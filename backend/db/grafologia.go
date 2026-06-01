package db

import (
	"database/sql"
	"fmt"
	"strings"
)

type EmpresaGrafologiaAnalisis struct {
	ID                   int64   `json:"id"`
	EmpresaID            int64   `json:"empresa_id"`
	Titulo               string  `json:"titulo"`
	ArchivoNombre        string  `json:"archivo_nombre"`
	ImagenURL            string  `json:"imagen_url"`
	ImagenMime           string  `json:"imagen_mime"`
	OCRTexto             string  `json:"ocr_texto,omitempty"`
	OCRMotor             string  `json:"ocr_motor"`
	Estado               string  `json:"estado"`
	Resumen              string  `json:"resumen"`
	MetricasJSON         string  `json:"-"`
	InterpretacionJSON   string  `json:"-"`
	PreprocesamientoJSON string  `json:"-"`
	Metricas             any     `json:"metricas,omitempty"`
	Interpretacion       any     `json:"interpretacion,omitempty"`
	Preprocesamiento     any     `json:"preprocesamiento,omitempty"`
	ReporteHTML          string  `json:"reporte_html,omitempty"`
	ConfianzaGlobal      float64 `json:"confianza_global"`
	UsuarioCreador       string  `json:"usuario_creador,omitempty"`
	FechaCreacion        string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion   string  `json:"fecha_actualizacion,omitempty"`
}

func EnsureEmpresaGrafologiaSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("conexion de base de datos no disponible")
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_grafologia_analisis (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			titulo TEXT,
			archivo_nombre TEXT,
			imagen_url TEXT,
			imagen_mime TEXT,
			ocr_texto TEXT,
			ocr_motor TEXT DEFAULT 'go_heuristico',
			estado TEXT DEFAULT 'completado',
			resumen TEXT,
			metricas_json TEXT,
			interpretacion_json TEXT,
			preprocesamiento_json TEXT,
			reporte_html TEXT,
			confianza_global REAL DEFAULT 0,
			usuario_creador TEXT,
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT))
		)`,
		`CREATE INDEX IF NOT EXISTS ix_grafologia_analisis_empresa ON empresa_grafologia_analisis(empresa_id, fecha_creacion DESC)`,
		`CREATE INDEX IF NOT EXISTS ix_grafologia_analisis_estado ON empresa_grafologia_analisis(empresa_id, estado)`,
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
		{"empresa_grafologia_analisis", "ocr_motor", "TEXT DEFAULT 'go_heuristico'"},
		{"empresa_grafologia_analisis", "reporte_html", "TEXT"},
		{"empresa_grafologia_analisis", "preprocesamiento_json", "TEXT"},
		{"empresa_grafologia_analisis", "confianza_global", "REAL DEFAULT 0"},
		{"empresa_grafologia_analisis", "fecha_actualizacion", "TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT))"},
	}
	for _, col := range columns {
		if err := ensureColumnIfMissing(dbConn, col.table, col.name, col.def); err != nil {
			return err
		}
	}
	return nil
}

func InsertEmpresaGrafologiaAnalisis(dbConn *sql.DB, item EmpresaGrafologiaAnalisis) (int64, error) {
	if item.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id invalido")
	}
	if strings.TrimSpace(item.Titulo) == "" {
		item.Titulo = "Analisis grafológico"
	}
	if strings.TrimSpace(item.Estado) == "" {
		item.Estado = "completado"
	}
	if strings.TrimSpace(item.OCRMotor) == "" {
		item.OCRMotor = "go_heuristico"
	}
	var id int64
	err := queryRowSQLCompat(dbConn, `INSERT INTO empresa_grafologia_analisis (
		empresa_id, titulo, archivo_nombre, imagen_url, imagen_mime, ocr_texto, ocr_motor,
		estado, resumen, metricas_json, interpretacion_json, preprocesamiento_json, reporte_html, confianza_global, usuario_creador
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) RETURNING id`,
		item.EmpresaID, item.Titulo, item.ArchivoNombre, item.ImagenURL, item.ImagenMime, item.OCRTexto, item.OCRMotor,
		item.Estado, item.Resumen, item.MetricasJSON, item.InterpretacionJSON, item.PreprocesamientoJSON, item.ReporteHTML, item.ConfianzaGlobal, item.UsuarioCreador,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func ListEmpresaGrafologiaAnalisis(dbConn *sql.DB, empresaID int64, limit int) ([]EmpresaGrafologiaAnalisis, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id invalido")
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := querySQLCompat(dbConn, `SELECT id, empresa_id, titulo, archivo_nombre, imagen_url, imagen_mime, ocr_motor,
		estado, resumen, confianza_global, usuario_creador, fecha_creacion, fecha_actualizacion
		FROM empresa_grafologia_analisis
		WHERE empresa_id = $1 AND COALESCE(estado,'') <> 'eliminado'
		ORDER BY id DESC LIMIT $2`, empresaID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []EmpresaGrafologiaAnalisis{}
	for rows.Next() {
		var item EmpresaGrafologiaAnalisis
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.Titulo, &item.ArchivoNombre, &item.ImagenURL, &item.ImagenMime, &item.OCRMotor, &item.Estado, &item.Resumen, &item.ConfianzaGlobal, &item.UsuarioCreador, &item.FechaCreacion, &item.FechaActualizacion); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func GetEmpresaGrafologiaAnalisis(dbConn *sql.DB, empresaID, id int64) (EmpresaGrafologiaAnalisis, error) {
	var item EmpresaGrafologiaAnalisis
	err := queryRowSQLCompat(dbConn, `SELECT id, empresa_id, titulo, archivo_nombre, imagen_url, imagen_mime, ocr_texto, ocr_motor,
		estado, resumen, metricas_json, interpretacion_json, COALESCE(preprocesamiento_json,''), reporte_html, confianza_global, usuario_creador, fecha_creacion, fecha_actualizacion
		FROM empresa_grafologia_analisis
		WHERE empresa_id = $1 AND id = $2 AND COALESCE(estado,'') <> 'eliminado'`, empresaID, id).Scan(
		&item.ID, &item.EmpresaID, &item.Titulo, &item.ArchivoNombre, &item.ImagenURL, &item.ImagenMime, &item.OCRTexto, &item.OCRMotor,
		&item.Estado, &item.Resumen, &item.MetricasJSON, &item.InterpretacionJSON, &item.PreprocesamientoJSON, &item.ReporteHTML, &item.ConfianzaGlobal, &item.UsuarioCreador, &item.FechaCreacion, &item.FechaActualizacion,
	)
	if err != nil {
		return EmpresaGrafologiaAnalisis{}, err
	}
	return item, nil
}
