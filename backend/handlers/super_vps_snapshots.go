package handlers

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	superVPSSnapshotConfigEnabled            = "super.vps_snapshot.enabled"
	superVPSSnapshotConfigAutoEnabled        = "super.vps_snapshot.auto_enabled"
	superVPSSnapshotConfigIntervalHours      = "super.vps_snapshot.interval_hours"
	superVPSSnapshotConfigDailyTime          = "super.vps_snapshot.daily_time"
	superVPSSnapshotConfigRetentionDays      = "super.vps_snapshot.retention_days"
	superVPSSnapshotConfigDeleteOldLocal     = "super.vps_snapshot.delete_old_local"
	superVPSSnapshotConfigCloudEnabled       = "super.vps_snapshot.cloud_enabled"
	superVPSSnapshotConfigCloudProvider      = "super.vps_snapshot.cloud_provider"
	superVPSSnapshotConfigRcloneRemotePath   = "super.vps_snapshot.rclone_remote_path"
	superVPSSnapshotConfigDeleteOldCloud     = "super.vps_snapshot.delete_old_cloud"
	superVPSSnapshotConfigIncludeProject     = "super.vps_snapshot.include_project"
	superVPSSnapshotConfigIncludePostgres    = "super.vps_snapshot.include_postgres"
	superVPSSnapshotConfigIncludeVolumes     = "super.vps_snapshot.include_volumes"
	superVPSSnapshotConfigIncludeDockerImage = "super.vps_snapshot.include_docker_images"
	superVPSSnapshotConfigLastAutoRun        = "super.vps_snapshot.last_auto_run"
	superVPSSnapshotConfigLastResult         = "super.vps_snapshot.last_result"
)

var superVPSSnapshotRunning int32

type superVPSSnapshotConfig struct {
	Enabled             bool   `json:"enabled"`
	AutoEnabled         bool   `json:"auto_enabled"`
	IntervalHours       int    `json:"interval_hours"`
	DailyTime           string `json:"daily_time,omitempty"`
	RetentionDays       int    `json:"retention_days"`
	DeleteOldLocal      bool   `json:"delete_old_local"`
	CloudEnabled        bool   `json:"cloud_enabled"`
	CloudProvider       string `json:"cloud_provider"`
	RcloneRemotePath    string `json:"rclone_remote_path"`
	DeleteOldCloud      bool   `json:"delete_old_cloud"`
	IncludeProject      bool   `json:"include_project"`
	IncludePostgres     bool   `json:"include_postgres"`
	IncludeVolumes      bool   `json:"include_volumes"`
	IncludeDockerImages bool   `json:"include_docker_images"`
	LastAutoRun         string `json:"last_auto_run,omitempty"`
	LastResult          string `json:"last_result,omitempty"`
	RcloneAvailable     bool   `json:"rclone_available"`
	DockerAvailable     bool   `json:"docker_available"`
	SnapshotDir         string `json:"snapshot_dir"`
}

type superVPSSnapshotCreateRequest struct {
	UploadCloud         *bool  `json:"upload_cloud,omitempty"`
	Automatico          bool   `json:"automatico,omitempty"`
	IncludeEnvSecrets   bool   `json:"include_env_secrets,omitempty"`
	ConfirmEnvSecrets   string `json:"confirm_env_secrets,omitempty"`
	IncludeDockerImages *bool  `json:"include_docker_images,omitempty"`
	DeleteOldLocal      *bool  `json:"delete_old_local,omitempty"`
	DeleteOldCloud      *bool  `json:"delete_old_cloud,omitempty"`
	RetentionDays       int    `json:"retention_days,omitempty"`
	UsuarioCreador      string `json:"usuario_creador,omitempty"`
	DownloadAfterCreate bool   `json:"download_after_create,omitempty"`
}

type superVPSSnapshotManifest struct {
	Version             string   `json:"version"`
	CreatedAt           string   `json:"created_at"`
	CreatedBy           string   `json:"created_by,omitempty"`
	RuntimeOS           string   `json:"runtime_os"`
	ProjectRoot         string   `json:"project_root"`
	ArchiveName         string   `json:"archive_name"`
	Includes            []string `json:"includes"`
	Warnings            []string `json:"warnings,omitempty"`
	RestoreInstructions []string `json:"restore_instructions"`
}

type superVPSSnapshotCreateResult struct {
	OK       bool                      `json:"ok"`
	Item     dbpkg.SuperVPSSnapshotLog `json:"item,omitempty"`
	Config   superVPSSnapshotConfig    `json:"config,omitempty"`
	Warnings []string                  `json:"warnings,omitempty"`
	Error    string                    `json:"error,omitempty"`
}

func SuperVPSSnapshotsHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminEmail, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper)
		if !ok {
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "" {
			action = "status"
		}

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "status":
				cfg := getSuperVPSSnapshotConfig(dbSuper)
				logs, _ := dbpkg.ListSuperVPSSnapshotLogs(dbSuper, 30)
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "config": cfg, "items": sanitizeSuperVPSSnapshotLogs(logs)})
				return
			case "download":
				id, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("id")), 10, 64)
				serveSuperVPSSnapshotDownload(w, r, dbSuper, id)
				return
			default:
				writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "accion no soportada"})
				return
			}
		case http.MethodPut:
			var cfg superVPSSnapshotConfig
			if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "payload invalido"})
				return
			}
			if err := saveSuperVPSSnapshotConfig(dbSuper, cfg); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "no se pudo guardar la configuracion de snapshots"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "config": getSuperVPSSnapshotConfig(dbSuper)})
			return
		case http.MethodPost:
			switch action {
			case "create", "crear", "download_now":
				var payload superVPSSnapshotCreateRequest
				if r.Body != nil {
					_ = json.NewDecoder(r.Body).Decode(&payload)
				}
				payload.UsuarioCreador = firstNonBlank(payload.UsuarioCreador, adminEmail)
				result := CreateSuperVPSSnapshot(dbSuper, payload)
				status := http.StatusOK
				if !result.OK {
					status = http.StatusInternalServerError
				}
				writeJSON(w, status, result)
				return
			default:
				writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "accion no soportada"})
				return
			}
		default:
			w.Header().Set("Allow", "GET, POST, PUT")
			writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"ok": false, "error": "method not allowed"})
			return
		}
	}
}

func getSuperVPSSnapshotConfig(dbSuper *sql.DB) superVPSSnapshotConfig {
	read := func(key string) string {
		value, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, key)
		return strings.TrimSpace(value)
	}
	intervalHours := parsePositiveInt(read(superVPSSnapshotConfigIntervalHours), 24, 1, 24*30)
	retentionDays := parsePositiveInt(read(superVPSSnapshotConfigRetentionDays), 14, 1, 3650)
	cfg := superVPSSnapshotConfig{
		Enabled:             parseEmpresaUsuarioBool(read(superVPSSnapshotConfigEnabled), true),
		AutoEnabled:         parseEmpresaUsuarioBool(read(superVPSSnapshotConfigAutoEnabled), false),
		IntervalHours:       intervalHours,
		DailyTime:           normalizeSuperVPSSnapshotDailyTime(read(superVPSSnapshotConfigDailyTime)),
		RetentionDays:       retentionDays,
		DeleteOldLocal:      parseEmpresaUsuarioBool(read(superVPSSnapshotConfigDeleteOldLocal), false),
		CloudEnabled:        parseEmpresaUsuarioBool(read(superVPSSnapshotConfigCloudEnabled), false),
		CloudProvider:       normalizeSuperVPSSnapshotCloudProvider(read(superVPSSnapshotConfigCloudProvider)),
		RcloneRemotePath:    sanitizeSuperVPSSnapshotRemotePath(read(superVPSSnapshotConfigRcloneRemotePath)),
		DeleteOldCloud:      parseEmpresaUsuarioBool(read(superVPSSnapshotConfigDeleteOldCloud), false),
		IncludeProject:      parseEmpresaUsuarioBool(read(superVPSSnapshotConfigIncludeProject), true),
		IncludePostgres:     parseEmpresaUsuarioBool(read(superVPSSnapshotConfigIncludePostgres), true),
		IncludeVolumes:      parseEmpresaUsuarioBool(read(superVPSSnapshotConfigIncludeVolumes), true),
		IncludeDockerImages: parseEmpresaUsuarioBool(read(superVPSSnapshotConfigIncludeDockerImage), false),
		LastAutoRun:         read(superVPSSnapshotConfigLastAutoRun),
		LastResult:          read(superVPSSnapshotConfigLastResult),
		SnapshotDir:         superVPSSnapshotDir(),
	}
	cfg.RcloneAvailable = commandAvailable("rclone")
	cfg.DockerAvailable = commandAvailable("docker")
	return cfg
}

