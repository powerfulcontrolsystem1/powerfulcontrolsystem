package handlers

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	maxSoporteComprasIAUploadBytes     = 15 << 20
	maxSoporteComprasIAExtractionBytes = 128 << 10
	soporteComprasIAExtractionTimeout  = 90 * time.Second
	soporteComprasIAAdvisoryNamespace  = int64(0x534349)
	soporteComprasIAStorageNamespace   = int64(0x534351)
	soporteComprasIAPurgeNamespace     = int64(0x534350)
)

var soporteComprasIAAllowedExt = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".webp": true,
	".pdf":  true,
	".xml":  true,
}

var (
	soporteComprasIAAntivirusClean            atomic.Uint64
	soporteComprasIAAntivirusMalware          atomic.Uint64
	soporteComprasIAAntivirusUnavailable      atomic.Uint64
	soporteComprasIAAntivirusBypassed         atomic.Uint64
	soporteComprasIAExtractionConsistent      atomic.Uint64
	soporteComprasIAExtractionHumanReview     atomic.Uint64
	soporteComprasIAExtractionProviderError   atomic.Uint64
	soporteComprasIAExtractionInvalidResponse atomic.Uint64
	soporteComprasIAExtractionCanceled        atomic.Uint64
	soporteComprasIAExtractionTimeouts        atomic.Uint64
	soporteComprasIAExtractionPersistence     atomic.Uint64
	soporteComprasIAFileIntegrityFailures     atomic.Uint64
)

// SupportAntivirusMetrics contains process-level, tenant-free counters for the
// private purchase-support scanner. Result values are intentionally bounded.
type SupportAntivirusMetrics struct {
	Clean       uint64
	Malware     uint64
	Unavailable uint64
	Bypassed    uint64
	Required    bool
	Configured  bool
}

// SupportAntivirusOperationalMetrics returns aggregate scanner state without
// company, user, filename, route, provider response or attachment contents.
func SupportAntivirusOperationalMetrics() SupportAntivirusMetrics {
	return SupportAntivirusMetrics{
		Clean:       soporteComprasIAAntivirusClean.Load(),
		Malware:     soporteComprasIAAntivirusMalware.Load(),
		Unavailable: soporteComprasIAAntivirusUnavailable.Load(),
		Bypassed:    soporteComprasIAAntivirusBypassed.Load(),
		Required:    parseBoolSoporteComprasIA(os.Getenv("PCS_SUPPORTS_CLAMAV_REQUIRED")),
		Configured:  strings.TrimSpace(os.Getenv("PCS_SUPPORTS_CLAMAV_ADDR")) != "",
	}
}

// SupportExtractionMetrics contains bounded process-level outcomes for the IA
// extraction pipeline. It deliberately excludes tenant and document data.
type SupportExtractionMetrics struct {
	Consistent      uint64
	HumanReview     uint64
	ProviderError   uint64
	InvalidResponse uint64
	Canceled        uint64
	Timeout         uint64
	Persistence     uint64
}

func SupportExtractionOperationalMetrics() SupportExtractionMetrics {
	return SupportExtractionMetrics{
		Consistent:      soporteComprasIAExtractionConsistent.Load(),
		HumanReview:     soporteComprasIAExtractionHumanReview.Load(),
		ProviderError:   soporteComprasIAExtractionProviderError.Load(),
		InvalidResponse: soporteComprasIAExtractionInvalidResponse.Load(),
		Canceled:        soporteComprasIAExtractionCanceled.Load(),
		Timeout:         soporteComprasIAExtractionTimeouts.Load(),
		Persistence:     soporteComprasIAExtractionPersistence.Load(),
	}
}

func recordSupportExtractionOutcome(outcome string) {
	switch outcome {
	case "consistent":
		soporteComprasIAExtractionConsistent.Add(1)
	case "human_review":
		soporteComprasIAExtractionHumanReview.Add(1)
	case "provider_error":
		soporteComprasIAExtractionProviderError.Add(1)
	case "invalid_response":
		soporteComprasIAExtractionInvalidResponse.Add(1)
	case "canceled":
		soporteComprasIAExtractionCanceled.Add(1)
	case "timeout":
		soporteComprasIAExtractionTimeouts.Add(1)
	case "persistence_error":
		soporteComprasIAExtractionPersistence.Add(1)
	}
}

// SupportFileOperationalMetrics exposes only an aggregate integrity failure
// counter. It never reports a tenant, document, hash or private path.
func SupportFileOperationalMetrics() uint64 {
	return soporteComprasIAFileIntegrityFailures.Load()
}

// EmpresaSoportesComprasIAHandler administra la captura inteligente de compras y gastos.
func EmpresaSoportesComprasIAHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil || empresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "" {
			action = "dashboard"
		}

		switch r.Method {
		case http.MethodGet:
			handleSoportesComprasIAGet(w, r, dbEmp, empresaID, action)
		case http.MethodPost, http.MethodPut:
			handleSoportesComprasIAMutate(w, r, dbEmp, dbSuper, empresaID, action)
		default:
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
		}
	}
}

func handleSoportesComprasIAGet(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, empresaID int64, action string) {
	switch action {
	case "dashboard":
		dashboard, err := dbpkg.BuildEmpresaSoportesComprasIADashboardContext(r.Context(), dbEmp, empresaID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "dashboard": dashboard})
	case "soportes":
		estado := r.URL.Query().Get("estado")
		registro := r.URL.Query().Get("registro")
		rows, err := dbpkg.ListEmpresaSoportesComprasIARegistroContext(r.Context(), dbEmp, empresaID, estado, registro, 300)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for i := range rows {
			exposeSoporteComprasIAURL(&rows[i])
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "soportes": rows})
	case "retencion_preview":
		days := parseSoporteComprasIARetentionDays(r.URL.Query().Get("retencion_dias"))
		rows, err := dbpkg.ListEmpresaSoportesComprasIARetencionContext(r.Context(), dbEmp, empresaID, days, 300)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		bytes := soporteComprasIARetentionBytes(rows)
		for i := range rows {
			exposeSoporteComprasIAURL(&rows[i])
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok": true, "retencion_dias": days, "candidatos": len(rows), "bytes": bytes, "soportes": rows,
		})
	case "cuarentena_preview":
		thresholdMinutes := parseSoporteComprasIAQuarantineThreshold(r.URL.Query().Get("umbral_minutos"))
		pending, err := dbpkg.ListEmpresaSoportesComprasIARegistroContext(r.Context(), dbEmp, empresaID, "", "purga_pendiente", 500)
		if err != nil {
			log.Printf("[soportes_compras_ia] diagnostico cuarentena empresa_id=%d: %v", empresaID, err)
			http.Error(w, "no se pudo consultar el diagnostico de cuarentena", http.StatusInternalServerError)
			return
		}
		files, bytes, err := soporteComprasIAQuarantineStats(empresaID)
		if err != nil {
			log.Printf("[soportes_compras_ia] filesystem cuarentena empresa_id=%d: %v", empresaID, err)
			http.Error(w, "no se pudo inspeccionar la cuarentena privada", http.StatusInternalServerError)
			return
		}
		stale := countSoporteComprasIAStalePending(pending, time.Now(), time.Duration(thresholdMinutes)*time.Minute)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok": true, "registros_pendientes": len(pending), "archivos_cuarentena": files,
			"bytes": bytes, "umbral_minutos": thresholdMinutes,
			"pendientes_vencidos": stale, "requiere_revision": len(pending) != files || stale > 0,
		})
	case "descargar":
		downloadSoporteComprasIA(w, r, dbEmp, empresaID)
	case "eventos":
		soporteID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("soporte_id")), 10, 64)
		rows, err := dbpkg.ListEmpresaSoportesComprasIAEventosContext(r.Context(), dbEmp, empresaID, soporteID, 200)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "eventos": rows})
	default:
		http.Error(w, "accion no soportada", http.StatusBadRequest)
	}
}

