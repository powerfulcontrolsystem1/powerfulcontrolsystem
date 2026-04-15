package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func ensureChatTareasTestBase(t *testing.T, dbEmp *sql.DB) {
	t.Helper()
	ensurePermsEmpresasSchema(t, dbEmp)
	ensureEmpresaUsersSchema(t, dbEmp)
	if err := dbpkg.EnsureEmpresaChatTareasSchema(dbEmp); err != nil {
		t.Fatalf("ensure chat_tareas schema: %v", err)
	}
}

func seedChatTareasUser(t *testing.T, dbEmp *sql.DB, empresaID int64, email, name string) int64 {
	t.Helper()
	res, err := dbEmp.Exec(`INSERT INTO users (
		email,
		name,
		role,
		empresa_id,
		documento_identidad,
		rol_usuario_id,
		email_confirmado,
		estado,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador
	) VALUES (?, ?, 'vendedor', ?, 'DOC-CHAT', 2, 1, 'activo', datetime('now','localtime'), datetime('now','localtime'), ?)
	`, strings.ToLower(strings.TrimSpace(email)), strings.TrimSpace(name), empresaID, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		t.Fatalf("insert users seed: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("users seed last insert id: %v", err)
	}
	return id
}

func cleanupChatTareasUploadArtifacts(t *testing.T, empresaID int64) {
	t.Helper()
	suffix := "empresa_" + strconv.FormatInt(empresaID, 10)
	bases := []string{
		filepath.Join("web", "uploads", "chat_tareas", suffix),
		filepath.Join("..", "web", "uploads", "chat_tareas", suffix),
	}
	t.Cleanup(func() {
		for _, base := range bases {
			_ = os.RemoveAll(base)
		}
	})
}

func TestEmpresaChatTareasMensajesHandlerDerivesUsuarioActor(t *testing.T) {
	dbEmp := openTestSQLite(t, "chat_tareas_mensajes_actor.db")
	ensureChatTareasTestBase(t, dbEmp)

	const empresaID int64 = 101
	const ownerEmail = "admin.owner@empresa.com"
	const userEmail = "vendedor@empresa.com"
	const userName = "Vendedor Uno"
	cleanupChatTareasUploadArtifacts(t, empresaID)

	seedPermsEmpresa(t, dbEmp, empresaID, ownerEmail)
	userID := seedChatTareasUser(t, dbEmp, empresaID, userEmail, userName)

	convID, err := dbpkg.CreateChatConversacion(dbEmp, dbpkg.ChatConversacion{
		EmpresaID:          empresaID,
		Titulo:             "Conversacion actor",
		Descripcion:        "Validar autor usuario",
		Prioridad:          "media",
		EstadoConversacion: "abierta",
		UsuarioCreador:     ownerEmail,
		Estado:             "activo",
	})
	if err != nil {
		t.Fatalf("create conversacion: %v", err)
	}

	h := EmpresaChatTareasMensajesHandler(dbEmp)
	body := `{"empresa_id":101,"conversacion_id":` + strconv.FormatInt(convID, 10) + `,"autor_tipo":"admin","autor_email":"spoof@empresa.com","autor_nombre":"Impostor","contenido":"Mensaje de prueba"}`
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/chat_tareas/mensajes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", userEmail))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rr.Code, rr.Body.String())
	}

	msgs, err := dbpkg.GetChatMensajes(dbEmp, empresaID, convID, true, 50, 0)
	if err != nil {
		t.Fatalf("get chat mensajes: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	msg := msgs[0]
	if msg.AutorTipo != "usuario" {
		t.Fatalf("expected autor_tipo usuario, got %q", msg.AutorTipo)
	}
	if msg.AutorRefID != userID {
		t.Fatalf("expected autor_ref_id %d, got %d", userID, msg.AutorRefID)
	}
	if !strings.EqualFold(msg.AutorEmail, userEmail) {
		t.Fatalf("expected autor_email %q, got %q", userEmail, msg.AutorEmail)
	}
	if strings.TrimSpace(msg.AutorNombre) != userName {
		t.Fatalf("expected autor_nombre %q, got %q", userName, msg.AutorNombre)
	}

	parts, err := dbpkg.GetChatParticipantes(dbEmp, empresaID, convID, true)
	if err != nil {
		t.Fatalf("get participantes: %v", err)
	}
	if len(parts) == 0 {
		t.Fatalf("expected sender participant auto-registration")
	}
}

func TestEmpresaChatTareasAdjuntoUploadAllowsDocx(t *testing.T) {
	dbEmp := openTestSQLite(t, "chat_tareas_adjunto_docx.db")
	ensureChatTareasTestBase(t, dbEmp)

	const empresaID int64 = 102
	const ownerEmail = "admin.owner@empresa.com"
	const userEmail = "operador@empresa.com"
	cleanupChatTareasUploadArtifacts(t, empresaID)

	seedPermsEmpresa(t, dbEmp, empresaID, ownerEmail)
	seedChatTareasUser(t, dbEmp, empresaID, userEmail, "Operador")

	convID, err := dbpkg.CreateChatConversacion(dbEmp, dbpkg.ChatConversacion{
		EmpresaID:          empresaID,
		Titulo:             "Conversacion adjuntos",
		Descripcion:        "Validar docx",
		Prioridad:          "media",
		EstadoConversacion: "abierta",
		UsuarioCreador:     ownerEmail,
		Estado:             "activo",
	})
	if err != nil {
		t.Fatalf("create conversacion: %v", err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("empresa_id", strconv.FormatInt(empresaID, 10)); err != nil {
		t.Fatalf("write empresa_id field: %v", err)
	}
	if err := writer.WriteField("conversacion_id", strconv.FormatInt(convID, 10)); err != nil {
		t.Fatalf("write conversacion_id field: %v", err)
	}
	if err := writer.WriteField("contenido", "Documento DOCX de prueba"); err != nil {
		t.Fatalf("write contenido field: %v", err)
	}
	part, err := writer.CreateFormFile("archivo", "manual.docx")
	if err != nil {
		t.Fatalf("create multipart file: %v", err)
	}
	if _, err := part.Write([]byte("contenido binario simulado")); err != nil {
		t.Fatalf("write multipart file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}

	h := EmpresaChatTareasAdjuntoUploadHandler(dbEmp)
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/chat_tareas/mensajes/adjunto", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", userEmail))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, rr.Body.String())
	}
	if got := strings.ToLower(strings.TrimSpace(toString(resp["tipo_archivo"]))); got != "archivo" {
		t.Fatalf("expected tipo_archivo=archivo, got %q", got)
	}

	msgs, err := dbpkg.GetChatMensajes(dbEmp, empresaID, convID, true, 50, 0)
	if err != nil {
		t.Fatalf("get chat mensajes: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message with attachment, got %d", len(msgs))
	}
	if len(msgs[0].Adjuntos) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(msgs[0].Adjuntos))
	}
	if !strings.HasSuffix(strings.ToLower(msgs[0].Adjuntos[0].NombreArchivo), ".docx") {
		t.Fatalf("expected .docx attachment, got %q", msgs[0].Adjuntos[0].NombreArchivo)
	}
}

