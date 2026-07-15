package db

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// EmpresaBuzonActor identifica al usuario/admin autenticado dentro de una empresa.
type EmpresaBuzonActor struct {
	Tipo      string `json:"tipo"`
	Ref       string `json:"ref"`
	Email     string `json:"email,omitempty"`
	Nombre    string `json:"nombre,omitempty"`
	Rol       string `json:"rol,omitempty"`
	UsuarioID int64  `json:"usuario_id,omitempty"`
}

// EmpresaBuzonMensaje es un mensaje privado visible solo para su destinatario.
type EmpresaBuzonMensaje struct {
	ID                     int64                 `json:"id"`
	EmpresaID              int64                 `json:"empresa_id"`
	DestinatarioTipo       string                `json:"destinatario_tipo"`
	DestinatarioRef        string                `json:"destinatario_ref"`
	DestinatarioEmail      string                `json:"destinatario_email,omitempty"`
	DestinatarioNombre     string                `json:"destinatario_nombre,omitempty"`
	RemitenteTipo          string                `json:"remitente_tipo,omitempty"`
	RemitenteRef           string                `json:"remitente_ref,omitempty"`
	RemitenteEmail         string                `json:"remitente_email,omitempty"`
	RemitenteNombre        string                `json:"remitente_nombre,omitempty"`
	Titulo                 string                `json:"titulo"`
	Mensaje                string                `json:"mensaje"`
	Tipo                   string                `json:"tipo,omitempty"`
	Prioridad              string                `json:"prioridad,omitempty"`
	Modulo                 string                `json:"modulo,omitempty"`
	ReferenciaTipo         string                `json:"referencia_tipo,omitempty"`
	ReferenciaID           int64                 `json:"referencia_id,omitempty"`
	EnlaceURL              string                `json:"enlace_url,omitempty"`
	TareaEstado            string                `json:"tarea_estado,omitempty"`
	TareaVenceEn           string                `json:"tarea_vence_en,omitempty"`
	TareaCerradaEn         string                `json:"tarea_cerrada_en,omitempty"`
	TareaCierreDescripcion string                `json:"tarea_cierre_descripcion,omitempty"`
	Leido                  int                   `json:"leido"`
	LeidoEn                string                `json:"leido_en,omitempty"`
	Estado                 string                `json:"estado,omitempty"`
	FechaCreacion          string                `json:"fecha_creacion,omitempty"`
	FechaActualizacion     string                `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador         string                `json:"usuario_creador,omitempty"`
	Adjuntos               []EmpresaBuzonAdjunto `json:"adjuntos,omitempty"`
}

// EmpresaBuzonAdjunto representa un archivo, foto o audio asociado a un mensaje del buzon.
type EmpresaBuzonAdjunto struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	MensajeID          int64   `json:"mensaje_id"`
	TipoArchivo        string  `json:"tipo_archivo,omitempty"`
	NombreArchivo      string  `json:"nombre_archivo,omitempty"`
	MimeType           string  `json:"mime_type,omitempty"`
	FileURL            string  `json:"file_url"`
	TamanoBytes        int64   `json:"tamano_bytes,omitempty"`
	DuracionSegundos   float64 `json:"duracion_segundos,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// EmpresaChatMensaje es un mensaje del chat general de la empresa.
type EmpresaChatMensaje struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	RemitenteTipo      string `json:"remitente_tipo"`
	RemitenteRef       string `json:"remitente_ref"`
	RemitenteEmail     string `json:"remitente_email,omitempty"`
	RemitenteNombre    string `json:"remitente_nombre,omitempty"`
	Mensaje            string `json:"mensaje"`
	Estado             string `json:"estado,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
}

type EmpresaBodegaTransferNotificationResult struct {
	Created    int      `json:"created"`
	Recipients []string `json:"recipients,omitempty"`
}

var (
	empresaBuzonSchemaMu    sync.Mutex
	empresaBuzonSchemaReady bool
)

func EnsureEmpresaBuzonSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	if dbConn == nil {
		return fmt.Errorf("db connection is required")
	}
	empresaBuzonSchemaMu.Lock()
	defer empresaBuzonSchemaMu.Unlock()
	if empresaBuzonSchemaReady {
		return nil
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_buzon_mensajes (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			destinatario_tipo TEXT NOT NULL DEFAULT 'usuario',
			destinatario_ref TEXT NOT NULL,
			destinatario_email TEXT,
			destinatario_nombre TEXT,
			remitente_tipo TEXT DEFAULT 'sistema',
			remitente_ref TEXT,
			remitente_email TEXT,
			remitente_nombre TEXT,
			titulo TEXT NOT NULL,
			mensaje TEXT NOT NULL,
			tipo TEXT DEFAULT 'general',
			prioridad TEXT DEFAULT 'normal',
			modulo TEXT,
			referencia_tipo TEXT,
			referencia_id BIGINT DEFAULT 0,
			enlace_url TEXT,
			tarea_estado TEXT DEFAULT '',
			tarea_vence_en TEXT,
			tarea_cerrada_en TEXT,
			tarea_cierre_descripcion TEXT,
			leido INTEGER DEFAULT 0,
			leido_en TEXT,
			estado TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP::text,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP::text,
			usuario_creador TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_chat_mensajes (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			remitente_tipo TEXT NOT NULL DEFAULT 'usuario',
			remitente_ref TEXT NOT NULL,
			remitente_email TEXT,
			remitente_nombre TEXT,
			mensaje TEXT NOT NULL,
			estado TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP::text,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP::text,
			usuario_creador TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_buzon_adjuntos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			mensaje_id BIGINT NOT NULL,
			tipo_archivo TEXT DEFAULT 'archivo',
			nombre_archivo TEXT,
			mime_type TEXT,
			file_url TEXT NOT NULL,
			tamano_bytes BIGINT DEFAULT 0,
			duracion_segundos REAL DEFAULT 0,
			estado TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP::text,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP::text,
			usuario_creador TEXT,
			observaciones TEXT
		)`,
		`ALTER TABLE empresa_buzon_mensajes ADD COLUMN IF NOT EXISTS destinatario_email TEXT`,
		`ALTER TABLE empresa_buzon_mensajes ADD COLUMN IF NOT EXISTS destinatario_nombre TEXT`,
		`ALTER TABLE empresa_buzon_mensajes ADD COLUMN IF NOT EXISTS referencia_id BIGINT DEFAULT 0`,
		`ALTER TABLE empresa_buzon_mensajes ADD COLUMN IF NOT EXISTS enlace_url TEXT`,
		`ALTER TABLE empresa_buzon_mensajes ADD COLUMN IF NOT EXISTS tarea_estado TEXT DEFAULT ''`,
		`ALTER TABLE empresa_buzon_mensajes ADD COLUMN IF NOT EXISTS tarea_vence_en TEXT`,
		`ALTER TABLE empresa_buzon_mensajes ADD COLUMN IF NOT EXISTS tarea_cerrada_en TEXT`,
		`ALTER TABLE empresa_buzon_mensajes ADD COLUMN IF NOT EXISTS tarea_cierre_descripcion TEXT`,
		`ALTER TABLE empresa_buzon_mensajes ADD COLUMN IF NOT EXISTS leido INTEGER DEFAULT 0`,
		`ALTER TABLE empresa_buzon_mensajes ADD COLUMN IF NOT EXISTS leido_en TEXT`,
		`ALTER TABLE empresa_chat_mensajes ADD COLUMN IF NOT EXISTS remitente_nombre TEXT`,
		`ALTER TABLE empresa_buzon_adjuntos ADD COLUMN IF NOT EXISTS duracion_segundos REAL DEFAULT 0`,
		`ALTER TABLE empresa_buzon_adjuntos ADD COLUMN IF NOT EXISTS observaciones TEXT`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_buzon_destinatario ON empresa_buzon_mensajes (empresa_id, destinatario_tipo, destinatario_ref, leido, id DESC)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_buzon_referencia ON empresa_buzon_mensajes (empresa_id, modulo, referencia_tipo, referencia_id)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_chat_empresa ON empresa_chat_mensajes (empresa_id, id DESC)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_buzon_adjuntos_mensaje ON empresa_buzon_adjuntos (empresa_id, mensaje_id, estado)`,
	}
	for _, stmt := range stmts {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	empresaBuzonSchemaReady = true
	return nil
}

func ResolveEmpresaBuzonActor(dbEmp, dbSuper *sql.DB, empresaID int64, email string) (EmpresaBuzonActor, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if empresaID <= 0 {
		return EmpresaBuzonActor{}, fmt.Errorf("empresa_id es obligatorio")
	}
	if email == "" || email == "sistema" {
		return EmpresaBuzonActor{}, fmt.Errorf("usuario autenticado requerido")
	}
	if dbEmp != nil {
		if user, err := GetEmpresaUsuarioByEmailScoped(dbEmp, email, empresaID); err == nil && user != nil && user.EmpresaID == empresaID {
			return actorFromEmpresaUsuario(*user), nil
		} else if err != nil && err != sql.ErrNoRows {
			return EmpresaBuzonActor{}, err
		}
	}
	actor := EmpresaBuzonActor{Tipo: "admin", Ref: email, Email: email, Nombre: email, Rol: "administrador"}
	if dbSuper != nil {
		if admin, err := GetAdminByEmailFull(dbSuper, email); err == nil && admin != nil {
			actor.Nombre = firstNonEmptyDB(strings.TrimSpace(admin.Name), email)
			actor.Rol = strings.TrimSpace(admin.Role)
		}
	}
	normalizeEmpresaBuzonActor(&actor)
	return actor, nil
}

func CreateEmpresaBuzonMensaje(dbConn *sql.DB, msg EmpresaBuzonMensaje) (EmpresaBuzonMensaje, error) {
	if err := EnsureEmpresaBuzonSchema(dbConn); err != nil {
		return EmpresaBuzonMensaje{}, err
	}
	normalizeEmpresaBuzonMensaje(&msg)
	if msg.EmpresaID <= 0 {
		return EmpresaBuzonMensaje{}, fmt.Errorf("empresa_id es obligatorio")
	}
	if msg.DestinatarioRef == "" {
		return EmpresaBuzonMensaje{}, fmt.Errorf("destinatario es obligatorio")
	}
	if msg.Titulo == "" {
		return EmpresaBuzonMensaje{}, fmt.Errorf("titulo es obligatorio")
	}
	if msg.Mensaje == "" {
		return EmpresaBuzonMensaje{}, fmt.Errorf("mensaje es obligatorio")
	}
	var id int64
	err := queryRowSQLCompat(dbConn, `INSERT INTO empresa_buzon_mensajes (
		empresa_id, destinatario_tipo, destinatario_ref, destinatario_email, destinatario_nombre,
		remitente_tipo, remitente_ref, remitente_email, remitente_nombre,
		titulo, mensaje, tipo, prioridad, modulo, referencia_tipo, referencia_id, enlace_url,
		tarea_estado, tarea_vence_en, tarea_cerrada_en, tarea_cierre_descripcion,
		leido, estado, fecha_creacion, fecha_actualizacion, usuario_creador
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 'activo', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?)
	RETURNING id`,
		msg.EmpresaID, msg.DestinatarioTipo, msg.DestinatarioRef, msg.DestinatarioEmail, msg.DestinatarioNombre,
		msg.RemitenteTipo, msg.RemitenteRef, msg.RemitenteEmail, msg.RemitenteNombre,
		msg.Titulo, msg.Mensaje, msg.Tipo, msg.Prioridad, msg.Modulo, msg.ReferenciaTipo, msg.ReferenciaID, msg.EnlaceURL,
		msg.TareaEstado, msg.TareaVenceEn, msg.TareaCerradaEn, msg.TareaCierreDescripcion,
		firstNonEmptyDB(msg.UsuarioCreador, msg.RemitenteEmail, msg.RemitenteRef, "sistema"),
	).Scan(&id)
	if err != nil {
		return EmpresaBuzonMensaje{}, err
	}
	return GetEmpresaBuzonMensajeByID(dbConn, msg.EmpresaID, id)
}

func GetEmpresaBuzonMensajeByID(dbConn *sql.DB, empresaID, id int64) (EmpresaBuzonMensaje, error) {
	if err := EnsureEmpresaBuzonSchema(dbConn); err != nil {
		return EmpresaBuzonMensaje{}, err
	}
	row := queryRowSQLCompat(dbConn, empresaBuzonMensajeSelectSQL()+` WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, id)
	return scanEmpresaBuzonMensaje(row)
}

