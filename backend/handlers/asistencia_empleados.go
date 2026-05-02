package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaAsistenciaEmpleadosHandler gestiona el modulo de control de asistencia por empresa.
func EmpresaAsistenciaEmpleadosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))

		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			switch action {
			case "config", "configuracion":
				cfg, err := dbpkg.GetEmpresaAsistenciaConfiguracion(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar la configuracion de asistencia", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, cfg)
				return

			case "dashboard", "resumen", "resumen_operativo":
				includeInactive := queryBool(r, "include_inactive")
				desde := strings.TrimSpace(r.URL.Query().Get("desde"))
				hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
				estadoAsistencia := strings.TrimSpace(r.URL.Query().Get("estado_asistencia"))
				q := strings.TrimSpace(r.URL.Query().Get("q"))
				dashboard, err := buildAsistenciaDashboard(dbEmp, empresaID, includeInactive, desde, hasta, estadoAsistencia, q)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, dashboard)
				return

			case "periodos_cerrados", "cierres_periodo", "cierre_periodo":
				includeInactive := queryBool(r, "include_inactive")
				desde := strings.TrimSpace(r.URL.Query().Get("desde"))
				hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
				limit, err := parseIntQueryOptional(r, "limit")
				if err != nil {
					http.Error(w, "limit invalido", http.StatusBadRequest)
					return
				}
				rows, err := dbpkg.ListEmpresaAsistenciaPeriodosCerrados(dbEmp, empresaID, includeInactive, desde, hasta, limit)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			}

			includeInactive := queryBool(r, "include_inactive")
			desde := strings.TrimSpace(r.URL.Query().Get("desde"))
			hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
			estadoAsistencia := strings.TrimSpace(r.URL.Query().Get("estado_asistencia"))
			q := strings.TrimSpace(r.URL.Query().Get("q"))
			limit, err := parseIntQueryOptional(r, "limit")
			if err != nil {
				http.Error(w, "limit invalido", http.StatusBadRequest)
				return
			}

			rows, err := dbpkg.ListEmpresaAsistenciaEmpleados(dbEmp, empresaID, includeInactive, desde, hasta, estadoAsistencia, q, limit)
			if err != nil {
				http.Error(w, "No se pudo listar la asistencia", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, rows)
			return

		case http.MethodPost:
			switch action {
			case "cerrar_periodo", "cierre_periodo":
				var payload dbpkg.EmpresaAsistenciaPeriodoCierre
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if payload.EmpresaID <= 0 {
					if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
						payload.EmpresaID = empresaID
					}
				}
				if payload.EmpresaID <= 0 {
					http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
					return
				}
				adminEmail := strings.TrimSpace(adminEmailFromRequest(r))
				payload.UsuarioCreador = adminEmail
				if strings.TrimSpace(payload.CerradoPor) == "" {
					payload.CerradoPor = adminEmail
				}

				id, err := dbpkg.CreateEmpresaAsistenciaPeriodoCierre(dbEmp, payload)
				if err != nil {
					if errors.Is(err, dbpkg.ErrAsistenciaPeriodoSolapado) {
						http.Error(w, err.Error(), http.StatusConflict)
						return
					}
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			}

			var payload dbpkg.EmpresaAsistenciaEmpleado
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 {
				if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
					payload.EmpresaID = empresaID
				}
			}
			if payload.EmpresaID <= 0 {
				http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
				return
			}
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			id, err := dbpkg.CreateEmpresaAsistenciaEmpleado(dbEmp, payload)
			if err != nil {
				if errors.Is(err, dbpkg.ErrAsistenciaPeriodoCerrado) {
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			return

		case http.MethodPut:
			switch action {
			case "config", "configuracion":
				var payload dbpkg.EmpresaAsistenciaConfiguracion
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if payload.EmpresaID <= 0 {
					if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
						payload.EmpresaID = empresaID
					}
				}
				if payload.EmpresaID <= 0 {
					http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
					return
				}
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				id, err := dbpkg.UpsertEmpresaAsistenciaConfiguracion(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				cfg, err := dbpkg.GetEmpresaAsistenciaConfiguracion(dbEmp, payload.EmpresaID)
				if err != nil {
					writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id, "configuracion": cfg})
				return

			case "activar", "desactivar":
				empresaID, err := parseEmpresaIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				id, err := parseInt64Query(r, "id")
				if err != nil {
					http.Error(w, "id es obligatorio", http.StatusBadRequest)
					return
				}
				estado := "activo"
				if action == "desactivar" {
					estado = "inactivo"
				}
				if err := dbpkg.SetEmpresaAsistenciaEmpleadoEstado(dbEmp, empresaID, id, estado); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "registro de asistencia no encontrado", http.StatusNotFound)
						return
					}
					if errors.Is(err, dbpkg.ErrAsistenciaPeriodoCerrado) {
						http.Error(w, err.Error(), http.StatusConflict)
						return
					}
					http.Error(w, "No se pudo actualizar el estado del registro", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": estado})
				return

			case "marcar_entrada":
				var payload dbpkg.EmpresaAsistenciaEmpleado
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if payload.EmpresaID <= 0 {
					if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
						payload.EmpresaID = empresaID
					}
				}
				if payload.ID <= 0 {
					if id, err := parseInt64QueryOptional(r, "id"); err == nil && id > 0 {
						payload.ID = id
					}
				}
				if payload.EmpresaID <= 0 || payload.ID <= 0 {
					http.Error(w, "empresa_id e id son obligatorios", http.StatusBadRequest)
					return
				}
				if err := dbpkg.MarkEmpresaAsistenciaEntrada(dbEmp, payload.EmpresaID, payload.ID, payload.HoraEntrada, payload.MinutosTarde, payload.EstadoAsistencia, payload.Novedad); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "registro de asistencia no encontrado", http.StatusNotFound)
						return
					}
					if errors.Is(err, dbpkg.ErrAsistenciaPeriodoCerrado) {
						http.Error(w, err.Error(), http.StatusConflict)
						return
					}
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "accion": "marcar_entrada"})
				return

			case "marcar_salida":
				var payload dbpkg.EmpresaAsistenciaEmpleado
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if payload.EmpresaID <= 0 {
					if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
						payload.EmpresaID = empresaID
					}
				}
				if payload.ID <= 0 {
					if id, err := parseInt64QueryOptional(r, "id"); err == nil && id > 0 {
						payload.ID = id
					}
				}
				if payload.EmpresaID <= 0 || payload.ID <= 0 {
					http.Error(w, "empresa_id e id son obligatorios", http.StatusBadRequest)
					return
				}
				if err := dbpkg.MarkEmpresaAsistenciaSalida(dbEmp, payload.EmpresaID, payload.ID, payload.HoraSalida, payload.Novedad); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "registro de asistencia no encontrado", http.StatusNotFound)
						return
					}
					if errors.Is(err, dbpkg.ErrAsistenciaPeriodoCerrado) {
						http.Error(w, err.Error(), http.StatusConflict)
						return
					}
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "accion": "marcar_salida"})
				return
			}

			var payload dbpkg.EmpresaAsistenciaEmpleado
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 || payload.ID <= 0 {
				http.Error(w, "empresa_id e id son obligatorios", http.StatusBadRequest)
				return
			}
			if err := dbpkg.UpdateEmpresaAsistenciaEmpleado(dbEmp, payload); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "registro de asistencia no encontrado", http.StatusNotFound)
					return
				}
				if errors.Is(err, dbpkg.ErrAsistenciaPeriodoCerrado) {
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return

		case http.MethodDelete:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			id, err := parseInt64Query(r, "id")
			if err != nil {
				http.Error(w, "id es obligatorio", http.StatusBadRequest)
				return
			}
			if err := dbpkg.DeleteEmpresaAsistenciaEmpleado(dbEmp, empresaID, id); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "registro de asistencia no encontrado", http.StatusNotFound)
					return
				}
				if errors.Is(err, dbpkg.ErrAsistenciaPeriodoCerrado) {
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}
				http.Error(w, "No se pudo eliminar el registro de asistencia", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

type asistenciaDashboard struct {
	EmpresaID         int64             `json:"empresa_id"`
	Desde             string            `json:"desde"`
	Hasta             string            `json:"hasta"`
	TotalRegistros    int               `json:"total_registros"`
	EmpleadosUnicos   int               `json:"empleados_unicos"`
	Presentes         int               `json:"presentes"`
	Tardes            int               `json:"tardes"`
	Ausentes          int               `json:"ausentes"`
	Permisos          int               `json:"permisos"`
	Incapacidades     int               `json:"incapacidades"`
	Vacaciones        int               `json:"vacaciones"`
	Pendientes        int               `json:"pendientes"`
	TurnosAbiertos    int               `json:"turnos_abiertos"`
	MinutosTardeTotal int               `json:"minutos_tarde_total"`
	HorasTrabajadas   float64           `json:"horas_trabajadas"`
	PeriodosCerrados  int               `json:"periodos_cerrados"`
	Estados           map[string]int    `json:"estados"`
	Alertas           []string          `json:"alertas,omitempty"`
	TopTardanzas      []asistenciaFicha `json:"top_tardanzas,omitempty"`
}

type asistenciaFicha struct {
	EmpleadoNombre    string  `json:"empleado_nombre"`
	EmpleadoCodigo    string  `json:"empleado_codigo,omitempty"`
	EmpleadoDocumento string  `json:"empleado_documento,omitempty"`
	FechaAsistencia   string  `json:"fecha_asistencia,omitempty"`
	Turno             string  `json:"turno,omitempty"`
	MinutosTarde      int     `json:"minutos_tarde"`
	HorasTrabajadas   float64 `json:"horas_trabajadas"`
	EstadoAsistencia  string  `json:"estado_asistencia,omitempty"`
}

func buildAsistenciaDashboard(dbEmp *sql.DB, empresaID int64, includeInactive bool, desde, hasta, estadoAsistencia, q string) (*asistenciaDashboard, error) {
	rows, err := dbpkg.ListEmpresaAsistenciaEmpleados(dbEmp, empresaID, includeInactive, desde, hasta, estadoAsistencia, q, 5000)
	if err != nil {
		return nil, err
	}
	cierres, err := dbpkg.ListEmpresaAsistenciaPeriodosCerrados(dbEmp, empresaID, false, desde, hasta, 500)
	if err != nil {
		return nil, err
	}
	out := &asistenciaDashboard{
		EmpresaID:        empresaID,
		Desde:            strings.TrimSpace(desde),
		Hasta:            strings.TrimSpace(hasta),
		Estados:          make(map[string]int),
		Alertas:          make([]string, 0),
		TopTardanzas:     make([]asistenciaFicha, 0),
		PeriodosCerrados: len(cierres),
	}
	unique := make(map[string]struct{})
	tardanzas := make([]asistenciaFicha, 0)
	for _, item := range rows {
		out.TotalRegistros++
		key := strings.ToLower(strings.TrimSpace(item.EmpleadoDocumento))
		if key == "" {
			key = strings.ToLower(strings.TrimSpace(item.EmpleadoCodigo))
		}
		if key == "" {
			key = strings.ToLower(strings.TrimSpace(item.EmpleadoNombre))
		}
		if key != "" {
			unique[key] = struct{}{}
		}
		estado := strings.ToLower(strings.TrimSpace(item.EstadoAsistencia))
		if estado == "" {
			estado = "pendiente"
		}
		out.Estados[estado]++
		switch estado {
		case "presente":
			out.Presentes++
		case "tarde":
			out.Tardes++
		case "ausente":
			out.Ausentes++
		case "permiso":
			out.Permisos++
		case "incapacidad":
			out.Incapacidades++
		case "vacaciones":
			out.Vacaciones++
		default:
			out.Pendientes++
		}
		if strings.TrimSpace(item.HoraEntrada) != "" && strings.TrimSpace(item.HoraSalida) == "" && estado != "ausente" && estado != "permiso" && estado != "incapacidad" && estado != "vacaciones" {
			out.TurnosAbiertos++
		}
		out.MinutosTardeTotal += item.MinutosTarde
		out.HorasTrabajadas += item.HorasTrabajadas
		if item.MinutosTarde > 0 {
			tardanzas = append(tardanzas, asistenciaFicha{
				EmpleadoNombre:    item.EmpleadoNombre,
				EmpleadoCodigo:    item.EmpleadoCodigo,
				EmpleadoDocumento: item.EmpleadoDocumento,
				FechaAsistencia:   item.FechaAsistencia,
				Turno:             item.Turno,
				MinutosTarde:      item.MinutosTarde,
				HorasTrabajadas:   item.HorasTrabajadas,
				EstadoAsistencia:  item.EstadoAsistencia,
			})
		}
	}
	out.EmpleadosUnicos = len(unique)
	sort.Slice(tardanzas, func(i, j int) bool {
		if tardanzas[i].MinutosTarde == tardanzas[j].MinutosTarde {
			return tardanzas[i].FechaAsistencia > tardanzas[j].FechaAsistencia
		}
		return tardanzas[i].MinutosTarde > tardanzas[j].MinutosTarde
	})
	if len(tardanzas) > 5 {
		tardanzas = tardanzas[:5]
	}
	out.TopTardanzas = tardanzas

	if out.TotalRegistros == 0 {
		out.Alertas = append(out.Alertas, "No hay registros de asistencia para el rango consultado.")
	}
	if out.TurnosAbiertos > 0 {
		out.Alertas = append(out.Alertas, "Hay turnos con entrada marcada pero sin salida registrada.")
	}
	if out.Tardes > 0 {
		out.Alertas = append(out.Alertas, "Existen registros con tardanza que conviene revisar con supervisión.")
	}
	if out.Pendientes > 0 {
		out.Alertas = append(out.Alertas, "Hay asistencias pendientes sin clasificar completamente.")
	}
	if out.Ausentes > 0 {
		out.Alertas = append(out.Alertas, "Se detectaron ausencias en el período actual.")
	}
	return out, nil
}
