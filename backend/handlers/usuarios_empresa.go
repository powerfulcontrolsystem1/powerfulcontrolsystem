package handlers

import (
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

const (
	empresaUsuarioMaxIntentosFallidos = 5
	empresaUsuarioVentanaIntentos     = 15 * time.Minute
	empresaUsuarioBloqueoDuracion     = 15 * time.Minute
	empresaUsuarioRecuperacionTTL     = 30 * time.Minute
)

// EmpresaRolesDeUsuarioHandler devuelve los roles disponibles para la empresa seleccionada.
func EmpresaRolesDeUsuarioHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		tipoEmpresaID, _, err := resolveTipoEmpresaIDForEmpresa(dbEmp, dbSuper, empresaID)
		if err != nil {
			// Si aún no hay relación tipo->empresa, devolvemos vacío para no romper la UI.
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]dbpkg.RolDeUsuario{})
			return
		}

		includeInactive := r.URL.Query().Get("include_inactive") == "1"
		roles, err := dbpkg.GetRolesDeUsuario(dbSuper, tipoEmpresaID, includeInactive)
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
			var payload struct {
				EmpresaID          int64  `json:"empresa_id"`
				Email              string `json:"email"`
				Nombre             string `json:"nombre"`
				DocumentoIdentidad string `json:"documento_identidad"`
				RolUsuarioID       int64  `json:"rol_usuario_id"`
				Observaciones      string `json:"observaciones"`
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
				rolNombre,
				strings.TrimSpace(payload.Observaciones),
				adminEmailFromRequest(r),
				token,
				expira,
			)
			if err != nil {
				if strings.Contains(strings.ToLower(err.Error()), "unique") {
					http.Error(w, "ya existe un usuario con ese correo", http.StatusConflict)
					return
				}
				http.Error(w, "failed to create user: "+err.Error(), http.StatusInternalServerError)
				return
			}

			confirmURL, mailErr := sendEmpresaUsuarioConfirmationEmail(r, dbSuper, payload.EmpresaID, strings.TrimSpace(payload.Email), strings.TrimSpace(payload.Nombre), token)
			if mailErr != nil {
				// Regla de negocio: si no se envía correo, no se registra usuario.
				rollbackErr := dbpkg.DeleteEmpresaUsuario(dbEmp, payload.EmpresaID, id)
				if rollbackErr != nil {
					http.Error(w, "no se pudo enviar el correo de validación y tampoco revertir el usuario: "+rollbackErr.Error(), http.StatusInternalServerError)
					return
				}
				http.Error(w, "no se pudo enviar el correo de validación; el usuario no fue registrado: "+mailErr.Error()+" | enlace: "+confirmURL, http.StatusBadGateway)
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

				confirmURL, mailErr := sendEmpresaUsuarioConfirmationEmail(r, dbSuper, empresaID, item.Email, item.Nombre, token)
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
				Observaciones      string `json:"observaciones"`
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
				rolNombre,
				strings.TrimSpace(payload.Observaciones),
				resetConfirm,
				confirmToken,
				confirmExpira,
			); err != nil {
				if strings.Contains(strings.ToLower(err.Error()), "unique") {
					http.Error(w, "ya existe un usuario con ese correo", http.StatusConflict)
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
				confirmURL, mailErr := sendEmpresaUsuarioConfirmationEmail(r, dbSuper, payload.EmpresaID, strings.TrimSpace(payload.Email), strings.TrimSpace(payload.Nombre), confirmToken)
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
			EmpresaID int64  `json:"empresa_id"`
			Email     string `json:"email"`
			Password  string `json:"password"`
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

		item, err := dbpkg.GetEmpresaUsuarioByEmailScoped(dbEmp, email, payload.EmpresaID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "credenciales inválidas", http.StatusUnauthorized)
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
		if strings.EqualFold(strings.TrimSpace(item.Estado), "inactivo") {
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
				"email":                   item.Email,
				"message":                 "Primer ingreso: debes crear tu contraseña para continuar.",
			})
			return
		}

		if strings.TrimSpace(payload.Password) == "" {
			http.Error(w, "password es obligatorio", http.StatusBadRequest)
			return
		}
		if !verifyEmpresaUsuarioPassword(payload.Password, item) {
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

		if err := createEmpresaUsuarioSessionAndRespond(w, r, dbSuper, item); err != nil {
			log.Printf("[usuarios_empresa] failed to create session (login) empresa_id=%d email=%s error=%v", item.EmpresaID, item.Email, err)
			http.Error(w, "No se pudo iniciar sesión del usuario", http.StatusInternalServerError)
			return
		}
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
		if len(payload.Password) < 8 {
			http.Error(w, "la contraseña debe tener al menos 8 caracteres", http.StatusBadRequest)
			return
		}
		if payload.PasswordConfirm != "" && payload.Password != payload.PasswordConfirm {
			http.Error(w, "la confirmación de contraseña no coincide", http.StatusBadRequest)
			return
		}

		item, err := dbpkg.GetEmpresaUsuarioByEmailScoped(dbEmp, email, payload.EmpresaID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "usuario no encontrado", http.StatusNotFound)
				return
			}
			log.Printf("[usuarios_empresa] failed to query user (set_password) empresa_id=%d email=%s error=%v", payload.EmpresaID, email, err)
			http.Error(w, "No se pudo validar el usuario", http.StatusInternalServerError)
			return
		}
		if !strings.EqualFold(strings.TrimSpace(item.DocumentoIdentidad), documento) {
			http.Error(w, "documento inválido", http.StatusUnauthorized)
			return
		}
		if item.EmailConfirmado != 1 {
			http.Error(w, "debes confirmar tu correo antes de crear contraseña", http.StatusForbidden)
			return
		}
		if strings.EqualFold(strings.TrimSpace(item.Estado), "inactivo") {
			http.Error(w, "tu usuario está inactivo", http.StatusForbidden)
			return
		}
		if item.PasswordSet == 1 && strings.TrimSpace(item.PasswordHash) != "" {
			http.Error(w, "el usuario ya tiene contraseña configurada", http.StatusConflict)
			return
		}

		hash, salt, err := generateEmpresaUsuarioPasswordHash(payload.Password)
		if err != nil {
			http.Error(w, "no se pudo generar password hash", http.StatusInternalServerError)
			return
		}
		if err := dbpkg.SetEmpresaUsuarioPassword(dbEmp, item.EmpresaID, item.ID, hash, salt); err != nil {
			log.Printf("[usuarios_empresa] failed to set password empresa_id=%d id=%d email=%s error=%v", item.EmpresaID, item.ID, item.Email, err)
			http.Error(w, "No se pudo actualizar la contraseña", http.StatusInternalServerError)
			return
		}

		item.PasswordHash = hash
		item.PasswordSalt = salt
		item.PasswordSet = 1

		if err := createEmpresaUsuarioSessionAndRespond(w, r, dbSuper, item); err != nil {
			log.Printf("[usuarios_empresa] failed to create session (set_password) empresa_id=%d email=%s error=%v", item.EmpresaID, item.Email, err)
			http.Error(w, "No se pudo iniciar sesión del usuario", http.StatusInternalServerError)
			return
		}
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
			EmpresaID int64  `json:"empresa_id"`
			Email     string `json:"email"`
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

		respondAccepted := func(delivery string) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":       true,
				"delivery": delivery,
				"message":  "Si el correo existe, enviaremos instrucciones para recuperar la contraseña.",
			})
		}

		item, err := dbpkg.GetEmpresaUsuarioByEmailScoped(dbEmp, email, payload.EmpresaID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
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

		if _, mailErr := sendEmpresaUsuarioPasswordRecoveryEmail(r, dbSuper, item.EmpresaID, item.Email, item.Nombre, token); mailErr != nil {
			log.Printf("[usuarios_empresa] password recovery email not sent empresa_id=%d id=%d email=%s error=%v", item.EmpresaID, item.ID, item.Email, mailErr)
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
			http.Error(w, "empresa_id, email y token son obligatorios", http.StatusBadRequest)
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
		if len(payload.Password) < 8 {
			http.Error(w, "la contraseña debe tener al menos 8 caracteres", http.StatusBadRequest)
			return
		}
		if payload.PasswordConfirm != "" && payload.PasswordConfirm != payload.Password {
			http.Error(w, "la confirmación de contraseña no coincide", http.StatusBadRequest)
			return
		}

		item, err := dbpkg.GetEmpresaUsuarioByEmailScoped(dbEmp, email, payload.EmpresaID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "token de recuperación inválido", http.StatusUnauthorized)
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
		empresaID, err := dbpkg.ConfirmEmpresaUsuarioByToken(dbEmp, token)
		if err != nil {
			if qEmpresaID, qErr := parseInt64QueryOptional(r, "empresa_id"); qErr == nil && qEmpresaID > 0 {
				empresaID = qEmpresaID
			}
			loginURL := "/login_usuario.html"
			if empresaID > 0 {
				loginURL += "?empresa_id=" + strconv.FormatInt(empresaID, 10)
			}
			msg := html.EscapeString(err.Error())
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "<html><body style='font-family:sans-serif;background:#10141f;color:#e9eefb;padding:24px'><h2>No se pudo confirmar el correo</h2><p>%s</p><p><a href='%s' style='color:#7fb2ff'>Volver al login de usuario</a></p></body></html>", msg, html.EscapeString(loginURL))
			return
		}
		loginURL := "/login_usuario.html"
		if empresaID > 0 {
			loginURL += "?empresa_id=" + strconv.FormatInt(empresaID, 10)
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, "<html><body style='font-family:sans-serif;background:#10141f;color:#e9eefb;padding:24px'><h2>Correo confirmado correctamente</h2><p>Tu cuenta ya está confirmada.</p><p><a href='%s' style='color:#7fb2ff'>Ir al login de usuario</a></p></body></html>", html.EscapeString(loginURL))
	}
}

