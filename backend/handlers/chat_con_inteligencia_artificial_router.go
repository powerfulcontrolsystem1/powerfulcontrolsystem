package handlers

import (
	"database/sql"
	"net/http"
)

// RegisterEmpresaChatIARoutes registra rutas del modulo de chat con IA por empresa.
func RegisterEmpresaChatIARoutes(dbEmp, dbSuper *sql.DB) {
	ctrl := NewEmpresaAIChatController(dbEmp, dbSuper)
	http.HandleFunc("/api/empresa/chat_con_inteligencia_artificial/modelos", WithEmpresaSeguridadPermissions(dbEmp, dbSuper, ctrl.ModelosHandler))
	http.HandleFunc("/api/empresa/chat_con_inteligencia_artificial/modelo_preferido", WithEmpresaSeguridadPermissions(dbEmp, dbSuper, ctrl.ModeloPreferidoHandler))
	http.HandleFunc("/api/empresa/chat_con_inteligencia_artificial/consultar", WithEmpresaSeguridadPermissions(dbEmp, dbSuper, ctrl.ConsultarHandler))
	http.HandleFunc("/api/empresa/chat_con_inteligencia_artificial/consultar_con_adjunto", WithEmpresaSeguridadPermissions(dbEmp, dbSuper, ctrl.ConsultarConAdjuntoHandler))
	http.HandleFunc("/api/empresa/chat_con_inteligencia_artificial/consultar_stream", WithEmpresaSeguridadPermissions(dbEmp, dbSuper, ctrl.ConsultarStreamHandler))
	http.HandleFunc("/api/empresa/chat_con_inteligencia_artificial/historial", WithEmpresaSeguridadPermissions(dbEmp, dbSuper, ctrl.HistorialHandler))
	http.HandleFunc("/api/empresa/chat_documentos/generar", WithEmpresaSeguridadPermissions(dbEmp, dbSuper, DynamicDocumentGenerateHandler(dbEmp, dbSuper)))
	http.HandleFunc("/api/empresa/chat_documentos/exportar", WithEmpresaSeguridadPermissions(dbEmp, dbSuper, DynamicDocumentChatExportHandler(dbEmp, dbSuper)))
	http.HandleFunc("/api/empresa/ia/importar_desde_foto", WithEmpresaSeguridadPermissions(dbEmp, dbSuper, EmpresaIAImportarDesdeFotoHandler(dbEmp)))
	http.HandleFunc("/api/empresa/ia_pedidos_estacion/ejecutar", WithEmpresaVentasPermissions(dbEmp, dbSuper, ctrl.IaPedidosEstacionEjecutarHandler))
}
