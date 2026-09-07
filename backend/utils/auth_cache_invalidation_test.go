package utils

import (
	"testing"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func TestInvalidateAuthCacheForAdminRemovesEverySessionImmediately(t *testing.T) {
	authMiddlewareCacheMu.Lock()
	originalSessions := authMiddlewareSessionCache
	originalAdmins := authMiddlewareAdminCache
	authMiddlewareSessionCache = map[string]authSessionCacheEntry{
		"token-a": {Session: &dbpkg.Session{AdminEmail: "admin@example.test"}, CachedAt: time.Now()},
		"token-b": {Session: &dbpkg.Session{AdminEmail: "other@example.test"}, CachedAt: time.Now()},
	}
	authMiddlewareAdminCache = map[string]authAdminCacheEntry{
		"admin@example.test": {Admin: &dbpkg.Admin{Email: "admin@example.test"}, CachedAt: time.Now()},
	}
	authMiddlewareCacheMu.Unlock()
	defer func() {
		authMiddlewareCacheMu.Lock()
		authMiddlewareSessionCache = originalSessions
		authMiddlewareAdminCache = originalAdmins
		authMiddlewareCacheMu.Unlock()
	}()

	InvalidateAuthCacheForAdmin("ADMIN@example.test")
	authMiddlewareCacheMu.Lock()
	defer authMiddlewareCacheMu.Unlock()
	if _, ok := authMiddlewareSessionCache["token-a"]; ok {
		t.Fatal("revoked administrator session remained in cache")
	}
	if _, ok := authMiddlewareAdminCache["admin@example.test"]; ok {
		t.Fatal("administrator privilege cache remained after invalidation")
	}
	if _, ok := authMiddlewareSessionCache["token-b"]; !ok {
		t.Fatal("unrelated administrator session was removed")
	}
}
