package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestEmpresaNominaSueldosHandlerFlow(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_nomina_handler.db")
	if err := dbpkg.EnsureEmpresaNominaSchema(dbEmp); err != nil {
		t.Fatalf("ensure nomina schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaAsistenciaSchema(dbEmp); err != nil {
		t.Fatalf("ensure asistencia schema: %v", err)
	}

	h := EmpresaNominaSueldosHandler(dbEmp)

	cfgBody := `{"empresa_id":9,"horas_ordinarias_semana":44,"horas_ordinarias_dia":8,"dias_nomina_mes":30,"recargo_nocturno_porcentaje":35,"hora_extra_diurna_porcentaje":25,"hora_extra_nocturna_porcentaje":75,"recargo_dominical_diurno_porcentaje":75,"recargo_dominical_nocturno_porcentaje":110,"hora_extra_dominical_diurna_porcentaje":100,"hora_extra_dominical_nocturna_porcentaje":150,"deduccion_salud_porcentaje":4,"deduccion_pension_porcentaje":4}`
	cfgReq := httptest.NewRequest(http.MethodPut, "/api/empresa/nomina?empresa_id=9&action=config", strings.NewReader(cfgBody))
	cfgReq.Header.Set("Content-Type", "application/json")
	cfgRR := httptest.NewRecorder()
	h.ServeHTTP(cfgRR, cfgReq)
	if cfgRR.Code != http.StatusOK {
		t.Fatalf("config expected=%d got=%d body=%s", http.StatusOK, cfgRR.Code, cfgRR.Body.String())
	}

	empleadoBody := `{"empresa_id":9,"empleado_id":9001,"empleado_codigo":"EMP-9001","empleado_nombre":"Daniel Castro","empleado_documento":"90112233","cargo":"Analista","salario_basico_mensual":2100000,"auxilio_transporte_mensual":162000,"jornada_horas_dia":8,"incluir_auxilio_transporte":true}`
	empleadoReq := httptest.NewRequest(http.MethodPost, "/api/empresa/nomina?action=empleado", strings.NewReader(empleadoBody))
	empleadoReq.Header.Set("Content-Type", "application/json")
	empleadoRR := httptest.NewRecorder()
	h.ServeHTTP(empleadoRR, empleadoReq)
	if empleadoRR.Code != http.StatusCreated {
		t.Fatalf("create empleado expected=%d got=%d body=%s", http.StatusCreated, empleadoRR.Code, empleadoRR.Body.String())
	}

	var empleadoResp map[string]interface{}
	if err := json.Unmarshal(empleadoRR.Body.Bytes(), &empleadoResp); err != nil {
		t.Fatalf("decode empleado response: %v", err)
	}
	empleadoNominaID := int64(empleadoResp["id"].(float64))
	if empleadoNominaID <= 0 {
		t.Fatalf("expected empleado id > 0, got %d", empleadoNominaID)
	}

	festivoBody := `{"empresa_id":9,"fecha_festivo":"2026-04-03","descripcion":"Festivo handler"}`
	festivoReq := httptest.NewRequest(http.MethodPost, "/api/empresa/nomina?action=festivo", strings.NewReader(festivoBody))
	festivoReq.Header.Set("Content-Type", "application/json")
	festivoRR := httptest.NewRecorder()
	h.ServeHTTP(festivoRR, festivoReq)
	if festivoRR.Code != http.StatusCreated {
		t.Fatalf("create festivo expected=%d got=%d body=%s", http.StatusCreated, festivoRR.Code, festivoRR.Body.String())
	}

	if _, err := dbpkg.CreateEmpresaAsistenciaEmpleado(dbEmp, dbpkg.EmpresaAsistenciaEmpleado{
		EmpresaID:         9,
		EmpleadoID:        9001,
		EmpleadoCodigo:    "EMP-9001",
		EmpleadoNombre:    "Daniel Castro",
		EmpleadoDocumento: "90112233",
		Cargo:             "Analista",
		FechaAsistencia:   "2026-04-01",
		HoraEntrada:       "08:00:00",
		HoraSalida:        "18:00:00",
		HorasTrabajadas:   10,
		EstadoAsistencia:  "presente",
		UsuarioCreador:    "qa@empresa.com",
	}); err != nil {
		t.Fatalf("seed asistencia extra diurna: %v", err)
	}
	if _, err := dbpkg.CreateEmpresaAsistenciaEmpleado(dbEmp, dbpkg.EmpresaAsistenciaEmpleado{
		EmpresaID:         9,
		EmpleadoID:        9001,
		EmpleadoCodigo:    "EMP-9001",
		EmpleadoNombre:    "Daniel Castro",
		EmpleadoDocumento: "90112233",
		Cargo:             "Analista",
		FechaAsistencia:   "2026-04-02",
		HoraEntrada:       "21:00:00",
		HoraSalida:        "23:00:00",
		HorasTrabajadas:   2,
		EstadoAsistencia:  "presente",
		UsuarioCreador:    "qa@empresa.com",
	}); err != nil {
		t.Fatalf("seed asistencia recargo nocturno: %v", err)
	}
	if _, err := dbpkg.CreateEmpresaAsistenciaEmpleado(dbEmp, dbpkg.EmpresaAsistenciaEmpleado{
		EmpresaID:         9,
		EmpleadoID:        9001,
		EmpleadoCodigo:    "EMP-9001",
		EmpleadoNombre:    "Daniel Castro",
		EmpleadoDocumento: "90112233",
		Cargo:             "Analista",
		FechaAsistencia:   "2026-04-03",
		HoraEntrada:       "08:00:00",
		HoraSalida:        "12:00:00",
		HorasTrabajadas:   4,
		EstadoAsistencia:  "presente",
		UsuarioCreador:    "qa@empresa.com",
	}); err != nil {
		t.Fatalf("seed asistencia festivo: %v", err)
	}

	calcBody := `{"empresa_id":9,"periodo_desde":"2026-04-01","periodo_hasta":"2026-04-10","empleado_nomina_id":` + strconv.FormatInt(empleadoNominaID, 10) + `,"overwrite":true}`
	calcReq := httptest.NewRequest(http.MethodPost, "/api/empresa/nomina?action=calcular", strings.NewReader(calcBody))
	calcReq.Header.Set("Content-Type", "application/json")
	calcRR := httptest.NewRecorder()
	h.ServeHTTP(calcRR, calcReq)
	if calcRR.Code != http.StatusOK {
		t.Fatalf("calcular expected=%d got=%d body=%s", http.StatusOK, calcRR.Code, calcRR.Body.String())
	}

	var calcResp dbpkg.EmpresaNominaCalculoResult
	if err := json.Unmarshal(calcRR.Body.Bytes(), &calcResp); err != nil {
		t.Fatalf("decode calcular response: %v", err)
	}
	if calcResp.Calculados != 1 {
		t.Fatalf("expected calculados=1, got %d", calcResp.Calculados)
	}
	if len(calcResp.Liquidaciones) != 1 {
		t.Fatalf("expected liquidaciones=1, got %d", len(calcResp.Liquidaciones))
	}
	if calcResp.Liquidaciones[0].HorasExtraDiurnas < 1.9 {
		t.Fatalf("expected horas_extra_diurnas >= 1.9, got %.2f", calcResp.Liquidaciones[0].HorasExtraDiurnas)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/empresa/nomina?empresa_id=9&action=liquidaciones&periodo_desde=2026-04-01&periodo_hasta=2026-04-30&limit=20", nil)
	listRR := httptest.NewRecorder()
	h.ServeHTTP(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("list liquidaciones expected=%d got=%d body=%s", http.StatusOK, listRR.Code, listRR.Body.String())
	}
	var rows []dbpkg.EmpresaNominaLiquidacion
	if err := json.Unmarshal(listRR.Body.Bytes(), &rows); err != nil {
		t.Fatalf("decode list liquidaciones: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 liquidacion row, got %d", len(rows))
	}
}
