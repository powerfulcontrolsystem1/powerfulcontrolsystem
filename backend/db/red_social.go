package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type PublicacionRedSocial struct {
	ID            int       `json:"id"`
	EmpresaID     int       `json:"empresa_id"`
	EmpresaNombre string    `json:"empresa_nombre,omitempty"`
	EmpresaSlug   string    `json:"empresa_slug,omitempty"`
	Nombre        string    `json:"nombre"`
	Descripcion   string    `json:"descripcion"`
	FotoURL       string    `json:"foto_url"`
	YoutubeURL    string    `json:"youtube_url"`
	FechaCreacion time.Time `json:"fecha_creacion"`
	Estado        string    `json:"estado"`

	// Campos derivados para feed público (Facebook-like).
	ReaccionesResumen map[string]int `json:"reacciones_resumen,omitempty"`
	ComentariosTotal  int            `json:"comentarios_total,omitempty"`
	UserReaction      string         `json:"user_reaction,omitempty"`
}

func EnsureEmpresaPublicacionesRedSocialSchema(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS empresa_publicaciones_red_social (
		id BIGSERIAL PRIMARY KEY,
		empresa_id INTEGER NOT NULL,
		nombre TEXT NOT NULL,
		descripcion TEXT NOT NULL,
		foto_url TEXT,
		youtube_url TEXT,
		fecha_creacion DATETIME DEFAULT CURRENT_TIMESTAMP,
		estado TEXT DEFAULT 'activo'
	);`
	if shouldUsePostgresCompat(db) {
		query = `
		CREATE TABLE IF NOT EXISTS empresa_publicaciones_red_social (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			nombre TEXT NOT NULL,
			descripcion TEXT NOT NULL,
			foto_url TEXT,
			youtube_url TEXT,
			fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			estado TEXT DEFAULT 'activo'
		);`
	}
	_, err := execSQLCompat(db, query)
	if err != nil {
		return fmt.Errorf("error creando empresa_publicaciones_red_social: %v", err)
	}
	// Migración suave de columnas nuevas.
	_ = ensureColumnIfMissing(db, "empresa_publicaciones_red_social", "foto_url", "TEXT")
	_ = ensureColumnIfMissing(db, "empresa_publicaciones_red_social", "youtube_url", "TEXT")
	return nil
}

type PublicacionRedSocialComentario struct {
	ID            int       `json:"id"`
	PublicacionID int       `json:"publicacion_id"`
	EmpresaID     int       `json:"empresa_id"`
	ActorKey      string    `json:"actor_key,omitempty"`
	Nombre        string    `json:"nombre,omitempty"`
	Contenido     string    `json:"contenido"`
	FechaCreacion time.Time `json:"fecha_creacion"`
	Estado        string    `json:"estado"`
}

type PublicacionRedSocialReaccion struct {
	PublicacionID int       `json:"publicacion_id"`
	EmpresaID     int       `json:"empresa_id"`
	ActorKey      string    `json:"actor_key"`
	Reaccion      string    `json:"reaccion"`
	Fecha         time.Time `json:"fecha"`
}

type RedSocialEmpresaSeguida struct {
	EmpresaID              int    `json:"empresa_id"`
	EmpresaNombre          string `json:"empresa_nombre,omitempty"`
	Seguido                bool   `json:"seguido"`
	UltimaPublicacionID    int    `json:"ultima_publicacion_id,omitempty"`
	UltimaPublicacion      string `json:"ultima_publicacion,omitempty"`
	UltimaPublicacionFecha string `json:"ultima_publicacion_fecha,omitempty"`
	Nuevas                 int    `json:"nuevas"`
}

func EnsureEmpresaRedSocialInteraccionesSchema(db *sql.DB) error {
	if err := EnsureEmpresaPublicacionesRedSocialSchema(db); err != nil {
		return err
	}

	comments := `
	CREATE TABLE IF NOT EXISTS empresa_publicaciones_red_social_comentarios (
		id BIGSERIAL PRIMARY KEY,
		publicacion_id INTEGER NOT NULL,
		empresa_id INTEGER NOT NULL,
		actor_key TEXT,
		nombre TEXT,
		contenido TEXT NOT NULL,
		fecha_creacion DATETIME DEFAULT CURRENT_TIMESTAMP,
		estado TEXT DEFAULT 'activo'
	);`
	reactions := `
	CREATE TABLE IF NOT EXISTS empresa_publicaciones_red_social_reacciones (
		id BIGSERIAL PRIMARY KEY,
		publicacion_id INTEGER NOT NULL,
		empresa_id INTEGER NOT NULL,
		actor_key TEXT NOT NULL,
		reaccion TEXT NOT NULL,
		fecha DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	follows := `
	CREATE TABLE IF NOT EXISTS empresa_publicaciones_red_social_seguidores (
		id BIGSERIAL PRIMARY KEY,
		empresa_id INTEGER NOT NULL,
		actor_key TEXT NOT NULL,
		fecha_creacion DATETIME DEFAULT CURRENT_TIMESTAMP,
		ultima_vista_publicacion_id INTEGER DEFAULT 0,
		estado TEXT DEFAULT 'activo'
	);`
	uniq := `CREATE UNIQUE INDEX IF NOT EXISTS ux_red_social_reacciones_unique ON empresa_publicaciones_red_social_reacciones(publicacion_id, actor_key);`
	uniqFollow := `CREATE UNIQUE INDEX IF NOT EXISTS ux_red_social_seguidores_unique ON empresa_publicaciones_red_social_seguidores(empresa_id, actor_key);`
	ixC := `CREATE INDEX IF NOT EXISTS ix_red_social_comentarios_post_fecha ON empresa_publicaciones_red_social_comentarios(publicacion_id, fecha_creacion DESC);`
	ixR := `CREATE INDEX IF NOT EXISTS ix_red_social_reacciones_post_fecha ON empresa_publicaciones_red_social_reacciones(publicacion_id, fecha DESC);`
	ixF := `CREATE INDEX IF NOT EXISTS ix_red_social_seguidores_actor ON empresa_publicaciones_red_social_seguidores(actor_key, estado);`

	if shouldUsePostgresCompat(db) {
		comments = `
		CREATE TABLE IF NOT EXISTS empresa_publicaciones_red_social_comentarios (
			id BIGSERIAL PRIMARY KEY,
			publicacion_id BIGINT NOT NULL,
			empresa_id BIGINT NOT NULL,
			actor_key TEXT,
			nombre TEXT,
			contenido TEXT NOT NULL,
			fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			estado TEXT DEFAULT 'activo'
		);`
		reactions = `
		CREATE TABLE IF NOT EXISTS empresa_publicaciones_red_social_reacciones (
			id BIGSERIAL PRIMARY KEY,
			publicacion_id BIGINT NOT NULL,
			empresa_id BIGINT NOT NULL,
			actor_key TEXT NOT NULL,
			reaccion TEXT NOT NULL,
			fecha TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`
		follows = `
		CREATE TABLE IF NOT EXISTS empresa_publicaciones_red_social_seguidores (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			actor_key TEXT NOT NULL,
			fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			ultima_vista_publicacion_id BIGINT DEFAULT 0,
			estado TEXT DEFAULT 'activo'
		);`
	}

	stmts := []string{
		comments,
		reactions,
		follows,
		uniq,
		uniqFollow,
		ixC,
		ixR,
		ixF,
	}
	for _, stmt := range stmts {
		if _, err := execSQLCompat(db, stmt); err != nil {
			return err
		}
	}

	_ = ensureColumnIfMissing(db, "empresa_publicaciones_red_social_comentarios", "actor_key", "TEXT")
	_ = ensureColumnIfMissing(db, "empresa_publicaciones_red_social_comentarios", "nombre", "TEXT")
	_ = ensureColumnIfMissing(db, "empresa_publicaciones_red_social_reacciones", "actor_key", "TEXT")
	_ = ensureColumnIfMissing(db, "empresa_publicaciones_red_social_reacciones", "reaccion", "TEXT")
	_ = ensureColumnIfMissing(db, "empresa_publicaciones_red_social_seguidores", "actor_key", "TEXT")
	_ = ensureColumnIfMissing(db, "empresa_publicaciones_red_social_seguidores", "ultima_vista_publicacion_id", "INTEGER DEFAULT 0")
	_ = ensureColumnIfMissing(db, "empresa_publicaciones_red_social_seguidores", "estado", "TEXT DEFAULT 'activo'")
	return nil
}

func normalizeReaccion(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "me_gusta", "me_encanta", "me_importa", "me_divierte", "me_asombra", "me_entristece", "me_enoja":
		return v
	default:
		return ""
	}
}

