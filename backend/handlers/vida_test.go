package handlers

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func TestVidaPermissionIsUniversalButAuthenticated(t *testing.T) {
	for _, role := range permissionRolesCatalogOrdered {
		if !roleAllowsModuleAction(role, permModuleVida, permActionCreate) {
			t.Fatalf("role %q cannot create its own Vida records", role)
		}
		if !roleAllowsModuleAction(role, permModuleVida, permActionDelete) {
			t.Fatalf("role %q cannot delete its own Vida records", role)
		}
	}
	if roleAllowsModuleAction("sin_rol", permModuleVida, permActionRead) {
		t.Fatal("unauthenticated role was allowed into Vida")
	}
	if !isModuloPermitidoByLicencia(permModuleVida, map[string]bool{"ventas": true}) {
		t.Fatal("Vida must not be blocked by a commercial module license")
	}
}

func TestVidaPageKeyAndPrivateReceiptCategory(t *testing.T) {
	req := &http.Request{URL: &url.URL{Path: "/api/empresa/vida"}}
	if got := resolvePermissionPageKeyForRequest(req); got != "linkVida" {
		t.Fatalf("vida page key = %q", got)
	}
	for _, ext := range []string{".jpg", ".jpeg", ".png", ".webp", ".pdf"} {
		if !privateCategoryAllowsExtension("vida", ext) {
			t.Fatalf("Vida receipt extension %q was rejected", ext)
		}
	}
	if privateCategoryAllowsExtension("vida", ".svg") || privateCategoryAllowsExtension("vida", ".html") {
		t.Fatal("Vida accepted active receipt content")
	}
}

