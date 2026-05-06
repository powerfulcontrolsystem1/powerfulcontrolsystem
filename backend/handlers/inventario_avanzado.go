package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaInventarioAvanzadoHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "" {
				action = "dashboard"
			}
			productoID, _ := parseInt64QueryOptional(r, "producto_id")
			bodegaID, _ := parseInt64QueryOptional(r, "bodega_id")
			switch action {
			case "dashboard":
				row, err := dbpkg.BuildEmpresaInventarioAvanzadoDashboard(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo cargar inventario avanzado", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
			case "lotes":
				rows, err := dbpkg.ListEmpresaInventarioLotesAvanzados(dbEmp, empresaID, productoID, bodegaID, r.URL.Query().Get("estado"), 200)
				if err != nil {
					http.Error(w, "No se pudieron listar lotes", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
			case "seriales":
				rows, err := dbpkg.ListEmpresaInventarioSerialesAvanzados(dbEmp, empresaID, productoID, bodegaID, r.URL.Query().Get("estado"), 200)
				if err != nil {
					http.Error(w, "No se pudieron listar seriales", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
			case "reservas":
				rows, err := dbpkg.ListEmpresaInventarioReservasAvanzadas(dbEmp, empresaID, r.URL.Query().Get("estado"), 200)
				if err != nil {
					http.Error(w, "No se pudieron listar reservas", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
			case "valorizacion":
				rows, err := dbpkg.ListEmpresaInventarioValorizacionAvanzada(dbEmp, empresaID, 200)
				if err != nil {
					http.Error(w, "No se pudo calcular valorizacion", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
			default:
				http.Error(w, "action no soportada", http.StatusBadRequest)
			}
		case http.MethodPost:
			var payload struct {
				Action    string                                 `json:"action"`
				EmpresaID int64                                  `json:"empresa_id"`
				Lote      dbpkg.EmpresaInventarioLoteAvanzado    `json:"lote"`
				Serial    dbpkg.EmpresaInventarioSerialAvanzado  `json:"serial"`
				Reserva   dbpkg.EmpresaInventarioReservaAvanzada `json:"reserva"`
				ReservaID int64                                  `json:"reserva_id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload invalido", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 {
				payload.EmpresaID, _ = parseInt64QueryOptional(r, "empresa_id")
			}
			if payload.EmpresaID <= 0 {
				http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
				return
			}
			action := strings.ToLower(strings.TrimSpace(payload.Action))
			if action == "" {
				action = strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			}
			usuario := strings.TrimSpace(adminEmailFromRequest(r))
			switch action {
			case "seed_demo":
				id, err := dbpkg.SeedEmpresaInventarioAvanzadoDemo(dbEmp, payload.EmpresaID, usuario)
				if err != nil {
					http.Error(w, "No se pudo crear demo de inventario avanzado", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			case "lote":
				payload.Lote.EmpresaID = payload.EmpresaID
				if payload.Lote.UsuarioCreador == "" {
					payload.Lote.UsuarioCreador = usuario
				}
				id, err := dbpkg.CreateEmpresaInventarioLoteAvanzado(dbEmp, payload.Lote)
				if err != nil {
					http.Error(w, "No se pudo guardar lote: "+err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			case "serial":
				payload.Serial.EmpresaID = payload.EmpresaID
				if payload.Serial.UsuarioCreador == "" {
					payload.Serial.UsuarioCreador = usuario
				}
				id, err := dbpkg.CreateEmpresaInventarioSerialAvanzado(dbEmp, payload.Serial)
				if err != nil {
					http.Error(w, "No se pudo guardar serial: "+err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			case "reserva":
				payload.Reserva.EmpresaID = payload.EmpresaID
				if payload.Reserva.UsuarioCreador == "" {
					payload.Reserva.UsuarioCreador = usuario
				}
				id, err := dbpkg.CreateEmpresaInventarioReservaAvanzada(dbEmp, payload.Reserva)
				if errors.Is(err, dbpkg.ErrStockInsuficiente) {
					http.Error(w, "Stock insuficiente para reservar", http.StatusConflict)
					return
				}
				if err != nil {
					http.Error(w, "No se pudo crear reserva: "+err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			case "confirmar_reserva":
				if payload.ReservaID <= 0 {
					payload.ReservaID, _ = parseInt64QueryOptional(r, "reserva_id")
				}
				if payload.ReservaID <= 0 {
					http.Error(w, "reserva_id es obligatorio", http.StatusBadRequest)
					return
				}
				err := dbpkg.ConfirmarEmpresaInventarioReservaAvanzada(dbEmp, payload.EmpresaID, payload.ReservaID, usuario)
				if errors.Is(err, dbpkg.ErrStockInsuficiente) {
					http.Error(w, "Stock insuficiente para confirmar", http.StatusConflict)
					return
				}
				if err != nil {
					http.Error(w, "No se pudo confirmar reserva: "+err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			default:
				http.Error(w, "action no soportada", http.StatusBadRequest)
			}
		default:
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		}
	}
}
