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
	Marca                 string  `json:"marca,omitempty"`
	Modelo                string  `json:"modelo,omitempty"`
	TipoDispositivo       string  `json:"tipo_dispositivo,omitempty"`
	Proveedor             string  `json:"proveedor,omitempty"`
	IdentificadorHardware string  `json:"identificador_hardware,omitempty"`
	TelefonoSIM           string  `json:"telefono_sim,omitempty"`
	PlacaActivo           string  `json:"placa_activo,omitempty"`
	ActivoReferencia      string  `json:"activo_referencia,omitempty"`
	IntervaloReporteSeg   int     `json:"intervalo_reporte_segundos,omitempty"`
	Protocolo             string  `json:"protocolo,omitempty"`
	UltimaBateriaPct      float64 `json:"ultima_bateria_porcentaje,omitempty"`
	UltimaSenalPct        float64 `json:"ultima_senal_porcentaje,omitempty"`
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
	BateriaPorcentaje  float64 `json:"bateria_porcentaje,omitempty"`
	SenalPorcentaje    float64 `json:"senal_porcentaje,omitempty"`
	Evento             string  `json:"evento,omitempty"`
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
			marca TEXT,
			modelo TEXT,
			tipo_dispositivo TEXT DEFAULT 'gps_tracker',
			proveedor TEXT,
			identificador_hardware TEXT,
			telefono_sim TEXT,
			placa_activo TEXT,
			activo_referencia TEXT,
			intervalo_reporte_segundos INTEGER DEFAULT 10,
			protocolo TEXT DEFAULT 'manual',
			ultima_bateria_porcentaje REAL DEFAULT 0,
			ultima_senal_porcentaje REAL DEFAULT 0,
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
			bateria_porcentaje REAL DEFAULT 0,
			senal_porcentaje REAL DEFAULT 0,
			evento TEXT DEFAULT 'posicion',
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
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_dispositivos", "marca", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_dispositivos", "modelo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_dispositivos", "tipo_dispositivo", "TEXT DEFAULT 'gps_tracker'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_dispositivos", "proveedor", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_dispositivos", "identificador_hardware", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_dispositivos", "telefono_sim", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_dispositivos", "placa_activo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_dispositivos", "activo_referencia", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_dispositivos", "intervalo_reporte_segundos", "INTEGER DEFAULT 10"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_dispositivos", "protocolo", "TEXT DEFAULT 'manual'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_dispositivos", "ultima_bateria_porcentaje", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_dispositivos", "ultima_senal_porcentaje", "REAL DEFAULT 0"); err != nil {
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
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_recorridos", "bateria_porcentaje", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_recorridos", "senal_porcentaje", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_gps_recorridos", "evento", "TEXT DEFAULT 'posicion'"); err != nil {
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

	for _, stmt := range []string{
		`CREATE INDEX IF NOT EXISTS ix_empresa_gps_dispositivos_empresa_marca ON empresa_gps_dispositivos(empresa_id, marca, modelo);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_gps_dispositivos_empresa_hardware ON empresa_gps_dispositivos(empresa_id, identificador_hardware);`,
	} {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}

// CreateEmpresaGPSDispositivo crea un dispositivo GPS para una empresa.
func CreateEmpresaGPSDispositivo(dbConn *sql.DB, d EmpresaGPSDispositivo) (int64, error) {
	d = normalizeEmpresaGPSDispositivo(d)
	codigo := sanitizeGPSCode(d.Codigo)
	if codigo == "" {
		codigo = defaultGPSCode(d.EmpresaID, d.Nombre)
	}

	res, err := dbConn.Exec(`INSERT INTO empresa_gps_dispositivos (
		empresa_id, codigo, nombre, descripcion,
		marca, modelo, tipo_dispositivo, proveedor, identificador_hardware,
		telefono_sim, placa_activo, activo_referencia, intervalo_reporte_segundos, protocolo,
		usuario_creador, estado, observaciones,
		fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, COALESCE(NULLIF(?, ''), 'activo'), ?, datetime('now','localtime'), datetime('now','localtime'))`,
		d.EmpresaID,
		codigo,
		d.Nombre,
		d.Descripcion,
		d.Marca,
		d.Modelo,
		d.TipoDispositivo,
		d.Proveedor,
		d.IdentificadorHardware,
		d.TelefonoSIM,
		d.PlacaActivo,
		d.ActivoReferencia,
		d.IntervaloReporteSeg,
		d.Protocolo,
		d.UsuarioCreador,
		d.Estado,
		d.Observaciones,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetEmpresaGPSDispositivos lista dispositivos GPS por empresa.
func GetEmpresaGPSDispositivos(dbConn *sql.DB, empresaID int64, includeInactive bool, q string) ([]EmpresaGPSDispositivo, error) {
	query := `SELECT
		id, empresa_id, COALESCE(codigo, ''), COALESCE(nombre, ''), COALESCE(descripcion, ''),
		COALESCE(marca, ''), COALESCE(modelo, ''), COALESCE(tipo_dispositivo, 'gps_tracker'),
		COALESCE(proveedor, ''), COALESCE(identificador_hardware, ''), COALESCE(telefono_sim, ''),
		COALESCE(placa_activo, ''), COALESCE(activo_referencia, ''),
		COALESCE(intervalo_reporte_segundos, 10), COALESCE(protocolo, 'manual'),
		COALESCE(ultima_bateria_porcentaje, 0), COALESCE(ultima_senal_porcentaje, 0),
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
		query += ` AND (
			LOWER(codigo) LIKE ? OR LOWER(nombre) LIKE ? OR LOWER(COALESCE(descripcion,'')) LIKE ?
			OR LOWER(COALESCE(marca,'')) LIKE ? OR LOWER(COALESCE(modelo,'')) LIKE ?
			OR LOWER(COALESCE(proveedor,'')) LIKE ? OR LOWER(COALESCE(identificador_hardware,'')) LIKE ?
			OR LOWER(COALESCE(telefono_sim,'')) LIKE ? OR LOWER(COALESCE(placa_activo,'')) LIKE ?
			OR LOWER(COALESCE(activo_referencia,'')) LIKE ?
		)`
		like := "%" + strings.ToLower(q) + "%"
		args = append(args, like, like, like, like, like, like, like, like, like, like)
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
			&d.Marca,
			&d.Modelo,
			&d.TipoDispositivo,
			&d.Proveedor,
			&d.IdentificadorHardware,
			&d.TelefonoSIM,
			&d.PlacaActivo,
			&d.ActivoReferencia,
			&d.IntervaloReporteSeg,
			&d.Protocolo,
			&d.UltimaBateriaPct,
			&d.UltimaSenalPct,
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
	d = normalizeEmpresaGPSDispositivo(d)
	codigo := sanitizeGPSCode(d.Codigo)
	if codigo == "" {
		codigo = defaultGPSCode(d.EmpresaID, d.Nombre)
	}
	res, err := dbConn.Exec(`UPDATE empresa_gps_dispositivos
	SET
		codigo = ?,
		nombre = ?,
		descripcion = ?,
		marca = ?,
		modelo = ?,
		tipo_dispositivo = ?,
		proveedor = ?,
		identificador_hardware = ?,
		telefono_sim = ?,
		placa_activo = ?,
		activo_referencia = ?,
		intervalo_reporte_segundos = ?,
		protocolo = ?,
		observaciones = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`,
		codigo,
		d.Nombre,
		d.Descripcion,
		d.Marca,
		d.Modelo,
		d.TipoDispositivo,
		d.Proveedor,
		d.IdentificadorHardware,
		d.TelefonoSIM,
		d.PlacaActivo,
		d.ActivoReferencia,
		d.IntervaloReporteSeg,
		d.Protocolo,
		d.Observaciones,
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
	p = normalizeEmpresaGPSRecorrido(p)
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
		bateria_porcentaje, senal_porcentaje, evento,
		fuente, capturado_en, usuario_creador, estado, observaciones,
		fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, COALESCE(NULLIF(?, ''), 'activo'), ?, datetime('now','localtime'), datetime('now','localtime'))`,
		p.EmpresaID,
		p.DispositivoID,
		p.Latitud,
		p.Longitud,
		p.PrecisionMetros,
		p.VelocidadKMH,
		p.RumboGrados,
		p.AltitudMetros,
		p.BateriaPorcentaje,
		p.SenalPorcentaje,
		p.Evento,
		fuente,
		capturadoEn,
		p.UsuarioCreador,
		p.Estado,
		p.Observaciones,
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
		ultima_bateria_porcentaje = ?,
		ultima_senal_porcentaje = ?,
		ultimo_reporte_en = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`,
		p.Latitud,
		p.Longitud,
		p.PrecisionMetros,
		p.VelocidadKMH,
		p.BateriaPorcentaje,
		p.SenalPorcentaje,
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
		COALESCE(bateria_porcentaje, 0), COALESCE(senal_porcentaje, 0),
		COALESCE(evento, 'posicion'),
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
			&p.BateriaPorcentaje,
			&p.SenalPorcentaje,
			&p.Evento,
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
	p = normalizeEmpresaGPSRecorrido(p)
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
		bateria_porcentaje = ?,
		senal_porcentaje = ?,
		evento = ?,
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
		p.BateriaPorcentaje,
		p.SenalPorcentaje,
		p.Evento,
		fuente,
		capturadoEn,
		p.Observaciones,
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
		ultima_bateria_porcentaje = ?,
		ultima_senal_porcentaje = ?,
		ultimo_reporte_en = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`,
		p.Latitud,
		p.Longitud,
		p.PrecisionMetros,
		p.VelocidadKMH,
		p.BateriaPorcentaje,
		p.SenalPorcentaje,
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

func normalizeEmpresaGPSDispositivo(d EmpresaGPSDispositivo) EmpresaGPSDispositivo {
	d.Nombre = strings.TrimSpace(d.Nombre)
	d.Descripcion = strings.TrimSpace(d.Descripcion)
	d.Marca = strings.TrimSpace(d.Marca)
	d.Modelo = strings.TrimSpace(d.Modelo)
	d.Proveedor = strings.TrimSpace(d.Proveedor)
	d.IdentificadorHardware = strings.TrimSpace(d.IdentificadorHardware)
	d.TelefonoSIM = strings.TrimSpace(d.TelefonoSIM)
	d.PlacaActivo = strings.TrimSpace(strings.ToUpper(d.PlacaActivo))
	d.ActivoReferencia = strings.TrimSpace(d.ActivoReferencia)
	d.UsuarioCreador = strings.TrimSpace(d.UsuarioCreador)
	d.Estado = strings.TrimSpace(strings.ToLower(d.Estado))
	d.Observaciones = strings.TrimSpace(d.Observaciones)
	d.TipoDispositivo = normalizeGPSCatalogValue(d.TipoDispositivo, "gps_tracker")
	d.Protocolo = normalizeGPSCatalogValue(d.Protocolo, "manual")
	if d.IntervaloReporteSeg <= 0 {
		d.IntervaloReporteSeg = 10
	}
	if d.IntervaloReporteSeg < 5 {
		d.IntervaloReporteSeg = 5
	}
	if d.IntervaloReporteSeg > 86400 {
		d.IntervaloReporteSeg = 86400
	}
	if d.Estado != "" && d.Estado != "activo" && d.Estado != "inactivo" {
		d.Estado = "activo"
	}
	return d
}

func normalizeEmpresaGPSRecorrido(p EmpresaGPSRecorrido) EmpresaGPSRecorrido {
	p.PrecisionMetros = clampMinFloat(p.PrecisionMetros, 0)
	p.VelocidadKMH = clampMinFloat(p.VelocidadKMH, 0)
	p.RumboGrados = clampFloat(p.RumboGrados, 0, 359.99)
	p.BateriaPorcentaje = clampFloat(p.BateriaPorcentaje, 0, 100)
	p.SenalPorcentaje = clampFloat(p.SenalPorcentaje, 0, 100)
	p.Evento = normalizeGPSCatalogValue(p.Evento, "posicion")
	p.Fuente = normalizeGPSCatalogValue(p.Fuente, "manual")
	p.UsuarioCreador = strings.TrimSpace(p.UsuarioCreador)
	p.Estado = strings.TrimSpace(strings.ToLower(p.Estado))
	p.Observaciones = strings.TrimSpace(p.Observaciones)
	if p.Estado != "" && p.Estado != "activo" && p.Estado != "inactivo" {
		p.Estado = "activo"
	}
	return p
}

func normalizeGPSCatalogValue(raw, fallback string) string {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		return fallback
	}
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "-", "_")
	for strings.Contains(value, "__") {
		value = strings.ReplaceAll(value, "__", "_")
	}
	return strings.Trim(value, "_")
}

func clampMinFloat(value, min float64) float64 {
	if value < min {
		return min
	}
	return value
}

func clampFloat(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
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
