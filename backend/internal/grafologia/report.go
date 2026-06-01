package grafologia

import (
	"bytes"
	"fmt"
	"html"
	"strings"
)

func RenderHTMLReport(title string, result AnalysisResult) string {
	if strings.TrimSpace(title) == "" {
		title = "Informe grafológico GRAFOLOGIX"
	}
	var b strings.Builder
	b.WriteString(`<!doctype html><html lang="es"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">`)
	b.WriteString(`<title>`)
	b.WriteString(html.EscapeString(title))
	b.WriteString(`</title><style>`)
	b.WriteString(`:root{font-family:Arial,sans-serif;color:#111;background:#fff}body{margin:0;padding:24px;background:#fff}.sheet{max-width:900px;margin:0 auto}.top{border-bottom:2px solid #111;padding-bottom:12px;margin-bottom:16px}.muted{color:#555}.grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:8px 18px}.row{display:flex;justify-content:space-between;border-bottom:1px solid #ddd;padding:7px 0}.bar{height:8px;background:#eee;border-radius:999px;overflow:hidden}.fill{height:100%;background:#111}.section{margin-top:20px}.note{font-size:12px;color:#333;border-top:1px solid #ddd;margin-top:20px;padding-top:12px}@media print{body{padding:0}.sheet{max-width:none}.no-print{display:none}}`)
	b.WriteString(`</style></head><body><main class="sheet">`)
	b.WriteString(`<button class="no-print" onclick="window.print()">Imprimir / guardar PDF</button>`)
	b.WriteString(`<section class="top"><h1>`)
	b.WriteString(html.EscapeString(title))
	b.WriteString(`</h1><p class="muted">Generado por GRAFOLOGIX · `)
	b.WriteString(html.EscapeString(result.GeneratedAt))
	b.WriteString(` · Confianza global `)
	b.WriteString(formatPercent(result.GlobalTrust))
	b.WriteString(`</p><p>`)
	b.WriteString(html.EscapeString(result.Summary))
	b.WriteString(`</p></section>`)
	b.WriteString(`<section class="section"><h2>Resumen técnico</h2><div class="grid">`)
	summaryRows := [][2]string{
		{"Resolución", intString(result.Image.Width) + " x " + intString(result.Image.Height)},
		{"Densidad de tinta", formatPercent(result.Image.InkDensity * 100)},
		{"Contraste", formatPercent(result.Image.Contrast)},
		{"Líneas detectadas", intString(result.Image.LinesDetected)},
		{"Palabras estimadas", intString(result.Image.WordsEstimated)},
		{"Letras estimadas", intString(result.Image.LettersEstimated)},
	}
	for _, row := range summaryRows {
		b.WriteString(`<div class="row"><strong>`)
		b.WriteString(html.EscapeString(row[0]))
		b.WriteString(`</strong><span>`)
		b.WriteString(html.EscapeString(row[1]))
		b.WriteString(`</span></div>`)
	}
	b.WriteString(`</div></section>`)
	if result.Preprocess != nil {
		b.WriteString(`<section class="section"><h2>Calidad de imagen</h2><div class="grid">`)
		qualityRows := [][2]string{
			{"Contraste", formatPercent(result.Preprocess.Quality.Contrast)},
			{"Densidad de tinta", formatPercent(result.Preprocess.Quality.InkDensity * 100)},
			{"Nitidez estimada", formatPercent(result.Preprocess.Quality.Sharpness)},
			{"Umbral Otsu", intString(result.Preprocess.Threshold)},
			{"Sugerencia de recorte", yesNo(result.Preprocess.Quality.CropSuggested)},
			{"Advertencia de iluminación", yesNo(result.Preprocess.Quality.LightingWarning)},
			{"Advertencia de resolución", yesNo(result.Preprocess.Quality.ResolutionWarning)},
			{"Caja de tinta", intString(result.Preprocess.InkBox.MinX) + "," + intString(result.Preprocess.InkBox.MinY) + " - " + intString(result.Preprocess.InkBox.MaxX) + "," + intString(result.Preprocess.InkBox.MaxY)},
		}
		for _, row := range qualityRows {
			b.WriteString(`<div class="row"><strong>`)
			b.WriteString(html.EscapeString(row[0]))
			b.WriteString(`</strong><span>`)
			b.WriteString(html.EscapeString(row[1]))
			b.WriteString(`</span></div>`)
		}
		b.WriteString(`</div></section>`)
	}
	b.WriteString(`<section class="section"><h2>Tabla de métricas</h2>`)
	for _, m := range result.Metrics {
		b.WriteString(`<div class="row"><div><strong>`)
		b.WriteString(html.EscapeString(m.Name))
		b.WriteString(`</strong><br><span class="muted">`)
		b.WriteString(html.EscapeString(m.Explanation))
		b.WriteString(`</span></div><span>`)
		b.WriteString(html.EscapeString(m.Value))
		b.WriteString(` · `)
		b.WriteString(formatPercent(m.Confidence))
		b.WriteString(`</span></div>`)
	}
	b.WriteString(`</section>`)
	if result.Preprocess != nil && len(result.Preprocess.ImageURLs) > 0 {
		b.WriteString(`<section class="section"><h2>Preprocesamiento visual</h2><div class="grid">`)
		for _, key := range []string{"grayscale", "binary", "edges", "lines"} {
			if url := strings.TrimSpace(result.Preprocess.ImageURLs[key]); url != "" {
				b.WriteString(`<div><strong>`)
				b.WriteString(html.EscapeString(preprocessLabel(key)))
				b.WriteString(`</strong><br><img alt="`)
				b.WriteString(html.EscapeString(preprocessLabel(key)))
				b.WriteString(`" src="`)
				b.WriteString(html.EscapeString(url))
				b.WriteString(`" style="width:100%;max-height:230px;object-fit:contain;border:1px solid #ddd;margin-top:6px"></div>`)
			}
		}
		b.WriteString(`</div></section>`)
	}
	b.WriteString(`<section class="section"><h2>Interpretación orientativa</h2>`)
	for _, t := range result.Traits {
		b.WriteString(`<div class="row"><div style="flex:1;padding-right:14px"><strong>`)
		b.WriteString(html.EscapeString(t.Name))
		b.WriteString(`</strong><br><span class="muted">`)
		b.WriteString(html.EscapeString(t.Explanation))
		b.WriteString(`</span><div class="bar"><div class="fill" style="width:`)
		b.WriteString(formatPercent(t.Percent))
		b.WriteString(`"></div></div></div><span>`)
		b.WriteString(formatPercent(t.Percent))
		b.WriteString(` · `)
		b.WriteString(html.EscapeString(t.Level))
		b.WriteString(`</span></div>`)
	}
	b.WriteString(`</section><section class="section"><h2>Conclusión</h2><p>`)
	b.WriteString(html.EscapeString(conclusion(result)))
	b.WriteString(`</p></section><section class="note">`)
	for _, note := range result.TechnicalNotes {
		b.WriteString(`<p>`)
		b.WriteString(html.EscapeString(note))
		b.WriteString(`</p>`)
	}
	b.WriteString(`</section></main></body></html>`)
	return b.String()
}

