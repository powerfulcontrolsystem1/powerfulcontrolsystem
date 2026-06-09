package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	PowerfulSystemEmpresaName             = "Powerful Control System"
	PowerfulSystemEmpresaNameLegacyTypo   = "Powerful Control Systen"
	PowerfulSystemEmpresaConfigKey        = "licencias.facturacion_empresa_sistema_id"
	PowerfulSystemEmpresaLicenseCode      = "PCS_SYSTEM_INTERNAL_PERPETUAL"
	PowerfulSystemEmpresaLicenseDays      = 36500
	powerfulSystemEmpresaUsuarioCreador   = "sistema.powerful_control_system"
	powerfulSystemEmpresaObservacionMarca = "Empresa interna del SaaS para facturar licencias de Powerful Control System."
)

var powerfulSystemEmpresaFullModules = []string{
	"ventas",
	"inventario",
	"finanzas",
	"contabilidad_colombia",
	"contabilidad_colombia_avanzada",
	"centros_costo",
	"cierre_fiscal",
	"activos_fijos_niif_fiscal",
	"declaraciones_tributarias",
	"clientes",
	"crm_unificado",
	"compras",
	"facturacion",
	"facturacion_ecuador",
	"facturacion_panama",
	"seguridad",
	"venta_publica",
	"reservas_hotel",
	"chat_tareas",
	"gimnasio",
	"taxi_system",
	"domicilios",
	"parqueadero",
	"apartamentos_turisticos",
	"propiedad_horizontal",
	"alquileres",
	"odontologia",
	"turnos_atencion",
	"control_electrico",
	"energia_solar",
	"camaras",
	"grafologia",
	"carnets",
	"horarios_trabajadores",
	"asistencia_empleados",
	"vehiculos_registro",
	"hoja_vida_operativa",
	"ubicacion_gps",
	"produccion_mrp",
	"logistica_wms",
	"tesoreria_presupuesto",
	"nomina_sueldos",
	"importaciones_costeo",
	"aiu_construccion",
	"cobranza",
	"reportes",
	"portal_contador",
	"portal_terceros_certificados",
	"soportes_compras_ia",
	"bancos_pagos",
	"gestion_documental",
	"cumplimiento_kyc",
	"contratos_obligaciones",
	"calidad_procesos",
	"drogueria_farmacia",
	"auditoria",
	"backups",
	"documentos_onlyoffice",
}

// PowerfulSystemEmpresaLicenseModules devuelve los modulos empresariales que la
// empresa interna necesita para operar como cualquier empresa real del sistema.
func PowerfulSystemEmpresaLicenseModules() string {
	return strings.Join(powerfulSystemEmpresaFullModules, ",")
}

// IsPowerfulSystemEmpresaName reconoce la empresa interna del sistema sin crear duplicados
// por diferencias de mayusculas, espacios o el nombre historico escrito con "Systen".
func IsPowerfulSystemEmpresaName(name string) bool {
	normalized := normalizePowerfulSystemEmpresaName(name)
	switch normalized {
	case "powerful control system", "powerful control systen":
		return true
	default:
		return false
	}
}

// IsPowerfulSystemEmpresa retorna true si el registro corresponde a la empresa interna del SaaS.
func IsPowerfulSystemEmpresa(empresa *Empresa) bool {
	if empresa == nil {
		return false
	}
	if IsPowerfulSystemEmpresaName(empresa.Nombre) {
		return true
	}
	obs := strings.ToLower(strings.TrimSpace(empresa.Observaciones))
	return strings.Contains(obs, "empresa interna del saas") && strings.Contains(obs, "powerful control system")
}

