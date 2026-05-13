package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	mantenimientoKeyActivo       = "mantenimiento_activo"
	mantenimientoKeyAvisoActivo  = "mantenimiento_programado.aviso_activo"
	mantenimientoKeyFecha        = "mantenimiento_programado.fecha"
	mantenimientoKeyHoraInicio   = "mantenimiento_programado.hora_inicio"
	mantenimientoKeyHoraFin      = "mantenimiento_programado.hora_fin"
	mantenimientoKeyZonaHoraria  = "mantenimiento_programado.zona_horaria"
	mantenimientoKeyMensaje      = "mantenimiento_programado.mensaje_publico"
	defaultMantenimientoTimezone = "America/Bogota"
)

// MantenimientoPayload define los datos de configuracion de mantenimiento.
type MantenimientoPayload struct {
	Activo      bool   `json:"mantenimiento_activo"`
	AvisoActivo bool   `json:"aviso_activo"`
	Fecha       string `json:"fecha"`
	HoraInicio  string `json:"hora_inicio"`
	HoraFin     string `json:"hora_fin"`
	ZonaHoraria string `json:"zona_horaria"`
	Mensaje     string `json:"mensaje_publico"`
}

type mantenimientoConfig struct {
	Activo      bool   `json:"mantenimiento_activo"`
	AvisoActivo bool   `json:"aviso_activo"`
	Fecha       string `json:"fecha"`
	HoraInicio  string `json:"hora_inicio"`
	HoraFin     string `json:"hora_fin"`
	ZonaHoraria string `json:"zona_horaria"`
	Mensaje     string `json:"mensaje_publico"`
	Visible     bool   `json:"visible"`
}

// SuperMantenimientoConfigHandler maneja GET y PUT de modo mantenimiento.
func SuperMantenimientoConfigHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, loadMantenimientoConfig(dbSuper))
			return

		case http.MethodPut:
			var payload MantenimientoPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			cfg, err := normalizeMantenimientoPayload(payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if err := saveMantenimientoConfig(dbSuper, cfg); err != nil {
				log.Printf("[mantenimiento] failed to update: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "config": loadMantenimientoConfig(dbSuper)})
			return

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// EmpresaMantenimientoProgramadoHandler publica el aviso activo para el panel de empresa.
func EmpresaMantenimientoProgramadoHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		cfg := loadMantenimientoConfig(dbSuper)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":                   true,
			"visible":              cfg.Visible,
			"aviso_activo":         cfg.AvisoActivo,
			"fecha":                cfg.Fecha,
			"hora_inicio":          cfg.HoraInicio,
			"hora_fin":             cfg.HoraFin,
			"zona_horaria":         cfg.ZonaHoraria,
			"mensaje_publico":      cfg.Mensaje,
			"mantenimiento_activo": cfg.Activo,
		})
	}
}

func loadMantenimientoConfig(dbSuper *sql.DB) mantenimientoConfig {
	cfg := mantenimientoConfig{
		ZonaHoraria: defaultMantenimientoTimezone,
	}
	cfg.Activo = strings.EqualFold(readMantenimientoConfigValue(dbSuper, mantenimientoKeyActivo), "true")
	cfg.AvisoActivo = strings.EqualFold(readMantenimientoConfigValue(dbSuper, mantenimientoKeyAvisoActivo), "true")
	cfg.Fecha = strings.TrimSpace(readMantenimientoConfigValue(dbSuper, mantenimientoKeyFecha))
	cfg.HoraInicio = strings.TrimSpace(readMantenimientoConfigValue(dbSuper, mantenimientoKeyHoraInicio))
	cfg.HoraFin = strings.TrimSpace(readMantenimientoConfigValue(dbSuper, mantenimientoKeyHoraFin))
	if tz := strings.TrimSpace(readMantenimientoConfigValue(dbSuper, mantenimientoKeyZonaHoraria)); tz != "" {
		cfg.ZonaHoraria = tz
	}
	cfg.Mensaje = strings.TrimSpace(readMantenimientoConfigValue(dbSuper, mantenimientoKeyMensaje))
	if cfg.Mensaje == "" {
		cfg.Mensaje = buildMantenimientoMensaje(cfg)
	}
	cfg.Visible = cfg.AvisoActivo && strings.TrimSpace(cfg.Mensaje) != ""
	return cfg
}

func readMantenimientoConfigValue(dbSuper *sql.DB, key string) string {
	if dbSuper == nil {
		return ""
	}
	val, _, err := dbpkg.GetConfigValue(dbSuper, key)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("[mantenimiento] error reading key=%s: %v", key, err)
		}
		return ""
	}
	return val
}

