package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	dbpkg "github.com/you/pos-backend/db"
)

type iaPedidoAccion struct {
	Tipo         string  `json:"tipo"`
	EstacionID   int64   `json:"estacion_id"`
	ProductoID   int64   `json:"producto_id"`
	ProductoHint string  `json:"producto_hint"`
	Cantidad     float64 `json:"cantidad"`
}

type iaPedidoModelOut struct {
	Acciones         []iaPedidoAccion `json:"acciones"`
	RespuestaNatural string           `json:"respuesta_natural"`
}

type estacionConfigNombre struct {
	ID     int64  `json:"id"`
	Nombre string `json:"nombre"`
}

func stripModelJSONFence(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSpace(s)
		if strings.HasPrefix(strings.ToLower(s), "json") {
			s = strings.TrimSpace(s[4:])
		}
		if i := strings.LastIndex(s, "```"); i >= 0 {
			s = strings.TrimSpace(s[:i])
		}
	}
	return strings.TrimSpace(s)
}

func parseEstacionesNombresFromPref(valor string) []estacionConfigNombre {
	valor = strings.TrimSpace(valor)
	if valor == "" {
		return nil
	}
	raw := valor
	for i := 0; i < 3; i++ {
		var next interface{}
		if err := json.Unmarshal([]byte(raw), &next); err != nil {
			return nil
		}
		if s, ok := next.(string); ok {
			raw = strings.TrimSpace(s)
			continue
		}
		b, err := json.Marshal(next)
		if err != nil {
			return nil
		}
		var cfg struct {
			Estaciones []struct {
				ID     interface{} `json:"id"`
				Nombre string      `json:"nombre"`
			} `json:"estaciones"`
		}
		if err := json.Unmarshal(b, &cfg); err != nil {
			return nil
		}
		out := make([]estacionConfigNombre, 0, len(cfg.Estaciones))
		for _, e := range cfg.Estaciones {
			id := int64(0)
			switch v := e.ID.(type) {
			case float64:
				id = int64(v)
			case string:
				id, _ = strconv.ParseInt(strings.TrimSpace(v), 10, 64)
			}
			nombre := strings.TrimSpace(e.Nombre)
			if id > 0 {
				out = append(out, estacionConfigNombre{ID: id, Nombre: nombre})
			}
		}
		return out
	}
	return nil
}

func stationNameByID(rows []estacionConfigNombre, id int64) string {
	for _, r := range rows {
		if r.ID == id {
			if strings.TrimSpace(r.Nombre) != "" {
				return r.Nombre
			}
			break
		}
	}
	return fmt.Sprintf("Estación %d", id)
}

