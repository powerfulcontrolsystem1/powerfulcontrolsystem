package db

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

type EmpresaCRMMetaComercial struct {
	ID                int64   `json:"id"`
	EmpresaID         int64   `json:"empresa_id"`
	Periodo           string  `json:"periodo"`
	Propietario       string  `json:"propietario"`
	Canal             string  `json:"canal"`
	MetaValor         float64 `json:"meta_valor"`
	MetaLeads         int     `json:"meta_leads"`
	MetaConversionPct float64 `json:"meta_conversion_pct"`
	Estado            string  `json:"estado"`
	UsuarioCreador    string  `json:"usuario_creador,omitempty"`
	FechaCreacion     string  `json:"fecha_creacion,omitempty"`
}

type EmpresaCRMAgendaItem struct {
	Tipo        string  `json:"tipo"`
	Referencia  string  `json:"referencia"`
	Nombre      string  `json:"nombre"`
	Responsable string  `json:"responsable"`
	Fecha       string  `json:"fecha"`
	Estado      string  `json:"estado"`
	Valor       float64 `json:"valor"`
}

type EmpresaCRMEmbudoEstado struct {
	Estado               string  `json:"estado"`
	Leads                int     `json:"leads"`
	Valor                float64 `json:"valor"`
	Forecast             float64 `json:"forecast"`
	ProbabilidadPromedio float64 `json:"probabilidad_promedio"`
}

type EmpresaCRMLeadScore struct {
	ID              int64   `json:"id"`
	Codigo          string  `json:"codigo"`
	Nombre          string  `json:"nombre"`
	EmpresaOrigen   string  `json:"empresa_origen"`
	EstadoLead      string  `json:"estado_lead"`
	ValorPotencial  float64 `json:"valor_potencial"`
	Probabilidad    float64 `json:"probabilidad"`
	Interacciones   int     `json:"interacciones"`
	Score           float64 `json:"score"`
	Recomendacion   string  `json:"recomendacion"`
	ProximoContacto string  `json:"proximo_contacto,omitempty"`
}

type EmpresaCRMResponsableRendimiento struct {
	Responsable          string  `json:"responsable"`
	LeadsActivos         int     `json:"leads_activos"`
	LeadsVencidos        int     `json:"leads_vencidos"`
	ValorPipeline        float64 `json:"valor_pipeline"`
	ForecastPonderado    float64 `json:"forecast_ponderado"`
	ProbabilidadPromedio float64 `json:"probabilidad_promedio"`
}

type EmpresaCRMCanalRendimiento struct {
	Canal             string  `json:"canal"`
	Leads             int     `json:"leads"`
	Ganados           int     `json:"ganados"`
	Perdidos          int     `json:"perdidos"`
	ValorPipeline     float64 `json:"valor_pipeline"`
	ForecastPonderado float64 `json:"forecast_ponderado"`
	ConversionPct     float64 `json:"conversion_pct"`
}

type EmpresaCRMAccionPrioritaria struct {
	Prioridad   int     `json:"prioridad"`
	Severidad   string  `json:"severidad"`
	Titulo      string  `json:"titulo"`
	Detalle     string  `json:"detalle"`
	Responsable string  `json:"responsable,omitempty"`
	Referencia  string  `json:"referencia,omitempty"`
	Fecha       string  `json:"fecha,omitempty"`
	Valor       float64 `json:"valor,omitempty"`
	Accion      string  `json:"accion"`
}

