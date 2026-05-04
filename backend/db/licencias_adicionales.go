package db

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"
)

type EmpresaLicenciaAdicional struct {
	ID             int64   `json:"id"`
	EmpresaID      int64   `json:"empresa_id"`
	LicenciaID     int64   `json:"licencia_id"`
	LicenciaNombre string  `json:"licencia_nombre,omitempty"`
	CodigoFuncion  string  `json:"codigo_funcion,omitempty"`
	Descripcion    string  `json:"descripcion,omitempty"`
	Valor          float64 `json:"valor"`
	DuracionDias   int     `json:"duracion_dias"`
	ModulosHab     string  `json:"modulos_habilitados,omitempty"`
	FechaInicio    string  `json:"fecha_inicio,omitempty"`
	FechaFin       string  `json:"fecha_fin,omitempty"`
	Activo         int     `json:"activo"`
	AutoRenovar    int     `json:"auto_renovar"`
	Estado         string  `json:"estado,omitempty"`
	UsuarioCreador string  `json:"usuario_creador,omitempty"`
	Observaciones  string  `json:"observaciones,omitempty"`
	FechaCreacion  string  `json:"fecha_creacion,omitempty"`
}

type EmpresaLicenciaBundleItem struct {
	Kind          string  `json:"kind"`
	LicenciaID    int64   `json:"licencia_id"`
	Nombre        string  `json:"nombre"`
	CodigoFuncion string  `json:"codigo_funcion,omitempty"`
	ModulosHab    string  `json:"modulos_habilitados,omitempty"`
	ValorBase     float64 `json:"valor_base"`
	ValorCobrar   float64 `json:"valor_cobrar"`
	Prorrateada   bool    `json:"prorrateada"`
	FechaInicio   string  `json:"fecha_inicio,omitempty"`
	FechaFin      string  `json:"fecha_fin,omitempty"`
	AutoRenovar   bool    `json:"auto_renovar"`
	YaActiva      bool    `json:"ya_activa"`
}

type EmpresaLicenciaBundleSummary struct {
	EmpresaID               int64                       `json:"empresa_id"`
	CheckoutMode            string                      `json:"checkout_mode"`
	BaseLicencia            *Licencia                   `json:"base_licencia,omitempty"`
	ActivasAdicionales      []EmpresaLicenciaAdicional  `json:"activas_adicionales"`
	ItemsCobro              []EmpresaLicenciaBundleItem `json:"items_cobro"`
	TotalPeriodicoSiguiente float64                     `json:"total_periodico_siguiente"`
	TotalCheckout           float64                     `json:"total_checkout"`
	FechaCorteBase          string                      `json:"fecha_corte_base,omitempty"`
	TieneBaseVigente        bool                        `json:"tiene_base_vigente"`
}

