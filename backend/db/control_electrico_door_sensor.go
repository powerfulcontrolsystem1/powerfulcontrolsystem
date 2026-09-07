package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

const (
	ControlElectricoUsoDomotica      = "domotica"
	ControlElectricoUsoSensorPuertas = "sensor_puertas"
	ControlElectricoDoorInputCount   = 4
	ControlElectricoDoorMaxOutputs   = 16
)

type EmpresaControlElectricoDoorScanConfig struct {
	InputPins  []int  `json:"input_pins"`
	OutputPins []int  `json:"output_pins"`
	DelayMS    int    `json:"delay_ms"`
	InputPull  string `json:"input_pull"`
}

type EmpresaControlElectricoDoorReading struct {
	OutputIndex int `json:"output_index"`
	InputIndex  int `json:"input_index"`
	Value       int `json:"value"`
}

type EmpresaControlElectricoDoorTransition struct {
	DeviceID   string
	EstacionID int64
	Previous   string
	State      string
	Changed    bool
}

func NormalizeControlElectricoUsoTipo(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "sensor_puertas", "sensor-puertas", "puertas", "door_sensor", "door-sensor":
		return ControlElectricoUsoSensorPuertas
	default:
		return ControlElectricoUsoDomotica
	}
}

func NormalizeControlElectricoDoorOutputCount(value int) int {
	if value < 1 {
		return 16
	}
	if value > ControlElectricoDoorMaxOutputs {
		return ControlElectricoDoorMaxOutputs
	}
	return value
}

func NormalizeControlElectricoDoorDelayMS(value int) int {
	if value < 10 {
		return 100
	}
	if value > 5000 {
		return 5000
	}
	return value
}

func BuildEmpresaControlElectricoDoorScanConfig(outputCount, delayMS int) EmpresaControlElectricoDoorScanConfig {
	outputCount = NormalizeControlElectricoDoorOutputCount(outputCount)
	outputs := make([]int, 0, outputCount)
	for pin := 4; pin < 4+outputCount; pin++ {
		outputs = append(outputs, pin)
	}
	return EmpresaControlElectricoDoorScanConfig{
		InputPins:  []int{0, 1, 2, 3},
		OutputPins: outputs,
		DelayMS:    NormalizeControlElectricoDoorDelayMS(delayMS),
		InputPull:  "up",
	}
}

func empresaControlElectricoDoorDeviceID(empresaID, raspberryID int64, outputIndex, inputIndex int) string {
	return NormalizeEmpresaSensorDeviceID(fmt.Sprintf("rpi-door-e%d-r%d-o%02d-i%d", empresaID, raspberryID, outputIndex, inputIndex))
}

