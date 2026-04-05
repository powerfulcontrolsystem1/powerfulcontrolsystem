package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	permActionRead    = "R"
	permActionCreate  = "C"
	permActionUpdate  = "U"
	permActionDelete  = "D"
	permActionApprove = "A"

	permModuleVentas      = "ventas"
	permModuleInventario  = "inventario"
	permModuleFinanzas    = "finanzas"
	permModuleClientes    = "clientes"
	permModuleCompras     = "compras"
	permModuleFacturacion = "facturacion"
	permModuleSeguridad   = "seguridad"
)

var permissionModulesCatalogOrdered = []string{
	permModuleVentas,
	permModuleInventario,
	permModuleFinanzas,
	permModuleClientes,
	permModuleCompras,
	permModuleFacturacion,
	permModuleSeguridad,
}

var permissionActionsCatalogOrdered = []string{
	permActionRead,
	permActionCreate,
	permActionUpdate,
	permActionDelete,
	permActionApprove,
}

var permissionRolesCatalogOrdered = []string{
	"super_administrador",
	"admin_empresa",
	"supervisor_sucursal",
	"cajero",
	"inventario",
	"compras",
	"contabilidad",
	"auditor",
}

type permissionModuleMatrixRow struct {
	Modulo   string          `json:"modulo"`
	Read     bool            `json:"read"`
	Create   bool            `json:"create"`
	Update   bool            `json:"update"`
	Delete   bool            `json:"delete"`
	Approve  bool            `json:"approve"`
	Acciones map[string]bool `json:"acciones"`
}

type permissionSummary struct {
	ModulosTotal        int `json:"modulos_total"`
	ModulosLectura      int `json:"modulos_lectura"`
	ModulosAprobacion   int `json:"modulos_aprobacion"`
	AccionesHabilitadas int `json:"acciones_habilitadas"`
}

type empresaPermisosRolMatriz struct {
	Rol     string                      `json:"rol"`
	Modulos []permissionModuleMatrixRow `json:"modulos"`
	Resumen permissionSummary           `json:"resumen"`
}

type empresaPermisosContextResponse struct {
	EmpresaID        int64                       `json:"empresa_id"`
	AdminEmail       string                      `json:"admin_email"`
	Rol              string                      `json:"rol"`
	AccionesCatalogo []string                    `json:"acciones_catalogo"`
	Modulos          []permissionModuleMatrixRow `json:"modulos"`
	Resumen          permissionSummary           `json:"resumen"`
	IncluyeMatriz    bool                        `json:"incluye_matriz"`
	MatrizRoles      []empresaPermisosRolMatriz  `json:"matriz_roles,omitempty"`
}

// EmpresaPermisosContextoHandler expone el contexto de permisos efectivo por rol/modulo.
// Endpoint recomendado: GET /api/empresa/permisos_contexto?empresa_id={id}[&include_matrix=1]
func EmpresaPermisosContextoHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}

		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		adminEmail := strings.ToLower(strings.TrimSpace(adminEmailFromRequest(r)))
		role := normalizePermissionRole(adminRoleFromRequest(r))
		if role == "" && dbSuper != nil && adminEmail != "" && adminEmail != "sistema" {
			admin, err := dbpkg.GetAdminByEmail(dbSuper, adminEmail)
			if err == nil && admin != nil {
				role = normalizePermissionRole(admin.Role)
			} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
				log.Printf("[authz] permisos_contexto get admin email=%s error: %v", adminEmail, err)
			}
		}
		if role == "" {
			role = "sin_rol"
		}

		modulos := buildPermissionModuleMatrixForRole(role)
		resp := empresaPermisosContextResponse{
			EmpresaID:        empresaID,
			AdminEmail:       adminEmail,
			Rol:              role,
			AccionesCatalogo: append([]string{}, permissionActionsCatalogOrdered...),
			Modulos:          modulos,
			Resumen:          summarizePermissionModules(modulos),
			IncluyeMatriz:    false,
		}

		if queryBool(r, "include_matrix") {
			resp.IncluyeMatriz = true
			resp.MatrizRoles = make([]empresaPermisosRolMatriz, 0, len(permissionRolesCatalogOrdered))
			for _, catalogRole := range permissionRolesCatalogOrdered {
				rows := buildPermissionModuleMatrixForRole(catalogRole)
				resp.MatrizRoles = append(resp.MatrizRoles, empresaPermisosRolMatriz{
					Rol:     catalogRole,
					Modulos: rows,
					Resumen: summarizePermissionModules(rows),
				})
			}
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

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

// WithEmpresaClientesPermissions aplica control de alcance por empresa y permisos por rol para clientes.
func WithEmpresaClientesPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleClientes, resolveClientesPermissionAction, next)
}

