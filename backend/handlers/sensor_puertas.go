package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

// PublicSensorPuertasHandler recibe heartbeats públicos desde dispositivos (Raspberry Pi)
// Usa query param `action=heartbeat` y método POST con JSON {"device_id":"...","state":"..."}
func PublicSensorPuertasHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		action := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("action")))
		switch action {
		case "heartbeat":
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			var payload struct {
				DeviceID string `json:"device_id"`
				State    string `json:"state"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "json invalido", http.StatusBadRequest)
				return
			}
			// Prefer token-based auth if header is present
			token := strings.TrimSpace(r.Header.Get("X-Device-Token"))
			if token != "" {
				dev, err := dbpkg.GetEmpresaSensorByToken(dbEmp, token)
				if err != nil {
					if err == sql.ErrNoRows {
						writeJSON(w, http.StatusNotFound, map[string]interface{}{"ok": false, "message": "device not registered"})
						return
					}
					log.Printf("[sensor_puertas] get device by token error: %v", err)
					http.Error(w, "error interno", http.StatusInternalServerError)
					return
				}
				empresaID, estacionID, err := dbpkg.UpdateDeviceHeartbeat(dbEmp, dev.DeviceID, payload.State)
				if err != nil {
					log.Printf("[sensor_puertas] heartbeat error: %v", err)
					http.Error(w, "error interno", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "empresa_id": empresaID, "estacion_id": estacionID})
				return
			}

			if strings.TrimSpace(payload.DeviceID) == "" {
				http.Error(w, "device_id obligatorio", http.StatusBadRequest)
				return
			}
			empresaID, estacionID, err := dbpkg.UpdateDeviceHeartbeat(dbEmp, payload.DeviceID, payload.State)
			if err != nil {
				if err == sql.ErrNoRows {
					writeJSON(w, http.StatusNotFound, map[string]interface{}{"ok": false, "message": "device not registered"})
					return
				}
				log.Printf("[sensor_puertas] heartbeat error: %v", err)
				http.Error(w, "error interno", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "empresa_id": empresaID, "estacion_id": estacionID})
			return
		case "message":
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			var payload struct {
				DeviceID string `json:"device_id"`
				Message  string `json:"message"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "json invalido", http.StatusBadRequest)
				return
			}
			// Prefer token-based auth if header present
			token := strings.TrimSpace(r.Header.Get("X-Device-Token"))
			if token != "" {
				dev, err := dbpkg.GetEmpresaSensorByToken(dbEmp, token)
				if err != nil {
					if err == sql.ErrNoRows {
						writeJSON(w, http.StatusNotFound, map[string]interface{}{"ok": false, "message": "device not registered"})
						return
					}
					log.Printf("[sensor_puertas] get device by token error: %v", err)
					http.Error(w, "error interno", http.StatusInternalServerError)
					return
				}
				msgID, empresaID, estacionID, err := dbpkg.InsertEmpresaSensorMessage(dbEmp, dev.DeviceID, payload.Message)
				if err != nil {
					if err == sql.ErrNoRows {
						writeJSON(w, http.StatusNotFound, map[string]interface{}{"ok": false, "message": "device not registered"})
						return
					}
					log.Printf("[sensor_puertas] insert message error: %v", err)
					http.Error(w, "error interno", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "message_id": msgID, "empresa_id": empresaID, "estacion_id": estacionID})
				return
			}

			if strings.TrimSpace(payload.DeviceID) == "" {
				http.Error(w, "device_id obligatorio", http.StatusBadRequest)
				return
			}
			msgID, empresaID, estacionID, err := dbpkg.InsertEmpresaSensorMessage(dbEmp, payload.DeviceID, payload.Message)
			if err != nil {
				if err == sql.ErrNoRows {
					writeJSON(w, http.StatusNotFound, map[string]interface{}{"ok": false, "message": "device not registered"})
					return
				}
				log.Printf("[sensor_puertas] insert message error: %v", err)
				http.Error(w, "error interno", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "message_id": msgID, "empresa_id": empresaID, "estacion_id": estacionID})
			return
		default:
			http.Error(w, "action no soportada", http.StatusBadRequest)
			return
		}
	}
}

// EmpresaSensorMessagesHandler lista los mensajes recibidos para una empresa (protegido)
func EmpresaSensorMessagesHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		messages, err := dbpkg.GetEmpresaSensorMessagesByEmpresa(dbEmp, empresaID)
		if err != nil {
			log.Printf("[sensor_puertas] get messages error: %v", err)
			http.Error(w, "error interno", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, messages)
	}
}

// EmpresaSensorConfigHandler gestiona CRUD ligero para asociar dispositivos a estaciones (protegido)
// GET -> lista dispositivos de la empresa
// POST/PUT -> crea/actualiza mapping { device_id, estacion_id }
func EmpresaSensorConfigHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			devices, err := dbpkg.GetEmpresaSensorsByEmpresa(dbEmp, empresaID)
			if err != nil {
				log.Printf("[sensor_puertas] get devices error: %v", err)
				http.Error(w, "error interno", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, devices)
			return
		case http.MethodPost, http.MethodPut:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			var payload struct {
				DeviceID    string `json:"device_id"`
				EstacionID  int64  `json:"estacion_id,omitempty"`
				DeviceToken string `json:"device_token,omitempty"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "json invalido", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.DeviceID) == "" {
				http.Error(w, "device_id obligatorio", http.StatusBadRequest)
				return
			}
			p := dbpkg.EmpresaSensorDevice{
				EmpresaID:      empresaID,
				DeviceID:       payload.DeviceID,
				DeviceToken:    payload.DeviceToken,
				EstacionID:     payload.EstacionID,
				UsuarioCreador: strings.TrimSpace(adminEmailFromRequest(r)),
				Estado:         "activo",
			}
			id, err := dbpkg.UpsertEmpresaSensorDevice(dbEmp, &p)
			if err != nil {
				log.Printf("[sensor_puertas] upsert error: %v", err)
				http.Error(w, "error interno", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
			return
		default:
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
	}
}
