package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	empresaStorageQuotaEnabledKey    = "empresa_storage.quota_enabled"
	empresaStorageDefaultLimitMBKey  = "empresa_storage.default_limit_mb"
	empresaStorageWarnPercentKey     = "empresa_storage.warn_percent"
	empresaStorageBlockUploadsKey    = "empresa_storage.block_uploads_over_limit"
	empresaStorageMaxUploadMBKey     = "empresa_storage.max_upload_mb"
	defaultEmpresaStorageLimitMB     = int64(1024)
	defaultEmpresaStorageWarnPercent = 80
	defaultEmpresaStorageMaxUploadMB = int64(10)
)

type empresaStorageConfig struct {
	QuotaEnabled   bool  `json:"quota_enabled"`
	DefaultLimitMB int64 `json:"default_limit_mb"`
	WarnPercent    int   `json:"warn_percent"`
	BlockUploads   bool  `json:"block_uploads_over_limit"`
	MaxUploadMB    int64 `json:"max_upload_mb"`
}

type empresaStorageUsage struct {
	EmpresaID      int64   `json:"empresa_id"`
	UsedBytes      int64   `json:"used_bytes"`
	LimitBytes     int64   `json:"limit_bytes"`
	UsedMB         float64 `json:"used_mb"`
	LimitMB        int64   `json:"limit_mb"`
	UsagePercent   float64 `json:"usage_percent"`
	Warn           bool    `json:"warn"`
	OverLimit      bool    `json:"over_limit"`
	Message        string  `json:"message"`
	QuotaEnabled   bool    `json:"quota_enabled"`
	CanUpload      bool    `json:"can_upload"`
	MaxUploadBytes int64   `json:"max_upload_bytes"`
	MaxUploadMB    int64   `json:"max_upload_mb"`
}

