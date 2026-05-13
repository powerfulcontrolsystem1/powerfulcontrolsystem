package vpssecurity

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/you/pos-backend/vpssecurity/config"
	storepkg "github.com/you/pos-backend/vpssecurity/logstore"
	"github.com/you/pos-backend/vpssecurity/reports"
	"github.com/you/pos-backend/vpssecurity/scanner"
)

var ErrScanRunning = errors.New("already running")

type StartRequest struct {
	TargetHost string `json:"target_host"`
	PortList   string `json:"port_list"`
	Profile    string `json:"profile"`
	Trigger    string `json:"trigger"`
}

type JobStatus struct {
	ScanID      string              `json:"scan_id,omitempty"`
	Status      string              `json:"status"`
	Active      bool                `json:"active"`
	StartedAt   string              `json:"started_at,omitempty"`
	CompletedAt string              `json:"completed_at,omitempty"`
	Error       string              `json:"error,omitempty"`
	Report      *reports.ScanReport `json:"report,omitempty"`
}

type Service struct {
	manager  *config.Manager
	store    *storepkg.Store
	executor scanner.Executor

	mu      sync.Mutex
	current *jobState
}

type jobState struct {
	status      JobStatus
	request     StartRequest
	triggeredBy string
}

func NewService(manager *config.Manager, store *storepkg.Store, executor scanner.Executor) (*Service, error) {
	if manager == nil {
		manager = config.NewManager("")
	}
	if store == nil {
		store = storepkg.NewStore(config.DefaultDataDir())
	}
	if err := store.Ensure(); err != nil {
		return nil, err
	}
	if executor == nil {
		executor = scanner.SystemExecutor{}
	}
	return &Service{manager: manager, store: store, executor: executor}, nil
}

func (s *Service) Config() (config.Settings, error) {
	return s.manager.Load()
}

func (s *Service) SaveConfig(settings config.Settings) (config.Settings, error) {
	return s.manager.Save(settings)
}

func (s *Service) StartScan(ctx context.Context, req StartRequest, triggeredBy string) (JobStatus, error) {
	settings, err := s.manager.Load()
	if err != nil {
		return JobStatus{}, err
	}
	applyOverrides(&settings, req)
	s.mu.Lock()
	if s.current != nil && s.current.status.Active {
		status := s.current.status
		s.mu.Unlock()
		return status, ErrScanRunning
	}
	status := JobStatus{
		ScanID:    newScanID(),
		Status:    "running",
		Active:    true,
		StartedAt: time.Now().UTC().Format(time.RFC3339),
	}
	job := &jobState{status: status, request: req, triggeredBy: strings.TrimSpace(triggeredBy)}
	s.current = job
	s.mu.Unlock()
	go s.run(job, settings)
	return status, nil
}

func applyOverrides(settings *config.Settings, req StartRequest) {
	if settings == nil {
		return
	}
	if strings.TrimSpace(req.TargetHost) != "" {
		settings.TargetHost = strings.TrimSpace(req.TargetHost)
	}
	if strings.TrimSpace(req.PortList) != "" {
		settings.PortList = strings.TrimSpace(req.PortList)
	}
	if strings.TrimSpace(req.Profile) != "" {
		settings.Profile = strings.ToLower(strings.TrimSpace(req.Profile))
	}
}

func (s *Service) run(job *jobState, settings config.Settings) {
	report, err := RunOnce(context.Background(), settings, s.store, s.executor, defaultTrigger(job.request.Trigger), job.triggeredBy, job.status.ScanID)
	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil {
		job.status.Status = "failed"
		job.status.Active = false
		job.status.CompletedAt = time.Now().UTC().Format(time.RFC3339)
		job.status.Error = err.Error()
		return
	}
	job.status.Status = "completed"
	job.status.Active = false
	job.status.CompletedAt = report.CompletedAt
	job.status.Report = report
	job.status.Error = ""
	if s.current == job {
		s.current = nil
	}
}

func defaultTrigger(trigger string) string {
	if strings.TrimSpace(trigger) == "" {
		return "super_panel"
	}
	return strings.TrimSpace(trigger)
}

