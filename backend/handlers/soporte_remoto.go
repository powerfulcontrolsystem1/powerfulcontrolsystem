package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type empresaSoporteRemotoConfigPayload struct {
	Habilitado                 *bool  `json:"habilitado"`
	ProveedorPreferido         string `json:"proveedor_preferido"`
	ModoOperacion              string `json:"modo_operacion"`
	RequiereAprobacionOperador *bool  `json:"requiere_aprobacion_operador"`
	AutoCerrarMinutos          int    `json:"auto_cerrar_minutos"`
	Observaciones              string `json:"observaciones"`
}

type empresaSoporteRemotoDispositivoPayload struct {
	ID                int64  `json:"id"`
	CodigoDispositivo string `json:"codigo_dispositivo"`
	NombreEquipo      string `json:"nombre_equipo"`
	AliasOperativo    string `json:"alias_operativo"`
	Ubicacion         string `json:"ubicacion"`
	SistemaOperativo  string `json:"sistema_operativo"`
	AgenteVersion     string `json:"agente_version"`
	StreamURL         string `json:"stream_url"`
	EstadoConexion    string `json:"estado_conexion"`
	AccesoPIN         string `json:"acceso_pin"`
	Observaciones     string `json:"observaciones"`
}

type empresaSoporteRemotoSesionPayload struct {
	DispositivoID  int64  `json:"dispositivo_id"`
	CodigoSesion   string `json:"codigo_sesion"`
	OperadorNombre string `json:"operador_nombre"`
	OperadorEmail  string `json:"operador_email"`
	Motivo         string `json:"motivo"`
	DuracionMin    int    `json:"duracion_min"`
	EstadoSesion   string `json:"estado_sesion"`
	Observaciones  string `json:"observaciones"`
}

type empresaSoporteRemotoHeartbeatPayload struct {
	EmpresaID         int64  `json:"empresa_id"`
	CodigoDispositivo string `json:"codigo_dispositivo"`
	AccesoPIN         string `json:"acceso_pin"`
	StreamURL         string `json:"stream_url"`
	SistemaOperativo  string `json:"sistema_operativo"`
	AgenteVersion     string `json:"agente_version"`
}

func empresaSoporteRemotoNormalizeAction(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "config", "configuracion":
		return "config"
	case "dispositivos", "listar_dispositivos":
		return "dispositivos"
	case "detalle_dispositivo", "dispositivo":
		return "detalle_dispositivo"
	case "crear_dispositivo":
		return "crear_dispositivo"
	case "actualizar_dispositivo":
		return "actualizar_dispositivo"
	case "activar_dispositivo":
		return "activar_dispositivo"
	case "desactivar_dispositivo":
		return "desactivar_dispositivo"
	case "sesiones", "listar_sesiones":
		return "sesiones"
	case "solicitar_sesion", "crear_sesion":
		return "solicitar_sesion"
	case "aprobar_sesion":
		return "aprobar_sesion"
	case "finalizar_sesion":
		return "finalizar_sesion"
	case "resolver_visualizacion", "resolver_view":
		return "resolver_visualizacion"
	case "export_sesiones", "export":
		return "export_sesiones"
	case "heartbeat_dispositivo", "heartbeat":
		return "heartbeat_dispositivo"
	default:
		return ""
	}
}

func empresaSoporteRemotoBuildViewerURL(r *http.Request, empresaID int64, codigoSesion, token string) string {
	if r == nil {
		return ""
	}
	base := &url.URL{Path: "/administrar_empresa/soporte_remoto_view.html"}
	q := base.Query()
	q.Set("empresa_id", strconv.FormatInt(empresaID, 10))
	q.Set("codigo_sesion", strings.TrimSpace(codigoSesion))
	q.Set("token", strings.TrimSpace(token))
	base.RawQuery = q.Encode()
	return base.String()
}

func empresaSoporteRemotoMaskStreamURL(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil {
		if len(value) <= 12 {
			return value
		}
		return value[:12] + "..."
	}
	masked := parsed.Scheme + "://" + parsed.Host
	if masked == "://" {
		return value
	}
	if parsed.Path != "" {
		masked += parsed.Path
	}
	return masked
}