func RenderWordReport(title string, result AnalysisResult) []byte {
	htmlDoc := RenderHTMLReport(title, result)
	htmlDoc = strings.Replace(htmlDoc, `<!doctype html><html`, `<html xmlns:o="urn:schemas-microsoft-com:office:office" xmlns:w="urn:schemas-microsoft-com:office:word" xmlns="http://www.w3.org/TR/REC-html40"`, 1)
	return []byte(htmlDoc)
}

func RenderTextReport(title string, result AnalysisResult) string {
	if strings.TrimSpace(title) == "" {
		title = "Informe grafologico GRAFOLOGIX"
	}
	var b strings.Builder
	b.WriteString(cleanPDFText(title) + "\n")
	b.WriteString("Generado por GRAFOLOGIX\n")
	b.WriteString("Fecha: " + result.GeneratedAt + "\n")
	b.WriteString("Confianza global: " + formatPercent(result.GlobalTrust) + "\n\n")
	b.WriteString("RESUMEN GENERAL\n")
	b.WriteString(cleanPDFText(result.Summary) + "\n\n")
	b.WriteString("METRICAS\n")
	for _, m := range result.Metrics {
		b.WriteString("- " + cleanPDFText(m.Name) + ": " + cleanPDFText(m.Value) + " | Confianza " + formatPercent(m.Confidence) + "\n")
		b.WriteString("  " + cleanPDFText(m.Explanation) + "\n")
	}
	b.WriteString("\nINTERPRETACION ORIENTATIVA\n")
	for _, t := range result.Traits {
		b.WriteString("- " + cleanPDFText(t.Name) + ": " + formatPercent(t.Percent) + " | " + cleanPDFText(t.Level) + "\n")
		b.WriteString("  " + cleanPDFText(t.Explanation) + "\n")
	}
	if result.Preprocess != nil {
		b.WriteString("\nCALIDAD DE IMAGEN\n")
		b.WriteString("- Contraste: " + formatPercent(result.Preprocess.Quality.Contrast) + "\n")
		b.WriteString("- Densidad de tinta: " + formatPercent(result.Preprocess.Quality.InkDensity*100) + "\n")
		b.WriteString("- Nitidez estimada: " + formatPercent(result.Preprocess.Quality.Sharpness) + "\n")
		b.WriteString("- Umbral Otsu: " + intString(result.Preprocess.Threshold) + "\n")
	}
	b.WriteString("\nOBSERVACIONES TECNICAS\n")
	for _, note := range result.TechnicalNotes {
		b.WriteString("- " + cleanPDFText(note) + "\n")
	}
	return b.String()
}