func RunOnce(ctx context.Context, settings config.Settings, store *storepkg.Store, executor scanner.Executor, trigger, triggeredBy, scanID string) (*reports.ScanReport, error) {
	if strings.TrimSpace(scanID) == "" {
		scanID = newScanID()
	}
	if store == nil {
		store = storepkg.NewStore(settings.DataDir)
		if err := store.Ensure(); err != nil {
			return nil, err
		}
	}
	startedAt := time.Now().UTC()
	runResult := scanner.RunAudit(ctx, settings, executor)
	report := &reports.ScanReport{
		ScanID:      scanID,
		Status:      "completed",
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		StartedAt:   startedAt.Format(time.RFC3339),
		CompletedAt: time.Now().UTC().Format(time.RFC3339),
		DurationMs:  time.Since(startedAt).Milliseconds(),
		Trigger:     trigger,
		TriggeredBy: strings.TrimSpace(triggeredBy),
		TargetHost:  settings.TargetHost,
		Profile:     settings.Profile,
		Config: reports.ConfigSnapshot{
			TargetHost:           settings.TargetHost,
			PortList:             settings.PortList,
			Profile:              settings.Profile,
			Cron:                 settings.Schedule.Cron,
			EnabledTools:         enabledTools(settings),
			VulnerabilityScanner: settings.VulnerabilityScan.Provider,
		},
		Tools:      runResult.Tools,
		SystemInfo: runResult.SystemInfo,
		Findings:   runResult.Findings,
		Notes:      append([]string(nil), runResult.Notes...),
		Errors:     append([]string(nil), runResult.Errors...),
		Reports:    buildReportLinks(scanID),
	}
	report.Summary.HardeningIndex = runResult.HardeningIndex
	reports.ApplySummary(report)
	if previous, err := store.LoadLatest(); err == nil && previous != nil && previous.ScanID != report.ScanID {
		report.Comparison = reports.Compare(report, previous)
	}
	generated, err := reports.GenerateArtifacts(*report)
	if err != nil {
		return nil, err
	}
	rawArtifacts := make(map[string][]byte, len(runResult.Artifacts))
	for _, artifact := range runResult.Artifacts {
		rawArtifacts[artifact.Name] = artifact.Content
	}
	if err := store.Save(report, generated, rawArtifacts); err != nil {
		return nil, err
	}
	return report, nil
}

func enabledTools(settings config.Settings) []string {
	enabled := make([]string, 0, 3)
	if settings.Lynis.Enabled {
		enabled = append(enabled, "lynis")
	}
	if settings.Nmap.Enabled {
		enabled = append(enabled, "nmap")
	}
	if settings.VulnerabilityScan.Enabled {
		enabled = append(enabled, settings.VulnerabilityScan.Provider)
	}
	return enabled
}

func buildReportLinks(scanID string) map[string]string {
	return map[string]string{
		"json": fmt.Sprintf("/super/api/security/vps/report?scan_id=%s&format=json", scanID),
		"txt":  fmt.Sprintf("/super/api/security/vps/report?scan_id=%s&format=txt", scanID),
		"html": fmt.Sprintf("/super/api/security/vps/report?scan_id=%s&format=html", scanID),
		"csv":  fmt.Sprintf("/super/api/security/vps/report?scan_id=%s&format=csv", scanID),
		"pdf":  fmt.Sprintf("/super/api/security/vps/report?scan_id=%s&format=pdf", scanID),
		"xls":  fmt.Sprintf("/super/api/security/vps/report?scan_id=%s&format=xls", scanID),
	}
}

func (s *Service) Status(scanID string) (JobStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current != nil && (scanID == "" || s.current.status.ScanID == scanID) {
		return s.current.status, nil
	}
	if strings.TrimSpace(scanID) == "" {
		report, err := s.store.LoadLatest()
		if err != nil || report == nil {
			return JobStatus{Status: "idle", Active: false}, nil
		}
		return JobStatus{ScanID: report.ScanID, Status: "completed", Active: false, StartedAt: report.StartedAt, CompletedAt: report.CompletedAt, Report: report}, nil
	}
	report, err := s.store.Load(scanID)
	if err != nil {
		return JobStatus{}, err
	}
	return JobStatus{ScanID: report.ScanID, Status: "completed", Active: false, StartedAt: report.StartedAt, CompletedAt: report.CompletedAt, Report: report}, nil
}

func (s *Service) History(limit int) ([]reports.HistoryEntry, error) {
	return s.store.List(limit)
}

func (s *Service) Report(scanID string) (*reports.ScanReport, error) {
	return s.store.Load(scanID)
}

func (s *Service) ReportArtifact(scanID, format string) ([]byte, string, string, error) {
	return s.store.LoadArtifact(scanID, format)
}

func (s *Service) Compare(scanID, otherScanID string) (reports.Comparison, error) {
	current, err := s.store.Load(scanID)
	if err != nil {
		return reports.Comparison{}, err
	}
	var previous *reports.ScanReport
	if strings.TrimSpace(otherScanID) != "" {
		previous, err = s.store.Load(otherScanID)
		if err != nil {
			return reports.Comparison{}, err
		}
	} else {
		previous, err = s.store.LoadPreviousBefore(scanID)
		if err != nil {
			return reports.Comparison{}, err
		}
	}
	if previous == nil {
		return reports.Comparison{}, nil
	}
	return reports.Compare(current, previous), nil
}

func newScanID() string {
	buffer := make([]byte, 4)
	if _, err := rand.Read(buffer); err != nil {
		return fmt.Sprintf("scan-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("scan-%s-%s", time.Now().UTC().Format("20060102T150405Z"), hex.EncodeToString(buffer))
}
