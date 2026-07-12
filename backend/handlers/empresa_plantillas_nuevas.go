package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

type empresaVerticalNuevoPermiso struct {
	Modulo string
	Page   string
	Titulo string
}

type empresaVerticalNuevoCatalogoItem struct {
	ID                   string                                         `json:"id"`
	Modulo               string                                         `json:"module"`
	Page                 string                                         `json:"page"`
	Titulo               string                                         `json:"title"`
	TituloFull           string                                         `json:"full_title"`
	Resumen              string                                         `json:"summary"`
	Secciones            []string                                       `json:"sections"`
	IntegrationStatus    string                                         `json:"integration_status"`
	OperationalVisible   bool                                           `json:"operational_visible"`
	CoreModules          []string                                       `json:"core_modules"`
	DuplicatesCore       []string                                       `json:"duplicates_core"`
	ProduccionMasiva     bool                                           `json:"produccion_masiva"`
	PrioridadProduccion  int                                            `json:"prioridad_produccion,omitempty"`
	DecisionPreconfig    string                                         `json:"decision_preconfig,omitempty"`
	IntegracionPreconfig *dbpkg.TipoEmpresaPreconfigIntegracionVertical `json:"integracion_preconfig,omitempty"`
	Plantilla            dbpkg.EmpresaModuloColombiaPlantilla           `json:"plantilla"`
}

func empresaPlantillasNuevosPermisos() []empresaVerticalNuevoPermiso {
	catalog := dbpkg.NuevasPlantillasTipoEmpresaCatalog()
	out := make([]empresaVerticalNuevoPermiso, 0, len(catalog))
	for _, item := range catalog {
		out = append(out, empresaVerticalNuevoPermiso{
			Modulo: strings.ToLower(strings.TrimSpace(item.Modulo)),
			Page:   nuevoVerticalPageKey(item.Modulo),
			Titulo: strings.TrimSpace(item.Nombre),
		})
	}
	return out
}

func EmpresaPlantillasNuevosCatalogoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		items := buildEmpresaPlantillasNuevosCatalogo()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"total": len(items),
			"items": items,
		})
	}
}

func SuperPlantillasNuevosCatalogoHandler(dbSuper ...*sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var superDB *sql.DB
		if len(dbSuper) > 0 {
			superDB = dbSuper[0]
		}
		if _, ok := paginaPrincipalRequireSuperAdmin(w, r, superDB); !ok {
			return
		}
		if r.Method == http.MethodPost {
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action != "asegurar_v1_licencias" && action != "asegurar_produccion_masiva" && action != "asegurar_20_licencias" {
				http.Error(w, "accion no permitida", http.StatusBadRequest)
				return
			}
			if superDB == nil {
				http.Error(w, "db super no disponible", http.StatusInternalServerError)
				return
			}
			tipos, licencias, err := dbpkg.EnsureNuevasPlantillasProduccionMasivaLicencias(superDB, "super.plantillas_20")
			if err != nil {
				http.Error(w, "no se pudieron asegurar plantillas de produccion: "+err.Error(), http.StatusInternalServerError)
				return
			}
			items := buildEmpresaPlantillasNuevosCatalogo()
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":                   true,
				"tipos_asegurados":     tipos,
				"licencias_aseguradas": licencias,
				"total":                len(items),
				"items":                items,
			})
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		items := buildEmpresaPlantillasNuevosCatalogo()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"total": len(items),
			"items": items,
		})
	}
}

func PublicPlantillasNuevosCatalogoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		items := buildEmpresaPlantillasNuevosCatalogo()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":    true,
			"total": len(items),
			"items": items,
		})
	}
}

