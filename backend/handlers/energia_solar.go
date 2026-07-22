package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

type energiaSolarCatalogItem struct {
	Proveedor       string   `json:"proveedor"`
	Nombre          string   `json:"nombre"`
	Plataforma      string   `json:"plataforma"`
	Modelos         []string `json:"modelos"`
	Baterias        []string `json:"baterias"`
	MetricasClave   []string `json:"metricas_clave"`
	ApiBaseSugerida string   `json:"api_base_sugerida,omitempty"`
	Nota            string   `json:"nota"`
}

type energiaSolarWritePayload struct {
	Sistema dbpkg.EmpresaEnergiaSolarSistema `json:"sistema"`
	Alerta  dbpkg.EmpresaEnergiaSolarAlerta  `json:"alerta"`
	Lectura dbpkg.EmpresaEnergiaSolarLectura `json:"lectura"`
}

func EmpresaEnergiaSolarHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := dbpkg.EmpresaEnergiaSolarSchemaReady(dbEmp); err != nil {
			log.Printf("[energia_solar] schema readiness empresa_id=%d error: %v", empresaID, err)
			http.Error(w, "No se pudo preparar el modulo de energia solar", http.StatusInternalServerError)
			return
		}

		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch r.Method {
		case http.MethodGet:
			handleEmpresaEnergiaSolarGET(w, r, dbEmp, empresaID, action)
		case http.MethodPost, http.MethodPut:
			handleEmpresaEnergiaSolarWrite(w, r, dbEmp, dbSuper, empresaID, action)
		default:
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
		}
	}
}

func handleEmpresaEnergiaSolarGET(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, empresaID int64, action string) {
	switch action {
	case "", "dashboard":
		dashboard, err := buildEmpresaEnergiaSolarDashboard(dbEmp, empresaID)
		if err != nil {
			log.Printf("[energia_solar] dashboard empresa_id=%d error: %v", empresaID, err)
			http.Error(w, "No se pudo cargar el dashboard solar", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, dashboard)
	case "catalogo":
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "proveedores": energiaSolarProviderCatalog(), "baterias": energiaSolarBatteryCatalog(), "alertas": energiaSolarAlertCatalog()})
	case "sistemas":
		includeInactive := strings.TrimSpace(r.URL.Query().Get("include_inactive")) == "1"
		items, err := dbpkg.ListEmpresaEnergiaSolarSistemas(dbEmp, empresaID, includeInactive)
		if err != nil {
			log.Printf("[energia_solar] sistemas empresa_id=%d error: %v", empresaID, err)
			http.Error(w, "No se pudieron listar los sistemas solares", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "items": items})
	case "alertas":
		sistemaID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("sistema_id")), 10, 64)
		includeInactive := strings.TrimSpace(r.URL.Query().Get("include_inactive")) == "1"
		items, err := dbpkg.ListEmpresaEnergiaSolarAlertas(dbEmp, empresaID, sistemaID, includeInactive)
		if err != nil {
			log.Printf("[energia_solar] alertas empresa_id=%d sistema_id=%d error: %v", empresaID, sistemaID, err)
			http.Error(w, "No se pudieron listar las alertas solares", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "items": items})
	case "lecturas":
		sistemaID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("sistema_id")), 10, 64)
		limit, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("limit")))
		items, err := dbpkg.ListEmpresaEnergiaSolarLecturas(dbEmp, empresaID, sistemaID, limit)
		if err != nil {
			log.Printf("[energia_solar] lecturas empresa_id=%d sistema_id=%d error: %v", empresaID, sistemaID, err)
			http.Error(w, "No se pudieron listar las lecturas solares", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "items": items})
	case "eventos":
		sistemaID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("sistema_id")), 10, 64)
		limit, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("limit")))
		items, err := dbpkg.ListEmpresaEnergiaSolarEventos(dbEmp, empresaID, sistemaID, limit)
		if err != nil {
			log.Printf("[energia_solar] eventos empresa_id=%d sistema_id=%d error: %v", empresaID, sistemaID, err)
			http.Error(w, "No se pudieron listar los eventos solares", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "items": items})
	default:
		http.Error(w, "accion no soportada", http.StatusBadRequest)
	}
}