// GmailConfigHandler gestiona configuración de envío SMTP por Gmail.
func GmailConfigHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			smtpEmail, _, _, smtpEmailUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "gmail.smtp_email")
			appPass, appPassEnc, _, appPassUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "gmail.smtp_app_password")
			fromName, _, _, fromNameUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "gmail.smtp_from_name")
			host, _, _, hostUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "gmail.smtp_host")
			port, _, _, portUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "gmail.smtp_port")
			baseURL, _, _, baseURLUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "gmail.confirm_base_url")

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
				if appPassEnc {
					masked = "********"
				} else if len(appPass) > 8 {
					masked = appPass[:2] + "****" + appPass[len(appPass)-2:]
				} else {
					masked = "****"
				}
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"smtp_email_set":            strings.TrimSpace(smtpEmail) != "",
				"smtp_email":                smtpEmail,
				"smtp_email_updated":        smtpEmailUpdated,
				"smtp_app_password_set":     strings.TrimSpace(appPass) != "",
				"smtp_app_password_masked":  masked,
				"smtp_app_password_updated": appPassUpdated,
				"smtp_from_name":            fromName,
				"smtp_from_name_updated":    fromNameUpdated,
				"smtp_host":                 host,
				"smtp_host_updated":         hostUpdated,
				"smtp_port":                 port,
				"smtp_port_updated":         portUpdated,
				"confirm_base_url":          baseURL,
				"confirm_base_url_updated":  baseURLUpdated,
				"encryption_available":      utils.EncryptionAvailable(),
			})
			return

		case http.MethodPost, http.MethodPut:
			var payload struct {
				SMTPEmail      string `json:"smtp_email"`
				SMTPAppPass    string `json:"smtp_app_password"`
				SMTPFromName   string `json:"smtp_from_name"`
				SMTPHost       string `json:"smtp_host"`
				SMTPPort       string `json:"smtp_port"`
				ConfirmBaseURL string `json:"confirm_base_url"`
				Encrypt        bool   `json:"encrypt"`
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

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"saved": true})
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
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
	tipoEmpresaID, _, err := resolveTipoEmpresaIDForEmpresa(dbEmp, dbSuper, empresaID)
	if err != nil {
		return "", err
	}

	row := dbSuper.QueryRow(`SELECT nombre, COALESCE(estado, 'activo') FROM roles_de_usuario WHERE id = ? AND tipo_empresa_id = ? LIMIT 1`, rolID, tipoEmpresaID)
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

