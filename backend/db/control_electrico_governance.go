package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	DefaultControlElectricoTunnelMonthlyLimitMB = int64(2048)
	DefaultControlElectricoTunnelWarningPercent = 80
)

var ErrControlElectricoTunnelBandwidthExceeded = errors.New("limite mensual del tunel alcanzado")

type EmpresaControlElectricoTunnelPolicy struct {
	EmpresaID          int64  `json:"empresa_id"`
	LimiteMensualMB    int64  `json:"limite_mensual_mb"`
	AlertaPorcentaje   int    `json:"alerta_porcentaje"`
	BloquearAlSuperar  bool   `json:"bloquear_al_superar"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

type EmpresaControlElectricoTunnelBandwidthStatus struct {
	EmpresaID       int64  `json:"empresa_id"`
	Mes             string `json:"mes"`
	UsoBytes        int64  `json:"uso_bytes"`
	LimiteBytes     int64  `json:"limite_bytes"`
	LimiteMensualMB int64  `json:"limite_mensual_mb"`
	Porcentaje      int    `json:"porcentaje"`
	Nivel           string `json:"nivel"`
	Bloqueado       bool   `json:"bloqueado"`
}

type EmpresaControlElectricoDisconnectCandidate struct {
	EmpresaID       int64  `json:"empresa_id"`
	RaspberryID     int64  `json:"raspberry_id"`
	RaspberryNombre string `json:"raspberry_nombre"`
	DeviceUID       string `json:"device_uid"`
	LastSeen        string `json:"last_seen"`
	AlertEmail      string `json:"alert_email"`
	GraceMinutes    int    `json:"grace_minutes"`
}

func defaultEmpresaControlElectricoTunnelPolicy(empresaID int64) EmpresaControlElectricoTunnelPolicy {
	return EmpresaControlElectricoTunnelPolicy{
		EmpresaID:         empresaID,
		LimiteMensualMB:   DefaultControlElectricoTunnelMonthlyLimitMB,
		AlertaPorcentaje:  DefaultControlElectricoTunnelWarningPercent,
		BloquearAlSuperar: true,
	}
}

func normalizeEmpresaControlElectricoTunnelPolicy(policy *EmpresaControlElectricoTunnelPolicy) {
	if policy == nil {
		return
	}
	if policy.LimiteMensualMB < 50 {
		policy.LimiteMensualMB = 50
	}
	if policy.LimiteMensualMB > 1024*1024 {
		policy.LimiteMensualMB = 1024 * 1024
	}
	if policy.AlertaPorcentaje < 10 {
		policy.AlertaPorcentaje = 10
	}
	if policy.AlertaPorcentaje > 100 {
		policy.AlertaPorcentaje = 100
	}
	policy.UsuarioCreador = truncateControlElectricoText(strings.TrimSpace(policy.UsuarioCreador), 180)
	policy.Observaciones = truncateControlElectricoText(strings.TrimSpace(policy.Observaciones), 500)
}

func GetEmpresaControlElectricoTunnelPolicy(dbConn *sql.DB, empresaID int64) (EmpresaControlElectricoTunnelPolicy, error) {
	if dbConn == nil || empresaID <= 0 {
		return EmpresaControlElectricoTunnelPolicy{}, errors.New("empresa_id invalido")
	}
	policy := defaultEmpresaControlElectricoTunnelPolicy(empresaID)
	var block int
	err := queryRowSQLCompat(dbConn, `SELECT empresa_id, COALESCE(limite_mensual_mb,2048), COALESCE(alerta_porcentaje,80), COALESCE(bloquear_al_superar,1), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,''), COALESCE(observaciones,'') FROM empresa_control_electrico_limites_tunel WHERE empresa_id=? LIMIT 1`, empresaID).
		Scan(&policy.EmpresaID, &policy.LimiteMensualMB, &policy.AlertaPorcentaje, &block, &policy.FechaActualizacion, &policy.UsuarioCreador, &policy.Observaciones)
	if errors.Is(err, sql.ErrNoRows) {
		return policy, nil
	}
	if err != nil {
		return EmpresaControlElectricoTunnelPolicy{}, err
	}
	policy.BloquearAlSuperar = block == 1
	normalizeEmpresaControlElectricoTunnelPolicy(&policy)
	return policy, nil
}

func UpsertEmpresaControlElectricoTunnelPolicy(dbConn *sql.DB, policy EmpresaControlElectricoTunnelPolicy) (EmpresaControlElectricoTunnelPolicy, error) {
	if dbConn == nil || policy.EmpresaID <= 0 {
		return EmpresaControlElectricoTunnelPolicy{}, errors.New("empresa_id invalido")
	}
	normalizeEmpresaControlElectricoTunnelPolicy(&policy)
	_, err := execSQLCompat(dbConn, `INSERT INTO empresa_control_electrico_limites_tunel (empresa_id, limite_mensual_mb, alerta_porcentaje, bloquear_al_superar, fecha_creacion, fecha_actualizacion, usuario_creador, observaciones) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?) ON CONFLICT (empresa_id) DO UPDATE SET limite_mensual_mb=EXCLUDED.limite_mensual_mb, alerta_porcentaje=EXCLUDED.alerta_porcentaje, bloquear_al_superar=EXCLUDED.bloquear_al_superar, fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=EXCLUDED.usuario_creador, observaciones=EXCLUDED.observaciones`,
		policy.EmpresaID, policy.LimiteMensualMB, policy.AlertaPorcentaje, boolInt(policy.BloquearAlSuperar), policy.UsuarioCreador, policy.Observaciones)
	if err != nil {
		return EmpresaControlElectricoTunnelPolicy{}, err
	}
	return GetEmpresaControlElectricoTunnelPolicy(dbConn, policy.EmpresaID)
}

func EmpresaControlElectricoTunnelMonthUsage(dbConn *sql.DB, empresaID int64, now time.Time) (int64, error) {
	if dbConn == nil || empresaID <= 0 {
		return 0, errors.New("empresa_id invalido")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	monthStart := time.Date(now.UTC().Year(), now.UTC().Month(), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	var total int64
	err := queryRowSQLCompat(dbConn, `SELECT COALESCE(SUM(COALESCE(bytes_rx,0)+COALESCE(bytes_tx,0)),0) FROM empresa_control_electrico_trafico_diario WHERE empresa_id=? AND fecha>=?`, empresaID, monthStart).Scan(&total)
	return total, err
}

func EvaluateEmpresaControlElectricoTunnelBandwidth(dbConn *sql.DB, empresaID int64, now time.Time) (EmpresaControlElectricoTunnelBandwidthStatus, error) {
	policy, err := GetEmpresaControlElectricoTunnelPolicy(dbConn, empresaID)
	if err != nil {
		return EmpresaControlElectricoTunnelBandwidthStatus{}, err
	}
	usage, err := EmpresaControlElectricoTunnelMonthUsage(dbConn, empresaID, now)
	if err != nil {
		return EmpresaControlElectricoTunnelBandwidthStatus{}, err
	}
	return BuildEmpresaControlElectricoTunnelBandwidthStatus(policy, usage, now), nil
}

// BuildEmpresaControlElectricoTunnelBandwidthStatus centraliza la clasificación
// usada tanto por el túnel como por el panel de Super Administrador.
func BuildEmpresaControlElectricoTunnelBandwidthStatus(policy EmpresaControlElectricoTunnelPolicy, usage int64, now time.Time) EmpresaControlElectricoTunnelBandwidthStatus {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	normalizeEmpresaControlElectricoTunnelPolicy(&policy)
	if usage < 0 {
		usage = 0
	}
	limitBytes := policy.LimiteMensualMB * 1024 * 1024
	percentage := 0
	if limitBytes > 0 {
		percentage = int((usage * 100) / limitBytes)
	}
	level := "normal"
	if percentage >= policy.AlertaPorcentaje {
		level = "advertencia"
	}
	exceeded := limitBytes > 0 && usage >= limitBytes
	if exceeded {
		level = "excedido"
	}
	return EmpresaControlElectricoTunnelBandwidthStatus{
		EmpresaID: policy.EmpresaID, Mes: now.UTC().Format("2006-01"), UsoBytes: usage,
		LimiteBytes: limitBytes, LimiteMensualMB: policy.LimiteMensualMB,
		Porcentaje: percentage, Nivel: level, Bloqueado: exceeded && policy.BloquearAlSuperar,
	}
}

func CheckEmpresaControlElectricoTunnelBandwidth(dbConn *sql.DB, empresaID int64, now time.Time) (EmpresaControlElectricoTunnelBandwidthStatus, error) {
	status, err := EvaluateEmpresaControlElectricoTunnelBandwidth(dbConn, empresaID, now)
	if err != nil {
		return status, err
	}
	if status.Bloqueado {
		return status, ErrControlElectricoTunnelBandwidthExceeded
	}
	return status, nil
}

func ListEmpresaControlElectricoDisconnectCandidates(dbConn *sql.DB) ([]EmpresaControlElectricoDisconnectCandidate, error) {
	if dbConn == nil {
		return nil, errors.New("conexion no disponible")
	}
	rows, err := querySQLCompat(dbConn, `SELECT r.empresa_id, r.id, COALESCE(r.nombre,''), COALESCE(r.device_uid,''), COALESCE(r.last_seen,''), COALESCE(c.disconnect_alert_email,''), COALESCE(c.disconnect_grace_minutes,5) FROM empresa_control_electrico_raspberry_pis r JOIN empresa_control_electrico_config c ON c.empresa_id=r.empresa_id WHERE COALESCE(c.disconnect_alert_enabled,0)=1 AND COALESCE(r.tunnel_enabled,0)=1 AND LOWER(COALESCE(r.estado,'activo'))='activo' AND NULLIF(BTRIM(COALESCE(r.last_seen,'')),'') IS NOT NULL AND CAST(NULLIF(r.last_seen,'') AS TIMESTAMP) <= CURRENT_TIMESTAMP-(COALESCE(c.disconnect_grace_minutes,5)*INTERVAL '1 minute') AND COALESCE(r.disconnect_alerted_for_last_seen,'')<>COALESCE(r.last_seen,'') ORDER BY r.empresa_id,r.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []EmpresaControlElectricoDisconnectCandidate{}
	for rows.Next() {
		var item EmpresaControlElectricoDisconnectCandidate
		if err := rows.Scan(&item.EmpresaID, &item.RaspberryID, &item.RaspberryNombre, &item.DeviceUID, &item.LastSeen, &item.AlertEmail, &item.GraceMinutes); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func MarkEmpresaControlElectricoDisconnectAlerted(dbConn *sql.DB, candidate EmpresaControlElectricoDisconnectCandidate) error {
	if dbConn == nil || candidate.EmpresaID <= 0 || candidate.RaspberryID <= 0 || strings.TrimSpace(candidate.LastSeen) == "" {
		return errors.New("candidato de desconexion invalido")
	}
	result, err := execSQLCompat(dbConn, `UPDATE empresa_control_electrico_raspberry_pis SET disconnect_alerted_for_last_seen=?, disconnect_alerted_at=CURRENT_TIMESTAMP, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=? AND COALESCE(last_seen,'')=? AND COALESCE(disconnect_alerted_for_last_seen,'')<>?`, candidate.LastSeen, candidate.EmpresaID, candidate.RaspberryID, candidate.LastSeen, candidate.LastSeen)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected != 1 {
		return fmt.Errorf("la Raspberry cambio de estado antes de registrar la alerta")
	}
	return nil
}
