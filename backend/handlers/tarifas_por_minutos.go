package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaTarifasPorMinutosHandler gestiona tarifas por minutos por estacion y dia de semana.
func EmpresaTarifasPorMinutosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleTarifasPorMinutosGet(w, r, dbEmp)
			return
		case http.MethodPost:
			handleTarifasPorMinutosCreate(w, r, dbEmp)
			return
		case http.MethodPut:
			handleTarifasPorMinutosUpdate(w, r, dbEmp)
			return
		case http.MethodDelete:
			handleTarifasPorMinutosDelete(w, r, dbEmp)
			return
		default:
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
	}
}

func handleTarifasPorMinutosGet(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))

	switch action {
	case "config", "configuracion":
		cfg, err := dbpkg.GetEmpresaTarifaPorMinutosConfiguracion(dbEmp, empresaID)
		if err != nil {
			writeTarifasPorMinutosError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, cfg)
		return

	case "", "listar", "list":
		filter := dbpkg.EmpresaTarifaPorMinutosFilter{}
		filter.IncludeInactive = queryBool(r, "include_inactive")

		estacionID, err := parseInt64QueryOptional(r, "estacion_id")
		if err != nil || estacionID < 0 {
			http.Error(w, "estacion_id invalido", http.StatusBadRequest)
			return
		}
		filter.EstacionID = estacionID

		diaSemana, err := resolveTarifaDiaSemana(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		filter.DiaSemana = diaSemana

		limit, err := parseIntQueryOptional(r, "limit")
		if err != nil || limit < 0 {
			http.Error(w, "limit invalido", http.StatusBadRequest)
			return
		}
		filter.Limit = limit

		rows, err := dbpkg.ListEmpresaTarifasPorMinutos(dbEmp, empresaID, filter)
		if err != nil {
			writeTarifasPorMinutosError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, rows)
		return

	case "detalle", "get", "by_id":
		id, err := parseInt64Query(r, "id")
		if err != nil {
			http.Error(w, "id es obligatorio", http.StatusBadRequest)
			return
		}
		item, err := dbpkg.GetEmpresaTarifaPorMinutosByID(dbEmp, empresaID, id)
		if err != nil {
			writeTarifasPorMinutosError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
		return

	case "aplicable", "resolver", "tarifa_actual":
		estacionID, err := parseInt64Query(r, "estacion_id")
		if err != nil {
			http.Error(w, "estacion_id es obligatorio", http.StatusBadRequest)
			return
		}
		diaSemana, err := resolveTarifaDiaSemana(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		item, err := dbpkg.GetEmpresaTarifaPorMinutosAplicable(dbEmp, empresaID, estacionID, diaSemana)
		if err != nil {
			writeTarifasPorMinutosError(w, err)
			return
		}
		if item == nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":          false,
				"tarifa":      nil,
				"estacion_id": estacionID,
				"dia_semana":  diaSemana,
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":          true,
			"tarifa":      item,
			"estacion_id": estacionID,
			"dia_semana":  diaSemana,
		})
		return

	case "calcular", "simular":
		estacionID, err := parseInt64Query(r, "estacion_id")
		if err != nil {
			http.Error(w, "estacion_id es obligatorio", http.StatusBadRequest)
			return
		}
		minutosConsumidos, err := resolveTarifaMinutosConsumidosQuery(r)
		if err != nil {
			http.Error(w, "minutos_consumidos invalido", http.StatusBadRequest)
			return
		}
		diaSemana, err := resolveTarifaDiaSemana(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		tarifa, err := dbpkg.GetEmpresaTarifaPorMinutosAplicable(dbEmp, empresaID, estacionID, diaSemana)
		if err != nil {
			writeTarifasPorMinutosError(w, err)
			return
		}
		if tarifa == nil {
			http.Error(w, "no existe tarifa activa para la estacion y dia indicado", http.StatusNotFound)
			return
		}
		cfg, err := dbpkg.GetEmpresaTarifaPorMinutosConfiguracion(dbEmp, empresaID)
		if err != nil {
			writeTarifasPorMinutosError(w, err)
			return
		}
		detalle := dbpkg.CalcularDetalleTarifaPorMinutos(*tarifa, minutosConsumidos, *cfg)
		detalle.DiaSemana = diaSemana

		traceID, documentoCodigo, periodoContable, err := dbpkg.RegisterTarifaPorMinutosCalculoContable(
			dbEmp,
			empresaID,
			*tarifa,
			*cfg,
			diaSemana,
			minutosConsumidos,
			detalle,
			strings.TrimSpace(adminEmailFromRequest(r)),
			resolveTarifaPorMinutosRequestID(r),
		)
		if err != nil {
			http.Error(w, "no se pudo registrar trazabilidad contable del calculo", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":                       true,
			"tarifa":                   tarifa,
			"tarifa_id":                tarifa.ID,
			"estacion_id":              tarifa.EstacionID,
			"dia_semana":               diaSemana,
			"minutos_consumidos":       detalle.MinutosConsumidos,
			"bloques_extra":            detalle.BloquesExtra,
			"monto_base":               detalle.MontoBase,
			"monto_extra":              detalle.MontoExtra,
			"monto_subtotal":           detalle.MontoSubtotal,
			"monto_redondeado":         detalle.MontoRedondeado,
			"ajuste_redondeo":          detalle.AjusteRedondeo,
			"monto_minimo_aplicado":    detalle.MontoMinimoAplicado,
			"monto_maximo_aplicado":    detalle.MontoMaximoAplicado,
			"monto_minimo_diario":      cfg.MontoMinimoDiario,
			"monto_maximo_diario":      cfg.MontoMaximoDiario,
			"monto_total":              detalle.MontoTotal,
			"moneda":                   detalle.Moneda,
			"redondeo_modo":            cfg.RedondeoModo,
			"redondeo_unidad":          cfg.RedondeoUnidad,
			"trazabilidad_contable_id": traceID,
			"documento_tipo":           "tarifa_por_minutos_calculo",
			"documento_codigo":         documentoCodigo,
			"periodo_contable":         periodoContable,
		})
		return

	default:
		http.Error(w, "action invalida. Use: listar, detalle, aplicable, calcular o config", http.StatusBadRequest)
		return
	}
}

func handleTarifasPorMinutosCreate(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	var payload dbpkg.EmpresaTarifaPorMinutos
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID <= 0 {
		if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
			payload.EmpresaID = empresaID
		}
	}
	if payload.EmpresaID <= 0 {
		http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.UsuarioCreador) == "" {
		payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
	}

	id, err := dbpkg.CreateEmpresaTarifaPorMinutos(dbEmp, payload)
	if err != nil {
		writeTarifasPorMinutosError(w, err)
		return
	}
	item, err := dbpkg.GetEmpresaTarifaPorMinutosByID(dbEmp, payload.EmpresaID, id)
	if err != nil {
		writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func handleTarifasPorMinutosUpdate(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	if action == "config" || action == "configuracion" {
		var payload dbpkg.EmpresaTarifaPorMinutosConfiguracion
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "JSON invalido", http.StatusBadRequest)
			return
		}
		if payload.EmpresaID <= 0 {
			if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
				payload.EmpresaID = empresaID
			}
		}
		if payload.EmpresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(payload.UsuarioCreador) == "" {
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
		}
		cfg, err := dbpkg.UpsertEmpresaTarifaPorMinutosConfiguracion(dbEmp, payload)
		if err != nil {
			writeTarifasPorMinutosError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, cfg)
		return
	}
	if action == "aplicar_todas_estaciones" || action == "aplicar_todas" || action == "aplicar_global" {
		var payload dbpkg.EmpresaTarifaPorMinutos
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "JSON invalido", http.StatusBadRequest)
			return
		}
		if payload.EmpresaID <= 0 {
			if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
				payload.EmpresaID = empresaID
			}
		}
		if payload.EmpresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(payload.UsuarioCreador) == "" {
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
		}
		result, err := dbpkg.ApplyEmpresaTarifaPorMinutosToAllStations(dbEmp, payload)
		if err != nil {
			writeTarifasPorMinutosError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
		return
	}
	if action == "activar" || action == "desactivar" {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id, err := parseInt64Query(r, "id")
		if err != nil {
			http.Error(w, "id es obligatorio", http.StatusBadRequest)
			return
		}
		nextEstado := "activo"
		if action == "desactivar" {
			nextEstado = "inactivo"
		}
		if err := dbpkg.SetEmpresaTarifaPorMinutosEstado(dbEmp, empresaID, id, nextEstado); err != nil {
			writeTarifasPorMinutosError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": nextEstado})
		return
	}

	var payload dbpkg.EmpresaTarifaPorMinutos
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID <= 0 {
		if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
			payload.EmpresaID = empresaID
		}
	}
	if payload.ID <= 0 {
		if id, err := parseInt64QueryOptional(r, "id"); err == nil && id > 0 {
			payload.ID = id
		}
	}
	if payload.EmpresaID <= 0 || payload.ID <= 0 {
		http.Error(w, "empresa_id e id son obligatorios", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.UsuarioCreador) == "" {
		payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
	}

	if err := dbpkg.UpdateEmpresaTarifaPorMinutos(dbEmp, payload); err != nil {
		writeTarifasPorMinutosError(w, err)
		return
	}
	item, err := dbpkg.GetEmpresaTarifaPorMinutosByID(dbEmp, payload.EmpresaID, payload.ID)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func handleTarifasPorMinutosDelete(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := parseInt64Query(r, "id")
	if err != nil {
		http.Error(w, "id es obligatorio", http.StatusBadRequest)
		return
	}
	if err := dbpkg.DeleteEmpresaTarifaPorMinutos(dbEmp, empresaID, id); err != nil {
		writeTarifasPorMinutosError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

func resolveTarifaDiaSemana(r *http.Request) (int, error) {
	diaSemana, err := parseIntQueryOptional(r, "dia_semana")
	if err != nil {
		return 0, errors.New("dia_semana invalido")
	}
	if diaSemana > 0 {
		return diaSemana, nil
	}

	fechaRaw := strings.TrimSpace(firstNonEmptyStr(
		r.URL.Query().Get("fecha"),
		r.URL.Query().Get("fecha_operacion"),
		r.URL.Query().Get("fecha_hora"),
	))
	if fechaRaw == "" {
		return 0, nil
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02",
	}
	for _, layout := range layouts {
		tm, parseErr := time.ParseInLocation(layout, fechaRaw, time.Local)
		if parseErr == nil {
			return dbpkg.DayOfWeekISO(tm), nil
		}
	}
	return 0, errors.New("fecha invalida para resolver dia_semana")
}

func resolveTarifaMinutosConsumidosQuery(r *http.Request) (float64, error) {
	raw := strings.TrimSpace(firstNonEmptyStr(
		r.URL.Query().Get("minutos_consumidos_decimal"),
		r.URL.Query().Get("minutos_consumidos"),
	))
	if raw == "" {
		return 0, errors.New("minutos_consumidos es obligatorio")
	}
	raw = strings.ReplaceAll(raw, ",", ".")
	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, errors.New("minutos_consumidos invalido")
	}
	if val <= 0 {
		return 0, errors.New("minutos_consumidos debe ser mayor a cero")
	}
	return val, nil
}

func resolveTarifaPorMinutosRequestID(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("X-Request-ID")); v != "" {
		return v
	}
	if v := strings.TrimSpace(r.URL.Query().Get("request_id")); v != "" {
		return v
	}
	return ""
}

func writeTarifasPorMinutosError(w http.ResponseWriter, err error) {
	if err == nil {
		http.Error(w, "error no especificado", http.StatusInternalServerError)
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "tarifa por minutos no encontrada", http.StatusNotFound)
		return
	}

	msg := strings.TrimSpace(err.Error())
	msgLower := strings.ToLower(msg)

	if strings.Contains(msgLower, "unique constraint failed") {
		http.Error(w, "ya existe una tarifa para la estacion y rango de dias indicado", http.StatusConflict)
		return
	}
	if containsAny(msgLower,
		"obligatorio",
		"invalido",
		"debe",
		"negativo",
		"dia_semana",
		"no se encontraron estaciones",
		"redondeo",
		"monto_minimo_diario",
		"monto_maximo_diario",
		"constraint",
	) {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	http.Error(w, "No se pudo procesar tarifas por minutos", http.StatusInternalServerError)
}

func containsAny(text string, parts ...string) bool {
	for _, part := range parts {
		if strings.Contains(text, part) {
			return true
		}
	}
	return false
}
