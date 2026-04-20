package handlers

import (
"encoding/json"
"net/http"
"os/exec"
"runtime"
"strings"
)

type ServicioEstado struct {
ID     string `json:"id"`
Nombre string `json:"nombre"`
Estado  string `json:"estado"`
Detalle string `json:"detalle"`
}

func SuperServidoresListHandler() http.HandlerFunc {
return func(w http.ResponseWriter, r *http.Request) {
rustdeskStatus := checkSystemctlStatus("rustdesk-hbbs")

servicios := []ServicioEstado{
{
ID:     "rustdesk",
Nombre: "RustDesk (Soporte Remoto)",
Estado:  rustdeskStatus,
Detalle: "Servidor ID/Relay para soporte remoto de clientes a travÃ©s de VPS.",
},
}

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
json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}
}

func checkSystemctlStatus(service string) string {
if runtime.GOOS == "windows" {
return "inactive" // dev fallback
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
return nil // dev fallback - fake success
}
cmd := exec.Command("sudo", "systemctl", accion, service)
err := cmd.Run()
return err
}
