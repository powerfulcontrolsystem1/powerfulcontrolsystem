package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type controlElectricoCommandPayload struct {
	EmpresaID      int64  `json:"empresa_id"`
	EstacionID     int64  `json:"estacion_id"`
	RaspberryID    int64  `json:"raspberry_id,omitempty"`
	EstacionCodigo string `json:"estacion_codigo,omitempty"`
	EstacionNombre string `json:"estacion_nombre,omitempty"`
	RelayID        int64  `json:"relay_id"`
	SalidaCodigo   string `json:"salida_codigo,omitempty"`
	TipoCarga      string `json:"tipo_carga,omitempty"`
	RelayName      string `json:"relay_name,omitempty"`
	GPIOPin        int    `json:"gpio_pin"`
	Estado         string `json:"estado"`
	ActiveHigh     bool   `json:"active_high"`
	PulsoMS        int    `json:"pulso_ms"`
	Origen         string `json:"origen,omitempty"`
	Actor          string `json:"actor,omitempty"`
}

type controlElectricoDispatchResult struct {
	OK           bool   `json:"ok"`
	Skipped      bool   `json:"skipped,omitempty"`
	Message      string `json:"message,omitempty"`
	HTTPStatus   int    `json:"http_status,omitempty"`
	ResponseBody string `json:"response_body,omitempty"`
	Error        string `json:"error,omitempty"`
	URL          string `json:"url,omitempty"`
}

