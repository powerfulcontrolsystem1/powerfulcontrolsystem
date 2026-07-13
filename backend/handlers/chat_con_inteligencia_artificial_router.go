package handlers

import (
	"database/sql"
	"net/http"
)

// RegisterEmpresaChatIARoutes registra rutas del modulo de chat con IA por empresa.
func RegisterEmpresaChatIARoutes(dbEmp, dbSuper *sql.DB) {
	ctrl := NewEmpresaAIChatController(dbEmp, dbSuper)
	http.HandleFunc("/api/empresa/chat_con_inteligencia_artificial/modelos", WithEmpresaVentasPermissions(dbEmp, dbSuper, ctrl.ModelosHandler))
	http.HandleFunc("/api/empresa/chat_con_inteligencia_artificial/modelo_preferido", WithEmpresaVentasPermissions(dbEmp, dbSuper, ctrl.ModeloPreferidoHandler))
	http.HandleFunc("/api/empresa/chat_con_inteligencia_artificial/consultar", WithEmpresaVentasPermissions(dbEmp, dbSuper, ctrl.ConsultarHandler))
	http.HandleFunc("/api/empresa/chat_con_inteligencia_artificial/consultar_con_adjunto", WithEmpresaVentasPermissions(dbEmp, dbSuper, ctrl.ConsultarConAdjuntoHandler))
	http.HandleFunc("/api/empresa/chat_con_inteligencia_artificial/consultar_stream", WithEmpresaVentasPermissions(dbEmp, dbSuper, ctrl.ConsultarStreamHandler))
	http.HandleFunc("/api/empresa/chat_con_inteligencia_artificial/historial", WithEmpresaVentasPermissions(dbEmp, dbSuper, ctrl.HistorialHandler))
	http.HandleFunc("/api/empresa/chat_documentos/generar", WithEmpresaSeguridadPermissions(dbEmp, dbSuper, DynamicDocumentGenerateHandler(dbEmp, dbSuper)))
	http.HandleFunc("/api/empresa/chat_documentos/exportar", WithEmpresaSeguridadPermissions(dbEmp, dbSuper, DynamicDocumentChatExportHandler(dbEmp, dbSuper)))
	http.HandleFunc("/api/empresa/chat_documentos/compartir_email", WithEmpresaSeguridadPermissions(dbEmp, dbSuper, DynamicDocumentEmailShareHandler(dbEmp, dbSuper)))
	http.HandleFunc("/api/empresa/ia/importar_desde_foto", WithEmpresaSeguridadPermissions(dbEmp, dbSuper, EmpresaIAImportarDesdeFotoHandler(dbEmp)))
	http.HandleFunc("/api/empresa/ia_pedidos_estacion/ejecutar", WithEmpresaVentasPermissions(dbEmp, dbSuper, ctrl.IaPedidosEstacionEjecutarHandler))
	http.HandleFunc("/api/empresa/ia_radio/activar", WithEmpresaVentasPermissions(dbEmp, dbSuper, EmpresaIARadioHandler(dbSuper, dbEmp)))
	// El orquestador empresarial no acepta endpoints ni acciones elegidas por el modelo.
	// This wrapper establishes only the authenticated company scope. Every tool
	// then validates its own module/action against the same server snapshot.
	http.HandleFunc("/api/empresa/ai/enterprise", WithEmpresaAIEnterprisePermissions(dbEmp, dbSuper, EmpresaAIEnterpriseHandler(dbEmp, dbSuper)))
}
