package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

type AyudaTicket struct {
	ID                 int64  `json:"id"`
	Codigo             string `json:"codigo"`
	EmpresaID          int64  `json:"empresa_id"`
	EmpresaNombre      string `json:"empresa_nombre"`
	SolicitanteNombre  string `json:"solicitante_nombre"`
	SolicitanteEmail   string `json:"solicitante_email"`
	ContactoTelefono   string `json:"contacto_telefono"`
	ContactoPreferido  string `json:"contacto_preferido"`
	Origen             string `json:"origen"`
	Modulo             string `json:"modulo"`
	Ruta               string `json:"ruta"`
	Asunto             string `json:"asunto"`
	Categoria          string `json:"categoria"`
	Prioridad          string `json:"prioridad"`
	Estado             string `json:"estado"`
	UltimoMensaje      string `json:"ultimo_mensaje"`
	ContextoJSON       string `json:"contexto_json"`
	AsignadoA          string `json:"asignado_a"`
	CerradoEn          string `json:"cerrado_en"`
	FechaCreacion      string `json:"fecha_creacion"`
	FechaActualizacion string `json:"fecha_actualizacion"`
	UsuarioCreador     string `json:"usuario_creador"`
}

type AyudaTicketMensaje struct {
	ID             int64  `json:"id"`
	TicketID       int64  `json:"ticket_id"`
	AutorTipo      string `json:"autor_tipo"`
	AutorNombre    string `json:"autor_nombre"`
	AutorEmail     string `json:"autor_email"`
	Mensaje        string `json:"mensaje"`
	Interno        int    `json:"interno"`
	FechaCreacion  string `json:"fecha_creacion"`
	UsuarioCreador string `json:"usuario_creador"`
}

type AyudaTicketCreateRequest struct {
	EmpresaID         int64
	EmpresaNombre     string
	SolicitanteNombre string
	SolicitanteEmail  string
	ContactoTelefono  string
	ContactoPreferido string
	Origen            string
	Modulo            string
	Ruta              string
	Asunto            string
	Categoria         string
	Prioridad         string
	Mensaje           string
	Contexto          map[string]interface{}
	UsuarioCreador    string
}

type AyudaTicketFilter struct {
	EmpresaID int64
	Estado    string
	Prioridad string
	Query     string
	Limit     int
}

type AyudaTicketDetalle struct {
	Ticket   AyudaTicket          `json:"ticket"`
	Mensajes []AyudaTicketMensaje `json:"mensajes"`
}

var (
	ayudaTicketsSchemaMu    sync.Mutex
	ayudaTicketsSchemaReady bool
)

func EnsureAyudaTicketsSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	if dbConn == nil || !isPostgresDialect() {
		return nil
	}
	ayudaTicketsSchemaMu.Lock()
	defer ayudaTicketsSchemaMu.Unlock()
	if ayudaTicketsSchemaReady {
		return nil
	}
	statements := []string{
		`CREATE TABLE IF NOT EXISTS super_tickets_ayuda (
			id BIGSERIAL PRIMARY KEY,
			codigo TEXT NOT NULL UNIQUE,
			empresa_id BIGINT,
			empresa_nombre TEXT,
			solicitante_nombre TEXT,
			solicitante_email TEXT,
			contacto_telefono TEXT,
			contacto_preferido TEXT DEFAULT 'email',
			origen TEXT DEFAULT 'sistema',
			modulo TEXT,
			ruta TEXT,
			asunto TEXT NOT NULL,
			categoria TEXT DEFAULT 'general',
			prioridad TEXT DEFAULT 'media',
			estado TEXT DEFAULT 'nuevo',
			ultimo_mensaje TEXT,
			contexto_json TEXT,
			asignado_a TEXT,
			cerrado_en TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP::text,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP::text,
			usuario_creador TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS super_ticket_ayuda_mensajes (
			id BIGSERIAL PRIMARY KEY,
			ticket_id BIGINT NOT NULL,
			autor_tipo TEXT DEFAULT 'usuario',
			autor_nombre TEXT,
			autor_email TEXT,
			mensaje TEXT NOT NULL,
			interno INTEGER DEFAULT 0,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP::text,
			usuario_creador TEXT
		)`,
		`ALTER TABLE super_tickets_ayuda ADD COLUMN IF NOT EXISTS empresa_id BIGINT`,
		`ALTER TABLE super_tickets_ayuda ADD COLUMN IF NOT EXISTS empresa_nombre TEXT`,
		`ALTER TABLE super_tickets_ayuda ADD COLUMN IF NOT EXISTS solicitante_nombre TEXT`,
		`ALTER TABLE super_tickets_ayuda ADD COLUMN IF NOT EXISTS solicitante_email TEXT`,
		`ALTER TABLE super_tickets_ayuda ADD COLUMN IF NOT EXISTS contacto_telefono TEXT`,
		`ALTER TABLE super_tickets_ayuda ADD COLUMN IF NOT EXISTS contacto_preferido TEXT DEFAULT 'email'`,
		`ALTER TABLE super_tickets_ayuda ADD COLUMN IF NOT EXISTS origen TEXT DEFAULT 'sistema'`,
		`ALTER TABLE super_tickets_ayuda ADD COLUMN IF NOT EXISTS modulo TEXT`,
		`ALTER TABLE super_tickets_ayuda ADD COLUMN IF NOT EXISTS ruta TEXT`,
		`ALTER TABLE super_tickets_ayuda ADD COLUMN IF NOT EXISTS categoria TEXT DEFAULT 'general'`,
		`ALTER TABLE super_tickets_ayuda ADD COLUMN IF NOT EXISTS prioridad TEXT DEFAULT 'media'`,
		`ALTER TABLE super_tickets_ayuda ADD COLUMN IF NOT EXISTS estado TEXT DEFAULT 'nuevo'`,
		`ALTER TABLE super_tickets_ayuda ADD COLUMN IF NOT EXISTS ultimo_mensaje TEXT`,
		`ALTER TABLE super_tickets_ayuda ADD COLUMN IF NOT EXISTS contexto_json TEXT`,
		`ALTER TABLE super_tickets_ayuda ADD COLUMN IF NOT EXISTS asignado_a TEXT`,
		`ALTER TABLE super_tickets_ayuda ADD COLUMN IF NOT EXISTS cerrado_en TEXT`,
		`ALTER TABLE super_tickets_ayuda ADD COLUMN IF NOT EXISTS fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP::text`,
		`ALTER TABLE super_tickets_ayuda ADD COLUMN IF NOT EXISTS usuario_creador TEXT`,
		`ALTER TABLE super_ticket_ayuda_mensajes ADD COLUMN IF NOT EXISTS interno INTEGER DEFAULT 0`,
		`CREATE INDEX IF NOT EXISTS ix_super_tickets_ayuda_empresa ON super_tickets_ayuda (empresa_id)`,
		`CREATE INDEX IF NOT EXISTS ix_super_tickets_ayuda_estado ON super_tickets_ayuda (estado, prioridad)`,
		`CREATE INDEX IF NOT EXISTS ix_super_ticket_ayuda_mensajes_ticket ON super_ticket_ayuda_mensajes (ticket_id, id)`,
	}
	for _, stmt := range statements {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}
	ayudaTicketsSchemaReady = true
	return nil
}

