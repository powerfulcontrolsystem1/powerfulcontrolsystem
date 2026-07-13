package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func queryBool(r *http.Request, key string) bool {
	v := strings.TrimSpace(strings.ToLower(r.URL.Query().Get(key)))
	return v == "1" || v == "true" || v == "si" || v == "yes"
}

func EmpresaChatTareasArchivoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil || empresaID <= 0 {
			http.Error(w, "empresa_id invalido", http.StatusBadRequest)
			return
		}
		serveEmpresaPrivateFile(w, r, empresaID, "chat_tareas")
	}
}

func safeAuthorName(name, fallback string) string {
	name = strings.TrimSpace(name)
	if name != "" {
		return name
	}
	return strings.TrimSpace(fallback)
}

func chatConversacionExists(dbEmp *sql.DB, empresaID, conversacionID int64) (bool, error) {
	if dbEmp == nil || empresaID <= 0 || conversacionID <= 0 {
		return false, nil
	}
	var id int64
	err := dbpkg.QueryRowCompat(dbEmp, `SELECT id FROM chat_tareas_conversaciones WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, conversacionID).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return id > 0, nil
}

func chatTareaExists(dbEmp *sql.DB, empresaID, tareaID int64) (bool, error) {
	if dbEmp == nil || empresaID <= 0 || tareaID <= 0 {
		return false, nil
	}
	var id int64
	err := dbpkg.QueryRowCompat(dbEmp, `SELECT id FROM chat_tareas WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, tareaID).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return id > 0, nil
}

func ensureChatConversacionExists(dbEmp *sql.DB, empresaID, conversacionID int64) error {
	exists, err := chatConversacionExists(dbEmp, empresaID, conversacionID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("conversacion_id no corresponde a una conversacion valida de esta empresa")
	}
	return nil
}

func ensureChatTareaExists(dbEmp *sql.DB, empresaID, tareaID int64) error {
	exists, err := chatTareaExists(dbEmp, empresaID, tareaID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("tarea_id no corresponde a una tarea valida de esta empresa")
	}
	return nil
}

func normalizeChatParticipanteForEmpresa(dbEmp *sql.DB, empresaID int64, participant dbpkg.ChatParticipante) (dbpkg.ChatParticipante, error) {
	participant.EmpresaID = empresaID
	participant.ParticipanteTipo = strings.ToLower(strings.TrimSpace(participant.ParticipanteTipo))
	participant.Email = normalizeChatActorEmail(participant.Email)
	participant.Nombre = strings.TrimSpace(participant.Nombre)
	if participant.ParticipanteTipo == "" {
		if participant.ParticipanteRefID > 0 {
			participant.ParticipanteTipo = "usuario"
		} else {
			participant.ParticipanteTipo = "admin"
		}
	}
	if participant.ParticipanteTipo != "usuario" {
		return participant, nil
	}

	var scopedUser *dbpkg.EmpresaUsuario
	if participant.ParticipanteRefID > 0 {
		userByID, err := dbpkg.GetEmpresaUsuarioByID(dbEmp, empresaID, participant.ParticipanteRefID)
		if err != nil {
			if err == sql.ErrNoRows {
				return participant, fmt.Errorf("el participante seleccionado no pertenece a esta empresa")
			}
			return participant, err
		}
		scopedUser = userByID
	}
	if participant.Email != "" {
		userByEmail, err := dbpkg.GetEmpresaUsuarioByEmailScoped(dbEmp, participant.Email, empresaID)
		if err != nil {
			if err == sql.ErrNoRows {
				return participant, fmt.Errorf("el participante seleccionado no pertenece a esta empresa")
			}
			return participant, err
		}
		if scopedUser != nil && userByEmail.ID != scopedUser.ID {
			return participant, fmt.Errorf("email y participante_ref_id no corresponden al mismo usuario de la empresa")
		}
		scopedUser = userByEmail
	}
	if scopedUser == nil {
		return participant, fmt.Errorf("el participante seleccionado no pertenece a esta empresa")
	}

	participant.ParticipanteRefID = scopedUser.ID
	participant.Email = normalizeChatActorEmail(scopedUser.Email)
	participant.Nombre = safeAuthorName(scopedUser.Nombre, participant.Email)
	return participant, nil
}

