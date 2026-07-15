package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

type EmpresaApartamentoTuristicoConfig struct {
	EmpresaID                int64   `json:"empresa_id"`
	NombreSistema            string  `json:"nombre_sistema"`
	Moneda                   string  `json:"moneda"`
	HoraCheckIn              string  `json:"hora_check_in"`
	HoraCheckOut             string  `json:"hora_check_out"`
	DepositoPorcentaje       float64 `json:"deposito_porcentaje"`
	ImpuestoPorcentaje       float64 `json:"impuesto_porcentaje"`
	AutoProgramarLimpieza    bool    `json:"auto_programar_limpieza"`
	PermitirReservasPublicas bool    `json:"permitir_reservas_publicas"`
	RequerirDocumentoHuesped bool    `json:"requerir_documento_huesped"`
	FechaActualizacion       string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador           string  `json:"usuario_creador,omitempty"`
}

type EmpresaApartamentoTuristicoUnidad struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	ServicioID         int64   `json:"servicio_id,omitempty"`
	Codigo             string  `json:"codigo"`
	Nombre             string  `json:"nombre"`
	Tipo               string  `json:"tipo"`
	Ubicacion          string  `json:"ubicacion,omitempty"`
	Capacidad          int     `json:"capacidad"`
	Habitaciones       int     `json:"habitaciones"`
	Camas              int     `json:"camas"`
	Banos              int     `json:"banos"`
	PrecioBaseNoche    float64 `json:"precio_base_noche"`
	TarifaLimpieza     float64 `json:"tarifa_limpieza"`
	DepositoSugerido   float64 `json:"deposito_sugerido"`
	EstadoOperativo    string  `json:"estado_operativo"`
	EstadoOcupacion    string  `json:"estado_ocupacion"`
	UrlFoto            string  `json:"url_foto,omitempty"`
	Amenidades         string  `json:"amenidades,omitempty"`
	ReglasCasa         string  `json:"reglas_casa,omitempty"`
	Notas              string  `json:"notas,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
}

type EmpresaApartamentoTuristicoTarifa struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	ApartamentoID      int64   `json:"apartamento_id,omitempty"`
	ApartamentoNombre  string  `json:"apartamento_nombre,omitempty"`
	Nombre             string  `json:"nombre"`
	Canal              string  `json:"canal"`
	FechaDesde         string  `json:"fecha_desde,omitempty"`
	FechaHasta         string  `json:"fecha_hasta,omitempty"`
	PrecioNoche        float64 `json:"precio_noche"`
	MinimoNoches       int     `json:"minimo_noches"`
	DescuentoSemanal   float64 `json:"descuento_semanal"`
	DescuentoMensual   float64 `json:"descuento_mensual"`
	Estado             string  `json:"estado"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
}

type EmpresaApartamentoTuristicoReserva struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	ApartamentoID      int64   `json:"apartamento_id"`
	ClienteID          int64   `json:"cliente_id,omitempty"`
	ServicioID         int64   `json:"servicio_id,omitempty"`
	CarritoID          int64   `json:"carrito_id,omitempty"`
	CarritoItemID      int64   `json:"carrito_item_id,omitempty"`
	ApartamentoNombre  string  `json:"apartamento_nombre,omitempty"`
	CodigoReserva      string  `json:"codigo_reserva"`
	HuespedNombre      string  `json:"huesped_nombre"`
	HuespedDocumento   string  `json:"huesped_documento,omitempty"`
	HuespedTelefono    string  `json:"huesped_telefono,omitempty"`
	HuespedEmail       string  `json:"huesped_email,omitempty"`
	CantidadHuespedes  int     `json:"cantidad_huespedes"`
	FechaEntrada       string  `json:"fecha_entrada"`
	FechaSalida        string  `json:"fecha_salida"`
	Noches             int     `json:"noches"`
	Canal              string  `json:"canal"`
	MetodoPago         string  `json:"metodo_pago,omitempty"`
	EstadoReserva      string  `json:"estado_reserva"`
	EstadoPago         string  `json:"estado_pago"`
	Subtotal           float64 `json:"subtotal"`
	Limpieza           float64 `json:"limpieza"`
	Impuestos          float64 `json:"impuestos"`
	Deposito           float64 `json:"deposito"`
	Total              float64 `json:"total"`
	SaldoPendiente     float64 `json:"saldo_pendiente"`
	CodigoAcceso       string  `json:"codigo_acceso,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
	FechaCheckIn       string  `json:"fecha_check_in,omitempty"`
	FechaCheckOut      string  `json:"fecha_check_out,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
}

