package db

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

var (
	ErrTaxiDriverAuthInvalid   = errors.New("credenciales del conductor invalidas")
	ErrTaxiCustomerAuthInvalid = errors.New("credenciales del cliente invalidas")
	ErrTaxiOfferUnavailable    = errors.New("la oferta ya no esta disponible")
	ErrTaxiRequestUnavailable  = errors.New("la solicitud ya no esta disponible")
)

type EmpresaTaxiConfig struct {
	EmpresaID                  int64   `json:"empresa_id"`
	NombreSistema              string  `json:"nombre_sistema"`
	NombrePortal               string  `json:"nombre_portal"`
	RadioBusquedaKM            float64 `json:"radio_busqueda_km"`
	ConductoresPorRonda        int     `json:"conductores_por_ronda"`
	TimeoutOfertaSegundos      int     `json:"timeout_oferta_segundos"`
	PermitirRegistroCliente    bool    `json:"permitir_registro_cliente"`
	PermitirUbicacionCliente   bool    `json:"permitir_ubicacion_cliente"`
	PermitirDespachoAutomatico bool    `json:"permitir_despacho_automatico"`
	LatitudBase                float64 `json:"latitud_base"`
	LongitudBase               float64 `json:"longitud_base"`
	FechaActualizacion         string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador             string  `json:"usuario_creador,omitempty"`
}

type EmpresaTaxiDriver struct {
	ID                    int64   `json:"id"`
	EmpresaID             int64   `json:"empresa_id"`
	Codigo                string  `json:"codigo"`
	Nombre                string  `json:"nombre"`
	Documento             string  `json:"documento"`
	Telefono              string  `json:"telefono,omitempty"`
	Email                 string  `json:"email,omitempty"`
	VehiculoPlaca         string  `json:"vehiculo_placa,omitempty"`
	VehiculoModelo        string  `json:"vehiculo_modelo,omitempty"`
	VehiculoTipo          string  `json:"vehiculo_tipo,omitempty"`
	VehiculoColor         string  `json:"vehiculo_color,omitempty"`
	LicenciaConduccion    string  `json:"licencia_conduccion,omitempty"`
	GPSDispositivoID      int64   `json:"gps_dispositivo_id,omitempty"`
	GPSCodigo             string  `json:"gps_codigo,omitempty"`
	GPSTipo               string  `json:"gps_tipo,omitempty"`
	GPSProveedor          string  `json:"gps_proveedor,omitempty"`
	GPSProtocolo          string  `json:"gps_protocolo,omitempty"`
	Pin                   string  `json:"pin,omitempty"`
	TokenSesion           string  `json:"token_sesion,omitempty"`
	TokenExpira           string  `json:"token_expira,omitempty"`
	UltimaLatitud         float64 `json:"ultima_latitud,omitempty"`
	UltimaLongitud        float64 `json:"ultima_longitud,omitempty"`
	UltimaPrecisionMetros float64 `json:"ultima_precision_metros,omitempty"`
	UltimaVelocidadKMH    float64 `json:"ultima_velocidad_kmh,omitempty"`
	UltimoReporteEn       string  `json:"ultimo_reporte_en,omitempty"`
	Online                bool    `json:"online"`
	Disponible            bool    `json:"disponible"`
	Estado                string  `json:"estado,omitempty"`
	FechaCreacion         string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion    string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador        string  `json:"usuario_creador,omitempty"`
	Observaciones         string  `json:"observaciones,omitempty"`
}

