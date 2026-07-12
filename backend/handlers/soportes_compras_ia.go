package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const maxSoporteComprasIAUploadBytes = 15 << 20

var soporteComprasIAAllowedExt = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".webp": true,
	".pdf":  true,
	".xml":  true,
}

// EmpresaSoportesComprasIAHandler administra la captura inteligente de compras y gastos.
func EmpresaSoportesComprasIAHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil || empresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "" {
			action = "dashboard"
		}

		switch r.Method {
		case http.MethodGet:
			handleSoportesComprasIAGet(w, r, dbEmp, empresaID, action)
		case http.MethodPost, http.MethodPut:
			handleSoportesComprasIAMutate(w, r, dbEmp, dbSuper, empresaID, action)
		default:
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
		}
	}
}

func handleSoportesComprasIAGet(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, empresaID int64, action string) {
	switch action {
	case "dashboard":
		dashboard, err := dbpkg.BuildEmpresaSoportesComprasIADashboard(dbEmp, empresaID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "dashboard": dashboard})
	case "soportes":
		estado := r.URL.Query().Get("estado")
		rows, err := dbpkg.ListEmpresaSoportesComprasIA(dbEmp, empresaID, estado, 300)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for i := range rows {
			exposeSoporteComprasIAURL(&rows[i])
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "soportes": rows})
	case "descargar":
		downloadSoporteComprasIA(w, r, dbEmp, empresaID)
	case "eventos":
		soporteID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("soporte_id")), 10, 64)
		rows, err := dbpkg.ListEmpresaSoportesComprasIAEventos(dbEmp, empresaID, soporteID, 200)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "eventos": rows})
	default:
		http.Error(w, "accion no soportada", http.StatusBadRequest)
	}
}

func handleSoportesComprasIAMutate(w http.ResponseWriter, r *http.Request, dbEmp, dbSuper *sql.DB, empresaID int64, action string) {
	usuario := strings.TrimSpace(adminEmailFromRequest(r))
	switch action {
	case "radicar":
		row, err := radicarSoporteComprasIA(r, dbEmp, empresaID, usuario)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		exposeSoporteComprasIAURL(&row)
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "soporte": row})
	case "extraer_ia":
		row, err := extraerSoporteComprasIAGPT55(r, dbEmp, dbSuper, empresaID, usuario)
		if err != nil {
			status := http.StatusBadRequest
			if errors.Is(err, errSoporteComprasIAIADesactivada) || errors.Is(err, errSoporteComprasIAModeloNoDisponible) {
				status = http.StatusServiceUnavailable
			}
			http.Error(w, err.Error(), status)
			return
		}
		exposeSoporteComprasIAURL(&row)
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "soporte": row})
	case "aprobar":
		payload, _ := decodeSoporteComprasIAActionPayload(r)
		row, err := dbpkg.UpdateEmpresaSoporteComprasIAEstado(dbEmp, empresaID, payload.SoporteID, "aprobado", usuario, payload.Observaciones)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		exposeSoporteComprasIAURL(&row)
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "soporte": row})
	case "rechazar":
		payload, _ := decodeSoporteComprasIAActionPayload(r)
		row, err := dbpkg.UpdateEmpresaSoporteComprasIAEstado(dbEmp, empresaID, payload.SoporteID, "rechazado", usuario, payload.Observaciones)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		exposeSoporteComprasIAURL(&row)
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "soporte": row})
	case "contabilizar":
		payload, _ := decodeSoporteComprasIAActionPayload(r)
		row, err := dbpkg.ContabilizarEmpresaSoporteComprasIA(dbEmp, empresaID, payload.SoporteID, usuario)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		exposeSoporteComprasIAURL(&row)
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "soporte": row})
	case "seed_demo":
		row, err := seedSoporteComprasIADemo(dbEmp, empresaID, usuario)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "soporte": row})
	default:
		http.Error(w, "accion no soportada", http.StatusBadRequest)
	}
}

type soporteComprasIAActionPayload struct {
	SoporteID     int64  `json:"soporte_id"`
	Observaciones string `json:"observaciones"`
}

func decodeSoporteComprasIAActionPayload(r *http.Request) (soporteComprasIAActionPayload, error) {
	var p soporteComprasIAActionPayload
	if strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
		err := json.NewDecoder(r.Body).Decode(&p)
		return p, err
	}
	p.SoporteID, _ = strconv.ParseInt(strings.TrimSpace(r.FormValue("soporte_id")), 10, 64)
	p.Observaciones = strings.TrimSpace(r.FormValue("observaciones"))
	return p, nil
}

