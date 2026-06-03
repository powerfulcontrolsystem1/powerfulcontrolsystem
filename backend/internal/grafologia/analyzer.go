package grafologia

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"sort"
	"strings"
	"time"
)

const EngineVersion = "grafologix-go-1.0"

type AnalysisResult struct {
	Version        string            `json:"version"`
	GeneratedAt    string            `json:"generated_at"`
	Subject        *SubjectInfo      `json:"subject,omitempty"`
	Summary        string            `json:"summary"`
	GlobalTrust    float64           `json:"global_trust"`
	Image          ImageSummary      `json:"image"`
	Metrics        []Metric          `json:"metrics"`
	Traits         []Trait           `json:"traits"`
	Preprocess     *PreprocessResult `json:"preprocess,omitempty"`
	TechnicalNotes []string          `json:"technical_notes"`
	OCRText        string            `json:"ocr_text,omitempty"`
	Raw            json.RawMessage   `json:"raw,omitempty"`
}

type SubjectInfo struct {
	ClienteID              int64  `json:"cliente_id,omitempty"`
	ClienteNombre          string `json:"cliente_nombre,omitempty"`
	ClienteDocumento       string `json:"cliente_documento,omitempty"`
	PersonaDescripcion     string `json:"persona_descripcion,omitempty"`
	PersonaCaracteristicas string `json:"persona_caracteristicas,omitempty"`
}

type ImageSummary struct {
	Width            int     `json:"width"`
	Height           int     `json:"height"`
	InkDensity       float64 `json:"ink_density"`
	Contrast         float64 `json:"contrast"`
	AverageDarkness  float64 `json:"average_darkness"`
	LinesDetected    int     `json:"lines_detected"`
	WordsEstimated   int     `json:"words_estimated"`
	LettersEstimated int     `json:"letters_estimated"`
}

type Metric struct {
	Key         string         `json:"key"`
	Name        string         `json:"name"`
	Value       string         `json:"value"`
	Category    string         `json:"category"`
	Score       float64        `json:"score"`
	Confidence  float64        `json:"confidence"`
	Explanation string         `json:"explanation"`
	Details     []MetricDetail `json:"details,omitempty"`
}

type MetricDetail struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Unit  string `json:"unit,omitempty"`
	Note  string `json:"note,omitempty"`
}

type Trait struct {
	Key         string  `json:"key"`
	Name        string  `json:"name"`
	Percent     float64 `json:"percent"`
	Level       string  `json:"level"`
	Confidence  float64 `json:"confidence"`
	Explanation string  `json:"explanation"`
}

type lineBand struct {
	Start int
	End   int
}

type inkBox struct {
	MinX int
	MinY int
	MaxX int
	MaxY int
}

type analysisWorkspace struct {
	width      int
	height     int
	gray       []uint8
	threshold  uint8
	ink        []bool
	inkCount   int
	inkDarkSum float64
	box        inkBox
	lines      []lineBand
	rowCounts  []int
	colCounts  []int
	words      int
	letters    int
}

func AnalyzeImageBytes(data []byte, ocrText string) (AnalysisResult, error) {
	if len(data) == 0 {
		return AnalysisResult{}, fmt.Errorf("imagen requerida")
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return AnalysisResult{}, fmt.Errorf("no se pudo leer la imagen manuscrita: %w", err)
	}
	ws := buildWorkspace(img)
	if ws.width == 0 || ws.height == 0 || ws.inkCount == 0 {
		return AnalysisResult{}, fmt.Errorf("la imagen no tiene suficientes trazos detectables")
	}
	metrics := buildMetrics(ws)
	traits := buildTraits(metrics)
	globalTrust := estimateGlobalTrust(ws, metrics)
	result := AnalysisResult{
		Version:     EngineVersion,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Summary:     buildSummary(metrics, traits),
		GlobalTrust: round2(globalTrust),
		Image: ImageSummary{
			Width:            ws.width,
			Height:           ws.height,
			InkDensity:       round4(float64(ws.inkCount) / float64(ws.width*ws.height)),
			Contrast:         round2(imageContrast(ws.gray)),
			AverageDarkness:  round2(ws.inkDarkSum / math.Max(1, float64(ws.inkCount))),
			LinesDetected:    len(ws.lines),
			WordsEstimated:   ws.words,
			LettersEstimated: ws.letters,
		},
		Metrics: metrics,
		Traits:  traits,
		TechnicalNotes: []string{
			"Analisis heuristico orientativo: no es diagnostico psicologico, medico, juridico ni prueba de seleccion de personal.",
			"El motor inicial usa geometria de imagen en Go puro; Tesseract OCR puede complementar el texto cuando este instalado en la VPS.",
			"Las metricas dependen de iluminacion, resolucion, angulo de captura y calidad del manuscrito.",
		},
		OCRText: strings.TrimSpace(ocrText),
	}
	return result, nil
}

