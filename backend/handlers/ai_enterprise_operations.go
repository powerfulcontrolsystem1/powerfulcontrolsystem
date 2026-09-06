package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	aipkg "github.com/you/pos-backend/ai"
	dbpkg "github.com/you/pos-backend/db"
)

func decodeEnterpriseTool(raw string, dst interface{}) error {
	if len(raw) > 64<<10 {
		return fmt.Errorf("argumentos demasiado extensos")
	}
	d := json.NewDecoder(strings.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(dst); err != nil {
		return fmt.Errorf("argumentos inválidos")
	}
	if err := d.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("argumentos inválidos")
	}
	return nil
}

// Explanation-only requests must not prepare writes, even if they mention creation.
func empresaAIExplanationOnly(question string) bool {
	q := foldEmpresaAICommandText(question)
	if strings.HasPrefix(strings.TrimLeft(q, "¿?¡! "), "como ") {
		return true
	}
	for _, marker := range []string{"explica", "como puedo", "como se ", "como crear", "como agregar", "paso a paso", "ensename", "no se como", "como hago"} {
		if strings.Contains(q, marker) {
			return true
		}
	}
	return false
}

func (c *EmpresaAIChatController) authorizedDirectDocumentResponse(empresaID int64, user, question string) (string, bool) {
	if empresaAIExplanationOnly(question) {
		return "", false
	}
	snapshot, err := getEmpresaPermissionSnapshot(c.dbEmp, c.dbSuper, user, empresaID)
	if err != nil {
		return "", false
	}
	if allowed, _ := empresaPermissionSnapshotAllowsAdditionalModule(snapshot, permModuleInventario, permActionRead, ""); !allowed {
		return "", false
	}
	return buildEmpresaAIDirectDocumentResponse(c.dbEmp, empresaID, question)
}

func enterpriseAIChatTools(ctx aipkg.ExecutionContext, question string) []map[string]interface{} {
	if !enterpriseAIAgentModeEnabled() {
		return nil
	}
	available := enterpriseAIAvailableTools(ctx)
	var out []map[string]interface{}
	for _, tool := range aipkg.ResponsesToolDefinitions(ctx) {
		name, _ := tool["name"].(string)
		def, ok := available[name]
		if ok && !(empresaAIExplanationOnly(question) && def.Confirmation == "required") {
			out = append(out, tool)
		}
	}
	return out
}

func empresaAIGuidance(snapshot empresaPermissionSnapshot) string {
	var b strings.Builder
	b.WriteString("\nAyuda de PCS: cuando pidan explicar, enseña pasos numerados, pantalla, campos y resultado esperado. No prepares escrituras si solo piden instrucciones. No inventes botones ni capacidades. Si falta una pantalla o dato, pregunta cuál está viendo. Las herramientas disponibles son el límite de ejecución; para otras funciones ofrece orientación y reconoce el límite. Fecha actual: " + time.Now().Format("2006-01-02") + ".\nPantallas autorizadas (usa estos nombres):\n")
	for _, page := range permissionPagesCatalogOrdered {
		if ok, _ := empresaPermissionSnapshotAllowsAdditionalModule(snapshot, page.Modulo, page.Accion, page.PaginaClave); ok {
			b.WriteString("- " + page.Grupo + ": " + page.Titulo + "\n")
		}
	}
	b.WriteString("Para crear productos: Inventario / Productos, Nuevo producto, nombre, precio e impuesto; categoría y bodega según configuración; revisar y guardar. Para agregar consumos: Estaciones, abrir la habitación o mesa existente, buscar producto, indicar cantidad, revisar cuenta. Agregar un consumo no cobra ni cierra la cuenta. Para reportes: Reportes, seleccionar tipo y periodo, generar y descargar. Ajusta estos pasos a las pantallas autorizadas; no concedas permisos ausentes.\n")
	return b.String()
}

