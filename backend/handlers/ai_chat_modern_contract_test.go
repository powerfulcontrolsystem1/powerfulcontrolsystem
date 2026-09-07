package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmpresaAIConversationIDValidation(t *testing.T) {
	valid := "chat-123e4567-e89b-12d3-a456-426614174000"
	if got := normalizeEmpresaAIConversationID(valid); got != valid {
		t.Fatalf("conversation id = %q, want %q", got, valid)
	}
	for _, value := range []string{"", "short", "chat con espacios", "../../otra-ruta", strings.Repeat("a", 97)} {
		if got := normalizeEmpresaAIConversationID(value); got != "" {
			t.Fatalf("invalid conversation id %q was accepted as %q", value, got)
		}
	}
}

func TestEmpresaAIHistoryScopeRequiresAdministrativePermission(t *testing.T) {
	for _, raw := range []string{"", "usuario", " USUARIO "} {
		got, err := normalizeEmpresaAIHistoryScope(raw, false)
		if err != nil || got != "usuario" {
			t.Fatalf("scope %q = %q, %v; want usuario", raw, got, err)
		}
	}
	if _, err := normalizeEmpresaAIHistoryScope("empresa", false); err == nil {
		t.Fatal("company-wide history was allowed without administrative permission")
	}
	if got, err := normalizeEmpresaAIHistoryScope("empresa", true); err != nil || got != "empresa" {
		t.Fatalf("admin company scope = %q, %v; want empresa", got, err)
	}
	if _, err := normalizeEmpresaAIHistoryScope("todos", true); err == nil {
		t.Fatal("unknown history scope was accepted")
	}
}

func TestAIChatUsesOneGlobalMultimodalModelAndUserHistory(t *testing.T) {
	controllerRaw, err := os.ReadFile("chat_con_inteligencia_artificial_controller.go")
	if err != nil {
		t.Fatal(err)
	}
	controller := string(controllerRaw)
	for _, marker := range []string{
		`ConversationID string`,
		`"history_scope":`,
		`"can_view_company_history":`,
		`ListEmpresaAIConsultasRecientesPorUsuario`,
		`ListEmpresaAIConsultasRecientes(c.dbEmp, empresaID, limit)`,
		`"global_super"`,
	} {
		if !strings.Contains(controller, marker) {
			t.Fatalf("enterprise chat controller missing marker %q", marker)
		}
	}

	logicRaw, err := os.ReadFile("super_chat_ia_logica.go")
	if err != nil {
		t.Fatal(err)
	}
	logic := string(logicRaw)
	if !strings.Contains(logic, `defaultChatIAEmpresaModeloOperacion`) ||
		!strings.Contains(logic, `defaultChatIAEmpresaModeloAdjuntos`) ||
		strings.Count(logic, `= "openai:gpt-5.6-terra"`) < 2 {
		t.Fatal("the default global chat and attachment model must be GPT-5.6 Terra")
	}

	drawerRaw, err := os.ReadFile(filepath.Join("..", "..", "web", "js", "ai_chat_drawer.js"))
	if err != nil {
		t.Fatal(err)
	}
	drawer := string(drawerRaw)
	for _, marker := range []string{
		`conversation_id: state.conversationID`,
		`formData.set('conversation_id', state.conversationID)`,
		`function loadChatHistory(scope)`,
		`Todos los usuarios`,
		`modelField.hidden = true`,
	} {
		if !strings.Contains(drawer, marker) {
			t.Fatalf("chat drawer missing marker %q", marker)
		}
	}
}
