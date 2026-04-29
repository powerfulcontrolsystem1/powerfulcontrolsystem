package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// TipoEmpresaPreconfiguracion define la plantilla que se aplica al crear una empresa por tipo.
type TipoEmpresaPreconfiguracion struct {
	ID                 int64  `json:"id"`
	TipoEmpresaID      int64  `json:"tipo_empresa_id"`
	TipoEmpresaNombre  string `json:"tipo_empresa_nombre,omitempty"`
	Enabled            bool   `json:"enabled"`
	Nombre             string `json:"nombre"`
	Descripcion        string `json:"descripcion,omitempty"`
	ConfigJSON         string `json:"config_json"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
}

type TipoEmpresaPreconfigTemplate struct {
	Estaciones TipoEmpresaPreconfigEstaciones `json:"estaciones"`
	Productos  []TipoEmpresaPreconfigProducto `json:"productos"`
}

type TipoEmpresaPreconfigEstaciones struct {
	Enabled     bool   `json:"enabled"`
	Cantidad    int    `json:"cantidad"`
	Prefijo     string `json:"prefijo"`
	CardSize    string `json:"card_size"`
	CajaEnabled bool   `json:"caja_enabled"`
}

type TipoEmpresaPreconfigProducto struct {
	SKU                  string  `json:"sku"`
	Nombre               string  `json:"nombre"`
	Categoria            string  `json:"categoria,omitempty"`
	Descripcion          string  `json:"descripcion,omitempty"`
	UnidadMedida         string  `json:"unidad_medida,omitempty"`
	Costo                float64 `json:"costo"`
	Precio               float64 `json:"precio"`
	ImpuestoPorcentaje   float64 `json:"impuesto_porcentaje"`
	StockMinimo          float64 `json:"stock_minimo"`
	StockInicial         float64 `json:"stock_inicial"`
	ReferenciaInventario string  `json:"referencia_inventario,omitempty"`
}

// EnsureTipoEmpresaPreconfiguracionSchema crea/migra la tabla de plantillas por tipo de empresa.
func EnsureTipoEmpresaPreconfiguracionSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS tipo_empresa_preconfiguraciones (
			id BIGSERIAL PRIMARY KEY,
			tipo_empresa_id BIGINT NOT NULL UNIQUE,
			enabled INTEGER DEFAULT 0,
			nombre TEXT,
			descripcion TEXT,
			config_json TEXT,
			fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo'
		);`,
		`CREATE INDEX IF NOT EXISTS ix_tipo_empresa_preconfiguraciones_tipo ON tipo_empresa_preconfiguraciones(tipo_empresa_id);`,
	}
	for _, stmt := range stmts {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	for _, col := range []struct {
		name string
		def  string
	}{
		{"enabled", "INTEGER DEFAULT 0"},
		{"nombre", "TEXT"},
		{"descripcion", "TEXT"},
		{"config_json", "TEXT"},
		{"fecha_actualizacion", "TEXT"},
		{"usuario_creador", "TEXT"},
		{"estado", "TEXT DEFAULT 'activo'"},
	} {
		if err := ensureColumnIfMissing(dbConn, "tipo_empresa_preconfiguraciones", col.name, col.def); err != nil {
			return err
		}
	}
	return nil
}

func scanTipoEmpresaPreconfiguracion(row scanner) (*TipoEmpresaPreconfiguracion, error) {
	var item TipoEmpresaPreconfiguracion
	var enabled int
	if err := row.Scan(
		&item.ID,
		&item.TipoEmpresaID,
		&item.TipoEmpresaNombre,
		&enabled,
		&item.Nombre,
		&item.Descripcion,
		&item.ConfigJSON,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
	); err != nil {
		return nil, err
	}
	item.Enabled = enabled == 1
	return &item, nil
}

type scanner interface {
	Scan(dest ...interface{}) error
}

