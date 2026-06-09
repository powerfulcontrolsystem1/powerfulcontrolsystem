package db

import (
	"database/sql"
	"errors"
	"strings"
)

type EmpresaConfiguracionGeneral struct {
	ID                                  int64   `json:"id"`
	EmpresaID                           int64   `json:"empresa_id"`
	ImprimirOrdenServicio               bool    `json:"imprimir_orden_servicio"`
	AreaDespacho                        string  `json:"area_despacho,omitempty"`
	CopiasOrdenServicio                 int64   `json:"copias_orden_servicio,omitempty"`
	NotaOrdenServicio                   string  `json:"nota_orden_servicio,omitempty"`
	DescuentosHabilitados               bool    `json:"descuentos_habilitados"`
	PermitirDescuentoPorcentaje         bool    `json:"permitir_descuento_porcentaje"`
	PermitirDescuentoCodigo             bool    `json:"permitir_descuento_codigo"`
	PermitirDescuentoValor              bool    `json:"permitir_descuento_valor"`
	CodigosDescuento                    string  `json:"codigos_descuento,omitempty"`
	LectorCodigoBarrasHabilitado        bool    `json:"lector_codigo_barras_habilitado"`
	LectorCodigoBarrasAutofoco          bool    `json:"lector_codigo_barras_autofoco"`
	LectorCodigoBarrasAcumular          bool    `json:"lector_codigo_barras_acumular"`
	CajaNombre                          string  `json:"caja_nombre,omitempty"`
	CajaCodigo                          string  `json:"caja_codigo,omitempty"`
	CajaActiva                          bool    `json:"caja_activa"`
	CajasSimultaneasHabilitadas         bool    `json:"cajas_simultaneas_habilitadas"`
	MaxCajasSimultaneasEmpresa          int64   `json:"max_cajas_simultaneas_empresa,omitempty"`
	CajonMonederoHabilitado             bool    `json:"cajon_monedero_habilitado"`
	AbrirCajonAlPagarCarrito            bool    `json:"abrir_cajon_al_pagar_carrito"`
	AbrirCajonAlCerrarTransaccion       bool    `json:"abrir_cajon_al_cerrar_transaccion"`
	CajonMonederoMetodo                 string  `json:"cajon_monedero_metodo,omitempty"`
	CajonMonederoImpresoraFuncionalidad string  `json:"cajon_monedero_impresora_funcionalidad,omitempty"`
	CajonMonederoComando                string  `json:"cajon_monedero_comando,omitempty"`
	CajaObservaciones                   string  `json:"caja_observaciones,omitempty"`
	ClimaCiudad                         string  `json:"clima_ciudad,omitempty"`
	ClimaRegion                         string  `json:"clima_region,omitempty"`
	ClimaPais                           string  `json:"clima_pais,omitempty"`
	ClimaPaisCodigo                     string  `json:"clima_pais_codigo,omitempty"`
	ClimaMoneda                         string  `json:"clima_moneda,omitempty"`
	ClimaLatitud                        float64 `json:"clima_latitud,omitempty"`
	ClimaLongitud                       float64 `json:"clima_longitud,omitempty"`
	ClimaNombre                         string  `json:"clima_nombre,omitempty"`
	ClimaFuente                         string  `json:"clima_fuente,omitempty"`
	FechaCreacion                       string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion                  string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador                      string  `json:"usuario_creador,omitempty"`
	Estado                              string  `json:"estado,omitempty"`
	Observaciones                       string  `json:"observaciones,omitempty"`
}

func EnsureEmpresaConfiguracionGeneralSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}

	idDefinition := "BIGSERIAL PRIMARY KEY"
	createdDefault := "CURRENT_TIMESTAMP"
	updatedDefault := "CURRENT_TIMESTAMP"
	if isPostgresDialect() {
		idDefinition = "BIGSERIAL PRIMARY KEY"
		createdDefault = "CAST(CURRENT_TIMESTAMP AS TEXT)"
		updatedDefault = "CAST(CURRENT_TIMESTAMP AS TEXT)"
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_configuracion_general (
			id ` + idDefinition + `,
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
			caja_nombre TEXT,
			caja_codigo TEXT,
			caja_activa INTEGER DEFAULT 1,
			cajas_simultaneas_habilitadas INTEGER DEFAULT 1,
			max_cajas_simultaneas_empresa INTEGER DEFAULT 0,
			cajon_monedero_habilitado INTEGER DEFAULT 0,
			abrir_cajon_al_pagar_carrito INTEGER DEFAULT 0,
			abrir_cajon_al_cerrar_transaccion INTEGER DEFAULT 0,
			cajon_monedero_metodo TEXT DEFAULT 'impresion_pos',
			cajon_monedero_impresora_funcionalidad TEXT DEFAULT 'cajon_monedero',
			cajon_monedero_comando TEXT DEFAULT 'escpos_pulse',
			caja_observaciones TEXT,
			clima_ciudad TEXT,
			clima_region TEXT,
			clima_pais TEXT,
			clima_pais_codigo TEXT,
			clima_moneda TEXT,
			clima_latitud REAL DEFAULT 0,
			clima_longitud REAL DEFAULT 0,
			clima_nombre TEXT,
			clima_fuente TEXT,
			fecha_creacion TEXT DEFAULT (` + createdDefault + `),
			fecha_actualizacion TEXT DEFAULT (` + updatedDefault + `),
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

	if err := ensurePostgresTableIDSequence(dbConn, "empresa_configuracion_general"); err != nil {
		return err
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
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "caja_nombre", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "caja_codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "caja_activa", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "cajas_simultaneas_habilitadas", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "max_cajas_simultaneas_empresa", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "cajon_monedero_habilitado", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "abrir_cajon_al_pagar_carrito", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "abrir_cajon_al_cerrar_transaccion", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "cajon_monedero_metodo", "TEXT DEFAULT 'impresion_pos'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "cajon_monedero_impresora_funcionalidad", "TEXT DEFAULT 'cajon_monedero'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "cajon_monedero_comando", "TEXT DEFAULT 'escpos_pulse'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "caja_observaciones", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "clima_ciudad", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "clima_region", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "clima_pais", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "clima_pais_codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "clima_moneda", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "clima_latitud", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "clima_longitud", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "clima_nombre", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "clima_fuente", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_general", "fecha_creacion", "TEXT"); err != nil {
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

	row := queryRowSQLCompat(dbConn, `SELECT
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
		COALESCE(caja_nombre, ''),
		COALESCE(caja_codigo, ''),
		COALESCE(caja_activa, 1),
		COALESCE(cajas_simultaneas_habilitadas, 1),
		COALESCE(max_cajas_simultaneas_empresa, 0),
		COALESCE(cajon_monedero_habilitado, 0),
		COALESCE(abrir_cajon_al_pagar_carrito, 0),
		COALESCE(abrir_cajon_al_cerrar_transaccion, 0),
		COALESCE(cajon_monedero_metodo, 'impresion_pos'),
		COALESCE(cajon_monedero_impresora_funcionalidad, 'cajon_monedero'),
		COALESCE(cajon_monedero_comando, 'escpos_pulse'),
		COALESCE(caja_observaciones, ''),
		COALESCE(clima_ciudad, ''),
		COALESCE(clima_region, ''),
		COALESCE(clima_pais, ''),
		COALESCE(clima_pais_codigo, ''),
		COALESCE(clima_moneda, ''),
		COALESCE(clima_latitud, 0),
		COALESCE(clima_longitud, 0),
		COALESCE(clima_nombre, ''),
		COALESCE(clima_fuente, ''),
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
	var cajaActiva int
	var cajasSimultaneasHabilitadas int
	var cajonMonederoHabilitado int
	var abrirCajonAlPagarCarrito int
	var abrirCajonAlCerrarTransaccion int
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
		&out.CajaNombre,
		&out.CajaCodigo,
		&cajaActiva,
		&cajasSimultaneasHabilitadas,
		&out.MaxCajasSimultaneasEmpresa,
		&cajonMonederoHabilitado,
		&abrirCajonAlPagarCarrito,
		&abrirCajonAlCerrarTransaccion,
		&out.CajonMonederoMetodo,
		&out.CajonMonederoImpresoraFuncionalidad,
		&out.CajonMonederoComando,
		&out.CajaObservaciones,
		&out.ClimaCiudad,
		&out.ClimaRegion,
		&out.ClimaPais,
		&out.ClimaPaisCodigo,
		&out.ClimaMoneda,
		&out.ClimaLatitud,
		&out.ClimaLongitud,
		&out.ClimaNombre,
		&out.ClimaFuente,
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
	out.CajaActiva = cajaActiva > 0
	out.CajasSimultaneasHabilitadas = cajasSimultaneasHabilitadas > 0
	out.CajonMonederoHabilitado = cajonMonederoHabilitado > 0
	out.AbrirCajonAlPagarCarrito = abrirCajonAlPagarCarrito > 0
	out.AbrirCajonAlCerrarTransaccion = abrirCajonAlCerrarTransaccion > 0
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
	err := queryRowSQLCompat(dbConn, "SELECT id FROM empresa_configuracion_general WHERE empresa_id = ? LIMIT 1", cfg.EmpresaID).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	if err == sql.ErrNoRows {
		nowExpr := sqlNowTextExpr()
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
			caja_nombre,
			caja_codigo,
			caja_activa,
			cajas_simultaneas_habilitadas,
			max_cajas_simultaneas_empresa,
			cajon_monedero_habilitado,
			abrir_cajon_al_pagar_carrito,
			abrir_cajon_al_cerrar_transaccion,
			cajon_monedero_metodo,
			cajon_monedero_impresora_funcionalidad,
			cajon_monedero_comando,
			caja_observaciones,
			clima_ciudad,
			clima_region,
			clima_pais,
			clima_pais_codigo,
			clima_moneda,
			clima_latitud,
			clima_longitud,
			clima_nombre,
			clima_fuente,
			fecha_creacion,
			fecha_actualizacion,
			usuario_creador,
			estado,
			observaciones
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			`+nowExpr+`,
			`+nowExpr+`,
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
			cfg.CajaNombre,
			cfg.CajaCodigo,
			configGeneralBoolToInt(cfg.CajaActiva),
			configGeneralBoolToInt(cfg.CajasSimultaneasHabilitadas),
			cfg.MaxCajasSimultaneasEmpresa,
			configGeneralBoolToInt(cfg.CajonMonederoHabilitado),
			configGeneralBoolToInt(cfg.AbrirCajonAlPagarCarrito),
			configGeneralBoolToInt(cfg.AbrirCajonAlCerrarTransaccion),
			cfg.CajonMonederoMetodo,
			cfg.CajonMonederoImpresoraFuncionalidad,
			cfg.CajonMonederoComando,
			cfg.CajaObservaciones,
			cfg.ClimaCiudad,
			cfg.ClimaRegion,
			cfg.ClimaPais,
			cfg.ClimaPaisCodigo,
			cfg.ClimaMoneda,
			cfg.ClimaLatitud,
			cfg.ClimaLongitud,
			cfg.ClimaNombre,
			cfg.ClimaFuente,
			cfg.UsuarioCreador,
			cfg.Estado,
			cfg.Observaciones,
		)
		if insertErr != nil {
			return 0, insertErr
		}
		return insertedID, nil
	}

	_, updateErr := execSQLCompat(dbConn, `UPDATE empresa_configuracion_general SET
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
		caja_nombre = ?,
		caja_codigo = ?,
		caja_activa = ?,
		cajas_simultaneas_habilitadas = ?,
		max_cajas_simultaneas_empresa = ?,
		cajon_monedero_habilitado = ?,
		abrir_cajon_al_pagar_carrito = ?,
		abrir_cajon_al_cerrar_transaccion = ?,
		cajon_monedero_metodo = ?,
		cajon_monedero_impresora_funcionalidad = ?,
		cajon_monedero_comando = ?,
		caja_observaciones = ?,
		clima_ciudad = ?,
		clima_region = ?,
		clima_pais = ?,
		clima_pais_codigo = ?,
		clima_moneda = ?,
		clima_latitud = ?,
		clima_longitud = ?,
		clima_nombre = ?,
		clima_fuente = ?,
		fecha_actualizacion = `+sqlNowTextExpr()+`,
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
		cfg.CajaNombre,
		cfg.CajaCodigo,
		configGeneralBoolToInt(cfg.CajaActiva),
		configGeneralBoolToInt(cfg.CajasSimultaneasHabilitadas),
		cfg.MaxCajasSimultaneasEmpresa,
		configGeneralBoolToInt(cfg.CajonMonederoHabilitado),
		configGeneralBoolToInt(cfg.AbrirCajonAlPagarCarrito),
		configGeneralBoolToInt(cfg.AbrirCajonAlCerrarTransaccion),
		cfg.CajonMonederoMetodo,
		cfg.CajonMonederoImpresoraFuncionalidad,
		cfg.CajonMonederoComando,
		cfg.CajaObservaciones,
		cfg.ClimaCiudad,
		cfg.ClimaRegion,
		cfg.ClimaPais,
		cfg.ClimaPaisCodigo,
		cfg.ClimaMoneda,
		cfg.ClimaLatitud,
		cfg.ClimaLongitud,
		cfg.ClimaNombre,
		cfg.ClimaFuente,
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
		EmpresaID:                           empresaID,
		CopiasOrdenServicio:                 1,
		DescuentosHabilitados:               true,
		PermitirDescuentoPorcentaje:         true,
		PermitirDescuentoCodigo:             true,
		PermitirDescuentoValor:              true,
		LectorCodigoBarrasHabilitado:        true,
		LectorCodigoBarrasAutofoco:          true,
		LectorCodigoBarrasAcumular:          true,
		CajaActiva:                          true,
		CajasSimultaneasHabilitadas:         true,
		CajonMonederoMetodo:                 "impresion_pos",
		CajonMonederoImpresoraFuncionalidad: "cajon_monedero",
		CajonMonederoComando:                "escpos_pulse",
		Estado:                              "activo",
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
	cfg.CajaNombre = strings.TrimSpace(cfg.CajaNombre)
	if len(cfg.CajaNombre) > 120 {
		cfg.CajaNombre = cfg.CajaNombre[:120]
	}
	cfg.CajaCodigo = strings.TrimSpace(cfg.CajaCodigo)
	if len(cfg.CajaCodigo) > 80 {
		cfg.CajaCodigo = cfg.CajaCodigo[:80]
	}
	if cfg.MaxCajasSimultaneasEmpresa < 0 {
		cfg.MaxCajasSimultaneasEmpresa = 0
	}
	if cfg.MaxCajasSimultaneasEmpresa > 99 {
		cfg.MaxCajasSimultaneasEmpresa = 99
	}
	cfg.CajonMonederoMetodo = strings.TrimSpace(strings.ToLower(cfg.CajonMonederoMetodo))
	switch cfg.CajonMonederoMetodo {
	case "impresion_pos", "manual":
	default:
		cfg.CajonMonederoMetodo = "impresion_pos"
	}
	cfg.CajonMonederoImpresoraFuncionalidad = strings.TrimSpace(cfg.CajonMonederoImpresoraFuncionalidad)
	if cfg.CajonMonederoImpresoraFuncionalidad == "" {
		cfg.CajonMonederoImpresoraFuncionalidad = "cajon_monedero"
	}
	if len(cfg.CajonMonederoImpresoraFuncionalidad) > 80 {
		cfg.CajonMonederoImpresoraFuncionalidad = cfg.CajonMonederoImpresoraFuncionalidad[:80]
	}
	cfg.CajonMonederoComando = strings.TrimSpace(strings.ToLower(cfg.CajonMonederoComando))
	switch cfg.CajonMonederoComando {
	case "escpos_pulse", "driver_auto_open":
	default:
		cfg.CajonMonederoComando = "escpos_pulse"
	}
	cfg.CajaObservaciones = strings.TrimSpace(cfg.CajaObservaciones)
	if len(cfg.CajaObservaciones) > 800 {
		cfg.CajaObservaciones = cfg.CajaObservaciones[:800]
	}
	cfg.ClimaCiudad = strings.TrimSpace(cfg.ClimaCiudad)
	if len(cfg.ClimaCiudad) > 120 {
		cfg.ClimaCiudad = cfg.ClimaCiudad[:120]
	}
	cfg.ClimaRegion = strings.TrimSpace(cfg.ClimaRegion)
	if len(cfg.ClimaRegion) > 120 {
		cfg.ClimaRegion = cfg.ClimaRegion[:120]
	}
	cfg.ClimaPais = strings.TrimSpace(cfg.ClimaPais)
	if len(cfg.ClimaPais) > 120 {
		cfg.ClimaPais = cfg.ClimaPais[:120]
	}
	cfg.ClimaPaisCodigo = strings.ToUpper(strings.TrimSpace(cfg.ClimaPaisCodigo))
	if len(cfg.ClimaPaisCodigo) > 2 {
		cfg.ClimaPaisCodigo = cfg.ClimaPaisCodigo[:2]
	}
	cfg.ClimaMoneda = strings.ToUpper(strings.TrimSpace(cfg.ClimaMoneda))
	if len(cfg.ClimaMoneda) > 3 {
		cfg.ClimaMoneda = cfg.ClimaMoneda[:3]
	}
	cfg.ClimaNombre = strings.TrimSpace(cfg.ClimaNombre)
	if len(cfg.ClimaNombre) > 240 {
		cfg.ClimaNombre = cfg.ClimaNombre[:240]
	}
	cfg.ClimaFuente = strings.ToLower(strings.TrimSpace(cfg.ClimaFuente))
	switch cfg.ClimaFuente {
	case "", "gps", "manual", "perfil", "navegador", "ip", "fallback":
	default:
		cfg.ClimaFuente = "manual"
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
