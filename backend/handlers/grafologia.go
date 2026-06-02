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
			handleEmpresaGrafologiaPOST(w, r, dbEmp, dbSuper, empresaID, action)
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
		item.Preprocesamiento = decodeJSONList(item.PreprocesamientoJSON)
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
			item.Preprocesamiento = decodeJSONList(item.PreprocesamientoJSON)
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "item": item})
			return
		}
		result := buildGrafologiaResultFromItem(item)
		if format == "pdf" {
			pdf := grafologix.RenderPDFReport(item.Titulo, result)
			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("Content-Disposition", "inline; filename=grafologix_reporte.pdf")
			_, _ = w.Write(pdf)
			return
		}
		if format == "doc" || format == "word" {
			w.Header().Set("Content-Type", "application/msword; charset=utf-8")
			w.Header().Set("Content-Disposition", "attachment; filename=grafologix_reporte.doc")
			_, _ = w.Write(grafologix.RenderWordReport(item.Titulo, result))
			return
		}
		if format == "csv" {
			w.Header().Set("Content-Type", "text/csv; charset=utf-8")
			w.Header().Set("Content-Disposition", "attachment; filename=grafologix_resultados.csv")
			_, _ = w.Write([]byte(grafologix.RenderCSVReport(result)))
			return
		}
		if format == "txt" {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Header().Set("Content-Disposition", "attachment; filename=grafologix_reporte.txt")
			_, _ = w.Write([]byte(grafologix.RenderTextReport(item.Titulo, result)))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(item.ReporteHTML))
	default:
		http.Error(w, "accion no soportada", http.StatusBadRequest)
	}
}

func handleEmpresaGrafologiaPOST(w http.ResponseWriter, r *http.Request, dbEmp, dbSuper *sql.DB, empresaID int64, action string) {
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
		clienteID, err := parseGrafologiaInt64Form(r, "cliente_id")
		if err != nil {
			http.Error(w, "cliente_id invalido", http.StatusBadRequest)
			return
		}
		personaDescripcion := limitGrafologiaText(r.FormValue("persona_descripcion"), 1200)
		personaCaracteristicas := limitGrafologiaText(r.FormValue("persona_caracteristicas"), 1200)
		personaNombre := limitGrafologiaText(r.FormValue("persona_nombre"), 180)
		clienteNombre := personaNombre
		clienteDocumento := limitGrafologiaText(r.FormValue("cliente_documento"), 80)
		var cliente *dbpkg.Cliente
		if clienteID > 0 {
			cliente, err = dbpkg.GetClienteByID(dbEmp, empresaID, clienteID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "cliente no encontrado para esta empresa", http.StatusBadRequest)
					return
				}
				log.Printf("[grafologia] cliente empresa_id=%d cliente_id=%d error: %v", empresaID, clienteID, err)
				http.Error(w, "No se pudo validar el cliente asociado", http.StatusInternalServerError)
				return
			}
			clienteNombre = strings.TrimSpace(cliente.NombreRazonSocial)
			clienteDocumento = strings.TrimSpace(strings.TrimSpace(cliente.TipoDocumento) + " " + strings.TrimSpace(cliente.NumeroDocumento))
		}
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
		result.Subject = &grafologix.SubjectInfo{
			ClienteID:              clienteID,
			ClienteNombre:          clienteNombre,
			ClienteDocumento:       clienteDocumento,
			PersonaDescripcion:     personaDescripcion,
			PersonaCaracteristicas: personaCaracteristicas,
		}
		preprocess, err := grafologix.GeneratePreprocessArtifacts(data)
		if err != nil {
			log.Printf("[grafologia] preprocess empresa_id=%d error: %v", empresaID, err)
		} else {
			preprocess.ImageURLs = saveEmpresaGrafologiaArtifacts(preprocess.ImageBytes, empresaID, fileName)
			preprocess.ImageBytes = nil
			result.Preprocess = &preprocess
		}
		metricsJSON, _ := json.Marshal(result.Metrics)
		traitsJSON, _ := json.Marshal(result.Traits)
		preprocessJSON, _ := json.Marshal(result.Preprocess)
		title := strings.TrimSpace(r.FormValue("titulo"))
		if title == "" {
			title = "Informe grafológico GRAFOLOGIX"
		}
		htmlReport := grafologix.RenderHTMLReport(title, result)
		item := dbpkg.EmpresaGrafologiaAnalisis{
			EmpresaID:              empresaID,
			ClienteID:              clienteID,
			ClienteNombre:          clienteNombre,
			ClienteDocumento:       clienteDocumento,
			PersonaDescripcion:     personaDescripcion,
			PersonaCaracteristicas: personaCaracteristicas,
			Titulo:                 title,
			ArchivoNombre:          fileName,
			ImagenURL:              imageURL,
			ImagenMime:             mimeType,
			OCRTexto:               ocrTexto,
			OCRMotor:               ocrMotor,
			Estado:                 "completado",
			Resumen:                result.Summary,
			MetricasJSON:           string(metricsJSON),
			InterpretacionJSON:     string(traitsJSON),
			PreprocesamientoJSON:   string(preprocessJSON),
			ReporteHTML:            htmlReport,
			ConfianzaGlobal:        result.GlobalTrust,
			UsuarioCreador:         adminEmailFromRequest(r),
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
			"cliente":     cliente,
			"resultado":   result,
			"reporte_url": fmt.Sprintf("/api/empresa/grafologia?empresa_id=%d&action=reporte&id=%d&format=html", empresaID, id),
		})
	case "analizar_ia", "analizar_gpt55":
		handleEmpresaGrafologiaAnalizarIA(w, r, dbEmp, dbSuper, empresaID)
	default:
		http.Error(w, "accion no soportada", http.StatusBadRequest)
	}
}

