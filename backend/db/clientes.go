package db

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Cliente representa un tercero/cliente por empresa, preparado para facturación electrónica en Colombia.
type Cliente struct {
	ID                        int64  `json:"id"`
	EmpresaID                 int64  `json:"empresa_id"`
	TipoDocumento             string `json:"tipo_documento"`
	NumeroDocumento           string `json:"numero_documento"`
	DigitoVerificacion        string `json:"digito_verificacion,omitempty"`
	TipoPersona               string `json:"tipo_persona,omitempty"`
	NombreRazonSocial         string `json:"nombre_razon_social"`
	NombreComercial           string `json:"nombre_comercial,omitempty"`
	RegimenFiscal             string `json:"regimen_fiscal,omitempty"`
	ResponsabilidadTributaria string `json:"responsabilidad_tributaria,omitempty"`
	Email                     string `json:"email,omitempty"`
	Telefono                  string `json:"telefono,omitempty"`
	Direccion                 string `json:"direccion,omitempty"`
	Pais                      string `json:"pais,omitempty"`
	Departamento              string `json:"departamento,omitempty"`
	Municipio                 string `json:"municipio,omitempty"`
	CodigoPostal              string `json:"codigo_postal,omitempty"`
	FechaCreacion             string `json:"fecha_creacion,omitempty"`
	FechaActualizacion        string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador            string `json:"usuario_creador,omitempty"`
	Estado                    string `json:"estado,omitempty"`
	Observaciones             string `json:"observaciones,omitempty"`
}

// ClientePerfilComercial representa el perfil analitico del cliente por empresa.
type ClientePerfilComercial struct {
	Cliente           Cliente `json:"cliente"`
	NumeroCompras     int64   `json:"numero_compras"`
	MontoCompras      float64 `json:"monto_compras"`
	TicketPromedio    float64 `json:"ticket_promedio"`
	PrimeraCompra     string  `json:"primera_compra"`
	UltimaCompra      string  `json:"ultima_compra"`
	DiasSinCompra     int     `json:"dias_sin_compra"`
	Segmento          string  `json:"segmento"`
	PerfilActualizado string  `json:"perfil_actualizado"`
}

// ClienteCompraHistorial representa una compra registrada para el historial del cliente.
type ClienteCompraHistorial struct {
	CarritoID      int64   `json:"carrito_id"`
	Codigo         string  `json:"codigo"`
	Nombre         string  `json:"nombre"`
	CanalVenta     string  `json:"canal_venta"`
	Moneda         string  `json:"moneda"`
	FechaOperacion string  `json:"fecha_operacion"`
	EstadoVenta    string  `json:"estado_venta"`
	MontoTotal     float64 `json:"monto_total"`
	ItemsActivos   int64   `json:"items_activos"`
}

// ClienteSegmentacionResumen representa el agregado de clientes por segmento.
type ClienteSegmentacionResumen struct {
	Segmento             string  `json:"segmento"`
	Clientes             int64   `json:"clientes"`
	Compras              int64   `json:"compras"`
	MontoCompras         float64 `json:"monto_compras"`
	TicketPromedioGlobal float64 `json:"ticket_promedio_global"`
}

type clienteComprasMetricas struct {
	NumeroCompras  int64
	MontoCompras   float64
	PrimeraCompra  string
	UltimaCompra   string
	DiasSinCompra  int
	TicketPromedio float64
	Segmento       string
}

// EnsureEmpresaClientesSchema crea y migra la tabla clientes en empresas.db.
func EnsureEmpresaClientesSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS clientes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			tipo_documento TEXT NOT NULL DEFAULT 'NIT',
			numero_documento TEXT NOT NULL,
			digito_verificacion TEXT,
			tipo_persona TEXT DEFAULT 'juridica',
			nombre_razon_social TEXT NOT NULL,
			nombre_comercial TEXT,
			regimen_fiscal TEXT,
			responsabilidad_tributaria TEXT,
			email TEXT,
			telefono TEXT,
			direccion TEXT,
			pais TEXT DEFAULT 'CO',
			departamento TEXT,
			municipio TEXT,
			codigo_postal TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, tipo_documento, numero_documento)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_clientes_empresa_nombre ON clientes(empresa_id, nombre_razon_social);`,
		`CREATE INDEX IF NOT EXISTS ix_clientes_empresa_documento ON clientes(empresa_id, tipo_documento, numero_documento);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureClientesColumns(dbConn); err != nil {
		return err
	}
	return nil
}

