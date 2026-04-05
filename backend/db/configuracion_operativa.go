package db

import (
	"database/sql"
	"fmt"
	"strings"
)

// EmpresaConfiguracionOperativa define controles de cobro por empresa.
type EmpresaConfiguracionOperativa struct {
	ID                              int64                              `json:"id"`
	EmpresaID                       int64                              `json:"empresa_id"`
	MetodoPagoEfectivo              bool                               `json:"metodo_pago_efectivo"`
	MetodoPagoTarjetaCredito        bool                               `json:"metodo_pago_tarjeta_credito"`
	MetodoPagoTarjetaDebito         bool                               `json:"metodo_pago_tarjeta_debito"`
	MetodoPagoTransferenciaBancaria bool                               `json:"metodo_pago_transferencia_bancaria"`
	MetodoPagoMixto                 bool                               `json:"metodo_pago_mixto"`
	MetodoPagoCodigoDescuento       bool                               `json:"metodo_pago_codigo_descuento"`
	HabilitarPropinas               bool                               `json:"habilitar_propinas"`
	HabilitarComisiones             bool                               `json:"habilitar_comisiones"`
	Roles                           []EmpresaConfiguracionOperativaRol `json:"roles,omitempty"`
	FechaCreacion                   string                             `json:"fecha_creacion,omitempty"`
	FechaActualizacion              string                             `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador                  string                             `json:"usuario_creador,omitempty"`
	Estado                          string                             `json:"estado,omitempty"`
	Observaciones                   string                             `json:"observaciones,omitempty"`
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
	res, err := dbConn.Exec(`INSERT INTO empresa_configuracion_operativa (
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
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'), ?, ?, ?)
	ON CONFLICT(empresa_id) DO UPDATE SET
		metodo_pago_efectivo = excluded.metodo_pago_efectivo,
		metodo_pago_tarjeta_credito = excluded.metodo_pago_tarjeta_credito,
		metodo_pago_tarjeta_debito = excluded.metodo_pago_tarjeta_debito,
		metodo_pago_transferencia_bancaria = excluded.metodo_pago_transferencia_bancaria,
		metodo_pago_mixto = excluded.metodo_pago_mixto,
		metodo_pago_codigo_descuento = excluded.metodo_pago_codigo_descuento,
		habilitar_propinas = excluded.habilitar_propinas,
		habilitar_comisiones = excluded.habilitar_comisiones,
		fecha_actualizacion = datetime('now','localtime'),
		usuario_creador = CASE
			WHEN trim(excluded.usuario_creador) <> '' THEN excluded.usuario_creador
			ELSE empresa_configuracion_operativa.usuario_creador
		END,
		estado = excluded.estado,
		observaciones = excluded.observaciones`,
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
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err == nil && id > 0 {
		return id, nil
	}
	if err := dbConn.QueryRow(`SELECT id FROM empresa_configuracion_operativa WHERE empresa_id = ? LIMIT 1`, payload.EmpresaID).Scan(&id); err != nil {
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

	res, err := dbConn.Exec(`INSERT INTO empresa_configuracion_operativa_roles (
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
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'), ?, ?, ?)
	ON CONFLICT(empresa_id, rol) DO UPDATE SET
		metodo_pago_efectivo = excluded.metodo_pago_efectivo,
		metodo_pago_tarjeta_credito = excluded.metodo_pago_tarjeta_credito,
		metodo_pago_tarjeta_debito = excluded.metodo_pago_tarjeta_debito,
		metodo_pago_transferencia_bancaria = excluded.metodo_pago_transferencia_bancaria,
		metodo_pago_mixto = excluded.metodo_pago_mixto,
		metodo_pago_codigo_descuento = excluded.metodo_pago_codigo_descuento,
		habilitar_propinas = excluded.habilitar_propinas,
		habilitar_comisiones = excluded.habilitar_comisiones,
		fecha_actualizacion = datetime('now','localtime'),
		usuario_creador = CASE
			WHEN trim(excluded.usuario_creador) <> '' THEN excluded.usuario_creador
			ELSE empresa_configuracion_operativa_roles.usuario_creador
		END,
		estado = excluded.estado,
		observaciones = excluded.observaciones`,
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
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err == nil && id > 0 {
		return id, nil
	}
	if err := dbConn.QueryRow(`SELECT id FROM empresa_configuracion_operativa_roles WHERE empresa_id = ? AND rol = ? LIMIT 1`, payload.EmpresaID, payload.Rol).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// ResolveEmpresaConfiguracionOperativaParaRol calcula permisos efectivos para un rol.
func ResolveEmpresaConfiguracionOperativaParaRol(cfg *EmpresaConfiguracionOperativa, role string) EmpresaConfiguracionOperativaPermisos {
	resolved := EmpresaConfiguracionOperativaPermisos{
		Rol:                             normalizeConfiguracionOperativaRol(role),
		MetodoPagoEfectivo:              true,
		MetodoPagoTarjetaCredito:        true,
		MetodoPagoTarjetaDebito:         true,
		MetodoPagoTransferenciaBancaria: true,
		MetodoPagoMixto:                 true,
		MetodoPagoCodigoDescuento:       true,
		HabilitarPropinas:               true,
		HabilitarComisiones:             true,
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
		break
	}
	return resolved
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