type EmpresaApartamentoTuristicoTarea struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	ApartamentoID      int64   `json:"apartamento_id"`
	ApartamentoNombre  string  `json:"apartamento_nombre,omitempty"`
	ReservaID          int64   `json:"reserva_id,omitempty"`
	Tipo               string  `json:"tipo"`
	Prioridad          string  `json:"prioridad"`
	Estado             string  `json:"estado"`
	Responsable        string  `json:"responsable,omitempty"`
	FechaProgramada    string  `json:"fecha_programada,omitempty"`
	FechaCierre        string  `json:"fecha_cierre,omitempty"`
	CostoEstimado      float64 `json:"costo_estimado"`
	CostoReal          float64 `json:"costo_real"`
	Descripcion        string  `json:"descripcion,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
}

type EmpresaApartamentoTuristicoDashboard struct {
	EmpresaID        int64                                `json:"empresa_id"`
	Apartamentos     int                                  `json:"apartamentos"`
	Disponibles      int                                  `json:"disponibles"`
	Ocupados         int                                  `json:"ocupados"`
	Limpieza         int                                  `json:"limpieza"`
	Mantenimiento    int                                  `json:"mantenimiento"`
	ReservasActivas  int                                  `json:"reservas_activas"`
	CheckInsHoy      int                                  `json:"checkins_hoy"`
	CheckOutsHoy     int                                  `json:"checkouts_hoy"`
	IngresosMes      float64                              `json:"ingresos_mes"`
	Config           EmpresaApartamentoTuristicoConfig    `json:"config"`
	Unidades         []EmpresaApartamentoTuristicoUnidad  `json:"unidades"`
	Reservas         []EmpresaApartamentoTuristicoReserva `json:"reservas"`
	TareasPendientes []EmpresaApartamentoTuristicoTarea   `json:"tareas_pendientes"`
}

func EnsureEmpresaApartamentosTuristicosSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_apartamentos_turisticos_config (
			empresa_id BIGINT PRIMARY KEY,
			nombre_sistema TEXT DEFAULT 'Apartamentos turisticos',
			moneda TEXT DEFAULT 'COP',
			hora_check_in TEXT DEFAULT '15:00',
			hora_check_out TEXT DEFAULT '11:00',
			deposito_porcentaje NUMERIC(7,2) DEFAULT 30,
			impuesto_porcentaje NUMERIC(7,2) DEFAULT 0,
			auto_programar_limpieza INTEGER DEFAULT 1,
			permitir_reservas_publicas INTEGER DEFAULT 1,
			requerir_documento_huesped INTEGER DEFAULT 1,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_apartamentos_turisticos_unidades (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			servicio_id BIGINT,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			tipo TEXT DEFAULT 'apartamento',
			ubicacion TEXT,
			capacidad INTEGER DEFAULT 2,
			habitaciones INTEGER DEFAULT 1,
			camas INTEGER DEFAULT 1,
			banos INTEGER DEFAULT 1,
			precio_base_noche NUMERIC(14,2) DEFAULT 0,
			tarifa_limpieza NUMERIC(14,2) DEFAULT 0,
			deposito_sugerido NUMERIC(14,2) DEFAULT 0,
			estado_operativo TEXT DEFAULT 'activo',
			estado_ocupacion TEXT DEFAULT 'disponible',
			url_foto TEXT,
			amenidades TEXT,
			reglas_casa TEXT,
			notas TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_apart_tur_unidad_codigo ON empresa_apartamentos_turisticos_unidades(empresa_id, codigo)`,
		`CREATE INDEX IF NOT EXISTS ix_apart_tur_unidad_estado ON empresa_apartamentos_turisticos_unidades(empresa_id, estado_operativo, estado_ocupacion)`,
		`CREATE TABLE IF NOT EXISTS empresa_apartamentos_turisticos_tarifas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			apartamento_id BIGINT DEFAULT 0,
			nombre TEXT NOT NULL,
			canal TEXT DEFAULT 'directo',
			fecha_desde TEXT,
			fecha_hasta TEXT,
			precio_noche NUMERIC(14,2) DEFAULT 0,
			minimo_noches INTEGER DEFAULT 1,
			descuento_semanal NUMERIC(7,2) DEFAULT 0,
			descuento_mensual NUMERIC(7,2) DEFAULT 0,
			estado TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_apart_tur_tarifas_lookup ON empresa_apartamentos_turisticos_tarifas(empresa_id, apartamento_id, estado)`,
		`CREATE TABLE IF NOT EXISTS empresa_apartamentos_turisticos_reservas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			apartamento_id BIGINT NOT NULL,
			cliente_id BIGINT,
			servicio_id BIGINT,
			carrito_id BIGINT,
			carrito_item_id BIGINT,
			codigo_reserva TEXT NOT NULL,
			huesped_nombre TEXT NOT NULL,
			huesped_documento TEXT,
			huesped_telefono TEXT,
			huesped_email TEXT,
			cantidad_huespedes INTEGER DEFAULT 1,
			fecha_entrada TEXT NOT NULL,
			fecha_salida TEXT NOT NULL,
			noches INTEGER DEFAULT 1,
			canal TEXT DEFAULT 'directo',
			metodo_pago TEXT DEFAULT 'efectivo',
			estado_reserva TEXT DEFAULT 'confirmada',
			estado_pago TEXT DEFAULT 'pendiente',
			subtotal NUMERIC(14,2) DEFAULT 0,
			limpieza NUMERIC(14,2) DEFAULT 0,
			impuestos NUMERIC(14,2) DEFAULT 0,
			deposito NUMERIC(14,2) DEFAULT 0,
			total NUMERIC(14,2) DEFAULT 0,
			saldo_pendiente NUMERIC(14,2) DEFAULT 0,
			codigo_acceso TEXT,
			observaciones TEXT,
			fecha_check_in TEXT,
			fecha_check_out TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_apart_tur_reserva_codigo ON empresa_apartamentos_turisticos_reservas(empresa_id, codigo_reserva)`,
		`CREATE INDEX IF NOT EXISTS ix_apart_tur_reservas_fechas ON empresa_apartamentos_turisticos_reservas(empresa_id, apartamento_id, fecha_entrada, fecha_salida, estado_reserva)`,
		`CREATE TABLE IF NOT EXISTS empresa_apartamentos_turisticos_tareas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			apartamento_id BIGINT NOT NULL,
			reserva_id BIGINT DEFAULT 0,
			tipo TEXT DEFAULT 'limpieza',
			prioridad TEXT DEFAULT 'media',
			estado TEXT DEFAULT 'pendiente',
			responsable TEXT,
			fecha_programada TEXT,
			fecha_cierre TEXT,
			costo_estimado NUMERIC(14,2) DEFAULT 0,
			costo_real NUMERIC(14,2) DEFAULT 0,
			descripcion TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_apart_tur_tareas_estado ON empresa_apartamentos_turisticos_tareas(empresa_id, estado, fecha_programada)`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	extraColumns := []struct {
		table  string
		column string
		def    string
	}{
		{"empresa_apartamentos_turisticos_unidades", "servicio_id", "BIGINT"},
		{"empresa_apartamentos_turisticos_reservas", "cliente_id", "BIGINT"},
		{"empresa_apartamentos_turisticos_reservas", "servicio_id", "BIGINT"},
		{"empresa_apartamentos_turisticos_reservas", "carrito_id", "BIGINT"},
		{"empresa_apartamentos_turisticos_reservas", "carrito_item_id", "BIGINT"},
		{"empresa_apartamentos_turisticos_reservas", "metodo_pago", "TEXT DEFAULT 'efectivo'"},
	}
	for _, col := range extraColumns {
		if err := ensureColumnIfMissing(dbConn, col.table, col.column, col.def); err != nil {
			return err
		}
	}
	for _, stmt := range []string{
		`CREATE INDEX IF NOT EXISTS ix_apart_tur_unidad_servicio ON empresa_apartamentos_turisticos_unidades(empresa_id, servicio_id)`,
		`CREATE INDEX IF NOT EXISTS ix_apart_tur_reserva_carrito ON empresa_apartamentos_turisticos_reservas(empresa_id, carrito_id)`,
	} {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func defaultApartamentoTuristicoConfig(empresaID int64) EmpresaApartamentoTuristicoConfig {
	return EmpresaApartamentoTuristicoConfig{
		EmpresaID:                empresaID,
		NombreSistema:            "Apartamentos turisticos",
		Moneda:                   "COP",
		HoraCheckIn:              "15:00",
		HoraCheckOut:             "11:00",
		DepositoPorcentaje:       30,
		ImpuestoPorcentaje:       0,
		AutoProgramarLimpieza:    true,
		PermitirReservasPublicas: true,
		RequerirDocumentoHuesped: true,
	}
}

func GetEmpresaApartamentoTuristicoConfig(dbConn *sql.DB, empresaID int64) (EmpresaApartamentoTuristicoConfig, error) {
	if err := EnsureEmpresaApartamentosTuristicosSchema(dbConn); err != nil {
		return EmpresaApartamentoTuristicoConfig{}, err
	}
	cfg := defaultApartamentoTuristicoConfig(empresaID)
	var autoLimpieza, publicas, doc int
	err := QueryRowCompat(dbConn, `SELECT empresa_id, COALESCE(nombre_sistema,''), COALESCE(moneda,'COP'), COALESCE(hora_check_in,'15:00'), COALESCE(hora_check_out,'11:00'), COALESCE(deposito_porcentaje,30), COALESCE(impuesto_porcentaje,0), COALESCE(auto_programar_limpieza,1), COALESCE(permitir_reservas_publicas,1), COALESCE(requerir_documento_huesped,1), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'') FROM empresa_apartamentos_turisticos_config WHERE empresa_id = ?`, empresaID).Scan(&cfg.EmpresaID, &cfg.NombreSistema, &cfg.Moneda, &cfg.HoraCheckIn, &cfg.HoraCheckOut, &cfg.DepositoPorcentaje, &cfg.ImpuestoPorcentaje, &autoLimpieza, &publicas, &doc, &cfg.FechaActualizacion, &cfg.UsuarioCreador)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return cfg, nil
		}
		return EmpresaApartamentoTuristicoConfig{}, err
	}
	cfg.AutoProgramarLimpieza = autoLimpieza > 0
	cfg.PermitirReservasPublicas = publicas > 0
	cfg.RequerirDocumentoHuesped = doc > 0
	return normalizeApartamentoTuristicoConfig(cfg), nil
}

