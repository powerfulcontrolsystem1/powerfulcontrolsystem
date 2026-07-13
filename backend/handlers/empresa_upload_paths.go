package handlers

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	empresaUploadsRootDirName              = "empresas"
	empresaFacturacionElectronicaDirName   = "facturacion_electronica"
	empresaFirmaElectronicaDirName         = "firma_electronica"
	empresaCapturasDIANDirName             = "capturas_dian"
	empresaFirmaElectronicaPrivateDirPerms = 0o700
)

func empresaUploadsFolderName(dbEmp *sql.DB, empresaID int64) string {
	name := ""
	if dbEmp != nil && empresaID > 0 {
		if empresa, err := dbpkg.GetEmpresaByScopeID(dbEmp, empresaID); err == nil && empresa != nil {
			name = empresa.Nombre
		}
	}
	slug := sanitizeDomoticaStorageSlug(name)
	if slug == "" {
		slug = "empresa"
	}
	return fmt.Sprintf("empresa_%d_%s", empresaID, slug)
}

func empresaUploadsBaseDir(dbEmp *sql.DB, empresaID int64) (string, string) {
	folder := empresaUploadsFolderName(dbEmp, empresaID)
	return filepath.Join(resolveWebRootDir(), "uploads", empresaUploadsRootDirName, folder), folder
}

