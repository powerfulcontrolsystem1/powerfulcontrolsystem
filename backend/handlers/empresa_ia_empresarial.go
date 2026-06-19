package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	empresaIAEmpresarialUsageID = "openai:gpt-5.4-mini:centro_ia_empresarial"
	empresaIAEmpresarialLimit   = 12
)

type empresaIAFuncion struct {
	ID          string `json:"id"`
	Titulo      string `json:"titulo"`
	Area        string `json:"area"`
	Descripcion string `json:"descripcion"`
	Prompt      string `json:"prompt"`
}

type empresaIAEmpresarialPayload struct {
	EmpresaID int64  `json:"empresa_id"`
	Accion    string `json:"accion"`
	Consulta  string `json:"consulta"`
	Desde     string `json:"desde"`
	Hasta     string `json:"hasta"`
	AgentID   string `json:"agent_id,omitempty"`
}

type empresaIAEmpresarialSnapshot struct {
	EmpresaID       int64                  `json:"empresa_id"`
	EmpresaNombre   string                 `json:"empresa_nombre"`
	EmpresaNIT      string                 `json:"empresa_nit,omitempty"`
	Desde           string                 `json:"desde"`
	Hasta           string                 `json:"hasta"`
	GeneradoEn      string                 `json:"generado_en"`
	Metricas        map[string]interface{} `json:"metricas"`
	Alertas         []string               `json:"alertas"`
	Recomendaciones []string               `json:"recomendaciones_base"`
}

func EmpresaIAEmpresarialHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		payload, err := parseEmpresaIAEmpresarialPayload(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if payload.EmpresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}
		payload.Accion = normalizeEmpresaIAAccion(payload.Accion)
		if payload.Accion == "" {
			payload.Accion = "diagnostico_erp"
		}

		snapshot := buildEmpresaIAEmpresarialSnapshot(dbEmp, payload)
		resp := map[string]interface{}{
			"ok":        true,
			"funciones": empresaIAEmpresarialFunciones(),
			"agentes":   empresaAIChatAgentCatalog(),
			"snapshot":  snapshot,
		}
		if r.Method == http.MethodPost {
			resp["accion"] = payload.Accion
			resp["ia"] = buildEmpresaIAEmpresarialResponse(r, dbEmp, dbSuper, payload, snapshot)
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

func parseEmpresaIAEmpresarialPayload(r *http.Request) (empresaIAEmpresarialPayload, error) {
	var payload empresaIAEmpresarialPayload
	if r.Method == http.MethodPost && r.Body != nil {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && err != io.EOF {
			return payload, fmt.Errorf("JSON invalido")
		}
	}
	q := r.URL.Query()
	if payload.EmpresaID <= 0 {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			return payload, err
		}
		payload.EmpresaID = empresaID
	}
	if strings.TrimSpace(payload.Accion) == "" {
		payload.Accion = strings.TrimSpace(q.Get("action"))
	}
	if strings.TrimSpace(payload.Consulta) == "" {
		payload.Consulta = strings.TrimSpace(q.Get("consulta"))
	}
	if strings.TrimSpace(payload.Desde) == "" {
		payload.Desde = strings.TrimSpace(q.Get("desde"))
	}
	if strings.TrimSpace(payload.Hasta) == "" {
		payload.Hasta = strings.TrimSpace(q.Get("hasta"))
	}
	if strings.TrimSpace(payload.AgentID) == "" {
		payload.AgentID = strings.TrimSpace(q.Get("agent_id"))
	}
	payload.AgentID = normalizeEmpresaAIChatAgentID(payload.AgentID)
	if payload.AgentID == "general" {
		payload.AgentID = defaultEmpresaIAAgentForAction(payload.Accion)
	}
	payload.Desde, payload.Hasta = normalizeEmpresaIADateRange(payload.Desde, payload.Hasta)
	if len([]rune(payload.Consulta)) > 1800 {
		payload.Consulta = string([]rune(payload.Consulta)[:1800])
	}
	return payload, nil
}

func normalizeEmpresaIADateRange(desde, hasta string) (string, string) {
	now := time.Now()
	h := strings.TrimSpace(hasta)
	if h == "" {
		h = now.Format("2006-01-02")
	}
	d := strings.TrimSpace(desde)
	if d == "" {
		d = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	}
	return d, h
}