func buildWorkspace(img image.Image) analysisWorkspace {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	gray := make([]uint8, 0, w*h)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			luma := uint8((299*uint32(r>>8) + 587*uint32(g>>8) + 114*uint32(b>>8)) / 1000)
			gray = append(gray, luma)
		}
	}
	threshold := otsuThreshold(gray)
	ink := make([]bool, len(gray))
	rowCounts := make([]int, h)
	colCounts := make([]int, w)
	box := inkBox{MinX: w, MinY: h, MaxX: 0, MaxY: 0}
	var inkCount int
	var darkSum float64
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x
			isInk := gray[i] <= threshold
			if isInk {
				ink[i] = true
				inkCount++
				darkSum += (255 - float64(gray[i])) / 255 * 100
				rowCounts[y]++
				colCounts[x]++
				if x < box.MinX {
					box.MinX = x
				}
				if y < box.MinY {
					box.MinY = y
				}
				if x > box.MaxX {
					box.MaxX = x
				}
				if y > box.MaxY {
					box.MaxY = y
				}
			}
		}
	}
	if inkCount == 0 {
		box = inkBox{}
	}
	lines := segmentLines(rowCounts, w)
	words, letters := estimateWordsAndLetters(ink, w, h, lines)
	return analysisWorkspace{
		width:      w,
		height:     h,
		gray:       gray,
		threshold:  threshold,
		ink:        ink,
		inkCount:   inkCount,
		inkDarkSum: darkSum,
		box:        box,
		lines:      lines,
		rowCounts:  rowCounts,
		colCounts:  colCounts,
		words:      words,
		letters:    letters,
	}
}

func buildMetrics(ws analysisWorkspace) []Metric {
	slCat, slScore, slTrust, slExp := metricSlant(ws)
	pressureCat, pressureScore, pressureTrust, pressureExp := metricPressure(ws)
	sizeCat, sizeScore, sizeTrust, sizeExp := metricLetterSize(ws)
	spCat, spScore, spTrust, spExp := metricSpacing(ws)
	contCat, contScore, contTrust, contExp := metricContinuity(ws)
	dirCat, dirScore, dirTrust, dirExp := metricBaseline(ws)
	marginCat, marginScore, marginTrust, marginExp := metricMargins(ws)
	speedCat, speedScore, speedTrust, speedExp := metricSpeed(contScore, pressureScore, spScore)
	regCat, regScore, regTrust, regExp := metricRegularity(ws)
	formCat, formScore, formTrust, formExp := metricLetterForm(ws)
	metrics := []Metric{
		metric("inclinacion", "Inclinacion de escritura", slCat, slCat, slScore, slTrust, slExp),
		metric("presion", "Presion del trazo", pressureCat, pressureCat, pressureScore, pressureTrust, pressureExp),
		metric("tamano_letra", "Tamano de letra", sizeCat, sizeCat, sizeScore, sizeTrust, sizeExp),
		metric("espaciado", "Espaciado", spCat, spCat, spScore, spTrust, spExp),
		metric("continuidad", "Continuidad", contCat, contCat, contScore, contTrust, contExp),
		metric("direccion_lineas", "Direccion de lineas", dirCat, dirCat, dirScore, dirTrust, dirExp),
		metric("margenes", "Margenes", marginCat, marginCat, marginScore, marginTrust, marginExp),
		metric("velocidad", "Velocidad estimada", speedCat, speedCat, speedScore, speedTrust, speedExp),
		metric("regularidad", "Regularidad", regCat, regCat, regScore, regTrust, regExp),
		metric("forma_letras", "Forma de letras", formCat, formCat, formScore, formTrust, formExp),
	}
	for i := range metrics {
		metrics[i].Details = metricDetails(ws, metrics[i].Key)
	}
	return metrics
}

func buildTraits(metrics []Metric) []Trait {
	scores := map[string]float64{}
	for _, m := range metrics {
		scores[m.Key] = m.Score
	}
	regularity := scores["regularidad"]
	pressure := scores["presion"]
	size := scores["tamano_letra"]
	spacing := scores["espaciado"]
	continuity := scores["continuidad"]
	slanta := scores["inclinacion"]
	baseline := scores["direccion_lineas"]
	margins := scores["margenes"]
	form := scores["forma_letras"]
	traits := []Trait{
		trait("organizacion", "Nivel de organizacion", avg(regularity, margins, baseline), "Regularidad, margenes y linea base sostienen esta lectura."),
		trait("extroversion", "Extroversion", avg(slanta, size, spacing), "Mayor inclinacion a la derecha, amplitud y espacios abiertos elevan la puntuacion."),
		trait("introversion", "Introversion", 100-avg(slanta, size, spacing), "La lectura aumenta cuando la escritura es mas vertical o izquierda y compacta."),
		trait("impulsividad", "Impulsividad", avg(100-regularity, continuity, pressure), "Trazos firmes, continuidad alta e irregularidad aumentan la estimacion."),
		trait("estabilidad_emocional", "Estabilidad emocional", avg(regularity, 100-math.Abs(pressure-55), baseline), "Se pondera equilibrio de presion, regularidad y direccion estable."),
		trait("creatividad", "Creatividad", avg(form, 100-math.Abs(spacing-55), 100-math.Abs(size-55)), "Variacion de forma, tamano medio y espaciado no extremo aportan a la lectura."),
		trait("disciplina", "Disciplina", avg(regularity, margins, 100-math.Abs(pressure-55)), "Orden espacial y consistencia del trazo favorecen disciplina operativa."),
		trait("sociabilidad", "Sociabilidad", avg(slanta, spacing, size), "Se estima por apertura espacial, tamano y orientacion del trazo."),
		trait("concentracion", "Concentracion", avg(regularity, 100-math.Abs(size-45), 100-math.Abs(spacing-45)), "Mejor puntuacion cuando hay compactacion controlada y consistencia."),
		trait("seguridad_personal", "Seguridad personal", avg(pressure, baseline, slanta), "Presion firme, direccion controlada y avance del trazo aumentan la lectura."),
		trait("liderazgo", "Liderazgo", avg(size, pressure, baseline, regularity), "Combina presencia grafica, firmeza y estructura."),
		trait("adaptabilidad", "Adaptabilidad", avg(form, 100-math.Abs(continuity-55), regularity), "Se favorece cuando la escritura mezcla formas y mantiene equilibrio."),
	}
	sort.SliceStable(traits, func(i, j int) bool { return traits[i].Percent > traits[j].Percent })
	return traits
}