func CountEmpresaBuzonUnread(dbConn *sql.DB, empresaID int64, actor EmpresaBuzonActor) (int64, error) {
	if err := EnsureEmpresaBuzonSchema(dbConn); err != nil {
		return 0, err
	}
	normalizeEmpresaBuzonActor(&actor)
	var total int64
	err := queryRowSQLCompat(dbConn, `SELECT COUNT(*)
		FROM empresa_buzon_mensajes
		WHERE empresa_id = ? AND destinatario_tipo = ? AND destinatario_ref = ?
		  AND COALESCE(leido, 0) = 0 AND COALESCE(estado, 'activo') = 'activo'`,
		empresaID, actor.Tipo, actor.Ref).Scan(&total)
	return total, err
}

func ListEmpresaBuzonMensajes(dbConn *sql.DB, empresaID int64, actor EmpresaBuzonActor, includeRead bool, limit int) ([]EmpresaBuzonMensaje, error) {
	if err := EnsureEmpresaBuzonSchema(dbConn); err != nil {
		return nil, err
	}
	normalizeEmpresaBuzonActor(&actor)
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	// #nosec G202 -- SQL structure is assembled only from server-side allowlists; all external values remain bound parameters.
	query := empresaBuzonMensajeSelectSQL() + ` WHERE empresa_id = ? AND destinatario_tipo = ? AND destinatario_ref = ? AND COALESCE(estado, 'activo') = 'activo'`
	args := []interface{}{empresaID, actor.Tipo, actor.Ref}
	if !includeRead {
		query += ` AND COALESCE(leido, 0) = 0`
	}
	query += ` ORDER BY id DESC LIMIT ?`
	args = append(args, limit)
	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EmpresaBuzonMensaje, 0)
	for rows.Next() {
		msg, err := scanEmpresaBuzonMensajeRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := attachEmpresaBuzonAdjuntos(dbConn, empresaID, out); err != nil {
		return nil, err
	}
	return out, nil
}

