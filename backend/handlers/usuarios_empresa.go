package handlers

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

const (
	empresaUsuarioMaxIntentosFallidos  = 5
	empresaUsuarioVentanaIntentos      = 15 * time.Minute
	empresaUsuarioBloqueoDuracion      = 15 * time.Minute
	empresaUsuarioRecuperacionTTL      = 30 * time.Minute
	empresaUsuarioLoginPublicSubdomain = "usuarios"

	superCorreoNotificacionTipoPruebaGmail = "prueba_gmail_super"
	superGmailTestRecipient                = "powerfulcontrolsystem@gmail.com"

	empresaUsuarioPasswordMinLengthDefault     = 8
	empresaUsuarioPasswordRequireUpperDefault  = true
	empresaUsuarioPasswordRequireLowerDefault  = true
	empresaUsuarioPasswordRequireDigitDefault  = true
	empresaUsuarioPasswordRequireSymbolDefault = false
	empresaUsuarioPasswordRotationDaysDefault  = 0
)

type empresaUsuarioPasswordPolicy struct {
	MinLength     int
	RequireUpper  bool
	RequireLower  bool
	RequireDigit  bool
	RequireSymbol bool
	RotationDays  int
}

var (
	empresaUsuarioPasswordPolicyMu       sync.Mutex
	empresaUsuarioPasswordPolicyCached   empresaUsuarioPasswordPolicy
	empresaUsuarioPasswordPolicyLoadedAt time.Time
)

const empresaUsuarioPasswordPolicyCacheTTL = 30 * time.Second

func defaultEmpresaUsuarioPasswordPolicy() empresaUsuarioPasswordPolicy {
	return empresaUsuarioPasswordPolicy{
		MinLength:     empresaUsuarioPasswordMinLengthDefault,
		RequireUpper:  empresaUsuarioPasswordRequireUpperDefault,
		RequireLower:  empresaUsuarioPasswordRequireLowerDefault,
		RequireDigit:  empresaUsuarioPasswordRequireDigitDefault,
		RequireSymbol: empresaUsuarioPasswordRequireSymbolDefault,
		RotationDays:  empresaUsuarioPasswordRotationDaysDefault,
	}
}

func empresaUsuarioPasswordPolicyToMap(policy empresaUsuarioPasswordPolicy) map[string]interface{} {
	return map[string]interface{}{
		"min_length":     policy.MinLength,
		"require_upper":  policy.RequireUpper,
		"require_lower":  policy.RequireLower,
		"require_digit":  policy.RequireDigit,
		"require_symbol": policy.RequireSymbol,
		"rotation_days":  policy.RotationDays,
	}
}

func IsEmpresaUsuarioLoginSubdomainRequest(r *http.Request) bool {
	host := strings.ToLower(splitHostPortSafe(resolveOAuthHost(r)))
	return host == empresaUsuarioLoginPublicSubdomain+".powerfulcontrolsystem.com"
}

func resolveEmpresaUsuarioLoginURLFromBase(baseURL, empresaSlug, dominioPublico string, empresaID int64) string {
	_ = empresaSlug
	_ = dominioPublico
	_ = empresaID
	trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if trimmed == "" {
		trimmed = "https://powerfulcontrolsystem.com"
	}

	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Host == "" {
		return trimmed + "/login_usuario.html"
	}

	host := strings.ToLower(splitHostPortSafe(parsed.Host))
	if host == "www.powerfulcontrolsystem.com" || strings.HasSuffix(host, ".powerfulcontrolsystem.com") {
		parsed.Scheme = "https"
		parsed.Host = "powerfulcontrolsystem.com"
	}
	parsed.Path = "/login_usuario.html"
	parsed.RawPath = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

func resolveEmpresaUsuarioLoginURL(r *http.Request, dbEmp, dbSuper *sql.DB, empresaID int64) string {
	baseURL := resolveBaseURLForConfirmation(r, dbSuper)
	if dbEmp != nil && empresaID > 0 {
		if cfg, err := dbpkg.GetEmpresaVentaPublicaConfig(dbEmp, empresaID); err == nil {
			return resolveEmpresaUsuarioLoginURLFromBase(baseURL, cfg.EmpresaSlug, cfg.EmpresaSlug, empresaID)
		}
	}
	return resolveEmpresaUsuarioLoginURLFromBase(baseURL, "", "", empresaID)
}

var errEmpresaUsuarioEmailAmbiguo = errors.New("empresa user email resolves to multiple companies")

func resolveUniqueEmpresaUsuarioByEmail(dbEmp *sql.DB, email string, empresaID int64) (*dbpkg.EmpresaUsuario, error) {
	if empresaID > 0 {
		return dbpkg.GetEmpresaUsuarioByEmailScoped(dbEmp, email, empresaID)
	}
	items, err := dbpkg.GetEmpresaUsuariosByEmail(dbEmp, email)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, sql.ErrNoRows
	}
	if len(items) > 1 {
		return nil, errEmpresaUsuarioEmailAmbiguo
	}
	return &items[0], nil
}

func resolveEmpresaUsuarioForPasswordLogin(dbEmp *sql.DB, email, password string, empresaID int64) (*dbpkg.EmpresaUsuario, bool, error) {
	if empresaID > 0 {
		item, err := dbpkg.GetEmpresaUsuarioByEmailScoped(dbEmp, email, empresaID)
		return item, false, err
	}

	items, err := dbpkg.GetEmpresaUsuariosByEmail(dbEmp, email)
	if err != nil {
		return nil, false, err
	}
	if len(items) == 0 {
		return nil, false, sql.ErrNoRows
	}
	if len(items) == 1 {
		return &items[0], false, nil
	}

	password = strings.TrimSpace(password)
	if password == "" {
		return nil, false, errEmpresaUsuarioEmailAmbiguo
	}
	matches := make([]*dbpkg.EmpresaUsuario, 0, 1)
	for i := range items {
		item := &items[i]
		if item.PasswordSet == 1 && strings.TrimSpace(item.PasswordHash) != "" && strings.TrimSpace(item.PasswordSalt) != "" && verifyEmpresaUsuarioPassword(password, item) {
			matches = append(matches, item)
		}
	}
	if len(matches) == 1 {
		return matches[0], true, nil
	}
	if len(matches) > 1 {
		return nil, false, errEmpresaUsuarioEmailAmbiguo
	}
	return nil, false, sql.ErrNoRows
}

func resolveEmpresaUsuarioForPasswordReset(dbEmp *sql.DB, email, token string, empresaID int64) (*dbpkg.EmpresaUsuario, error) {
	if empresaID > 0 {
		return dbpkg.GetEmpresaUsuarioByEmailScoped(dbEmp, email, empresaID)
	}
	items, err := dbpkg.GetEmpresaUsuariosByEmail(dbEmp, email)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, sql.ErrNoRows
	}
	token = strings.TrimSpace(token)
	var matched *dbpkg.EmpresaUsuario
	for i := range items {
		storedToken := strings.TrimSpace(items[i].PasswordResetToken)
		if storedToken != "" && token != "" && subtle.ConstantTimeCompare([]byte(token), []byte(storedToken)) == 1 {
			if matched != nil {
				return nil, errEmpresaUsuarioEmailAmbiguo
			}
			matched = &items[i]
		}
	}
	if matched == nil {
		return nil, sql.ErrNoRows
	}
	return matched, nil
}

func empresaUsuarioContractAccepted(item *dbpkg.EmpresaUsuario, contract *dbpkg.SuperContractVersion) bool {
	if item == nil || contract == nil {
		return false
	}
	return item.AceptaContrato == 1 && item.ContratoVersionAceptada >= contract.Version
}

func writeEmpresaUsuarioContractRequirement(w http.ResponseWriter, item *dbpkg.EmpresaUsuario, contract *dbpkg.SuperContractVersion, message string) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"ok":                           false,
		"contract_acceptance_required": true,
		"message":                      message,
	}
	if item != nil {
		response["empresa_id"] = item.EmpresaID
		response["email"] = item.Email
	}
	if contract != nil {
		response["contract"] = contract
	}
	_ = json.NewEncoder(w).Encode(response)
}

func empresaUsuarioPublicPayload(item *dbpkg.EmpresaUsuario) map[string]interface{} {
	if item == nil {
		return nil
	}
	return map[string]interface{}{
		"id":                          item.ID,
		"empresa_id":                  item.EmpresaID,
		"email":                       item.Email,
		"nombre":                      item.Nombre,
		"documento_identidad":         item.DocumentoIdentidad,
		"rol_usuario_id":              item.RolUsuarioID,
		"rol_nombre":                  item.RolNombre,
		"foto_url":                    item.FotoURL,
		"control_aseo_estaciones":     item.ControlAseoEstaciones,
		"email_confirmado":            item.EmailConfirmado,
		"email_confirmado_en":         item.EmailConfirmadoEn,
		"estado":                      item.Estado,
		"fecha_creacion":              item.FechaCreacion,
		"fecha_actualizacion":         item.FechaActualizacion,
		"observaciones":               item.Observaciones,
		"puede_reenviar_confirmacion": item.EmailConfirmado != 1,
	}
}

func writeEmpresaUsuarioDuplicateResponse(w http.ResponseWriter, empresaID int64, email string, existing *dbpkg.EmpresaUsuario, message string) {
	if strings.TrimSpace(message) == "" {
		message = "Ya existe un usuario de esta empresa con ese correo. Revisa la lista y, si esta pendiente de confirmacion, usa Reenviar confirmacion."
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusConflict)
	response := map[string]interface{}{
		"ok":                      false,
		"code":                    "empresa_usuario_email_duplicado",
		"empresa_id":              empresaID,
		"email":                   strings.TrimSpace(email),
		"error":                   message,
		"message":                 message,
		"can_resend_confirmation": existing != nil && existing.EmailConfirmado != 1,
		"email_confirmado":        0,
		"usuario_existente":       empresaUsuarioPublicPayload(existing),
	}
	if existing != nil {
		response["email_confirmado"] = existing.EmailConfirmado
		response["estado"] = existing.Estado
	}
	_ = json.NewEncoder(w).Encode(response)
}

func empresaUsuarioEstadoBloqueaPrimerIngreso(item *dbpkg.EmpresaUsuario) bool {
	if item == nil {
		return true
	}
	if !strings.EqualFold(strings.TrimSpace(item.Estado), "inactivo") {
		return false
	}
	if item.EmailConfirmado != 1 {
		return false
	}
	return item.PasswordSet == 1 || strings.TrimSpace(item.PasswordHash) != "" || strings.TrimSpace(item.PasswordSalt) != ""
}

func ensureEmpresaUsuarioCurrentContractAccepted(dbEmp, dbSuper *sql.DB, item *dbpkg.EmpresaUsuario, acceptRequested bool) (*dbpkg.SuperContractVersion, bool, error) {
	contract, err := dbpkg.GetCurrentSuperContract(dbSuper)
	if err != nil {
		return nil, false, err
	}
	if empresaUsuarioContractAccepted(item, contract) {
		return contract, true, nil
	}
	if !acceptRequested {
		return contract, false, nil
	}
	if err := dbpkg.SetEmpresaUsuarioContratoAceptado(dbEmp, item.EmpresaID, item.ID, contract.Version); err != nil {
		return nil, false, err
	}
	item.AceptaContrato = 1
	item.ContratoVersionAceptada = contract.Version
	return contract, true, nil
}

