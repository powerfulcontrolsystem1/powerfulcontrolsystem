package db

import (
	"database/sql"
	"fmt"
	"strings"
)

const (
	CamaraProtocoloRTSP   = "rtsp"
	CamaraProtocoloONVIF  = "onvif"
	CamaraProtocoloHLS    = "hls"
	CamaraProtocoloWebRTC = "webrtc"
	CamaraProtocoloMJPEG  = "mjpeg"
	CamaraProtocoloIframe = "iframe"
)

type EmpresaCamara struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Nombre             string `json:"nombre"`
	Ubicacion          string `json:"ubicacion"`
	DVRNombre          string `json:"dvr_nombre"`
	DVRHost            string `json:"dvr_host"`
	Canal              string `json:"canal"`
	Fabricante         string `json:"fabricante"`
	Modelo             string `json:"modelo"`
	ProtocoloOrigen    string `json:"protocolo_origen"`
	URLStream          string `json:"url_stream"`
	URLSnapshot        string `json:"url_snapshot"`
	URLEmbed           string `json:"url_embed"`
	VisorTipo          string `json:"visor_tipo"`
	UsuarioRef         string `json:"usuario_ref"`
	PasswordRef        string `json:"password_ref"`
	EstacionID         int64  `json:"estacion_id"`
	CargarEnEstaciones bool   `json:"cargar_en_estaciones"`
	Orden              int    `json:"orden"`
	Activa             bool   `json:"activa"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

func EnsureEmpresaCamarasSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("conexion de base de datos no disponible")
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_camaras (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			nombre TEXT NOT NULL,
			ubicacion TEXT,
			dvr_nombre TEXT,
			dvr_host TEXT,
			canal TEXT,
			fabricante TEXT,
			modelo TEXT,
			protocolo_origen TEXT DEFAULT 'rtsp',
			url_stream TEXT,
			url_snapshot TEXT,
			url_embed TEXT,
			visor_tipo TEXT DEFAULT 'auto',
			usuario_ref TEXT,
			password_ref TEXT,
			estacion_id INTEGER DEFAULT 0,
			cargar_en_estaciones INTEGER DEFAULT 1,
			orden INTEGER DEFAULT 0,
			activa INTEGER DEFAULT 1,
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_camaras_empresa ON empresa_camaras(empresa_id, estado, activa)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_camaras_estacion ON empresa_camaras(empresa_id, estacion_id)`,
	}
	for _, stmt := range stmts {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	columns := []struct {
		table string
		name  string
		def   string
	}{
		{"empresa_camaras", "url_snapshot", "TEXT"},
		{"empresa_camaras", "url_embed", "TEXT"},
		{"empresa_camaras", "visor_tipo", "TEXT DEFAULT 'auto'"},
		{"empresa_camaras", "usuario_ref", "TEXT"},
		{"empresa_camaras", "password_ref", "TEXT"},
		{"empresa_camaras", "estacion_id", "INTEGER DEFAULT 0"},
		{"empresa_camaras", "cargar_en_estaciones", "INTEGER DEFAULT 1"},
	}
	for _, col := range columns {
		if err := ensureColumnIfMissing(dbConn, col.table, col.name, col.def); err != nil {
			return err
		}
	}
	return nil
}