func RenderCSVReport(result AnalysisResult) string {
	var b strings.Builder
	b.WriteString("tipo,clave,nombre,valor,nivel_o_categoria,puntaje,confianza,explicacion\n")
	for _, m := range result.Metrics {
		writeCSVRow(&b, []string{"metrica", m.Key, m.Name, m.Value, m.Category, sprintf2(m.Score), sprintf2(m.Confidence), m.Explanation})
	}
	for _, t := range result.Traits {
		writeCSVRow(&b, []string{"interpretacion", t.Key, t.Name, sprintf2(t.Percent), t.Level, sprintf2(t.Percent), sprintf2(t.Confidence), t.Explanation})
	}
	return b.String()
}

func RenderPDFReport(title string, result AnalysisResult) []byte {
	if strings.TrimSpace(title) == "" {
		title = "Informe grafologico GRAFOLOGIX"
	}
	lines := []string{
		title,
		"Generado por GRAFOLOGIX",
		"Fecha: " + result.GeneratedAt,
		"Confianza global: " + formatPercent(result.GlobalTrust),
		"",
		"Resumen general",
		result.Summary,
		"",
		"Metricas principales",
	}
	for _, m := range result.Metrics {
		lines = append(lines, "- "+m.Name+": "+m.Value+" ("+formatPercent(m.Confidence)+")")
	}
	lines = append(lines, "", "Interpretacion orientativa")
	for _, t := range result.Traits {
		lines = append(lines, "- "+t.Name+": "+formatPercent(t.Percent)+" - "+t.Level)
	}
	lines = append(lines, "", "Observaciones tecnicas")
	for _, note := range result.TechnicalNotes {
		lines = append(lines, "- "+note)
	}
	if len(lines) > 58 {
		lines = append(lines[:57], "... informe resumido; use HTML para el detalle completo.")
	}

	var content strings.Builder
	content.WriteString("BT\n/F1 11 Tf\n50 800 Td\n14 TL\n")
	for _, line := range lines {
		for _, wrapped := range wrapPDFLine(cleanPDFText(line), 92) {
			content.WriteString("(")
			content.WriteString(escapePDFString(wrapped))
			content.WriteString(") Tj\nT*\n")
		}
	}
	content.WriteString("ET\n")

	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content.String()), content.String()),
	}
	var pdf bytes.Buffer
	pdf.WriteString("%PDF-1.4\n")
	offsets := make([]int, 0, len(objects)+1)
	offsets = append(offsets, 0)
	for i, obj := range objects {
		offsets = append(offsets, pdf.Len())
		pdf.WriteString(fmt.Sprintf("%d 0 obj\n%s\nendobj\n", i+1, obj))
	}
	xref := pdf.Len()
	pdf.WriteString(fmt.Sprintf("xref\n0 %d\n0000000000 65535 f \n", len(objects)+1))
	for i := 1; i < len(offsets); i++ {
		pdf.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}
	pdf.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref))
	return pdf.Bytes()
}

