package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	ColombiaDefaultsVersion              = "CO-2026-06"
	ColombiaSalarioMinimoMensual2026     = 1750905
	ColombiaAuxilioTransporteMensual2026 = 249095
	colombiaDefaultsMigrationVersion     = "20260607_colombia_impuestos_nomina_defaults"
	colombiaDefaultsPrefClave            = "preconfiguracion_colombia_fiscal_nomina"
	colombiaDefaultsMigrationDescription = "Preconfigura impuestos Colombia y parametros base de nomina en empresas existentes"
	colombiaDefaultsUsuarioSistema       = "sistema.preconfiguracion_colombia"
	colombiaDefaultsObservacionPreprod   = "[preproduccion_2026-06-07] defaults Colombia impuestos y nomina; revisar con contador antes de produccion"
)

type EmpresaColombiaDefaultsResult struct {
	EmpresaID             int64    `json:"empresa_id"`
	Version               string   `json:"version"`
	Impuestos             int      `json:"impuestos"`
	NominaConfiguracionID int64    `json:"nomina_configuracion_id"`
	ConceptosNomina       int      `json:"conceptos_nomina"`
	BodegaID              int64    `json:"bodega_id"`
	MarkerID              int64    `json:"marker_id"`
	Errores               []string `json:"errores,omitempty"`
}

type EmpresasColombiaDefaultsBackfillResult struct {
	Version   string                          `json:"version"`
	Empresas  int                             `json:"empresas"`
	Aplicadas int                             `json:"aplicadas"`
	Errores   []string                        `json:"errores,omitempty"`
	Items     []EmpresaColombiaDefaultsResult `json:"items,omitempty"`
}

type CatalogoLegalVersion struct {
	ID             int64  `json:"id"`
	PaisCodigo     string `json:"pais_codigo"`
	Version        string `json:"version"`
	Nombre         string `json:"nombre"`
	VigenciaDesde  string `json:"vigencia_desde"`
	VigenciaHasta  string `json:"vigencia_hasta,omitempty"`
	Estado         string `json:"estado"`
	Fuente         string `json:"fuente,omitempty"`
	Descripcion    string `json:"descripcion,omitempty"`
	FechaCreacion  string `json:"fecha_creacion,omitempty"`
	ActualizadoPor string `json:"actualizado_por,omitempty"`
}

type CatalogoLegalParametro struct {
	ID             int64   `json:"id"`
	VersionID      int64   `json:"version_id"`
	Codigo         string  `json:"codigo"`
	Grupo          string  `json:"grupo"`
	Nombre         string  `json:"nombre"`
	TipoValor      string  `json:"tipo_valor"`
	ValorNumero    float64 `json:"valor_numero"`
	ValorTexto     string  `json:"valor_texto,omitempty"`
	Habilitado     int     `json:"habilitado"`
	Orden          int     `json:"orden"`
	Observaciones  string  `json:"observaciones,omitempty"`
	FechaCreacion  string  `json:"fecha_creacion,omitempty"`
	ActualizadoPor string  `json:"actualizado_por,omitempty"`
}

type EmpresaParametrosLegalesEstado struct {
	EmpresaID             int64                    `json:"empresa_id"`
	PaisCodigo            string                   `json:"pais_codigo"`
	VersionAplicada       string                   `json:"version_aplicada,omitempty"`
	VersionDisponible     string                   `json:"version_disponible,omitempty"`
	NombreDisponible      string                   `json:"nombre_disponible,omitempty"`
	VigenciaDesde         string                   `json:"vigencia_desde,omitempty"`
	AutoActualizar        bool                     `json:"auto_actualizar"`
	Pendiente             bool                     `json:"pendiente"`
	FechaUltimaRevision   string                   `json:"fecha_ultima_revision,omitempty"`
	FechaUltimaAplicacion string                   `json:"fecha_ultima_aplicacion,omitempty"`
	ModoUltimaAplicacion  string                   `json:"modo_ultima_aplicacion,omitempty"`
	Parametros            []CatalogoLegalParametro `json:"parametros,omitempty"`
	Mensaje               string                   `json:"mensaje,omitempty"`
}

type EmpresaParametrosLegalesApplyResult struct {
	EmpresaColombiaDefaultsResult
	VersionAplicada  string `json:"version_aplicada"`
	AutoActualizar   bool   `json:"auto_actualizar"`
	ModoAplicacion   string `json:"modo_aplicacion"`
	RegistroAplicado int64  `json:"registro_aplicado"`
}

type EmpresaParametrosLegalesWorkerResult struct {
	Version           string   `json:"version"`
	EmpresasRevisadas int      `json:"empresas_revisadas"`
	Pendientes        int      `json:"pendientes"`
	Aplicadas         int      `json:"aplicadas"`
	Errores           []string `json:"errores,omitempty"`
}

func EnsureCatalogoLegalPaisSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	if dbConn == nil {
		return fmt.Errorf("db nil")
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS catalogo_legal_pais_versiones (
			id BIGSERIAL PRIMARY KEY,
			pais_codigo TEXT NOT NULL,
			version TEXT NOT NULL,
			nombre TEXT NOT NULL,
			vigencia_desde TEXT,
			vigencia_hasta TEXT,
			estado TEXT DEFAULT 'vigente',
			fuente TEXT,
			descripcion TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			actualizado_por TEXT,
			UNIQUE(pais_codigo, version)
		);`,
		`CREATE TABLE IF NOT EXISTS catalogo_legal_pais_parametros (
			id BIGSERIAL PRIMARY KEY,
			version_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			grupo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			tipo_valor TEXT DEFAULT 'numero',
			valor_numero REAL DEFAULT 0,
			valor_texto TEXT,
			habilitado INTEGER DEFAULT 1,
			orden INTEGER DEFAULT 0,
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			actualizado_por TEXT,
			UNIQUE(version_id, codigo)
		);`,
		`CREATE TABLE IF NOT EXISTS empresa_parametros_legales_aplicados (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			pais_codigo TEXT NOT NULL DEFAULT 'CO',
			version TEXT NOT NULL,
			auto_actualizar INTEGER DEFAULT 0,
			fecha_ultima_revision TEXT,
			fecha_aplicacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			modo_aplicacion TEXT DEFAULT 'manual',
			estado TEXT DEFAULT 'activo',
			resumen_json TEXT,
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			UNIQUE(empresa_id, pais_codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_catalogo_legal_versiones_pais_estado ON catalogo_legal_pais_versiones(pais_codigo, estado, vigencia_desde);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_parametros_legales_empresa ON empresa_parametros_legales_aplicados(empresa_id, pais_codigo);`,
	}
	for _, stmt := range stmts {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	cols := []struct {
		table string
		name  string
		def   string
	}{
		{"catalogo_legal_pais_versiones", "fuente", "TEXT"},
		{"catalogo_legal_pais_versiones", "descripcion", "TEXT"},
		{"catalogo_legal_pais_versiones", "fecha_actualizacion", "TEXT"},
		{"catalogo_legal_pais_parametros", "valor_texto", "TEXT"},
		{"catalogo_legal_pais_parametros", "habilitado", "INTEGER DEFAULT 1"},
		{"catalogo_legal_pais_parametros", "orden", "INTEGER DEFAULT 0"},
		{"catalogo_legal_pais_parametros", "fecha_actualizacion", "TEXT"},
		{"empresa_parametros_legales_aplicados", "auto_actualizar", "INTEGER DEFAULT 0"},
		{"empresa_parametros_legales_aplicados", "fecha_ultima_revision", "TEXT"},
		{"empresa_parametros_legales_aplicados", "modo_aplicacion", "TEXT DEFAULT 'manual'"},
		{"empresa_parametros_legales_aplicados", "resumen_json", "TEXT"},
		{"empresa_parametros_legales_aplicados", "fecha_actualizacion", "TEXT"},
	}
	for _, col := range cols {
		if err := ensureColumnIfMissing(dbConn, col.table, col.name, col.def); err != nil {
			return err
		}
	}
	return nil
}

func SeedCatalogoLegalColombiaBase(dbConn *sql.DB, usuario string) error {
	if err := EnsureCatalogoLegalPaisSchema(dbConn); err != nil {
		return err
	}
	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = colombiaDefaultsUsuarioSistema
	}
	versionID, err := insertSQLCompat(dbConn, `INSERT INTO catalogo_legal_pais_versiones (
		pais_codigo, version, nombre, vigencia_desde, vigencia_hasta, estado, fuente, descripcion,
		fecha_creacion, fecha_actualizacion, actualizado_por
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?)
	ON CONFLICT(pais_codigo, version) DO UPDATE SET
		nombre = excluded.nombre,
		vigencia_desde = excluded.vigencia_desde,
		vigencia_hasta = excluded.vigencia_hasta,
		estado = excluded.estado,
		fuente = excluded.fuente,
		descripcion = excluded.descripcion,
		fecha_actualizacion = CURRENT_TIMESTAMP,
		actualizado_por = excluded.actualizado_por
	RETURNING id`,
		"CO",
		ColombiaDefaultsVersion,
		"Parametros legales Colombia 2026",
		"2026-01-01",
		"2026-12-31",
		"vigente",
		"Catalogo interno auditado con decretos nacionales y Estatuto Tributario",
		"Salario minimo, auxilio de transporte, impuestos y porcentajes base Colombia para preproduccion.",
		usuario,
	)
	if err != nil {
		return err
	}
	params := []CatalogoLegalParametro{
		{Codigo: "NOM_SALARIO_MINIMO", Grupo: "nomina", Nombre: "Salario minimo mensual", TipoValor: "moneda", ValorNumero: ColombiaSalarioMinimoMensual2026, Habilitado: 1, Orden: 10},
		{Codigo: "NOM_AUXILIO_TRANSPORTE", Grupo: "nomina", Nombre: "Auxilio transporte legal mensual", TipoValor: "moneda", ValorNumero: ColombiaAuxilioTransporteMensual2026, Habilitado: 1, Orden: 20},
		{Codigo: "NOM_HORAS_SEMANA", Grupo: "nomina", Nombre: "Horas ordinarias semana", TipoValor: "numero", ValorNumero: 44, Habilitado: 1, Orden: 30},
		{Codigo: "NOM_DIAS_MES", Grupo: "nomina", Nombre: "Dias base nomina mes", TipoValor: "numero", ValorNumero: 30, Habilitado: 1, Orden: 40},
		{Codigo: "NOM_SALUD_EMPLEADO", Grupo: "nomina", Nombre: "Deduccion salud empleado", TipoValor: "porcentaje", ValorNumero: 4, Habilitado: 1, Orden: 50},
		{Codigo: "NOM_PENSION_EMPLEADO", Grupo: "nomina", Nombre: "Deduccion pension empleado", TipoValor: "porcentaje", ValorNumero: 4, Habilitado: 1, Orden: 60},
		{Codigo: "IMP_IVA", Grupo: "impuestos", Nombre: "IVA tarifa general", TipoValor: "porcentaje", ValorNumero: 19, ValorTexto: "IVA", Habilitado: 1, Orden: 100},
		{Codigo: "IMP_IVA_0", Grupo: "impuestos", Nombre: "IVA 0% / exento / excluido", TipoValor: "porcentaje", ValorNumero: 0, ValorTexto: "IVA_0", Habilitado: 1, Orden: 110},
		{Codigo: "IMP_INC_8", Grupo: "impuestos", Nombre: "Impuesto nacional al consumo", TipoValor: "porcentaje", ValorNumero: 8, ValorTexto: "INC_8", Habilitado: 0, Orden: 120},
		{Codigo: "RET_RETEIVA", Grupo: "retenciones", Nombre: "Retencion a titulo de IVA", TipoValor: "porcentaje", ValorNumero: 15, ValorTexto: "RETEIVA", Habilitado: 0, Orden: 200},
	}
	for _, p := range params {
		if _, err := insertSQLCompat(dbConn, `INSERT INTO catalogo_legal_pais_parametros (
			version_id, codigo, grupo, nombre, tipo_valor, valor_numero, valor_texto, habilitado, orden,
			observaciones, fecha_creacion, fecha_actualizacion, actualizado_por
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?)
		ON CONFLICT(version_id, codigo) DO UPDATE SET
			grupo = excluded.grupo,
			nombre = excluded.nombre,
			tipo_valor = excluded.tipo_valor,
			valor_numero = excluded.valor_numero,
			valor_texto = excluded.valor_texto,
			habilitado = excluded.habilitado,
			orden = excluded.orden,
			observaciones = excluded.observaciones,
			fecha_actualizacion = CURRENT_TIMESTAMP,
			actualizado_por = excluded.actualizado_por
		RETURNING id`,
			versionID,
			strings.TrimSpace(p.Codigo),
			strings.TrimSpace(p.Grupo),
			strings.TrimSpace(p.Nombre),
			strings.TrimSpace(p.TipoValor),
			p.ValorNumero,
			strings.TrimSpace(p.ValorTexto),
			p.Habilitado,
			p.Orden,
			"Parametro legal versionado "+ColombiaDefaultsVersion,
			usuario,
		); err != nil {
			return err
		}
	}
	return nil
}

func GetLatestCatalogoLegalVersion(dbConn *sql.DB, paisCodigo string) (CatalogoLegalVersion, error) {
	if err := SeedCatalogoLegalColombiaBase(dbConn, colombiaDefaultsUsuarioSistema); err != nil {
		return CatalogoLegalVersion{}, err
	}
	paisCodigo = strings.ToUpper(strings.TrimSpace(paisCodigo))
	if paisCodigo == "" {
		paisCodigo = "CO"
	}
	row := queryRowSQLCompat(dbConn, `SELECT id, pais_codigo, version, COALESCE(nombre,''), COALESCE(vigencia_desde,''), COALESCE(vigencia_hasta,''), COALESCE(estado,'vigente'), COALESCE(fuente,''), COALESCE(descripcion,''), COALESCE(fecha_creacion,''), COALESCE(actualizado_por,'')
		FROM catalogo_legal_pais_versiones
		WHERE pais_codigo = ? AND LOWER(COALESCE(estado,'vigente')) IN ('vigente','publicado','activo')
		ORDER BY COALESCE(vigencia_desde,'') DESC, version DESC
		LIMIT 1`, paisCodigo)
	var out CatalogoLegalVersion
	if err := row.Scan(&out.ID, &out.PaisCodigo, &out.Version, &out.Nombre, &out.VigenciaDesde, &out.VigenciaHasta, &out.Estado, &out.Fuente, &out.Descripcion, &out.FechaCreacion, &out.ActualizadoPor); err != nil {
		return out, err
	}
	return out, nil
}

func ListCatalogoLegalParametros(dbConn *sql.DB, versionID int64) ([]CatalogoLegalParametro, error) {
	rows, err := querySQLCompat(dbConn, `SELECT id, version_id, codigo, COALESCE(grupo,''), COALESCE(nombre,''), COALESCE(tipo_valor,'numero'), COALESCE(valor_numero,0), COALESCE(valor_texto,''), COALESCE(habilitado,1), COALESCE(orden,0), COALESCE(observaciones,''), COALESCE(fecha_creacion,''), COALESCE(actualizado_por,'')
		FROM catalogo_legal_pais_parametros
		WHERE version_id = ?
		ORDER BY orden ASC, grupo ASC, codigo ASC`, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]CatalogoLegalParametro, 0)
	for rows.Next() {
		var p CatalogoLegalParametro
		if err := rows.Scan(&p.ID, &p.VersionID, &p.Codigo, &p.Grupo, &p.Nombre, &p.TipoValor, &p.ValorNumero, &p.ValorTexto, &p.Habilitado, &p.Orden, &p.Observaciones, &p.FechaCreacion, &p.ActualizadoPor); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func GetEmpresaParametrosLegalesEstado(dbConn *sql.DB, empresaID int64, paisCodigo string) (EmpresaParametrosLegalesEstado, error) {
	if empresaID <= 0 {
		return EmpresaParametrosLegalesEstado{}, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureCatalogoLegalPaisSchema(dbConn); err != nil {
		return EmpresaParametrosLegalesEstado{}, err
	}
	latest, err := GetLatestCatalogoLegalVersion(dbConn, paisCodigo)
	if err != nil {
		return EmpresaParametrosLegalesEstado{}, err
	}
	params, _ := ListCatalogoLegalParametros(dbConn, latest.ID)
	out := EmpresaParametrosLegalesEstado{
		EmpresaID:         empresaID,
		PaisCodigo:        latest.PaisCodigo,
		VersionDisponible: latest.Version,
		NombreDisponible:  latest.Nombre,
		VigenciaDesde:     latest.VigenciaDesde,
		Parametros:        params,
	}
	row := queryRowSQLCompat(dbConn, `SELECT COALESCE(version,''), COALESCE(auto_actualizar,0), COALESCE(fecha_ultima_revision,''), COALESCE(fecha_aplicacion,''), COALESCE(modo_aplicacion,'')
		FROM empresa_parametros_legales_aplicados
		WHERE empresa_id = ? AND pais_codigo = ?
		LIMIT 1`, empresaID, latest.PaisCodigo)
	var autoInt int
	if err := row.Scan(&out.VersionAplicada, &autoInt, &out.FechaUltimaRevision, &out.FechaUltimaAplicacion, &out.ModoUltimaAplicacion); err != nil && err != sql.ErrNoRows {
		return out, err
	}
	if strings.EqualFold(strings.TrimSpace(out.VersionAplicada), "SIN_APLICAR") {
		out.VersionAplicada = ""
	}
	out.AutoActualizar = autoInt == 1
	out.Pendiente = strings.TrimSpace(out.VersionAplicada) == "" || strings.TrimSpace(out.VersionAplicada) != strings.TrimSpace(out.VersionDisponible)
	if out.Pendiente {
		out.Mensaje = "Hay parametros legales disponibles para aplicar."
	} else {
		out.Mensaje = "Parametros legales al dia."
	}
	_, _ = execSQLCompat(dbConn, `UPDATE empresa_parametros_legales_aplicados SET fecha_ultima_revision = CURRENT_TIMESTAMP, fecha_actualizacion = CURRENT_TIMESTAMP WHERE empresa_id = ? AND pais_codigo = ?`, empresaID, latest.PaisCodigo)
	return out, nil
}

func SetEmpresaParametrosLegalesAutoActualizar(dbConn *sql.DB, empresaID int64, paisCodigo string, enabled bool, usuario string) (EmpresaParametrosLegalesEstado, error) {
	if empresaID <= 0 {
		return EmpresaParametrosLegalesEstado{}, fmt.Errorf("empresa_id es obligatorio")
	}
	latest, err := GetLatestCatalogoLegalVersion(dbConn, paisCodigo)
	if err != nil {
		return EmpresaParametrosLegalesEstado{}, err
	}
	estadoActual, _ := GetEmpresaParametrosLegalesEstado(dbConn, empresaID, latest.PaisCodigo)
	versionRegistro := strings.TrimSpace(estadoActual.VersionAplicada)
	if versionRegistro == "" {
		versionRegistro = "SIN_APLICAR"
	}
	auto := 0
	if enabled {
		auto = 1
	}
	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = colombiaDefaultsUsuarioSistema
	}
	_, err = insertSQLCompat(dbConn, `INSERT INTO empresa_parametros_legales_aplicados (
		empresa_id, pais_codigo, version, auto_actualizar, fecha_ultima_revision, fecha_aplicacion,
		usuario_creador, modo_aplicacion, estado, resumen_json, observaciones, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, 'preferencia', 'activo', '{}', ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT(empresa_id, pais_codigo) DO UPDATE SET
		auto_actualizar = excluded.auto_actualizar,
		fecha_ultima_revision = CURRENT_TIMESTAMP,
		fecha_actualizacion = CURRENT_TIMESTAMP,
		usuario_creador = excluded.usuario_creador,
		observaciones = excluded.observaciones
	RETURNING id`,
		empresaID,
		latest.PaisCodigo,
		versionRegistro,
		auto,
		usuario,
		"Preferencia de actualizacion legal por empresa",
	)
	if err != nil {
		return EmpresaParametrosLegalesEstado{}, err
	}
	return GetEmpresaParametrosLegalesEstado(dbConn, empresaID, latest.PaisCodigo)
}

func ApplyEmpresaParametrosLegalesLatest(dbConn *sql.DB, empresaID int64, paisCodigo, usuario, modo string) (EmpresaParametrosLegalesApplyResult, error) {
	out := EmpresaParametrosLegalesApplyResult{}
	if empresaID <= 0 {
		return out, fmt.Errorf("empresa_id es obligatorio")
	}
	latest, err := GetLatestCatalogoLegalVersion(dbConn, paisCodigo)
	if err != nil {
		return out, err
	}
	prev, _ := GetEmpresaParametrosLegalesEstado(dbConn, empresaID, latest.PaisCodigo)
	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = colombiaDefaultsUsuarioSistema
	}
	modo = strings.ToLower(strings.TrimSpace(modo))
	if modo == "" {
		modo = "manual"
	}
	defaults, err := ApplyEmpresaColombiaDefaults(dbConn, empresaID, usuario)
	out.EmpresaColombiaDefaultsResult = defaults
	out.VersionAplicada = latest.Version
	out.AutoActualizar = prev.AutoActualizar
	out.ModoAplicacion = modo
	if err != nil {
		return out, err
	}
	summaryRaw, _ := json.Marshal(map[string]interface{}{
		"version":             latest.Version,
		"modo":                modo,
		"impuestos":           defaults.Impuestos,
		"conceptos_nomina":    defaults.ConceptosNomina,
		"nomina_config_id":    defaults.NominaConfiguracionID,
		"bodega_id":           defaults.BodegaID,
		"version_anterior":    prev.VersionAplicada,
		"vigencia_desde":      latest.VigenciaDesde,
		"catalogo_version_id": latest.ID,
	})
	auto := 0
	if prev.AutoActualizar {
		auto = 1
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_parametros_legales_aplicados (
		empresa_id, pais_codigo, version, auto_actualizar, fecha_ultima_revision, fecha_aplicacion,
		usuario_creador, modo_aplicacion, estado, resumen_json, observaciones, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?, 'activo', ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT(empresa_id, pais_codigo) DO UPDATE SET
		version = excluded.version,
		auto_actualizar = excluded.auto_actualizar,
		fecha_ultima_revision = CURRENT_TIMESTAMP,
		fecha_aplicacion = CURRENT_TIMESTAMP,
		usuario_creador = excluded.usuario_creador,
		modo_aplicacion = excluded.modo_aplicacion,
		estado = 'activo',
		resumen_json = excluded.resumen_json,
		observaciones = excluded.observaciones,
		fecha_actualizacion = CURRENT_TIMESTAMP
	RETURNING id`,
		empresaID,
		latest.PaisCodigo,
		latest.Version,
		auto,
		usuario,
		modo,
		string(summaryRaw),
		"Parametros legales aplicados desde catalogo versionado",
	)
	if err != nil {
		return out, err
	}
	out.RegistroAplicado = id
	return out, nil
}

func ApplyParametrosLegalesToExistingEmpresas(dbConn *sql.DB) (*EmpresasColombiaDefaultsBackfillResult, error) {
	res := &EmpresasColombiaDefaultsBackfillResult{Version: ColombiaDefaultsVersion}
	if dbConn == nil {
		return res, fmt.Errorf("db nil")
	}
	err := ApplySchemaMigration(dbConn, "empresas", "20260607_catalogo_legal_colombia_aplicado", "Registra catalogo legal Colombia versionado por empresa", func(tx *sql.DB) error {
		empresas, err := GetEmpresas(tx)
		if err != nil {
			return err
		}
		res.Empresas = len(empresas)
		for _, empresa := range empresas {
			empresaID := empresa.EmpresaID
			if empresaID <= 0 {
				empresaID = empresa.ID
			}
			estado := strings.ToLower(strings.TrimSpace(empresa.Estado))
			if empresaID <= 0 || estado == "eliminada" || estado == "eliminado" {
				continue
			}
			item, err := ApplyEmpresaParametrosLegalesLatest(tx, empresaID, "CO", colombiaDefaultsUsuarioSistema, "backfill")
			res.Items = append(res.Items, item.EmpresaColombiaDefaultsResult)
			if err != nil {
				res.Errores = append(res.Errores, fmt.Sprintf("empresa_id=%d: %v", empresaID, err))
				continue
			}
			res.Aplicadas++
		}
		if len(res.Errores) > 0 {
			return fmt.Errorf("%s", strings.Join(res.Errores, "; "))
		}
		return nil
	})
	return res, err
}

func CheckAndApplyEmpresaParametrosLegalesAuto(dbConn *sql.DB, usuario string) (*EmpresaParametrosLegalesWorkerResult, error) {
	if err := SeedCatalogoLegalColombiaBase(dbConn, usuario); err != nil {
		return nil, err
	}
	latest, err := GetLatestCatalogoLegalVersion(dbConn, "CO")
	if err != nil {
		return nil, err
	}
	res := &EmpresaParametrosLegalesWorkerResult{Version: latest.Version}
	empresas, err := GetEmpresas(dbConn)
	if err != nil {
		return res, err
	}
	for _, empresa := range empresas {
		empresaID := empresa.EmpresaID
		if empresaID <= 0 {
			empresaID = empresa.ID
		}
		estadoEmpresa := strings.ToLower(strings.TrimSpace(empresa.Estado))
		if empresaID <= 0 || estadoEmpresa == "eliminada" || estadoEmpresa == "eliminado" {
			continue
		}
		res.EmpresasRevisadas++
		estado, err := GetEmpresaParametrosLegalesEstado(dbConn, empresaID, "CO")
		if err != nil {
			res.Errores = append(res.Errores, fmt.Sprintf("empresa_id=%d: %v", empresaID, err))
			continue
		}
		if estado.Pendiente {
			res.Pendientes++
		}
		if estado.Pendiente && estado.AutoActualizar {
			if _, err := ApplyEmpresaParametrosLegalesLatest(dbConn, empresaID, "CO", usuario, "automatico"); err != nil {
				res.Errores = append(res.Errores, fmt.Sprintf("empresa_id=%d aplicar: %v", empresaID, err))
				continue
			}
			res.Aplicadas++
		}
	}
	if len(res.Errores) > 0 {
		return res, fmt.Errorf("%s", strings.Join(res.Errores, "; "))
	}
	return res, nil
}

func StartEmpresaParametrosLegalesWorker(dbConn *sql.DB, interval time.Duration, stop <-chan struct{}) {
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	run := func() {
		res, err := CheckAndApplyEmpresaParametrosLegalesAuto(dbConn, "sistema.worker_parametros_legales")
		if err != nil {
			fmt.Printf("[parametros_legales_worker] error: %v\n", err)
			return
		}
		if res != nil {
			fmt.Printf("[parametros_legales_worker] version=%s empresas=%d pendientes=%d aplicadas=%d\n", res.Version, res.EmpresasRevisadas, res.Pendientes, res.Aplicadas)
		}
	}
	run()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			run()
		case <-stop:
			return
		}
	}
}

