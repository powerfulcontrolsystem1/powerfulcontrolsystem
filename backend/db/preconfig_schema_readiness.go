package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type schemaReadinessCheck struct {
	name  string
	query string
}

func requireSchemaReadiness(dbConn *sql.DB, scope string, checks []schemaReadinessCheck) error {
	if dbConn == nil {
		return errors.New("conexion de base de datos no disponible")
	}
	for _, check := range checks {
		var marker int
		err := queryRowSQLCompat(dbConn, check.query).Scan(&marker)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return fmt.Errorf("esquema de %s no disponible (%s): %w", scope, check.name, err)
		}
	}
	return nil
}

// requiredTableColumnsExist centraliza la inspeccion PostgreSQL usada por los
// verificadores de esquema. Solo consulta metadata; nunca crea ni altera DDL.
func requiredTableColumnsExist(dbConn *sql.DB, tableName string, columns []string) (bool, error) {
	if len(columns) == 0 {
		return true, nil
	}
	found := make(map[string]bool, len(columns))
	rows, err := querySQLCompat(dbConn, `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = ANY (current_schemas(false))
		  AND table_name = ?
	`, strings.TrimSpace(tableName))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return false, err
		}
		found[strings.ToLower(strings.TrimSpace(columnName))] = true
	}
	if err := rows.Err(); err != nil {
		return false, err
	}
	for _, columnName := range columns {
		if !found[strings.ToLower(strings.TrimSpace(columnName))] {
			return false, nil
		}
	}
	return true, nil
}

