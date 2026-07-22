package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrReservaHotelConflicto       = errors.New("reserva_hotel_conflicto")
	ErrReservaHotelExpirada        = errors.New("reserva_hotel_expirada")
	ErrReservaHotelNoReconvertible = errors.New("reserva_hotel_no_reconvertible")
	reservasHotelSchemaMu          sync.Mutex
	reservasHotelSchemaReady       bool
	reservasHotelPoliciesMu        sync.Mutex
	reservasHotelPoliciesLastRun   = map[int64]time.Time{}
)

const (
	reservaHotelExpiracionDefaultMin = 30
	reservaHotelNoShowToleranciaMin  = 90
	reservaHotelPoliciesCooldown     = 15 * time.Second
)

// ReservaHotel representa una reserva de estacion por empresa.
type ReservaHotel struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	CarritoID          int64   `json:"carrito_id"`
	EstacionID         int64   `json:"estacion_id"`
	EstacionCodigo     string  `json:"estacion_codigo,omitempty"`
	EstacionNombre     string  `json:"estacion_nombre,omitempty"`
	CodigoReserva      string  `json:"codigo_reserva"`
	ClienteNombre      string  `json:"cliente_nombre"`
	ClienteDocumento   string  `json:"cliente_documento,omitempty"`
	ClienteEmail       string  `json:"cliente_email,omitempty"`
	ClienteTelefono    string  `json:"cliente_telefono,omitempty"`
	CantidadHuespedes  int64   `json:"cantidad_huespedes"`
	FechaEntrada       string  `json:"fecha_entrada"`
	FechaSalida        string  `json:"fecha_salida"`
	MontoTotal         float64 `json:"monto_total"`
	Moneda             string  `json:"moneda"`
	EstadoReserva      string  `json:"estado_reserva"`
	EstadoPago         string  `json:"estado_pago"`
	ReferenciaPago     string  `json:"referencia_pago,omitempty"`
	PagoConfirmadoEn   string  `json:"pago_confirmado_en,omitempty"`
	FechaExpiracion    string  `json:"fecha_expiracion,omitempty"`
	ConfirmadoPor      string  `json:"confirmado_por,omitempty"`
	CanalOrigen        string  `json:"canal_origen,omitempty"`
	RequestID          string  `json:"request_id,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// ReservaHotelFilter define filtros para consultar reservas hoteleras por empresa.
type ReservaHotelFilter struct {
	EstacionID    int64
	EstadoReserva string
	EstadoPago    string
	Search        string
	FechaDesde    string
	FechaHasta    string
	Limit         int
	Offset        int
}

// ReservaHotelEstacion representa una estacion y su disponibilidad para un rango.
type ReservaHotelEstacion struct {
	CarritoID       int64  `json:"carrito_id"`
	EstacionID      int64  `json:"estacion_id"`
	EstacionCodigo  string `json:"estacion_codigo"`
	EstacionNombre  string `json:"estacion_nombre"`
	Estado          string `json:"estado"`
	EstadoCarrito   string `json:"estado_carrito"`
	Disponible      bool   `json:"disponible"`
	ReservasActivas int64  `json:"reservas_activas"`
}

// EnsureEmpresaReservasHotelSchema crea/migra el esquema de reservas de hotel por empresa.
func EnsureEmpresaReservasHotelSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("db connection is nil")
	}
	reservasHotelSchemaMu.Lock()
	defer reservasHotelSchemaMu.Unlock()

	if reservasHotelSchemaReady {
		return nil
	}
	ready, err := reservasHotelSchemaLooksReady(dbConn)
	if err == nil && ready {
		reservasHotelSchemaReady = true
		return nil
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS reservas_hotel (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			carrito_id INTEGER NOT NULL,
			estacion_id INTEGER NOT NULL,
			codigo_reserva TEXT NOT NULL,
			cliente_nombre TEXT NOT NULL,
			cliente_documento TEXT,
			cliente_email TEXT,
			cliente_telefono TEXT,
			cantidad_huespedes INTEGER DEFAULT 1,
			fecha_entrada TEXT NOT NULL,
			fecha_salida TEXT NOT NULL,
			monto_total REAL DEFAULT 0,
			moneda TEXT DEFAULT 'COP',
			estado_reserva TEXT DEFAULT 'pendiente_pago',
			estado_pago TEXT DEFAULT 'pendiente',
			referencia_pago TEXT,
			pago_confirmado_en TEXT,
			fecha_expiracion TEXT,
			confirmado_por TEXT,
			canal_origen TEXT DEFAULT 'web_reservas',
			request_id TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_reservas_hotel_empresa_codigo ON reservas_hotel(empresa_id, codigo_reserva);`,
		`CREATE INDEX IF NOT EXISTS ix_reservas_hotel_empresa_estado ON reservas_hotel(empresa_id, estado_reserva, estado_pago, estado);`,
		`CREATE INDEX IF NOT EXISTS ix_reservas_hotel_empresa_carrito_fechas ON reservas_hotel(empresa_id, carrito_id, fecha_entrada, fecha_salida);`,
		`CREATE INDEX IF NOT EXISTS ix_reservas_hotel_empresa_expiracion ON reservas_hotel(empresa_id, fecha_expiracion);`,
		`CREATE INDEX IF NOT EXISTS ix_reservas_hotel_empresa_estado_id ON reservas_hotel(empresa_id, estado, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_reservas_hotel_empresa_pendientes_fecha ON reservas_hotel(empresa_id, estado, estado_reserva, estado_pago, fecha_creacion, fecha_expiracion);`,
		`CREATE INDEX IF NOT EXISTS ix_reservas_hotel_empresa_confirmadas_entrada ON reservas_hotel(empresa_id, estado, estado_reserva, estado_pago, fecha_entrada);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "reservas_hotel", "carrito_id", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "reservas_hotel", "estacion_id", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "reservas_hotel", "codigo_reserva", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "reservas_hotel", "cliente_nombre", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "reservas_hotel", "cliente_documento", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "reservas_hotel", "cliente_email", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "reservas_hotel", "cliente_telefono", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "reservas_hotel", "cantidad_huespedes", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "reservas_hotel", "fecha_entrada", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "reservas_hotel", "fecha_salida", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "reservas_hotel", "monto_total", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "reservas_hotel", "moneda", "TEXT DEFAULT 'COP'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "reservas_hotel", "estado_reserva", "TEXT DEFAULT 'pendiente_pago'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "reservas_hotel", "estado_pago", "TEXT DEFAULT 'pendiente'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "reservas_hotel", "referencia_pago", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "reservas_hotel", "pago_confirmado_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "reservas_hotel", "fecha_expiracion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "reservas_hotel", "confirmado_por", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "reservas_hotel", "canal_origen", "TEXT DEFAULT 'web_reservas'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "reservas_hotel", "request_id", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "reservas_hotel", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "reservas_hotel", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "reservas_hotel", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "reservas_hotel", "observaciones", "TEXT"); err != nil {
		return err
	}
	reservasHotelSchemaReady = true
	return nil
}

// EmpresaReservasHotelSchemaReady verifies the table and indexes owned by the
// migrator without provisioning them while serving HTTP traffic.
func EmpresaReservasHotelSchemaReady(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("db connection is nil")
	}
	ready, err := reservasHotelSchemaLooksReady(dbConn)
	if err != nil {
		return fmt.Errorf("esquema de reservas no disponible: %w", err)
	}
	if !ready {
		return fmt.Errorf("esquema de reservas no disponible")
	}
	return nil
}

func reservasHotelSchemaLooksReady(dbConn *sql.DB) (bool, error) {
	ok, err := tableExists(dbConn, "reservas_hotel")
	if err != nil || !ok {
		return false, err
	}

	requiredIndexes := []string{
		"ux_reservas_hotel_empresa_codigo",
		"ix_reservas_hotel_empresa_estado",
		"ix_reservas_hotel_empresa_carrito_fechas",
		"ix_reservas_hotel_empresa_expiracion",
		"ix_reservas_hotel_empresa_estado_id",
		"ix_reservas_hotel_empresa_pendientes_fecha",
		"ix_reservas_hotel_empresa_confirmadas_entrada",
	}
	for _, indexName := range requiredIndexes {
		indexOK, idxErr := reservasHotelIndexExists(dbConn, indexName)
		if idxErr != nil || !indexOK {
			return false, idxErr
		}
	}

	return true, nil
}

func reservasHotelIndexExists(dbConn *sql.DB, indexName string) (bool, error) {
	var exists bool
	err := queryRowSQLCompat(dbConn, `
		SELECT EXISTS (
			SELECT 1
			FROM pg_indexes
			WHERE schemaname = ANY (current_schemas(false))
			  AND indexname = ?
		)
	`, indexName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func nextReservaHotelCodigo() string {
	return fmt.Sprintf("RSV-%d", time.Now().UnixNano())
}

func normalizeReservaHotelMoneda(v string) string {
	trim := strings.TrimSpace(strings.ToUpper(v))
	if trim == "" {
		return "COP"
	}
	return trim
}

func normalizeReservaHotelEstado(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "confirmada":
		return "confirmada"
	case "en_curso":
		return "en_curso"
	case "cancelada":
		return "cancelada"
	case "expirada":
		return "expirada"
	case "no_show", "noshow":
		return "no_show"
	case "pendiente_pago", "pendiente":
		return "pendiente_pago"
	default:
		return "pendiente_pago"
	}
}

func normalizeReservaHotelEstadoPago(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "confirmado":
		return "confirmado"
	case "cancelado":
		return "cancelado"
	case "expirado":
		return "expirado"
	case "pendiente":
		return "pendiente"
	default:
		return "pendiente"
	}
}

func parseReservaHotelpcs_ts(raw string) (time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, fmt.Errorf("fecha vacia")
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"2006-01-02",
	}
	for _, layout := range layouts {
		ts, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			return ts, nil
		}
	}
	return time.Time{}, fmt.Errorf("formato de fecha invalido: %s", value)
}

func normalizeReservaHotelDateRange(fechaEntrada, fechaSalida string) (string, string, error) {
	entrada, err := parseReservaHotelpcs_ts(fechaEntrada)
	if err != nil {
		return "", "", fmt.Errorf("fecha_entrada invalida")
	}
	salida, err := parseReservaHotelpcs_ts(fechaSalida)
	if err != nil {
		return "", "", fmt.Errorf("fecha_salida invalida")
	}
	if !salida.After(entrada) {
		return "", "", fmt.Errorf("fecha_salida debe ser mayor que fecha_entrada")
	}
	return entrada.Format("2006-01-02 15:04:05"), salida.Format("2006-01-02 15:04:05"), nil
}

func parseReservaHotelEstacionID(referenciaExterna, codigo string, empresaID int64) int64 {
	ref := strings.ToUpper(strings.TrimSpace(referenciaExterna))
	if strings.HasPrefix(ref, "ESTACION_") {
		n, err := strconv.ParseInt(strings.TrimPrefix(ref, "ESTACION_"), 10, 64)
		if err == nil && n > 0 {
			return n
		}
	}
	prefix := strings.ToUpper(fmt.Sprintf("EST-%d-", empresaID))
	code := strings.ToUpper(strings.TrimSpace(codigo))
	if strings.HasPrefix(code, prefix) {
		n, err := strconv.ParseInt(strings.TrimPrefix(code, prefix), 10, 64)
		if err == nil && n > 0 {
			return n
		}
	}
	return 0
}

func resolveReservaHotelStation(dbConn *sql.DB, empresaID, estacionID int64) (int64, string, string, string, error) {
	row := dbConn.QueryRow(`SELECT
		id,
		COALESCE(codigo, ''),
		COALESCE(nombre, ''),
		COALESCE(moneda, 'COP')
	FROM carritos_compras
	WHERE empresa_id = ?
	AND (
		upper(COALESCE(referencia_externa, '')) = upper(?)
		OR upper(COALESCE(codigo, '')) = upper(?)
	)
	ORDER BY id DESC
	LIMIT 1`,
		empresaID,
		fmt.Sprintf("ESTACION_%d", estacionID),
		fmt.Sprintf("EST-%d-%d", empresaID, estacionID),
	)

	var carritoID int64
	var codigo, nombre, moneda string
	if err := row.Scan(&carritoID, &codigo, &nombre, &moneda); err != nil {
		return 0, "", "", "", err
	}
	if strings.TrimSpace(nombre) == "" {
		nombre = fmt.Sprintf("Estacion %d", estacionID)
	}
	return carritoID, strings.TrimSpace(codigo), strings.TrimSpace(nombre), normalizeReservaHotelMoneda(moneda), nil
}

func hasReservaHotelConflict(dbConn *sql.DB, empresaID, estacionID, carritoID int64, fechaEntrada, fechaSalida string, ignoreID int64) (bool, error) {
	query := `SELECT COUNT(1)
	FROM reservas_hotel
	WHERE empresa_id = ?
		AND (
			estacion_id = ?
			OR carrito_id = ?
		)
		AND COALESCE(estado, 'activo') = 'activo'
		AND estado_reserva IN ('pendiente_pago', 'confirmada', 'en_curso')
		AND (
			estado_reserva <> 'pendiente_pago'
			OR COALESCE(fecha_expiracion, '') = ''
			OR pcs_ts(fecha_expiracion) > CURRENT_TIMESTAMP
		)
		AND pcs_ts(fecha_entrada) < pcs_ts(?)
		AND pcs_ts(fecha_salida) > pcs_ts(?)`
	args := []interface{}{empresaID, estacionID, carritoID, fechaSalida, fechaEntrada}
	if ignoreID > 0 {
		query += ` AND id <> ?`
		args = append(args, ignoreID)
	}

	var total int64
	if err := dbConn.QueryRow(query, args...).Scan(&total); err != nil {
		return false, err
	}
	return total > 0, nil
}

// ExpirePendientesReservasHotel marca como expiradas las reservas pendientes cuyo limite ya vencio.
func ExpirePendientesReservasHotel(dbConn *sql.DB, empresaID int64) (int64, error) {
	return expirePendientesReservasHotelAvanzado(dbConn, empresaID)
}

func expirePendientesReservasHotelAvanzado(dbConn *sql.DB, empresaID int64) (int64, error) {
	query := `UPDATE reservas_hotel
	SET
		estado_reserva = 'expirada',
		estado_pago = 'expirado',
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE COALESCE(estado, 'activo') = 'activo'
		AND estado_reserva = 'pendiente_pago'
		AND estado_pago = 'pendiente'
		AND (
			(
				COALESCE(fecha_expiracion, '') <> ''
				AND pcs_ts(fecha_expiracion) <= CURRENT_TIMESTAMP
			)
			OR (
				COALESCE(fecha_expiracion, '') = ''
				AND pcs_ts(fecha_creacion, '+' || ? || ' minutes') <= CURRENT_TIMESTAMP
			)
		)`
	args := []interface{}{strconv.Itoa(reservaHotelExpiracionDefaultMin)}
	if empresaID > 0 {
		query += ` AND empresa_id = ?`
		args = append(args, empresaID)
	}
	res, err := dbConn.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	count, _ := res.RowsAffected()
	return count, nil
}

func markReservasHotelNoShow(dbConn *sql.DB, empresaID int64, toleranciaMin int) (int64, error) {
	if toleranciaMin <= 0 {
		toleranciaMin = reservaHotelNoShowToleranciaMin
	}
	query := `UPDATE reservas_hotel
	SET
		estado_reserva = 'no_show',
		observaciones = CASE
			WHEN trim(COALESCE(observaciones, '')) = '' THEN 'no_show automatico por inasistencia'
			ELSE trim(observaciones) || ' | no_show automatico por inasistencia'
		END,
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE COALESCE(estado, 'activo') = 'activo'
		AND estado_reserva = 'confirmada'
		AND estado_pago = 'confirmado'
		AND pcs_ts(fecha_entrada, '+' || ? || ' minutes') <= CURRENT_TIMESTAMP`
	args := []interface{}{strconv.Itoa(toleranciaMin)}
	if empresaID > 0 {
		query += ` AND empresa_id = ?`
		args = append(args, empresaID)
	}
	res, err := dbConn.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	count, _ := res.RowsAffected()
	return count, nil
}

// ApplyReservasHotelOperationalPolicies aplica expiracion pendiente y politica no_show por empresa.
func ApplyReservasHotelOperationalPolicies(dbConn *sql.DB, empresaID int64) (int64, int64, error) {
	startedAt := time.Now()
	defer func() {
		PerfLogf("[perf][reservas] ApplyReservasHotelOperationalPolicies empresa=%d dur=%s", empresaID, time.Since(startedAt))
	}()
	if empresaID > 0 {
		reservasHotelPoliciesMu.Lock()
		lastRun := reservasHotelPoliciesLastRun[empresaID]
		if !lastRun.IsZero() && time.Since(lastRun) < reservaHotelPoliciesCooldown {
			reservasHotelPoliciesMu.Unlock()
			return 0, 0, nil
		}
		reservasHotelPoliciesLastRun[empresaID] = time.Now()
		reservasHotelPoliciesMu.Unlock()
	}

	expiradas, err := expirePendientesReservasHotelAvanzado(dbConn, empresaID)
	if err != nil {
		if empresaID > 0 {
			reservasHotelPoliciesMu.Lock()
			delete(reservasHotelPoliciesLastRun, empresaID)
			reservasHotelPoliciesMu.Unlock()
		}
		return 0, 0, err
	}
	noShow, err := markReservasHotelNoShow(dbConn, empresaID, reservaHotelNoShowToleranciaMin)
	if err != nil {
		if empresaID > 0 {
			reservasHotelPoliciesMu.Lock()
			delete(reservasHotelPoliciesLastRun, empresaID)
			reservasHotelPoliciesMu.Unlock()
		}
		return expiradas, 0, err
	}
	return expiradas, noShow, nil
}

// CreateReservaHotel crea una reserva para una estacion.
func CreateReservaHotel(dbConn *sql.DB, payload ReservaHotel) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	if payload.EstacionID <= 0 {
		return 0, fmt.Errorf("estacion_id es obligatorio")
	}
	if strings.TrimSpace(payload.ClienteNombre) == "" {
		return 0, fmt.Errorf("cliente_nombre es obligatorio")
	}

	fechaEntrada, fechaSalida, err := normalizeReservaHotelDateRange(payload.FechaEntrada, payload.FechaSalida)
	if err != nil {
		return 0, err
	}
	if _, _, err := ApplyReservasHotelOperationalPolicies(dbConn, payload.EmpresaID); err != nil {
		return 0, err
	}

	carritoID, estacionCodigo, estacionNombre, monedaCarrito, err := resolveReservaHotelStation(dbConn, payload.EmpresaID, payload.EstacionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("estacion no encontrada")
		}
		return 0, err
	}

	conflict, err := hasReservaHotelConflict(dbConn, payload.EmpresaID, payload.EstacionID, carritoID, fechaEntrada, fechaSalida, 0)
	if err != nil {
		return 0, err
	}
	if conflict {
		return 0, ErrReservaHotelConflicto
	}

	codigo := strings.TrimSpace(payload.CodigoReserva)
	if codigo == "" {
		codigo = nextReservaHotelCodigo()
	}

	cantidadHuespedes := payload.CantidadHuespedes
	if cantidadHuespedes <= 0 {
		cantidadHuespedes = 1
	}

	estadoReserva := normalizeReservaHotelEstado(payload.EstadoReserva)
	estadoPago := normalizeReservaHotelEstadoPago(payload.EstadoPago)
	if estadoReserva != "confirmada" {
		estadoReserva = "pendiente_pago"
		estadoPago = "pendiente"
	}

	montoTotal := payload.MontoTotal
	if montoTotal < 0 {
		montoTotal = 0
	}

	fechaExpiracion := strings.TrimSpace(payload.FechaExpiracion)
	if fechaExpiracion != "" {
		parsedExp, err := parseReservaHotelpcs_ts(fechaExpiracion)
		if err != nil {
			return 0, fmt.Errorf("fecha_expiracion invalida")
		}
		fechaExpiracion = parsedExp.Format("2006-01-02 15:04:05")
	}

	canalOrigen := strings.TrimSpace(payload.CanalOrigen)
	if canalOrigen == "" {
		canalOrigen = "web_reservas"
	}

	usuarioCreador := strings.TrimSpace(payload.UsuarioCreador)
	if usuarioCreador == "" {
		usuarioCreador = "portal_reservas"
	}

	id, err := insertSQLCompat(dbConn, `INSERT INTO reservas_hotel (
		empresa_id,
		carrito_id,
		estacion_id,
		codigo_reserva,
		cliente_nombre,
		cliente_documento,
		cliente_email,
		cliente_telefono,
		cantidad_huespedes,
		fecha_entrada,
		fecha_salida,
		monto_total,
		moneda,
		estado_reserva,
		estado_pago,
		referencia_pago,
		pago_confirmado_en,
		fecha_expiracion,
		confirmado_por,
		canal_origen,
		request_id,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '', NULL, COALESCE(NULLIF(?, ''), pcs_ts('now','localtime', '+30 minutes')), '', ?, ?, ?, 'activo', ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		payload.EmpresaID,
		carritoID,
		payload.EstacionID,
		codigo,
		strings.TrimSpace(payload.ClienteNombre),
		strings.TrimSpace(payload.ClienteDocumento),
		strings.TrimSpace(payload.ClienteEmail),
		strings.TrimSpace(payload.ClienteTelefono),
		cantidadHuespedes,
		fechaEntrada,
		fechaSalida,
		round2(montoTotal),
		normalizeReservaHotelMoneda(firstNonEmpty(payload.Moneda, monedaCarrito)),
		estadoReserva,
		estadoPago,
		fechaExpiracion,
		canalOrigen,
		strings.TrimSpace(payload.RequestID),
		usuarioCreador,
		strings.TrimSpace(payload.Observaciones),
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "ux_reservas_hotel_empresa_codigo") {
			return 0, fmt.Errorf("codigo_reserva duplicado")
		}
		return 0, err
	}

	if strings.TrimSpace(payload.EstacionCodigo) == "" {
		payload.EstacionCodigo = estacionCodigo
	}
	if strings.TrimSpace(payload.EstacionNombre) == "" {
		payload.EstacionNombre = estacionNombre
	}

	return id, nil
}