func normalizeEmpresaIAAccion(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "catalogo", "dashboard", "snapshot":
		return ""
	case "factura", "factura_borrador", "borrador_factura", "cotizacion":
		return "borrador_factura"
	case "cobranza", "pagos", "record_payment":
		return "cobranza_pagos"
	case "inventario", "productos", "stock":
		return "inventario_productos"
	case "conciliacion", "conciliacion_bancaria", "bancos":
		return "conciliacion_bancaria"
	case "cumplimiento", "dian", "impuestos":
		return "cumplimiento_dian"
	case "compras", "gastos", "soportes":
		return "compras_gastos"
	case "diagnostico", "diagnostico_erp", "analisis":
		return "diagnostico_erp"
	default:
		if v == "" {
			return ""
		}
		return "diagnostico_erp"
	}
}

func defaultEmpresaIAAgentForAction(accion string) string {
	switch normalizeEmpresaIAAccion(accion) {
	case "borrador_factura", "cobranza_pagos":
		return "ventas"
	case "inventario_productos":
		return "inventario"
	case "compras_gastos":
		return "compras"
	case "cumplimiento_dian":
		return "impuestos"
	default:
		return "general"
	}
}

func empresaIAEmpresarialFunciones() []empresaIAFuncion {
	return []empresaIAFuncion{
		{ID: "diagnostico_erp", Titulo: "Diagnostico ERP", Area: "Gerencia", Descripcion: "Analiza ventas, clientes, productos, finanzas y alertas con datos reales de la empresa.", Prompt: "Haz un diagnostico ejecutivo de la empresa y prioriza 5 acciones."},
		{ID: "borrador_factura", Titulo: "Borrador factura/cotizacion", Area: "Facturacion", Descripcion: "Convierte una instruccion en borrador revisable de cliente, items, impuestos y datos faltantes; no emite documentos automaticamente.", Prompt: "Prepara un borrador de factura o cotizacion con datos faltantes y validaciones antes de emitir."},
		{ID: "cobranza_pagos", Titulo: "Cobranza y pagos", Area: "Cartera", Descripcion: "Propone seguimiento de pagos, recaudo y conciliacion del ciclo de cobranza.", Prompt: "Revisa la cobranza, pagos pendientes y acciones de recaudo del periodo."},
		{ID: "inventario_productos", Titulo: "Inventario inteligente", Area: "Inventario", Descripcion: "Detecta faltantes, catalogo incompleto, bajo stock y productos que merecen revision.", Prompt: "Analiza inventario, productos, rotacion y prioridades de reposicion."},
		{ID: "conciliacion_bancaria", Titulo: "Conciliacion bancaria", Area: "Finanzas", Descripcion: "Compara ventas cerradas contra ingresos financieros y sugiere cruces de conciliacion.", Prompt: "Compara ventas e ingresos financieros y senala diferencias de conciliacion."},
		{ID: "compras_gastos", Titulo: "Compras y gastos IA", Area: "Compras", Descripcion: "Sugiere controles sobre soportes de gastos, proveedores, duplicados y contabilizacion.", Prompt: "Revisa compras y gastos registrados y recomienda controles antes de contabilizar."},
		{ID: "cumplimiento_dian", Titulo: "Cumplimiento DIAN", Area: "Fiscal", Descripcion: "Resume riesgos de facturacion electronica, impuestos, renta y soportes que debe validar el contador.", Prompt: "Revisa riesgos DIAN, impuestos y cierre contable sin asumir habilitaciones no verificadas."},
	}
}

