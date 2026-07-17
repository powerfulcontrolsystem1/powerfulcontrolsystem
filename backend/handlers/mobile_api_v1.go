package handlers

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

// Mobile API v1 is an additive facade. Legacy browser endpoints remain
// compatible while mobile clients get a stable JSON envelope and pagination.
type mobileAPIEnvelope struct {
	OK        bool         `json:"ok"`
	Data      interface{}  `json:"data,omitempty"`
	Meta      interface{}  `json:"meta,omitempty"`
	Error     *mobileError `json:"error,omitempty"`
	RequestID string       `json:"request_id"`
}

type mobileError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type mobileResponseCapture struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (c *mobileResponseCapture) Header() http.Header { return c.header }
func (c *mobileResponseCapture) WriteHeader(status int) {
	if c.status == 0 {
		c.status = status
	}
}
func (c *mobileResponseCapture) Write(p []byte) (int, error) {
	if c.status == 0 {
		c.status = http.StatusOK
	}
	return c.body.Write(p)
}

func mobileAPIRequestID() string {
	token, err := utils.GenerateSecureToken(12)
	if err != nil {
		return "mobile-request"
	}
	return "m_" + token
}

func mobileAPIErrorCode(status int) string {
	switch status {
	case http.StatusUnauthorized:
		return "unauthenticated"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusTooManyRequests:
		return "rate_limited"
	case http.StatusBadRequest:
		return "invalid_request"
	default:
		return "internal_error"
	}
}

func writeMobileAPIJSON(w http.ResponseWriter, status int, payload mobileAPIEnvelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func mobileAPIJSON(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		capture := &mobileResponseCapture{header: make(http.Header)}
		next(capture, r)
		status := capture.status
		if status == 0 {
			status = http.StatusOK
		}
		requestID := mobileAPIRequestID()
		w.Header().Set("X-Request-ID", requestID)
		if status >= 400 {
			writeMobileAPIJSON(w, status, mobileAPIEnvelope{OK: false, Error: &mobileError{Code: mobileAPIErrorCode(status), Message: "No fue posible completar la solicitud."}, RequestID: requestID})
			return
		}
		if capture.body.Len() == 0 {
			writeMobileAPIJSON(w, status, mobileAPIEnvelope{OK: true, RequestID: requestID})
			return
		}
		var envelope mobileAPIEnvelope
		// Some v1 handlers build the envelope themselves while others reuse a
		// legacy handler. Preserve the former even when the handler intentionally
		// leaves request_id to this boundary; otherwise clients would receive a
		// nested {data:{ok:...}} response depending on the endpoint used.
		if json.Unmarshal(capture.body.Bytes(), &envelope) == nil && (envelope.OK || envelope.Error != nil || envelope.Data != nil || envelope.Meta != nil) {
			if envelope.RequestID == "" {
				envelope.RequestID = requestID
			}
			for k, values := range capture.header {
				w.Header()[k] = append([]string{}, values...)
			}
			writeMobileAPIJSON(w, status, envelope)
			return
		}
		var legacy interface{}
		if json.Unmarshal(capture.body.Bytes(), &legacy) != nil {
			legacy = map[string]interface{}{}
		}
		writeMobileAPIJSON(w, status, mobileAPIEnvelope{OK: true, Data: legacy, RequestID: requestID})
	}
}

func mobileBearerSessionAdapter(dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := r.Cookie("session_token"); err == nil {
			next(w, r)
			return
		}
		auth := strings.TrimSpace(r.Header.Get("Authorization"))
		if !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			next(w, r)
			return
		}
		token := strings.TrimSpace(auth[len("Bearer "):])
		if len(token) < 32 || len(token) > 512 {
			next(w, r)
			return
		}
		if session, err := dbpkg.GetSessionByToken(dbSuper, token); err != nil || session == nil {
			next(w, r)
			return
		}
		r2 := r.Clone(r.Context())
		r2.AddCookie(&http.Cookie{Name: "session_token", Value: token, Path: "/", HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode})
		next(w, r2)
	}
}

func mobileCurrentSession(r *http.Request, dbSuper *sql.DB) (*dbpkg.Session, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie == nil || strings.TrimSpace(cookie.Value) == "" {
		return nil, fmt.Errorf("unauthenticated")
	}
	return dbpkg.GetSessionByToken(dbSuper, cookie.Value)
}

