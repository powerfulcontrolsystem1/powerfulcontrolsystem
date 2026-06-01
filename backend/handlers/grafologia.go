package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	grafologix "github.com/you/pos-backend/internal/grafologia"
)

type grafologiaCatalogo struct {
	Metricas       []string          `json:"metricas"`
	Interpretacion []string          `json:"interpretacion"`
	Exportaciones  []string          `json:"exportaciones"`
	OCR            map[string]string `json:"ocr"`
	Advertencia    string            `json:"advertencia"`
}

func EmpresaGrafologiaHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := dbpkg.EnsureEmpresaGrafologiaSchema(dbEmp); err != nil {
			log.Printf("[grafologia] ensure schema empresa_id=%d error: %v", empresaID, err)
			http.Error(w, "No se pudo preparar el modulo de grafologia", http.StatusInternalServerError)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch r.Method {
		case http.MethodGet:
			handleEmpresaGrafologiaGET(w, r, dbEmp, empresaID, action)
		case http.MethodPost:
			handleEmpresaGrafologiaPOST(w, r, dbEmp, empresaID, action)
		default:
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
		}
		_ = dbSuper
	}
}

func handleEmpresaGrafologiaGET(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, empresaID int64, action string) {
	switch action {
	case "", "dashboard":
		limit, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("limit")))
		items, err := dbpkg.ListEmpresaGrafologiaAnalisis(dbEmp, empresaID, limit)
		if err != nil {
			log.Printf("[grafologia] list empresa_id=%d error: %v", empresaID, err)
			http.Error(w, "No se pudieron listar los analisis grafológicos", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":         true,
			"empresa_id": empresaID,
			"items":      items,
			"resumen": map[string]interface{}{
				"analisis": len(items),
				"motor":    grafologix.EngineVersion,
			},
		})
	case "catalogo":
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "catalogo": buildGrafologiaCatalogo()})
	case "analisis":
		id, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("id")), 10, 64)
		item, err := dbpkg.GetEmpresaGrafologiaAnalisis(dbEmp, empresaID, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "analisis no encontrado para esta empresa", http.StatusNotFound)
				return
			}
			http.Error(w, "No se pudo cargar el analisis", http.StatusInternalServerError)
			return
		}
		item.Metricas = decodeJSONList(item.MetricasJSON)
		item.Interpretacion = decodeJSONList(item.InterpretacionJSON)
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "item": item})
	case "reporte":
		id, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("id")), 10, 64)
		format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
		item, err := dbpkg.GetEmpresaGrafologiaAnalisis(dbEmp, empresaID, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "analisis no encontrado para esta empresa", http.StatusNotFound)
				return
			}
			http.Error(w, "No se pudo cargar el reporte", http.StatusInternalServerError)
			return
		}
		if format == "json" {
			item.Metricas = decodeJSONList(item.MetricasJSON)
			item.Interpretacion = decodeJSONList(item.InterpretacionJSON)
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "item": item})
			return
		}
		if format == "pdf" {
			pdf := grafologix.RenderPDFReport(item.Titulo, buildGrafologiaResultFromItem(item))
			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("Content-Disposition", "inline; filename=grafologix_reporte.pdf")
			_, _ = w.Write(pdf)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(item.ReporteHTML))
	default:
		http.Error(w, "accion no soportada", http.StatusBadRequest)
	}
}

