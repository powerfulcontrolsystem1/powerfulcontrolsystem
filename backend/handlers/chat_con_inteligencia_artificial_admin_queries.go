package handlers

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func buildEmpresaAIAdminDBDirectResponse(dbEmp, dbSuper *sql.DB, empresaID int64, adminEmail, pregunta string) (string, bool, error) {
	folded := foldEmpresaAICommandText(pregunta)
	if !empresaAIWantsUserSummaryQuestion(folded) {
		return "", false, nil
	}
	if dbEmp == nil || empresaID <= 0 {
		return "No pude consultar usuarios porque falta el contexto de empresa.", true, nil
	}
	allowed, role, err := empresaAIAdminRoleCanReadCompanyDB(dbEmp, dbSuper, empresaID, adminEmail)
	if err != nil {
		return "", true, fmt.Errorf("no se pudo validar permisos administrativos para consulta IA")
	}
	if !allowed {
		return "No puedo mostrar conteos administrativos de usuarios con este rol. Esta consulta esta limitada a super administrador, administrador total o administrador de la empresa activa.", true, nil
	}

	users, err := dbpkg.GetEmpresaUsuarios(dbEmp, empresaID, true)
	if err != nil {
		return "", true, fmt.Errorf("no se pudo consultar usuarios de la empresa")
	}
	return formatEmpresaAIUsuariosResumen(empresaID, role, users), true, nil
}

func empresaAIWantsUserSummaryQuestion(folded string) bool {
	folded = strings.TrimSpace(folded)
	if folded == "" {
		return false
	}
	hasUser := strings.Contains(folded, "usuario") ||
		strings.Contains(folded, "usuarios") ||
		strings.Contains(folded, "users") ||
		strings.Contains(folded, "colaborador") ||
		strings.Contains(folded, "empleado")
	hasCount := strings.Contains(folded, "cuantos") ||
		strings.Contains(folded, "cuantas") ||
		strings.Contains(folded, "cantidad") ||
		strings.Contains(folded, "conteo") ||
		strings.Contains(folded, "total") ||
		strings.Contains(folded, "resumen") ||
		strings.Contains(folded, "registrados") ||
		strings.Contains(folded, "registradas")
	return hasUser && hasCount
}

func empresaAIAdminRoleCanReadCompanyDB(dbEmp, dbSuper *sql.DB, empresaID int64, adminEmail string) (bool, string, error) {
	adminEmail = strings.ToLower(strings.TrimSpace(adminEmail))
	if adminEmail == "" || adminEmail == "sistema" || empresaID <= 0 {
		return false, "", nil
	}

	if dbSuper != nil {
		admin, err := dbpkg.GetAdminByEmail(dbSuper, adminEmail)
		if err == nil && admin != nil {
			role := normalizePermissionRole(admin.Role)
			if empresaAIAdministrativeDBReadRole(role) {
				return true, role, nil
			}
		} else if err != nil && err != sql.ErrNoRows {
			return false, "", err
		}
	}

	if dbEmp != nil {
		user, err := dbpkg.GetEmpresaUsuarioByEmailScoped(dbEmp, adminEmail, empresaID)
		if err == nil && user != nil {
			role := normalizePermissionRole(user.RolNombre)
			if empresaAIAdministrativeDBReadRole(role) {
				return true, role, nil
			}
			return false, role, nil
		} else if err != nil && err != sql.ErrNoRows {
			return false, "", err
		}
	}

	return false, "", nil
}

func empresaAIAdministrativeDBReadRole(role string) bool {
	switch normalizePermissionRole(role) {
	case "super_administrador", "administrador_total", "admin_empresa":
		return true
	default:
		return false
	}
}

func formatEmpresaAIUsuariosResumen(empresaID int64, role string, users []dbpkg.EmpresaUsuario) string {
	total := len(users)
	var activos, inactivos, pendientesCorreo, sinPassword int
	roles := map[string]int{}
	for _, user := range users {
		estado := strings.ToLower(strings.TrimSpace(user.Estado))
		if estado == "" || estado == "activo" {
			activos++
		} else {
			inactivos++
		}
		if user.EmailConfirmado == 0 {
			pendientesCorreo++
		}
		if user.PasswordSet == 0 {
			sinPassword++
		}
		roleName := normalizePermissionRole(user.RolNombre)
		if roleName == "" {
			roleName = "sin_rol"
		}
		roles[roleName]++
	}
	roleNames := make([]string, 0, len(roles))
	for roleName := range roles {
		roleNames = append(roleNames, roleName)
	}
	sort.Strings(roleNames)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Consulta real de base de datos ejecutada para la empresa_id %d.\n\n", empresaID))
	b.WriteString(fmt.Sprintf("En esta empresa hay **%d usuarios registrados**.\n\n", total))
	b.WriteString("Resumen:\n")
	b.WriteString(fmt.Sprintf("- Activos: %d\n", activos))
	b.WriteString(fmt.Sprintf("- Inactivos u otros estados: %d\n", inactivos))
	b.WriteString(fmt.Sprintf("- Pendientes de confirmar correo: %d\n", pendientesCorreo))
	b.WriteString(fmt.Sprintf("- Sin contrasena configurada: %d\n", sinPassword))
	if len(roleNames) > 0 {
		b.WriteString("\nPor rol:\n")
		for _, roleName := range roleNames {
			b.WriteString(fmt.Sprintf("- %s: %d\n", roleName, roles[roleName]))
		}
	}
	b.WriteString("\nAlcance aplicado: solo tabla `users` filtrada por `empresa_id`; no consulte otras empresas, secretos, tokens ni claves. Rol autorizado: ")
	b.WriteString(normalizePermissionRole(role))
	b.WriteString(".")
	return b.String()
}
