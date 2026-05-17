package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
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
	mantenimientoKeyAvisosJSON   = "mantenimiento_programado.avisos_json"
	defaultMantenimientoTimezone = "America/Bogota"
)

// MantenimientoPayload define los datos de configuracion de mantenimiento.
type MantenimientoPayload struct {
	AvisoID     string `json:"aviso_id"`
	Activo      bool   `json:"mantenimiento_activo"`
	AvisoActivo bool   `json:"aviso_activo"`
	Fecha       string `json:"fecha"`
	HoraInicio  string `json:"hora_inicio"`
	HoraFin     string `json:"hora_fin"`
	ZonaHoraria string `json:"zona_horaria"`
	Mensaje     string `json:"mensaje_publico"`
}

type mantenimientoConfig struct {
	AvisoID     string               `json:"aviso_id,omitempty"`
	Activo      bool                 `json:"mantenimiento_activo"`
	AvisoActivo bool                 `json:"aviso_activo"`
	Fecha       string               `json:"fecha"`
	HoraInicio  string               `json:"hora_inicio"`
	HoraFin     string               `json:"hora_fin"`
	ZonaHoraria string               `json:"zona_horaria"`
	Mensaje     string               `json:"mensaje_publico"`
	Visible     bool                 `json:"visible"`
	Avisos      []mantenimientoAviso `json:"avisos_programados"`
}

