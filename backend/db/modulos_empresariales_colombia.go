package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

type EmpresaModuloColombiaRegistro struct {
	ID                 int64                  `json:"id"`
	EmpresaID          int64                  `json:"empresa_id"`
	Modulo             string                 `json:"modulo"`
	Tipo               string                 `json:"tipo"`
	Codigo             string                 `json:"codigo"`
	Nombre             string                 `json:"nombre"`
	Tercero            string                 `json:"tercero,omitempty"`
	Responsable        string                 `json:"responsable,omitempty"`
	Categoria          string                 `json:"categoria,omitempty"`
	Referencia         string                 `json:"referencia,omitempty"`
	Prioridad          string                 `json:"prioridad"`
	Estado             string                 `json:"estado"`
	Fecha              string                 `json:"fecha,omitempty"`
	FechaVencimiento   string                 `json:"fecha_vencimiento,omitempty"`
	Valor              float64                `json:"valor"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	MetadataJSON       string                 `json:"-"`
	UsuarioCreador     string                 `json:"usuario_creador,omitempty"`
	FechaCreacion      string                 `json:"fecha_creacion,omitempty"`
	FechaActualizacion string                 `json:"fecha_actualizacion,omitempty"`
}

type EmpresaModuloColombiaEvento struct {
	ID             int64  `json:"id"`
	EmpresaID      int64  `json:"empresa_id"`
	Modulo         string `json:"modulo"`
	RegistroID     int64  `json:"registro_id"`
	Evento         string `json:"evento"`
	EstadoAnterior string `json:"estado_anterior,omitempty"`
	EstadoNuevo    string `json:"estado_nuevo,omitempty"`
	Detalle        string `json:"detalle,omitempty"`
	Usuario        string `json:"usuario,omitempty"`
	FechaCreacion  string `json:"fecha_creacion,omitempty"`
}

type EmpresaModuloColombiaEvidencia struct {
	ID            int64  `json:"id"`
	EmpresaID     int64  `json:"empresa_id"`
	Modulo        string `json:"modulo"`
	RegistroID    int64  `json:"registro_id"`
	Tipo          string `json:"tipo"`
	Nombre        string `json:"nombre"`
	URL           string `json:"url,omitempty"`
	Descripcion   string `json:"descripcion,omitempty"`
	Usuario       string `json:"usuario,omitempty"`
	FechaCreacion string `json:"fecha_creacion,omitempty"`
}

type EmpresaModuloColombiaAprobacion struct {
	ID               int64  `json:"id"`
	EmpresaID        int64  `json:"empresa_id"`
	Modulo           string `json:"modulo"`
	RegistroID       int64  `json:"registro_id"`
	Nivel            string `json:"nivel"`
	SolicitadoA      string `json:"solicitado_a"`
	SolicitadoPor    string `json:"solicitado_por,omitempty"`
	Estado           string `json:"estado"`
	Comentario       string `json:"comentario,omitempty"`
	DecisionPor      string `json:"decision_por,omitempty"`
	FechaDecision    string `json:"fecha_decision,omitempty"`
	FechaCreacion    string `json:"fecha_creacion,omitempty"`
	FechaVencimiento string `json:"fecha_vencimiento,omitempty"`
}

type EmpresaModuloColombiaTarea struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Modulo             string `json:"modulo"`
	RegistroID         int64  `json:"registro_id"`
	Titulo             string `json:"titulo"`
	Responsable        string `json:"responsable,omitempty"`
	Prioridad          string `json:"prioridad"`
	Estado             string `json:"estado"`
	FechaVencimiento   string `json:"fecha_vencimiento,omitempty"`
	Comentario         string `json:"comentario,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
}

type EmpresaModuloColombiaDashboard struct {
	EmpresaID        int64                           `json:"empresa_id"`
	Modulo           string                          `json:"modulo"`
	Titulo           string                          `json:"titulo"`
	TotalRegistros   int                             `json:"total_registros"`
	Abiertos         int                             `json:"abiertos"`
	EnProceso        int                             `json:"en_proceso"`
	Aprobados        int                             `json:"aprobados"`
	Vencidos         int                             `json:"vencidos"`
	ValorTotal       float64                         `json:"valor_total"`
	Registros        []EmpresaModuloColombiaRegistro `json:"registros"`
	EventosRecientes []EmpresaModuloColombiaEvento   `json:"eventos_recientes"`
	Alertas          []string                        `json:"alertas"`
}

type EmpresaModuloColombiaMetrica struct {
	Clave string  `json:"clave"`
	Total int     `json:"total"`
	Valor float64 `json:"valor"`
}

type EmpresaModuloColombiaReporte struct {
	EmpresaID        int64                          `json:"empresa_id"`
	Modulo           string                         `json:"modulo"`
	Titulo           string                         `json:"titulo"`
	PorEstado        []EmpresaModuloColombiaMetrica `json:"por_estado"`
	PorTipo          []EmpresaModuloColombiaMetrica `json:"por_tipo"`
	PorCategoria     []EmpresaModuloColombiaMetrica `json:"por_categoria"`
	PorPrioridad     []EmpresaModuloColombiaMetrica `json:"por_prioridad"`
	Vencidos         int                            `json:"vencidos"`
	Vencen7Dias      int                            `json:"vencen_7_dias"`
	Vencen30Dias     int                            `json:"vencen_30_dias"`
	CriticosAbiertos int                            `json:"criticos_abiertos"`
	SinResponsable   int                            `json:"sin_responsable"`
	ValorPendiente   float64                        `json:"valor_pendiente"`
	ValorVencido     float64                        `json:"valor_vencido"`
	UltimaActividad  string                         `json:"ultima_actividad,omitempty"`
	Recomendaciones  []string                       `json:"recomendaciones"`
}

type EmpresaModuloColombiaAgendaItem struct {
	Tipo             string `json:"tipo"`
	RegistroID       int64  `json:"registro_id"`
	ReferenciaID     int64  `json:"referencia_id,omitempty"`
	Codigo           string `json:"codigo,omitempty"`
	Titulo           string `json:"titulo"`
	Responsable      string `json:"responsable,omitempty"`
	Estado           string `json:"estado"`
	Prioridad        string `json:"prioridad,omitempty"`
	FechaVencimiento string `json:"fecha_vencimiento,omitempty"`
	Severidad        string `json:"severidad"`
	Detalle          string `json:"detalle,omitempty"`
}

type EmpresaModuloColombiaAgenda struct {
	EmpresaID              int64                             `json:"empresa_id"`
	Modulo                 string                            `json:"modulo"`
	FechaCorte             string                            `json:"fecha_corte"`
	RegistrosVencidos      int                               `json:"registros_vencidos"`
	RegistrosProximos      int                               `json:"registros_proximos"`
	TareasVencidas         int                               `json:"tareas_vencidas"`
	TareasProximas         int                               `json:"tareas_proximas"`
	AprobacionesVencidas   int                               `json:"aprobaciones_vencidas"`
	AprobacionesPendientes int                               `json:"aprobaciones_pendientes"`
	TotalAlertas           int                               `json:"total_alertas"`
	Items                  []EmpresaModuloColombiaAgendaItem `json:"items"`
	Recomendaciones        []string                          `json:"recomendaciones"`
}

type EmpresaModuloColombiaResponsableResumen struct {
	Responsable            string `json:"responsable"`
	RegistrosAbiertos      int    `json:"registros_abiertos"`
	RegistrosVencidos      int    `json:"registros_vencidos"`
	TareasAbiertas         int    `json:"tareas_abiertas"`
	TareasVencidas         int    `json:"tareas_vencidas"`
	AprobacionesPendientes int    `json:"aprobaciones_pendientes"`
	TotalPendiente         int    `json:"total_pendiente"`
	Recomendacion          string `json:"recomendacion"`
}

type EmpresaModuloColombiaSLA struct {
	EmpresaID       int64          `json:"empresa_id"`
	Modulo          string         `json:"modulo"`
	FechaCorte      string         `json:"fecha_corte"`
	TotalAbiertos   int            `json:"total_abiertos"`
	Vencidos        int            `json:"vencidos"`
	Proximos7Dias   int            `json:"proximos_7_dias"`
	SinVencimiento  int            `json:"sin_vencimiento"`
	TareasAbiertas  int            `json:"tareas_abiertas"`
	TareasVencidas  int            `json:"tareas_vencidas"`
	CumplimientoPct float64        `json:"cumplimiento_pct"`
	Semaforo        string         `json:"semaforo"`
	Buckets         map[string]int `json:"buckets"`
	Recomendaciones []string       `json:"recomendaciones"`
}

type EmpresaModuloColombiaRiesgo struct {
	EmpresaID              int64    `json:"empresa_id"`
	Modulo                 string   `json:"modulo"`
	Score                  int      `json:"score"`
	Nivel                  string   `json:"nivel"`
	RegistrosVencidos      int      `json:"registros_vencidos"`
	CriticosAbiertos       int      `json:"criticos_abiertos"`
	SinResponsable         int      `json:"sin_responsable"`
	SinEvidencia           int      `json:"sin_evidencia"`
	AprobacionesPendientes int      `json:"aprobaciones_pendientes"`
	TareasAbiertas         int      `json:"tareas_abiertas"`
	TareasVencidas         int      `json:"tareas_vencidas"`
	Factores               []string `json:"factores"`
	Recomendaciones        []string `json:"recomendaciones"`
}

type EmpresaModuloColombiaExportacion struct {
	EmpresaID  int64                             `json:"empresa_id"`
	Modulo     string                            `json:"modulo"`
	Titulo     string                            `json:"titulo"`
	FechaCorte string                            `json:"fecha_corte"`
	Secciones  []EmpresaModuloColombiaCSVSeccion `json:"secciones"`
}

type EmpresaModuloColombiaCSVSeccion struct {
	Nombre  string     `json:"nombre"`
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}

type EmpresaModuloColombiaBusqueda struct {
	EmpresaID  int64                           `json:"empresa_id"`
	Modulo     string                          `json:"modulo"`
	Total      int                             `json:"total"`
	Filtros    EmpresaModuloColombiaFiltro     `json:"filtros"`
	Registros  []EmpresaModuloColombiaRegistro `json:"registros"`
	FechaCorte string                          `json:"fecha_corte"`
}

type EmpresaModuloColombiaFiltro struct {
	Texto        string `json:"texto,omitempty"`
	Estado       string `json:"estado,omitempty"`
	Tipo         string `json:"tipo,omitempty"`
	Categoria    string `json:"categoria,omitempty"`
	Prioridad    string `json:"prioridad,omitempty"`
	Responsable  string `json:"responsable,omitempty"`
	Vencidos     bool   `json:"vencidos,omitempty"`
	ProximosDias int    `json:"proximos_dias,omitempty"`
}

type EmpresaModuloColombiaExpediente struct {
	EmpresaID     int64                             `json:"empresa_id"`
	Modulo        string                            `json:"modulo"`
	Registro      EmpresaModuloColombiaRegistro     `json:"registro"`
	Eventos       []EmpresaModuloColombiaEvento     `json:"eventos"`
	Evidencias    []EmpresaModuloColombiaEvidencia  `json:"evidencias"`
	Aprobaciones  []EmpresaModuloColombiaAprobacion `json:"aprobaciones"`
	Tareas        []EmpresaModuloColombiaTarea      `json:"tareas"`
	Resumen       map[string]int                    `json:"resumen"`
	Recomendacion string                            `json:"recomendacion"`
}

type EmpresaModuloColombiaImportResult struct {
	Total     int      `json:"total"`
	Guardados int      `json:"guardados"`
	Errores   []string `json:"errores"`
}

type EmpresaModuloColombiaPlanAccionResult struct {
	AlertasRevisadas int      `json:"alertas_revisadas"`
	TareasCreadas    int      `json:"tareas_creadas"`
	Omitidas         int      `json:"omitidas"`
	Detalles         []string `json:"detalles"`
}

type EmpresaModuloColombiaAccionMasiva struct {
	RegistroIDs []int64 `json:"registro_ids"`
	Estado      string  `json:"estado,omitempty"`
	Prioridad   string  `json:"prioridad,omitempty"`
	Responsable string  `json:"responsable,omitempty"`
	Detalle     string  `json:"detalle,omitempty"`
}

type EmpresaModuloColombiaAccionMasivaResult struct {
	Total        int      `json:"total"`
	Actualizados int      `json:"actualizados"`
	Omitidos     int      `json:"omitidos"`
	Detalles     []string `json:"detalles"`
}

type EmpresaModuloColombiaPlantilla struct {
	Modulo             string   `json:"modulo"`
	Titulo             string   `json:"titulo"`
	SeccionesFlujo     []string `json:"secciones_flujo,omitempty"`
	Tipos              []string `json:"tipos"`
	Categorias         []string `json:"categorias"`
	EstadosFlujo       []string `json:"estados_flujo"`
	AccionesSugeridas  []string `json:"acciones_sugeridas"`
	EtiquetaTercero    string   `json:"etiqueta_tercero"`
	EtiquetaReferencia string   `json:"etiqueta_referencia"`
	MetadataEjemplo    string   `json:"metadata_ejemplo"`
}

type EmpresaModuloColombiaDiagnosticoCheck struct {
	Clave         string `json:"clave"`
	Titulo        string `json:"titulo"`
	OK            bool   `json:"ok"`
	Informativo   bool   `json:"informativo,omitempty"`
	Detalle       string `json:"detalle,omitempty"`
	Recomendacion string `json:"recomendacion,omitempty"`
}

type EmpresaModuloColombiaDiagnostico struct {
	EmpresaID         int64                                   `json:"empresa_id"`
	Modulo            string                                  `json:"modulo"`
	Titulo            string                                  `json:"titulo"`
	Estado            string                                  `json:"estado"`
	Puntuacion        int                                     `json:"puntuacion"`
	TotalObligatorios int                                     `json:"total_obligatorios"`
	OKObligatorios    int                                     `json:"ok_obligatorios"`
	Checks            []EmpresaModuloColombiaDiagnosticoCheck `json:"checks"`
	Recomendaciones   []string                                `json:"recomendaciones,omitempty"`
}

var empresaModuloColombiaTitulos = map[string]string{
	"bancos_pagos":           "Bancos y pagos masivos Colombia",
	"gestion_documental":     "Gestion documental empresarial",
	"cumplimiento_kyc":       "Cumplimiento KYC/KYB y riesgo LAFT",
	"contratos_obligaciones": "Contratos, obligaciones y firma electronica",
	"calidad_procesos":       "Calidad, procesos y no conformidades",
	"drogueria_farmacia":     "Drogueria y farmacia",
}

