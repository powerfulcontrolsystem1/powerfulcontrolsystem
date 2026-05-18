package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type empresaConfiguracionGuiadaState struct {
	EmpresaID              int64                                 `json:"empresa_id"`
	EmpresaNombre          string                                `json:"empresa_nombre,omitempty"`
	TipoEmpresaID          int64                                 `json:"tipo_empresa_id,omitempty"`
	TipoEmpresaNombre      string                                `json:"tipo_empresa_nombre,omitempty"`
	Operacion              dbpkg.TipoEmpresaPreconfigOperacion   `json:"operacion"`
	Estaciones             dbpkg.TipoEmpresaPreconfigEstaciones  `json:"estaciones"`
	Asistente              dbpkg.TipoEmpresaPreconfigAsistenteIA `json:"asistente_ia"`
	Advanced               *dbpkg.EmpresaConfiguracionAvanzada   `json:"advanced,omitempty"`
	Operativa              *dbpkg.EmpresaConfiguracionOperativa  `json:"operativa,omitempty"`
	ResumenAnterior        map[string]interface{}                `json:"resumen_anterior,omitempty"`
	PendientesAnteriores   []string                              `json:"pendientes_anteriores,omitempty"`
	ConfiguradaAnterior    bool                                  `json:"configurada_anterior"`
	FechaConfiguracionPrev string                                `json:"fecha_configuracion_prev,omitempty"`
}

type empresaConfiguracionGuiadaQuestion struct {
	ID           string   `json:"id"`
	Label        string   `json:"label"`
	Prompt       string   `json:"prompt"`
	Type         string   `json:"type"`
	Required     bool     `json:"required"`
	Placeholder  string   `json:"placeholder,omitempty"`
	Help         string   `json:"help,omitempty"`
	DefaultValue string   `json:"default_value,omitempty"`
	Options      []string `json:"options,omitempty"`
}

func EmpresaConfiguracionGuiadaHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			state, err := loadEmpresaConfiguracionGuiadaState(dbEmp, empresaID)
			if err != nil {
				http.Error(w, "no se pudo cargar la configuracion guiada: "+err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":         true,
				"estado":     state,
				"wizard":     buildEmpresaConfiguracionGuiadaWizard(state),
				"resumen":    state.ResumenAnterior,
				"pendientes": state.PendientesAnteriores,
			})
			return

		case http.MethodPost:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "" {
				action = "aplicar"
			}
			if action != "aplicar" {
				http.Error(w, "accion no soportada", http.StatusBadRequest)
				return
			}

			var payload struct {
				Answers map[string]interface{} `json:"answers"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload invalido", http.StatusBadRequest)
				return
			}
			state, err := loadEmpresaConfiguracionGuiadaState(dbEmp, empresaID)
			if err != nil {
				http.Error(w, "no se pudo cargar el contexto guiado: "+err.Error(), http.StatusInternalServerError)
				return
			}
			result, err := applyEmpresaConfiguracionGuiada(dbEmp, state, payload.Answers, strings.TrimSpace(adminEmailFromRequest(r)))
			if err != nil {
				http.Error(w, "no se pudo aplicar la configuracion guiada: "+err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":        true,
				"resultado": result,
			})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

func loadEmpresaConfiguracionGuiadaState(dbEmp *sql.DB, empresaID int64) (*empresaConfiguracionGuiadaState, error) {
	state := &empresaConfiguracionGuiadaState{EmpresaID: empresaID}

	var tipoID int64
	var tipoNombre string
	if err := dbEmp.QueryRow(`SELECT COALESCE(nombre, ''), COALESCE(tipo_id, 0), COALESCE(tipo_nombre, '') FROM empresas WHERE id = ? LIMIT 1`, empresaID).Scan(&state.EmpresaNombre, &tipoID, &tipoNombre); err != nil {
		return nil, err
	}
	state.TipoEmpresaID = tipoID
	state.TipoEmpresaNombre = strings.TrimSpace(tipoNombre)

	if cfg, err := dbpkg.GetEmpresaConfiguracionAvanzada(dbEmp, empresaID); err == nil {
		state.Advanced = cfg
	}
	if op, err := dbpkg.GetEmpresaConfiguracionOperativa(dbEmp, empresaID); err == nil {
		state.Operativa = op
	}

	if pref, err := dbpkg.GetEmpresaEstacionPref(dbEmp, empresaID, 0, "preconfiguracion_tipo_empresa_asistente_ia"); err == nil && pref != nil && strings.TrimSpace(pref.Valor) != "" {
		var payload struct {
			TipoEmpresaID     int64                                 `json:"tipo_empresa_id"`
			TipoEmpresaNombre string                                `json:"tipo_empresa_nombre"`
			Operacion         dbpkg.TipoEmpresaPreconfigOperacion   `json:"operacion"`
			Estaciones        dbpkg.TipoEmpresaPreconfigEstaciones  `json:"estaciones"`
			AsistenteIA       dbpkg.TipoEmpresaPreconfigAsistenteIA `json:"asistente_ia"`
		}
		if json.Unmarshal([]byte(pref.Valor), &payload) == nil {
			if payload.TipoEmpresaID > 0 {
				state.TipoEmpresaID = payload.TipoEmpresaID
			}
			if strings.TrimSpace(payload.TipoEmpresaNombre) != "" {
				state.TipoEmpresaNombre = strings.TrimSpace(payload.TipoEmpresaNombre)
			}
			state.Operacion = payload.Operacion
			state.Estaciones = payload.Estaciones
			state.Asistente = payload.AsistenteIA
		}
	}

	if strings.TrimSpace(state.Operacion.NombreEstacionSingular) == "" {
		state.Operacion.NombreEstacionSingular = defaultGuidedStationSingular(state.TipoEmpresaNombre)
	}
	if strings.TrimSpace(state.Operacion.NombreEstacionPlural) == "" {
		state.Operacion.NombreEstacionPlural = defaultGuidedStationPlural(state.Operacion.NombreEstacionSingular)
	}
	if strings.TrimSpace(state.Estaciones.Prefijo) == "" {
		state.Estaciones.Prefijo = state.Operacion.NombreEstacionSingular
	}

	if pref, err := dbpkg.GetEmpresaEstacionPref(dbEmp, empresaID, 0, "configuracion_guiada_resumen"); err == nil && pref != nil && strings.TrimSpace(pref.Valor) != "" {
		var meta map[string]interface{}
		if json.Unmarshal([]byte(pref.Valor), &meta) == nil {
			state.ResumenAnterior = meta
			state.ConfiguradaAnterior = true
			if ts, _ := meta["aplicado_en"].(string); strings.TrimSpace(ts) != "" {
				state.FechaConfiguracionPrev = ts
			}
			if pending, ok := meta["pendientes"].([]interface{}); ok {
				state.PendientesAnteriores = make([]string, 0, len(pending))
				for _, item := range pending {
					text := strings.TrimSpace(fmt.Sprintf("%v", item))
					if text != "" {
						state.PendientesAnteriores = append(state.PendientesAnteriores, text)
					}
				}
			}
		}
	}

	return state, nil
}

func buildEmpresaConfiguracionGuiadaWizard(state *empresaConfiguracionGuiadaState) map[string]interface{} {
	singular := strings.TrimSpace(state.Operacion.NombreEstacionSingular)
	if singular == "" {
		singular = "estación"
	}
	plural := strings.TrimSpace(state.Operacion.NombreEstacionPlural)
	if plural == "" {
		plural = defaultGuidedStationPlural(singular)
	}
	kind := strings.ToLower(strings.TrimSpace(state.Operacion.TipoNegocio))
	if kind == "" {
		kind = strings.ToLower(strings.TrimSpace(state.TipoEmpresaNombre))
	}
	questions := []empresaConfiguracionGuiadaQuestion{
		{
			ID:           "nombre_comercial",
			Label:        "Nombre comercial",
			Prompt:       "¿Cómo quieres que aparezca el nombre comercial de esta empresa dentro del sistema y en sus documentos internos?",
			Type:         "text",
			Required:     true,
			Placeholder:  "Ej: Restaurante Pepita",
			DefaultValue: firstNonEmptyGuidedValue(valueOrEmpty(state.Advanced, func(cfg *dbpkg.EmpresaConfiguracionAvanzada) string { return cfg.NombreComercial }), state.EmpresaNombre),
		},
		{
			ID:           "cantidad_estaciones",
			Label:        "Cantidad operativa",
			Prompt:       fmt.Sprintf("¿Cuántas %s operativas quieres dejar creadas desde ahora?", strings.ToLower(plural)),
			Type:         "number",
			Required:     state.Operacion.UsaEstaciones || state.Estaciones.Enabled,
			Placeholder:  "Ej: 12",
			DefaultValue: strconv.Itoa(maxIntGuided(state.Estaciones.Cantidad, inferGuidedStationCountFromState(state))),
			Help:         fmt.Sprintf("El robot usará este número para crear o ajustar %s con un nombre consistente.", strings.ToLower(plural)),
		},
		{
			ID:           "prefijo_estaciones",
			Label:        "Nombre base de estaciones",
			Prompt:       fmt.Sprintf("¿Qué nombre base quieres usar para cada %s? Ejemplo: %s 1, %s 2.", strings.ToLower(singular), singular, singular),
			Type:         "text",
			Required:     state.Operacion.UsaEstaciones || state.Estaciones.Enabled,
			Placeholder:  singular,
			DefaultValue: firstNonEmptyGuidedValue(state.Estaciones.Prefijo, singular),
		},
		{
			ID:           "venta_directa",
			Label:        "Venta directa",
			Prompt:       "¿Quieres dejar activa la venta directa para operar sin entrar siempre por estaciones?",
			Type:         "boolean",
			Required:     true,
			DefaultValue: guidedBoolString(state.Operacion.VentaDirectaEnabled),
		},
		{
			ID:           "modo_documento_venta",
			Label:        "Documento de venta",
			Prompt:       "¿Qué documento quieres usar por defecto al vender: comprobante de pago o factura electrónica?",
			Type:         "select",
			Required:     true,
			DefaultValue: firstNonEmptyGuidedValue(valueOrEmpty(state.Advanced, func(cfg *dbpkg.EmpresaConfiguracionAvanzada) string { return cfg.ModoDocumentoVenta }), "comprobante_pago"),
			Options:      []string{"comprobante_pago", "factura_electronica"},
		},
		{
			ID:           "imprimir_venta",
			Label:        "Imprimir venta",
			Prompt:       "¿Quieres imprimir automáticamente la venta cuando se cierre el cobro?",
			Type:         "boolean",
			Required:     true,
			DefaultValue: guidedBoolString(boolOrFalseGuided(state.Advanced, func(cfg *dbpkg.EmpresaConfiguracionAvanzada) bool { return cfg.ImprimirVenta })),
		},
		{
			ID:           "imprimir_factura_electronica",
			Label:        "Imprimir factura electrónica",
			Prompt:       "¿Quieres imprimir automáticamente la factura electrónica cuando se emita?",
			Type:         "boolean",
			Required:     true,
			DefaultValue: guidedBoolString(boolOrFalseGuided(state.Advanced, func(cfg *dbpkg.EmpresaConfiguracionAvanzada) bool { return cfg.ImprimirFacturaElectronica })),
		},
	}

	if guidedTypeNeedsKitchen(kind) {
		questions = append(questions, empresaConfiguracionGuiadaQuestion{
			ID:           "usa_impresion_cocina",
			Label:        "Impresión en cocina o despacho",
			Prompt:       "¿Manejas impresión de comandas para cocina, barra o despacho?",
			Type:         "boolean",
			Required:     true,
			DefaultValue: "si",
		})
	}

	if guidedTypeNeedsTips(kind) {
		questions = append(questions, empresaConfiguracionGuiadaQuestion{
			ID:           "habilitar_propinas",
			Label:        "Propinas",
			Prompt:       "¿Quieres dejar activado el manejo de propinas desde el inicio?",
			Type:         "boolean",
			Required:     true,
			DefaultValue: guidedBoolString(boolOrFalseGuided(state.Operativa, func(cfg *dbpkg.EmpresaConfiguracionOperativa) bool { return cfg.HabilitarPropinas })),
		})
	}

	return map[string]interface{}{
		"title":       "Configuración guiada inicial",
		"description": "El robot te hace preguntas concretas y aplica la configuración operativa base de la empresa sin pedirte que recorras todo el sistema a mano.",
		"questions":   questions,
	}
}

func applyEmpresaConfiguracionGuiada(dbEmp *sql.DB, state *empresaConfiguracionGuiadaState, answers map[string]interface{}, usuario string) (map[string]interface{}, error) {
	if state == nil || state.EmpresaID <= 0 {
		return nil, fmt.Errorf("empresa invalida")
	}
	if strings.TrimSpace(usuario) == "" {
		usuario = "sistema.configuracion_guiada"
	}

	nombreComercial := strings.TrimSpace(answerStringGuided(answers["nombre_comercial"]))
	if nombreComercial == "" {
		nombreComercial = firstNonEmptyGuidedValue(valueOrEmpty(state.Advanced, func(cfg *dbpkg.EmpresaConfiguracionAvanzada) string { return cfg.NombreComercial }), state.EmpresaNombre)
	}
	cantidadEstaciones := maxIntGuided(answerIntGuided(answers["cantidad_estaciones"]), inferGuidedStationCountFromState(state))
	prefijoEstaciones := strings.TrimSpace(answerStringGuided(answers["prefijo_estaciones"]))
	if prefijoEstaciones == "" {
		prefijoEstaciones = firstNonEmptyGuidedValue(state.Estaciones.Prefijo, state.Operacion.NombreEstacionSingular, "Estacion")
	}
	ventaDirecta := answerBoolGuided(answers["venta_directa"], state.Operacion.VentaDirectaEnabled)
	modoDocumento := normalizeGuidedDocumentMode(answerStringGuided(answers["modo_documento_venta"]))
	imprimirVenta := answerBoolGuided(answers["imprimir_venta"], boolOrFalseGuided(state.Advanced, func(cfg *dbpkg.EmpresaConfiguracionAvanzada) bool { return cfg.ImprimirVenta }))
	imprimirFE := answerBoolGuided(answers["imprimir_factura_electronica"], boolOrFalseGuided(state.Advanced, func(cfg *dbpkg.EmpresaConfiguracionAvanzada) bool { return cfg.ImprimirFacturaElectronica }))
	usaImpresionCocina := answerBoolGuided(answers["usa_impresion_cocina"], false)
	habilitarPropinas := answerBoolGuided(answers["habilitar_propinas"], boolOrFalseGuided(state.Operativa, func(cfg *dbpkg.EmpresaConfiguracionOperativa) bool { return cfg.HabilitarPropinas }))

	operacion := state.Operacion
	operacion.VentaDirectaEnabled = ventaDirecta
	operacion.UsaEstaciones = cantidadEstaciones > 0
	estaciones := state.Estaciones
	estaciones.Enabled = cantidadEstaciones > 0
	estaciones.Cantidad = cantidadEstaciones
	estaciones.Prefijo = prefijoEstaciones

	if estaciones.Enabled {
		adaptacion := dbpkg.TipoEmpresaPreconfigAdaptacionNucleo{
			FuenteUnica:                        true,
			UsuariosDesdeNucleo:                true,
			ProductosServiciosDesdeNucleo:      true,
			EstacionesComoRecursosConfigurados: true,
			EntidadEstacionSingular:            strings.TrimSpace(defaultString(operacion.NombreEstacionSingular, estaciones.Prefijo)),
			EntidadEstacionPlural:              strings.TrimSpace(defaultString(operacion.NombreEstacionPlural, pluralizeStationName(defaultString(operacion.NombreEstacionSingular, estaciones.Prefijo)))),
			UsuariosOperativos:                 operacion.RolesOperativos,
		}
		rawConfig, estacionesCreadas := buildEmpresaEstacionesPreconfig(estaciones, adaptacion)
		if _, err := dbpkg.UpsertEmpresaEstacionPref(dbEmp, dbpkg.EmpresaEstacionPref{
			EmpresaID:      state.EmpresaID,
			EstacionID:     0,
			Clave:          "estaciones_config",
			Valor:          rawConfig,
			UsuarioCreador: usuario,
			Estado:         "activo",
			Observaciones:  "[configuracion_guiada] estaciones configuradas por asistente",
		}); err != nil {
			return nil, err
		}
		_ = estacionesCreadas
	}

	rawOperacion, _ := json.Marshal(map[string]any{
		"tipo_negocio":             strings.TrimSpace(operacion.TipoNegocio),
		"nombre_estacion_singular": strings.TrimSpace(operacion.NombreEstacionSingular),
		"nombre_estacion_plural":   strings.TrimSpace(operacion.NombreEstacionPlural),
		"usa_estaciones":           operacion.UsaEstaciones,
		"venta_directa_enabled":    operacion.VentaDirectaEnabled,
		"venta_directa_nombre":     strings.TrimSpace(defaultString(operacion.VentaDirectaNombre, "Venta directa")),
		"venta_directa_url":        "/administrar_empresa/carrito_de_compras.html?modo=venta_directa&perm_page=linkVentaDirecta",
		"carrito_rapido_url":       "/administrar_empresa/carrito_de_compras.html?modo=venta_directa&perm_page=linkVentaDirecta",
		"comisiones_enabled":       operacion.ComisionesEnabled,
		"comision_rol":             strings.TrimSpace(operacion.ComisionRol),
		"comision_filtro":          strings.TrimSpace(operacion.ComisionFiltro),
		"comision_porcentaje":      operacion.ComisionPorcentaje,
		"roles_operativos":         operacion.RolesOperativos,
		"configuracion_guiada":     true,
		"fecha_actualizacion":      time.Now().Format(time.RFC3339),
	})
	if _, err := dbpkg.UpsertEmpresaEstacionPref(dbEmp, dbpkg.EmpresaEstacionPref{
		EmpresaID:      state.EmpresaID,
		EstacionID:     0,
		Clave:          "preconfiguracion_tipo_empresa_operacion",
		Valor:          string(rawOperacion),
		UsuarioCreador: usuario,
		Estado:         "activo",
		Observaciones:  "[configuracion_guiada] operación inicial ajustada por asistente",
	}); err != nil {
		return nil, err
	}

	cfgOperativa := dbpkg.EmpresaConfiguracionOperativa{
		EmpresaID:                       state.EmpresaID,
		MetodoPagoEfectivo:              true,
		MetodoPagoTarjetaCredito:        true,
		MetodoPagoTarjetaDebito:         true,
		MetodoPagoTransferenciaBancaria: true,
		MetodoPagoMixto:                 true,
		MetodoPagoCodigoDescuento:       true,
		HabilitarPropinas:               habilitarPropinas,
		HabilitarComisiones:             operacion.ComisionesEnabled,
		UsuarioCreador:                  usuario,
		Estado:                          "activo",
		Observaciones:                   "[configuracion_guiada] configuración operativa base",
	}
	if _, err := dbpkg.UpsertEmpresaConfiguracionOperativa(dbEmp, cfgOperativa); err != nil {
		return nil, err
	}

	cfgAvanzada := dbpkg.EmpresaConfiguracionAvanzada{EmpresaID: state.EmpresaID}
	if state.Advanced != nil {
		cfgAvanzada = *state.Advanced
	}
	cfgAvanzada.EmpresaID = state.EmpresaID
	cfgAvanzada.UsuarioCreador = usuario
	cfgAvanzada.Estado = "activo"
	cfgAvanzada.NombreComercial = nombreComercial
	if strings.TrimSpace(cfgAvanzada.RazonSocial) == "" {
		cfgAvanzada.RazonSocial = nombreComercial
	}
	cfgAvanzada.ModoDocumentoVenta = modoDocumento
	cfgAvanzada.ImprimirVenta = imprimirVenta
	cfgAvanzada.ImprimirFacturaElectronica = imprimirFE
	cfgAvanzada.FacturacionElectronicaActiva = modoDocumento == "factura_electronica"
	cfgAvanzada.FormatoImpresion = firstNonEmptyGuidedValue(cfgAvanzada.FormatoImpresion, "pos")
	cfgAvanzada.Observaciones = strings.TrimSpace(strings.Join([]string{
		strings.TrimSpace(cfgAvanzada.Observaciones),
		"[configuracion_guiada] configuración comercial y documental actualizada",
	}, " | "))
	if _, err := dbpkg.UpsertEmpresaConfiguracionAvanzada(dbEmp, cfgAvanzada); err != nil {
		return nil, err
	}

	pendientes := make([]string, 0)
	if usaImpresionCocina {
		if err := dbpkg.EnsureEmpresaImpresorasSchema(dbEmp); err == nil {
			impresoras, listErr := dbpkg.ListEmpresaImpresorasByEmpresa(dbEmp, state.EmpresaID, false)
			if listErr == nil {
				var cocinaPrinterID int64
				for _, item := range impresoras {
					if strings.Contains(strings.ToLower(strings.TrimSpace(item.AreaOperativa)), "cocina") && strings.TrimSpace(strings.ToLower(item.Estado)) != "inactivo" {
						cocinaPrinterID = item.ID
						break
					}
				}
				if cocinaPrinterID > 0 {
					_, _ = dbpkg.UpsertEmpresaImpresoraFuncionalidad(dbEmp, dbpkg.EmpresaImpresoraFuncionalidad{
						EmpresaID:      state.EmpresaID,
						Funcionalidad:  "cocina",
						ImpresoraID:    cocinaPrinterID,
						Prioridad:      10,
						UsuarioCreador: usuario,
						Estado:         "activo",
						Observaciones:  "[configuracion_guiada] asignación automática de cocina",
					})
				} else {
					pendientes = append(pendientes, "Configurar una impresora activa para la funcionalidad de cocina o despacho.")
				}
			}
		}
	}

	resumen := map[string]interface{}{
		"empresa_id":                   state.EmpresaID,
		"empresa_nombre":               state.EmpresaNombre,
		"tipo_empresa_nombre":          state.TipoEmpresaNombre,
		"nombre_comercial":             nombreComercial,
		"cantidad_estaciones":          cantidadEstaciones,
		"prefijo_estaciones":           prefijoEstaciones,
		"venta_directa":                ventaDirecta,
		"modo_documento_venta":         modoDocumento,
		"imprimir_venta":               imprimirVenta,
		"imprimir_factura_electronica": imprimirFE,
		"usa_impresion_cocina":         usaImpresionCocina,
		"habilitar_propinas":           habilitarPropinas,
		"aplicado_en":                  time.Now().Format(time.RFC3339),
		"pendientes":                   pendientes,
	}
	rawResumen, _ := json.Marshal(resumen)
	_, _ = dbpkg.UpsertEmpresaEstacionPref(dbEmp, dbpkg.EmpresaEstacionPref{
		EmpresaID:      state.EmpresaID,
		EstacionID:     0,
		Clave:          "configuracion_guiada_resumen",
		Valor:          string(rawResumen),
		UsuarioCreador: usuario,
		Estado:         "activo",
		Observaciones:  "[configuracion_guiada] resumen de configuración aplicada",
	})

	return map[string]interface{}{
		"mensaje":    buildConfiguracionGuiadaSuccessMessage(state, cantidadEstaciones, pendientes),
		"resumen":    resumen,
		"pendientes": pendientes,
		"acciones": []map[string]string{
			{"label": "Revisar impresora", "url": "/administrar_empresa/configuracion_impresora.html"},
			{"label": "Revisar estaciones", "url": "/administrar_empresa/configuracion_de_estaciones.html"},
			{"label": "Abrir configuración", "url": "/administrar_empresa/configuracion.html"},
		},
	}, nil
}

func buildConfiguracionGuiadaSuccessMessage(state *empresaConfiguracionGuiadaState, cantidadEstaciones int, pendientes []string) string {
	partes := []string{"Listo. Apliqué la configuración guiada base de la empresa."}
	if cantidadEstaciones > 0 {
		partes = append(partes, fmt.Sprintf("Dejé %d %s preparadas.", cantidadEstaciones, strings.ToLower(defaultGuidedStationPlural(state.Operacion.NombreEstacionSingular))))
	}
	if len(pendientes) > 0 {
		partes = append(partes, "Todavía quedan pendientes controlados: "+strings.Join(pendientes, " "))
	}
	return strings.Join(partes, " ")
}

func firstNonEmptyGuidedValue(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func valueOrEmpty[T any](obj *T, getter func(*T) string) string {
	if obj == nil {
		return ""
	}
	return strings.TrimSpace(getter(obj))
}

func boolOrFalseGuided[T any](obj *T, getter func(*T) bool) bool {
	if obj == nil {
		return false
	}
	return getter(obj)
}

func guidedBoolString(v bool) string {
	if v {
		return "si"
	}
	return "no"
}

func answerStringGuided(v interface{}) string {
	return strings.TrimSpace(fmt.Sprintf("%v", v))
}

func answerIntGuided(v interface{}) int {
	switch vv := v.(type) {
	case float64:
		return int(vv)
	case int:
		return vv
	case int64:
		return int(vv)
	case string:
		n, _ := strconv.Atoi(strings.TrimSpace(vv))
		return n
	default:
		n, _ := strconv.Atoi(strings.TrimSpace(fmt.Sprintf("%v", v)))
		return n
	}
}

func answerBoolGuided(v interface{}, fallback bool) bool {
	raw := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", v)))
	switch raw {
	case "si", "sí", "true", "1", "yes", "y", "on":
		return true
	case "no", "false", "0", "off":
		return false
	default:
		return fallback
	}
}

func normalizeGuidedDocumentMode(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "factura" || raw == "factura_electronica" {
		return "factura_electronica"
	}
	return "comprobante_pago"
}

func maxIntGuided(values ...int) int {
	max := 0
	for _, value := range values {
		if value > max {
			max = value
		}
	}
	if max <= 0 {
		return 1
	}
	if max > 200 {
		return 200
	}
	return max
}

func inferGuidedStationCountFromState(state *empresaConfiguracionGuiadaState) int {
	if state == nil {
		return 1
	}
	if state.Estaciones.Cantidad > 0 {
		return state.Estaciones.Cantidad
	}
	switch {
	case guidedTypeContains(state.TipoEmpresaNombre, "hotel", "hostal", "hospedaje"):
		return 12
	case guidedTypeContains(state.TipoEmpresaNombre, "motel"):
		return 10
	case guidedTypeContains(state.TipoEmpresaNombre, "restaurante", "restaurant", "bar", "cafeteria", "cafetería"):
		return 8
	case guidedTypeContains(state.TipoEmpresaNombre, "salon", "salón", "belleza", "spa", "barberia", "barbería"):
		return 6
	default:
		return 4
	}
}

func defaultGuidedStationSingular(tipo string) string {
	switch {
	case guidedTypeContains(tipo, "hotel", "hostal", "hospedaje", "motel"):
		return "Estacion"
	case guidedTypeContains(tipo, "restaurante", "restaurant", "bar", "cafeteria", "cafetería"):
		return "Mesa"
	case guidedTypeContains(tipo, "salon", "salón", "belleza", "spa", "barberia", "barbería"):
		return "Silla"
	case guidedTypeContains(tipo, "lavadero", "autolavado", "bahia", "bahía"):
		return "Bahia"
	default:
		return "Estacion"
	}
}

func defaultGuidedStationPlural(singular string) string {
	singular = strings.TrimSpace(singular)
	if singular == "" {
		return "Estaciones"
	}
	lower := strings.ToLower(singular)
	if strings.HasSuffix(lower, "ion") {
		return singular + "es"
	}
	if strings.HasSuffix(lower, "s") {
		return singular
	}
	return singular + "s"
}

func guidedTypeNeedsKitchen(raw string) bool {
	return guidedTypeContains(raw, "restaurante", "restaurant", "bar", "cafeteria", "cafetería", "panaderia", "panadería")
}

func guidedTypeNeedsTips(raw string) bool {
	return guidedTypeContains(raw, "restaurante", "restaurant", "bar", "salon", "salón", "belleza", "spa", "barberia", "barbería", "hotel", "motel")
}

func guidedTypeContains(raw string, tokens ...string) bool {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return false
	}
	for _, token := range tokens {
		if strings.Contains(normalized, strings.ToLower(strings.TrimSpace(token))) {
			return true
		}
	}
	return false
}

func sortedUniqueOptionsGuided(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, item := range values {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		key := strings.ToLower(item)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}