// EmpresaBuzonHandler gestiona buzon privado, campana y chat empresarial.
func EmpresaBuzonHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID := parseEmpresaIDFromContext(r)
		if empresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}
		actor, ok := resolveEmpresaBuzonRequestActor(w, r, dbEmp, dbSuper, empresaID)
		if !ok {
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "" {
			action = "resumen"
		}

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "resumen":
				messages, unread, err := empresaBuzonResumen(dbEmp, empresaID, actor)
				if err != nil {
					log.Printf("[empresa_buzon] resumen empresa_id=%d error: %v", empresaID, err)
					writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "No se pudo cargar el buzon"})
					return
				}
				chat, _ := dbpkg.ListEmpresaChatMensajes(dbEmp, empresaID, 12)
				usage := buildEmpresaStorageUsage(dbEmp, dbSuper, empresaID)
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "actor": actor, "unread": unread, "mensajes": messages, "chat": chat, "storage": usage})
			case "mensajes":
				includeRead := queryBool(r, "include_read")
				limit := parsePositiveIntDefault(r.URL.Query().Get("limit"), 50)
				messages, err := dbpkg.ListEmpresaBuzonMensajes(dbEmp, empresaID, actor, includeRead, limit)
				if err != nil {
					log.Printf("[empresa_buzon] mensajes empresa_id=%d error: %v", empresaID, err)
					writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "No se pudieron cargar los mensajes"})
					return
				}
				unread, _ := dbpkg.CountEmpresaBuzonUnread(dbEmp, empresaID, actor)
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "unread": unread, "mensajes": messages})
			case "chat":
				limit := parsePositiveIntDefault(r.URL.Query().Get("limit"), 60)
				chat, err := dbpkg.ListEmpresaChatMensajes(dbEmp, empresaID, limit)
				if err != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "No se pudo cargar el chat"})
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "chat": chat})
			case "usuarios":
				items, err := listEmpresaBuzonRecipients(dbEmp, dbSuper, empresaID, actor)
				if err != nil {
					log.Printf("[empresa_buzon] usuarios empresa_id=%d error: %v", empresaID, err)
					writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "No se pudieron cargar usuarios"})
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "usuarios": items})
			case "storage":
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "storage": buildEmpresaStorageUsage(dbEmp, dbSuper, empresaID)})
			default:
				http.Error(w, "accion no soportada", http.StatusBadRequest)
			}
		case http.MethodPost:
			contentType := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
			if strings.Contains(contentType, "multipart/form-data") || action == "adjunto" {
				handleEmpresaBuzonAttachmentUpload(w, r, dbEmp, dbSuper, empresaID, actor)
				return
			}
			switch action {
			case "chat":
				var payload struct {
					Mensaje string `json:"mensaje"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "payload invalido", http.StatusBadRequest)
					return
				}
				item, err := dbpkg.CreateEmpresaChatMensaje(dbEmp, empresaID, actor, payload.Mensaje)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, "chat_empresarial", "mensaje_chat_creado", "empresa_chat_mensajes", item.ID, http.StatusCreated, map[string]interface{}{
					"actor_tipo": actor.Tipo,
					"actor_ref":  actor.Ref,
				}, "mensaje creado en chat empresarial")
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "mensaje": item})
			case "mensaje":
				msg, err := createEmpresaBuzonManualMessageFromJSON(r, dbEmp, dbSuper, empresaID, actor)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				auditModulo := "buzon"
				auditAccion := "mensaje_enviado"
				if strings.EqualFold(strings.TrimSpace(msg.Tipo), "tarea") || strings.EqualFold(strings.TrimSpace(msg.Modulo), "tareas_buzon") {
					auditModulo = "tareas_buzon"
					auditAccion = "tarea_asignada"
				}
				registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, auditModulo, auditAccion, "empresa_buzon_mensajes", msg.ID, http.StatusCreated, map[string]interface{}{
					"destinatario_tipo": msg.DestinatarioTipo,
					"destinatario_ref":  msg.DestinatarioRef,
					"prioridad":         msg.Prioridad,
					"tarea_estado":      msg.TareaEstado,
					"tiene_enlace":      strings.TrimSpace(msg.EnlaceURL) != "",
				}, "mensaje interno o tarea enviado desde buzon")
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "mensaje": msg})
			case "finalizar_tarea":
				var payload struct {
					ID          int64  `json:"id"`
					Descripcion string `json:"descripcion"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "payload invalido", http.StatusBadRequest)
					return
				}
				msg, err := dbpkg.CompleteEmpresaBuzonTarea(dbEmp, empresaID, payload.ID, actor, payload.Descripcion)
				if err != nil {
					if err == sql.ErrNoRows {
						writeJSON(w, http.StatusNotFound, map[string]interface{}{"ok": false, "error": "tarea no encontrada o no asignada a este usuario"})
						return
					}
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, "tareas_buzon", "tarea_finalizada", "empresa_buzon_mensajes", msg.ID, http.StatusOK, map[string]interface{}{
					"actor_tipo":    actor.Tipo,
					"actor_ref":     actor.Ref,
					"tarea_estado":  msg.TareaEstado,
					"tiene_cierre":  strings.TrimSpace(payload.Descripcion) != "",
					"destinatario":  msg.DestinatarioRef,
					"remitente_ref": msg.RemitenteRef,
				}, "tarea finalizada desde buzon empresarial")
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "mensaje": msg})
			default:
				http.Error(w, "accion no soportada", http.StatusBadRequest)
			}
		case http.MethodPut:
			if action != "leer" {
				http.Error(w, "accion no soportada", http.StatusBadRequest)
				return
			}
			var payload struct {
				ID int64 `json:"id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload invalido", http.StatusBadRequest)
				return
			}
			if err := dbpkg.MarkEmpresaBuzonMensajeRead(dbEmp, empresaID, payload.ID, actor); err != nil {
				if err == sql.ErrNoRows {
					writeJSON(w, http.StatusNotFound, map[string]interface{}{"ok": false, "error": "mensaje no encontrado"})
					return
				}
				writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "No se pudo marcar como leido"})
				return
			}
			unread, _ := dbpkg.CountEmpresaBuzonUnread(dbEmp, empresaID, actor)
			registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, "buzon", "mensaje_marcado_leido", "empresa_buzon_mensajes", payload.ID, http.StatusOK, map[string]interface{}{
				"actor_tipo": actor.Tipo,
				"actor_ref":  actor.Ref,
				"unread":     unread,
			}, "mensaje marcado como leido por el usuario")
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "unread": unread})
		default:
			w.Header().Set("Allow", "GET, POST, PUT")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// SuperEmpresaStorageConfigHandler administra cuotas globales de almacenamiento por empresa.
func SuperEmpresaStorageConfigHandler(dbSuper, dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "empresas" {
				items := listEmpresaStorageUsages(dbEmp, dbSuper)
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "config": getEmpresaStorageConfig(dbSuper), "empresas": items})
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "config": getEmpresaStorageConfig(dbSuper)})
		case http.MethodPut, http.MethodPost:
			var payload empresaStorageConfig
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload invalido", http.StatusBadRequest)
				return
			}
			cfg := normalizeEmpresaStorageConfig(payload)
			if err := saveEmpresaStorageConfig(dbSuper, cfg, strings.TrimSpace(adminEmailFromRequest(r))); err != nil {
				http.Error(w, "No se pudo guardar configuracion: "+err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "config": cfg})
		case http.MethodDelete:
			empresaID := parsePositiveInt64(strings.TrimSpace(r.URL.Query().Get("empresa_id")))
			if empresaID <= 0 {
				http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
				return
			}
			if err := cleanupEmpresaBuzonOldFiles(dbEmp, empresaID, parsePositiveIntDefault(r.URL.Query().Get("days"), 180)); err != nil {
				http.Error(w, "No se pudo limpiar archivos: "+err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "storage": buildEmpresaStorageUsage(dbEmp, dbSuper, empresaID)})
		default:
			w.Header().Set("Allow", "GET, POST, PUT, DELETE")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func empresaBuzonResumen(dbEmp *sql.DB, empresaID int64, actor dbpkg.EmpresaBuzonActor) ([]dbpkg.EmpresaBuzonMensaje, int64, error) {
	messages, err := dbpkg.ListEmpresaBuzonMensajes(dbEmp, empresaID, actor, true, 10)
	if err != nil {
		return nil, 0, err
	}
	unread, err := dbpkg.CountEmpresaBuzonUnread(dbEmp, empresaID, actor)
	if err != nil {
		return nil, 0, err
	}
	return messages, unread, nil
}

func resolveEmpresaBuzonRequestActor(w http.ResponseWriter, r *http.Request, dbEmp, dbSuper *sql.DB, empresaID int64) (dbpkg.EmpresaBuzonActor, bool) {
	email := strings.TrimSpace(adminEmailFromRequest(r))
	actor, err := dbpkg.ResolveEmpresaBuzonActor(dbEmp, dbSuper, empresaID, email)
	if err != nil {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return dbpkg.EmpresaBuzonActor{}, false
	}
	return actor, true
}

func createEmpresaBuzonManualMessageFromJSON(r *http.Request, dbEmp, dbSuper *sql.DB, empresaID int64, actor dbpkg.EmpresaBuzonActor) (dbpkg.EmpresaBuzonMensaje, error) {
	var payload struct {
		DestinatarioRef   string `json:"destinatario_ref"`
		DestinatarioEmail string `json:"destinatario_email"`
		Titulo            string `json:"titulo"`
		Mensaje           string `json:"mensaje"`
		EnlaceURL         string `json:"enlace_url"`
		Tipo              string `json:"tipo"`
		Prioridad         string `json:"prioridad"`
		TareaVenceEn      string `json:"tarea_vence_en"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return dbpkg.EmpresaBuzonMensaje{}, fmt.Errorf("payload invalido")
	}
	recipient, err := resolveEmpresaBuzonRecipient(dbEmp, dbSuper, empresaID, payload.DestinatarioRef, payload.DestinatarioEmail)
	if err != nil {
		return dbpkg.EmpresaBuzonMensaje{}, err
	}
	tipo := strings.ToLower(strings.TrimSpace(payload.Tipo))
	if tipo != "tarea" {
		tipo = "interno"
	}
	modulo := "buzon"
	if tipo == "tarea" {
		modulo = "tareas_buzon"
	}
	return dbpkg.CreateEmpresaBuzonMensaje(dbEmp, dbpkg.EmpresaBuzonMensaje{
		EmpresaID:          empresaID,
		DestinatarioTipo:   recipient.Tipo,
		DestinatarioRef:    recipient.Ref,
		DestinatarioEmail:  recipient.Email,
		DestinatarioNombre: recipient.Nombre,
		RemitenteTipo:      actor.Tipo,
		RemitenteRef:       actor.Ref,
		RemitenteEmail:     actor.Email,
		RemitenteNombre:    actor.Nombre,
		Titulo:             payload.Titulo,
		Mensaje:            payload.Mensaje,
		Tipo:               tipo,
		Prioridad:          firstNonEmptyString(payload.Prioridad, "normal"),
		Modulo:             modulo,
		EnlaceURL:          payload.EnlaceURL,
		TareaEstado:        "pendiente",
		TareaVenceEn:       payload.TareaVenceEn,
		UsuarioCreador:     actor.Email,
	})
}

