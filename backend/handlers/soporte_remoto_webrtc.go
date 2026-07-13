package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

const (
	soporteRemotoSignalingReadLimit       = int64(64 << 10)
	soporteRemotoSignalingCheckInterval   = 10 * time.Second
	soporteRemotoSignalingDefaultIdleTime = 5 * time.Minute
)

var soporteRemotoSignalingUpgrader = websocket.Upgrader{
	CheckOrigin: utils.IsSameOriginRequest,
}

type soporteRemotoSignalingPeer struct {
	conn         *websocket.Conn
	empresaID    int64
	sessionID    int64
	codigoSesion string
	role         string
	lastActivity atomic.Int64
	writeMu      sync.Mutex
}

func (p *soporteRemotoSignalingPeer) write(messageType int, payload []byte) error {
	p.writeMu.Lock()
	defer p.writeMu.Unlock()
	_ = p.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return p.conn.WriteMessage(messageType, payload)
}

func (p *soporteRemotoSignalingPeer) closeWithReason(code int, reason string) {
	p.writeMu.Lock()
	defer p.writeMu.Unlock()
	_ = p.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, reason), time.Now().Add(2*time.Second))
	_ = p.conn.Close()
}

var soporteRemotoSignalingHub = struct {
	sync.RWMutex
	peers map[string]*soporteRemotoSignalingPeer
}{peers: make(map[string]*soporteRemotoSignalingPeer)}

func soporteRemotoSignalingRole(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "host":
		return "host"
	case "viewer":
		return "viewer"
	default:
		return ""
	}
}

func soporteRemotoSignalingPeerKey(empresaID int64, codigoSesion, role string) string {
	return fmt.Sprintf("%d:%s:%s", empresaID, strings.TrimSpace(codigoSesion), soporteRemotoSignalingRole(role))
}

func soporteRemotoSignalingReserve(key string) bool {
	soporteRemotoSignalingHub.Lock()
	defer soporteRemotoSignalingHub.Unlock()
	if _, exists := soporteRemotoSignalingHub.peers[key]; exists {
		return false
	}
	soporteRemotoSignalingHub.peers[key] = nil
	return true
}

func soporteRemotoSignalingSetPeer(key string, peer *soporteRemotoSignalingPeer) {
	soporteRemotoSignalingHub.Lock()
	defer soporteRemotoSignalingHub.Unlock()
	soporteRemotoSignalingHub.peers[key] = peer
}

func soporteRemotoSignalingRelease(key string, peer *soporteRemotoSignalingPeer) {
	soporteRemotoSignalingHub.Lock()
	defer soporteRemotoSignalingHub.Unlock()
	current, exists := soporteRemotoSignalingHub.peers[key]
	if exists && (peer == nil || current == peer) {
		delete(soporteRemotoSignalingHub.peers, key)
	}
}

func soporteRemotoSignalingTarget(key string) *soporteRemotoSignalingPeer {
	soporteRemotoSignalingHub.RLock()
	defer soporteRemotoSignalingHub.RUnlock()
	return soporteRemotoSignalingHub.peers[key]
}

func soporteRemotoSignalingIdleTimeout() time.Duration {
	raw := strings.TrimSpace(os.Getenv("PCS_WEBRTC_IDLE_TIMEOUT_SECONDS"))
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds < 60 {
		return soporteRemotoSignalingDefaultIdleTime
	}
	if seconds > 1800 {
		seconds = 1800
	}
	return time.Duration(seconds) * time.Second
}

func soporteRemotoSignalingEmpresaID(r *http.Request) (int64, error) {
	contextEmpresaID := parseEmpresaIDFromContext(r)
	if contextEmpresaID <= 0 {
		return 0, errors.New("contexto de empresa requerido")
	}
	queryEmpresaID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("empresa_id")), 10, 64)
	if err != nil || queryEmpresaID != contextEmpresaID {
		return 0, errors.New("empresa fuera del contexto autorizado")
	}
	return contextEmpresaID, nil
}

func registrarAuditoriaSignalingNoBloqueante(dbEmp *sql.DB, r *http.Request, empresaID, sessionID int64, event string, status int) {
	registrarAuditoriaModuloEmpresaNoBloqueante(
		dbEmp,
		r,
		empresaID,
		"soporte_remoto",
		event,
		"webrtc_signaling",
		sessionID,
		status,
		map[string]interface{}{"canal": "websocket"},
		"evento de senalizacion WebRTC",
	)
}

