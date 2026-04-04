package handlers

import (
	"database/sql"
	"net/http"
)

// RegisterEmpresaChatIARoutes registra rutas del modulo de chat con IA por empresa.
func RegisterEmpresaChatIARoutes(dbEmp, dbSuper *sql.DB) {
	ctrl := NewEmpresaAIChatController(dbEmp, dbSuper)
	http.HandleFunc("/api/empresa/chat_con_inteligencia_artificial/modelos", ctrl.ModelosHandler)
	http.HandleFunc("/api/empresa/chat_con_inteligencia_artificial/modelo_preferido", ctrl.ModeloPreferidoHandler)
	http.HandleFunc("/api/empresa/chat_con_inteligencia_artificial/consultar", ctrl.ConsultarHandler)
	http.HandleFunc("/api/empresa/chat_con_inteligencia_artificial/historial", ctrl.HistorialHandler)
}
