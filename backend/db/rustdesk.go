package db

import (
	"database/sql"
)

type RustDeskDevice struct {
	ID         int64  `json:"id"`
	EmpresaID  int64  `json:"empresa_id"`
	RustDeskID string `json:"rustdesk_id"`
	Password   string `json:"password,omitempty"`
	Nombre     string `json:"nombre"`
	Rol        string `json:"rol"`
	Estado     string `json:"estado"`
}

func EnsureRustDeskSchema(dbEmp *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS empresa_rustdesk_devices (
		id BIGSERIAL PRIMARY KEY,
		empresa_id INTEGER NOT NULL,
		rustdesk_id TEXT NOT NULL,
		password TEXT,
		nombre TEXT,
		rol TEXT DEFAULT 'soporte',
		estado TEXT DEFAULT 'activo',
		fecha_creacion DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(empresa_id, rustdesk_id)
	)`
	_, err := dbEmp.Exec(query)
	return err
}

func GetRustDeskDevices(dbEmp *sql.DB, empresaID int64) ([]RustDeskDevice, error) {
	rows, err := dbEmp.Query("SELECT id, empresa_id, rustdesk_id, nombre, rol, estado FROM empresa_rustdesk_devices WHERE empresa_id = ? AND estado = 'activo'", empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []RustDeskDevice
	for rows.Next() {
		var d RustDeskDevice
		if err := rows.Scan(&d.ID, &d.EmpresaID, &d.RustDeskID, &d.Nombre, &d.Rol, &d.Estado); err == nil {
			devices = append(devices, d)
		}
	}
	return devices, nil
}

func RegisterRustDeskDevice(dbEmp *sql.DB, d RustDeskDevice) error {
	_, err := dbEmp.Exec("INSERT INTO empresa_rustdesk_devices (empresa_id, rustdesk_id, password, nombre, rol, estado) VALUES (?, ?, ?, ?, ?, ?) ON CONFLICT(empresa_id, rustdesk_id) DO UPDATE SET nombre = excluded.nombre, rol = excluded.rol",
		d.EmpresaID, d.RustDeskID, d.Password, d.Nombre, d.Rol, d.Estado)
	return err
}
