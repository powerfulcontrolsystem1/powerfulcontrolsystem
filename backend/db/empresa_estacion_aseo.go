package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

const empresaEstacionEstadoSuciaClave = "estacion_estado_sucia"

// EmpresaEstacionAseoEvento registra el ciclo de aseo de una estacion.
type EmpresaEstacionAseoEvento struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	EstacionID         int64  `json:"estacion_id"`
	EstacionNombre     string `json:"estacion_nombre,omitempty"`
	SuciaDesde         string `json:"sucia_desde,omitempty"`
	AseoFin            string `json:"aseo_fin,omitempty"`
	DuracionSegundos   int64  `json:"duracion_segundos"`
	UsuarioID          int64  `json:"usuario_id,omitempty"`
	UsuarioEmail       string `json:"usuario_email,omitempty"`
	UsuarioNombre      string `json:"usuario_nombre,omitempty"`
	RolNombre          string `json:"rol_nombre,omitempty"`
	Origen             string `json:"origen,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
}

type EmpresaEstacionAseoFiltro struct {
	EmpresaID  int64
	EstacionID int64
	UsuarioID  int64
	Desde      string
	Hasta      string
	Estado     string
	Limit      int
}

type EmpresaEstacionAseoFinalizarInput struct {
	EmpresaID      int64
	EstacionID     int64
	EstacionNombre string
	UsuarioID      int64
	UsuarioEmail   string
	UsuarioNombre  string
	RolNombre      string
	Observaciones  string
	Origen         string
}

func EnsureEmpresaEstacionAseoSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_estacion_aseo_eventos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			estacion_id BIGINT NOT NULL,
			estacion_nombre TEXT,
			sucia_desde TEXT NOT NULL,
			aseo_fin TEXT,
			duracion_segundos BIGINT DEFAULT 0,
			usuario_id BIGINT DEFAULT 0,
			usuario_email TEXT,
			usuario_nombre TEXT,
			rol_nombre TEXT,
			origen TEXT,
			estado TEXT DEFAULT 'pendiente',
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime'))
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_estacion_aseo_empresa_estacion_estado ON empresa_estacion_aseo_eventos(empresa_id, estacion_id, estado);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_estacion_aseo_empresa_fin ON empresa_estacion_aseo_eventos(empresa_id, aseo_fin);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_estacion_aseo_empresa_usuario ON empresa_estacion_aseo_eventos(empresa_id, usuario_id);`,
	}
	for _, stmt := range stmts {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	columns := []struct {
		name string
		def  string
	}{
		{"estacion_nombre", "TEXT"},
		{"sucia_desde", "TEXT"},
		{"aseo_fin", "TEXT"},
		{"duracion_segundos", "BIGINT DEFAULT 0"},
		{"usuario_id", "BIGINT DEFAULT 0"},
		{"usuario_email", "TEXT"},
		{"usuario_nombre", "TEXT"},
		{"rol_nombre", "TEXT"},
		{"origen", "TEXT"},
		{"estado", "TEXT DEFAULT 'pendiente'"},
		{"observaciones", "TEXT"},
		{"fecha_creacion", "TEXT DEFAULT (datetime('now','localtime'))"},
		{"fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"},
	}
	for _, column := range columns {
		if err := ensureColumnIfMissing(dbConn, "empresa_estacion_aseo_eventos", column.name, column.def); err != nil {
			return err
		}
	}
	return nil
}

func IsEmpresaEstacionDirtyValue(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "sucia", "dirty", "si", "sí":
		return true
	default:
		return false
	}
}

func ResolveEmpresaEstacionNombre(dbConn *sql.DB, empresaID, estacionID int64) string {
	if empresaID <= 0 || estacionID <= 0 {
		return ""
	}
	pref, err := GetEmpresaEstacionPref(dbConn, empresaID, 0, "estaciones_config")
	if err != nil || pref == nil {
		return ""
	}
	cfg, err := parseEmpresaEstacionesConfig(pref.Valor)
	if err != nil || cfg == nil {
		return ""
	}
	for _, est := range cfg.Estaciones {
		if est.ID == estacionID {
			return strings.TrimSpace(est.Nombre)
		}
	}
	return ""
}

