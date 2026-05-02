package db

import (
	"database/sql"
	"strings"
	"sync"
)

var (
	asesorComercialSchemaMu    sync.Mutex
	asesorComercialSchemaReady bool
)

// AsesorComercial representa al administrador aceptado para operar como asesor comercial.
type AsesorComercial struct {
	ID                 int64   `json:"id"`
	AdminEmail         string  `json:"admin_email"`
	AdminNombre        string  `json:"admin_nombre"`
	Codigo             string  `json:"codigo"`
	PorcentajeComision float64 `json:"porcentaje_comision"`
	MesesAsociacion    int     `json:"meses_asociacion"`
	EstadoInvitacion   string  `json:"estado_invitacion"`
	InvitadoPorEmail   string  `json:"invitado_por_email"`
	InvitacionExpiraEn string  `json:"invitacion_expira_en"`
	AceptadoEn         string  `json:"aceptado_en"`
	FechaCreacion      string  `json:"fecha_creacion"`
	FechaActualizacion string  `json:"fecha_actualizacion"`
	Estado             string  `json:"estado"`
	Observaciones      string  `json:"observaciones"`
}

// AsesorComercialComision registra una venta/renovacion asociada a un asesor.
type AsesorComercialComision struct {
	ID                 int64   `json:"id"`
	AsesorID           int64   `json:"asesor_id"`
	AsesorCodigo       string  `json:"asesor_codigo"`
	AsesorEmail        string  `json:"asesor_email"`
	EmpresaID          int64   `json:"empresa_id"`
	EmpresaNombre      string  `json:"empresa_nombre"`
	LicenciaID         int64   `json:"licencia_id"`
	PagoProvider       string  `json:"pago_provider"`
	PagoID             int64   `json:"pago_id"`
	TransactionID      string  `json:"transaction_id"`
	Referencia         string  `json:"referencia"`
	ValorPagado        float64 `json:"valor_pagado"`
	PorcentajeComision float64 `json:"porcentaje_comision"`
	MontoComision      float64 `json:"monto_comision"`
	FechaPago          string  `json:"fecha_pago"`
	AsociadoDesde      string  `json:"asociado_desde"`
	AsociadoHasta      string  `json:"asociado_hasta"`
	Pagado             int     `json:"pagado"`
	FechaPagoComision  string  `json:"fecha_pago_comision"`
	PagadoPor          string  `json:"pagado_por"`
	Estado             string  `json:"estado"`
	Observaciones      string  `json:"observaciones"`
	FechaCreacion      string  `json:"fecha_creacion"`
	FechaActualizacion string  `json:"fecha_actualizacion"`
}

