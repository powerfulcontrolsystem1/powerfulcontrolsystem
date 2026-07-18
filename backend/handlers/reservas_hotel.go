package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type reservasHotelCacheEntry struct {
	Status   int
	Body     []byte
	Headers  map[string]string
	CachedAt time.Time
}

var (
	reservasHotelReadCacheMu sync.Mutex
	reservasHotelReadCache   = map[string]reservasHotelCacheEntry{}
)

const reservasHotelReadCacheTTL = 10 * time.Second

// EmpresaReservasHotelHandler gestiona el modulo de reservas por empresa.
func EmpresaReservasHotelHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := dbpkg.EmpresaCarritosSchemaReady(dbEmp); err != nil {
			http.Error(w, "El modulo de reservas aun no esta listo. Contacta al administrador para completar la migracion.", http.StatusServiceUnavailable)
			return
		}
		if err := dbpkg.EnsureEmpresaReservasHotelSchema(dbEmp); err != nil {
			http.Error(w, "No se pudo preparar el modulo de reservas", http.StatusInternalServerError)
			return
		}
		switch r.Method {
		case http.MethodGet:
			handleReservasHotelGet(w, r, dbEmp)
			return
		case http.MethodPost:
			handleReservasHotelCreate(w, r, dbEmp)
			return
		case http.MethodPut:
			handleReservasHotelUpdate(w, r, dbEmp)
			return
		case http.MethodDelete:
			handleReservasHotelDelete(w, r, dbEmp)
			return
		default:
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
	}
}

func handleReservasHotelGet(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	cacheKey := buildReservasHotelCacheKey(r, empresaID, action)
	if cached := getReservasHotelCachedResponse(cacheKey); cached != nil {
		for k, v := range cached.Headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(cached.Status)
		_, _ = w.Write(cached.Body)
		return
	}
	switch action {
	case "aplicar_politicas", "sincronizar_estado", "sincronizar_politicas":
		expiradas, noShow, err := dbpkg.ApplyReservasHotelOperationalPolicies(dbEmp, empresaID)
		if err != nil {
			http.Error(w, "No se pudo aplicar la politica operativa de reservas", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":        true,
			"expiradas": expiradas,
			"no_show":   noShow,
		})
		return

	case "", "listar", "list":
		if _, _, err := dbpkg.ApplyReservasHotelOperationalPolicies(dbEmp, empresaID); err != nil {
			http.Error(w, "No se pudo aplicar la politica operativa de reservas", http.StatusInternalServerError)
			return
		}
		limit, err := parseIntQueryOptional(r, "limit")
		if err != nil {
			http.Error(w, "limit invalido", http.StatusBadRequest)
			return
		}
		offset, err := parseIntQueryOptional(r, "offset")
		if err != nil {
			http.Error(w, "offset invalido", http.StatusBadRequest)
			return
		}
		if offset < 0 {
			http.Error(w, "offset invalido", http.StatusBadRequest)
			return
		}

		estacionID, err := parseInt64QueryOptional(r, "estacion_id")
		if err != nil {
			http.Error(w, "estacion_id invalido", http.StatusBadRequest)
			return
		}
		if estacionID < 0 {
			http.Error(w, "estacion_id invalido", http.StatusBadRequest)
			return
		}

		filter := dbpkg.ReservaHotelFilter{
			EstacionID:    estacionID,
			EstadoReserva: strings.TrimSpace(r.URL.Query().Get("estado_reserva")),
			EstadoPago:    strings.TrimSpace(r.URL.Query().Get("estado_pago")),
			Search:        strings.TrimSpace(firstNonEmptyStr(r.URL.Query().Get("search"), r.URL.Query().Get("q"))),
			FechaDesde:    strings.TrimSpace(firstNonEmptyStr(r.URL.Query().Get("fecha_desde"), r.URL.Query().Get("desde"))),
			FechaHasta:    strings.TrimSpace(firstNonEmptyStr(r.URL.Query().Get("fecha_hasta"), r.URL.Query().Get("hasta"))),
			Limit:         limit,
			Offset:        offset,
		}

		total, err := dbpkg.CountReservasHotelByEmpresaRaw(dbEmp, empresaID, filter)
		if err != nil {
			http.Error(w, "No se pudo consultar total de reservas", http.StatusInternalServerError)
			return
		}
		rows, err := dbpkg.ListReservasHotelByEmpresaRaw(dbEmp, empresaID, filter)
		if err != nil {
			http.Error(w, "No se pudo listar reservas", http.StatusInternalServerError)
			return
		}

		if filter.Limit <= 0 {
			filter.Limit = 80
		}
		writeReservasHotelCachedJSON(w, cacheKey, http.StatusOK, rows, map[string]string{
			"X-Total-Count": strconv.FormatInt(total, 10),
			"X-Page-Limit":  strconv.Itoa(filter.Limit),
			"X-Page-Offset": strconv.Itoa(filter.Offset),
		})
		return

	case "detalle", "get", "by_id", "by_codigo":
		id, err := parseInt64QueryOptional(r, "id")
		if err != nil {
			http.Error(w, "id invalido", http.StatusBadRequest)
			return
		}
		codigoReserva := strings.TrimSpace(r.URL.Query().Get("codigo_reserva"))

		var item *dbpkg.ReservaHotel
		if id > 0 {
			item, err = dbpkg.GetReservaHotelByID(dbEmp, empresaID, id)
		} else if codigoReserva != "" {
			item, err = dbpkg.GetReservaHotelByCodigo(dbEmp, empresaID, codigoReserva)
		} else {
			http.Error(w, "id o codigo_reserva es obligatorio", http.StatusBadRequest)
			return
		}
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "reserva no encontrada", http.StatusNotFound)
				return
			}
			http.Error(w, "No se pudo consultar la reserva", http.StatusInternalServerError)
			return
		}
		writeReservasHotelCachedJSON(w, cacheKey, http.StatusOK, item, nil)
		return

	case "disponibilidad", "estaciones_disponibles":
		fechaEntrada := strings.TrimSpace(firstNonEmptyStr(r.URL.Query().Get("fecha_entrada"), r.URL.Query().Get("desde")))
		fechaSalida := strings.TrimSpace(firstNonEmptyStr(r.URL.Query().Get("fecha_salida"), r.URL.Query().Get("hasta")))
		if fechaEntrada == "" || fechaSalida == "" {
			http.Error(w, "fecha_entrada y fecha_salida son obligatorias", http.StatusBadRequest)
			return
		}

		rows, err := dbpkg.ListReservasHotelEstacionesDisponibles(dbEmp, empresaID, fechaEntrada, fechaSalida)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeReservasHotelCachedJSON(w, cacheKey, http.StatusOK, rows, nil)
		return
	default:
		http.Error(w, "action invalida. Use: listar, detalle, disponibilidad o aplicar_politicas", http.StatusBadRequest)
		return
	}
}