func buildEmpresaIAEmpresarialSnapshot(dbEmp *sql.DB, payload empresaIAEmpresarialPayload) empresaIAEmpresarialSnapshot {
	metrics := map[string]interface{}{}
	alertas := []string{}
	recs := []string{}
	var nombre, nit string
	_ = dbpkg.QueryRowCompat(dbEmp, `SELECT COALESCE(nombre, ''), COALESCE(nit, '') FROM empresas WHERE id = ? LIMIT 1`, payload.EmpresaID).Scan(&nombre, &nit)

	var ventasCount int64
	var ventasTotal float64
	_ = dbpkg.QueryRowCompat(dbEmp, `SELECT COUNT(1), COALESCE(SUM(COALESCE(total,0)),0)
		FROM carritos_compras
		WHERE empresa_id = ?
		  AND LOWER(COALESCE(estado,'activo')) <> 'inactivo'
		  AND LOWER(COALESCE(estado_carrito,'')) IN ('cerrado','pagado','finalizado')
		  AND LEFT(COALESCE(NULLIF(pagado_en,''), NULLIF(fecha_actualizacion,''), fecha_creacion),10) BETWEEN ? AND ?`,
		payload.EmpresaID, payload.Desde, payload.Hasta).Scan(&ventasCount, &ventasTotal)
	metrics["ventas_cerradas"] = ventasCount
	metrics["ventas_total"] = ventasTotal

	var ingresos, egresos float64
	_ = dbpkg.QueryRowCompat(dbEmp, `SELECT COALESCE(SUM(COALESCE(NULLIF(total_neto,0), NULLIF(total,0), monto, 0)),0)
		FROM empresa_finanzas_movimientos
		WHERE empresa_id = ? AND LOWER(COALESCE(tipo_movimiento,'')) = 'ingreso'
		  AND LOWER(COALESCE(estado,'activo')) = 'activo'
		  AND LEFT(COALESCE(NULLIF(fecha_movimiento,''), fecha_creacion),10) BETWEEN ? AND ?`,
		payload.EmpresaID, payload.Desde, payload.Hasta).Scan(&ingresos)
	_ = dbpkg.QueryRowCompat(dbEmp, `SELECT COALESCE(SUM(COALESCE(NULLIF(total_neto,0), NULLIF(total,0), monto, 0)),0)
		FROM empresa_finanzas_movimientos
		WHERE empresa_id = ? AND LOWER(COALESCE(tipo_movimiento,'')) = 'egreso'
		  AND LOWER(COALESCE(estado,'activo')) = 'activo'
		  AND LEFT(COALESCE(NULLIF(fecha_movimiento,''), fecha_creacion),10) BETWEEN ? AND ?`,
		payload.EmpresaID, payload.Desde, payload.Hasta).Scan(&egresos)
	metrics["ingresos_financieros"] = ingresos
	metrics["egresos_financieros"] = egresos
	metrics["balance_financiero"] = ingresos - egresos

	var clientes, productos, servicios int64
	_ = dbpkg.QueryRowCompat(dbEmp, `SELECT COUNT(1) FROM clientes WHERE empresa_id = ? AND COALESCE(estado,'activo') <> 'inactivo'`, payload.EmpresaID).Scan(&clientes)
	_ = dbpkg.QueryRowCompat(dbEmp, `SELECT COUNT(1) FROM productos WHERE empresa_id = ? AND COALESCE(estado,'activo') = 'activo'`, payload.EmpresaID).Scan(&productos)
	_ = dbpkg.QueryRowCompat(dbEmp, `SELECT COUNT(1) FROM servicios WHERE empresa_id = ? AND COALESCE(estado,'activo') = 'activo'`, payload.EmpresaID).Scan(&servicios)
	metrics["clientes_activos"] = clientes
	metrics["productos_activos"] = productos
	metrics["servicios_activos"] = servicios

	var bajoStock int64
	_ = dbpkg.QueryRowCompat(dbEmp, `SELECT COUNT(1)
		FROM productos p
		LEFT JOIN (
		  SELECT empresa_id, producto_id, SUM(COALESCE(cantidad,0)) cantidad
		  FROM inventario_existencias
		  WHERE empresa_id = ? AND COALESCE(estado,'activo') <> 'inactivo'
		  GROUP BY empresa_id, producto_id
		) ex ON ex.empresa_id = p.empresa_id AND ex.producto_id = p.id
		WHERE p.empresa_id = ?
		  AND COALESCE(p.estado,'activo') = 'activo'
		  AND COALESCE(p.stock_minimo,0) > 0
		  AND COALESCE(ex.cantidad,0) <= COALESCE(p.stock_minimo,0)`,
		payload.EmpresaID, payload.EmpresaID).Scan(&bajoStock)
	metrics["productos_bajo_stock"] = bajoStock

	diferenciaConciliacion := ventasTotal - ingresos
	metrics["diferencia_ventas_vs_ingresos"] = diferenciaConciliacion

	if ventasCount == 0 {
		alertas = append(alertas, "No hay ventas cerradas en el periodo seleccionado.")
	}
	if clientes == 0 {
		alertas = append(alertas, "No hay clientes activos registrados para cruzar facturacion o cartera.")
	}
	if productos == 0 && servicios == 0 {
		alertas = append(alertas, "No hay catalogo activo de productos o servicios para facturacion asistida.")
	}
	if bajoStock > 0 {
		alertas = append(alertas, fmt.Sprintf("%d productos estan en o por debajo del stock minimo.", bajoStock))
	}
	if ventasTotal > 0 && absFloat64(diferenciaConciliacion) > ventasTotal*0.05 {
		alertas = append(alertas, "Las ventas cerradas y los ingresos financieros difieren mas del 5%; conviene conciliar.")
	}
	if len(alertas) == 0 {
		alertas = append(alertas, "No se detectaron alertas criticas con los datos resumidos del periodo.")
	}
	recs = append(recs,
		"Revisar que clientes y productos tengan datos fiscales completos antes de emitir documentos electronicos.",
		"Conciliar ventas cerradas contra ingresos financieros antes de cierre contable.",
		"Usar borradores IA solo como preparacion; la emision, pagos y anulaciones deben confirmarse en los modulos operativos.",
	)

	return empresaIAEmpresarialSnapshot{
		EmpresaID: payload.EmpresaID, EmpresaNombre: strings.TrimSpace(nombre), EmpresaNIT: strings.TrimSpace(nit),
		Desde: payload.Desde, Hasta: payload.Hasta, GeneradoEn: time.Now().Format("2006-01-02 15:04:05"),
		Metricas: metrics, Alertas: alertas, Recomendaciones: recs,
	}
}