func ensureClientesColumns(dbConn *sql.DB) error {
	rows, err := dbConn.Query(`PRAGMA table_info(clientes);`)
	if err != nil {
		return err
	}
	defer rows.Close()

	existing := map[string]bool{}
	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return err
		}
		existing[name] = true
	}

	addIfMissing := func(colDef, name string) error {
		if existing[name] {
			return nil
		}
		q := fmt.Sprintf("ALTER TABLE clientes ADD COLUMN %s;", colDef)
		_, err := dbConn.Exec(q)
		return err
	}

	if err := addIfMissing("tipo_documento TEXT DEFAULT 'NIT'", "tipo_documento"); err != nil {
		return err
	}
	if err := addIfMissing("numero_documento TEXT", "numero_documento"); err != nil {
		return err
	}
	if err := addIfMissing("digito_verificacion TEXT", "digito_verificacion"); err != nil {
		return err
	}
	if err := addIfMissing("tipo_persona TEXT DEFAULT 'juridica'", "tipo_persona"); err != nil {
		return err
	}
	if err := addIfMissing("nombre_razon_social TEXT", "nombre_razon_social"); err != nil {
		return err
	}
	if err := addIfMissing("nombre_comercial TEXT", "nombre_comercial"); err != nil {
		return err
	}
	if err := addIfMissing("regimen_fiscal TEXT", "regimen_fiscal"); err != nil {
		return err
	}
	if err := addIfMissing("responsabilidad_tributaria TEXT", "responsabilidad_tributaria"); err != nil {
		return err
	}
	if err := addIfMissing("email TEXT", "email"); err != nil {
		return err
	}
	if err := addIfMissing("telefono TEXT", "telefono"); err != nil {
		return err
	}
	if err := addIfMissing("direccion TEXT", "direccion"); err != nil {
		return err
	}
	if err := addIfMissing("pais TEXT DEFAULT 'CO'", "pais"); err != nil {
		return err
	}
	if err := addIfMissing("departamento TEXT", "departamento"); err != nil {
		return err
	}
	if err := addIfMissing("municipio TEXT", "municipio"); err != nil {
		return err
	}
	if err := addIfMissing("codigo_postal TEXT", "codigo_postal"); err != nil {
		return err
	}
	if err := addIfMissing("fecha_actualizacion TEXT", "fecha_actualizacion"); err != nil {
		return err
	}
	if err := addIfMissing("usuario_creador TEXT", "usuario_creador"); err != nil {
		return err
	}
	if err := addIfMissing("estado TEXT DEFAULT 'activo'", "estado"); err != nil {
		return err
	}
	if err := addIfMissing("observaciones TEXT", "observaciones"); err != nil {
		return err
	}
	return nil
}

