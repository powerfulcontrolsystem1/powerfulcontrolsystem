package db

import (
	"context"
	"database/sql"
)

const portalVisitasSchemaFingerprint = "portal-visitas-paises:v1:pais-fecha-contador"

// applyPortalVisitasSchemaTx creates the public portal counter from the
// migration role. The public HTTP handler only reads or increments this table.
func applyPortalVisitasSchemaTx(_ context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS portal_visitas_paises (
		pais_codigo TEXT NOT NULL,
		fecha DATE NOT NULL DEFAULT CURRENT_DATE,
		visitas BIGINT NOT NULL DEFAULT 0,
		actualizado_en TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		PRIMARY KEY (pais_codigo, fecha)
	)`)
	return err
}
