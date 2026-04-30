package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaHojaVidaOperativaHandler gestiona hojas de vida universales por empresa.
func EmpresaHojaVidaOperativaHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := dbpkg.EnsureEmpresaHojaVidaOperativaSchema(dbEmp); err != nil {
			log.Printf("[hoja_vida] ensure schema error: %v", err)
			http.Error(w, "No se pudo preparar hoja de vida operativa", http.StatusInternalServerError)
			return
		}

		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch r.Method {
		case http.MethodGet:
			handleEmpresaHojaVidaGet(w, r, dbEmp, action)
		case http.MethodPost:
			handleEmpresaHojaVidaPost(w, r, dbEmp, action)
		case http.MethodPut:
			handleEmpresaHojaVidaPut(w, r, dbEmp, action)
		case http.MethodDelete:
			handleEmpresaHojaVidaDelete(w, r, dbEmp, action)
		default:
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		}
	}
}

func handleEmpresaHojaVidaGet(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, action string) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	limit, err := parseIntQueryOptional(r, "limit")
	if err != nil {
		http.Error(w, "limit invalido", http.StatusBadRequest)
		return
	}

	switch action {
	case "eventos":
		entidadID, err := parseInt64Query(r, "entidad_id")
		if err != nil || entidadID <= 0 {
			http.Error(w, "entidad_id es obligatorio", http.StatusBadRequest)
			return
		}
		rows, err := dbpkg.ListEmpresaHojaVidaEventos(dbEmp, empresaID, entidadID, limit)
		if err != nil {
			log.Printf("[hoja_vida] list eventos empresa_id=%d entidad_id=%d error: %v", empresaID, entidadID, err)
			http.Error(w, "No se pudieron consultar eventos", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, rows)
		return

	case "alertas":
		entidadID, _ := parseInt64QueryOptional(r, "entidad_id")
		estadoAlerta := strings.TrimSpace(r.URL.Query().Get("estado_alerta"))
		rows, err := dbpkg.ListEmpresaHojaVidaAlertas(dbEmp, empresaID, entidadID, estadoAlerta, limit)
		if err != nil {
			log.Printf("[hoja_vida] list alertas empresa_id=%d entidad_id=%d error: %v", empresaID, entidadID, err)
			http.Error(w, "No se pudieron consultar alertas", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, rows)
		return

	case "reporte", "resumen":
		reporte, err := dbpkg.GetEmpresaHojaVidaReporte(dbEmp, empresaID)
		if err != nil {
			log.Printf("[hoja_vida] reporte empresa_id=%d error: %v", empresaID, err)
			http.Error(w, "No se pudo consultar reporte", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, reporte)
		return
	}

	tipo := strings.TrimSpace(r.URL.Query().Get("tipo_entidad"))
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	includeInactive := queryBool(r, "include_inactive")
	rows, err := dbpkg.ListEmpresaHojaVidaEntidades(dbEmp, empresaID, tipo, q, includeInactive, limit)
	if err != nil {
		log.Printf("[hoja_vida] list entidades empresa_id=%d error: %v", empresaID, err)
		http.Error(w, "No se pudieron consultar hojas de vida", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func handleEmpresaHojaVidaPost(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, action string) {
	switch action {
	case "evento", "eventos":
		var payload dbpkg.EmpresaHojaVidaEvento
		if err := decodeJSON(r, &payload); err != nil {
			http.Error(w, "JSON invalido", http.StatusBadRequest)
			return
		}
		if err := hydrateEmpresaID(r, &payload.EmpresaID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if payload.EntidadID <= 0 || strings.TrimSpace(payload.Titulo) == "" {
			http.Error(w, "entidad_id y titulo son obligatorios", http.StatusBadRequest)
			return
		}
		payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
		id, err := dbpkg.CreateEmpresaHojaVidaEvento(dbEmp, payload)
		if err != nil {
			log.Printf("[hoja_vida] create evento empresa_id=%d error: %v", payload.EmpresaID, err)
			http.Error(w, "No se pudo crear evento", http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
		return

	case "alerta", "alertas":
		var payload dbpkg.EmpresaHojaVidaAlerta
		if err := decodeJSON(r, &payload); err != nil {
			http.Error(w, "JSON invalido", http.StatusBadRequest)
			return
		}
		if err := hydrateEmpresaID(r, &payload.EmpresaID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if payload.EntidadID <= 0 || strings.TrimSpace(payload.Titulo) == "" {
			http.Error(w, "entidad_id y titulo son obligatorios", http.StatusBadRequest)
			return
		}
		payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
		id, err := dbpkg.CreateEmpresaHojaVidaAlerta(dbEmp, payload)
		if err != nil {
			log.Printf("[hoja_vida] create alerta empresa_id=%d error: %v", payload.EmpresaID, err)
			http.Error(w, "No se pudo crear alerta", http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
		return
	}

	var payload dbpkg.EmpresaHojaVidaEntidad
	if err := decodeJSON(r, &payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if err := hydrateEmpresaID(r, &payload.EmpresaID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.Nombre) == "" {
		http.Error(w, "nombre es obligatorio", http.StatusBadRequest)
		return
	}
	payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
	id, err := dbpkg.CreateEmpresaHojaVidaEntidad(dbEmp, payload)
	if err != nil {
		log.Printf("[hoja_vida] create entidad empresa_id=%d error: %v", payload.EmpresaID, err)
		http.Error(w, "No se pudo crear hoja de vida", http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
}

func handleEmpresaHojaVidaPut(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, action string) {
	switch action {
	case "alerta_estado", "completar_alerta", "reabrir_alerta":
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id, err := parseInt64Query(r, "id")
		if err != nil || id <= 0 {
			http.Error(w, "id es obligatorio", http.StatusBadRequest)
			return
		}
		estado := strings.TrimSpace(r.URL.Query().Get("estado_alerta"))
		if estado == "" && action == "completar_alerta" {
			estado = "completada"
		}
		if estado == "" && action == "reabrir_alerta" {
			estado = "pendiente"
		}
		if err := dbpkg.SetEmpresaHojaVidaAlertaEstado(dbEmp, empresaID, id, estado); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "alerta no encontrada", http.StatusNotFound)
				return
			}
			http.Error(w, "No se pudo actualizar alerta", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado_alerta": estado})
		return
	}

	var payload dbpkg.EmpresaHojaVidaEntidad
	if err := decodeJSON(r, &payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if err := hydrateEmpresaID(r, &payload.EmpresaID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if payload.ID <= 0 || strings.TrimSpace(payload.Nombre) == "" {
		http.Error(w, "id y nombre son obligatorios", http.StatusBadRequest)
		return
	}
	if err := dbpkg.UpdateEmpresaHojaVidaEntidad(dbEmp, payload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "hoja de vida no encontrada", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo actualizar hoja de vida", http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

func handleEmpresaHojaVidaDelete(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, action string) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := parseInt64Query(r, "id")
	if err != nil || id <= 0 {
		http.Error(w, "id es obligatorio", http.StatusBadRequest)
		return
	}

	switch action {
	case "evento", "eventos":
		err = dbpkg.DeleteEmpresaHojaVidaEvento(dbEmp, empresaID, id)
	case "alerta", "alertas":
		err = dbpkg.DeleteEmpresaHojaVidaAlerta(dbEmp, empresaID, id)
	default:
		err = dbpkg.DeleteEmpresaHojaVidaEntidad(dbEmp, empresaID, id)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "registro no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo eliminar registro", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

func decodeJSON(r *http.Request, out interface{}) error {
	return json.NewDecoder(r.Body).Decode(out)
}

func hydrateEmpresaID(r *http.Request, target *int64) error {
	if target == nil {
		return errBadRequest("empresa_id es obligatorio")
	}
	if *target <= 0 {
		if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
			*target = empresaID
		}
	}
	if *target <= 0 {
		return errBadRequest("empresa_id es obligatorio")
	}
	return nil
}