func EmpresaImpuestosCatalogoBase(pais string) []EmpresaImpuestoConfig {
	pais = strings.ToUpper(strings.TrimSpace(pais))
	switch pais {
	case "EC":
		return []EmpresaImpuestoConfig{
			{PaisCodigo: "EC", Codigo: "IVA", Nombre: "IVA tarifa general", Tipo: "impuesto", TasaPorcentaje: 15, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "EC", Codigo: "IVA_0", Nombre: "IVA 0% / exento", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "EC", Codigo: "ICE", Nombre: "ICE consumos especiales", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "EC", Codigo: "RET_IVA", Nombre: "Retencion IVA segun SRI", Tipo: "retencion", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "compras"},
			{PaisCodigo: "EC", Codigo: "RET_IR", Nombre: "Retencion IR segun tabla SRI", Tipo: "retencion", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "compras"},
		}
	case "PA":
		return []EmpresaImpuestoConfig{
			{PaisCodigo: "PA", Codigo: "ITBMS_7", Nombre: "ITBMS 7% tasa general", Tipo: "impuesto", TasaPorcentaje: 7, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "PA", Codigo: "ITBMS_10", Nombre: "ITBMS 10% rubros especiales", Tipo: "impuesto", TasaPorcentaje: 10, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "PA", Codigo: "ITBMS_15", Nombre: "ITBMS 15% rubros especiales", Tipo: "impuesto", TasaPorcentaje: 15, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "PA", Codigo: "ISC", Nombre: "ISC selectivo al consumo", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "PA", Codigo: "RET_ITBMS", Nombre: "Retencion ITBMS segun condicion", Tipo: "retencion", TasaPorcentaje: 50, Habilitado: 0, AplicaEn: "compras"},
		}
	case "CR":
		return []EmpresaImpuestoConfig{
			{PaisCodigo: "CR", Codigo: "IVA_13", Nombre: "IVA 13% tarifa general", Tipo: "impuesto", TasaPorcentaje: 13, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "CR", Codigo: "IVA_4", Nombre: "IVA 4% tarifa reducida", Tipo: "impuesto", TasaPorcentaje: 4, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "CR", Codigo: "IVA_2", Nombre: "IVA 2% tarifa reducida", Tipo: "impuesto", TasaPorcentaje: 2, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "CR", Codigo: "IVA_1", Nombre: "IVA 1% tarifa reducida", Tipo: "impuesto", TasaPorcentaje: 1, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "CR", Codigo: "EXENTO", Nombre: "Exento / no sujeto", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 1, AplicaEn: "ventas"},
		}
	case "AR":
		return []EmpresaImpuestoConfig{
			{PaisCodigo: "AR", Codigo: "IVA_21", Nombre: "IVA 21% tarifa general", Tipo: "impuesto", TasaPorcentaje: 21, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "AR", Codigo: "IVA_105", Nombre: "IVA 10.5% tarifa reducida", Tipo: "impuesto", TasaPorcentaje: 10.5, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "AR", Codigo: "IVA_27", Nombre: "IVA 27% tarifa diferencial", Tipo: "impuesto", TasaPorcentaje: 27, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "AR", Codigo: "EXENTO", Nombre: "Exento / no gravado", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "AR", Codigo: "RET_GAN", Nombre: "Retencion ganancias segun regimen", Tipo: "retencion", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "compras"},
			{PaisCodigo: "AR", Codigo: "IIBB", Nombre: "Ingresos brutos jurisdiccional", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "ventas"},
		}
	case "VE":
		return []EmpresaImpuestoConfig{
			{PaisCodigo: "VE", Codigo: "IVA_16", Nombre: "IVA 16% tarifa general", Tipo: "impuesto", TasaPorcentaje: 16, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "VE", Codigo: "IVA_8", Nombre: "IVA 8% tarifa reducida", Tipo: "impuesto", TasaPorcentaje: 8, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "VE", Codigo: "IVA_31", Nombre: "IVA adicional 31% rubros especiales", Tipo: "impuesto", TasaPorcentaje: 31, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "VE", Codigo: "EXENTO", Nombre: "Exento / no sujeto", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "VE", Codigo: "IGTF", Nombre: "IGTF segun medio de pago", Tipo: "impuesto", TasaPorcentaje: 3, Habilitado: 0, AplicaEn: "ventas"},
		}
	default:
		return EmpresaImpuestosCatalogoColombia()
	}
}

