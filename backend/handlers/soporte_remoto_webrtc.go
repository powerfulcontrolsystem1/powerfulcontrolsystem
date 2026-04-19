package handlers

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Map to hold signaling connections for each empresa
var WebrtcHub = struct {
	sync.RWMutex
	Connections map[string]*websocket.Conn
}{Connections: make(map[string]*websocket.Conn)}

func SoporteRemotoSignalingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID := r.URL.Query().Get("empresa_id")
		if empresaID == "" {
			http.Error(w, "missing empresa_id", http.StatusBadRequest)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Error upgrading websocket:", err)
			return
		}

		// Save the connection (oversimplified signaling for 1-1 session)
		WebrtcHub.Lock()
		// Determine role from query: "role=host" or "role=viewer"
		role := r.URL.Query().Get("role")
		key := empresaID + "_" + role
		WebrtcHub.Connections[key] = conn
		WebrtcHub.Unlock()

		defer func() {
			WebrtcHub.Lock()
			delete(WebrtcHub.Connections, key)
			WebrtcHub.Unlock()
			conn.Close()
		}()

		// Basic signaling: host sends to viewer, viewer sends to host
		targetRole := "viewer"
		if role == "viewer" {
			targetRole = "host"
		}

		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}

			WebrtcHub.RLock()
			targetConn, ok := WebrtcHub.Connections[empresaID+"_"+targetRole]
			WebrtcHub.RUnlock()

			if ok {
				targetConn.WriteMessage(mt, msg)
			}
		}
	}
}
