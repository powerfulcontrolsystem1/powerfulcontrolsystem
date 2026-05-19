package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

// EmpresaConfiguracionOperativa define controles de cobro por empresa.
type EmpresaConfiguracionOperativa struct {
	ID                              int64                                   `json:"id"`
	EmpresaID                       int64                                   `json:"empresa_id"`
	MetodoPagoEfectivo              bool                                    `json:"metodo_pago_efectivo"`
	MetodoPagoTarjetaCredito        bool                                    `json:"metodo_pago_tarjeta_credito"`
	MetodoPagoTarjetaDebito         bool                                    `json:"metodo_pago_tarjeta_debito"`
	MetodoPagoTransferenciaBancaria bool                                    `json:"metodo_pago_transferencia_bancaria"`
	MetodoPagoMixto                 bool                                    `json:"metodo_pago_mixto"`
	MetodoPagoCodigoDescuento       bool                                    `json:"metodo_pago_codigo_descuento"`
	HabilitarPropinas               bool                                    `json:"habilitar_propinas"`
	HabilitarComisiones             bool                                    `json:"habilitar_comisiones"`
	Roles                           []EmpresaConfiguracionOperativaRol      `json:"roles,omitempty"`
	Politicas                       []EmpresaConfiguracionOperativaPolitica `json:"politicas,omitempty"`
	FechaCreacion                   string                                  `json:"fecha_creacion,omitempty"`
	FechaActualizacion              string                                  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador                  string                                  `json:"usuario_creador,omitempty"`
	Estado                          string                                  `json:"estado,omitempty"`
	Observaciones                   string                                  `json:"observaciones,omitempty"`
}

// EmpresaConfiguracionOperativaPolitica define reglas contextuales por canal/sucursal/turno.
type EmpresaConfiguracionOperativaPolitica struct {
	ID                              int64  `json:"id"`
	EmpresaID                       int64  `json:"empresa_id"`
	CanalVenta                      string `json:"canal_venta,omitempty"`
	SucursalID                      int64  `json:"sucursal_id,omitempty"`
	Turno                           string `json:"turno,omitempty"`
	Prioridad                       int    `json:"prioridad"`
	MetodoPagoEfectivo              bool   `json:"metodo_pago_efectivo"`
	MetodoPagoTarjetaCredito        bool   `json:"metodo_pago_tarjeta_credito"`
	MetodoPagoTarjetaDebito         bool   `json:"metodo_pago_tarjeta_debito"`
	MetodoPagoTransferenciaBancaria bool   `json:"metodo_pago_transferencia_bancaria"`
	MetodoPagoMixto                 bool   `json:"metodo_pago_mixto"`
	MetodoPagoCodigoDescuento       bool   `json:"metodo_pago_codigo_descuento"`
	HabilitarPropinas               bool   `json:"habilitar_propinas"`
	HabilitarComisiones             bool   `json:"habilitar_comisiones"`
	FechaCreacion                   string `json:"fecha_creacion,omitempty"`
	FechaActualizacion              string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador                  string `json:"usuario_creador,omitempty"`
	Estado                          string `json:"estado,omitempty"`
	Observaciones                   string `json:"observaciones,omitempty"`
}

// EmpresaConfiguracionOperativaContexto representa el contexto operativo de cobro.
type EmpresaConfiguracionOperativaContexto struct {
	Rol        string `json:"rol,omitempty"`
	CanalVenta string `json:"canal_venta,omitempty"`
	SucursalID int64  `json:"sucursal_id,omitempty"`
	Turno      string `json:"turno,omitempty"`
}

// EmpresaConfiguracionOperativaHistorialSnapshot guarda snapshots para rollback operativo.
type EmpresaConfiguracionOperativaHistorialSnapshot struct {
	ID                  int64  `json:"id"`
	EmpresaID           int64  `json:"empresa_id"`
	Evento              string `json:"evento"`
	RollbackDeHistorial int64  `json:"rollback_de_historial,omitempty"`
	SnapshotJSON        string `json:"snapshot_json"`
	SimulacionJSON      string `json:"simulacion_json,omitempty"`
	FechaCreacion       string `json:"fecha_creacion,omitempty"`
	FechaActualizacion  string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador      string `json:"usuario_creador,omitempty"`
	Estado              string `json:"estado,omitempty"`
	Observaciones       string `json:"observaciones,omitempty"`
}

// EmpresaConfiguracionOperativaRol define controles de cobro para un rol en una empresa.
type EmpresaConfiguracionOperativaRol struct {
	ID                              int64  `json:"id"`
	EmpresaID                       int64  `json:"empresa_id"`
	Rol                             string `json:"rol"`
	MetodoPagoEfectivo              bool   `json:"metodo_pago_efectivo"`
	MetodoPagoTarjetaCredito        bool   `json:"metodo_pago_tarjeta_credito"`
	MetodoPagoTarjetaDebito         bool   `json:"metodo_pago_tarjeta_debito"`
	MetodoPagoTransferenciaBancaria bool   `json:"metodo_pago_transferencia_bancaria"`
	MetodoPagoMixto                 bool   `json:"metodo_pago_mixto"`
	MetodoPagoCodigoDescuento       bool   `json:"metodo_pago_codigo_descuento"`
	HabilitarPropinas               bool   `json:"habilitar_propinas"`
	HabilitarComisiones             bool   `json:"habilitar_comisiones"`
	FechaCreacion                   string `json:"fecha_creacion,omitempty"`
	FechaActualizacion              string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador                  string `json:"usuario_creador,omitempty"`
	Estado                          string `json:"estado,omitempty"`
	Observaciones                   string `json:"observaciones,omitempty"`
}