func GetEmpresaModuloColombiaPlantilla(modulo string) EmpresaModuloColombiaPlantilla {
	modulo = NormalizeEmpresaModuloColombia(modulo)
	base := EmpresaModuloColombiaPlantilla{
		Modulo:             modulo,
		Titulo:             empresaModuloColombiaTitulos[modulo],
		SeccionesFlujo:     GetEmpresaModuloColombiaSeccionesFlujo(modulo),
		EstadosFlujo:       []string{"pendiente", "en_revision", "en_proceso", "aprobado", "cerrado", "rechazado", "cancelado"},
		AccionesSugeridas:  []string{"seguimiento", "comentario", "aprobacion", "evidencia", "cierre"},
		EtiquetaTercero:    "Tercero / area",
		EtiquetaReferencia: "Referencia",
		MetadataEjemplo:    `{"nota":"detalle operativo"}`,
	}
	if plantilla, ok := empresaModuloColombiaPlantillasPlantillas[modulo]; ok {
		base.Titulo = plantilla.Titulo
		base.SeccionesFlujo = append([]string{}, GetEmpresaModuloColombiaSeccionesFlujo(modulo)...)
		base.Tipos = append([]string{}, plantilla.Tipos...)
		base.Categorias = append([]string{}, plantilla.Categorias...)
		base.EstadosFlujo = append([]string{}, plantilla.EstadosFlujo...)
		base.AccionesSugeridas = append([]string{}, plantilla.AccionesSugeridas...)
		base.EtiquetaTercero = plantilla.EtiquetaTercero
		base.EtiquetaReferencia = plantilla.EtiquetaReferencia
		base.MetadataEjemplo = plantilla.MetadataEjemplo
		return base
	}
	switch modulo {
	case "bancos_pagos":
		base.Tipos = []string{"conciliacion", "extracto", "pago_masivo", "rechazo_bancario", "aprobacion_pago"}
		base.Categorias = []string{"banco", "proveedores", "nomina", "cartera", "comisiones", "impuestos"}
		base.EstadosFlujo = []string{"pendiente", "en_revision", "aprobado", "pagado", "rechazado", "cancelado"}
		base.EtiquetaTercero = "Banco / beneficiario"
		base.EtiquetaReferencia = "Cuenta / lote / extracto"
		base.MetadataEjemplo = `{"banco":"Bancolombia","cuenta":"Ahorros 1234","archivo":"CSV/OFX","aprobadores":2}`
	case "gestion_documental":
		base.Tipos = []string{"radicado", "expediente", "version", "aprobacion", "retencion", "vencimiento"}
		base.Categorias = []string{"compras", "contratos", "laboral", "tributario", "operacion", "legal"}
		base.EtiquetaTercero = "Origen / tercero"
		base.EtiquetaReferencia = "Expediente / radicado"
		base.MetadataEjemplo = `{"version":"v1","retencion":"7 anos","etiquetas":"legal,compras","flujo":"area->gerencia"}`
	case "cumplimiento_kyc":
		base.Tipos = []string{"evaluacion", "alerta", "revision_listas", "beneficiario_final", "aprobacion_riesgo"}
		base.Categorias = []string{"cliente", "proveedor", "empleado", "contratista", "LAFT", "PEP"}
		base.EstadosFlujo = []string{"pendiente", "en_revision", "aprobado", "rechazado", "cerrado"}
		base.EtiquetaTercero = "Tercero evaluado"
		base.EtiquetaReferencia = "NIT / documento / caso"
		base.MetadataEjemplo = `{"riesgo":"medio","listas":"pendiente","beneficiario_final":true,"senal_alerta":"valor inusual"}`
	case "contratos_obligaciones":
		base.Tipos = []string{"contrato", "obligacion", "poliza", "hito", "renovacion", "firma"}
		base.Categorias = []string{"servicios", "arrendamiento", "laboral", "proveedor", "mantenimiento", "obra"}
		base.EtiquetaTercero = "Contraparte"
		base.EtiquetaReferencia = "Contrato / poliza"
		base.MetadataEjemplo = `{"firma":"pendiente","renovacion":"2026-12-31","alerta_dias":30,"valor_mensual":800000}`
	case "calidad_procesos":
		base.Tipos = []string{"proceso", "checklist", "auditoria", "no_conformidad", "accion_correctiva", "accion_preventiva"}
		base.Categorias = []string{"servicio", "finanzas", "inventario", "seguridad", "operacion", "cliente"}
		base.EstadosFlujo = []string{"pendiente", "en_proceso", "en_revision", "cumplido", "cerrado", "rechazado"}
		base.EtiquetaTercero = "Proceso / area"
		base.EtiquetaReferencia = "Checklist / auditoria"
		base.MetadataEjemplo = `{"hallazgos":2,"causa_raiz":"pendiente","accion_correctiva":"definir","responsable_cierre":"Supervisor"}`
	case "drogueria_farmacia":
		base.Tipos = []string{"medicamento", "lote", "formula_medica", "controlado", "vencimiento", "devolucion", "farmacovigilancia", "inventario_sanitario"}
		base.Categorias = []string{"otc", "rx", "controlados", "dispositivos_medicos", "aseo_salud", "cadena_frio", "compras", "dispensacion"}
		base.EstadosFlujo = []string{"pendiente", "en_revision", "aprobado", "dispensado", "observado", "bloqueado", "cerrado", "rechazado"}
		base.AccionesSugeridas = []string{"dispensacion", "validacion_formula", "bloqueo_lote", "farmacovigilancia", "conteo", "devolucion", "cierre"}
		base.EtiquetaTercero = "Paciente / proveedor / laboratorio"
		base.EtiquetaReferencia = "INVIMA / lote / formula"
		base.MetadataEjemplo = `{"registro_invima":"INVIMA 2026M-000000","lote":"L-001","vence":"2026-12-31","requiere_formula":true,"controlado":false,"cadena_frio":false}`
	default:
		base.Tipos = []string{"general", "seguimiento", "aprobacion"}
		base.Categorias = []string{"general", "operacion", "finanzas"}
	}
	return base
}

func BuildEmpresaModuloColombiaDiagnostico(dbConn *sql.DB, empresaID int64, modulo string) (EmpresaModuloColombiaDiagnostico, error) {
	modulo = NormalizeEmpresaModuloColombia(modulo)
	if modulo == "" {
		return EmpresaModuloColombiaDiagnostico{}, errors.New("modulo no soportado")
	}
	plantilla := GetEmpresaModuloColombiaPlantilla(modulo)
	totalRegistros := 0
	dbDetalle := "Conexion operativa"
	dbOK := true
	if dbConn != nil {
		registros, err := ListEmpresaModuloColombiaRegistros(dbConn, empresaID, modulo, "", 1)
		if err != nil {
			dbOK = false
			dbDetalle = err.Error()
		} else {
			totalRegistros = len(registros)
		}
	}
	return buildEmpresaModuloColombiaDiagnostico(empresaID, modulo, plantilla, totalRegistros, dbOK, dbDetalle), nil
}

func buildEmpresaModuloColombiaDiagnostico(empresaID int64, modulo string, plantilla EmpresaModuloColombiaPlantilla, totalRegistros int, dbOK bool, dbDetalle string) EmpresaModuloColombiaDiagnostico {
	metadataOK := false
	if strings.TrimSpace(plantilla.MetadataEjemplo) != "" {
		var meta map[string]interface{}
		metadataOK = json.Unmarshal([]byte(plantilla.MetadataEjemplo), &meta) == nil
	}
	if strings.TrimSpace(dbDetalle) == "" {
		dbDetalle = "Conexion operativa"
	}
	checks := []EmpresaModuloColombiaDiagnosticoCheck{
		{Clave: "empresa_contexto", Titulo: "Empresa detectada", OK: empresaID > 0, Detalle: fmt.Sprintf("empresa_id %d", empresaID), Recomendacion: "Abrir el modulo desde el panel de una empresa activa."},
		{Clave: "modulo_soportado", Titulo: "Modulo soportado", OK: modulo != "" && plantilla.Modulo == modulo && plantilla.Titulo != "", Detalle: plantilla.Titulo, Recomendacion: "Registrar el modulo en el catalogo empresarial antes de activarlo."},
		{Clave: "base_datos", Titulo: "Base de datos operativa", OK: dbOK, Detalle: dbDetalle, Recomendacion: "Revisar migraciones y conexion de la base empresarial."},
		{Clave: "ruta_trabajo", Titulo: "Ruta de trabajo", OK: len(plantilla.SeccionesFlujo) >= 4, Detalle: fmt.Sprintf("%d secciones", len(plantilla.SeccionesFlujo)), Recomendacion: "Definir las secciones principales del submenu operativo."},
		{Clave: "tipos_categorias", Titulo: "Tipos y categorias", OK: len(plantilla.Tipos) >= 2 && len(plantilla.Categorias) >= 2, Detalle: fmt.Sprintf("%d tipos / %d categorias", len(plantilla.Tipos), len(plantilla.Categorias)), Recomendacion: "Definir al menos dos tipos y dos categorias para clasificar registros."},
		{Clave: "estados_acciones", Titulo: "Estados y acciones", OK: len(plantilla.EstadosFlujo) >= 3 && len(plantilla.AccionesSugeridas) >= 3, Detalle: fmt.Sprintf("%d estados / %d acciones", len(plantilla.EstadosFlujo), len(plantilla.AccionesSugeridas)), Recomendacion: "Completar el flujo minimo de estados y acciones sugeridas."},
		{Clave: "etiquetas", Titulo: "Etiquetas operativas", OK: plantilla.EtiquetaTercero != "" && plantilla.EtiquetaReferencia != "", Detalle: strings.TrimSpace(plantilla.EtiquetaTercero + " / " + plantilla.EtiquetaReferencia), Recomendacion: "Configurar las etiquetas de tercero y referencia para el negocio."},
		{Clave: "metadata_json", Titulo: "Metadata JSON", OK: metadataOK, Detalle: map[bool]string{true: "Ejemplo valido", false: "Revisar JSON de ejemplo"}[metadataOK], Recomendacion: "Corregir metadata_ejemplo para que sea JSON valido."},
		{Clave: "registros_operativos", Titulo: "Registros operativos", OK: totalRegistros > 0, Informativo: true, Detalle: map[bool]string{true: fmt.Sprintf("%d registro(s) recientes", totalRegistros), false: "Listo para cargar el primer registro"}[totalRegistros > 0]},
	}
	recomendaciones := []string{}
	okObligatorios := 0
	totalObligatorios := 0
	for _, check := range checks {
		if check.Informativo {
			continue
		}
		totalObligatorios++
		if check.OK {
			okObligatorios++
			continue
		}
		if check.Recomendacion != "" {
			recomendaciones = append(recomendaciones, check.Recomendacion)
		}
	}
	if len(recomendaciones) == 0 && totalRegistros == 0 {
		recomendaciones = append(recomendaciones, "Cargar registros demo o crear el primer registro para activar metricas reales.")
	}
	estado := "revisar"
	if okObligatorios == totalObligatorios {
		estado = "listo"
	}
	puntuacion := 0
	if totalObligatorios > 0 {
		puntuacion = int(float64(okObligatorios) / float64(totalObligatorios) * 100)
	}
	return EmpresaModuloColombiaDiagnostico{
		EmpresaID:         empresaID,
		Modulo:            modulo,
		Titulo:            plantilla.Titulo,
		Estado:            estado,
		Puntuacion:        puntuacion,
		TotalObligatorios: totalObligatorios,
		OKObligatorios:    okObligatorios,
		Checks:            checks,
		Recomendaciones:   recomendaciones,
	}
}

func EnsureEmpresaModulosColombiaSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_modulos_colombia_registros (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			modulo TEXT NOT NULL,
			tipo TEXT DEFAULT 'general',
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			tercero TEXT,
			responsable TEXT,
			categoria TEXT,
			referencia TEXT,
			prioridad TEXT DEFAULT 'normal',
			estado TEXT DEFAULT 'borrador',
			fecha TEXT,
			fecha_vencimiento TEXT,
			valor REAL DEFAULT 0,
			metadata_json TEXT,
			usuario_creador TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(empresa_id,modulo,codigo)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_modulos_colombia_registros_empresa ON empresa_modulos_colombia_registros(empresa_id,modulo,estado,fecha_vencimiento)`,
		`CREATE TABLE IF NOT EXISTS empresa_modulos_colombia_eventos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			modulo TEXT NOT NULL,
			registro_id INTEGER DEFAULT 0,
			evento TEXT NOT NULL,
			estado_anterior TEXT,
			estado_nuevo TEXT,
			detalle TEXT,
			usuario TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS ix_modulos_colombia_eventos_empresa ON empresa_modulos_colombia_eventos(empresa_id,modulo,registro_id)`,
		`CREATE TABLE IF NOT EXISTS empresa_modulos_colombia_evidencias (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			modulo TEXT NOT NULL,
			registro_id INTEGER NOT NULL,
			tipo TEXT DEFAULT 'soporte',
			nombre TEXT NOT NULL,
			url TEXT,
			descripcion TEXT,
			usuario TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS ix_modulos_colombia_evidencias_empresa ON empresa_modulos_colombia_evidencias(empresa_id,modulo,registro_id)`,
		`CREATE TABLE IF NOT EXISTS empresa_modulos_colombia_aprobaciones (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			modulo TEXT NOT NULL,
			registro_id INTEGER NOT NULL,
			nivel TEXT DEFAULT 'operativo',
			solicitado_a TEXT NOT NULL,
			solicitado_por TEXT,
			estado TEXT DEFAULT 'pendiente',
			comentario TEXT,
			decision_por TEXT,
			fecha_decision TEXT,
			fecha_vencimiento TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS ix_modulos_colombia_aprobaciones_empresa ON empresa_modulos_colombia_aprobaciones(empresa_id,modulo,registro_id,estado)`,
		`CREATE TABLE IF NOT EXISTS empresa_modulos_colombia_tareas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			modulo TEXT NOT NULL,
			registro_id INTEGER NOT NULL,
			titulo TEXT NOT NULL,
			responsable TEXT,
			prioridad TEXT DEFAULT 'normal',
			estado TEXT DEFAULT 'pendiente',
			fecha_vencimiento TEXT,
			comentario TEXT,
			usuario_creador TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS ix_modulos_colombia_tareas_empresa ON empresa_modulos_colombia_tareas(empresa_id,modulo,registro_id,estado,fecha_vencimiento)`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func BuildEmpresaModuloColombiaDashboard(dbConn *sql.DB, empresaID int64, modulo string) (EmpresaModuloColombiaDashboard, error) {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return EmpresaModuloColombiaDashboard{}, err
	}
	modulo = NormalizeEmpresaModuloColombia(modulo)
	d := EmpresaModuloColombiaDashboard{EmpresaID: empresaID, Modulo: modulo, Titulo: empresaModuloColombiaTitulos[modulo]}
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_modulos_colombia_registros WHERE empresa_id=? AND modulo=?`, empresaID, modulo).Scan(&d.TotalRegistros)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_modulos_colombia_registros WHERE empresa_id=? AND modulo=? AND estado IN ('borrador','pendiente','abierto')`, empresaID, modulo).Scan(&d.Abiertos)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_modulos_colombia_registros WHERE empresa_id=? AND modulo=? AND estado IN ('en_revision','en_proceso','en_gestion')`, empresaID, modulo).Scan(&d.EnProceso)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_modulos_colombia_registros WHERE empresa_id=? AND modulo=? AND estado IN ('aprobado','cerrado','cumplido','pagado','resuelto')`, empresaID, modulo).Scan(&d.Aprobados)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_modulos_colombia_registros WHERE empresa_id=? AND modulo=? AND COALESCE(fecha_vencimiento,'')<>'' AND fecha_vencimiento<? AND estado NOT IN ('cerrado','cancelado','cumplido','pagado','resuelto')`, empresaID, modulo, time.Now().Format("2006-01-02")).Scan(&d.Vencidos)
	_ = QueryRowCompat(dbConn, `SELECT COALESCE(SUM(valor),0) FROM empresa_modulos_colombia_registros WHERE empresa_id=? AND modulo=? AND estado NOT IN ('cancelado')`, empresaID, modulo).Scan(&d.ValorTotal)
	d.Registros, _ = ListEmpresaModuloColombiaRegistros(dbConn, empresaID, modulo, "", 200)
	d.EventosRecientes, _ = ListEmpresaModuloColombiaEventos(dbConn, empresaID, modulo, 50)
	if d.Vencidos > 0 {
		d.Alertas = append(d.Alertas, "Hay registros vencidos que requieren gestion.")
	}
	if d.TotalRegistros == 0 {
		d.Alertas = append(d.Alertas, "No hay datos cargados para este modulo.")
	}
	if len(d.Alertas) == 0 {
		d.Alertas = append(d.Alertas, "Modulo sin alertas criticas.")
	}
	return d, nil
}

