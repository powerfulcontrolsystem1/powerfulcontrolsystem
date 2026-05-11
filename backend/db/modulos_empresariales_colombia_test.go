package db

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNormalizeEmpresaModuloColombia(t *testing.T) {
	if got := NormalizeEmpresaModuloColombia(" bancos_pagos "); got != "bancos_pagos" {
		t.Fatalf("NormalizeEmpresaModuloColombia = %q", got)
	}
	if got := NormalizeEmpresaModuloColombia(" drogueria_farmacia "); got != "drogueria_farmacia" {
		t.Fatalf("NormalizeEmpresaModuloColombia drogueria = %q", got)
	}
	if got := NormalizeEmpresaModuloColombia("modulo_inexistente"); got != "" {
		t.Fatalf("modulo desconocido deberia quedar vacio, got %q", got)
	}
}

func TestNormalizeModuloColombiaCodigo(t *testing.T) {
	if got := normalizeModuloColombiaCodigo(" doc/001 "); got != "DOC-001" {
		t.Fatalf("normalizeModuloColombiaCodigo = %q", got)
	}
}

func TestNormalizeEmpresaModuloColombiaRegistroDefaults(t *testing.T) {
	got := normalizeEmpresaModuloColombiaRegistro(EmpresaModuloColombiaRegistro{
		Modulo:    "helpdesk",
		Codigo:    " tck/001 ",
		Nombre:    "Ticket demo",
		Prioridad: "rara",
		Estado:    "desconocido",
		Metadata:  map[string]interface{}{"sla_horas": float64(8)},
	})
	if got.Codigo != "TCK-001" || got.Prioridad != "normal" || got.Estado != "pendiente" || got.Fecha == "" {
		t.Fatalf("defaults inesperados: %#v", got)
	}
	var meta map[string]interface{}
	if err := json.Unmarshal([]byte(got.MetadataJSON), &meta); err != nil {
		t.Fatalf("metadata json invalido: %v", err)
	}
	if meta["sla_horas"] != float64(8) {
		t.Fatalf("metadata inesperado: %#v", meta)
	}
}

func TestGetEmpresaModuloColombiaPlantillaPorModulo(t *testing.T) {
	got := GetEmpresaModuloColombiaPlantilla("bancos_pagos")
	if got.Modulo != "bancos_pagos" || got.Titulo == "" {
		t.Fatalf("plantilla basica inesperada: %#v", got)
	}
	if len(got.Tipos) == 0 || len(got.EstadosFlujo) == 0 || got.EtiquetaTercero == "" {
		t.Fatalf("plantilla incompleta: %#v", got)
	}
	if got.Tipos[0] != "conciliacion" {
		t.Fatalf("tipos inesperados para bancos: %#v", got.Tipos)
	}
	farma := GetEmpresaModuloColombiaPlantilla("drogueria_farmacia")
	if farma.Modulo != "drogueria_farmacia" || !strings.Contains(strings.Join(farma.Tipos, ","), "lote") {
		t.Fatalf("plantilla farmacia incompleta: %#v", farma)
	}
	if farma.EtiquetaReferencia == "" || !strings.Contains(farma.MetadataEjemplo, "registro_invima") {
		t.Fatalf("plantilla farmacia sin trazabilidad sanitaria: %#v", farma)
	}
	if !strings.Contains(strings.Join(farma.Tipos, ","), "formula_medica") || !strings.Contains(strings.Join(farma.Categorias, ","), "controlados") {
		t.Fatalf("plantilla farmacia debe ser expediente sanitario, no inventario paralelo: %#v", farma)
	}
}

func TestBuildEmpresaModuloColombiaDiagnostico(t *testing.T) {
	plantilla := GetEmpresaModuloColombiaPlantilla("agencia_viajes")
	got := buildEmpresaModuloColombiaDiagnostico(7, "agencia_viajes", plantilla, 0, true, "")
	if got.Estado != "listo" || got.Puntuacion != 100 {
		t.Fatalf("diagnostico listo inesperado: %#v", got)
	}
	if got.TotalObligatorios == 0 || got.OKObligatorios != got.TotalObligatorios {
		t.Fatalf("conteo de checks inesperado: %#v", got)
	}
	if len(got.Checks) < 7 {
		t.Fatalf("checks insuficientes: %#v", got.Checks)
	}
	foundInfo := false
	for _, check := range got.Checks {
		if check.Clave == "registros_operativos" {
			foundInfo = check.Informativo && !check.OK
		}
	}
	if !foundInfo {
		t.Fatalf("check informativo de registros inesperado: %#v", got.Checks)
	}
}

