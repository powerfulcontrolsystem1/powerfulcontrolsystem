package db

import (
	"database/sql"
	"fmt"
	"strings"
)

const (
	LicenciaCodigoTrial15Global     = "PLAN_GLOBAL_TRIAL_15"
	LicenciaCodigoBasicoGlobal      = "PLAN_GLOBAL_BASICO_1000"
	LicenciaCodigoProfesionalGlobal = "PLAN_GLOBAL_PROFESIONAL_2000"
	LicenciaCodigoEmpresarialGlobal = "PLAN_GLOBAL_EMPRESARIAL_4000"
)

type GlobalLicenciaPlan struct {
	Codigo                 string
	Nombre                 string
	Descripcion            string
	Valor                  float64
	DuracionDias           int
	MaxDocumentosMensuales int
	MaxCajasSimultaneas    int
}

func DefaultGlobalLicenciaPlans() []GlobalLicenciaPlan {
	return []GlobalLicenciaPlan{
		{
			Codigo:                 LicenciaCodigoTrial15Global,
			Nombre:                 "Prueba gratis 15 dias",
			Descripcion:            "Licencia gratuita de prueba para cualquier tipo de empresa. Solo puede activarse una vez por empresa.",
			Valor:                  0,
			DuracionDias:           15,
			MaxDocumentosMensuales: 250,
			MaxCajasSimultaneas:    2,
		},
		{
			Codigo:                 LicenciaCodigoBasicoGlobal,
			Nombre:                 "Plan Basico mensual - 1000 documentos",
			Descripcion:            "Plan mensual global para cualquier tipo de empresa con hasta 1000 documentos o ventas.",
			Valor:                  60000,
			DuracionDias:           30,
			MaxDocumentosMensuales: 1000,
			MaxCajasSimultaneas:    2,
		},
		{
			Codigo:                 LicenciaCodigoProfesionalGlobal,
			Nombre:                 "Plan Profesional mensual - 2000 documentos",
			Descripcion:            "Plan mensual global para cualquier tipo de empresa con hasta 2000 documentos o ventas.",
			Valor:                  100000,
			DuracionDias:           30,
			MaxDocumentosMensuales: 2000,
			MaxCajasSimultaneas:    3,
		},
		{
			Codigo:                 LicenciaCodigoEmpresarialGlobal,
			Nombre:                 "Plan Empresarial mensual - 4000 documentos",
			Descripcion:            "Plan mensual global para cualquier tipo de empresa con hasta 4000 documentos o ventas.",
			Valor:                  150000,
			DuracionDias:           30,
			MaxDocumentosMensuales: 4000,
			MaxCajasSimultaneas:    4,
		},
	}
}

func globalLicenciaPlanCodes() []string {
	plans := DefaultGlobalLicenciaPlans()
	out := make([]string, 0, len(plans))
	for _, plan := range plans {
		out = append(out, strings.TrimSpace(plan.Codigo))
	}
	return out
}

// EnsureLicenciasCatalogoGlobal asegura el catalogo comercial unico compartido por todos los tipos de empresa.
func EnsureLicenciasCatalogoGlobal(dbConn *sql.DB, usuario string) (int, error) {
	if dbConn == nil {
		return 0, nil
	}
	if err := EnsureLicenciasSchema(dbConn); err != nil {
		return 0, err
	}
	return ensureLicenciasCatalogoGlobalNoSchema(dbConn, usuario)
}

func ensureLicenciasCatalogoGlobalNoSchema(dbConn *sql.DB, usuario string) (int, error) {
	if dbConn == nil {
		return 0, nil
	}
	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = "sistema.licencias_globales"
	}
	count := 0
	for _, plan := range DefaultGlobalLicenciaPlans() {
		if err := upsertGlobalLicenciaPlan(dbConn, usuario, plan); err != nil {
			return count, err
		}
		count++
	}
	if err := hideLegacyCatalogLicenciasNoSchema(dbConn); err != nil {
		return count, err
	}
	return count, nil
}