type EmpresaTaxiCustomer struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Nombre             string `json:"nombre"`
	Documento          string `json:"documento,omitempty"`
	Telefono           string `json:"telefono"`
	Email              string `json:"email,omitempty"`
	Pin                string `json:"pin,omitempty"`
	TokenSesion        string `json:"token_sesion,omitempty"`
	TokenExpira        string `json:"token_expira,omitempty"`
	Estado             string `json:"estado,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
}

type EmpresaTaxiRequest struct {
	ID                       int64   `json:"id"`
	EmpresaID                int64   `json:"empresa_id"`
	CustomerID               int64   `json:"customer_id,omitempty"`
	ConductorID              int64   `json:"conductor_id,omitempty"`
	CodigoServicio           string  `json:"codigo_servicio"`
	ClienteNombre            string  `json:"cliente_nombre"`
	ClienteTelefono          string  `json:"cliente_telefono"`
	ClienteDocumento         string  `json:"cliente_documento,omitempty"`
	RecogerTexto             string  `json:"recoger_texto"`
	RecogerLatitud           float64 `json:"recoger_latitud"`
	RecogerLongitud          float64 `json:"recoger_longitud"`
	DestinoTexto             string  `json:"destino_texto,omitempty"`
	DestinoLatitud           float64 `json:"destino_latitud,omitempty"`
	DestinoLongitud          float64 `json:"destino_longitud,omitempty"`
	ComparteUbicacionCliente bool    `json:"comparte_ubicacion_cliente"`
	MetodoSolicitud          string  `json:"metodo_solicitud,omitempty"`
	Estado                   string  `json:"estado"`
	Canal                    string  `json:"canal,omitempty"`
	Notas                    string  `json:"notas,omitempty"`
	DistanciaEstimadaKM      float64 `json:"distancia_estimada_km,omitempty"`
	TiempoEstimadoMin        float64 `json:"tiempo_estimado_min,omitempty"`
	TarifaEstimada           float64 `json:"tarifa_estimada,omitempty"`
	ConductorNombre          string  `json:"conductor_nombre,omitempty"`
	ConductorTelefono        string  `json:"conductor_telefono,omitempty"`
	VehiculoPlaca            string  `json:"vehiculo_placa,omitempty"`
	VehiculoModelo           string  `json:"vehiculo_modelo,omitempty"`
	FechaSolicitud           string  `json:"fecha_solicitud,omitempty"`
	FechaAceptacion          string  `json:"fecha_aceptacion,omitempty"`
	FechaInicio              string  `json:"fecha_inicio,omitempty"`
	FechaCierre              string  `json:"fecha_cierre,omitempty"`
}

type EmpresaTaxiOffer struct {
	ID               int64   `json:"id"`
	EmpresaID        int64   `json:"empresa_id"`
	RequestID        int64   `json:"request_id"`
	ConductorID      int64   `json:"conductor_id"`
	ConductorNombre  string  `json:"conductor_nombre,omitempty"`
	VehiculoPlaca    string  `json:"vehiculo_placa,omitempty"`
	DistanciaKM      float64 `json:"distancia_km,omitempty"`
	TiempoAproximado float64 `json:"tiempo_aproximado_min,omitempty"`
	Estado           string  `json:"estado"`
	FechaOferta      string  `json:"fecha_oferta,omitempty"`
	FechaRespuesta   string  `json:"fecha_respuesta,omitempty"`
	Observaciones    string  `json:"observaciones,omitempty"`
}

type EmpresaTaxiRoutePoint struct {
	ID              int64   `json:"id"`
	EmpresaID       int64   `json:"empresa_id"`
	RequestID       int64   `json:"request_id,omitempty"`
	ConductorID     int64   `json:"conductor_id,omitempty"`
	ActorTipo       string  `json:"actor_tipo"`
	Latitud         float64 `json:"latitud"`
	Longitud        float64 `json:"longitud"`
	PrecisionMetros float64 `json:"precision_metros,omitempty"`
	VelocidadKMH    float64 `json:"velocidad_kmh,omitempty"`
	RumboGrados     float64 `json:"rumbo_grados,omitempty"`
	CapturadoEn     string  `json:"capturado_en,omitempty"`
}

type EmpresaTaxiDashboard struct {
	EmpresaID              int64                `json:"empresa_id"`
	SolicitudesPendientes  int                  `json:"solicitudes_pendientes"`
	ServiciosActivos       int                  `json:"servicios_activos"`
	ConductoresOnline      int                  `json:"conductores_online"`
	ConductoresDisponibles int                  `json:"conductores_disponibles"`
	ClientesRegistrados    int                  `json:"clientes_registrados"`
	Requests               []EmpresaTaxiRequest `json:"requests"`
	Drivers                []EmpresaTaxiDriver  `json:"drivers"`
	Offers                 []EmpresaTaxiOffer   `json:"offers"`
}

func EnsureEmpresaTaxiSystemSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_taxi_config (
			empresa_id BIGINT PRIMARY KEY,
			nombre_sistema TEXT DEFAULT 'Taxi system',
			nombre_portal TEXT DEFAULT 'Solicita tu servicio',
			radio_busqueda_km NUMERIC(10,2) DEFAULT 7,
			conductores_por_ronda INTEGER DEFAULT 5,
			timeout_oferta_segundos INTEGER DEFAULT 25,
			permitir_registro_cliente INTEGER DEFAULT 1,
			permitir_ubicacion_cliente INTEGER DEFAULT 1,
			permitir_despacho_automatico INTEGER DEFAULT 1,
			latitud_base NUMERIC(12,8) DEFAULT 0,
			longitud_base NUMERIC(12,8) DEFAULT 0,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_taxi_drivers (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			documento TEXT NOT NULL,
			telefono TEXT,
			email TEXT,
			vehiculo_placa TEXT,
			vehiculo_modelo TEXT,
			vehiculo_tipo TEXT,
			vehiculo_color TEXT,
			licencia_conduccion TEXT,
			gps_dispositivo_id BIGINT DEFAULT 0,
			gps_codigo TEXT,
			gps_tipo TEXT DEFAULT 'app_movil',
			gps_proveedor TEXT,
			gps_protocolo TEXT DEFAULT 'app_movil',
			pin_hash TEXT,
			pin_salt TEXT,
			token_sesion TEXT,
			token_expira TEXT,
			ultima_latitud NUMERIC(12,8) DEFAULT 0,
			ultima_longitud NUMERIC(12,8) DEFAULT 0,
			ultima_precision_metros NUMERIC(10,2) DEFAULT 0,
			ultima_velocidad_kmh NUMERIC(10,2) DEFAULT 0,
			ultimo_reporte_en TEXT,
			online INTEGER DEFAULT 0,
			disponible INTEGER DEFAULT 1,
			estado TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			observaciones TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_taxi_driver_codigo ON empresa_taxi_drivers(empresa_id, codigo)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_taxi_driver_documento ON empresa_taxi_drivers(empresa_id, documento)`,
		`CREATE TABLE IF NOT EXISTS empresa_taxi_customers (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			nombre TEXT NOT NULL,
			documento TEXT,
			telefono TEXT NOT NULL,
			email TEXT,
			pin_hash TEXT,
			pin_salt TEXT,
			token_sesion TEXT,
			token_expira TEXT,
			estado TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_taxi_customer_phone ON empresa_taxi_customers(empresa_id, telefono)`,
		`CREATE TABLE IF NOT EXISTS empresa_taxi_requests (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			customer_id BIGINT,
			conductor_id BIGINT,
			codigo_servicio TEXT NOT NULL,
			cliente_nombre TEXT NOT NULL,
			cliente_telefono TEXT NOT NULL,
			cliente_documento TEXT,
			recoger_texto TEXT NOT NULL,
			recoger_latitud NUMERIC(12,8) NOT NULL,
			recoger_longitud NUMERIC(12,8) NOT NULL,
			destino_texto TEXT,
			destino_latitud NUMERIC(12,8) DEFAULT 0,
			destino_longitud NUMERIC(12,8) DEFAULT 0,
			comparte_ubicacion_cliente INTEGER DEFAULT 0,
			metodo_solicitud TEXT DEFAULT 'taxi',
			estado TEXT DEFAULT 'pendiente',
			canal TEXT DEFAULT 'web',
			notas TEXT,
			distancia_estimada_km NUMERIC(10,2) DEFAULT 0,
			tiempo_estimado_min NUMERIC(10,2) DEFAULT 0,
			tarifa_estimada NUMERIC(12,2) DEFAULT 0,
			fecha_solicitud TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_aceptacion TEXT,
			fecha_inicio TEXT,
			fecha_cierre TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_taxi_requests_estado ON empresa_taxi_requests(empresa_id, estado, fecha_solicitud DESC)`,
		`CREATE TABLE IF NOT EXISTS empresa_taxi_offers (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			request_id BIGINT NOT NULL,
			conductor_id BIGINT NOT NULL,
			distancia_km NUMERIC(10,2) DEFAULT 0,
			tiempo_aproximado_min NUMERIC(10,2) DEFAULT 0,
			estado TEXT DEFAULT 'pendiente',
			fecha_oferta TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_respuesta TEXT,
			observaciones TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_taxi_offers_driver_estado ON empresa_taxi_offers(empresa_id, conductor_id, estado, fecha_oferta DESC)`,
		`CREATE TABLE IF NOT EXISTS empresa_taxi_route_points (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			request_id BIGINT,
			conductor_id BIGINT,
			actor_tipo TEXT NOT NULL,
			latitud NUMERIC(12,8) NOT NULL,
			longitud NUMERIC(12,8) NOT NULL,
			precision_metros NUMERIC(10,2) DEFAULT 0,
			velocidad_kmh NUMERIC(10,2) DEFAULT 0,
			rumbo_grados NUMERIC(10,2) DEFAULT 0,
			capturado_en TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_taxi_route_points_req ON empresa_taxi_route_points(empresa_id, request_id, capturado_en DESC)`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	for _, col := range []struct {
		name string
		def  string
	}{
		{"gps_dispositivo_id", "BIGINT DEFAULT 0"},
		{"gps_codigo", "TEXT"},
		{"gps_tipo", "TEXT DEFAULT 'app_movil'"},
		{"gps_proveedor", "TEXT"},
		{"gps_protocolo", "TEXT DEFAULT 'app_movil'"},
	} {
		if err := ensureColumnIfMissing(dbConn, "empresa_taxi_drivers", col.name, col.def); err != nil {
			return err
		}
	}
	return nil
}

func defaultEmpresaTaxiConfig(empresaID int64) EmpresaTaxiConfig {
	return EmpresaTaxiConfig{
		EmpresaID:                  empresaID,
		NombreSistema:              "Taxi system",
		NombrePortal:               "Solicita tu servicio",
		RadioBusquedaKM:            7,
		ConductoresPorRonda:        5,
		TimeoutOfertaSegundos:      25,
		PermitirRegistroCliente:    true,
		PermitirUbicacionCliente:   true,
		PermitirDespachoAutomatico: true,
	}
}

