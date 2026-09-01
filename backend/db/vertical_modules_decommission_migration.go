package db

import (
	"context"
	"database/sql"
)

const empresaVerticalModulesDecommissionFingerprint = `
ALTER TABLE IF EXISTS empresa_venta_publica_ordenes DROP COLUMN IF EXISTS taxi_request_id;

DROP TABLE IF EXISTS empresa_taxi_route_points CASCADE;
DROP TABLE IF EXISTS empresa_taxi_offers CASCADE;
DROP TABLE IF EXISTS empresa_taxi_requests CASCADE;
DROP TABLE IF EXISTS empresa_taxi_customers CASCADE;
DROP TABLE IF EXISTS empresa_taxi_drivers CASCADE;
DROP TABLE IF EXISTS empresa_taxi_config CASCADE;

DROP TABLE IF EXISTS empresa_apartamentos_turisticos_tareas CASCADE;
DROP TABLE IF EXISTS empresa_apartamentos_turisticos_reservas CASCADE;
DROP TABLE IF EXISTS empresa_apartamentos_turisticos_tarifas CASCADE;
DROP TABLE IF EXISTS empresa_apartamentos_turisticos_unidades CASCADE;
DROP TABLE IF EXISTS empresa_apartamentos_turisticos_config CASCADE;

DROP TABLE IF EXISTS empresa_propiedad_horizontal_asambleas CASCADE;
DROP TABLE IF EXISTS empresa_propiedad_horizontal_pqrs CASCADE;
DROP TABLE IF EXISTS empresa_propiedad_horizontal_recaudos CASCADE;
DROP TABLE IF EXISTS empresa_propiedad_horizontal_cargos CASCADE;
DROP TABLE IF EXISTS empresa_propiedad_horizontal_personas CASCADE;
DROP TABLE IF EXISTS empresa_propiedad_horizontal_unidades CASCADE;
DROP TABLE IF EXISTS empresa_propiedad_horizontal_config CASCADE;

DROP TABLE IF EXISTS empresa_gimnasio_eventos_acceso CASCADE;
DROP TABLE IF EXISTS empresa_gimnasio_dispositivos_acceso CASCADE;
DROP TABLE IF EXISTS empresa_gimnasio_credenciales CASCADE;
DROP TABLE IF EXISTS empresa_gimnasio_acceso_config CASCADE;
DROP TABLE IF EXISTS empresa_gimnasio_pagos CASCADE;
DROP TABLE IF EXISTS empresa_gimnasio_asistencias CASCADE;
DROP TABLE IF EXISTS empresa_gimnasio_inscripciones CASCADE;
DROP TABLE IF EXISTS empresa_gimnasio_clases CASCADE;
DROP TABLE IF EXISTS empresa_gimnasio_entrenadores CASCADE;
DROP TABLE IF EXISTS empresa_gimnasio_socios CASCADE;
DROP TABLE IF EXISTS empresa_gimnasio_planes CASCADE;

DROP TABLE IF EXISTS empresa_odontologia_pagos CASCADE;
DROP TABLE IF EXISTS empresa_odontologia_presupuestos CASCADE;
DROP TABLE IF EXISTS empresa_odontologia_tratamientos CASCADE;
DROP TABLE IF EXISTS empresa_odontologia_odontogramas CASCADE;
DROP TABLE IF EXISTS empresa_odontologia_historias CASCADE;
DROP TABLE IF EXISTS empresa_odontologia_citas CASCADE;
DROP TABLE IF EXISTS empresa_odontologia_consultorios CASCADE;
DROP TABLE IF EXISTS empresa_odontologia_profesionales CASCADE;
DROP TABLE IF EXISTS empresa_odontologia_pacientes CASCADE;

DO $$
BEGIN
	IF to_regclass('empresa_modulos_colombia_eventos') IS NOT NULL THEN
		DELETE FROM empresa_modulos_colombia_eventos WHERE modulo = 'drogueria_farmacia';
	END IF;
	IF to_regclass('empresa_modulos_colombia_evidencias') IS NOT NULL THEN
		DELETE FROM empresa_modulos_colombia_evidencias WHERE modulo = 'drogueria_farmacia';
	END IF;
	IF to_regclass('empresa_modulos_colombia_aprobaciones') IS NOT NULL THEN
		DELETE FROM empresa_modulos_colombia_aprobaciones WHERE modulo = 'drogueria_farmacia';
	END IF;
	IF to_regclass('empresa_modulos_colombia_tareas') IS NOT NULL THEN
		DELETE FROM empresa_modulos_colombia_tareas WHERE modulo = 'drogueria_farmacia';
	END IF;
	IF to_regclass('empresa_modulos_colombia_registros') IS NOT NULL THEN
		DELETE FROM empresa_modulos_colombia_registros WHERE modulo = 'drogueria_farmacia';
	END IF;
END $$;`

func applyEmpresaVerticalModulesDecommissionTx(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, empresaVerticalModulesDecommissionFingerprint)
	return err
}