func handleSoportesComprasIAMutate(w http.ResponseWriter, r *http.Request, dbEmp, dbSuper *sql.DB, empresaID int64, action string) {
	usuario := strings.TrimSpace(adminEmailFromRequest(r))
	switch action {
	case "radicar":
		row, err := radicarSoporteComprasIA(r, dbEmp, dbSuper, empresaID, usuario)
		if err != nil {
			if errors.Is(err, errSoporteComprasIAMalware) {
				writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{"ok": false, "error": "El antivirus rechazo el archivo del soporte"})
				return
			}
			if errors.Is(err, errSoporteComprasIAAntivirusUnavailable) {
				writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{"ok": false, "error": "El antivirus de soportes no esta disponible. Intenta nuevamente."})
				return
			}
			if errors.Is(err, errSoporteComprasIAStorageQuota) {
				writeJSON(w, http.StatusInsufficientStorage, map[string]interface{}{
					"ok":    false,
					"error": "La empresa alcanzo el limite de almacenamiento configurado",
				})
				return
			}
			if errors.Is(err, errSoporteComprasIAStorageBusy) {
				writeJSON(w, http.StatusConflict, map[string]interface{}{
					"ok":    false,
					"error": "Otra carga de la empresa esta finalizando. Intenta nuevamente.",
				})
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		exposeSoporteComprasIAURL(&row)
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "soporte": row})
	case "extraer_ia":
		row, err := extraerSoporteComprasIAGPT55(r, dbEmp, dbSuper, empresaID, usuario)
		if err != nil {
			status := http.StatusInternalServerError
			message := "No se pudo completar la extraccion IA. Intenta nuevamente."
			if errors.Is(err, errSoporteComprasIAIADesactivada) || errors.Is(err, errSoporteComprasIAModeloNoDisponible) {
				status = http.StatusServiceUnavailable
				message = err.Error()
			} else if errors.Is(err, errSoporteComprasIAProcesamientoEnCurso) {
				status = http.StatusConflict
				message = "El soporte ya se esta procesando con IA. Espera el resultado antes de reintentar."
			} else if errors.Is(err, errSoporteComprasIASolicitudInvalida) {
				status = http.StatusBadRequest
				message = "Selecciona un soporte valido antes de ejecutar la extraccion IA."
			} else if errors.Is(err, errSoporteComprasIAEstadoNoExtraible) {
				status = http.StatusConflict
				message = "El estado actual del soporte no permite ejecutar nuevamente la extraccion IA."
			} else if errors.Is(err, errSoporteComprasIASinAdjunto) {
				status = http.StatusUnprocessableEntity
				message = "El soporte no tiene un archivo privado disponible para analizar."
			} else if errors.Is(err, context.DeadlineExceeded) {
				status = http.StatusGatewayTimeout
				message = "El proveedor de IA excedio el tiempo de espera. El soporte conserva su estado anterior."
			} else if errors.Is(err, context.Canceled) {
				status = http.StatusRequestTimeout
				message = "La extraccion IA fue cancelada. El soporte conserva su estado anterior."
			} else if errors.Is(err, errSoporteComprasIAProveedor) {
				status = http.StatusBadGateway
				message = publicAIProviderError(err)
			} else if isProviderLimitError(err) {
				status = http.StatusTooManyRequests
				message = publicAIProviderError(err)
			}
			log.Printf("[soportes_compras_ia] extraccion empresa_id=%d status=%d: %v", empresaID, status, err)
			http.Error(w, message, status)
			return
		}
		exposeSoporteComprasIAURL(&row)
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "soporte": row})
	case "editar_revision":
		payload, err := decodeSoporteComprasIARevisionPayload(r)
		if err != nil {
			http.Error(w, "datos de revision invalidos", http.StatusBadRequest)
			return
		}
		row, err := dbpkg.UpdateEmpresaSoporteComprasIARevision(dbEmp, empresaID, payload.SoporteID, payload.EmpresaSoporteComprasIA, usuario)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		exposeSoporteComprasIAURL(&row)
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "soporte": row})
	case "aprobar":
		payload, err := decodeSoporteComprasIAActionPayload(r)
		if err != nil || payload.SoporteID <= 0 {
			http.Error(w, "soporte_id es obligatorio", http.StatusBadRequest)
			return
		}
		row, err := dbpkg.UpdateEmpresaSoporteComprasIAEstadoContext(r.Context(), dbEmp, empresaID, payload.SoporteID, "aprobado", usuario, payload.Observaciones)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		exposeSoporteComprasIAURL(&row)
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "soporte": row})
	case "rechazar":
		payload, err := decodeSoporteComprasIAActionPayload(r)
		if err != nil || payload.SoporteID <= 0 {
			http.Error(w, "soporte_id es obligatorio", http.StatusBadRequest)
			return
		}
		row, err := dbpkg.UpdateEmpresaSoporteComprasIAEstadoContext(r.Context(), dbEmp, empresaID, payload.SoporteID, "rechazado", usuario, payload.Observaciones)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		exposeSoporteComprasIAURL(&row)
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "soporte": row})
	case "contabilizar":
		payload, err := decodeSoporteComprasIAActionPayload(r)
		if err != nil || payload.SoporteID <= 0 {
			http.Error(w, "soporte_id es obligatorio", http.StatusBadRequest)
			return
		}
		row, err := dbpkg.ContabilizarEmpresaSoporteComprasIA(dbEmp, empresaID, payload.SoporteID, usuario)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		exposeSoporteComprasIAURL(&row)
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "soporte": row})
	case "eliminar", "restaurar":
		payload, err := decodeSoporteComprasIAActionPayload(r)
		if err != nil || payload.SoporteID <= 0 {
			http.Error(w, "soporte_id es obligatorio", http.StatusBadRequest)
			return
		}
		next := "eliminado"
		if action == "restaurar" {
			next = "activo"
		}
		row, err := dbpkg.UpdateEmpresaSoporteComprasIARegistroEstadoContext(r.Context(), dbEmp, empresaID, payload.SoporteID, next, usuario, payload.Observaciones)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		exposeSoporteComprasIAURL(&row)
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "soporte": row})
	case "purgar":
		payload, err := decodeSoporteComprasIAActionPayload(r)
		if err != nil || payload.SoporteID <= 0 {
			http.Error(w, "datos de depuracion invalidos", http.StatusBadRequest)
			return
		}
		releasePurge, err := acquireSoporteComprasIAPurgeLock(r, dbEmp, empresaID)
		if err != nil {
			log.Printf("[soportes_compras_ia] lock depuracion empresa_id=%d soporte_id=%d: %v", empresaID, payload.SoporteID, err)
			http.Error(w, "otra depuracion de la empresa esta en curso; reintenta", http.StatusConflict)
			return
		}
		defer releasePurge()
		row, err := dbpkg.GetEmpresaSoporteComprasIA(dbEmp, empresaID, payload.SoporteID)
		if err != nil {
			http.Error(w, "soporte no disponible", http.StatusNotFound)
			return
		}
		quarantine, err := prepareSoporteComprasIAPurgeFile(row)
		if err != nil {
			http.Error(w, "el archivo privado no se pudo preparar de forma segura", http.StatusConflict)
			return
		}
		if _, err := dbpkg.BeginPurgeEmpresaSoporteComprasIA(dbEmp, empresaID, payload.SoporteID, payload.RetentionDays, payload.Confirmation, usuario, payload.Observaciones); err != nil {
			if rollbackErr := quarantine.rollback(); rollbackErr != nil {
				log.Printf("[soportes_compras_ia] ALERTA rollback de cuarentena fallo empresa_id=%d soporte_id=%d: %v", empresaID, payload.SoporteID, rollbackErr)
			}
			if isSoporteComprasIAPurgePublicValidation(err) {
				http.Error(w, err.Error(), http.StatusBadRequest)
			} else {
				log.Printf("[soportes_compras_ia] inicio de depuracion empresa_id=%d soporte_id=%d: %v", empresaID, payload.SoporteID, err)
				http.Error(w, "no se pudo iniciar la depuracion del soporte", http.StatusInternalServerError)
			}
			return
		}
		if err := quarantine.commit(); err != nil {
			log.Printf("[soportes_compras_ia] depuracion pendiente empresa_id=%d soporte_id=%d: %v", empresaID, payload.SoporteID, err)
			http.Error(w, "la depuracion quedo pendiente y puede reintentarse de forma segura", http.StatusServiceUnavailable)
			return
		}
		purged, err := dbpkg.FinalizePurgeEmpresaSoporteComprasIA(dbEmp, empresaID, payload.SoporteID, usuario)
		if err != nil {
			log.Printf("[soportes_compras_ia] archivo eliminado con metadato pendiente empresa_id=%d soporte_id=%d: %v", empresaID, payload.SoporteID, err)
			http.Error(w, "el archivo fue eliminado y la auditoria final quedo pendiente; reintenta", http.StatusServiceUnavailable)
			return
		}
		exposeSoporteComprasIAURL(&purged)
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "soporte": purged})
	case "seed_demo":
		row, err := seedSoporteComprasIADemo(dbEmp, empresaID, usuario)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "soporte": row})
	default:
		http.Error(w, "accion no soportada", http.StatusBadRequest)
	}
}

func isSoporteComprasIAPurgePublicValidation(err error) bool {
	if err == nil {
		return false
	}
	message := err.Error()
	for _, allowed := range []string{
		"empresa_id y soporte_id son obligatorios",
		"retencion_dias debe estar entre 1 y 3650",
		"el motivo es obligatorio",
		"el motivo no puede superar 500 caracteres",
		"la confirmacion no coincide con el codigo del soporte",
		"solo un soporte de la papelera puede depurarse",
		"un soporte contabilizado no puede depurarse porque conserva trazabilidad contable",
		"el soporte aun no cumple la retencion configurada",
	} {
		if message == allowed {
			return true
		}
	}
	return false
}

type soporteComprasIAActionPayload struct {
	SoporteID     int64  `json:"soporte_id"`
	Observaciones string `json:"observaciones"`
	RetentionDays int    `json:"retencion_dias"`
	Confirmation  string `json:"confirmacion"`
}

type soporteComprasIARevisionPayload struct {
	SoporteID int64 `json:"soporte_id"`
	dbpkg.EmpresaSoporteComprasIA
}