func handleEmpresaGrafologiaAnalizarIA(w http.ResponseWriter, r *http.Request, dbEmp, dbSuper *sql.DB, empresaID int64) {
	if !isSuperAIEnabled(dbSuper) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":    false,
			"code":  "ai_disabled",
			"error": "La IA esta desactivada desde super administrador.",
		})
		return
	}
	empresaChatEnabled, _, _, err := getChatIAEmpresaEnabled(dbSuper)
	if err != nil {
		http.Error(w, "No se pudo consultar configuracion de chat IA", http.StatusInternalServerError)
		return
	}
	if !empresaChatEnabled {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":    false,
			"code":  "ai_empresa_chat_disabled",
			"error": "El chat con IA para empresas esta desactivado desde super administrador.",
		})
		return
	}
	maxGPT55, _, _, err := getChatIAEmpresaMaxGPT55ConsultasDia(dbSuper)
	if err != nil {
		http.Error(w, "No se pudo consultar limite GPT-5.5", http.StatusInternalServerError)
		return
	}
	if maxGPT55 == 0 {
		writeJSON(w, http.StatusTooManyRequests, map[string]interface{}{
			"ok":    false,
			"code":  "ai_empresa_gpt55_blocked",
			"error": "Las consultas con GPT-5.5 estan bloqueadas para empresas.",
		})
		return
	}
	model, ok := availableEmpresaAIModelMap(dbSuper)["openai:gpt-5.5"]
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":    false,
			"code":  "ai_gpt55_unavailable",
			"error": "GPT-5.5 no esta disponible o el proveedor OpenAI esta deshabilitado.",
		})
		return
	}
	fechaUso := time.Now().Format("2006-01-02")
	usoActual, err := dbpkg.GetEmpresaAIUsoDiario(dbEmp, empresaID, model.Provider, model.ID, fechaUso)
	if err != nil {
		http.Error(w, "No se pudo consultar uso diario GPT-5.5", http.StatusInternalServerError)
		return
	}
	if usoActual.Consultas >= maxGPT55 {
		writeJSON(w, http.StatusTooManyRequests, map[string]interface{}{
			"ok":    false,
			"code":  "ai_empresa_gpt55_limit_reached",
			"error": "Se alcanzo el limite diario de consultas GPT-5.5 para esta empresa.",
			"usage": map[string]interface{}{
				"daily_used":      usoActual.Consultas,
				"daily_limit":     maxGPT55,
				"daily_remaining": 0,
			},
		})
		return
	}
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
		http.Error(w, "solo se permiten imagenes para analisis GPT-5.5", http.StatusBadRequest)
		return
	}

	clienteID, err := parseGrafologiaInt64Form(r, "cliente_id")
	if err != nil {
		http.Error(w, "cliente_id invalido", http.StatusBadRequest)
		return
	}
	personaDescripcion := limitGrafologiaText(r.FormValue("persona_descripcion"), 1200)
	personaCaracteristicas := limitGrafologiaText(r.FormValue("persona_caracteristicas"), 1200)
	personaNombre := limitGrafologiaText(r.FormValue("persona_nombre"), 180)
	clienteNombre := personaNombre
	clienteDocumento := limitGrafologiaText(r.FormValue("cliente_documento"), 80)
	if clienteID > 0 {
		cliente, err := dbpkg.GetClienteByID(dbEmp, empresaID, clienteID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "cliente no encontrado para esta empresa", http.StatusBadRequest)
				return
			}
			log.Printf("[grafologia_ia] cliente empresa_id=%d cliente_id=%d error: %v", empresaID, clienteID, err)
			http.Error(w, "No se pudo validar el cliente asociado", http.StatusInternalServerError)
			return
		}
		clienteNombre = strings.TrimSpace(cliente.NombreRazonSocial)
		clienteDocumento = strings.TrimSpace(strings.TrimSpace(cliente.TipoDocumento) + " " + strings.TrimSpace(cliente.NumeroDocumento))
	}

	ocrTexto := limitGrafologiaText(r.FormValue("ocr_texto"), 3000)
	titulo := limitGrafologiaText(r.FormValue("titulo"), 220)
	if titulo == "" {
		titulo = "Analisis grafologico con GPT-5.5"
	}
	pregunta := buildGrafologiaIAPrompt(titulo, clienteNombre, clienteDocumento, personaDescripcion, personaCaracteristicas, ocrTexto)
	systemPrompt := buildGrafologiaIASystemPrompt()
	ctrl := &EmpresaAIChatController{dbEmp: dbEmp, dbSuper: dbSuper, client: &http.Client{Timeout: 75 * time.Second}}
	att := &aiAttachment{
		Filename: strings.TrimSpace(header.Filename),
		MimeType: mimeType,
		Bytes:    data,
	}
	respuesta, promptTokens, completionTokens, err := ctrl.generateResponseWithSystemPromptAndAttachment(model, pregunta, nil, systemPrompt, att)
	if err != nil {
		if isAICredentialUnavailableError(err) {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":    false,
				"code":  "ai_credentials_unavailable",
				"error": "No se pudo generar analisis GPT-5.5: " + err.Error(),
			})
			return
		}
		http.Error(w, "No se pudo generar analisis GPT-5.5: "+err.Error(), http.StatusBadGateway)
		return
	}
	respuesta = strings.TrimSpace(respuesta)
	if respuesta == "" {
		http.Error(w, "GPT-5.5 no devolvio contenido", http.StatusBadGateway)
		return
	}
	if len([]rune(respuesta)) > 12000 {
		respuesta = string([]rune(respuesta)[:12000])
	}

	planActual := strings.ToLower(strings.TrimSpace(usoActual.PlanActual))
	if planActual == "" {
		planActual = "free"
	}
	if _, err := dbpkg.RegisterEmpresaAIConsulta(dbEmp, dbpkg.EmpresaAIConsulta{
		EmpresaID:        empresaID,
		Provider:         model.Provider,
		ModelID:          model.ID,
		Pregunta:         pregunta,
		Respuesta:        respuesta,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
		FechaConsulta:    time.Now().Format("2006-01-02 15:04:05"),
		PlanActual:       planActual,
		UsuarioCreador:   adminEmailFromRequest(r),
		Estado:           "activo",
		Observaciones:    "grafologia_gpt55",
	}); err != nil {
		http.Error(w, "No se pudo registrar auditoria de IA", http.StatusInternalServerError)
		return
	}
	usoActualizado, _ := dbpkg.GetEmpresaAIUsoDiario(dbEmp, empresaID, model.Provider, model.ID, fechaUso)
	restante := maxGPT55 - usoActualizado.Consultas
	if restante < 0 {
		restante = 0
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":             true,
		"empresa_id":     empresaID,
		"provider":       model.Provider,
		"model_id":       model.ID,
		"display_name":   model.DisplayName,
		"upstream_model": model.UpstreamModel,
		"respuesta":      respuesta,
		"subject": map[string]interface{}{
			"cliente_id":              clienteID,
			"cliente_nombre":          clienteNombre,
			"cliente_documento":       clienteDocumento,
			"persona_descripcion":     personaDescripcion,
			"persona_caracteristicas": personaCaracteristicas,
		},
		"usage": map[string]interface{}{
			"plan":              planActual,
			"daily_used":        usoActualizado.Consultas,
			"daily_limit":       maxGPT55,
			"daily_remaining":   restante,
			"prompt_tokens":     promptTokens,
			"completion_tokens": completionTokens,
		},
	})
}

