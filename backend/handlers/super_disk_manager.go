package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

const superDiskManagerConfirmation = "LIBERAR ESPACIO"

var superDiskManagerRunning int32

type superDiskManagerRequest struct {
	Action       string   `json:"action"`
	Categories   []string `json:"categories,omitempty"`
	Confirmation string   `json:"confirmation,omitempty"`
}

type superDiskManagerProxyResponse struct {
	OK          bool            `json:"ok"`
	Disk        json.RawMessage `json:"disk"`
	Categories  json.RawMessage `json:"categories,omitempty"`
	FreedBytes  int64           `json:"freed_bytes,omitempty"`
	Cleaned     []string        `json:"cleaned,omitempty"`
	GeneratedAt string          `json:"generated_at,omitempty"`
	Error       string          `json:"error,omitempty"`
}

// SuperDiskManagerHandler is deliberately a thin authenticated proxy to the
// internal allowlisted disk manager. The browser never sends paths, commands,
// Docker IDs, or a host address.
func SuperDiskManagerHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper); !ok {
			return
		}
		switch r.Method {
		case http.MethodGet:
			response, status := callSuperDiskManager(superDiskManagerRequest{Action: "status"})
			writeJSON(w, status, response)
		case http.MethodPost:
			var request superDiskManagerRequest
			if r.Body == nil || json.NewDecoder(io.LimitReader(r.Body, 8<<10)).Decode(&request) != nil {
				writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "solicitud invalida"})
				return
			}
			if strings.ToLower(strings.TrimSpace(request.Action)) != "cleanup" || strings.TrimSpace(request.Confirmation) != superDiskManagerConfirmation || len(request.Categories) == 0 {
				writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "confirme LIBERAR ESPACIO y seleccione al menos una categoria"})
				return
			}
			if !atomic.CompareAndSwapInt32(&superDiskManagerRunning, 0, 1) {
				writeJSON(w, http.StatusConflict, map[string]interface{}{"ok": false, "error": "ya hay una limpieza en curso"})
				return
			}
			defer atomic.StoreInt32(&superDiskManagerRunning, 0)
			request.Action = "cleanup"
			response, status := callSuperDiskManager(request)
			writeJSON(w, status, response)
		default:
			w.Header().Set("Allow", "GET, POST")
			writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"ok": false, "error": "metodo no permitido"})
		}
	}
}

func callSuperDiskManager(payload superDiskManagerRequest) (superDiskManagerProxyResponse, int) {
	if !superDiskManagerValidCategories(payload.Categories) {
		return superDiskManagerProxyResponse{OK: false, Error: "seleccion de limpieza invalida"}, http.StatusBadRequest
	}
	secret := strings.TrimSpace(os.Getenv("CONFIG_ENC_KEY"))
	if secret == "" {
		return superDiskManagerProxyResponse{OK: false, Error: "el administrador de disco no esta disponible"}, http.StatusServiceUnavailable
	}
	body, err := json.Marshal(superDiskManagerRequest{Action: strings.ToLower(strings.TrimSpace(payload.Action)), Categories: superDiskManagerSortedCategories(payload.Categories)})
	if err != nil {
		return superDiskManagerProxyResponse{OK: false, Error: "no se pudo preparar la solicitud"}, http.StatusInternalServerError
	}
	stamp := time.Now().UTC().Format(time.RFC3339)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(stamp + "\n"))
	_, _ = mac.Write(body)
	request, err := http.NewRequest(http.MethodPost, superDiskManagerURL(), bytes.NewReader(body))
	if err != nil {
		return superDiskManagerProxyResponse{OK: false, Error: "el administrador de disco no esta disponible"}, http.StatusServiceUnavailable
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-PCS-Timestamp", stamp)
	request.Header.Set("X-PCS-Signature", hex.EncodeToString(mac.Sum(nil)))
	client := &http.Client{Timeout: 6 * time.Minute}
	response, err := client.Do(request)
	if err != nil {
		return superDiskManagerProxyResponse{OK: false, Error: "el administrador de disco no responde"}, http.StatusServiceUnavailable
	}
	defer response.Body.Close()
	var result superDiskManagerProxyResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&result); err != nil {
		return superDiskManagerProxyResponse{OK: false, Error: "respuesta invalida del administrador de disco"}, http.StatusBadGateway
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices || !result.OK {
		return superDiskManagerProxyResponse{OK: false, Error: "la operacion de disco no pudo completarse"}, http.StatusBadGateway
	}
	return result, http.StatusOK
}

func superDiskManagerURL() string {
	value := strings.TrimSpace(os.Getenv("PCS_DISK_MANAGER_URL"))
	if value == "" {
		return "http://disk-manager:8083/v1/disk"
	}
	return strings.TrimRight(value, "/") + "/v1/disk"
}

func superDiskManagerValidCategories(categories []string) bool {
	if len(categories) == 0 {
		return true
	}
	allowed := map[string]bool{"containers": true, "images": true, "builder_cache": true, "anonymous_volumes": true}
	seen := map[string]bool{}
	for _, category := range categories {
		category = strings.ToLower(strings.TrimSpace(category))
		if !allowed[category] || seen[category] {
			return false
		}
		seen[category] = true
	}
	return true
}

func superDiskManagerSortedCategories(categories []string) []string {
	result := make([]string, 0, len(categories))
	for _, category := range categories {
		if value := strings.ToLower(strings.TrimSpace(category)); value != "" {
			result = append(result, value)
		}
	}
	sort.Strings(result)
	return result
}
