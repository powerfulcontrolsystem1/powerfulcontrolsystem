package db

import (
	"database/sql"
	"fmt"
	"strings"
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
