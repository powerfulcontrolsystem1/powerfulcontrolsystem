package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type EmpresaCobranzaCuentaFiltro struct {
	Estado  string
	Query   string
	MoraMin int
	Limit   int
}

type EmpresaCobranzaCuenta struct {
	ID               int64   `json:"id"`
	EmpresaID        int64   `json:"empresa_id"`
	Codigo           string  `json:"codigo"`
	ClienteID        int64   `json:"cliente_id"`
	ClienteNombre    string  `json:"cliente_nombre"`
	DocumentoTipo    string  `json:"documento_tipo"`
	DocumentoCodigo  string  `json:"documento_codigo"`
	FechaEmision     string  `json:"fecha_emision"`
	FechaVencimiento string  `json:"fecha_vencimiento"`
	DiasMora         int     `json:"dias_mora"`
	ValorOriginal    float64 `json:"valor_original"`
	ValorPagado      float64 `json:"valor_pagado"`
	Saldo            float64 `json:"saldo"`
	EstadoCartera    string  `json:"estado_cartera"`
	Moneda           string  `json:"moneda"`
	Observaciones    string  `json:"observaciones"`
}

type EmpresaCobranzaPlantilla struct {
	ID            int64  `json:"id"`
	EmpresaID     int64  `json:"empresa_id"`
	Codigo        string `json:"codigo"`
	Nombre        string `json:"nombre"`
	Canal         string `json:"canal"`
	Asunto        string `json:"asunto"`
	Cuerpo        string `json:"cuerpo"`
	DiasMoraDesde int    `json:"dias_mora_desde"`
	DiasMoraHasta int    `json:"dias_mora_hasta"`
	Prioridad     int    `json:"prioridad"`
	Activa        bool   `json:"activa"`
	Usuario       string `json:"usuario_creador"`
	Estado        string `json:"estado"`
	Observaciones string `json:"observaciones"`
	FechaCreacion string `json:"fecha_creacion"`
}

type EmpresaCobranzaCampana struct {
	ID                   int64   `json:"id"`
	EmpresaID            int64   `json:"empresa_id"`
	Codigo               string  `json:"codigo"`
	Nombre               string  `json:"nombre"`
	Tipo                 string  `json:"tipo"`
	CanalPrincipal       string  `json:"canal_principal"`
	Segmento             string  `json:"segmento"`
	FechaInicio          string  `json:"fecha_inicio"`
	FechaFin             string  `json:"fecha_fin"`
	EstadoCampana        string  `json:"estado_campana"`
	MetaRecaudo          float64 `json:"meta_recaudo"`
	ValorAsignado        float64 `json:"valor_asignado"`
	ValorRecuperado      float64 `json:"valor_recuperado"`
	ContactosProgramados int     `json:"contactos_programados"`
	ContactosRealizados  int     `json:"contactos_realizados"`
	Usuario              string  `json:"usuario_creador"`
	Estado               string  `json:"estado"`
	Observaciones        string  `json:"observaciones"`
	FechaCreacion        string  `json:"fecha_creacion"`
}

type EmpresaCobranzaGestion struct {
	ID                   int64   `json:"id"`
	EmpresaID            int64   `json:"empresa_id"`
	CuentaID             int64   `json:"cuenta_id"`
	CampanaID            int64   `json:"campana_id"`
	PlantillaID          int64   `json:"plantilla_id"`
	ClienteID            int64   `json:"cliente_id"`
	ClienteNombre        string  `json:"cliente_nombre"`
	DocumentoCodigo      string  `json:"documento_codigo"`
	Canal                string  `json:"canal"`
	Resultado            string  `json:"resultado"`
	FechaGestion         string  `json:"fecha_gestion"`
	FechaProximoContacto string  `json:"fecha_proximo_contacto"`
	ValorCompromiso      float64 `json:"valor_compromiso"`
	PromesaFecha         string  `json:"promesa_fecha"`
	PromesaEstado        string  `json:"promesa_estado"`
	Mensaje              string  `json:"mensaje"`
	Contacto             string  `json:"contacto"`
	Usuario              string  `json:"usuario_creador"`
	Estado               string  `json:"estado"`
	Observaciones        string  `json:"observaciones"`
	FechaCreacion        string  `json:"fecha_creacion"`
}

type EmpresaCobranzaPromesa struct {
	ID                int64   `json:"id"`
	EmpresaID         int64   `json:"empresa_id"`
	CuentaID          int64   `json:"cuenta_id"`
	GestionID         int64   `json:"gestion_id"`
	ClienteNombre     string  `json:"cliente_nombre"`
	DocumentoCodigo   string  `json:"documento_codigo"`
	ValorPrometido    float64 `json:"valor_prometido"`
	FechaPromesa      string  `json:"fecha_promesa"`
	EstadoPromesa     string  `json:"estado_promesa"`
	FechaCumplimiento string  `json:"fecha_cumplimiento"`
	Usuario           string  `json:"usuario_creador"`
	Observaciones     string  `json:"observaciones"`
	FechaCreacion     string  `json:"fecha_creacion"`
}

type EmpresaCobranzaDashboard struct {
	EmpresaID           int64                    `json:"empresa_id"`
	SaldoTotal          float64                  `json:"saldo_total"`
	SaldoVencido        float64                  `json:"saldo_vencido"`
	SaldoPorVencer      float64                  `json:"saldo_por_vencer"`
	SaldoMoraCritica    float64                  `json:"saldo_mora_critica"`
	CuentasTotal        int                      `json:"cuentas_total"`
	CuentasVencidas     int                      `json:"cuentas_vencidas"`
	CuentasPorVencer    int                      `json:"cuentas_por_vencer"`
	PromesasPendientes  int                      `json:"promesas_pendientes"`
	PromesasIncumplidas int                      `json:"promesas_incumplidas"`
	GestionesHoy        int                      `json:"gestiones_hoy"`
	CampanasActivas     int                      `json:"campanas_activas"`
	RecuperadoMes       float64                  `json:"recuperado_mes"`
	CuentasPrioritarias []EmpresaCobranzaCuenta  `json:"cuentas_prioritarias"`
	UltimasGestiones    []EmpresaCobranzaGestion `json:"ultimas_gestiones"`
	Campanas            []EmpresaCobranzaCampana `json:"campanas"`
	Alertas             []string                 `json:"alertas"`
}