func MarkEmpresaBuzonMensajeRead(dbConn *sql.DB, empresaID, id int64, actor EmpresaBuzonActor) error {
	if err := EnsureEmpresaBuzonSchema(dbConn); err != nil {
		return err
	}
	normalizeEmpresaBuzonActor(&actor)
	res, err := execSQLCompat(dbConn, `UPDATE empresa_buzon_mensajes
		SET leido = 1, leido_en = CURRENT_TIMESTAMP::text, fecha_actualizacion = CURRENT_TIMESTAMP::text
		WHERE empresa_id = ? AND id = ? AND destinatario_tipo = ? AND destinatario_ref = ?`,
		empresaID, id, actor.Tipo, actor.Ref)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func CompleteEmpresaBuzonTarea(dbConn *sql.DB, empresaID, id int64, actor EmpresaBuzonActor, descripcion string) (EmpresaBuzonMensaje, error) {
	if err := EnsureEmpresaBuzonSchema(dbConn); err != nil {
		return EmpresaBuzonMensaje{}, err
	}
	normalizeEmpresaBuzonActor(&actor)
	descripcion = truncateTextDB(strings.TrimSpace(descripcion), 3000)
	if descripcion == "" {
		return EmpresaBuzonMensaje{}, fmt.Errorf("descripcion de cierre es obligatoria")
	}
	res, err := execSQLCompat(dbConn, `UPDATE empresa_buzon_mensajes
		SET tarea_estado = 'finalizada', tarea_cerrada_en = CURRENT_TIMESTAMP::text, tarea_cierre_descripcion = ?,
		    leido = 1, leido_en = COALESCE(NULLIF(leido_en, ''), CURRENT_TIMESTAMP::text), fecha_actualizacion = CURRENT_TIMESTAMP::text
		WHERE empresa_id = ? AND id = ? AND destinatario_tipo = ? AND destinatario_ref = ?
		  AND COALESCE(tipo, '') = 'tarea' AND COALESCE(estado, 'activo') = 'activo'`,
		descripcion, empresaID, id, actor.Tipo, actor.Ref)
	if err != nil {
		return EmpresaBuzonMensaje{}, err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return EmpresaBuzonMensaje{}, sql.ErrNoRows
	}
	return GetEmpresaBuzonMensajeByID(dbConn, empresaID, id)
}

func CreateEmpresaChatMensaje(dbConn *sql.DB, empresaID int64, actor EmpresaBuzonActor, mensaje string) (EmpresaChatMensaje, error) {
	if err := EnsureEmpresaBuzonSchema(dbConn); err != nil {
		return EmpresaChatMensaje{}, err
	}
	normalizeEmpresaBuzonActor(&actor)
	mensaje = truncateTextDB(strings.TrimSpace(mensaje), 2000)
	if empresaID <= 0 {
		return EmpresaChatMensaje{}, fmt.Errorf("empresa_id es obligatorio")
	}
	if mensaje == "" {
		return EmpresaChatMensaje{}, fmt.Errorf("mensaje es obligatorio")
	}
	var id int64
	err := queryRowSQLCompat(dbConn, `INSERT INTO empresa_chat_mensajes (
		empresa_id, remitente_tipo, remitente_ref, remitente_email, remitente_nombre,
		mensaje, estado, fecha_creacion, fecha_actualizacion, usuario_creador
	) VALUES (?, ?, ?, ?, ?, ?, 'activo', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?)
	RETURNING id`, empresaID, actor.Tipo, actor.Ref, actor.Email, actor.Nombre, mensaje, firstNonEmptyDB(actor.Email, actor.Ref, "sistema")).Scan(&id)
	if err != nil {
		return EmpresaChatMensaje{}, err
	}
	items, err := ListEmpresaChatMensajes(dbConn, empresaID, 1)
	if err != nil {
		return EmpresaChatMensaje{}, err
	}
	for _, item := range items {
		if item.ID == id {
			return item, nil
		}
	}
	return EmpresaChatMensaje{ID: id, EmpresaID: empresaID, RemitenteTipo: actor.Tipo, RemitenteRef: actor.Ref, RemitenteEmail: actor.Email, RemitenteNombre: actor.Nombre, Mensaje: mensaje, Estado: "activo"}, nil
}

func CreateEmpresaBuzonAdjunto(dbConn *sql.DB, payload EmpresaBuzonAdjunto) (EmpresaBuzonAdjunto, error) {
	if err := EnsureEmpresaBuzonSchema(dbConn); err != nil {
		return EmpresaBuzonAdjunto{}, err
	}
	payload.NombreArchivo = truncateTextDB(strings.TrimSpace(payload.NombreArchivo), 240)
	payload.MimeType = truncateTextDB(strings.TrimSpace(payload.MimeType), 160)
	payload.TipoArchivo = normalizeEmpresaBuzonTipoAdjunto(payload.TipoArchivo, payload.MimeType, payload.FileURL)
	payload.FileURL = normalizeEmpresaBuzonFileURL(payload.FileURL)
	payload.UsuarioCreador = truncateTextDB(strings.TrimSpace(payload.UsuarioCreador), 180)
	payload.Observaciones = truncateTextDB(strings.TrimSpace(payload.Observaciones), 500)
	if payload.EmpresaID <= 0 || payload.MensajeID <= 0 {
		return EmpresaBuzonAdjunto{}, fmt.Errorf("empresa_id y mensaje_id son obligatorios")
	}
	if payload.FileURL == "" {
		return EmpresaBuzonAdjunto{}, fmt.Errorf("file_url es obligatorio")
	}
	var exists int
	if err := queryRowSQLCompat(dbConn, `SELECT 1 FROM empresa_buzon_mensajes WHERE empresa_id = ? AND id = ? LIMIT 1`, payload.EmpresaID, payload.MensajeID).Scan(&exists); err != nil {
		if err == sql.ErrNoRows {
			return EmpresaBuzonAdjunto{}, fmt.Errorf("mensaje no encontrado para esta empresa")
		}
		return EmpresaBuzonAdjunto{}, err
	}
	var id int64
	err := queryRowSQLCompat(dbConn, `INSERT INTO empresa_buzon_adjuntos (
		empresa_id, mensaje_id, tipo_archivo, nombre_archivo, mime_type, file_url,
		tamano_bytes, duracion_segundos, estado, fecha_creacion, fecha_actualizacion, usuario_creador, observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'activo', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?)
	RETURNING id`, payload.EmpresaID, payload.MensajeID, payload.TipoArchivo, payload.NombreArchivo, payload.MimeType, payload.FileURL, payload.TamanoBytes, payload.DuracionSegundos, payload.UsuarioCreador, payload.Observaciones).Scan(&id)
	if err != nil {
		return EmpresaBuzonAdjunto{}, err
	}
	items, err := ListEmpresaBuzonAdjuntosByMensajeIDs(dbConn, payload.EmpresaID, []int64{payload.MensajeID})
	if err != nil {
		return EmpresaBuzonAdjunto{}, err
	}
	for _, item := range items[payload.MensajeID] {
		if item.ID == id {
			return item, nil
		}
	}
	payload.ID = id
	payload.Estado = "activo"
	return payload, nil
}

func ListEmpresaBuzonAdjuntosByMensajeIDs(dbConn *sql.DB, empresaID int64, messageIDs []int64) (map[int64][]EmpresaBuzonAdjunto, error) {
	if err := EnsureEmpresaBuzonSchema(dbConn); err != nil {
		return nil, err
	}
	cleanIDs := make([]int64, 0, len(messageIDs))
	seen := map[int64]bool{}
	for _, id := range messageIDs {
		if id > 0 && !seen[id] {
			seen[id] = true
			cleanIDs = append(cleanIDs, id)
		}
	}
	out := make(map[int64][]EmpresaBuzonAdjunto)
	if len(cleanIDs) == 0 {
		return out, nil
	}
	placeholders := make([]string, 0, len(cleanIDs))
	args := []interface{}{empresaID}
	for _, id := range cleanIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	// #nosec G202 -- SQL structure is assembled only from server-side allowlists; all external values remain bound parameters.
	query := `SELECT id, empresa_id, mensaje_id, COALESCE(tipo_archivo, ''), COALESCE(nombre_archivo, ''),
		COALESCE(mime_type, ''), COALESCE(file_url, ''), COALESCE(tamano_bytes, 0), COALESCE(duracion_segundos, 0),
		COALESCE(estado, 'activo'), COALESCE(fecha_creacion, ''), COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''), COALESCE(observaciones, '')
		FROM empresa_buzon_adjuntos
		WHERE empresa_id = ? AND mensaje_id IN (` + strings.Join(placeholders, ",") + `)
		  AND COALESCE(estado, 'activo') = 'activo'
		ORDER BY id ASC`
	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item EmpresaBuzonAdjunto
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.MensajeID, &item.TipoArchivo, &item.NombreArchivo, &item.MimeType, &item.FileURL, &item.TamanoBytes, &item.DuracionSegundos, &item.Estado, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador, &item.Observaciones); err != nil {
			return nil, err
		}
		out[item.MensajeID] = append(out[item.MensajeID], item)
	}
	return out, rows.Err()
}