func EmpresaImpuestosCatalogoColombia() []EmpresaImpuestoConfig {
	return []EmpresaImpuestoConfig{
		{PaisCodigo: "CO", Codigo: "IVA", Nombre: "IVA tarifa general 19%", Tipo: "impuesto", TasaPorcentaje: 19, Habilitado: 1, AplicaEn: "ventas"},
		{PaisCodigo: "CO", Codigo: "IVA_0", Nombre: "IVA 0% / exento / excluido", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 1, AplicaEn: "ventas"},
		{PaisCodigo: "CO", Codigo: "INC_8", Nombre: "Impuesto nacional al consumo 8%", Tipo: "impuesto", TasaPorcentaje: 8, Habilitado: 0, AplicaEn: "ventas"},
		{PaisCodigo: "CO", Codigo: "ICA", Nombre: "ICA municipal variable", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "ventas"},
		{PaisCodigo: "CO", Codigo: "RETEFUENTE", Nombre: "Retencion en la fuente renta", Tipo: "retencion", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "compras"},
		{PaisCodigo: "CO", Codigo: "RETEIVA", Nombre: "Retencion a titulo de IVA 15%", Tipo: "retencion", TasaPorcentaje: 15, Habilitado: 0, AplicaEn: "compras"},
		{PaisCodigo: "CO", Codigo: "RETEICA", Nombre: "Retencion a titulo de ICA", Tipo: "retencion", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "compras"},
	}
}

