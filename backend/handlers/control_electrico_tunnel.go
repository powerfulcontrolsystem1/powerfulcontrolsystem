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
	auditMetadata, _ := json.Marshal(map[string]interface{}{
		"uso_tipo": pi.UsoTipo, "puerta_reles_salida": pi.PuertaRelesSalida, "puerta_delay_ms": pi.PuertaDelayMS,
	})
	_, _ = dbpkg.InsertEmpresaControlElectricoEvento(dbEmp, dbpkg.EmpresaControlElectricoEvento{
		EmpresaID:    empresaID,
		RaspberryID:  pi.ID,
		Comando:      "provisionar_tunel",
		Resultado:    "instalador_generado",
		Actor:        actor,
		Origen:       "panel_domotica",
		MetadataJSON: string(auditMetadata),
	})
	filenameCode := sanitizeDomoticaASCII(firstNonEmpty(pi.Codigo, pi.Nombre, "raspberry"))
	if filenameCode == "" {
		filenameCode = "raspberry"
	}
	w.Header().Set("Content-Type", "text/x-shellscript; charset=utf-8")
	prefix := "instalar-pcs-domotica-"
	if pi.UsoTipo == dbpkg.ControlElectricoUsoSensorPuertas {
		prefix = "instalar-pcs-sensores-puertas-"
	}
	w.Header().Set("Content-Disposition", `attachment; filename="`+prefix+filenameCode+`.sh"`)
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
		if status, limitErr := dbpkg.CheckEmpresaControlElectricoTunnelBandwidth(dbEmp, device.EmpresaID, time.Now()); limitErr != nil {
			if errors.Is(limitErr, dbpkg.ErrControlElectricoTunnelBandwidthExceeded) {
				w.Header().Set("Retry-After", "3600")
				writeDomoticaTunnelJSON(w, http.StatusTooManyRequests, map[string]interface{}{"ok": false, "error": "limite mensual de transferencia alcanzado", "month": status.Mes})
				return
			}
			log.Printf("[domotica_tunnel] bandwidth check empresa_id=%d raspberry_id=%d error: %v", device.EmpresaID, device.RaspberryID, limitErr)
			writeDomoticaTunnelJSON(w, http.StatusServiceUnavailable, map[string]interface{}{"ok": false, "error": "control de transferencia no disponible"})
			return
		}
		switch action {
		case "poll":
			handleDomoticaTunnelPoll(w, r, dbEmp, device, requestBytes)
		case "ack":
			handleDomoticaTunnelAck(w, r, dbEmp, device, requestBytes)
		case "input":
			handleDomoticaTunnelInput(w, r, dbEmp, device, requestBytes)
		case "door_scan":
			handleDomoticaTunnelDoorScan(w, r, dbEmp, device, requestBytes)
		case "telemetry":
			handleDomoticaTunnelTelemetry(w, r, dbEmp, device, requestBytes)
		case "solar_telemetry":
			handleDomoticaTunnelSolarTelemetry(w, r, dbEmp, device, requestBytes)
		case "relay_topology":
			handleDomoticaTunnelRelayTopology(w, r, dbEmp, device, requestBytes)
		default:
			writeDomoticaTunnelJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "action no soportada"})
		}
	}
}

