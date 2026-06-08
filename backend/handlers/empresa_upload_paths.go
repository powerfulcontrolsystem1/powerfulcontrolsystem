package handlers

import (
	"database/sql"
	"fmt"
	"os"
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
	dir, _, folder := empresaUploadsSubdir(
		dbEmp,
		empresaID,
		empresaFacturacionElectronicaDirName,
		empresaFirmaElectronicaDirName,
	)
	return dir, folder
}

func ensureEmpresaUploadFolders(dbEmp *sql.DB, empresaID int64) (string, error) {
	baseDir, folder := empresaUploadsBaseDir(dbEmp, empresaID)
	publicImagesDir := filepath.Join(baseDir, "imagenes")
	facturacionDir := filepath.Join(baseDir, empresaFacturacionElectronicaDirName)
	firmaDir := filepath.Join(facturacionDir, empresaFirmaElectronicaDirName)
	capturasDIANDir := filepath.Join(facturacionDir, empresaCapturasDIANDirName)
	for _, item := range []struct {
		path string
		perm os.FileMode
	}{
		{baseDir, 0o755},
		{publicImagesDir, 0o755},
		{facturacionDir, empresaFirmaElectronicaPrivateDirPerms},
		{firmaDir, empresaFirmaElectronicaPrivateDirPerms},
		{capturasDIANDir, empresaFirmaElectronicaPrivateDirPerms},
	} {
		if err := os.MkdirAll(item.path, item.perm); err != nil {
			return folder, err
		}
		_ = os.Chmod(item.path, item.perm)
	}
	return folder, nil
}
