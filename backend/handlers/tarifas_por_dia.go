package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaTarifasPorDiaHandler gestiona tarifas diarias por estacion.
func EmpresaTarifasPorDiaHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleTarifasPorDiaGet(w, r, dbEmp)
			return
		case http.MethodPost:
			handleTarifasPorDiaCreate(w, r, dbEmp)
			return
		case http.MethodPut:
			handleTarifasPorDiaUpdate(w, r, dbEmp)
			return
		case http.MethodDelete:
			handleTarifasPorDiaDelete(w, r, dbEmp)
			return
		default:
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
	}
}

func handleTarifasPorDiaGet(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))

	switch action {
	case "", "listar", "list":
		filter := dbpkg.EmpresaTarifaPorDiaFilter{}
		filter.IncludeInactive = queryBool(r, "include_inactive")

		estacionID, err := parseInt64QueryOptional(r, "estacion_id")
		if err != nil || estacionID < 0 {
			http.Error(w, "estacion_id invalido", http.StatusBadRequest)
			return
		}
		filter.EstacionID = estacionID

		limit, err := parseIntQueryOptional(r, "limit")
		if err != nil || limit < 0 {
			http.Error(w, "limit invalido", http.StatusBadRequest)
			return
		}
		filter.Limit = limit

		rows, err := dbpkg.ListEmpresaTarifasPorDia(dbEmp, empresaID, filter)
		if err != nil {
			writeTarifasPorDiaError(w, err)
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
		item, err := dbpkg.GetEmpresaTarifaPorDiaByID(dbEmp, empresaID, id)
		if err != nil {
			writeTarifasPorDiaError(w, err)
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
		item, err := dbpkg.GetEmpresaTarifaPorDiaAplicable(dbEmp, empresaID, estacionID)
		if err != nil {
			writeTarifasPorDiaError(w, err)
			return
		}
		if item == nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":          false,
				"tarifa":      nil,
				"estacion_id": estacionID,
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":          true,
			"tarifa":      item,
			"estacion_id": estacionID,
		})
		return

	case "calcular", "simular":
		estacionID, err := parseInt64Query(r, "estacion_id")
		if err != nil {
			http.Error(w, "estacion_id es obligatorio", http.StatusBadRequest)
			return
		}

		tarifa, err := dbpkg.GetEmpresaTarifaPorDiaActiva(dbEmp, empresaID, estacionID)
		if err != nil {
			writeTarifasPorDiaError(w, err)
			return
		}
		if tarifa == nil {
			http.Error(w, "no existe tarifa activa para la estacion indicada", http.StatusNotFound)
			return
		}

		fechaInicio, err := resolveTarifaPorDiaDateTimeQuery(r, "activado_en", time.Now())
		if err != nil {
			http.Error(w, "activado_en invalido", http.StatusBadRequest)
			return
		}
		fechaCorte, err := resolveTarifaPorDiaDateTimeQuery(r, "fecha_corte", time.Now())
		if err != nil {
			http.Error(w, "fecha_corte invalida", http.StatusBadRequest)
			return
		}
		if fechaCorte.Before(fechaInicio) {
			http.Error(w, "fecha_corte debe ser mayor o igual a activado_en", http.StatusBadRequest)
			return
		}

		detalle := dbpkg.CalcularDetalleTarifaPorDia(*tarifa, fechaInicio, fechaCorte)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":                      true,
			"tarifa":                  tarifa,
			"tarifa_id":               tarifa.ID,
			"estacion_id":             tarifa.EstacionID,
			"activado_en":             detalle.FechaInicio,
			"fecha_corte":             detalle.FechaCorte,
			"dias_cobrados":           detalle.DiasCobrados,
			"valor_dia":               detalle.ValorDia,
			"monto_total":             detalle.MontoTotal,
			"hora_check_in":           detalle.HoraCheckIn,
			"hora_check_out":          detalle.HoraCheckOut,
			"aplicar_automaticamente": tarifa.AplicarAutomaticamente,
			"moneda":                  detalle.Moneda,
		})
		return

	default:
		http.Error(w, "action invalida. Use: listar, detalle, aplicable o calcular", http.StatusBadRequest)
		return
	}
}

func handleTarifasPorDiaCreate(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	var payload dbpkg.EmpresaTarifaPorDia
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

	id, err := dbpkg.CreateEmpresaTarifaPorDia(dbEmp, payload)
	if err != nil {
		writeTarifasPorDiaError(w, err)
		return
	}
	item, err := dbpkg.GetEmpresaTarifaPorDiaByID(dbEmp, payload.EmpresaID, id)
	if err != nil {
		writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func handleTarifasPorDiaUpdate(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
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
		if err := dbpkg.SetEmpresaTarifaPorDiaEstado(dbEmp, empresaID, id, nextEstado); err != nil {
			writeTarifasPorDiaError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": nextEstado})
		return
	}

	var payload dbpkg.EmpresaTarifaPorDia
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

	if err := dbpkg.UpdateEmpresaTarifaPorDia(dbEmp, payload); err != nil {
		writeTarifasPorDiaError(w, err)
		return
	}
	item, err := dbpkg.GetEmpresaTarifaPorDiaByID(dbEmp, payload.EmpresaID, payload.ID)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func handleTarifasPorDiaDelete(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
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
	if err := dbpkg.DeleteEmpresaTarifaPorDia(dbEmp, empresaID, id); err != nil {
		writeTarifasPorDiaError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

func parseTarifaPorDiaHandlerDateTime(raw string) (time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, errors.New("fecha vacia")
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"2006-01-02",
	}
	for _, layout := range layouts {
		ts, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			return ts, nil
		}
	}
	return time.Time{}, errors.New("fecha invalida")
}

func resolveTarifaPorDiaDateTimeQuery(r *http.Request, key string, fallback time.Time) (time.Time, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return fallback, nil
	}
	return parseTarifaPorDiaHandlerDateTime(raw)
}

func writeTarifasPorDiaError(w http.ResponseWriter, err error) {
	if err == nil {
		http.Error(w, "error no especificado", http.StatusInternalServerError)
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "tarifa por dia no encontrada", http.StatusNotFound)
		return
	}

	msg := strings.TrimSpace(err.Error())
	msgLower := strings.ToLower(msg)

	if strings.Contains(msgLower, "unique constraint failed") {
		http.Error(w, "ya existe una tarifa por dia para la estacion indicada", http.StatusConflict)
		return
	}
	if containsAny(msgLower,
		"obligatorio",
		"invalido",
		"debe",
		"constraint",
		"hora_check",
		"valor_dia",
	) {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	http.Error(w, "No se pudo procesar tarifas por dia", http.StatusInternalServerError)
}