// EnsurePowerfulSystemEmpresa resuelve la empresa emisora interna usada para facturar
// licencias. Primero respeta la configuracion existente, luego busca la empresa ya
// creada por nombre y solo crea una si no existe ninguna coincidencia.
func EnsurePowerfulSystemEmpresa(dbEmp, dbSuper *sql.DB) (*Empresa, error) {
	if dbEmp == nil {
		return nil, fmt.Errorf("base de datos de empresas no disponible")
	}

	if empresa, err := getPowerfulSystemEmpresaFromConfig(dbEmp, dbSuper); err != nil {
		return nil, err
	} else if empresa != nil {
		if err := EnsurePowerfulSystemEmpresaPerpetualLicense(dbSuper, empresa.EmpresaID, powerfulSystemEmpresaUsuarioCreador); err != nil {
			return nil, err
		}
		return empresa, nil
	}

	empresa, err := findPowerfulSystemEmpresaByName(dbEmp)
	if err != nil {
		return nil, err
	}
	if empresa == nil {
		id, _, createErr := CreateEmpresaIdempotente(
			dbEmp,
			0,
			"Sistema",
			PowerfulSystemEmpresaName,
			"",
			powerfulSystemEmpresaObservacionMarca,
			powerfulSystemEmpresaUsuarioCreador,
		)
		if createErr != nil {
			return nil, createErr
		}
		empresa, err = GetEmpresaByScopeID(dbEmp, id)
		if err != nil {
			return nil, err
		}
	}
	if empresa == nil {
		return nil, fmt.Errorf("no se pudo resolver la empresa interna Powerful Control System")
	}

	if err := savePowerfulSystemEmpresaConfig(dbSuper, empresa.EmpresaID); err != nil {
		return nil, err
	}
	if err := EnsurePowerfulSystemEmpresaPerpetualLicense(dbSuper, empresa.EmpresaID, powerfulSystemEmpresaUsuarioCreador); err != nil {
		return nil, err
	}
	return empresa, nil
}

// EnsurePowerfulSystemEmpresaPerpetualLicense mantiene una licencia interna activa
// de 100 anos para que la empresa del sistema opere como empresa normal sin
// depender de renovaciones comerciales mensuales.
func EnsurePowerfulSystemEmpresaPerpetualLicense(dbSuper *sql.DB, empresaID int64, usuario string) error {
	if dbSuper == nil || empresaID <= 0 {
		return nil
	}
	if strings.TrimSpace(usuario) == "" {
		usuario = powerfulSystemEmpresaUsuarioCreador
	}
	if err := EnsureLicenciasSchema(dbSuper); err != nil {
		return err
	}

	var id int64
	err := queryRowSQLCompat(dbSuper, `SELECT id
		FROM licencias
		WHERE empresa_id = ? AND codigo_funcion = ?
		ORDER BY id ASC
		LIMIT 1`, empresaID, PowerfulSystemEmpresaLicenseCode).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	nowExpr := sqlNowExpr()
	fechaFin := time.Now().AddDate(100, 0, 0).Format("2006-01-02 23:59:59")
	modulos := PowerfulSystemEmpresaLicenseModules()
	if id > 0 {
		_, err = execSQLCompat(dbSuper, `UPDATE licencias
			SET tipo_id = 0,
				pais_codigo = 'CO',
				nombre = ?,
				descripcion = ?,
				valor = 0,
				duracion_dias = ?,
				max_documentos_mensuales = 999999,
				max_cajas_simultaneas = 99,
				modulos_habilitados = ?,
				es_adicional = 0,
				super_rol_habilitado = 0,
				fecha_inicio = CASE WHEN fecha_inicio IS NULL THEN `+nowExpr+` ELSE fecha_inicio END,
				fecha_fin = ?,
				activo = 1,
				fecha_actualizacion = `+nowExpr+`,
				usuario_creador = CASE WHEN COALESCE(usuario_creador, '') = '' THEN ? ELSE usuario_creador END,
				estado = 'activo',
				observaciones = ?
			WHERE id = ?`,
			"Licencia interna 100 anos Powerful Control System",
			"Licencia tecnica interna de 100 anos para que la empresa emisora del SaaS facture licencias y opere modulos internos como una empresa normal.",
			PowerfulSystemEmpresaLicenseDays,
			modulos,
			fechaFin,
			usuario,
			"Licencia interna del sistema por 100 anos; no se ofrece en el catalogo comercial.",
			id,
		)
		if err != nil {
			return err
		}
		InvalidateLicenciaPermisoPolicyCacheForEmpresa(empresaID)
		return nil
	}

	_, err = insertSQLCompat(dbSuper, `INSERT INTO licencias (
		empresa_id, tipo_id, pais_codigo, nombre, descripcion, valor, duracion_dias,
		max_documentos_mensuales, max_cajas_simultaneas, modulos_habilitados,
		es_adicional, codigo_funcion, super_rol_habilitado, fecha_inicio, fecha_fin,
		activo, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
	) VALUES (?, 0, 'CO', ?, ?, 0, ?, 999999, 99, ?, 0, ?, 0, `+nowExpr+`, ?, 1, `+nowExpr+`, `+nowExpr+`, ?, 'activo', ?)`,
		empresaID,
		"Licencia interna 100 anos Powerful Control System",
		"Licencia tecnica interna de 100 anos para que la empresa emisora del SaaS facture licencias y opere modulos internos como una empresa normal.",
		PowerfulSystemEmpresaLicenseDays,
		modulos,
		PowerfulSystemEmpresaLicenseCode,
		fechaFin,
		usuario,
		"Licencia interna del sistema por 100 anos; no se ofrece en el catalogo comercial.",
	)
	if err != nil {
		return err
	}
	InvalidateLicenciaPermisoPolicyCacheForEmpresa(empresaID)
	return nil
}

