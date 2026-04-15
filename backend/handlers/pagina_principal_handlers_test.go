package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPaginaPrincipalNormalizeConfigCompletesLandingFields(t *testing.T) {
	defaults := paginaPrincipalDefaultConfig()
	normalized := paginaPrincipalNormalizeConfig(paginaPrincipalConfig{
		Cantidad: 1,
		Tarjetas: []paginaPrincipalCard{
			{
				Titulo:      "Tarjeta personalizada",
				Descripcion: "Oferta principal personalizada.",
				ImagenURL:   "/img/punto_venta.png",
				Enlace:      "/login.html",
			},
		},
	})

	if normalized.Cantidad != 1 {
		t.Fatalf("expected cantidad 1, got %d", normalized.Cantidad)
	}
	if len(normalized.Tarjetas) != 1 {
		t.Fatalf("expected 1 tarjeta, got %d", len(normalized.Tarjetas))
	}

	card := normalized.Tarjetas[0]
	base := defaults.Tarjetas[0]
	if card.DetalleEtiqueta != base.DetalleEtiqueta {
		t.Fatalf("expected detalle_etiqueta %q, got %q", base.DetalleEtiqueta, card.DetalleEtiqueta)
	}
	if card.DetalleTitular != base.DetalleTitular {
		t.Fatalf("expected detalle_titular %q, got %q", base.DetalleTitular, card.DetalleTitular)
	}
	if card.DetalleParrafoUno != base.DetalleParrafoUno {
		t.Fatalf("expected detalle_parrafo_uno %q, got %q", base.DetalleParrafoUno, card.DetalleParrafoUno)
	}
	if card.DetalleParrafoDos != base.DetalleParrafoDos {
		t.Fatalf("expected detalle_parrafo_dos %q, got %q", base.DetalleParrafoDos, card.DetalleParrafoDos)
	}
	if len(card.DetallePuntos) != len(base.DetallePuntos) {
		t.Fatalf("expected %d detalle_puntos, got %d", len(base.DetallePuntos), len(card.DetallePuntos))
	}
	for idx, point := range base.DetallePuntos {
		if card.DetallePuntos[idx] != point {
			t.Fatalf("expected detalle_puntos[%d] %q, got %q", idx, point, card.DetallePuntos[idx])
		}
	}
}

func TestPublicPaginaPrincipalHandlerExposesLandingFields(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_pagina_principal_landing_config.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	cfg := paginaPrincipalConfig{
		Cantidad: 1,
		Tarjetas: []paginaPrincipalCard{
			{
				Titulo:            "Hotel boutique",
				Descripcion:       "Operacion personalizada para hospedaje premium.",
				ImagenURL:         "/img/settings-color.svg",
				Enlace:            "/login.html",
				DetalleEtiqueta:   "Hospedaje premium",
				DetalleTitular:    "Administra reservas, check-in y cargos con mayor detalle.",
				DetalleParrafoUno: "Primer parrafo ampliado de la landing.",
				DetalleParrafoDos: "Segundo parrafo ampliado de la landing.",
				DetallePuntos:     []string{"Capacidad 1", "Capacidad 2", "Capacidad 3"},
			},
		},
	}
	if err := paginaPrincipalSaveConfig(dbSuper, cfg, "super@demo.com"); err != nil {
		t.Fatalf("save pagina principal config: %v", err)
	}

	h := PublicPaginaPrincipalHandler(dbSuper)
	req := httptest.NewRequest(http.MethodGet, "/api/public/pagina_principal", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp struct {
		Cantidad int                   `json:"cantidad"`
		Tarjetas []paginaPrincipalCard `json:"tarjetas"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, rr.Body.String())
	}
	if resp.Cantidad != 1 {
		t.Fatalf("expected cantidad 1, got %d", resp.Cantidad)
	}
	if len(resp.Tarjetas) != 1 {
		t.Fatalf("expected 1 tarjeta, got %d", len(resp.Tarjetas))
	}

	card := resp.Tarjetas[0]
	if card.Titulo != cfg.Tarjetas[0].Titulo {
		t.Fatalf("expected titulo %q, got %q", cfg.Tarjetas[0].Titulo, card.Titulo)
	}
	if card.DetalleEtiqueta != cfg.Tarjetas[0].DetalleEtiqueta {
		t.Fatalf("expected detalle_etiqueta %q, got %q", cfg.Tarjetas[0].DetalleEtiqueta, card.DetalleEtiqueta)
	}
	if card.DetalleTitular != cfg.Tarjetas[0].DetalleTitular {
		t.Fatalf("expected detalle_titular %q, got %q", cfg.Tarjetas[0].DetalleTitular, card.DetalleTitular)
	}
	if card.DetalleParrafoUno != cfg.Tarjetas[0].DetalleParrafoUno {
		t.Fatalf("expected detalle_parrafo_uno %q, got %q", cfg.Tarjetas[0].DetalleParrafoUno, card.DetalleParrafoUno)
	}
	if card.DetalleParrafoDos != cfg.Tarjetas[0].DetalleParrafoDos {
		t.Fatalf("expected detalle_parrafo_dos %q, got %q", cfg.Tarjetas[0].DetalleParrafoDos, card.DetalleParrafoDos)
	}
	if len(card.DetallePuntos) != len(cfg.Tarjetas[0].DetallePuntos) {
		t.Fatalf("expected %d detalle_puntos, got %d", len(cfg.Tarjetas[0].DetallePuntos), len(card.DetallePuntos))
	}
	for idx, point := range cfg.Tarjetas[0].DetallePuntos {
		if card.DetallePuntos[idx] != point {
			t.Fatalf("expected detalle_puntos[%d] %q, got %q", idx, point, card.DetallePuntos[idx])
		}
	}
}