func UpsertEmpresaCamara(dbConn *sql.DB, item EmpresaCamara) (int64, error) {
	if dbConn == nil {
		return 0, fmt.Errorf("conexion de base de datos no disponible")
	}
	if item.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id requerido")
	}
	item.Nombre = strings.TrimSpace(item.Nombre)
	if item.Nombre == "" {
		return 0, fmt.Errorf("nombre de camara requerido")
	}
	item.ProtocoloOrigen = normalizeEmpresaCamaraProtocolo(item.ProtocoloOrigen)
	item.VisorTipo = normalizeEmpresaCamaraVisor(item.VisorTipo)
	item.Estado = normalizeEmpresaCamaraEstado(item.Estado)
	if item.Orden < 0 {
		item.Orden = 0
	}
	if item.ID > 0 {
		_, err := execSQLCompat(dbConn, `UPDATE empresa_camaras SET nombre=?, ubicacion=?, dvr_nombre=?, dvr_host=?, canal=?, fabricante=?, modelo=?, protocolo_origen=?, url_stream=?, url_snapshot=?, url_embed=?, visor_tipo=?, usuario_ref=?, password_ref=?, estacion_id=?, cargar_en_estaciones=?, orden=?, activa=?, fecha_actualizacion=CAST(CURRENT_TIMESTAMP AS TEXT), usuario_creador=?, estado=?, observaciones=? WHERE empresa_id=? AND id=?`,
			item.Nombre, strings.TrimSpace(item.Ubicacion), strings.TrimSpace(item.DVRNombre), strings.TrimSpace(item.DVRHost), strings.TrimSpace(item.Canal), strings.TrimSpace(item.Fabricante), strings.TrimSpace(item.Modelo), item.ProtocoloOrigen, strings.TrimSpace(item.URLStream), strings.TrimSpace(item.URLSnapshot), strings.TrimSpace(item.URLEmbed), item.VisorTipo, strings.TrimSpace(item.UsuarioRef), strings.TrimSpace(item.PasswordRef), item.EstacionID, item.CargarEnEstaciones, item.Orden, item.Activa, strings.TrimSpace(item.UsuarioCreador), item.Estado, strings.TrimSpace(item.Observaciones), item.EmpresaID, item.ID)
		if err != nil {
			return 0, err
		}
		return item.ID, nil
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_camaras (empresa_id, nombre, ubicacion, dvr_nombre, dvr_host, canal, fabricante, modelo, protocolo_origen, url_stream, url_snapshot, url_embed, visor_tipo, usuario_ref, password_ref, estacion_id, cargar_en_estaciones, orden, activa, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CAST(CURRENT_TIMESTAMP AS TEXT), CAST(CURRENT_TIMESTAMP AS TEXT), ?, ?, ?)`,
		item.EmpresaID, item.Nombre, strings.TrimSpace(item.Ubicacion), strings.TrimSpace(item.DVRNombre), strings.TrimSpace(item.DVRHost), strings.TrimSpace(item.Canal), strings.TrimSpace(item.Fabricante), strings.TrimSpace(item.Modelo), item.ProtocoloOrigen, strings.TrimSpace(item.URLStream), strings.TrimSpace(item.URLSnapshot), strings.TrimSpace(item.URLEmbed), item.VisorTipo, strings.TrimSpace(item.UsuarioRef), strings.TrimSpace(item.PasswordRef), item.EstacionID, item.CargarEnEstaciones, item.Orden, item.Activa, strings.TrimSpace(item.UsuarioCreador), item.Estado, strings.TrimSpace(item.Observaciones))
}

func ListEmpresaCamaras(dbConn *sql.DB, empresaID int64, includeInactive bool) ([]EmpresaCamara, error) {
	if dbConn == nil {
		return nil, fmt.Errorf("conexion de base de datos no disponible")
	}
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id requerido")
	}
	where := "WHERE empresa_id=?"
	if !includeInactive {
		where += " AND COALESCE(estado,'activo')='activo' AND COALESCE(activa,1)=1"
	}
	rows, err := querySQLCompat(dbConn, `SELECT id, empresa_id, COALESCE(nombre,''), COALESCE(ubicacion,''), COALESCE(dvr_nombre,''), COALESCE(dvr_host,''), COALESCE(canal,''), COALESCE(fabricante,''), COALESCE(modelo,''), COALESCE(protocolo_origen,'rtsp'), COALESCE(url_stream,''), COALESCE(url_snapshot,''), COALESCE(url_embed,''), COALESCE(visor_tipo,'auto'), COALESCE(usuario_ref,''), COALESCE(password_ref,''), COALESCE(estacion_id,0), COALESCE(cargar_en_estaciones,1), COALESCE(orden,0), COALESCE(activa,1), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,''), COALESCE(estado,'activo'), COALESCE(observaciones,'') FROM empresa_camaras `+where+` ORDER BY COALESCE(orden,0), id`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []EmpresaCamara{}
	for rows.Next() {
		var item EmpresaCamara
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.Nombre, &item.Ubicacion, &item.DVRNombre, &item.DVRHost, &item.Canal, &item.Fabricante, &item.Modelo, &item.ProtocoloOrigen, &item.URLStream, &item.URLSnapshot, &item.URLEmbed, &item.VisorTipo, &item.UsuarioRef, &item.PasswordRef, &item.EstacionID, &item.CargarEnEstaciones, &item.Orden, &item.Activa, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador, &item.Estado, &item.Observaciones); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func GetEmpresaCamara(dbConn *sql.DB, empresaID, camaraID int64) (*EmpresaCamara, error) {
	items, err := ListEmpresaCamaras(dbConn, empresaID, true)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].ID == camaraID {
			return &items[i], nil
		}
	}
	return nil, sql.ErrNoRows
}

func DesactivarEmpresaCamara(dbConn *sql.DB, empresaID, camaraID int64, actor string) error {
	if dbConn == nil {
		return fmt.Errorf("conexion de base de datos no disponible")
	}
	if empresaID <= 0 || camaraID <= 0 {
		return fmt.Errorf("empresa_id y camara_id requeridos")
	}
	_, err := execSQLCompat(dbConn, `UPDATE empresa_camaras SET activa=0, estado='inactivo', fecha_actualizacion=CAST(CURRENT_TIMESTAMP AS TEXT), usuario_creador=? WHERE empresa_id=? AND id=?`, strings.TrimSpace(actor), empresaID, camaraID)
	return err
}

func normalizeEmpresaCamaraProtocolo(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case CamaraProtocoloONVIF:
		return CamaraProtocoloONVIF
	case CamaraProtocoloHLS:
		return CamaraProtocoloHLS
	case CamaraProtocoloWebRTC:
		return CamaraProtocoloWebRTC
	case CamaraProtocoloMJPEG:
		return CamaraProtocoloMJPEG
	case CamaraProtocoloIframe:
		return CamaraProtocoloIframe
	default:
		return CamaraProtocoloRTSP
	}
}

func normalizeEmpresaCamaraVisor(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "hls", "webrtc", "mjpeg", "iframe", "rtsp_gateway":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "auto"
	}
}

func normalizeEmpresaCamaraEstado(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "inactivo", "pausado":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "activo"
	}
}
