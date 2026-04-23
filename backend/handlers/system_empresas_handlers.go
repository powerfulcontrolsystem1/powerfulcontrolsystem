package handlers

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type empresaImpactoDesactivacion struct {
	EmpresaID            int64  `json:"empresa_id"`
	EmpresaNombre        string `json:"empresa_nombre,omitempty"`
	EstadoActual         string `json:"estado_actual,omitempty"`
	UsuariosActivos      int64  `json:"usuarios_activos"`
	CarritosAbiertos     int64  `json:"carritos_abiertos"`
	ReservasVigentes     int64  `json:"reservas_vigentes"`
	LicenciasActivas     int64  `json:"licencias_activas"`
	Bloqueos             int64  `json:"bloqueos"`
	RequiereConfirmacion bool   `json:"requiere_confirmacion"`
}

func superTableExists(dbConn *sql.DB, tableName string) bool {
	if dbConn == nil {
		return false
	}
	var total int
	err := dbConn.QueryRow(`
		SELECT COUNT(1)
		FROM information_schema.tables
		WHERE table_schema = ANY (current_schemas(false))
		  AND table_name = ?
	`, strings.TrimSpace(tableName)).Scan(&total)
	return err == nil && total > 0
}

func superCountIfTableExists(dbConn *sql.DB, tableName, query string, args ...interface{}) (int64, error) {
	if !superTableExists(dbConn, tableName) {
		return 0, nil
	}
	var total int64
	if err := dbConn.QueryRow(query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func parseBoolQueryValue(raw string) bool {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "1", "true", "t", "si", "sí", "y", "yes", "on":
		return true
	default:
		return false
	}
}

func resolveRequesterAdminScope(dbSuper *sql.DB, r *http.Request) (*dbpkg.Admin, string, error) {
	requesterEmail := strings.ToLower(strings.TrimSpace(adminEmailFromRequest(r)))
	if requesterEmail == "" || requesterEmail == "sistema" {
		return nil, "", nil
	}
	admin, err := dbpkg.GetAdminByEmailFull(dbSuper, requesterEmail)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, requesterEmail, nil
		}
		return nil, "", err
	}
	principalEmail, err := dbpkg.ResolveAdminPrincipalEmail(dbSuper, requesterEmail)
	if err != nil {
		return nil, "", err
	}
	if principalEmail == "" {
		principalEmail = requesterEmail
	}
	return admin, principalEmail, nil
}

func adminEmailMatchesPrincipalScope(dbSuper *sql.DB, principalEmail, targetEmail string) (bool, error) {
	principalEmail = strings.ToLower(strings.TrimSpace(principalEmail))
	targetEmail = strings.ToLower(strings.TrimSpace(targetEmail))
	if principalEmail == "" || targetEmail == "" {
		return principalEmail == "", nil
	}
	if principalEmail == targetEmail {
		return true, nil
	}
	resolved, err := dbpkg.ResolveAdminPrincipalEmail(dbSuper, targetEmail)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return strings.EqualFold(strings.TrimSpace(resolved), principalEmail), nil
}

func empresaBelongsToPrincipalScope(dbSuper *sql.DB, principalEmail, empresaCreator string) (bool, error) {
	principalEmail = strings.ToLower(strings.TrimSpace(principalEmail))
	creator := strings.ToLower(strings.TrimSpace(empresaCreator))
	if principalEmail == "" || creator == "" {
		return true, nil
	}
	if creator == principalEmail {
		return true, nil
	}
	return adminEmailMatchesPrincipalScope(dbSuper, principalEmail, creator)
}

func filterEmpresasByPrincipalScope(dbSuper *sql.DB, principalEmail string, empresas []dbpkg.Empresa) ([]dbpkg.Empresa, error) {
	if strings.TrimSpace(principalEmail) == "" {
		return empresas, nil
	}
	filtered := make([]dbpkg.Empresa, 0, len(empresas))
	for _, empresa := range empresas {
		ok, err := empresaBelongsToPrincipalScope(dbSuper, principalEmail, empresa.UsuarioCreador)
		if err != nil {
			return nil, err
		}
		if ok {
			filtered = append(filtered, empresa)
		}
	}
	return filtered, nil
}