func StartEmpresaEstacionAseoEvento(dbConn *sql.DB, empresaID, estacionID int64, estacionNombre, suciaDesde, usuario string) (int64, error) {
	if empresaID <= 0 || estacionID <= 0 {
		return 0, errors.New("empresa_id y estacion_id son obligatorios")
	}
	if err := EnsureEmpresaEstacionAseoSchema(dbConn); err != nil {
		return 0, err
	}
	estacionNombre = strings.TrimSpace(estacionNombre)
	if estacionNombre == "" {
		estacionNombre = ResolveEmpresaEstacionNombre(dbConn, empresaID, estacionID)
	}
	startedAt := normalizeEmpresaAseoTimestamp(suciaDesde)
	if startedAt == "" {
		startedAt = time.Now().Format("2006-01-02 15:04:05")
	}
	existingID, err := getPendingEmpresaEstacionAseoID(dbConn, empresaID, estacionID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	if existingID > 0 {
		_, err := execSQLCompat(dbConn, `UPDATE empresa_estacion_aseo_eventos
			SET estacion_nombre = COALESCE(NULLIF(?, ''), estacion_nombre),
				sucia_desde = ?,
				origen = COALESCE(NULLIF(?, ''), origen),
				fecha_actualizacion = datetime('now','localtime')
			WHERE id = ? AND empresa_id = ?`,
			estacionNombre,
			startedAt,
			strings.TrimSpace(usuario),
			existingID,
			empresaID,
		)
		return existingID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_estacion_aseo_eventos (
		empresa_id, estacion_id, estacion_nombre, sucia_desde, origen, estado, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, 'pendiente', datetime('now','localtime'), datetime('now','localtime'))`,
		empresaID,
		estacionID,
		estacionNombre,
		startedAt,
		strings.TrimSpace(usuario),
	)
}

func FinalizarEmpresaEstacionAseo(dbConn *sql.DB, input EmpresaEstacionAseoFinalizarInput) (*EmpresaEstacionAseoEvento, error) {
	if input.EmpresaID <= 0 || input.EstacionID <= 0 {
		return nil, errors.New("empresa_id y estacion_id son obligatorios")
	}
	if err := EnsureEmpresaEstacionPrefsSchema(dbConn); err != nil {
		return nil, err
	}
	if err := EnsureEmpresaEstacionAseoSchema(dbConn); err != nil {
		return nil, err
	}
	pref, err := GetEmpresaEstacionPref(dbConn, input.EmpresaID, input.EstacionID, empresaEstacionEstadoSuciaClave)
	if err != nil {
		return nil, err
	}
	if pref == nil || !IsEmpresaEstacionDirtyValue(pref.Valor) {
		return nil, fmt.Errorf("la estacion no esta marcada como sucia")
	}
	evento, err := getPendingEmpresaEstacionAseo(dbConn, input.EmpresaID, input.EstacionID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if evento == nil {
		nombre := strings.TrimSpace(input.EstacionNombre)
		if nombre == "" {
			nombre = ResolveEmpresaEstacionNombre(dbConn, input.EmpresaID, input.EstacionID)
		}
		id, err := StartEmpresaEstacionAseoEvento(dbConn, input.EmpresaID, input.EstacionID, nombre, pref.FechaActualizacion, pref.UsuarioCreador)
		if err != nil {
			return nil, err
		}
		evento, err = getEmpresaEstacionAseoByID(dbConn, input.EmpresaID, id)
		if err != nil {
			return nil, err
		}
	}

	fin := time.Now()
	inicio, ok := parseEmpresaAseoTimestamp(evento.SuciaDesde)
	if !ok {
		inicio = fin
	}
	duracion := int64(fin.Sub(inicio).Seconds())
	if duracion < 0 {
		duracion = 0
	}
	usuarioEmail := strings.TrimSpace(input.UsuarioEmail)
	usuarioNombre := strings.TrimSpace(input.UsuarioNombre)
	rolNombre := strings.TrimSpace(input.RolNombre)
	origen := strings.TrimSpace(input.Origen)
	if origen == "" {
		origen = "estaciones"
	}
	observaciones := strings.TrimSpace(input.Observaciones)
	finStr := fin.Format("2006-01-02 15:04:05")
	_, err = execSQLCompat(dbConn, `UPDATE empresa_estacion_aseo_eventos
		SET estacion_nombre = COALESCE(NULLIF(?, ''), estacion_nombre),
			aseo_fin = ?,
			duracion_segundos = ?,
			usuario_id = ?,
			usuario_email = ?,
			usuario_nombre = ?,
			rol_nombre = ?,
			origen = ?,
			estado = 'finalizado',
			observaciones = ?,
			fecha_actualizacion = datetime('now','localtime')
		WHERE id = ? AND empresa_id = ?`,
		strings.TrimSpace(input.EstacionNombre),
		finStr,
		duracion,
		input.UsuarioID,
		usuarioEmail,
		usuarioNombre,
		rolNombre,
		origen,
		observaciones,
		evento.ID,
		input.EmpresaID,
	)
	if err != nil {
		return nil, err
	}
	_, err = UpsertEmpresaEstacionPref(dbConn, EmpresaEstacionPref{
		EmpresaID:      input.EmpresaID,
		EstacionID:     input.EstacionID,
		Clave:          empresaEstacionEstadoSuciaClave,
		Valor:          "0",
		UsuarioCreador: usuarioEmail,
		Estado:         "activo",
		Observaciones:  "aseo reportado",
	})
	if err != nil {
		return nil, err
	}
	return getEmpresaEstacionAseoByID(dbConn, input.EmpresaID, evento.ID)
}

func ListEmpresaEstacionAseoEventos(dbConn *sql.DB, filtro EmpresaEstacionAseoFiltro) ([]EmpresaEstacionAseoEvento, error) {
	if filtro.EmpresaID <= 0 {
		return nil, errors.New("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaEstacionAseoSchema(dbConn); err != nil {
		return nil, err
	}
	where := []string{"empresa_id = ?"}
	args := []interface{}{filtro.EmpresaID}
	if filtro.EstacionID > 0 {
		where = append(where, "estacion_id = ?")
		args = append(args, filtro.EstacionID)
	}
	if filtro.UsuarioID > 0 {
		where = append(where, "usuario_id = ?")
		args = append(args, filtro.UsuarioID)
	}
	estado := strings.ToLower(strings.TrimSpace(filtro.Estado))
	if estado == "" {
		estado = "finalizado"
	}
	if estado != "todos" {
		where = append(where, "LOWER(COALESCE(estado, '')) = ?")
		args = append(args, estado)
	}
	if strings.TrimSpace(filtro.Desde) != "" {
		where = append(where, "COALESCE(NULLIF(aseo_fin, ''), sucia_desde) >= ?")
		args = append(args, strings.TrimSpace(filtro.Desde))
	}
	if strings.TrimSpace(filtro.Hasta) != "" {
		where = append(where, "COALESCE(NULLIF(aseo_fin, ''), sucia_desde) <= ?")
		args = append(args, strings.TrimSpace(filtro.Hasta))
	}
	limit := filtro.Limit
	if limit <= 0 || limit > 1000 {
		limit = 500
	}
	args = append(args, limit)
	rows, err := querySQLCompat(dbConn, `SELECT
		id, empresa_id, estacion_id, COALESCE(estacion_nombre, ''), COALESCE(sucia_desde, ''),
		COALESCE(aseo_fin, ''), COALESCE(duracion_segundos, 0), COALESCE(usuario_id, 0),
		COALESCE(usuario_email, ''), COALESCE(usuario_nombre, ''), COALESCE(rol_nombre, ''),
		COALESCE(origen, ''), COALESCE(estado, ''), COALESCE(observaciones, ''),
		COALESCE(fecha_creacion, ''), COALESCE(fecha_actualizacion, '')
	FROM empresa_estacion_aseo_eventos
	WHERE `+strings.Join(where, " AND ")+`
	ORDER BY COALESCE(NULLIF(aseo_fin, ''), sucia_desde) DESC, id DESC
	LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaEstacionAseoEvento, 0)
	for rows.Next() {
		var item EmpresaEstacionAseoEvento
		if err := scanEmpresaEstacionAseo(rows.Scan, &item); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func getPendingEmpresaEstacionAseoID(dbConn *sql.DB, empresaID, estacionID int64) (int64, error) {
	row := queryRowSQLCompat(dbConn, `SELECT id FROM empresa_estacion_aseo_eventos
		WHERE empresa_id = ? AND estacion_id = ? AND LOWER(COALESCE(estado, 'pendiente')) = 'pendiente'
		ORDER BY id DESC LIMIT 1`, empresaID, estacionID)
	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func getPendingEmpresaEstacionAseo(dbConn *sql.DB, empresaID, estacionID int64) (*EmpresaEstacionAseoEvento, error) {
	row := queryRowSQLCompat(dbConn, `SELECT
		id, empresa_id, estacion_id, COALESCE(estacion_nombre, ''), COALESCE(sucia_desde, ''),
		COALESCE(aseo_fin, ''), COALESCE(duracion_segundos, 0), COALESCE(usuario_id, 0),
		COALESCE(usuario_email, ''), COALESCE(usuario_nombre, ''), COALESCE(rol_nombre, ''),
		COALESCE(origen, ''), COALESCE(estado, ''), COALESCE(observaciones, ''),
		COALESCE(fecha_creacion, ''), COALESCE(fecha_actualizacion, '')
	FROM empresa_estacion_aseo_eventos
	WHERE empresa_id = ? AND estacion_id = ? AND LOWER(COALESCE(estado, 'pendiente')) = 'pendiente'
	ORDER BY id DESC LIMIT 1`, empresaID, estacionID)
	var item EmpresaEstacionAseoEvento
	if err := scanEmpresaEstacionAseo(row.Scan, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func getEmpresaEstacionAseoByID(dbConn *sql.DB, empresaID, id int64) (*EmpresaEstacionAseoEvento, error) {
	row := queryRowSQLCompat(dbConn, `SELECT
		id, empresa_id, estacion_id, COALESCE(estacion_nombre, ''), COALESCE(sucia_desde, ''),
		COALESCE(aseo_fin, ''), COALESCE(duracion_segundos, 0), COALESCE(usuario_id, 0),
		COALESCE(usuario_email, ''), COALESCE(usuario_nombre, ''), COALESCE(rol_nombre, ''),
		COALESCE(origen, ''), COALESCE(estado, ''), COALESCE(observaciones, ''),
		COALESCE(fecha_creacion, ''), COALESCE(fecha_actualizacion, '')
	FROM empresa_estacion_aseo_eventos
	WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, id)
	var item EmpresaEstacionAseoEvento
	if err := scanEmpresaEstacionAseo(row.Scan, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

type empresaEstacionAseoScanner func(dest ...interface{}) error

func scanEmpresaEstacionAseo(scan empresaEstacionAseoScanner, item *EmpresaEstacionAseoEvento) error {
	return scan(
		&item.ID,
		&item.EmpresaID,
		&item.EstacionID,
		&item.EstacionNombre,
		&item.SuciaDesde,
		&item.AseoFin,
		&item.DuracionSegundos,
		&item.UsuarioID,
		&item.UsuarioEmail,
		&item.UsuarioNombre,
		&item.RolNombre,
		&item.Origen,
		&item.Estado,
		&item.Observaciones,
		&item.FechaCreacion,
		&item.FechaActualizacion,
	)
}

func normalizeEmpresaAseoTimestamp(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if parsed, ok := parseEmpresaAseoTimestamp(raw); ok {
		return parsed.Format("2006-01-02 15:04:05")
	}
	return raw
}

func parseEmpresaAseoTimestamp(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.999999",
		"2006-01-02T15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return parsed, true
		}
	}
	if len(raw) >= len("2006-01-02 15:04:05") {
		short := strings.ReplaceAll(raw[:len("2006-01-02 15:04:05")], "T", " ")
		if parsed, err := time.ParseInLocation("2006-01-02 15:04:05", short, time.Local); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}