func handleEmpresaBuzonAttachmentUpload(w http.ResponseWriter, r *http.Request, dbEmp, dbSuper *sql.DB, empresaID int64, actor dbpkg.EmpresaBuzonActor) {
	cfg := getEmpresaStorageConfig(dbSuper)
	maxBytes := cfg.MaxUploadMB * 1024 * 1024
	if maxBytes <= 0 {
		maxBytes = defaultEmpresaStorageMaxUploadMB * 1024 * 1024
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes+(1<<20))
	if err := r.ParseMultipartForm(maxBytes + (1 << 20)); err != nil {
		http.Error(w, "payload multipart invalido o archivo demasiado grande", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("archivo")
	if err != nil {
		file, header, err = r.FormFile("adjunto")
	}
	if err != nil {
		http.Error(w, "archivo es obligatorio", http.StatusBadRequest)
		return
	}
	defer file.Close()
	if header.Size > maxBytes {
		http.Error(w, "archivo supera el maximo permitido", http.StatusBadRequest)
		return
	}
	usage := buildEmpresaStorageUsage(dbEmp, dbSuper, empresaID)
	if !usage.CanUpload || (usage.QuotaEnabled && usage.LimitBytes > 0 && usage.UsedBytes+header.Size > usage.LimitBytes && cfg.BlockUploads) {
		writeJSON(w, http.StatusInsufficientStorage, map[string]interface{}{"ok": false, "error": "La empresa alcanzo el limite de almacenamiento configurado", "storage": usage})
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	contentType := strings.TrimSpace(header.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = mime.TypeByExtension(ext)
	}
	if ext == "" {
		ext = inferExtension(contentType)
	}
	if ext == "" {
		ext = ".bin"
	}
	if !isAllowedAttachmentExt(ext) {
		http.Error(w, "extension de archivo no permitida", http.StatusBadRequest)
		return
	}

	mensajeID := parsePositiveInt64(strings.TrimSpace(r.FormValue("mensaje_id")))
	recipient, err := resolveEmpresaBuzonRecipient(dbEmp, dbSuper, empresaID, r.FormValue("destinatario_ref"), r.FormValue("destinatario_email"))
	if err != nil && mensajeID <= 0 {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if mensajeID <= 0 {
		tipo := strings.ToLower(strings.TrimSpace(r.FormValue("tipo")))
		if tipo != "tarea" {
			tipo = "interno"
		}
		modulo := "buzon"
		if tipo == "tarea" {
			modulo = "tareas_buzon"
		}
		msg, err := dbpkg.CreateEmpresaBuzonMensaje(dbEmp, dbpkg.EmpresaBuzonMensaje{
			EmpresaID:          empresaID,
			DestinatarioTipo:   recipient.Tipo,
			DestinatarioRef:    recipient.Ref,
			DestinatarioEmail:  recipient.Email,
			DestinatarioNombre: recipient.Nombre,
			RemitenteTipo:      actor.Tipo,
			RemitenteRef:       actor.Ref,
			RemitenteEmail:     actor.Email,
			RemitenteNombre:    actor.Nombre,
			Titulo:             firstNonEmptyString(r.FormValue("titulo"), "Mensaje con adjunto"),
			Mensaje:            firstNonEmptyString(r.FormValue("mensaje"), "Adjunto: "+strings.TrimSpace(header.Filename)),
			Tipo:               tipo,
			Prioridad:          firstNonEmptyString(r.FormValue("prioridad"), "normal"),
			Modulo:             modulo,
			EnlaceURL:          r.FormValue("enlace_url"),
			TareaEstado:        "pendiente",
			TareaVenceEn:       r.FormValue("tarea_vence_en"),
			UsuarioCreador:     actor.Email,
		})
		if err != nil {
			http.Error(w, "No se pudo crear el mensaje", http.StatusInternalServerError)
			return
		}
		mensajeID = msg.ID
	} else {
		msg, err := dbpkg.GetEmpresaBuzonMensajeByID(dbEmp, empresaID, mensajeID)
		if err != nil {
			http.Error(w, "mensaje no encontrado", http.StatusNotFound)
			return
		}
		if !empresaBuzonActorCanAttach(actor, msg) {
			http.Error(w, "no autorizado para adjuntar a este mensaje", http.StatusForbidden)
			return
		}
	}

	absDir, publicDir, _ := empresaUploadsSubdir(dbEmp, empresaID, "mensajeria", "buzon", fmt.Sprintf("mensaje_%d", mensajeID))
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		http.Error(w, "No se pudo preparar carpeta de adjuntos", http.StatusInternalServerError)
		return
	}
	fileName := fmt.Sprintf("adjunto_%d%s", time.Now().UnixNano(), ext)
	absPath := filepath.Join(absDir, fileName)
	out, err := os.Create(absPath)
	if err != nil {
		http.Error(w, "No se pudo crear archivo", http.StatusInternalServerError)
		return
	}
	size, err := io.Copy(out, file)
	closeErr := out.Close()
	if err != nil || closeErr != nil {
		_ = os.Remove(absPath)
		http.Error(w, "No se pudo guardar archivo", http.StatusInternalServerError)
		return
	}
	fileURL := strings.TrimRight(publicDir, "/") + "/" + fileName
	duracion, _ := parseFloat64FormOptional(r, "duracion_segundos")
	adj, err := dbpkg.CreateEmpresaBuzonAdjunto(dbEmp, dbpkg.EmpresaBuzonAdjunto{
		EmpresaID:        empresaID,
		MensajeID:        mensajeID,
		TipoArchivo:      inferAttachmentType(r.FormValue("tipo_archivo"), contentType, ext),
		NombreArchivo:    header.Filename,
		MimeType:         contentType,
		FileURL:          fileURL,
		TamanoBytes:      size,
		DuracionSegundos: duracion,
		UsuarioCreador:   actor.Email,
	})
	if err != nil {
		_ = os.Remove(absPath)
		http.Error(w, "No se pudo guardar metadata del adjunto", http.StatusInternalServerError)
		return
	}
	registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, "buzon", "adjunto_cargado", "empresa_buzon_adjuntos", adj.ID, http.StatusCreated, map[string]interface{}{
		"mensaje_id":        mensajeID,
		"tipo_archivo":      adj.TipoArchivo,
		"mime_type":         adj.MimeType,
		"tamano_bytes":      adj.TamanoBytes,
		"duracion_segundos": adj.DuracionSegundos,
	}, "adjunto cargado al buzon empresarial")
	writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "mensaje_id": mensajeID, "adjunto": adj, "storage": buildEmpresaStorageUsage(dbEmp, dbSuper, empresaID)})
}