func buildEmpresaIAEmpresarialResponse(r *http.Request, dbEmp, dbSuper *sql.DB, payload empresaIAEmpresarialPayload, snapshot empresaIAEmpresarialSnapshot) map[string]interface{} {
	if !isSuperAIEnabled(dbSuper) {
		return map[string]interface{}{"ok": false, "code": "ai_disabled", "error": "La IA esta desactivada desde super administrador; el snapshot real queda disponible."}
	}
	empresaChatEnabled, _, _, err := getChatIAEmpresaEnabled(dbSuper)
	if err != nil || !empresaChatEnabled {
		return map[string]interface{}{"ok": false, "code": "ai_empresa_disabled", "error": "El chat IA empresarial esta desactivado; el snapshot real queda disponible."}
	}
	model, ok := availableEmpresaAIModelMap(dbSuper)["openai:gpt-5.4-mini"]
	if !ok {
		return map[string]interface{}{"ok": false, "code": "ai_model_missing", "error": "GPT-5.4 mini no esta disponible en el catalogo IA."}
	}
	fechaUso := time.Now().Format("2006-01-02")
	uso, err := dbpkg.GetEmpresaAIUsoDiario(dbEmp, payload.EmpresaID, model.Provider, empresaIAEmpresarialUsageID, fechaUso)
	if err != nil {
		return map[string]interface{}{"ok": false, "code": "usage_error", "error": "No se pudo consultar uso diario de IA."}
	}
	if uso.Consultas >= empresaIAEmpresarialLimit {
		return map[string]interface{}{
			"ok": false, "code": "centro_ia_limit_reached",
			"error": "Se alcanzo el limite diario de funciones IA empresariales.",
			"usage": reportesIAUsagePayload(uso.Consultas, empresaIAEmpresarialLimit, 0, 0),
		}
	}
	if payload.AgentID != "general" {
		user := adminEmailFromRequest(r)
		if user == "" {
			user = googleAccountFromRequest(r)
		}
		if _, _, err := reserveAgenteInternetLightUsage(dbEmp, dbSuper, payload.EmpresaID, user); err != nil {
			return map[string]interface{}{"ok": false, "code": "empresa_agent_limit_reached", "error": err.Error()}
		}
	}
	consulta := strings.TrimSpace(payload.Consulta)
	if consulta == "" {
		consulta = defaultEmpresaIAConsulta(payload.Accion)
	}
	raw, _ := json.Marshal(snapshot)
	system := "Eres un agente IA ERP y contable para Powerful Control System. Responde en espanol profesional y accionable. Usa solo el snapshot real filtrado por empresa_id; no inventes cifras, NIT, estados DIAN ni registros. No ejecutes mutaciones: si el usuario pide factura, pago, cliente o producto, entrega un borrador revisable, datos faltantes y siguiente boton/ruta sugerida. No digas que emitiste, registraste o conciliaste algo. Entrega secciones cortas: diagnostico, hallazgos, riesgos, siguiente accion y datos faltantes. Si la accion es borrador_factura, usa formato JSON visible con cliente, items, impuestos sugeridos, faltantes y advertencias.\n\nACCION_SOLICITADA: " + payload.Accion + "\nSNAPSHOT_REAL_JSON:\n" + truncateText(string(raw), 9000)
	system += "\n\n" + buildEmpresaAIChatAgentInstruction(payload.AgentID)
	ctrl := &EmpresaAIChatController{dbEmp: dbEmp, dbSuper: dbSuper, client: &http.Client{Timeout: 45 * time.Second}}
	respuesta, pt, ct, err := ctrl.generateResponseWithSystemPrompt(model, consulta, nil, system)
	if err != nil {
		return map[string]interface{}{"ok": false, "code": "ai_error", "error": "No se pudo ejecutar la funcion IA: " + err.Error()}
	}
	respuesta = strings.TrimSpace(respuesta)
	if _, err := dbpkg.RegisterEmpresaAIConsulta(dbEmp, dbpkg.EmpresaAIConsulta{
		EmpresaID: payload.EmpresaID, Provider: model.Provider, ModelID: empresaIAEmpresarialUsageID,
		Pregunta: consulta, Respuesta: respuesta, PromptTokens: pt, CompletionTokens: ct, TotalTokens: pt + ct,
		FechaConsulta: time.Now().Format("2006-01-02 15:04:05"), PlanActual: strings.TrimSpace(uso.PlanActual),
		UsuarioCreador: adminEmailFromRequest(r), Estado: "activo", Observaciones: "centro_ia_empresarial:" + payload.Accion + " agente=" + payload.AgentID,
	}); err != nil {
		return map[string]interface{}{"ok": false, "code": "usage_register_error", "error": "No se pudo registrar uso IA."}
	}
	usoNuevo, _ := dbpkg.GetEmpresaAIUsoDiario(dbEmp, payload.EmpresaID, model.Provider, empresaIAEmpresarialUsageID, fechaUso)
	return map[string]interface{}{
		"ok":        true,
		"modelo":    "openai:gpt-5.4-mini",
		"agent_id":  payload.AgentID,
		"respuesta": respuesta,
		"usage":     reportesIAUsagePayload(usoNuevo.Consultas, empresaIAEmpresarialLimit, pt, ct),
	}
}

func defaultEmpresaIAConsulta(accion string) string {
	switch normalizeEmpresaIAAccion(accion) {
	case "borrador_factura":
		return "Prepara un borrador revisable de factura o cotizacion con datos disponibles, datos faltantes y advertencias antes de emitir."
	case "cobranza_pagos":
		return "Analiza cobranza y pagos del periodo, prioriza acciones de recaudo y conciliacion."
	case "inventario_productos":
		return "Analiza productos, servicios y stock para sugerir reposicion, limpieza de catalogo y controles."
	case "conciliacion_bancaria":
		return "Compara ventas cerradas contra ingresos financieros y sugiere pasos de conciliacion bancaria."
	case "compras_gastos":
		return "Revisa compras y gastos del periodo, riesgos de duplicados y controles antes de contabilizar."
	case "cumplimiento_dian":
		return "Revisa riesgos de cumplimiento DIAN, impuestos y cierre contable con los datos reales disponibles."
	default:
		return "Haz un diagnostico ejecutivo de la empresa y prioriza acciones operativas, financieras y contables."
	}
}

func absFloat64(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
