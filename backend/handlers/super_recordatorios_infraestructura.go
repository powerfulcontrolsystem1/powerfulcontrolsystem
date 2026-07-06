package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const superInfraReminderConfigKey = "super.recordatorios_infraestructura.items_json"

type superInfraReminderItem struct {
	ID               string `json:"id"`
	Tipo             string `json:"tipo"`
	Nombre           string `json:"nombre"`
	Proveedor        string `json:"proveedor,omitempty"`
	FechaVencimiento string `json:"fecha_vencimiento"`
	DiasAviso        int    `json:"dias_aviso"`
	Email            string `json:"email,omitempty"`
	WhatsApp         string `json:"whatsapp,omitempty"`
	EmailEnabled     bool   `json:"email_enabled"`
	WhatsAppEnabled  bool   `json:"whatsapp_enabled"`
	Activo           bool   `json:"activo"`
	Notas            string `json:"notas,omitempty"`
}

func SuperRecordatoriosInfraestructuraHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			items, err := loadSuperInfraReminderItems(dbSuper)
			if err != nil {
				http.Error(w, "No se pudieron cargar los recordatorios", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "items": items})
		case http.MethodPut, http.MethodPost:
			var payload struct {
				Items []superInfraReminderItem `json:"items"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			items, err := normalizeSuperInfraReminderItems(payload.Items)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			raw, _ := json.Marshal(items)
			if err := dbpkg.SetConfigValue(dbSuper, superInfraReminderConfigKey, string(raw), false); err != nil {
				http.Error(w, "No se pudieron guardar los recordatorios", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "items": items})
		default:
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
		}
	}
}

func loadSuperInfraReminderItems(dbSuper *sql.DB) ([]superInfraReminderItem, error) {
	raw, _, _, _, err := dbpkg.GetConfigEntry(dbSuper, superInfraReminderConfigKey)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(raw) == "" {
		return []superInfraReminderItem{}, nil
	}
	var items []superInfraReminderItem
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return []superInfraReminderItem{}, nil
	}
	out, err := normalizeSuperInfraReminderItems(items)
	if err != nil {
		return []superInfraReminderItem{}, nil
	}
	return out, nil
}

func normalizeSuperInfraReminderItems(input []superInfraReminderItem) ([]superInfraReminderItem, error) {
	out := make([]superInfraReminderItem, 0, len(input))
	seen := map[string]bool{}
	for _, item := range input {
		item.ID = strings.TrimSpace(item.ID)
		if item.ID == "" {
			item.ID = fmt.Sprintf("recordatorio_%d", time.Now().UnixNano()+int64(len(out)))
		}
		if seen[item.ID] {
			continue
		}
		seen[item.ID] = true
		item.Tipo = strings.ToLower(strings.TrimSpace(item.Tipo))
		if item.Tipo == "" {
			item.Tipo = "dominio"
		}
		item.Nombre = strings.TrimSpace(item.Nombre)
		if item.Nombre == "" {
			return nil, fmt.Errorf("nombre es obligatorio")
		}
		item.FechaVencimiento = strings.TrimSpace(item.FechaVencimiento)
		if _, err := time.Parse("2006-01-02", item.FechaVencimiento); err != nil {
			return nil, fmt.Errorf("fecha de vencimiento invalida para %s", item.Nombre)
		}
		if item.DiasAviso <= 0 {
			item.DiasAviso = 30
		}
		if item.DiasAviso > 365 {
			item.DiasAviso = 365
		}
		item.Email = strings.ToLower(strings.TrimSpace(item.Email))
		if item.Email != "" {
			if _, err := mail.ParseAddress(item.Email); err != nil {
				return nil, fmt.Errorf("email invalido para %s", item.Nombre)
			}
		}
		item.WhatsApp = normalizeWhatsAppPhone(item.WhatsApp)
		item.Proveedor = strings.TrimSpace(item.Proveedor)
		item.Notas = strings.TrimSpace(item.Notas)
		out = append(out, item)
	}
	return out, nil
}