func ListEmpresaChatMensajes(dbConn *sql.DB, empresaID int64, limit int) ([]EmpresaChatMensaje, error) {
	if err := EnsureEmpresaBuzonSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 40
	}
	rows, err := dbConn.Query(`SELECT id, empresa_id, COALESCE(remitente_tipo, ''), COALESCE(remitente_ref, ''),
		COALESCE(remitente_email, ''), COALESCE(remitente_nombre, ''), COALESCE(mensaje, ''),
		COALESCE(estado, 'activo'), COALESCE(fecha_creacion, ''), COALESCE(fecha_actualizacion, ''), COALESCE(usuario_creador, '')
		FROM empresa_chat_mensajes
		WHERE empresa_id = ? AND COALESCE(estado, 'activo') = 'activo'
		ORDER BY id DESC LIMIT ?`, empresaID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EmpresaChatMensaje, 0)
	for rows.Next() {
		var item EmpresaChatMensaje
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.RemitenteTipo, &item.RemitenteRef, &item.RemitenteEmail, &item.RemitenteNombre, &item.Mensaje, &item.Estado, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, rows.Err()
}

func CreateEmpresaBodegaTransferNotification(dbConn *sql.DB, empresaID, productoID, bodegaOrigenID, bodegaDestinoID int64, cantidad float64, referencia, actorEmail string) (EmpresaBodegaTransferNotificationResult, error) {
	if err := EnsureEmpresaBuzonSchema(dbConn); err != nil {
		return EmpresaBodegaTransferNotificationResult{}, err
	}
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		return EmpresaBodegaTransferNotificationResult{}, err
	}
	if empresaID <= 0 || productoID <= 0 || bodegaOrigenID <= 0 || bodegaDestinoID <= 0 {
		return EmpresaBodegaTransferNotificationResult{}, fmt.Errorf("datos de traslado incompletos")
	}
	producto, err := GetProductoByID(dbConn, empresaID, productoID)
	if err != nil {
		return EmpresaBodegaTransferNotificationResult{}, err
	}
	origen, err := getEmpresaBodegaByID(dbConn, empresaID, bodegaOrigenID)
	if err != nil {
		return EmpresaBodegaTransferNotificationResult{}, err
	}
	destino, err := getEmpresaBodegaByID(dbConn, empresaID, bodegaDestinoID)
	if err != nil {
		return EmpresaBodegaTransferNotificationResult{}, err
	}
	remitente := EmpresaBuzonActor{Tipo: "sistema", Ref: "inventario", Nombre: "Inventario PCS", Email: strings.ToLower(strings.TrimSpace(actorEmail))}
	normalizeEmpresaBuzonActor(&remitente)
	recipients, err := resolveBodegaTransferRecipients(dbConn, empresaID, destino.Responsable, actorEmail)
	if err != nil {
		return EmpresaBodegaTransferNotificationResult{}, err
	}
	if len(recipients) == 0 {
		return EmpresaBodegaTransferNotificationResult{}, nil
	}
	titulo := "Traslado recibido en " + firstNonEmptyDB(destino.Nombre, "bodega destino")
	mensaje := fmt.Sprintf("Se trasladaron %.4g unidades de %s desde %s hacia %s.", cantidad, firstNonEmptyDB(producto.Nombre, fmt.Sprintf("producto %d", productoID)), firstNonEmptyDB(origen.Nombre, "bodega origen"), firstNonEmptyDB(destino.Nombre, "bodega destino"))
	if strings.TrimSpace(referencia) != "" {
		mensaje += " Referencia: " + truncateTextDB(strings.TrimSpace(referencia), 120) + "."
	}
	enlace := fmt.Sprintf("/administrar_empresa/administrar_productos.html?view=bodegas&empresa_id=%d", empresaID)
	result := EmpresaBodegaTransferNotificationResult{Recipients: make([]string, 0, len(recipients))}
	for _, recipient := range recipients {
		msg, err := CreateEmpresaBuzonMensaje(dbConn, EmpresaBuzonMensaje{
			EmpresaID:          empresaID,
			DestinatarioTipo:   recipient.Tipo,
			DestinatarioRef:    recipient.Ref,
			DestinatarioEmail:  recipient.Email,
			DestinatarioNombre: recipient.Nombre,
			RemitenteTipo:      remitente.Tipo,
			RemitenteRef:       remitente.Ref,
			RemitenteEmail:     remitente.Email,
			RemitenteNombre:    remitente.Nombre,
			Titulo:             titulo,
			Mensaje:            mensaje,
			Tipo:               "inventario_traslado",
			Prioridad:          "normal",
			Modulo:             "inventario",
			ReferenciaTipo:     "bodega_destino",
			ReferenciaID:       bodegaDestinoID,
			EnlaceURL:          enlace,
			UsuarioCreador:     actorEmail,
		})
		if err != nil {
			return result, err
		}
		if msg.ID > 0 {
			result.Created++
			result.Recipients = append(result.Recipients, firstNonEmptyDB(recipient.Email, recipient.Nombre, recipient.Ref))
		}
	}
	return result, nil
}