// EmpresaRolesDeUsuarioHandler devuelve los roles disponibles para la empresa seleccionada.
func EmpresaRolesDeUsuarioHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if _, err := parseEmpresaIDQuery(r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		includeInactive := r.URL.Query().Get("include_inactive") == "1"
		roles, err := dbpkg.GetRolesDeUsuarioCatalogoGlobal(dbSuper, includeInactive)
		if err != nil {
			http.Error(w, "failed to query roles_de_usuario: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(roles)
	}
}

// EmpresaUsuariosHandler maneja CRUD de usuarios por empresa con confirmación de correo.
func EmpresaUsuariosHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			includeInactive := r.URL.Query().Get("include_inactive") == "1"
			items, err := dbpkg.GetEmpresaUsuarios(dbEmp, empresaID, includeInactive)
			if err != nil {
				http.Error(w, "failed to query users: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(items)
			return

		case http.MethodPost:
			if strings.TrimSpace(r.URL.Query().Get("action")) == "foto" {
				empresaID, err := parseEmpresaIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				userID, photoURL, err := handleEmpresaUsuarioFotoUpload(r, dbEmp, dbSuper, empresaID)
				if err != nil {
					log.Printf("[usuarios_empresa] upload foto empresa_id=%d error: %v", empresaID, err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"ok":       true,
					"id":       userID,
					"foto_url": photoURL,
				})
				return
			}

			var payload struct {
				EmpresaID          int64  `json:"empresa_id"`
				Email              string `json:"email"`
				Nombre             string `json:"nombre"`
				DocumentoIdentidad string `json:"documento_identidad"`
				RolUsuarioID       int64  `json:"rol_usuario_id"`
				ControlAseo        int    `json:"control_aseo_estaciones"`
				Observaciones      string `json:"observaciones"`
				MensajeInvitacion  string `json:"mensaje_invitacion"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if err := validateEmpresaUsuarioPayload(payload.EmpresaID, payload.Email, payload.Nombre, payload.RolUsuarioID); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			rolNombre, err := resolveRolNombreValidoParaEmpresa(dbEmp, dbSuper, payload.EmpresaID, payload.RolUsuarioID)
			if err != nil {
				http.Error(w, "rol no válido para la empresa: "+err.Error(), http.StatusBadRequest)
				return
			}

			token, expira, err := newEmailConfirmationTokenAndExpiration()
			if err != nil {
				http.Error(w, "failed to generate confirmation token", http.StatusInternalServerError)
				return
			}

			id, err := dbpkg.CreateEmpresaUsuario(
				dbEmp,
				payload.EmpresaID,
				strings.TrimSpace(payload.Email),
				strings.TrimSpace(payload.Nombre),
				strings.TrimSpace(payload.DocumentoIdentidad),
				payload.RolUsuarioID,
				payload.ControlAseo,
				rolNombre,
				strings.TrimSpace(payload.Observaciones),
				adminEmailFromRequest(r),
				token,
				expira,
			)
			if err != nil {
				if strings.Contains(strings.ToLower(err.Error()), "unique") {
					email := strings.TrimSpace(payload.Email)
					existing, lookupErr := dbpkg.GetEmpresaUsuarioByEmailScoped(dbEmp, email, payload.EmpresaID)
					if lookupErr == nil && existing != nil {
						writeEmpresaUsuarioDuplicateResponse(w, payload.EmpresaID, email, existing, "Ya existe un usuario de esta empresa con ese correo. Lo dejamos ubicado en la lista; si esta pendiente de confirmacion, usa el boton Reenviar confirmacion para enviarle una nueva invitacion al correo electronico.")
						return
					}
					writeEmpresaUsuarioDuplicateResponse(w, payload.EmpresaID, email, nil, "Este correo ya existe en el sistema, pero no se encontro como usuario de esta empresa. Vuelve a cargar la lista; si no aparece, pide al super administrador revisar usuarios heredados o usa otro correo.")
					return
				}
				http.Error(w, "failed to create user: "+err.Error(), http.StatusInternalServerError)
				return
			}

			confirmURL, mailErr := sendEmpresaUsuarioConfirmationEmail(r, dbEmp, dbSuper, payload.EmpresaID, strings.TrimSpace(payload.Email), strings.TrimSpace(payload.Nombre), token, strings.TrimSpace(payload.MensajeInvitacion))
			if mailErr != nil {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"id":                          id,
					"email_confirmation_required": true,
					"email_sent":                  false,
					"email_error":                 mailErr.Error(),
					"confirm_url_preview":         confirmURL,
				})
				return
			}

			w.Header().Set("Content-Type", "application/json")
			resp := map[string]interface{}{
				"id":                          id,
				"email_confirmation_required": true,
				"email_sent":                  true,
			}
			json.NewEncoder(w).Encode(resp)
			return

		case http.MethodPut:
			action := strings.TrimSpace(r.URL.Query().Get("action"))
			if action == "activar" {
				empresaID, err := parseEmpresaIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				id, err := parseInt64Query(r, "id")
				if err != nil {
					http.Error(w, "id required", http.StatusBadRequest)
					return
				}
				item, err := dbpkg.GetEmpresaUsuarioByID(dbEmp, empresaID, id)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "user not found", http.StatusNotFound)
						return
					}
					log.Printf("[usuarios_empresa] failed to query user (activar) empresa_id=%d id=%d error=%v", empresaID, id, err)
					http.Error(w, "No se pudo validar el usuario", http.StatusInternalServerError)
					return
				}
				estado := "inactivo"
				if r.URL.Query().Get("activo") == "1" || strings.EqualFold(r.URL.Query().Get("estado"), "activo") {
					estado = "activo"
				}
				if estado == "activo" && item.EmailConfirmado != 1 {
					http.Error(w, "no se puede activar el usuario hasta que confirme su correo", http.StatusConflict)
					return
				}
				if err := dbpkg.SetEmpresaUsuarioEstado(dbEmp, empresaID, id, estado); err != nil {
					http.Error(w, "failed to set estado: "+err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			if action == "reenviar_confirmacion" {
				empresaID, err := parseEmpresaIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				id, err := parseInt64Query(r, "id")
				if err != nil {
					http.Error(w, "id required", http.StatusBadRequest)
					return
				}
				item, err := dbpkg.GetEmpresaUsuarioByID(dbEmp, empresaID, id)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "user not found", http.StatusNotFound)
						return
					}
					log.Printf("[usuarios_empresa] failed to query user (reenviar_confirmacion) empresa_id=%d id=%d error=%v", empresaID, id, err)
					http.Error(w, "No se pudo validar el usuario", http.StatusInternalServerError)
					return
				}
				if item.EmailConfirmado == 1 {
					http.Error(w, "el correo ya está confirmado", http.StatusConflict)
					return
				}

				token, expira, err := newEmailConfirmationTokenAndExpiration()
				if err != nil {
					http.Error(w, "failed to generate confirmation token", http.StatusInternalServerError)
					return
				}
				if err := dbpkg.SetEmpresaUsuarioConfirmToken(dbEmp, empresaID, id, token, expira); err != nil {
					http.Error(w, "failed to set confirmation token: "+err.Error(), http.StatusInternalServerError)
					return
				}

				// leer optional mensaje_invitacion desde el body
				var resendPayload struct {
					MensajeInvitacion string `json:"mensaje_invitacion"`
				}
				if err := json.NewDecoder(r.Body).Decode(&resendPayload); err != nil && err != io.EOF {
					// ignore decode errors for empty body, but log others
					log.Printf("[usuarios_empresa] warning decoding resend payload: %v", err)
				}
				confirmURL, mailErr := sendEmpresaUsuarioConfirmationEmail(r, dbEmp, dbSuper, empresaID, item.Email, item.Nombre, token, strings.TrimSpace(resendPayload.MensajeInvitacion))
				w.Header().Set("Content-Type", "application/json")
				resp := map[string]interface{}{
					"resent":     true,
					"email_sent": mailErr == nil,
				}
				if mailErr != nil {
					resp["email_error"] = mailErr.Error()
					resp["confirm_url_preview"] = confirmURL
				}
				json.NewEncoder(w).Encode(resp)
				return
			}

			id, err := parseInt64Query(r, "id")
			if err != nil {
				http.Error(w, "id required", http.StatusBadRequest)
				return
			}
			var payload struct {
				EmpresaID          int64  `json:"empresa_id"`
				Email              string `json:"email"`
				Nombre             string `json:"nombre"`
				DocumentoIdentidad string `json:"documento_identidad"`
				RolUsuarioID       int64  `json:"rol_usuario_id"`
				ControlAseo        int    `json:"control_aseo_estaciones"`
				Observaciones      string `json:"observaciones"`
				MensajeInvitacion  string `json:"mensaje_invitacion"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if err := validateEmpresaUsuarioPayload(payload.EmpresaID, payload.Email, payload.Nombre, payload.RolUsuarioID); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			existing, err := dbpkg.GetEmpresaUsuarioByID(dbEmp, payload.EmpresaID, id)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "user not found", http.StatusNotFound)
					return
				}
				log.Printf("[usuarios_empresa] failed to query user (update) empresa_id=%d id=%d error=%v", payload.EmpresaID, id, err)
				http.Error(w, "No se pudo validar el usuario", http.StatusInternalServerError)
				return
			}

			rolNombre, err := resolveRolNombreValidoParaEmpresa(dbEmp, dbSuper, payload.EmpresaID, payload.RolUsuarioID)
			if err != nil {
				http.Error(w, "rol no válido para la empresa: "+err.Error(), http.StatusBadRequest)
				return
			}

			resetConfirm := !strings.EqualFold(strings.TrimSpace(existing.Email), strings.TrimSpace(payload.Email))
			confirmToken := ""
			confirmExpira := ""
			if resetConfirm {
				confirmToken, confirmExpira, err = newEmailConfirmationTokenAndExpiration()
				if err != nil {
					http.Error(w, "failed to generate confirmation token", http.StatusInternalServerError)
					return
				}
			}

			if err := dbpkg.UpdateEmpresaUsuario(
				dbEmp,
				id,
				payload.EmpresaID,
				strings.TrimSpace(payload.Email),
				strings.TrimSpace(payload.Nombre),
				strings.TrimSpace(payload.DocumentoIdentidad),
				payload.RolUsuarioID,
				payload.ControlAseo,
				rolNombre,
				strings.TrimSpace(payload.Observaciones),
				resetConfirm,
				confirmToken,
				confirmExpira,
			); err != nil {
				if strings.Contains(strings.ToLower(err.Error()), "unique") {
					http.Error(w, "Ese correo ya esta asignado a otro usuario de esta empresa. Revisa la lista de usuarios; si la cuenta esta pendiente, usa Reenviar confirmacion para enviar nuevamente la invitacion.", http.StatusConflict)
					return
				}
				http.Error(w, "failed to update user: "+err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			resp := map[string]interface{}{
				"updated":                     true,
				"email_reconfirmation_needed": resetConfirm,
			}
			if resetConfirm {
				confirmURL, mailErr := sendEmpresaUsuarioConfirmationEmail(r, dbEmp, dbSuper, payload.EmpresaID, strings.TrimSpace(payload.Email), strings.TrimSpace(payload.Nombre), confirmToken, strings.TrimSpace(payload.MensajeInvitacion))
				resp["email_sent"] = mailErr == nil
				if mailErr != nil {
					resp["email_error"] = mailErr.Error()
					resp["confirm_url_preview"] = confirmURL
				}
			}
			json.NewEncoder(w).Encode(resp)
			return

		case http.MethodDelete:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			id, err := parseInt64Query(r, "id")
			if err != nil {
				http.Error(w, "id required", http.StatusBadRequest)
				return
			}
			if err := dbpkg.DeleteEmpresaUsuario(dbEmp, empresaID, id); err != nil {
				http.Error(w, "failed to delete user: "+err.Error(), http.StatusInternalServerError)
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

// EmpresaUsuarioLoginHandler valida credenciales de usuario de empresa y crea sesión de acceso.
func EmpresaUsuarioLoginHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			EmpresaID      int64  `json:"empresa_id"`
			Email          string `json:"email"`
			Password       string `json:"password"`
			RecaptchaToken string `json:"recaptcha_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		email := strings.TrimSpace(payload.Email)
		if email == "" {
			http.Error(w, "email es obligatorio", http.StatusBadRequest)
			return
		}
		if _, err := mail.ParseAddress(email); err != nil {
			http.Error(w, "email inválido", http.StatusBadRequest)
			return
		}
		if payload.EmpresaID <= 0 {
			if qEmpresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && qEmpresaID > 0 {
				payload.EmpresaID = qEmpresaID
			}
		}
		if err := validateRecaptchaToken(dbSuper, r, payload.RecaptchaToken); err != nil {
			writeRecaptchaValidationError(w, err)
			return
		}

		item, passwordAlreadyVerified, err := resolveEmpresaUsuarioForPasswordLogin(dbEmp, email, payload.Password, payload.EmpresaID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "credenciales inválidas", http.StatusUnauthorized)
				return
			}
			if errors.Is(err, errEmpresaUsuarioEmailAmbiguo) {
				http.Error(w, "este correo esta asociado a mas de una empresa; solicita al administrador usar un correo unico por empresa", http.StatusConflict)
				return
			}
			log.Printf("[usuarios_empresa] failed to query user (login) empresa_id=%d email=%s error=%v", payload.EmpresaID, email, err)
			http.Error(w, "No se pudo validar el usuario", http.StatusInternalServerError)
			return
		}

		if item.EmailConfirmado != 1 {
			http.Error(w, "debes confirmar tu correo antes de iniciar sesión", http.StatusForbidden)
			return
		}
		if strings.EqualFold(strings.TrimSpace(item.Estado), "inactivo") && item.EmailConfirmado == 1 {
			http.Error(w, "tu usuario está inactivo", http.StatusForbidden)
			return
		}
		if blocked, lockUntil := dbpkg.IsEmpresaUsuarioLocked(item, time.Now()); blocked {
			http.Error(w, "usuario bloqueado temporalmente por intentos fallidos hasta "+lockUntil, http.StatusTooManyRequests)
			return
		}

		if item.PasswordSet != 1 || strings.TrimSpace(item.PasswordHash) == "" || strings.TrimSpace(item.PasswordSalt) == "" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":                      false,
				"password_setup_required": true,
				"empresa_id":              item.EmpresaID,
				"email":                   item.Email,
				"message":                 "Primer ingreso: abre la invitacion enviada por el administrador para crear tu contrasena.",
			})
			return
		}

		if strings.TrimSpace(payload.Password) == "" {
			http.Error(w, "password es obligatorio", http.StatusBadRequest)
			return
		}
		if !passwordAlreadyVerified && !verifyEmpresaUsuarioPassword(payload.Password, item) {
			_, lockUntil, registerErr := dbpkg.RegisterEmpresaUsuarioLoginFailure(
				dbEmp,
				item.EmpresaID,
				item.ID,
				empresaUsuarioMaxIntentosFallidos,
				empresaUsuarioVentanaIntentos,
				empresaUsuarioBloqueoDuracion,
			)
			if registerErr != nil {
				log.Printf("[usuarios_empresa] failed to register login failure empresa_id=%d id=%d email=%s error=%v", item.EmpresaID, item.ID, item.Email, registerErr)
			}
			if strings.TrimSpace(lockUntil) != "" {
				http.Error(w, "usuario bloqueado temporalmente por intentos fallidos hasta "+lockUntil, http.StatusTooManyRequests)
				return
			}
			http.Error(w, "credenciales inválidas", http.StatusUnauthorized)
			return
		}

		if err := dbpkg.ClearEmpresaUsuarioLoginFailures(dbEmp, item.EmpresaID, item.ID); err != nil {
			log.Printf("[usuarios_empresa] failed to clear login failures empresa_id=%d id=%d email=%s error=%v", item.EmpresaID, item.ID, item.Email, err)
			http.Error(w, "No se pudo restablecer la seguridad de acceso", http.StatusInternalServerError)
			return
		}

		policy := resolveEmpresaUsuarioPasswordPolicy(dbSuper)
		if rotationRequired, edadDias := empresaUsuarioPasswordRotationRequired(item, policy, time.Now()); rotationRequired {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":                         false,
				"password_rotation_required": true,
				"empresa_id":                 item.EmpresaID,
				"email":                      item.Email,
				"password_age_days":          edadDias,
				"message":                    "Debes cambiar tu contraseña antes de continuar por politica de seguridad.",
				"password_policy":            empresaUsuarioPasswordPolicyToMap(policy),
			})
			return
		}

		if err := createEmpresaUsuarioSessionAndRespond(w, r, dbSuper, item); err != nil {
			log.Printf("[usuarios_empresa] failed to create session (login) empresa_id=%d email=%s error=%v", item.EmpresaID, item.Email, err)
			http.Error(w, "No se pudo iniciar sesión del usuario", http.StatusInternalServerError)
			return
		}
		warmEmpresaPermissionSnapshot(dbEmp, dbSuper, item)
	}
}

// EmpresaUsuarioSetPasswordHandler define la contraseña en el primer ingreso y abre sesión.
func EmpresaUsuarioSetPasswordHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			EmpresaID          int64  `json:"empresa_id"`
			Email              string `json:"email"`
			DocumentoIdentidad string `json:"documento_identidad"`
			Password           string `json:"password"`
			PasswordConfirm    string `json:"password_confirm"`
			TokenInvitacion    string `json:"token_invitacion"`
			AcceptContract     bool   `json:"accept_contract"`
			RecaptchaToken     string `json:"recaptcha_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		email := strings.TrimSpace(payload.Email)
		documento := strings.TrimSpace(payload.DocumentoIdentidad)
		if email == "" || documento == "" {
			http.Error(w, "email y documento_identidad son obligatorios", http.StatusBadRequest)
			return
		}
		if _, err := mail.ParseAddress(email); err != nil {
			http.Error(w, "email inválido", http.StatusBadRequest)
			return
		}
		if payload.EmpresaID <= 0 {
			if qEmpresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && qEmpresaID > 0 {
				payload.EmpresaID = qEmpresaID
			}
		}
		if strings.TrimSpace(payload.Password) == "" {
			http.Error(w, "debes ingresar una contraseña", http.StatusBadRequest)
			return
		}
		if payload.PasswordConfirm != "" && payload.Password != payload.PasswordConfirm {
			http.Error(w, "la confirmación de contraseña no coincide", http.StatusBadRequest)
			return
		}
		if err := validateRecaptchaToken(dbSuper, r, payload.RecaptchaToken); err != nil {
			writeRecaptchaValidationError(w, err)
			return
		}
		invitationToken := strings.TrimSpace(payload.TokenInvitacion)
		if invitationToken == "" {
			http.Error(w, "el registro solo puede completarse desde la invitacion enviada por correo por el administrador", http.StatusForbidden)
			return
		}

		policy := resolveEmpresaUsuarioPasswordPolicy(dbSuper)
		if err := validateEmpresaUsuarioPasswordWithPolicy(payload.Password, policy); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		item, err := dbpkg.GetEmpresaUsuarioByConfirmToken(dbEmp, invitationToken)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "usuario no encontrado", http.StatusNotFound)
				return
			}
			log.Printf("[usuarios_empresa] failed to query user (set_password) empresa_id=%d email=%s error=%v", payload.EmpresaID, email, err)
			http.Error(w, "No se pudo validar el usuario", http.StatusInternalServerError)
			return
		}
		if payload.EmpresaID > 0 && item.EmpresaID != payload.EmpresaID {
			http.Error(w, "invitacion invalida para la empresa indicada", http.StatusUnauthorized)
			return
		}
		if !strings.EqualFold(strings.TrimSpace(item.Email), email) {
			http.Error(w, "invitacion invalida para el correo indicado", http.StatusUnauthorized)
			return
		}
		if !strings.EqualFold(strings.TrimSpace(item.DocumentoIdentidad), documento) {
			http.Error(w, "documento inválido", http.StatusUnauthorized)
			return
		}
		if status, msg := validateEmpresaUsuarioInvitationToken(item, invitationToken, time.Now()); status != http.StatusOK {
			http.Error(w, msg, status)
			return
		}
		if empresaUsuarioEstadoBloqueaPrimerIngreso(item) {
			http.Error(w, "tu usuario está inactivo", http.StatusForbidden)
			return
		}
		if item.PasswordSet == 1 && strings.TrimSpace(item.PasswordHash) != "" {
			http.Error(w, "el usuario ya tiene contraseña configurada", http.StatusConflict)
			return
		}

		contract, accepted, err := ensureEmpresaUsuarioCurrentContractAccepted(dbEmp, dbSuper, item, payload.AcceptContract)
		if err != nil {
			log.Printf("[usuarios_empresa] failed to verify contract acceptance (set_password) empresa_id=%d email=%s error=%v", item.EmpresaID, item.Email, err)
			http.Error(w, "No se pudo validar el contrato vigente", http.StatusInternalServerError)
			return
		}
		if !accepted {
			writeEmpresaUsuarioContractRequirement(w, item, contract, "Debes aceptar el contrato vigente antes de completar tu registro.")
			return
		}

		hash, salt, err := generateEmpresaUsuarioPasswordHash(payload.Password)
		if err != nil {
			http.Error(w, "no se pudo generar password hash", http.StatusInternalServerError)
			return
		}
		if err := dbpkg.CompleteEmpresaUsuarioInvitationPassword(dbEmp, item.EmpresaID, item.ID, hash, salt); err != nil {
			log.Printf("[usuarios_empresa] failed to set password empresa_id=%d id=%d email=%s error=%v", item.EmpresaID, item.ID, item.Email, err)
			http.Error(w, "No se pudo actualizar la contraseña", http.StatusInternalServerError)
			return
		}

		item.PasswordHash = hash
		item.PasswordSalt = salt
		item.PasswordSet = 1
		item.EmailConfirmado = 1
		item.EmailConfirmToken = ""
		item.EmailConfirmExpira = ""
		item.Estado = "activo"

		if err := createEmpresaUsuarioSessionAndRespond(w, r, dbSuper, item); err != nil {
			log.Printf("[usuarios_empresa] failed to create session (set_password) empresa_id=%d email=%s error=%v", item.EmpresaID, item.Email, err)
			http.Error(w, "No se pudo iniciar sesión del usuario", http.StatusInternalServerError)
			return
		}
		warmEmpresaPermissionSnapshot(dbEmp, dbSuper, item)
	}
}