func GetEmpresaTaxiConfig(dbConn *sql.DB, empresaID int64) (EmpresaTaxiConfig, error) {
	if err := EnsureEmpresaTaxiSystemSchema(dbConn); err != nil {
		return EmpresaTaxiConfig{}, err
	}
	cfg := defaultEmpresaTaxiConfig(empresaID)
	var regCli, ubCli, auto int
	err := QueryRowCompat(dbConn, `SELECT empresa_id, COALESCE(nombre_sistema,''), COALESCE(nombre_portal,''), COALESCE(radio_busqueda_km,7), COALESCE(conductores_por_ronda,5), COALESCE(timeout_oferta_segundos,25), COALESCE(permitir_registro_cliente,1), COALESCE(permitir_ubicacion_cliente,1), COALESCE(permitir_despacho_automatico,1), COALESCE(latitud_base,0), COALESCE(longitud_base,0), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'') FROM empresa_taxi_config WHERE empresa_id = ?`, empresaID).Scan(
		&cfg.EmpresaID, &cfg.NombreSistema, &cfg.NombrePortal, &cfg.RadioBusquedaKM, &cfg.ConductoresPorRonda, &cfg.TimeoutOfertaSegundos, &regCli, &ubCli, &auto, &cfg.LatitudBase, &cfg.LongitudBase, &cfg.FechaActualizacion, &cfg.UsuarioCreador,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return cfg, nil
		}
		return EmpresaTaxiConfig{}, err
	}
	cfg.PermitirRegistroCliente = regCli > 0
	cfg.PermitirUbicacionCliente = ubCli > 0
	cfg.PermitirDespachoAutomatico = auto > 0
	return cfg, nil
}

func UpsertEmpresaTaxiConfig(dbConn *sql.DB, cfg EmpresaTaxiConfig) error {
	if err := EnsureEmpresaTaxiSystemSchema(dbConn); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.NombreSistema) == "" {
		cfg.NombreSistema = "Taxi system"
	}
	if strings.TrimSpace(cfg.NombrePortal) == "" {
		cfg.NombrePortal = "Solicita tu servicio"
	}
	if cfg.RadioBusquedaKM <= 0 {
		cfg.RadioBusquedaKM = 7
	}
	if cfg.ConductoresPorRonda <= 0 {
		cfg.ConductoresPorRonda = 5
	}
	if cfg.TimeoutOfertaSegundos <= 0 {
		cfg.TimeoutOfertaSegundos = 25
	}
	_, err := ExecCompat(dbConn, `INSERT INTO empresa_taxi_config (empresa_id, nombre_sistema, nombre_portal, radio_busqueda_km, conductores_por_ronda, timeout_oferta_segundos, permitir_registro_cliente, permitir_ubicacion_cliente, permitir_despacho_automatico, latitud_base, longitud_base, fecha_actualizacion, usuario_creador)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (empresa_id) DO UPDATE SET
			nombre_sistema = EXCLUDED.nombre_sistema,
			nombre_portal = EXCLUDED.nombre_portal,
			radio_busqueda_km = EXCLUDED.radio_busqueda_km,
			conductores_por_ronda = EXCLUDED.conductores_por_ronda,
			timeout_oferta_segundos = EXCLUDED.timeout_oferta_segundos,
			permitir_registro_cliente = EXCLUDED.permitir_registro_cliente,
			permitir_ubicacion_cliente = EXCLUDED.permitir_ubicacion_cliente,
			permitir_despacho_automatico = EXCLUDED.permitir_despacho_automatico,
			latitud_base = EXCLUDED.latitud_base,
			longitud_base = EXCLUDED.longitud_base,
			fecha_actualizacion = EXCLUDED.fecha_actualizacion,
			usuario_creador = EXCLUDED.usuario_creador`,
		cfg.EmpresaID, strings.TrimSpace(cfg.NombreSistema), strings.TrimSpace(cfg.NombrePortal), cfg.RadioBusquedaKM, cfg.ConductoresPorRonda, cfg.TimeoutOfertaSegundos, taxiBoolToInt(cfg.PermitirRegistroCliente), taxiBoolToInt(cfg.PermitirUbicacionCliente), taxiBoolToInt(cfg.PermitirDespachoAutomatico), cfg.LatitudBase, cfg.LongitudBase, time.Now().Format("2006-01-02 15:04:05"), strings.TrimSpace(cfg.UsuarioCreador),
	)
	return err
}