func decodeSoporteComprasIAActionPayload(r *http.Request) (soporteComprasIAActionPayload, error) {
	var p soporteComprasIAActionPayload
	if strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
		err := json.NewDecoder(r.Body).Decode(&p)
		return p, err
	}
	p.SoporteID, _ = strconv.ParseInt(strings.TrimSpace(r.FormValue("soporte_id")), 10, 64)
	p.Observaciones = strings.TrimSpace(r.FormValue("observaciones"))
	p.RetentionDays, _ = strconv.Atoi(strings.TrimSpace(r.FormValue("retencion_dias")))
	p.Confirmation = strings.TrimSpace(r.FormValue("confirmacion"))
	return p, nil
}

func decodeSoporteComprasIARevisionPayload(r *http.Request) (soporteComprasIARevisionPayload, error) {
	var p soporteComprasIARevisionPayload
	if !strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
		return p, errors.New("se requiere JSON")
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&p); err != nil {
		return p, err
	}
	if p.SoporteID <= 0 {
		return p, errors.New("soporte_id es obligatorio")
	}
	return p, nil
}

func radicarSoporteComprasIA(r *http.Request, dbEmp, dbSuper *sql.DB, empresaID int64, usuario string) (dbpkg.EmpresaSoporteComprasIA, error) {
	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	var row dbpkg.EmpresaSoporteComprasIA
	row.EmpresaID = empresaID
	row.Usuario = usuario
	row.ModeloIA = dbpkg.EmpresaSoporteComprasIAModeloDefault
	row.EstadoSoporte = "radicado"
	row.RequiereRevisionHumana = true

	if strings.Contains(contentType, "multipart/form-data") {
		att, err := parseSingleAttachmentFromMultipart(r, "archivo", maxSoporteComprasIAUploadBytes)
		if err != nil {
			return row, err
		}
		if att == nil {
			att, err = parseSingleAttachmentFromMultipart(r, "soporte", maxSoporteComprasIAUploadBytes)
			if err != nil {
				return row, err
			}
		}
		if att == nil {
			att, err = parseSingleAttachmentFromMultipart(r, "file", maxSoporteComprasIAUploadBytes)
			if err != nil {
				return row, err
			}
		}
		row = soporteComprasIAFromForm(r, row)
		if att != nil {
			// Scan the raw multipart bytes before format-specific parsing. This lets
			// the official EICAR probe be rejected even though it is intentionally
			// not a valid business document, and ensures no untrusted bytes reach
			// a parser or storage when antivirus is required.
			if err := scanSoporteComprasIAAttachment(att); err != nil {
				return row, err
			}
			if err := validateSoporteComprasIAAttachment(att); err != nil {
				return row, err
			}
			release, err := acquireSoporteComprasIAStorageLock(r, dbEmp, empresaID)
			if err != nil {
				return row, err
			}
			defer release()
			if err := validateSoporteComprasIAStorageUpload(dbEmp, dbSuper, empresaID, int64(len(att.Bytes))); err != nil {
				return row, err
			}
			url, name, mimeType, hash, origen, err := saveSoporteComprasIAAttachment(att, empresaID)
			if err != nil {
				return row, err
			}
			row.ArchivoURL = url
			row.ArchivoNombre = name
			row.ArchivoMime = mimeType
			row.ArchivoHash = hash
			row.Origen = origen
		}
		created, err := dbpkg.CreateEmpresaSoporteComprasIA(dbEmp, row)
		if err != nil {
			cleanupUnpersistedSoporteComprasIAAttachment(row.ArchivoURL)
			return row, err
		}
		return created, nil
	}

	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&row); err != nil {
		return row, err
	}
	row.EmpresaID = empresaID
	row.Usuario = usuario
	row.ModeloIA = dbpkg.EmpresaSoporteComprasIAModeloDefault
	row = sanitizeManualSoporteComprasIA(row)
	if row.EstadoSoporte == "" {
		row.EstadoSoporte = "radicado"
	}
	return dbpkg.CreateEmpresaSoporteComprasIA(dbEmp, row)
}

func sanitizeManualSoporteComprasIA(row dbpkg.EmpresaSoporteComprasIA) dbpkg.EmpresaSoporteComprasIA {
	row.ArchivoURL = ""
	row.ArchivoNombre = ""
	row.ArchivoMime = ""
	row.ArchivoHash = ""
	row.Origen = "manual"
	return row
}

func validateSoporteComprasIAStorageUpload(dbEmp, dbSuper *sql.DB, empresaID, attachmentBytes int64) error {
	return validateSoporteComprasIAStorageUploadWithUsage(
		getEmpresaStorageConfig(dbSuper),
		buildEmpresaStorageUsage(dbEmp, dbSuper, empresaID),
		attachmentBytes,
	)
}

func validateSoporteComprasIAStorageUploadWithUsage(cfg empresaStorageConfig, usage empresaStorageUsage, attachmentBytes int64) error {
	if attachmentBytes <= 0 {
		return nil
	}
	maxBytes := int64(maxSoporteComprasIAUploadBytes)
	if configuredMax := cfg.MaxUploadMB * 1024 * 1024; configuredMax > 0 && configuredMax < maxBytes {
		maxBytes = configuredMax
	}
	if attachmentBytes > maxBytes {
		return fmt.Errorf("el soporte supera el maximo de almacenamiento permitido")
	}
	if cfg.QuotaEnabled && cfg.BlockUploads && usage.LimitBytes > 0 && usage.UsedBytes+attachmentBytes > usage.LimitBytes {
		return errSoporteComprasIAStorageQuota
	}
	return nil
}

func validateSoporteComprasIAAttachment(att *aiAttachment) error {
	if att == nil || len(att.Bytes) == 0 {
		return errors.New("el archivo del soporte esta vacio")
	}
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(att.Filename)))
	if !soporteComprasIAAllowedExt[ext] {
		return errors.New("extension no permitida para soporte")
	}
	detected := strings.ToLower(strings.TrimSpace(http.DetectContentType(att.Bytes)))
	switch ext {
	case ".png":
		if detected != "image/png" {
			return errors.New("el contenido no corresponde a una imagen PNG")
		}
		att.MimeType = "image/png"
	case ".jpg", ".jpeg":
		if detected != "image/jpeg" {
			return errors.New("el contenido no corresponde a una imagen JPEG")
		}
		att.MimeType = "image/jpeg"
	case ".webp":
		if detected != "image/webp" {
			return errors.New("el contenido no corresponde a una imagen WebP")
		}
		att.MimeType = "image/webp"
	case ".pdf":
		if !bytes.HasPrefix(bytes.TrimSpace(att.Bytes), []byte("%PDF-")) {
			return errors.New("el contenido no corresponde a un PDF")
		}
		att.MimeType = "application/pdf"
	case ".xml":
		if err := validateSoporteComprasIAXML(att.Bytes); err != nil {
			return err
		}
		att.MimeType = "application/xml"
	default:
		return errors.New("tipo de soporte no permitido")
	}
	return nil
}

func validateSoporteComprasIAXML(raw []byte) error {
	if len(bytes.TrimSpace(raw)) == 0 || !bytes.HasPrefix(bytes.TrimSpace(raw), []byte("<")) || bytes.IndexByte(raw, 0) >= 0 {
		return errors.New("el XML del soporte es invalido")
	}
	decoder := xml.NewDecoder(bytes.NewReader(raw))
	decoder.Strict = true
	depth := 0
	tokens := 0
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return errors.New("el XML del soporte no esta bien formado")
		}
		tokens++
		if tokens > 250000 {
			return errors.New("el XML del soporte supera la complejidad permitida")
		}
		switch value := token.(type) {
		case xml.Directive:
			return errors.New("el XML no permite DTD, entidades ni directivas")
		case xml.ProcInst:
			if !strings.EqualFold(strings.TrimSpace(value.Target), "xml") {
				return errors.New("el XML no permite instrucciones de procesamiento")
			}
		case xml.StartElement:
			depth++
			if depth > 128 {
				return errors.New("el XML supera la profundidad permitida")
			}
			switch strings.ToLower(strings.TrimSpace(value.Name.Local)) {
			case "script", "iframe", "object", "embed", "applet":
				return errors.New("el XML contiene elementos activos no permitidos")
			}
		case xml.EndElement:
			depth--
			if depth < 0 {
				return errors.New("el XML del soporte no esta bien formado")
			}
		}
	}
	if depth != 0 || tokens == 0 {
		return errors.New("el XML del soporte no esta bien formado")
	}
	return nil
}

