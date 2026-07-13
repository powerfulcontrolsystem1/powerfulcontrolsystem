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
	Habilitado                 *bool   `json:"habilitado"`
	ProveedorPreferido         *string `json:"proveedor_preferido"`
	ModoOperacion              *string `json:"modo_operacion"`
	RequiereAprobacionOperador *bool   `json:"requiere_aprobacion_operador"`
	AutoCerrarMinutos          int     `json:"auto_cerrar_minutos"`
	MaxConexionesMes           int     `json:"max_conexiones_mes"`
	MaxMinutosMes              int     `json:"max_minutos_mes"`
	MaxMinutosDiaRustDesk      *int    `json:"max_minutos_dia_rustdesk"`
	MaxDispositivos            int     `json:"max_dispositivos"`
	PortalPublicoHabilitado    *bool   `json:"portal_publico_habilitado"`
	RustDeskServerHost         *string `json:"rustdesk_server_host"`
	RustDeskServerKey          *string `json:"rustdesk_server_key"`
	ClienteWindowsURL          *string `json:"cliente_windows_url"`
	ClienteLinuxURL            *string `json:"cliente_linux_url"`
	ClienteMacURL              *string `json:"cliente_mac_url"`
	ServidorWindowsURL         *string `json:"servidor_windows_url"`
	ServidorLinuxURL           *string `json:"servidor_linux_url"`
	ServidorMacURL             *string `json:"servidor_mac_url"`
	CarpetaTransferencia       *string `json:"carpeta_transferencia"`
	InstruccionesPublicas      *string `json:"instrucciones_publicas"`
	Observaciones              *string `json:"observaciones"`
}

type empresaSoporteRemotoDispositivoPayload struct {
	ID                      int64  `json:"id"`
	CodigoDispositivo       string `json:"codigo_dispositivo"`
	NombreEquipo            string `json:"nombre_equipo"`
	AliasOperativo          string `json:"alias_operativo"`
	Ubicacion               string `json:"ubicacion"`
	SistemaOperativo        string `json:"sistema_operativo"`
	AgenteVersion           string `json:"agente_version"`
	StreamURL               string `json:"stream_url"`
	RustDeskDeviceID        string `json:"rustdesk_device_id"`
	RustDeskPassword        string `json:"rustdesk_password"`
	CarpetaTransferencia    string `json:"carpeta_transferencia"`
	AccesoPublicoHabilitado *bool  `json:"acceso_publico_habilitado"`
	EstadoConexion          string `json:"estado_conexion"`
	AccesoPIN               string `json:"acceso_pin"`
	Observaciones           string `json:"observaciones"`
}

type empresaSoporteRemotoSesionPayload struct {
	DispositivoID  int64  `json:"dispositivo_id"`
	CodigoSesion   string `json:"codigo_sesion"`
	Role           string `json:"role"`
	OperadorNombre string `json:"operador_nombre"`
	OperadorEmail  string `json:"operador_email"`
	Motivo         string `json:"motivo"`
	DuracionMin    int    `json:"duracion_min"`
	EstadoSesion   string `json:"estado_sesion"`
	Observaciones  string `json:"observaciones"`
}

type empresaSoporteRemotoAccessBundle struct {
	Proveedor             string `json:"proveedor"`
	ModoOperacion         string `json:"modo_operacion"`
	EmbedURL              string `json:"embed_url,omitempty"`
	PortalPublicoURL      string `json:"portal_publico_url,omitempty"`
	RequiereCliente       bool   `json:"requiere_cliente"`
	RustDeskServerHost    string `json:"rustdesk_server_host,omitempty"`
	RustDeskServerKey     string `json:"rustdesk_server_key,omitempty"`
	RustDeskDeviceID      string `json:"rustdesk_device_id,omitempty"`
	RustDeskPassword      string `json:"rustdesk_password,omitempty"`
	ClienteWindowsURL     string `json:"cliente_windows_url,omitempty"`
	ClienteLinuxURL       string `json:"cliente_linux_url,omitempty"`
	ClienteMacURL         string `json:"cliente_mac_url,omitempty"`
	ServidorWindowsURL    string `json:"servidor_windows_url,omitempty"`
	ServidorLinuxURL      string `json:"servidor_linux_url,omitempty"`
	ServidorMacURL        string `json:"servidor_mac_url,omitempty"`
	CarpetaTransferencia  string `json:"carpeta_transferencia,omitempty"`
	InstruccionesPublicas string `json:"instrucciones_publicas,omitempty"`
}

