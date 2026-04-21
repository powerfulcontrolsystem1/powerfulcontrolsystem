package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type superSoporteRemotoSessionPayload struct {
	EmpresaID      int64  `json:"empresa_id"`
	DispositivoID  int64  `json:"dispositivo_id"`
	CodigoSesion   string `json:"codigo_sesion"`
	OperadorNombre string `json:"operador_nombre"`
	OperadorEmail  string `json:"operador_email"`
	Motivo         string `json:"motivo"`
	DuracionMin    int    `json:"duracion_min"`
	Observaciones  string `json:"observaciones"`
}

type superSoporteRemotoEmpresaResumen struct {
	Empresa dbpkg.Empresa                    `json:"empresa"`
	Config  dbpkg.EmpresaSoporteRemotoConfig `json:"config"`
	Uso     dbpkg.EmpresaSoporteRemotoUso    `json:"uso"`
}

func normalizeSuperSoporteRemotoAction(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "empresas", "resumen_empresas":
		return "empresas"
	case "config", "configuracion":
		return "config"
	case "dispositivos":
		return "dispositivos"
	case "sesiones":
		return "sesiones"
	case "reporte", "reportes", "export":
		return "reporte"
	case "solicitar_sesion":
		return "solicitar_sesion"
	case "aprobar_sesion":
		return "aprobar_sesion"
	case "finalizar_sesion":
		return "finalizar_sesion"
	default:
		return ""
	}
}

func SuperSoporteRemotoHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		action := normalizeSuperSoporteRemotoAction(r.URL.Query().Get("action"))
		if action == "" {
			http.Error(w, "action invalida", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "empresas":
				superSoporteRemotoEmpresasGet(w, r, dbEmp)
			case "config":
				superSoporteRemotoConfigGet(w, r, dbEmp)
			case "dispositivos":
				superSoporteRemotoDispositivosGet(w, r, dbEmp)
			case "sesiones":
				superSoporteRemotoSesionesGet(w, r, dbEmp)
			case "reporte":
				superSoporteRemotoReporteGet(w, r, dbEmp)
			default:
				http.Error(w, "action invalida", http.StatusBadRequest)
			}
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			switch action {
			case "config":
				superSoporteRemotoConfigUpsert(w, r, dbEmp)
			case "solicitar_sesion":
				superSoporteRemotoSesionCreate(w, r, dbEmp)
			case "aprobar_sesion":
				superSoporteRemotoSesionEstado(w, r, dbEmp, "aprobada")
			case "finalizar_sesion":
				superSoporteRemotoSesionEstado(w, r, dbEmp, "finalizada")
			default:
				http.Error(w, "action invalida", http.StatusBadRequest)
			}
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func superSoporteRemotoEmpresasGet(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresas, err := dbpkg.GetEmpresas(dbEmp)
	if err != nil {
		http.Error(w, "No se pudieron consultar empresas", http.StatusInternalServerError)
		return
	}
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	rows := make([]superSoporteRemotoEmpresaResumen, 0)
	for _, empresa := range empresas {
		if q != "" {
			haystack := strings.ToLower(strings.TrimSpace(empresa.Nombre + " " + empresa.Nit + " " + empresa.Observaciones))
			if !strings.Contains(haystack, q) {
				continue
			}
		}
		cfg, err := dbpkg.GetEmpresaSoporteRemotoConfig(dbEmp, empresa.EmpresaID)
		if err != nil {
			continue
		}
		applyRustDeskDownloadDefaults(&cfg)
		uso, err := dbpkg.GetEmpresaSoporteRemotoUso(dbEmp, empresa.EmpresaID)
		if err != nil {
			continue
		}
		rows = append(rows, superSoporteRemotoEmpresaResumen{Empresa: empresa, Config: cfg, Uso: uso})
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "total": len(rows), "rows": rows})
}

func superSoporteRemotoConfigGet(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
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

func superSoporteRemotoConfigUpsert(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
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

func superSoporteRemotoDispositivosGet(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
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
		http.Error(w, "No se pudieron consultar dispositivos", http.StatusInternalServerError)
		return
	}
	uso, _ := dbpkg.GetEmpresaSoporteRemotoUso(dbEmp, empresaID)
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "total": total, "rows": rows, "uso": uso})
}