// UpdateReservaHotel actualiza una reserva pendiente de pago por empresa.
func UpdateReservaHotel(dbConn *sql.DB, payload ReservaHotel) error {
	if payload.EmpresaID <= 0 || payload.ID <= 0 {
		return fmt.Errorf("empresa_id e id son obligatorios")
	}
	if _, _, err := ApplyReservasHotelOperationalPolicies(dbConn, payload.EmpresaID); err != nil {
		return err
	}

	current, err := GetReservaHotelByID(dbConn, payload.EmpresaID, payload.ID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(strings.ToLower(current.Estado)) != "activo" {
		return fmt.Errorf("solo se permiten reservas activas")
	}
	if current.EstadoReserva != "pendiente_pago" || current.EstadoPago != "pendiente" {
		return fmt.Errorf("solo se pueden editar reservas pendientes de pago")
	}

	estacionID := payload.EstacionID
	if estacionID <= 0 {
		estacionID = current.EstacionID
	}
	if estacionID <= 0 {
		return fmt.Errorf("estacion_id es obligatorio")
	}

	carritoID, _, _, monedaCarrito, err := resolveReservaHotelStation(dbConn, payload.EmpresaID, estacionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("estacion no encontrada")
		}
		return err
	}

	fechaEntrada, fechaSalida, err := normalizeReservaHotelDateRange(
		firstNonEmpty(payload.FechaEntrada, current.FechaEntrada),
		firstNonEmpty(payload.FechaSalida, current.FechaSalida),
	)
	if err != nil {
		return err
	}

	conflict, err := hasReservaHotelConflict(dbConn, payload.EmpresaID, estacionID, carritoID, fechaEntrada, fechaSalida, payload.ID)
	if err != nil {
		return err
	}
	if conflict {
		return ErrReservaHotelConflicto
	}

	clienteNombre := strings.TrimSpace(firstNonEmpty(payload.ClienteNombre, current.ClienteNombre))
	if clienteNombre == "" {
		return fmt.Errorf("cliente_nombre es obligatorio")
	}

	cantidadHuespedes := payload.CantidadHuespedes
	if cantidadHuespedes <= 0 {
		cantidadHuespedes = current.CantidadHuespedes
	}
	if cantidadHuespedes <= 0 {
		cantidadHuespedes = 1
	}

	montoTotal := payload.MontoTotal
	if montoTotal < 0 {
		return fmt.Errorf("monto_total no puede ser negativo")
	}
	if payload.MontoTotal == 0 && current.MontoTotal > 0 {
		montoTotal = current.MontoTotal
	}

	fechaExpiracion := strings.TrimSpace(firstNonEmpty(payload.FechaExpiracion, current.FechaExpiracion))
	if fechaExpiracion == "" {
		fechaExpiracion = ""
	} else if parsedExp, err := parseReservaHotelpcs_ts(fechaExpiracion); err == nil {
		fechaExpiracion = parsedExp.Format("2006-01-02 15:04:05")
	} else {
		return fmt.Errorf("fecha_expiracion invalida")
	}

	res, err := dbConn.Exec(`UPDATE reservas_hotel
	SET
		carrito_id = ?,
		estacion_id = ?,
		codigo_reserva = ?,
		cliente_nombre = ?,
		cliente_documento = ?,
		cliente_email = ?,
		cliente_telefono = ?,
		cantidad_huespedes = ?,
		fecha_entrada = ?,
		fecha_salida = ?,
		monto_total = ?,
		moneda = ?,
		fecha_expiracion = COALESCE(NULLIF(?, ''), pcs_ts('now','localtime', '+30 minutes')),
		canal_origen = ?,
		request_id = ?,
		observaciones = ?,
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE empresa_id = ?
		AND id = ?
		AND COALESCE(estado, 'activo') = 'activo'`,
		carritoID,
		estacionID,
		firstNonEmpty(strings.TrimSpace(payload.CodigoReserva), strings.TrimSpace(current.CodigoReserva)),
		clienteNombre,
		strings.TrimSpace(firstNonEmpty(payload.ClienteDocumento, current.ClienteDocumento)),
		strings.TrimSpace(firstNonEmpty(payload.ClienteEmail, current.ClienteEmail)),
		strings.TrimSpace(firstNonEmpty(payload.ClienteTelefono, current.ClienteTelefono)),
		cantidadHuespedes,
		fechaEntrada,
		fechaSalida,
		round2(montoTotal),
		normalizeReservaHotelMoneda(firstNonEmpty(payload.Moneda, current.Moneda, monedaCarrito)),
		fechaExpiracion,
		strings.TrimSpace(firstNonEmpty(payload.CanalOrigen, current.CanalOrigen)),
		strings.TrimSpace(firstNonEmpty(payload.RequestID, current.RequestID)),
		strings.TrimSpace(firstNonEmpty(payload.Observaciones, current.Observaciones)),
		payload.EmpresaID,
		payload.ID,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "ux_reservas_hotel_empresa_codigo") {
			return fmt.Errorf("codigo_reserva duplicado")
		}
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SetReservaHotelEstado cambia estado activo/inactivo del registro de reserva.
func SetReservaHotelEstado(dbConn *sql.DB, empresaID, reservaID int64, estado string) error {
	if empresaID <= 0 || reservaID <= 0 {
		return fmt.Errorf("empresa_id e id son obligatorios")
	}
	nextEstado := strings.ToLower(strings.TrimSpace(estado))
	if nextEstado != "activo" && nextEstado != "inactivo" {
		nextEstado = "activo"
	}

	res, err := dbConn.Exec(`UPDATE reservas_hotel
	SET
		estado = ?,
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE empresa_id = ?
		AND id = ?`, nextEstado, empresaID, reservaID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteReservaHotel elimina una reserva por empresa.
func DeleteReservaHotel(dbConn *sql.DB, empresaID, reservaID int64) error {
	if empresaID <= 0 || reservaID <= 0 {
		return fmt.Errorf("empresa_id e id son obligatorios")
	}
	res, err := dbConn.Exec(`DELETE FROM reservas_hotel WHERE empresa_id = ? AND id = ?`, empresaID, reservaID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// CountReservasHotelByEmpresa cuenta reservas por empresa usando filtros operativos.
func CountReservasHotelByEmpresa(dbConn *sql.DB, empresaID int64, filter ReservaHotelFilter) (int64, error) {
	startedAt := time.Now()
	defer func() {
		PerfLogf("[perf][reservas] CountReservasHotelByEmpresa empresa=%d limit=%d offset=%d dur=%s", empresaID, filter.Limit, filter.Offset, time.Since(startedAt))
	}()
	if empresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	if _, _, err := ApplyReservasHotelOperationalPolicies(dbConn, empresaID); err != nil {
		return 0, err
	}
	return CountReservasHotelByEmpresaRaw(dbConn, empresaID, filter)
}

// CountReservasHotelByEmpresaRaw cuenta reservas por empresa sin reaplicar politicas operativas.
func CountReservasHotelByEmpresaRaw(dbConn *sql.DB, empresaID int64, filter ReservaHotelFilter) (int64, error) {
	if empresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	where, args := buildReservaHotelFilterClause(empresaID, filter)
	query := `SELECT COUNT(1)
	FROM reservas_hotel r
	WHERE ` + where

	var total int64
	if err := dbConn.QueryRow(query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

// ListReservasHotelByEmpresa lista reservas de hotel por empresa.
func ListReservasHotelByEmpresa(dbConn *sql.DB, empresaID int64, filter ReservaHotelFilter) ([]ReservaHotel, error) {
	startedAt := time.Now()
	defer func() {
		PerfLogf("[perf][reservas] ListReservasHotelByEmpresa empresa=%d limit=%d offset=%d dur=%s", empresaID, filter.Limit, filter.Offset, time.Since(startedAt))
	}()
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if _, _, err := ApplyReservasHotelOperationalPolicies(dbConn, empresaID); err != nil {
		return nil, err
	}
	return ListReservasHotelByEmpresaRaw(dbConn, empresaID, filter)
}

// ListReservasHotelByEmpresaRaw lista reservas de hotel por empresa SIN aplicar
// las politicas operativas automaticas (expiracion/no_show).
func ListReservasHotelByEmpresaRaw(dbConn *sql.DB, empresaID int64, filter ReservaHotelFilter) ([]ReservaHotel, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}

	if filter.Limit <= 0 {
		filter.Limit = 80
	}
	if filter.Limit > 500 {
		filter.Limit = 500
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	where, args := buildReservaHotelFilterClause(empresaID, filter)
	args = append(args, filter.Limit, filter.Offset)

	// #nosec G202 -- SQL structure is assembled only from server-side allowlists; all external values remain bound parameters.
	query := `SELECT
		r.id,
		r.empresa_id,
		r.carrito_id,
		r.estacion_id,
		COALESCE(c.codigo, ''),
		COALESCE(c.nombre, ''),
		COALESCE(r.codigo_reserva, ''),
		COALESCE(r.cliente_nombre, ''),
		COALESCE(r.cliente_documento, ''),
		COALESCE(r.cliente_email, ''),
		COALESCE(r.cliente_telefono, ''),
		COALESCE(r.cantidad_huespedes, 1),
		COALESCE(r.fecha_entrada, ''),
		COALESCE(r.fecha_salida, ''),
		COALESCE(r.monto_total, 0),
		COALESCE(r.moneda, 'COP'),
		COALESCE(r.estado_reserva, 'pendiente_pago'),
		COALESCE(r.estado_pago, 'pendiente'),
		COALESCE(r.referencia_pago, ''),
		COALESCE(r.pago_confirmado_en, ''),
		COALESCE(r.fecha_expiracion, ''),
		COALESCE(r.confirmado_por, ''),
		COALESCE(r.canal_origen, ''),
		COALESCE(r.request_id, ''),
		COALESCE(r.fecha_creacion, ''),
		COALESCE(r.fecha_actualizacion, ''),
		COALESCE(r.usuario_creador, ''),
		COALESCE(r.estado, 'activo'),
		COALESCE(r.observaciones, '')
	FROM reservas_hotel r
	LEFT JOIN carritos_compras c ON c.empresa_id = r.empresa_id AND c.id = r.carrito_id
	WHERE ` + where + `
	ORDER BY r.id DESC
	LIMIT ? OFFSET ?`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ReservaHotel, 0)
	for rows.Next() {
		var item ReservaHotel
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.CarritoID,
			&item.EstacionID,
			&item.EstacionCodigo,
			&item.EstacionNombre,
			&item.CodigoReserva,
			&item.ClienteNombre,
			&item.ClienteDocumento,
			&item.ClienteEmail,
			&item.ClienteTelefono,
			&item.CantidadHuespedes,
			&item.FechaEntrada,
			&item.FechaSalida,
			&item.MontoTotal,
			&item.Moneda,
			&item.EstadoReserva,
			&item.EstadoPago,
			&item.ReferenciaPago,
			&item.PagoConfirmadoEn,
			&item.FechaExpiracion,
			&item.ConfirmadoPor,
			&item.CanalOrigen,
			&item.RequestID,
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func buildReservaHotelFilterClause(empresaID int64, filter ReservaHotelFilter) (string, []interface{}) {
	where := []string{"r.empresa_id = ?", buildReservaHotelEstadoActivoClause("r.estado")}
	args := []interface{}{empresaID}

	if filter.EstacionID > 0 {
		where = append(where, "r.estacion_id = ?")
		args = append(args, filter.EstacionID)
	}
	if estadoReserva := strings.TrimSpace(strings.ToLower(filter.EstadoReserva)); estadoReserva != "" {
		where = append(where, buildReservaHotelExactTextClause("r.estado_reserva"))
		args = append(args, estadoReserva)
	}
	if estadoPago := strings.TrimSpace(strings.ToLower(filter.EstadoPago)); estadoPago != "" {
		where = append(where, buildReservaHotelExactTextClause("r.estado_pago"))
		args = append(args, estadoPago)
	}
	if strings.TrimSpace(filter.FechaDesde) != "" {
		if parsed, err := parseReservaHotelpcs_ts(filter.FechaDesde); err == nil {
			where = append(where, buildReservaHotelDateGTEClause("r.fecha_entrada"))
			args = append(args, parsed.Format("2006-01-02 15:04:05"))
		}
	}
	if strings.TrimSpace(filter.FechaHasta) != "" {
		if parsed, err := parseReservaHotelpcs_ts(filter.FechaHasta); err == nil {
			where = append(where, buildReservaHotelDateLTEClause("r.fecha_salida"))
			args = append(args, parsed.Format("2006-01-02 15:04:05"))
		}
	}
	search := strings.TrimSpace(filter.Search)
	if search != "" {
		pat := "%" + strings.ToLower(search) + "%"
		where = append(where, `(lower(COALESCE(r.codigo_reserva, '')) LIKE ? OR lower(COALESCE(r.cliente_nombre, '')) LIKE ? OR lower(COALESCE(r.cliente_documento, '')) LIKE ? OR lower(COALESCE(r.cliente_email, '')) LIKE ? OR lower(COALESCE(c.nombre, '')) LIKE ?)`)
		args = append(args, pat, pat, pat, pat, pat)
	}
	return strings.Join(where, " AND "), args
}

func buildReservaHotelEstadoActivoClause(column string) string {
	column = strings.TrimSpace(column)
	if column == "" {
		column = "r.estado"
	}
	if isPostgresDialect() {
		return "(" + column + " = 'activo' OR " + column + " IS NULL)"
	}
	return "COALESCE(" + column + ", 'activo') = 'activo'"
}

func buildReservaHotelExactTextClause(column string) string {
	column = strings.TrimSpace(column)
	if column == "" {
		column = "r.estado_reserva"
	}
	if isPostgresDialect() {
		return column + " = ?"
	}
	return "lower(COALESCE(" + column + ", '')) = ?"
}

func buildReservaHotelDateGTEClause(column string) string {
	column = strings.TrimSpace(column)
	if column == "" {
		column = "r.fecha_entrada"
	}
	if isPostgresDialect() {
		return column + " >= ?"
	}
	return "pcs_ts(" + column + ") >= pcs_ts(?)"
}

func buildReservaHotelDateLTEClause(column string) string {
	column = strings.TrimSpace(column)
	if column == "" {
		column = "r.fecha_salida"
	}
	if isPostgresDialect() {
		return column + " <= ?"
	}
	return "pcs_ts(" + column + ") <= pcs_ts(?)"
}

// GetReservaHotelByID obtiene una reserva puntual por empresa.
func GetReservaHotelByID(dbConn *sql.DB, empresaID, reservaID int64) (*ReservaHotel, error) {
	if empresaID <= 0 || reservaID <= 0 {
		return nil, fmt.Errorf("empresa_id e id son obligatorios")
	}
	if _, _, err := ApplyReservasHotelOperationalPolicies(dbConn, empresaID); err != nil {
		return nil, err
	}

	row := dbConn.QueryRow(`SELECT
		r.id,
		r.empresa_id,
		r.carrito_id,
		r.estacion_id,
		COALESCE(c.codigo, ''),
		COALESCE(c.nombre, ''),
		COALESCE(r.codigo_reserva, ''),
		COALESCE(r.cliente_nombre, ''),
		COALESCE(r.cliente_documento, ''),
		COALESCE(r.cliente_email, ''),
		COALESCE(r.cliente_telefono, ''),
		COALESCE(r.cantidad_huespedes, 1),
		COALESCE(r.fecha_entrada, ''),
		COALESCE(r.fecha_salida, ''),
		COALESCE(r.monto_total, 0),
		COALESCE(r.moneda, 'COP'),
		COALESCE(r.estado_reserva, 'pendiente_pago'),
		COALESCE(r.estado_pago, 'pendiente'),
		COALESCE(r.referencia_pago, ''),
		COALESCE(r.pago_confirmado_en, ''),
		COALESCE(r.fecha_expiracion, ''),
		COALESCE(r.confirmado_por, ''),
		COALESCE(r.canal_origen, ''),
		COALESCE(r.request_id, ''),
		COALESCE(r.fecha_creacion, ''),
		COALESCE(r.fecha_actualizacion, ''),
		COALESCE(r.usuario_creador, ''),
		COALESCE(r.estado, 'activo'),
		COALESCE(r.observaciones, '')
	FROM reservas_hotel r
	LEFT JOIN carritos_compras c ON c.empresa_id = r.empresa_id AND c.id = r.carrito_id
	WHERE r.empresa_id = ? AND r.id = ? AND COALESCE(r.estado, 'activo') = 'activo'
	LIMIT 1`, empresaID, reservaID)

	item := &ReservaHotel{}
	if err := row.Scan(
		&item.ID,
		&item.EmpresaID,
		&item.CarritoID,
		&item.EstacionID,
		&item.EstacionCodigo,
		&item.EstacionNombre,
		&item.CodigoReserva,
		&item.ClienteNombre,
		&item.ClienteDocumento,
		&item.ClienteEmail,
		&item.ClienteTelefono,
		&item.CantidadHuespedes,
		&item.FechaEntrada,
		&item.FechaSalida,
		&item.MontoTotal,
		&item.Moneda,
		&item.EstadoReserva,
		&item.EstadoPago,
		&item.ReferenciaPago,
		&item.PagoConfirmadoEn,
		&item.FechaExpiracion,
		&item.ConfirmadoPor,
		&item.CanalOrigen,
		&item.RequestID,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	); err != nil {
		return nil, err
	}
	return item, nil
}

// GetReservaHotelByCodigo obtiene una reserva por codigo de reserva.
func GetReservaHotelByCodigo(dbConn *sql.DB, empresaID int64, codigo string) (*ReservaHotel, error) {
	trimCode := strings.TrimSpace(codigo)
	if empresaID <= 0 || trimCode == "" {
		return nil, fmt.Errorf("empresa_id y codigo_reserva son obligatorios")
	}
	row := dbConn.QueryRow(`SELECT id FROM reservas_hotel WHERE empresa_id = ? AND codigo_reserva = ? AND COALESCE(estado, 'activo') = 'activo' LIMIT 1`, empresaID, trimCode)
	var id int64
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	return GetReservaHotelByID(dbConn, empresaID, id)
}

// ConfirmReservaHotelPago confirma el pago y la reserva.
func ConfirmReservaHotelPago(dbConn *sql.DB, empresaID, reservaID int64, referenciaPago, confirmadoPor, observaciones string) error {
	if empresaID <= 0 || reservaID <= 0 {
		return fmt.Errorf("empresa_id e id son obligatorios")
	}
	if _, _, err := ApplyReservasHotelOperationalPolicies(dbConn, empresaID); err != nil {
		return err
	}

	current, err := GetReservaHotelByID(dbConn, empresaID, reservaID)
	if err != nil {
		return err
	}
	if current.EstadoReserva == "expirada" {
		return ErrReservaHotelExpirada
	}
	if current.EstadoReserva != "pendiente_pago" || current.EstadoPago != "pendiente" {
		return fmt.Errorf("la reserva no esta pendiente de pago")
	}
	if strings.TrimSpace(current.FechaExpiracion) != "" {
		expiresAt, err := parseReservaHotelpcs_ts(current.FechaExpiracion)
		if err == nil && !expiresAt.After(time.Now()) {
			if _, expireErr := ExpirePendientesReservasHotel(dbConn, empresaID); expireErr != nil {
				return expireErr
			}
			return ErrReservaHotelExpirada
		}
	}

	conflict, err := hasReservaHotelConflict(dbConn, empresaID, current.EstacionID, current.CarritoID, current.FechaEntrada, current.FechaSalida, reservaID)
	if err != nil {
		return err
	}
	if conflict {
		return ErrReservaHotelConflicto
	}

	ref := strings.TrimSpace(referenciaPago)
	if ref == "" {
		ref = "confirmacion_manual"
	}
	by := strings.TrimSpace(confirmadoPor)
	if by == "" {
		by = "sistema"
	}

	_, err = dbConn.Exec(`UPDATE reservas_hotel
	SET
		estado_reserva = 'confirmada',
		estado_pago = 'confirmado',
		referencia_pago = ?,
		pago_confirmado_en = CURRENT_TIMESTAMP,
		confirmado_por = ?,
		observaciones = ?,
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE empresa_id = ? AND id = ? AND COALESCE(estado, 'activo') = 'activo'`,
		ref,
		by,
		strings.TrimSpace(observaciones),
		empresaID,
		reservaID,
	)
	return err
}

// ConvertReservaHotelToCarrito reconvierte una reserva confirmada a flujo de carrito operativo.
func ConvertReservaHotelToCarrito(dbConn *sql.DB, empresaID, reservaID int64, usuario string) (int64, error) {
	if empresaID <= 0 || reservaID <= 0 {
		return 0, fmt.Errorf("empresa_id e id son obligatorios")
	}
	if _, _, err := ApplyReservasHotelOperationalPolicies(dbConn, empresaID); err != nil {
		return 0, err
	}

	current, err := GetReservaHotelByID(dbConn, empresaID, reservaID)
	if err != nil {
		return 0, err
	}
	if !strings.EqualFold(strings.TrimSpace(current.Estado), "activo") {
		return 0, ErrReservaHotelNoReconvertible
	}

	estadoReserva := strings.ToLower(strings.TrimSpace(current.EstadoReserva))
	switch estadoReserva {
	case "en_curso":
		return current.CarritoID, nil
	case "confirmada":
		// estado valido para reconversion.
	case "pendiente_pago":
		return 0, fmt.Errorf("debe confirmar pago antes de reconvertir a carrito")
	default:
		return 0, ErrReservaHotelNoReconvertible
	}

	if current.CarritoID <= 0 {
		return 0, fmt.Errorf("carrito asociado no disponible")
	}

	by := strings.TrimSpace(usuario)
	if by == "" {
		by = "sistema"
	}
	obs := strings.TrimSpace(current.Observaciones)
	if obs == "" {
		obs = "reconversion a carrito operativa"
	} else {
		obs = obs + " | reconversion a carrito operativa"
	}

	tx, err := dbConn.Begin()
	if err != nil {
		return 0, err
	}

	resCarrito, err := tx.Exec(`UPDATE carritos_compras
	SET
		estado = 'activo',
		estado_carrito = 'abierto',
		activado_en = CASE
			WHEN trim(COALESCE(activado_en, '')) = '' THEN CURRENT_TIMESTAMP
			ELSE activado_en
		END,
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE empresa_id = ? AND id = ?`, empresaID, current.CarritoID)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	affectedCarrito, _ := resCarrito.RowsAffected()
	if affectedCarrito == 0 {
		_ = tx.Rollback()
		return 0, sql.ErrNoRows
	}

	resReserva, err := tx.Exec(`UPDATE reservas_hotel
	SET
		estado_reserva = 'en_curso',
		confirmado_por = ?,
		observaciones = ?,
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE empresa_id = ?
		AND id = ?
		AND COALESCE(estado, 'activo') = 'activo'`, by, obs, empresaID, reservaID)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	affectedReserva, _ := resReserva.RowsAffected()
	if affectedReserva == 0 {
		_ = tx.Rollback()
		return 0, sql.ErrNoRows
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return current.CarritoID, nil
}

// CancelReservaHotel cancela una reserva y libera la estacion para nuevas reservas.
func CancelReservaHotel(dbConn *sql.DB, empresaID, reservaID int64, motivo, usuario string) error {
	if empresaID <= 0 || reservaID <= 0 {
		return fmt.Errorf("empresa_id e id son obligatorios")
	}
	by := strings.TrimSpace(usuario)
	if by == "" {
		by = "sistema"
	}
	obs := strings.TrimSpace(motivo)
	if obs == "" {
		obs = "reserva cancelada"
	}
	_, err := dbConn.Exec(`UPDATE reservas_hotel
	SET
		estado_reserva = 'cancelada',
		estado_pago = CASE WHEN estado_pago = 'confirmado' THEN 'confirmado' ELSE 'cancelado' END,
		confirmado_por = ?,
		observaciones = ?,
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE empresa_id = ? AND id = ? AND COALESCE(estado, 'activo') = 'activo'`,
		by,
		obs,
		empresaID,
		reservaID,
	)
	return err
}

// ListReservasHotelEstacionesDisponibles lista estaciones y disponibilidad para un rango de fechas.
func ListReservasHotelEstacionesDisponibles(dbConn *sql.DB, empresaID int64, fechaEntrada, fechaSalida string) ([]ReservaHotelEstacion, error) {
	startedAt := time.Now()
	defer func() {
		PerfLogf("[perf][reservas] ListReservasHotelEstacionesDisponibles empresa=%d dur=%s", empresaID, time.Since(startedAt))
	}()
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	entradaNorm, salidaNorm, err := normalizeReservaHotelDateRange(fechaEntrada, fechaSalida)
	if err != nil {
		return nil, err
	}
	if _, _, err := ApplyReservasHotelOperationalPolicies(dbConn, empresaID); err != nil {
		return nil, err
	}

	prefixCode := strings.ToUpper(fmt.Sprintf("EST-%d-%%", empresaID))
	rows, err := dbConn.Query(`SELECT
		c.id,
		COALESCE(c.codigo, ''),
		COALESCE(c.nombre, ''),
		COALESCE(c.referencia_externa, ''),
		COALESCE(c.estado, 'activo'),
		COALESCE(c.estado_carrito, 'abierto'),
		COALESCE(SUM(
			CASE
				WHEN r.id IS NULL THEN 0
				ELSE 1
			END
		), 0) AS reservas_activas
	FROM carritos_compras c
	LEFT JOIN reservas_hotel r ON r.empresa_id = c.empresa_id
		AND r.carrito_id = c.id
		AND COALESCE(r.estado, 'activo') = 'activo'
		AND r.estado_reserva IN ('pendiente_pago', 'confirmada', 'en_curso')
		AND (
			r.estado_reserva <> 'pendiente_pago'
			OR COALESCE(r.fecha_expiracion, '') = ''
			OR pcs_ts(r.fecha_expiracion) > CURRENT_TIMESTAMP
		)
		AND pcs_ts(r.fecha_entrada) < pcs_ts(?)
		AND pcs_ts(r.fecha_salida) > pcs_ts(?)
	WHERE c.empresa_id = ?
		AND (
			upper(COALESCE(c.referencia_externa, '')) LIKE 'ESTACION_%'
			OR upper(COALESCE(c.codigo, '')) LIKE ?
		)
	GROUP BY c.id, c.codigo, c.nombre, c.referencia_externa, c.estado, c.estado_carrito
	ORDER BY c.id ASC`, salidaNorm, entradaNorm, empresaID, prefixCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ReservaHotelEstacion, 0)
	for rows.Next() {
		var item ReservaHotelEstacion
		var referenciaExterna string
		if err := rows.Scan(
			&item.CarritoID,
			&item.EstacionCodigo,
			&item.EstacionNombre,
			&referenciaExterna,
			&item.Estado,
			&item.EstadoCarrito,
			&item.ReservasActivas,
		); err != nil {
			return nil, err
		}
		item.EstacionID = parseReservaHotelEstacionID(referenciaExterna, item.EstacionCodigo, empresaID)
		item.Disponible = strings.EqualFold(item.Estado, "activo") && item.ReservasActivas == 0
		if strings.TrimSpace(item.EstacionNombre) == "" {
			if item.EstacionID > 0 {
				item.EstacionNombre = fmt.Sprintf("Estacion %d", item.EstacionID)
			} else {
				item.EstacionNombre = fmt.Sprintf("Estacion %d", item.CarritoID)
			}
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		trim := strings.TrimSpace(v)
		if trim != "" {
			return trim
		}
	}
	return ""
}