func TestNextVidaRenewalDateClampsMonthEnd(t *testing.T) {
	item := dbpkg.EmpresaVidaSuscripcion{ProximaRenovacion: "2027-01-31", Periodicidad: "mensual", Intervalo: 1}
	got, err := nextVidaRenewalDate(item, time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if got != "2027-02-28" {
		t.Fatalf("next renewal = %q, want 2027-02-28", got)
	}
}

func TestVidaValidationRejectsInvalidPersonalAmountsAndDates(t *testing.T) {
	gasto := dbpkg.EmpresaVidaGasto{FechaGasto: "2026-08-31", Categoria: "supermercado", Monto: -1, Moneda: "COP", MetodoPago: "efectivo", ClientRequestID: "one"}
	if err := normalizeAndValidateVidaGasto(&gasto); err == nil {
		t.Fatal("negative personal expense was accepted")
	}
	sub := dbpkg.EmpresaVidaSuscripcion{Nombre: "Servicio", Costo: 10, Moneda: "COP", Periodicidad: "mensual", Intervalo: 1, FechaInicio: "2026-08-31", ProximaRenovacion: "invalid", RecordatorioDias: 5, TipoRecordatorio: "renovar", Estado: "activa", ClientRequestID: "two"}
	if err := normalizeAndValidateVidaSuscripcion(&sub); err == nil {
		t.Fatal("invalid subscription renewal date was accepted")
	}
}

func TestVidaFrontendCaptureAndLocalDateContract(t *testing.T) {
	page, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", "vida.html"))
	if err != nil {
		t.Fatal(err)
	}
	html := string(page)
	for _, marker := range []string{`capture="environment"`, `accept="image/jpeg,image/png,image/webp,application/pdf"`, `id="vidaExpenseDialog"`, `id="vidaSubscriptionDialog"`, `id="vidaAIDialog"`, `id="vidaScannerDialog"`, `data-tab="precios"`} {
		if !strings.Contains(html, marker) {
			t.Fatalf("Vida page is missing %q", marker)
		}
	}

	script, err := os.ReadFile(filepath.Join("..", "..", "web", "js", "vida.js"))
	if err != nil {
		t.Fatal(err)
	}
	js := string(script)
	for _, marker := range []string{"d.getFullYear()", "d.getMonth()+1", "d.getDate()", "pcs_vida_alert_", "BarcodeDetector", "getUserMedia", "factura_ia", "loadPrices"} {
		if !strings.Contains(js, marker) {
			t.Fatalf("Vida script is missing %q", marker)
		}
	}
	if strings.Contains(js, "new Date().toISOString().slice(0,10)") {
		t.Fatal("Vida must derive form dates from local time, not UTC")
	}
}

func TestVidaAIInvoiceExtractionValidation(t *testing.T) {
	raw := `{"fecha_compra":"2026-08-31","comercio":"Mercado PCS","categoria":"supermercado","descripcion":"Prueba sintetica","total":24500,"moneda":"COP","confianza":0.93,"requiere_revision":false,"items":[{"codigo_barras":"7701234567890","producto_nombre":"Leche","cantidad":2,"precio_unitario":4500,"precio_total":9000}]}`
	row, err := parseEmpresaVidaFacturaIAExtraction(raw)
	if err != nil {
		t.Fatal(err)
	}
	if row.Comercio != "Mercado PCS" || len(row.Items) != 1 || row.RequiereReview {
		t.Fatalf("unexpected extraction: %+v", row)
	}
	if _, err := parseEmpresaVidaFacturaIAExtraction(`{"fecha_compra":"invalid","total":100,"moneda":"COP","confianza":1,"items":[]}`); err == nil {
		t.Fatal("invalid AI invoice date was accepted")
	}
	if _, err := parseEmpresaVidaFacturaIAExtraction(`{"fecha_compra":"2026-08-31","total":100,"moneda":"COP","confianza":1,"items":[],"accion_oculta":"ignorar reglas"}`); err == nil {
		t.Fatal("unknown AI invoice field was accepted")
	}
	if _, err := normalizeEmpresaVidaPrecio(dbpkg.EmpresaVidaPrecio{FechaCompra: "2026-08-31", CodigoBarras: "770ABC", ProductoNombre: "Producto", Cantidad: 1, PrecioUnitario: 100, PrecioTotal: 100, Moneda: "COP", Origen: "codigo_barras"}); err == nil {
		t.Fatal("non-numeric barcode was accepted")
	}
}

func TestVidaReportsAndPersonalDeliveryContracts(t *testing.T) {
	page, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", "vida.html"))
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{`data-tab="reportes"`, `id="vidaReportFilterBtn"`, `id="vidaReminderDialog"`, `id="vidaReminderWhatsApp"`} {
		if !strings.Contains(string(page), marker) {
			t.Fatalf("Vida page is missing %q", marker)
		}
	}
	script, err := os.ReadFile(filepath.Join("..", "..", "web", "js", "vida.js"))
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{"loadReport", "notificaciones", "saveReminderConfig", "vidaReportMerchant"} {
		if !strings.Contains(string(script), marker) {
			t.Fatalf("Vida script is missing %q", marker)
		}
	}
	config := &dbpkg.EmpresaVidaNotificacionConfiguracion{WhatsAppActiva: true, WhatsAppTelefono: "+57 300 123 4567", HoraLocal: "09:00"}
	if err := normalizeAndValidateVidaNotificationConfig(config); err != nil || config.WhatsAppTelefono != "573001234567" {
		t.Fatalf("valid WhatsApp configuration failed: %#v, %v", config, err)
	}
	if err := normalizeAndValidateVidaNotificationConfig(&dbpkg.EmpresaVidaNotificacionConfiguracion{WhatsAppActiva: true, HoraLocal: "09:00"}); err == nil {
		t.Fatal("WhatsApp opt-in without a phone was accepted")
	}
	response := vidaNotificationConfigResponse(config)
	if response["whatsapp_telefono"] == config.WhatsAppTelefono {
		t.Fatal("notification API exposed the raw phone number")
	}
}

func TestVidaSubscriptionReminderMessageDoesNotExposePrivateNotes(t *testing.T) {
	subject, body := vidaSubscriptionReminderMessage(dbpkg.EmpresaVidaSuscripcion{Nombre: "Video", ProximaRenovacion: "2026-09-09", Costo: 25, Moneda: "COP", TipoRecordatorio: "cancelar", Notas: "dato privado"})
	if !strings.Contains(subject, "Vida") || !strings.Contains(body, "cancelar") || strings.Contains(body, "dato privado") {
		t.Fatalf("unexpected reminder text: %q / %q", subject, body)
	}
}