func empresaSoporteRemotoComposeSessionsDataset(empresaID int64, rows []dbpkg.EmpresaSoporteRemotoSession, total int64) empresaReporteDataset {
	datasetRows := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		datasetRows = append(datasetRows, map[string]interface{}{
			"empresa_id":          row.EmpresaID,
			"codigo_sesion":       row.CodigoSesion,
			"dispositivo_id":      row.DispositivoID,
			"dispositivo_codigo":  row.DispositivoCodigo,
			"dispositivo_nombre":  row.DispositivoNombre,
			"solicitada_por":      row.SolicitadaPor,
			"operador_nombre":     row.OperadorNombre,
			"operador_email":      row.OperadorEmail,
			"estado_sesion":       row.EstadoSesion,
			"motivo":              row.Motivo,
			"url_visualizacion":   empresaSoporteRemotoMaskStreamURL(row.URLVisualizacion),
			"iniciada_en":         row.IniciadaEn,
			"expira_en":           row.ExpiraEn,
			"finalizada_en":       row.FinalizadaEn,
			"fecha_creacion":      row.FechaCreacion,
			"fecha_actualizacion": row.FechaActualizacion,
		})
	}
	return empresaReporteDataset{
		Key:         "operativo_soporte_remoto_sesiones",
		Title:       "Soporte remoto - sesiones",
		Level:       "operativo",
		Description: "Trazabilidad de sesiones de visualizacion remota por empresa.",
		EmpresaID:   empresaID,
		Desde:       "",
		Hasta:       "",
		GeneratedAt: time.Now().In(time.Local).Format("2006-01-02 15:04:05"),
		Columns: []string{
			"empresa_id",
			"codigo_sesion",
			"dispositivo_id",
			"dispositivo_codigo",
			"dispositivo_nombre",
			"solicitada_por",
			"operador_nombre",
			"operador_email",
			"estado_sesion",
			"motivo",
			"url_visualizacion",
			"iniciada_en",
			"expira_en",
			"finalizada_en",
			"fecha_creacion",
			"fecha_actualizacion",
		},
		Rows:     datasetRows,
		RowCount: len(datasetRows),
		Summary: map[string]interface{}{
			"sesiones_total":         total,
			"sesiones_exportadas":    len(datasetRows),
			"modulo":                 "soporte_remoto",
			"empresa_id":             empresaID,
			"formato_trazable":       "json,csv,txt,xls,pdf",
			"generated_at_localtime": time.Now().In(time.Local).Format("2006-01-02 15:04:05"),
		},
	}
}

// EmpresaSoporteRemotoHandler administra soporte remoto por empresa desde panel administrador.
func EmpresaSoporteRemotoHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		action := empresaSoporteRemotoNormalizeAction(r.URL.Query().Get("action"))
		if action == "" {
			http.Error(w, "action invalida", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "config":
				empresaSoporteRemotoConfigGet(w, r, dbEmp)
			case "dispositivos":
				empresaSoporteRemotoDispositivosGet(w, r, dbEmp)
			case "detalle_dispositivo":
				empresaSoporteRemotoDispositivoDetailGet(w, r, dbEmp)
			case "sesiones":
				empresaSoporteRemotoSesionesGet(w, r, dbEmp)
			case "resolver_visualizacion":
				empresaSoporteRemotoResolverVisualizacion(w, r, dbEmp)
			case "export_sesiones":
				empresaSoporteRemotoSesionesExport(w, r, dbEmp)
			default:
				http.Error(w, "action invalida", http.StatusBadRequest)
			}

		case http.MethodPost:
			switch action {
			case "config":
				empresaSoporteRemotoConfigUpsert(w, r, dbEmp)
			case "crear_dispositivo":
				empresaSoporteRemotoDispositivoCreate(w, r, dbEmp)
			case "solicitar_sesion":
				empresaSoporteRemotoSesionCreate(w, r, dbEmp)
			case "aprobar_sesion":
				empresaSoporteRemotoSesionAprobar(w, r, dbEmp)
			case "finalizar_sesion":
				empresaSoporteRemotoSesionFinalizar(w, r, dbEmp)
			case "heartbeat_dispositivo":
				empresaSoporteRemotoHeartbeat(w, r, dbEmp)
			default:
				http.Error(w, "action invalida", http.StatusBadRequest)
			}

		case http.MethodPut, http.MethodPatch:
			switch action {
			case "actualizar_dispositivo":
				empresaSoporteRemotoDispositivoUpdate(w, r, dbEmp)
			case "activar_dispositivo":
				empresaSoporteRemotoDispositivoToggle(w, r, dbEmp, "activo")
			case "desactivar_dispositivo":
				empresaSoporteRemotoDispositivoToggle(w, r, dbEmp, "inactivo")
			case "aprobar_sesion":
				empresaSoporteRemotoSesionAprobar(w, r, dbEmp)
			case "finalizar_sesion":
				empresaSoporteRemotoSesionFinalizar(w, r, dbEmp)
			default:
				http.Error(w, "action invalida", http.StatusBadRequest)
			}

		case http.MethodDelete:
			if action == "desactivar_dispositivo" {
				empresaSoporteRemotoDispositivoToggle(w, r, dbEmp, "inactivo")
				return
			}
			http.Error(w, "action invalida", http.StatusBadRequest)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// PublicEmpresaSoporteRemotoAgentHandler expone operaciones de heartbeat/estado para plugin de agencia remota.
func PublicEmpresaSoporteRemotoAgentHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		action := empresaSoporteRemotoNormalizeAction(r.URL.Query().Get("action"))
		switch action {
		case "heartbeat_dispositivo":
			empresaSoporteRemotoHeartbeat(w, r, dbEmp)
		case "aprobar_sesion", "finalizar_sesion":
			empresaSoporteRemotoSesionAgentUpdate(w, r, dbEmp, action)
		default:
			http.Error(w, "action invalida", http.StatusBadRequest)
		}
	}
}

