package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type AlquileresLicenciaPlan struct {
	Nombre                 string
	Descripcion            string
	Valor                  float64
	DuracionDias           int
	MaxDocumentosMensuales int
	ModulosHabilitados     string
}

func DefaultAlquileresLicenciaModules() string {
	modules := []string{
		"ventas",
		"inventario",
		"compras",
		"logistica_wms",
		"finanzas",
		"contabilidad_colombia",
		"bancos_pagos",
		"tesoreria_presupuesto",
		"cobranza",
		"gestion_documental",
		"contratos_obligaciones",
		"calidad_procesos",
		"alquileres",
		"clientes",
		"facturacion",
		"seguridad",
	}
	return strings.Join(modules, ",")
}

func DefaultAlquileresLicenciaPlans() []AlquileresLicenciaPlan {
	modules := DefaultAlquileresLicenciaModules()
	plans := DefaultGlobalLicenciaPlans()
	out := make([]AlquileresLicenciaPlan, 0, len(plans))
	for _, plan := range plans {
		nombre := plan.Nombre
		descripcion := plan.Descripcion
		if plan.Valor == 0 && plan.DuracionDias == 15 {
			nombre = "Alquileres prueba 15 dias"
			descripcion = "Licencia de prueba para alquiler universal: herramientas, motos, maquinaria, objetos, contratos, garantias y devoluciones."
		}
		out = append(out, AlquileresLicenciaPlan{
			Nombre:                 nombre,
			Descripcion:            descripcion,
			Valor:                  plan.Valor,
			DuracionDias:           plan.DuracionDias,
			MaxDocumentosMensuales: plan.MaxDocumentosMensuales,
			ModulosHabilitados:     modules,
		})
	}
	return out
}

func EnsureAlquileresTipoEmpresaYLicencias(dbConn *sql.DB, usuario string) (tipoID int64, licenciasAseguradas int, err error) {
	if dbConn == nil {
		return 0, 0, errors.New("db connection is nil")
	}
	if err := EnsureLicenciasSchema(dbConn); err != nil {
		return 0, 0, err
	}
	if err := EnsureCanonicalTiposEmpresaPreconfigurables(dbConn); err != nil {
		return 0, 0, err
	}
	tipoID, err = ensureAlquileresTipoEmpresa(dbConn)
	if err != nil {
		return 0, 0, err
	}
	licenciasAseguradas, err = EnsureLicenciasCatalogoGlobal(dbConn, usuario)
	if err != nil {
		return tipoID, licenciasAseguradas, err
	}
	return tipoID, licenciasAseguradas, nil
}

func ensureAlquileresTipoEmpresa(dbConn *sql.DB) (int64, error) {
	tipos, err := GetTiposEmpresas(dbConn)
	if err != nil {
		return 0, err
	}
	for _, tipo := range tipos {
		if isTipoEmpresaAlquilerObjetos(tipo.Nombre) {
			if strings.EqualFold(strings.TrimSpace(tipo.Estado), "inactivo") {
				if err := SetTipoEmpresaActivo(dbConn, tipo.ID, "activo"); err != nil {
					return 0, err
				}
			}
			return tipo.ID, nil
		}
	}
	return CreateTipoEmpresa(dbConn, "Alquileres de herramientas, motos y objetos", "Herramientas, motos, maquinaria, mobiliario, objetos, contratos, garantias, devoluciones y mantenimiento.")
}

func ensureAlquileresLicenciaPlan(dbConn *sql.DB, tipoID int64, usuario string, plan AlquileresLicenciaPlan) error {
	if tipoID <= 0 {
		return fmt.Errorf("tipo_id alquileres invalido")
	}
	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = "sistema.alquileres"
	}
	nowExpr := sqlNowExpr()
	var existingID int64
	err := queryRowSQLCompat(dbConn, `SELECT id
		FROM licencias
		WHERE tipo_id = ?
			AND COALESCE(empresa_id, 0) = 0
			AND LOWER(TRIM(COALESCE(nombre, ''))) = LOWER(TRIM(?))
		LIMIT 1`, tipoID, plan.Nombre).Scan(&existingID)
	if err == nil && existingID > 0 {
		_, err = execSQLCompat(dbConn, `UPDATE licencias
			SET pais_codigo = 'CO',
				descripcion = ?,
				valor = ?,
				duracion_dias = ?,
				max_documentos_mensuales = ?,
				max_cajas_simultaneas = ?,
				modulos_habilitados = ?,
				es_adicional = 0,
				codigo_funcion = '',
				super_rol_habilitado = 0,
				activo = 1,
				estado = 'activo',
				usuario_creador = COALESCE(NULLIF(TRIM(usuario_creador), ''), ?),
				fecha_actualizacion = `+nowExpr+`
			WHERE id = ?`, plan.Descripcion, plan.Valor, plan.DuracionDias, plan.MaxDocumentosMensuales, DefaultLicenciaMaxCajasSimultaneas(plan.MaxDocumentosMensuales), plan.ModulosHabilitados, usuario, existingID)
		return err
	}
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	id, err := CreateLicenciaAdvancedWithLimits(dbConn, tipoID, "CO", plan.Nombre, plan.Descripcion, plan.Valor, plan.DuracionDias, plan.ModulosHabilitados, 0, "", 0, plan.MaxDocumentosMensuales)
	if err != nil {
		return err
	}
	_, _ = execSQLCompat(dbConn, "UPDATE licencias SET usuario_creador = COALESCE(NULLIF(TRIM(usuario_creador), ''), ?), fecha_actualizacion = "+nowExpr+" WHERE id = ?", usuario, id)
	return nil
}