func normalizeListLimitOffset(limit, offset, defaultLimit, maxLimit int) (int, int) {
	if limit <= 0 {
		limit = defaultLimit
	}
	if maxLimit > 0 && limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// firstNonBlankValue returns the first non-empty value after trimming it. It is
// shared by repositories so defaults are applied consistently.
func firstNonBlankValue(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// escapedContainsPattern builds a PostgreSQL LIKE pattern for clauses that use
// ESCAPE '!'. Keeping the escaping here prevents repositories from drifting.
func escapedContainsPattern(raw string) string {
	value := strings.TrimSpace(raw)
	value = strings.ReplaceAll(value, "!", "!!")
	value = strings.ReplaceAll(value, "%", "!%")
	value = strings.ReplaceAll(value, "_", "!_")
	return "%" + value + "%"
}

func currentSchemaIndexExists(dbConn *sql.DB, indexName string) (bool, error) {
	if dbConn == nil {
		return false, errors.New("conexion de base de datos no disponible")
	}
	var exists bool
	err := queryRowSQLCompat(dbConn, `
		SELECT EXISTS (
			SELECT 1
			FROM pg_indexes
			WHERE schemaname = ANY (current_schemas(false))
			  AND indexname = ?
		)
	`, strings.TrimSpace(indexName)).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// EmpresaProductosSchemaReady valida el contrato minimo usado al aplicar una
// preconfiguracion. No crea tablas ni indices durante la solicitud HTTP.
func EmpresaProductosSchemaReady(dbConn *sql.DB) error {
	checks := []schemaReadinessCheck{
		{"bodegas", `SELECT id FROM bodegas WHERE 1=0`},
		{"categorias_productos", `SELECT id FROM categorias_productos WHERE 1=0`},
		{"productos", `SELECT id FROM productos WHERE 1=0`},
		{"proveedores", `SELECT id FROM proveedores WHERE 1=0`},
		{"servicios", `SELECT id FROM servicios WHERE 1=0`},
		{"inventario_existencias", `SELECT id FROM inventario_existencias WHERE 1=0`},
		{"inventario_movimientos", `SELECT id FROM inventario_movimientos WHERE 1=0`},
	}
	return requireSchemaReadiness(dbConn, "productos e inventario", checks)
}

// EmpresaUsuariosAuthSchemaReady valida las columnas de autenticacion que usa
// la preconfiguracion sin ejecutar DDL ni limpiar usuarios reservados.
func EmpresaUsuariosAuthSchemaReady(dbConn *sql.DB) error {
	return requireSchemaReadiness(dbConn, "usuarios empresariales", []schemaReadinessCheck{
		{"users", `SELECT id, empresa_id, email, password_hash, estado FROM users WHERE 1=0`},
	})
}

// EmpresaConfiguracionOperativaSchemaReady valida el esquema operativo ya
// migrado sin modificarlo desde el flujo empresarial.
func EmpresaConfiguracionOperativaSchemaReady(dbConn *sql.DB) error {
	checks := []schemaReadinessCheck{
		{"configuracion", `SELECT id, empresa_id FROM empresa_configuracion_operativa WHERE 1=0`},
		{"roles", `SELECT id, empresa_id, rol FROM empresa_configuracion_operativa_roles WHERE 1=0`},
		{"politicas", `SELECT id, empresa_id FROM empresa_configuracion_operativa_politicas WHERE 1=0`},
		{"historial", `SELECT id, empresa_id FROM empresa_configuracion_operativa_historial WHERE 1=0`},
	}
	return requireSchemaReadiness(dbConn, "configuracion operativa", checks)
}

// EmpresaComisionesServicioSchemaReady valida las tablas de comisiones sin DDL.
func EmpresaComisionesServicioSchemaReady(dbConn *sql.DB) error {
	checks := []schemaReadinessCheck{
		{"configuracion", `SELECT id, empresa_id FROM empresa_comisiones_servicio_configuracion WHERE 1=0`},
		{"escalas", `SELECT id, empresa_id FROM empresa_comisiones_servicio_escalas WHERE 1=0`},
		{"movimientos", `SELECT id, empresa_id FROM empresa_comisiones_servicio_movimientos WHERE 1=0`},
	}
	return requireSchemaReadiness(dbConn, "comisiones de servicio", checks)
}

// EmpresaTarifasPorDiaSchemaReady valida la tabla de tarifas diarias sin DDL.
func EmpresaTarifasPorDiaSchemaReady(dbConn *sql.DB) error {
	return requireSchemaReadiness(dbConn, "tarifas por dia", []schemaReadinessCheck{
		{"empresa_tarifas_por_dia", `SELECT id, empresa_id, estacion_id FROM empresa_tarifas_por_dia WHERE 1=0`},
	})
}

// EmpresaEventosContablesSchemaReady valida el contrato de eventos antes de
// registrar integraciones contables desde HTTP, sin preparar el esquema.
func EmpresaEventosContablesSchemaReady(dbConn *sql.DB) error {
	return requireSchemaReadiness(dbConn, "eventos contables", []schemaReadinessCheck{
		{"empresa_eventos_contables", `SELECT id, empresa_id, modulo, evento, entidad FROM empresa_eventos_contables WHERE 1=0`},
	})
}

// EmpresaPermisosFinosSchemaReady valida tablas y columnas de autorización sin
// otorgar al API capacidad DDL durante GET/POST.
func EmpresaPermisosFinosSchemaReady(dbConn *sql.DB) error {
	checks := []schemaReadinessCheck{
		{"modulos", `SELECT id, empresa_id, modulo, accion, permitido FROM empresa_permisos_modulos WHERE 1=0`},
		{"paginas", `SELECT id, empresa_id, pagina_clave, permitido FROM empresa_permisos_paginas WHERE 1=0`},
	}
	return requireSchemaReadiness(dbConn, "permisos finos empresariales", checks)
}

// EmpresaRappiSchemaReady valida integración y bitácora ya migradas.
func EmpresaRappiSchemaReady(dbConn *sql.DB) error {
	checks := []schemaReadinessCheck{
		{"configuracion", `SELECT id, empresa_id FROM empresa_rappi_configuracion WHERE 1=0`},
		{"ordenes", `SELECT id, empresa_id, rappi_order_id FROM empresa_rappi_ordenes WHERE 1=0`},
	}
	return requireSchemaReadiness(dbConn, "integracion Rappi", checks)
}

// RolesPermisosSchemaReady valida permisos por rol sin DDL desde el panel.
func RolesPermisosSchemaReady(dbConn *sql.DB) error {
	checks := []schemaReadinessCheck{
		{"modulos", `SELECT id, rol_id, modulo, accion, permitido FROM roles_de_usuario_permisos WHERE 1=0`},
		{"paginas", `SELECT id, rol_id, pagina_clave, permitido FROM roles_de_usuario_paginas_permisos WHERE 1=0`},
	}
	return requireSchemaReadiness(dbConn, "permisos por rol", checks)
}

// SuperAlertasSchemaReady valida configuración y eventos ya migrados.
func SuperAlertasSchemaReady(dbConn *sql.DB) error {
	checks := []schemaReadinessCheck{
		{"configuracion", `SELECT id FROM super_alertas_config WHERE 1=0`},
		{"eventos", `SELECT id, tipo, fecha_evento FROM super_alertas_eventos WHERE 1=0`},
	}
	return requireSchemaReadiness(dbConn, "alertas de plataforma", checks)
}

// SuperMantenimientoAgentesSchemaReady valida agentes y hallazgos ya migrados.
func SuperMantenimientoAgentesSchemaReady(dbConn *sql.DB) error {
	checks := []schemaReadinessCheck{
		{"agentes", `SELECT id, codigo FROM super_mantenimiento_agentes WHERE 1=0`},
		{"hallazgos", `SELECT id, agente_codigo FROM super_mantenimiento_agente_hallazgos WHERE 1=0`},
	}
	return requireSchemaReadiness(dbConn, "agentes de mantenimiento", checks)
}

// PlantillasLicenciasSchemaReady valida el contrato comercial necesario para
// sincronizar plantillas y planes sin ejecutar DDL desde el panel.
func PlantillasLicenciasSchemaReady(dbConn *sql.DB) error {
	checks := []schemaReadinessCheck{
		{"licencias", `SELECT id, tipo_id, nombre, modulos_habilitados FROM licencias WHERE 1=0`},
		{"tipos_de_empresas", `SELECT id, nombre, estado FROM tipos_de_empresas WHERE 1=0`},
		{"preconfiguraciones", `SELECT id, tipo_empresa_id, config_json FROM tipo_empresa_preconfiguraciones WHERE 1=0`},
	}
	return requireSchemaReadiness(dbConn, "plantillas y licencias", checks)
}