func empresaBuzonMensajeSelectSQL() string {
	return `SELECT id, empresa_id, COALESCE(destinatario_tipo, ''), COALESCE(destinatario_ref, ''),
		COALESCE(destinatario_email, ''), COALESCE(destinatario_nombre, ''),
		COALESCE(remitente_tipo, ''), COALESCE(remitente_ref, ''), COALESCE(remitente_email, ''), COALESCE(remitente_nombre, ''),
		COALESCE(titulo, ''), COALESCE(mensaje, ''), COALESCE(tipo, ''), COALESCE(prioridad, ''),
		COALESCE(modulo, ''), COALESCE(referencia_tipo, ''), COALESCE(referencia_id, 0), COALESCE(enlace_url, ''),
		COALESCE(tarea_estado, ''), COALESCE(tarea_vence_en, ''), COALESCE(tarea_cerrada_en, ''), COALESCE(tarea_cierre_descripcion, ''),
		COALESCE(leido, 0), COALESCE(leido_en, ''), COALESCE(estado, 'activo'), COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''), COALESCE(usuario_creador, '')
		FROM empresa_buzon_mensajes`
}

type empresaBuzonScanner interface {
	Scan(dest ...interface{}) error
}

func scanEmpresaBuzonMensaje(row empresaBuzonScanner) (EmpresaBuzonMensaje, error) {
	var msg EmpresaBuzonMensaje
	err := row.Scan(&msg.ID, &msg.EmpresaID, &msg.DestinatarioTipo, &msg.DestinatarioRef, &msg.DestinatarioEmail, &msg.DestinatarioNombre, &msg.RemitenteTipo, &msg.RemitenteRef, &msg.RemitenteEmail, &msg.RemitenteNombre, &msg.Titulo, &msg.Mensaje, &msg.Tipo, &msg.Prioridad, &msg.Modulo, &msg.ReferenciaTipo, &msg.ReferenciaID, &msg.EnlaceURL, &msg.TareaEstado, &msg.TareaVenceEn, &msg.TareaCerradaEn, &msg.TareaCierreDescripcion, &msg.Leido, &msg.LeidoEn, &msg.Estado, &msg.FechaCreacion, &msg.FechaActualizacion, &msg.UsuarioCreador)
	return msg, err
}