func saveSuperVPSSnapshotConfig(dbSuper *sql.DB, cfg superVPSSnapshotConfig) error {
	if dbSuper == nil {
		return fmt.Errorf("db super no disponible")
	}
	cfg.CloudProvider = normalizeSuperVPSSnapshotCloudProvider(cfg.CloudProvider)
	cfg.RcloneRemotePath = sanitizeSuperVPSSnapshotRemotePath(cfg.RcloneRemotePath)
	cfg.IntervalHours = clampInt(cfg.IntervalHours, 1, 24*30, 24)
	if strings.TrimSpace(cfg.DailyTime) != "" && normalizeSuperVPSSnapshotDailyTime(cfg.DailyTime) == "" {
		return fmt.Errorf("hora diaria invalida; use formato HH:MM")
	}
	cfg.DailyTime = normalizeSuperVPSSnapshotDailyTime(cfg.DailyTime)
	cfg.RetentionDays = clampInt(cfg.RetentionDays, 1, 3650, 14)
	if !cfg.IncludePostgres && !cfg.IncludeProject && !cfg.IncludeVolumes {
		return fmt.Errorf("seleccione al menos PostgreSQL, proyecto o volumenes para el snapshot")
	}
	pairs := map[string]string{
		superVPSSnapshotConfigEnabled:            strconv.FormatBool(cfg.Enabled),
		superVPSSnapshotConfigAutoEnabled:        strconv.FormatBool(cfg.AutoEnabled),
		superVPSSnapshotConfigIntervalHours:      strconv.Itoa(cfg.IntervalHours),
		superVPSSnapshotConfigDailyTime:          cfg.DailyTime,
		superVPSSnapshotConfigRetentionDays:      strconv.Itoa(cfg.RetentionDays),
		superVPSSnapshotConfigDeleteOldLocal:     strconv.FormatBool(cfg.DeleteOldLocal),
		superVPSSnapshotConfigCloudEnabled:       strconv.FormatBool(cfg.CloudEnabled),
		superVPSSnapshotConfigCloudProvider:      cfg.CloudProvider,
		superVPSSnapshotConfigRcloneRemotePath:   cfg.RcloneRemotePath,
		superVPSSnapshotConfigDeleteOldCloud:     strconv.FormatBool(cfg.DeleteOldCloud),
		superVPSSnapshotConfigIncludeProject:     strconv.FormatBool(cfg.IncludeProject),
		superVPSSnapshotConfigIncludePostgres:    strconv.FormatBool(cfg.IncludePostgres),
		superVPSSnapshotConfigIncludeVolumes:     strconv.FormatBool(cfg.IncludeVolumes),
		superVPSSnapshotConfigIncludeDockerImage: strconv.FormatBool(cfg.IncludeDockerImages),
	}
	for key, value := range pairs {
		if err := dbpkg.SetConfigValue(dbSuper, key, value, false); err != nil {
			return err
		}
	}
	return nil
}

func CreateSuperVPSSnapshot(dbSuper *sql.DB, req superVPSSnapshotCreateRequest) superVPSSnapshotCreateResult {
	if !atomic.CompareAndSwapInt32(&superVPSSnapshotRunning, 0, 1) {
		return superVPSSnapshotCreateResult{OK: false, Error: "ya hay un snapshot VPS en ejecucion"}
	}
	defer atomic.StoreInt32(&superVPSSnapshotRunning, 0)

	cfg := getSuperVPSSnapshotConfig(dbSuper)
	if !cfg.Enabled && !req.Automatico {
		return superVPSSnapshotCreateResult{OK: false, Config: cfg, Error: "snapshot VPS desactivado por configuracion"}
	}
	if req.Automatico && !cfg.AutoEnabled {
		return superVPSSnapshotCreateResult{OK: false, Config: cfg, Error: "snapshot automatico VPS desactivado"}
	}
	includeImages := cfg.IncludeDockerImages
	if req.IncludeDockerImages != nil {
		includeImages = *req.IncludeDockerImages
	}
	uploadCloud := cfg.CloudEnabled
	if req.UploadCloud != nil {
		uploadCloud = *req.UploadCloud
	}
	deleteOldLocal := cfg.DeleteOldLocal
	if req.DeleteOldLocal != nil {
		deleteOldLocal = *req.DeleteOldLocal
	}
	deleteOldCloud := cfg.DeleteOldCloud
	if req.DeleteOldCloud != nil {
		deleteOldCloud = *req.DeleteOldCloud
	}
	retentionDays := cfg.RetentionDays
	if req.RetentionDays > 0 {
		retentionDays = clampInt(req.RetentionDays, 1, 3650, cfg.RetentionDays)
	}
	includeSecrets := req.IncludeEnvSecrets && !req.Automatico && strings.EqualFold(strings.TrimSpace(req.ConfirmEnvSecrets), "INCLUIR_ENV")
	now := time.Now()
	code := "VPS-" + now.Format("20060102-150405")
	fileName := "pcs-vps-snapshot-" + now.Format("20060102-150405") + ".tar.gz"
	outDir := superVPSSnapshotDir()
	_ = os.MkdirAll(outDir, 0o700)
	outPath := filepath.Join(outDir, fileName)

	item := dbpkg.SuperVPSSnapshotLog{
		Codigo:          code,
		FileName:        fileName,
		FilePath:        outPath,
		Estado:          "en_proceso",
		CloudProvider:   cfg.CloudProvider,
		CloudDestino:    cfg.RcloneRemotePath,
		Automatico:      boolToSmallInt(req.Automatico),
		IncluyeSecretos: boolToSmallInt(includeSecrets),
		IncluyeImagenes: boolToSmallInt(includeImages),
		UsuarioCreador:  strings.TrimSpace(req.UsuarioCreador),
	}
	id, err := dbpkg.InsertSuperVPSSnapshotLog(dbSuper, item)
	if err == nil {
		item.ID = id
	}

	manifest, warnings, err := buildSuperVPSSnapshotArchive(outPath, item, cfg, includeSecrets, includeImages)
	if err != nil {
		item.Estado = "error"
		item.Error = superVPSSnapshotFailureMessage()
		_ = dbpkg.UpdateSuperVPSSnapshotLog(dbSuper, item)
		return superVPSSnapshotCreateResult{OK: false, Config: cfg, Item: sanitizeSuperVPSSnapshotLog(item), Warnings: sanitizeSuperVPSSnapshotWarnings(warnings), Error: superVPSSnapshotFailureMessage()}
	}
	item.ManifestJSON = jsonString(manifest)
	if info, statErr := os.Stat(outPath); statErr == nil {
		item.TamanoBytes = info.Size()
	}
	item.HashSHA256 = fileSHA256(outPath)
	item.Estado = "completado"

	if uploadCloud {
		cloudEstado, cloudMsg := uploadSuperVPSSnapshotWithRclone(outPath, cfg)
		item.CloudEstado = cloudEstado
		item.CloudMensaje = cloudMsg
		if cloudEstado == "error" {
			warnings = append(warnings, "Subida nube no completada. Revise el estado de la integracion de respaldo.")
		}
		if deleteOldCloud && cloudEstado == "subido" {
			if msg := cleanupSuperVPSSnapshotCloud(cfg, retentionDays); strings.TrimSpace(msg) != "" {
				warnings = append(warnings, msg)
			}
		}
	}
	if deleteOldLocal {
		deleted, cleanErr := cleanupOldSuperVPSSnapshotsLocal(outDir, retentionDays)
		if cleanErr != nil {
			warnings = append(warnings, "No se pudieron limpiar copias locales antiguas.")
		} else if deleted > 0 {
			warnings = append(warnings, fmt.Sprintf("Copias locales antiguas eliminadas: %d", deleted))
		}
	}
	_ = dbpkg.UpdateSuperVPSSnapshotLog(dbSuper, item)
	_ = dbpkg.SetConfigValue(dbSuper, superVPSSnapshotConfigLastResult, fmt.Sprintf("snapshot=%s tamano=%d warnings=%d", code, item.TamanoBytes, len(warnings)), false)
	if req.Automatico {
		_ = dbpkg.SetConfigValue(dbSuper, superVPSSnapshotConfigLastAutoRun, time.Now().Format(time.RFC3339), false)
	}
	return superVPSSnapshotCreateResult{OK: true, Config: getSuperVPSSnapshotConfig(dbSuper), Item: sanitizeSuperVPSSnapshotLog(item), Warnings: sanitizeSuperVPSSnapshotWarnings(warnings)}
}