func BuildEmpresaModuloColombiaReporte(dbConn *sql.DB, empresaID int64, modulo string) (EmpresaModuloColombiaReporte, error) {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return EmpresaModuloColombiaReporte{}, err
	}
	modulo = NormalizeEmpresaModuloColombia(modulo)
	if empresaID <= 0 || modulo == "" {
		return EmpresaModuloColombiaReporte{}, errors.New("empresa_id y modulo son requeridos")
	}
	today := time.Now()
	todayStr := today.Format("2006-01-02")
	weekStr := today.AddDate(0, 0, 7).Format("2006-01-02")
	monthStr := today.AddDate(0, 0, 30).Format("2006-01-02")

	r := EmpresaModuloColombiaReporte{
		EmpresaID: empresaID,
		Modulo:    modulo,
		Titulo:    empresaModuloColombiaTitulos[modulo],
	}
	r.PorEstado, _ = listEmpresaModuloColombiaMetricas(dbConn, empresaID, modulo, "estado", 20)
	r.PorTipo, _ = listEmpresaModuloColombiaMetricas(dbConn, empresaID, modulo, "tipo", 20)
	r.PorCategoria, _ = listEmpresaModuloColombiaMetricas(dbConn, empresaID, modulo, "categoria", 20)
	r.PorPrioridad, _ = listEmpresaModuloColombiaMetricas(dbConn, empresaID, modulo, "prioridad", 10)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_modulos_colombia_registros WHERE empresa_id=? AND modulo=? AND COALESCE(fecha_vencimiento,'')<>'' AND fecha_vencimiento<? AND estado NOT IN ('cerrado','cancelado','cumplido','pagado','resuelto','rechazado')`, empresaID, modulo, todayStr).Scan(&r.Vencidos)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_modulos_colombia_registros WHERE empresa_id=? AND modulo=? AND COALESCE(fecha_vencimiento,'')<>'' AND fecha_vencimiento>=? AND fecha_vencimiento<=? AND estado NOT IN ('cerrado','cancelado','cumplido','pagado','resuelto','rechazado')`, empresaID, modulo, todayStr, weekStr).Scan(&r.Vencen7Dias)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_modulos_colombia_registros WHERE empresa_id=? AND modulo=? AND COALESCE(fecha_vencimiento,'')<>'' AND fecha_vencimiento>=? AND fecha_vencimiento<=? AND estado NOT IN ('cerrado','cancelado','cumplido','pagado','resuelto','rechazado')`, empresaID, modulo, todayStr, monthStr).Scan(&r.Vencen30Dias)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_modulos_colombia_registros WHERE empresa_id=? AND modulo=? AND prioridad IN ('critica','urgente') AND estado NOT IN ('cerrado','cancelado','cumplido','pagado','resuelto','rechazado')`, empresaID, modulo).Scan(&r.CriticosAbiertos)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_modulos_colombia_registros WHERE empresa_id=? AND modulo=? AND COALESCE(TRIM(responsable),'')='' AND estado NOT IN ('cerrado','cancelado','cumplido','pagado','resuelto','rechazado')`, empresaID, modulo).Scan(&r.SinResponsable)
	_ = QueryRowCompat(dbConn, `SELECT COALESCE(SUM(valor),0) FROM empresa_modulos_colombia_registros WHERE empresa_id=? AND modulo=? AND estado NOT IN ('cerrado','cancelado','cumplido','pagado','resuelto','rechazado')`, empresaID, modulo).Scan(&r.ValorPendiente)
	_ = QueryRowCompat(dbConn, `SELECT COALESCE(SUM(valor),0) FROM empresa_modulos_colombia_registros WHERE empresa_id=? AND modulo=? AND COALESCE(fecha_vencimiento,'')<>'' AND fecha_vencimiento<? AND estado NOT IN ('cerrado','cancelado','cumplido','pagado','resuelto','rechazado')`, empresaID, modulo, todayStr).Scan(&r.ValorVencido)
	_ = QueryRowCompat(dbConn, `SELECT COALESCE(MAX(fecha_creacion),'') FROM empresa_modulos_colombia_eventos WHERE empresa_id=? AND modulo=?`, empresaID, modulo).Scan(&r.UltimaActividad)
	r.Recomendaciones = recomendacionesModuloColombia(r)
	return r, nil
}

func BuildEmpresaModuloColombiaAgenda(dbConn *sql.DB, empresaID int64, modulo string) (EmpresaModuloColombiaAgenda, error) {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return EmpresaModuloColombiaAgenda{}, err
	}
	modulo = NormalizeEmpresaModuloColombia(modulo)
	if empresaID <= 0 || modulo == "" {
		return EmpresaModuloColombiaAgenda{}, errors.New("empresa_id y modulo son requeridos")
	}
	today := time.Now()
	todayStr := today.Format("2006-01-02")
	weekStr := today.AddDate(0, 0, 7).Format("2006-01-02")
	agenda := EmpresaModuloColombiaAgenda{
		EmpresaID:  empresaID,
		Modulo:     modulo,
		FechaCorte: todayStr,
		Items:      []EmpresaModuloColombiaAgendaItem{},
	}
	registros, _ := ListEmpresaModuloColombiaRegistros(dbConn, empresaID, modulo, "", 500)
	for _, row := range registros {
		if row.FechaVencimiento == "" || isModuloColombiaEstadoFinal(row.Estado) {
			continue
		}
		if row.FechaVencimiento < todayStr {
			agenda.RegistrosVencidos++
			agenda.Items = append(agenda.Items, EmpresaModuloColombiaAgendaItem{Tipo: "registro", RegistroID: row.ID, Codigo: row.Codigo, Titulo: row.Nombre, Responsable: row.Responsable, Estado: row.Estado, Prioridad: row.Prioridad, FechaVencimiento: row.FechaVencimiento, Severidad: "critica", Detalle: "Registro vencido"})
		} else if row.FechaVencimiento <= weekStr {
			agenda.RegistrosProximos++
			agenda.Items = append(agenda.Items, EmpresaModuloColombiaAgendaItem{Tipo: "registro", RegistroID: row.ID, Codigo: row.Codigo, Titulo: row.Nombre, Responsable: row.Responsable, Estado: row.Estado, Prioridad: row.Prioridad, FechaVencimiento: row.FechaVencimiento, Severidad: "alta", Detalle: "Registro vence en los proximos 7 dias"})
		}
	}
	tareas, _ := ListEmpresaModuloColombiaTareas(dbConn, empresaID, modulo, 0, "", 500)
	for _, row := range tareas {
		if row.FechaVencimiento == "" || row.Estado == "cumplida" || row.Estado == "cancelada" {
			continue
		}
		if row.FechaVencimiento < todayStr {
			agenda.TareasVencidas++
			agenda.Items = append(agenda.Items, EmpresaModuloColombiaAgendaItem{Tipo: "tarea", RegistroID: row.RegistroID, ReferenciaID: row.ID, Titulo: row.Titulo, Responsable: row.Responsable, Estado: row.Estado, Prioridad: row.Prioridad, FechaVencimiento: row.FechaVencimiento, Severidad: "critica", Detalle: "Tarea vencida"})
		} else if row.FechaVencimiento <= weekStr {
			agenda.TareasProximas++
			agenda.Items = append(agenda.Items, EmpresaModuloColombiaAgendaItem{Tipo: "tarea", RegistroID: row.RegistroID, ReferenciaID: row.ID, Titulo: row.Titulo, Responsable: row.Responsable, Estado: row.Estado, Prioridad: row.Prioridad, FechaVencimiento: row.FechaVencimiento, Severidad: "media", Detalle: "Tarea vence en los proximos 7 dias"})
		}
	}
	aprobaciones, _ := ListEmpresaModuloColombiaAprobaciones(dbConn, empresaID, modulo, 0, "pendiente", 500)
	for _, row := range aprobaciones {
		agenda.AprobacionesPendientes++
		if row.FechaVencimiento != "" && row.FechaVencimiento < todayStr {
			agenda.AprobacionesVencidas++
			agenda.Items = append(agenda.Items, EmpresaModuloColombiaAgendaItem{Tipo: "aprobacion", RegistroID: row.RegistroID, ReferenciaID: row.ID, Titulo: "Aprobacion pendiente", Responsable: row.SolicitadoA, Estado: row.Estado, Prioridad: row.Nivel, FechaVencimiento: row.FechaVencimiento, Severidad: "critica", Detalle: row.Comentario})
		} else if row.FechaVencimiento != "" && row.FechaVencimiento <= weekStr {
			agenda.Items = append(agenda.Items, EmpresaModuloColombiaAgendaItem{Tipo: "aprobacion", RegistroID: row.RegistroID, ReferenciaID: row.ID, Titulo: "Aprobacion pendiente", Responsable: row.SolicitadoA, Estado: row.Estado, Prioridad: row.Nivel, FechaVencimiento: row.FechaVencimiento, Severidad: "alta", Detalle: row.Comentario})
		}
	}
	sort.SliceStable(agenda.Items, func(i, j int) bool {
		if agenda.Items[i].Severidad != agenda.Items[j].Severidad {
			return moduloColombiaSeveridadRank(agenda.Items[i].Severidad) < moduloColombiaSeveridadRank(agenda.Items[j].Severidad)
		}
		return agenda.Items[i].FechaVencimiento < agenda.Items[j].FechaVencimiento
	})
	if len(agenda.Items) > 80 {
		agenda.Items = agenda.Items[:80]
	}
	agenda.TotalAlertas = len(agenda.Items)
	agenda.Recomendaciones = recomendacionesModuloColombiaAgenda(agenda)
	return agenda, nil
}

func BuildEmpresaModuloColombiaResponsables(dbConn *sql.DB, empresaID int64, modulo string) ([]EmpresaModuloColombiaResponsableResumen, error) {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return nil, err
	}
	modulo = NormalizeEmpresaModuloColombia(modulo)
	if empresaID <= 0 || modulo == "" {
		return nil, errors.New("empresa_id y modulo son requeridos")
	}
	today := time.Now().Format("2006-01-02")
	resumen := map[string]*EmpresaModuloColombiaResponsableResumen{}
	registros, _ := ListEmpresaModuloColombiaRegistros(dbConn, empresaID, modulo, "", 1000)
	for _, row := range registros {
		if isModuloColombiaEstadoFinal(row.Estado) {
			continue
		}
		r := getModuloColombiaResponsableResumen(resumen, row.Responsable)
		r.RegistrosAbiertos++
		if row.FechaVencimiento != "" && row.FechaVencimiento < today {
			r.RegistrosVencidos++
		}
	}
	tareas, _ := ListEmpresaModuloColombiaTareas(dbConn, empresaID, modulo, 0, "", 1000)
	for _, row := range tareas {
		if row.Estado != "pendiente" && row.Estado != "en_proceso" {
			continue
		}
		r := getModuloColombiaResponsableResumen(resumen, row.Responsable)
		r.TareasAbiertas++
		if row.FechaVencimiento != "" && row.FechaVencimiento < today {
			r.TareasVencidas++
		}
	}
	aprobaciones, _ := ListEmpresaModuloColombiaAprobaciones(dbConn, empresaID, modulo, 0, "pendiente", 1000)
	for _, row := range aprobaciones {
		r := getModuloColombiaResponsableResumen(resumen, row.SolicitadoA)
		r.AprobacionesPendientes++
	}
	out := make([]EmpresaModuloColombiaResponsableResumen, 0, len(resumen))
	for _, row := range resumen {
		row.TotalPendiente = row.RegistrosAbiertos + row.TareasAbiertas + row.AprobacionesPendientes
		row.Recomendacion = recomendacionModuloColombiaResponsable(*row)
		out = append(out, *row)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].TotalPendiente != out[j].TotalPendiente {
			return out[i].TotalPendiente > out[j].TotalPendiente
		}
		return out[i].Responsable < out[j].Responsable
	})
	if len(out) > 30 {
		out = out[:30]
	}
	return out, nil
}

func BuildEmpresaModuloColombiaSLA(dbConn *sql.DB, empresaID int64, modulo string) (EmpresaModuloColombiaSLA, error) {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return EmpresaModuloColombiaSLA{}, err
	}
	modulo = NormalizeEmpresaModuloColombia(modulo)
	if empresaID <= 0 || modulo == "" {
		return EmpresaModuloColombiaSLA{}, errors.New("empresa_id y modulo son requeridos")
	}
	today := time.Now()
	todayStr := today.Format("2006-01-02")
	weekStr := today.AddDate(0, 0, 7).Format("2006-01-02")
	sla := EmpresaModuloColombiaSLA{
		EmpresaID:  empresaID,
		Modulo:     modulo,
		FechaCorte: todayStr,
		Buckets:    map[string]int{"sin_vencimiento": 0, "vencido": 0, "0_7": 0, "8_30": 0, "mas_30": 0},
	}
	registros, _ := ListEmpresaModuloColombiaRegistros(dbConn, empresaID, modulo, "", 1000)
	for _, row := range registros {
		if isModuloColombiaEstadoFinal(row.Estado) {
			continue
		}
		sla.TotalAbiertos++
		bucket := bucketModuloColombiaVencimiento(row.FechaVencimiento, today)
		sla.Buckets[bucket]++
		switch bucket {
		case "sin_vencimiento":
			sla.SinVencimiento++
		case "vencido":
			sla.Vencidos++
		case "0_7":
			sla.Proximos7Dias++
		}
		if row.FechaVencimiento != "" && row.FechaVencimiento <= weekStr && row.FechaVencimiento >= todayStr {
			sla.Proximos7Dias = maxIntModuloColombia(sla.Proximos7Dias, sla.Buckets["0_7"])
		}
	}
	tareas, _ := ListEmpresaModuloColombiaTareas(dbConn, empresaID, modulo, 0, "", 1000)
	for _, row := range tareas {
		if row.Estado != "pendiente" && row.Estado != "en_proceso" {
			continue
		}
		sla.TareasAbiertas++
		if row.FechaVencimiento != "" && row.FechaVencimiento < todayStr {
			sla.TareasVencidas++
		}
	}
	sla.CumplimientoPct = calcularCumplimientoModuloColombia(sla.TotalAbiertos+sla.TareasAbiertas, sla.Vencidos+sla.TareasVencidas)
	sla.Semaforo = semaforoModuloColombiaSLA(sla.CumplimientoPct, sla.Vencidos+sla.TareasVencidas)
	sla.Recomendaciones = recomendacionesModuloColombiaSLA(sla)
	return sla, nil
}

func BuildEmpresaModuloColombiaRiesgo(dbConn *sql.DB, empresaID int64, modulo string) (EmpresaModuloColombiaRiesgo, error) {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return EmpresaModuloColombiaRiesgo{}, err
	}
	modulo = NormalizeEmpresaModuloColombia(modulo)
	if empresaID <= 0 || modulo == "" {
		return EmpresaModuloColombiaRiesgo{}, errors.New("empresa_id y modulo son requeridos")
	}
	reporte, _ := BuildEmpresaModuloColombiaReporte(dbConn, empresaID, modulo)
	agenda, _ := BuildEmpresaModuloColombiaAgenda(dbConn, empresaID, modulo)
	sla, _ := BuildEmpresaModuloColombiaSLA(dbConn, empresaID, modulo)
	riesgo := EmpresaModuloColombiaRiesgo{
		EmpresaID:              empresaID,
		Modulo:                 modulo,
		RegistrosVencidos:      agenda.RegistrosVencidos,
		CriticosAbiertos:       reporte.CriticosAbiertos,
		SinResponsable:         reporte.SinResponsable,
		AprobacionesPendientes: agenda.AprobacionesPendientes,
		TareasAbiertas:         sla.TareasAbiertas,
		TareasVencidas:         agenda.TareasVencidas,
		Factores:               []string{},
	}
	registros, _ := ListEmpresaModuloColombiaRegistros(dbConn, empresaID, modulo, "", 500)
	for _, row := range registros {
		if isModuloColombiaEstadoFinal(row.Estado) {
			continue
		}
		evidencias, _ := ListEmpresaModuloColombiaEvidencias(dbConn, empresaID, modulo, row.ID, 1)
		if len(evidencias) == 0 {
			riesgo.SinEvidencia++
		}
	}
	riesgo.Score, riesgo.Factores = scoreModuloColombiaRiesgo(riesgo)
	riesgo.Nivel = nivelModuloColombiaRiesgo(riesgo.Score)
	riesgo.Recomendaciones = recomendacionesModuloColombiaRiesgo(riesgo)
	return riesgo, nil
}

func BuildEmpresaModuloColombiaExportacion(dbConn *sql.DB, empresaID int64, modulo string) (EmpresaModuloColombiaExportacion, error) {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return EmpresaModuloColombiaExportacion{}, err
	}
	modulo = NormalizeEmpresaModuloColombia(modulo)
	if empresaID <= 0 || modulo == "" {
		return EmpresaModuloColombiaExportacion{}, errors.New("empresa_id y modulo son requeridos")
	}
	registros, err := ListEmpresaModuloColombiaRegistros(dbConn, empresaID, modulo, "", 1000)
	if err != nil {
		return EmpresaModuloColombiaExportacion{}, err
	}
	reporte, err := BuildEmpresaModuloColombiaReporte(dbConn, empresaID, modulo)
	if err != nil {
		return EmpresaModuloColombiaExportacion{}, err
	}
	agenda, err := BuildEmpresaModuloColombiaAgenda(dbConn, empresaID, modulo)
	if err != nil {
		return EmpresaModuloColombiaExportacion{}, err
	}
	responsables, err := BuildEmpresaModuloColombiaResponsables(dbConn, empresaID, modulo)
	if err != nil {
		return EmpresaModuloColombiaExportacion{}, err
	}
	sla, err := BuildEmpresaModuloColombiaSLA(dbConn, empresaID, modulo)
	if err != nil {
		return EmpresaModuloColombiaExportacion{}, err
	}
	riesgo, err := BuildEmpresaModuloColombiaRiesgo(dbConn, empresaID, modulo)
	if err != nil {
		return EmpresaModuloColombiaExportacion{}, err
	}
	eventos, _ := ListEmpresaModuloColombiaEventos(dbConn, empresaID, modulo, 300)
	evidencias, _ := ListEmpresaModuloColombiaEvidencias(dbConn, empresaID, modulo, 0, 500)
	aprobaciones, _ := ListEmpresaModuloColombiaAprobaciones(dbConn, empresaID, modulo, 0, "", 500)
	tareas, _ := ListEmpresaModuloColombiaTareas(dbConn, empresaID, modulo, 0, "", 500)

	out := EmpresaModuloColombiaExportacion{
		EmpresaID:  empresaID,
		Modulo:     modulo,
		Titulo:     empresaModuloColombiaTitulos[modulo],
		FechaCorte: time.Now().Format("2006-01-02 15:04:05"),
		Secciones:  []EmpresaModuloColombiaCSVSeccion{},
	}
	out.Secciones = append(out.Secciones,
		EmpresaModuloColombiaCSVSeccion{
			Nombre:  "resumen_ejecutivo",
			Headers: []string{"campo", "valor"},
			Rows: [][]string{
				{"modulo", out.Titulo},
				{"fecha_corte", out.FechaCorte},
				{"vencidos", fmt.Sprintf("%d", reporte.Vencidos)},
				{"vencen_7_dias", fmt.Sprintf("%d", reporte.Vencen7Dias)},
				{"vencen_30_dias", fmt.Sprintf("%d", reporte.Vencen30Dias)},
				{"criticos_abiertos", fmt.Sprintf("%d", reporte.CriticosAbiertos)},
				{"sin_responsable", fmt.Sprintf("%d", reporte.SinResponsable)},
				{"valor_pendiente", formatFloatModuloColombia(reporte.ValorPendiente)},
				{"valor_vencido", formatFloatModuloColombia(reporte.ValorVencido)},
				{"cumplimiento_sla_pct", formatFloatModuloColombia(sla.CumplimientoPct)},
				{"semaforo_sla", sla.Semaforo},
				{"score_riesgo", fmt.Sprintf("%d", riesgo.Score)},
				{"nivel_riesgo", riesgo.Nivel},
			},
		},
		exportacionModuloColombiaRegistros(registros),
		exportacionModuloColombiaAgenda(agenda.Items),
		exportacionModuloColombiaResponsables(responsables),
		exportacionModuloColombiaMetricas("metricas_estado", reporte.PorEstado),
		exportacionModuloColombiaMetricas("metricas_tipo", reporte.PorTipo),
		exportacionModuloColombiaMetricas("metricas_categoria", reporte.PorCategoria),
		exportacionModuloColombiaMetricas("metricas_prioridad", reporte.PorPrioridad),
		exportacionModuloColombiaSLA(sla),
		exportacionModuloColombiaRiesgo(riesgo),
		exportacionModuloColombiaTareas(tareas),
		exportacionModuloColombiaAprobaciones(aprobaciones),
		exportacionModuloColombiaEvidencias(evidencias),
		exportacionModuloColombiaEventos(eventos),
		exportacionModuloColombiaLista("recomendaciones", []string{"origen", "detalle"}, append(exportacionModuloColombiaPrefijo("reporte", reporte.Recomendaciones), append(exportacionModuloColombiaPrefijo("agenda", agenda.Recomendaciones), append(exportacionModuloColombiaPrefijo("sla", sla.Recomendaciones), exportacionModuloColombiaPrefijo("riesgo", riesgo.Recomendaciones)...)...)...)),
	)
	return out, nil
}

func GenerarEmpresaModuloColombiaPlanAccion(dbConn *sql.DB, empresaID int64, modulo, usuario string) (EmpresaModuloColombiaPlanAccionResult, error) {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return EmpresaModuloColombiaPlanAccionResult{}, err
	}
	modulo = NormalizeEmpresaModuloColombia(modulo)
	if empresaID <= 0 || modulo == "" {
		return EmpresaModuloColombiaPlanAccionResult{}, errors.New("empresa_id y modulo son requeridos")
	}
	agenda, err := BuildEmpresaModuloColombiaAgenda(dbConn, empresaID, modulo)
	if err != nil {
		return EmpresaModuloColombiaPlanAccionResult{}, err
	}
	result := EmpresaModuloColombiaPlanAccionResult{AlertasRevisadas: len(agenda.Items), Detalles: []string{}}
	for i, item := range agenda.Items {
		if i >= 25 {
			result.Omitidas += len(agenda.Items) - i
			result.Detalles = append(result.Detalles, "Se limitaron las tareas automaticas a las primeras 25 alertas.")
			break
		}
		if item.RegistroID <= 0 {
			result.Omitidas++
			continue
		}
		titulo := tituloPlanAccionModuloColombia(item)
		if existeTareaAbiertaModuloColombia(dbConn, empresaID, modulo, item.RegistroID, titulo) {
			result.Omitidas++
			result.Detalles = append(result.Detalles, "Ya existe tarea abierta: "+titulo)
			continue
		}
		vence := item.FechaVencimiento
		if strings.TrimSpace(vence) == "" {
			vence = time.Now().AddDate(0, 0, 2).Format("2006-01-02")
		}
		_, err := CrearEmpresaModuloColombiaTarea(dbConn, EmpresaModuloColombiaTarea{
			EmpresaID:        empresaID,
			Modulo:           modulo,
			RegistroID:       item.RegistroID,
			Titulo:           titulo,
			Responsable:      item.Responsable,
			Prioridad:        prioridadPlanAccionModuloColombia(item.Severidad),
			Estado:           "pendiente",
			FechaVencimiento: vence,
			Comentario:       strings.TrimSpace(item.Detalle),
			UsuarioCreador:   usuario,
		})
		if err != nil {
			result.Omitidas++
			result.Detalles = append(result.Detalles, fmt.Sprintf("No se pudo crear tarea para registro %d: %v", item.RegistroID, err))
			continue
		}
		result.TareasCreadas++
	}
	_ = RegistrarEmpresaModuloColombiaEvento(dbConn, empresaID, modulo, 0, "plan_accion_generado", "", "", fmt.Sprintf("Tareas creadas: %d, omitidas: %d", result.TareasCreadas, result.Omitidas), usuario)
	return result, nil
}

func GetEmpresaModuloColombiaExpediente(dbConn *sql.DB, empresaID int64, modulo string, registroID int64) (EmpresaModuloColombiaExpediente, error) {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return EmpresaModuloColombiaExpediente{}, err
	}
	modulo = NormalizeEmpresaModuloColombia(modulo)
	if empresaID <= 0 || modulo == "" || registroID <= 0 {
		return EmpresaModuloColombiaExpediente{}, errors.New("empresa_id, modulo y registro_id son requeridos")
	}
	registro, err := GetEmpresaModuloColombiaRegistro(dbConn, empresaID, modulo, registroID)
	if err != nil {
		return EmpresaModuloColombiaExpediente{}, err
	}
	eventos, _ := ListEmpresaModuloColombiaEventosPorRegistro(dbConn, empresaID, modulo, registroID, 80)
	evidencias, _ := ListEmpresaModuloColombiaEvidencias(dbConn, empresaID, modulo, registroID, 80)
	aprobaciones, _ := ListEmpresaModuloColombiaAprobaciones(dbConn, empresaID, modulo, registroID, "", 80)
	tareas, _ := ListEmpresaModuloColombiaTareas(dbConn, empresaID, modulo, registroID, "", 80)
	resumen := map[string]int{
		"eventos":                 len(eventos),
		"evidencias":              len(evidencias),
		"aprobaciones":            len(aprobaciones),
		"aprobaciones_pendientes": countModuloColombiaAprobacionesEstado(aprobaciones, "pendiente"),
		"tareas":                  len(tareas),
		"tareas_abiertas":         countModuloColombiaTareasAbiertas(tareas),
	}
	return EmpresaModuloColombiaExpediente{
		EmpresaID:     empresaID,
		Modulo:        modulo,
		Registro:      registro,
		Eventos:       eventos,
		Evidencias:    evidencias,
		Aprobaciones:  aprobaciones,
		Tareas:        tareas,
		Resumen:       resumen,
		Recomendacion: recomendacionEmpresaModuloColombiaExpediente(registro, resumen),
	}, nil
}

func UpsertEmpresaModuloColombiaRegistro(dbConn *sql.DB, row EmpresaModuloColombiaRegistro) (int64, error) {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return 0, err
	}
	row = normalizeEmpresaModuloColombiaRegistro(row)
	if row.EmpresaID <= 0 || row.Modulo == "" || row.Codigo == "" || row.Nombre == "" {
		return 0, errors.New("empresa_id, modulo, codigo y nombre son requeridos")
	}
	if row.ID > 0 {
		old := ""
		_ = QueryRowCompat(dbConn, `SELECT COALESCE(estado,'') FROM empresa_modulos_colombia_registros WHERE empresa_id=? AND modulo=? AND id=?`, row.EmpresaID, row.Modulo, row.ID).Scan(&old)
		_, err := ExecCompat(dbConn, `UPDATE empresa_modulos_colombia_registros SET tipo=?,codigo=?,nombre=?,tercero=?,responsable=?,categoria=?,referencia=?,prioridad=?,estado=?,fecha=?,fecha_vencimiento=?,valor=?,metadata_json=?,usuario_creador=?,fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND modulo=? AND id=?`,
			row.Tipo, row.Codigo, row.Nombre, row.Tercero, row.Responsable, row.Categoria, row.Referencia, row.Prioridad, row.Estado, row.Fecha, row.FechaVencimiento, row.Valor, row.MetadataJSON, row.UsuarioCreador, row.EmpresaID, row.Modulo, row.ID)
		if err == nil {
			_ = registrarEmpresaModuloColombiaEvento(dbConn, row.EmpresaID, row.Modulo, row.ID, "registro_actualizado", old, row.Estado, row.Nombre, row.UsuarioCreador)
		}
		return row.ID, err
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_modulos_colombia_registros (empresa_id,modulo,tipo,codigo,nombre,tercero,responsable,categoria,referencia,prioridad,estado,fecha,fecha_vencimiento,valor,metadata_json,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (empresa_id,modulo,codigo) DO UPDATE SET tipo=EXCLUDED.tipo,nombre=EXCLUDED.nombre,tercero=EXCLUDED.tercero,responsable=EXCLUDED.responsable,categoria=EXCLUDED.categoria,referencia=EXCLUDED.referencia,prioridad=EXCLUDED.prioridad,estado=EXCLUDED.estado,fecha=EXCLUDED.fecha,fecha_vencimiento=EXCLUDED.fecha_vencimiento,valor=EXCLUDED.valor,metadata_json=EXCLUDED.metadata_json,usuario_creador=EXCLUDED.usuario_creador,fecha_actualizacion=CURRENT_TIMESTAMP`,
		row.EmpresaID, row.Modulo, row.Tipo, row.Codigo, row.Nombre, row.Tercero, row.Responsable, row.Categoria, row.Referencia, row.Prioridad, row.Estado, row.Fecha, row.FechaVencimiento, row.Valor, row.MetadataJSON, row.UsuarioCreador)
	if err == nil {
		_ = registrarEmpresaModuloColombiaEvento(dbConn, row.EmpresaID, row.Modulo, id, "registro_guardado", "", row.Estado, row.Nombre, row.UsuarioCreador)
	}
	return id, err
}