func ListEmpresaTaxiDrivers(dbConn *sql.DB, empresaID int64, onlyOnline bool) ([]EmpresaTaxiDriver, error) {
	if err := EnsureEmpresaTaxiSystemSchema(dbConn); err != nil {
		return nil, err
	}
	query := `SELECT id, empresa_id, COALESCE(codigo,''), COALESCE(nombre,''), COALESCE(documento,''), COALESCE(telefono,''), COALESCE(email,''), COALESCE(vehiculo_placa,''), COALESCE(vehiculo_modelo,''), COALESCE(vehiculo_tipo,''), COALESCE(vehiculo_color,''), COALESCE(licencia_conduccion,''), COALESCE(gps_dispositivo_id,0), COALESCE(gps_codigo,''), COALESCE(gps_tipo,'app_movil'), COALESCE(gps_proveedor,''), COALESCE(gps_protocolo,'app_movil'), COALESCE(token_sesion,''), COALESCE(token_expira,''), COALESCE(ultima_latitud,0), COALESCE(ultima_longitud,0), COALESCE(ultima_precision_metros,0), COALESCE(ultima_velocidad_kmh,0), COALESCE(ultimo_reporte_en,''), COALESCE(online,0), COALESCE(disponible,1), COALESCE(estado,'activo'), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,''), COALESCE(observaciones,'') FROM empresa_taxi_drivers WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	if onlyOnline {
		query += ` AND online = 1 AND estado = 'activo'`
	}
	query += ` ORDER BY online DESC, disponible DESC, nombre ASC`
	rows, err := ExecQueryCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]EmpresaTaxiDriver, 0)
	for rows.Next() {
		var item EmpresaTaxiDriver
		var online, disponible int
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.Codigo, &item.Nombre, &item.Documento, &item.Telefono, &item.Email, &item.VehiculoPlaca, &item.VehiculoModelo, &item.VehiculoTipo, &item.VehiculoColor, &item.LicenciaConduccion, &item.GPSDispositivoID, &item.GPSCodigo, &item.GPSTipo, &item.GPSProveedor, &item.GPSProtocolo, &item.TokenSesion, &item.TokenExpira, &item.UltimaLatitud, &item.UltimaLongitud, &item.UltimaPrecisionMetros, &item.UltimaVelocidadKMH, &item.UltimoReporteEn, &online, &disponible, &item.Estado, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador, &item.Observaciones); err != nil {
			return nil, err
		}
		item.Online = online > 0
		item.Disponible = disponible > 0
		items = append(items, item)
	}
	return items, rows.Err()
}

func CreateEmpresaTaxiDriver(dbConn *sql.DB, item EmpresaTaxiDriver) (int64, error) {
	if err := EnsureEmpresaTaxiSystemSchema(dbConn); err != nil {
		return 0, err
	}
	item.Codigo = strings.ToUpper(strings.TrimSpace(item.Codigo))
	item.Documento = strings.TrimSpace(item.Documento)
	item.Nombre = strings.TrimSpace(item.Nombre)
	if item.Codigo == "" {
		item.Codigo = fmt.Sprintf("DRV-%d", time.Now().Unix()%100000)
	}
	if item.Nombre == "" || item.Documento == "" {
		return 0, fmt.Errorf("nombre y documento son obligatorios")
	}
	salt, hash := hashTaxiPin(item.Pin)
	res, err := ExecCompat(dbConn, `INSERT INTO empresa_taxi_drivers (empresa_id, codigo, nombre, documento, telefono, email, vehiculo_placa, vehiculo_modelo, vehiculo_tipo, vehiculo_color, licencia_conduccion, gps_dispositivo_id, gps_codigo, gps_tipo, gps_proveedor, gps_protocolo, pin_hash, pin_salt, online, disponible, estado, fecha_creacion, fecha_actualizacion, usuario_creador, observaciones)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 1, COALESCE(NULLIF(?, ''), 'activo'), ?, ?, ?, ?)`,
		item.EmpresaID, item.Codigo, item.Nombre, item.Documento, strings.TrimSpace(item.Telefono), strings.TrimSpace(item.Email), strings.ToUpper(strings.TrimSpace(item.VehiculoPlaca)), strings.TrimSpace(item.VehiculoModelo), strings.TrimSpace(item.VehiculoTipo), strings.TrimSpace(item.VehiculoColor), strings.TrimSpace(item.LicenciaConduccion), item.GPSDispositivoID, strings.TrimSpace(item.GPSCodigo), firstTaxiState(item.GPSTipo, "app_movil"), strings.TrimSpace(item.GPSProveedor), firstTaxiState(item.GPSProtocolo, "app_movil"), hash, salt, strings.TrimSpace(item.Estado), time.Now().Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"), strings.TrimSpace(item.UsuarioCreador), strings.TrimSpace(item.Observaciones))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func UpdateEmpresaTaxiDriver(dbConn *sql.DB, item EmpresaTaxiDriver) error {
	if err := EnsureEmpresaTaxiSystemSchema(dbConn); err != nil {
		return err
	}
	if item.ID <= 0 || item.EmpresaID <= 0 {
		return fmt.Errorf("id y empresa_id son obligatorios")
	}
	item.Nombre = strings.TrimSpace(item.Nombre)
	item.Documento = strings.TrimSpace(item.Documento)
	if item.Nombre == "" || item.Documento == "" {
		return fmt.Errorf("nombre y documento son obligatorios")
	}
	query := `UPDATE empresa_taxi_drivers SET codigo = ?, nombre = ?, documento = ?, telefono = ?, email = ?, vehiculo_placa = ?, vehiculo_modelo = ?, vehiculo_tipo = ?, vehiculo_color = ?, licencia_conduccion = ?, gps_dispositivo_id = ?, gps_codigo = ?, gps_tipo = ?, gps_proveedor = ?, gps_protocolo = ?, estado = ?, observaciones = ?, fecha_actualizacion = ?`
	args := []interface{}{strings.ToUpper(strings.TrimSpace(item.Codigo)), item.Nombre, item.Documento, strings.TrimSpace(item.Telefono), strings.TrimSpace(item.Email), strings.ToUpper(strings.TrimSpace(item.VehiculoPlaca)), strings.TrimSpace(item.VehiculoModelo), strings.TrimSpace(item.VehiculoTipo), strings.TrimSpace(item.VehiculoColor), strings.TrimSpace(item.LicenciaConduccion), item.GPSDispositivoID, strings.TrimSpace(item.GPSCodigo), firstTaxiState(item.GPSTipo, "app_movil"), strings.TrimSpace(item.GPSProveedor), firstTaxiState(item.GPSProtocolo, "app_movil"), firstTaxiState(item.Estado, "activo"), strings.TrimSpace(item.Observaciones), time.Now().Format("2006-01-02 15:04:05")}
	if strings.TrimSpace(item.Pin) != "" {
		salt, hash := hashTaxiPin(item.Pin)
		query += `, pin_hash = ?, pin_salt = ?`
		args = append(args, hash, salt)
	}
	query += ` WHERE id = ? AND empresa_id = ?`
	args = append(args, item.ID, item.EmpresaID)
	_, err := ExecCompat(dbConn, query, args...)
	return err
}

func TaxiDriverLogin(dbConn *sql.DB, empresaID int64, documento, pin string) (EmpresaTaxiDriver, error) {
	if err := EnsureEmpresaTaxiSystemSchema(dbConn); err != nil {
		return EmpresaTaxiDriver{}, err
	}
	var item EmpresaTaxiDriver
	var hash, salt string
	var online, disponible int
	err := QueryRowCompat(dbConn, `SELECT id, empresa_id, COALESCE(codigo,''), COALESCE(nombre,''), COALESCE(documento,''), COALESCE(telefono,''), COALESCE(email,''), COALESCE(vehiculo_placa,''), COALESCE(vehiculo_modelo,''), COALESCE(vehiculo_tipo,''), COALESCE(vehiculo_color,''), COALESCE(licencia_conduccion,''), COALESCE(gps_dispositivo_id,0), COALESCE(gps_codigo,''), COALESCE(gps_tipo,'app_movil'), COALESCE(gps_proveedor,''), COALESCE(gps_protocolo,'app_movil'), COALESCE(pin_hash,''), COALESCE(pin_salt,''), COALESCE(online,0), COALESCE(disponible,1), COALESCE(estado,'activo') FROM empresa_taxi_drivers WHERE empresa_id = ? AND documento = ?`, empresaID, strings.TrimSpace(documento)).Scan(
		&item.ID, &item.EmpresaID, &item.Codigo, &item.Nombre, &item.Documento, &item.Telefono, &item.Email, &item.VehiculoPlaca, &item.VehiculoModelo, &item.VehiculoTipo, &item.VehiculoColor, &item.LicenciaConduccion, &item.GPSDispositivoID, &item.GPSCodigo, &item.GPSTipo, &item.GPSProveedor, &item.GPSProtocolo, &hash, &salt, &online, &disponible, &item.Estado,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EmpresaTaxiDriver{}, ErrTaxiDriverAuthInvalid
		}
		return EmpresaTaxiDriver{}, err
	}
	if strings.ToLower(strings.TrimSpace(item.Estado)) != "activo" || !verifyTaxiPin(pin, salt, hash) {
		return EmpresaTaxiDriver{}, ErrTaxiDriverAuthInvalid
	}
	item.TokenSesion = newTaxiToken()
	item.TokenExpira = time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04:05")
	item.Online = online > 0
	item.Disponible = disponible > 0
	_, err = ExecCompat(dbConn, `UPDATE empresa_taxi_drivers SET token_sesion = ?, token_expira = ?, fecha_actualizacion = ? WHERE id = ? AND empresa_id = ?`, item.TokenSesion, item.TokenExpira, time.Now().Format("2006-01-02 15:04:05"), item.ID, item.EmpresaID)
	if err != nil {
		return EmpresaTaxiDriver{}, err
	}
	return item, nil
}

func ResolveTaxiDriverByToken(dbConn *sql.DB, empresaID int64, token string) (EmpresaTaxiDriver, error) {
	var item EmpresaTaxiDriver
	var online, disponible int
	err := QueryRowCompat(dbConn, `SELECT id, empresa_id, COALESCE(codigo,''), COALESCE(nombre,''), COALESCE(documento,''), COALESCE(telefono,''), COALESCE(email,''), COALESCE(vehiculo_placa,''), COALESCE(vehiculo_modelo,''), COALESCE(vehiculo_tipo,''), COALESCE(vehiculo_color,''), COALESCE(licencia_conduccion,''), COALESCE(gps_dispositivo_id,0), COALESCE(gps_codigo,''), COALESCE(gps_tipo,'app_movil'), COALESCE(gps_proveedor,''), COALESCE(gps_protocolo,'app_movil'), COALESCE(token_sesion,''), COALESCE(token_expira,''), COALESCE(ultima_latitud,0), COALESCE(ultima_longitud,0), COALESCE(ultima_precision_metros,0), COALESCE(ultima_velocidad_kmh,0), COALESCE(ultimo_reporte_en,''), COALESCE(online,0), COALESCE(disponible,1), COALESCE(estado,'activo') FROM empresa_taxi_drivers WHERE empresa_id = ? AND token_sesion = ?`, empresaID, strings.TrimSpace(token)).Scan(
		&item.ID, &item.EmpresaID, &item.Codigo, &item.Nombre, &item.Documento, &item.Telefono, &item.Email, &item.VehiculoPlaca, &item.VehiculoModelo, &item.VehiculoTipo, &item.VehiculoColor, &item.LicenciaConduccion, &item.GPSDispositivoID, &item.GPSCodigo, &item.GPSTipo, &item.GPSProveedor, &item.GPSProtocolo, &item.TokenSesion, &item.TokenExpira, &item.UltimaLatitud, &item.UltimaLongitud, &item.UltimaPrecisionMetros, &item.UltimaVelocidadKMH, &item.UltimoReporteEn, &online, &disponible, &item.Estado,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EmpresaTaxiDriver{}, ErrTaxiDriverAuthInvalid
		}
		return EmpresaTaxiDriver{}, err
	}
	if taxiTokenExpired(item.TokenExpira) || strings.ToLower(strings.TrimSpace(item.Estado)) != "activo" {
		return EmpresaTaxiDriver{}, ErrTaxiDriverAuthInvalid
	}
	item.Online = online > 0
	item.Disponible = disponible > 0
	return item, nil
}

func UpdateTaxiDriverPresence(dbConn *sql.DB, empresaID, driverID int64, online, available bool) error {
	_, err := ExecCompat(dbConn, `UPDATE empresa_taxi_drivers SET online = ?, disponible = ?, fecha_actualizacion = ? WHERE empresa_id = ? AND id = ?`, taxiBoolToInt(online), taxiBoolToInt(available), time.Now().Format("2006-01-02 15:04:05"), empresaID, driverID)
	return err
}

func UpdateTaxiDriverLocation(dbConn *sql.DB, empresaID, driverID, requestID int64, point EmpresaTaxiRoutePoint) error {
	if err := EnsureEmpresaTaxiSystemSchema(dbConn); err != nil {
		return err
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := ExecCompat(dbConn, `UPDATE empresa_taxi_drivers SET ultima_latitud = ?, ultima_longitud = ?, ultima_precision_metros = ?, ultima_velocidad_kmh = ?, ultimo_reporte_en = ?, online = 1, fecha_actualizacion = ? WHERE empresa_id = ? AND id = ?`, point.Latitud, point.Longitud, point.PrecisionMetros, point.VelocidadKMH, now, now, empresaID, driverID)
	if err != nil {
		return err
	}
	_, err = ExecCompat(dbConn, `INSERT INTO empresa_taxi_route_points (empresa_id, request_id, conductor_id, actor_tipo, latitud, longitud, precision_metros, velocidad_kmh, rumbo_grados, capturado_en) VALUES (?, ?, ?, 'conductor', ?, ?, ?, ?, ?, ?)`,
		empresaID, requestID, driverID, point.Latitud, point.Longitud, point.PrecisionMetros, point.VelocidadKMH, point.RumboGrados, now)
	return err
}

func RegisterTaxiCustomer(dbConn *sql.DB, item EmpresaTaxiCustomer) (EmpresaTaxiCustomer, error) {
	if err := EnsureEmpresaTaxiSystemSchema(dbConn); err != nil {
		return EmpresaTaxiCustomer{}, err
	}
	item.Nombre = strings.TrimSpace(item.Nombre)
	item.Telefono = strings.TrimSpace(item.Telefono)
	if item.Nombre == "" || item.Telefono == "" || strings.TrimSpace(item.Pin) == "" {
		return EmpresaTaxiCustomer{}, fmt.Errorf("nombre, telefono y pin son obligatorios")
	}
	salt, hash := hashTaxiPin(item.Pin)
	res, err := ExecCompat(dbConn, `INSERT INTO empresa_taxi_customers (empresa_id, nombre, documento, telefono, email, pin_hash, pin_salt, estado, fecha_creacion, fecha_actualizacion) VALUES (?, ?, ?, ?, ?, ?, ?, 'activo', ?, ?)`, item.EmpresaID, item.Nombre, strings.TrimSpace(item.Documento), item.Telefono, strings.TrimSpace(item.Email), hash, salt, time.Now().Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		return EmpresaTaxiCustomer{}, err
	}
	item.ID, _ = res.LastInsertId()
	item.Pin = ""
	return item, nil
}

func TaxiCustomerLogin(dbConn *sql.DB, empresaID int64, telefono, pin string) (EmpresaTaxiCustomer, error) {
	var item EmpresaTaxiCustomer
	var hash, salt string
	err := QueryRowCompat(dbConn, `SELECT id, empresa_id, COALESCE(nombre,''), COALESCE(documento,''), COALESCE(telefono,''), COALESCE(email,''), COALESCE(pin_hash,''), COALESCE(pin_salt,''), COALESCE(estado,'activo') FROM empresa_taxi_customers WHERE empresa_id = ? AND telefono = ?`, empresaID, strings.TrimSpace(telefono)).Scan(&item.ID, &item.EmpresaID, &item.Nombre, &item.Documento, &item.Telefono, &item.Email, &hash, &salt, &item.Estado)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EmpresaTaxiCustomer{}, ErrTaxiCustomerAuthInvalid
		}
		return EmpresaTaxiCustomer{}, err
	}
	if strings.ToLower(strings.TrimSpace(item.Estado)) != "activo" || !verifyTaxiPin(pin, salt, hash) {
		return EmpresaTaxiCustomer{}, ErrTaxiCustomerAuthInvalid
	}
	item.TokenSesion = newTaxiToken()
	item.TokenExpira = time.Now().Add(7 * 24 * time.Hour).Format("2006-01-02 15:04:05")
	_, err = ExecCompat(dbConn, `UPDATE empresa_taxi_customers SET token_sesion = ?, token_expira = ?, fecha_actualizacion = ? WHERE id = ? AND empresa_id = ?`, item.TokenSesion, item.TokenExpira, time.Now().Format("2006-01-02 15:04:05"), item.ID, item.EmpresaID)
	if err != nil {
		return EmpresaTaxiCustomer{}, err
	}
	return item, nil
}

func ResolveTaxiCustomerByToken(dbConn *sql.DB, empresaID int64, token string) (EmpresaTaxiCustomer, error) {
	var item EmpresaTaxiCustomer
	err := QueryRowCompat(dbConn, `SELECT id, empresa_id, COALESCE(nombre,''), COALESCE(documento,''), COALESCE(telefono,''), COALESCE(email,''), COALESCE(token_sesion,''), COALESCE(token_expira,''), COALESCE(estado,'activo') FROM empresa_taxi_customers WHERE empresa_id = ? AND token_sesion = ?`, empresaID, strings.TrimSpace(token)).Scan(&item.ID, &item.EmpresaID, &item.Nombre, &item.Documento, &item.Telefono, &item.Email, &item.TokenSesion, &item.TokenExpira, &item.Estado)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EmpresaTaxiCustomer{}, ErrTaxiCustomerAuthInvalid
		}
		return EmpresaTaxiCustomer{}, err
	}
	if taxiTokenExpired(item.TokenExpira) || strings.ToLower(strings.TrimSpace(item.Estado)) != "activo" {
		return EmpresaTaxiCustomer{}, ErrTaxiCustomerAuthInvalid
	}
	return item, nil
}

func CreateTaxiRequest(dbConn *sql.DB, item EmpresaTaxiRequest) (EmpresaTaxiRequest, error) {
	if err := EnsureEmpresaTaxiSystemSchema(dbConn); err != nil {
		return EmpresaTaxiRequest{}, err
	}
	cfg, err := GetEmpresaTaxiConfig(dbConn, item.EmpresaID)
	if err != nil {
		return EmpresaTaxiRequest{}, err
	}
	if strings.TrimSpace(item.ClienteNombre) == "" || strings.TrimSpace(item.ClienteTelefono) == "" || strings.TrimSpace(item.RecogerTexto) == "" {
		return EmpresaTaxiRequest{}, fmt.Errorf("cliente_nombre, cliente_telefono y recoger_texto son obligatorios")
	}
	item.CodigoServicio = fmt.Sprintf("TX-%s", time.Now().Format("060102-150405"))
	item.Estado = "pendiente"
	item.MetodoSolicitud = firstTaxiState(item.MetodoSolicitud, "taxi")
	item.Canal = firstTaxiState(item.Canal, "web")
	if item.DestinoLatitud != 0 || item.DestinoLongitud != 0 {
		item.DistanciaEstimadaKM = taxiDistanceKM(item.RecogerLatitud, item.RecogerLongitud, item.DestinoLatitud, item.DestinoLongitud)
		item.TiempoEstimadoMin = math.Round((item.DistanciaEstimadaKM/28.0)*60*10) / 10
		item.TarifaEstimada = math.Round((6000+(item.DistanciaEstimadaKM*1900))*100) / 100
	}
	res, err := ExecCompat(dbConn, `INSERT INTO empresa_taxi_requests (empresa_id, customer_id, codigo_servicio, cliente_nombre, cliente_telefono, cliente_documento, recoger_texto, recoger_latitud, recoger_longitud, destino_texto, destino_latitud, destino_longitud, comparte_ubicacion_cliente, metodo_solicitud, estado, canal, notas, distancia_estimada_km, tiempo_estimado_min, tarifa_estimada, fecha_solicitud)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'pendiente', ?, ?, ?, ?, ?, ?)`,
		item.EmpresaID, nullableID(item.CustomerID), item.CodigoServicio, item.ClienteNombre, item.ClienteTelefono, strings.TrimSpace(item.ClienteDocumento), item.RecogerTexto, item.RecogerLatitud, item.RecogerLongitud, strings.TrimSpace(item.DestinoTexto), item.DestinoLatitud, item.DestinoLongitud, taxiBoolToInt(item.ComparteUbicacionCliente), item.MetodoSolicitud, item.Canal, strings.TrimSpace(item.Notas), item.DistanciaEstimadaKM, item.TiempoEstimadoMin, item.TarifaEstimada, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		return EmpresaTaxiRequest{}, err
	}
	item.ID, _ = res.LastInsertId()
	item.FechaSolicitud = time.Now().Format("2006-01-02 15:04:05")
	if cfg.PermitirDespachoAutomatico {
		_, _ = DispatchTaxiRequestToNearbyDrivers(dbConn, item.EmpresaID, item.ID, 0)
	}
	return GetTaxiRequestByID(dbConn, item.EmpresaID, item.ID)
}

func GetTaxiRequestByID(dbConn *sql.DB, empresaID, requestID int64) (EmpresaTaxiRequest, error) {
	var item EmpresaTaxiRequest
	var comparte int
	err := QueryRowCompat(dbConn, `SELECT r.id, r.empresa_id, COALESCE(r.customer_id,0), COALESCE(r.conductor_id,0), COALESCE(r.codigo_servicio,''), COALESCE(r.cliente_nombre,''), COALESCE(r.cliente_telefono,''), COALESCE(r.cliente_documento,''), COALESCE(r.recoger_texto,''), COALESCE(r.recoger_latitud,0), COALESCE(r.recoger_longitud,0), COALESCE(r.destino_texto,''), COALESCE(r.destino_latitud,0), COALESCE(r.destino_longitud,0), COALESCE(r.comparte_ubicacion_cliente,0), COALESCE(r.metodo_solicitud,''), COALESCE(r.estado,'pendiente'), COALESCE(r.canal,'web'), COALESCE(r.notas,''), COALESCE(r.distancia_estimada_km,0), COALESCE(r.tiempo_estimado_min,0), COALESCE(r.tarifa_estimada,0), COALESCE(r.fecha_solicitud,''), COALESCE(r.fecha_aceptacion,''), COALESCE(r.fecha_inicio,''), COALESCE(r.fecha_cierre,''), COALESCE(d.nombre,''), COALESCE(d.telefono,''), COALESCE(d.vehiculo_placa,''), COALESCE(d.vehiculo_modelo,'') FROM empresa_taxi_requests r LEFT JOIN empresa_taxi_drivers d ON d.id = r.conductor_id AND d.empresa_id = r.empresa_id WHERE r.empresa_id = ? AND r.id = ?`, empresaID, requestID).Scan(
		&item.ID, &item.EmpresaID, &item.CustomerID, &item.ConductorID, &item.CodigoServicio, &item.ClienteNombre, &item.ClienteTelefono, &item.ClienteDocumento, &item.RecogerTexto, &item.RecogerLatitud, &item.RecogerLongitud, &item.DestinoTexto, &item.DestinoLatitud, &item.DestinoLongitud, &comparte, &item.MetodoSolicitud, &item.Estado, &item.Canal, &item.Notas, &item.DistanciaEstimadaKM, &item.TiempoEstimadoMin, &item.TarifaEstimada, &item.FechaSolicitud, &item.FechaAceptacion, &item.FechaInicio, &item.FechaCierre, &item.ConductorNombre, &item.ConductorTelefono, &item.VehiculoPlaca, &item.VehiculoModelo,
	)
	if err != nil {
		return EmpresaTaxiRequest{}, err
	}
	item.ComparteUbicacionCliente = comparte > 0
	return item, nil
}

func ListTaxiRequests(dbConn *sql.DB, empresaID int64, estado string, limit int) ([]EmpresaTaxiRequest, error) {
	if err := EnsureEmpresaTaxiSystemSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 100
	}
	query := `SELECT id FROM empresa_taxi_requests WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	if strings.TrimSpace(estado) != "" {
		query += ` AND estado = ?`
		args = append(args, strings.TrimSpace(estado))
	}
	query += ` ORDER BY fecha_solicitud DESC LIMIT ?`
	args = append(args, limit)
	rows, err := ExecQueryCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]EmpresaTaxiRequest, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		item, err := GetTaxiRequestByID(dbConn, empresaID, id)
		if err == nil {
			items = append(items, item)
		}
	}
	return items, rows.Err()
}