func handleEmpresaEnergiaSolarWrite(w http.ResponseWriter, r *http.Request, dbEmp, dbSuper *sql.DB, empresaID int64, action string) {
	switch action {
	case "sistema":
		var payload energiaSolarWritePayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "JSON invalido", http.StatusBadRequest)
			return
		}
		item := payload.Sistema
		item.EmpresaID = empresaID
		item.UsuarioCreador = adminEmailFromRequest(r)
		if item.Estado == "" {
			item.Estado = "activo"
		}
		if !strings.EqualFold(strings.TrimSpace(item.Estado), "inactivo") && !item.Activo {
			item.Activo = true
		}
		id, err := dbpkg.UpsertEmpresaEnergiaSolarSistema(dbEmp, item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_ = ensureEmpresaEnergiaSolarDefaultAlerts(dbEmp, empresaID, id, item.UsuarioCreador)
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
	case "alerta":
		var payload energiaSolarWritePayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "JSON invalido", http.StatusBadRequest)
			return
		}
		item := payload.Alerta
		item.EmpresaID = empresaID
		item.UsuarioCreador = adminEmailFromRequest(r)
		if item.Estado == "" {
			item.Estado = "activo"
		}
		if !strings.EqualFold(strings.TrimSpace(item.Estado), "inactivo") && !item.Activo {
			item.Activo = true
		}
		id, err := dbpkg.UpsertEmpresaEnergiaSolarAlerta(dbEmp, item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
	case "lectura":
		var payload energiaSolarWritePayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "JSON invalido", http.StatusBadRequest)
			return
		}
		lectura := payload.Lectura
		lectura.EmpresaID = empresaID
		lectura.UsuarioCreador = adminEmailFromRequest(r)
		if lectura.Raw == nil {
			lectura.Raw = map[string]interface{}{"origen": "manual_o_gateway", "empresa_id": empresaID}
		}
		lecturaID, err := dbpkg.InsertEmpresaEnergiaSolarLectura(dbEmp, lectura)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		eventos := evaluateEmpresaEnergiaSolarAlertas(r, dbEmp, dbSuper, empresaID, lectura.SistemaID, lecturaID, lectura)
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": lecturaID, "eventos": eventos})
	case "probar_alerta":
		sistemaID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("sistema_id")), 10, 64)
		if sistemaID <= 0 {
			http.Error(w, "sistema_id requerido", http.StatusBadRequest)
			return
		}
		sistema, err := dbpkg.GetEmpresaEnergiaSolarSistema(dbEmp, empresaID, sistemaID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "sistema solar no encontrado para esta empresa", http.StatusNotFound)
				return
			}
			http.Error(w, "No se pudo cargar el sistema solar", http.StatusInternalServerError)
			return
		}
		emailOK, emailErr := sendEmpresaEnergiaSolarEmail(dbSuper, empresaID, sistema.EmailAlertas, "Prueba de alerta energia solar", "Esta es una prueba de alerta del modulo Energia solar para "+sistema.Nombre+".", adminEmailFromRequest(r))
		eventoID, _ := dbpkg.InsertEmpresaEnergiaSolarEvento(dbEmp, dbpkg.EmpresaEnergiaSolarEvento{
			EmpresaID:      empresaID,
			SistemaID:      sistemaID,
			Tipo:           "prueba_alerta",
			Severidad:      "info",
			Mensaje:        "Prueba manual de alerta solar",
			EmailEnviado:   emailOK,
			EmailError:     emailErr,
			UsuarioCreador: adminEmailFromRequest(r),
			Estado:         "activo",
		})
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "evento_id": eventoID, "email_enviado": emailOK, "email_error": emailErr})
	default:
		http.Error(w, "accion no soportada", http.StatusBadRequest)
	}
}

