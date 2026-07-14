package handlers

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type licenciaComprobanteItem struct {
	dbpkg.EmpresaLicenciaPagoResumen
	FacturaDisponible bool `json:"factura_disponible"`
}

// EmpresaLicenciasComprobantesHandler expone compras y documentos de licencia
// exclusivamente para la empresa validada por el middleware empresarial.
func EmpresaLicenciasComprobantesHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		empresaID := parseEmpresaIDFromContext(r)
		if empresaID <= 0 {
			http.Error(w, "empresa no disponible", http.StatusForbidden)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "" || action == "list" {
			handleEmpresaLicenciasComprobantesList(w, r, dbEmp, dbSuper, empresaID)
			return
		}
		if action == "download" {
			handleEmpresaLicenciasComprobantesDownload(w, r, dbEmp, dbSuper, empresaID)
			return
		}
		http.Error(w, "accion no permitida", http.StatusBadRequest)
	}
}

func handleEmpresaLicenciasComprobantesList(w http.ResponseWriter, r *http.Request, dbEmp, dbSuper *sql.DB, empresaID int64) {
	items, err := dbpkg.ListEmpresaLicenciaPagos(dbSuper, empresaID, 50)
	if err != nil {
		http.Error(w, "No se pudo consultar el historial de licencias", http.StatusInternalServerError)
		return
	}
	issuer, _ := dbpkg.GetPowerfulSystemEmpresa(dbEmp, dbSuper)
	out := make([]licenciaComprobanteItem, 0, len(items))
	for _, item := range items {
		entry := licenciaComprobanteItem{EmpresaLicenciaPagoResumen: item}
		if issuer != nil && issuer.EmpresaID > 0 {
			if doc, docErr := findLicenciaFacturaForEmpresa(dbEmp, issuer.EmpresaID, empresaID, item); docErr == nil && doc != nil {
				entry.FacturaDisponible = true
			}
		}
		out = append(out, entry)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "items": out})
}

func handleEmpresaLicenciasComprobantesDownload(w http.ResponseWriter, r *http.Request, dbEmp, dbSuper *sql.DB, empresaID int64) {
	provider := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("provider")))
	paymentID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("payment_id")), 10, 64)
	if err != nil || paymentID <= 0 {
		http.Error(w, "comprobante no valido", http.StatusBadRequest)
		return
	}
	payment, err := dbpkg.GetEmpresaLicenciaPago(dbSuper, empresaID, provider, paymentID)
	if err == sql.ErrNoRows {
		http.Error(w, "comprobante no encontrado", http.StatusNotFound)
		return
	}
	if err != nil || payment == nil {
		http.Error(w, "No se pudo preparar el comprobante", http.StatusInternalServerError)
		return
	}
	empresa, err := dbpkg.GetEmpresaByScopeID(dbEmp, empresaID)
	if err != nil || empresa == nil {
		http.Error(w, "empresa no encontrada", http.StatusNotFound)
		return
	}
	documentType := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("document")))
	var pdf []byte
	var filename string
	if documentType == "factura" {
		issuer, issuerErr := dbpkg.GetPowerfulSystemEmpresa(dbEmp, dbSuper)
		if issuerErr != nil || issuer == nil || issuer.EmpresaID <= 0 {
			http.Error(w, "Factura de licencia no disponible", http.StatusNotFound)
			return
		}
		doc, docErr := findLicenciaFacturaForEmpresa(dbEmp, issuer.EmpresaID, empresaID, *payment)
		if docErr != nil || doc == nil {
			http.Error(w, "Factura de licencia no disponible", http.StatusNotFound)
			return
		}
		pdf, filename = buildLicenciaFacturaElectronicaPDF(*doc, empresa.Nombre, payment.LicenciaNombre, payment.Proveedor, payment.Referencia)
	} else if documentType == "comprobante" {
		pdf, filename = buildLicenciaPaymentReceiptPDF(empresa, *payment)
	} else {
		http.Error(w, "tipo de documento no permitido", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdf)
	registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, "licencias", "documento_descargado", "pagos_licencia", payment.ID, http.StatusOK, map[string]interface{}{
		"proveedor": payment.Proveedor,
		"documento": documentType,
	}, "descarga de documento de compra de licencia")
}

