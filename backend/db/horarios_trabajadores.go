package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// HorarioTrabajador define la estructura del horario asignado a un empleado/usuario.
type HorarioTrabajador struct {
	ID             int64              `json:"id"`
	EmpresaID      int64              `json:"empresa_id"`
	UsuarioID      *int64             `json:"usuario_id"` // Puede no ser un usuario, sino un nombre suelto
	NombreEmpleado string             `json:"nombre_empleado"`
	Fecha          string             `json:"fecha"`       // YYYY-MM-DD
	HoraInicio     string             `json:"hora_inicio"` // HH:MM
	HoraFin        string             `json:"hora_fin"`    // HH:MM
	Estado         string             `json:"estado"`      // 'agendado', 'completado', 'ausente', 'cancelado'
	Observaciones  *string            `json:"observaciones"`
	FechaCreacion  string             `json:"fecha_creacion"`
	UsuarioCreador string             `json:"usuario_creador"`
}

// EnsureHorariosTrabajadoresSchema crea la tabla para el calendario de turnos.
func EnsureHorariosTrabajadoresSchema() error {
	dbConn := GetDB()
	if dbConn == nil {
		return fmt.Errorf("base de datos de empresas no inicializada")
	}

	schema := `CREATE TABLE IF NOT EXISTS empresa_horarios_trabajadores (
		id BIGSERIAL PRIMARY KEY,
		empresa_id BIGINT NOT NULL,
		usuario_id BIGINT,
		nombre_empleado TEXT NOT NULL,
		fecha TEXT NOT NULL,
		hora_inicio TEXT NOT NULL,
		hora_fin TEXT NOT NULL,
		estado TEXT DEFAULT 'agendado',
		observaciones TEXT,
		fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
		usuario_creador TEXT
	);`

	if _, err := dbConn.Exec(schema); err != nil {
		return fmt.Errorf("error creando tabla empresa_horarios_trabajadores: %v", err)
	}

	// Índices
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_empresa_horarios_trabajadores_empresa ON empresa_horarios_trabajadores(empresa_id);`,
		`CREATE INDEX IF NOT EXISTS idx_empresa_horarios_trabajadores_fecha ON empresa_horarios_trabajadores(empresa_id, fecha);`,
	}
	for _, idx := range indexes {
		if _, err := dbConn.Exec(idx); err != nil {
			log.Printf("[Advertencia] No se pudo crear el índice '%s': %v", idx, err)
		}
	}

	return nil
}

// GetHorariosTrabajadoresByrango returns schedules within a date range for a specific company.
func GetHorariosTrabajadoresByRango(empresaID int64, fechaInicio, fechaFin string) ([]HorarioTrabajador, error) {
	dbConn := GetDB()
	if dbConn == nil {
		return nil, fmt.Errorf("base de datos no inicializada")
	}

	query := `
		SELECT id, empresa_id, usuario_id, nombre_empleado, fecha, hora_inicio, hora_fin, estado, observaciones, fecha_creacion, usuario_creador
		FROM empresa_horarios_trabajadores
		WHERE empresa_id = $1 AND fecha >= $2 AND fecha <= $3
		ORDER BY fecha ASC, hora_inicio ASC
	`
	rows, err := ExecQueryCompat(dbConn, query, empresaID, fechaInicio, fechaFin)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []HorarioTrabajador
	for rows.Next() {
		var h HorarioTrabajador
		var uid sql.NullInt64
		var obs sql.NullString
		if err := rows.Scan(
			&h.ID, &h.EmpresaID, &uid, &h.NombreEmpleado, &h.Fecha, &h.HoraInicio, &h.HoraFin,
			&h.Estado, &obs, &h.FechaCreacion, &h.UsuarioCreador,
		); err != nil {
			return nil, err
		}
		if uid.Valid {
			h.UsuarioID = &uid.Int64
		}
		if obs.Valid {
			h.Observaciones = &obs.String
		}
		result = append(result, h)
	}
	return result, nil
}

// GetHorariosTrabajadorByUsuario returns recent or range schedules for a specific worker.
func GetHorariosTrabajadorByUsuario(empresaID int64, usuarioID int64) ([]HorarioTrabajador, error) {
	dbConn := GetDB()
	if dbConn == nil {
		return nil, fmt.Errorf("base de datos no inicializada")
	}

	query := `
		SELECT id, empresa_id, usuario_id, nombre_empleado, fecha, hora_inicio, hora_fin, estado, observaciones, fecha_creacion, usuario_creador
		FROM empresa_horarios_trabajadores
		WHERE empresa_id = $1 AND usuario_id = $2
		ORDER BY fecha DESC, hora_inicio DESC
		LIMIT 100
	`
	rows, err := ExecQueryCompat(dbConn, query, empresaID, usuarioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []HorarioTrabajador
	for rows.Next() {
		var h HorarioTrabajador
		var uid sql.NullInt64
		var obs sql.NullString
		if err := rows.Scan(
			&h.ID, &h.EmpresaID, &uid, &h.NombreEmpleado, &h.Fecha, &h.HoraInicio, &h.HoraFin,
			&h.Estado, &obs, &h.FechaCreacion, &h.UsuarioCreador,
		); err != nil {
			return nil, err
		}
		if uid.Valid {
			h.UsuarioID = &uid.Int64
		}
		if obs.Valid {
			h.Observaciones = &obs.String
		}
		result = append(result, h)
	}
	return result, nil
}

// CreateHorarioTrabajador inserts a new schedule into the DB.
func CreateHorarioTrabajador(h *HorarioTrabajador) error {
	dbConn := GetDB()
	if dbConn == nil {
		return fmt.Errorf("db no inicializada")
	}

	query := `
		INSERT INTO empresa_horarios_trabajadores 
			(empresa_id, usuario_id, nombre_empleado, fecha, hora_inicio, hora_fin, estado, observaciones, usuario_creador, fecha_creacion)
		VALUES 
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	fechaStr := time.Now().Format("2006-01-02 15:04:05")

	_, err := ExecCompat(dbConn, query,
		h.EmpresaID,
		h.UsuarioID,
		h.NombreEmpleado,
		h.Fecha,
		h.HoraInicio,
		h.HoraFin,
		h.Estado,
		h.Observaciones,
		h.UsuarioCreador,
		fechaStr,
	)
	return err
}

// UpdateHorarioTrabajador modifies schedule info
func UpdateHorarioTrabajador(h *HorarioTrabajador) error {
	dbConn := GetDB()
	if dbConn == nil {
		return fmt.Errorf("db no inicializada")
	}

	query := `
		UPDATE empresa_horarios_trabajadores
		SET usuario_id = $1, nombre_empleado = $2, fecha = $3, hora_inicio = $4, hora_fin = $5, estado = $6, observaciones = $7
		WHERE id = $8 AND empresa_id = $9
	`
	res, err := ExecCompat(dbConn, query,
		h.UsuarioID,
		h.NombreEmpleado,
		h.Fecha,
		h.HoraInicio,
		h.HoraFin,
		h.Estado,
		h.Observaciones,
		h.ID,
		h.EmpresaID,
	)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no se encontró el horario o no hubo cambios")
	}

	return nil
}

// DeleteHorarioTrabajador
func DeleteHorarioTrabajador(id int64, empresaID int64) error {
	dbConn := GetDB()
	if dbConn == nil {
		return fmt.Errorf("db no inicializada")
	}

	query := `DELETE FROM empresa_horarios_trabajadores WHERE id = $1 AND empresa_id = $2`
	_, err := ExecCompat(dbConn, query, id, empresaID)
	return err
}
