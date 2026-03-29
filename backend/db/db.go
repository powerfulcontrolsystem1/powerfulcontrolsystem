package db

import (
	"database/sql"
)

// UpsertUser inserta o actualiza un usuario en la base de datos de empresas (registro por empresa)
func UpsertUser(dbConn *sql.DB, email, name string) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("INSERT OR IGNORE INTO users (email, name) VALUES (?, ?)", email, name); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE users SET name = ? WHERE email = ?", name, email); err != nil {
		return err
	}
	return tx.Commit()
}

// EnsureUserEmpresa crea una empresa por defecto para el usuario si no tiene una asociada
func EnsureUserEmpresa(dbConn *sql.DB, email, empresaNombre string) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var userID int64
	var empresaID sql.NullInt64
	row := tx.QueryRow("SELECT id, empresa_id FROM users WHERE email = ?", email)
	if err := row.Scan(&userID, &empresaID); err != nil {
		return err
	}

	if empresaID.Valid {
		// ya tiene empresa asociada
		return tx.Commit()
	}

	res, err := tx.Exec("INSERT INTO empresas (nombre, usuario_creador) VALUES (?, ?)", empresaNombre, email)
	if err != nil {
		return err
	}
	newEmpresaID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	if _, err := tx.Exec("UPDATE users SET empresa_id = ? WHERE id = ?", newEmpresaID, userID); err != nil {
		return err
	}

	return tx.Commit()
}

// UpsertAdministrador inserta o actualiza un registro en la tabla administradores de la base superadministrador
// Si se inserta por primera vez, asigna el rol provisto (usualmente 'administrador').
// Ahora acepta un campo `photo` con la URL de la foto del perfil.
func UpsertAdministrador(dbConn *sql.DB, email, name, role, photo string) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("INSERT OR IGNORE INTO administradores (email, name, role, photo, fecha_creacion, fecha_actualizacion, estado) VALUES (?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'), 'activo')", email, name, role, photo); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE administradores SET name = ?, role = ?, photo = ?, fecha_actualizacion = datetime('now','localtime') WHERE email = ?", name, role, photo, email); err != nil {
		return err
	}

	return tx.Commit()
}

// UpdateAdministrador actualiza el nombre y rol de un administrador por id
func UpdateAdministrador(dbConn *sql.DB, id int64, name, role string) error {
	_, err := dbConn.Exec("UPDATE administradores SET name = ?, role = ?, fecha_actualizacion = datetime('now','localtime') WHERE id = ?", name, role, id)
	return err
}

// DeleteAdministrador elimina un administrador por id
func DeleteAdministrador(dbConn *sql.DB, id int64) error {
	_, err := dbConn.Exec("DELETE FROM administradores WHERE id = ?", id)
	return err
}

// SetAdministradorEstado activa/desactiva un administrador (estado: 'activo'/'inactivo')
func SetAdministradorEstado(dbConn *sql.DB, id int64, estado string) error {
	_, err := dbConn.Exec("UPDATE administradores SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE id = ?", estado, id)
	return err
}

// CreateSession registra una sesión en la tabla sesiones de superadministrador
func CreateSession(dbConn *sql.DB, adminEmail, ip, userAgent, token string) error {
	_, err := dbConn.Exec("INSERT INTO sesiones (admin_email, token, ip, user_agent, fecha_inicio, activo, fecha_creacion) VALUES (?, ?, ?, ?, datetime('now','localtime'), 1, datetime('now','localtime'))", adminEmail, token, ip, userAgent)
	return err
}

// Admin representa un registro en la tabla administradores
type Admin struct {
	ID                 int64  `json:"id"`
	Email              string `json:"email"`
	Name               string `json:"name"`
	Role               string `json:"role"`
	Photo              string `json:"photo,omitempty"`
	FechaCreacion      string `json:"fecha_creacion"`
	FechaActualizacion string `json:"fecha_actualizacion"`
	Estado             string `json:"estado"`
}

// NOTE: tipos_de_licencia CRUD removed per project decision (frontend/page/link removed).

// Licencia representa una licencia asignada (nuevo CRUD)
type Licencia struct {
	ID            int64   `json:"id"`
	EmpresaID     int64   `json:"empresa_id"`
	TipoID        int64   `json:"tipo_id"`
	TipoNombre    string  `json:"tipo_nombre,omitempty"`
	Nombre        string  `json:"nombre"`
	Descripcion   string  `json:"descripcion"`
	Valor         float64 `json:"valor"`
	DuracionDias  int     `json:"duracion_dias"`
	FechaCreacion string  `json:"fecha_creacion"`
	Activo        int     `json:"activo"`
}