func handleReservasHotelCreate(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	var payload dbpkg.ReservaHotel
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
	payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
	if strings.TrimSpace(payload.CanalOrigen) == "" {
		payload.CanalOrigen = "panel_empresa"
	}
	if strings.TrimSpace(payload.RequestID) == "" {
		payload.RequestID = strings.TrimSpace(r.Header.Get("X-Request-ID"))
	}

	id, err := dbpkg.CreateReservaHotel(dbEmp, payload)
	if err != nil {
		writeReservaHotelError(w, err)
		return
	}
	invalidateReservasHotelCache(payload.EmpresaID)
	item, err := dbpkg.GetReservaHotelByID(dbEmp, payload.EmpresaID, id)
	if err != nil {
		writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func handleReservasHotelUpdate(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
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
		nextEstado := "activo"
		if action == "desactivar" {
			nextEstado = "inactivo"
		}
		if err := dbpkg.SetReservaHotelEstado(dbEmp, empresaID, id, nextEstado); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "reserva no encontrada", http.StatusNotFound)
				return
			}
			http.Error(w, "No se pudo actualizar el estado", http.StatusInternalServerError)
			return
		}
		invalidateReservasHotelCache(empresaID)
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": nextEstado})
		return
	}

	var payload dbpkg.ReservaHotel
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID <= 0 {
		if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
			payload.EmpresaID = empresaID
		}
	}
	if payload.ID <= 0 {
		if id, err := parseInt64QueryOptional(r, "id"); err == nil && id > 0 {
			payload.ID = id
		}
	}
	if payload.EmpresaID <= 0 || payload.ID <= 0 {
		http.Error(w, "empresa_id e id son obligatorios", http.StatusBadRequest)
		return
	}

	switch action {
	case "confirmar_pago", "confirmar":
		confirmadoPor := strings.TrimSpace(firstNonEmptyStr(payload.ConfirmadoPor, adminEmailFromRequest(r)))
		err := dbpkg.ConfirmReservaHotelPago(dbEmp, payload.EmpresaID, payload.ID, payload.ReferenciaPago, confirmadoPor, payload.Observaciones)
		if err != nil {
			writeReservaHotelError(w, err)
			return
		}
		invalidateReservasHotelCache(payload.EmpresaID)
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "accion": "confirmar_pago"})
		return

	case "cancelar", "cancelar_reserva":
		usuario := strings.TrimSpace(firstNonEmptyStr(payload.ConfirmadoPor, adminEmailFromRequest(r)))
		err := dbpkg.CancelReservaHotel(dbEmp, payload.EmpresaID, payload.ID, payload.Observaciones, usuario)
		if err != nil {
			writeReservaHotelError(w, err)
			return
		}
		invalidateReservasHotelCache(payload.EmpresaID)
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "accion": "cancelar"})
		return

	case "convertir_carrito", "reconvertir_carrito":
		usuario := strings.TrimSpace(firstNonEmptyStr(payload.ConfirmadoPor, adminEmailFromRequest(r)))
		carritoID, err := dbpkg.ConvertReservaHotelToCarrito(dbEmp, payload.EmpresaID, payload.ID, usuario)
		if err != nil {
			writeReservaHotelError(w, err)
			return
		}
		invalidateReservasHotelCache(payload.EmpresaID)
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "accion": "convertir_carrito", "reserva_id": payload.ID, "carrito_id": carritoID})
		return

	default:
		if err := dbpkg.UpdateReservaHotel(dbEmp, payload); err != nil {
			writeReservaHotelError(w, err)
			return
		}
		invalidateReservasHotelCache(payload.EmpresaID)
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
		return
	}
}