func ApplyEmpresaColombiaDefaults(dbConn *sql.DB, empresaID int64, usuario string) (EmpresaColombiaDefaultsResult, error) {
	res := EmpresaColombiaDefaultsResult{EmpresaID: empresaID, Version: ColombiaDefaultsVersion}
	if dbConn == nil {
		return res, fmt.Errorf("db nil")
	}
	if empresaID <= 0 {
		return res, fmt.Errorf("empresa_id es obligatorio")
	}
	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = colombiaDefaultsUsuarioSistema
	}
	if err := EnsureEmpresaImpuestosSchema(dbConn); err != nil {
		return res, err
	}
	if err := EnsureEmpresaNominaSchema(dbConn); err != nil {
		return res, err
	}
	if err := EnsureEmpresaEstacionPrefsSchema(dbConn); err != nil {
		return res, err
	}

	bodegaID, err := EnsureEmpresaBodega1(dbConn, empresaID, usuario)
	if err != nil {
		res.Errores = append(res.Errores, "bodega_1: "+err.Error())
	} else {
		res.BodegaID = bodegaID
	}

	for _, imp := range EmpresaImpuestosCatalogoColombia() {
		imp.EmpresaID = empresaID
		imp.UsuarioCreador = usuario
		imp.Estado = "activo"
		imp.Observaciones = colombiaDefaultsObservacionPreprod + " version=" + ColombiaDefaultsVersion
		if _, err := UpsertEmpresaImpuesto(dbConn, imp); err != nil {
			res.Errores = append(res.Errores, fmt.Sprintf("impuesto %s: %v", imp.Codigo, err))
			continue
		}
		res.Impuestos++
	}

	cfg := defaultEmpresaNominaConfiguracion(empresaID)
	cfg.UsuarioCreador = usuario
	cfg.Observaciones = colombiaDefaultsObservacionPreprod + " version=" + ColombiaDefaultsVersion
	id, err := UpsertEmpresaNominaConfiguracion(dbConn, cfg)
	if err != nil {
		res.Errores = append(res.Errores, "nomina_configuracion: "+err.Error())
	} else {
		res.NominaConfiguracionID = id
	}

	if err := SeedEmpresaNominaColombiaConceptosBase(dbConn, empresaID, usuario); err != nil {
		res.Errores = append(res.Errores, "nomina_conceptos: "+err.Error())
	} else {
		res.ConceptosNomina = len(nominaColombiaConceptosProfesionales(empresaID, usuario))
	}

	markerRaw, _ := json.Marshal(map[string]interface{}{
		"version":                    ColombiaDefaultsVersion,
		"salario_minimo_mensual":     ColombiaSalarioMinimoMensual2026,
		"auxilio_transporte_mensual": ColombiaAuxilioTransporteMensual2026,
		"impuestos":                  res.Impuestos,
		"conceptos_nomina":           res.ConceptosNomina,
		"bodega_id":                  res.BodegaID,
		"bodega_nombre":              "Bodega 1",
		"observacion":                "preconfiguracion Colombia aplicada en preproduccion",
	})
	markerID, err := UpsertEmpresaEstacionPref(dbConn, EmpresaEstacionPref{
		EmpresaID:      empresaID,
		EstacionID:     0,
		Clave:          colombiaDefaultsPrefClave,
		Valor:          string(markerRaw),
		UsuarioCreador: usuario,
		Estado:         "activo",
		Observaciones:  colombiaDefaultsObservacionPreprod,
	})
	if err != nil {
		res.Errores = append(res.Errores, "marker: "+err.Error())
	} else {
		res.MarkerID = markerID
	}

	if len(res.Errores) > 0 {
		return res, fmt.Errorf("%s", strings.Join(res.Errores, "; "))
	}
	return res, nil
}

