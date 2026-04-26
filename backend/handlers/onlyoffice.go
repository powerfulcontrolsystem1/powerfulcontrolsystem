package handlers

import (
	"archive/zip"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// OnlyOffice / Documentos por empresa
//
// - Almacenamiento: /data/empresas/{empresa_id}/documentos/
// - Listado/subida/descarga: protegido por permisos de empresa.
// - Acceso desde OnlyOffice: por token temporal (no sesión), evita cruce entre empresas.
// - JWT OnlyOffice HS256: implementado con stdlib (HMAC-SHA256).

const (
	onlyOfficeDefaultDataRoot = "/data/empresas"
	onlyOfficeConfigKeyDSURL  = "onlyoffice.document_server_url" // ej: http://onlyoffice:80
	onlyOfficeConfigKeyJWT    = "onlyoffice.jwt_secret"          // secreto HS256
)

type onlyOfficeAccessTokenClaims struct {
	EmpresaID int64  `json:"empresa_id"`
	Path      string `json:"path"`   // ruta relativa dentro de /data/empresas/{empresa_id}/documentos/
	Action    string `json:"action"` // file|callback
	ExpUnix   int64  `json:"exp"`
	Nonce     string `json:"nonce,omitempty"`
}

type onlyOfficeDocListItem struct {
	Name      string `json:"name"`
	SizeBytes int64  `json:"size_bytes"`
	UpdatedAt string `json:"updated_at"`
	Ext       string `json:"ext"`
}

type onlyOfficeEditorConfigRequest struct {
	EmpresaID   int64  `json:"empresa_id"`
	FileName   string `json:"file_name"`
	Mode       string `json:"mode"` // edit|view
	UserID     string `json:"user_id,omitempty"`
	UserName   string `json:"user_name,omitempty"`
	Download   bool   `json:"download,omitempty"`
}

// OnlyOffice callback payload (subset usado).
type onlyOfficeCallbackPayload struct {
	Status int    `json:"status"`
	URL    string `json:"url"`
	Key    string `json:"key"`
	Error  int    `json:"error"`
}

type onlyOfficeCreateDocRequest struct {
	Tipo   string `json:"tipo"`             // word|excel|powerpoint|docx|xlsx|pptx
	Nombre string `json:"nombre,omitempty"` // opcional, sin path
}

func onlyOfficeDataRoot() string {
	if v := strings.TrimSpace(os.Getenv("PCS_DATA_ROOT")); v != "" {
		return v
	}
	return onlyOfficeDefaultDataRoot
}

func onlyOfficeEmpresaDocsDir(empresaID int64) (string, error) {
	if empresaID <= 0 {
		return "", fmt.Errorf("empresa_id invalido")
	}
	root := onlyOfficeDataRoot()
	dir := filepath.Join(root, "empresas", fmt.Sprintf("%d", empresaID), "documentos")
	// Normalizamos a formato del FS local.
	dir = filepath.Clean(dir)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return "", err
	}
	return dir, nil
}

func onlyOfficeSafeBaseName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("nombre de archivo invalido")
	}
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." {
		return "", fmt.Errorf("nombre de archivo invalido")
	}
	return name, nil
}

func onlyOfficeRelPathForEmpresaFile(fileName string) (string, error) {
	base, err := onlyOfficeSafeBaseName(fileName)
	if err != nil {
		return "", err
	}
	// Solo un nivel: documentos/<archivo>
	return base, nil
}

func onlyOfficeJoinEmpresaFile(empresaID int64, rel string) (string, error) {
	dir, err := onlyOfficeEmpresaDocsDir(empresaID)
	if err != nil {
		return "", err
	}
	rel = strings.TrimSpace(rel)
	rel = filepath.Base(rel)
	full := filepath.Join(dir, rel)
	full = filepath.Clean(full)
	// Garantizar confinamiento dentro de dir.
	dirL := strings.ToLower(dir)
	fullL := strings.ToLower(full)
	if !strings.HasPrefix(fullL, dirL+strings.ToLower(string(filepath.Separator))) && !strings.EqualFold(full, filepath.Join(dir, rel)) {
		return "", fmt.Errorf("ruta no permitida")
	}
	return full, nil
}

