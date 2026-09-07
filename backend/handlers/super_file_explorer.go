package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

type superFileExplorerRoot struct {
	Label string `json:"label"`
	Path  string `json:"path"`
}

type superFileExplorerEntry struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Type       string `json:"type"`
	Extension  string `json:"extension"`
	Size       int64  `json:"size"`
	SizePretty string `json:"size_pretty"`
	ModifiedAt string `json:"modified_at"`
	Mode       string `json:"mode"`
	IsDir      bool   `json:"is_dir"`
	IsSymlink  bool   `json:"is_symlink"`
	Hidden     bool   `json:"hidden"`
	Error      string `json:"error,omitempty"`
}

type superFileExplorerResponse struct {
	OK           bool                     `json:"ok"`
	GeneratedAt  string                   `json:"generated_at"`
	RuntimeOS    string                   `json:"runtime_os"`
	CurrentPath  string                   `json:"current_path"`
	ParentPath   string                   `json:"parent_path"`
	Roots        []superFileExplorerRoot  `json:"roots"`
	Entries      []superFileExplorerEntry `json:"entries"`
	TotalEntries int                      `json:"total_entries"`
	Error        string                   `json:"error,omitempty"`
}

func SuperFileExplorerHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper); !ok {
			return
		}
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"ok": false, "error": "method not allowed"})
			return
		}

		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "" {
			action = "list"
		}
		if action != "list" {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "accion no soportada"})
			return
		}

		roots := superFileExplorerRoots()
		targetPath := superFileExplorerResolvePath(r.URL.Query().Get("path"), roots)
		entries, err := superFileExplorerListEntries(targetPath)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, os.ErrPermission) {
				status = http.StatusForbidden
			} else if errors.Is(err, os.ErrNotExist) {
				status = http.StatusNotFound
			}
			writeJSON(w, status, superFileExplorerResponse{
				OK:          false,
				GeneratedAt: time.Now().UTC().Format(time.RFC3339),
				RuntimeOS:   runtime.GOOS,
				CurrentPath: targetPath,
				ParentPath:  superFileExplorerParentPath(targetPath),
				Roots:       roots,
				Error:       "ruta no disponible o sin permisos de lectura",
			})
			return
		}

		writeJSON(w, http.StatusOK, superFileExplorerResponse{
			OK:           true,
			GeneratedAt:  time.Now().UTC().Format(time.RFC3339),
			RuntimeOS:    runtime.GOOS,
			CurrentPath:  targetPath,
			ParentPath:   superFileExplorerParentPath(targetPath),
			Roots:        roots,
			Entries:      entries,
			TotalEntries: len(entries),
		})
	}
}

func superFileExplorerRoots() []superFileExplorerRoot {
	if runtime.GOOS != "windows" {
		return []superFileExplorerRoot{{Label: "/", Path: string(filepath.Separator)}}
	}

	roots := make([]superFileExplorerRoot, 0, 4)
	for drive := 'A'; drive <= 'Z'; drive++ {
		path := string(drive) + `:\`
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			roots = append(roots, superFileExplorerRoot{Label: path, Path: path})
		}
	}
	if len(roots) == 0 {
		if wd, err := os.Getwd(); err == nil {
			volume := filepath.VolumeName(wd)
			if volume != "" {
				roots = append(roots, superFileExplorerRoot{Label: volume + `\`, Path: volume + `\`})
			}
		}
	}
	return roots
}

func superFileExplorerResolvePath(rawPath string, roots []superFileExplorerRoot) string {
	candidate := strings.TrimSpace(rawPath)
	if candidate == "" {
		if len(roots) > 0 {
			return filepath.Clean(roots[0].Path)
		}
		return filepath.Clean(string(filepath.Separator))
	}

	if runtime.GOOS == "windows" && len(candidate) == 2 && candidate[1] == ':' {
		candidate += `\`
	}
	if !filepath.IsAbs(candidate) {
		if abs, err := filepath.Abs(candidate); err == nil {
			candidate = abs
		}
	}
	return filepath.Clean(candidate)
}

func superFileExplorerListEntries(targetPath string) ([]superFileExplorerEntry, error) {
	info, err := os.Stat(targetPath)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, os.ErrPermission
	}

	dirEntries, err := os.ReadDir(targetPath)
	if err != nil {
		return nil, err
	}

	entries := make([]superFileExplorerEntry, 0, len(dirEntries))
	for _, dirEntry := range dirEntries {
		name := dirEntry.Name()
		entryPath := filepath.Join(targetPath, name)
		entry := superFileExplorerEntry{
			Name:      name,
			Path:      entryPath,
			Extension: strings.TrimPrefix(strings.ToLower(filepath.Ext(name)), "."),
			Hidden:    strings.HasPrefix(name, "."),
		}

		info, err := dirEntry.Info()
		if err != nil {
			entry.Type = "unknown"
			entry.Error = "sin permisos de lectura"
			entries = append(entries, entry)
			continue
		}

		entry.Mode = info.Mode().String()
		entry.ModifiedAt = info.ModTime().UTC().Format(time.RFC3339)
		entry.IsSymlink = info.Mode()&os.ModeSymlink != 0
		entry.IsDir = dirEntry.IsDir()
		entry.Type = "file"
		if entry.IsDir {
			entry.Type = "folder"
		} else if entry.IsSymlink {
			entry.Type = "symlink"
		}
		if !entry.IsDir {
			entry.Size = info.Size()
			entry.SizePretty = superFileExplorerFormatBytes(info.Size())
		}
		entries = append(entries, entry)
	}

	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
	return entries, nil
}

func superFileExplorerParentPath(path string) string {
	clean := filepath.Clean(strings.TrimSpace(path))
	parent := filepath.Dir(clean)
	if parent == "." || parent == clean {
		return ""
	}
	if runtime.GOOS == "windows" && strings.TrimRight(parent, `\`) == filepath.VolumeName(parent) {
		if strings.HasSuffix(parent, `\`) {
			return parent
		}
		return parent + `\`
	}
	return parent
}

func superFileExplorerFormatBytes(size int64) string {
	if size < 0 {
		return ""
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	value := float64(size)
	unitIndex := 0
	for value >= 1024 && unitIndex < len(units)-1 {
		value /= 1024
		unitIndex++
	}
	if unitIndex == 0 {
		return strconv.FormatInt(size, 10) + " " + units[unitIndex]
	}
	return strconv.FormatFloat(value, 'f', 1, 64) + " " + units[unitIndex]
}
