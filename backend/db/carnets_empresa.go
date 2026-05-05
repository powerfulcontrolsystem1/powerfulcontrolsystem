package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type EmpresaCarnetPlantilla struct {
	ID                  int64  `json:"id"`
	EmpresaID           int64  `json:"empresa_id"`
	Nombre              string `json:"nombre"`
	Tipo                string `json:"tipo"`
	Orientacion         string `json:"orientacion"`
	AnchoMM             int    `json:"ancho_mm"`
	AltoMM              int    `json:"alto_mm"`
	ColorPrimario       string `json:"color_primario"`
	ColorSecundario     string `json:"color_secundario"`
	ColorTexto          string `json:"color_texto"`
	MostrarLogo         bool   `json:"mostrar_logo"`
	MostrarFoto         bool   `json:"mostrar_foto"`
	MostrarQR           bool   `json:"mostrar_qr"`
	MostrarCodigoBarras bool   `json:"mostrar_codigo_barras"`
	CamposVisibles      string `json:"campos_visibles"`
	DisenoJSON          string `json:"diseno_json"`
	EsPredeterminada    bool   `json:"es_predeterminada"`
	FechaCreacion       string `json:"fecha_creacion,omitempty"`
	FechaActualizacion  string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador      string `json:"usuario_creador,omitempty"`
	Estado              string `json:"estado,omitempty"`
	Observaciones       string `json:"observaciones,omitempty"`
}

type EmpresaCarnet struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	PlantillaID        int64  `json:"plantilla_id"`
	PlantillaNombre    string `json:"plantilla_nombre,omitempty"`
	UsuarioID          int64  `json:"usuario_id"`
	Codigo             string `json:"codigo"`
	TipoPersona        string `json:"tipo_persona"`
	NombreCompleto     string `json:"nombre_completo"`
	Documento          string `json:"documento"`
	Cargo              string `json:"cargo"`
	Area               string `json:"area"`
	Email              string `json:"email"`
	Telefono           string `json:"telefono"`
	FotoURL            string `json:"foto_url"`
	NivelAcceso        string `json:"nivel_acceso"`
	GrupoSanguineo     string `json:"grupo_sanguineo"`
	ContactoEmergencia string `json:"contacto_emergencia"`
	TelefonoEmergencia string `json:"telefono_emergencia"`
	FechaEmision       string `json:"fecha_emision"`
	FechaVencimiento   string `json:"fecha_vencimiento"`
	QRPayload          string `json:"qr_payload"`
	EstadoCarnet       string `json:"estado_carnet"`
	UltimaImpresion    string `json:"ultima_impresion,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

type EmpresaCarnetEvento struct {
	ID             int64  `json:"id"`
	EmpresaID      int64  `json:"empresa_id"`
	CarnetID       int64  `json:"carnet_id"`
	Evento         string `json:"evento"`
	Detalle        string `json:"detalle"`
	UsuarioCreador string `json:"usuario_creador,omitempty"`
	FechaCreacion  string `json:"fecha_creacion,omitempty"`
}

type EmpresaCarnetPersonaFuente struct {
	ID        int64  `json:"id"`
	Nombre    string `json:"nombre"`
	Documento string `json:"documento,omitempty"`
	Email     string `json:"email,omitempty"`
	Cargo     string `json:"cargo,omitempty"`
	Estado    string `json:"estado,omitempty"`
	Origen    string `json:"origen"`
}

type EmpresaCarnetsDashboard struct {
	EmpresaID         int64                    `json:"empresa_id"`
	Total             int64                    `json:"total"`
	Vigentes          int64                    `json:"vigentes"`
	Vencidos          int64                    `json:"vencidos"`
	Suspendidos       int64                    `json:"suspendidos"`
	Revocados         int64                    `json:"revocados"`
	PlantillasActivas int64                    `json:"plantillas_activas"`
	Recientes         []EmpresaCarnet          `json:"recientes"`
	PorTipo           map[string]int64         `json:"por_tipo"`
	Plantillas        []EmpresaCarnetPlantilla `json:"plantillas"`
}

func EnsureEmpresaCarnetsSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_carnets_plantillas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			nombre TEXT NOT NULL,
			tipo TEXT DEFAULT 'empleado',
			orientacion TEXT DEFAULT 'vertical',
			ancho_mm INTEGER DEFAULT 54,
			alto_mm INTEGER DEFAULT 86,
			color_primario TEXT DEFAULT '#1f6feb',
			color_secundario TEXT DEFAULT '#0f172a',
			color_texto TEXT DEFAULT '#ffffff',
			mostrar_logo INTEGER DEFAULT 1,
			mostrar_foto INTEGER DEFAULT 1,
			mostrar_qr INTEGER DEFAULT 1,
			mostrar_codigo_barras INTEGER DEFAULT 0,
			campos_visibles TEXT DEFAULT 'documento,cargo,area,email,nivel_acceso,vencimiento',
			diseno_json TEXT,
			es_predeterminada INTEGER DEFAULT 0,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_carnets_plantillas_empresa ON empresa_carnets_plantillas(empresa_id, estado, tipo);`,
		`CREATE TABLE IF NOT EXISTS empresa_carnets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			plantilla_id INTEGER,
			usuario_id INTEGER,
			codigo TEXT NOT NULL,
			tipo_persona TEXT DEFAULT 'empleado',
			nombre_completo TEXT NOT NULL,
			documento TEXT,
			cargo TEXT,
			area TEXT,
			email TEXT,
			telefono TEXT,
			foto_url TEXT,
			nivel_acceso TEXT DEFAULT 'general',
			grupo_sanguineo TEXT,
			contacto_emergencia TEXT,
			telefono_emergencia TEXT,
			fecha_emision TEXT,
			fecha_vencimiento TEXT,
			qr_payload TEXT,
			estado_carnet TEXT DEFAULT 'vigente',
			ultima_impresion TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_carnets_codigo ON empresa_carnets(empresa_id, codigo);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_carnets_empresa_estado ON empresa_carnets(empresa_id, estado, estado_carnet);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_carnets_usuario ON empresa_carnets(empresa_id, usuario_id);`,
		`CREATE TABLE IF NOT EXISTS empresa_carnets_eventos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			carnet_id INTEGER NOT NULL,
			evento TEXT NOT NULL,
			detalle TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo'
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_carnets_eventos_empresa ON empresa_carnets_eventos(empresa_id, carnet_id, id DESC);`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func boolToDBInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func dbIntToBool(v int) bool {
	return v != 0
}

