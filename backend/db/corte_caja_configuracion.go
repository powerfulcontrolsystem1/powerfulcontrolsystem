package db

import (
	"database/sql"
	"errors"
	"strings"
)

// EmpresaCorteCajaConfiguracion controla que bloques y metricas aparecen en el
// reporte de corte de caja por empresa.
type EmpresaCorteCajaConfiguracion struct {
	ID                     int64  `json:"id"`
	EmpresaID              int64  `json:"empresa_id"`
	MostrarResumen         bool   `json:"mostrar_resumen"`
	MostrarNumeroFacturas  bool   `json:"mostrar_numero_facturas"`
	MostrarTotalVentas     bool   `json:"mostrar_total_ventas"`
	MostrarEfectivo        bool   `json:"mostrar_efectivo"`
	MostrarDebito          bool   `json:"mostrar_debito"`
	MostrarCredito         bool   `json:"mostrar_credito"`
	MostrarTransferencias  bool   `json:"mostrar_transferencias"`
	MostrarOtrosMedios     bool   `json:"mostrar_otros_medios"`
	MostrarIngresos        bool   `json:"mostrar_ingresos"`
	MostrarEgresos         bool   `json:"mostrar_egresos"`
	MostrarAnulaciones     bool   `json:"mostrar_anulaciones"`
	MostrarDevoluciones    bool   `json:"mostrar_devoluciones"`
	MostrarCajaEsperada    bool   `json:"mostrar_caja_esperada"`
	MostrarDiferenciaCaja  bool   `json:"mostrar_diferencia_caja"`
	MostrarVentasDetalle   bool   `json:"mostrar_ventas_detalle"`
	MostrarMovimientos     bool   `json:"mostrar_movimientos"`
	MostrarItems           bool   `json:"mostrar_items"`
	MostrarSensoresPuertas bool   `json:"mostrar_sensores_puertas"`
	MostrarAuditoria       bool   `json:"mostrar_auditoria"`
	FormatoImpresion       string `json:"formato_impresion,omitempty"`
	FechaCreacion          string `json:"fecha_creacion,omitempty"`
	FechaActualizacion     string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador         string `json:"usuario_creador,omitempty"`
	Estado                 string `json:"estado,omitempty"`
	Observaciones          string `json:"observaciones,omitempty"`
}

func DefaultEmpresaCorteCajaConfiguracion(empresaID int64) EmpresaCorteCajaConfiguracion {
	return normalizeEmpresaCorteCajaConfiguracion(EmpresaCorteCajaConfiguracion{
		EmpresaID:              empresaID,
		MostrarResumen:         true,
		MostrarNumeroFacturas:  true,
		MostrarTotalVentas:     true,
		MostrarEfectivo:        true,
		MostrarDebito:          true,
		MostrarCredito:         true,
		MostrarTransferencias:  true,
		MostrarOtrosMedios:     true,
		MostrarIngresos:        true,
		MostrarEgresos:         true,
		MostrarAnulaciones:     true,
		MostrarDevoluciones:    true,
		MostrarCajaEsperada:    true,
		MostrarDiferenciaCaja:  true,
		MostrarVentasDetalle:   true,
		MostrarMovimientos:     true,
		MostrarItems:           true,
		MostrarSensoresPuertas: true,
		MostrarAuditoria:       true,
		FormatoImpresion:       "carta",
		Estado:                 "activo",
	})
}