func buildSuperVPSSnapshotArchive(outPath string, item dbpkg.SuperVPSSnapshotLog, cfg superVPSSnapshotConfig, includeSecrets, includeImages bool) (superVPSSnapshotManifest, []string, error) {
	root, err := superDockerPortableResolveRoot()
	if err != nil {
		root = resolveProjectRootDir()
	}
	warnings := []string{}
	includes := []string{}
	tempDir, err := os.MkdirTemp("", "pcs-vps-snapshot-*")
	if err != nil {
		return superVPSSnapshotManifest{}, warnings, err
	}
	defer os.RemoveAll(tempDir)

	// #nosec G304 -- path is normalized and constrained to a server-controlled root before this operation.
	file, err := os.Create(outPath)
	if err != nil {
		return superVPSSnapshotManifest{}, warnings, err
	}
	defer file.Close()
	gz := gzip.NewWriter(file)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()

	manifest := superVPSSnapshotManifest{
		Version:     "pcs-vps-snapshot.v1",
		CreatedAt:   time.Now().Format(time.RFC3339),
		CreatedBy:   item.UsuarioCreador,
		RuntimeOS:   runtime.GOOS,
		ProjectRoot: root,
		ArchiveName: item.FileName,
		RestoreInstructions: []string{
			"1. Restaurar el proyecto desde project/ o combinar con el paquete Docker portable.",
			"2. Restaurar PostgreSQL desde postgres/pg_dumpall.sql si existe.",
			"3. Restaurar volumenes desde docker-volumes/*.tar.gz con docker run y el volumen destino.",
			"4. Si se incluyo docker-images/docker-images.tar, cargar con docker load -i docker-images.tar.",
			"5. Revisar MANIFEST.json y RESTAURAR_VPS.md antes de levantar servicios.",
		},
	}
	if cfg.IncludeProject {
		if err := superVPSSnapshotAddProject(tw, root); err != nil {
			warnings = append(warnings, "No se pudo incluir proyecto portable: "+err.Error())
		} else {
			includes = append(includes, "project")
		}
	}
	if cfg.IncludePostgres {
		if dumpPath, dumpWarn := superVPSSnapshotCreatePostgresDump(tempDir); dumpPath != "" {
			if err := addFileToTar(tw, dumpPath, "postgres/pg_dumpall.sql"); err != nil {
				warnings = append(warnings, "No se pudo agregar dump PostgreSQL: "+err.Error())
			} else {
				includes = append(includes, "postgres")
			}
		} else if dumpWarn != "" {
			warnings = append(warnings, dumpWarn)
		}
	}
	if cfg.IncludeVolumes {
		added, volWarnings := superVPSSnapshotAddDockerVolumes(tw, tempDir)
		warnings = append(warnings, volWarnings...)
		if added > 0 {
			includes = append(includes, "docker-volumes")
		} else {
			if localAdded, localWarnings := superVPSSnapshotAddLocalRuntimeDirs(tw, root); localAdded > 0 {
				includes = append(includes, "runtime-local")
			} else {
				warnings = append(warnings, localWarnings...)
			}
		}
	}
	if includeImages {
		if imagePath, imgWarn := superVPSSnapshotCreateDockerImages(tempDir); imagePath != "" {
			if err := addFileToTar(tw, imagePath, "docker-images/docker-images.tar"); err != nil {
				warnings = append(warnings, "No se pudo agregar imagenes Docker: "+err.Error())
			} else {
				includes = append(includes, "docker-images")
			}
		} else if imgWarn != "" {
			warnings = append(warnings, imgWarn)
		}
	}
	if includeSecrets {
		envPath := filepath.Join(root, "deploy", ".env.platform")
		if fileExists(envPath) {
			if err := addFileToTar(tw, envPath, "secrets/env.platform.backup"); err != nil {
				warnings = append(warnings, "No se pudo incluir .env.platform: "+err.Error())
			} else {
				includes = append(includes, "secrets")
			}
		} else {
			warnings = append(warnings, "No se encontro deploy/.env.platform para incluir secretos")
		}
	}
	sort.Strings(includes)
	manifest.Includes = includes
	manifest.Warnings = warnings
	if err := addBytesToTar(tw, "MANIFEST.json", []byte(jsonPretty(manifest)+"\n"), 0o600); err != nil {
		return manifest, warnings, err
	}
	if err := addBytesToTar(tw, "RESTAURAR_VPS.md", []byte(superVPSSnapshotRestoreReadme(manifest)+"\n"), 0o600); err != nil {
		return manifest, warnings, err
	}
	return manifest, warnings, nil
}