func EnsureEmpresaLicenciasAdicionalesSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return nil
	}
	statements := []string{
		`CREATE TABLE IF NOT EXISTS empresa_licencias_adicionales (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			licencia_id BIGINT NOT NULL,
			fecha_inicio TEXT,
			fecha_fin TEXT,
			activo INTEGER DEFAULT 1,
			auto_renovar INTEGER DEFAULT 1,
			estado TEXT DEFAULT 'activa',
			usuario_creador TEXT,
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT),
			fecha_actualizacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT)
		)`,
		`ALTER TABLE empresa_licencias_adicionales ADD COLUMN IF NOT EXISTS empresa_id BIGINT`,
		`ALTER TABLE empresa_licencias_adicionales ADD COLUMN IF NOT EXISTS licencia_id BIGINT`,
		`ALTER TABLE empresa_licencias_adicionales ADD COLUMN IF NOT EXISTS fecha_inicio TEXT`,
		`ALTER TABLE empresa_licencias_adicionales ADD COLUMN IF NOT EXISTS fecha_fin TEXT`,
		`ALTER TABLE empresa_licencias_adicionales ADD COLUMN IF NOT EXISTS activo INTEGER DEFAULT 1`,
		`ALTER TABLE empresa_licencias_adicionales ADD COLUMN IF NOT EXISTS auto_renovar INTEGER DEFAULT 1`,
		`ALTER TABLE empresa_licencias_adicionales ADD COLUMN IF NOT EXISTS estado TEXT DEFAULT 'activa'`,
		`ALTER TABLE empresa_licencias_adicionales ADD COLUMN IF NOT EXISTS usuario_creador TEXT`,
		`ALTER TABLE empresa_licencias_adicionales ADD COLUMN IF NOT EXISTS observaciones TEXT`,
		`ALTER TABLE empresa_licencias_adicionales ADD COLUMN IF NOT EXISTS fecha_creacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT)`,
		`ALTER TABLE empresa_licencias_adicionales ADD COLUMN IF NOT EXISTS fecha_actualizacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_licencias_adicionales_empresa_licencia ON empresa_licencias_adicionales(empresa_id, licencia_id)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_licencias_adicionales_empresa_estado ON empresa_licencias_adicionales(empresa_id, activo, fecha_fin DESC)`,
	}
	for _, stmt := range statements {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func ListEmpresaLicenciasAdicionales(dbConn *sql.DB, empresaID int64, includeInactive bool) ([]EmpresaLicenciaAdicional, error) {
	if err := EnsureLicenciasSchema(dbConn); err != nil {
		return nil, err
	}
	if err := EnsureEmpresaLicenciasAdicionalesSchema(dbConn); err != nil {
		return nil, err
	}
	query := `SELECT a.id, a.empresa_id, a.licencia_id, COALESCE(l.nombre,''), COALESCE(l.codigo_funcion,''), COALESCE(l.descripcion,''), COALESCE(l.valor,0), COALESCE(l.duracion_dias,0), COALESCE(l.modulos_habilitados,''), COALESCE(a.fecha_inicio,''), COALESCE(a.fecha_fin,''), COALESCE(a.activo,1), COALESCE(a.auto_renovar,1), COALESCE(a.estado,'activa'), COALESCE(a.usuario_creador,''), COALESCE(a.observaciones,''), COALESCE(a.fecha_creacion,'')
	FROM empresa_licencias_adicionales a
	LEFT JOIN licencias l ON l.id = a.licencia_id
	WHERE a.empresa_id = ?`
	args := []interface{}{empresaID}
	if !includeInactive {
		if isPostgresDialect() {
			query += ` AND COALESCE(a.activo,1) = 1 AND (COALESCE(a.fecha_fin,'') = '' OR CAST(a.fecha_fin AS TIMESTAMP) >= CURRENT_TIMESTAMP)`
		} else {
			query += ` AND COALESCE(a.activo,1) = 1 AND (COALESCE(a.fecha_fin,'') = '' OR datetime(a.fecha_fin) >= datetime('now','localtime'))`
		}
	}
	query += ` ORDER BY COALESCE(a.activo,1) DESC, COALESCE(a.fecha_fin,'') DESC, a.id DESC`
	rows, err := querySQLCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EmpresaLicenciaAdicional, 0)
	for rows.Next() {
		var item EmpresaLicenciaAdicional
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.LicenciaID, &item.LicenciaNombre, &item.CodigoFuncion, &item.Descripcion, &item.Valor, &item.DuracionDias, &item.ModulosHab, &item.FechaInicio, &item.FechaFin, &item.Activo, &item.AutoRenovar, &item.Estado, &item.UsuarioCreador, &item.Observaciones, &item.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func GetEmpresaLicenciaAdicionalByEmpresaYLicencia(dbConn *sql.DB, empresaID, licenciaID int64) (*EmpresaLicenciaAdicional, error) {
	if err := EnsureEmpresaLicenciasAdicionalesSchema(dbConn); err != nil {
		return nil, err
	}
	query := `SELECT a.id, a.empresa_id, a.licencia_id, COALESCE(l.nombre,''), COALESCE(l.codigo_funcion,''), COALESCE(l.descripcion,''), COALESCE(l.valor,0), COALESCE(l.duracion_dias,0), COALESCE(l.modulos_habilitados,''), COALESCE(a.fecha_inicio,''), COALESCE(a.fecha_fin,''), COALESCE(a.activo,1), COALESCE(a.auto_renovar,1), COALESCE(a.estado,'activa'), COALESCE(a.usuario_creador,''), COALESCE(a.observaciones,''), COALESCE(a.fecha_creacion,'')
	FROM empresa_licencias_adicionales a
	LEFT JOIN licencias l ON l.id = a.licencia_id
	WHERE a.empresa_id = ? AND a.licencia_id = ?
	LIMIT 1`
	row := queryRowSQLCompat(dbConn, query, empresaID, licenciaID)
	var item EmpresaLicenciaAdicional
	if err := row.Scan(&item.ID, &item.EmpresaID, &item.LicenciaID, &item.LicenciaNombre, &item.CodigoFuncion, &item.Descripcion, &item.Valor, &item.DuracionDias, &item.ModulosHab, &item.FechaInicio, &item.FechaFin, &item.Activo, &item.AutoRenovar, &item.Estado, &item.UsuarioCreador, &item.Observaciones, &item.FechaCreacion); err != nil {
		return nil, err
	}
	return &item, nil
}

func UpsertEmpresaLicenciaAdicional(dbConn *sql.DB, empresaID, licenciaID int64, fechaInicio, fechaFin string, autoRenovar bool, usuarioCreador, observaciones string) (int64, error) {
	if empresaID <= 0 || licenciaID <= 0 {
		return 0, fmt.Errorf("empresa_id y licencia_id son obligatorios")
	}
	if err := EnsureLicenciasSchema(dbConn); err != nil {
		return 0, err
	}
	if err := EnsureEmpresaLicenciasAdicionalesSchema(dbConn); err != nil {
		return 0, err
	}
	lic, err := GetLicenciaByID(dbConn, licenciaID)
	if err != nil {
		return 0, err
	}
	if lic == nil {
		return 0, fmt.Errorf("licencia adicional no encontrada")
	}
	if lic.EsAdicional != 1 {
		return 0, fmt.Errorf("la licencia seleccionada no esta marcada como adicional")
	}
	autoRenovarInt := 0
	if autoRenovar {
		autoRenovarInt = 1
	}
	nowExpr := sqlNowExpr()
	existing, err := GetEmpresaLicenciaAdicionalByEmpresaYLicencia(dbConn, empresaID, licenciaID)
	if err == nil && existing != nil && existing.ID > 0 {
		_, err = execSQLCompat(dbConn, "UPDATE empresa_licencias_adicionales SET fecha_inicio = ?, fecha_fin = ?, activo = 1, auto_renovar = ?, estado = 'activa', usuario_creador = COALESCE(NULLIF(?, ''), usuario_creador), observaciones = COALESCE(NULLIF(?, ''), observaciones), fecha_actualizacion = "+nowExpr+" WHERE id = ?", fechaInicio, fechaFin, autoRenovarInt, strings.TrimSpace(usuarioCreador), strings.TrimSpace(observaciones), existing.ID)
		if err != nil {
			return 0, err
		}
		InvalidateLicenciaPermisoPolicyCacheForEmpresa(empresaID)
		return existing.ID, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	id, err := insertSQLCompat(dbConn, "INSERT INTO empresa_licencias_adicionales (empresa_id, licencia_id, fecha_inicio, fecha_fin, activo, auto_renovar, estado, usuario_creador, observaciones, fecha_creacion, fecha_actualizacion) VALUES (?, ?, ?, ?, 1, ?, 'activa', ?, ?, "+nowExpr+", "+nowExpr+")", empresaID, licenciaID, fechaInicio, fechaFin, autoRenovarInt, strings.TrimSpace(usuarioCreador), strings.TrimSpace(observaciones))
	if err == nil {
		InvalidateLicenciaPermisoPolicyCacheForEmpresa(empresaID)
	}
	return id, err
}

func SetEmpresaLicenciaAdicionalEstado(dbConn *sql.DB, empresaID, licenciaID int64, activo bool, autoRenovar *bool, observaciones string) error {
	if err := EnsureEmpresaLicenciasAdicionalesSchema(dbConn); err != nil {
		return err
	}
	nowExpr := sqlNowExpr()
	autoExpr := "auto_renovar"
	args := []interface{}{}
	if autoRenovar != nil {
		autoExpr = "?"
		if *autoRenovar {
			args = append(args, 1)
		} else {
			args = append(args, 0)
		}
	}
	activoInt := 0
	estado := "inactiva"
	if activo {
		activoInt = 1
		estado = "activa"
	}
	query := "UPDATE empresa_licencias_adicionales SET activo = ?, auto_renovar = " + autoExpr + ", estado = ?, observaciones = COALESCE(NULLIF(?, ''), observaciones), fecha_actualizacion = " + nowExpr + " WHERE empresa_id = ? AND licencia_id = ?"
	args = append([]interface{}{activoInt}, args...)
	args = append(args, estado, strings.TrimSpace(observaciones), empresaID, licenciaID)
	if _, err := execSQLCompat(dbConn, query, args...); err != nil {
		return err
	}
	InvalidateLicenciaPermisoPolicyCacheForEmpresa(empresaID)
	return nil
}

func SetLicenciaFechas(dbConn *sql.DB, licenciaID int64, fechaInicio, fechaFin string) error {
	nowExpr := sqlNowExpr()
	if _, err := execSQLCompat(dbConn, "UPDATE licencias SET fecha_inicio = ?, fecha_fin = ?, activo = 1, estado = 'activo', fecha_actualizacion = "+nowExpr+" WHERE id = ?", strings.TrimSpace(fechaInicio), strings.TrimSpace(fechaFin), licenciaID); err != nil {
		return err
	}
	invalidateLicenciaPermisoPolicyCacheForLicencia(dbConn, licenciaID)
	return nil
}

func roundAddonAmount(value float64) float64 {
	if value < 0 {
		value = 0
	}
	return math.Round(value*100) / 100
}

func parseMaybeTime(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func BuildEmpresaLicenciaBundleSummary(dbConn *sql.DB, empresaID int64, checkoutMode string, selectedAddonIDs []int64) (*EmpresaLicenciaBundleSummary, error) {
	base, err := GetActiveLicenciaByEmpresa(dbConn, empresaID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	activeAddons, err := ListEmpresaLicenciasAdicionales(dbConn, empresaID, false)
	if err != nil {
		return nil, err
	}
	summary := &EmpresaLicenciaBundleSummary{
		EmpresaID:               empresaID,
		CheckoutMode:            strings.TrimSpace(checkoutMode),
		BaseLicencia:            base,
		ActivasAdicionales:      activeAddons,
		ItemsCobro:              make([]EmpresaLicenciaBundleItem, 0),
		TotalPeriodicoSiguiente: 0,
		TotalCheckout:           0,
		TieneBaseVigente:        base != nil,
	}
	activeAddonIDs := map[int64]EmpresaLicenciaAdicional{}
	for _, addon := range activeAddons {
		activeAddonIDs[addon.LicenciaID] = addon
		if addon.Activo == 1 && addon.AutoRenovar == 1 {
			summary.TotalPeriodicoSiguiente += addon.Valor
		}
	}
	if base != nil {
		summary.TotalPeriodicoSiguiente += base.Valor
		summary.FechaCorteBase = strings.TrimSpace(base.FechaFin)
		if strings.EqualFold(summary.CheckoutMode, "empresa_bundle") {
			summary.ItemsCobro = append(summary.ItemsCobro, EmpresaLicenciaBundleItem{
				Kind:        "licencia_base",
				LicenciaID:  base.ID,
				Nombre:      base.Nombre,
				ValorBase:   roundAddonAmount(base.Valor),
				ValorCobrar: roundAddonAmount(base.Valor),
				ModulosHab:  base.ModulosHab,
				FechaInicio: base.FechaInicio,
				FechaFin:    base.FechaFin,
				AutoRenovar: true,
				YaActiva:    true,
				Prorrateada: false,
			})
			summary.TotalCheckout += base.Valor
		}
	}

	baseDuration := 30
	baseRemainingRatio := 1.0
	if base != nil && base.DuracionDias > 0 {
		baseDuration = base.DuracionDias
	}
	if base != nil {
		if endTime, ok := parseMaybeTime(base.FechaFin); ok && endTime.After(time.Now()) && baseDuration > 0 {
			remainingDays := math.Ceil(endTime.Sub(time.Now()).Hours() / 24)
			if remainingDays < 1 {
				remainingDays = 1
			}
			baseRemainingRatio = remainingDays / float64(baseDuration)
			if baseRemainingRatio > 1 {
				baseRemainingRatio = 1
			}
		}
	}

	for _, addon := range activeAddons {
		if addon.Activo == 1 && addon.AutoRenovar == 1 && strings.EqualFold(summary.CheckoutMode, "empresa_bundle") {
			item := EmpresaLicenciaBundleItem{
				Kind:          "adicional_activa",
				LicenciaID:    addon.LicenciaID,
				Nombre:        addon.LicenciaNombre,
				CodigoFuncion: addon.CodigoFuncion,
				ModulosHab:    addon.ModulosHab,
				ValorBase:     roundAddonAmount(addon.Valor),
				ValorCobrar:   roundAddonAmount(addon.Valor),
				FechaInicio:   addon.FechaInicio,
				FechaFin:      addon.FechaFin,
				AutoRenovar:   addon.AutoRenovar == 1,
				YaActiva:      true,
			}
			summary.ItemsCobro = append(summary.ItemsCobro, item)
			summary.TotalCheckout += addon.Valor
		}
	}

	seenSelected := map[int64]struct{}{}
	for _, selectedID := range selectedAddonIDs {
		if selectedID <= 0 {
			continue
		}
		if _, seen := seenSelected[selectedID]; seen {
			continue
		}
		seenSelected[selectedID] = struct{}{}
		lic, err := GetLicenciaByID(dbConn, selectedID)
		if err != nil {
			return nil, err
		}
		if lic == nil || lic.EsAdicional != 1 {
			continue
		}
		if _, already := activeAddonIDs[selectedID]; already {
			continue
		}
		item := EmpresaLicenciaBundleItem{
			Kind:          "adicional_nueva",
			LicenciaID:    lic.ID,
			Nombre:        lic.Nombre,
			CodigoFuncion: lic.CodigoFuncion,
			ModulosHab:    lic.ModulosHab,
			ValorBase:     roundAddonAmount(lic.Valor),
			AutoRenovar:   true,
			YaActiva:      false,
		}
		if strings.EqualFold(summary.CheckoutMode, "empresa_addons") {
			item.Prorrateada = true
			item.ValorCobrar = roundAddonAmount(lic.Valor * baseRemainingRatio)
		} else {
			item.ValorCobrar = roundAddonAmount(lic.Valor)
		}
		summary.ItemsCobro = append(summary.ItemsCobro, item)
		summary.TotalCheckout += item.ValorCobrar
		summary.TotalPeriodicoSiguiente += lic.Valor
	}
	summary.TotalCheckout = roundAddonAmount(summary.TotalCheckout)
	summary.TotalPeriodicoSiguiente = roundAddonAmount(summary.TotalPeriodicoSiguiente)
	return summary, nil
}
