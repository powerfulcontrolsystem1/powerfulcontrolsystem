package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

type empresaVerticalIntegracionItem struct {
	ID                   string   `json:"id"`
	Modulo               string   `json:"module"`
	Page                 string   `json:"page"`
	Titulo               string   `json:"title"`
	IntegrationStatus    string   `json:"integration_status"`
	OperationalVisible   bool     `json:"operational_visible"`
	CoreModules          []string `json:"core_modules"`
	TemplateActivates    []string `json:"template_activates,omitempty"`
	TablesTouched        []string `json:"tables_touched,omitempty"`
	RequiredPermissions  []string `json:"required_permissions,omitempty"`
	SaleFlow             []string `json:"sale_flow,omitempty"`
	ReportsProduced      []string `json:"reports_produced,omitempty"`
	FinancialCoreModules []string `json:"financial_core_modules,omitempty"`
	IncomeFlow           []string `json:"income_flow,omitempty"`
	ExpenseFlow          []string `json:"expense_flow,omitempty"`
	FinancialTables      []string `json:"financial_tables,omitempty"`
	FinancialReports     []string `json:"financial_reports,omitempty"`
	DuplicatesCore       []string `json:"duplicates_core"`
	OwnFlowAllowed       []string `json:"own_flow_allowed"`
	Decision             string   `json:"decision"`
	AliasDe              string   `json:"alias_of,omitempty"`
	FusedModules         []string `json:"fused_modules,omitempty"`
	SupportModules       []string `json:"support_modules,omitempty"`
	SimilarTemplates     []string `json:"similar_templates,omitempty"`
	Motivo               string   `json:"reason"`
	ProfessionalReady    bool     `json:"professional_ready"`
	ReadinessScore       int      `json:"readiness_score"`
	ReadinessChecks      []string `json:"readiness_checks,omitempty"`
	ConfigurationScope   []string `json:"configuration_scope,omitempty"`
}

var empresaPlantillasCoreModules = []string{"clientes", "inventario", "ventas", "pagos", "finanzas", "facturacion", "reportes", "seguridad"}

func EmpresaPlantillasIntegracionCatalogoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		items := buildEmpresaPlantillasIntegracionCatalogo()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"total": len(items),
			"items": items,
		})
	}
}

func SuperPlantillasIntegracionCatalogoHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper); !ok {
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		items := buildEmpresaPlantillasIntegracionCatalogo()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"total": len(items),
			"items": items,
		})
	}
}

func PublicPlantillasIntegracionCatalogoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		items := buildEmpresaPlantillasIntegracionCatalogo()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":    true,
			"total": len(items),
			"items": items,
		})
	}
}

func buildEmpresaPlantillasIntegracionCatalogo() []empresaVerticalIntegracionItem {
	items := nuevasPlantillasIntegracionItems()
	for idx := range items {
		items[idx] = enrichEmpresaVerticalReadiness(items[idx])
	}
	return items
}

