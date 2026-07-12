package handlers

import (
	"archive/tar"
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

type superDockerPortableStatus struct {
	OK                 bool     `json:"ok"`
	GeneratedAt        string   `json:"generated_at"`
	RuntimeOS          string   `json:"runtime_os"`
	ProjectRoot        string   `json:"project_root"`
	ArchiveName        string   `json:"archive_name"`
	Ready              bool     `json:"ready"`
	ComposeFile        string   `json:"compose_file"`
	EnvExample         string   `json:"env_example"`
	DockerfileBackend  string   `json:"dockerfile_backend"`
	DockerfileFrontend string   `json:"dockerfile_frontend"`
	IncludedDirs       []string `json:"included_dirs"`
	IncludedFiles      []string `json:"included_files"`
	ExcludedPolicy     []string `json:"excluded_policy"`
	Profiles           []string `json:"profiles"`
	EstimatedFiles     int      `json:"estimated_files"`
	EstimatedSize      int64    `json:"estimated_size"`
	EstimatedSizeHuman string   `json:"estimated_size_human"`
	Warnings           []string `json:"warnings"`
	Error              string   `json:"error,omitempty"`
}

type superDockerPortableWalkStats struct {
	Files int
	Size  int64
}

func SuperDockerPortabilidadHandler(dbSuper *sql.DB) http.HandlerFunc {
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
			action = "status"
		}

		root, err := superDockerPortableResolveRoot()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, superDockerPortableStatus{
				OK:          false,
				GeneratedAt: time.Now().UTC().Format(time.RFC3339),
				RuntimeOS:   runtime.GOOS,
				Error:       "no se encontro una raiz de proyecto portable con deploy/docker-compose.platform.yml",
			})
			return
		}

		switch action {
		case "status":
			status := superDockerPortableBuildStatus(root)
			status.OK = true
			writeJSON(w, http.StatusOK, status)
		case "download":
			status := superDockerPortableBuildStatus(root)
			if !status.Ready {
				writeJSON(w, http.StatusConflict, status)
				return
			}
			superDockerPortableDownload(w, r, root, status.ArchiveName)
		default:
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "accion no soportada"})
		}
	}
}

func superDockerPortableResolveRoot() (string, error) {
	candidates := []string{}
	if envRoot := strings.TrimSpace(os.Getenv("PCS_PROJECT_EXPORT_ROOT")); envRoot != "" {
		candidates = append(candidates, envRoot)
	}
	candidates = append(candidates, "/app/project_export", "/app")
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, wd, filepath.Dir(wd), filepath.Join(wd, ".."))
	}
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates, exeDir, filepath.Dir(exeDir), filepath.Join(exeDir, ".."), filepath.Join(exeDir, "..", ".."))
	}

	seen := map[string]bool{}
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		abs, err := filepath.Abs(candidate)
		if err != nil {
			abs = candidate
		}
		abs = filepath.Clean(abs)
		if seen[abs] {
			continue
		}
		seen[abs] = true
		if superDockerPortableRootLooksValid(abs) {
			return abs, nil
		}
	}
	return "", os.ErrNotExist
}

func superDockerPortableRootLooksValid(root string) bool {
	requiredFiles := []string{
		filepath.Join(root, "deploy", "docker-compose.platform.yml"),
		filepath.Join(root, "deploy", ".env.platform.example"),
		filepath.Join(root, "deploy", "docker", "backend.Dockerfile"),
		filepath.Join(root, "deploy", "docker", "frontend.Dockerfile"),
		filepath.Join(root, "backend", "go.mod"),
	}
	for _, file := range requiredFiles {
		info, err := os.Stat(file)
		if err != nil || info.IsDir() {
			return false
		}
	}
	if info, err := os.Stat(filepath.Join(root, "web")); err != nil || !info.IsDir() {
		return false
	}
	return true
}