func metric(key, name, value, category string, score, confidence float64, explanation string) Metric {
	return Metric{Key: key, Name: name, Value: value, Category: category, Score: clamp(round2(score), 0, 100), Confidence: clamp(round2(confidence), 0, 100), Explanation: explanation}
}

func trait(key, name string, percent float64, explanation string) Trait {
	percent = clamp(round2(percent), 0, 100)
	level := "Medio"
	if percent >= 70 {
		level = "Alto"
	} else if percent < 40 {
		level = "Bajo"
	}
	return Trait{Key: key, Name: name, Percent: percent, Level: level, Confidence: round2(62 + math.Min(18, math.Abs(percent-50)/2)), Explanation: explanation}
}

func metricSlant(ws analysisWorkspace) (string, float64, float64, string) {
	if ws.inkCount < 40 {
		return "Vertical", 50, 30, "No hay suficientes trazos para estimar inclinacion con fuerza."
	}
	var n, sumY, sumX, sumYY, sumYX float64
	for y := ws.box.MinY; y <= ws.box.MaxY; y++ {
		var sx, c float64
		for x := ws.box.MinX; x <= ws.box.MaxX; x++ {
			if ws.ink[y*ws.width+x] {
				sx += float64(x)
				c++
			}
		}
		if c > 0 {
			xavg := sx / c
			yy := float64(y - ws.box.MinY)
			n++
			sumY += yy
			sumX += xavg
			sumYY += yy * yy
			sumYX += yy * xavg
		}
	}
	den := n*sumYY - sumY*sumY
	slope := 0.0
	if math.Abs(den) > 0.001 {
		slope = (n*sumYX - sumY*sumX) / den
	}
	if slope > 0.08 {
		return "Derecha", 74 + math.Min(20, slope*50), 72, "Los centroides de tinta avanzan hacia la derecha al descender por el trazo."
	}
	if slope < -0.08 {
		return "Izquierda", 28 + math.Max(-18, slope*45), 72, "Los centroides de tinta retroceden hacia la izquierda al descender por el trazo."
	}
	return "Vertical", 52, 76, "La proyeccion del trazo se mantiene cerca del eje vertical."
}

func metricPressure(ws analysisWorkspace) (string, float64, float64, string) {
	darkness := ws.inkDarkSum / math.Max(1, float64(ws.inkCount))
	density := float64(ws.inkCount) / float64(ws.width*ws.height)
	score := clamp(darkness*0.8+density*500, 0, 100)
	if score >= 68 {
		return "Alta", score, 70, "La tinta detectada es densa y con alto contraste respecto al fondo."
	}
	if score <= 42 {
		return "Baja", score, 68, "La tinta detectada tiene menor oscuridad o baja densidad visual."
	}
	return "Media", score, 76, "El trazo presenta densidad y contraste intermedios."
}

func metricLetterSize(ws analysisWorkspace) (string, float64, float64, string) {
	heights := make([]float64, 0, len(ws.lines))
	for _, line := range ws.lines {
		heights = append(heights, float64(line.End-line.Start+1))
	}
	avgHeight := average(heights)
	score := clamp(avgHeight/float64(ws.height)*420, 0, 100)
	if avgHeight >= 42 {
		return "Grande", score, 74, "La altura promedio de linea manuscrita es amplia para la imagen."
	}
	if avgHeight <= 24 {
		return "Pequena", score, 72, "La altura promedio de linea manuscrita es compacta."
	}
	return "Mediana", score, 78, "La altura promedio de linea manuscrita esta en un rango medio."
}

