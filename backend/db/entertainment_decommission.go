package db

import "database/sql"

// DecommissionRemovedEntertainmentArtifacts removes data owned exclusively by
// the retired games and emulator module. It is intentionally idempotent so a
// restart cannot fail when the table was already removed.
func DecommissionRemovedEntertainmentArtifacts(dbSuper *sql.DB) error {
	if dbSuper == nil {
		return nil
	}
	_, err := execSQLCompat(dbSuper, `DROP TABLE IF EXISTS super_juegos_records`)
	return err
}
