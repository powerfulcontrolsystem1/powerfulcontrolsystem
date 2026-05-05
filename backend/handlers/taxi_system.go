package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaTaxiSystemHandler(dbEmp *sql.DB, dbSuper ...*sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		adminEmail := strings.TrimSpace(adminEmailFromRequest(r))

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "", "dashboard":
				row, err := dbpkg.BuildEmpresaTaxiDashboard(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar el dashboard taxi", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "config":
				row, err := dbpkg.GetEmpresaTaxiConfig(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar la configuracion taxi", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "drivers":
				onlyOnline := queryBool(r, "online")
				rows, err := dbpkg.ListEmpresaTaxiDrivers(dbEmp, empresaID, onlyOnline)
				if err != nil {
					http.Error(w, "No se pudieron listar los conductores", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "requests":
				state := strings.TrimSpace(r.URL.Query().Get("estado"))
				rows, err := dbpkg.ListTaxiRequests(dbEmp, empresaID, state, 100)
				if err != nil {
					http.Error(w, "No se pudieron listar los servicios", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "request":
				requestID, err := parseInt64Query(r, "request_id")
				if err != nil {
					http.Error(w, "request_id es obligatorio", http.StatusBadRequest)
					return
				}
				row, err := dbpkg.GetTaxiRequestByID(dbEmp, empresaID, requestID)
				if err != nil {
					http.Error(w, "No se pudo consultar el servicio", http.StatusNotFound)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "route":
				requestID, err := parseInt64Query(r, "request_id")
				if err != nil {
					http.Error(w, "request_id es obligatorio", http.StatusBadRequest)
					return
				}
				rows, err := dbpkg.ListTaxiRoutePoints(dbEmp, empresaID, requestID, 500)
				if err != nil {
					http.Error(w, "No se pudo consultar la ruta", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "gps_devices":
				if err := dbpkg.EnsureEmpresaUbicacionGPSSchema(dbEmp); err != nil {
					http.Error(w, "No se pudo preparar el modulo GPS", http.StatusInternalServerError)
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
			}

		case http.MethodPost:
			switch action {
			case "config":
				var payload dbpkg.EmpresaTaxiConfig
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				if err := dbpkg.UpsertEmpresaTaxiConfig(dbEmp, payload); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "drivers":
				var payload dbpkg.EmpresaTaxiDriver
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				id, err := dbpkg.CreateEmpresaTaxiDriver(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "dispatch":
				requestID, err := parseInt64QueryOptional(r, "request_id")
				if err != nil || requestID <= 0 {
					var payload struct {
						RequestID int64 `json:"request_id"`
					}
					_ = json.NewDecoder(r.Body).Decode(&payload)
					requestID = payload.RequestID
				}
				if requestID <= 0 {
					http.Error(w, "request_id es obligatorio", http.StatusBadRequest)
					return
				}
				rows, err := dbpkg.DispatchTaxiRequestToNearbyDrivers(dbEmp, empresaID, requestID, 0)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "request_state":
				var payload struct {
					RequestID int64  `json:"request_id"`
					State     string `json:"state"`
					Notes     string `json:"notes"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if payload.RequestID <= 0 {
					http.Error(w, "request_id es obligatorio", http.StatusBadRequest)
					return
				}
				row, err := dbpkg.UpdateTaxiRequestState(dbEmp, empresaID, payload.RequestID, 0, payload.State, payload.Notes)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "gps_devices":
				if err := dbpkg.EnsureEmpresaUbicacionGPSSchema(dbEmp); err != nil {
					http.Error(w, "No se pudo preparar el modulo GPS", http.StatusInternalServerError)
					return
				}
				var payload dbpkg.EmpresaGPSDispositivo
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.Nombre = strings.TrimSpace(payload.Nombre)
				if payload.Nombre == "" {
					http.Error(w, "nombre es obligatorio", http.StatusBadRequest)
					return
				}
				if len(dbSuper) > 0 && dbSuper[0] != nil {
					maxGPS, err := MaxGPSDispositivosPorEmpresa(dbSuper[0])
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
						http.Error(w, fmt.Sprintf("La empresa alcanzo el maximo de dispositivos GPS permitidos (%d).", maxGPS), http.StatusConflict)
						return
					}
				}
				payload.UsuarioCreador = adminEmail
				id, err := dbpkg.CreateEmpresaGPSDispositivo(dbEmp, payload)
				if err != nil {
					http.Error(w, "No se pudo crear el dispositivo GPS", http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			}

		case http.MethodPut:
			switch action {
			case "drivers":
				var payload dbpkg.EmpresaTaxiDriver
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				if err := dbpkg.UpdateEmpresaTaxiDriver(dbEmp, payload); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "gps_devices":
				if err := dbpkg.EnsureEmpresaUbicacionGPSSchema(dbEmp); err != nil {
					http.Error(w, "No se pudo preparar el modulo GPS", http.StatusInternalServerError)
					return
				}
				var payload dbpkg.EmpresaGPSDispositivo
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.Nombre = strings.TrimSpace(payload.Nombre)
				if payload.ID <= 0 || payload.Nombre == "" {
					http.Error(w, "id y nombre son obligatorios", http.StatusBadRequest)
					return
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
			}
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

func PublicTaxiSystemHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "", "config":
				row, err := dbpkg.GetEmpresaTaxiConfig(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar la configuracion", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "request":
				requestID, err := parseInt64Query(r, "request_id")
				if err != nil {
					http.Error(w, "request_id es obligatorio", http.StatusBadRequest)
					return
				}
				row, err := dbpkg.GetTaxiRequestByID(dbEmp, empresaID, requestID)
				if err != nil {
					http.Error(w, "No se pudo consultar el servicio", http.StatusNotFound)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "route":
				requestID, err := parseInt64Query(r, "request_id")
				if err != nil {
					http.Error(w, "request_id es obligatorio", http.StatusBadRequest)
					return
				}
				rows, err := dbpkg.ListTaxiRoutePoints(dbEmp, empresaID, requestID, 500)
				if err != nil {
					http.Error(w, "No se pudo consultar la ruta", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			}

		case http.MethodPost:
			switch action {
			case "register_customer":
				cfg, err := dbpkg.GetEmpresaTaxiConfig(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo validar la configuracion", http.StatusInternalServerError)
					return
				}
				if !cfg.PermitirRegistroCliente {
					http.Error(w, "El registro de clientes esta deshabilitado", http.StatusForbidden)
					return
				}
				var payload dbpkg.EmpresaTaxiCustomer
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				row, err := dbpkg.RegisterTaxiCustomer(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, row)
				return
			case "login_customer":
				var payload struct {
					Telefono string `json:"telefono"`
					Pin      string `json:"pin"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				row, err := dbpkg.TaxiCustomerLogin(dbEmp, empresaID, payload.Telefono, payload.Pin)
				if err != nil {
					status := http.StatusBadRequest
					if errors.Is(err, dbpkg.ErrTaxiCustomerAuthInvalid) {
						status = http.StatusUnauthorized
					}
					http.Error(w, err.Error(), status)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "request_service":
				var payload dbpkg.EmpresaTaxiRequest
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				if token := strings.TrimSpace(r.Header.Get("X-Taxi-Customer-Token")); token != "" {
					if customer, err := dbpkg.ResolveTaxiCustomerByToken(dbEmp, empresaID, token); err == nil {
						payload.CustomerID = customer.ID
						if strings.TrimSpace(payload.ClienteNombre) == "" {
							payload.ClienteNombre = customer.Nombre
						}
						if strings.TrimSpace(payload.ClienteTelefono) == "" {
							payload.ClienteTelefono = customer.Telefono
						}
						if strings.TrimSpace(payload.ClienteDocumento) == "" {
							payload.ClienteDocumento = customer.Documento
						}
					}
				}
				row, err := dbpkg.CreateTaxiRequest(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, row)
				return
			case "share_location":
				var payload struct {
					RequestID       int64   `json:"request_id"`
					CustomerToken   string  `json:"customer_token"`
					Latitud         float64 `json:"latitud"`
					Longitud        float64 `json:"longitud"`
					PrecisionMetros float64 `json:"precision_metros"`
					VelocidadKMH    float64 `json:"velocidad_kmh"`
					RumboGrados     float64 `json:"rumbo_grados"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				customer, err := dbpkg.ResolveTaxiCustomerByToken(dbEmp, empresaID, firstNonEmptyString(payload.CustomerToken, r.Header.Get("X-Taxi-Customer-Token")))
				if err != nil {
					http.Error(w, "No se pudo validar el cliente", http.StatusUnauthorized)
					return
				}
				err = dbpkg.AddTaxiCustomerRoutePoint(dbEmp, empresaID, payload.RequestID, customer.ID, dbpkg.EmpresaTaxiRoutePoint{
					Latitud:         payload.Latitud,
					Longitud:        payload.Longitud,
					PrecisionMetros: payload.PrecisionMetros,
					VelocidadKMH:    payload.VelocidadKMH,
					RumboGrados:     payload.RumboGrados,
				})
				if err != nil {
					http.Error(w, "No se pudo registrar la ubicacion del cliente", http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "login_driver":
				var payload struct {
					Documento string `json:"documento"`
					Pin       string `json:"pin"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				row, err := dbpkg.TaxiDriverLogin(dbEmp, empresaID, payload.Documento, payload.Pin)
				if err != nil {
					status := http.StatusBadRequest
					if errors.Is(err, dbpkg.ErrTaxiDriverAuthInvalid) {
						status = http.StatusUnauthorized
					}
					http.Error(w, err.Error(), status)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "driver_presence":
				var payload struct {
					Token      string `json:"token"`
					Online     bool   `json:"online"`
					Disponible bool   `json:"disponible"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				driver, err := dbpkg.ResolveTaxiDriverByToken(dbEmp, empresaID, firstNonEmptyString(payload.Token, r.Header.Get("X-Taxi-Driver-Token")))
				if err != nil {
					http.Error(w, "No se pudo validar el conductor", http.StatusUnauthorized)
					return
				}
				if err := dbpkg.UpdateTaxiDriverPresence(dbEmp, empresaID, driver.ID, payload.Online, payload.Disponible); err != nil {
					http.Error(w, "No se pudo actualizar la presencia", http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "driver_location":
				var payload struct {
					Token           string  `json:"token"`
					RequestID       int64   `json:"request_id"`
					Latitud         float64 `json:"latitud"`
					Longitud        float64 `json:"longitud"`
					PrecisionMetros float64 `json:"precision_metros"`
					VelocidadKMH    float64 `json:"velocidad_kmh"`
					RumboGrados     float64 `json:"rumbo_grados"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				driver, err := dbpkg.ResolveTaxiDriverByToken(dbEmp, empresaID, firstNonEmptyString(payload.Token, r.Header.Get("X-Taxi-Driver-Token")))
				if err != nil {
					http.Error(w, "No se pudo validar el conductor", http.StatusUnauthorized)
					return
				}
				if err := dbpkg.UpdateTaxiDriverLocation(dbEmp, empresaID, driver.ID, payload.RequestID, dbpkg.EmpresaTaxiRoutePoint{
					Latitud:         payload.Latitud,
					Longitud:        payload.Longitud,
					PrecisionMetros: payload.PrecisionMetros,
					VelocidadKMH:    payload.VelocidadKMH,
					RumboGrados:     payload.RumboGrados,
				}); err != nil {
					http.Error(w, "No se pudo registrar la ubicacion", http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "driver_offers":
				token := strings.TrimSpace(r.Header.Get("X-Taxi-Driver-Token"))
				var payload struct {
					Token string `json:"token"`
				}
				_ = json.NewDecoder(r.Body).Decode(&payload)
				driver, err := dbpkg.ResolveTaxiDriverByToken(dbEmp, empresaID, firstNonEmptyString(payload.Token, token))
				if err != nil {
					http.Error(w, "No se pudo validar el conductor", http.StatusUnauthorized)
					return
				}
				rows, err := dbpkg.ListTaxiOffersForDriver(dbEmp, empresaID, driver.ID)
				if err != nil {
					http.Error(w, "No se pudieron consultar las ofertas", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "respond_offer":
				var payload struct {
					Token         string `json:"token"`
					OfferID       int64  `json:"offer_id"`
					Accept        bool   `json:"accept"`
					Observaciones string `json:"observaciones"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				driver, err := dbpkg.ResolveTaxiDriverByToken(dbEmp, empresaID, firstNonEmptyString(payload.Token, r.Header.Get("X-Taxi-Driver-Token")))
				if err != nil {
					http.Error(w, "No se pudo validar el conductor", http.StatusUnauthorized)
					return
				}
				row, err := dbpkg.RespondTaxiOffer(dbEmp, empresaID, payload.OfferID, driver.ID, payload.Accept, payload.Observaciones)
				if err != nil {
					status := http.StatusBadRequest
					if errors.Is(err, dbpkg.ErrTaxiOfferUnavailable) {
						status = http.StatusConflict
					}
					http.Error(w, err.Error(), status)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "driver_request_state":
				var payload struct {
					Token     string `json:"token"`
					RequestID int64  `json:"request_id"`
					State     string `json:"state"`
					Notes     string `json:"notes"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				driver, err := dbpkg.ResolveTaxiDriverByToken(dbEmp, empresaID, firstNonEmptyString(payload.Token, r.Header.Get("X-Taxi-Driver-Token")))
				if err != nil {
					http.Error(w, "No se pudo validar el conductor", http.StatusUnauthorized)
					return
				}
				row, err := dbpkg.UpdateTaxiRequestState(dbEmp, empresaID, payload.RequestID, driver.ID, payload.State, payload.Notes)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			}
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}