func nuevasPlantillasIntegracionItems() []empresaVerticalIntegracionItem {
	catalog := dbpkg.NuevasPlantillasTipoEmpresaCatalog()
	out := make([]empresaVerticalIntegracionItem, 0, len(catalog))
	for _, item := range catalog {
		modulo := strings.ToLower(strings.TrimSpace(item.Modulo))
		if modulo == "" {
			continue
		}
		plantilla := dbpkg.GetEmpresaModuloColombiaPlantilla(modulo)
		integracion := dbpkg.BuildTipoEmpresaPreconfigIntegracionVertical(modulo)
		if integracion == nil {
			continue
		}
		page := nuevoVerticalPageKey(modulo)
		title := strings.TrimSpace(firstNonEmptyString(item.Nombre, plantilla.Titulo, modulo))
		reason := strings.TrimSpace(integracion.MotivoDecision)
		if reason == "" {
			reason = "Plantilla real conectada al nucleo comun sin duplicar clientes, productos, ventas, pagos ni reportes."
		}
		out = append(out, empresaVerticalIntegracionItem{
			ID:                   page,
			Modulo:               modulo,
			Page:                 page,
			Titulo:               title,
			IntegrationStatus:    strings.TrimSpace(integracion.EstadoIntegracion),
			OperationalVisible:   true,
			CoreModules:          append([]string{}, empresaPlantillasCoreModules...),
			TemplateActivates:    copyStringSlice(integracion.TemplateActivates),
			TablesTouched:        copyStringSlice(integracion.TablesTouched),
			RequiredPermissions:  copyStringSlice(integracion.RequiredPermissions),
			SaleFlow:             copyStringSlice(integracion.SaleFlow),
			ReportsProduced:      copyStringSlice(integracion.ReportsProduced),
			FinancialCoreModules: copyStringSlice(integracion.FinancialCoreModules),
			IncomeFlow:           copyStringSlice(integracion.IncomeFlow),
			ExpenseFlow:          copyStringSlice(integracion.ExpenseFlow),
			FinancialTables:      copyStringSlice(integracion.FinancialTables),
			FinancialReports:     copyStringSlice(integracion.FinancialReports),
			DuplicatesCore:       []string{},
			OwnFlowAllowed:       copyStringSlice(plantilla.SeccionesFlujo),
			Decision:             strings.TrimSpace(integracion.Decision),
			Motivo:               reason,
			ConfigurationScope:   []string{"tipo_empresa_preconfiguracion", "licencia", "roles", "menu", "datos_guia", "reportes"},
		})
	}
	return out
}

func enrichEmpresaVerticalReadiness(item empresaVerticalIntegracionItem) empresaVerticalIntegracionItem {
	checks := make([]string, 0, 8)
	total := 0
	ok := 0
	add := func(name string, passed bool) {
		total++
		if passed {
			ok++
			checks = append(checks, name)
		}
	}

	status := strings.ToLower(strings.TrimSpace(item.IntegrationStatus))
	add("visible_operativo", item.OperationalVisible)
	add("sin_duplicados_del_nucleo", len(item.DuplicatesCore) == 0)
	add("plantilla_de_configuracion", len(item.TemplateActivates) > 0)
	add("tablas_y_datos_declarados", len(item.TablesTouched) > 0)
	add("permisos_declarados", len(item.RequiredPermissions) > 0)
	add("flujo_de_venta_o_operacion", len(item.SaleFlow) > 0)
	add("reportes_declarados", len(item.ReportsProduced) > 0)
	add("nucleo_financiero_declarado", hasAllStringValues(item.FinancialCoreModules, []string{"finanzas", "ventas", "pagos", "reportes"}))
	add("ingresos_del_nucleo_declarados", len(item.IncomeFlow) > 0)
	add("egresos_del_nucleo_declarados", len(item.ExpenseFlow) > 0)
	add("tablas_financieras_declaradas", hasAllStringValues(item.FinancialTables, []string{"empresa_finanzas_movimientos"}))
	add("reportes_financieros_declarados", len(item.FinancialReports) > 0)
	if status == "integrado_soporte" {
		add("nucleo_soporte_reportes_seguridad", hasAllStringValues(item.CoreModules, []string{"seguridad", "reportes"}))
	} else {
		add("nucleo_comercial_completo", hasAllStringValues(item.CoreModules, []string{"clientes", "inventario", "ventas", "pagos", "finanzas", "reportes"}))
	}

	score := 0
	if total > 0 {
		score = (ok * 100) / total
	}
	item.ReadinessScore = score
	item.ProfessionalReady = score == 100
	item.ReadinessChecks = checks
	if len(item.ConfigurationScope) == 0 {
		item.ConfigurationScope = []string{"tipo_empresa", "licencia", "roles", "menu", "datos_guia", "reportes"}
	}
	return item
}

func hasAllStringValues(values []string, required []string) bool {
	seen := map[string]bool{}
	for _, value := range values {
		clean := strings.ToLower(strings.TrimSpace(value))
		if clean != "" {
			seen[clean] = true
		}
	}
	for _, value := range required {
		if !seen[strings.ToLower(strings.TrimSpace(value))] {
			return false
		}
	}
	return true
}

func copyStringSlice(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}
