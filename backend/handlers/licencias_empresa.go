package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func parseLicenciaIDsCSV(raw string) []int64 {
	parts := strings.Split(strings.TrimSpace(raw), ",")
	out := make([]int64, 0, len(parts))
	seen := map[int64]struct{}{}
	for _, part := range parts {
		id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err != nil || id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func EmpresaLicenciasAdicionalesHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseInt64Query(r, "empresa_id")
		if err != nil || empresaID <= 0 {
			http.Error(w, "empresa_id invalido", http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "" {
			action = "resumen"
		}

		switch r.Method {
		case http.MethodGet:
			if action != "resumen" {
				http.Error(w, "accion no soportada", http.StatusBadRequest)
				return
			}
			empresa, _ := dbpkg.GetEmpresaByID(dbSuper, empresaID)
			base, baseErr := dbpkg.GetActiveLicenciaByEmpresa(dbSuper, empresaID)
			if baseErr != nil && baseErr != sql.ErrNoRows {
				http.Error(w, "no se pudo cargar la licencia base", http.StatusInternalServerError)
				return
			}
			addons, err := dbpkg.ListEmpresaLicenciasAdicionales(dbSuper, empresaID, true)
			if err != nil {
				http.Error(w, "no se pudieron cargar las licencias adicionales", http.StatusInternalServerError)
				return
			}
			paisCodigo := "CO"
			if base != nil && strings.TrimSpace(base.PaisCodigo) != "" {
				paisCodigo = strings.TrimSpace(base.PaisCodigo)
			}
			catalog, err := dbpkg.GetLicenciasFilteredByPais(dbSuper, true, "", false, paisCodigo)
			if err != nil {
				http.Error(w, "no se pudo cargar el catalogo de licencias", http.StatusInternalServerError)
				return
			}
			activeAddonCatalog := map[int64]struct{}{}
			for _, addon := range addons {
				if addon.Activo == 1 {
					activeAddonCatalog[addon.LicenciaID] = struct{}{}
				}
			}
			availableAddons := make([]dbpkg.Licencia, 0)
			for _, item := range catalog {
				if item.EsAdicional != 1 || item.EmpresaID > 0 {
					continue
				}
				if _, exists := activeAddonCatalog[item.ID]; exists {
					continue
				}
				availableAddons = append(availableAddons, item)
			}
			bundle, err := dbpkg.BuildEmpresaLicenciaBundleSummary(dbSuper, empresaID, "empresa_bundle", nil)
			if err != nil {
				http.Error(w, "no se pudo construir el resumen agrupado", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"empresa": func() interface{} {
					if empresa == nil {
						return nil
					}
					return map[string]interface{}{
						"id":          empresa.ID,
						"empresa_id":  empresa.EmpresaID,
						"nombre":      empresa.Nombre,
						"tipo_id":     empresa.TipoID,
						"tipo_nombre": empresa.TipoNombre,
					}
				}(),
				"base_licencia":         base,
				"licencias_adicionales": addons,
				"catalogo_adicionales":  availableAddons,
				"bundle_summary":        bundle,
			})
			return
		case http.MethodPost:
			switch action {
			case "desactivar_adicional", "activar_adicional", "auto_renovar":
				var payload struct {
					LicenciaID    int64  `json:"licencia_id"`
					Activo        *bool  `json:"activo,omitempty"`
					AutoRenovar   *bool  `json:"auto_renovar,omitempty"`
					Observaciones string `json:"observaciones,omitempty"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "payload invalido", http.StatusBadRequest)
					return
				}
				if payload.LicenciaID <= 0 {
					http.Error(w, "licencia_id invalido", http.StatusBadRequest)
					return
				}
				activo := true
				switch action {
				case "desactivar_adicional":
					activo = false
				case "activar_adicional":
					activo = true
				case "auto_renovar":
					current, err := dbpkg.GetEmpresaLicenciaAdicionalByEmpresaYLicencia(dbSuper, empresaID, payload.LicenciaID)
					if err != nil {
						http.Error(w, "no se pudo localizar la licencia adicional", http.StatusNotFound)
						return
					}
					activo = current.Activo == 1
				}
				if err := dbpkg.SetEmpresaLicenciaAdicionalEstado(dbSuper, empresaID, payload.LicenciaID, activo, payload.AutoRenovar, payload.Observaciones); err != nil {
					http.Error(w, "no se pudo actualizar la licencia adicional: "+err.Error(), http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			default:
				http.Error(w, "accion no soportada", http.StatusBadRequest)
				return
			}
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