func upsertGlobalLicenciaPlan(dbConn *sql.DB, usuario string, plan GlobalLicenciaPlan) error {
	nowExpr := sqlNowExpr()
	code := strings.TrimSpace(plan.Codigo)
	if code == "" {
		return fmt.Errorf("codigo de licencia global requerido")
	}
	var existingID int64
	err := queryRowSQLCompat(dbConn, `SELECT id
		FROM licencias
		WHERE COALESCE(empresa_id, 0) = 0
			AND COALESCE(es_adicional, 0) = 0
			AND UPPER(TRIM(COALESCE(codigo_funcion, ''))) = UPPER(TRIM(?))
		ORDER BY id ASC
		LIMIT 1`, code).Scan(&existingID)
	if err == sql.ErrNoRows {
		err = queryRowSQLCompat(dbConn, `SELECT id
			FROM licencias
			WHERE COALESCE(empresa_id, 0) = 0
				AND COALESCE(es_adicional, 0) = 0
				AND LOWER(TRIM(COALESCE(nombre, ''))) = LOWER(TRIM(?))
			ORDER BY id ASC
			LIMIT 1`, plan.Nombre).Scan(&existingID)
	}
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if existingID > 0 {
		_, err = execSQLCompat(dbConn, `UPDATE licencias
			SET tipo_id = 0,
				pais_codigo = 'GLOBAL',
				nombre = ?,
				descripcion = ?,
				valor = ?,
				duracion_dias = ?,
				max_documentos_mensuales = ?,
				max_cajas_simultaneas = ?,
				modulos_habilitados = '',
				es_adicional = 0,
				codigo_funcion = ?,
				super_rol_habilitado = 0,
				fecha_inicio = '',
				fecha_fin = '',
				activo = 1,
				estado = 'activo',
				usuario_creador = COALESCE(NULLIF(TRIM(usuario_creador), ''), ?),
				observaciones = 'Plan global compartido por todos los tipos de empresa.',
				fecha_actualizacion = `+nowExpr+`
			WHERE id = ?`,
			plan.Nombre, plan.Descripcion, plan.Valor, plan.DuracionDias, plan.MaxDocumentosMensuales, plan.MaxCajasSimultaneas, code, usuario, existingID)
		return err
	}
	_, err = insertSQLCompat(dbConn, `INSERT INTO licencias (
		tipo_id,
		pais_codigo,
		nombre,
		descripcion,
		valor,
		duracion_dias,
		max_documentos_mensuales,
		max_cajas_simultaneas,
		modulos_habilitados,
		es_adicional,
		codigo_funcion,
		super_rol_habilitado,
		fecha_inicio,
		fecha_fin,
		activo,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (0, 'GLOBAL', ?, ?, ?, ?, ?, ?, '', 0, ?, 0, '', '', 1, `+nowExpr+`, `+nowExpr+`, ?, 'activo', 'Plan global compartido por todos los tipos de empresa.')`,
		plan.Nombre, plan.Descripcion, plan.Valor, plan.DuracionDias, plan.MaxDocumentosMensuales, plan.MaxCajasSimultaneas, code, usuario)
	return err
}

func hideLegacyCatalogLicenciasNoSchema(dbConn *sql.DB) error {
	codes := globalLicenciaPlanCodes()
	if len(codes) == 0 {
		return nil
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(codes)), ",")
	args := make([]interface{}, 0, len(codes))
	for _, code := range codes {
		args = append(args, strings.ToUpper(strings.TrimSpace(code)))
	}
	query := fmt.Sprintf(`UPDATE licencias
		SET activo = 0,
			estado = CASE
				WHEN LOWER(COALESCE(estado, '')) = 'eliminada' THEN estado
				ELSE 'catalogo_oculto'
			END,
			observaciones = COALESCE(NULLIF(TRIM(observaciones), ''), 'Oculta por migracion a catalogo global de 4 planes.'),
			fecha_actualizacion = `+sqlNowExpr()+`
		WHERE COALESCE(empresa_id, 0) = 0
			AND COALESCE(es_adicional, 0) = 0
			AND UPPER(TRIM(COALESCE(codigo_funcion, ''))) NOT IN (%s)`, placeholders)
	_, err := execSQLCompat(dbConn, query, args...)
	return err
}
