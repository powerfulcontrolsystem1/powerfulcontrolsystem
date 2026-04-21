package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestPaginaPrincipalNormalizeConfigCompletesLandingFields(t *testing.T) {
	defaults := paginaPrincipalDefaultConfig()
	normalized := paginaPrincipalNormalizeConfig(paginaPrincipalConfig{
		Cantidad: 1,
		Estilos: paginaPrincipalVisualSettings{
			IndexCardSize:   "large",
			IndexTextSize:   "small",
			LandingCardSize: "grande",
			LandingTextSize: "valor-invalido",
		},
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
	if normalized.Estilos.IndexCardSize != paginaPrincipalVisualSizeLarge {
		t.Fatalf("expected index_card_size %q, got %q", paginaPrincipalVisualSizeLarge, normalized.Estilos.IndexCardSize)
	}
	if normalized.Estilos.IndexTextSize != paginaPrincipalVisualSizeSmall {
		t.Fatalf("expected index_text_size %q, got %q", paginaPrincipalVisualSizeSmall, normalized.Estilos.IndexTextSize)
	}
	if normalized.Estilos.LandingCardSize != paginaPrincipalVisualSizeLarge {
		t.Fatalf("expected landing_card_size %q, got %q", paginaPrincipalVisualSizeLarge, normalized.Estilos.LandingCardSize)
	}
	if normalized.Estilos.LandingTextSize != defaults.Estilos.LandingTextSize {
		t.Fatalf("expected landing_text_size %q, got %q", defaults.Estilos.LandingTextSize, normalized.Estilos.LandingTextSize)
	}
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
	if err := dbpkg.SetConfigValue(dbSuper, "portal.whatsapp_contact_number", "573001112233", false); err != nil {
		t.Fatalf("seed portal.whatsapp_contact_number: %v", err)
	}

	cfg := paginaPrincipalConfig{
		Cantidad: 1,
		Estilos: paginaPrincipalVisualSettings{
			IndexCardSize:   paginaPrincipalVisualSizeSmall,
			IndexTextSize:   paginaPrincipalVisualSizeLarge,
			LandingCardSize: paginaPrincipalVisualSizeLarge,
			LandingTextSize: paginaPrincipalVisualSizeSmall,
		},
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
		Cantidad int                           `json:"cantidad"`
		Tarjetas []paginaPrincipalCard         `json:"tarjetas"`
		Estilos  paginaPrincipalVisualSettings `json:"estilos"`
		WhatsAppContactNumber string           `json:"whatsapp_contact_number"`
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
	if resp.Estilos.IndexCardSize != cfg.Estilos.IndexCardSize {
		t.Fatalf("expected index_card_size %q, got %q", cfg.Estilos.IndexCardSize, resp.Estilos.IndexCardSize)
	}
	if resp.Estilos.IndexTextSize != cfg.Estilos.IndexTextSize {
		t.Fatalf("expected index_text_size %q, got %q", cfg.Estilos.IndexTextSize, resp.Estilos.IndexTextSize)
	}
	if resp.Estilos.LandingCardSize != cfg.Estilos.LandingCardSize {
		t.Fatalf("expected landing_card_size %q, got %q", cfg.Estilos.LandingCardSize, resp.Estilos.LandingCardSize)
	}
	if resp.Estilos.LandingTextSize != cfg.Estilos.LandingTextSize {
		t.Fatalf("expected landing_text_size %q, got %q", cfg.Estilos.LandingTextSize, resp.Estilos.LandingTextSize)
	}
	if resp.WhatsAppContactNumber != "573001112233" {
		t.Fatalf("expected whatsapp_contact_number %q, got %q", "573001112233", resp.WhatsAppContactNumber)
	}
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

func TestPaginaPrincipalSaveAndLoadPreservesExpandedCardCount(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_pagina_principal_expanded_cards.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	cfg := paginaPrincipalDefaultConfig()
	cfg.Cantidad = 7
	cfg.Tarjetas = append(cfg.Tarjetas,
		paginaPrincipalCard{
			Titulo:            "Lavanderia industrial",
			Descripcion:       "Operacion ampliada para lavado y seguimiento por servicio.",
			ImagenURL:         "/img/punto_venta.png",
			Enlace:            "/login.html",
			DetalleEtiqueta:   "Lavado especializado",
			DetalleTitular:    "Amplia el catalogo publico con mas tarjetas persistentes.",
			DetalleParrafoUno: "Tarjeta adicional para validar persistencia de cantidades mayores al set inicial.",
			DetalleParrafoDos: "Debe conservarse despues de guardar y volver a cargar desde configuracion super.",
			DetallePuntos:     []string{"Servicio 1", "Servicio 2"},
		},
		paginaPrincipalCard{
			Titulo:            "Parqueadero inteligente",
			Descripcion:       "Control de acceso y permanencia con trazabilidad.",
			ImagenURL:         "/img/punto_venta.png",
			Enlace:            "/login.html",
			DetalleEtiqueta:   "Accesos y sensores",
			DetalleTitular:    "Segunda tarjeta extra para verificar cantidad ampliada.",
			DetalleParrafoUno: "La configuracion debe persistir todas las tarjetas solicitadas por el editor.",
			DetalleParrafoDos: "Al recargar, la cantidad y el contenido de las tarjetas extra deben seguir presentes.",
			DetallePuntos:     []string{"Acceso 1", "Acceso 2"},
		},
	)

	if err := paginaPrincipalSaveConfig(dbSuper, cfg, "super@demo.com"); err != nil {
		t.Fatalf("save pagina principal config: %v", err)
	}

	loaded, _, updatedBy, err := paginaPrincipalLoadConfig(dbSuper)
	if err != nil {
		t.Fatalf("load pagina principal config: %v", err)
	}
	if updatedBy != "super@demo.com" {
		t.Fatalf("expected updated_by %q, got %q", "super@demo.com", updatedBy)
	}
	if loaded.Cantidad != 7 {
		t.Fatalf("expected cantidad 7, got %d", loaded.Cantidad)
	}
	if len(loaded.Tarjetas) != 7 {
		t.Fatalf("expected 7 tarjetas, got %d", len(loaded.Tarjetas))
	}
	if loaded.Tarjetas[5].Titulo != "Lavanderia industrial" {
		t.Fatalf("expected tarjeta 6 titulo %q, got %q", "Lavanderia industrial", loaded.Tarjetas[5].Titulo)
	}
	if loaded.Tarjetas[6].Titulo != "Parqueadero inteligente" {
		t.Fatalf("expected tarjeta 7 titulo %q, got %q", "Parqueadero inteligente", loaded.Tarjetas[6].Titulo)
	}
}
