package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaPortalTercerosCertificadosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "" {
			action = "dashboard"
		}
		usuario := strings.TrimSpace(adminEmailFromRequest(r))
		if usuario == "" {
			usuario = "sistema"
		}

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "dashboard":
				row, err := dbpkg.BuildEmpresaPortalTercerosCertificadosDashboard(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar el portal de terceros", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "terceros":
				rows, err := dbpkg.ListEmpresaPortalTerceros(dbEmp, empresaID, r.URL.Query().Get("q"), 500)
				if err != nil {
					http.Error(w, "No se pudieron listar terceros", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "certificados":
				rows, err := dbpkg.ListEmpresaCertificadosTributarios(dbEmp, empresaID, r.URL.Query().Get("estado"), r.URL.Query().Get("tipo"), 500)
				if err != nil {
					http.Error(w, "No se pudieron listar certificados", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "descargas":
				rows, err := dbpkg.ListEmpresaCertificadosTributariosDescargas(dbEmp, empresaID, 500)
				if err != nil {
					http.Error(w, "No se pudo listar la bitacora de descargas", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			}
		case http.MethodPost, http.MethodPut:
			switch action {
			case "tercero":
				var payload dbpkg.EmpresaPortalTercero
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.UpsertEmpresaPortalTercero(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			case "certificado":
				var payload dbpkg.EmpresaCertificadoTributario
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.UpsertEmpresaCertificadoTributario(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			case "seed_demo":
				if err := dbpkg.SeedEmpresaPortalTercerosCertificadosDemo(dbEmp, empresaID, usuario); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			}
		}
		http.Error(w, fmt.Sprintf("Metodo o accion no permitida: %s", action), http.StatusMethodNotAllowed)
	}
}

func PublicCertificadosTributariosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		if token == "" {
			http.Error(w, "token requerido", http.StatusBadRequest)
			return
		}
		cert, err := dbpkg.GetEmpresaCertificadoTributarioByToken(dbEmp, token)
		if err != nil {
			http.Error(w, "Certificado no encontrado o no disponible", http.StatusNotFound)
			return
		}
		if r.Method == http.MethodPost {
			_, _ = dbpkg.CreateEmpresaCertificadoTributarioDescarga(dbEmp, dbpkg.EmpresaCertificadoTributarioDescarga{
				EmpresaID:       cert.EmpresaID,
				CertificadoID:   cert.ID,
				TerceroID:       cert.TerceroID,
				Canal:           "portal_publico",
				IP:              clientIPForCertificados(r),
				UserAgent:       r.UserAgent(),
				ValidacionClave: token,
			})
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}
		writeJSON(w, http.StatusOK, cert)
	}
}

func clientIPForCertificados(r *http.Request) string {
	if r == nil {
		return ""
	}
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		if v := strings.TrimSpace(r.Header.Get(header)); v != "" {
			if idx := strings.Index(v, ","); idx >= 0 {
				return strings.TrimSpace(v[:idx])
			}
			return v
		}
	}
	return strings.TrimSpace(r.RemoteAddr)
}