func onlyOfficeB64URL(b []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}

func onlyOfficeSignToken(secret string, claims onlyOfficeAccessTokenClaims) (string, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return "", fmt.Errorf("onlyoffice jwt_secret no configurado")
	}
	raw, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(raw)
	sig := mac.Sum(nil)
	return onlyOfficeB64URL(raw) + "." + onlyOfficeB64URL(sig), nil
}

func onlyOfficeVerifyToken(secret, token string) (onlyOfficeAccessTokenClaims, error) {
	var out onlyOfficeAccessTokenClaims
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return out, fmt.Errorf("onlyoffice jwt_secret no configurado")
	}
	token = strings.TrimSpace(token)
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return out, fmt.Errorf("token invalido")
	}
	payloadB64 := parts[0]
	sigB64 := parts[1]
	payloadRaw, err := base64.URLEncoding.DecodeString(payloadB64 + strings.Repeat("=", (4-len(payloadB64)%4)%4))
	if err != nil {
		return out, fmt.Errorf("token invalido")
	}
	sigRaw, err := base64.URLEncoding.DecodeString(sigB64 + strings.Repeat("=", (4-len(sigB64)%4)%4))
	if err != nil {
		return out, fmt.Errorf("token invalido")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payloadRaw)
	expected := mac.Sum(nil)
	if !hmac.Equal(sigRaw, expected) {
		return out, fmt.Errorf("token firma invalida")
	}
	if err := json.Unmarshal(payloadRaw, &out); err != nil {
		return out, fmt.Errorf("token invalido")
	}
	if out.EmpresaID <= 0 {
		return out, fmt.Errorf("token invalido")
	}
	if out.ExpUnix > 0 && time.Now().Unix() > out.ExpUnix {
		return out, fmt.Errorf("token expirado")
	}
	// path debe ser simple (basename)
	out.Path = filepath.Base(strings.TrimSpace(out.Path))
	if out.Path == "" || out.Path == "." || out.Path == ".." {
		return out, fmt.Errorf("token invalido")
	}
	return out, nil
}

// JWT HS256 para OnlyOffice (token que va dentro de la config).
func onlyOfficeJWTSignHS256(secret string, payload any) (string, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return "", fmt.Errorf("onlyoffice jwt_secret no configurado")
	}
	header := map[string]any{"alg": "HS256", "typ": "JWT"}
	hb, _ := json.Marshal(header)
	pb, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	h64 := onlyOfficeB64URL(hb)
	p64 := onlyOfficeB64URL(pb)
	signing := h64 + "." + p64
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signing))
	sig := onlyOfficeB64URL(mac.Sum(nil))
	return signing + "." + sig, nil
}

func onlyOfficeGuessFileType(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".docx":
		return "docx"
	case ".xlsx":
		return "xlsx"
	case ".pptx":
		return "pptx"
	case ".doc":
		return "doc"
	case ".xls":
		return "xls"
	case ".ppt":
		return "ppt"
	default:
		// OnlyOffice soporta muchos, pero limitamos a office clásico.
		return strings.TrimPrefix(ext, ".")
	}
}

func onlyOfficeDocumentTypeByExt(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".doc", ".docx", ".odt", ".rtf", ".txt":
		return "word"
	case ".xls", ".xlsx", ".ods", ".csv":
		return "cell"
	case ".ppt", ".pptx", ".odp":
		return "slide"
	default:
		return "word"
	}
}

func onlyOfficeMIMEByExt(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case ".doc":
		return "application/msword"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".ppt":
		return "application/vnd.ms-powerpoint"
	default:
		return "application/octet-stream"
	}
}

