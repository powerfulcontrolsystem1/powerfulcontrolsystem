package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type queueCapacityConfigCacheEntry struct {
	config    dbpkg.QueueCapacityConfig
	expiresAt time.Time
}

var queueCapacityConfigCache = struct {
	sync.Mutex
	items map[string]queueCapacityConfigCacheEntry
}{items: map[string]queueCapacityConfigCacheEntry{}}

func defaultQueueRateLimitForLane(lane string) int64 {
	switch lane {
	case dbpkg.QueueLanePrinting:
		return 120
	case dbpkg.QueueLaneProductAdd:
		return 240
	case dbpkg.QueueLaneFiscal:
		return 30
	default:
		return defaultEmpresaAPIRequestsPerMinute
	}
}

func queueCapacityRateLimit(dbSuper *sql.DB, lane string) int64 {
	fallback := defaultQueueRateLimitForLane(lane)
	cfg, err := getQueueCapacityConfigCached(dbSuper, lane)
	if err != nil || cfg.RateLimitPerMinute <= 0 {
		return fallback
	}
	return cfg.RateLimitPerMinute
}

func getQueueCapacityConfigCached(dbSuper *sql.DB, lane string) (dbpkg.QueueCapacityConfig, error) {
	if dbSuper == nil {
		return dbpkg.QueueCapacityConfig{Lane: lane, RateLimitPerMinute: defaultQueueRateLimitForLane(lane)}, nil
	}
	now := time.Now()
	queueCapacityConfigCache.Lock()
	entry, ok := queueCapacityConfigCache.items[lane]
	queueCapacityConfigCache.Unlock()
	if ok && now.Before(entry.expiresAt) && entry.config.RateLimitPerMinute > 0 {
		return entry.config, nil
	}
	cfg, err := dbpkg.GetQueueCapacityConfig(dbSuper, lane)
	if err != nil {
		log.Printf("[queue_capacity] no se pudo leer limite lane=%s error_type=%T", lane, err)
		return dbpkg.QueueCapacityConfig{}, err
	}
	queueCapacityConfigCache.Lock()
	queueCapacityConfigCache.items[lane] = queueCapacityConfigCacheEntry{config: cfg, expiresAt: now.Add(dbpkg.QueueCapacityConfigCacheTTL())}
	queueCapacityConfigCache.Unlock()
	return cfg, nil
}

func invalidateQueueCapacityConfigCache() {
	queueCapacityConfigCache.Lock()
	queueCapacityConfigCache.items = map[string]queueCapacityConfigCacheEntry{}
	queueCapacityConfigCache.Unlock()
}

type queueCapacityEvaluation struct {
	OK          bool                          `json:"ok"`
	Configs     []dbpkg.QueueCapacityConfig   `json:"configs"`
	Snapshots   []dbpkg.QueueCapacitySnapshot `json:"snapshots"`
	Candidates  []superAlertCandidate         `json:"candidates"`
	EvaluatedAt string                        `json:"evaluated_at"`
	Error       string                        `json:"error,omitempty"`
}

type queueCapacitySaveRequest struct {
	Configs []dbpkg.QueueCapacityConfig `json:"configs"`
}