func scanEmpresaBuzonMensajeRows(rows *sql.Rows) (EmpresaBuzonMensaje, error) {
	return scanEmpresaBuzonMensaje(rows)
}

func normalizeEmpresaBuzonMensaje(msg *EmpresaBuzonMensaje) {
	normalizeEmpresaBuzonActorFields(&msg.DestinatarioTipo, &msg.DestinatarioRef, &msg.DestinatarioEmail, &msg.DestinatarioNombre)
	normalizeEmpresaBuzonActorFields(&msg.RemitenteTipo, &msg.RemitenteRef, &msg.RemitenteEmail, &msg.RemitenteNombre)
	msg.Titulo = truncateTextDB(strings.TrimSpace(msg.Titulo), 180)
	msg.Mensaje = truncateTextDB(strings.TrimSpace(msg.Mensaje), 3000)
	msg.Tipo = firstNonEmptyDB(truncateTextDB(strings.TrimSpace(msg.Tipo), 80), "general")
	msg.Prioridad = normalizeEmpresaBuzonPrioridad(msg.Prioridad)
	msg.Modulo = truncateTextDB(strings.TrimSpace(msg.Modulo), 80)
	msg.ReferenciaTipo = truncateTextDB(strings.TrimSpace(msg.ReferenciaTipo), 80)
	msg.EnlaceURL = normalizeEmpresaBuzonLink(msg.EnlaceURL)
	msg.TareaEstado = normalizeEmpresaBuzonTareaEstado(msg.TareaEstado, msg.Tipo)
	msg.TareaVenceEn = truncateTextDB(strings.TrimSpace(msg.TareaVenceEn), 80)
	msg.TareaCerradaEn = truncateTextDB(strings.TrimSpace(msg.TareaCerradaEn), 80)
	msg.TareaCierreDescripcion = truncateTextDB(strings.TrimSpace(msg.TareaCierreDescripcion), 3000)
	msg.UsuarioCreador = truncateTextDB(strings.TrimSpace(msg.UsuarioCreador), 180)
}

