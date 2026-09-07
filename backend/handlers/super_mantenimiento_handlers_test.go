package handlers

import (
	"testing"
	"time"
)

func TestMantenimientoAvisosSeleccionaPrimerActivo(t *testing.T) {
	avisos := normalizeMantenimientoAvisos([]mantenimientoAviso{
		{ID: "b", AvisoActivo: true, Fecha: "2026-06-02", HoraInicio: "03:00", HoraFin: "04:00", ZonaHoraria: "America/Bogota", Mensaje: "Segundo"},
		{ID: "a", AvisoActivo: true, Fecha: "2026-06-01", HoraInicio: "02:00", HoraFin: "03:00", ZonaHoraria: "America/Bogota", Mensaje: "Primero"},
	})

	selected := selectMantenimientoAvisoPrincipal(avisos)
	if selected == nil {
		t.Fatal("se esperaba aviso activo")
	}
	if selected.ID != "a" {
		t.Fatalf("aviso seleccionado = %s, want a", selected.ID)
	}
}

func TestMantenimientoAvisoUpsertYDesactivacion(t *testing.T) {
	base := mantenimientoConfig{
		AvisoID:     "ventana-1",
		AvisoActivo: true,
		Fecha:       "2026-06-01",
		HoraInicio:  "02:00",
		HoraFin:     "03:00",
		ZonaHoraria: "America/Bogota",
		Mensaje:     "Mantenimiento de prueba",
	}

	avisos := upsertMantenimientoAviso(nil, mantenimientoAvisoFromConfig(base))
	if len(avisos) != 1 {
		t.Fatalf("avisos len = %d, want 1", len(avisos))
	}
	if !avisos[0].AvisoActivo {
		t.Fatal("el aviso debe quedar activo")
	}

	avisos[0].AvisoActivo = false
	selected := selectMantenimientoAvisoPrincipal(normalizeMantenimientoAvisos(avisos))
	if selected != nil {
		t.Fatal("no debe haber aviso principal al desactivar todos")
	}
}

func TestCleanMantenimientoID(t *testing.T) {
	got := cleanMantenimientoID(" Aviso Junio 01!! ")
	if got != "avisojunio01" {
		t.Fatalf("cleanMantenimientoID = %q, want avisojunio01", got)
	}
}

func TestFiltrarMantenimientoAvisosVigentesEliminaViejosEInactivos(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	next, removed := filtrarMantenimientoAvisosVigentes([]mantenimientoAviso{
		{ID: "pasado-activo", AvisoActivo: true, Fecha: "2026-06-01", HoraInicio: "02:00", HoraFin: "03:00", ZonaHoraria: "America/Bogota", Mensaje: "Viejo"},
		{ID: "futuro-inactivo", AvisoActivo: false, Fecha: "2026-06-20", HoraInicio: "02:00", HoraFin: "03:00", ZonaHoraria: "America/Bogota", Mensaje: "Inactivo"},
		{ID: "futuro-activo", AvisoActivo: true, Fecha: "2026-06-20", HoraInicio: "02:00", HoraFin: "03:00", ZonaHoraria: "America/Bogota", Mensaje: "Vigente"},
	}, now)
	if removed != 2 {
		t.Fatalf("removed = %d, want 2", removed)
	}
	if len(next) != 1 || next[0].ID != "futuro-activo" {
		t.Fatalf("avisos vigentes = %#v, want solo futuro-activo", next)
	}
}