func getPowerfulSystemEmpresaFromConfig(dbEmp, dbSuper *sql.DB) (*Empresa, error) {
	if dbSuper == nil {
		return nil, nil
	}
	raw, _, _, _, err := GetConfigEntry(dbSuper, PowerfulSystemEmpresaConfigKey)
	if err != nil {
		if isMissingTableError(err) {
			return nil, nil
		}
		return nil, err
	}
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return nil, nil
	}
	empresa, err := GetEmpresaByScopeID(dbEmp, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if empresa == nil {
		return nil, nil
	}
	if !IsPowerfulSystemEmpresa(empresa) {
		return nil, nil
	}
	return empresa, nil
}

func savePowerfulSystemEmpresaConfig(dbSuper *sql.DB, empresaID int64) error {
	if dbSuper == nil || empresaID <= 0 {
		return nil
	}
	if err := SetConfigValue(dbSuper, PowerfulSystemEmpresaConfigKey, strconv.FormatInt(empresaID, 10), false); err != nil {
		if isMissingTableError(err) {
			return nil
		}
		return err
	}
	return nil
}

func findPowerfulSystemEmpresaByName(dbEmp *sql.DB) (*Empresa, error) {
	if dbEmp == nil {
		return nil, nil
	}
	candidates := []string{PowerfulSystemEmpresaName, PowerfulSystemEmpresaNameLegacyTypo}
	for _, candidate := range candidates {
		var id int64
		err := queryRowSQLCompat(dbEmp, `SELECT id
			FROM empresas
			WHERE LOWER(TRIM(COALESCE(nombre, ''))) = LOWER(TRIM(?))
			ORDER BY id ASC
			LIMIT 1`, candidate).Scan(&id)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return nil, err
		}
		if id > 0 {
			return GetEmpresaByScopeID(dbEmp, id)
		}
	}

	rows, err := dbEmp.Query(`SELECT id, COALESCE(nombre, '')
		FROM empresas
		WHERE LOWER(COALESCE(nombre, '')) LIKE '%powerful%control%syste%'
		ORDER BY id ASC`)
	if err != nil {
		if isMissingTableError(err) {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var nombre string
		if err := rows.Scan(&id, &nombre); err != nil {
			return nil, err
		}
		if id > 0 && IsPowerfulSystemEmpresaName(nombre) {
			return GetEmpresaByScopeID(dbEmp, id)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return nil, nil
}

func normalizePowerfulSystemEmpresaName(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" {
		return ""
	}
	var b strings.Builder
	lastWasSpace := true
	for _, r := range raw {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastWasSpace = false
			continue
		}
		if !lastWasSpace {
			b.WriteRune(' ')
			lastWasSpace = true
		}
	}
	return strings.TrimSpace(b.String())
}
