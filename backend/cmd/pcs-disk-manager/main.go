// pcs-disk-manager is an internal-only, allowlisted Docker cleanup service.
// It deliberately exposes no shell, paths, Docker IDs, or arbitrary commands.
package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const maxRequestAge = time.Minute

var allowedCategories = map[string]cleanupCategory{
	"containers": {
		ID: "containers", Title: "Contenedores detenidos", Description: "Contenedores detenidos hace más de 24 horas.",
	},
	"images": {
		ID: "images", Title: "Imágenes Docker sin uso", Description: "Imágenes que ningún contenedor usa y que tienen más de 7 días.",
	},
	"builder_cache": {
		ID: "builder_cache", Title: "Caché de compilación", Description: "Caché BuildKit no usada durante más de 7 días.",
	},
	"anonymous_volumes": {
		ID: "anonymous_volumes", Title: "Volúmenes anónimos huérfanos", Description: "Volúmenes anónimos sin referencias de contenedores.",
	},
}

type cleanupCategory struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	CandidateCount int    `json:"candidate_count"`
	EstimatedBytes int64  `json:"estimated_bytes"`
	EstimateKnown  bool   `json:"estimate_known"`
}

type diskUsage struct {
	TotalBytes int64 `json:"total_bytes"`
	UsedBytes  int64 `json:"used_bytes"`
	FreeBytes  int64 `json:"free_bytes"`
}

type cleanupRequest struct {
	Action     string   `json:"action"`
	Categories []string `json:"categories,omitempty"`
}

type cleanupResponse struct {
	OK          bool              `json:"ok"`
	Disk        diskUsage         `json:"disk"`
	Categories  []cleanupCategory `json:"categories,omitempty"`
	FreedBytes  int64             `json:"freed_bytes,omitempty"`
	Cleaned     []string          `json:"cleaned,omitempty"`
	GeneratedAt string            `json:"generated_at"`
	Error       string            `json:"error,omitempty"`
}

type dockerSummary struct {
	Type        string `json:"Type"`
	TotalCount  string `json:"TotalCount"`
	Active      string `json:"Active"`
	Size        string `json:"Size"`
	Reclaimable string `json:"Reclaimable"`
}

type diskManager struct {
	secret []byte
	mu     sync.Mutex
}

func main() {
	secret := strings.TrimSpace(os.Getenv("PCS_DISK_MANAGER_SHARED_SECRET"))
	if secret == "" {
		log.Fatal("PCS_DISK_MANAGER_SHARED_SECRET es obligatorio")
	}
	address := strings.TrimSpace(os.Getenv("PCS_DISK_MANAGER_ADDR"))
	if address == "" {
		address = ":8083"
	}
	manager := &diskManager{secret: []byte(secret)}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", manager.health)
	mux.HandleFunc("/v1/disk", manager.handleDisk)
	server := &http.Server{Addr: address, Handler: mux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 90 * time.Second, IdleTimeout: 30 * time.Second}
	log.Printf("pcs disk manager escuchando en %s", address)
	log.Fatal(server.ListenAndServe())
}

