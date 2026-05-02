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
	"sync"
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

	permissionApprovalHeaderBy       = "X-Permission-Approved-By"
	permissionApprovalHeaderCode     = "X-Permission-Approval-Code"
	permissionApprovalHeaderReason   = "X-Permission-Approval-Reason"
	permissionApprovalHeaderRequired = "X-Permission-Approval-Required"
)

type permissionApprovalEvidence struct {
	ApprovedBy   string
	ApprovalCode string
	Reason       string
}

type empresaRateLimitBucket struct {
	WindowStart time.Time
	Count       int64
}

type empresaPermissionSnapshot struct {
	AdminRole         string
	EffectiveRole     string
	CanAccess         bool
	AllowedModules    map[string]bool
	RoleModuleActions map[string]bool
	AllowedPages      map[string]bool
	LoadedAt          time.Time
}

type empresaPermissionSnapshotInflight struct {
	done     chan struct{}
	snapshot empresaPermissionSnapshot
	err      error
}

type permissionRoleOverrideCacheEntry struct {
	ModuleOverrides map[string]bool
	PageOverrides   map[string]bool
	LoadedAt        time.Time
}

type empresaPermissionOverrideCacheEntry struct {
	ModuleOverrides map[string]bool
	PageOverrides   map[string]bool
	Ctx             *empresaPermisosFinosCtx
	LoadedAt        time.Time
}

type permissionRoleModuleMatrixCacheEntry struct {
	Rows     []permissionModuleMatrixRow
	LoadedAt time.Time
}

var (
	empresaRateLimitMu              sync.Mutex
	empresaRateLimitBuckets         = map[string]empresaRateLimitBucket{}
	empresaPermissionCacheMu        sync.Mutex
	empresaPermissionCache          = map[string]empresaPermissionSnapshot{}
	empresaPermissionInflight       = map[string]*empresaPermissionSnapshotInflight{}
	rolePermissionModuleMatrixCache = map[string]permissionRoleModuleMatrixCacheEntry{}
	rolePermissionOverrideCache     = map[string]permissionRoleOverrideCacheEntry{}
	empresaPermissionOverrideCache  = map[int64]empresaPermissionOverrideCacheEntry{}
)

const empresaPermissionCacheTTL = 60 * time.Second
const permissionOverrideCacheTTL = 60 * time.Second

var legacyPermissionVisibleTextReplacer = strings.NewReplacer(
	"Operaci\u00c3\u00b3n", "Operaci\u00f3n",
	"Configuraci\u00c3\u00b3n", "Configuraci\u00f3n",
	"Facturaci\u00c3\u00b3n", "Facturaci\u00f3n",
	"electr\u00c3\u00b3nica", "electr\u00f3nica",
	"cat\u00c3\u00a1logo", "cat\u00e1logo",
	"c\u00c3\u00b3digos", "c\u00f3digos",
	"c\u00c3\u00b3digo", "c\u00f3digo",
	"\u00c3\u00b3rdenes", "\u00f3rdenes",
	"N\u00c3\u00b3mina", "N\u00f3mina",
	"veh\u00c3\u00adculos", "veh\u00edculos",
	"veh\u00c3\u00adculo", "veh\u00edculo",
	"Auditor\u00c3\u00ada", "Auditor\u00eda",
	"Cr\u00c3\u00a9ditos", "Cr\u00e9ditos",
	"cr\u00c3\u00a9ditos", "cr\u00e9ditos",
	"Ubicaci\u00c3\u00b3n", "Ubicaci\u00f3n",
	"Aprobaci\u00c3\u00b3n", "Aprobaci\u00f3n",
	"d\u00c3\u00ada", "d\u00eda",
	"Gr\u00c3\u00a1ficos", "Gr\u00e1ficos",
	"estad\u00c3\u00adsticas", "estad\u00edsticas",
	"m\u00c3\u00b3dulo", "m\u00f3dulo",
	"acci\u00c3\u00b3n", "acci\u00f3n",
	"integraci\u00c3\u00b3n", "integraci\u00f3n",
)

func sanitizeLegacyPermissionVisibleText(value string) string {
	clean := strings.TrimSpace(value)
	if clean == "" {
		return ""
	}
	return strings.TrimSpace(legacyPermissionVisibleTextReplacer.Replace(clean))
}

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

// Etiquetas cortas para UI (super: permisos por rol) y documentaciÃƒÂ³n.
var permissionActionDisplayNames = map[string]string{
	permActionRead:    "Leer / consultar",
	permActionCreate:  "Crear / registrar",
	permActionUpdate:  "Actualizar / modificar",
	permActionDelete:  "Eliminar / anular",
	permActionApprove: "Aprobar / auditar",
}

// permissionModuleDisplayNames nombres de negocio por clave de mÃƒÂ³dulo.
var permissionModuleDisplayNames = map[string]string{
	permModuleVentas:      "Ventas y servicio al cliente",
	permModuleInventario:  "Inventario y almacÃƒÂ©n",
	permModuleFinanzas:    "Finanzas, caja y reportes",
	permModuleClientes:    "Clientes y cartera comercial",
	permModuleCompras:     "Compras y proveedores",
	permModuleFacturacion: "FacturaciÃƒÂ³n electrÃƒÂ³nica (DIAN)",
	permModuleSeguridad:   "Seguridad, usuarios e integraciÃƒÂ³n",
}

var permissionRolesCatalogOrdered = []string{
	"super_administrador",
	"administrador_total",
	"admin_empresa",
	"supervisor_sucursal",
	"cajero",
	"inventario",
	"compras",
	"contabilidad",
	"auditor",
}

type permissionPageRule struct {
	PaginaClave   string `json:"pagina_clave"`
	Modulo        string `json:"modulo,omitempty"`
	Accion        string `json:"accion,omitempty"`
	AlwaysVisible bool   `json:"always_visible,omitempty"`
	Titulo        string `json:"titulo,omitempty"`
	Grupo         string `json:"grupo,omitempty"`
}