// EmpresaUsuarioRequestPasswordRecoveryHandler genera un token de recuperación de contraseña.
func EmpresaUsuarioRequestPasswordRecoveryHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			EmpresaID      int64  `json:"empresa_id"`
			Email          string `json:"email"`
			RecaptchaToken string `json:"recaptcha_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		email := strings.TrimSpace(payload.Email)
		if email == "" {
			http.Error(w, "email es obligatorio", http.StatusBadRequest)
			return
		}
		if _, err := mail.ParseAddress(email); err != nil {
			http.Error(w, "email inválido", http.StatusBadRequest)
			return
		}
		if payload.EmpresaID <= 0 {
			if qEmpresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && qEmpresaID > 0 {
				payload.EmpresaID = qEmpresaID
			}
		}
		if err := validateRecaptchaToken(dbSuper, r, payload.RecaptchaToken); err != nil {
			writeRecaptchaValidationError(w, err)
			return
		}

		respondAccepted := func(delivery string) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":       true,
				"delivery": delivery,
				"message":  "Si el correo existe, enviaremos instrucciones para recuperar la contraseña.",
			})
		}

		item, err := resolveUniqueEmpresaUsuarioByEmail(dbEmp, email, payload.EmpresaID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) || errors.Is(err, errEmpresaUsuarioEmailAmbiguo) {
				respondAccepted("masked")
				return
			}
			log.Printf("[usuarios_empresa] failed to query user (password_recovery_request) empresa_id=%d email=%s error=%v", payload.EmpresaID, email, err)
			http.Error(w, "No se pudo procesar la solicitud", http.StatusInternalServerError)
			return
		}
		if item.EmailConfirmado != 1 || strings.EqualFold(strings.TrimSpace(item.Estado), "inactivo") {
			respondAccepted("masked")
			return
		}

		token, expira, err := newPasswordRecoveryTokenAndExpiration()
		if err != nil {
			http.Error(w, "failed to generate recovery token", http.StatusInternalServerError)
			return
		}
		if err := dbpkg.SetEmpresaUsuarioPasswordResetToken(dbEmp, item.EmpresaID, item.ID, token, expira); err != nil {
			log.Printf("[usuarios_empresa] failed to set recovery token empresa_id=%d id=%d email=%s error=%v", item.EmpresaID, item.ID, item.Email, err)
			http.Error(w, "No se pudo registrar la recuperación", http.StatusInternalServerError)
			return
		}

		if _, mailErr := sendEmpresaUsuarioPasswordRecoveryEmail(r, dbEmp, dbSuper, item.EmpresaID, item.Email, item.Nombre, token); mailErr != nil {
			log.Printf("[usuarios_empresa] password recovery email not sent empresa_id=%d id=%d email=%s error=%v", item.EmpresaID, item.ID, item.Email, mailErr)
			respondAccepted("manual")
			return
		}

		respondAccepted("email")
	}
}

// EmpresaUsuarioRequestInvitationRecoveryHandler reenvia una invitacion pendiente sin revelar si el correo existe.
func EmpresaUsuarioRequestInvitationRecoveryHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			EmpresaID      int64  `json:"empresa_id"`
			Email          string `json:"email"`
			RecaptchaToken string `json:"recaptcha_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		email := strings.TrimSpace(payload.Email)
		if email == "" {
			http.Error(w, "email es obligatorio", http.StatusBadRequest)
			return
		}
		if _, err := mail.ParseAddress(email); err != nil {
			http.Error(w, "email inválido", http.StatusBadRequest)
			return
		}
		if payload.EmpresaID <= 0 {
			if qEmpresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && qEmpresaID > 0 {
				payload.EmpresaID = qEmpresaID
			}
		}
		if err := validateRecaptchaToken(dbSuper, r, payload.RecaptchaToken); err != nil {
			writeRecaptchaValidationError(w, err)
			return
		}

		respondAccepted := func(delivery string) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":       true,
				"delivery": delivery,
				"message":  "Si ese correo tiene una invitacion pendiente, enviaremos nuevamente el email de invitacion.",
			})
		}

		item, err := resolveUniqueEmpresaUsuarioByEmail(dbEmp, email, payload.EmpresaID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) || errors.Is(err, errEmpresaUsuarioEmailAmbiguo) {
				respondAccepted("masked")
				return
			}
			log.Printf("[usuarios_empresa] failed to query user (invitation_recovery) empresa_id=%d email=%s error=%v", payload.EmpresaID, email, err)
			http.Error(w, "No se pudo procesar la solicitud", http.StatusInternalServerError)
			return
		}
		if strings.EqualFold(strings.TrimSpace(item.Estado), "inactivo") || (item.PasswordSet == 1 && strings.TrimSpace(item.PasswordHash) != "") {
			respondAccepted("masked")
			return
		}

		token, expira, err := newEmailConfirmationTokenAndExpiration()
		if err != nil {
			http.Error(w, "failed to generate invitation token", http.StatusInternalServerError)
			return
		}
		if err := dbpkg.SetEmpresaUsuarioConfirmToken(dbEmp, item.EmpresaID, item.ID, token, expira); err != nil {
			log.Printf("[usuarios_empresa] failed to set invitation recovery token empresa_id=%d id=%d email=%s error=%v", item.EmpresaID, item.ID, item.Email, err)
			http.Error(w, "No se pudo preparar la invitacion", http.StatusInternalServerError)
			return
		}

		if _, mailErr := sendEmpresaUsuarioConfirmationEmail(r, dbEmp, dbSuper, item.EmpresaID, item.Email, item.Nombre, token, "Reenvio de invitacion solicitado desde el portal de usuarios."); mailErr != nil {
			log.Printf("[usuarios_empresa] invitation recovery email not sent empresa_id=%d id=%d email=%s error=%v", item.EmpresaID, item.ID, item.Email, mailErr)
			respondAccepted("manual")
			return
		}

		respondAccepted("email")
	}
}

