package db

import (
	"database/sql"
	"fmt"
	"strings"
)

const (
	LicenciaCodigoTrial15Global     = "PLAN_GLOBAL_TRIAL_15"
	LicenciaCodigoTrial1DiaGlobal   = "PLAN_GLOBAL_TRIAL_1D_1000"
	LicenciaCodigoBasicoGlobal      = "PLAN_GLOBAL_BASICO_1000"
	LicenciaCodigoProfesionalGlobal = "PLAN_GLOBAL_PROFESIONAL_2000"
	// Codigo historico conservado para no romper renovaciones de clientes existentes.
	LicenciaCodigoEmpresarialGlobal = "PLAN_GLOBAL_EMPRESARIAL_4000"
	LicenciaCodigoAnual12000Global  = "PLAN_GLOBAL_ANUAL_12000"
	LicenciaCodigoAnual24000Global  = "PLAN_GLOBAL_ANUAL_24000"
	LicenciaCodigoAnual36000Global  = "PLAN_GLOBAL_ANUAL_36000"
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
			MaxCajasSimultaneas:    0,
		},
		{
			Codigo:                 LicenciaCodigoTrial1DiaGlobal,
			Nombre:                 "1 dia de prueba",
			Descripcion:            "Licencia de prueba de 1 dia para validar pagos, activacion y operacion real antes de contratar un plan mensual o anual.",
			Valor:                  1000,
			DuracionDias:           1,
			MaxDocumentosMensuales: 250,
			MaxCajasSimultaneas:    0,
		},
		{
			Codigo:                 LicenciaCodigoBasicoGlobal,
			Nombre:                 "Plan mensual COP 60000",
			Descripcion:            "Plan mensual global para cualquier tipo de empresa por COP 60000.",
			Valor:                  60000,
			DuracionDias:           30,
			MaxDocumentosMensuales: 1000,
			MaxCajasSimultaneas:    0,
		},
		{
			Codigo:                 LicenciaCodigoProfesionalGlobal,
			Nombre:                 "Plan mensual COP 110000",
			Descripcion:            "Plan mensual global para cualquier tipo de empresa por COP 110000.",
			Valor:                  110000,
			DuracionDias:           30,
			MaxDocumentosMensuales: 2000,
			MaxCajasSimultaneas:    0,
		},
		{
			Codigo:                 LicenciaCodigoEmpresarialGlobal,
			Nombre:                 "Plan mensual COP 200000",
			Descripcion:            "Plan mensual global para cualquier tipo de empresa por COP 200000.",
			Valor:                  200000,
			DuracionDias:           30,
			MaxDocumentosMensuales: 4000,
			MaxCajasSimultaneas:    0,
		},
		{
			Codigo:                 LicenciaCodigoAnual12000Global,
			Nombre:                 "Plan anual COP 600000",
			Descripcion:            "Plan anual global para cualquier tipo de empresa por COP 600000 con cupo de 12000 documentos electronicos.",
			Valor:                  600000,
			DuracionDias:           365,
			MaxDocumentosMensuales: 12000,
			MaxCajasSimultaneas:    0,
		},
		{
			Codigo:                 LicenciaCodigoAnual24000Global,
			Nombre:                 "Plan anual COP 1100000",
			Descripcion:            "Plan anual global para cualquier tipo de empresa por COP 1100000 con cupo de 24000 documentos electronicos.",
			Valor:                  1100000,
			DuracionDias:           365,
			MaxDocumentosMensuales: 24000,
			MaxCajasSimultaneas:    0,
		},
		{
			Codigo:                 LicenciaCodigoAnual36000Global,
			Nombre:                 "Plan anual COP 2200000",
			Descripcion:            "Plan anual global para cualquier tipo de empresa por COP 2200000 con cupo de 36000 documentos electronicos.",
			Valor:                  2200000,
			DuracionDias:           365,
			MaxDocumentosMensuales: 36000,
			MaxCajasSimultaneas:    0,
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

func IsGlobalLicenciaPlanCode(code string) bool {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return false
	}
	for _, allowed := range globalLicenciaPlanCodes() {
		if strings.ToUpper(strings.TrimSpace(allowed)) == code {
			return true
		}
	}
	return false
}

func IsGlobalLicenciaCatalogItem(lic Licencia) bool {
	if lic.EmpresaID > 0 || lic.EsAdicional == 1 {
		return false
	}
	pais := strings.ToUpper(strings.TrimSpace(lic.PaisCodigo))
	if pais == "" {
		pais = "GLOBAL"
	}
	if pais != "GLOBAL" && pais != "*" {
		return false
	}
	if lic.TipoID != 0 {
		return false
	}
	return IsGlobalLicenciaPlanCode(lic.CodigoFuncion)
}

func FilterGlobalLicenciaCatalog(rows []Licencia) []Licencia {
	out := make([]Licencia, 0, len(DefaultGlobalLicenciaPlans()))
	for _, item := range rows {
		if IsGlobalLicenciaCatalogItem(item) {
			out = append(out, item)
		}
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
	if err := deleteExtraCatalogLicenciasNoSchema(dbConn); err != nil {
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
				activo = COALESCE(activo, 1),
				estado = CASE WHEN COALESCE(activo, 1) = 1 THEN 'activo' ELSE 'inactivo' END,
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

func deleteExtraCatalogLicenciasNoSchema(dbConn *sql.DB) error {
	codes := globalLicenciaPlanCodes()
	if len(codes) == 0 {
		return nil
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(codes)), ",")
	args := make([]interface{}, 0, len(codes))
	for _, code := range codes {
		args = append(args, strings.ToUpper(strings.TrimSpace(code)))
	}
	deleteNonCanonical := fmt.Sprintf(`DELETE FROM licencias
		WHERE COALESCE(empresa_id, 0) = 0
			AND UPPER(TRIM(COALESCE(codigo_funcion, ''))) NOT IN (%s)`, placeholders)
	if _, err := execSQLCompat(dbConn, deleteNonCanonical, args...); err != nil {
		return err
	}
	duplicateArgs := make([]interface{}, 0, len(args)*2)
	duplicateArgs = append(duplicateArgs, args...)
	duplicateArgs = append(duplicateArgs, args...)
	deleteDuplicates := fmt.Sprintf(`DELETE FROM licencias
		WHERE COALESCE(empresa_id, 0) = 0
			AND UPPER(TRIM(COALESCE(codigo_funcion, ''))) IN (%s)
			AND id NOT IN (
				SELECT keep_id FROM (
					SELECT MIN(id) AS keep_id
					FROM licencias
					WHERE COALESCE(empresa_id, 0) = 0
						AND UPPER(TRIM(COALESCE(codigo_funcion, ''))) IN (%s)
					GROUP BY UPPER(TRIM(COALESCE(codigo_funcion, '')))
				) canonical_keep
			)`, placeholders, placeholders)
	_, err := execSQLCompat(dbConn, deleteDuplicates, duplicateArgs...)
	return err
}
