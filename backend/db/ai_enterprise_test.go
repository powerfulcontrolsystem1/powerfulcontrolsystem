package db

import (
	"os"
	"strings"
	"testing"
)

func TestEmpresaAIEnterpriseSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaAIEnterpriseSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}

func TestEmpresaAIOperationsDoNotCreateSchemaAtRuntime(t *testing.T) {
	body, err := os.ReadFile("ai_enterprise.go")
	if err != nil {
		t.Fatalf("read AI enterprise source: %v", err)
	}
	source := string(body)
	for _, function := range []string{
		"RecordEmpresaAIExecution",
		"CreateOrRefreshEmpresaAIConversation",
		"CreateEmpresaAIProposal",
		"GetEmpresaAIProposal",
		"BeginEmpresaAIProposalExecution",
	} {
		start := strings.Index(source, "func "+function)
		if start < 0 {
			t.Fatalf("%s not found", function)
		}
		section := source[start:]
		if next := strings.Index(section[1:], "\nfunc "); next >= 0 {
			section = section[:next+1]
		}
		if strings.Contains(section, "EnsureEmpresaAIEnterpriseSchema(") {
			t.Fatalf("%s must verify, not create, the AI schema", function)
		}
		if !strings.Contains(section, "EmpresaAIEnterpriseSchemaReady(") {
			t.Fatalf("%s must verify the AI schema", function)
		}
	}
}

func TestNormalizeEmpresaAIHotelRoomPlanRejectsAmbiguousOrUnsafeInput(t *testing.T) {
	plan := EmpresaAIHotelRoomPlan{EstacionID: 1, NombreHabitacion: "Habitacion 1", Moneda: "COP", HoraCheckIn: "14:00", HoraCheckOut: "13:00", Tarifas: []EmpresaAIHotelRoomRate{{Personas: 2, Valor: 100000}, {Personas: 2, Valor: 200000}}}
	if err := NormalizeEmpresaAIHotelRoomPlan(&plan); err == nil {
		t.Fatal("expected duplicate occupancy to be rejected")
	}
	plan.Tarifas[1].Personas = 4
	if err := NormalizeEmpresaAIHotelRoomPlan(&plan); err != nil {
		t.Fatalf("valid hotel plan rejected: %v", err)
	}
}

func TestNormalizeEmpresaAIHotelRoomPlanRejectsMissingStation(t *testing.T) {
	plan := EmpresaAIHotelRoomPlan{NombreHabitacion: "Habitacion 1", Tarifas: []EmpresaAIHotelRoomRate{{Personas: 2, Valor: 100000}}}
	if err := NormalizeEmpresaAIHotelRoomPlan(&plan); err == nil {
		t.Fatal("expected station validation error")
	}
}

func TestNormalizeEmpresaAIProductCreatePlan(t *testing.T) {
	valid := EmpresaAIProductCreatePlan{Nombre: "Producto de prueba", Precio: 5000, ImpuestoPorcentaje: 19, BodegaID: 4, StockInicial: 10}
	if err := NormalizeEmpresaAIProductCreatePlan(&valid); err != nil {
		t.Fatalf("valid product plan rejected: %v", err)
	}
	for _, plan := range []EmpresaAIProductCreatePlan{
		{Nombre: "", Precio: 1},
		{Nombre: "Producto", Precio: -1},
		{Nombre: "Producto", Precio: 1, ImpuestoPorcentaje: 101},
		{Nombre: "Producto", Precio: 1, StockInicial: 1},
	} {
		if err := NormalizeEmpresaAIProductCreatePlan(&plan); err == nil {
			t.Fatalf("invalid product plan was accepted: %#v", plan)
		}
	}
}