var permissionPagesCatalogOrdered = []permissionPageRule{
	{PaginaClave: "linkInicio", AlwaysVisible: true, Titulo: "Inicio (tablero)", Grupo: "Acceso general"},
	{PaginaClave: "linkPanelEmpresa", AlwaysVisible: true, Titulo: "Panel de empresa", Grupo: "Acceso general"},
	{PaginaClave: "linkVentas", Modulo: permModuleVentas, Accion: permActionRead, Titulo: "Punto de venta / TPV", Grupo: "OperaciÃƒÂ³n y venta"},
	{PaginaClave: "linkCarritoCompras", Modulo: permModuleVentas, Accion: permActionCreate, Titulo: "Carritos de compra", Grupo: "OperaciÃƒÂ³n y venta"},
	{PaginaClave: "linkVentaDirecta", Modulo: permModuleVentas, Accion: permActionCreate, Titulo: "Venta directa", Grupo: "OperaciÃƒÂ³n y venta"},
	{PaginaClave: "linkGimnasio", Modulo: permModuleVentas, Accion: permActionCreate, Titulo: "GestiÃ³n de gimnasio", Grupo: "OperaciÃ³n y venta"},
	{PaginaClave: "linkVentaPublica", Modulo: permModuleVentas, Accion: permActionCreate, Titulo: "Venta pÃƒÂºblica (eÃ¢â‚¬â€˜commerce)", Grupo: "OperaciÃƒÂ³n y venta"},
	{PaginaClave: "linkProductos", Modulo: permModuleInventario, Accion: permActionCreate, Titulo: "Productos y servicios", Grupo: "Inventario y catÃƒÂ¡logo"},
	{PaginaClave: "linkCombosProductos", Modulo: permModuleInventario, Accion: permActionCreate, Titulo: "Combos y paquetes", Grupo: "Inventario y catÃƒÂ¡logo"},
	{PaginaClave: "linkGeneradorCodigosBarras", Modulo: permModuleInventario, Accion: permActionUpdate, Titulo: "Generador de cÃƒÂ³digos de barras", Grupo: "Inventario y catÃƒÂ¡logo"},
	{PaginaClave: "linkCodigosDescuento", Modulo: permModuleVentas, Accion: permActionCreate, Titulo: "CÃƒÂ³digos de descuento", Grupo: "OperaciÃƒÂ³n y venta"},
	{PaginaClave: "linkCompras", Modulo: permModuleCompras, Accion: permActionCreate, Titulo: "Compras y ÃƒÂ³rdenes", Grupo: "Compras"},
	{PaginaClave: "linkConfiguracion", Modulo: permModuleSeguridad, Accion: permActionUpdate, Titulo: "ConfiguraciÃƒÂ³n de empresa", Grupo: "Seguridad e integraciÃƒÂ³n"},
	{PaginaClave: "linkConfiguracionImpresora", Modulo: permModuleSeguridad, Accion: permActionUpdate, Titulo: "ConfiguraciÃƒÂ³n de impresora", Grupo: "Seguridad e integraciÃƒÂ³n"},
	{PaginaClave: "linkUsuarios", Modulo: permModuleSeguridad, Accion: permActionUpdate, Titulo: "Usuarios y accesos", Grupo: "Seguridad e integraciÃƒÂ³n"},
	{PaginaClave: "linkHorariosTrabajadores", Modulo: permModuleSeguridad, Accion: permActionUpdate, Titulo: "Horarios laborales", Grupo: "Seguridad e integraciÃƒÂ³n"},
	{PaginaClave: "linkAsistenciaEmpleados", Modulo: permModuleSeguridad, Accion: permActionUpdate, Titulo: "Asistencia de empleados", Grupo: "Seguridad e integraciÃƒÂ³n"},
	{PaginaClave: "linkNominaSueldos", Modulo: permModuleFinanzas, Accion: permActionCreate, Titulo: "NÃƒÂ³mina y sueldos", Grupo: "Finanzas y nÃƒÂ³mina"},
	{PaginaClave: "linkVehiculosRegistro", Modulo: permModuleSeguridad, Accion: permActionCreate, Titulo: "Registro de vehÃƒÂ­culos", Grupo: "Seguridad e integraciÃƒÂ³n"},
	{PaginaClave: "linkHojaVidaOperativa", Modulo: permModuleSeguridad, Accion: permActionUpdate, Titulo: "Hoja de vida operativa", Grupo: "Seguridad e integraciÃƒÂ³n"},
	{PaginaClave: "linkAuditoria", Modulo: permModuleSeguridad, Accion: permActionRead, Titulo: "AuditorÃƒÂ­a de acciones", Grupo: "Seguridad e integraciÃƒÂ³n"},
	{PaginaClave: "linkChatTareas", Modulo: permModuleVentas, Accion: permActionCreate, Titulo: "Chat y tareas", Grupo: "OperaciÃƒÂ³n y venta"},
	{PaginaClave: "linkClientes", Modulo: permModuleClientes, Accion: permActionCreate, Titulo: "Clientes y CRM bÃƒÂ¡sico", Grupo: "Clientes"},
	{PaginaClave: "linkCRMComercial", Modulo: permModuleClientes, Accion: permActionCreate, Titulo: "CRM comercial y embudo", Grupo: "Clientes"},
	{PaginaClave: "linkFacturacionElectronica", Modulo: permModuleFacturacion, Accion: permActionCreate, Titulo: "FacturaciÃƒÂ³n electrÃƒÂ³nica (emitir)", Grupo: "FacturaciÃƒÂ³n DIAN"},
	{PaginaClave: "linkFacturasElectronicas", Modulo: permModuleFacturacion, Accion: permActionRead, Titulo: "Documentos y consultas FE", Grupo: "FacturaciÃƒÂ³n DIAN"},
	{PaginaClave: "linkERPExtendido", Modulo: permModuleSeguridad, Accion: permActionUpdate, Titulo: "Integraciones / ERP extendido", Grupo: "Seguridad e integraciÃƒÂ³n"},
	{PaginaClave: "linkChatIA", Modulo: permModuleVentas, Accion: permActionRead, Titulo: "Asistente IA (chat empresarial)", Grupo: "OperaciÃƒÂ³n y venta"},
	{PaginaClave: "linkChatIAGlobal", Modulo: permModuleSeguridad, Accion: permActionRead, Titulo: "Chat IA global (super)", Grupo: "Seguridad e integraciÃƒÂ³n"},
	{PaginaClave: "linkFinanzas", Modulo: permModuleFinanzas, Accion: permActionCreate, Titulo: "Finanzas y movimientos", Grupo: "Finanzas y reportes"},
	{PaginaClave: "linkCreditos", Modulo: permModuleFinanzas, Accion: permActionCreate, Titulo: "CrÃƒÂ©ditos y cartera", Grupo: "Finanzas y reportes"},
	{PaginaClave: "linkBackups", Modulo: permModuleSeguridad, Accion: permActionApprove, Titulo: "Backups empresariales", Grupo: "Seguridad e integraciÃƒÂ³n"},
	{PaginaClave: "linkSoporteRemoto", Modulo: permModuleSeguridad, Accion: permActionApprove, Titulo: "Soporte remoto", Grupo: "Seguridad e integraciÃƒÂ³n"},
	{PaginaClave: "linkPropinas", Modulo: permModuleFinanzas, Accion: permActionCreate, Titulo: "Propinas", Grupo: "Finanzas y reportes"},
	{PaginaClave: "linkComisiones", Modulo: permModuleFinanzas, Accion: permActionCreate, Titulo: "Comisiones de personal", Grupo: "Finanzas y reportes"},
	{PaginaClave: "linkUbicacionGPS", Modulo: permModuleInventario, Accion: permActionCreate, Titulo: "UbicaciÃƒÂ³n / GPS (activos)", Grupo: "Inventario y catÃƒÂ¡logo"},
	{PaginaClave: "linkConfigEstaciones", Modulo: permModuleVentas, Accion: permActionApprove, Titulo: "AprobaciÃƒÂ³n: configuraciÃƒÂ³n de estaciones", Grupo: "OperaciÃƒÂ³n y venta"},
	{PaginaClave: "linkConfiguracionSensoresRaspberry", Modulo: permModuleSeguridad, Accion: permActionUpdate, Titulo: "Raspberry Pi y sensores", Grupo: "Seguridad e integraciÃƒÂ³n"},
	{PaginaClave: "linkTarifasPorMinutos", Modulo: permModuleVentas, Accion: permActionCreate, Titulo: "Tarifas por minutos", Grupo: "OperaciÃƒÂ³n y venta"},
	{PaginaClave: "linkTarifasPorDia", Modulo: permModuleVentas, Accion: permActionCreate, Titulo: "Tarifas por dÃƒÂ­a", Grupo: "OperaciÃƒÂ³n y venta"},
	{PaginaClave: "linkEstaciones", Modulo: permModuleVentas, Accion: permActionUpdate, Titulo: "Estaciones y terminales", Grupo: "OperaciÃƒÂ³n y venta"},
	{PaginaClave: "linkReservasHotel", Modulo: permModuleVentas, Accion: permActionCreate, Titulo: "Reservas (hotel / habitaciones)", Grupo: "OperaciÃƒÂ³n y venta"},
	{PaginaClave: "linkReportes", Modulo: permModuleFinanzas, Accion: permActionRead, Titulo: "Reportes e informes", Grupo: "Finanzas y reportes"},
	{PaginaClave: "linkReportesIAChat", Modulo: permModuleFinanzas, Accion: permActionRead, Titulo: "Chat IA de reportes", Grupo: "Finanzas y reportes"},
	{PaginaClave: "linkCalculadora", Modulo: permModuleFinanzas, Accion: permActionRead, Titulo: "Calculadora financiera", Grupo: "Finanzas y reportes"},
	{PaginaClave: "linkGraficosEstadisticas", Modulo: permModuleFinanzas, Accion: permActionRead, Titulo: "AnalÃƒÂ­tica ejecutiva avanzada", Grupo: "Finanzas y reportes"},
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

type permissionPageAccessRow struct {
	PaginaClave   string `json:"pagina_clave"`
	Modulo        string `json:"modulo,omitempty"`
	Accion        string `json:"accion,omitempty"`
	Permitido     bool   `json:"permitido"`
	AlwaysVisible bool   `json:"always_visible,omitempty"`
	Titulo        string `json:"titulo,omitempty"`
	Grupo         string `json:"grupo,omitempty"`
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
	RolEfectivo      string                      `json:"rol_efectivo,omitempty"`
	AccionesCatalogo []string                    `json:"acciones_catalogo"`
	Modulos          []permissionModuleMatrixRow `json:"modulos"`
	Paginas          map[string]bool             `json:"paginas,omitempty"`
	Resumen          permissionSummary           `json:"resumen"`
	Licencia         *empresaPermisosLicenciaCtx `json:"licencia,omitempty"`
	EmpresaPolicy    *empresaPermisosFinosCtx    `json:"empresa_policy,omitempty"`
	IncluyeMatriz    bool                        `json:"incluye_matriz"`
	MatrizRoles      []empresaPermisosRolMatriz  `json:"matriz_roles,omitempty"`
}

type empresaPermisosLicenciaCtx struct {
	LicenciaID         int64    `json:"licencia_id,omitempty"`
	Nombre             string   `json:"nombre,omitempty"`
	ModulosHabilitados []string `json:"modulos_habilitados,omitempty"`
	SuperRolHabilitado bool     `json:"super_rol_habilitado"`
	RestringeModulos   bool     `json:"restringe_modulos"`
}

type empresaPermisosFinosCtx struct {
	ReglasModulo int  `json:"reglas_modulo"`
	ReglasPagina int  `json:"reglas_pagina"`
	Activo       bool `json:"activo"`
}

type empresaPermisoModuloPayload struct {
	Modulo    string `json:"modulo"`
	Accion    string `json:"accion"`
	Permitido bool   `json:"permitido"`
}

type empresaPermisoPaginaPayload struct {
	PaginaClave string `json:"pagina_clave"`
	Permitido   bool   `json:"permitido"`
}

type empresaPermisosFinosPayload struct {
	EmpresaID      int64                         `json:"empresa_id"`
	PermisosModulo []empresaPermisoModuloPayload `json:"permisos_modulo"`
	PermisosPagina []empresaPermisoPaginaPayload `json:"permisos_pagina"`
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

		licenciaPolicy, err := dbpkg.GetLicenciaPermisoPolicyByEmpresa(dbSuper, empresaID)
		if err != nil {
			log.Printf("[authz] permisos_contexto licencia empresa=%d error: %v", empresaID, err)
		}

		allowedModules, allowedModulesList := parseLicenciaModulosCSV("")
		if licenciaPolicy != nil {
			allowedModules, allowedModulesList = parseLicenciaModulosCSV(licenciaPolicy.ModulosHabilitados)
		}
		effectiveRole := resolveEffectiveRoleByLicencia(role, licenciaPolicy)

		modulos := buildPermissionModuleMatrixForRoleDynamic(dbSuper, effectiveRole)
		modulos = applyLicenciaRestriccionesToModuleRows(modulos, allowedModules)
		empresaModuleOverrides, empresaPageOverrides, empresaPolicyCtx := loadEmpresaPermissionOverrides(dbSuper, empresaID)
		modulos = applyEmpresaRestriccionesToModuleRows(modulos, empresaModuleOverrides)
		paginas := buildPermissionPagesMapForRoleDynamic(dbSuper, effectiveRole, modulos)
		paginas = applyEmpresaPageRestrictionsToMap(paginas, empresaPageOverrides)

		var licenciaCtx *empresaPermisosLicenciaCtx
		if licenciaPolicy != nil {
			licenciaCtx = &empresaPermisosLicenciaCtx{
				LicenciaID:         licenciaPolicy.LicenciaID,
				Nombre:             strings.TrimSpace(licenciaPolicy.Nombre),
				ModulosHabilitados: append([]string{}, allowedModulesList...),
				SuperRolHabilitado: licenciaPolicy.SuperRolHabilitado,
				RestringeModulos:   len(allowedModules) > 0,
			}
		}

		resp := empresaPermisosContextResponse{
			EmpresaID:        empresaID,
			AdminEmail:       adminEmail,
			Rol:              role,
			RolEfectivo:      effectiveRole,
			AccionesCatalogo: append([]string{}, permissionActionsCatalogOrdered...),
			Modulos:          modulos,
			Paginas:          paginas,
			Resumen:          summarizePermissionModules(modulos),
			Licencia:         licenciaCtx,
			EmpresaPolicy:    empresaPolicyCtx,
			IncluyeMatriz:    false,
		}

		if queryBool(r, "include_matrix") {
			resp.IncluyeMatriz = true
			resp.MatrizRoles = make([]empresaPermisosRolMatriz, 0, len(permissionRolesCatalogOrdered))
			for _, catalogRole := range permissionRolesCatalogOrdered {
				rows := buildPermissionModuleMatrixForRoleDynamic(dbSuper, catalogRole)
				rows = applyLicenciaRestriccionesToModuleRows(rows, allowedModules)
				rows = applyEmpresaRestriccionesToModuleRows(rows, empresaModuleOverrides)
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

// EmpresaPermisosFinosHandler administra el techo fino de modulos, acciones y paginas para una empresa.
func EmpresaPermisosFinosHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := dbpkg.EnsureEmpresaPermisosFinosSchema(dbSuper); err != nil {
			http.Error(w, "failed to ensure empresa permisos finos schema: "+err.Error(), http.StatusInternalServerError)
			return
		}

		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			modulos := buildEmpresaPermisosDefaultModuleRows()
			moduleItems, err := dbpkg.ListEmpresaPermisosModuloByEmpresaID(dbSuper, empresaID)
			if err != nil {
				http.Error(w, "failed to load empresa modulo permisos: "+err.Error(), http.StatusInternalServerError)
				return
			}
			moduleOverrides := make(map[string]bool, len(moduleItems))
			for _, item := range moduleItems {
				moduleOverrides[permissionModuleActionKey(item.Modulo, item.Accion)] = item.Permitido
			}
			modulos = applyEmpresaRestriccionesToModuleRows(modulos, moduleOverrides)

			pageItems, err := dbpkg.ListEmpresaPermisosPaginaByEmpresaID(dbSuper, empresaID)
			if err != nil {
				http.Error(w, "failed to load empresa pagina permisos: "+err.Error(), http.StatusInternalServerError)
				return
			}
			pageOverrides := make(map[string]bool, len(pageItems))
			for _, item := range pageItems {
				pageOverrides[strings.TrimSpace(item.PaginaClave)] = item.Permitido
			}
			paginas := buildPermissionPagesCatalogFromModuleRows(modulos, pageOverrides)

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"empresa_id":          empresaID,
				"acciones_catalogo":   append([]string{}, permissionActionsCatalogOrdered...),
				"acciones_etiqueta":   PermissionActionDisplayNameMap(),
				"modulos_catalogo":    append([]string{}, permissionModulesCatalogOrdered...),
				"modulos_etiqueta":    PermissionModuleDisplayNameMap(),
				"modulos":             modulos,
				"paginas":             paginas,
				"reglas_modulo":       len(moduleItems),
				"reglas_pagina":       len(pageItems),
				"comportamiento_base": "sin reglas guardadas, la empresa no restringe el catalogo; licencia y rol siguen aplicando",
			})
			return

		case http.MethodPut:
			var payload empresaPermisosFinosPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			empresaID := payload.EmpresaID
			if empresaID <= 0 {
				if qID, err := parseOptionalInt64Query(r, "empresa_id"); err == nil && qID > 0 {
					empresaID = qID
				}
			}
			if empresaID <= 0 {
				http.Error(w, "empresa_id required", http.StatusBadRequest)
				return
			}

			moduleRows := make([]dbpkg.EmpresaPermisoModulo, 0, len(payload.PermisosModulo))
			for _, item := range payload.PermisosModulo {
				moduleRows = append(moduleRows, dbpkg.EmpresaPermisoModulo{
					EmpresaID: empresaID,
					Modulo:    strings.ToLower(strings.TrimSpace(item.Modulo)),
					Accion:    strings.ToUpper(strings.TrimSpace(item.Accion)),
					Permitido: item.Permitido,
				})
			}

			pageRows := make([]dbpkg.EmpresaPermisoPagina, 0, len(payload.PermisosPagina))
			for _, item := range payload.PermisosPagina {
				pageRows = append(pageRows, dbpkg.EmpresaPermisoPagina{
					EmpresaID:   empresaID,
					PaginaClave: strings.TrimSpace(item.PaginaClave),
					Permitido:   item.Permitido,
				})
			}

			if err := dbpkg.ReplaceEmpresaPermisosFinos(dbSuper, empresaID, moduleRows, pageRows, adminEmailFromRequest(r)); err != nil {
				http.Error(w, "failed to save empresa permisos finos: "+err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusNoContent)
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
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

func empresaRateLimitScopeForRequest(r *http.Request) string {
	path := ""
	if r != nil && r.URL != nil {
		path = strings.TrimSpace(r.URL.Path)
	}
	if strings.HasPrefix(path, "/api/empresa/db_admin") {
		return "db_admin"
	}
	return "api"
}

func empresaRateLimitMaxForRequest(dbSuper *sql.DB, r *http.Request) int64 {
	if dbSuper == nil {
		if empresaRateLimitScopeForRequest(r) == "db_admin" {
			return defaultEmpresaDBQueriesPerMinute
		}
		return defaultEmpresaAPIRequestsPerMinute
	}
	switch empresaRateLimitScopeForRequest(r) {
	case "db_admin":
		value, _, _, err := getLimitacionInt64(dbSuper, superEmpresaLimitDBQueriesPerMinuteKey, defaultEmpresaDBQueriesPerMinute)
		if err != nil {
			log.Printf("[rate_limit] no se pudo leer limite db_admin: %v", err)
			return defaultEmpresaDBQueriesPerMinute
		}
		return value
	default:
		value, _, _, err := getLimitacionInt64(dbSuper, superEmpresaLimitAPIRequestsPerMinuteKey, defaultEmpresaAPIRequestsPerMinute)
		if err != nil {
			log.Printf("[rate_limit] no se pudo leer limite api: %v", err)
			return defaultEmpresaAPIRequestsPerMinute
		}
		return value
	}
}

func checkEmpresaRateLimitAt(now time.Time, empresaID int64, scope string, maxPerMinute int64) (allowed bool, remaining int64, retryAfterSeconds int64, current int64) {
	if empresaID <= 0 || maxPerMinute <= 0 {
		return true, 0, 0, 0
	}
	scope = strings.TrimSpace(strings.ToLower(scope))
	if scope == "" {
		scope = "api"
	}
	if now.IsZero() {
		now = time.Now()
	}
	windowStart := now.Truncate(time.Minute)
	key := scope + ":" + strconv.FormatInt(empresaID, 10)

	empresaRateLimitMu.Lock()
	defer empresaRateLimitMu.Unlock()

	bucket := empresaRateLimitBuckets[key]
	if bucket.WindowStart.IsZero() || !bucket.WindowStart.Equal(windowStart) {
		bucket = empresaRateLimitBucket{WindowStart: windowStart}
	}
	if bucket.Count >= maxPerMinute {
		retryAfter := int64(bucket.WindowStart.Add(time.Minute).Sub(now).Seconds())
		if retryAfter < 1 {
			retryAfter = 1
		}
		empresaRateLimitBuckets[key] = bucket
		return false, 0, retryAfter, bucket.Count
	}

	bucket.Count++
	empresaRateLimitBuckets[key] = bucket
	remaining = maxPerMinute - bucket.Count
	if remaining < 0 {
		remaining = 0
	}
	return true, remaining, 0, bucket.Count
}

func applyEmpresaRateLimitHeaders(w http.ResponseWriter, limit, remaining, retryAfterSeconds int64) {
	if w == nil {
		return
	}
	if limit > 0 {
		w.Header().Set("X-Empresa-RateLimit-Limit", strconv.FormatInt(limit, 10))
		w.Header().Set("X-Empresa-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
	}
	if retryAfterSeconds > 0 {
		w.Header().Set("Retry-After", strconv.FormatInt(retryAfterSeconds, 10))
	}
}

func withEmpresaRolePermissions(dbEmp, dbSuper *sql.DB, module string, resolveAction func(*http.Request) string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		defer func() {
			dbpkg.PerfLogf("[perf][authz] module=%s method=%s path=%s dur=%s", module, r.Method, r.URL.Path, time.Since(startedAt))
		}()
		empresaID := extractEmpresaIDForPermissions(r)
		if empresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}

		action := defaultPermissionActionFromMethod(r.Method)
		if resolveAction != nil {
			action = normalizePermissionAction(resolveAction(r), action)
		}

		rateLimit := empresaRateLimitMaxForRequest(dbSuper, r)
		rateScope := empresaRateLimitScopeForRequest(r)
		allowedByRate, remaining, retryAfter, current := checkEmpresaRateLimitAt(time.Now(), empresaID, rateScope, rateLimit)
		applyEmpresaRateLimitHeaders(w, rateLimit, remaining, retryAfter)
		if !allowedByRate {
			path := ""
			if r.URL != nil {
				path = strings.TrimSpace(r.URL.Path)
			}
			log.Printf("[rate_limit] empresa_id=%d scope=%s limite=%d actual=%d path=%s", empresaID, rateScope, rateLimit, current, path)
			http.Error(w, "limite de consumo por empresa excedido; intenta de nuevo en unos segundos", http.StatusTooManyRequests)
			registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusTooManyRequests, 0)
			return
		}

		adminEmail := strings.ToLower(strings.TrimSpace(adminEmailFromRequest(r)))
		if adminEmail == "" || adminEmail == "sistema" {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusUnauthorized, 0)
			return
		}

		snapshotStartedAt := time.Now()
		snapshot, err := getEmpresaPermissionSnapshot(dbEmp, dbSuper, adminEmail, empresaID)
		dbpkg.PerfLogf("[perf][authz] module=%s snapshot empresa=%d email=%s dur=%s", module, empresaID, adminEmail, time.Since(snapshotStartedAt))
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "unauthenticated", http.StatusUnauthorized)
				registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusUnauthorized, 0)
				return
			}
			log.Printf("[authz] snapshot module=%s email=%s empresa_id=%d error: %v", module, adminEmail, empresaID, err)
			http.Error(w, "No se pudo validar permisos del usuario", http.StatusInternalServerError)
			registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusInternalServerError, 0)
			return
		}

		if !snapshot.CanAccess {
			http.Error(w, "forbidden: empresa_id fuera del alcance del usuario autenticado", http.StatusForbidden)
			registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusForbidden, 0)
			return
		}

		role := snapshot.AdminRole
		skipLicenciaModuloCheck := module == permModuleSeguridad && strings.HasPrefix(strings.TrimSpace(r.URL.Path), "/api/empresa/permisos_contexto")
		if !skipLicenciaModuloCheck && !isModuloPermitidoByLicencia(module, snapshot.AllowedModules) {
			http.Error(w, "forbidden: modulo no habilitado por licencia activa", http.StatusForbidden)
			registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusForbidden, 0)
			return
		}

		requestPath := strings.TrimSpace(r.URL.Path)
		effectiveRole := snapshot.EffectiveRole
		skipRoleModuloCheck := module == permModuleSeguridad && strings.HasPrefix(requestPath, "/api/empresa/permisos_contexto")
		if !skipRoleModuloCheck && !snapshot.RoleModuleActions[permissionModuleActionKey(module, action)] {
			http.Error(w, "forbidden: rol sin permiso para la accion solicitada", http.StatusForbidden)
			registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusForbidden, 0)
			return
		}
		if pageKey := resolvePermissionPageKeyForRequest(r); pageKey != "" {
			if !snapshot.AllowedPages[pageKey] {
				http.Error(w, "forbidden: rol sin acceso a la funcionalidad solicitada", http.StatusForbidden)
				registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusForbidden, 0)
				return
			}
		}

		if permissionChangeRequiresApproval(module, r, action) {
			evidence, err := extractPermissionApprovalEvidence(r)
			if err != nil {
				http.Error(w, "no se pudo validar evidencia de aprobacion para el cambio de permisos", http.StatusBadRequest)
				registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusBadRequest, 0)
				return
			}
			if evidence.ApprovedBy == "" || evidence.ApprovalCode == "" {
				http.Error(w, "se requiere aprobacion trazable (aprobado_por y codigo_aprobacion) para cambios de permisos", http.StatusBadRequest)
				registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusBadRequest, 0)
				return
			}

			r.Header.Set(permissionApprovalHeaderRequired, "1")
			r.Header.Set(permissionApprovalHeaderBy, evidence.ApprovedBy)
			r.Header.Set(permissionApprovalHeaderCode, evidence.ApprovalCode)
			if evidence.Reason != "" {
				r.Header.Set(permissionApprovalHeaderReason, evidence.Reason)
			}
		}

		ctx := context.WithValue(r.Context(), "adminRole", role)
		ctx = context.WithValue(ctx, "adminRoleEfectivo", effectiveRole)
		ctx = context.WithValue(ctx, "empresaID", empresaID)
		r = r.WithContext(ctx)

		w.Header().Set("X-Empresa-ID", strconv.FormatInt(empresaID, 10))
		r.Header.Set("X-Admin-Role", role)
		r.Header.Set("X-Admin-Role-Efectivo", effectiveRole)

		auditStart := time.Now()
		auditRW := &auditCaptureResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(auditRW, r)
		dbpkg.PerfLogf("[perf][authz] module=%s next empresa=%d path=%s dur=%s", module, empresaID, requestPath, time.Since(auditStart))
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func trimWithLimit(raw string, maxLen int) string {
	v := strings.TrimSpace(raw)
	if maxLen > 0 && len(v) > maxLen {
		return v[:maxLen]
	}
	return v
}