func TestBuildEmpresaModuloColombiaDiagnosticoDetectaFallos(t *testing.T) {
	plantilla := EmpresaModuloColombiaPlantilla{
		Modulo:          "helpdesk",
		Titulo:          "Helpdesk",
		Tipos:           []string{"ticket"},
		Categorias:      []string{"soporte"},
		EstadosFlujo:    []string{"abierto"},
		MetadataEjemplo: "{mal",
	}
	got := buildEmpresaModuloColombiaDiagnostico(0, "helpdesk", plantilla, 0, false, "sin conexion")
	if got.Estado != "revisar" || got.Puntuacion >= 100 {
		t.Fatalf("diagnostico debio requerir revision: %#v", got)
	}
	if len(got.Recomendaciones) == 0 {
		t.Fatalf("recomendaciones esperadas: %#v", got)
	}
}

func TestNormalizeModuloColombiaEvento(t *testing.T) {
	if got := normalizeModuloColombiaEvento(" Cambio Estado / Manual "); got != "cambio_estado_manual" {
		t.Fatalf("normalizeModuloColombiaEvento = %q", got)
	}
	if got := normalizeModuloColombiaEvento(""); got != "seguimiento" {
		t.Fatalf("evento vacio = %q", got)
	}
}

func TestRecomendacionesModuloColombiaReporte(t *testing.T) {
	got := recomendacionesModuloColombia(EmpresaModuloColombiaReporte{
		Vencidos:         1,
		Vencen7Dias:      2,
		CriticosAbiertos: 3,
		SinResponsable:   4,
	})
	if len(got) != 4 {
		t.Fatalf("recomendaciones esperadas 4, got %#v", got)
	}
	ok := recomendacionesModuloColombia(EmpresaModuloColombiaReporte{})
	if len(ok) != 1 || ok[0] == "" {
		t.Fatalf("recomendacion saludable inesperada: %#v", ok)
	}
}

func TestSanitizeModuloColombiaURL(t *testing.T) {
	if got := sanitizeModuloColombiaURL(" javascript:alert(1) "); got != "" {
		t.Fatalf("javascript deberia bloquearse, got %q", got)
	}
	if got := sanitizeModuloColombiaURL("https://example.com/soporte.pdf"); got != "https://example.com/soporte.pdf" {
		t.Fatalf("url valida inesperada: %q", got)
	}
}

func TestNormalizeModuloColombiaAprobacion(t *testing.T) {
	if got := normalizeModuloColombiaAprobacionEstado(" APROBADO "); got != "aprobado" {
		t.Fatalf("estado aprobacion inesperado: %q", got)
	}
	if got := normalizeModuloColombiaAprobacionEstado("raro"); got != "pendiente" {
		t.Fatalf("estado aprobacion por defecto inesperado: %q", got)
	}
	if got := normalizeModuloColombiaNivelAprobacion(" Juridico "); got != "juridico" {
		t.Fatalf("nivel aprobacion inesperado: %q", got)
	}
	if got := normalizeModuloColombiaNivelAprobacion("otro"); got != "operativo" {
		t.Fatalf("nivel aprobacion por defecto inesperado: %q", got)
	}
}

func TestNormalizeModuloColombiaTareaEstado(t *testing.T) {
	if got := normalizeModuloColombiaTareaEstado(" EN_PROCESO "); got != "en_proceso" {
		t.Fatalf("estado tarea inesperado: %q", got)
	}
	if got := normalizeModuloColombiaTareaEstado("raro"); got != "pendiente" {
		t.Fatalf("estado tarea por defecto inesperado: %q", got)
	}
}

func TestRecomendacionEmpresaModuloColombiaExpediente(t *testing.T) {
	row := EmpresaModuloColombiaRegistro{Estado: "en_proceso"}
	if got := recomendacionEmpresaModuloColombiaExpediente(row, map[string]int{"aprobaciones_pendientes": 1}); got == "" || got == "Expediente con trazabilidad suficiente para seguimiento operativo." {
		t.Fatalf("recomendacion de aprobacion inesperada: %q", got)
	}
	if got := recomendacionEmpresaModuloColombiaExpediente(row, map[string]int{"evidencias": 0}); got == "" {
		t.Fatalf("recomendacion sin evidencias vacia")
	}
	if !isModuloColombiaEstadoFinal("cerrado") || isModuloColombiaEstadoFinal("en_proceso") {
		t.Fatalf("estado final inesperado")
	}
}