func findLicenciaFacturaForEmpresa(dbEmp *sql.DB, issuerEmpresaID, buyerEmpresaID int64, payment dbpkg.EmpresaLicenciaPagoResumen) (*dbpkg.EmpresaDocumentoFacturacion, error) {
	code := buildLicenciaFacturaDocumentoCodigo(payment.Proveedor, payment.Referencia, payment.LicenciaID, buyerEmpresaID)
	doc, err := dbpkg.GetEmpresaDocumentoFacturacionByCodigo(dbEmp, issuerEmpresaID, "factura_electronica", code)
	if err != nil || doc == nil {
		return nil, err
	}
	if doc.EntidadRelacionadaID != buyerEmpresaID {
		return nil, sql.ErrNoRows
	}
	return doc, nil
}

func buildLicenciaPaymentReceiptPDF(empresa *dbpkg.Empresa, payment dbpkg.EmpresaLicenciaPagoResumen) ([]byte, string) {
	companyName := "Empresa usuaria"
	if empresa != nil && strings.TrimSpace(empresa.Nombre) != "" {
		companyName = strings.TrimSpace(empresa.Nombre)
	}
	status := strings.TrimSpace(payment.Estado)
	if status == "" {
		status = "registrado"
	}
	var content bytes.Buffer
	pdfLine(&content, "q 0 0 0 RG 0.90 0.95 1 rg 1 w 54 781 38 28 re B Q")
	pdfText(&content, "F2", 13, 62, 790, "PCS")
	pdfLine(&content, "q 0 0 0 RG 1.4 w 46 760 m 548 760 l S Q")
	pdfText(&content, "F2", 22, 104, 790, "Powerful Control System")
	pdfText(&content, "F2", 15, 54, 742, "Comprobante de compra de licencia")
	lines := []string{
		"Empresa: " + companyName,
		"Licencia: " + emptyPDFValue(payment.LicenciaNombre, "Licencia del sistema"),
		"Pasarela: " + emptyPDFValue(strings.ToUpper(payment.Proveedor), "Sistema"),
		"Referencia: " + emptyPDFValue(payment.Referencia, "Sin referencia"),
		"Transaccion: " + emptyPDFValue(payment.TransaccionID, "Pendiente de asignacion"),
		"Estado reportado: " + status,
		"Fecha de registro: " + emptyPDFValue(payment.FechaCreacion, time.Now().Format("2006-01-02 15:04:05")),
	}
	y := 710
	for _, line := range lines {
		pdfText(&content, "F1", 10, 54, y, line)
		y -= 16
	}
	y -= 10
	for _, line := range wrapPDFText("Este comprobante acredita el registro de la solicitud o pago en la pasarela indicada. La factura electronica se habilita para descarga cuando su emision fiscal quede registrada.", 92) {
		pdfText(&content, "F1", 10, 54, y, line)
		y -= 13
	}
	pdfLine(&content, "q 0 0 0 RG 0.8 w 46 52 m 548 52 l S Q")
	pdfText(&content, "F1", 8, 54, 38, "Documento generado automaticamente por Powerful Control System.")
	filenameRef := strings.TrimSpace(payment.Referencia)
	if filenameRef == "" {
		filenameRef = fmt.Sprintf("pago-%d", payment.ID)
	}
	base := strings.TrimSuffix(licenciaFacturaElectronicaPDFFilename(filenameRef), ".pdf")
	base = strings.TrimPrefix(base, "factura-electronica-")
	return assembleSimplePDF(content.Bytes()), "comprobante-licencia-" + base + ".pdf"
}