type EmpresaCobranzaConfiguracion struct {
	EmpresaID       int64  `json:"empresa_id"`
	AutoActivo      bool   `json:"auto_activo"`
	EmailActivo     bool   `json:"email_activo"`
	WhatsAppActivo  bool   `json:"whatsapp_activo"`
	DiasAntes       int    `json:"dias_antes"`
	FrecuenciaDias  int    `json:"frecuencia_dias"`
	Asunto          string `json:"asunto"`
	Mensaje         string `json:"mensaje"`
	HoraLocal       string `json:"hora_local"`
	UltimaEjecucion string `json:"ultima_ejecucion,omitempty"`
	Usuario         string `json:"usuario_creador,omitempty"`
}

func EnsureEmpresaCobranzaSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if err := EnsureEmpresaModulosFaltantesSchema(dbConn); err != nil {
		return err
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_cobranza_plantillas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			canal TEXT DEFAULT 'email',
			asunto TEXT,
			cuerpo TEXT,
			dias_mora_desde INTEGER DEFAULT 0,
			dias_mora_hasta INTEGER DEFAULT 9999,
			prioridad INTEGER DEFAULT 1,
			activa INTEGER DEFAULT 1,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_cobranza_plantillas_empresa ON empresa_cobranza_plantillas(empresa_id, canal, activa);`,
		`CREATE TABLE IF NOT EXISTS empresa_cobranza_campanas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			tipo TEXT DEFAULT 'preventiva',
			canal_principal TEXT DEFAULT 'whatsapp',
			segmento TEXT DEFAULT 'todas',
			fecha_inicio TEXT,
			fecha_fin TEXT,
			estado_campana TEXT DEFAULT 'borrador',
			meta_recaudo REAL DEFAULT 0,
			valor_asignado REAL DEFAULT 0,
			valor_recuperado REAL DEFAULT 0,
			contactos_programados INTEGER DEFAULT 0,
			contactos_realizados INTEGER DEFAULT 0,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_cobranza_campanas_empresa ON empresa_cobranza_campanas(empresa_id, estado_campana, fecha_inicio);`,
		`CREATE TABLE IF NOT EXISTS empresa_cobranza_gestiones (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			cuenta_id INTEGER DEFAULT 0,
			campana_id INTEGER DEFAULT 0,
			plantilla_id INTEGER DEFAULT 0,
			cliente_id INTEGER DEFAULT 0,
			cliente_nombre TEXT,
			documento_codigo TEXT,
			canal TEXT DEFAULT 'llamada',
			resultado TEXT DEFAULT 'registrada',
			fecha_gestion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_proximo_contacto TEXT,
			valor_compromiso REAL DEFAULT 0,
			promesa_fecha TEXT,
			promesa_estado TEXT DEFAULT 'sin_promesa',
			mensaje TEXT,
			contacto TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_cobranza_gestiones_empresa ON empresa_cobranza_gestiones(empresa_id, fecha_gestion DESC, resultado);`,
		`CREATE TABLE IF NOT EXISTS empresa_cobranza_promesas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			cuenta_id INTEGER DEFAULT 0,
			gestion_id INTEGER DEFAULT 0,
			cliente_nombre TEXT,
			documento_codigo TEXT,
			valor_prometido REAL DEFAULT 0,
			fecha_promesa TEXT,
			estado_promesa TEXT DEFAULT 'pendiente',
			fecha_cumplimiento TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_cobranza_promesas_empresa ON empresa_cobranza_promesas(empresa_id, estado_promesa, fecha_promesa);`,
		`CREATE TABLE IF NOT EXISTS empresa_cobranza_configuracion (
			empresa_id INTEGER PRIMARY KEY,
			auto_activo INTEGER DEFAULT 0,
			email_activo INTEGER DEFAULT 1,
			whatsapp_activo INTEGER DEFAULT 0,
			dias_antes INTEGER DEFAULT 1,
			frecuencia_dias INTEGER DEFAULT 3,
			asunto TEXT DEFAULT 'Recordatorio de pago',
			mensaje TEXT DEFAULT 'Hola {{cliente}}, el documento {{documento}} registra un saldo de {{saldo}} con vencimiento {{vencimiento}}.',
			hora_local TEXT DEFAULT '09:00',
			ultima_ejecucion TEXT,
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS empresa_cobranza_envios (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			cuenta_id INTEGER NOT NULL,
			canal TEXT NOT NULL,
			dedupe_key TEXT NOT NULL,
			destino TEXT,
			resultado TEXT,
			detalle TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			UNIQUE(empresa_id, dedupe_key)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_cobranza_envios_empresa ON empresa_cobranza_envios(empresa_id, cuenta_id, canal, fecha_creacion DESC);`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func GetEmpresaCobranzaConfiguracion(dbConn *sql.DB, empresaID int64) (EmpresaCobranzaConfiguracion, error) {
	if err := EnsureEmpresaCobranzaSchema(dbConn); err != nil {
		return EmpresaCobranzaConfiguracion{}, err
	}
	cfg := EmpresaCobranzaConfiguracion{EmpresaID: empresaID, EmailActivo: true, DiasAntes: 1, FrecuenciaDias: 3, Asunto: "Recordatorio de pago", Mensaje: "Hola {{cliente}}, el documento {{documento}} registra un saldo de {{saldo}} con vencimiento {{vencimiento}}.", HoraLocal: "09:00"}
	var autoActivo, emailActivo, whatsappActivo int
	err := QueryRowCompat(dbConn, `SELECT empresa_id,COALESCE(auto_activo,0),COALESCE(email_activo,1),COALESCE(whatsapp_activo,0),COALESCE(dias_antes,1),COALESCE(frecuencia_dias,3),COALESCE(asunto,'Recordatorio de pago'),COALESCE(mensaje,''),COALESCE(hora_local,'09:00'),COALESCE(ultima_ejecucion,''),COALESCE(usuario_creador,'') FROM empresa_cobranza_configuracion WHERE empresa_id=?`, empresaID).Scan(&cfg.EmpresaID, &autoActivo, &emailActivo, &whatsappActivo, &cfg.DiasAntes, &cfg.FrecuenciaDias, &cfg.Asunto, &cfg.Mensaje, &cfg.HoraLocal, &cfg.UltimaEjecucion, &cfg.Usuario)
	if err == sql.ErrNoRows {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	cfg.AutoActivo = autoActivo != 0
	cfg.EmailActivo = emailActivo != 0
	cfg.WhatsAppActivo = whatsappActivo != 0
	if cfg.DiasAntes < 0 {
		cfg.DiasAntes = 0
	}
	if cfg.DiasAntes > 90 {
		cfg.DiasAntes = 90
	}
	if cfg.FrecuenciaDias < 1 {
		cfg.FrecuenciaDias = 1
	}
	if cfg.FrecuenciaDias > 90 {
		cfg.FrecuenciaDias = 90
	}
	return cfg, nil
}

func UpsertEmpresaCobranzaConfiguracion(dbConn *sql.DB, cfg EmpresaCobranzaConfiguracion) (EmpresaCobranzaConfiguracion, error) {
	if cfg.EmpresaID <= 0 {
		return cfg, errors.New("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaCobranzaSchema(dbConn); err != nil {
		return cfg, err
	}
	if cfg.DiasAntes < 0 {
		cfg.DiasAntes = 0
	}
	if cfg.DiasAntes > 90 {
		cfg.DiasAntes = 90
	}
	if cfg.FrecuenciaDias < 1 {
		cfg.FrecuenciaDias = 1
	}
	if cfg.FrecuenciaDias > 90 {
		cfg.FrecuenciaDias = 90
	}
	if strings.TrimSpace(cfg.Asunto) == "" {
		cfg.Asunto = "Recordatorio de pago"
	}
	if strings.TrimSpace(cfg.Mensaje) == "" {
		cfg.Mensaje = "Hola {{cliente}}, el documento {{documento}} registra un saldo de {{saldo}} con vencimiento {{vencimiento}}."
	}
	if strings.TrimSpace(cfg.HoraLocal) == "" {
		cfg.HoraLocal = "09:00"
	}
	_, err := ExecCompat(dbConn, `INSERT INTO empresa_cobranza_configuracion (empresa_id,auto_activo,email_activo,whatsapp_activo,dias_antes,frecuencia_dias,asunto,mensaje,hora_local,fecha_actualizacion,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP,?) ON CONFLICT (empresa_id) DO UPDATE SET auto_activo=excluded.auto_activo,email_activo=excluded.email_activo,whatsapp_activo=excluded.whatsapp_activo,dias_antes=excluded.dias_antes,frecuencia_dias=excluded.frecuencia_dias,asunto=excluded.asunto,mensaje=excluded.mensaje,hora_local=excluded.hora_local,fecha_actualizacion=CURRENT_TIMESTAMP,usuario_creador=excluded.usuario_creador`, cfg.EmpresaID, boolToInt(cfg.AutoActivo), boolToInt(cfg.EmailActivo), boolToInt(cfg.WhatsAppActivo), cfg.DiasAntes, cfg.FrecuenciaDias, strings.TrimSpace(cfg.Asunto), strings.TrimSpace(cfg.Mensaje), strings.TrimSpace(cfg.HoraLocal), strings.TrimSpace(cfg.Usuario))
	if err != nil {
		return cfg, err
	}
	return GetEmpresaCobranzaConfiguracion(dbConn, cfg.EmpresaID)
}

func ListEmpresaCobranzaConfiguracionesActivas(dbConn *sql.DB) ([]EmpresaCobranzaConfiguracion, error) {
	if err := EnsureEmpresaCobranzaSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT empresa_id FROM empresa_cobranza_configuracion WHERE COALESCE(auto_activo,0)=1 ORDER BY empresa_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCobranzaConfiguracion{}
	for rows.Next() {
		var empresaID int64
		if err := rows.Scan(&empresaID); err != nil {
			return nil, err
		}
		cfg, err := GetEmpresaCobranzaConfiguracion(dbConn, empresaID)
		if err != nil {
			return nil, err
		}
		out = append(out, cfg)
	}
	return out, rows.Err()
}

func RegisterEmpresaCobranzaEnvio(dbConn *sql.DB, empresaID, cuentaID int64, canal, dedupeKey, destino, resultado, detalle, usuario string) (bool, error) {
	res, err := ExecCompat(dbConn, `INSERT INTO empresa_cobranza_envios (empresa_id,cuenta_id,canal,dedupe_key,destino,resultado,detalle,usuario_creador) VALUES (?,?,?,?,?,?,?,?) ON CONFLICT (empresa_id,dedupe_key) DO NOTHING`, empresaID, cuentaID, strings.TrimSpace(canal), strings.TrimSpace(dedupeKey), strings.TrimSpace(destino), strings.TrimSpace(resultado), strings.TrimSpace(detalle), strings.TrimSpace(usuario))
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func UpdateEmpresaCobranzaEnvioResultado(dbConn *sql.DB, empresaID int64, dedupeKey, resultado, detalle string) error {
	_, err := ExecCompat(dbConn, `UPDATE empresa_cobranza_envios SET resultado=?,detalle=? WHERE empresa_id=? AND dedupe_key=?`, strings.TrimSpace(resultado), strings.TrimSpace(detalle), empresaID, strings.TrimSpace(dedupeKey))
	return err
}

func MarkEmpresaCobranzaUltimaEjecucion(dbConn *sql.DB, empresaID int64, value string) error {
	_, err := ExecCompat(dbConn, `UPDATE empresa_cobranza_configuracion SET ultima_ejecucion=?,fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=?`, strings.TrimSpace(value), empresaID)
	return err
}

func ListEmpresaCobranzaCuentas(dbConn *sql.DB, empresaID int64, filtro EmpresaCobranzaCuentaFiltro) ([]EmpresaCobranzaCuenta, error) {
	if err := EnsureEmpresaCobranzaSchema(dbConn); err != nil {
		return nil, err
	}
	return listEmpresaCobranzaCuentas(dbConn, empresaID, filtro)
}

func listEmpresaCobranzaCuentas(dbConn *sql.DB, empresaID int64, filtro EmpresaCobranzaCuentaFiltro) ([]EmpresaCobranzaCuenta, error) {
	limit := filtro.Limit
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	where := []string{"empresa_id=?", "COALESCE(estado,'activo')='activo'", "COALESCE(saldo,0)>0"}
	args := []interface{}{empresaID}
	estado := normalizeCobranzaEstadoCartera(filtro.Estado)
	if estado != "" && estado != "todas" {
		where = append(where, "LOWER(COALESCE(estado_cartera,''))=?")
		args = append(args, estado)
	}
	q := strings.ToLower(strings.TrimSpace(filtro.Query))
	if q != "" {
		where = append(where, "(LOWER(COALESCE(codigo,'')) LIKE ? OR LOWER(COALESCE(cliente_nombre,'')) LIKE ? OR LOWER(COALESCE(documento_codigo,'')) LIKE ?)")
		like := "%" + q + "%"
		args = append(args, like, like, like)
	}
	args = append(args, limit)
	query := `SELECT id,empresa_id,COALESCE(codigo,''),COALESCE(cliente_id,0),COALESCE(cliente_nombre,''),COALESCE(documento_tipo,''),COALESCE(documento_codigo,''),COALESCE(fecha_emision,''),COALESCE(fecha_vencimiento,''),COALESCE(dias_mora,0),COALESCE(valor_original,0),COALESCE(valor_pagado,0),COALESCE(saldo,0),COALESCE(estado_cartera,'pendiente'),COALESCE(moneda,'COP'),COALESCE(observaciones,'') FROM empresa_cuentas_por_cobrar WHERE ` + strings.Join(where, " AND ") + ` ORDER BY fecha_vencimiento ASC, saldo DESC LIMIT ?`
	rows, err := ExecQueryCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCobranzaCuenta{}
	for rows.Next() {
		var row EmpresaCobranzaCuenta
		if err := rows.Scan(&row.ID, &row.EmpresaID, &row.Codigo, &row.ClienteID, &row.ClienteNombre, &row.DocumentoTipo, &row.DocumentoCodigo, &row.FechaEmision, &row.FechaVencimiento, &row.DiasMora, &row.ValorOriginal, &row.ValorPagado, &row.Saldo, &row.EstadoCartera, &row.Moneda, &row.Observaciones); err != nil {
			return nil, err
		}
		row.DiasMora = cobranzaDiasMora(row.FechaVencimiento, row.DiasMora, time.Now())
		if filtro.MoraMin > 0 && row.DiasMora < filtro.MoraMin {
			continue
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func BuildEmpresaCobranzaDashboard(dbConn *sql.DB, empresaID int64) (EmpresaCobranzaDashboard, error) {
	if err := EnsureEmpresaCobranzaSchema(dbConn); err != nil {
		return EmpresaCobranzaDashboard{}, err
	}
	cuentas, err := listEmpresaCobranzaCuentas(dbConn, empresaID, EmpresaCobranzaCuentaFiltro{Limit: 500})
	if err != nil {
		return EmpresaCobranzaDashboard{}, err
	}
	dash := EmpresaCobranzaDashboard{EmpresaID: empresaID, CuentasPrioritarias: []EmpresaCobranzaCuenta{}, UltimasGestiones: []EmpresaCobranzaGestion{}, Campanas: []EmpresaCobranzaCampana{}}
	for _, c := range cuentas {
		dash.CuentasTotal++
		dash.SaldoTotal += c.Saldo
		if c.DiasMora > 0 {
			dash.CuentasVencidas++
			dash.SaldoVencido += c.Saldo
			if c.DiasMora >= 60 {
				dash.SaldoMoraCritica += c.Saldo
			}
		} else {
			dash.CuentasPorVencer++
			dash.SaldoPorVencer += c.Saldo
		}
		if len(dash.CuentasPrioritarias) < 12 && (c.DiasMora >= 15 || c.Saldo >= 1000000) {
			dash.CuentasPrioritarias = append(dash.CuentasPrioritarias, c)
		}
	}
	dash.PromesasPendientes, _ = countEmpresaCobranza(dbConn, `SELECT COUNT(1) FROM empresa_cobranza_promesas WHERE empresa_id=? AND estado_promesa='pendiente'`, empresaID)
	dash.PromesasIncumplidas, _ = countEmpresaCobranza(dbConn, `SELECT COUNT(1) FROM empresa_cobranza_promesas WHERE empresa_id=? AND estado_promesa='incumplida'`, empresaID)
	dash.GestionesHoy, _ = countEmpresaCobranza(dbConn, `SELECT COUNT(1) FROM empresa_cobranza_gestiones WHERE empresa_id=? AND substr(COALESCE(fecha_gestion,''),1,10)=?`, empresaID, time.Now().Format("2006-01-02"))
	dash.CampanasActivas, _ = countEmpresaCobranza(dbConn, `SELECT COUNT(1) FROM empresa_cobranza_campanas WHERE empresa_id=? AND estado_campana='activa' AND COALESCE(estado,'activo')='activo'`, empresaID)
	dash.RecuperadoMes, _ = sumEmpresaCobranza(dbConn, `SELECT COALESCE(SUM(valor_recuperado),0) FROM empresa_cobranza_campanas WHERE empresa_id=? AND substr(COALESCE(fecha_actualizacion,''),1,7)=?`, empresaID, time.Now().Format("2006-01"))
	dash.UltimasGestiones, _ = listEmpresaCobranzaGestiones(dbConn, empresaID, 10)
	dash.Campanas, _ = listEmpresaCobranzaCampanas(dbConn, empresaID, 10)
	if dash.SaldoMoraCritica > 0 {
		dash.Alertas = append(dash.Alertas, "Hay cartera con mora igual o superior a 60 dias; prioriza llamada y acuerdo formal.")
	}
	if dash.PromesasIncumplidas > 0 {
		dash.Alertas = append(dash.Alertas, "Existen promesas incumplidas que requieren escalamiento.")
	}
	if dash.CuentasTotal == 0 {
		dash.Alertas = append(dash.Alertas, "No hay cuentas por cobrar abiertas para gestionar.")
	}
	return dash, nil
}

func UpsertEmpresaCobranzaPlantilla(dbConn *sql.DB, row EmpresaCobranzaPlantilla) (int64, error) {
	if err := EnsureEmpresaCobranzaSchema(dbConn); err != nil {
		return 0, err
	}
	row.Nombre = strings.TrimSpace(row.Nombre)
	if row.EmpresaID <= 0 || row.Nombre == "" {
		return 0, errors.New("empresa_id y nombre son obligatorios")
	}
	row.Canal = normalizeCobranzaCanal(row.Canal)
	row.Estado = normalizeCobranzaEstado(row.Estado)
	if row.Codigo == "" {
		row.Codigo = nextEmpresaCobranzaCode(dbConn, row.EmpresaID, "empresa_cobranza_plantillas", "COB-PL")
	}
	activa := 0
	if row.Activa {
		activa = 1
	}
	if row.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_cobranza_plantillas SET nombre=?,canal=?,asunto=?,cuerpo=?,dias_mora_desde=?,dias_mora_hasta=?,prioridad=?,activa=?,fecha_actualizacion=CURRENT_TIMESTAMP,usuario_creador=?,estado=?,observaciones=? WHERE id=? AND empresa_id=?`, row.Nombre, row.Canal, row.Asunto, row.Cuerpo, row.DiasMoraDesde, row.DiasMoraHasta, row.Prioridad, activa, row.Usuario, row.Estado, row.Observaciones, row.ID, row.EmpresaID)
		return row.ID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_cobranza_plantillas (empresa_id,codigo,nombre,canal,asunto,cuerpo,dias_mora_desde,dias_mora_hasta,prioridad,activa,usuario_creador,estado,observaciones) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`, row.EmpresaID, row.Codigo, row.Nombre, row.Canal, row.Asunto, row.Cuerpo, row.DiasMoraDesde, row.DiasMoraHasta, row.Prioridad, activa, row.Usuario, row.Estado, row.Observaciones)
}

func UpsertEmpresaCobranzaCampana(dbConn *sql.DB, row EmpresaCobranzaCampana) (int64, error) {
	if err := EnsureEmpresaCobranzaSchema(dbConn); err != nil {
		return 0, err
	}
	row.Nombre = strings.TrimSpace(row.Nombre)
	if row.EmpresaID <= 0 || row.Nombre == "" {
		return 0, errors.New("empresa_id y nombre son obligatorios")
	}
	row.Tipo = normalizeCobranzaTipoCampana(row.Tipo)
	row.CanalPrincipal = normalizeCobranzaCanal(row.CanalPrincipal)
	row.EstadoCampana = normalizeCobranzaEstadoCampana(row.EstadoCampana)
	row.Estado = normalizeCobranzaEstado(row.Estado)
	if row.Codigo == "" {
		row.Codigo = nextEmpresaCobranzaCode(dbConn, row.EmpresaID, "empresa_cobranza_campanas", "COB-CA")
	}
	if row.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_cobranza_campanas SET nombre=?,tipo=?,canal_principal=?,segmento=?,fecha_inicio=?,fecha_fin=?,estado_campana=?,meta_recaudo=?,valor_asignado=?,valor_recuperado=?,contactos_programados=?,contactos_realizados=?,fecha_actualizacion=CURRENT_TIMESTAMP,usuario_creador=?,estado=?,observaciones=? WHERE id=? AND empresa_id=?`, row.Nombre, row.Tipo, row.CanalPrincipal, row.Segmento, row.FechaInicio, row.FechaFin, row.EstadoCampana, row.MetaRecaudo, row.ValorAsignado, row.ValorRecuperado, row.ContactosProgramados, row.ContactosRealizados, row.Usuario, row.Estado, row.Observaciones, row.ID, row.EmpresaID)
		return row.ID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_cobranza_campanas (empresa_id,codigo,nombre,tipo,canal_principal,segmento,fecha_inicio,fecha_fin,estado_campana,meta_recaudo,valor_asignado,valor_recuperado,contactos_programados,contactos_realizados,usuario_creador,estado,observaciones) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, row.EmpresaID, row.Codigo, row.Nombre, row.Tipo, row.CanalPrincipal, row.Segmento, row.FechaInicio, row.FechaFin, row.EstadoCampana, row.MetaRecaudo, row.ValorAsignado, row.ValorRecuperado, row.ContactosProgramados, row.ContactosRealizados, row.Usuario, row.Estado, row.Observaciones)
}

