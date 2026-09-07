package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaDomiciliosHandler(dbEmp *sql.DB) http.HandlerFunc {
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
				row, err := dbpkg.BuildEmpresaDomiciliosDashboard(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar el dashboard de domicilios", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "config":
				row, err := dbpkg.GetEmpresaDomiciliosConfig(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar la configuracion", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "restaurants":
				rows, err := dbpkg.ListDomicilioRestaurants(dbEmp, empresaID, queryBool(r, "active"))
				if err != nil {
					http.Error(w, "No se pudieron listar restaurantes", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "couriers":
				rows, err := dbpkg.ListDomicilioCouriers(dbEmp, empresaID, queryBool(r, "online"))
				if err != nil {
					http.Error(w, "No se pudieron listar domiciliarios", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "menu":
				restaurantID, _ := parseInt64QueryOptional(r, "restaurant_id")
				rows, err := dbpkg.ListDomicilioMenuItems(dbEmp, empresaID, restaurantID, queryBool(r, "available"))
				if err != nil {
					http.Error(w, "No se pudo listar el menu", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "orders":
				rows, err := dbpkg.ListDomicilioOrders(dbEmp, empresaID, strings.TrimSpace(r.URL.Query().Get("estado")), 150)
				if err != nil {
					http.Error(w, "No se pudieron listar pedidos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "order":
				id, err := parseInt64Query(r, "order_id")
				if err != nil {
					http.Error(w, "order_id es obligatorio", http.StatusBadRequest)
					return
				}
				row, err := dbpkg.GetDomicilioOrderByID(dbEmp, empresaID, id)
				if err != nil {
					http.Error(w, "Pedido no encontrado", http.StatusNotFound)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "tracking":
				id, err := parseInt64Query(r, "order_id")
				if err != nil {
					http.Error(w, "order_id es obligatorio", http.StatusBadRequest)
					return
				}
				rows, err := dbpkg.ListDomicilioTracking(dbEmp, empresaID, id, 600)
				if err != nil {
					http.Error(w, "No se pudo consultar tracking", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			}
		case http.MethodPost:
			switch action {
			case "config":
				var payload dbpkg.EmpresaDomiciliosConfig
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.UsuarioCreador = empresaID, adminEmail
				if err := dbpkg.UpsertEmpresaDomiciliosConfig(dbEmp, payload); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "restaurants":
				var payload dbpkg.EmpresaDomicilioRestaurant
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.UsuarioCreador = empresaID, adminEmail
				id, err := dbpkg.CreateDomicilioRestaurant(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "couriers":
				var payload dbpkg.EmpresaDomicilioCourier
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.UsuarioCreador = empresaID, adminEmail
				id, err := dbpkg.CreateDomicilioCourier(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "menu":
				var payload dbpkg.EmpresaDomicilioMenuItem
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				id, err := dbpkg.UpsertDomicilioMenuItem(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			case "order":
				var payload dbpkg.EmpresaDomicilioOrder
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.Canal = empresaID, "central"
				row, err := dbpkg.CreateDomicilioOrder(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, row)
				return
			case "dispatch":
				var payload struct {
					OrderID int64 `json:"order_id"`
				}
				_ = json.NewDecoder(r.Body).Decode(&payload)
				if payload.OrderID <= 0 {
					if id, err := parseInt64QueryOptional(r, "order_id"); err == nil {
						payload.OrderID = id
					}
				}
				if payload.OrderID <= 0 {
					http.Error(w, "order_id es obligatorio", http.StatusBadRequest)
					return
				}
				rows, err := dbpkg.DispatchDomicilioOrder(dbEmp, empresaID, payload.OrderID, 0)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "order_state":
				var payload struct {
					OrderID       int64  `json:"order_id"`
					State         string `json:"state"`
					Notes         string `json:"notes"`
					CodigoEntrega string `json:"codigo_entrega"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				row, err := dbpkg.UpdateDomicilioOrderState(dbEmp, empresaID, payload.OrderID, 0, "central", payload.State, payload.Notes, payload.CodigoEntrega)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if strings.EqualFold(strings.TrimSpace(payload.State), "listo") {
					_, _ = dbpkg.DispatchDomicilioOrder(dbEmp, empresaID, payload.OrderID, 0)
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "seed_demo":
				if err := dbpkg.SeedEmpresaDomiciliosDemo(dbEmp, empresaID, adminEmail); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			}
		case http.MethodPut:
			switch action {
			case "restaurants":
				var payload dbpkg.EmpresaDomicilioRestaurant
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				if err := dbpkg.UpdateDomicilioRestaurant(dbEmp, payload); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "couriers":
				var payload dbpkg.EmpresaDomicilioCourier
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				if err := dbpkg.UpdateDomicilioCourier(dbEmp, payload); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			}
		}
		http.Error(w, "accion o metodo no soportado", http.StatusMethodNotAllowed)
	}
}

func PublicDomiciliosHandler(dbEmp *sql.DB) http.HandlerFunc {
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
			case "", "catalog":
				cfg, err := dbpkg.GetEmpresaDomiciliosConfig(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar configuracion", http.StatusInternalServerError)
					return
				}
				rests, _ := dbpkg.ListDomicilioRestaurants(dbEmp, empresaID, true)
				menu, _ := dbpkg.ListDomicilioMenuItems(dbEmp, empresaID, 0, true)
				writeJSON(w, http.StatusOK, map[string]interface{}{"config": cfg, "restaurants": rests, "menu": menu})
				return
			case "order":
				id, err := parseInt64Query(r, "order_id")
				if err != nil {
					http.Error(w, "order_id es obligatorio", http.StatusBadRequest)
					return
				}
				token := strings.TrimSpace(r.URL.Query().Get("token"))
				row, err := dbpkg.GetDomicilioOrderByCustomerToken(dbEmp, empresaID, id, token)
				if err != nil {
					http.Error(w, "Pedido no encontrado", http.StatusNotFound)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "tracking":
				id, err := parseInt64Query(r, "order_id")
				if err != nil {
					http.Error(w, "order_id es obligatorio", http.StatusBadRequest)
					return
				}
				rows, err := dbpkg.ListDomicilioTracking(dbEmp, empresaID, id, 500)
				if err != nil {
					http.Error(w, "No se pudo consultar tracking", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			}
		case http.MethodPost:
			switch action {
			case "order":
				var payload dbpkg.EmpresaDomicilioOrder
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.Canal = empresaID, "web"
				row, err := dbpkg.CreateDomicilioOrder(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, row)
				return
			case "courier_login":
				var payload struct {
					Documento string `json:"documento"`
					Pin       string `json:"pin"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				row, err := dbpkg.DomicilioCourierLogin(dbEmp, empresaID, payload.Documento, payload.Pin)
				if err != nil {
					status := http.StatusBadRequest
					if errors.Is(err, dbpkg.ErrDomicilioCourierAuthInvalid) {
						status = http.StatusUnauthorized
					}
					http.Error(w, err.Error(), status)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "restaurant_login":
				var payload struct {
					Codigo string `json:"codigo"`
					Pin    string `json:"pin"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				row, err := dbpkg.DomicilioRestaurantLogin(dbEmp, empresaID, payload.Codigo, payload.Pin)
				if err != nil {
					status := http.StatusBadRequest
					if errors.Is(err, dbpkg.ErrDomicilioRestaurantAuthInvalid) {
						status = http.StatusUnauthorized
					}
					http.Error(w, err.Error(), status)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "courier_presence":
				courier, ok := resolveDomicilioCourierFromRequest(w, r, dbEmp, empresaID)
				if !ok {
					return
				}
				var payload struct {
					Online     bool `json:"online"`
					Disponible bool `json:"disponible"`
				}
				_ = json.NewDecoder(r.Body).Decode(&payload)
				if err := dbpkg.UpdateDomicilioCourierPresence(dbEmp, empresaID, courier.ID, payload.Online, payload.Disponible); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "courier_location":
				courier, ok := resolveDomicilioCourierFromRequest(w, r, dbEmp, empresaID)
				if !ok {
					return
				}
				var payload struct {
					OrderID         int64   `json:"order_id"`
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
				err := dbpkg.UpdateDomicilioCourierLocation(dbEmp, empresaID, courier.ID, payload.OrderID, dbpkg.EmpresaDomicilioTrackPoint{Latitud: payload.Latitud, Longitud: payload.Longitud, PrecisionMetros: payload.PrecisionMetros, VelocidadKMH: payload.VelocidadKMH, RumboGrados: payload.RumboGrados})
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "courier_offers":
				courier, ok := resolveDomicilioCourierFromRequest(w, r, dbEmp, empresaID)
				if !ok {
					return
				}
				rows, err := dbpkg.ListDomicilioOffersForCourier(dbEmp, empresaID, courier.ID)
				if err != nil {
					http.Error(w, "No se pudieron consultar ofertas", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "courier_orders":
				courier, ok := resolveDomicilioCourierFromRequest(w, r, dbEmp, empresaID)
				if !ok {
					return
				}
				rows, err := dbpkg.ListDomicilioOrdersForCourier(dbEmp, empresaID, courier.ID)
				if err != nil {
					http.Error(w, "No se pudieron consultar pedidos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "respond_offer":
				courier, ok := resolveDomicilioCourierFromRequest(w, r, dbEmp, empresaID)
				if !ok {
					return
				}
				var payload struct {
					OfferID       int64  `json:"offer_id"`
					Accept        bool   `json:"accept"`
					Observaciones string `json:"observaciones"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				row, err := dbpkg.RespondDomicilioOffer(dbEmp, empresaID, payload.OfferID, courier.ID, payload.Accept, payload.Observaciones)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "order_state":
				var payload struct {
					OrderID       int64  `json:"order_id"`
					State         string `json:"state"`
					Notes         string `json:"notes"`
					CodigoEntrega string `json:"codigo_entrega"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				actorTipo := "publico"
				var actorID int64
				if c, err := dbpkg.ResolveDomicilioCourierByToken(dbEmp, empresaID, strings.TrimSpace(r.Header.Get("X-Domicilio-Courier-Token"))); err == nil {
					actorTipo, actorID = "domiciliario", c.ID
				}
				if actorTipo == "publico" {
					if rest, err := dbpkg.ResolveDomicilioRestaurantByToken(dbEmp, empresaID, strings.TrimSpace(r.Header.Get("X-Domicilio-Restaurant-Token"))); err == nil {
						actorTipo, actorID = "restaurante", rest.ID
					}
				}
				row, err := dbpkg.UpdateDomicilioOrderState(dbEmp, empresaID, payload.OrderID, actorID, actorTipo, payload.State, payload.Notes, payload.CodigoEntrega)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if strings.EqualFold(strings.TrimSpace(payload.State), "listo") {
					_, _ = dbpkg.DispatchDomicilioOrder(dbEmp, empresaID, payload.OrderID, 0)
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "restaurant_orders":
				rest, ok := resolveDomicilioRestaurantFromRequest(w, r, dbEmp, empresaID)
				if !ok {
					return
				}
				rows, err := dbpkg.ListDomicilioOrdersForRestaurant(dbEmp, empresaID, rest.ID)
				if err != nil {
					http.Error(w, "No se pudieron consultar pedidos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			}
		}
		http.Error(w, "accion o metodo no soportado", http.StatusMethodNotAllowed)
	}
}

func resolveDomicilioCourierFromRequest(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, empresaID int64) (dbpkg.EmpresaDomicilioCourier, bool) {
	token := strings.TrimSpace(r.Header.Get("X-Domicilio-Courier-Token"))
	if token == "" {
		token = strings.TrimSpace(r.URL.Query().Get("token"))
	}
	courier, err := dbpkg.ResolveDomicilioCourierByToken(dbEmp, empresaID, token)
	if err != nil {
		http.Error(w, "No se pudo validar el domiciliario", http.StatusUnauthorized)
		return dbpkg.EmpresaDomicilioCourier{}, false
	}
	return courier, true
}

func resolveDomicilioRestaurantFromRequest(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, empresaID int64) (dbpkg.EmpresaDomicilioRestaurant, bool) {
	token := strings.TrimSpace(r.Header.Get("X-Domicilio-Restaurant-Token"))
	if token == "" {
		token = strings.TrimSpace(r.URL.Query().Get("token"))
	}
	rest, err := dbpkg.ResolveDomicilioRestaurantByToken(dbEmp, empresaID, token)
	if err != nil {
		http.Error(w, "No se pudo validar el restaurante", http.StatusUnauthorized)
		return dbpkg.EmpresaDomicilioRestaurant{}, false
	}
	return rest, true
}
