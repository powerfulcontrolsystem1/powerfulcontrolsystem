package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/mail"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

// AccountHandler devuelve el perfil compacto de la cuenta asociada a la cookie de sesión.
// Combina el registro de administrador (dbSuper) y, si existe, el usuario de empresa (dbEmp).
func AccountHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("session_token")
		if err != nil || c == nil || c.Value == "" {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}

		s, err := dbpkg.GetSessionByToken(dbSuper, c.Value)
		if err != nil || s == nil {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}

		// Intentar obtener admin desde dbSuper (incluye campos extendidos como telefono)
		adminFull, _ := dbpkg.GetAdminByEmailFull(dbSuper, s.AdminEmail)

		// Intentar obtener usuario de empresa (si existe tabla/registro)
		var empresaUser *dbpkg.EmpresaUsuario
		if dbEmp != nil {
			if eu, err := dbpkg.GetEmpresaUsuarioByEmail(dbEmp, s.AdminEmail); err == nil && eu != nil {
				empresaUser = eu
			}
		}

		// Payload de respuesta
		type Account struct {
			Email       string                `json:"email"`
			Name        string                `json:"name"`
			Photo       string                `json:"photo,omitempty"`
			Role        string                `json:"role,omitempty"`
			IsSuper     bool                  `json:"is_super,omitempty"`
			Admin       *dbpkg.Admin          `json:"admin,omitempty"`
			EmpresaUser *dbpkg.EmpresaUsuario `json:"empresa_user,omitempty"`
		}

		var out Account
		if adminFull != nil {
			out.Email = adminFull.Email
			out.Name = adminFull.Name
			out.Photo = adminFull.Photo
			out.Role = utils.ManagedAdminRole(adminFull.Email, adminFull.Role)
			out.IsSuper = utils.IsSuperPanelRole(out.Role)
			out.Admin = adminFull
			out.Admin.Role = out.Role
		} else {
			out.Email = s.AdminEmail
		}
		if empresaUser != nil {
			out.EmpresaUser = empresaUser
		}

		w.Header().Set("Content-Type", "application/json")
		encodeJSONResponse(w, out)
	}
}

