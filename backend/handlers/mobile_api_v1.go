package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
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
		if json.Unmarshal(capture.body.Bytes(), &envelope) == nil && envelope.RequestID != "" {
			for k, values := range capture.header {
				w.Header()[k] = append([]string{}, values...)
			}
			w.WriteHeader(status)
			_, _ = w.Write(capture.body.Bytes())
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
	if raw := strings.TrimSpace(r.URL.Query().Get("offset")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			return 0, 0, fmt.Errorf("offset invalido")
		}
		offset = n
	}
	return limit, offset, nil
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
		writeJSON(w, http.StatusOK, mobileAPIEnvelope{OK: true, Data: mobileSelectFields(items, r.URL.Query().Get("fields"), allowed), Meta: map[string]interface{}{"limit": limit, "offset": offset, "next_offset": next, "returned": len(items)}, RequestID: mobileAPIRequestID()})
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
		writeJSON(w, http.StatusOK, mobileAPIEnvelope{OK: true, Data: mobileSelectFields(items, r.URL.Query().Get("fields"), allowed), Meta: map[string]interface{}{"limit": limit, "offset": offset, "next_offset": next, "returned": len(items)}, RequestID: mobileAPIRequestID()})
	}
}

func RegisterMobileAPIV1Routes(dbEmp, dbSuper *sql.DB) {
	http.HandleFunc("/api/v1/auth/mobile-session", mobileAPIJSON(MobileSessionTokenHandler(dbSuper)))
	http.HandleFunc("/api/v1/me", mobileAPIJSON(mobileBearerSessionAdapter(dbSuper, MobileMeHandler(dbSuper))))
	http.HandleFunc("/api/v1/empresa/productos", mobileAPIJSON(mobileBearerSessionAdapter(dbSuper, WithEmpresaInventarioPermissions(dbEmp, dbSuper, MobileProductosHandler(dbEmp)))))
	http.HandleFunc("/api/v1/empresa/clientes", mobileAPIJSON(mobileBearerSessionAdapter(dbSuper, WithEmpresaClientesPermissions(dbEmp, dbSuper, MobileClientesHandler(dbEmp)))))
}
