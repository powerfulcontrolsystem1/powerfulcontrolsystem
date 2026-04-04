package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	permActionRead    = "R"
	permActionCreate  = "C"
	permActionUpdate  = "U"
	permActionDelete  = "D"
	permActionApprove = "A"

	permModuleVentas     = "ventas"
	permModuleInventario = "inventario"
	permModuleFinanzas   = "finanzas"
)

// WithEmpresaVentasPermissions aplica control de alcance por empresa y permisos por rol para ventas.
func WithEmpresaVentasPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleVentas, resolveVentasPermissionAction, next)
}

// WithEmpresaInventarioPermissions aplica control de alcance por empresa y permisos por rol para inventario.
func WithEmpresaInventarioPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleInventario, resolveInventarioPermissionAction, next)
}

// WithEmpresaFinanzasPermissions aplica control de alcance por empresa y permisos por rol para finanzas.
func WithEmpresaFinanzasPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleFinanzas, resolveFinanzasPermissionAction, next)
}

func withEmpresaRolePermissions(dbEmp, dbSuper *sql.DB, module string, resolveAction func(*http.Request) string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID := extractEmpresaIDForPermissions(r)
		if empresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}

		adminEmail := strings.ToLower(strings.TrimSpace(adminEmailFromRequest(r)))
		if adminEmail == "" || adminEmail == "sistema" {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}

		admin, err := dbpkg.GetAdminByEmail(dbSuper, adminEmail)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "unauthenticated", http.StatusUnauthorized)
				return
			}
			log.Printf("[authz] get admin email=%s error: %v", adminEmail, err)
			http.Error(w, "No se pudo validar permisos del usuario", http.StatusInternalServerError)
			return
		}

		canAccess, err := dbpkg.CanAdminAccessEmpresaIA(dbEmp, dbSuper, adminEmail, empresaID)
		if err != nil {
			log.Printf("[authz] alcance empresa module=%s email=%s empresa_id=%d error: %v", module, adminEmail, empresaID, err)
			http.Error(w, "No se pudo validar alcance de empresa", http.StatusInternalServerError)
			return
		}
		if !canAccess {
			http.Error(w, "forbidden: empresa_id fuera del alcance del usuario autenticado", http.StatusForbidden)
			return
		}

		action := defaultPermissionActionFromMethod(r.Method)
		if resolveAction != nil {
			action = normalizePermissionAction(resolveAction(r), action)
		}
		role := normalizePermissionRole(admin.Role)
		if !roleAllowsModuleAction(role, module, action) {
			http.Error(w, "forbidden: rol sin permiso para la accion solicitada", http.StatusForbidden)
			return
		}

		w.Header().Set("X-Empresa-ID", strconv.FormatInt(empresaID, 10))
		next.ServeHTTP(w, r)
	}
}

func extractEmpresaIDForPermissions(r *http.Request) int64 {
	if id, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && id > 0 {
		return id
	}
	if id := parsePositiveInt64(strings.TrimSpace(r.Header.Get("X-Empresa-ID"))); id > 0 {
		return id
	}

	method := strings.ToUpper(strings.TrimSpace(r.Method))
	if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch && method != http.MethodDelete {
		return 0
	}

	contentType := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
	if strings.Contains(contentType, "application/json") {
		return extractEmpresaIDFromJSONBody(r)
	}
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		if err := r.ParseForm(); err == nil {
			if id := parsePositiveInt64(strings.TrimSpace(r.FormValue("empresa_id"))); id > 0 {
				return id
			}
		}
	}
	if strings.Contains(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(12 << 20); err == nil {
			if id := parsePositiveInt64(strings.TrimSpace(r.FormValue("empresa_id"))); id > 0 {
				return id
			}
		}
	}

	return 0
}

