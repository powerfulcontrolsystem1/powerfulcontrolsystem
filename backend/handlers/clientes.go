package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"strings"

	db "github.com/you/pos-backend/db"
)

// EmpresaClientesHandler administra CRUD de clientes por empresa.
func EmpresaClientesHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			switch action {
			case "perfil":
				clienteID, err := parseClienteIDFromQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				perfil, err := db.GetClientePerfilComercialByEmpresa(dbEmp, empresaID, clienteID)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "Cliente no encontrado", http.StatusNotFound)
						return
					}
					log.Printf("[clientes] perfil empresa_id=%d id=%d error: %v", empresaID, clienteID, err)
					http.Error(w, "No se pudo obtener el perfil del cliente", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, perfil)
				return

			case "historial":
				clienteID, err := parseClienteIDFromQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				limit, err := parseIntQueryOptional(r, "limit")
				if err != nil {
					http.Error(w, "limit invalido", http.StatusBadRequest)
					return
				}
				rows, err := db.GetClienteHistorialComprasByEmpresa(dbEmp, empresaID, clienteID, limit)
				if err != nil {
					log.Printf("[clientes] historial empresa_id=%d id=%d error: %v", empresaID, clienteID, err)
					http.Error(w, "No se pudo obtener el historial del cliente", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return

			case "segmentacion", "segmentos":
				includeInactive := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("include_inactive")), "1") ||
					strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("include_inactive")), "true")
				q := strings.TrimSpace(r.URL.Query().Get("q"))
				rows, err := db.GetClientesSegmentacionByEmpresa(dbEmp, empresaID, includeInactive, q)
				if err != nil {
					log.Printf("[clientes] segmentacion empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo obtener la segmentacion de clientes", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			}

			includeInactive := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("include_inactive")), "1") ||
				strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("include_inactive")), "true")
			q := strings.TrimSpace(r.URL.Query().Get("q"))

			items, err := db.GetClientesByEmpresa(dbEmp, empresaID, includeInactive, q)
			if err != nil {
				log.Printf("[clientes] list empresa_id=%d error: %v", empresaID, err)
				http.Error(w, "No se pudieron listar los clientes", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, items)
			return

		case http.MethodPost:
			var payload db.Cliente
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if err := validateClientePayload(payload); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))

			newID, err := db.CreateCliente(dbEmp, payload)
			if err != nil {
				var dupErr *db.ClienteDuplicadoError
				if errors.As(err, &dupErr) {
					log.Printf("[clientes] create duplicado empresa_id=%d doc=%s-%s error: %v", payload.EmpresaID, payload.TipoDocumento, payload.NumeroDocumento, err)
					http.Error(w, dupErr.Error(), http.StatusConflict)
					return
				}
				log.Printf("[clientes] create empresa_id=%d doc=%s-%s error: %v", payload.EmpresaID, payload.TipoDocumento, payload.NumeroDocumento, err)
				http.Error(w, "No se pudo crear el cliente (verifique que el documento no este duplicado)", http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]interface{}{
				"ok": true,
				"id": newID,
			})
			return

		case http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "activar" || action == "desactivar" {
				empresaID, errEmp := parseEmpresaIDQuery(r)
				if errEmp != nil {
					http.Error(w, errEmp.Error(), http.StatusBadRequest)
					return
				}
				id, errID := parseInt64Query(r, "id")
				if errID != nil {
					http.Error(w, errID.Error(), http.StatusBadRequest)
					return
				}
				estado := "activo"
				if action == "desactivar" {
					estado = "inactivo"
				}
				if err := db.SetClienteEstado(dbEmp, empresaID, id, estado); err != nil {
					log.Printf("[clientes] set estado empresa_id=%d id=%d estado=%s error: %v", empresaID, id, estado, err)
					http.Error(w, "No se pudo actualizar el estado del cliente", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": estado})
				return
			}

			var payload db.Cliente
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.ID <= 0 {
				http.Error(w, "id es obligatorio", http.StatusBadRequest)
				return
			}
			if err := validateClientePayload(payload); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := db.UpdateCliente(dbEmp, payload); err != nil {
				var dupErr *db.ClienteDuplicadoError
				if errors.As(err, &dupErr) {
					log.Printf("[clientes] update duplicado empresa_id=%d id=%d error: %v", payload.EmpresaID, payload.ID, err)
					http.Error(w, dupErr.Error(), http.StatusConflict)
					return
				}
				log.Printf("[clientes] update empresa_id=%d id=%d error: %v", payload.EmpresaID, payload.ID, err)
				http.Error(w, "No se pudo actualizar el cliente", http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return

		case http.MethodDelete:
			empresaID, errEmp := parseEmpresaIDQuery(r)
			if errEmp != nil {
				http.Error(w, errEmp.Error(), http.StatusBadRequest)
				return
			}
			id, errID := parseInt64Query(r, "id")
			if errID != nil {
				http.Error(w, errID.Error(), http.StatusBadRequest)
				return
			}
			if err := db.DeleteCliente(dbEmp, empresaID, id); err != nil {
				log.Printf("[clientes] delete empresa_id=%d id=%d error: %v", empresaID, id, err)
				http.Error(w, "No se pudo eliminar el cliente", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

func validateClientePayload(payload db.Cliente) error {
	if payload.EmpresaID <= 0 {
		return errBadRequest("empresa_id es obligatorio")
	}
	if strings.TrimSpace(payload.TipoDocumento) == "" {
		payload.TipoDocumento = "NIT"
	}
	if strings.TrimSpace(payload.NumeroDocumento) == "" {
		return errBadRequest("numero_documento es obligatorio")
	}
	if strings.TrimSpace(payload.NombreRazonSocial) == "" {
		return errBadRequest("nombre_razon_social es obligatorio")
	}
	if email := strings.TrimSpace(payload.Email); email != "" {
		if _, err := mail.ParseAddress(email); err != nil {
			return errBadRequest("email invalido")
		}
	}
	return nil
}

type badRequestError struct{ msg string }

func (e badRequestError) Error() string { return e.msg }

func errBadRequest(msg string) error { return badRequestError{msg: msg} }

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("[clientes] write json response error: %v", err)
	}
}

func parseClienteIDFromQuery(r *http.Request) (int64, error) {
	clienteID, err := parseInt64QueryOptional(r, "cliente_id")
	if err != nil {
		return 0, errBadRequest("cliente_id invalido")
	}
	if clienteID <= 0 {
		clienteID, err = parseInt64QueryOptional(r, "id")
		if err != nil {
			return 0, errBadRequest("id/cliente_id invalido")
		}
	}
	if clienteID <= 0 {
		return 0, errBadRequest("cliente_id es obligatorio")
	}
	return clienteID, nil
}
