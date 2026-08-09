package handlers

import (
	"bytes"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

//go:embed templates/instalar_domotica_raspberry.sh.tmpl
var domoticaRaspberryInstallerTemplate string

var (
	domoticaTunnelSchemaOnce sync.Once
	domoticaTunnelSchemaErr  error
)

type domoticaInstallerTemplateData struct {
	BaseURLJSON         string
	DeviceUIDJSON       string
	EnrollmentTokenJSON string
}

func ensureDomoticaTunnelSchemaReady(dbEmp *sql.DB) error {
	domoticaTunnelSchemaOnce.Do(func() {
		domoticaTunnelSchemaErr = dbpkg.EmpresaControlElectricoTunnelSchemaReady(dbEmp)
	})
	return domoticaTunnelSchemaErr
}

func handleDomoticaRaspberryInstallerDownload(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, empresaID int64, actor string) error {
	if err := ensureDomoticaTunnelSchemaReady(dbEmp); err != nil {
		return err
	}
	baseURL, err := domoticaTunnelPublicBaseURL()
	if err != nil {
		return err
	}
	r.Body = http.MaxBytesReader(w, r.Body, 32*1024)
	var payload struct {
		RaspberryID int64 `json:"raspberry_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.RaspberryID <= 0 {
		return errors.New("raspberry_id requerido")
	}
	pi, enrollmentToken, err := dbpkg.ProvisionEmpresaControlElectricoRaspberryTunnel(dbEmp, empresaID, payload.RaspberryID, actor)
	if err != nil {
		return err
	}
	baseJSON, _ := json.Marshal(baseURL)
	deviceJSON, _ := json.Marshal(pi.DeviceUID)
	tokenJSON, _ := json.Marshal(enrollmentToken)
	tmpl, err := template.New("installer").Parse(domoticaRaspberryInstallerTemplate)
	if err != nil {
		return err
	}
	var body bytes.Buffer
	if err := tmpl.Execute(&body, domoticaInstallerTemplateData{BaseURLJSON: string(baseJSON), DeviceUIDJSON: string(deviceJSON), EnrollmentTokenJSON: string(tokenJSON)}); err != nil {
		return err
	}
	_, _ = dbpkg.InsertEmpresaControlElectricoEvento(dbEmp, dbpkg.EmpresaControlElectricoEvento{
		EmpresaID:   empresaID,
		RaspberryID: pi.ID,
		Comando:     "provisionar_tunel",
		Resultado:   "instalador_generado",
		Actor:       actor,
		Origen:      "panel_domotica",
	})
	filenameCode := sanitizeDomoticaASCII(firstNonEmpty(pi.Codigo, pi.Nombre, "raspberry"))
	if filenameCode == "" {
		filenameCode = "raspberry"
	}
	w.Header().Set("Content-Type", "text/x-shellscript; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="instalar-pcs-domotica-`+filenameCode+`.sh"`)
	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body.Bytes())
	return err
}

func domoticaTunnelPublicBaseURL() (string, error) {
	raw := ""
	for _, key := range []string{"PCS_DOMOTICA_PUBLIC_BASE_URL", "PCS_PUBLIC_BASE_URL", "PUBLIC_BASE_URL"} {
		if raw = strings.TrimSpace(os.Getenv(key)); raw != "" {
			break
		}
	}
	if raw == "" {
		raw = "https://powerfulcontrolsystem.com"
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Hostname() == "" || parsed.User != nil {
		return "", errors.New("URL publica de domotica invalida")
	}
	if parsed.Scheme != "https" && !(parsed.Scheme == "http" && (parsed.Hostname() == "localhost" || net.ParseIP(parsed.Hostname()).IsLoopback())) {
		return "", errors.New("el tunel de domotica exige HTTPS")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return strings.TrimRight(parsed.String(), "/"), nil
}

// PublicDomoticaRaspberryTunnelHandler atiende exclusivamente dispositivos
// previamente aprovisionados. La empresa se resuelve desde la identidad
// criptografica del dispositivo, nunca desde empresa_id enviado por la placa.
func PublicDomoticaRaspberryTunnelHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := ensureDomoticaTunnelSchemaReady(dbEmp); err != nil {
			log.Printf("[domotica_tunnel] schema readiness error: %v", err)
			http.Error(w, "servicio no disponible", http.StatusServiceUnavailable)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 256*1024)
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		requestBytes := r.ContentLength
		if requestBytes < 0 {
			requestBytes = 0
		}
		if action == "enroll" {
			handleDomoticaTunnelEnroll(w, r, dbEmp, requestBytes)
			return
		}
		device, err := authenticateDomoticaTunnelRequest(dbEmp, r)
		if err != nil {
			writeDomoticaTunnelJSON(w, http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "dispositivo no autorizado"})
			return
		}
		switch action {
		case "poll":
			handleDomoticaTunnelPoll(w, r, dbEmp, device, requestBytes)
		case "ack":
			handleDomoticaTunnelAck(w, r, dbEmp, device, requestBytes)
		case "input":
			handleDomoticaTunnelInput(w, r, dbEmp, device, requestBytes)
		case "telemetry":
			handleDomoticaTunnelTelemetry(w, r, dbEmp, device, requestBytes)
		default:
			writeDomoticaTunnelJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "action no soportada"})
		}
	}
}

