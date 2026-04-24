package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// EmpresaGPSDispositivo representa un dispositivo GPS registrado por empresa.
type EmpresaGPSDispositivo struct {
	ID                    int64   `json:"id"`
	EmpresaID             int64   `json:"empresa_id"`
	Codigo                string  `json:"codigo"`
	Nombre                string  `json:"nombre"`
	Descripcion           string  `json:"descripcion,omitempty"`
	UltimaLatitud         float64 `json:"ultima_latitud,omitempty"`
	UltimaLongitud        float64 `json:"ultima_longitud,omitempty"`
	UltimaPrecisionMetros float64 `json:"ultima_precision_metros,omitempty"`
	UltimaVelocidadKMH    float64 `json:"ultima_velocidad_kmh,omitempty"`
	UltimoReporteEn       string  `json:"ultimo_reporte_en,omitempty"`
	FechaCreacion         string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion    string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador        string  `json:"usuario_creador,omitempty"`
	Estado                string  `json:"estado,omitempty"`
	Observaciones         string  `json:"observaciones,omitempty"`
}

// EmpresaGPSRecorrido representa un punto de recorrido para un dispositivo GPS.
type EmpresaGPSRecorrido struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	DispositivoID      int64   `json:"dispositivo_id"`
	Latitud            float64 `json:"latitud"`
	Longitud           float64 `json:"longitud"`
	PrecisionMetros    float64 `json:"precision_metros,omitempty"`
	VelocidadKMH       float64 `json:"velocidad_kmh,omitempty"`
	RumboGrados        float64 `json:"rumbo_grados,omitempty"`
	AltitudMetros      float64 `json:"altitud_metros,omitempty"`
	Fuente             string  `json:"fuente,omitempty"`
	CapturadoEn        string  `json:"capturado_en,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// EnsureEmpresaUbicacionGPSSchema crea y migra las tablas del modulo de ubicacion GPS por empresa.
func EnsureEmpresaUbicacionGPSSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_gps_dispositivos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			descripcion TEXT,
			ultima_latitud REAL,
			ultima_longitud REAL,
			ultima_precision_metros REAL DEFAULT 0,
			ultima_velocidad_kmh REAL DEFAULT 0,
			ultimo_reporte_en TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE TABLE IF NOT EXISTS empresa_gps_recorridos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			dispositivo_id INTEGER NOT NULL,
			latitud REAL NOT NULL,
			longitud REAL NOT NULL,
			precision_metros REAL DEFAULT 0,
			velocidad_kmh REAL DEFAULT 0,
			rumbo_grados REAL DEFAULT 0,
			altitud_metros REAL DEFAULT 0,
			fuente TEXT DEFAULT 'manual',
			capturado_en TEXT DEFAULT (datetime('now','localtime')),
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_gps_dispositivos_empresa_estado ON empresa_gps_dispositivos(empresa_id, estado, nombre);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_gps_dispositivos_empresa_codigo ON empresa_gps_dispositivos(empresa_id, codigo);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_gps_recorridos_empresa_dispositivo_fecha ON empresa_gps_recorridos(empresa_id, dispositivo_id, capturado_en);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_gps_recorridos_empresa_fecha ON empresa_gps_recorridos(empresa_id, capturado_en);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_gps_dispositivos", "codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_dispositivos", "nombre", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_dispositivos", "descripcion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_dispositivos", "ultima_latitud", "REAL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_dispositivos", "ultima_longitud", "REAL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_dispositivos", "ultima_precision_metros", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_dispositivos", "ultima_velocidad_kmh", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_dispositivos", "ultimo_reporte_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_dispositivos", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_dispositivos", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_dispositivos", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_dispositivos", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_gps_recorridos", "precision_metros", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_recorridos", "velocidad_kmh", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_recorridos", "rumbo_grados", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_recorridos", "altitud_metros", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_recorridos", "fuente", "TEXT DEFAULT 'manual'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_recorridos", "capturado_en", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_recorridos", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_recorridos", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_recorridos", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_recorridos", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

// CreateEmpresaGPSDispositivo crea un dispositivo GPS para una empresa.
func CreateEmpresaGPSDispositivo(dbConn *sql.DB, d EmpresaGPSDispositivo) (int64, error) {
	codigo := sanitizeGPSCode(d.Codigo)
	if codigo == "" {
		codigo = defaultGPSCode(d.EmpresaID, d.Nombre)
	}

	res, err := dbConn.Exec(`INSERT INTO empresa_gps_dispositivos (
		empresa_id, codigo, nombre, descripcion,
		usuario_creador, estado, observaciones,
		fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, COALESCE(NULLIF(?, ''), 'activo'), ?, datetime('now','localtime'), datetime('now','localtime'))`,
		d.EmpresaID,
		codigo,
		strings.TrimSpace(d.Nombre),
		strings.TrimSpace(d.Descripcion),
		strings.TrimSpace(d.UsuarioCreador),
		strings.TrimSpace(d.Estado),
		strings.TrimSpace(d.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetEmpresaGPSDispositivos lista dispositivos GPS por empresa.
func GetEmpresaGPSDispositivos(dbConn *sql.DB, empresaID int64, includeInactive bool, q string) ([]EmpresaGPSDispositivo, error) {
	query := `SELECT
		id, empresa_id, codigo, nombre, COALESCE(descripcion, ''),
		COALESCE(ultima_latitud, 0), COALESCE(ultima_longitud, 0),
		COALESCE(ultima_precision_metros, 0), COALESCE(ultima_velocidad_kmh, 0),
		COALESCE(ultimo_reporte_en, ''), COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''), COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'), COALESCE(observaciones, '')
	FROM empresa_gps_dispositivos
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	if !includeInactive {
		query += ` AND estado = 'activo'`
	}
	q = strings.TrimSpace(q)
	if q != "" {
		query += ` AND (LOWER(codigo) LIKE ? OR LOWER(nombre) LIKE ? OR LOWER(COALESCE(descripcion,'')) LIKE ?)`
		like := "%" + strings.ToLower(q) + "%"
		args = append(args, like, like, like)
	}
	query += ` ORDER BY nombre ASC, id ASC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaGPSDispositivo, 0)
	for rows.Next() {
		var d EmpresaGPSDispositivo
		if err := rows.Scan(
			&d.ID,
			&d.EmpresaID,
			&d.Codigo,
			&d.Nombre,
			&d.Descripcion,
			&d.UltimaLatitud,
			&d.UltimaLongitud,
			&d.UltimaPrecisionMetros,
			&d.UltimaVelocidadKMH,
			&d.UltimoReporteEn,
			&d.FechaCreacion,
			&d.FechaActualizacion,
			&d.UsuarioCreador,
			&d.Estado,
			&d.Observaciones,
		); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// CountEmpresaGPSDispositivos cuenta todos los dispositivos GPS registrados para la empresa (cualquier estado).
func CountEmpresaGPSDispositivos(dbConn *sql.DB, empresaID int64) (int64, error) {
	var n int64
	err := dbConn.QueryRow(`SELECT COUNT(*) FROM empresa_gps_dispositivos WHERE empresa_id = ?`, empresaID).Scan(&n)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// UpdateEmpresaGPSDispositivo actualiza datos base de un dispositivo GPS.
func UpdateEmpresaGPSDispositivo(dbConn *sql.DB, d EmpresaGPSDispositivo) error {
	codigo := sanitizeGPSCode(d.Codigo)
	if codigo == "" {
		codigo = defaultGPSCode(d.EmpresaID, d.Nombre)
	}
	res, err := dbConn.Exec(`UPDATE empresa_gps_dispositivos
	SET
		codigo = ?,
		nombre = ?,
		descripcion = ?,
		observaciones = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`,
		codigo,
		strings.TrimSpace(d.Nombre),
		strings.TrimSpace(d.Descripcion),
		strings.TrimSpace(d.Observaciones),
		d.EmpresaID,
		d.ID,
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

// SetEmpresaGPSDispositivoEstado activa o desactiva un dispositivo GPS.
func SetEmpresaGPSDispositivoEstado(dbConn *sql.DB, empresaID, id int64, estado string) error {
	estado = strings.TrimSpace(strings.ToLower(estado))
	if estado != "activo" && estado != "inactivo" {
		estado = "activo"
	}
	res, err := dbConn.Exec(`UPDATE empresa_gps_dispositivos
	SET estado = ?, fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`, estado, empresaID, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteEmpresaGPSDispositivo elimina un dispositivo GPS y sus puntos de recorrido.
func DeleteEmpresaGPSDispositivo(dbConn *sql.DB, empresaID, id int64) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM empresa_gps_recorridos WHERE empresa_id = ? AND dispositivo_id = ?`, empresaID, id); err != nil {
		_ = tx.Rollback()
		return err
	}

	res, err := tx.Exec(`DELETE FROM empresa_gps_dispositivos WHERE empresa_id = ? AND id = ?`, empresaID, id)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		_ = tx.Rollback()
		return sql.ErrNoRows
	}

	return tx.Commit()
}

