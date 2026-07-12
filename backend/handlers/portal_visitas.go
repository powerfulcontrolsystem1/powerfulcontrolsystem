package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

type portalVisitasCountryRow struct {
	PaisCodigo string `json:"pais_codigo"`
	PaisNombre string `json:"pais_nombre"`
	Visitas    int64  `json:"visitas"`
}

type portalVisitasResponse struct {
	Ok             bool                      `json:"ok"`
	TotalVisitas   int64                     `json:"total_visitas"`
	Paises         []portalVisitasCountryRow `json:"paises"`
	PaisRegistrado string                    `json:"pais_registrado,omitempty"`
}

var portalVisitasSchema = struct {
	sync.Mutex
	ready bool
}{}

func PublicPortalVisitasHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if db == nil {
			http.Error(w, `{"ok":false,"error":"base de datos no disponible"}`, http.StatusServiceUnavailable)
			return
		}
		if err := ensurePortalVisitasSchema(db); err != nil {
			http.Error(w, `{"ok":false,"error":"no se pudo preparar contador de visitas"}`, http.StatusInternalServerError)
			return
		}

		switch r.Method {
		case http.MethodGet:
			writePortalVisitasResponse(w, db, "")
		case http.MethodPost:
			country := detectPortalVisitCountry(r)
			if country == "" {
				country = "CO"
			}
			if err := incrementPortalVisitCountry(db, country); err != nil {
				http.Error(w, `{"ok":false,"error":"no se pudo registrar la visita"}`, http.StatusInternalServerError)
				return
			}
			writePortalVisitasResponse(w, db, country)
		default:
			http.Error(w, `{"ok":false,"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	}
}

func ensurePortalVisitasSchema(db *sql.DB) error {
	portalVisitasSchema.Lock()
	defer portalVisitasSchema.Unlock()
	if portalVisitasSchema.ready {
		return nil
	}
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS portal_visitas_paises (
	pais_codigo TEXT NOT NULL,
	fecha DATE NOT NULL DEFAULT CURRENT_DATE,
	visitas BIGINT NOT NULL DEFAULT 0,
	actualizado_en TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (pais_codigo, fecha)
)`)
	if err != nil {
		return err
	}
	portalVisitasSchema.ready = true
	return nil
}

func detectPortalVisitCountry(r *http.Request) string {
	if r == nil {
		return ""
	}
	if query := normalizeCountryCode(r.URL.Query().Get("pais")); query != "" {
		return query
	}
	var payload struct {
		PaisCodigo string `json:"pais_codigo"`
		Country    string `json:"country"`
	}
	if strings.Contains(strings.ToLower(r.Header.Get("Content-Type")), "application/json") && r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&payload)
		if bodyCountry := normalizeCountryCode(payload.PaisCodigo); bodyCountry != "" {
			return bodyCountry
		}
		if bodyCountry := normalizeCountryCode(payload.Country); bodyCountry != "" {
			return bodyCountry
		}
	}
	for _, header := range []string{"CF-IPCountry", "X-Country-Code", "X-Geo-Country"} {
		if country := normalizeCountryCode(r.Header.Get(header)); country != "" {
			return country
		}
	}
	return ""
}

func incrementPortalVisitCountry(db *sql.DB, country string) error {
	_, err := db.Exec(`
INSERT INTO portal_visitas_paises (pais_codigo, fecha, visitas, actualizado_en)
VALUES ($1, CURRENT_DATE, 1, NOW())
ON CONFLICT (pais_codigo, fecha)
DO UPDATE SET visitas = portal_visitas_paises.visitas + 1, actualizado_en = NOW()`, country)
	return err
}

func writePortalVisitasResponse(w http.ResponseWriter, db *sql.DB, registeredCountry string) {
	rows, err := db.Query(`
SELECT pais_codigo, SUM(visitas)::BIGINT AS visitas
FROM portal_visitas_paises
GROUP BY pais_codigo
ORDER BY visitas DESC, pais_codigo ASC
LIMIT 48`)
	if err != nil {
		http.Error(w, `{"ok":false,"error":"no se pudieron cargar las visitas"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	resp := portalVisitasResponse{Ok: true, Paises: []portalVisitasCountryRow{}, PaisRegistrado: registeredCountry}
	for rows.Next() {
		var item portalVisitasCountryRow
		if err := rows.Scan(&item.PaisCodigo, &item.Visitas); err != nil {
			http.Error(w, `{"ok":false,"error":"no se pudieron leer las visitas"}`, http.StatusInternalServerError)
			return
		}
		item.PaisCodigo = normalizeCountryCode(item.PaisCodigo)
		item.PaisNombre = portalCountryName(item.PaisCodigo)
		resp.TotalVisitas += item.Visitas
		resp.Paises = append(resp.Paises, item)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, `{"ok":false,"error":"no se pudieron recorrer las visitas"}`, http.StatusInternalServerError)
		return
	}
	encodeJSONResponse(w, resp)
}

func portalCountryName(code string) string {
	switch normalizeCountryCode(code) {
	case "CO":
		return "Colombia"
	case "PA":
		return "Panama"
	case "EC":
		return "Ecuador"
	case "US":
		return "Estados Unidos"
	case "MX":
		return "Mexico"
	case "PE":
		return "Peru"
	case "CL":
		return "Chile"
	case "AR":
		return "Argentina"
	case "BR":
		return "Brasil"
	case "VE":
		return "Venezuela"
	case "ES":
		return "Espana"
	case "CA":
		return "Canada"
	case "DO":
		return "Republica Dominicana"
	case "CR":
		return "Costa Rica"
	default:
		if code == "" {
			return "Pais no detectado"
		}
		return code
	}
}