// MobileSessionTokenHandler creates a distinct short-lived server session for a
// device after a browser/OAuth login. Only its raw value is returned once; the
// database stores the established SHA-256 verifier through CreateSession.
func MobileSessionTokenHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		current, err := mobileCurrentSession(r, dbSuper)
		if err != nil || current == nil {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		var body struct {
			DeviceName string `json:"device_name"`
		}
		_ = json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&body)
		device := strings.TrimSpace(body.DeviceName)
		if len(device) > 80 {
			http.Error(w, "invalid device name", http.StatusBadRequest)
			return
		}
		token, err := utils.GenerateSecureToken(32)
		if err != nil {
			http.Error(w, "token unavailable", http.StatusInternalServerError)
			return
		}
		if err := dbpkg.CreateSession(dbSuper, current.AdminEmail, r.RemoteAddr, "mobile-api/"+device, token); err != nil {
			http.Error(w, "session unavailable", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, mobileAPIEnvelope{OK: true, Data: map[string]interface{}{"access_token": token, "token_type": "Bearer", "scope": "authenticated_account"}, RequestID: mobileAPIRequestID()})
	}
}

func MobileMeHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		session, err := mobileCurrentSession(r, dbSuper)
		if err != nil || session == nil {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		writeJSON(w, http.StatusOK, mobileAPIEnvelope{OK: true, Data: map[string]interface{}{"email": session.AdminEmail}, RequestID: mobileAPIRequestID()})
	}
}

func mobileLimitOffset(r *http.Request) (int, int, error) {
	limit, offset := 50, 0
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > 100 {
			return 0, 0, fmt.Errorf("limit invalido")
		}
		limit = n
	}
	rawOffset := strings.TrimSpace(r.URL.Query().Get("offset"))
	rawCursor := strings.TrimSpace(r.URL.Query().Get("cursor"))
	if rawOffset != "" && rawCursor != "" {
		return 0, 0, fmt.Errorf("offset y cursor no se pueden combinar")
	}
	if rawOffset != "" {
		n, err := strconv.Atoi(rawOffset)
		if err != nil || n < 0 {
			return 0, 0, fmt.Errorf("offset invalido")
		}
		offset = n
	}
	if rawCursor != "" {
		decoded, err := base64.RawURLEncoding.DecodeString(rawCursor)
		if err != nil || len(decoded) > 32 || !strings.HasPrefix(string(decoded), "v1:") {
			return 0, 0, fmt.Errorf("cursor invalido")
		}
		n, err := strconv.Atoi(strings.TrimPrefix(string(decoded), "v1:"))
		if err != nil || n < 0 {
			return 0, 0, fmt.Errorf("cursor invalido")
		}
		offset = n
	}
	return limit, offset, nil
}

func mobilePageMeta(limit, offset, nextOffset, returned int) map[string]interface{} {
	nextCursor := ""
	if nextOffset >= 0 {
		nextCursor = base64.RawURLEncoding.EncodeToString([]byte("v1:" + strconv.Itoa(nextOffset)))
	}
	return map[string]interface{}{
		"limit": limit, "offset": offset, "next_offset": nextOffset,
		"next_cursor": nextCursor, "returned": returned,
	}
}

func mobileSelectFields(items interface{}, raw string, allowed map[string]bool) interface{} {
	requested := strings.Split(strings.TrimSpace(raw), ",")
	if len(requested) == 0 || strings.TrimSpace(raw) == "" {
		return items
	}
	fields := make(map[string]bool, len(requested))
	for _, field := range requested {
		field = strings.TrimSpace(field)
		if allowed[field] {
			fields[field] = true
		}
	}
	if len(fields) == 0 {
		return items
	}
	encoded, err := json.Marshal(items)
	if err != nil {
		return items
	}
	var maps []map[string]interface{}
	if json.Unmarshal(encoded, &maps) != nil {
		return items
	}
	for _, item := range maps {
		for key := range item {
			if !fields[key] {
				delete(item, key)
			}
		}
	}
	return maps
}