func empresaSoporteRemotoConfigGet(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cfg, err := dbpkg.GetEmpresaSoporteRemotoConfig(dbEmp, empresaID)
	if err != nil {
		http.Error(w, "No se pudo consultar configuracion de soporte remoto", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "config": cfg})
}

func empresaSoporteRemotoConfigUpsert(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var payload empresaSoporteRemotoConfigPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	current, err := dbpkg.GetEmpresaSoporteRemotoConfig(dbEmp, empresaID)
	if err != nil {
		http.Error(w, "No se pudo consultar configuracion actual", http.StatusInternalServerError)
		return
	}
	if payload.Habilitado != nil {
		current.Habilitado = *payload.Habilitado
	}
	if payload.RequiereAprobacionOperador != nil {
		current.RequiereAprobacionOperador = *payload.RequiereAprobacionOperador
	}
	if strings.TrimSpace(payload.ProveedorPreferido) != "" {
		current.ProveedorPreferido = payload.ProveedorPreferido
	}
	if strings.TrimSpace(payload.ModoOperacion) != "" {
		current.ModoOperacion = payload.ModoOperacion
	}
	if payload.AutoCerrarMinutos > 0 {
		current.AutoCerrarMinutos = payload.AutoCerrarMinutos
	}
	current.UsuarioCreador = adminEmailFromRequest(r)
	current.Observaciones = strings.TrimSpace(payload.Observaciones)
	if _, err := dbpkg.UpsertEmpresaSoporteRemotoConfig(dbEmp, current); err != nil {
		http.Error(w, "No se pudo guardar configuracion de soporte remoto: "+err.Error(), http.StatusBadRequest)
		return
	}
	cfg, err := dbpkg.GetEmpresaSoporteRemotoConfig(dbEmp, empresaID)
	if err != nil {
		http.Error(w, "Configuracion guardada, pero no se pudo consultar", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "config": cfg})
}

func empresaSoporteRemotoDispositivosGet(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	limit, err := parseIntQueryOptional(r, "limit")
	if err != nil {
		http.Error(w, "limit invalido", http.StatusBadRequest)
		return
	}
	offset, err := parseIntQueryOptional(r, "offset")
	if err != nil {
		http.Error(w, "offset invalido", http.StatusBadRequest)
		return
	}
	rows, total, err := dbpkg.ListEmpresaSoporteRemotoDispositivos(dbEmp, empresaID, dbpkg.EmpresaSoporteRemotoDispositivoFilter{
		IncludeInactive: queryBool(r, "include_inactive"),
		Q:               strings.TrimSpace(r.URL.Query().Get("q")),
		Limit:           limit,
		Offset:          offset,
	})
	if err != nil {
		http.Error(w, "No se pudo consultar dispositivos", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "total": total, "rows": rows})
}

func empresaSoporteRemotoDispositivoDetailGet(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	deviceID, err := parseInt64QueryOptional(r, "id")
	if err != nil || deviceID <= 0 {
		http.Error(w, "id es obligatorio", http.StatusBadRequest)
		return
	}
	item, err := dbpkg.GetEmpresaSoporteRemotoDispositivoByID(dbEmp, empresaID, deviceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "dispositivo no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo consultar dispositivo", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "dispositivo": item})
}

