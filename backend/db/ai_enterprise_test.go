package db

import "testing"

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