// ListTipoEmpresaPreconfiguraciones devuelve las plantillas guardadas.
func ListTipoEmpresaPreconfiguraciones(dbConn *sql.DB) ([]TipoEmpresaPreconfiguracion, error) {
	if err := EnsureTipoEmpresaPreconfiguracionSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		p.id, p.tipo_empresa_id, COALESCE(t.nombre, ''), COALESCE(p.enabled, 0),
		COALESCE(p.nombre, ''), COALESCE(p.descripcion, ''), COALESCE(p.config_json, ''),
		COALESCE(p.fecha_creacion, ''), COALESCE(p.fecha_actualizacion, ''),
		COALESCE(p.usuario_creador, ''), COALESCE(NULLIF(TRIM(p.estado), ''), 'activo')
	FROM tipo_empresa_preconfiguraciones p
	LEFT JOIN tipos_de_empresas t ON t.id = p.tipo_empresa_id
	ORDER BY p.tipo_empresa_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]TipoEmpresaPreconfiguracion, 0)
	for rows.Next() {
		item, err := scanTipoEmpresaPreconfiguracion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

// GetTipoEmpresaPreconfiguracionByTipoID devuelve una plantilla por tipo, o nil si no existe.
func GetTipoEmpresaPreconfiguracionByTipoID(dbConn *sql.DB, tipoEmpresaID int64) (*TipoEmpresaPreconfiguracion, error) {
	if tipoEmpresaID <= 0 {
		return nil, nil
	}
	if err := EnsureTipoEmpresaPreconfiguracionSchema(dbConn); err != nil {
		return nil, err
	}
	row := queryRowSQLCompat(dbConn, `SELECT
		p.id, p.tipo_empresa_id, COALESCE(t.nombre, ''), COALESCE(p.enabled, 0),
		COALESCE(p.nombre, ''), COALESCE(p.descripcion, ''), COALESCE(p.config_json, ''),
		COALESCE(p.fecha_creacion, ''), COALESCE(p.fecha_actualizacion, ''),
		COALESCE(p.usuario_creador, ''), COALESCE(NULLIF(TRIM(p.estado), ''), 'activo')
	FROM tipo_empresa_preconfiguraciones p
	LEFT JOIN tipos_de_empresas t ON t.id = p.tipo_empresa_id
	WHERE p.tipo_empresa_id = ? LIMIT 1`, tipoEmpresaID)
	item, err := scanTipoEmpresaPreconfiguracion(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return item, err
}

// UpsertTipoEmpresaPreconfiguracion crea o actualiza una plantilla por tipo de empresa.
func UpsertTipoEmpresaPreconfiguracion(dbConn *sql.DB, item TipoEmpresaPreconfiguracion) (int64, error) {
	if item.TipoEmpresaID <= 0 {
		return 0, errors.New("tipo_empresa_id invalido")
	}
	if err := EnsureTipoEmpresaPreconfiguracionSchema(dbConn); err != nil {
		return 0, err
	}
	enabled := 0
	if item.Enabled {
		enabled = 1
	}
	item.Nombre = strings.TrimSpace(item.Nombre)
	if item.Nombre == "" {
		item.Nombre = "Preconfiguracion inicial"
	}
	item.Estado = strings.ToLower(strings.TrimSpace(item.Estado))
	if item.Estado == "" {
		item.Estado = "activo"
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO tipo_empresa_preconfiguraciones (
		tipo_empresa_id, enabled, nombre, descripcion, config_json,
		fecha_creacion, fecha_actualizacion, usuario_creador, estado
	) VALUES (
		?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?
	) ON CONFLICT(tipo_empresa_id) DO UPDATE SET
		enabled = excluded.enabled,
		nombre = excluded.nombre,
		descripcion = excluded.descripcion,
		config_json = excluded.config_json,
		fecha_actualizacion = CURRENT_TIMESTAMP,
		usuario_creador = CASE WHEN trim(excluded.usuario_creador) <> '' THEN excluded.usuario_creador ELSE tipo_empresa_preconfiguraciones.usuario_creador END,
		estado = COALESCE(NULLIF(TRIM(excluded.estado), ''), 'activo')
	RETURNING id`,
		item.TipoEmpresaID,
		enabled,
		item.Nombre,
		strings.TrimSpace(item.Descripcion),
		strings.TrimSpace(item.ConfigJSON),
		strings.TrimSpace(item.UsuarioCreador),
		item.Estado,
	)
	if err != nil {
		return 0, fmt.Errorf("upsert tipo empresa preconfiguracion: %w", err)
	}
	return id, nil
}

// DefaultTipoEmpresaPreconfiguracion entrega una plantilla profesional sugerida para tipos conocidos.
func DefaultTipoEmpresaPreconfiguracion(tipoEmpresaID int64, tipoNombre string) TipoEmpresaPreconfiguracion {
	template := DefaultTipoEmpresaPreconfigTemplate(tipoNombre)
	raw, _ := json.Marshal(template)
	enabled := len(template.Productos) > 0 || template.Estaciones.Cantidad > 0
	nombre := "Preconfiguracion inicial"
	if isTipoEmpresaRestaurante(tipoNombre) {
		nombre = "Restaurante basico"
	}
	return TipoEmpresaPreconfiguracion{
		TipoEmpresaID:     tipoEmpresaID,
		TipoEmpresaNombre: strings.TrimSpace(tipoNombre),
		Enabled:           enabled,
		Nombre:            nombre,
		Descripcion:       "Plantilla inicial aplicada automaticamente al crear empresas nuevas de este tipo.",
		ConfigJSON:        string(raw),
		Estado:            "activo",
	}
}

// ResolveTipoEmpresaPreconfiguracion devuelve la configuracion guardada o la sugerida por defecto.
func ResolveTipoEmpresaPreconfiguracion(dbConn *sql.DB, tipoEmpresaID int64, tipoNombre string) (*TipoEmpresaPreconfiguracion, error) {
	if tipoEmpresaID > 0 {
		saved, err := GetTipoEmpresaPreconfiguracionByTipoID(dbConn, tipoEmpresaID)
		if err != nil {
			return nil, err
		}
		if saved != nil && strings.ToLower(strings.TrimSpace(saved.Estado)) != "inactivo" {
			if strings.TrimSpace(saved.TipoEmpresaNombre) == "" {
				saved.TipoEmpresaNombre = tipoNombre
			}
			return saved, nil
		}
	}
	def := DefaultTipoEmpresaPreconfiguracion(tipoEmpresaID, tipoNombre)
	return &def, nil
}

func DefaultTipoEmpresaPreconfigTemplate(tipoNombre string) TipoEmpresaPreconfigTemplate {
	if !isTipoEmpresaRestaurante(tipoNombre) {
		return TipoEmpresaPreconfigTemplate{}
	}
	return NormalizeTipoEmpresaPreconfigTemplate(TipoEmpresaPreconfigTemplate{
		Estaciones: TipoEmpresaPreconfigEstaciones{
			Enabled:     true,
			Cantidad:    5,
			Prefijo:     "Mesa",
			CardSize:    "medium",
			CajaEnabled: true,
		},
		Productos: []TipoEmpresaPreconfigProducto{
			{SKU: "DEMO-REST-001", Nombre: "Hamburguesa clasica", Categoria: "Comidas", UnidadMedida: "unidad", Costo: 9000, Precio: 18000, ImpuestoPorcentaje: 0, StockMinimo: 5},
			{SKU: "DEMO-REST-002", Nombre: "Perro caliente", Categoria: "Comidas", UnidadMedida: "unidad", Costo: 6000, Precio: 12000, ImpuestoPorcentaje: 0, StockMinimo: 5},
			{SKU: "DEMO-REST-003", Nombre: "Gaseosa personal", Categoria: "Bebidas", UnidadMedida: "unidad", Costo: 2200, Precio: 4000, ImpuestoPorcentaje: 0, StockMinimo: 12},
			{SKU: "DEMO-REST-004", Nombre: "Agua botella", Categoria: "Bebidas", UnidadMedida: "unidad", Costo: 1800, Precio: 3500, ImpuestoPorcentaje: 0, StockMinimo: 12},
			{SKU: "DEMO-REST-005", Nombre: "Menu del dia", Categoria: "Almuerzos", UnidadMedida: "unidad", Costo: 12000, Precio: 22000, ImpuestoPorcentaje: 0, StockMinimo: 3},
		},
	})
}

func ParseTipoEmpresaPreconfigTemplate(raw string) (TipoEmpresaPreconfigTemplate, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return TipoEmpresaPreconfigTemplate{}, nil
	}
	var template TipoEmpresaPreconfigTemplate
	if err := json.Unmarshal([]byte(raw), &template); err != nil {
		return TipoEmpresaPreconfigTemplate{}, err
	}
	return NormalizeTipoEmpresaPreconfigTemplate(template), nil
}

func MarshalTipoEmpresaPreconfigTemplate(template TipoEmpresaPreconfigTemplate) (string, error) {
	normalized := NormalizeTipoEmpresaPreconfigTemplate(template)
	raw, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func NormalizeTipoEmpresaPreconfigTemplate(template TipoEmpresaPreconfigTemplate) TipoEmpresaPreconfigTemplate {
	if template.Estaciones.Cantidad < 0 {
		template.Estaciones.Cantidad = 0
	}
	if template.Estaciones.Cantidad > 200 {
		template.Estaciones.Cantidad = 200
	}
	template.Estaciones.Prefijo = strings.TrimSpace(template.Estaciones.Prefijo)
	if template.Estaciones.Prefijo == "" {
		template.Estaciones.Prefijo = "Estacion"
	}
	template.Estaciones.CardSize = strings.ToLower(strings.TrimSpace(template.Estaciones.CardSize))
	if template.Estaciones.CardSize == "" {
		template.Estaciones.CardSize = "medium"
	}
	if !template.Estaciones.Enabled {
		template.Estaciones.Cantidad = 0
	}

	productos := make([]TipoEmpresaPreconfigProducto, 0, len(template.Productos))
	seenSKU := map[string]bool{}
	for idx, p := range template.Productos {
		p.Nombre = strings.TrimSpace(p.Nombre)
		if p.Nombre == "" {
			continue
		}
		p.SKU = strings.ToUpper(strings.TrimSpace(p.SKU))
		if p.SKU == "" {
			p.SKU = fmt.Sprintf("DEMO-%03d", idx+1)
		}
		if seenSKU[p.SKU] {
			continue
		}
		seenSKU[p.SKU] = true
		p.Categoria = strings.TrimSpace(p.Categoria)
		p.Descripcion = strings.TrimSpace(p.Descripcion)
		p.UnidadMedida = strings.TrimSpace(p.UnidadMedida)
		if p.UnidadMedida == "" {
			p.UnidadMedida = "unidad"
		}
		if p.Precio < 0 {
			p.Precio = 0
		}
		if p.Costo < 0 {
			p.Costo = 0
		}
		if p.StockMinimo < 0 {
			p.StockMinimo = 0
		}
		if p.StockInicial < 0 {
			p.StockInicial = 0
		}
		productos = append(productos, p)
	}
	template.Productos = productos
	return template
}

func isTipoEmpresaRestaurante(tipoNombre string) bool {
	n := strings.ToLower(strings.TrimSpace(tipoNombre))
	for _, token := range []string{"restaurante", "restaurant", "comida", "bar", "cafeteria", "cafetería"} {
		if strings.Contains(n, token) {
			return true
		}
	}
	return false
}