const (
	rustDeskClientWindowsURL = "https://rustdesk.com/download"
	rustDeskClientLinuxURL   = "https://rustdesk.com/download"
	rustDeskClientMacURL     = "https://rustdesk.com/download"
	rustDeskServerWindowsURL = "https://rustdesk.com/docs/en/self-host/install/"
	rustDeskServerLinuxURL   = "https://rustdesk.com/docs/en/self-host/install/"
)

func applyRustDeskDownloadDefaults(cfg *dbpkg.EmpresaSoporteRemotoConfig) {
	if cfg == nil {
		return
	}
	if strings.TrimSpace(cfg.ClienteWindowsURL) == "" {
		cfg.ClienteWindowsURL = rustDeskClientWindowsURL
	}
	if strings.TrimSpace(cfg.ClienteLinuxURL) == "" {
		cfg.ClienteLinuxURL = rustDeskClientLinuxURL
	}
	if strings.TrimSpace(cfg.ClienteMacURL) == "" {
		cfg.ClienteMacURL = rustDeskClientMacURL
	}
	if strings.TrimSpace(cfg.ServidorWindowsURL) == "" {
		cfg.ServidorWindowsURL = rustDeskServerWindowsURL
	}
	if strings.TrimSpace(cfg.ServidorLinuxURL) == "" {
		cfg.ServidorLinuxURL = rustDeskServerLinuxURL
	}
	if strings.TrimSpace(cfg.ServidorMacURL) == "" {
		cfg.ServidorMacURL = ""
	}
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
	case "crear_token_senalizacion", "signaling_token":
		return "crear_token_senalizacion"
	case "export_sesiones", "export":
		return "export_sesiones"
	case "heartbeat_dispositivo", "heartbeat":
		return "heartbeat_dispositivo"
	case "resolver_acceso_publico", "acceso_publico":
		return "resolver_acceso_publico"
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

func empresaSoporteRemotoBuildPublicPortalURL(r *http.Request, empresaID int64, codigoSesion, token string) string {
	if r == nil {
		return ""
	}
	base := &url.URL{Path: "/soporte_remoto_acceso.html"}
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

func empresaSoporteRemotoBuildAccessBundle(r *http.Request, cfg dbpkg.EmpresaSoporteRemotoConfig, device dbpkg.EmpresaSoporteRemotoDispositivo, empresaID int64, codigoSesion, token string) empresaSoporteRemotoAccessBundle {
	bundle := empresaSoporteRemotoAccessBundle{
		Proveedor:             cfg.ProveedorPreferido,
		ModoOperacion:         cfg.ModoOperacion,
		EmbedURL:              strings.TrimSpace(device.StreamURL),
		PortalPublicoURL:      empresaSoporteRemotoBuildPublicPortalURL(r, empresaID, codigoSesion, token),
		RustDeskServerHost:    strings.TrimSpace(cfg.RustDeskServerHost),
		RustDeskServerKey:     strings.TrimSpace(cfg.RustDeskServerKey),
		RustDeskDeviceID:      strings.TrimSpace(device.RustDeskDeviceID),
		ClienteWindowsURL:     strings.TrimSpace(cfg.ClienteWindowsURL),
		ClienteLinuxURL:       strings.TrimSpace(cfg.ClienteLinuxURL),
		ClienteMacURL:         strings.TrimSpace(cfg.ClienteMacURL),
		ServidorWindowsURL:    strings.TrimSpace(cfg.ServidorWindowsURL),
		ServidorLinuxURL:      strings.TrimSpace(cfg.ServidorLinuxURL),
		ServidorMacURL:        strings.TrimSpace(cfg.ServidorMacURL),
		CarpetaTransferencia:  strings.TrimSpace(device.CarpetaTransferencia),
		InstruccionesPublicas: strings.TrimSpace(cfg.InstruccionesPublicas),
	}
	if bundle.CarpetaTransferencia == "" {
		bundle.CarpetaTransferencia = strings.TrimSpace(cfg.CarpetaTransferencia)
	}
	if bundle.RustDeskDeviceID != "" {
		bundle.RequiereCliente = true
		if password, err := dbpkg.ResolveEmpresaSoporteRemotoRustDeskPassword(device); err == nil {
			bundle.RustDeskPassword = strings.TrimSpace(password)
		}
	}
	if bundle.EmbedURL != "" {
		bundle.RequiereCliente = false
	}
	return bundle
}

func empresaSoporteRemotoApplyConfigPayload(current *dbpkg.EmpresaSoporteRemotoConfig, payload empresaSoporteRemotoConfigPayload, actor string) {
	if current == nil {
		return
	}
	if payload.Habilitado != nil {
		current.Habilitado = *payload.Habilitado
	}
	if payload.RequiereAprobacionOperador != nil {
		current.RequiereAprobacionOperador = *payload.RequiereAprobacionOperador
	}
	if payload.ProveedorPreferido != nil {
		current.ProveedorPreferido = strings.TrimSpace(*payload.ProveedorPreferido)
	}
	if payload.ModoOperacion != nil {
		current.ModoOperacion = strings.TrimSpace(*payload.ModoOperacion)
	}
	if payload.AutoCerrarMinutos > 0 {
		current.AutoCerrarMinutos = payload.AutoCerrarMinutos
	}
	if payload.MaxConexionesMes >= 0 {
		current.MaxConexionesMes = payload.MaxConexionesMes
	}
	if payload.MaxMinutosMes >= 0 {
		current.MaxMinutosMes = payload.MaxMinutosMes
	}
	if payload.MaxMinutosDiaRustDesk != nil && *payload.MaxMinutosDiaRustDesk >= 0 {
		current.MaxMinutosDiaRustDesk = *payload.MaxMinutosDiaRustDesk
	}
	if payload.MaxDispositivos >= 0 {
		current.MaxDispositivos = payload.MaxDispositivos
	}
	if payload.PortalPublicoHabilitado != nil {
		current.PortalPublicoHabilitado = *payload.PortalPublicoHabilitado
	}
	if payload.RustDeskServerHost != nil {
		current.RustDeskServerHost = strings.TrimSpace(*payload.RustDeskServerHost)
	}
	if payload.RustDeskServerKey != nil {
		current.RustDeskServerKey = strings.TrimSpace(*payload.RustDeskServerKey)
	}
	if payload.ClienteWindowsURL != nil {
		current.ClienteWindowsURL = strings.TrimSpace(*payload.ClienteWindowsURL)
	}
	if payload.ClienteLinuxURL != nil {
		current.ClienteLinuxURL = strings.TrimSpace(*payload.ClienteLinuxURL)
	}
	if payload.ClienteMacURL != nil {
		current.ClienteMacURL = strings.TrimSpace(*payload.ClienteMacURL)
	}
	if payload.ServidorWindowsURL != nil {
		current.ServidorWindowsURL = strings.TrimSpace(*payload.ServidorWindowsURL)
	}
	if payload.ServidorLinuxURL != nil {
		current.ServidorLinuxURL = strings.TrimSpace(*payload.ServidorLinuxURL)
	}
	if payload.ServidorMacURL != nil {
		current.ServidorMacURL = strings.TrimSpace(*payload.ServidorMacURL)
	}
	if payload.CarpetaTransferencia != nil {
		current.CarpetaTransferencia = strings.TrimSpace(*payload.CarpetaTransferencia)
	}
	if payload.InstruccionesPublicas != nil {
		current.InstruccionesPublicas = strings.TrimSpace(*payload.InstruccionesPublicas)
	}
	if payload.Observaciones != nil {
		current.Observaciones = strings.TrimSpace(*payload.Observaciones)
	}
	current.UsuarioCreador = strings.TrimSpace(actor)
}

func empresaSoporteRemotoComposeSessionsDataset(empresaID int64, rows []dbpkg.EmpresaSoporteRemotoSession, total int64) empresaReporteDataset {
	datasetRows := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		datasetRows = append(datasetRows, map[string]interface{}{
			"empresa_id":              row.EmpresaID,
			"codigo_sesion":           row.CodigoSesion,
			"dispositivo_id":          row.DispositivoID,
			"dispositivo_codigo":      row.DispositivoCodigo,
			"dispositivo_nombre":      row.DispositivoNombre,
			"solicitada_por":          row.SolicitadaPor,
			"operador_nombre":         row.OperadorNombre,
			"operador_email":          row.OperadorEmail,
			"estado_sesion":           row.EstadoSesion,
			"duracion_min_solicitada": row.DuracionMinSolicitada,
			"duracion_min_consumida":  row.DuracionMinConsumida,
			"bloqueada_por_limite":    row.BloqueadaPorLimite,
			"motivo":                  row.Motivo,
			"url_visualizacion":       empresaSoporteRemotoMaskStreamURL(row.URLVisualizacion),
			"iniciada_en":             row.IniciadaEn,
			"expira_en":               row.ExpiraEn,
			"finalizada_en":           row.FinalizadaEn,
			"fecha_creacion":          row.FechaCreacion,
			"fecha_actualizacion":     row.FechaActualizacion,
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
			"duracion_min_solicitada",
			"duracion_min_consumida",
			"bloqueada_por_limite",
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
			"incluye_consumo_plan":   true,
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
			case "crear_token_senalizacion":
				empresaSoporteRemotoSignalingCredentialCreate(w, r, dbEmp)
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

func empresaSoporteRemotoSignalingCredentialCreate(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID := parseEmpresaIDFromContext(r)
	if empresaID <= 0 {
		http.Error(w, "contexto de empresa requerido", http.StatusUnauthorized)
		return
	}
	var payload empresaSoporteRemotoSesionPayload
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	payload.CodigoSesion = strings.TrimSpace(payload.CodigoSesion)
	if payload.CodigoSesion == "" {
		http.Error(w, "codigo_sesion es obligatorio", http.StatusBadRequest)
		return
	}
	credential, err := dbpkg.CreateEmpresaSoporteRemotoSignalingCredential(
		dbEmp,
		empresaID,
		payload.CodigoSesion,
		payload.Role,
		adminEmailFromRequest(r),
	)
	if err != nil {
		if errors.Is(err, dbpkg.ErrSoporteRemotoSignalingCredential) || errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "sesion no disponible para senalizacion", http.StatusForbidden)
			return
		}
		http.Error(w, "No se pudo crear credencial de senalizacion", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "credential": credential})
}

// PublicEmpresaSoporteRemotoAgentHandler expone operaciones de heartbeat/estado para plugin de agencia remota.
func PublicEmpresaSoporteRemotoAgentHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		action := empresaSoporteRemotoNormalizeAction(r.URL.Query().Get("action"))
		switch action {
		case "resolver_acceso_publico":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			empresaSoporteRemotoResolverAccesoPublico(w, r, dbEmp)
		case "heartbeat_dispositivo":
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			empresaSoporteRemotoHeartbeat(w, r, dbEmp)
		case "aprobar_sesion", "finalizar_sesion":
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
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
	applyRustDeskDownloadDefaults(&cfg)
	uso, err := dbpkg.GetEmpresaSoporteRemotoUso(dbEmp, empresaID)
	if err != nil {
		http.Error(w, "No se pudo consultar consumo de soporte remoto", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "config": cfg, "uso": uso})
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
	empresaSoporteRemotoApplyConfigPayload(&current, payload, adminEmailFromRequest(r))
	if _, err := dbpkg.UpsertEmpresaSoporteRemotoConfig(dbEmp, current); err != nil {
		http.Error(w, "No se pudo guardar configuracion de soporte remoto: "+err.Error(), http.StatusBadRequest)
		return
	}
	cfg, err := dbpkg.GetEmpresaSoporteRemotoConfig(dbEmp, empresaID)
	if err != nil {
		http.Error(w, "Configuracion guardada, pero no se pudo consultar", http.StatusInternalServerError)
		return
	}
	applyRustDeskDownloadDefaults(&cfg)
	uso, err := dbpkg.GetEmpresaSoporteRemotoUso(dbEmp, empresaID)
	if err != nil {
		http.Error(w, "Configuracion guardada, pero no se pudo consultar el consumo", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "config": cfg, "uso": uso})
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
		EmpresaID:               empresaID,
		CodigoDispositivo:       payload.CodigoDispositivo,
		NombreEquipo:            payload.NombreEquipo,
		AliasOperativo:          payload.AliasOperativo,
		Ubicacion:               payload.Ubicacion,
		SistemaOperativo:        payload.SistemaOperativo,
		AgenteVersion:           payload.AgenteVersion,
		StreamURL:               payload.StreamURL,
		RustDeskDeviceID:        payload.RustDeskDeviceID,
		RustDeskPasswordEnc:     payload.RustDeskPassword,
		CarpetaTransferencia:    payload.CarpetaTransferencia,
		AccesoPublicoHabilitado: payload.AccesoPublicoHabilitado == nil || *payload.AccesoPublicoHabilitado,
		EstadoConexion:          payload.EstadoConexion,
		UsuarioCreador:          adminEmailFromRequest(r),
		Estado:                  "activo",
		Observaciones:           payload.Observaciones,
	}, payload.AccesoPIN)
	if err != nil {
		if errors.Is(err, dbpkg.ErrSoporteRemotoPlanLimit) {
			uso, _ := dbpkg.GetEmpresaSoporteRemotoUso(dbEmp, empresaID)
			writeJSON(w, http.StatusPreconditionFailed, map[string]interface{}{"ok": false, "error": err.Error(), "uso": uso})
			return
		}
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
		ID:                      payload.ID,
		EmpresaID:               empresaID,
		CodigoDispositivo:       payload.CodigoDispositivo,
		NombreEquipo:            payload.NombreEquipo,
		AliasOperativo:          payload.AliasOperativo,
		Ubicacion:               payload.Ubicacion,
		SistemaOperativo:        payload.SistemaOperativo,
		AgenteVersion:           payload.AgenteVersion,
		StreamURL:               payload.StreamURL,
		RustDeskDeviceID:        payload.RustDeskDeviceID,
		RustDeskPasswordEnc:     payload.RustDeskPassword,
		CarpetaTransferencia:    payload.CarpetaTransferencia,
		AccesoPublicoHabilitado: payload.AccesoPublicoHabilitado == nil || *payload.AccesoPublicoHabilitado,
		EstadoConexion:          payload.EstadoConexion,
		UsuarioCreador:          adminEmailFromRequest(r),
		Estado:                  "activo",
		Observaciones:           payload.Observaciones,
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
	uso, _ := dbpkg.GetEmpresaSoporteRemotoUso(dbEmp, empresaID)
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "total": total, "rows": rows, "uso": uso})
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
		if errors.Is(err, dbpkg.ErrSoporteRemotoPlanLimit) {
			uso, _ := dbpkg.GetEmpresaSoporteRemotoUso(dbEmp, empresaID)
			writeJSON(w, http.StatusPreconditionFailed, map[string]interface{}{"ok": false, "error": err.Error(), "uso": uso})
			return
		}
		http.Error(w, "No se pudo crear sesion de soporte remoto: "+err.Error(), http.StatusBadRequest)
		return
	}

	viewerURL := empresaSoporteRemotoBuildViewerURL(r, empresaID, session.CodigoSesion, session.TokenVisualizacionRaw)
	uso, _ := dbpkg.GetEmpresaSoporteRemotoUso(dbEmp, empresaID)
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"ok":                 true,
		"session":            session,
		"viewer_url":         viewerURL,
		"portal_publico_url": empresaSoporteRemotoBuildPublicPortalURL(r, empresaID, session.CodigoSesion, session.TokenVisualizacionRaw),
		"uso":                uso,
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
		if errors.Is(err, dbpkg.ErrSoporteRemotoPlanLimit) {
			uso, _ := dbpkg.GetEmpresaSoporteRemotoUso(dbEmp, empresaID)
			writeJSON(w, http.StatusPreconditionFailed, map[string]interface{}{"ok": false, "error": err.Error(), "uso": uso})
			return
		}
		http.Error(w, "No se pudo actualizar sesion", http.StatusInternalServerError)
		return
	}
	session, _ := dbpkg.GetEmpresaSoporteRemotoSessionByCodigo(dbEmp, empresaID, payload.CodigoSesion)
	uso, _ := dbpkg.GetEmpresaSoporteRemotoUso(dbEmp, empresaID)
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "session": session, "uso": uso})
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
	w.Header().Set("Cache-Control", "no-store")
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
	access := empresaSoporteRemotoAccessBundle{EmbedURL: session.URLVisualizacion}
	if cfg, cfgErr := dbpkg.GetEmpresaSoporteRemotoConfig(dbEmp, empresaID); cfgErr == nil {
		if device, deviceErr := dbpkg.GetEmpresaSoporteRemotoDispositivoByID(dbEmp, empresaID, session.DispositivoID); deviceErr == nil {
			access = empresaSoporteRemotoBuildAccessBundle(r, cfg, device, empresaID, session.CodigoSesion, token)
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":               true,
		"session":          session,
		"acceso_permitido": accesoPermitido,
		"motivo_bloqueo":   motivoBloqueo,
		"embed_url":        session.URLVisualizacion,
		"access":           access,
		"proveedor_hint":   "Configura stream_url del dispositivo a una URL web embebible (noVNC/Guacamole/RustDesk Web)",
	})
}

