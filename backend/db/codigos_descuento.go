package db

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"sort"
	"strconv"
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
	SegmentoCliente    string  `json:"segmento_cliente,omitempty"`
	CanalVenta         string  `json:"canal_venta,omitempty"`
	HorarioDesde       string  `json:"horario_desde,omitempty"`
	HorarioHasta       string  `json:"horario_hasta,omitempty"`
	DiasSemana         string  `json:"dias_semana,omitempty"`
	MaxUsosPorCliente  int64   `json:"max_usos_por_cliente,omitempty"`
	VentanaHorasFraude int64   `json:"ventana_horas_fraude,omitempty"`
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
	SegmentoCliente  string  `json:"segmento_cliente,omitempty"`
	CanalVenta       string  `json:"canal_venta,omitempty"`
	FechaVencimiento string  `json:"fecha_vencimiento,omitempty"`
}

// CodigoDescuentoContexto permite validar reglas avanzadas por canal/cliente/horario.
type CodigoDescuentoContexto struct {
	ClienteID      int64     `json:"cliente_id,omitempty"`
	CanalVenta     string    `json:"canal_venta,omitempty"`
	CarritoID      int64     `json:"carrito_id,omitempty"`
	RequestID      string    `json:"request_id,omitempty"`
	FechaOperacion time.Time `json:"-"`
}

// CodigoDescuentoRedencion representa la trazabilidad de uso de codigos de descuento.
type CodigoDescuentoRedencion struct {
	ID                  int64   `json:"id"`
	EmpresaID           int64   `json:"empresa_id"`
	CodigoDescuentoID   int64   `json:"codigo_descuento_id"`
	CarritoID           int64   `json:"carrito_id,omitempty"`
	ClienteID           int64   `json:"cliente_id,omitempty"`
	Codigo              string  `json:"codigo"`
	CanalVenta          string  `json:"canal_venta,omitempty"`
	SegmentoCliente     string  `json:"segmento_cliente,omitempty"`
	MontoBase           float64 `json:"monto_base"`
	ValorDescuento      float64 `json:"valor_descuento"`
	EstadoRedencion     string  `json:"estado_redencion"`
	Motivo              string  `json:"motivo,omitempty"`
	ReferenciaOperacion string  `json:"referencia_operacion,omitempty"`
	FechaRedencion      string  `json:"fecha_redencion,omitempty"`
	FechaCreacion       string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion  string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador      string  `json:"usuario_creador,omitempty"`
	Estado              string  `json:"estado,omitempty"`
	Observaciones       string  `json:"observaciones,omitempty"`
}

const (
	defaultCodigoDescuentoTipo         = "valor_fijo"
	defaultCodigoDescuentoMoneda       = "COP"
	defaultCodigoDescuentoSegmento     = "todos"
	defaultCodigoDescuentoCanalVenta   = "todos"
	defaultCodigoDescuentoVentanaHoras = int64(24)
)

var codigoDescuentoCharset = []rune("ABCDEFGHJKLMNPQRSTUVWXYZ23456789")