func superVPSSnapshotAddProject(tw *tar.Writer, root string) error {
	if err := superDockerPortableWriteReadme(tw, "project"); err != nil {
		return err
	}
	return superDockerPortableWalk(root, func(path, rel string, info os.FileInfo) error {
		if info == nil || info.IsDir() {
			return nil
		}
		return superDockerPortableAddFile(tw, root, path, rel, "project", info)
	})
}

func superVPSSnapshotCreatePostgresDump(tempDir string) (string, string) {
	out := filepath.Join(tempDir, "pg_dumpall.sql")
	if commandAvailable("docker") {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, "docker", "exec", "pcs-postgres", "sh", "-lc", `pg_dumpall -U "$POSTGRES_USER"`)
		data, err := cmd.Output()
		if err == nil && len(data) > 0 {
			if writeErr := os.WriteFile(out, data, 0o600); writeErr == nil {
				return out, ""
			}
		}
	}
	if commandAvailable("pg_dumpall") {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, "pg_dumpall")
		data, err := cmd.Output()
		if err == nil && len(data) > 0 {
			if writeErr := os.WriteFile(out, data, 0o600); writeErr == nil {
				return out, ""
			}
		}
	}
	return "", "No se genero pg_dumpall: docker pcs-postgres o pg_dumpall no disponible/accesible"
}

func superVPSSnapshotAddDockerVolumes(tw *tar.Writer, tempDir string) (int, []string) {
	warnings := []string{}
	if !commandAvailable("docker") {
		return 0, []string{"Docker no disponible para empaquetar volumenes"}
	}
	volumes := []string{
		"powerful-control-system_pcs_web_uploads",
		"powerful-control-system_pcs_downloads",
		"powerful-control-system_pcs_backend_logs",
		"powerful-control-system_pcs_backups",
		"powerful-control-system_pcs_postgres_data",
		"powerful-control-system_pcs_letsencrypt",
		"powerful-control-system_pcs_certbot_www",
	}
	added := 0
	for _, volume := range volumes {
		if !dockerVolumeExists(volume) {
			continue
		}
		safeName := sanitizeFileName(volume) + ".tar.gz"
		tempArchive := filepath.Join(tempDir, safeName)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		// #nosec G204 -- Docker is fixed; volume comes from Docker inventory and safeName is sanitized.
		cmd := exec.CommandContext(ctx, "docker", "run", "--rm", "-v", volume+":/volume:ro", "-v", tempDir+":/backup", "alpine:3.20", "sh", "-lc", "cd /volume && tar -czf /backup/"+safeName+" .")
		err := cmd.Run()
		cancel()
		if err != nil {
			warnings = append(warnings, "No se pudo empaquetar volumen "+volume)
			continue
		}
		if err := addFileToTar(tw, tempArchive, "docker-volumes/"+safeName); err != nil {
			warnings = append(warnings, "No se pudo agregar volumen "+volume+": "+err.Error())
			continue
		}
		added++
	}
	if added == 0 {
		warnings = append(warnings, "No se encontraron volumenes Docker PCS para respaldo")
	}
	return added, warnings
}

