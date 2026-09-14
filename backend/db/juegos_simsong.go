package db

import (
	"context"
	"database/sql"
)

// Keep the historical key so retirement does not destroy personal save data.
// The handler's live catalog no longer exposes selva.
const simsongSchemaSQL = `ALTER TABLE juegos_partidas DROP CONSTRAINT juegos_partidas_juego_check;
ALTER TABLE juegos_partidas ADD CONSTRAINT juegos_partidas_juego_check
CHECK (juego IN ('pacman','tetris','buscaminas','solitario','selva','sorpresa','simsong'));`

func applySimsongSchemaTx(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, simsongSchemaSQL)
	return err
}
