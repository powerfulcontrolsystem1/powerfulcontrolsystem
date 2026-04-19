package handlers

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func TestMotelCalisoVentaPDF(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_motel_caliso.db")
	ensureClientesSchema(t, dbEmp)
	ensureCarritosVentasSchema(t, dbEmp)
	if err := dbpkg.EnsureEmpresaFacturacionElectronicaSchema(dbEmp); err != nil {
		t.Fatalf("ensure facturacion schema: %v", err)
	}

	// Ensure categories and products schema
	// if err := dbpkg.EnsureEmpresasCategoriasSchema(dbEmp); err != nil {
	// 	t.Fatalf("ensure categorias: %v", err)
	// }
	// if err := dbpkg.EnsureEmpresasProductosSchema(dbEmp); err != nil {
	// 	t.Fatalf("ensure productos: %v", err)
	// }

	// Create mock receipt data
	receiptLines := []string{
		"----------------------------------------",
		"             MOTEL CALISO               ",
		"        NIT: 900.123.456-7              ",
		"    Dir: Calle 1 # 2-3, La Ciudad       ",
		"----------------------------------------",
		fmt.Sprintf("Fecha: %s", time.Now().Format("2006-01-02 15:04:05")),
		"Estacion: Habitacion 10",
		"----------------------------------------",
		fmt.Sprintf("%-20s %10s", "Habitacion", "$ 45,000.00"),
		fmt.Sprintf("%-20s %10s", "Agua Mineral 500ml", "$  3,000.00"),
		fmt.Sprintf("%-20s %10s", "Servicio Cuarto", "$  5,000.00"),
		"----------------------------------------",
		fmt.Sprintf("%-20s %10s", "TOTAL:", "$ 53,000.00"),
		"----------------------------------------",
		"   Gracias por su discreta visita!      ",
		"----------------------------------------",
	}

	// Simple PDF manual generation
	pdfData := generateSimplePDF(receiptLines)
	outF := "../Motel_Caliso_Factura_10.pdf"
	if err := os.WriteFile(outF, pdfData, 0644); err != nil {
		t.Fatalf("Could not write PDF file: %v", err)
	}
	t.Logf("Factura guardada exitosamente en %s", outF)
}

func generateSimplePDF(lines []string) []byte {
	var streamBuilder bytes.Buffer
	streamBuilder.WriteString("q\nBT\n/F1 12 Tf\n10 750 Td\n15 TL\n")
	for _, line := range lines {
		escaped := strings.ReplaceAll(line, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "(", "\\(")
		escaped = strings.ReplaceAll(escaped, ")", "\\)")
		streamBuilder.WriteString(fmt.Sprintf("(%s) Tj T*\n", escaped))
	}
	streamBuilder.WriteString("ET\nQ\n")
	stream := streamBuilder.String()

	var pdf bytes.Buffer
	offsets := make([]int, 6)
	pdf.WriteString("%PDF-1.4\n")
	offsets[1] = pdf.Len()
	pdf.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")
	offsets[2] = pdf.Len()
	pdf.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")
	offsets[3] = pdf.Len()
	pdf.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>\nendobj\n")
	offsets[4] = pdf.Len()
	pdf.WriteString(fmt.Sprintf("4 0 obj\n<< /Length %d >>\nstream\n%sendstream\nendobj\n", len(stream), stream))
	offsets[5] = pdf.Len()
	pdf.WriteString("5 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Courier >>\nendobj\n")
	
	startXRef := pdf.Len()
	pdf.WriteString("xref\n0 6\n")
	pdf.WriteString("0000000000 65535 f \n")
	for i := 1; i <= 5; i++ {
		pdf.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}
	pdf.WriteString("trailer\n<< /Size 6 /Root 1 0 R >>\n")
	pdf.WriteString(fmt.Sprintf("startxref\n%d\n%%%%EOF", startXRef))
	return pdf.Bytes()
}