// EnsureAsesorComercialSchema prepara el modulo de asesores comerciales en pcs_superadministrador.
func EnsureAsesorComercialSchema(dbConn *sql.DB) error {
	if dbConn == nil || !isPostgresDialect() {
		return nil
	}
	asesorComercialSchemaMu.Lock()
	defer asesorComercialSchemaMu.Unlock()

	if asesorComercialSchemaReady {
		return nil
	}
	ready, err := asesorComercialSchemaLooksReady(dbConn)
	if err == nil && ready {
		asesorComercialSchemaReady = true
		return nil
	}
	statements := []string{
		`CREATE TABLE IF NOT EXISTS asesores_comerciales (
			id BIGSERIAL PRIMARY KEY,
			admin_email TEXT NOT NULL,
			admin_nombre TEXT,
			codigo TEXT NOT NULL UNIQUE,
			porcentaje_comision NUMERIC(12,4) DEFAULT 0,
			meses_asociacion INTEGER DEFAULT 6,
			estado_invitacion TEXT DEFAULT 'pendiente',
			invitacion_token_hash TEXT,
			invitacion_expira_en TEXT,
			invitado_por_email TEXT,
			aceptado_en TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP::text,
			fecha_actualizacion TEXT,
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		)`,
		`ALTER TABLE asesores_comerciales ADD COLUMN IF NOT EXISTS admin_nombre TEXT`,
		`ALTER TABLE asesores_comerciales ADD COLUMN IF NOT EXISTS porcentaje_comision NUMERIC(12,4) DEFAULT 0`,
		`ALTER TABLE asesores_comerciales ADD COLUMN IF NOT EXISTS meses_asociacion INTEGER DEFAULT 6`,
		`ALTER TABLE asesores_comerciales ADD COLUMN IF NOT EXISTS estado_invitacion TEXT DEFAULT 'pendiente'`,
		`ALTER TABLE asesores_comerciales ADD COLUMN IF NOT EXISTS invitacion_token_hash TEXT`,
		`ALTER TABLE asesores_comerciales ADD COLUMN IF NOT EXISTS invitacion_expira_en TEXT`,
		`ALTER TABLE asesores_comerciales ADD COLUMN IF NOT EXISTS invitado_por_email TEXT`,
		`ALTER TABLE asesores_comerciales ADD COLUMN IF NOT EXISTS aceptado_en TEXT`,
		`ALTER TABLE asesores_comerciales ADD COLUMN IF NOT EXISTS fecha_actualizacion TEXT`,
		`ALTER TABLE asesores_comerciales ADD COLUMN IF NOT EXISTS usuario_creador TEXT`,
		`ALTER TABLE asesores_comerciales ADD COLUMN IF NOT EXISTS estado TEXT DEFAULT 'activo'`,
		`ALTER TABLE asesores_comerciales ADD COLUMN IF NOT EXISTS observaciones TEXT`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_asesores_comerciales_email ON asesores_comerciales (LOWER(admin_email)) WHERE estado <> 'inactivo'`,
		`CREATE INDEX IF NOT EXISTS ix_asesores_comerciales_codigo ON asesores_comerciales (codigo)`,
		`CREATE INDEX IF NOT EXISTS ix_asesores_comerciales_token ON asesores_comerciales (invitacion_token_hash)`,
		`CREATE TABLE IF NOT EXISTS asesor_comercial_comisiones (
			id BIGSERIAL PRIMARY KEY,
			asesor_id BIGINT,
			asesor_codigo TEXT NOT NULL,
			asesor_email TEXT NOT NULL,
			empresa_id BIGINT NOT NULL,
			empresa_nombre TEXT,
			licencia_id BIGINT,
			pago_provider TEXT,
			pago_id BIGINT,
			transaction_id TEXT,
			referencia TEXT,
			valor_pagado NUMERIC(14,2) DEFAULT 0,
			porcentaje_comision NUMERIC(12,4) DEFAULT 0,
			monto_comision NUMERIC(14,2) DEFAULT 0,
			fecha_pago TEXT,
			asociado_desde TEXT,
			asociado_hasta TEXT,
			pagado INTEGER DEFAULT 0,
			fecha_pago_comision TEXT,
			pagado_por TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP::text,
			fecha_actualizacion TEXT,
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		)`,
		`ALTER TABLE asesor_comercial_comisiones ADD COLUMN IF NOT EXISTS empresa_nombre TEXT`,
		`ALTER TABLE asesor_comercial_comisiones ADD COLUMN IF NOT EXISTS pago_provider TEXT`,
		`ALTER TABLE asesor_comercial_comisiones ADD COLUMN IF NOT EXISTS pago_id BIGINT`,
		`ALTER TABLE asesor_comercial_comisiones ADD COLUMN IF NOT EXISTS transaction_id TEXT`,
		`ALTER TABLE asesor_comercial_comisiones ADD COLUMN IF NOT EXISTS referencia TEXT`,
		`ALTER TABLE asesor_comercial_comisiones ADD COLUMN IF NOT EXISTS valor_pagado NUMERIC(14,2) DEFAULT 0`,
		`ALTER TABLE asesor_comercial_comisiones ADD COLUMN IF NOT EXISTS porcentaje_comision NUMERIC(12,4) DEFAULT 0`,
		`ALTER TABLE asesor_comercial_comisiones ADD COLUMN IF NOT EXISTS monto_comision NUMERIC(14,2) DEFAULT 0`,
		`ALTER TABLE asesor_comercial_comisiones ADD COLUMN IF NOT EXISTS fecha_pago TEXT`,
		`ALTER TABLE asesor_comercial_comisiones ADD COLUMN IF NOT EXISTS asociado_desde TEXT`,
		`ALTER TABLE asesor_comercial_comisiones ADD COLUMN IF NOT EXISTS asociado_hasta TEXT`,
		`ALTER TABLE asesor_comercial_comisiones ADD COLUMN IF NOT EXISTS pagado INTEGER DEFAULT 0`,
		`ALTER TABLE asesor_comercial_comisiones ADD COLUMN IF NOT EXISTS fecha_pago_comision TEXT`,
		`ALTER TABLE asesor_comercial_comisiones ADD COLUMN IF NOT EXISTS pagado_por TEXT`,
		`ALTER TABLE asesor_comercial_comisiones ADD COLUMN IF NOT EXISTS fecha_actualizacion TEXT`,
		`ALTER TABLE asesor_comercial_comisiones ADD COLUMN IF NOT EXISTS usuario_creador TEXT`,
		`ALTER TABLE asesor_comercial_comisiones ADD COLUMN IF NOT EXISTS estado TEXT DEFAULT 'activo'`,
		`ALTER TABLE asesor_comercial_comisiones ADD COLUMN IF NOT EXISTS observaciones TEXT`,
		`CREATE INDEX IF NOT EXISTS ix_asesor_comisiones_codigo ON asesor_comercial_comisiones (asesor_codigo)`,
		`CREATE INDEX IF NOT EXISTS ix_asesor_comisiones_email ON asesor_comercial_comisiones (LOWER(asesor_email))`,
		`CREATE INDEX IF NOT EXISTS ix_asesor_comisiones_empresa ON asesor_comercial_comisiones (empresa_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_asesor_comisiones_ref ON asesor_comercial_comisiones (pago_provider, referencia) WHERE COALESCE(referencia, '') <> ''`,
	}
	for _, stmt := range statements {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}
	asesorComercialSchemaReady = true
	return nil
}

func asesorComercialSchemaLooksReady(dbConn *sql.DB) (bool, error) {
	if dbConn == nil || !isPostgresDialect() {
		return false, nil
	}
	requiredTables := []string{
		"asesores_comerciales",
		"asesor_comercial_comisiones",
	}
	for _, tableName := range requiredTables {
		ok, err := tableExists(dbConn, tableName)
		if err != nil || !ok {
			return false, err
		}
	}

	requiredIndexes := []string{
		"ux_asesores_comerciales_email",
		"ix_asesores_comerciales_codigo",
		"ix_asesores_comerciales_token",
		"ix_asesor_comisiones_empresa",
		"ux_asesor_comisiones_ref",
	}
	for _, indexName := range requiredIndexes {
		ok, err := asesorComercialIndexExists(dbConn, indexName)
		if err != nil || !ok {
			return false, err
		}
	}
	return true, nil
}

func asesorComercialIndexExists(dbConn *sql.DB, indexName string) (bool, error) {
	if dbConn == nil {
		return false, nil
	}
	var exists bool
	err := queryRowSQLCompat(dbConn, `
		SELECT EXISTS (
			SELECT 1
			FROM pg_indexes
			WHERE schemaname = current_schema()
			  AND indexname = ?
		)
	`, strings.TrimSpace(indexName)).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func CreateAsesorComercial(dbConn *sql.DB, item AsesorComercial, tokenHash string) (int64, error) {
	if err := EnsureAsesorComercialSchema(dbConn); err != nil {
		return 0, err
	}
	nowExpr := sqlNowExpr()
	query := `INSERT INTO asesores_comerciales
		(admin_email, admin_nombre, codigo, porcentaje_comision, meses_asociacion, estado_invitacion, invitacion_token_hash, invitacion_expira_en, invitado_por_email, usuario_creador, estado, observaciones, fecha_creacion, fecha_actualizacion)
		VALUES (?, ?, ?, ?, ?, 'pendiente', ?, ?, ?, ?, 'activo', ?, ` + nowExpr + `, ` + nowExpr + `)
		ON CONFLICT (codigo) DO UPDATE SET
			admin_email = EXCLUDED.admin_email,
			admin_nombre = EXCLUDED.admin_nombre,
			porcentaje_comision = EXCLUDED.porcentaje_comision,
			meses_asociacion = EXCLUDED.meses_asociacion,
			estado_invitacion = 'pendiente',
			invitacion_token_hash = EXCLUDED.invitacion_token_hash,
			invitacion_expira_en = EXCLUDED.invitacion_expira_en,
			invitado_por_email = EXCLUDED.invitado_por_email,
			fecha_actualizacion = ` + nowExpr + `,
			estado = 'activo',
			observaciones = EXCLUDED.observaciones
		RETURNING id`
	return insertSQLCompat(dbConn, query,
		strings.TrimSpace(item.AdminEmail),
		strings.TrimSpace(item.AdminNombre),
		strings.ToUpper(strings.TrimSpace(item.Codigo)),
		item.PorcentajeComision,
		item.MesesAsociacion,
		strings.TrimSpace(tokenHash),
		strings.TrimSpace(item.InvitacionExpiraEn),
		strings.TrimSpace(item.InvitadoPorEmail),
		strings.TrimSpace(item.InvitadoPorEmail),
		strings.TrimSpace(item.Observaciones),
	)
}

func ListAsesoresComerciales(dbConn *sql.DB) ([]AsesorComercial, error) {
	if err := EnsureAsesorComercialSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT id, admin_email, COALESCE(admin_nombre,''), codigo, COALESCE(porcentaje_comision,0), COALESCE(meses_asociacion,6), COALESCE(estado_invitacion,''), COALESCE(invitado_por_email,''), COALESCE(invitacion_expira_en,''), COALESCE(aceptado_en,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(estado,'activo'), COALESCE(observaciones,'') FROM asesores_comerciales WHERE COALESCE(estado,'activo') <> 'inactivo' ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []AsesorComercial{}
	for rows.Next() {
		var item AsesorComercial
		if err := rows.Scan(&item.ID, &item.AdminEmail, &item.AdminNombre, &item.Codigo, &item.PorcentajeComision, &item.MesesAsociacion, &item.EstadoInvitacion, &item.InvitadoPorEmail, &item.InvitacionExpiraEn, &item.AceptadoEn, &item.FechaCreacion, &item.FechaActualizacion, &item.Estado, &item.Observaciones); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func GetAsesorComercialByCode(dbConn *sql.DB, code string) (*AsesorComercial, error) {
	if err := EnsureAsesorComercialSchema(dbConn); err != nil {
		return nil, err
	}
	return scanAsesorComercial(queryRowSQLCompat(dbConn, `SELECT id, admin_email, COALESCE(admin_nombre,''), codigo, COALESCE(porcentaje_comision,0), COALESCE(meses_asociacion,6), COALESCE(estado_invitacion,''), COALESCE(invitado_por_email,''), COALESCE(invitacion_expira_en,''), COALESCE(aceptado_en,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(estado,'activo'), COALESCE(observaciones,'') FROM asesores_comerciales WHERE UPPER(codigo) = UPPER(?) AND COALESCE(estado,'activo') <> 'inactivo' LIMIT 1`, strings.TrimSpace(code)))
}

func GetAsesorComercialByEmail(dbConn *sql.DB, email string) (*AsesorComercial, error) {
	if err := EnsureAsesorComercialSchema(dbConn); err != nil {
		return nil, err
	}
	return scanAsesorComercial(queryRowSQLCompat(dbConn, `SELECT id, admin_email, COALESCE(admin_nombre,''), codigo, COALESCE(porcentaje_comision,0), COALESCE(meses_asociacion,6), COALESCE(estado_invitacion,''), COALESCE(invitado_por_email,''), COALESCE(invitacion_expira_en,''), COALESCE(aceptado_en,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(estado,'activo'), COALESCE(observaciones,'') FROM asesores_comerciales WHERE LOWER(admin_email) = LOWER(?) AND COALESCE(estado,'activo') <> 'inactivo' LIMIT 1`, strings.TrimSpace(email)))
}

func GetAsesorComercialByTokenHash(dbConn *sql.DB, tokenHash string) (*AsesorComercial, error) {
	if err := EnsureAsesorComercialSchema(dbConn); err != nil {
		return nil, err
	}
	return scanAsesorComercial(queryRowSQLCompat(dbConn, `SELECT id, admin_email, COALESCE(admin_nombre,''), codigo, COALESCE(porcentaje_comision,0), COALESCE(meses_asociacion,6), COALESCE(estado_invitacion,''), COALESCE(invitado_por_email,''), COALESCE(invitacion_expira_en,''), COALESCE(aceptado_en,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(estado,'activo'), COALESCE(observaciones,'') FROM asesores_comerciales WHERE invitacion_token_hash = ? AND COALESCE(estado,'activo') <> 'inactivo' LIMIT 1`, strings.TrimSpace(tokenHash)))
}

func scanAsesorComercial(row *sql.Row) (*AsesorComercial, error) {
	var item AsesorComercial
	if err := row.Scan(&item.ID, &item.AdminEmail, &item.AdminNombre, &item.Codigo, &item.PorcentajeComision, &item.MesesAsociacion, &item.EstadoInvitacion, &item.InvitadoPorEmail, &item.InvitacionExpiraEn, &item.AceptadoEn, &item.FechaCreacion, &item.FechaActualizacion, &item.Estado, &item.Observaciones); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func AcceptAsesorComercialInvitation(dbConn *sql.DB, id int64, acceptedAt, actor string) error {
	if err := EnsureAsesorComercialSchema(dbConn); err != nil {
		return err
	}
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, `UPDATE asesores_comerciales SET estado_invitacion = 'aceptada', aceptado_en = ?, invitacion_token_hash = '', fecha_actualizacion = `+nowExpr+`, usuario_creador = ? WHERE id = ?`, strings.TrimSpace(acceptedAt), strings.TrimSpace(actor), id)
	return err
}

func UpdateAsesorComercial(dbConn *sql.DB, id int64, pct float64, meses int, observaciones, actor string) error {
	if err := EnsureAsesorComercialSchema(dbConn); err != nil {
		return err
	}
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, `UPDATE asesores_comerciales SET porcentaje_comision = ?, meses_asociacion = ?, observaciones = ?, fecha_actualizacion = `+nowExpr+`, usuario_creador = ? WHERE id = ?`, pct, meses, strings.TrimSpace(observaciones), strings.TrimSpace(actor), id)
	return err
}

func InactivateAsesorComercial(dbConn *sql.DB, id int64, actor string) error {
	if err := EnsureAsesorComercialSchema(dbConn); err != nil {
		return err
	}
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, `UPDATE asesores_comerciales SET estado = 'inactivo', fecha_actualizacion = `+nowExpr+`, usuario_creador = ? WHERE id = ?`, strings.TrimSpace(actor), id)
	return err
}

func CreateAsesorComercialComision(dbConn *sql.DB, item AsesorComercialComision) (int64, error) {
	if err := EnsureAsesorComercialSchema(dbConn); err != nil {
		return 0, err
	}
	if strings.TrimSpace(item.PagoProvider) != "" && strings.TrimSpace(item.Referencia) != "" {
		existing, err := GetAsesorComercialComisionByProviderReference(dbConn, item.PagoProvider, item.Referencia)
		if err != nil {
			return 0, err
		}
		if existing != nil {
			return existing.ID, nil
		}
	}
	nowExpr := sqlNowExpr()
	query := `INSERT INTO asesor_comercial_comisiones
		(asesor_id, asesor_codigo, asesor_email, empresa_id, empresa_nombre, licencia_id, pago_provider, pago_id, transaction_id, referencia, valor_pagado, porcentaje_comision, monto_comision, fecha_pago, asociado_desde, asociado_hasta, pagado, usuario_creador, estado, observaciones, fecha_creacion, fecha_actualizacion)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'activo', ?, ` + nowExpr + `, ` + nowExpr + `)`
	return insertSQLCompat(dbConn, query, item.AsesorID, strings.ToUpper(strings.TrimSpace(item.AsesorCodigo)), strings.TrimSpace(item.AsesorEmail), item.EmpresaID, strings.TrimSpace(item.EmpresaNombre), item.LicenciaID, strings.ToLower(strings.TrimSpace(item.PagoProvider)), item.PagoID, strings.TrimSpace(item.TransactionID), strings.TrimSpace(item.Referencia), item.ValorPagado, item.PorcentajeComision, item.MontoComision, strings.TrimSpace(item.FechaPago), strings.TrimSpace(item.AsociadoDesde), strings.TrimSpace(item.AsociadoHasta), item.Pagado, strings.TrimSpace(item.UsuarioCreador()), strings.TrimSpace(item.Observaciones))
}

func (item AsesorComercialComision) UsuarioCreador() string {
	if strings.TrimSpace(item.PagadoPor) != "" {
		return strings.TrimSpace(item.PagadoPor)
	}
	return strings.TrimSpace(item.AsesorEmail)
}

func GetAsesorComercialComisionByProviderReference(dbConn *sql.DB, provider, reference string) (*AsesorComercialComision, error) {
	if err := EnsureAsesorComercialSchema(dbConn); err != nil {
		return nil, err
	}
	row := queryRowSQLCompat(dbConn, `SELECT id, asesor_id, asesor_codigo, asesor_email, empresa_id, COALESCE(empresa_nombre,''), COALESCE(licencia_id,0), COALESCE(pago_provider,''), COALESCE(pago_id,0), COALESCE(transaction_id,''), COALESCE(referencia,''), COALESCE(valor_pagado,0), COALESCE(porcentaje_comision,0), COALESCE(monto_comision,0), COALESCE(fecha_pago,''), COALESCE(asociado_desde,''), COALESCE(asociado_hasta,''), COALESCE(pagado,0), COALESCE(fecha_pago_comision,''), COALESCE(pagado_por,''), COALESCE(estado,'activo'), COALESCE(observaciones,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,'') FROM asesor_comercial_comisiones WHERE LOWER(COALESCE(pago_provider,'')) = LOWER(?) AND referencia = ? LIMIT 1`, strings.TrimSpace(provider), strings.TrimSpace(reference))
	return scanAsesorComercialComision(row)
}

func GetActiveAsesorComercialAssociationByEmpresa(dbConn *sql.DB, empresaID int64) (*AsesorComercialComision, error) {
	if err := EnsureAsesorComercialSchema(dbConn); err != nil {
		return nil, err
	}
	row := queryRowSQLCompat(dbConn, `SELECT id, asesor_id, asesor_codigo, asesor_email, empresa_id, COALESCE(empresa_nombre,''), COALESCE(licencia_id,0), COALESCE(pago_provider,''), COALESCE(pago_id,0), COALESCE(transaction_id,''), COALESCE(referencia,''), COALESCE(valor_pagado,0), COALESCE(porcentaje_comision,0), COALESCE(monto_comision,0), COALESCE(fecha_pago,''), COALESCE(asociado_desde,''), COALESCE(asociado_hasta,''), COALESCE(pagado,0), COALESCE(fecha_pago_comision,''), COALESCE(pagado_por,''), COALESCE(estado,'activo'), COALESCE(observaciones,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,'') FROM asesor_comercial_comisiones WHERE empresa_id = ? AND COALESCE(estado,'activo') = 'activo' AND (COALESCE(asociado_hasta,'') = '' OR asociado_hasta >= CURRENT_DATE::text) ORDER BY id DESC LIMIT 1`, empresaID)
	return scanAsesorComercialComision(row)
}

func ListAsesorComercialComisiones(dbConn *sql.DB, asesorEmail string, includeExpired bool) ([]AsesorComercialComision, error) {
	if err := EnsureAsesorComercialSchema(dbConn); err != nil {
		return nil, err
	}
	query := `SELECT id, asesor_id, asesor_codigo, asesor_email, empresa_id, COALESCE(empresa_nombre,''), COALESCE(licencia_id,0), COALESCE(pago_provider,''), COALESCE(pago_id,0), COALESCE(transaction_id,''), COALESCE(referencia,''), COALESCE(valor_pagado,0), COALESCE(porcentaje_comision,0), COALESCE(monto_comision,0), COALESCE(fecha_pago,''), COALESCE(asociado_desde,''), COALESCE(asociado_hasta,''), COALESCE(pagado,0), COALESCE(fecha_pago_comision,''), COALESCE(pagado_por,''), COALESCE(estado,'activo'), COALESCE(observaciones,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,'') FROM asesor_comercial_comisiones WHERE COALESCE(estado,'activo') = 'activo'`
	args := []interface{}{}
	if strings.TrimSpace(asesorEmail) != "" {
		query += ` AND LOWER(asesor_email) = LOWER(?)`
		args = append(args, strings.TrimSpace(asesorEmail))
	}
	if !includeExpired {
		query += ` AND (COALESCE(asociado_hasta,'') = '' OR asociado_hasta >= CURRENT_DATE::text)`
	}
	query += ` ORDER BY fecha_pago DESC, id DESC`
	rows, err := querySQLCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []AsesorComercialComision{}
	for rows.Next() {
		item, err := scanAsesorComercialComisionRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

func MarkAsesorComercialComisionPagada(dbConn *sql.DB, id int64, pagadoPor, observaciones string) error {
	if err := EnsureAsesorComercialSchema(dbConn); err != nil {
		return err
	}
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, `UPDATE asesor_comercial_comisiones SET pagado = 1, fecha_pago_comision = `+nowExpr+`, pagado_por = ?, observaciones = CASE WHEN ? <> '' THEN ? ELSE observaciones END, fecha_actualizacion = `+nowExpr+` WHERE id = ?`, strings.TrimSpace(pagadoPor), strings.TrimSpace(observaciones), strings.TrimSpace(observaciones), id)
	return err
}

func scanAsesorComercialComision(row *sql.Row) (*AsesorComercialComision, error) {
	var item AsesorComercialComision
	if err := row.Scan(&item.ID, &item.AsesorID, &item.AsesorCodigo, &item.AsesorEmail, &item.EmpresaID, &item.EmpresaNombre, &item.LicenciaID, &item.PagoProvider, &item.PagoID, &item.TransactionID, &item.Referencia, &item.ValorPagado, &item.PorcentajeComision, &item.MontoComision, &item.FechaPago, &item.AsociadoDesde, &item.AsociadoHasta, &item.Pagado, &item.FechaPagoComision, &item.PagadoPor, &item.Estado, &item.Observaciones, &item.FechaCreacion, &item.FechaActualizacion); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanAsesorComercialComisionRows(row rowScanner) (*AsesorComercialComision, error) {
	var item AsesorComercialComision
	if err := row.Scan(&item.ID, &item.AsesorID, &item.AsesorCodigo, &item.AsesorEmail, &item.EmpresaID, &item.EmpresaNombre, &item.LicenciaID, &item.PagoProvider, &item.PagoID, &item.TransactionID, &item.Referencia, &item.ValorPagado, &item.PorcentajeComision, &item.MontoComision, &item.FechaPago, &item.AsociadoDesde, &item.AsociadoHasta, &item.Pagado, &item.FechaPagoComision, &item.PagadoPor, &item.Estado, &item.Observaciones, &item.FechaCreacion, &item.FechaActualizacion); err != nil {
		return nil, err
	}
	return &item, nil
}
