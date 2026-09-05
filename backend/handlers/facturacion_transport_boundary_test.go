package handlers

import (
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestFacturacionTransportRejectsSimulatedProductionForEveryCountry(t *testing.T) {
	for _, country := range []string{"CO", "EC", "PA", "CR", "AR", "VE"} {
		for _, provider := range []string{"manual", "interno", "local", "proveedor_externo"} {
			t.Run(country+"/"+provider, func(t *testing.T) {
				cfg := &dbpkg.FacturacionElectronicaPaisConfig{EmpresaID: 12, PaisCodigo: country, Ambiente: "produccion", Proveedor: provider, APIBaseURL: "mock://ok", Estado: "activo"}
				doc := dbpkg.EmpresaDocumentoFacturacion{EmpresaID: 12, PaisCodigo: country, AmbienteFE: "produccion", TipoDocumento: "factura_electronica", DocumentoCodigo: "QA-TRANSPORT"}
				got := dispatchFacturacionProveedor(nil, cfg, facturacionOperacionPayload{EmpresaID: 12, PaisCodigo: country}, doc, "emitir", false)
				if got.Success || !got.FinalFailure || got.ReferenciaExterna != "" || got.RespuestaJSON != "" {
					t.Fatalf("simulation must fail before any fiscal response: %#v", got)
				}
			})
		}
	}
}

func TestFacturacionProductionRequiresImplementedCountryAdapter(t *testing.T) {
	for _, country := range []string{"CO", "EC", "PA", "CR", "AR", "VE"} {
		t.Run(country, func(t *testing.T) {
			cfg := &dbpkg.FacturacionElectronicaPaisConfig{EmpresaID: 12, PaisCodigo: country, Ambiente: "produccion", Proveedor: "externo", APIBaseURL: "https://example.com/dian.gov.co", Estado: "activo"}
			doc := dbpkg.EmpresaDocumentoFacturacion{EmpresaID: 12, PaisCodigo: country, AmbienteFE: "produccion", TipoDocumento: "factura_electronica"}
			got := dispatchFacturacionProveedor(nil, cfg, facturacionOperacionPayload{}, doc, "emitir", false)
			if got.Success || !got.FinalFailure || !strings.Contains(got.Error, "adaptador fiscal especifico") {
				t.Fatalf("unimplemented provider reached dispatch: %#v", got)
			}
			if country != "CO" {
				status := facturacionProveedorConnectionStatus(cfg)
				if status["online"] != false || status["estado_conexion"] != "sin_adaptador_fiscal" || status["accion_recomendada"] != "bloquear_facturacion_electronica" {
					t.Fatalf("unimplemented country advertised as operational: %#v", status)
				}
			}
		})
	}
}

func TestFacturacionLocalProvidersRejectedWithoutMockURL(t *testing.T) {
	for _, provider := range []string{"manual", "interno", "local", ""} {
		cfg := &dbpkg.FacturacionElectronicaPaisConfig{EmpresaID: 12, PaisCodigo: "EC", Ambiente: "produccion", Proveedor: provider, Estado: "activo"}
		doc := dbpkg.EmpresaDocumentoFacturacion{EmpresaID: 12, PaisCodigo: "EC", AmbienteFE: "produccion", TipoDocumento: "factura_electronica"}
		got := dispatchFacturacionProveedor(nil, cfg, facturacionOperacionPayload{}, doc, "emitir", false)
		if got.Success || !got.FinalFailure || !strings.Contains(got.Error, "proveedor fiscal real") {
			t.Fatalf("local provider %q accepted: %#v", provider, got)
		}
	}
}

func TestFacturacionTransportRejectsCrossTenantCountryAndEnvironmentBeforeIO(t *testing.T) {
	tests := []struct {
		name   string
		change func(*dbpkg.FacturacionElectronicaPaisConfig, *facturacionOperacionPayload, *dbpkg.EmpresaDocumentoFacturacion)
	}{
		{"configuration tenant", func(c *dbpkg.FacturacionElectronicaPaisConfig, p *facturacionOperacionPayload, d *dbpkg.EmpresaDocumentoFacturacion) {
			c.EmpresaID = 13
		}},
		{"inactive configuration", func(c *dbpkg.FacturacionElectronicaPaisConfig, p *facturacionOperacionPayload, d *dbpkg.EmpresaDocumentoFacturacion) {
			c.Estado = "inactivo"
		}},
		{"payload tenant", func(c *dbpkg.FacturacionElectronicaPaisConfig, p *facturacionOperacionPayload, d *dbpkg.EmpresaDocumentoFacturacion) {
			p.EmpresaID = 13
		}},
		{"invalid tenant", func(c *dbpkg.FacturacionElectronicaPaisConfig, p *facturacionOperacionPayload, d *dbpkg.EmpresaDocumentoFacturacion) {
			d.EmpresaID = 0
		}},
		{"configuration country", func(c *dbpkg.FacturacionElectronicaPaisConfig, p *facturacionOperacionPayload, d *dbpkg.EmpresaDocumentoFacturacion) {
			c.PaisCodigo = "EC"
		}},
		{"payload country", func(c *dbpkg.FacturacionElectronicaPaisConfig, p *facturacionOperacionPayload, d *dbpkg.EmpresaDocumentoFacturacion) {
			p.PaisCodigo = "PA"
		}},
		{"missing document country", func(c *dbpkg.FacturacionElectronicaPaisConfig, p *facturacionOperacionPayload, d *dbpkg.EmpresaDocumentoFacturacion) {
			d.PaisCodigo = ""
		}},
		{"sandbox document", func(c *dbpkg.FacturacionElectronicaPaisConfig, p *facturacionOperacionPayload, d *dbpkg.EmpresaDocumentoFacturacion) {
			d.AmbienteFE = "sandbox"
		}},
		{"missing document environment", func(c *dbpkg.FacturacionElectronicaPaisConfig, p *facturacionOperacionPayload, d *dbpkg.EmpresaDocumentoFacturacion) {
			d.AmbienteFE = ""
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &dbpkg.FacturacionElectronicaPaisConfig{EmpresaID: 12, PaisCodigo: "CO", Ambiente: "produccion", Proveedor: "dian", Estado: "activo"}
			payload := facturacionOperacionPayload{EmpresaID: 12, PaisCodigo: "CO"}
			doc := dbpkg.EmpresaDocumentoFacturacion{EmpresaID: 12, PaisCodigo: "CO", AmbienteFE: "produccion", TipoDocumento: "factura_electronica"}
			tc.change(cfg, &payload, &doc)
			// nil DB proves rejection precedes configuration/secret reads and IO.
			got := dispatchFacturacionProveedor(nil, cfg, payload, doc, "emitir", false)
			if got.Success || !got.FinalFailure || got.Error == "" {
				t.Fatalf("invalid fiscal boundary accepted: %#v", got)
			}
		})
	}
}

func TestFacturacionMockCannotAdvertiseProductionConnectivity(t *testing.T) {
	cfg := &dbpkg.FacturacionElectronicaPaisConfig{EmpresaID: 12, PaisCodigo: "CO", Ambiente: "produccion", Proveedor: "externo", APIBaseURL: "mock://ok", Estado: "activo"}
	got := facturacionProveedorConnectionStatus(cfg)
	if got["online"] != false || got["accion_recomendada"] != "bloquear_facturacion_electronica" {
		t.Fatalf("simulation advertised as online: %#v", got)
	}
}