func mobileLegacyAction(next http.HandlerFunc, action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clone := r.Clone(r.Context())
		query := clone.URL.Query()
		if strings.TrimSpace(action) != "" {
			query.Set("action", action)
		}
		clone.URL = cloneURLWithQuery(clone.URL, query)
		next(w, clone)
	}
}

func cloneURLWithQuery(source *url.URL, query url.Values) *url.URL {
	copyURL := *source
	copyURL.RawQuery = query.Encode()
	return &copyURL
}

// mobileNormalizeEmpresaJSON makes the tenant selected by the validated route
// authoritative. A mobile client may send empresa_id for convenience, but it
// cannot select a different tenant through a JSON body.
func mobileNormalizeEmpresaJSON(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodDelete || r.Body == nil {
			next(w, r)
			return
		}
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil || empresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}
		body := []byte{}
		if r.Body != nil {
			body, err = io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
			if err != nil {
				http.Error(w, "payload invalido", http.StatusBadRequest)
				return
			}
		}
		if len(bytes.TrimSpace(body)) == 0 {
			r.Body = io.NopCloser(bytes.NewReader(body))
			next(w, r)
			return
		}
		var payload map[string]interface{}
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "payload JSON invalido", http.StatusBadRequest)
			return
		}
		payload["empresa_id"] = empresaID
		normalized, err := json.Marshal(payload)
		if err != nil {
			http.Error(w, "payload invalido", http.StatusBadRequest)
			return
		}
		clone := r.Clone(r.Context())
		clone.Body = io.NopCloser(bytes.NewReader(normalized))
		clone.ContentLength = int64(len(normalized))
		next(w, clone)
	}
}

func validMobileIdempotencyKey(key string) bool {
	key = strings.TrimSpace(key)
	if len(key) < 16 || len(key) > 200 {
		return false
	}
	for _, char := range key {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_' || char == '.' {
			continue
		}
		return false
	}
	return true
}

// mobileIdempotentMutation persists successful financial/document mutations.
// The key is hashed, scoped to the tenant and operation, and never logged.
func mobileIdempotentMutation(dbEmp *sql.DB, operation string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
		if !validMobileIdempotencyKey(key) {
			http.Error(w, "Idempotency-Key valido es obligatorio", http.StatusBadRequest)
			return
		}
		empresaID := parseEmpresaIDFromContext(r)
		queryEmpresaID, err := parseEmpresaIDQuery(r)
		if err != nil || empresaID <= 0 || (queryEmpresaID > 0 && queryEmpresaID != empresaID) {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}
		body := []byte{}
		if r.Body != nil {
			body, err = io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
			if err != nil {
				http.Error(w, "payload invalido", http.StatusBadRequest)
				return
			}
		}
		claim, claimed, err := dbpkg.ClaimMobileAPIIdempotency(dbEmp, empresaID, operation, key, string(body))
		if err != nil {
			if errors.Is(err, dbpkg.ErrMobileAPIIdempotencyConflict) {
				http.Error(w, "Idempotency-Key ya fue usado con otra solicitud", http.StatusConflict)
				return
			}
			http.Error(w, "no se pudo preparar la operacion", http.StatusServiceUnavailable)
			return
		}
		if !claimed {
			if claim.Status == "completado" && claim.ResponseCode >= 200 && claim.ResponseCode < 300 {
				w.Header().Set("Idempotency-Replayed", "true")
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(claim.ResponseCode)
				_, _ = io.WriteString(w, claim.ResponseJSON)
				return
			}
			w.Header().Set("Retry-After", "2")
			http.Error(w, "operacion en proceso; reintenta", http.StatusConflict)
			return
		}

		capture := &mobileResponseCapture{header: make(http.Header)}
		clone := r.Clone(r.Context())
		clone.Body = io.NopCloser(bytes.NewReader(body))
		mobileNormalizeEmpresaJSON(next).ServeHTTP(capture, clone)
		status := capture.status
		if status == 0 {
			status = http.StatusOK
		}
		if status >= 200 && status < 300 {
			if err := dbpkg.CompleteMobileAPIIdempotency(dbEmp, claim, status, capture.body.String()); err != nil {
				http.Error(w, "no se pudo cerrar la operacion", http.StatusServiceUnavailable)
				return
			}
		} else {
			_ = dbpkg.AbandonMobileAPIIdempotency(dbEmp, claim)
		}
		for header, values := range capture.header {
			w.Header()[header] = append([]string{}, values...)
		}
		w.WriteHeader(status)
		_, _ = w.Write(capture.body.Bytes())
	}
}

