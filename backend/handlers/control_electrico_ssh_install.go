package handlers

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/secure"
	"golang.org/x/crypto/ssh"
)

const domoticaSSHInstallTimeout = 90 * time.Second

var errDomoticaSSHHostKey = errors.New("confirmacion de huella SSH requerida")

type domoticaSSHInstallRequest struct {
	RaspberryID         int64  `json:"raspberry_id"`
	Host                string `json:"host"`
	Port                int    `json:"port"`
	Username            string `json:"username"`
	Password            string `json:"password"`
	SudoPassword        string `json:"sudo_password"`
	HostKeyFingerprint  string `json:"host_key_fingerprint"`
	SaveCredentials     bool   `json:"save_credentials"`
	UseSavedCredentials bool   `json:"use_saved_credentials"`
}

type domoticaSSHOutput struct {
	buf   bytes.Buffer
	limit int
}

func (w *domoticaSSHOutput) Write(p []byte) (int, error) {
	original := len(p)
	remaining := w.limit - w.buf.Len()
	if remaining > 0 {
		if len(p) > remaining {
			p = p[:remaining]
		}
		_, _ = w.buf.Write(p)
	}
	return original, nil
}

func handleDomoticaRaspberrySSHInstall(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, empresaID int64, actor string) {
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	var payload domoticaSSHInstallRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "JSON invalido"})
		return
	}
	target, status, message := prepareDomoticaSSHInstallRequest(dbEmp, empresaID, &payload)
	if status != 0 {
		writeJSON(w, status, map[string]interface{}{"ok": false, "error": message})
		return
	}
	expectedFingerprint := strings.TrimSpace(payload.HostKeyFingerprint)
	client, observedFingerprint, connectErr := dialDomoticaSSH(target, payload.Username, payload.Password, expectedFingerprint)
	if connectErr != nil {
		if observedFingerprint != "" && (expectedFingerprint == "" || errors.Is(connectErr, errDomoticaSSHHostKey)) {
			message := "Confirma la huella SSH mostrada antes de instalar."
			if expectedFingerprint != "" {
				message = "La huella SSH cambio o no coincide. No se envio ninguna credencial ni instalador."
			}
			writeJSON(w, http.StatusConflict, map[string]interface{}{
				"ok": false, "requires_host_key_confirmation": true,
				"host_key_fingerprint": observedFingerprint, "error": message,
			})
			return
		}
		writeJSON(w, http.StatusBadGateway, map[string]interface{}{"ok": false, "error": "El VPS no pudo abrir la conexion SSH. Verifica IP publica/VPN, puerto, firewall y credenciales."})
		return
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(r.Context(), domoticaSSHInstallTimeout)
	defer cancel()
	lockConn, locked, err := acquireDomoticaSSHInstallLock(ctx, dbEmp, empresaID, payload.RaspberryID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "No se pudo reservar la instalacion"})
		return
	}
	if !locked {
		writeJSON(w, http.StatusConflict, map[string]interface{}{"ok": false, "error": "Ya existe una instalacion SSH en curso para esta Raspberry"})
		return
	}
	defer releaseDomoticaSSHInstallLock(lockConn, empresaID, payload.RaspberryID)

	_, installer, err := generateDomoticaRaspberryInstaller(dbEmp, empresaID, payload.RaspberryID, actor, "ssh_directo_vps")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "No se pudo preparar el instalador empresarial"})
		return
	}
	remotePath, err := randomDomoticaSSHInstallerPath()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "No se pudo preparar la instalacion"})
		return
	}
	if _, err := runDomoticaSSHSession(client, "umask 077 && cat > "+remotePath, bytes.NewReader(installer)); err != nil {
		recordDomoticaSSHInstallEvent(dbEmp, empresaID, payload.RaspberryID, actor, "error", "No se pudo transferir el instalador")
		writeJSON(w, http.StatusBadGateway, map[string]interface{}{"ok": false, "error": "No se pudo transferir el instalador por SSH"})
		return
	}
	defer func() { _, _ = runDomoticaSSHSession(client, "rm -f -- "+remotePath, nil) }()

	command := "/bin/sh " + remotePath
	var stdin io.Reader
	if payload.Username != "root" {
		command = "sudo -S -p '' /bin/sh " + remotePath
		stdin = strings.NewReader(payload.SudoPassword + "\n")
	}
	output, err := runDomoticaSSHSession(client, command, stdin)
	if err != nil {
		recordDomoticaSSHInstallEvent(dbEmp, empresaID, payload.RaspberryID, actor, "error", "La instalacion remota no finalizo")
		writeJSON(w, http.StatusBadGateway, map[string]interface{}{"ok": false, "error": domoticaSSHInstallFailureMessage(output)})
		return
	}
	connected := waitForDomoticaAgentVersion(dbEmp, empresaID, payload.RaspberryID, domoticaAgentVersion, 25*time.Second)
	credentialSaved := payload.UseSavedCredentials
	warning := ""
	if payload.SaveCredentials {
		if saveErr := dbpkg.SaveEmpresaControlElectricoSSHCredentials(dbEmp, empresaID, payload.RaspberryID, payload.Host, normalizedDomoticaSSHPort(payload.Port), payload.Username, payload.Password, payload.SudoPassword, observedFingerprint, actor); saveErr != nil {
			warning = "El agente se instalo, pero la credencial SSH no pudo guardarse cifrada."
		} else {
			credentialSaved = true
		}
	}
	recordDomoticaSSHInstallEvent(dbEmp, empresaID, payload.RaspberryID, actor, "ok", "")
	message = "Agente instalado por SSH; la Raspberry abrira el tunel HTTPS saliente."
	if connected {
		message = "Agente actualizado y tunel HTTPS confirmado."
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok": true, "connected": connected, "agent_version": domoticaAgentVersion,
		"host_key_fingerprint": observedFingerprint, "credentials_saved": credentialSaved,
		"message": message, "warning": warning,
	})
}

