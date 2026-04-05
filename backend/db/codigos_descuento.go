package db

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// CodigoDescuento representa un codigo promocional por empresa.
type CodigoDescuento struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	Codigo             string  `json:"codigo"`
	TipoDescuento      string  `json:"tipo_descuento"`
	Valor              float64 `json:"valor"`
	Moneda             string  `json:"moneda,omitempty"`
	MontoMinimoCompra  float64 `json:"monto_minimo_compra"`
	FechaVencimiento   string  `json:"fecha_vencimiento,omitempty"`
	UsosMaximos        int64   `json:"usos_maximos"`
	UsosActuales       int64   `json:"usos_actuales"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// CodigoDescuentoAplicado resume el valor efectivo aplicado para una venta.
type CodigoDescuentoAplicado struct {
	ID               int64   `json:"id"`
	EmpresaID        int64   `json:"empresa_id"`
	Codigo           string  `json:"codigo"`
	TipoDescuento    string  `json:"tipo_descuento"`
	ValorConfigurado float64 `json:"valor_configurado"`
	ValorAplicado    float64 `json:"valor_aplicado"`
	Moneda           string  `json:"moneda,omitempty"`
	FechaVencimiento string  `json:"fecha_vencimiento,omitempty"`
}

const (
	defaultCodigoDescuentoTipo   = "valor_fijo"
	defaultCodigoDescuentoMoneda = "COP"
)

var codigoDescuentoCharset = []rune("ABCDEFGHJKLMNPQRSTUVWXYZ23456789")

// EnsureEmpresaCodigosDescuentoSchema crea/migra la tabla de codigos de descuento por empresa.
func EnsureEmpresaCodigosDescuentoSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS codigos_de_descuento (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			tipo_descuento TEXT DEFAULT 'valor_fijo',
			valor REAL DEFAULT 0,
			moneda TEXT DEFAULT 'COP',
			monto_minimo_compra REAL DEFAULT 0,
			fecha_vencimiento TEXT,
			usos_maximos INTEGER DEFAULT 1,
			usos_actuales INTEGER DEFAULT 0,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_codigos_descuento_empresa_codigo ON codigos_de_descuento(empresa_id, codigo);`,
		`CREATE INDEX IF NOT EXISTS ix_codigos_descuento_empresa_estado ON codigos_de_descuento(empresa_id, estado);`,
		`CREATE INDEX IF NOT EXISTS ix_codigos_descuento_empresa_vencimiento ON codigos_de_descuento(empresa_id, fecha_vencimiento);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "codigos_de_descuento", "empresa_id", "INTEGER NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_de_descuento", "codigo", "TEXT NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_de_descuento", "tipo_descuento", "TEXT DEFAULT 'valor_fijo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_de_descuento", "valor", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_de_descuento", "moneda", "TEXT DEFAULT 'COP'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_de_descuento", "monto_minimo_compra", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_de_descuento", "fecha_vencimiento", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_de_descuento", "usos_maximos", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_de_descuento", "usos_actuales", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_de_descuento", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_de_descuento", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_de_descuento", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_de_descuento", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

func normalizeCodigoDescuentoTipo(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	switch v {
	case "porcentaje", "percent", "pct":
		return "porcentaje"
	case "valor", "valor_fijo", "fixed", "monto":
		return "valor_fijo"
	default:
		return defaultCodigoDescuentoTipo
	}
}

func normalizeCodigoDescuentoEstado(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "inactivo" {
		return "inactivo"
	}
	return "activo"
}

func normalizeCodigoDescuento(v string) string {
	raw := strings.ToUpper(strings.TrimSpace(v))
	if raw == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(raw))
	for _, r := range raw {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func normalizeCodigoDescuentoMoneda(v string) string {
	v = strings.ToUpper(strings.TrimSpace(v))
	if v == "" {
		return defaultCodigoDescuentoMoneda
	}
	return v
}

func randomCodigoDescuentoChars(size int) string {
	if size <= 0 {
		size = 6
	}
	b := make([]rune, 0, size)
	max := big.NewInt(int64(len(codigoDescuentoCharset)))
	for i := 0; i < size; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			idx := time.Now().UnixNano() % int64(len(codigoDescuentoCharset))
			b = append(b, codigoDescuentoCharset[idx])
			continue
		}
		b = append(b, codigoDescuentoCharset[n.Int64()])
	}
	return string(b)
}

// GenerateCodigoDescuentoAutomatico genera un codigo unico legible para promociones.
func GenerateCodigoDescuentoAutomatico(prefix string) string {
	base := normalizeCodigoDescuento(prefix)
	if base == "" {
		base = "DSCT"
	}
	fecha := time.Now().Format("060102")
	return base + fecha + randomCodigoDescuentoChars(4)
}

func parseCodigoDescuentoDate(raw string) (time.Time, error) {
	trim := strings.TrimSpace(raw)
	if trim == "" {
		return time.Time{}, nil
	}
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}
	for _, f := range formats {
		if tm, err := time.ParseInLocation(f, trim, time.Local); err == nil {
			return tm, nil
		}
	}
	return time.Time{}, fmt.Errorf("fecha_vencimiento invalida")
}

func validateCodigoDescuentoPayload(payload *CodigoDescuento, requireID bool) error {
	if payload == nil {
		return fmt.Errorf("payload de codigo de descuento invalido")
	}
	if requireID && payload.ID <= 0 {
		return fmt.Errorf("id es obligatorio")
	}
	if payload.EmpresaID <= 0 {
		return fmt.Errorf("empresa_id es obligatorio")
	}

	payload.Codigo = normalizeCodigoDescuento(payload.Codigo)
	if payload.Codigo == "" {
		payload.Codigo = GenerateCodigoDescuentoAutomatico("DSCT")
	}
	payload.TipoDescuento = normalizeCodigoDescuentoTipo(payload.TipoDescuento)
	payload.Moneda = normalizeCodigoDescuentoMoneda(payload.Moneda)
	payload.Estado = normalizeCodigoDescuentoEstado(payload.Estado)

	if payload.Valor <= 0 {
		return fmt.Errorf("valor debe ser mayor a cero")
	}
	if payload.TipoDescuento == "porcentaje" && payload.Valor > 100 {
		return fmt.Errorf("valor porcentaje no puede superar 100")
	}
	if payload.MontoMinimoCompra < 0 {
		return fmt.Errorf("monto_minimo_compra no puede ser negativo")
	}
	if payload.UsosMaximos <= 0 {
		payload.UsosMaximos = 1
	}
	if payload.UsosActuales < 0 {
		payload.UsosActuales = 0
	}
	if payload.UsosActuales > payload.UsosMaximos {
		return fmt.Errorf("usos_actuales no puede superar usos_maximos")
	}
	if payload.FechaVencimiento != "" {
		if _, err := parseCodigoDescuentoDate(payload.FechaVencimiento); err != nil {
			return err
		}
	}

	payload.Valor = round2(payload.Valor)
	payload.MontoMinimoCompra = round2(payload.MontoMinimoCompra)
	return nil
}

// CreateCodigoDescuento crea un codigo de descuento por empresa.
func CreateCodigoDescuento(dbConn *sql.DB, payload CodigoDescuento) (int64, error) {
	autoGenerate := strings.TrimSpace(payload.Codigo) == ""
	const maxAttempts = 4

	for attempt := 0; attempt < maxAttempts; attempt++ {
		candidate := payload
		if autoGenerate {
			candidate.Codigo = GenerateCodigoDescuentoAutomatico("DSCT")
		}
		if err := validateCodigoDescuentoPayload(&candidate, false); err != nil {
			return 0, err
		}

		res, err := dbConn.Exec(`INSERT INTO codigos_de_descuento (
			empresa_id,
			codigo,
			tipo_descuento,
			valor,
			moneda,
			monto_minimo_compra,
			fecha_vencimiento,
			usos_maximos,
			usos_actuales,
			fecha_creacion,
			fecha_actualizacion,
			usuario_creador,
			estado,
			observaciones
		) VALUES (?, ?, ?, ?, ?, ?, NULLIF(?, ''), ?, ?, datetime('now','localtime'), datetime('now','localtime'), ?, ?, ?)`,
			candidate.EmpresaID,
			candidate.Codigo,
			candidate.TipoDescuento,
			candidate.Valor,
			candidate.Moneda,
			candidate.MontoMinimoCompra,
			strings.TrimSpace(candidate.FechaVencimiento),
			candidate.UsosMaximos,
			candidate.UsosActuales,
			strings.TrimSpace(candidate.UsuarioCreador),
			candidate.Estado,
			strings.TrimSpace(candidate.Observaciones),
		)
		if err == nil {
			return res.LastInsertId()
		}

		if !autoGenerate || !strings.Contains(strings.ToLower(err.Error()), "unique") {
			return 0, err
		}
	}

	return 0, fmt.Errorf("no se pudo generar un codigo de descuento unico")
}

// GetCodigosDescuentoByEmpresa lista codigos por empresa con filtros opcionales.
func GetCodigosDescuentoByEmpresa(dbConn *sql.DB, empresaID int64, filtro, estado string, incluirInactivos bool, limit, offset int) ([]CodigoDescuento, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	query := `SELECT
		id,
		empresa_id,
		COALESCE(codigo, ''),
		COALESCE(tipo_descuento, 'valor_fijo'),
		COALESCE(valor, 0),
		COALESCE(moneda, 'COP'),
		COALESCE(monto_minimo_compra, 0),
		COALESCE(fecha_vencimiento, ''),
		COALESCE(usos_maximos, 1),
		COALESCE(usos_actuales, 0),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM codigos_de_descuento
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	if strings.TrimSpace(estado) != "" {
		query += ` AND COALESCE(estado, 'activo') = ?`
		args = append(args, normalizeCodigoDescuentoEstado(estado))
	} else if !incluirInactivos {
		query += ` AND COALESCE(estado, 'activo') = 'activo'`
	}

	filtro = strings.TrimSpace(filtro)
	if filtro != "" {
		like := "%" + filtro + "%"
		query += ` AND (codigo LIKE ? OR observaciones LIKE ?)`
		args = append(args, like, like)
	}

	query += ` ORDER BY id DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]CodigoDescuento, 0)
	for rows.Next() {
		var item CodigoDescuento
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.Codigo,
			&item.TipoDescuento,
			&item.Valor,
			&item.Moneda,
			&item.MontoMinimoCompra,
			&item.FechaVencimiento,
			&item.UsosMaximos,
			&item.UsosActuales,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}

	return out, nil
}

// GetCodigoDescuentoByID obtiene un codigo de descuento puntual por empresa.
func GetCodigoDescuentoByID(dbConn *sql.DB, empresaID, codigoID int64) (*CodigoDescuento, error) {
	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(codigo, ''),
		COALESCE(tipo_descuento, 'valor_fijo'),
		COALESCE(valor, 0),
		COALESCE(moneda, 'COP'),
		COALESCE(monto_minimo_compra, 0),
		COALESCE(fecha_vencimiento, ''),
		COALESCE(usos_maximos, 1),
		COALESCE(usos_actuales, 0),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM codigos_de_descuento
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, codigoID)

	var item CodigoDescuento
	if err := row.Scan(
		&item.ID,
		&item.EmpresaID,
		&item.Codigo,
		&item.TipoDescuento,
		&item.Valor,
		&item.Moneda,
		&item.MontoMinimoCompra,
		&item.FechaVencimiento,
		&item.UsosMaximos,
		&item.UsosActuales,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	); err != nil {
		return nil, err
	}

	return &item, nil
}

func getCodigoDescuentoByCode(dbConn *sql.DB, empresaID int64, codigo string) (*CodigoDescuento, error) {
	codigo = normalizeCodigoDescuento(codigo)
	if codigo == "" {
		return nil, fmt.Errorf("codigo de descuento obligatorio")
	}

	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(codigo, ''),
		COALESCE(tipo_descuento, 'valor_fijo'),
		COALESCE(valor, 0),
		COALESCE(moneda, 'COP'),
		COALESCE(monto_minimo_compra, 0),
		COALESCE(fecha_vencimiento, ''),
		COALESCE(usos_maximos, 1),
		COALESCE(usos_actuales, 0),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM codigos_de_descuento
	WHERE empresa_id = ? AND codigo = ?
	LIMIT 1`, empresaID, codigo)

	var item CodigoDescuento
	if err := row.Scan(
		&item.ID,
		&item.EmpresaID,
		&item.Codigo,
		&item.TipoDescuento,
		&item.Valor,
		&item.Moneda,
		&item.MontoMinimoCompra,
		&item.FechaVencimiento,
		&item.UsosMaximos,
		&item.UsosActuales,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	); err != nil {
		return nil, err
	}

	return &item, nil
}

