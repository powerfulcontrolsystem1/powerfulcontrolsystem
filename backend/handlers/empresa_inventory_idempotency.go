package handlers

import (
	"database/sql"
	"net/http"
	"strings"
)

// EmpresaInventoryIdempotentMutation applies the tenant-scoped durable
// idempotency ledger to a critical browser inventory mutation. It must be
// placed inside WithEmpresaInventarioPermissions so the authenticated tenant
// context exists before the request is claimed.
func EmpresaInventoryIdempotentMutation(dbEmp *sql.DB, operation string, next http.HandlerFunc) http.HandlerFunc {
	return mobileIdempotentWhenMutating(dbEmp, "inventory.web."+operation, next)
}

// EmpresaPurchasesIdempotentMutation protects browser purchase mutations with
// the same durable, tenant-scoped ledger used by inventory operations.
func EmpresaPurchasesIdempotentMutation(dbEmp *sql.DB, operation string, next http.HandlerFunc) http.HandlerFunc {
	return mobileIdempotentWhenMutating(dbEmp, "purchases.web."+operation, next)
}

// EmpresaPrinterQueueIdempotentMutation protects queue creation without
// imposing an idempotency key on read-only printer resolution or on the
// naturally idempotent configuration upserts sharing the same endpoint.
func EmpresaPrinterQueueIdempotentMutation(dbEmp *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		isQueueCreation := action == "cola_trabajo" || action == "crear_trabajo" || action == "trabajo"
		if isQueueCreation && (r.Method == http.MethodPost || r.Method == http.MethodPut) {
			mobileIdempotentMutation(dbEmp, "printers.web.queue_create", next).ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	}
}
