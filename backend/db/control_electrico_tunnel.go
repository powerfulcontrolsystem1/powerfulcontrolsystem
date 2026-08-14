package db

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	controlElectricoTunnelEnrollmentTTL = 24 * time.Hour
	controlElectricoTunnelCommandTTL    = 2 * time.Minute
	controlElectricoRestoreCommandTTL   = 10 * time.Minute
)

type EmpresaControlElectricoTunnelDevice struct {
	EmpresaID   int64
	RaspberryID int64
	DeviceUID   string
	Nombre      string
	Codigo      string
}

type EmpresaControlElectricoTunnelCommand struct {
	ID             int64  `json:"-"`
	EmpresaID      int64  `json:"-"`
	RaspberryID    int64  `json:"-"`
	CommandUID     string `json:"command_id"`
	ReleID         int64  `json:"rele_id,omitempty"`
	EstacionID     int64  `json:"estacion_id,omitempty"`
	GPIOPin        int    `json:"gpio_pin"`
	EstadoObjetivo string `json:"estado"`
	PayloadJSON    string `json:"-"`
	Estado         string `json:"-"`
	Intentos       int    `json:"-"`
	Resultado      string `json:"-"`
	Error          string `json:"-"`
	AlreadyFinal   bool   `json:"-"`
}

type EmpresaControlElectricoInputConfig struct {
	ReglaID              int64  `json:"rule_id"`
	GPIOPin              int    `json:"gpio_pin"`
	Pull                 string `json:"pull"`
	DebounceMS           int    `json:"debounce_ms"`
	ValorActivo          string `json:"active_value"`
	SensorCodigo         string `json:"sensor_code"`
	Accion               string `json:"action"`
	ReleID               int64  `json:"target_relay_id,omitempty"`
	EstacionID           int64  `json:"station_id,omitempty"`
	Nombre               string `json:"name,omitempty"`
	AlarmaHabilitada     bool   `json:"alarm_enabled"`
	TemporizadorSegundos int    `json:"timer_seconds,omitempty"`
}

type EmpresaControlElectricoTraficoRaspberry struct {
	EmpresaID       int64  `json:"empresa_id"`
	RaspberryID     int64  `json:"raspberry_id"`
	Codigo          string `json:"codigo"`
	Nombre          string `json:"nombre"`
	DeviceUID       string `json:"device_uid"`
	TunnelEnabled   bool   `json:"tunnel_enabled"`
	TunnelStatus    string `json:"tunnel_status"`
	LastSeen        string `json:"last_seen"`
	LastIP          string `json:"last_ip"`
	AgentVersion    string `json:"agent_version"`
	BytesRx         int64  `json:"bytes_rx"`
	BytesTx         int64  `json:"bytes_tx"`
	TodayBytesRx    int64  `json:"today_bytes_rx"`
	TodayBytesTx    int64  `json:"today_bytes_tx"`
	TodayRequests   int64  `json:"today_requests"`
	MonthBytesRx    int64  `json:"month_bytes_rx"`
	MonthBytesTx    int64  `json:"month_bytes_tx"`
	LastTunnelError string `json:"last_tunnel_error,omitempty"`
}

func EmpresaControlElectricoTunnelSchemaReady(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("conexion de base de datos no disponible")
	}
	checks := []string{
		`SELECT device_uid FROM empresa_control_electrico_raspberry_pis WHERE 1=0`,
		`SELECT raspberry_id FROM empresa_control_electrico_reglas WHERE 1=0`,
		`SELECT id FROM empresa_control_electrico_comandos WHERE 1=0`,
		`SELECT id FROM empresa_control_electrico_trafico_diario WHERE 1=0`,
		`SELECT empresa_id FROM empresa_control_electrico_limites_tunel WHERE 1=0`,
	}
	for _, query := range checks {
		var marker interface{}
		err := queryRowSQLCompat(dbConn, query).Scan(&marker)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return fmt.Errorf("esquema de tunel domotica no disponible: %w", err)
		}
	}
	return nil
}

func generateControlElectricoTunnelSecret(bytes int) (string, error) {
	if bytes < 16 {
		bytes = 16
	}
	raw := make([]byte, bytes)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func controlElectricoTunnelTokenHash(raw string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(raw)))
	return hex.EncodeToString(sum[:])
}

func generateControlElectricoTunnelDeviceUID() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return "RPI-" + strings.ToUpper(hex.EncodeToString(raw)), nil
}