func radicarSoporteComprasIA(r *http.Request, dbEmp *sql.DB, empresaID int64, usuario string) (dbpkg.EmpresaSoporteComprasIA, error) {
	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	var row dbpkg.EmpresaSoporteComprasIA
	row.EmpresaID = empresaID
	row.Usuario = usuario
	row.ModeloIA = dbpkg.EmpresaSoporteComprasIAModeloDefault
	row.EstadoSoporte = "radicado"
	row.RequiereRevisionHumana = true

	if strings.Contains(contentType, "multipart/form-data") {
		att, err := parseSingleAttachmentFromMultipart(r, "archivo", maxSoporteComprasIAUploadBytes)
		if err != nil {
			return row, err
		}
		if att == nil {
			att, err = parseSingleAttachmentFromMultipart(r, "soporte", maxSoporteComprasIAUploadBytes)
			if err != nil {
				return row, err
			}
		}
		if att == nil {
			att, err = parseSingleAttachmentFromMultipart(r, "file", maxSoporteComprasIAUploadBytes)
			if err != nil {
				return row, err
			}
		}
		row = soporteComprasIAFromForm(r, row)
		if att != nil {
			url, name, mimeType, hash, origen, err := saveSoporteComprasIAAttachment(att, empresaID)
			if err != nil {
				return row, err
			}
			row.ArchivoURL = url
			row.ArchivoNombre = name
			row.ArchivoMime = mimeType
			row.ArchivoHash = hash
			row.Origen = origen
		}
		return dbpkg.CreateEmpresaSoporteComprasIA(dbEmp, row)
	}

	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&row); err != nil {
		return row, err
	}
	row.EmpresaID = empresaID
	row.Usuario = usuario
	row.ModeloIA = dbpkg.EmpresaSoporteComprasIAModeloDefault
	if row.EstadoSoporte == "" {
		row.EstadoSoporte = "radicado"
	}
	return dbpkg.CreateEmpresaSoporteComprasIA(dbEmp, row)
}

func soporteComprasIAFromForm(r *http.Request, row dbpkg.EmpresaSoporteComprasIA) dbpkg.EmpresaSoporteComprasIA {
	row.TipoSoporte = strings.TrimSpace(r.FormValue("tipo_soporte"))
	row.DocumentoTipo = strings.TrimSpace(r.FormValue("documento_tipo"))
	row.DocumentoNumero = strings.TrimSpace(r.FormValue("documento_numero"))
	row.ProveedorNombre = strings.TrimSpace(r.FormValue("proveedor_nombre"))
	row.ProveedorNIT = strings.TrimSpace(r.FormValue("proveedor_nit"))
	row.FechaDocumento = strings.TrimSpace(r.FormValue("fecha_documento"))
	row.FechaVencimiento = strings.TrimSpace(r.FormValue("fecha_vencimiento"))
	row.CategoriaContable = strings.TrimSpace(r.FormValue("categoria_contable"))
	row.CentroCosto = strings.TrimSpace(r.FormValue("centro_costo"))
	row.Moneda = strings.TrimSpace(r.FormValue("moneda"))
	row.Observaciones = strings.TrimSpace(r.FormValue("observaciones"))
	row.ImpactaInventario = parseBoolSoporteComprasIA(r.FormValue("impacta_inventario"))
	row.Subtotal = parseFloatSoporteComprasIA(r.FormValue("subtotal"))
	row.ImpuestoIVA = parseFloatSoporteComprasIA(r.FormValue("impuesto_iva"))
	row.RetencionFuente = parseFloatSoporteComprasIA(r.FormValue("retencion_fuente"))
	row.RetencionICA = parseFloatSoporteComprasIA(r.FormValue("retencion_ica"))
	row.RetencionIVA = parseFloatSoporteComprasIA(r.FormValue("retencion_iva"))
	row.Total = parseFloatSoporteComprasIA(r.FormValue("total"))
	return row
}

