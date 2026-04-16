package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

type superContractPayload struct {
	Titulo         string `json:"titulo"`
	Resumen        string `json:"resumen"`
	Contenido      string `json:"contenido"`
	NotaAceptacion string `json:"nota_aceptacion"`
	ResumenCambio  string `json:"resumen_cambio"`
}

func parseSuperContractVersion(raw string) (int, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, nil
	}
	version, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if version < 0 {
		version = 0
	}
	return version, nil
}

func PublicContratoHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		version, err := parseSuperContractVersion(r.URL.Query().Get("version"))
		if err != nil {
			http.Error(w, "invalid version", http.StatusBadRequest)
			return
		}

		if err := dbpkg.EnsureSuperContractSchema(dbSuper); err != nil {
			http.Error(w, "failed to prepare contract schema: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var contract *dbpkg.SuperContractVersion
		if version > 0 {
			contract, err = dbpkg.GetSuperContractVersionByNumber(dbSuper, version)
			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "contract version not found", http.StatusNotFound)
					return
				}
				http.Error(w, "failed to load contract version: "+err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			contract, err = dbpkg.GetCurrentSuperContract(dbSuper)
			if err != nil {
				http.Error(w, "failed to load current contract: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":       true,
			"contrato": contract,
		})
	}
}

func SuperContratoHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminEmail, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper)
		if !ok {
			return
		}

		if err := dbpkg.EnsureSuperContractSchema(dbSuper); err != nil {
			http.Error(w, "failed to prepare contract schema: "+err.Error(), http.StatusInternalServerError)
			return
		}

		switch r.Method {
		case http.MethodGet:
			current, err := dbpkg.GetCurrentSuperContract(dbSuper)
			if err != nil {
				http.Error(w, "failed to load current contract: "+err.Error(), http.StatusInternalServerError)
				return
			}
			history, err := dbpkg.ListSuperContractVersions(dbSuper, 25)
			if err != nil {
				http.Error(w, "failed to load contract history: "+err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":          true,
				"admin_email": adminEmail,
				"current":     current,
				"history":     history,
			})
			return

		case http.MethodPost, http.MethodPut:
			var payload superContractPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.Titulo) == "" {
				http.Error(w, "titulo is required", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.Contenido) == "" {
				http.Error(w, "contenido is required", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.ResumenCambio) == "" {
				http.Error(w, "resumen_cambio is required", http.StatusBadRequest)
				return
			}

			saved, noChanges, err := dbpkg.SaveSuperContractVersion(dbSuper, dbpkg.SuperContractVersion{
				Titulo:         payload.Titulo,
				Resumen:        payload.Resumen,
				Contenido:      payload.Contenido,
				NotaAceptacion: payload.NotaAceptacion,
				ResumenCambio:  payload.ResumenCambio,
				UsuarioCreador: adminEmail,
				Estado:         "activo",
			})
			if err != nil {
				http.Error(w, "failed to save contract version: "+err.Error(), http.StatusInternalServerError)
				return
			}
			history, err := dbpkg.ListSuperContractVersions(dbSuper, 25)
			if err != nil {
				http.Error(w, "failed to load contract history: "+err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":         true,
				"saved":      !noChanges,
				"no_changes": noChanges,
				"current":    saved,
				"history":    history,
			})
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}