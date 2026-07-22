package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	EnergiaSolarProviderVictron   = "victron_vrm"
	EnergiaSolarProviderSMA       = "sma_sunny_portal"
	EnergiaSolarProviderSolarEdge = "solaredge_monitoring"
	EnergiaSolarProviderLocal     = "gateway_local"
)

type EmpresaEnergiaSolarSistema struct {
	ID                  int64   `json:"id"`
	EmpresaID           int64   `json:"empresa_id"`
	Proveedor           string  `json:"proveedor"`
	Modelo              string  `json:"modelo"`
	Nombre              string  `json:"nombre"`
	Ubicacion           string  `json:"ubicacion"`
	CapacidadKwp        float64 `json:"capacidad_kwp"`
	BateriaMarca        string  `json:"bateria_marca"`
	BateriaModelo       string  `json:"bateria_modelo"`
	BateriaSerial       string  `json:"bateria_serial"`
	BMSProtocolo        string  `json:"bms_protocolo"`
	CapacidadBateriaKwh float64 `json:"capacidad_bateria_kwh"`
	ApiBaseURL          string  `json:"api_base_url"`
	ApiKeyRef           string  `json:"api_key_ref"`
	InstalacionRef      string  `json:"instalacion_ref"`
	LocalGatewayURL     string  `json:"local_gateway_url"`
	IntervaloSegundos   int     `json:"intervalo_segundos"`
	EmailAlertas        string  `json:"email_alertas"`
	AlertasEmailActivas bool    `json:"alertas_email_activas"`
	Activo              bool    `json:"activo"`
	FechaCreacion       string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion  string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador      string  `json:"usuario_creador,omitempty"`
	Estado              string  `json:"estado,omitempty"`
	Observaciones       string  `json:"observaciones,omitempty"`
}

type EmpresaEnergiaSolarAlerta struct {
	ID               int64   `json:"id"`
	EmpresaID        int64   `json:"empresa_id"`
	SistemaID        int64   `json:"sistema_id"`
	Tipo             string  `json:"tipo"`
	Nombre           string  `json:"nombre"`
	Operador         string  `json:"operador"`
	Umbral           float64 `json:"umbral"`
	UmbralSecundario float64 `json:"umbral_secundario,omitempty"`
	Severidad        string  `json:"severidad"`
	EnviarEmail      bool    `json:"enviar_email"`
	Activo           bool    `json:"activo"`
	FechaCreacion    string  `json:"fecha_creacion,omitempty"`
	UsuarioCreador   string  `json:"usuario_creador,omitempty"`
	Estado           string  `json:"estado,omitempty"`
}

type EmpresaEnergiaSolarLectura struct {
	ID                int64                  `json:"id"`
	EmpresaID         int64                  `json:"empresa_id"`
	SistemaID         int64                  `json:"sistema_id"`
	FechaLectura      string                 `json:"fecha_lectura,omitempty"`
	PotenciaSolarW    float64                `json:"potencia_solar_w"`
	ProduccionDiaKwh  float64                `json:"produccion_dia_kwh"`
	ConsumoW          float64                `json:"consumo_w"`
	BateriaSOC        float64                `json:"bateria_soc_pct"`
	BateriaSOH        float64                `json:"bateria_soh_pct"`
	BateriaVoltaje    float64                `json:"bateria_voltaje_v"`
	BateriaCorrienteA float64                `json:"bateria_corriente_a"`
	BateriaCargaW     float64                `json:"bateria_carga_w"`
	BateriaDescargaW  float64                `json:"bateria_descarga_w"`
	BateriaCiclos     float64                `json:"bateria_ciclos"`
	CeldaVoltajeMin   float64                `json:"celda_voltaje_min_v"`
	CeldaVoltajeMax   float64                `json:"celda_voltaje_max_v"`
	InversorPotenciaW float64                `json:"inversor_potencia_w"`
	RedPotenciaW      float64                `json:"red_potencia_w"`
	TemperaturaC      float64                `json:"temperatura_c"`
	EstadoPaneles     string                 `json:"estado_paneles"`
	EstadoBateria     string                 `json:"estado_bateria"`
	EstadoInversor    string                 `json:"estado_inversor"`
	Raw               map[string]interface{} `json:"raw,omitempty"`
	RawJSON           string                 `json:"-"`
	FechaCreacion     string                 `json:"fecha_creacion,omitempty"`
	UsuarioCreador    string                 `json:"usuario_creador,omitempty"`
}

