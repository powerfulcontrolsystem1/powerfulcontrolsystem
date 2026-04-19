package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/you/pos-backend/db"
)

func EmpresaRustDeskDevicesHandler(dbEmp *sql.DB, dbSup *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			devices, err := db.GetRustDeskDevices(dbEmp, empresaID)
			if err != nil {
				log.Printf("Error listando devices rustdesk para la empresa %d: %v", empresaID, err)
				http.Error(w, "Error listando dispositivos", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, devices)

		case http.MethodPost:
			var req db.RustDeskDevice
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			req.EmpresaID = empresaID
			req.Estado = "activo"

			if err := db.RegisterRustDeskDevice(dbEmp, req); err != nil {
				log.Printf("Error registrando device rustdesk empresa %d: %v", empresaID, err)
				http.Error(w, "Error en registro", http.StatusInternalServerError)
				return
			}

			writeJSON(w, http.StatusOK, map[string]string{"message": "Dispositivo configurado en el servidor RustDesk."})

		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	}
}