func EnsureEmpresaCorteCajaConfiguracionSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_corte_caja_configuracion (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL UNIQUE,
			mostrar_resumen INTEGER DEFAULT 1,
			mostrar_numero_facturas INTEGER DEFAULT 1,
			mostrar_total_ventas INTEGER DEFAULT 1,
			mostrar_efectivo INTEGER DEFAULT 1,
			mostrar_debito INTEGER DEFAULT 1,
			mostrar_credito INTEGER DEFAULT 1,
			mostrar_transferencias INTEGER DEFAULT 1,
			mostrar_otros_medios INTEGER DEFAULT 1,
			mostrar_ingresos INTEGER DEFAULT 1,
			mostrar_egresos INTEGER DEFAULT 1,
			mostrar_anulaciones INTEGER DEFAULT 1,
			mostrar_devoluciones INTEGER DEFAULT 1,
			mostrar_caja_esperada INTEGER DEFAULT 1,
			mostrar_diferencia_caja INTEGER DEFAULT 1,
			mostrar_ventas_detalle INTEGER DEFAULT 1,
			mostrar_movimientos INTEGER DEFAULT 1,
			mostrar_items INTEGER DEFAULT 1,
			mostrar_sensores_puertas INTEGER DEFAULT 1,
			mostrar_auditoria INTEGER DEFAULT 1,
			formato_impresion TEXT DEFAULT 'carta',
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_corte_caja_configuracion_empresa ON empresa_corte_caja_configuracion(empresa_id);`,
	}
	for _, stmt := range stmts {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	columns := []struct {
		name string
		def  string
	}{
		{"mostrar_resumen", "INTEGER DEFAULT 1"},
		{"mostrar_numero_facturas", "INTEGER DEFAULT 1"},
		{"mostrar_total_ventas", "INTEGER DEFAULT 1"},
		{"mostrar_efectivo", "INTEGER DEFAULT 1"},
		{"mostrar_debito", "INTEGER DEFAULT 1"},
		{"mostrar_credito", "INTEGER DEFAULT 1"},
		{"mostrar_transferencias", "INTEGER DEFAULT 1"},
		{"mostrar_otros_medios", "INTEGER DEFAULT 1"},
		{"mostrar_ingresos", "INTEGER DEFAULT 1"},
		{"mostrar_egresos", "INTEGER DEFAULT 1"},
		{"mostrar_anulaciones", "INTEGER DEFAULT 1"},
		{"mostrar_devoluciones", "INTEGER DEFAULT 1"},
		{"mostrar_caja_esperada", "INTEGER DEFAULT 1"},
		{"mostrar_diferencia_caja", "INTEGER DEFAULT 1"},
		{"mostrar_ventas_detalle", "INTEGER DEFAULT 1"},
		{"mostrar_movimientos", "INTEGER DEFAULT 1"},
		{"mostrar_items", "INTEGER DEFAULT 1"},
		{"mostrar_sensores_puertas", "INTEGER DEFAULT 1"},
		{"mostrar_auditoria", "INTEGER DEFAULT 1"},
		{"formato_impresion", "TEXT DEFAULT 'carta'"},
		{"fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"},
		{"usuario_creador", "TEXT"},
		{"estado", "TEXT DEFAULT 'activo'"},
		{"observaciones", "TEXT"},
	}
	for _, col := range columns {
		if err := ensureColumnIfMissing(dbConn, "empresa_corte_caja_configuracion", col.name, col.def); err != nil {
			return err
		}
	}
	return nil
}

func GetEmpresaCorteCajaConfiguracion(dbConn *sql.DB, empresaID int64) (*EmpresaCorteCajaConfiguracion, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, errors.New("empresa_id invalido")
	}
	if err := EnsureEmpresaCorteCajaConfiguracionSchema(dbConn); err != nil {
		return nil, err
	}
	row := queryRowSQLCompat(dbConn, `SELECT
		id, empresa_id,
		COALESCE(mostrar_resumen, 1),
		COALESCE(mostrar_numero_facturas, 1),
		COALESCE(mostrar_total_ventas, 1),
		COALESCE(mostrar_efectivo, 1),
		COALESCE(mostrar_debito, 1),
		COALESCE(mostrar_credito, 1),
		COALESCE(mostrar_transferencias, 1),
		COALESCE(mostrar_otros_medios, 1),
		COALESCE(mostrar_ingresos, 1),
		COALESCE(mostrar_egresos, 1),
		COALESCE(mostrar_anulaciones, 1),
		COALESCE(mostrar_devoluciones, 1),
		COALESCE(mostrar_caja_esperada, 1),
		COALESCE(mostrar_diferencia_caja, 1),
		COALESCE(mostrar_ventas_detalle, 1),
		COALESCE(mostrar_movimientos, 1),
		COALESCE(mostrar_items, 1),
		COALESCE(mostrar_sensores_puertas, 1),
		COALESCE(mostrar_auditoria, 1),
		COALESCE(formato_impresion, 'carta'),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_corte_caja_configuracion
	WHERE empresa_id = ?
	LIMIT 1`, empresaID)

	var out EmpresaCorteCajaConfiguracion
	var b [19]int
	err := row.Scan(
		&out.ID, &out.EmpresaID,
		&b[0], &b[1], &b[2], &b[3], &b[4], &b[5], &b[6], &b[7], &b[8], &b[9],
		&b[10], &b[11], &b[12], &b[13], &b[14], &b[15], &b[16], &b[17], &b[18],
		&out.FormatoImpresion,
		&out.FechaCreacion, &out.FechaActualizacion, &out.UsuarioCreador, &out.Estado, &out.Observaciones,
	)
	if err == sql.ErrNoRows {
		cfg := DefaultEmpresaCorteCajaConfiguracion(empresaID)
		id, upsertErr := UpsertEmpresaCorteCajaConfiguracion(dbConn, cfg)
		if upsertErr != nil {
			return nil, upsertErr
		}
		cfg.ID = id
		return &cfg, nil
	}
	if err != nil {
		return nil, err
	}
	out.MostrarResumen = b[0] > 0
	out.MostrarNumeroFacturas = b[1] > 0
	out.MostrarTotalVentas = b[2] > 0
	out.MostrarEfectivo = b[3] > 0
	out.MostrarDebito = b[4] > 0
	out.MostrarCredito = b[5] > 0
	out.MostrarTransferencias = b[6] > 0
	out.MostrarOtrosMedios = b[7] > 0
	out.MostrarIngresos = b[8] > 0
	out.MostrarEgresos = b[9] > 0
	out.MostrarAnulaciones = b[10] > 0
	out.MostrarDevoluciones = b[11] > 0
	out.MostrarCajaEsperada = b[12] > 0
	out.MostrarDiferenciaCaja = b[13] > 0
	out.MostrarVentasDetalle = b[14] > 0
	out.MostrarMovimientos = b[15] > 0
	out.MostrarItems = b[16] > 0
	out.MostrarSensoresPuertas = b[17] > 0
	out.MostrarAuditoria = b[18] > 0
	out = normalizeEmpresaCorteCajaConfiguracion(out)
	return &out, nil
}

func UpsertEmpresaCorteCajaConfiguracion(dbConn *sql.DB, cfg EmpresaCorteCajaConfiguracion) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}
	if cfg.EmpresaID <= 0 {
		return 0, errors.New("empresa_id invalido")
	}
	if err := EnsureEmpresaCorteCajaConfiguracionSchema(dbConn); err != nil {
		return 0, err
	}
	cfg = normalizeEmpresaCorteCajaConfiguracion(cfg)

	var existingID int64
	err := queryRowSQLCompat(dbConn, "SELECT id FROM empresa_corte_caja_configuracion WHERE empresa_id = ? LIMIT 1", cfg.EmpresaID).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	args := []interface{}{
		cfg.EmpresaID,
		corteCajaConfigBoolToInt(cfg.MostrarResumen),
		corteCajaConfigBoolToInt(cfg.MostrarNumeroFacturas),
		corteCajaConfigBoolToInt(cfg.MostrarTotalVentas),
		corteCajaConfigBoolToInt(cfg.MostrarEfectivo),
		corteCajaConfigBoolToInt(cfg.MostrarDebito),
		corteCajaConfigBoolToInt(cfg.MostrarCredito),
		corteCajaConfigBoolToInt(cfg.MostrarTransferencias),
		corteCajaConfigBoolToInt(cfg.MostrarOtrosMedios),
		corteCajaConfigBoolToInt(cfg.MostrarIngresos),
		corteCajaConfigBoolToInt(cfg.MostrarEgresos),
		corteCajaConfigBoolToInt(cfg.MostrarAnulaciones),
		corteCajaConfigBoolToInt(cfg.MostrarDevoluciones),
		corteCajaConfigBoolToInt(cfg.MostrarCajaEsperada),
		corteCajaConfigBoolToInt(cfg.MostrarDiferenciaCaja),
		corteCajaConfigBoolToInt(cfg.MostrarVentasDetalle),
		corteCajaConfigBoolToInt(cfg.MostrarMovimientos),
		corteCajaConfigBoolToInt(cfg.MostrarItems),
		corteCajaConfigBoolToInt(cfg.MostrarSensoresPuertas),
		corteCajaConfigBoolToInt(cfg.MostrarAuditoria),
		cfg.FormatoImpresion,
		cfg.UsuarioCreador,
		cfg.Estado,
		cfg.Observaciones,
	}
	if err == sql.ErrNoRows {
		return insertSQLCompat(dbConn, `INSERT INTO empresa_corte_caja_configuracion (
			empresa_id,
			mostrar_resumen, mostrar_numero_facturas, mostrar_total_ventas,
			mostrar_efectivo, mostrar_debito, mostrar_credito, mostrar_transferencias, mostrar_otros_medios,
			mostrar_ingresos, mostrar_egresos, mostrar_anulaciones, mostrar_devoluciones,
			mostrar_caja_esperada, mostrar_diferencia_caja,
			mostrar_ventas_detalle, mostrar_movimientos, mostrar_items, mostrar_sensores_puertas, mostrar_auditoria,
			formato_impresion, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			datetime('now','localtime'), datetime('now','localtime'), ?, ?, ?
		)`, args...)
	}

	updateArgs := append(args[1:], cfg.EmpresaID)
	_, updateErr := execSQLCompat(dbConn, `UPDATE empresa_corte_caja_configuracion SET
		mostrar_resumen = ?,
		mostrar_numero_facturas = ?,
		mostrar_total_ventas = ?,
		mostrar_efectivo = ?,
		mostrar_debito = ?,
		mostrar_credito = ?,
		mostrar_transferencias = ?,
		mostrar_otros_medios = ?,
		mostrar_ingresos = ?,
		mostrar_egresos = ?,
		mostrar_anulaciones = ?,
		mostrar_devoluciones = ?,
		mostrar_caja_esperada = ?,
		mostrar_diferencia_caja = ?,
		mostrar_ventas_detalle = ?,
		mostrar_movimientos = ?,
		mostrar_items = ?,
		mostrar_sensores_puertas = ?,
		mostrar_auditoria = ?,
		formato_impresion = ?,
		fecha_actualizacion = datetime('now','localtime'),
		usuario_creador = ?,
		estado = ?,
		observaciones = ?
	WHERE empresa_id = ?`, updateArgs...)
	if updateErr != nil {
		return 0, updateErr
	}
	return existingID, nil
}

func EmpresaCorteCajaReportesDesdeConfiguracion(cfg *EmpresaCorteCajaConfiguracion) []string {
	if cfg == nil {
		defaultCfg := DefaultEmpresaCorteCajaConfiguracion(0)
		cfg = &defaultCfg
	}
	out := []string{}
	if cfg.MostrarResumen {
		out = append(out, "resumen")
	}
	if cfg.MostrarMovimientos || cfg.MostrarIngresos || cfg.MostrarEgresos {
		out = append(out, "movimientos")
	}
	if cfg.MostrarVentasDetalle || cfg.MostrarNumeroFacturas || cfg.MostrarTotalVentas || cfg.MostrarEfectivo || cfg.MostrarDebito || cfg.MostrarCredito {
		out = append(out, "ventas")
	}
	if cfg.MostrarAnulaciones || cfg.MostrarDevoluciones {
		out = append(out, "anulaciones")
	}
	if cfg.MostrarItems {
		out = append(out, "items")
	}
	if cfg.MostrarSensoresPuertas {
		out = append(out, "sensores")
	}
	if cfg.MostrarAuditoria {
		out = append(out, "auditoria")
	}
	if len(out) == 0 {
		return []string{"resumen"}
	}
	return out
}

func normalizeEmpresaCorteCajaConfiguracion(cfg EmpresaCorteCajaConfiguracion) EmpresaCorteCajaConfiguracion {
	cfg.FormatoImpresion = strings.ToLower(strings.TrimSpace(cfg.FormatoImpresion))
	switch cfg.FormatoImpresion {
	case "pos", "ticket", "ejecutivo", "compacto":
		if cfg.FormatoImpresion == "ticket" {
			cfg.FormatoImpresion = "pos"
		}
		if cfg.FormatoImpresion == "compacto" {
			cfg.FormatoImpresion = "ejecutivo"
		}
	default:
		cfg.FormatoImpresion = "carta"
	}
	cfg.Estado = strings.TrimSpace(strings.ToLower(cfg.Estado))
	if cfg.Estado == "" {
		cfg.Estado = "activo"
	}
	if cfg.Estado != "inactivo" {
		cfg.Estado = "activo"
	}
	cfg.Observaciones = strings.TrimSpace(cfg.Observaciones)
	if len(cfg.Observaciones) > 800 {
		cfg.Observaciones = cfg.Observaciones[:800]
	}
	cfg.UsuarioCreador = strings.TrimSpace(cfg.UsuarioCreador)
	return cfg
}

func corteCajaConfigBoolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
