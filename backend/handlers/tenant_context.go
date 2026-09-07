package handlers

import (
	"context"
	"net/http"
	"strings"
)

type tenantContextKey struct{}

// TenantContext is the validated multiempresa identity propagated after the
// authorization boundary. Handlers must use EmpresaID from this context as the
// source of authority instead of trusting query, header or JSON values.
type TenantContext struct {
	EmpresaID     int64
	AdminEmail    string
	AdminRole     string
	EffectiveRole string
	Module        string
	Action        string
}

func TenantContextFromRequest(r *http.Request) (TenantContext, bool) {
	if r == nil {
		return TenantContext{}, false
	}
	tenant, ok := r.Context().Value(tenantContextKey{}).(TenantContext)
	return tenant, ok && tenant.EmpresaID > 0
}

func requestWithTenantContext(r *http.Request, tenant TenantContext) *http.Request {
	if r == nil || tenant.EmpresaID <= 0 {
		return r
	}
	tenant.AdminEmail = strings.ToLower(strings.TrimSpace(tenant.AdminEmail))
	tenant.AdminRole = strings.TrimSpace(tenant.AdminRole)
	tenant.EffectiveRole = strings.TrimSpace(tenant.EffectiveRole)
	tenant.Module = strings.TrimSpace(tenant.Module)
	tenant.Action = strings.TrimSpace(tenant.Action)
	ctx := context.WithValue(r.Context(), tenantContextKey{}, tenant)
	// Transitional compatibility for handlers that still consume legacy keys.
	ctx = context.WithValue(ctx, "empresaID", tenant.EmpresaID)
	if tenant.AdminRole != "" {
		ctx = context.WithValue(ctx, "adminRole", tenant.AdminRole)
	}
	if tenant.EffectiveRole != "" {
		ctx = context.WithValue(ctx, "adminRoleEfectivo", tenant.EffectiveRole)
	}
	return r.WithContext(ctx)
}
