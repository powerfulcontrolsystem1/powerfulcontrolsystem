package handlers

import (
"encoding/json"
	"fmt"
	"net"
"net/http"
"os/exec"
"runtime"
"strings"
	"time"
)

type ServicioEstado struct {
	ID          string            `json:"id"`
	Nombre      string            `json:"nombre"`
	Estado      string            `json:"estado"`
	Detalle     string            `json:"detalle"`
	Componentes map[string]string `json:"componentes,omitempty"`
	Prueba      *ServicioPrueba   `json:"prueba,omitempty"`
}

type ServicioPrueba struct {
	OK        bool              `json:"ok"`
	Resumen   string            `json:"resumen"`
	Revisado  string            `json:"revisado"`
	Puertos   map[string]string `json:"puertos,omitempty"`
	Servicios map[string]string `json:"servicios,omitempty"`
}

func buildRustDeskServiceState(includeProbe bool) ServicioEstado {
	hbbsStatus := checkSystemctlStatus("rustdesk-hbbs")
	hbbrStatus := checkSystemctlStatus("rustdesk-hbbr")
	overall := "inactive"
	switch {
	case hbbsStatus == "active" && hbbrStatus == "active":
		overall = "active"
	case hbbsStatus == "error" || hbbrStatus == "error":
		overall = "error"
	case hbbsStatus == "active" || hbbrStatus == "active":
		overall = "degraded"
	}
	state := ServicioEstado{
		ID:     "rustdesk",
		Nombre: "RustDesk (Soporte Remoto)",
		Estado: overall,
		Detalle: "Servidor ID/Relay para soporte remoto de clientes a traves de VPS.",
		Componentes: map[string]string{
			"rustdesk-hbbs": hbbsStatus,
			"rustdesk-hbbr": hbbrStatus,
		},
	}
	if includeProbe {
		probe := probeRustDeskService()
		state.Prueba = &probe
	}
	return state
}

func SuperServidoresListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		servicios := []ServicioEstado{buildRustDeskServiceState(false)}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "servicios": servicios})
	}
}

func SuperServidoresToggleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			ID     string `json:"id"`
			Accion string `json:"accion"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		if payload.ID == "rustdesk" {
			if payload.Accion == "start" || payload.Accion == "stop" || payload.Accion == "restart" {
				err1 := runSystemctl(payload.Accion, "rustdesk-hbbs")
				err2 := runSystemctl(payload.Accion, "rustdesk-hbbr")
				if err1 != nil {
					http.Error(w, err1.Error(), http.StatusInternalServerError)
					return
				}
				if err2 != nil {
					http.Error(w, err2.Error(), http.StatusInternalServerError)
					return
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "servicio": buildRustDeskServiceState(false)})
	}
}

func SuperServidoresProbeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		service := strings.TrimSpace(r.URL.Query().Get("id"))
		if service == "" {
			service = "rustdesk"
		}
		if service != "rustdesk" {
			http.Error(w, "servicio no soportado", http.StatusBadRequest)
			return
		}
		state := buildRustDeskServiceState(true)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "servicio": state})
	}
}

func checkSystemctlStatus(service string) string {
	if runtime.GOOS == "windows" {
		return "inactive"
	}
	cmd := exec.Command("systemctl", "is-active", service)
	out, err := cmd.Output()
	if err != nil {
		return "error"
	}
	res := strings.TrimSpace(string(out))
	if res == "active" {
		return "active"
	}
	return "inactive"
}

func runSystemctl(accion string, service string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	cmd := exec.Command("sudo", "systemctl", accion, service)
	err := cmd.Run()
	return err
}

func probeRustDeskService() ServicioPrueba {
	probe := ServicioPrueba{
		OK:       false,
		Resumen:  "Comprobacion no disponible en este entorno.",
		Revisado: time.Now().In(time.Local).Format("2006-01-02 15:04:05"),
		Puertos:  map[string]string{},
		Servicios: map[string]string{
			"rustdesk-hbbs": checkSystemctlStatus("rustdesk-hbbs"),
			"rustdesk-hbbr": checkSystemctlStatus("rustdesk-hbbr"),
		},
	}
	if runtime.GOOS == "windows" {
		probe.Resumen = "Entorno local Windows: la prueba real se ejecuta en el VPS Linux donde corre RustDesk."
		return probe
	}

	ports := []int{21114, 21115, 21116, 21117, 21118, 21119}
	openPorts := 0
	for _, port := range ports {
		address := fmt.Sprintf("127.0.0.1:%d", port)
		conn, err := net.DialTimeout("tcp", address, 700*time.Millisecond)
		if err != nil {
			probe.Puertos[fmt.Sprintf("%d", port)] = "cerrado"
			continue
		}
		_ = conn.Close()
		probe.Puertos[fmt.Sprintf("%d", port)] = "abierto"
		openPorts++
	}
	if probe.Servicios["rustdesk-hbbs"] == "active" && probe.Servicios["rustdesk-hbbr"] == "active" && openPorts > 0 {
		probe.OK = true
		probe.Resumen = fmt.Sprintf("RustDesk responde: hbbs/hbbr activos y %d puerto(s) locales abiertos.", openPorts)
		return probe
	}
	probe.Resumen = fmt.Sprintf("RustDesk con alertas: hbbs=%s, hbbr=%s, puertos abiertos=%d.", probe.Servicios["rustdesk-hbbs"], probe.Servicios["rustdesk-hbbr"], openPorts)
	return probe
}