func normalizeCarnetTipo(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "empleado", "usuario", "contratista", "visitante", "temporal", "directivo":
		return v
	default:
		return "empleado"
	}
}

func normalizeCarnetEstado(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "vigente", "suspendido", "vencido", "revocado", "pendiente":
		return v
	default:
		return "vigente"
	}
}

func normalizeCarnetOrientacion(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	if v == "horizontal" {
		return v
	}
	return "vertical"
}

func defaultEmpresaCarnetPlantilla(empresaID int64) EmpresaCarnetPlantilla {
	return EmpresaCarnetPlantilla{
		EmpresaID:           empresaID,
		Nombre:              "Carnet corporativo moderno",
		Tipo:                "empleado",
		Orientacion:         "vertical",
		AnchoMM:             54,
		AltoMM:              86,
		ColorPrimario:       "#1f6feb",
		ColorSecundario:     "#0f172a",
		ColorTexto:          "#ffffff",
		MostrarLogo:         true,
		MostrarFoto:         true,
		MostrarQR:           true,
		MostrarCodigoBarras: false,
		CamposVisibles:      "documento,cargo,area,email,nivel_acceso,vencimiento",
		EsPredeterminada:    true,
		Estado:              "activo",
	}
}

func normalizeCarnetPlantilla(p EmpresaCarnetPlantilla) EmpresaCarnetPlantilla {
	if p.EmpresaID <= 0 {
		return p
	}
	if strings.TrimSpace(p.Nombre) == "" {
		p.Nombre = "Carnet corporativo"
	}
	p.Tipo = normalizeCarnetTipo(p.Tipo)
	p.Orientacion = normalizeCarnetOrientacion(p.Orientacion)
	if p.AnchoMM <= 0 {
		p.AnchoMM = 54
	}
	if p.AltoMM <= 0 {
		p.AltoMM = 86
	}
	if strings.TrimSpace(p.ColorPrimario) == "" {
		p.ColorPrimario = "#1f6feb"
	}
	if strings.TrimSpace(p.ColorSecundario) == "" {
		p.ColorSecundario = "#0f172a"
	}
	if strings.TrimSpace(p.ColorTexto) == "" {
		p.ColorTexto = "#ffffff"
	}
	if strings.TrimSpace(p.CamposVisibles) == "" {
		p.CamposVisibles = "documento,cargo,area,email,nivel_acceso,vencimiento"
	}
	if strings.TrimSpace(p.Estado) == "" {
		p.Estado = "activo"
	}
	return p
}

