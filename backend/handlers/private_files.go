package handlers

import (
	"crypto/rand"
	"database/sql"
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

type PrivateFilesMigrationResult struct {
	DryRun   bool  `json:"dry_run"`
	Scanned  int64 `json:"scanned"`
	Eligible int64 `json:"eligible"`
	Migrated int64 `json:"migrated"`
	Missing  int64 `json:"missing"`
	Rejected int64 `json:"rejected"`
}

type privateFilesMigrationSource struct {
	table    string
	column   string
	category string
	route    string
	fileRef  bool
}

var privateFilesMigrationSources = []privateFilesMigrationSource{
	{table: "chat_tareas_adjuntos", column: "file_url", category: "chat_tareas", route: "/api/empresa/chat_tareas/archivo"},
	{table: "chat_tareas", column: "nota_voz_url", category: "chat_tareas", route: "/api/empresa/chat_tareas/archivo"},
	{table: "empresa_buzon_adjuntos", column: "file_url", category: "buzon", route: "/api/empresa/buzon/archivo"},
	{table: "empresa_finanzas_movimientos", column: "comprobante_url", category: "finanzas", route: "/api/empresa/finanzas/archivo"},
	{table: "empresa_grafologia_analisis", column: "imagen_url", category: "grafologia", route: "/api/empresa/grafologia/archivo"},
	{table: "empresa_dian_configuracion", column: "certificado_clave_ref", category: "dian", fileRef: true},
	{table: "empresa_dian_configuracion", column: "certificado_url", category: "dian", fileRef: true},
}

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

// MigrateLegacyPrivateUploads moves legacy business attachments out of the web
// root. Dry-run is the default in the command wrapper and performs no writes.
func MigrateLegacyPrivateUploads(dbConn *sql.DB, webRoot string, apply bool) (PrivateFilesMigrationResult, error) {
	result := PrivateFilesMigrationResult{DryRun: !apply}
	if dbConn == nil {
		return result, errors.New("conexion empresarial no disponible")
	}
	absWebRoot, err := filepath.Abs(filepath.Clean(strings.TrimSpace(webRoot)))
	if err != nil || strings.TrimSpace(webRoot) == "" {
		return result, errors.New("raiz web invalida")
	}
	for _, source := range privateFilesMigrationSources {
		// #nosec G201 -- table and column names come only from the fixed migration catalog above.
		query := fmt.Sprintf("SELECT id, empresa_id, COALESCE(%s, '') FROM %s WHERE COALESCE(%s, '') <> ''", source.column, source.table, source.column)
		rows, err := dbConn.Query(query)
		if err != nil {
			return result, fmt.Errorf("no se pudo consultar inventario privado: %w", err)
		}
		for rows.Next() {
			var id, empresaID int64
			var oldRef string
			if err := rows.Scan(&id, &empresaID, &oldRef); err != nil {
				_ = rows.Close()
				return result, err
			}
			if !isLegacyPrivateReference(oldRef, source.fileRef) {
				continue
			}
			result.Scanned++
			legacyPath, err := resolveLegacyPrivatePath(absWebRoot, oldRef)
			if err != nil {
				result.Missing++
				continue
			}
			ext := strings.ToLower(filepath.Ext(legacyPath))
			if !privateCategoryAllowsExtension(source.category, ext) || validatePrivateFileContent(legacyPath, ext) != nil {
				result.Rejected++
				continue
			}
			result.Eligible++
			if !apply {
				continue
			}
			input, err := os.Open(legacyPath) // #nosec G304 -- resolved below the configured legacy web root without symlinks.
			if err != nil {
				result.Missing++
				continue
			}
			name, newPath, _, saveErr := saveEmpresaPrivateUpload(empresaID, source.category, ext, input, 20<<20)
			closeErr := input.Close()
			if saveErr != nil || closeErr != nil {
				result.Rejected++
				continue
			}
			newRef := empresaPrivateDownloadURL(source.route, empresaID, name)
			if source.fileRef {
				newRef = "file:" + newPath
			}
			// #nosec G201 -- table and column names come only from the fixed migration catalog above.
			update := fmt.Sprintf("UPDATE %s SET %s = $1 WHERE id = $2 AND empresa_id = $3 AND %s = $4", source.table, source.column, source.column)
			res, err := dbConn.Exec(update, newRef, id, empresaID, oldRef)
			if err != nil {
				_ = os.Remove(newPath)
				_ = rows.Close()
				return result, fmt.Errorf("no se pudo actualizar referencia privada: %w", err)
			}
			affected, err := res.RowsAffected()
			if err != nil || affected != 1 {
				_ = os.Remove(newPath)
				_ = rows.Close()
				return result, errors.New("la migracion privada no actualizo exactamente una fila")
			}
			if err := os.Remove(legacyPath); err != nil {
				return result, errors.New("referencia migrada pero el archivo heredado no pudo retirarse")
			}
			result.Migrated++
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			return result, err
		}
		if err := rows.Close(); err != nil {
			return result, err
		}
	}
	return result, nil
}

func isLegacyPrivateReference(ref string, fileRef bool) bool {
	ref = strings.TrimSpace(ref)
	if fileRef {
		return strings.HasPrefix(ref, "file:") && strings.Contains(filepath.ToSlash(ref), "/uploads/")
	}
	return strings.HasPrefix(filepath.ToSlash(ref), "/uploads/")
}

func resolveLegacyPrivatePath(webRoot, ref string) (string, error) {
	raw := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(ref), "file:"))
	if strings.HasPrefix(filepath.ToSlash(raw), "/uploads/") {
		raw = filepath.Join(webRoot, filepath.FromSlash(strings.TrimPrefix(filepath.ToSlash(raw), "/")))
	}
	return resolveExistingPrivateFileUnderRoot(webRoot, raw)
}
