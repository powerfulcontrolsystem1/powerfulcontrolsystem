package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var blockedPrivateExtensions = map[string]bool{
	".bat": true, ".cmd": true, ".com": true, ".dll": true, ".exe": true,
	".htm": true, ".html": true, ".js": true, ".mjs": true, ".ps1": true,
	".sh": true, ".svg": true,
}

var allowedPrivateExtensions = map[string]bool{
	".aac": true, ".csv": true, ".doc": true, ".docx": true, ".gif": true,
	".jpeg": true, ".jpg": true, ".json": true, ".m4a": true, ".mp3": true,
	".odt": true, ".ods": true, ".odp": true, ".ogg": true, ".pdf": true,
	".pem": true, ".png": true, ".ppt": true, ".pptx": true, ".rtf": true,
	".tif": true, ".tiff": true, ".txt": true,
	".wav": true, ".webm": true, ".webp": true, ".xls": true, ".xlsx": true,
	".xml": true,
}

func empresaPrivateCategoryRoot(empresaID int64, category string) (string, error) {
	category = strings.ToLower(strings.TrimSpace(category))
	switch category {
	case "buzon", "chat_tareas", "dian", "finanzas", "grafologia", "soportes_compras_ia":
	default:
		return "", errors.New("categoria privada no permitida")
	}
	if empresaID <= 0 {
		return "", errors.New("empresa invalida")
	}
	base := strings.TrimSpace(os.Getenv("PCS_PRIVATE_STORAGE_DIR"))
	if base == "" {
		base = filepath.Join(resolveProjectRootDir(), "private_storage")
	}
	return filepath.Join(base, category, "empresa_"+strconv.FormatInt(empresaID, 10)), nil
}

func saveEmpresaPrivateUpload(empresaID int64, category, extension string, source io.Reader, maxBytes int64) (string, string, int64, error) {
	extension = strings.ToLower(strings.TrimSpace(extension))
	if extension == "" || len(extension) > 12 || !privateCategoryAllowsExtension(category, extension) || blockedPrivateExtensions[extension] || strings.ContainsAny(extension, `/\\`) {
		return "", "", 0, errors.New("extension privada no permitida")
	}
	root, err := empresaPrivateCategoryRoot(empresaID, category)
	if err != nil {
		return "", "", 0, err
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return "", "", 0, err
	}
	nameBytes := make([]byte, 32)
	if _, err := rand.Read(nameBytes); err != nil {
		return "", "", 0, err
	}
	fileName := hex.EncodeToString(nameBytes) + extension
	path := filepath.Join(root, fileName)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600) // #nosec G304 -- random basename under validated private tenant root.
	if err != nil {
		return "", "", 0, err
	}
	if maxBytes <= 0 {
		maxBytes = 20 << 20
	}
	written, copyErr := io.Copy(file, io.LimitReader(source, maxBytes+1))
	closeErr := file.Close()
	if copyErr != nil || closeErr != nil || written > maxBytes {
		_ = os.Remove(path)
		if written > maxBytes {
			return "", "", 0, errors.New("archivo privado supera el limite permitido")
		}
		return "", "", 0, errors.New("no se pudo guardar archivo privado")
	}
	if err := validatePrivateFileContent(path, extension); err != nil {
		_ = os.Remove(path)
		return "", "", 0, err
	}
	return fileName, path, written, nil
}

func privateCategoryAllowsExtension(category, extension string) bool {
	if !allowedPrivateExtensions[extension] {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "dian":
		return extension == ".pem" || extension == ".png" || extension == ".jpg" || extension == ".jpeg" || extension == ".tif" || extension == ".tiff"
	case "grafologia":
		return extension == ".png" || extension == ".jpg" || extension == ".jpeg" || extension == ".gif" || extension == ".webp"
	case "finanzas":
		return empresaComprobanteAllowedExt[extension]
	case "buzon", "chat_tareas":
		return isAllowedAttachmentExt(extension) && extension != ".svg"
	case "soportes_compras_ia":
		return soporteComprasIAAllowedExt[extension]
	default:
		return false
	}
}

func validatePrivateFileContent(path, extension string) error {
	file, err := os.Open(path) // #nosec G304 -- caller supplies a path created under a private tenant root.
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	detected := strings.ToLower(http.DetectContentType(buf[:n]))
	prefix := strings.ToLower(strings.TrimSpace(string(buf[:n])))
	if strings.Contains(detected, "text/html") || strings.Contains(detected, "javascript") || strings.Contains(prefix, "<svg") || strings.HasPrefix(prefix, "#!") {
		return errors.New("contenido activo no permitido")
	}
	if extension == ".pem" && !strings.HasPrefix(prefix, "-----begin ") {
		return errors.New("contenido PEM invalido")
	}
	if !privateExtensionMatchesContent(strings.ToLower(extension), detected) {
		return errors.New("el contenido no coincide con la extension permitida")
	}
	return nil
}

func privateExtensionMatchesContent(extension, detected string) bool {
	switch extension {
	case ".png":
		return detected == "image/png"
	case ".jpg", ".jpeg":
		return detected == "image/jpeg"
	case ".gif":
		return detected == "image/gif"
	case ".webp":
		return detected == "image/webp"
	case ".pdf":
		return detected == "application/pdf"
	case ".tif", ".tiff":
		return detected == "image/tiff" || detected == "image/x-tiff"
	case ".pem":
		return strings.HasPrefix(detected, "text/plain")
	case ".txt", ".csv", ".json", ".xml", ".rtf":
		return strings.HasPrefix(detected, "text/") || detected == "application/json" || detected == "application/xml"
	case ".mp3", ".wav", ".ogg", ".m4a", ".webm", ".aac":
		return strings.HasPrefix(detected, "audio/") || strings.HasPrefix(detected, "video/webm") || detected == "application/ogg" || detected == "application/octet-stream"
	case ".doc", ".xls", ".ppt":
		return detected == "application/x-ole-storage" || detected == "application/octet-stream"
	case ".docx", ".xlsx", ".pptx", ".odt", ".ods", ".odp":
		return detected == "application/zip" || detected == "application/octet-stream"
	default:
		return false
	}
}

func empresaPrivateDownloadURL(route string, empresaID int64, ref string) string {
	return route + "?empresa_id=" + strconv.FormatInt(empresaID, 10) + "&ref=" + url.QueryEscape(ref)
}

func resolveEmpresaPrivateFile(empresaID int64, category, ref string) (string, error) {
	root, err := empresaPrivateCategoryRoot(empresaID, category)
	if err != nil {
		return "", err
	}
	ref = filepath.Clean(filepath.FromSlash(strings.TrimSpace(ref)))
	if ref == "." || ref == ".." || filepath.IsAbs(ref) || strings.HasPrefix(ref, ".."+string(os.PathSeparator)) {
		return "", errors.New("referencia privada invalida")
	}
	candidate := filepath.Join(root, ref)
	return resolveExistingPrivateFileUnderRoot(root, candidate)
}

func serveEmpresaPrivateFile(w http.ResponseWriter, r *http.Request, empresaID int64, category string) {
	path, err := resolveEmpresaPrivateFile(empresaID, category, r.URL.Query().Get("ref"))
	if err != nil {
		http.Error(w, "archivo no disponible", http.StatusNotFound)
		return
	}
	file, err := os.Open(path) // #nosec G304 -- path is resolved under the validated private tenant root without symlinks.
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
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(path)))
	http.ServeContent(w, r, filepath.Base(path), info.ModTime(), file)
}