func TestEmpresaChatTareasConversacionesAddsOwnerAdminParticipant(t *testing.T) {
	dbEmp := openTestSQLite(t, "chat_tareas_conversacion_owner_participant.db")
	ensureChatTareasTestBase(t, dbEmp)

	const empresaID int64 = 103
	const ownerEmail = "propietario@empresa.com"
	const userEmail = "agente@empresa.com"
	cleanupChatTareasUploadArtifacts(t, empresaID)

	seedPermsEmpresa(t, dbEmp, empresaID, ownerEmail)
	userID := seedChatTareasUser(t, dbEmp, empresaID, userEmail, "Agente Empresa")

	h := EmpresaChatTareasConversacionesHandler(dbEmp)
	payload := `{"empresa_id":103,"titulo":"Soporte cliente","descripcion":"Conversacion usuario-admin","prioridad":"alta"}`
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/chat_tareas/conversaciones", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", userEmail))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rr.Code, rr.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v body=%s", err, rr.Body.String())
	}
	convID := int64(body["id"].(float64))

	participants, err := dbpkg.GetChatParticipantes(dbEmp, empresaID, convID, true)
	if err != nil {
		t.Fatalf("get participantes: %v", err)
	}
	if len(participants) < 2 {
		t.Fatalf("expected at least 2 participants (usuario + admin propietario), got %d", len(participants))
	}

	foundUser := false
	foundOwnerAdmin := false
	for _, p := range participants {
		email := strings.ToLower(strings.TrimSpace(p.Email))
		tipo := strings.ToLower(strings.TrimSpace(p.ParticipanteTipo))
		if email == strings.ToLower(userEmail) && tipo == "usuario" && p.ParticipanteRefID == userID {
			foundUser = true
		}
		if email == strings.ToLower(ownerEmail) && tipo == "admin" {
			foundOwnerAdmin = true
		}
	}

	if !foundUser {
		t.Fatalf("expected usuario participant for %q", userEmail)
	}
	if !foundOwnerAdmin {
		t.Fatalf("expected admin participant for owner email %q", ownerEmail)
	}
}