func metricSpacing(ws analysisWorkspace) (string, float64, float64, string) {
	lineGaps := make([]float64, 0)
	for i := 1; i < len(ws.lines); i++ {
		lineGaps = append(lineGaps, float64(ws.lines[i].Start-ws.lines[i-1].End))
	}
	avgGap := average(lineGaps)
	relative := 0.0
	if len(ws.lines) > 0 {
		relative = avgGap / math.Max(1, float64(ws.box.MaxY-ws.box.MinY+1)/float64(len(ws.lines)))
	}
	score := clamp(45+relative*35+float64(ws.words)/math.Max(1, float64(ws.letters))*120, 0, 100)
	if score >= 68 {
		return "Amplio", score, 68, "Se detectan separaciones marcadas entre palabras o entre lineas."
	}
	if score <= 40 {
		return "Estrecho", score, 66, "Las proyecciones muestran poca separacion entre grupos de tinta."
	}
	return "Medio", score, 74, "El espaciado general se mantiene equilibrado."
}

func metricContinuity(ws analysisWorkspace) (string, float64, float64, string) {
	runs := horizontalRuns(ws)
	if len(runs) == 0 {
		return "Separada", 30, 35, "No se detectaron suficientes recorridos horizontales."
	}
	var longRuns int
	for _, r := range runs {
		if r >= 8 {
			longRuns++
		}
	}
	score := clamp(float64(longRuns)/float64(len(runs))*120, 0, 100)
	if score >= 70 {
		return "Ligada", score, 70, "Hay alta continuidad de trazos horizontales dentro de las lineas."
	}
	if score <= 38 {
		return "Separada", score, 70, "Predominan grupos cortos y separados."
	}
	return "Semi ligada", score, 75, "Combina trazos conectados con pausas visibles."
}

func metricBaseline(ws analysisWorkspace) (string, float64, float64, string) {
	deltas := make([]float64, 0, len(ws.lines))
	for _, line := range ws.lines {
		leftY := centroidYInRange(ws, line, ws.box.MinX, ws.box.MinX+(ws.box.MaxX-ws.box.MinX)/3)
		rightY := centroidYInRange(ws, line, ws.box.MaxX-(ws.box.MaxX-ws.box.MinX)/3, ws.box.MaxX)
		if leftY > 0 && rightY > 0 {
			deltas = append(deltas, rightY-leftY)
		}
	}
	delta := average(deltas)
	if delta < -3 {
		return "Ascendente", 72 + math.Min(18, math.Abs(delta)*2), 68, "El extremo derecho de las lineas tiende a subir respecto al inicio."
	}
	if delta > 3 {
		return "Descendente", 35 - math.Min(20, delta*2), 68, "El extremo derecho de las lineas tiende a bajar respecto al inicio."
	}
	return "Recta", 58, 76, "La linea base permanece estable en la mayoria de renglones."
}

func metricMargins(ws analysisWorkspace) (string, float64, float64, string) {
	left := float64(ws.box.MinX) / math.Max(1, float64(ws.width))
	right := float64(ws.width-1-ws.box.MaxX) / math.Max(1, float64(ws.width))
	top := float64(ws.box.MinY) / math.Max(1, float64(ws.height))
	bottom := float64(ws.height-1-ws.box.MaxY) / math.Max(1, float64(ws.height))
	balance := 100 - (math.Abs(left-right)+math.Abs(top-bottom))*180
	score := clamp(balance, 0, 100)
	if score >= 70 {
		return "Equilibrados", score, 72, "Los margenes detectados guardan proporciones similares."
	}
	if left < 0.04 || right < 0.04 || top < 0.04 || bottom < 0.04 {
		return "Reducidos", score, 70, "Alguno de los margenes queda muy cerca del borde de la imagen."
	}
	return "Irregulares", score, 70, "Los margenes tienen diferencias visibles entre lados."
}

func metricSpeed(continuity, pressure, spacing float64) (string, float64, float64, string) {
	score := clamp(continuity*0.45+(100-math.Abs(pressure-55))*0.25+spacing*0.30, 0, 100)
	if score >= 68 {
		return "Rapida", score, 60, "La continuidad y amplitud del trazo sugieren ejecucion agil."
	}
	if score <= 40 {
		return "Lenta", score, 60, "La separacion y menor continuidad sugieren ejecucion pausada."
	}
	return "Media", score, 66, "Los indicadores de continuidad y espaciado sugieren ritmo intermedio."
}

func metricRegularity(ws analysisWorkspace) (string, float64, float64, string) {
	heights := make([]float64, 0, len(ws.lines))
	gaps := make([]float64, 0)
	for i, line := range ws.lines {
		heights = append(heights, float64(line.End-line.Start+1))
		if i > 0 {
			gaps = append(gaps, float64(line.Start-ws.lines[i-1].End))
		}
	}
	score := 100 - coefficientVariation(heights)*90 - coefficientVariation(gaps)*55
	score = clamp(score, 0, 100)
	if score >= 70 {
		return "Alta", score, 75, "Alturas y espacios entre lineas son consistentes."
	}
	if score <= 42 {
		return "Baja", score, 70, "Se detectan variaciones fuertes entre lineas o espacios."
	}
	return "Media", score, 76, "La escritura mantiene regularidad aceptable con algunas variaciones."
}

