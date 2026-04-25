package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type publicPrivMsgLimiter struct {
	mu      sync.Mutex
	records map[string]*publicPortalChatUsage
}

var privMsgLimiter = &publicPrivMsgLimiter{records: map[string]*publicPortalChatUsage{}}

func (l *publicPrivMsgLimiter) allow(key string, max int, window time.Duration) (bool, int) {
	ok, _, retry := portalChatLimiter.allow(key, max, window)
	return ok, retry
}

func publicClientIP(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); v != "" {
		parts := strings.Split(v, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if v := strings.TrimSpace(r.Header.Get("X-Real-IP")); v != "" {
		return v
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func randomPublicMessageID() string {
	var b [12]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func normalizePublicSource(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "red_social", "red_social_comercial", "social":
		return "red_social"
	case "venta_publica", "venta", "store":
		return "venta_publica"
	default:
		return "portal_publico"
	}
}

// PublicMensajesPrivadosHandler permite a un visitante enviar un mensaje privado a una empresa.
// El mensaje se materializa como una conversación en el módulo Chat y tareas de esa empresa.
func PublicMensajesPrivadosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		if dbEmp == nil {
			http.Error(w, "db empresas no disponible", http.StatusInternalServerError)
			return
		}

		ip := publicClientIP(r)
		key := "privmsg:" + ip
		if ok, retry := privMsgLimiter.allow(key, 4, 5*time.Minute); !ok {
			w.Header().Set("Retry-After", fmt.Sprintf("%d", retry))
			writeJSON(w, http.StatusTooManyRequests, map[string]interface{}{
				"ok":                 false,
				"code":               "rate_limited",
				"error":              "Has enviado demasiados mensajes. Espera unos minutos e inténtalo de nuevo.",
				"retry_after_seconds": retry,
			})
			return
		}

		var body struct {
			EmpresaID    int64  `json:"empresa_id"`
			EmpresaSlug  string `json:"empresa_slug"`
			Origen       string `json:"origen"` // red_social | venta_publica
			OrigenRef    string `json:"origen_ref"`
			Nombre       string `json:"nombre"`
			Contacto     string `json:"contacto"` // whatsapp o email
			Mensaje      string `json:"mensaje"`
			ConsentTerms bool   `json:"consent_terms"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "JSON invalido", http.StatusBadRequest)
			return
		}

		empresaID := body.EmpresaID
		if empresaID <= 0 && strings.TrimSpace(body.EmpresaSlug) != "" {
			if id, err := dbpkg.ResolveEmpresaIDByVentaPublicaSlug(dbEmp, body.EmpresaSlug); err == nil {
				empresaID = id
			}
		}
		if empresaID <= 0 {
			http.Error(w, "empresa_id o empresa_slug es obligatorio", http.StatusBadRequest)
			return
		}

		origen := normalizePublicSource(body.Origen)
		nombre := strings.TrimSpace(body.Nombre)
		if nombre == "" {
			nombre = "Visitante"
		}
		contacto := strings.TrimSpace(body.Contacto)
		msg := strings.TrimSpace(body.Mensaje)
		if msg == "" {
			http.Error(w, "mensaje es obligatorio", http.StatusBadRequest)
			return
		}
		if len(msg) > 2000 {
			msg = msg[:2000]
		}
		if len(contacto) > 180 {
			contacto = contacto[:180]
		}
		if len(nombre) > 140 {
			nombre = nombre[:140]
		}

		_ = dbpkg.EnsureEmpresaChatTareasSchema(dbEmp)

		ref := strings.TrimSpace(body.OrigenRef)
		if len(ref) > 220 {
			ref = ref[:220]
		}
		conversationTitle := "Mensaje privado (Portal público)"
		if origen == "red_social" {
			conversationTitle = "Mensaje privado (Red social comercial)"
		}
		if origen == "venta_publica" {
			conversationTitle = "Mensaje privado (Venta pública)"
		}
		convDesc := "Origen: " + origen
		if ref != "" {
			convDesc += " — Ref: " + ref
		}
		if contacto != "" {
			convDesc += "\nContacto: " + contacto
		}
		convDesc += "\nID público: " + randomPublicMessageID()

		convID, err := dbpkg.CreateChatConversacion(dbEmp, dbpkg.ChatConversacion{
			EmpresaID:          empresaID,
			Titulo:             conversationTitle,
			Descripcion:        convDesc,
			Prioridad:          "media",
			EstadoConversacion: "abierta",
			UsuarioCreador:     "publico",
			Estado:             "activo",
			Observaciones:      "",
		})
		if err != nil {
			http.Error(w, "no se pudo crear conversación", http.StatusInternalServerError)
			return
		}

		// Registrar participante público (para trazabilidad).
		_, _ = dbpkg.CreateChatParticipante(dbEmp, dbpkg.ChatParticipante{
			EmpresaID:         empresaID,
			ConversacionID:    convID,
			ParticipanteTipo:  "publico",
			ParticipanteRefID: 0,
			Nombre:            nombre,
			Email:             contacto,
			UsuarioCreador:    "publico",
			Estado:            "activo",
		})

		_, err = dbpkg.CreateChatMensaje(dbEmp, dbpkg.ChatMensaje{
			EmpresaID:      empresaID,
			ConversacionID: convID,
			AutorTipo:      "publico",
			AutorRefID:     0,
			AutorNombre:    nombre,
			AutorEmail:     contacto,
			Contenido:      msg,
			TipoMensaje:    "texto",
			UsuarioCreador: "publico",
			Estado:         "activo",
		})
		if err != nil {
			http.Error(w, "no se pudo guardar mensaje", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":             true,
			"empresa_id":      empresaID,
			"conversacion_id": convID,
			"message":         "Mensaje enviado. La empresa te responderá por el mismo canal si dejaste tu contacto.",
		})
	}
}

