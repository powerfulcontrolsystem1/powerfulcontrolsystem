package db

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"strings"

	"github.com/you/pos-backend/secure"
)

// EmpresaSensorDevice representa un dispositivo sensor (Raspberry) asociado a una empresa/estación
type EmpresaSensorDevice struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	DeviceID           string `json:"device_id"`
	DeviceToken        string `json:"device_token,omitempty"`
	EstacionID         int64  `json:"estacion_id,omitempty"`
	LastState          string `json:"last_state,omitempty"`
	LastSeen           string `json:"last_seen,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// EnsureEmpresaSensorPuertasSchema crea/migra las tablas del módulo sensor de puertas por empresa
func EnsureEmpresaSensorPuertasSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_sensor_puertas_devices (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			device_id TEXT NOT NULL,
			device_token_hash TEXT,
			device_token_enc TEXT,
			estacion_id INTEGER,
			last_state TEXT,
			last_seen TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT,
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_sensor_device_empresa_device ON empresa_sensor_puertas_devices(empresa_id, device_id);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_sensor_device_empresa ON empresa_sensor_puertas_devices(empresa_id);`,
		`CREATE TABLE IF NOT EXISTS empresa_sensor_puertas_messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER,
			device_id TEXT,
			estacion_id INTEGER,
			message_text TEXT,
			raw_text TEXT,
			received_at TEXT DEFAULT (datetime('now','localtime')),
			procesado INTEGER DEFAULT 0
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_sensor_messages_empresa ON empresa_sensor_puertas_messages(empresa_id);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_sensor_messages_device ON empresa_sensor_puertas_messages(device_id);`,
	}

	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_sensor_puertas_devices", "empresa_id", "INTEGER NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_sensor_puertas_devices", "device_id", "TEXT NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_sensor_puertas_devices", "estacion_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_sensor_puertas_devices", "last_state", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_sensor_puertas_devices", "last_seen", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_sensor_puertas_devices", "fecha_creacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_sensor_puertas_devices", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_sensor_puertas_devices", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_sensor_puertas_devices", "device_token_hash", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_sensor_puertas_devices", "device_token_enc", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_sensor_puertas_devices", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_sensor_puertas_devices", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

// GetEmpresaSensorByDeviceID busca un dispositivo por su identificador (case-insensitive)
func GetEmpresaSensorByDeviceID(dbConn *sql.DB, deviceID string) (*EmpresaSensorDevice, error) {
	idv := strings.TrimSpace(strings.ToLower(deviceID))
	var p EmpresaSensorDevice
	row := dbConn.QueryRow(`SELECT id, empresa_id, device_id, COALESCE(estacion_id,0), last_state, last_seen, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones FROM empresa_sensor_puertas_devices WHERE lower(device_id) = ? AND estado = 'activo' LIMIT 1`, idv)
	var estacion int64
	var observ sql.NullString
	if err := row.Scan(&p.ID, &p.EmpresaID, &p.DeviceID, &estacion, &p.LastState, &p.LastSeen, &p.FechaCreacion, &p.FechaActualizacion, &p.UsuarioCreador, &p.Estado, &observ); err != nil {
		return nil, err
	}
	p.Observaciones = observ.String
	p.EstacionID = estacion
	return &p, nil
}

// GetEmpresaSensorByToken busca un dispositivo asociado a un token (hash) activo
func GetEmpresaSensorByToken(dbConn *sql.DB, token string) (*EmpresaSensorDevice, error) {
	sum := sha256.Sum256([]byte(token))
	hash := base64.StdEncoding.EncodeToString(sum[:])
	var p EmpresaSensorDevice
	row := dbConn.QueryRow(`SELECT id, empresa_id, device_id, COALESCE(estacion_id,0), last_state, last_seen, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones FROM empresa_sensor_puertas_devices WHERE device_token_hash = ? AND estado = 'activo' LIMIT 1`, hash)
	var estacion int64
	var observ sql.NullString
	if err := row.Scan(&p.ID, &p.EmpresaID, &p.DeviceID, &estacion, &p.LastState, &p.LastSeen, &p.FechaCreacion, &p.FechaActualizacion, &p.UsuarioCreador, &p.Estado, &observ); err != nil {
		return nil, err
	}
	p.Observaciones = observ.String
	p.EstacionID = estacion
	return &p, nil
}