func buildGrafologiaIASystemPrompt() string {
	return strings.TrimSpace(`Eres un analista profesional de GRAFOLOGIX. Analiza la imagen manuscrita con criterio visual, matematico y heuristico, en espanol claro.

Reglas:
- Usa la imagen como fuente principal y el texto OCR solo como apoyo.
- No emitas diagnosticos clinicos, legales, laborales definitivos ni decisiones automaticas de contratacion.
- No infieras rasgos sensibles o protegidos como salud, religion, orientacion, origen etnico, ideologia o condiciones medicas.
- Si la imagen no permite concluir algo, indicalo con baja confianza y recomienda una nueva captura.
- Entrega resultados orientativos con porcentajes y confianza.

Formato obligatorio:
1. Resumen general.
2. Calidad de la imagen y observaciones tecnicas.
3. Metricas grafologicas: inclinacion, presion, tamano, espaciado entre letras/palabras/lineas, continuidad, direccion de lineas, margenes, velocidad estimada, regularidad y forma de letras.
4. Interpretacion orientativa: organizacion, extroversion/introversion, impulsividad, estabilidad emocional, creatividad, disciplina, sociabilidad, concentracion, seguridad personal, liderazgo y adaptabilidad.
5. Alertas de baja confianza o datos insuficientes.
6. Conclusiones y recomendaciones.`)
}

