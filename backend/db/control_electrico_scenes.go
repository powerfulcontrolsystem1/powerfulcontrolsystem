package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type EmpresaControlElectricoEscenaItem struct {
	ID             int64  `json:"id,omitempty"`
	EmpresaID      int64  `json:"empresa_id,omitempty"`
	EscenaID       int64  `json:"escena_id,omitempty"`
	ReleID         int64  `json:"rele_id"`
	EstadoObjetivo string `json:"estado_objetivo"`
	Orden          int    `json:"orden"`
}

type EmpresaControlElectricoEscena struct {
	ID                 int64                               `json:"id"`
	EmpresaID          int64                               `json:"empresa_id"`
	Nombre             string                              `json:"nombre"`
	Descripcion        string                              `json:"descripcion,omitempty"`
	FechaCreacion      string                              `json:"fecha_creacion,omitempty"`
	FechaActualizacion string                              `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string                              `json:"usuario_creador,omitempty"`
	Estado             string                              `json:"estado,omitempty"`
	Items              []EmpresaControlElectricoEscenaItem `json:"items"`
}

func ListEmpresaControlElectricoEscenas(dbConn *sql.DB, empresaID int64, includeInactive bool) ([]EmpresaControlElectricoEscena, error) {
	if dbConn == nil || empresaID <= 0 {
		return nil, errors.New("empresa_id invalido")
	}
	query := `SELECT id, empresa_id, nombre, COALESCE(descripcion,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,''), COALESCE(estado,'activo') FROM empresa_control_electrico_escenas WHERE empresa_id=?`
	if !includeInactive {
		query += ` AND LOWER(COALESCE(estado,'activo'))='activo'`
	}
	query += ` ORDER BY nombre,id`
	rows, err := querySQLCompat(dbConn, query, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaControlElectricoEscena{}
	for rows.Next() {
		var scene EmpresaControlElectricoEscena
		if err := rows.Scan(&scene.ID, &scene.EmpresaID, &scene.Nombre, &scene.Descripcion, &scene.FechaCreacion, &scene.FechaActualizacion, &scene.UsuarioCreador, &scene.Estado); err != nil {
			return nil, err
		}
		scene.Items = []EmpresaControlElectricoEscenaItem{}
		out = append(out, scene)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	index := map[int64]int{}
	for i := range out {
		index[out[i].ID] = i
	}
	itemRows, err := querySQLCompat(dbConn, `SELECT id, empresa_id, escena_id, rele_id, COALESCE(estado_objetivo,'off'), COALESCE(orden,0) FROM empresa_control_electrico_escena_items WHERE empresa_id=? ORDER BY escena_id,orden,id`, empresaID)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()
	for itemRows.Next() {
		var item EmpresaControlElectricoEscenaItem
		if err := itemRows.Scan(&item.ID, &item.EmpresaID, &item.EscenaID, &item.ReleID, &item.EstadoObjetivo, &item.Orden); err != nil {
			return nil, err
		}
		if pos, ok := index[item.EscenaID]; ok {
			out[pos].Items = append(out[pos].Items, item)
		}
	}
	return out, itemRows.Err()
}

func UpsertEmpresaControlElectricoEscena(dbConn *sql.DB, scene EmpresaControlElectricoEscena) (int64, error) {
	if dbConn == nil || scene.EmpresaID <= 0 {
		return 0, errors.New("empresa_id invalido")
	}
	scene.Nombre = strings.TrimSpace(scene.Nombre)
	if scene.Nombre == "" || len(scene.Nombre) > 120 {
		return 0, errors.New("nombre de escena invalido")
	}
	if len(scene.Items) == 0 || len(scene.Items) > 100 {
		return 0, errors.New("la escena debe contener entre 1 y 100 aparatos")
	}
	tx, err := dbConn.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	for index := range scene.Items {
		item := &scene.Items[index]
		item.EstadoObjetivo = strings.ToLower(strings.TrimSpace(item.EstadoObjetivo))
		if item.ReleID <= 0 || (item.EstadoObjetivo != "on" && item.EstadoObjetivo != "off") {
			return 0, errors.New("aparato o estado objetivo invalido")
		}
		var exists int
		if err := queryRowTxSQLCompat(tx, `SELECT COUNT(1) FROM empresa_control_electrico_reles WHERE empresa_id=? AND id=? AND LOWER(COALESCE(estado,'activo'))='activo'`, scene.EmpresaID, item.ReleID).Scan(&exists); err != nil || exists != 1 {
			return 0, fmt.Errorf("un aparato de la escena no pertenece a la empresa")
		}
		item.Orden = index
	}
	if scene.ID > 0 {
		result, err := execTxSQLCompat(tx, `UPDATE empresa_control_electrico_escenas SET nombre=?, descripcion=?, fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=?, estado='activo' WHERE empresa_id=? AND id=?`, scene.Nombre, strings.TrimSpace(scene.Descripcion), strings.TrimSpace(scene.UsuarioCreador), scene.EmpresaID, scene.ID)
		if err != nil {
			return 0, err
		}
		affected, _ := result.RowsAffected()
		if affected != 1 {
			return 0, sql.ErrNoRows
		}
	} else {
		if err := queryRowTxSQLCompat(tx, `INSERT INTO empresa_control_electrico_escenas (empresa_id,nombre,descripcion,usuario_creador,estado) VALUES (?,?,?,?, 'activo') RETURNING id`, scene.EmpresaID, scene.Nombre, strings.TrimSpace(scene.Descripcion), strings.TrimSpace(scene.UsuarioCreador)).Scan(&scene.ID); err != nil {
			return 0, err
		}
	}
	if _, err := execTxSQLCompat(tx, `DELETE FROM empresa_control_electrico_escena_items WHERE empresa_id=? AND escena_id=?`, scene.EmpresaID, scene.ID); err != nil {
		return 0, err
	}
	for _, item := range scene.Items {
		if _, err := execTxSQLCompat(tx, `INSERT INTO empresa_control_electrico_escena_items (empresa_id,escena_id,rele_id,estado_objetivo,orden) VALUES (?,?,?,?,?)`, scene.EmpresaID, scene.ID, item.ReleID, item.EstadoObjetivo, item.Orden); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return scene.ID, nil
}

func SetEmpresaControlElectricoEscenaEstado(dbConn *sql.DB, empresaID, sceneID int64, state string) error {
	state = strings.ToLower(strings.TrimSpace(state))
	if state != "activo" && state != "inactivo" {
		return errors.New("estado invalido")
	}
	result, err := execSQLCompat(dbConn, `UPDATE empresa_control_electrico_escenas SET estado=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, state, empresaID, sceneID)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected != 1 {
		return sql.ErrNoRows
	}
	return nil
}