type EmpresaCRMVentasAvanzadasDashboard struct {
	EmpresaID            int64                              `json:"empresa_id"`
	Periodo              string                             `json:"periodo"`
	LeadsActivos         int                                `json:"leads_activos"`
	LeadsGanados         int                                `json:"leads_ganados"`
	LeadsPerdidos        int                                `json:"leads_perdidos"`
	LeadsVencidos        int                                `json:"leads_vencidos"`
	LeadsSinContacto     int                                `json:"leads_sin_contacto"`
	LeadsEstancados      int                                `json:"leads_estancados"`
	AgendaHoy            int                                `json:"agenda_hoy"`
	CampanasActivas      int                                `json:"campanas_activas"`
	ValorPipeline        float64                            `json:"valor_pipeline"`
	ForecastPonderado    float64                            `json:"forecast_ponderado"`
	ValorRiesgo          float64                            `json:"valor_riesgo"`
	SaludComercialPct    float64                            `json:"salud_comercial_pct"`
	CotizacionesAbiertas int                                `json:"cotizaciones_abiertas"`
	CotizacionesValor    float64                            `json:"cotizaciones_valor"`
	PedidosAbiertos      int                                `json:"pedidos_abiertos"`
	PedidosValor         float64                            `json:"pedidos_valor"`
	MetaValor            float64                            `json:"meta_valor"`
	CumplimientoMetaPct  float64                            `json:"cumplimiento_meta_pct"`
	ConversionPct        float64                            `json:"conversion_pct"`
	TicketPromedio       float64                            `json:"ticket_promedio"`
	Alertas              []string                           `json:"alertas"`
	Embudo               []EmpresaCRMEmbudoEstado           `json:"embudo"`
	Agenda               []EmpresaCRMAgendaItem             `json:"agenda"`
	TopLeads             []EmpresaCRMLeadScore              `json:"top_leads"`
	Metas                []EmpresaCRMMetaComercial          `json:"metas"`
	Responsables         []EmpresaCRMResponsableRendimiento `json:"responsables"`
	Canales              []EmpresaCRMCanalRendimiento       `json:"canales"`
	AccionesPrioritarias []EmpresaCRMAccionPrioritaria      `json:"acciones_prioritarias"`
}

func EnsureEmpresaCRMVentasAvanzadasSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	if err := EnsureEmpresaModulosFaltantesSchema(dbConn); err != nil {
		return err
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_crm_metas_comerciales (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			periodo TEXT NOT NULL,
			propietario TEXT DEFAULT '',
			canal TEXT DEFAULT '',
			meta_valor REAL DEFAULT 0,
			meta_leads INTEGER DEFAULT 0,
			meta_conversion_pct REAL DEFAULT 0,
			estado TEXT DEFAULT 'activo',
			usuario_creador TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(empresa_id, periodo, propietario, canal)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_crm_metas_empresa_periodo ON empresa_crm_metas_comerciales(empresa_id, periodo, estado)`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func UpsertEmpresaCRMMetaComercial(dbConn *sql.DB, meta EmpresaCRMMetaComercial) (int64, error) {
	meta = normalizeCRMMetaComercial(meta)
	if meta.EmpresaID <= 0 || meta.Periodo == "" {
		return 0, errors.New("empresa_id y periodo son requeridos")
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_crm_metas_comerciales
		(empresa_id,periodo,propietario,canal,meta_valor,meta_leads,meta_conversion_pct,estado,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?)
		ON CONFLICT (empresa_id,periodo,propietario,canal) DO UPDATE SET
			meta_valor=EXCLUDED.meta_valor,
			meta_leads=EXCLUDED.meta_leads,
			meta_conversion_pct=EXCLUDED.meta_conversion_pct,
			estado=EXCLUDED.estado,
			usuario_creador=EXCLUDED.usuario_creador,
			fecha_actualizacion=CURRENT_TIMESTAMP`,
		meta.EmpresaID, meta.Periodo, meta.Propietario, meta.Canal, meta.MetaValor, meta.MetaLeads, meta.MetaConversionPct, meta.Estado, meta.UsuarioCreador)
}