func superDockerPortableBuildStatus(root string) superDockerPortableStatus {
	now := time.Now().UTC()
	status := superDockerPortableStatus{
		GeneratedAt:        now.Format(time.RFC3339),
		RuntimeOS:          runtime.GOOS,
		ProjectRoot:        root,
		ArchiveName:        "powerfulcontrolsystem-docker-portable-" + now.Format("20060102-150405") + ".tar.gz",
		ComposeFile:        filepath.Join(root, "deploy", "docker-compose.platform.yml"),
		EnvExample:         filepath.Join(root, "deploy", ".env.platform.example"),
		DockerfileBackend:  filepath.Join(root, "deploy", "docker", "backend.Dockerfile"),
		DockerfileFrontend: filepath.Join(root, "deploy", "docker", "frontend.Dockerfile"),
		IncludedDirs:       []string{"backend", "web", "deploy", "scripts", "documentos"},
		IncludedFiles:      []string{".dockerignore", "AGENTS.md", "CHANGELOG.md", "README.md"},
		ExcludedPolicy: []string{
			"secretos y entornos: .env*, deploy/.env.platform, backend/.env*",
			"datos privados/runtime: web/uploads, descargas, backup, backups, logs",
			"caches y builds: .git, node_modules, tmp, test_runs, binarios, evidencias QA",
			"llaves: *.pem, *.key, *.ppk",
		},
		Profiles: []string{"edge", "certbot", "office", "voice", "rustdesk"},
	}

	required := []string{status.ComposeFile, status.EnvExample, status.DockerfileBackend, status.DockerfileFrontend}
	for _, file := range required {
		if info, err := os.Stat(file); err != nil || info.IsDir() {
			status.Warnings = append(status.Warnings, "Falta "+file)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "deploy", "scripts", "vps-docker-preflight.sh")); err != nil {
		status.Warnings = append(status.Warnings, "Falta script de preflight Docker VPS")
	}
	if _, err := os.Stat(filepath.Join(root, "documentos", "docker_vps_operacion.md")); err != nil {
		status.Warnings = append(status.Warnings, "Falta documento operativo Docker VPS")
	}

	stats, err := superDockerPortableCollectStats(root)
	if err != nil {
		status.Warnings = append(status.Warnings, "No se pudo estimar el paquete: "+err.Error())
	} else {
		status.EstimatedFiles = stats.Files
		status.EstimatedSize = stats.Size
		status.EstimatedSizeHuman = superFileExplorerFormatBytes(stats.Size)
	}

	status.Ready = len(status.Warnings) == 0
	return status
}

func superDockerPortableCollectStats(root string) (superDockerPortableWalkStats, error) {
	var stats superDockerPortableWalkStats
	err := superDockerPortableWalk(root, func(path, rel string, info os.FileInfo) error {
		if info == nil || info.IsDir() {
			return nil
		}
		stats.Files++
		stats.Size += info.Size()
		return nil
	})
	return stats, err
}

func superDockerPortableDownload(w http.ResponseWriter, r *http.Request, root, archiveName string) {
	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+archiveName+`"`)
	w.Header().Set("Cache-Control", "no-store")

	gz := gzip.NewWriter(w)
	defer gz.Close()

	tw := tar.NewWriter(gz)
	defer tw.Close()

	prefix := strings.TrimSuffix(archiveName, ".tar.gz")
	if err := superDockerPortableWriteReadme(tw, prefix); err != nil {
		return
	}

	_ = superDockerPortableWalk(root, func(path, rel string, info os.FileInfo) error {
		if info == nil || info.IsDir() {
			return nil
		}
		return superDockerPortableAddFile(tw, root, path, rel, prefix, info)
	})
}