// EmpresaUsuarioResetPasswordHandler permite restablecer contraseña con token de recuperación.
func EmpresaUsuarioResetPasswordHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			EmpresaID       int64  `json:"empresa_id"`
			Email           string `json:"email"`
			Token           string `json:"token"`
			Password        string `json:"password"`
			PasswordConfirm string `json:"password_confirm"`
			RecaptchaToken  string `json:"recaptcha_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		email := strings.TrimSpace(payload.Email)
		token := strings.TrimSpace(payload.Token)
		if payload.EmpresaID <= 0 {
			if qEmpresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && qEmpresaID > 0 {
				payload.EmpresaID = qEmpresaID
			}
		}
		if email == "" || token == "" {
			http.Error(w, "email y token son obligatorios", http.StatusBadRequest)
			return
		}
		if _, err := mail.ParseAddress(email); err != nil {
			http.Error(w, "email inválido", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(payload.Password) == "" {
			http.Error(w, "debes ingresar una contraseña", http.StatusBadRequest)
			return
		}
		if payload.PasswordConfirm != "" && payload.PasswordConfirm != payload.Password {
			http.Error(w, "la confirmación de contraseña no coincide", http.StatusBadRequest)
			return
		}
		if err := validateRecaptchaToken(dbSuper, r, payload.RecaptchaToken); err != nil {
			writeRecaptchaValidationError(w, err)
			return
		}

		policy := resolveEmpresaUsuarioPasswordPolicy(dbSuper)
		if err := validateEmpresaUsuarioPasswordWithPolicy(payload.Password, policy); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		item, err := resolveEmpresaUsuarioForPasswordReset(dbEmp, email, token, payload.EmpresaID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "token de recuperación inválido", http.StatusUnauthorized)
				return
			}
			if errors.Is(err, errEmpresaUsuarioEmailAmbiguo) {
				http.Error(w, "este correo esta asociado a mas de una empresa; solicita un nuevo token de recuperacion", http.StatusConflict)
				return
			}
			log.Printf("[usuarios_empresa] failed to query user (password_reset) empresa_id=%d email=%s error=%v", payload.EmpresaID, email, err)
			http.Error(w, "No se pudo validar el usuario", http.StatusInternalServerError)
			return
		}
		if item.EmailConfirmado != 1 {
			http.Error(w, "debes confirmar tu correo antes de restablecer contraseña", http.StatusForbidden)
			return
		}
		if strings.EqualFold(strings.TrimSpace(item.Estado), "inactivo") {
			http.Error(w, "tu usuario está inactivo", http.StatusForbidden)
			return
		}

		storedToken := strings.TrimSpace(item.PasswordResetToken)
		if storedToken == "" || subtle.ConstantTimeCompare([]byte(token), []byte(storedToken)) != 1 {
			http.Error(w, "token de recuperación inválido", http.StatusUnauthorized)
			return
		}
		expiraAt, ok := parseEmpresaUsuarioDateTime(strings.TrimSpace(item.PasswordResetExpira))
		if !ok || time.Now().After(expiraAt) {
			_ = dbpkg.ClearEmpresaUsuarioPasswordResetToken(dbEmp, item.EmpresaID, item.ID)
			http.Error(w, "token de recuperación expirado", http.StatusUnauthorized)
			return
		}

		hash, salt, err := generateEmpresaUsuarioPasswordHash(payload.Password)
		if err != nil {
			http.Error(w, "no se pudo generar password hash", http.StatusInternalServerError)
			return
		}
		if err := dbpkg.SetEmpresaUsuarioPassword(dbEmp, item.EmpresaID, item.ID, hash, salt); err != nil {
			log.Printf("[usuarios_empresa] failed to reset password empresa_id=%d id=%d email=%s error=%v", item.EmpresaID, item.ID, item.Email, err)
			http.Error(w, "No se pudo actualizar la contraseña", http.StatusInternalServerError)
			return
		}

		item.PasswordHash = hash
		item.PasswordSalt = salt
		item.PasswordSet = 1

		if err := createEmpresaUsuarioSessionAndRespond(w, r, dbSuper, item); err != nil {
			log.Printf("[usuarios_empresa] failed to create session (password_reset) empresa_id=%d email=%s error=%v", item.EmpresaID, item.Email, err)
			http.Error(w, "No se pudo iniciar sesión del usuario", http.StatusInternalServerError)
			return
		}
		warmEmpresaPermissionSnapshot(dbEmp, dbSuper, item)
	}
}

// EmpresaUsuarioChangePasswordHandler permite cambiar contraseña con credenciales actuales.
func EmpresaUsuarioChangePasswordHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			EmpresaID            int64  `json:"empresa_id"`
			Email                string `json:"email"`
			CurrentPassword      string `json:"current_password"`
			PasswordActual       string `json:"password_actual"`
			NewPassword          string `json:"new_password"`
			PasswordNueva        string `json:"password_nueva"`
			NewPasswordConfirm   string `json:"new_password_confirm"`
			PasswordNuevaConfirm string `json:"password_nueva_confirm"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		if payload.EmpresaID <= 0 {
			if qEmpresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && qEmpresaID > 0 {
				payload.EmpresaID = qEmpresaID
			}
		}

		email := strings.TrimSpace(payload.Email)
		if email == "" {
			http.Error(w, "email es obligatorio", http.StatusBadRequest)
			return
		}
		if _, err := mail.ParseAddress(email); err != nil {
			http.Error(w, "email inválido", http.StatusBadRequest)
			return
		}

		currentPassword := payload.CurrentPassword
		if strings.TrimSpace(currentPassword) == "" {
			currentPassword = payload.PasswordActual
		}
		newPassword := payload.NewPassword
		if strings.TrimSpace(newPassword) == "" {
			newPassword = payload.PasswordNueva
		}
		newPasswordConfirm := payload.NewPasswordConfirm
		if strings.TrimSpace(newPasswordConfirm) == "" {
			newPasswordConfirm = payload.PasswordNuevaConfirm
		}

		if strings.TrimSpace(currentPassword) == "" || strings.TrimSpace(newPassword) == "" {
			http.Error(w, "current_password y new_password son obligatorios", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(newPasswordConfirm) != "" && newPasswordConfirm != newPassword {
			http.Error(w, "la confirmación de contraseña no coincide", http.StatusBadRequest)
			return
		}

		item, currentPasswordAlreadyVerified, err := resolveEmpresaUsuarioForPasswordLogin(dbEmp, email, currentPassword, payload.EmpresaID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "credenciales inválidas", http.StatusUnauthorized)
				return
			}
			if errors.Is(err, errEmpresaUsuarioEmailAmbiguo) {
				http.Error(w, "este correo esta asociado a mas de una empresa; solicita al administrador usar un correo unico por empresa", http.StatusConflict)
				return
			}
			log.Printf("[usuarios_empresa] failed to query user (change_password) empresa_id=%d email=%s error=%v", payload.EmpresaID, email, err)
			http.Error(w, "No se pudo validar el usuario", http.StatusInternalServerError)
			return
		}

		if item.EmailConfirmado != 1 {
			http.Error(w, "debes confirmar tu correo antes de cambiar contraseña", http.StatusForbidden)
			return
		}
		if strings.EqualFold(strings.TrimSpace(item.Estado), "inactivo") {
			http.Error(w, "tu usuario está inactivo", http.StatusForbidden)
			return
		}
		if item.PasswordSet != 1 || strings.TrimSpace(item.PasswordHash) == "" || strings.TrimSpace(item.PasswordSalt) == "" {
			http.Error(w, "debes establecer tu contraseña inicial antes de cambiarla", http.StatusConflict)
			return
		}
		if !currentPasswordAlreadyVerified && !verifyEmpresaUsuarioPassword(currentPassword, item) {
			http.Error(w, "credenciales inválidas", http.StatusUnauthorized)
			return
		}
		if currentPassword == newPassword {
			http.Error(w, "la nueva contraseña debe ser diferente a la actual", http.StatusBadRequest)
			return
		}

		policy := resolveEmpresaUsuarioPasswordPolicy(dbSuper)
		if err := validateEmpresaUsuarioPasswordWithPolicy(newPassword, policy); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		hash, salt, err := generateEmpresaUsuarioPasswordHash(newPassword)
		if err != nil {
			http.Error(w, "no se pudo generar password hash", http.StatusInternalServerError)
			return
		}
		if err := dbpkg.SetEmpresaUsuarioPassword(dbEmp, item.EmpresaID, item.ID, hash, salt); err != nil {
			log.Printf("[usuarios_empresa] failed to change password empresa_id=%d id=%d email=%s error=%v", item.EmpresaID, item.ID, item.Email, err)
			http.Error(w, "No se pudo actualizar la contraseña", http.StatusInternalServerError)
			return
		}

		item.PasswordHash = hash
		item.PasswordSalt = salt
		item.PasswordSet = 1

		if err := createEmpresaUsuarioSessionAndRespond(w, r, dbSuper, item); err != nil {
			log.Printf("[usuarios_empresa] failed to create session (change_password) empresa_id=%d email=%s error=%v", item.EmpresaID, item.Email, err)
			http.Error(w, "No se pudo iniciar sesión del usuario", http.StatusInternalServerError)
			return
		}
		warmEmpresaPermissionSnapshot(dbEmp, dbSuper, item)
	}
}