// SuperQueueCapacityHandler exposes aggregate, super-only queue health. It
// does not return job payloads, document contents or tenant credentials.
func SuperQueueCapacityHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminEmail, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper)
		if !ok {
			return
		}
		switch r.Method {
		case http.MethodGet:
			eval := EvaluateQueueCapacity(dbEmp, dbSuper, false)
			status := http.StatusOK
			if !eval.OK {
				status = http.StatusServiceUnavailable
			}
			writeJSON(w, status, eval)
		case http.MethodPut:
			var request queueCapacitySaveRequest
			decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&request); err != nil {
				http.Error(w, "configuracion de colas invalida", http.StatusBadRequest)
				return
			}
			if err := dbpkg.SaveQueueCapacityConfigs(r.Context(), dbSuper, request.Configs, adminEmail); err != nil {
				log.Printf("[queue_capacity] guardar request_id=%s error_type=%T", resolveAuditoriaRequestID(r), err)
				http.Error(w, "No se pudo guardar la configuracion de capacidad", http.StatusBadRequest)
				return
			}
			invalidateQueueCapacityConfigCache()
			configs, err := dbpkg.GetQueueCapacityConfigs(dbSuper)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "No se pudo recargar la configuracion"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "configs": configs})
		case http.MethodPost:
			if strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action"))) != "evaluate" {
				http.Error(w, "action no soportada", http.StatusBadRequest)
				return
			}
			eval := EvaluateQueueCapacity(dbEmp, dbSuper, true)
			status := http.StatusOK
			if !eval.OK {
				status = http.StatusServiceUnavailable
			}
			writeJSON(w, status, eval)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func EvaluateQueueCapacity(dbEmp, dbSuper *sql.DB, forceSend bool) queueCapacityEvaluation {
	eval := queueCapacityEvaluation{OK: true, EvaluatedAt: time.Now().Format("2006-01-02 15:04:05"), Candidates: []superAlertCandidate{}}
	if dbSuper == nil {
		eval.OK = false
		eval.Error = "base super administrador no disponible"
		return eval
	}
	configs, err := dbpkg.GetQueueCapacityConfigs(dbSuper)
	if err != nil {
		eval.OK = false
		eval.Error = "configuracion de capacidad no disponible"
		return eval
	}
	eval.Configs = configs
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	eval.Snapshots = dbpkg.GetQueueCapacitySnapshots(ctx, dbEmp, dbSuper, configs)
	byLane := map[string]dbpkg.QueueCapacityConfig{}
	for _, cfg := range configs {
		byLane[cfg.Lane] = cfg
	}
	alertConfig, alertErr := dbpkg.GetSuperAlertasConfig(dbSuper)
	for _, snapshot := range eval.Snapshots {
		cfg := byLane[snapshot.Lane]
		candidate := queueCapacityCandidate(snapshot, cfg)
		if !snapshot.QueryOK {
			eval.OK = false
			candidate.Severidad = "warning"
			candidate.Detalle = snapshot.Error
			eval.Candidates = append(eval.Candidates, candidate)
			continue
		}
		if candidate.Triggered && cfg.AlertsEnabled && alertErr == nil && alertConfig.Enabled {
			dispatchQueueCapacityCandidate(dbSuper, alertConfig, snapshot, &candidate, eval.EvaluatedAt, forceSend)
		}
		eval.Candidates = append(eval.Candidates, candidate)
	}
	if !eval.OK {
		eval.Error = "Una o mas colas no pudieron medirse; revisa migraciones y conectividad PostgreSQL."
	}
	return eval
}

func queueCapacityCandidate(snapshot dbpkg.QueueCapacitySnapshot, cfg dbpkg.QueueCapacityConfig) superAlertCandidate {
	title := "Capacidad normal: " + snapshot.Label
	detail := fmt.Sprintf("Pendientes=%d, procesando=%d, fallidos=%d, empresas activas=%d, antiguedad=%.0fs.",
		snapshot.Pending, snapshot.Processing, snapshot.Failed, snapshot.ActiveTenants, snapshot.OldestSeconds)
	triggered := false
	if snapshot.Lane == dbpkg.QueueLaneProductAdd {
		title = "Presion del carril para agregar productos"
		detail = fmt.Sprintf("Solicitudes del minuto=%d, maximo de una empresa=%d, limite por empresa=%d/min.",
			snapshot.RequestsCurrentMinute, snapshot.BusiestTenantPending, cfg.RateLimitPerMinute)
		triggered = cfg.RateLimitPerMinute > 0 && snapshot.BusiestTenantPending >= cfg.RateLimitPerMinute
	} else {
		triggered = (cfg.PendingAlertThreshold > 0 && snapshot.Pending >= cfg.PendingAlertThreshold) ||
			(cfg.OldestAlertSeconds > 0 && snapshot.OldestSeconds >= float64(cfg.OldestAlertSeconds)) ||
			(cfg.MaxPendingPerTenant > 0 && snapshot.BusiestTenantPending >= cfg.MaxPendingPerTenant)
		if triggered {
			title = "Saturacion de cola: " + snapshot.Label
		}
	}
	return superAlertCandidate{
		Tipo: "queue_capacity_" + snapshot.Lane, Severidad: severityForQueueSaturation(snapshot.SaturationPercent),
		Titulo: title, Detalle: detail, Valor: math.Round(snapshot.SaturationPercent*100) / 100,
		Umbral: 100, Unidad: "%", Triggered: triggered,
	}
}