func extractJSONBodyMap(r *http.Request) (map[string]interface{}, error) {
	if r == nil || r.Body == nil {
		return nil, nil
	}
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		r.Body = io.NopCloser(bytes.NewReader(raw))
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewReader(raw))
	if len(raw) == 0 {
		return nil, nil
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, nil
	}
	return payload, nil
}

func extractStringField(payload map[string]interface{}, keys ...string) string {
	if payload == nil {
		return ""
	}
	for _, key := range keys {
		if value, ok := payload[key]; ok {
			switch typed := value.(type) {
			case string:
				if trimmed := strings.TrimSpace(typed); trimmed != "" {
					return trimmed
				}
			case float64:
				if typed > 0 {
					return strings.TrimSpace(strconv.FormatFloat(typed, 'f', -1, 64))
				}
			case int64:
				if typed > 0 {
					return strings.TrimSpace(strconv.FormatInt(typed, 10))
				}
			case int:
				if typed > 0 {
					return strings.TrimSpace(strconv.Itoa(typed))
				}
			}
		}
	}
	return ""
}

func extractPermissionApprovalEvidence(r *http.Request) (permissionApprovalEvidence, error) {
	evidence := permissionApprovalEvidence{
		ApprovedBy: trimWithLimit(firstNonEmpty(
			r.URL.Query().Get("aprobado_por"),
			r.URL.Query().Get("approved_by"),
			r.Header.Get(permissionApprovalHeaderBy),
		), 160),
		ApprovalCode: trimWithLimit(firstNonEmpty(
			r.URL.Query().Get("codigo_aprobacion"),
			r.URL.Query().Get("approval_code"),
			r.Header.Get(permissionApprovalHeaderCode),
		), 160),
		Reason: trimWithLimit(firstNonEmpty(
			r.URL.Query().Get("motivo_aprobacion"),
			r.URL.Query().Get("approval_reason"),
			r.Header.Get(permissionApprovalHeaderReason),
		), 320),
	}

	payload, err := extractJSONBodyMap(r)
	if err != nil {
		return permissionApprovalEvidence{}, err
	}
	if payload == nil {
		return evidence, nil
	}

	evidence.ApprovedBy = trimWithLimit(firstNonEmpty(
		evidence.ApprovedBy,
		extractStringField(payload, "aprobado_por", "approved_by"),
	), 160)
	evidence.ApprovalCode = trimWithLimit(firstNonEmpty(
		evidence.ApprovalCode,
		extractStringField(payload, "codigo_aprobacion", "approval_code"),
	), 160)
	evidence.Reason = trimWithLimit(firstNonEmpty(
		evidence.Reason,
		extractStringField(payload, "motivo_aprobacion", "approval_reason"),
	), 320)

	aprobacionPayload, _ := payload["aprobacion"].(map[string]interface{})
	evidence.ApprovedBy = trimWithLimit(firstNonEmpty(
		evidence.ApprovedBy,
		extractStringField(aprobacionPayload, "aprobado_por", "approved_by"),
	), 160)
	evidence.ApprovalCode = trimWithLimit(firstNonEmpty(
		evidence.ApprovalCode,
		extractStringField(aprobacionPayload, "codigo_aprobacion", "approval_code"),
	), 160)
	evidence.Reason = trimWithLimit(firstNonEmpty(
		evidence.Reason,
		extractStringField(aprobacionPayload, "motivo_aprobacion", "approval_reason"),
	), 320)

	return evidence, nil
}

