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
	onlyOfficeDefaultDataRoot        = "/data/empresas"
	onlyOfficeConfigKeyDSURL         = "onlyoffice.document_server_url" // ej: http://onlyoffice:80
	onlyOfficeConfigKeyJWT           = "onlyoffice.jwt_secret"          // secreto HS256
	onlyOfficeFileTokenTTL           = 4 * time.Hour
	onlyOfficeCallbackTTL            = 24 * time.Hour
	onlyOfficeCallbackMaxBytes int64 = 32 << 20
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
	EmpresaID int64  `json:"empresa_id"`
	FileName  string `json:"file_name"`
	Mode      string `json:"mode"` // edit|view
	UserID    string `json:"user_id,omitempty"`
	UserName  string `json:"user_name,omitempty"`
	Download  bool   `json:"download,omitempty"`
}

// OnlyOffice callback payload (subset usado).
type onlyOfficeCallbackPayload struct {
	Status int    `json:"status"`
	URL    string `json:"url"`
	Key    string `json:"key"`
	Error  int    `json:"error"`
}

type onlyOfficeCreateDocRequest struct {
	Tipo         string `json:"tipo"`             // word|excel|powerpoint|docx|xlsx|pptx
	Nombre       string `json:"nombre,omitempty"` // opcional, sin path
	LocalSession bool   `json:"local_session,omitempty"`
}

func onlyOfficeBuildEmptyFileByExt(ext string) ([]byte, error) {
	switch strings.ToLower(strings.TrimSpace(ext)) {
	case "docx":
		return onlyOfficeBuildEmptyDOCX()
	case "xlsx":
		return onlyOfficeBuildEmptyXLSX()
	case "pptx":
		return onlyOfficeBuildBlankPPTX()
	default:
		return nil, fmt.Errorf("tipo no soportado")
	}
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
	root := filepath.Clean(onlyOfficeDataRoot())
	if !strings.EqualFold(filepath.Base(root), "empresas") {
		root = filepath.Join(root, "empresas")
	}
	dir := filepath.Join(root, fmt.Sprintf("%d", empresaID), "documentos")
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
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = filepath.Base(name)
	name = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		switch r {
		case ':', '*', '?', '"', '<', '>', '|':
			return '_'
		default:
			return r
		}
	}, name)
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

func onlyOfficeAttachConfigToken(secret string, cfg map[string]any) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("config onlyoffice invalida")
	}
	if doc, ok := cfg["document"].(map[string]any); ok {
		delete(doc, "token")
	}
	if ed, ok := cfg["editorConfig"].(map[string]any); ok {
		delete(ed, "token")
	}
	jwt, err := onlyOfficeJWTSignHS256(secret, cfg)
	if err != nil {
		return "", err
	}
	cfg["token"] = jwt
	return jwt, nil
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

func onlyOfficeIsInternalDocumentServerHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(strings.Trim(host, "[]")))
	if host == "" || host == "localhost" || strings.Contains(host, ".") || strings.Contains(host, ":") {
		return false
	}
	return host == "onlyoffice" ||
		host == "documentserver" ||
		host == "onlyoffice-documentserver" ||
		host == "pcs-onlyoffice-documentserver" ||
		strings.Contains(host, "onlyoffice")
}

func onlyOfficePublicDocumentServerURLFromBase(baseURL string) (string, bool) {
	u, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil || u == nil {
		return "", false
	}
	scheme := strings.ToLower(strings.TrimSpace(u.Scheme))
	if scheme != "http" && scheme != "https" {
		return "", false
	}
	host := strings.ToLower(strings.TrimSpace(u.Hostname()))
	if host == "" || host == "localhost" || strings.Contains(host, ":") {
		return "", false
	}
	if strings.HasPrefix(host, "127.") || strings.HasPrefix(host, "10.") || strings.HasPrefix(host, "192.168.") || strings.HasPrefix(host, "172.") {
		return "", false
	}
	host = strings.TrimPrefix(host, "www.")
	labels := strings.Split(host, ".")
	if len(labels) < 2 {
		return "", false
	}
	if len(labels) > 2 {
		switch labels[0] {
		case "app", "admin", "erp", "panel", "www":
			host = strings.Join(labels[1:], ".")
		}
	}
	if strings.HasPrefix(host, "onlyoffice.") {
		return scheme + "://" + host, true
	}
	return scheme + "://onlyoffice." + host, true
}

