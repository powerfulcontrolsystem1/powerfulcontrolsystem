package db

import (
	"database/sql"
	"errors"
	"sort"
	"strconv"
	"strings"
)

// RolDeUsuario define un rol configurable por tipo de empresa.
type RolDeUsuario struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id,omitempty"`
	TipoEmpresaID      int64  `json:"tipo_empresa_id"`
	TipoEmpresaNombre  string `json:"tipo_empresa_nombre,omitempty"`
	Nombre             string `json:"nombre"`
	Descripcion        string `json:"descripcion,omitempty"`
	Origen             string `json:"origen,omitempty"`
	RolBaseID          int64  `json:"rol_base_id,omitempty"`
	Personalizado      bool   `json:"personalizado,omitempty"`
	Asignable          bool   `json:"asignable"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// RolesDeUsuarioSchemaReady verifica el catálogo migrado sin DDL en peticiones.
func RolesDeUsuarioSchemaReady(dbConn *sql.DB) error {
	return requireSchemaReadiness(dbConn, "catalogo de roles", []schemaReadinessCheck{
		{"roles", `SELECT id, tipo_empresa_id, empresa_id, nombre, origen, rol_base_id, estado FROM roles_de_usuario WHERE 1=0`},
		{"tipos", `SELECT id, nombre FROM tipos_de_empresas WHERE 1=0`},
	})
}

// IsRolDeUsuarioAsignable impide convertir un rol de plataforma en usuario empresarial.
// En roles propios también debe verificarse el rol base con alcance empresarial.
func IsRolDeUsuarioAsignable(rol *RolDeUsuario) bool {
	if rol == nil || !strings.EqualFold(strings.TrimSpace(rol.Estado), "activo") {
		return false
	}
	switch normalizeRolCatalogKey(rol.Nombre) {
	case "super", "superadmin", "super_admin", "superadministrador", "super_administrador", "administrador_total", "admin_total", "admin_full", "full_admin":
		return false
	}
	return true
}

// EnsureRolesDeUsuarioSchema crea/migra la tabla base de roles por tipo de empresa.
func EnsureRolesDeUsuarioSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS roles_de_usuario (
			id BIGSERIAL PRIMARY KEY,
			tipo_empresa_id BIGINT NOT NULL,
			nombre TEXT NOT NULL,
			descripcion TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_roles_de_usuario_tipo ON roles_de_usuario(tipo_empresa_id);`,
		`CREATE INDEX IF NOT EXISTS ix_roles_de_usuario_tipo_nombre ON roles_de_usuario(tipo_empresa_id, nombre);`,
	}
	for _, stmt := range stmts {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	for _, col := range []struct {
		name string
		def  string
	}{
		{"tipo_empresa_id", "BIGINT DEFAULT 0"},
		{"empresa_id", "BIGINT DEFAULT 0"},
		{"nombre", "TEXT"},
		{"descripcion", "TEXT"},
		{"origen", "TEXT DEFAULT 'global'"},
		{"rol_base_id", "BIGINT DEFAULT 0"},
		{"fecha_creacion", "TEXT DEFAULT (CURRENT_TIMESTAMP)"},
		{"fecha_actualizacion", "TEXT DEFAULT (CURRENT_TIMESTAMP)"},
		{"usuario_creador", "TEXT"},
		{"estado", "TEXT DEFAULT 'activo'"},
		{"observaciones", "TEXT"},
	} {
		if err := ensureColumnIfMissing(dbConn, "roles_de_usuario", col.name, col.def); err != nil {
			return err
		}
	}
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS ix_roles_de_usuario_empresa ON roles_de_usuario(empresa_id);`,
		`CREATE INDEX IF NOT EXISTS ix_roles_de_usuario_empresa_nombre ON roles_de_usuario(empresa_id, nombre);`,
		`CREATE INDEX IF NOT EXISTS ix_roles_de_usuario_origen ON roles_de_usuario(origen);`,
	}
	for _, stmt := range indexes {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

// CreateRolDeUsuario crea un rol de usuario para un tipo de empresa.
func CreateRolDeUsuario(dbConn *sql.DB, tipoEmpresaID int64, nombre, descripcion, usuarioCreador string) (int64, error) {
	if err := RolesDeUsuarioSchemaReady(dbConn); err != nil {
		return 0, err
	}
	nombre = strings.TrimSpace(nombre)
	descripcion = strings.TrimSpace(descripcion)
	usuarioCreador = strings.TrimSpace(usuarioCreador)
	if tipoEmpresaID <= 0 || nombre == "" {
		return 0, errors.New("tipo_empresa_id y nombre son obligatorios")
	}
	if exists, err := roleNameExistsForTipo(dbConn, tipoEmpresaID, nombre, 0); err != nil {
		return 0, err
	} else if exists {
		return 0, errors.New("ya existe un rol con ese nombre para el tipo de empresa")
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO roles_de_usuario (
		tipo_empresa_id, nombre, descripcion, usuario_creador, estado, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, 'activo', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, tipoEmpresaID, nombre, descripcion, usuarioCreador)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpsertRolDeUsuarioByTipoNombre crea o reactiva un rol por tipo de empresa y nombre.
func UpsertRolDeUsuarioByTipoNombre(dbConn *sql.DB, tipoEmpresaID int64, nombre, descripcion, usuarioCreador string) (int64, bool, error) {
	if err := RolesDeUsuarioSchemaReady(dbConn); err != nil {
		return 0, false, err
	}
	nombre = strings.TrimSpace(nombre)
	descripcion = strings.TrimSpace(descripcion)
	usuarioCreador = strings.TrimSpace(usuarioCreador)
	if usuarioCreador == "" {
		usuarioCreador = "sistema.roles"
	}
	if tipoEmpresaID <= 0 || nombre == "" {
		return 0, false, errors.New("tipo_empresa_id y nombre son obligatorios")
	}
	var id int64
	err := queryRowSQLCompat(dbConn, `SELECT id
		FROM roles_de_usuario
		WHERE tipo_empresa_id = ? AND lower(trim(nombre)) = lower(trim(?))
		ORDER BY id DESC
		LIMIT 1`, tipoEmpresaID, nombre).Scan(&id)
	if err == nil {
		_, err = execSQLCompat(dbConn, `UPDATE roles_de_usuario
			SET descripcion = COALESCE(NULLIF(?, ''), descripcion),
				estado = 'activo',
				usuario_creador = COALESCE(NULLIF(?, ''), usuario_creador),
				fecha_actualizacion = CURRENT_TIMESTAMP
			WHERE id = ?`, descripcion, usuarioCreador, id)
		return id, false, err
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, false, err
	}
	id, err = CreateRolDeUsuario(dbConn, tipoEmpresaID, nombre, descripcion, usuarioCreador)
	return id, true, err
}

// GetRolesDeUsuario obtiene roles de usuario, con filtro opcional por tipo de empresa.
func GetRolesDeUsuario(dbConn *sql.DB, tipoEmpresaID int64, incluirInactivos bool) ([]RolDeUsuario, error) {
	if err := RolesDeUsuarioSchemaReady(dbConn); err != nil {
		return nil, err
	}
	query := `SELECT
		r.id,
		COALESCE(r.empresa_id, 0),
		r.tipo_empresa_id,
		COALESCE(t.nombre, ''),
		r.nombre,
		COALESCE(r.descripcion, ''),
		COALESCE(r.origen, 'global'),
		COALESCE(r.rol_base_id, 0),
		COALESCE(r.fecha_creacion, ''),
		COALESCE(r.fecha_actualizacion, ''),
		COALESCE(r.usuario_creador, ''),
		COALESCE(r.estado, 'activo'),
		COALESCE(r.observaciones, '')
	FROM roles_de_usuario r
	LEFT JOIN tipos_de_empresas t ON t.id = r.tipo_empresa_id
	WHERE COALESCE(r.empresa_id, 0) = 0`
	args := make([]interface{}, 0)

	if tipoEmpresaID > 0 {
		query += ` AND r.tipo_empresa_id = ?`
		args = append(args, tipoEmpresaID)
	}
	if !incluirInactivos {
		query += ` AND COALESCE(r.estado, 'activo') = 'activo'`
	}
	query += ` ORDER BY r.id DESC`

	rows, err := querySQLCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]RolDeUsuario, 0)
	for rows.Next() {
		var item RolDeUsuario
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.TipoEmpresaID,
			&item.TipoEmpresaNombre,
			&item.Nombre,
			&item.Descripcion,
			&item.Origen,
			&item.RolBaseID,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, err
		}
		item.Personalizado = item.EmpresaID > 0 || strings.EqualFold(strings.TrimSpace(item.Origen), "empresa")
		item.Asignable = IsRolDeUsuarioAsignable(&item)
		out = append(out, item)
	}
	return out, rows.Err()
}

// GetRolesDeUsuarioCatalogoGlobal obtiene un catalogo unico de roles para asignacion
// empresarial. No filtra por tipo de empresa: deduplica por nombre normalizado para
// que todos los tipos vean los mismos roles sin repetir opciones en el selector.
func GetRolesDeUsuarioCatalogoGlobal(dbConn *sql.DB, incluirInactivos bool) ([]RolDeUsuario, error) {
	roles, err := GetRolesDeUsuario(dbConn, 0, incluirInactivos)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(roles, func(i, j int) bool {
		leftKey := normalizeRolCatalogKey(roles[i].Nombre)
		rightKey := normalizeRolCatalogKey(roles[j].Nombre)
		if leftKey != rightKey {
			return leftKey < rightKey
		}
		leftActive := !strings.EqualFold(strings.TrimSpace(roles[i].Estado), "inactivo")
		rightActive := !strings.EqualFold(strings.TrimSpace(roles[j].Estado), "inactivo")
		if leftActive != rightActive {
			return leftActive
		}
		leftPreferred := preferredRolCatalogDisplayRank(leftKey, roles[i].Nombre)
		rightPreferred := preferredRolCatalogDisplayRank(rightKey, roles[j].Nombre)
		if leftPreferred != rightPreferred {
			return leftPreferred < rightPreferred
		}
		return roles[i].ID < roles[j].ID
	})
	seen := map[string]bool{}
	out := make([]RolDeUsuario, 0, len(roles))
	for _, item := range roles {
		key := normalizeRolCatalogKey(item.Nombre)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		item.TipoEmpresaID = 0
		item.TipoEmpresaNombre = "Todos los tipos de empresa"
		item.EmpresaID = 0
		item.Origen = "global"
		item.Personalizado = false
		out = append(out, item)
	}
	return out, nil
}

// GetRolesDeUsuarioCatalogoEmpresa obtiene roles globales mas roles propios de una empresa.
func GetRolesDeUsuarioCatalogoEmpresa(dbConn *sql.DB, empresaID int64, incluirInactivos bool) ([]RolDeUsuario, error) {
	if empresaID <= 0 {
		return nil, errors.New("empresa_id es obligatorio")
	}
	globales, err := GetRolesDeUsuarioCatalogoGlobal(dbConn, incluirInactivos)
	if err != nil {
		return nil, err
	}
	personalizados, err := GetRolesDeUsuarioEmpresa(dbConn, empresaID, incluirInactivos)
	if err != nil {
		return nil, err
	}
	out := make([]RolDeUsuario, 0, len(globales)+len(personalizados))
	bases := make(map[int64]RolDeUsuario, len(globales))
	if len(personalizados) > 0 {
		allBases, err := GetRolesDeUsuario(dbConn, 0, true)
		if err != nil {
			return nil, err
		}
		for _, base := range allBases {
			bases[base.ID] = base
		}
	}
	for _, item := range globales {
		if IsRolDeUsuarioAsignable(&RolDeUsuario{Nombre: item.Nombre, Estado: "activo"}) {
			out = append(out, item)
		}
	}
	for _, item := range personalizados {
		base, exists := bases[item.RolBaseID]
		item.Asignable = IsRolDeUsuarioAsignable(&item) && exists && IsRolDeUsuarioAsignable(&base)
		out = append(out, item)
	}
	return out, nil
}

// GetRolesDeUsuarioEmpresa lista los roles personalizados de una empresa.
func GetRolesDeUsuarioEmpresa(dbConn *sql.DB, empresaID int64, incluirInactivos bool) ([]RolDeUsuario, error) {
	if err := RolesDeUsuarioSchemaReady(dbConn); err != nil {
		return nil, err
	}
	if empresaID <= 0 {
		return nil, errors.New("empresa_id es obligatorio")
	}
	query := `SELECT
		r.id,
		COALESCE(r.empresa_id, 0),
		COALESCE(r.tipo_empresa_id, 0),
		'Rol personalizado de esta empresa',
		COALESCE(r.nombre, ''),
		COALESCE(r.descripcion, ''),
		COALESCE(r.origen, 'empresa'),
		COALESCE(r.rol_base_id, 0),
		COALESCE(r.fecha_creacion, ''),
		COALESCE(r.fecha_actualizacion, ''),
		COALESCE(r.usuario_creador, ''),
		COALESCE(r.estado, 'activo'),
		COALESCE(r.observaciones, '')
	FROM roles_de_usuario r
	WHERE COALESCE(r.empresa_id, 0) = ?`
	args := []interface{}{empresaID}
	if !incluirInactivos {
		query += ` AND COALESCE(r.estado, 'activo') = 'activo'`
	}
	query += ` ORDER BY lower(trim(r.nombre)) ASC, r.id ASC`
	rows, err := querySQLCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]RolDeUsuario, 0)
	for rows.Next() {
		var item RolDeUsuario
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.TipoEmpresaID,
			&item.TipoEmpresaNombre,
			&item.Nombre,
			&item.Descripcion,
			&item.Origen,
			&item.RolBaseID,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, err
		}
		item.Origen = "empresa"
		item.Personalizado = true
		item.Asignable = IsRolDeUsuarioAsignable(&item)
		out = append(out, item)
	}
	return out, rows.Err()
}

// CreateEmpresaRolDeUsuario crea un rol personalizado para una empresa.
func CreateEmpresaRolDeUsuario(dbConn *sql.DB, empresaID int64, nombre, descripcion string, rolBaseID int64, usuarioCreador string) (int64, error) {
	if err := RolesDeUsuarioSchemaReady(dbConn); err != nil {
		return 0, err
	}
	empresaID = normalizePositiveInt64(empresaID)
	nombre = strings.TrimSpace(nombre)
	descripcion = strings.TrimSpace(descripcion)
	usuarioCreador = strings.TrimSpace(usuarioCreador)
	if empresaID <= 0 || nombre == "" || rolBaseID <= 0 {
		return 0, errors.New("empresa_id, nombre y rol_base_id son obligatorios")
	}
	base, err := GetRolDeUsuarioByIDEmpresaScope(dbConn, empresaID, rolBaseID)
	if err != nil {
		return 0, err
	}
	if base.EmpresaID != 0 || !IsRolDeUsuarioAsignable(base) {
		return 0, errors.New("el rol base debe ser global, activo y empresarial")
	}
	if !IsRolDeUsuarioAsignable(&RolDeUsuario{Nombre: nombre, Estado: "activo"}) {
		return 0, errors.New("el nombre esta reservado para un rol de plataforma")
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if err := lockEmpresaRoleCatalog(tx, empresaID, nombre, 0); err != nil {
		return 0, err
	}
	var id int64
	err = tx.QueryRow(`INSERT INTO roles_de_usuario (
		tipo_empresa_id, empresa_id, nombre, descripcion, origen, rol_base_id, usuario_creador, estado, fecha_creacion, fecha_actualizacion
	) VALUES (0, ?, ?, ?, 'empresa', ?, ?, 'activo', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING id`,
		empresaID, nombre, descripcion, rolBaseID, usuarioCreador).Scan(&id)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateEmpresaRolDeUsuario actualiza un rol personalizado de una empresa.
func UpdateEmpresaRolDeUsuario(dbConn *sql.DB, empresaID, rolID int64, nombre, descripcion string, rolBaseID int64) error {
	if err := RolesDeUsuarioSchemaReady(dbConn); err != nil {
		return err
	}
	nombre = strings.TrimSpace(nombre)
	descripcion = strings.TrimSpace(descripcion)
	if empresaID <= 0 || rolID <= 0 || nombre == "" || rolBaseID <= 0 {
		return errors.New("empresa_id, rol_id, nombre y rol_base_id son obligatorios")
	}
	if _, err := GetRolDeUsuarioEmpresaByID(dbConn, empresaID, rolID); err != nil {
		return err
	}
	base, err := GetRolDeUsuarioByIDEmpresaScope(dbConn, empresaID, rolBaseID)
	if err != nil {
		return err
	}
	if base.EmpresaID != 0 || !IsRolDeUsuarioAsignable(base) {
		return errors.New("el rol base debe ser global, activo y empresarial")
	}
	if !IsRolDeUsuarioAsignable(&RolDeUsuario{Nombre: nombre, Estado: "activo"}) {
		return errors.New("el nombre esta reservado para un rol de plataforma")
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := lockEmpresaRoleCatalog(tx, empresaID, nombre, rolID); err != nil {
		return err
	}
	res, err := tx.Exec(`UPDATE roles_de_usuario
		SET nombre = ?, descripcion = ?, rol_base_id = ?, fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE id = ? AND COALESCE(empresa_id, 0) = ?`, nombre, descripcion, rolBaseID, rolID, empresaID)
	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); err == nil && affected == 0 {
		return sql.ErrNoRows
	}
	return tx.Commit()
}

// La exclusión por empresa evita duplicados por reintentos concurrentes sin
// bloquear catálogos de otros tenants ni modificar migraciones históricas.
func lockEmpresaRoleCatalog(tx *sql.Tx, empresaID int64, nombre string, excludeID int64) error {
	if _, err := tx.Exec(`SELECT pg_advisory_xact_lock(hashtextextended(?, 0))`, "pcs_roles_empresa:"+strconv.FormatInt(empresaID, 10)); err != nil {
		return err
	}
	var exists bool
	if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM roles_de_usuario WHERE empresa_id = ? AND lower(trim(nombre)) = lower(trim(?)) AND id <> ?)`, empresaID, nombre, excludeID).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return errors.New("ya existe un rol personalizado con ese nombre en esta empresa")
	}
	return nil
}

