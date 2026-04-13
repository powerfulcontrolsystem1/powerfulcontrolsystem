package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

// AsesoresHandler maneja CRUD para vendedores/asesores (alias vendedor_de_licencia)
func AsesoresHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			q := r.URL.Query()
			// opcional: soporte ?id= para obtener uno
			idStr := q.Get("id")
			if idStr != "" {
				id, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil {
					http.Error(w, "invalid id", http.StatusBadRequest)
					return
				}
				// buscar en la lista por id
				lista, err := dbpkg.ListAsesores(dbSuper)
				if err != nil {
					http.Error(w, "failed to query vendedores: "+err.Error(), http.StatusInternalServerError)
					return
				}
				for _, a := range lista {
					if a.ID == id {
						w.Header().Set("Content-Type", "application/json")
						json.NewEncoder(w).Encode(a)
						return
					}
				}
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			lista, err := dbpkg.ListAsesores(dbSuper)
			if err != nil {
				http.Error(w, "failed to query vendedores: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(lista)
			return
		case http.MethodPost:
			var payload struct{
				Email string `json:"email"`
				Nombre string `json:"nombre"`
				Rol string `json:"rol"`
				Notas string `json:"notas"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.Email) == "" {
				http.Error(w, "email required", http.StatusBadRequest)
				return
			}
			id, err := dbpkg.CreateAsesor(dbSuper, payload.Email, payload.Nombre, payload.Rol, payload.Notas)
			if err != nil {
				http.Error(w, "failed to create vendedor: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
			return
		case http.MethodPut:
			q := r.URL.Query()
			idStr := q.Get("id")
			if idStr == "" {
				http.Error(w, "id required", http.StatusBadRequest)
				return
			}
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				http.Error(w, "invalid id", http.StatusBadRequest)
				return
			}
			var payload struct{
				Email string `json:"email"`
				Nombre string `json:"nombre"`
				Rol string `json:"rol"`
				Notas string `json:"notas"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if err := dbpkg.UpdateAsesor(dbSuper, id, payload.Email, payload.Nombre, payload.Rol, payload.Notas); err != nil {
				http.Error(w, "failed to update vendedor: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		case http.MethodDelete:
			q := r.URL.Query()
			idStr := q.Get("id")
			if idStr == "" {
				http.Error(w, "id required", http.StatusBadRequest)
				return
			}
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				http.Error(w, "invalid id", http.StatusBadRequest)
				return
			}
			if err := dbpkg.DeleteAsesor(dbSuper, id); err != nil {
				http.Error(w, "failed to delete vendedor: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

// AsesorComercialHandler maneja planes de comision para vendedores (vendedor_de_licencia)
func AsesorComercialHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			q := r.URL.Query()
			asesorID := q.Get("asesor_id")
			if asesorID != "" {
				plan, err := dbpkg.GetAsesorComercialPlanByAsesorID(dbSuper, asesorID)
				if err != nil {
					http.Error(w, "failed to query plan: "+err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(plan)
				return
			}
			plans, err := dbpkg.ListAsesorComercialPlans(dbSuper)
			if err != nil {
				http.Error(w, "failed to query plans: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(plans)
			return
		case http.MethodPost:
			var payload struct{
				AsesorID string `json:"asesor_id"`
				VendedorID string `json:"vendedor_id,omitempty"`
				AsesorEmail string `json:"asesor_email"`
				EmpresaID int64 `json:"empresa_id,omitempty"`
				ComisionVentaPct float64 `json:"comision_venta_pct"`
				ComisionPagoPct float64 `json:"comision_pago_pct"`
				MesesRenovacion int `json:"meses_renovacion"`
				Notas string `json:"notas"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			// aceptar alias vendedor_id
			if strings.TrimSpace(payload.AsesorID) == "" && strings.TrimSpace(payload.VendedorID) != "" {
				payload.AsesorID = payload.VendedorID
			}
			if strings.TrimSpace(payload.AsesorID) == "" {
				http.Error(w, "asesor_id (o vendedor_id) required", http.StatusBadRequest)
				return
			}
			id, err := dbpkg.CreateAsesorComercialPlan(dbSuper, payload.AsesorID, payload.AsesorEmail, payload.EmpresaID, payload.ComisionVentaPct, payload.ComisionPagoPct, payload.MesesRenovacion, payload.Notas)
			if err != nil {
				http.Error(w, "failed to create plan: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
			return
		case http.MethodPut:
			q := r.URL.Query()
			idStr := q.Get("id")
			if idStr == "" {
				http.Error(w, "id required", http.StatusBadRequest)
				return
			}
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				http.Error(w, "invalid id", http.StatusBadRequest)
				return
			}
			var payload struct{
				ComisionVentaPct float64 `json:"comision_venta_pct"`
				ComisionPagoPct float64 `json:"comision_pago_pct"`
				MesesRenovacion int `json:"meses_renovacion"`
				Notas string `json:"notas"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if err := dbpkg.UpdateAsesorComercialPlan(dbSuper, id, payload.ComisionVentaPct, payload.ComisionPagoPct, payload.MesesRenovacion, payload.Notas); err != nil {
				http.Error(w, "failed to update plan: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		case http.MethodDelete:
			q := r.URL.Query()
			idStr := q.Get("id")
			if idStr == "" {
				http.Error(w, "id required", http.StatusBadRequest)
				return
			}
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				http.Error(w, "invalid id", http.StatusBadRequest)
				return
			}
			if err := dbpkg.DeleteAsesorComercialPlan(dbSuper, id); err != nil {
				http.Error(w, "failed to delete plan: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}