type EmpresaEnergiaSolarEvento struct {
	ID             int64  `json:"id"`
	EmpresaID      int64  `json:"empresa_id"`
	SistemaID      int64  `json:"sistema_id"`
	AlertaID       int64  `json:"alerta_id"`
	LecturaID      int64  `json:"lectura_id"`
	Tipo           string `json:"tipo"`
	Severidad      string `json:"severidad"`
	Mensaje        string `json:"mensaje"`
	EmailEnviado   bool   `json:"email_enviado"`
	EmailError     string `json:"email_error,omitempty"`
	FechaCreacion  string `json:"fecha_creacion,omitempty"`
	UsuarioCreador string `json:"usuario_creador,omitempty"`
	Estado         string `json:"estado,omitempty"`
}

func EnsureEmpresaEnergiaSolarSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("conexion de base de datos no disponible")
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_energia_solar_sistemas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			proveedor TEXT NOT NULL,
			modelo TEXT,
			nombre TEXT,
			ubicacion TEXT,
			capacidad_kwp REAL DEFAULT 0,
			bateria_marca TEXT,
			bateria_modelo TEXT,
			bateria_serial TEXT,
			bms_protocolo TEXT,
			capacidad_bateria_kwh REAL DEFAULT 0,
			api_base_url TEXT,
			api_key_ref TEXT,
			instalacion_ref TEXT,
			local_gateway_url TEXT,
			intervalo_segundos INTEGER DEFAULT 300,
			email_alertas TEXT,
			alertas_email_activas INTEGER DEFAULT 1,
			activo INTEGER DEFAULT 1,
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_energia_solar_alertas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			sistema_id INTEGER NOT NULL,
			tipo TEXT NOT NULL,
			nombre TEXT,
			operador TEXT DEFAULT '<',
			umbral REAL DEFAULT 0,
			umbral_secundario REAL DEFAULT 0,
			severidad TEXT DEFAULT 'media',
			enviar_email INTEGER DEFAULT 1,
			activo INTEGER DEFAULT 1,
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo'
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_energia_solar_lecturas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			sistema_id INTEGER NOT NULL,
			fecha_lectura TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			potencia_solar_w REAL DEFAULT 0,
			produccion_dia_kwh REAL DEFAULT 0,
			consumo_w REAL DEFAULT 0,
			bateria_soc_pct REAL DEFAULT 0,
			bateria_soh_pct REAL DEFAULT 0,
			bateria_voltaje_v REAL DEFAULT 0,
			bateria_corriente_a REAL DEFAULT 0,
			bateria_carga_w REAL DEFAULT 0,
			bateria_descarga_w REAL DEFAULT 0,
			bateria_ciclos REAL DEFAULT 0,
			celda_voltaje_min_v REAL DEFAULT 0,
			celda_voltaje_max_v REAL DEFAULT 0,
			inversor_potencia_w REAL DEFAULT 0,
			red_potencia_w REAL DEFAULT 0,
			temperatura_c REAL DEFAULT 0,
			estado_paneles TEXT,
			estado_bateria TEXT,
			estado_inversor TEXT,
			raw_json TEXT,
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_energia_solar_eventos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			sistema_id INTEGER NOT NULL,
			alerta_id INTEGER DEFAULT 0,
			lectura_id INTEGER DEFAULT 0,
			tipo TEXT NOT NULL,
			severidad TEXT DEFAULT 'media',
			mensaje TEXT,
			email_enviado INTEGER DEFAULT 0,
			email_error TEXT,
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo'
		)`,
		`CREATE INDEX IF NOT EXISTS ix_energia_solar_sistemas_empresa ON empresa_energia_solar_sistemas(empresa_id, estado, activo)`,
		`CREATE INDEX IF NOT EXISTS ix_energia_solar_alertas_sistema ON empresa_energia_solar_alertas(empresa_id, sistema_id, estado, activo)`,
		`CREATE INDEX IF NOT EXISTS ix_energia_solar_lecturas_sistema ON empresa_energia_solar_lecturas(empresa_id, sistema_id, fecha_lectura)`,
		`CREATE INDEX IF NOT EXISTS ix_energia_solar_eventos_empresa ON empresa_energia_solar_eventos(empresa_id, sistema_id, fecha_creacion)`,
	}
	for _, stmt := range stmts {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	columns := []struct {
		table string
		name  string
		def   string
	}{
		{"empresa_energia_solar_sistemas", "bateria_marca", "TEXT"},
		{"empresa_energia_solar_sistemas", "bateria_modelo", "TEXT"},
		{"empresa_energia_solar_sistemas", "bateria_serial", "TEXT"},
		{"empresa_energia_solar_sistemas", "bms_protocolo", "TEXT"},
		{"empresa_energia_solar_lecturas", "bateria_soh_pct", "REAL DEFAULT 0"},
		{"empresa_energia_solar_lecturas", "bateria_corriente_a", "REAL DEFAULT 0"},
		{"empresa_energia_solar_lecturas", "bateria_descarga_w", "REAL DEFAULT 0"},
		{"empresa_energia_solar_lecturas", "bateria_ciclos", "REAL DEFAULT 0"},
		{"empresa_energia_solar_lecturas", "celda_voltaje_min_v", "REAL DEFAULT 0"},
		{"empresa_energia_solar_lecturas", "celda_voltaje_max_v", "REAL DEFAULT 0"},
	}
	for _, col := range columns {
		if err := ensureColumnIfMissing(dbConn, col.table, col.name, col.def); err != nil {
			return err
		}
	}
	return nil
}

// EmpresaEnergiaSolarSchemaReady verifies migration-owned solar tables without
// executing DDL during an enterprise request.
func EmpresaEnergiaSolarSchemaReady(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("conexion de base de datos no disponible")
	}
	for _, table := range []string{
		"empresa_energia_solar_sistemas",
		"empresa_energia_solar_alertas",
		"empresa_energia_solar_lecturas",
		"empresa_energia_solar_eventos",
	} {
		var marker int
		err := queryRowSQLCompat(dbConn, "SELECT 1 FROM "+table+" WHERE 1=0").Scan(&marker)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return fmt.Errorf("esquema de energia solar no disponible (%s): %w", table, err)
		}
	}
	return nil
}

func NormalizeEnergiaSolarProvider(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "victron", "victron_vrm", "vrm":
		return EnergiaSolarProviderVictron
	case "sma", "sma_sunny_portal", "sunny_portal", "ennexos":
		return EnergiaSolarProviderSMA
	case "solaredge", "solaredge_monitoring":
		return EnergiaSolarProviderSolarEdge
	case "gateway", "gateway_local", "local":
		return EnergiaSolarProviderLocal
	default:
		return ""
	}
}

func UpsertEmpresaEnergiaSolarSistema(dbConn *sql.DB, item EmpresaEnergiaSolarSistema) (int64, error) {
	if item.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id invalido")
	}
	item.Proveedor = NormalizeEnergiaSolarProvider(item.Proveedor)
	if item.Proveedor == "" {
		return 0, fmt.Errorf("proveedor solar no soportado")
	}
	if strings.TrimSpace(item.Nombre) == "" {
		item.Nombre = "Sistema solar"
	}
	if item.IntervaloSegundos <= 0 {
		item.IntervaloSegundos = 300
	}
	if item.IntervaloSegundos < 30 {
		item.IntervaloSegundos = 30
	}
	if item.IntervaloSegundos > 86400 {
		item.IntervaloSegundos = 86400
	}
	if item.ApiKeyRef != "" && !strings.HasPrefix(strings.ToLower(strings.TrimSpace(item.ApiKeyRef)), "env:") {
		return 0, fmt.Errorf("api_key_ref debe usar referencia segura env:NOMBRE_VARIABLE")
	}
	activeInt := 0
	if item.Activo {
		activeInt = 1
	}
	emailInt := 0
	if item.AlertasEmailActivas {
		emailInt = 1
	}
	if item.ID > 0 {
		_, err := execSQLCompat(dbConn, `UPDATE empresa_energia_solar_sistemas
			SET proveedor=?, modelo=?, nombre=?, ubicacion=?, capacidad_kwp=?,
				bateria_marca=?, bateria_modelo=?, bateria_serial=?, bms_protocolo=?, capacidad_bateria_kwh=?,
				api_base_url=?, api_key_ref=?, instalacion_ref=?, local_gateway_url=?, intervalo_segundos=?,
				email_alertas=?, alertas_email_activas=?, activo=?, fecha_actualizacion=CAST(CURRENT_TIMESTAMP AS TEXT),
				usuario_creador=?, estado=?, observaciones=?
			WHERE empresa_id=? AND id=?`,
			item.Proveedor, strings.TrimSpace(item.Modelo), strings.TrimSpace(item.Nombre), strings.TrimSpace(item.Ubicacion),
			item.CapacidadKwp, strings.TrimSpace(item.BateriaMarca), strings.TrimSpace(item.BateriaModelo),
			strings.TrimSpace(item.BateriaSerial), strings.TrimSpace(item.BMSProtocolo), item.CapacidadBateriaKwh,
			strings.TrimSpace(item.ApiBaseURL), strings.TrimSpace(item.ApiKeyRef),
			strings.TrimSpace(item.InstalacionRef), strings.TrimSpace(item.LocalGatewayURL), item.IntervaloSegundos,
			strings.TrimSpace(item.EmailAlertas), emailInt, activeInt, strings.TrimSpace(item.UsuarioCreador),
			firstNonBlankDB(item.Estado, "activo"), strings.TrimSpace(item.Observaciones), item.EmpresaID, item.ID)
		return item.ID, err
	}
	var id int64
	err := queryRowSQLCompat(dbConn, `INSERT INTO empresa_energia_solar_sistemas (
		empresa_id, proveedor, modelo, nombre, ubicacion, capacidad_kwp,
		bateria_marca, bateria_modelo, bateria_serial, bms_protocolo, capacidad_bateria_kwh,
		api_base_url, api_key_ref, instalacion_ref, local_gateway_url, intervalo_segundos,
		email_alertas, alertas_email_activas, activo, usuario_creador, estado, observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`,
		item.EmpresaID, item.Proveedor, strings.TrimSpace(item.Modelo), strings.TrimSpace(item.Nombre), strings.TrimSpace(item.Ubicacion),
		item.CapacidadKwp, strings.TrimSpace(item.BateriaMarca), strings.TrimSpace(item.BateriaModelo),
		strings.TrimSpace(item.BateriaSerial), strings.TrimSpace(item.BMSProtocolo), item.CapacidadBateriaKwh,
		strings.TrimSpace(item.ApiBaseURL), strings.TrimSpace(item.ApiKeyRef),
		strings.TrimSpace(item.InstalacionRef), strings.TrimSpace(item.LocalGatewayURL), item.IntervaloSegundos,
		strings.TrimSpace(item.EmailAlertas), emailInt, activeInt, strings.TrimSpace(item.UsuarioCreador),
		firstNonBlankDB(item.Estado, "activo"), strings.TrimSpace(item.Observaciones)).Scan(&id)
	return id, err
}

func ListEmpresaEnergiaSolarSistemas(dbConn *sql.DB, empresaID int64, includeInactive bool) ([]EmpresaEnergiaSolarSistema, error) {
	where := "empresa_id=?"
	if !includeInactive {
		where += " AND COALESCE(estado,'activo') <> 'inactivo' AND COALESCE(activo,1)=1"
	}
	rows, err := querySQLCompat(dbConn, `SELECT id, empresa_id, proveedor, COALESCE(modelo,''), COALESCE(nombre,''), COALESCE(ubicacion,''),
		COALESCE(capacidad_kwp,0), COALESCE(bateria_marca,''), COALESCE(bateria_modelo,''), COALESCE(bateria_serial,''),
		COALESCE(bms_protocolo,''), COALESCE(capacidad_bateria_kwh,0), COALESCE(api_base_url,''), COALESCE(api_key_ref,''),
		COALESCE(instalacion_ref,''), COALESCE(local_gateway_url,''), COALESCE(intervalo_segundos,300),
		COALESCE(email_alertas,''), COALESCE(alertas_email_activas,1), COALESCE(activo,1),
		COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,''), COALESCE(estado,'activo'), COALESCE(observaciones,'')
		FROM empresa_energia_solar_sistemas WHERE `+where+` ORDER BY id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []EmpresaEnergiaSolarSistema{}
	for rows.Next() {
		var it EmpresaEnergiaSolarSistema
		var emailInt, activeInt int
		if err := rows.Scan(&it.ID, &it.EmpresaID, &it.Proveedor, &it.Modelo, &it.Nombre, &it.Ubicacion, &it.CapacidadKwp, &it.BateriaMarca, &it.BateriaModelo, &it.BateriaSerial, &it.BMSProtocolo, &it.CapacidadBateriaKwh, &it.ApiBaseURL, &it.ApiKeyRef, &it.InstalacionRef, &it.LocalGatewayURL, &it.IntervaloSegundos, &it.EmailAlertas, &emailInt, &activeInt, &it.FechaCreacion, &it.FechaActualizacion, &it.UsuarioCreador, &it.Estado, &it.Observaciones); err != nil {
			return nil, err
		}
		it.AlertasEmailActivas = emailInt != 0
		it.Activo = activeInt != 0
		items = append(items, it)
	}
	return items, rows.Err()
}

