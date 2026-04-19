package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/you/pos-backend/db"
)

func PublicacionesRedSocialHandler(dbEmpresas *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			pubs, err := db.GetPublicacionesRedSocialActivas(dbEmpresas)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(pubs)
			return
		}
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

func EmpresaPublicacionesRedSocialHandler(dbEmpresas *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, ok := r.Context().Value("empresa_id").(int)
		if !ok || empresaID == 0 {
			http.Error(w, "Acceso denegado o empresa no seleccionada", http.StatusForbidden)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet {
			pubs, err := db.GetPublicacionesRedSocialByEmpresa(dbEmpresas, empresaID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(pubs)
			return
		}

		if r.Method == http.MethodPost {
			var p db.PublicacionRedSocial
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			p.EmpresaID = empresaID
			if p.Estado == "" {
				p.Estado = "activo"
			}
			if err := db.InsertPublicacionRedSocial(dbEmpresas, &p); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(p)
			return
		}

		if r.Method == http.MethodPut {
			idStr := strings.TrimPrefix(r.URL.Path, "/api/empresa/publicaciones/")
			id, _ := strconv.Atoi(idStr)
			var p db.PublicacionRedSocial
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			p.EmpresaID = empresaID
			p.ID = id
			if err := db.UpdatePublicacionRedSocial(dbEmpresas, &p); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}

		if r.Method == http.MethodDelete {
			idStr := strings.TrimPrefix(r.URL.Path, "/api/empresa/publicaciones/")
			id, _ := strconv.Atoi(idStr)
			if err := db.DeletePublicacionRedSocial(dbEmpresas, id, empresaID); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}