func metricLetterForm(ws analysisWorkspace) (string, float64, float64, string) {
	corners := 0
	transitions := 0
	for y := ws.box.MinY + 1; y < ws.box.MaxY; y++ {
		for x := ws.box.MinX + 1; x < ws.box.MaxX; x++ {
			if !ws.ink[y*ws.width+x] {
				continue
			}
			hv := boolToInt(ws.ink[y*ws.width+x-1]) + boolToInt(ws.ink[y*ws.width+x+1]) + boolToInt(ws.ink[(y-1)*ws.width+x]) + boolToInt(ws.ink[(y+1)*ws.width+x])
			diag := boolToInt(ws.ink[(y-1)*ws.width+x-1]) + boolToInt(ws.ink[(y-1)*ws.width+x+1]) + boolToInt(ws.ink[(y+1)*ws.width+x-1]) + boolToInt(ws.ink[(y+1)*ws.width+x+1])
			transitions++
			if diag > hv+1 || hv <= 1 {
				corners++
			}
		}
	}
	score := clamp(float64(corners)/math.Max(1, float64(transitions))*180, 0, 100)
	if score >= 65 {
		return "Angulosas", score, 60, "Predominan cambios bruscos o diagonales en la vecindad de los trazos."
	}
	if score <= 35 {
		return "Redondeadas", score, 60, "Predominan transiciones suaves y pocos vertices detectados."
	}
	return "Mixtas", score, 66, "La forma combina curvas y angulos."
}

func metricDetails(ws analysisWorkspace, key string) []MetricDetail {
	stats := collectWritingStats(ws)
	switch key {
	case "inclinacion":
		slope := slantSlope(ws)
		angle := math.Atan(slope) * 180 / math.Pi
		return []MetricDetail{
			detail("Angulo estimado del eje de tinta", signed2(angle), "grados", "Positivo inclina hacia la derecha; negativo hacia la izquierda."),
			detail("Pendiente horizontal por pixel vertical", sprintf2(slope), "px/px", ""),
			detail("Filas con tinta usadas", intString(stats.InkedRows), "filas", ""),
			detail("Caja analizada", boxText(ws.box), "px", ""),
		}
	case "presion":
		darkness := ws.inkDarkSum / math.Max(1, float64(ws.inkCount))
		density := float64(ws.inkCount) / math.Max(1, float64(ws.width*ws.height))
		return []MetricDetail{
			detail("Oscuridad promedio del trazo", sprintf2(round2(darkness)), "%", ""),
			detail("Densidad de tinta global", sprintf2(round4(density)*100), "%", ""),
			detail("Pixeles de tinta detectados", intString(ws.inkCount), "px", ""),
			detail("Umbral Otsu usado", intString(int(ws.threshold)), "0-255", ""),
		}
	case "tamano_letra":
		return []MetricDetail{
			detail("Altura promedio de renglon", sprintf2(stats.AvgLineHeight), "px", "Se usa como aproximacion del tamano de letra."),
			detail("Altura minima de renglon", sprintf2(stats.MinLineHeight), "px", ""),
			detail("Altura maxima de renglon", sprintf2(stats.MaxLineHeight), "px", ""),
			detail("Lineas detectadas", intString(len(ws.lines)), "lineas", ""),
		}
	case "espaciado":
		return []MetricDetail{
			detail("Separacion promedio entre letras/componentes", sprintf2(stats.AvgLetterGap), "px", "Estimacion por huecos pequenos entre grupos de tinta."),
			detail("Separacion promedio entre palabras", sprintf2(stats.AvgWordGap), "px", "Estimacion por huecos amplios entre grupos de tinta."),
			detail("Separacion promedio entre lineas", sprintf2(stats.AvgLineGap), "px", ""),
			detail("Relacion palabras/letras", sprintf2(float64(ws.words)/math.Max(1, float64(ws.letters))), "ratio", ""),
		}
	case "continuidad":
		return []MetricDetail{
			detail("Recorridos horizontales detectados", intString(stats.HorizontalRunCount), "trazos", ""),
			detail("Longitud promedio de recorrido", sprintf2(stats.AvgHorizontalRun), "px", ""),
			detail("Recorridos largos", intString(stats.LongHorizontalRunCount), "trazos", "Recorridos de 8 px o mas."),
			detail("Indice de continuidad", sprintf2(stats.LongRunRatio*100), "%", ""),
		}
	case "direccion_lineas":
		return []MetricDetail{
			detail("Delta vertical promedio derecha-izquierda", signed2(stats.AvgBaselineDelta), "px", "Negativo sugiere linea ascendente; positivo descendente."),
			detail("Angulo de linea base estimado", signed2(stats.AvgBaselineAngle), "grados", ""),
			detail("Lineas con direccion medible", intString(stats.BaselineSamples), "lineas", ""),
			detail("Ancho de escritura usado", intString(maxInt(0, ws.box.MaxX-ws.box.MinX+1)), "px", ""),
		}
	case "margenes":
		return []MetricDetail{
			detail("Margen izquierdo", marginValue(ws.box.MinX, ws.width), "", ""),
			detail("Margen derecho", marginValue(ws.width-1-ws.box.MaxX, ws.width), "", ""),
			detail("Margen superior", marginValue(ws.box.MinY, ws.height), "", ""),
			detail("Margen inferior", marginValue(ws.height-1-ws.box.MaxY, ws.height), "", ""),
		}
	case "velocidad":
		return []MetricDetail{
			detail("Continuidad usada", sprintf2(stats.LongRunRatio*100), "%", ""),
			detail("Separacion palabra/letra", sprintf2(stats.GapRatio), "ratio", ""),
			detail("Presion aproximada", sprintf2((ws.inkDarkSum/math.Max(1, float64(ws.inkCount)))*0.8+float64(ws.inkCount)/math.Max(1, float64(ws.width*ws.height))*500), "puntos", ""),
			detail("Nota", "Velocidad estimada por geometria", "", "No mide tiempo real de escritura."),
		}
	case "regularidad":
		return []MetricDetail{
			detail("Variacion de altura de lineas", sprintf2(stats.LineHeightCV*100), "%", ""),
			detail("Variacion de espacios entre lineas", sprintf2(stats.LineGapCV*100), "%", ""),
			detail("Altura promedio", sprintf2(stats.AvgLineHeight), "px", ""),
			detail("Espacio promedio entre lineas", sprintf2(stats.AvgLineGap), "px", ""),
		}
	case "forma_letras":
		return []MetricDetail{
			detail("Puntos de forma evaluados", intString(stats.ShapeTransitions), "puntos", ""),
			detail("Vertices/angulos detectados", intString(stats.ShapeCorners), "puntos", ""),
			detail("Indice de angulosidad", sprintf2(stats.ShapeCornerRatio*100), "%", ""),
			detail("Letras estimadas", intString(ws.letters), "letras", ""),
		}
	default:
		return nil
	}
}