func CreateEmpresaCobranzaGestion(dbConn *sql.DB, row EmpresaCobranzaGestion) (EmpresaCobranzaGestion, *EmpresaCobranzaPromesa, error) {
	if err := EnsureEmpresaCobranzaSchema(dbConn); err != nil {
		return row, nil, err
	}
	if row.EmpresaID <= 0 {
		return row, nil, errors.New("empresa_id es obligatorio")
	}
	if row.CuentaID > 0 && strings.TrimSpace(row.ClienteNombre) == "" {
		cuenta, _ := GetEmpresaCobranzaCuenta(dbConn, row.EmpresaID, row.CuentaID)
		row.ClienteID = cuenta.ClienteID
		row.ClienteNombre = cuenta.ClienteNombre
		row.DocumentoCodigo = cuenta.DocumentoCodigo
	}
	row.Canal = normalizeCobranzaCanal(row.Canal)
	row.Resultado = normalizeCobranzaResultado(row.Resultado)
	if strings.TrimSpace(row.FechaGestion) == "" {
		row.FechaGestion = time.Now().Format("2006-01-02 15:04:05")
	}
	if row.PromesaEstado == "" {
		if row.ValorCompromiso > 0 && strings.TrimSpace(row.PromesaFecha) != "" {
			row.PromesaEstado = "pendiente"
		} else {
			row.PromesaEstado = "sin_promesa"
		}
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_cobranza_gestiones (empresa_id,cuenta_id,campana_id,plantilla_id,cliente_id,cliente_nombre,documento_codigo,canal,resultado,fecha_gestion,fecha_proximo_contacto,valor_compromiso,promesa_fecha,promesa_estado,mensaje,contacto,usuario_creador,estado,observaciones) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, row.EmpresaID, row.CuentaID, row.CampanaID, row.PlantillaID, row.ClienteID, row.ClienteNombre, row.DocumentoCodigo, row.Canal, row.Resultado, row.FechaGestion, row.FechaProximoContacto, row.ValorCompromiso, row.PromesaFecha, row.PromesaEstado, row.Mensaje, row.Contacto, row.Usuario, normalizeCobranzaEstado(row.Estado), row.Observaciones)
	if err != nil {
		return row, nil, err
	}
	row.ID = id
	var promesa *EmpresaCobranzaPromesa
	if row.ValorCompromiso > 0 && strings.TrimSpace(row.PromesaFecha) != "" {
		p := EmpresaCobranzaPromesa{EmpresaID: row.EmpresaID, CuentaID: row.CuentaID, GestionID: row.ID, ClienteNombre: row.ClienteNombre, DocumentoCodigo: row.DocumentoCodigo, ValorPrometido: row.ValorCompromiso, FechaPromesa: row.PromesaFecha, EstadoPromesa: "pendiente", Usuario: row.Usuario, Observaciones: "Creada desde gestion de cobranza"}
		pid, err := UpsertEmpresaCobranzaPromesa(dbConn, p)
		if err != nil {
			return row, nil, err
		}
		p.ID = pid
		promesa = &p
	}
	return row, promesa, nil
}