func scanSoporteComprasIAAttachment(att *aiAttachment) (scanErr error) {
	addr := strings.TrimSpace(os.Getenv("PCS_SUPPORTS_CLAMAV_ADDR"))
	required := parseBoolSoporteComprasIA(os.Getenv("PCS_SUPPORTS_CLAMAV_REQUIRED"))
	defer func() {
		switch {
		case addr == "" && !required:
			soporteComprasIAAntivirusBypassed.Add(1)
		case scanErr == nil:
			soporteComprasIAAntivirusClean.Add(1)
		case errors.Is(scanErr, errSoporteComprasIAMalware):
			soporteComprasIAAntivirusMalware.Add(1)
		default:
			soporteComprasIAAntivirusUnavailable.Add(1)
		}
	}()
	if addr == "" {
		if required {
			return errSoporteComprasIAAntivirusUnavailable
		}
		return nil
	}
	if att == nil || len(att.Bytes) == 0 {
		return errors.New("el archivo del soporte esta vacio")
	}
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		return fmt.Errorf("%w: conexion no disponible", errSoporteComprasIAAntivirusUnavailable)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(15 * time.Second))
	if _, err := conn.Write([]byte("zINSTREAM\x00")); err != nil {
		return fmt.Errorf("%w: no se pudo iniciar el analisis", errSoporteComprasIAAntivirusUnavailable)
	}
	for offset := 0; offset < len(att.Bytes); {
		end := offset + 32*1024
		if end > len(att.Bytes) {
			end = len(att.Bytes)
		}
		chunk := att.Bytes[offset:end]
		// #nosec G115 -- chunk is bounded above by 32 KiB immediately above.
		if err := binary.Write(conn, binary.BigEndian, uint32(len(chunk))); err != nil {
			return fmt.Errorf("%w: envio interrumpido", errSoporteComprasIAAntivirusUnavailable)
		}
		if _, err := conn.Write(chunk); err != nil {
			return fmt.Errorf("%w: envio interrumpido", errSoporteComprasIAAntivirusUnavailable)
		}
		offset = end
	}
	if err := binary.Write(conn, binary.BigEndian, uint32(0)); err != nil {
		return fmt.Errorf("%w: no se pudo finalizar el analisis", errSoporteComprasIAAntivirusUnavailable)
	}
	response, err := bufio.NewReader(io.LimitReader(conn, 4096)).ReadString(0)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: respuesta invalida", errSoporteComprasIAAntivirusUnavailable)
	}
	response = strings.TrimSpace(strings.TrimRight(response, "\x00"))
	upper := strings.ToUpper(response)
	if strings.Contains(upper, " FOUND") || strings.HasSuffix(upper, "FOUND") {
		return errSoporteComprasIAMalware
	}
	if !strings.HasSuffix(upper, "OK") {
		return fmt.Errorf("%w: resultado no concluyente", errSoporteComprasIAAntivirusUnavailable)
	}
	return nil
}

func cleanupUnpersistedSoporteComprasIAAttachment(privateURL string) {
	path, err := safeSoporteComprasIAPathFromURL(privateURL)
	if err != nil {
		return
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("[soportes_compras_ia] no se pudo limpiar adjunto sin fila: %v", err)
	}
}

type soporteComprasIAPurgeFile struct {
	original   string
	quarantine string
}

func prepareSoporteComprasIAPurgeFile(row dbpkg.EmpresaSoporteComprasIA) (soporteComprasIAPurgeFile, error) {
	if strings.TrimSpace(row.ArchivoURL) == "" {
		return soporteComprasIAPurgeFile{}, nil
	}
	path, err := safeSoporteComprasIAPathCandidateForEmpresa(row.ArchivoURL, row.EmpresaID)
	if err != nil {
		return soporteComprasIAPurgeFile{}, err
	}
	info, statErr := os.Lstat(path)
	if statErr == nil {
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return soporteComprasIAPurgeFile{}, errors.New("el archivo privado no es regular")
		}
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return soporteComprasIAPurgeFile{}, statErr
	} else {
		matches, globErr := filepath.Glob(path + ".purge-????????????????????????")
		if globErr != nil {
			return soporteComprasIAPurgeFile{}, globErr
		}
		valid := make([]string, 0, len(matches))
		for _, match := range matches {
			matchInfo, matchErr := os.Lstat(match)
			if matchErr == nil && matchInfo.Mode().IsRegular() && matchInfo.Mode()&os.ModeSymlink == 0 {
				valid = append(valid, match)
			}
		}
		if len(valid) == 1 {
			return soporteComprasIAPurgeFile{original: path, quarantine: valid[0]}, nil
		}
		if len(valid) > 1 {
			return soporteComprasIAPurgeFile{}, errors.New("existen varias cuarentenas para el soporte")
		}
		if strings.EqualFold(strings.TrimSpace(row.Estado), "purga_pendiente") {
			// La fase de archivo ya terminó y solo falta cerrar la transacción.
			return soporteComprasIAPurgeFile{}, nil
		}
		return soporteComprasIAPurgeFile{}, errors.New("archivo privado no disponible para depuracion")
	}
	random := make([]byte, 12)
	if _, err := rand.Read(random); err != nil {
		return soporteComprasIAPurgeFile{}, errors.New("no se pudo crear la cuarentena privada")
	}
	quarantine := path + ".purge-" + hex.EncodeToString(random)
	if err := os.Rename(path, quarantine); err != nil {
		return soporteComprasIAPurgeFile{}, err
	}
	return soporteComprasIAPurgeFile{original: path, quarantine: quarantine}, nil
}

func (f soporteComprasIAPurgeFile) rollback() error {
	if f.quarantine == "" {
		return nil
	}
	return os.Rename(f.quarantine, f.original)
}

func (f soporteComprasIAPurgeFile) commit() error {
	if f.quarantine == "" {
		return nil
	}
	if err := os.Remove(f.quarantine); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func parseSoporteComprasIARetentionDays(raw string) int {
	days, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || days < 1 || days > 3650 {
		return 90
	}
	return days
}

func soporteComprasIARetentionBytes(rows []dbpkg.EmpresaSoporteComprasIA) int64 {
	var total int64
	for i := range rows {
		path, err := safeSoporteComprasIAPathForEmpresa(rows[i].ArchivoURL, rows[i].EmpresaID)
		if err != nil {
			continue
		}
		info, err := os.Lstat(path)
		if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			continue
		}
		total += info.Size()
	}
	return total
}

func soporteComprasIAQuarantineStats(empresaID int64) (int, int64, error) {
	if empresaID <= 0 {
		return 0, 0, errors.New("empresa invalida para cuarentena")
	}
	tenantDir := filepath.Join(soporteComprasIAPrivateRoot(), "empresa_"+strconv.FormatInt(empresaID, 10))
	entries, err := os.ReadDir(tenantDir)
	if errors.Is(err, os.ErrNotExist) {
		return 0, 0, nil
	}
	if err != nil {
		return 0, 0, err
	}
	count := 0
	var total int64
	for _, entry := range entries {
		if entry.IsDir() || entry.Type()&os.ModeSymlink != 0 || !isSoporteComprasIAQuarantineName(entry.Name()) {
			continue
		}
		info, infoErr := entry.Info()
		if infoErr != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			continue
		}
		count++
		total += info.Size()
	}
	return count, total, nil
}

func isSoporteComprasIAQuarantineName(name string) bool {
	marker := strings.LastIndex(name, ".purge-")
	if marker <= 0 {
		return false
	}
	suffix := name[marker+len(".purge-"):]
	if len(suffix) != 24 {
		return false
	}
	_, err := hex.DecodeString(suffix)
	return err == nil
}

func parseSoporteComprasIAQuarantineThreshold(raw string) int {
	minutes, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || minutes < 5 || minutes > 1440 {
		return 15
	}
	return minutes
}

func countSoporteComprasIAStalePending(rows []dbpkg.EmpresaSoporteComprasIA, now time.Time, threshold time.Duration) int {
	if threshold <= 0 {
		threshold = 15 * time.Minute
	}
	count := 0
	for i := range rows {
		raw := strings.TrimSpace(rows[i].FechaActualizacion)
		if raw == "" {
			raw = strings.TrimSpace(rows[i].FechaCreacion)
		}
		parsed, ok := parseSoporteComprasIAOperationalTime(raw, now.Location())
		if ok && !parsed.After(now) && now.Sub(parsed) >= threshold {
			count++
		}
	}
	return count
}