func (m *diskManager) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]bool{"ok": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (m *diskManager) handleDisk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, cleanupResponse{OK: false, Error: "metodo no permitido"})
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	body, err := readAll(r.Body)
	if err != nil || !m.validSignature(r.Header.Get("X-PCS-Timestamp"), r.Header.Get("X-PCS-Signature"), body) {
		writeJSON(w, http.StatusUnauthorized, cleanupResponse{OK: false, Error: "solicitud interna no autorizada"})
		return
	}
	var req cleanupRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, cleanupResponse{OK: false, Error: "solicitud invalida"})
		return
	}

	switch strings.ToLower(strings.TrimSpace(req.Action)) {
	case "status":
		response, err := buildStatus()
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, cleanupResponse{OK: false, Error: "no se pudo consultar Docker"})
			return
		}
		writeJSON(w, http.StatusOK, response)
	case "cleanup":
		categories, err := normalizeCategories(req.Categories)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, cleanupResponse{OK: false, Error: "seleccion de limpieza invalida"})
			return
		}
		m.mu.Lock()
		defer m.mu.Unlock()
		before, err := getDiskUsage()
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, cleanupResponse{OK: false, Error: "no se pudo consultar el disco"})
			return
		}
		cleaned, err := runCleanup(categories)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, cleanupResponse{OK: false, Error: "la limpieza no pudo completarse"})
			return
		}
		after, err := getDiskUsage()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, cleanupResponse{OK: false, Error: "la limpieza termino sin poder medir el disco"})
			return
		}
		freed := after.FreeBytes - before.FreeBytes
		if freed < 0 {
			freed = 0
		}
		writeJSON(w, http.StatusOK, cleanupResponse{OK: true, Disk: after, FreedBytes: freed, Cleaned: cleaned, GeneratedAt: time.Now().UTC().Format(time.RFC3339)})
	default:
		writeJSON(w, http.StatusBadRequest, cleanupResponse{OK: false, Error: "accion no soportada"})
	}
}

func (m *diskManager) validSignature(rawTime, signature string, body []byte) bool {
	stamp, err := time.Parse(time.RFC3339, strings.TrimSpace(rawTime))
	if err != nil || time.Since(stamp).Abs() > maxRequestAge {
		return false
	}
	provided, err := hex.DecodeString(strings.TrimSpace(signature))
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(stamp.UTC().Format(time.RFC3339) + "\n"))
	_, _ = mac.Write(body)
	expected := mac.Sum(nil)
	return len(provided) == len(expected) && subtle.ConstantTimeCompare(provided, expected) == 1
}

func buildStatus() (cleanupResponse, error) {
	disk, err := getDiskUsage()
	if err != nil {
		return cleanupResponse{}, err
	}
	summary, err := dockerSystemDF()
	if err != nil {
		return cleanupResponse{}, err
	}
	volumes, err := anonymousDanglingVolumes()
	if err != nil {
		return cleanupResponse{}, err
	}
	categories := make([]cleanupCategory, 0, len(allowedCategories))
	for _, id := range []string{"containers", "images", "builder_cache", "anonymous_volumes"} {
		item := allowedCategories[id]
		switch id {
		case "containers":
			item.CandidateCount = summary["containers"].TotalCount - summary["containers"].Active
			item.EstimatedBytes = summary["containers"].Reclaimable
			item.EstimateKnown = true
		case "images":
			item.CandidateCount = summary["images"].TotalCount - summary["images"].Active
			item.EstimatedBytes = summary["images"].Reclaimable
			item.EstimateKnown = true
		case "builder_cache":
			item.CandidateCount = summary["build cache"].TotalCount - summary["build cache"].Active
			item.EstimatedBytes = summary["build cache"].Reclaimable
			item.EstimateKnown = true
		case "anonymous_volumes":
			item.CandidateCount = len(volumes)
			item.EstimateKnown = false
		}
		categories = append(categories, item)
	}
	return cleanupResponse{OK: true, Disk: disk, Categories: categories, GeneratedAt: time.Now().UTC().Format(time.RFC3339)}, nil
}

type summaryCounts struct {
	TotalCount, Active int
	Reclaimable        int64
}

func dockerSystemDF() (map[string]summaryCounts, error) {
	output, err := runDocker(20*time.Second, "system", "df", "--format", "{{json .}}")
	if err != nil {
		return nil, err
	}
	result := map[string]summaryCounts{}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var row dockerSummary
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return nil, err
		}
		total, _ := strconv.Atoi(strings.TrimSpace(row.TotalCount))
		active, _ := strconv.Atoi(strings.TrimSpace(row.Active))
		result[strings.ToLower(strings.TrimSpace(row.Type))] = summaryCounts{TotalCount: total, Active: active, Reclaimable: parseDockerBytes(row.Reclaimable)}
	}
	return result, nil
}

