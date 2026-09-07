package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// EmpresaImpuestoConfig representa un impuesto configurable por empresa.
// Nota: es un catálogo operativo (habilitar/deshabilitar + tasa por defecto). No reemplaza el cálculo por item/factura.
type EmpresaImpuestoConfig struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	PaisCodigo         string  `json:"pais_codigo"`
	Codigo             string  `json:"codigo"`
	Nombre             string  `json:"nombre"`
	Tipo               string  `json:"tipo"` // impuesto | retencion
	TasaPorcentaje     float64 `json:"tasa_porcentaje"`
	Habilitado         int     `json:"habilitado"`
	AplicaEn           string  `json:"aplica_en,omitempty"` // ventas | compras | ambos
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

type EmpresaImpuestosResumen struct {
	Desde                 string  `json:"desde"`
	Hasta                 string  `json:"hasta"`
	VentasCerradas        int64   `json:"ventas_cerradas"`
	VentasGravadas        int64   `json:"ventas_gravadas"`
	BaseGravable          float64 `json:"base_gravable"`
	ImpuestoGenerado      float64 `json:"impuesto_generado"`
	TotalFacturado        float64 `json:"total_facturado"`
	TasaEfectiva          float64 `json:"tasa_efectiva"`
	ImpuestosActivos      int     `json:"impuestos_activos"`
	RetencionesActivas    int     `json:"retenciones_activas"`
	SaldoFiscalEstimado   float64 `json:"saldo_fiscal_estimado"`
	PromedioImpuestoVenta float64 `json:"promedio_impuesto_venta"`
}

type EmpresaImpuestoReporteCodigo struct {
	Codigo           string  `json:"codigo"`
	Nombre           string  `json:"nombre"`
	Tipo             string  `json:"tipo"`
	Lineas           int64   `json:"lineas"`
	BaseGravable     float64 `json:"base_gravable"`
	ImpuestoGenerado float64 `json:"impuesto_generado"`
	TasaConfigurada  float64 `json:"tasa_configurada"`
	TasaEfectiva     float64 `json:"tasa_efectiva"`
	Participacion    float64 `json:"participacion"`
	AplicaEn         string  `json:"aplica_en"`
	Habilitado       int     `json:"habilitado"`
}

type EmpresaImpuestoReporteDiario struct {
	Fecha            string  `json:"fecha"`
	Ventas           int64   `json:"ventas"`
	BaseGravable     float64 `json:"base_gravable"`
	ImpuestoGenerado float64 `json:"impuesto_generado"`
	TasaEfectiva     float64 `json:"tasa_efectiva"`
	TotalFacturado   float64 `json:"total_facturado"`
}

type EmpresaImpuestoAsientoSugerido struct {
	Orden          int     `json:"orden"`
	CuentaSugerida string  `json:"cuenta_sugerida"`
	Concepto       string  `json:"concepto"`
	Naturaleza     string  `json:"naturaleza"`
	Monto          float64 `json:"monto"`
	Tipo           string  `json:"tipo"`
}

type EmpresaImpuestosDashboard struct {
	EmpresaID  int64                            `json:"empresa_id"`
	GeneradoEn string                           `json:"generado_en"`
	Resumen    EmpresaImpuestosResumen          `json:"resumen"`
	PorCodigo  []EmpresaImpuestoReporteCodigo   `json:"por_codigo"`
	Diario     []EmpresaImpuestoReporteDiario   `json:"diario"`
	Asientos   []EmpresaImpuestoAsientoSugerido `json:"asientos"`
}

func normalizeImpuestoTipo(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v != "retencion" {
		return "impuesto"
	}
	return v
}

func normalizeAplicaEn(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "ventas", "compras", "ambos":
		return v
	default:
		return "ventas"
	}
}

func EnsureEmpresaImpuestosSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("db nil")
	}
	query := `
	CREATE TABLE IF NOT EXISTS empresa_impuestos_config (
		id BIGSERIAL PRIMARY KEY,
		empresa_id INTEGER NOT NULL,
		pais_codigo TEXT DEFAULT 'CO',
		codigo TEXT NOT NULL,
		nombre TEXT NOT NULL,
		tipo TEXT DEFAULT 'impuesto',
		tasa_porcentaje REAL DEFAULT 0,
		habilitado INTEGER DEFAULT 1,
		aplica_en TEXT DEFAULT 'ventas',
		fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
		fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT,
		UNIQUE(empresa_id, codigo)
	);`
	if shouldUsePostgresCompat(dbConn) {
		query = `
		CREATE TABLE IF NOT EXISTS empresa_impuestos_config (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			pais_codigo TEXT DEFAULT 'CO',
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			tipo TEXT DEFAULT 'impuesto',
			tasa_porcentaje DOUBLE PRECISION DEFAULT 0,
			habilitado INTEGER DEFAULT 1,
			aplica_en TEXT DEFAULT 'ventas',
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`
	}
	if _, err := execSQLCompat(dbConn, query); err != nil {
		return err
	}
	_ = ensureColumnIfMissing(dbConn, "empresa_impuestos_config", "pais_codigo", "TEXT DEFAULT 'CO'")
	_ = ensureColumnIfMissing(dbConn, "empresa_impuestos_config", "tipo", "TEXT DEFAULT 'impuesto'")
	_ = ensureColumnIfMissing(dbConn, "empresa_impuestos_config", "tasa_porcentaje", "DOUBLE PRECISION DEFAULT 0")
	_ = ensureColumnIfMissing(dbConn, "empresa_impuestos_config", "habilitado", "INTEGER DEFAULT 1")
	_ = ensureColumnIfMissing(dbConn, "empresa_impuestos_config", "aplica_en", "TEXT DEFAULT 'ventas'")
	_ = ensureColumnIfMissing(dbConn, "empresa_impuestos_config", "fecha_actualizacion", "TEXT")
	return nil
}

func ListEmpresaImpuestos(dbConn *sql.DB, empresaID int64) ([]EmpresaImpuestoConfig, error) {
	if err := EnsureEmpresaImpuestosSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		id, empresa_id,
		COALESCE(pais_codigo,'CO'),
		COALESCE(codigo,''),
		COALESCE(nombre,''),
		COALESCE(tipo,'impuesto'),
		COALESCE(tasa_porcentaje,0),
		COALESCE(habilitado,1),
		COALESCE(aplica_en,'ventas'),
		COALESCE(fecha_creacion,''),
		COALESCE(fecha_actualizacion,''),
		COALESCE(usuario_creador,''),
		COALESCE(estado,'activo'),
		COALESCE(observaciones,'')
	FROM empresa_impuestos_config
	WHERE empresa_id = ?
	ORDER BY tipo ASC, codigo ASC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EmpresaImpuestoConfig, 0)
	for rows.Next() {
		var it EmpresaImpuestoConfig
		if err := rows.Scan(
			&it.ID, &it.EmpresaID, &it.PaisCodigo, &it.Codigo, &it.Nombre, &it.Tipo, &it.TasaPorcentaje,
			&it.Habilitado, &it.AplicaEn, &it.FechaCreacion, &it.FechaActualizacion, &it.UsuarioCreador, &it.Estado, &it.Observaciones,
		); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, nil
}