func ListEmpresaModuloColombiaRegistros(dbConn *sql.DB, empresaID int64, modulo, estado string, limit int) ([]EmpresaModuloColombiaRegistro, error) {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 1000 {
		limit = 300
	}
	modulo = NormalizeEmpresaModuloColombia(modulo)
	args := []interface{}{empresaID, modulo}
	where := "empresa_id=? AND modulo=?"
	if strings.TrimSpace(estado) != "" {
		estado = normalizeModuloColombiaEstado(estado)
	}
	if estado != "" && estado != "todos" {
		where += " AND estado=?"
		args = append(args, estado)
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,modulo,COALESCE(tipo,''),codigo,nombre,COALESCE(tercero,''),COALESCE(responsable,''),COALESCE(categoria,''),COALESCE(referencia,''),COALESCE(prioridad,'normal'),COALESCE(estado,'borrador'),COALESCE(fecha,''),COALESCE(fecha_vencimiento,''),COALESCE(valor,0),COALESCE(metadata_json,''),COALESCE(usuario_creador,''),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,'') FROM empresa_modulos_colombia_registros WHERE %s ORDER BY id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaModuloColombiaRegistro{}
	for rows.Next() {
		var x EmpresaModuloColombiaRegistro
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Modulo, &x.Tipo, &x.Codigo, &x.Nombre, &x.Tercero, &x.Responsable, &x.Categoria, &x.Referencia, &x.Prioridad, &x.Estado, &x.Fecha, &x.FechaVencimiento, &x.Valor, &x.MetadataJSON, &x.UsuarioCreador, &x.FechaCreacion, &x.FechaActualizacion); err != nil {
			return nil, err
		}
		x.Metadata = decodeModuloColombiaMetadata(x.MetadataJSON)
		out = append(out, x)
	}
	return out, rows.Err()
}