// ProvisionEmpresaControlElectricoRaspberryTunnel crea o rota el secreto de
// enrolamiento. El secreto plano se devuelve una sola vez para incorporarlo al
// instalador descargable; la base conserva solamente su huella.
func ProvisionEmpresaControlElectricoRaspberryTunnel(dbConn *sql.DB, empresaID, raspberryID int64, actor string) (*EmpresaControlElectricoRaspberry, string, error) {
	if dbConn == nil || empresaID <= 0 || raspberryID <= 0 {
		return nil, "", errors.New("empresa_id y raspberry_id son obligatorios")
	}
	var deviceUID string
	err := queryRowSQLCompat(dbConn, `SELECT COALESCE(device_uid,'') FROM empresa_control_electrico_raspberry_pis WHERE empresa_id=? AND id=? AND LOWER(COALESCE(estado,'activo'))='activo'`, empresaID, raspberryID).Scan(&deviceUID)
	if err != nil {
		return nil, "", err
	}
	if strings.TrimSpace(deviceUID) == "" {
		deviceUID, err = generateControlElectricoTunnelDeviceUID()
		if err != nil {
			return nil, "", err
		}
	}
	enrollmentToken, err := generateControlElectricoTunnelSecret(32)
	if err != nil {
		return nil, "", err
	}
	expires := time.Now().UTC().Add(controlElectricoTunnelEnrollmentTTL).Format(time.RFC3339)
	// Un instalador nuevo no invalida el agente activo antes de tiempo. El token
	// operativo se reemplaza atomica y definitivamente cuando el nuevo agente
	// completa el enrolamiento; así una descarga o un SSH fallido no deja la
	// Raspberry desconectada.
	result, err := execSQLCompat(dbConn, `UPDATE empresa_control_electrico_raspberry_pis SET device_uid=?, enrollment_token_hash=?, enrollment_expires_at=?, tunnel_enabled=1, tunnel_status=CASE WHEN device_token_hash IS NULL THEN 'pendiente_instalacion' ELSE tunnel_status END, last_tunnel_error='', fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=? WHERE empresa_id=? AND id=?`,
		deviceUID, controlElectricoTunnelTokenHash(enrollmentToken), expires, truncateControlElectricoText(actor, 180), empresaID, raspberryID)
	if err != nil {
		return nil, "", err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return nil, "", sql.ErrNoRows
	}
	item, err := GetEmpresaControlElectricoRaspberryByID(dbConn, empresaID, raspberryID, false)
	return item, enrollmentToken, err
}

func EnrollEmpresaControlElectricoRaspberryTunnel(dbConn *sql.DB, deviceUID, enrollmentToken string) (*EmpresaControlElectricoTunnelDevice, string, error) {
	deviceUID = strings.TrimSpace(deviceUID)
	enrollmentToken = strings.TrimSpace(enrollmentToken)
	if dbConn == nil || deviceUID == "" || enrollmentToken == "" {
		return nil, "", sql.ErrNoRows
	}
	var device EmpresaControlElectricoTunnelDevice
	var expiresRaw string
	err := queryRowSQLCompat(dbConn, `SELECT empresa_id, id, COALESCE(device_uid,''), COALESCE(nombre,''), COALESCE(codigo,''), COALESCE(enrollment_expires_at,'') FROM empresa_control_electrico_raspberry_pis WHERE device_uid=? AND enrollment_token_hash=? AND COALESCE(tunnel_enabled,0)=1 AND LOWER(COALESCE(estado,'activo'))='activo' LIMIT 1`,
		deviceUID, controlElectricoTunnelTokenHash(enrollmentToken)).Scan(&device.EmpresaID, &device.RaspberryID, &device.DeviceUID, &device.Nombre, &device.Codigo, &expiresRaw)
	if err != nil {
		return nil, "", err
	}
	expires, err := time.Parse(time.RFC3339, strings.TrimSpace(expiresRaw))
	if err != nil || time.Now().UTC().After(expires) {
		return nil, "", sql.ErrNoRows
	}
	deviceToken, err := generateControlElectricoTunnelSecret(48)
	if err != nil {
		return nil, "", err
	}
	result, err := execSQLCompat(dbConn, `UPDATE empresa_control_electrico_raspberry_pis SET device_token_hash=?, enrollment_token_hash=NULL, enrollment_expires_at=NULL, tunnel_status='conectado', last_seen=CURRENT_TIMESTAMP, last_tunnel_error='', fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=? AND enrollment_token_hash=?`,
		controlElectricoTunnelTokenHash(deviceToken), device.EmpresaID, device.RaspberryID, controlElectricoTunnelTokenHash(enrollmentToken))
	if err != nil {
		return nil, "", err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return nil, "", sql.ErrNoRows
	}
	return &device, deviceToken, nil
}

