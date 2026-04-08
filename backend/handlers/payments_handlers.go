package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

// TiposLicenciasHandler placeholder (removed from UI)
func TiposLicenciasHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "tipos_de_licencia API removed", http.StatusNotFound)
	}
}

// LicenciasHandler maneja CRUD de licencias
func LicenciasHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			q := r.URL.Query()
			parseTruthy := func(v string) bool {
				switch strings.ToLower(strings.TrimSpace(v)) {
				case "1", "true", "si", "yes", "activo":
					return true
				default:
					return false
				}
			}

			soloActivas := parseTruthy(q.Get("activo"))
			conEmpresa := parseTruthy(q.Get("con_empresa"))
			usuarioCreador := strings.TrimSpace(q.Get("usuario_creador"))

			// scope=mine permite filtrar por el administrador autenticado sin exponer email en la URL.
			if strings.EqualFold(strings.TrimSpace(q.Get("scope")), "mine") && usuarioCreador == "" {
				c, err := r.Cookie("session_token")
				if err != nil || c == nil || strings.TrimSpace(c.Value) == "" {
					http.Error(w, "unauthenticated", http.StatusUnauthorized)
					return
				}
				s, err := dbpkg.GetSessionByToken(dbSuper, c.Value)
				if err != nil || s == nil {
					http.Error(w, "unauthenticated", http.StatusUnauthorized)
					return
				}
				usuarioCreador = strings.TrimSpace(s.AdminEmail)
			}

			licencias, err := dbpkg.GetLicenciasFiltered(dbSuper, soloActivas, usuarioCreador, conEmpresa)
			if err != nil {
				log.Println("GET /super/api/licencias error:", err)
				http.Error(w, "failed to query licencias: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(licencias)
			return
		case http.MethodPost:
			var payload struct {
				TipoID       int64   `json:"tipo_id"`
				Nombre       string  `json:"nombre"`
				Descripcion  string  `json:"descripcion"`
				Valor        float64 `json:"valor"`
				DuracionDias int     `json:"duracion_dias"`
				ModulosHab   string  `json:"modulos_habilitados"`
				SuperRol     int     `json:"super_rol_habilitado"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}

			log.Printf("POST /super/api/licencias payload: TipoID=%d Nombre=%q", payload.TipoID, payload.Nombre)
			if payload.Nombre == "" {
				http.Error(w, "nombre required", http.StatusBadRequest)
				return
			}
			id, err := dbpkg.CreateLicencia(dbSuper, payload.TipoID, payload.Nombre, payload.Descripcion, payload.Valor, payload.DuracionDias, payload.ModulosHab, payload.SuperRol)
			if err != nil {
				log.Println("POST /super/api/licencias error:", err)
				http.Error(w, "failed to create licencia: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
			return
		case http.MethodPut:
			q := r.URL.Query()
			idStr := q.Get("id")
			if idStr == "" {
				http.Error(w, "id required", http.StatusBadRequest)
				return
			}
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				http.Error(w, "invalid id", http.StatusBadRequest)
				return
			}
			// soporte para acción de activar/desactivar vía query param
			if q.Get("action") == "activar" {
				activoStr := q.Get("activo")
				if activoStr == "" {
					http.Error(w, "activo required (0 or 1)", http.StatusBadRequest)
					return
				}
				act, err := strconv.Atoi(activoStr)
				if err != nil || (act != 0 && act != 1) {
					http.Error(w, "invalid activo value", http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetLicenciaActivo(dbSuper, id, act); err != nil {
					log.Println("ACTIVAR /super/api/licencias error:", err)
					http.Error(w, "failed to set activo: "+err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
			// actualización normal (payload JSON)
			var payloadUpdate struct {
				TipoID       int64   `json:"tipo_id"`
				Nombre       string  `json:"nombre"`
				Descripcion  string  `json:"descripcion"`
				Valor        float64 `json:"valor"`
				DuracionDias int     `json:"duracion_dias"`
				ModulosHab   string  `json:"modulos_habilitados"`
				SuperRol     int     `json:"super_rol_habilitado"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payloadUpdate); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if err := dbpkg.UpdateLicencia(dbSuper, id, payloadUpdate.TipoID, payloadUpdate.Nombre, payloadUpdate.Descripcion, payloadUpdate.Valor, payloadUpdate.DuracionDias, payloadUpdate.ModulosHab, payloadUpdate.SuperRol); err != nil {
				log.Println("PUT /super/api/licencias error:", err)
				http.Error(w, "failed to update licencia: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		case http.MethodDelete:
			q := r.URL.Query()
			idStr := q.Get("id")
			if idStr == "" {
				http.Error(w, "id required", http.StatusBadRequest)
				return
			}
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				http.Error(w, "invalid id", http.StatusBadRequest)
				return
			}
			if err := dbpkg.DeleteLicencia(dbSuper, id); err != nil {
				log.Println("DELETE /super/api/licencias error:", err)
				http.Error(w, "failed to delete licencia: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

func getDecryptedConfigValue(dbSuper *sql.DB, key string) (string, error) {
	v, enc, err := dbpkg.GetConfigValue(dbSuper, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	if v == "" {
		return "", nil
	}
	if !enc {
		return v, nil
	}
	dec, derr := utils.DecryptString(v)
	if derr != nil {
		return "", derr
	}
	return dec, nil
}

func isApprovedPaymentStatus(status string) bool {
	status = strings.ToLower(strings.TrimSpace(status))
	return status == "approved" || status == "accredited"
}

func activateLicenciaByIDs(dbSuper *sql.DB, licenciaID, empresaID int64) (bool, error) {
	if licenciaID <= 0 || empresaID <= 0 {
		return false, nil
	}
	lic, err := dbpkg.GetLicenciaByID(dbSuper, licenciaID)
	if err != nil {
		return false, err
	}
	if lic == nil {
		return false, nil
	}
	now := time.Now()
	fechaInicio := now.Format("2006-01-02 15:04:05")
	fechaFin := now.AddDate(0, 0, lic.DuracionDias).Format("2006-01-02 15:04:05")
	if err := dbpkg.ActivateLicenciaForEmpresa(dbSuper, licenciaID, empresaID, fechaInicio, fechaFin); err != nil {
		return false, err
	}
	return true, nil
}

func extractWompiWebhookPaymentInfo(obj map[string]interface{}) (string, string, string) {
	get := func(v interface{}) string {
		s := strings.TrimSpace(fmt.Sprint(v))
		if s == "<nil>" {
			return ""
		}
		return s
	}

	var transactionID, reference, status string
	data, _ := obj["data"].(map[string]interface{})
	if tx, ok := data["transaction"].(map[string]interface{}); ok {
		transactionID = get(tx["id"])
		reference = get(tx["reference"])
		status = get(tx["status"])
	}
	if transactionID == "" {
		transactionID = get(data["id"])
	}
	if reference == "" {
		reference = get(data["reference"])
	}
	if status == "" {
		status = get(data["status"])
	}
	if transactionID == "" {
		transactionID = get(obj["transaction_id"])
	}
	if reference == "" {
		reference = get(obj["reference"])
	}
	if status == "" {
		status = get(obj["status"])
	}

	status = strings.ToUpper(strings.TrimSpace(status))
	return transactionID, reference, status
}

func parseSignatureCandidates(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	seen := map[string]struct{}{}
	out := make([]string, 0)
	add := func(v string) {
		v = strings.TrimSpace(v)
		v = strings.Trim(v, `"`)
		if v == "" {
			return
		}
		key := strings.ToLower(v)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, v)
	}

	add(raw)
	parts := strings.Split(raw, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "=") {
			split := strings.SplitN(part, "=", 2)
			add(split[1])
			continue
		}
		add(part)
	}

	return out
}

func signatureMatch(candidate, expected string) bool {
	left := []byte(strings.ToLower(strings.TrimSpace(candidate)))
	right := []byte(strings.ToLower(strings.TrimSpace(expected)))
	if len(left) == 0 || len(right) == 0 || len(left) != len(right) {
		return false
	}
	return subtle.ConstantTimeCompare(left, right) == 1
}

func verifyWompiWebhookSignature(dbSuper *sql.DB, r *http.Request, body []byte, obj map[string]interface{}) error {
	integrityKey, err := getDecryptedConfigValue(dbSuper, "wompi.integrity_key")
	if err != nil {
		return err
	}
	integrityKey = strings.TrimSpace(integrityKey)
	if integrityKey == "" {
		return nil
	}

	rawSignature := ""
	headerKeys := []string{"X-Wompi-Signature", "X-Event-Checksum", "X-Signature"}
	for _, hk := range headerKeys {
		if v := strings.TrimSpace(r.Header.Get(hk)); v != "" {
			rawSignature = v
			break
		}
	}
	if rawSignature == "" {
		if sigObj, ok := obj["signature"].(map[string]interface{}); ok {
			rawSignature = strings.TrimSpace(fmt.Sprint(sigObj["checksum"]))
			if rawSignature == "" || rawSignature == "<nil>" {
				rawSignature = strings.TrimSpace(fmt.Sprint(sigObj["signature"]))
			}
		}
	}
	if rawSignature == "" || rawSignature == "<nil>" {
		log.Println("warning: wompi webhook received without signature checksum; skipping strict validation")
		return nil
	}

	candidates := parseSignatureCandidates(rawSignature)
	if len(candidates) == 0 {
		return errors.New("invalid wompi signature format")
	}

	h := hmac.New(sha256.New, []byte(integrityKey))
	h.Write(body)
	hmacHex := hex.EncodeToString(h.Sum(nil))
	hmacB64 := base64.StdEncoding.EncodeToString(h.Sum(nil))

	shaBodyPlus := sha256.Sum256(append(append([]byte{}, body...), []byte(integrityKey)...))
	shaKeyPlus := sha256.Sum256(append([]byte(integrityKey), body...))
	bodyHex := hex.EncodeToString(shaBodyPlus[:])
	keyHex := hex.EncodeToString(shaKeyPlus[:])

	for _, candidate := range candidates {
		if signatureMatch(candidate, hmacHex) || signatureMatch(candidate, hmacB64) || signatureMatch(candidate, bodyHex) || signatureMatch(candidate, keyHex) {
			return nil
		}
	}

	return errors.New("invalid wompi signature")
}

func normalizeWompiMode(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "sandbox", "test", "testing", "sambox", "pruebas":
		return "sandbox"
	case "production", "prod", "live", "real", "reales":
		return "production"
	default:
		return ""
	}
}

func wompiModeFromKeys(publicKey, privateKey string) string {
	if strings.HasPrefix(privateKey, "prv_test_") || strings.HasPrefix(publicKey, "pub_test_") {
		return "sandbox"
	}
	if strings.TrimSpace(publicKey) != "" || strings.TrimSpace(privateKey) != "" {
		return "production"
	}
	return ""
}

func resolveWompiMode(dbSuper *sql.DB, publicKey, privateKey string) (string, string) {
	if configuredMode, _, err := dbpkg.GetConfigValue(dbSuper, "wompi.mode"); err == nil {
		if normalized := normalizeWompiMode(configuredMode); normalized != "" {
			return normalized, "manual"
		}
	}
	if inferred := wompiModeFromKeys(publicKey, privateKey); inferred != "" {
		return inferred, "keys"
	}
	return "sandbox", "default"
}

func wompiBaseURLFromMode(mode string) string {
	if normalizeWompiMode(mode) == "sandbox" {
		return "https://sandbox.wompi.co/v1"
	}
	return "https://production.wompi.co/v1"
}

func fetchWompiAcceptanceInfo(baseURL, publicKey string) (string, string, string, string, error) {
	if strings.TrimSpace(publicKey) == "" {
		return "", "", "", "", fmt.Errorf("wompi.public_key no configurada")
	}
	merchantURL := strings.TrimRight(baseURL, "/") + "/merchants/" + url.PathEscape(publicKey)
	req, err := http.NewRequest("GET", merchantURL, nil)
	if err != nil {
		return "", "", "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+publicKey)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", "", "", "", fmt.Errorf("wompi merchants error %s: %s", resp.Status, string(body))
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(body, &obj); err != nil {
		return "", "", "", "", err
	}
	data, _ := obj["data"].(map[string]interface{})
	presignedAcceptance, _ := data["presigned_acceptance"].(map[string]interface{})
	presignedPersonal, _ := data["presigned_personal_data_auth"].(map[string]interface{})
	acceptanceToken := strings.TrimSpace(fmt.Sprint(presignedAcceptance["acceptance_token"]))
	personalToken := strings.TrimSpace(fmt.Sprint(presignedPersonal["acceptance_token"]))
	acceptancePermalink := strings.TrimSpace(fmt.Sprint(presignedAcceptance["permalink"]))
	personalPermalink := strings.TrimSpace(fmt.Sprint(presignedPersonal["permalink"]))
	if acceptanceToken == "" || acceptanceToken == "<nil>" {
		acceptanceToken = ""
	}
	if personalToken == "" || personalToken == "<nil>" {
		personalToken = ""
	}
	if acceptancePermalink == "<nil>" {
		acceptancePermalink = ""
	}
	if personalPermalink == "<nil>" {
		personalPermalink = ""
	}
	return acceptanceToken, personalToken, acceptancePermalink, personalPermalink, nil
}

// WompiConfigHandler gestiona credenciales de Wompi para pagos alternativos con Nequi.
func WompiConfigHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			pub, _, _, pubUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "wompi.public_key")
			prv, prvEnc, _, prvUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "wompi.private_key")
			integrity, intEnc, _, intUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "wompi.integrity_key")
			modeRaw, _, _, modeUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "wompi.mode")

			pubSet := pub != ""
			prvSet := prv != ""
			intSet := integrity != ""

			pubMasked := ""
			if pubSet {
				if len(pub) > 16 {
					pubMasked = pub[:8] + "..." + pub[len(pub)-6:]
				} else {
					pubMasked = pub
				}
			}

			prvMasked := ""
			if prvSet {
				if prvEnc {
					prvMasked = "********"
				} else if len(prv) > 10 {
					prvMasked = prv[:4] + "****" + prv[len(prv)-4:]
				} else {
					prvMasked = "****"
				}
			}

			integrityMasked := ""
			if intSet {
				if intEnc {
					integrityMasked = "********"
				} else if len(integrity) > 10 {
					integrityMasked = integrity[:4] + "****" + integrity[len(integrity)-4:]
				} else {
					integrityMasked = "****"
				}
			}

			configuredMode := normalizeWompiMode(modeRaw)
			mode := configuredMode
			modeSource := "manual"
			if mode == "" {
				mode = wompiModeFromKeys(pub, prv)
				if mode != "" {
					modeSource = "keys"
				} else {
					mode = "sandbox"
					modeSource = "default"
				}
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"public_key_set":        pubSet,
				"public_key_masked":     pubMasked,
				"public_key_updated":    pubUpdated,
				"private_key_set":       prvSet,
				"private_key_masked":    prvMasked,
				"private_key_updated":   prvUpdated,
				"integrity_key_set":     intSet,
				"integrity_key_masked":  integrityMasked,
				"integrity_key_updated": intUpdated,
				"encryption_available":  utils.EncryptionAvailable(),
				"mode":                  mode,
				"mode_set":              configuredMode != "",
				"mode_source":           modeSource,
				"mode_updated":          modeUpdated,
			})
			return

		case http.MethodPost, http.MethodPut:
			var payload struct {
				PublicKey    string `json:"public_key"`
				PrivateKey   string `json:"private_key"`
				IntegrityKey string `json:"integrity_key"`
				Mode         string `json:"mode"`
				Encrypt      bool   `json:"encrypt"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
				return
			}
			modeInput := strings.TrimSpace(payload.Mode)
			normalizedMode := normalizeWompiMode(modeInput)
			if modeInput != "" && normalizedMode == "" {
				http.Error(w, "mode inválido: usa sandbox o real", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.PublicKey) == "" && strings.TrimSpace(payload.PrivateKey) == "" && strings.TrimSpace(payload.IntegrityKey) == "" && normalizedMode == "" {
				http.Error(w, "at least one value is required (mode o llaves)", http.StatusBadRequest)
				return
			}

			if payload.PublicKey != "" && !strings.HasPrefix(payload.PublicKey, "pub_") {
				http.Error(w, "public_key inválida: debe iniciar con pub_", http.StatusBadRequest)
				return
			}
			if payload.PrivateKey != "" && !strings.HasPrefix(payload.PrivateKey, "prv_") {
				http.Error(w, "private_key inválida: debe iniciar con prv_", http.StatusBadRequest)
				return
			}
			if payload.IntegrityKey != "" && !strings.Contains(payload.IntegrityKey, "integrity") {
				http.Error(w, "integrity_key inválida: prefijo esperado *_integrity_*", http.StatusBadRequest)
				return
			}

			// Requerir cifrado obligatorio para llaves sensibles.
			if payload.PublicKey != "" {
				if err := dbpkg.SetConfigValue(dbSuper, "wompi.public_key", payload.PublicKey, false); err != nil {
					http.Error(w, "failed to save wompi.public_key: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			saveSensitive := func(key, value string) error {
				if value == "" {
					return nil
				}
				if !utils.EncryptionAvailable() {
					return fmt.Errorf("encryption required: CONFIG_ENC_KEY not set")
				}
				encVal, err := utils.EncryptString(value)
				if err != nil {
					return err
				}
				return dbpkg.SetConfigValue(dbSuper, key, encVal, true)
			}

			if err := saveSensitive("wompi.private_key", payload.PrivateKey); err != nil {
				http.Error(w, "failed to save wompi.private_key: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if err := saveSensitive("wompi.integrity_key", payload.IntegrityKey); err != nil {
				http.Error(w, "failed to save wompi.integrity_key: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if normalizedMode != "" {
				if err := dbpkg.SetConfigValue(dbSuper, "wompi.mode", normalizedMode, false); err != nil {
					http.Error(w, "failed to save wompi.mode: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"saved": true, "mode": normalizedMode})
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

// WompiTermsHandler devuelve links de términos y autorizaciones para cumplimiento de aceptación.
func WompiTermsHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		publicKey, err := getDecryptedConfigValue(dbSuper, "wompi.public_key")
		if err != nil {
			http.Error(w, "failed to read wompi.public_key: "+err.Error(), http.StatusInternalServerError)
			return
		}
		privateKey, _ := getDecryptedConfigValue(dbSuper, "wompi.private_key")
		if strings.TrimSpace(publicKey) == "" {
			http.Error(w, "wompi.public_key not configured", http.StatusInternalServerError)
			return
		}
		mode, modeSource := resolveWompiMode(dbSuper, publicKey, privateKey)
		baseURL := wompiBaseURLFromMode(mode)
		_, _, acceptancePermalink, personalPermalink, ferr := fetchWompiAcceptanceInfo(baseURL, publicKey)
		if ferr != nil {
			http.Error(w, "failed to fetch acceptance tokens: "+ferr.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"provider":                    "wompi",
			"payment_method":              "NEQUI",
			"mode":                        mode,
			"mode_source":                 modeSource,
			"api_base_url":                baseURL,
			"acceptance_permalink":        acceptancePermalink,
			"personal_data_permalink":     personalPermalink,
			"sandbox_phone_approved":      "3991111111",
			"sandbox_phone_declined":      "3992222222",
			"sandbox_phone_error_example": "3993333333",
		})
	}
}

// WompiCreateNequiTransactionHandler crea una transacción Wompi usando método de pago NEQUI.
func WompiCreateNequiTransactionHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			LicenciaID    int64  `json:"licencia_id"`
			EmpresaID     int64  `json:"empresa_id,omitempty"`
			PhoneNumber   string `json:"phone_number"`
			CustomerEmail string `json:"customer_email,omitempty"`
			AcceptTerms   bool   `json:"accept_terms"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
			return
		}

		if payload.LicenciaID <= 0 {
			http.Error(w, "licencia_id inválido", http.StatusBadRequest)
			return
		}
		phone := strings.TrimSpace(payload.PhoneNumber)
		if ok, _ := regexp.MatchString(`^3\d{9}$`, phone); !ok {
			http.Error(w, "phone_number inválido: usa 10 dígitos colombianos (ej. 3991111111 en sandbox)", http.StatusBadRequest)
			return
		}

		publicKey, err := getDecryptedConfigValue(dbSuper, "wompi.public_key")
		if err != nil {
			http.Error(w, "failed to read wompi.public_key: "+err.Error(), http.StatusInternalServerError)
			return
		}
		privateKey, err := getDecryptedConfigValue(dbSuper, "wompi.private_key")
		if err != nil {
			http.Error(w, "failed to read wompi.private_key: "+err.Error(), http.StatusInternalServerError)
			return
		}
		integrityKey, err := getDecryptedConfigValue(dbSuper, "wompi.integrity_key")
		if err != nil {
			http.Error(w, "failed to read wompi.integrity_key: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if strings.TrimSpace(publicKey) == "" || strings.TrimSpace(privateKey) == "" || strings.TrimSpace(integrityKey) == "" {
			http.Error(w, "Wompi no configurado: faltan llaves (public/private/integrity)", http.StatusInternalServerError)
			return
		}

		lic, err := dbpkg.GetLicenciaByID(dbSuper, payload.LicenciaID)
		if err != nil || lic == nil {
			http.Error(w, "licencia not found", http.StatusBadRequest)
			return
		}

		amountInCents := int64(math.Round(lic.Valor * 100))
		if amountInCents <= 0 {
			http.Error(w, "valor de licencia inválido para Wompi", http.StatusBadRequest)
			return
		}

		mode, _ := resolveWompiMode(dbSuper, publicKey, privateKey)
		baseURL := wompiBaseURLFromMode(mode)
		acceptanceToken, personalToken, acceptancePermalink, personalPermalink, ferr := fetchWompiAcceptanceInfo(baseURL, publicKey)
		if ferr != nil {
			http.Error(w, "failed to fetch Wompi acceptance data: "+ferr.Error(), http.StatusBadGateway)
			return
		}

		if !payload.AcceptTerms {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":                   "Debes aceptar términos y autorización de datos para continuar con Nequi",
				"acceptance_permalink":    acceptancePermalink,
				"personal_data_permalink": personalPermalink,
			})
			return
		}

		if acceptanceToken == "" || personalToken == "" {
			http.Error(w, "Wompi no devolvió tokens de aceptación válidos", http.StatusBadGateway)
			return
		}

		email := strings.TrimSpace(payload.CustomerEmail)
		if email == "" {
			email = strings.TrimSpace(r.Header.Get("X-Admin-Email"))
		}
		if email == "" {
			http.Error(w, "customer_email requerido para crear la transacción", http.StatusBadRequest)
			return
		}

		reference := fmt.Sprintf("WOMPI-LIC-%d-EMP-%d-%d", payload.LicenciaID, payload.EmpresaID, time.Now().UnixNano())
		signatureSource := fmt.Sprintf("%s%dCOP%s", reference, amountInCents, integrityKey)
		signatureHash := sha256.Sum256([]byte(signatureSource))
		signature := hex.EncodeToString(signatureHash[:])

		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		redirectURL := fmt.Sprintf("%s://%s/pagar_licencia.html?status=pending&provider=nequi", scheme, r.Host)

		reqBody := map[string]interface{}{
			"acceptance_token":     acceptanceToken,
			"accept_personal_auth": personalToken,
			"amount_in_cents":      amountInCents,
			"currency":             "COP",
			"customer_email":       email,
			"reference":            reference,
			"signature":            signature,
			"redirect_url":         redirectURL,
			"payment_method": map[string]interface{}{
				"type":         "NEQUI",
				"phone_number": phone,
			},
		}

		bodyBytes, _ := json.Marshal(reqBody)
		apiURL := strings.TrimRight(baseURL, "/") + "/transactions"
		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(bodyBytes))
		if err != nil {
			http.Error(w, "failed to create request: "+err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+privateKey)

		client := &http.Client{Timeout: 20 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "request error: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		if resp.StatusCode >= 400 {
			log.Println("Wompi API error:", resp.Status, string(respBody))
			http.Error(w, "wompi API error: "+string(respBody), http.StatusBadGateway)
			return
		}

		var wompiResp map[string]interface{}
		if err := json.Unmarshal(respBody, &wompiResp); err != nil {
			http.Error(w, "invalid response from wompi: "+err.Error(), http.StatusInternalServerError)
			return
		}

		data, _ := wompiResp["data"].(map[string]interface{})
		transactionID := strings.TrimSpace(fmt.Sprint(data["id"]))
		status := strings.ToUpper(strings.TrimSpace(fmt.Sprint(data["status"])))
		respReference := strings.TrimSpace(fmt.Sprint(data["reference"]))
		if transactionID == "" || transactionID == "<nil>" {
			http.Error(w, "wompi response sin transaction id", http.StatusBadGateway)
			return
		}
		if status == "" || status == "<nil>" {
			status = "PENDING"
		}
		if respReference == "" || respReference == "<nil>" {
			respReference = reference
		}

		if _, err := dbpkg.CreateWompiPaymentRecord(dbSuper, payload.LicenciaID, payload.EmpresaID, transactionID, respReference, status, string(respBody)); err != nil {
			log.Println("warning: failed to record Wompi transaction in DB:", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"provider":                "wompi",
			"payment_method":          "NEQUI",
			"mode":                    mode,
			"transaction_id":          transactionID,
			"reference":               respReference,
			"status":                  status,
			"acceptance_permalink":    acceptancePermalink,
			"personal_data_permalink": personalPermalink,
			"data":                    data,
		})
	}
}

