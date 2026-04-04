package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaFinanzasMovimientosHandler gestiona CRUD de ingresos/egresos por empresa.
func EmpresaFinanzasMovimientosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			includeInactive := queryBool(r, "include_inactive")
			tipo := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("tipo")))
			desde := strings.TrimSpace(r.URL.Query().Get("desde"))
			hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
			periodo := strings.TrimSpace(r.URL.Query().Get("periodo"))
			q := strings.TrimSpace(r.URL.Query().Get("q"))
			limit, err := parseIntQueryOptional(r, "limit")
			if err != nil {
				http.Error(w, "limit invalido", http.StatusBadRequest)
				return
			}
			rows, err := dbpkg.ListEmpresaFinanzasMovimientos(dbEmp, empresaID, dbpkg.EmpresaFinanzasMovimientoFilter{
				Tipo:            tipo,
				Desde:           desde,
				Hasta:           hasta,
				Periodo:         periodo,
				Q:               q,
				IncludeInactive: includeInactive,
				Limit:           limit,
			})
			if err != nil {
				http.Error(w, "No se pudieron listar los movimientos financieros", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, rows)
			return

		case http.MethodPost:
			var payload dbpkg.EmpresaFinanzasMovimiento
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 {
				if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
					payload.EmpresaID = empresaID
				}
			}
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			id, err := dbpkg.CreateEmpresaFinanzasMovimiento(dbEmp, payload)
			if err != nil {
				if errors.Is(err, dbpkg.ErrPeriodoFinancieroCerrado) {
					http.Error(w, "el periodo contable del movimiento esta cerrado", http.StatusConflict)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			return

		case http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "activar" || action == "desactivar" || action == "anular" {
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
				estado := "activo"
				if action == "desactivar" {
					estado = "inactivo"
				}
				if action == "anular" {
					estado = "anulado"
				}
				if err := dbpkg.SetEmpresaFinanzasMovimientoEstado(dbEmp, empresaID, id, estado); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "movimiento no encontrado", http.StatusNotFound)
						return
					}
					if errors.Is(err, dbpkg.ErrPeriodoFinancieroCerrado) {
						http.Error(w, "el periodo contable del movimiento esta cerrado", http.StatusConflict)
						return
					}
					http.Error(w, "No se pudo actualizar el estado", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": estado})
				return
			}

			var payload dbpkg.EmpresaFinanzasMovimiento
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 || payload.ID <= 0 {
				http.Error(w, "id y empresa_id son obligatorios", http.StatusBadRequest)
				return
			}
			if payload.UsuarioCreador == "" {
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			}
			if err := dbpkg.UpdateEmpresaFinanzasMovimiento(dbEmp, payload); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "movimiento no encontrado", http.StatusNotFound)
					return
				}
				if errors.Is(err, dbpkg.ErrPeriodoFinancieroCerrado) {
					http.Error(w, "el periodo contable del movimiento esta cerrado", http.StatusConflict)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return

		case http.MethodDelete:
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
			if err := dbpkg.DeleteEmpresaFinanzasMovimiento(dbEmp, empresaID, id); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "movimiento no encontrado", http.StatusNotFound)
					return
				}
				if errors.Is(err, dbpkg.ErrPeriodoFinancieroCerrado) {
					http.Error(w, "el periodo contable del movimiento esta cerrado", http.StatusConflict)
					return
				}
				http.Error(w, "No se pudo eliminar el movimiento", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

// EmpresaFinanzasConfiguracionHandler gestiona configuracion por empresa del modulo financiero.
func EmpresaFinanzasConfiguracionHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			cfg, err := dbpkg.GetEmpresaFinanzasConfiguracion(dbEmp, empresaID)
			if err != nil {
				http.Error(w, "No se pudo consultar la configuracion financiera", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, cfg)
			return

		case http.MethodPost, http.MethodPut:
			var payload dbpkg.EmpresaFinanzasConfiguracion
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
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			id, err := dbpkg.UpsertEmpresaFinanzasConfiguracion(dbEmp, payload)
			if err != nil {
				http.Error(w, "No se pudo guardar la configuracion financiera", http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

// EmpresaFinanzasPeriodosHandler gestiona periodos contables por empresa.
func EmpresaFinanzasPeriodosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			includeInactive := queryBool(r, "include_inactive")
			rows, err := dbpkg.ListEmpresaFinanzasPeriodos(dbEmp, empresaID, includeInactive)
			if err != nil {
				http.Error(w, "No se pudieron listar los periodos", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, rows)
			return

		case http.MethodPost:
			var payload dbpkg.EmpresaFinanzasPeriodo
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
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			id, err := dbpkg.UpsertEmpresaFinanzasPeriodo(dbEmp, payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
			return

		case http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "cerrar" || action == "reabrir" {
				empresaID, err := parseEmpresaIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				periodo := strings.TrimSpace(r.URL.Query().Get("periodo"))
				if periodo == "" {
					http.Error(w, "periodo es obligatorio", http.StatusBadRequest)
					return
				}
				estado := "cerrado"
				if action == "reabrir" {
					estado = "abierto"
				}
				if err := dbpkg.SetEmpresaFinanzasPeriodoEstado(dbEmp, empresaID, periodo, estado, strings.TrimSpace(adminEmailFromRequest(r)), "actualizacion desde API"); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "periodo": periodo, "estado": estado})
				return
			}

			var payload dbpkg.EmpresaFinanzasPeriodo
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
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			id, err := dbpkg.UpsertEmpresaFinanzasPeriodo(dbEmp, payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}