// EnsureEmpresaCodigosDescuentoSchema crea/migra la tabla de codigos de descuento por empresa.
func EnsureEmpresaCodigosDescuentoSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS codigos_de_descuento (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			tipo_descuento TEXT DEFAULT 'valor_fijo',
			valor REAL DEFAULT 0,
			moneda TEXT DEFAULT 'COP',
			monto_minimo_compra REAL DEFAULT 0,
			segmento_cliente TEXT DEFAULT 'todos',
			canal_venta TEXT DEFAULT 'todos',
			horario_desde TEXT,
			horario_hasta TEXT,
			dias_semana TEXT,
			max_usos_por_cliente INTEGER DEFAULT 0,
			ventana_horas_fraude INTEGER DEFAULT 24,
			fecha_vencimiento TEXT,
			usos_maximos INTEGER DEFAULT 1,
			usos_actuales INTEGER DEFAULT 0,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS codigos_descuento_redenciones (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo_descuento_id INTEGER NOT NULL,
			carrito_id INTEGER,
			cliente_id INTEGER,
			codigo TEXT NOT NULL,
			canal_venta TEXT DEFAULT 'mostrador',
			segmento_cliente TEXT DEFAULT 'desconocido',
			monto_base REAL DEFAULT 0,
			valor_descuento REAL DEFAULT 0,
			estado_redencion TEXT DEFAULT 'aplicada',
			motivo TEXT,
			referencia_operacion TEXT,
			fecha_redencion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_codigos_descuento_empresa_codigo ON codigos_de_descuento(empresa_id, codigo);`,
		`CREATE INDEX IF NOT EXISTS ix_codigos_descuento_empresa_estado ON codigos_de_descuento(empresa_id, estado);`,
		`CREATE INDEX IF NOT EXISTS ix_codigos_descuento_empresa_vencimiento ON codigos_de_descuento(empresa_id, fecha_vencimiento);`,
		`CREATE INDEX IF NOT EXISTS ix_codigos_descuento_redenciones_empresa_codigo_cliente_fecha ON codigos_descuento_redenciones(empresa_id, codigo_descuento_id, cliente_id, fecha_redencion DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_codigos_descuento_redenciones_empresa_carrito ON codigos_descuento_redenciones(empresa_id, carrito_id, estado_redencion);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_codigos_descuento_redenciones_aplicada ON codigos_descuento_redenciones(empresa_id, codigo_descuento_id, carrito_id, estado_redencion);`,
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
	if err := ensureColumnIfMissing(dbConn, "codigos_de_descuento", "segmento_cliente", "TEXT DEFAULT 'todos'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_de_descuento", "canal_venta", "TEXT DEFAULT 'todos'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_de_descuento", "horario_desde", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_de_descuento", "horario_hasta", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_de_descuento", "dias_semana", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_de_descuento", "max_usos_por_cliente", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_de_descuento", "ventana_horas_fraude", "INTEGER DEFAULT 24"); err != nil {
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

	if err := ensureColumnIfMissing(dbConn, "codigos_descuento_redenciones", "empresa_id", "INTEGER NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_descuento_redenciones", "codigo_descuento_id", "INTEGER NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_descuento_redenciones", "carrito_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_descuento_redenciones", "cliente_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_descuento_redenciones", "codigo", "TEXT NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_descuento_redenciones", "canal_venta", "TEXT DEFAULT 'mostrador'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_descuento_redenciones", "segmento_cliente", "TEXT DEFAULT 'desconocido'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_descuento_redenciones", "monto_base", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_descuento_redenciones", "valor_descuento", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_descuento_redenciones", "estado_redencion", "TEXT DEFAULT 'aplicada'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_descuento_redenciones", "motivo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_descuento_redenciones", "referencia_operacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_descuento_redenciones", "fecha_redencion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_descuento_redenciones", "fecha_creacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_descuento_redenciones", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_descuento_redenciones", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_descuento_redenciones", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "codigos_descuento_redenciones", "observaciones", "TEXT"); err != nil {
		return err
	}

	// Regla comercial vigente: cada codigo de descuento se consume una sola vez
	// por empresa. Se normalizan registros heredados para que no conserven cupos
	// mayores ni queden redenciones historicas con contador disponible.
	if _, err := execSQLCompat(dbConn, `UPDATE codigos_de_descuento
		SET usos_maximos = 1,
			usos_actuales = CASE WHEN COALESCE(usos_actuales, 0) > 1 THEN 1 ELSE COALESCE(usos_actuales, 0) END,
			fecha_actualizacion = `+sqlNowExpr()+`
		WHERE COALESCE(usos_maximos, 1) <> 1
			OR COALESCE(usos_actuales, 0) > 1`); err != nil {
		return err
	}
	if _, err := execSQLCompat(dbConn, `UPDATE codigos_de_descuento
		SET usos_actuales = 1,
			fecha_actualizacion = `+sqlNowExpr()+`
		WHERE COALESCE(usos_actuales, 0) < 1
			AND EXISTS (
				SELECT 1
				FROM codigos_descuento_redenciones r
				WHERE r.empresa_id = codigos_de_descuento.empresa_id
					AND r.codigo_descuento_id = codigos_de_descuento.id
			)`); err != nil {
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

func normalizeCodigoDescuentoSegmento(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	v = strings.ReplaceAll(v, " ", "_")
	if v == "" || v == "all" {
		return defaultCodigoDescuentoSegmento
	}
	var b strings.Builder
	b.Grow(len(v))
	for _, r := range v {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return defaultCodigoDescuentoSegmento
	}
	return out
}

func normalizeCodigoDescuentoCanal(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" || v == "all" {
		return defaultCodigoDescuentoCanalVenta
	}
	v = strings.ReplaceAll(v, " ", "_")
	if v == "todos" {
		return defaultCodigoDescuentoCanalVenta
	}
	return defaultCanalVenta(v)
}

func normalizeCodigoDescuentoHorario(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	formats := []string{"15:04", "15:04:05"}
	for _, format := range formats {
		if tm, err := time.Parse(format, raw); err == nil {
			return tm.Format("15:04"), nil
		}
	}
	return "", fmt.Errorf("horario invalido; use HH:MM")
}

func parseCodigoDescuentoHorarioMin(raw string) (int, error) {
	normalized, err := normalizeCodigoDescuentoHorario(raw)
	if err != nil {
		return 0, err
	}
	if normalized == "" {
		return 0, fmt.Errorf("horario vacio")
	}
	parts := strings.Split(normalized, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("horario invalido")
	}
	hh, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("hora invalida")
	}
	mm, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("minuto invalido")
	}
	if hh < 0 || hh > 23 || mm < 0 || mm > 59 {
		return 0, fmt.Errorf("horario invalido")
	}
	return hh*60 + mm, nil
}

func normalizeCodigoDescuentoDiasSemana(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	replacer := strings.NewReplacer(";", ",", " ", ",", "|", ",")
	raw = replacer.Replace(raw)
	parts := strings.Split(raw, ",")
	uniq := map[int]struct{}{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		v, err := strconv.Atoi(part)
		if err != nil {
			return "", fmt.Errorf("dias_semana invalido; use valores entre 1 y 7")
		}
		if v < 1 || v > 7 {
			return "", fmt.Errorf("dias_semana invalido; use valores entre 1 y 7")
		}
		uniq[v] = struct{}{}
	}
	if len(uniq) == 0 {
		return "", nil
	}
	keys := make([]int, 0, len(uniq))
	for k := range uniq {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	values := make([]string, 0, len(keys))
	for _, k := range keys {
		values = append(values, strconv.Itoa(k))
	}
	return strings.Join(values, ","), nil
}

func parseCodigoDescuentoDiasSemanaSet(raw string) map[int]struct{} {
	set := map[int]struct{}{}
	for _, part := range strings.Split(strings.TrimSpace(raw), ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if v, err := strconv.Atoi(part); err == nil && v >= 1 && v <= 7 {
			set[v] = struct{}{}
		}
	}
	return set
}

func isoWeekday(t time.Time) int {
	wd := int(t.Weekday())
	if wd == 0 {
		return 7
	}
	return wd
}

func nowOrDefault(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now()
	}
	return t
}

func resolveCodigoDescuentoContexto(ctx CodigoDescuentoContexto) CodigoDescuentoContexto {
	ctx.CanalVenta = normalizeCodigoDescuentoCanal(ctx.CanalVenta)
	ctx.FechaOperacion = nowOrDefault(ctx.FechaOperacion)
	return ctx
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
	// Formato moderno y sencillo: PREFIJO-XXXX-XXXX (dos grupos de 4 caracteres)
	return base + "-" + randomCodigoDescuentoChars(4) + "-" + randomCodigoDescuentoChars(4)
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
	payload.SegmentoCliente = normalizeCodigoDescuentoSegmento(payload.SegmentoCliente)
	payload.CanalVenta = normalizeCodigoDescuentoCanal(payload.CanalVenta)

	if payload.Valor <= 0 {
		return fmt.Errorf("valor debe ser mayor a cero")
	}
	if payload.TipoDescuento == "porcentaje" && payload.Valor > 100 {
		return fmt.Errorf("valor porcentaje no puede superar 100")
	}
	if payload.MontoMinimoCompra < 0 {
		return fmt.Errorf("monto_minimo_compra no puede ser negativo")
	}
	payload.UsosMaximos = 1
	if payload.UsosActuales < 0 {
		payload.UsosActuales = 0
	}
	if payload.UsosActuales > 1 {
		payload.UsosActuales = 1
	}
	if payload.MaxUsosPorCliente < 0 {
		return fmt.Errorf("max_usos_por_cliente no puede ser negativo")
	}
	if payload.VentanaHorasFraude <= 0 {
		payload.VentanaHorasFraude = defaultCodigoDescuentoVentanaHoras
	}
	if payload.VentanaHorasFraude > 24*30 {
		return fmt.Errorf("ventana_horas_fraude no puede superar 720 horas")
	}

	horarioDesde, err := normalizeCodigoDescuentoHorario(payload.HorarioDesde)
	if err != nil {
		return err
	}
	horarioHasta, err := normalizeCodigoDescuentoHorario(payload.HorarioHasta)
	if err != nil {
		return err
	}
	if (horarioDesde == "" && horarioHasta != "") || (horarioDesde != "" && horarioHasta == "") {
		return fmt.Errorf("debe configurar horario_desde y horario_hasta juntos")
	}
	payload.HorarioDesde = horarioDesde
	payload.HorarioHasta = horarioHasta

	diasSemana, err := normalizeCodigoDescuentoDiasSemana(payload.DiasSemana)
	if err != nil {
		return err
	}
	payload.DiasSemana = diasSemana
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

		id, err := insertSQLCompat(dbConn, `INSERT INTO codigos_de_descuento (
			empresa_id,
			codigo,
			tipo_descuento,
			valor,
			moneda,
			monto_minimo_compra,
			segmento_cliente,
			canal_venta,
			horario_desde,
			horario_hasta,
			dias_semana,
			max_usos_por_cliente,
			ventana_horas_fraude,
			fecha_vencimiento,
			usos_maximos,
			usos_actuales,
			fecha_creacion,
			fecha_actualizacion,
			usuario_creador,
			estado,
			observaciones
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), ?, ?, NULLIF(?, ''), ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?, ?)`,
			candidate.EmpresaID,
			candidate.Codigo,
			candidate.TipoDescuento,
			candidate.Valor,
			candidate.Moneda,
			candidate.MontoMinimoCompra,
			candidate.SegmentoCliente,
			candidate.CanalVenta,
			strings.TrimSpace(candidate.HorarioDesde),
			strings.TrimSpace(candidate.HorarioHasta),
			strings.TrimSpace(candidate.DiasSemana),
			candidate.MaxUsosPorCliente,
			candidate.VentanaHorasFraude,
			strings.TrimSpace(candidate.FechaVencimiento),
			candidate.UsosMaximos,
			candidate.UsosActuales,
			strings.TrimSpace(candidate.UsuarioCreador),
			candidate.Estado,
			strings.TrimSpace(candidate.Observaciones),
		)
		if err == nil {
			return id, nil
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
		COALESCE(segmento_cliente, 'todos'),
		COALESCE(canal_venta, 'todos'),
		COALESCE(horario_desde, ''),
		COALESCE(horario_hasta, ''),
		COALESCE(dias_semana, ''),
		COALESCE(max_usos_por_cliente, 0),
		COALESCE(ventana_horas_fraude, 24),
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
			&item.SegmentoCliente,
			&item.CanalVenta,
			&item.HorarioDesde,
			&item.HorarioHasta,
			&item.DiasSemana,
			&item.MaxUsosPorCliente,
			&item.VentanaHorasFraude,
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
		COALESCE(segmento_cliente, 'todos'),
		COALESCE(canal_venta, 'todos'),
		COALESCE(horario_desde, ''),
		COALESCE(horario_hasta, ''),
		COALESCE(dias_semana, ''),
		COALESCE(max_usos_por_cliente, 0),
		COALESCE(ventana_horas_fraude, 24),
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
		&item.SegmentoCliente,
		&item.CanalVenta,
		&item.HorarioDesde,
		&item.HorarioHasta,
		&item.DiasSemana,
		&item.MaxUsosPorCliente,
		&item.VentanaHorasFraude,
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
		COALESCE(segmento_cliente, 'todos'),
		COALESCE(canal_venta, 'todos'),
		COALESCE(horario_desde, ''),
		COALESCE(horario_hasta, ''),
		COALESCE(dias_semana, ''),
		COALESCE(max_usos_por_cliente, 0),
		COALESCE(ventana_horas_fraude, 24),
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
		&item.SegmentoCliente,
		&item.CanalVenta,
		&item.HorarioDesde,
		&item.HorarioHasta,
		&item.DiasSemana,
		&item.MaxUsosPorCliente,
		&item.VentanaHorasFraude,
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
		segmento_cliente = ?,
		canal_venta = ?,
		horario_desde = NULLIF(?, ''),
		horario_hasta = NULLIF(?, ''),
		dias_semana = NULLIF(?, ''),
		max_usos_por_cliente = ?,
		ventana_horas_fraude = ?,
		fecha_vencimiento = NULLIF(?, ''),
		usos_maximos = ?,
		usos_actuales = ?,
		estado = ?,
		observaciones = ?,
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE id = ? AND empresa_id = ?`,
		payload.Codigo,
		payload.TipoDescuento,
		payload.Valor,
		payload.Moneda,
		payload.MontoMinimoCompra,
		payload.SegmentoCliente,
		payload.CanalVenta,
		strings.TrimSpace(payload.HorarioDesde),
		strings.TrimSpace(payload.HorarioHasta),
		strings.TrimSpace(payload.DiasSemana),
		payload.MaxUsosPorCliente,
		payload.VentanaHorasFraude,
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
	SET estado = ?, fecha_actualizacion = CURRENT_TIMESTAMP
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
	if item.UsosActuales >= 1 {
		return fmt.Errorf("codigo de descuento ya usado por esta empresa")
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

func codigoDescuentoHoraEnVentana(minutoActual, minutoDesde, minutoHasta int) bool {
	if minutoDesde <= minutoHasta {
		return minutoActual >= minutoDesde && minutoActual <= minutoHasta
	}
	return minutoActual >= minutoDesde || minutoActual <= minutoHasta
}

func validateCodigoDescuentoReglasContexto(dbConn *sql.DB, item *CodigoDescuento, ctx CodigoDescuentoContexto) (string, error) {
	if item == nil {
		return "", fmt.Errorf("codigo de descuento no encontrado")
	}
	ctx = resolveCodigoDescuentoContexto(ctx)

	canalRegla := normalizeCodigoDescuentoCanal(item.CanalVenta)
	if canalRegla != defaultCodigoDescuentoCanalVenta {
		if ctx.CanalVenta == "" || ctx.CanalVenta == defaultCodigoDescuentoCanalVenta {
			return "", fmt.Errorf("el codigo de descuento requiere un canal_venta especifico")
		}
		if ctx.CanalVenta != canalRegla {
			return "", fmt.Errorf("el codigo de descuento no aplica para el canal de venta actual")
		}
	}

	if strings.TrimSpace(item.HorarioDesde) != "" || strings.TrimSpace(item.HorarioHasta) != "" {
		desdeMin, err := parseCodigoDescuentoHorarioMin(item.HorarioDesde)
		if err != nil {
			return "", err
		}
		hastaMin, err := parseCodigoDescuentoHorarioMin(item.HorarioHasta)
		if err != nil {
			return "", err
		}
		ahoraMin := ctx.FechaOperacion.Hour()*60 + ctx.FechaOperacion.Minute()
		if !codigoDescuentoHoraEnVentana(ahoraMin, desdeMin, hastaMin) {
			return "", fmt.Errorf("el codigo de descuento no aplica en el horario actual")
		}
	}

	diasPermitidos := parseCodigoDescuentoDiasSemanaSet(item.DiasSemana)
	if len(diasPermitidos) > 0 {
		if _, ok := diasPermitidos[isoWeekday(ctx.FechaOperacion)]; !ok {
			return "", fmt.Errorf("el codigo de descuento no aplica para el dia actual")
		}
	}

	segmentoDetectado := defaultCodigoDescuentoSegmento
	segmentoRegla := normalizeCodigoDescuentoSegmento(item.SegmentoCliente)
	if segmentoRegla != defaultCodigoDescuentoSegmento {
		if dbConn == nil {
			return "", fmt.Errorf("no se pudo validar segmento del cliente")
		}
		if ctx.ClienteID <= 0 {
			return "", fmt.Errorf("el codigo de descuento requiere cliente asociado para validar segmento")
		}
		perfil, err := GetClientePerfilComercialByEmpresa(dbConn, item.EmpresaID, ctx.ClienteID)
		if err != nil {
			if err == sql.ErrNoRows {
				return "", fmt.Errorf("cliente no encontrado para validar segmento de descuento")
			}
			return "", err
		}
		segmentoDetectado = normalizeCodigoDescuentoSegmento(perfil.Segmento)
		if segmentoDetectado == "" {
			segmentoDetectado = "nuevo"
		}
		if segmentoDetectado != segmentoRegla {
			return "", fmt.Errorf("el codigo de descuento no aplica para el segmento del cliente")
		}
	}

	return segmentoDetectado, nil
}

func validateCodigoDescuentoAntiFraudeContexto(dbConn *sql.DB, item *CodigoDescuento, ctx CodigoDescuentoContexto) error {
	if item == nil || dbConn == nil {
		return nil
	}
	ctx = resolveCodigoDescuentoContexto(ctx)

	if ctx.CarritoID > 0 {
		var reused int64
		err := dbConn.QueryRow(`SELECT COUNT(1)
			FROM codigos_descuento_redenciones
			WHERE empresa_id = ?
				AND codigo_descuento_id = ?
				AND carrito_id = ?`,
			item.EmpresaID,
			item.ID,
			ctx.CarritoID,
		).Scan(&reused)
		if err != nil {
			return err
		}
		if reused > 0 {
			return fmt.Errorf("el codigo de descuento ya fue aplicado en este carrito")
		}
	}

	var usosEmpresa int64
	err := dbConn.QueryRow(`SELECT COUNT(1)
		FROM codigos_descuento_redenciones
		WHERE empresa_id = ?
			AND codigo_descuento_id = ?`,
		item.EmpresaID,
		item.ID,
	).Scan(&usosEmpresa)
	if err != nil {
		return err
	}
	if usosEmpresa > 0 {
		return fmt.Errorf("codigo de descuento ya usado por esta empresa")
	}

	if item.MaxUsosPorCliente > 0 {
		if ctx.ClienteID <= 0 {
			return fmt.Errorf("el codigo de descuento requiere cliente para control antifraude")
		}
		ventana := item.VentanaHorasFraude
		if ventana <= 0 {
			ventana = defaultCodigoDescuentoVentanaHoras
		}
		desde := ctx.FechaOperacion.Add(-time.Duration(ventana) * time.Hour).Format("2006-01-02 15:04:05")
		var usosCliente int64
		err := dbConn.QueryRow(`SELECT COUNT(1)
			FROM codigos_descuento_redenciones
			WHERE empresa_id = ?
				AND codigo_descuento_id = ?
				AND cliente_id = ?
				AND COALESCE(fecha_redencion, fecha_creacion, '') >= ?`,
			item.EmpresaID,
			item.ID,
			ctx.ClienteID,
			desde,
		).Scan(&usosCliente)
		if err != nil {
			return err
		}
		if usosCliente >= item.MaxUsosPorCliente {
			return fmt.Errorf("limite de uso por cliente alcanzado para este codigo")
		}
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
	return ResolveCodigoDescuentoParaMontoConContexto(dbConn, empresaID, codigo, montoBase, CodigoDescuentoContexto{})
}

// ResolveCodigoDescuentoParaMontoConContexto valida un codigo de descuento incluyendo reglas por canal, segmento, horario y antifraude.
func ResolveCodigoDescuentoParaMontoConContexto(dbConn *sql.DB, empresaID int64, codigo string, montoBase float64, ctx CodigoDescuentoContexto) (*CodigoDescuentoAplicado, error) {
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
	segmentoDetectado, err := validateCodigoDescuentoReglasContexto(dbConn, item, ctx)
	if err != nil {
		return nil, err
	}
	if err := validateCodigoDescuentoAntiFraudeContexto(dbConn, item, ctx); err != nil {
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
		SegmentoCliente:  segmentoDetectado,
		CanalVenta:       normalizeCodigoDescuentoCanal(item.CanalVenta),
		FechaVencimiento: item.FechaVencimiento,
	}, nil
}

func resolveClienteSegmentoTx(tx *sql.Tx, empresaID, clienteID int64) string {
	if tx == nil || clienteID <= 0 {
		return "desconocido"
	}
	var compras int64
	var ultima sql.NullString
	err := queryRowTxSQLCompat(tx, `SELECT
		COALESCE(COUNT(1), 0),
		MAX(NULLIF(pagado_en, ''))
	FROM carritos_compras
	WHERE empresa_id = ?
		AND cliente_id = ?
		AND COALESCE(estado_carrito, 'abierto') = 'cerrado'`, empresaID, clienteID).Scan(&compras, &ultima)
	if err != nil {
		return "desconocido"
	}
	if compras <= 1 {
		return "nuevo"
	}
	if ultima.Valid {
		if tm, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(ultima.String), time.Local); err == nil {
			dias := int(time.Since(tm).Hours() / 24)
			if dias <= 30 {
				return "frecuente"
			}
			if dias <= 90 {
				return "recurrente"
			}
		}
	}
	return "en_riesgo"
}

func validateCodigoDescuentoAntiFraudeTx(tx *sql.Tx, item *CodigoDescuento, clienteID, carritoID int64, ahora time.Time) error {
	if tx == nil || item == nil {
		return nil
	}
	if carritoID > 0 {
		var reused int64
		err := queryRowTxSQLCompat(tx, `SELECT COUNT(1)
			FROM codigos_descuento_redenciones
			WHERE empresa_id = ?
				AND codigo_descuento_id = ?
				AND carrito_id = ?`,
			item.EmpresaID,
			item.ID,
			carritoID,
		).Scan(&reused)
		if err != nil {
			return err
		}
		if reused > 0 {
			return fmt.Errorf("el codigo de descuento ya fue aplicado en este carrito")
		}
	}

	var usosEmpresa int64
	err := queryRowTxSQLCompat(tx, `SELECT COUNT(1)
		FROM codigos_descuento_redenciones
		WHERE empresa_id = ?
			AND codigo_descuento_id = ?`,
		item.EmpresaID,
		item.ID,
	).Scan(&usosEmpresa)
	if err != nil {
		return err
	}
	if usosEmpresa > 0 {
		return fmt.Errorf("codigo de descuento ya usado por esta empresa")
	}

	if item.MaxUsosPorCliente > 0 {
		if clienteID <= 0 {
			return fmt.Errorf("el codigo de descuento requiere cliente para control antifraude")
		}
		ventana := item.VentanaHorasFraude
		if ventana <= 0 {
			ventana = defaultCodigoDescuentoVentanaHoras
		}
		desde := ahora.Add(-time.Duration(ventana) * time.Hour).Format("2006-01-02 15:04:05")
		var usosCliente int64
		err := queryRowTxSQLCompat(tx, `SELECT COUNT(1)
			FROM codigos_descuento_redenciones
			WHERE empresa_id = ?
				AND codigo_descuento_id = ?
				AND cliente_id = ?
				AND COALESCE(fecha_redencion, fecha_creacion, '') >= ?`,
			item.EmpresaID,
			item.ID,
			clienteID,
			desde,
		).Scan(&usosCliente)
		if err != nil {
			return err
		}
		if usosCliente >= item.MaxUsosPorCliente {
			return fmt.Errorf("limite de uso por cliente alcanzado para este codigo")
		}
	}
	return nil
}

func markCodigoDescuentoUsoTx(tx *sql.Tx, empresaID, codigoID, carritoID int64, valorAplicado float64, usuarioCreador, referenciaOperacion string) error {
	if tx == nil || codigoID <= 0 {
		return nil
	}
	ahora := time.Now()

	var item CodigoDescuento
	err := queryRowTxSQLCompat(tx, `SELECT
		id,
		empresa_id,
		COALESCE(codigo, ''),
		COALESCE(tipo_descuento, 'valor_fijo'),
		COALESCE(valor, 0),
		COALESCE(moneda, 'COP'),
		COALESCE(monto_minimo_compra, 0),
		COALESCE(segmento_cliente, 'todos'),
		COALESCE(canal_venta, 'todos'),
		COALESCE(horario_desde, ''),
		COALESCE(horario_hasta, ''),
		COALESCE(dias_semana, ''),
		COALESCE(max_usos_por_cliente, 0),
		COALESCE(ventana_horas_fraude, 24),
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
		&item.SegmentoCliente,
		&item.CanalVenta,
		&item.HorarioDesde,
		&item.HorarioHasta,
		&item.DiasSemana,
		&item.MaxUsosPorCliente,
		&item.VentanaHorasFraude,
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

	var clienteID int64
	var canalVenta string
	var montoBase float64
	err = queryRowTxSQLCompat(tx, `SELECT
		COALESCE(cliente_id, 0),
		COALESCE(canal_venta, 'mostrador'),
		COALESCE(total, 0)
	FROM carritos_compras
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, carritoID).Scan(&clienteID, &canalVenta, &montoBase)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("carrito no encontrado para aplicar codigo de descuento")
		}
		return err
	}

	if err := validateCodigoDescuentoAplicacion(&item, round2(montoBase)); err != nil {
		return err
	}
	if err := validateCodigoDescuentoAntiFraudeTx(tx, &item, clienteID, carritoID, ahora); err != nil {
		return err
	}

	res, err := execTxSQLCompat(tx, `UPDATE codigos_de_descuento
	SET usos_actuales = usos_actuales + 1,
		fecha_actualizacion = `+sqlNowExpr()+`
	WHERE empresa_id = ? AND id = ?
		AND usos_actuales < 1`, empresaID, codigoID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("codigo de descuento sin usos disponibles")
	}

	segmentoCliente := resolveClienteSegmentoTx(tx, empresaID, clienteID)
	if strings.TrimSpace(segmentoCliente) == "" {
		segmentoCliente = "desconocido"
	}
	_, err = execTxSQLCompat(tx, `INSERT INTO codigos_descuento_redenciones (
		empresa_id,
		codigo_descuento_id,
		carrito_id,
		cliente_id,
		codigo,
		canal_venta,
		segmento_cliente,
		monto_base,
		valor_descuento,
		estado_redencion,
		motivo,
		referencia_operacion,
		fecha_redencion,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'aplicada', ?, ?, `+sqlNowExpr()+`, `+sqlNowExpr()+`, `+sqlNowExpr()+`, ?, 'activo', ?)`,
		empresaID,
		codigoID,
		carritoID,
		clienteID,
		item.Codigo,
		normalizeCodigoDescuentoCanal(canalVenta),
		segmentoCliente,
		round2(montoBase),
		round2(valorAplicado),
		"uso_aplicado_en_cierre_de_carrito",
		strings.TrimSpace(referenciaOperacion),
		strings.TrimSpace(usuarioCreador),
		"redencion confirmada al cierre de carrito",
	)
	if err != nil {
		lower := strings.ToLower(strings.TrimSpace(err.Error()))
		if strings.Contains(lower, "unique") {
			return fmt.Errorf("el codigo de descuento ya fue aplicado en este carrito")
		}
		return err
	}
	return nil
}

func revertCodigoDescuentoUsoPorCarritoTx(tx *sql.Tx, empresaID, carritoID int64, estadoDestino, motivo, usuario string) error {
	if tx == nil || empresaID <= 0 || carritoID <= 0 {
		return nil
	}
	estadoDestino = strings.ToLower(strings.TrimSpace(estadoDestino))
	if estadoDestino != "anulada" && estadoDestino != "revertida" {
		estadoDestino = "revertida"
	}
	motivo = strings.TrimSpace(motivo)
	usuario = strings.TrimSpace(usuario)

	rows, err := queryTxSQLCompat(tx, `SELECT id, codigo_descuento_id
	FROM codigos_descuento_redenciones
	WHERE empresa_id = ?
		AND carrito_id = ?
		AND COALESCE(estado_redencion, 'aplicada') = 'aplicada'`, empresaID, carritoID)
	if err != nil {
		return err
	}

	type redencionDescuento struct {
		id       int64
		codigoID int64
	}
	redenciones := make([]redencionDescuento, 0)
	for rows.Next() {
		var redencion redencionDescuento
		if err := rows.Scan(&redencion.id, &redencion.codigoID); err != nil {
			_ = rows.Close()
			return err
		}
		redenciones = append(redenciones, redencion)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}

	for _, redencion := range redenciones {
		if _, err := execTxSQLCompat(tx, `UPDATE codigos_descuento_redenciones
		SET estado_redencion = ?,
			motivo = CASE
				WHEN trim(COALESCE(motivo, '')) = '' THEN ?
				ELSE trim(COALESCE(motivo, '')) || ' | ' || ?
			END,
			fecha_actualizacion = `+sqlNowExpr()+`,
			usuario_creador = CASE WHEN trim(COALESCE(usuario_creador, '')) = '' THEN ? ELSE usuario_creador END
		WHERE id = ?`, estadoDestino, motivo, motivo, usuario, redencion.id); err != nil {
			return err
		}
	}

	return nil
}

// ListCodigoDescuentoRedencionesByEmpresa lista la trazabilidad de redenciones por empresa.
func ListCodigoDescuentoRedencionesByEmpresa(dbConn *sql.DB, empresaID, codigoID int64, estado string, limit int) ([]CodigoDescuentoRedencion, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id invalido")
	}
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	estado = strings.ToLower(strings.TrimSpace(estado))

	query := `SELECT
		id,
		empresa_id,
		COALESCE(codigo_descuento_id, 0),
		COALESCE(carrito_id, 0),
		COALESCE(cliente_id, 0),
		COALESCE(codigo, ''),
		COALESCE(canal_venta, ''),
		COALESCE(segmento_cliente, ''),
		COALESCE(monto_base, 0),
		COALESCE(valor_descuento, 0),
		COALESCE(estado_redencion, 'aplicada'),
		COALESCE(motivo, ''),
		COALESCE(referencia_operacion, ''),
		COALESCE(fecha_redencion, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM codigos_descuento_redenciones
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	if codigoID > 0 {
		query += ` AND codigo_descuento_id = ?`
		args = append(args, codigoID)
	}
	if estado != "" {
		query += ` AND COALESCE(estado_redencion, 'aplicada') = ?`
		args = append(args, estado)
	}
	query += ` ORDER BY id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]CodigoDescuentoRedencion, 0)
	for rows.Next() {
		var item CodigoDescuentoRedencion
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.CodigoDescuentoID,
			&item.CarritoID,
			&item.ClienteID,
			&item.Codigo,
			&item.CanalVenta,
			&item.SegmentoCliente,
			&item.MontoBase,
			&item.ValorDescuento,
			&item.EstadoRedencion,
			&item.Motivo,
			&item.ReferenciaOperacion,
			&item.FechaRedencion,
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

	return out, rows.Err()
}