func empresaBuzonActorCanAttach(actor dbpkg.EmpresaBuzonActor, msg dbpkg.EmpresaBuzonMensaje) bool {
	actorRef := strings.TrimSpace(actor.Ref)
	actorEmail := strings.ToLower(strings.TrimSpace(actor.Email))
	if strings.TrimSpace(msg.DestinatarioTipo) == strings.TrimSpace(actor.Tipo) && strings.TrimSpace(msg.DestinatarioRef) == actorRef {
		return true
	}
	if actorEmail != "" {
		if strings.EqualFold(strings.TrimSpace(msg.UsuarioCreador), actorEmail) || strings.EqualFold(strings.TrimSpace(msg.RemitenteEmail), actorEmail) {
			return true
		}
	}
	return false
}

func listEmpresaBuzonRecipients(dbEmp, dbSuper *sql.DB, empresaID int64, actor dbpkg.EmpresaBuzonActor) ([]map[string]interface{}, error) {
	out := make([]map[string]interface{}, 0)
	seen := map[string]bool{}
	add := func(tipo, ref, nombre, rol, email, estado string, id int64) {
		tipo = strings.ToLower(strings.TrimSpace(tipo))
		ref = strings.TrimSpace(ref)
		email = strings.ToLower(strings.TrimSpace(email))
		nombre = strings.TrimSpace(nombre)
		rol = strings.TrimSpace(rol)
		estado = strings.ToLower(strings.TrimSpace(estado))
		if tipo == "" || ref == "" {
			return
		}
		if estado == "eliminado" || estado == "borrado" || estado == "deleted" {
			return
		}
		key := tipo + ":" + strings.ToLower(ref)
		if seen[key] {
			return
		}
		seen[key] = true
		value := tipo + ":" + ref
		if tipo == "usuario" && id > 0 {
			value = "usuario:" + strconv.FormatInt(id, 10)
		}
		if nombre == "" {
			nombre = firstNonEmptyString(email, ref)
		}
		if rol == "" {
			rol = "usuario"
		}
		out = append(out, map[string]interface{}{
			"id":     id,
			"tipo":   tipo,
			"ref":    ref,
			"value":  value,
			"email":  email,
			"nombre": nombre,
			"rol":    rol,
			"estado": estado,
		})
	}

	users, err := dbpkg.GetEmpresaUsuarios(dbEmp, empresaID, true)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		estado := strings.ToLower(strings.TrimSpace(user.Estado))
		if estado == "" {
			estado = "activo"
		}
		if user.EmailConfirmado == 0 {
			estado = "pendiente"
		}
		add("usuario", strconv.FormatInt(user.ID, 10), user.Nombre, user.RolNombre, user.Email, estado, user.ID)
	}

	if dbEmp != nil {
		var ownerEmail string
		if err := dbEmp.QueryRow(`SELECT COALESCE(usuario_creador, '') FROM empresas WHERE id = ? LIMIT 1`, empresaID).Scan(&ownerEmail); err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		ownerEmail = strings.ToLower(strings.TrimSpace(ownerEmail))
		if ownerEmail != "" {
			name := ownerEmail
			role := "administrador"
			if dbSuper != nil {
				if admin, err := dbpkg.GetAdminByEmailFull(dbSuper, ownerEmail); err == nil && admin != nil {
					name = firstNonEmptyString(admin.Name, ownerEmail)
					role = firstNonEmptyString(admin.Role, role)
				}
			}
			add("admin", ownerEmail, name, role, ownerEmail, "activo", 0)
		}
	}

	if dbSuper != nil {
		accesos, err := dbpkg.ListAdminEmpresaCompartidaAccesosByEmpresa(dbSuper, empresaID)
		if err != nil {
			return nil, err
		}
		for _, acceso := range accesos {
			if !strings.EqualFold(strings.TrimSpace(acceso.Estado), "activo") || strings.TrimSpace(acceso.FechaRevocada) != "" {
				continue
			}
			email := strings.ToLower(strings.TrimSpace(acceso.AdminEmail))
			role := firstNonEmptyString(acceso.NivelAcceso, "administrador compartido")
			name := firstNonEmptyString(acceso.AdminName, email)
			if admin, err := dbpkg.GetAdminByEmailFull(dbSuper, email); err == nil && admin != nil {
				name = firstNonEmptyString(admin.Name, name)
				role = firstNonEmptyString(admin.Role, role)
			}
			add("admin", email, name, role, email, "activo", 0)
		}
	}

	if actor.Ref != "" {
		add(actor.Tipo, actor.Ref, actor.Nombre, actor.Rol, actor.Email, "activo", actor.UsuarioID)
	}
	return out, nil
}