// WithEmpresaComprasPermissions aplica control de alcance por empresa y permisos por rol para compras/proveedores.
func WithEmpresaComprasPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleCompras, resolveComprasPermissionAction, next)
}

// WithEmpresaFacturacionPermissions aplica control de alcance por empresa y permisos por rol para facturacion.
func WithEmpresaFacturacionPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleFacturacion, resolveFacturacionPermissionAction, next)
}

// WithEmpresaSeguridadPermissions aplica control de alcance por empresa y permisos por rol para seguridad/usuarios.
func WithEmpresaSeguridadPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleSeguridad, resolveSeguridadPermissionAction, next)
}

// WithEmpresaPublicScope aplica validacion minima de alcance por empresa para endpoints publicos
// que no pueden exigir autenticacion previa (por ejemplo login y primer establecimiento de password).
func WithEmpresaPublicScope(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID := extractEmpresaIDForPermissions(r)
		if empresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), "empresaID", empresaID)
		r = r.WithContext(ctx)
		w.Header().Set("X-Empresa-ID", strconv.FormatInt(empresaID, 10))

		next.ServeHTTP(w, r)
	}
}

func withEmpresaRolePermissions(dbEmp, dbSuper *sql.DB, module string, resolveAction func(*http.Request) string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID := extractEmpresaIDForPermissions(r)
		if empresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}

		action := defaultPermissionActionFromMethod(r.Method)
		if resolveAction != nil {
			action = normalizePermissionAction(resolveAction(r), action)
		}

		adminEmail := strings.ToLower(strings.TrimSpace(adminEmailFromRequest(r)))
		if adminEmail == "" || adminEmail == "sistema" {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusUnauthorized, 0)
			return
		}

		admin, err := dbpkg.GetAdminByEmail(dbSuper, adminEmail)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "unauthenticated", http.StatusUnauthorized)
				registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusUnauthorized, 0)
				return
			}
			log.Printf("[authz] get admin email=%s error: %v", adminEmail, err)
			http.Error(w, "No se pudo validar permisos del usuario", http.StatusInternalServerError)
			registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusInternalServerError, 0)
			return
		}

		canAccess, err := dbpkg.CanAdminAccessEmpresaIA(dbEmp, dbSuper, adminEmail, empresaID)
		if err != nil {
			log.Printf("[authz] alcance empresa module=%s email=%s empresa_id=%d error: %v", module, adminEmail, empresaID, err)
			http.Error(w, "No se pudo validar alcance de empresa", http.StatusInternalServerError)
			registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusInternalServerError, 0)
			return
		}
		if !canAccess {
			http.Error(w, "forbidden: empresa_id fuera del alcance del usuario autenticado", http.StatusForbidden)
			registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusForbidden, 0)
			return
		}

		role := normalizePermissionRole(admin.Role)
		if !roleAllowsModuleAction(role, module, action) {
			http.Error(w, "forbidden: rol sin permiso para la accion solicitada", http.StatusForbidden)
			registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusForbidden, 0)
			return
		}

		ctx := context.WithValue(r.Context(), "adminRole", role)
		ctx = context.WithValue(ctx, "empresaID", empresaID)
		r = r.WithContext(ctx)

		w.Header().Set("X-Empresa-ID", strconv.FormatInt(empresaID, 10))
		r.Header.Set("X-Admin-Role", role)

		auditStart := time.Now()
		auditRW := &auditCaptureResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(auditRW, r)
		registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, auditRW.status, time.Since(auditStart))
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
	case "cerrar", "reabrir", "pagar_estacion", "activar_estacion", "pagar", "suspender", "suspender_venta", "reactivar", "reabrir_venta":
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
	case "cerrar", "reabrir", "aprobar", "procesar_asientos", "procesar":
		return permActionApprove
	case "anular":
		return permActionDelete
	case "activar", "desactivar":
		return permActionUpdate
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func resolveClientesPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	if action == "activar" || action == "desactivar" {
		return permActionUpdate
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func resolveComprasPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	if action == "activar" || action == "desactivar" {
		return permActionUpdate
	}
	if action == "anular" || action == "cancelar" {
		return permActionDelete
	}
	if action == "aprobar" || action == "cerrar" || action == "emitir" || action == "emitir_orden" || action == "recepcionar" || action == "recepcionar_compra" || action == "contabilizar" || action == "contabilizar_compra" {
		return permActionApprove
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func resolveFacturacionPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	if action == "activar" || action == "desactivar" {
		return permActionUpdate
	}
	if action == "aprobar" || action == "emitir" || action == "emitir_factura" || action == "emitir_documento" || action == "nota_credito" || action == "emitir_nota_credito" {
		return permActionApprove
	}
	if action == "anular" {
		return permActionDelete
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func resolveSeguridadPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	switch action {
	case "activar", "desactivar":
		return permActionUpdate
	case "reenviar_confirmacion":
		return permActionApprove
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

	case permModuleClientes:
		switch action {
		case permActionRead:
			return roleIn(role, allReadRoles...)
		case permActionCreate, permActionUpdate, permActionApprove:
			return roleIn(role, "admin_empresa", "supervisor_sucursal", "cajero")
		case permActionDelete:
			return false
		}

	case permModuleCompras:
		switch action {
		case permActionRead:
			return roleIn(role, allReadRoles...)
		case permActionCreate, permActionUpdate, permActionApprove:
			return roleIn(role, "admin_empresa", "supervisor_sucursal", "compras")
		case permActionDelete:
			return false
		}

	case permModuleFacturacion:
		switch action {
		case permActionRead:
			return roleIn(role, allReadRoles...)
		case permActionCreate, permActionUpdate, permActionApprove:
			return roleIn(role, "admin_empresa", "cajero")
		case permActionDelete:
			return false
		}

	case permModuleSeguridad:
		switch action {
		case permActionRead:
			return roleIn(role, allReadRoles...)
		case permActionCreate, permActionUpdate, permActionDelete, permActionApprove:
			return roleIn(role, "admin_empresa")
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

func buildPermissionModuleMatrixForRole(role string) []permissionModuleMatrixRow {
	normalizedRole := normalizePermissionRole(role)
	out := make([]permissionModuleMatrixRow, 0, len(permissionModulesCatalogOrdered))
	for _, modulo := range permissionModulesCatalogOrdered {
		readAllowed := roleAllowsModuleAction(normalizedRole, modulo, permActionRead)
		createAllowed := roleAllowsModuleAction(normalizedRole, modulo, permActionCreate)
		updateAllowed := roleAllowsModuleAction(normalizedRole, modulo, permActionUpdate)
		deleteAllowed := roleAllowsModuleAction(normalizedRole, modulo, permActionDelete)
		approveAllowed := roleAllowsModuleAction(normalizedRole, modulo, permActionApprove)

		out = append(out, permissionModuleMatrixRow{
			Modulo:  modulo,
			Read:    readAllowed,
			Create:  createAllowed,
			Update:  updateAllowed,
			Delete:  deleteAllowed,
			Approve: approveAllowed,
			Acciones: map[string]bool{
				permActionRead:    readAllowed,
				permActionCreate:  createAllowed,
				permActionUpdate:  updateAllowed,
				permActionDelete:  deleteAllowed,
				permActionApprove: approveAllowed,
			},
		})
	}
	return out
}

func summarizePermissionModules(rows []permissionModuleMatrixRow) permissionSummary {
	summary := permissionSummary{ModulosTotal: len(rows)}
	for _, row := range rows {
		if row.Read {
			summary.ModulosLectura++
			summary.AccionesHabilitadas++
		}
		if row.Create {
			summary.AccionesHabilitadas++
		}
		if row.Update {
			summary.AccionesHabilitadas++
		}
		if row.Delete {
			summary.AccionesHabilitadas++
		}
		if row.Approve {
			summary.ModulosAprobacion++
			summary.AccionesHabilitadas++
		}
	}
	return summary
}