func superDockerPortableWriteReadme(tw *tar.Writer, prefix string) error {
	content := strings.Join([]string{
		"# Powerful Control System - paquete Docker portable",
		"",
		"Este paquete fue generado desde el panel de Super Administrador.",
		"",
		"## Que incluye",
		"",
		"- Codigo fuente necesario para construir backend/frontend.",
		"- `deploy/docker-compose.platform.yml` con PostgreSQL, backend, frontend, edge, certbot y perfiles opcionales.",
		"- `deploy/.env.platform.example` como plantilla de variables.",
		"- Scripts y documentacion operativa de migracion VPS.",
		"",
		"## Que no incluye",
		"",
		"- Secretos reales (`deploy/.env.platform`, `backend/.env*`, claves privadas).",
		"- Datos de clientes, uploads, descargas, backups, logs o volumenes PostgreSQL.",
		"- Caches, evidencias QA, binarios generados o carpetas temporales.",
		"",
		"## Levantar en otra VPS",
		"",
		"1. Copia el paquete al nuevo servidor y descomprímelo en `/root/powerfulcontrolsystem`.",
		"2. Copia `deploy/.env.platform.example` como `deploy/.env.platform`, cambia los secretos y deja permisos `600`.",
		"3. Restaura los volumenes Docker y el backup PostgreSQL si estas migrando datos reales.",
		"4. Ejecuta `bash deploy/scripts/vps-docker-preflight.sh`.",
		"5. Ejecuta `docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml up -d --build`.",
		"6. Para publicar 80/443 desde Docker, valida DNS y ejecuta `CONFIRM_DOCKER_EDGE=YES bash deploy/scripts/vps-docker-edge-up.sh`.",
		"",
		"Consulta tambien `documentos/docker_vps_operacion.md` y `documentos/gobernanza_tecnica/runbooks/runbook_recuperacion_desastre_docker_vps.md`.",
		"",
	}, "\n")
	header := &tar.Header{
		Name:    prefix + "/LEEME_MIGRACION_DOCKER.md",
		Mode:    0644,
		Size:    int64(len([]byte(content))),
		ModTime: time.Now(),
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	_, err := tw.Write([]byte(content))
	return err
}

func superDockerPortableAddFile(tw *tar.Writer, root, path, rel, prefix string, info os.FileInfo) error {
	// #nosec G304 -- path is normalized and constrained to a server-controlled root before this operation.
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return nil
	}
	header.Name = prefix + "/" + filepath.ToSlash(rel)
	header.Mode = int64(info.Mode().Perm())
	if header.Mode == 0 {
		header.Mode = 0644
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	_, err = io.Copy(tw, file)
	return err
}

func superDockerPortableWalk(root string, visit func(path, rel string, info os.FileInfo) error) error {
	root = filepath.Clean(root)
	allowedTop := map[string]bool{
		"backend":    true,
		"web":        true,
		"deploy":     true,
		"scripts":    true,
		"documentos": true,
	}
	allowedRootFiles := map[string]bool{
		".dockerignore": true,
		"AGENTS.md":     true,
		"CHANGELOG.md":  true,
		"README.md":     true,
	}

	return filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if path == root {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		rel = filepath.Clean(rel)
		if strings.HasPrefix(rel, "..") {
			return nil
		}

		parts := strings.Split(filepath.ToSlash(rel), "/")
		top := parts[0]
		if len(parts) == 1 && !entry.IsDir() && !allowedRootFiles[top] {
			return nil
		}
		if entry.IsDir() {
			if len(parts) == 1 && !allowedTop[top] {
				return filepath.SkipDir
			}
			if superDockerPortableShouldExclude(rel, true) {
				return filepath.SkipDir
			}
			return nil
		}
		if !allowedTop[top] && !allowedRootFiles[top] {
			return nil
		}
		if superDockerPortableShouldExclude(rel, false) {
			return nil
		}
		info, err := entry.Info()
		if err != nil || info == nil {
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		return visit(path, rel, info)
	})
}

func superDockerPortableShouldExclude(rel string, isDir bool) bool {
	clean := filepath.ToSlash(filepath.Clean(rel))
	lower := strings.ToLower(clean)
	base := strings.ToLower(filepath.Base(clean))
	segments := strings.Split(lower, "/")
	blockedSegments := map[string]bool{
		".git": true, ".github": true, ".vscode": true, ".idea": true, ".codex": true,
		".cache": true, ".gocache": true, ".gotmp": true, "node_modules": true,
		"backup": true, "backups": true, "logs": true, "tmp": true, "test_runs": true,
		"bin": true, "dist": true, "build": true, "evidencias_qa": true,
		"uploads": true, "descargas": true, "tmp_tools": true, "juegos": true,
	}
	for _, segment := range segments {
		if blockedSegments[segment] {
			return true
		}
		if strings.HasPrefix(segment, ".env") && !strings.HasSuffix(segment, ".example") {
			return true
		}
		if strings.HasPrefix(segment, ".codex-") {
			return true
		}
	}
	if strings.HasPrefix(base, ".env") && !strings.HasSuffix(base, ".example") {
		return true
	}
	if strings.HasSuffix(base, ".pem") || strings.HasSuffix(base, ".key") || strings.HasSuffix(base, ".ppk") {
		return true
	}
	if strings.Contains(lower, "deploy/.env.platform") && !strings.HasSuffix(lower, ".example") {
		return true
	}
	if strings.Contains(lower, "backend/.env") {
		return true
	}
	if isDir {
		return false
	}
	if strings.HasSuffix(base, ".log") || strings.HasSuffix(base, ".tmp") || strings.HasSuffix(base, ".bak") || strings.HasSuffix(base, ".db") || strings.HasSuffix(base, ".exe") {
		return true
	}
	return false
}

func superDockerPortableJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func superDockerPortableAssertReadyForTests(root string) error {
	if !superDockerPortableRootLooksValid(root) {
		return errors.New("raiz docker portable invalida: " + fmt.Sprint(root))
	}
	return nil
}

func superDockerPortableSortedIncludedForTests(root string) ([]string, error) {
	files := []string{}
	err := superDockerPortableWalk(root, func(path, rel string, info os.FileInfo) error {
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	sort.Strings(files)
	return files, err
}