func buildEmpresaEnergiaSolarDashboard(dbEmp *sql.DB, empresaID int64) (map[string]interface{}, error) {
	sistemas, err := dbpkg.ListEmpresaEnergiaSolarSistemas(dbEmp, empresaID, false)
	if err != nil {
		return nil, err
	}
	lecturas, err := dbpkg.ListEmpresaEnergiaSolarLecturas(dbEmp, empresaID, 0, 120)
	if err != nil {
		return nil, err
	}
	eventos, err := dbpkg.ListEmpresaEnergiaSolarEventos(dbEmp, empresaID, 0, 80)
	if err != nil {
		return nil, err
	}
	latestBySystem := map[int64]dbpkg.EmpresaEnergiaSolarLectura{}
	var totalSolar, totalInversor, socSum, sohSum float64
	var socCount, sohCount int
	for _, lectura := range lecturas {
		if _, ok := latestBySystem[lectura.SistemaID]; ok {
			continue
		}
		latestBySystem[lectura.SistemaID] = lectura
		totalSolar += lectura.PotenciaSolarW
		totalInversor += lectura.InversorPotenciaW
		if lectura.BateriaSOC > 0 {
			socSum += lectura.BateriaSOC
			socCount++
		}
		if lectura.BateriaSOH > 0 {
			sohSum += lectura.BateriaSOH
			sohCount++
		}
	}
	activeEvents := 0
	for _, evento := range eventos {
		if strings.EqualFold(strings.TrimSpace(evento.Estado), "activo") && evento.Severidad != "info" {
			activeEvents++
		}
	}
	var socAvg, sohAvg float64
	if socCount > 0 {
		socAvg = socSum / float64(socCount)
	}
	if sohCount > 0 {
		sohAvg = sohSum / float64(sohCount)
	}
	return map[string]interface{}{
		"ok": true,
		"kpis": map[string]interface{}{
			"sistemas_activos":         len(sistemas),
			"potencia_solar_w":         totalSolar,
			"potencia_inversor_w":      totalInversor,
			"bateria_soc_promedio":     socAvg,
			"bateria_soh_promedio":     sohAvg,
			"alertas_activas_reciente": activeEvents,
		},
		"sistemas": sistemas,
		"lecturas": lecturas,
		"eventos":  eventos,
	}, nil
}

func ensureEmpresaEnergiaSolarDefaultAlerts(dbEmp *sql.DB, empresaID, sistemaID int64, usuario string) error {
	existing, err := dbpkg.ListEmpresaEnergiaSolarAlertas(dbEmp, empresaID, sistemaID, true)
	if err != nil || len(existing) > 0 {
		return err
	}
	defaults := []dbpkg.EmpresaEnergiaSolarAlerta{
		{Tipo: "bateria_soc_baja", Nombre: "Bateria con carga baja", Operador: "<", Umbral: 20, Severidad: "alta", EnviarEmail: true, Activo: true},
		{Tipo: "bateria_no_carga", Nombre: "Bateria no esta cargando", Operador: "<=", Umbral: 5, Severidad: "alta", EnviarEmail: true, Activo: true},
		{Tipo: "bateria_soh_baja", Nombre: "Salud de bateria baja", Operador: "<", Umbral: 80, Severidad: "media", EnviarEmail: true, Activo: true},
		{Tipo: "panel_sin_produccion", Nombre: "Paneles sin produccion", Operador: "<", Umbral: 50, Severidad: "media", EnviarEmail: true, Activo: true},
		{Tipo: "temperatura_alta", Nombre: "Temperatura de bateria/inversor alta", Operador: ">", Umbral: 55, Severidad: "alta", EnviarEmail: true, Activo: true},
		{Tipo: "celda_desbalanceada", Nombre: "Desbalance de celdas", Operador: ">", Umbral: 0.15, Severidad: "alta", EnviarEmail: true, Activo: true},
		{Tipo: "inversor_error", Nombre: "Error reportado por inversor", Operador: "estado", Umbral: 0, Severidad: "critica", EnviarEmail: true, Activo: true},
	}
	for _, alerta := range defaults {
		alerta.EmpresaID = empresaID
		alerta.SistemaID = sistemaID
		alerta.UsuarioCreador = usuario
		alerta.Estado = "activo"
		if _, err := dbpkg.UpsertEmpresaEnergiaSolarAlerta(dbEmp, alerta); err != nil {
			return err
		}
	}
	return nil
}