// GetEmpresaSensorsByEmpresa lista los dispositivos registrados para una empresa
func GetEmpresaSensorsByEmpresa(dbConn *sql.DB, empresaID int64) ([]EmpresaSensorDevice, error) {
	q := `SELECT id, empresa_id, device_id, COALESCE(estacion_id,0), last_state, last_seen, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones FROM empresa_sensor_puertas_devices WHERE empresa_id = ? AND estado = 'activo'`
	rows, err := dbConn.Query(q, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaSensorDevice{}
	for rows.Next() {
		var p EmpresaSensorDevice
		var estacion int64
		var observ sql.NullString
		if err := rows.Scan(&p.ID, &p.EmpresaID, &p.DeviceID, &estacion, &p.LastState, &p.LastSeen, &p.FechaCreacion, &p.FechaActualizacion, &p.UsuarioCreador, &p.Estado, &observ); err != nil {
			return nil, err
		}
		p.Observaciones = observ.String
		p.EstacionID = estacion
		out = append(out, p)
	}
	return out, nil
}

// UpsertEmpresaSensorDevice crea o actualiza el mapeo device -> empresa/estacion
func UpsertEmpresaSensorDevice(dbConn *sql.DB, p *EmpresaSensorDevice) (int64, error) {
	device := strings.TrimSpace(strings.ToLower(p.DeviceID))

	var tokenHash sql.NullString
	var tokenEnc sql.NullString
	if strings.TrimSpace(p.DeviceToken) != "" {
		sum := sha256.Sum256([]byte(p.DeviceToken))
		tokenHash.String = base64.StdEncoding.EncodeToString(sum[:])
		tokenHash.Valid = true
		enc, err := secure.EncryptString(p.DeviceToken)
		if err != nil {
			return 0, err
		}
		tokenEnc.String = enc
		tokenEnc.Valid = true
	}

	var existingID int64
	row := dbConn.QueryRow(`SELECT id FROM empresa_sensor_puertas_devices WHERE empresa_id = ? AND lower(device_id) = ? LIMIT 1`, p.EmpresaID, device)
	if err := row.Scan(&existingID); err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	if existingID > 0 {
		if tokenHash.Valid {
			_, err := dbConn.Exec(`UPDATE empresa_sensor_puertas_devices SET estacion_id = NULLIF(?,0), fecha_actualizacion = datetime('now','localtime'), usuario_creador = ?, estado = COALESCE(NULLIF(?, ''), 'activo'), device_token_hash = ?, device_token_enc = ? WHERE id = ?`, p.EstacionID, p.UsuarioCreador, p.Estado, tokenHash.String, tokenEnc.String, existingID)
			if err != nil {
				return 0, err
			}
			return existingID, nil
		}
		_, err := dbConn.Exec(`UPDATE empresa_sensor_puertas_devices SET estacion_id = NULLIF(?,0), fecha_actualizacion = datetime('now','localtime'), usuario_creador = ?, estado = COALESCE(NULLIF(?, ''), 'activo') WHERE id = ?`, p.EstacionID, p.UsuarioCreador, p.Estado, existingID)
		if err != nil {
			return 0, err
		}
		return existingID, nil
	}

	if tokenHash.Valid {
		res, err := dbConn.Exec(`INSERT INTO empresa_sensor_puertas_devices (empresa_id, device_id, estacion_id, last_state, last_seen, device_token_hash, device_token_enc, fecha_creacion, fecha_actualizacion, usuario_creador, estado) VALUES (?, ?, NULLIF(?,0), ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'), ?, COALESCE(NULLIF(?, ''), 'activo'))`, p.EmpresaID, device, p.EstacionID, p.LastState, p.LastSeen, tokenHash.String, tokenEnc.String, p.UsuarioCreador, p.Estado)
		if err != nil {
			return 0, err
		}
		return res.LastInsertId()
	}

	res, err := dbConn.Exec(`INSERT INTO empresa_sensor_puertas_devices (empresa_id, device_id, estacion_id, last_state, last_seen, fecha_creacion, fecha_actualizacion, usuario_creador, estado) VALUES (?, ?, NULLIF(?,0), ?, ?, datetime('now','localtime'), datetime('now','localtime'), ?, COALESCE(NULLIF(?, ''), 'activo'))`, p.EmpresaID, device, p.EstacionID, p.LastState, p.LastSeen, p.UsuarioCreador, p.Estado)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateDeviceHeartbeat actualiza el estado y last_seen del dispositivo y devuelve la empresa/estacion asociada
func UpdateDeviceHeartbeat(dbConn *sql.DB, deviceID, state string) (int64, int64, error) {
	device := strings.TrimSpace(strings.ToLower(deviceID))
	var id int64
	var empresaID int64
	var estacion sql.NullInt64
	row := dbConn.QueryRow(`SELECT id, empresa_id, estacion_id FROM empresa_sensor_puertas_devices WHERE lower(device_id) = ? AND estado = 'activo' LIMIT 1`, device)
	if err := row.Scan(&id, &empresaID, &estacion); err != nil {
		return 0, 0, err
	}

	nowState := strings.TrimSpace(state)
	if _, err := dbConn.Exec(`UPDATE empresa_sensor_puertas_devices SET last_state = ?, last_seen = datetime('now','localtime'), fecha_actualizacion = datetime('now','localtime') WHERE id = ?`, nowState, id); err != nil {
		return 0, 0, err
	}

	var estID int64
	if estacion.Valid {
		estID = estacion.Int64
	}
	return empresaID, estID, nil
}

// EmpresaSensorMessage representa un mensaje recibido desde un dispositivo
type EmpresaSensorMessage struct {
	ID         int64  `json:"id"`
	EmpresaID  int64  `json:"empresa_id"`
	DeviceID   string `json:"device_id"`
	EstacionID int64  `json:"estacion_id,omitempty"`
	Message    string `json:"message_text"`
	ReceivedAt string `json:"received_at"`
	Procesado  int    `json:"procesado,omitempty"`
}

// InsertEmpresaSensorMessage registra un mensaje recibido desde un dispositivo y retorna el id insertado
func InsertEmpresaSensorMessage(dbConn *sql.DB, deviceID, messageText string) (int64, int64, int64, error) {
	// primero resolver el dispositivo y empresa asociada
	dev, err := GetEmpresaSensorByDeviceID(dbConn, deviceID)
	if err != nil {
		return 0, 0, 0, err
	}

	res, err := dbConn.Exec(`INSERT INTO empresa_sensor_puertas_messages (empresa_id, device_id, estacion_id, message_text, raw_text, received_at) VALUES (?, ?, NULLIF(?,0), ?, ?, datetime('now','localtime'))`, dev.EmpresaID, dev.DeviceID, dev.EstacionID, messageText, messageText)
	if err != nil {
		return 0, 0, 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, 0, 0, err
	}
	return id, dev.EmpresaID, dev.EstacionID, nil
}

// GetEmpresaSensorMessagesByEmpresa lista mensajes recibidos para una empresa (ordenados por fecha descendente)
func GetEmpresaSensorMessagesByEmpresa(dbConn *sql.DB, empresaID int64) ([]EmpresaSensorMessage, error) {
	q := `SELECT id, empresa_id, device_id, COALESCE(estacion_id,0), message_text, received_at, procesado FROM empresa_sensor_puertas_messages WHERE empresa_id = ? ORDER BY received_at DESC LIMIT 1000`
	rows, err := dbConn.Query(q, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaSensorMessage{}
	for rows.Next() {
		var m EmpresaSensorMessage
		var estacion int64
		if err := rows.Scan(&m.ID, &m.EmpresaID, &m.DeviceID, &estacion, &m.Message, &m.ReceivedAt, &m.Procesado); err != nil {
			return nil, err
		}
		m.EstacionID = estacion
		out = append(out, m)
	}
	return out, nil
}