func ListEmpresaCRMMetasComerciales(dbConn *sql.DB, empresaID int64, periodo string) ([]EmpresaCRMMetaComercial, error) {
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if strings.TrimSpace(periodo) != "" {
		where += " AND periodo=?"
		args = append(args, strings.TrimSpace(periodo))
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,periodo,COALESCE(propietario,''),COALESCE(canal,''),COALESCE(meta_valor,0),COALESCE(meta_leads,0),COALESCE(meta_conversion_pct,0),COALESCE(estado,'activo'),COALESCE(usuario_creador,''),COALESCE(fecha_creacion,'') FROM empresa_crm_metas_comerciales WHERE %s ORDER BY periodo DESC, propietario ASC, canal ASC`, where), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCRMMetaComercial{}
	for rows.Next() {
		var x EmpresaCRMMetaComercial
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Periodo, &x.Propietario, &x.Canal, &x.MetaValor, &x.MetaLeads, &x.MetaConversionPct, &x.Estado, &x.UsuarioCreador, &x.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func BuildEmpresaCRMVentasAvanzadasDashboard(dbConn *sql.DB, empresaID int64, periodo string) (EmpresaCRMVentasAvanzadasDashboard, error) {
	if strings.TrimSpace(periodo) == "" {
		periodo = time.Now().Format("2006-01")
	}
	d := EmpresaCRMVentasAvanzadasDashboard{EmpresaID: empresaID, Periodo: periodo}
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM crm_leads WHERE empresa_id=? AND estado='activo' AND estado_lead NOT IN ('ganado','perdido','descalificado','cerrado')`, empresaID).Scan(&d.LeadsActivos)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM crm_leads WHERE empresa_id=? AND estado='activo' AND estado_lead='ganado'`, empresaID).Scan(&d.LeadsGanados)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM crm_leads WHERE empresa_id=? AND estado='activo' AND estado_lead IN ('perdido','descalificado')`, empresaID).Scan(&d.LeadsPerdidos)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM crm_leads WHERE empresa_id=? AND estado='activo' AND estado_lead NOT IN ('ganado','perdido','descalificado','cerrado') AND COALESCE(proximo_contacto,'')<>'' AND pcs_ts(proximo_contacto) < CURRENT_TIMESTAMP`, empresaID).Scan(&d.LeadsVencidos)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM crm_leads l LEFT JOIN (SELECT empresa_id, lead_id, COUNT(1) AS interacciones FROM crm_interacciones WHERE empresa_id=? AND estado='activo' GROUP BY empresa_id, lead_id) i ON i.empresa_id=l.empresa_id AND i.lead_id=l.id WHERE l.empresa_id=? AND l.estado='activo' AND l.estado_lead NOT IN ('ganado','perdido','descalificado','cerrado') AND COALESCE(i.interacciones,0)=0`, empresaID, empresaID).Scan(&d.LeadsSinContacto)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1), COALESCE(SUM(valor_potencial),0) FROM crm_leads WHERE empresa_id=? AND estado='activo' AND estado_lead NOT IN ('ganado','perdido','descalificado','cerrado') AND COALESCE(fecha_actualizacion,fecha_creacion,'')<>'' AND pcs_ts(COALESCE(fecha_actualizacion,fecha_creacion)) < pcs_ts('now','-14 days')`, empresaID).Scan(&d.LeadsEstancados, &d.ValorRiesgo)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM crm_campanas WHERE empresa_id=? AND estado='activo' AND estado_campana='activa'`, empresaID).Scan(&d.CampanasActivas)
	_ = QueryRowCompat(dbConn, `SELECT COALESCE(SUM(valor_potencial),0), COALESCE(SUM(valor_potencial*(COALESCE(probabilidad,0)/100.0)),0) FROM crm_leads WHERE empresa_id=? AND estado='activo' AND estado_lead NOT IN ('perdido','descalificado','cerrado')`, empresaID).Scan(&d.ValorPipeline, &d.ForecastPonderado)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1), COALESCE(SUM(total),0) FROM empresa_cotizaciones_venta WHERE empresa_id=? AND estado='activo' AND estado_documento IN ('borrador','emitida','aprobada')`, empresaID).Scan(&d.CotizacionesAbiertas, &d.CotizacionesValor)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1), COALESCE(SUM(total),0) FROM empresa_pedidos_venta WHERE empresa_id=? AND estado='activo' AND estado_pedido NOT IN ('cerrado','cancelado')`, empresaID).Scan(&d.PedidosAbiertos, &d.PedidosValor)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM (
		SELECT proximo_contacto AS fecha FROM crm_leads WHERE empresa_id=? AND estado='activo' AND COALESCE(proximo_contacto,'')<>''
		UNION ALL
		SELECT proxima_accion AS fecha FROM crm_interacciones WHERE empresa_id=? AND estado='activo' AND COALESCE(proxima_accion,'')<>''
	) WHERE substr(fecha,1,10)=CURRENT_DATE`, empresaID, empresaID).Scan(&d.AgendaHoy)
	metas, err := ListEmpresaCRMMetasComerciales(dbConn, empresaID, periodo)
	if err != nil {
		return d, err
	}
	d.Metas = metas
	for _, meta := range metas {
		if meta.Estado == "activo" {
			d.MetaValor += meta.MetaValor
		}
	}
	if d.MetaValor > 0 {
		d.CumplimientoMetaPct = crmRound((d.ForecastPonderado / d.MetaValor) * 100)
	}
	d.ConversionPct = crmConversionPct(d.LeadsGanados, d.LeadsPerdidos)
	if d.LeadsGanados > 0 {
		d.TicketPromedio = crmRound(d.PedidosValor / float64(d.LeadsGanados))
	}
	embudo, err := buildEmpresaCRMEmbudo(dbConn, empresaID)
	if err != nil {
		return d, err
	}
	d.Embudo = embudo
	agenda, err := buildEmpresaCRMAgenda(dbConn, empresaID)
	if err != nil {
		return d, err
	}
	d.Agenda = agenda
	top, err := ListEmpresaCRMLeadScores(dbConn, empresaID, 10)
	if err != nil {
		return d, err
	}
	d.TopLeads = top
	responsables, err := buildEmpresaCRMResponsables(dbConn, empresaID)
	if err != nil {
		return d, err
	}
	d.Responsables = responsables
	canales, err := buildEmpresaCRMCanales(dbConn, empresaID)
	if err != nil {
		return d, err
	}
	d.Canales = canales
	d.SaludComercialPct = crmCommercialHealthPct(d)
	d.Alertas = buildEmpresaCRMAlertas(d)
	d.AccionesPrioritarias = buildEmpresaCRMAccionesPrioritarias(d)
	return d, nil
}

func ListEmpresaCRMLeadScores(dbConn *sql.DB, empresaID int64, limit int) ([]EmpresaCRMLeadScore, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT l.id,l.codigo,COALESCE(l.nombre,''),COALESCE(l.empresa_origen,''),COALESCE(l.estado_lead,'nuevo'),COALESCE(l.valor_potencial,0),COALESCE(l.probabilidad,0),COALESCE(l.proximo_contacto,''),COALESCE(i.interacciones,0)
		FROM crm_leads l
		LEFT JOIN (SELECT empresa_id, lead_id, COUNT(1) AS interacciones FROM crm_interacciones WHERE empresa_id=? AND estado='activo' GROUP BY empresa_id, lead_id) i ON i.empresa_id=l.empresa_id AND i.lead_id=l.id
		WHERE l.empresa_id=? AND l.estado='activo'
		ORDER BY COALESCE(l.valor_potencial,0)*(COALESCE(l.probabilidad,0)/100.0) DESC, l.id DESC
		LIMIT %d`, limit), empresaID, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCRMLeadScore{}
	for rows.Next() {
		var x EmpresaCRMLeadScore
		if err := rows.Scan(&x.ID, &x.Codigo, &x.Nombre, &x.EmpresaOrigen, &x.EstadoLead, &x.ValorPotencial, &x.Probabilidad, &x.ProximoContacto, &x.Interacciones); err != nil {
			return nil, err
		}
		x.Score = crmLeadScore(x)
		x.Recomendacion = crmLeadRecommendation(x)
		out = append(out, x)
	}
	return out, rows.Err()
}