// WompiTransactionStatusHandler consulta estado de la transacción y activa licencia si quedó APPROVED.
func WompiTransactionStatusHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		transactionID := strings.TrimSpace(r.URL.Query().Get("id"))
		if transactionID == "" {
			transactionID = strings.TrimSpace(r.URL.Query().Get("transaction_id"))
		}
		if transactionID == "" {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}

		publicKey, err := getDecryptedConfigValue(dbSuper, "wompi.public_key")
		if err != nil {
			http.Error(w, "failed to read wompi.public_key: "+err.Error(), http.StatusInternalServerError)
			return
		}
		privateKey, err := getDecryptedConfigValue(dbSuper, "wompi.private_key")
		if err != nil {
			http.Error(w, "failed to read wompi.private_key: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if strings.TrimSpace(publicKey) == "" {
			http.Error(w, "wompi.public_key not configured", http.StatusInternalServerError)
			return
		}

		mode, _ := resolveWompiMode(dbSuper, publicKey, privateKey)
		baseURL := wompiBaseURLFromMode(mode)
		statusURL := strings.TrimRight(baseURL, "/") + "/transactions/" + url.PathEscape(transactionID)

		fetchStatus := func(authKey string) ([]byte, int, error) {
			req, err := http.NewRequest("GET", statusURL, nil)
			if err != nil {
				return nil, 0, err
			}
			req.Header.Set("Authorization", "Bearer "+authKey)
			client := &http.Client{Timeout: 15 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				return nil, 0, err
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			return body, resp.StatusCode, nil
		}

		respBody, statusCode, err := fetchStatus(publicKey)
		if err != nil {
			http.Error(w, "request error: "+err.Error(), http.StatusBadGateway)
			return
		}
		if statusCode >= 400 && strings.TrimSpace(privateKey) != "" {
			if body2, code2, err2 := fetchStatus(privateKey); err2 == nil {
				respBody = body2
				statusCode = code2
			}
		}
		if statusCode >= 400 {
			http.Error(w, "wompi API error: "+string(respBody), http.StatusBadGateway)
			return
		}

		var wompiResp map[string]interface{}
		if err := json.Unmarshal(respBody, &wompiResp); err != nil {
			http.Error(w, "invalid response from wompi: "+err.Error(), http.StatusInternalServerError)
			return
		}
		data, _ := wompiResp["data"].(map[string]interface{})
		status := strings.ToUpper(strings.TrimSpace(fmt.Sprint(data["status"])))
		reference := strings.TrimSpace(fmt.Sprint(data["reference"]))

		if err := dbpkg.UpdateWompiPaymentRecordByTransaction(dbSuper, transactionID, status, string(respBody)); err != nil {
			log.Println("warning: failed to update Wompi payment record:", err)
		}

		var licenciaID sql.NullInt64
		var empresaID sql.NullInt64
		row := dbSuper.QueryRow("SELECT licencia_id, empresa_id FROM pagos_wompi WHERE transaction_id = ? LIMIT 1", transactionID)
		if err := row.Scan(&licenciaID, &empresaID); err != nil {
			log.Println("warning: pagos_wompi record not found for tx:", transactionID, "err:", err)
		}

		if strings.EqualFold(status, "APPROVED") && licenciaID.Valid && empresaID.Valid {
			lic, err := dbpkg.GetLicenciaByID(dbSuper, licenciaID.Int64)
			if err == nil && lic != nil {
				now := time.Now()
				fechaInicio := now.Format("2006-01-02 15:04:05")
				fechaFin := now.AddDate(0, 0, lic.DuracionDias).Format("2006-01-02 15:04:05")
				if err := dbpkg.ActivateLicenciaForEmpresa(dbSuper, licenciaID.Int64, empresaID.Int64, fechaInicio, fechaFin); err != nil {
					log.Println("failed to activate licencia from Wompi:", err)
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"provider":       "wompi",
			"mode":           mode,
			"transaction_id": transactionID,
			"reference":      reference,
			"status":         status,
			"data":           data,
		})
	}
}

// WompiWebhookHandler procesa notificaciones servidor-servidor de Wompi.
func WompiWebhookHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		var obj map[string]interface{}
		if err := json.Unmarshal(body, &obj); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		if err := verifyWompiWebhookSignature(dbSuper, r, body, obj); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		transactionID, reference, status := extractWompiWebhookPaymentInfo(obj)
		if strings.TrimSpace(transactionID) == "" && strings.TrimSpace(reference) == "" {
			http.Error(w, "transaction_id or reference required", http.StatusBadRequest)
			return
		}
		if status == "" {
			status = "PENDING"
		}

		if transactionID != "" {
			if err := dbpkg.UpdateWompiPaymentRecordByTransaction(dbSuper, transactionID, status, string(body)); err != nil {
				log.Println("warning: failed to update Wompi record by transaction_id:", err)
			}
		}
		if reference != "" {
			if err := dbpkg.UpdateWompiPaymentRecordByReference(dbSuper, reference, status, string(body)); err != nil {
				log.Println("warning: failed to update Wompi record by reference:", err)
			}
		}

		licenciaID, empresaID, hasContext, ctxErr := dbpkg.GetWompiPaymentContext(dbSuper, transactionID, reference)
		if ctxErr != nil {
			log.Println("warning: failed to resolve Wompi payment context:", ctxErr)
		}

		activated := false
		if isApprovedPaymentStatus(status) && hasContext {
			act, actErr := activateLicenciaByIDs(dbSuper, licenciaID, empresaID)
			if actErr != nil {
				log.Println("warning: failed to activate licencia from Wompi webhook:", actErr)
			} else {
				activated = act
			}
		}

		ventaDigitalContextFound := false
		ventaDigitalDeliverySent := false
		ventaDigitalDeliveryStage := "not_processed"
		if strings.TrimSpace(status) != "" {
			foundVD, deliveredVD, deliveryStageVD, vdErr := processVentaDigitalPaymentStatusUpdate(r, dbSuper, transactionID, reference, status, string(body))
			ventaDigitalContextFound = foundVD
			if strings.TrimSpace(deliveryStageVD) != "" {
				ventaDigitalDeliveryStage = deliveryStageVD
			}
			if vdErr != nil {
				log.Println("warning: failed to process venta_digital webhook update:", vdErr)
			} else {
				ventaDigitalDeliverySent = deliveredVD
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":                           true,
			"provider":                     "wompi",
			"transaction_id":               transactionID,
			"reference":                    reference,
			"status":                       status,
			"context_found":                hasContext,
			"licencia_id":                  licenciaID,
			"empresa_id":                   empresaID,
			"activated":                    activated,
			"venta_digital_context_found":  ventaDigitalContextFound,
			"venta_digital_delivery_sent":  ventaDigitalDeliverySent,
			"venta_digital_delivery_stage": ventaDigitalDeliveryStage,
		})
	}
}