func UpsertEmpresaCobranzaPromesa(dbConn *sql.DB, row EmpresaCobranzaPromesa) (int64, error) {
	if err := EnsureEmpresaCobranzaSchema(dbConn); err != nil {
		return 0, err
	}
	if row.EmpresaID <= 0 || row.ValorPrometido <= 0 {
		return 0, errors.New("empresa_id y valor_prometido son obligatorios")
	}
	row.EstadoPromesa = normalizeCobranzaEstadoPromesa(row.EstadoPromesa)
	if row.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_cobranza_promesas SET cuenta_id=?,gestion_id=?,cliente_nombre=?,documento_codigo=?,valor_prometido=?,fecha_promesa=?,estado_promesa=?,fecha_cumplimiento=?,fecha_actualizacion=CURRENT_TIMESTAMP,usuario_creador=?,observaciones=? WHERE id=? AND empresa_id=?`, row.CuentaID, row.GestionID, row.ClienteNombre, row.DocumentoCodigo, row.ValorPrometido, row.FechaPromesa, row.EstadoPromesa, row.FechaCumplimiento, row.Usuario, row.Observaciones, row.ID, row.EmpresaID)
		return row.ID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_cobranza_promesas (empresa_id,cuenta_id,gestion_id,cliente_nombre,documento_codigo,valor_prometido,fecha_promesa,estado_promesa,fecha_cumplimiento,usuario_creador,observaciones) VALUES (?,?,?,?,?,?,?,?,?,?,?)`, row.EmpresaID, row.CuentaID, row.GestionID, row.ClienteNombre, row.DocumentoCodigo, row.ValorPrometido, row.FechaPromesa, row.EstadoPromesa, row.FechaCumplimiento, row.Usuario, row.Observaciones)
}

