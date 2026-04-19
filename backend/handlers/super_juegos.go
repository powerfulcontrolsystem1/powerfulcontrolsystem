package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

// PublicJuegosRecordsHandler maneja GET y POST para el ranking global de los juegos (Buscaminas, Solitario, Pacman).
func PublicJuegosRecordsHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet {
			// Consultar el top 10 de un juego específico
			juego := strings.TrimSpace(r.URL.Query().Get("juego"))
			if juego == "" {
				http.Error(w, `{"error": "se requiere parámetro 'juego' (ej. buscaminas, solitario, pacman)"}`, http.StatusBadRequest)
				return
			}

			limitStr := strings.TrimSpace(r.URL.Query().Get("limit"))
			limit := 10
			if limitStr != "" {
				l, err := strconv.Atoi(limitStr)
				if err == nil && l > 0 && l <= 100 {
					limit = l
				}
			}

			tops, err := dbpkg.GetTopSuperJuegoRecords(dbSuper, juego, limit)
			if err != nil {
				http.Error(w, `{"error": "error al obtener records: `+err.Error()+`"}`, http.StatusInternalServerError)
				return
			}

			if tops == nil {
				tops = []dbpkg.SuperJuegoRecord{}
			}

			json.NewEncoder(w).Encode(tops)
			return
		}

		if r.Method == http.MethodPost {
			var rec dbpkg.SuperJuegoRecord
			if err := json.NewDecoder(r.Body).Decode(&rec); err != nil {
				http.Error(w, `{"error": "payload JSON inválido"}`, http.StatusBadRequest)
				return
			}

			rec.Juego = strings.TrimSpace(rec.Juego)
			rec.NombreJugador = strings.TrimSpace(rec.NombreJugador)
			rec.EmpresaID = strings.TrimSpace(rec.EmpresaID)

			if rec.Juego == "" {
				http.Error(w, `{"error": "el campo 'juego' es obligatorio"}`, http.StatusBadRequest)
				return
			}
			if rec.NombreJugador == "" {
				rec.NombreJugador = "Anónimo"
			}
			if rec.EmpresaID == "" {
				// Intenta extraer el empresa_id del context si existe (ej. middleware), sino 'Publico'
				ctxEmp := r.Context().Value("empresa_id")
				if ctxEmp != nil {
					rec.EmpresaID = fmt.Sprintf("%v", ctxEmp)
				} else {
					rec.EmpresaID = "Publico"
				}
			}

			if rec.Puntaje < 0 {
				rec.Puntaje = 0
			}
			if rec.Nivel < 1 {
				rec.Nivel = 1
			}

			id, err := dbpkg.SaveSuperJuegoRecord(dbSuper, rec)
			if err != nil {
				http.Error(w, `{"error": "error al guardar el record: `+err.Error()+`"}`, http.StatusInternalServerError)
				return
			}

			rec.ID = int(id)
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(rec)
			return
		}

		http.Error(w, `{"error": "método no permitido"}`, http.StatusMethodNotAllowed)
	}
}