func buildGrafologiaIAPrompt(titulo, clienteNombre, clienteDocumento, personaDescripcion, personaCaracteristicas, ocrTexto string) string {
	var b strings.Builder
	b.WriteString("Genera un informe grafologico profesional usando GPT-5.5 vision para el manuscrito adjunto.\n")
	if strings.TrimSpace(titulo) != "" {
		b.WriteString("\nTitulo del informe: ")
		b.WriteString(strings.TrimSpace(titulo))
		b.WriteString("\n")
	}
	if strings.TrimSpace(clienteNombre) != "" || strings.TrimSpace(clienteDocumento) != "" || strings.TrimSpace(personaDescripcion) != "" || strings.TrimSpace(personaCaracteristicas) != "" {
		b.WriteString("\nPersona asociada al manuscrito:\n")
		if strings.TrimSpace(clienteNombre) != "" {
			b.WriteString("- Nombre/cliente: ")
			b.WriteString(strings.TrimSpace(clienteNombre))
			b.WriteString("\n")
		}
		if strings.TrimSpace(clienteDocumento) != "" {
			b.WriteString("- Documento/referencia: ")
			b.WriteString(strings.TrimSpace(clienteDocumento))
			b.WriteString("\n")
		}
		if strings.TrimSpace(personaDescripcion) != "" {
			b.WriteString("- Descripcion autorizada: ")
			b.WriteString(strings.TrimSpace(personaDescripcion))
			b.WriteString("\n")
		}
		if strings.TrimSpace(personaCaracteristicas) != "" {
			b.WriteString("- Caracteristicas registradas: ")
			b.WriteString(strings.TrimSpace(personaCaracteristicas))
			b.WriteString("\n")
		}
	}
	if strings.TrimSpace(ocrTexto) != "" {
		b.WriteString("\nTexto OCR o transcripcion opcional:\n")
		b.WriteString(strings.TrimSpace(ocrTexto))
		b.WriteString("\n")
	}
	b.WriteString("\nDevuelve el informe con porcentajes, nivel de confianza y explicaciones concretas por cada metrica. Manten un tono profesional, compacto y apto para imprimir o anexar al historial del cliente.")
	return b.String()
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
		Exportaciones: []string{"HTML imprimible", "PDF real", "Word compatible", "JSON", "CSV", "TXT"},
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
	var preprocess grafologix.PreprocessResult
	_ = json.Unmarshal([]byte(strings.TrimSpace(item.MetricasJSON)), &metrics)
	_ = json.Unmarshal([]byte(strings.TrimSpace(item.InterpretacionJSON)), &traits)
	hasPreprocess := json.Unmarshal([]byte(strings.TrimSpace(item.PreprocesamientoJSON)), &preprocess) == nil && preprocess.Width > 0
	result := grafologix.AnalysisResult{
		Version:     grafologix.EngineVersion,
		GeneratedAt: item.FechaCreacion,
		Subject: &grafologix.SubjectInfo{
			ClienteID:              item.ClienteID,
			ClienteNombre:          item.ClienteNombre,
			ClienteDocumento:       item.ClienteDocumento,
			PersonaDescripcion:     item.PersonaDescripcion,
			PersonaCaracteristicas: item.PersonaCaracteristicas,
		},
		Summary:     item.Resumen,
		GlobalTrust: item.ConfianzaGlobal,
		Metrics:     metrics,
		Traits:      traits,
		TechnicalNotes: []string{
			"Analisis heuristico orientativo: no es diagnostico psicologico, medico, juridico ni prueba de seleccion de personal.",
			"El PDF resume el informe; la vista HTML conserva el detalle completo y las barras visuales.",
		},
	}
	if hasPreprocess {
		result.Preprocess = &preprocess
	}
	return result
}

