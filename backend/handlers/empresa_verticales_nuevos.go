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
	ID         string                               `json:"id"`
	Modulo     string                               `json:"module"`
	Page       string                               `json:"page"`
	Titulo     string                               `json:"title"`
	TituloFull string                               `json:"full_title"`
	Resumen    string                               `json:"summary"`
	Secciones  []string                             `json:"sections"`
	Plantilla  dbpkg.EmpresaModuloColombiaPlantilla `json:"plantilla"`
}

func empresaVerticalesNuevosPermisos() []empresaVerticalNuevoPermiso {
	catalog := dbpkg.NuevosVerticalesTipoEmpresaCatalog()
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

func EmpresaVerticalesNuevosCatalogoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		items := buildEmpresaVerticalesNuevosCatalogo()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"total": len(items),
			"items": items,
		})
	}
}

func SuperVerticalesNuevosCatalogoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		items := buildEmpresaVerticalesNuevosCatalogo()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"total": len(items),
			"items": items,
		})
	}
}

func PublicVerticalesNuevosCatalogoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		items := buildEmpresaVerticalesNuevosCatalogo()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":    true,
			"total": len(items),
			"items": items,
		})
	}
}

func buildEmpresaVerticalesNuevosCatalogo() []empresaVerticalNuevoCatalogoItem {
	catalog := dbpkg.NuevosVerticalesTipoEmpresaCatalog()
	out := make([]empresaVerticalNuevoCatalogoItem, 0, len(catalog))
	for _, item := range catalog {
		modulo := strings.ToLower(strings.TrimSpace(item.Modulo))
		if modulo == "" {
			continue
		}
		page := nuevoVerticalPageKey(modulo)
		plantilla := dbpkg.GetEmpresaModuloColombiaPlantilla(modulo)
		out = append(out, empresaVerticalNuevoCatalogoItem{
			ID:         page,
			Modulo:     modulo,
			Page:       page,
			Titulo:     strings.TrimSpace(item.Nombre),
			TituloFull: strings.TrimSpace(item.Nombre),
			Resumen:    strings.TrimSpace(item.Observaciones),
			Secciones:  append([]string{}, plantilla.SeccionesFlujo...),
			Plantilla:  plantilla,
		})
	}
	return out
}

var empresaVerticalesNuevosModuloSet = func() map[string]bool {
	permisos := empresaVerticalesNuevosPermisos()
	out := make(map[string]bool, len(permisos))
	for _, item := range permisos {
		out[item.Modulo] = true
	}
	return out
}()

var empresaVerticalesNuevosPageByAPIPath = func() map[string]string {
	permisos := empresaVerticalesNuevosPermisos()
	out := make(map[string]string, len(permisos))
	for _, item := range permisos {
		out["/api/empresa/"+item.Modulo] = item.Page
	}
	return out
}()

func init() {
	modulos := NuevosVerticalesEmpresaModules()
	permissionPagesCatalogOrdered = append(permissionPagesCatalogOrdered, permissionPageRule{
		PaginaClave: "linkNuevosVerticales",
		AnyModules:  modulos,
		Accion:      permActionCreate,
		Titulo:      "20 nuevos verticales empresariales",
		Grupo:       "Verticales de negocio",
	})
	for _, item := range empresaVerticalesNuevosPermisos() {
		permissionModulesCatalogOrdered = append(permissionModulesCatalogOrdered, item.Modulo)
		permissionModuleDisplayNames[item.Modulo] = item.Titulo
		permissionPagesCatalogOrdered = append(permissionPagesCatalogOrdered, permissionPageRule{
			PaginaClave: item.Page,
			Modulo:      item.Modulo,
			Accion:      permActionCreate,
			Titulo:      item.Titulo,
			Grupo:       "Verticales de negocio",
		})
	}
}

func NuevosVerticalesEmpresaModules() []string {
	permisos := empresaVerticalesNuevosPermisos()
	out := make([]string, 0, len(permisos))
	for _, item := range permisos {
		out = append(out, item.Modulo)
	}
	return out
}

func isPermModuleNuevoVertical(module string) bool {
	return empresaVerticalesNuevosModuloSet[strings.ToLower(strings.TrimSpace(module))]
}

func permissionPageForNuevoVerticalAPIPath(path string) (string, bool) {
	page, ok := empresaVerticalesNuevosPageByAPIPath[strings.ToLower(strings.TrimSpace(path))]
	return page, ok
}

func WithEmpresaModuloVerticalPermissions(dbEmp, dbSuper *sql.DB, modulo string, next http.HandlerFunc) http.HandlerFunc {
	clean := strings.ToLower(strings.TrimSpace(modulo))
	if !isPermModuleNuevoVertical(clean) {
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "modulo vertical no soportado", http.StatusBadRequest)
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