// UpdateCodigoDescuento actualiza un codigo de descuento existente.
func UpdateCodigoDescuento(dbConn *sql.DB, payload CodigoDescuento) error {
	if err := validateCodigoDescuentoPayload(&payload, true); err != nil {
		return err
	}

	res, err := dbConn.Exec(`UPDATE codigos_de_descuento SET
		codigo = ?,
		tipo_descuento = ?,
		valor = ?,
		moneda = ?,
		monto_minimo_compra = ?,
		fecha_vencimiento = NULLIF(?, ''),
		usos_maximos = ?,
		usos_actuales = ?,
		estado = ?,
		observaciones = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE id = ? AND empresa_id = ?`,
		payload.Codigo,
		payload.TipoDescuento,
		payload.Valor,
		payload.Moneda,
		payload.MontoMinimoCompra,
		strings.TrimSpace(payload.FechaVencimiento),
		payload.UsosMaximos,
		payload.UsosActuales,
		payload.Estado,
		strings.TrimSpace(payload.Observaciones),
		payload.ID,
		payload.EmpresaID,
	)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteCodigoDescuento elimina un codigo de descuento por empresa.
func DeleteCodigoDescuento(dbConn *sql.DB, empresaID, codigoID int64) error {
	res, err := dbConn.Exec(`DELETE FROM codigos_de_descuento WHERE empresa_id = ? AND id = ?`, empresaID, codigoID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SetCodigoDescuentoEstado activa o desactiva un codigo por empresa.
func SetCodigoDescuentoEstado(dbConn *sql.DB, empresaID, codigoID int64, estado string) error {
	nuevoEstado := normalizeCodigoDescuentoEstado(estado)
	res, err := dbConn.Exec(`UPDATE codigos_de_descuento
	SET estado = ?, fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`, nuevoEstado, empresaID, codigoID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func validateCodigoDescuentoAplicacion(item *CodigoDescuento, montoBase float64) error {
	if item == nil {
		return fmt.Errorf("codigo de descuento no encontrado")
	}
	if strings.TrimSpace(strings.ToLower(item.Estado)) != "activo" {
		return fmt.Errorf("codigo de descuento inactivo")
	}
	if item.UsosMaximos > 0 && item.UsosActuales >= item.UsosMaximos {
		return fmt.Errorf("codigo de descuento sin usos disponibles")
	}
	if item.FechaVencimiento != "" {
		vence, err := parseCodigoDescuentoDate(item.FechaVencimiento)
		if err != nil {
			return err
		}
		if !vence.IsZero() {
			if len(strings.TrimSpace(item.FechaVencimiento)) <= len("2006-01-02") {
				finDelDia := time.Date(vence.Year(), vence.Month(), vence.Day(), 23, 59, 59, 0, time.Local)
				if time.Now().After(finDelDia) {
					return fmt.Errorf("codigo de descuento vencido")
				}
			} else if time.Now().After(vence) {
				return fmt.Errorf("codigo de descuento vencido")
			}
		}
	}
	if montoBase < item.MontoMinimoCompra {
		return fmt.Errorf("el carrito no cumple el monto minimo para aplicar este codigo")
	}
	return nil
}

func calcularValorDescuento(item *CodigoDescuento, montoBase float64) float64 {
	if item == nil || montoBase <= 0 {
		return 0
	}
	tipo := normalizeCodigoDescuentoTipo(item.TipoDescuento)
	if tipo == "porcentaje" {
		return round2(montoBase * (item.Valor / 100))
	}
	if item.Valor <= 0 {
		return 0
	}
	return round2(item.Valor)
}

// ResolveCodigoDescuentoParaMonto valida un codigo y devuelve el valor aplicado para el monto dado.
func ResolveCodigoDescuentoParaMonto(dbConn *sql.DB, empresaID int64, codigo string, montoBase float64) (*CodigoDescuentoAplicado, error) {
	item, err := getCodigoDescuentoByCode(dbConn, empresaID, codigo)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("codigo de descuento no encontrado")
		}
		return nil, err
	}

	montoBase = round2(montoBase)
	if montoBase < 0 {
		montoBase = 0
	}
	if err := validateCodigoDescuentoAplicacion(item, montoBase); err != nil {
		return nil, err
	}

	valorAplicado := calcularValorDescuento(item, montoBase)
	if valorAplicado <= 0 {
		return nil, fmt.Errorf("codigo de descuento sin valor aplicable")
	}
	if valorAplicado > montoBase {
		valorAplicado = montoBase
	}

	return &CodigoDescuentoAplicado{
		ID:               item.ID,
		EmpresaID:        item.EmpresaID,
		Codigo:           item.Codigo,
		TipoDescuento:    normalizeCodigoDescuentoTipo(item.TipoDescuento),
		ValorConfigurado: item.Valor,
		ValorAplicado:    round2(valorAplicado),
		Moneda:           item.Moneda,
		FechaVencimiento: item.FechaVencimiento,
	}, nil
}

func markCodigoDescuentoUsoTx(tx *sql.Tx, empresaID, codigoID int64) error {
	if tx == nil || codigoID <= 0 {
		return nil
	}

	var item CodigoDescuento
	err := tx.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(codigo, ''),
		COALESCE(tipo_descuento, 'valor_fijo'),
		COALESCE(valor, 0),
		COALESCE(moneda, 'COP'),
		COALESCE(monto_minimo_compra, 0),
		COALESCE(fecha_vencimiento, ''),
		COALESCE(usos_maximos, 1),
		COALESCE(usos_actuales, 0),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM codigos_de_descuento
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, codigoID).Scan(
		&item.ID,
		&item.EmpresaID,
		&item.Codigo,
		&item.TipoDescuento,
		&item.Valor,
		&item.Moneda,
		&item.MontoMinimoCompra,
		&item.FechaVencimiento,
		&item.UsosMaximos,
		&item.UsosActuales,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("codigo de descuento no encontrado")
		}
		return err
	}
	if err := validateCodigoDescuentoAplicacion(&item, item.MontoMinimoCompra); err != nil {
		return err
	}

	res, err := tx.Exec(`UPDATE codigos_de_descuento
	SET usos_actuales = usos_actuales + 1,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`, empresaID, codigoID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("codigo de descuento no encontrado")
	}
	return nil
}