// Legacy context includes finance, customers and sales together. Restrict that
// aggregate to users entitled to every included domain; other roles use tools.
func (c *EmpresaAIChatController) roleScopedChatContext(empresaID int64, question, user, page, model string) (string, error) {
	snapshot, err := getEmpresaPermissionSnapshot(c.dbEmp, c.dbSuper, user, empresaID)
	if err != nil || !snapshot.CanAccess {
		return "", fmt.Errorf("contexto no autorizado")
	}
	base := "Empresa activa validada. Consulta datos operativos únicamente con herramientas autorizadas. No inventes cifras."
	all := empresaAIAdministrativeDBReadRole(snapshot.EffectiveRole)
	for _, module := range []string{permModuleInventario, permModuleVentas, permModuleClientes, permModuleFinanzas, permModuleCompras, permModuleReportes, permModuleSeguridad} {
		ok, _ := empresaPermissionSnapshotAllowsAdditionalModule(snapshot, module, "R", "")
		all = all && ok
	}
	if all {
		opts := c.contextoPreguntaOptionsForAccount(empresaID, user, model)
		// Arbitrary configuration/JSON tables may contain secrets or personal
		// memory despite having empresa_id. Business tools provide other domains.
		opts.DBQueryAllowedTables = map[string]bool{"productos": true, "categorias_productos": true, "bodegas": true, "inventario_existencias": true, "carritos_compras": true, "carrito_compra_items": true, "empresa_finanzas_movimientos": true}
		base, err = dbpkg.BuildEmpresaAIContextoForQuestionWithOptions(c.dbEmp, empresaID, question, user, page, opts)
		if err != nil {
			return "", err
		}
	}
	return base + empresaAIGuidance(snapshot), nil
}

type enterpriseStationProductArgs struct {
	EstacionID int64 `json:"estacion_id"`
	ProductoID int64 `json:"producto_id"`
	Cantidad   int64 `json:"cantidad"`
}

func validateEnterpriseStationProductArgs(p enterpriseStationProductArgs) error {
	if p.EstacionID <= 0 || p.ProductoID <= 0 || p.Cantidad < 1 || p.Cantidad > 99 {
		return fmt.Errorf("indica estación, producto y cantidad entre 1 y 99")
	}
	return nil
}