func resolveEmpresaBuzonRecipient(dbEmp, dbSuper *sql.DB, empresaID int64, ref, email string) (dbpkg.EmpresaBuzonActor, error) {
	ref = strings.TrimSpace(ref)
	email = strings.ToLower(strings.TrimSpace(email))
	lookup := firstNonEmptyString(ref, email)
	if lookup == "" {
		return dbpkg.EmpresaBuzonActor{}, fmt.Errorf("destinatario es obligatorio")
	}

	tipoPrefix := ""
	if parts := strings.SplitN(lookup, ":", 2); len(parts) == 2 {
		tipoPrefix = strings.ToLower(strings.TrimSpace(parts[0]))
		lookup = strings.TrimSpace(parts[1])
	}
	if tipoPrefix == "usuario" || tipoPrefix == "" {
		if user, err := dbpkg.ResolveEmpresaUsuarioByReference(dbEmp, empresaID, lookup); err == nil && user != nil {
			return dbpkg.EmpresaBuzonActor{Tipo: "usuario", Ref: strconv.FormatInt(user.ID, 10), Email: strings.ToLower(strings.TrimSpace(user.Email)), Nombre: user.Nombre, Rol: user.RolNombre, UsuarioID: user.ID}, nil
		}
	}
	if tipoPrefix == "usuario" {
		return dbpkg.EmpresaBuzonActor{}, fmt.Errorf("destinatario no encontrado en la empresa")
	}
	if tipoPrefix == "admin" || strings.Contains(lookup, "@") {
		adminEmail := strings.ToLower(strings.TrimSpace(lookup))
		if adminEmail == "" || !strings.Contains(adminEmail, "@") {
			return dbpkg.EmpresaBuzonActor{}, fmt.Errorf("destinatario admin invalido")
		}
		allowed := false
		if dbSuper != nil {
			canAccess, err := dbpkg.CanAdminAccessEmpresaIA(dbEmp, dbSuper, adminEmail, empresaID)
			if err != nil {
				return dbpkg.EmpresaBuzonActor{}, err
			}
			allowed = canAccess
		}
		if !allowed && dbEmp != nil {
			var ownerEmail string
			if err := dbEmp.QueryRow(`SELECT COALESCE(usuario_creador, '') FROM empresas WHERE id = ? LIMIT 1`, empresaID).Scan(&ownerEmail); err != nil && err != sql.ErrNoRows {
				return dbpkg.EmpresaBuzonActor{}, err
			}
			allowed = strings.EqualFold(strings.TrimSpace(ownerEmail), adminEmail)
		}
		if !allowed {
			return dbpkg.EmpresaBuzonActor{}, fmt.Errorf("destinatario no pertenece al alcance de la empresa")
		}
		name := adminEmail
		role := "administrador"
		if dbSuper != nil {
			if admin, err := dbpkg.GetAdminByEmailFull(dbSuper, adminEmail); err == nil && admin != nil {
				name = firstNonEmptyString(admin.Name, adminEmail)
				role = firstNonEmptyString(admin.Role, role)
			}
		}
		return dbpkg.EmpresaBuzonActor{Tipo: "admin", Ref: adminEmail, Email: adminEmail, Nombre: name, Rol: role}, nil
	}
	if user, err := dbpkg.ResolveEmpresaUsuarioByReference(dbEmp, empresaID, lookup); err == nil && user != nil {
		return dbpkg.EmpresaBuzonActor{Tipo: "usuario", Ref: strconv.FormatInt(user.ID, 10), Email: strings.ToLower(strings.TrimSpace(user.Email)), Nombre: user.Nombre, Rol: user.RolNombre, UsuarioID: user.ID}, nil
	}
	return dbpkg.EmpresaBuzonActor{}, fmt.Errorf("destinatario no encontrado en la empresa")
}

