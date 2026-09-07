// Package httpguard centraliza contratos HTTP pequenos compartidos por los
// procesos API y worker.
package httpguard

import "net/http"

// AllowGetOrHead rechaza otros metodos y publica el encabezado Allow correcto.
func AllowGetOrHead(w http.ResponseWriter, r *http.Request) bool {
	if r != nil && (r.Method == http.MethodGet || r.Method == http.MethodHead) {
		return true
	}
	w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	return false
}
