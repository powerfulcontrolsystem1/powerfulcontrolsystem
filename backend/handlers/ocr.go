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
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	ocrConfigEnabledKey     = "ocr.enabled"
	ocrConfigEngineKey      = "ocr.engine"
	ocrConfigTesseractKey   = "ocr.tesseract_bin"
	ocrConfigPDFToPPMKey    = "ocr.pdftoppm_bin"
	ocrConfigLangKey        = "ocr.lang"
	ocrConfigPSMKey         = "ocr.psm"
	ocrConfigMaxUploadMBKey = "ocr.max_upload_mb"
	ocrConfigMaxPDFPagesKey = "ocr.max_pdf_pages"
)

type ocrRuntimeConfig struct {
	Enabled      bool   `json:"enabled"`
	Engine       string `json:"engine"`
	TesseractBin string `json:"tesseract_bin"`
	PDFToPPMBin  string `json:"pdftoppm_bin"`
	Lang         string `json:"lang"`
	PSM          string `json:"psm"`
	MaxUploadMB  int64  `json:"max_upload_mb"`
	MaxPDFPages  int    `json:"max_pdf_pages"`
}

type ocrFieldSuggestion struct {
	Modulo     string  `json:"modulo"`
	Campo      string  `json:"campo"`
	Etiqueta   string  `json:"etiqueta"`
	Valor      string  `json:"valor"`
	Confianza  float64 `json:"confianza"`
	Fuente     string  `json:"fuente"`
	Aplicacion string  `json:"aplicacion"`
}

func EmpresaOCRHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := dbpkg.EnsureEmpresaOCRSchema(dbEmp); err != nil {
			log.Printf("[ocr] ensure schema empresa_id=%d error: %v", empresaID, err)
			http.Error(w, "No se pudo preparar el modulo OCR", http.StatusInternalServerError)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch r.Method {
		case http.MethodGet:
			handleEmpresaOCRGET(w, r, dbEmp, dbSuper, empresaID, action)
		case http.MethodPost:
			handleEmpresaOCRPOST(w, r, dbEmp, dbSuper, empresaID, action)
		default:
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
		}
	}
}

func handleEmpresaOCRGET(w http.ResponseWriter, r *http.Request, dbEmp, dbSuper *sql.DB, empresaID int64, action string) {
	switch action {
	case "", "dashboard", "list":
		limit, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("limit")))
		items, err := dbpkg.ListEmpresaOCRDocumentos(dbEmp, empresaID, limit)
		if err != nil {
			log.Printf("[ocr] list empresa_id=%d error: %v", empresaID, err)
			http.Error(w, "No se pudo listar el historial OCR", http.StatusInternalServerError)
			return
		}
		cfg := resolveOCRRuntimeConfig(dbSuper)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":         true,
			"empresa_id": empresaID,
			"config":     publicOCRConfig(cfg),
			"items":      items,
			"resumen": map[string]interface{}{
				"documentos": len(items),
				"motor":      cfg.Engine,
				"idioma":     cfg.Lang,
			},
		})
	case "documento":
		id, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("id")), 10, 64)
		item, err := dbpkg.GetEmpresaOCRDocumento(dbEmp, empresaID, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "documento OCR no encontrado para esta empresa", http.StatusNotFound)
				return
			}
			http.Error(w, "No se pudo cargar el documento OCR", http.StatusInternalServerError)
			return
		}
		item.Campos = decodeJSONList(item.CamposJSON)
		item.Sugerencias = decodeJSONList(item.SugerenciasJSON)
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "item": item})
	case "catalogo":
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok": true,
			"catalogo": map[string]interface{}{
				"motor_recomendado": "Tesseract OCR + pdftoppm",
				"sin_ia":            true,
				"tipos_documento": []map[string]string{
					{"id": "dian", "nombre": "DIAN / facturacion electronica"},
					{"id": "inventario", "nombre": "Inventario y productos"},
					{"id": "usuarios", "nombre": "Usuarios y empleados"},
					{"id": "clientes", "nombre": "Clientes y terceros"},
					{"id": "general", "nombre": "Documento general"},
				},
			},
		})
	default:
		http.Error(w, "accion no soportada", http.StatusBadRequest)
	}
}