func TestRecomendacionesModuloColombiaAgenda(t *testing.T) {
	got := recomendacionesModuloColombiaAgenda(EmpresaModuloColombiaAgenda{RegistrosVencidos: 1, AprobacionesPendientes: 2, TareasProximas: 1})
	if len(got) != 3 {
		t.Fatalf("recomendaciones de agenda inesperadas: %#v", got)
	}
	ok := recomendacionesModuloColombiaAgenda(EmpresaModuloColombiaAgenda{})
	if len(ok) != 1 || ok[0] == "" {
		t.Fatalf("agenda saludable inesperada: %#v", ok)
	}
	if moduloColombiaSeveridadRank("critica") >= moduloColombiaSeveridadRank("media") {
		t.Fatalf("orden de severidad inesperado")
	}
}

func TestCountModuloColombiaTareasAbiertas(t *testing.T) {
	got := countModuloColombiaTareasAbiertas([]EmpresaModuloColombiaTarea{
		{Estado: "pendiente"},
		{Estado: "en_proceso"},
		{Estado: "cumplida"},
		{Estado: "cancelada"},
	})
	if got != 2 {
		t.Fatalf("tareas abiertas = %d", got)
	}
}

func TestPlanAccionModuloColombiaHelpers(t *testing.T) {
	item := EmpresaModuloColombiaAgendaItem{Tipo: "registro", Titulo: "Vence contrato", Severidad: "critica"}
	if got := tituloPlanAccionModuloColombia(item); got != "Plan de accion - registro: Vence contrato" {
		t.Fatalf("titulo plan accion = %q", got)
	}
	if got := prioridadPlanAccionModuloColombia("critica"); got != "urgente" {
		t.Fatalf("prioridad critica = %q", got)
	}
	if got := prioridadPlanAccionModuloColombia("alta"); got != "alta" {
		t.Fatalf("prioridad alta = %q", got)
	}
	if got := prioridadPlanAccionModuloColombia("media"); got != "normal" {
		t.Fatalf("prioridad media = %q", got)
	}
}

func TestResponsableModuloColombiaHelpers(t *testing.T) {
	rows := map[string]*EmpresaModuloColombiaResponsableResumen{}
	got := getModuloColombiaResponsableResumen(rows, "")
	if got.Responsable != "Sin responsable" {
		t.Fatalf("responsable vacio = %q", got.Responsable)
	}
	if rec := recomendacionModuloColombiaResponsable(EmpresaModuloColombiaResponsableResumen{RegistrosVencidos: 1}); rec == "Carga operativa controlada." {
		t.Fatalf("recomendacion de vencido inesperada: %q", rec)
	}
	if rec := recomendacionModuloColombiaResponsable(EmpresaModuloColombiaResponsableResumen{TotalPendiente: 9}); rec != "Rebalancear carga o reasignar tareas." {
		t.Fatalf("recomendacion de carga inesperada: %q", rec)
	}
}

func TestSLAModuloColombiaHelpers(t *testing.T) {
	today, _ := time.Parse("2006-01-02", "2026-05-06")
	if got := bucketModuloColombiaVencimiento("", today); got != "sin_vencimiento" {
		t.Fatalf("bucket vacio = %q", got)
	}
	if got := bucketModuloColombiaVencimiento("2026-05-05", today); got != "vencido" {
		t.Fatalf("bucket vencido = %q", got)
	}
	if got := bucketModuloColombiaVencimiento("2026-05-10", today); got != "0_7" {
		t.Fatalf("bucket 0_7 = %q", got)
	}
	if got := calcularCumplimientoModuloColombia(10, 2); got != 80 {
		t.Fatalf("cumplimiento = %.2f", got)
	}
	if got := semaforoModuloColombiaSLA(79, 0); got != "rojo" {
		t.Fatalf("semaforo rojo = %q", got)
	}
	if got := semaforoModuloColombiaSLA(90, 0); got != "amarillo" {
		t.Fatalf("semaforo amarillo = %q", got)
	}
	if got := semaforoModuloColombiaSLA(100, 0); got != "verde" {
		t.Fatalf("semaforo verde = %q", got)
	}
}

