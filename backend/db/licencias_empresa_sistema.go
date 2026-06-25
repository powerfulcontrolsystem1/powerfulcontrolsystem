package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

const (
	PowerfulSystemEmpresaName             = "Powerful Control System"
	PowerfulSystemEmpresaNameLegacyTypo   = "Powerful Control Systen"
	PowerfulSystemEmpresaConfigKey        = "licencias.facturacion_empresa_sistema_id"
	PowerfulSystemEmpresaLicenseCode      = "PCS_SYSTEM_INTERNAL_PERPETUAL"
	PowerfulSystemEmpresaDefaultLogoURL   = "/img/Logo pcs 1.png"
	powerfulSystemEmpresaUsuarioCreador   = "sistema.powerful_control_system"
	powerfulSystemEmpresaObservacionMarca = "Empresa interna del SaaS para facturar licencias de Powerful Control System."
)

// IsPowerfulSystemEmpresaName reconoce la empresa interna del sistema sin crear duplicados
// por diferencias de mayusculas, espacios o el nombre historico escrito con "Systen".
func IsPowerfulSystemEmpresaName(name string) bool {
	normalized := normalizePowerfulSystemEmpresaName(name)
	switch normalized {
	case "powerful control system", "powerful control systen":
		return true
	default:
		return false
	}
}

// IsPowerfulSystemEmpresa retorna true si el registro corresponde a la empresa interna del SaaS.
func IsPowerfulSystemEmpresa(empresa *Empresa) bool {
	if empresa == nil {
		return false
	}
	if IsPowerfulSystemEmpresaName(empresa.Nombre) {
		return true
	}
	obs := strings.ToLower(strings.TrimSpace(empresa.Observaciones))
	return strings.Contains(obs, "empresa interna del saas") && strings.Contains(obs, "powerful control system")
}

// EnsurePowerfulSystemEmpresa resuelve la empresa emisora interna usada para facturar
// licencias. Primero respeta la configuracion existente, luego busca la empresa ya
// creada por nombre y solo crea una si no existe ninguna coincidencia.
func EnsurePowerfulSystemEmpresa(dbEmp, dbSuper *sql.DB) (*Empresa, error) {
	if dbEmp == nil {
		return nil, fmt.Errorf("base de datos de empresas no disponible")
	}

	if empresa, err := getPowerfulSystemEmpresaFromConfig(dbEmp, dbSuper); err != nil {
		return nil, err
	} else if empresa != nil {
		if err := DisablePowerfulSystemEmpresaInternalLicense(dbSuper, empresa.EmpresaID, powerfulSystemEmpresaUsuarioCreador); err != nil {
			return nil, err
		}
		if err := EnsurePowerfulSystemEmpresaDefaultLogo(dbEmp, empresa.EmpresaID); err != nil {
			return nil, err
		}
		return empresa, nil
	}

	empresa, err := findPowerfulSystemEmpresaByName(dbEmp)
	if err != nil {
		return nil, err
	}
	if empresa == nil {
		id, _, createErr := CreateEmpresaIdempotente(
			dbEmp,
			0,
			"Sistema",
			PowerfulSystemEmpresaName,
			"",
			powerfulSystemEmpresaObservacionMarca,
			powerfulSystemEmpresaUsuarioCreador,
		)
		if createErr != nil {
			return nil, createErr
		}
		empresa, err = GetEmpresaByScopeID(dbEmp, id)
		if err != nil {
			return nil, err
		}
	}
	if empresa == nil {
		return nil, fmt.Errorf("no se pudo resolver la empresa interna Powerful Control System")
	}

	if err := savePowerfulSystemEmpresaConfig(dbSuper, empresa.EmpresaID); err != nil {
		return nil, err
	}
	if err := DisablePowerfulSystemEmpresaInternalLicense(dbSuper, empresa.EmpresaID, powerfulSystemEmpresaUsuarioCreador); err != nil {
		return nil, err
	}
	if err := EnsurePowerfulSystemEmpresaDefaultLogo(dbEmp, empresa.EmpresaID); err != nil {
		return nil, err
	}
	return empresa, nil
}

// EnsurePowerfulSystemEmpresaDefaultLogo mantiene un logo corporativo visible para
// la empresa interna PCS sin pisar logos personalizados subidos a /uploads/.
func EnsurePowerfulSystemEmpresaDefaultLogo(dbEmp *sql.DB, empresaID int64) error {
	if dbEmp == nil || empresaID <= 0 {
		return nil
	}
	cfg, err := GetEmpresaConfiguracionAvanzada(dbEmp, empresaID)
	if err != nil {
		return err
	}
	logoURL := strings.TrimSpace(cfg.LogoURL)
	shouldSetCorporateLogo := logoURL == "" || strings.HasPrefix(logoURL, "/img/")
	if shouldSetCorporateLogo {
		cfg.LogoURL = PowerfulSystemEmpresaDefaultLogoURL
		cfg.MostrarLogoEmpresa = true
	}
	if strings.TrimSpace(cfg.LogoFacturaURL) == "" || strings.HasPrefix(strings.TrimSpace(cfg.LogoFacturaURL), "/img/") {
		cfg.LogoFacturaURL = PowerfulSystemEmpresaDefaultLogoURL
		cfg.MostrarLogoFactura = true
	}
	cfg.MostrarLogo = cfg.MostrarLogoEmpresa || cfg.MostrarLogoFactura || cfg.MostrarLogoSistema
	cfg.UsuarioCreador = powerfulSystemEmpresaUsuarioCreador
	_, err = UpsertEmpresaConfiguracionAvanzada(dbEmp, *cfg)
	return err
}