func handleDomoticaTunnelEnroll(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, requestBytes int64) {
	var payload struct {
		DeviceUID       string `json:"device_uid"`
		EnrollmentToken string `json:"enrollment_token"`
		Hostname        string `json:"hostname"`
		LocalIP         string `json:"local_ip"`
		AgentVersion    string `json:"agent_version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeDomoticaTunnelJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "json invalido"})
		return
	}
	device, deviceToken, err := dbpkg.EnrollEmpresaControlElectricoRaspberryTunnel(dbEmp, payload.DeviceUID, payload.EnrollmentToken)
	if err != nil {
		writeDomoticaTunnelJSON(w, http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "aprovisionamiento invalido o vencido"})
		return
	}
	response := map[string]interface{}{"ok": true, "device_token": deviceToken, "poll_seconds": 20, "server_time": time.Now().UTC().Format(time.RFC3339)}
	body := marshalDomoticaTunnelJSON(response)
	_ = dbpkg.RecordEmpresaControlElectricoTunnelTraffic(dbEmp, device, requestBytes, int64(len(body)), domoticaTunnelRemoteIP(r), payload.AgentVersion, "")
	writeDomoticaTunnelJSONBytes(w, http.StatusOK, body)
}

func authenticateDomoticaTunnelRequest(dbEmp *sql.DB, r *http.Request) (*dbpkg.EmpresaControlElectricoTunnelDevice, error) {
	deviceUID := strings.TrimSpace(r.Header.Get("X-PCS-Device-ID"))
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(auth) < 8 || !strings.EqualFold(auth[:7], "Bearer ") {
		return nil, sql.ErrNoRows
	}
	return dbpkg.AuthenticateEmpresaControlElectricoRaspberryTunnel(dbEmp, deviceUID, strings.TrimSpace(auth[7:]))
}

func handleDomoticaTunnelPoll(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, device *dbpkg.EmpresaControlElectricoTunnelDevice, requestBytes int64) {
	var payload struct {
		Hostname     string `json:"hostname"`
		LocalIP      string `json:"local_ip"`
		AgentVersion string `json:"agent_version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeDomoticaTunnelJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "json invalido"})
		return
	}
	deadline := time.Now().Add(20 * time.Second)
	var command *dbpkg.EmpresaControlElectricoTunnelCommand
	for time.Now().Before(deadline) {
		claimed, err := dbpkg.ClaimEmpresaControlElectricoTunnelCommand(dbEmp, device.EmpresaID, device.RaspberryID)
		if err == nil {
			command = claimed
			break
		}
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("[domotica_tunnel] claim empresa_id=%d raspberry_id=%d error: %v", device.EmpresaID, device.RaspberryID, err)
			break
		}
		select {
		case <-r.Context().Done():
			return
		case <-time.After(400 * time.Millisecond):
		}
	}
	inputs, err := dbpkg.ListEmpresaControlElectricoInputConfigs(dbEmp, device.EmpresaID, device.RaspberryID)
	if err != nil {
		inputs = []dbpkg.EmpresaControlElectricoInputConfig{}
	}
	response := map[string]interface{}{"ok": true, "inputs": inputs, "server_time": time.Now().UTC().Format(time.RFC3339)}
	if command != nil {
		var commandPayload map[string]interface{}
		if err := json.Unmarshal([]byte(command.PayloadJSON), &commandPayload); err == nil {
			commandPayload["command_id"] = command.CommandUID
			response["command"] = commandPayload
		}
	}
	body := marshalDomoticaTunnelJSON(response)
	remoteIP := firstNonEmpty(strings.TrimSpace(payload.LocalIP), domoticaTunnelRemoteIP(r))
	if err := dbpkg.RecordEmpresaControlElectricoTunnelTraffic(dbEmp, device, requestBytes, int64(len(body)), remoteIP, payload.AgentVersion, ""); err != nil {
		log.Printf("[domotica_tunnel] traffic empresa_id=%d raspberry_id=%d error: %v", device.EmpresaID, device.RaspberryID, err)
	}
	writeDomoticaTunnelJSONBytes(w, http.StatusOK, body)
}