func handleReservasHotelDelete(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
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
	if err := dbpkg.DeleteReservaHotel(dbEmp, empresaID, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "reserva no encontrada", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo eliminar la reserva", http.StatusInternalServerError)
		return
	}
	invalidateReservasHotelCache(empresaID)
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

func buildReservasHotelCacheKey(r *http.Request, empresaID int64, action string) string {
	if r == nil || r.URL == nil || empresaID <= 0 || !strings.EqualFold(strings.TrimSpace(r.Method), http.MethodGet) {
		return ""
	}
	return strings.Join([]string{
		strconv.FormatInt(empresaID, 10),
		strings.ToLower(strings.TrimSpace(action)),
		r.URL.RawQuery,
	}, "|")
}

func getReservasHotelCachedResponse(key string) *reservasHotelCacheEntry {
	if strings.TrimSpace(key) == "" {
		return nil
	}
	reservasHotelReadCacheMu.Lock()
	defer reservasHotelReadCacheMu.Unlock()
	entry, ok := reservasHotelReadCache[key]
	if !ok {
		return nil
	}
	if time.Since(entry.CachedAt) > reservasHotelReadCacheTTL {
		delete(reservasHotelReadCache, key)
		return nil
	}
	copied := reservasHotelCacheEntry{
		Status:   entry.Status,
		Body:     append([]byte(nil), entry.Body...),
		Headers:  map[string]string{},
		CachedAt: entry.CachedAt,
	}
	for k, v := range entry.Headers {
		copied.Headers[k] = v
	}
	return &copied
}

func writeReservasHotelCachedJSON(w http.ResponseWriter, key string, status int, payload interface{}, headers map[string]string) {
	body, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "No se pudo serializar la respuesta", http.StatusInternalServerError)
		return
	}
	if headers == nil {
		headers = map[string]string{}
	}
	headers["Content-Type"] = "application/json; charset=utf-8"
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(status)
	_, _ = w.Write(body)

	if strings.TrimSpace(key) == "" {
		return
	}
	cachedHeaders := map[string]string{}
	for k, v := range headers {
		cachedHeaders[k] = v
	}
	reservasHotelReadCacheMu.Lock()
	reservasHotelReadCache[key] = reservasHotelCacheEntry{
		Status:   status,
		Body:     append([]byte(nil), body...),
		Headers:  cachedHeaders,
		CachedAt: time.Now(),
	}
	reservasHotelReadCacheMu.Unlock()
}

func invalidateReservasHotelCache(empresaID int64) {
	if empresaID <= 0 {
		return
	}
	prefix := strconv.FormatInt(empresaID, 10) + "|"
	reservasHotelReadCacheMu.Lock()
	for key := range reservasHotelReadCache {
		if strings.HasPrefix(key, prefix) {
			delete(reservasHotelReadCache, key)
		}
	}
	reservasHotelReadCacheMu.Unlock()
}

func writeReservaHotelError(w http.ResponseWriter, err error) {
	if err == nil {
		http.Error(w, "error no especificado", http.StatusInternalServerError)
		return
	}
	switch {
	case errors.Is(err, dbpkg.ErrReservaHotelConflicto):
		http.Error(w, "conflicto de reserva en el rango de fechas solicitado", http.StatusConflict)
	case errors.Is(err, dbpkg.ErrReservaHotelExpirada):
		http.Error(w, "la reserva se encuentra expirada", http.StatusConflict)
	case errors.Is(err, dbpkg.ErrReservaHotelNoReconvertible):
		http.Error(w, "la reserva no esta en un estado valido para reconversion a carrito", http.StatusConflict)
	case errors.Is(err, sql.ErrNoRows):
		http.Error(w, "reserva no encontrada", http.StatusNotFound)
	default:
		msg := strings.TrimSpace(err.Error())
		if msg == "" {
			msg = "operacion invalida"
		}
		http.Error(w, msg, http.StatusBadRequest)
	}
}

func firstNonEmptyStr(values ...string) string {
	for _, v := range values {
		trim := strings.TrimSpace(v)
		if trim != "" {
			return trim
		}
	}
	return ""
}