// mobileIdempotentWhenMutating keeps reads cacheable and simple while making
// every state change retry-safe. It is deliberately applied after the tenant
// permission middleware so the idempotency row is always scoped to the
// authorized empresa in the request context.
func mobileIdempotentWhenMutating(dbEmp *sql.DB, operation string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			mobileIdempotentMutation(dbEmp, operation, next).ServeHTTP(w, r)
		default:
			next.ServeHTTP(w, r)
		}
	}
}

func MobileProductosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}
		limit, offset, err := mobileLimitOffset(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		bodegaID, _ := parseInt64QueryOptional(r, "bodega_id")
		categoriaID, _ := parseInt64QueryOptional(r, "categoria_id")
		items, err := dbpkg.GetProductosByEmpresa(dbEmp, empresaID, r.URL.Query().Get("q"), r.URL.Query().Get("estado"), bodegaID, categoriaID, limit, offset)
		if err != nil {
			http.Error(w, "catalog unavailable", http.StatusServiceUnavailable)
			return
		}
		next := -1
		if len(items) == limit {
			next = offset + len(items)
		}
		allowed := map[string]bool{"id": true, "nombre": true, "codigo_barras": true, "codigo_sku": true, "sku": true, "precio_venta": true, "stock_actual": true, "estado": true, "categoria_id": true, "impuesto_porcentaje": true, "tipo": true}
		writeJSON(w, http.StatusOK, mobileAPIEnvelope{OK: true, Data: mobileSelectFields(items, r.URL.Query().Get("fields"), allowed), Meta: mobilePageMeta(limit, offset, next, len(items)), RequestID: mobileAPIRequestID()})
	}
}

func MobileClientesHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}
		limit, offset, err := mobileLimitOffset(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		includeInactive := r.URL.Query().Get("include_inactive") == "1" || strings.EqualFold(r.URL.Query().Get("include_inactive"), "true")
		all, err := dbpkg.GetClientesByEmpresa(dbEmp, empresaID, includeInactive, r.URL.Query().Get("q"))
		if err != nil {
			http.Error(w, "customers unavailable", http.StatusServiceUnavailable)
			return
		}
		if offset > len(all) {
			offset = len(all)
		}
		end := offset + limit
		if end > len(all) {
			end = len(all)
		}
		items := all[offset:end]
		next := -1
		if end < len(all) {
			next = end
		}
		allowed := map[string]bool{"id": true, "nombre": true, "numero_documento": true, "tipo_documento": true, "telefono": true, "email": true, "estado": true, "direccion": true, "ciudad": true}
		writeJSON(w, http.StatusOK, mobileAPIEnvelope{OK: true, Data: mobileSelectFields(items, r.URL.Query().Get("fields"), allowed), Meta: mobilePageMeta(limit, offset, next, len(items)), RequestID: mobileAPIRequestID()})
	}
}

func MobileCarritosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}
		if id, err := parseOptionalInt64CarritoQuery(r, "id"); err != nil {
			http.Error(w, "id invalido", http.StatusBadRequest)
			return
		} else if id > 0 {
			if isStationBoardOnlyCarritoRequest(r) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			item, err := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, id)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "carrito no encontrado", http.StatusNotFound)
					return
				}
				http.Error(w, "carrito no disponible", http.StatusServiceUnavailable)
				return
			}
			if err := ensureCarritoStationAccessForCarrito(dbEmp, empresaID, adminEmailFromRequest(r), item); err != nil {
				writeCarritoStationAccessError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, mobileAPIEnvelope{OK: true, Data: item, RequestID: mobileAPIRequestID()})
			return
		}
		limit, offset, err := mobileLimitOffset(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		includeInactive := r.URL.Query().Get("include_inactive") == "1" || strings.EqualFold(r.URL.Query().Get("include_inactive"), "true")
		items, err := dbpkg.GetCarritosCompraByEmpresa(dbEmp, empresaID, includeInactive, strings.TrimSpace(r.URL.Query().Get("q")))
		if err != nil {
			http.Error(w, "carritos no disponibles", http.StatusServiceUnavailable)
			return
		}
		items, err = filterCarritosByStationAccess(dbEmp, empresaID, adminEmailFromRequest(r), items)
		if err != nil {
			writeCarritoStationAccessError(w, err)
			return
		}
		if isStationBoardOnlyCarritoRequest(r) {
			stationOnly := make([]dbpkg.CarritoCompra, 0, len(items))
			for _, item := range items {
				stationID, _, _ := dbpkg.ResolveCarritoStationIdentity(&item)
				if stationID > 0 {
					stationOnly = append(stationOnly, item)
				}
			}
			items = stationOnly
		}
		statusFilter := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("estado")))
		if statusFilter != "" && statusFilter != "todos" {
			filtered := make([]dbpkg.CarritoCompra, 0, len(items))
			for _, item := range items {
				paid := strings.TrimSpace(item.PagadoEn) != "" || strings.EqualFold(item.EstadoCarrito, "cerrado") || strings.EqualFold(item.EstadoCarrito, "pagado")
				if (statusFilter == "pagado" && paid) || (statusFilter == "abierto" && !paid) || strings.EqualFold(statusFilter, strings.TrimSpace(item.EstadoCarrito)) {
					filtered = append(filtered, item)
				}
			}
			items = filtered
		}
		if offset > len(items) {
			offset = len(items)
		}
		end := offset + limit
		if end > len(items) {
			end = len(items)
		}
		next := -1
		if end < len(items) {
			next = end
		}
		allowed := map[string]bool{"id": true, "codigo": true, "nombre": true, "cliente_id": true, "cliente_nombre": true, "estado_carrito": true, "estado_venta": true, "moneda": true, "subtotal": true, "descuento_total": true, "impuesto_total": true, "total": true, "metodo_pago": true, "total_pagado": true, "pagado_en": true, "fecha_actualizacion": true, "item_count": true, "canal_venta": true}
		writeJSON(w, http.StatusOK, mobileAPIEnvelope{OK: true, Data: mobileSelectFields(items[offset:end], r.URL.Query().Get("fields"), allowed), Meta: mobilePageMeta(limit, offset, next, end-offset), RequestID: mobileAPIRequestID()})
	}
}

// MobileVentasHandler is the historical POS view. A sale is a paid/closed
// carrito, so it deliberately uses the same source of truth as the web POS.
func MobileVentasHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		clone := r.Clone(r.Context())
		query := clone.URL.Query()
		if strings.TrimSpace(query.Get("estado")) == "" {
			query.Set("estado", "pagado")
		}
		clone.URL = cloneURLWithQuery(clone.URL, query)
		MobileCarritosHandler(dbEmp).ServeHTTP(w, clone)
	}
}

func MobileCarritoItemsHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if isStationBoardOnlyCarritoRequest(r) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}
		carritoID, err := parseInt64Query(r, "carrito_id")
		if err != nil || carritoID <= 0 {
			http.Error(w, "carrito_id es obligatorio", http.StatusBadRequest)
			return
		}
		if err := ensureCarritoStationAccessByID(dbEmp, empresaID, carritoID, adminEmailFromRequest(r)); err != nil {
			writeCarritoStationAccessError(w, err)
			return
		}
		limit, offset, err := mobileLimitOffset(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		includeInactive := r.URL.Query().Get("include_inactive") == "1" || strings.EqualFold(r.URL.Query().Get("include_inactive"), "true")
		all, err := dbpkg.GetCarritoCompraItems(dbEmp, empresaID, carritoID, includeInactive)
		if err != nil {
			http.Error(w, "items no disponibles", http.StatusServiceUnavailable)
			return
		}
		if offset > len(all) {
			offset = len(all)
		}
		end := offset + limit
		if end > len(all) {
			end = len(all)
		}
		next := -1
		if end < len(all) {
			next = end
		}
		allowed := map[string]bool{"id": true, "carrito_id": true, "tipo_item": true, "referencia_id": true, "codigo_item": true, "descripcion": true, "unidad_medida": true, "cantidad": true, "precio_unitario": true, "descuento_porcentaje": true, "impuesto_porcentaje": true, "total_linea": true, "estado": true, "fecha_actualizacion": true}
		writeJSON(w, http.StatusOK, mobileAPIEnvelope{OK: true, Data: mobileSelectFields(all[offset:end], r.URL.Query().Get("fields"), allowed), Meta: mobilePageMeta(limit, offset, next, end-offset), RequestID: mobileAPIRequestID()})
	}
}

func MobileDocumentosFacturacionHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}
		limit, offset, err := mobileLimitOffset(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		cajero := strings.TrimSpace(r.URL.Query().Get("cajero"))
		if normalizePermissionRole(effectiveAdminRoleFromRequest(r)) == "cajero" {
			cajero = strings.TrimSpace(adminEmailFromRequest(r))
		}
		items, err := dbpkg.ListEmpresaDocumentosFacturacionByEmpresa(dbEmp, dbpkg.EmpresaDocumentoFacturacionListFilter{
			EmpresaID: empresaID, TipoDocumento: strings.TrimSpace(r.URL.Query().Get("tipo_documento")), EstadoDocumento: strings.TrimSpace(r.URL.Query().Get("estado_documento")),
			IncludeInactive: r.URL.Query().Get("include_inactive") == "1" || strings.EqualFold(r.URL.Query().Get("include_inactive"), "true"),
			ClienteQuery:    strings.TrimSpace(r.URL.Query().Get("cliente")), DocumentoQuery: strings.TrimSpace(r.URL.Query().Get("documento")), CajeroQuery: cajero,
			FechaDesde: strings.TrimSpace(r.URL.Query().Get("fecha_desde")), FechaHasta: strings.TrimSpace(r.URL.Query().Get("fecha_hasta")), Query: strings.TrimSpace(r.URL.Query().Get("q")), Limit: limit, Offset: offset,
		})
		if err != nil {
			http.Error(w, "documentos no disponibles", http.StatusServiceUnavailable)
			return
		}
		next := -1
		if len(items) == limit {
			next = offset + len(items)
		}
		allowed := map[string]bool{"id": true, "tipo_documento": true, "documento_codigo": true, "numero_legal": true, "pais_codigo": true, "ambiente_fe": true, "estado_documento": true, "monto_total": true, "moneda": true, "fecha_documento": true, "fecha_creacion": true, "usuario_creador": true, "cliente_nombre": true, "cliente_documento": true}
		writeJSON(w, http.StatusOK, mobileAPIEnvelope{OK: true, Data: mobileSelectFields(items, r.URL.Query().Get("fields"), allowed), Meta: mobilePageMeta(limit, offset, next, len(items)), RequestID: mobileAPIRequestID()})
	}
}