func GetEmpresaEnergiaSolarSistema(dbConn *sql.DB, empresaID, sistemaID int64) (*EmpresaEnergiaSolarSistema, error) {
	items, err := ListEmpresaEnergiaSolarSistemas(dbConn, empresaID, true)
	if err != nil {
		return nil, err
	}
	for _, it := range items {
		if it.ID == sistemaID {
			return &it, nil
		}
	}
	return nil, sql.ErrNoRows
}

func UpsertEmpresaEnergiaSolarAlerta(dbConn *sql.DB, it EmpresaEnergiaSolarAlerta) (int64, error) {
	if it.EmpresaID <= 0 || it.SistemaID <= 0 {
		return 0, fmt.Errorf("empresa_id y sistema_id son obligatorios")
	}
	if strings.TrimSpace(it.Tipo) == "" {
		return 0, fmt.Errorf("tipo de alerta requerido")
	}
	if strings.TrimSpace(it.Nombre) == "" {
		it.Nombre = it.Tipo
	}
	if strings.TrimSpace(it.Operador) == "" {
		it.Operador = "<"
	}
	if strings.TrimSpace(it.Severidad) == "" {
		it.Severidad = "media"
	}
	activeInt := 0
	if it.Activo {
		activeInt = 1
	}
	emailInt := 0
	if it.EnviarEmail {
		emailInt = 1
	}
	if it.ID > 0 {
		_, err := execSQLCompat(dbConn, `UPDATE empresa_energia_solar_alertas SET tipo=?, nombre=?, operador=?, umbral=?, umbral_secundario=?, severidad=?, enviar_email=?, activo=?, fecha_actualizacion=CAST(CURRENT_TIMESTAMP AS TEXT), usuario_creador=?, estado=? WHERE empresa_id=? AND id=?`,
			strings.TrimSpace(it.Tipo), strings.TrimSpace(it.Nombre), strings.TrimSpace(it.Operador), it.Umbral, it.UmbralSecundario, strings.TrimSpace(it.Severidad), emailInt, activeInt, strings.TrimSpace(it.UsuarioCreador), firstNonBlankDB(it.Estado, "activo"), it.EmpresaID, it.ID)
		return it.ID, err
	}
	var id int64
	err := queryRowSQLCompat(dbConn, `INSERT INTO empresa_energia_solar_alertas (empresa_id, sistema_id, tipo, nombre, operador, umbral, umbral_secundario, severidad, enviar_email, activo, usuario_creador, estado) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`,
		it.EmpresaID, it.SistemaID, strings.TrimSpace(it.Tipo), strings.TrimSpace(it.Nombre), strings.TrimSpace(it.Operador), it.Umbral, it.UmbralSecundario, strings.TrimSpace(it.Severidad), emailInt, activeInt, strings.TrimSpace(it.UsuarioCreador), firstNonBlankDB(it.Estado, "activo")).Scan(&id)
	return id, err
}