func getEmpresaStorageConfig(dbSuper *sql.DB) empresaStorageConfig {
	cfg := empresaStorageConfig{QuotaEnabled: true, DefaultLimitMB: defaultEmpresaStorageLimitMB, WarnPercent: defaultEmpresaStorageWarnPercent, BlockUploads: true, MaxUploadMB: defaultEmpresaStorageMaxUploadMB}
	cfg.QuotaEnabled = readBoolConfigWithDefaultLocal(dbSuper, empresaStorageQuotaEnabledKey, cfg.QuotaEnabled)
	cfg.BlockUploads = readBoolConfigWithDefaultLocal(dbSuper, empresaStorageBlockUploadsKey, cfg.BlockUploads)
	cfg.DefaultLimitMB = readInt64ConfigWithDefaultLocal(dbSuper, empresaStorageDefaultLimitMBKey, cfg.DefaultLimitMB)
	cfg.MaxUploadMB = readInt64ConfigWithDefaultLocal(dbSuper, empresaStorageMaxUploadMBKey, cfg.MaxUploadMB)
	cfg.WarnPercent = int(readInt64ConfigWithDefaultLocal(dbSuper, empresaStorageWarnPercentKey, int64(cfg.WarnPercent)))
	return normalizeEmpresaStorageConfig(cfg)
}

