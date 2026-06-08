package db

import (
	"database/sql"
	"fmt"
	"strings"
)

type EmpresaOCRDocumento struct {
	ID              int64   `json:"id"`
	EmpresaID       int64   `json:"empresa_id"`
	TipoDocumento   string  `json:"tipo_documento"`
	Titulo          string  `json:"titulo"`
	ArchivoNombre   string  `json:"archivo_nombre"`
	ArchivoURL      string  `json:"archivo_url"`
	ArchivoMime     string  `json:"archivo_mime"`
	OCRMotor        string  `json:"ocr_motor"`
	Idioma          string  `json:"idioma"`
	Estado          string  `json:"estado"`
	TextoExtraido   string  `json:"texto_extraido,omitempty"`
	CamposJSON      string  `json:"-"`
	SugerenciasJSON string  `json:"-"`
	Campos          any     `json:"campos,omitempty"`
	Sugerencias     any     `json:"sugerencias,omitempty"`
	Confianza       float64 `json:"confianza"`
	UsuarioCreador  string  `json:"usuario_creador,omitempty"`
	FechaCreacion   string  `json:"fecha_creacion,omitempty"`
}

func EnsureEmpresaOCRSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("conexion de base de datos no disponible")
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_ocr_documentos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			tipo_documento TEXT,
			titulo TEXT,
			archivo_nombre TEXT,
			archivo_url TEXT,
			archivo_mime TEXT,
			ocr_motor TEXT DEFAULT 'tesseract_cli',
			idioma TEXT DEFAULT 'spa+eng',
			estado TEXT DEFAULT 'procesado',
			texto_extraido TEXT,
			campos_json TEXT,
			sugerencias_json TEXT,
			confianza REAL DEFAULT 0,
			usuario_creador TEXT,
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT))
		)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_ocr_documentos_empresa ON empresa_ocr_documentos(empresa_id, fecha_creacion DESC)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_ocr_documentos_tipo ON empresa_ocr_documentos(empresa_id, tipo_documento)`,
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
		{"empresa_ocr_documentos", "tipo_documento", "TEXT"},
		{"empresa_ocr_documentos", "titulo", "TEXT"},
		{"empresa_ocr_documentos", "archivo_nombre", "TEXT"},
		{"empresa_ocr_documentos", "archivo_url", "TEXT"},
		{"empresa_ocr_documentos", "archivo_mime", "TEXT"},
		{"empresa_ocr_documentos", "ocr_motor", "TEXT DEFAULT 'tesseract_cli'"},
		{"empresa_ocr_documentos", "idioma", "TEXT DEFAULT 'spa+eng'"},
		{"empresa_ocr_documentos", "estado", "TEXT DEFAULT 'procesado'"},
		{"empresa_ocr_documentos", "texto_extraido", "TEXT"},
		{"empresa_ocr_documentos", "campos_json", "TEXT"},
		{"empresa_ocr_documentos", "sugerencias_json", "TEXT"},
		{"empresa_ocr_documentos", "confianza", "REAL DEFAULT 0"},
		{"empresa_ocr_documentos", "usuario_creador", "TEXT"},
		{"empresa_ocr_documentos", "fecha_creacion", "TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT))"},
	}
	for _, col := range columns {
		if err := ensureColumnIfMissing(dbConn, col.table, col.name, col.def); err != nil {
			return err
		}
	}
	return nil
}

func InsertEmpresaOCRDocumento(dbConn *sql.DB, item EmpresaOCRDocumento) (int64, error) {
	if item.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id invalido")
	}
	if strings.TrimSpace(item.TipoDocumento) == "" {
		item.TipoDocumento = "general"
	}
	if strings.TrimSpace(item.Estado) == "" {
		item.Estado = "procesado"
	}
	if strings.TrimSpace(item.OCRMotor) == "" {
		item.OCRMotor = "tesseract_cli"
	}
	var id int64
	err := queryRowSQLCompat(dbConn, `INSERT INTO empresa_ocr_documentos (
		empresa_id, tipo_documento, titulo, archivo_nombre, archivo_url, archivo_mime,
		ocr_motor, idioma, estado, texto_extraido, campos_json, sugerencias_json, confianza, usuario_creador
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING id`,
		item.EmpresaID, item.TipoDocumento, item.Titulo, item.ArchivoNombre, item.ArchivoURL, item.ArchivoMime,
		item.OCRMotor, item.Idioma, item.Estado, item.TextoExtraido, item.CamposJSON, item.SugerenciasJSON, item.Confianza, item.UsuarioCreador,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func ListEmpresaOCRDocumentos(dbConn *sql.DB, empresaID int64, limit int) ([]EmpresaOCRDocumento, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id invalido")
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := querySQLCompat(dbConn, `SELECT id, empresa_id, COALESCE(tipo_documento,''), COALESCE(titulo,''), COALESCE(archivo_nombre,''), COALESCE(archivo_url,''), COALESCE(archivo_mime,''), COALESCE(ocr_motor,''), COALESCE(idioma,''), COALESCE(estado,''), COALESCE(confianza,0), COALESCE(usuario_creador,''), COALESCE(fecha_creacion,'')
		FROM empresa_ocr_documentos
		WHERE empresa_id = $1
		ORDER BY id DESC LIMIT $2`, empresaID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []EmpresaOCRDocumento{}
	for rows.Next() {
		var item EmpresaOCRDocumento
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.TipoDocumento, &item.Titulo, &item.ArchivoNombre, &item.ArchivoURL, &item.ArchivoMime, &item.OCRMotor, &item.Idioma, &item.Estado, &item.Confianza, &item.UsuarioCreador, &item.FechaCreacion); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func GetEmpresaOCRDocumento(dbConn *sql.DB, empresaID, id int64) (EmpresaOCRDocumento, error) {
	var item EmpresaOCRDocumento
	err := queryRowSQLCompat(dbConn, `SELECT id, empresa_id, COALESCE(tipo_documento,''), COALESCE(titulo,''), COALESCE(archivo_nombre,''), COALESCE(archivo_url,''), COALESCE(archivo_mime,''), COALESCE(ocr_motor,''), COALESCE(idioma,''), COALESCE(estado,''), COALESCE(texto_extraido,''), COALESCE(campos_json,''), COALESCE(sugerencias_json,''), COALESCE(confianza,0), COALESCE(usuario_creador,''), COALESCE(fecha_creacion,'')
		FROM empresa_ocr_documentos
		WHERE empresa_id = $1 AND id = $2`, empresaID, id).Scan(
		&item.ID, &item.EmpresaID, &item.TipoDocumento, &item.Titulo, &item.ArchivoNombre, &item.ArchivoURL, &item.ArchivoMime, &item.OCRMotor, &item.Idioma, &item.Estado, &item.TextoExtraido, &item.CamposJSON, &item.SugerenciasJSON, &item.Confianza, &item.UsuarioCreador, &item.FechaCreacion,
	)
	if err != nil {
		return EmpresaOCRDocumento{}, err
	}
	return item, nil
}
