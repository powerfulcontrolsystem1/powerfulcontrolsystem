package db

import (
	"testing"
	"time"
)

func TestAdministradorLoginFailuresLockAndExpire(t *testing.T) {
	now := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	state := AdministradorLoginAttemptState{}
	for attempt := 1; attempt <= AdminLoginMaxAttempts; attempt++ {
		state = nextAdministradorLoginAttemptState(state, now.Add(time.Duration(attempt-1)*time.Second), AdminLoginMaxAttempts, AdminLoginWindow, AdminLoginLock)
		if state.FailedAttempts != attempt {
			t.Fatalf("attempt %d produced count %d", attempt, state.FailedAttempts)
		}
	}
	if !state.Locked(now.Add(10 * time.Second)) {
		t.Fatal("the fifth failed attempt must lock the administrator login")
	}
	afterLock := state.LockedUntil.Add(time.Second)
	state = nextAdministradorLoginAttemptState(state, afterLock, AdminLoginMaxAttempts, AdminLoginWindow, AdminLoginLock)
	if state.FailedAttempts != 1 || state.Locked(afterLock) {
		t.Fatalf("expired lock must start a fresh window: %#v", state)
	}
}

func TestAdministradorLoginFailureWindowResets(t *testing.T) {
	now := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	state := nextAdministradorLoginAttemptState(AdministradorLoginAttemptState{}, now, 5, time.Minute, time.Minute)
	state = nextAdministradorLoginAttemptState(state, now.Add(2*time.Minute), 5, time.Minute, time.Minute)
	if state.FailedAttempts != 1 || !state.WindowStarted.Equal(now.Add(2*time.Minute)) {
		t.Fatalf("expired failure window was not reset: %#v", state)
	}
}
