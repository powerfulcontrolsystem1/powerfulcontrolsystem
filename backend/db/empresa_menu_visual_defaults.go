package db

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

const EmpresaMenuVisualDefaultConfigJSON = `{"version":3,"enabled":true,"hidden_links":["linkImportacionesCosteo","linkProduccionMRP","linkLogisticaWMS","linkCRMComercial","linkUsuarios","linkClientes","linkPortalUsuarios","linkMiHorario","linkHorariosTrabajadores","linkAsistenciaEmpleados","linkCarnets","linkVehiculosRegistro","linkHojaVidaOperativa","linkLicenciaSistema"]}`

const empresaMenuVisualDefaultsSQL = `INSERT INTO empresa_estacion_prefs (
 empresa_id, estacion_id, clave, valor, fecha_creacion, fecha_actualizacion,
 usuario_creador, estado, observaciones
)
SELECT DISTINCT COALESCE(NULLIF(e.empresa_id, 0), e.id), 0, 'menu_visual_config', $1,
 CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'sistema.menu_visual_predeterminado',
 'activo', 'Seleccion predeterminada de menus empresariales 2026-09-14'
FROM empresas e
WHERE COALESCE(NULLIF(e.empresa_id, 0), e.id) > 0
ON CONFLICT (empresa_id, estacion_id, clave) DO UPDATE SET
 valor = EXCLUDED.valor,
 fecha_actualizacion = CURRENT_TIMESTAMP,
 usuario_creador = EXCLUDED.usuario_creador,
 estado = 'activo',
 observaciones = EXCLUDED.observaciones`

const empresaMenuVisualDefaultsFingerprint = empresaMenuVisualDefaultsSQL + "\n-- default_config=" + EmpresaMenuVisualDefaultConfigJSON

func applyEmpresaMenuVisualDefaultsTx(ctx context.Context, tx *sql.Tx) error {
	if tx == nil {
		return errors.New("migration transaction is required")
	}
	_, err := tx.ExecContext(ctx, empresaMenuVisualDefaultsSQL, EmpresaMenuVisualDefaultConfigJSON)
	return err
}

// EnsureEmpresaMenuVisualDefault creates the company preference without
// overwriting a later selection made from Configuracion > Menu visible.
func EnsureEmpresaMenuVisualDefault(dbConn *sql.DB, empresaID int64, usuario string) error {
	if dbConn == nil || empresaID <= 0 {
		return nil
	}
	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = "sistema.menu_visual_predeterminado"
	}
	_, err := execSQLCompat(dbConn, `INSERT INTO empresa_estacion_prefs (
		empresa_id, estacion_id, clave, valor, fecha_creacion, fecha_actualizacion,
		usuario_creador, estado, observaciones
	) VALUES (?, 0, 'menu_visual_config', ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, 'activo', ?)
	ON CONFLICT (empresa_id, estacion_id, clave) DO NOTHING`,
		empresaID,
		EmpresaMenuVisualDefaultConfigJSON,
		usuario,
		"Seleccion predeterminada de menus empresariales 2026-09-14",
	)
	return err
}