func saveSoporteComprasIAAttachment(att *aiAttachment, empresaID int64) (string, string, string, string, string, error) {
	if att == nil || len(att.Bytes) == 0 {
		return "", "", "", "", "manual", nil
	}
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(att.Filename)))
	if ext == "" {
		ext = extFromSoporteComprasIAMime(att.MimeType)
	}
	if !soporteComprasIAAllowedExt[ext] {
		return "", "", "", "", "", fmt.Errorf("extension no permitida para soporte")
	}
	root := soporteComprasIAPrivateRoot()
	dir := filepath.Join(root, fmt.Sprintf("empresa_%d", empresaID))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", "", "", "", "", err
	}
	randomName := make([]byte, 32)
	if _, err := rand.Read(randomName); err != nil {
		return "", "", "", "", "", fmt.Errorf("no se pudo generar nombre privado: %w", err)
	}
	fileName := hex.EncodeToString(randomName) + ext
	absPath := filepath.Join(dir, fileName)
	if err := os.WriteFile(absPath, att.Bytes, 0o600); err != nil {
		return "", "", "", "", "", err
	}
	mimeType := strings.TrimSpace(att.MimeType)
	if mimeType == "" {
		mimeType = mimeFromSoporteComprasIAExt(ext)
	}
	origen := origenFromSoporteComprasIAExt(ext, mimeType)
	url := "private://soportes_compras_ia/" + fmt.Sprintf("empresa_%d", empresaID) + "/" + fileName
	return url, fileName, mimeType, dbpkg.EmpresaSoporteComprasIAHashBytes(att.Bytes), origen, nil
}

func soporteComprasIAPrivateRoot() string {
	if configured := strings.TrimSpace(os.Getenv("PCS_PRIVATE_STORAGE_DIR")); configured != "" {
		return filepath.Join(configured, "soportes_compras_ia")
	}
	return filepath.Join(resolveProjectRootDir(), "private_storage", "soportes_compras_ia")
}

func soporteComprasIADownloadURL(empresaID, soporteID int64) string {
	return "/api/empresa/soportes_compras_ia?empresa_id=" + strconv.FormatInt(empresaID, 10) + "&action=descargar&soporte_id=" + strconv.FormatInt(soporteID, 10)
}

func exposeSoporteComprasIAURL(row *dbpkg.EmpresaSoporteComprasIA) {
	if row != nil && row.ID > 0 && strings.HasPrefix(strings.TrimSpace(row.ArchivoURL), "private://") {
		row.ArchivoURL = soporteComprasIADownloadURL(row.EmpresaID, row.ID)
	}
}

func downloadSoporteComprasIA(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, empresaID int64) {
	soporteID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("soporte_id")), 10, 64)
	if err != nil || soporteID <= 0 {
		http.Error(w, "soporte invalido", http.StatusBadRequest)
		return
	}
	row, err := dbpkg.GetEmpresaSoporteComprasIA(dbEmp, empresaID, soporteID)
	if err != nil {
		http.Error(w, "archivo no disponible", http.StatusNotFound)
		return
	}
	path, err := safeSoporteComprasIAPathFromURL(row.ArchivoURL)
	if err != nil {
		http.Error(w, "archivo no disponible", http.StatusNotFound)
		return
	}
	file, err := os.Open(path) // #nosec G304 -- path was resolved under the private tenant root without symlinks.
	if err != nil {
		http.Error(w, "archivo no disponible", http.StatusNotFound)
		return
	}
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		http.Error(w, "archivo no disponible", http.StatusNotFound)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", strings.TrimSpace(row.ArchivoMime))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", sanitizeComprobanteBaseName(row.ArchivoNombre)+filepath.Ext(path)))
	http.ServeContent(w, r, filepath.Base(path), info.ModTime(), file)
}

var (
	errSoporteComprasIAIADesactivada      = errors.New("la IA esta desactivada desde configuracion avanzada")
	errSoporteComprasIAModeloNoDisponible = errors.New("el modelo openai:gpt-5.5 no esta disponible en el catalogo de IA")
)

