package handlers

import (
    "database/sql"
    "encoding/json"
    "log"
    "net/mail"
    "net/http"
    "strings"

    dbpkg "github.com/you/pos-backend/db"
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
            Email       string                 `json:"email"`
            Name        string                 `json:"name"`
            Photo       string                 `json:"photo,omitempty"`
            Role        string                 `json:"role,omitempty"`
            IsSuper     bool                   `json:"is_super,omitempty"`
            Admin       *dbpkg.Admin           `json:"admin,omitempty"`
            EmpresaUser *dbpkg.EmpresaUsuario `json:"empresa_user,omitempty"`
        }

        var out Account
        if adminFull != nil {
            out.Email = adminFull.Email
            out.Name = adminFull.Name
            out.Photo = adminFull.Photo
            out.Role = adminFull.Role
            out.IsSuper = strings.Contains(strings.ToLower(adminFull.Role), "super")
            out.Admin = adminFull
        } else {
            out.Email = s.AdminEmail
        }
        if empresaUser != nil {
            out.EmpresaUser = empresaUser
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(out)
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
        var payload struct{
            Email string `json:"email"`
            Name string `json:"name"`
            Telefono string `json:"telefono"`
        }
        if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
            http.Error(w, "invalid payload", http.StatusBadRequest)
            return
        }
        payload.Email = strings.TrimSpace(payload.Email)
        payload.Name = strings.TrimSpace(payload.Name)
        payload.Telefono = strings.TrimSpace(payload.Telefono)

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
        if err := dbpkg.UpdateAdministradorProfile(dbSuper, admin.ID, payload.Name, payload.Telefono, newEmail); err != nil {
            log.Println("AccountUpdateProfileHandler update error:", err)
            http.Error(w, "failed to update profile", http.StatusInternalServerError)
            return
        }

        // Si cambió el email, reasignar sesiones activas
        if strings.ToLower(strings.TrimSpace(newEmail)) != strings.ToLower(strings.TrimSpace(admin.Email)) {
            if err := dbpkg.ReassignSessionsAdminEmail(dbSuper, admin.Email, newEmail); err != nil {
                log.Println("AccountUpdateProfileHandler reassign sessions error:", err)
            }
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
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
        var payload struct{
            CurrentPassword string `json:"current_password"`
            NewPassword string `json:"new_password"`
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
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
    }
}