func normalizeActorKey(raw string) string {
	v := strings.TrimSpace(raw)
	if len(v) < 8 {
		return ""
	}
	if len(v) > 96 {
		v = v[:96]
	}
	return v
}

func clampInt(v, def, min, max int) int {
	if v <= 0 {
		v = def
	}
	if v < min {
		v = min
	}
	if v > max {
		v = max
	}
	return v
}

func GetPublicacionesRedSocialActivas(db *sql.DB, limit, offset int) ([]PublicacionRedSocial, error) {
	if err := EnsureEmpresaRedSocialInteraccionesSchema(db); err != nil {
		return nil, err
	}
	limit = clampInt(limit, 20, 1, 50)
	if offset < 0 {
		offset = 0
	}
	query := `SELECT p.id, p.empresa_id, COALESCE(e.nombre, ''), COALESCE(vpc.empresa_slug, ''), p.nombre, p.descripcion, COALESCE(p.foto_url,''), COALESCE(p.youtube_url,''), p.fecha_creacion, p.estado
	          FROM empresa_publicaciones_red_social p
	          LEFT JOIN empresas e ON e.id = p.empresa_id OR COALESCE(e.empresa_id, 0) = p.empresa_id
	          LEFT JOIN empresa_venta_publica_configuracion vpc ON vpc.empresa_id = p.empresa_id AND COALESCE(vpc.estado, 'activo') <> 'inactivo'
	          WHERE p.estado = 'activo' ORDER BY p.fecha_creacion DESC LIMIT ? OFFSET ?`
	rows, err := querySQLCompat(db, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pubs []PublicacionRedSocial
	for rows.Next() {
		var p PublicacionRedSocial
		var youtube string
		if err := rows.Scan(&p.ID, &p.EmpresaID, &p.EmpresaNombre, &p.EmpresaSlug, &p.Nombre, &p.Descripcion, &p.FotoURL, &youtube, &p.FechaCreacion, &p.Estado); err != nil {
			return nil, err
		}
		p.YoutubeURL = youtube
		p.ReaccionesResumen, _ = GetPublicacionReaccionesResumen(db, p.ID)
		p.ComentariosTotal, _ = GetPublicacionComentariosTotal(db, p.ID)
		pubs = append(pubs, p)
	}
	if pubs == nil {
		pubs = []PublicacionRedSocial{}
	}
	return pubs, nil
}

func GetPublicacionesRedSocialByEmpresa(db *sql.DB, empresaID int, limit, offset int) ([]PublicacionRedSocial, error) {
	if err := EnsureEmpresaRedSocialInteraccionesSchema(db); err != nil {
		return nil, err
	}
	limit = clampInt(limit, 50, 1, 200)
	if offset < 0 {
		offset = 0
	}
	query := `SELECT p.id, p.empresa_id, COALESCE(e.nombre, ''), COALESCE(vpc.empresa_slug, ''), p.nombre, p.descripcion, COALESCE(p.foto_url,''), COALESCE(p.youtube_url,''), p.fecha_creacion, p.estado
	          FROM empresa_publicaciones_red_social p
	          LEFT JOIN empresas e ON e.id = p.empresa_id OR COALESCE(e.empresa_id, 0) = p.empresa_id
	          LEFT JOIN empresa_venta_publica_configuracion vpc ON vpc.empresa_id = p.empresa_id AND COALESCE(vpc.estado, 'activo') <> 'inactivo'
	          WHERE p.empresa_id = ? ORDER BY p.fecha_creacion DESC LIMIT ? OFFSET ?`
	rows, err := querySQLCompat(db, query, empresaID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pubs []PublicacionRedSocial
	for rows.Next() {
		var p PublicacionRedSocial
		var youtube string
		if err := rows.Scan(&p.ID, &p.EmpresaID, &p.EmpresaNombre, &p.EmpresaSlug, &p.Nombre, &p.Descripcion, &p.FotoURL, &youtube, &p.FechaCreacion, &p.Estado); err != nil {
			return nil, err
		}
		p.YoutubeURL = youtube
		p.ReaccionesResumen, _ = GetPublicacionReaccionesResumen(db, p.ID)
		p.ComentariosTotal, _ = GetPublicacionComentariosTotal(db, p.ID)
		pubs = append(pubs, p)
	}
	if pubs == nil {
		pubs = []PublicacionRedSocial{}
	}
	return pubs, nil
}

func InsertPublicacionRedSocial(db *sql.DB, p *PublicacionRedSocial) error {
	if err := EnsureEmpresaPublicacionesRedSocialSchema(db); err != nil {
		return err
	}
	query := `INSERT INTO empresa_publicaciones_red_social (empresa_id, nombre, descripcion, foto_url, youtube_url, estado)
	          VALUES (?, ?, ?, ?, ?, ?)`
	id, err := insertSQLCompat(db, query, p.EmpresaID, p.Nombre, p.Descripcion, p.FotoURL, p.YoutubeURL, p.Estado)
	if err != nil {
		return err
	}
	p.ID = int(id)
	return nil
}

func UpdatePublicacionRedSocial(db *sql.DB, p *PublicacionRedSocial) error {
	if err := EnsureEmpresaPublicacionesRedSocialSchema(db); err != nil {
		return err
	}
	query := `UPDATE empresa_publicaciones_red_social SET nombre=?, descripcion=?, foto_url=?, youtube_url=?, estado=? WHERE id=? AND empresa_id=?`
	_, err := execSQLCompat(db, query, p.Nombre, p.Descripcion, p.FotoURL, p.YoutubeURL, p.Estado, p.ID, p.EmpresaID)
	return err
}

func DeletePublicacionRedSocial(db *sql.DB, id, empresaID int) error {
	if err := EnsureEmpresaRedSocialInteraccionesSchema(db); err != nil {
		return err
	}
	query := `DELETE FROM empresa_publicaciones_red_social WHERE id=? AND empresa_id=?`
	_, err := execSQLCompat(db, query, id, empresaID)
	return err
}

func ListPublicacionComentarios(db *sql.DB, publicacionID int, limit, offset int) ([]PublicacionRedSocialComentario, error) {
	if err := EnsureEmpresaRedSocialInteraccionesSchema(db); err != nil {
		return nil, err
	}
	if publicacionID <= 0 {
		return []PublicacionRedSocialComentario{}, nil
	}
	limit = clampInt(limit, 20, 1, 100)
	if offset < 0 {
		offset = 0
	}
	rows, err := querySQLCompat(db, `SELECT
		id,
		COALESCE(publicacion_id, 0),
		COALESCE(empresa_id, 0),
		COALESCE(actor_key, ''),
		COALESCE(nombre, ''),
		COALESCE(contenido, ''),
		fecha_creacion,
		COALESCE(estado, 'activo')
	FROM empresa_publicaciones_red_social_comentarios
	WHERE publicacion_id = ?
		AND COALESCE(estado, 'activo') = 'activo'
	ORDER BY fecha_creacion DESC
	LIMIT ? OFFSET ?`, publicacionID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]PublicacionRedSocialComentario, 0)
	for rows.Next() {
		var item PublicacionRedSocialComentario
		if err := rows.Scan(&item.ID, &item.PublicacionID, &item.EmpresaID, &item.ActorKey, &item.Nombre, &item.Contenido, &item.FechaCreacion, &item.Estado); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	if out == nil {
		out = []PublicacionRedSocialComentario{}
	}
	return out, nil
}

func CreatePublicacionComentario(db *sql.DB, payload PublicacionRedSocialComentario) (int64, error) {
	if err := EnsureEmpresaRedSocialInteraccionesSchema(db); err != nil {
		return 0, err
	}
	if payload.PublicacionID <= 0 {
		return 0, fmt.Errorf("publicacion_id es obligatorio")
	}
	payload.ActorKey = normalizeActorKey(payload.ActorKey)
	payload.Nombre = strings.TrimSpace(payload.Nombre)
	payload.Contenido = strings.TrimSpace(payload.Contenido)
	if payload.Contenido == "" {
		return 0, fmt.Errorf("contenido es obligatorio")
	}
	if len(payload.Contenido) > 1500 {
		payload.Contenido = payload.Contenido[:1500]
	}
	if len(payload.Nombre) > 120 {
		payload.Nombre = payload.Nombre[:120]
	}
	id, err := insertSQLCompat(db, `INSERT INTO empresa_publicaciones_red_social_comentarios
		(publicacion_id, empresa_id, actor_key, nombre, contenido, estado)
	VALUES (?, ?, ?, ?, ?, 'activo')`,
		payload.PublicacionID, payload.EmpresaID, payload.ActorKey, payload.Nombre, payload.Contenido,
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func UpsertPublicacionReaccion(db *sql.DB, publicacionID int, empresaID int, actorKey, reaccion string) error {
	if err := EnsureEmpresaRedSocialInteraccionesSchema(db); err != nil {
		return err
	}
	if publicacionID <= 0 {
		return fmt.Errorf("publicacion_id es obligatorio")
	}
	actorKey = normalizeActorKey(actorKey)
	if actorKey == "" {
		return fmt.Errorf("actor_key es obligatorio")
	}
	reaccion = normalizeReaccion(reaccion)
	if reaccion == "" {
		return fmt.Errorf("reaccion invalida")
	}
	if shouldUsePostgresCompat(db) {
		_, err := execSQLCompat(db, `INSERT INTO empresa_publicaciones_red_social_reacciones
			(publicacion_id, empresa_id, actor_key, reaccion)
			VALUES ($1,$2,$3,$4)
			ON CONFLICT (publicacion_id, actor_key)
			DO UPDATE SET reaccion = EXCLUDED.reaccion, fecha = CURRENT_TIMESTAMP`,
			publicacionID, empresaID, actorKey, reaccion)
		return err
	}
	_, err := execSQLCompat(db, `INSERT INTO empresa_publicaciones_red_social_reacciones
		(publicacion_id, empresa_id, actor_key, reaccion)
		VALUES (?,?,?,?)
		ON CONFLICT(publicacion_id, actor_key)
		DO UPDATE SET reaccion=excluded.reaccion, fecha=CURRENT_TIMESTAMP`,
		publicacionID, empresaID, actorKey, reaccion)
	return err
}

func DeletePublicacionReaccion(db *sql.DB, publicacionID int, actorKey string) error {
	if err := EnsureEmpresaRedSocialInteraccionesSchema(db); err != nil {
		return err
	}
	if publicacionID <= 0 {
		return fmt.Errorf("publicacion_id es obligatorio")
	}
	actorKey = normalizeActorKey(actorKey)
	if actorKey == "" {
		return fmt.Errorf("actor_key es obligatorio")
	}
	_, err := execSQLCompat(db, `DELETE FROM empresa_publicaciones_red_social_reacciones WHERE publicacion_id=? AND actor_key=?`, publicacionID, actorKey)
	return err
}

func GetPublicacionReaccionesResumen(db *sql.DB, publicacionID int) (map[string]int, error) {
	if err := EnsureEmpresaRedSocialInteraccionesSchema(db); err != nil {
		return map[string]int{}, err
	}
	if publicacionID <= 0 {
		return map[string]int{}, nil
	}
	rows, err := querySQLCompat(db, `SELECT COALESCE(reaccion,''), COUNT(1)
		FROM empresa_publicaciones_red_social_reacciones
		WHERE publicacion_id = ?
		GROUP BY reaccion`, publicacionID)
	if err != nil {
		return map[string]int{}, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var reaccion string
		var count int
		if err := rows.Scan(&reaccion, &count); err != nil {
			return map[string]int{}, err
		}
		reaccion = normalizeReaccion(reaccion)
		if reaccion == "" {
			continue
		}
		out[reaccion] = count
	}
	return out, nil
}

func GetPublicacionComentariosTotal(db *sql.DB, publicacionID int) (int, error) {
	if err := EnsureEmpresaRedSocialInteraccionesSchema(db); err != nil {
		return 0, err
	}
	if publicacionID <= 0 {
		return 0, nil
	}
	var total int
	if err := queryRowSQLCompat(db, `SELECT COUNT(1)
		FROM empresa_publicaciones_red_social_comentarios
		WHERE publicacion_id = ?
		  AND COALESCE(estado, 'activo') = 'activo'`, publicacionID).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func GetUserReaction(db *sql.DB, publicacionID int, actorKey string) (string, error) {
	if err := EnsureEmpresaRedSocialInteraccionesSchema(db); err != nil {
		return "", err
	}
	if publicacionID <= 0 {
		return "", nil
	}
	actorKey = normalizeActorKey(actorKey)
	if actorKey == "" {
		return "", nil
	}
	var reaccion string
	err := queryRowSQLCompat(db, `SELECT COALESCE(reaccion,'')
		FROM empresa_publicaciones_red_social_reacciones
		WHERE publicacion_id = ? AND actor_key = ?
		ORDER BY id DESC
		LIMIT 1`, publicacionID, actorKey).Scan(&reaccion)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return normalizeReaccion(reaccion), nil
}

func SetRedSocialEmpresaSeguida(db *sql.DB, empresaID int, actorKey string, seguir bool) error {
	if err := EnsureEmpresaRedSocialInteraccionesSchema(db); err != nil {
		return err
	}
	if empresaID <= 0 {
		return fmt.Errorf("empresa_id es obligatorio")
	}
	actorKey = normalizeActorKey(actorKey)
	if actorKey == "" {
		return fmt.Errorf("actor_key es obligatorio")
	}
	if shouldUsePostgresCompat(db) {
		_, err := execSQLCompat(db, `INSERT INTO empresa_publicaciones_red_social_seguidores
			(empresa_id, actor_key, estado)
			VALUES ($1,$2,$3)
			ON CONFLICT (empresa_id, actor_key)
			DO UPDATE SET estado = EXCLUDED.estado`,
			empresaID, actorKey, map[bool]string{true: "activo", false: "inactivo"}[seguir])
		return err
	}
	_, err := execSQLCompat(db, `INSERT INTO empresa_publicaciones_red_social_seguidores
		(empresa_id, actor_key, estado)
		VALUES (?,?,?)
		ON CONFLICT(empresa_id, actor_key)
		DO UPDATE SET estado=excluded.estado`,
		empresaID, actorKey, map[bool]string{true: "activo", false: "inactivo"}[seguir])
	return err
}

func IsRedSocialEmpresaSeguida(db *sql.DB, empresaID int, actorKey string) (bool, error) {
	if err := EnsureEmpresaRedSocialInteraccionesSchema(db); err != nil {
		return false, err
	}
	actorKey = normalizeActorKey(actorKey)
	if empresaID <= 0 || actorKey == "" {
		return false, nil
	}
	var estado string
	err := queryRowSQLCompat(db, `SELECT COALESCE(estado,'') FROM empresa_publicaciones_red_social_seguidores WHERE empresa_id=? AND actor_key=? LIMIT 1`, empresaID, actorKey).Scan(&estado)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return strings.EqualFold(strings.TrimSpace(estado), "activo"), nil
}

func ListRedSocialSeguimientos(db *sql.DB, actorKey string) ([]RedSocialEmpresaSeguida, error) {
	if err := EnsureEmpresaRedSocialInteraccionesSchema(db); err != nil {
		return nil, err
	}
	actorKey = normalizeActorKey(actorKey)
	if actorKey == "" {
		return []RedSocialEmpresaSeguida{}, nil
	}
	rows, err := querySQLCompat(db, `SELECT
		s.empresa_id,
		COALESCE(e.nombre, ''),
		COALESCE(s.ultima_vista_publicacion_id, 0),
		COALESCE(lp.id, 0),
		COALESCE(lp.nombre, ''),
		COALESCE(CAST(lp.fecha_creacion AS TEXT), ''),
		COALESCE((
			SELECT COUNT(1)
			FROM empresa_publicaciones_red_social p
			WHERE p.empresa_id = s.empresa_id
			  AND p.estado = 'activo'
			  AND p.id > COALESCE(s.ultima_vista_publicacion_id, 0)
		), 0)
	FROM empresa_publicaciones_red_social_seguidores s
	LEFT JOIN empresas e ON e.id = s.empresa_id OR COALESCE(e.empresa_id, 0) = s.empresa_id
	LEFT JOIN empresa_publicaciones_red_social lp ON lp.id = (
		SELECT p2.id
		FROM empresa_publicaciones_red_social p2
		WHERE p2.empresa_id = s.empresa_id AND p2.estado = 'activo'
		ORDER BY p2.fecha_creacion DESC
		LIMIT 1
	)
	WHERE s.actor_key = ? AND COALESCE(s.estado, 'activo') = 'activo'
	ORDER BY lp.fecha_creacion DESC`, actorKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []RedSocialEmpresaSeguida{}
	for rows.Next() {
		var item RedSocialEmpresaSeguida
		var lastSeen int
		if err := rows.Scan(&item.EmpresaID, &item.EmpresaNombre, &lastSeen, &item.UltimaPublicacionID, &item.UltimaPublicacion, &item.UltimaPublicacionFecha, &item.Nuevas); err != nil {
			return nil, err
		}
		item.Seguido = true
		out = append(out, item)
	}
	return out, nil
}

func MarkRedSocialSeguimientoVisto(db *sql.DB, empresaID int, actorKey string) error {
	if err := EnsureEmpresaRedSocialInteraccionesSchema(db); err != nil {
		return err
	}
	actorKey = normalizeActorKey(actorKey)
	if empresaID <= 0 || actorKey == "" {
		return nil
	}
	_, err := execSQLCompat(db, `UPDATE empresa_publicaciones_red_social_seguidores
		SET ultima_vista_publicacion_id = COALESCE((
			SELECT MAX(id) FROM empresa_publicaciones_red_social WHERE empresa_id = ? AND estado = 'activo'
		), ultima_vista_publicacion_id)
		WHERE empresa_id = ? AND actor_key = ?`, empresaID, empresaID, actorKey)
	return err
}