func extraerSoporteComprasIAGPT55(r *http.Request, dbEmp, dbSuper *sql.DB, empresaID int64, usuario string) (dbpkg.EmpresaSoporteComprasIA, error) {
	if !isSuperAIEnabled(dbSuper) {
		return dbpkg.EmpresaSoporteComprasIA{}, errSoporteComprasIAIADesactivada
	}
	model, ok := availableEmpresaAIModelMap(dbSuper)[dbpkg.EmpresaSoporteComprasIAModeloDefault]
	if !ok {
		return dbpkg.EmpresaSoporteComprasIA{}, errSoporteComprasIAModeloNoDisponible
	}
	if _, _, err := reserveEmpresaAgentAdvancedUsage(dbEmp, dbSuper, empresaID, usuario); err != nil {
		return dbpkg.EmpresaSoporteComprasIA{}, err
	}
	payload, _ := decodeSoporteComprasIAActionPayload(r)
	if payload.SoporteID <= 0 {
		payload.SoporteID, _ = strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("soporte_id")), 10, 64)
	}
	if payload.SoporteID <= 0 {
		return dbpkg.EmpresaSoporteComprasIA{}, errors.New("soporte_id es obligatorio")
	}
	current, err := dbpkg.GetEmpresaSoporteComprasIA(dbEmp, empresaID, payload.SoporteID)
	if err != nil {
		return dbpkg.EmpresaSoporteComprasIA{}, err
	}
	att, err := loadSoporteComprasIAAttachment(current)
	if err != nil {
		return dbpkg.EmpresaSoporteComprasIA{}, err
	}

	ctrl := NewEmpresaAIChatController(dbEmp, dbSuper)
	systemPrompt := soporteComprasIASystemPrompt()
	pregunta := "Extrae y normaliza este soporte de compra o gasto de Colombia. Responde solo JSON valido, sin explicaciones."
	respuesta, promptTokens, completionTokens, err := ctrl.callOpenAIResponsesWithSystemPrompt(model, pregunta, nil, systemPrompt, att)
	if err != nil {
		return dbpkg.EmpresaSoporteComprasIA{}, err
	}
	extracted, compactJSON, err := parseSoporteComprasIAExtraction(respuesta)
	if err != nil {
		return dbpkg.EmpresaSoporteComprasIA{}, err
	}
	extracted.ExtraccionJSON = compactJSON
	extracted.RespuestaIA = respuesta
	extracted.ModeloIA = model.ID
	extracted.Usuario = usuario
	updated, err := dbpkg.UpdateEmpresaSoporteComprasIAExtraccion(dbEmp, empresaID, payload.SoporteID, extracted, usuario)
	if err != nil {
		return dbpkg.EmpresaSoporteComprasIA{}, err
	}
	_, _ = dbpkg.RegisterEmpresaAIConsulta(dbEmp, dbpkg.EmpresaAIConsulta{
		EmpresaID:        empresaID,
		Provider:         model.Provider,
		ModelID:          model.ID,
		Pregunta:         fmt.Sprintf("captura_inteligente_compras_gastos soporte_id=%d codigo=%s", payload.SoporteID, current.Codigo),
		Respuesta:        respuesta,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
		UsuarioCreador:   usuario,
		Estado:           "activo",
		Observaciones:    "Extraccion IA GPT-5.5 de soporte de compra o gasto",
	})
	return updated, nil
}

func soporteComprasIASystemPrompt() string {
	return `Eres un motor profesional de captura inteligente con IA GPT-5.5 para documentos empresariales en Colombia.
Lee fotos, PDFs o XML de facturas de compra, documentos soporte, cuentas de cobro, recibos, gastos, ingresos, comprobantes de caja y cartas/listas de precios.
Devuelve exclusivamente JSON valido con estas claves:
{
  "tipo_soporte": "compra|gasto|ingreso|documento_soporte|recibo|servicio|carta_precios",
  "proveedor_nombre": "",
  "proveedor_nit": "",
  "documento_tipo": "factura_compra|documento_soporte|cuenta_cobro|recibo_caja|gasto|ingreso|lista_precios|otro",
  "documento_numero": "",
  "fecha_documento": "YYYY-MM-DD",
  "fecha_vencimiento": "YYYY-MM-DD",
  "subtotal": 0,
  "impuesto_iva": 0,
  "retencion_fuente": 0,
  "retencion_ica": 0,
  "retencion_iva": 0,
  "total": 0,
  "moneda": "COP",
  "categoria_contable": "",
  "centro_costo": "",
  "impacta_inventario": false,
  "confianza_ia": 0.0,
  "requiere_revision_humana": true,
  "lineas_detectadas": [],
  "observaciones": ""
}
Usa numeros sin separadores de miles. Si falta un dato, deja cadena vacia o 0. Marca requiere_revision_humana=true cuando haya baja confianza, documento borroso, totales inconsistentes o datos tributarios incompletos.`
}