type mantenimientoAviso struct {
	ID          string `json:"id"`
	AvisoActivo bool   `json:"aviso_activo"`
	Fecha       string `json:"fecha"`
	HoraInicio  string `json:"hora_inicio"`
	HoraFin     string `json:"hora_fin"`
	ZonaHoraria string `json:"zona_horaria"`
	Mensaje     string `json:"mensaje_publico"`
	CreadoEn    string `json:"creado_en,omitempty"`
	Actualizado string `json:"actualizado_en,omitempty"`
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

		case http.MethodPost:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action != "desactivar" && action != "eliminar" {
				http.Error(w, "accion no soportada", http.StatusBadRequest)
				return
			}
			var payload struct {
				ID string `json:"id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if err := updateMantenimientoAvisoState(dbSuper, payload.ID, action); err != nil {
				log.Printf("[mantenimiento] failed to %s aviso: %v", action, err)
				http.Error(w, err.Error(), http.StatusBadRequest)
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
	cfg.Avisos = loadMantenimientoAvisos(dbSuper, cfg)
	if selected := selectMantenimientoAvisoPrincipal(cfg.Avisos); selected != nil {
		applyMantenimientoAvisoToConfig(&cfg, *selected)
	} else if len(cfg.Avisos) > 0 {
		cfg.AvisoActivo = false
		cfg.Fecha = ""
		cfg.HoraInicio = ""
		cfg.HoraFin = ""
		cfg.Mensaje = ""
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
	avisos := loadMantenimientoAvisos(dbSuper, mantenimientoConfig{})
	if shouldPersistMantenimientoAviso(cfg) {
		avisos = upsertMantenimientoAviso(avisos, mantenimientoAvisoFromConfig(cfg))
	}
	return saveMantenimientoState(dbSuper, cfg.Activo, avisos)
}

func saveMantenimientoState(dbSuper *sql.DB, activo bool, avisos []mantenimientoAviso) error {
	avisos = normalizeMantenimientoAvisos(avisos)
	if err := saveMantenimientoAvisos(dbSuper, avisos); err != nil {
		return err
	}
	current := mantenimientoConfig{Activo: activo, ZonaHoraria: defaultMantenimientoTimezone}
	if selected := selectMantenimientoAvisoPrincipal(avisos); selected != nil {
		applyMantenimientoAvisoToConfig(&current, *selected)
	}
	items := map[string]string{
		mantenimientoKeyActivo:      boolConfigValue(activo),
		mantenimientoKeyAvisoActivo: boolConfigValue(current.AvisoActivo),
		mantenimientoKeyFecha:       current.Fecha,
		mantenimientoKeyHoraInicio:  current.HoraInicio,
		mantenimientoKeyHoraFin:     current.HoraFin,
		mantenimientoKeyZonaHoraria: current.ZonaHoraria,
		mantenimientoKeyMensaje:     current.Mensaje,
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
		AvisoID:     cleanMantenimientoID(payload.AvisoID),
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

func loadMantenimientoAvisos(dbSuper *sql.DB, legacy mantenimientoConfig) []mantenimientoAviso {
	raw := strings.TrimSpace(readMantenimientoConfigValue(dbSuper, mantenimientoKeyAvisosJSON))
	var avisos []mantenimientoAviso
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &avisos); err != nil {
			log.Printf("[mantenimiento] avisos_json invalido: %v", err)
		}
	}
	avisos = normalizeMantenimientoAvisos(avisos)
	if len(avisos) == 0 && legacyHasMantenimientoAviso(legacy) {
		avisos = append(avisos, mantenimientoAvisoFromConfig(legacy))
	}
	return avisos
}

func saveMantenimientoAvisos(dbSuper *sql.DB, avisos []mantenimientoAviso) error {
	raw, err := json.Marshal(normalizeMantenimientoAvisos(avisos))
	if err != nil {
		return err
	}
	return dbpkg.SetConfigValue(dbSuper, mantenimientoKeyAvisosJSON, string(raw), false)
}

func updateMantenimientoAvisoState(dbSuper *sql.DB, rawID, action string) error {
	id := cleanMantenimientoID(rawID)
	if id == "" {
		return fmt.Errorf("id de aviso requerido")
	}
	cfg := loadMantenimientoConfig(dbSuper)
	avisos := cfg.Avisos
	next := make([]mantenimientoAviso, 0, len(avisos))
	found := false
	now := time.Now().UTC().Format(time.RFC3339)
	for _, aviso := range avisos {
		if aviso.ID != id {
			next = append(next, aviso)
			continue
		}
		found = true
		if action == "eliminar" {
			continue
		}
		aviso.AvisoActivo = false
		aviso.Actualizado = now
		next = append(next, aviso)
	}
	if !found {
		return fmt.Errorf("aviso no encontrado")
	}
	return saveMantenimientoState(dbSuper, cfg.Activo, next)
}

func normalizeMantenimientoAvisos(in []mantenimientoAviso) []mantenimientoAviso {
	out := make([]mantenimientoAviso, 0, len(in))
	seen := make(map[string]bool)
	for _, aviso := range in {
		aviso.ID = cleanMantenimientoID(aviso.ID)
		if aviso.ID == "" {
			aviso.ID = generateMantenimientoAvisoID(aviso)
		}
		if seen[aviso.ID] {
			continue
		}
		seen[aviso.ID] = true
		aviso.Fecha, _ = normalizeMantenimientoDate(aviso.Fecha)
		aviso.HoraInicio, _ = normalizeMantenimientoTime(aviso.HoraInicio, "hora_inicio")
		aviso.HoraFin, _ = normalizeMantenimientoTime(aviso.HoraFin, "hora_fin")
		aviso.ZonaHoraria = cleanMantenimientoText(aviso.ZonaHoraria, 80)
		if aviso.ZonaHoraria == "" {
			aviso.ZonaHoraria = defaultMantenimientoTimezone
		}
		aviso.Mensaje = cleanMantenimientoText(aviso.Mensaje, 260)
		if aviso.Mensaje == "" {
			aviso.Mensaje = buildMantenimientoMensaje(mantenimientoConfig{Fecha: aviso.Fecha, HoraInicio: aviso.HoraInicio, HoraFin: aviso.HoraFin})
		}
		if aviso.CreadoEn == "" {
			aviso.CreadoEn = time.Now().UTC().Format(time.RFC3339)
		}
		if aviso.Actualizado == "" {
			aviso.Actualizado = aviso.CreadoEn
		}
		out = append(out, aviso)
	}
	sort.SliceStable(out, func(i, j int) bool {
		left := out[i].Fecha + " " + out[i].HoraInicio + " " + out[i].ID
		right := out[j].Fecha + " " + out[j].HoraInicio + " " + out[j].ID
		return left < right
	})
	return out
}

func selectMantenimientoAvisoPrincipal(avisos []mantenimientoAviso) *mantenimientoAviso {
	for i := range avisos {
		if avisos[i].AvisoActivo {
			return &avisos[i]
		}
	}
	return nil
}

func applyMantenimientoAvisoToConfig(cfg *mantenimientoConfig, aviso mantenimientoAviso) {
	cfg.AvisoID = aviso.ID
	cfg.AvisoActivo = aviso.AvisoActivo
	cfg.Fecha = aviso.Fecha
	cfg.HoraInicio = aviso.HoraInicio
	cfg.HoraFin = aviso.HoraFin
	cfg.ZonaHoraria = aviso.ZonaHoraria
	cfg.Mensaje = aviso.Mensaje
}

func mantenimientoAvisoFromConfig(cfg mantenimientoConfig) mantenimientoAviso {
	now := time.Now().UTC().Format(time.RFC3339)
	aviso := mantenimientoAviso{
		ID:          cleanMantenimientoID(cfg.AvisoID),
		AvisoActivo: cfg.AvisoActivo,
		Fecha:       cfg.Fecha,
		HoraInicio:  cfg.HoraInicio,
		HoraFin:     cfg.HoraFin,
		ZonaHoraria: cfg.ZonaHoraria,
		Mensaje:     cfg.Mensaje,
		CreadoEn:    now,
		Actualizado: now,
	}
	if aviso.ID == "" {
		aviso.ID = generateMantenimientoAvisoID(aviso)
	}
	return aviso
}

func upsertMantenimientoAviso(avisos []mantenimientoAviso, aviso mantenimientoAviso) []mantenimientoAviso {
	aviso = normalizeMantenimientoAvisos([]mantenimientoAviso{aviso})[0]
	for i := range avisos {
		if avisos[i].ID == aviso.ID {
			if avisos[i].CreadoEn != "" {
				aviso.CreadoEn = avisos[i].CreadoEn
			}
			aviso.Actualizado = time.Now().UTC().Format(time.RFC3339)
			avisos[i] = aviso
			return normalizeMantenimientoAvisos(avisos)
		}
	}
	return normalizeMantenimientoAvisos(append(avisos, aviso))
}

func shouldPersistMantenimientoAviso(cfg mantenimientoConfig) bool {
	return cfg.AvisoID != "" || cfg.AvisoActivo || cfg.Fecha != "" || cfg.HoraInicio != "" || cfg.HoraFin != "" || cfg.Mensaje != ""
}

func legacyHasMantenimientoAviso(cfg mantenimientoConfig) bool {
	return cfg.AvisoActivo || cfg.Fecha != "" || cfg.HoraInicio != "" || cfg.HoraFin != "" || cfg.Mensaje != ""
}

func generateMantenimientoAvisoID(aviso mantenimientoAviso) string {
	base := strings.TrimSpace(aviso.Fecha + "-" + aviso.HoraInicio + "-" + aviso.HoraFin + "-" + aviso.Mensaje)
	if base == "" {
		base = fmt.Sprintf("aviso-%d", time.Now().UnixNano())
	}
	var b strings.Builder
	for _, r := range strings.ToLower(base) {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	id := strings.Trim(b.String(), "-")
	if id == "" {
		id = fmt.Sprintf("aviso-%d", time.Now().UnixNano())
	}
	if len(id) > 80 {
		id = id[:80]
	}
	return id
}

func cleanMantenimientoID(raw string) string {
	raw = strings.TrimSpace(raw)
	if len(raw) > 100 {
		raw = raw[:100]
	}
	var b strings.Builder
	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + ('a' - 'A'))
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_':
			b.WriteRune(r)
		}
	}
	return strings.Trim(b.String(), "-_")
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
