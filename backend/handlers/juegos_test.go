package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJuegosRejectAnonymous(t *testing.T) {
	for _, action := range []string{"", "&action=records"} {
		r := httptest.NewRequest(http.MethodGet, "/api/juegos?juego=pacman"+action, nil)
		w := httptest.NewRecorder()
		JuegosHandler(nil, nil)(w, r)
		if w.Code != 401 {
			t.Fatalf("anonymous %s got %d", action, w.Code)
		}
	}
}
func TestJuegosPayload(t *testing.T) {
	cases := []struct {
		body string
		ok   bool
	}{
		{`{"version":0,"puntaje":100,"estado":{"schema":1,"game":"pacman","score":100,"data":{}}}`, true},
		{`{"version":0,"puntaje":-1,"estado":{"schema":1,"game":"pacman","score":-1}}`, false},
		{`{"version":0,"puntaje":1,"usuario_id":"otra","estado":{"schema":1,"game":"pacman","score":1}}`, false},
		{`{"version":0,"puntaje":1,"empresa_id":2,"estado":{"schema":1,"game":"pacman","score":1}}`, false},
		{`{"version":0,"puntaje":1,"estado":{"schema":1,"game":"tetris","score":1}}`, false},
		{`{"version":0,"puntaje":1,"estado":{"schema":1,"game":"pacman","score":2}}`, false},
		{`{"version":0,"puntaje":1,"estado":null}`, false},
		{`{"version":0,"puntaje":1,"estado":{"schema":1,"game":"pacman","score":1}} {}`, false},
		{`{"version":0,"puntaje":100000001,"estado":{"schema":1,"game":"pacman","score":100000001}}`, false},
	}
	for _, tc := range cases {
		r := httptest.NewRequest(http.MethodPut, "/api/juegos?juego=pacman", strings.NewReader(tc.body))
		r.Header.Set("Content-Type", "application/json")
		_, err := decodeJuegoSave(httptest.NewRecorder(), r)
		if (err == nil) != tc.ok {
			t.Errorf("valid=%v for %s", err == nil, tc.body)
		}
	}
}
func TestJuegosDisplayName(t *testing.T) {
	if juegoDisplayName("private@example.test") != "Jugador PCS" {
		t.Fatal("email exposed")
	}
	if juegoDisplayName(" Ana   Pérez ") != "Ana Pérez" {
		t.Fatal("invalid display name")
	}
}
