package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	panelEmpresaFavoritosEnabledKey = "panel_empresa.favoritos_enabled"
	panelEmpresaEmailEnabledKey     = "panel_empresa.email_enabled"
	panelEmpresaNoticiasEnabledKey  = "panel_empresa.noticias_enabled"
	panelEmpresaBuzonEnabledKey     = "panel_empresa.buzon_enabled"
	panelEmpresaChatEnabledKey      = "panel_empresa.chat_enabled"
)

type empresaPanelConfigResponse struct {
	OK               bool  `json:"ok"`
	EmpresaID        int64 `json:"empresa_id"`
	FavoritosEnabled bool  `json:"favoritos_enabled"`
	EmailEnabled     bool  `json:"email_enabled"`
	NoticiasEnabled  bool  `json:"noticias_enabled"`
	BuzonEnabled     bool  `json:"buzon_enabled"`
	ChatEnabled      bool  `json:"chat_enabled"`
	DefaultEnabled   bool  `json:"default_enabled"`
	PreferenciasBase bool  `json:"preferencias_base"`
}

type empresaPanelConfigPayload struct {
	EmpresaID        int64 `json:"empresa_id"`
	FavoritosEnabled *bool `json:"favoritos_enabled"`
	EmailEnabled     *bool `json:"email_enabled"`
	NoticiasEnabled  *bool `json:"noticias_enabled"`
	BuzonEnabled     *bool `json:"buzon_enabled"`
	ChatEnabled      *bool `json:"chat_enabled"`
}

func parsePanelEmpresaBool(raw string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "on", "activo", "enabled", "si", "yes":
		return true
	case "0", "false", "off", "inactivo", "disabled", "no":
		return false
	default:
		return fallback
	}
}

func panelEmpresaBoolValue(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

func empresaPanelConfigEmpresaID(r *http.Request, payloadID int64) int64 {
	if payloadID > 0 {
		return payloadID
	}
	raw := strings.TrimSpace(r.URL.Query().Get("empresa_id"))
	if raw == "" {
		raw = strings.TrimSpace(r.URL.Query().Get("id"))
	}
	if raw == "" {
		return 0
	}
	id, _ := strconv.ParseInt(raw, 10, 64)
	return id
}

func getEmpresaPanelPref(dbEmp *sql.DB, empresaID int64, key string, fallback bool) bool {
	if dbEmp == nil || empresaID <= 0 {
		return fallback
	}
	if err := dbpkg.EnsureEmpresaEstacionPrefsSchema(dbEmp); err != nil {
		return fallback
	}
	pref, err := dbpkg.GetEmpresaEstacionPref(dbEmp, empresaID, 0, key)
	if err != nil || pref == nil {
		return fallback
	}
	return parsePanelEmpresaBool(pref.Valor, fallback)
}

func setEmpresaPanelPref(dbEmp *sql.DB, empresaID int64, key string, enabled bool, usuario string) error {
	if err := dbpkg.EnsureEmpresaEstacionPrefsSchema(dbEmp); err != nil {
		return err
	}
	_, err := dbpkg.UpsertEmpresaEstacionPref(dbEmp, dbpkg.EmpresaEstacionPref{
		EmpresaID:      empresaID,
		EstacionID:     0,
		Clave:          key,
		Valor:          panelEmpresaBoolValue(enabled),
		UsuarioCreador: strings.TrimSpace(usuario),
		Estado:         "activo",
		Observaciones:  "Configuracion de tarjetas del panel de administrar empresa.",
	})
	return err
}

func empresaPanelConfigResponseFromDB(dbEmp *sql.DB, empresaID int64) empresaPanelConfigResponse {
	return empresaPanelConfigResponse{
		OK:               true,
		EmpresaID:        empresaID,
		FavoritosEnabled: getEmpresaPanelPref(dbEmp, empresaID, panelEmpresaFavoritosEnabledKey, true),
		EmailEnabled:     getEmpresaPanelPref(dbEmp, empresaID, panelEmpresaEmailEnabledKey, true),
		NoticiasEnabled:  getEmpresaPanelPref(dbEmp, empresaID, panelEmpresaNoticiasEnabledKey, true),
		BuzonEnabled:     getEmpresaPanelPref(dbEmp, empresaID, panelEmpresaBuzonEnabledKey, true),
		ChatEnabled:      getEmpresaPanelPref(dbEmp, empresaID, panelEmpresaChatEnabledKey, false),
		DefaultEnabled:   true,
		PreferenciasBase: true,
	}
}

// EmpresaPanelConfiguracionHandler permite activar u ocultar tarjetas del panel por empresa.
func EmpresaPanelConfiguracionHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			empresaID := empresaPanelConfigEmpresaID(r, 0)
			if empresaID <= 0 {
				http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
				return
			}
			encodeJSONResponse(w, empresaPanelConfigResponseFromDB(dbEmp, empresaID))
		case http.MethodPost, http.MethodPut:
			var payload empresaPanelConfigPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload invalido", http.StatusBadRequest)
				return
			}
			empresaID := empresaPanelConfigEmpresaID(r, payload.EmpresaID)
			if empresaID <= 0 {
				http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
				return
			}
			usuario := adminEmailFromRequest(r)
			if payload.FavoritosEnabled != nil {
				if err := setEmpresaPanelPref(dbEmp, empresaID, panelEmpresaFavoritosEnabledKey, *payload.FavoritosEnabled, usuario); err != nil {
					http.Error(w, "no se pudo guardar favoritos", http.StatusInternalServerError)
					return
				}
			}
			if payload.EmailEnabled != nil {
				if err := setEmpresaPanelPref(dbEmp, empresaID, panelEmpresaEmailEnabledKey, *payload.EmailEnabled, usuario); err != nil {
					http.Error(w, "no se pudo guardar email corporativo", http.StatusInternalServerError)
					return
				}
			}
			if payload.NoticiasEnabled != nil {
				if err := setEmpresaPanelPref(dbEmp, empresaID, panelEmpresaNoticiasEnabledKey, *payload.NoticiasEnabled, usuario); err != nil {
					http.Error(w, "no se pudo guardar noticias", http.StatusInternalServerError)
					return
				}
			}
			if payload.BuzonEnabled != nil {
				if err := setEmpresaPanelPref(dbEmp, empresaID, panelEmpresaBuzonEnabledKey, *payload.BuzonEnabled, usuario); err != nil {
					http.Error(w, "no se pudo guardar buzon", http.StatusInternalServerError)
					return
				}
			}
			if payload.ChatEnabled != nil {
				if err := setEmpresaPanelPref(dbEmp, empresaID, panelEmpresaChatEnabledKey, *payload.ChatEnabled, usuario); err != nil {
					http.Error(w, "no se pudo guardar chat", http.StatusInternalServerError)
					return
				}
			}
			encodeJSONResponse(w, empresaPanelConfigResponseFromDB(dbEmp, empresaID))
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