// EmpresaControlElectricoHandler administra el modulo de control electrico por Raspberry Pi.
func EmpresaControlElectricoHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := dbpkg.EnsureEmpresaControlElectricoSchema(dbEmp); err != nil {
			log.Printf("[control_electrico] ensure schema error: %v", err)
			http.Error(w, "No se pudo preparar control electrico", http.StatusInternalServerError)
			return
		}
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "", "resumen":
				cfg, err := dbpkg.GetEmpresaControlElectricoConfig(dbEmp, empresaID, true)
				if err != nil {
					log.Printf("[control_electrico] get config empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo cargar configuracion electrica", http.StatusInternalServerError)
					return
				}
				if _, err := dbpkg.EnsureEmpresaControlElectricoPrimaryRaspberry(dbEmp, cfg); err != nil {
					log.Printf("[control_electrico] ensure primary raspberry empresa_id=%d error: %v", empresaID, err)
				}
				cfg.APIToken = ""
				estaciones, err := dbpkg.ListEmpresaControlElectricoEstaciones(dbEmp, empresaID)
				if err != nil {
					log.Printf("[control_electrico] list estaciones empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudieron cargar estaciones electricas", http.StatusInternalServerError)
					return
				}
				eventos, err := dbpkg.ListEmpresaControlElectricoEventos(dbEmp, empresaID, 25)
				if err != nil {
					log.Printf("[control_electrico] list eventos empresa_id=%d error: %v", empresaID, err)
					eventos = []dbpkg.EmpresaControlElectricoEvento{}
				}
				reles, err := dbpkg.ListEmpresaControlElectricoReles(dbEmp, empresaID, true)
				if err != nil {
					log.Printf("[control_electrico] list reles resumen empresa_id=%d error: %v", empresaID, err)
					reles = []dbpkg.EmpresaControlElectricoRele{}
				}
				raspberries, err := dbpkg.ListEmpresaControlElectricoRaspberry(dbEmp, empresaID, true)
				if err != nil {
					log.Printf("[control_electrico] list raspberry resumen empresa_id=%d error: %v", empresaID, err)
					raspberries = []dbpkg.EmpresaControlElectricoRaspberry{}
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"config":        cfg,
					"raspberry_pis": raspberries,
					"estaciones":    estaciones,
					"reles":         reles,
					"eventos":       eventos,
				})
				return
			case "config":
				cfg, err := dbpkg.GetEmpresaControlElectricoConfig(dbEmp, empresaID, false)
				if err != nil {
					log.Printf("[control_electrico] get config empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo cargar configuracion electrica", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, cfg)
				return
			case "reles":
				reles, err := dbpkg.ListEmpresaControlElectricoReles(dbEmp, empresaID, controlElectricoIncludeInactive(r))
				if err != nil {
					log.Printf("[control_electrico] list reles empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudieron cargar reles", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, reles)
				return
			case "raspberry_pis":
				rows, err := dbpkg.ListEmpresaControlElectricoRaspberry(dbEmp, empresaID, controlElectricoIncludeInactive(r))
				if err != nil {
					log.Printf("[control_electrico] list raspberry_pis empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudieron cargar Raspberry Pi", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "estacion_controls":
				estacionID, err := parseInt64QueryOptional(r, "estacion_id")
				if err != nil || estacionID <= 0 {
					http.Error(w, "estacion_id requerido", http.StatusBadRequest)
					return
				}
				cfg, err := dbpkg.GetEmpresaControlElectricoConfig(dbEmp, empresaID, false)
				if err != nil {
					log.Printf("[control_electrico] get station config empresa_id=%d estacion_id=%d error: %v", empresaID, estacionID, err)
					http.Error(w, "No se pudo cargar configuracion electrica", http.StatusInternalServerError)
					return
				}
				reles, err := dbpkg.ListEmpresaControlElectricoRelesByEstacion(dbEmp, empresaID, estacionID, false)
				if err != nil {
					log.Printf("[control_electrico] list station reles empresa_id=%d estacion_id=%d error: %v", empresaID, estacionID, err)
					http.Error(w, "No se pudieron cargar controles electricos de la estacion", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"config":      cfg,
					"estacion_id": estacionID,
					"reles":       reles,
				})
				return
			case "eventos":
				limit := controlElectricoParseLimit(r, 50)
				eventos, err := dbpkg.ListEmpresaControlElectricoEventos(dbEmp, empresaID, limit)
				if err != nil {
					log.Printf("[control_electrico] list eventos empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudieron cargar eventos electricos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, eventos)
				return
			default:
				http.Error(w, "action no soportada", http.StatusBadRequest)
				return
			}

		case http.MethodPost, http.MethodPut:
			switch action {
			case "config":
				var payload dbpkg.EmpresaControlElectricoConfig
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				id, err := dbpkg.UpsertEmpresaControlElectricoConfig(dbEmp, &payload)
				if err != nil {
					log.Printf("[control_electrico] upsert config empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo guardar configuracion electrica", http.StatusInternalServerError)
					return
				}
				cfg, _ := dbpkg.GetEmpresaControlElectricoConfig(dbEmp, empresaID, false)
				if cfgWithToken, err := dbpkg.GetEmpresaControlElectricoConfig(dbEmp, empresaID, true); err == nil {
					_, _ = dbpkg.EnsureEmpresaControlElectricoPrimaryRaspberry(dbEmp, cfgWithToken)
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id, "config": cfg})
				return

			case "raspberry_pi":
				var payload dbpkg.EmpresaControlElectricoRaspberry
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				id, err := dbpkg.UpsertEmpresaControlElectricoRaspberry(dbEmp, &payload)
				if err != nil {
					log.Printf("[control_electrico] upsert raspberry empresa_id=%d error: %v", empresaID, err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				rows, _ := dbpkg.ListEmpresaControlElectricoRaspberry(dbEmp, empresaID, true)
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id, "raspberry_pis": rows})
				return

			case "rele":
				var payload dbpkg.EmpresaControlElectricoRele
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				id, err := dbpkg.UpsertEmpresaControlElectricoRele(dbEmp, &payload)
				if err != nil {
					log.Printf("[control_electrico] upsert rele empresa_id=%d estacion_id=%d error: %v", empresaID, payload.EstacionID, err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return

			case "rele_foto":
				releID, imageURL, err := handleControlElectricoReleFotoUpload(r, dbEmp, empresaID)
				if err != nil {
					log.Printf("[control_electrico] upload foto empresa_id=%d error: %v", empresaID, err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "rele_id": releID, "image_url": imageURL})
				return

			case "probar_rele":
				var payload struct {
					EstacionID int64  `json:"estacion_id"`
					ReleID     int64  `json:"rele_id"`
					Estado     string `json:"estado"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				target, err := controlElectricoParseTargetState(payload.Estado)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				result := controlElectricoDispatchManual(dbEmp, empresaID, payload.EstacionID, payload.ReleID, target, strings.TrimSpace(adminEmailFromRequest(r)), "prueba_manual")
				status := http.StatusOK
				if !result.OK && !result.Skipped {
					status = http.StatusBadGateway
				}
				writeJSON(w, status, result)
				return

			case "sincronizar":
				estaciones, err := dbpkg.ListEmpresaControlElectricoEstaciones(dbEmp, empresaID)
				if err != nil {
					log.Printf("[control_electrico] sync list estaciones empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudieron cargar estaciones para sincronizar", http.StatusInternalServerError)
					return
				}
				results := make([]map[string]interface{}, 0, len(estaciones))
				actor := strings.TrimSpace(adminEmailFromRequest(r))
				for _, estacion := range estaciones {
					result := DispatchEmpresaControlElectricoEstacion(dbEmp, empresaID, estacion.EstacionID, estacion.Activa, actor, "sincronizacion_manual")
					results = append(results, map[string]interface{}{
						"estacion_id": estacion.EstacionID,
						"activa":      estacion.Activa,
						"resultado":   result,
					})
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "results": results})
				return
			case "ejecutar_programacion":
				executed, err := ejecutarControlElectricoProgramacionPendiente(dbEmp, time.Now(), empresaID)
				if err != nil {
					log.Printf("[control_electrico] ejecutar programacion empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo evaluar la programacion electrica", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "comandos_ejecutados": executed})
				return
			default:
				http.Error(w, "action no soportada", http.StatusBadRequest)
				return
			}

		case http.MethodDelete:
			if action == "raspberry_pi" {
				raspberryID, err := parseInt64QueryOptional(r, "id")
				if err != nil || raspberryID <= 0 {
					http.Error(w, "id requerido", http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetEmpresaControlElectricoRaspberryEstado(dbEmp, empresaID, raspberryID, "inactivo"); err != nil {
					log.Printf("[control_electrico] delete raspberry empresa_id=%d id=%d error: %v", empresaID, raspberryID, err)
					http.Error(w, "No se pudo desactivar Raspberry Pi", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			}
			releID, err := parseInt64QueryOptional(r, "id")
			if err != nil || releID <= 0 {
				http.Error(w, "id requerido", http.StatusBadRequest)
				return
			}
			if err := dbpkg.SetEmpresaControlElectricoReleEstado(dbEmp, empresaID, releID, "inactivo"); err != nil {
				log.Printf("[control_electrico] delete rele empresa_id=%d id=%d error: %v", empresaID, releID, err)
				http.Error(w, "No se pudo desactivar rele", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

func handleControlElectricoReleFotoUpload(r *http.Request, dbEmp *sql.DB, empresaID int64) (int64, string, error) {
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		return 0, "", fmt.Errorf("payload multipart invalido")
	}
	releID, err := parseInt64Form(r, "rele_id")
	if err != nil || releID <= 0 {
		return 0, "", fmt.Errorf("rele_id requerido")
	}
	if _, err := dbpkg.GetEmpresaControlElectricoReleByID(dbEmp, empresaID, releID); err != nil {
		if err == sql.ErrNoRows {
			return 0, "", fmt.Errorf("rele no encontrado")
		}
		return 0, "", err
	}
	file, header, err := r.FormFile("foto")
	if err != nil {
		return 0, "", fmt.Errorf("foto requerida")
	}
	defer file.Close()
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(header.Filename)))
	allowed := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true}
	if !allowed[ext] {
		return 0, "", fmt.Errorf("extension de imagen no permitida")
	}
	webRoot := resolveWebRootDir()
	dir := filepath.Join(webRoot, "uploads", "control_electrico", fmt.Sprintf("empresa_%d", empresaID))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, "", fmt.Errorf("no se pudo preparar carpeta de imagenes")
	}
	fileName := fmt.Sprintf("rele_%d_%d%s", releID, time.Now().UnixNano(), ext)
	absPath := filepath.Join(dir, fileName)
	out, err := os.Create(absPath)
	if err != nil {
		return 0, "", fmt.Errorf("no se pudo crear imagen")
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		return 0, "", fmt.Errorf("no se pudo guardar imagen")
	}
	imageURL := "/uploads/control_electrico/empresa_" + strconv.FormatInt(empresaID, 10) + "/" + fileName
	if err := dbpkg.UpdateEmpresaControlElectricoReleImagen(dbEmp, empresaID, releID, imageURL); err != nil {
		return 0, "", err
	}
	return releID, imageURL, nil
}

// StartControlElectricoProgramacionWorker ejecuta horarios ON/OFF de relays programados.
func StartControlElectricoProgramacionWorker(dbEmp *sql.DB, interval time.Duration, stop <-chan struct{}) {
	if interval <= 0 {
		interval = time.Minute
	}
	run := func(origin string) {
		count, err := EjecutarControlElectricoProgramacionPendiente(dbEmp, time.Now())
		if err != nil {
			log.Printf("[control_electrico] programacion %s error: %v", origin, err)
			return
		}
		if count > 0 {
			log.Printf("[control_electrico] programacion %s comandos_ejecutados=%d", origin, count)
		}
	}
	run("startup")
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			run("ticker")
		case <-stop:
			return
		}
	}
}

// EjecutarControlElectricoProgramacionPendiente evalua horarios activos y dispara comandos vencidos.
func EjecutarControlElectricoProgramacionPendiente(dbEmp *sql.DB, now time.Time) (int, error) {
	return ejecutarControlElectricoProgramacionPendiente(dbEmp, now, 0)
}

func ejecutarControlElectricoProgramacionPendiente(dbEmp *sql.DB, now time.Time, empresaIDFilter int64) (int, error) {
	if dbEmp == nil {
		return 0, fmt.Errorf("dbEmp nil")
	}
	if err := dbpkg.EnsureEmpresaControlElectricoSchema(dbEmp); err != nil {
		return 0, err
	}
	reles, err := dbpkg.ListEmpresaControlElectricoRelesProgramados(dbEmp)
	if err != nil {
		return 0, err
	}
	executed := 0
	for i := range reles {
		rele := reles[i]
		if empresaIDFilter > 0 && rele.EmpresaID != empresaIDFilter {
			continue
		}
		for _, due := range controlElectricoProgramacionDue(rele, now) {
			cfg, err := dbpkg.GetEmpresaControlElectricoConfig(dbEmp, rele.EmpresaID, true)
			if err != nil {
				log.Printf("[control_electrico] config programacion empresa_id=%d rele_id=%d error: %v", rele.EmpresaID, rele.ID, err)
				continue
			}
			if cfg == nil || !cfg.Habilitado {
				continue
			}
			result := dispatchControlElectricoRele(dbEmp, cfg, &rele, due.EstadoObjetivo, "sistema.control_electrico", "programacion_horaria")
			if result.OK {
				if err := dbpkg.MarkEmpresaControlElectricoReleProgramacion(dbEmp, rele.EmpresaID, rele.ID, due.EstadoObjetivo, due.EjecutadoEn); err != nil {
					log.Printf("[control_electrico] marcar programacion empresa_id=%d rele_id=%d error: %v", rele.EmpresaID, rele.ID, err)
				}
				executed++
			}
		}
	}
	return executed, nil
}

type controlElectricoProgramacionDueItem struct {
	EstadoObjetivo string
	EjecutadoEn    string
}

func controlElectricoProgramacionDue(rele dbpkg.EmpresaControlElectricoRele, now time.Time) []controlElectricoProgramacionDueItem {
	if !rele.ProgramacionHabilitada {
		return nil
	}
	loc, err := time.LoadLocation(strings.TrimSpace(rele.ProgramacionTimezone))
	if err != nil {
		loc, _ = time.LoadLocation("America/Bogota")
	}
	local := now.In(loc)
	if !controlElectricoProgramacionDiaActivo(rele.ProgramacionDias, local.Weekday()) {
		return nil
	}
	currentMinute := local.Format("15:04")
	currentStamp := local.Format("2006-01-02 15:04:05")
	out := []controlElectricoProgramacionDueItem{}
	if currentMinute == strings.TrimSpace(rele.HoraEncendido) && !controlElectricoProgramacionEjecutadaHoy(rele.UltimaProgramacionOn, local) {
		out = append(out, controlElectricoProgramacionDueItem{EstadoObjetivo: "on", EjecutadoEn: currentStamp})
	}
	if currentMinute == strings.TrimSpace(rele.HoraApagado) && !controlElectricoProgramacionEjecutadaHoy(rele.UltimaProgramacionOff, local) {
		out = append(out, controlElectricoProgramacionDueItem{EstadoObjetivo: "off", EjecutadoEn: currentStamp})
	}
	return out
}

func controlElectricoProgramacionDiaActivo(raw string, weekday time.Weekday) bool {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "", "todos", "diario", "daily":
		return true
	case "lunes_viernes", "laborales", "weekdays":
		return weekday >= time.Monday && weekday <= time.Friday
	case "sabado_domingo", "fines_semana", "weekend":
		return weekday == time.Saturday || weekday == time.Sunday
	}
	target := strconv.Itoa(int(weekday))
	for _, part := range strings.Split(value, ",") {
		if strings.TrimSpace(part) == target {
			return true
		}
	}
	return false
}

func controlElectricoProgramacionEjecutadaHoy(raw string, local time.Time) bool {
	value := strings.TrimSpace(raw)
	if len(value) >= 10 {
		return value[:10] == local.Format("2006-01-02")
	}
	return false
}

// DispatchEmpresaControlElectricoEstacion envia ON/OFF a la Raspberry Pi para una estacion.
func DispatchEmpresaControlElectricoEstacion(dbEmp *sql.DB, empresaID, estacionID int64, activa bool, actor, origen string) controlElectricoDispatchResult {
	if dbEmp == nil || empresaID <= 0 || estacionID <= 0 {
		return controlElectricoDispatchResult{Skipped: true, Message: "estacion no valida"}
	}
	if err := dbpkg.EnsureEmpresaControlElectricoSchema(dbEmp); err != nil {
		return controlElectricoDispatchResult{OK: false, Error: err.Error()}
	}
	cfg, err := dbpkg.GetEmpresaControlElectricoConfig(dbEmp, empresaID, true)
	if err != nil {
		return controlElectricoDispatchResult{OK: false, Error: err.Error()}
	}
	if cfg == nil || !cfg.Habilitado {
		return controlElectricoDispatchResult{Skipped: true, Message: "control electrico no habilitado"}
	}
	if !cfg.AutoSyncEstaciones && !controlElectricoOrigenManual(origen) {
		return controlElectricoDispatchResult{Skipped: true, Message: "sincronizacion automatica desactivada"}
	}
	reles, err := dbpkg.ListEmpresaControlElectricoRelesByEstacion(dbEmp, empresaID, estacionID, false)
	if err != nil {
		return controlElectricoDispatchResult{OK: false, Error: err.Error()}
	}
	if len(reles) == 0 {
		return controlElectricoDispatchResult{Skipped: true, Message: "estacion sin rele configurado"}
	}
	targetState := "off"
	if activa {
		targetState = "on"
	}
	total := 0
	failed := 0
	skipped := 0
	lastResult := controlElectricoDispatchResult{OK: true}
	for i := range reles {
		rele := reles[i]
		if !strings.EqualFold(strings.TrimSpace(rele.Modo), "seguimiento_estacion") {
			skipped++
			continue
		}
		total++
		result := dispatchControlElectricoRele(dbEmp, cfg, &rele, targetState, actor, origen)
		lastResult = result
		if !result.OK {
			failed++
		}
	}
	if total == 0 {
		return controlElectricoDispatchResult{Skipped: true, Message: "estacion sin salidas automaticas"}
	}
	if failed > 0 {
		lastResult.Message = fmt.Sprintf("sincronizacion parcial: %d/%d salidas con error", failed, total)
		return lastResult
	}
	return controlElectricoDispatchResult{OK: true, Message: fmt.Sprintf("salidas sincronizadas: %d, omitidas: %d", total, skipped), URL: lastResult.URL}
}

func controlElectricoDispatchManual(dbEmp *sql.DB, empresaID, estacionID, releID int64, activa bool, actor, origen string) controlElectricoDispatchResult {
	if dbEmp == nil || empresaID <= 0 {
		return controlElectricoDispatchResult{Skipped: true, Message: "empresa no valida"}
	}
	cfg, err := dbpkg.GetEmpresaControlElectricoConfig(dbEmp, empresaID, true)
	if err != nil {
		return controlElectricoDispatchResult{OK: false, Error: err.Error()}
	}
	if cfg == nil || !cfg.Habilitado {
		return controlElectricoDispatchResult{Skipped: true, Message: "control electrico no habilitado"}
	}
	var rele *dbpkg.EmpresaControlElectricoRele
	if releID > 0 {
		rele, err = dbpkg.GetEmpresaControlElectricoReleByID(dbEmp, empresaID, releID)
	} else {
		rele, err = dbpkg.GetEmpresaControlElectricoReleByEstacion(dbEmp, empresaID, estacionID)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return controlElectricoDispatchResult{Skipped: true, Message: "rele no configurado"}
		}
		return controlElectricoDispatchResult{OK: false, Error: err.Error()}
	}
	targetState := "off"
	if activa {
		targetState = "on"
	}
	return dispatchControlElectricoRele(dbEmp, cfg, rele, targetState, actor, origen)
}

func dispatchControlElectricoRele(dbEmp *sql.DB, cfg *dbpkg.EmpresaControlElectricoConfig, rele *dbpkg.EmpresaControlElectricoRele, targetState, actor, origen string) controlElectricoDispatchResult {
	dispatchCfg, raspberryID, err := resolveControlElectricoDispatchConfig(dbEmp, cfg, rele)
	if err != nil {
		result := controlElectricoDispatchResult{OK: false, Error: err.Error()}
		_, _ = dbpkg.InsertEmpresaControlElectricoEvento(dbEmp, dbpkg.EmpresaControlElectricoEvento{
			EmpresaID:      cfg.EmpresaID,
			EstacionID:     rele.EstacionID,
			ReleID:         rele.ID,
			RaspberryID:    rele.RaspberryID,
			GPIOPin:        rele.GPIOPin,
			Comando:        "set_relay",
			EstadoObjetivo: targetState,
			Resultado:      "error",
			RaspberryIP:    rele.RaspberryIP,
			Error:          result.Error,
			Actor:          actor,
			Origen:         origen,
		})
		return result
	}
	result := sendControlElectricoRelayCommand(dispatchCfg, rele, targetState, strings.TrimSpace(actor), strings.TrimSpace(origen))
	evento := dbpkg.EmpresaControlElectricoEvento{
		EmpresaID:      dispatchCfg.EmpresaID,
		EstacionID:     rele.EstacionID,
		ReleID:         rele.ID,
		RaspberryID:    raspberryID,
		GPIOPin:        rele.GPIOPin,
		Comando:        "set_relay",
		EstadoObjetivo: targetState,
		Resultado:      "error",
		HTTPStatus:     result.HTTPStatus,
		RaspberryIP:    dispatchCfg.RaspberryIP,
		ResponseBody:   result.ResponseBody,
		Error:          result.Error,
		Actor:          actor,
		Origen:         origen,
	}
	if result.OK {
		evento.Resultado = "ok"
		_ = dbpkg.UpdateEmpresaControlElectricoReleRuntime(dbEmp, dispatchCfg.EmpresaID, rele.ID, targetState, "set_relay", "")
	} else {
		_ = dbpkg.UpdateEmpresaControlElectricoReleRuntime(dbEmp, dispatchCfg.EmpresaID, rele.ID, rele.UltimoEstado, "set_relay", result.Error)
	}
	if _, err := dbpkg.InsertEmpresaControlElectricoEvento(dbEmp, evento); err != nil {
		log.Printf("[control_electrico] insert evento empresa_id=%d estacion_id=%d error: %v", dispatchCfg.EmpresaID, rele.EstacionID, err)
	}
	return result
}

func resolveControlElectricoDispatchConfig(dbEmp *sql.DB, cfg *dbpkg.EmpresaControlElectricoConfig, rele *dbpkg.EmpresaControlElectricoRele) (*dbpkg.EmpresaControlElectricoConfig, int64, error) {
	if cfg == nil {
		return nil, 0, fmt.Errorf("configuracion electrica no disponible")
	}
	if rele != nil && rele.RaspberryID > 0 {
		pi, err := dbpkg.GetEmpresaControlElectricoRaspberryByID(dbEmp, cfg.EmpresaID, rele.RaspberryID, true)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, rele.RaspberryID, fmt.Errorf("Raspberry Pi asignada no esta activa")
			}
			return nil, rele.RaspberryID, err
		}
		return controlElectricoConfigFromRaspberry(cfg, pi), pi.ID, nil
	}
	if strings.TrimSpace(cfg.RaspberryIP) == "" {
		return nil, 0, fmt.Errorf("rele sin Raspberry Pi asignada y sin Raspberry principal")
	}
	return cfg, 0, nil
}

func controlElectricoConfigFromRaspberry(base *dbpkg.EmpresaControlElectricoConfig, pi *dbpkg.EmpresaControlElectricoRaspberry) *dbpkg.EmpresaControlElectricoConfig {
	out := *base
	out.RaspberryIP = pi.RaspberryIP
	out.RaspberryPort = pi.RaspberryPort
	out.APIPath = pi.APIPath
	out.APIToken = pi.APIToken
	out.TimeoutMS = pi.TimeoutMS
	return &out
}

func sendControlElectricoRelayCommand(cfg *dbpkg.EmpresaControlElectricoConfig, rele *dbpkg.EmpresaControlElectricoRele, estado, actor, origen string) controlElectricoDispatchResult {
	endpoint, err := buildControlElectricoEndpoint(cfg)
	if err != nil {
		return controlElectricoDispatchResult{OK: false, Error: err.Error()}
	}
	payload := controlElectricoCommandPayload{
		EmpresaID:      cfg.EmpresaID,
		EstacionID:     rele.EstacionID,
		RaspberryID:    rele.RaspberryID,
		EstacionCodigo: rele.EstacionCodigo,
		EstacionNombre: rele.EstacionNombre,
		RelayID:        rele.ID,
		SalidaCodigo:   rele.SalidaCodigo,
		TipoCarga:      rele.TipoCarga,
		RelayName:      rele.RelayName,
		GPIOPin:        rele.GPIOPin,
		Estado:         estado,
		ActiveHigh:     rele.ActiveHigh,
		PulsoMS:        rele.PulsoMS,
		Origen:         origen,
		Actor:          actor,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return controlElectricoDispatchResult{OK: false, Error: err.Error(), URL: endpoint}
	}
	timeout := time.Duration(cfg.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = time.Duration(dbpkg.DefaultControlElectricoTimeoutMS) * time.Millisecond
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return controlElectricoDispatchResult{OK: false, Error: err.Error(), URL: endpoint}
	}
	req.Header.Set("Content-Type", "application/json")
	if token := strings.TrimSpace(cfg.APIToken); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-Control-Electrico-Token", token)
	}
	resp, err := (&http.Client{Timeout: timeout}).Do(req)
	if err != nil {
		return controlElectricoDispatchResult{OK: false, Error: err.Error(), URL: endpoint}
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	result := controlElectricoDispatchResult{
		OK:           resp.StatusCode >= 200 && resp.StatusCode < 300,
		HTTPStatus:   resp.StatusCode,
		ResponseBody: strings.TrimSpace(string(raw)),
		URL:          endpoint,
	}
	if !result.OK {
		result.Error = fmt.Sprintf("raspberry respondio HTTP %d", resp.StatusCode)
	}
	return result
}

func buildControlElectricoEndpoint(cfg *dbpkg.EmpresaControlElectricoConfig) (string, error) {
	if cfg == nil || strings.TrimSpace(cfg.RaspberryIP) == "" {
		return "", fmt.Errorf("raspberry_ip obligatorio")
	}
	host := strings.TrimSpace(cfg.RaspberryIP)
	if strings.Contains(host, "://") {
		parsed, err := url.Parse(host)
		if err != nil {
			return "", err
		}
		if parsed.Path == "" || parsed.Path == "/" {
			parsed.Path = cfg.APIPath
		}
		return parsed.String(), nil
	}
	if _, _, err := net.SplitHostPort(host); err != nil {
		host = net.JoinHostPort(host, strconv.Itoa(cfg.RaspberryPort))
	}
	return (&url.URL{Scheme: "http", Host: host, Path: cfg.APIPath}).String(), nil
}

func dispatchControlElectricoEstacionAsync(dbEmp *sql.DB, carrito *dbpkg.CarritoCompra, activa bool, actor, origen string) {
	if dbEmp == nil || carrito == nil {
		return
	}
	estacionID, _, _ := dbpkg.ResolveCarritoStationIdentity(carrito)
	if estacionID <= 0 {
		return
	}
	empresaID := carrito.EmpresaID
	go func() {
		result := DispatchEmpresaControlElectricoEstacion(dbEmp, empresaID, estacionID, activa, actor, origen)
		if !result.OK && !result.Skipped {
			log.Printf("[control_electrico] dispatch async empresa_id=%d estacion_id=%d activa=%v origen=%s error=%s", empresaID, estacionID, activa, origen, result.Error)
		}
	}()
}

func controlElectricoParseTargetState(raw string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "on", "encender", "encendido", "activo", "activa", "abrir":
		return true, nil
	case "0", "false", "off", "apagar", "apagado", "inactivo", "inactiva", "cerrar":
		return false, nil
	default:
		return false, fmt.Errorf("estado debe ser on/off")
	}
}

func controlElectricoOrigenManual(origen string) bool {
	switch strings.ToLower(strings.TrimSpace(origen)) {
	case "prueba_manual", "sincronizacion_manual":
		return true
	default:
		return false
	}
}

func controlElectricoIncludeInactive(r *http.Request) bool {
	raw := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("include_inactive")))
	return raw == "1" || raw == "true" || raw == "si" || raw == "yes"
}

func controlElectricoParseLimit(r *http.Request, fallback int) int {
	raw := strings.TrimSpace(r.URL.Query().Get("limit"))
	if raw == "" {
		return fallback
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit <= 0 {
		return fallback
	}
	return limit
}