// handleDomoticaTunnelRelayTopology lets an enrolled controller report the
// electrical polarity of its own configured outputs. It cannot select another
// company, Raspberry, relay, or station: all of those are resolved server-side
// from the authenticated tunnel device.
func handleDomoticaTunnelRelayTopology(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, device *dbpkg.EmpresaControlElectricoTunnelDevice, requestBytes int64) {
	var payload struct {
		Relays []struct {
			GPIOPin    int  `json:"gpio_pin"`
			ActiveHigh bool `json:"active_high"`
		} `json:"relays"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || len(payload.Relays) == 0 || len(payload.Relays) > 28 {
		writeDomoticaTunnelJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "topologia de relays invalida"})
		return
	}
	wanted := make(map[int]bool, len(payload.Relays))
	for _, relay := range payload.Relays {
		if relay.GPIOPin < 0 || relay.GPIOPin > 27 {
			writeDomoticaTunnelJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "gpio_pin invalido"})
			return
		}
		if _, duplicate := wanted[relay.GPIOPin]; duplicate {
			writeDomoticaTunnelJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "gpio_pin repetido"})
			return
		}
		wanted[relay.GPIOPin] = relay.ActiveHigh
	}
	relays, err := dbpkg.ListEmpresaControlElectricoReles(dbEmp, device.EmpresaID, false)
	if err != nil {
		recordAndWriteDomoticaTunnel(w, r, dbEmp, device, requestBytes, map[string]interface{}{"ok": false, "error": "no se pudieron consultar relays"}, err.Error())
		return
	}
	updated := 0
	for i := range relays {
		relay := &relays[i]
		activeHigh, found := wanted[relay.GPIOPin]
		if !found || relay.RaspberryID != device.RaspberryID || relay.ActiveHigh == activeHigh {
			continue
		}
		relay.ActiveHigh = activeHigh
		relay.UsuarioCreador = device.DeviceUID
		if _, err := dbpkg.UpsertEmpresaControlElectricoRele(dbEmp, relay); err != nil {
			recordAndWriteDomoticaTunnel(w, r, dbEmp, device, requestBytes, map[string]interface{}{"ok": false, "error": "no se pudo actualizar polaridad"}, err.Error())
			return
		}
		updated++
	}
	recordAndWriteDomoticaTunnel(w, r, dbEmp, device, requestBytes, map[string]interface{}{"ok": true, "updated": updated}, "")
}

// handleDomoticaTunnelSolarTelemetry accepts VE.Direct metrics only from an
// enrolled Raspberry. The tenant and controller are resolved from its tunnel
// token, never from fields sent by the device.
func handleDomoticaTunnelSolarTelemetry(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, device *dbpkg.EmpresaControlElectricoTunnelDevice, requestBytes int64) {
	var payload struct {
		Modelo  string                           `json:"modelo"`
		Lectura dbpkg.EmpresaEnergiaSolarLectura `json:"lectura"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeDomoticaTunnelJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "telemetria solar invalida"})
		return
	}
	if err := dbpkg.EmpresaEnergiaSolarSchemaReady(dbEmp); err != nil {
		recordAndWriteDomoticaTunnel(w, r, dbEmp, device, requestBytes, map[string]interface{}{"ok": false, "error": "modulo solar no disponible"}, err.Error())
		return
	}
	ref := "raspberry:" + strconv.FormatInt(device.RaspberryID, 10) + ":vedirect"
	sistemas, err := dbpkg.ListEmpresaEnergiaSolarSistemas(dbEmp, device.EmpresaID, true)
	if err != nil {
		recordAndWriteDomoticaTunnel(w, r, dbEmp, device, requestBytes, map[string]interface{}{"ok": false, "error": "no se pudo consultar sistema solar"}, err.Error())
		return
	}
	var sistemaID int64
	for _, sistema := range sistemas {
		if strings.EqualFold(strings.TrimSpace(sistema.InstalacionRef), ref) {
			sistemaID = sistema.ID
			break
		}
	}
	if sistemaID == 0 {
		sistemaID, err = dbpkg.UpsertEmpresaEnergiaSolarSistema(dbEmp, dbpkg.EmpresaEnergiaSolarSistema{
			EmpresaID: device.EmpresaID, Proveedor: dbpkg.EnergiaSolarProviderVictron,
			Modelo: firstNonEmpty(strings.TrimSpace(payload.Modelo), "BlueSolar MPPT VE.Direct"),
			Nombre: "Victron solar - Raspberry Pi", Ubicacion: "Gateway VE.Direct por tunel PCS",
			InstalacionRef: ref, LocalGatewayURL: "tunnel://raspberry/" + strconv.FormatInt(device.RaspberryID, 10) + "/vedirect",
			IntervaloSegundos: 60, Activo: true, Estado: "activo", UsuarioCreador: device.DeviceUID,
			Observaciones: "Telemetria VE.Direct autenticada desde Raspberry Pi; sin acceso entrante al VPS.",
		})
		if err != nil {
			recordAndWriteDomoticaTunnel(w, r, dbEmp, device, requestBytes, map[string]interface{}{"ok": false, "error": "no se pudo crear sistema solar"}, err.Error())
			return
		}
	}
	payload.Lectura.EmpresaID = device.EmpresaID
	payload.Lectura.SistemaID = sistemaID
	payload.Lectura.UsuarioCreador = device.DeviceUID
	if payload.Lectura.Raw == nil {
		payload.Lectura.Raw = map[string]interface{}{}
	}
	payload.Lectura.Raw["origen"] = "vedirect_tunel_raspberry"
	payload.Lectura.Raw["raspberry_id"] = device.RaspberryID
	lecturaID, err := dbpkg.InsertEmpresaEnergiaSolarLectura(dbEmp, payload.Lectura)
	if err != nil {
		recordAndWriteDomoticaTunnel(w, r, dbEmp, device, requestBytes, map[string]interface{}{"ok": false, "error": "no se pudo guardar telemetria solar"}, err.Error())
		return
	}
	recordAndWriteDomoticaTunnel(w, r, dbEmp, device, requestBytes, map[string]interface{}{"ok": true, "sistema_id": sistemaID, "lectura_id": lecturaID}, "")
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
	metadata, _ := json.Marshal(map[string]interface{}{"agent_version": truncateHTTPText(payload.AgentVersion, 80)})
	_, _ = dbpkg.InsertEmpresaControlElectricoEvento(dbEmp, dbpkg.EmpresaControlElectricoEvento{
		EmpresaID: device.EmpresaID, RaspberryID: device.RaspberryID, Comando: "raspberry_enrolada",
		Resultado: "conectado", Actor: device.DeviceUID, Origen: "tunel_raspberry", MetadataJSON: string(metadata),
	})
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
		BootID       string `json:"boot_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeDomoticaTunnelJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "json invalido"})
		return
	}
	isDoorSensor := device.UsoTipo == dbpkg.ControlElectricoUsoSensorPuertas
	recoveryQueued := 0
	if !isDoorSensor && strings.TrimSpace(payload.BootID) != "" {
		var recoveryErr error
		recoveryQueued, recoveryErr = dbpkg.QueueEmpresaControlElectricoTunnelRestoreOnBoot(dbEmp, device, payload.BootID)
		if recoveryErr != nil {
			log.Printf("[domotica_tunnel] restore queue empresa_id=%d raspberry_id=%d error: %v", device.EmpresaID, device.RaspberryID, recoveryErr)
		}
	}
	pollWait := 20 * time.Second
	if isDoorSensor {
		pollWait = 5 * time.Second
	}
	deadline := time.Now().Add(pollWait)
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
	inputs := []dbpkg.EmpresaControlElectricoInputConfig{}
	if !isDoorSensor {
		var err error
		inputs, err = dbpkg.ListEmpresaControlElectricoInputConfigs(dbEmp, device.EmpresaID, device.RaspberryID)
		if err != nil {
			inputs = []dbpkg.EmpresaControlElectricoInputConfig{}
		}
	}
	response := map[string]interface{}{"ok": true, "usage_type": device.UsoTipo, "inputs": inputs, "server_time": time.Now().UTC().Format(time.RFC3339)}
	if isDoorSensor {
		response["door_scan"] = dbpkg.BuildEmpresaControlElectricoDoorScanConfig(device.PuertaRelesSalida, device.PuertaDelayMS)
	}
	if recoveryQueued > 0 {
		response["recovery_queued"] = recoveryQueued
	}
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

func handleDomoticaTunnelDoorScan(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, device *dbpkg.EmpresaControlElectricoTunnelDevice, requestBytes int64) {
	if device.UsoTipo != dbpkg.ControlElectricoUsoSensorPuertas {
		writeDomoticaTunnelJSON(w, http.StatusForbidden, map[string]interface{}{"ok": false, "error": "Raspberry no configurada para sensores de puertas"})
		return
	}
	var payload struct {
		Readings []dbpkg.EmpresaControlElectricoDoorReading `json:"readings"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || len(payload.Readings) == 0 {
		writeDomoticaTunnelJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "lecturas de puertas invalidas"})
		return
	}
	for _, reading := range payload.Readings {
		if reading.OutputIndex > device.PuertaRelesSalida {
			writeDomoticaTunnelJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "salida fuera de la configuracion empresarial"})
			return
		}
	}
	transitions, err := dbpkg.ApplyEmpresaControlElectricoDoorScan(dbEmp, device.EmpresaID, device.RaspberryID, payload.Readings)
	if err != nil {
		recordAndWriteDomoticaTunnel(w, r, dbEmp, device, requestBytes, map[string]interface{}{"ok": false, "error": "no se pudieron guardar las lecturas de puertas"}, err.Error())
		return
	}
	changed := 0
	autoActivated := 0
	warnings := make([]string, 0)
	for _, transition := range transitions {
		if !transition.Changed {
			continue
		}
		changed++
		activated, warning := maybeAutoActivateStationFromSensor(dbEmp, device.EmpresaID, transition.EstacionID, transition.State)
		if activated {
			autoActivated++
		}
		if warning != "" {
			warnings = append(warnings, warning)
		}
	}
	response := map[string]interface{}{"ok": true, "readings": len(transitions), "changed": changed, "auto_activated": autoActivated}
	if len(warnings) > 0 {
		response["warnings"] = warnings
	}
	recordAndWriteDomoticaTunnel(w, r, dbEmp, device, requestBytes, response, "")
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
