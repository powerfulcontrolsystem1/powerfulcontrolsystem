package db

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type HotelTarjetaAcceso struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	EstacionID         int64  `json:"estacion_id"`
	EstacionCodigo     string `json:"estacion_codigo,omitempty"`
	EstacionNombre     string `json:"estacion_nombre,omitempty"`
	CodigoTarjeta      string `json:"codigo_tarjeta"`
	CardUID            string `json:"card_uid,omitempty"`
	AccessCode         string `json:"access_code,omitempty"`
	HuespedNombre      string `json:"huesped_nombre,omitempty"`
	ReservaID          int64  `json:"reserva_id,omitempty"`
	VigenteDesde       string `json:"vigente_desde"`
	VigenteHasta       string `json:"vigente_hasta"`
	MaxUsos            int64  `json:"max_usos"`
	UsosRealizados     int64  `json:"usos_realizados"`
	UltimoUsoEn        string `json:"ultimo_uso_en,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado"`
	Observaciones      string `json:"observaciones,omitempty"`
}

type HotelTarjetaAccesoFilter struct {
	EstacionID      int64
	ReservaID       int64
	IncludeInactive bool
	Limit           int
}

type HotelTarjetaValidacion struct {
	OK             bool                   `json:"ok"`
	Permitido      bool                   `json:"permitido"`
	Motivo         string                 `json:"motivo,omitempty"`
	Tarjeta        *HotelTarjetaAcceso    `json:"tarjeta,omitempty"`
	EventoID       int64                  `json:"evento_id,omitempty"`
	DeviceID       string                 `json:"device_id,omitempty"`
	EstacionID     int64                  `json:"estacion_id,omitempty"`
	ComandoPuerta  string                 `json:"comando_puerta,omitempty"`
	Provisioning   map[string]string      `json:"provisioning,omitempty"`
	MetadataSalida map[string]interface{} `json:"metadata_salida,omitempty"`
}

func EnsureHotelTarjetasAccesoSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS hotel_tarjetas_acceso (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			estacion_id INTEGER NOT NULL,
			estacion_codigo TEXT,
			estacion_nombre TEXT,
			codigo_tarjeta TEXT NOT NULL,
			card_uid_hash TEXT,
			access_code_hash TEXT,
			huesped_nombre TEXT,
			reserva_id INTEGER,
			vigente_desde TEXT NOT NULL,
			vigente_hasta TEXT NOT NULL,
			max_usos INTEGER DEFAULT 0,
			usos_realizados INTEGER DEFAULT 0,
			ultimo_uso_en TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT,
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_hotel_tarjetas_empresa_codigo ON hotel_tarjetas_acceso(empresa_id, codigo_tarjeta);`,
		`CREATE INDEX IF NOT EXISTS ix_hotel_tarjetas_lookup ON hotel_tarjetas_acceso(empresa_id, estacion_id, estado);`,
		`CREATE INDEX IF NOT EXISTS ix_hotel_tarjetas_reserva ON hotel_tarjetas_acceso(empresa_id, reserva_id);`,
		`CREATE TABLE IF NOT EXISTS hotel_tarjetas_acceso_eventos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			tarjeta_id INTEGER,
			estacion_id INTEGER,
			device_id TEXT,
			resultado TEXT NOT NULL,
			motivo TEXT,
			fecha_evento TEXT DEFAULT (CURRENT_TIMESTAMP),
			metadata_json TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_hotel_tarjetas_eventos_empresa ON hotel_tarjetas_acceso_eventos(empresa_id, fecha_evento);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}
	columns := map[string]string{
		"estacion_codigo":     "TEXT",
		"estacion_nombre":     "TEXT",
		"card_uid_hash":       "TEXT",
		"access_code_hash":    "TEXT",
		"huesped_nombre":      "TEXT",
		"reserva_id":          "INTEGER",
		"max_usos":            "INTEGER DEFAULT 0",
		"usos_realizados":     "INTEGER DEFAULT 0",
		"ultimo_uso_en":       "TEXT",
		"fecha_actualizacion": "TEXT",
		"usuario_creador":     "TEXT",
		"estado":              "TEXT DEFAULT 'activo'",
		"observaciones":       "TEXT",
	}
	for name, typ := range columns {
		if err := ensureColumnIfMissing(dbConn, "hotel_tarjetas_acceso", name, typ); err != nil {
			return err
		}
	}
	return nil
}