func empresaSoporteRemotoResolverAccesoPublico(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	w.Header().Set("Cache-Control", "no-store")
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
		http.Error(w, "No se pudo resolver acceso remoto", http.StatusInternalServerError)
		return
	}
	cfg, cfgErr := dbpkg.GetEmpresaSoporteRemotoConfig(dbEmp, empresaID)
	if cfgErr != nil {
		http.Error(w, "No se pudo cargar configuracion de soporte remoto", http.StatusInternalServerError)
		return
	}
	if !cfg.PortalPublicoHabilitado {
		http.Error(w, "portal publico deshabilitado", http.StatusForbidden)
		return
	}
	device, deviceErr := dbpkg.GetEmpresaSoporteRemotoDispositivoByID(dbEmp, empresaID, session.DispositivoID)
	if deviceErr != nil {
		http.Error(w, "No se pudo cargar el dispositivo remoto", http.StatusInternalServerError)
		return
	}
	access := empresaSoporteRemotoBuildAccessBundle(r, cfg, device, empresaID, session.CodigoSesion, token)
	accesoPermitido := session.EstadoSesion == "activa" || session.EstadoSesion == "aprobada"
	motivoBloqueo := ""
	if !accesoPermitido {
		motivoBloqueo = "sesion no activa: " + session.EstadoSesion
	}
	if !device.AccesoPublicoHabilitado {
		accesoPermitido = false
		motivoBloqueo = "dispositivo con acceso publico deshabilitado"
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":               true,
		"session":          session,
		"acceso_permitido": accesoPermitido,
		"motivo_bloqueo":   motivoBloqueo,
		"access":           access,
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
		if errors.Is(err, dbpkg.ErrSoporteRemotoPlanLimit) {
			uso, _ := dbpkg.GetEmpresaSoporteRemotoUso(dbEmp, hb.EmpresaID)
			writeJSON(w, http.StatusPreconditionFailed, map[string]interface{}{"ok": false, "error": err.Error(), "uso": uso})
			return
		}
		http.Error(w, "No se pudo actualizar estado de sesion", http.StatusInternalServerError)
		return
	}
	session, _ = dbpkg.GetEmpresaSoporteRemotoSessionByCodigo(dbEmp, hb.EmpresaID, codigoSesion)
	uso, _ := dbpkg.GetEmpresaSoporteRemotoUso(dbEmp, hb.EmpresaID)
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "session": session, "uso": uso})
}
