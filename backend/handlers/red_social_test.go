package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestPublicacionesRedSocialHandlerFiltersByEmpresaWhenProvided(t *testing.T) {
	dbEmp := openTestSQLite(t, "red_social_publica.db")
	if err := dbpkg.EnsureEmpresaPublicacionesRedSocialSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaPublicacionesRedSocialSchema: %v", err)
	}

	if _, err := dbEmp.Exec(`INSERT INTO empresa_publicaciones_red_social (empresa_id, nombre, descripcion, foto_url, estado) VALUES (?, ?, ?, ?, ?)`, 7, "Publicacion Calipso", "Visible en empresa 7", "", "activo"); err != nil {
		t.Fatalf("insert publicacion A: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO empresa_publicaciones_red_social (empresa_id, nombre, descripcion, foto_url, estado) VALUES (?, ?, ?, ?, ?)`, 9, "Publicacion Ajena", "Visible en empresa 9", "", "activo"); err != nil {
		t.Fatalf("insert publicacion B: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/public/publicaciones?empresa_id=7", nil)
	rr := httptest.NewRecorder()

	PublicacionesRedSocialHandler(dbEmp).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var pubs []dbpkg.PublicacionRedSocial
	if err := json.Unmarshal(rr.Body.Bytes(), &pubs); err != nil {
		t.Fatalf("decode publicaciones response: %v body=%s", err, rr.Body.String())
	}
	if len(pubs) != 1 {
		t.Fatalf("expected 1 publication, got %d", len(pubs))
	}
	if pubs[0].EmpresaID != 7 {
		t.Fatalf("expected empresa_id 7, got %d", pubs[0].EmpresaID)
	}
	if !strings.Contains(pubs[0].Nombre, "Calipso") {
		t.Fatalf("unexpected publication returned: %+v", pubs[0])
	}
}