func CreateAyudaTicket(dbConn *sql.DB, req AyudaTicketCreateRequest) (AyudaTicket, error) {
	if err := EnsureAyudaTicketsSchema(dbConn); err != nil {
		return AyudaTicket{}, err
	}
	req.Asunto = truncateTextDB(strings.TrimSpace(req.Asunto), 180)
	req.Mensaje = truncateTextDB(strings.TrimSpace(req.Mensaje), 4000)
	if req.EmpresaID <= 0 {
		return AyudaTicket{}, fmt.Errorf("empresa_id es obligatorio")
	}
	if req.Asunto == "" {
		return AyudaTicket{}, fmt.Errorf("asunto es obligatorio")
	}
	if req.Mensaje == "" {
		return AyudaTicket{}, fmt.Errorf("mensaje es obligatorio")
	}
	req.Categoria = normalizeAyudaCategoria(req.Categoria)
	req.Prioridad = normalizeAyudaPrioridad(req.Prioridad)
	req.Origen = firstNonEmptyDB(strings.TrimSpace(req.Origen), "sistema")
	req.UsuarioCreador = firstNonEmptyDB(strings.TrimSpace(req.UsuarioCreador), strings.TrimSpace(req.SolicitanteEmail), "sistema")
	req.ContactoPreferido = normalizeAyudaContactoPreferido(req.ContactoPreferido)
	contextoJSON := normalizeAyudaContextoJSON(req.Contexto)
	codigo := nextAyudaTicketCodigo(dbConn)

	var id int64
	err := QueryRowCompat(dbConn, `INSERT INTO super_tickets_ayuda
		(codigo, empresa_id, empresa_nombre, solicitante_nombre, solicitante_email, contacto_telefono, contacto_preferido, origen, modulo, ruta, asunto, categoria, prioridad, estado, ultimo_mensaje, contexto_json, fecha_creacion, fecha_actualizacion, usuario_creador)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'nuevo', ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?)
		RETURNING id`,
		codigo,
		req.EmpresaID,
		truncateTextDB(strings.TrimSpace(req.EmpresaNombre), 180),
		truncateTextDB(strings.TrimSpace(req.SolicitanteNombre), 140),
		truncateTextDB(strings.TrimSpace(strings.ToLower(req.SolicitanteEmail)), 180),
		truncateTextDB(strings.TrimSpace(req.ContactoTelefono), 80),
		req.ContactoPreferido,
		truncateTextDB(req.Origen, 80),
		truncateTextDB(strings.TrimSpace(req.Modulo), 120),
		truncateTextDB(strings.TrimSpace(req.Ruta), 500),
		req.Asunto,
		req.Categoria,
		req.Prioridad,
		req.Mensaje,
		contextoJSON,
		req.UsuarioCreador,
	).Scan(&id)
	if err != nil {
		return AyudaTicket{}, err
	}
	if err := AddAyudaTicketMensaje(dbConn, id, AyudaTicketMensaje{
		AutorTipo:      "usuario",
		AutorNombre:    req.SolicitanteNombre,
		AutorEmail:     req.SolicitanteEmail,
		Mensaje:        req.Mensaje,
		UsuarioCreador: req.UsuarioCreador,
	}); err != nil {
		return AyudaTicket{}, err
	}
	return GetAyudaTicket(dbConn, id)
}

