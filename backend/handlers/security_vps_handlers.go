package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/you/pos-backend/vpssecurity"
	"github.com/you/pos-backend/vpssecurity/config"
	"github.com/you/pos-backend/vpssecurity/reports"
)

type securityVPSService interface {
	Config() (config.Settings, error)
	SaveConfig(settings config.Settings) (config.Settings, error)
	StartScan(ctx context.Context, req vpssecurity.StartRequest, triggeredBy string) (vpssecurity.JobStatus, error)
	Status(scanID string) (vpssecurity.JobStatus, error)
	History(limit int) ([]reports.HistoryEntry, error)
	ReportArtifact(scanID, format string) ([]byte, string, string, error)
	Compare(scanID, otherScanID string) (reports.Comparison, error)
}

func SecurityVPSConfigHandler(dbSuper *sql.DB, service securityVPSService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminEmail, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper)
		if !ok {
			return
		}
		switch r.Method {
		case http.MethodGet:
			settings, err := service.Config()
			if err != nil {
				writeSecurityVPSPublicError(w, r, http.StatusInternalServerError, "cargar configuracion", err, nil)
				return
			}
			status, _ := service.Status("")
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":          true,
				"config":      settings,
				"admin_email": adminEmail,
				"status":      status,
			})
		case http.MethodPut:
			var payload config.Settings
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "payload invalido"})
				return
			}
			settings, err := service.SaveConfig(payload)
			if err != nil {
				writeSecurityVPSPublicError(w, r, http.StatusInternalServerError, "guardar configuracion", err, nil)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "config": settings, "admin_email": adminEmail})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func SecurityVPSRunHandler(dbSuper *sql.DB, service securityVPSService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminEmail, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper)
		if !ok {
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var payload vpssecurity.StartRequest
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&payload)
		}
		status, err := service.StartScan(r.Context(), payload, adminEmail)
		if err != nil {
			code := http.StatusInternalServerError
			if errors.Is(err, vpssecurity.ErrScanRunning) {
				code = http.StatusConflict
			}
			writeSecurityVPSPublicError(w, r, code, "iniciar escaneo", err, map[string]interface{}{"status": status})
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]interface{}{"ok": true, "status": status})
	}
}

func SecurityVPSStatusHandler(dbSuper *sql.DB, service securityVPSService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper); !ok {
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		status, err := service.Status(strings.TrimSpace(r.URL.Query().Get("scan_id")))
		if err != nil {
			writeSecurityVPSPublicError(w, r, http.StatusNotFound, "consultar estado", err, nil)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "status": status})
	}
}

func SecurityVPSHistoryHandler(dbSuper *sql.DB, service securityVPSService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper); !ok {
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		limit := 20
		if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
				limit = parsed
			}
		}
		history, err := service.History(limit)
		if err != nil {
			writeSecurityVPSPublicError(w, r, http.StatusInternalServerError, "consultar historial", err, nil)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "items": history})
	}
}

func SecurityVPSReportHandler(dbSuper *sql.DB, service securityVPSService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper); !ok {
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		scanID := strings.TrimSpace(r.URL.Query().Get("scan_id"))
		if scanID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "scan_id es obligatorio"})
			return
		}
		format := strings.TrimSpace(r.URL.Query().Get("format"))
		if format == "" {
			format = "json"
		}
		content, fileName, contentType, err := service.ReportArtifact(scanID, format)
		if err != nil {
			writeSecurityVPSPublicError(w, r, http.StatusNotFound, "descargar reporte", err, nil)
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileName))
		_, _ = w.Write(content)
	}
}

func SecurityVPSCompareHandler(dbSuper *sql.DB, service securityVPSService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper); !ok {
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		scanID := strings.TrimSpace(r.URL.Query().Get("scan_id"))
		if scanID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "scan_id es obligatorio"})
			return
		}
		comparison, err := service.Compare(scanID, strings.TrimSpace(r.URL.Query().Get("other_scan_id")))
		if err != nil {
			writeSecurityVPSPublicError(w, r, http.StatusNotFound, "comparar escaneos", err, nil)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "comparison": comparison})
	}
}

func writeSecurityVPSPublicError(w http.ResponseWriter, r *http.Request, status int, operation string, err error, extra map[string]interface{}) {
	requestID := resolveAuditoriaRequestID(r)
	log.Printf("[security_vps] operation=%s request_id=%s error_type=%T", operation, requestID, err)
	payload := map[string]interface{}{
		"ok":    false,
		"code":  "security_vps_error",
		"error": "No se pudo completar la operacion de seguridad VPS.",
	}
	for key, value := range extra {
		payload[key] = value
	}
	if requestID != "" {
		payload["request_id"] = requestID
	}
	writeJSON(w, status, payload)
}