func TestEmpresaChatTareasCitasSharedByEmpresa(t *testing.T) {
	dbEmp := openTestSQLite(t, "chat_tareas_citas_shared_empresa.db")
	ensureChatTareasTestBase(t, dbEmp)

	const empresaID int64 = 104
	const ownerEmail = "admin.owner@empresa.com"
	const userCreatorEmail = "agenda.creador@empresa.com"
	const userViewerEmail = "agenda.viewer@empresa.com"

	seedPermsEmpresa(t, dbEmp, empresaID, ownerEmail)
	seedChatTareasUser(t, dbEmp, empresaID, userCreatorEmail, "Creador Agenda")
	seedChatTareasUser(t, dbEmp, empresaID, userViewerEmail, "Viewer Agenda")

	h := EmpresaChatTareasCitasHandler(dbEmp)

	createBody := `{"empresa_id":104,"titulo":"Reunion semanal","descripcion":"Seguimiento comercial","tipo_cita":"reunion","fecha_inicio":"2026-04-20 10:00:00","fecha_fin":"2026-04-20 11:00:00","notificar_minutos_antes":45}`
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/empresa/chat_tareas/citas", strings.NewReader(createBody))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate = reqCreate.WithContext(context.WithValue(reqCreate.Context(), "adminEmail", userCreatorEmail))
	rrCreate := httptest.NewRecorder()

	h.ServeHTTP(rrCreate, reqCreate)

	if rrCreate.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rrCreate.Code, rrCreate.Body.String())
	}

	var createResp map[string]interface{}
	if err := json.Unmarshal(rrCreate.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v body=%s", err, rrCreate.Body.String())
	}
	id := int64(createResp["id"].(float64))
	if id <= 0 {
		t.Fatalf("expected cita id > 0, got %d", id)
	}

	listURL := "/api/empresa/chat_tareas/citas?empresa_id=104&include_inactive=1"
	reqList := httptest.NewRequest(http.MethodGet, listURL, nil)
	reqList = reqList.WithContext(context.WithValue(reqList.Context(), "adminEmail", userViewerEmail))
	rrList := httptest.NewRecorder()

	h.ServeHTTP(rrList, reqList)

	if rrList.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rrList.Code, rrList.Body.String())
	}

	var citas []dbpkg.ChatCita
	if err := json.Unmarshal(rrList.Body.Bytes(), &citas); err != nil {
		t.Fatalf("decode citas response: %v body=%s", err, rrList.Body.String())
	}
	if len(citas) != 1 {
		t.Fatalf("expected 1 cita, got %d", len(citas))
	}
	if strings.TrimSpace(citas[0].Titulo) != "Reunion semanal" {
		t.Fatalf("expected titulo 'Reunion semanal', got %q", citas[0].Titulo)
	}
	if !strings.EqualFold(strings.TrimSpace(citas[0].CreadoPorEmail), userCreatorEmail) {
		t.Fatalf("expected creado_por_email %q, got %q", userCreatorEmail, citas[0].CreadoPorEmail)
	}

	workflowURL := "/api/empresa/chat_tareas/citas?empresa_id=104&id=" + strconv.FormatInt(id, 10) + "&action=cancelar"
	reqWorkflow := httptest.NewRequest(http.MethodPut, workflowURL, nil)
	reqWorkflow = reqWorkflow.WithContext(context.WithValue(reqWorkflow.Context(), "adminEmail", userViewerEmail))
	rrWorkflow := httptest.NewRecorder()

	h.ServeHTTP(rrWorkflow, reqWorkflow)

	if rrWorkflow.Code != http.StatusOK {
		t.Fatalf("expected 200 workflow update, got %d body=%s", rrWorkflow.Code, rrWorkflow.Body.String())
	}

	listCanceledURL := "/api/empresa/chat_tareas/citas?empresa_id=104&include_inactive=1&estado_cita=cancelada"
	reqCanceled := httptest.NewRequest(http.MethodGet, listCanceledURL, nil)
	reqCanceled = reqCanceled.WithContext(context.WithValue(reqCanceled.Context(), "adminEmail", userCreatorEmail))
	rrCanceled := httptest.NewRecorder()

	h.ServeHTTP(rrCanceled, reqCanceled)

	if rrCanceled.Code != http.StatusOK {
		t.Fatalf("expected 200 canceled list, got %d body=%s", rrCanceled.Code, rrCanceled.Body.String())
	}

	var canceled []dbpkg.ChatCita
	if err := json.Unmarshal(rrCanceled.Body.Bytes(), &canceled); err != nil {
		t.Fatalf("decode canceled citas response: %v body=%s", err, rrCanceled.Body.String())
	}
	if len(canceled) != 1 {
		t.Fatalf("expected 1 canceled cita, got %d", len(canceled))
	}
	if normalize := strings.ToLower(strings.TrimSpace(canceled[0].EstadoCita)); normalize != "cancelada" {
		t.Fatalf("expected estado_cita cancelada, got %q", canceled[0].EstadoCita)
	}
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch typed := v.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return strings.TrimSpace("")
	}
}