// ActivateLicenciaSinPagoHandler activa una licencia manualmente para avanzar en pruebas internas.
func ActivateLicenciaSinPagoHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			LicenciaID int64  `json:"licencia_id"`
			EmpresaID  int64  `json:"empresa_id"`
			Motivo     string `json:"motivo,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
			return
		}
		if payload.LicenciaID <= 0 {
			http.Error(w, "licencia_id inválido", http.StatusBadRequest)
			return
		}
		if payload.EmpresaID <= 0 {
			http.Error(w, "empresa_id inválido", http.StatusBadRequest)
			return
		}

		lic, err := dbpkg.GetLicenciaByID(dbSuper, payload.LicenciaID)
		if err != nil || lic == nil {
			http.Error(w, "licencia not found", http.StatusBadRequest)
			return
		}

		now := time.Now()
		fechaInicio := now.Format("2006-01-02 15:04:05")
		fechaFin := now.AddDate(0, 0, lic.DuracionDias).Format("2006-01-02 15:04:05")
		if err := dbpkg.ActivateLicenciaForEmpresa(dbSuper, payload.LicenciaID, payload.EmpresaID, fechaInicio, fechaFin); err != nil {
			http.Error(w, "failed to activate licencia: "+err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("Licencia activada sin pago: licencia=%d empresa=%d motivo=%q", payload.LicenciaID, payload.EmpresaID, payload.Motivo)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"activated":      true,
			"provider":       "manual",
			"payment_method": "ACTIVAR_SIN_PAGO",
			"licencia_id":    payload.LicenciaID,
			"empresa_id":     payload.EmpresaID,
			"fecha_inicio":   fechaInicio,
			"fecha_fin":      fechaFin,
			"redirect_url":   fmt.Sprintf("/administrar_empresa.html?id=%d", payload.EmpresaID),
		})
	}
}