func superVPSSnapshotAddLocalRuntimeDirs(tw *tar.Writer, root string) (int, []string) {
	targets := map[string]string{
		filepath.Join(root, "web", "uploads"):  "runtime-local/web_uploads",
		filepath.Join(root, "descargas"):       "runtime-local/descargas",
		filepath.Join(root, "backend", "logs"): "runtime-local/backend_logs",
	}
	if privateRoot := strings.TrimSpace(os.Getenv("PCS_PRIVATE_STORAGE_DIR")); privateRoot != "" {
		targets[privateRoot] = "runtime-local/private_storage"
	}
	added := 0
	warnings := []string{}
	for path, prefix := range targets {
		if !dirExists(path) {
			continue
		}
		if err := addDirToTar(tw, path, prefix); err != nil {
			warnings = append(warnings, "No se pudo incluir "+path+": "+err.Error())
			continue
		}
		added++
	}
	if added == 0 {
		warnings = append(warnings, "No se encontraron directorios runtime locales para respaldo")
	}
	return added, warnings
}

func superVPSSnapshotCreateDockerImages(tempDir string) (string, string) {
	if !commandAvailable("docker") {
		return "", "Docker no disponible para guardar imagenes"
	}
	containers := []string{"pcs-backend", "pcs-frontend", "pcs-edge", "pcs-postgres", "pcs-onlyoffice"}
	images := []string{}
	seen := map[string]bool{}
	for _, c := range containers {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		// #nosec G204 -- fixed Docker command; container names originate from Docker inventory.
		out, err := exec.CommandContext(ctx, "docker", "inspect", "--format", "{{.Image}}", c).Output()
		cancel()
		img := strings.TrimSpace(string(out))
		if err == nil && img != "" && !seen[img] {
			seen[img] = true
			images = append(images, img)
		}
	}
	if len(images) == 0 {
		return "", "No se encontraron imagenes de contenedores PCS activos para docker save"
	}
	out := filepath.Join(tempDir, "docker-images.tar")
	args := append([]string{"image", "save", "-o", out}, images...)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	// #nosec G204 -- fixed Docker executable; image identifiers originate from Docker inspect.
	err := exec.CommandContext(ctx, "docker", args...).Run()
	cancel()
	if err != nil {
		return "", "No se pudo ejecutar docker image save"
	}
	return out, ""
}

func serveSuperVPSSnapshotDownload(w http.ResponseWriter, r *http.Request, dbSuper *sql.DB, id int64) {
	item, err := dbpkg.GetSuperVPSSnapshotLog(dbSuper, id)
	if err != nil || item == nil {
		writeJSON(w, http.StatusNotFound, map[string]interface{}{"ok": false, "error": "snapshot no encontrado"})
		return
	}
	path, ok := safeSuperVPSSnapshotPath(item.FilePath)
	if !ok || !fileExists(path) {
		writeJSON(w, http.StatusNotFound, map[string]interface{}{"ok": false, "error": "archivo de snapshot no disponible"})
		return
	}
	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filepath.Base(path)+`"`)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, r, path)
}

func uploadSuperVPSSnapshotWithRclone(path string, cfg superVPSSnapshotConfig) (string, string) {
	if !cfg.CloudEnabled {
		return "omitido", "subida nube desactivada"
	}
	if !commandAvailable("rclone") {
		return "error", "rclone no esta instalado o no esta en PATH"
	}
	remote := sanitizeSuperVPSSnapshotRemotePath(cfg.RcloneRemotePath)
	if remote == "" || !strings.Contains(remote, ":") {
		return "error", "ruta rclone invalida; usa formato remoto:carpeta"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Minute)
	defer cancel()
	// #nosec G204 -- fixed rclone executable; remote syntax and local snapshot path are validated above.
	out, err := exec.CommandContext(ctx, "rclone", "copy", path, remote, "--checksum").CombinedOutput()
	if err != nil {
		return "error", sanitizeCommandOutput(string(out), 500)
	}
	return "subido", "archivo copiado a " + remote
}