func DispatchTaxiRequestToNearbyDrivers(dbConn *sql.DB, empresaID, requestID, actorDriverID int64) ([]EmpresaTaxiOffer, error) {
	if err := EnsureEmpresaTaxiSystemSchema(dbConn); err != nil {
		return nil, err
	}
	req, err := GetTaxiRequestByID(dbConn, empresaID, requestID)
	if err != nil {
		return nil, err
	}
	if !(req.Estado == "pendiente" || req.Estado == "ofertado") {
		return nil, ErrTaxiRequestUnavailable
	}
	cfg, err := GetEmpresaTaxiConfig(dbConn, empresaID)
	if err != nil {
		return nil, err
	}
	_, _ = ExecCompat(dbConn, `UPDATE empresa_taxi_offers SET estado = 'expirada', fecha_respuesta = ?, observaciones = 'redespacho central' WHERE empresa_id = ? AND request_id = ? AND estado = 'pendiente'`, time.Now().Format("2006-01-02 15:04:05"), empresaID, requestID)
	drivers, err := ListEmpresaTaxiDrivers(dbConn, empresaID, true)
	if err != nil {
		return nil, err
	}
	type ranked struct {
		Driver  EmpresaTaxiDriver
		DistKM  float64
		TimeMin float64
	}
	candidates := make([]ranked, 0)
	for _, d := range drivers {
		if !d.Disponible || d.ID == actorDriverID {
			continue
		}
		if d.UltimaLatitud == 0 && d.UltimaLongitud == 0 {
			continue
		}
		dist := taxiDistanceKM(req.RecogerLatitud, req.RecogerLongitud, d.UltimaLatitud, d.UltimaLongitud)
		if cfg.RadioBusquedaKM > 0 && dist > cfg.RadioBusquedaKM {
			continue
		}
		candidates = append(candidates, ranked{Driver: d, DistKM: dist, TimeMin: math.Round((dist/28.0)*60*10) / 10})
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].DistKM < candidates[j].DistKM })
	if len(candidates) > cfg.ConductoresPorRonda {
		candidates = candidates[:cfg.ConductoresPorRonda]
	}
	offers := make([]EmpresaTaxiOffer, 0, len(candidates))
	for _, c := range candidates {
		res, err := ExecCompat(dbConn, `INSERT INTO empresa_taxi_offers (empresa_id, request_id, conductor_id, distancia_km, tiempo_aproximado_min, estado, fecha_oferta) VALUES (?, ?, ?, ?, ?, 'pendiente', ?)`,
			empresaID, requestID, c.Driver.ID, c.DistKM, c.TimeMin, time.Now().Format("2006-01-02 15:04:05"))
		if err != nil {
			return nil, err
		}
		offerID, _ := res.LastInsertId()
		offers = append(offers, EmpresaTaxiOffer{
			ID:               offerID,
			EmpresaID:        empresaID,
			RequestID:        requestID,
			ConductorID:      c.Driver.ID,
			ConductorNombre:  c.Driver.Nombre,
			VehiculoPlaca:    c.Driver.VehiculoPlaca,
			DistanciaKM:      c.DistKM,
			TiempoAproximado: c.TimeMin,
			Estado:           "pendiente",
			FechaOferta:      time.Now().Format("2006-01-02 15:04:05"),
		})
	}
	newState := "pendiente"
	if len(offers) > 0 {
		newState = "ofertado"
	}
	_, _ = ExecCompat(dbConn, `UPDATE empresa_taxi_requests SET estado = ?, fecha_solicitud = COALESCE(fecha_solicitud, ?) WHERE empresa_id = ? AND id = ?`, newState, time.Now().Format("2006-01-02 15:04:05"), empresaID, requestID)
	return offers, nil
}