func prepareDomoticaSSHInstallRequest(dbEmp *sql.DB, empresaID int64, payload *domoticaSSHInstallRequest) (string, int, string) {
	pi, err := dbpkg.GetEmpresaControlElectricoRaspberryByID(dbEmp, empresaID, payload.RaspberryID, false)
	if err != nil || pi == nil || !strings.EqualFold(strings.TrimSpace(pi.TipoControlador), "raspberry_gpio") {
		return "", http.StatusNotFound, "Raspberry Pi no disponible para esta empresa"
	}
	if payload.UseSavedCredentials {
		profile, password, sudoPassword, resolveErr := dbpkg.ResolveEmpresaControlElectricoSSHCredentials(dbEmp, empresaID, payload.RaspberryID)
		if resolveErr != nil {
			return "", http.StatusBadRequest, "No se pudo usar la credencial SSH guardada. Reemplazala o verifica la clave de cifrado del servidor."
		}
		payload.Host = profile.Host
		payload.Port = profile.Port
		payload.Username = profile.Username
		payload.Password = password
		payload.SudoPassword = sudoPassword
		payload.HostKeyFingerprint = profile.HostKeyFingerprint
		payload.SaveCredentials = false
	} else if payload.SaveCredentials && !secure.EncryptionAvailable() {
		return "", http.StatusServiceUnavailable, "El cifrado seguro de credenciales no esta disponible en el servidor"
	}
	target, err := resolveDomoticaSSHTarget(payload.Host, payload.Port)
	if err != nil {
		return "", http.StatusBadRequest, err.Error()
	}
	payload.Username = strings.TrimSpace(payload.Username)
	if !validDomoticaSSHUsername(payload.Username) || len(payload.Password) < 1 || len(payload.Password) > 512 || len(payload.SudoPassword) > 512 {
		return "", http.StatusBadRequest, "Usuario o credencial SSH invalida"
	}
	if payload.SudoPassword == "" {
		payload.SudoPassword = payload.Password
	}
	return target, 0, ""
}