func empresaUploadsSubdir(dbEmp *sql.DB, empresaID int64, parts ...string) (string, string, string) {
	baseDir, folder := empresaUploadsBaseDir(dbEmp, empresaID)
	cleanParts := make([]string, 0, len(parts))
	for _, part := range parts {
		clean := strings.Trim(strings.TrimSpace(part), `/\`)
		if clean == "" || clean == "." {
			continue
		}
		cleanParts = append(cleanParts, clean)
	}
	allParts := append([]string{baseDir}, cleanParts...)
	publicParts := append([]string{"/uploads", empresaUploadsRootDirName, folder}, cleanParts...)
	return filepath.Join(allParts...), strings.Join(publicParts, "/"), folder
}

func empresaFacturacionFirmaElectronicaDir(dbEmp *sql.DB, empresaID int64) (string, string) {
	folder := empresaUploadsFolderName(dbEmp, empresaID)
	dir, _ := empresaPrivateCategoryRoot(empresaID, "dian")
	return dir, folder
}

func ensureEmpresaUploadFolders(dbEmp *sql.DB, empresaID int64) (string, error) {
	baseDir, folder := empresaUploadsBaseDir(dbEmp, empresaID)
	publicImagesDir := filepath.Join(baseDir, "imagenes")
	dianPrivateDir, err := empresaPrivateCategoryRoot(empresaID, "dian")
	if err != nil {
		return "", err
	}
	for _, item := range []struct {
		path string
		perm os.FileMode
	}{
		{baseDir, 0o755},
		{publicImagesDir, 0o755},
		{dianPrivateDir, empresaFirmaElectronicaPrivateDirPerms},
	} {
		if err := os.MkdirAll(item.path, item.perm); err != nil {
			return folder, err
		}
		_ = os.Chmod(item.path, item.perm)
	}
	return folder, nil
}

func empresaUploadedPublicURLAbsPath(dbEmp *sql.DB, empresaID int64, publicURL string) (string, bool) {
	raw := strings.TrimSpace(publicURL)
	if raw == "" || empresaID <= 0 {
		return "", false
	}
	if strings.HasPrefix(strings.ToLower(raw), "http://") || strings.HasPrefix(strings.ToLower(raw), "https://") {
		parsed, err := url.Parse(raw)
		if err != nil {
			return "", false
		}
		raw = parsed.Path
	}
	if !strings.HasPrefix(raw, "/") {
		return "", false
	}
	baseDir, folder := empresaUploadsBaseDir(dbEmp, empresaID)
	cleanURL := path.Clean("/" + strings.TrimLeft(raw, "/"))
	prefix := "/uploads/" + empresaUploadsRootDirName + "/" + folder + "/"
	if !strings.HasPrefix(cleanURL, prefix) {
		return "", false
	}
	relURL := strings.TrimPrefix(cleanURL, prefix)
	if relURL == "" || strings.HasPrefix(relURL, "../") || strings.Contains(relURL, "/../") {
		return "", false
	}
	abs := filepath.Join(baseDir, filepath.FromSlash(relURL))
	baseClean, err := filepath.Abs(filepath.Clean(baseDir))
	if err != nil {
		return "", false
	}
	absClean, err := filepath.Abs(filepath.Clean(abs))
	if err != nil {
		return "", false
	}
	rel, err := filepath.Rel(baseClean, absClean)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", false
	}
	return absClean, true
}

func deleteEmpresaUploadedPublicURL(dbEmp *sql.DB, empresaID int64, publicURL string) bool {
	abs, ok := empresaUploadedPublicURLAbsPath(dbEmp, empresaID, publicURL)
	if !ok {
		return false
	}
	info, err := os.Stat(abs)
	if err != nil || info.IsDir() {
		return false
	}
	return os.Remove(abs) == nil
}

func legacyEmpresaProductoUploadAbsPath(empresaID int64, publicURL string) (string, bool) {
	raw := strings.TrimSpace(publicURL)
	if raw == "" || empresaID <= 0 {
		return "", false
	}
	if strings.HasPrefix(strings.ToLower(raw), "http://") || strings.HasPrefix(strings.ToLower(raw), "https://") {
		parsed, err := url.Parse(raw)
		if err != nil {
			return "", false
		}
		raw = parsed.Path
	}
	cleanURL := path.Clean("/" + strings.TrimLeft(raw, "/"))
	prefix := "/uploads/productos/empresa_" + fmt.Sprintf("%d", empresaID) + "/"
	if !strings.HasPrefix(cleanURL, prefix) {
		return "", false
	}
	relURL := strings.TrimPrefix(cleanURL, prefix)
	if relURL == "" || strings.HasPrefix(relURL, "../") || strings.Contains(relURL, "/../") {
		return "", false
	}
	baseDir := filepath.Join(resolveWebRootDir(), "uploads", "productos", fmt.Sprintf("empresa_%d", empresaID))
	abs := filepath.Join(baseDir, filepath.FromSlash(relURL))
	baseClean, err := filepath.Abs(filepath.Clean(baseDir))
	if err != nil {
		return "", false
	}
	absClean, err := filepath.Abs(filepath.Clean(abs))
	if err != nil {
		return "", false
	}
	rel, err := filepath.Rel(baseClean, absClean)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", false
	}
	return absClean, true
}

func deleteEmpresaProductoUploadedPublicURL(dbEmp *sql.DB, empresaID int64, publicURL string) bool {
	if deleteEmpresaUploadedPublicURL(dbEmp, empresaID, publicURL) {
		return true
	}
	abs, ok := legacyEmpresaProductoUploadAbsPath(empresaID, publicURL)
	if !ok {
		return false
	}
	info, err := os.Stat(abs)
	if err != nil || info.IsDir() {
		return false
	}
	return os.Remove(abs) == nil
}

func deleteFileRefIfInsideDir(fileRef string, baseDir string) bool {
	raw := strings.TrimSpace(fileRef)
	if raw == "" || strings.TrimSpace(baseDir) == "" {
		return false
	}
	if strings.HasPrefix(raw, "file:") {
		raw = strings.TrimPrefix(raw, "file:")
	}
	if raw == "" {
		return false
	}
	baseClean, err := filepath.Abs(filepath.Clean(baseDir))
	if err != nil {
		return false
	}
	absClean, err := filepath.Abs(filepath.Clean(raw))
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(baseClean, absClean)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return false
	}
	info, err := os.Stat(absClean)
	if err != nil || info.IsDir() {
		return false
	}
	return os.Remove(absClean) == nil
}

func deleteEmpresaDIANFileRef(dbEmp *sql.DB, empresaID int64, fileRef string) bool {
	privateDir, _ := empresaFacturacionFirmaElectronicaDir(dbEmp, empresaID)
	if deleteFileRefIfInsideDir(fileRef, privateDir) {
		return true
	}
	legacyDir, _, _ := empresaUploadsSubdir(
		dbEmp,
		empresaID,
		empresaFacturacionElectronicaDirName,
		empresaFirmaElectronicaDirName,
	)
	return deleteFileRefIfInsideDir(fileRef, legacyDir)
}