func handleEmpresaGrafologiaPOST(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, empresaID int64, action string) {
	switch action {
	case "analizar":
		if err := r.ParseMultipartForm(18 << 20); err != nil {
			http.Error(w, "payload multipart invalido", http.StatusBadRequest)
			return
		}
		file, header, err := r.FormFile("imagen")
		if err != nil {
			file, header, err = r.FormFile("file")
		}
		if err != nil {
			http.Error(w, "imagen manuscrita requerida", http.StatusBadRequest)
			return
		}
		defer file.Close()
		data, err := io.ReadAll(io.LimitReader(file, 16<<20))
		if err != nil {
			http.Error(w, "No se pudo leer la imagen", http.StatusBadRequest)
			return
		}
		mimeType := detectGrafologiaMime(data, header.Filename)
		if !strings.HasPrefix(mimeType, "image/") {
			http.Error(w, "solo se permiten imagenes", http.StatusBadRequest)
			return
		}
		imageURL, fileName, absPath, err := saveEmpresaGrafologiaImage(data, header.Filename, empresaID)
		if err != nil {
			log.Printf("[grafologia] save empresa_id=%d error: %v", empresaID, err)
			http.Error(w, "No se pudo guardar la imagen del analisis", http.StatusInternalServerError)
			return
		}
		ocrTexto := strings.TrimSpace(r.FormValue("ocr_texto"))
		ocrMotor := "go_heuristico"
		if ocrTexto == "" {
			if text, ok := runOptionalTesseractOCR(r.Context(), absPath); ok {
				ocrTexto = text
				ocrMotor = "tesseract_cli"
			}
		}
		result, err := grafologix.AnalyzeImageBytes(data, ocrTexto)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		metricsJSON, _ := json.Marshal(result.Metrics)
		traitsJSON, _ := json.Marshal(result.Traits)
		title := strings.TrimSpace(r.FormValue("titulo"))
		if title == "" {
			title = "Informe grafológico GRAFOLOGIX"
		}
		htmlReport := grafologix.RenderHTMLReport(title, result)
		item := dbpkg.EmpresaGrafologiaAnalisis{
			EmpresaID:          empresaID,
			Titulo:             title,
			ArchivoNombre:      fileName,
			ImagenURL:          imageURL,
			ImagenMime:         mimeType,
			OCRTexto:           ocrTexto,
			OCRMotor:           ocrMotor,
			Estado:             "completado",
			Resumen:            result.Summary,
			MetricasJSON:       string(metricsJSON),
			InterpretacionJSON: string(traitsJSON),
			ReporteHTML:        htmlReport,
			ConfianzaGlobal:    result.GlobalTrust,
			UsuarioCreador:     adminEmailFromRequest(r),
		}
		id, err := dbpkg.InsertEmpresaGrafologiaAnalisis(dbEmp, item)
		if err != nil {
			log.Printf("[grafologia] insert empresa_id=%d error: %v", empresaID, err)
			http.Error(w, "No se pudo guardar el analisis", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":          true,
			"id":          id,
			"empresa_id":  empresaID,
			"imagen_url":  imageURL,
			"resultado":   result,
			"reporte_url": fmt.Sprintf("/api/empresa/grafologia?empresa_id=%d&action=reporte&id=%d&format=html", empresaID, id),
		})
	default:
		http.Error(w, "accion no soportada", http.StatusBadRequest)
	}
}

func buildGrafologiaCatalogo() grafologiaCatalogo {
	return grafologiaCatalogo{
		Metricas: []string{
			"Inclinacion de escritura", "Presion del trazo", "Tamano de letra", "Espaciado", "Continuidad",
			"Direccion de lineas", "Margenes", "Velocidad estimada", "Regularidad", "Forma de letras",
		},
		Interpretacion: []string{
			"Organizacion", "Extroversion", "Introversion", "Impulsividad", "Estabilidad emocional",
			"Creatividad", "Disciplina", "Sociabilidad", "Concentracion", "Seguridad personal", "Liderazgo", "Adaptabilidad",
		},
		Exportaciones: []string{"HTML imprimible", "JSON", "PDF desde impresion del navegador"},
		OCR: map[string]string{
			"go":        "Analisis geometrico integrado en Go puro.",
			"tesseract": "OCR libre opcional por CLI cuando GRAFOLOGIA_TESSERACT_ENABLED=1.",
			"opencv":    "Recomendado como sidecar futuro para perspectiva avanzada, Canny y Hough sin acoplar dependencias Go.",
		},
		Advertencia: "La grafologia se maneja como lectura heuristica orientativa, no como diagnostico psicologico ni decision automatizada.",
	}
}

