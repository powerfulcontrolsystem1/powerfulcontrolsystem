package db

import (
	"context"
	"database/sql"
	"fmt"
)

const empresaDIANNominaElectronicaFingerprint = "empresa-dian-nomina-electronica:v2:monthly-source-profiles-numbering-provider"

// applyEmpresaDIANNominaElectronicaTx adds the data that the DIAN payroll
// annex requires without guessing it for legacy employees or payroll rows.
// Existing records remain drafts until their explicit fiscal profile is
// completed and a paid payroll settlement can be sealed as source data.
func applyEmpresaDIANNominaElectronicaTx(ctx context.Context, tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}

	var incompatible int64
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*)
		FROM empresa_contabilidad_nomina_electronica
		WHERE COALESCE(salario_base, 0) < 0
		   OR COALESCE(devengados, 0) < 0
		   OR COALESCE(deducciones, 0) < 0
		   OR COALESCE(total, 0) < 0
		   OR COALESCE(deducciones, 0) > COALESCE(devengados, 0)
		   OR ABS(COALESCE(salario_base, 0)::numeric - ROUND(COALESCE(salario_base, 0)::numeric, 2)) > 0.005
		   OR ABS(COALESCE(devengados, 0)::numeric - ROUND(COALESCE(devengados, 0)::numeric, 2)) > 0.005
		   OR ABS(COALESCE(deducciones, 0)::numeric - ROUND(COALESCE(deducciones, 0)::numeric, 2)) > 0.005
		   OR ABS(COALESCE(total, 0)::numeric - ROUND(COALESCE(total, 0)::numeric, 2)) > 0.005`).Scan(&incompatible); err != nil {
		return fmt.Errorf("inspect electronic payroll monetary precision: %w", err)
	}
	if incompatible > 0 {
		return fmt.Errorf("empresa_contabilidad_nomina_electronica contains %d monetary rows requiring manual reconciliation", incompatible)
	}

	for _, statement := range empresaDIANNominaElectronicaMigrationStatements() {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("migrate DIAN electronic payroll schema: %w", err)
		}
	}
	return nil
}

func empresaDIANNominaElectronicaMigrationStatements() []string {
	statements := empresaDIANNominaElectronicaMoneyStatements()
	statements = append(statements, empresaDIANNominaElectronicaSchemaStatements()...)
	statements = append(statements, empresaDIANNominaElectronicaConstraintStatements()...)
	return statements
}

func empresaDIANNominaElectronicaMoneyStatements() []string {
	return []string{
		`UPDATE empresa_contabilidad_nomina_electronica
			SET salario_base = COALESCE(salario_base, 0), devengados = COALESCE(devengados, 0),
				deducciones = COALESCE(deducciones, 0), total = COALESCE(total, 0)`,
		`ALTER TABLE empresa_contabilidad_nomina_electronica
			ALTER COLUMN salario_base TYPE NUMERIC(18,2) USING ROUND(COALESCE(salario_base, 0)::numeric, 2),
			ALTER COLUMN devengados TYPE NUMERIC(18,2) USING ROUND(COALESCE(devengados, 0)::numeric, 2),
			ALTER COLUMN deducciones TYPE NUMERIC(18,2) USING ROUND(COALESCE(deducciones, 0)::numeric, 2),
			ALTER COLUMN total TYPE NUMERIC(18,2) USING ROUND(COALESCE(total, 0)::numeric, 2)`,
		`ALTER TABLE empresa_contabilidad_nomina_electronica
			ALTER COLUMN salario_base SET DEFAULT 0,
			ALTER COLUMN salario_base SET NOT NULL,
			ALTER COLUMN devengados SET DEFAULT 0,
			ALTER COLUMN devengados SET NOT NULL,
			ALTER COLUMN deducciones SET DEFAULT 0,
			ALTER COLUMN deducciones SET NOT NULL,
			ALTER COLUMN total SET DEFAULT 0,
			ALTER COLUMN total SET NOT NULL`,
	}
}

func empresaDIANNominaElectronicaSchemaStatements() []string {
	return []string{
		`ALTER TABLE empresa_contabilidad_nomina_electronica
			ADD COLUMN IF NOT EXISTS liquidacion_id BIGINT NOT NULL DEFAULT 0,
			ADD COLUMN IF NOT EXISTS empleado_nomina_id BIGINT NOT NULL DEFAULT 0,
			ADD COLUMN IF NOT EXISTS periodo_reporte TEXT NOT NULL DEFAULT '',
			ADD COLUMN IF NOT EXISTS numero_legal TEXT NOT NULL DEFAULT '',
			ADD COLUMN IF NOT EXISTS fecha_emision_legal TIMESTAMPTZ,
			ADD COLUMN IF NOT EXISTS configuracion_dian_json JSONB NOT NULL DEFAULT '{}'::jsonb,
			ADD COLUMN IF NOT EXISTS fuente_fiscal_json JSONB NOT NULL DEFAULT '{}'::jsonb,
			ADD COLUMN IF NOT EXISTS fuente_fiscal_sellada BOOLEAN NOT NULL DEFAULT FALSE,
			ADD COLUMN IF NOT EXISTS intentos INTEGER NOT NULL DEFAULT 0,
			ADD COLUMN IF NOT EXISTS fecha_ultimo_intento TIMESTAMPTZ`,
		`CREATE TABLE IF NOT EXISTS empresa_nomina_dian_perfiles (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			empleado_nomina_id BIGINT NOT NULL,
			tipo_documento_dian TEXT NOT NULL DEFAULT '',
			primer_apellido TEXT NOT NULL DEFAULT '',
			segundo_apellido TEXT NOT NULL DEFAULT '',
			primer_nombre TEXT NOT NULL DEFAULT '',
			otros_nombres TEXT NOT NULL DEFAULT '',
			tipo_trabajador_dian TEXT NOT NULL DEFAULT '',
			subtipo_trabajador_dian TEXT NOT NULL DEFAULT '',
			alto_riesgo_pension BOOLEAN NOT NULL DEFAULT FALSE,
			salario_integral BOOLEAN NOT NULL DEFAULT FALSE,
			fecha_retiro DATE,
			lugar_trabajo_pais TEXT NOT NULL DEFAULT 'CO',
			lugar_trabajo_departamento_codigo_dane TEXT NOT NULL DEFAULT '',
			lugar_trabajo_municipio_codigo_dane TEXT NOT NULL DEFAULT '',
			lugar_trabajo_direccion TEXT NOT NULL DEFAULT '',
			tipo_contrato_dian TEXT NOT NULL DEFAULT '',
			metodo_pago_dian TEXT NOT NULL DEFAULT '',
			banco TEXT NOT NULL DEFAULT '',
			tipo_cuenta TEXT NOT NULL DEFAULT '',
			numero_cuenta TEXT NOT NULL DEFAULT '',
			usuario_creador TEXT NOT NULL DEFAULT '',
			estado TEXT NOT NULL DEFAULT 'activo',
			observaciones TEXT NOT NULL DEFAULT '',
			fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT uq_empresa_nomina_dian_perfil UNIQUE (empresa_id, empleado_nomina_id),
			CONSTRAINT ck_empresa_nomina_dian_perfil_estado CHECK (estado IN ('activo', 'inactivo'))
		)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_nomina_dian_perfiles_empresa
			ON empresa_nomina_dian_perfiles (empresa_id, estado, empleado_nomina_id)`,
		`ALTER TABLE empresa_nomina_configuracion
			ADD COLUMN IF NOT EXISTS periodo_nomina_dian INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE empresa_dian_configuracion
			ADD COLUMN IF NOT EXISTS software_proveedor_nit TEXT NOT NULL DEFAULT '',
			ADD COLUMN IF NOT EXISTS software_proveedor_dv TEXT NOT NULL DEFAULT '',
			ADD COLUMN IF NOT EXISTS software_proveedor_razon_social TEXT NOT NULL DEFAULT '',
			ADD COLUMN IF NOT EXISTS software_proveedor_primer_apellido TEXT NOT NULL DEFAULT '',
			ADD COLUMN IF NOT EXISTS software_proveedor_segundo_apellido TEXT NOT NULL DEFAULT '',
			ADD COLUMN IF NOT EXISTS software_proveedor_primer_nombre TEXT NOT NULL DEFAULT '',
			ADD COLUMN IF NOT EXISTS software_proveedor_otros_nombres TEXT NOT NULL DEFAULT ''`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_empresa_nomina_electronica_liquidacion
			ON empresa_contabilidad_nomina_electronica (empresa_id, liquidacion_id)
			WHERE liquidacion_id > 0`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_empresa_nomina_electronica_empleado_periodo
			ON empresa_contabilidad_nomina_electronica (empresa_id, empleado_nomina_id, periodo_reporte)
			WHERE empleado_nomina_id > 0 AND periodo_reporte <> ''`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_empresa_nomina_electronica_numero_legal
			ON empresa_contabilidad_nomina_electronica (empresa_id, numero_legal)
			WHERE numero_legal <> ''`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_nomina_electronica_estado
			ON empresa_contabilidad_nomina_electronica (empresa_id, estado_dian, id)`,
	}
}

func empresaDIANNominaElectronicaConstraintStatements() []string {
	return []string{
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint
				WHERE conname = 'ck_empresa_nomina_electronica_periodo_reporte'
				  AND conrelid = 'empresa_contabilidad_nomina_electronica'::regclass
			) THEN
				ALTER TABLE empresa_contabilidad_nomina_electronica
					ADD CONSTRAINT ck_empresa_nomina_electronica_periodo_reporte
					CHECK (periodo_reporte = '' OR periodo_reporte ~ '^[0-9]{4}-(0[1-9]|1[0-2])$');
			END IF;
		END $$`,
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint
				WHERE conname = 'ck_empresa_nomina_electronica_importes'
				  AND conrelid = 'empresa_contabilidad_nomina_electronica'::regclass
			) THEN
				ALTER TABLE empresa_contabilidad_nomina_electronica
					ADD CONSTRAINT ck_empresa_nomina_electronica_importes
					CHECK (salario_base >= 0 AND devengados >= 0 AND deducciones >= 0 AND total >= 0 AND deducciones <= devengados);
			END IF;
		END $$`,
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint
				WHERE conname = 'ck_empresa_nomina_electronica_cune'
				  AND conrelid = 'empresa_contabilidad_nomina_electronica'::regclass
			) THEN
				ALTER TABLE empresa_contabilidad_nomina_electronica
					ADD CONSTRAINT ck_empresa_nomina_electronica_cune
					CHECK (COALESCE(cune, '') = '' OR COALESCE(cune, '') ~ '^[0-9a-fA-F]{96}$');
			END IF;
		END $$`,
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint
				WHERE conname = 'ck_empresa_nomina_periodo_dian'
				  AND conrelid = 'empresa_nomina_configuracion'::regclass
			) THEN
				ALTER TABLE empresa_nomina_configuracion
					ADD CONSTRAINT ck_empresa_nomina_periodo_dian CHECK (periodo_nomina_dian BETWEEN 0 AND 6);
			END IF;
		END $$`,
	}
}