func CreateEmpresaCRMCotizacionDesdeLead(dbConn *sql.DB, empresaID, leadID int64, codigo, usuario string) (int64, error) {
	var lead EmpresaCRMLeadScore
	if err := QueryRowCompat(dbConn, `SELECT id,codigo,COALESCE(nombre,''),COALESCE(empresa_origen,''),COALESCE(estado_lead,'nuevo'),COALESCE(valor_potencial,0),COALESCE(probabilidad,0),COALESCE(proximo_contacto,'') FROM crm_leads WHERE empresa_id=? AND id=? AND estado='activo'`, empresaID, leadID).Scan(&lead.ID, &lead.Codigo, &lead.Nombre, &lead.EmpresaOrigen, &lead.EstadoLead, &lead.ValorPotencial, &lead.Probabilidad, &lead.ProximoContacto); err != nil {
		return 0, err
	}
	codigo = strings.ToUpper(strings.TrimSpace(codigo))
	if codigo == "" {
		codigo = "COT-CRM-" + time.Now().Format("20060102150405")
	}
	nombre := strings.TrimSpace(lead.Nombre)
	if nombre == "" {
		nombre = strings.TrimSpace(lead.EmpresaOrigen)
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_cotizaciones_venta
		(empresa_id,codigo,cliente_nombre,fecha_documento,vigencia_hasta,estado_documento,subtotal,total,moneda,notas,origen,usuario_creador,estado,observaciones)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		empresaID, codigo, nombre, time.Now().Format("2006-01-02"), time.Now().AddDate(0, 0, 15).Format("2006-01-02"), "emitida", lead.ValorPotencial, lead.ValorPotencial, "COP", "Cotizacion generada desde lead "+lead.Codigo, "crm_avanzado", strings.TrimSpace(usuario), "activo", "lead_id="+fmt.Sprint(leadID))
	if err != nil {
		return 0, err
	}
	_, _ = ExecCompat(dbConn, `UPDATE crm_leads SET estado_lead='propuesta', fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=? AND estado_lead IN ('nuevo','contactado','calificado')`, empresaID, leadID)
	return id, nil
}

