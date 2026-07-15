package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type EmpresaAgenteUsoDiario struct {
	EmpresaID          int64  `json:"empresa_id"`
	FechaUso           string `json:"fecha_uso"`
	SegundosUsados     int64  `json:"segundos_usados"`
	ConsultasAvanzadas int64  `json:"consultas_avanzadas"`
	ConsultasLigeras   int64  `json:"consultas_ligeras"`
	UsuarioCreador     string `json:"usuario_creador"`
}

func EnsureEmpresaAgentesUsoSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	if dbConn == nil {
		return nil
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_agentes_uso_diario (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			fecha_uso TEXT NOT NULL,
			segundos_usados INTEGER DEFAULT 0,
			consultas_avanzadas INTEGER DEFAULT 0,
			consultas_ligeras INTEGER DEFAULT 0,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, fecha_uso)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_agentes_uso_empresa_fecha ON empresa_agentes_uso_diario(empresa_id, fecha_uso DESC);`,
	}
	for _, stmt := range stmts {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func GetEmpresaAgenteUsoDiario(dbConn *sql.DB, empresaID int64, fechaUso string) (EmpresaAgenteUsoDiario, error) {
	var out EmpresaAgenteUsoDiario
	if empresaID <= 0 {
		return out, fmt.Errorf("empresa_id requerido")
	}
	if err := EnsureEmpresaAgentesUsoSchema(dbConn); err != nil {
		return out, err
	}
	fechaUso = strings.TrimSpace(fechaUso)
	if fechaUso == "" {
		fechaUso = time.Now().Format("2006-01-02")
	}
	out.EmpresaID = empresaID
	out.FechaUso = fechaUso
	err := queryRowSQLCompat(dbConn, `SELECT empresa_id, fecha_uso, COALESCE(segundos_usados,0), COALESCE(consultas_avanzadas,0), COALESCE(consultas_ligeras,0), COALESCE(usuario_creador,'')
		FROM empresa_agentes_uso_diario WHERE empresa_id = ? AND fecha_uso = ? LIMIT 1`, empresaID, fechaUso).
		Scan(&out.EmpresaID, &out.FechaUso, &out.SegundosUsados, &out.ConsultasAvanzadas, &out.ConsultasLigeras, &out.UsuarioCreador)
	if err == sql.ErrNoRows {
		return out, nil
	}
	return out, err
}

func AddEmpresaAgenteUsoDiario(dbConn *sql.DB, in EmpresaAgenteUsoDiario) error {
	if in.EmpresaID <= 0 {
		return fmt.Errorf("empresa_id requerido")
	}
	if err := EnsureEmpresaAgentesUsoSchema(dbConn); err != nil {
		return err
	}
	in.FechaUso = strings.TrimSpace(in.FechaUso)
	if in.FechaUso == "" {
		in.FechaUso = time.Now().Format("2006-01-02")
	}
	if in.SegundosUsados < 0 {
		in.SegundosUsados = 0
	}
	if in.ConsultasAvanzadas < 0 {
		in.ConsultasAvanzadas = 0
	}
	if in.ConsultasLigeras < 0 {
		in.ConsultasLigeras = 0
	}
	_, err := execSQLCompat(dbConn, `INSERT INTO empresa_agentes_uso_diario
		(empresa_id, fecha_uso, segundos_usados, consultas_avanzadas, consultas_ligeras, usuario_creador, fecha_actualizacion)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT (empresa_id, fecha_uso) DO UPDATE SET
			segundos_usados = COALESCE(empresa_agentes_uso_diario.segundos_usados,0) + EXCLUDED.segundos_usados,
			consultas_avanzadas = COALESCE(empresa_agentes_uso_diario.consultas_avanzadas,0) + EXCLUDED.consultas_avanzadas,
			consultas_ligeras = COALESCE(empresa_agentes_uso_diario.consultas_ligeras,0) + EXCLUDED.consultas_ligeras,
			usuario_creador = EXCLUDED.usuario_creador,
			fecha_actualizacion = CURRENT_TIMESTAMP`,
		in.EmpresaID, in.FechaUso, in.SegundosUsados, in.ConsultasAvanzadas, in.ConsultasLigeras, strings.TrimSpace(in.UsuarioCreador))
	return err
}
