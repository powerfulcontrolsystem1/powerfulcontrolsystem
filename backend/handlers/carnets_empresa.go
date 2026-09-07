package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaCarnetsHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		usuario := strings.TrimSpace(adminEmailFromRequest(r))

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "", "dashboard":
				row, err := dbpkg.BuildEmpresaCarnetsDashboard(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar el dashboard de carnets", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "plantillas":
				rows, err := dbpkg.ListEmpresaCarnetPlantillas(dbEmp, empresaID, queryBool(r, "include_inactive"))
				if err != nil {
					http.Error(w, "No se pudieron listar plantillas", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "personas":
				limit, _ := parseIntQueryOptional(r, "limit")
				rows, err := dbpkg.ListEmpresaCarnetPersonasFuente(dbEmp, empresaID, strings.TrimSpace(r.URL.Query().Get("q")), limit)
				if err != nil {
					http.Error(w, "No se pudieron listar usuarios fuente", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "eventos":
				carnetID, err := parseInt64Query(r, "carnet_id")
				if err != nil {
					http.Error(w, "carnet_id es obligatorio", http.StatusBadRequest)
					return
				}
				limit, _ := parseIntQueryOptional(r, "limit")
				rows, err := dbpkg.ListEmpresaCarnetEventos(dbEmp, empresaID, carnetID, limit)
				if err != nil {
					http.Error(w, "No se pudieron listar eventos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "carnets", "listar":
				limit, _ := parseIntQueryOptional(r, "limit")
				rows, err := dbpkg.ListEmpresaCarnets(dbEmp, empresaID, queryBool(r, "include_inactive"), strings.TrimSpace(r.URL.Query().Get("estado_carnet")), strings.TrimSpace(r.URL.Query().Get("tipo_persona")), strings.TrimSpace(r.URL.Query().Get("q")), limit)
				if err != nil {
					http.Error(w, "No se pudieron listar carnets", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			}

		case http.MethodPost:
			switch action {
			case "plantilla":
				var payload dbpkg.EmpresaCarnetPlantilla
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.UpsertEmpresaCarnetPlantilla(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			case "seed_demo":
				if err := dbpkg.SeedEmpresaCarnetsDemo(dbEmp, empresaID, usuario); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "", "carnet":
				var payload dbpkg.EmpresaCarnet
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.CreateEmpresaCarnet(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			}

		case http.MethodPut:
			switch action {
			case "plantilla":
				var payload dbpkg.EmpresaCarnetPlantilla
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.UpsertEmpresaCarnetPlantilla(dbEmp, payload)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "plantilla no encontrada", http.StatusNotFound)
						return
					}
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			case "estado":
				var payload struct {
					ID           int64  `json:"id"`
					EstadoCarnet string `json:"estado_carnet"`
					Detalle      string `json:"detalle"`
				}
				_ = json.NewDecoder(r.Body).Decode(&payload)
				if payload.ID <= 0 {
					if id, err := parseInt64QueryOptional(r, "id"); err == nil && id > 0 {
						payload.ID = id
					}
				}
				if strings.TrimSpace(payload.EstadoCarnet) == "" {
					payload.EstadoCarnet = strings.TrimSpace(r.URL.Query().Get("estado_carnet"))
				}
				if payload.ID <= 0 {
					http.Error(w, "id es obligatorio", http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetEmpresaCarnetEstado(dbEmp, empresaID, payload.ID, payload.EstadoCarnet, usuario, payload.Detalle); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "carnet no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "impreso":
				var payload struct {
					ID int64 `json:"id"`
				}
				_ = json.NewDecoder(r.Body).Decode(&payload)
				if payload.ID <= 0 {
					if id, err := parseInt64QueryOptional(r, "id"); err == nil && id > 0 {
						payload.ID = id
					}
				}
				if payload.ID <= 0 {
					http.Error(w, "id es obligatorio", http.StatusBadRequest)
					return
				}
				if err := dbpkg.MarkEmpresaCarnetImpreso(dbEmp, empresaID, payload.ID, usuario); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "carnet no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "", "carnet":
				var payload dbpkg.EmpresaCarnet
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				if err := dbpkg.UpdateEmpresaCarnet(dbEmp, payload); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "carnet no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			}
		}

		http.Error(w, "accion o metodo no soportado", http.StatusMethodNotAllowed)
	}
}
