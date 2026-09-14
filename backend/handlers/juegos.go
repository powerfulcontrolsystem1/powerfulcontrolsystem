package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

var juegosPermitidos = map[string]bool{"pacman": true, "tetris": true, "buscaminas": true, "solitario": true, "selva": true, "sorpresa": true}

// Juegos is a free personal accessory, independent of business permissions and
// licences. Typed, active session identity is mandatory even for leaderboards.
func juegoIdentity(r *http.Request, super, emp *sql.DB) (string, string, error) {
	c, err := r.Cookie("session_token")
	if err != nil || super == nil {
		return "", "", errors.New("sesión requerida")
	}
	s, err := dbpkg.GetSessionByToken(super, c.Value)
	if err != nil || s == nil {
		return "", "", errors.New("sesión inválida")
	}
	if s.PrincipalType == "empresa_usuario" {
		u, err := dbpkg.GetEmpresaUsuarioByID(emp, s.EmpresaID, s.PrincipalID)
		if err != nil || u == nil || !strings.EqualFold(u.Estado, "activo") || u.EmailConfirmado != 1 || !strings.EqualFold(u.Email, s.AdminEmail) {
			return "", "", errors.New("usuario inactivo")
		}
		var active bool
		err = emp.QueryRowContext(r.Context(), `SELECT EXISTS(SELECT 1 FROM empresas WHERE COALESCE(empresa_id,id)=$1 AND lower(trim(COALESCE(estado,'activo')))='activo')`, s.EmpresaID).Scan(&active)
		if err != nil || !active {
			return "", "", errors.New("empresa inactiva")
		}
		return fmt.Sprintf("empresa:%d:usuario:%d", s.EmpresaID, u.ID), juegoDisplayName(u.Nombre), nil
	}
	if s.PrincipalType != "admin" && s.PrincipalType != "" {
		return "", "", errors.New("identidad inválida")
	}
	a, err := dbpkg.GetAdminAuthorizationIdentity(super, s.AdminEmail)
	if err != nil || a == nil || !strings.EqualFold(a.Estado, "activo") || (s.PrincipalID > 0 && s.PrincipalID != a.ID) {
		return "", "", errors.New("usuario inactivo")
	}
	if err := super.QueryRowContext(r.Context(), `SELECT COALESCE(name,'') FROM administradores WHERE id=$1`, a.ID).Scan(&a.Name); err != nil {
		return "", "", err
	}
	return fmt.Sprintf("admin:%d", a.ID), juegoDisplayName(a.Name), nil
}

func juegoDisplayName(name string) string {
	name = strings.Join(strings.Fields(name), " ")
	if name == "" || strings.Contains(name, "@") {
		return "Jugador PCS"
	}
	runes := []rune(name)
	if len(runes) > 80 {
		name = string(runes[:80])
	}
	return name
}

type juegoSaveRequest struct {
	Version int64           `json:"version"`
	Puntaje int64           `json:"puntaje"`
	Estado  json.RawMessage `json:"estado"`
}

func decodeJuegoSave(w http.ResponseWriter, r *http.Request) (juegoSaveRequest, error) {
	var p juegoSaveRequest
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		return p, errors.New("se requiere JSON")
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512<<10))
	d.DisallowUnknownFields()
	if err := d.Decode(&p); err != nil {
		return p, err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return p, errors.New("JSON adicional")
	}
	var state struct {
		Schema int    `json:"schema"`
		Game   string `json:"game"`
		Score  int64  `json:"score"`
	}
	if json.Unmarshal(p.Estado, &state) != nil || state.Schema != 1 || state.Game != r.URL.Query().Get("juego") || state.Score != p.Puntaje || p.Version < 0 || p.Puntaje < 0 || p.Puntaje > 100000000 {
		return p, errors.New("estado inválido")
	}
	return p, nil
}

func JuegosHandler(super, emp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private, no-store")
		if r.Method != http.MethodGet && r.Method != http.MethodPut {
			w.Header().Set("Allow", "GET, PUT")
			http.Error(w, "Método no permitido", 405)
			return
		}
		owner, name, err := juegoIdentity(r, super, emp)
		if err != nil {
			http.Error(w, "Inicia sesión para entrar a Juegos", 401)
			return
		}
		// There is no caller-supplied identity or company scope for personal storage.
		for key := range r.URL.Query() {
			if key != "juego" && key != "action" {
				http.Error(w, "Parámetro no permitido", 400)
				return
			}
		}
		game := r.URL.Query().Get("juego")
		action := r.URL.Query().Get("action")
		if !juegosPermitidos[game] {
			http.Error(w, "Juego no válido", 400)
			return
		}
		if r.Method == http.MethodGet {
			if action == "records" {
				items, err := dbpkg.GetJuegoRecords(r.Context(), super, game)
				if err != nil {
					http.Error(w, "No se pueden cargar los récords", 503)
					return
				}
				writeJSON(w, 200, map[string]any{"records": items})
				return
			}
			if action != "" {
				http.Error(w, "Acción no válida", 400)
				return
			}
			p, err := dbpkg.GetJuegoPartida(r.Context(), super, owner, game)
			if err != nil {
				http.Error(w, "No se puede cargar la partida", 503)
				return
			}
			writeJSON(w, 200, map[string]any{"partida": p, "nombre": name})
			return
		}
		if action != "" || r.Header.Get("Sec-Fetch-Site") == "cross-site" {
			http.Error(w, "Acción no permitida", 403)
			return
		}
		p, err := decodeJuegoSave(w, r)
		if err != nil {
			http.Error(w, "Partida o puntaje inválido", 400)
			return
		}
		version, err := dbpkg.SaveJuegoPartida(r.Context(), super, owner, name, game, p.Version, p.Puntaje, p.Estado)
		if errors.Is(err, dbpkg.ErrJuegoConflict) {
			http.Error(w, "La partida cambió en otro dispositivo. Recarga antes de continuar.", 409)
			return
		}
		if err != nil {
			http.Error(w, "No se pudo guardar la partida", 503)
			return
		}
		writeJSON(w, 200, map[string]any{"version": version, "ok": true})
	}
}