func parseSoporteComprasIAOperationalTime(raw string, location *time.Location) (time.Time, bool) {
	if location == nil {
		location = time.UTC
	}
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02 15:04:05.999999999Z07:00", "2006-01-02 15:04:05Z07:00", "2006-01-02 15:04:05.999999999-07", "2006-01-02 15:04:05-07"} {
		if parsed, err := time.Parse(layout, strings.TrimSpace(raw)); err == nil {
			return parsed, true
		}
	}
	for _, layout := range []string{"2006-01-02 15:04:05.999999999", "2006-01-02 15:04:05", "2006-01-02"} {
		if parsed, err := time.ParseInLocation(layout, strings.TrimSpace(raw), location); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

// acquireSoporteComprasIAStorageLock serializa la comprobacion de cuota, la
// escritura privada y el registro del soporte para una empresa entre replicas.
// El candado vive en PostgreSQL, por lo que no depende de memoria del proceso.
func acquireSoporteComprasIAStorageLock(r *http.Request, dbEmp *sql.DB, empresaID int64) (func(), error) {
	return acquireSoporteComprasIAAdvisoryLock(r, dbEmp, soporteComprasIAStorageNamespace, empresaID, "cuota")
}

func acquireSoporteComprasIAPurgeLock(r *http.Request, dbEmp *sql.DB, empresaID int64) (func(), error) {
	return acquireSoporteComprasIAAdvisoryLock(r, dbEmp, soporteComprasIAPurgeNamespace, empresaID, "depuracion")
}

func acquireSoporteComprasIAAdvisoryLock(r *http.Request, dbEmp *sql.DB, namespace, empresaID int64, operation string) (func(), error) {
	if r == nil || dbEmp == nil || empresaID <= 0 || empresaID > math.MaxInt32 {
		return nil, errors.New("empresa invalida para lock de soportes")
	}
	conn, err := dbEmp.Conn(r.Context())
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	for {
		var locked bool
		err = conn.QueryRowContext(ctx, `SELECT pg_try_advisory_lock($1::integer,$2::integer)`, soporteComprasIAStorageNamespace, empresaID).Scan(&locked)
		if err != nil {
			_ = conn.Close()
			return nil, err
		}
		if locked {
			return func() {
				releaseCtx, releaseCancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer releaseCancel()
				var unlocked bool
				if err := conn.QueryRowContext(releaseCtx, `SELECT pg_advisory_unlock($1::integer,$2::integer)`, soporteComprasIAStorageNamespace, empresaID).Scan(&unlocked); err != nil || !unlocked {
					log.Printf("[soportes_compras_ia] no se pudo liberar lock de %s empresa_id=%d: %v", operation, empresaID, err)
				}
				_ = conn.Close()
			}, nil
		}
		select {
		case <-ctx.Done():
			_ = conn.Close()
			return nil, errSoporteComprasIAStorageBusy
		case <-time.After(25 * time.Millisecond):
		}
	}
}

func soporteComprasIAFromForm(r *http.Request, row dbpkg.EmpresaSoporteComprasIA) dbpkg.EmpresaSoporteComprasIA {
	row.TipoSoporte = strings.TrimSpace(r.FormValue("tipo_soporte"))
	row.DocumentoTipo = strings.TrimSpace(r.FormValue("documento_tipo"))
	row.DocumentoNumero = strings.TrimSpace(r.FormValue("documento_numero"))
	row.ProveedorNombre = strings.TrimSpace(r.FormValue("proveedor_nombre"))
	row.ProveedorNIT = strings.TrimSpace(r.FormValue("proveedor_nit"))
	row.FechaDocumento = strings.TrimSpace(r.FormValue("fecha_documento"))
	row.FechaVencimiento = strings.TrimSpace(r.FormValue("fecha_vencimiento"))
	row.CategoriaContable = strings.TrimSpace(r.FormValue("categoria_contable"))
	row.CentroCosto = strings.TrimSpace(r.FormValue("centro_costo"))
	row.Moneda = strings.TrimSpace(r.FormValue("moneda"))
	row.Observaciones = strings.TrimSpace(r.FormValue("observaciones"))
	row.ImpactaInventario = parseBoolSoporteComprasIA(r.FormValue("impacta_inventario"))
	row.Subtotal = parseFloatSoporteComprasIA(r.FormValue("subtotal"))
	row.ImpuestoIVA = parseFloatSoporteComprasIA(r.FormValue("impuesto_iva"))
	row.RetencionFuente = parseFloatSoporteComprasIA(r.FormValue("retencion_fuente"))
	row.RetencionICA = parseFloatSoporteComprasIA(r.FormValue("retencion_ica"))
	row.RetencionIVA = parseFloatSoporteComprasIA(r.FormValue("retencion_iva"))
	row.Total = parseFloatSoporteComprasIA(r.FormValue("total"))
	return row
}

func saveSoporteComprasIAAttachment(att *aiAttachment, empresaID int64) (string, string, string, string, string, error) {
	if att == nil || len(att.Bytes) == 0 {
		return "", "", "", "", "manual", nil
	}
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(att.Filename)))
	if ext == "" {
		ext = extFromSoporteComprasIAMime(att.MimeType)
	}
	if !soporteComprasIAAllowedExt[ext] {
		return "", "", "", "", "", fmt.Errorf("extension no permitida para soporte")
	}
	root := soporteComprasIAPrivateRoot()
	dir := filepath.Join(root, fmt.Sprintf("empresa_%d", empresaID))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", "", "", "", "", err
	}
	randomName := make([]byte, 32)
	if _, err := rand.Read(randomName); err != nil {
		return "", "", "", "", "", fmt.Errorf("no se pudo generar nombre privado: %w", err)
	}
	fileName := hex.EncodeToString(randomName) + ext
	absPath := filepath.Join(dir, fileName)
	if err := os.WriteFile(absPath, att.Bytes, 0o600); err != nil {
		return "", "", "", "", "", err
	}
	mimeType := mimeFromSoporteComprasIAExt(ext)
	origen := origenFromSoporteComprasIAExt(ext, mimeType)
	url := "private://soportes_compras_ia/" + fmt.Sprintf("empresa_%d", empresaID) + "/" + fileName
	return url, fileName, mimeType, dbpkg.EmpresaSoporteComprasIAHashBytes(att.Bytes), origen, nil
}

func soporteComprasIAPrivateRoot() string {
	if configured := strings.TrimSpace(os.Getenv("PCS_PRIVATE_STORAGE_DIR")); configured != "" {
		return filepath.Join(configured, "soportes_compras_ia")
	}
	return filepath.Join(resolveProjectRootDir(), "private_storage", "soportes_compras_ia")
}

func soporteComprasIADownloadURL(empresaID, soporteID int64) string {
	return "/api/empresa/soportes_compras_ia?empresa_id=" + strconv.FormatInt(empresaID, 10) + "&action=descargar&soporte_id=" + strconv.FormatInt(soporteID, 10)
}

func exposeSoporteComprasIAURL(row *dbpkg.EmpresaSoporteComprasIA) {
	if row != nil && !strings.EqualFold(strings.TrimSpace(row.Estado), "activo") {
		row.ArchivoURL = ""
		return
	}
	if row != nil && row.ID > 0 && strings.HasPrefix(strings.TrimSpace(row.ArchivoURL), "private://") {
		row.ArchivoURL = soporteComprasIADownloadURL(row.EmpresaID, row.ID)
	}
}

func downloadSoporteComprasIA(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, empresaID int64) {
	soporteID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("soporte_id")), 10, 64)
	if err != nil || soporteID <= 0 {
		http.Error(w, "soporte invalido", http.StatusBadRequest)
		return
	}
	row, err := dbpkg.GetEmpresaSoporteComprasIAActivo(dbEmp, empresaID, soporteID)
	if err != nil {
		http.Error(w, "archivo no disponible", http.StatusNotFound)
		return
	}
	path, err := safeSoporteComprasIAPathForEmpresa(row.ArchivoURL, row.EmpresaID)
	if err != nil {
		http.Error(w, "archivo no disponible", http.StatusNotFound)
		return
	}
	file, err := os.Open(path) // #nosec G304 -- path was resolved under the private tenant root without symlinks.
	if err != nil {
		http.Error(w, "archivo no disponible", http.StatusNotFound)
		return
	}
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		http.Error(w, "archivo no disponible", http.StatusNotFound)
		return
	}
	if info.Size() < 1 || info.Size() > maxSoporteComprasIAUploadBytes {
		_ = supportFileIntegrityFailure()
		recordSoporteComprasIAIntegrityIncident(dbEmp, row, adminEmailFromRequest(r), "descarga")
		http.Error(w, "integridad del archivo no valida", http.StatusConflict)
		return
	}
	if err := verifySoporteComprasIAReaderIntegrity(file, row.ArchivoHash); err != nil {
		recordSoporteComprasIAIntegrityIncident(dbEmp, row, adminEmailFromRequest(r), "descarga")
		http.Error(w, "integridad del archivo no valida", http.StatusConflict)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "sandbox; default-src 'none'")
	w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Frame-Options", "DENY")
	downloadMIME := "application/octet-stream"
	if strings.TrimSpace(row.ArchivoHash) != "" {
		downloadMIME = safeSoporteComprasIADownloadMIME(row.ArchivoMime, path)
	}
	w.Header().Set("Content-Type", downloadMIME)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", sanitizeComprobanteBaseName(row.ArchivoNombre)+filepath.Ext(path)))
	http.ServeContent(w, r, filepath.Base(path), info.ModTime(), file)
}

var (
	errSoporteComprasIAIADesactivada        = errors.New("la IA esta desactivada desde configuracion avanzada")
	errSoporteComprasIAModeloNoDisponible   = errors.New("el modelo openai:gpt-5.5 no esta disponible en el catalogo de IA")
	errSoporteComprasIAProcesamientoEnCurso = errors.New("el soporte ya se esta procesando con IA")
	errSoporteComprasIAProveedor            = errors.New("fallo del proveedor IA para soportes de compras")
	errSoporteComprasIASolicitudInvalida    = errors.New("solicitud de extraccion IA invalida")
	errSoporteComprasIAEstadoNoExtraible    = errors.New("estado de soporte no extraible")
	errSoporteComprasIASinAdjunto           = errors.New("soporte sin adjunto privado disponible")
	errSoporteComprasIAStorageQuota         = errors.New("limite de almacenamiento empresarial alcanzado")
	errSoporteComprasIAStorageBusy          = errors.New("otra carga de almacenamiento empresarial esta en curso")
	errSoporteComprasIAMalware              = errors.New("el antivirus rechazo el archivo del soporte")
	errSoporteComprasIAAntivirusUnavailable = errors.New("el antivirus de soportes no esta disponible")
	errSoporteComprasIAIntegrity            = errors.New("la integridad del archivo del soporte no es valida")
)