func resolveProductoForPedidoIA(empresaID int64, productos []dbpkg.Producto, productoID int64, hint string) (*dbpkg.Producto, string) {
	hint = strings.TrimSpace(hint)
	if productoID > 0 {
		for i := range productos {
			if productos[i].ID == productoID {
				return &productos[i], ""
			}
		}
		return nil, "producto_id no encontrado en catálogo activo"
	}
	if hint == "" {
		return nil, "falta producto_id o producto_hint"
	}
	hintLower := strings.ToLower(hint)
	tokens := strings.FieldsFunc(hintLower, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	bestIdx := -1
	bestScore := 0
	for i := range productos {
		n := strings.ToLower(productos[i].Nombre)
		score := 0
		if strings.Contains(n, hintLower) {
			score += 5
		}
		for _, t := range tokens {
			if len(t) < 2 {
				continue
			}
			if strings.Contains(n, t) {
				score += 2
			}
		}
		if score > bestScore {
			bestScore = score
			bestIdx = i
		}
	}
	if bestIdx >= 0 && bestScore > 0 {
		return &productos[bestIdx], ""
	}
	return nil, "no se pudo emparejar el producto con el catálogo"
}

func ensureCarritoEstacionParaPedidoIA(dbEmp *sql.DB, empresaID, estacionID int64, nombreEstacion, usuario string) (*dbpkg.CarritoCompra, error) {
	if estacionID <= 0 {
		return nil, fmt.Errorf("estacion_id invalido para pedido")
	}
	codigo := fmt.Sprintf("EST-%d-%d", empresaID, estacionID)
	cart, err := dbpkg.GetCarritoCompraByCodigo(dbEmp, empresaID, codigo)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if cart == nil || errors.Is(err, sql.ErrNoRows) {
		newID, cerr := dbpkg.CreateCarritoCompra(dbEmp, dbpkg.CarritoCompra{
			EmpresaID:         empresaID,
			Codigo:            codigo,
			Nombre:            nombreEstacion,
			CanalVenta:        "mostrador",
			ReferenciaExterna: fmt.Sprintf("estacion:%d", estacionID),
			Observaciones:     "Carrito automático de estación",
			UsuarioCreador:    usuario,
		})
		if cerr != nil {
			return nil, cerr
		}
		cart, err = dbpkg.GetCarritoCompraByID(dbEmp, empresaID, newID)
		if err != nil {
			return nil, err
		}
	}
	if cart == nil {
		return nil, fmt.Errorf("carrito no disponible")
	}

	paid := isCarritoVentaPagada(cart)
	reg := normalizeCarritoRegistroEstado(cart.Estado)
	op := normalizeCarritoOperativoEstado(cart.EstadoCarrito)
	if !paid && reg == "activo" && op == "abierto" {
		return cart, nil
	}

	resetItems := paid
	if err := dbpkg.ActivateCarritoStationSession(dbEmp, empresaID, cart.ID, resetItems); err != nil {
		fresh, ferr := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, cart.ID)
		if ferr == nil && fresh != nil && !isCarritoVentaPagada(fresh) {
			r2 := normalizeCarritoRegistroEstado(fresh.Estado)
			o2 := normalizeCarritoOperativoEstado(fresh.EstadoCarrito)
			if r2 == "activo" && o2 == "abierto" {
				return fresh, nil
			}
		}
		return nil, err
	}
	return dbpkg.GetCarritoCompraByID(dbEmp, empresaID, cart.ID)
}

func buildIAPedidosSystemPrompt(estaciones []estacionConfigNombre, productos []dbpkg.Producto) string {
	estJSON, _ := json.Marshal(estaciones)
	type prodLite struct {
		ID     int64  `json:"id"`
		Nombre string `json:"nombre"`
		SKU    string `json:"sku,omitempty"`
	}
	pl := make([]prodLite, 0, len(productos))
	for _, p := range productos {
		pl = append(pl, prodLite{ID: p.ID, Nombre: p.Nombre, SKU: p.SKU})
	}
	prodJSON, _ := json.Marshal(pl)
	return "Eres un intérprete de pedidos para un POS en español. " +
		"El usuario puede pedir agregar productos a una mesa, habitación o estación (por nombre o número). " +
		"Debes responder SOLO con un JSON válido (sin markdown, sin texto fuera del JSON) con esta forma exacta:\n" +
		`{"acciones":[{"tipo":"agregar_producto","estacion_id":<entero>,"producto_id":<entero o 0>,"producto_hint":"<texto opcional si producto_id es 0>","cantidad":<entero>=1}],"respuesta_natural":"<mensaje breve en español para el usuario>"}` + "\n" +
		"Reglas: estacion_id debe existir en ESTACIONES. producto_id debe ser un id de PRODUCTOS si hay coincidencia clara; si no, usa 0 y llena producto_hint con palabras del producto pedido. " +
		"Si no entiendes el pedido o falta información, devuelve acciones vacías y explica en respuesta_natural.\n" +
		"ESTACIONES (id y nombre):\n" + string(estJSON) + "\n" +
		"PRODUCTOS (id, nombre, sku):\n" + string(prodJSON)
}