func evaluateEmpresaEnergiaSolarAlertas(r *http.Request, dbEmp, dbSuper *sql.DB, empresaID, sistemaID, lecturaID int64, lectura dbpkg.EmpresaEnergiaSolarLectura) []dbpkg.EmpresaEnergiaSolarEvento {
	out := []dbpkg.EmpresaEnergiaSolarEvento{}
	sistema, err := dbpkg.GetEmpresaEnergiaSolarSistema(dbEmp, empresaID, sistemaID)
	if err != nil {
		return out
	}
	alertas, err := dbpkg.ListEmpresaEnergiaSolarAlertas(dbEmp, empresaID, sistemaID, false)
	if err != nil {
		return out
	}
	for _, alerta := range alertas {
		triggered, value, mensaje := energiaSolarAlertaTriggered(alerta, lectura)
		if !triggered {
			continue
		}
		emailOK := false
		emailErr := ""
		if sistema.AlertasEmailActivas && alerta.EnviarEmail {
			subject := "Alerta energia solar - " + alerta.Nombre
			body := fmt.Sprintf("Empresa ID: %d\r\nSistema: %s\r\nProveedor: %s\r\nBateria: %s %s\r\nAlerta: %s\r\nValor detectado: %.2f\r\nDetalle: %s\r\n", empresaID, sistema.Nombre, sistema.Proveedor, sistema.BateriaMarca, sistema.BateriaModelo, alerta.Nombre, value, mensaje)
			emailOK, emailErr = sendEmpresaEnergiaSolarEmail(dbSuper, empresaID, sistema.EmailAlertas, subject, body, adminEmailFromRequest(r))
		}
		evento := dbpkg.EmpresaEnergiaSolarEvento{
			EmpresaID:      empresaID,
			SistemaID:      sistemaID,
			AlertaID:       alerta.ID,
			LecturaID:      lecturaID,
			Tipo:           alerta.Tipo,
			Severidad:      alerta.Severidad,
			Mensaje:        mensaje,
			EmailEnviado:   emailOK,
			EmailError:     emailErr,
			UsuarioCreador: adminEmailFromRequest(r),
			Estado:         "activo",
		}
		if id, err := dbpkg.InsertEmpresaEnergiaSolarEvento(dbEmp, evento); err == nil {
			evento.ID = id
		}
		out = append(out, evento)
	}
	return out
}

func energiaSolarAlertaTriggered(alerta dbpkg.EmpresaEnergiaSolarAlerta, lectura dbpkg.EmpresaEnergiaSolarLectura) (bool, float64, string) {
	tipo := strings.ToLower(strings.TrimSpace(alerta.Tipo))
	switch tipo {
	case "inversor_error":
		if energiaSolarStatusHasError(lectura.EstadoInversor) {
			return true, 1, "El inversor reporto estado de error: " + lectura.EstadoInversor
		}
		return false, 0, ""
	case "bms_error":
		if energiaSolarStatusHasError(lectura.EstadoBateria) {
			return true, 1, "El BMS o bateria reporto estado de error: " + lectura.EstadoBateria
		}
		return false, 0, ""
	}
	value := energiaSolarMetricValue(tipo, lectura)
	triggered := compareEnergiaSolarMetric(value, alerta.Operador, alerta.Umbral)
	if !triggered {
		return false, value, ""
	}
	return true, value, fmt.Sprintf("%s: valor %.2f cumple condicion %s %.2f", alerta.Nombre, value, alerta.Operador, alerta.Umbral)
}