func BuscarEmpresaModuloColombiaRegistros(dbConn *sql.DB, empresaID int64, modulo string, filtro EmpresaModuloColombiaFiltro, limit int) (EmpresaModuloColombiaBusqueda, error) {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return EmpresaModuloColombiaBusqueda{}, err
	}
	modulo = NormalizeEmpresaModuloColombia(modulo)
	if empresaID <= 0 || modulo == "" {
		return EmpresaModuloColombiaBusqueda{}, errors.New("empresa_id y modulo son requeridos")
	}
	filtro = normalizeEmpresaModuloColombiaFiltro(filtro)
	if limit <= 0 || limit > 1000 {
		limit = 300
	}
	args := []interface{}{empresaID, modulo}
	where := []string{"empresa_id=?", "modulo=?"}
	if filtro.Texto != "" {
		like := "%" + strings.ToLower(filtro.Texto) + "%"
		where = append(where, `(LOWER(COALESCE(codigo,'')) LIKE ? OR LOWER(COALESCE(nombre,'')) LIKE ? OR LOWER(COALESCE(tercero,'')) LIKE ? OR LOWER(COALESCE(referencia,'')) LIKE ? OR LOWER(COALESCE(metadata_json,'')) LIKE ?)`)
		args = append(args, like, like, like, like, like)
	}
	if filtro.Estado != "" && filtro.Estado != "todos" {
		where = append(where, "estado=?")
		args = append(args, filtro.Estado)
	}
	if filtro.Tipo != "" {
		where = append(where, "tipo=?")
		args = append(args, filtro.Tipo)
	}
	if filtro.Categoria != "" {
		where = append(where, "categoria=?")
		args = append(args, filtro.Categoria)
	}
	if filtro.Prioridad != "" {
		where = append(where, "prioridad=?")
		args = append(args, filtro.Prioridad)
	}
	if filtro.Responsable != "" {
		where = append(where, "LOWER(COALESCE(responsable,'')) LIKE ?")
		args = append(args, "%"+strings.ToLower(filtro.Responsable)+"%")
	}
	today := time.Now().Format("2006-01-02")
	if filtro.Vencidos {
		where = append(where, "COALESCE(fecha_vencimiento,'')<>'' AND fecha_vencimiento<? AND estado NOT IN ('cerrado','cancelado','cumplido','pagado','resuelto','rechazado')")
		args = append(args, today)
	} else if filtro.ProximosDias > 0 {
		until := time.Now().AddDate(0, 0, filtro.ProximosDias).Format("2006-01-02")
		where = append(where, "COALESCE(fecha_vencimiento,'')<>'' AND fecha_vencimiento>=? AND fecha_vencimiento<=? AND estado NOT IN ('cerrado','cancelado','cumplido','pagado','resuelto','rechazado')")
		args = append(args, today, until)
	}
	query := fmt.Sprintf(`SELECT id,empresa_id,modulo,COALESCE(tipo,''),codigo,nombre,COALESCE(tercero,''),COALESCE(responsable,''),COALESCE(categoria,''),COALESCE(referencia,''),COALESCE(prioridad,'normal'),COALESCE(estado,'borrador'),COALESCE(fecha,''),COALESCE(fecha_vencimiento,''),COALESCE(valor,0),COALESCE(metadata_json,''),COALESCE(usuario_creador,''),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,'') FROM empresa_modulos_colombia_registros WHERE %s ORDER BY CASE prioridad WHEN 'urgente' THEN 0 WHEN 'critica' THEN 1 WHEN 'alta' THEN 2 ELSE 3 END, COALESCE(fecha_vencimiento,'9999-12-31') ASC, id DESC LIMIT %d`, strings.Join(where, " AND "), limit)
	rows, err := ExecQueryCompat(dbConn, query, args...)
	if err != nil {
		return EmpresaModuloColombiaBusqueda{}, err
	}
	defer rows.Close()
	out := EmpresaModuloColombiaBusqueda{EmpresaID: empresaID, Modulo: modulo, Filtros: filtro, FechaCorte: time.Now().Format("2006-01-02 15:04:05")}
	for rows.Next() {
		var x EmpresaModuloColombiaRegistro
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Modulo, &x.Tipo, &x.Codigo, &x.Nombre, &x.Tercero, &x.Responsable, &x.Categoria, &x.Referencia, &x.Prioridad, &x.Estado, &x.Fecha, &x.FechaVencimiento, &x.Valor, &x.MetadataJSON, &x.UsuarioCreador, &x.FechaCreacion, &x.FechaActualizacion); err != nil {
			return EmpresaModuloColombiaBusqueda{}, err
		}
		x.Metadata = decodeModuloColombiaMetadata(x.MetadataJSON)
		out.Registros = append(out.Registros, x)
	}
	if err := rows.Err(); err != nil {
		return EmpresaModuloColombiaBusqueda{}, err
	}
	out.Total = len(out.Registros)
	return out, nil
}

func GetEmpresaModuloColombiaRegistro(dbConn *sql.DB, empresaID int64, modulo string, registroID int64) (EmpresaModuloColombiaRegistro, error) {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return EmpresaModuloColombiaRegistro{}, err
	}
	modulo = NormalizeEmpresaModuloColombia(modulo)
	var x EmpresaModuloColombiaRegistro
	err := QueryRowCompat(dbConn, `SELECT id,empresa_id,modulo,COALESCE(tipo,''),codigo,nombre,COALESCE(tercero,''),COALESCE(responsable,''),COALESCE(categoria,''),COALESCE(referencia,''),COALESCE(prioridad,'normal'),COALESCE(estado,'borrador'),COALESCE(fecha,''),COALESCE(fecha_vencimiento,''),COALESCE(valor,0),COALESCE(metadata_json,''),COALESCE(usuario_creador,''),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,'') FROM empresa_modulos_colombia_registros WHERE empresa_id=? AND modulo=? AND id=?`, empresaID, modulo, registroID).Scan(&x.ID, &x.EmpresaID, &x.Modulo, &x.Tipo, &x.Codigo, &x.Nombre, &x.Tercero, &x.Responsable, &x.Categoria, &x.Referencia, &x.Prioridad, &x.Estado, &x.Fecha, &x.FechaVencimiento, &x.Valor, &x.MetadataJSON, &x.UsuarioCreador, &x.FechaCreacion, &x.FechaActualizacion)
	if err != nil {
		return EmpresaModuloColombiaRegistro{}, err
	}
	x.Metadata = decodeModuloColombiaMetadata(x.MetadataJSON)
	return x, nil
}