// ConfirmarCorreoUsuarioHandler confirma el correo desde un enlace enviado al usuario.
func ConfirmarCorreoUsuarioHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		if token == "" {
			http.Error(w, "token required", http.StatusBadRequest)
			return
		}
		item, err := dbpkg.GetEmpresaUsuarioByConfirmToken(dbEmp, token)
		if err != nil {
			loginURL := "/login_usuario.html"
			msg := html.EscapeString(err.Error())
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "<html><body style='font-family:sans-serif;background:#10141f;color:#e9eefb;padding:24px'><h2>No se pudo abrir la invitacion</h2><p>%s</p><p><a href='%s' style='color:#7fb2ff'>Volver al login de usuario</a></p></body></html>", msg, html.EscapeString(loginURL))
			return
		}
		if status, msg := validateEmpresaUsuarioInvitationToken(item, token, time.Now()); status != http.StatusOK {
			loginURL := "/login_usuario.html"
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(status)
			fmt.Fprintf(w, "<html><body style='font-family:sans-serif;background:#10141f;color:#e9eefb;padding:24px'><h2>No se pudo abrir la invitacion</h2><p>%s</p><p><a href='%s' style='color:#7fb2ff'>Volver al login de usuario</a></p></body></html>", html.EscapeString(msg), html.EscapeString(loginURL))
			return
		}
		invitationURL := buildEmpresaUsuarioInvitationURL(r, dbEmp, nil, item.EmpresaID, item.Email, token)
		http.Redirect(w, r, invitationURL, http.StatusSeeOther)
	}
}