func UpdateEmpresaCobranzaPromesaEstado(dbConn *sql.DB, empresaID, promesaID int64, estado, usuario, observaciones string) (EmpresaCobranzaPromesa, error) {
	if err := EnsureEmpresaCobranzaSchema(dbConn); err != nil {
		return EmpresaCobranzaPromesa{}, err
	}
	estado = normalizeCobranzaEstadoPromesa(estado)
	fechaCumplimiento := ""
	if estado == "cumplida" {
		fechaCumplimiento = time.Now().Format("2006-01-02")
	}
	_, err := ExecCompat(dbConn, `UPDATE empresa_cobranza_promesas SET estado_promesa=?,fecha_cumplimiento=?,fecha_actualizacion=CURRENT_TIMESTAMP,usuario_creador=?,observaciones=? WHERE empresa_id=? AND id=?`, estado, fechaCumplimiento, usuario, observaciones, empresaID, promesaID)
	if err != nil {
		return EmpresaCobranzaPromesa{}, err
	}
	return GetEmpresaCobranzaPromesa(dbConn, empresaID, promesaID)
}

func GetEmpresaCobranzaCuenta(dbConn *sql.DB, empresaID, cuentaID int64) (EmpresaCobranzaCuenta, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,COALESCE(codigo,''),COALESCE(cliente_id,0),COALESCE(cliente_nombre,''),COALESCE(documento_tipo,''),COALESCE(documento_codigo,''),COALESCE(fecha_emision,''),COALESCE(fecha_vencimiento,''),COALESCE(dias_mora,0),COALESCE(valor_original,0),COALESCE(valor_pagado,0),COALESCE(saldo,0),COALESCE(estado_cartera,'pendiente'),COALESCE(moneda,'COP'),COALESCE(observaciones,'') FROM empresa_cuentas_por_cobrar WHERE empresa_id=? AND id=?`, empresaID, cuentaID)
	if err != nil {
		return EmpresaCobranzaCuenta{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return EmpresaCobranzaCuenta{}, sql.ErrNoRows
	}
	var row EmpresaCobranzaCuenta
	err = rows.Scan(&row.ID, &row.EmpresaID, &row.Codigo, &row.ClienteID, &row.ClienteNombre, &row.DocumentoTipo, &row.DocumentoCodigo, &row.FechaEmision, &row.FechaVencimiento, &row.DiasMora, &row.ValorOriginal, &row.ValorPagado, &row.Saldo, &row.EstadoCartera, &row.Moneda, &row.Observaciones)
	row.DiasMora = cobranzaDiasMora(row.FechaVencimiento, row.DiasMora, time.Now())
	return row, err
}

func ListEmpresaCobranzaPlantillas(dbConn *sql.DB, empresaID int64, limit int) ([]EmpresaCobranzaPlantilla, error) {
	if err := EnsureEmpresaCobranzaSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 300 {
		limit = 100
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,COALESCE(codigo,''),COALESCE(nombre,''),COALESCE(canal,'email'),COALESCE(asunto,''),COALESCE(cuerpo,''),COALESCE(dias_mora_desde,0),COALESCE(dias_mora_hasta,9999),COALESCE(prioridad,1),COALESCE(activa,1),COALESCE(usuario_creador,''),COALESCE(estado,'activo'),COALESCE(observaciones,''),COALESCE(fecha_creacion,'') FROM empresa_cobranza_plantillas WHERE empresa_id=? AND COALESCE(estado,'activo')='activo' ORDER BY prioridad DESC, dias_mora_desde LIMIT ?`, empresaID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCobranzaPlantilla{}
	for rows.Next() {
		var row EmpresaCobranzaPlantilla
		var activa int
		if err := rows.Scan(&row.ID, &row.EmpresaID, &row.Codigo, &row.Nombre, &row.Canal, &row.Asunto, &row.Cuerpo, &row.DiasMoraDesde, &row.DiasMoraHasta, &row.Prioridad, &activa, &row.Usuario, &row.Estado, &row.Observaciones, &row.FechaCreacion); err != nil {
			return nil, err
		}
		row.Activa = activa != 0
		out = append(out, row)
	}
	return out, rows.Err()
}

