package handlers

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestValidateNextcloudBaseURL(t *testing.T) {
	t.Setenv("PCS_NEXTCLOUD_ALLOW_PRIVATE_HOSTS", "")
	valid, err := validateNextcloudBaseURL("https://nube.example.com/nextcloud/")
	if err != nil || valid != "https://nube.example.com/nextcloud" {
		t.Fatalf("URL valida = %q, err=%v", valid, err)
	}
	for _, raw := range []string{
		"http://nube.example.com", "https://user:pass@nube.example.com",
		"https://nube.example.com/?token=secret", "https://127.0.0.1",
		"https://169.254.169.254/latest/meta-data",
	} {
		if _, err := validateNextcloudBaseURL(raw); err == nil {
			t.Fatalf("se acepto URL riesgosa %q", raw)
		}
	}
}

func TestNextcloudTemporaryPasswordUsesCryptographicEntropy(t *testing.T) {
	first, err := newNextcloudTemporaryPassword()
	if err != nil {
		t.Fatal(err)
	}
	second, err := newNextcloudTemporaryPassword()
	if err != nil {
		t.Fatal(err)
	}
	if first == second || strings.Contains(first, "empresa") {
		t.Fatal("credencial temporal predecible")
	}
	raw, err := base64.RawURLEncoding.DecodeString(first)
	if err != nil || len(raw) < 32 {
		t.Fatalf("entropia insuficiente: bytes=%d err=%v", len(raw), err)
	}
}

func TestNextcloudOCSValidatesEnvelopeAndDoesNotFollowRedirects(t *testing.T) {
	t.Setenv("PCS_NEXTCLOUD_ALLOW_PRIVATE_HOSTS", "true")
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("format") != "json" || r.Header.Get("OCS-APIRequest") != "true" {
			t.Error("faltan cabeceras OCS")
		}
		user, password, ok := r.BasicAuth()
		if !ok || user != "admin" || password != "credential" {
			t.Error("autenticacion OCS inesperada")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ocs":{"meta":{"status":"ok","statuscode":100,"message":"OK"},"data":{}}}`))
	}))
	defer server.Close()

	meta, err := nextcloudOCS(context.Background(), server.Client(), http.MethodPost, server.URL, "/ocs/v1.php/cloud/users", "admin", "credential", url.Values{"userid": {"pcs_empresa_42"}})
	if err != nil || !nextcloudOCSSuccess(meta) {
		t.Fatalf("respuesta OCS valida rechazada: meta=%+v err=%v", meta, err)
	}

	redirectTarget := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("el cliente siguio un redirect y pudo reenviar credenciales")
	}))
	defer redirectTarget.Close()
	redirectSource := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Redirect(w, httptest.NewRequest(http.MethodGet, redirectTarget.URL, nil), redirectTarget.URL, http.StatusFound)
	}))
	defer redirectSource.Close()
	client := redirectSource.Client()
	client.CheckRedirect = newNextcloudHTTPClient().CheckRedirect
	if _, err := nextcloudOCS(context.Background(), client, http.MethodGet, redirectSource.URL, "/ocs/v1.php/cloud/capabilities", "admin", "credential", nil); err == nil {
		t.Fatal("redirect OCS aceptado")
	}
}