// SyncEmpresaControlElectricoDoorSensorChannels crea los cuatro canales de
// lectura por cada salida selectora. Las asociaciones de estacion existentes
// se conservan al cambiar delay o cantidad de salidas.
func SyncEmpresaControlElectricoDoorSensorChannels(dbConn *sql.DB, empresaID, raspberryID int64, outputCount int, enabled bool, actor string) error {
	if dbConn == nil || empresaID <= 0 || raspberryID <= 0 {
		return errors.New("empresa_id y raspberry_id son obligatorios")
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := execTxSQLCompat(tx, `UPDATE empresa_sensor_puertas_devices SET estado='inactivo', fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND source_raspberry_id=?`, empresaID, raspberryID); err != nil {
		return err
	}
	if enabled {
		if _, err := execTxSQLCompat(tx, `UPDATE empresa_control_electrico_comandos SET estado='expirado', error='Raspberry reasignada a sensores de puertas', completado_en=CURRENT_TIMESTAMP WHERE empresa_id=? AND raspberry_id=? AND estado IN ('pendiente','entregado')`, empresaID, raspberryID); err != nil {
			return err
		}
		outputCount = NormalizeControlElectricoDoorOutputCount(outputCount)
		for outputIndex := 1; outputIndex <= outputCount; outputIndex++ {
			for inputIndex := 1; inputIndex <= ControlElectricoDoorInputCount; inputIndex++ {
				deviceID := empresaControlElectricoDoorDeviceID(empresaID, raspberryID, outputIndex, inputIndex)
				observaciones := fmt.Sprintf("Canal automatico Raspberry %d: OUT%d / IN%d", raspberryID, outputIndex, inputIndex)
				_, err := execTxSQLCompat(tx, `INSERT INTO empresa_sensor_puertas_devices (empresa_id, device_id, estacion_id, last_state, last_seen, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones, source_raspberry_id, selector_output, selector_input) VALUES (?, ?, NULL, '', NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, 'activo', ?, ?, ?, ?) ON CONFLICT (empresa_id, device_id) DO UPDATE SET source_raspberry_id=EXCLUDED.source_raspberry_id, selector_output=EXCLUDED.selector_output, selector_input=EXCLUDED.selector_input, fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=EXCLUDED.usuario_creador, estado='activo'`,
					empresaID, deviceID, truncateControlElectricoText(actor, 180), observaciones, raspberryID, outputIndex, inputIndex)
				if err != nil {
					return err
				}
			}
		}
	}
	return tx.Commit()
}

// ApplyEmpresaControlElectricoDoorScan actualiza un lote autenticado de
// lecturas. El canal se resuelve por empresa + Raspberry + coordenadas; el
// payload nunca puede seleccionar otra empresa ni otra estacion.
func ApplyEmpresaControlElectricoDoorScan(dbConn *sql.DB, empresaID, raspberryID int64, readings []EmpresaControlElectricoDoorReading) ([]EmpresaControlElectricoDoorTransition, error) {
	if dbConn == nil || empresaID <= 0 || raspberryID <= 0 {
		return nil, errors.New("dispositivo de sensores invalido")
	}
	if len(readings) < 1 || len(readings) > ControlElectricoDoorInputCount*ControlElectricoDoorMaxOutputs {
		return nil, errors.New("cantidad de lecturas invalida")
	}
	seen := make(map[string]struct{}, len(readings))
	for _, reading := range readings {
		if reading.OutputIndex < 1 || reading.OutputIndex > ControlElectricoDoorMaxOutputs || reading.InputIndex < 1 || reading.InputIndex > ControlElectricoDoorInputCount || (reading.Value != 0 && reading.Value != 1) {
			return nil, errors.New("canal o valor de sensor invalido")
		}
		key := fmt.Sprintf("%d:%d", reading.OutputIndex, reading.InputIndex)
		if _, duplicate := seen[key]; duplicate {
			return nil, errors.New("canal de sensor repetido")
		}
		seen[key] = struct{}{}
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	transitions := make([]EmpresaControlElectricoDoorTransition, 0, len(readings))
	for _, reading := range readings {
		var id int64
		var deviceID, previous string
		var estacionID int64
		err := queryRowTxSQLCompat(tx, `SELECT id, device_id, COALESCE(estacion_id,0), COALESCE(last_state,'') FROM empresa_sensor_puertas_devices WHERE empresa_id=? AND source_raspberry_id=? AND selector_output=? AND selector_input=? AND LOWER(COALESCE(estado,'activo'))='activo' FOR UPDATE`,
			empresaID, raspberryID, reading.OutputIndex, reading.InputIndex).Scan(&id, &deviceID, &estacionID, &previous)
		if err != nil {
			return nil, err
		}
		state := "closed"
		if reading.Value == 1 {
			state = "open"
		}
		if _, err := execTxSQLCompat(tx, `UPDATE empresa_sensor_puertas_devices SET last_state=?, last_seen=CURRENT_TIMESTAMP, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, state, empresaID, id); err != nil {
			return nil, err
		}
		if previous != state {
			raw := fmt.Sprintf("raspberry_id=%d;out=%d;in=%d;previous=%s;state=%s", raspberryID, reading.OutputIndex, reading.InputIndex, previous, state)
			if _, err := execTxSQLCompat(tx, `INSERT INTO empresa_sensor_puertas_messages (empresa_id, device_id, estacion_id, message_text, raw_text, received_at) VALUES (?, ?, NULLIF(?,0), ?, ?, CURRENT_TIMESTAMP)`, empresaID, deviceID, estacionID, state, raw); err != nil {
				return nil, err
			}
		}
		transitions = append(transitions, EmpresaControlElectricoDoorTransition{
			DeviceID: deviceID, EstacionID: estacionID, Previous: previous, State: state, Changed: previous != "" && previous != state,
		})
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return transitions, nil
}