func TestRiesgoModuloColombiaHelpers(t *testing.T) {
	score, factores := scoreModuloColombiaRiesgo(EmpresaModuloColombiaRiesgo{
		RegistrosVencidos:      2,
		CriticosAbiertos:       1,
		AprobacionesPendientes: 1,
		TareasVencidas:         1,
		SinResponsable:         1,
		SinEvidencia:           1,
	})
	if score <= 0 || len(factores) == 0 {
		t.Fatalf("score/factores inesperados: %d %#v", score, factores)
	}
	if got := nivelModuloColombiaRiesgo(80); got != "alto" {
		t.Fatalf("nivel alto = %q", got)
	}
	if got := nivelModuloColombiaRiesgo(40); got != "medio" {
		t.Fatalf("nivel medio = %q", got)
	}
	if got := nivelModuloColombiaRiesgo(10); got != "bajo" {
		t.Fatalf("nivel bajo = %q", got)
	}
	if got := minIntModuloColombia(3, 5); got != 3 {
		t.Fatalf("min = %d", got)
	}
}

func TestExportacionModuloColombiaHelpers(t *testing.T) {
	registros := exportacionModuloColombiaRegistros([]EmpresaModuloColombiaRegistro{{ID: 7, Codigo: "KYC-001", Nombre: "Debida diligencia", Valor: 1200.5}})
	if registros.Nombre != "registros" || len(registros.Rows) != 1 || registros.Rows[0][0] != "7" || registros.Rows[0][12] != "1200.5" {
		t.Fatalf("exportacion registros inesperada: %#v", registros)
	}
	prefijo := exportacionModuloColombiaPrefijo("riesgo", []string{"", "Atender vencidos"})
	if len(prefijo) != 1 || prefijo[0][0] != "riesgo" {
		t.Fatalf("prefijo exportacion inesperado: %#v", prefijo)
	}
	if got := formatFloatModuloColombia(1000.00); got != "1000" {
		t.Fatalf("float entero = %q", got)
	}
	if got := formatFloatModuloColombia(1000.25); got != "1000.25" {
		t.Fatalf("float decimal = %q", got)
	}
}

func TestNormalizeEmpresaModuloColombiaFiltro(t *testing.T) {
	got := normalizeEmpresaModuloColombiaFiltro(EmpresaModuloColombiaFiltro{
		Texto:        "  Contrato ABC  ",
		Estado:       "",
		Tipo:         "Pago Masivo",
		Categoria:    "Centro/Costo",
		Prioridad:    "rara",
		Responsable:  "  Ana  ",
		Vencidos:     true,
		ProximosDias: 999,
	})
	if got.Texto != "contrato abc" || got.Estado != "" || got.Tipo != "pago_masivo" || got.Categoria != "centro_costo" {
		t.Fatalf("filtro texto/tipo/categoria inesperado: %#v", got)
	}
	if got.Prioridad != "normal" || got.Responsable != "Ana" || !got.Vencidos || got.ProximosDias != 0 {
		t.Fatalf("filtro prioridad/responsable/vencidos inesperado: %#v", got)
	}
	proximo := normalizeEmpresaModuloColombiaFiltro(EmpresaModuloColombiaFiltro{ProximosDias: 400})
	if proximo.ProximosDias != 365 {
		t.Fatalf("limite proximos dias inesperado: %#v", proximo)
	}
}

func TestAccionMasivaModuloColombiaHelpers(t *testing.T) {
	accion := normalizeEmpresaModuloColombiaAccionMasiva(EmpresaModuloColombiaAccionMasiva{
		RegistroIDs: []int64{3, 3, 0, 5},
		Estado:      " En Proceso ",
		Prioridad:   "URGENTE",
		Responsable: "  Jefe Operativo  ",
		Detalle:     "  Reasignacion semanal  ",
	})
	if accion.Estado != "en_proceso" || accion.Prioridad != "urgente" || accion.Responsable != "Jefe Operativo" || accion.Detalle != "Reasignacion semanal" {
		t.Fatalf("accion masiva normalizada inesperada: %#v", accion)
	}
	ids := uniqueModuloColombiaIDs(accion.RegistroIDs)
	if len(ids) != 2 || ids[0] != 3 || ids[1] != 5 {
		t.Fatalf("ids unicos inesperados: %#v", ids)
	}
	detalle := detalleAccionMasivaModuloColombia(accion, accion.Detalle)
	if !strings.Contains(detalle, "estado=en_proceso") || !strings.Contains(detalle, "responsable=Jefe Operativo") {
		t.Fatalf("detalle accion masiva inesperado: %q", detalle)
	}
}