func extraerSoporteComprasIAGPT55(r *http.Request, dbEmp, dbSuper *sql.DB, empresaID int64, usuario string) (dbpkg.EmpresaSoporteComprasIA, error) {
	if !isSuperAIEnabled(dbSuper) {
		return dbpkg.EmpresaSoporteComprasIA{}, errSoporteComprasIAIADesactivada
	}
	model, ok := availableEmpresaAIModelMap(dbSuper)[dbpkg.EmpresaSoporteComprasIAModeloDefault]
	if !ok {
		return dbpkg.EmpresaSoporteComprasIA{}, errSoporteComprasIAModeloNoDisponible
	}
	payload, err := decodeSoporteComprasIAActionPayload(r)
	if err != nil && !errors.Is(err, io.EOF) {
		return dbpkg.EmpresaSoporteComprasIA{}, errSoporteComprasIASolicitudInvalida
	}
	if payload.SoporteID <= 0 {
		payload.SoporteID, _ = strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("soporte_id")), 10, 64)
	}
	if payload.SoporteID <= 0 {
		return dbpkg.EmpresaSoporteComprasIA{}, errSoporteComprasIASolicitudInvalida
	}
	current, err := dbpkg.GetEmpresaSoporteComprasIAActivo(dbEmp, empresaID, payload.SoporteID)
	if err != nil {
		return dbpkg.EmpresaSoporteComprasIA{}, err
	}
	switch strings.ToLower(strings.TrimSpace(current.EstadoSoporte)) {
	case "radicado", "extraido", "en_revision":
	default:
		return dbpkg.EmpresaSoporteComprasIA{}, errSoporteComprasIAEstadoNoExtraible
	}
	att, err := loadSoporteComprasIAAttachment(current)
	if err != nil {
		if errors.Is(err, errSoporteComprasIAIntegrity) {
			recordSoporteComprasIAIntegrityIncident(dbEmp, current, usuario, "extraccion_ia")
		}
		return dbpkg.EmpresaSoporteComprasIA{}, fmt.Errorf("%w: %v", errSoporteComprasIASinAdjunto, err)
	}
	release, err := acquireSoporteComprasIAExtractionLock(r, dbEmp, empresaID, payload.SoporteID)
	if err != nil {
		return dbpkg.EmpresaSoporteComprasIA{}, err
	}
	defer release()
	if _, _, err := reserveEmpresaAgentAdvancedUsage(dbEmp, dbSuper, empresaID, usuario); err != nil {
		return dbpkg.EmpresaSoporteComprasIA{}, err
	}

	ctrl := NewEmpresaAIChatController(dbEmp, dbSuper)
	systemPrompt := soporteComprasIASystemPrompt()
	pregunta := "Extrae y normaliza este soporte de compra o gasto de Colombia. Responde solo JSON valido, sin explicaciones."
	providerCtx, cancelProvider := context.WithTimeout(r.Context(), soporteComprasIAExtractionTimeout)
	defer cancelProvider()
	respuesta, promptTokens, completionTokens, err := ctrl.callOpenAIResponsesWithSystemPromptContext(providerCtx, model, pregunta, nil, systemPrompt, att, nil, nil)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			outcome := "canceled"
			if errors.Is(err, context.DeadlineExceeded) {
				outcome = "timeout"
			}
			recordSupportExtractionOutcome(outcome)
			if refundErr := dbpkg.RefundEmpresaAgenteUsoDiario(dbEmp, dbpkg.EmpresaAgenteUsoDiario{
				EmpresaID: empresaID, FechaUso: time.Now().Format("2006-01-02"), ConsultasAvanzadas: 1, SegundosUsados: 5,
			}); refundErr != nil {
				log.Printf("[soportes_compras_ia] no se pudo devolver reserva IA interrumpida empresa_id=%d resultado=%s: %v", empresaID, outcome, refundErr)
			}
		} else {
			recordSupportExtractionOutcome("provider_error")
		}
		return dbpkg.EmpresaSoporteComprasIA{}, fmt.Errorf("%w: %w", errSoporteComprasIAProveedor, err)
	}
	extracted, compactJSON, err := parseSoporteComprasIAExtraction(respuesta)
	if err != nil {
		recordSupportExtractionOutcome("invalid_response")
		return dbpkg.EmpresaSoporteComprasIA{}, fmt.Errorf("%w: respuesta invalida", errSoporteComprasIAProveedor)
	}
	extracted = evaluateSoporteComprasIAExtraction(extracted)
	extracted.ExtraccionJSON = compactJSON
	extracted.RespuestaIA = respuesta
	extracted.ModeloIA = model.ID
	extracted.Usuario = usuario
	updated, err := dbpkg.UpdateEmpresaSoporteComprasIAExtraccion(dbEmp, empresaID, payload.SoporteID, extracted, usuario)
	if err != nil {
		recordSupportExtractionOutcome("persistence_error")
		return dbpkg.EmpresaSoporteComprasIA{}, err
	}
	if extracted.RequiereRevisionHumana {
		recordSupportExtractionOutcome("human_review")
	} else {
		recordSupportExtractionOutcome("consistent")
	}
	_, _ = dbpkg.RegisterEmpresaAIConsulta(dbEmp, dbpkg.EmpresaAIConsulta{
		EmpresaID:        empresaID,
		Provider:         model.Provider,
		ModelID:          model.ID,
		Pregunta:         fmt.Sprintf("captura_inteligente_compras_gastos soporte_id=%d codigo=%s", payload.SoporteID, current.Codigo),
		Respuesta:        respuesta,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
		UsuarioCreador:   usuario,
		Estado:           "activo",
		Observaciones:    "Extraccion IA GPT-5.5 de soporte de compra o gasto",
	})
	return updated, nil
}

func acquireSoporteComprasIAExtractionLock(r *http.Request, dbEmp *sql.DB, empresaID, soporteID int64) (func(), error) {
	if dbEmp == nil || empresaID <= 0 || soporteID <= 0 || empresaID > math.MaxInt32 || soporteID > math.MaxInt32 {
		return nil, errors.New("identificador de soporte no valido")
	}
	conn, err := dbEmp.Conn(r.Context())
	if err != nil {
		return nil, err
	}
	var locked bool
	if err := conn.QueryRowContext(r.Context(), `SELECT pg_try_advisory_lock($1::integer,$2::integer)`, soporteComprasIAAdvisoryNamespace, soporteID).Scan(&locked); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if !locked {
		_ = conn.Close()
		return nil, errSoporteComprasIAProcesamientoEnCurso
	}
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		var unlocked bool
		if err := conn.QueryRowContext(ctx, `SELECT pg_advisory_unlock($1::integer,$2::integer)`, soporteComprasIAAdvisoryNamespace, soporteID).Scan(&unlocked); err != nil || !unlocked {
			log.Printf("[soportes_compras_ia] no se pudo liberar lock empresa_id=%d soporte_id=%d: %v", empresaID, soporteID, err)
		}
		_ = conn.Close()
	}, nil
}

func evaluateSoporteComprasIAExtraction(row dbpkg.EmpresaSoporteComprasIA) dbpkg.EmpresaSoporteComprasIA {
	expected := row.Subtotal + row.ImpuestoIVA - row.RetencionFuente - row.RetencionICA - row.RetencionIVA
	tolerance := math.Max(1, math.Abs(expected)*0.01)
	invalidMoney := false
	for _, value := range []float64{row.Subtotal, row.ImpuestoIVA, row.RetencionFuente, row.RetencionICA, row.RetencionIVA, row.Total} {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
			invalidMoney = true
			break
		}
	}
	if strings.TrimSpace(row.ProveedorNombre) == "" || strings.TrimSpace(row.DocumentoNumero) == "" ||
		strings.TrimSpace(row.DocumentoTipo) == "" || strings.TrimSpace(row.FechaDocumento) == "" ||
		!strings.EqualFold(strings.TrimSpace(row.Moneda), "COP") || row.Total <= 0 ||
		row.ConfianzaIA < 0.85 || row.ConfianzaIA > 1 || math.IsNaN(row.ConfianzaIA) || math.IsInf(row.ConfianzaIA, 0) ||
		invalidMoney || math.IsNaN(expected) || math.IsInf(expected, 0) || math.Abs(row.Total-expected) > tolerance {
		row.RequiereRevisionHumana = true
	}
	return row
}

func soporteComprasIASystemPrompt() string {
	return `Eres un motor profesional de captura inteligente con IA GPT-5.5 para documentos empresariales en Colombia.
Lee fotos, PDFs o XML de facturas de compra, documentos soporte, cuentas de cobro, recibos, gastos, ingresos, comprobantes de caja y cartas/listas de precios.
Devuelve exclusivamente JSON valido con estas claves:
{
  "tipo_soporte": "compra|gasto|ingreso|documento_soporte|recibo|servicio|carta_precios",
  "proveedor_nombre": "",
  "proveedor_nit": "",
  "documento_tipo": "factura_compra|documento_soporte|cuenta_cobro|recibo_caja|gasto|ingreso|lista_precios|otro",
  "documento_numero": "",
  "fecha_documento": "YYYY-MM-DD",
  "fecha_vencimiento": "YYYY-MM-DD",
  "subtotal": 0,
  "impuesto_iva": 0,
  "retencion_fuente": 0,
  "retencion_ica": 0,
  "retencion_iva": 0,
  "total": 0,
  "moneda": "COP",
  "categoria_contable": "",
  "centro_costo": "",
  "impacta_inventario": false,
  "confianza_ia": 0.0,
  "requiere_revision_humana": true,
  "lineas_detectadas": [],
  "observaciones": ""
}
Usa numeros sin separadores de miles. Si falta un dato, deja cadena vacia o 0. Marca requiere_revision_humana=true cuando haya baja confianza, documento borroso, totales inconsistentes o datos tributarios incompletos.`
}