func normalizeEmpresaStorageConfig(cfg empresaStorageConfig) empresaStorageConfig {
	if cfg.DefaultLimitMB <= 0 {
		cfg.DefaultLimitMB = defaultEmpresaStorageLimitMB
	}
	if cfg.DefaultLimitMB > 1024*1024 {
		cfg.DefaultLimitMB = 1024 * 1024
	}
	if cfg.WarnPercent < 50 || cfg.WarnPercent > 99 {
		cfg.WarnPercent = defaultEmpresaStorageWarnPercent
	}
	if cfg.MaxUploadMB <= 0 {
		cfg.MaxUploadMB = defaultEmpresaStorageMaxUploadMB
	}
	if cfg.MaxUploadMB > 100 {
		cfg.MaxUploadMB = 100
	}
	return cfg
}

func saveEmpresaStorageConfig(dbSuper *sql.DB, cfg empresaStorageConfig, actor string) error {
	if dbSuper == nil {
		return fmt.Errorf("dbSuper nil")
	}
	pairs := map[string]string{
		empresaStorageQuotaEnabledKey:   strconv.FormatBool(cfg.QuotaEnabled),
		empresaStorageDefaultLimitMBKey: strconv.FormatInt(cfg.DefaultLimitMB, 10),
		empresaStorageWarnPercentKey:    strconv.Itoa(cfg.WarnPercent),
		empresaStorageBlockUploadsKey:   strconv.FormatBool(cfg.BlockUploads),
		empresaStorageMaxUploadMBKey:    strconv.FormatInt(cfg.MaxUploadMB, 10),
	}
	for key, value := range pairs {
		if err := dbpkg.SetConfigValue(dbSuper, key, value, false); err != nil {
			return err
		}
	}
	if strings.TrimSpace(actor) != "" {
		_ = dbpkg.SetConfigValue(dbSuper, "empresa_storage.updated_by", strings.TrimSpace(actor), false)
	}
	return nil
}