// DisablePowerfulSystemEmpresaInternalLicense retira la licencia tecnica antigua
// de Powerful Control System para que esta empresa use el mismo ciclo comercial
// de compra, vencimiento y bloqueo que las demas empresas.
func DisablePowerfulSystemEmpresaInternalLicense(dbSuper *sql.DB, empresaID int64, usuario string) error {
	if dbSuper == nil || empresaID <= 0 {
		return nil
	}
	if strings.TrimSpace(usuario) == "" {
		usuario = powerfulSystemEmpresaUsuarioCreador
	}
	if err := EnsureLicenciasSchema(dbSuper); err != nil {
		return err
	}

	nowExpr := sqlNowExpr()
	nota := "Licencia tecnica interna retirada; Powerful Control System debe comprar o renovar licencias como cualquier empresa."
	res, err := execSQLCompat(dbSuper, `UPDATE licencias
		SET activo = 0,
			estado = 'retirada',
			fecha_fin = CASE WHEN COALESCE(fecha_fin, '') = '' THEN `+nowExpr+` ELSE fecha_fin END,
			fecha_actualizacion = `+nowExpr+`,
			usuario_creador = CASE WHEN COALESCE(usuario_creador, '') = '' THEN ? ELSE usuario_creador END,
			observaciones = TRIM(COALESCE(observaciones, '') || CASE WHEN COALESCE(observaciones, '') = '' THEN '' ELSE ' | ' END || ?)
		WHERE empresa_id = ?
			AND UPPER(TRIM(COALESCE(codigo_funcion, ''))) = UPPER(TRIM(?))
			AND COALESCE(activo, 1) = 1`,
		usuario,
		nota,
		empresaID,
		PowerfulSystemEmpresaLicenseCode,
	)
	if err != nil {
		return err
	}
	if res != nil {
		if rows, rowsErr := res.RowsAffected(); rowsErr == nil && rows > 0 {
			InvalidateLicenciaPermisoPolicyCacheForEmpresa(empresaID)
		}
	}
	return nil
}

func getPowerfulSystemEmpresaFromConfig(dbEmp, dbSuper *sql.DB) (*Empresa, error) {
	if dbSuper == nil {
		return nil, nil
	}
	raw, _, _, _, err := GetConfigEntry(dbSuper, PowerfulSystemEmpresaConfigKey)
	if err != nil {
		if isMissingTableError(err) {
			return nil, nil
		}
		return nil, err
	}
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return nil, nil
	}
	empresa, err := GetEmpresaByScopeID(dbEmp, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if empresa == nil {
		return nil, nil
	}
	if !IsPowerfulSystemEmpresa(empresa) {
		return nil, nil
	}
	return empresa, nil
}

func savePowerfulSystemEmpresaConfig(dbSuper *sql.DB, empresaID int64) error {
	if dbSuper == nil || empresaID <= 0 {
		return nil
	}
	if err := SetConfigValue(dbSuper, PowerfulSystemEmpresaConfigKey, strconv.FormatInt(empresaID, 10), false); err != nil {
		if isMissingTableError(err) {
			return nil
		}
		return err
	}
	return nil
}

func findPowerfulSystemEmpresaByName(dbEmp *sql.DB) (*Empresa, error) {
	if dbEmp == nil {
		return nil, nil
	}
	candidates := []string{PowerfulSystemEmpresaName, PowerfulSystemEmpresaNameLegacyTypo}
	for _, candidate := range candidates {
		var id int64
		err := queryRowSQLCompat(dbEmp, `SELECT id
			FROM empresas
			WHERE LOWER(TRIM(COALESCE(nombre, ''))) = LOWER(TRIM(?))
			ORDER BY id ASC
			LIMIT 1`, candidate).Scan(&id)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return nil, err
		}
		if id > 0 {
			return GetEmpresaByScopeID(dbEmp, id)
		}
	}

	rows, err := dbEmp.Query(`SELECT id, COALESCE(nombre, '')
		FROM empresas
		WHERE LOWER(COALESCE(nombre, '')) LIKE '%powerful%control%syste%'
		ORDER BY id ASC`)
	if err != nil {
		if isMissingTableError(err) {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var nombre string
		if err := rows.Scan(&id, &nombre); err != nil {
			return nil, err
		}
		if id > 0 && IsPowerfulSystemEmpresaName(nombre) {
			return GetEmpresaByScopeID(dbEmp, id)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return nil, nil
}

func normalizePowerfulSystemEmpresaName(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" {
		return ""
	}
	var b strings.Builder
	lastWasSpace := true
	for _, r := range raw {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastWasSpace = false
			continue
		}
		if !lastWasSpace {
			b.WriteRune(' ')
			lastWasSpace = true
		}
	}
	return strings.TrimSpace(b.String())
}