func EnsureEmpresaCarnetDefaultTemplate(dbConn *sql.DB, empresaID int64, usuario string) (int64, error) {
	if empresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaCarnetsSchema(dbConn); err != nil {
		return 0, err
	}
	var id int64
	err := QueryRowCompat(dbConn, `SELECT id FROM empresa_carnets_plantillas WHERE empresa_id=? AND COALESCE(estado,'activo')='activo' ORDER BY es_predeterminada DESC, id ASC LIMIT 1`, empresaID).Scan(&id)
	if err == nil && id > 0 {
		return id, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	p := defaultEmpresaCarnetPlantilla(empresaID)
	p.UsuarioCreador = usuario
	return UpsertEmpresaCarnetPlantilla(dbConn, p)
}

func UpsertEmpresaCarnetPlantilla(dbConn *sql.DB, p EmpresaCarnetPlantilla) (int64, error) {
	if err := EnsureEmpresaCarnetsSchema(dbConn); err != nil {
		return 0, err
	}
	p = normalizeCarnetPlantilla(p)
	if p.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	if p.EsPredeterminada {
		if _, err := ExecCompat(dbConn, `UPDATE empresa_carnets_plantillas SET es_predeterminada=0 WHERE empresa_id=?`, p.EmpresaID); err != nil {
			return 0, err
		}
	}
	if p.ID > 0 {
		res, err := ExecCompat(dbConn, `UPDATE empresa_carnets_plantillas SET nombre=?, tipo=?, orientacion=?, ancho_mm=?, alto_mm=?, color_primario=?, color_secundario=?, color_texto=?, mostrar_logo=?, mostrar_foto=?, mostrar_qr=?, mostrar_codigo_barras=?, campos_visibles=?, diseno_json=?, es_predeterminada=?, fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=?, estado=?, observaciones=? WHERE empresa_id=? AND id=?`,
			strings.TrimSpace(p.Nombre), p.Tipo, p.Orientacion, p.AnchoMM, p.AltoMM, strings.TrimSpace(p.ColorPrimario), strings.TrimSpace(p.ColorSecundario), strings.TrimSpace(p.ColorTexto), boolToDBInt(p.MostrarLogo), boolToDBInt(p.MostrarFoto), boolToDBInt(p.MostrarQR), boolToDBInt(p.MostrarCodigoBarras), strings.TrimSpace(p.CamposVisibles), strings.TrimSpace(p.DisenoJSON), boolToDBInt(p.EsPredeterminada), strings.TrimSpace(p.UsuarioCreador), strings.TrimSpace(p.Estado), strings.TrimSpace(p.Observaciones), p.EmpresaID, p.ID)
		if err != nil {
			return 0, err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return 0, sql.ErrNoRows
		}
		return p.ID, nil
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_carnets_plantillas (empresa_id, nombre, tipo, orientacion, ancho_mm, alto_mm, color_primario, color_secundario, color_texto, mostrar_logo, mostrar_foto, mostrar_qr, mostrar_codigo_barras, campos_visibles, diseno_json, es_predeterminada, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?, ?)`,
		p.EmpresaID, strings.TrimSpace(p.Nombre), p.Tipo, p.Orientacion, p.AnchoMM, p.AltoMM, strings.TrimSpace(p.ColorPrimario), strings.TrimSpace(p.ColorSecundario), strings.TrimSpace(p.ColorTexto), boolToDBInt(p.MostrarLogo), boolToDBInt(p.MostrarFoto), boolToDBInt(p.MostrarQR), boolToDBInt(p.MostrarCodigoBarras), strings.TrimSpace(p.CamposVisibles), strings.TrimSpace(p.DisenoJSON), boolToDBInt(p.EsPredeterminada), strings.TrimSpace(p.UsuarioCreador), strings.TrimSpace(p.Estado), strings.TrimSpace(p.Observaciones))
}

func ListEmpresaCarnetPlantillas(dbConn *sql.DB, empresaID int64, includeInactive bool) ([]EmpresaCarnetPlantilla, error) {
	if err := EnsureEmpresaCarnetsSchema(dbConn); err != nil {
		return nil, err
	}
	query := `SELECT id, empresa_id, COALESCE(nombre,''), COALESCE(tipo,''), COALESCE(orientacion,''), COALESCE(ancho_mm,54), COALESCE(alto_mm,86), COALESCE(color_primario,''), COALESCE(color_secundario,''), COALESCE(color_texto,''), COALESCE(mostrar_logo,1), COALESCE(mostrar_foto,1), COALESCE(mostrar_qr,1), COALESCE(mostrar_codigo_barras,0), COALESCE(campos_visibles,''), COALESCE(diseno_json,''), COALESCE(es_predeterminada,0), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,''), COALESCE(estado,'activo'), COALESCE(observaciones,'') FROM empresa_carnets_plantillas WHERE empresa_id=?`
	args := []interface{}{empresaID}
	if !includeInactive {
		query += ` AND COALESCE(estado,'activo')='activo'`
	}
	query += ` ORDER BY es_predeterminada DESC, id DESC`
	rows, err := ExecQueryCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCarnetPlantilla{}
	for rows.Next() {
		var p EmpresaCarnetPlantilla
		var logo, foto, qr, bar, pred int
		if err := rows.Scan(&p.ID, &p.EmpresaID, &p.Nombre, &p.Tipo, &p.Orientacion, &p.AnchoMM, &p.AltoMM, &p.ColorPrimario, &p.ColorSecundario, &p.ColorTexto, &logo, &foto, &qr, &bar, &p.CamposVisibles, &p.DisenoJSON, &pred, &p.FechaCreacion, &p.FechaActualizacion, &p.UsuarioCreador, &p.Estado, &p.Observaciones); err != nil {
			return nil, err
		}
		p.MostrarLogo, p.MostrarFoto, p.MostrarQR, p.MostrarCodigoBarras, p.EsPredeterminada = dbIntToBool(logo), dbIntToBool(foto), dbIntToBool(qr), dbIntToBool(bar), dbIntToBool(pred)
		out = append(out, p)
	}
	return out, rows.Err()
}

func normalizeEmpresaCarnet(c EmpresaCarnet) EmpresaCarnet {
	c.TipoPersona = normalizeCarnetTipo(c.TipoPersona)
	c.EstadoCarnet = normalizeCarnetEstado(c.EstadoCarnet)
	if strings.TrimSpace(c.Estado) == "" {
		c.Estado = "activo"
	}
	if strings.TrimSpace(c.NivelAcceso) == "" {
		c.NivelAcceso = "general"
	}
	if strings.TrimSpace(c.FechaEmision) == "" {
		c.FechaEmision = time.Now().Format("2006-01-02")
	}
	return c
}

func nextEmpresaCarnetCodigo(dbConn *sql.DB, empresaID int64, tipo string) string {
	prefix := "CRN"
	switch normalizeCarnetTipo(tipo) {
	case "empleado":
		prefix = "EMP"
	case "usuario":
		prefix = "USR"
	case "contratista":
		prefix = "CON"
	case "visitante":
		prefix = "VIS"
	case "directivo":
		prefix = "DIR"
	}
	var count int64
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_carnets WHERE empresa_id=?`, empresaID).Scan(&count)
	return fmt.Sprintf("%s-%d-%05d", prefix, empresaID, count+1)
}