func UpsertEmpresaApartamentoTuristicoConfig(dbConn *sql.DB, cfg EmpresaApartamentoTuristicoConfig) error {
	if err := EnsureEmpresaApartamentosTuristicosSchema(dbConn); err != nil {
		return err
	}
	cfg = normalizeApartamentoTuristicoConfig(cfg)
	_, err := ExecCompat(dbConn, `INSERT INTO empresa_apartamentos_turisticos_config (empresa_id,nombre_sistema,moneda,hora_check_in,hora_check_out,deposito_porcentaje,impuesto_porcentaje,auto_programar_limpieza,permitir_reservas_publicas,requerir_documento_huesped,fecha_actualizacion,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP,?)
		ON CONFLICT (empresa_id) DO UPDATE SET nombre_sistema=EXCLUDED.nombre_sistema, moneda=EXCLUDED.moneda, hora_check_in=EXCLUDED.hora_check_in, hora_check_out=EXCLUDED.hora_check_out, deposito_porcentaje=EXCLUDED.deposito_porcentaje, impuesto_porcentaje=EXCLUDED.impuesto_porcentaje, auto_programar_limpieza=EXCLUDED.auto_programar_limpieza, permitir_reservas_publicas=EXCLUDED.permitir_reservas_publicas, requerir_documento_huesped=EXCLUDED.requerir_documento_huesped, fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=EXCLUDED.usuario_creador`,
		cfg.EmpresaID, cfg.NombreSistema, cfg.Moneda, cfg.HoraCheckIn, cfg.HoraCheckOut, cfg.DepositoPorcentaje, cfg.ImpuestoPorcentaje, apartTurBoolInt(cfg.AutoProgramarLimpieza), apartTurBoolInt(cfg.PermitirReservasPublicas), apartTurBoolInt(cfg.RequerirDocumentoHuesped), cfg.UsuarioCreador)
	return err
}