// CreateEmpresaGPSRecorrido registra un punto de recorrido GPS.
func CreateEmpresaGPSRecorrido(dbConn *sql.DB, p EmpresaGPSRecorrido) (int64, error) {
	capturadoEn := strings.TrimSpace(p.CapturadoEn)
	if capturadoEn == "" {
		capturadoEn = time.Now().Format("2006-01-02 15:04:05")
	}
	fuente := strings.TrimSpace(p.Fuente)
	if fuente == "" {
		fuente = "manual"
	}

	tx, err := dbConn.Begin()
	if err != nil {
		return 0, err
	}

	res, err := tx.Exec(`INSERT INTO empresa_gps_recorridos (
		empresa_id, dispositivo_id, latitud, longitud,
		precision_metros, velocidad_kmh, rumbo_grados, altitud_metros,
		fuente, capturado_en, usuario_creador, estado, observaciones,
		fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, COALESCE(NULLIF(?, ''), 'activo'), ?, datetime('now','localtime'), datetime('now','localtime'))`,
		p.EmpresaID,
		p.DispositivoID,
		p.Latitud,
		p.Longitud,
		p.PrecisionMetros,
		p.VelocidadKMH,
		p.RumboGrados,
		p.AltitudMetros,
		fuente,
		capturadoEn,
		strings.TrimSpace(p.UsuarioCreador),
		strings.TrimSpace(p.Estado),
		strings.TrimSpace(p.Observaciones),
	)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	upd, err := tx.Exec(`UPDATE empresa_gps_dispositivos
	SET
		ultima_latitud = ?,
		ultima_longitud = ?,
		ultima_precision_metros = ?,
		ultima_velocidad_kmh = ?,
		ultimo_reporte_en = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`,
		p.Latitud,
		p.Longitud,
		p.PrecisionMetros,
		p.VelocidadKMH,
		capturadoEn,
		p.EmpresaID,
		p.DispositivoID,
	)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	affected, _ := upd.RowsAffected()
	if affected == 0 {
		_ = tx.Rollback()
		return 0, fmt.Errorf("dispositivo gps no encontrado")
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

// ListEmpresaGPSRecorridos lista puntos de recorrido por empresa.
func ListEmpresaGPSRecorridos(dbConn *sql.DB, empresaID, dispositivoID int64, includeInactive bool, desdeMinutos, limit int) ([]EmpresaGPSRecorrido, error) {
	if limit <= 0 || limit > 5000 {
		limit = 600
	}

	query := `SELECT
		id, empresa_id, dispositivo_id, latitud, longitud,
		COALESCE(precision_metros, 0), COALESCE(velocidad_kmh, 0),
		COALESCE(rumbo_grados, 0), COALESCE(altitud_metros, 0),
		COALESCE(fuente, 'manual'), COALESCE(capturado_en, ''),
		COALESCE(fecha_creacion, ''), COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''), COALESCE(estado, 'activo'), COALESCE(observaciones, '')
	FROM empresa_gps_recorridos
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	if dispositivoID > 0 {
		query += ` AND dispositivo_id = ?`
		args = append(args, dispositivoID)
	}
	if !includeInactive {
		query += ` AND estado = 'activo'`
	}
	if desdeMinutos > 0 {
		query += ` AND datetime(capturado_en) >= datetime('now','localtime', ?)`
		args = append(args, fmt.Sprintf("-%d minutes", desdeMinutos))
	}
	query += ` ORDER BY datetime(capturado_en) ASC, id ASC LIMIT ?`
	args = append(args, limit)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaGPSRecorrido, 0)
	for rows.Next() {
		var p EmpresaGPSRecorrido
		if err := rows.Scan(
			&p.ID,
			&p.EmpresaID,
			&p.DispositivoID,
			&p.Latitud,
			&p.Longitud,
			&p.PrecisionMetros,
			&p.VelocidadKMH,
			&p.RumboGrados,
			&p.AltitudMetros,
			&p.Fuente,
			&p.CapturadoEn,
			&p.FechaCreacion,
			&p.FechaActualizacion,
			&p.UsuarioCreador,
			&p.Estado,
			&p.Observaciones,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// UpdateEmpresaGPSRecorrido actualiza un punto de recorrido.
func UpdateEmpresaGPSRecorrido(dbConn *sql.DB, p EmpresaGPSRecorrido) error {
	capturadoEn := strings.TrimSpace(p.CapturadoEn)
	if capturadoEn == "" {
		capturadoEn = time.Now().Format("2006-01-02 15:04:05")
	}
	fuente := strings.TrimSpace(p.Fuente)
	if fuente == "" {
		fuente = "manual"
	}

	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}

	res, err := tx.Exec(`UPDATE empresa_gps_recorridos
	SET
		dispositivo_id = ?,
		latitud = ?,
		longitud = ?,
		precision_metros = ?,
		velocidad_kmh = ?,
		rumbo_grados = ?,
		altitud_metros = ?,
		fuente = ?,
		capturado_en = ?,
		observaciones = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`,
		p.DispositivoID,
		p.Latitud,
		p.Longitud,
		p.PrecisionMetros,
		p.VelocidadKMH,
		p.RumboGrados,
		p.AltitudMetros,
		fuente,
		capturadoEn,
		strings.TrimSpace(p.Observaciones),
		p.EmpresaID,
		p.ID,
	)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		_ = tx.Rollback()
		return sql.ErrNoRows
	}

	if _, err := tx.Exec(`UPDATE empresa_gps_dispositivos
	SET
		ultima_latitud = ?,
		ultima_longitud = ?,
		ultima_precision_metros = ?,
		ultima_velocidad_kmh = ?,
		ultimo_reporte_en = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`,
		p.Latitud,
		p.Longitud,
		p.PrecisionMetros,
		p.VelocidadKMH,
		capturadoEn,
		p.EmpresaID,
		p.DispositivoID,
	); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// SetEmpresaGPSRecorridoEstado activa o desactiva un punto de recorrido.
func SetEmpresaGPSRecorridoEstado(dbConn *sql.DB, empresaID, id int64, estado string) error {
	estado = strings.TrimSpace(strings.ToLower(estado))
	if estado != "activo" && estado != "inactivo" {
		estado = "activo"
	}
	res, err := dbConn.Exec(`UPDATE empresa_gps_recorridos
	SET estado = ?, fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`, estado, empresaID, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteEmpresaGPSRecorrido elimina un punto de recorrido.
func DeleteEmpresaGPSRecorrido(dbConn *sql.DB, empresaID, id int64) error {
	res, err := dbConn.Exec(`DELETE FROM empresa_gps_recorridos WHERE empresa_id = ? AND id = ?`, empresaID, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func sanitizeGPSCode(raw string) string {
	clean := strings.TrimSpace(strings.ToUpper(raw))
	if clean == "" {
		return ""
	}
	clean = strings.ReplaceAll(clean, " ", "-")
	clean = strings.ReplaceAll(clean, "_", "-")
	for strings.Contains(clean, "--") {
		clean = strings.ReplaceAll(clean, "--", "-")
	}
	return strings.Trim(clean, "-")
}

func defaultGPSCode(empresaID int64, nombre string) string {
	base := sanitizeGPSCode(nombre)
	if base == "" {
		base = "DISPOSITIVO"
	}
	return fmt.Sprintf("GPS-%d-%s-%d", empresaID, base, time.Now().Unix())
}