func ensureEmpresaInRequesterScope(dbEmp, dbSuper *sql.DB, r *http.Request, empresaID int64) (string, bool, error) {
	requesterEmail := strings.ToLower(strings.TrimSpace(adminEmailFromRequest(r)))
	_, principalEmail, err := resolveRequesterAdminScope(dbSuper, r)
	if err != nil {
		return "", false, err
	}
	if principalEmail == "" || empresaID <= 0 {
		return principalEmail, true, nil
	}
	ok, err := dbpkg.CanAdminAccessEmpresaIA(dbEmp, dbSuper, requesterEmail, empresaID)
	if err != nil {
		return principalEmail, false, err
	}
	return principalEmail, ok, nil
}

func buildEmpresaImpactoDesactivacion(dbEmp, dbSuper *sql.DB, empresaID int64) (*empresaImpactoDesactivacion, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id invalido")
	}

	empresa, err := dbpkg.GetEmpresaByID(dbEmp, empresaID)
	if err != nil {
		return nil, err
	}

	impacto := &empresaImpactoDesactivacion{
		EmpresaID:     empresaID,
		EmpresaNombre: strings.TrimSpace(empresa.Nombre),
		EstadoActual:  strings.TrimSpace(empresa.Estado),
	}

	usuariosActivos, err := superCountIfTableExists(
		dbEmp,
		"users",
		`SELECT COUNT(1) FROM users WHERE empresa_id = ? AND LOWER(COALESCE(estado, 'activo')) = 'activo'`,
		empresaID,
	)
	if err != nil {
		return nil, err
	}
	impacto.UsuariosActivos = usuariosActivos

	carritosAbiertos, err := superCountIfTableExists(
		dbEmp,
		"carritos_compras",
		`SELECT COUNT(1)
		 FROM carritos_compras
		 WHERE empresa_id = ?
		   AND LOWER(COALESCE(estado, 'activo')) = 'activo'
		   AND LOWER(COALESCE(estado_carrito, 'abierto')) NOT IN ('cerrado', 'cancelado', 'anulado')`,
		empresaID,
	)
	if err != nil {
		return nil, err
	}
	impacto.CarritosAbiertos = carritosAbiertos

	reservasVigentes, err := superCountIfTableExists(
		dbEmp,
		"reservas_hotel",
		`SELECT COUNT(1)
		 FROM reservas_hotel
		 WHERE empresa_id = ?
		   AND LOWER(COALESCE(estado, 'activo')) = 'activo'
		   AND LOWER(COALESCE(estado_reserva, 'pendiente_pago')) IN ('pendiente_pago', 'confirmada')`,
		empresaID,
	)
	if err != nil {
		return nil, err
	}
	impacto.ReservasVigentes = reservasVigentes

	licenciasActivas, err := superCountIfTableExists(
		dbSuper,
		"licencias",
		`SELECT COUNT(1)
		 FROM licencias
		 WHERE empresa_id = ?
		   AND (COALESCE(CAST(activo AS TEXT), '0') IN ('1', 'activo', 'true'))`,
		empresaID,
	)
	if err != nil {
		return nil, err
	}
	impacto.LicenciasActivas = licenciasActivas

	impacto.Bloqueos = impacto.UsuariosActivos + impacto.CarritosAbiertos + impacto.ReservasVigentes + impacto.LicenciasActivas
	impacto.RequiereConfirmacion = impacto.Bloqueos > 0

	return impacto, nil
}

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
			addr := net.JoinHostPort(ip, strconv.Itoa(p))
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