func removeChatUploadFile(path string) {
	if strings.TrimSpace(path) == "" {
		return
	}
	_ = os.Remove(path)
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

type chatActor struct {
	Tipo           string
	RefID          int64
	Nombre         string
	Email          string
	UsuarioCreador string
}

func normalizeChatActorEmail(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	if v == "" {
		return ""
	}
	return v
}

func getEmpresaOwnerEmail(dbEmp *sql.DB, empresaID int64) string {
	if dbEmp == nil || empresaID <= 0 {
		return ""
	}

	var owner string
	if err := dbEmp.QueryRow(`SELECT COALESCE(usuario_creador, '') FROM empresas WHERE id = ? LIMIT 1`, empresaID).Scan(&owner); err != nil {
		return ""
	}
	return normalizeChatActorEmail(owner)
}

func resolveChatActor(dbEmp *sql.DB, r *http.Request, empresaID int64) chatActor {
	email := normalizeChatActorEmail(adminEmailFromRequest(r))
	if email == "" || email == "sistema" {
		return chatActor{
			Tipo:           "sistema",
			RefID:          0,
			Nombre:         "Sistema",
			Email:          "sistema",
			UsuarioCreador: "sistema",
		}
	}

	actor := chatActor{
		Tipo:           "admin",
		RefID:          0,
		Nombre:         safeAuthorName("", email),
		Email:          email,
		UsuarioCreador: email,
	}

	if dbEmp != nil && empresaID > 0 {
		if user, err := dbpkg.GetEmpresaUsuarioByEmailScoped(dbEmp, email, empresaID); err == nil && user != nil {
			actor.Tipo = "usuario"
			actor.RefID = user.ID
			actor.Email = normalizeChatActorEmail(user.Email)
			actor.Nombre = safeAuthorName(user.Nombre, actor.Email)
			actor.UsuarioCreador = actor.Email
		}
	}

	return actor
}

func ensureChatActorParticipant(dbEmp *sql.DB, empresaID, conversacionID int64, actor chatActor) {
	if dbEmp == nil || empresaID <= 0 || conversacionID <= 0 {
		return
	}
	email := normalizeChatActorEmail(actor.Email)
	if email == "" || email == "sistema" {
		return
	}
	_, _ = dbpkg.CreateChatParticipante(dbEmp, dbpkg.ChatParticipante{
		EmpresaID:         empresaID,
		ConversacionID:    conversacionID,
		ParticipanteTipo:  actor.Tipo,
		ParticipanteRefID: actor.RefID,
		Nombre:            safeAuthorName(actor.Nombre, email),
		Email:             email,
		UsuarioCreador:    actor.UsuarioCreador,
		Estado:            "activo",
	})
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
		".doc":  true,
		".docx": true,
		".xls":  true,
		".xlsx": true,
		".ppt":  true,
		".pptx": true,
		".rtf":  true,
		".odt":  true,
		".ods":  true,
		".odp":  true,
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

			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "chat_usuarios" || action == "canal_general" {
				actor := resolveChatActor(dbEmp, r, empresaID)
				item, created, err := dbpkg.EnsureChatUsuariosGeneralConversacion(dbEmp, empresaID, actor.UsuarioCreador)
				if err != nil {
					http.Error(w, "No se pudo preparar el canal general", http.StatusInternalServerError)
					return
				}
				ensureChatActorParticipant(dbEmp, empresaID, item.ID, actor)
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":           true,
					"created":      created,
					"conversacion": item,
				})
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

			normalizedParticipants := make([]dbpkg.ChatParticipante, 0, len(payload.Participantes))
			for _, participant := range payload.Participantes {
				normalizedParticipant, err := normalizeChatParticipanteForEmpresa(dbEmp, payload.EmpresaID, participant)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				normalizedParticipants = append(normalizedParticipants, normalizedParticipant)
			}

			actor := resolveChatActor(dbEmp, r, payload.EmpresaID)
			payload.UsuarioCreador = actor.UsuarioCreador
			if strings.TrimSpace(payload.EstadoConversacion) == "" {
				payload.EstadoConversacion = "abierta"
			}

			newID, err := dbpkg.CreateChatConversacion(dbEmp, payload.ChatConversacion)
			if err != nil {
				http.Error(w, "No se pudo crear la conversacion", http.StatusBadRequest)
				return
			}

			ensureChatActorParticipant(dbEmp, payload.EmpresaID, newID, actor)

			empresaOwnerEmail := getEmpresaOwnerEmail(dbEmp, payload.EmpresaID)
			if empresaOwnerEmail != "" && !strings.EqualFold(empresaOwnerEmail, actor.Email) {
				_, _ = dbpkg.CreateChatParticipante(dbEmp, dbpkg.ChatParticipante{
					EmpresaID:         payload.EmpresaID,
					ConversacionID:    newID,
					ParticipanteTipo:  "admin",
					ParticipanteRefID: 0,
					Nombre:            empresaOwnerEmail,
					Email:             empresaOwnerEmail,
					UsuarioCreador:    actor.UsuarioCreador,
					Estado:            "activo",
				})
			}

			for _, p := range normalizedParticipants {
				p.EmpresaID = payload.EmpresaID
				p.ConversacionID = newID
				p.UsuarioCreador = actor.UsuarioCreador
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
			if err := ensureChatConversacionExists(dbEmp, payload.EmpresaID, payload.ConversacionID); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.Email) == "" && payload.ParticipanteRefID <= 0 {
				http.Error(w, "email o participante_ref_id es obligatorio", http.StatusBadRequest)
				return
			}
			normalizedParticipant, err := normalizeChatParticipanteForEmpresa(dbEmp, payload.EmpresaID, payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			payload = normalizedParticipant
			payload.UsuarioCreador = resolveChatActor(dbEmp, r, payload.EmpresaID).UsuarioCreador
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
			if err := ensureChatConversacionExists(dbEmp, payload.EmpresaID, payload.ConversacionID); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.Contenido) == "" {
				http.Error(w, "contenido es obligatorio", http.StatusBadRequest)
				return
			}
			actor := resolveChatActor(dbEmp, r, payload.EmpresaID)
			payload.UsuarioCreador = actor.UsuarioCreador
			payload.AutorTipo = actor.Tipo
			payload.AutorRefID = actor.RefID
			payload.AutorEmail = actor.Email
			payload.AutorNombre = safeAuthorName("", actor.Nombre)
			if strings.TrimSpace(payload.AutorNombre) == "" {
				payload.AutorNombre = safeAuthorName(payload.AutorNombre, payload.AutorEmail)
			}
			if strings.TrimSpace(payload.TipoMensaje) == "" {
				payload.TipoMensaje = "texto"
			}
			ensureChatActorParticipant(dbEmp, payload.EmpresaID, payload.ConversacionID, actor)
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
		w.Header().Set("X-Empresa-ID", strconv.FormatInt(empresaID, 10))
		conversacionID, err := parseInt64Form(r, "conversacion_id")
		if err != nil || conversacionID <= 0 {
			http.Error(w, "conversacion_id required", http.StatusBadRequest)
			return
		}
		if err := ensureChatConversacionExists(dbEmp, empresaID, conversacionID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
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

		fileName, absPath, size, err := saveEmpresaPrivateUpload(empresaID, "chat_tareas", ext, file, 20<<20)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		actor := resolveChatActor(dbEmp, r, empresaID)
		autorNombre := safeAuthorName("", actor.Nombre)
		if strings.TrimSpace(autorNombre) == "" {
			autorNombre = safeAuthorName(autorNombre, actor.Email)
		}
		contenido := strings.TrimSpace(r.FormValue("contenido"))
		if contenido == "" {
			contenido = "Adjunto: " + strings.TrimSpace(header.Filename)
		}

		ensureChatActorParticipant(dbEmp, empresaID, conversacionID, actor)

		msgID, err := dbpkg.CreateChatMensaje(dbEmp, dbpkg.ChatMensaje{
			EmpresaID:      empresaID,
			ConversacionID: conversacionID,
			AutorTipo:      actor.Tipo,
			AutorRefID:     actor.RefID,
			AutorNombre:    autorNombre,
			AutorEmail:     actor.Email,
			Contenido:      contenido,
			TipoMensaje:    "texto",
			UsuarioCreador: actor.UsuarioCreador,
			Estado:         "activo",
		})
		if err != nil {
			removeChatUploadFile(absPath)
			http.Error(w, "failed to create message", http.StatusInternalServerError)
			return
		}

		duracion, _ := parseFloat64FormOptional(r, "duracion_segundos")
		tipoAdjunto := inferAttachmentType(r.FormValue("tipo_archivo"), contentType, ext)
		fileURL := empresaPrivateDownloadURL("/api/empresa/chat_tareas/archivo", empresaID, fileName)

		adjID, err := dbpkg.CreateChatAdjunto(dbEmp, dbpkg.ChatAdjunto{
			EmpresaID:        empresaID,
			MensajeID:        msgID,
			TipoArchivo:      tipoAdjunto,
			NombreArchivo:    strings.TrimSpace(header.Filename),
			MimeType:         contentType,
			FileURL:          fileURL,
			TamanoBytes:      size,
			DuracionSegundos: duracion,
			UsuarioCreador:   actor.UsuarioCreador,
			Estado:           "activo",
		})
		if err != nil {
			_ = dbpkg.DeleteChatMensaje(dbEmp, empresaID, conversacionID, msgID)
			removeChatUploadFile(absPath)
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

// EmpresaChatTareasTareaNotaVozUploadHandler sube una nota de voz y la asocia a una tarea.
func EmpresaChatTareasTareaNotaVozUploadHandler(dbEmp *sql.DB) http.HandlerFunc {
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
		w.Header().Set("X-Empresa-ID", strconv.FormatInt(empresaID, 10))

		tareaID, err := parseInt64Form(r, "tarea_id")
		if err != nil || tareaID <= 0 {
			http.Error(w, "tarea_id required", http.StatusBadRequest)
			return
		}
		if err := ensureChatTareaExists(dbEmp, empresaID, tareaID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("archivo")
		if err != nil {
			file, header, err = r.FormFile("audio")
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
			ext = ".webm"
		}
		if !isAllowedAttachmentExt(ext) {
			http.Error(w, "attachment extension not allowed", http.StatusBadRequest)
			return
		}
		if inferAttachmentType("", contentType, ext) != "audio" {
			http.Error(w, "only audio files are allowed for nota_voz", http.StatusBadRequest)
			return
		}

		fileName, absPath, size, err := saveEmpresaPrivateUpload(empresaID, "chat_tareas", ext, file, 20<<20)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		duracion, _ := parseFloat64FormOptional(r, "duracion_segundos")
		fileURL := empresaPrivateDownloadURL("/api/empresa/chat_tareas/archivo", empresaID, fileName)
		if err := dbpkg.SetChatTareaNotaVoz(dbEmp, empresaID, tareaID, fileURL, contentType, size, duracion); err != nil {
			removeChatUploadFile(absPath)
			http.Error(w, "failed to update tarea nota_voz", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusCreated, map[string]interface{}{
			"ok":                         true,
			"tarea_id":                   tareaID,
			"nota_voz_url":               fileURL,
			"nota_voz_mime_type":         contentType,
			"nota_voz_tamano_bytes":      size,
			"nota_voz_duracion_segundos": duracion,
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
			if payload.ConversacionID > 0 {
				if err := ensureChatConversacionExists(dbEmp, payload.EmpresaID, payload.ConversacionID); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			}
			if strings.TrimSpace(payload.Titulo) == "" {
				http.Error(w, "titulo es obligatorio", http.StatusBadRequest)
				return
			}

			actor := resolveChatActor(dbEmp, r, payload.EmpresaID)
			payload.UsuarioCreador = actor.UsuarioCreador
			payload.CreadoPorTipo = actor.Tipo
			payload.CreadoPorEmail = actor.Email
			if strings.TrimSpace(payload.EstadoTarea) == "" {
				payload.EstadoTarea = "pendiente"
			}
			if payload.ConversacionID > 0 {
				ensureChatActorParticipant(dbEmp, payload.EmpresaID, payload.ConversacionID, actor)
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
			if payload.ConversacionID > 0 {
				if err := ensureChatConversacionExists(dbEmp, payload.EmpresaID, payload.ConversacionID); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
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

// EmpresaChatTareasCitasHandler gestiona agenda de citas compartidas por empresa.
func EmpresaChatTareasCitasHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			includeInactive := queryBool(r, "include_inactive")
			desde := strings.TrimSpace(r.URL.Query().Get("desde"))
			hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
			estadoCita := strings.TrimSpace(r.URL.Query().Get("estado_cita"))
			q := strings.TrimSpace(r.URL.Query().Get("q"))

			rows, err := dbpkg.GetChatCitas(dbEmp, empresaID, desde, hasta, includeInactive, estadoCita, q)
			if err != nil {
				http.Error(w, "No se pudieron listar las citas", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, rows)
			return

		case http.MethodPost:
			var payload dbpkg.ChatCita
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 {
				http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
				return
			}
			if payload.ConversacionID > 0 {
				if err := ensureChatConversacionExists(dbEmp, payload.EmpresaID, payload.ConversacionID); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			}
			if strings.TrimSpace(payload.Titulo) == "" {
				http.Error(w, "titulo es obligatorio", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.FechaInicio) == "" {
				http.Error(w, "fecha_inicio es obligatoria", http.StatusBadRequest)
				return
			}

			actor := resolveChatActor(dbEmp, r, payload.EmpresaID)
			payload.UsuarioCreador = actor.UsuarioCreador
			payload.CreadoPorTipo = actor.Tipo
			payload.CreadoPorRefID = actor.RefID
			payload.CreadoPorEmail = actor.Email
			payload.CreadoPorNombre = safeAuthorName(payload.CreadoPorNombre, actor.Nombre)
			if strings.TrimSpace(payload.EstadoCita) == "" {
				payload.EstadoCita = "programada"
			}
			if strings.TrimSpace(payload.Visibilidad) == "" {
				payload.Visibilidad = "empresa"
			}
			if payload.NotificarMinutosAntes <= 0 {
				payload.NotificarMinutosAntes = 30
			}
			if strings.TrimSpace(payload.FechaFin) == "" {
				payload.FechaFin = payload.FechaInicio
			}

			id, err := dbpkg.CreateChatCita(dbEmp, payload)
			if err != nil {
				log.Printf("EmpresaChatTareasCitasHandler CreateChatCita error empresa_id=%d conversacion_id=%d: %v", payload.EmpresaID, payload.ConversacionID, err)
				http.Error(w, "No se pudo crear la cita", http.StatusBadRequest)
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
				if err := dbpkg.SetChatCitaEstado(dbEmp, empresaID, id, estado); err != nil {
					http.Error(w, "No se pudo actualizar estado de la cita", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": estado})
				return
			}

			if action == "cancelar" || action == "completar" || action == "reprogramar" {
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

				estadoCita := "programada"
				switch action {
				case "cancelar":
					estadoCita = "cancelada"
				case "completar":
					estadoCita = "completada"
				case "reprogramar":
					estadoCita = "programada"
				}

				if err := dbpkg.SetChatCitaWorkflowEstado(dbEmp, empresaID, id, estadoCita); err != nil {
					http.Error(w, "No se pudo actualizar estado operativo de la cita", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado_cita": estadoCita})
				return
			}

			if action == "marcar_recordatorio" || action == "limpiar_recordatorio" {
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

				sent := action == "marcar_recordatorio"
				if strings.TrimSpace(r.URL.Query().Get("sent")) != "" {
					sent = queryBool(r, "sent")
				}

				if err := dbpkg.SetChatCitaReminderSent(dbEmp, empresaID, id, sent); err != nil {
					http.Error(w, "No se pudo actualizar estado de recordatorio", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "recordatorio_enviado": sent})
				return
			}

			var payload dbpkg.ChatCita
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
			if strings.TrimSpace(payload.FechaInicio) == "" {
				http.Error(w, "fecha_inicio es obligatoria", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.FechaFin) == "" {
				payload.FechaFin = payload.FechaInicio
			}
			if payload.NotificarMinutosAntes <= 0 {
				payload.NotificarMinutosAntes = 30
			}
			if strings.TrimSpace(payload.Visibilidad) == "" {
				payload.Visibilidad = "empresa"
			}

			if err := dbpkg.UpdateChatCita(dbEmp, payload); err != nil {
				http.Error(w, "No se pudo actualizar la cita", http.StatusBadRequest)
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
			if err := dbpkg.DeleteChatCita(dbEmp, empresaID, id); err != nil {
				http.Error(w, "No se pudo eliminar la cita", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

// EmpresaChatTareasPapeleraHandler permite listar y restaurar elementos eliminados (papelera temporal).
// GET  /api/empresa/chat_tareas/papelera?empresa_id=1&tipo=conversacion|mensaje|tarea|cita&limit&offset
// POST /api/empresa/chat_tareas/papelera?action=restaurar  { empresa_id, tipo, id }
func EmpresaChatTareasPapeleraHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			tipo := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("tipo")))
			limit, _ := parseIntQueryOptional(r, "limit")
			offset, _ := parseIntQueryOptional(r, "offset")
			rows, err := dbpkg.ListChatTrash(dbEmp, empresaID, tipo, limit, offset)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "rows": rows})
			return
		}

		if r.Method == http.MethodPost {
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action != "restaurar" {
				http.Error(w, "action invalida", http.StatusBadRequest)
				return
			}
			var payload struct {
				EmpresaID int64  `json:"empresa_id"`
				Tipo      string `json:"tipo"`
				ID        int64  `json:"id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 || payload.ID <= 0 {
				http.Error(w, "empresa_id e id son obligatorios", http.StatusBadRequest)
				return
			}
			if err := dbpkg.RestoreChatEntity(dbEmp, payload.EmpresaID, payload.Tipo, payload.ID); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}