func ListEmpresaCobranzaCampanas(dbConn *sql.DB, empresaID int64, limit int) ([]EmpresaCobranzaCampana, error) {
	if err := EnsureEmpresaCobranzaSchema(dbConn); err != nil {
		return nil, err
	}
	return listEmpresaCobranzaCampanas(dbConn, empresaID, limit)
}

func listEmpresaCobranzaCampanas(dbConn *sql.DB, empresaID int64, limit int) ([]EmpresaCobranzaCampana, error) {
	if limit <= 0 || limit > 300 {
		limit = 100
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,COALESCE(codigo,''),COALESCE(nombre,''),COALESCE(tipo,'preventiva'),COALESCE(canal_principal,'whatsapp'),COALESCE(segmento,'todas'),COALESCE(fecha_inicio,''),COALESCE(fecha_fin,''),COALESCE(estado_campana,'borrador'),COALESCE(meta_recaudo,0),COALESCE(valor_asignado,0),COALESCE(valor_recuperado,0),COALESCE(contactos_programados,0),COALESCE(contactos_realizados,0),COALESCE(usuario_creador,''),COALESCE(estado,'activo'),COALESCE(observaciones,''),COALESCE(fecha_creacion,'') FROM empresa_cobranza_campanas WHERE empresa_id=? AND COALESCE(estado,'activo')='activo' ORDER BY fecha_creacion DESC LIMIT ?`, empresaID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCobranzaCampana{}
	for rows.Next() {
		var row EmpresaCobranzaCampana
		if err := rows.Scan(&row.ID, &row.EmpresaID, &row.Codigo, &row.Nombre, &row.Tipo, &row.CanalPrincipal, &row.Segmento, &row.FechaInicio, &row.FechaFin, &row.EstadoCampana, &row.MetaRecaudo, &row.ValorAsignado, &row.ValorRecuperado, &row.ContactosProgramados, &row.ContactosRealizados, &row.Usuario, &row.Estado, &row.Observaciones, &row.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func ListEmpresaCobranzaGestiones(dbConn *sql.DB, empresaID int64, limit int) ([]EmpresaCobranzaGestion, error) {
	if err := EnsureEmpresaCobranzaSchema(dbConn); err != nil {
		return nil, err
	}
	return listEmpresaCobranzaGestiones(dbConn, empresaID, limit)
}

func listEmpresaCobranzaGestiones(dbConn *sql.DB, empresaID int64, limit int) ([]EmpresaCobranzaGestion, error) {
	if limit <= 0 || limit > 300 {
		limit = 100
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,COALESCE(cuenta_id,0),COALESCE(campana_id,0),COALESCE(plantilla_id,0),COALESCE(cliente_id,0),COALESCE(cliente_nombre,''),COALESCE(documento_codigo,''),COALESCE(canal,'llamada'),COALESCE(resultado,'registrada'),COALESCE(fecha_gestion,''),COALESCE(fecha_proximo_contacto,''),COALESCE(valor_compromiso,0),COALESCE(promesa_fecha,''),COALESCE(promesa_estado,'sin_promesa'),COALESCE(mensaje,''),COALESCE(contacto,''),COALESCE(usuario_creador,''),COALESCE(estado,'activo'),COALESCE(observaciones,''),COALESCE(fecha_creacion,'') FROM empresa_cobranza_gestiones WHERE empresa_id=? AND COALESCE(estado,'activo')='activo' ORDER BY fecha_gestion DESC, id DESC LIMIT ?`, empresaID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCobranzaGestion{}
	for rows.Next() {
		var row EmpresaCobranzaGestion
		if err := rows.Scan(&row.ID, &row.EmpresaID, &row.CuentaID, &row.CampanaID, &row.PlantillaID, &row.ClienteID, &row.ClienteNombre, &row.DocumentoCodigo, &row.Canal, &row.Resultado, &row.FechaGestion, &row.FechaProximoContacto, &row.ValorCompromiso, &row.PromesaFecha, &row.PromesaEstado, &row.Mensaje, &row.Contacto, &row.Usuario, &row.Estado, &row.Observaciones, &row.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func ListEmpresaCobranzaPromesas(dbConn *sql.DB, empresaID int64, estado string, limit int) ([]EmpresaCobranzaPromesa, error) {
	if err := EnsureEmpresaCobranzaSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 300 {
		limit = 100
	}
	where := "empresa_id=?"
	args := []interface{}{empresaID}
	if e := normalizeCobranzaEstadoPromesa(estado); e != "" && e != "todas" {
		where += " AND estado_promesa=?"
		args = append(args, e)
	}
	args = append(args, limit)
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,COALESCE(cuenta_id,0),COALESCE(gestion_id,0),COALESCE(cliente_nombre,''),COALESCE(documento_codigo,''),COALESCE(valor_prometido,0),COALESCE(fecha_promesa,''),COALESCE(estado_promesa,'pendiente'),COALESCE(fecha_cumplimiento,''),COALESCE(usuario_creador,''),COALESCE(observaciones,''),COALESCE(fecha_creacion,'') FROM empresa_cobranza_promesas WHERE `+where+` ORDER BY fecha_promesa ASC, id DESC LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCobranzaPromesa{}
	for rows.Next() {
		var row EmpresaCobranzaPromesa
		if err := rows.Scan(&row.ID, &row.EmpresaID, &row.CuentaID, &row.GestionID, &row.ClienteNombre, &row.DocumentoCodigo, &row.ValorPrometido, &row.FechaPromesa, &row.EstadoPromesa, &row.FechaCumplimiento, &row.Usuario, &row.Observaciones, &row.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func GetEmpresaCobranzaPromesa(dbConn *sql.DB, empresaID, promesaID int64) (EmpresaCobranzaPromesa, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,COALESCE(cuenta_id,0),COALESCE(gestion_id,0),COALESCE(cliente_nombre,''),COALESCE(documento_codigo,''),COALESCE(valor_prometido,0),COALESCE(fecha_promesa,''),COALESCE(estado_promesa,'pendiente'),COALESCE(fecha_cumplimiento,''),COALESCE(usuario_creador,''),COALESCE(observaciones,''),COALESCE(fecha_creacion,'') FROM empresa_cobranza_promesas WHERE empresa_id=? AND id=?`, empresaID, promesaID)
	if err != nil {
		return EmpresaCobranzaPromesa{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return EmpresaCobranzaPromesa{}, sql.ErrNoRows
	}
	var row EmpresaCobranzaPromesa
	err = rows.Scan(&row.ID, &row.EmpresaID, &row.CuentaID, &row.GestionID, &row.ClienteNombre, &row.DocumentoCodigo, &row.ValorPrometido, &row.FechaPromesa, &row.EstadoPromesa, &row.FechaCumplimiento, &row.Usuario, &row.Observaciones, &row.FechaCreacion)
	return row, err
}

func SeedEmpresaCobranzaDemo(dbConn *sql.DB, empresaID int64, usuario string) error {
	if err := EnsureEmpresaCobranzaSchema(dbConn); err != nil {
		return err
	}
	if empresaID <= 0 {
		return errors.New("empresa_id es obligatorio")
	}
	if usuario == "" {
		usuario = "sistema"
	}
	cxcCount, _ := countEmpresaCobranza(dbConn, `SELECT COUNT(1) FROM empresa_cuentas_por_cobrar WHERE empresa_id=?`, empresaID)
	if cxcCount == 0 {
		now := time.Now()
		demos := []EmpresaCobranzaCuenta{
			{Codigo: "CXC-COB-001", ClienteNombre: "Cliente Hotel Corporativo SAS", DocumentoTipo: "factura", DocumentoCodigo: "FE-9001", FechaEmision: now.AddDate(0, -2, 0).Format("2006-01-02"), FechaVencimiento: now.AddDate(0, -1, -5).Format("2006-01-02"), ValorOriginal: 1850000, ValorPagado: 350000, Saldo: 1500000, EstadoCartera: "vencida", Moneda: "COP", Observaciones: "Cuenta demo para gestion de cobranza"},
			{Codigo: "CXC-COB-002", ClienteNombre: "Agencia Viajes Norte", DocumentoTipo: "factura", DocumentoCodigo: "FE-9002", FechaEmision: now.AddDate(0, -1, -10).Format("2006-01-02"), FechaVencimiento: now.AddDate(0, 0, 8).Format("2006-01-02"), ValorOriginal: 980000, ValorPagado: 0, Saldo: 980000, EstadoCartera: "pendiente", Moneda: "COP", Observaciones: "Cuenta demo por vencer"},
			{Codigo: "CXC-COB-003", ClienteNombre: "Constructora Horizonte", DocumentoTipo: "factura", DocumentoCodigo: "AIU-230", FechaEmision: now.AddDate(0, -4, 0).Format("2006-01-02"), FechaVencimiento: now.AddDate(0, -3, 0).Format("2006-01-02"), ValorOriginal: 5200000, ValorPagado: 1200000, Saldo: 4000000, EstadoCartera: "vencida", Moneda: "COP", Observaciones: "Cuenta demo mora critica"},
		}
		for _, d := range demos {
			dias := cobranzaDiasMora(d.FechaVencimiento, 0, now)
			_, _ = insertSQLCompat(dbConn, `INSERT INTO empresa_cuentas_por_cobrar (empresa_id,codigo,cliente_nombre,documento_tipo,documento_codigo,fecha_emision,fecha_vencimiento,dias_mora,valor_original,valor_pagado,saldo,estado_cartera,moneda,usuario_creador,estado,observaciones) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, empresaID, d.Codigo, d.ClienteNombre, d.DocumentoTipo, d.DocumentoCodigo, d.FechaEmision, d.FechaVencimiento, dias, d.ValorOriginal, d.ValorPagado, d.Saldo, d.EstadoCartera, d.Moneda, usuario, "activo", d.Observaciones)
		}
	}
	plantillas, _ := countEmpresaCobranza(dbConn, `SELECT COUNT(1) FROM empresa_cobranza_plantillas WHERE empresa_id=?`, empresaID)
	if plantillas == 0 {
		_, _ = UpsertEmpresaCobranzaPlantilla(dbConn, EmpresaCobranzaPlantilla{EmpresaID: empresaID, Nombre: "Recordatorio preventivo", Canal: "whatsapp", Asunto: "Recordatorio de pago", Cuerpo: "Hola {{cliente}}, tu documento {{documento}} tiene saldo {{saldo}}. Puedes realizar el pago por los canales autorizados.", DiasMoraDesde: 0, DiasMoraHasta: 5, Prioridad: 1, Activa: true, Usuario: usuario})
		_, _ = UpsertEmpresaCobranzaPlantilla(dbConn, EmpresaCobranzaPlantilla{EmpresaID: empresaID, Nombre: "Mora critica con compromiso", Canal: "email", Asunto: "Acuerdo de pago pendiente", Cuerpo: "Estimado {{cliente}}, requerimos acordar fecha de pago del documento {{documento}} por {{saldo}}.", DiasMoraDesde: 30, DiasMoraHasta: 9999, Prioridad: 3, Activa: true, Usuario: usuario})
	}
	campanas, _ := countEmpresaCobranza(dbConn, `SELECT COUNT(1) FROM empresa_cobranza_campanas WHERE empresa_id=?`, empresaID)
	if campanas == 0 {
		_, _ = UpsertEmpresaCobranzaCampana(dbConn, EmpresaCobranzaCampana{EmpresaID: empresaID, Nombre: "Recuperacion cartera vencida", Tipo: "recuperacion", CanalPrincipal: "whatsapp", Segmento: "mora mayor a 15 dias", FechaInicio: time.Now().Format("2006-01-02"), EstadoCampana: "activa", MetaRecaudo: 5000000, ValorAsignado: 5500000, ContactosProgramados: 25, Usuario: usuario})
	}
	return nil
}

func countEmpresaCobranza(dbConn *sql.DB, query string, args ...interface{}) (int, error) {
	var n int
	err := QueryRowCompat(dbConn, query, args...).Scan(&n)
	return n, err
}

func sumEmpresaCobranza(dbConn *sql.DB, query string, args ...interface{}) (float64, error) {
	var n float64
	err := QueryRowCompat(dbConn, query, args...).Scan(&n)
	return n, err
}

func nextEmpresaCobranzaCode(dbConn *sql.DB, empresaID int64, table, prefix string) string {
	var count int
	_ = QueryRowCompat(dbConn, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE empresa_id=?", table), empresaID).Scan(&count)
	return fmt.Sprintf("%s-%04d", prefix, count+1)
}

func cobranzaDiasMora(fechaVencimiento string, fallback int, today time.Time) int {
	fechaVencimiento = strings.TrimSpace(fechaVencimiento)
	if fechaVencimiento == "" {
		if fallback < 0 {
			return 0
		}
		return fallback
	}
	for _, layout := range []string{"2006-01-02", "2006-01-02 15:04:05", time.RFC3339} {
		if parsed, err := time.ParseInLocation(layout, fechaVencimiento, today.Location()); err == nil {
			diff := int(today.Truncate(24*time.Hour).Sub(parsed.Truncate(24*time.Hour)).Hours() / 24)
			if diff < 0 {
				return 0
			}
			return diff
		}
	}
	if fallback < 0 {
		return 0
	}
	return fallback
}

func normalizeCobranzaEstadoCartera(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "", "todas", "todos":
		return "todas"
	case "vencida", "vencido", "mora":
		return "vencida"
	case "pagada", "pagado":
		return "pagada"
	case "castigada", "castigado":
		return "castigada"
	default:
		return "pendiente"
	}
}

func normalizeCobranzaCanal(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "email", "correo":
		return "email"
	case "sms":
		return "sms"
	case "whatsapp", "wa":
		return "whatsapp"
	case "llamada", "telefono", "telefonico":
		return "llamada"
	default:
		return "llamada"
	}
}

func normalizeCobranzaResultado(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "enviado", "enviado_simulado", "sin_respuesta", "contactado", "promesa_pago", "reprogramado", "escalado", "fallido":
		return v
	case "promesa":
		return "promesa_pago"
	default:
		return "registrada"
	}
}

func normalizeCobranzaTipoCampana(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "preventiva", "recuperacion", "juridica", "masiva", "vip":
		return v
	default:
		return "preventiva"
	}
}

func normalizeCobranzaEstadoCampana(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "borrador", "activa", "pausada", "finalizada":
		return v
	default:
		return "borrador"
	}
}

func normalizeCobranzaEstadoPromesa(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "pendiente", "cumplida", "incumplida", "cancelada", "todas":
		return v
	default:
		return "pendiente"
	}
}

func normalizeCobranzaEstado(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "inactivo" || v == "archivado" {
		return v
	}
	return "activo"
}