func UpsertEmpresaImpuesto(dbConn *sql.DB, payload EmpresaImpuestoConfig) (int64, error) {
	if err := EnsureEmpresaImpuestosSchema(dbConn); err != nil {
		return 0, err
	}
	payload.Codigo = strings.ToUpper(strings.TrimSpace(payload.Codigo))
	payload.Nombre = strings.TrimSpace(payload.Nombre)
	payload.PaisCodigo = strings.ToUpper(strings.TrimSpace(payload.PaisCodigo))
	payload.Tipo = normalizeImpuestoTipo(payload.Tipo)
	payload.AplicaEn = normalizeAplicaEn(payload.AplicaEn)
	if payload.Habilitado != 1 {
		payload.Habilitado = 0
	}
	if payload.EmpresaID <= 0 || payload.Codigo == "" || payload.Nombre == "" {
		return 0, fmt.Errorf("empresa_id, codigo y nombre son obligatorios")
	}

	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_impuestos_config (
		empresa_id, pais_codigo, codigo, nombre, tipo, tasa_porcentaje, habilitado, aplica_en,
		usuario_creador, estado, observaciones, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT (empresa_id, codigo) DO UPDATE SET
		pais_codigo = excluded.pais_codigo,
		nombre = excluded.nombre,
		tipo = excluded.tipo,
		tasa_porcentaje = excluded.tasa_porcentaje,
		habilitado = excluded.habilitado,
		aplica_en = excluded.aplica_en,
		observaciones = excluded.observaciones,
		fecha_actualizacion = CURRENT_TIMESTAMP
	RETURNING id`,
		payload.EmpresaID, payload.PaisCodigo, payload.Codigo, payload.Nombre, payload.Tipo, payload.TasaPorcentaje, payload.Habilitado, payload.AplicaEn,
		strings.TrimSpace(payload.UsuarioCreador), normalizeChatEstado(payload.Estado), strings.TrimSpace(payload.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func EmpresaImpuestosDashboardData(dbConn *sql.DB, empresaID int64, desde string, hasta string) (EmpresaImpuestosDashboard, error) {
	if err := EnsureEmpresaImpuestosSchema(dbConn); err != nil {
		return EmpresaImpuestosDashboard{}, err
	}
	desde = strings.TrimSpace(desde)
	hasta = strings.TrimSpace(hasta)
	if desde == "" || hasta == "" {
		now := time.Now()
		if hasta == "" {
			hasta = now.Format("2006-01-02")
		}
		if desde == "" {
			desde = now.AddDate(0, 0, -29).Format("2006-01-02")
		}
	}

	configs, err := ListEmpresaImpuestos(dbConn, empresaID)
	if err != nil {
		return EmpresaImpuestosDashboard{}, err
	}
	activeImpuestos := 0
	activeRetenciones := 0
	configByCode := map[string]EmpresaImpuestoConfig{}
	for _, cfg := range configs {
		code := strings.ToUpper(strings.TrimSpace(cfg.Codigo))
		if code != "" {
			configByCode[code] = cfg
		}
		if cfg.Habilitado != 1 {
			continue
		}
		if strings.ToLower(strings.TrimSpace(cfg.Tipo)) == "retencion" {
			activeRetenciones++
		} else {
			activeImpuestos++
		}
	}

	dateExpr := "substr(COALESCE(NULLIF(c.pagado_en,''), NULLIF(c.fecha_actualizacion,''), c.fecha_creacion), 1, 10)"
	baseWhere := " FROM carritos_compras c WHERE c.empresa_id = ? AND LOWER(COALESCE(c.estado_carrito,'')) = 'cerrado' AND " + dateExpr + " >= ? AND " + dateExpr + " <= ? "

	var ventasCerradas int64
	if err := dbConn.QueryRow("SELECT COUNT(1)"+baseWhere, empresaID, desde, hasta).Scan(&ventasCerradas); err != nil {
		return EmpresaImpuestosDashboard{}, err
	}

	itemJoinWhere := `
		FROM carrito_compra_items i
		INNER JOIN carritos_compras c ON c.id = i.carrito_id AND c.empresa_id = i.empresa_id
		WHERE c.empresa_id = ?
		  AND LOWER(COALESCE(c.estado_carrito,'')) = 'cerrado'
		  AND LOWER(COALESCE(i.estado,'activo')) <> 'anulado'
		  AND ` + dateExpr + ` >= ?
		  AND ` + dateExpr + ` <= ?`

	var baseGravable float64
	var impuestoGenerado float64
	var totalFacturado float64
	if err := dbConn.QueryRow(`SELECT
		COALESCE(SUM(COALESCE(i.base_gravable,0)),0),
		COALESCE(SUM(COALESCE(i.valor_impuesto,0)),0),
		COALESCE(SUM(COALESCE(i.total_linea,0)),0)
	`+itemJoinWhere, empresaID, desde, hasta).Scan(&baseGravable, &impuestoGenerado, &totalFacturado); err != nil {
		return EmpresaImpuestosDashboard{}, err
	}

	var ventasGravadas int64
	if err := dbConn.QueryRow(`SELECT COUNT(DISTINCT c.id)
	`+itemJoinWhere+` AND COALESCE(i.valor_impuesto,0) > 0`, empresaID, desde, hasta).Scan(&ventasGravadas); err != nil {
		return EmpresaImpuestosDashboard{}, err
	}

	tasaEfectiva := 0.0
	if baseGravable > 0 {
		tasaEfectiva = (impuestoGenerado / baseGravable) * 100
	}
	promedioImpuestoVenta := 0.0
	if ventasCerradas > 0 {
		promedioImpuestoVenta = impuestoGenerado / float64(ventasCerradas)
	}

	rowsCodigo, err := querySQLCompat(dbConn, `SELECT
		COALESCE(NULLIF(TRIM(i.impuesto_codigo),''),'SIN_CODIGO') AS codigo,
		COUNT(1) AS lineas,
		COALESCE(SUM(COALESCE(i.base_gravable,0)),0) AS base_gravable,
		COALESCE(SUM(COALESCE(i.valor_impuesto,0)),0) AS impuesto_generado,
		COALESCE(AVG(COALESCE(i.impuesto_porcentaje,0)),0) AS tasa_configurada,
		CASE
			WHEN COALESCE(SUM(COALESCE(i.base_gravable,0)),0) > 0
				THEN (COALESCE(SUM(COALESCE(i.valor_impuesto,0)),0) * 100.0) / COALESCE(SUM(COALESCE(i.base_gravable,0)),0)
			ELSE 0
		END AS tasa_efectiva
	`+itemJoinWhere+`
		GROUP BY COALESCE(NULLIF(TRIM(i.impuesto_codigo),''),'SIN_CODIGO')
		ORDER BY impuesto_generado DESC, base_gravable DESC`, empresaID, desde, hasta)
	if err != nil {
		return EmpresaImpuestosDashboard{}, err
	}
	defer rowsCodigo.Close()

	porCodigo := make([]EmpresaImpuestoReporteCodigo, 0)
	for rowsCodigo.Next() {
		var row EmpresaImpuestoReporteCodigo
		if err := rowsCodigo.Scan(&row.Codigo, &row.Lineas, &row.BaseGravable, &row.ImpuestoGenerado, &row.TasaConfigurada, &row.TasaEfectiva); err != nil {
			return EmpresaImpuestosDashboard{}, err
		}
		cfg, ok := configByCode[strings.ToUpper(strings.TrimSpace(row.Codigo))]
		if ok {
			row.Nombre = cfg.Nombre
			row.Tipo = cfg.Tipo
			row.AplicaEn = cfg.AplicaEn
			row.Habilitado = cfg.Habilitado
			if row.TasaConfigurada == 0 && cfg.TasaPorcentaje > 0 {
				row.TasaConfigurada = cfg.TasaPorcentaje
			}
		} else {
			row.Nombre = row.Codigo
			row.Tipo = "impuesto"
			row.AplicaEn = "ventas"
			row.Habilitado = 1
		}
		if impuestoGenerado > 0 {
			row.Participacion = (row.ImpuestoGenerado / impuestoGenerado) * 100
		}
		porCodigo = append(porCodigo, row)
	}

	rowsDiario, err := querySQLCompat(dbConn, `SELECT
		`+dateExpr+` AS fecha,
		COUNT(DISTINCT c.id) AS ventas,
		COALESCE(SUM(COALESCE(i.base_gravable,0)),0) AS base_gravable,
		COALESCE(SUM(COALESCE(i.valor_impuesto,0)),0) AS impuesto_generado,
		COALESCE(SUM(COALESCE(i.total_linea,0)),0) AS total_facturado
	`+itemJoinWhere+`
		GROUP BY `+dateExpr+`
		ORDER BY fecha DESC`, empresaID, desde, hasta)
	if err != nil {
		return EmpresaImpuestosDashboard{}, err
	}
	defer rowsDiario.Close()

	diario := make([]EmpresaImpuestoReporteDiario, 0)
	for rowsDiario.Next() {
		var row EmpresaImpuestoReporteDiario
		if err := rowsDiario.Scan(&row.Fecha, &row.Ventas, &row.BaseGravable, &row.ImpuestoGenerado, &row.TotalFacturado); err != nil {
			return EmpresaImpuestosDashboard{}, err
		}
		if row.BaseGravable > 0 {
			row.TasaEfectiva = (row.ImpuestoGenerado / row.BaseGravable) * 100
		}
		diario = append(diario, row)
	}

	return EmpresaImpuestosDashboard{
		EmpresaID:  empresaID,
		GeneradoEn: time.Now().Format("2006-01-02 15:04:05"),
		Resumen: EmpresaImpuestosResumen{
			Desde:                 desde,
			Hasta:                 hasta,
			VentasCerradas:        ventasCerradas,
			VentasGravadas:        ventasGravadas,
			BaseGravable:          baseGravable,
			ImpuestoGenerado:      impuestoGenerado,
			TotalFacturado:        totalFacturado,
			TasaEfectiva:          tasaEfectiva,
			ImpuestosActivos:      activeImpuestos,
			RetencionesActivas:    activeRetenciones,
			SaldoFiscalEstimado:   impuestoGenerado,
			PromedioImpuestoVenta: promedioImpuestoVenta,
		},
		PorCodigo: porCodigo,
		Diario:    diario,
		Asientos:  buildImpuestosAsientosSugeridos(baseGravable, impuestoGenerado, totalFacturado, porCodigo),
	}, nil
}

func buildImpuestosAsientosSugeridos(baseGravable float64, impuestoGenerado float64, totalFacturado float64, porCodigo []EmpresaImpuestoReporteCodigo) []EmpresaImpuestoAsientoSugerido {
	asientos := make([]EmpresaImpuestoAsientoSugerido, 0, len(porCodigo)+2)
	if totalFacturado > 0 {
		asientos = append(asientos, EmpresaImpuestoAsientoSugerido{
			Orden:          1,
			CuentaSugerida: "1105 / 1305",
			Concepto:       "Caja, bancos o clientes por ventas del período",
			Naturaleza:     "debito",
			Monto:          totalFacturado,
			Tipo:           "activo",
		})
	}
	if baseGravable > 0 {
		asientos = append(asientos, EmpresaImpuestoAsientoSugerido{
			Orden:          len(asientos) + 1,
			CuentaSugerida: "4135",
			Concepto:       "Ingresos gravados del período",
			Naturaleza:     "credito",
			Monto:          baseGravable,
			Tipo:           "ingreso",
		})
	}
	for _, row := range porCodigo {
		if row.ImpuestoGenerado <= 0 {
			continue
		}
		asientos = append(asientos, EmpresaImpuestoAsientoSugerido{
			Orden:          len(asientos) + 1,
			CuentaSugerida: impuestosCuentaSugerida(row.Codigo),
			Concepto:       "Pasivo fiscal sugerido por " + strings.ToUpper(strings.TrimSpace(row.Codigo)),
			Naturaleza:     "credito",
			Monto:          row.ImpuestoGenerado,
			Tipo:           "pasivo",
		})
	}
	if len(asientos) == 0 && impuestoGenerado > 0 {
		asientos = append(asientos, EmpresaImpuestoAsientoSugerido{
			Orden:          1,
			CuentaSugerida: "2408",
			Concepto:       "Impuesto por pagar del período",
			Naturaleza:     "credito",
			Monto:          impuestoGenerado,
			Tipo:           "pasivo",
		})
	}
	return asientos
}

func impuestosCuentaSugerida(codigo string) string {
	code := strings.ToUpper(strings.TrimSpace(codigo))
	switch {
	case strings.Contains(code, "IVA"), strings.Contains(code, "ITBMS"):
		return "2408"
	case strings.Contains(code, "ICA"):
		return "2368"
	case strings.Contains(code, "RET"):
		return "2365"
	default:
		return "2408"
	}
}