func parseSoporteComprasIAExtraction(raw string) (dbpkg.EmpresaSoporteComprasIA, string, error) {
	if len(raw) == 0 || len(raw) > maxSoporteComprasIAExtractionBytes {
		return dbpkg.EmpresaSoporteComprasIA{}, "", errors.New("respuesta IA fuera del limite permitido")
	}
	candidate := extractJSONCandidate(raw)
	if len(candidate) == 0 || len(candidate) > maxSoporteComprasIAExtractionBytes {
		return dbpkg.EmpresaSoporteComprasIA{}, "", errors.New("JSON IA fuera del limite permitido")
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(candidate), &data); err != nil {
		return dbpkg.EmpresaSoporteComprasIA{}, "", fmt.Errorf("respuesta IA no es JSON valido: %w", err)
	}
	if err := validateSoporteComprasIAExtractionSchema(data); err != nil {
		return dbpkg.EmpresaSoporteComprasIA{}, "", err
	}
	compactBytes, _ := json.Marshal(data)
	row := dbpkg.EmpresaSoporteComprasIA{
		TipoSoporte:            stringFromMap(data, "tipo_soporte"),
		ProveedorNombre:        stringFromMap(data, "proveedor_nombre"),
		ProveedorNIT:           stringFromMap(data, "proveedor_nit"),
		DocumentoTipo:          stringFromMap(data, "documento_tipo"),
		DocumentoNumero:        stringFromMap(data, "documento_numero"),
		FechaDocumento:         normalizeDateString(stringFromMap(data, "fecha_documento")),
		FechaVencimiento:       normalizeDateString(stringFromMap(data, "fecha_vencimiento")),
		Subtotal:               floatFromMap(data, "subtotal"),
		ImpuestoIVA:            floatFromMap(data, "impuesto_iva"),
		RetencionFuente:        floatFromMap(data, "retencion_fuente"),
		RetencionICA:           floatFromMap(data, "retencion_ica"),
		RetencionIVA:           floatFromMap(data, "retencion_iva"),
		Total:                  floatFromMap(data, "total"),
		Moneda:                 stringFromMap(data, "moneda"),
		CategoriaContable:      stringFromMap(data, "categoria_contable"),
		CentroCosto:            stringFromMap(data, "centro_costo"),
		ImpactaInventario:      boolFromMap(data, "impacta_inventario"),
		ConfianzaIA:            floatFromMap(data, "confianza_ia"),
		RequiereRevisionHumana: boolFromMap(data, "requiere_revision_humana"),
		Observaciones:          stringFromMap(data, "observaciones"),
	}
	if err := validateSoporteComprasIAExtractionValues(row); err != nil {
		return dbpkg.EmpresaSoporteComprasIA{}, "", err
	}
	return row, string(compactBytes), nil
}

func validateSoporteComprasIAExtractionSchema(data map[string]interface{}) error {
	allowed := map[string]struct{}{
		"tipo_soporte": {}, "proveedor_nombre": {}, "proveedor_nit": {},
		"documento_tipo": {}, "documento_numero": {}, "fecha_documento": {},
		"fecha_vencimiento": {}, "subtotal": {}, "impuesto_iva": {},
		"retencion_fuente": {}, "retencion_ica": {}, "retencion_iva": {},
		"total": {}, "moneda": {}, "categoria_contable": {}, "centro_costo": {},
		"impacta_inventario": {}, "confianza_ia": {}, "requiere_revision_humana": {},
		"lineas_detectadas": {}, "observaciones": {},
	}
	for key := range data {
		if _, ok := allowed[key]; !ok {
			return fmt.Errorf("respuesta IA contiene campo no permitido: %s", key)
		}
	}
	for key, value := range data {
		if key == "lineas_detectadas" {
			if value == nil {
				continue
			}
			lines, ok := value.([]interface{})
			if !ok || len(lines) > 500 {
				return errors.New("respuesta IA contiene lineas_detectadas invalidas")
			}
			continue
		}
		switch value.(type) {
		case nil, string, float64, bool:
		default:
			return fmt.Errorf("respuesta IA contiene tipo invalido para %s", key)
		}
	}
	return nil
}

func validateSoporteComprasIAExtractionValues(row dbpkg.EmpresaSoporteComprasIA) error {
	textLimits := map[string]struct {
		value string
		limit int
	}{
		"tipo_soporte": {row.TipoSoporte, 40}, "proveedor_nombre": {row.ProveedorNombre, 200},
		"proveedor_nit": {row.ProveedorNIT, 40}, "documento_tipo": {row.DocumentoTipo, 50},
		"documento_numero": {row.DocumentoNumero, 100}, "moneda": {row.Moneda, 8},
		"categoria_contable": {row.CategoriaContable, 160}, "centro_costo": {row.CentroCosto, 160},
		"observaciones": {row.Observaciones, 2000},
	}
	for field, item := range textLimits {
		if len(item.value) > item.limit {
			return fmt.Errorf("respuesta IA excede limite de %s", field)
		}
	}
	for _, value := range []float64{row.Subtotal, row.ImpuestoIVA, row.RetencionFuente, row.RetencionICA, row.RetencionIVA, row.Total} {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 1e15 {
			return errors.New("respuesta IA contiene valor monetario invalido")
		}
	}
	if math.IsNaN(row.ConfianzaIA) || math.IsInf(row.ConfianzaIA, 0) || row.ConfianzaIA < 0 || row.ConfianzaIA > 1 {
		return errors.New("respuesta IA contiene confianza invalida")
	}
	return nil
}

func extractJSONCandidate(raw string) string {
	clean := strings.TrimSpace(raw)
	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimPrefix(clean, "```")
	clean = strings.TrimSuffix(clean, "```")
	clean = strings.TrimSpace(clean)
	start := strings.Index(clean, "{")
	end := strings.LastIndex(clean, "}")
	if start >= 0 && end > start {
		return clean[start : end+1]
	}
	return clean
}

func loadSoporteComprasIAAttachment(row dbpkg.EmpresaSoporteComprasIA) (*aiAttachment, error) {
	if strings.TrimSpace(row.ArchivoURL) == "" {
		return nil, errors.New("el soporte no tiene archivo adjunto para analisis IA")
	}
	path, err := safeSoporteComprasIAPathForEmpresa(row.ArchivoURL, row.EmpresaID)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path) // #nosec G304 -- path is an existing regular file resolved inside the private support root above.
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() < 1 || info.Size() > maxSoporteComprasIAUploadBytes {
		return nil, supportFileIntegrityFailure()
	}
	b, err := io.ReadAll(io.LimitReader(file, maxSoporteComprasIAUploadBytes+1))
	if err != nil || len(b) < 1 || len(b) > maxSoporteComprasIAUploadBytes {
		return nil, supportFileIntegrityFailure()
	}
	if err := verifySoporteComprasIABytesIntegrity(b, row.ArchivoHash); err != nil {
		return nil, err
	}
	att := &aiAttachment{Filename: row.ArchivoNombre, MimeType: row.ArchivoMime, Bytes: b}
	if err := validateSoporteComprasIAAttachment(att); err != nil {
		return nil, supportFileIntegrityFailure()
	}
	return att, nil
}

func verifySoporteComprasIAReaderIntegrity(file *os.File, expectedHash string) error {
	if file == nil {
		return supportFileIntegrityFailure()
	}
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return supportFileIntegrityFailure()
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return supportFileIntegrityFailure()
	}
	return verifySoporteComprasIADigest(hasher.Sum(nil), expectedHash)
}

func verifySoporteComprasIABytesIntegrity(content []byte, expectedHash string) error {
	sum := sha256.Sum256(content)
	return verifySoporteComprasIADigest(sum[:], expectedHash)
}

func verifySoporteComprasIADigest(actual []byte, expectedHash string) error {
	expectedHash = strings.ToLower(strings.TrimSpace(expectedHash))
	if expectedHash == "" {
		return nil
	}
	expected, err := hex.DecodeString(expectedHash)
	if err != nil || len(expected) != sha256.Size || len(actual) != sha256.Size || subtle.ConstantTimeCompare(actual, expected) != 1 {
		return supportFileIntegrityFailure()
	}
	return nil
}

func supportFileIntegrityFailure() error {
	soporteComprasIAFileIntegrityFailures.Add(1)
	return errSoporteComprasIAIntegrity
}