func buildEmpresaStorageUsage(dbEmp, dbSuper *sql.DB, empresaID int64) empresaStorageUsage {
	cfg := getEmpresaStorageConfig(dbSuper)
	baseDir, _ := empresaUploadsBaseDir(dbEmp, empresaID)
	used := dirSizeBytes(baseDir)
	limit := cfg.DefaultLimitMB * 1024 * 1024
	pct := 0.0
	if limit > 0 {
		pct = (float64(used) / float64(limit)) * 100
	}
	usage := empresaStorageUsage{
		EmpresaID:      empresaID,
		UsedBytes:      used,
		LimitBytes:     limit,
		UsedMB:         float64(used) / 1024.0 / 1024.0,
		LimitMB:        cfg.DefaultLimitMB,
		UsagePercent:   pct,
		Warn:           cfg.QuotaEnabled && pct >= float64(cfg.WarnPercent),
		OverLimit:      cfg.QuotaEnabled && limit > 0 && used >= limit,
		QuotaEnabled:   cfg.QuotaEnabled,
		CanUpload:      true,
		MaxUploadBytes: cfg.MaxUploadMB * 1024 * 1024,
		MaxUploadMB:    cfg.MaxUploadMB,
	}
	if usage.Warn {
		usage.Message = "La empresa esta cerca del limite de almacenamiento."
	}
	if usage.OverLimit {
		usage.Message = "La empresa alcanzo el limite de almacenamiento."
		if cfg.BlockUploads {
			usage.CanUpload = false
		}
	}
	return usage
}

func dirSizeBytes(root string) int64 {
	root = strings.TrimSpace(root)
	if root == "" {
		return 0
	}
	var total int64
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return nil
		}
		if info, statErr := d.Info(); statErr == nil {
			total += info.Size()
		}
		return nil
	})
	return total
}

func listEmpresaStorageUsages(dbEmp, dbSuper *sql.DB) []empresaStorageUsage {
	if dbEmp == nil {
		return nil
	}
	empresas, err := dbpkg.GetEmpresas(dbEmp)
	if err != nil {
		return nil
	}
	out := make([]empresaStorageUsage, 0, len(empresas))
	for _, empresa := range empresas {
		id := empresa.EmpresaID
		if id <= 0 {
			id = empresa.ID
		}
		if id > 0 {
			out = append(out, buildEmpresaStorageUsage(dbEmp, dbSuper, id))
		}
	}
	return out
}

func cleanupEmpresaBuzonOldFiles(dbEmp *sql.DB, empresaID int64, days int) error {
	if days <= 0 {
		days = 180
	}
	absDir, _, _ := empresaUploadsSubdir(dbEmp, empresaID, "mensajeria", "buzon")
	if _, err := os.Stat(absDir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	return filepath.WalkDir(absDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return nil
		}
		info, statErr := d.Info()
		if statErr != nil {
			return nil
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(path)
		}
		return nil
	})
}

func readBoolConfigWithDefaultLocal(dbSuper *sql.DB, key string, fallback bool) bool {
	raw, _, _, _, err := dbpkg.GetConfigEntry(dbSuper, key)
	if err != nil {
		return fallback
	}
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "si", "yes", "on", "activo", "enabled":
		return true
	case "0", "false", "no", "off", "inactivo", "disabled":
		return false
	default:
		return fallback
	}
}

func readInt64ConfigWithDefaultLocal(dbSuper *sql.DB, key string, fallback int64) int64 {
	raw, _, _, _, err := dbpkg.GetConfigEntry(dbSuper, key)
	if err != nil {
		return fallback
	}
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return fallback
	}
	return value
}

func parsePositiveIntDefault(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
