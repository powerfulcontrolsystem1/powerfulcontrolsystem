package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// EmpresaAIConsulta representa una consulta/respuesta de IA por empresa.
type EmpresaAIConsulta struct {
	ID               int64  `json:"id"`
	EmpresaID        int64  `json:"empresa_id"`
	Provider         string `json:"provider"`
	ModelID          string `json:"model_id"`
	Pregunta         string `json:"pregunta"`
	Respuesta        string `json:"respuesta"`
	PromptTokens     int64  `json:"prompt_tokens"`
	CompletionTokens int64  `json:"completion_tokens"`
	TotalTokens      int64  `json:"total_tokens"`
	FechaConsulta    string `json:"fecha_consulta"`
	PlanActual       string `json:"plan_actual"`
	FechaCreacion    string `json:"fecha_creacion"`
	FechaActualiz    string `json:"fecha_actualizacion"`
	UsuarioCreador   string `json:"usuario_creador"`
	Estado           string `json:"estado"`
	Observaciones    string `json:"observaciones"`
}

// EmpresaAIUsoDiario representa uso diario por empresa/proveedor/modelo.
type EmpresaAIUsoDiario struct {
	ID            int64  `json:"id"`
	EmpresaID     int64  `json:"empresa_id"`
	Provider      string `json:"provider"`
	ModelID       string `json:"model_id"`
	FechaUso      string `json:"fecha_uso"`
	Consultas     int64  `json:"consultas_total"`
	TokensTotal   int64  `json:"tokens_total"`
	PlanActual    string `json:"plan_actual"`
	FechaCreacion string `json:"fecha_creacion"`
	FechaActualiz string `json:"fecha_actualizacion"`
	UsuarioCread  string `json:"usuario_creador"`
	Estado        string `json:"estado"`
	Observaciones string `json:"observaciones"`
}

// EmpresaAIModeloPreferido vincula el modelo IA preferido por empresa y cuenta Google.
type EmpresaAIModeloPreferido struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	AdminEmail         string `json:"admin_email"`
	Provider           string `json:"provider"`
	ModelID            string `json:"model_id"`
	FechaCreacion      string `json:"fecha_creacion"`
	FechaActualizacion string `json:"fecha_actualizacion"`
	UsuarioCreador     string `json:"usuario_creador"`
	Estado             string `json:"estado"`
	Observaciones      string `json:"observaciones"`
}