func empresaSoporteRemotoDispositivoCreate(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var payload empresaSoporteRemotoDispositivoPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	id, err := dbpkg.CreateEmpresaSoporteRemotoDispositivo(dbEmp, dbpkg.EmpresaSoporteRemotoDispositivo{
		EmpresaID:         empresaID,
		CodigoDispositivo: payload.CodigoDispositivo,
		NombreEquipo:      payload.NombreEquipo,
		AliasOperativo:    payload.AliasOperativo,
		Ubicacion:         payload.Ubicacion,
		SistemaOperativo:  payload.SistemaOperativo,
		AgenteVersion:     payload.AgenteVersion,
		StreamURL:         payload.StreamURL,
		EstadoConexion:    payload.EstadoConexion,
		UsuarioCreador:    adminEmailFromRequest(r),
		Estado:            "activo",
		Observaciones:     payload.Observaciones,
	}, payload.AccesoPIN)
	if err != nil {
		http.Error(w, "No se pudo crear dispositivo: "+err.Error(), http.StatusBadRequest)
		return
	}
	item, err := dbpkg.GetEmpresaSoporteRemotoDispositivoByID(dbEmp, empresaID, id)
	if err != nil {
		http.Error(w, "Dispositivo creado, pero no se pudo consultar", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "dispositivo": item})
}

func empresaSoporteRemotoDispositivoUpdate(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var payload empresaSoporteRemotoDispositivoPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.ID <= 0 {
		if qID, qErr := parseInt64QueryOptional(r, "id"); qErr == nil && qID > 0 {
			payload.ID = qID
		}
	}
	if payload.ID <= 0 {
		http.Error(w, "id es obligatorio", http.StatusBadRequest)
		return
	}
	err = dbpkg.UpdateEmpresaSoporteRemotoDispositivo(dbEmp, dbpkg.EmpresaSoporteRemotoDispositivo{
		ID:                payload.ID,
		EmpresaID:         empresaID,
		CodigoDispositivo: payload.CodigoDispositivo,
		NombreEquipo:      payload.NombreEquipo,
		AliasOperativo:    payload.AliasOperativo,
		Ubicacion:         payload.Ubicacion,
		SistemaOperativo:  payload.SistemaOperativo,
		AgenteVersion:     payload.AgenteVersion,
		StreamURL:         payload.StreamURL,
		EstadoConexion:    payload.EstadoConexion,
		UsuarioCreador:    adminEmailFromRequest(r),
		Estado:            "activo",
		Observaciones:     payload.Observaciones,
	}, payload.AccesoPIN)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "dispositivo no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo actualizar dispositivo: "+err.Error(), http.StatusBadRequest)
		return
	}
	item, err := dbpkg.GetEmpresaSoporteRemotoDispositivoByID(dbEmp, empresaID, payload.ID)
	if err != nil {
		http.Error(w, "Dispositivo actualizado, pero no se pudo consultar", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "dispositivo": item})
}

func empresaSoporteRemotoDispositivoToggle(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, estado string) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	deviceID, err := parseInt64QueryOptional(r, "id")
	if err != nil || deviceID <= 0 {
		http.Error(w, "id es obligatorio", http.StatusBadRequest)
		return
	}
	if err := dbpkg.SetEmpresaSoporteRemotoDispositivoEstadoByID(dbEmp, empresaID, deviceID, estado); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "dispositivo no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo cambiar estado de dispositivo", http.StatusInternalServerError)
		return
	}
	item, _ := dbpkg.GetEmpresaSoporteRemotoDispositivoByID(dbEmp, empresaID, deviceID)
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "dispositivo": item})
}

func empresaSoporteRemotoSesionesGet(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	limit, err := parseIntQueryOptional(r, "limit")
	if err != nil {
		http.Error(w, "limit invalido", http.StatusBadRequest)
		return
	}
	offset, err := parseIntQueryOptional(r, "offset")
	if err != nil {
		http.Error(w, "offset invalido", http.StatusBadRequest)
		return
	}
	rows, total, err := dbpkg.ListEmpresaSoporteRemotoSesiones(dbEmp, empresaID, dbpkg.EmpresaSoporteRemotoSessionFilter{
		IncludeInactive: queryBool(r, "include_inactive"),
		EstadoSesion:    strings.TrimSpace(r.URL.Query().Get("estado_sesion")),
		Q:               strings.TrimSpace(r.URL.Query().Get("q")),
		Limit:           limit,
		Offset:          offset,
	})
	if err != nil {
		http.Error(w, "No se pudo consultar sesiones de soporte remoto", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "total": total, "rows": rows})
}