func onlyOfficeBrowserDocumentServerURL(r *http.Request, dbSuper *sql.DB, configured string) (string, bool) {
	configured = strings.TrimRight(strings.TrimSpace(configured), "/")
	for _, key := range []string{"ONLYOFFICE_PUBLIC_DOCUMENT_SERVER_URL", "ONLYOFFICE_BROWSER_DOCUMENT_SERVER_URL"} {
		if v := strings.TrimRight(strings.TrimSpace(os.Getenv(key)), "/"); v != "" {
			return v, v != configured
		}
	}
	u, err := url.Parse(configured)
	if err != nil || u == nil || !onlyOfficeIsInternalDocumentServerHost(u.Hostname()) {
		return configured, false
	}
	publicURL, ok := onlyOfficePublicDocumentServerURLFromBase(resolveBaseURLForConfirmation(r, dbSuper))
	if !ok {
		return configured, false
	}
	return publicURL, publicURL != configured
}

func onlyOfficeCallbackDownloadURLAllowed(dbSuper *sql.DB, raw string) bool {
	target, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || target == nil || (target.Scheme != "http" && target.Scheme != "https") || target.Hostname() == "" {
		return false
	}
	configured, _, err := onlyOfficeResolveDocumentServerURL(dbSuper)
	if err != nil {
		return false
	}
	server, err := url.Parse(strings.TrimSpace(configured))
	if err != nil || server == nil {
		return false
	}
	return strings.EqualFold(target.Scheme, server.Scheme) && strings.EqualFold(target.Hostname(), server.Hostname()) && target.Port() == server.Port()
}