func ListAyudaTickets(dbConn *sql.DB, filter AyudaTicketFilter) ([]AyudaTicket, error) {
	if err := EnsureAyudaTicketsSchema(dbConn); err != nil {
		return nil, err
	}
	limit := filter.Limit
	if limit <= 0 || limit > 300 {
		limit = 100
	}
	where := []string{"1=1"}
	args := []interface{}{}
	if filter.EmpresaID > 0 {
		where = append(where, "empresa_id = ?")
		args = append(args, filter.EmpresaID)
	}
	if estado := normalizeAyudaEstadoFilter(filter.Estado); estado != "" {
		where = append(where, "LOWER(COALESCE(estado,'')) = ?")
		args = append(args, estado)
	}
	if prioridad := normalizeAyudaPrioridadFilter(filter.Prioridad); prioridad != "" {
		where = append(where, "LOWER(COALESCE(prioridad,'')) = ?")
		args = append(args, prioridad)
	}
	if q := strings.TrimSpace(filter.Query); q != "" {
		where = append(where, "(LOWER(COALESCE(codigo,'')) LIKE LOWER(?) OR LOWER(COALESCE(asunto,'')) LIKE LOWER(?) OR LOWER(COALESCE(empresa_nombre,'')) LIKE LOWER(?) OR LOWER(COALESCE(solicitante_email,'')) LIKE LOWER(?) OR LOWER(COALESCE(contacto_telefono,'')) LIKE LOWER(?) OR LOWER(COALESCE(categoria,'')) LIKE LOWER(?) OR LOWER(COALESCE(modulo,'')) LIKE LOWER(?))")
		like := "%" + q + "%"
		args = append(args, like, like, like, like, like, like, like)
	}
	args = append(args, limit)
	rows, err := ExecQueryCompat(dbConn, `SELECT id, COALESCE(codigo,''), COALESCE(empresa_id,0), COALESCE(empresa_nombre,''), COALESCE(solicitante_nombre,''), COALESCE(solicitante_email,''), COALESCE(contacto_telefono,''), COALESCE(contacto_preferido,''), COALESCE(origen,''), COALESCE(modulo,''), COALESCE(ruta,''), COALESCE(asunto,''), COALESCE(categoria,''), COALESCE(prioridad,''), COALESCE(estado,''), COALESCE(ultimo_mensaje,''), COALESCE(contexto_json,''), COALESCE(asignado_a,''), COALESCE(cerrado_en,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'')
		FROM super_tickets_ayuda
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY CASE LOWER(COALESCE(estado,'')) WHEN 'nuevo' THEN 0 WHEN 'en_revision' THEN 1 WHEN 'respondido' THEN 2 ELSE 3 END, id DESC
		LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []AyudaTicket{}
	for rows.Next() {
		item, err := scanAyudaTicketRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func GetAyudaTicket(dbConn *sql.DB, id int64) (AyudaTicket, error) {
	if err := EnsureAyudaTicketsSchema(dbConn); err != nil {
		return AyudaTicket{}, err
	}
	row := QueryRowCompat(dbConn, `SELECT id, COALESCE(codigo,''), COALESCE(empresa_id,0), COALESCE(empresa_nombre,''), COALESCE(solicitante_nombre,''), COALESCE(solicitante_email,''), COALESCE(contacto_telefono,''), COALESCE(contacto_preferido,''), COALESCE(origen,''), COALESCE(modulo,''), COALESCE(ruta,''), COALESCE(asunto,''), COALESCE(categoria,''), COALESCE(prioridad,''), COALESCE(estado,''), COALESCE(ultimo_mensaje,''), COALESCE(contexto_json,''), COALESCE(asignado_a,''), COALESCE(cerrado_en,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'')
		FROM super_tickets_ayuda WHERE id = ? LIMIT 1`, id)
	return scanAyudaTicketRow(row)
}

func GetAyudaTicketDetalle(dbConn *sql.DB, id int64) (AyudaTicketDetalle, error) {
	ticket, err := GetAyudaTicket(dbConn, id)
	if err != nil {
		return AyudaTicketDetalle{}, err
	}
	mensajes, err := ListAyudaTicketMensajes(dbConn, id)
	if err != nil {
		return AyudaTicketDetalle{}, err
	}
	return AyudaTicketDetalle{Ticket: ticket, Mensajes: mensajes}, nil
}

func ListAyudaTicketMensajes(dbConn *sql.DB, ticketID int64) ([]AyudaTicketMensaje, error) {
	if err := EnsureAyudaTicketsSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT id, ticket_id, COALESCE(autor_tipo,''), COALESCE(autor_nombre,''), COALESCE(autor_email,''), COALESCE(mensaje,''), COALESCE(interno,0), COALESCE(fecha_creacion,''), COALESCE(usuario_creador,'')
		FROM super_ticket_ayuda_mensajes WHERE ticket_id = ? ORDER BY id ASC`, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []AyudaTicketMensaje{}
	for rows.Next() {
		var item AyudaTicketMensaje
		if err := rows.Scan(&item.ID, &item.TicketID, &item.AutorTipo, &item.AutorNombre, &item.AutorEmail, &item.Mensaje, &item.Interno, &item.FechaCreacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func AddAyudaTicketMensaje(dbConn *sql.DB, ticketID int64, msg AyudaTicketMensaje) error {
	if err := EnsureAyudaTicketsSchema(dbConn); err != nil {
		return err
	}
	msg.Mensaje = truncateTextDB(strings.TrimSpace(msg.Mensaje), 4000)
	if ticketID <= 0 || msg.Mensaje == "" {
		return fmt.Errorf("ticket y mensaje son obligatorios")
	}
	msg.AutorTipo = normalizeAyudaAutorTipo(msg.AutorTipo)
	msg.UsuarioCreador = firstNonEmptyDB(strings.TrimSpace(msg.UsuarioCreador), strings.TrimSpace(msg.AutorEmail), "sistema")
	msg.AutorEmail = truncateTextDB(strings.TrimSpace(strings.ToLower(msg.AutorEmail)), 180)
	msg.AutorNombre = truncateTextDB(strings.TrimSpace(msg.AutorNombre), 140)
	interno := msg.Interno
	if interno != 1 {
		interno = 0
	}
	_, err := ExecCompat(dbConn, `INSERT INTO super_ticket_ayuda_mensajes
		(ticket_id, autor_tipo, autor_nombre, autor_email, mensaje, interno, fecha_creacion, usuario_creador)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?)`,
		ticketID, msg.AutorTipo, msg.AutorNombre, msg.AutorEmail, msg.Mensaje, interno, msg.UsuarioCreador)
	if err != nil {
		return err
	}
	if interno == 1 {
		_, err = ExecCompat(dbConn, `UPDATE super_tickets_ayuda
			SET fecha_actualizacion = CURRENT_TIMESTAMP
			WHERE id = ?`, ticketID)
		return err
	}
	nextState := "en_revision"
	if msg.AutorTipo == "super" {
		nextState = "respondido"
	}
	_, err = ExecCompat(dbConn, `UPDATE super_tickets_ayuda
		SET ultimo_mensaje = ?, estado = CASE WHEN LOWER(COALESCE(estado,'')) = 'cerrado' AND ? <> 'en_revision' THEN estado ELSE ? END, fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE id = ?`, msg.Mensaje, nextState, nextState, ticketID)
	return err
}

func UpdateAyudaTicketEstado(dbConn *sql.DB, ticketID int64, estado, prioridad, asignadoA, usuario string) (AyudaTicket, error) {
	if err := EnsureAyudaTicketsSchema(dbConn); err != nil {
		return AyudaTicket{}, err
	}
	estado = normalizeAyudaEstado(estado)
	prioridad = normalizeAyudaPrioridadOptional(prioridad)
	if estado == "" && prioridad == "" && strings.TrimSpace(asignadoA) == "" {
		return GetAyudaTicket(dbConn, ticketID)
	}
	assign := truncateTextDB(strings.TrimSpace(asignadoA), 180)
	usuario = firstNonEmptyDB(strings.TrimSpace(usuario), "sistema")
	if estado == "cerrado" {
		_, err := ExecCompat(dbConn, `UPDATE super_tickets_ayuda
			SET estado = CASE WHEN ? = '' THEN estado ELSE ? END,
				prioridad = CASE WHEN ? = '' THEN prioridad ELSE ? END,
				asignado_a = CASE WHEN ? = '' THEN asignado_a ELSE ? END,
				cerrado_en = CURRENT_TIMESTAMP,
				fecha_actualizacion = CURRENT_TIMESTAMP,
				usuario_creador = ?
			WHERE id = ?`, estado, estado, prioridad, prioridad, assign, assign, usuario, ticketID)
		if err != nil {
			return AyudaTicket{}, err
		}
		return GetAyudaTicket(dbConn, ticketID)
	}
	_, err := ExecCompat(dbConn, `UPDATE super_tickets_ayuda
		SET estado = CASE WHEN ? = '' THEN estado ELSE ? END,
			prioridad = CASE WHEN ? = '' THEN prioridad ELSE ? END,
			asignado_a = CASE WHEN ? = '' THEN asignado_a ELSE ? END,
			fecha_actualizacion = CURRENT_TIMESTAMP,
			usuario_creador = ?
		WHERE id = ?`, estado, estado, prioridad, prioridad, assign, assign, usuario, ticketID)
	if err != nil {
		return AyudaTicket{}, err
	}
	return GetAyudaTicket(dbConn, ticketID)
}

func nextAyudaTicketCodigo(dbConn *sql.DB) string {
	prefix := "AYU-" + time.Now().Format("20060102") + "-"
	var count int64
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM super_tickets_ayuda WHERE codigo LIKE ?`, prefix+"%").Scan(&count)
	return fmt.Sprintf("%s%05d", prefix, count+1)
}

func scanAyudaTicketRow(row *sql.Row) (AyudaTicket, error) {
	var item AyudaTicket
	err := row.Scan(&item.ID, &item.Codigo, &item.EmpresaID, &item.EmpresaNombre, &item.SolicitanteNombre, &item.SolicitanteEmail, &item.ContactoTelefono, &item.ContactoPreferido, &item.Origen, &item.Modulo, &item.Ruta, &item.Asunto, &item.Categoria, &item.Prioridad, &item.Estado, &item.UltimoMensaje, &item.ContextoJSON, &item.AsignadoA, &item.CerradoEn, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador)
	return item, err
}

func scanAyudaTicketRows(rows *sql.Rows) (AyudaTicket, error) {
	var item AyudaTicket
	err := rows.Scan(&item.ID, &item.Codigo, &item.EmpresaID, &item.EmpresaNombre, &item.SolicitanteNombre, &item.SolicitanteEmail, &item.ContactoTelefono, &item.ContactoPreferido, &item.Origen, &item.Modulo, &item.Ruta, &item.Asunto, &item.Categoria, &item.Prioridad, &item.Estado, &item.UltimoMensaje, &item.ContextoJSON, &item.AsignadoA, &item.CerradoEn, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador)
	return item, err
}

func normalizeAyudaCategoria(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "facturacion", "pagos", "licencias", "usuarios", "seguridad", "tecnico", "operacion", "configuracion":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "general"
	}
}

func normalizeAyudaPrioridad(value string) string {
	if v := normalizeAyudaPrioridadFilter(value); v != "" {
		return v
	}
	return "media"
}

func normalizeAyudaPrioridadOptional(value string) string {
	return normalizeAyudaPrioridadFilter(value)
}

func normalizeAyudaPrioridadFilter(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "baja", "media", "alta", "critica":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}

func normalizeAyudaEstado(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "nuevo", "en_revision", "respondido", "cerrado":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}

func normalizeAyudaEstadoFilter(value string) string {
	return normalizeAyudaEstado(value)
}

func normalizeAyudaAutorTipo(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "super", "sistema":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "usuario"
	}
}

func normalizeAyudaContactoPreferido(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "telefono", "whatsapp", "email":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "email"
	}
}

func normalizeAyudaContextoJSON(value map[string]interface{}) string {
	if len(value) == 0 {
		return ""
	}
	safe := make(map[string]interface{}, len(value))
	for key, raw := range value {
		k := strings.ToLower(strings.TrimSpace(key))
		switch k {
		case "titulo", "title", "url", "ruta", "modulo", "viewport", "screen", "user_agent", "idioma", "tema", "online", "hora_local":
			safe[k] = truncateTextDB(fmt.Sprint(raw), 700)
		}
	}
	if len(safe) == 0 {
		return ""
	}
	b, err := json.Marshal(safe)
	if err != nil {
		return ""
	}
	return truncateTextDB(string(b), 2500)
}

func firstNonEmptyDB(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func truncateTextDB(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 || len([]rune(value)) <= max {
		return value
	}
	runes := []rune(value)
	return strings.TrimSpace(string(runes[:max]))
}