func onlyOfficeZipBytes(files map[string]string) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			_ = zw.Close()
			return nil, err
		}
		if _, err := w.Write([]byte(content)); err != nil {
			_ = zw.Close()
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func onlyOfficeBuildEmptyDOCX() ([]byte, error) {
	// OOXML mínimo para un documento vacío.
	return onlyOfficeZipBytes(map[string]string{
		"[Content_Types].xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`,
		"_rels/.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`,
		"word/document.xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t></w:t></w:r></w:p>
    <w:sectPr/>
  </w:body>
</w:document>`,
	})
}

func onlyOfficeBuildEmptyXLSX() ([]byte, error) {
	// OOXML mínimo con una hoja Sheet1.
	return onlyOfficeZipBytes(map[string]string{
		"[Content_Types].xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
</Types>`,
		"_rels/.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`,
		"xl/workbook.xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"
 xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="Sheet1" sheetId="1" r:id="rId1"/>
  </sheets>
</workbook>`,
		"xl/_rels/workbook.xml.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
</Relationships>`,
		"xl/worksheets/sheet1.xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData/>
</worksheet>`,
	})
}

func onlyOfficeBuildEmptyPPTX() ([]byte, error) {
	// OOXML mínimo con una diapositiva.
	return onlyOfficeZipBytes(map[string]string{
		"[Content_Types].xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/ppt/presentation.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"/>
  <Override PartName="/ppt/slides/slide1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slide+xml"/>
</Types>`,
		"_rels/.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="ppt/presentation.xml"/>
</Relationships>`,
		"ppt/presentation.xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:presentation xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"
 xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <p:sldIdLst>
    <p:sldId id="256" r:id="rId1"/>
  </p:sldIdLst>
</p:presentation>`,
		"ppt/_rels/presentation.xml.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide1.xml"/>
</Relationships>`,
		"ppt/slides/slide1.xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"
 xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main">
  <p:cSld>
    <p:spTree>
      <p:nvGrpSpPr>
        <p:cNvPr id="1" name=""/>
        <p:cNvGrpSpPr/>
        <p:nvPr/>
      </p:nvGrpSpPr>
      <p:grpSpPr/>
    </p:spTree>
  </p:cSld>
</p:sld>`,
	})
}

func onlyOfficeNormalizeCreateTipo(tipo string) (ext string) {
	t := strings.ToLower(strings.TrimSpace(tipo))
	switch t {
	case "word", "docx", "documento", "doc":
		return "docx"
	case "excel", "xlsx", "hoja", "sheet", "xls":
		return "xlsx"
	case "powerpoint", "pptx", "presentacion", "presentación", "ppt":
		return "pptx"
	default:
		return ""
	}
}

func onlyOfficeDefaultNameByExt(ext string) string {
	ts := time.Now().Format("20060102_150405")
	switch ext {
	case "docx":
		return "Documento_" + ts + ".docx"
	case "xlsx":
		return "Hoja_" + ts + ".xlsx"
	case "pptx":
		return "Presentacion_" + ts + ".pptx"
	default:
		return "Documento_" + ts
	}
}

func onlyOfficeEnsureUniqueName(empresaID int64, name string) (string, error) {
	base, err := onlyOfficeSafeBaseName(name)
	if err != nil {
		return "", err
	}
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	tryName := base
	for i := 0; i < 200; i++ {
		full, err := onlyOfficeJoinEmpresaFile(empresaID, tryName)
		if err != nil {
			return "", err
		}
		_, statErr := os.Stat(full)
		if statErr != nil && errors.Is(statErr, os.ErrNotExist) {
			return tryName, nil
		}
		tryName = fmt.Sprintf("%s_%d%s", stem, i+1, ext)
	}
	return "", fmt.Errorf("no se pudo reservar un nombre único")
}

func OnlyOfficeDocumentosHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if dbSuper != nil && !isOnlyOfficeEnabled(dbSuper) {
			http.Error(w, "OnlyOffice está desactivado por super administrador.", http.StatusServiceUnavailable)
			return
		}
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "" {
			action = "list"
		}
		switch action {
		case "list":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			dir, err := onlyOfficeEmpresaDocsDir(empresaID)
			if err != nil {
				http.Error(w, "no se pudo preparar carpeta de documentos", http.StatusInternalServerError)
				return
			}
			entries, err := os.ReadDir(dir)
			if err != nil {
				http.Error(w, "no se pudo listar documentos", http.StatusInternalServerError)
				return
			}
			out := make([]onlyOfficeDocListItem, 0, len(entries))
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				info, ierr := e.Info()
				if ierr != nil {
					continue
				}
				name := e.Name()
				out = append(out, onlyOfficeDocListItem{
					Name:      name,
					SizeBytes: info.Size(),
					UpdatedAt: info.ModTime().Format("2006-01-02 15:04:05"),
					Ext:       strings.ToLower(strings.TrimPrefix(filepath.Ext(name), ".")),
				})
			}
			sort.SliceStable(out, func(i, j int) bool {
				return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
			})
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "empresa_id": empresaID, "items": out})
			return

		case "upload":
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			if err := r.ParseMultipartForm(32 << 20); err != nil {
				http.Error(w, "multipart invalido", http.StatusBadRequest)
				return
			}
			f, hdr, err := r.FormFile("file")
			if err != nil {
				http.Error(w, "file requerido", http.StatusBadRequest)
				return
			}
			defer f.Close()
			name, err := onlyOfficeSafeBaseName(hdr.Filename)
			if err != nil {
				http.Error(w, "nombre de archivo invalido", http.StatusBadRequest)
				return
			}
			full, err := onlyOfficeJoinEmpresaFile(empresaID, name)
			if err != nil {
				http.Error(w, "ruta invalida", http.StatusBadRequest)
				return
			}
			tmp := full + ".uploading"
			out, err := os.Create(tmp)
			if err != nil {
				http.Error(w, "no se pudo escribir el archivo", http.StatusInternalServerError)
				return
			}
			_, copyErr := io.Copy(out, f)
			_ = out.Close()
			if copyErr != nil {
				_ = os.Remove(tmp)
				http.Error(w, "no se pudo escribir el archivo", http.StatusInternalServerError)
				return
			}
			_ = os.Rename(tmp, full)
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "empresa_id": empresaID, "name": name})
			return

		case "create":
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			var req onlyOfficeCreateDocRequest
			body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
			if len(bytes.TrimSpace(body)) > 0 {
				_ = json.Unmarshal(body, &req)
			}
			ext := onlyOfficeNormalizeCreateTipo(req.Tipo)
			if ext == "" {
				http.Error(w, "tipo invalido (word/excel/powerpoint)", http.StatusBadRequest)
				return
			}
			name := strings.TrimSpace(req.Nombre)
			if name == "" {
				name = onlyOfficeDefaultNameByExt(ext)
			}
			if filepath.Ext(name) == "" {
				name = name + "." + ext
			}
			// Forzar extensión correcta
			if strings.ToLower(strings.TrimPrefix(filepath.Ext(name), ".")) != ext {
				name = strings.TrimSuffix(name, filepath.Ext(name)) + "." + ext
			}
			name, err = onlyOfficeEnsureUniqueName(empresaID, name)
			if err != nil {
				http.Error(w, "nombre de archivo invalido", http.StatusBadRequest)
				return
			}
			full, err := onlyOfficeJoinEmpresaFile(empresaID, name)
			if err != nil {
				http.Error(w, "ruta invalida", http.StatusBadRequest)
				return
			}
			var fileBytes []byte
			switch ext {
			case "docx":
				fileBytes, err = onlyOfficeBuildEmptyDOCX()
			case "xlsx":
				fileBytes, err = onlyOfficeBuildEmptyXLSX()
			case "pptx":
				fileBytes, err = onlyOfficeBuildEmptyPPTX()
			default:
				err = fmt.Errorf("tipo no soportado")
			}
			if err != nil {
				http.Error(w, "no se pudo crear documento", http.StatusInternalServerError)
				return
			}
			tmp := full + ".new"
			if writeErr := os.WriteFile(tmp, fileBytes, 0640); writeErr != nil {
				http.Error(w, "no se pudo escribir el archivo", http.StatusInternalServerError)
				return
			}
			_ = os.Rename(tmp, full)
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "empresa_id": empresaID, "name": name})
			return

		case "editor_config":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			if dbSuper == nil {
				http.Error(w, "db_super no disponible", http.StatusInternalServerError)
				return
			}
			fileName := strings.TrimSpace(r.URL.Query().Get("file"))
			base, err := onlyOfficeSafeBaseName(fileName)
			if err != nil {
				http.Error(w, "file invalido", http.StatusBadRequest)
				return
			}
			full, err := onlyOfficeJoinEmpresaFile(empresaID, base)
			if err != nil {
				http.Error(w, "file invalido", http.StatusBadRequest)
				return
			}
			if _, err := os.Stat(full); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					http.Error(w, "archivo no encontrado", http.StatusNotFound)
					return
				}
				http.Error(w, "no se pudo validar archivo", http.StatusInternalServerError)
				return
			}

			dsURL, _, _ := onlyOfficeResolveDocumentServerURL(dbSuper)
			jwtSecret, _, _ := onlyOfficeResolveJWTSecret(dbSuper)
			dsURL = strings.TrimRight(strings.TrimSpace(dsURL), "/")
			if dsURL == "" {
				http.Error(w, "OnlyOffice Document Server no configurado (onlyoffice.document_server_url)", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(jwtSecret) == "" {
				http.Error(w, "OnlyOffice JWT no configurado (onlyoffice.jwt_secret)", http.StatusBadRequest)
				return
			}

			mode := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("mode")))
			if mode != "view" {
				mode = "edit"
			}
			canEdit := mode == "edit"

			// Token temporal para servir archivo y para callback.
			exp := time.Now().Add(15 * time.Minute).Unix()
			fileTok, _ := onlyOfficeSignToken(jwtSecret, onlyOfficeAccessTokenClaims{
				EmpresaID: empresaID,
				Path:      base,
				Action:    "file",
				ExpUnix:   exp,
			})
			cbTok, _ := onlyOfficeSignToken(jwtSecret, onlyOfficeAccessTokenClaims{
				EmpresaID: empresaID,
				Path:      base,
				Action:    "callback",
				ExpUnix:   exp,
			})

			baseURL := resolveBaseURLForConfirmation(r, dbSuper)
			documentURL := strings.TrimRight(baseURL, "/") + "/api/onlyoffice/file?token=" + urlQueryEscape(fileTok)
			callbackURL := strings.TrimRight(baseURL, "/") + "/api/onlyoffice/callback?token=" + urlQueryEscape(cbTok)

			// key debe cambiar cuando cambie el archivo; usamos mtime+size como base.
			st, _ := os.Stat(full)
			keySeed := fmt.Sprintf("%d|%d|%s|%d", empresaID, st.Size(), st.ModTime().UTC().Format(time.RFC3339Nano), st.ModTime().UnixNano())
			sum := sha256.Sum256([]byte(keySeed))
			docKey := fmt.Sprintf("pcs-%x", sum[:16])

			ooCfg := map[string]any{
				"documentType": onlyOfficeDocumentTypeByExt(base),
				"document": map[string]any{
					"fileType": onlyOfficeGuessFileType(base),
					"key":      docKey,
					"title":    base,
					"url":      documentURL,
					"permissions": map[string]any{
						"edit":     canEdit,
						"download": true,
						"print":    true,
						"review":   false,
						"comment":  canEdit,
					},
				},
				"editorConfig": map[string]any{
					"mode": mode,
					"callbackUrl": callbackURL,
					"user": map[string]any{
						"id":   strings.TrimSpace(adminEmailFromRequest(r)),
						"name": strings.TrimSpace(adminEmailFromRequest(r)),
					},
					"customization": map[string]any{
						"forcesave": true,
					},
				},
			}

			jwt, err := onlyOfficeJWTSignHS256(jwtSecret, ooCfg)
			if err != nil {
				http.Error(w, "no se pudo firmar jwt", http.StatusInternalServerError)
				return
			}
			ooCfg["token"] = jwt
			if doc, ok := ooCfg["document"].(map[string]any); ok {
				doc["token"] = jwt
			}
			if ed, ok := ooCfg["editorConfig"].(map[string]any); ok {
				ed["token"] = jwt
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"ok":            true,
				"empresa_id":    empresaID,
				"ds_url":        dsURL,
				"onlyofficeCfg": ooCfg,
			})
			return

		default:
			http.Error(w, "action invalida (list, upload, editor_config)", http.StatusBadRequest)
			return
		}
	}
}