// CreateLicencia inserta una nueva licencia en dbSuper
func CreateLicencia(dbConn *sql.DB, tipoID int64, nombre, descripcion string, valor float64, duracionDias int) (int64, error) {
	res, err := dbConn.Exec("INSERT INTO licencias (tipo_id, nombre, descripcion, valor, duracion_dias, fecha_creacion, activo) VALUES (?, ?, ?, ?, ?, datetime('now','localtime'), 1)", tipoID, nombre, descripcion, valor, duracionDias)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetLicencias obtiene todas las licencias (con nombre de tipo si existe)
func GetLicencias(dbConn *sql.DB) ([]Licencia, error) {
	q := `SELECT l.id, l.empresa_id, l.tipo_id, t.nombre, l.nombre, l.descripcion, l.valor, l.duracion_dias, l.fecha_creacion, l.activo
		FROM licencias l LEFT JOIN tipos_de_empresas t ON l.tipo_id = t.id
		ORDER BY l.id DESC`
	rows, err := dbConn.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Licencia
	for rows.Next() {
		var lic Licencia
		var empresaID sql.NullInt64
		var tipoNombre sql.NullString
		var descripcion sql.NullString
		var fechaCreacion sql.NullString
		if err := rows.Scan(&lic.ID, &empresaID, &lic.TipoID, &tipoNombre, &lic.Nombre, &descripcion, &lic.Valor, &lic.DuracionDias, &fechaCreacion, &lic.Activo); err != nil {
			return nil, err
		}
		if empresaID.Valid {
			lic.EmpresaID = empresaID.Int64
		}
		if tipoNombre.Valid {
			lic.TipoNombre = tipoNombre.String
		}
		if descripcion.Valid {
			lic.Descripcion = descripcion.String
		}
		if fechaCreacion.Valid {
			lic.FechaCreacion = fechaCreacion.String
		}
		out = append(out, lic)
	}
	return out, nil
}

// GetLicenciaByID devuelve una licencia por id
func GetLicenciaByID(dbConn *sql.DB, id int64) (*Licencia, error) {
	q := `SELECT id, empresa_id, tipo_id, nombre, descripcion, valor, duracion_dias, fecha_creacion, activo FROM licencias WHERE id = ? LIMIT 1`
	row := dbConn.QueryRow(q, id)
	var lic Licencia
	var empresaID sql.NullInt64
	var descripcion sql.NullString
	var fechaCreacion sql.NullString
	if err := row.Scan(&lic.ID, &empresaID, &lic.TipoID, &lic.Nombre, &descripcion, &lic.Valor, &lic.DuracionDias, &fechaCreacion, &lic.Activo); err != nil {
		return nil, err
	}
	if empresaID.Valid {
		lic.EmpresaID = empresaID.Int64
	}
	if descripcion.Valid {
		lic.Descripcion = descripcion.String
	}
	if fechaCreacion.Valid {
		lic.FechaCreacion = fechaCreacion.String
	}
	return &lic, nil
}

// UpdateLicencia actualiza campos editables de una licencia
func UpdateLicencia(dbConn *sql.DB, id, tipoID int64, nombre, descripcion string, valor float64, duracionDias int) error {
	_, err := dbConn.Exec("UPDATE licencias SET tipo_id = ?, nombre = ?, descripcion = ?, valor = ?, duracion_dias = ?, fecha_actualizacion = datetime('now','localtime') WHERE id = ?", tipoID, nombre, descripcion, valor, duracionDias, id)
	return err
}

// DeleteLicencia elimina una licencia por id
func DeleteLicencia(dbConn *sql.DB, id int64) error {
	_, err := dbConn.Exec("DELETE FROM licencias WHERE id = ?", id)
	return err
}

// SetLicenciaActivo activa/desactiva una licencia (activo: 1 o 0)
func SetLicenciaActivo(dbConn *sql.DB, id int64, activo int) error {
	_, err := dbConn.Exec("UPDATE licencias SET activo = ?, fecha_actualizacion = datetime('now','localtime') WHERE id = ?", activo, id)
	return err
}

// Session representa una sesión del administrador registrada en la tabla sesiones
type Session struct {
	ID            int64  `json:"id"`
	AdminEmail    string `json:"admin_email"`
	Token         string `json:"token"`
	IP            string `json:"ip"`
	UserAgent     string `json:"user_agent"`
	FechaInicio   string `json:"fecha_inicio"`
	FechaCreacion string `json:"fecha_creacion"`
	Activo        int    `json:"activo"`
}

// GetSessionByToken devuelve una sesión activa por token
func GetSessionByToken(dbConn *sql.DB, token string) (*Session, error) {
	row := dbConn.QueryRow("SELECT id, admin_email, token, ip, user_agent, fecha_inicio, fecha_creacion, activo FROM sesiones WHERE token = ? AND activo = 1 LIMIT 1", token)
	var s Session
	var fechaInicio sql.NullString
	var fechaCreacion sql.NullString
	if err := row.Scan(&s.ID, &s.AdminEmail, &s.Token, &s.IP, &s.UserAgent, &fechaInicio, &fechaCreacion, &s.Activo); err != nil {
		return nil, err
	}
	if fechaInicio.Valid {
		s.FechaInicio = fechaInicio.String
	}
	if fechaCreacion.Valid {
		s.FechaCreacion = fechaCreacion.String
	}
	return &s, nil
}

// GetAdminByEmail devuelve el administrador por email
func GetAdminByEmail(dbConn *sql.DB, email string) (*Admin, error) {
	row := dbConn.QueryRow("SELECT id, email, name, role, photo, fecha_creacion, fecha_actualizacion, estado FROM administradores WHERE email = ? LIMIT 1", email)
	var a Admin
	var photo sql.NullString
	if err := row.Scan(&a.ID, &a.Email, &a.Name, &a.Role, &photo, &a.FechaCreacion, &a.FechaActualizacion, &a.Estado); err != nil {
		return nil, err
	}
	if photo.Valid { a.Photo = photo.String }
	return &a, nil
}

// GetAdministradores lista todos los administradores
func GetAdministradores(dbConn *sql.DB) ([]Admin, error) {
	rows, err := dbConn.Query("SELECT id, email, name, role, photo, fecha_creacion, fecha_actualizacion, estado FROM administradores ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Admin
	for rows.Next() {
		var a Admin
		var photo sql.NullString
		if err := rows.Scan(&a.ID, &a.Email, &a.Name, &a.Role, &photo, &a.FechaCreacion, &a.FechaActualizacion, &a.Estado); err != nil {
			return nil, err
		}
		if photo.Valid { a.Photo = photo.String }
		out = append(out, a)
	}
	return out, nil
}

// GetSesiones lista las sesiones registradas
func GetSesiones(dbConn *sql.DB) ([]Session, error) {
	rows, err := dbConn.Query("SELECT id, admin_email, token, ip, user_agent, fecha_inicio, fecha_creacion, activo FROM sesiones ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Session
	for rows.Next() {
		var s Session
		var fechaInicio sql.NullString
		var fechaCreacion sql.NullString
		if err := rows.Scan(&s.ID, &s.AdminEmail, &s.Token, &s.IP, &s.UserAgent, &fechaInicio, &fechaCreacion, &s.Activo); err != nil {
			return nil, err
		}
		if fechaInicio.Valid {
			s.FechaInicio = fechaInicio.String
		}
		if fechaCreacion.Valid {
			s.FechaCreacion = fechaCreacion.String
		}
		out = append(out, s)
	}
	return out, nil
}

// TipoEmpresa representa un tipo de empresa
type TipoEmpresa struct {
	ID            int64  `json:"id"`
	Nombre        string `json:"nombre"`
	Observaciones string `json:"observaciones"`
	FechaCreacion string `json:"fecha_creacion"`
	Estado        string `json:"estado"`
}

// CreateTipoEmpresa inserta un nuevo tipo de empresa
func CreateTipoEmpresa(dbConn *sql.DB, nombre, observaciones string) (int64, error) {
	res, err := dbConn.Exec("INSERT INTO tipos_de_empresas (nombre, observaciones, fecha_creacion) VALUES (?, ?, datetime('now','localtime'))", nombre, observaciones)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetTiposEmpresas obtiene todos los tipos de empresa
func GetTiposEmpresas(dbConn *sql.DB) ([]TipoEmpresa, error) {
	rows, err := dbConn.Query("SELECT id, nombre, observaciones, fecha_creacion, estado FROM tipos_de_empresas ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TipoEmpresa
	for rows.Next() {
		var t TipoEmpresa
		if err := rows.Scan(&t.ID, &t.Nombre, &t.Observaciones, &t.FechaCreacion, &t.Estado); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

// UpdateTipoEmpresa actualiza un tipo de empresa por id
func UpdateTipoEmpresa(dbConn *sql.DB, id int64, nombre, observaciones string) error {
	_, err := dbConn.Exec("UPDATE tipos_de_empresas SET nombre = ?, observaciones = ?, fecha_actualizacion = datetime('now','localtime') WHERE id = ?", nombre, observaciones, id)
	return err
}

// DeleteTipoEmpresa elimina un tipo de empresa por id
func DeleteTipoEmpresa(dbConn *sql.DB, id int64) error {
	_, err := dbConn.Exec("DELETE FROM tipos_de_empresas WHERE id = ?", id)
	return err
}

// SetTipoEmpresaActivo activa/desactiva un tipo de empresa (activo: 'activo'/'inactivo' o 1/0)
func SetTipoEmpresaActivo(dbConn *sql.DB, id int64, estado string) error {
	_, err := dbConn.Exec("UPDATE tipos_de_empresas SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE id = ?", estado, id)
	return err
}

// Empresa representa una empresa registrada en empresas.db
type Empresa struct {
	ID                 int64  `json:"id"`
	Nombre             string `json:"nombre"`
	Nit                string `json:"nit,omitempty"`
	TipoID             int64  `json:"tipo_id,omitempty"`
	TipoNombre         string `json:"tipo_nombre,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// CreateEmpresa inserta una nueva empresa en la base empresas.db
func CreateEmpresa(dbConn *sql.DB, tipoID int64, tipoNombre, nombre, nit, observaciones, usuarioCreador string) (int64, error) {
	res, err := dbConn.Exec("INSERT INTO empresas (tipo_id, tipo_nombre, nombre, nit, observaciones, usuario_creador, fecha_creacion, estado) VALUES (?, ?, ?, ?, ?, ?, datetime('now','localtime'), 'activo')", tipoID, tipoNombre, nombre, nit, observaciones, usuarioCreador)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetEmpresas obtiene todas las empresas
func GetEmpresas(dbConn *sql.DB) ([]Empresa, error) {
	rows, err := dbConn.Query("SELECT id, nombre, nit, tipo_id, tipo_nombre, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones FROM empresas ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Empresa
	for rows.Next() {
		var e Empresa
		var nit sql.NullString
		var tipoID sql.NullInt64
		var tipoNombre sql.NullString
		var fechaCre sql.NullString
		var fechaAct sql.NullString
		var usuario sql.NullString
		var estado sql.NullString
		var obs sql.NullString
		if err := rows.Scan(&e.ID, &e.Nombre, &nit, &tipoID, &tipoNombre, &fechaCre, &fechaAct, &usuario, &estado, &obs); err != nil {
			return nil, err
		}
		if nit.Valid {
			e.Nit = nit.String
		}
		if tipoID.Valid {
			e.TipoID = tipoID.Int64
		}
		if tipoNombre.Valid {
			e.TipoNombre = tipoNombre.String
		}
		if fechaCre.Valid {
			e.FechaCreacion = fechaCre.String
		}
		if fechaAct.Valid {
			e.FechaActualizacion = fechaAct.String
		}
		if usuario.Valid {
			e.UsuarioCreador = usuario.String
		}
		if estado.Valid {
			e.Estado = estado.String
		}
		if obs.Valid {
			e.Observaciones = obs.String
		}
		out = append(out, e)
	}
	return out, nil
}

// GetEmpresaByID devuelve una empresa por id
func GetEmpresaByID(dbConn *sql.DB, id int64) (*Empresa, error) {
	row := dbConn.QueryRow("SELECT id, nombre, nit, tipo_id, tipo_nombre, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones FROM empresas WHERE id = ? LIMIT 1", id)
	var e Empresa
	var nit sql.NullString
	var tipoID sql.NullInt64
	var tipoNombre sql.NullString
	var fechaCre sql.NullString
	var fechaAct sql.NullString
	var usuario sql.NullString
	var estado sql.NullString
	var obs sql.NullString
	if err := row.Scan(&e.ID, &e.Nombre, &nit, &tipoID, &tipoNombre, &fechaCre, &fechaAct, &usuario, &estado, &obs); err != nil {
		return nil, err
	}
	if nit.Valid {
		e.Nit = nit.String
	}
	if tipoID.Valid {
		e.TipoID = tipoID.Int64
	}
	if tipoNombre.Valid {
		e.TipoNombre = tipoNombre.String
	}
	if fechaCre.Valid {
		e.FechaCreacion = fechaCre.String
	}
	if fechaAct.Valid {
		e.FechaActualizacion = fechaAct.String
	}
	if usuario.Valid {
		e.UsuarioCreador = usuario.String
	}
	if estado.Valid {
		e.Estado = estado.String
	}
	if obs.Valid {
		e.Observaciones = obs.String
	}
	return &e, nil
}

// UpdateEmpresa actualiza campos editables de una empresa
func UpdateEmpresa(dbConn *sql.DB, id, tipoID int64, tipoNombre, nombre, nit, observaciones string) error {
	_, err := dbConn.Exec("UPDATE empresas SET tipo_id = ?, tipo_nombre = ?, nombre = ?, nit = ?, observaciones = ?, fecha_actualizacion = datetime('now','localtime') WHERE id = ?", tipoID, tipoNombre, nombre, nit, observaciones, id)
	return err
}

// DeleteEmpresa elimina una empresa por id
func DeleteEmpresa(dbConn *sql.DB, id int64) error {
	_, err := dbConn.Exec("DELETE FROM empresas WHERE id = ?", id)
	return err
}

// SetEmpresaEstado activa/desactiva una empresa (estado: 'activo'/'inactivo')
func SetEmpresaEstado(dbConn *sql.DB, id int64, estado string) error {
	_, err := dbConn.Exec("UPDATE empresas SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE id = ?", estado, id)
	return err
}

// Metric representa una muestra de métricas del sistema
type Metric struct {
	ID            int64   `json:"id"`
	Timestamp     string  `json:"timestamp"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemTotal      uint64  `json:"mem_total"`
	MemUsed       uint64  `json:"mem_used"`
	MemPercent    float64 `json:"mem_percent"`
	NetRecv       uint64  `json:"net_recv"`
	NetSent       uint64  `json:"net_sent"`
	FechaCreacion string  `json:"fecha_creacion"`
}

// InitMetricsTable crea la tabla metrics en la base de datos si no existe
func InitMetricsTable(dbConn *sql.DB) error {
	create := `CREATE TABLE IF NOT EXISTS metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp TEXT DEFAULT (datetime('now','localtime')),
		cpu_percent REAL,
		mem_total INTEGER,
		mem_used INTEGER,
		mem_percent REAL,
		net_recv INTEGER,
		net_sent INTEGER,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT,
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`
	_, err := dbConn.Exec(create)
	return err
}

// InsertMetric inserta una muestra de métricas en la tabla metrics
func InsertMetric(dbConn *sql.DB, cpuPercent float64, memTotal, memUsed uint64, memPercent float64, netRecv, netSent uint64) error {
	_, err := dbConn.Exec("INSERT INTO metrics (cpu_percent, mem_total, mem_used, mem_percent, net_recv, net_sent) VALUES (?, ?, ?, ?, ?, ?)",
		cpuPercent, memTotal, memUsed, memPercent, netRecv, netSent)
	return err
}

// GetLatestMetric obtiene la última muestra registrada
func GetLatestMetric(dbConn *sql.DB) (*Metric, error) {
	row := dbConn.QueryRow("SELECT id, timestamp, cpu_percent, mem_total, mem_used, mem_percent, net_recv, net_sent, fecha_creacion FROM metrics ORDER BY id DESC LIMIT 1")
	var m Metric
	var timestamp sql.NullString
	var fechaCre sql.NullString
	if err := row.Scan(&m.ID, &timestamp, &m.CPUPercent, &m.MemTotal, &m.MemUsed, &m.MemPercent, &m.NetRecv, &m.NetSent, &fechaCre); err != nil {
		return nil, err
	}
	if timestamp.Valid {
		m.Timestamp = timestamp.String
	}
	if fechaCre.Valid {
		m.FechaCreacion = fechaCre.String
	}
	return &m, nil
}

// GetMetricsHistory devuelve las últimas 'limit' muestras (ordenadas de más antiguo a más reciente)
func GetMetricsHistory(dbConn *sql.DB, limit int) ([]Metric, error) {
	q := "SELECT id, timestamp, cpu_percent, mem_total, mem_used, mem_percent, net_recv, net_sent, fecha_creacion FROM metrics ORDER BY id DESC LIMIT ?"
	rows, err := dbConn.Query(q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Metric
	for rows.Next() {
		var m Metric
		var timestamp sql.NullString
		var fechaCre sql.NullString
		if err := rows.Scan(&m.ID, &timestamp, &m.CPUPercent, &m.MemTotal, &m.MemUsed, &m.MemPercent, &m.NetRecv, &m.NetSent, &fechaCre); err != nil {
			return nil, err
		}
		if timestamp.Valid {
			m.Timestamp = timestamp.String
		}
		if fechaCre.Valid {
			m.FechaCreacion = fechaCre.String
		}
		out = append(out, m)
	}
	// invertir slice para devolver de más antiguo a más reciente
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}
