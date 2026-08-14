package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/you/pos-backend/secure"
)

// EmpresaControlElectricoSSHProfile contiene unicamente datos no secretos.
// PasswordEnc y SudoPasswordEnc nunca forman parte de este contrato JSON.
type EmpresaControlElectricoSSHProfile struct {
	RaspberryID           int64  `json:"raspberry_id"`
	Host                  string `json:"host,omitempty"`
	Port                  int    `json:"port"`
	Username              string `json:"username,omitempty"`
	HostKeyFingerprint    string `json:"host_key_fingerprint,omitempty"`
	CredentialsConfigured bool   `json:"credentials_configured"`
	UpdatedAt             string `json:"updated_at,omitempty"`
	UpdatedBy             string `json:"updated_by,omitempty"`
}

func controlElectricoSSHPurpose(empresaID, raspberryID int64, kind string) string {
	return fmt.Sprintf("domotica-ssh-empresa-%d-raspberry-%d-%s", empresaID, raspberryID, kind)
}

// GetEmpresaControlElectricoSSHProfile obtiene solo el perfil visible y obliga
// el filtro compuesto empresa/Raspberry.
func GetEmpresaControlElectricoSSHProfile(dbConn *sql.DB, empresaID, raspberryID int64) (EmpresaControlElectricoSSHProfile, error) {
	if dbConn == nil || empresaID <= 0 || raspberryID <= 0 {
		return EmpresaControlElectricoSSHProfile{}, errors.New("empresa_id y raspberry_id son obligatorios")
	}
	profile := EmpresaControlElectricoSSHProfile{RaspberryID: raspberryID, Port: 22}
	var passwordEnc string
	err := queryRowSQLCompat(dbConn, `SELECT COALESCE(ssh_host,''), COALESCE(ssh_port,22), COALESCE(ssh_username,''), COALESCE(ssh_password_enc,''), COALESCE(ssh_host_key_fingerprint,''), COALESCE(ssh_credentials_updated_at,''), COALESCE(ssh_credentials_updated_by,'') FROM empresa_control_electrico_raspberry_pis WHERE empresa_id=? AND id=? AND LOWER(COALESCE(estado,'activo'))='activo' LIMIT 1`, empresaID, raspberryID).Scan(
		&profile.Host, &profile.Port, &profile.Username, &passwordEnc, &profile.HostKeyFingerprint, &profile.UpdatedAt, &profile.UpdatedBy,
	)
	if err != nil {
		return EmpresaControlElectricoSSHProfile{}, err
	}
	profile.CredentialsConfigured = strings.TrimSpace(passwordEnc) != ""
	return profile, nil
}

// SaveEmpresaControlElectricoSSHCredentials cifra las claves en un dominio
// criptografico exclusivo de la empresa y la Raspberry antes de persistirlas.
func SaveEmpresaControlElectricoSSHCredentials(dbConn *sql.DB, empresaID, raspberryID int64, host string, port int, username, password, sudoPassword, fingerprint, actor string) error {
	if dbConn == nil || empresaID <= 0 || raspberryID <= 0 {
		return errors.New("empresa_id y raspberry_id son obligatorios")
	}
	if password == "" {
		return errors.New("la credencial SSH es obligatoria")
	}
	if sudoPassword == "" {
		sudoPassword = password
	}
	passwordEnc, err := secure.EncryptStringForPurpose(controlElectricoSSHPurpose(empresaID, raspberryID, "password"), password)
	if err != nil {
		return fmt.Errorf("no se pudo cifrar la credencial SSH: %w", err)
	}
	sudoPasswordEnc, err := secure.EncryptStringForPurpose(controlElectricoSSHPurpose(empresaID, raspberryID, "sudo-password"), sudoPassword)
	if err != nil {
		return fmt.Errorf("no se pudo cifrar la credencial sudo: %w", err)
	}
	result, err := execSQLCompat(dbConn, `UPDATE empresa_control_electrico_raspberry_pis SET ssh_host=?, ssh_port=?, ssh_username=?, ssh_password_enc=?, ssh_sudo_password_enc=?, ssh_host_key_fingerprint=?, ssh_credentials_updated_at=CURRENT_TIMESTAMP, ssh_credentials_updated_by=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=? AND LOWER(COALESCE(estado,'activo'))='activo'`, strings.TrimSpace(host), port, strings.TrimSpace(username), passwordEnc, sudoPasswordEnc, strings.TrimSpace(fingerprint), strings.TrimSpace(actor), empresaID, raspberryID)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected != 1 {
		return sql.ErrNoRows
	}
	return nil
}

// ResolveEmpresaControlElectricoSSHCredentials descifra solo para una
// instalacion autorizada del mismo tenant y dispositivo.
func ResolveEmpresaControlElectricoSSHCredentials(dbConn *sql.DB, empresaID, raspberryID int64) (EmpresaControlElectricoSSHProfile, string, string, error) {
	profile := EmpresaControlElectricoSSHProfile{RaspberryID: raspberryID, Port: 22}
	var passwordEnc, sudoPasswordEnc string
	err := queryRowSQLCompat(dbConn, `SELECT COALESCE(ssh_host,''), COALESCE(ssh_port,22), COALESCE(ssh_username,''), COALESCE(ssh_password_enc,''), COALESCE(ssh_sudo_password_enc,''), COALESCE(ssh_host_key_fingerprint,''), COALESCE(ssh_credentials_updated_at,''), COALESCE(ssh_credentials_updated_by,'') FROM empresa_control_electrico_raspberry_pis WHERE empresa_id=? AND id=? AND LOWER(COALESCE(estado,'activo'))='activo' LIMIT 1`, empresaID, raspberryID).Scan(
		&profile.Host, &profile.Port, &profile.Username, &passwordEnc, &sudoPasswordEnc, &profile.HostKeyFingerprint, &profile.UpdatedAt, &profile.UpdatedBy,
	)
	if err != nil {
		return EmpresaControlElectricoSSHProfile{}, "", "", err
	}
	if strings.TrimSpace(passwordEnc) == "" {
		return EmpresaControlElectricoSSHProfile{}, "", "", errors.New("no hay credencial SSH guardada")
	}
	password, err := secure.DecryptStringForPurpose(controlElectricoSSHPurpose(empresaID, raspberryID, "password"), passwordEnc)
	if err != nil {
		return EmpresaControlElectricoSSHProfile{}, "", "", fmt.Errorf("no se pudo descifrar la credencial SSH")
	}
	sudoPassword, err := secure.DecryptStringForPurpose(controlElectricoSSHPurpose(empresaID, raspberryID, "sudo-password"), sudoPasswordEnc)
	if err != nil {
		return EmpresaControlElectricoSSHProfile{}, "", "", fmt.Errorf("no se pudo descifrar la credencial sudo")
	}
	profile.CredentialsConfigured = true
	return profile, password, sudoPassword, nil
}

func DeleteEmpresaControlElectricoSSHCredentials(dbConn *sql.DB, empresaID, raspberryID int64, actor string) error {
	if dbConn == nil || empresaID <= 0 || raspberryID <= 0 {
		return errors.New("empresa_id y raspberry_id son obligatorios")
	}
	result, err := execSQLCompat(dbConn, `UPDATE empresa_control_electrico_raspberry_pis SET ssh_password_enc=NULL, ssh_sudo_password_enc=NULL, ssh_credentials_updated_at=CURRENT_TIMESTAMP, ssh_credentials_updated_by=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=? AND LOWER(COALESCE(estado,'activo'))='activo'`, strings.TrimSpace(actor), empresaID, raspberryID)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected != 1 {
		return sql.ErrNoRows
	}
	return nil
}
