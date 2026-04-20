package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	
	"powerfulcontrolsystem/backend/db"
)

// EmpresaHorariosHandler maneja las peticiones CRUD de Horarios de Trabajadores.
func EmpresaHorariosHandler(w http.ResponseWriter, r *http.Request) {
	empresaIDStr := r.URL.Query().Get("empresa_id")
	empresaID, err := strconv.ParseInt(empresaIDStr, 10, 64)
	if err != nil || empresaID == 0 {
		http.Error(w, "empresa_id inválido o faltante", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		action := r.URL.Query().Get("action")
		if action == "by_user" {
			usuarioIDStr := r.URL.Query().Get("usuario_id")
			usuarioID, err := strconv.ParseInt(usuarioIDStr, 10, 64)
			if err != nil {
				http.Error(w, "usuario_id inválido", http.StatusBadRequest)
				return
			}
			horarios, err := db.GetHorariosTrabajadorByUsuario(empresaID, usuarioID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(horarios)
			return
		}

		// Por defecto, lista un rango de fecha
		fechaInicio := r.URL.Query().Get("fecha_inicio")
		fechaFin := r.URL.Query().Get("fecha_fin")
		if fechaInicio == "" || fechaFin == "" {
			http.Error(w, "Faltan parámetros fecha_inicio y/o fecha_fin (YYYY-MM-DD)", http.StatusBadRequest)
			return
		}

		horarios, err := db.GetHorariosTrabajadoresByRango(empresaID, fechaInicio, fechaFin)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(horarios)

	case http.MethodPost:
		var input db.HorarioTrabajador
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Payload json inválido: "+err.Error(), http.StatusBadRequest)
			return
		}
		input.EmpresaID = empresaID
		
		// FIXME: Extraer sessionToken y buscar creador, por ahora default o de session
		// Get creador del contexto si estuviese (el middleware lo hace).
		input.UsuarioCreador = "Sistema/Admin" // Hardcoded o lo traería del auth 
		
		if err := db.CreateHorarioTrabajador(&input); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})

	case http.MethodPut:
		var input db.HorarioTrabajador
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Payload json inválido: "+err.Error(), http.StatusBadRequest)
			return
		}
		input.EmpresaID = empresaID

		if err := db.UpdateHorarioTrabajador(&input); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})

	case http.MethodDelete:
		idStr := r.URL.Query().Get("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "id inválido", http.StatusBadRequest)
			return
		}
		if err := db.DeleteHorarioTrabajador(id, empresaID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})

	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}
