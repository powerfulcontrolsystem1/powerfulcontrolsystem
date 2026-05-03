package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type qaStep struct {
	Module string `json:"module"`
	Action string `json:"action"`
	Method string `json:"method"`
	Path   string `json:"path"`
	Status int    `json:"status"`
	OK     bool   `json:"ok"`
	ID     int64  `json:"id,omitempty"`
	Sample string `json:"sample,omitempty"`
}

type qaReport struct {
	EmpresaID int64    `json:"empresa_id"`
	StartedAt string   `json:"started_at"`
	Steps     []qaStep `json:"steps"`
	Failures  []qaStep `json:"failures"`
}

func main() {
	base := flag.String("base", "http://127.0.0.1:8080", "URL base del servidor")
	empresaID := flag.Int64("empresa_id", 7, "empresa_id a probar")
	email := flag.String("email", "qa.admin.calipso@powerfulcontrolsystem.local", "usuario QA de empresa")
	password := flag.String("password", "QaCalipso2026", "password QA")
	flag.Parse()

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, Timeout: 120 * time.Second}
	report := qaReport{EmpresaID: *empresaID, StartedAt: time.Now().Format(time.RFC3339)}
	runID := time.Now().Format("20060102-150405")

	do := func(module, action, method, path string, payload interface{}) (map[string]interface{}, qaStep) {
		var body io.Reader
		if payload != nil {
			raw, _ := json.Marshal(payload)
			body = bytes.NewReader(raw)
		}
		req, err := http.NewRequest(method, strings.TrimRight(*base, "/")+path, body)
		if err != nil {
			log.Fatal(err)
		}
		if payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		resp, err := client.Do(req)
		if err != nil {
			step := qaStep{Module: module, Action: action, Method: method, Path: path, Status: 0, OK: false, Sample: err.Error()}
			report.Steps = append(report.Steps, step)
			report.Failures = append(report.Failures, step)
			return nil, step
		}
		defer resp.Body.Close()
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		step := qaStep{Module: module, Action: action, Method: method, Path: path, Status: resp.StatusCode, OK: resp.StatusCode >= 200 && resp.StatusCode < 300, Sample: strings.TrimSpace(string(raw))}
		var decoded map[string]interface{}
		if len(raw) > 0 {
			var decodedAny interface{}
			_ = json.Unmarshal(raw, &decodedAny)
			if m, ok := decodedAny.(map[string]interface{}); ok {
				decoded = m
				if v, ok := decoded["id"].(float64); ok {
					step.ID = int64(v)
				}
				if credito, ok := decoded["credito"].(map[string]interface{}); ok && step.ID == 0 {
					if v, ok := credito["id"].(float64); ok {
						step.ID = int64(v)
					}
				}
			} else if rows, ok := decodedAny.([]interface{}); ok {
				for _, row := range rows {
					m, ok := row.(map[string]interface{})
					if !ok {
						continue
					}
					disponible, hasDisponible := m["disponible"].(bool)
					if hasDisponible && !disponible {
						continue
					}
					if v, ok := m["estacion_id"].(float64); ok && v > 0 {
						step.ID = int64(v)
						if !hasDisponible || disponible {
							break
						}
					}
				}
			}
		}
		if len(step.Sample) > 280 {
			step.Sample = step.Sample[:280]
		}
		report.Steps = append(report.Steps, step)
		if !step.OK {
			report.Failures = append(report.Failures, step)
		}
		return decoded, step
	}

	loginBody := map[string]interface{}{
		"empresa_id":      *empresaID,
		"email":           strings.TrimSpace(*email),
		"password":        *password,
		"accept_contract": true,
		"recaptcha_token": "qa-local",
	}
	_, loginStep := do("usuarios", "login_usuario", http.MethodPost, "/api/empresa/usuarios/login", loginBody)
	if !loginStep.OK {
		writeReport(report)
		log.Fatalf("login_usuario fallo status=%d sample=%s", loginStep.Status, loginStep.Sample)
	}

	get := func(module, action, path string) {
		do(module, action, http.MethodGet, path, nil)
	}
	postID := func(module, action, path string, payload map[string]interface{}) int64 {
		_, step := do(module, action, http.MethodPost, path, payload)
		if !step.OK {
			return 0
		}
		return step.ID
	}
	post := func(module, action, path string, payload map[string]interface{}) {
		do(module, action, http.MethodPost, path, payload)
	}
	put := func(module, action, path string) {
		do(module, action, http.MethodPut, path, nil)
	}
	putPayload := func(module, action, path string, payload map[string]interface{}) {
		do(module, action, http.MethodPut, path, payload)
	}

	q := fmt.Sprintf("empresa_id=%d", *empresaID)
	today := time.Now().Format("2006-01-02")
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	stationID := int64(900000) + time.Now().UnixNano()%1000000
	stationCode := fmt.Sprintf("EST-%d-%d", *empresaID, stationID)
	stationName := "QA Habitacion Reserva Calipso " + runID

	postID("carritos_compra", "crear_estacion_reserva_qa", "/api/empresa/carritos_compra?"+q, map[string]interface{}{
		"empresa_id": *empresaID, "codigo": stationCode, "nombre": stationName, "canal_venta": "estacion", "estado_carrito": "cerrado", "moneda": "COP", "referencia_externa": fmt.Sprintf("ESTACION_%d", stationID), "metodo_pago": "efectivo", "referencia_pago": "QA-" + runID, "observaciones": "QA estacion aislada para reservas Calipso",
	})

	get("gimnasio", "dashboard", "/api/empresa/gimnasio?"+q+"&action=dashboard")
	planID := postID("gimnasio", "crear_plan", "/api/empresa/gimnasio?action=planes", map[string]interface{}{
		"empresa_id": *empresaID, "nombre": "QA Plan Calipso " + runID, "precio": 99000, "duracion_dias": 30, "clases_incluidas": 12, "acceso_ilimitado": true, "estado": "activo",
	})
	socioID := postID("gimnasio", "crear_socio", "/api/empresa/gimnasio?action=socios", map[string]interface{}{
		"empresa_id": *empresaID, "codigo": "QA-GYM-" + runID, "nombre_completo": "QA Socio Calipso " + runID, "documento": "GYM" + runID, "telefono": "3000000000", "email": "qa.gym." + runID + "@powerfulcontrolsystem.local", "objetivo": "Acondicionamiento", "estado": "activo", "plan_id": planID, "fecha_inicio_plan": today, "fecha_fin_plan": tomorrow,
	})
	entrenadorID := postID("gimnasio", "crear_entrenador", "/api/empresa/gimnasio?action=entrenadores", map[string]interface{}{
		"empresa_id": *empresaID, "nombre_completo": "QA Coach Calipso " + runID, "especialidad": "Funcional", "estado": "activo",
	})
	claseID := postID("gimnasio", "crear_clase", "/api/empresa/gimnasio?action=clases", map[string]interface{}{
		"empresa_id": *empresaID, "nombre": "QA Clase Calipso " + runID, "categoria": "Funcional", "entrenador_id": entrenadorID, "sede": "Calipso", "canal": "presencial", "cupos": 10, "duracion_minutos": 45, "fecha_programada": today + " 18:00:00", "estado": "programada", "precio": 0,
	})
	postID("gimnasio", "crear_inscripcion", "/api/empresa/gimnasio?action=inscripciones", map[string]interface{}{"empresa_id": *empresaID, "socio_id": socioID, "clase_id": claseID, "estado": "inscrito"})
	postID("gimnasio", "registrar_asistencia", "/api/empresa/gimnasio?action=checkin", map[string]interface{}{"empresa_id": *empresaID, "socio_id": socioID, "clase_id": claseID, "tipo_acceso": "clase", "canal": "recepcion", "sede": "Calipso"})
	postID("gimnasio", "registrar_pago", "/api/empresa/gimnasio?action=pagos", map[string]interface{}{"empresa_id": *empresaID, "socio_id": socioID, "plan_id": planID, "concepto": "QA mensualidad", "monto": 99000, "moneda": "COP", "metodo_pago": "efectivo", "estado": "pagado", "referencia": "QA-" + runID})
	get("gimnasio", "listar_socios", "/api/empresa/gimnasio?"+q+"&action=socios")

	get("creditos", "resumen", "/api/empresa/creditos?"+q+"&action=resumen")
	creditoID := postID("creditos", "crear_credito", "/api/empresa/creditos?"+q+"&action=crear", map[string]interface{}{
		"codigo": "QA-CRED-" + runID, "cliente_nombre": "QA Cliente Credito Calipso " + runID, "tipo_credito": "consumo", "monto_aprobado": 250000, "cupo_credito": 250000, "saldo_actual": 250000, "tasa_interes": 0, "tasa_mora": 0, "plazo_dias": 30, "plazo_cuotas": 1, "fecha_inicio": today, "fecha_vencimiento": tomorrow, "estado_credito": "activo", "observaciones": "QA credito Calipso",
	})
	post("creditos", "registrar_abono", "/api/empresa/creditos?"+q+"&action=abono", map[string]interface{}{"credito_id": creditoID, "monto": 50000, "metodo_pago": "efectivo", "referencia_pago": "QA-" + runID, "fecha_movimiento": today, "registrar_evento_contable": false, "procesar_asientos": false})
	get("creditos", "estado_cuenta", fmt.Sprintf("/api/empresa/creditos?%s&action=estado_cuenta&id=%d", q, creditoID))

	get("tarifas_motel", "listar", "/api/empresa/tarifas_motel?"+q+"&include_inactive=1")
	tarifaMotelID := postID("tarifas_motel", "crear_plan_motel", "/api/empresa/tarifas_motel?"+q, map[string]interface{}{
		"empresa_id": *empresaID, "estacion_id": stationID, "estacion_codigo": stationCode, "estacion_nombre": stationName, "nombre_plan": "QA Express Calipso " + runID, "tipo_plan": "express", "categoria_habitacion": "suite", "dia_semana_desde": 1, "dia_semana_hasta": 7, "hora_inicio": "00:00", "hora_fin": "23:59", "minutos_incluidos": 180, "valor_base": 85000, "minutos_extra": 30, "valor_extra": 15000, "cobrar_por_fraccion": true, "tolerancia_minutos": 10, "moneda": "COP", "prioridad": 1, "aplicar_automaticamente": true, "estado": "activo", "observaciones": "QA tarifa motel Calipso",
	})
	get("tarifas_motel", "calcular", fmt.Sprintf("/api/empresa/tarifas_motel?%s&action=calcular&id=%d&minutos_consumidos=245", q, tarifaMotelID))

	putPayload("tarifas_por_minutos", "configurar", "/api/empresa/tarifas_por_minutos?"+q+"&action=config", map[string]interface{}{"empresa_id": *empresaID, "redondeo_modo": "matematico", "redondeo_unidad": 100, "monto_minimo_diario": 0, "monto_maximo_diario": 0, "margen_tolerancia_entrada_minutos": 5, "sensor_auto_activar_estacion": false, "margen_desactivacion_habilitado": true, "margen_desactivacion_minutos": 10, "estado": "activo", "observaciones": "QA config minutos Calipso"})
	putPayload("tarifas_por_minutos", "aplicar_todas_estaciones", "/api/empresa/tarifas_por_minutos?"+q+"&action=aplicar_todas_estaciones", map[string]interface{}{"empresa_id": *empresaID, "estacion_id": stationID, "estacion_codigo": stationCode, "estacion_nombre": stationName, "dia_semana_desde": 1, "dia_semana_hasta": 7, "minutos_base": 180, "valor_base": 80000, "minutos_extra": 30, "valor_extra": 12000, "cobrar_por_fraccion": true, "moneda": "COP", "prioridad": 1, "estado": "activo", "observaciones": "QA tarifa minutos Calipso"})
	get("tarifas_por_minutos", "resolver", fmt.Sprintf("/api/empresa/tarifas_por_minutos?%s&action=aplicable&estacion_id=%d&dia_semana=1", q, stationID))
	get("tarifas_por_minutos", "calcular", fmt.Sprintf("/api/empresa/tarifas_por_minutos?%s&action=calcular&estacion_id=%d&dia_semana=1&minutos_consumidos=245", q, stationID))

	putPayload("tarifas_por_dia", "aplicar_todas_estaciones", "/api/empresa/tarifas_por_dia?"+q+"&action=aplicar_todas_estaciones", map[string]interface{}{"empresa_id": *empresaID, "estacion_id": stationID, "estacion_codigo": stationCode, "estacion_nombre": stationName, "servicio_nombre": "hospedaje", "valor_dia": 180000, "hora_check_in": "15:00", "hora_check_out": "12:00", "moneda": "COP", "prioridad": 1, "aplicar_automaticamente": true, "estado": "activo", "observaciones": "QA tarifa dia Calipso"})
	get("tarifas_por_dia", "calcular", fmt.Sprintf("/api/empresa/tarifas_por_dia?%s&action=calcular&estacion_id=%d&activado_en=%s%%2015:00:00&fecha_corte=%s%%2012:00:00", q, stationID, today, tomorrow))

	reservaInicio := time.Now().Add(72 * time.Hour).Format("2006-01-02 15:04:05")
	reservaFin := time.Now().Add(74 * time.Hour).Format("2006-01-02 15:04:05")
	_, disponibilidadStep := do("reservas_hotel", "disponibilidad", http.MethodGet, fmt.Sprintf("/api/empresa/reservas_hotel?%s&action=disponibilidad&fecha_entrada=%s&fecha_salida=%s", q, strings.ReplaceAll(reservaInicio, " ", "%20"), strings.ReplaceAll(reservaFin, " ", "%20")), nil)
	reservaStationID := stationID
	if disponibilidadStep.ID > 0 {
		reservaStationID = disponibilidadStep.ID
	}
	reservaID := postID("reservas_hotel", "crear_reserva", "/api/empresa/reservas_hotel?"+q, map[string]interface{}{"empresa_id": *empresaID, "estacion_id": reservaStationID, "codigo_reserva": "QA-RES-" + runID, "cliente_nombre": "QA Huesped Calipso " + runID, "cliente_documento": "RES" + runID, "cliente_email": "qa.reserva." + runID + "@powerfulcontrolsystem.local", "cliente_telefono": "3020000000", "cantidad_huespedes": 2, "fecha_entrada": reservaInicio, "fecha_salida": reservaFin, "monto_total": 85000, "moneda": "COP", "estado_reserva": "pendiente_pago", "estado_pago": "pendiente", "canal_origen": "recepcion", "request_id": "qa-" + runID, "observaciones": "QA reserva Calipso"})
	if reservaID > 0 {
		get("reservas_hotel", "detalle", fmt.Sprintf("/api/empresa/reservas_hotel?%s&action=detalle&id=%d", q, reservaID))
		putPayload("reservas_hotel", "confirmar_pago", "/api/empresa/reservas_hotel?"+q+"&action=confirmar_pago", map[string]interface{}{"empresa_id": *empresaID, "id": reservaID, "referencia_pago": "QA-PAGO-" + runID, "observaciones": "QA pago confirmado"})
	}

	get("alquileres", "dashboard", "/api/empresa/alquileres?"+q+"&action=dashboard")
	post("alquileres", "configurar", "/api/empresa/alquileres?"+q+"&action=config", map[string]interface{}{"nombre_sistema": "Alquileres QA Calipso", "moneda": "COP", "permitir_reservas": true, "permitir_gps": true, "requerir_deposito": true, "permitir_kilometraje": true, "requerir_checklist": true, "permitir_entrega_domicilio": false, "alertar_vencimiento_horas": 4, "deposito_base_sugerido": 50000})
	categoriaAlquilerID := postID("alquileres", "crear_categoria", "/api/empresa/alquileres?"+q+"&action=categorias", map[string]interface{}{"codigo": "QA-CAT-" + runID, "nombre": "QA Categoria Calipso " + runID, "tipo_activo": "habitacion", "descripcion": "Categoria QA", "estado": "activo"})
	activoAlquilerID := postID("alquileres", "crear_activo", "/api/empresa/alquileres?"+q+"&action=activos", map[string]interface{}{"codigo": "QA-ACT-" + runID, "nombre": "QA Activo Calipso " + runID, "categoria_id": categoriaAlquilerID, "tipo_activo": "habitacion", "marca": "Calipso", "modelo": "Suite", "serie": "SER" + runID, "placa": "QA" + runID, "sede": "Calipso", "estado": "disponible", "valor_reposicion": 2500000, "costo_base_hora": 45000, "deposito_sugerido": 50000, "usa_gps": true, "requiere_checklist": true, "requiere_licencia": false, "notas": "QA activo Calipso"})
	tarifaAlquilerID := postID("alquileres", "crear_tarifa", "/api/empresa/alquileres?"+q+"&action=tarifas", map[string]interface{}{"codigo": "QA-TAR-" + runID, "nombre": "QA Tarifa Calipso " + runID, "categoria_id": categoriaAlquilerID, "modalidad_cobro": "hora", "precio_base": 45000, "precio_hora": 45000, "precio_dia": 180000, "precio_semana": 900000, "precio_mes": 2800000, "kilometros_incluidos": 0, "deposito_minimo": 50000, "estado": "activo"})
	contratoAlquilerID := postID("alquileres", "crear_contrato", "/api/empresa/alquileres?"+q+"&action=contratos", map[string]interface{}{"codigo": "QA-CON-" + runID, "tipo_registro": "alquiler", "activo_id": activoAlquilerID, "cliente_nombre": "QA Cliente Alquiler Calipso " + runID, "cliente_documento": "ALQ" + runID, "cliente_telefono": "3030000000", "cliente_email": "qa.alquiler." + runID + "@powerfulcontrolsystem.local", "responsable_empresa": "QA Recepcion", "tarifa_id": tarifaAlquilerID, "modalidad_cobro": "hora", "fecha_reserva": today + " 10:00:00", "fecha_inicio": today + " 11:00:00", "fecha_fin_prevista": today + " 14:00:00", "estado": "activo", "cantidad": 1, "horas_planeadas": 3, "dias_planeados": 0, "deposito": 50000, "valor_base": 135000, "descuento": 0, "impuestos": 0, "total": 135000, "saldo_pendiente": 0, "origen_entrega": "Recepcion", "destino_devolucion": "Recepcion", "observaciones": "QA contrato Calipso", "requiere_garantia": true, "gps_tracking_activo": true})
	postID("alquileres", "crear_mantenimiento", "/api/empresa/alquileres?"+q+"&action=mantenimientos", map[string]interface{}{"activo_id": activoAlquilerID, "tipo": "preventivo", "prioridad": "media", "estado": "programado", "fecha_programada": tomorrow + " 09:00:00", "proveedor": "QA Proveedor", "costo_estimado": 25000, "descripcion": "QA mantenimiento Calipso", "observaciones": "QA"})
	postID("alquileres", "registrar_ubicacion", "/api/empresa/alquileres?"+q+"&action=ubicaciones", map[string]interface{}{"activo_id": activoAlquilerID, "contrato_id": contratoAlquilerID, "latitud": 3.4516, "longitud": -76.5320, "velocidad": 0, "precision_metros": 8, "fuente": "qa", "referencia": "Calipso QA"})
	post("alquileres", "cerrar_contrato", "/api/empresa/alquileres?"+q+"&action=cambiar_estado", map[string]interface{}{"contrato_id": contratoAlquilerID, "estado": "cerrado", "observaciones": "QA cierre contrato", "responsable": "QA Recepcion"})
	get("alquileres", "listar_contratos", "/api/empresa/alquileres?"+q+"&action=contratos")

	get("odontologia", "dashboard", "/api/empresa/odontologia?"+q+"&action=dashboard")
	pacienteID := postID("odontologia", "crear_paciente", "/api/empresa/odontologia?"+q+"&action=pacientes", map[string]interface{}{
		"nombre_completo": "QA Paciente Calipso " + runID, "documento": "OD" + runID, "telefono": "3010000000", "email": "qa.odonto." + runID + "@powerfulcontrolsystem.local", "estado": "activo",
	})
	proID := postID("odontologia", "crear_profesional", "/api/empresa/odontologia?"+q+"&action=profesionales", map[string]interface{}{
		"nombre_completo": "QA Odontologa Calipso " + runID, "especialidad": "Rehabilitacion", "registro_profesional": "QA-" + runID, "estado": "activo",
	})
	consultorioID := postID("odontologia", "crear_consultorio", "/api/empresa/odontologia?"+q+"&action=consultorios", map[string]interface{}{
		"nombre": "QA Consultorio " + runID, "sede": "Calipso", "sillon": "1", "estado": "activo",
	})
	citaID := postID("odontologia", "crear_cita", "/api/empresa/odontologia?"+q+"&action=citas", map[string]interface{}{
		"paciente_id": pacienteID, "profesional_id": proID, "consultorio_id": consultorioID, "fecha_hora": tomorrow + " 09:30:00", "duracion_minutos": 30, "motivo": "Control QA", "estado": "programada", "canal": "recepcion",
	})
	tratID := postID("odontologia", "crear_tratamiento", "/api/empresa/odontologia?"+q+"&action=tratamientos", map[string]interface{}{
		"paciente_id": pacienteID, "profesional_id": proID, "nombre": "QA Profilaxis " + runID, "categoria": "Preventiva", "sesiones_total": 1, "costo_estimado": 120000, "estado": "activo",
	})
	presID := postID("odontologia", "crear_presupuesto", "/api/empresa/odontologia?"+q+"&action=presupuestos", map[string]interface{}{
		"paciente_id": pacienteID, "tratamiento_id": tratID, "nombre": "QA Presupuesto " + runID, "valor_total": 120000, "cuota_inicial": 20000, "saldo": 100000, "estado": "vigente",
	})
	postID("odontologia", "registrar_pago", "/api/empresa/odontologia?"+q+"&action=pagos", map[string]interface{}{"paciente_id": pacienteID, "presupuesto_id": presID, "concepto": "QA abono odontologia", "monto": 20000, "metodo_pago": "efectivo", "referencia": "QA-" + runID, "estado": "aplicado"})
	put("odontologia", "confirmar_cita", fmt.Sprintf("/api/empresa/odontologia?%s&action=estado_cita&id=%d&estado=confirmada", q, citaID))
	get("odontologia", "listar_citas", "/api/empresa/odontologia?"+q+"&action=citas&fecha="+tomorrow)

	get("turnos", "dashboard", "/api/empresa/turnos_atencion?"+q+"&action=dashboard&fecha="+today)
	post("turnos", "configurar", "/api/empresa/turnos_atencion?"+q+"&action=config", map[string]interface{}{"nombre_sistema": "Turnos QA Calipso", "nombre_pantalla": "Calipso", "prefijo_general": "Q", "tiempo_llamado_segundos": 20, "permitir_emision_publica": true, "mostrar_tickets_completados": true})
	servID := postID("turnos", "crear_servicio", "/api/empresa/turnos_atencion?"+q+"&action=servicios", map[string]interface{}{"codigo": "QAS" + runID, "nombre": "QA Servicio " + runID, "descripcion": "Servicio QA", "prefijo": "Q", "prioridad": 1, "color": "#2563eb", "estado": "activo"})
	puestoID := postID("turnos", "crear_puesto", "/api/empresa/turnos_atencion?"+q+"&action=puestos", map[string]interface{}{"codigo": "QAP" + runID, "nombre": "QA Puesto " + runID, "area": "Recepcion", "ubicacion": "Calipso", "servicios_permitidos": fmt.Sprintf("%d", servID), "estado": "activo"})
	ticket, ticketStep := do("turnos", "emitir_ticket", http.MethodPost, "/api/empresa/turnos_atencion?"+q+"&action=emitir_ticket", map[string]interface{}{"servicio_id": servID, "documento_cliente": "QA" + runID, "nombre_cliente": "QA Cliente Turnos " + runID, "canal_emision": "recepcion"})
	ticketID := ticketStep.ID
	if ticketID == 0 {
		if v, ok := ticket["id"].(float64); ok {
			ticketID = int64(v)
		}
	}
	post("turnos", "llamar_siguiente", "/api/empresa/turnos_atencion?"+q+"&action=llamar_siguiente", map[string]interface{}{"puesto_id": puestoID})
	post("turnos", "cerrar_ticket", "/api/empresa/turnos_atencion?"+q+"&action=cambiar_estado", map[string]interface{}{"ticket_id": ticketID, "puesto_id": puestoID, "estado": "completado", "observaciones": "QA completado"})
	get("turnos", "display_publico", "/api/public/turnos_atencion?"+q+"&action=display&fecha="+today)

	writeReport(report)
	if len(report.Failures) > 0 {
		for _, failure := range report.Failures {
			log.Printf("FAIL module=%s action=%s status=%d path=%s sample=%s", failure.Module, failure.Action, failure.Status, failure.Path, failure.Sample)
		}
		log.Fatalf("RESULTADO_QA_MODULOS failures=%d", len(report.Failures))
	}
	fmt.Printf("RESULTADO_QA_MODULOS_OK steps=%d reporte=%s\n", len(report.Steps), reportPath())
}

func writeReport(report qaReport) {
	out, _ := json.MarshalIndent(report, "", "  ")
	_ = os.MkdirAll(filepath.Dir(reportPath()), 0755)
	_ = os.WriteFile(reportPath(), out, 0600)
}

func reportPath() string {
	return filepath.Join("tmp_tools", "qa_calipso_modulos", "reporte_modulos_calipso.json")
}