func normalizeHotelTarjetaEstado(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "inactivo", "bloqueada", "bloqueado", "revocada", "revocado":
		return "inactivo"
	case "vencida", "vencido":
		return "vencido"
	default:
		return "activo"
	}
}

func hashHotelTarjetaSecret(raw string) string {
	value := strings.TrimSpace(strings.ToUpper(raw))
	if value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(value))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func GenerateHotelAccessCode() (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(b[:])), nil
}

func normalizeHotelTarjetaPayload(payload *HotelTarjetaAcceso) error {
	if payload == nil {
		return fmt.Errorf("payload invalido")
	}
	if payload.EmpresaID <= 0 {
		return fmt.Errorf("empresa_id es obligatorio")
	}
	if payload.EstacionID <= 0 {
		return fmt.Errorf("estacion_id es obligatorio")
	}
	payload.CodigoTarjeta = strings.TrimSpace(payload.CodigoTarjeta)
	if payload.CodigoTarjeta == "" {
		return fmt.Errorf("codigo_tarjeta es obligatorio")
	}
	payload.EstacionCodigo = strings.TrimSpace(payload.EstacionCodigo)
	payload.EstacionNombre = strings.TrimSpace(payload.EstacionNombre)
	payload.HuespedNombre = strings.TrimSpace(payload.HuespedNombre)
	payload.VigenteDesde = strings.TrimSpace(payload.VigenteDesde)
	payload.VigenteHasta = strings.TrimSpace(payload.VigenteHasta)
	if _, err := parseHotelTarjetaDateTime(payload.VigenteDesde); err != nil {
		return fmt.Errorf("vigente_desde invalido")
	}
	desde, _ := parseHotelTarjetaDateTime(payload.VigenteDesde)
	hasta, err := parseHotelTarjetaDateTime(payload.VigenteHasta)
	if err != nil {
		return fmt.Errorf("vigente_hasta invalido")
	}
	if hasta.Before(desde) || hasta.Equal(desde) {
		return fmt.Errorf("vigente_hasta debe ser mayor a vigente_desde")
	}
	if payload.MaxUsos < 0 {
		payload.MaxUsos = 0
	}
	payload.Estado = normalizeHotelTarjetaEstado(payload.Estado)
	payload.UsuarioCreador = strings.TrimSpace(payload.UsuarioCreador)
	payload.Observaciones = strings.TrimSpace(payload.Observaciones)
	return nil
}

func parseHotelTarjetaDateTime(raw string) (time.Time, error) {
	value := strings.TrimSpace(raw)
	layouts := []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05", "2006-01-02 15:04", "2006-01-02T15:04", "2006-01-02"}
	for _, layout := range layouts {
		if ts, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return ts, nil
		}
	}
	return time.Time{}, fmt.Errorf("fecha invalida")
}