func energiaSolarMetricValue(tipo string, lectura dbpkg.EmpresaEnergiaSolarLectura) float64 {
	switch tipo {
	case "bateria_soc_baja":
		return lectura.BateriaSOC
	case "bateria_no_carga":
		return lectura.BateriaCargaW
	case "bateria_soh_baja":
		return lectura.BateriaSOH
	case "temperatura_alta":
		return lectura.TemperaturaC
	case "panel_sin_produccion":
		return lectura.PotenciaSolarW
	case "voltaje_bateria_bajo":
		return lectura.BateriaVoltaje
	case "descarga_alta":
		return lectura.BateriaDescargaW
	case "celda_desbalanceada":
		if lectura.CeldaVoltajeMax <= 0 || lectura.CeldaVoltajeMin <= 0 {
			return 0
		}
		return lectura.CeldaVoltajeMax - lectura.CeldaVoltajeMin
	default:
		return 0
	}
}

func compareEnergiaSolarMetric(value float64, op string, threshold float64) bool {
	switch strings.TrimSpace(op) {
	case "<":
		return value < threshold
	case "<=":
		return value <= threshold
	case ">":
		return value > threshold
	case ">=":
		return value >= threshold
	case "=", "==":
		return value == threshold
	default:
		return value < threshold
	}
}

func energiaSolarStatusHasError(value string) bool {
	clean := strings.ToLower(strings.TrimSpace(value))
	if clean == "" {
		return false
	}
	keywords := []string{"error", "fault", "falla", "fallo", "alarm", "alarma", "offline", "dañado", "danado", "no_carga", "no carga"}
	for _, keyword := range keywords {
		if strings.Contains(clean, keyword) {
			return true
		}
	}
	return false
}

func sendEmpresaEnergiaSolarEmail(dbSuper *sql.DB, empresaID int64, rawRecipients, subject, body, usuario string) (bool, string) {
	recipients := parseEnergiaSolarRecipients(rawRecipients)
	if len(recipients) == 0 {
		return false, "sin destinatarios configurados"
	}
	if isEmpresaUsuarioMailTestMode(dbSuper) {
		for _, to := range recipients {
			metadataJSON := fmt.Sprintf(`{"tipo":"energia_solar","empresa_id":%d}`, empresaID)
			if err := captureEmpresaUsuarioMailNotification(dbSuper, "energia_solar_alerta", empresaID, to, subject, body, "", metadataJSON, usuario); err != nil {
				return false, err.Error()
			}
		}
		return true, ""
	}
	for _, to := range recipients {
		if err := sendServerStartupEmail(dbSuper, to, subject, body); err != nil {
			return false, err.Error()
		}
	}
	return true, ""
}

func parseEnergiaSolarRecipients(raw string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, part := range strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || r == ';' || r == '\n' || r == '\r' || r == '\t' }) {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		addr, err := mail.ParseAddress(part)
		if err != nil {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(addr.Address))
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, addr.Address)
	}
	return out
}