func permissionChangeRequiresApproval(module string, r *http.Request, action string) bool {
	if module != permModuleSeguridad {
		return false
	}

	switch strings.ToUpper(strings.TrimSpace(action)) {
	case permActionCreate, permActionUpdate, permActionDelete, permActionApprove:
	default:
		return false
	}

	path := strings.ToLower(strings.TrimSpace(r.URL.Path))
	if path == "/api/empresa/roles_de_usuario" {
		return !strings.EqualFold(strings.TrimSpace(r.Method), http.MethodGet)
	}
	if path == "/api/empresa/permisos_empresa" {
		return !strings.EqualFold(strings.TrimSpace(r.Method), http.MethodGet)
	}
	if path != "/api/empresa/usuarios" {
		return false
	}

	queryAction := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	if queryAction == "reenviar_confirmacion" || queryAction == "activar" {
		return false
	}

	return true
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
	case "cerrar", "reabrir", "pagar_estacion", "activar_estacion", "pagar", "suspender", "suspender_venta", "reactivar", "reabrir_venta", "convertir_pedido", "convertir_documento_final":
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
	case "cerrar", "reabrir", "aprobar", "procesar_asientos", "procesar", "conciliar_bancaria_auto", "conciliar_bancos", "conciliar_bancaria_automatica", "aprobar_workflow", "aprobar_reverso", "aprobar_refinanciacion", "rechazar_workflow", "rechazar_reverso", "rechazar_refinanciacion":
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
	if action == "aprobar" || action == "cerrar" || action == "emitir" || action == "emitir_orden" || action == "recepcionar" || action == "recepcionar_compra" || action == "recepcionar_parcial_compra" || action == "contabilizar" || action == "contabilizar_compra" || action == "solicitar_aprobacion" || action == "aprobar_compra" || action == "rechazar_compra" || action == "validar_documentos" {
		return permActionApprove
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func resolveFacturacionPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	if action == "activar" || action == "desactivar" {
		return permActionUpdate
	}
	if (action == "procesar_reintentos" || action == "reconciliar_estados" || action == "firmar_xml_real" || action == "enviar_documento_real" || action == "reconexion_dian" || action == "consultar_acuse_real") && (r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch) {
		return permActionApprove
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
	case "solicitar_aprobacion", "iniciar_aprobacion":
		return permActionUpdate
	case "versionar":
		return permActionApprove
	case "restaurar", "restore", "rollback_backup":
		return permActionApprove
	case "depurar_fecha", "purgar_fecha", "eliminar_hasta_fecha", "depurar_hasta_fecha":
		return permActionApprove
	case "sync_manual", "rotar_credencial", "rotar_credenciales":
		return permActionApprove
	case "aprobar", "rechazar", "vincular_nomina", "enlazar_nomina":
		return permActionApprove
	case "reenviar_confirmacion":
		return permActionApprove
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func normalizePermissionRole(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "super_administrador", "superadmin", "super":
		return "super_administrador"
	case "administrador_total", "admin_total", "admin_full", "full_admin":
		return "administrador_total"
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
	if role == "administrador_total" {
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

func roleAllowsModuleActionWithOverrides(dbSuper *sql.DB, role, module, action string) bool {
	normalizedRole := normalizePermissionRole(role)
	normalizedModule := strings.ToLower(strings.TrimSpace(module))
	normalizedAction := strings.ToUpper(strings.TrimSpace(action))

	allowed := roleAllowsModuleAction(normalizedRole, normalizedModule, normalizedAction)
	if dbSuper == nil || normalizedRole == "" || normalizedRole == "sin_rol" {
		return allowed
	}

	found, permitido, err := dbpkg.LookupRolPermisoModuloByRoleName(dbSuper, normalizedRole, normalizedModule, normalizedAction)
	if err != nil {
		if isPermissionMissingTableError(err) {
			return allowed
		}
		log.Printf("[authz] modulo override lookup role=%s modulo=%s accion=%s error: %v", normalizedRole, normalizedModule, normalizedAction, err)
		return allowed
	}
	if found {
		return permitido
	}

	return allowed
}

func empresaAllowsModuleActionWithOverrides(dbSuper *sql.DB, empresaID int64, module, action string) bool {
	if dbSuper == nil || empresaID <= 0 {
		return true
	}
	normalizedModule := strings.ToLower(strings.TrimSpace(module))
	normalizedAction := strings.ToUpper(strings.TrimSpace(action))
	if normalizedModule == "" || normalizedAction == "" {
		return true
	}
	found, permitido, err := dbpkg.LookupEmpresaPermisoModulo(dbSuper, empresaID, normalizedModule, normalizedAction)
	if err != nil {
		if isPermissionMissingTableError(err) {
			return true
		}
		log.Printf("[authz] empresa permiso lookup empresa_id=%d modulo=%s accion=%s error: %v", empresaID, normalizedModule, normalizedAction, err)
		return true
	}
	if found {
		return permitido
	}
	return true
}

func loadPermissionOverridesByRoleName(dbSuper *sql.DB, role string) (map[string]bool, map[string]bool, error) {
	moduleOverrides := map[string]bool{}
	pageOverrides := map[string]bool{}
	if dbSuper == nil {
		return moduleOverrides, pageOverrides, nil
	}

	normalizedRole := normalizePermissionRole(role)
	if normalizedRole == "" || normalizedRole == "sin_rol" {
		return moduleOverrides, pageOverrides, nil
	}

	empresaPermissionCacheMu.Lock()
	if cached, ok := rolePermissionOverrideCache[normalizedRole]; ok && time.Since(cached.LoadedAt) < permissionOverrideCacheTTL {
		empresaPermissionCacheMu.Unlock()
		return clonePermissionBoolMap(cached.ModuleOverrides), clonePermissionBoolMap(cached.PageOverrides), nil
	}
	empresaPermissionCacheMu.Unlock()

	rolID, err := dbpkg.ResolveRolDeUsuarioIDByNombre(dbSuper, normalizedRole)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || isPermissionMissingTableError(err) {
			return moduleOverrides, pageOverrides, nil
		}
		return nil, nil, err
	}

	var (
		modulos    []dbpkg.RolPermisoModulo
		paginas    []dbpkg.RolPermisoPagina
		modulosErr error
		paginasErr error
	)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		modulos, modulosErr = dbpkg.ListRolPermisosModuloByRolID(dbSuper, rolID)
	}()
	go func() {
		defer wg.Done()
		paginas, paginasErr = dbpkg.ListRolPermisosPaginaByRolID(dbSuper, rolID)
	}()
	wg.Wait()
	if modulosErr != nil {
		if isPermissionMissingTableError(modulosErr) {
			return moduleOverrides, pageOverrides, nil
		}
		return nil, nil, modulosErr
	}
	for _, item := range modulos {
		moduleOverrides[permissionModuleActionKey(item.Modulo, item.Accion)] = item.Permitido
	}

	if paginasErr != nil {
		if isPermissionMissingTableError(paginasErr) {
			return moduleOverrides, pageOverrides, nil
		}
		return nil, nil, paginasErr
	}
	for _, item := range paginas {
		pageOverrides[strings.TrimSpace(item.PaginaClave)] = item.Permitido
	}

	empresaPermissionCacheMu.Lock()
	rolePermissionOverrideCache[normalizedRole] = permissionRoleOverrideCacheEntry{
		ModuleOverrides: clonePermissionBoolMap(moduleOverrides),
		PageOverrides:   clonePermissionBoolMap(pageOverrides),
		LoadedAt:        time.Now(),
	}
	empresaPermissionCacheMu.Unlock()

	return moduleOverrides, pageOverrides, nil
}

func loadEmpresaPermissionOverrides(dbSuper *sql.DB, empresaID int64) (map[string]bool, map[string]bool, *empresaPermisosFinosCtx) {
	moduleOverrides := map[string]bool{}
	pageOverrides := map[string]bool{}
	ctx := &empresaPermisosFinosCtx{}
	if dbSuper == nil || empresaID <= 0 {
		return moduleOverrides, pageOverrides, ctx
	}

	empresaPermissionCacheMu.Lock()
	if cached, ok := empresaPermissionOverrideCache[empresaID]; ok && time.Since(cached.LoadedAt) < permissionOverrideCacheTTL {
		empresaPermissionCacheMu.Unlock()
		return clonePermissionBoolMap(cached.ModuleOverrides), clonePermissionBoolMap(cached.PageOverrides), cloneEmpresaPermisosFinosCtx(cached.Ctx)
	}
	empresaPermissionCacheMu.Unlock()

	var (
		modulos    []dbpkg.EmpresaPermisoModulo
		paginas    []dbpkg.EmpresaPermisoPagina
		modulosErr error
		paginasErr error
	)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		modulos, modulosErr = dbpkg.ListEmpresaPermisosModuloByEmpresaID(dbSuper, empresaID)
	}()
	go func() {
		defer wg.Done()
		paginas, paginasErr = dbpkg.ListEmpresaPermisosPaginaByEmpresaID(dbSuper, empresaID)
	}()
	wg.Wait()
	if modulosErr != nil {
		if !isPermissionMissingTableError(modulosErr) {
			log.Printf("[authz] load empresa modulo overrides empresa_id=%d error: %v", empresaID, modulosErr)
		}
		modulos = []dbpkg.EmpresaPermisoModulo{}
	}
	for _, item := range modulos {
		moduleOverrides[permissionModuleActionKey(item.Modulo, item.Accion)] = item.Permitido
	}

	if paginasErr != nil {
		if !isPermissionMissingTableError(paginasErr) {
			log.Printf("[authz] load empresa page overrides empresa_id=%d error: %v", empresaID, paginasErr)
		}
		paginas = []dbpkg.EmpresaPermisoPagina{}
	}
	for _, item := range paginas {
		pageOverrides[strings.TrimSpace(item.PaginaClave)] = item.Permitido
	}

	ctx.ReglasModulo = len(moduleOverrides)
	ctx.ReglasPagina = len(pageOverrides)
	ctx.Activo = ctx.ReglasModulo > 0 || ctx.ReglasPagina > 0

	empresaPermissionCacheMu.Lock()
	empresaPermissionOverrideCache[empresaID] = empresaPermissionOverrideCacheEntry{
		ModuleOverrides: clonePermissionBoolMap(moduleOverrides),
		PageOverrides:   clonePermissionBoolMap(pageOverrides),
		Ctx:             cloneEmpresaPermisosFinosCtx(ctx),
		LoadedAt:        time.Now(),
	}
	empresaPermissionCacheMu.Unlock()
	return moduleOverrides, pageOverrides, ctx
}

func clonePermissionBoolMap(input map[string]bool) map[string]bool {
	if len(input) == 0 {
		return map[string]bool{}
	}
	out := make(map[string]bool, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func cloneEmpresaPermisosFinosCtx(input *empresaPermisosFinosCtx) *empresaPermisosFinosCtx {
	if input == nil {
		return &empresaPermisosFinosCtx{}
	}
	out := *input
	return &out
}

func clonePermissionModuleRows(input []permissionModuleMatrixRow) []permissionModuleMatrixRow {
	if len(input) == 0 {
		return []permissionModuleMatrixRow{}
	}
	out := make([]permissionModuleMatrixRow, 0, len(input))
	for _, row := range input {
		copied := row
		copied.Acciones = clonePermissionBoolMap(row.Acciones)
		out = append(out, copied)
	}
	return out
}

func permissionModuleActionKey(modulo, accion string) string {
	return strings.ToLower(strings.TrimSpace(modulo)) + "|" + strings.ToUpper(strings.TrimSpace(accion))
}

func setPermissionActionOnModuleRow(row *permissionModuleMatrixRow, action string, permitido bool) {
	normalizedAction := strings.ToUpper(strings.TrimSpace(action))
	switch normalizedAction {
	case permActionRead:
		row.Read = permitido
	case permActionCreate:
		row.Create = permitido
	case permActionUpdate:
		row.Update = permitido
	case permActionDelete:
		row.Delete = permitido
	case permActionApprove:
		row.Approve = permitido
	}
	if row.Acciones == nil {
		row.Acciones = map[string]bool{}
	}
	row.Acciones[normalizedAction] = permitido
}

func isPermissionMissingTableError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(msg, "no such table") || strings.Contains(msg, "does not exist")
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

func buildPermissionModuleMatrixForRoleDynamic(dbSuper *sql.DB, role string) []permissionModuleMatrixRow {
	normalizedRole := normalizePermissionRole(role)
	empresaPermissionCacheMu.Lock()
	if cached, ok := rolePermissionModuleMatrixCache[normalizedRole]; ok && time.Since(cached.LoadedAt) < permissionOverrideCacheTTL {
		empresaPermissionCacheMu.Unlock()
		return clonePermissionModuleRows(cached.Rows)
	}
	empresaPermissionCacheMu.Unlock()

	rows := buildPermissionModuleMatrixForRole(role)
	moduleOverrides, _, err := loadPermissionOverridesByRoleName(dbSuper, role)
	if err != nil {
		log.Printf("[authz] load permission overrides role=%s error: %v", role, err)
		return rows
	}
	if len(moduleOverrides) == 0 {
		return rows
	}

	for idx := range rows {
		row := &rows[idx]
		for _, action := range permissionActionsCatalogOrdered {
			if permitido, ok := moduleOverrides[permissionModuleActionKey(row.Modulo, action)]; ok {
				setPermissionActionOnModuleRow(row, action, permitido)
			}
		}
	}

	empresaPermissionCacheMu.Lock()
	rolePermissionModuleMatrixCache[normalizedRole] = permissionRoleModuleMatrixCacheEntry{
		Rows:     clonePermissionModuleRows(rows),
		LoadedAt: time.Now(),
	}
	empresaPermissionCacheMu.Unlock()
	return rows
}

func buildEmpresaPermisosDefaultModuleRows() []permissionModuleMatrixRow {
	out := make([]permissionModuleMatrixRow, 0, len(permissionModulesCatalogOrdered))
	for _, modulo := range permissionModulesCatalogOrdered {
		out = append(out, permissionModuleMatrixRow{
			Modulo:  modulo,
			Read:    true,
			Create:  true,
			Update:  true,
			Delete:  true,
			Approve: true,
			Acciones: map[string]bool{
				permActionRead:    true,
				permActionCreate:  true,
				permActionUpdate:  true,
				permActionDelete:  true,
				permActionApprove: true,
			},
		})
	}
	return out
}

func buildPermissionPagesMapForRoleDynamic(dbSuper *sql.DB, role string, modulos []permissionModuleMatrixRow) map[string]bool {
	_, pageOverrides, err := loadPermissionOverridesByRoleName(dbSuper, role)
	if err != nil {
		log.Printf("[authz] load page overrides role=%s error: %v", role, err)
		pageOverrides = map[string]bool{}
	}
	return buildPermissionPagesMapFromModuleRows(modulos, pageOverrides)
}

func buildPermissionPagesCatalogForRoleDynamic(dbSuper *sql.DB, role string, modulos []permissionModuleMatrixRow) []permissionPageAccessRow {
	_, pageOverrides, err := loadPermissionOverridesByRoleName(dbSuper, role)
	if err != nil {
		log.Printf("[authz] load page catalog overrides role=%s error: %v", role, err)
		pageOverrides = map[string]bool{}
	}
	return buildPermissionPagesCatalogFromModuleRows(modulos, pageOverrides)
}

func buildPermissionPagesMapFromModuleRows(modulos []permissionModuleMatrixRow, pageOverrides map[string]bool) map[string]bool {
	rows := buildPermissionPagesCatalogFromModuleRows(modulos, pageOverrides)
	out := make(map[string]bool, len(rows))
	for _, row := range rows {
		out[row.PaginaClave] = row.Permitido
	}
	return out
}

func buildPermissionPagesCatalogFromModuleRows(modulos []permissionModuleMatrixRow, pageOverrides map[string]bool) []permissionPageAccessRow {
	moduleRows := make(map[string]permissionModuleMatrixRow, len(modulos))
	for _, row := range modulos {
		moduleRows[strings.ToLower(strings.TrimSpace(row.Modulo))] = row
	}

	out := make([]permissionPageAccessRow, 0, len(permissionPagesCatalogOrdered))
	for _, rule := range permissionPagesCatalogOrdered {
		permitido := true
		if !rule.AlwaysVisible {
			permitido = false
			if moduleRow, ok := moduleRows[strings.ToLower(strings.TrimSpace(rule.Modulo))]; ok {
				permitido = moduleRow.Acciones[strings.ToUpper(strings.TrimSpace(rule.Accion))]
			}
		}
		if override, ok := pageOverrides[rule.PaginaClave]; ok {
			permitido = override
		}

		titulo := sanitizeLegacyPermissionVisibleText(rule.Titulo)
		if titulo == "" {
			titulo = sanitizeLegacyPermissionVisibleText(rule.PaginaClave)
		}
		grupo := sanitizeLegacyPermissionVisibleText(rule.Grupo)
		if grupo == "" {
			grupo = "Otras"
		}
		out = append(out, permissionPageAccessRow{
			PaginaClave:   rule.PaginaClave,
			Modulo:        sanitizeLegacyPermissionVisibleText(rule.Modulo),
			Accion:        sanitizeLegacyPermissionVisibleText(rule.Accion),
			Permitido:     permitido,
			AlwaysVisible: rule.AlwaysVisible,
			Titulo:        titulo,
			Grupo:         grupo,
		})
	}

	return out
}

func parseLicenciaModulosCSV(raw string) (map[string]bool, []string) {
	allowed := map[string]bool{}
	ordered := make([]string, 0)
	for _, chunk := range strings.Split(raw, ",") {
		modulo := strings.ToLower(strings.TrimSpace(chunk))
		if modulo == "" || !isPermissionModuleKnown(modulo) {
			continue
		}
		if allowed[modulo] {
			continue
		}
		allowed[modulo] = true
		ordered = append(ordered, modulo)
	}
	return allowed, ordered
}

func isPermissionModuleKnown(modulo string) bool {
	target := strings.ToLower(strings.TrimSpace(modulo))
	if target == "" {
		return false
	}
	for _, known := range permissionModulesCatalogOrdered {
		if target == strings.ToLower(strings.TrimSpace(known)) {
			return true
		}
	}
	return false
}

func isModuloPermitidoByLicencia(modulo string, allowed map[string]bool) bool {
	if len(allowed) == 0 {
		return true
	}
	key := strings.ToLower(strings.TrimSpace(modulo))
	if key == "" {
		return false
	}
	return allowed[key]
}

func applyLicenciaRestriccionesToModuleRows(rows []permissionModuleMatrixRow, allowed map[string]bool) []permissionModuleMatrixRow {
	if len(allowed) == 0 {
		return rows
	}
	out := make([]permissionModuleMatrixRow, 0, len(rows))
	for _, row := range rows {
		next := row
		next.Acciones = map[string]bool{}
		for _, action := range permissionActionsCatalogOrdered {
			next.Acciones[action] = row.Acciones[action]
		}
		if !isModuloPermitidoByLicencia(next.Modulo, allowed) {
			setPermissionActionOnModuleRow(&next, permActionRead, false)
			setPermissionActionOnModuleRow(&next, permActionCreate, false)
			setPermissionActionOnModuleRow(&next, permActionUpdate, false)
			setPermissionActionOnModuleRow(&next, permActionDelete, false)
			setPermissionActionOnModuleRow(&next, permActionApprove, false)
		}
		out = append(out, next)
	}
	return out
}

func applyEmpresaRestriccionesToModuleRows(rows []permissionModuleMatrixRow, overrides map[string]bool) []permissionModuleMatrixRow {
	if len(overrides) == 0 {
		return rows
	}
	out := make([]permissionModuleMatrixRow, 0, len(rows))
	for _, row := range rows {
		next := row
		next.Acciones = map[string]bool{}
		for _, action := range permissionActionsCatalogOrdered {
			next.Acciones[action] = row.Acciones[action]
		}
		for _, action := range permissionActionsCatalogOrdered {
			if permitido, ok := overrides[permissionModuleActionKey(next.Modulo, action)]; ok {
				setPermissionActionOnModuleRow(&next, action, permitido && row.Acciones[action])
			}
		}
		out = append(out, next)
	}
	return out
}

func applyEmpresaPageRestrictionsToMap(paginas map[string]bool, overrides map[string]bool) map[string]bool {
	if len(overrides) == 0 {
		return paginas
	}
	out := make(map[string]bool, len(paginas))
	for k, v := range paginas {
		out[k] = v
	}
	for key, permitido := range overrides {
		clean := strings.TrimSpace(key)
		if clean == "" {
			continue
		}
		out[clean] = permitido && out[clean]
	}
	return out
}

func resolvePermissionPageKeyForRequest(r *http.Request) string {
	if r == nil || r.URL == nil {
		return ""
	}
	path := strings.ToLower(strings.TrimSpace(r.URL.Path))
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	switch {
	case strings.HasPrefix(path, "/api/empresa/crm/"):
		return "linkCRMComercial"
	case path == "/api/empresa/clientes":
		return "linkClientes"
	case strings.HasPrefix(path, "/api/empresa/chat_tareas"):
		return "linkChatTareas"
	case strings.HasPrefix(path, "/api/empresa/chat_con_inteligencia_artificial"):
		return "linkChatIA"
	case strings.HasPrefix(path, "/api/empresa/facturacion_electronica"):
		if action == "emitir" || !strings.EqualFold(strings.TrimSpace(r.Method), http.MethodGet) {
			return "linkFacturacionElectronica"
		}
		return "linkFacturasElectronicas"
	case strings.HasPrefix(path, "/api/empresa/finanzas/") || path == "/api/empresa/corte_caja":
		return "linkFinanzas"
	case strings.HasPrefix(path, "/api/empresa/creditos") ||
		strings.HasPrefix(path, "/api/empresa/cuentas_por_cobrar") ||
		strings.HasPrefix(path, "/api/empresa/cuentas_por_pagar"):
		return "linkCreditos"
	case path == "/api/empresa/propinas":
		return "linkPropinas"
	case path == "/api/empresa/comisiones":
		return "linkComisiones"
	case path == "/api/empresa/codigos_de_descuento":
		return "linkCodigosDescuento"
	case path == "/api/empresa/venta_publica":
		return "linkVentaPublica"
	case path == "/api/empresa/carritos_compra":
		if strings.Contains(action, "estacion") {
			return "linkEstaciones"
		}
		return "linkCarritoCompras"
	case strings.HasPrefix(path, "/api/empresa/estaciones") ||
		strings.HasPrefix(path, "/api/empresa/estacion_") ||
		strings.HasPrefix(path, "/api/empresa/ventas_estacion"):
		return "linkEstaciones"
	case path == "/api/empresa/reservas_hotel":
		return "linkReservasHotel"
	case path == "/api/empresa/tarifas_por_minutos":
		return "linkTarifasPorMinutos"
	case path == "/api/empresa/tarifas_por_dia":
		return "linkTarifasPorDia"
	case path == "/api/empresa/nomina_sueldos":
		return "linkNominaSueldos"
	case path == "/api/empresa/horarios_trabajadores":
		return "linkHorariosTrabajadores"
	case path == "/api/empresa/asistencia_empleados":
		return "linkAsistenciaEmpleados"
	case path == "/api/empresa/vehiculos_registro":
		return "linkVehiculosRegistro"
	case path == "/api/empresa/gimnasio":
		return "linkGimnasio"
	case strings.HasPrefix(path, "/api/empresa/reportes"):
		if strings.Contains(path, "/ia") {
			return "linkReportesIAChat"
		}
		return "linkReportes"
	}
	return ""
}

func getEmpresaPermissionSnapshot(dbEmp, dbSuper *sql.DB, adminEmail string, empresaID int64) (empresaPermissionSnapshot, error) {
	startedAt := time.Now()
	defer func() {
		dbpkg.PerfLogf("[perf][authz] getEmpresaPermissionSnapshot empresa=%d email=%s dur=%s", empresaID, strings.ToLower(strings.TrimSpace(adminEmail)), time.Since(startedAt))
	}()
	cacheKey := strings.ToLower(strings.TrimSpace(adminEmail)) + "|" + strconv.FormatInt(empresaID, 10)
	if strings.TrimSpace(adminEmail) == "" || empresaID <= 0 {
		return empresaPermissionSnapshot{}, sql.ErrNoRows
	}
	var snapshotResult empresaPermissionSnapshot
	var snapshotErr error

	empresaPermissionCacheMu.Lock()
	if cached, ok := empresaPermissionCache[cacheKey]; ok && time.Since(cached.LoadedAt) < empresaPermissionCacheTTL {
		empresaPermissionCacheMu.Unlock()
		return cached, nil
	}
	if inflight, ok := empresaPermissionInflight[cacheKey]; ok {
		empresaPermissionCacheMu.Unlock()
		<-inflight.done
		return inflight.snapshot, inflight.err
	}
	inflight := &empresaPermissionSnapshotInflight{done: make(chan struct{})}
	empresaPermissionInflight[cacheKey] = inflight
	empresaPermissionCacheMu.Unlock()
	defer func() {
		empresaPermissionCacheMu.Lock()
		delete(empresaPermissionInflight, cacheKey)
		inflight.snapshot = snapshotResult
		inflight.err = snapshotErr
		close(inflight.done)
		empresaPermissionCacheMu.Unlock()
	}()
	stepStarted := time.Now()
	admin, err := dbpkg.GetAdminByEmail(dbSuper, adminEmail)
	if err != nil {
		snapshotErr = err
		return empresaPermissionSnapshot{}, err
	}
	dbpkg.PerfLogf("[perf][authz] snapshot empresa=%d email=%s step=admin dur=%s", empresaID, strings.ToLower(strings.TrimSpace(adminEmail)), time.Since(stepStarted))
	role := normalizePermissionRole(admin.Role)

	var (
		canAccess              bool
		canAccessErr           error
		licenciaPolicy         *dbpkg.LicenciaPermisoPolicy
		licenciaErr            error
		moduleRows             []permissionModuleMatrixRow
		empresaModuleOverrides map[string]bool
		empresaPageOverrides   map[string]bool
	)

	var snapshotWG sync.WaitGroup
	snapshotWG.Add(4)

	go func() {
		defer snapshotWG.Done()
		step := time.Now()
		canAccess, canAccessErr = dbpkg.CanAdminAccessEmpresaIA(dbEmp, dbSuper, adminEmail, empresaID)
		dbpkg.PerfLogf("[perf][authz] snapshot empresa=%d email=%s step=access dur=%s", empresaID, strings.ToLower(strings.TrimSpace(adminEmail)), time.Since(step))
	}()

	go func() {
		defer snapshotWG.Done()
		step := time.Now()
		licenciaPolicy, licenciaErr = dbpkg.GetLicenciaPermisoPolicyByEmpresa(dbSuper, empresaID)
		dbpkg.PerfLogf("[perf][authz] snapshot empresa=%d email=%s step=licencia dur=%s", empresaID, strings.ToLower(strings.TrimSpace(adminEmail)), time.Since(step))
	}()

	go func() {
		defer snapshotWG.Done()
		step := time.Now()
		moduleRows = buildPermissionModuleMatrixForRoleDynamic(dbSuper, role)
		dbpkg.PerfLogf("[perf][authz] snapshot empresa=%d email=%s step=module_rows dur=%s", empresaID, strings.ToLower(strings.TrimSpace(adminEmail)), time.Since(step))
	}()

	go func() {
		defer snapshotWG.Done()
		step := time.Now()
		empresaModuleOverrides, empresaPageOverrides, _ = loadEmpresaPermissionOverrides(dbSuper, empresaID)
		dbpkg.PerfLogf("[perf][authz] snapshot empresa=%d email=%s step=empresa_overrides dur=%s", empresaID, strings.ToLower(strings.TrimSpace(adminEmail)), time.Since(step))
	}()

	snapshotWG.Wait()
	if canAccessErr != nil {
		snapshotErr = canAccessErr
		return empresaPermissionSnapshot{}, canAccessErr
	}
	if licenciaErr != nil {
		snapshotErr = licenciaErr
		return empresaPermissionSnapshot{}, licenciaErr
	}

	allowedModules, _ := parseLicenciaModulosCSV("")
	if licenciaPolicy != nil {
		allowedModules, _ = parseLicenciaModulosCSV(licenciaPolicy.ModulosHabilitados)
	}
	effectiveRole := resolveEffectiveRoleByLicencia(role, licenciaPolicy)
	if effectiveRole != role {
		stepStarted = time.Now()
		moduleRows = buildPermissionModuleMatrixForRoleDynamic(dbSuper, effectiveRole)
		dbpkg.PerfLogf("[perf][authz] snapshot empresa=%d email=%s step=module_rows_effective dur=%s", empresaID, strings.ToLower(strings.TrimSpace(adminEmail)), time.Since(stepStarted))
	}
	moduleRows = applyLicenciaRestriccionesToModuleRows(moduleRows, allowedModules)
	moduleRows = applyEmpresaRestriccionesToModuleRows(moduleRows, empresaModuleOverrides)
	stepStarted = time.Now()
	allowedPages := buildPermissionPagesMapForRoleDynamic(dbSuper, effectiveRole, moduleRows)
	dbpkg.PerfLogf("[perf][authz] snapshot empresa=%d email=%s step=allowed_pages dur=%s", empresaID, strings.ToLower(strings.TrimSpace(adminEmail)), time.Since(stepStarted))
	allowedPages = applyEmpresaPageRestrictionsToMap(allowedPages, empresaPageOverrides)

	roleModuleActions := map[string]bool{}
	for _, row := range moduleRows {
		for _, permissionAction := range permissionActionsCatalogOrdered {
			roleModuleActions[permissionModuleActionKey(row.Modulo, permissionAction)] = row.Acciones[permissionAction]
		}
	}

	snapshot := empresaPermissionSnapshot{
		AdminRole:         role,
		EffectiveRole:     effectiveRole,
		CanAccess:         canAccess,
		AllowedModules:    allowedModules,
		RoleModuleActions: roleModuleActions,
		AllowedPages:      allowedPages,
		LoadedAt:          time.Now(),
	}

	empresaPermissionCacheMu.Lock()
	empresaPermissionCache[cacheKey] = snapshot
	empresaPermissionCacheMu.Unlock()
	snapshotResult = snapshot
	return snapshot, nil
}

func resolveEffectiveRoleByLicencia(role string, licenciaPolicy *dbpkg.LicenciaPermisoPolicy) string {
	resolved := normalizePermissionRole(role)
	if licenciaPolicy == nil || !licenciaPolicy.SuperRolHabilitado {
		return resolved
	}
	if resolved == "supervisor_sucursal" {
		return "admin_empresa"
	}
	return resolved
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

// PermissionModuleDisplayNameMap devuelve etiquetas de negocio por clave de mÃƒÂ³dulo (API super: permisos por rol).
func PermissionModuleDisplayNameMap() map[string]string {
	out := make(map[string]string, len(permissionModulesCatalogOrdered))
	for _, m := range permissionModulesCatalogOrdered {
		if lab, ok := permissionModuleDisplayNames[m]; ok && strings.TrimSpace(lab) != "" {
			out[m] = sanitizeLegacyPermissionVisibleText(lab)
		} else {
			out[m] = sanitizeLegacyPermissionVisibleText(m)
		}
	}
	return out
}

// PermissionActionDisplayNameMap devuelve etiquetas por letra de acciÃƒÂ³n (R/C/U/D/A).
func PermissionActionDisplayNameMap() map[string]string {
	out := make(map[string]string, len(permissionActionsCatalogOrdered))
	for _, a := range permissionActionsCatalogOrdered {
		if lab, ok := permissionActionDisplayNames[a]; ok && strings.TrimSpace(lab) != "" {
			out[a] = sanitizeLegacyPermissionVisibleText(lab)
		} else {
			out[a] = sanitizeLegacyPermissionVisibleText(a)
		}
	}
	return out
}