type writingStats struct {
	AvgLineHeight          float64
	MinLineHeight          float64
	MaxLineHeight          float64
	AvgLineGap             float64
	AvgLetterGap           float64
	AvgWordGap             float64
	GapRatio               float64
	HorizontalRunCount     int
	AvgHorizontalRun       float64
	LongHorizontalRunCount int
	LongRunRatio           float64
	AvgBaselineDelta       float64
	AvgBaselineAngle       float64
	BaselineSamples        int
	LineHeightCV           float64
	LineGapCV              float64
	ShapeTransitions       int
	ShapeCorners           int
	ShapeCornerRatio       float64
	InkedRows              int
}

func collectWritingStats(ws analysisWorkspace) writingStats {
	heights := make([]float64, 0, len(ws.lines))
	lineGaps := make([]float64, 0)
	letterGaps := make([]float64, 0)
	wordGaps := make([]float64, 0)
	baselineDeltas := make([]float64, 0)
	baselineAngles := make([]float64, 0)
	for i, line := range ws.lines {
		heights = append(heights, float64(line.End-line.Start+1))
		if i > 0 {
			lineGaps = append(lineGaps, float64(line.Start-ws.lines[i-1].End))
		}
		runs, gaps := lineRunsAndGaps(ws, line)
		wordBreak := medianInt(gaps) + maxInt(5, (line.End-line.Start)/2)
		for _, gap := range gaps {
			if gap >= wordBreak {
				wordGaps = append(wordGaps, float64(gap))
			} else if gap > 0 {
				letterGaps = append(letterGaps, float64(gap))
			}
		}
		_ = runs
		leftY := centroidYInRange(ws, line, ws.box.MinX, ws.box.MinX+(ws.box.MaxX-ws.box.MinX)/3)
		rightY := centroidYInRange(ws, line, ws.box.MaxX-(ws.box.MaxX-ws.box.MinX)/3, ws.box.MaxX)
		if leftY > 0 && rightY > 0 {
			delta := rightY - leftY
			width := math.Max(1, float64(ws.box.MaxX-ws.box.MinX+1))
			baselineDeltas = append(baselineDeltas, delta)
			baselineAngles = append(baselineAngles, math.Atan(delta/width)*180/math.Pi)
		}
	}
	runs := horizontalRuns(ws)
	runValues := make([]float64, 0, len(runs))
	longRuns := 0
	for _, run := range runs {
		runValues = append(runValues, float64(run))
		if run >= 8 {
			longRuns++
		}
	}
	corners, transitions := shapeCounts(ws)
	inkedRows := 0
	for _, count := range ws.rowCounts {
		if count > 0 {
			inkedRows++
		}
	}
	return writingStats{
		AvgLineHeight:          average(heights),
		MinLineHeight:          minFloat(heights),
		MaxLineHeight:          maxFloat(heights),
		AvgLineGap:             average(lineGaps),
		AvgLetterGap:           average(letterGaps),
		AvgWordGap:             average(wordGaps),
		GapRatio:               average(wordGaps) / math.Max(1, average(letterGaps)),
		HorizontalRunCount:     len(runs),
		AvgHorizontalRun:       average(runValues),
		LongHorizontalRunCount: longRuns,
		LongRunRatio:           float64(longRuns) / math.Max(1, float64(len(runs))),
		AvgBaselineDelta:       average(baselineDeltas),
		AvgBaselineAngle:       average(baselineAngles),
		BaselineSamples:        len(baselineDeltas),
		LineHeightCV:           coefficientVariation(heights),
		LineGapCV:              coefficientVariation(lineGaps),
		ShapeTransitions:       transitions,
		ShapeCorners:           corners,
		ShapeCornerRatio:       float64(corners) / math.Max(1, float64(transitions)),
		InkedRows:              inkedRows,
	}
}