// EmpresaConfiguracionOperativaPermisos es la configuracion efectiva para una operacion.
type EmpresaConfiguracionOperativaPermisos struct {
	Rol                             string `json:"rol"`
	CanalVenta                      string `json:"canal_venta,omitempty"`
	SucursalID                      int64  `json:"sucursal_id,omitempty"`
	Turno                           string `json:"turno,omitempty"`
	Fuente                          string `json:"fuente,omitempty"`
	PoliticaID                      int64  `json:"politica_id,omitempty"`
	PoliticaAplicada                bool   `json:"politica_aplicada,omitempty"`
	MetodoPagoEfectivo              bool   `json:"metodo_pago_efectivo"`
	MetodoPagoTarjetaCredito        bool   `json:"metodo_pago_tarjeta_credito"`
	MetodoPagoTarjetaDebito         bool   `json:"metodo_pago_tarjeta_debito"`
	MetodoPagoTransferenciaBancaria bool   `json:"metodo_pago_transferencia_bancaria"`
	MetodoPagoMixto                 bool   `json:"metodo_pago_mixto"`
	MetodoPagoCodigoDescuento       bool   `json:"metodo_pago_codigo_descuento"`
	HabilitarPropinas               bool   `json:"habilitar_propinas"`
	HabilitarComisiones             bool   `json:"habilitar_comisiones"`
}

func defaultEmpresaConfiguracionOperativa(empresaID int64) EmpresaConfiguracionOperativa {
	return EmpresaConfiguracionOperativa{
		EmpresaID:                       empresaID,
		MetodoPagoEfectivo:              true,
		MetodoPagoTarjetaCredito:        true,
		MetodoPagoTarjetaDebito:         true,
		MetodoPagoTransferenciaBancaria: true,
		MetodoPagoMixto:                 true,
		MetodoPagoCodigoDescuento:       true,
		HabilitarPropinas:               true,
		HabilitarComisiones:             true,
		Estado:                          "activo",
	}
}

func defaultEmpresaConfiguracionOperativaRol(empresaID int64, rol string) EmpresaConfiguracionOperativaRol {
	return EmpresaConfiguracionOperativaRol{
		EmpresaID:                       empresaID,
		Rol:                             normalizeConfiguracionOperativaRol(rol),
		MetodoPagoEfectivo:              true,
		MetodoPagoTarjetaCredito:        true,
		MetodoPagoTarjetaDebito:         true,
		MetodoPagoTransferenciaBancaria: true,
		MetodoPagoMixto:                 true,
		MetodoPagoCodigoDescuento:       true,
		HabilitarPropinas:               true,
		HabilitarComisiones:             true,
		Estado:                          "activo",
	}
}

func defaultEmpresaConfiguracionOperativaPolitica(empresaID int64) EmpresaConfiguracionOperativaPolitica {
	return EmpresaConfiguracionOperativaPolitica{
		EmpresaID:                       empresaID,
		CanalVenta:                      "",
		SucursalID:                      0,
		Turno:                           "",
		Prioridad:                       100,
		MetodoPagoEfectivo:              true,
		MetodoPagoTarjetaCredito:        true,
		MetodoPagoTarjetaDebito:         true,
		MetodoPagoTransferenciaBancaria: true,
		MetodoPagoMixto:                 true,
		MetodoPagoCodigoDescuento:       true,
		HabilitarPropinas:               true,
		HabilitarComisiones:             true,
		Estado:                          "activo",
	}
}

func normalizeConfiguracionOperativaRol(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "super_administrador", "superadmin", "super":
		return "super_administrador"
	case "administrador", "admin", "admin_empresa":
		return "admin_empresa"
	case "supervisor", "supervisor_sucursal":
		return "supervisor_sucursal"
	case "cajero":
		return "cajero"
	case "inventario":
		return "inventario"
	case "compras":
		return "compras"
	case "contabilidad", "contador":
		return "contabilidad"
	case "auditor":
		return "auditor"
	default:
		return strings.ToLower(strings.TrimSpace(raw))
	}
}

func normalizeConfiguracionOperativaEstado(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "inactivo") {
		return "inactivo"
	}
	return "activo"
}

func normalizeConfiguracionOperativaCanal(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" || value == "todos" || value == "all" || value == "*" {
		return ""
	}
	switch value {
	case "mostrador", "app", "estacion", "reserva", "online", "delivery", "kiosko", "kiosk":
		return value
	default:
		return value
	}
}

func normalizeConfiguracionOperativaTurno(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" || value == "todos" || value == "all" || value == "*" {
		return ""
	}
	return value
}

func normalizeConfiguracionOperativaEvento(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "rollback", "publicar", "simular":
		return value
	default:
		return "publicar"
	}
}

