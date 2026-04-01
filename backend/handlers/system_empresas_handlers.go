package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func MeHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("session_token")
		if err != nil || c == nil || c.Value == "" {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		s, err := dbpkg.GetSessionByToken(dbSuper, c.Value)
		if err != nil || s == nil {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		admin, err := dbpkg.GetAdminByEmail(dbSuper, s.AdminEmail)
		if err != nil || admin == nil {
			http.Error(w, "no admin found", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(admin)
	}
}

// SecurityPortsHandler intenta conexiones TCP a una lista de puertos y devuelve su estado.
func SecurityPortsHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		ip := q.Get("ip")
		if ip == "" {
			ip = "127.0.0.1"
		}
		portsParam := q.Get("ports")
		var ports []int
		if portsParam == "" {
			ports = []int{22, 23, 80, 443, 3306, 5432, 8080, 8443}
		} else {
			for _, s := range strings.Split(portsParam, ",") {
				s = strings.TrimSpace(s)
				if s == "" {
					continue
				}
				if n, err := strconv.Atoi(s); err == nil {
					ports = append(ports, n)
				}
			}
		}
		timeout := 500 * time.Millisecond
		if tms := q.Get("timeout_ms"); tms != "" {
			if ms, err := strconv.Atoi(tms); err == nil && ms > 0 {
				timeout = time.Duration(ms) * time.Millisecond
			}
		}

		type Entry struct {
			Puerto   int    `json:"puerto"`
			Estado   string `json:"estado"`
			IP       string `json:"ip"`
			Firewall string `json:"firewall"`
		}
		var resp []Entry
		for _, p := range ports {
			addr := fmt.Sprintf("%s:%d", ip, p)
			conn, err := net.DialTimeout("tcp", addr, timeout)
			estado := "cerrado"
			if err == nil {
				estado = "abierto"
				conn.Close()
			}
			resp = append(resp, Entry{Puerto: p, Estado: estado, IP: ip, Firewall: "Desconocido"})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// EmpresasHandler maneja CRUD de empresas en la base empresas.db
func EmpresasHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Si se pasa ?id=<id> devolver una sola empresa
			q := r.URL.Query()
			idStr := q.Get("id")
			if idStr != "" {
				id, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil {
					http.Error(w, "invalid id", http.StatusBadRequest)
					return
				}
				empresa, err := dbpkg.GetEmpresaByID(dbEmp, id)
				if err != nil {
					if err == sql.ErrNoRows {
						http.Error(w, "empresa not found", http.StatusNotFound)
					} else {
						log.Println("GET /super/api/empresas?id= error:", err)
						http.Error(w, "failed to query empresa: "+err.Error(), http.StatusInternalServerError)
					}
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(empresa)
				return
			}
			// Sin id: devolver lista completa
			empresas, err := dbpkg.GetEmpresas(dbEmp)
			if err != nil {
				log.Println("GET /super/api/empresas error:", err)
				http.Error(w, "failed to query empresas: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(empresas)
			return
		case http.MethodPost:
			var payload struct {
				TipoID         int64  `json:"tipo_id"`
				TipoNombre     string `json:"tipo_nombre"`
				Nombre         string `json:"nombre"`
				Nit            string `json:"nit"`
				Observaciones  string `json:"observaciones"`
				UsuarioCreador string `json:"usuario_creador"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if payload.Nombre == "" {
				http.Error(w, "nombre required", http.StatusBadRequest)
				return
			}
			id, err := dbpkg.CreateEmpresa(dbEmp, payload.TipoID, payload.TipoNombre, payload.Nombre, payload.Nit, payload.Observaciones, payload.UsuarioCreador)
			if err != nil {
				log.Println("POST /super/api/empresas error:", err)
				http.Error(w, "failed to create empresa: "+err.Error(), http.StatusInternalServerError)
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
			if q.Get("action") == "activar" {
				activoStr := q.Get("activo")
				if activoStr == "" {
					http.Error(w, "activo required (0 or 1)", http.StatusBadRequest)
					return
				}
				act, err := strconv.Atoi(activoStr)
				if err != nil || (act != 0 && act != 1) {
					http.Error(w, "invalid activo value", http.StatusBadRequest)
					return
				}
				estado := "inactivo"
				if act == 1 {
					estado = "activo"
				}
				if err := dbpkg.SetEmpresaEstado(dbEmp, id, estado); err != nil {
					log.Println("ACTIVAR /super/api/empresas error:", err)
					http.Error(w, "failed to set estado: "+err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
			var payloadUpdate struct {
				TipoID        int64  `json:"tipo_id"`
				TipoNombre    string `json:"tipo_nombre"`
				Nombre        string `json:"nombre"`
				Nit           string `json:"nit"`
				Observaciones string `json:"observaciones"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payloadUpdate); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if err := dbpkg.UpdateEmpresa(dbEmp, id, payloadUpdate.TipoID, payloadUpdate.TipoNombre, payloadUpdate.Nombre, payloadUpdate.Nit, payloadUpdate.Observaciones); err != nil {
				log.Println("PUT /super/api/empresas error:", err)
				http.Error(w, "failed to update empresa: "+err.Error(), http.StatusInternalServerError)
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
			if err := dbpkg.DeleteEmpresa(dbEmp, id); err != nil {
				log.Println("DELETE /super/api/empresas error:", err)
				http.Error(w, "failed to delete empresa: "+err.Error(), http.StatusInternalServerError)
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