func handleEmpresaOCRPOST(w http.ResponseWriter, r *http.Request, dbEmp, dbSuper *sql.DB, empresaID int64, action string) {
	switch action {
	case "", "procesar", "analizar":
		cfg := resolveOCRRuntimeConfig(dbSuper)
		if !cfg.Enabled {
			http.Error(w, "OCR deshabilitado desde super administrador", http.StatusForbidden)
			return
		}
		maxBytes := cfg.MaxUploadMB << 20
		if maxBytes <= 0 {
			maxBytes = 20 << 20
		}
		if err := r.ParseMultipartForm(maxBytes); err != nil {
			http.Error(w, "payload multipart invalido", http.StatusBadRequest)
			return
		}
		file, header, err := r.FormFile("documento")
		if err != nil {
			file, header, err = r.FormFile("file")
		}
		if err != nil {
			http.Error(w, "documento requerido", http.StatusBadRequest)
			return
		}
		defer file.Close()
		data, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
		if err != nil || int64(len(data)) > maxBytes {
			http.Error(w, "documento supera el tamano permitido", http.StatusBadRequest)
			return
		}
		mimeType := detectOCRMime(data, header.Filename)
		if !isOCRAllowedMime(mimeType) {
			http.Error(w, "tipo de archivo no soportado para OCR", http.StatusBadRequest)
			return
		}
		tipoDocumento := normalizeOCRType(r.FormValue("tipo_documento"))
		titulo := limitOCRText(r.FormValue("titulo"), 180)
		if titulo == "" {
			titulo = "Documento OCR " + strings.ToUpper(tipoDocumento)
		}
		fileURL, fileName, absPath, err := saveEmpresaOCRFile(data, header.Filename, empresaID)
		if err != nil {
			log.Printf("[ocr] save empresa_id=%d error: %v", empresaID, err)
			http.Error(w, "No se pudo guardar el documento OCR", http.StatusInternalServerError)
			return
		}
		text, engine, runErr := runOCRForFile(r.Context(), cfg, absPath, mimeType)
		if runErr != nil {
			log.Printf("[ocr] run empresa_id=%d file=%s error: %v", empresaID, fileName, runErr)
			http.Error(w, "No se pudo ejecutar OCR: "+sanitizePublicError(runErr.Error()), http.StatusBadGateway)
			return
		}
		fields := extractOCRFields(tipoDocumento, text)
		fieldsJSON, _ := json.Marshal(fields)
		suggestionsJSON, _ := json.Marshal(groupOCRSuggestions(fields))
		confidence := estimateOCRConfidence(text, fields)
		item := dbpkg.EmpresaOCRDocumento{
			EmpresaID:       empresaID,
			TipoDocumento:   tipoDocumento,
			Titulo:          titulo,
			ArchivoNombre:   fileName,
			ArchivoURL:      fileURL,
			ArchivoMime:     mimeType,
			OCRMotor:        engine,
			Idioma:          cfg.Lang,
			Estado:          "procesado",
			TextoExtraido:   limitOCRText(text, 60000),
			CamposJSON:      string(fieldsJSON),
			SugerenciasJSON: string(suggestionsJSON),
			Confianza:       confidence,
			UsuarioCreador:  adminEmailFromRequest(r),
		}
		id, err := dbpkg.InsertEmpresaOCRDocumento(dbEmp, item)
		if err != nil {
			log.Printf("[ocr] insert empresa_id=%d error: %v", empresaID, err)
			http.Error(w, "No se pudo guardar el resultado OCR", http.StatusInternalServerError)
			return
		}
		item.ID = id
		item.Campos = fields
		item.Sugerencias = groupOCRSuggestions(fields)
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "item": item})
	default:
		http.Error(w, "accion no soportada", http.StatusBadRequest)
	}
}

func OCRConfigHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			cfg := resolveOCRRuntimeConfig(dbSuper)
			status := testOCRRuntime(r.Context(), cfg)
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "config": publicOCRConfig(cfg), "status": status})
		case http.MethodPost, http.MethodPut:
			if strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("action")), "test") {
				cfg := resolveOCRRuntimeConfig(dbSuper)
				status := testOCRRuntime(r.Context(), cfg)
				code := http.StatusOK
				if !status["tesseract_ok"].(bool) {
					code = http.StatusBadGateway
				}
				writeJSON(w, code, map[string]interface{}{"ok": code == http.StatusOK, "status": status})
				return
			}
			var payload struct {
				Enabled      *bool  `json:"enabled"`
				Engine       string `json:"engine"`
				TesseractBin string `json:"tesseract_bin"`
				PDFToPPMBin  string `json:"pdftoppm_bin"`
				Lang         string `json:"lang"`
				PSM          string `json:"psm"`
				MaxUploadMB  int64  `json:"max_upload_mb"`
				MaxPDFPages  int    `json:"max_pdf_pages"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload invalido: "+err.Error(), http.StatusBadRequest)
				return
			}
			if payload.Enabled != nil {
				v := "0"
				if *payload.Enabled {
					v = "1"
				}
				if err := dbpkg.SetConfigValue(dbSuper, ocrConfigEnabledKey, v, false); err != nil {
					http.Error(w, "No se pudo guardar ocr.enabled", http.StatusInternalServerError)
					return
				}
			}
			pairs := map[string]string{
				ocrConfigEngineKey:    normalizeOCREngine(payload.Engine),
				ocrConfigTesseractKey: strings.TrimSpace(payload.TesseractBin),
				ocrConfigPDFToPPMKey:  strings.TrimSpace(payload.PDFToPPMBin),
				ocrConfigLangKey:      normalizeOCRLang(payload.Lang),
				ocrConfigPSMKey:       normalizeOCRPSM(payload.PSM),
			}
			for key, value := range pairs {
				if value == "" {
					continue
				}
				if err := dbpkg.SetConfigValue(dbSuper, key, value, false); err != nil {
					http.Error(w, "No se pudo guardar "+key, http.StatusInternalServerError)
					return
				}
			}
			if payload.MaxUploadMB > 0 {
				if payload.MaxUploadMB > 80 {
					payload.MaxUploadMB = 80
				}
				_ = dbpkg.SetConfigValue(dbSuper, ocrConfigMaxUploadMBKey, strconv.FormatInt(payload.MaxUploadMB, 10), false)
			}
			if payload.MaxPDFPages > 0 {
				if payload.MaxPDFPages > 12 {
					payload.MaxPDFPages = 12
				}
				_ = dbpkg.SetConfigValue(dbSuper, ocrConfigMaxPDFPagesKey, strconv.Itoa(payload.MaxPDFPages), false)
			}
			cfg := resolveOCRRuntimeConfig(dbSuper)
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "config": publicOCRConfig(cfg)})
		default:
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
		}
	}
}

func resolveOCRRuntimeConfig(dbSuper *sql.DB) ocrRuntimeConfig {
	cfg := ocrRuntimeConfig{Enabled: true, Engine: "tesseract", TesseractBin: "tesseract", PDFToPPMBin: "pdftoppm", Lang: "spa+eng", PSM: "6", MaxUploadMB: 20, MaxPDFPages: 4}
	if strings.TrimSpace(os.Getenv("GRAFOLOGIA_TESSERACT_LANG")) != "" {
		cfg.Lang = strings.TrimSpace(os.Getenv("GRAFOLOGIA_TESSERACT_LANG"))
	}
	if strings.TrimSpace(os.Getenv("GRAFOLOGIA_TESSERACT_BIN")) != "" {
		cfg.TesseractBin = strings.TrimSpace(os.Getenv("GRAFOLOGIA_TESSERACT_BIN"))
	}
	if dbSuper != nil {
		if v := readOCRConfig(dbSuper, ocrConfigEnabledKey); v != "" {
			cfg.Enabled = parseOCRBool(v, cfg.Enabled)
		}
		if v := readOCRConfig(dbSuper, ocrConfigEngineKey); v != "" {
			cfg.Engine = normalizeOCREngine(v)
		}
		if v := readOCRConfig(dbSuper, ocrConfigTesseractKey); v != "" {
			cfg.TesseractBin = v
		}
		if v := readOCRConfig(dbSuper, ocrConfigPDFToPPMKey); v != "" {
			cfg.PDFToPPMBin = v
		}
		if v := readOCRConfig(dbSuper, ocrConfigLangKey); v != "" {
			cfg.Lang = normalizeOCRLang(v)
		}
		if v := readOCRConfig(dbSuper, ocrConfigPSMKey); v != "" {
			cfg.PSM = normalizeOCRPSM(v)
		}
		if v := readOCRConfig(dbSuper, ocrConfigMaxUploadMBKey); v != "" {
			if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 && n <= 80 {
				cfg.MaxUploadMB = n
			}
		}
		if v := readOCRConfig(dbSuper, ocrConfigMaxPDFPagesKey); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 12 {
				cfg.MaxPDFPages = n
			}
		}
	}
	if strings.TrimSpace(os.Getenv("OCR_PDFTOPPM_BIN")) != "" {
		cfg.PDFToPPMBin = strings.TrimSpace(os.Getenv("OCR_PDFTOPPM_BIN"))
	}
	return cfg
}

func readOCRConfig(dbSuper *sql.DB, key string) string {
	value, err := getDecryptedConfigValue(dbSuper, key)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(value)
}

func publicOCRConfig(cfg ocrRuntimeConfig) map[string]interface{} {
	return map[string]interface{}{
		"enabled":       cfg.Enabled,
		"engine":        cfg.Engine,
		"tesseract_bin": cfg.TesseractBin,
		"pdftoppm_bin":  cfg.PDFToPPMBin,
		"lang":          cfg.Lang,
		"psm":           cfg.PSM,
		"max_upload_mb": cfg.MaxUploadMB,
		"max_pdf_pages": cfg.MaxPDFPages,
		"sin_ia":        true,
	}
}

func runOCRForFile(ctx context.Context, cfg ocrRuntimeConfig, absPath, mimeType string) (string, string, error) {
	if strings.EqualFold(mimeType, "application/pdf") {
		return runOCRForPDF(ctx, cfg, absPath)
	}
	return runTesseractImage(ctx, cfg, absPath)
}

func runOCRForPDF(ctx context.Context, cfg ocrRuntimeConfig, absPath string) (string, string, error) {
	tmpDir, err := os.MkdirTemp("", "pcs-ocr-pdf-*")
	if err != nil {
		return "", "", err
	}
	defer os.RemoveAll(tmpDir)
	prefix := filepath.Join(tmpDir, "page")
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(25+cfg.MaxPDFPages*12)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(timeoutCtx, cfg.PDFToPPMBin, "-r", "180", "-png", "-f", "1", "-l", strconv.Itoa(cfg.MaxPDFPages), absPath, prefix)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", "tesseract_cli+pdftoppm", fmt.Errorf("pdftoppm: %s", strings.TrimSpace(string(output)))
	}
	pages, _ := filepath.Glob(prefix + "-*.png")
	sort.Strings(pages)
	if len(pages) == 0 {
		return "", "tesseract_cli+pdftoppm", fmt.Errorf("pdftoppm no genero paginas legibles")
	}
	var parts []string
	for _, page := range pages {
		text, _, err := runTesseractImage(ctx, cfg, page)
		if err != nil {
			return "", "tesseract_cli+pdftoppm", err
		}
		if strings.TrimSpace(text) != "" {
			parts = append(parts, strings.TrimSpace(text))
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n--- pagina OCR ---\n\n")), "tesseract_cli+pdftoppm", nil
}

func runTesseractImage(ctx context.Context, cfg ocrRuntimeConfig, imagePath string) (string, string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 35*time.Second)
	defer cancel()
	cmd := exec.CommandContext(timeoutCtx, cfg.TesseractBin, imagePath, "stdout", "-l", cfg.Lang, "--psm", cfg.PSM)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", "tesseract_cli", fmt.Errorf("tesseract: %s", strings.TrimSpace(string(output)))
	}
	text := strings.TrimSpace(string(output))
	if text == "" {
		return "", "tesseract_cli", fmt.Errorf("OCR no encontro texto util")
	}
	return text, "tesseract_cli", nil
}

func testOCRRuntime(ctx context.Context, cfg ocrRuntimeConfig) map[string]interface{} {
	out := map[string]interface{}{"enabled": cfg.Enabled, "tesseract_ok": false, "pdftoppm_ok": false, "lang": cfg.Lang}
	tctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	if data, err := exec.CommandContext(tctx, cfg.TesseractBin, "--version").CombinedOutput(); err == nil {
		out["tesseract_ok"] = true
		out["tesseract_version"] = firstOCRLine(string(data))
	} else {
		out["tesseract_error"] = sanitizePublicError(err.Error())
	}
	pctx, pcancel := context.WithTimeout(ctx, 8*time.Second)
	defer pcancel()
	if data, err := exec.CommandContext(pctx, cfg.PDFToPPMBin, "-v").CombinedOutput(); err == nil || len(data) > 0 {
		out["pdftoppm_ok"] = true
		out["pdftoppm_version"] = firstOCRLine(string(data))
	} else {
		out["pdftoppm_error"] = sanitizePublicError(err.Error())
	}
	return out
}

func extractOCRFields(tipoDocumento, text string) []ocrFieldSuggestion {
	clean := normalizeOCRText(text)
	fields := []ocrFieldSuggestion{}
	add := func(mod, campo, etiqueta, valor, fuente, aplicacion string, conf float64) {
		valor = strings.TrimSpace(valor)
		if valor == "" || len(valor) > 500 {
			return
		}
		for _, f := range fields {
			if f.Modulo == mod && f.Campo == campo && strings.EqualFold(f.Valor, valor) {
				return
			}
		}
		fields = append(fields, ocrFieldSuggestion{Modulo: mod, Campo: campo, Etiqueta: etiqueta, Valor: valor, Fuente: fuente, Aplicacion: aplicacion, Confianza: conf})
	}
	uuidRe := regexp.MustCompile(`(?i)\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b`)
	for _, v := range uuidRe.FindAllString(clean, 5) {
		campo := "identificador_uuid"
		label := "Identificador UUID"
		if strings.Contains(strings.ToLower(clean), "testsetid") {
			campo = "test_set_id"
			label = "TestSetId DIAN"
		}
		add("facturacion_electronica", campo, label, v, "uuid detectado", "Revisar en Facturacion electronica > Configuracion DIAN Colombia", 0.82)
	}
	patterns := []struct{ campo, label, re string }{
		{"pin_software", "PIN software", `(?i)\bpin\b[^A-Za-z0-9]{0,20}([A-Za-z0-9]{3,30})`},
		{"clave_tecnica", "Clave tecnica DIAN", `(?i)clave\s+t[eé]cnica[^A-Za-z0-9]{0,40}([a-fA-F0-9]{20,120})`},
		{"prefijo", "Prefijo de numeracion", `(?i)\bprefijo\b[^A-Za-z0-9]{0,20}([A-Z0-9]{1,8})`},
		{"numero_resolucion", "Numero de resolucion", `(?i)(?:n[uú]mero\s+)?resoluci[oó]n[^A-Za-z0-9]{0,25}([0-9-]{4,40})`},
		{"rango_desde", "Rango desde", `(?i)rango\s+desde[^A-Za-z0-9]{0,20}([0-9]{3,20})`},
		{"rango_hasta", "Rango hasta", `(?i)rango\s+hasta[^A-Za-z0-9]{0,20}([0-9]{3,20})`},
	}
	for _, p := range patterns {
		if m := regexp.MustCompile(p.re).FindStringSubmatch(clean); len(m) > 1 {
			add("facturacion_electronica", p.campo, p.label, m[len(m)-1], p.label, "Revisar antes de guardar en DIAN Colombia", 0.78)
		}
	}
	for _, url := range regexp.MustCompile(`https?://[^\s<>"']+`).FindAllString(clean, 6) {
		add("facturacion_electronica", "url_servicio", "URL servicio detectada", strings.TrimRight(url, ".,;"), "url", "Revisar ambiente habilitacion/produccion", 0.7)
	}
	for _, email := range regexp.MustCompile(`(?i)\b[A-Z0-9._%+\-]+@[A-Z0-9.\-]+\.[A-Z]{2,}\b`).FindAllString(clean, 20) {
		add("usuarios", "email", "Correo detectado", email, "correo", "Puede usarse para usuario, cliente o proveedor", 0.74)
	}
	for _, phone := range regexp.MustCompile(`(?:\+?57\s*)?(?:3[0-9]{9}|[0-9]{7,10})`).FindAllString(clean, 20) {
		add("clientes", "telefono", "Telefono detectado", phone, "telefono", "Revisar antes de crear cliente o usuario", 0.58)
	}
	if tipoDocumento == "inventario" || strings.Contains(strings.ToLower(clean), "producto") || strings.Contains(strings.ToLower(clean), "item") {
		if m := regexp.MustCompile(`(?i)(?:producto|item|art[ií]culo)\D{0,40}([A-Za-z0-9 ÁÉÍÓÚáéíóúñÑ._-]{3,80})`).FindStringSubmatch(clean); len(m) > 1 {
			add("inventario", "nombre_producto", "Nombre de producto", strings.TrimSpace(m[1]), "producto", "Revisar en Inventario > Productos", 0.62)
		}
		if m := regexp.MustCompile(`(?i)(?:precio|valor|total)\D{0,20}\$?\s*([0-9][0-9.,]{1,18})`).FindStringSubmatch(clean); len(m) > 1 {
			add("inventario", "precio", "Precio o valor", m[1], "valor", "Revisar impuesto y moneda antes de crear producto", 0.58)
		}
	}
	return fields
}