func cleanupSuperVPSSnapshotCloud(cfg superVPSSnapshotConfig, retentionDays int) string {
	remote := sanitizeSuperVPSSnapshotRemotePath(cfg.RcloneRemotePath)
	if !commandAvailable("rclone") || remote == "" || retentionDays <= 0 {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	// #nosec G204 -- fixed rclone executable; remote and retention are validated server-side.
	out, err := exec.CommandContext(ctx, "rclone", "delete", remote, "--min-age", strconv.Itoa(retentionDays)+"d", "--include", "pcs-vps-snapshot-*.tar.gz").CombinedOutput()
	if err != nil {
		return "No se pudo limpiar nube: " + sanitizeCommandOutput(string(out), 300)
	}
	return "Limpieza nube ejecutada para copias mayores a " + strconv.Itoa(retentionDays) + " dias"
}

func cleanupOldSuperVPSSnapshotsLocal(dir string, retentionDays int) (int, error) {
	if retentionDays <= 0 {
		return 0, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	deleted := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "pcs-vps-snapshot-") || !strings.HasSuffix(entry.Name(), ".tar.gz") {
			continue
		}
		full := filepath.Join(dir, entry.Name())
		info, err := entry.Info()
		if err != nil || info.ModTime().After(cutoff) {
			continue
		}
		if safePath, ok := safeSuperVPSSnapshotPath(full); ok {
			if err := os.Remove(safePath); err == nil {
				deleted++
			}
		}
	}
	return deleted, nil
}

func StartSuperVPSSnapshotWorker(dbSuper *sql.DB, interval time.Duration, stop <-chan struct{}) {
	if dbSuper == nil {
		return
	}
	if interval <= 0 {
		interval = time.Hour
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			maybeRunSuperVPSSnapshotAutomatic(dbSuper)
		case <-stop:
			return
		}
	}
}

func maybeRunSuperVPSSnapshotAutomatic(dbSuper *sql.DB) {
	cfg := getSuperVPSSnapshotConfig(dbSuper)
	if !cfg.Enabled || !cfg.AutoEnabled {
		return
	}
	if cfg.CloudEnabled && strings.TrimSpace(cfg.RcloneRemotePath) == "" {
		_ = dbpkg.SetConfigValue(dbSuper, superVPSSnapshotConfigLastResult, "omitido: nube activa sin ruta rclone", false)
		return
	}
	if due := superVPSSnapshotAutomaticDue(cfg, time.Now()); !due {
		return
	}
	result := CreateSuperVPSSnapshot(dbSuper, superVPSSnapshotCreateRequest{Automatico: true, UploadCloud: &cfg.CloudEnabled, UsuarioCreador: "sistema.vps_snapshot_worker"})
	if !result.OK {
		_ = dbpkg.SetConfigValue(dbSuper, superVPSSnapshotConfigLastResult, "error: "+result.Error, false)
		log.Printf("[super_vps_snapshot] automatic error: %s", result.Error)
	}
}

func normalizeSuperVPSSnapshotDailyTime(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	if _, err := time.Parse("15:04", value); err != nil {
		return ""
	}
	return value
}

// superVPSSnapshotAutomaticDue supports a fixed daily hour when configured;
// otherwise it retains the existing interval schedule for compatibility.
func superVPSSnapshotAutomaticDue(cfg superVPSSnapshotConfig, now time.Time) bool {
	last, lastErr := time.Parse(time.RFC3339, strings.TrimSpace(cfg.LastAutoRun))
	hasLast := lastErr == nil
	if cfg.DailyTime != "" {
		clock, _ := time.Parse("15:04", cfg.DailyTime)
		dueAt := time.Date(now.Year(), now.Month(), now.Day(), clock.Hour(), clock.Minute(), 0, 0, now.Location())
		if now.Before(dueAt) {
			return false
		}
		return !hasLast || last.In(now.Location()).Format("2006-01-02") != now.Format("2006-01-02")
	}
	return !hasLast || now.Sub(last) >= time.Duration(cfg.IntervalHours)*time.Hour
}

// RunSuperVPSSnapshotScheduled executes one due-check without owning a timer.
func RunSuperVPSSnapshotScheduled(dbSuper *sql.DB) error {
	if dbSuper == nil {
		return fmt.Errorf("super database unavailable")
	}
	maybeRunSuperVPSSnapshotAutomatic(dbSuper)
	return nil
}

func sanitizeSuperVPSSnapshotLogs(items []dbpkg.SuperVPSSnapshotLog) []dbpkg.SuperVPSSnapshotLog {
	out := make([]dbpkg.SuperVPSSnapshotLog, 0, len(items))
	for _, item := range items {
		out = append(out, sanitizeSuperVPSSnapshotLog(item))
	}
	return out
}

func sanitizeSuperVPSSnapshotLog(item dbpkg.SuperVPSSnapshotLog) dbpkg.SuperVPSSnapshotLog {
	item.FilePath = ""
	item.ManifestJSON = ""
	item.Error = ""
	item.CloudMensaje = ""
	return item
}

