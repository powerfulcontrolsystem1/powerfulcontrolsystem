package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/vpssecurity"
	"github.com/you/pos-backend/vpssecurity/config"
	"github.com/you/pos-backend/vpssecurity/reports"
)

type fakeSecurityVPSService struct {
	config          config.Settings
	savedConfig     config.Settings
	startStatus     vpssecurity.JobStatus
	startErr        error
	status          vpssecurity.JobStatus
	history         []reports.HistoryEntry
	reportContent   []byte
	reportFileName  string
	reportType      string
	reportErr       error
	comparison      reports.Comparison
	comparisonErr   error
	lastStartReq    vpssecurity.StartRequest
	lastTriggeredBy string
}

func (f *fakeSecurityVPSService) Config() (config.Settings, error) {
	return f.config, nil
}

func (f *fakeSecurityVPSService) SaveConfig(settings config.Settings) (config.Settings, error) {
	f.savedConfig = settings
	f.config = settings
	return settings, nil
}

func (f *fakeSecurityVPSService) StartScan(ctx context.Context, req vpssecurity.StartRequest, triggeredBy string) (vpssecurity.JobStatus, error) {
	f.lastStartReq = req
	f.lastTriggeredBy = triggeredBy
	return f.startStatus, f.startErr
}

func (f *fakeSecurityVPSService) Status(scanID string) (vpssecurity.JobStatus, error) {
	return f.status, nil
}

func (f *fakeSecurityVPSService) History(limit int) ([]reports.HistoryEntry, error) {
	return f.history, nil
}

func (f *fakeSecurityVPSService) ReportArtifact(scanID, format string) ([]byte, string, string, error) {
	if f.reportErr != nil {
		return nil, "", "", f.reportErr
	}
	return f.reportContent, f.reportFileName, f.reportType, nil
}

func (f *fakeSecurityVPSService) Compare(scanID, otherScanID string) (reports.Comparison, error) {
	return f.comparison, f.comparisonErr
}

func seedSuperAuthForSecurityVPS(t *testing.T) *sql.DB {
	t.Helper()
	dbSuper := openTestSQLite(t, "security_vps_super.db")
	ensureSuperSchema(t, dbSuper)
	if err := dbpkg.UpsertAdministrador(dbSuper, "super@pcs.com", "Super", "super_administrador", ""); err != nil {
		t.Fatalf("seed super admin: %v", err)
	}
	if err := dbpkg.CreateSession(dbSuper, "super@pcs.com", "127.0.0.1", "test-agent", "token-security-vps"); err != nil {
		t.Fatalf("seed super session: %v", err)
	}
	return dbSuper
}

func TestSecurityVPSConfigHandlerGetAndPut(t *testing.T) {
	dbSuper := seedSuperAuthForSecurityVPS(t)
	service := &fakeSecurityVPSService{config: config.Settings{TargetHost: "127.0.0.1", PortList: "22,80,443", Profile: "full"}}

	req := httptest.NewRequest(http.MethodGet, "/super/api/security/vps/config", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "token-security-vps"})
	rr := httptest.NewRecorder()
	SecurityVPSConfigHandler(dbSuper, service).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("config get status inesperado: %d body=%s", rr.Code, rr.Body.String())
	}
	var getPayload map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &getPayload); err != nil {
		t.Fatalf("decode get config: %v", err)
	}
	if ok, _ := getPayload["ok"].(bool); !ok {
		t.Fatalf("expected ok=true in get config")
	}
	configBlock, _ := getPayload["config"].(map[string]interface{})
	if got := configBlock["target_host"].(string); got != "127.0.0.1" {
		t.Fatalf("target_host inesperado: %s", got)
	}

	body := bytes.NewBufferString(`{"target_host":"10.0.0.10","port_list":"22,443","profile":"quick"}`)
	putReq := httptest.NewRequest(http.MethodPut, "/super/api/security/vps/config", body)
	putReq.AddCookie(&http.Cookie{Name: "session_token", Value: "token-security-vps"})
	putReq.Header.Set("Content-Type", "application/json")
	putRR := httptest.NewRecorder()
	SecurityVPSConfigHandler(dbSuper, service).ServeHTTP(putRR, putReq)
	if putRR.Code != http.StatusOK {
		t.Fatalf("config put status inesperado: %d body=%s", putRR.Code, putRR.Body.String())
	}
	if service.savedConfig.TargetHost != "10.0.0.10" || service.savedConfig.Profile != "quick" {
		t.Fatalf("config not persisted in fake service: %+v", service.savedConfig)
	}
}