// SecurityProcessesHandler devuelve la lista de procesos activos y uso de memoria.
func SecurityProcessesHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		q := r.URL.Query()
		limit := 200
		if l := q.Get("limit"); l != "" {
			if n, err := strconv.Atoi(l); err == nil && n > 0 {
				limit = n
			}
		}

		type ProcEntry struct {
			PID      int64  `json:"pid"`
			Name     string `json:"name"`
			MemoryKB int64  `json:"memory_kb"`
		}

		var procs []ProcEntry

		if runtime.GOOS == "windows" {
			// tasklist CSV: "Image Name","PID","Session Name","Session#","Mem Usage"
			cmd := exec.Command("tasklist", "/FO", "CSV", "/NH")
			out, err := cmd.Output()
			if err != nil {
				http.Error(w, "failed to list processes", http.StatusInternalServerError)
				return
			}
			rdr := csv.NewReader(bytes.NewReader(out))
			rdr.FieldsPerRecord = -1
			for {
				rec, err := rdr.Read()
				if err == io.EOF {
					break
				}
				if err != nil {
					continue
				}
				if len(rec) < 5 {
					continue
				}
				name := strings.Trim(rec[0], " \"\r\n")
				pidStr := strings.Trim(rec[1], " \"\r\n")
				memStr := strings.Trim(rec[4], " \"\r\n")
				pid, _ := strconv.ParseInt(strings.TrimSpace(pidStr), 10, 64)
				// Mem string example: "12,345 K"
				memClean := strings.ReplaceAll(memStr, ",", "")
				memClean = strings.ReplaceAll(memClean, "K", "")
				memClean = strings.ReplaceAll(memClean, "k", "")
				memClean = strings.TrimSpace(memClean)
				memVal, _ := strconv.ParseInt(memClean, 10, 64)
				procs = append(procs, ProcEntry{PID: pid, Name: name, MemoryKB: memVal})
			}
		} else {
			// Unix-like: ps -eo pid,comm,rss  (rss in KB)
			cmd := exec.Command("ps", "-eo", "pid,comm,rss")
			out, err := cmd.Output()
			if err != nil {
				http.Error(w, "failed to list processes", http.StatusInternalServerError)
				return
			}
			lines := strings.Split(strings.TrimSpace(string(out)), "\n")
			start := 0
			if len(lines) > 0 && strings.Contains(strings.ToUpper(lines[0]), "PID") {
				start = 1
			}
			for i := start; i < len(lines); i++ {
				if len(procs) >= limit*10 { // safety cap while parsing
					break
				}
				f := strings.Fields(lines[i])
				if len(f) < 3 {
					continue
				}
				pid, _ := strconv.ParseInt(f[0], 10, 64)
				name := f[1]
				rss, _ := strconv.ParseInt(f[len(f)-1], 10, 64)
				procs = append(procs, ProcEntry{PID: pid, Name: name, MemoryKB: rss})
			}
		}

		// Ordenar por memoria descendente
		sort.Slice(procs, func(i, j int) bool { return procs[i].MemoryKB > procs[j].MemoryKB })

		if len(procs) > limit {
			procs = procs[:limit]
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(procs)
	}
}

