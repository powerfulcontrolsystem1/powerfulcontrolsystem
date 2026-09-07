package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaHotelTarjetasAccesoHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if action == "detalle" || action == "get" {
				id, err := parseInt64Query(r, "id")
				if err != nil {
					http.Error(w, "id es obligatorio", http.StatusBadRequest)
					return
				}
				item, err := dbpkg.GetHotelTarjetaAccesoByID(dbEmp, empresaID, id)
				if err != nil {
					writeHotelTarjetasError(w, err)
					return
				}
				writeJSON(w, http.StatusOK, item)
				return
			}
			filter := dbpkg.HotelTarjetaAccesoFilter{IncludeInactive: queryBool(r, "include_inactive")}
			filter.EstacionID, _ = parseInt64QueryOptional(r, "estacion_id")
			filter.ReservaID, _ = parseInt64QueryOptional(r, "reserva_id")
			filter.Limit, _ = parseIntQueryOptional(r, "limit")
			rows, err := dbpkg.ListHotelTarjetasAcceso(dbEmp, empresaID, filter)
			if err != nil {
				writeHotelTarjetasError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, rows)
			return
		case http.MethodPost:
			if action == "validar" || action == "validate" {
				handleHotelTarjetaValidate(w, r, dbEmp)
				return
			}
			var payload dbpkg.HotelTarjetaAcceso
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 {
				if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil {
					payload.EmpresaID = empresaID
				}
			}
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			id, accessCode, err := dbpkg.CreateHotelTarjetaAcceso(dbEmp, payload)
			if err != nil {
				writeHotelTarjetasError(w, err)
				return
			}
			item, err := dbpkg.GetHotelTarjetaAccesoByID(dbEmp, payload.EmpresaID, id)
			if err != nil {
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id, "access_code": accessCode})
				return
			}
			item.AccessCode = accessCode
			writeJSON(w, http.StatusCreated, item)
			return
		case http.MethodPut:
			if action == "activar" || action == "desactivar" || action == "bloquear" || action == "revocar" {
				empresaID, err := parseEmpresaIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				id, err := parseInt64Query(r, "id")
				if err != nil {
					http.Error(w, "id es obligatorio", http.StatusBadRequest)
					return
				}
				estado := "activo"
				if action != "activar" {
					estado = "inactivo"
				}
				if err := dbpkg.SetHotelTarjetaAccesoEstado(dbEmp, empresaID, id, estado); err != nil {
					writeHotelTarjetasError(w, err)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": estado})
				return
			}
			var payload dbpkg.HotelTarjetaAcceso
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 {
				if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil {
					payload.EmpresaID = empresaID
				}
			}
			if payload.ID <= 0 {
				if id, err := parseInt64QueryOptional(r, "id"); err == nil {
					payload.ID = id
				}
			}
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			if err := dbpkg.UpdateHotelTarjetaAcceso(dbEmp, payload); err != nil {
				writeHotelTarjetasError(w, err)
				return
			}
			item, err := dbpkg.GetHotelTarjetaAccesoByID(dbEmp, payload.EmpresaID, payload.ID)
			if err != nil {
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			}
			writeJSON(w, http.StatusOK, item)
			return
		default:
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
	}
}

func PublicHotelTarjetasAccesoHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		handleHotelTarjetaValidate(w, r, dbEmp)
	}
}

func handleHotelTarjetaValidate(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	var payload struct {
		EmpresaID  int64  `json:"empresa_id"`
		EstacionID int64  `json:"estacion_id"`
		CardUID    string `json:"card_uid"`
		AccessCode string `json:"access_code"`
		DeviceID   string `json:"device_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID <= 0 {
		if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil {
			payload.EmpresaID = empresaID
		}
	}
	if payload.EstacionID <= 0 {
		payload.EstacionID, _ = parseInt64QueryOptional(r, "estacion_id")
	}
	result, err := dbpkg.ValidateHotelTarjetaAcceso(dbEmp, payload.EmpresaID, payload.EstacionID, payload.CardUID, payload.AccessCode, payload.DeviceID)
	if err != nil {
		writeHotelTarjetasError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func writeHotelTarjetasError(w http.ResponseWriter, err error) {
	if err == nil {
		http.Error(w, "error no especificado", http.StatusInternalServerError)
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "tarjeta no encontrada", http.StatusNotFound)
		return
	}
	msg := strings.TrimSpace(err.Error())
	lower := strings.ToLower(msg)
	if strings.Contains(lower, "obligatorio") || strings.Contains(lower, "invalido") || strings.Contains(lower, "debe") || strings.Contains(lower, "unique") {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	http.Error(w, "No se pudo procesar tarjetas de hotel", http.StatusInternalServerError)
}