func normalizeEmpresaBuzonActor(actor *EmpresaBuzonActor) {
	if actor == nil {
		return
	}
	normalizeEmpresaBuzonActorFields(&actor.Tipo, &actor.Ref, &actor.Email, &actor.Nombre)
	actor.Rol = truncateTextDB(strings.TrimSpace(actor.Rol), 80)
}

func normalizeEmpresaBuzonActorFields(tipo, ref, email, nombre *string) {
	t := strings.ToLower(strings.TrimSpace(*tipo))
	if t != "admin" && t != "usuario" && t != "sistema" {
		t = "usuario"
	}
	e := strings.ToLower(strings.TrimSpace(*email))
	r := strings.TrimSpace(*ref)
	if t == "admin" && r == "" {
		r = e
	}
	if t == "sistema" && r == "" {
		r = "sistema"
	}
	if t == "usuario" && r == "" {
		r = e
	}
	*tipo = truncateTextDB(t, 40)
	*ref = truncateTextDB(r, 180)
	*email = truncateTextDB(e, 180)
	*nombre = truncateTextDB(strings.TrimSpace(*nombre), 160)
}

func normalizeEmpresaBuzonPrioridad(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "alta", "urgente":
		return "alta"
	case "baja":
		return "baja"
	default:
		return "normal"
	}
}

func normalizeEmpresaBuzonTareaEstado(value, tipo string) string {
	if strings.ToLower(strings.TrimSpace(tipo)) != "tarea" {
		return ""
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "finalizada", "cerrada":
		return "finalizada"
	case "cancelada":
		return "cancelada"
	default:
		return "pendiente"
	}
}

func normalizeEmpresaBuzonLink(value string) string {
	value = truncateTextDB(strings.TrimSpace(value), 500)
	if value == "" || strings.HasPrefix(value, "/administrar_empresa/") {
		return value
	}
	return ""
}

func normalizeEmpresaBuzonFileURL(value string) string {
	value = truncateTextDB(strings.TrimSpace(value), 500)
	if value == "" || strings.HasPrefix(value, "/uploads/empresas/") {
		return value
	}
	return ""
}