func parseIAPedidosModelJSON(raw string) (*iaPedidoModelOut, error) {
	raw = stripModelJSONFence(raw)
	if raw == "" {
		return nil, fmt.Errorf("respuesta vacia del modelo")
	}
	var out iaPedidoModelOut
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// IaPedidosEstacionEjecutarHandler interpreta un mensaje con IA y agrega ítems al carrito de la estación indicada.
func (c *EmpresaAIChatController) IaPedidosEstacionEjecutarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	if !isSuperAIEnabled(c.dbSuper) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":             false,
			"code":           "ai_disabled",
			"error":          "La IA está desactivada desde configuración avanzada.",
			"service_status": superAIServiceStatus(c.dbSuper),
		})
		return
	}

	var body struct {
		EmpresaID int64                  `json:"empresa_id"`
		ModelID   string                 `json:"model_id"`
		Mensaje   string                 `json:"mensaje"`
		Historial []empresaAIChatMensaje `json:"historial"`
		AgentID   string                 `json:"agent_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "JSON invalido"})
		return
	}
	if body.EmpresaID <= 0 {
		if eid, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && eid > 0 {
			body.EmpresaID = eid
		}
	}
	if body.EmpresaID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "empresa_id es obligatorio"})
		return
	}
	googleAccount := googleAccountFromRequest(r)
	if googleAccount == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "No se pudo identificar la cuenta del usuario autenticado"})
		return
	}
	if err := c.ensureEmpresaAccessByAccount(googleAccount, body.EmpresaID); err != nil {
		writeJSON(w, http.StatusForbidden, map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}

	body.ModelID = strings.TrimSpace(body.ModelID)
	if body.ModelID == "" {
		body.ModelID = firstAvailableEmpresaAIModelID(c.dbSuper)
	}
	catalog := availableEmpresaAIModelMap(c.dbSuper)
	if len(catalog) == 0 {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":             false,
			"code":           "ai_models_unavailable",
			"error":          "No hay proveedores IA habilitados para esta empresa.",
			"service_status": superAIServiceStatus(c.dbSuper),
		})
		return
	}
	model, ok := catalog[body.ModelID]
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "model_id no soportado o desactivado"})
		return
	}

	body.Mensaje = strings.TrimSpace(body.Mensaje)
	if body.Mensaje == "" {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "mensaje es obligatorio"})
		return
	}
	if len([]rune(body.Mensaje)) > 2500 {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "mensaje supera el maximo permitido (2500 caracteres)"})
		return
	}
	agentID := normalizeEmpresaAIChatAgentID(body.AgentID)
	if agentID == "general" {
		agentID = "ventas"
	}
	if _, _, err := reserveAgenteInternetLightUsage(c.dbEmp, c.dbSuper, body.EmpresaID, googleAccount); err != nil {
		writeJSON(w, http.StatusTooManyRequests, map[string]interface{}{"ok": false, "code": "empresa_agent_limit_reached", "error": err.Error()})
		return
	}

	fechaUso := time.Now().Format("2006-01-02")
	usoActual, err := dbpkg.GetEmpresaAIUsoDiario(c.dbEmp, body.EmpresaID, model.Provider, model.ID, fechaUso)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "No se pudo consultar uso diario"})
		return
	}
	planActual := strings.ToLower(strings.TrimSpace(usoActual.PlanActual))
	if planActual == "" {
		planActual = "free"
	}
	empresaChatEnabled, _, _, err := getChatIAEmpresaEnabled(c.dbSuper)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "No se pudo consultar configuración de chat IA"})
		return
	}
	if !empresaChatEnabled {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":    false,
			"code":  "ai_empresa_chat_disabled",
			"error": "El chat con IA para empresas está desactivado desde configuración lógica del chat con IA.",
		})
		return
	}
	empresaMaxConsultas, _, _, err := getChatIAEmpresaMaxConsultasDia(c.dbSuper)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "No se pudo consultar configuración de límites IA"})
		return
	}
	effectiveLimit := effectiveDailyLimitBySuperConfig(empresaMaxConsultas, model.FreeDailyLimit)
	if effectiveLimit == 0 {
		writeJSON(w, http.StatusTooManyRequests, map[string]interface{}{
			"ok":    false,
			"code":  "ai_empresa_chat_blocked",
			"error": "El chat con IA para empresas está bloqueado por configuración (límite en 0).",
		})
		return
	}
	if usoActual.Consultas >= effectiveLimit {
		c.writeLimitReached(w, body.EmpresaID, model, usoActual.Consultas)
		return
	}

	pref, err := dbpkg.GetEmpresaEstacionPref(c.dbEmp, body.EmpresaID, 0, "estaciones_config")
	if err != nil || pref == nil || strings.TrimSpace(pref.Valor) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"ok":    false,
			"error": "No hay configuración de estaciones guardada para esta empresa",
		})
		return
	}
	estaciones := parseEstacionesNombresFromPref(pref.Valor)
	if len(estaciones) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"ok":    false,
			"error": "No se pudieron leer nombres de estaciones desde la configuración",
		})
		return
	}
	productos, err := dbpkg.GetProductosByEmpresa(c.dbEmp, body.EmpresaID, "", "activo", 0, 0, 320, 0)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "No se pudo cargar catálogo de productos"})
		return
	}
	if len(productos) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"ok":    false,
			"error": "No hay productos activos para sugerir al modelo",
		})
		return
	}

	systemPrompt := buildIAPedidosSystemPrompt(estaciones, productos)
	systemPrompt += "\n\n" + buildEmpresaAIChatAgentInstruction(agentID)
	respText, promptTokens, completionTokens, err := c.generateResponseWithSystemPrompt(model, body.Mensaje, body.Historial, systemPrompt)
	if err != nil {
		if isProviderLimitError(err) {
			c.writeLimitReached(w, body.EmpresaID, model, usoActual.Consultas)
			return
		}
		writeJSON(w, http.StatusBadGateway, map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}
	respText = strings.TrimSpace(respText)
	if respText == "" {
		writeJSON(w, http.StatusBadGateway, map[string]interface{}{"ok": false, "error": "El proveedor no devolvio contenido"})
		return
	}

	parsed, err := parseIAPedidosModelJSON(respText)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]interface{}{
			"ok":          false,
			"error":       "No se pudo interpretar la respuesta del modelo como JSON",
			"raw_model":   truncateText(respText, 800),
			"parse_error": err.Error(),
		})
		return
	}

	validStation := make(map[int64]bool)
	for _, e := range estaciones {
		validStation[e.ID] = true
	}

	usuario := strings.TrimSpace(adminEmailFromRequest(r))
	if usuario == "" {
		usuario = googleAccount
	}

	resultados := make([]map[string]interface{}, 0, len(parsed.Acciones))
	tuvoExito := false
	for _, ac := range parsed.Acciones {
		if strings.TrimSpace(strings.ToLower(ac.Tipo)) != "" && strings.TrimSpace(strings.ToLower(ac.Tipo)) != "agregar_producto" {
			continue
		}
		if ac.EstacionID <= 0 || !validStation[ac.EstacionID] {
			resultados = append(resultados, map[string]interface{}{
				"ok":          false,
				"estacion_id": ac.EstacionID,
				"error":       "estacion_id no válido para esta empresa",
			})
			continue
		}
		qty := int64(ac.Cantidad)
		if qty < 1 {
			qty = 1
		}
		if qty > 99 {
			qty = 99
		}
		prod, perr := resolveProductoForPedidoIA(body.EmpresaID, productos, ac.ProductoID, ac.ProductoHint)
		if perr != "" || prod == nil {
			resultados = append(resultados, map[string]interface{}{
				"ok":          false,
				"estacion_id": ac.EstacionID,
				"error":       perr,
			})
			continue
		}
		nombreEst := stationNameByID(estaciones, ac.EstacionID)
		cart, cerr := ensureCarritoEstacionParaPedidoIA(c.dbEmp, body.EmpresaID, ac.EstacionID, nombreEst, usuario)
		if cerr != nil || cart == nil {
			resultados = append(resultados, map[string]interface{}{
				"ok":          false,
				"estacion_id": ac.EstacionID,
				"error":       fmt.Sprintf("carrito: %v", cerr),
			})
			continue
		}
		codigoItem := strings.TrimSpace(prod.SKU)
		if codigoItem == "" {
			codigoItem = strings.TrimSpace(prod.CodigoBarras)
		}
		unidad := strings.TrimSpace(prod.UnidadMedida)
		if unidad == "" {
			unidad = "unidad"
		}
		item := dbpkg.CarritoCompraItem{
			EmpresaID:           body.EmpresaID,
			CarritoID:           cart.ID,
			TipoItem:            "producto",
			ReferenciaID:        prod.ID,
			CodigoItem:          codigoItem,
			Descripcion:         prod.Nombre,
			UnidadMedida:        unidad,
			Cantidad:            float64(qty),
			PrecioUnitario:      prod.Precio,
			DescuentoPorcentaje: 0,
			ImpuestoPorcentaje:  prod.ImpuestoPorcentaje,
			ImpuestoCodigo:      "IVA",
			Observaciones:       "Agregado vía IA pedidos estación",
			UsuarioCreador:      usuario,
		}
		if err := validateCarritoItemPayload(item); err != nil {
			resultados = append(resultados, map[string]interface{}{
				"ok":          false,
				"estacion_id": ac.EstacionID,
				"error":       err.Error(),
			})
			continue
		}
		itemID, ierr := dbpkg.CreateCarritoCompraItem(c.dbEmp, item)
		if ierr != nil {
			resultados = append(resultados, map[string]interface{}{
				"ok":          false,
				"estacion_id": ac.EstacionID,
				"producto_id": prod.ID,
				"producto":    prod.Nombre,
				"error":       ierr.Error(),
			})
			continue
		}
		tuvoExito = true
		resultados = append(resultados, map[string]interface{}{
			"ok":           true,
			"estacion_id":  ac.EstacionID,
			"carrito_id":   cart.ID,
			"item_id":      itemID,
			"producto_id":  prod.ID,
			"producto":     prod.Nombre,
			"cantidad":     qty,
			"estacion_nom": nombreEst,
		})
	}

	auditResp := respText
	if len(auditResp) > 8000 {
		auditResp = string([]rune(auditResp)[:8000])
	}
	if err := dbpkg.UpsertEmpresaAIModeloPreferido(c.dbEmp, body.EmpresaID, googleAccount, model.ID, googleAccount); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "No se pudo registrar modelo preferido"})
		return
	}
	_, err = dbpkg.RegisterEmpresaAIConsulta(c.dbEmp, dbpkg.EmpresaAIConsulta{
		EmpresaID:        body.EmpresaID,
		Provider:         model.Provider,
		ModelID:          model.ID,
		Pregunta:         body.Mensaje,
		Respuesta:        auditResp,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
		FechaConsulta:    time.Now().Format("2006-01-02 15:04:05"),
		PlanActual:       planActual,
		UsuarioCreador:   googleAccount,
		Estado:           "activo",
		Observaciones:    "ia_pedidos_estacion agente=" + agentID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "No se pudo registrar auditoria de consulta"})
		return
	}

	usoActualizado, err := dbpkg.GetEmpresaAIUsoDiario(c.dbEmp, body.EmpresaID, model.Provider, model.ID, fechaUso)
	if err != nil {
		usoActualizado = usoActual
	}
	restante := effectiveLimit - usoActualizado.Consultas
	if restante < 0 {
		restante = 0
	}

	natural := strings.TrimSpace(parsed.RespuestaNatural)
	if natural == "" {
		natural = "Pedido interpretado."
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                  tuvoExito,
		"empresa_id":          body.EmpresaID,
		"model_id":            model.ID,
		"provider":            model.Provider,
		"agent_id":            agentID,
		"respuesta_natural":   natural,
		"acciones_ejecutadas": resultados,
		"raw_model":           truncateText(respText, 1200),
		"usage": map[string]interface{}{
			"plan":              planActual,
			"daily_used":        usoActualizado.Consultas,
			"daily_limit":       effectiveLimit,
			"daily_remaining":   restante,
			"prompt_tokens":     promptTokens,
			"completion_tokens": completionTokens,
		},
	})
}