func groupOCRSuggestions(fields []ocrFieldSuggestion) map[string][]ocrFieldSuggestion {
	grouped := map[string][]ocrFieldSuggestion{}
	for _, f := range fields {
		grouped[f.Modulo] = append(grouped[f.Modulo], f)
	}
	return grouped
}

func estimateOCRConfidence(text string, fields []ocrFieldSuggestion) float64 {
	words := len(strings.Fields(text))
	score := 0.35
	if words > 8 {
		score += 0.2
	}
	if words > 40 {
		score += 0.2
	}
	if len(fields) > 0 {
		score += 0.15
	}
	if score > 0.92 {
		score = 0.92
	}
	return score
}

func saveEmpresaOCRFile(data []byte, originalFilename string, empresaID int64) (string, string, string, error) {
	folder := fmt.Sprintf("empresa_%d", empresaID)
	dir := filepath.Join(resolveWebRootDir(), "uploads", "empresas", folder, "ocr")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", "", err
	}
	ext := strings.ToLower(filepath.Ext(originalFilename))
	if ext == "" || len(ext) > 12 {
		ext = ".bin"
	}
	base := sanitizeOCRFilename(strings.TrimSuffix(filepath.Base(originalFilename), filepath.Ext(originalFilename)))
	if base == "" {
		base = "documento"
	}
	fileName := fmt.Sprintf("%s_%d%s", base, time.Now().UnixNano(), ext)
	absPath := filepath.Join(dir, fileName)
	if err := os.WriteFile(absPath, data, 0o640); err != nil {
		return "", "", "", err
	}
	publicURL := "/uploads/empresas/" + folder + "/ocr/" + fileName
	return publicURL, fileName, absPath, nil
}