func CreateEmpresaCarnet(dbConn *sql.DB, c EmpresaCarnet) (int64, error) {
	if err := EnsureEmpresaCarnetsSchema(dbConn); err != nil {
		return 0, err
	}
	c = normalizeEmpresaCarnet(c)
	if c.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	if strings.TrimSpace(c.NombreCompleto) == "" {
		return 0, fmt.Errorf("nombre_completo es obligatorio")
	}
	if c.PlantillaID <= 0 {
		id, err := EnsureEmpresaCarnetDefaultTemplate(dbConn, c.EmpresaID, c.UsuarioCreador)
		if err != nil {
			return 0, err
		}
		c.PlantillaID = id
	}
	if strings.TrimSpace(c.Codigo) == "" {
		c.Codigo = nextEmpresaCarnetCodigo(dbConn, c.EmpresaID, c.TipoPersona)
	}
	if strings.TrimSpace(c.QRPayload) == "" {
		c.QRPayload = fmt.Sprintf("PCS:CARNET:%d:%s", c.EmpresaID, strings.TrimSpace(c.Codigo))
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_carnets (empresa_id, plantilla_id, usuario_id, codigo, tipo_persona, nombre_completo, documento, cargo, area, email, telefono, foto_url, nivel_acceso, grupo_sanguineo, contacto_emergencia, telefono_emergencia, fecha_emision, fecha_vencimiento, qr_payload, estado_carnet, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?, ?)`,
		c.EmpresaID, c.PlantillaID, c.UsuarioID, strings.TrimSpace(c.Codigo), c.TipoPersona, strings.TrimSpace(c.NombreCompleto), strings.TrimSpace(c.Documento), strings.TrimSpace(c.Cargo), strings.TrimSpace(c.Area), strings.TrimSpace(c.Email), strings.TrimSpace(c.Telefono), strings.TrimSpace(c.FotoURL), strings.TrimSpace(c.NivelAcceso), strings.TrimSpace(c.GrupoSanguineo), strings.TrimSpace(c.ContactoEmergencia), strings.TrimSpace(c.TelefonoEmergencia), strings.TrimSpace(c.FechaEmision), strings.TrimSpace(c.FechaVencimiento), strings.TrimSpace(c.QRPayload), c.EstadoCarnet, strings.TrimSpace(c.UsuarioCreador), strings.TrimSpace(c.Estado), strings.TrimSpace(c.Observaciones))
	if err != nil {
		return 0, err
	}
	_ = AddEmpresaCarnetEvento(dbConn, c.EmpresaID, id, "emitido", "Carnet emitido", c.UsuarioCreador)
	return id, nil
}

func UpdateEmpresaCarnet(dbConn *sql.DB, c EmpresaCarnet) error {
	if err := EnsureEmpresaCarnetsSchema(dbConn); err != nil {
		return err
	}
	c = normalizeEmpresaCarnet(c)
	if c.EmpresaID <= 0 || c.ID <= 0 {
		return fmt.Errorf("empresa_id e id son obligatorios")
	}
	res, err := ExecCompat(dbConn, `UPDATE empresa_carnets SET plantilla_id=?, usuario_id=?, codigo=?, tipo_persona=?, nombre_completo=?, documento=?, cargo=?, area=?, email=?, telefono=?, foto_url=?, nivel_acceso=?, grupo_sanguineo=?, contacto_emergencia=?, telefono_emergencia=?, fecha_emision=?, fecha_vencimiento=?, qr_payload=?, estado_carnet=?, fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=?, estado=?, observaciones=? WHERE empresa_id=? AND id=?`,
		c.PlantillaID, c.UsuarioID, strings.TrimSpace(c.Codigo), c.TipoPersona, strings.TrimSpace(c.NombreCompleto), strings.TrimSpace(c.Documento), strings.TrimSpace(c.Cargo), strings.TrimSpace(c.Area), strings.TrimSpace(c.Email), strings.TrimSpace(c.Telefono), strings.TrimSpace(c.FotoURL), strings.TrimSpace(c.NivelAcceso), strings.TrimSpace(c.GrupoSanguineo), strings.TrimSpace(c.ContactoEmergencia), strings.TrimSpace(c.TelefonoEmergencia), strings.TrimSpace(c.FechaEmision), strings.TrimSpace(c.FechaVencimiento), strings.TrimSpace(c.QRPayload), c.EstadoCarnet, strings.TrimSpace(c.UsuarioCreador), strings.TrimSpace(c.Estado), strings.TrimSpace(c.Observaciones), c.EmpresaID, c.ID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	_ = AddEmpresaCarnetEvento(dbConn, c.EmpresaID, c.ID, "actualizado", "Carnet actualizado", c.UsuarioCreador)
	return nil
}

func SetEmpresaCarnetEstado(dbConn *sql.DB, empresaID, id int64, estadoCarnet, usuario, detalle string) error {
	estadoCarnet = normalizeCarnetEstado(estadoCarnet)
	res, err := ExecCompat(dbConn, `UPDATE empresa_carnets SET estado_carnet=?, fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=? WHERE empresa_id=? AND id=?`, estadoCarnet, strings.TrimSpace(usuario), empresaID, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return AddEmpresaCarnetEvento(dbConn, empresaID, id, estadoCarnet, detailOrDefault(detalle, "Cambio de estado"), usuario)
}

func MarkEmpresaCarnetImpreso(dbConn *sql.DB, empresaID, id int64, usuario string) error {
	res, err := ExecCompat(dbConn, `UPDATE empresa_carnets SET ultima_impresion=CURRENT_TIMESTAMP, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, empresaID, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return AddEmpresaCarnetEvento(dbConn, empresaID, id, "impreso", "Carnet marcado como impreso/exportado", usuario)
}

func detailOrDefault(v, fallback string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return fallback
	}
	return v
}