func ApplyColombiaDefaultsToExistingEmpresas(dbConn *sql.DB) (*EmpresasColombiaDefaultsBackfillResult, error) {
	res := &EmpresasColombiaDefaultsBackfillResult{Version: ColombiaDefaultsVersion}
	if dbConn == nil {
		return res, fmt.Errorf("db nil")
	}
	err := ApplySchemaMigration(dbConn, "empresas", colombiaDefaultsMigrationVersion, colombiaDefaultsMigrationDescription, func(tx *sql.DB) error {
		empresas, err := GetEmpresas(tx)
		if err != nil {
			return err
		}
		res.Empresas = len(empresas)
		for _, empresa := range empresas {
			empresaID := empresa.EmpresaID
			if empresaID <= 0 {
				empresaID = empresa.ID
			}
			estado := strings.ToLower(strings.TrimSpace(empresa.Estado))
			if empresaID <= 0 || estado == "eliminada" || estado == "eliminado" {
				continue
			}
			item, err := ApplyEmpresaColombiaDefaults(tx, empresaID, colombiaDefaultsUsuarioSistema)
			res.Items = append(res.Items, item)
			if err != nil {
				res.Errores = append(res.Errores, fmt.Sprintf("empresa_id=%d: %v", empresaID, err))
				continue
			}
			res.Aplicadas++
		}
		if len(res.Errores) > 0 {
			return fmt.Errorf("%s", strings.Join(res.Errores, "; "))
		}
		return nil
	})
	return res, err
}