func saveMantenimientoConfig(dbSuper *sql.DB, cfg mantenimientoConfig) error {
	items := map[string]string{
		mantenimientoKeyActivo:      boolConfigValue(cfg.Activo),
		mantenimientoKeyAvisoActivo: boolConfigValue(cfg.AvisoActivo),
		mantenimientoKeyFecha:       cfg.Fecha,
		mantenimientoKeyHoraInicio:  cfg.HoraInicio,
		mantenimientoKeyHoraFin:     cfg.HoraFin,
		mantenimientoKeyZonaHoraria: cfg.ZonaHoraria,
		mantenimientoKeyMensaje:     cfg.Mensaje,
	}
	for key, value := range items {
		if err := dbpkg.SetConfigValue(dbSuper, key, value, false); err != nil {
			return err
		}
	}
	return nil
}

func normalizeMantenimientoPayload(payload MantenimientoPayload) (mantenimientoConfig, error) {
	cfg := mantenimientoConfig{
		Activo:      payload.Activo,
		AvisoActivo: payload.AvisoActivo,
		ZonaHoraria: cleanMantenimientoText(payload.ZonaHoraria, 80),
		Mensaje:     cleanMantenimientoText(payload.Mensaje, 260),
	}
	if cfg.ZonaHoraria == "" {
		cfg.ZonaHoraria = defaultMantenimientoTimezone
	}
	var err error
	cfg.Fecha, err = normalizeMantenimientoDate(payload.Fecha)
	if err != nil {
		return cfg, err
	}
	cfg.HoraInicio, err = normalizeMantenimientoTime(payload.HoraInicio, "hora_inicio")
	if err != nil {
		return cfg, err
	}
	cfg.HoraFin, err = normalizeMantenimientoTime(payload.HoraFin, "hora_fin")
	if err != nil {
		return cfg, err
	}
	if cfg.AvisoActivo && (cfg.Fecha == "" || cfg.HoraInicio == "" || cfg.HoraFin == "") {
		return cfg, fmt.Errorf("fecha, hora_inicio y hora_fin son obligatorias para publicar el aviso")
	}
	return cfg, nil
}

func normalizeMantenimientoDate(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return "", fmt.Errorf("fecha invalida, usa YYYY-MM-DD")
	}
	return t.Format("2006-01-02"), nil
}

func normalizeMantenimientoTime(raw, field string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	t, err := time.Parse("15:04", raw)
	if err != nil {
		return "", fmt.Errorf("%s invalida, usa HH:MM", field)
	}
	return t.Format("15:04"), nil
}

func cleanMantenimientoText(raw string, maxLen int) string {
	clean := strings.Join(strings.Fields(strings.TrimSpace(raw)), " ")
	if maxLen > 0 && len(clean) > maxLen {
		clean = clean[:maxLen]
	}
	return clean
}

func buildMantenimientoMensaje(cfg mantenimientoConfig) string {
	if cfg.Fecha == "" || cfg.HoraInicio == "" || cfg.HoraFin == "" {
		return ""
	}
	return fmt.Sprintf("Mantenimiento programado para el %s, de %s a %s.",
		formatMantenimientoDateSpanish(cfg.Fecha),
		formatMantenimientoTimeSpanish(cfg.HoraInicio),
		formatMantenimientoTimeSpanish(cfg.HoraFin),
	)
}

func formatMantenimientoDateSpanish(raw string) string {
	t, err := time.Parse("2006-01-02", strings.TrimSpace(raw))
	if err != nil {
		return raw
	}
	months := []string{"enero", "febrero", "marzo", "abril", "mayo", "junio", "julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre"}
	return fmt.Sprintf("%d de %s de %d", t.Day(), months[int(t.Month())-1], t.Year())
}

func formatMantenimientoTimeSpanish(raw string) string {
	t, err := time.Parse("15:04", strings.TrimSpace(raw))
	if err != nil {
		return raw
	}
	hour := t.Hour()
	minute := t.Minute()
	suffix := "a. m."
	if hour >= 12 {
		suffix = "p. m."
	}
	hour12 := hour % 12
	if hour12 == 0 {
		hour12 = 12
	}
	if minute == 0 {
		return fmt.Sprintf("%d %s", hour12, suffix)
	}
	return fmt.Sprintf("%d:%02d %s", hour12, minute, suffix)
}

func boolConfigValue(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
