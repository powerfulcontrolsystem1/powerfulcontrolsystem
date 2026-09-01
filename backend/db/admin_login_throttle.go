package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const (
	AdminLoginMaxAttempts = 5
	AdminLoginWindow      = 15 * time.Minute
	AdminLoginLock        = 15 * time.Minute
)

type AdministradorLoginAttemptState struct {
	FailedAttempts int
	WindowStarted  time.Time
	LockedUntil    time.Time
}

func (state AdministradorLoginAttemptState) Locked(now time.Time) bool {
	return !state.LockedUntil.IsZero() && now.Before(state.LockedUntil)
}

func GetAdministradorLoginAttemptState(dbConn *sql.DB, email string) (AdministradorLoginAttemptState, error) {
	if dbConn == nil || strings.TrimSpace(email) == "" {
		return AdministradorLoginAttemptState{}, fmt.Errorf("administrator email is required")
	}
	var state AdministradorLoginAttemptState
	err := queryRowSQLCompat(dbConn, `SELECT failed_attempts, window_started_at, locked_until
		FROM administrador_login_intentos
		WHERE LOWER(email) = LOWER(?)`, strings.TrimSpace(email)).Scan(
		&state.FailedAttempts,
		&state.WindowStarted,
		&state.LockedUntil,
	)
	if err == sql.ErrNoRows {
		return AdministradorLoginAttemptState{}, nil
	}
	return state, err
}

func nextAdministradorLoginAttemptState(state AdministradorLoginAttemptState, now time.Time, maxAttempts int, window, lockDuration time.Duration) AdministradorLoginAttemptState {
	if maxAttempts <= 0 {
		maxAttempts = AdminLoginMaxAttempts
	}
	if window <= 0 {
		window = AdminLoginWindow
	}
	if lockDuration <= 0 {
		lockDuration = AdminLoginLock
	}
	if state.Locked(now) {
		return state
	}
	if state.WindowStarted.IsZero() || now.Sub(state.WindowStarted) > window {
		state.FailedAttempts = 0
		state.WindowStarted = now
	}
	state.FailedAttempts++
	if state.FailedAttempts >= maxAttempts {
		state.LockedUntil = now.Add(lockDuration)
	}
	return state
}

func RegisterAdministradorLoginFailure(dbConn *sql.DB, email string, now time.Time) (AdministradorLoginAttemptState, error) {
	if dbConn == nil || strings.TrimSpace(email) == "" {
		return AdministradorLoginAttemptState{}, fmt.Errorf("administrator email is required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	email = strings.ToLower(strings.TrimSpace(email))
	tx, err := dbConn.Begin()
	if err != nil {
		return AdministradorLoginAttemptState{}, err
	}
	defer rollbackTransaction(tx)
	if _, err := execTxSQLCompat(tx, `INSERT INTO administrador_login_intentos
		(email, failed_attempts, window_started_at, locked_until, updated_at)
		VALUES (?, 0, ?, ?, ?)
		ON CONFLICT (email) DO NOTHING`, email, now, time.Unix(0, 0).UTC(), now); err != nil {
		return AdministradorLoginAttemptState{}, err
	}
	var state AdministradorLoginAttemptState
	if err := queryRowTxSQLCompat(tx, `SELECT failed_attempts, window_started_at, locked_until
		FROM administrador_login_intentos
		WHERE email = ? FOR UPDATE`, email).Scan(&state.FailedAttempts, &state.WindowStarted, &state.LockedUntil); err != nil {
		return AdministradorLoginAttemptState{}, err
	}
	state = nextAdministradorLoginAttemptState(state, now, AdminLoginMaxAttempts, AdminLoginWindow, AdminLoginLock)
	if _, err := execTxSQLCompat(tx, `UPDATE administrador_login_intentos
		SET failed_attempts = ?, window_started_at = ?, locked_until = ?, updated_at = ?
		WHERE email = ?`, state.FailedAttempts, state.WindowStarted, state.LockedUntil, now, email); err != nil {
		return AdministradorLoginAttemptState{}, err
	}
	if err := tx.Commit(); err != nil {
		return AdministradorLoginAttemptState{}, err
	}
	return state, nil
}

func ClearAdministradorLoginFailures(dbConn *sql.DB, email string) error {
	if dbConn == nil || strings.TrimSpace(email) == "" {
		return fmt.Errorf("administrator email is required")
	}
	_, err := execSQLCompat(dbConn, `DELETE FROM administrador_login_intentos WHERE LOWER(email) = LOWER(?)`, strings.TrimSpace(email))
	return err
}