// CreateCliente crea un cliente para una empresa.
func CreateCliente(dbConn *sql.DB, payload Cliente) (int64, error) {
	if strings.TrimSpace(payload.TipoDocumento) == "" {
		payload.TipoDocumento = "NIT"
	}
	if strings.TrimSpace(payload.Pais) == "" {
		payload.Pais = "CO"
	}
	res, err := dbConn.Exec(`INSERT INTO clientes (
		empresa_id,
		tipo_documento,
		numero_documento,
		digito_verificacion,
		tipo_persona,
		nombre_razon_social,
		nombre_comercial,
		regimen_fiscal,
		responsabilidad_tributaria,
		email,
		telefono,
		direccion,
		pais,
		departamento,
		municipio,
		codigo_postal,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'activo', ?, datetime('now','localtime'), datetime('now','localtime'))`,
		payload.EmpresaID,
		strings.TrimSpace(payload.TipoDocumento),
		strings.TrimSpace(payload.NumeroDocumento),
		strings.TrimSpace(payload.DigitoVerificacion),
		strings.TrimSpace(payload.TipoPersona),
		strings.TrimSpace(payload.NombreRazonSocial),
		strings.TrimSpace(payload.NombreComercial),
		strings.TrimSpace(payload.RegimenFiscal),
		strings.TrimSpace(payload.ResponsabilidadTributaria),
		strings.TrimSpace(payload.Email),
		strings.TrimSpace(payload.Telefono),
		strings.TrimSpace(payload.Direccion),
		strings.TrimSpace(payload.Pais),
		strings.TrimSpace(payload.Departamento),
		strings.TrimSpace(payload.Municipio),
		strings.TrimSpace(payload.CodigoPostal),
		strings.TrimSpace(payload.UsuarioCreador),
		strings.TrimSpace(payload.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetClientesByEmpresa lista clientes por empresa.
func GetClientesByEmpresa(dbConn *sql.DB, empresaID int64, includeInactive bool, q string) ([]Cliente, error) {
	query := `SELECT
		id,
		empresa_id,
		COALESCE(tipo_documento, 'NIT'),
		COALESCE(numero_documento, ''),
		COALESCE(digito_verificacion, ''),
		COALESCE(tipo_persona, 'juridica'),
		COALESCE(nombre_razon_social, ''),
		COALESCE(nombre_comercial, ''),
		COALESCE(regimen_fiscal, ''),
		COALESCE(responsabilidad_tributaria, ''),
		COALESCE(email, ''),
		COALESCE(telefono, ''),
		COALESCE(direccion, ''),
		COALESCE(pais, 'CO'),
		COALESCE(departamento, ''),
		COALESCE(municipio, ''),
		COALESCE(codigo_postal, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM clientes
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	if !includeInactive {
		query += ` AND COALESCE(estado, 'activo') = 'activo'`
	}
	q = strings.TrimSpace(q)
	if q != "" {
		query += ` AND (
			lower(COALESCE(nombre_razon_social, '')) LIKE lower(?) OR
			lower(COALESCE(nombre_comercial, '')) LIKE lower(?) OR
			lower(COALESCE(numero_documento, '')) LIKE lower(?)
		)`
		pat := "%" + q + "%"
		args = append(args, pat, pat, pat)
	}
	query += ` ORDER BY id DESC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Cliente, 0)
	for rows.Next() {
		var item Cliente
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.TipoDocumento,
			&item.NumeroDocumento,
			&item.DigitoVerificacion,
			&item.TipoPersona,
			&item.NombreRazonSocial,
			&item.NombreComercial,
			&item.RegimenFiscal,
			&item.ResponsabilidadTributaria,
			&item.Email,
			&item.Telefono,
			&item.Direccion,
			&item.Pais,
			&item.Departamento,
			&item.Municipio,
			&item.CodigoPostal,
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
	return out, nil
}

// GetClienteByID devuelve un cliente puntual por empresa.
func GetClienteByID(dbConn *sql.DB, empresaID, clienteID int64) (*Cliente, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id invalido")
	}
	if clienteID <= 0 {
		return nil, fmt.Errorf("cliente_id invalido")
	}

	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(tipo_documento, 'NIT'),
		COALESCE(numero_documento, ''),
		COALESCE(digito_verificacion, ''),
		COALESCE(tipo_persona, 'juridica'),
		COALESCE(nombre_razon_social, ''),
		COALESCE(nombre_comercial, ''),
		COALESCE(regimen_fiscal, ''),
		COALESCE(responsabilidad_tributaria, ''),
		COALESCE(email, ''),
		COALESCE(telefono, ''),
		COALESCE(direccion, ''),
		COALESCE(pais, 'CO'),
		COALESCE(departamento, ''),
		COALESCE(municipio, ''),
		COALESCE(codigo_postal, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM clientes
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, clienteID)

	var item Cliente
	if err := row.Scan(
		&item.ID,
		&item.EmpresaID,
		&item.TipoDocumento,
		&item.NumeroDocumento,
		&item.DigitoVerificacion,
		&item.TipoPersona,
		&item.NombreRazonSocial,
		&item.NombreComercial,
		&item.RegimenFiscal,
		&item.ResponsabilidadTributaria,
		&item.Email,
		&item.Telefono,
		&item.Direccion,
		&item.Pais,
		&item.Departamento,
		&item.Municipio,
		&item.CodigoPostal,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	); err != nil {
		return nil, err
	}

	return &item, nil
}

// UpdateCliente actualiza un cliente por empresa.
func UpdateCliente(dbConn *sql.DB, payload Cliente) error {
	if strings.TrimSpace(payload.TipoDocumento) == "" {
		payload.TipoDocumento = "NIT"
	}
	if strings.TrimSpace(payload.Pais) == "" {
		payload.Pais = "CO"
	}
	_, err := dbConn.Exec(`UPDATE clientes SET
		tipo_documento = ?,
		numero_documento = ?,
		digito_verificacion = ?,
		tipo_persona = ?,
		nombre_razon_social = ?,
		nombre_comercial = ?,
		regimen_fiscal = ?,
		responsabilidad_tributaria = ?,
		email = ?,
		telefono = ?,
		direccion = ?,
		pais = ?,
		departamento = ?,
		municipio = ?,
		codigo_postal = ?,
		observaciones = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE id = ? AND empresa_id = ?`,
		strings.TrimSpace(payload.TipoDocumento),
		strings.TrimSpace(payload.NumeroDocumento),
		strings.TrimSpace(payload.DigitoVerificacion),
		strings.TrimSpace(payload.TipoPersona),
		strings.TrimSpace(payload.NombreRazonSocial),
		strings.TrimSpace(payload.NombreComercial),
		strings.TrimSpace(payload.RegimenFiscal),
		strings.TrimSpace(payload.ResponsabilidadTributaria),
		strings.TrimSpace(payload.Email),
		strings.TrimSpace(payload.Telefono),
		strings.TrimSpace(payload.Direccion),
		strings.TrimSpace(payload.Pais),
		strings.TrimSpace(payload.Departamento),
		strings.TrimSpace(payload.Municipio),
		strings.TrimSpace(payload.CodigoPostal),
		strings.TrimSpace(payload.Observaciones),
		payload.ID,
		payload.EmpresaID,
	)
	return err
}

// DeleteCliente elimina un cliente por empresa.
func DeleteCliente(dbConn *sql.DB, empresaID, id int64) error {
	_, err := dbConn.Exec(`DELETE FROM clientes WHERE empresa_id = ? AND id = ?`, empresaID, id)
	return err
}

// SetClienteEstado activa o desactiva un cliente por empresa.
func SetClienteEstado(dbConn *sql.DB, empresaID, id int64, estado string) error {
	_, err := dbConn.Exec(`UPDATE clientes SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ? AND id = ?`, estado, empresaID, id)
	return err
}

// GetClientePerfilComercialByEmpresa devuelve perfil e indicadores de compra para un cliente.
func GetClientePerfilComercialByEmpresa(dbConn *sql.DB, empresaID, clienteID int64) (ClientePerfilComercial, error) {
	if empresaID <= 0 {
		return ClientePerfilComercial{}, fmt.Errorf("empresa_id invalido")
	}
	if clienteID <= 0 {
		return ClientePerfilComercial{}, fmt.Errorf("cliente_id invalido")
	}

	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(tipo_documento, 'NIT'),
		COALESCE(numero_documento, ''),
		COALESCE(digito_verificacion, ''),
		COALESCE(tipo_persona, 'juridica'),
		COALESCE(nombre_razon_social, ''),
		COALESCE(nombre_comercial, ''),
		COALESCE(regimen_fiscal, ''),
		COALESCE(responsabilidad_tributaria, ''),
		COALESCE(email, ''),
		COALESCE(telefono, ''),
		COALESCE(direccion, ''),
		COALESCE(pais, 'CO'),
		COALESCE(departamento, ''),
		COALESCE(municipio, ''),
		COALESCE(codigo_postal, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM clientes
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, clienteID)

	var cliente Cliente
	if err := row.Scan(
		&cliente.ID,
		&cliente.EmpresaID,
		&cliente.TipoDocumento,
		&cliente.NumeroDocumento,
		&cliente.DigitoVerificacion,
		&cliente.TipoPersona,
		&cliente.NombreRazonSocial,
		&cliente.NombreComercial,
		&cliente.RegimenFiscal,
		&cliente.ResponsabilidadTributaria,
		&cliente.Email,
		&cliente.Telefono,
		&cliente.Direccion,
		&cliente.Pais,
		&cliente.Departamento,
		&cliente.Municipio,
		&cliente.CodigoPostal,
		&cliente.FechaCreacion,
		&cliente.FechaActualizacion,
		&cliente.UsuarioCreador,
		&cliente.Estado,
		&cliente.Observaciones,
	); err != nil {
		return ClientePerfilComercial{}, err
	}

	metricas, err := getClienteComprasMetricas(dbConn, empresaID, clienteID)
	if err != nil {
		return ClientePerfilComercial{}, err
	}

	return ClientePerfilComercial{
		Cliente:           cliente,
		NumeroCompras:     metricas.NumeroCompras,
		MontoCompras:      metricas.MontoCompras,
		TicketPromedio:    metricas.TicketPromedio,
		PrimeraCompra:     metricas.PrimeraCompra,
		UltimaCompra:      metricas.UltimaCompra,
		DiasSinCompra:     metricas.DiasSinCompra,
		Segmento:          metricas.Segmento,
		PerfilActualizado: time.Now().In(time.Local).Format("2006-01-02 15:04:05"),
	}, nil
}

// GetClienteHistorialComprasByEmpresa lista compras del cliente en orden descendente por fecha.
func GetClienteHistorialComprasByEmpresa(dbConn *sql.DB, empresaID, clienteID int64, limit int) ([]ClienteCompraHistorial, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id invalido")
	}
	if clienteID <= 0 {
		return nil, fmt.Errorf("cliente_id invalido")
	}
	if limit <= 0 || limit > 200 {
		limit = 30
	}

	rows, err := dbConn.Query(`SELECT
		c.id,
		COALESCE(c.codigo, ''),
		COALESCE(c.nombre, ''),
		COALESCE(c.canal_venta, 'mostrador'),
		COALESCE(c.moneda, 'COP'),
		COALESCE(NULLIF(c.pagado_en, ''), NULLIF(c.fecha_actualizacion, ''), COALESCE(c.fecha_creacion, '')),
		COALESCE(c.estado_carrito, 'abierto'),
		COALESCE(c.estado, 'activo'),
		COALESCE(c.pagado_en, ''),
		COALESCE(CASE WHEN c.total_pagado > 0 THEN c.total_pagado ELSE c.total END, 0),
		COALESCE((
			SELECT COUNT(1)
			FROM carrito_compra_items i
			WHERE i.empresa_id = c.empresa_id
				AND i.carrito_id = c.id
				AND COALESCE(i.estado, 'activo') = 'activo'
		), 0)
	FROM carritos_compras c
	WHERE c.empresa_id = ?
		AND COALESCE(c.estado, 'activo') = 'activo'
		AND COALESCE(c.cliente_id, 0) = ?
	ORDER BY datetime(COALESCE(NULLIF(c.pagado_en, ''), NULLIF(c.fecha_actualizacion, ''), COALESCE(c.fecha_creacion, ''))) DESC, c.id DESC
	LIMIT ?`, empresaID, clienteID, limit)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such table") {
			return []ClienteCompraHistorial{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	out := make([]ClienteCompraHistorial, 0)
	for rows.Next() {
		var item ClienteCompraHistorial
		var estadoCarrito string
		var estadoRegistro string
		var pagadoEn string
		if err := rows.Scan(
			&item.CarritoID,
			&item.Codigo,
			&item.Nombre,
			&item.CanalVenta,
			&item.Moneda,
			&item.FechaOperacion,
			&estadoCarrito,
			&estadoRegistro,
			&pagadoEn,
			&item.MontoTotal,
			&item.ItemsActivos,
		); err != nil {
			return nil, err
		}
		item.EstadoVenta = resolveCarritoEstadoVenta(estadoCarrito, estadoRegistro, pagadoEn)
		out = append(out, item)
	}

	return out, nil
}

// GetClientesSegmentacionByEmpresa devuelve el consolidado de clientes por segmento.
func GetClientesSegmentacionByEmpresa(dbConn *sql.DB, empresaID int64, includeInactive bool, q string) ([]ClienteSegmentacionResumen, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id invalido")
	}

	clientes, err := GetClientesByEmpresa(dbConn, empresaID, includeInactive, q)
	if err != nil {
		return nil, err
	}
	if len(clientes) == 0 {
		return []ClienteSegmentacionResumen{}, nil
	}

	agg := map[string]*ClienteSegmentacionResumen{}
	for _, cliente := range clientes {
		metricas, err := getClienteComprasMetricas(dbConn, empresaID, cliente.ID)
		if err != nil {
			return nil, err
		}
		segmento := metricas.Segmento
		if segmento == "" {
			segmento = "nuevo"
		}

		group := agg[segmento]
		if group == nil {
			group = &ClienteSegmentacionResumen{Segmento: segmento}
			agg[segmento] = group
		}
		group.Clientes++
		group.Compras += metricas.NumeroCompras
		group.MontoCompras += metricas.MontoCompras
	}

	out := make([]ClienteSegmentacionResumen, 0, len(agg))
	for _, row := range agg {
		if row.Compras > 0 {
			row.TicketPromedioGlobal = row.MontoCompras / float64(row.Compras)
		}
		out = append(out, *row)
	}

	priority := map[string]int{
		"estrategico": 0,
		"frecuente":   1,
		"activo":      2,
		"nuevo":       3,
		"inactivo":    4,
	}

	sort.Slice(out, func(i, j int) bool {
		pi, okI := priority[out[i].Segmento]
		pj, okJ := priority[out[j].Segmento]
		if !okI {
			pi = 99
		}
		if !okJ {
			pj = 99
		}
		if pi != pj {
			return pi < pj
		}
		if out[i].Clientes != out[j].Clientes {
			return out[i].Clientes > out[j].Clientes
		}
		return out[i].Segmento < out[j].Segmento
	})

	return out, nil
}

func getClienteComprasMetricas(dbConn *sql.DB, empresaID, clienteID int64) (clienteComprasMetricas, error) {
	row := dbConn.QueryRow(`SELECT
		COUNT(1),
		COALESCE(SUM(CASE WHEN total_pagado > 0 THEN total_pagado ELSE total END), 0),
		COALESCE(MIN(COALESCE(NULLIF(pagado_en, ''), NULLIF(fecha_actualizacion, ''), COALESCE(fecha_creacion, ''))), ''),
		COALESCE(MAX(COALESCE(NULLIF(pagado_en, ''), NULLIF(fecha_actualizacion, ''), COALESCE(fecha_creacion, ''))), '')
	FROM carritos_compras
	WHERE empresa_id = ?
		AND COALESCE(estado, 'activo') = 'activo'
		AND COALESCE(cliente_id, 0) = ?
		AND (
			COALESCE(estado_carrito, 'abierto') = 'cerrado'
			OR COALESCE(total_pagado, 0) > 0
		)`, empresaID, clienteID)

	var m clienteComprasMetricas
	if err := row.Scan(&m.NumeroCompras, &m.MontoCompras, &m.PrimeraCompra, &m.UltimaCompra); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such table") {
			m.Segmento = "nuevo"
			return m, nil
		}
		return clienteComprasMetricas{}, err
	}
	if m.NumeroCompras > 0 {
		m.TicketPromedio = m.MontoCompras / float64(m.NumeroCompras)
	}
	if last, ok := parseSQLiteDate(m.UltimaCompra); ok {
		delta := int(time.Since(last).Hours() / 24)
		if delta < 0 {
			delta = 0
		}
		m.DiasSinCompra = delta
	}
	m.Segmento = resolveClienteSegmento(m.NumeroCompras, m.MontoCompras, m.DiasSinCompra)
	return m, nil
}

func resolveClienteSegmento(numeroCompras int64, montoCompras float64, diasSinCompra int) string {
	switch {
	case numeroCompras <= 0:
		return "nuevo"
	case montoCompras >= 3000000 || numeroCompras >= 8:
		return "estrategico"
	case diasSinCompra <= 45 && numeroCompras >= 3:
		return "frecuente"
	case diasSinCompra <= 120:
		return "activo"
	default:
		return "inactivo"
	}
}

func parseSQLiteDate(raw string) (time.Time, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, false
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if ts, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return ts, true
		}
	}
	return time.Time{}, false
}