func ListEmpresaCarnets(dbConn *sql.DB, empresaID int64, includeInactive bool, estadoCarnet, tipoPersona, q string, limit int) ([]EmpresaCarnet, error) {
	if err := EnsureEmpresaCarnetsSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	query := `SELECT c.id, c.empresa_id, COALESCE(c.plantilla_id,0), COALESCE(p.nombre,''), COALESCE(c.usuario_id,0), COALESCE(c.codigo,''), COALESCE(c.tipo_persona,''), COALESCE(c.nombre_completo,''), COALESCE(c.documento,''), COALESCE(c.cargo,''), COALESCE(c.area,''), COALESCE(c.email,''), COALESCE(c.telefono,''), COALESCE(c.foto_url,''), COALESCE(c.nivel_acceso,''), COALESCE(c.grupo_sanguineo,''), COALESCE(c.contacto_emergencia,''), COALESCE(c.telefono_emergencia,''), COALESCE(c.fecha_emision,''), COALESCE(c.fecha_vencimiento,''), COALESCE(c.qr_payload,''), COALESCE(c.estado_carnet,''), COALESCE(c.ultima_impresion,''), COALESCE(c.fecha_creacion,''), COALESCE(c.fecha_actualizacion,''), COALESCE(c.usuario_creador,''), COALESCE(c.estado,'activo'), COALESCE(c.observaciones,'') FROM empresa_carnets c LEFT JOIN empresa_carnets_plantillas p ON p.empresa_id=c.empresa_id AND p.id=c.plantilla_id WHERE c.empresa_id=?`
	args := []interface{}{empresaID}
	if !includeInactive {
		query += ` AND COALESCE(c.estado,'activo')='activo'`
	}
	if estado := strings.TrimSpace(estadoCarnet); estado != "" {
		query += ` AND c.estado_carnet=?`
		args = append(args, normalizeCarnetEstado(estado))
	}
	if tipo := strings.TrimSpace(tipoPersona); tipo != "" {
		query += ` AND c.tipo_persona=?`
		args = append(args, normalizeCarnetTipo(tipo))
	}
	if search := strings.ToLower(strings.TrimSpace(q)); search != "" {
		query += ` AND (LOWER(COALESCE(c.nombre_completo,'')) LIKE ? OR LOWER(COALESCE(c.documento,'')) LIKE ? OR LOWER(COALESCE(c.codigo,'')) LIKE ? OR LOWER(COALESCE(c.email,'')) LIKE ? OR LOWER(COALESCE(c.cargo,'')) LIKE ?)`
		like := "%" + search + "%"
		args = append(args, like, like, like, like, like)
	}
	query += ` ORDER BY c.id DESC LIMIT ?`
	args = append(args, limit)
	rows, err := ExecQueryCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCarnet{}
	for rows.Next() {
		var c EmpresaCarnet
		if err := rows.Scan(&c.ID, &c.EmpresaID, &c.PlantillaID, &c.PlantillaNombre, &c.UsuarioID, &c.Codigo, &c.TipoPersona, &c.NombreCompleto, &c.Documento, &c.Cargo, &c.Area, &c.Email, &c.Telefono, &c.FotoURL, &c.NivelAcceso, &c.GrupoSanguineo, &c.ContactoEmergencia, &c.TelefonoEmergencia, &c.FechaEmision, &c.FechaVencimiento, &c.QRPayload, &c.EstadoCarnet, &c.UltimaImpresion, &c.FechaCreacion, &c.FechaActualizacion, &c.UsuarioCreador, &c.Estado, &c.Observaciones); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func AddEmpresaCarnetEvento(dbConn *sql.DB, empresaID, carnetID int64, evento, detalle, usuario string) error {
	if empresaID <= 0 || carnetID <= 0 {
		return nil
	}
	_, err := insertSQLCompat(dbConn, `INSERT INTO empresa_carnets_eventos (empresa_id, carnet_id, evento, detalle, fecha_creacion, usuario_creador, estado) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, ?, 'activo')`, empresaID, carnetID, strings.TrimSpace(evento), strings.TrimSpace(detalle), strings.TrimSpace(usuario))
	return err
}

func ListEmpresaCarnetEventos(dbConn *sql.DB, empresaID, carnetID int64, limit int) ([]EmpresaCarnetEvento, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT id, empresa_id, carnet_id, COALESCE(evento,''), COALESCE(detalle,''), COALESCE(usuario_creador,''), COALESCE(fecha_creacion,'') FROM empresa_carnets_eventos WHERE empresa_id=? AND carnet_id=? ORDER BY id DESC LIMIT ?`, empresaID, carnetID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCarnetEvento{}
	for rows.Next() {
		var e EmpresaCarnetEvento
		if err := rows.Scan(&e.ID, &e.EmpresaID, &e.CarnetID, &e.Evento, &e.Detalle, &e.UsuarioCreador, &e.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func ListEmpresaCarnetPersonasFuente(dbConn *sql.DB, empresaID int64, q string, limit int) ([]EmpresaCarnetPersonaFuente, error) {
	if err := EnsureEmpresaUsuariosAuthSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 300 {
		limit = 100
	}
	query := `SELECT id, COALESCE(name,''), COALESCE(documento_identidad,''), COALESCE(email,''), COALESCE(role,''), COALESCE(estado,'activo') FROM users WHERE empresa_id=?`
	args := []interface{}{empresaID}
	if search := strings.ToLower(strings.TrimSpace(q)); search != "" {
		query += ` AND (LOWER(COALESCE(name,'')) LIKE ? OR LOWER(COALESCE(email,'')) LIKE ? OR LOWER(COALESCE(documento_identidad,'')) LIKE ? OR LOWER(COALESCE(role,'')) LIKE ?)`
		like := "%" + search + "%"
		args = append(args, like, like, like, like)
	}
	query += ` ORDER BY name, email LIMIT ?`
	args = append(args, limit)
	rows, err := ExecQueryCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCarnetPersonaFuente{}
	for rows.Next() {
		var p EmpresaCarnetPersonaFuente
		if err := rows.Scan(&p.ID, &p.Nombre, &p.Documento, &p.Email, &p.Cargo, &p.Estado); err != nil {
			return nil, err
		}
		p.Origen = "usuario"
		out = append(out, p)
	}
	return out, rows.Err()
}

func BuildEmpresaCarnetsDashboard(dbConn *sql.DB, empresaID int64) (*EmpresaCarnetsDashboard, error) {
	if err := EnsureEmpresaCarnetsSchema(dbConn); err != nil {
		return nil, err
	}
	d := &EmpresaCarnetsDashboard{EmpresaID: empresaID, PorTipo: map[string]int64{}}
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_carnets WHERE empresa_id=? AND COALESCE(estado,'activo')='activo'`, empresaID).Scan(&d.Total)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_carnets WHERE empresa_id=? AND COALESCE(estado,'activo')='activo' AND estado_carnet='vigente'`, empresaID).Scan(&d.Vigentes)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_carnets WHERE empresa_id=? AND COALESCE(estado,'activo')='activo' AND estado_carnet='vencido'`, empresaID).Scan(&d.Vencidos)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_carnets WHERE empresa_id=? AND COALESCE(estado,'activo')='activo' AND estado_carnet='suspendido'`, empresaID).Scan(&d.Suspendidos)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_carnets WHERE empresa_id=? AND COALESCE(estado,'activo')='activo' AND estado_carnet='revocado'`, empresaID).Scan(&d.Revocados)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_carnets_plantillas WHERE empresa_id=? AND COALESCE(estado,'activo')='activo'`, empresaID).Scan(&d.PlantillasActivas)
	rows, err := ExecQueryCompat(dbConn, `SELECT tipo_persona, COUNT(1) FROM empresa_carnets WHERE empresa_id=? AND COALESCE(estado,'activo')='activo' GROUP BY tipo_persona`, empresaID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var tipo string
			var total int64
			if rows.Scan(&tipo, &total) == nil {
				d.PorTipo[tipo] = total
			}
		}
	}
	recientes, err := ListEmpresaCarnets(dbConn, empresaID, false, "", "", "", 8)
	if err != nil {
		return nil, err
	}
	plantillas, err := ListEmpresaCarnetPlantillas(dbConn, empresaID, false)
	if err != nil {
		return nil, err
	}
	d.Recientes = recientes
	d.Plantillas = plantillas
	return d, nil
}

func SeedEmpresaCarnetsDemo(dbConn *sql.DB, empresaID int64, usuario string) error {
	templateID, err := EnsureEmpresaCarnetDefaultTemplate(dbConn, empresaID, usuario)
	if err != nil {
		return err
	}
	examples := []EmpresaCarnet{
		{EmpresaID: empresaID, PlantillaID: templateID, TipoPersona: "empleado", NombreCompleto: "Laura Ramirez", Documento: "CC 1020304050", Cargo: "Administradora", Area: "Operacion", Email: "laura.demo@empresa.local", NivelAcceso: "administracion", GrupoSanguineo: "O+", FechaVencimiento: time.Now().AddDate(1, 0, 0).Format("2006-01-02"), UsuarioCreador: usuario},
		{EmpresaID: empresaID, PlantillaID: templateID, TipoPersona: "usuario", NombreCompleto: "Carlos Gomez", Documento: "CC 900100200", Cargo: "Cajero", Area: "Ventas", Email: "carlos.demo@empresa.local", NivelAcceso: "pos", GrupoSanguineo: "A+", FechaVencimiento: time.Now().AddDate(1, 0, 0).Format("2006-01-02"), UsuarioCreador: usuario},
		{EmpresaID: empresaID, PlantillaID: templateID, TipoPersona: "contratista", NombreCompleto: "Diana Torres", Documento: "CC 80100200", Cargo: "Soporte tecnico", Area: "Mantenimiento", Email: "diana.demo@proveedor.local", NivelAcceso: "temporal", GrupoSanguineo: "B+", FechaVencimiento: time.Now().AddDate(0, 6, 0).Format("2006-01-02"), UsuarioCreador: usuario},
	}
	for _, item := range examples {
		if _, err := CreateEmpresaCarnet(dbConn, item); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate") && !strings.Contains(strings.ToLower(err.Error()), "unique") {
			return err
		}
	}
	return nil
}