func normalizeEmpresaBuzonTipoAdjunto(value, mimeType, fileURL string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "imagen" || value == "audio" || value == "archivo" || value == "foto" || value == "documento" {
		if value == "foto" {
			return "imagen"
		}
		if value == "documento" {
			return "archivo"
		}
		return value
	}
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	if strings.HasPrefix(mimeType, "image/") {
		return "imagen"
	}
	if strings.HasPrefix(mimeType, "audio/") {
		return "audio"
	}
	ext := strings.ToLower(strings.TrimSpace(fileURL))
	if strings.HasSuffix(ext, ".png") || strings.HasSuffix(ext, ".jpg") || strings.HasSuffix(ext, ".jpeg") || strings.HasSuffix(ext, ".webp") || strings.HasSuffix(ext, ".gif") {
		return "imagen"
	}
	if strings.HasSuffix(ext, ".mp3") || strings.HasSuffix(ext, ".wav") || strings.HasSuffix(ext, ".ogg") || strings.HasSuffix(ext, ".webm") || strings.HasSuffix(ext, ".m4a") {
		return "audio"
	}
	return "archivo"
}

func attachEmpresaBuzonAdjuntos(dbConn *sql.DB, empresaID int64, messages []EmpresaBuzonMensaje) error {
	ids := make([]int64, 0, len(messages))
	for _, msg := range messages {
		ids = append(ids, msg.ID)
	}
	adjuntos, err := ListEmpresaBuzonAdjuntosByMensajeIDs(dbConn, empresaID, ids)
	if err != nil {
		return err
	}
	for i := range messages {
		messages[i].Adjuntos = adjuntos[messages[i].ID]
	}
	return nil
}

func actorFromEmpresaUsuario(user EmpresaUsuario) EmpresaBuzonActor {
	actor := EmpresaBuzonActor{
		Tipo:      "usuario",
		Ref:       fmt.Sprintf("%d", user.ID),
		Email:     strings.ToLower(strings.TrimSpace(user.Email)),
		Nombre:    strings.TrimSpace(user.Nombre),
		Rol:       strings.TrimSpace(user.RolNombre),
		UsuarioID: user.ID,
	}
	normalizeEmpresaBuzonActor(&actor)
	return actor
}

func getEmpresaBodegaByID(dbConn *sql.DB, empresaID, bodegaID int64) (Bodega, error) {
	row := queryRowSQLCompat(dbConn, `SELECT id, empresa_id, COALESCE(codigo, ''), COALESCE(nombre, ''), COALESCE(ubicacion, ''),
		COALESCE(responsable, ''), COALESCE(fecha_creacion, ''), COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''), COALESCE(estado, 'activo'), COALESCE(observaciones, '')
		FROM bodegas WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, bodegaID)
	var b Bodega
	err := row.Scan(&b.ID, &b.EmpresaID, &b.Codigo, &b.Nombre, &b.Ubicacion, &b.Responsable, &b.FechaCreacion, &b.FechaActualizacion, &b.UsuarioCreador, &b.Estado, &b.Observaciones)
	return b, err
}

func resolveBodegaTransferRecipients(dbConn *sql.DB, empresaID int64, responsable, actorEmail string) ([]EmpresaBuzonActor, error) {
	recipients := make([]EmpresaBuzonActor, 0)
	seen := map[string]bool{}
	add := func(actor EmpresaBuzonActor) {
		normalizeEmpresaBuzonActor(&actor)
		if actor.Ref == "" {
			return
		}
		key := actor.Tipo + "|" + actor.Ref
		if seen[key] {
			return
		}
		seen[key] = true
		recipients = append(recipients, actor)
	}

	responsable = strings.TrimSpace(responsable)
	if responsable != "" {
		if user, err := ResolveEmpresaUsuarioByReference(dbConn, empresaID, responsable); err == nil && user != nil && user.EmpresaID == empresaID {
			add(actorFromEmpresaUsuario(*user))
		} else if strings.Contains(responsable, "@") {
			add(EmpresaBuzonActor{Tipo: "admin", Ref: strings.ToLower(responsable), Email: strings.ToLower(responsable), Nombre: responsable})
		}
	}

	users, err := GetEmpresaUsuarios(dbConn, empresaID, false)
	if err != nil {
		return recipients, err
	}
	for _, user := range users {
		role := normalizeEmpresaBuzonRole(user.RolNombre)
		if role == "jefe_bodega" || role == "responsable_bodega" || role == "inventario" || role == "supervisor" || role == "administrador" || role == "admin" {
			add(actorFromEmpresaUsuario(user))
		}
	}
	if len(recipients) == 0 && strings.TrimSpace(actorEmail) != "" {
		if user, err := GetEmpresaUsuarioByEmailScoped(dbConn, strings.TrimSpace(actorEmail), empresaID); err == nil && user != nil {
			add(actorFromEmpresaUsuario(*user))
		} else {
			add(EmpresaBuzonActor{Tipo: "admin", Ref: strings.ToLower(strings.TrimSpace(actorEmail)), Email: strings.ToLower(strings.TrimSpace(actorEmail)), Nombre: strings.TrimSpace(actorEmail)})
		}
	}
	return recipients, nil
}

func normalizeEmpresaBuzonRole(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ñ", "n", " ", "_", "-", "_")
	value = replacer.Replace(value)
	for strings.Contains(value, "__") {
		value = strings.ReplaceAll(value, "__", "_")
	}
	switch value {
	case "responsable_de_bodega", "responsable_bodega", "bodeguero", "almacenista":
		return "responsable_bodega"
	}
	return value
}