func parseGrafologiaInt64Form(r *http.Request, key string) (int64, error) {
	value := strings.TrimSpace(r.FormValue(key))
	if value == "" {
		return 0, nil
	}
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("%s invalido", key)
	}
	return n, nil
}

func limitGrafologiaText(value string, max int) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\x00", ""))
	if max <= 0 || len([]rune(value)) <= max {
		return value
	}
	runes := []rune(value)
	return strings.TrimSpace(string(runes[:max]))
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

func saveEmpresaGrafologiaArtifacts(artifacts map[string][]byte, empresaID int64, baseFileName string) map[string]string {
	if len(artifacts) == 0 {
		return map[string]string{}
	}
	folder := fmt.Sprintf("empresa_%d", empresaID)
	dir := filepath.Join(resolveWebRootDir(), "uploads", "empresas", folder, "imagenes", "grafologia", "procesado")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Printf("[grafologia] no se pudo crear carpeta de artefactos empresa_id=%d: %v", empresaID, err)
		return map[string]string{}
	}
	base := strings.TrimSuffix(filepath.Base(baseFileName), filepath.Ext(baseFileName))
	base = sanitizeGrafologiaFilename(base)
	if base == "" {
		base = "analisis"
	}
	out := map[string]string{}
	for key, data := range artifacts {
		cleanKey := sanitizeGrafologiaFilename(key)
		if cleanKey == "" || len(data) == 0 {
			continue
		}
		name := base + "_" + cleanKey + ".png"
		abs := filepath.Join(dir, name)
		if err := os.WriteFile(abs, data, 0o644); err != nil {
			log.Printf("[grafologia] no se pudo guardar artefacto %s empresa_id=%d: %v", cleanKey, empresaID, err)
			continue
		}
		out[cleanKey] = "/uploads/empresas/" + folder + "/imagenes/grafologia/procesado/" + name
	}
	return out
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