// SetEmpresaRolDeUsuarioEstado activa/desactiva un rol personalizado de una empresa.
func SetEmpresaRolDeUsuarioEstado(dbConn *sql.DB, empresaID, rolID int64, estado string) error {
	if empresaID <= 0 || rolID <= 0 {
		return sql.ErrNoRows
	}
	if err := RolesDeUsuarioSchemaReady(dbConn); err != nil {
		return err
	}
	estado = strings.ToLower(strings.TrimSpace(estado))
	if estado != "activo" && estado != "inactivo" {
		return errors.New("estado de rol invalido")
	}
	res, err := execSQLCompat(dbConn, `UPDATE roles_de_usuario
		SET estado = ?, fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE id = ? AND COALESCE(empresa_id, 0) = ?`, estado, rolID, empresaID)
	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); err == nil && affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetRolDeUsuarioEmpresaByID valida que un rol pertenezca a la empresa.
func GetRolDeUsuarioEmpresaByID(dbConn *sql.DB, empresaID, rolID int64) (*RolDeUsuario, error) {
	if empresaID <= 0 || rolID <= 0 {
		return nil, sql.ErrNoRows
	}
	rol, err := getRolDeUsuarioByIDScoped(dbConn, rolID, empresaID, true)
	if err != nil {
		return nil, err
	}
	if rol.EmpresaID != empresaID || empresaID <= 0 {
		return nil, sql.ErrNoRows
	}
	rol.Origen = "empresa"
	rol.Personalizado = true
	return rol, nil
}

