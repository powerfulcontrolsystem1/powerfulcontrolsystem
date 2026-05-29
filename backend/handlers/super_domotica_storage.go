package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

type superDomoticaStorageRow struct {
	EmpresaID    int64  `json:"empresa_id"`
	Nombre       string `json:"nombre"`
	Folder       string `json:"folder"`
	MaxImageKB   int64  `json:"max_image_kb"`
	UsadoBytes   int64  `json:"usado_bytes"`
	UsadoMB      string `json:"usado_mb"`
	Imagenes     int64  `json:"imagenes"`
	FolderExists bool   `json:"folder_exists"`
	PublicPath   string `json:"public_path"`
}

// SuperDomoticaStorageHandler administra limites y carpetas de imagenes empresariales por empresa.
func SuperDomoticaStorageHandler(dbSuper, dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.EqualFold(r.Method, http.MethodGet) && !strings.EqualFold(r.Method, http.MethodPut) && !strings.EqualFold(r.Method, http.MethodPost) {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		switch r.Method {
		case http.MethodGet:
			rows, err := buildSuperDomoticaStorageRows(dbSuper, dbEmp)
			if err != nil {
				http.Error(w, "No se pudo cargar almacenamiento de domotica", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":                   true,
				"default_max_image_kb": domoticaStorageDefaultMaxKB(dbSuper),
				"empresas":             rows,
			})
		case http.MethodPost, http.MethodPut:
			var payload struct {
				EmpresaID         int64 `json:"empresa_id"`
				DefaultMaxImageKB int64 `json:"default_max_image_kb"`
				MaxImageKB        int64 `json:"max_image_kb"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.DefaultMaxImageKB > 0 {
				if err := dbpkg.SetConfigValue(dbSuper, "domotica.storage.default_max_image_kb", strconv.FormatInt(normalizeDomoticaStorageKB(payload.DefaultMaxImageKB), 10), false); err != nil {
					http.Error(w, "No se pudo guardar limite general", http.StatusInternalServerError)
					return
				}
			}
			if payload.EmpresaID > 0 && payload.MaxImageKB > 0 {
				key := fmt.Sprintf("domotica.storage.empresa.%d.max_image_kb", payload.EmpresaID)
				if err := dbpkg.SetConfigValue(dbSuper, key, strconv.FormatInt(normalizeDomoticaStorageKB(payload.MaxImageKB), 10), false); err != nil {
					http.Error(w, "No se pudo guardar limite de empresa", http.StatusInternalServerError)
					return
				}
			}
			rows, _ := buildSuperDomoticaStorageRows(dbSuper, dbEmp)
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "empresas": rows, "default_max_image_kb": domoticaStorageDefaultMaxKB(dbSuper)})
		}
	}
}

func buildSuperDomoticaStorageRows(dbSuper, dbEmp *sql.DB) ([]superDomoticaStorageRow, error) {
	empresas, err := dbpkg.GetEmpresas(dbEmp)
	if err != nil {
		return nil, err
	}
	rows := make([]superDomoticaStorageRow, 0, len(empresas))
	for _, empresa := range empresas {
		empresaID := empresa.EmpresaID
		if empresaID <= 0 {
			empresaID = empresa.ID
		}
		folder := domoticaEmpresaStorageFolder(dbEmp, empresaID)
		dir := filepath.Join(resolveWebRootDir(), "uploads", "empresas", folder, "imagenes")
		used, images, exists := domoticaStorageDirUsage(dir)
		rows = append(rows, superDomoticaStorageRow{
			EmpresaID:    empresaID,
			Nombre:       empresa.Nombre,
			Folder:       folder,
			MaxImageKB:   domoticaStorageEmpresaMaxKB(dbSuper, empresaID),
			UsadoBytes:   used,
			UsadoMB:      fmt.Sprintf("%.2f", float64(used)/(1024*1024)),
			Imagenes:     images,
			FolderExists: exists,
			PublicPath:   "/uploads/empresas/" + folder + "/imagenes/",
		})
	}
	return rows, nil
}

func domoticaStorageDefaultMaxKB(dbSuper *sql.DB) int64 {
	if dbSuper != nil {
		if raw, _, err := dbpkg.GetConfigValue(dbSuper, "domotica.storage.default_max_image_kb"); err == nil {
			if kb, parseErr := strconv.ParseInt(strings.TrimSpace(raw), 10, 64); parseErr == nil {
				return normalizeDomoticaStorageKB(kb)
			}
		}
	}
	return 2048
}

func domoticaStorageEmpresaMaxKB(dbSuper *sql.DB, empresaID int64) int64 {
	if dbSuper != nil && empresaID > 0 {
		key := fmt.Sprintf("domotica.storage.empresa.%d.max_image_kb", empresaID)
		if raw, _, err := dbpkg.GetConfigValue(dbSuper, key); err == nil {
			if kb, parseErr := strconv.ParseInt(strings.TrimSpace(raw), 10, 64); parseErr == nil && kb > 0 {
				return normalizeDomoticaStorageKB(kb)
			}
		}
	}
	return domoticaStorageDefaultMaxKB(dbSuper)
}

func normalizeDomoticaStorageKB(kb int64) int64 {
	if kb < 128 {
		return 128
	}
	if kb > 20480 {
		return 20480
	}
	return kb
}

func domoticaStorageDirUsage(dir string) (int64, int64, bool) {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return 0, 0, false
	}
	var used int64
	var images int64
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return nil
		}
		if info, statErr := d.Info(); statErr == nil {
			used += info.Size()
			images++
		}
		return nil
	})
	return used, images, true
}