func anonymousDanglingVolumes() ([]string, error) {
	output, err := runDocker(15*time.Second, "volume", "ls", "-q", "--filter", "dangling=true", "--filter", "label=com.docker.volume.anonymous")
	if err != nil {
		return nil, err
	}
	items := []string{}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if value := strings.TrimSpace(line); value != "" {
			items = append(items, value)
		}
	}
	sort.Strings(items)
	return items, nil
}

func normalizeCategories(values []string) ([]string, error) {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(strings.ToLower(value))
		if _, ok := allowedCategories[value]; !ok || seen[value] {
			return nil, errors.New("categoria invalida")
		}
		seen[value] = true
		result = append(result, value)
	}
	if len(result) == 0 {
		return nil, errors.New("sin categorias")
	}
	sort.Strings(result)
	return result, nil
}

func runCleanup(categories []string) ([]string, error) {
	cleaned := make([]string, 0, len(categories))
	for _, category := range categories {
		var err error
		switch category {
		case "containers":
			_, err = runDocker(90*time.Second, "container", "prune", "-f", "--filter", "until=24h")
		case "images":
			_, err = runDocker(5*time.Minute, "image", "prune", "-a", "-f", "--filter", "until=168h")
		case "builder_cache":
			_, err = runDocker(5*time.Minute, "builder", "prune", "-a", "-f", "--filter", "until=168h")
		case "anonymous_volumes":
			var volumes []string
			volumes, err = anonymousDanglingVolumes()
			if err == nil {
				for _, volume := range volumes {
					if _, err = runDocker(90*time.Second, "volume", "rm", volume); err != nil {
						break
					}
				}
			}
		}
		if err != nil {
			return nil, err
		}
		cleaned = append(cleaned, category)
	}
	return cleaned, nil
}

func runDocker(timeout time.Duration, args ...string) ([]byte, error) {
	ctx, cancel := contextWithTimeout(timeout)
	defer cancel()
	return exec.CommandContext(ctx, "docker", args...).CombinedOutput()
}

func getDiskUsage() (diskUsage, error) {
	ctx, cancel := contextWithTimeout(10 * time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, "df", "-B1", "--output=size,used,avail", "/").Output()
	if err != nil {
		return diskUsage{}, err
	}
	fields := strings.Fields(string(output))
	if len(fields) < 6 {
		return diskUsage{}, errors.New("salida df invalida")
	}
	total, totalErr := strconv.ParseInt(fields[len(fields)-3], 10, 64)
	used, usedErr := strconv.ParseInt(fields[len(fields)-2], 10, 64)
	free, freeErr := strconv.ParseInt(fields[len(fields)-1], 10, 64)
	if totalErr != nil || usedErr != nil || freeErr != nil {
		return diskUsage{}, errors.New("valores df invalidos")
	}
	return diskUsage{TotalBytes: total, UsedBytes: used, FreeBytes: free}, nil
}

func parseDockerBytes(raw string) int64 {
	value := strings.TrimSpace(strings.SplitN(raw, "(", 2)[0])
	if value == "" {
		return 0
	}
	numberEnd := 0
	for numberEnd < len(value) {
		char := value[numberEnd]
		if (char >= '0' && char <= '9') || char == '.' || char == ',' {
			numberEnd++
			continue
		}
		break
	}
	if numberEnd == 0 {
		return 0
	}
	number, err := strconv.ParseFloat(strings.ReplaceAll(value[:numberEnd], ",", "."), 64)
	if err != nil {
		return 0
	}
	unit := strings.ToUpper(strings.TrimSpace(value[numberEnd:]))
	multiplier := float64(1)
	switch unit {
	case "", "B":
	case "KB", "KIB":
		multiplier = 1024
	case "MB", "MIB":
		multiplier = 1024 * 1024
	case "GB", "GIB":
		multiplier = 1024 * 1024 * 1024
	case "TB", "TIB":
		multiplier = 1024 * 1024 * 1024 * 1024
	default:
		return 0
	}
	return int64(number * multiplier)
}

func readAll(body io.ReadCloser) ([]byte, error) { defer body.Close(); return io.ReadAll(body) }

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func contextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}
