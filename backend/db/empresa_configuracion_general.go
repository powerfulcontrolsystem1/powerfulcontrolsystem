package db

import (
	"database/sql"
	"errors"
	"strings"
)

type EmpresaConfiguracionGeneral struct {
	ID                           int64  `json:"id"`
	EmpresaID                    int64  `json:"empresa_id"`
	ImprimirOrdenServicio        bool   `json:"imprimir_orden_servicio"`
	AreaDespacho                 string `json:"area_despacho,omitempty"`
	CopiasOrdenServicio          int64  `json:"copias_orden_servicio,omitempty"`
	NotaOrdenServicio            string `json:"nota_orden_servicio,omitempty"`
	DescuentosHabilitados        bool   `json:"descuentos_habilitados"`
	PermitirDescuentoPorcentaje  bool   `json:"permitir_descuento_porcentaje"`
	PermitirDescuentoCodigo      bool   `json:"permitir_descuento_codigo"`
	PermitirDescuentoValor       bool   `json:"permitir_descuento_valor"`
	CodigosDescuento             string `json:"codigos_descuento,omitempty"`
	LectorCodigoBarrasHabilitado bool   `json:"lector_codigo_barras_habilitado"`
	LectorCodigoBarrasAutofoco   bool   `json:"lector_codigo_barras_autofoco"`
	LectorCodigoBarrasAcumular   bool   `json:"lector_codigo_barras_acumular"`
	FechaCreacion                string `json:"fecha_creacion,omitempty"`
	FechaActualizacion           string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador               string `json:"usuario_creador,omitempty"`
	Estado                       string `json:"estado,omitempty"`
	Observaciones                string `json:"observaciones,omitempty"`
}

func EnsureEmpresaConfiguracionGeneralSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_configuracion_general (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL UNIQUE,
			imprimir_orden_servicio INTEGER DEFAULT 0,
			area_despacho TEXT,
			copias_orden_servicio INTEGER DEFAULT 1,
			nota_orden_servicio TEXT,
			descuentos_habilitados INTEGER DEFAULT 1,
			permitir_descuento_porcentaje INTEGER DEFAULT 1,
			permitir_descuento_codigo INTEGER DEFAULT 1,
			permitir_descuento_valor INTEGER DEFAULT 1,
			codigos_descuento TEXT,
			lector_codigo_barras_habilitado INTEGER DEFAULT 1,
			lector_codigo_barras_autofoco INTEGER DEFAULT 1,
			lector_codigo_barras_acumular INTEGER DEFAULT 1,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_configuracion_general_empresa ON empresa_configuracion_general(empresa_id);`,
	}

	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "imprimir_orden_servicio", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "area_despacho", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "copias_orden_servicio", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "nota_orden_servicio", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "descuentos_habilitados", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "permitir_descuento_porcentaje", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "permitir_descuento_codigo", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "permitir_descuento_valor", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "codigos_descuento", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "lector_codigo_barras_habilitado", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "lector_codigo_barras_autofoco", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "lector_codigo_barras_acumular", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

func GetEmpresaConfiguracionGeneral(dbConn *sql.DB, empresaID int64) (*EmpresaConfiguracionGeneral, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, errors.New("empresa_id invalido")
	}
	if err := EnsureEmpresaConfiguracionGeneralSchema(dbConn); err != nil {
		return nil, err
	}

	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(imprimir_orden_servicio, 0),
		COALESCE(area_despacho, ''),
		COALESCE(copias_orden_servicio, 1),
		COALESCE(nota_orden_servicio, ''),
		COALESCE(descuentos_habilitados, 1),
		COALESCE(permitir_descuento_porcentaje, 1),
		COALESCE(permitir_descuento_codigo, 1),
		COALESCE(permitir_descuento_valor, 1),
		COALESCE(codigos_descuento, ''),
		COALESCE(lector_codigo_barras_habilitado, 1),
		COALESCE(lector_codigo_barras_autofoco, 1),
		COALESCE(lector_codigo_barras_acumular, 1),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_configuracion_general
	WHERE empresa_id = ?
	LIMIT 1`, empresaID)

	var out EmpresaConfiguracionGeneral
	var imprimirOrdenServicio int
	var descuentosHabilitados int
	var permitirDescuentoPorcentaje int
	var permitirDescuentoCodigo int
	var permitirDescuentoValor int
	var lectorCodigoBarrasHabilitado int
	var lectorCodigoBarrasAutofoco int
	var lectorCodigoBarrasAcumular int
	err := row.Scan(
		&out.ID,
		&out.EmpresaID,
		&imprimirOrdenServicio,
		&out.AreaDespacho,
		&out.CopiasOrdenServicio,
		&out.NotaOrdenServicio,
		&descuentosHabilitados,
		&permitirDescuentoPorcentaje,
		&permitirDescuentoCodigo,
		&permitirDescuentoValor,
		&out.CodigosDescuento,
		&lectorCodigoBarrasHabilitado,
		&lectorCodigoBarrasAutofoco,
		&lectorCodigoBarrasAcumular,
		&out.FechaCreacion,
		&out.FechaActualizacion,
		&out.UsuarioCreador,
		&out.Estado,
		&out.Observaciones,
	)
	if err == sql.ErrNoRows {
		defaultCfg := defaultEmpresaConfiguracionGeneral(empresaID)
		id, upsertErr := UpsertEmpresaConfiguracionGeneral(dbConn, defaultCfg)
		if upsertErr != nil {
			return nil, upsertErr
		}
		defaultCfg.ID = id
		return &defaultCfg, nil
	}
	if err != nil {
		return nil, err
	}

	out.ImprimirOrdenServicio = imprimirOrdenServicio > 0
	out.DescuentosHabilitados = descuentosHabilitados > 0
	out.PermitirDescuentoPorcentaje = permitirDescuentoPorcentaje > 0
	out.PermitirDescuentoCodigo = permitirDescuentoCodigo > 0
	out.PermitirDescuentoValor = permitirDescuentoValor > 0
	out.LectorCodigoBarrasHabilitado = lectorCodigoBarrasHabilitado > 0
	out.LectorCodigoBarrasAutofoco = lectorCodigoBarrasAutofoco > 0
	out.LectorCodigoBarrasAcumular = lectorCodigoBarrasAcumular > 0
	out = normalizeEmpresaConfiguracionGeneral(out)

	return &out, nil
}