func newPasswordRecoveryTokenAndExpiration() (string, string, error) {
	token, err := utils.GenerateSecureToken(32)
	if err != nil {
		return "", "", err
	}
	expira := time.Now().Add(empresaUsuarioRecuperacionTTL).Format("2006-01-02 15:04:05")
	return token, expira, nil
}

func resolveBaseURLForConfirmation(r *http.Request, dbSuper *sql.DB) string {
	if configured, err := getDecryptedConfigValue(dbSuper, "gmail.confirm_base_url"); err == nil {
		configured = strings.TrimSpace(configured)
		if configured != "" {
			return strings.TrimRight(configured, "/")
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

func sendEmpresaUsuarioConfirmationEmail(r *http.Request, dbSuper *sql.DB, empresaID int64, toEmail, toName, token string) (string, error) {
	smtpEmail, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_email")
	if err != nil {
		return "", err
	}
	smtpEmail = strings.TrimSpace(smtpEmail)
	if smtpEmail == "" {
		return "", fmt.Errorf("gmail.smtp_email no configurado")
	}

	smtpPass, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_app_password")
	if err != nil {
		return "", err
	}
	smtpPass = strings.TrimSpace(smtpPass)
	if smtpPass == "" {
		return "", fmt.Errorf("gmail.smtp_app_password no configurado")
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

	baseURL := resolveBaseURLForConfirmation(r, dbSuper)
	confirmURL := strings.TrimRight(baseURL, "/") + "/auth/confirmar_correo?token=" + url.QueryEscape(token)
	if empresaID > 0 {
		confirmURL += "&empresa_id=" + strconv.FormatInt(empresaID, 10)
	}
	loginURL := strings.TrimRight(baseURL, "/") + "/login_usuario.html"
	if empresaID > 0 {
		loginURL += "?empresa_id=" + strconv.FormatInt(empresaID, 10)
	}

	mailHostForAuth := smtpHost
	if strings.Contains(smtpHost, ":") {
		if h, _, err := net.SplitHostPort(smtpHost); err == nil {
			mailHostForAuth = h
		}
	}
	addr := smtpHost
	if !strings.Contains(addr, ":") {
		addr = smtpHost + ":" + smtpPort
	}

	auth := smtp.PlainAuth("", smtpEmail, smtpPass, mailHostForAuth)
	safeName := strings.TrimSpace(toName)
	if safeName == "" {
		safeName = "usuario"
	}

	subject := "Confirma tu correo - Powerful Control System"
	body := "Hola " + safeName + ",\r\n\r\n" +
		"Tu cuenta fue creada y necesita confirmar el correo para quedar habilitada.\r\n" +
		"Haz clic en este enlace:\r\n" +
		confirmURL + "\r\n\r\n" +
		"Después de confirmar, inicia sesión aquí:\r\n" +
		loginURL + "\r\n\r\n" +
		"Si no solicitaste esta cuenta, ignora este mensaje.\r\n"

	msg := "From: " + fromName + " <" + smtpEmail + ">\r\n" +
		"To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body

	if err := smtp.SendMail(addr, auth, smtpEmail, []string{toEmail}, []byte(msg)); err != nil {
		return confirmURL, err
	}
	return confirmURL, nil
}

func sendEmpresaUsuarioPasswordRecoveryEmail(r *http.Request, dbSuper *sql.DB, empresaID int64, toEmail, toName, token string) (string, error) {
	smtpEmail, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_email")
	if err != nil {
		return "", err
	}
	smtpEmail = strings.TrimSpace(smtpEmail)
	if smtpEmail == "" {
		return "", fmt.Errorf("gmail.smtp_email no configurado")
	}

	smtpPass, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_app_password")
	if err != nil {
		return "", err
	}
	smtpPass = strings.TrimSpace(smtpPass)
	if smtpPass == "" {
		return "", fmt.Errorf("gmail.smtp_app_password no configurado")
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

	baseURL := resolveBaseURLForConfirmation(r, dbSuper)
	resetHintURL := strings.TrimRight(baseURL, "/") + "/login_usuario.html"
	sep := "?"
	if empresaID > 0 {
		resetHintURL += "?empresa_id=" + strconv.FormatInt(empresaID, 10)
		sep = "&"
	}
	resetHintURL += sep + "email=" + url.QueryEscape(toEmail) + "&token_recuperacion=" + url.QueryEscape(token)

	mailHostForAuth := smtpHost
	if strings.Contains(smtpHost, ":") {
		if h, _, err := net.SplitHostPort(smtpHost); err == nil {
			mailHostForAuth = h
		}
	}
	addr := smtpHost
	if !strings.Contains(addr, ":") {
		addr = smtpHost + ":" + smtpPort
	}

	auth := smtp.PlainAuth("", smtpEmail, smtpPass, mailHostForAuth)
	safeName := strings.TrimSpace(toName)
	if safeName == "" {
		safeName = "usuario"
	}

	subject := "Recuperacion de contraseña - Powerful Control System"
	body := "Hola " + safeName + ",\r\n\r\n" +
		"Recibimos una solicitud para restablecer tu contraseña.\r\n" +
		"Token de recuperación (vigencia limitada):\r\n" +
		token + "\r\n\r\n" +
		"Abre el login de usuario y usa el token para completar el restablecimiento:\r\n" +
		resetHintURL + "\r\n\r\n" +
		"Si no solicitaste este cambio, ignora este mensaje.\r\n"

	msg := "From: " + fromName + " <" + smtpEmail + ">\r\n" +
		"To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body

	if err := smtp.SendMail(addr, auth, smtpEmail, []string{toEmail}, []byte(msg)); err != nil {
		return resetHintURL, err
	}
	return resetHintURL, nil
}

func createEmpresaUsuarioSessionAndRespond(w http.ResponseWriter, r *http.Request, dbSuper *sql.DB, item *dbpkg.EmpresaUsuario) error {
	if err := dbpkg.UpsertAdministrador(dbSuper, item.Email, item.Nombre, "administrador", ""); err != nil {
		return fmt.Errorf("failed to upsert admin: %w", err)
	}

	token, err := utils.GenerateSecureToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate session token: %w", err)
	}
	if err := dbpkg.CreateSession(dbSuper, item.Email, r.RemoteAddr, r.UserAgent(), token); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400,
		Secure:   (r.TLS != nil),
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)

	redirectURL := "/administrar_empresa.html?id=" + strconv.FormatInt(item.EmpresaID, 10)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":           true,
		"empresa_id":   item.EmpresaID,
		"usuario_id":   item.ID,
		"redirect_url": redirectURL,
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