func parseSoporteComprasIAExtraction(raw string) (dbpkg.EmpresaSoporteComprasIA, string, error) {
	candidate := extractJSONCandidate(raw)
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(candidate), &data); err != nil {
		return dbpkg.EmpresaSoporteComprasIA{}, "", fmt.Errorf("respuesta IA no es JSON valido: %w", err)
	}
	compactBytes, _ := json.Marshal(data)
	row := dbpkg.EmpresaSoporteComprasIA{
		TipoSoporte:            stringFromMap(data, "tipo_soporte"),
		ProveedorNombre:        stringFromMap(data, "proveedor_nombre"),
		ProveedorNIT:           stringFromMap(data, "proveedor_nit"),
		DocumentoTipo:          stringFromMap(data, "documento_tipo"),
		DocumentoNumero:        stringFromMap(data, "documento_numero"),
		FechaDocumento:         normalizeDateString(stringFromMap(data, "fecha_documento")),
		FechaVencimiento:       normalizeDateString(stringFromMap(data, "fecha_vencimiento")),
		Subtotal:               floatFromMap(data, "subtotal"),
		ImpuestoIVA:            floatFromMap(data, "impuesto_iva"),
		RetencionFuente:        floatFromMap(data, "retencion_fuente"),
		RetencionICA:           floatFromMap(data, "retencion_ica"),
		RetencionIVA:           floatFromMap(data, "retencion_iva"),
		Total:                  floatFromMap(data, "total"),
		Moneda:                 stringFromMap(data, "moneda"),
		CategoriaContable:      stringFromMap(data, "categoria_contable"),
		CentroCosto:            stringFromMap(data, "centro_costo"),
		ImpactaInventario:      boolFromMap(data, "impacta_inventario"),
		ConfianzaIA:            floatFromMap(data, "confianza_ia"),
		RequiereRevisionHumana: boolFromMap(data, "requiere_revision_humana"),
		Observaciones:          stringFromMap(data, "observaciones"),
	}
	return row, string(compactBytes), nil
}

func extractJSONCandidate(raw string) string {
	clean := strings.TrimSpace(raw)
	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimPrefix(clean, "```")
	clean = strings.TrimSuffix(clean, "```")
	clean = strings.TrimSpace(clean)
	start := strings.Index(clean, "{")
	end := strings.LastIndex(clean, "}")
	if start >= 0 && end > start {
		return clean[start : end+1]
	}
	return clean
}

func loadSoporteComprasIAAttachment(row dbpkg.EmpresaSoporteComprasIA) (*aiAttachment, error) {
	if strings.TrimSpace(row.ArchivoURL) == "" {
		return nil, errors.New("el soporte no tiene archivo adjunto para analisis IA")
	}
	path, err := safeSoporteComprasIAPathFromURL(row.ArchivoURL)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path) // #nosec G304 -- path is an existing regular file resolved inside the private support root above.
	if err != nil {
		return nil, err
	}
	return &aiAttachment{Filename: row.ArchivoNombre, MimeType: row.ArchivoMime, Bytes: b}, nil
}

func safeSoporteComprasIAPathFromURL(url string) (string, error) {
	clean := strings.TrimSpace(strings.TrimPrefix(url, "private://soportes_compras_ia/"))
	if clean == strings.TrimSpace(url) || clean == "" {
		return "", errors.New("ruta de soporte no permitida")
	}
	root := filepath.Clean(soporteComprasIAPrivateRoot())
	abs := filepath.Clean(filepath.Join(root, filepath.FromSlash(clean)))
	if abs != root && !strings.HasPrefix(abs, root+string(os.PathSeparator)) {
		return "", errors.New("ruta de soporte fuera del directorio permitido")
	}
	return resolveExistingPrivateFileUnderRoot(root, abs)
}

// resolveExistingPrivateFileUnderRoot rejects traversal and symlink escapes before
// an attachment is sent to an external analysis service.
func resolveExistingPrivateFileUnderRoot(root, candidate string) (string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", errors.New("directorio privado no disponible")
	}
	absCandidate, err := filepath.Abs(candidate)
	if err != nil {
		return "", errors.New("archivo privado no disponible")
	}
	rel, err := filepath.Rel(absRoot, absCandidate)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || filepath.IsAbs(rel) {
		return "", errors.New("ruta de soporte fuera del directorio permitido")
	}
	info, err := os.Lstat(absCandidate)
	if err != nil || !info.Mode().IsRegular() {
		return "", errors.New("archivo de soporte no disponible")
	}
	resolvedRoot, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		return "", errors.New("directorio privado no disponible")
	}
	resolvedCandidate, err := filepath.EvalSymlinks(absCandidate)
	if err != nil {
		return "", errors.New("archivo de soporte no disponible")
	}
	resolvedRel, err := filepath.Rel(resolvedRoot, resolvedCandidate)
	if err != nil || strings.HasPrefix(resolvedRel, ".."+string(os.PathSeparator)) || resolvedRel == ".." || filepath.IsAbs(resolvedRel) {
		return "", errors.New("enlace de soporte fuera del directorio permitido")
	}
	return resolvedCandidate, nil
}

