package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func resolveProjectRootDir() string {
	wd, err := os.Getwd()
	if err != nil || strings.TrimSpace(wd) == "" {
		return "."
	}
	base := strings.ToLower(filepath.Base(wd))
	// Normalmente el backend corre con CWD=backend (por scripts). En ese caso,
	// el root del repo está un nivel arriba.
	if base == "backend" {
		return filepath.Dir(wd)
	}
	return wd
}

func backupRootDir() string {
	return filepath.Join(resolveProjectRootDir(), "backup")
}

func ensureDir(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("empty dir path")
	}
	return os.MkdirAll(path, 0o755)
}

func superAdminBackupDir() string {
	return filepath.Join(backupRootDir(), "super_administrador")
}

func empresaBackupDir(empresaID int64) string {
	return filepath.Join(backupRootDir(), "empresas", fmt.Sprintf("%d", empresaID))
}

func writeJSONBackupFile(dir string, fileName string, payload interface{}) (string, error) {
	if err := ensureDir(dir); err != nil {
		return "", err
	}
	name := strings.TrimSpace(fileName)
	if name == "" {
		name = "backup.json"
	}
	// Evitar paths raros.
	name = filepath.Base(name)
	fullPath := filepath.Join(dir, name)

	// Marshal con indent para lectura humana.
	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", err
	}
	body = append(body, '\n')
	if err := os.WriteFile(fullPath, body, 0o644); err != nil {
		return "", err
	}
	return fullPath, nil
}

func backupTimestampForFile() string {
	// FS-safe: 20060102_150405
	return time.Now().In(time.Local).Format("20060102_150405")
}

func runtimeOSLabel() string {
	if v := strings.TrimSpace(runtime.GOOS); v != "" {
		return v
	}
	return "unknown"
}

