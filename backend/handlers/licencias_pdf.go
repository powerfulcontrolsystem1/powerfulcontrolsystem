package handlers

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type licenciaSoftwarePDFInput struct {
	Title       string
	Body        string
	CompanyName string
	LicenseName string
	LicenseCode string
	IssuedAt    time.Time
}

func buildLicenciaSoftwarePDFForEmpresa(dbSuper *sql.DB, empresa *dbpkg.Empresa, lic *dbpkg.Licencia, provider, reference, amountPaid string, issuedAt time.Time) ([]byte, string, error) {
	if empresa == nil {
		return nil, "", fmt.Errorf("empresa requerida")
	}
	if lic == nil {
		return nil, "", fmt.Errorf("licencia requerida")
	}
	if issuedAt.IsZero() {
		issuedAt = time.Now()
	}
	safeEmpresa := strings.TrimSpace(empresa.Nombre)
	if safeEmpresa == "" {
		safeEmpresa = fmt.Sprintf("Empresa %d", empresa.EmpresaID)
	}
	effectiveEmpresaID := empresa.EmpresaID
	if effectiveEmpresaID <= 0 {
		effectiveEmpresaID = empresa.ID
	}
	provider = strings.TrimSpace(provider)
	if provider == "" {
		provider = "Sistema"
	}
	reference = strings.TrimSpace(reference)
	if reference == "" {
		reference = "Descarga desde Administrar empresa"
	}
	amountPaid = strings.TrimSpace(amountPaid)
	if amountPaid == "" {
		amountPaid = fmt.Sprintf("%.0f", lic.Valor)
	}
	licenseCode := fmt.Sprintf("PCS-%d-%d", effectiveEmpresaID, lic.ID)
	pdfTitle, pdfBody, _, err := applySuperEmailTemplate(dbSuper, superEmailTemplateKeyLicenciaSoftwarePDF, map[string]string{
		"company_name": safeEmpresa,
		"company_nit":  strings.TrimSpace(empresa.Nit),
		"license_name": strings.TrimSpace(lic.Nombre),
		"license_code": licenseCode,
		"provider":     provider,
		"reference":    reference,
		"issue_date":   issuedAt.Format("2006-01-02 15:04:05"),
		"start_date":   strings.TrimSpace(lic.FechaInicio),
		"end_date":     strings.TrimSpace(lic.FechaFin),
		"amount_paid":  amountPaid,
		"system_name":  "Powerful Control System",
	})
	if err != nil {
		return nil, "", err
	}
	pdfBytes := buildLicenciaSoftwarePDF(licenciaSoftwarePDFInput{
		Title:       pdfTitle,
		Body:        pdfBody,
		CompanyName: safeEmpresa,
		LicenseName: strings.TrimSpace(lic.Nombre),
		LicenseCode: licenseCode,
		IssuedAt:    issuedAt,
	})
	return pdfBytes, licenciaSoftwarePDFFilename(safeEmpresa, effectiveEmpresaID), nil
}

func EmpresaLicenciaSistemaPDFHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		empresaID := parseEmpresaIDFromContext(r)
		if empresaID <= 0 {
			if id, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && id > 0 {
				empresaID = id
			}
		}
		if empresaID <= 0 {
			http.Error(w, "empresa_id requerido", http.StatusBadRequest)
			return
		}
		empresa, err := dbpkg.GetEmpresaByScopeID(dbEmp, empresaID)
		if err != nil || empresa == nil {
			http.Error(w, "empresa no encontrada", http.StatusNotFound)
			return
		}
		lic, err := dbpkg.GetActiveLicenciaByEmpresa(dbSuper, empresaID)
		if err != nil || lic == nil {
			http.Error(w, "No hay licencia activa para descargar", http.StatusNotFound)
			return
		}
		pdfBytes, filename, err := buildLicenciaSoftwarePDFForEmpresa(dbSuper, empresa, lic, "Sistema", "Descarga desde Administrar empresa", fmt.Sprintf("%.0f", lic.Valor), time.Now())
		if err != nil {
			http.Error(w, "No se pudo generar la licencia en PDF", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(pdfBytes)
	}
}