func recordSoporteComprasIAIntegrityIncident(dbEmp *sql.DB, row dbpkg.EmpresaSoporteComprasIA, actor, channel string) {
	if dbEmp == nil || row.EmpresaID <= 0 || row.ID <= 0 {
		return
	}
	if err := dbpkg.MarkEmpresaSoporteComprasIAIntegrityIncident(dbEmp, row.EmpresaID, row.ID,
		strings.TrimSpace(actor), strings.TrimSpace(channel)); err != nil {
		log.Printf("[soportes_compras_ia] auditoria integridad no disponible empresa_id=%d soporte_id=%d", row.EmpresaID, row.ID)
	}
}

func safeSoporteComprasIADownloadMIME(storedMIME, path string) string {
	storedMIME = strings.ToLower(strings.TrimSpace(storedMIME))
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png":
		if storedMIME == "image/png" {
			return storedMIME
		}
	case ".jpg", ".jpeg":
		if storedMIME == "image/jpeg" {
			return storedMIME
		}
	case ".webp":
		if storedMIME == "image/webp" {
			return storedMIME
		}
	case ".pdf":
		if storedMIME == "application/pdf" {
			return storedMIME
		}
	case ".xml":
		if storedMIME == "application/xml" || storedMIME == "text/xml" {
			return "application/xml"
		}
	}
	return "application/octet-stream"
}

func safeSoporteComprasIAPathFromURL(url string) (string, error) {
	clean := strings.TrimSpace(strings.TrimPrefix(url, "private://soportes_compras_ia/"))
	if clean == strings.TrimSpace(url) || clean == "" {
		return "", errors.New("ruta de soporte no permitida")
	}
	root := filepath.Clean(soporteComprasIAPrivateRoot())
	abs := filepath.Clean(filepath.Join(root, filepath.FromSlash(clean)))
	if abs != root && !strings.HasPrefix(abs, root+string(os.PathSeparator)) {
		return "", errors.New("ruta de soporte fuera del directorio permitido")
	}
	return resolveExistingPrivateFileUnderRoot(root, abs)
}

func safeSoporteComprasIAPathForEmpresa(url string, empresaID int64) (string, error) {
	candidate, err := safeSoporteComprasIAPathCandidateForEmpresa(url, empresaID)
	if err != nil {
		return "", err
	}
	return resolveExistingPrivateFileUnderRoot(soporteComprasIAPrivateRoot(), candidate)
}

func safeSoporteComprasIAPathCandidateForEmpresa(url string, empresaID int64) (string, error) {
	if empresaID <= 0 {
		return "", errors.New("empresa invalida para archivo privado")
	}
	prefix := "private://soportes_compras_ia/empresa_" + strconv.FormatInt(empresaID, 10) + "/"
	if !strings.HasPrefix(strings.TrimSpace(url), prefix) {
		return "", errors.New("el archivo privado no pertenece a la empresa")
	}
	clean := strings.TrimPrefix(strings.TrimSpace(url), "private://soportes_compras_ia/")
	root, err := filepath.Abs(filepath.Clean(soporteComprasIAPrivateRoot()))
	if err != nil {
		return "", errors.New("directorio privado no disponible")
	}
	candidate, err := filepath.Abs(filepath.Clean(filepath.Join(root, filepath.FromSlash(clean))))
	if err != nil {
		return "", errors.New("archivo privado no disponible")
	}
	rel, err := filepath.Rel(root, candidate)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || filepath.IsAbs(rel) {
		return "", errors.New("ruta de soporte fuera del directorio permitido")
	}
	parentResolved, err := filepath.EvalSymlinks(filepath.Dir(candidate))
	if err != nil {
		return "", errors.New("directorio privado no disponible")
	}
	rootResolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", errors.New("directorio privado no disponible")
	}
	resolvedRel, err := filepath.Rel(rootResolved, parentResolved)
	if err != nil || resolvedRel == ".." || strings.HasPrefix(resolvedRel, ".."+string(os.PathSeparator)) || filepath.IsAbs(resolvedRel) {
		return "", errors.New("directorio privado fuera del tenant")
	}
	return candidate, nil
}

// resolveExistingPrivateFileUnderRoot rejects traversal and symlink escapes before
// an attachment is sent to an external analysis service.
func resolveExistingPrivateFileUnderRoot(root, candidate string) (string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", errors.New("directorio privado no disponible")
	}
	absCandidate, err := filepath.Abs(candidate)
	if err != nil {
		return "", errors.New("archivo privado no disponible")
	}
	rel, err := filepath.Rel(absRoot, absCandidate)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || filepath.IsAbs(rel) {
		return "", errors.New("ruta de soporte fuera del directorio permitido")
	}
	info, err := os.Lstat(absCandidate)
	if err != nil || !info.Mode().IsRegular() {
		return "", errors.New("archivo de soporte no disponible")
	}
	resolvedRoot, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		return "", errors.New("directorio privado no disponible")
	}
	resolvedCandidate, err := filepath.EvalSymlinks(absCandidate)
	if err != nil {
		return "", errors.New("archivo de soporte no disponible")
	}
	resolvedRel, err := filepath.Rel(resolvedRoot, resolvedCandidate)
	if err != nil || strings.HasPrefix(resolvedRel, ".."+string(os.PathSeparator)) || resolvedRel == ".." || filepath.IsAbs(resolvedRel) {
		return "", errors.New("enlace de soporte fuera del directorio permitido")
	}
	return resolvedCandidate, nil
}

func seedSoporteComprasIADemo(dbEmp *sql.DB, empresaID int64, usuario string) (dbpkg.EmpresaSoporteComprasIA, error) {
	row, err := dbpkg.CreateEmpresaSoporteComprasIA(dbEmp, dbpkg.EmpresaSoporteComprasIA{
		EmpresaID:              empresaID,
		TipoSoporte:            "gasto",
		Origen:                 "manual",
		ProveedorNombre:        "Papeleria Centro Empresarial SAS",
		ProveedorNIT:           "901234567-8",
		DocumentoTipo:          "factura_compra",
		DocumentoNumero:        "FE-1024",
		FechaDocumento:         time.Now().Format("2006-01-02"),
		FechaVencimiento:       time.Now().AddDate(0, 0, 15).Format("2006-01-02"),
		Subtotal:               180000,
		ImpuestoIVA:            34200,
		Total:                  214200,
		Moneda:                 "COP",
		CategoriaContable:      "Gastos administrativos",
		CentroCosto:            "Administracion",
		ConfianzaIA:            0.94,
		ModeloIA:               dbpkg.EmpresaSoporteComprasIAModeloDefault,
		RequiereRevisionHumana: true,
		Usuario:                usuario,
		Observaciones:          "Soporte de ejemplo para probar captura inteligente.",
	})
	if err != nil {
		return row, err
	}
	extraction := row
	raw, _ := json.Marshal(map[string]interface{}{
		"proveedor_nombre": row.ProveedorNombre,
		"proveedor_nit":    row.ProveedorNIT,
		"documento_tipo":   row.DocumentoTipo,
		"documento_numero": row.DocumentoNumero,
		"total":            row.Total,
		"confianza_ia":     row.ConfianzaIA,
	})
	extraction.ExtraccionJSON = string(raw)
	return dbpkg.UpdateEmpresaSoporteComprasIAExtraccion(dbEmp, empresaID, row.ID, extraction, usuario)
}

func extFromSoporteComprasIAMime(mimeType string) string {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	case "application/pdf":
		return ".pdf"
	case "application/xml", "text/xml":
		return ".xml"
	default:
		return ".bin"
	}
}

func mimeFromSoporteComprasIAExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".pdf":
		return "application/pdf"
	case ".xml":
		return "application/xml"
	default:
		return "application/octet-stream"
	}
}

func origenFromSoporteComprasIAExt(ext, mimeType string) string {
	if strings.HasPrefix(strings.ToLower(mimeType), "image/") {
		return "foto"
	}
	switch strings.ToLower(ext) {
	case ".pdf":
		return "pdf"
	case ".xml":
		return "xml"
	default:
		return "manual"
	}
}

func parseFloatSoporteComprasIA(raw string) float64 {
	raw = strings.ReplaceAll(strings.TrimSpace(raw), ",", ".")
	v, _ := strconv.ParseFloat(raw, 64)
	return v
}

func parseBoolSoporteComprasIA(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "on", "si", "sí", "yes":
		return true
	default:
		return false
	}
}

func stringFromMap(data map[string]interface{}, key string) string {
	if data == nil {
		return ""
	}
	raw, ok := data[key]
	if !ok || raw == nil {
		return ""
	}
	switch v := raw.(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func floatFromMap(data map[string]interface{}, key string) float64 {
	if data == nil {
		return 0
	}
	switch v := data[key].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case string:
		return parseFloatSoporteComprasIA(v)
	default:
		return 0
	}
}

func boolFromMap(data map[string]interface{}, key string) bool {
	if data == nil {
		return false
	}
	switch v := data[key].(type) {
	case bool:
		return v
	case string:
		return parseBoolSoporteComprasIA(v)
	case float64:
		return v != 0
	default:
		return false
	}
}

func normalizeDateString(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	layouts := []string{"2006-01-02", "02/01/2006", "02-01-2006", time.RFC3339}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.Format("2006-01-02")
		}
	}
	if len(raw) >= 10 {
		if t, err := time.Parse("2006-01-02", raw[:10]); err == nil {
			return t.Format("2006-01-02")
		}
	}
	return ""
}