func dispatchEnterpriseAIOperation(dbEmp *sql.DB, r *http.Request, ctx aipkg.ExecutionContext, call openAIResponsesFunctionCall) (string, error) {
	if !enterpriseAIRequireTool(ctx, call.Name) {
		return "", fmt.Errorf("herramienta no autorizada")
	}
	if call.Name == aipkg.ToolCatalogCreateProduct {
		return dispatchEnterpriseAIResponsesFunctionCall(dbEmp, r, ctx, call)
	}
	var result interface{}
	switch call.Name {
	case aipkg.ToolCatalogSearchProducts, aipkg.ToolSalesInspectStation:
		var args struct {
			Q string `json:"q"`
		}
		if err := decodeEnterpriseTool(call.Arguments, &args); err != nil || len(args.Q) > 160 {
			return "", fmt.Errorf("búsqueda inválida")
		}
		products, err := dbpkg.GetProductosByEmpresa(dbEmp, ctx.EmpresaID, args.Q, "activo", 0, 0, 20, 0)
		if err != nil {
			return "", err
		}
		// Explicit provider projection: no costs, account data or contact fields.
		items := make([]map[string]interface{}, 0, len(products))
		for _, p := range products {
			items = append(items, map[string]interface{}{"producto_id": p.ID, "nombre": p.Nombre, "sku": p.SKU, "precio": p.Precio, "impuesto_porcentaje": p.ImpuestoPorcentaje})
		}
		out := map[string]interface{}{"productos": items}
		if call.Name == aipkg.ToolCatalogSearchProducts {
			categories, err := dbpkg.GetCategoriasProductoByEmpresa(dbEmp, ctx.EmpresaID, false, "")
			if err != nil {
				return "", err
			}
			warehouses, err := dbpkg.GetBodegasByEmpresa(dbEmp, ctx.EmpresaID, false)
			if err != nil {
				return "", err
			}
			cats := make([]map[string]interface{}, 0)
			stores := make([]map[string]interface{}, 0)
			for i, c := range categories {
				if i >= 50 {
					break
				}
				cats = append(cats, map[string]interface{}{"id": c.ID, "nombre": c.Nombre})
			}
			for i, b := range warehouses {
				if i >= 50 {
					break
				}
				stores = append(stores, map[string]interface{}{"id": b.ID, "nombre": b.Nombre})
			}
			out["categorias"] = cats
			out["bodegas"] = stores
		}
		if call.Name == aipkg.ToolSalesInspectStation {
			pref, err := dbpkg.GetEmpresaEstacionPref(dbEmp, ctx.EmpresaID, 0, "estaciones_config")
			if err != nil {
				return "", err
			}
			stations := make([]map[string]interface{}, 0)
			if pref != nil {
				for _, station := range parseEstacionesNombresFromPref(pref.Valor) {
					cart, err := dbpkg.GetCarritoCompraByStation(dbEmp, ctx.EmpresaID, station.ID)
					if err == sql.ErrNoRows {
						continue
					}
					if err != nil {
						return "", err
					}
					if cart != nil {
						stations = append(stations, map[string]interface{}{"estacion_id": station.ID, "nombre": station.Nombre, "cuenta_abierta": cart.Estado == "activo" && cart.EstadoCarrito == "abierto" && !isCarritoVentaPagada(cart)})
					}
					if len(stations) >= 100 {
						break
					}
				}
			}
			out["estaciones"] = stations
		}
		result = out
	case aipkg.ToolSalesAddStationProduct:
		if !enterpriseAIWriteToolEnabled(call.Name) {
			return "", fmt.Errorf("herramienta desactivada")
		}
		var args enterpriseStationProductArgs
		if err := decodeEnterpriseTool(call.Arguments, &args); err != nil {
			return "", err
		}
		if err := validateEnterpriseStationProductArgs(args); err != nil {
			return "", err
		}
		cart, err := dbpkg.GetCarritoCompraByStation(dbEmp, ctx.EmpresaID, args.EstacionID)
		if err != nil || cart == nil || cart.Estado != "activo" || cart.EstadoCarrito != "abierto" || isCarritoVentaPagada(cart) {
			return "", fmt.Errorf("la habitación o estación debe tener una cuenta abierta")
		}
		product, err := dbpkg.GetProductoByID(dbEmp, ctx.EmpresaID, args.ProductoID)
		if err != nil || product == nil || product.Estado != "activo" {
			return "", fmt.Errorf("producto no disponible")
		}
		plan := dbpkg.EmpresaAIStationProductPlan{EstacionID: args.EstacionID, CarritoID: cart.ID, ProductoID: product.ID, Cantidad: args.Cantidad, Precio: product.Precio, Impuesto: product.ImpuestoPorcentaje, ActivadoEn: cart.ActivadoEn}
		raw, _ := json.Marshal(plan)
		id, err := aipkg.NewOpaqueID("proposal")
		if err != nil {
			return "", err
		}
		summary := fmt.Sprintf("Agregar %d × %s a %s. Precio unitario: %.2f; impuesto: %.2f%%. No cobra ni cierra la cuenta.", args.Cantidad, product.Nombre, cart.Nombre, product.Precio, product.ImpuestoPorcentaje)
		p, err := dbpkg.CreateEmpresaAIProposal(dbEmp, dbpkg.EmpresaAIProposal{ProposalID: id, ConversationID: ctx.ConversationID, EmpresaID: ctx.EmpresaID, UsuarioCreador: ctx.UserID, ToolName: call.Name, RiskLevel: "medium", PlanJSON: string(raw), Resumen: summary, EstadoAnterior: `{}`, EstadoEsperado: `{}`, RollbackPolicy: "transactional_before_commit", Estado: dbpkg.AIProposalAwaitingConfirmation}, 15*time.Minute)
		if err != nil {
			return "", err
		}
		result = map[string]interface{}{"status": p.Estado, "proposal_id": p.ProposalID, "summary": p.Resumen, "confirmation_required": true}
	case aipkg.ToolReportsGenerate:
		var args struct {
			Dataset string `json:"dataset"`
			Desde   string `json:"desde"`
			Hasta   string `json:"hasta"`
		}
		if err := decodeEnterpriseTool(call.Arguments, &args); err != nil {
			return "", err
		}
		key, permission := enterpriseAIReportDataset(args.Dataset)
		if key == "" || !aipkg.ToolAllowed(aipkg.ToolDefinition{RequiredPermissions: []string{permission}}, ctx.Permissions) {
			return "", fmt.Errorf("reporte no autorizado")
		}
		from, e1 := time.Parse("2006-01-02", args.Desde)
		to, e2 := time.Parse("2006-01-02", args.Hasta)
		if e1 != nil || e2 != nil || to.Before(from) || to.Sub(from) > 366*24*time.Hour {
			return "", fmt.Errorf("indica un periodo válido de hasta un año")
		}
		b := newReportesAIBuilder(r.Context(), dbEmp, ctx.EmpresaID, args.Desde, args.Hasta)
		b.maxRows = 50
		ds, err := b.buildDataset(key)
		if err != nil {
			return "", err
		}
		q := url.Values{"empresa_id": {fmt.Sprint(ctx.EmpresaID)}, "action": {"export"}, "dataset": {key}, "desde": {args.Desde}, "hasta": {args.Hasta}, "format": {"pdf"}}
		result = map[string]interface{}{"titulo": ds.Title, "desde": ds.Desde, "hasta": ds.Hasta, "columnas": ds.Columns, "filas": ds.Rows, "resumen": ds.Summary, "limite_filas": 50, "descarga_pdf": "/api/empresa/reportes?" + q.Encode()}
	default:
		return "", fmt.Errorf("herramienta IA no registrada")
	}
	_ = dbpkg.RecordEmpresaAIExecution(dbEmp, dbpkg.EmpresaAIExecution{EmpresaID: ctx.EmpresaID, UsuarioID: ctx.UserID, ConversationID: ctx.ConversationID, ToolName: call.Name, Modo: ctx.Mode, RiskLevel: aipkg.Registry()[call.Name].RiskLevel, Resultado: "completed", CategoriasJSON: `["internal"]`})
	raw, err := json.Marshal(result)
	return string(raw), err
}