func lineRunsAndGaps(ws analysisWorkspace, line lineBand) ([][2]int, []int) {
	colActive := make([]bool, ws.width)
	for x := 0; x < ws.width; x++ {
		count := 0
		for y := line.Start; y <= line.End; y++ {
			if ws.ink[y*ws.width+x] {
				count++
			}
		}
		colActive[x] = count >= 1
	}
	runs := activeRuns(colActive)
	return runs, inactiveGapsBetweenRuns(runs)
}

func slantSlope(ws analysisWorkspace) float64 {
	if ws.inkCount < 40 {
		return 0
	}
	var n, sumY, sumX, sumYY, sumYX float64
	for y := ws.box.MinY; y <= ws.box.MaxY; y++ {
		var sx, c float64
		for x := ws.box.MinX; x <= ws.box.MaxX; x++ {
			if ws.ink[y*ws.width+x] {
				sx += float64(x)
				c++
			}
		}
		if c > 0 {
			xavg := sx / c
			yy := float64(y - ws.box.MinY)
			n++
			sumY += yy
			sumX += xavg
			sumYY += yy * yy
			sumYX += yy * xavg
		}
	}
	den := n*sumYY - sumY*sumY
	if math.Abs(den) <= 0.001 {
		return 0
	}
	return (n*sumYX - sumY*sumX) / den
}

func shapeCounts(ws analysisWorkspace) (int, int) {
	corners := 0
	transitions := 0
	for y := ws.box.MinY + 1; y < ws.box.MaxY; y++ {
		for x := ws.box.MinX + 1; x < ws.box.MaxX; x++ {
			if !ws.ink[y*ws.width+x] {
				continue
			}
			hv := boolToInt(ws.ink[y*ws.width+x-1]) + boolToInt(ws.ink[y*ws.width+x+1]) + boolToInt(ws.ink[(y-1)*ws.width+x]) + boolToInt(ws.ink[(y+1)*ws.width+x])
			diag := boolToInt(ws.ink[(y-1)*ws.width+x-1]) + boolToInt(ws.ink[(y-1)*ws.width+x+1]) + boolToInt(ws.ink[(y+1)*ws.width+x-1]) + boolToInt(ws.ink[(y+1)*ws.width+x+1])
			transitions++
			if diag > hv+1 || hv <= 1 {
				corners++
			}
		}
	}
	return corners, transitions
}

func detail(label, value, unit, note string) MetricDetail {
	return MetricDetail{Label: label, Value: value, Unit: unit, Note: note}
}

func marginValue(px, total int) string {
	if px < 0 {
		px = 0
	}
	pct := float64(px) / math.Max(1, float64(total)) * 100
	return intString(px) + " px / " + sprintf2(pct) + "%"
}

func boxText(box inkBox) string {
	return intString(box.MinX) + "," + intString(box.MinY) + " - " + intString(box.MaxX) + "," + intString(box.MaxY)
}

func signed2(v float64) string {
	if v > 0 {
		return "+" + sprintf2(v)
	}
	return sprintf2(v)
}

func segmentLines(rowCounts []int, width int) []lineBand {
	threshold := maxInt(3, int(float64(width)*0.008))
	lines := []lineBand{}
	start := -1
	quiet := 0
	for y, c := range rowCounts {
		active := c >= threshold
		if active && start < 0 {
			start = y
			quiet = 0
		}
		if start >= 0 && !active {
			quiet++
			if quiet >= 3 {
				end := y - quiet
				if end-start >= 4 {
					lines = append(lines, lineBand{Start: start, End: end})
				}
				start = -1
				quiet = 0
			}
		}
		if active {
			quiet = 0
		}
	}
	if start >= 0 && len(rowCounts)-1-start >= 4 {
		lines = append(lines, lineBand{Start: start, End: len(rowCounts) - 1})
	}
	return lines
}

func estimateWordsAndLetters(ink []bool, w, h int, lines []lineBand) (int, int) {
	totalWords := 0
	totalLetters := 0
	for _, line := range lines {
		colActive := make([]bool, w)
		for x := 0; x < w; x++ {
			count := 0
			for y := line.Start; y <= line.End; y++ {
				if ink[y*w+x] {
					count++
				}
			}
			colActive[x] = count >= 1
		}
		runs := activeRuns(colActive)
		gaps := inactiveGapsBetweenRuns(runs)
		wordBreak := medianInt(gaps) + maxInt(5, (line.End-line.Start)/2)
		words := 1
		for _, g := range gaps {
			if g >= wordBreak {
				words++
			}
		}
		if len(runs) == 0 {
			words = 0
		}
		totalWords += words
		totalLetters += maxInt(words, len(runs)/2)
	}
	return totalWords, totalLetters
}