func buildLicenciaSoftwarePDF(input licenciaSoftwarePDFInput) []byte {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = "Licencia de software Powerful Control System"
	}
	body := strings.TrimSpace(input.Body)
	if body == "" {
		body = "Licencia de uso del software Powerful Control System."
	}
	issuedAt := input.IssuedAt
	if issuedAt.IsZero() {
		issuedAt = time.Now()
	}

	var content bytes.Buffer
	pdfLine(&content, "q 0 0 0 RG 0.90 0.95 1 rg 1 w 54 781 38 28 re B Q")
	pdfText(&content, "F2", 13, 62, 790, "PCS")
	pdfLine(&content, "q 0 0 0 RG 1.4 w 46 760 m 548 760 l S Q")
	pdfText(&content, "F2", 22, 104, 790, "Powerful Control System")
	pdfText(&content, "F1", 9, 105, 775, "Sistema de facturacion electronica con domotica integrada")
	pdfText(&content, "F2", 15, 54, 742, title)

	y := 715
	meta := []string{
		"Empresa: " + emptyPDFValue(input.CompanyName, "Empresa usuaria"),
		"Licencia: " + emptyPDFValue(input.LicenseName, "Plan activo"),
		"Codigo: " + emptyPDFValue(input.LicenseCode, "LICENCIA"),
		"Emitida: " + issuedAt.Format("2006-01-02 15:04:05"),
	}
	for _, line := range meta {
		pdfText(&content, "F1", 10, 54, y, line)
		y -= 15
	}
	y -= 8

	for _, paragraph := range splitPDFParagraphs(body) {
		if y < 72 {
			break
		}
		lines := wrapPDFText(paragraph, 92)
		if len(lines) == 0 {
			y -= 10
			continue
		}
		for _, line := range lines {
			if y < 72 {
				break
			}
			font := "F1"
			size := 10
			if isPDFHeading(line) {
				font = "F2"
				size = 11
			}
			pdfText(&content, font, size, 54, y, line)
			y -= 13
		}
		y -= 6
	}

	pdfLine(&content, "q 0 0 0 RG 0.8 w 46 52 m 548 52 l S Q")
	pdfText(&content, "F1", 8, 54, 38, "Documento generado automaticamente por Powerful Control System.")

	return assembleSimplePDF(content.Bytes())
}

func assembleSimplePDF(content []byte) []byte {
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 4 0 R /F2 5 0 R >> >> /Contents 6 0 R >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content),
	}

	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n")
	offsets := make([]int, 0, len(objects)+1)
	offsets = append(offsets, 0)
	for i, obj := range objects {
		offsets = append(offsets, out.Len())
		fmt.Fprintf(&out, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}
	xrefAt := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n", len(objects)+1)
	out.WriteString("0000000000 65535 f \n")
	for i := 1; i < len(offsets); i++ {
		fmt.Fprintf(&out, "%010d 00000 n \n", offsets[i])
	}
	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xrefAt)
	return out.Bytes()
}

func pdfText(buf *bytes.Buffer, font string, size, x, y int, text string) {
	text = normalizePDFText(text)
	fmt.Fprintf(buf, "BT /%s %d Tf %d %d Td (%s) Tj ET\n", font, size, x, y, escapeLicensePDFString(text))
}

func pdfLine(buf *bytes.Buffer, command string) {
	buf.WriteString(command)
	buf.WriteByte('\n')
}

func splitPDFParagraphs(input string) []string {
	input = strings.ReplaceAll(strings.ReplaceAll(input, "\r\n", "\n"), "\r", "\n")
	raw := strings.Split(input, "\n")
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		out = append(out, strings.TrimSpace(item))
	}
	return out
}

func wrapPDFText(input string, max int) []string {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}
	words := strings.Fields(input)
	if len(words) == 0 {
		return nil
	}
	lines := make([]string, 0, 4)
	current := ""
	for _, word := range words {
		if current == "" {
			current = word
			continue
		}
		if len([]rune(current))+1+len([]rune(word)) > max {
			lines = append(lines, current)
			current = word
			continue
		}
		current += " " + word
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func escapeLicensePDFString(input string) string {
	input = strings.ReplaceAll(input, `\`, `\\`)
	input = strings.ReplaceAll(input, "(", `\(`)
	input = strings.ReplaceAll(input, ")", `\)`)
	return input
}

func normalizePDFText(input string) string {
	replacer := strings.NewReplacer(
		"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ñ", "n",
		"Á", "A", "É", "E", "Í", "I", "Ó", "O", "Ú", "U", "Ñ", "N",
		"ü", "u", "Ü", "U", "¿", "", "¡", "", "–", "-", "—", "-",
	)
	input = replacer.Replace(input)
	return regexp.MustCompile(`[^\x09\x0A\x0D\x20-\x7E]`).ReplaceAllString(input, "")
}

func isPDFHeading(input string) bool {
	text := strings.TrimSpace(input)
	if text == "" || len(text) > 64 {
		return false
	}
	return !strings.Contains(text, ".") && !strings.Contains(text, ":") && strings.EqualFold(text, strings.ToUpper(text))
}

func emptyPDFValue(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func licenciaSoftwarePDFFilename(companyName string, empresaID int64) string {
	base := strings.ToLower(normalizePDFText(strings.TrimSpace(companyName)))
	base = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(base, "-")
	base = strings.Trim(base, "-")
	if base == "" {
		base = fmt.Sprintf("empresa-%d", empresaID)
	}
	if len(base) > 64 {
		base = strings.Trim(base[:64], "-")
	}
	return "licencia-powerful-control-system-" + base + ".pdf"
}
