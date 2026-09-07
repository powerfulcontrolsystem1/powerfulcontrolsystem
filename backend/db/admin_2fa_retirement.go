package db

import (
	"context"
	"database/sql"
)

const admin2FARetirementFingerprint = "retire administrator TOTP: clear authenticators and recovery codes, remove policy config, revoke active sessions"

// applyAdmin2FARetirementTx removes all usable administrator second-factor
// material. The now-inert legacy columns and table remain temporarily so a
// rollback to the preceding binary does not fail on missing schema objects.
func applyAdmin2FARetirementTx(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, `
		DO $retire_admin_2fa$
		BEGIN
			IF EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_schema = current_schema()
				  AND table_name = 'administradores'
				  AND column_name = 'totp_secret'
			) THEN
				EXECUTE 'UPDATE administradores
					SET totp_enabled = 0,
						totp_secret = '''',
						totp_confirmado_en = '''',
						totp_last_counter = -1';
			END IF;
			IF to_regclass(current_schema() || '.administrador_totp_recovery_codes') IS NOT NULL THEN
				EXECUTE 'DELETE FROM administrador_totp_recovery_codes';
			END IF;
		END
		$retire_admin_2fa$
	`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM configuraciones
		WHERE config_key IN (
			'security.admin_2fa.enabled',
			'security.admin_2fa.enabled.updated_by'
		)
	`); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `
		UPDATE sesiones
		SET activo = 0,
			fecha_fin = CAST(CURRENT_TIMESTAMP AS TEXT)
		WHERE activo = 1
	`)
	return err
}