func handleDomoticaRaspberrySSHProfile(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, empresaID int64) {
	raspberryID, err := parseInt64QueryOptional(r, "raspberry_id")
	if err != nil || raspberryID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "raspberry_id requerido"})
		return
	}
	profile, err := dbpkg.GetEmpresaControlElectricoSSHProfile(dbEmp, empresaID, raspberryID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, sql.ErrNoRows) {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]interface{}{"ok": false, "error": "No se pudo cargar la configuracion SSH"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "ssh_config": profile})
}

func handleDomoticaRaspberrySSHProfileDelete(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, empresaID int64, actor string) {
	raspberryID, err := parseInt64QueryOptional(r, "raspberry_id")
	if err != nil || raspberryID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "raspberry_id requerido"})
		return
	}
	if err := dbpkg.DeleteEmpresaControlElectricoSSHCredentials(dbEmp, empresaID, raspberryID, actor); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, sql.ErrNoRows) {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]interface{}{"ok": false, "error": "No se pudo eliminar la credencial SSH"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

func resolveDomoticaSSHTarget(rawHost string, port int) (string, error) {
	host := strings.Trim(strings.TrimSpace(rawHost), "[]")
	ip := net.ParseIP(host)
	if ip == nil {
		return "", errors.New("La instalacion SSH exige una direccion IP literal")
	}
	if ip.IsLoopback() || ip.IsUnspecified() || ip.IsMulticast() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return "", errors.New("La direccion SSH no esta permitida")
	}
	if ip.IsPrivate() && !domoticaSSHPrivateAddressAllowed(ip) {
		return "", errors.New("La IP es privada y no pertenece a una red VPN autorizada por el VPS. Configura PCS_DOMOTICA_SSH_ALLOWED_CIDRS o usa el instalador descargable")
	}
	port = normalizedDomoticaSSHPort(port)
	if port < 1 || port > 65535 {
		return "", errors.New("Puerto SSH invalido")
	}
	return net.JoinHostPort(ip.String(), strconv.Itoa(port)), nil
}

func normalizedDomoticaSSHPort(port int) int {
	if port == 0 {
		return 22
	}
	return port
}

func domoticaSSHPrivateAddressAllowed(ip net.IP) bool {
	for _, rawCIDR := range strings.Split(os.Getenv("PCS_DOMOTICA_SSH_ALLOWED_CIDRS"), ",") {
		_, network, err := net.ParseCIDR(strings.TrimSpace(rawCIDR))
		if err == nil && network.Contains(ip) {
			return true
		}
	}
	return false
}

func validDomoticaSSHUsername(value string) bool {
	if value == "" || len(value) > 64 {
		return false
	}
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || strings.ContainsRune("._-", char) {
			continue
		}
		return false
	}
	return true
}

func dialDomoticaSSH(target, username, password, expectedFingerprint string) (*ssh.Client, string, error) {
	var observed string
	config := &ssh.ClientConfig{
		User: username, Auth: []ssh.AuthMethod{ssh.Password(password)}, Timeout: 12 * time.Second,
		HostKeyCallback: func(_ string, _ net.Addr, key ssh.PublicKey) error {
			observed = ssh.FingerprintSHA256(key)
			if expectedFingerprint == "" || subtle.ConstantTimeCompare([]byte(observed), []byte(expectedFingerprint)) != 1 {
				return errDomoticaSSHHostKey
			}
			return nil
		},
	}
	conn, err := net.DialTimeout("tcp", target, 12*time.Second)
	if err != nil {
		return nil, observed, err
	}
	if deadlineErr := conn.SetDeadline(time.Now().Add(domoticaSSHInstallTimeout)); deadlineErr != nil {
		closeErr := conn.Close()
		if closeErr != nil {
			return nil, observed, fmt.Errorf("configurar limite SSH: %w; cerrar socket: %v", deadlineErr, closeErr)
		}
		return nil, observed, deadlineErr
	}
	clientConn, channels, requests, err := ssh.NewClientConn(conn, target, config)
	if err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			return nil, observed, fmt.Errorf("abrir SSH: %w; cerrar socket: %v", err, closeErr)
		}
		return nil, observed, err
	}
	return ssh.NewClient(clientConn, channels, requests), observed, nil
}