func enterpriseAIReportDataset(name string) (string, string) {
	switch name {
	case "ventas":
		return reporteDatasetVentasDiariasMetodoPago, "ventas:R"
	case "productos":
		return reporteDatasetOperativoTopProductos, "ventas:R"
	case "inventario":
		return reporteDatasetOperativoInventario, "inventario:R"
	case "compras":
		return reporteDatasetOperativoCompras, "compras:R"
	case "resultados":
		return reporteDatasetContableEstadoResultados, "finanzas:R"
	case "caja":
		return reporteDatasetContableFlujoCaja, "finanzas:R"
	default:
		return "", ""
	}
}

func enterpriseAIConfirmStationProduct(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, ctx aipkg.ExecutionContext, p *dbpkg.EmpresaAIProposal) {
	var plan dbpkg.EmpresaAIStationProductPlan
	if err := json.Unmarshal([]byte(p.PlanJSON), &plan); err != nil {
		http.Error(w, "plan inválido", http.StatusConflict)
		return
	}
	id, err := dbpkg.AddEmpresaAIStationProduct(dbEmp, ctx.EmpresaID, plan, ctx.UserID)
	if err != nil {
		_ = dbpkg.FinishEmpresaAIProposal(dbEmp, ctx.EmpresaID, p.ProposalID, dbpkg.AIProposalFailed, `{"error":"cuenta o producto modificados"}`)
		http.Error(w, "No se agregó el consumo. Revisa que la cuenta siga abierta, que el precio no haya cambiado y que exista stock.", http.StatusConflict)
		return
	}
	raw, _ := json.Marshal(map[string]interface{}{"item_id": id, "carrito_id": plan.CarritoID, "verified": true})
	if err := dbpkg.FinishEmpresaAIProposal(dbEmp, ctx.EmpresaID, p.ProposalID, dbpkg.AIProposalCompleted, string(raw)); err != nil {
		http.Error(w, "Consumo agregado; cierre de propuesta pendiente de conciliación. No repitas la operación.", http.StatusInternalServerError)
		return
	}
	registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, ctx.EmpresaID, "ventas", "consumo_ia_confirmado", "carrito_compra_items", id, http.StatusOK, map[string]interface{}{"proposal_id": p.ProposalID}, "consumo confirmado desde chat")
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "status": dbpkg.AIProposalCompleted, "result": json.RawMessage(raw)})
}
