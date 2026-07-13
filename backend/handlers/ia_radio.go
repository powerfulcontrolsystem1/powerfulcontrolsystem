package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

// EmpresaIARadioHandler limita la IA a encender/apagar la emisora por empresa.
func EmpresaIARadioHandler(dbSuper, dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			EmpresaID          int64   `json:"empresa_id"`
			RadioOnlineEnabled *bool   `json:"radio_online_enabled"`
			Activo             *bool   `json:"activo"`
			Accion             string  `json:"accion"`
			RadioCountry       *string `json:"radio_country"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "payload invalido: "+err.Error(), http.StatusBadRequest)
			return
		}
		empresaID := payload.EmpresaID
		if empresaID <= 0 {
			if id, ok := r.Context().Value("empresaID").(int64); ok && id > 0 {
				empresaID = id
			}
		}
		if empresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}

		enabled := true
		if payload.RadioOnlineEnabled != nil {
			enabled = *payload.RadioOnlineEnabled
		} else if payload.Activo != nil {
			enabled = *payload.Activo
		} else {
			action := strings.ToLower(strings.TrimSpace(payload.Accion))
			if action == "apagar" || action == "desactivar" || action == "off" {
				enabled = false
			}
		}

		usuario := strings.TrimSpace(adminEmailFromRequest(r))
		if err := setChatFlotanteEmpresaPref(dbEmp, empresaID, chatFlotanteRadioOnlineEnabledKey, chatFlotanteBoolValue(enabled), usuario); err != nil {
			http.Error(w, "No se pudo guardar la configuracion de radio", http.StatusInternalServerError)
			return
		}
		if payload.RadioCountry != nil {
			value := normalizeChatFlotanteRadioCountry(*payload.RadioCountry)
			if err := setChatFlotanteEmpresaPref(dbEmp, empresaID, chatFlotanteRadioCountryKey, value, usuario); err != nil {
				http.Error(w, "No se pudo guardar la configuracion de radio", http.StatusInternalServerError)
				return
			}
		}

		resp := chatFlotantePrefsResponse(dbSuper, dbEmp, empresaID)
		resp["ok"] = true
		resp["accion"] = "radio_online"
		writeJSON(w, http.StatusOK, resp)
	}
}