func randomDomoticaSSHInstallerPath() (string, error) {
	raw := make([]byte, 12)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return "/tmp/pcs-domotica-" + hex.EncodeToString(raw) + ".sh", nil
}

func runDomoticaSSHSession(client *ssh.Client, command string, stdin io.Reader) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	output := &domoticaSSHOutput{limit: 16 * 1024}
	session.Stdout = output
	session.Stderr = output
	if stdin != nil {
		session.Stdin = stdin
	}
	err = session.Run(command)
	return strings.TrimSpace(output.buf.String()), err
}

func acquireDomoticaSSHInstallLock(ctx context.Context, dbConn *sql.DB, empresaID, raspberryID int64) (*sql.Conn, bool, error) {
	conn, err := dbConn.Conn(ctx)
	if err != nil {
		return nil, false, err
	}
	var locked bool
	err = conn.QueryRowContext(ctx, `SELECT pg_try_advisory_lock($1::integer,$2::integer)`, empresaID, raspberryID).Scan(&locked)
	if err != nil || !locked {
		if closeErr := conn.Close(); closeErr != nil && err == nil {
			return nil, false, closeErr
		} else if closeErr != nil {
			return nil, false, fmt.Errorf("reservar instalacion: %w; cerrar conexion: %v", err, closeErr)
		}
		return nil, locked, err
	}
	return conn, true, nil
}

func releaseDomoticaSSHInstallLock(conn *sql.Conn, empresaID, raspberryID int64) {
	if conn == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := conn.ExecContext(ctx, `SELECT pg_advisory_unlock($1::integer,$2::integer)`, empresaID, raspberryID); err != nil {
		log.Printf("[domotica_ssh] liberar lock empresa_id=%d raspberry_id=%d error: %v", empresaID, raspberryID, err)
	}
	if err := conn.Close(); err != nil {
		log.Printf("[domotica_ssh] cierre de lock empresa_id=%d raspberry_id=%d error: %v", empresaID, raspberryID, err)
	}
}

func waitForDomoticaAgentVersion(dbConn *sql.DB, empresaID, raspberryID int64, version string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		pi, err := dbpkg.GetEmpresaControlElectricoRaspberryByID(dbConn, empresaID, raspberryID, false)
		if err == nil && pi != nil && strings.TrimSpace(pi.AgentVersion) == version && domoticaTunnelSeenRecently(pi.LastSeen, 90*time.Second) {
			return true
		}
		time.Sleep(750 * time.Millisecond)
	}
	return false
}

func recordDomoticaSSHInstallEvent(dbConn *sql.DB, empresaID, raspberryID int64, actor, result, errorText string) {
	_, _ = dbpkg.InsertEmpresaControlElectricoEvento(dbConn, dbpkg.EmpresaControlElectricoEvento{
		EmpresaID: empresaID, RaspberryID: raspberryID, Comando: "instalar_ssh",
		Resultado: result, Error: errorText, Actor: actor, Origen: "ssh_directo_vps",
	})
}

func domoticaSSHInstallFailureMessage(output string) string {
	if strings.Contains(strings.ToLower(output), "sudo") {
		return "SSH conectado, pero sudo rechazo la instalacion. Verifica la clave de sudo y los permisos del usuario."
	}
	return "SSH conectado, pero la instalacion del agente no finalizo correctamente."
}