func energiaSolarProviderCatalog() []energiaSolarCatalogItem {
	return []energiaSolarCatalogItem{
		{
			Proveedor:       dbpkg.EnergiaSolarProviderVictron,
			Nombre:          "Victron Energy",
			Plataforma:      "VRM Portal / VictronConnect / Venus OS",
			Modelos:         []string{"Cerbo GX", "Venus GX", "SmartSolar MPPT", "MultiPlus-II", "EasySolar-II GX"},
			Baterias:        []string{"Victron Lithium NG", "Pylontech US5000", "BYD Battery-Box", "Baterias LiFePO4 con BMS CAN-bus"},
			MetricasClave:   []string{"produccion solar", "SOC", "SOH", "carga/descarga", "temperatura", "alarmas VRM", "estado inversor"},
			ApiBaseSugerida: "https://vrmapi.victronenergy.com/v2",
			Nota:            "Recomendado cuando la instalacion usa equipo GX o controladoras Victron con monitoreo remoto.",
		},
		{
			Proveedor:       dbpkg.EnergiaSolarProviderSMA,
			Nombre:          "SMA",
			Plataforma:      "Sunny Portal powered by ennexOS",
			Modelos:         []string{"Sunny Boy", "Sunny Tripower", "Sunny Island", "SMA Data Manager M"},
			Baterias:        []string{"BYD Battery-Box Premium", "SMA Home Storage", "LG Energy Solution RESU", "Bateria compatible con inversor hibrido SMA"},
			MetricasClave:   []string{"energia FV", "rendimiento", "autoconsumo", "estado de planta", "fallas de inversor", "bateria hibrida"},
			ApiBaseSugerida: "https://ennexos.sunnyportal.com",
			Nota:            "Orientado a plantas residenciales y comerciales con Data Manager o inversores SMA registrados.",
		},
		{
			Proveedor:       dbpkg.EnergiaSolarProviderSolarEdge,
			Nombre:          "SolarEdge",
			Plataforma:      "SolarEdge Monitoring Platform",
			Modelos:         []string{"Home Hub Inverter", "HD-Wave", "Three Phase Inverter", "Power Optimizer"},
			Baterias:        []string{"SolarEdge Home Battery", "Tesla Powerwall por integracion externa", "Baterias compatibles por inversor hibrido"},
			MetricasClave:   []string{"datos por modulo", "estado de inversor", "produccion", "optimizadores", "bateria", "eventos de sitio"},
			ApiBaseSugerida: "https://monitoringapi.solaredge.com",
			Nota:            "Adecuado cuando se requiere diagnostico por panel/optimizador e inversor SolarEdge.",
		},
		{
			Proveedor:     dbpkg.EnergiaSolarProviderLocal,
			Nombre:        "Gateway local / BMS",
			Plataforma:    "Modbus TCP/RTU, CAN-bus, MQTT o API local",
			Modelos:       []string{"Controladora local", "Raspberry Pi gateway", "Medidor industrial", "BMS CAN/RS485"},
			Baterias:      []string{"Pylontech US5000", "BYD Battery-Box", "Enphase IQ Battery", "Tesla Powerwall Gateway", "Banco LiFePO4 generico"},
			MetricasClave: []string{"SOC", "SOH", "voltaje por celda", "corriente", "ciclos", "temperatura", "alarmas BMS"},
			Nota:          "Para instalaciones que exponen datos locales sin depender de nube del fabricante.",
		},
	}
}

func energiaSolarBatteryCatalog() []map[string]interface{} {
	return []map[string]interface{}{
		{"marca": "Tesla", "modelo": "Powerwall", "metricas": []string{"SOC", "potencia", "energia solar/red/casa", "modo respaldo", "estado Gateway"}},
		{"marca": "BYD", "modelo": "Battery-Box Premium HVS/HVM/LVS/LVL", "metricas": []string{"SOC", "SOH", "BMS", "temperatura", "estado/fallas", "modulos"}},
		{"marca": "Pylontech", "modelo": "US5000 / US3000C", "metricas": []string{"SOC", "SOH", "voltaje", "corriente", "ciclos", "alarmas BMS"}},
		{"marca": "Enphase", "modelo": "IQ Battery", "metricas": []string{"SOC", "respaldo", "microinversores", "energia", "estado de comunicacion"}},
		{"marca": "Victron", "modelo": "Lithium NG / Smart Lithium", "metricas": []string{"SOC", "voltaje", "temperatura", "BMS", "alarmas VRM"}},
	}
}

func energiaSolarAlertCatalog() []map[string]interface{} {
	return []map[string]interface{}{
		{"tipo": "bateria_soc_baja", "nombre": "Bateria con carga baja", "operador": "<", "umbral": 20},
		{"tipo": "bateria_no_carga", "nombre": "Bateria no carga", "operador": "<=", "umbral": 5},
		{"tipo": "bateria_soh_baja", "nombre": "Salud de bateria baja", "operador": "<", "umbral": 80},
		{"tipo": "temperatura_alta", "nombre": "Temperatura alta", "operador": ">", "umbral": 55},
		{"tipo": "panel_sin_produccion", "nombre": "Panel sin produccion", "operador": "<", "umbral": 50},
		{"tipo": "celda_desbalanceada", "nombre": "Desbalance de celdas", "operador": ">", "umbral": 0.15},
		{"tipo": "inversor_error", "nombre": "Error de inversor", "operador": "estado", "umbral": 0},
		{"tipo": "bms_error", "nombre": "Error de BMS/bateria", "operador": "estado", "umbral": 0},
	}
}