func extractEmpresaIDFromJSONBody(r *http.Request) int64 {
	if r.Body == nil {
		return 0
	}
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		r.Body = io.NopCloser(bytes.NewReader(raw))
		return 0
	}
	r.Body = io.NopCloser(bytes.NewReader(raw))
	if len(raw) == 0 {
		return 0
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return 0
	}

	if v, ok := payload["empresa_id"]; ok {
		if id := toPositiveInt64(v); id > 0 {
			return id
		}
	}
	if v, ok := payload["empresaId"]; ok {
		if id := toPositiveInt64(v); id > 0 {
			return id
		}
	}
	if empresaObj, ok := payload["empresa"].(map[string]interface{}); ok {
		if v, exists := empresaObj["id"]; exists {
			if id := toPositiveInt64(v); id > 0 {
				return id
			}
		}
	}
	return 0
}

func toPositiveInt64(v interface{}) int64 {
	switch n := v.(type) {
	case float64:
		if n > 0 {
			return int64(n)
		}
	case int64:
		if n > 0 {
			return n
		}
	case int:
		if n > 0 {
			return int64(n)
		}
	case string:
		return parsePositiveInt64(n)
	}
	return 0
}

func parsePositiveInt64(raw string) int64 {
	v := strings.TrimSpace(raw)
	if v == "" {
		return 0
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil || n <= 0 {
		return 0
	}
	return n
}

func defaultPermissionActionFromMethod(method string) string {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return permActionRead
	case http.MethodPost:
		return permActionCreate
	case http.MethodPut, http.MethodPatch:
		return permActionUpdate
	case http.MethodDelete:
		return permActionDelete
	default:
		return permActionRead
	}
}

func normalizePermissionAction(candidate, fallback string) string {
	v := strings.ToUpper(strings.TrimSpace(candidate))
	if v == "" {
		return fallback
	}
	switch v {
	case permActionRead, permActionCreate, permActionUpdate, permActionDelete, permActionApprove:
		return v
	default:
		return fallback
	}
}

func resolveVentasPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	switch action {
	case "cerrar", "reabrir", "pagar_estacion", "activar_estacion":
		return permActionApprove
	case "activar", "desactivar":
		return permActionUpdate
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func resolveInventarioPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	if action == "activar" || action == "desactivar" {
		return permActionUpdate
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func resolveFinanzasPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	switch action {
	case "cerrar", "reabrir":
		return permActionApprove
	case "anular":
		return permActionDelete
	case "activar", "desactivar":
		return permActionUpdate
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func normalizePermissionRole(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "super_administrador", "superadmin", "super":
		return "super_administrador"
	case "administrador", "admin", "admin_empresa":
		return "admin_empresa"
	case "supervisor", "supervisor_sucursal":
		return "supervisor_sucursal"
	case "cajero":
		return "cajero"
	case "inventario":
		return "inventario"
	case "compras":
		return "compras"
	case "contabilidad", "contador":
		return "contabilidad"
	case "auditor":
		return "auditor"
	default:
		return strings.ToLower(strings.TrimSpace(raw))
	}
}

func roleAllowsModuleAction(role, module, action string) bool {
	if role == "super_administrador" {
		return true
	}

	allReadRoles := []string{"admin_empresa", "supervisor_sucursal", "cajero", "inventario", "compras", "contabilidad", "auditor"}

	switch module {
	case permModuleVentas:
		switch action {
		case permActionRead:
			return roleIn(role, allReadRoles...)
		case permActionCreate, permActionUpdate, permActionDelete, permActionApprove:
			return roleIn(role, "admin_empresa", "supervisor_sucursal", "cajero")
		}

	case permModuleInventario:
		switch action {
		case permActionRead:
			return roleIn(role, allReadRoles...)
		case permActionCreate, permActionUpdate, permActionDelete, permActionApprove:
			return roleIn(role, "admin_empresa", "supervisor_sucursal", "inventario")
		}

	case permModuleFinanzas:
		switch action {
		case permActionRead:
			return roleIn(role, allReadRoles...)
		case permActionCreate, permActionUpdate, permActionApprove:
			return roleIn(role, "admin_empresa", "contabilidad")
		case permActionDelete:
			return roleIn(role, "contabilidad")
		}
	}

	return false
}

func roleIn(role string, allowed ...string) bool {
	role = strings.TrimSpace(strings.ToLower(role))
	if role == "" {
		return false
	}
	for _, it := range allowed {
		if role == strings.TrimSpace(strings.ToLower(it)) {
			return true
		}
	}
	return false
}