func severityForQueueSaturation(percent float64) string {
	if percent >= 150 {
		return "critical"
	}
	if percent >= 100 {
		return "warning"
	}
	return "info"
}

func dispatchQueueCapacityCandidate(dbSuper *sql.DB, alertCfg dbpkg.SuperAlertaConfig, snapshot dbpkg.QueueCapacitySnapshot, candidate *superAlertCandidate, evaluatedAt string, forceSend bool) {
	if candidate == nil {
		return
	}
	if !forceSend {
		if recent, err := dbpkg.SuperAlertaRecentlySent(dbSuper, candidate.Tipo, alertCfg.CooldownMinutes); err == nil && recent {
			candidate.SkippedCooldown = true
			return
		}
	}
	subject := "[PCS] " + candidate.Titulo
	body := fmt.Sprintf("Alerta de capacidad de colas de Powerful Control System.\r\n\r\n%s\r\nSaturacion: %.2f%%\r\nEmpresa con mayor presion: %d (%d trabajos/solicitudes)\r\nEvaluada: %s\r\n\r\nRevisa Super administrador > Capacidad de colas.",
		candidate.Detalle, snapshot.SaturationPercent, snapshot.BusiestTenantID, snapshot.BusiestTenantPending, evaluatedAt)
	sent, sendErr := sendSuperSystemAlertEmail(dbSuper, alertCfg.RecipientEmail, subject, body, candidate.Tipo, "sistema")
	candidate.CorreoEnviado = sent
	if sendErr != nil {
		candidate.CorreoError = "no se pudo entregar la alerta por el canal configurado"
	}
	metadata, _ := json.Marshal(map[string]interface{}{
		"lane": snapshot.Lane, "pending": snapshot.Pending, "processing": snapshot.Processing,
		"failed": snapshot.Failed, "active_tenants": snapshot.ActiveTenants,
		"busiest_tenant_id": snapshot.BusiestTenantID, "busiest_tenant_pending": snapshot.BusiestTenantPending,
		"oldest_seconds": snapshot.OldestSeconds, "requests_current_minute": snapshot.RequestsCurrentMinute,
	})
	event := dbpkg.SuperAlertaEvento{Tipo: candidate.Tipo, Severidad: candidate.Severidad, Titulo: candidate.Titulo,
		Detalle: candidate.Detalle, Valor: candidate.Valor, Umbral: candidate.Umbral, Unidad: candidate.Unidad,
		Destinatario: alertCfg.RecipientEmail, Asunto: subject, Cuerpo: body, CorreoEnviado: sent,
		MetadataJSON: string(metadata), UsuarioCreador: "sistema", Observaciones: "evaluacion_capacidad_colas"}
	if sendErr != nil {
		event.CorreoError = "entrega de alerta fallida"
	}
	if _, err := dbpkg.CreateSuperAlertaEvento(dbSuper, event); err != nil {
		log.Printf("[queue_capacity] registrar alerta lane=%s error_type=%T", snapshot.Lane, err)
	}
}