func handleDomoticaTunnelAck(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, device *dbpkg.EmpresaControlElectricoTunnelDevice, requestBytes int64) {
	var payload struct {
		CommandID string `json:"command_id"`
		OK        bool   `json:"ok"`
		Result    string `json:"result"`
		Error     string `json:"error"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || strings.TrimSpace(payload.CommandID) == "" {
		writeDomoticaTunnelJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "command_id requerido"})
		return
	}
	command, err := dbpkg.CompleteEmpresaControlElectricoTunnelCommand(dbEmp, device.EmpresaID, device.RaspberryID, payload.CommandID, payload.OK, payload.Result, payload.Error, device.DeviceUID)
	if err != nil {
		writeDomoticaTunnelJSON(w, http.StatusNotFound, map[string]interface{}{"ok": false, "error": "comando no encontrado"})
		return
	}
	if command.AlreadyFinal {
		recordAndWriteDomoticaTunnel(w, r, dbEmp, device, requestBytes, map[string]interface{}{"ok": true, "duplicate": true}, "")
		return
	}
	response := map[string]interface{}{"ok": true}
	recordAndWriteDomoticaTunnel(w, r, dbEmp, device, requestBytes, response, "")
}

func handleDomoticaTunnelInput(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, device *dbpkg.EmpresaControlElectricoTunnelDevice, requestBytes int64) {
	var payload struct {
		GPIOPin    int    `json:"gpio_pin"`
		Value      string `json:"value"`
		SensorCode string `json:"sensor_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.GPIOPin < 0 || payload.GPIOPin > 27 {
		writeDomoticaTunnelJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "entrada GPIO invalida"})
		return
	}
	if strings.TrimSpace(payload.SensorCode) == "" {
		payload.SensorCode = "gpio:" + strconv.Itoa(payload.GPIOPin)
	}
	metadata, _ := json.Marshal(map[string]interface{}{"raspberry_id": device.RaspberryID, "device_uid": device.DeviceUID, "gpio_pin": payload.GPIOPin})
	results, err := evaluarControlElectricoReglasOrigen(dbEmp, device.EmpresaID, payload.SensorCode, payload.Value, device.DeviceUID, string(metadata), device.RaspberryID, payload.GPIOPin)
	if err != nil {
		recordAndWriteDomoticaTunnel(w, r, dbEmp, device, requestBytes, map[string]interface{}{"ok": false, "error": "no se pudo procesar la entrada"}, err.Error())
		return
	}
	response := map[string]interface{}{"ok": true, "results": results}
	recordAndWriteDomoticaTunnel(w, r, dbEmp, device, requestBytes, response, "")
}

func handleDomoticaTunnelTelemetry(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, device *dbpkg.EmpresaControlElectricoTunnelDevice, requestBytes int64) {
	var payload dbpkg.EmpresaControlElectricoLectura
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.ReleID <= 0 {
		writeDomoticaTunnelJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "telemetria invalida"})
		return
	}
	rele, err := dbpkg.GetEmpresaControlElectricoReleByID(dbEmp, device.EmpresaID, payload.ReleID)
	if err != nil || (rele.RaspberryID > 0 && rele.RaspberryID != device.RaspberryID) {
		writeDomoticaTunnelJSON(w, http.StatusForbidden, map[string]interface{}{"ok": false, "error": "aparato no pertenece al dispositivo"})
		return
	}
	payload.EmpresaID = device.EmpresaID
	payload.EstacionID = rele.EstacionID
	payload.Origen = "tunel_raspberry"
	id, err := dbpkg.InsertEmpresaControlElectricoLectura(dbEmp, payload)
	if err != nil {
		recordAndWriteDomoticaTunnel(w, r, dbEmp, device, requestBytes, map[string]interface{}{"ok": false, "error": "no se pudo guardar telemetria"}, err.Error())
		return
	}
	recordAndWriteDomoticaTunnel(w, r, dbEmp, device, requestBytes, map[string]interface{}{"ok": true, "id": id}, "")
}

func recordAndWriteDomoticaTunnel(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, device *dbpkg.EmpresaControlElectricoTunnelDevice, requestBytes int64, response interface{}, tunnelError string) {
	body := marshalDomoticaTunnelJSON(response)
	_ = dbpkg.RecordEmpresaControlElectricoTunnelTraffic(dbEmp, device, requestBytes, int64(len(body)), domoticaTunnelRemoteIP(r), "", tunnelError)
	status := http.StatusOK
	if tunnelError != "" {
		status = http.StatusInternalServerError
	}
	writeDomoticaTunnelJSONBytes(w, status, body)
}

func domoticaTunnelRemoteIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	if forwarded := firstForwardedValue(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		return truncateHTTPText(forwarded, 120)
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return host
	}
	return truncateHTTPText(strings.TrimSpace(r.RemoteAddr), 120)
}

func marshalDomoticaTunnelJSON(payload interface{}) []byte {
	body, err := json.Marshal(payload)
	if err != nil {
		return []byte(`{"ok":false,"error":"respuesta invalida"}`)
	}
	return append(body, '\n')
}

func writeDomoticaTunnelJSON(w http.ResponseWriter, status int, payload interface{}) {
	writeDomoticaTunnelJSONBytes(w, status, marshalDomoticaTunnelJSON(payload))
}

func writeDomoticaTunnelJSONBytes(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func truncateHTTPText(raw string, limit int) string {
	raw = strings.TrimSpace(raw)
	if limit > 0 && len(raw) > limit {
		return raw[:limit]
	}
	return raw
}

func domoticaTunnelTrafficHuman(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(bytes)/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB", float64(bytes)/(1024*1024*1024))
}