func TestSecurityVPSHandlersRunStatusHistoryReportAndCompare(t *testing.T) {
	dbSuper := seedSuperAuthForSecurityVPS(t)
	service := &fakeSecurityVPSService{
		startStatus: vpssecurity.JobStatus{ScanID: "scan-1", Status: "running", Active: true},
		status:      vpssecurity.JobStatus{ScanID: "scan-1", Status: "completed", Active: false, Report: &reports.ScanReport{ScanID: "scan-1", Status: "completed", GeneratedAt: "2026-04-16T10:00:00Z", TargetHost: "127.0.0.1", Profile: "full"}},
		history: []reports.HistoryEntry{{ScanID: "scan-1", GeneratedAt: "2026-04-16T10:00:00Z", TargetHost: "127.0.0.1", Profile: "full", Status: "completed", TotalFindings: 3, HighestSeverity: "ALTO"}},
		reportContent:  []byte("reporte txt"),
		reportFileName: "report.txt",
		reportType:     "text/plain; charset=utf-8",
		comparison: reports.Comparison{PreviousScanID: "scan-0", NewFindings: 1, ResolvedFindings: 2, Summary: "1 hallazgo nuevo; 2 resueltos"},
	}

	runReq := httptest.NewRequest(http.MethodPost, "/super/api/security/vps/run", bytes.NewBufferString(`{"target_host":"10.10.10.10","profile":"full"}`))
	runReq.AddCookie(&http.Cookie{Name: "session_token", Value: "token-security-vps"})
	runReq.Header.Set("Content-Type", "application/json")
	runRR := httptest.NewRecorder()
	SecurityVPSRunHandler(dbSuper, service).ServeHTTP(runRR, runReq)
	if runRR.Code != http.StatusAccepted {
		t.Fatalf("run status inesperado: %d body=%s", runRR.Code, runRR.Body.String())
	}
	if service.lastStartReq.TargetHost != "10.10.10.10" || service.lastTriggeredBy != "super@pcs.com" {
		t.Fatalf("run handler did not forward request correctly: req=%+v admin=%s", service.lastStartReq, service.lastTriggeredBy)
	}

	statusReq := httptest.NewRequest(http.MethodGet, "/super/api/security/vps/status?scan_id=scan-1", nil)
	statusReq.AddCookie(&http.Cookie{Name: "session_token", Value: "token-security-vps"})
	statusRR := httptest.NewRecorder()
	SecurityVPSStatusHandler(dbSuper, service).ServeHTTP(statusRR, statusReq)
	if statusRR.Code != http.StatusOK {
		t.Fatalf("status handler returned %d body=%s", statusRR.Code, statusRR.Body.String())
	}

	historyReq := httptest.NewRequest(http.MethodGet, "/super/api/security/vps/history?limit=5", nil)
	historyReq.AddCookie(&http.Cookie{Name: "session_token", Value: "token-security-vps"})
	historyRR := httptest.NewRecorder()
	SecurityVPSHistoryHandler(dbSuper, service).ServeHTTP(historyRR, historyReq)
	if historyRR.Code != http.StatusOK {
		t.Fatalf("history handler returned %d body=%s", historyRR.Code, historyRR.Body.String())
	}

	reportReq := httptest.NewRequest(http.MethodGet, "/super/api/security/vps/report?scan_id=scan-1&format=txt", nil)
	reportReq.AddCookie(&http.Cookie{Name: "session_token", Value: "token-security-vps"})
	reportRR := httptest.NewRecorder()
	SecurityVPSReportHandler(dbSuper, service).ServeHTTP(reportRR, reportReq)
	if reportRR.Code != http.StatusOK {
		t.Fatalf("report handler returned %d body=%s", reportRR.Code, reportRR.Body.String())
	}
	if got := reportRR.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Fatalf("content-type inesperado: %s", got)
	}
	if body := reportRR.Body.String(); body != "reporte txt" {
		t.Fatalf("report body inesperado: %q", body)
	}

	compareReq := httptest.NewRequest(http.MethodGet, "/super/api/security/vps/compare?scan_id=scan-1", nil)
	compareReq.AddCookie(&http.Cookie{Name: "session_token", Value: "token-security-vps"})
	compareRR := httptest.NewRecorder()
	SecurityVPSCompareHandler(dbSuper, service).ServeHTTP(compareRR, compareReq)
	if compareRR.Code != http.StatusOK {
		t.Fatalf("compare handler returned %d body=%s", compareRR.Code, compareRR.Body.String())
	}
}

func TestSecurityVPSRunHandlerReturnsConflictWhenScanIsRunning(t *testing.T) {
	dbSuper := seedSuperAuthForSecurityVPS(t)
	service := &fakeSecurityVPSService{
		startStatus: vpssecurity.JobStatus{ScanID: "scan-active", Status: "running", Active: true},
		startErr:    vpssecurity.ErrScanRunning,
	}
	req := httptest.NewRequest(http.MethodPost, "/super/api/security/vps/run", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "token-security-vps"})
	rr := httptest.NewRecorder()
	SecurityVPSRunHandler(dbSuper, service).ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected conflict, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestSecurityVPSReportHandlerReturnsNotFound(t *testing.T) {
	dbSuper := seedSuperAuthForSecurityVPS(t)
	service := &fakeSecurityVPSService{reportErr: errors.New("not found")}
	req := httptest.NewRequest(http.MethodGet, "/super/api/security/vps/report?scan_id=scan-missing&format=json", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "token-security-vps"})
	rr := httptest.NewRecorder()
	SecurityVPSReportHandler(dbSuper, service).ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rr.Code, rr.Body.String())
	}
}