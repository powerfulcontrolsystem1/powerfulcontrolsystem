package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func parseTruthy(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "si", "yes", "activo":
		return true
	default:
		return false
	}
}

// EmpresaFacturacionElectronicaHandler gestiona configuración FE por empresa y país.
func EmpresaFacturacionElectronicaHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			paisCodigo := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("pais_codigo")))
			incluirInactivas := parseTruthy(r.URL.Query().Get("incluir_inactivas"))

			if paisCodigo != "" {
				cfg, err := dbpkg.GetFacturacionElectronicaPaisConfig(dbEmp, empresaID, paisCodigo)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "No se pudo consultar la configuración de facturación electrónica", http.StatusInternalServerError)
					return
				}
				if cfg == nil {
					http.Error(w, "No se pudo resolver la configuración", http.StatusInternalServerError)
					return
				}
				if errors.Is(err, sql.ErrNoRows) {
					pais, source, derr := dbpkg.DetectFacturacionPais(dbEmp, empresaID, r.URL.Query().Get("tz"), r.URL.Query().Get("lang"))
					if derr == nil {
						cfg.PaisCodigo = pais.Codigo
						cfg.PaisNombre = pais.Nombre
						cfg.BanderaPais = pais.Bandera
						cfg.MonedaCodigo = pais.Moneda
						if cfg.Observaciones == "" {
							cfg.Observaciones = "Pais detectado por " + source
						}
					}
				}
				writeJSON(w, http.StatusOK, cfg)
				return
			}

			items, err := dbpkg.ListFacturacionElectronicaPaisConfigs(dbEmp, empresaID, incluirInactivas)
			if err != nil {
				http.Error(w, "No se pudo listar la configuración de facturación electrónica", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"empresa_id": empresaID,
				"items":      items,
			})
			return

		case http.MethodPost, http.MethodPut:
			var payload dbpkg.FacturacionElectronicaPaisConfig
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
			if strings.TrimSpace(payload.PaisCodigo) == "" {
				http.Error(w, "pais_codigo es obligatorio", http.StatusBadRequest)
				return
			}
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))

			id, err := dbpkg.UpsertFacturacionElectronicaPaisConfig(dbEmp, payload)
			if err != nil {
				http.Error(w, "No se pudo guardar la configuración de facturación electrónica", http.StatusBadRequest)
				return
			}
			cfg, err := dbpkg.GetFacturacionElectronicaPaisConfig(dbEmp, payload.EmpresaID, payload.PaisCodigo)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "No se pudo recuperar la configuración guardada", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":            true,
				"id":            id,
				"configuracion": cfg,
			})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

// EmpresaFacturacionElectronicaPaisDetectadoHandler detecta automáticamente país FE.
func EmpresaFacturacionElectronicaPaisDetectadoHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}

		empresaID, err := parseInt64QueryOptional(r, "empresa_id")
		if err != nil {
			http.Error(w, "empresa_id inválido", http.StatusBadRequest)
			return
		}

		tz := strings.TrimSpace(r.URL.Query().Get("tz"))
		if tz == "" {
			tz = strings.TrimSpace(r.URL.Query().Get("timezone"))
		}
		lang := strings.TrimSpace(r.URL.Query().Get("lang"))
		if lang == "" {
			acceptLang := strings.TrimSpace(r.Header.Get("Accept-Language"))
			if idx := strings.Index(acceptLang, ","); idx > 0 {
				lang = strings.TrimSpace(acceptLang[:idx])
			} else {
				lang = acceptLang
			}
		}

		pais, source, err := dbpkg.DetectFacturacionPais(dbEmp, empresaID, tz, lang)
		if err != nil {
			http.Error(w, "No se pudo detectar el país", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"empresa_id":  empresaID,
			"pais_codigo": pais.Codigo,
			"pais_nombre": pais.Nombre,
			"bandera":     pais.Bandera,
			"moneda":      pais.Moneda,
			"source":      source,
		})
	}
}

// EmpresaFacturacionElectronicaPaisesDisponiblesHandler retorna catálogo de países FE soportados.
func EmpresaFacturacionElectronicaPaisesDisponiblesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"items": dbpkg.ListPaisesFacturacionDisponibles(),
		})
	}
}