// EmpresasHandler maneja CRUD de empresas en la base empresas.db
func EmpresasHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			requesterEmail := strings.ToLower(strings.TrimSpace(adminEmailFromRequest(r)))
			_, principalEmail, err := resolveRequesterAdminScope(dbSuper, r)
			if err != nil {
				log.Println("GET /super/api/empresas scope error:", err)
				http.Error(w, "failed to resolve admin scope: "+err.Error(), http.StatusInternalServerError)
				return
			}
			// Si se pasa ?id=<id> devolver una sola empresa
			q := r.URL.Query()
			action := strings.ToLower(strings.TrimSpace(q.Get("action")))
			idStr := q.Get("id")
			if idStr != "" {
				id, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil {
					http.Error(w, "invalid id", http.StatusBadRequest)
					return
				}
				if _, ok, err := ensureEmpresaInRequesterScope(dbEmp, dbSuper, r, id); err != nil {
					log.Printf("GET /super/api/empresas?id=%d scope error: %v", id, err)
					http.Error(w, "failed to validate empresa scope: "+err.Error(), http.StatusInternalServerError)
					return
				} else if !ok {
					http.Error(w, "empresa fuera del alcance del administrador autenticado", http.StatusForbidden)
					return
				}

				if action == "impacto" || action == "impacto_desactivacion" {
					impacto, err := buildEmpresaImpactoDesactivacion(dbEmp, dbSuper, id)
					if err != nil {
						if err == sql.ErrNoRows {
							http.Error(w, "empresa not found", http.StatusNotFound)
							return
						}
						log.Printf("GET /super/api/empresas?action=%s&id=%d impacto error: %v", action, id, err)
						http.Error(w, "failed to evaluate empresa impact: "+err.Error(), http.StatusInternalServerError)
						return
					}
					writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "impacto": impacto})
					return
				}

				if action == "resumen_descarga" {
					snapshot, err := buildEmpresaInfoExportSnapshot(dbEmp, dbSuper, id, 3)
					if err != nil {
						if err == sql.ErrNoRows {
							http.Error(w, "empresa not found", http.StatusNotFound)
							return
						}
						log.Printf("GET /super/api/empresas?action=%s&id=%d resumen descarga error: %v", action, id, err)
						http.Error(w, "failed to build empresa export summary: "+err.Error(), http.StatusInternalServerError)
						return
					}
					writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "snapshot": snapshot})
					return
				}

				if action == "exportar_informacion" {
					format := strings.TrimSpace(q.Get("format"))
					if format == "" {
						http.Error(w, "format required", http.StatusBadRequest)
						return
					}
					snapshot, err := buildEmpresaInfoExportSnapshot(dbEmp, dbSuper, id, 0)
					if err != nil {
						if err == sql.ErrNoRows {
							http.Error(w, "empresa not found", http.StatusNotFound)
							return
						}
						log.Printf("GET /super/api/empresas?action=%s&id=%d export error: %v", action, id, err)
						http.Error(w, "failed to build empresa export: "+err.Error(), http.StatusInternalServerError)
						return
					}
					if err := writeEmpresaInfoExport(w, snapshot, format); err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
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
				if err := decorateEmpresaAccessForRequester(dbSuper, requesterEmail, principalEmail, empresa); err != nil {
					log.Println("GET /super/api/empresas?id= decorate access error:", err)
					http.Error(w, "failed to decorate empresa access: "+err.Error(), http.StatusInternalServerError)
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
			empresas, err = decorateEmpresasByEffectiveAccess(dbSuper, requesterEmail, principalEmail, empresas)
			if err != nil {
				log.Println("GET /super/api/empresas effective access error:", err)
				http.Error(w, "failed to resolve empresa access: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(empresas)
			return
		case http.MethodPost:
			_, principalEmail, err := resolveRequesterAdminScope(dbSuper, r)
			if err != nil {
				log.Println("POST /super/api/empresas scope error:", err)
				http.Error(w, "failed to resolve admin scope: "+err.Error(), http.StatusInternalServerError)
				return
			}
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
			if principalEmail != "" {
				payload.UsuarioCreador = principalEmail
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
			action := strings.ToLower(strings.TrimSpace(q.Get("action")))
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
			if _, ok, err := ensureEmpresaInRequesterScope(dbEmp, dbSuper, r, id); err != nil {
				log.Printf("PUT /super/api/empresas id=%d scope error: %v", id, err)
				http.Error(w, "failed to validate empresa scope: "+err.Error(), http.StatusInternalServerError)
				return
			} else if !ok {
				http.Error(w, "empresa fuera del alcance del administrador autenticado", http.StatusForbidden)
				return
			}
			if _, _, owner, err := ensureEmpresaOwnerAccess(dbEmp, dbSuper, r, id); err != nil {
				log.Printf("PUT /super/api/empresas id=%d owner error: %v", id, err)
				http.Error(w, "failed to validate empresa owner: "+err.Error(), http.StatusInternalServerError)
				return
			} else if !owner {
				http.Error(w, "solo el administrador propietario puede modificar o desactivar la empresa", http.StatusForbidden)
				return
			}

			if action == "activar" || action == "rehabilitar" || action == "desactivar" {
				estadoObjetivo := ""
				if action == "desactivar" {
					estadoObjetivo = "inactivo"
				} else {
					activoStr := strings.TrimSpace(q.Get("activo"))
					if activoStr == "" {
						estadoObjetivo = "activo"
					} else {
						act, err := strconv.Atoi(activoStr)
						if err != nil || (act != 0 && act != 1) {
							http.Error(w, "invalid activo value", http.StatusBadRequest)
							return
						}
						if act == 1 {
							estadoObjetivo = "activo"
						} else {
							estadoObjetivo = "inactivo"
						}
					}
				}

				impacto, err := buildEmpresaImpactoDesactivacion(dbEmp, dbSuper, id)
				if err != nil {
					if err == sql.ErrNoRows {
						http.Error(w, "empresa not found", http.StatusNotFound)
						return
					}
					log.Printf("PUT /super/api/empresas action=%s id=%d impacto error: %v", action, id, err)
					http.Error(w, "failed to evaluate empresa impact: "+err.Error(), http.StatusInternalServerError)
					return
				}

				force := parseBoolQueryValue(q.Get("force"))
				if estadoObjetivo == "inactivo" && impacto.RequiereConfirmacion && !force {
					writeJSON(w, http.StatusConflict, map[string]interface{}{
						"ok":                    false,
						"id":                    id,
						"estado":                strings.TrimSpace(impacto.EstadoActual),
						"requiere_confirmacion": true,
						"message":               "La empresa tiene operaciones activas. Confirma force=1 para desactivar.",
						"impacto":               impacto,
					})
					return
				}

				if err := dbpkg.SetEmpresaEstado(dbEmp, id, estadoObjetivo); err != nil {
					log.Printf("PUT /super/api/empresas action=%s id=%d set estado error: %v", action, id, err)
					http.Error(w, "failed to set estado: "+err.Error(), http.StatusInternalServerError)
					return
				}

				impactoPost, err := buildEmpresaImpactoDesactivacion(dbEmp, dbSuper, id)
				if err != nil {
					log.Printf("PUT /super/api/empresas action=%s id=%d post-impact warning: %v", action, id, err)
					impactoPost = impacto
					impactoPost.EstadoActual = estadoObjetivo
				}

				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":      true,
					"id":      id,
					"estado":  estadoObjetivo,
					"impacto": impactoPost,
				})
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
			action := strings.ToLower(strings.TrimSpace(q.Get("action")))
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
			if _, ok, err := ensureEmpresaInRequesterScope(dbEmp, dbSuper, r, id); err != nil {
				log.Printf("DELETE /super/api/empresas id=%d scope error: %v", id, err)
				http.Error(w, "failed to validate empresa scope: "+err.Error(), http.StatusInternalServerError)
				return
			} else if !ok {
				http.Error(w, "empresa fuera del alcance del administrador autenticado", http.StatusForbidden)
				return
			}
			if _, _, owner, err := ensureEmpresaOwnerAccess(dbEmp, dbSuper, r, id); err != nil {
				log.Printf("DELETE /super/api/empresas id=%d owner error: %v", id, err)
				http.Error(w, "failed to validate empresa owner: "+err.Error(), http.StatusInternalServerError)
				return
			} else if !owner {
				http.Error(w, "solo el administrador propietario puede eliminar la empresa", http.StatusForbidden)
				return
			}
			if action == "eliminar_total" || action == "purge" || action == "eliminacion_total" {
				empresa, err := dbpkg.GetEmpresaByID(dbEmp, id)
				if err != nil {
					if err == sql.ErrNoRows {
						http.Error(w, "empresa not found", http.StatusNotFound)
						return
					}
					log.Printf("DELETE /super/api/empresas action=%s id=%d load error: %v", action, id, err)
					http.Error(w, "failed to query empresa: "+err.Error(), http.StatusInternalServerError)
					return
				}

				var payload struct {
					ConfirmacionNombre string `json:"confirmacion_nombre"`
				}
				if r.Body != nil {
					defer r.Body.Close()
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && err != io.EOF {
						http.Error(w, "invalid payload", http.StatusBadRequest)
						return
					}
				}
				if strings.TrimSpace(payload.ConfirmacionNombre) == "" {
					http.Error(w, "confirmacion_nombre required", http.StatusBadRequest)
					return
				}
				if !strings.EqualFold(strings.TrimSpace(payload.ConfirmacionNombre), strings.TrimSpace(empresa.Nombre)) {
					http.Error(w, "la confirmacion no coincide con el nombre de la empresa", http.StatusConflict)
					return
				}

				result, err := dbpkg.DeleteEmpresaCascade(dbEmp, dbSuper, id)
				if err != nil {
					log.Printf("DELETE /super/api/empresas action=%s id=%d cascade error: %v", action, id, err)
					http.Error(w, "failed to purge empresa: "+err.Error(), http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "result": result})
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
