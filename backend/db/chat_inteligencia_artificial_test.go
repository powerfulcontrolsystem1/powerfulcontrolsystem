package db

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func openChatIATestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "chat_ia_test.db")
	dbConn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	dbConn.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = dbConn.Close()
	})
	return dbConn
}

func TestEmpresaAIModeloPreferidoUpsertAndGet(t *testing.T) {
	dbConn := openChatIATestDB(t)
	if err := EnsureEmpresaAIChatSchema(dbConn); err != nil {
		t.Fatalf("ensure chat ia schema: %v", err)
	}

	modelID, err := GetEmpresaAIModeloPreferido(dbConn, 10, "admin@example.com")
	if err != nil {
		t.Fatalf("get modelo preferido vacio: %v", err)
	}
	if modelID != "" {
		t.Fatalf("expected modelo preferido vacio, got %q", modelID)
	}

	if err := UpsertEmpresaAIModeloPreferido(dbConn, 10, "Admin@Example.com", "google:gemini-2.0-flash", "tester"); err != nil {
		t.Fatalf("upsert modelo preferido inicial: %v", err)
	}
	modelID, err = GetEmpresaAIModeloPreferido(dbConn, 10, "admin@example.com")
	if err != nil {
		t.Fatalf("get modelo preferido inicial: %v", err)
	}
	if modelID != "google:gemini-2.0-flash" {
		t.Fatalf("expected model google:gemini-2.0-flash, got %q", modelID)
	}

	if err := UpsertEmpresaAIModeloPreferido(dbConn, 10, "admin@example.com", "google:gemini-1.5-flash", "tester"); err != nil {
		t.Fatalf("upsert modelo preferido update: %v", err)
	}
	modelID, err = GetEmpresaAIModeloPreferido(dbConn, 10, "admin@example.com")
	if err != nil {
		t.Fatalf("get modelo preferido update: %v", err)
	}
	if modelID != "google:gemini-1.5-flash" {
		t.Fatalf("expected model google:gemini-1.5-flash, got %q", modelID)
	}

	var provider string
	err = dbConn.QueryRow(`SELECT COALESCE(provider, '') FROM empresa_ai_modelo_preferido WHERE empresa_id = ? AND admin_email = ? LIMIT 1`, 10, "admin@example.com").Scan(&provider)
	if err != nil {
		t.Fatalf("query provider preferido: %v", err)
	}
	if provider != "google" {
		t.Fatalf("expected provider google, got %q", provider)
	}
}

func TestRegisterEmpresaAIConsultaAcumulaUsoDiario(t *testing.T) {
	dbConn := openChatIATestDB(t)
	if err := EnsureEmpresaAIChatSchema(dbConn); err != nil {
		t.Fatalf("ensure chat ia schema: %v", err)
	}

	_, err := RegisterEmpresaAIConsulta(dbConn, EmpresaAIConsulta{
		EmpresaID:        6,
		Provider:         "google",
		ModelID:          "google:gemini-2.0-flash",
		Pregunta:         "primera pregunta",
		Respuesta:        "primera respuesta",
		PromptTokens:     8,
		CompletionTokens: 12,
		TotalTokens:      20,
		FechaConsulta:    "2026-04-03 10:00:00",
		PlanActual:       "free",
		UsuarioCreador:   "admin@example.com",
		Estado:           "activo",
	})
	if err != nil {
		t.Fatalf("register consulta 1: %v", err)
	}

	_, err = RegisterEmpresaAIConsulta(dbConn, EmpresaAIConsulta{
		EmpresaID:        6,
		Provider:         "google",
		ModelID:          "google:gemini-2.0-flash",
		Pregunta:         "segunda pregunta",
		Respuesta:        "segunda respuesta",
		PromptTokens:     5,
		CompletionTokens: 7,
		TotalTokens:      12,
		FechaConsulta:    "2026-04-03 11:00:00",
		PlanActual:       "free",
		UsuarioCreador:   "admin@example.com",
		Estado:           "activo",
	})
	if err != nil {
		t.Fatalf("register consulta 2: %v", err)
	}

	uso, err := GetEmpresaAIUsoDiario(dbConn, 6, "google", "google:gemini-2.0-flash", "2026-04-03")
	if err != nil {
		t.Fatalf("get uso diario: %v", err)
	}
	if uso.Consultas != 2 {
		t.Fatalf("expected 2 consultas, got %d", uso.Consultas)
	}
	if uso.TokensTotal != 32 {
		t.Fatalf("expected 32 tokens acumulados, got %d", uso.TokensTotal)
	}

	rows, err := ListEmpresaAIConsultasRecientes(dbConn, 6, 10)
	if err != nil {
		t.Fatalf("list consultas recientes: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 consultas recientes, got %d", len(rows))
	}
}