func empresaSoporteRemotoSesionCreate(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cfg, err := dbpkg.GetEmpresaSoporteRemotoConfig(dbEmp, empresaID)
	if err != nil {
		http.Error(w, "No se pudo cargar configuracion de soporte remoto", http.StatusInternalServerError)
		return
	}
	if !cfg.Habilitado {
		http.Error(w, "El soporte remoto esta deshabilitado para esta empresa", http.StatusPreconditionFailed)
		return
	}

	var payload empresaSoporteRemotoSesionPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.DispositivoID <= 0 {
		http.Error(w, "dispositivo_id es obligatorio", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.OperadorEmail) != "" {
		if _, err := mail.ParseAddress(strings.TrimSpace(payload.OperadorEmail)); err != nil {
			http.Error(w, "operador_email invalido", http.StatusBadRequest)
			return
		}
	}

	session, err := dbpkg.CreateEmpresaSoporteRemotoSession(
		dbEmp,
		empresaID,
		payload.DispositivoID,
		adminEmailFromRequest(r),
		payload.OperadorNombre,
		payload.OperadorEmail,
		payload.Motivo,
		payload.DuracionMin,
		cfg.RequiereAprobacionOperador,
	)
	if err != nil {
		http.Error(w, "No se pudo crear sesion de soporte remoto: "+err.Error(), http.StatusBadRequest)
		return
	}

	viewerURL := empresaSoporteRemotoBuildViewerURL(r, empresaID, session.CodigoSesion, session.TokenVisualizacionRaw)
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"ok":         true,
		"session":    session,
		"viewer_url": viewerURL,
	})
}

func empresaSoporteRemotoSesionAprobar(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var payload empresaSoporteRemotoSesionPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	payload.CodigoSesion = strings.TrimSpace(payload.CodigoSesion)
	if payload.CodigoSesion == "" {
		payload.CodigoSesion = strings.TrimSpace(r.URL.Query().Get("codigo_sesion"))
	}
	if payload.CodigoSesion == "" {
		http.Error(w, "codigo_sesion es obligatorio", http.StatusBadRequest)
		return
	}
	estado := strings.TrimSpace(payload.EstadoSesion)
	if estado == "" {
		estado = "aprobada"
	}
	if err := dbpkg.SetEmpresaSoporteRemotoSessionEstadoByCodigo(dbEmp, empresaID, payload.CodigoSesion, estado, payload.Observaciones); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "sesion no encontrada", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo actualizar sesion", http.StatusInternalServerError)
		return
	}
	session, _ := dbpkg.GetEmpresaSoporteRemotoSessionByCodigo(dbEmp, empresaID, payload.CodigoSesion)
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "session": session})
}

func empresaSoporteRemotoSesionFinalizar(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var payload empresaSoporteRemotoSesionPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	payload.CodigoSesion = strings.TrimSpace(payload.CodigoSesion)
	if payload.CodigoSesion == "" {
		payload.CodigoSesion = strings.TrimSpace(r.URL.Query().Get("codigo_sesion"))
	}
	if payload.CodigoSesion == "" {
		http.Error(w, "codigo_sesion es obligatorio", http.StatusBadRequest)
		return
	}
	if err := dbpkg.SetEmpresaSoporteRemotoSessionEstadoByCodigo(dbEmp, empresaID, payload.CodigoSesion, "finalizada", payload.Observaciones); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "sesion no encontrada", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo finalizar sesion", http.StatusInternalServerError)
		return
	}
	session, _ := dbpkg.GetEmpresaSoporteRemotoSessionByCodigo(dbEmp, empresaID, payload.CodigoSesion)
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "session": session})
}

func empresaSoporteRemotoResolverVisualizacion(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	codigoSesion := strings.TrimSpace(r.URL.Query().Get("codigo_sesion"))
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if codigoSesion == "" || token == "" {
		http.Error(w, "codigo_sesion y token son obligatorios", http.StatusBadRequest)
		return
	}

	session, err := dbpkg.ResolveEmpresaSoporteRemotoViewerSession(dbEmp, empresaID, codigoSesion, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "sesion/token invalido", http.StatusUnauthorized)
			return
		}
		http.Error(w, "No se pudo resolver visualizacion", http.StatusInternalServerError)
		return
	}

	accesoPermitido := session.EstadoSesion == "activa" || session.EstadoSesion == "aprobada"
	motivoBloqueo := ""
	if !accesoPermitido {
		motivoBloqueo = "sesion no activa: " + session.EstadoSesion
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":               true,
		"session":          session,
		"acceso_permitido": accesoPermitido,
		"motivo_bloqueo":   motivoBloqueo,
		"embed_url":        session.URLVisualizacion,
		"proveedor_hint":   "Configura stream_url del dispositivo a una URL web embebible (noVNC/Guacamole/RustDesk Web)",
	})
}