func UpsertEmpresaConfiguracionGeneral(dbConn *sql.DB, cfg EmpresaConfiguracionGeneral) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}
	if cfg.EmpresaID <= 0 {
		return 0, errors.New("empresa_id invalido")
	}
	if err := EnsureEmpresaConfiguracionGeneralSchema(dbConn); err != nil {
		return 0, err
	}

	cfg = normalizeEmpresaConfiguracionGeneral(cfg)

	var existingID int64
	err := dbConn.QueryRow("SELECT id FROM empresa_configuracion_general WHERE empresa_id = ? LIMIT 1", cfg.EmpresaID).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	if err == sql.ErrNoRows {
		insertedID, insertErr := insertSQLCompat(dbConn, `INSERT INTO empresa_configuracion_general (
			empresa_id,
			imprimir_orden_servicio,
			area_despacho,
			copias_orden_servicio,
			nota_orden_servicio,
			descuentos_habilitados,
			permitir_descuento_porcentaje,
			permitir_descuento_codigo,
			permitir_descuento_valor,
			codigos_descuento,
			lector_codigo_barras_habilitado,
			lector_codigo_barras_autofoco,
			lector_codigo_barras_acumular,
			fecha_creacion,
			fecha_actualizacion,
			usuario_creador,
			estado,
			observaciones
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			datetime('now','localtime'),
			datetime('now','localtime'),
			?, ?, ?
		)`,
			cfg.EmpresaID,
			configGeneralBoolToInt(cfg.ImprimirOrdenServicio),
			cfg.AreaDespacho,
			cfg.CopiasOrdenServicio,
			cfg.NotaOrdenServicio,
			configGeneralBoolToInt(cfg.DescuentosHabilitados),
			configGeneralBoolToInt(cfg.PermitirDescuentoPorcentaje),
			configGeneralBoolToInt(cfg.PermitirDescuentoCodigo),
			configGeneralBoolToInt(cfg.PermitirDescuentoValor),
			cfg.CodigosDescuento,
			configGeneralBoolToInt(cfg.LectorCodigoBarrasHabilitado),
			configGeneralBoolToInt(cfg.LectorCodigoBarrasAutofoco),
			configGeneralBoolToInt(cfg.LectorCodigoBarrasAcumular),
			cfg.UsuarioCreador,
			cfg.Estado,
			cfg.Observaciones,
		)
		if insertErr != nil {
			return 0, insertErr
		}
		return insertedID, nil
	}

	_, updateErr := dbConn.Exec(`UPDATE empresa_configuracion_general SET
		imprimir_orden_servicio = ?,
		area_despacho = ?,
		copias_orden_servicio = ?,
		nota_orden_servicio = ?,
		descuentos_habilitados = ?,
		permitir_descuento_porcentaje = ?,
		permitir_descuento_codigo = ?,
		permitir_descuento_valor = ?,
		codigos_descuento = ?,
		lector_codigo_barras_habilitado = ?,
		lector_codigo_barras_autofoco = ?,
		lector_codigo_barras_acumular = ?,
		fecha_actualizacion = datetime('now','localtime'),
		usuario_creador = ?,
		estado = ?,
		observaciones = ?
	WHERE empresa_id = ?`,
		configGeneralBoolToInt(cfg.ImprimirOrdenServicio),
		cfg.AreaDespacho,
		cfg.CopiasOrdenServicio,
		cfg.NotaOrdenServicio,
		configGeneralBoolToInt(cfg.DescuentosHabilitados),
		configGeneralBoolToInt(cfg.PermitirDescuentoPorcentaje),
		configGeneralBoolToInt(cfg.PermitirDescuentoCodigo),
		configGeneralBoolToInt(cfg.PermitirDescuentoValor),
		cfg.CodigosDescuento,
		configGeneralBoolToInt(cfg.LectorCodigoBarrasHabilitado),
		configGeneralBoolToInt(cfg.LectorCodigoBarrasAutofoco),
		configGeneralBoolToInt(cfg.LectorCodigoBarrasAcumular),
		cfg.UsuarioCreador,
		cfg.Estado,
		cfg.Observaciones,
		cfg.EmpresaID,
	)
	if updateErr != nil {
		return 0, updateErr
	}

	return existingID, nil
}