func urlQueryEscape(v string) string {
	return url.QueryEscape(v)
}

// Endpoint público: sirve el archivo a OnlyOffice (token temporal).
func OnlyOfficeFilePublicHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if dbSuper != nil && !isOnlyOfficeEnabled(dbSuper) {
			http.Error(w, "OnlyOffice disabled", http.StatusServiceUnavailable)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if dbSuper == nil {
			http.Error(w, "db_super no disponible", http.StatusInternalServerError)
			return
		}
		jwtSecret, _, _ := onlyOfficeResolveJWTSecret(dbSuper)
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		claims, err := onlyOfficeVerifyToken(jwtSecret, token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if claims.Action != "file" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		full, err := onlyOfficeJoinEmpresaFile(claims.EmpresaID, claims.Path)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		f, err := os.Open(full)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		defer f.Close()
		w.Header().Set("Content-Type", onlyOfficeMIMEByExt(claims.Path))
		http.ServeContent(w, r, claims.Path, time.Now(), f)
	}
}

// Endpoint público: callback de OnlyOffice (guardado).
func OnlyOfficeCallbackPublicHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if dbSuper != nil && !isOnlyOfficeEnabled(dbSuper) {
			http.Error(w, "OnlyOffice disabled", http.StatusServiceUnavailable)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if dbSuper == nil {
			http.Error(w, "db_super no disponible", http.StatusInternalServerError)
			return
		}
		jwtSecret, _, _ := onlyOfficeResolveJWTSecret(dbSuper)
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		claims, err := onlyOfficeVerifyToken(jwtSecret, token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if claims.Action != "callback" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var payload onlyOfficeCallbackPayload
		raw, _ := io.ReadAll(io.LimitReader(r.Body, 2<<20))
		_ = json.Unmarshal(raw, &payload)

		// Respuesta esperada por OnlyOffice: {"error":0}
		resp := map[string]any{"error": 0}

		// status 2 = documento listo para guardarse; status 6 = forcesave.
		if payload.Status != 2 && payload.Status != 6 {
			writeJSON(w, http.StatusOK, resp)
			return
		}
		if strings.TrimSpace(payload.URL) == "" {
			writeJSON(w, http.StatusOK, resp)
			return
		}

		full, err := onlyOfficeJoinEmpresaFile(claims.EmpresaID, claims.Path)
		if err != nil {
			writeJSON(w, http.StatusOK, resp)
			return
		}

		// Descargar el archivo desde OnlyOffice y reemplazar el local (atomic).
		client := &http.Client{Timeout: 45 * time.Second}
		req, _ := http.NewRequest(http.MethodGet, strings.TrimSpace(payload.URL), nil)
		// Algunos deployments requieren JWT en header para descargas desde Document Server.
		if jwtSecret != "" {
			if dlJWT, err := onlyOfficeJWTSignHS256(jwtSecret, map[string]any{
				"payload": map[string]any{"url": strings.TrimSpace(payload.URL)},
			}); err == nil && strings.TrimSpace(dlJWT) != "" {
				req.Header.Set("Authorization", "Bearer "+dlJWT)
			}
		}

		res, err := client.Do(req)
		if err != nil || res == nil {
			writeJSON(w, http.StatusOK, resp)
			return
		}
		defer res.Body.Close()
		if res.StatusCode < 200 || res.StatusCode > 299 {
			writeJSON(w, http.StatusOK, resp)
			return
		}

		tmp := full + ".onlyoffice"
		out, err := os.Create(tmp)
		if err != nil {
			writeJSON(w, http.StatusOK, resp)
			return
		}
		_, copyErr := io.Copy(out, res.Body)
		_ = out.Close()
		if copyErr != nil {
			_ = os.Remove(tmp)
			writeJSON(w, http.StatusOK, resp)
			return
		}
		_ = os.Rename(tmp, full)

		writeJSON(w, http.StatusOK, resp)
	}
}