func detectOCRMime(data []byte, original string) string {
	ext := strings.ToLower(filepath.Ext(original))
	if ext != "" {
		if mt := mime.TypeByExtension(ext); mt != "" {
			return strings.ToLower(strings.Split(mt, ";")[0])
		}
	}
	return strings.ToLower(strings.Split(http.DetectContentType(data), ";")[0])
}

func isOCRAllowedMime(mimeType string) bool {
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	return strings.HasPrefix(mimeType, "image/") || mimeType == "application/pdf"
}

func sanitizeOCRFilename(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			continue
		}
		if r == '-' || r == '_' {
			b.WriteRune(r)
			continue
		}
		if r == ' ' || r == '.' {
			b.WriteRune('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

func normalizeOCRType(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "dian", "facturacion", "facturacion_electronica", "inventario", "usuarios", "clientes", "general":
		return v
	default:
		return "general"
	}
}

func normalizeOCREngine(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" || v == "tesseract_cli" {
		return "tesseract"
	}
	if v != "tesseract" {
		return "tesseract"
	}
	return v
}

func normalizeOCRLang(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "spa+eng"
	}
	return regexp.MustCompile(`[^A-Za-z0-9_+.-]`).ReplaceAllString(v, "")
}

func normalizeOCRPSM(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "6"
	}
	if _, err := strconv.Atoi(v); err != nil {
		return "6"
	}
	return v
}

func parseOCRBool(v string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "si", "sí", "yes", "on", "activo", "enabled":
		return true
	case "0", "false", "no", "off", "inactivo", "disabled":
		return false
	default:
		return fallback
	}
}

func normalizeOCRText(text string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
}

func limitOCRText(value string, max int) string {
	value = strings.TrimSpace(value)
	if max > 0 && len([]rune(value)) > max {
		return string([]rune(value)[:max])
	}
	return value
}

func firstOCRLine(value string) string {
	for _, line := range strings.Split(value, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

func sanitizePublicError(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\n", " ")
	if len(value) > 240 {
		value = value[:240]
	}
	return value
}
