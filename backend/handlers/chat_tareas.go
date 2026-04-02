package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func queryBool(r *http.Request, key string) bool {
	v := strings.TrimSpace(strings.ToLower(r.URL.Query().Get(key)))
	return v == "1" || v == "true" || v == "si" || v == "yes"
}

func safeAuthorName(name, fallback string) string {
	name = strings.TrimSpace(name)
	if name != "" {
		return name
	}
	return strings.TrimSpace(fallback)
}

func parseInt64FormOptional(r *http.Request, key string) (int64, error) {
	raw := strings.TrimSpace(r.FormValue(key))
	if raw == "" {
		return 0, nil
	}
	return strconv.ParseInt(raw, 10, 64)
}

func parseFloat64FormOptional(r *http.Request, key string) (float64, error) {
	raw := strings.TrimSpace(r.FormValue(key))
	if raw == "" {
		return 0, nil
	}
	return strconv.ParseFloat(raw, 64)
}

func inferAttachmentType(requested, contentType, ext string) string {
	req := strings.ToLower(strings.TrimSpace(requested))
	if req == "imagen" || req == "audio" || req == "archivo" || req == "otro" {
		return req
	}
	if strings.HasPrefix(strings.ToLower(contentType), "image/") {
		return "imagen"
	}
	if strings.HasPrefix(strings.ToLower(contentType), "audio/") {
		return "audio"
	}
	ext = strings.ToLower(strings.TrimSpace(ext))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg":
		return "imagen"
	case ".mp3", ".wav", ".ogg", ".m4a", ".webm", ".aac":
		return "audio"
	default:
		return "archivo"
	}
}

func inferExtension(contentType string) string {
	if contentType == "" {
		return ""
	}
	exts, _ := mime.ExtensionsByType(contentType)
	if len(exts) > 0 {
		return strings.ToLower(exts[0])
	}
	return ""
}

func isAllowedAttachmentExt(ext string) bool {
	allowed := map[string]bool{
		".png":  true,
		".jpg":  true,
		".jpeg": true,
		".gif":  true,
		".webp": true,
		".svg":  true,
		".mp3":  true,
		".wav":  true,
		".ogg":  true,
		".m4a":  true,
		".webm": true,
		".aac":  true,
		".pdf":  true,
		".txt":  true,
		".csv":  true,
		".json": true,
	}
	return allowed[strings.ToLower(strings.TrimSpace(ext))]
}

