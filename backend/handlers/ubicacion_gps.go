package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaUbicacionGPSDispositivosHandler gestiona CRUD de dispositivos GPS por empresa.
func EmpresaUbicacionGPSDispositivosHandler(dbEmp *sql.DB, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := dbpkg.EnsureEmpresaUbicacionGPSSchema(dbEmp); err != nil {
			http.Error(w, "No se pudo preparar el modulo GPS", http.StatusInternalServerError)
			return
		}
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			includeInactive := queryBool(r, "include_inactive")
			q := strings.TrimSpace(r.URL.Query().Get("q"))
			rows, err := dbpkg.GetEmpresaGPSDispositivos(dbEmp, empresaID, includeInactive, q)
			if err != nil {
				http.Error(w, "No se pudieron listar los dispositivos GPS", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, rows)
			return

		case http.MethodPost:
			var payload dbpkg.EmpresaGPSDispositivo
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
			payload.Nombre = strings.TrimSpace(payload.Nombre)
			if payload.Nombre == "" {
				http.Error(w, "nombre es obligatorio", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.Codigo) == "" {
				payload.Codigo = fmt.Sprintf("GPS-%d-%d", payload.EmpresaID, time.Now().Unix())
			}
			maxGPS, err := MaxGPSDispositivosPorEmpresa(dbSuper)
			if err != nil {
				http.Error(w, "No se pudo validar limite de dispositivos GPS: "+err.Error(), http.StatusInternalServerError)
				return
			}
			existing, err := dbpkg.CountEmpresaGPSDispositivos(dbEmp, payload.EmpresaID)
			if err != nil {
				http.Error(w, "No se pudo contar dispositivos GPS", http.StatusInternalServerError)
				return
			}
			if maxGPS >= 0 && existing >= maxGPS {
				http.Error(w, fmt.Sprintf("La empresa alcanzo el maximo de dispositivos GPS permitidos (%d). Elimina uno inactivo o solicita al super administrador un mayor tope en configuracion avanzada.", maxGPS), http.StatusConflict)
				return
			}
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			id, err := dbpkg.CreateEmpresaGPSDispositivo(dbEmp, payload)
			if err != nil {
				http.Error(w, "No se pudo crear el dispositivo GPS", http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			return

		case http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "activar" || action == "desactivar" {
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
				if action == "desactivar" {
					estado = "inactivo"
				}
				if err := dbpkg.SetEmpresaGPSDispositivoEstado(dbEmp, empresaID, id, estado); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "dispositivo gps no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, "No se pudo actualizar el estado del dispositivo GPS", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": estado})
				return
			}

			var payload dbpkg.EmpresaGPSDispositivo
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.ID <= 0 || payload.EmpresaID <= 0 {
				http.Error(w, "id y empresa_id son obligatorios", http.StatusBadRequest)
				return
			}
			payload.Nombre = strings.TrimSpace(payload.Nombre)
			if payload.Nombre == "" {
				http.Error(w, "nombre es obligatorio", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.Codigo) == "" {
				payload.Codigo = fmt.Sprintf("GPS-%d-%d", payload.EmpresaID, payload.ID)
			}
			if err := dbpkg.UpdateEmpresaGPSDispositivo(dbEmp, payload); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "dispositivo gps no encontrado", http.StatusNotFound)
					return
				}
				http.Error(w, "No se pudo actualizar el dispositivo GPS", http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return

		case http.MethodDelete:
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
			if err := dbpkg.DeleteEmpresaGPSDispositivo(dbEmp, empresaID, id); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "dispositivo gps no encontrado", http.StatusNotFound)
					return
				}
				http.Error(w, "No se pudo eliminar el dispositivo GPS", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

// EmpresaUbicacionGPSRecorridosHandler gestiona registro y consulta de recorridos GPS.
func EmpresaUbicacionGPSRecorridosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := dbpkg.EnsureEmpresaUbicacionGPSSchema(dbEmp); err != nil {
			http.Error(w, "No se pudo preparar el modulo GPS", http.StatusInternalServerError)
			return
		}
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			dispositivoID, err := parseInt64QueryOptional(r, "dispositivo_id")
			if err != nil {
				http.Error(w, "dispositivo_id invalido", http.StatusBadRequest)
				return
			}
			includeInactive := queryBool(r, "include_inactive")
			desdeMinutos, err := parseIntQueryOptional(r, "desde_minutos")
			if err != nil {
				http.Error(w, "desde_minutos invalido", http.StatusBadRequest)
				return
			}
			if desdeMinutos < 0 {
				desdeMinutos = 0
			}

			limit, err := parseIntQueryOptional(r, "limit")
			if err != nil {
				http.Error(w, "limit invalido", http.StatusBadRequest)
				return
			}
			if limit <= 0 {
				limit = 600
			}
			rows, err := dbpkg.ListEmpresaGPSRecorridos(dbEmp, empresaID, dispositivoID, includeInactive, desdeMinutos, limit)
			if err != nil {
				http.Error(w, "No se pudieron listar los recorridos GPS", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, rows)
			return

		case http.MethodPost:
			var payload dbpkg.EmpresaGPSRecorrido
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 {
				if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
					payload.EmpresaID = empresaID
				}
			}
			if payload.EmpresaID <= 0 || payload.DispositivoID <= 0 {
				http.Error(w, "empresa_id y dispositivo_id son obligatorios", http.StatusBadRequest)
				return
			}
			if err := validateGPSCoordinates(payload.Latitud, payload.Longitud); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := validateGPSTelemetry(payload); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			if strings.TrimSpace(payload.Fuente) == "" {
				payload.Fuente = "manual"
			}
			id, err := dbpkg.CreateEmpresaGPSRecorrido(dbEmp, payload)
			if err != nil {
				http.Error(w, "No se pudo registrar el punto GPS", http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			return

		case http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "activar" || action == "desactivar" {
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
				if action == "desactivar" {
					estado = "inactivo"
				}
				if err := dbpkg.SetEmpresaGPSRecorridoEstado(dbEmp, empresaID, id, estado); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "punto gps no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, "No se pudo actualizar el estado del punto GPS", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": estado})
				return
			}

			var payload dbpkg.EmpresaGPSRecorrido
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.ID <= 0 || payload.EmpresaID <= 0 || payload.DispositivoID <= 0 {
				http.Error(w, "id, empresa_id y dispositivo_id son obligatorios", http.StatusBadRequest)
				return
			}
			if err := validateGPSCoordinates(payload.Latitud, payload.Longitud); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := validateGPSTelemetry(payload); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := dbpkg.UpdateEmpresaGPSRecorrido(dbEmp, payload); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "punto gps no encontrado", http.StatusNotFound)
					return
				}
				http.Error(w, "No se pudo actualizar el punto GPS", http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return

		case http.MethodDelete:
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
			if err := dbpkg.DeleteEmpresaGPSRecorrido(dbEmp, empresaID, id); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "punto gps no encontrado", http.StatusNotFound)
					return
				}
				http.Error(w, "No se pudo eliminar el punto GPS", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

func validateGPSCoordinates(lat, lng float64) error {
	if lat < -90 || lat > 90 {
		return fmt.Errorf("latitud fuera de rango")
	}
	if lng < -180 || lng > 180 {
		return fmt.Errorf("longitud fuera de rango")
	}
	return nil
}

func validateGPSTelemetry(p dbpkg.EmpresaGPSRecorrido) error {
	if p.PrecisionMetros < 0 || p.PrecisionMetros > 100000 {
		return fmt.Errorf("precision GPS fuera de rango")
	}
	if p.VelocidadKMH < 0 || p.VelocidadKMH > 1000 {
		return fmt.Errorf("velocidad GPS fuera de rango")
	}
	if p.RumboGrados < 0 || p.RumboGrados >= 360 {
		return fmt.Errorf("rumbo GPS fuera de rango")
	}
	if p.BateriaPorcentaje < 0 || p.BateriaPorcentaje > 100 || p.SenalPorcentaje < 0 || p.SenalPorcentaje > 100 {
		return fmt.Errorf("bateria o senal GPS fuera de rango")
	}
	if p.AltitudMetros < -11000 || p.AltitudMetros > 100000 {
		return fmt.Errorf("altitud GPS fuera de rango")
	}
	return nil
}
