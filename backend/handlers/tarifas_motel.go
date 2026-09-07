package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaTarifasMotelHandler gestiona planes tarifarios especializados para moteles.
func EmpresaTarifasMotelHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete {
			if err := dbpkg.EmpresaTarifasMotelSchemaReady(dbEmp); err != nil {
				writeTarifasMotelError(w, err)
				return
			}
		}
		switch r.Method {
		case http.MethodGet:
			handleTarifasMotelGet(w, r, dbEmp)
		case http.MethodPost:
			handleTarifasMotelCreate(w, r, dbEmp)
		case http.MethodPut:
			handleTarifasMotelUpdate(w, r, dbEmp)
		case http.MethodDelete:
			handleTarifasMotelDelete(w, r, dbEmp)
		default:
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		}
	}
}

func handleTarifasMotelGet(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	switch action {
	case "detalle", "get", "by_id":
		id, err := parseInt64Query(r, "id")
		if err != nil {
			http.Error(w, "id es obligatorio", http.StatusBadRequest)
			return
		}
		item, err := dbpkg.GetEmpresaTarifaMotelByID(dbEmp, empresaID, id)
		if err != nil {
			writeTarifasMotelError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
		return
	case "calcular", "simular":
		id, err := parseInt64Query(r, "id")
		if err != nil {
			http.Error(w, "id es obligatorio", http.StatusBadRequest)
			return
		}
		minutos, err := resolveTarifaMinutosConsumidosQuery(r)
		if err != nil {
			http.Error(w, "minutos_consumidos invalido", http.StatusBadRequest)
			return
		}
		item, err := dbpkg.GetEmpresaTarifaMotelByID(dbEmp, empresaID, id)
		if err != nil {
			writeTarifasMotelError(w, err)
			return
		}
		detalle := dbpkg.CalcularDetalleTarifaMotel(*item, minutos)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":      true,
			"tarifa":  item,
			"detalle": detalle,
		})
		return
	case "", "listar", "list":
		filter := dbpkg.EmpresaTarifaMotelFilter{IncludeInactive: queryBool(r, "include_inactive")}
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
		filter.TipoPlan = strings.TrimSpace(r.URL.Query().Get("tipo_plan"))
		rows, err := dbpkg.ListEmpresaTarifasMotel(dbEmp, empresaID, filter)
		if err != nil {
			writeTarifasMotelError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, rows)
		return
	default:
		http.Error(w, "accion no soportada", http.StatusBadRequest)
	}
}

func handleTarifasMotelCreate(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	var payload dbpkg.EmpresaTarifaMotel
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
	id, err := dbpkg.CreateEmpresaTarifaMotel(dbEmp, payload)
	if err != nil {
		writeTarifasMotelError(w, err)
		return
	}
	item, err := dbpkg.GetEmpresaTarifaMotelByID(dbEmp, payload.EmpresaID, id)
	if err != nil {
		writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func handleTarifasMotelUpdate(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
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
		next := "activo"
		if action == "desactivar" {
			next = "inactivo"
		}
		if err := dbpkg.SetEmpresaTarifaMotelEstado(dbEmp, empresaID, id, next); err != nil {
			writeTarifasMotelError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": next})
		return
	}

	var payload dbpkg.EmpresaTarifaMotel
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
	if err := dbpkg.UpdateEmpresaTarifaMotel(dbEmp, payload); err != nil {
		writeTarifasMotelError(w, err)
		return
	}
	item, err := dbpkg.GetEmpresaTarifaMotelByID(dbEmp, payload.EmpresaID, payload.ID)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func handleTarifasMotelDelete(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
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
	if err := dbpkg.DeleteEmpresaTarifaMotel(dbEmp, empresaID, id); err != nil {
		writeTarifasMotelError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

func writeTarifasMotelError(w http.ResponseWriter, err error) {
	if err == nil {
		http.Error(w, "error no especificado", http.StatusInternalServerError)
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "registro no encontrado", http.StatusNotFound)
		return
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "obligatorio") || strings.Contains(msg, "invalido") || strings.Contains(msg, "negativos") {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