func SeedEmpresaCRMVentasAvanzadasDemo(dbConn *sql.DB, empresaID int64, usuario string) (int64, error) {
	if err := EnsureEmpresaCRMVentasAvanzadasSchema(dbConn); err != nil {
		return 0, err
	}
	run := time.Now().Format("20060102150405")
	leadID, err := insertSQLCompat(dbConn, `INSERT INTO crm_leads
		(empresa_id,codigo,nombre,empresa_origen,email,telefono,canal_origen,estado_lead,valor_potencial,probabilidad,propietario,proximo_contacto,notas,usuario_creador,estado)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (empresa_id,codigo) DO UPDATE SET
			nombre=EXCLUDED.nombre,
			empresa_origen=EXCLUDED.empresa_origen,
			valor_potencial=EXCLUDED.valor_potencial,
			probabilidad=EXCLUDED.probabilidad,
			proximo_contacto=EXCLUDED.proximo_contacto,
			fecha_actualizacion=CURRENT_TIMESTAMP`,
		empresaID, "LEAD-CRM-"+run, "Cliente QA CRM "+run, "Empresa QA Comercial", "crm.qa@calipso.local", "3000000000", "web", "calificado", 1850000, 65, usuario, time.Now().Add(24*time.Hour).Format("2006-01-02 15:04:05"), "QA CRM avanzado", usuario, "activo")
	if err != nil {
		return 0, err
	}
	if _, err := insertSQLCompat(dbConn, `INSERT INTO crm_interacciones
		(empresa_id,codigo,lead_id,tipo_interaccion,fecha_interaccion,resumen,resultado,usuario_responsable,proxima_accion,estado_interaccion,usuario_creador,estado)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
		empresaID, "INT-CRM-"+run, leadID, "llamada", time.Now().Format("2006-01-02 15:04:05"), "Contacto QA CRM avanzado", "Interesado en paquete empresarial", usuario, time.Now().Add(48*time.Hour).Format("2006-01-02 15:04:05"), "cerrada", usuario, "activo"); err != nil {
		return 0, err
	}
	if _, err := UpsertEmpresaCRMMetaComercial(dbConn, EmpresaCRMMetaComercial{EmpresaID: empresaID, Periodo: time.Now().Format("2006-01"), Propietario: usuario, Canal: "web", MetaValor: 2500000, MetaLeads: 12, MetaConversionPct: 25, Estado: "activo", UsuarioCreador: usuario}); err != nil {
		return 0, err
	}
	if _, err := CreateEmpresaCRMCotizacionDesdeLead(dbConn, empresaID, leadID, "COT-CRM-"+run, usuario); err != nil {
		return 0, err
	}
	return leadID, nil
}

func buildEmpresaCRMEmbudo(dbConn *sql.DB, empresaID int64) ([]EmpresaCRMEmbudoEstado, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT COALESCE(estado_lead,'nuevo'), COUNT(1), COALESCE(SUM(valor_potencial),0), COALESCE(SUM(valor_potencial*(COALESCE(probabilidad,0)/100.0)),0), COALESCE(AVG(probabilidad),0) FROM crm_leads WHERE empresa_id=? AND estado='activo' GROUP BY COALESCE(estado_lead,'nuevo') ORDER BY 2 DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCRMEmbudoEstado{}
	for rows.Next() {
		var x EmpresaCRMEmbudoEstado
		if err := rows.Scan(&x.Estado, &x.Leads, &x.Valor, &x.Forecast, &x.ProbabilidadPromedio); err != nil {
			return nil, err
		}
		x.Valor = crmRound(x.Valor)
		x.Forecast = crmRound(x.Forecast)
		x.ProbabilidadPromedio = crmRound(x.ProbabilidadPromedio)
		out = append(out, x)
	}
	return out, rows.Err()
}

func buildEmpresaCRMAgenda(dbConn *sql.DB, empresaID int64) ([]EmpresaCRMAgendaItem, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT 'lead', codigo, COALESCE(nombre,''), COALESCE(propietario,''), COALESCE(proximo_contacto,''), COALESCE(estado_lead,''), COALESCE(valor_potencial,0) FROM crm_leads WHERE empresa_id=? AND estado='activo' AND COALESCE(proximo_contacto,'')<>'' UNION ALL SELECT 'interaccion', codigo, COALESCE(resumen,''), COALESCE(usuario_responsable,''), COALESCE(proxima_accion,''), COALESCE(estado_interaccion,''), 0 FROM crm_interacciones WHERE empresa_id=? AND estado='activo' AND COALESCE(proxima_accion,'')<>'' ORDER BY 5 ASC LIMIT 20`, empresaID, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCRMAgendaItem{}
	for rows.Next() {
		var x EmpresaCRMAgendaItem
		if err := rows.Scan(&x.Tipo, &x.Referencia, &x.Nombre, &x.Responsable, &x.Fecha, &x.Estado, &x.Valor); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func buildEmpresaCRMResponsables(dbConn *sql.DB, empresaID int64) ([]EmpresaCRMResponsableRendimiento, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT COALESCE(NULLIF(TRIM(propietario),''),'Sin asignar') AS responsable,
		COUNT(1),
		SUM(CASE WHEN COALESCE(proximo_contacto,'')<>'' AND pcs_ts(proximo_contacto) < CURRENT_TIMESTAMP THEN 1 ELSE 0 END),
		COALESCE(SUM(valor_potencial),0),
		COALESCE(SUM(valor_potencial*(COALESCE(probabilidad,0)/100.0)),0),
		COALESCE(AVG(probabilidad),0)
		FROM crm_leads
		WHERE empresa_id=? AND estado='activo' AND estado_lead NOT IN ('ganado','perdido','descalificado','cerrado')
		GROUP BY COALESCE(NULLIF(TRIM(propietario),''),'Sin asignar')
		ORDER BY 5 DESC, 2 DESC
		LIMIT 12`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCRMResponsableRendimiento{}
	for rows.Next() {
		var x EmpresaCRMResponsableRendimiento
		if err := rows.Scan(&x.Responsable, &x.LeadsActivos, &x.LeadsVencidos, &x.ValorPipeline, &x.ForecastPonderado, &x.ProbabilidadPromedio); err != nil {
			return nil, err
		}
		x.ValorPipeline = crmRound(x.ValorPipeline)
		x.ForecastPonderado = crmRound(x.ForecastPonderado)
		x.ProbabilidadPromedio = crmRound(x.ProbabilidadPromedio)
		out = append(out, x)
	}
	return out, rows.Err()
}

func buildEmpresaCRMCanales(dbConn *sql.DB, empresaID int64) ([]EmpresaCRMCanalRendimiento, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT COALESCE(NULLIF(TRIM(canal_origen),''),'Sin canal') AS canal,
		COUNT(1),
		SUM(CASE WHEN estado_lead='ganado' THEN 1 ELSE 0 END),
		SUM(CASE WHEN estado_lead IN ('perdido','descalificado') THEN 1 ELSE 0 END),
		COALESCE(SUM(valor_potencial),0),
		COALESCE(SUM(valor_potencial*(COALESCE(probabilidad,0)/100.0)),0)
		FROM crm_leads
		WHERE empresa_id=? AND estado='activo'
		GROUP BY COALESCE(NULLIF(TRIM(canal_origen),''),'Sin canal')
		ORDER BY 6 DESC, 2 DESC
		LIMIT 12`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCRMCanalRendimiento{}
	for rows.Next() {
		var x EmpresaCRMCanalRendimiento
		if err := rows.Scan(&x.Canal, &x.Leads, &x.Ganados, &x.Perdidos, &x.ValorPipeline, &x.ForecastPonderado); err != nil {
			return nil, err
		}
		x.ValorPipeline = crmRound(x.ValorPipeline)
		x.ForecastPonderado = crmRound(x.ForecastPonderado)
		x.ConversionPct = crmConversionPct(x.Ganados, x.Perdidos)
		out = append(out, x)
	}
	return out, rows.Err()
}

func normalizeCRMMetaComercial(meta EmpresaCRMMetaComercial) EmpresaCRMMetaComercial {
	meta.Periodo = strings.TrimSpace(meta.Periodo)
	if meta.Periodo == "" {
		meta.Periodo = time.Now().Format("2006-01")
	}
	meta.Propietario = strings.TrimSpace(meta.Propietario)
	meta.Canal = strings.ToLower(strings.TrimSpace(meta.Canal))
	meta.MetaValor = crmRound(math.Max(0, meta.MetaValor))
	if meta.MetaLeads < 0 {
		meta.MetaLeads = 0
	}
	meta.MetaConversionPct = crmRound(math.Max(0, math.Min(100, meta.MetaConversionPct)))
	meta.Estado = strings.ToLower(strings.TrimSpace(meta.Estado))
	if meta.Estado != "inactivo" {
		meta.Estado = "activo"
	}
	meta.UsuarioCreador = strings.TrimSpace(meta.UsuarioCreador)
	return meta
}

func crmLeadScore(x EmpresaCRMLeadScore) float64 {
	score := x.Probabilidad
	if x.ValorPotencial >= 5000000 {
		score += 20
	} else if x.ValorPotencial >= 1000000 {
		score += 12
	}
	if x.Interacciones >= 3 {
		score += 10
	} else if x.Interacciones > 0 {
		score += 5
	}
	switch x.EstadoLead {
	case "negociacion", "propuesta":
		score += 10
	case "calificado":
		score += 6
	}
	return crmRound(math.Max(0, math.Min(100, score)))
}

func crmLeadRecommendation(x EmpresaCRMLeadScore) string {
	if x.Score >= 80 {
		return "priorizar_cierre"
	}
	if x.Score >= 60 {
		return "enviar_propuesta"
	}
	if x.Interacciones == 0 {
		return "contactar"
	}
	return "nutrir_seguimiento"
}

func crmCommercialHealthPct(d EmpresaCRMVentasAvanzadasDashboard) float64 {
	score := 100.0
	score -= math.Min(30, float64(d.LeadsVencidos)*6)
	score -= math.Min(20, float64(d.LeadsSinContacto)*4)
	score -= math.Min(20, float64(d.LeadsEstancados)*5)
	if d.MetaValor > 0 && d.CumplimientoMetaPct < 80 {
		score -= math.Min(18, (80-d.CumplimientoMetaPct)*0.45)
	}
	if d.CampanasActivas == 0 {
		score -= 8
	}
	if d.LeadsActivos == 0 {
		score -= 20
	}
	return crmRound(math.Max(0, math.Min(100, score)))
}

func buildEmpresaCRMAlertas(d EmpresaCRMVentasAvanzadasDashboard) []string {
	alertas := []string{}
	if d.LeadsVencidos > 0 {
		alertas = append(alertas, fmt.Sprintf("%d leads tienen seguimiento vencido.", d.LeadsVencidos))
	}
	if d.LeadsSinContacto > 0 {
		alertas = append(alertas, fmt.Sprintf("%d leads activos aun no tienen interaccion registrada.", d.LeadsSinContacto))
	}
	if d.LeadsEstancados > 0 {
		alertas = append(alertas, fmt.Sprintf("%d oportunidades llevan mas de 14 dias sin actualizacion.", d.LeadsEstancados))
	}
	if d.MetaValor <= 0 {
		alertas = append(alertas, "No hay meta comercial activa para el periodo.")
	} else if d.CumplimientoMetaPct < 70 {
		alertas = append(alertas, fmt.Sprintf("El forecast cubre %.0f%% de la meta del periodo.", d.CumplimientoMetaPct))
	}
	if d.LeadsActivos == 0 {
		alertas = append(alertas, "No hay leads activos en el embudo comercial.")
	}
	if d.CampanasActivas == 0 {
		alertas = append(alertas, "No hay campanas activas alimentando el embudo.")
	}
	return alertas
}

func buildEmpresaCRMAccionesPrioritarias(d EmpresaCRMVentasAvanzadasDashboard) []EmpresaCRMAccionPrioritaria {
	out := []EmpresaCRMAccionPrioritaria{}
	if d.LeadsVencidos > 0 {
		out = append(out, EmpresaCRMAccionPrioritaria{
			Prioridad: 1,
			Severidad: "alta",
			Titulo:    "Recuperar seguimientos vencidos",
			Detalle:   fmt.Sprintf("%d leads tienen fecha de contacto vencida y pueden enfriar el pipeline.", d.LeadsVencidos),
			Valor:     d.ValorRiesgo,
			Accion:    "Reasignar responsable o registrar contacto hoy",
		})
	}
	if d.LeadsSinContacto > 0 {
		out = append(out, EmpresaCRMAccionPrioritaria{
			Prioridad: 2,
			Severidad: "media",
			Titulo:    "Asignar primer contacto",
			Detalle:   fmt.Sprintf("%d leads activos no tienen interaccion registrada.", d.LeadsSinContacto),
			Accion:    "Crear seguimiento inicial con responsable y proxima accion",
		})
	}
	if d.LeadsEstancados > 0 {
		out = append(out, EmpresaCRMAccionPrioritaria{
			Prioridad: 3,
			Severidad: "media",
			Titulo:    "Reactivar oportunidades estancadas",
			Detalle:   fmt.Sprintf("%d oportunidades llevan mas de 14 dias sin actualizacion.", d.LeadsEstancados),
			Valor:     d.ValorRiesgo,
			Accion:    "Actualizar etapa, descartar o programar cierre",
		})
	}
	if d.MetaValor > 0 && d.CumplimientoMetaPct < 70 {
		out = append(out, EmpresaCRMAccionPrioritaria{
			Prioridad: 4,
			Severidad: "media",
			Titulo:    "Cerrar brecha contra meta",
			Detalle:   fmt.Sprintf("El forecast ponderado cubre %.0f%% de la meta del periodo.", d.CumplimientoMetaPct),
			Accion:    "Priorizar leads con score alto y activar campanas de demanda",
		})
	}
	for _, lead := range d.TopLeads {
		if len(out) >= 8 {
			break
		}
		if lead.Recomendacion != "priorizar_cierre" {
			continue
		}
		nombre := strings.TrimSpace(lead.Nombre)
		if nombre == "" {
			nombre = strings.TrimSpace(lead.EmpresaOrigen)
		}
		out = append(out, EmpresaCRMAccionPrioritaria{
			Prioridad:   5 + len(out),
			Severidad:   "alta",
			Titulo:      "Cerrar oportunidad de alto score",
			Detalle:     nombre,
			Responsable: "",
			Referencia:  lead.Codigo,
			Fecha:       lead.ProximoContacto,
			Valor:       lead.ValorPotencial,
			Accion:      "Enviar propuesta final o confirmar decision de compra",
		})
	}
	if len(out) == 0 {
		out = append(out, EmpresaCRMAccionPrioritaria{
			Prioridad: 1,
			Severidad: "baja",
			Titulo:    "Mantener cadencia comercial",
			Detalle:   "El CRM no tiene alertas criticas para el periodo.",
			Accion:    "Revisar pipeline y registrar proximas acciones semanalmente",
		})
	}
	return out
}

func crmConversionPct(ganados, perdidos int) float64 {
	total := ganados + perdidos
	if total <= 0 {
		return 0
	}
	return crmRound((float64(ganados) / float64(total)) * 100)
}

func crmRound(v float64) float64 {
	return math.Round(v*100) / 100
}