func writeCSVRow(b *strings.Builder, values []string) {
	for i, value := range values {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteByte('"')
		b.WriteString(strings.ReplaceAll(cleanPDFText(value), `"`, `""`))
		b.WriteByte('"')
	}
	b.WriteByte('\n')
}

func yesNo(value bool) string {
	if value {
		return "Si"
	}
	return "No"
}

func conclusion(result AnalysisResult) string {
	if len(result.Traits) == 0 {
		return "La muestra requiere una imagen de mayor calidad para producir una conclusión útil."
	}
	top := result.Traits[0]
	return "La lectura destaca " + strings.ToLower(top.Name) + " con nivel " + strings.ToLower(top.Level) + ". Debe usarse como apoyo exploratorio y contrastarse con información humana y operativa adicional."
}

func formatPercent(v float64) string {
	return strings.TrimRight(strings.TrimRight(floatString(v), "0"), ".") + "%"
}

func floatString(v float64) string {
	return strings.ReplaceAll(strings.TrimRight(strings.TrimRight(sprintf2(v), "0"), "."), ",", ".")
}

func sprintf2(v float64) string {
	const digits = "0123456789"
	if v < 0 {
		return "-" + sprintf2(-v)
	}
	n := int(v*100 + 0.5)
	whole := n / 100
	frac := n % 100
	if frac < 10 {
		return intString(whole) + ".0" + string(digits[frac])
	}
	return intString(whole) + "." + string(digits[frac/10]) + string(digits[frac%10])
}

func intString(v int) string {
	if v == 0 {
		return "0"
	}
	neg := false
	if v < 0 {
		neg = true
		v = -v
	}
	buf := make([]byte, 0, 12)
	for v > 0 {
		buf = append(buf, byte('0'+v%10))
		v /= 10
	}
	if neg {
		buf = append(buf, '-')
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}

func cleanPDFText(value string) string {
	replacer := strings.NewReplacer(
		"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ñ", "n",
		"Á", "A", "É", "E", "Í", "I", "Ó", "O", "Ú", "U", "Ñ", "N",
		"ü", "u", "Ü", "U", "·", "-", "–", "-", "—", "-", "“", "\"", "”", "\"",
	)
	value = replacer.Replace(value)
	var b strings.Builder
	for _, r := range value {
		if r == '\n' || r == '\t' || (r >= 32 && r <= 126) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func wrapPDFLine(value string, max int) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return []string{""}
	}
	words := strings.Fields(value)
	lines := []string{}
	current := ""
	for _, word := range words {
		if len(current)+len(word)+1 > max && current != "" {
			lines = append(lines, current)
			current = word
			continue
		}
		if current == "" {
			current = word
		} else {
			current += " " + word
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func escapePDFString(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "(", "\\(")
	value = strings.ReplaceAll(value, ")", "\\)")
	return value
}

func preprocessLabel(key string) string {
	switch key {
	case "grayscale":
		return "Escala de grises"
	case "binary":
		return "Binarización"
	case "edges":
		return "Bordes"
	case "lines":
		return "Líneas y márgenes"
	default:
		return key
	}
}