func seedSoporteComprasIADemo(dbEmp *sql.DB, empresaID int64, usuario string) (dbpkg.EmpresaSoporteComprasIA, error) {
	row, err := dbpkg.CreateEmpresaSoporteComprasIA(dbEmp, dbpkg.EmpresaSoporteComprasIA{
		EmpresaID:              empresaID,
		TipoSoporte:            "gasto",
		Origen:                 "manual",
		ProveedorNombre:        "Papeleria Centro Empresarial SAS",
		ProveedorNIT:           "901234567-8",
		DocumentoTipo:          "factura_compra",
		DocumentoNumero:        "FE-1024",
		FechaDocumento:         time.Now().Format("2006-01-02"),
		FechaVencimiento:       time.Now().AddDate(0, 0, 15).Format("2006-01-02"),
		Subtotal:               180000,
		ImpuestoIVA:            34200,
		Total:                  214200,
		Moneda:                 "COP",
		CategoriaContable:      "Gastos administrativos",
		CentroCosto:            "Administracion",
		ConfianzaIA:            0.94,
		ModeloIA:               dbpkg.EmpresaSoporteComprasIAModeloDefault,
		RequiereRevisionHumana: true,
		Usuario:                usuario,
		Observaciones:          "Soporte de ejemplo para probar captura inteligente.",
	})
	if err != nil {
		return row, err
	}
	extraction := row
	raw, _ := json.Marshal(map[string]interface{}{
		"proveedor_nombre": row.ProveedorNombre,
		"proveedor_nit":    row.ProveedorNIT,
		"documento_tipo":   row.DocumentoTipo,
		"documento_numero": row.DocumentoNumero,
		"total":            row.Total,
		"confianza_ia":     row.ConfianzaIA,
	})
	extraction.ExtraccionJSON = string(raw)
	return dbpkg.UpdateEmpresaSoporteComprasIAExtraccion(dbEmp, empresaID, row.ID, extraction, usuario)
}

func extFromSoporteComprasIAMime(mimeType string) string {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	case "application/pdf":
		return ".pdf"
	case "application/xml", "text/xml":
		return ".xml"
	default:
		return ".bin"
	}
}

func mimeFromSoporteComprasIAExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".pdf":
		return "application/pdf"
	case ".xml":
		return "application/xml"
	default:
		return "application/octet-stream"
	}
}

func origenFromSoporteComprasIAExt(ext, mimeType string) string {
	if strings.HasPrefix(strings.ToLower(mimeType), "image/") {
		return "foto"
	}
	switch strings.ToLower(ext) {
	case ".pdf":
		return "pdf"
	case ".xml":
		return "xml"
	default:
		return "manual"
	}
}

func parseFloatSoporteComprasIA(raw string) float64 {
	raw = strings.ReplaceAll(strings.TrimSpace(raw), ",", ".")
	v, _ := strconv.ParseFloat(raw, 64)
	return v
}

func parseBoolSoporteComprasIA(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "on", "si", "sí", "yes":
		return true
	default:
		return false
	}
}

func stringFromMap(data map[string]interface{}, key string) string {
	if data == nil {
		return ""
	}
	raw, ok := data[key]
	if !ok || raw == nil {
		return ""
	}
	switch v := raw.(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func floatFromMap(data map[string]interface{}, key string) float64 {
	if data == nil {
		return 0
	}
	switch v := data[key].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case string:
		return parseFloatSoporteComprasIA(v)
	default:
		return 0
	}
}

func boolFromMap(data map[string]interface{}, key string) bool {
	if data == nil {
		return false
	}
	switch v := data[key].(type) {
	case bool:
		return v
	case string:
		return parseBoolSoporteComprasIA(v)
	case float64:
		return v != 0
	default:
		return false
	}
}

func normalizeDateString(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	layouts := []string{"2006-01-02", "02/01/2006", "02-01-2006", time.RFC3339}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.Format("2006-01-02")
		}
	}
	if len(raw) >= 10 {
		return raw[:10]
	}
	return raw
}