func superSoporteRemotoSesionesGet(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
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
		http.Error(w, "No se pudieron consultar sesiones", http.StatusInternalServerError)
		return
	}
	uso, _ := dbpkg.GetEmpresaSoporteRemotoUso(dbEmp, empresaID)
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "total": total, "rows": rows, "uso": uso})
}

func superSoporteRemotoReporteGet(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	if format == "" {
		format = "json"
	}
	empresas, err := dbpkg.GetEmpresas(dbEmp)
	if err != nil {
		http.Error(w, "No se pudieron consultar empresas", http.StatusInternalServerError)
		return
	}
	rows := make([]map[string]interface{}, 0)
	for _, empresa := range empresas {
		haystack := strings.ToLower(strings.TrimSpace(empresa.Nombre + " " + empresa.Nit + " " + empresa.Observaciones))
		if q != "" && !strings.Contains(haystack, q) {
			continue
		}
		cfg, cfgErr := dbpkg.GetEmpresaSoporteRemotoConfig(dbEmp, empresa.EmpresaID)
		if cfgErr != nil {
			continue
		}
		uso, usoErr := dbpkg.GetEmpresaSoporteRemotoUso(dbEmp, empresa.EmpresaID)
		if usoErr != nil {
			continue
		}
		rows = append(rows, map[string]interface{}{
			"empresa_id": empresa.EmpresaID,
			"empresa": empresa.Nombre,
			"nit": empresa.Nit,
			"proveedor": cfg.ProveedorPreferido,
			"modo_operacion": cfg.ModoOperacion,
			"portal_publico_habilitado": cfg.PortalPublicoHabilitado,
			"rustdesk_server_host": cfg.RustDeskServerHost,
			"max_minutos_dia_rustdesk": cfg.MaxMinutosDiaRustDesk,
			"cliente_windows_url": cfg.ClienteWindowsURL,
			"cliente_linux_url": cfg.ClienteLinuxURL,
			"cliente_mac_url": cfg.ClienteMacURL,
			"servidor_windows_url": cfg.ServidorWindowsURL,
			"servidor_linux_url": cfg.ServidorLinuxURL,
			"servidor_mac_url": cfg.ServidorMacURL,
			"carpeta_transferencia": cfg.CarpetaTransferencia,
			"instrucciones_publicas": cfg.InstruccionesPublicas,
			"dia_referencia": uso.DiaReferencia,
			"minutos_consumidos_dia_rustdesk": uso.MinutosConsumidosDiaRustDesk,
			"minutos_disponibles_dia_rustdesk": uso.MinutosDisponiblesDiaRustDesk,
			"dispositivos_activos": uso.DispositivosActivos,
			"dispositivos_online": uso.DispositivosOnline,
			"sesiones_mes": uso.SesionesMes,
			"intentos_bloqueados_mes": uso.IntentosBloqueadosMes,
			"minutos_consumidos_mes": uso.MinutosConsumidosMes,
			"max_dispositivos": uso.MaxDispositivos,
			"max_conexiones_mes": uso.MaxConexionesMes,
			"max_minutos_mes": uso.MaxMinutosMes,
			"bloqueo_motivo": uso.BloqueoMotivo,
		})
	}
	ds := empresaReporteDataset{
		Key:         "super_soporte_remoto_multiempresa",
		Title:       "Soporte remoto multiempresa",
		Level:       "operativo",
		Description: "Resumen consolidado de soporte remoto, acceso tecnico y cupos por empresa.",
		GeneratedAt: time.Now().In(time.Local).Format("2006-01-02 15:04:05"),
		Columns: []string{
			"empresa_id", "empresa", "nit", "proveedor", "modo_operacion", "portal_publico_habilitado",
			"rustdesk_server_host", "max_minutos_dia_rustdesk", "cliente_windows_url", "cliente_linux_url", "cliente_mac_url", "servidor_windows_url", "servidor_linux_url", "servidor_mac_url", "carpeta_transferencia", "instrucciones_publicas",
			"dia_referencia", "minutos_consumidos_dia_rustdesk", "minutos_disponibles_dia_rustdesk",
			"dispositivos_activos", "dispositivos_online", "sesiones_mes", "intentos_bloqueados_mes",
			"minutos_consumidos_mes", "max_dispositivos", "max_conexiones_mes", "max_minutos_mes", "bloqueo_motivo",
		},
		Rows:     rows,
		RowCount: len(rows),
		Summary: map[string]interface{}{"empresas_exportadas": len(rows), "formato_trazable": "json,csv,txt,xls,pdf"},
	}
	if err := writeReportesDatasetExport(w, ds, format); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func superSoporteRemotoSesionCreate(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	var payload superSoporteRemotoSessionPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID <= 0 || payload.DispositivoID <= 0 {
		http.Error(w, "empresa_id y dispositivo_id son obligatorios", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.OperadorEmail) != "" {
		if _, err := mail.ParseAddress(strings.TrimSpace(payload.OperadorEmail)); err != nil {
			http.Error(w, "operador_email invalido", http.StatusBadRequest)
			return
		}
	}
	cfg, err := dbpkg.GetEmpresaSoporteRemotoConfig(dbEmp, payload.EmpresaID)
	if err != nil {
		http.Error(w, "No se pudo consultar la configuracion remota", http.StatusInternalServerError)
		return
	}
	if !cfg.Habilitado {
		http.Error(w, "El soporte remoto esta deshabilitado para esta empresa", http.StatusPreconditionFailed)
		return
	}
	session, err := dbpkg.CreateEmpresaSoporteRemotoSession(dbEmp, payload.EmpresaID, payload.DispositivoID, adminEmailFromRequest(r), payload.OperadorNombre, payload.OperadorEmail, payload.Motivo, payload.DuracionMin, cfg.RequiereAprobacionOperador)
	if err != nil {
		if errors.Is(err, dbpkg.ErrSoporteRemotoPlanLimit) {
			uso, _ := dbpkg.GetEmpresaSoporteRemotoUso(dbEmp, payload.EmpresaID)
			writeJSON(w, http.StatusPreconditionFailed, map[string]interface{}{"ok": false, "error": err.Error(), "uso": uso})
			return
		}
		http.Error(w, "No se pudo crear sesion", http.StatusBadRequest)
		return
	}
	viewerURL := empresaSoporteRemotoBuildViewerURL(r, payload.EmpresaID, session.CodigoSesion, session.TokenVisualizacionRaw)
	uso, _ := dbpkg.GetEmpresaSoporteRemotoUso(dbEmp, payload.EmpresaID)
	writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "session": session, "viewer_url": viewerURL, "portal_publico_url": empresaSoporteRemotoBuildPublicPortalURL(r, payload.EmpresaID, session.CodigoSesion, session.TokenVisualizacionRaw), "uso": uso})
}