func defaultEmpresaConfiguracionGeneral(empresaID int64) EmpresaConfiguracionGeneral {
	return normalizeEmpresaConfiguracionGeneral(EmpresaConfiguracionGeneral{
		EmpresaID:                    empresaID,
		CopiasOrdenServicio:          1,
		DescuentosHabilitados:        true,
		PermitirDescuentoPorcentaje:  true,
		PermitirDescuentoCodigo:      true,
		PermitirDescuentoValor:       true,
		LectorCodigoBarrasHabilitado: true,
		LectorCodigoBarrasAutofoco:   true,
		LectorCodigoBarrasAcumular:   true,
		Estado:                       "activo",
	})
}

func normalizeEmpresaConfiguracionGeneral(cfg EmpresaConfiguracionGeneral) EmpresaConfiguracionGeneral {
	cfg.AreaDespacho = strings.TrimSpace(cfg.AreaDespacho)
	if cfg.CopiasOrdenServicio < 1 {
		cfg.CopiasOrdenServicio = 1
	}
	if cfg.CopiasOrdenServicio > 5 {
		cfg.CopiasOrdenServicio = 5
	}
	cfg.NotaOrdenServicio = strings.TrimSpace(cfg.NotaOrdenServicio)
	if len(cfg.NotaOrdenServicio) > 800 {
		cfg.NotaOrdenServicio = cfg.NotaOrdenServicio[:800]
	}
	cfg.CodigosDescuento = strings.TrimSpace(cfg.CodigosDescuento)
	if len(cfg.CodigosDescuento) > 2000 {
		cfg.CodigosDescuento = cfg.CodigosDescuento[:2000]
	}
	cfg.Estado = strings.TrimSpace(cfg.Estado)
	if cfg.Estado == "" {
		cfg.Estado = "activo"
	}
	cfg.Observaciones = strings.TrimSpace(cfg.Observaciones)
	return cfg
}

func configGeneralBoolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}