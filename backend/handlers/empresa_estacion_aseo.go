package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaEstacionAseoHandler maneja el control operativo de aseo por estacion.
func EmpresaEstacionAseoHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		usuario, role := resolveEmpresaAseoUsuarioActual(dbEmp, r, empresaID)
		puedeFinalizarAseo := empresaAseoRoleCanClean(role)
		controlHabilitado := (usuario != nil && usuario.ControlAseoEstaciones == 1) || puedeFinalizarAseo
		puedeGestionar := empresaAseoRoleCanManage(role)
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "" {
			action = "contexto"
		}

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "contexto":
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":                      true,
					"control_aseo_habilitado": controlHabilitado,
					"puede_ver_reporte":       puedeGestionar,
					"usuario":                 empresaAseoUsuarioResponse(usuario, adminEmailFromRequest(r), role),
				})
				return
			case "reporte":
				if !puedeGestionar {
					http.Error(w, "forbidden: no tienes permiso para ver el reporte de aseo", http.StatusForbidden)
					return
				}
				filtro := dbpkg.EmpresaEstacionAseoFiltro{
					EmpresaID:  empresaID,
					EstacionID: parseInt64QueryOptionalDefault(r, "estacion_id", 0),
					UsuarioID:  parseInt64QueryOptionalDefault(r, "usuario_id", 0),
					Desde:      normalizeEmpresaAseoDateBoundary(r.URL.Query().Get("desde"), false),
					Hasta:      normalizeEmpresaAseoDateBoundary(r.URL.Query().Get("hasta"), true),
					Estado:     strings.TrimSpace(r.URL.Query().Get("estado")),
					Limit:      int(parseInt64QueryOptionalDefault(r, "limit", 300)),
				}
				items, err := dbpkg.ListEmpresaEstacionAseoEventos(dbEmp, filtro)
				if err != nil {
					log.Printf("[estacion_aseo] reporte empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo cargar el reporte de aseo", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":      true,
					"items":   items,
					"summary": buildEmpresaAseoSummary(items),
				})
				return
			default:
				http.Error(w, "action no soportada", http.StatusBadRequest)
				return
			}

		case http.MethodPost:
			if action != "finalizar" {
				http.Error(w, "action no soportada", http.StatusBadRequest)
				return
			}
			if !controlHabilitado && !puedeGestionar {
				http.Error(w, "forbidden: usuario sin control de aseo habilitado", http.StatusForbidden)
				return
			}
			var payload struct {
				EstacionID     int64  `json:"estacion_id"`
				EstacionNombre string `json:"estacion_nombre"`
				Observaciones  string `json:"observaciones"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.EstacionID <= 0 {
				http.Error(w, "estacion_id es obligatorio", http.StatusBadRequest)
				return
			}
			input := dbpkg.EmpresaEstacionAseoFinalizarInput{
				EmpresaID:      empresaID,
				EstacionID:     payload.EstacionID,
				EstacionNombre: strings.TrimSpace(payload.EstacionNombre),
				UsuarioEmail:   strings.TrimSpace(adminEmailFromRequest(r)),
				RolNombre:      role,
				Observaciones:  strings.TrimSpace(payload.Observaciones),
				Origen:         "estaciones",
			}
			if usuario != nil {
				input.UsuarioID = usuario.ID
				input.UsuarioEmail = usuario.Email
				input.UsuarioNombre = usuario.Nombre
				input.RolNombre = usuario.RolNombre
			}
			evento, err := dbpkg.FinalizarEmpresaEstacionAseo(dbEmp, input)
			if err != nil {
				if strings.Contains(strings.ToLower(err.Error()), "no esta marcada") {
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}
				log.Printf("[estacion_aseo] finalizar empresa_id=%d estacion_id=%d error: %v", empresaID, payload.EstacionID, err)
				http.Error(w, "No se pudo reportar el aseo", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "evento": evento})
			return

		default:
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
	}
}

func resolveEmpresaAseoUsuarioActual(dbEmp *sql.DB, r *http.Request, empresaID int64) (*dbpkg.EmpresaUsuario, string) {
	email := strings.TrimSpace(adminEmailFromRequest(r))
	role := strings.TrimSpace(adminRoleFromRequest(r))
	if email == "" {
		return nil, normalizePermissionRole(role)
	}
	item, err := dbpkg.GetEmpresaUsuarioByEmailScoped(dbEmp, email, empresaID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("[estacion_aseo] usuario actual empresa_id=%d email=%s error: %v", empresaID, redactEmailForLog(email), err)
		}
		return nil, normalizePermissionRole(role)
	}
	if strings.TrimSpace(role) == "" {
		role = item.RolNombre
	}
	return item, normalizePermissionRole(role)
}

func empresaAseoUsuarioResponse(item *dbpkg.EmpresaUsuario, fallbackEmail, role string) map[string]interface{} {
	if item == nil {
		return map[string]interface{}{
			"email":      strings.TrimSpace(fallbackEmail),
			"rol_nombre": role,
		}
	}
	return map[string]interface{}{
		"id":                      item.ID,
		"email":                   item.Email,
		"nombre":                  item.Nombre,
		"rol_nombre":              item.RolNombre,
		"control_aseo_estaciones": item.ControlAseoEstaciones,
	}
}

func empresaAseoRoleCanManage(role string) bool {
	role = normalizePermissionRole(role)
	switch role {
	case "super_administrador", "administrador_total", "admin_empresa", "supervisor_sucursal", "auditor", "reportes":
		return true
	}
	return strings.Contains(role, "admin") || strings.Contains(role, "gerente") || strings.Contains(role, "supervisor")
}

func empresaAseoRoleCanClean(role string) bool {
	return normalizePermissionRole(role) == "servicio_limpieza"
}

func parseInt64QueryOptionalDefault(r *http.Request, key string, fallback int64) int64 {
	value, err := parseInt64QueryOptional(r, key)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func normalizeEmpresaAseoDateBoundary(raw string, endOfDay bool) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if len(raw) == len("2006-01-02") && raw[4] == '-' && raw[7] == '-' {
		if endOfDay {
			return raw + " 23:59:59"
		}
		return raw + " 00:00:00"
	}
	return strings.ReplaceAll(raw, "T", " ")
}

func buildEmpresaAseoSummary(items []dbpkg.EmpresaEstacionAseoEvento) map[string]interface{} {
	total := int64(len(items))
	var sum int64
	var min int64
	var max int64
	for i, item := range items {
		d := item.DuracionSegundos
		if d < 0 {
			d = 0
		}
		sum += d
		if i == 0 || d < min {
			min = d
		}
		if d > max {
			max = d
		}
	}
	avg := int64(0)
	if total > 0 {
		avg = sum / total
	}
	return map[string]interface{}{
		"total":             total,
		"promedio_segundos": avg,
		"minimo_segundos":   min,
		"maximo_segundos":   max,
	}
}