func superSoporteRemotoSesionEstado(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, estado string) {
	var payload superSoporteRemotoSessionPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID <= 0 {
		if empresaID, err := parseEmpresaIDQuery(r); err == nil {
			payload.EmpresaID = empresaID
		}
	}
	payload.CodigoSesion = strings.TrimSpace(payload.CodigoSesion)
	if payload.CodigoSesion == "" {
		payload.CodigoSesion = strings.TrimSpace(r.URL.Query().Get("codigo_sesion"))
	}
	if payload.EmpresaID <= 0 || payload.CodigoSesion == "" {
		http.Error(w, "empresa_id y codigo_sesion son obligatorios", http.StatusBadRequest)
		return
	}
	if err := dbpkg.SetEmpresaSoporteRemotoSessionEstadoByCodigo(dbEmp, payload.EmpresaID, payload.CodigoSesion, estado, payload.Observaciones); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "sesion no encontrada", http.StatusNotFound)
			return
		}
		if errors.Is(err, dbpkg.ErrSoporteRemotoPlanLimit) {
			uso, _ := dbpkg.GetEmpresaSoporteRemotoUso(dbEmp, payload.EmpresaID)
			writeJSON(w, http.StatusPreconditionFailed, map[string]interface{}{"ok": false, "error": err.Error(), "uso": uso})
			return
		}
		http.Error(w, "No se pudo actualizar la sesion", http.StatusInternalServerError)
		return
	}
	session, _ := dbpkg.GetEmpresaSoporteRemotoSessionByCodigo(dbEmp, payload.EmpresaID, payload.CodigoSesion)
	uso, _ := dbpkg.GetEmpresaSoporteRemotoUso(dbEmp, payload.EmpresaID)
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "session": session, "uso": uso})
}