func ListEmpresaEnergiaSolarAlertas(dbConn *sql.DB, empresaID, sistemaID int64, includeInactive bool) ([]EmpresaEnergiaSolarAlerta, error) {
	where := "empresa_id=?"
	args := []interface{}{empresaID}
	if sistemaID > 0 {
		where += " AND sistema_id=?"
		args = append(args, sistemaID)
	}
	if !includeInactive {
		where += " AND COALESCE(estado,'activo') <> 'inactivo' AND COALESCE(activo,1)=1"
	}
	rows, err := querySQLCompat(dbConn, `SELECT id, empresa_id, sistema_id, tipo, COALESCE(nombre,''), COALESCE(operador,'<'), COALESCE(umbral,0), COALESCE(umbral_secundario,0), COALESCE(severidad,'media'), COALESCE(enviar_email,1), COALESCE(activo,1), COALESCE(fecha_creacion,''), COALESCE(usuario_creador,''), COALESCE(estado,'activo') FROM empresa_energia_solar_alertas WHERE `+where+` ORDER BY id DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []EmpresaEnergiaSolarAlerta{}
	for rows.Next() {
		var it EmpresaEnergiaSolarAlerta
		var emailInt, activeInt int
		if err := rows.Scan(&it.ID, &it.EmpresaID, &it.SistemaID, &it.Tipo, &it.Nombre, &it.Operador, &it.Umbral, &it.UmbralSecundario, &it.Severidad, &emailInt, &activeInt, &it.FechaCreacion, &it.UsuarioCreador, &it.Estado); err != nil {
			return nil, err
		}
		it.EnviarEmail = emailInt != 0
		it.Activo = activeInt != 0
		items = append(items, it)
	}
	return items, rows.Err()
}

func InsertEmpresaEnergiaSolarLectura(dbConn *sql.DB, it EmpresaEnergiaSolarLectura) (int64, error) {
	if it.EmpresaID <= 0 || it.SistemaID <= 0 {
		return 0, fmt.Errorf("empresa_id y sistema_id son obligatorios")
	}
	raw := strings.TrimSpace(it.RawJSON)
	if raw == "" && it.Raw != nil {
		if b, err := json.Marshal(it.Raw); err == nil {
			raw = string(b)
		}
	}
	var id int64
	err := queryRowSQLCompat(dbConn, `INSERT INTO empresa_energia_solar_lecturas (
		empresa_id, sistema_id, potencia_solar_w, produccion_dia_kwh, consumo_w, bateria_soc_pct,
		bateria_soh_pct, bateria_voltaje_v, bateria_corriente_a, bateria_carga_w, bateria_descarga_w,
		bateria_ciclos, celda_voltaje_min_v, celda_voltaje_max_v, inversor_potencia_w, red_potencia_w, temperatura_c,
		estado_paneles, estado_bateria, estado_inversor, raw_json, usuario_creador
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`,
		it.EmpresaID, it.SistemaID, it.PotenciaSolarW, it.ProduccionDiaKwh, it.ConsumoW, it.BateriaSOC,
		it.BateriaSOH, it.BateriaVoltaje, it.BateriaCorrienteA, it.BateriaCargaW, it.BateriaDescargaW,
		it.BateriaCiclos, it.CeldaVoltajeMin, it.CeldaVoltajeMax, it.InversorPotenciaW, it.RedPotenciaW, it.TemperaturaC,
		strings.TrimSpace(it.EstadoPaneles), strings.TrimSpace(it.EstadoBateria), strings.TrimSpace(it.EstadoInversor), raw, strings.TrimSpace(it.UsuarioCreador)).Scan(&id)
	return id, err
}

func ListEmpresaEnergiaSolarLecturas(dbConn *sql.DB, empresaID, sistemaID int64, limit int) ([]EmpresaEnergiaSolarLectura, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	where := "empresa_id=?"
	args := []interface{}{empresaID}
	if sistemaID > 0 {
		where += " AND sistema_id=?"
		args = append(args, sistemaID)
	}
	args = append(args, limit)
	rows, err := querySQLCompat(dbConn, `SELECT id, empresa_id, sistema_id, COALESCE(fecha_lectura,''), COALESCE(potencia_solar_w,0), COALESCE(produccion_dia_kwh,0), COALESCE(consumo_w,0), COALESCE(bateria_soc_pct,0), COALESCE(bateria_soh_pct,0), COALESCE(bateria_voltaje_v,0), COALESCE(bateria_corriente_a,0), COALESCE(bateria_carga_w,0), COALESCE(bateria_descarga_w,0), COALESCE(bateria_ciclos,0), COALESCE(celda_voltaje_min_v,0), COALESCE(celda_voltaje_max_v,0), COALESCE(inversor_potencia_w,0), COALESCE(red_potencia_w,0), COALESCE(temperatura_c,0), COALESCE(estado_paneles,''), COALESCE(estado_bateria,''), COALESCE(estado_inversor,''), COALESCE(raw_json,''), COALESCE(fecha_creacion,''), COALESCE(usuario_creador,'') FROM empresa_energia_solar_lecturas WHERE `+where+` ORDER BY id DESC LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []EmpresaEnergiaSolarLectura{}
	for rows.Next() {
		var it EmpresaEnergiaSolarLectura
		if err := rows.Scan(&it.ID, &it.EmpresaID, &it.SistemaID, &it.FechaLectura, &it.PotenciaSolarW, &it.ProduccionDiaKwh, &it.ConsumoW, &it.BateriaSOC, &it.BateriaSOH, &it.BateriaVoltaje, &it.BateriaCorrienteA, &it.BateriaCargaW, &it.BateriaDescargaW, &it.BateriaCiclos, &it.CeldaVoltajeMin, &it.CeldaVoltajeMax, &it.InversorPotenciaW, &it.RedPotenciaW, &it.TemperaturaC, &it.EstadoPaneles, &it.EstadoBateria, &it.EstadoInversor, &it.RawJSON, &it.FechaCreacion, &it.UsuarioCreador); err != nil {
			return nil, err
		}
		if it.RawJSON != "" {
			_ = json.Unmarshal([]byte(it.RawJSON), &it.Raw)
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func InsertEmpresaEnergiaSolarEvento(dbConn *sql.DB, it EmpresaEnergiaSolarEvento) (int64, error) {
	var id int64
	emailInt := 0
	if it.EmailEnviado {
		emailInt = 1
	}
	err := queryRowSQLCompat(dbConn, `INSERT INTO empresa_energia_solar_eventos (empresa_id, sistema_id, alerta_id, lectura_id, tipo, severidad, mensaje, email_enviado, email_error, usuario_creador, estado) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`,
		it.EmpresaID, it.SistemaID, it.AlertaID, it.LecturaID, strings.TrimSpace(it.Tipo), firstNonBlankDB(it.Severidad, "media"), strings.TrimSpace(it.Mensaje), emailInt, strings.TrimSpace(it.EmailError), strings.TrimSpace(it.UsuarioCreador), firstNonBlankDB(it.Estado, "activo")).Scan(&id)
	return id, err
}

func ListEmpresaEnergiaSolarEventos(dbConn *sql.DB, empresaID, sistemaID int64, limit int) ([]EmpresaEnergiaSolarEvento, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	where := "empresa_id=?"
	args := []interface{}{empresaID}
	if sistemaID > 0 {
		where += " AND sistema_id=?"
		args = append(args, sistemaID)
	}
	args = append(args, limit)
	rows, err := querySQLCompat(dbConn, `SELECT id, empresa_id, sistema_id, COALESCE(alerta_id,0), COALESCE(lectura_id,0), tipo, COALESCE(severidad,'media'), COALESCE(mensaje,''), COALESCE(email_enviado,0), COALESCE(email_error,''), COALESCE(fecha_creacion,''), COALESCE(usuario_creador,''), COALESCE(estado,'activo') FROM empresa_energia_solar_eventos WHERE `+where+` ORDER BY id DESC LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []EmpresaEnergiaSolarEvento{}
	for rows.Next() {
		var it EmpresaEnergiaSolarEvento
		var emailInt int
		if err := rows.Scan(&it.ID, &it.EmpresaID, &it.SistemaID, &it.AlertaID, &it.LecturaID, &it.Tipo, &it.Severidad, &it.Mensaje, &emailInt, &it.EmailError, &it.FechaCreacion, &it.UsuarioCreador, &it.Estado); err != nil {
			return nil, err
		}
		it.EmailEnviado = emailInt != 0
		items = append(items, it)
	}
	return items, rows.Err()
}

func firstNonBlankDB(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