func empresaSoporteRemotoSesionesExport(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	if format == "" {
		format = "json"
	}
	rows, total, err := dbpkg.ListEmpresaSoporteRemotoSesiones(dbEmp, empresaID, dbpkg.EmpresaSoporteRemotoSessionFilter{
		IncludeInactive: queryBool(r, "include_inactive"),
		EstadoSesion:    strings.TrimSpace(r.URL.Query().Get("estado_sesion")),
		Q:               strings.TrimSpace(r.URL.Query().Get("q")),
		Limit:           1000,
		Offset:          0,
	})
	if err != nil {
		http.Error(w, "No se pudo exportar sesiones", http.StatusInternalServerError)
		return
	}
	ds := empresaSoporteRemotoComposeSessionsDataset(empresaID, rows, total)
	if err := writeReportesDatasetExport(w, ds, format); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func empresaSoporteRemotoHeartbeat(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	var payload empresaSoporteRemotoHeartbeatPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID <= 0 {
		if empresaID, err := parseEmpresaIDQuery(r); err == nil && empresaID > 0 {
			payload.EmpresaID = empresaID
		}
	}
	if payload.EmpresaID <= 0 {
		http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.CodigoDispositivo) == "" {
		http.Error(w, "codigo_dispositivo es obligatorio", http.StatusBadRequest)
		return
	}
	item, err := dbpkg.RegisterEmpresaSoporteRemotoDispositivoHeartbeat(
		dbEmp,
		payload.EmpresaID,
		payload.CodigoDispositivo,
		payload.AccesoPIN,
		payload.StreamURL,
		payload.SistemaOperativo,
		payload.AgenteVersion,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "dispositivo no autorizado", http.StatusUnauthorized)
			return
		}
		http.Error(w, "No se pudo registrar heartbeat", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "dispositivo": item})
}

func empresaSoporteRemotoSesionAgentUpdate(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, action string) {
	var hb empresaSoporteRemotoHeartbeatPayload
	if err := json.NewDecoder(r.Body).Decode(&hb); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if hb.EmpresaID <= 0 || strings.TrimSpace(hb.CodigoDispositivo) == "" {
		http.Error(w, "empresa_id y codigo_dispositivo son obligatorios", http.StatusBadRequest)
		return
	}
	device, err := dbpkg.ValidateEmpresaSoporteRemotoDispositivoAccess(dbEmp, hb.EmpresaID, hb.CodigoDispositivo, hb.AccesoPIN)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "dispositivo no autorizado", http.StatusUnauthorized)
			return
		}
		http.Error(w, "No se pudo validar dispositivo", http.StatusInternalServerError)
		return
	}

	codigoSesion := strings.TrimSpace(r.URL.Query().Get("codigo_sesion"))
	if codigoSesion == "" {
		http.Error(w, "codigo_sesion es obligatorio", http.StatusBadRequest)
		return
	}
	session, err := dbpkg.GetEmpresaSoporteRemotoSessionByCodigo(dbEmp, hb.EmpresaID, codigoSesion)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "sesion no encontrada", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo consultar sesion", http.StatusInternalServerError)
		return
	}
	if session.DispositivoID != device.ID {
		http.Error(w, "sesion no corresponde al dispositivo", http.StatusForbidden)
		return
	}

	estadoDestino := "aprobada"
	if action == "finalizar_sesion" {
		estadoDestino = "finalizada"
	}
	if err := dbpkg.SetEmpresaSoporteRemotoSessionEstadoByCodigo(dbEmp, hb.EmpresaID, codigoSesion, estadoDestino, "actualizado por agente remoto"); err != nil {
		http.Error(w, "No se pudo actualizar estado de sesion", http.StatusInternalServerError)
		return
	}
	session, _ = dbpkg.GetEmpresaSoporteRemotoSessionByCodigo(dbEmp, hb.EmpresaID, codigoSesion)
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "session": session})
}