// EnsureEmpresaAIChatSchema crea/ajusta el esquema del modulo de chat con IA por empresa.
func EnsureEmpresaAIChatSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_ai_consultas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			provider TEXT NOT NULL,
			model_id TEXT NOT NULL,
			pregunta TEXT NOT NULL,
			respuesta TEXT,
			prompt_tokens INTEGER DEFAULT 0,
			completion_tokens INTEGER DEFAULT 0,
			total_tokens INTEGER DEFAULT 0,
			fecha_consulta TEXT DEFAULT (datetime('now','localtime')),
			plan_actual TEXT DEFAULT 'free',
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS empresa_ai_uso_diario (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			provider TEXT NOT NULL,
			model_id TEXT NOT NULL,
			fecha_uso TEXT NOT NULL,
			consultas_total INTEGER DEFAULT 0,
			tokens_total INTEGER DEFAULT 0,
			plan_actual TEXT DEFAULT 'free',
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, provider, model_id, fecha_uso)
		);`,
		`CREATE TABLE IF NOT EXISTS empresa_ai_modelo_preferido (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			admin_email TEXT NOT NULL,
			provider TEXT,
			model_id TEXT NOT NULL,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, admin_email)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_ai_consultas_empresa_fecha ON empresa_ai_consultas(empresa_id, fecha_consulta DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_ai_consultas_empresa_modelo ON empresa_ai_consultas(empresa_id, provider, model_id);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_ai_uso_diario_empresa_fecha ON empresa_ai_uso_diario(empresa_id, fecha_uso DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_ai_modelo_preferido_empresa_admin ON empresa_ai_modelo_preferido(empresa_id, admin_email, estado);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_ai_consultas", "prompt_tokens", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ai_consultas", "completion_tokens", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ai_consultas", "total_tokens", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ai_consultas", "fecha_consulta", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ai_consultas", "plan_actual", "TEXT DEFAULT 'free'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ai_consultas", "fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ai_consultas", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ai_consultas", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ai_consultas", "observaciones", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ai_uso_diario", "consultas_total", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ai_uso_diario", "tokens_total", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ai_uso_diario", "plan_actual", "TEXT DEFAULT 'free'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ai_uso_diario", "fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ai_uso_diario", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ai_uso_diario", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ai_uso_diario", "observaciones", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ai_modelo_preferido", "provider", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ai_modelo_preferido", "model_id", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ai_modelo_preferido", "fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ai_modelo_preferido", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ai_modelo_preferido", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_ai_modelo_preferido", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

// GetEmpresaAIModeloPreferido obtiene el modelo IA vinculado a la cuenta Google del admin.
func GetEmpresaAIModeloPreferido(dbConn *sql.DB, empresaID int64, adminEmail string) (string, error) {
	if empresaID <= 0 {
		return "", fmt.Errorf("empresa_id es obligatorio")
	}
	adminEmail = aiNormalizeAdminEmail(adminEmail)
	if adminEmail == "" {
		return "", fmt.Errorf("admin_email es obligatorio")
	}

	var modelID string
	err := dbConn.QueryRow(`SELECT COALESCE(model_id, '')
	FROM empresa_ai_modelo_preferido
	WHERE empresa_id = ? AND admin_email = ? AND COALESCE(estado, 'activo') <> 'inactivo'
	LIMIT 1`, empresaID, adminEmail).Scan(&modelID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(modelID), nil
}

// UpsertEmpresaAIModeloPreferido registra/actualiza el modelo preferido por empresa y cuenta Google.
func UpsertEmpresaAIModeloPreferido(dbConn *sql.DB, empresaID int64, adminEmail, modelID, usuarioCreador string) error {
	if empresaID <= 0 {
		return fmt.Errorf("empresa_id es obligatorio")
	}
	adminEmail = aiNormalizeAdminEmail(adminEmail)
	if adminEmail == "" {
		return fmt.Errorf("admin_email es obligatorio")
	}
	modelID = aiNormalizeModelID(modelID)
	if modelID == "" {
		return fmt.Errorf("model_id es obligatorio")
	}
	provider := aiProviderFromModelID(modelID)
	if strings.TrimSpace(usuarioCreador) == "" {
		usuarioCreador = adminEmail
	}

	_, err := dbConn.Exec(`INSERT INTO empresa_ai_modelo_preferido (
		empresa_id,
		admin_email,
		provider,
		model_id,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, 'activo', 'preferencia de modelo IA por cuenta Google', datetime('now','localtime'), datetime('now','localtime'))
	ON CONFLICT(empresa_id, admin_email) DO UPDATE SET
		provider = excluded.provider,
		model_id = excluded.model_id,
		usuario_creador = excluded.usuario_creador,
		estado = 'activo',
		observaciones = excluded.observaciones,
		fecha_actualizacion = datetime('now','localtime')`,
		empresaID,
		adminEmail,
		provider,
		modelID,
		strings.TrimSpace(usuarioCreador),
	)
	return err
}

// CanAdminAccessEmpresaIA valida acceso del admin a empresa_id.
// Super administrador puede acceder a cualquier empresa.
func CanAdminAccessEmpresaIA(dbEmp, dbSuper *sql.DB, adminEmail string, empresaID int64) (bool, error) {
	adminEmail = strings.TrimSpace(strings.ToLower(adminEmail))
	if empresaID <= 0 {
		return false, nil
	}
	if adminEmail == "" {
		return false, nil
	}

	if dbSuper != nil {
		if adm, err := GetAdminByEmail(dbSuper, adminEmail); err == nil {
			if strings.EqualFold(strings.TrimSpace(adm.Role), "super_administrador") {
				return true, nil
			}
		}
	}

	var creador string
	err := dbEmp.QueryRow(`SELECT COALESCE(usuario_creador, '') FROM empresas WHERE id = ? LIMIT 1`, empresaID).Scan(&creador)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	creador = strings.TrimSpace(strings.ToLower(creador))
	if creador == "" {
		return true, nil
	}
	return creador == adminEmail, nil
}

// GetEmpresaAIUsoDiario obtiene el uso diario para un modelo.
func GetEmpresaAIUsoDiario(dbConn *sql.DB, empresaID int64, provider, modelID, fechaUso string) (*EmpresaAIUsoDiario, error) {
	provider = aiNormalizeProvider(provider)
	modelID = aiNormalizeModelID(modelID)
	fechaUso = aiNormalizeFechaUso(fechaUso)
	if fechaUso == "" {
		fechaUso = time.Now().Format("2006-01-02")
	}

	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(provider, ''),
		COALESCE(model_id, ''),
		COALESCE(fecha_uso, ''),
		COALESCE(consultas_total, 0),
		COALESCE(tokens_total, 0),
		COALESCE(plan_actual, 'free'),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_ai_uso_diario
	WHERE empresa_id = ? AND provider = ? AND model_id = ? AND fecha_uso = ?
	LIMIT 1`, empresaID, provider, modelID, fechaUso)

	var out EmpresaAIUsoDiario
	err := row.Scan(
		&out.ID,
		&out.EmpresaID,
		&out.Provider,
		&out.ModelID,
		&out.FechaUso,
		&out.Consultas,
		&out.TokensTotal,
		&out.PlanActual,
		&out.FechaCreacion,
		&out.FechaActualiz,
		&out.UsuarioCread,
		&out.Estado,
		&out.Observaciones,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return &EmpresaAIUsoDiario{
				EmpresaID:  empresaID,
				Provider:   provider,
				ModelID:    modelID,
				FechaUso:   fechaUso,
				PlanActual: "free",
				Estado:     "activo",
			}, nil
		}
		return nil, err
	}
	return &out, nil
}

// RegisterEmpresaAIConsulta inserta la consulta IA y acumula el uso diario.
func RegisterEmpresaAIConsulta(dbConn *sql.DB, in EmpresaAIConsulta) (int64, error) {
	in.EmpresaID = maxInt64(in.EmpresaID, 0)
	if in.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	in.Provider = aiNormalizeProvider(in.Provider)
	if in.Provider == "" {
		return 0, fmt.Errorf("provider es obligatorio")
	}
	in.ModelID = aiNormalizeModelID(in.ModelID)
	if in.ModelID == "" {
		return 0, fmt.Errorf("model_id es obligatorio")
	}
	in.Pregunta = strings.TrimSpace(in.Pregunta)
	if in.Pregunta == "" {
		return 0, fmt.Errorf("pregunta es obligatoria")
	}
	if in.PromptTokens < 0 {
		in.PromptTokens = 0
	}
	if in.CompletionTokens < 0 {
		in.CompletionTokens = 0
	}
	if in.TotalTokens <= 0 {
		in.TotalTokens = in.PromptTokens + in.CompletionTokens
	}
	if in.TotalTokens < 0 {
		in.TotalTokens = 0
	}
	in.PlanActual = aiNormalizePlan(in.PlanActual)
	if in.UsuarioCreador == "" {
		in.UsuarioCreador = "sistema"
	}
	in.Estado = aiNormalizeEstado(in.Estado)
	if in.FechaConsulta == "" {
		in.FechaConsulta = time.Now().Format("2006-01-02 15:04:05")
	}
	fechaUso := aiNormalizeFechaUso(in.FechaConsulta)
	if fechaUso == "" {
		fechaUso = time.Now().Format("2006-01-02")
	}

	tx, err := dbConn.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	res, err := tx.Exec(`INSERT INTO empresa_ai_consultas (
		empresa_id,
		provider,
		model_id,
		pregunta,
		respuesta,
		prompt_tokens,
		completion_tokens,
		total_tokens,
		fecha_consulta,
		plan_actual,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))`,
		in.EmpresaID,
		in.Provider,
		in.ModelID,
		in.Pregunta,
		strings.TrimSpace(in.Respuesta),
		in.PromptTokens,
		in.CompletionTokens,
		in.TotalTokens,
		in.FechaConsulta,
		in.PlanActual,
		in.UsuarioCreador,
		in.Estado,
		strings.TrimSpace(in.Observaciones),
	)
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(`INSERT INTO empresa_ai_uso_diario (
		empresa_id,
		provider,
		model_id,
		fecha_uso,
		consultas_total,
		tokens_total,
		plan_actual,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, 1, ?, ?, ?, 'activo', ?, datetime('now','localtime'), datetime('now','localtime'))
	ON CONFLICT(empresa_id, provider, model_id, fecha_uso) DO UPDATE SET
		consultas_total = empresa_ai_uso_diario.consultas_total + 1,
		tokens_total = empresa_ai_uso_diario.tokens_total + excluded.tokens_total,
		plan_actual = excluded.plan_actual,
		usuario_creador = excluded.usuario_creador,
		observaciones = excluded.observaciones,
		fecha_actualizacion = datetime('now','localtime')`,
		in.EmpresaID,
		in.Provider,
		in.ModelID,
		fechaUso,
		in.TotalTokens,
		in.PlanActual,
		in.UsuarioCreador,
		strings.TrimSpace(in.Observaciones),
	)
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return id, nil
}

// ListEmpresaAIConsultasRecientes devuelve historial de consultas por empresa.
func ListEmpresaAIConsultasRecientes(dbConn *sql.DB, empresaID int64, limit int) ([]EmpresaAIConsulta, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}

	rows, err := dbConn.Query(`SELECT
		id,
		empresa_id,
		COALESCE(provider, ''),
		COALESCE(model_id, ''),
		COALESCE(pregunta, ''),
		COALESCE(respuesta, ''),
		COALESCE(prompt_tokens, 0),
		COALESCE(completion_tokens, 0),
		COALESCE(total_tokens, 0),
		COALESCE(fecha_consulta, ''),
		COALESCE(plan_actual, 'free'),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_ai_consultas
	WHERE empresa_id = ?
	ORDER BY datetime(fecha_consulta) DESC, id DESC
	LIMIT ?`, empresaID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaAIConsulta, 0)
	for rows.Next() {
		var it EmpresaAIConsulta
		if err := rows.Scan(
			&it.ID,
			&it.EmpresaID,
			&it.Provider,
			&it.ModelID,
			&it.Pregunta,
			&it.Respuesta,
			&it.PromptTokens,
			&it.CompletionTokens,
			&it.TotalTokens,
			&it.FechaConsulta,
			&it.PlanActual,
			&it.FechaCreacion,
			&it.FechaActualiz,
			&it.UsuarioCreador,
			&it.Estado,
			&it.Observaciones,
		); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// BuildEmpresaAIContexto resume datos de la empresa para orientar la respuesta IA.
func BuildEmpresaAIContexto(dbConn *sql.DB, empresaID int64) (string, error) {
	if empresaID <= 0 {
		return "", fmt.Errorf("empresa_id es obligatorio")
	}

	var nombre, nit string
	err := dbConn.QueryRow(`SELECT COALESCE(nombre, ''), COALESCE(nit, '') FROM empresas WHERE id = ? LIMIT 1`, empresaID).Scan(&nombre, &nit)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("empresa no encontrada")
		}
		return "", err
	}

	var b strings.Builder
	b.WriteString("CONTEXTO EMPRESA\n")
	b.WriteString(fmt.Sprintf("- empresa_id: %d\n", empresaID))
	b.WriteString(fmt.Sprintf("- nombre: %s\n", strings.TrimSpace(nombre)))
	b.WriteString(fmt.Sprintf("- nit: %s\n", strings.TrimSpace(nit)))

	if ok, err := tableExists(dbConn, "clientes"); err == nil && ok {
		var totalClientes int64
		_ = dbConn.QueryRow(`SELECT COUNT(1) FROM clientes WHERE empresa_id = ? AND COALESCE(estado, 'activo') <> 'inactivo'`, empresaID).Scan(&totalClientes)
		b.WriteString(fmt.Sprintf("- clientes_activos: %d\n", totalClientes))
	}

	if ok, err := tableExists(dbConn, "productos"); err == nil && ok {
		var totalProductos int64
		_ = dbConn.QueryRow(`SELECT COUNT(1) FROM productos WHERE empresa_id = ? AND COALESCE(estado, 'activo') = 'activo'`, empresaID).Scan(&totalProductos)
		b.WriteString(fmt.Sprintf("- productos_activos: %d\n", totalProductos))
	}

	if ok, err := tableExists(dbConn, "carritos_compras"); err == nil && ok {
		var ventasCerradas int64
		var ventasTotal float64
		_ = dbConn.QueryRow(`SELECT COUNT(1), COALESCE(SUM(total), 0) FROM carritos_compras WHERE empresa_id = ? AND LOWER(COALESCE(estado_carrito, '')) = 'cerrado'`, empresaID).Scan(&ventasCerradas, &ventasTotal)
		b.WriteString(fmt.Sprintf("- ventas_cerradas: %d\n", ventasCerradas))
		b.WriteString(fmt.Sprintf("- ventas_total: %.2f\n", ventasTotal))
	}

	if ok, err := tableExists(dbConn, "empresa_finanzas_movimientos"); err == nil && ok {
		var ingresos float64
		var egresos float64
		_ = dbConn.QueryRow(`SELECT COALESCE(SUM(total), 0) FROM empresa_finanzas_movimientos WHERE empresa_id = ? AND LOWER(COALESCE(tipo_movimiento, '')) = 'ingreso' AND LOWER(COALESCE(estado, 'activo')) = 'activo'`, empresaID).Scan(&ingresos)
		_ = dbConn.QueryRow(`SELECT COALESCE(SUM(total), 0) FROM empresa_finanzas_movimientos WHERE empresa_id = ? AND LOWER(COALESCE(tipo_movimiento, '')) = 'egreso' AND LOWER(COALESCE(estado, 'activo')) = 'activo'`, empresaID).Scan(&egresos)
		b.WriteString(fmt.Sprintf("- finanzas_ingresos: %.2f\n", ingresos))
		b.WriteString(fmt.Sprintf("- finanzas_egresos: %.2f\n", egresos))
		b.WriteString(fmt.Sprintf("- finanzas_balance: %.2f\n", ingresos-egresos))
	}

	return b.String(), nil
}

func aiNormalizeProvider(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

func aiNormalizeModelID(v string) string {
	return strings.TrimSpace(v)
}

func aiNormalizeAdminEmail(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

func aiProviderFromModelID(modelID string) string {
	v := aiNormalizeModelID(modelID)
	if v == "" {
		return ""
	}
	if idx := strings.Index(v, ":"); idx > 0 {
		return strings.ToLower(strings.TrimSpace(v[:idx]))
	}
	return ""
}

func aiNormalizePlan(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "pro" || v == "enterprise" {
		return v
	}
	return "free"
}

func aiNormalizeEstado(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "inactivo" || v == "anulado" {
		return v
	}
	return "activo"
}

func aiNormalizeFechaUso(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	if len(v) >= 10 {
		prefix := v[:10]
		if _, err := time.Parse("2006-01-02", prefix); err == nil {
			return prefix
		}
	}
	layouts := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		time.RFC3339,
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, v); err == nil {
			return t.Format("2006-01-02")
		}
	}
	return ""
}

func maxInt64(v, min int64) int64 {
	if v < min {
		return min
	}
	return v
}