// SoporteRemotoSignalingHandler exige autenticacion empresarial previa por el
// middleware, origen exacto y una credencial de senalizacion de un solo uso.
func SoporteRemotoSignalingHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if strings.TrimSpace(r.Header.Get("Origin")) == "" || !utils.IsSameOriginRequest(r) {
			http.Error(w, "origin no autorizado", http.StatusForbidden)
			return
		}
		empresaID, err := soporteRemotoSignalingEmpresaID(r)
		if err != nil {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		codigoSesion := strings.TrimSpace(r.URL.Query().Get("codigo_sesion"))
		role := soporteRemotoSignalingRole(r.URL.Query().Get("role"))
		tokenRaw := strings.TrimSpace(r.URL.Query().Get("token"))
		nonceRaw := strings.TrimSpace(r.URL.Query().Get("nonce"))
		if codigoSesion == "" || role == "" || tokenRaw == "" || nonceRaw == "" {
			http.Error(w, "credencial de senalizacion requerida", http.StatusBadRequest)
			return
		}

		peerKey := soporteRemotoSignalingPeerKey(empresaID, codigoSesion, role)
		if !soporteRemotoSignalingReserve(peerKey) {
			http.Error(w, "rol de senalizacion ya conectado", http.StatusConflict)
			return
		}
		reserved := true
		defer func() {
			if reserved {
				soporteRemotoSignalingRelease(peerKey, nil)
			}
		}()

		session, err := dbpkg.ConsumeEmpresaSoporteRemotoSignalingCredential(dbEmp, empresaID, codigoSesion, role, tokenRaw, nonceRaw)
		if err != nil {
			registrarAuditoriaSignalingNoBloqueante(dbEmp, r, empresaID, 0, "signaling_rechazado", http.StatusUnauthorized)
			http.Error(w, "credencial de senalizacion invalida", http.StatusUnauthorized)
			return
		}
		registrarAuditoriaSignalingNoBloqueante(dbEmp, r, empresaID, session.ID, "signaling_aceptado", http.StatusSwitchingProtocols)

		conn, err := soporteRemotoSignalingUpgrader.Upgrade(w, r, nil)
		if err != nil {
			registrarAuditoriaSignalingNoBloqueante(dbEmp, r, empresaID, session.ID, "signaling_upgrade_fallido", http.StatusBadRequest)
			return
		}
		peer := &soporteRemotoSignalingPeer{
			conn:         conn,
			empresaID:    empresaID,
			sessionID:    session.ID,
			codigoSesion: codigoSesion,
			role:         role,
		}
		peer.lastActivity.Store(time.Now().UnixNano())
		soporteRemotoSignalingSetPeer(peerKey, peer)
		reserved = false
		registrarAuditoriaSignalingNoBloqueante(dbEmp, r, empresaID, session.ID, "signaling_conectado", http.StatusSwitchingProtocols)

		done := make(chan struct{})
		defer func() {
			close(done)
			soporteRemotoSignalingRelease(peerKey, peer)
			_ = conn.Close()
			registrarAuditoriaSignalingNoBloqueante(dbEmp, r, empresaID, session.ID, "signaling_cerrado", http.StatusOK)
		}()

		conn.SetReadLimit(soporteRemotoSignalingReadLimit)
		_ = conn.SetReadDeadline(time.Now().Add(soporteRemotoSignalingIdleTimeout()))
		conn.SetPongHandler(func(string) error {
			return conn.SetReadDeadline(time.Now().Add(soporteRemotoSignalingIdleTimeout()))
		})

		go soporteRemotoMonitorSignalingPeer(dbEmp, peer, done)

		targetRole := "viewer"
		if role == "viewer" {
			targetRole = "host"
		}
		targetKey := soporteRemotoSignalingPeerKey(empresaID, codigoSesion, targetRole)
		for {
			messageType, payload, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if messageType != websocket.TextMessage && messageType != websocket.BinaryMessage {
				continue
			}
			peer.lastActivity.Store(time.Now().UnixNano())
			target := soporteRemotoSignalingTarget(targetKey)
			if target != nil {
				if err := target.write(messageType, payload); err != nil {
					return
				}
			}
		}
	}
}

func soporteRemotoMonitorSignalingPeer(dbEmp *sql.DB, peer *soporteRemotoSignalingPeer, done <-chan struct{}) {
	ticker := time.NewTicker(soporteRemotoSignalingCheckInterval)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			last := time.Unix(0, peer.lastActivity.Load())
			if time.Since(last) > soporteRemotoSignalingIdleTimeout() {
				peer.closeWithReason(websocket.CloseNormalClosure, "inactividad")
				return
			}
			if !dbpkg.IsEmpresaSoporteRemotoSessionActive(dbEmp, peer.empresaID, peer.sessionID) {
				peer.closeWithReason(websocket.ClosePolicyViolation, "sesion revocada")
				return
			}
			peer.writeMu.Lock()
			err := peer.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(2*time.Second))
			peer.writeMu.Unlock()
			if err != nil {
				return
			}
		}
	}
}
