package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	canAdminAccessEmpresaIACacheMu  sync.Mutex
	canAdminAccessEmpresaIACache    = map[string]cachedAdminEmpresaAccessIA{}
	canAdminAccessEmpresaIACacheTTL = 60 * time.Second
)

func InvalidateCanAdminAccessEmpresaIACache(empresaID int64, adminEmail string) {
	adminEmail = strings.TrimSpace(strings.ToLower(adminEmail))
	if empresaID <= 0 || adminEmail == "" {
		return
	}
	cacheKey := fmt.Sprintf("%d|%s", empresaID, adminEmail)
	canAdminAccessEmpresaIACacheMu.Lock()
	delete(canAdminAccessEmpresaIACache, cacheKey)
	canAdminAccessEmpresaIACacheMu.Unlock()
}

func InvalidateCanAdminAccessEmpresaIAAdminCache(adminEmail string) {
	adminEmail = strings.TrimSpace(strings.ToLower(adminEmail))
	if adminEmail == "" {
		return
	}
	suffix := "|" + adminEmail
	canAdminAccessEmpresaIACacheMu.Lock()
	for key := range canAdminAccessEmpresaIACache {
		if strings.HasSuffix(key, suffix) {
			delete(canAdminAccessEmpresaIACache, key)
		}
	}
	canAdminAccessEmpresaIACacheMu.Unlock()
}

type cachedAdminEmpresaAccessIA struct {
	Allowed  bool
	LoadedAt time.Time
}

// GetEmpresaAIUsoDiarioOpenAITokensGlobal retorna el consumo del día (consultas/tokens) agregado
// para todas las empresas en el proveedor indicado (ej: "openai").
func GetEmpresaAIUsoDiarioOpenAITokensGlobal(dbConn *sql.DB, provider string, fechaUso string) (int64, int64, error) {
	if dbConn == nil {
		return 0, 0, nil
	}
	if strings.TrimSpace(fechaUso) == "" {
		fechaUso = time.Now().Format("2006-01-02")
	}
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		provider = "openai"
	}
	if err := EnsureEmpresaAIChatSchema(dbConn); err != nil {
		return 0, 0, err
	}

	var consultas int64
	var tokens int64
	// Agrupar por día, sumar de todas las empresas y modelos.
	if err := queryRowSQLCompat(dbConn, `SELECT
		COALESCE(SUM(COALESCE(consultas_total, 0)), 0) AS consultas,
		COALESCE(SUM(COALESCE(tokens_total, 0)), 0) AS tokens
		FROM empresa_ai_uso_diario
		WHERE LOWER(COALESCE(provider,'')) = ?
		  AND COALESCE(fecha_uso,'') = ?`, provider, fechaUso).Scan(&consultas, &tokens); err != nil {
		return 0, 0, err
	}
	return consultas, tokens, nil
}

// GetSuperAIUsoDiarioOpenAITokensGlobal retorna el consumo del día (consultas/tokens) agregado
// del chat global de super administrador para el proveedor indicado.
// Si adminEmail está vacío, agrega todas las cuentas super.
func GetSuperAIUsoDiarioOpenAITokensGlobal(dbConn *sql.DB, adminEmail string, provider string, fechaUso string) (int64, int64, error) {
	if dbConn == nil {
		return 0, 0, nil
	}
	if strings.TrimSpace(fechaUso) == "" {
		fechaUso = time.Now().Format("2006-01-02")
	}
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		provider = "openai"
	}
	if err := EnsureSuperAIChatSchema(dbConn); err != nil {
		return 0, 0, err
	}

	adminEmail = strings.TrimSpace(strings.ToLower(adminEmail))

	var consultas int64
	var tokens int64
	if adminEmail == "" {
		if err := queryRowSQLCompat(dbConn, `SELECT
			COALESCE(SUM(COALESCE(consultas_total, 0)), 0) AS consultas,
			COALESCE(SUM(COALESCE(tokens_total, 0)), 0) AS tokens
			FROM super_ai_uso_diario
			WHERE LOWER(COALESCE(provider,'')) = ?
			  AND COALESCE(fecha_uso,'') = ?`, provider, fechaUso).Scan(&consultas, &tokens); err != nil {
			return 0, 0, err
		}
		return consultas, tokens, nil
	}

	if err := queryRowSQLCompat(dbConn, `SELECT
		COALESCE(SUM(COALESCE(consultas_total, 0)), 0) AS consultas,
		COALESCE(SUM(COALESCE(tokens_total, 0)), 0) AS tokens
		FROM super_ai_uso_diario
		WHERE LOWER(COALESCE(provider,'')) = ?
		  AND COALESCE(fecha_uso,'') = ?
		  AND LOWER(COALESCE(admin_email,'')) = ?`, provider, fechaUso, adminEmail).Scan(&consultas, &tokens); err != nil {
		return 0, 0, err
	}
	return consultas, tokens, nil
}

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