func buildEmpresaPlantillasNuevosCatalogo() []empresaVerticalNuevoCatalogoItem {
	catalog := dbpkg.NuevasPlantillasTipoEmpresaCatalog()
	out := make([]empresaVerticalNuevoCatalogoItem, 0, len(catalog))
	for _, item := range catalog {
		modulo := strings.ToLower(strings.TrimSpace(item.Modulo))
		if modulo == "" {
			continue
		}
		page := nuevoVerticalPageKey(modulo)
		plantilla := dbpkg.GetEmpresaModuloColombiaPlantilla(modulo)
		integracion := dbpkg.BuildTipoEmpresaPreconfigIntegracionVertical(modulo)
		out = append(out, empresaVerticalNuevoCatalogoItem{
			ID:                   page,
			Modulo:               modulo,
			Page:                 page,
			Titulo:               strings.TrimSpace(item.Nombre),
			TituloFull:           strings.TrimSpace(item.Nombre),
			Resumen:              strings.TrimSpace(item.Observaciones),
			Secciones:            append([]string{}, plantilla.SeccionesFlujo...),
			IntegrationStatus:    "plantilla_integrada_nucleo",
			OperationalVisible:   true,
			CoreModules:          []string{"clientes", "inventario", "ventas", "pagos", "facturacion", "reportes", "seguridad"},
			DuplicatesCore:       []string{},
			ProduccionMasiva:     integracion != nil && integracion.ProduccionMasiva,
			PrioridadProduccion:  dbpkg.NuevoVerticalProduccionMasivaRank(modulo),
			DecisionPreconfig:    integrationDecision(integracion),
			IntegracionPreconfig: integracion,
			Plantilla:            plantilla,
		})
	}
	return out
}

func integrationDecision(integracion *dbpkg.TipoEmpresaPreconfigIntegracionVertical) string {
	if integracion == nil {
		return ""
	}
	return strings.TrimSpace(integracion.Decision)
}

var empresaPlantillasNuevosModuloSet = func() map[string]bool {
	permisos := empresaPlantillasNuevosPermisos()
	out := make(map[string]bool, len(permisos))
	for _, item := range permisos {
		out[item.Modulo] = true
	}
	return out
}()

var empresaPlantillasNuevosPageByAPIPath = func() map[string]string {
	permisos := empresaPlantillasNuevosPermisos()
	out := make(map[string]string, len(permisos))
	for _, item := range permisos {
		out["/api/empresa/"+item.Modulo] = item.Page
	}
	return out
}()

func init() {
	modulos := NuevasPlantillasEmpresaModules()
	permissionPagesCatalogOrdered = append(permissionPagesCatalogOrdered, permissionPageRule{
		PaginaClave: "linkNuevasPlantillas",
		AnyModules:  modulos,
		Accion:      permActionCreate,
		Titulo:      "20 nuevas plantillas empresariales",
		Grupo:       "Plantillas de negocio",
	})
	for _, item := range empresaPlantillasNuevosPermisos() {
		permissionModulesCatalogOrdered = append(permissionModulesCatalogOrdered, item.Modulo)
		permissionModuleDisplayNames[item.Modulo] = item.Titulo
		permissionPagesCatalogOrdered = append(permissionPagesCatalogOrdered, permissionPageRule{
			PaginaClave: item.Page,
			Modulo:      item.Modulo,
			Accion:      permActionCreate,
			Titulo:      item.Titulo,
			Grupo:       "Plantillas de negocio",
		})
	}
}

func NuevasPlantillasEmpresaModules() []string {
	permisos := empresaPlantillasNuevosPermisos()
	out := make([]string, 0, len(permisos))
	for _, item := range permisos {
		out = append(out, item.Modulo)
	}
	return out
}

func isPermModuleNuevoVertical(module string) bool {
	return empresaPlantillasNuevosModuloSet[strings.ToLower(strings.TrimSpace(module))]
}

func permissionPageForNuevoVerticalAPIPath(path string) (string, bool) {
	page, ok := empresaPlantillasNuevosPageByAPIPath[strings.ToLower(strings.TrimSpace(path))]
	return page, ok
}

func WithEmpresaModuloVerticalPermissions(dbEmp, dbSuper *sql.DB, modulo string, next http.HandlerFunc) http.HandlerFunc {
	clean := strings.ToLower(strings.TrimSpace(modulo))
	if !isPermModuleNuevoVertical(clean) {
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "modulo de plantilla no soportado", http.StatusBadRequest)
		}
	}
	return withEmpresaRolePermissions(dbEmp, dbSuper, clean, resolveVerticalPermissionAction, next)
}

func nuevoVerticalPageKey(module string) string {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(module)), "_")
	var b strings.Builder
	b.WriteString("link")
	for _, part := range parts {
		if part == "" {
			continue
		}
		if part == "tms" {
			b.WriteString("TMS")
			continue
		}
		b.WriteString(strings.ToUpper(part[:1]))
		if len(part) > 1 {
			b.WriteString(part[1:])
		}
	}
	return b.String()
}