// copyOnlyOfficeCallbackFile enforces the callback size limit without ever
// accepting a truncated document. The caller keeps the destination temporary
// until this function completes successfully.
func copyOnlyOfficeCallbackFile(dst io.Writer, src io.Reader) error {
	written, err := io.Copy(dst, io.LimitReader(src, onlyOfficeCallbackMaxBytes+1))
	if err != nil {
		return err
	}
	if written > onlyOfficeCallbackMaxBytes {
		return fmt.Errorf("onlyoffice callback document exceeds allowed size")
	}
	return nil
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

func onlyOfficeBuildBlankPPTX() ([]byte, error) {
	// PresentationML minimo: OnlyOffice requiere master, layout, theme y relaciones.
	return onlyOfficeZipBytes(map[string]string{
		"[Content_Types].xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/>
  <Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>
  <Override PartName="/ppt/presentation.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"/>
  <Override PartName="/ppt/presProps.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presProps+xml"/>
  <Override PartName="/ppt/slides/slide1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slide+xml"/>
  <Override PartName="/ppt/slideMasters/slideMaster1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideMaster+xml"/>
  <Override PartName="/ppt/slideLayouts/slideLayout1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideLayout+xml"/>
  <Override PartName="/ppt/theme/theme1.xml" ContentType="application/vnd.openxmlformats-officedocument.theme+xml"/>
</Types>`,
		"_rels/.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="ppt/presentation.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>
  <Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/>
</Relationships>`,
		"docProps/app.xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties" xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes">
  <Application>Powerful Control System</Application>
  <PresentationFormat>On-screen Show (4:3)</PresentationFormat>
  <Slides>1</Slides>
  <Company>Powerful Control System</Company>
  <AppVersion>1.0</AppVersion>
</Properties>`,
		"docProps/core.xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <dc:creator>Powerful Control System</dc:creator>
  <cp:lastModifiedBy>Powerful Control System</cp:lastModifiedBy>
  <dcterms:created xsi:type="dcterms:W3CDTF">2026-01-01T00:00:00Z</dcterms:created>
  <dcterms:modified xsi:type="dcterms:W3CDTF">2026-01-01T00:00:00Z</dcterms:modified>
</cp:coreProperties>`,
		"ppt/presentation.xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:presentation xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"
 xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <p:sldMasterIdLst>
    <p:sldMasterId id="2147483648" r:id="rId1"/>
  </p:sldMasterIdLst>
  <p:sldIdLst>
    <p:sldId id="256" r:id="rId2"/>
  </p:sldIdLst>
  <p:sldSz cx="9144000" cy="6858000" type="screen4x3"/>
  <p:notesSz cx="6858000" cy="9144000"/>
  <p:defaultTextStyle/>
</p:presentation>`,
		"ppt/_rels/presentation.xml.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster" Target="slideMasters/slideMaster1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide1.xml"/>
  <Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/presProps" Target="presProps.xml"/>
  <Relationship Id="rId4" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme" Target="theme/theme1.xml"/>
</Relationships>`,
		"ppt/presProps.xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:presentationPr xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"/>`,
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
      <p:grpSpPr><a:xfrm/></p:grpSpPr>
      <p:sp>
        <p:nvSpPr>
          <p:cNvPr id="2" name="Title 1"/>
          <p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr>
          <p:nvPr><p:ph/></p:nvPr>
        </p:nvSpPr>
        <p:spPr/>
        <p:txBody>
          <a:bodyPr/>
          <a:lstStyle/>
          <a:p><a:endParaRPr lang="es-CO"/></a:p>
        </p:txBody>
      </p:sp>
    </p:spTree>
  </p:cSld>
  <p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr>
</p:sld>`,
		"ppt/slides/_rels/slide1.xml.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout" Target="../slideLayouts/slideLayout1.xml"/>
</Relationships>`,
		"ppt/slideLayouts/slideLayout1.xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sldLayout xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"
 xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" type="title" preserve="1">
  <p:cSld name="Title Slide">
    <p:spTree>
      <p:nvGrpSpPr>
        <p:cNvPr id="1" name=""/>
        <p:cNvGrpSpPr/>
        <p:nvPr/>
      </p:nvGrpSpPr>
      <p:grpSpPr><a:xfrm/></p:grpSpPr>
      <p:sp>
        <p:nvSpPr>
          <p:cNvPr id="2" name=""/>
          <p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr>
          <p:nvPr><p:ph/></p:nvPr>
        </p:nvSpPr>
        <p:spPr/>
        <p:txBody>
          <a:bodyPr/>
          <a:lstStyle/>
          <a:p><a:endParaRPr lang="es-CO"/></a:p>
        </p:txBody>
      </p:sp>
    </p:spTree>
  </p:cSld>
  <p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr>
</p:sldLayout>`,
		"ppt/slideLayouts/_rels/slideLayout1.xml.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster" Target="../slideMasters/slideMaster1.xml"/>
</Relationships>`,
		"ppt/slideMasters/slideMaster1.xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sldMaster xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"
 xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"
 xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <p:cSld>
    <p:spTree>
      <p:nvGrpSpPr>
        <p:cNvPr id="1" name=""/>
        <p:cNvGrpSpPr/>
        <p:nvPr/>
      </p:nvGrpSpPr>
      <p:grpSpPr><a:xfrm/></p:grpSpPr>
      <p:sp>
        <p:nvSpPr>
          <p:cNvPr id="2" name="Title Placeholder 1"/>
          <p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr>
          <p:nvPr><p:ph type="title"/></p:nvPr>
        </p:nvSpPr>
        <p:spPr/>
        <p:txBody>
          <a:bodyPr/>
          <a:lstStyle/>
          <a:p/>
        </p:txBody>
      </p:sp>
    </p:spTree>
  </p:cSld>
  <p:clrMap bg1="lt1" tx1="dk1" bg2="lt2" tx2="dk2" accent1="accent1" accent2="accent2" accent3="accent3" accent4="accent4" accent5="accent5" accent6="accent6" hlink="hlink" folHlink="folHlink"/>
  <p:sldLayoutIdLst>
    <p:sldLayoutId id="2147483649" r:id="rId1"/>
  </p:sldLayoutIdLst>
  <p:txStyles>
    <p:titleStyle/>
    <p:bodyStyle/>
    <p:otherStyle/>
  </p:txStyles>
</p:sldMaster>`,
		"ppt/slideMasters/_rels/slideMaster1.xml.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout" Target="../slideLayouts/slideLayout1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme" Target="../theme/theme1.xml"/>
</Relationships>`,
		"ppt/theme/theme1.xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" name="Office Theme">
  <a:themeElements>
    <a:clrScheme name="Office">
      <a:dk1><a:sysClr val="windowText" lastClr="000000"/></a:dk1>
      <a:lt1><a:sysClr val="window" lastClr="FFFFFF"/></a:lt1>
      <a:dk2><a:srgbClr val="1F497D"/></a:dk2>
      <a:lt2><a:srgbClr val="EEECE1"/></a:lt2>
      <a:accent1><a:srgbClr val="4F81BD"/></a:accent1>
      <a:accent2><a:srgbClr val="C0504D"/></a:accent2>
      <a:accent3><a:srgbClr val="9BBB59"/></a:accent3>
      <a:accent4><a:srgbClr val="8064A2"/></a:accent4>
      <a:accent5><a:srgbClr val="4BACC6"/></a:accent5>
      <a:accent6><a:srgbClr val="F79646"/></a:accent6>
      <a:hlink><a:srgbClr val="0000FF"/></a:hlink>
      <a:folHlink><a:srgbClr val="800080"/></a:folHlink>
    </a:clrScheme>
    <a:fontScheme name="Office">
      <a:majorFont><a:latin typeface="Calibri"/><a:ea typeface=""/><a:cs typeface=""/></a:majorFont>
      <a:minorFont><a:latin typeface="Calibri"/><a:ea typeface=""/><a:cs typeface=""/></a:minorFont>
    </a:fontScheme>
    <a:fmtScheme name="Office">
      <a:fillStyleLst>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
        <a:gradFill rotWithShape="1"><a:gsLst><a:gs pos="0"><a:schemeClr val="phClr"><a:tint val="50000"/><a:satMod val="300000"/></a:schemeClr></a:gs><a:gs pos="100000"><a:schemeClr val="phClr"><a:tint val="15000"/><a:satMod val="350000"/></a:schemeClr></a:gs></a:gsLst><a:lin ang="16200000" scaled="1"/></a:gradFill>
        <a:noFill/>
      </a:fillStyleLst>
      <a:lnStyleLst>
        <a:ln w="9525" cap="flat" cmpd="sng" algn="ctr"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:prstDash val="solid"/></a:ln>
        <a:ln w="25400" cap="flat" cmpd="sng" algn="ctr"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:prstDash val="solid"/></a:ln>
        <a:ln w="38100" cap="flat" cmpd="sng" algn="ctr"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:prstDash val="solid"/></a:ln>
      </a:lnStyleLst>
      <a:effectStyleLst>
        <a:effectStyle><a:effectLst/></a:effectStyle>
        <a:effectStyle><a:effectLst/></a:effectStyle>
        <a:effectStyle><a:effectLst/></a:effectStyle>
      </a:effectStyleLst>
      <a:bgFillStyleLst>
        <a:solidFill><a:schemeClr val="phClr"/></a:solidFill>
        <a:gradFill rotWithShape="1"><a:gsLst><a:gs pos="0"><a:schemeClr val="phClr"><a:tint val="50000"/><a:satMod val="300000"/></a:schemeClr></a:gs><a:gs pos="100000"><a:schemeClr val="phClr"><a:tint val="50000"/><a:satMod val="300000"/></a:schemeClr></a:gs></a:gsLst><a:lin ang="16200000" scaled="1"/></a:gradFill>
        <a:noFill/>
      </a:bgFillStyleLst>
    </a:fmtScheme>
  </a:themeElements>
  <a:objectDefaults/>
  <a:extraClrSchemeLst/>
</a:theme>`,
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
			writeJSON(w, http.StatusOK, map[string]any{
				"ok":      false,
				"enabled": false,
				"error":   "OnlyOffice está desactivado por super administrador.",
			})
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
			// #nosec G304 -- path is normalized and constrained to a server-controlled root before this operation.
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

		case "create", "create_local", "create_edit_local":
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
			if action == "create_local" {
				name, err = onlyOfficeSafeBaseName(name)
				if err != nil {
					http.Error(w, "nombre de archivo invalido", http.StatusBadRequest)
					return
				}
				fileBytes, err := onlyOfficeBuildEmptyFileByExt(ext)
				if err != nil {
					http.Error(w, "no se pudo crear documento", http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", onlyOfficeMIMEByExt(name))
				w.Header().Set("Content-Disposition", "attachment; filename=\""+name+"\"")
				w.Header().Set("X-PCS-Storage", "cliente")
				_, _ = w.Write(fileBytes)
				return
			}
			localSession := action == "create_edit_local" || req.LocalSession
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
			fileBytes, err := onlyOfficeBuildEmptyFileByExt(ext)
			if err != nil {
				http.Error(w, "no se pudo crear documento", http.StatusInternalServerError)
				return
			}
			tmp := full + ".new"
			if writeErr := os.WriteFile(tmp, fileBytes, 0600); writeErr != nil {
				http.Error(w, "no se pudo escribir el archivo", http.StatusInternalServerError)
				return
			}
			_ = os.Rename(tmp, full)
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "empresa_id": empresaID, "name": name, "local_session": localSession})
			return

		case "download":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
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
			deleteAfter := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("delete")), "1") || strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("delete")), "true")
			w.Header().Set("Content-Type", onlyOfficeMIMEByExt(base))
			w.Header().Set("Content-Disposition", "attachment; filename=\""+base+"\"")
			w.Header().Set("X-PCS-Storage", "cliente")
			if deleteAfter {
				// #nosec G304 -- path is normalized and constrained to a server-controlled root before this operation.
				fileBytes, err := os.ReadFile(full)
				if err != nil {
					if errors.Is(err, os.ErrNotExist) {
						http.Error(w, "archivo no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, "no se pudo leer archivo", http.StatusInternalServerError)
					return
				}
				_, _ = w.Write(fileBytes)
				_ = os.Remove(full)
				return
			}
			// #nosec G304 -- path is normalized and constrained to a server-controlled root before this operation.
			f, err := os.Open(full)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					http.Error(w, "archivo no encontrado", http.StatusNotFound)
					return
				}
				http.Error(w, "no se pudo leer archivo", http.StatusInternalServerError)
				return
			}
			defer f.Close()
			info, _ := f.Stat()
			modTime := time.Now()
			if info != nil {
				modTime = info.ModTime()
			}
			http.ServeContent(w, r, base, modTime, f)
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
			// Asegurar que exista un JWT secret persistido (si hay cifrado disponible).
			// Esto evita que el editor falle con "token de seguridad no está configurado" cuando el super aún no lo ha guardado.
			_, _ = onlyOfficeEnsureJWTSecret(dbSuper)
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
				writeJSON(w, http.StatusOK, map[string]any{
					"ok":         false,
					"enabled":    true,
					"configured": false,
					"error":      "OnlyOffice Document Server no configurado (onlyoffice.document_server_url)",
				})
				return
			}
			browserDSURL, dsURLRewritten := onlyOfficeBrowserDocumentServerURL(r, dbSuper, dsURL)
			if strings.TrimSpace(jwtSecret) == "" {
				writeJSON(w, http.StatusOK, map[string]any{
					"ok":         false,
					"enabled":    true,
					"configured": false,
					"error":      "OnlyOffice JWT no configurado (onlyoffice.jwt_secret)",
				})
				return
			}

			mode := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("mode")))
			if mode != "view" {
				mode = "edit"
			}
			canEdit := mode == "edit"

			// Token temporal para servir archivo y para callback.
			now := time.Now()
			fileExp := now.Add(onlyOfficeFileTokenTTL).Unix()
			callbackExp := now.Add(onlyOfficeCallbackTTL).Unix()
			fileTok, _ := onlyOfficeSignToken(jwtSecret, onlyOfficeAccessTokenClaims{
				EmpresaID: empresaID,
				Path:      base,
				Action:    "file",
				ExpUnix:   fileExp,
			})
			cbTok, _ := onlyOfficeSignToken(jwtSecret, onlyOfficeAccessTokenClaims{
				EmpresaID: empresaID,
				Path:      base,
				Action:    "callback",
				ExpUnix:   callbackExp,
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
					"mode":        mode,
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

			if _, signErr := onlyOfficeAttachConfigToken(jwtSecret, ooCfg); signErr != nil {
				http.Error(w, "no se pudo firmar jwt", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"ok":               true,
				"empresa_id":       empresaID,
				"ds_url":           browserDSURL,
				"ds_url_rewritten": dsURLRewritten,
				"onlyofficeCfg":    ooCfg,
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
			writeJSON(w, http.StatusOK, map[string]any{"ok": false, "enabled": false, "error": "OnlyOffice disabled"})
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
		if strings.TrimSpace(jwtSecret) == "" {
			writeJSON(w, http.StatusOK, map[string]any{"ok": false, "enabled": true, "configured": false, "error": "OnlyOffice JWT no configurado (onlyoffice.jwt_secret)"})
			return
		}
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		claims, err := onlyOfficeVerifyToken(jwtSecret, token)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": "unauthorized"})
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
		// #nosec G304 -- path is normalized and constrained to a server-controlled root before this operation.
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
			writeJSON(w, http.StatusOK, map[string]any{"ok": false, "enabled": false, "error": "OnlyOffice disabled"})
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
		if strings.TrimSpace(jwtSecret) == "" {
			writeJSON(w, http.StatusOK, map[string]any{"ok": false, "enabled": true, "configured": false, "error": "OnlyOffice JWT no configurado (onlyoffice.jwt_secret)"})
			return
		}
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		claims, err := onlyOfficeVerifyToken(jwtSecret, token)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]any{"ok": false, "error": "unauthorized"})
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
		if !onlyOfficeCallbackDownloadURLAllowed(dbSuper, payload.URL) {
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
		// #nosec G304 -- path is normalized and constrained to a server-controlled root before this operation.
		out, err := os.Create(tmp)
		if err != nil {
			writeJSON(w, http.StatusOK, resp)
			return
		}
		copyErr := copyOnlyOfficeCallbackFile(out, res.Body)
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