func ListTaxiOffersForDriver(dbConn *sql.DB, empresaID, driverID int64) ([]EmpresaTaxiOffer, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT o.id, o.empresa_id, o.request_id, o.conductor_id, COALESCE(d.nombre,''), COALESCE(d.vehiculo_placa,''), COALESCE(o.distancia_km,0), COALESCE(o.tiempo_aproximado_min,0), COALESCE(o.estado,'pendiente'), COALESCE(o.fecha_oferta,''), COALESCE(o.fecha_respuesta,''), COALESCE(o.observaciones,'') FROM empresa_taxi_offers o LEFT JOIN empresa_taxi_drivers d ON d.id = o.conductor_id AND d.empresa_id = o.empresa_id WHERE o.empresa_id = ? AND o.conductor_id = ? AND o.estado = 'pendiente' ORDER BY o.fecha_oferta DESC LIMIT 20`, empresaID, driverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]EmpresaTaxiOffer, 0)
	for rows.Next() {
		var item EmpresaTaxiOffer
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.RequestID, &item.ConductorID, &item.ConductorNombre, &item.VehiculoPlaca, &item.DistanciaKM, &item.TiempoAproximado, &item.Estado, &item.FechaOferta, &item.FechaRespuesta, &item.Observaciones); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func RespondTaxiOffer(dbConn *sql.DB, empresaID, offerID, driverID int64, accept bool, observations string) (EmpresaTaxiRequest, error) {
	if err := EnsureEmpresaTaxiSystemSchema(dbConn); err != nil {
		return EmpresaTaxiRequest{}, err
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return EmpresaTaxiRequest{}, err
	}
	defer tx.Rollback()

	var requestID int64
	var currentState string
	if err := queryRowTxSQLCompat(tx, `SELECT request_id, COALESCE(estado,'pendiente') FROM empresa_taxi_offers WHERE empresa_id = ? AND id = ? AND conductor_id = ?`, empresaID, offerID, driverID).Scan(&requestID, &currentState); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EmpresaTaxiRequest{}, ErrTaxiOfferUnavailable
		}
		return EmpresaTaxiRequest{}, err
	}
	if currentState != "pendiente" {
		return EmpresaTaxiRequest{}, ErrTaxiOfferUnavailable
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	if accept {
		var reqState string
		var assigned int64
		if err := queryRowTxSQLCompat(tx, `SELECT COALESCE(estado,'pendiente'), COALESCE(conductor_id,0) FROM empresa_taxi_requests WHERE empresa_id = ? AND id = ?`, empresaID, requestID).Scan(&reqState, &assigned); err != nil {
			return EmpresaTaxiRequest{}, err
		}
		if assigned > 0 || !(reqState == "pendiente" || reqState == "ofertado") {
			if _, err := execTxSQLCompat(tx, `UPDATE empresa_taxi_offers SET estado = 'expirada', fecha_respuesta = ?, observaciones = ? WHERE empresa_id = ? AND id = ?`, now, "solicitud ya asignada", empresaID, offerID); err != nil {
				return EmpresaTaxiRequest{}, err
			}
			return EmpresaTaxiRequest{}, ErrTaxiOfferUnavailable
		}
		if _, err := execTxSQLCompat(tx, `UPDATE empresa_taxi_offers SET estado = 'aceptada', fecha_respuesta = ?, observaciones = ? WHERE empresa_id = ? AND id = ?`, now, strings.TrimSpace(observations), empresaID, offerID); err != nil {
			return EmpresaTaxiRequest{}, err
		}
		if _, err := execTxSQLCompat(tx, `UPDATE empresa_taxi_offers SET estado = 'cerrada', fecha_respuesta = ?, observaciones = 'otra unidad acepto el servicio' WHERE empresa_id = ? AND request_id = ? AND id <> ? AND estado = 'pendiente'`, now, empresaID, requestID, offerID); err != nil {
			return EmpresaTaxiRequest{}, err
		}
		if _, err := execTxSQLCompat(tx, `UPDATE empresa_taxi_requests SET conductor_id = ?, estado = 'aceptada', fecha_aceptacion = ? WHERE empresa_id = ? AND id = ?`, driverID, now, empresaID, requestID); err != nil {
			return EmpresaTaxiRequest{}, err
		}
		if _, err := execTxSQLCompat(tx, `UPDATE empresa_taxi_drivers SET disponible = 0, fecha_actualizacion = ? WHERE empresa_id = ? AND id = ?`, now, empresaID, driverID); err != nil {
			return EmpresaTaxiRequest{}, err
		}
	} else {
		if _, err := execTxSQLCompat(tx, `UPDATE empresa_taxi_offers SET estado = 'rechazada', fecha_respuesta = ?, observaciones = ? WHERE empresa_id = ? AND id = ?`, now, strings.TrimSpace(observations), empresaID, offerID); err != nil {
			return EmpresaTaxiRequest{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return EmpresaTaxiRequest{}, err
	}
	return GetTaxiRequestByID(dbConn, empresaID, requestID)
}

func UpdateTaxiRequestState(dbConn *sql.DB, empresaID, requestID, driverID int64, state, notes string) (EmpresaTaxiRequest, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	switch state {
	case "en_camino":
		_, _ = ExecCompat(dbConn, `UPDATE empresa_taxi_requests SET estado = 'en_camino', notas = ? WHERE empresa_id = ? AND id = ? AND conductor_id = ?`, strings.TrimSpace(notes), empresaID, requestID, driverID)
	case "abordo":
		_, _ = ExecCompat(dbConn, `UPDATE empresa_taxi_requests SET estado = 'abordo', fecha_inicio = ?, notas = ? WHERE empresa_id = ? AND id = ? AND conductor_id = ?`, now, strings.TrimSpace(notes), empresaID, requestID, driverID)
	case "completado":
		_, _ = ExecCompat(dbConn, `UPDATE empresa_taxi_requests SET estado = 'completado', fecha_cierre = ?, notas = ? WHERE empresa_id = ? AND id = ? AND conductor_id = ?`, now, strings.TrimSpace(notes), empresaID, requestID, driverID)
		_, _ = ExecCompat(dbConn, `UPDATE empresa_taxi_drivers SET disponible = 1, fecha_actualizacion = ? WHERE empresa_id = ? AND id = ?`, now, empresaID, driverID)
	case "cancelado":
		_, _ = ExecCompat(dbConn, `UPDATE empresa_taxi_requests SET estado = 'cancelado', fecha_cierre = ?, notas = ? WHERE empresa_id = ? AND id = ?`, now, strings.TrimSpace(notes), empresaID, requestID)
		if driverID > 0 {
			_, _ = ExecCompat(dbConn, `UPDATE empresa_taxi_drivers SET disponible = 1, fecha_actualizacion = ? WHERE empresa_id = ? AND id = ?`, now, empresaID, driverID)
		}
	}
	return GetTaxiRequestByID(dbConn, empresaID, requestID)
}

func ListTaxiRoutePoints(dbConn *sql.DB, empresaID, requestID int64, limit int) ([]EmpresaTaxiRoutePoint, error) {
	if limit <= 0 {
		limit = 300
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT id, empresa_id, COALESCE(request_id,0), COALESCE(conductor_id,0), COALESCE(actor_tipo,''), COALESCE(latitud,0), COALESCE(longitud,0), COALESCE(precision_metros,0), COALESCE(velocidad_kmh,0), COALESCE(rumbo_grados,0), COALESCE(capturado_en,'') FROM empresa_taxi_route_points WHERE empresa_id = ? AND request_id = ? ORDER BY capturado_en ASC LIMIT ?`, empresaID, requestID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EmpresaTaxiRoutePoint, 0)
	for rows.Next() {
		var item EmpresaTaxiRoutePoint
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.RequestID, &item.ConductorID, &item.ActorTipo, &item.Latitud, &item.Longitud, &item.PrecisionMetros, &item.VelocidadKMH, &item.RumboGrados, &item.CapturadoEn); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func AddTaxiCustomerRoutePoint(dbConn *sql.DB, empresaID, requestID, customerID int64, point EmpresaTaxiRoutePoint) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := ExecCompat(dbConn, `INSERT INTO empresa_taxi_route_points (empresa_id, request_id, conductor_id, actor_tipo, latitud, longitud, precision_metros, velocidad_kmh, rumbo_grados, capturado_en) VALUES (?, ?, ?, 'cliente', ?, ?, ?, ?, ?, ?)`, empresaID, requestID, customerID, point.Latitud, point.Longitud, point.PrecisionMetros, point.VelocidadKMH, point.RumboGrados, now)
	return err
}