func ListEmpresaModuloColombiaEventos(dbConn *sql.DB, empresaID int64, modulo string, limit int) ([]EmpresaModuloColombiaEvento, error) {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	modulo = NormalizeEmpresaModuloColombia(modulo)
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,modulo,COALESCE(registro_id,0),evento,COALESCE(estado_anterior,''),COALESCE(estado_nuevo,''),COALESCE(detalle,''),COALESCE(usuario,''),COALESCE(fecha_creacion,'') FROM empresa_modulos_colombia_eventos WHERE empresa_id=? AND modulo=? ORDER BY id DESC LIMIT %d`, limit), empresaID, modulo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaModuloColombiaEvento{}
	for rows.Next() {
		var x EmpresaModuloColombiaEvento
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Modulo, &x.RegistroID, &x.Evento, &x.EstadoAnterior, &x.EstadoNuevo, &x.Detalle, &x.Usuario, &x.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func ListEmpresaModuloColombiaEventosPorRegistro(dbConn *sql.DB, empresaID int64, modulo string, registroID int64, limit int) ([]EmpresaModuloColombiaEvento, error) {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	modulo = NormalizeEmpresaModuloColombia(modulo)
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,modulo,COALESCE(registro_id,0),evento,COALESCE(estado_anterior,''),COALESCE(estado_nuevo,''),COALESCE(detalle,''),COALESCE(usuario,''),COALESCE(fecha_creacion,'') FROM empresa_modulos_colombia_eventos WHERE empresa_id=? AND modulo=? AND registro_id=? ORDER BY id DESC LIMIT %d`, limit), empresaID, modulo, registroID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaModuloColombiaEvento{}
	for rows.Next() {
		var x EmpresaModuloColombiaEvento
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Modulo, &x.RegistroID, &x.Evento, &x.EstadoAnterior, &x.EstadoNuevo, &x.Detalle, &x.Usuario, &x.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func ListEmpresaModuloColombiaEvidencias(dbConn *sql.DB, empresaID int64, modulo string, registroID int64, limit int) ([]EmpresaModuloColombiaEvidencia, error) {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	modulo = NormalizeEmpresaModuloColombia(modulo)
	args := []interface{}{empresaID, modulo}
	where := "empresa_id=? AND modulo=?"
	if registroID > 0 {
		where += " AND registro_id=?"
		args = append(args, registroID)
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,modulo,COALESCE(registro_id,0),COALESCE(tipo,'soporte'),nombre,COALESCE(url,''),COALESCE(descripcion,''),COALESCE(usuario,''),COALESCE(fecha_creacion,'') FROM empresa_modulos_colombia_evidencias WHERE %s ORDER BY id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaModuloColombiaEvidencia{}
	for rows.Next() {
		var x EmpresaModuloColombiaEvidencia
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Modulo, &x.RegistroID, &x.Tipo, &x.Nombre, &x.URL, &x.Descripcion, &x.Usuario, &x.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func ListEmpresaModuloColombiaAprobaciones(dbConn *sql.DB, empresaID int64, modulo string, registroID int64, estado string, limit int) ([]EmpresaModuloColombiaAprobacion, error) {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	modulo = NormalizeEmpresaModuloColombia(modulo)
	args := []interface{}{empresaID, modulo}
	where := "empresa_id=? AND modulo=?"
	if registroID > 0 {
		where += " AND registro_id=?"
		args = append(args, registroID)
	}
	if strings.TrimSpace(estado) != "" {
		estado = normalizeModuloColombiaAprobacionEstado(estado)
	}
	if estado != "" && estado != "todos" {
		where += " AND estado=?"
		args = append(args, estado)
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,modulo,COALESCE(registro_id,0),COALESCE(nivel,'operativo'),solicitado_a,COALESCE(solicitado_por,''),COALESCE(estado,'pendiente'),COALESCE(comentario,''),COALESCE(decision_por,''),COALESCE(fecha_decision,''),COALESCE(fecha_creacion,''),COALESCE(fecha_vencimiento,'') FROM empresa_modulos_colombia_aprobaciones WHERE %s ORDER BY id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaModuloColombiaAprobacion{}
	for rows.Next() {
		var x EmpresaModuloColombiaAprobacion
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Modulo, &x.RegistroID, &x.Nivel, &x.SolicitadoA, &x.SolicitadoPor, &x.Estado, &x.Comentario, &x.DecisionPor, &x.FechaDecision, &x.FechaCreacion, &x.FechaVencimiento); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func ListEmpresaModuloColombiaTareas(dbConn *sql.DB, empresaID int64, modulo string, registroID int64, estado string, limit int) ([]EmpresaModuloColombiaTarea, error) {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	modulo = NormalizeEmpresaModuloColombia(modulo)
	args := []interface{}{empresaID, modulo}
	where := "empresa_id=? AND modulo=?"
	if registroID > 0 {
		where += " AND registro_id=?"
		args = append(args, registroID)
	}
	if strings.TrimSpace(estado) != "" {
		estado = normalizeModuloColombiaTareaEstado(estado)
	}
	if estado != "" && estado != "todos" {
		where += " AND estado=?"
		args = append(args, estado)
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,modulo,COALESCE(registro_id,0),titulo,COALESCE(responsable,''),COALESCE(prioridad,'normal'),COALESCE(estado,'pendiente'),COALESCE(fecha_vencimiento,''),COALESCE(comentario,''),COALESCE(usuario_creador,''),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,'') FROM empresa_modulos_colombia_tareas WHERE %s ORDER BY CASE estado WHEN 'pendiente' THEN 0 WHEN 'en_proceso' THEN 1 ELSE 2 END, COALESCE(fecha_vencimiento,'9999-12-31') ASC, id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaModuloColombiaTarea{}
	for rows.Next() {
		var x EmpresaModuloColombiaTarea
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Modulo, &x.RegistroID, &x.Titulo, &x.Responsable, &x.Prioridad, &x.Estado, &x.FechaVencimiento, &x.Comentario, &x.UsuarioCreador, &x.FechaCreacion, &x.FechaActualizacion); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func CambiarEstadoEmpresaModuloColombiaRegistro(dbConn *sql.DB, empresaID int64, modulo string, registroID int64, estado, detalle, usuario string) error {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return err
	}
	modulo = NormalizeEmpresaModuloColombia(modulo)
	estado = normalizeModuloColombiaEstado(estado)
	if empresaID <= 0 || modulo == "" || registroID <= 0 || estado == "" || estado == "todos" {
		return errors.New("empresa_id, modulo, registro_id y estado son requeridos")
	}
	old := ""
	if err := QueryRowCompat(dbConn, `SELECT COALESCE(estado,'') FROM empresa_modulos_colombia_registros WHERE empresa_id=? AND modulo=? AND id=?`, empresaID, modulo, registroID).Scan(&old); err != nil {
		return err
	}
	if _, err := ExecCompat(dbConn, `UPDATE empresa_modulos_colombia_registros SET estado=?,fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND modulo=? AND id=?`, estado, empresaID, modulo, registroID); err != nil {
		return err
	}
	if strings.TrimSpace(detalle) == "" {
		detalle = "Cambio de estado"
	}
	return RegistrarEmpresaModuloColombiaEvento(dbConn, empresaID, modulo, registroID, "estado_actualizado", old, estado, detalle, usuario)
}

func AplicarEmpresaModuloColombiaAccionMasiva(dbConn *sql.DB, empresaID int64, modulo string, accion EmpresaModuloColombiaAccionMasiva, usuario string) (EmpresaModuloColombiaAccionMasivaResult, error) {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return EmpresaModuloColombiaAccionMasivaResult{}, err
	}
	modulo = NormalizeEmpresaModuloColombia(modulo)
	accion = normalizeEmpresaModuloColombiaAccionMasiva(accion)
	if empresaID <= 0 || modulo == "" {
		return EmpresaModuloColombiaAccionMasivaResult{}, errors.New("empresa_id y modulo son requeridos")
	}
	result := EmpresaModuloColombiaAccionMasivaResult{Total: len(accion.RegistroIDs), Detalles: []string{}}
	if len(accion.RegistroIDs) == 0 {
		return result, errors.New("selecciona al menos un registro")
	}
	if len(accion.RegistroIDs) > 200 {
		return result, errors.New("maximo 200 registros por accion masiva")
	}
	if accion.Estado == "" && accion.Prioridad == "" && accion.Responsable == "" {
		return result, errors.New("indica estado, prioridad o responsable para actualizar")
	}
	for _, registroID := range uniqueModuloColombiaIDs(accion.RegistroIDs) {
		if registroID <= 0 {
			result.Omitidos++
			continue
		}
		var oldEstado string
		if err := QueryRowCompat(dbConn, `SELECT COALESCE(estado,'') FROM empresa_modulos_colombia_registros WHERE empresa_id=? AND modulo=? AND id=?`, empresaID, modulo, registroID).Scan(&oldEstado); err != nil {
			result.Omitidos++
			result.Detalles = append(result.Detalles, fmt.Sprintf("registro %d no encontrado", registroID))
			continue
		}
		setParts := []string{"fecha_actualizacion=CURRENT_TIMESTAMP"}
		args := []interface{}{}
		if accion.Estado != "" {
			setParts = append(setParts, "estado=?")
			args = append(args, accion.Estado)
		}
		if accion.Prioridad != "" {
			setParts = append(setParts, "prioridad=?")
			args = append(args, accion.Prioridad)
		}
		if accion.Responsable != "" {
			setParts = append(setParts, "responsable=?")
			args = append(args, accion.Responsable)
		}
		args = append(args, empresaID, modulo, registroID)
		res, err := ExecCompat(dbConn, fmt.Sprintf(`UPDATE empresa_modulos_colombia_registros SET %s WHERE empresa_id=? AND modulo=? AND id=?`, strings.Join(setParts, ",")), args...)
		if err != nil {
			result.Omitidos++
			result.Detalles = append(result.Detalles, fmt.Sprintf("registro %d: %v", registroID, err))
			continue
		}
		if n, _ := res.RowsAffected(); n == 0 {
			result.Omitidos++
			continue
		}
		result.Actualizados++
		nuevoEstado := oldEstado
		if accion.Estado != "" {
			nuevoEstado = accion.Estado
		}
		detalle := accion.Detalle
		if detalle == "" {
			detalle = "Accion masiva aplicada"
		}
		_ = RegistrarEmpresaModuloColombiaEvento(dbConn, empresaID, modulo, registroID, "accion_masiva", oldEstado, nuevoEstado, detalleAccionMasivaModuloColombia(accion, detalle), usuario)
	}
	result.Total = len(uniqueModuloColombiaIDs(accion.RegistroIDs))
	result.Omitidos = result.Total - result.Actualizados
	return result, nil
}

func CerrarEmpresaModuloColombiaRegistroControlado(dbConn *sql.DB, empresaID int64, modulo string, registroID int64, detalle, usuario string) error {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return err
	}
	modulo = NormalizeEmpresaModuloColombia(modulo)
	detalle = strings.TrimSpace(detalle)
	usuario = strings.TrimSpace(usuario)
	if empresaID <= 0 || modulo == "" || registroID <= 0 {
		return errors.New("empresa_id, modulo y registro_id son requeridos")
	}
	row, err := GetEmpresaModuloColombiaRegistro(dbConn, empresaID, modulo, registroID)
	if err != nil {
		return err
	}
	if isModuloColombiaEstadoFinal(row.Estado) {
		return errors.New("el registro ya esta en estado final")
	}
	evidencias, _ := ListEmpresaModuloColombiaEvidencias(dbConn, empresaID, modulo, registroID, 1)
	if len(evidencias) == 0 {
		return errors.New("no se puede cerrar sin al menos una evidencia o soporte")
	}
	aprobaciones, _ := ListEmpresaModuloColombiaAprobaciones(dbConn, empresaID, modulo, registroID, "pendiente", 50)
	if len(aprobaciones) > 0 {
		return fmt.Errorf("no se puede cerrar con %d aprobacion(es) pendiente(s)", len(aprobaciones))
	}
	tareas, _ := ListEmpresaModuloColombiaTareas(dbConn, empresaID, modulo, registroID, "", 100)
	abiertas := countModuloColombiaTareasAbiertas(tareas)
	if abiertas > 0 {
		return fmt.Errorf("no se puede cerrar con %d tarea(s) abierta(s)", abiertas)
	}
	if _, err := ExecCompat(dbConn, `UPDATE empresa_modulos_colombia_registros SET estado='cerrado',fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND modulo=? AND id=?`, empresaID, modulo, registroID); err != nil {
		return err
	}
	if detalle == "" {
		detalle = "Cierre controlado validado"
	}
	return RegistrarEmpresaModuloColombiaEvento(dbConn, empresaID, modulo, registroID, "cierre_controlado", row.Estado, "cerrado", detalle, usuario)
}

func RegistrarEmpresaModuloColombiaEvento(dbConn *sql.DB, empresaID int64, modulo string, registroID int64, evento, old, nuevo, detalle, usuario string) error {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return err
	}
	modulo = NormalizeEmpresaModuloColombia(modulo)
	evento = normalizeModuloColombiaEvento(evento)
	if empresaID <= 0 || modulo == "" || evento == "" {
		return errors.New("empresa_id, modulo y evento son requeridos")
	}
	return registrarEmpresaModuloColombiaEvento(dbConn, empresaID, modulo, registroID, evento, old, nuevo, detalle, usuario)
}

func RegistrarEmpresaModuloColombiaEvidencia(dbConn *sql.DB, row EmpresaModuloColombiaEvidencia) (int64, error) {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return 0, err
	}
	row.Modulo = NormalizeEmpresaModuloColombia(row.Modulo)
	row.Tipo = normalizeModuloColombiaEvento(row.Tipo)
	row.Nombre = strings.TrimSpace(row.Nombre)
	row.URL = sanitizeModuloColombiaURL(row.URL)
	row.Descripcion = strings.TrimSpace(row.Descripcion)
	row.Usuario = strings.TrimSpace(row.Usuario)
	if row.EmpresaID <= 0 || row.Modulo == "" || row.RegistroID <= 0 || row.Nombre == "" {
		return 0, errors.New("empresa_id, modulo, registro_id y nombre son requeridos")
	}
	if err := validarEmpresaModuloColombiaRegistroExiste(dbConn, row.EmpresaID, row.Modulo, row.RegistroID); err != nil {
		return 0, err
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_modulos_colombia_evidencias (empresa_id,modulo,registro_id,tipo,nombre,url,descripcion,usuario) VALUES (?,?,?,?,?,?,?,?)`, row.EmpresaID, row.Modulo, row.RegistroID, row.Tipo, row.Nombre, row.URL, row.Descripcion, row.Usuario)
	if err != nil {
		return 0, err
	}
	_ = RegistrarEmpresaModuloColombiaEvento(dbConn, row.EmpresaID, row.Modulo, row.RegistroID, "evidencia_agregada", "", "", row.Nombre, row.Usuario)
	return id, nil
}

func SolicitarEmpresaModuloColombiaAprobacion(dbConn *sql.DB, row EmpresaModuloColombiaAprobacion) (int64, error) {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return 0, err
	}
	row.Modulo = NormalizeEmpresaModuloColombia(row.Modulo)
	row.Nivel = normalizeModuloColombiaNivelAprobacion(row.Nivel)
	row.SolicitadoA = strings.TrimSpace(row.SolicitadoA)
	row.SolicitadoPor = strings.TrimSpace(row.SolicitadoPor)
	row.Estado = "pendiente"
	row.Comentario = strings.TrimSpace(row.Comentario)
	row.FechaVencimiento = strings.TrimSpace(row.FechaVencimiento)
	if row.EmpresaID <= 0 || row.Modulo == "" || row.RegistroID <= 0 || row.SolicitadoA == "" {
		return 0, errors.New("empresa_id, modulo, registro_id y solicitado_a son requeridos")
	}
	if err := validarEmpresaModuloColombiaRegistroExiste(dbConn, row.EmpresaID, row.Modulo, row.RegistroID); err != nil {
		return 0, err
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_modulos_colombia_aprobaciones (empresa_id,modulo,registro_id,nivel,solicitado_a,solicitado_por,estado,comentario,fecha_vencimiento) VALUES (?,?,?,?,?,?,?,?,?)`, row.EmpresaID, row.Modulo, row.RegistroID, row.Nivel, row.SolicitadoA, row.SolicitadoPor, row.Estado, row.Comentario, row.FechaVencimiento)
	if err != nil {
		return 0, err
	}
	_ = RegistrarEmpresaModuloColombiaEvento(dbConn, row.EmpresaID, row.Modulo, row.RegistroID, "aprobacion_solicitada", "", "pendiente", fmt.Sprintf("%s -> %s", row.Nivel, row.SolicitadoA), row.SolicitadoPor)
	return id, nil
}

func DecidirEmpresaModuloColombiaAprobacion(dbConn *sql.DB, empresaID int64, modulo string, aprobacionID int64, decision, comentario, usuario string) error {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return err
	}
	modulo = NormalizeEmpresaModuloColombia(modulo)
	decision = normalizeModuloColombiaAprobacionEstado(decision)
	comentario = strings.TrimSpace(comentario)
	usuario = strings.TrimSpace(usuario)
	if empresaID <= 0 || modulo == "" || aprobacionID <= 0 || (decision != "aprobado" && decision != "rechazado") {
		return errors.New("empresa_id, modulo, aprobacion_id y decision aprobado/rechazado son requeridos")
	}
	var registroID int64
	var old string
	if err := QueryRowCompat(dbConn, `SELECT COALESCE(registro_id,0),COALESCE(estado,'pendiente') FROM empresa_modulos_colombia_aprobaciones WHERE empresa_id=? AND modulo=? AND id=?`, empresaID, modulo, aprobacionID).Scan(&registroID, &old); err != nil {
		return err
	}
	if _, err := ExecCompat(dbConn, `UPDATE empresa_modulos_colombia_aprobaciones SET estado=?,comentario=?,decision_por=?,fecha_decision=CURRENT_TIMESTAMP WHERE empresa_id=? AND modulo=? AND id=?`, decision, comentario, usuario, empresaID, modulo, aprobacionID); err != nil {
		return err
	}
	if decision == "aprobado" {
		_, _ = ExecCompat(dbConn, `UPDATE empresa_modulos_colombia_registros SET estado='aprobado',fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND modulo=? AND id=? AND estado NOT IN ('cerrado','cancelado','cumplido','pagado','resuelto')`, empresaID, modulo, registroID)
	}
	return RegistrarEmpresaModuloColombiaEvento(dbConn, empresaID, modulo, registroID, "aprobacion_decidida", old, decision, comentario, usuario)
}

func CrearEmpresaModuloColombiaTarea(dbConn *sql.DB, row EmpresaModuloColombiaTarea) (int64, error) {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return 0, err
	}
	row.Modulo = NormalizeEmpresaModuloColombia(row.Modulo)
	row.Titulo = strings.TrimSpace(row.Titulo)
	row.Responsable = strings.TrimSpace(row.Responsable)
	row.Prioridad = normalizeModuloColombiaPrioridad(row.Prioridad)
	row.Estado = normalizeModuloColombiaTareaEstado(row.Estado)
	row.FechaVencimiento = strings.TrimSpace(row.FechaVencimiento)
	row.Comentario = strings.TrimSpace(row.Comentario)
	row.UsuarioCreador = strings.TrimSpace(row.UsuarioCreador)
	if row.EmpresaID <= 0 || row.Modulo == "" || row.RegistroID <= 0 || row.Titulo == "" {
		return 0, errors.New("empresa_id, modulo, registro_id y titulo son requeridos")
	}
	if row.Estado == "" || row.Estado == "todos" {
		row.Estado = "pendiente"
	}
	if err := validarEmpresaModuloColombiaRegistroExiste(dbConn, row.EmpresaID, row.Modulo, row.RegistroID); err != nil {
		return 0, err
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_modulos_colombia_tareas (empresa_id,modulo,registro_id,titulo,responsable,prioridad,estado,fecha_vencimiento,comentario,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?)`, row.EmpresaID, row.Modulo, row.RegistroID, row.Titulo, row.Responsable, row.Prioridad, row.Estado, row.FechaVencimiento, row.Comentario, row.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	_ = RegistrarEmpresaModuloColombiaEvento(dbConn, row.EmpresaID, row.Modulo, row.RegistroID, "tarea_creada", "", row.Estado, row.Titulo, row.UsuarioCreador)
	return id, nil
}

func CambiarEstadoEmpresaModuloColombiaTarea(dbConn *sql.DB, empresaID int64, modulo string, tareaID int64, estado, comentario, usuario string) error {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return err
	}
	modulo = NormalizeEmpresaModuloColombia(modulo)
	estado = normalizeModuloColombiaTareaEstado(estado)
	comentario = strings.TrimSpace(comentario)
	usuario = strings.TrimSpace(usuario)
	if empresaID <= 0 || modulo == "" || tareaID <= 0 || estado == "" || estado == "todos" {
		return errors.New("empresa_id, modulo, tarea_id y estado son requeridos")
	}
	var registroID int64
	var old string
	var titulo string
	if err := QueryRowCompat(dbConn, `SELECT COALESCE(registro_id,0),COALESCE(estado,'pendiente'),titulo FROM empresa_modulos_colombia_tareas WHERE empresa_id=? AND modulo=? AND id=?`, empresaID, modulo, tareaID).Scan(&registroID, &old, &titulo); err != nil {
		return err
	}
	if _, err := ExecCompat(dbConn, `UPDATE empresa_modulos_colombia_tareas SET estado=?,comentario=?,fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND modulo=? AND id=?`, estado, comentario, empresaID, modulo, tareaID); err != nil {
		return err
	}
	detalle := strings.TrimSpace(comentario)
	if detalle == "" {
		detalle = titulo
	}
	return RegistrarEmpresaModuloColombiaEvento(dbConn, empresaID, modulo, registroID, "tarea_actualizada", old, estado, detalle, usuario)
}

func SeedEmpresaModuloColombiaDemo(dbConn *sql.DB, empresaID int64, modulo, usuario string) error {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return err
	}
	modulo = NormalizeEmpresaModuloColombia(modulo)
	rows := demoEmpresaModuloColombiaRows(empresaID, modulo, usuario)
	for _, row := range rows {
		if _, err := UpsertEmpresaModuloColombiaRegistro(dbConn, row); err != nil {
			return err
		}
	}
	return nil
}

func ImportEmpresaModuloColombiaRegistros(dbConn *sql.DB, empresaID int64, modulo string, rows []EmpresaModuloColombiaRegistro, usuario string) (EmpresaModuloColombiaImportResult, error) {
	if err := EnsureEmpresaModulosColombiaSchema(dbConn); err != nil {
		return EmpresaModuloColombiaImportResult{}, err
	}
	modulo = NormalizeEmpresaModuloColombia(modulo)
	if empresaID <= 0 || modulo == "" {
		return EmpresaModuloColombiaImportResult{}, errors.New("empresa_id y modulo son requeridos")
	}
	result := EmpresaModuloColombiaImportResult{Total: len(rows), Errores: []string{}}
	if len(rows) > 1000 {
		return result, errors.New("maximo 1000 registros por importacion")
	}
	for i, row := range rows {
		row.EmpresaID = empresaID
		row.Modulo = modulo
		row.UsuarioCreador = usuario
		if strings.TrimSpace(row.Codigo) == "" || strings.TrimSpace(row.Nombre) == "" {
			result.Errores = append(result.Errores, fmt.Sprintf("fila %d: codigo y nombre son requeridos", i+1))
			continue
		}
		if _, err := UpsertEmpresaModuloColombiaRegistro(dbConn, row); err != nil {
			result.Errores = append(result.Errores, fmt.Sprintf("fila %d: %v", i+1, err))
			continue
		}
		result.Guardados++
	}
	_ = RegistrarEmpresaModuloColombiaEvento(dbConn, empresaID, modulo, 0, "importacion_masiva", "", "", fmt.Sprintf("Importacion masiva: %d/%d registros guardados", result.Guardados, result.Total), usuario)
	return result, nil
}

func listEmpresaModuloColombiaMetricas(dbConn *sql.DB, empresaID int64, modulo, campo string, limit int) ([]EmpresaModuloColombiaMetrica, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	col := ""
	switch campo {
	case "estado":
		col = "estado"
	case "tipo":
		col = "tipo"
	case "categoria":
		col = "categoria"
	case "prioridad":
		col = "prioridad"
	default:
		return nil, errors.New("campo de metrica no soportado")
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT COALESCE(NULLIF(TRIM(%[1]s),''),'sin_dato') AS clave, COUNT(1), COALESCE(SUM(valor),0) FROM empresa_modulos_colombia_registros WHERE empresa_id=? AND modulo=? GROUP BY COALESCE(NULLIF(TRIM(%[1]s),''),'sin_dato') ORDER BY COUNT(1) DESC, clave ASC LIMIT %d`, col, limit), empresaID, modulo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaModuloColombiaMetrica{}
	for rows.Next() {
		var x EmpresaModuloColombiaMetrica
		if err := rows.Scan(&x.Clave, &x.Total, &x.Valor); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func exportacionModuloColombiaRegistros(rows []EmpresaModuloColombiaRegistro) EmpresaModuloColombiaCSVSeccion {
	out := EmpresaModuloColombiaCSVSeccion{Nombre: "registros", Headers: []string{"id", "codigo", "tipo", "nombre", "tercero", "responsable", "categoria", "referencia", "prioridad", "estado", "fecha", "fecha_vencimiento", "valor", "usuario_creador", "fecha_creacion", "fecha_actualizacion"}}
	for _, row := range rows {
		out.Rows = append(out.Rows, []string{fmt.Sprintf("%d", row.ID), row.Codigo, row.Tipo, row.Nombre, row.Tercero, row.Responsable, row.Categoria, row.Referencia, row.Prioridad, row.Estado, row.Fecha, row.FechaVencimiento, formatFloatModuloColombia(row.Valor), row.UsuarioCreador, row.FechaCreacion, row.FechaActualizacion})
	}
	return out
}

func exportacionModuloColombiaAgenda(rows []EmpresaModuloColombiaAgendaItem) EmpresaModuloColombiaCSVSeccion {
	out := EmpresaModuloColombiaCSVSeccion{Nombre: "agenda_alertas", Headers: []string{"tipo", "registro_id", "referencia_id", "codigo", "titulo", "responsable", "estado", "prioridad", "fecha_vencimiento", "severidad", "detalle"}}
	for _, row := range rows {
		out.Rows = append(out.Rows, []string{row.Tipo, fmt.Sprintf("%d", row.RegistroID), fmt.Sprintf("%d", row.ReferenciaID), row.Codigo, row.Titulo, row.Responsable, row.Estado, row.Prioridad, row.FechaVencimiento, row.Severidad, row.Detalle})
	}
	return out
}

func exportacionModuloColombiaResponsables(rows []EmpresaModuloColombiaResponsableResumen) EmpresaModuloColombiaCSVSeccion {
	out := EmpresaModuloColombiaCSVSeccion{Nombre: "responsables_carga", Headers: []string{"responsable", "registros_abiertos", "registros_vencidos", "tareas_abiertas", "tareas_vencidas", "aprobaciones_pendientes", "total_pendiente", "recomendacion"}}
	for _, row := range rows {
		out.Rows = append(out.Rows, []string{row.Responsable, fmt.Sprintf("%d", row.RegistrosAbiertos), fmt.Sprintf("%d", row.RegistrosVencidos), fmt.Sprintf("%d", row.TareasAbiertas), fmt.Sprintf("%d", row.TareasVencidas), fmt.Sprintf("%d", row.AprobacionesPendientes), fmt.Sprintf("%d", row.TotalPendiente), row.Recomendacion})
	}
	return out
}

func exportacionModuloColombiaMetricas(nombre string, rows []EmpresaModuloColombiaMetrica) EmpresaModuloColombiaCSVSeccion {
	out := EmpresaModuloColombiaCSVSeccion{Nombre: nombre, Headers: []string{"clave", "total", "valor"}}
	for _, row := range rows {
		out.Rows = append(out.Rows, []string{row.Clave, fmt.Sprintf("%d", row.Total), formatFloatModuloColombia(row.Valor)})
	}
	return out
}

func exportacionModuloColombiaSLA(row EmpresaModuloColombiaSLA) EmpresaModuloColombiaCSVSeccion {
	return EmpresaModuloColombiaCSVSeccion{
		Nombre:  "sla_cumplimiento",
		Headers: []string{"fecha_corte", "total_abiertos", "vencidos", "proximos_7_dias", "sin_vencimiento", "tareas_abiertas", "tareas_vencidas", "cumplimiento_pct", "semaforo", "bucket_sin_vencimiento", "bucket_vencido", "bucket_0_7", "bucket_8_30", "bucket_mas_30"},
		Rows: [][]string{{
			row.FechaCorte,
			fmt.Sprintf("%d", row.TotalAbiertos),
			fmt.Sprintf("%d", row.Vencidos),
			fmt.Sprintf("%d", row.Proximos7Dias),
			fmt.Sprintf("%d", row.SinVencimiento),
			fmt.Sprintf("%d", row.TareasAbiertas),
			fmt.Sprintf("%d", row.TareasVencidas),
			formatFloatModuloColombia(row.CumplimientoPct),
			row.Semaforo,
			fmt.Sprintf("%d", row.Buckets["sin_vencimiento"]),
			fmt.Sprintf("%d", row.Buckets["vencido"]),
			fmt.Sprintf("%d", row.Buckets["0_7"]),
			fmt.Sprintf("%d", row.Buckets["8_30"]),
			fmt.Sprintf("%d", row.Buckets["mas_30"]),
		}},
	}
}

func exportacionModuloColombiaRiesgo(row EmpresaModuloColombiaRiesgo) EmpresaModuloColombiaCSVSeccion {
	return EmpresaModuloColombiaCSVSeccion{
		Nombre:  "matriz_riesgo",
		Headers: []string{"score", "nivel", "registros_vencidos", "criticos_abiertos", "sin_responsable", "sin_evidencia", "aprobaciones_pendientes", "tareas_abiertas", "tareas_vencidas", "factores"},
		Rows: [][]string{{
			fmt.Sprintf("%d", row.Score),
			row.Nivel,
			fmt.Sprintf("%d", row.RegistrosVencidos),
			fmt.Sprintf("%d", row.CriticosAbiertos),
			fmt.Sprintf("%d", row.SinResponsable),
			fmt.Sprintf("%d", row.SinEvidencia),
			fmt.Sprintf("%d", row.AprobacionesPendientes),
			fmt.Sprintf("%d", row.TareasAbiertas),
			fmt.Sprintf("%d", row.TareasVencidas),
			strings.Join(row.Factores, " | "),
		}},
	}
}

func exportacionModuloColombiaTareas(rows []EmpresaModuloColombiaTarea) EmpresaModuloColombiaCSVSeccion {
	out := EmpresaModuloColombiaCSVSeccion{Nombre: "tareas", Headers: []string{"id", "registro_id", "titulo", "responsable", "prioridad", "estado", "fecha_vencimiento", "comentario", "usuario_creador", "fecha_creacion", "fecha_actualizacion"}}
	for _, row := range rows {
		out.Rows = append(out.Rows, []string{fmt.Sprintf("%d", row.ID), fmt.Sprintf("%d", row.RegistroID), row.Titulo, row.Responsable, row.Prioridad, row.Estado, row.FechaVencimiento, row.Comentario, row.UsuarioCreador, row.FechaCreacion, row.FechaActualizacion})
	}
	return out
}

func exportacionModuloColombiaAprobaciones(rows []EmpresaModuloColombiaAprobacion) EmpresaModuloColombiaCSVSeccion {
	out := EmpresaModuloColombiaCSVSeccion{Nombre: "aprobaciones", Headers: []string{"id", "registro_id", "nivel", "solicitado_a", "solicitado_por", "estado", "comentario", "decision_por", "fecha_decision", "fecha_vencimiento", "fecha_creacion"}}
	for _, row := range rows {
		out.Rows = append(out.Rows, []string{fmt.Sprintf("%d", row.ID), fmt.Sprintf("%d", row.RegistroID), row.Nivel, row.SolicitadoA, row.SolicitadoPor, row.Estado, row.Comentario, row.DecisionPor, row.FechaDecision, row.FechaVencimiento, row.FechaCreacion})
	}
	return out
}

func exportacionModuloColombiaEvidencias(rows []EmpresaModuloColombiaEvidencia) EmpresaModuloColombiaCSVSeccion {
	out := EmpresaModuloColombiaCSVSeccion{Nombre: "evidencias", Headers: []string{"id", "registro_id", "tipo", "nombre", "url", "descripcion", "usuario", "fecha_creacion"}}
	for _, row := range rows {
		out.Rows = append(out.Rows, []string{fmt.Sprintf("%d", row.ID), fmt.Sprintf("%d", row.RegistroID), row.Tipo, row.Nombre, row.URL, row.Descripcion, row.Usuario, row.FechaCreacion})
	}
	return out
}

func exportacionModuloColombiaEventos(rows []EmpresaModuloColombiaEvento) EmpresaModuloColombiaCSVSeccion {
	out := EmpresaModuloColombiaCSVSeccion{Nombre: "bitacora", Headers: []string{"id", "registro_id", "evento", "estado_anterior", "estado_nuevo", "detalle", "usuario", "fecha_creacion"}}
	for _, row := range rows {
		out.Rows = append(out.Rows, []string{fmt.Sprintf("%d", row.ID), fmt.Sprintf("%d", row.RegistroID), row.Evento, row.EstadoAnterior, row.EstadoNuevo, row.Detalle, row.Usuario, row.FechaCreacion})
	}
	return out
}

func exportacionModuloColombiaPrefijo(origen string, rows []string) [][]string {
	out := [][]string{}
	for _, row := range rows {
		if strings.TrimSpace(row) == "" {
			continue
		}
		out = append(out, []string{origen, row})
	}
	return out
}

func exportacionModuloColombiaLista(nombre string, headers []string, rows [][]string) EmpresaModuloColombiaCSVSeccion {
	return EmpresaModuloColombiaCSVSeccion{Nombre: nombre, Headers: headers, Rows: rows}
}

func formatFloatModuloColombia(v float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", v), "0"), ".")
}

func recomendacionesModuloColombia(r EmpresaModuloColombiaReporte) []string {
	out := []string{}
	if r.Vencidos > 0 {
		out = append(out, fmt.Sprintf("Priorizar %d registro(s) vencido(s) y dejar evidencia de gestion.", r.Vencidos))
	}
	if r.Vencen7Dias > 0 {
		out = append(out, fmt.Sprintf("Programar seguimiento para %d registro(s) que vencen en los proximos 7 dias.", r.Vencen7Dias))
	}
	if r.CriticosAbiertos > 0 {
		out = append(out, fmt.Sprintf("Escalar %d registro(s) critico(s) o urgente(s) que siguen abiertos.", r.CriticosAbiertos))
	}
	if r.SinResponsable > 0 {
		out = append(out, fmt.Sprintf("Asignar responsable a %d registro(s) para mantener trazabilidad.", r.SinResponsable))
	}
	if len(out) == 0 {
		out = append(out, "Indicadores dentro de parametros; mantener seguimiento periodico.")
	}
	return out
}

func recomendacionesModuloColombiaAgenda(a EmpresaModuloColombiaAgenda) []string {
	out := []string{}
	if a.RegistrosVencidos+a.TareasVencidas+a.AprobacionesVencidas > 0 {
		out = append(out, "Atender primero los vencidos criticos y dejar evidencia del cierre.")
	}
	if a.AprobacionesPendientes > 0 {
		out = append(out, fmt.Sprintf("Resolver %d aprobacion(es) pendiente(s) para desbloquear el flujo.", a.AprobacionesPendientes))
	}
	if a.TareasProximas > 0 || a.RegistrosProximos > 0 {
		out = append(out, "Programar responsables para los vencimientos de los proximos 7 dias.")
	}
	if len(out) == 0 {
		out = append(out, "Agenda sin alertas criticas; mantener revision periodica.")
	}
	return out
}

func moduloColombiaSeveridadRank(v string) int {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "critica":
		return 0
	case "alta":
		return 1
	case "media":
		return 2
	default:
		return 3
	}
}

func getModuloColombiaResponsableResumen(rows map[string]*EmpresaModuloColombiaResponsableResumen, responsable string) *EmpresaModuloColombiaResponsableResumen {
	key := strings.TrimSpace(responsable)
	if key == "" {
		key = "Sin responsable"
	}
	if rows[key] == nil {
		rows[key] = &EmpresaModuloColombiaResponsableResumen{Responsable: key}
	}
	return rows[key]
}

func recomendacionModuloColombiaResponsable(row EmpresaModuloColombiaResponsableResumen) string {
	if row.RegistrosVencidos+row.TareasVencidas > 0 {
		return "Priorizar vencidos antes de tomar nuevos compromisos."
	}
	if row.AprobacionesPendientes > 0 {
		return "Resolver aprobaciones pendientes para desbloquear el flujo."
	}
	if row.TotalPendiente > 8 {
		return "Rebalancear carga o reasignar tareas."
	}
	return "Carga operativa controlada."
}

func bucketModuloColombiaVencimiento(fecha string, today time.Time) string {
	fecha = strings.TrimSpace(fecha)
	if fecha == "" {
		return "sin_vencimiento"
	}
	due, err := time.Parse("2006-01-02", fecha)
	if err != nil {
		return "sin_vencimiento"
	}
	t0, _ := time.Parse("2006-01-02", today.Format("2006-01-02"))
	days := int(due.Sub(t0).Hours() / 24)
	switch {
	case days < 0:
		return "vencido"
	case days <= 7:
		return "0_7"
	case days <= 30:
		return "8_30"
	default:
		return "mas_30"
	}
}

func calcularCumplimientoModuloColombia(total int, vencidos int) float64 {
	if total <= 0 {
		return 100
	}
	ok := total - vencidos
	if ok < 0 {
		ok = 0
	}
	return float64(ok) * 100 / float64(total)
}

func semaforoModuloColombiaSLA(cumplimiento float64, vencidos int) string {
	if vencidos > 0 || cumplimiento < 80 {
		return "rojo"
	}
	if cumplimiento < 95 {
		return "amarillo"
	}
	return "verde"
}

func recomendacionesModuloColombiaSLA(sla EmpresaModuloColombiaSLA) []string {
	out := []string{}
	if sla.Vencidos+sla.TareasVencidas > 0 {
		out = append(out, "Activar plan de accion para vencidos y tareas atrasadas.")
	}
	if sla.SinVencimiento > 0 {
		out = append(out, "Asignar fecha de vencimiento a registros sin SLA definido.")
	}
	if sla.Proximos7Dias > 0 {
		out = append(out, "Revisar compromisos que vencen esta semana.")
	}
	if len(out) == 0 {
		out = append(out, "SLA controlado; mantener monitoreo periodico.")
	}
	return out
}

func maxIntModuloColombia(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func scoreModuloColombiaRiesgo(r EmpresaModuloColombiaRiesgo) (int, []string) {
	score := 0
	factores := []string{}
	add := func(points int, label string) {
		if points <= 0 {
			return
		}
		score += points
		factores = append(factores, label)
	}
	add(minIntModuloColombia(r.RegistrosVencidos*12, 30), fmt.Sprintf("%d registro(s) vencido(s)", r.RegistrosVencidos))
	add(minIntModuloColombia(r.CriticosAbiertos*10, 20), fmt.Sprintf("%d critico(s) abierto(s)", r.CriticosAbiertos))
	add(minIntModuloColombia(r.AprobacionesPendientes*6, 18), fmt.Sprintf("%d aprobacion(es) pendiente(s)", r.AprobacionesPendientes))
	add(minIntModuloColombia(r.TareasVencidas*8, 16), fmt.Sprintf("%d tarea(s) vencida(s)", r.TareasVencidas))
	add(minIntModuloColombia(r.TareasAbiertas*2, 10), fmt.Sprintf("%d tarea(s) abierta(s)", r.TareasAbiertas))
	add(minIntModuloColombia(r.SinResponsable*4, 12), fmt.Sprintf("%d registro(s) sin responsable", r.SinResponsable))
	add(minIntModuloColombia(r.SinEvidencia*3, 12), fmt.Sprintf("%d registro(s) sin evidencia", r.SinEvidencia))
	if score > 100 {
		score = 100
	}
	if len(factores) == 0 {
		factores = append(factores, "Sin factores criticos detectados")
	}
	return score, factores
}

func nivelModuloColombiaRiesgo(score int) string {
	switch {
	case score >= 70:
		return "alto"
	case score >= 35:
		return "medio"
	default:
		return "bajo"
	}
}

func recomendacionesModuloColombiaRiesgo(r EmpresaModuloColombiaRiesgo) []string {
	out := []string{}
	if r.RegistrosVencidos > 0 || r.TareasVencidas > 0 {
		out = append(out, "Atender vencidos antes de crear nuevos compromisos.")
	}
	if r.AprobacionesPendientes > 0 {
		out = append(out, "Resolver aprobaciones pendientes para reducir bloqueo operativo.")
	}
	if r.SinEvidencia > 0 {
		out = append(out, "Adjuntar evidencias a los registros abiertos sin soporte.")
	}
	if r.SinResponsable > 0 {
		out = append(out, "Asignar responsables a registros abiertos sin propietario.")
	}
	if len(out) == 0 {
		out = append(out, "Riesgo operativo bajo; mantener seguimiento periodico.")
	}
	return out
}

func minIntModuloColombia(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func tituloPlanAccionModuloColombia(item EmpresaModuloColombiaAgendaItem) string {
	base := strings.TrimSpace(item.Titulo)
	if base == "" {
		base = "alerta operativa"
	}
	return "Plan de accion - " + strings.ReplaceAll(item.Tipo, "_", " ") + ": " + base
}

func prioridadPlanAccionModuloColombia(severidad string) string {
	switch strings.ToLower(strings.TrimSpace(severidad)) {
	case "critica":
		return "urgente"
	case "alta":
		return "alta"
	default:
		return "normal"
	}
}

func existeTareaAbiertaModuloColombia(dbConn *sql.DB, empresaID int64, modulo string, registroID int64, titulo string) bool {
	var total int
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_modulos_colombia_tareas WHERE empresa_id=? AND modulo=? AND registro_id=? AND titulo=? AND estado IN ('pendiente','en_proceso')`, empresaID, modulo, registroID, strings.TrimSpace(titulo)).Scan(&total)
	return total > 0
}

func countModuloColombiaAprobacionesEstado(rows []EmpresaModuloColombiaAprobacion, estado string) int {
	total := 0
	for _, row := range rows {
		if row.Estado == estado {
			total++
		}
	}
	return total
}

func countModuloColombiaTareasAbiertas(rows []EmpresaModuloColombiaTarea) int {
	total := 0
	for _, row := range rows {
		if row.Estado == "pendiente" || row.Estado == "en_proceso" {
			total++
		}
	}
	return total
}

func recomendacionEmpresaModuloColombiaExpediente(row EmpresaModuloColombiaRegistro, resumen map[string]int) string {
	if resumen["aprobaciones_pendientes"] > 0 {
		return "Tiene aprobaciones pendientes; conviene resolverlas antes del cierre."
	}
	if resumen["tareas_abiertas"] > 0 {
		return "Tiene tareas abiertas; asigna seguimiento hasta cumplirlas o cancelarlas."
	}
	if resumen["evidencias"] == 0 {
		return "No tiene evidencias; adjunta soportes para fortalecer la trazabilidad."
	}
	if row.FechaVencimiento != "" && row.FechaVencimiento < time.Now().Format("2006-01-02") && !isModuloColombiaEstadoFinal(row.Estado) {
		return "El registro esta vencido; prioriza cierre, evidencia o aprobacion."
	}
	return "Expediente con trazabilidad suficiente para seguimiento operativo."
}

func normalizeModuloColombiaEvento(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	repl := strings.NewReplacer(" ", "_", "-", "_", "/", "_", "\\", "_")
	v = repl.Replace(v)
	for strings.Contains(v, "__") {
		v = strings.ReplaceAll(v, "__", "_")
	}
	v = strings.Trim(v, "_")
	if v == "" {
		return "seguimiento"
	}
	return v
}

func NormalizeEmpresaModuloColombia(v string) string {
	clean := strings.ToLower(strings.TrimSpace(v))
	if _, ok := empresaModuloColombiaPlantillasPlantillas[clean]; ok {
		return clean
	}
	switch clean {
	case "bancos_pagos", "gestion_documental", "cumplimiento_kyc", "contratos_obligaciones", "calidad_procesos", "drogueria_farmacia":
		return clean
	default:
		return ""
	}
}

func normalizeEmpresaModuloColombiaRegistro(row EmpresaModuloColombiaRegistro) EmpresaModuloColombiaRegistro {
	row.Modulo = NormalizeEmpresaModuloColombia(row.Modulo)
	row.Codigo = normalizeModuloColombiaCodigo(row.Codigo)
	if row.Codigo == "" {
		row.Codigo = strings.ToUpper(row.Modulo) + "-" + time.Now().Format("20060102150405")
	}
	if strings.TrimSpace(row.Tipo) == "" {
		row.Tipo = "general"
	}
	row.Prioridad = normalizeModuloColombiaPrioridad(row.Prioridad)
	row.Estado = normalizeModuloColombiaEstado(row.Estado)
	if row.Fecha == "" {
		row.Fecha = time.Now().Format("2006-01-02")
	}
	if row.MetadataJSON == "" && len(row.Metadata) > 0 {
		if b, err := json.Marshal(row.Metadata); err == nil {
			row.MetadataJSON = string(b)
		}
	}
	if row.MetadataJSON == "" {
		row.MetadataJSON = "{}"
	}
	return row
}

func normalizeEmpresaModuloColombiaFiltro(f EmpresaModuloColombiaFiltro) EmpresaModuloColombiaFiltro {
	f.Texto = strings.ToLower(strings.TrimSpace(f.Texto))
	rawEstado := strings.TrimSpace(f.Estado)
	f.Estado = normalizeModuloColombiaEstado(f.Estado)
	if rawEstado == "" {
		f.Estado = ""
	}
	f.Tipo = normalizeModuloColombiaSimple(f.Tipo)
	f.Categoria = normalizeModuloColombiaSimple(f.Categoria)
	f.Prioridad = strings.TrimSpace(f.Prioridad)
	if f.Prioridad != "" {
		f.Prioridad = normalizeModuloColombiaPrioridad(f.Prioridad)
	}
	f.Responsable = strings.TrimSpace(f.Responsable)
	if f.ProximosDias < 0 {
		f.ProximosDias = 0
	}
	if f.ProximosDias > 365 {
		f.ProximosDias = 365
	}
	if f.Vencidos {
		f.ProximosDias = 0
	}
	return f
}

func normalizeEmpresaModuloColombiaAccionMasiva(a EmpresaModuloColombiaAccionMasiva) EmpresaModuloColombiaAccionMasiva {
	if strings.TrimSpace(a.Estado) != "" {
		a.Estado = normalizeModuloColombiaEstado(normalizeModuloColombiaSimple(a.Estado))
		if a.Estado == "todos" {
			a.Estado = ""
		}
	}
	if strings.TrimSpace(a.Prioridad) != "" {
		a.Prioridad = normalizeModuloColombiaPrioridad(a.Prioridad)
	}
	a.Responsable = strings.TrimSpace(a.Responsable)
	a.Detalle = strings.TrimSpace(a.Detalle)
	return a
}

func uniqueModuloColombiaIDs(ids []int64) []int64 {
	seen := map[int64]bool{}
	out := []int64{}
	for _, id := range ids {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

func detalleAccionMasivaModuloColombia(a EmpresaModuloColombiaAccionMasiva, detalle string) string {
	parts := []string{}
	if a.Estado != "" {
		parts = append(parts, "estado="+a.Estado)
	}
	if a.Prioridad != "" {
		parts = append(parts, "prioridad="+a.Prioridad)
	}
	if a.Responsable != "" {
		parts = append(parts, "responsable="+a.Responsable)
	}
	if detalle != "" {
		parts = append(parts, detalle)
	}
	return strings.Join(parts, " | ")
}

func normalizeModuloColombiaSimple(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	v = strings.NewReplacer(" ", "_", "-", "_", "/", "_", "\\", "_").Replace(v)
	for strings.Contains(v, "__") {
		v = strings.ReplaceAll(v, "__", "_")
	}
	return strings.Trim(v, "_")
}

func normalizeModuloColombiaCodigo(v string) string {
	v = strings.ToUpper(strings.TrimSpace(v))
	repl := strings.NewReplacer(" ", "-", "_", "-", "/", "-", "\\", "-", ".", "-")
	v = repl.Replace(v)
	for strings.Contains(v, "--") {
		v = strings.ReplaceAll(v, "--", "-")
	}
	return strings.Trim(v, "-")
}

func normalizeModuloColombiaPrioridad(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "baja", "normal", "alta", "critica", "urgente":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "normal"
	}
}

func normalizeModuloColombiaEstado(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "borrador", "pendiente", "abierto", "en_revision", "en_proceso", "en_gestion", "aprobado", "cerrado", "cumplido", "pagado", "resuelto", "rechazado", "cancelado", "todos":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "pendiente"
	}
}

func isModuloColombiaEstadoFinal(v string) bool {
	switch normalizeModuloColombiaEstado(v) {
	case "cerrado", "cancelado", "cumplido", "pagado", "resuelto", "rechazado":
		return true
	default:
		return false
	}
}

func normalizeModuloColombiaAprobacionEstado(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "pendiente", "aprobado", "rechazado", "cancelado", "todos":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "pendiente"
	}
}

func normalizeModuloColombiaTareaEstado(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "pendiente", "en_proceso", "cumplida", "cancelada", "todos":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "pendiente"
	}
}

func normalizeModuloColombiaNivelAprobacion(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "operativo", "supervisor", "contable", "gerencia", "cumplimiento", "juridico":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "operativo"
	}
}

func sanitizeModuloColombiaURL(v string) string {
	v = strings.TrimSpace(v)
	lower := strings.ToLower(v)
	if strings.HasPrefix(lower, "javascript:") || strings.HasPrefix(lower, "data:") {
		return ""
	}
	return v
}

func decodeModuloColombiaMetadata(raw string) map[string]interface{} {
	out := map[string]interface{}{}
	if strings.TrimSpace(raw) == "" {
		return out
	}
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func validarEmpresaModuloColombiaRegistroExiste(dbConn *sql.DB, empresaID int64, modulo string, registroID int64) error {
	var exists int
	if err := QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_modulos_colombia_registros WHERE empresa_id=? AND modulo=? AND id=?`, empresaID, modulo, registroID).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return errors.New("registro no existe para esta empresa y modulo")
	}
	return nil
}

func registrarEmpresaModuloColombiaEvento(dbConn *sql.DB, empresaID int64, modulo string, registroID int64, evento, old, nuevo, detalle, usuario string) error {
	_, err := insertSQLCompat(dbConn, `INSERT INTO empresa_modulos_colombia_eventos (empresa_id,modulo,registro_id,evento,estado_anterior,estado_nuevo,detalle,usuario) VALUES (?,?,?,?,?,?,?,?)`, empresaID, modulo, registroID, evento, old, nuevo, detalle, usuario)
	return err
}

func demoEmpresaModuloColombiaRows(empresaID int64, modulo, usuario string) []EmpresaModuloColombiaRegistro {
	today := time.Now()
	due := today.AddDate(0, 0, 15).Format("2006-01-02")
	base := EmpresaModuloColombiaRegistro{EmpresaID: empresaID, Modulo: modulo, Prioridad: "alta", Estado: "en_proceso", Fecha: today.Format("2006-01-02"), FechaVencimiento: due, UsuarioCreador: usuario}
	switch modulo {
	case "bancos_pagos":
		return []EmpresaModuloColombiaRegistro{
			mergeModuloDemo(base, "BANCO-001", "Conciliacion cuenta Bancolombia", "conciliacion", "Bancolombia", "Tesoreria", "extracto", "EXT-2026-05", 2450000, map[string]interface{}{"cuenta": "Ahorros 1234", "formato": "OFX/CSV", "reglas": "fecha, tercero, valor"}),
			mergeModuloDemo(base, "PAGO-001", "Lote pagos proveedores mayo", "pago_masivo", "Proveedores", "Tesoreria", "pagos", "LOTE-PROV-001", 5800000, map[string]interface{}{"banco": "Bancolombia", "archivo": "pendiente", "aprobadores": 2}),
		}
	case "gestion_documental":
		return []EmpresaModuloColombiaRegistro{
			mergeModuloDemo(base, "EXP-COMPRAS-001", "Expediente compras y soportes mayo", "expediente", "Proveedor demo", "Administracion", "compras", "EXP-2026-05", 0, map[string]interface{}{"retencion": "7 anos", "versiones": 3, "etiquetas": "compras,soportes"}),
			mergeModuloDemo(base, "APR-001", "Aprobacion contrato mantenimiento", "aprobacion", "Contratista", "Gerencia", "contratos", "CTR-MANT-001", 1200000, map[string]interface{}{"flujo": "area->gerencia->contabilidad"}),
		}
	case "cumplimiento_kyc":
		return []EmpresaModuloColombiaRegistro{
			mergeModuloDemo(base, "KYC-PROV-001", "Debida diligencia proveedor critico", "evaluacion", "Proveedor critico", "Cumplimiento", "proveedor", "NIT-900000001", 0, map[string]interface{}{"riesgo": "medio", "listas": "pendiente", "beneficiario_final": true}),
			mergeModuloDemo(base, "ALERTA-LAFT-001", "Alerta por cambio inusual de pagos", "alerta", "Cliente demo", "Oficial cumplimiento", "LAFT", "MOV-001", 3500000, map[string]interface{}{"senal": "valor superior al promedio", "decision": "revisar"}),
		}
	case "contratos_obligaciones":
		return []EmpresaModuloColombiaRegistro{
			mergeModuloDemo(base, "CTR-001", "Contrato de mantenimiento anual", "contrato", "Proveedor mantenimiento", "Administracion", "servicios", "POL-001", 9600000, map[string]interface{}{"firma": "pendiente", "renovacion": today.AddDate(1, 0, 0).Format("2006-01-02")}),
			mergeModuloDemo(base, "OBL-001", "Poliza de cumplimiento", "obligacion", "Aseguradora", "Gerencia", "poliza", "POL-001", 0, map[string]interface{}{"vigencia": due, "alerta_dias": 15}),
		}
	case "calidad_procesos":
		return []EmpresaModuloColombiaRegistro{
			mergeModuloDemo(base, "NC-001", "No conformidad en alistamiento de estacion", "no_conformidad", "Operacion", "Calidad", "servicio", "CHK-HAB-001", 0, map[string]interface{}{"accion_correctiva": "reforzar checklist", "responsable_cierre": "Supervisor"}),
			mergeModuloDemo(base, "AUD-001", "Auditoria interna proceso de caja", "auditoria", "Caja", "Auditor", "finanzas", "PROC-CAJA", 0, map[string]interface{}{"hallazgos": 2, "checklist": "cierre diario"}),
		}
	case "drogueria_farmacia":
		return []EmpresaModuloColombiaRegistro{
			mergeModuloDemo(base, "FARMA-LOT-001", "Control lote acetaminofen 500 mg", "lote", "Laboratorio demo", "Regente farmacia", "rx", "INVIMA-2026M-000001", 145000, map[string]interface{}{"lote": "ACET-0526", "vence": today.AddDate(0, 8, 0).Format("2006-01-02"), "stock": 120, "requiere_formula": false}),
			mergeModuloDemo(base, "FARMA-RX-001", "Validacion formula antibiotico", "formula_medica", "Paciente demo", "Auxiliar farmacia", "dispensacion", "FORM-001", 82000, map[string]interface{}{"medico": "Registro medico demo", "requiere_formula": true, "entrega_parcial": false}),
			mergeModuloDemo(base, "FARMA-CTRL-001", "Seguimiento medicamento controlado", "controlado", "Paciente controlado", "Director tecnico", "controlados", "REC-CTRL-001", 0, map[string]interface{}{"libro_control": "pendiente", "cantidad_dispensada": 1, "requiere_firma": true}),
		}
	default:
		plantilla := GetEmpresaModuloColombiaPlantilla(modulo)
		if plantilla.Titulo == "" || len(plantilla.Tipos) == 0 || len(plantilla.Categorias) == 0 {
			return nil
		}
		tipoPrincipal := plantilla.Tipos[0]
		tipoControl := plantilla.Tipos[len(plantilla.Tipos)-1]
		categoriaPrincipal := plantilla.Categorias[0]
		categoriaControl := plantilla.Categorias[len(plantilla.Categorias)-1]
		tercero := strings.TrimSpace(plantilla.EtiquetaTercero)
		if tercero == "" {
			tercero = "Cliente / area"
		}
		referencia := strings.TrimSpace(plantilla.EtiquetaReferencia)
		if referencia == "" {
			referencia = "Referencia"
		}
		prefix := strings.ToUpper(strings.ReplaceAll(modulo, "_", "-"))
		return []EmpresaModuloColombiaRegistro{
			mergeModuloDemo(base, prefix+"-001", plantilla.Titulo+" - operacion principal", tipoPrincipal, tercero+" demo", "Coordinador", categoriaPrincipal, referencia+" 001", 850000, map[string]interface{}{"origen": "demo", "modulo": modulo, "flujo": "registro->seguimiento->cierre"}),
			mergeModuloDemo(base, prefix+"-002", plantilla.Titulo+" - control y seguimiento", tipoControl, tercero+" control", "Supervisor", categoriaControl, referencia+" 002", 0, map[string]interface{}{"origen": "demo", "sla_dias": 3, "requiere_evidencia": true}),
		}
	}
}

func mergeModuloDemo(base EmpresaModuloColombiaRegistro, codigo, nombre, tipo, tercero, responsable, categoria, referencia string, valor float64, metadata map[string]interface{}) EmpresaModuloColombiaRegistro {
	base.Codigo = codigo
	base.Nombre = nombre
	base.Tipo = tipo
	base.Tercero = tercero
	base.Responsable = responsable
	base.Categoria = categoria
	base.Referencia = referencia
	base.Valor = valor
	base.Metadata = metadata
	return base
}