func decodeJSONList(raw string) interface{} {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []interface{}{}
	}
	var out interface{}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return []interface{}{}
	}
	return out
}

func buildGrafologiaResultFromItem(item dbpkg.EmpresaGrafologiaAnalisis) grafologix.AnalysisResult {
	var metrics []grafologix.Metric
	var traits []grafologix.Trait
	_ = json.Unmarshal([]byte(strings.TrimSpace(item.MetricasJSON)), &metrics)
	_ = json.Unmarshal([]byte(strings.TrimSpace(item.InterpretacionJSON)), &traits)
	return grafologix.AnalysisResult{
		Version:     grafologix.EngineVersion,
		GeneratedAt: item.FechaCreacion,
		Summary:     item.Resumen,
		GlobalTrust: item.ConfianzaGlobal,
		Metrics:     metrics,
		Traits:      traits,
		TechnicalNotes: []string{
			"Analisis heuristico orientativo: no es diagnostico psicologico, medico, juridico ni prueba de seleccion de personal.",
			"El PDF resume el informe; la vista HTML conserva el detalle completo y las barras visuales.",
		},
	}
}

func detectGrafologiaMime(data []byte, original string) string {
	mimeType := http.DetectContentType(data)
	if strings.EqualFold(mimeType, "application/octet-stream") {
		if ext := strings.ToLower(filepath.Ext(original)); ext != "" {
			if byExt := mime.TypeByExtension(ext); strings.TrimSpace(byExt) != "" {
				mimeType = byExt
			}
		}
	}
	return strings.ToLower(strings.TrimSpace(mimeType))
}

func saveEmpresaGrafologiaImage(data []byte, originalFilename string, empresaID int64) (string, string, string, error) {
	folder := fmt.Sprintf("empresa_%d", empresaID)
	dir := filepath.Join(resolveWebRootDir(), "uploads", "empresas", folder, "imagenes", "grafologia")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", "", err
	}
	ext := strings.ToLower(filepath.Ext(originalFilename))
	if ext == "" || len(ext) > 8 {
		ext = ".png"
	}
	base := sanitizeGrafologiaFilename(strings.TrimSuffix(filepath.Base(originalFilename), filepath.Ext(originalFilename)))
	if base == "" {
		base = "manuscrito"
	}
	fileName := fmt.Sprintf("%s_%d%s", base, time.Now().UnixNano(), ext)
	absPath := filepath.Join(dir, fileName)
	if err := os.WriteFile(absPath, data, 0o644); err != nil {
		return "", "", "", err
	}
	publicURL := "/uploads/empresas/" + folder + "/imagenes/grafologia/" + fileName
	return publicURL, fileName, absPath, nil
}

func sanitizeGrafologiaFilename(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	prevDash := false
	for _, r := range value {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if ok {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash {
			b.WriteByte('-')
			prevDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func runOptionalTesseractOCR(ctx context.Context, imagePath string) (string, bool) {
	if strings.TrimSpace(os.Getenv("GRAFOLOGIA_TESSERACT_ENABLED")) != "1" {
		return "", false
	}
	bin := strings.TrimSpace(os.Getenv("GRAFOLOGIA_TESSERACT_BIN"))
	if bin == "" {
		bin = "tesseract"
	}
	lang := strings.TrimSpace(os.Getenv("GRAFOLOGIA_TESSERACT_LANG"))
	if lang == "" {
		lang = "spa+eng"
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(timeoutCtx, bin, imagePath, "stdout", "-l", lang, "--psm", "6")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("[grafologia] tesseract opcional no disponible: %v", err)
		return "", false
	}
	text := strings.TrimSpace(string(output))
	return text, text != ""
}