func CreateHotelTarjetaAcceso(dbConn *sql.DB, payload HotelTarjetaAcceso) (int64, string, error) {
	if err := EnsureHotelTarjetasAccesoSchema(dbConn); err != nil {
		return 0, "", err
	}
	if err := normalizeHotelTarjetaPayload(&payload); err != nil {
		return 0, "", err
	}
	accessCode := strings.TrimSpace(payload.AccessCode)
	if accessCode == "" {
		var err error
		accessCode, err = GenerateHotelAccessCode()
		if err != nil {
			return 0, "", err
		}
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO hotel_tarjetas_acceso (
		empresa_id, estacion_id, estacion_codigo, estacion_nombre, codigo_tarjeta,
		card_uid_hash, access_code_hash, huesped_nombre, reserva_id, vigente_desde,
		vigente_hasta, max_usos, usos_realizados, usuario_creador, estado, observaciones,
		fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NULLIF(?,0), ?, ?, ?, 0, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		payload.EmpresaID, payload.EstacionID, payload.EstacionCodigo, payload.EstacionNombre, payload.CodigoTarjeta,
		hashHotelTarjetaSecret(payload.CardUID), hashHotelTarjetaSecret(accessCode), payload.HuespedNombre, payload.ReservaID,
		payload.VigenteDesde, payload.VigenteHasta, payload.MaxUsos, payload.UsuarioCreador, payload.Estado, payload.Observaciones,
	)
	return id, accessCode, err
}

func UpdateHotelTarjetaAcceso(dbConn *sql.DB, payload HotelTarjetaAcceso) error {
	if payload.ID <= 0 {
		return fmt.Errorf("id es obligatorio")
	}
	if err := normalizeHotelTarjetaPayload(&payload); err != nil {
		return err
	}
	cardHash := hashHotelTarjetaSecret(payload.CardUID)
	codeHash := hashHotelTarjetaSecret(payload.AccessCode)
	query := `UPDATE hotel_tarjetas_acceso SET
		estacion_id = ?, estacion_codigo = ?, estacion_nombre = ?, codigo_tarjeta = ?,
		huesped_nombre = ?, reserva_id = NULLIF(?,0), vigente_desde = ?, vigente_hasta = ?,
		max_usos = ?, usuario_creador = ?, estado = ?, observaciones = ?,
		fecha_actualizacion = CURRENT_TIMESTAMP`
	args := []interface{}{payload.EstacionID, payload.EstacionCodigo, payload.EstacionNombre, payload.CodigoTarjeta, payload.HuespedNombre, payload.ReservaID, payload.VigenteDesde, payload.VigenteHasta, payload.MaxUsos, payload.UsuarioCreador, payload.Estado, payload.Observaciones}
	if cardHash != "" {
		query += `, card_uid_hash = ?`
		args = append(args, cardHash)
	}
	if codeHash != "" {
		query += `, access_code_hash = ?`
		args = append(args, codeHash)
	}
	query += ` WHERE empresa_id = ? AND id = ?`
	args = append(args, payload.EmpresaID, payload.ID)
	res, err := dbConn.Exec(query, args...)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func scanHotelTarjetaAcceso(scanner interface {
	Scan(dest ...interface{}) error
}) (*HotelTarjetaAcceso, error) {
	item := &HotelTarjetaAcceso{}
	var reserva sql.NullInt64
	if err := scanner.Scan(&item.ID, &item.EmpresaID, &item.EstacionID, &item.EstacionCodigo, &item.EstacionNombre, &item.CodigoTarjeta, &item.HuespedNombre, &reserva, &item.VigenteDesde, &item.VigenteHasta, &item.MaxUsos, &item.UsosRealizados, &item.UltimoUsoEn, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador, &item.Estado, &item.Observaciones); err != nil {
		return nil, err
	}
	if reserva.Valid {
		item.ReservaID = reserva.Int64
	}
	return item, nil
}

func hotelTarjetaSelectSQL() string {
	return `SELECT id, empresa_id, estacion_id, COALESCE(estacion_codigo,''), COALESCE(estacion_nombre,''), codigo_tarjeta, COALESCE(huesped_nombre,''), reserva_id, vigente_desde, vigente_hasta, COALESCE(max_usos,0), COALESCE(usos_realizados,0), COALESCE(ultimo_uso_en,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,''), COALESCE(estado,'activo'), COALESCE(observaciones,'') FROM hotel_tarjetas_acceso`
}

func ListHotelTarjetasAcceso(dbConn *sql.DB, empresaID int64, filter HotelTarjetaAccesoFilter) ([]HotelTarjetaAcceso, error) {
	if err := EnsureHotelTarjetasAccesoSchema(dbConn); err != nil {
		return nil, err
	}
	if filter.Limit <= 0 {
		filter.Limit = 300
	}
	// #nosec G202 -- SQL structure is assembled only from server-side allowlists; all external values remain bound parameters.
	query := hotelTarjetaSelectSQL() + ` WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	if !filter.IncludeInactive {
		query += ` AND COALESCE(estado,'activo') = 'activo'`
	}
	if filter.EstacionID > 0 {
		query += ` AND estacion_id = ?`
		args = append(args, filter.EstacionID)
	}
	if filter.ReservaID > 0 {
		query += ` AND reserva_id = ?`
		args = append(args, filter.ReservaID)
	}
	query += ` ORDER BY id DESC LIMIT ?`
	args = append(args, filter.Limit)
	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]HotelTarjetaAcceso, 0)
	for rows.Next() {
		item, err := scanHotelTarjetaAcceso(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

func GetHotelTarjetaAccesoByID(dbConn *sql.DB, empresaID, id int64) (*HotelTarjetaAcceso, error) {
	row := dbConn.QueryRow(hotelTarjetaSelectSQL()+` WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, id)
	return scanHotelTarjetaAcceso(row)
}

func SetHotelTarjetaAccesoEstado(dbConn *sql.DB, empresaID, id int64, estado string) error {
	res, err := dbConn.Exec(`UPDATE hotel_tarjetas_acceso SET estado = ?, fecha_actualizacion = CURRENT_TIMESTAMP WHERE empresa_id = ? AND id = ?`, normalizeHotelTarjetaEstado(estado), empresaID, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func ValidateHotelTarjetaAcceso(dbConn *sql.DB, empresaID, estacionID int64, cardUID, accessCode, deviceID string) (*HotelTarjetaValidacion, error) {
	if err := EnsureHotelTarjetasAccesoSchema(dbConn); err != nil {
		return nil, err
	}
	cardHash := hashHotelTarjetaSecret(cardUID)
	codeHash := hashHotelTarjetaSecret(accessCode)
	if cardHash == "" && codeHash == "" {
		return nil, fmt.Errorf("card_uid o access_code es obligatorio")
	}
	query := hotelTarjetaSelectSQL() + ` WHERE empresa_id = ? AND COALESCE(estado,'activo') = 'activo'`
	args := []interface{}{empresaID}
	if estacionID > 0 {
		query += ` AND estacion_id = ?`
		args = append(args, estacionID)
	}
	if cardHash != "" && codeHash != "" {
		query += ` AND (card_uid_hash = ? OR access_code_hash = ?)`
		args = append(args, cardHash, codeHash)
	} else if cardHash != "" {
		query += ` AND card_uid_hash = ?`
		args = append(args, cardHash)
	} else {
		query += ` AND access_code_hash = ?`
		args = append(args, codeHash)
	}
	query += ` ORDER BY id DESC LIMIT 1`
	item, err := scanHotelTarjetaAcceso(dbConn.QueryRow(query, args...))
	if err != nil {
		if err == sql.ErrNoRows {
			eventID, _ := insertHotelTarjetaEvento(dbConn, empresaID, 0, estacionID, deviceID, "denegado", "tarjeta_no_encontrada")
			return &HotelTarjetaValidacion{OK: true, Permitido: false, Motivo: "tarjeta_no_encontrada", EventoID: eventID, EstacionID: estacionID, DeviceID: strings.TrimSpace(deviceID)}, nil
		}
		return nil, err
	}
	now := time.Now()
	desde, _ := parseHotelTarjetaDateTime(item.VigenteDesde)
	hasta, _ := parseHotelTarjetaDateTime(item.VigenteHasta)
	permitido := true
	motivo := "acceso_permitido"
	if now.Before(desde) {
		permitido, motivo = false, "tarjeta_aun_no_vigente"
	} else if now.After(hasta) {
		permitido, motivo = false, "tarjeta_vencida"
	} else if item.MaxUsos > 0 && item.UsosRealizados >= item.MaxUsos {
		permitido, motivo = false, "max_usos_alcanzado"
	}
	resultado := "permitido"
	if !permitido {
		resultado = "denegado"
	} else {
		_, _ = dbConn.Exec(`UPDATE hotel_tarjetas_acceso SET usos_realizados = COALESCE(usos_realizados,0) + 1, ultimo_uso_en = CURRENT_TIMESTAMP, fecha_actualizacion = CURRENT_TIMESTAMP WHERE empresa_id = ? AND id = ?`, empresaID, item.ID)
		item.UsosRealizados++
	}
	eventID, _ := insertHotelTarjetaEvento(dbConn, empresaID, item.ID, item.EstacionID, deviceID, resultado, motivo)
	resp := &HotelTarjetaValidacion{OK: true, Permitido: permitido, Motivo: motivo, Tarjeta: item, EventoID: eventID, DeviceID: strings.TrimSpace(deviceID), EstacionID: item.EstacionID}
	if permitido {
		resp.ComandoPuerta = "unlock"
	}
	return resp, nil
}

func insertHotelTarjetaEvento(dbConn *sql.DB, empresaID, tarjetaID, estacionID int64, deviceID, resultado, motivo string) (int64, error) {
	return insertSQLCompat(dbConn, `INSERT INTO hotel_tarjetas_acceso_eventos (empresa_id, tarjeta_id, estacion_id, device_id, resultado, motivo, metadata_json, fecha_evento) VALUES (?, NULLIF(?,0), NULLIF(?,0), ?, ?, ?, '{}', CURRENT_TIMESTAMP)`, empresaID, tarjetaID, estacionID, strings.TrimSpace(deviceID), resultado, motivo)
}