func BuildEmpresaTaxiDashboard(dbConn *sql.DB, empresaID int64) (EmpresaTaxiDashboard, error) {
	if err := EnsureEmpresaTaxiSystemSchema(dbConn); err != nil {
		return EmpresaTaxiDashboard{}, err
	}
	d := EmpresaTaxiDashboard{EmpresaID: empresaID}
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_taxi_requests WHERE empresa_id = ? AND estado IN ('pendiente','ofertado')`, empresaID).Scan(&d.SolicitudesPendientes)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_taxi_requests WHERE empresa_id = ? AND estado IN ('aceptada','en_camino','abordo')`, empresaID).Scan(&d.ServiciosActivos)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_taxi_drivers WHERE empresa_id = ? AND online = 1 AND estado = 'activo'`, empresaID).Scan(&d.ConductoresOnline)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_taxi_drivers WHERE empresa_id = ? AND online = 1 AND disponible = 1 AND estado = 'activo'`, empresaID).Scan(&d.ConductoresDisponibles)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_taxi_customers WHERE empresa_id = ? AND estado = 'activo'`, empresaID).Scan(&d.ClientesRegistrados)
	d.Requests, _ = ListTaxiRequests(dbConn, empresaID, "", 40)
	d.Drivers, _ = ListEmpresaTaxiDrivers(dbConn, empresaID, false)
	rows, err := ExecQueryCompat(dbConn, `SELECT o.id, o.empresa_id, o.request_id, o.conductor_id, COALESCE(d.nombre,''), COALESCE(d.vehiculo_placa,''), COALESCE(o.distancia_km,0), COALESCE(o.tiempo_aproximado_min,0), COALESCE(o.estado,'pendiente'), COALESCE(o.fecha_oferta,''), COALESCE(o.fecha_respuesta,''), COALESCE(o.observaciones,'') FROM empresa_taxi_offers o LEFT JOIN empresa_taxi_drivers d ON d.id = o.conductor_id AND d.empresa_id = o.empresa_id WHERE o.empresa_id = ? ORDER BY o.fecha_oferta DESC LIMIT 40`, empresaID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var item EmpresaTaxiOffer
			if err := rows.Scan(&item.ID, &item.EmpresaID, &item.RequestID, &item.ConductorID, &item.ConductorNombre, &item.VehiculoPlaca, &item.DistanciaKM, &item.TiempoAproximado, &item.Estado, &item.FechaOferta, &item.FechaRespuesta, &item.Observaciones); err == nil {
				d.Offers = append(d.Offers, item)
			}
		}
	}
	return d, nil
}

func hashTaxiPin(pin string) (string, string) {
	saltBytes := make([]byte, 8)
	_, _ = rand.Read(saltBytes)
	salt := hex.EncodeToString(saltBytes)
	sum := sha256.Sum256([]byte(salt + "::" + strings.TrimSpace(pin)))
	return salt, hex.EncodeToString(sum[:])
}

func verifyTaxiPin(pin, salt, want string) bool {
	if strings.TrimSpace(pin) == "" || strings.TrimSpace(salt) == "" || strings.TrimSpace(want) == "" {
		return false
	}
	sum := sha256.Sum256([]byte(strings.TrimSpace(salt) + "::" + strings.TrimSpace(pin)))
	return hex.EncodeToString(sum[:]) == strings.TrimSpace(want)
}

func taxiDistanceKM(lat1, lng1, lat2, lng2 float64) float64 {
	const r = 6371.0
	latRad1 := lat1 * math.Pi / 180
	latRad2 := lat2 * math.Pi / 180
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(latRad1)*math.Cos(latRad2)*math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return math.Round(r*c*100) / 100
}

func newTaxiToken() string {
	b := make([]byte, 18)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func taxiTokenExpired(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return true
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", raw, time.Local)
	if err != nil {
		return true
	}
	return time.Now().After(t)
}

func taxiBoolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func firstTaxiState(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return strings.TrimSpace(v)
}

func nullableID(v int64) interface{} {
	if v <= 0 {
		return nil
	}
	return v
}