// AccountUpdateProfileHandler permite al usuario autenticado actualizar su perfil (email, nombre, telefono).
func AccountUpdateProfileHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		c, err := r.Cookie("session_token")
		if err != nil || c == nil || c.Value == "" {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		s, err := dbpkg.GetSessionByToken(dbSuper, c.Value)
		if err != nil || s == nil {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		admin, err := dbpkg.GetAdminByEmailFull(dbSuper, s.AdminEmail)
		if err != nil || admin == nil {
			http.Error(w, "account not found", http.StatusNotFound)
			return
		}
		var payload struct {
			Email    string `json:"email"`
			Name     string `json:"name"`
			Telefono string `json:"telefono"`
			Pais     string `json:"pais"`
			Ciudad   string `json:"ciudad"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		payload.Email = strings.TrimSpace(payload.Email)
		payload.Name = strings.TrimSpace(payload.Name)
		payload.Telefono = strings.TrimSpace(payload.Telefono)
		payload.Pais = strings.TrimSpace(payload.Pais)
		payload.Ciudad = strings.TrimSpace(payload.Ciudad)

		if payload.Email != "" {
			if _, err := mail.ParseAddress(payload.Email); err != nil {
				http.Error(w, "invalid email", http.StatusBadRequest)
				return
			}
		}

		newEmail := admin.Email
		if payload.Email != "" {
			newEmail = payload.Email
		}

		// Actualizar perfil (email, telefono, name)
		if err := dbpkg.UpdateAdministradorProfile(dbSuper, admin.ID, payload.Name, payload.Telefono, newEmail, payload.Pais, payload.Ciudad); err != nil {
			log.Println("AccountUpdateProfileHandler update error:", err)
			http.Error(w, "failed to update profile", http.StatusInternalServerError)
			return
		}

		// A change of login identifier is a security event. Existing sessions
		// must not silently inherit the new identity.
		if strings.ToLower(strings.TrimSpace(newEmail)) != strings.ToLower(strings.TrimSpace(admin.Email)) {
			if err := dbpkg.RevokeSessionsByAdminEmail(dbSuper, admin.Email); err != nil {
				log.Println("AccountUpdateProfileHandler revoke previous sessions error:", err)
				http.Error(w, "failed to protect sessions", http.StatusInternalServerError)
				return
			}
			utils.InvalidateAuthCacheForAdmin(admin.Email)
			if err := issueReplacementAdminSession(w, r, dbSuper, newEmail); err != nil {
				log.Println("AccountUpdateProfileHandler rotate sessions error:", err)
				http.Error(w, "failed to protect sessions", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		encodeJSONResponse(w, map[string]interface{}{"ok": true})
	}
}

// AccountChangePasswordHandler permite cambiar la contraseña del usuario autenticado.
func AccountChangePasswordHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		c, err := r.Cookie("session_token")
		if err != nil || c == nil || c.Value == "" {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		s, err := dbpkg.GetSessionByToken(dbSuper, c.Value)
		if err != nil || s == nil {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		admin, err := dbpkg.GetAdminByEmailFull(dbSuper, s.AdminEmail)
		if err != nil || admin == nil {
			http.Error(w, "account not found", http.StatusNotFound)
			return
		}
		var payload struct {
			CurrentPassword string `json:"current_password"`
			NewPassword     string `json:"new_password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		payload.CurrentPassword = strings.TrimSpace(payload.CurrentPassword)
		payload.NewPassword = strings.TrimSpace(payload.NewPassword)
		if payload.CurrentPassword == "" || payload.NewPassword == "" {
			http.Error(w, "current and new password required", http.StatusBadRequest)
			return
		}
		// verificar contraseña actual
		expected := hashEmpresaUsuarioPassword(payload.CurrentPassword, admin.PasswordSalt)
		if expected != strings.TrimSpace(admin.PasswordHash) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		// generar hash y guardar
		hash, salt, herr := generateEmpresaUsuarioPasswordHash(payload.NewPassword)
		if herr != nil {
			log.Println("AccountChangePasswordHandler hash error:", herr)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if err := dbpkg.SetAdministradorPassword(dbSuper, admin.Email, hash, salt); err != nil {
			log.Println("AccountChangePasswordHandler set password error:", err)
			http.Error(w, "failed to update password", http.StatusInternalServerError)
			return
		}
		if err := issueReplacementAdminSession(w, r, dbSuper, admin.Email); err != nil {
			log.Println("AccountChangePasswordHandler rotate sessions error:", err)
			http.Error(w, "failed to protect sessions", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		encodeJSONResponse(w, map[string]interface{}{"ok": true})
	}
}

// AccountSetGooglePasswordHandler permite definir la primera contraseña local para una cuenta autenticada por Google.
func AccountSetGooglePasswordHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		c, err := r.Cookie("session_token")
		if err != nil || c == nil || c.Value == "" {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		s, err := dbpkg.GetSessionByToken(dbSuper, c.Value)
		if err != nil || s == nil {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		admin, err := dbpkg.GetAdminByEmailFull(dbSuper, s.AdminEmail)
		if err != nil || admin == nil {
			http.Error(w, "account not found", http.StatusNotFound)
			return
		}
		if admin.PasswordSet == 1 && strings.TrimSpace(admin.PasswordHash) != "" {
			w.Header().Set("Content-Type", "application/json")
			encodeJSONResponse(w, map[string]interface{}{"ok": true, "redirect_url": resolveAdminPostLoginRedirect(admin), "message": "Tu cuenta ya tiene una contraseña activa."})
			return
		}

		var payload struct {
			Password        string `json:"password"`
			PasswordConfirm string `json:"password_confirm"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		payload.Password = strings.TrimSpace(payload.Password)
		payload.PasswordConfirm = strings.TrimSpace(payload.PasswordConfirm)
		if payload.Password == "" || payload.PasswordConfirm == "" {
			http.Error(w, "password and confirmation required", http.StatusBadRequest)
			return
		}
		if len(payload.Password) < minAdminPasswordLength {
			http.Error(w, "password must be at least 8 characters", http.StatusBadRequest)
			return
		}
		if payload.Password != payload.PasswordConfirm {
			http.Error(w, "passwords do not match", http.StatusBadRequest)
			return
		}

		hash, salt, herr := generateEmpresaUsuarioPasswordHash(payload.Password)
		if herr != nil {
			log.Println("AccountSetGooglePasswordHandler hash error:", herr)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if err := dbpkg.SetAdministradorPassword(dbSuper, admin.Email, hash, salt); err != nil {
			log.Println("AccountSetGooglePasswordHandler set password error:", err)
			http.Error(w, "failed to update password", http.StatusInternalServerError)
			return
		}
		if err := issueReplacementAdminSession(w, r, dbSuper, admin.Email); err != nil {
			log.Println("AccountSetGooglePasswordHandler rotate sessions error:", err)
			http.Error(w, "failed to protect sessions", http.StatusInternalServerError)
			return
		}

		updatedAdmin, _ := dbpkg.GetAdminByEmailFull(dbSuper, admin.Email)
		redirectURL := resolveAdminPostLoginRedirect(updatedAdmin)
		if redirectURL == googlePasswordSetupPagePath {
			if updatedAdmin != nil && strings.EqualFold(strings.TrimSpace(updatedAdmin.Role), "super_administrador") {
				redirectURL = "/super_administrador.html"
			} else {
				redirectURL = "/seleccionar_empresa.html"
			}
		}

		w.Header().Set("Content-Type", "application/json")
		encodeJSONResponse(w, map[string]interface{}{"ok": true, "redirect_url": redirectURL, "message": "Contraseña registrada correctamente."})
	}
}