func AuthenticateEmpresaControlElectricoRaspberryTunnel(dbConn *sql.DB, deviceUID, deviceToken string) (*EmpresaControlElectricoTunnelDevice, error) {
	deviceUID = strings.TrimSpace(deviceUID)
	deviceToken = strings.TrimSpace(deviceToken)
	if dbConn == nil || deviceUID == "" || deviceToken == "" {
		return nil, sql.ErrNoRows
	}
	var device EmpresaControlElectricoTunnelDevice
	err := queryRowSQLCompat(dbConn, `SELECT empresa_id, id, COALESCE(device_uid,''), COALESCE(nombre,''), COALESCE(codigo,'') FROM empresa_control_electrico_raspberry_pis WHERE device_uid=? AND device_token_hash=? AND COALESCE(tunnel_enabled,0)=1 AND LOWER(COALESCE(estado,'activo'))='activo' LIMIT 1`,
		deviceUID, controlElectricoTunnelTokenHash(deviceToken)).Scan(&device.EmpresaID, &device.RaspberryID, &device.DeviceUID, &device.Nombre, &device.Codigo)
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func RecordEmpresaControlElectricoTunnelTraffic(dbConn *sql.DB, device *EmpresaControlElectricoTunnelDevice, bytesRX, bytesTX int64, remoteIP, agentVersion, tunnelError string) error {
	if dbConn == nil || device == nil || device.EmpresaID <= 0 || device.RaspberryID <= 0 {
		return errors.New("dispositivo de tunel invalido")
	}
	if bytesRX < 0 || bytesRX > 16*1024*1024 {
		bytesRX = 0
	}
	if bytesTX < 0 || bytesTX > 16*1024*1024 {
		bytesTX = 0
	}
	status := "conectado"
	if strings.TrimSpace(tunnelError) != "" {
		status = "error"
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := execTxSQLCompat(tx, `UPDATE empresa_control_electrico_raspberry_pis SET tunnel_status=?, last_seen=CURRENT_TIMESTAMP, last_ip=?, agent_version=COALESCE(NULLIF(?,''),agent_version), bytes_rx=COALESCE(bytes_rx,0)+?, bytes_tx=COALESCE(bytes_tx,0)+?, last_tunnel_error=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`,
		status, truncateControlElectricoText(remoteIP, 120), truncateControlElectricoText(agentVersion, 80), bytesRX, bytesTX, truncateControlElectricoText(tunnelError, 500), device.EmpresaID, device.RaspberryID); err != nil {
		return err
	}
	date := time.Now().UTC().Format("2006-01-02")
	if _, err := execTxSQLCompat(tx, `INSERT INTO empresa_control_electrico_trafico_diario (empresa_id, raspberry_id, fecha, bytes_rx, bytes_tx, solicitudes, fecha_actualizacion) VALUES (?, ?, ?, ?, ?, 1, CURRENT_TIMESTAMP) ON CONFLICT (empresa_id, raspberry_id, fecha) DO UPDATE SET bytes_rx=empresa_control_electrico_trafico_diario.bytes_rx+EXCLUDED.bytes_rx, bytes_tx=empresa_control_electrico_trafico_diario.bytes_tx+EXCLUDED.bytes_tx, solicitudes=empresa_control_electrico_trafico_diario.solicitudes+1, fecha_actualizacion=CURRENT_TIMESTAMP`,
		device.EmpresaID, device.RaspberryID, date, bytesRX, bytesTX); err != nil {
		return err
	}
	return tx.Commit()
}

func QueueEmpresaControlElectricoTunnelCommand(dbConn *sql.DB, empresaID, raspberryID, releID, estacionID int64, gpioPin int, targetState string, payload interface{}, actor, origen string) (*EmpresaControlElectricoTunnelCommand, error) {
	if dbConn == nil || empresaID <= 0 || raspberryID <= 0 || releID < 0 {
		return nil, errors.New("empresa y raspberry son obligatorias")
	}
	commandUID, err := generateControlElectricoTunnelSecret(24)
	if err != nil {
		return nil, err
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	if len(body) > 16*1024 {
		return nil, errors.New("comando de tunel demasiado grande")
	}
	targetState = strings.ToLower(strings.TrimSpace(targetState))
	if targetState != "on" && targetState != "off" && targetState != "operacion" {
		return nil, errors.New("estado objetivo invalido")
	}
	availableAt := time.Now().UTC()
	if targetState == "on" {
		availableAt, _, err = reserveEmpresaControlElectricoActivationSlot(dbConn, empresaID)
		if err != nil {
			return nil, err
		}
	}
	expires := availableAt.Add(controlElectricoTunnelCommandTTL).Format(time.RFC3339Nano)
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_control_electrico_comandos (empresa_id, raspberry_id, command_uid, rele_id, estacion_id, gpio_pin, estado_objetivo, payload_json, estado, intentos, solicitado_en, disponible_desde, expira_en, usuario_creador, origen) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'pendiente', 0, CURRENT_TIMESTAMP, ?, ?, ?, ?)`,
		empresaID, raspberryID, commandUID, releID, estacionID, gpioPin, targetState, string(body), availableAt.Format(time.RFC3339Nano), expires, truncateControlElectricoText(actor, 180), truncateControlElectricoText(origen, 100))
	if err != nil {
		return nil, err
	}
	return &EmpresaControlElectricoTunnelCommand{ID: id, EmpresaID: empresaID, RaspberryID: raspberryID, CommandUID: commandUID, ReleID: releID, EstacionID: estacionID, GPIOPin: gpioPin, EstadoObjetivo: targetState, PayloadJSON: string(body), Estado: "pendiente"}, nil
}

// reserveEmpresaControlElectricoActivationSlot reserves the next ON slot for
// one company. The configuration row is locked transactionally, so commands
// from different cashiers, stations or Raspberry Pis cannot energize loads at
// the same instant.
func reserveEmpresaControlElectricoActivationSlot(dbConn *sql.DB, empresaID int64) (time.Time, int, error) {
	if dbConn == nil || empresaID <= 0 {
		return time.Time{}, 0, errors.New("empresa invalida para cola de activacion")
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return time.Time{}, 0, err
	}
	defer tx.Rollback()
	slot, delay, err := reserveEmpresaControlElectricoActivationSlotTx(tx, empresaID)
	if err != nil {
		return time.Time{}, 0, err
	}
	if err := tx.Commit(); err != nil {
		return time.Time{}, 0, err
	}
	return slot, delay, nil
}

func reserveEmpresaControlElectricoActivationSlotTx(tx *sql.Tx, empresaID int64) (time.Time, int, error) {
	if tx == nil || empresaID <= 0 {
		return time.Time{}, 0, errors.New("empresa invalida para cola de activacion")
	}
	if _, err := execTxSQLCompat(tx, `INSERT INTO empresa_control_electrico_config (empresa_id, habilitado, raspberry_port, api_path, timeout_ms, auto_sync_estaciones, activation_delay_seconds, estado) VALUES (?, 1, 8081, '/api/gpio/relay', 2500, 1, 1, 'activo') ON CONFLICT (empresa_id) DO NOTHING`, empresaID); err != nil {
		return time.Time{}, 0, err
	}
	var delay int
	var nextRaw string
	if err := queryRowTxSQLCompat(tx, `SELECT COALESCE(activation_delay_seconds,1), COALESCE(next_activation_at,'') FROM empresa_control_electrico_config WHERE empresa_id=? FOR UPDATE`, empresaID).Scan(&delay, &nextRaw); err != nil {
		return time.Time{}, 0, err
	}
	if delay < 1 {
		delay = 1
	}
	if delay > 60 {
		delay = 60
	}
	now := time.Now().UTC()
	slot := now
	if next, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(nextRaw)); err == nil && next.After(slot) {
		slot = next
	}
	next := slot.Add(time.Duration(delay) * time.Second)
	if _, err := execTxSQLCompat(tx, `UPDATE empresa_control_electrico_config SET activation_delay_seconds=?, next_activation_at=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=?`, delay, next.Format(time.RFC3339Nano), empresaID); err != nil {
		return time.Time{}, 0, err
	}
	return slot, delay, nil
}

// QueueEmpresaControlElectricoTunnelRestoreOnBoot recrea, una sola vez por
// arranque, los comandos ON previamente confirmados del dispositivo. La
// transaccion evita duplicados en los reintentos de long polling y conserva el
// filtro estricto por empresa y Raspberry.
func QueueEmpresaControlElectricoTunnelRestoreOnBoot(dbConn *sql.DB, device *EmpresaControlElectricoTunnelDevice, bootID string) (int, error) {
	if dbConn == nil || device == nil || device.EmpresaID <= 0 || device.RaspberryID <= 0 {
		return 0, errors.New("dispositivo de tunel invalido")
	}
	bootID = truncateControlElectricoText(strings.TrimSpace(bootID), 96)
	if bootID == "" {
		return 0, nil
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	updated, err := execTxSQLCompat(tx, `UPDATE empresa_control_electrico_raspberry_pis SET last_boot_id=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=? AND COALESCE(last_boot_id,'')<>?`, bootID, device.EmpresaID, device.RaspberryID, bootID)
	if err != nil {
		return 0, err
	}
	changed, err := updated.RowsAffected()
	if err != nil {
		return 0, err
	}
	if changed == 0 {
		return 0, tx.Commit()
	}
	rows, err := queryTxSQLCompat(tx, `SELECT id, COALESCE(estacion_id,0), COALESCE(gpio_pin,0), COALESCE(active_high,1), COALESCE(pulso_ms,0), COALESCE(salida_codigo,''), COALESCE(relay_name,''), COALESCE(tipo_carga,'') FROM empresa_control_electrico_reles WHERE empresa_id=? AND raspberry_id=? AND LOWER(COALESCE(estado,'activo'))='activo' AND LOWER(COALESCE(ultimo_estado,''))='on' ORDER BY estacion_id, gpio_pin, id`, device.EmpresaID, device.RaspberryID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var releID, estacionID int64
		var gpioPin, activeHigh, pulsoMS int
		var salidaCodigo, relayName, tipoCarga string
		if err := rows.Scan(&releID, &estacionID, &gpioPin, &activeHigh, &pulsoMS, &salidaCodigo, &relayName, &tipoCarga); err != nil {
			return 0, err
		}
		commandUID, err := generateControlElectricoTunnelSecret(24)
		if err != nil {
			return 0, err
		}
		availableAt, delay, err := reserveEmpresaControlElectricoActivationSlotTx(tx, device.EmpresaID)
		if err != nil {
			return 0, err
		}
		payload, err := json.Marshal(map[string]interface{}{
			"relay_id": releID, "station_id": estacionID, "gpio_pin": gpioPin, "estado": "on",
			"active_high": activeHigh == 1, "pulso_ms": pulsoMS, "salida_codigo": salidaCodigo,
			"relay_name": relayName, "tipo_carga": tipoCarga, "origen": "raspberry_recovery",
			"restore_delay_ms": delay * 1000,
		})
		if err != nil {
			return 0, err
		}
		expires := availableAt.Add(controlElectricoRestoreCommandTTL).Format(time.RFC3339Nano)
		if _, err := execTxSQLCompat(tx, `INSERT INTO empresa_control_electrico_comandos (empresa_id, raspberry_id, command_uid, rele_id, estacion_id, gpio_pin, estado_objetivo, payload_json, estado, intentos, solicitado_en, disponible_desde, expira_en, usuario_creador, origen) VALUES (?, ?, ?, ?, ?, ?, 'on', ?, 'pendiente', 0, CURRENT_TIMESTAMP, ?, ?, ?, 'raspberry_recovery')`, device.EmpresaID, device.RaspberryID, commandUID, releID, estacionID, gpioPin, string(payload), availableAt.Format(time.RFC3339Nano), expires, device.DeviceUID); err != nil {
			return 0, err
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return count, nil
}

func ClaimEmpresaControlElectricoTunnelCommand(dbConn *sql.DB, empresaID, raspberryID int64) (*EmpresaControlElectricoTunnelCommand, error) {
	if dbConn == nil || empresaID <= 0 || raspberryID <= 0 {
		return nil, sql.ErrNoRows
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	// Las fechas de la cola se conservan como texto RFC3339 por compatibilidad
	// histórica. Deben compararse como TIMESTAMPTZ: un CAST a TIMESTAMP elimina
	// la Z/offset y retrasa los comandos UTC varias horas cuando PostgreSQL usa
	// America/Bogota u otra zona distinta de UTC.
	_, _ = execTxSQLCompat(tx, `UPDATE empresa_control_electrico_comandos SET estado='pendiente', entregado_en=NULL WHERE empresa_id=? AND raspberry_id=? AND estado='entregado' AND intentos<3 AND CAST(NULLIF(entregado_en,'') AS TIMESTAMPTZ)<CURRENT_TIMESTAMP-INTERVAL '30 seconds'`, empresaID, raspberryID)
	_, _ = execTxSQLCompat(tx, `UPDATE empresa_control_electrico_comandos SET estado='expirado', completado_en=CURRENT_TIMESTAMP WHERE empresa_id=? AND raspberry_id=? AND estado IN ('pendiente','entregado') AND CAST(NULLIF(expira_en,'') AS TIMESTAMPTZ)<CURRENT_TIMESTAMP`, empresaID, raspberryID)
	var command EmpresaControlElectricoTunnelCommand
	err = queryRowTxSQLCompat(tx, `SELECT id, empresa_id, raspberry_id, command_uid, COALESCE(rele_id,0), COALESCE(estacion_id,0), COALESCE(gpio_pin,0), COALESCE(estado_objetivo,''), COALESCE(payload_json,''), COALESCE(estado,''), COALESCE(intentos,0), COALESCE(resultado,''), COALESCE(error,'') FROM empresa_control_electrico_comandos WHERE empresa_id=? AND raspberry_id=? AND estado='pendiente' AND (COALESCE(disponible_desde,'')='' OR CAST(NULLIF(disponible_desde,'') AS TIMESTAMPTZ)<=CURRENT_TIMESTAMP) ORDER BY id FOR UPDATE SKIP LOCKED LIMIT 1`, empresaID, raspberryID).
		Scan(&command.ID, &command.EmpresaID, &command.RaspberryID, &command.CommandUID, &command.ReleID, &command.EstacionID, &command.GPIOPin, &command.EstadoObjetivo, &command.PayloadJSON, &command.Estado, &command.Intentos, &command.Resultado, &command.Error)
	if err != nil {
		return nil, err
	}
	if _, err := execTxSQLCompat(tx, `UPDATE empresa_control_electrico_comandos SET estado='entregado', entregado_en=CURRENT_TIMESTAMP, intentos=intentos+1 WHERE id=? AND empresa_id=? AND raspberry_id=?`, command.ID, empresaID, raspberryID); err != nil {
		return nil, err
	}
	command.Estado = "entregado"
	command.Intentos++
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &command, nil
}

func GetEmpresaControlElectricoTunnelCommand(dbConn *sql.DB, empresaID, raspberryID int64, commandUID string) (*EmpresaControlElectricoTunnelCommand, error) {
	var command EmpresaControlElectricoTunnelCommand
	err := queryRowSQLCompat(dbConn, `SELECT id, empresa_id, raspberry_id, command_uid, COALESCE(rele_id,0), COALESCE(estacion_id,0), COALESCE(gpio_pin,0), COALESCE(estado_objetivo,''), COALESCE(payload_json,''), COALESCE(estado,''), COALESCE(intentos,0), COALESCE(resultado,''), COALESCE(error,'') FROM empresa_control_electrico_comandos WHERE empresa_id=? AND raspberry_id=? AND command_uid=? LIMIT 1`, empresaID, raspberryID, strings.TrimSpace(commandUID)).
		Scan(&command.ID, &command.EmpresaID, &command.RaspberryID, &command.CommandUID, &command.ReleID, &command.EstacionID, &command.GPIOPin, &command.EstadoObjetivo, &command.PayloadJSON, &command.Estado, &command.Intentos, &command.Resultado, &command.Error)
	if err != nil {
		return nil, err
	}
	return &command, nil
}

// CompleteEmpresaControlElectricoTunnelCommand aplica el ACK, el estado visible,
// la lectura y la bitacora en una sola transaccion. El lock por comando hace
// idempotentes los reintentos y los ACK concurrentes del agente.
func CompleteEmpresaControlElectricoTunnelCommand(dbConn *sql.DB, empresaID, raspberryID int64, commandUID string, ok bool, resultText, errorText, actor string) (*EmpresaControlElectricoTunnelCommand, error) {
	if dbConn == nil || empresaID <= 0 || raspberryID <= 0 || strings.TrimSpace(commandUID) == "" {
		return nil, sql.ErrNoRows
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var command EmpresaControlElectricoTunnelCommand
	err = queryRowTxSQLCompat(tx, `SELECT id, empresa_id, raspberry_id, command_uid, COALESCE(rele_id,0), COALESCE(estacion_id,0), COALESCE(gpio_pin,0), COALESCE(estado_objetivo,''), COALESCE(payload_json,''), COALESCE(estado,''), COALESCE(intentos,0), COALESCE(resultado,''), COALESCE(error,'') FROM empresa_control_electrico_comandos WHERE empresa_id=? AND raspberry_id=? AND command_uid=? FOR UPDATE`, empresaID, raspberryID, strings.TrimSpace(commandUID)).
		Scan(&command.ID, &command.EmpresaID, &command.RaspberryID, &command.CommandUID, &command.ReleID, &command.EstacionID, &command.GPIOPin, &command.EstadoObjetivo, &command.PayloadJSON, &command.Estado, &command.Intentos, &command.Resultado, &command.Error)
	if err != nil {
		return nil, err
	}
	if command.Estado == "completado" || command.Estado == "error" || command.Estado == "expirado" {
		command.AlreadyFinal = true
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return &command, nil
	}
	if command.Estado != "pendiente" && command.Estado != "entregado" {
		return nil, sql.ErrNoRows
	}

	resultText = truncateControlElectricoText(resultText, 1200)
	errorText = truncateControlElectricoText(errorText, 800)
	status := "error"
	eventResult := "error"
	if ok {
		status = "completado"
		eventResult = "ok"
	}
	var releStationID int64
	var potenciaW float64
	var ultimoEstado string
	releErr := queryRowTxSQLCompat(tx, `SELECT COALESCE(estacion_id,0), COALESCE(potencia_w,0), COALESCE(ultimo_estado,'desconocido') FROM empresa_control_electrico_reles WHERE empresa_id=? AND id=? LIMIT 1`, empresaID, command.ReleID).Scan(&releStationID, &potenciaW, &ultimoEstado)
	if releErr != nil && releErr != sql.ErrNoRows {
		return nil, releErr
	}
	if releErr == nil {
		if ok {
			if _, err := execTxSQLCompat(tx, `UPDATE empresa_control_electrico_reles SET ultimo_estado=?, ultimo_comando='tunnel_ack', ultimo_error='', ultima_sincronizacion=CURRENT_TIMESTAMP, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, command.EstadoObjetivo, empresaID, command.ReleID); err != nil {
				return nil, err
			}
			consumoW := 0.0
			if command.EstadoObjetivo == "on" {
				consumoW = potenciaW
			}
			if _, err := execTxSQLCompat(tx, `INSERT INTO empresa_control_electrico_lecturas (empresa_id, estacion_id, rele_id, origen, estado, consumo_w, consumo_kwh, voltaje_v, corriente_a, fecha_lectura, metadata_json) VALUES (?, NULLIF(?,0), NULLIF(?,0), 'tunel_raspberry', ?, ?, 0, 0, 0, CURRENT_TIMESTAMP, '')`, empresaID, releStationID, command.ReleID, command.EstadoObjetivo, consumoW); err != nil {
				return nil, err
			}
		} else {
			if _, err := execTxSQLCompat(tx, `UPDATE empresa_control_electrico_reles SET ultimo_estado=?, ultimo_comando='tunnel_ack', ultimo_error=?, ultima_sincronizacion=CURRENT_TIMESTAMP, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, ultimoEstado, errorText, empresaID, command.ReleID); err != nil {
				return nil, err
			}
		}
	}
	if _, err := execTxSQLCompat(tx, `INSERT INTO empresa_control_electrico_eventos (empresa_id, estacion_id, rele_id, raspberry_id, gpio_pin, comando, estado_objetivo, resultado, http_status, raspberry_ip, response_body, error, fecha_evento, actor, origen, metadata_json) VALUES (?, NULLIF(?,0), NULLIF(?,0), ?, ?, 'tunnel_ack', ?, ?, 0, '', ?, ?, CURRENT_TIMESTAMP, ?, 'tunel_raspberry', '')`, empresaID, command.EstacionID, command.ReleID, raspberryID, command.GPIOPin, command.EstadoObjetivo, eventResult, resultText, errorText, truncateControlElectricoText(actor, 180)); err != nil {
		return nil, err
	}
	if _, err := execTxSQLCompat(tx, `UPDATE empresa_control_electrico_comandos SET estado=?, resultado=?, error=?, completado_en=CURRENT_TIMESTAMP WHERE id=? AND empresa_id=? AND raspberry_id=?`, status, resultText, errorText, command.ID, empresaID, raspberryID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	command.Estado = status
	command.Resultado = resultText
	command.Error = errorText
	return &command, nil
}

func WaitEmpresaControlElectricoTunnelCommand(dbConn *sql.DB, empresaID, raspberryID int64, commandUID string, timeout time.Duration) (*EmpresaControlElectricoTunnelCommand, error) {
	if timeout <= 0 {
		timeout = 2500 * time.Millisecond
	}
	deadline := time.Now().Add(timeout)
	for {
		command, err := GetEmpresaControlElectricoTunnelCommand(dbConn, empresaID, raspberryID, commandUID)
		if err != nil {
			return nil, err
		}
		switch command.Estado {
		case "completado", "error", "expirado":
			return command, nil
		}
		if time.Now().After(deadline) {
			return command, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func ListEmpresaControlElectricoInputConfigs(dbConn *sql.DB, empresaID, raspberryID int64) ([]EmpresaControlElectricoInputConfig, error) {
	rows, err := querySQLCompat(dbConn, `SELECT id, COALESCE(entrada_gpio_pin,-1), COALESCE(entrada_pull,'none'), COALESCE(debounce_ms,250), COALESCE(valor,'1'), COALESCE(sensor_codigo,''), COALESCE(accion,'alarma'), COALESCE(rele_id,0), COALESCE(estacion_id,0), COALESCE(nombre,''), COALESCE(alarma_habilitada,1), COALESCE(temporizador_segundos,0) FROM empresa_control_electrico_reglas WHERE empresa_id=? AND raspberry_id=? AND entrada_gpio_pin>=0 AND LOWER(COALESCE(estado,'activo'))='activo' ORDER BY entrada_gpio_pin,id`, empresaID, raspberryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []EmpresaControlElectricoInputConfig{}
	for rows.Next() {
		var item EmpresaControlElectricoInputConfig
		var alarm int
		if err := rows.Scan(&item.ReglaID, &item.GPIOPin, &item.Pull, &item.DebounceMS, &item.ValorActivo, &item.SensorCodigo, &item.Accion, &item.ReleID, &item.EstacionID, &item.Nombre, &alarm, &item.TemporizadorSegundos); err != nil {
			return nil, err
		}
		item.AlarmaHabilitada = alarm == 1
		result = append(result, item)
	}
	return result, rows.Err()
}

func ListEmpresaControlElectricoTraficoRaspberry(dbConn *sql.DB) ([]EmpresaControlElectricoTraficoRaspberry, error) {
	today := time.Now().UTC().Format("2006-01-02")
	monthStart := time.Now().UTC().Format("2006-01") + "-01"
	rows, err := querySQLCompat(dbConn, `SELECT r.empresa_id, r.id, COALESCE(r.codigo,''), COALESCE(r.nombre,''), COALESCE(r.device_uid,''), COALESCE(r.tunnel_enabled,0), COALESCE(r.tunnel_status,'sin_configurar'), COALESCE(r.last_seen,''), COALESCE(r.last_ip,''), COALESCE(r.agent_version,''), COALESCE(r.bytes_rx,0), COALESCE(r.bytes_tx,0), COALESCE(t.bytes_rx,0), COALESCE(t.bytes_tx,0), COALESCE(t.solicitudes,0), COALESCE(m.bytes_rx,0), COALESCE(m.bytes_tx,0), COALESCE(r.last_tunnel_error,'') FROM empresa_control_electrico_raspberry_pis r LEFT JOIN empresa_control_electrico_trafico_diario t ON t.empresa_id=r.empresa_id AND t.raspberry_id=r.id AND t.fecha=? LEFT JOIN (SELECT empresa_id,raspberry_id,SUM(bytes_rx) bytes_rx,SUM(bytes_tx) bytes_tx FROM empresa_control_electrico_trafico_diario WHERE fecha>=? GROUP BY empresa_id,raspberry_id) m ON m.empresa_id=r.empresa_id AND m.raspberry_id=r.id WHERE LOWER(COALESCE(r.estado,'activo'))='activo' ORDER BY r.empresa_id,r.nombre,r.id`, today, monthStart)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []EmpresaControlElectricoTraficoRaspberry{}
	for rows.Next() {
		var item EmpresaControlElectricoTraficoRaspberry
		var enabled int
		if err := rows.Scan(&item.EmpresaID, &item.RaspberryID, &item.Codigo, &item.Nombre, &item.DeviceUID, &enabled, &item.TunnelStatus, &item.LastSeen, &item.LastIP, &item.AgentVersion, &item.BytesRx, &item.BytesTx, &item.TodayBytesRx, &item.TodayBytesTx, &item.TodayRequests, &item.MonthBytesRx, &item.MonthBytesTx, &item.LastTunnelError); err != nil {
			return nil, err
		}
		item.TunnelEnabled = enabled == 1
		result = append(result, item)
	}
	return result, rows.Err()
}