// SuperAIConsulta representa una consulta/respuesta de IA en alcance global super.
type SuperAIConsulta struct {
	ID               int64  `json:"id"`
	AdminEmail       string `json:"admin_email"`
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

// SuperAIUsoDiario representa uso diario por administrador/proveedor/modelo.
type SuperAIUsoDiario struct {
	ID            int64  `json:"id"`
	AdminEmail    string `json:"admin_email"`
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

// SuperAIModeloPreferido vincula el modelo IA preferido por administrador super.
type SuperAIModeloPreferido struct {
	ID                 int64  `json:"id"`
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

// EnsureSuperAIChatSchema crea/ajusta el esquema del chat IA global para super administrador.
func EnsureSuperAIChatSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS super_ai_consultas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			admin_email TEXT NOT NULL,
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
		`CREATE TABLE IF NOT EXISTS super_ai_uso_diario (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			admin_email TEXT NOT NULL,
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
			UNIQUE(admin_email, provider, model_id, fecha_uso)
		);`,
		`CREATE TABLE IF NOT EXISTS super_ai_modelo_preferido (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			admin_email TEXT NOT NULL,
			provider TEXT,
			model_id TEXT NOT NULL,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(admin_email)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_super_ai_consultas_admin_fecha ON super_ai_consultas(admin_email, fecha_consulta DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_super_ai_consultas_admin_modelo ON super_ai_consultas(admin_email, provider, model_id);`,
		`CREATE INDEX IF NOT EXISTS ix_super_ai_uso_diario_admin_fecha ON super_ai_uso_diario(admin_email, fecha_uso DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_super_ai_modelo_preferido_admin ON super_ai_modelo_preferido(admin_email, estado);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "super_ai_consultas", "prompt_tokens", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_ai_consultas", "completion_tokens", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_ai_consultas", "total_tokens", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_ai_consultas", "fecha_consulta", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_ai_consultas", "plan_actual", "TEXT DEFAULT 'free'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_ai_consultas", "fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_ai_consultas", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_ai_consultas", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_ai_consultas", "observaciones", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_ai_uso_diario", "consultas_total", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_ai_uso_diario", "tokens_total", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_ai_uso_diario", "plan_actual", "TEXT DEFAULT 'free'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_ai_uso_diario", "fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_ai_uso_diario", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_ai_uso_diario", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_ai_uso_diario", "observaciones", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_ai_modelo_preferido", "provider", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_ai_modelo_preferido", "model_id", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_ai_modelo_preferido", "fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_ai_modelo_preferido", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_ai_modelo_preferido", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_ai_modelo_preferido", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

func shouldRepairAIChatSchema(err error) bool {
	return isMissingTableError(err) || isMissingColumnError(err)
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

	scanModeloPreferido := func() (string, error) {
		var modelID string
		err := queryRowSQLCompat(dbConn, `SELECT COALESCE(model_id, '')
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

	modelID, err := scanModeloPreferido()
	if err != nil {
		if !shouldRepairAIChatSchema(err) {
			return "", err
		}
		if schemaErr := EnsureEmpresaAIChatSchema(dbConn); schemaErr != nil {
			return "", err
		}
		return scanModeloPreferido()
	}
	return modelID, nil
}

// GetSuperAIModeloPreferido obtiene el modelo IA preferido por administrador super.
func GetSuperAIModeloPreferido(dbConn *sql.DB, adminEmail string) (string, error) {
	adminEmail = aiNormalizeAdminEmail(adminEmail)
	if adminEmail == "" {
		return "", fmt.Errorf("admin_email es obligatorio")
	}

	scanModeloPreferido := func() (string, error) {
		var modelID string
		err := queryRowSQLCompat(dbConn, `SELECT COALESCE(model_id, '')
	FROM super_ai_modelo_preferido
	WHERE admin_email = ? AND COALESCE(estado, 'activo') <> 'inactivo'
	LIMIT 1`, adminEmail).Scan(&modelID)
		if err != nil {
			if err == sql.ErrNoRows {
				return "", nil
			}
			return "", err
		}
		return strings.TrimSpace(modelID), nil
	}

	modelID, err := scanModeloPreferido()
	if err != nil {
		if !shouldRepairAIChatSchema(err) {
			return "", err
		}
		if schemaErr := EnsureSuperAIChatSchema(dbConn); schemaErr != nil {
			return "", err
		}
		return scanModeloPreferido()
	}
	return modelID, nil
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

	nowExpr := sqlNowExpr()
	query := `INSERT INTO empresa_ai_modelo_preferido (
		empresa_id,
		admin_email,
		provider,
		model_id,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, 'activo', 'preferencia de modelo IA por cuenta Google', ` + nowExpr + `, ` + nowExpr + `)
	ON CONFLICT(empresa_id, admin_email) DO UPDATE SET
		provider = excluded.provider,
		model_id = excluded.model_id,
		usuario_creador = excluded.usuario_creador,
		estado = 'activo',
		observaciones = excluded.observaciones,
		fecha_actualizacion = ` + nowExpr
	_, err := execSQLCompat(dbConn, query,
		empresaID,
		adminEmail,
		provider,
		modelID,
		strings.TrimSpace(usuarioCreador),
	)
	if err != nil {
		if !shouldRepairAIChatSchema(err) {
			return err
		}
		if schemaErr := EnsureEmpresaAIChatSchema(dbConn); schemaErr != nil {
			return err
		}
		_, err = execSQLCompat(dbConn, query,
			empresaID,
			adminEmail,
			provider,
			modelID,
			strings.TrimSpace(usuarioCreador),
		)
	}
	return err
}

// UpsertSuperAIModeloPreferido registra/actualiza el modelo preferido por administrador super.
func UpsertSuperAIModeloPreferido(dbConn *sql.DB, adminEmail, modelID, usuarioCreador string) error {
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

	nowExpr := sqlNowExpr()
	query := `INSERT INTO super_ai_modelo_preferido (
		admin_email,
		provider,
		model_id,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, 'activo', 'preferencia de modelo IA global super', ` + nowExpr + `, ` + nowExpr + `)
	ON CONFLICT(admin_email) DO UPDATE SET
		provider = excluded.provider,
		model_id = excluded.model_id,
		usuario_creador = excluded.usuario_creador,
		estado = 'activo',
		observaciones = excluded.observaciones,
		fecha_actualizacion = ` + nowExpr
	_, err := execSQLCompat(dbConn, query,
		adminEmail,
		provider,
		modelID,
		strings.TrimSpace(usuarioCreador),
	)
	if err != nil {
		if !shouldRepairAIChatSchema(err) {
			return err
		}
		if schemaErr := EnsureSuperAIChatSchema(dbConn); schemaErr != nil {
			return err
		}
		_, err = execSQLCompat(dbConn, query,
			adminEmail,
			provider,
			modelID,
			strings.TrimSpace(usuarioCreador),
		)
	}
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
	cacheKey := fmt.Sprintf("%d|%s", empresaID, adminEmail)
	canAdminAccessEmpresaIACacheMu.Lock()
	if cached, ok := canAdminAccessEmpresaIACache[cacheKey]; ok && time.Since(cached.LoadedAt) < canAdminAccessEmpresaIACacheTTL {
		canAdminAccessEmpresaIACacheMu.Unlock()
		return cached.Allowed, nil
	}
	canAdminAccessEmpresaIACacheMu.Unlock()

	if dbSuper != nil {
		if adm, err := GetAdminByEmail(dbSuper, adminEmail); err == nil {
			creator := strings.TrimSpace(strings.ToLower(adm.UsuarioCreador))
			if strings.EqualFold(strings.TrimSpace(adm.Role), "super_administrador") && (creator == "" || creator == adminEmail) {
				canAdminAccessEmpresaIACacheMu.Lock()
				canAdminAccessEmpresaIACache[cacheKey] = cachedAdminEmpresaAccessIA{Allowed: true, LoadedAt: time.Now()}
				canAdminAccessEmpresaIACacheMu.Unlock()
				return true, nil
			}
		}
	}

	var creador string
	err := dbEmp.QueryRow(`SELECT COALESCE(usuario_creador, '') FROM empresas WHERE id = ? LIMIT 1`, empresaID).Scan(&creador)
	if err != nil {
		if err == sql.ErrNoRows {
			canAdminAccessEmpresaIACacheMu.Lock()
			canAdminAccessEmpresaIACache[cacheKey] = cachedAdminEmpresaAccessIA{Allowed: false, LoadedAt: time.Now()}
			canAdminAccessEmpresaIACacheMu.Unlock()
			return false, nil
		}
		return false, err
	}
	creador = strings.TrimSpace(strings.ToLower(creador))
	if creador != "" && creador == adminEmail {
		canAdminAccessEmpresaIACacheMu.Lock()
		canAdminAccessEmpresaIACache[cacheKey] = cachedAdminEmpresaAccessIA{Allowed: true, LoadedAt: time.Now()}
		canAdminAccessEmpresaIACacheMu.Unlock()
		return true, nil
	}
	if dbSuper != nil {
		principalEmail, err := ResolveAdminPrincipalEmail(dbSuper, adminEmail)
		if err != nil {
			return false, err
		}
		principalEmail = strings.TrimSpace(strings.ToLower(principalEmail))
		if principalEmail != "" && principalEmail != adminEmail && creador != "" && creador == principalEmail {
			canAdminAccessEmpresaIACacheMu.Lock()
			canAdminAccessEmpresaIACache[cacheKey] = cachedAdminEmpresaAccessIA{Allowed: true, LoadedAt: time.Now()}
			canAdminAccessEmpresaIACacheMu.Unlock()
			return true, nil
		}
		if creador != "" {
			delegatedPrincipals, err := ListActiveAdminPrincipalDelegacionPrincipals(dbSuper, adminEmail)
			if err != nil {
				return false, err
			}
			for _, delegatedPrincipal := range delegatedPrincipals {
				if creador == strings.TrimSpace(strings.ToLower(delegatedPrincipal)) {
					canAdminAccessEmpresaIACacheMu.Lock()
					canAdminAccessEmpresaIACache[cacheKey] = cachedAdminEmpresaAccessIA{Allowed: true, LoadedAt: time.Now()}
					canAdminAccessEmpresaIACacheMu.Unlock()
					return true, nil
				}
			}
		}
		access, err := GetActiveAdminEmpresaCompartidaAcceso(dbSuper, empresaID, adminEmail)
		if err != nil {
			return false, err
		}
		if access != nil {
			canAdminAccessEmpresaIACacheMu.Lock()
			canAdminAccessEmpresaIACache[cacheKey] = cachedAdminEmpresaAccessIA{Allowed: true, LoadedAt: time.Now()}
			canAdminAccessEmpresaIACacheMu.Unlock()
			return true, nil
		}
		sharedBy, err := HasActiveAdminEmpresaCompartidaAccesoBySharer(dbSuper, empresaID, adminEmail)
		if err != nil {
			return false, err
		}
		if sharedBy {
			canAdminAccessEmpresaIACacheMu.Lock()
			canAdminAccessEmpresaIACache[cacheKey] = cachedAdminEmpresaAccessIA{Allowed: true, LoadedAt: time.Now()}
			canAdminAccessEmpresaIACacheMu.Unlock()
			return true, nil
		}
	}
	canAdminAccessEmpresaIACacheMu.Lock()
	canAdminAccessEmpresaIACache[cacheKey] = cachedAdminEmpresaAccessIA{Allowed: false, LoadedAt: time.Now()}
	canAdminAccessEmpresaIACacheMu.Unlock()
	return false, nil
}

// GetEmpresaAIUsoDiario obtiene el uso diario para un modelo.
func GetEmpresaAIUsoDiario(dbConn *sql.DB, empresaID int64, provider, modelID, fechaUso string) (*EmpresaAIUsoDiario, error) {
	provider = aiNormalizeProvider(provider)
	modelID = aiNormalizeModelID(modelID)
	fechaUso = aiNormalizeFechaUso(fechaUso)
	if fechaUso == "" {
		fechaUso = time.Now().Format("2006-01-02")
	}

	scanUsoDiario := func() (*EmpresaAIUsoDiario, error) {
		row := queryRowSQLCompat(dbConn, `SELECT
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

	uso, err := scanUsoDiario()
	if err != nil {
		if !shouldRepairAIChatSchema(err) {
			return nil, err
		}
		if schemaErr := EnsureEmpresaAIChatSchema(dbConn); schemaErr != nil {
			return nil, err
		}
		return scanUsoDiario()
	}
	return uso, nil
}

// GetSuperAIUsoDiario obtiene el uso diario para un modelo en el alcance global super.
func GetSuperAIUsoDiario(dbConn *sql.DB, adminEmail, provider, modelID, fechaUso string) (*SuperAIUsoDiario, error) {
	adminEmail = aiNormalizeAdminEmail(adminEmail)
	provider = aiNormalizeProvider(provider)
	modelID = aiNormalizeModelID(modelID)
	fechaUso = strings.TrimSpace(fechaUso)
	if adminEmail == "" {
		return &SuperAIUsoDiario{}, fmt.Errorf("admin_email es obligatorio")
	}
	if provider == "" || modelID == "" || fechaUso == "" {
		return &SuperAIUsoDiario{}, fmt.Errorf("provider, model_id y fecha_uso son obligatorios")
	}

	item := &SuperAIUsoDiario{
		AdminEmail: adminEmail,
		Provider:   provider,
		ModelID:    modelID,
		FechaUso:   fechaUso,
		PlanActual: "free",
	}
	scanUsoDiario := func() (*SuperAIUsoDiario, error) {
		err := queryRowSQLCompat(dbConn, `SELECT
		id,
		COALESCE(admin_email, ''),
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
	FROM super_ai_uso_diario
	WHERE admin_email = ? AND provider = ? AND model_id = ? AND fecha_uso = ?
	LIMIT 1`, adminEmail, provider, modelID, fechaUso).Scan(
			&item.ID,
			&item.AdminEmail,
			&item.Provider,
			&item.ModelID,
			&item.FechaUso,
			&item.Consultas,
			&item.TokensTotal,
			&item.PlanActual,
			&item.FechaCreacion,
			&item.FechaActualiz,
			&item.UsuarioCread,
			&item.Estado,
			&item.Observaciones,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return item, nil
			}
			return &SuperAIUsoDiario{}, err
		}
		if strings.TrimSpace(item.PlanActual) == "" {
			item.PlanActual = "free"
		}
		return item, nil
	}

	uso, err := scanUsoDiario()
	if err != nil {
		if !shouldRepairAIChatSchema(err) {
			return &SuperAIUsoDiario{}, err
		}
		if schemaErr := EnsureSuperAIChatSchema(dbConn); schemaErr != nil {
			return &SuperAIUsoDiario{}, err
		}
		return scanUsoDiario()
	}
	return uso, nil
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

	nowExpr := sqlNowExpr()
	runInsert := func() (int64, error) {
		tx, err := dbConn.Begin()
		if err != nil {
			return 0, err
		}
		defer func() {
			_ = tx.Rollback()
		}()

		consultaID, err := insertTxSQLCompat(tx, `INSERT INTO empresa_ai_consultas (
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
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, `+nowExpr+`, `+nowExpr+`)`,
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

		_, err = execTxSQLCompat(tx, `INSERT INTO empresa_ai_uso_diario (
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
	) VALUES (?, ?, ?, ?, 1, ?, ?, ?, 'activo', ?, `+nowExpr+`, `+nowExpr+`)
	ON CONFLICT(empresa_id, provider, model_id, fecha_uso) DO UPDATE SET
		consultas_total = empresa_ai_uso_diario.consultas_total + 1,
		tokens_total = empresa_ai_uso_diario.tokens_total + excluded.tokens_total,
		plan_actual = excluded.plan_actual,
		usuario_creador = excluded.usuario_creador,
		observaciones = excluded.observaciones,
		fecha_actualizacion = `+nowExpr,
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

		if err := tx.Commit(); err != nil {
			return 0, err
		}
		return consultaID, nil
	}

	id, err := runInsert()
	if err != nil {
		if !shouldRepairAIChatSchema(err) {
			return 0, err
		}
		if schemaErr := EnsureEmpresaAIChatSchema(dbConn); schemaErr != nil {
			return 0, err
		}
		return runInsert()
	}
	return id, nil
}

// RegisterSuperAIConsulta registra una consulta y acumula uso diario global super.
func RegisterSuperAIConsulta(dbConn *sql.DB, input SuperAIConsulta) (int64, error) {
	input.AdminEmail = aiNormalizeAdminEmail(input.AdminEmail)
	input.Provider = aiNormalizeProvider(input.Provider)
	input.ModelID = aiNormalizeModelID(input.ModelID)
	input.Pregunta = strings.TrimSpace(input.Pregunta)
	input.Respuesta = strings.TrimSpace(input.Respuesta)
	if input.AdminEmail == "" {
		return 0, fmt.Errorf("admin_email es obligatorio")
	}
	if input.Provider == "" || input.ModelID == "" {
		return 0, fmt.Errorf("provider y model_id son obligatorios")
	}
	if input.Pregunta == "" {
		return 0, fmt.Errorf("pregunta es obligatoria")
	}
	if strings.TrimSpace(input.PlanActual) == "" {
		input.PlanActual = "free"
	}
	if strings.TrimSpace(input.UsuarioCreador) == "" {
		input.UsuarioCreador = input.AdminEmail
	}
	if strings.TrimSpace(input.Estado) == "" {
		input.Estado = "activo"
	}
	if strings.TrimSpace(input.FechaConsulta) == "" {
		input.FechaConsulta = time.Now().Format("2006-01-02 15:04:05")
	}

	nowExpr := sqlNowExpr()
	runInsert := func() (int64, error) {
		tx, err := dbConn.Begin()
		if err != nil {
			return 0, err
		}
		defer func() {
			_ = tx.Rollback()
		}()

		consultaID, err := insertTxSQLCompat(tx, `INSERT INTO super_ai_consultas (
		admin_email,
		provider,
		model_id,
		pregunta,
		respuesta,
		prompt_tokens,
		completion_tokens,
		total_tokens,
		fecha_consulta,
		plan_actual,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, `+nowExpr+`, `+nowExpr+`, ?, ?, ?)`,
			input.AdminEmail,
			input.Provider,
			input.ModelID,
			input.Pregunta,
			input.Respuesta,
			input.PromptTokens,
			input.CompletionTokens,
			input.TotalTokens,
			input.FechaConsulta,
			input.PlanActual,
			strings.TrimSpace(input.UsuarioCreador),
			strings.TrimSpace(input.Estado),
			strings.TrimSpace(input.Observaciones),
		)
		if err != nil {
			return 0, err
		}

		fechaUso := input.FechaConsulta
		if len(fechaUso) >= 10 {
			fechaUso = fechaUso[:10]
		}
		if fechaUso == "" {
			fechaUso = time.Now().Format("2006-01-02")
		}
		_, err = execTxSQLCompat(tx, `INSERT INTO super_ai_uso_diario (
		admin_email,
		provider,
		model_id,
		fecha_uso,
		consultas_total,
		tokens_total,
		plan_actual,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, 1, ?, ?, `+nowExpr+`, `+nowExpr+`, ?, 'activo', 'uso diario chat IA global super')
	ON CONFLICT(admin_email, provider, model_id, fecha_uso) DO UPDATE SET
		consultas_total = COALESCE(super_ai_uso_diario.consultas_total, 0) + 1,
		tokens_total = COALESCE(super_ai_uso_diario.tokens_total, 0) + excluded.tokens_total,
		plan_actual = excluded.plan_actual,
		fecha_actualizacion = `+nowExpr+`,
		usuario_creador = excluded.usuario_creador,
		estado = 'activo'`,
			input.AdminEmail,
			input.Provider,
			input.ModelID,
			fechaUso,
			input.TotalTokens,
			input.PlanActual,
			strings.TrimSpace(input.UsuarioCreador),
		)
		if err != nil {
			return consultaID, err
		}

		if err := tx.Commit(); err != nil {
			return 0, err
		}
		return consultaID, nil
	}

	id, err := runInsert()
	if err != nil {
		if !shouldRepairAIChatSchema(err) {
			return 0, err
		}
		if schemaErr := EnsureSuperAIChatSchema(dbConn); schemaErr != nil {
			return 0, err
		}
		return runInsert()
	}

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

	queryRecentes := func() ([]EmpresaAIConsulta, error) {
		rows, err := querySQLCompat(dbConn, `SELECT
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

	items, err := queryRecentes()
	if err != nil {
		if !shouldRepairAIChatSchema(err) {
			return nil, err
		}
		if schemaErr := EnsureEmpresaAIChatSchema(dbConn); schemaErr != nil {
			return nil, err
		}
		return queryRecentes()
	}
	return items, nil
}

// ListSuperAIConsultasRecientes devuelve historial reciente por administrador super.
func ListSuperAIConsultasRecientes(dbConn *sql.DB, adminEmail string, limit int) ([]SuperAIConsulta, error) {
	adminEmail = aiNormalizeAdminEmail(adminEmail)
	if adminEmail == "" {
		return nil, fmt.Errorf("admin_email es obligatorio")
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	queryRecentes := func() ([]SuperAIConsulta, error) {
		rows, err := querySQLCompat(dbConn, `SELECT
		id,
		COALESCE(admin_email, ''),
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
	FROM super_ai_consultas
	WHERE admin_email = ? AND COALESCE(estado, 'activo') <> 'inactivo'
	ORDER BY fecha_consulta DESC, id DESC
	LIMIT ?`, adminEmail, limit)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		items := make([]SuperAIConsulta, 0, limit)
		for rows.Next() {
			var item SuperAIConsulta
			if err := rows.Scan(
				&item.ID,
				&item.AdminEmail,
				&item.Provider,
				&item.ModelID,
				&item.Pregunta,
				&item.Respuesta,
				&item.PromptTokens,
				&item.CompletionTokens,
				&item.TotalTokens,
				&item.FechaConsulta,
				&item.PlanActual,
				&item.FechaCreacion,
				&item.FechaActualiz,
				&item.UsuarioCreador,
				&item.Estado,
				&item.Observaciones,
			); err != nil {
				return nil, err
			}
			items = append(items, item)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return items, nil
	}

	items, err := queryRecentes()
	if err != nil {
		if !shouldRepairAIChatSchema(err) {
			return nil, err
		}
		if schemaErr := EnsureSuperAIChatSchema(dbConn); schemaErr != nil {
			return nil, err
		}
		return queryRecentes()
	}
	return items, nil
}

// BuildEmpresaAIContexto resume datos de la empresa para orientar la respuesta IA.
func BuildEmpresaAIContexto(dbConn *sql.DB, empresaID int64) (string, error) {
	if empresaID <= 0 {
		return "", fmt.Errorf("empresa_id es obligatorio")
	}

	var nombre, nit string
	err := queryRowSQLCompat(dbConn, `SELECT COALESCE(nombre, ''), COALESCE(nit, '') FROM empresas WHERE id = ? LIMIT 1`, empresaID).Scan(&nombre, &nit)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("empresa no encontrada")
		}
		return "", err
	}

	availableTables, err := aiAvailableTables(dbConn, []string{
		"clientes",
		"productos",
		"inventario_existencias",
		"carritos_compras",
		"carrito_compra_items",
		"empresa_finanzas_movimientos",
	})
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("CONTEXTO EMPRESA\n")
	b.WriteString(fmt.Sprintf("- empresa_id: %d\n", empresaID))
	b.WriteString(fmt.Sprintf("- nombre: %s\n", strings.TrimSpace(nombre)))
	b.WriteString(fmt.Sprintf("- nit: %s\n", strings.TrimSpace(nit)))
	b.WriteString(fmt.Sprintf("- tablas_contexto_disponibles: %s\n", strings.Join(availableTables, ", ")))

	if slicesContain(availableTables, "clientes") {
		var totalClientes int64
		_ = queryRowSQLCompat(dbConn, `SELECT COUNT(1) FROM clientes WHERE empresa_id = ? AND COALESCE(estado, 'activo') <> 'inactivo'`, empresaID).Scan(&totalClientes)
		b.WriteString(fmt.Sprintf("- clientes_activos: %d\n", totalClientes))
	}

	if slicesContain(availableTables, "productos") {
		var totalProductos int64
		_ = queryRowSQLCompat(dbConn, `SELECT COUNT(1) FROM productos WHERE empresa_id = ? AND COALESCE(estado, 'activo') = 'activo'`, empresaID).Scan(&totalProductos)
		b.WriteString(fmt.Sprintf("- productos_activos: %d\n", totalProductos))
	}

	if slicesContain(availableTables, "carritos_compras") {
		var ventasCerradas int64
		var ventasTotal float64
		_ = queryRowSQLCompat(dbConn, `SELECT COUNT(1), COALESCE(SUM(total), 0) FROM carritos_compras WHERE empresa_id = ? AND LOWER(COALESCE(estado_carrito, '')) IN ('cerrado', 'pagado', 'finalizado')`, empresaID).Scan(&ventasCerradas, &ventasTotal)
		b.WriteString(fmt.Sprintf("- ventas_cerradas: %d\n", ventasCerradas))
		b.WriteString(fmt.Sprintf("- ventas_total: %.2f\n", ventasTotal))
	}

	if slicesContain(availableTables, "empresa_finanzas_movimientos") {
		var ingresos float64
		var egresos float64
		_ = queryRowSQLCompat(dbConn, `SELECT COALESCE(SUM(COALESCE(NULLIF(total_neto, 0), NULLIF(total, 0), monto, 0)), 0) FROM empresa_finanzas_movimientos WHERE empresa_id = ? AND LOWER(COALESCE(tipo_movimiento, '')) = 'ingreso' AND LOWER(COALESCE(estado, 'activo')) = 'activo'`, empresaID).Scan(&ingresos)
		_ = queryRowSQLCompat(dbConn, `SELECT COALESCE(SUM(COALESCE(NULLIF(total_neto, 0), NULLIF(total, 0), monto, 0)), 0) FROM empresa_finanzas_movimientos WHERE empresa_id = ? AND LOWER(COALESCE(tipo_movimiento, '')) = 'egreso' AND LOWER(COALESCE(estado, 'activo')) = 'activo'`, empresaID).Scan(&egresos)
		b.WriteString(fmt.Sprintf("- finanzas_ingresos: %.2f\n", ingresos))
		b.WriteString(fmt.Sprintf("- finanzas_egresos: %.2f\n", egresos))
		b.WriteString(fmt.Sprintf("- finanzas_balance: %.2f\n", ingresos-egresos))
	}

	writeAIContextSection(&b, "TOP_PRODUCTOS_VENTA", empresaAITopProductos(dbConn, empresaID, availableTables, 5))
	writeAIContextSection(&b, "TOP_CLIENTES", empresaAITopClientes(dbConn, empresaID, availableTables, 5))
	writeAIContextSection(&b, "VENTAS_RECIENTES", empresaAIVentasRecientes(dbConn, empresaID, availableTables, 5))
	writeAIContextSection(&b, "ALERTAS_INVENTARIO", empresaAIAlertasInventario(dbConn, empresaID, availableTables, 5))
	writeAIContextSection(&b, "MOVIMIENTOS_FINANCIEROS_RECIENTES", empresaAIFinanzasRecientes(dbConn, empresaID, availableTables, 5))
	writeAIContextSection(&b, "PRECONFIGURACION_GUIA_IA", empresaAIPreconfiguracionGuia(dbConn, empresaID))

	return b.String(), nil
}

// BuildEmpresaAIContextoForQuestion amplía el contexto base con resultados de consultas seguras
// resueltas por intención, sin permitir SQL libre generado por IA.
type EmpresaAIContextoPreguntaOptions struct {
	Modelo            string
	DBQueryEnabled    bool
	DBQueryEnabledSet bool
	DBQueryMaxTables  int
	DBQueryRows       int
	DBQueryMaxChars   int
}

func BuildEmpresaAIContextoForQuestion(dbConn *sql.DB, empresaID int64, pregunta string, usuarioCreador string, paginaContexto string, modelo ...string) (string, error) {
	return BuildEmpresaAIContextoForQuestionWithOptions(dbConn, empresaID, pregunta, usuarioCreador, paginaContexto, EmpresaAIContextoPreguntaOptions{
		Modelo: aiContextModelName(modelo...),
	})
}

func empresaAIPreconfiguracionGuia(dbConn *sql.DB, empresaID int64) []string {
	if dbConn == nil || empresaID <= 0 {
		return nil
	}
	ok, err := tableExists(dbConn, "empresa_estacion_prefs")
	if err != nil || !ok {
		return nil
	}
	var raw string
	err = queryRowSQLCompat(dbConn, `SELECT COALESCE(valor, '')
		FROM empresa_estacion_prefs
		WHERE empresa_id = ?
		  AND estacion_id = 0
		  AND clave = 'preconfiguracion_tipo_empresa_asistente_ia'
		  AND COALESCE(estado, 'activo') <> 'inactivo'
		LIMIT 1`, empresaID).Scan(&raw)
	if err != nil || strings.TrimSpace(raw) == "" {
		return nil
	}
	var payload struct {
		TipoEmpresaNombre string                          `json:"tipo_empresa_nombre"`
		Preconfiguracion  string                          `json:"preconfiguracion"`
		Asistente         TipoEmpresaPreconfigAsistenteIA `json:"asistente_ia"`
		TareasGuia        []TipoEmpresaPreconfigTareaGuia `json:"tareas_guia"`
		UsuariosGuia      []TipoEmpresaPreconfigUsuario   `json:"usuarios_guia"`
		Estaciones        TipoEmpresaPreconfigEstaciones  `json:"estaciones"`
		ProductoSKUs      []string                        `json:"producto_skus"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return []string{"estado=presente", "raw=" + truncateText(raw, 900)}
	}
	lines := []string{
		"estado=presente",
		"tipo_empresa=" + strings.TrimSpace(payload.TipoEmpresaNombre),
		"plantilla=" + strings.TrimSpace(payload.Preconfiguracion),
		fmt.Sprintf("estaciones=%d prefijo=%s caja=%t", payload.Estaciones.Cantidad, strings.TrimSpace(payload.Estaciones.Prefijo), payload.Estaciones.CajaEnabled),
		fmt.Sprintf("productos_guia=%d usuarios_guia=%d tareas_guia=%d", len(payload.ProductoSKUs), len(payload.UsuariosGuia), len(payload.TareasGuia)),
	}
	if payload.Asistente.Enabled {
		lines = append(lines, "asistente_ia=activo rol="+truncateText(payload.Asistente.Rol, 220))
		for idx, instruction := range payload.Asistente.Instrucciones {
			if idx >= 4 {
				break
			}
			lines = append(lines, "instruccion_ia="+truncateText(instruction, 220))
		}
	}
	for idx, task := range payload.TareasGuia {
		if idx >= 6 {
			break
		}
		lines = append(lines, fmt.Sprintf("tarea=%s | %s | %s", truncateText(task.Modulo, 80), truncateText(task.Titulo, 120), truncateText(task.Descripcion, 180)))
	}
	for idx, user := range payload.UsuariosGuia {
		if idx >= 5 {
			break
		}
		lines = append(lines, fmt.Sprintf("usuario_guia=%s | rol=%s | %s", truncateText(user.Nombre, 120), truncateText(user.Rol, 80), truncateText(user.Observaciones, 160)))
	}
	return lines
}

func BuildEmpresaAIContextoForQuestionWithOptions(dbConn *sql.DB, empresaID int64, pregunta string, usuarioCreador string, paginaContexto string, opts EmpresaAIContextoPreguntaOptions) (string, error) {
	base, err := BuildEmpresaAIContexto(dbConn, empresaID)
	if err != nil {
		return "", err
	}
	safeContext, err := buildEmpresaAISafeIntentContext(dbConn, empresaID, pregunta, usuarioCreador)
	if err != nil {
		return "", err
	}
	if paginaContexto = strings.TrimSpace(paginaContexto); paginaContexto != "" {
		if safeContext != "" {
			safeContext += "\n"
		}
		safeContext += "PAGINA_CONTEXTO: " + paginaContexto
	}
	if auditContext := BuildEmpresaAuditoriaAIContextForQuestion(dbConn, empresaID, pregunta, usuarioCreador, aiContextModelName(opts.Modelo), 20, 2*time.Hour); strings.TrimSpace(auditContext) != "" {
		if safeContext != "" {
			safeContext += "\n"
		}
		safeContext += auditContext
	}
	if dbReadContext := buildEmpresaAIFullReadDBContext(dbConn, empresaID, pregunta, opts); strings.TrimSpace(dbReadContext) != "" {
		if safeContext != "" {
			safeContext += "\n"
		}
		safeContext += dbReadContext
	}
	if safeContext == "" {
		return base, nil
	}
	return base + "\n" + safeContext, nil
}

type empresaAIFullReadTable struct {
	Name        string
	Columns     []DBAdminColumn
	SafeColumns []string
	Score       int
	HasID       bool
	RowCount    int64
}

func buildEmpresaAIFullReadDBContext(dbConn *sql.DB, empresaID int64, pregunta string, opts EmpresaAIContextoPreguntaOptions) string {
	if dbConn == nil || empresaID <= 0 {
		return ""
	}
	if opts.DBQueryEnabledSet && !opts.DBQueryEnabled {
		return ""
	}
	maxTables := opts.DBQueryMaxTables
	if maxTables <= 0 {
		maxTables = 25
	}
	if maxTables > 100 {
		maxTables = 100
	}
	rowsPerTable := opts.DBQueryRows
	if rowsPerTable <= 0 {
		rowsPerTable = 8
	}
	if rowsPerTable > 30 {
		rowsPerTable = 30
	}
	maxChars := opts.DBQueryMaxChars
	if maxChars <= 0 {
		maxChars = 45000
	}
	if maxChars > 120000 {
		maxChars = 120000
	}

	tableNames, err := DBAdminListEmpresaTables(dbConn)
	if err != nil || len(tableNames) == 0 {
		return "BASE_DATOS_EMPRESA_LECTURA_TOTAL\n- estado=no_disponible\n- nota=no se pudo listar tablas consultables por empresa_id.\n"
	}
	terms := aiFullReadTerms(pregunta)
	tables := make([]empresaAIFullReadTable, 0, len(tableNames))
	for _, name := range tableNames {
		if !isSafePostgresUnquotedIdent(name) {
			continue
		}
		cols, err := DBAdminGetTableColumns(dbConn, name)
		if err != nil || len(cols) == 0 {
			continue
		}
		item := empresaAIFullReadTable{Name: name, Columns: cols}
		for _, col := range cols {
			colName := strings.TrimSpace(col.Name)
			if strings.EqualFold(colName, "id") {
				item.HasID = true
			}
			if strings.EqualFold(colName, "empresa_id") || aiFullReadSensitiveColumn(colName) {
				continue
			}
			if isSafePostgresUnquotedIdent(colName) {
				item.SafeColumns = append(item.SafeColumns, colName)
			}
		}
		item.Score = aiFullReadTableScore(item, terms)
		tables = append(tables, item)
	}
	if len(tables) == 0 {
		return ""
	}
	sort.SliceStable(tables, func(i, j int) bool {
		if tables[i].Score != tables[j].Score {
			return tables[i].Score > tables[j].Score
		}
		return tables[i].Name < tables[j].Name
	})
	if len(tables) > maxTables {
		tables = tables[:maxTables]
	}

	var b strings.Builder
	b.WriteString("BASE_DATOS_EMPRESA_LECTURA_TOTAL\n")
	b.WriteString(fmt.Sprintf("- empresa_id=%d\n", empresaID))
	b.WriteString(fmt.Sprintf("- modelo=%s\n", safeAIValue(aiContextModelName(opts.Modelo))))
	b.WriteString("- estado=activo_por_configuracion_super\n")
	b.WriteString("- regla=lectura total controlada: el backend puede consultar cualquier tabla con empresa_id; solo SELECT parametrizado; sin SQL libre del modelo; columnas sensibles omitidas.\n")
	b.WriteString(fmt.Sprintf("- tablas_consultables_detectadas=%d\n", len(tableNames)))
	b.WriteString(fmt.Sprintf("- tablas_entregadas_en_contexto=%d\n", len(tables)))
	b.WriteString(fmt.Sprintf("- filas_por_tabla=%d\n", rowsPerTable))
	b.WriteString("ESQUEMA_TABLAS_EMPRESA\n")
	for _, t := range tables {
		cols := make([]string, 0, len(t.Columns))
		for _, c := range t.Columns {
			name := strings.TrimSpace(c.Name)
			if name == "" || aiFullReadSensitiveColumn(name) {
				continue
			}
			cols = append(cols, fmt.Sprintf("%s:%s", safeAIValue(name), safeAIValue(c.Type)))
			if len(cols) >= 28 {
				cols = append(cols, "...")
				break
			}
		}
		b.WriteString(fmt.Sprintf("- tabla=%s columnas=%s\n", safeAIValue(t.Name), strings.Join(cols, ",")))
	}
	b.WriteString("CONSULTAS_DB_LECTURA_TOTAL_RESUELTAS\n")
	for _, t := range tables {
		lines, count := empresaAIFullReadTableRows(dbConn, empresaID, t, rowsPerTable)
		b.WriteString(fmt.Sprintf("- tabla=%s filas_empresa=%d filas_entregadas=%d score=%d\n", safeAIValue(t.Name), count, len(lines), t.Score))
		for _, line := range lines {
			b.WriteString("  - " + line + "\n")
			if b.Len() >= maxChars {
				b.WriteString("- nota=CONTEXTO_DB_LECTURA_TOTAL_RECORTADO\n")
				return b.String()
			}
		}
		if b.Len() >= maxChars {
			b.WriteString("- nota=CONTEXTO_DB_LECTURA_TOTAL_RECORTADO\n")
			return b.String()
		}
	}
	return b.String()
}

func aiFullReadTerms(pregunta string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, 12)
	add := func(v string) {
		v = strings.Trim(strings.TrimSpace(aiFoldText(v)), ".,;:()[]{}¿?¡!\"'")
		if len([]rune(v)) < 3 {
			return
		}
		if _, ok := seen[v]; ok {
			return
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	for _, term := range aiExtractSearchTerms(pregunta) {
		add(term)
	}
	stop := map[string]bool{"que": true, "quien": true, "como": true, "para": true, "por": true, "con": true, "los": true, "las": true, "una": true, "uno": true, "del": true, "base": true, "datos": true, "tabla": true, "tablas": true, "consulta": true, "consultar": true, "empresa": true, "sistema": true}
	for _, part := range strings.Fields(aiFoldText(pregunta)) {
		part = strings.Trim(part, ".,;:()[]{}¿?¡!\"'")
		if stop[part] {
			continue
		}
		add(part)
		if len(out) >= 12 {
			break
		}
	}
	return out
}

func aiFullReadSensitiveColumn(name string) bool {
	v := aiFoldText(name)
	for _, token := range []string{"password", "passwd", "contrasena", "contraseña", "token", "secret", "secreto", "hash", "salt", "api_key", "apikey", "private_key", "public_key", "jwt", "credential", "credencial", "firma", "certificate", "certificado", "pin", "otp"} {
		if strings.Contains(v, token) {
			return true
		}
	}
	return false
}

func aiFullReadTableScore(t empresaAIFullReadTable, terms []string) int {
	score := 0
	name := aiFoldText(t.Name)
	if len(terms) == 0 {
		score = 1
	}
	for _, term := range terms {
		if term == "" {
			continue
		}
		if strings.Contains(name, term) {
			score += 8
		}
		for _, col := range t.Columns {
			if strings.Contains(aiFoldText(col.Name), term) {
				score += 3
			}
		}
	}
	for _, preferred := range []string{"clientes", "productos", "carritos", "ventas", "finanzas", "inventario", "users", "usuarios", "auditoria", "compras", "facturacion"} {
		if strings.Contains(name, preferred) {
			score++
		}
	}
	return score
}

func empresaAIFullReadTableRows(dbConn *sql.DB, empresaID int64, t empresaAIFullReadTable, limit int) ([]string, int64) {
	var count int64
	countQuery := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE empresa_id = ?", quotePGIdent(t.Name))
	_ = queryRowSQLCompat(dbConn, countQuery, empresaID).Scan(&count)
	if len(t.SafeColumns) == 0 || count == 0 {
		return nil, count
	}
	cols := append([]string{}, t.SafeColumns...)
	if len(cols) > 12 {
		cols = cols[:12]
	}
	quotedCols := make([]string, 0, len(cols))
	for _, col := range cols {
		quotedCols = append(quotedCols, quotePGIdent(col))
	}
	order := ""
	if t.HasID {
		order = " ORDER BY " + quotePGIdent("id") + " DESC"
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE empresa_id = ?%s LIMIT ?", strings.Join(quotedCols, ", "), quotePGIdent(t.Name), order)
	rows, err := querySQLCompat(dbConn, query, empresaID, limit)
	if err != nil {
		return []string{"estado=error_lectura detalle=" + safeAIValue(err.Error())}, count
	}
	defer rows.Close()
	names, err := rows.Columns()
	if err != nil {
		return nil, count
	}
	out := make([]string, 0, limit)
	for rows.Next() {
		raw := make([]interface{}, len(names))
		ptrs := make([]interface{}, len(names))
		for i := range raw {
			ptrs[i] = &raw[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			continue
		}
		parts := make([]string, 0, len(names))
		for i, name := range names {
			parts = append(parts, fmt.Sprintf("%s=%s", safeAIValue(name), aiFullReadFormatValue(raw[i])))
		}
		out = append(out, strings.Join(parts, " | "))
	}
	return out, count
}

func aiFullReadFormatValue(v interface{}) string {
	if v == nil {
		return "null"
	}
	switch val := v.(type) {
	case []byte:
		return aiFullReadCompactValue(string(val))
	case string:
		return aiFullReadCompactValue(val)
	default:
		return aiFullReadCompactValue(fmt.Sprintf("%v", val))
	}
}

func aiFullReadCompactValue(v string) string {
	v = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(v, "\n", " "), "\r", " "))
	for strings.Contains(v, "  ") {
		v = strings.ReplaceAll(v, "  ", " ")
	}
	if v == "" {
		return "sin_dato"
	}
	r := []rune(v)
	if len(r) > 140 {
		v = string(r[:140]) + "..."
	}
	return v
}

// BuildSuperAIContexto resume contexto global consolidado del sistema para super administrador.
func BuildSuperAIContexto(dbEmp, dbSuper *sql.DB, adminEmail string) (string, error) {
	adminEmail = aiNormalizeAdminEmail(adminEmail)
	if adminEmail == "" {
		return "", fmt.Errorf("admin_email es obligatorio")
	}

	availableTables, err := aiAvailableTables(dbEmp, []string{
		"empresas",
		"clientes",
		"productos",
		"inventario_existencias",
		"carritos_compras",
		"empresa_finanzas_movimientos",
	})
	if err != nil {
		return "", err
	}

	var totalEmpresas, empresasActivas int64
	if err := queryRowSQLCompat(dbEmp, `SELECT COUNT(1), COALESCE(SUM(CASE WHEN LOWER(COALESCE(estado,'activo')) = 'activo' THEN 1 ELSE 0 END), 0) FROM empresas`).Scan(&totalEmpresas, &empresasActivas); err != nil {
		return "", err
	}

	var totalClientes int64
	if err := queryRowSQLCompat(dbEmp, `SELECT COUNT(1) FROM clientes`).Scan(&totalClientes); err != nil {
		if !isMissingTableError(err) {
			return "", err
		}
	}

	var totalProductos int64
	if err := queryRowSQLCompat(dbEmp, `SELECT COUNT(1) FROM productos`).Scan(&totalProductos); err != nil {
		if !isMissingTableError(err) {
			return "", err
		}
	}

	var totalCarritos, carritosPagados int64
	var ventasTotal float64
	if err := queryRowSQLCompat(dbEmp, `SELECT COUNT(1), COALESCE(SUM(CASE WHEN LOWER(COALESCE(estado_carrito,'')) IN ('pagado','cerrado','finalizado') THEN 1 ELSE 0 END), 0), COALESCE(SUM(total), 0) FROM carritos_compras`).Scan(&totalCarritos, &carritosPagados, &ventasTotal); err != nil {
		if !isMissingTableError(err) {
			return "", err
		}
	}

	var movimientosFinancieros int64
	var ingresosTotal, egresosTotal float64
	if err := queryRowSQLCompat(dbEmp, `SELECT COUNT(1),
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(tipo_movimiento,'')) = 'ingreso' THEN COALESCE(NULLIF(total_neto, 0), NULLIF(total, 0), monto, 0) ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(tipo_movimiento,'')) = 'egreso' THEN COALESCE(NULLIF(total_neto, 0), NULLIF(total, 0), monto, 0) ELSE 0 END), 0)
		FROM empresa_finanzas_movimientos`).Scan(&movimientosFinancieros, &ingresosTotal, &egresosTotal); err != nil {
		if !isMissingTableError(err) {
			return "", err
		}
	}

	var totalAdministradores int64
	if dbSuper != nil {
		if err := queryRowSQLCompat(dbSuper, `SELECT COUNT(1) FROM administradores`).Scan(&totalAdministradores); err != nil {
			if !isMissingTableError(err) {
				return "", err
			}
		}
	}

	contexto := []string{
		fmt.Sprintf("admin_super_consultando=%s", adminEmail),
		fmt.Sprintf("empresas_total=%d", totalEmpresas),
		fmt.Sprintf("empresas_activas=%d", empresasActivas),
		fmt.Sprintf("clientes_total=%d", totalClientes),
		fmt.Sprintf("productos_total=%d", totalProductos),
		fmt.Sprintf("carritos_total=%d", totalCarritos),
		fmt.Sprintf("carritos_pagados=%d", carritosPagados),
		fmt.Sprintf("ventas_total=%.2f", ventasTotal),
		fmt.Sprintf("movimientos_financieros_total=%d", movimientosFinancieros),
		fmt.Sprintf("ingresos_total=%.2f", ingresosTotal),
		fmt.Sprintf("egresos_total=%.2f", egresosTotal),
		fmt.Sprintf("administradores_total=%d", totalAdministradores),
		fmt.Sprintf("tablas_contexto_disponibles=%s", strings.Join(availableTables, ",")),
		"alcance=global_superadministrador",
		"restriccion=no revelar secretos, tokens, hashes, credenciales ni datos sensibles no solicitados expresamente",
	}

	contexto = append(contexto, aiContextSectionLines("empresas_top_ventas", superAITopEmpresasVentas(dbEmp, availableTables, 5))...)
	contexto = append(contexto, aiContextSectionLines("ventas_globales_recientes", superAIVentasRecientes(dbEmp, availableTables, 5))...)
	contexto = append(contexto, aiContextSectionLines("alertas_globales_inventario", superAIAlertasInventario(dbEmp, availableTables, 5))...)

	return strings.Join(contexto, "\n"), nil
}

// SuperAIContextoOpts modula el contexto inyectado al chat global de superadministrador.
// El bloque de metadatos de la base super (conteos/columnas) se adjunta siempre que exista conexión a dbSuper;
// EmpresaSoloLectura controla si los fragmentos de la base empresas son estrictamente lectura informativa.
type SuperAIContextoOpts struct {
	EmpresaSoloLectura bool
}

// BuildSuperAIContextoForQuestion amplía el contexto global con consultas seguras resueltas
// a partir de la intención de la pregunta.
func BuildSuperAIContextoForQuestion(dbEmp, dbSuper *sql.DB, adminEmail, pregunta string, opts SuperAIContextoOpts, modelo ...string) (string, error) {
	base, err := BuildSuperAIContexto(dbEmp, dbSuper, adminEmail)
	if err != nil {
		return "", err
	}
	if dbSuper != nil {
		if ampliado := buildSuperAIContextoAmpliadoSuper(dbSuper); ampliado != "" {
			base = base + "\n" + ampliado
		}
	}
	safeContext, err := buildSuperAISafeIntentContext(dbEmp, pregunta, opts.EmpresaSoloLectura)
	if err != nil {
		return "", err
	}
	if auditContext := BuildSuperAuditoriaAIContextForQuestion(dbEmp, pregunta, adminEmail, aiContextModelName(modelo...), 25, 2*time.Hour); strings.TrimSpace(auditContext) != "" {
		if safeContext != "" {
			safeContext += "\n"
		}
		safeContext += auditContext
	}
	if safeContext == "" {
		return base, nil
	}
	return base + "\n" + safeContext, nil
}

var aiPGUnquotedIdent = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func isSafePostgresUnquotedIdent(s string) bool {
	if s == "" || len(s) > 63 {
		return false
	}
	return aiPGUnquotedIdent.MatchString(s)
}

func quotePGIdent(s string) string {
	// Incluye reservas y mayúsculas: siempre entre comillas con escape estándar.
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

type superQTable struct {
	Schema string
	Table  string
}

// superAIListBaseTablesInSearchPath devuelve tablas base en los esquemas de búsqueda actuales.
func superAIListBaseTablesInSearchPath(dbConn *sql.DB) ([]superQTable, error) {
	if dbConn == nil {
		return nil, nil
	}
	rows, err := querySQLCompat(dbConn, `
		SELECT table_schema, table_name
		FROM information_schema.tables
		WHERE table_schema = ANY (current_schemas(false))
		  AND table_type = 'BASE TABLE'
		ORDER BY table_schema, table_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []superQTable
	for rows.Next() {
		var sch, tbl string
		if err := rows.Scan(&sch, &tbl); err != nil {
			return nil, err
		}
		sch = strings.TrimSpace(sch)
		tbl = strings.TrimSpace(tbl)
		if !isSafePostgresUnquotedIdent(sch) || !isSafePostgresUnquotedIdent(tbl) {
			continue
		}
		out = append(out, superQTable{Schema: sch, Table: tbl})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// superAIColumnLines lista columnas (nombre:tipo) para una tabla calificada.
func superAIColumnLines(dbConn *sql.DB, q superQTable) ([]string, error) {
	rows, err := querySQLCompat(dbConn, `
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_schema = ?
		  AND table_name = ?
		ORDER BY ordinal_position
	`, q.Schema, q.Table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var parts []string
	for rows.Next() {
		var col, typ string
		if err := rows.Scan(&col, &typ); err != nil {
			return nil, err
		}
		col = strings.TrimSpace(col)
		typ = strings.TrimSpace(typ)
		if col == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s:%s", col, typ))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return parts, nil
}

// buildSuperAIContextoAmpliadoSuper agrega metadatos de toda la base superadministrador en el
// search_path: conteo de filas por tabla, columnas (nombre:tipo) y conteos de administradores por rol.
// No inyecta valores de fila ni contenidos de configuración/secretos.
func buildSuperAIContextoAmpliadoSuper(dbSuper *sql.DB) string {
	if dbSuper == nil {
		return ""
	}
	tables, err := superAIListBaseTablesInSearchPath(dbSuper)
	if err != nil || len(tables) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("CONTEXTO_SUPER_ESQUEMA_COMPLETO\n")
	b.WriteString("nota=metadatos de lectura: filas por tabla, columnas (nombre:tipo) y reparto de roles; no se incluyen filas, tokens, contraseñas ni valores de configuración sensibles.\n")
	hasData := false
	for _, q := range tables {
		cq := fmt.Sprintf(`SELECT COUNT(1) FROM %s.%s`, quotePGIdent(q.Schema), quotePGIdent(q.Table))
		var n int64
		if err := queryRowSQLCompat(dbSuper, cq).Scan(&n); err != nil {
			continue
		}
		b.WriteString(fmt.Sprintf("- %s.%s filas=%d\n", q.Schema, q.Table, n))
		hasData = true
	}
	b.WriteString("CONTEXTO_SUPER_COLUMNAS_POR_TABLA\n")
	for _, q := range tables {
		lines, err := superAIColumnLines(dbSuper, q)
		if err != nil || len(lines) == 0 {
			continue
		}
		joined := strings.Join(lines, ", ")
		if len([]rune(joined)) > 4000 {
			joined = string([]rune(joined)[:4000]) + "..."
		}
		b.WriteString(fmt.Sprintf("- %s.%s: %s\n", q.Schema, q.Table, joined))
		hasData = true
	}
	// Resumen de administradores por rol (sin PII: solo conteos)
	if ok, _ := tableExists(dbSuper, "administradores"); ok {
		rows, err := querySQLCompat(dbSuper, `SELECT LOWER(COALESCE(role,'')) AS r, COUNT(1) FROM administradores GROUP BY LOWER(COALESCE(role,''))`)
		if err == nil {
			defer rows.Close()
			b.WriteString("CONTEXTO_SUPER_ADMIN_ROLES\n")
			for rows.Next() {
				var role string
				var c int64
				if err := rows.Scan(&role, &c); err != nil {
					break
				}
				if strings.TrimSpace(role) == "" {
					role = "sin_rol"
				}
				b.WriteString(fmt.Sprintf("- administradores_por_rol %s=%d\n", safeAIValue(role), c))
				hasData = true
			}
		}
	}
	s := b.String()
	if len([]rune(s)) > 100000 {
		r := []rune(s)
		s = string(r[:100000]) + "\n- nota=CONTEXTO_SUPER_RECORTADO (límite de caracteres; prioriza las tablas alfabéticamente iniciales).\n"
	}
	if !hasData {
		return ""
	}
	return s
}

func aiAvailableTables(dbConn *sql.DB, candidates []string) ([]string, error) {
	available := make([]string, 0, len(candidates))
	for _, tableName := range candidates {
		ok, err := tableExists(dbConn, tableName)
		if err != nil {
			return nil, err
		}
		if ok {
			available = append(available, tableName)
		}
	}
	sort.Strings(available)
	return available, nil
}

func aiAvailableColumns(dbConn *sql.DB, tableName string, candidates []string) ([]string, error) {
	if len(candidates) == 0 {
		return nil, nil
	}

	available := make([]string, 0, len(candidates))
	found := map[string]struct{}{}

	rows, err := querySQLCompat(dbConn, `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = ANY (current_schemas(false))
		  AND table_name = ?
	`, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return nil, err
		}
		found[strings.ToLower(strings.TrimSpace(columnName))] = struct{}{}
	}

	for _, candidate := range candidates {
		if _, ok := found[strings.ToLower(candidate)]; ok {
			available = append(available, candidate)
		}
	}
	return available, nil
}

func slicesContain(values []string, expected string) bool {
	for _, item := range values {
		if item == expected {
			return true
		}
	}
	return false
}

func writeAIContextSection(b *strings.Builder, sectionName string, lines []string) {
	if len(lines) == 0 {
		return
	}
	b.WriteString(sectionName)
	b.WriteString("\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		b.WriteString("- ")
		b.WriteString(line)
		b.WriteString("\n")
	}
}

func aiContextSectionLines(sectionName string, lines []string) []string {
	if len(lines) == 0 {
		return nil
	}
	out := make([]string, 0, len(lines)+1)
	out = append(out, sectionName)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out = append(out, "  - "+line)
	}
	return out
}

func empresaAITopProductos(dbConn *sql.DB, empresaID int64, availableTables []string, limit int) []string {
	if !slicesContain(availableTables, "carritos_compras") || !slicesContain(availableTables, "carrito_compra_items") {
		return nil
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		COALESCE(NULLIF(ci.descripcion, ''), NULLIF(ci.codigo_item, ''), 'item_sin_nombre') AS item_nombre,
		COALESCE(SUM(ci.cantidad), 0) AS cantidad_total,
		COALESCE(SUM(ci.total_linea), 0) AS total_vendido
	FROM carrito_compra_items ci
	JOIN carritos_compras c ON c.id = ci.carrito_id AND c.empresa_id = ci.empresa_id
	WHERE ci.empresa_id = ?
	  AND LOWER(COALESCE(c.estado_carrito, '')) IN ('cerrado', 'pagado', 'finalizado')
	GROUP BY COALESCE(NULLIF(ci.descripcion, ''), NULLIF(ci.codigo_item, ''), 'item_sin_nombre')
	ORDER BY cantidad_total DESC, total_vendido DESC
	LIMIT ?`, empresaID, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()

	out := make([]string, 0, limit)
	for rows.Next() {
		var nombre string
		var cantidad, total float64
		if err := rows.Scan(&nombre, &cantidad, &total); err != nil {
			return nil
		}
		out = append(out, fmt.Sprintf("%s | cantidad=%.2f | total=%.2f", strings.TrimSpace(nombre), cantidad, total))
	}
	return out
}

func empresaAITopClientes(dbConn *sql.DB, empresaID int64, availableTables []string, limit int) []string {
	if !slicesContain(availableTables, "clientes") || !slicesContain(availableTables, "carritos_compras") {
		return nil
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		COALESCE(NULLIF(cl.nombre_razon_social, ''), NULLIF(cl.numero_documento, ''), 'cliente_sin_nombre') AS cliente_nombre,
		COUNT(1) AS compras,
		COALESCE(SUM(c.total), 0) AS total_comprado
	FROM carritos_compras c
	JOIN clientes cl ON cl.id = c.cliente_id AND cl.empresa_id = c.empresa_id
	WHERE c.empresa_id = ?
	  AND LOWER(COALESCE(c.estado_carrito, '')) IN ('cerrado', 'pagado', 'finalizado')
	GROUP BY COALESCE(NULLIF(cl.nombre_razon_social, ''), NULLIF(cl.numero_documento, ''), 'cliente_sin_nombre')
	ORDER BY total_comprado DESC, compras DESC
	LIMIT ?`, empresaID, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()

	out := make([]string, 0, limit)
	for rows.Next() {
		var nombre string
		var compras int64
		var total float64
		if err := rows.Scan(&nombre, &compras, &total); err != nil {
			return nil
		}
		out = append(out, fmt.Sprintf("%s | compras=%d | total=%.2f", strings.TrimSpace(nombre), compras, total))
	}
	return out
}

func empresaAIVentasRecientes(dbConn *sql.DB, empresaID int64, availableTables []string, limit int) []string {
	if !slicesContain(availableTables, "carritos_compras") {
		return nil
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		COALESCE(NULLIF(c.codigo, ''), printf('venta_%d', c.id)),
		COALESCE(c.total, 0),
		COALESCE(c.metodo_pago, ''),
		COALESCE(c.pagado_en, c.fecha_actualizacion, c.fecha_creacion, '')
	FROM carritos_compras c
	WHERE c.empresa_id = ?
	  AND LOWER(COALESCE(c.estado_carrito, '')) IN ('cerrado', 'pagado', 'finalizado')
	ORDER BY COALESCE(c.pagado_en, c.fecha_actualizacion, c.fecha_creacion, '') DESC, c.id DESC
	LIMIT ?`, empresaID, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()

	out := make([]string, 0, limit)
	for rows.Next() {
		var codigo, metodo, fecha string
		var total float64
		if err := rows.Scan(&codigo, &total, &metodo, &fecha); err != nil {
			return nil
		}
		out = append(out, fmt.Sprintf("%s | total=%.2f | metodo=%s | fecha=%s", strings.TrimSpace(codigo), total, safeAIValue(metodo), safeAIValue(fecha)))
	}
	return out
}

func empresaAIAlertasInventario(dbConn *sql.DB, empresaID int64, availableTables []string, limit int) []string {
	if !slicesContain(availableTables, "productos") || !slicesContain(availableTables, "inventario_existencias") {
		return nil
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		COALESCE(p.nombre, ''),
		COALESCE(SUM(e.cantidad), 0) AS cantidad_actual,
		COALESCE(p.stock_minimo, 0) AS stock_minimo
	FROM productos p
	LEFT JOIN inventario_existencias e
	  ON e.producto_id = p.id
	 AND e.empresa_id = p.empresa_id
	WHERE p.empresa_id = ?
	  AND LOWER(COALESCE(p.estado, 'activo')) = 'activo'
	GROUP BY p.id, p.nombre, p.stock_minimo
	HAVING COALESCE(SUM(e.cantidad), 0) <= 0
	    OR (COALESCE(p.stock_minimo, 0) > 0 AND COALESCE(SUM(e.cantidad), 0) <= COALESCE(p.stock_minimo, 0))
	ORDER BY (COALESCE(p.stock_minimo, 0) - COALESCE(SUM(e.cantidad), 0)) DESC, p.nombre ASC
	LIMIT ?`, empresaID, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()

	out := make([]string, 0, limit)
	for rows.Next() {
		var nombre string
		var cantidad, minimo float64
		if err := rows.Scan(&nombre, &cantidad, &minimo); err != nil {
			return nil
		}
		out = append(out, fmt.Sprintf("%s | stock=%.2f | stock_minimo=%.2f", strings.TrimSpace(nombre), cantidad, minimo))
	}
	return out
}

func empresaAIFinanzasRecientes(dbConn *sql.DB, empresaID int64, availableTables []string, limit int) []string {
	if !slicesContain(availableTables, "empresa_finanzas_movimientos") {
		return nil
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		COALESCE(tipo_movimiento, ''),
		COALESCE(NULLIF(concepto, ''), NULLIF(descripcion, ''), NULLIF(numero_comprobante, ''), 'movimiento'),
		COALESCE(NULLIF(total_neto, 0), NULLIF(total, 0), monto, 0),
		COALESCE(fecha_movimiento, fecha_actualizacion, fecha_creacion, '')
	FROM empresa_finanzas_movimientos
	WHERE empresa_id = ?
	  AND LOWER(COALESCE(estado, 'activo')) = 'activo'
	ORDER BY COALESCE(fecha_movimiento, fecha_actualizacion, fecha_creacion, '') DESC, id DESC
	LIMIT ?`, empresaID, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()

	out := make([]string, 0, limit)
	for rows.Next() {
		var tipo, concepto, fecha string
		var total float64
		if err := rows.Scan(&tipo, &concepto, &total, &fecha); err != nil {
			return nil
		}
		out = append(out, fmt.Sprintf("%s | concepto=%s | total=%.2f | fecha=%s", safeAIValue(tipo), safeAIValue(concepto), total, safeAIValue(fecha)))
	}
	return out
}

func superAITopEmpresasVentas(dbConn *sql.DB, availableTables []string, limit int) []string {
	if !slicesContain(availableTables, "empresas") || !slicesContain(availableTables, "carritos_compras") {
		return nil
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		COALESCE(e.nombre, printf('empresa_%d', e.id)) AS empresa_nombre,
		COUNT(c.id) AS ventas,
		COALESCE(SUM(c.total), 0) AS total_vendido
	FROM empresas e
	LEFT JOIN carritos_compras c
	  ON c.empresa_id = e.id
	 AND LOWER(COALESCE(c.estado_carrito, '')) IN ('cerrado', 'pagado', 'finalizado')
	GROUP BY e.id, e.nombre
	ORDER BY total_vendido DESC, ventas DESC
	LIMIT ?`, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()

	out := make([]string, 0, limit)
	for rows.Next() {
		var nombre string
		var ventas int64
		var total float64
		if err := rows.Scan(&nombre, &ventas, &total); err != nil {
			return nil
		}
		out = append(out, fmt.Sprintf("%s | ventas=%d | total=%.2f", strings.TrimSpace(nombre), ventas, total))
	}
	return out
}

func superAIVentasRecientes(dbConn *sql.DB, availableTables []string, limit int) []string {
	if !slicesContain(availableTables, "empresas") || !slicesContain(availableTables, "carritos_compras") {
		return nil
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		COALESCE(e.nombre, printf('empresa_%d', e.id)) AS empresa_nombre,
		COALESCE(NULLIF(c.codigo, ''), printf('venta_%d', c.id)),
		COALESCE(c.total, 0),
		COALESCE(c.pagado_en, c.fecha_actualizacion, c.fecha_creacion, '')
	FROM carritos_compras c
	JOIN empresas e ON e.id = c.empresa_id
	WHERE LOWER(COALESCE(c.estado_carrito, '')) IN ('cerrado', 'pagado', 'finalizado')
	ORDER BY COALESCE(c.pagado_en, c.fecha_actualizacion, c.fecha_creacion, '') DESC, c.id DESC
	LIMIT ?`, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()

	out := make([]string, 0, limit)
	for rows.Next() {
		var empresa, codigo, fecha string
		var total float64
		if err := rows.Scan(&empresa, &codigo, &total, &fecha); err != nil {
			return nil
		}
		out = append(out, fmt.Sprintf("%s | %s | total=%.2f | fecha=%s", strings.TrimSpace(empresa), strings.TrimSpace(codigo), total, safeAIValue(fecha)))
	}
	return out
}

func superAIAlertasInventario(dbConn *sql.DB, availableTables []string, limit int) []string {
	if !slicesContain(availableTables, "empresas") || !slicesContain(availableTables, "productos") || !slicesContain(availableTables, "inventario_existencias") {
		return nil
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		empresa_nombre,
		COUNT(1) AS productos_alerta
	FROM (
		SELECT
			COALESCE(e.nombre, printf('empresa_%d', e.id)) AS empresa_nombre,
			p.id AS producto_id,
			COALESCE(SUM(ie.cantidad), 0) AS stock_actual,
			COALESCE(p.stock_minimo, 0) AS stock_minimo
		FROM productos p
		JOIN empresas e ON e.id = p.empresa_id
		LEFT JOIN inventario_existencias ie
		  ON ie.producto_id = p.id
		 AND ie.empresa_id = p.empresa_id
		WHERE LOWER(COALESCE(p.estado, 'activo')) = 'activo'
		GROUP BY e.id, e.nombre, p.id, p.stock_minimo
	) resumen
	WHERE stock_actual <= 0 OR (stock_minimo > 0 AND stock_actual <= stock_minimo)
	GROUP BY empresa_nombre
	ORDER BY productos_alerta DESC, empresa_nombre ASC
	LIMIT ?`, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()

	type item struct {
		empresa string
		count   int64
	}
	items := make([]item, 0)
	for rows.Next() {
		var empresa string
		var count int64
		if err := rows.Scan(&empresa, &count); err != nil {
			return nil
		}
		if count > 0 {
			items = append(items, item{empresa: strings.TrimSpace(empresa), count: count})
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].count == items[j].count {
			return items[i].empresa < items[j].empresa
		}
		return items[i].count > items[j].count
	})
	if len(items) > limit {
		items = items[:limit]
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, fmt.Sprintf("%s | productos_con_alerta=%d", item.empresa, item.count))
	}
	return out
}

func safeAIValue(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "sin_dato"
	}
	return v
}

type empresaAISafeIntent struct {
	Name  string
	Lines []string
}

func buildEmpresaAISafeIntentContext(dbConn *sql.DB, empresaID int64, pregunta string, usuarioCreador string) (string, error) {
	folded := aiFoldText(pregunta)
	if strings.TrimSpace(folded) == "" {
		return "", nil
	}

	availableTables, err := aiAvailableTables(dbConn, []string{
		"clientes",
		"productos",
		"inventario_existencias",
		"carritos_compras",
		"carrito_compra_items",
		"empresa_finanzas_movimientos",
	})
	if err != nil {
		return "", err
	}

	terms := aiExtractSearchTerms(pregunta)
	intents := make([]empresaAISafeIntent, 0, 4)

	if aiLooksLikeClientQuestion(folded) {
		lines := empresaAISafeVentasPorCliente(dbConn, empresaID, availableTables, terms, 5)
		if len(lines) == 0 {
			lines = empresaAITopClientes(dbConn, empresaID, availableTables, 5)
		}
		if len(lines) > 0 {
			intents = append(intents, empresaAISafeIntent{Name: "CONSULTA_SEGURA_CLIENTES", Lines: lines})
		}
	}

	if aiLooksLikeProductQuestion(folded) {
		lines := empresaAISafeProductoDetalle(dbConn, empresaID, availableTables, terms, 5)
		if len(lines) == 0 {
			lines = empresaAITopProductos(dbConn, empresaID, availableTables, 5)
		}
		if len(lines) > 0 {
			intents = append(intents, empresaAISafeIntent{Name: "CONSULTA_SEGURA_PRODUCTOS", Lines: lines})
		}
	}

	// Acción segura (best-effort): cambiar precio de un producto cuando viene explícito.
	// Reglas:
	// - Debe venir el nombre del producto entre comillas: "Cocacola pequeña"
	// - Debe venir un número en la pregunta: precio 3500 / a 3500 / $3500
	// - No ejecuta SQL libre; hace UPDATE parametrizado por empresa_id + nombre.
	if slicesContain(availableTables, "productos") && strings.Contains(folded, "precio") &&
		(strings.Contains(folded, "cambia") || strings.Contains(folded, "cambiale") || strings.Contains(folded, "cambiar")) {
		if actionLines := empresaAISafeUpdateProductoPrecio(dbConn, empresaID, pregunta, terms, usuarioCreador); len(actionLines) > 0 {
			intents = append(intents, empresaAISafeIntent{Name: "ACCION_SEGURA_ACTUALIZAR_PRECIO_PRODUCTO", Lines: actionLines})
		}
	}

	if aiLooksLikeNoRotationQuestion(folded) {
		lines := empresaAISafeProductosSinRotacion(dbConn, empresaID, availableTables, 30, 5)
		if len(lines) > 0 {
			intents = append(intents, empresaAISafeIntent{Name: "CONSULTA_SEGURA_PRODUCTOS_SIN_ROTACION_30D", Lines: lines})
		}
	}

	if aiLooksLikeInventoryQuestion(folded) {
		lines := empresaAIAlertasInventario(dbConn, empresaID, availableTables, 5)
		if len(lines) > 0 {
			intents = append(intents, empresaAISafeIntent{Name: "CONSULTA_SEGURA_ALERTAS_INVENTARIO", Lines: lines})
		}
	}

	if aiLooksLikeSalesQuestion(folded) {
		if strings.Contains(folded, "ayer") {
			if lines := empresaAISafeVentasAyer(dbConn, empresaID, availableTables); len(lines) > 0 {
				intents = append(intents, empresaAISafeIntent{Name: "CONSULTA_SEGURA_VENTAS_AYER", Lines: lines})
			}
		}
		lines := empresaAIVentasRecientes(dbConn, empresaID, availableTables, 5)
		if len(lines) > 0 {
			intents = append(intents, empresaAISafeIntent{Name: "CONSULTA_SEGURA_VENTAS_RECIENTES", Lines: lines})
		}
	}

	if aiLooksLikeFinanceQuestion(folded) {
		lines := empresaAISafeMovimientosFinancieros(dbConn, empresaID, availableTables, folded, 5)
		if len(lines) == 0 {
			lines = empresaAIFinanzasRecientes(dbConn, empresaID, availableTables, 5)
		}
		if len(lines) > 0 {
			intents = append(intents, empresaAISafeIntent{Name: "CONSULTA_SEGURA_FINANZAS", Lines: lines})
		}
	}

	if len(intents) == 0 {
		return "", nil
	}

	var b strings.Builder
	b.WriteString("CONSULTAS_SEGURAS_RESUELTAS\n")
	for _, intent := range intents {
		writeAIContextSection(&b, intent.Name, intent.Lines)
	}
	return b.String(), nil
}

var aiFirstNumberRE = regexp.MustCompile(`(?i)(?:\\$\\s*)?(\\d+(?:[\\.,]\\d+)?)`)

func truncateText(v string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(strings.TrimSpace(v))
	if len(r) <= max {
		return string(r)
	}
	return string(r[:max])
}

func empresaAISafeUpdateProductoPrecio(dbConn *sql.DB, empresaID int64, pregunta string, terms []string, usuarioCreador string) []string {
	if dbConn == nil || empresaID <= 0 {
		return nil
	}
	folded := aiFoldText(pregunta)
	if !strings.Contains(folded, "confirmar") && !strings.Contains(folded, "confirmo") && !strings.Contains(folded, "confirmación") {
		return []string{
			"resultado=requiere_confirmacion",
			"detalle=Para ejecutar esta acción escribe la misma instrucción agregando la palabra CONFIRMAR al final. Ejemplo: Cambiale el precio a \"Producto\" 3500 CONFIRMAR",
		}
	}
	if len(terms) == 0 {
		return []string{"resultado=sin_accion (falta nombre del producto entre comillas)"}
	}
	nmatch := aiFirstNumberRE.FindStringSubmatch(pregunta)
	if len(nmatch) < 2 {
		return []string{"resultado=sin_accion (falta valor numérico de precio)"}
	}
	raw := strings.ReplaceAll(nmatch[1], ".", "")
	raw = strings.ReplaceAll(raw, ",", ".")
	val, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || val < 0 {
		return []string{"resultado=sin_accion (precio inválido)"}
	}
	productoNombre := strings.TrimSpace(terms[0])
	if productoNombre == "" {
		return []string{"resultado=sin_accion (nombre vacío)"}
	}

	tx, err := dbConn.Begin()
	if err != nil {
		return []string{"resultado=error", "detalle=" + safeAIValue(err.Error())}
	}
	defer tx.Rollback()

	var productoID int64
	var costoAnterior, precioAnterior, impuestoAnterior float64
	row := queryRowTxSQLCompat(tx, `SELECT id, COALESCE(costo,0), COALESCE(precio,0), COALESCE(impuesto_porcentaje,0)
		FROM productos
		WHERE empresa_id = ? AND LOWER(COALESCE(nombre,'')) = LOWER(?) LIMIT 1`, empresaID, productoNombre)
	if scanErr := row.Scan(&productoID, &costoAnterior, &precioAnterior, &impuestoAnterior); scanErr != nil {
		if scanErr == sql.ErrNoRows {
			return []string{"resultado=sin_cambios", "detalle=Producto no encontrado", "producto=" + safeAIValue(productoNombre)}
		}
		return []string{"resultado=error", "detalle=" + safeAIValue(scanErr.Error())}
	}

	nowExpr := sqlNowExpr()
	if _, uerr := execTxSQLCompat(tx, "UPDATE productos SET precio = ?, fecha_actualizacion = "+nowExpr+" WHERE id = ? AND empresa_id = ?", val, productoID, empresaID); uerr != nil {
		return []string{"resultado=error", "detalle=" + safeAIValue(uerr.Error())}
	}

	u := strings.TrimSpace(usuarioCreador)
	if u == "" {
		u = "ia_chat"
	}
	ref := truncateText(strings.TrimSpace(pregunta), 180)
	_ = insertProductoPrecioHistorialTx(tx, ProductoPrecioHistorial{
		EmpresaID:        empresaID,
		ProductoID:       productoID,
		CostoAnterior:    costoAnterior,
		CostoNuevo:       costoAnterior,
		PrecioAnterior:   precioAnterior,
		PrecioNuevo:      val,
		ImpuestoAnterior: impuestoAnterior,
		ImpuestoNuevo:    impuestoAnterior,
		Motivo:           "ia_update_precio",
		Referencia:       ref,
		UsuarioCreador:   u,
		Estado:           "activo",
		Observaciones:    "Cambio ejecutado desde chat IA con confirmación explícita.",
	})

	// Auditoría empresarial: registrar el cambio como evento forense.
	metadataObj := map[string]interface{}{
		"origen":            "chat_ia",
		"motivo":            "ia_update_precio",
		"producto":          productoNombre,
		"producto_id":       productoID,
		"precio_anterior":   precioAnterior,
		"precio_nuevo":      val,
		"confirmacion":      true,
		"prompt_referencia": ref,
		"usuario_creador":   u,
		"tabla_afectada":    "productos",
		"campo_actualizado": "precio",
		"empresa_id":        empresaID,
	}
	if metaRaw, merr := json.Marshal(metadataObj); merr == nil && json.Valid(metaRaw) {
		_, _ = insertTxSQLCompat(tx, `INSERT INTO empresa_auditoria_eventos (
			empresa_id,
			modulo,
			accion,
			recurso,
			recurso_id,
			metodo_http,
			endpoint,
			resultado,
			codigo_http,
			request_id,
			ip_origen,
			user_agent,
			metadata_json,
			retencion_dias,
			fecha_evento,
			fecha_expiracion,
			fecha_creacion,
			fecha_actualizacion,
			usuario_creador,
			estado,
			observaciones
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			datetime('now','localtime'),
			datetime(datetime('now','localtime'), '+3650 days'),
			datetime('now','localtime'),
			datetime('now','localtime'),
			?, ?, ?
		)`,
			empresaID,
			"ia",
			"update_precio_producto",
			"productos",
			productoID,
			"IA",
			"/api/empresa/chat_con_inteligencia_artificial/consultar",
			"ok",
			200,
			"",
			"",
			"",
			string(metaRaw),
			int64(3650),
			u,
			"activo",
			"Cambio de precio ejecutado desde chat IA con confirmación explícita.",
		)
	}

	if err := tx.Commit(); err != nil {
		return []string{"resultado=error", "detalle=" + safeAIValue(err.Error())}
	}

	return []string{
		"resultado=ok",
		"producto=" + safeAIValue(productoNombre),
		fmt.Sprintf("precio_anterior=%.2f", precioAnterior),
		fmt.Sprintf("precio_nuevo=%.2f", val),
		fmt.Sprintf("producto_id=%d", productoID),
		"usuario=" + safeAIValue(u),
	}
}

func empresaAISafeVentasAyer(dbConn *sql.DB, empresaID int64, availableTables []string) []string {
	if !slicesContain(availableTables, "carritos_compras") {
		return nil
	}
	var total float64
	var count int64
	var q string
	var args []interface{}
	if isPostgresDialect() {
		q = `SELECT COUNT(1),
			COALESCE(SUM(COALESCE(NULLIF(total_pagado, 0), NULLIF(total, 0), 0)), 0)
			FROM carritos_compras
			WHERE empresa_id = ?
			  AND LOWER(COALESCE(estado, 'activo')) = 'activo'
			  AND LOWER(COALESCE(estado_carrito, '')) IN ('pagado','cerrado','finalizado')
			  AND DATE(COALESCE(NULLIF(pagado_en, ''), fecha_creacion)::timestamp) = (CURRENT_DATE - INTERVAL '1 day')::date`
		args = []interface{}{empresaID}
	} else {
		q = `SELECT COUNT(1),
			COALESCE(SUM(COALESCE(NULLIF(total_pagado, 0), NULLIF(total, 0), 0)), 0)
			FROM carritos_compras
			WHERE empresa_id = ?
			  AND LOWER(COALESCE(estado, 'activo')) = 'activo'
			  AND LOWER(COALESCE(estado_carrito, '')) IN ('pagado','cerrado','finalizado')
			  AND date(COALESCE(NULLIF(pagado_en,''), fecha_creacion)) = date('now','localtime','-1 day')`
		args = []interface{}{empresaID}
	}
	_ = queryRowSQLCompat(dbConn, q, args...).Scan(&count, &total)
	return []string{
		fmt.Sprintf("ventas_ayer_total=%.2f", total),
		fmt.Sprintf("ventas_ayer_transacciones=%d", count),
	}
}

func superAIEmpresasListadoSoloLectura(dbConn *sql.DB, limit int) []string {
	if dbConn == nil || limit <= 0 {
		return nil
	}
	ok, err := tableExists(dbConn, "empresas")
	if err != nil || !ok {
		return nil
	}
	rows, err := querySQLCompat(dbConn, `SELECT id, COALESCE(nombre, printf('empresa_%d', id)), COALESCE(CAST(nit AS TEXT), ''), COALESCE(estado, '')
		FROM empresas ORDER BY id ASC LIMIT ?`, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]string, 0, limit)
	for rows.Next() {
		var id int64
		var nombre, nit, estado string
		if err := rows.Scan(&id, &nombre, &nit, &estado); err != nil {
			return nil
		}
		out = append(out, fmt.Sprintf("id=%d nombre=%s nit=%s estado=%s", id, safeAIValue(nombre), safeAIValue(nit), safeAIValue(estado)))
	}
	return out
}

func superAIFinanzasMuestraGlobal(dbConn *sql.DB, availableTables []string, limit int) []string {
	if !slicesContain(availableTables, "empresa_finanzas_movimientos") || !slicesContain(availableTables, "empresas") {
		return nil
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		COALESCE(e.nombre, printf('empresa_%d', e.id)),
		m.id,
		COALESCE(m.tipo_movimiento, ''),
		COALESCE(NULLIF(m.concepto, ''), NULLIF(m.descripcion, ''), 'mov'),
		COALESCE(NULLIF(m.total_neto, 0), NULLIF(m.total, 0), m.monto, 0),
		COALESCE(m.fecha_movimiento, m.fecha_actualizacion, m.fecha_creacion, '')
	FROM empresa_finanzas_movimientos m
	JOIN empresas e ON e.id = m.empresa_id
	WHERE LOWER(COALESCE(m.estado, 'activo')) = 'activo'
	ORDER BY COALESCE(m.fecha_movimiento, m.fecha_actualizacion, m.fecha_creacion, '') DESC, m.id DESC
	LIMIT ?`, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]string, 0, limit)
	for rows.Next() {
		var enombre string
		var mid int64
		var tipo, concepto, fecha string
		var total float64
		if err := rows.Scan(&enombre, &mid, &tipo, &concepto, &total, &fecha); err != nil {
			return nil
		}
		out = append(out, fmt.Sprintf("empresa=%s mov_id=%d tipo=%s concepto=%s total=%.2f fecha=%s", safeAIValue(enombre), mid, safeAIValue(tipo), safeAIValue(concepto), total, safeAIValue(fecha)))
	}
	return out
}

func superAIProductosMuestraSoloLectura(dbConn *sql.DB, availableTables []string, limit int) []string {
	if !slicesContain(availableTables, "productos") || !slicesContain(availableTables, "empresas") {
		return nil
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		COALESCE(e.nombre, printf('empresa_%d', e.id)),
		p.id,
		COALESCE(p.nombre, ''),
		COALESCE(p.precio, 0)
	FROM productos p
	JOIN empresas e ON e.id = p.empresa_id
	WHERE LOWER(COALESCE(p.estado, 'activo')) = 'activo'
	ORDER BY p.id DESC
	LIMIT ?`, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]string, 0, limit)
	for rows.Next() {
		var enombre, pnombre string
		var pid int64
		var precio float64
		if err := rows.Scan(&enombre, &pid, &pnombre, &precio); err != nil {
			return nil
		}
		out = append(out, fmt.Sprintf("empresa=%s producto_id=%d nombre=%s precio=%.2f", safeAIValue(enombre), pid, safeAIValue(pnombre), precio))
	}
	return out
}

func superAIClientesResumenSoloLectura(dbConn *sql.DB, availableTables []string) []string {
	if !slicesContain(availableTables, "clientes") {
		return nil
	}
	var total int64
	if err := queryRowSQLCompat(dbConn, `SELECT COUNT(1) FROM clientes WHERE LOWER(COALESCE(estado, 'activo')) = 'activo'`).Scan(&total); err != nil {
		return nil
	}
	out := []string{fmt.Sprintf("clientes_activos_total=%d", total)}
	if !slicesContain(availableTables, "empresas") {
		return out
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		COALESCE(e.nombre, printf('empresa_%d', e.id)) AS en,
		COUNT(c.id) AS n
	FROM clientes c
	JOIN empresas e ON e.id = c.empresa_id
	WHERE LOWER(COALESCE(c.estado, 'activo')) = 'activo'
	GROUP BY e.id, e.nombre
	ORDER BY n DESC, e.nombre ASC
	LIMIT 8`)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		var n int64
		if err := rows.Scan(&name, &n); err != nil {
			return out
		}
		out = append(out, fmt.Sprintf("clientes_por_empresa=%s n=%d", safeAIValue(name), n))
	}
	return out
}

// buildSuperAIEmpresaSoloLecturaSnapshot añade bloques de solo lectura sobre la base empresas
// (agregados y filas puntuales acotadas), sin operaciones de escritura.
func buildSuperAIEmpresaSoloLecturaSnapshot(dbConn *sql.DB, availableTables []string) []empresaAISafeIntent {
	intents := make([]empresaAISafeIntent, 0, 6)
	if lines := superAIEmpresasListadoSoloLectura(dbConn, 35); len(lines) > 0 {
		intents = append(intents, empresaAISafeIntent{Name: "CONSULTA_SOLO_LECTURA_EMPRESAS_LISTA", Lines: lines})
	}
	if lines := superAIFinanzasMuestraGlobal(dbConn, availableTables, 12); len(lines) > 0 {
		intents = append(intents, empresaAISafeIntent{Name: "CONSULTA_SOLO_LECTURA_FINANZAS_MUESTRA", Lines: lines})
	}
	if lines := superAIProductosMuestraSoloLectura(dbConn, availableTables, 14); len(lines) > 0 {
		intents = append(intents, empresaAISafeIntent{Name: "CONSULTA_SOLO_LECTURA_PRODUCTOS_MUESTRA", Lines: lines})
	}
	if lines := superAIClientesResumenSoloLectura(dbConn, availableTables); len(lines) > 0 {
		intents = append(intents, empresaAISafeIntent{Name: "CONSULTA_SOLO_LECTURA_CLIENTES_RESUMEN", Lines: lines})
	}
	if lines := superAIVentasRecientes(dbConn, availableTables, 8); len(lines) > 0 {
		intents = append(intents, empresaAISafeIntent{Name: "CONSULTA_SOLO_LECTURA_VENTAS_RECIENTES", Lines: lines})
	}
	if lines := superAITopEmpresasVentas(dbConn, availableTables, 8); len(lines) > 0 {
		intents = append(intents, empresaAISafeIntent{Name: "CONSULTA_SOLO_LECTURA_TOP_EMPRESAS_VENTAS", Lines: lines})
	}
	if lines := superAIAlertasInventario(dbConn, availableTables, 6); len(lines) > 0 {
		intents = append(intents, empresaAISafeIntent{Name: "CONSULTA_SOLO_LECTURA_INVENTARIO_ALERTAS", Lines: lines})
	}
	return intents
}

func buildSuperAISafeIntentContext(dbConn *sql.DB, pregunta string, empresaSoloLectura bool) (string, error) {
	folded := aiFoldText(pregunta)
	availableTables, err := aiAvailableTables(dbConn, []string{
		"empresas",
		"productos",
		"inventario_existencias",
		"carritos_compras",
		"empresa_finanzas_movimientos",
		"clientes",
	})
	if err != nil {
		return "", err
	}

	intents := make([]empresaAISafeIntent, 0, 8)
	if empresaSoloLectura {
		intents = append(intents, buildSuperAIEmpresaSoloLecturaSnapshot(dbConn, availableTables)...)
	} else {
		if strings.TrimSpace(folded) == "" {
			return "", nil
		}
		if aiLooksLikeSalesQuestion(folded) {
			if lines := superAIVentasRecientes(dbConn, availableTables, 5); len(lines) > 0 {
				intents = append(intents, empresaAISafeIntent{Name: "CONSULTA_SEGURA_GLOBAL_VENTAS_RECIENTES", Lines: lines})
			}
			if lines := superAITopEmpresasVentas(dbConn, availableTables, 5); len(lines) > 0 {
				intents = append(intents, empresaAISafeIntent{Name: "CONSULTA_SEGURA_GLOBAL_EMPRESAS_TOP_VENTAS", Lines: lines})
			}
		}
		if aiLooksLikeInventoryQuestion(folded) || aiLooksLikeProductQuestion(folded) {
			if lines := superAIAlertasInventario(dbConn, availableTables, 5); len(lines) > 0 {
				intents = append(intents, empresaAISafeIntent{Name: "CONSULTA_SEGURA_GLOBAL_ALERTAS_INVENTARIO", Lines: lines})
			}
		}
		if aiLooksLikeFinanceQuestion(folded) {
			if lines := superAIFinanzasMuestraGlobal(dbConn, availableTables, 8); len(lines) > 0 {
				intents = append(intents, empresaAISafeIntent{Name: "CONSULTA_SEGURA_GLOBAL_FINANZAS_MUESTRA", Lines: lines})
			}
		}
		if aiLooksLikeClientQuestion(folded) {
			if lines := superAIClientesResumenSoloLectura(dbConn, availableTables); len(lines) > 0 {
				intents = append(intents, empresaAISafeIntent{Name: "CONSULTA_SEGURA_GLOBAL_CLIENTES_RESUMEN", Lines: lines})
			}
		}
	}
	if len(intents) == 0 {
		return "", nil
	}
	var b strings.Builder
	b.WriteString("CONSULTAS_SEGURAS_GLOBALES_RESUELTAS\n")
	if empresaSoloLectura {
		b.WriteString("nota=fragmentos de solo lectura de la base empresas; el modelo no puede editar ni eliminar datos; no sugieras escrituras.\n")
	}
	for _, intent := range intents {
		writeAIContextSection(&b, intent.Name, intent.Lines)
	}
	return b.String(), nil
}

func aiLooksLikeClientQuestion(folded string) bool {
	return strings.Contains(folded, "cliente") || strings.Contains(folded, "clientes")
}

func aiLooksLikeProductQuestion(folded string) bool {
	return strings.Contains(folded, "producto") || strings.Contains(folded, "productos") || strings.Contains(folded, "articulo") || strings.Contains(folded, "item")
}

func aiLooksLikeInventoryQuestion(folded string) bool {
	return strings.Contains(folded, "inventario") || strings.Contains(folded, "stock") || strings.Contains(folded, "existencia") || strings.Contains(folded, "reabaste")
}

func aiLooksLikeSalesQuestion(folded string) bool {
	return strings.Contains(folded, "venta") || strings.Contains(folded, "ventas") || strings.Contains(folded, "vend")
}

func aiLooksLikeFinanceQuestion(folded string) bool {
	return strings.Contains(folded, "finanza") || strings.Contains(folded, "flujo de caja") || strings.Contains(folded, "ingreso") || strings.Contains(folded, "egreso") || strings.Contains(folded, "gasto")
}

func aiLooksLikeNoRotationQuestion(folded string) bool {
	return strings.Contains(folded, "sin rotacion") || strings.Contains(folded, "no se vende") || strings.Contains(folded, "no se vend") || strings.Contains(folded, "poca rotacion") || strings.Contains(folded, "sin movimiento")
}

func aiFoldText(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	replacer := strings.NewReplacer(
		"á", "a",
		"é", "e",
		"í", "i",
		"ó", "o",
		"ú", "u",
		"ü", "u",
		"ñ", "n",
	)
	v = replacer.Replace(v)
	return strings.NewReplacer(
		"á", "a",
		"é", "e",
		"í", "i",
		"ó", "o",
		"ú", "u",
		"ü", "u",
		"ñ", "n",
	).Replace(v)
}

var aiQuotedTermRE = regexp.MustCompile(`["']([^"']{2,80})["']`)

func aiExtractSearchTerms(pregunta string) []string {
	matches := aiQuotedTermRE.FindAllStringSubmatch(pregunta, 4)
	seen := map[string]struct{}{}
	out := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		term := strings.TrimSpace(match[1])
		if term == "" {
			continue
		}
		folded := aiFoldText(term)
		if len([]rune(folded)) < 2 {
			continue
		}
		if _, ok := seen[folded]; ok {
			continue
		}
		seen[folded] = struct{}{}
		out = append(out, term)
	}
	return out
}

func aiLikePattern(term string) string {
	term = strings.TrimSpace(term)
	if term == "" {
		return "%"
	}
	return "%" + aiFoldText(term) + "%"
}

func empresaAISafeVentasPorCliente(dbConn *sql.DB, empresaID int64, availableTables, terms []string, limit int) []string {
	if !slicesContain(availableTables, "clientes") || !slicesContain(availableTables, "carritos_compras") {
		return nil
	}
	query := `SELECT
		COALESCE(NULLIF(cl.nombre_razon_social, ''), NULLIF(cl.numero_documento, ''), 'cliente_sin_nombre') AS cliente_nombre,
		COUNT(c.id) AS ventas,
		COALESCE(SUM(c.total), 0) AS total_vendido,
		COALESCE(MAX(COALESCE(c.pagado_en, c.fecha_actualizacion, c.fecha_creacion, '')), '') AS ultima_venta
	FROM clientes cl
	LEFT JOIN carritos_compras c
	  ON c.cliente_id = cl.id
	 AND c.empresa_id = cl.empresa_id
	 AND LOWER(COALESCE(c.estado_carrito, '')) IN ('cerrado', 'pagado', 'finalizado')
	WHERE cl.empresa_id = ?`
	args := []interface{}{empresaID}
	if len(terms) > 0 {
		query += ` AND (LOWER(COALESCE(cl.nombre_razon_social, '')) LIKE ? OR LOWER(COALESCE(cl.numero_documento, '')) LIKE ?)`
		pattern := aiLikePattern(terms[0])
		args = append(args, pattern, pattern)
	}
	query += ` GROUP BY COALESCE(NULLIF(cl.nombre_razon_social, ''), NULLIF(cl.numero_documento, ''), 'cliente_sin_nombre')
	ORDER BY total_vendido DESC, ventas DESC
	LIMIT ?`
	args = append(args, limit)

	rows, err := querySQLCompat(dbConn, query, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]string, 0, limit)
	for rows.Next() {
		var nombre, fecha string
		var ventas int64
		var total float64
		if err := rows.Scan(&nombre, &ventas, &total, &fecha); err != nil {
			return nil
		}
		out = append(out, fmt.Sprintf("%s | ventas=%d | total=%.2f | ultima_venta=%s", strings.TrimSpace(nombre), ventas, total, safeAIValue(fecha)))
	}
	return out
}

func empresaAISafeProductoDetalle(dbConn *sql.DB, empresaID int64, availableTables, terms []string, limit int) []string {
	if !slicesContain(availableTables, "productos") || len(terms) == 0 {
		return nil
	}
	productColumns, err := aiAvailableColumns(dbConn, "productos", []string{"sku", "codigo_barras", "stock_minimo"})
	if err != nil {
		return nil
	}
	itemColumns := []string{}
	if slicesContain(availableTables, "carrito_compra_items") {
		itemColumns, err = aiAvailableColumns(dbConn, "carrito_compra_items", []string{"codigo_item", "descripcion", "cantidad", "total_linea"})
		if err != nil {
			return nil
		}
	}

	hasSKU := slicesContain(productColumns, "sku")
	hasCodigoBarras := slicesContain(productColumns, "codigo_barras")
	hasStockMinimo := slicesContain(productColumns, "stock_minimo")
	hasVentaItems := slicesContain(availableTables, "carrito_compra_items") &&
		slicesContain(availableTables, "carritos_compras") &&
		slicesContain(itemColumns, "descripcion")

	skuSelect := `'sin_sku'`
	if hasSKU {
		skuSelect = `COALESCE(NULLIF(p.sku, ''), 'sin_sku')`
	}
	stockMinSelect := `0`
	if hasStockMinimo {
		stockMinSelect = `COALESCE(p.stock_minimo, 0)`
	}

	ventasJoin := `LEFT JOIN (
		SELECT '' AS producto_nombre, '' AS producto_codigo, 0 AS cantidad_vendida, 0 AS total_vendido
	) v ON 1=0`
	args := []interface{}{}
	if hasVentaItems {
		codigoItemExpr := `''`
		if slicesContain(itemColumns, "codigo_item") {
			codigoItemExpr = `LOWER(TRIM(COALESCE(ci.codigo_item, '')))`
		}
		cantidadExpr := `0`
		if slicesContain(itemColumns, "cantidad") {
			cantidadExpr = `COALESCE(SUM(ci.cantidad), 0)`
		}
		totalExpr := `0`
		if slicesContain(itemColumns, "total_linea") {
			totalExpr = `COALESCE(SUM(ci.total_linea), 0)`
		}
		productMatchExtra := ""
		if hasSKU {
			productMatchExtra = ` OR (COALESCE(p.sku, '') <> '' AND v.producto_codigo = LOWER(TRIM(COALESCE(p.sku, ''))))`
		}
		ventasJoin = fmt.Sprintf(`LEFT JOIN (
			SELECT
				LOWER(TRIM(COALESCE(ci.descripcion, ''))) AS producto_nombre,
				%s AS producto_codigo,
				%s AS cantidad_vendida,
				%s AS total_vendido
			FROM carrito_compra_items ci
			JOIN carritos_compras c ON c.id = ci.carrito_id AND c.empresa_id = ci.empresa_id
			WHERE ci.empresa_id = ?
			  AND LOWER(COALESCE(c.estado_carrito, '')) IN ('cerrado', 'pagado', 'finalizado')
			GROUP BY LOWER(TRIM(COALESCE(ci.descripcion, ''))), %s
		) v ON v.producto_nombre = LOWER(TRIM(COALESCE(p.nombre, '')))%s`,
			codigoItemExpr, cantidadExpr, totalExpr, codigoItemExpr, productMatchExtra)
		args = append(args, empresaID)
	}

	pattern := aiLikePattern(terms[0])
	whereParts := []string{`LOWER(COALESCE(p.nombre, '')) LIKE ?`}
	args = append(args, empresaID, pattern)
	if hasSKU {
		whereParts = append(whereParts, `LOWER(COALESCE(p.sku, '')) LIKE ?`)
		args = append(args, pattern)
	}
	if hasCodigoBarras {
		whereParts = append(whereParts, `LOWER(COALESCE(p.codigo_barras, '')) LIKE ?`)
		args = append(args, pattern)
	}
	args = append(args, limit)

	query := fmt.Sprintf(`SELECT
		COALESCE(p.nombre, ''),
		%s,
		COALESCE(SUM(ie.cantidad), 0) AS stock_actual,
		%s AS stock_minimo,
		COALESCE(MAX(v.cantidad_vendida), 0),
		COALESCE(MAX(v.total_vendido), 0)
	FROM productos p
	LEFT JOIN inventario_existencias ie
	  ON ie.producto_id = p.id
	 AND ie.empresa_id = p.empresa_id
	%s
	WHERE p.empresa_id = ?
	  AND (%s)
	GROUP BY p.id, p.nombre, %s, %s
	ORDER BY COALESCE(MAX(v.total_vendido), 0) DESC, p.nombre ASC
	LIMIT ?`, skuSelect, stockMinSelect, ventasJoin, strings.Join(whereParts, " OR "), skuSelect, stockMinSelect)

	rows, err := querySQLCompat(dbConn, query, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]string, 0, limit)
	for rows.Next() {
		var nombre, sku string
		var stock, minimo, cantidadVendida, totalVendido float64
		if err := rows.Scan(&nombre, &sku, &stock, &minimo, &cantidadVendida, &totalVendido); err != nil {
			return nil
		}
		out = append(out, fmt.Sprintf("%s | sku=%s | stock=%.2f | stock_minimo=%.2f | vendido=%.2f | total_vendido=%.2f", strings.TrimSpace(nombre), safeAIValue(sku), stock, minimo, cantidadVendida, totalVendido))
	}
	return out
}

func empresaAISafeProductosSinRotacion(dbConn *sql.DB, empresaID int64, availableTables []string, days, limit int) []string {
	if !slicesContain(availableTables, "productos") || !slicesContain(availableTables, "carrito_compra_items") || !slicesContain(availableTables, "carritos_compras") {
		return nil
	}
	productColumns, err := aiAvailableColumns(dbConn, "productos", []string{"sku", "stock_minimo", "estado"})
	if err != nil {
		return nil
	}
	itemColumns, err := aiAvailableColumns(dbConn, "carrito_compra_items", []string{"codigo_item", "descripcion"})
	if err != nil {
		return nil
	}
	if days <= 0 {
		days = 30
	}
	stockMinSelect := `0`
	if slicesContain(productColumns, "stock_minimo") {
		stockMinSelect = `COALESCE(p.stock_minimo, 0)`
	}
	estadoFilter := ""
	if slicesContain(productColumns, "estado") {
		estadoFilter = `AND LOWER(COALESCE(p.estado, 'activo')) = 'activo'`
	}
	productCodeMatch := ""
	if slicesContain(productColumns, "sku") && slicesContain(itemColumns, "codigo_item") {
		productCodeMatch = `OR (COALESCE(p.sku, '') <> '' AND LOWER(TRIM(COALESCE(ci.codigo_item, ''))) = LOWER(TRIM(COALESCE(p.sku, ''))))`
	}
	interval := fmt.Sprintf("-%d day", days)
	rows, err := querySQLCompat(dbConn, fmt.Sprintf(`SELECT
		COALESCE(p.nombre, ''),
		COALESCE(SUM(ie.cantidad), 0) AS stock_actual,
		%s AS stock_minimo
	FROM productos p
	LEFT JOIN inventario_existencias ie
	  ON ie.producto_id = p.id
	 AND ie.empresa_id = p.empresa_id
	WHERE p.empresa_id = ?
	  %s
	  AND NOT EXISTS (
		SELECT 1
		FROM carrito_compra_items ci
		JOIN carritos_compras c ON c.id = ci.carrito_id AND c.empresa_id = ci.empresa_id
		WHERE ci.empresa_id = p.empresa_id
		  AND LOWER(COALESCE(c.estado_carrito, '')) IN ('cerrado', 'pagado', 'finalizado')
		  AND date(COALESCE(c.pagado_en, c.fecha_actualizacion, c.fecha_creacion, '')) >= date('now', ?)
		  AND (
			LOWER(TRIM(COALESCE(ci.descripcion, ''))) = LOWER(TRIM(COALESCE(p.nombre, '')))
			%s
		  )
	  )
	GROUP BY p.id, p.nombre, %s
	ORDER BY stock_actual DESC, p.nombre ASC
	LIMIT ?`, stockMinSelect, estadoFilter, productCodeMatch, stockMinSelect), empresaID, interval, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]string, 0, limit)
	for rows.Next() {
		var nombre string
		var stock, minimo float64
		if err := rows.Scan(&nombre, &stock, &minimo); err != nil {
			return nil
		}
		out = append(out, fmt.Sprintf("%s | stock=%.2f | stock_minimo=%.2f | sin_ventas_%dd=true", strings.TrimSpace(nombre), stock, minimo, days))
	}
	return out
}

func empresaAISafeMovimientosFinancieros(dbConn *sql.DB, empresaID int64, availableTables []string, folded string, limit int) []string {
	if !slicesContain(availableTables, "empresa_finanzas_movimientos") {
		return nil
	}
	tipo := ""
	if strings.Contains(folded, "ingreso") {
		tipo = "ingreso"
	} else if strings.Contains(folded, "egreso") || strings.Contains(folded, "gasto") {
		tipo = "egreso"
	}

	query := `SELECT
		COALESCE(tipo_movimiento, ''),
		COALESCE(NULLIF(concepto, ''), NULLIF(descripcion, ''), NULLIF(numero_comprobante, ''), 'movimiento'),
		COALESCE(NULLIF(total_neto, 0), NULLIF(total, 0), monto, 0),
		COALESCE(fecha_movimiento, fecha_actualizacion, fecha_creacion, '')
	FROM empresa_finanzas_movimientos
	WHERE empresa_id = ?
	  AND LOWER(COALESCE(estado, 'activo')) = 'activo'`
	args := []interface{}{empresaID}
	if tipo != "" {
		query += ` AND LOWER(COALESCE(tipo_movimiento, '')) = ?`
		args = append(args, tipo)
	}
	query += ` ORDER BY COALESCE(fecha_movimiento, fecha_actualizacion, fecha_creacion, '') DESC, id DESC
	LIMIT ?`
	args = append(args, limit)
	rows, err := querySQLCompat(dbConn, query, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]string, 0, limit)
	for rows.Next() {
		var tipoMov, concepto, fecha string
		var total float64
		if err := rows.Scan(&tipoMov, &concepto, &total, &fecha); err != nil {
			return nil
		}
		out = append(out, fmt.Sprintf("%s | concepto=%s | total=%.2f | fecha=%s", safeAIValue(tipoMov), safeAIValue(concepto), total, safeAIValue(fecha)))
	}
	return out
}

func aiNormalizeProvider(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

func aiContextModelName(modelo ...string) string {
	for _, item := range modelo {
		if v := strings.TrimSpace(item); v != "" {
			return v
		}
	}
	return "openai:gpt-5.4-mini"
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