// EmpresaChatTareasConversacionesHandler gestiona conversaciones de chat/tareas por empresa.
func EmpresaChatTareasConversacionesHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			includeInactive := queryBool(r, "include_inactive")
			q := strings.TrimSpace(r.URL.Query().Get("q"))
			rows, err := dbpkg.GetChatConversaciones(dbEmp, empresaID, includeInactive, q)
			if err != nil {
				http.Error(w, "No se pudieron listar las conversaciones", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, rows)
			return

		case http.MethodPost:
			var payload struct {
				dbpkg.ChatConversacion
				Participantes []dbpkg.ChatParticipante `json:"participantes"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 {
				http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.Titulo) == "" {
				http.Error(w, "titulo es obligatorio", http.StatusBadRequest)
				return
			}

			adminEmail := strings.TrimSpace(adminEmailFromRequest(r))
			payload.UsuarioCreador = adminEmail
			if strings.TrimSpace(payload.EstadoConversacion) == "" {
				payload.EstadoConversacion = "abierta"
			}

			newID, err := dbpkg.CreateChatConversacion(dbEmp, payload.ChatConversacion)
			if err != nil {
				http.Error(w, "No se pudo crear la conversacion", http.StatusBadRequest)
				return
			}

			_, _ = dbpkg.CreateChatParticipante(dbEmp, dbpkg.ChatParticipante{
				EmpresaID:         payload.EmpresaID,
				ConversacionID:    newID,
				ParticipanteTipo:  "admin",
				ParticipanteRefID: 0,
				Nombre:            adminEmail,
				Email:             adminEmail,
				UsuarioCreador:    adminEmail,
				Estado:            "activo",
			})

			for _, p := range payload.Participantes {
				p.EmpresaID = payload.EmpresaID
				p.ConversacionID = newID
				p.UsuarioCreador = adminEmail
				if _, err := dbpkg.CreateChatParticipante(dbEmp, p); err != nil {
					http.Error(w, "Conversacion creada, pero fallo al registrar participantes", http.StatusInternalServerError)
					return
				}
			}

			writeJSON(w, http.StatusCreated, map[string]interface{}{
				"ok": true,
				"id": newID,
			})
			return

		case http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "activar" || action == "desactivar" {
				empresaID, errEmp := parseEmpresaIDQuery(r)
				if errEmp != nil {
					http.Error(w, errEmp.Error(), http.StatusBadRequest)
					return
				}
				id, errID := parseInt64Query(r, "id")
				if errID != nil {
					http.Error(w, errID.Error(), http.StatusBadRequest)
					return
				}
				estado := "activo"
				if action == "desactivar" {
					estado = "inactivo"
				}
				if err := dbpkg.SetChatConversacionEstado(dbEmp, empresaID, id, estado); err != nil {
					http.Error(w, "No se pudo actualizar el estado de la conversacion", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": estado})
				return
			}

			if action == "cerrar" || action == "reabrir" {
				empresaID, errEmp := parseEmpresaIDQuery(r)
				if errEmp != nil {
					http.Error(w, errEmp.Error(), http.StatusBadRequest)
					return
				}
				id, errID := parseInt64Query(r, "id")
				if errID != nil {
					http.Error(w, errID.Error(), http.StatusBadRequest)
					return
				}
				estado := "abierta"
				if action == "cerrar" {
					estado = "cerrada"
				}
				if err := dbpkg.SetChatConversacionOperacionEstado(dbEmp, empresaID, id, estado); err != nil {
					http.Error(w, "No se pudo actualizar el estado operativo de la conversacion", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado_conversacion": estado})
				return
			}

			var payload dbpkg.ChatConversacion
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.ID <= 0 || payload.EmpresaID <= 0 {
				http.Error(w, "id y empresa_id son obligatorios", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.Titulo) == "" {
				http.Error(w, "titulo es obligatorio", http.StatusBadRequest)
				return
			}
			if err := dbpkg.UpdateChatConversacion(dbEmp, payload); err != nil {
				http.Error(w, "No se pudo actualizar la conversacion", http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return

		case http.MethodDelete:
			empresaID, errEmp := parseEmpresaIDQuery(r)
			if errEmp != nil {
				http.Error(w, errEmp.Error(), http.StatusBadRequest)
				return
			}
			id, errID := parseInt64Query(r, "id")
			if errID != nil {
				http.Error(w, errID.Error(), http.StatusBadRequest)
				return
			}
			if err := dbpkg.DeleteChatConversacion(dbEmp, empresaID, id); err != nil {
				http.Error(w, "No se pudo eliminar la conversacion", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

// EmpresaChatTareasParticipantesHandler gestiona participantes por conversacion.
func EmpresaChatTareasParticipantesHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			conversacionID, errConv := parseInt64Query(r, "conversacion_id")
			if errConv != nil {
				http.Error(w, "conversacion_id es obligatorio", http.StatusBadRequest)
				return
			}
			includeInactive := queryBool(r, "include_inactive")
			rows, err := dbpkg.GetChatParticipantes(dbEmp, empresaID, conversacionID, includeInactive)
			if err != nil {
				http.Error(w, "No se pudieron listar participantes", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, rows)
			return

		case http.MethodPost:
			var payload dbpkg.ChatParticipante
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 || payload.ConversacionID <= 0 {
				http.Error(w, "empresa_id y conversacion_id son obligatorios", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.Email) == "" && payload.ParticipanteRefID <= 0 {
				http.Error(w, "email o participante_ref_id es obligatorio", http.StatusBadRequest)
				return
			}
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			id, err := dbpkg.CreateChatParticipante(dbEmp, payload)
			if err != nil {
				http.Error(w, "No se pudo crear el participante", http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			return

		case http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "activar" || action == "desactivar" {
				empresaID, errEmp := parseEmpresaIDQuery(r)
				if errEmp != nil {
					http.Error(w, errEmp.Error(), http.StatusBadRequest)
					return
				}
				conversacionID, errConv := parseInt64Query(r, "conversacion_id")
				if errConv != nil {
					http.Error(w, "conversacion_id es obligatorio", http.StatusBadRequest)
					return
				}
				id, errID := parseInt64Query(r, "id")
				if errID != nil {
					http.Error(w, "id es obligatorio", http.StatusBadRequest)
					return
				}
				estado := "activo"
				if action == "desactivar" {
					estado = "inactivo"
				}
				if err := dbpkg.SetChatParticipanteEstado(dbEmp, empresaID, conversacionID, id, estado); err != nil {
					http.Error(w, "No se pudo actualizar estado del participante", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": estado})
				return
			}

			var payload dbpkg.ChatParticipante
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.ID <= 0 || payload.EmpresaID <= 0 || payload.ConversacionID <= 0 {
				http.Error(w, "id, empresa_id y conversacion_id son obligatorios", http.StatusBadRequest)
				return
			}
			if err := dbpkg.UpdateChatParticipante(dbEmp, payload); err != nil {
				http.Error(w, "No se pudo actualizar el participante", http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return

		case http.MethodDelete:
			empresaID, errEmp := parseEmpresaIDQuery(r)
			if errEmp != nil {
				http.Error(w, errEmp.Error(), http.StatusBadRequest)
				return
			}
			conversacionID, errConv := parseInt64Query(r, "conversacion_id")
			if errConv != nil {
				http.Error(w, "conversacion_id es obligatorio", http.StatusBadRequest)
				return
			}
			id, errID := parseInt64Query(r, "id")
			if errID != nil {
				http.Error(w, "id es obligatorio", http.StatusBadRequest)
				return
			}
			if err := dbpkg.DeleteChatParticipante(dbEmp, empresaID, conversacionID, id); err != nil {
				http.Error(w, "No se pudo eliminar el participante", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

// EmpresaChatTareasMensajesHandler gestiona mensajes por conversacion.
func EmpresaChatTareasMensajesHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			conversacionID, errConv := parseInt64Query(r, "conversacion_id")
			if errConv != nil {
				http.Error(w, "conversacion_id es obligatorio", http.StatusBadRequest)
				return
			}
			includeInactive := queryBool(r, "include_inactive")
			limit, _ := parseIntQueryOptional(r, "limit")
			offset, _ := parseIntQueryOptional(r, "offset")
			rows, err := dbpkg.GetChatMensajes(dbEmp, empresaID, conversacionID, includeInactive, limit, offset)
			if err != nil {
				http.Error(w, "No se pudieron listar los mensajes", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, rows)
			return

		case http.MethodPost:
			var payload dbpkg.ChatMensaje
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 || payload.ConversacionID <= 0 {
				http.Error(w, "empresa_id y conversacion_id son obligatorios", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.Contenido) == "" {
				http.Error(w, "contenido es obligatorio", http.StatusBadRequest)
				return
			}
			adminEmail := strings.TrimSpace(adminEmailFromRequest(r))
			payload.UsuarioCreador = adminEmail
			payload.AutorEmail = strings.TrimSpace(payload.AutorEmail)
			if payload.AutorEmail == "" {
				payload.AutorEmail = adminEmail
			}
			payload.AutorNombre = safeAuthorName(payload.AutorNombre, payload.AutorEmail)
			if strings.TrimSpace(payload.TipoMensaje) == "" {
				payload.TipoMensaje = "texto"
			}
			id, err := dbpkg.CreateChatMensaje(dbEmp, payload)
			if err != nil {
				http.Error(w, "No se pudo crear el mensaje", http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			return

		case http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "activar" || action == "desactivar" {
				empresaID, errEmp := parseEmpresaIDQuery(r)
				if errEmp != nil {
					http.Error(w, errEmp.Error(), http.StatusBadRequest)
					return
				}
				conversacionID, errConv := parseInt64Query(r, "conversacion_id")
				if errConv != nil {
					http.Error(w, "conversacion_id es obligatorio", http.StatusBadRequest)
					return
				}
				id, errID := parseInt64Query(r, "id")
				if errID != nil {
					http.Error(w, "id es obligatorio", http.StatusBadRequest)
					return
				}
				estado := "activo"
				if action == "desactivar" {
					estado = "inactivo"
				}
				if err := dbpkg.SetChatMensajeEstado(dbEmp, empresaID, conversacionID, id, estado); err != nil {
					http.Error(w, "No se pudo actualizar estado del mensaje", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": estado})
				return
			}

			var payload dbpkg.ChatMensaje
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.ID <= 0 || payload.EmpresaID <= 0 || payload.ConversacionID <= 0 {
				http.Error(w, "id, empresa_id y conversacion_id son obligatorios", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.Contenido) == "" {
				http.Error(w, "contenido es obligatorio", http.StatusBadRequest)
				return
			}
			if err := dbpkg.UpdateChatMensaje(dbEmp, payload); err != nil {
				http.Error(w, "No se pudo actualizar el mensaje", http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return

		case http.MethodDelete:
			empresaID, errEmp := parseEmpresaIDQuery(r)
			if errEmp != nil {
				http.Error(w, errEmp.Error(), http.StatusBadRequest)
				return
			}
			conversacionID, errConv := parseInt64Query(r, "conversacion_id")
			if errConv != nil {
				http.Error(w, "conversacion_id es obligatorio", http.StatusBadRequest)
				return
			}
			id, errID := parseInt64Query(r, "id")
			if errID != nil {
				http.Error(w, "id es obligatorio", http.StatusBadRequest)
				return
			}
			if err := dbpkg.DeleteChatMensaje(dbEmp, empresaID, conversacionID, id); err != nil {
				http.Error(w, "No se pudo eliminar el mensaje", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

// EmpresaChatTareasAdjuntoUploadHandler sube adjuntos de imagen/audio/archivo y crea un mensaje asociado.
func EmpresaChatTareasAdjuntoUploadHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseMultipartForm(20 << 20); err != nil {
			http.Error(w, "invalid multipart payload", http.StatusBadRequest)
			return
		}

		empresaID, err := parseInt64Form(r, "empresa_id")
		if err != nil || empresaID <= 0 {
			http.Error(w, "empresa_id required", http.StatusBadRequest)
			return
		}
		conversacionID, err := parseInt64Form(r, "conversacion_id")
		if err != nil || conversacionID <= 0 {
			http.Error(w, "conversacion_id required", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("archivo")
		if err != nil {
			file, header, err = r.FormFile("adjunto")
		}
		if err != nil {
			http.Error(w, "archivo required", http.StatusBadRequest)
			return
		}
		defer file.Close()

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
			http.Error(w, "attachment extension not allowed", http.StatusBadRequest)
			return
		}

		webRoot := resolveWebRootDir()
		dir := filepath.Join(webRoot, "uploads", "chat_tareas", fmt.Sprintf("empresa_%d", empresaID), fmt.Sprintf("conversacion_%d", conversacionID))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			http.Error(w, "failed to prepare upload directory", http.StatusInternalServerError)
			return
		}

		fileName := fmt.Sprintf("mensaje_%d_%d%s", conversacionID, time.Now().UnixNano(), ext)
		absPath := filepath.Join(dir, fileName)
		out, err := os.Create(absPath)
		if err != nil {
			http.Error(w, "failed to create attachment file", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		size, err := io.Copy(out, file)
		if err != nil {
			http.Error(w, "failed to save attachment", http.StatusInternalServerError)
			return
		}

		adminEmail := strings.TrimSpace(adminEmailFromRequest(r))
		autorNombre := safeAuthorName(r.FormValue("autor_nombre"), adminEmail)
		autorEmail := strings.TrimSpace(r.FormValue("autor_email"))
		if autorEmail == "" {
			autorEmail = adminEmail
		}
		autorRefID, _ := parseInt64FormOptional(r, "autor_ref_id")
		contenido := strings.TrimSpace(r.FormValue("contenido"))
		if contenido == "" {
			contenido = "Adjunto: " + strings.TrimSpace(header.Filename)
		}

		msgID, err := dbpkg.CreateChatMensaje(dbEmp, dbpkg.ChatMensaje{
			EmpresaID:      empresaID,
			ConversacionID: conversacionID,
			AutorTipo:      strings.TrimSpace(r.FormValue("autor_tipo")),
			AutorRefID:     autorRefID,
			AutorNombre:    autorNombre,
			AutorEmail:     autorEmail,
			Contenido:      contenido,
			TipoMensaje:    "texto",
			UsuarioCreador: adminEmail,
			Estado:         "activo",
		})
		if err != nil {
			http.Error(w, "failed to create message", http.StatusInternalServerError)
			return
		}

		duracion, _ := parseFloat64FormOptional(r, "duracion_segundos")
		tipoAdjunto := inferAttachmentType(r.FormValue("tipo_archivo"), contentType, ext)
		fileURL := "/uploads/chat_tareas/empresa_" + strconv.FormatInt(empresaID, 10) + "/conversacion_" + strconv.FormatInt(conversacionID, 10) + "/" + fileName

		adjID, err := dbpkg.CreateChatAdjunto(dbEmp, dbpkg.ChatAdjunto{
			EmpresaID:        empresaID,
			MensajeID:        msgID,
			TipoArchivo:      tipoAdjunto,
			NombreArchivo:    strings.TrimSpace(header.Filename),
			MimeType:         contentType,
			FileURL:          fileURL,
			TamanoBytes:      size,
			DuracionSegundos: duracion,
			UsuarioCreador:   adminEmail,
			Estado:           "activo",
		})
		if err != nil {
			http.Error(w, "failed to create attachment metadata", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusCreated, map[string]interface{}{
			"ok":            true,
			"message_id":    msgID,
			"attachment_id": adjID,
			"file_url":      fileURL,
			"tipo_archivo":  tipoAdjunto,
			"tamano_bytes":  size,
		})
	}
}

// EmpresaChatTareasTareasHandler gestiona tareas por empresa.
func EmpresaChatTareasTareasHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			conversacionID, _ := parseInt64QueryOptional(r, "conversacion_id")
			includeInactive := queryBool(r, "include_inactive")
			estadoTarea := strings.TrimSpace(r.URL.Query().Get("estado_tarea"))
			q := strings.TrimSpace(r.URL.Query().Get("q"))
			rows, err := dbpkg.GetChatTareas(dbEmp, empresaID, conversacionID, includeInactive, estadoTarea, q)
			if err != nil {
				http.Error(w, "No se pudieron listar las tareas", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, rows)
			return

		case http.MethodPost:
			var payload dbpkg.ChatTarea
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 {
				http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.Titulo) == "" {
				http.Error(w, "titulo es obligatorio", http.StatusBadRequest)
				return
			}

			adminEmail := strings.TrimSpace(adminEmailFromRequest(r))
			payload.UsuarioCreador = adminEmail
			payload.CreadoPorTipo = "admin"
			payload.CreadoPorEmail = adminEmail
			if strings.TrimSpace(payload.EstadoTarea) == "" {
				payload.EstadoTarea = "pendiente"
			}

			id, err := dbpkg.CreateChatTarea(dbEmp, payload)
			if err != nil {
				http.Error(w, "No se pudo crear la tarea", http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			return

		case http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "activar" || action == "desactivar" {
				empresaID, errEmp := parseEmpresaIDQuery(r)
				if errEmp != nil {
					http.Error(w, errEmp.Error(), http.StatusBadRequest)
					return
				}
				id, errID := parseInt64Query(r, "id")
				if errID != nil {
					http.Error(w, "id es obligatorio", http.StatusBadRequest)
					return
				}
				estado := "activo"
				if action == "desactivar" {
					estado = "inactivo"
				}
				if err := dbpkg.SetChatTareaEstado(dbEmp, empresaID, id, estado); err != nil {
					http.Error(w, "No se pudo actualizar estado de la tarea", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": estado})
				return
			}

			if action == "completar" || action == "reabrir" || action == "cancelar" || action == "en_progreso" || action == "bloqueada" {
				empresaID, errEmp := parseEmpresaIDQuery(r)
				if errEmp != nil {
					http.Error(w, errEmp.Error(), http.StatusBadRequest)
					return
				}
				id, errID := parseInt64Query(r, "id")
				if errID != nil {
					http.Error(w, "id es obligatorio", http.StatusBadRequest)
					return
				}

				estado := "pendiente"
				porcentaje := 0
				switch action {
				case "completar":
					estado = "completada"
					porcentaje = 100
				case "reabrir":
					estado = "pendiente"
					porcentaje = 0
				case "cancelar":
					estado = "cancelada"
					porcentaje = 0
				case "en_progreso":
					estado = "en_progreso"
					porcentaje = 50
				case "bloqueada":
					estado = "bloqueada"
					porcentaje = 0
				}
				if pRaw := strings.TrimSpace(r.URL.Query().Get("porcentaje")); pRaw != "" {
					if pVal, err := strconv.Atoi(pRaw); err == nil {
						porcentaje = pVal
					}
				}

				if err := dbpkg.SetChatTareaWorkflowEstado(dbEmp, empresaID, id, estado, porcentaje); err != nil {
					http.Error(w, "No se pudo actualizar estado operativo de la tarea", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado_tarea": estado, "porcentaje_avance": porcentaje})
				return
			}

			var payload dbpkg.ChatTarea
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.ID <= 0 || payload.EmpresaID <= 0 {
				http.Error(w, "id y empresa_id son obligatorios", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.Titulo) == "" {
				http.Error(w, "titulo es obligatorio", http.StatusBadRequest)
				return
			}
			if err := dbpkg.UpdateChatTarea(dbEmp, payload); err != nil {
				http.Error(w, "No se pudo actualizar la tarea", http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return

		case http.MethodDelete:
			empresaID, errEmp := parseEmpresaIDQuery(r)
			if errEmp != nil {
				http.Error(w, errEmp.Error(), http.StatusBadRequest)
				return
			}
			id, errID := parseInt64Query(r, "id")
			if errID != nil {
				http.Error(w, "id es obligatorio", http.StatusBadRequest)
				return
			}
			if err := dbpkg.DeleteChatTarea(dbEmp, empresaID, id); err != nil {
				http.Error(w, "No se pudo eliminar la tarea", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}