func RegisterMobileAPIV1Routes(dbEmp, dbSuper *sql.DB) {
	http.HandleFunc("/api/v1/auth/mobile-session", mobileAPIJSON(MobileSessionTokenHandler(dbSuper)))
	http.HandleFunc("/api/v1/me", mobileAPIJSON(mobileBearerSessionAdapter(dbSuper, MobileMeHandler(dbSuper))))
	http.HandleFunc("/api/v1/empresa/productos", mobileAPIJSON(mobileBearerSessionAdapter(dbSuper, WithEmpresaInventarioPermissions(dbEmp, dbSuper, MobileProductosHandler(dbEmp)))))
	http.HandleFunc("/api/v1/empresa/clientes", mobileAPIJSON(mobileBearerSessionAdapter(dbSuper, WithEmpresaClientesPermissions(dbEmp, dbSuper, MobileClientesHandler(dbEmp)))))

	carritosLegacy := EmpresaCarritosCompraHandler(dbEmp, dbSuper)
	itemsLegacy := EmpresaCarritoItemsHandler(dbEmp)
	facturacionLegacy := EmpresaFacturacionElectronicaHandler(dbEmp, dbSuper)
	offlineLegacy := EmpresaOfflineVentasHandler(dbEmp, dbSuper)
	buzonLegacy := EmpresaBuzonHandler(dbEmp, dbSuper)

	carritosV1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			MobileCarritosHandler(dbEmp).ServeHTTP(w, r)
		case http.MethodPost:
			mobileIdempotentMutation(dbEmp, "carrito_crear", carritosLegacy).ServeHTTP(w, r)
		case http.MethodPut:
			mobileIdempotentMutation(dbEmp, "carrito_actualizar", carritosLegacy).ServeHTTP(w, r)
		case http.MethodDelete:
			mobileIdempotentMutation(dbEmp, "carrito_eliminar", carritosLegacy).ServeHTTP(w, r)
		default:
			w.Header().Set("Allow", "GET, POST, PUT, DELETE")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/api/v1/empresa/carritos", mobileAPIJSON(mobileBearerSessionAdapter(dbSuper, mobileNormalizeEmpresaJSON(WithEmpresaVentasPermissions(dbEmp, dbSuper, mobileIdempotentWhenMutating(dbEmp, "carritos_mutar", carritosV1))))))

	itemsV1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			MobileCarritoItemsHandler(dbEmp).ServeHTTP(w, r)
		case http.MethodPost:
			mobileIdempotentMutation(dbEmp, "carrito_item_crear", itemsLegacy).ServeHTTP(w, r)
		case http.MethodPut:
			mobileIdempotentMutation(dbEmp, "carrito_item_actualizar", itemsLegacy).ServeHTTP(w, r)
		case http.MethodDelete:
			mobileIdempotentMutation(dbEmp, "carrito_item_eliminar", itemsLegacy).ServeHTTP(w, r)
		default:
			w.Header().Set("Allow", "GET, POST, PUT, DELETE")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/api/v1/empresa/carritos/items", mobileAPIJSON(mobileBearerSessionAdapter(dbSuper, mobileNormalizeEmpresaJSON(WithEmpresaVentasPermissions(dbEmp, dbSuper, mobileIdempotentWhenMutating(dbEmp, "carrito_items_mutar", itemsV1))))))
	http.HandleFunc("/api/v1/empresa/ventas", mobileAPIJSON(mobileBearerSessionAdapter(dbSuper, WithEmpresaVentasPermissions(dbEmp, dbSuper, MobileVentasHandler(dbEmp)))))
	http.HandleFunc("/api/v1/empresa/pagos", mobileAPIJSON(mobileBearerSessionAdapter(dbSuper, mobileNormalizeEmpresaJSON(WithEmpresaVentasPermissions(dbEmp, dbSuper, mobileIdempotentMutation(dbEmp, "venta_cobrar", mobileLegacyAction(carritosLegacy, "pagar_estacion")))))))
	http.HandleFunc("/api/v1/empresa/ventas/offline/sync", mobileAPIJSON(mobileBearerSessionAdapter(dbSuper, mobileNormalizeEmpresaJSON(WithEmpresaVentasPermissions(dbEmp, dbSuper, mobileIdempotentMutation(dbEmp, "ventas_offline_sync", mobileLegacyAction(offlineLegacy, "sync")))))))

	http.HandleFunc("/api/v1/empresa/facturacion/documentos", mobileAPIJSON(mobileBearerSessionAdapter(dbSuper, WithEmpresaFacturacionPermissions(dbEmp, dbSuper, MobileDocumentosFacturacionHandler(dbEmp)))))
	http.HandleFunc("/api/v1/empresa/facturacion/emitir", mobileAPIJSON(mobileBearerSessionAdapter(dbSuper, mobileNormalizeEmpresaJSON(WithEmpresaFacturacionPermissions(dbEmp, dbSuper, mobileIdempotentMutation(dbEmp, "facturacion_emitir_desde_venta", mobileLegacyAction(facturacionLegacy, "facturar_desde_venta")))))))

	notificacionesV1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			mobileLegacyAction(buzonLegacy, "mensajes").ServeHTTP(w, r)
		case http.MethodPost:
			mobileIdempotentMutation(dbEmp, "notificacion_enviar", mobileLegacyAction(buzonLegacy, "mensaje")).ServeHTTP(w, r)
		case http.MethodPut:
			mobileIdempotentMutation(dbEmp, "notificacion_leer", mobileLegacyAction(buzonLegacy, "leer")).ServeHTTP(w, r)
		default:
			w.Header().Set("Allow", "GET, POST, PUT")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/api/v1/empresa/notificaciones", mobileAPIJSON(mobileBearerSessionAdapter(dbSuper, mobileNormalizeEmpresaJSON(WithEmpresaSelfServicePermissions(dbEmp, dbSuper, notificacionesV1)))))
}