// EnsureEmpresaConfiguracionOperativaSchema crea/migra tablas de configuracion operativa de cobro.
func EnsureEmpresaConfiguracionOperativaSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_configuracion_operativa (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL UNIQUE,
			metodo_pago_efectivo INTEGER DEFAULT 1,
			metodo_pago_tarjeta_credito INTEGER DEFAULT 1,
			metodo_pago_tarjeta_debito INTEGER DEFAULT 1,
			metodo_pago_transferencia_bancaria INTEGER DEFAULT 1,
			metodo_pago_mixto INTEGER DEFAULT 1,
			metodo_pago_codigo_descuento INTEGER DEFAULT 1,
			habilitar_propinas INTEGER DEFAULT 1,
			habilitar_comisiones INTEGER DEFAULT 1,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_configuracion_operativa_empresa ON empresa_configuracion_operativa(empresa_id);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_configuracion_operativa_estado ON empresa_configuracion_operativa(empresa_id, estado);`,
		`CREATE TABLE IF NOT EXISTS empresa_configuracion_operativa_roles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			rol TEXT NOT NULL,
			metodo_pago_efectivo INTEGER DEFAULT 1,
			metodo_pago_tarjeta_credito INTEGER DEFAULT 1,
			metodo_pago_tarjeta_debito INTEGER DEFAULT 1,
			metodo_pago_transferencia_bancaria INTEGER DEFAULT 1,
			metodo_pago_mixto INTEGER DEFAULT 1,
			metodo_pago_codigo_descuento INTEGER DEFAULT 1,
			habilitar_propinas INTEGER DEFAULT 1,
			habilitar_comisiones INTEGER DEFAULT 1,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_configuracion_operativa_roles_empresa_rol ON empresa_configuracion_operativa_roles(empresa_id, rol);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_configuracion_operativa_roles_estado ON empresa_configuracion_operativa_roles(empresa_id, estado, rol);`,
		`CREATE TABLE IF NOT EXISTS empresa_configuracion_operativa_politicas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			canal_venta TEXT DEFAULT '',
			sucursal_id INTEGER DEFAULT 0,
			turno TEXT DEFAULT '',
			prioridad INTEGER DEFAULT 100,
			metodo_pago_efectivo INTEGER DEFAULT 1,
			metodo_pago_tarjeta_credito INTEGER DEFAULT 1,
			metodo_pago_tarjeta_debito INTEGER DEFAULT 1,
			metodo_pago_transferencia_bancaria INTEGER DEFAULT 1,
			metodo_pago_mixto INTEGER DEFAULT 1,
			metodo_pago_codigo_descuento INTEGER DEFAULT 1,
			habilitar_propinas INTEGER DEFAULT 1,
			habilitar_comisiones INTEGER DEFAULT 1,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_configuracion_operativa_politicas_ctx ON empresa_configuracion_operativa_politicas(empresa_id, canal_venta, sucursal_id, turno);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_configuracion_operativa_politicas_estado ON empresa_configuracion_operativa_politicas(empresa_id, estado, prioridad, id DESC);`,
		`CREATE TABLE IF NOT EXISTS empresa_configuracion_operativa_historial (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			evento TEXT DEFAULT 'publicar',
			rollback_de_historial_id INTEGER DEFAULT 0,
			snapshot_json TEXT NOT NULL,
			simulacion_json TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_configuracion_operativa_historial_empresa_fecha ON empresa_configuracion_operativa_historial(empresa_id, fecha_creacion DESC, id DESC);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa", "metodo_pago_efectivo", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa", "metodo_pago_tarjeta_credito", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa", "metodo_pago_tarjeta_debito", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa", "metodo_pago_transferencia_bancaria", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa", "metodo_pago_mixto", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa", "metodo_pago_codigo_descuento", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa", "habilitar_propinas", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa", "habilitar_comisiones", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_roles", "metodo_pago_efectivo", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_roles", "metodo_pago_tarjeta_credito", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_roles", "metodo_pago_tarjeta_debito", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_roles", "metodo_pago_transferencia_bancaria", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_roles", "metodo_pago_mixto", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_roles", "metodo_pago_codigo_descuento", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_roles", "habilitar_propinas", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_roles", "habilitar_comisiones", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_roles", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_roles", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_roles", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_roles", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_politicas", "canal_venta", "TEXT DEFAULT ''"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_politicas", "sucursal_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_politicas", "turno", "TEXT DEFAULT ''"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_politicas", "prioridad", "INTEGER DEFAULT 100"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_politicas", "metodo_pago_efectivo", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_politicas", "metodo_pago_tarjeta_credito", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_politicas", "metodo_pago_tarjeta_debito", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_politicas", "metodo_pago_transferencia_bancaria", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_politicas", "metodo_pago_mixto", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_politicas", "metodo_pago_codigo_descuento", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_politicas", "habilitar_propinas", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_politicas", "habilitar_comisiones", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_politicas", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_politicas", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_politicas", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_politicas", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_historial", "evento", "TEXT DEFAULT 'publicar'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_historial", "rollback_de_historial_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_historial", "snapshot_json", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_historial", "simulacion_json", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_historial", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_historial", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_historial", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_operativa_historial", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

// GetEmpresaConfiguracionOperativa obtiene la configuracion operativa por empresa y sus reglas por rol.
func GetEmpresaConfiguracionOperativa(dbConn *sql.DB, empresaID int64) (*EmpresaConfiguracionOperativa, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}

	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(metodo_pago_efectivo, 1),
		COALESCE(metodo_pago_tarjeta_credito, 1),
		COALESCE(metodo_pago_tarjeta_debito, 1),
		COALESCE(metodo_pago_transferencia_bancaria, 1),
		COALESCE(metodo_pago_mixto, 1),
		COALESCE(metodo_pago_codigo_descuento, 1),
		COALESCE(habilitar_propinas, 1),
		COALESCE(habilitar_comisiones, 1),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_configuracion_operativa
	WHERE empresa_id = ?
	LIMIT 1`, empresaID)

	cfg := defaultEmpresaConfiguracionOperativa(empresaID)
	var metodoEfectivo int
	var metodoTarjetaCredito int
	var metodoTarjetaDebito int
	var metodoTransferencia int
	var metodoMixto int
	var metodoCodigo int
	var habilitarPropinas int
	var habilitarComisiones int
	if err := row.Scan(
		&cfg.ID,
		&cfg.EmpresaID,
		&metodoEfectivo,
		&metodoTarjetaCredito,
		&metodoTarjetaDebito,
		&metodoTransferencia,
		&metodoMixto,
		&metodoCodigo,
		&habilitarPropinas,
		&habilitarComisiones,
		&cfg.FechaCreacion,
		&cfg.FechaActualizacion,
		&cfg.UsuarioCreador,
		&cfg.Estado,
		&cfg.Observaciones,
	); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	} else {
		cfg.MetodoPagoEfectivo = metodoEfectivo == 1
		cfg.MetodoPagoTarjetaCredito = metodoTarjetaCredito == 1
		cfg.MetodoPagoTarjetaDebito = metodoTarjetaDebito == 1
		cfg.MetodoPagoTransferenciaBancaria = metodoTransferencia == 1
		cfg.MetodoPagoMixto = metodoMixto == 1
		cfg.MetodoPagoCodigoDescuento = metodoCodigo == 1
		cfg.HabilitarPropinas = habilitarPropinas == 1
		cfg.HabilitarComisiones = habilitarComisiones == 1
		cfg.Estado = normalizeConfiguracionOperativaEstado(cfg.Estado)
	}

	roles, err := ListEmpresaConfiguracionOperativaRoles(dbConn, empresaID, true)
	if err != nil {
		return nil, err
	}
	cfg.Roles = roles

	politicas, err := ListEmpresaConfiguracionOperativaPoliticas(dbConn, empresaID, true)
	if err != nil {
		return nil, err
	}
	cfg.Politicas = politicas
	return &cfg, nil
}

// ListEmpresaConfiguracionOperativaRoles lista reglas por rol.
func ListEmpresaConfiguracionOperativaRoles(dbConn *sql.DB, empresaID int64, includeInactive bool) ([]EmpresaConfiguracionOperativaRol, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}

	query := `SELECT
		id,
		empresa_id,
		COALESCE(rol, ''),
		COALESCE(metodo_pago_efectivo, 1),
		COALESCE(metodo_pago_tarjeta_credito, 1),
		COALESCE(metodo_pago_tarjeta_debito, 1),
		COALESCE(metodo_pago_transferencia_bancaria, 1),
		COALESCE(metodo_pago_mixto, 1),
		COALESCE(metodo_pago_codigo_descuento, 1),
		COALESCE(habilitar_propinas, 1),
		COALESCE(habilitar_comisiones, 1),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_configuracion_operativa_roles
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	if !includeInactive {
		query += ` AND COALESCE(estado, 'activo') = 'activo'`
	}
	query += ` ORDER BY rol ASC, id ASC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such table") {
			return []EmpresaConfiguracionOperativaRol{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaConfiguracionOperativaRol, 0)
	for rows.Next() {
		var row EmpresaConfiguracionOperativaRol
		var metodoEfectivo int
		var metodoTarjetaCredito int
		var metodoTarjetaDebito int
		var metodoTransferencia int
		var metodoMixto int
		var metodoCodigo int
		var habilitarPropinas int
		var habilitarComisiones int
		if err := rows.Scan(
			&row.ID,
			&row.EmpresaID,
			&row.Rol,
			&metodoEfectivo,
			&metodoTarjetaCredito,
			&metodoTarjetaDebito,
			&metodoTransferencia,
			&metodoMixto,
			&metodoCodigo,
			&habilitarPropinas,
			&habilitarComisiones,
			&row.FechaCreacion,
			&row.FechaActualizacion,
			&row.UsuarioCreador,
			&row.Estado,
			&row.Observaciones,
		); err != nil {
			return nil, err
		}
		row.Rol = normalizeConfiguracionOperativaRol(row.Rol)
		row.MetodoPagoEfectivo = metodoEfectivo == 1
		row.MetodoPagoTarjetaCredito = metodoTarjetaCredito == 1
		row.MetodoPagoTarjetaDebito = metodoTarjetaDebito == 1
		row.MetodoPagoTransferenciaBancaria = metodoTransferencia == 1
		row.MetodoPagoMixto = metodoMixto == 1
		row.MetodoPagoCodigoDescuento = metodoCodigo == 1
		row.HabilitarPropinas = habilitarPropinas == 1
		row.HabilitarComisiones = habilitarComisiones == 1
		row.Estado = normalizeConfiguracionOperativaEstado(row.Estado)
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// UpsertEmpresaConfiguracionOperativa crea o actualiza configuracion base por empresa.
func UpsertEmpresaConfiguracionOperativa(dbConn *sql.DB, payload EmpresaConfiguracionOperativa) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	payload.Estado = normalizeConfiguracionOperativaEstado(payload.Estado)
	nowExpr := sqlNowExpr()
	var id int64
	err := QueryRowCompat(dbConn, `INSERT INTO empresa_configuracion_operativa (
		empresa_id,
		metodo_pago_efectivo,
		metodo_pago_tarjeta_credito,
		metodo_pago_tarjeta_debito,
		metodo_pago_transferencia_bancaria,
		metodo_pago_mixto,
		metodo_pago_codigo_descuento,
		habilitar_propinas,
		habilitar_comisiones,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, `+nowExpr+`, `+nowExpr+`, ?, ?, ?)
	ON CONFLICT(empresa_id) DO UPDATE SET
		metodo_pago_efectivo = excluded.metodo_pago_efectivo,
		metodo_pago_tarjeta_credito = excluded.metodo_pago_tarjeta_credito,
		metodo_pago_tarjeta_debito = excluded.metodo_pago_tarjeta_debito,
		metodo_pago_transferencia_bancaria = excluded.metodo_pago_transferencia_bancaria,
		metodo_pago_mixto = excluded.metodo_pago_mixto,
		metodo_pago_codigo_descuento = excluded.metodo_pago_codigo_descuento,
		habilitar_propinas = excluded.habilitar_propinas,
		habilitar_comisiones = excluded.habilitar_comisiones,
		fecha_actualizacion = `+nowExpr+`,
		usuario_creador = CASE
			WHEN trim(excluded.usuario_creador) <> '' THEN excluded.usuario_creador
			ELSE empresa_configuracion_operativa.usuario_creador
		END,
		estado = excluded.estado,
		observaciones = excluded.observaciones
	RETURNING id`,
		payload.EmpresaID,
		boolToInt(payload.MetodoPagoEfectivo),
		boolToInt(payload.MetodoPagoTarjetaCredito),
		boolToInt(payload.MetodoPagoTarjetaDebito),
		boolToInt(payload.MetodoPagoTransferenciaBancaria),
		boolToInt(payload.MetodoPagoMixto),
		boolToInt(payload.MetodoPagoCodigoDescuento),
		boolToInt(payload.HabilitarPropinas),
		boolToInt(payload.HabilitarComisiones),
		strings.TrimSpace(payload.UsuarioCreador),
		payload.Estado,
		strings.TrimSpace(payload.Observaciones),
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpsertEmpresaConfiguracionOperativaRol crea o actualiza la configuracion de un rol.
func UpsertEmpresaConfiguracionOperativaRol(dbConn *sql.DB, payload EmpresaConfiguracionOperativaRol) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	payload.Rol = normalizeConfiguracionOperativaRol(payload.Rol)
	if payload.Rol == "" {
		return 0, fmt.Errorf("rol es obligatorio")
	}
	payload.Estado = normalizeConfiguracionOperativaEstado(payload.Estado)

	nowExpr := sqlNowExpr()
	var id int64
	err := QueryRowCompat(dbConn, `INSERT INTO empresa_configuracion_operativa_roles (
		empresa_id,
		rol,
		metodo_pago_efectivo,
		metodo_pago_tarjeta_credito,
		metodo_pago_tarjeta_debito,
		metodo_pago_transferencia_bancaria,
		metodo_pago_mixto,
		metodo_pago_codigo_descuento,
		habilitar_propinas,
		habilitar_comisiones,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, `+nowExpr+`, `+nowExpr+`, ?, ?, ?)
	ON CONFLICT(empresa_id, rol) DO UPDATE SET
		metodo_pago_efectivo = excluded.metodo_pago_efectivo,
		metodo_pago_tarjeta_credito = excluded.metodo_pago_tarjeta_credito,
		metodo_pago_tarjeta_debito = excluded.metodo_pago_tarjeta_debito,
		metodo_pago_transferencia_bancaria = excluded.metodo_pago_transferencia_bancaria,
		metodo_pago_mixto = excluded.metodo_pago_mixto,
		metodo_pago_codigo_descuento = excluded.metodo_pago_codigo_descuento,
		habilitar_propinas = excluded.habilitar_propinas,
		habilitar_comisiones = excluded.habilitar_comisiones,
		fecha_actualizacion = `+nowExpr+`,
		usuario_creador = CASE
			WHEN trim(excluded.usuario_creador) <> '' THEN excluded.usuario_creador
			ELSE empresa_configuracion_operativa_roles.usuario_creador
		END,
		estado = excluded.estado,
		observaciones = excluded.observaciones
	RETURNING id`,
		payload.EmpresaID,
		payload.Rol,
		boolToInt(payload.MetodoPagoEfectivo),
		boolToInt(payload.MetodoPagoTarjetaCredito),
		boolToInt(payload.MetodoPagoTarjetaDebito),
		boolToInt(payload.MetodoPagoTransferenciaBancaria),
		boolToInt(payload.MetodoPagoMixto),
		boolToInt(payload.MetodoPagoCodigoDescuento),
		boolToInt(payload.HabilitarPropinas),
		boolToInt(payload.HabilitarComisiones),
		strings.TrimSpace(payload.UsuarioCreador),
		payload.Estado,
		strings.TrimSpace(payload.Observaciones),
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ListEmpresaConfiguracionOperativaPoliticas lista reglas contextuales por empresa.
func ListEmpresaConfiguracionOperativaPoliticas(dbConn *sql.DB, empresaID int64, includeInactive bool) ([]EmpresaConfiguracionOperativaPolitica, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}

	query := `SELECT
		id,
		empresa_id,
		COALESCE(canal_venta, ''),
		COALESCE(sucursal_id, 0),
		COALESCE(turno, ''),
		COALESCE(prioridad, 100),
		COALESCE(metodo_pago_efectivo, 1),
		COALESCE(metodo_pago_tarjeta_credito, 1),
		COALESCE(metodo_pago_tarjeta_debito, 1),
		COALESCE(metodo_pago_transferencia_bancaria, 1),
		COALESCE(metodo_pago_mixto, 1),
		COALESCE(metodo_pago_codigo_descuento, 1),
		COALESCE(habilitar_propinas, 1),
		COALESCE(habilitar_comisiones, 1),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_configuracion_operativa_politicas
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	if !includeInactive {
		query += ` AND COALESCE(estado, 'activo') = 'activo'`
	}
	query += ` ORDER BY prioridad ASC, id DESC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such table") {
			return []EmpresaConfiguracionOperativaPolitica{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaConfiguracionOperativaPolitica, 0)
	for rows.Next() {
		var row EmpresaConfiguracionOperativaPolitica
		var metodoEfectivo int
		var metodoTarjetaCredito int
		var metodoTarjetaDebito int
		var metodoTransferencia int
		var metodoMixto int
		var metodoCodigo int
		var habilitarPropinas int
		var habilitarComisiones int
		if err := rows.Scan(
			&row.ID,
			&row.EmpresaID,
			&row.CanalVenta,
			&row.SucursalID,
			&row.Turno,
			&row.Prioridad,
			&metodoEfectivo,
			&metodoTarjetaCredito,
			&metodoTarjetaDebito,
			&metodoTransferencia,
			&metodoMixto,
			&metodoCodigo,
			&habilitarPropinas,
			&habilitarComisiones,
			&row.FechaCreacion,
			&row.FechaActualizacion,
			&row.UsuarioCreador,
			&row.Estado,
			&row.Observaciones,
		); err != nil {
			return nil, err
		}
		row.CanalVenta = normalizeConfiguracionOperativaCanal(row.CanalVenta)
		row.Turno = normalizeConfiguracionOperativaTurno(row.Turno)
		if row.SucursalID < 0 {
			row.SucursalID = 0
		}
		if row.Prioridad <= 0 {
			row.Prioridad = 100
		}
		row.MetodoPagoEfectivo = metodoEfectivo == 1
		row.MetodoPagoTarjetaCredito = metodoTarjetaCredito == 1
		row.MetodoPagoTarjetaDebito = metodoTarjetaDebito == 1
		row.MetodoPagoTransferenciaBancaria = metodoTransferencia == 1
		row.MetodoPagoMixto = metodoMixto == 1
		row.MetodoPagoCodigoDescuento = metodoCodigo == 1
		row.HabilitarPropinas = habilitarPropinas == 1
		row.HabilitarComisiones = habilitarComisiones == 1
		row.Estado = normalizeConfiguracionOperativaEstado(row.Estado)
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// UpsertEmpresaConfiguracionOperativaPolitica crea o actualiza una politica contextual.
func UpsertEmpresaConfiguracionOperativaPolitica(dbConn *sql.DB, payload EmpresaConfiguracionOperativaPolitica) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	payload.CanalVenta = normalizeConfiguracionOperativaCanal(payload.CanalVenta)
	payload.Turno = normalizeConfiguracionOperativaTurno(payload.Turno)
	if payload.SucursalID < 0 {
		payload.SucursalID = 0
	}
	if payload.Prioridad <= 0 {
		payload.Prioridad = 100
	}
	payload.Estado = normalizeConfiguracionOperativaEstado(payload.Estado)

	nowExpr := sqlNowExpr()
	var id int64
	err := QueryRowCompat(dbConn, `INSERT INTO empresa_configuracion_operativa_politicas (
		empresa_id,
		canal_venta,
		sucursal_id,
		turno,
		prioridad,
		metodo_pago_efectivo,
		metodo_pago_tarjeta_credito,
		metodo_pago_tarjeta_debito,
		metodo_pago_transferencia_bancaria,
		metodo_pago_mixto,
		metodo_pago_codigo_descuento,
		habilitar_propinas,
		habilitar_comisiones,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, `+nowExpr+`, `+nowExpr+`, ?, ?, ?)
	ON CONFLICT(empresa_id, canal_venta, sucursal_id, turno) DO UPDATE SET
		prioridad = excluded.prioridad,
		metodo_pago_efectivo = excluded.metodo_pago_efectivo,
		metodo_pago_tarjeta_credito = excluded.metodo_pago_tarjeta_credito,
		metodo_pago_tarjeta_debito = excluded.metodo_pago_tarjeta_debito,
		metodo_pago_transferencia_bancaria = excluded.metodo_pago_transferencia_bancaria,
		metodo_pago_mixto = excluded.metodo_pago_mixto,
		metodo_pago_codigo_descuento = excluded.metodo_pago_codigo_descuento,
		habilitar_propinas = excluded.habilitar_propinas,
		habilitar_comisiones = excluded.habilitar_comisiones,
		fecha_actualizacion = `+nowExpr+`,
		usuario_creador = CASE
			WHEN trim(excluded.usuario_creador) <> '' THEN excluded.usuario_creador
			ELSE empresa_configuracion_operativa_politicas.usuario_creador
		END,
		estado = excluded.estado,
		observaciones = excluded.observaciones
	RETURNING id`,
		payload.EmpresaID,
		payload.CanalVenta,
		payload.SucursalID,
		payload.Turno,
		payload.Prioridad,
		boolToInt(payload.MetodoPagoEfectivo),
		boolToInt(payload.MetodoPagoTarjetaCredito),
		boolToInt(payload.MetodoPagoTarjetaDebito),
		boolToInt(payload.MetodoPagoTransferenciaBancaria),
		boolToInt(payload.MetodoPagoMixto),
		boolToInt(payload.MetodoPagoCodigoDescuento),
		boolToInt(payload.HabilitarPropinas),
		boolToInt(payload.HabilitarComisiones),
		strings.TrimSpace(payload.UsuarioCreador),
		payload.Estado,
		strings.TrimSpace(payload.Observaciones),
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ResolveEmpresaConfiguracionOperativaParaRol calcula permisos efectivos para un rol.
func ResolveEmpresaConfiguracionOperativaParaRol(cfg *EmpresaConfiguracionOperativa, role string) EmpresaConfiguracionOperativaPermisos {
	return ResolveEmpresaConfiguracionOperativaConContexto(cfg, EmpresaConfiguracionOperativaContexto{Rol: role})
}

// ResolveEmpresaConfiguracionOperativaConContexto calcula permisos efectivos para rol/canal/sucursal/turno.
func ResolveEmpresaConfiguracionOperativaConContexto(cfg *EmpresaConfiguracionOperativa, contexto EmpresaConfiguracionOperativaContexto) EmpresaConfiguracionOperativaPermisos {
	resolved := EmpresaConfiguracionOperativaPermisos{
		Rol:                             normalizeConfiguracionOperativaRol(contexto.Rol),
		CanalVenta:                      normalizeConfiguracionOperativaCanal(contexto.CanalVenta),
		SucursalID:                      contexto.SucursalID,
		Turno:                           normalizeConfiguracionOperativaTurno(contexto.Turno),
		Fuente:                          "default",
		MetodoPagoEfectivo:              true,
		MetodoPagoTarjetaCredito:        true,
		MetodoPagoTarjetaDebito:         true,
		MetodoPagoTransferenciaBancaria: true,
		MetodoPagoMixto:                 true,
		MetodoPagoCodigoDescuento:       true,
		HabilitarPropinas:               true,
		HabilitarComisiones:             true,
	}
	if resolved.SucursalID < 0 {
		resolved.SucursalID = 0
	}
	if cfg == nil {
		return resolved
	}

	resolved.MetodoPagoEfectivo = cfg.MetodoPagoEfectivo
	resolved.MetodoPagoTarjetaCredito = cfg.MetodoPagoTarjetaCredito
	resolved.MetodoPagoTarjetaDebito = cfg.MetodoPagoTarjetaDebito
	resolved.MetodoPagoTransferenciaBancaria = cfg.MetodoPagoTransferenciaBancaria
	resolved.MetodoPagoMixto = cfg.MetodoPagoMixto
	resolved.MetodoPagoCodigoDescuento = cfg.MetodoPagoCodigoDescuento
	resolved.HabilitarPropinas = cfg.HabilitarPropinas
	resolved.HabilitarComisiones = cfg.HabilitarComisiones
	resolved.Fuente = "empresa"

	for _, row := range cfg.Roles {
		if normalizeConfiguracionOperativaRol(row.Rol) != resolved.Rol {
			continue
		}
		if normalizeConfiguracionOperativaEstado(row.Estado) != "activo" {
			continue
		}
		resolved.MetodoPagoEfectivo = row.MetodoPagoEfectivo
		resolved.MetodoPagoTarjetaCredito = row.MetodoPagoTarjetaCredito
		resolved.MetodoPagoTarjetaDebito = row.MetodoPagoTarjetaDebito
		resolved.MetodoPagoTransferenciaBancaria = row.MetodoPagoTransferenciaBancaria
		resolved.MetodoPagoMixto = row.MetodoPagoMixto
		resolved.MetodoPagoCodigoDescuento = row.MetodoPagoCodigoDescuento
		resolved.HabilitarPropinas = row.HabilitarPropinas
		resolved.HabilitarComisiones = row.HabilitarComisiones
		resolved.Fuente = "rol"
		break
	}

	bestIdx := -1
	bestScore := -1
	bestPriority := int(^uint(0) >> 1)
	for idx, policy := range cfg.Politicas {
		if normalizeConfiguracionOperativaEstado(policy.Estado) != "activo" {
			continue
		}
		if !configuracionOperativaPoliticaMatch(policy, contexto) {
			continue
		}
		score := configuracionOperativaPoliticaSpecificity(policy)
		priority := policy.Prioridad
		if priority <= 0 {
			priority = 100
		}
		if score > bestScore || (score == bestScore && priority < bestPriority) {
			bestIdx = idx
			bestScore = score
			bestPriority = priority
		}
	}

	if bestIdx >= 0 {
		policy := cfg.Politicas[bestIdx]
		resolved.MetodoPagoEfectivo = policy.MetodoPagoEfectivo
		resolved.MetodoPagoTarjetaCredito = policy.MetodoPagoTarjetaCredito
		resolved.MetodoPagoTarjetaDebito = policy.MetodoPagoTarjetaDebito
		resolved.MetodoPagoTransferenciaBancaria = policy.MetodoPagoTransferenciaBancaria
		resolved.MetodoPagoMixto = policy.MetodoPagoMixto
		resolved.MetodoPagoCodigoDescuento = policy.MetodoPagoCodigoDescuento
		resolved.HabilitarPropinas = policy.HabilitarPropinas
		resolved.HabilitarComisiones = policy.HabilitarComisiones
		resolved.PoliticaID = policy.ID
		resolved.PoliticaAplicada = true
		resolved.Fuente = "politica"
	}

	return resolved
}

func configuracionOperativaPoliticaMatch(policy EmpresaConfiguracionOperativaPolitica, contexto EmpresaConfiguracionOperativaContexto) bool {
	policyCanal := normalizeConfiguracionOperativaCanal(policy.CanalVenta)
	ctxCanal := normalizeConfiguracionOperativaCanal(contexto.CanalVenta)
	if policyCanal != "" && policyCanal != ctxCanal {
		return false
	}

	policyTurno := normalizeConfiguracionOperativaTurno(policy.Turno)
	ctxTurno := normalizeConfiguracionOperativaTurno(contexto.Turno)
	if policyTurno != "" && policyTurno != ctxTurno {
		return false
	}

	policySucursal := policy.SucursalID
	ctxSucursal := contexto.SucursalID
	if policySucursal < 0 {
		policySucursal = 0
	}
	if ctxSucursal < 0 {
		ctxSucursal = 0
	}
	if policySucursal > 0 && policySucursal != ctxSucursal {
		return false
	}

	return true
}

func configuracionOperativaPoliticaSpecificity(policy EmpresaConfiguracionOperativaPolitica) int {
	score := 0
	if normalizeConfiguracionOperativaCanal(policy.CanalVenta) != "" {
		score++
	}
	if policy.SucursalID > 0 {
		score++
	}
	if normalizeConfiguracionOperativaTurno(policy.Turno) != "" {
		score++
	}
	return score
}

// GetEmpresaConfiguracionOperativaPermisos obtiene permisos efectivos para empresa+rol.
func GetEmpresaConfiguracionOperativaPermisos(dbConn *sql.DB, empresaID int64, role string) (*EmpresaConfiguracionOperativaPermisos, error) {
	cfg, err := GetEmpresaConfiguracionOperativa(dbConn, empresaID)
	if err != nil {
		return nil, err
	}
	resolved := ResolveEmpresaConfiguracionOperativaParaRol(cfg, role)
	return &resolved, nil
}

// GetEmpresaConfiguracionOperativaPermisosContexto obtiene permisos efectivos para empresa+contexto operativo.
func GetEmpresaConfiguracionOperativaPermisosContexto(dbConn *sql.DB, empresaID int64, contexto EmpresaConfiguracionOperativaContexto) (*EmpresaConfiguracionOperativaPermisos, error) {
	cfg, err := GetEmpresaConfiguracionOperativa(dbConn, empresaID)
	if err != nil {
		return nil, err
	}
	resolved := ResolveEmpresaConfiguracionOperativaConContexto(cfg, contexto)
	return &resolved, nil
}

// CreateEmpresaConfiguracionOperativaHistorialSnapshot guarda un snapshot para trazabilidad y rollback.
func CreateEmpresaConfiguracionOperativaHistorialSnapshot(dbConn *sql.DB, payload EmpresaConfiguracionOperativaHistorialSnapshot) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	payload.Evento = normalizeConfiguracionOperativaEvento(payload.Evento)
	payload.Estado = normalizeConfiguracionOperativaEstado(payload.Estado)

	snapshotJSON := strings.TrimSpace(payload.SnapshotJSON)
	if snapshotJSON == "" {
		cfg, err := GetEmpresaConfiguracionOperativa(dbConn, payload.EmpresaID)
		if err != nil {
			return 0, err
		}
		raw, err := json.Marshal(cfg)
		if err != nil {
			return 0, err
		}
		snapshotJSON = string(raw)
	}

	nowExpr := sqlNowExpr()
	var id int64
	err := QueryRowCompat(dbConn, `INSERT INTO empresa_configuracion_operativa_historial (
		empresa_id,
		evento,
		rollback_de_historial_id,
		snapshot_json,
		simulacion_json,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, `+nowExpr+`, `+nowExpr+`, ?, ?, ?)
	RETURNING id`,
		payload.EmpresaID,
		payload.Evento,
		payload.RollbackDeHistorial,
		snapshotJSON,
		strings.TrimSpace(payload.SimulacionJSON),
		strings.TrimSpace(payload.UsuarioCreador),
		payload.Estado,
		strings.TrimSpace(payload.Observaciones),
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ListEmpresaConfiguracionOperativaHistorialSnapshots lista snapshots recientes por empresa.
func ListEmpresaConfiguracionOperativaHistorialSnapshots(dbConn *sql.DB, empresaID int64, limit int) ([]EmpresaConfiguracionOperativaHistorialSnapshot, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}

	rows, err := dbConn.Query(`SELECT
		id,
		empresa_id,
		COALESCE(evento, 'publicar'),
		COALESCE(rollback_de_historial_id, 0),
		COALESCE(snapshot_json, ''),
		COALESCE(simulacion_json, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_configuracion_operativa_historial
	WHERE empresa_id = ?
	ORDER BY id DESC
	LIMIT ?`, empresaID, limit)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such table") {
			return []EmpresaConfiguracionOperativaHistorialSnapshot{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaConfiguracionOperativaHistorialSnapshot, 0)
	for rows.Next() {
		var row EmpresaConfiguracionOperativaHistorialSnapshot
		if err := rows.Scan(
			&row.ID,
			&row.EmpresaID,
			&row.Evento,
			&row.RollbackDeHistorial,
			&row.SnapshotJSON,
			&row.SimulacionJSON,
			&row.FechaCreacion,
			&row.FechaActualizacion,
			&row.UsuarioCreador,
			&row.Estado,
			&row.Observaciones,
		); err != nil {
			return nil, err
		}
		row.Evento = normalizeConfiguracionOperativaEvento(row.Evento)
		row.Estado = normalizeConfiguracionOperativaEstado(row.Estado)
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// GetEmpresaConfiguracionOperativaHistorialSnapshotByID obtiene un snapshot especifico.
func GetEmpresaConfiguracionOperativaHistorialSnapshotByID(dbConn *sql.DB, empresaID, historialID int64) (*EmpresaConfiguracionOperativaHistorialSnapshot, error) {
	if empresaID <= 0 || historialID <= 0 {
		return nil, nil
	}

	var row EmpresaConfiguracionOperativaHistorialSnapshot
	err := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(evento, 'publicar'),
		COALESCE(rollback_de_historial_id, 0),
		COALESCE(snapshot_json, ''),
		COALESCE(simulacion_json, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_configuracion_operativa_historial
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, historialID).Scan(
		&row.ID,
		&row.EmpresaID,
		&row.Evento,
		&row.RollbackDeHistorial,
		&row.SnapshotJSON,
		&row.SimulacionJSON,
		&row.FechaCreacion,
		&row.FechaActualizacion,
		&row.UsuarioCreador,
		&row.Estado,
		&row.Observaciones,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	row.Evento = normalizeConfiguracionOperativaEvento(row.Evento)
	row.Estado = normalizeConfiguracionOperativaEstado(row.Estado)
	return &row, nil
}

// ApplyEmpresaConfiguracionOperativaRollback aplica un snapshot previo y registra el evento de rollback.
func ApplyEmpresaConfiguracionOperativaRollback(dbConn *sql.DB, empresaID, historialID int64, usuario, observaciones string) (int64, error) {
	target, err := GetEmpresaConfiguracionOperativaHistorialSnapshotByID(dbConn, empresaID, historialID)
	if err != nil {
		return 0, err
	}
	if target == nil {
		return 0, fmt.Errorf("snapshot de historial no encontrado")
	}

	raw := strings.TrimSpace(target.SnapshotJSON)
	if raw == "" {
		return 0, fmt.Errorf("snapshot vacio")
	}

	var restored EmpresaConfiguracionOperativa
	if err := json.Unmarshal([]byte(raw), &restored); err != nil {
		var envelope struct {
			Configuracion EmpresaConfiguracionOperativa `json:"configuracion"`
		}
		if err2 := json.Unmarshal([]byte(raw), &envelope); err2 != nil {
			return 0, fmt.Errorf("snapshot invalido: %w", err)
		}
		restored = envelope.Configuracion
	}
	restored.EmpresaID = empresaID
	if restored.Estado == "" {
		restored.Estado = "activo"
	}
	restored.Estado = normalizeConfiguracionOperativaEstado(restored.Estado)
	restored.UsuarioCreador = strings.TrimSpace(usuario)
	if strings.TrimSpace(restored.Observaciones) == "" {
		restored.Observaciones = strings.TrimSpace(observaciones)
	}

	if _, err := UpsertEmpresaConfiguracionOperativa(dbConn, restored); err != nil {
		return 0, err
	}

	if _, err := dbConn.Exec(`UPDATE empresa_configuracion_operativa_roles
		SET estado = 'inactivo', fecha_actualizacion = datetime('now','localtime')
		WHERE empresa_id = ?`, empresaID); err != nil {
		return 0, err
	}
	for _, role := range restored.Roles {
		role.EmpresaID = empresaID
		role.Rol = normalizeConfiguracionOperativaRol(role.Rol)
		if role.Rol == "" {
			continue
		}
		role.Estado = normalizeConfiguracionOperativaEstado(role.Estado)
		role.UsuarioCreador = strings.TrimSpace(usuario)
		if _, err := UpsertEmpresaConfiguracionOperativaRol(dbConn, role); err != nil {
			return 0, err
		}
	}

	if _, err := dbConn.Exec(`UPDATE empresa_configuracion_operativa_politicas
		SET estado = 'inactivo', fecha_actualizacion = datetime('now','localtime')
		WHERE empresa_id = ?`, empresaID); err != nil {
		return 0, err
	}
	for _, policy := range restored.Politicas {
		policy.EmpresaID = empresaID
		policy.CanalVenta = normalizeConfiguracionOperativaCanal(policy.CanalVenta)
		policy.Turno = normalizeConfiguracionOperativaTurno(policy.Turno)
		if policy.SucursalID < 0 {
			policy.SucursalID = 0
		}
		policy.Estado = normalizeConfiguracionOperativaEstado(policy.Estado)
		policy.UsuarioCreador = strings.TrimSpace(usuario)
		if _, err := UpsertEmpresaConfiguracionOperativaPolitica(dbConn, policy); err != nil {
			return 0, err
		}
	}

	currentCfg, err := GetEmpresaConfiguracionOperativa(dbConn, empresaID)
	if err != nil {
		return 0, err
	}
	currentRaw, err := json.Marshal(currentCfg)
	if err != nil {
		return 0, err
	}

	rollbackID, err := CreateEmpresaConfiguracionOperativaHistorialSnapshot(dbConn, EmpresaConfiguracionOperativaHistorialSnapshot{
		EmpresaID:           empresaID,
		Evento:              "rollback",
		RollbackDeHistorial: historialID,
		SnapshotJSON:        string(currentRaw),
		UsuarioCreador:      strings.TrimSpace(usuario),
		Estado:              "activo",
		Observaciones:       strings.TrimSpace(observaciones),
	})
	if err != nil {
		return 0, err
	}

	return rollbackID, nil
}

// IsMetodoPagoHabilitado valida si el metodo de pago esta permitido en la configuracion efectiva.
func (p EmpresaConfiguracionOperativaPermisos) IsMetodoPagoHabilitado(metodoPago string) bool {
	normalized := NormalizeMetodoPagoCarrito(metodoPago)
	switch normalized {
	case "efectivo":
		return p.MetodoPagoEfectivo
	case "tarjeta_credito":
		return p.MetodoPagoTarjetaCredito
	case "tarjeta_debito":
		return p.MetodoPagoTarjetaDebito
	case "transferencia_bancaria":
		return p.MetodoPagoTransferenciaBancaria
	case "mixto":
		return p.MetodoPagoMixto
	case "codigo_descuento":
		return p.MetodoPagoCodigoDescuento
	default:
		return false
	}
}