func horizontalRuns(ws analysisWorkspace) []int {
	runs := []int{}
	for _, line := range ws.lines {
		for y := line.Start; y <= line.End; y++ {
			current := 0
			for x := ws.box.MinX; x <= ws.box.MaxX; x++ {
				if ws.ink[y*ws.width+x] {
					current++
				} else if current > 0 {
					runs = append(runs, current)
					current = 0
				}
			}
			if current > 0 {
				runs = append(runs, current)
			}
		}
	}
	return runs
}

func centroidYInRange(ws analysisWorkspace, line lineBand, minX, maxX int) float64 {
	var sum, count float64
	if minX < ws.box.MinX {
		minX = ws.box.MinX
	}
	if maxX > ws.box.MaxX {
		maxX = ws.box.MaxX
	}
	for y := line.Start; y <= line.End; y++ {
		for x := minX; x <= maxX; x++ {
			if ws.ink[y*ws.width+x] {
				sum += float64(y)
				count++
			}
		}
	}
	if count == 0 {
		return 0
	}
	return sum / count
}

func buildSummary(metrics []Metric, traits []Trait) string {
	var slant, pressure, size, regularity string
	for _, m := range metrics {
		switch m.Key {
		case "inclinacion":
			slant = m.Value
		case "presion":
			pressure = m.Value
		case "tamano_letra":
			size = m.Value
		case "regularidad":
			regularity = m.Value
		}
	}
	top := "organizacion"
	if len(traits) > 0 {
		top = traits[0].Name
	}
	return fmt.Sprintf("La muestra presenta inclinacion %s, presion %s, tamano %s y regularidad %s. La dimension heuristica mas marcada es %s.", slant, strings.ToLower(pressure), strings.ToLower(size), strings.ToLower(regularity), top)
}

func estimateGlobalTrust(ws analysisWorkspace, metrics []Metric) float64 {
	conf := 0.0
	for _, m := range metrics {
		conf += m.Confidence
	}
	conf = conf / math.Max(1, float64(len(metrics)))
	if ws.width < 700 || ws.height < 450 {
		conf -= 8
	}
	if len(ws.lines) < 2 {
		conf -= 12
	}
	return clamp(conf, 25, 88)
}

func otsuThreshold(gray []uint8) uint8 {
	var hist [256]int
	for _, v := range gray {
		hist[v]++
	}
	total := len(gray)
	sum := 0.0
	for t := 0; t < 256; t++ {
		sum += float64(t * hist[t])
	}
	var sumB, wB, maxVar float64
	threshold := 160
	for t := 0; t < 256; t++ {
		wB += float64(hist[t])
		if wB == 0 {
			continue
		}
		wF := float64(total) - wB
		if wF == 0 {
			break
		}
		sumB += float64(t * hist[t])
		mB := sumB / wB
		mF := (sum - sumB) / wF
		between := wB * wF * (mB - mF) * (mB - mF)
		if between > maxVar {
			maxVar = between
			threshold = t
		}
	}
	if threshold < 80 {
		threshold = 110
	}
	if threshold > 210 {
		threshold = 210
	}
	return uint8(threshold)
}

func imageContrast(gray []uint8) float64 {
	mean := 0.0
	for _, v := range gray {
		mean += float64(v)
	}
	mean /= math.Max(1, float64(len(gray)))
	var variance float64
	for _, v := range gray {
		d := float64(v) - mean
		variance += d * d
	}
	return clamp(math.Sqrt(variance/math.Max(1, float64(len(gray))))/64*100, 0, 100)
}

func activeRuns(active []bool) [][2]int {
	runs := [][2]int{}
	start := -1
	for i, v := range active {
		if v && start < 0 {
			start = i
		}
		if start >= 0 && (!v || i == len(active)-1) {
			end := i
			if !v {
				end = i - 1
			}
			if end >= start {
				runs = append(runs, [2]int{start, end})
			}
			start = -1
		}
	}
	return runs
}

func inactiveGapsBetweenRuns(runs [][2]int) []int {
	gaps := []int{}
	for i := 1; i < len(runs); i++ {
		gaps = append(gaps, runs[i][0]-runs[i-1][1]-1)
	}
	return gaps
}

func medianInt(values []int) int {
	if len(values) == 0 {
		return 0
	}
	cp := append([]int{}, values...)
	sort.Ints(cp)
	return cp[len(cp)/2]
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func minFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

func maxFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

func coefficientVariation(values []float64) float64 {
	if len(values) < 2 {
		return 0.15
	}
	mean := average(values)
	if mean == 0 {
		return 0
	}
	variance := 0.0
	for _, v := range values {
		d := v - mean
		variance += d * d
	}
	return math.Sqrt(variance/float64(len(values))) / math.Abs(mean)
}

func avg(values ...float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += clamp(v, 0, 100)
	}
	return sum / float64(len(values))
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func round4(v float64) float64 {
	return math.Round(v*10000) / 10000
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
