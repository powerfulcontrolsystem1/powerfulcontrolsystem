package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type empresaDeleteFileCleanupResult struct {
	PathsEliminados []string `json:"paths_eliminados,omitempty"`
	Errores         []string `json:"errores,omitempty"`
}

func removeEmpresaOwnedDir(root, target string, result *empresaDeleteFileCleanupResult) {
	root = strings.TrimSpace(root)
	target = strings.TrimSpace(target)
	if root == "" || target == "" || result == nil {
		return
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		result.Errores = append(result.Errores, fmt.Sprintf("%s: %v", root, err))
		return
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		result.Errores = append(result.Errores, fmt.Sprintf("%s: %v", target, err))
		return
	}
	rel, err := filepath.Rel(absRoot, absTarget)
	if err != nil || rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." || filepath.IsAbs(rel) {
		result.Errores = append(result.Errores, fmt.Sprintf("%s: ruta fuera de alcance seguro", absTarget))
		return
	}
	if _, err := os.Stat(absTarget); err != nil {
		if os.IsNotExist(err) {
			return
		}
		result.Errores = append(result.Errores, fmt.Sprintf("%s: %v", absTarget, err))
		return
	}
	if err := os.RemoveAll(absTarget); err != nil {
		result.Errores = append(result.Errores, fmt.Sprintf("%s: %v", absTarget, err))
		return
	}
	result.PathsEliminados = append(result.PathsEliminados, absTarget)
}

func removeEmpresaOwnedFile(root, target string, result *empresaDeleteFileCleanupResult) {
	root = strings.TrimSpace(root)
	target = strings.TrimSpace(target)
	if root == "" || target == "" || result == nil {
		return
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		result.Errores = append(result.Errores, fmt.Sprintf("%s: %v", root, err))
		return
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		result.Errores = append(result.Errores, fmt.Sprintf("%s: %v", target, err))
		return
	}
	rel, err := filepath.Rel(absRoot, absTarget)
	if err != nil || rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." || filepath.IsAbs(rel) {
		result.Errores = append(result.Errores, fmt.Sprintf("%s: ruta fuera de alcance seguro", absTarget))
		return
	}
	if _, err := os.Stat(absTarget); err != nil {
		if os.IsNotExist(err) {
			return
		}
		result.Errores = append(result.Errores, fmt.Sprintf("%s: %v", absTarget, err))
		return
	}
	if err := os.Remove(absTarget); err != nil {
		result.Errores = append(result.Errores, fmt.Sprintf("%s: %v", absTarget, err))
		return
	}
	result.PathsEliminados = append(result.PathsEliminados, absTarget)
}

func cleanupEmpresaDynamicDocuments(empresaID int64, result *empresaDeleteFileCleanupResult) {
	if empresaID <= 0 || result == nil {
		return
	}
	dir, err := ensureDynamicDocumentDir()
	if err != nil {
		result.Errores = append(result.Errores, fmt.Sprintf("documentos temporales: %v", err))
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		result.Errores = append(result.Errores, fmt.Sprintf("%s: %v", dir, err))
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".record.json") {
			continue
		}
		recordPath := filepath.Join(dir, entry.Name())
		// #nosec G304 -- path is normalized and constrained to a server-controlled root before this operation.
		raw, err := os.ReadFile(recordPath)
		if err != nil {
			result.Errores = append(result.Errores, fmt.Sprintf("%s: %v", recordPath, err))
			continue
		}
		var record dynamicDocumentRecord
		if err := json.Unmarshal(raw, &record); err != nil || record.EmpresaID != empresaID || record.ID == "" {
			continue
		}
		matches, err := filepath.Glob(filepath.Join(dir, record.ID+".*"))
		if err != nil {
			result.Errores = append(result.Errores, fmt.Sprintf("%s: %v", record.ID, err))
			continue
		}
		for _, match := range matches {
			removeEmpresaOwnedFile(dir, match, result)
		}
	}
}

func cleanupEmpresaOwnedFiles(empresaID int64) empresaDeleteFileCleanupResult {
	result := empresaDeleteFileCleanupResult{}
	if empresaID <= 0 {
		result.Errores = append(result.Errores, "empresa_id invalido")
		return result
	}

	empresaDir := fmt.Sprintf("empresa_%d", empresaID)
	webRoot := resolveWebRootDir()
	uploadsRoot := filepath.Join(webRoot, "uploads")
	backupEmpresasRoot := filepath.Join(backupRootDir(), "empresas")

	for _, rel := range []string{
		filepath.Join("chat_tareas", empresaDir),
		filepath.Join("comprobantes", empresaDir),
		filepath.Join("dian", empresaDir),
		filepath.Join("productos", empresaDir),
		filepath.Join("red_social", empresaDir),
		filepath.Join("venta_publica", empresaDir),
	} {
		removeEmpresaOwnedDir(uploadsRoot, filepath.Join(uploadsRoot, rel), &result)
	}
	empresasUploadsRoot := filepath.Join(uploadsRoot, empresaUploadsRootDirName)
	for _, pattern := range []string{
		filepath.Join(empresasUploadsRoot, empresaDir),
		filepath.Join(empresasUploadsRoot, fmt.Sprintf("empresa_%d_*", empresaID)),
	} {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			result.Errores = append(result.Errores, fmt.Sprintf("%s: %v", pattern, err))
			continue
		}
		for _, match := range matches {
			removeEmpresaOwnedDir(empresasUploadsRoot, match, &result)
		}
	}

	removeEmpresaOwnedDir(backupEmpresasRoot, filepath.Join(backupEmpresasRoot, fmt.Sprintf("%d", empresaID)), &result)
	cleanupEmpresaDynamicDocuments(empresaID, &result)

	logFile := fmt.Sprintf("empresa_%d.log", empresaID)
	for _, logRoot := range []string{
		filepath.Join(".", "logs"),
		filepath.Join(resolveProjectRootDir(), "logs"),
		filepath.Join(resolveProjectRootDir(), "backend", "logs"),
	} {
		removeEmpresaOwnedFile(logRoot, filepath.Join(logRoot, logFile), &result)
	}
	return result
}
