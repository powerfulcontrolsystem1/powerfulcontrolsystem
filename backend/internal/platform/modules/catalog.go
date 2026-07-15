// Package modules is the canonical, dependency-free manifest of PCS domains.
// It is intentionally data-only so planning, tests and future feature-flag
// tooling can share module boundaries without importing HTTP handlers.
package modules

import (
	"fmt"
	"sort"
	"strings"
)

type Maturity string

const (
	MaturityExperimental Maturity = "experimental"
	MaturityPilot        Maturity = "pilot"
	MaturityStable       Maturity = "stable"
	MaturityDisabled     Maturity = "disabled"
)

type Descriptor struct {
	Name         string
	Version      string
	Maturity     Maturity
	Dependencies []string
	Permissions  []string
	Tables       []string
	Endpoints    []string
	Jobs         []string
	FeatureFlag  string
}

// Catalog provides the cross-domain boundaries used in the preproduction
// review. A descriptor is not an authorization source: handlers must still use
// the existing tenant and permission middleware for every request.
func Catalog() []Descriptor {
	return []Descriptor{
		{Name: "auth", Version: "v1", Maturity: MaturityStable, Tables: []string{"administradores", "sesiones"}, Endpoints: []string{"/auth", "/api/v1/auth"}},
		{Name: "empresas", Version: "v1", Maturity: MaturityStable, Dependencies: []string{"auth"}, Permissions: []string{"empresa"}, Tables: []string{"empresas", "empresa_usuarios"}, Endpoints: []string{"/api/empresa", "/api/v1/empresas"}},
		{Name: "usuarios", Version: "v1", Maturity: MaturityStable, Dependencies: []string{"auth", "empresas"}, Permissions: []string{"usuarios"}, Tables: []string{"empresa_usuarios"}, Endpoints: []string{"/api/empresa/usuarios"}},
		{Name: "clientes", Version: "v1", Maturity: MaturityStable, Dependencies: []string{"empresas"}, Permissions: []string{"clientes"}, Tables: []string{"clientes"}, Endpoints: []string{"/api/empresa/clientes", "/api/v1/empresa/clientes"}},
		{Name: "productos", Version: "v1", Maturity: MaturityStable, Dependencies: []string{"empresas"}, Permissions: []string{"inventario"}, Tables: []string{"productos", "producto_precio_historial"}, Endpoints: []string{"/api/empresa/productos", "/api/v1/empresa/productos"}},
		{Name: "licencias", Version: "v1", Maturity: MaturityStable, Dependencies: []string{"empresas"}, Permissions: []string{"licencias"}, Tables: []string{"licencias", "licencias_empresa"}, Endpoints: []string{"/api/empresa/licencias", "/api/public/licencias"}},
		{Name: "ventas", Version: "v1", Maturity: MaturityStable, Dependencies: []string{"empresas", "clientes", "inventario", "caja"}, Permissions: []string{"ventas"}, Tables: []string{"empresa_carritos", "empresa_carrito_items"}, Endpoints: []string{"/api/empresa/carritos", "/api/v1/empresa/ventas"}},
		{Name: "inventario", Version: "v1", Maturity: MaturityStable, Dependencies: []string{"empresas", "productos"}, Permissions: []string{"inventario"}, Tables: []string{"productos", "inventario_existencias", "inventario_movimientos"}, Endpoints: []string{"/api/empresa/inventario", "/api/v1/empresa/productos"}},
		{Name: "caja", Version: "v1", Maturity: MaturityStable, Dependencies: []string{"empresas", "ventas"}, Permissions: []string{"ventas"}, Tables: []string{"empresa_cajas", "empresa_cortes_caja"}, Endpoints: []string{"/api/empresa/corte_caja"}},
		{Name: "pagos", Version: "v1", Maturity: MaturityPilot, Dependencies: []string{"ventas", "licencias"}, Permissions: []string{"ventas"}, Tables: []string{"epayco_payments", "wompi_payments"}, Endpoints: []string{"/epayco", "/wompi", "/api/v1/empresa/pagos"}, Jobs: []string{"payment.confirmation"}},
		{Name: "facturacion", Version: "v1", Maturity: MaturityPilot, Dependencies: []string{"ventas", "inventario"}, Permissions: []string{"facturacion"}, Tables: []string{"empresa_facturacion_documentos"}, Endpoints: []string{"/api/empresa/facturacion", "/api/v1/empresa/facturacion"}, Jobs: []string{"dian.submit", "dian.status"}},
		{Name: "documentos", Version: "v1", Maturity: MaturityPilot, Dependencies: []string{"empresas"}, Permissions: []string{"seguridad"}, Tables: []string{"empresa_documentos_gestion"}, Endpoints: []string{"/api/empresa/documentos"}, Jobs: []string{"document.render"}},
		{Name: "notificaciones", Version: "v1", Maturity: MaturityPilot, Dependencies: []string{"empresas", "usuarios"}, Permissions: []string{"self_service"}, Tables: []string{"empresa_buzon_mensajes"}, Endpoints: []string{"/api/empresa/buzon", "/api/v1/empresa/notificaciones"}, Jobs: []string{"email.send", "whatsapp.send"}},
		{Name: "ia", Version: "v1", Maturity: MaturityPilot, Dependencies: []string{"empresas", "documentos"}, Permissions: []string{"ventas", "seguridad"}, Tables: []string{"empresa_ai_propuestas", "empresa_ai_ejecuciones"}, Endpoints: []string{"/api/empresa/chat_con_inteligencia_artificial"}},
		{Name: "soporte_remoto", Version: "v1", Maturity: MaturityPilot, Dependencies: []string{"auth", "empresas"}, Permissions: []string{"soporte_remoto"}, Tables: []string{"empresa_soporte_remoto_dispositivos"}, Endpoints: []string{"/api/empresa/soporte_remoto"}},
		{Name: "verticales", Version: "v1", Maturity: MaturityExperimental, Dependencies: []string{"empresas", "ventas", "inventario"}, Permissions: []string{"configuracion"}, Tables: []string{"empresa_estaciones", "empresa_tarifas"}, Endpoints: []string{"/api/empresa/estaciones"}, FeatureFlag: "modulos_verticales"},
	}
}

func Validate(descriptors []Descriptor) error {
	known := make(map[string]struct{}, len(descriptors))
	for _, descriptor := range descriptors {
		name := strings.TrimSpace(descriptor.Name)
		if name == "" || strings.TrimSpace(descriptor.Version) == "" {
			return fmt.Errorf("module name and version are required")
		}
		switch descriptor.Maturity {
		case MaturityExperimental, MaturityPilot, MaturityStable, MaturityDisabled:
		default:
			return fmt.Errorf("module %s has invalid maturity", name)
		}
		if _, exists := known[name]; exists {
			return fmt.Errorf("duplicate module %s", name)
		}
		known[name] = struct{}{}
	}
	for _, descriptor := range descriptors {
		for _, dependency := range descriptor.Dependencies {
			if _, exists := known[dependency]; !exists {
				return fmt.Errorf("module %s has unknown dependency %s", descriptor.Name, dependency)
			}
		}
	}
	return nil
}

func Names() []string {
	descriptors := Catalog()
	names := make([]string, 0, len(descriptors))
	for _, descriptor := range descriptors {
		names = append(names, descriptor.Name)
	}
	sort.Strings(names)
	return names
}
