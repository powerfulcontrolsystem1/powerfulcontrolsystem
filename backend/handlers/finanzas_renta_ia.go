package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	finanzasRentaIAModelUsageID = "openai:gpt-5.4-mini:renta_ia"
	finanzasRentaIADailyLimit   = 8
)

type empresaFinanzasRentaIAPayload struct {
	dbpkg.EmpresaFinanzasRentaInputs
	UsarIA   bool   `json:"usar_ia"`
	Pregunta string `json:"pregunta"`
}

// EmpresaFinanzasRentaIAHandler calcula una estimacion de renta por empresa y
// opcionalmente genera una explicacion con IA basada solo en el resultado.
func EmpresaFinanzasRentaIAHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}

		payload, err := parseEmpresaFinanzasRentaIAPayload(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if payload.EmpresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}

		resultado, err := dbpkg.GetEmpresaFinanzasRentaEstimacion(dbEmp, payload.EmpresaFinanzasRentaInputs)
		if err != nil {
			http.Error(w, "No se pudo calcular la renta estimada", http.StatusInternalServerError)
			return
		}

		resp := map[string]interface{}{
			"ok":        true,
			"resultado": resultado,
		}
		if payload.UsarIA {
			resp["ia"] = buildEmpresaFinanzasRentaIAResponse(r, dbEmp, dbSuper, payload, resultado)
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

func parseEmpresaFinanzasRentaIAPayload(r *http.Request) (empresaFinanzasRentaIAPayload, error) {
	var payload empresaFinanzasRentaIAPayload
	if r.Method == http.MethodPost {
		if r.Body != nil {
			defer r.Body.Close()
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && err != io.EOF {
				return payload, fmt.Errorf("JSON invalido")
			}
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
	if strings.TrimSpace(payload.Desde) == "" {
		payload.Desde = strings.TrimSpace(q.Get("desde"))
	}
	if strings.TrimSpace(payload.Hasta) == "" {
		payload.Hasta = strings.TrimSpace(q.Get("hasta"))
	}
	if payload.TarifaRenta <= 0 {
		payload.TarifaRenta = rentaFloatQuery(q.Get("tarifa_renta"))
	}
	if payload.IngresosNoConstitutivos <= 0 {
		payload.IngresosNoConstitutivos = rentaFloatQuery(q.Get("ingresos_no_constitutivos"))
	}
	if payload.RentasExentas <= 0 {
		payload.RentasExentas = rentaFloatQuery(q.Get("rentas_exentas"))
	}
	if payload.DeduccionesAdicionales <= 0 {
		payload.DeduccionesAdicionales = rentaFloatQuery(q.Get("deducciones_adicionales"))
	}
	if payload.DescuentosTributarios <= 0 {
		payload.DescuentosTributarios = rentaFloatQuery(q.Get("descuentos_tributarios"))
	}
	if payload.AnticipoRenta <= 0 {
		payload.AnticipoRenta = rentaFloatQuery(q.Get("anticipo_renta"))
	}
	if payload.RetencionesAdicionales <= 0 {
		payload.RetencionesAdicionales = rentaFloatQuery(q.Get("retenciones_adicionales"))
	}
	if payload.SobretasaPuntos <= 0 {
		payload.SobretasaPuntos = rentaFloatQuery(q.Get("sobretasa_puntos"))
	}
	payload.UsarVentasPOSComoIngreso = boolPayloadOrQuery(payload.UsarVentasPOSComoIngreso, q.Get("usar_ventas_pos_como_ingreso"))
	payload.UsarMovimientosComoIngreso = boolPayloadOrQuery(payload.UsarMovimientosComoIngreso, q.Get("usar_movimientos_como_ingreso"))
	payload.UsarComprasYNominaEgreso = boolPayloadOrQuery(payload.UsarComprasYNominaEgreso, q.Get("usar_compras_y_nomina_egreso"))
	payload.UsarMovimientosComoEgreso = boolPayloadOrQuery(payload.UsarMovimientosComoEgreso, q.Get("usar_movimientos_como_egreso"))
	payload.UsarIA = payload.UsarIA || queryBool(r, "usar_ia") || queryBool(r, "ia")
	if strings.TrimSpace(payload.Pregunta) == "" {
		payload.Pregunta = strings.TrimSpace(q.Get("pregunta"))
	}
	payload.EmpresaFinanzasRentaInputs = dbpkg.NormalizeEmpresaFinanzasRentaInputs(payload.EmpresaFinanzasRentaInputs)
	return payload, nil
}

func rentaFloatQuery(raw string) float64 {
	raw = strings.TrimSpace(strings.ReplaceAll(raw, ",", "."))
	if raw == "" {
		return 0
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0
	}
	return v
}

func boolPayloadOrQuery(payloadValue bool, raw string) bool {
	if payloadValue {
		return true
	}
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return false
	}
	return raw == "1" || raw == "true" || raw == "si" || raw == "sí" || raw == "yes" || raw == "on"
}

func buildEmpresaFinanzasRentaIAResponse(r *http.Request, dbEmp, dbSuper *sql.DB, payload empresaFinanzasRentaIAPayload, resultado dbpkg.EmpresaFinanzasRentaResultado) map[string]interface{} {
	if !isSuperAIEnabled(dbSuper) {
		return map[string]interface{}{
			"ok": false, "code": "ai_disabled",
			"error": "La IA esta desactivada desde super administrador; el calculo numerico queda disponible.",
		}
	}
	empresaChatEnabled, _, _, err := getChatIAEmpresaEnabled(dbSuper)
	if err != nil || !empresaChatEnabled {
		return map[string]interface{}{
			"ok": false, "code": "ai_empresa_disabled",
			"error": "El chat IA empresarial esta desactivado; el calculo numerico queda disponible.",
		}
	}
	model, ok := availableEmpresaAIModelMap(dbSuper)["openai:gpt-5.4-mini"]
	if !ok {
		return map[string]interface{}{
			"ok": false, "code": "ai_model_missing",
			"error": "GPT-5.4 mini no esta disponible; el calculo numerico queda disponible.",
		}
	}
	fechaUso := time.Now().Format("2006-01-02")
	uso, err := dbpkg.GetEmpresaAIUsoDiario(dbEmp, payload.EmpresaID, model.Provider, finanzasRentaIAModelUsageID, fechaUso)
	if err != nil {
		return map[string]interface{}{"ok": false, "code": "usage_error", "error": "No se pudo consultar uso diario de IA."}
	}
	if uso.Consultas >= finanzasRentaIADailyLimit {
		return map[string]interface{}{
			"ok": false, "code": "renta_ia_limit_reached",
			"error": "Se alcanzo el limite diario de analisis de renta IA para esta empresa.",
			"usage": reportesIAUsagePayload(uso.Consultas, finanzasRentaIADailyLimit, 0, 0),
		}
	}

	pregunta := strings.TrimSpace(payload.Pregunta)
	if pregunta == "" {
		pregunta = "Analiza esta estimacion de impuesto de renta de la empresa, explica la base usada, riesgos de doble conteo, saldo estimado y siguientes pasos para contador."
	}
	if len([]rune(pregunta)) > 1500 {
		pregunta = string([]rune(pregunta)[:1500])
	}
	raw, _ := json.Marshal(resultado)
	system := "Eres un contador tributario y analista financiero del sistema POS multiempresa. Responde en espanol claro y profesional. Usa solamente el JSON entregado por el backend; no inventes cifras ni asumas beneficios tributarios. Explica que es una estimacion gerencial, no declaracion oficial. Si hay alertas, priorizalas. Entrega: resumen ejecutivo, calculo clave, riesgos, recomendaciones y datos que debe validar el contador.\n\nRENTA_ESTIMADA_JSON:\n" + truncateText(string(raw), 9000)
	ctrl := &EmpresaAIChatController{dbEmp: dbEmp, dbSuper: dbSuper, client: &http.Client{Timeout: 45 * time.Second}}
	respuesta, pt, ct, err := ctrl.generateResponseWithSystemPrompt(model, pregunta, nil, system)
	if err != nil {
		return map[string]interface{}{"ok": false, "code": "ai_error", "error": "No se pudo generar analisis IA: " + err.Error()}
	}
	respuesta = strings.TrimSpace(respuesta)
	if _, err := dbpkg.RegisterEmpresaAIConsulta(dbEmp, dbpkg.EmpresaAIConsulta{
		EmpresaID: payload.EmpresaID, Provider: model.Provider, ModelID: finanzasRentaIAModelUsageID,
		Pregunta: pregunta, Respuesta: respuesta, PromptTokens: pt, CompletionTokens: ct, TotalTokens: pt + ct,
		FechaConsulta: time.Now().Format("2006-01-02 15:04:05"), PlanActual: strings.TrimSpace(uso.PlanActual),
		UsuarioCreador: adminEmailFromRequest(r), Estado: "activo", Observaciones: "finanzas_renta_ia",
	}); err != nil {
		return map[string]interface{}{"ok": false, "code": "usage_register_error", "error": "No se pudo registrar uso IA."}
	}
	usoNuevo, _ := dbpkg.GetEmpresaAIUsoDiario(dbEmp, payload.EmpresaID, model.Provider, finanzasRentaIAModelUsageID, fechaUso)
	return map[string]interface{}{
		"ok":        true,
		"modelo":    "openai:gpt-5.4-mini",
		"respuesta": respuesta,
		"usage":     reportesIAUsagePayload(usoNuevo.Consultas, finanzasRentaIADailyLimit, pt, ct),
	}
}