func superVPSSnapshotFailureMessage() string {
	return "No se pudo crear el snapshot VPS. Revise el estado del servidor e intente nuevamente."
}

func sanitizeSuperVPSSnapshotWarnings(warnings []string) []string {
	out := make([]string, 0, len(warnings))
	for _, warning := range warnings {
		warning = strings.TrimSpace(warning)
		if warning == "" {
			continue
		}
		out = append(out, warning)
	}
	return out
}

func superVPSSnapshotDir() string {
	return filepath.Join(backupRootDir(), "vps_snapshots")
}

func safeSuperVPSSnapshotPath(path string) (string, bool) {
	root, err := filepath.Abs(superVPSSnapshotDir())
	if err != nil {
		return "", false
	}
	full, err := filepath.Abs(strings.TrimSpace(path))
	if err != nil {
		return "", false
	}
	root = filepath.Clean(root)
	full = filepath.Clean(full)
	if full == root || !strings.HasPrefix(full, root+string(os.PathSeparator)) {
		return "", false
	}
	return full, true
}

func normalizeSuperVPSSnapshotCloudProvider(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "rclone", "google_drive", "gdrive", "mega", "onedrive", "s3":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return "rclone"
	}
}

func sanitizeSuperVPSSnapshotRemotePath(raw string) string {
	value := strings.TrimSpace(raw)
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", "")
	if len(value) > 240 {
		value = value[:240]
	}
	return value
}

func commandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func dockerVolumeExists(name string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, "docker", "volume", "inspect", name).Run() == nil
}

func addFileToTar(tw *tar.Writer, path, name string) error {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return err
	}
	// #nosec G304 -- path is normalized and constrained to a server-controlled root before this operation.
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = filepath.ToSlash(name)
	header.Mode = int64(info.Mode().Perm())
	if header.Mode == 0 {
		header.Mode = 0o600
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	_, err = io.Copy(tw, file)
	return err
}

func addBytesToTar(tw *tar.Writer, name string, body []byte, mode int64) error {
	if mode == 0 {
		mode = 0o600
	}
	header := &tar.Header{Name: filepath.ToSlash(name), Mode: mode, Size: int64(len(body)), ModTime: time.Now()}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	_, err := tw.Write(body)
	return err
}

func addDirToTar(tw *tar.Writer, root, prefix string) error {
	root = filepath.Clean(root)
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil || path == root {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil || strings.HasPrefix(rel, "..") {
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil || info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		return addFileToTar(tw, path, filepath.ToSlash(filepath.Join(prefix, rel)))
	})
}

func fileSHA256(path string) string {
	// #nosec G304 -- path is normalized and constrained to a server-controlled root before this operation.
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()
	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func sanitizeFileName(value string) string {
	value = strings.TrimSpace(value)
	var b strings.Builder
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	if b.Len() == 0 {
		return "archivo"
	}
	return b.String()
}

func sanitizeCommandOutput(value string, max int) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	if max > 0 && len(value) > max {
		value = value[:max]
	}
	return value
}

func parsePositiveInt(raw string, fallback, min, max int) int {
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return clampInt(v, min, max, fallback)
}

func clampInt(value, min, max, fallback int) int {
	if value <= 0 {
		value = fallback
	}
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func boolToSmallInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func jsonString(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func jsonPretty(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}

func superVPSSnapshotRestoreReadme(manifest superVPSSnapshotManifest) string {
	return strings.Join([]string{
		"# Restauracion de snapshot VPS PCS",
		"",
		"Este paquete contiene una copia operativa generada desde Super Administrador.",
		"",
		"## Contenido",
		"",
		"- `project/`: codigo portable y despliegue Docker sin secretos.",
		"- `postgres/pg_dumpall.sql`: dump logico de PostgreSQL cuando estuvo disponible.",
		"- `docker-volumes/`: volumenes persistentes empaquetados cuando Docker estuvo disponible.",
		"- `docker-images/`: imagenes Docker solo si se activo la opcion manual.",
		"- `secrets/`: solo aparece si el super administrador confirmo incluir `.env.platform`.",
		"",
		"## Pasos generales",
		"",
		"1. Preparar una VPS limpia con Docker Engine y Compose v2.",
		"2. Copiar `project/` a `/root/powerfulcontrolsystem`.",
		"3. Configurar `deploy/.env.platform` con secretos reales o restaurar `secrets/env.platform.backup` si se incluyo deliberadamente.",
		"4. Restaurar volumenes creando los volumenes destino y descomprimiendo cada tarball.",
		"5. Restaurar PostgreSQL con `psql` dentro del contenedor nuevo usando `postgres/pg_dumpall.sql`.",
		"6. Ejecutar `bash deploy/scripts/vps-docker-preflight.sh` y luego levantar Compose.",
		"",
		"## Advertencias del manifiesto",
		"",
		strings.Join(manifest.Warnings, "\n"),
	}, "\n")
}
