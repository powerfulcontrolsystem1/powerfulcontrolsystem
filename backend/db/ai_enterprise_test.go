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