// GetRolDeUsuarioByIDEmpresaScope permite roles globales o roles propios de la empresa.
func GetRolDeUsuarioByIDEmpresaScope(dbConn *sql.DB, empresaID, rolID int64) (*RolDeUsuario, error) {
	if empresaID <= 0 || rolID <= 0 {
		return nil, sql.ErrNoRows
	}
	rol, err := getRolDeUsuarioByIDScoped(dbConn, rolID, empresaID, false)
	if err != nil {
		return nil, err
	}
	if rol.EmpresaID > 0 && rol.EmpresaID != empresaID {
		return nil, sql.ErrNoRows
	}
	if rol.EmpresaID > 0 {
		rol.Origen = "empresa"
		rol.Personalizado = true
	} else {
		rol.Origen = "global"
	}
	return rol, nil
}

func roleNameExistsForEmpresa(dbConn *sql.DB, empresaID int64, nombre string, excludeID int64) (bool, error) {
	nombre = strings.TrimSpace(nombre)
	if empresaID <= 0 || nombre == "" {
		return false, nil
	}
	query := `SELECT COUNT(1) FROM roles_de_usuario WHERE COALESCE(empresa_id, 0) = ? AND lower(trim(nombre)) = lower(trim(?))`
	args := []interface{}{empresaID, nombre}
	if excludeID > 0 {
		query += ` AND id <> ?`
		args = append(args, excludeID)
	}
	var count int
	if err := queryRowSQLCompat(dbConn, query, args...).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func normalizePositiveInt64(value int64) int64 {
	if value < 0 {
		return 0
	}
	return value
}

func normalizeRolCatalogKey(nombre string) string {
	value := strings.ToLower(strings.TrimSpace(nombre))
	replacer := strings.NewReplacer(
		"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ñ", "n",
		"Á", "a", "É", "e", "Í", "i", "Ó", "o", "Ú", "u", "Ñ", "n",
		"_", " ", "-", " ", "/", " ", ".", " ",
	)
	value = replacer.Replace(value)
	key := strings.Join(strings.Fields(value), "_")
	aliases := map[string]string{
		"administrador":            "admin_empresa",
		"administrador_empresa":    "admin_empresa",
		"administrador_de_empresa": "admin_empresa",
		"admin":                    "admin_empresa",
		"supervisor":               "supervisor_sucursal",
		"caja":                     "cajero",
		"caja_turno":               "cajero",
		"caja_principal":           "cajero",
		"caja_hotel":               "cajero",
		"caja_bar":                 "cajero",
		"caja_salon":               "cajero",
		"caja_restaurante":         "cajero",
		"caja_pyme":                "cajero",
		"servicio_de_limpieza":     "servicio_limpieza",
		"limpieza":                 "servicio_limpieza",
		"aseadora":                 "servicio_limpieza",
		"jefe_de_bodega":           "jefe_bodega",
		"bodega":                   "jefe_bodega",
		"bodeguero":                "jefe_bodega",
		"recursos_humanos":         "recursos_humanos",
		"talento_humano":           "recursos_humanos",
		"rrhh":                     "recursos_humanos",
		"tecnico":                  "tecnico_solar",
		"tecnico_solar":            "tecnico_solar",
		"dueno":                    "empresario",
		"dueño":                    "empresario",
		"propietario":              "empresario",
	}
	if alias, ok := aliases[key]; ok {
		return alias
	}
	return key
}

func preferredRolCatalogDisplayRank(key, nombre string) int {
	if key == "cajero" && normalizeRolCatalogKey(nombre) == "cajero" && strings.EqualFold(strings.TrimSpace(nombre), "cajero") {
		return 0
	}
	return 1
}

// UpdateRolDeUsuario actualiza un rol de usuario.
func UpdateRolDeUsuario(dbConn *sql.DB, id, tipoEmpresaID int64, nombre, descripcion string) error {
	if err := RolesDeUsuarioSchemaReady(dbConn); err != nil {
		return err
	}
	nombre = strings.TrimSpace(nombre)
	descripcion = strings.TrimSpace(descripcion)
	if id <= 0 || tipoEmpresaID <= 0 || nombre == "" {
		return errors.New("id, tipo_empresa_id y nombre son obligatorios")
	}
	if exists, err := roleNameExistsForTipo(dbConn, tipoEmpresaID, nombre, id); err != nil {
		return err
	} else if exists {
		return errors.New("ya existe un rol con ese nombre para el tipo de empresa")
	}
	_, err := execSQLCompat(dbConn, `UPDATE roles_de_usuario
		SET tipo_empresa_id = ?, nombre = ?, descripcion = ?, fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE id = ?`, tipoEmpresaID, nombre, descripcion, id)
	return err
}

func roleNameExistsForTipo(dbConn *sql.DB, tipoEmpresaID int64, nombre string, excludeID int64) (bool, error) {
	nombre = strings.TrimSpace(nombre)
	if tipoEmpresaID <= 0 || nombre == "" {
		return false, nil
	}
	query := `SELECT COUNT(1) FROM roles_de_usuario WHERE tipo_empresa_id = ? AND lower(trim(nombre)) = lower(trim(?))`
	args := []interface{}{tipoEmpresaID, nombre}
	if excludeID > 0 {
		query += ` AND id <> ?`
		args = append(args, excludeID)
	}
	var count int
	if err := queryRowSQLCompat(dbConn, query, args...).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

// DeleteRolDeUsuario elimina un rol de usuario.
func DeleteRolDeUsuario(dbConn *sql.DB, id int64) error {
	if err := RolesDeUsuarioSchemaReady(dbConn); err != nil {
		return err
	}
	_, err := execSQLCompat(dbConn, `DELETE FROM roles_de_usuario WHERE id = ?`, id)
	return err
}

// SetRolDeUsuarioEstado activa/desactiva un rol de usuario.
func SetRolDeUsuarioEstado(dbConn *sql.DB, id int64, estado string) error {
	if err := RolesDeUsuarioSchemaReady(dbConn); err != nil {
		return err
	}
	_, err := execSQLCompat(dbConn, `UPDATE roles_de_usuario SET estado = ?, fecha_actualizacion = CURRENT_TIMESTAMP WHERE id = ?`, estado, id)
	return err
}

// DropTiposDeUsuarioTable elimina la tabla legada `tipos_de_usuario` (modulo retirado del producto).
func DropTiposDeUsuarioTable(dbConn *sql.DB) error {
	if dbConn == nil {
		return nil
	}
	_, err := execSQLCompat(dbConn, `DROP TABLE IF EXISTS tipos_de_usuario`)
	return err
}