// GmailConfigHandler gestiona configuración de envío SMTP por Gmail.
func GmailConfigHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			smtpEmail, _, _, smtpEmailUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "gmail.smtp_email")
			appPass, _, _, appPassUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "gmail.smtp_app_password")
			fromName, _, _, fromNameUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "gmail.smtp_from_name")
			host, _, _, hostUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "gmail.smtp_host")
			port, _, _, portUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "gmail.smtp_port")
			baseURL, _, _, baseURLUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "gmail.confirm_base_url")
			whatsAppNumber, _, _, whatsAppNumberUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "portal.whatsapp_contact_number")
			restartAlertTo, _, _, restartAlertUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "gmail.restart_alert_to")
			restartAlertEnabledRaw, _, _, restartAlertEnabledUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "gmail.restart_alert_enabled")
			restartAlertEnabled := parseEmpresaUsuarioBool(restartAlertEnabledRaw, true)

			if host == "" {
				host = "smtp.gmail.com"
			}
			if port == "" {
				port = "587"
			}
			if fromName == "" {
				fromName = "Powerful Control System"
			}

			masked := ""
			if appPass != "" {
				masked = "********"
			}
			appPassDecryptOK := true
			appPassDecryptMessage := ""
			if strings.TrimSpace(appPass) != "" {
				if _, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_app_password"); err != nil {
					appPassDecryptOK = false
					appPassDecryptMessage = friendlyEmpresaUsuarioMailConfigError(err).Error()
				}
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"smtp_email_set":                  strings.TrimSpace(smtpEmail) != "",
				"smtp_email":                      smtpEmail,
				"smtp_email_updated":              smtpEmailUpdated,
				"smtp_app_password_set":           strings.TrimSpace(appPass) != "",
				"smtp_app_password_decrypt_ok":    appPassDecryptOK,
				"smtp_app_password_error":         appPassDecryptMessage,
				"smtp_app_password_masked":        masked,
				"smtp_app_password_updated":       appPassUpdated,
				"smtp_from_name":                  fromName,
				"smtp_from_name_updated":          fromNameUpdated,
				"smtp_host":                       host,
				"smtp_host_updated":               hostUpdated,
				"smtp_port":                       port,
				"smtp_port_updated":               portUpdated,
				"confirm_base_url":                baseURL,
				"confirm_base_url_updated":        baseURLUpdated,
				"whatsapp_contact_number":         whatsAppNumber,
				"whatsapp_contact_number_set":     strings.TrimSpace(whatsAppNumber) != "",
				"whatsapp_contact_number_updated": whatsAppNumberUpdated,
				"restart_alert_to_set":            strings.TrimSpace(restartAlertTo) != "",
				"restart_alert_to":                restartAlertTo,
				"restart_alert_to_updated":        restartAlertUpdated,
				"restart_alert_enabled":           restartAlertEnabled,
				"restart_alert_enabled_set":       strings.TrimSpace(restartAlertEnabledRaw) != "",
				"restart_alert_enabled_updated":   restartAlertEnabledUpdated,
				"encryption_available":            utils.EncryptionAvailable(),
			})
			return

		case http.MethodPost, http.MethodPut:
			if strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("action")), "test") {
				if err := sendSuperGmailTestEmail(dbSuper, adminEmailFromRequest(r)); err != nil {
					status := http.StatusInternalServerError
					if isEmpresaUsuarioMailActionableConfigError(err) {
						status = http.StatusBadRequest
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(status)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"sent":      false,
						"recipient": superGmailTestRecipient,
						"error":     friendlyEmpresaUsuarioMailConfigError(err).Error(),
					})
					return
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"sent":      true,
					"recipient": superGmailTestRecipient,
					"message":   "Correo de prueba enviado correctamente a " + superGmailTestRecipient,
				})
				return
			}

			var payload struct {
				SMTPEmail             string `json:"smtp_email"`
				SMTPAppPass           string `json:"smtp_app_password"`
				SMTPFromName          string `json:"smtp_from_name"`
				SMTPHost              string `json:"smtp_host"`
				SMTPPort              string `json:"smtp_port"`
				ConfirmBaseURL        string `json:"confirm_base_url"`
				WhatsAppContactNumber string `json:"whatsapp_contact_number"`
				RestartAlertTo        string `json:"restart_alert_to"`
				RestartAlertEnabled   *bool  `json:"restart_alert_enabled"`
				Encrypt               bool   `json:"encrypt"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
				return
			}

			smtpEmail := strings.TrimSpace(payload.SMTPEmail)
			if smtpEmail != "" {
				if _, err := mail.ParseAddress(smtpEmail); err != nil {
					http.Error(w, "smtp_email inválido", http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, "gmail.smtp_email", smtpEmail, false); err != nil {
					http.Error(w, "failed to save gmail.smtp_email: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			if strings.TrimSpace(payload.SMTPAppPass) != "" {
				appPass := strings.TrimSpace(payload.SMTPAppPass)
				if !utils.EncryptionAvailable() {
					http.Error(w, "encryption required: CONFIG_ENC_KEY not set", http.StatusBadRequest)
					return
				}
				encVal, err := utils.EncryptString(appPass)
				if err != nil {
					http.Error(w, "encryption failed: "+err.Error(), http.StatusInternalServerError)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, "gmail.smtp_app_password", encVal, true); err != nil {
					http.Error(w, "failed to save gmail.smtp_app_password: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			if strings.TrimSpace(payload.SMTPFromName) != "" {
				if err := dbpkg.SetConfigValue(dbSuper, "gmail.smtp_from_name", strings.TrimSpace(payload.SMTPFromName), false); err != nil {
					http.Error(w, "failed to save gmail.smtp_from_name: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			smtpHost := strings.TrimSpace(payload.SMTPHost)
			if smtpHost != "" {
				if err := dbpkg.SetConfigValue(dbSuper, "gmail.smtp_host", smtpHost, false); err != nil {
					http.Error(w, "failed to save gmail.smtp_host: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			smtpPort := strings.TrimSpace(payload.SMTPPort)
			if smtpPort != "" {
				portInt, err := strconv.Atoi(smtpPort)
				if err != nil || portInt <= 0 || portInt > 65535 {
					http.Error(w, "smtp_port inválido", http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, "gmail.smtp_port", smtpPort, false); err != nil {
					http.Error(w, "failed to save gmail.smtp_port: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			confirmBaseURL := strings.TrimSpace(payload.ConfirmBaseURL)
			if confirmBaseURL != "" {
				u, err := url.ParseRequestURI(confirmBaseURL)
				if err != nil || u.Scheme == "" || u.Host == "" {
					http.Error(w, "confirm_base_url inválida", http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, "gmail.confirm_base_url", confirmBaseURL, false); err != nil {
					http.Error(w, "failed to save gmail.confirm_base_url: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			whatsAppContactNumber := strings.TrimSpace(payload.WhatsAppContactNumber)
			if whatsAppContactNumber != "" {
				normalizedWhatsApp := normalizePortalWhatsAppContactNumber(whatsAppContactNumber)
				if normalizedWhatsApp == "" {
					http.Error(w, "whatsapp_contact_number inválido", http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, "portal.whatsapp_contact_number", normalizedWhatsApp, false); err != nil {
					http.Error(w, "failed to save portal.whatsapp_contact_number: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			restartAlertTo := strings.TrimSpace(payload.RestartAlertTo)
			if restartAlertTo != "" {
				if _, err := mail.ParseAddress(restartAlertTo); err != nil {
					http.Error(w, "restart_alert_to inválido", http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, "gmail.restart_alert_to", restartAlertTo, false); err != nil {
					http.Error(w, "failed to save gmail.restart_alert_to: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			if payload.RestartAlertEnabled != nil {
				raw := "0"
				if *payload.RestartAlertEnabled {
					raw = "1"
				}
				if err := dbpkg.SetConfigValue(dbSuper, "gmail.restart_alert_enabled", raw, false); err != nil {
					http.Error(w, "failed to save gmail.restart_alert_enabled: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"saved": true})
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

func sendSuperGmailTestEmail(dbSuper *sql.DB, usuarioCreador string) error {
	if dbSuper == nil {
		return fmt.Errorf("db super no disponible")
	}

	smtpEmail, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_email")
	if err != nil {
		return friendlyEmpresaUsuarioMailConfigError(err)
	}
	smtpEmail = strings.TrimSpace(smtpEmail)
	if smtpEmail == "" {
		return fmt.Errorf("gmail.smtp_email no configurado")
	}

	smtpPass, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_app_password")
	if err != nil {
		return friendlyEmpresaUsuarioMailConfigError(err)
	}
	smtpPass = strings.TrimSpace(smtpPass)
	if smtpPass == "" {
		return fmt.Errorf("gmail.smtp_app_password no configurado")
	}

	smtpHost, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_host")
	smtpPort, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_port")
	fromName, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_from_name")

	smtpHost = strings.TrimSpace(smtpHost)
	if smtpHost == "" {
		smtpHost = "smtp.gmail.com"
	}
	smtpPort = strings.TrimSpace(smtpPort)
	if smtpPort == "" {
		smtpPort = "587"
	}
	fromName = strings.TrimSpace(fromName)
	if fromName == "" {
		fromName = "Powerful Control System"
	}

	stamp := time.Now().Format(time.RFC3339)
	subject := "Prueba Gmail - Configuracion avanzada Powerful Control System"
	body := "Esta es una prueba del boton Probar Gmail desde configuracion avanzada.\r\n\r\n" +
		"Fecha: " + stamp + "\r\n" +
		"Host SMTP: " + smtpHost + "\r\n" +
		"Puerto SMTP: " + smtpPort + "\r\n" +
		"Remitente: " + smtpEmail + "\r\n" +
		"Destino: " + superGmailTestRecipient + "\r\n"

	if isEmpresaUsuarioMailTestMode(dbSuper) {
		metadataJSON := fmt.Sprintf(`{"mail_mode":%q,"smtp_host":%q,"smtp_port":%q,"from":%q}`, "test", smtpHost, smtpPort, smtpEmail)
		return captureEmpresaUsuarioMailNotification(
			dbSuper,
			superCorreoNotificacionTipoPruebaGmail,
			0,
			superGmailTestRecipient,
			subject,
			body,
			"",
			metadataJSON,
			usuarioCreador,
		)
	}

	mailHostForAuth := smtpHost
	if strings.Contains(smtpHost, ":") {
		if host, _, err := net.SplitHostPort(smtpHost); err == nil && strings.TrimSpace(host) != "" {
			mailHostForAuth = host
		}
	}
	addr := smtpHost
	if !strings.Contains(addr, ":") {
		addr = net.JoinHostPort(smtpHost, smtpPort)
	}

	auth := smtp.PlainAuth("", smtpEmail, smtpPass, mailHostForAuth)
	msg := "From: " + fromName + " <" + smtpEmail + ">\r\n" +
		"To: " + superGmailTestRecipient + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body

	if err := smtp.SendMail(addr, auth, smtpEmail, []string{superGmailTestRecipient}, []byte(msg)); err != nil {
		return err
	}
	return nil
}

func handleEmpresaUsuarioFotoUpload(r *http.Request, dbEmp, dbSuper *sql.DB, empresaID int64) (int64, string, error) {
	if empresaID <= 0 {
		return 0, "", fmt.Errorf("empresa_id requerido")
	}
	maxBytes := domoticaStorageMaxImageBytes(dbSuper, empresaID)
	if err := r.ParseMultipartForm(maxBytes + (1 << 20)); err != nil {
		return 0, "", fmt.Errorf("payload multipart invalido")
	}
	userID, err := parseInt64Form(r, "usuario_id")
	if err != nil || userID <= 0 {
		userID, err = parseInt64Form(r, "id")
	}
	if err != nil || userID <= 0 {
		return 0, "", fmt.Errorf("usuario_id requerido")
	}
	if _, err := dbpkg.GetEmpresaUsuarioByID(dbEmp, empresaID, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, "", fmt.Errorf("usuario no encontrado")
		}
		return 0, "", err
	}
	file, header, err := r.FormFile("foto")
	if err != nil {
		return 0, "", fmt.Errorf("foto requerida")
	}
	defer file.Close()
	if header.Size > maxBytes {
		return 0, "", fmt.Errorf("la imagen supera el tamano maximo permitido de %d KB", maxBytes/1024)
	}
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(header.Filename)))
	allowed := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true}
	if !allowed[ext] {
		return 0, "", fmt.Errorf("extension de imagen no permitida")
	}
	folder := domoticaEmpresaStorageFolder(dbEmp, empresaID)
	dir := filepath.Join(resolveWebRootDir(), "uploads", "empresas", folder, "imagenes", "usuarios")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, "", fmt.Errorf("no se pudo preparar carpeta de imagenes")
	}
	fileName := fmt.Sprintf("usuario_%d_%d%s", userID, time.Now().UnixNano(), ext)
	absPath := filepath.Join(dir, fileName)
	out, err := os.Create(absPath)
	if err != nil {
		return 0, "", fmt.Errorf("no se pudo crear imagen")
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		return 0, "", fmt.Errorf("no se pudo guardar imagen")
	}
	photoURL := "/uploads/empresas/" + folder + "/imagenes/usuarios/" + fileName
	if err := dbpkg.UpdateEmpresaUsuarioFoto(dbEmp, empresaID, userID, photoURL); err != nil {
		return 0, "", err
	}
	return userID, photoURL, nil
}

func validateEmpresaUsuarioPayload(empresaID int64, email, nombre string, rolUsuarioID int64) error {
	if empresaID <= 0 {
		return fmt.Errorf("empresa_id required")
	}
	if strings.TrimSpace(nombre) == "" {
		return fmt.Errorf("nombre required")
	}
	if strings.TrimSpace(email) == "" {
		return fmt.Errorf("email required")
	}
	if _, err := mail.ParseAddress(strings.TrimSpace(email)); err != nil {
		return fmt.Errorf("email inválido")
	}
	if rolUsuarioID <= 0 {
		return fmt.Errorf("rol_usuario_id required")
	}
	return nil
}

func resolveTipoEmpresaIDForEmpresa(dbEmp, dbSuper *sql.DB, empresaID int64) (int64, *dbpkg.Empresa, error) {
	empresa, err := dbpkg.GetEmpresaByID(dbEmp, empresaID)
	if err != nil {
		return 0, nil, err
	}
	if empresa.TipoID > 0 {
		return empresa.TipoID, empresa, nil
	}

	candidateNames := []string{
		strings.TrimSpace(empresa.TipoNombre),
		strings.TrimSpace(empresa.Nombre),
	}
	for _, name := range candidateNames {
		if name == "" {
			continue
		}
		row := dbSuper.QueryRow(`SELECT id FROM tipos_de_empresas WHERE lower(nombre) = lower(?) LIMIT 1`, name)
		var tipoID int64
		if err := row.Scan(&tipoID); err == nil && tipoID > 0 {
			return tipoID, empresa, nil
		}
	}
	return 0, empresa, fmt.Errorf("empresa sin tipo de empresa asociado")
}

func resolveRolNombreValidoParaEmpresa(dbEmp, dbSuper *sql.DB, empresaID, rolID int64) (string, error) {
	if _, _, err := resolveTipoEmpresaIDForEmpresa(dbEmp, dbSuper, empresaID); err != nil {
		return "", err
	}

	row := dbSuper.QueryRow(`SELECT nombre, COALESCE(estado, 'activo') FROM roles_de_usuario WHERE id = ? LIMIT 1`, rolID)
	var nombre string
	var estado string
	if err := row.Scan(&nombre, &estado); err != nil {
		return "", err
	}
	if strings.TrimSpace(nombre) == "" {
		return "", fmt.Errorf("rol sin nombre")
	}
	if strings.EqualFold(strings.TrimSpace(estado), "inactivo") {
		return "", fmt.Errorf("el rol está inactivo")
	}
	return nombre, nil
}

func newEmailConfirmationTokenAndExpiration() (string, string, error) {
	token, err := utils.GenerateSecureToken(32)
	if err != nil {
		return "", "", err
	}
	expira := time.Now().Add(48 * time.Hour).Format("2006-01-02 15:04:05")
	return token, expira, nil
}

func validateEmpresaUsuarioInvitationToken(item *dbpkg.EmpresaUsuario, token string, now time.Time) (int, string) {
	if item == nil {
		return http.StatusForbidden, "invitacion invalida"
	}
	storedToken := strings.TrimSpace(item.EmailConfirmToken)
	token = strings.TrimSpace(token)
	if storedToken == "" || token == "" || subtle.ConstantTimeCompare([]byte(storedToken), []byte(token)) != 1 {
		return http.StatusForbidden, "invitacion invalida o ya utilizada"
	}
	expiraRaw := strings.TrimSpace(item.EmailConfirmExpira)
	if expiraRaw == "" {
		return http.StatusForbidden, "invitacion sin vencimiento valido"
	}
	expiraAt, err := time.ParseInLocation("2006-01-02 15:04:05", expiraRaw, time.Local)
	if err != nil {
		return http.StatusForbidden, "invitacion con vencimiento invalido"
	}
	if now.After(expiraAt) {
		return http.StatusGone, "invitacion expirada; solicita al administrador que reenvie la invitacion"
	}
	return http.StatusOK, ""
}

func newPasswordRecoveryTokenAndExpiration() (string, string, error) {
	token, err := utils.GenerateSecureToken(32)
	if err != nil {
		return "", "", err
	}
	expira := time.Now().Add(empresaUsuarioRecuperacionTTL).Format("2006-01-02 15:04:05")
	return token, expira, nil
}

func resolveEmpresaUsuarioPasswordPolicy(dbSuper *sql.DB) empresaUsuarioPasswordPolicy {
	empresaUsuarioPasswordPolicyMu.Lock()
	if !empresaUsuarioPasswordPolicyLoadedAt.IsZero() && time.Since(empresaUsuarioPasswordPolicyLoadedAt) < empresaUsuarioPasswordPolicyCacheTTL {
		cached := empresaUsuarioPasswordPolicyCached
		empresaUsuarioPasswordPolicyMu.Unlock()
		return cached
	}
	empresaUsuarioPasswordPolicyMu.Unlock()

	policy := defaultEmpresaUsuarioPasswordPolicy()

	policy.MinLength = parseEmpresaUsuarioInt(
		getEmpresaUsuarioConfigValue(dbSuper, "usuarios.password_min_length"),
		policy.MinLength,
		8,
		128,
	)
	policy.RequireUpper = parseEmpresaUsuarioBool(
		getEmpresaUsuarioConfigValue(dbSuper, "usuarios.password_require_uppercase"),
		policy.RequireUpper,
	)
	policy.RequireLower = parseEmpresaUsuarioBool(
		getEmpresaUsuarioConfigValue(dbSuper, "usuarios.password_require_lowercase"),
		policy.RequireLower,
	)
	policy.RequireDigit = parseEmpresaUsuarioBool(
		getEmpresaUsuarioConfigValue(dbSuper, "usuarios.password_require_digit"),
		policy.RequireDigit,
	)
	policy.RequireSymbol = parseEmpresaUsuarioBool(
		getEmpresaUsuarioConfigValue(dbSuper, "usuarios.password_require_symbol"),
		policy.RequireSymbol,
	)
	policy.RotationDays = parseEmpresaUsuarioInt(
		getEmpresaUsuarioConfigValue(dbSuper, "usuarios.password_rotation_days"),
		policy.RotationDays,
		0,
		3650,
	)

	empresaUsuarioPasswordPolicyMu.Lock()
	empresaUsuarioPasswordPolicyCached = policy
	empresaUsuarioPasswordPolicyLoadedAt = time.Now()
	empresaUsuarioPasswordPolicyMu.Unlock()

	return policy
}

func validateEmpresaUsuarioPasswordWithPolicy(password string, policy empresaUsuarioPasswordPolicy) error {
	if strings.TrimSpace(password) == "" {
		return fmt.Errorf("debes ingresar una contraseña")
	}

	runes := []rune(password)
	if len(runes) < policy.MinLength {
		return fmt.Errorf("la contraseña debe tener al menos %d caracteres", policy.MinLength)
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSymbol := false
	hasSpace := false
	for _, r := range runes {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsSpace(r):
			hasSpace = true
		default:
			hasSymbol = true
		}
	}

	if hasSpace {
		return fmt.Errorf("la contraseña no debe contener espacios")
	}

	missing := make([]string, 0)
	if policy.RequireUpper && !hasUpper {
		missing = append(missing, "una letra mayúscula")
	}
	if policy.RequireLower && !hasLower {
		missing = append(missing, "una letra minúscula")
	}
	if policy.RequireDigit && !hasDigit {
		missing = append(missing, "un número")
	}
	if policy.RequireSymbol && !hasSymbol {
		missing = append(missing, "un símbolo")
	}
	if len(missing) > 0 {
		return fmt.Errorf("la contraseña debe incluir %s", strings.Join(missing, ", "))
	}

	return nil
}

func empresaUsuarioPasswordRotationRequired(item *dbpkg.EmpresaUsuario, policy empresaUsuarioPasswordPolicy, now time.Time) (bool, int) {
	if item == nil || policy.RotationDays <= 0 {
		return false, 0
	}
	if item.PasswordSet != 1 || strings.TrimSpace(item.PasswordHash) == "" {
		return false, 0
	}
	if now.IsZero() {
		now = time.Now()
	}

	referenceCandidates := []string{
		strings.TrimSpace(item.PasswordActualizadaEn),
		strings.TrimSpace(item.FechaActualizacion),
		strings.TrimSpace(item.FechaCreacion),
	}

	referenceAt := time.Time{}
	for _, raw := range referenceCandidates {
		if parsed, ok := parseEmpresaUsuarioDateTime(raw); ok {
			referenceAt = parsed
			break
		}
	}
	if referenceAt.IsZero() {
		return true, 0
	}
	if now.Before(referenceAt) {
		return false, 0
	}

	ageDays := int(now.Sub(referenceAt).Hours() / 24)
	if ageDays >= policy.RotationDays {
		return true, ageDays
	}
	return false, ageDays
}

func getEmpresaUsuarioConfigValue(dbSuper *sql.DB, key string) string {
	if dbSuper == nil {
		return ""
	}
	v, err := getDecryptedConfigValue(dbSuper, key)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(v)
}

func parseEmpresaUsuarioInt(raw string, defaultValue, minValue, maxValue int) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultValue
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return defaultValue
	}
	if v < minValue {
		return minValue
	}
	if v > maxValue {
		return maxValue
	}
	return v
}

func parseEmpresaUsuarioBool(raw string, defaultValue bool) bool {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" {
		return defaultValue
	}
	switch raw {
	case "1", "true", "t", "si", "sí", "y", "yes", "on", "activo":
		return true
	case "0", "false", "f", "no", "n", "off", "inactivo":
		return false
	default:
		return defaultValue
	}
}

func isEmpresaUsuarioMailTestMode(dbSuper *sql.DB) bool {
	if parseEmpresaUsuarioBool(os.Getenv("PCS_MAIL_TEST_MODE"), false) {
		return true
	}
	return parseEmpresaUsuarioBool(getEmpresaUsuarioConfigValue(dbSuper, "gmail.smtp_test_mode"), false)
}

func isEmpresaUsuarioMailSecretDecryptError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(msg, "message authentication failed") || strings.Contains(msg, "cipher:")
}

func friendlyEmpresaUsuarioMailConfigError(err error) error {
	if err == nil {
		return nil
	}
	if isEmpresaUsuarioMailSecretDecryptError(err) {
		return fmt.Errorf("la contrasena SMTP guardada no se puede descifrar. Regraba Gmail SMTP en Super administrador > Mensajeria y alertas > Gmail SMTP para cifrarla con la clave actual del servidor")
	}
	return err
}

type empresaUsuarioSMTPConfig struct {
	Email    string
	Password string
	Host     string
	Port     string
	FromName string
}

func getEmpresaUsuarioGmailSMTPConfig(dbSuper *sql.DB) (empresaUsuarioSMTPConfig, error) {
	var cfg empresaUsuarioSMTPConfig
	smtpEmail, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_email")
	if err != nil {
		return cfg, friendlyEmpresaUsuarioMailConfigError(err)
	}
	cfg.Email = strings.TrimSpace(smtpEmail)
	if cfg.Email == "" {
		return cfg, fmt.Errorf("gmail.smtp_email no configurado")
	}

	smtpPass, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_app_password")
	if err != nil {
		return cfg, friendlyEmpresaUsuarioMailConfigError(err)
	}
	cfg.Password = strings.TrimSpace(smtpPass)
	if cfg.Password == "" {
		return cfg, fmt.Errorf("gmail.smtp_app_password no configurado")
	}

	cfg.Host, _ = getDecryptedConfigValue(dbSuper, "gmail.smtp_host")
	cfg.Port, _ = getDecryptedConfigValue(dbSuper, "gmail.smtp_port")
	cfg.FromName, _ = getDecryptedConfigValue(dbSuper, "gmail.smtp_from_name")

	cfg.Host = strings.TrimSpace(cfg.Host)
	if cfg.Host == "" {
		cfg.Host = "smtp.gmail.com"
	}
	cfg.Port = strings.TrimSpace(cfg.Port)
	if cfg.Port == "" {
		cfg.Port = "587"
	}
	cfg.FromName = strings.TrimSpace(cfg.FromName)
	if cfg.FromName == "" {
		cfg.FromName = "Powerful Control System"
	}
	return cfg, nil
}

func sendEmpresaUsuarioSMTPMessage(cfg empresaUsuarioSMTPConfig, toEmail string, msg []byte) error {
	mailHostForAuth := cfg.Host
	if strings.Contains(cfg.Host, ":") {
		if h, _, err := net.SplitHostPort(cfg.Host); err == nil {
			mailHostForAuth = h
		}
	}
	addr := cfg.Host
	if !strings.Contains(addr, ":") {
		addr = net.JoinHostPort(cfg.Host, cfg.Port)
	}
	auth := smtp.PlainAuth("", cfg.Email, cfg.Password, mailHostForAuth)
	return smtp.SendMail(addr, auth, cfg.Email, []string{toEmail}, msg)
}

func buildEmpresaUsuarioMultipartMessage(baseURL, fromName, fromEmail, toEmail, subject, bodyPlain, bodyHTML string) string {
	boundary := "==PCS_BOUNDARY_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	listUnsub := ""
	if u, err := url.Parse(baseURL); err == nil {
		host := u.Host
		if strings.Contains(host, ":") {
			host, _, _ = net.SplitHostPort(host)
		}
		if strings.TrimSpace(host) != "" {
			listUnsub = "<mailto:postmaster@" + strings.TrimSpace(host) + ">"
		}
	}

	headers := "From: " + strings.TrimSpace(fromName) + " <" + strings.TrimSpace(fromEmail) + ">\r\n" +
		"To: " + strings.TrimSpace(toEmail) + "\r\n" +
		"Subject: " + strings.TrimSpace(subject) + "\r\n"
	if listUnsub != "" {
		headers += "List-Unsubscribe: " + listUnsub + "\r\n"
	}
	headers += "MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/alternative; boundary=\"" + boundary + "\"\r\n\r\n"

	return headers +
		"--" + boundary + "\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n\r\n" +
		bodyPlain + "\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n\r\n" +
		bodyHTML + "\r\n" +
		"--" + boundary + "--\r\n"
}

func buildEmpresaUsuarioPlainMessage(fromName, fromEmail, toEmail, subject, body string) string {
	return "From: " + strings.TrimSpace(fromName) + " <" + strings.TrimSpace(fromEmail) + ">\r\n" +
		"To: " + strings.TrimSpace(toEmail) + "\r\n" +
		"Subject: " + strings.TrimSpace(subject) + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body
}

func sendEmpresaUsuarioGmailMultipart(dbSuper *sql.DB, baseURL, toEmail, subject, bodyPlain, bodyHTML string) error {
	cfg, err := getEmpresaUsuarioGmailSMTPConfig(dbSuper)
	if err != nil {
		return err
	}
	msg := buildEmpresaUsuarioMultipartMessage(baseURL, cfg.FromName, cfg.Email, toEmail, subject, bodyPlain, bodyHTML)
	return sendEmpresaUsuarioSMTPMessage(cfg, toEmail, []byte(msg))
}

func sendEmpresaUsuarioGmailPlain(dbSuper *sql.DB, toEmail, subject, body string) error {
	cfg, err := getEmpresaUsuarioGmailSMTPConfig(dbSuper)
	if err != nil {
		return err
	}
	msg := buildEmpresaUsuarioPlainMessage(cfg.FromName, cfg.Email, toEmail, subject, body)
	return sendEmpresaUsuarioSMTPMessage(cfg, toEmail, []byte(msg))
}

func empresaUsuarioMailuFallbackEnabled(dbSuper *sql.DB) bool {
	cfg := getCorporateEmailConfig(dbSuper)
	if cfg.Enabled {
		return true
	}
	return corporateEmailEnvBool([]string{"EMAIL_CORPORATIVO_ENABLED", "MAILU_ENABLED", "PCS_MAILU_SMTP_FALLBACK"}, false)
}

func empresaUsuarioMailuSender(dbSuper *sql.DB) (string, string) {
	cfg := getCorporateEmailConfig(dbSuper)
	domain := normalizeCorporateEmailDomain(cfg.Domain)
	if domain == "" {
		domain = normalizeCorporateEmailDomain(firstNonEmptyEnv("EMAIL_CORPORATIVO_DOMAIN", "MAILU_DOMAIN"))
	}
	if domain == "" {
		domain = "powerfulcontrolsystem.com"
	}
	fromName := strings.TrimSpace(getEmpresaUsuarioConfigValue(dbSuper, "gmail.smtp_from_name"))
	if fromName == "" {
		fromName = "Powerful Control System"
	}
	return fromName, "postmaster@" + domain
}

func sanitizeEmpresaUsuarioMailerError(err error, output []byte) string {
	parts := []string{}
	if err != nil {
		parts = append(parts, strings.TrimSpace(err.Error()))
	}
	if trimmed := strings.TrimSpace(string(output)); trimmed != "" {
		parts = append(parts, trimmed)
	}
	msg := strings.Join(parts, " - ")
	msg = strings.ReplaceAll(msg, "\r", " ")
	msg = strings.ReplaceAll(msg, "\n", " ")
	msg = strings.Join(strings.Fields(msg), " ")
	if len(msg) > 280 {
		msg = msg[:280] + "..."
	}
	if msg == "" {
		msg = "sin detalle"
	}
	return msg
}

func sendEmpresaUsuarioMailuMessage(dbSuper *sql.DB, fromEmail, toEmail string, msg []byte) error {
	if !empresaUsuarioMailuFallbackEnabled(dbSuper) {
		return fmt.Errorf("correo corporativo Mailu no habilitado")
	}

	if err := smtp.SendMail("mailu-smtp:25", nil, fromEmail, []string{toEmail}, msg); err == nil {
		return nil
	}

	candidates := [][]string{
		{"docker", "exec", "-i", "pcs-mailu-smtp", "sendmail", "-t", "-f", fromEmail},
		{"docker", "exec", "-i", "pcs-mailu-smtp", "/usr/sbin/sendmail", "-t", "-f", fromEmail},
		{"docker", "exec", "-i", "pcs-mailu-front", "sendmail", "-t", "-f", fromEmail},
	}
	var lastDetail string
	for _, args := range candidates {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		cmd := exec.CommandContext(ctx, args[0], args[1:]...)
		cmd.Stdin = strings.NewReader(string(msg))
		output, err := cmd.CombinedOutput()
		cancel()
		if err == nil {
			return nil
		}
		lastDetail = sanitizeEmpresaUsuarioMailerError(err, output)
	}
	return fmt.Errorf("Mailu no pudo enviar el correo: %s", lastDetail)
}

func sendEmpresaUsuarioMailuMultipart(dbSuper *sql.DB, baseURL, toEmail, subject, bodyPlain, bodyHTML string) error {
	fromName, fromEmail := empresaUsuarioMailuSender(dbSuper)
	msg := buildEmpresaUsuarioMultipartMessage(baseURL, fromName, fromEmail, toEmail, subject, bodyPlain, bodyHTML)
	return sendEmpresaUsuarioMailuMessage(dbSuper, fromEmail, toEmail, []byte(msg))
}

func sendEmpresaUsuarioMailuPlain(dbSuper *sql.DB, toEmail, subject, body string) error {
	fromName, fromEmail := empresaUsuarioMailuSender(dbSuper)
	msg := buildEmpresaUsuarioPlainMessage(fromName, fromEmail, toEmail, subject, body)
	return sendEmpresaUsuarioMailuMessage(dbSuper, fromEmail, toEmail, []byte(msg))
}

func isEmpresaUsuarioMailActionableConfigError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	return isEmpresaUsuarioMailSecretDecryptError(err) ||
		strings.Contains(msg, "no configurado") ||
		strings.Contains(msg, "smtp") ||
		strings.Contains(msg, "gmail")
}

func captureEmpresaUsuarioMailNotification(
	dbSuper *sql.DB,
	tipo string,
	empresaID int64,
	destinatario string,
	asunto string,
	cuerpo string,
	tokenRef string,
	metadataJSON string,
	usuarioCreador string,
) error {
	if dbSuper == nil {
		return fmt.Errorf("db super no disponible para captura de correo")
	}
	usuarioCreador = strings.TrimSpace(usuarioCreador)
	if usuarioCreador == "" {
		usuarioCreador = "sistema"
	}
	_, err := dbpkg.CreateSuperCorreoNotificacionPrueba(dbSuper, dbpkg.SuperCorreoNotificacionPrueba{
		Tipo:           tipo,
		EmpresaID:      empresaID,
		Destinatario:   destinatario,
		Asunto:         asunto,
		Cuerpo:         cuerpo,
		TokenRef:       tokenRef,
		MetadataJSON:   metadataJSON,
		UsuarioCreador: usuarioCreador,
		Estado:         "capturado",
		Observaciones:  "modo_pruebas_correo",
	})
	return err
}

func resolveBaseURLForConfirmation(r *http.Request, dbSuper *sql.DB) string {
	if dbSuper != nil {
		if configured, err := getDecryptedConfigValue(dbSuper, "gmail.confirm_base_url"); err == nil {
			configured = strings.TrimSpace(configured)
			if configured != "" {
				return strings.TrimRight(configured, "/")
			}
		}
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if xfProto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); xfProto != "" {
		scheme = xfProto
	}
	host := strings.TrimSpace(r.Host)
	if host == "" {
		host = "localhost:8080"
	}
	return scheme + "://" + host
}

func buildEmpresaUsuarioInvitationURL(r *http.Request, dbEmp, dbSuper *sql.DB, empresaID int64, toEmail, token string) string {
	loginURL := resolveEmpresaUsuarioLoginURL(r, dbEmp, dbSuper, empresaID)
	parsed, err := url.Parse(loginURL)
	if err != nil {
		baseURL := strings.TrimRight(resolveBaseURLForConfirmation(r, dbSuper), "/")
		parsed, _ = url.Parse(baseURL + "/login_usuario.html")
	}
	query := parsed.Query()
	query.Set("email", strings.TrimSpace(toEmail))
	query.Set("token_invitacion", strings.TrimSpace(token))
	query.Set("modo", "registro")
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func sendEmpresaUsuarioConfirmationEmail(r *http.Request, dbEmp, dbSuper *sql.DB, empresaID int64, toEmail, toName, token string, adminMessage string) (string, error) {
	baseURL := resolveBaseURLForConfirmation(r, dbSuper)
	confirmURL := buildEmpresaUsuarioInvitationURL(r, dbEmp, dbSuper, empresaID, toEmail, token)
	loginURL := resolveEmpresaUsuarioLoginURL(r, dbEmp, dbSuper, empresaID)

	safeName := strings.TrimSpace(toName)
	if safeName == "" {
		safeName = "usuario"
	}

	// intentar obtener nombre de la empresa para el mensaje
	empresaNombre := "la empresa"
	if dbEmp != nil && empresaID > 0 {
		if cfg, err := dbpkg.GetEmpresaVentaPublicaConfig(dbEmp, empresaID); err == nil {
			if strings.TrimSpace(cfg.NombreTienda) != "" {
				empresaNombre = strings.TrimSpace(cfg.NombreTienda)
			}
		}
	}

	adminEmail := adminEmailFromRequest(r)

	adminMessage = strings.TrimSpace(adminMessage)
	subject, bodyPlain, bodyHTML, err := applySuperEmailTemplate(dbSuper, superEmailTemplateKeyEmpresaConfirmation, map[string]string{
		"name":                     safeName,
		"company_name":             empresaNombre,
		"confirm_url":              confirmURL,
		"login_url":                loginURL,
		"admin_message":            adminMessage,
		"admin_message_block_text": templateParagraphText("Mensaje del administrador:", adminMessage),
		"admin_message_block_html": templateParagraphHTML("Mensaje del administrador:", adminMessage),
	})
	if err != nil {
		return confirmURL, err
	}

	if isEmpresaUsuarioMailTestMode(dbSuper) {
		metadataJSON := fmt.Sprintf(`{"confirm_url":%q,"login_url":%q,"mail_mode":"test","admin_message":%q,"admin_email":%q}`, confirmURL, loginURL, adminMessage, adminEmail)
		if err := captureEmpresaUsuarioMailNotification(
			dbSuper,
			dbpkg.SuperCorreoNotificacionTipoConfirmacion,
			empresaID,
			toEmail,
			subject,
			bodyPlain,
			token,
			metadataJSON,
			adminEmail,
		); err != nil {
			return confirmURL, err
		}
		return confirmURL, nil
	}

	if err := sendEmpresaUsuarioGmailMultipart(dbSuper, baseURL, toEmail, subject, bodyPlain, bodyHTML); err != nil {
		gmailErr := err
		if fallbackErr := sendEmpresaUsuarioMailuMultipart(dbSuper, baseURL, toEmail, subject, bodyPlain, bodyHTML); fallbackErr == nil {
			log.Printf("[usuarios_empresa] confirmacion enviada por Mailu interno tras fallo de Gmail empresa_id=%d", empresaID)
			return confirmURL, nil
		}
		return confirmURL, gmailErr
	}
	return confirmURL, nil
}

func sendEmpresaUsuarioPasswordRecoveryEmail(r *http.Request, dbEmp, dbSuper *sql.DB, empresaID int64, toEmail, toName, token string) (string, error) {
	resetURL, err := url.Parse(resolveEmpresaUsuarioLoginURL(r, dbEmp, dbSuper, empresaID))
	if err != nil {
		return "", err
	}
	query := resetURL.Query()
	query.Set("email", toEmail)
	query.Set("token_recuperacion", token)
	resetURL.RawQuery = query.Encode()
	resetHintURL := resetURL.String()

	safeName := strings.TrimSpace(toName)
	if safeName == "" {
		safeName = "usuario"
	}

	subject, body, _, err := applySuperEmailTemplate(dbSuper, superEmailTemplateKeyEmpresaPasswordRecovery, map[string]string{
		"name":      safeName,
		"token":     token,
		"reset_url": resetHintURL,
	})
	if err != nil {
		return resetHintURL, err
	}

	if isEmpresaUsuarioMailTestMode(dbSuper) {
		metadataJSON := fmt.Sprintf(`{"reset_hint_url":%q,"mail_mode":"test"}`, resetHintURL)
		if err := captureEmpresaUsuarioMailNotification(
			dbSuper,
			dbpkg.SuperCorreoNotificacionTipoRecuperacion,
			empresaID,
			toEmail,
			subject,
			body,
			token,
			metadataJSON,
			adminEmailFromRequest(r),
		); err != nil {
			return resetHintURL, err
		}
		return resetHintURL, nil
	}

	if err := sendEmpresaUsuarioGmailPlain(dbSuper, toEmail, subject, body); err != nil {
		gmailErr := err
		if fallbackErr := sendEmpresaUsuarioMailuPlain(dbSuper, toEmail, subject, body); fallbackErr == nil {
			log.Printf("[usuarios_empresa] recuperacion enviada por Mailu interno tras fallo de Gmail empresa_id=%d", empresaID)
			return resetHintURL, nil
		}
		return resetHintURL, gmailErr
	}
	return resetHintURL, nil
}

type empresaUsuarioSessionResult struct {
	EmpresaID   int64
	UsuarioID   int64
	Rol         string
	RolNombre   string
	Email       string
	RedirectURL string
	Apariencia  string
}

func createEmpresaUsuarioSession(w http.ResponseWriter, r *http.Request, dbSuper *sql.DB, item *dbpkg.EmpresaUsuario) (empresaUsuarioSessionResult, error) {
	var result empresaUsuarioSessionResult
	if item == nil {
		return result, fmt.Errorf("usuario de empresa requerido")
	}
	sessionRole := normalizePermissionRole(item.RolNombre)
	if sessionRole == "" || sessionRole == "sin_rol" {
		sessionRole = "admin_empresa"
	}
	if err := dbpkg.UpsertAdministrador(dbSuper, item.Email, item.Nombre, sessionRole, ""); err != nil {
		return result, fmt.Errorf("failed to upsert admin: %w", err)
	}
	if _, err := dbpkg.UpsertAdminEmpresaCompartidaAcceso(dbSuper, dbpkg.AdminEmpresaCompartidaAcceso{
		EmpresaID:          item.EmpresaID,
		AdminEmail:         item.Email,
		CompartidoPorEmail: item.UsuarioCreador,
		FechaAceptada:      time.Now().Format("2006-01-02 15:04:05"),
		UsuarioCreador:     item.UsuarioCreador,
		Estado:             "activo",
		Observaciones:      "Acceso operativo para usuario de empresa.",
	}); err != nil {
		return result, fmt.Errorf("failed to upsert company access: %w", err)
	}

	token, err := utils.GenerateSecureToken(32)
	if err != nil {
		return result, fmt.Errorf("failed to generate session token: %w", err)
	}
	if err := dbpkg.CreateSession(dbSuper, item.Email, r.RemoteAddr, r.UserAgent(), token); err != nil {
		return result, fmt.Errorf("failed to create session: %w", err)
	}

	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400,
		Secure:   SessionCookieSecure(r),
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
	SetBrowserSessionStateCookie(w, r, true)

	redirectURL := "/administrar_empresa.html?id=" + strconv.FormatInt(item.EmpresaID, 10)
	apariencia, appearanceErr := dbpkg.GetUsuarioApariencia(dbSuper, item.Email)
	if appearanceErr != nil {
		log.Println("createEmpresaUsuarioSessionAndRespond get appearance error:", appearanceErr)
		apariencia = ""
	}
	result = empresaUsuarioSessionResult{
		EmpresaID:   item.EmpresaID,
		UsuarioID:   item.ID,
		Rol:         sessionRole,
		RolNombre:   item.RolNombre,
		Email:       item.Email,
		RedirectURL: redirectURL,
		Apariencia:  apariencia,
	}
	return result, nil
}

func createEmpresaUsuarioSessionAndRespond(w http.ResponseWriter, r *http.Request, dbSuper *sql.DB, item *dbpkg.EmpresaUsuario) error {
	result, err := createEmpresaUsuarioSession(w, r, dbSuper, item)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":           true,
		"empresa_id":   result.EmpresaID,
		"usuario_id":   result.UsuarioID,
		"rol":          result.Rol,
		"rol_nombre":   result.RolNombre,
		"email":        result.Email,
		"redirect_url": result.RedirectURL,
		"apariencia":   result.Apariencia,
	})
	return nil
}

func hashEmpresaUsuarioPassword(password, salt string) string {
	sum := sha256.Sum256([]byte(salt + ":" + password))
	return hex.EncodeToString(sum[:])
}

func generateEmpresaUsuarioPasswordHash(password string) (string, string, error) {
	salt, err := utils.GenerateSecureToken(16)
	if err != nil {
		return "", "", err
	}
	return hashEmpresaUsuarioPassword(password, salt), salt, nil
}

func verifyEmpresaUsuarioPassword(password string, item *dbpkg.EmpresaUsuario) bool {
	if item == nil {
		return false
	}
	if strings.TrimSpace(item.PasswordHash) == "" || strings.TrimSpace(item.PasswordSalt) == "" {
		return false
	}
	return hashEmpresaUsuarioPassword(password, item.PasswordSalt) == strings.TrimSpace(item.PasswordHash)
}

func warmEmpresaPermissionSnapshot(dbEmp, dbSuper *sql.DB, item *dbpkg.EmpresaUsuario) {
	if dbEmp == nil || dbSuper == nil || item == nil {
		return
	}
	adminEmail := strings.ToLower(strings.TrimSpace(item.Email))
	empresaID := item.EmpresaID
	roleName := normalizePermissionRole(item.RolNombre)
	if adminEmail == "" || empresaID <= 0 {
		return
	}
	go func() {
		if roleName != "" && roleName != "sin_rol" {
			_, _, _ = loadPermissionOverridesByRoleName(dbSuper, roleName)
			_ = buildPermissionModuleMatrixForRoleDynamic(dbSuper, roleName)
		}
		_, _, _ = loadEmpresaPermissionOverrides(dbSuper, empresaID)
		if _, err := dbpkg.GetLicenciaPermisoPolicyByEmpresa(dbSuper, empresaID); err != nil {
			log.Printf("[usuarios_empresa] warm licencia policy empresa_id=%d email=%s error=%v", empresaID, adminEmail, err)
		}
		if _, err := dbpkg.CanAdminAccessEmpresaIA(dbEmp, dbSuper, adminEmail, empresaID); err != nil {
			log.Printf("[usuarios_empresa] warm admin access empresa_id=%d email=%s error=%v", empresaID, adminEmail, err)
		}
	}()
	go func() {
		time.Sleep(2 * time.Second)
		if _, err := getEmpresaPermissionSnapshot(dbEmp, dbSuper, adminEmail, empresaID); err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("[usuarios_empresa] warm permission snapshot empresa_id=%d email=%s error=%v", empresaID, adminEmail, err)
		}
	}()
}

func parseEmpresaUsuarioDateTime(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}