func ListEmpresaApartamentosTuristicosUnidades(dbConn *sql.DB, empresaID int64) ([]EmpresaApartamentoTuristicoUnidad, error) {
	if err := EnsureEmpresaApartamentosTuristicosSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,COALESCE(servicio_id,0),codigo,nombre,COALESCE(tipo,''),COALESCE(ubicacion,''),COALESCE(capacidad,2),COALESCE(habitaciones,1),COALESCE(camas,1),COALESCE(banos,1),COALESCE(precio_base_noche,0),COALESCE(tarifa_limpieza,0),COALESCE(deposito_sugerido,0),COALESCE(estado_operativo,'activo'),COALESCE(estado_ocupacion,'disponible'),COALESCE(url_foto,''),COALESCE(amenidades,''),COALESCE(reglas_casa,''),COALESCE(notas,''),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,''),COALESCE(usuario_creador,'') FROM empresa_apartamentos_turisticos_unidades WHERE empresa_id=? ORDER BY codigo`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaApartamentoTuristicoUnidad{}
	for rows.Next() {
		var x EmpresaApartamentoTuristicoUnidad
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.ServicioID, &x.Codigo, &x.Nombre, &x.Tipo, &x.Ubicacion, &x.Capacidad, &x.Habitaciones, &x.Camas, &x.Banos, &x.PrecioBaseNoche, &x.TarifaLimpieza, &x.DepositoSugerido, &x.EstadoOperativo, &x.EstadoOcupacion, &x.UrlFoto, &x.Amenidades, &x.ReglasCasa, &x.Notas, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func CreateEmpresaApartamentoTuristicoUnidad(dbConn *sql.DB, x EmpresaApartamentoTuristicoUnidad) (int64, error) {
	if err := EnsureEmpresaApartamentosTuristicosSchema(dbConn); err != nil {
		return 0, err
	}
	x = normalizeApartamentoTuristicoUnidad(x)
	if x.Codigo == "" || x.Nombre == "" {
		return 0, errors.New("codigo y nombre son obligatorios")
	}
	servicioID, err := ensureApartTurUnidadServicio(dbConn, x, x.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	x.ServicioID = servicioID
	var id int64
	err = QueryRowCompat(dbConn, `INSERT INTO empresa_apartamentos_turisticos_unidades (empresa_id,servicio_id,codigo,nombre,tipo,ubicacion,capacidad,habitaciones,camas,banos,precio_base_noche,tarifa_limpieza,deposito_sugerido,estado_operativo,estado_ocupacion,url_foto,amenidades,reglas_casa,notas,fecha_creacion,fecha_actualizacion,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,?) RETURNING id`,
		x.EmpresaID, nullableID(x.ServicioID), x.Codigo, x.Nombre, x.Tipo, x.Ubicacion, x.Capacidad, x.Habitaciones, x.Camas, x.Banos, x.PrecioBaseNoche, x.TarifaLimpieza, x.DepositoSugerido, x.EstadoOperativo, x.EstadoOcupacion, x.UrlFoto, x.Amenidades, x.ReglasCasa, x.Notas, x.UsuarioCreador).Scan(&id)
	return id, err
}

func ListEmpresaApartamentosTuristicosTarifas(dbConn *sql.DB, empresaID int64) ([]EmpresaApartamentoTuristicoTarifa, error) {
	if err := EnsureEmpresaApartamentosTuristicosSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT t.id,t.empresa_id,COALESCE(t.apartamento_id,0),COALESCE(u.nombre,''),COALESCE(t.nombre,''),COALESCE(t.canal,''),COALESCE(t.fecha_desde,''),COALESCE(t.fecha_hasta,''),COALESCE(t.precio_noche,0),COALESCE(t.minimo_noches,1),COALESCE(t.descuento_semanal,0),COALESCE(t.descuento_mensual,0),COALESCE(t.estado,'activo'),COALESCE(t.fecha_creacion,''),COALESCE(t.fecha_actualizacion,''),COALESCE(t.usuario_creador,'') FROM empresa_apartamentos_turisticos_tarifas t LEFT JOIN empresa_apartamentos_turisticos_unidades u ON u.id=t.apartamento_id AND u.empresa_id=t.empresa_id WHERE t.empresa_id=? ORDER BY t.id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaApartamentoTuristicoTarifa{}
	for rows.Next() {
		var x EmpresaApartamentoTuristicoTarifa
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.ApartamentoID, &x.ApartamentoNombre, &x.Nombre, &x.Canal, &x.FechaDesde, &x.FechaHasta, &x.PrecioNoche, &x.MinimoNoches, &x.DescuentoSemanal, &x.DescuentoMensual, &x.Estado, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func CreateEmpresaApartamentoTuristicoTarifa(dbConn *sql.DB, x EmpresaApartamentoTuristicoTarifa) (int64, error) {
	if err := EnsureEmpresaApartamentosTuristicosSchema(dbConn); err != nil {
		return 0, err
	}
	x.Nombre = strings.TrimSpace(x.Nombre)
	if x.Nombre == "" {
		x.Nombre = "Tarifa directa"
	}
	x.Canal = firstApartTurState(x.Canal, "directo")
	x.Estado = firstApartTurState(x.Estado, "activo")
	if x.MinimoNoches <= 0 {
		x.MinimoNoches = 1
	}
	var id int64
	err := QueryRowCompat(dbConn, `INSERT INTO empresa_apartamentos_turisticos_tarifas (empresa_id,apartamento_id,nombre,canal,fecha_desde,fecha_hasta,precio_noche,minimo_noches,descuento_semanal,descuento_mensual,estado,fecha_creacion,fecha_actualizacion,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,?) RETURNING id`,
		x.EmpresaID, x.ApartamentoID, x.Nombre, x.Canal, x.FechaDesde, x.FechaHasta, x.PrecioNoche, x.MinimoNoches, x.DescuentoSemanal, x.DescuentoMensual, x.Estado, x.UsuarioCreador).Scan(&id)
	return id, err
}

func ListEmpresaApartamentosTuristicosReservas(dbConn *sql.DB, empresaID int64, estado string, limit int) ([]EmpresaApartamentoTuristicoReserva, error) {
	if err := EnsureEmpresaApartamentosTuristicosSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	where := "r.empresa_id=?"
	args := []interface{}{empresaID}
	if strings.TrimSpace(estado) != "" {
		where += " AND LOWER(COALESCE(r.estado_reserva,''))=?"
		args = append(args, strings.ToLower(strings.TrimSpace(estado)))
	}
	args = append(args, limit)
	rows, err := ExecQueryCompat(dbConn, apartTurReservaSelect()+` WHERE `+where+` ORDER BY r.id DESC LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaApartamentoTuristicoReserva{}
	for rows.Next() {
		x, err := scanApartTurReserva(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func CreateEmpresaApartamentoTuristicoReserva(dbConn *sql.DB, x EmpresaApartamentoTuristicoReserva) (int64, error) {
	if err := EnsureEmpresaApartamentosTuristicosSchema(dbConn); err != nil {
		return 0, err
	}
	if x.ApartamentoID <= 0 || strings.TrimSpace(x.HuespedNombre) == "" {
		return 0, errors.New("apartamento y huesped son obligatorios")
	}
	inicio, fin, noches, err := normalizeApartTurDates(x.FechaEntrada, x.FechaSalida)
	if err != nil {
		return 0, err
	}
	conflict, err := ApartamentoTuristicoTieneConflicto(dbConn, x.EmpresaID, x.ApartamentoID, inicio, fin, 0)
	if err != nil {
		return 0, err
	}
	if conflict {
		return 0, errors.New("apartamento no disponible en esas fechas")
	}
	cfg, _ := GetEmpresaApartamentoTuristicoConfig(dbConn, x.EmpresaID)
	unit, err := GetEmpresaApartamentoTuristicoUnidad(dbConn, x.EmpresaID, x.ApartamentoID)
	if err != nil {
		return 0, err
	}
	price := resolveApartTurNightPrice(dbConn, x.EmpresaID, x.ApartamentoID, unit.PrecioBaseNoche)
	subtotal := roundApartTur(price * float64(noches))
	limpieza := roundApartTur(unit.TarifaLimpieza)
	impuestos := roundApartTur((subtotal + limpieza) * cfg.ImpuestoPorcentaje / 100)
	deposito := roundApartTur((subtotal + limpieza) * cfg.DepositoPorcentaje / 100)
	total := roundApartTur(subtotal + limpieza + impuestos)
	code, err := nextApartTurReservaCode(dbConn, x.EmpresaID)
	if err != nil {
		return 0, err
	}
	access := generateApartTurAccessCode()
	estado := firstApartTurState(x.EstadoReserva, "confirmada")
	pago := firstApartTurState(x.EstadoPago, "pendiente")
	canal := firstApartTurState(x.Canal, "directo")
	metodo := NormalizeMetodoPagoCarrito(x.MetodoPago)
	if metodo == "" {
		metodo = "efectivo"
	}
	clienteID, err := ensureApartTurClienteCore(dbConn, x, x.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	servicioID, err := ensureApartTurUnidadServicio(dbConn, unit, x.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	var id int64
	err = QueryRowCompat(dbConn, `INSERT INTO empresa_apartamentos_turisticos_reservas (empresa_id,apartamento_id,cliente_id,servicio_id,codigo_reserva,huesped_nombre,huesped_documento,huesped_telefono,huesped_email,cantidad_huespedes,fecha_entrada,fecha_salida,noches,canal,metodo_pago,estado_reserva,estado_pago,subtotal,limpieza,impuestos,deposito,total,saldo_pendiente,codigo_acceso,observaciones,fecha_creacion,fecha_actualizacion,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,?) RETURNING id`,
		x.EmpresaID, x.ApartamentoID, nullableID(clienteID), nullableID(servicioID), code, strings.TrimSpace(x.HuespedNombre), strings.TrimSpace(x.HuespedDocumento), strings.TrimSpace(x.HuespedTelefono), strings.TrimSpace(x.HuespedEmail), maxApartTurInt(x.CantidadHuespedes, 1), inicio, fin, noches, canal, metodo, estado, pago, subtotal, limpieza, impuestos, deposito, total, total, access, strings.TrimSpace(x.Observaciones), x.UsuarioCreador).Scan(&id)
	if err != nil {
		return 0, err
	}
	_, _ = ExecCompat(dbConn, `UPDATE empresa_apartamentos_turisticos_unidades SET estado_ocupacion='reservado', fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=? AND estado_ocupacion='disponible'`, x.EmpresaID, x.ApartamentoID)
	return id, nil
}

func GetEmpresaApartamentoTuristicoUnidad(dbConn *sql.DB, empresaID, id int64) (EmpresaApartamentoTuristicoUnidad, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,COALESCE(servicio_id,0),codigo,nombre,COALESCE(tipo,''),COALESCE(ubicacion,''),COALESCE(capacidad,2),COALESCE(habitaciones,1),COALESCE(camas,1),COALESCE(banos,1),COALESCE(precio_base_noche,0),COALESCE(tarifa_limpieza,0),COALESCE(deposito_sugerido,0),COALESCE(estado_operativo,'activo'),COALESCE(estado_ocupacion,'disponible'),COALESCE(url_foto,''),COALESCE(amenidades,''),COALESCE(reglas_casa,''),COALESCE(notas,''),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,''),COALESCE(usuario_creador,'') FROM empresa_apartamentos_turisticos_unidades WHERE empresa_id=? AND id=? LIMIT 1`, empresaID, id)
	if err != nil {
		return EmpresaApartamentoTuristicoUnidad{}, err
	}
	defer rows.Close()
	if rows.Next() {
		var x EmpresaApartamentoTuristicoUnidad
		err := rows.Scan(&x.ID, &x.EmpresaID, &x.ServicioID, &x.Codigo, &x.Nombre, &x.Tipo, &x.Ubicacion, &x.Capacidad, &x.Habitaciones, &x.Camas, &x.Banos, &x.PrecioBaseNoche, &x.TarifaLimpieza, &x.DepositoSugerido, &x.EstadoOperativo, &x.EstadoOcupacion, &x.UrlFoto, &x.Amenidades, &x.ReglasCasa, &x.Notas, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador)
		return x, err
	}
	return EmpresaApartamentoTuristicoUnidad{}, sql.ErrNoRows
}

func ApartamentoTuristicoTieneConflicto(dbConn *sql.DB, empresaID, aptID int64, inicio, fin string, ignoreID int64) (bool, error) {
	var count int
	err := QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_apartamentos_turisticos_reservas WHERE empresa_id=? AND apartamento_id=? AND id<>? AND LOWER(COALESCE(estado_reserva,'')) NOT IN ('cancelada','checkout','no_show') AND fecha_entrada < ? AND fecha_salida > ?`, empresaID, aptID, ignoreID, fin, inicio).Scan(&count)
	return count > 0, err
}

func CambiarEstadoApartamentoTuristicoReserva(dbConn *sql.DB, empresaID, reservaID int64, estado, usuario string) error {
	estado = strings.ToLower(strings.TrimSpace(estado))
	if estado == "" {
		return errors.New("estado es obligatorio")
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	extra := ""
	args := []interface{}{estado}
	if estado == "checkin" {
		extra = ", fecha_check_in=?"
		args = append(args, now)
	}
	if estado == "checkout" {
		extra = ", fecha_check_out=?"
		args = append(args, now)
	}
	args = append(args, usuario, empresaID, reservaID)
	_, err := ExecCompat(dbConn, `UPDATE empresa_apartamentos_turisticos_reservas SET estado_reserva=?`+extra+`, fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=COALESCE(NULLIF(?,''),usuario_creador) WHERE empresa_id=? AND id=?`, args...)
	if err != nil {
		return err
	}
	var aptID int64
	_ = QueryRowCompat(dbConn, `SELECT apartamento_id FROM empresa_apartamentos_turisticos_reservas WHERE empresa_id=? AND id=?`, empresaID, reservaID).Scan(&aptID)
	if aptID > 0 {
		nextState := "reservado"
		if estado == "checkin" {
			nextState = "ocupado"
		}
		if estado == "checkout" || estado == "cancelada" {
			nextState = "limpieza"
		}
		_, _ = ExecCompat(dbConn, `UPDATE empresa_apartamentos_turisticos_unidades SET estado_ocupacion=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, nextState, empresaID, aptID)
		cfg, _ := GetEmpresaApartamentoTuristicoConfig(dbConn, empresaID)
		if estado == "checkout" && cfg.AutoProgramarLimpieza {
			_, _ = ExecCompat(dbConn, `INSERT INTO empresa_apartamentos_turisticos_tareas (empresa_id,apartamento_id,reserva_id,tipo,prioridad,estado,fecha_programada,descripcion,fecha_creacion,fecha_actualizacion,usuario_creador) VALUES (?,?,?,?,?,?,?, ?,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,?)`, empresaID, aptID, reservaID, "limpieza", "alta", "pendiente", now, "Limpieza posterior a checkout", usuario)
		}
	}
	if estado == "checkout" {
		reserva, err := GetEmpresaApartamentoTuristicoReserva(dbConn, empresaID, reservaID)
		if err != nil {
			return err
		}
		if reserva.CarritoID <= 0 && reserva.Total > 0 {
			carritoID, itemID, clienteID, servicioID, err := createApartTurReservaCarrito(dbConn, reserva, usuario)
			if err != nil {
				return err
			}
			_, err = ExecCompat(dbConn, `UPDATE empresa_apartamentos_turisticos_reservas SET cliente_id=?, servicio_id=?, carrito_id=?, carrito_item_id=?, estado_pago='pagado', saldo_pendiente=0 WHERE empresa_id=? AND id=?`, nullableID(clienteID), nullableID(servicioID), nullableID(carritoID), nullableID(itemID), empresaID, reservaID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func CambiarEstadoApartamentoTuristicoUnidad(dbConn *sql.DB, empresaID, apartamentoID int64, estadoOperativo, estadoOcupacion, usuario string) error {
	if err := EnsureEmpresaApartamentosTuristicosSchema(dbConn); err != nil {
		return err
	}
	if apartamentoID <= 0 {
		return errors.New("apartamento_id es obligatorio")
	}
	estadoOperativo = firstApartTurState(estadoOperativo, "activo")
	estadoOcupacion = firstApartTurState(estadoOcupacion, "disponible")
	res, err := ExecCompat(dbConn, `UPDATE empresa_apartamentos_turisticos_unidades SET estado_operativo=?, estado_ocupacion=?, fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=COALESCE(NULLIF(?,''),usuario_creador) WHERE empresa_id=? AND id=?`, estadoOperativo, estadoOcupacion, usuario, empresaID, apartamentoID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func CreateEmpresaApartamentoTuristicoTarea(dbConn *sql.DB, x EmpresaApartamentoTuristicoTarea) (int64, error) {
	if err := EnsureEmpresaApartamentosTuristicosSchema(dbConn); err != nil {
		return 0, err
	}
	if x.ApartamentoID <= 0 {
		return 0, errors.New("apartamento_id es obligatorio")
	}
	x.Tipo = firstApartTurState(x.Tipo, "limpieza")
	x.Prioridad = firstApartTurState(x.Prioridad, "media")
	x.Estado = firstApartTurState(x.Estado, "pendiente")
	var id int64
	err := QueryRowCompat(dbConn, `INSERT INTO empresa_apartamentos_turisticos_tareas (empresa_id,apartamento_id,reserva_id,tipo,prioridad,estado,responsable,fecha_programada,costo_estimado,costo_real,descripcion,fecha_creacion,fecha_actualizacion,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,?) RETURNING id`, x.EmpresaID, x.ApartamentoID, x.ReservaID, x.Tipo, x.Prioridad, x.Estado, x.Responsable, x.FechaProgramada, x.CostoEstimado, x.CostoReal, x.Descripcion, x.UsuarioCreador).Scan(&id)
	return id, err
}

func CambiarEstadoApartamentoTuristicoTarea(dbConn *sql.DB, empresaID, tareaID int64, estado, usuario string, costoReal float64) error {
	if err := EnsureEmpresaApartamentosTuristicosSchema(dbConn); err != nil {
		return err
	}
	if tareaID <= 0 {
		return errors.New("tarea_id es obligatorio")
	}
	estado = firstApartTurState(estado, "completada")
	extra := ""
	args := []interface{}{estado}
	if estado == "completada" || estado == "cerrada" || estado == "cancelada" {
		extra = ", fecha_cierre=CURRENT_TIMESTAMP"
	}
	if costoReal >= 0 {
		extra += ", costo_real=?"
		args = append(args, costoReal)
	}
	args = append(args, usuario, empresaID, tareaID)
	res, err := ExecCompat(dbConn, `UPDATE empresa_apartamentos_turisticos_tareas SET estado=?`+extra+`, fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=COALESCE(NULLIF(?,''),usuario_creador) WHERE empresa_id=? AND id=?`, args...)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func ListEmpresaApartamentoTuristicoTareas(dbConn *sql.DB, empresaID int64, estado string, limit int) ([]EmpresaApartamentoTuristicoTarea, error) {
	if err := EnsureEmpresaApartamentosTuristicosSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	where := "t.empresa_id=?"
	args := []interface{}{empresaID}
	if strings.TrimSpace(estado) != "" {
		where += " AND LOWER(COALESCE(t.estado,''))=?"
		args = append(args, strings.ToLower(strings.TrimSpace(estado)))
	}
	args = append(args, limit)
	rows, err := ExecQueryCompat(dbConn, `SELECT t.id,t.empresa_id,t.apartamento_id,COALESCE(u.nombre,''),COALESCE(t.reserva_id,0),COALESCE(t.tipo,''),COALESCE(t.prioridad,''),COALESCE(t.estado,''),COALESCE(t.responsable,''),COALESCE(t.fecha_programada,''),COALESCE(t.fecha_cierre,''),COALESCE(t.costo_estimado,0),COALESCE(t.costo_real,0),COALESCE(t.descripcion,''),COALESCE(t.fecha_creacion,''),COALESCE(t.fecha_actualizacion,''),COALESCE(t.usuario_creador,'') FROM empresa_apartamentos_turisticos_tareas t LEFT JOIN empresa_apartamentos_turisticos_unidades u ON u.id=t.apartamento_id AND u.empresa_id=t.empresa_id WHERE `+where+` ORDER BY t.id DESC LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaApartamentoTuristicoTarea{}
	for rows.Next() {
		var x EmpresaApartamentoTuristicoTarea
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.ApartamentoID, &x.ApartamentoNombre, &x.ReservaID, &x.Tipo, &x.Prioridad, &x.Estado, &x.Responsable, &x.FechaProgramada, &x.FechaCierre, &x.CostoEstimado, &x.CostoReal, &x.Descripcion, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func BuildEmpresaApartamentoTuristicoDashboard(dbConn *sql.DB, empresaID int64) (EmpresaApartamentoTuristicoDashboard, error) {
	cfg, err := GetEmpresaApartamentoTuristicoConfig(dbConn, empresaID)
	if err != nil {
		return EmpresaApartamentoTuristicoDashboard{}, err
	}
	units, err := ListEmpresaApartamentosTuristicosUnidades(dbConn, empresaID)
	if err != nil {
		return EmpresaApartamentoTuristicoDashboard{}, err
	}
	reservas, err := ListEmpresaApartamentosTuristicosReservas(dbConn, empresaID, "", 80)
	if err != nil {
		return EmpresaApartamentoTuristicoDashboard{}, err
	}
	tareas, err := ListEmpresaApartamentoTuristicoTareas(dbConn, empresaID, "pendiente", 80)
	if err != nil {
		return EmpresaApartamentoTuristicoDashboard{}, err
	}
	out := EmpresaApartamentoTuristicoDashboard{EmpresaID: empresaID, Config: cfg, Unidades: units, Reservas: reservas, TareasPendientes: tareas, Apartamentos: len(units)}
	for _, u := range units {
		switch strings.ToLower(u.EstadoOcupacion) {
		case "ocupado":
			out.Ocupados++
		case "limpieza":
			out.Limpieza++
		case "mantenimiento":
			out.Mantenimiento++
		default:
			if strings.ToLower(u.EstadoOperativo) == "activo" {
				out.Disponibles++
			}
		}
	}
	today := time.Now().Format("2006-01-02")
	month := time.Now().Format("2006-01")
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_apartamentos_turisticos_reservas WHERE empresa_id=? AND LOWER(COALESCE(estado_reserva,'')) IN ('confirmada','checkin')`, empresaID).Scan(&out.ReservasActivas)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_apartamentos_turisticos_reservas WHERE empresa_id=? AND fecha_entrada=?`, empresaID, today).Scan(&out.CheckInsHoy)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_apartamentos_turisticos_reservas WHERE empresa_id=? AND fecha_salida=?`, empresaID, today).Scan(&out.CheckOutsHoy)
	_ = QueryRowCompat(dbConn, `SELECT COALESCE(SUM(COALESCE(total,0)),0) FROM empresa_apartamentos_turisticos_reservas WHERE empresa_id=? AND CAST(COALESCE(fecha_creacion,'') AS TEXT) >= ?`, empresaID, month+"-01").Scan(&out.IngresosMes)
	return out, nil
}

func normalizeApartamentoTuristicoConfig(cfg EmpresaApartamentoTuristicoConfig) EmpresaApartamentoTuristicoConfig {
	cfg.NombreSistema = strings.TrimSpace(cfg.NombreSistema)
	if cfg.NombreSistema == "" {
		cfg.NombreSistema = "Apartamentos turisticos"
	}
	cfg.Moneda = strings.ToUpper(strings.TrimSpace(cfg.Moneda))
	if cfg.Moneda == "" {
		cfg.Moneda = "COP"
	}
	if strings.TrimSpace(cfg.HoraCheckIn) == "" {
		cfg.HoraCheckIn = "15:00"
	}
	if strings.TrimSpace(cfg.HoraCheckOut) == "" {
		cfg.HoraCheckOut = "11:00"
	}
	if cfg.DepositoPorcentaje < 0 {
		cfg.DepositoPorcentaje = 0
	}
	if cfg.ImpuestoPorcentaje < 0 {
		cfg.ImpuestoPorcentaje = 0
	}
	return cfg
}

func normalizeApartamentoTuristicoUnidad(x EmpresaApartamentoTuristicoUnidad) EmpresaApartamentoTuristicoUnidad {
	x.Codigo = strings.ToUpper(strings.TrimSpace(x.Codigo))
	x.Nombre = strings.TrimSpace(x.Nombre)
	x.Tipo = firstApartTurState(x.Tipo, "apartamento")
	x.EstadoOperativo = firstApartTurState(x.EstadoOperativo, "activo")
	x.EstadoOcupacion = firstApartTurState(x.EstadoOcupacion, "disponible")
	if x.Capacidad <= 0 {
		x.Capacidad = 2
	}
	if x.Habitaciones <= 0 {
		x.Habitaciones = 1
	}
	if x.Camas <= 0 {
		x.Camas = 1
	}
	if x.Banos <= 0 {
		x.Banos = 1
	}
	return x
}

func normalizeApartTurDates(start, end string) (string, string, int, error) {
	inicio, err := time.Parse("2006-01-02", strings.TrimSpace(start))
	if err != nil {
		return "", "", 0, errors.New("fecha_entrada invalida")
	}
	fin, err := time.Parse("2006-01-02", strings.TrimSpace(end))
	if err != nil {
		return "", "", 0, errors.New("fecha_salida invalida")
	}
	if !fin.After(inicio) {
		return "", "", 0, errors.New("fecha_salida debe ser posterior a fecha_entrada")
	}
	noches := int(math.Ceil(fin.Sub(inicio).Hours() / 24))
	if noches < 1 {
		noches = 1
	}
	return inicio.Format("2006-01-02"), fin.Format("2006-01-02"), noches, nil
}

func resolveApartTurNightPrice(dbConn *sql.DB, empresaID, aptID int64, fallback float64) float64 {
	var price float64
	err := QueryRowCompat(dbConn, `SELECT COALESCE(precio_noche,0) FROM empresa_apartamentos_turisticos_tarifas WHERE empresa_id=? AND estado='activo' AND (apartamento_id=? OR COALESCE(apartamento_id,0)=0) ORDER BY apartamento_id DESC, id DESC LIMIT 1`, empresaID, aptID).Scan(&price)
	if err == nil && price > 0 {
		return price
	}
	return fallback
}

func nextApartTurReservaCode(dbConn *sql.DB, empresaID int64) (string, error) {
	prefix := "APT-" + time.Now().Format("20060102") + "-"
	var count int
	if err := QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_apartamentos_turisticos_reservas WHERE empresa_id=? AND codigo_reserva LIKE ?`, empresaID, prefix+"%").Scan(&count); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%04d", prefix, count+1), nil
}

func generateApartTurAccessCode() string {
	buf := make([]byte, 3)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}
	return strings.ToUpper(hex.EncodeToString(buf))
}

func apartTurBoolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
func firstApartTurState(v, fallback string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return fallback
	}
	return v
}
func maxApartTurInt(v, fallback int) int {
	if v <= 0 {
		return fallback
	}
	return v
}
func roundApartTur(v float64) float64 { return math.Round(v*100) / 100 }

func apartTurCoreCode(prefix string, parts ...string) string {
	var b strings.Builder
	for _, part := range parts {
		for _, r := range strings.ToUpper(strings.TrimSpace(part)) {
			if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				b.WriteRune(r)
				continue
			}
			if b.Len() > 0 && b.String()[b.Len()-1] != '-' {
				b.WriteRune('-')
			}
		}
		if b.Len() > 0 && b.String()[b.Len()-1] != '-' {
			b.WriteRune('-')
		}
	}
	code := strings.Trim(b.String(), "-")
	if code == "" {
		code = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	if len(code) > 42 {
		code = code[:42]
	}
	return strings.Trim(strings.ToUpper(strings.TrimSpace(prefix)), "-") + "-" + strings.Trim(code, "-")
}

func ensureApartTurClienteCore(dbConn *sql.DB, reserva EmpresaApartamentoTuristicoReserva, usuario string) (int64, error) {
	if reserva.ClienteID > 0 {
		return reserva.ClienteID, nil
	}
	if strings.TrimSpace(reserva.HuespedNombre) == "" && strings.TrimSpace(reserva.HuespedDocumento) == "" && strings.TrimSpace(reserva.HuespedTelefono) == "" && strings.TrimSpace(reserva.HuespedEmail) == "" {
		return 0, nil
	}
	if err := EnsureEmpresaClientesSchema(dbConn); err != nil {
		return 0, err
	}
	if documentoNorm := normalizeClienteDocumentoValue(reserva.HuespedDocumento); documentoNorm != "" {
		query := fmt.Sprintf(`SELECT id FROM clientes WHERE empresa_id = ? AND %s = ? LIMIT 1`, clienteDocumentoSQLExpr("numero_documento"))
		if id, err := findClienteDuplicateID(dbConn, query, reserva.EmpresaID, documentoNorm); err != nil {
			return 0, err
		} else if id > 0 {
			return id, nil
		}
	}
	if telefonoNorm := normalizeClienteTelefonoValue(reserva.HuespedTelefono); telefonoNorm != "" {
		query := fmt.Sprintf(`SELECT id FROM clientes WHERE empresa_id = ? AND %s = ? LIMIT 1`, clienteTelefonoSQLExpr("telefono"))
		if id, err := findClienteDuplicateID(dbConn, query, reserva.EmpresaID, telefonoNorm); err != nil {
			return 0, err
		} else if id > 0 {
			return id, nil
		}
	}
	if emailNorm := normalizeClienteEmailValue(reserva.HuespedEmail); emailNorm != "" {
		if id, err := findClienteDuplicateID(dbConn, `SELECT id FROM clientes WHERE empresa_id = ? AND lower(trim(COALESCE(email, ''))) = ? LIMIT 1`, reserva.EmpresaID, emailNorm); err != nil {
			return 0, err
		} else if id > 0 {
			return id, nil
		}
	}
	tipoDocumento := "CC"
	numeroDocumento := strings.TrimSpace(reserva.HuespedDocumento)
	if numeroDocumento == "" {
		tipoDocumento = "OTRO"
		numeroDocumento = apartTurCoreCode("APT-CLI", reserva.HuespedTelefono, reserva.HuespedEmail, reserva.HuespedNombre)
	}
	nombre := strings.TrimSpace(reserva.HuespedNombre)
	if nombre == "" {
		nombre = "Huesped apartamentos " + strings.TrimSpace(reserva.HuespedTelefono)
	}
	id, err := CreateCliente(dbConn, Cliente{
		EmpresaID:         reserva.EmpresaID,
		TipoDocumento:     tipoDocumento,
		NumeroDocumento:   numeroDocumento,
		TipoPersona:       "natural",
		NombreRazonSocial: nombre,
		NombreComercial:   nombre,
		Email:             strings.TrimSpace(reserva.HuespedEmail),
		Telefono:          strings.TrimSpace(reserva.HuespedTelefono),
		Pais:              "CO",
		UsuarioCreador:    strings.TrimSpace(usuario),
		Estado:            "activo",
		Observaciones:     "Cliente creado/sincronizado desde apartamentos turisticos.",
	})
	if err != nil {
		var dup *ClienteDuplicadoError
		if errors.As(err, &dup) && dup.ClienteID > 0 {
			return dup.ClienteID, nil
		}
		return 0, err
	}
	return id, nil
}

func ensureApartTurUnidadServicio(dbConn *sql.DB, unidad EmpresaApartamentoTuristicoUnidad, usuario string) (int64, error) {
	if unidad.ServicioID > 0 {
		return unidad.ServicioID, nil
	}
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		return 0, err
	}
	code := apartTurCoreCode("APT-UNIDAD", unidad.Codigo)
	if strings.TrimSpace(unidad.Codigo) == "" && unidad.ID > 0 {
		code = apartTurCoreCode("APT-UNIDAD", fmt.Sprintf("%d", unidad.ID))
	}
	var servicioID int64
	err := QueryRowCompat(dbConn, `SELECT id FROM servicios WHERE empresa_id=? AND codigo=? LIMIT 1`, unidad.EmpresaID, code).Scan(&servicioID)
	if err == nil {
		_, _ = ExecCompat(dbConn, `UPDATE servicios SET nombre=?, descripcion=?, categoria='apartamentos_turisticos', precio=?, estado='activo', fecha_actualizacion=? WHERE empresa_id=? AND id=?`, strings.TrimSpace(unidad.Nombre), strings.TrimSpace(unidad.Tipo+" "+unidad.Ubicacion), unidad.PrecioBaseNoche, time.Now().Format("2006-01-02 15:04:05"), unidad.EmpresaID, servicioID)
		if unidad.ID > 0 {
			_, _ = ExecCompat(dbConn, `UPDATE empresa_apartamentos_turisticos_unidades SET servicio_id=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, nullableID(servicioID), unidad.EmpresaID, unidad.ID)
		}
		return servicioID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	servicioID, err = CreateServicio(dbConn, Servicio{
		EmpresaID:      unidad.EmpresaID,
		Codigo:         code,
		Nombre:         strings.TrimSpace(unidad.Nombre),
		Descripcion:    strings.TrimSpace("Alojamiento turistico: " + unidad.Tipo + " " + unidad.Ubicacion),
		Categoria:      "apartamentos_turisticos",
		Precio:         unidad.PrecioBaseNoche,
		Estado:         "activo",
		UsuarioCreador: strings.TrimSpace(usuario),
		Observaciones:  "Servicio sincronizado desde apartamentos turisticos.",
	})
	if err != nil {
		return 0, err
	}
	if unidad.ID > 0 {
		_, _ = ExecCompat(dbConn, `UPDATE empresa_apartamentos_turisticos_unidades SET servicio_id=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, nullableID(servicioID), unidad.EmpresaID, unidad.ID)
	}
	return servicioID, nil
}

func ensureApartTurStaticServicio(dbConn *sql.DB, empresaID int64, code, nombre, descripcion string, precio float64, usuario string) (int64, error) {
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		return 0, err
	}
	var servicioID int64
	err := QueryRowCompat(dbConn, `SELECT id FROM servicios WHERE empresa_id=? AND codigo=? LIMIT 1`, empresaID, strings.TrimSpace(code)).Scan(&servicioID)
	if err == nil {
		_, _ = ExecCompat(dbConn, `UPDATE servicios SET nombre=?, descripcion=?, categoria='apartamentos_turisticos', precio=?, estado='activo', fecha_actualizacion=? WHERE empresa_id=? AND id=?`, strings.TrimSpace(nombre), strings.TrimSpace(descripcion), precio, time.Now().Format("2006-01-02 15:04:05"), empresaID, servicioID)
		return servicioID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	return CreateServicio(dbConn, Servicio{
		EmpresaID:      empresaID,
		Codigo:         strings.TrimSpace(code),
		Nombre:         strings.TrimSpace(nombre),
		Descripcion:    strings.TrimSpace(descripcion),
		Categoria:      "apartamentos_turisticos",
		Precio:         precio,
		Estado:         "activo",
		UsuarioCreador: strings.TrimSpace(usuario),
		Observaciones:  "Servicio sincronizado desde apartamentos turisticos.",
	})
}

func createApartTurReservaCarrito(dbConn *sql.DB, reserva EmpresaApartamentoTuristicoReserva, usuario string) (int64, int64, int64, int64, error) {
	if reserva.Total <= 0 {
		return reserva.CarritoID, reserva.CarritoItemID, reserva.ClienteID, reserva.ServicioID, nil
	}
	if err := EnsureEmpresaCarritosSchema(dbConn); err != nil {
		return 0, 0, 0, 0, err
	}
	clienteID, err := ensureApartTurClienteCore(dbConn, reserva, usuario)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	unidad, err := GetEmpresaApartamentoTuristicoUnidad(dbConn, reserva.EmpresaID, reserva.ApartamentoID)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	servicioID, err := ensureApartTurUnidadServicio(dbConn, unidad, usuario)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	metodo := NormalizeMetodoPagoCarrito(reserva.MetodoPago)
	if metodo == "" {
		metodo = "efectivo"
	}
	cfg, _ := GetEmpresaApartamentoTuristicoConfig(dbConn, reserva.EmpresaID)
	referenciaExterna := fmt.Sprintf("apartamentos_turisticos:reserva:%d:%s", reserva.ID, reserva.CodigoReserva)
	var carritoExistente int64
	err = QueryRowCompat(dbConn, `SELECT id FROM carritos_compras WHERE empresa_id=? AND referencia_externa=? LIMIT 1`, reserva.EmpresaID, referenciaExterna).Scan(&carritoExistente)
	if err == nil && carritoExistente > 0 {
		return carritoExistente, reserva.CarritoItemID, clienteID, servicioID, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, 0, 0, 0, err
	}
	carritoID, err := CreateCarritoCompra(dbConn, CarritoCompra{
		EmpresaID:         reserva.EmpresaID,
		Codigo:            apartTurCoreCode("APT-RES", fmt.Sprintf("%d", reserva.ID), reserva.CodigoReserva),
		Nombre:            "Reserva apartamento - " + strings.TrimSpace(reserva.HuespedNombre),
		CanalVenta:        "apartamentos_turisticos",
		ClienteID:         clienteID,
		EstadoCarrito:     "abierto",
		Moneda:            strings.ToUpper(strings.TrimSpace(cfg.Moneda)),
		ReferenciaExterna: referenciaExterna,
		MetodoPago:        metodo,
		ReferenciaPago:    reserva.CodigoReserva,
		UsuarioCreador:    strings.TrimSpace(usuario),
		Observaciones:     "Venta central generada desde reserva de apartamentos turisticos.",
	})
	if err != nil {
		return 0, 0, 0, 0, err
	}
	noches := reserva.Noches
	if noches <= 0 {
		noches = 1
	}
	precioNoche := reserva.Subtotal / float64(noches)
	if precioNoche <= 0 {
		precioNoche = unidad.PrecioBaseNoche
	}
	itemID, err := CreateCarritoCompraItem(dbConn, CarritoCompraItem{
		EmpresaID:          reserva.EmpresaID,
		CarritoID:          carritoID,
		TipoItem:           "servicio",
		ReferenciaID:       servicioID,
		CodigoItem:         apartTurCoreCode("APT-NOCHE", reserva.CodigoReserva),
		Descripcion:        strings.TrimSpace("Alojamiento " + reserva.ApartamentoNombre),
		UnidadMedida:       "noche",
		Cantidad:           float64(noches),
		PrecioUnitario:     precioNoche,
		ImpuestoPorcentaje: cfg.ImpuestoPorcentaje,
		UsuarioCreador:     strings.TrimSpace(usuario),
		Estado:             "activo",
		Observaciones:      fmt.Sprintf("%s a %s", reserva.FechaEntrada, reserva.FechaSalida),
	})
	if err != nil {
		return 0, 0, 0, 0, err
	}
	if reserva.Limpieza > 0 {
		limpiezaID, err := ensureApartTurStaticServicio(dbConn, reserva.EmpresaID, "APT-LIMPIEZA", "Limpieza apartamentos turisticos", "Servicio central para tarifa de limpieza de reservas.", reserva.Limpieza, usuario)
		if err != nil {
			return 0, 0, 0, 0, err
		}
		if _, err := CreateCarritoCompraItem(dbConn, CarritoCompraItem{
			EmpresaID:          reserva.EmpresaID,
			CarritoID:          carritoID,
			TipoItem:           "servicio",
			ReferenciaID:       limpiezaID,
			CodigoItem:         apartTurCoreCode("APT-LIMP", reserva.CodigoReserva),
			Descripcion:        "Limpieza apartamentos turisticos",
			UnidadMedida:       "servicio",
			Cantidad:           1,
			PrecioUnitario:     reserva.Limpieza,
			ImpuestoPorcentaje: cfg.ImpuestoPorcentaje,
			UsuarioCreador:     strings.TrimSpace(usuario),
			Estado:             "activo",
		}); err != nil {
			return 0, 0, 0, 0, err
		}
	}
	if err := PayCarritoStationSession(dbConn, reserva.EmpresaID, carritoID, metodo, reserva.CodigoReserva, "", "", 0, 0, reserva.Total, 0, 0, "", "", 0, strings.TrimSpace(usuario)); err != nil {
		return 0, 0, 0, 0, err
	}
	return carritoID, itemID, clienteID, servicioID, nil
}

func GetEmpresaApartamentoTuristicoReserva(dbConn *sql.DB, empresaID, reservaID int64) (EmpresaApartamentoTuristicoReserva, error) {
	if err := EnsureEmpresaApartamentosTuristicosSchema(dbConn); err != nil {
		return EmpresaApartamentoTuristicoReserva{}, err
	}
	return scanApartTurReserva(QueryRowCompat(dbConn, apartTurReservaSelect()+` WHERE r.empresa_id=? AND r.id=?`, empresaID, reservaID))
}

type apartTurReservaScanner interface {
	Scan(dest ...interface{}) error
}

func apartTurReservaSelect() string {
	return `SELECT r.id,r.empresa_id,r.apartamento_id,COALESCE(r.cliente_id,0),COALESCE(r.servicio_id,0),COALESCE(r.carrito_id,0),COALESCE(r.carrito_item_id,0),COALESCE(u.nombre,''),r.codigo_reserva,r.huesped_nombre,COALESCE(r.huesped_documento,''),COALESCE(r.huesped_telefono,''),COALESCE(r.huesped_email,''),COALESCE(r.cantidad_huespedes,1),r.fecha_entrada,r.fecha_salida,COALESCE(r.noches,1),COALESCE(r.canal,''),COALESCE(r.metodo_pago,''),COALESCE(r.estado_reserva,''),COALESCE(r.estado_pago,''),COALESCE(r.subtotal,0),COALESCE(r.limpieza,0),COALESCE(r.impuestos,0),COALESCE(r.deposito,0),COALESCE(r.total,0),COALESCE(r.saldo_pendiente,0),COALESCE(r.codigo_acceso,''),COALESCE(r.observaciones,''),COALESCE(r.fecha_check_in,''),COALESCE(r.fecha_check_out,''),COALESCE(r.fecha_creacion,''),COALESCE(r.fecha_actualizacion,''),COALESCE(r.usuario_creador,'') FROM empresa_apartamentos_turisticos_reservas r LEFT JOIN empresa_apartamentos_turisticos_unidades u ON u.id=r.apartamento_id AND u.empresa_id=r.empresa_id`
}

func scanApartTurReserva(row apartTurReservaScanner) (EmpresaApartamentoTuristicoReserva, error) {
	var x EmpresaApartamentoTuristicoReserva
	err := row.Scan(&x.ID, &x.EmpresaID, &x.ApartamentoID, &x.ClienteID, &x.ServicioID, &x.CarritoID, &x.CarritoItemID, &x.ApartamentoNombre, &x.CodigoReserva, &x.HuespedNombre, &x.HuespedDocumento, &x.HuespedTelefono, &x.HuespedEmail, &x.CantidadHuespedes, &x.FechaEntrada, &x.FechaSalida, &x.Noches, &x.Canal, &x.MetodoPago, &x.EstadoReserva, &x.EstadoPago, &x.Subtotal, &x.Limpieza, &x.Impuestos, &x.Deposito, &x.Total, &x.SaldoPendiente, &x.CodigoAcceso, &x.Observaciones, &x.FechaCheckIn, &x.FechaCheckOut, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador)
	return x, err
}
