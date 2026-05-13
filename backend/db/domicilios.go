package db

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

var (
	ErrDomicilioCourierAuthInvalid    = errors.New("credenciales del domiciliario invalidas")
	ErrDomicilioRestaurantAuthInvalid = errors.New("credenciales del restaurante invalidas")
	ErrDomicilioOfferUnavailable      = errors.New("la oferta de domicilio ya no esta disponible")
)

type EmpresaDomiciliosConfig struct {
	EmpresaID                   int64   `json:"empresa_id"`
	NombreSistema               string  `json:"nombre_sistema"`
	NombrePortal                string  `json:"nombre_portal"`
	Moneda                      string  `json:"moneda"`
	RadioCoberturaKM            float64 `json:"radio_cobertura_km"`
	RadioAsignacionKM           float64 `json:"radio_asignacion_km"`
	TarifaBase                  float64 `json:"tarifa_base"`
	TarifaKM                    float64 `json:"tarifa_km"`
	ComisionPorcentaje          float64 `json:"comision_porcentaje"`
	TiempoPreparacionDefaultMin int     `json:"tiempo_preparacion_default_min"`
	DomiciliariosPorRonda       int     `json:"domiciliarios_por_ronda"`
	AutoAsignar                 bool    `json:"auto_asignar"`
	PermitirPedidosPublicos     bool    `json:"permitir_pedidos_publicos"`
	ExigirCodigoEntrega         bool    `json:"exigir_codigo_entrega"`
	LatitudBase                 float64 `json:"latitud_base"`
	LongitudBase                float64 `json:"longitud_base"`
	FechaActualizacion          string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador              string  `json:"usuario_creador,omitempty"`
}

type EmpresaDomicilioRestaurant struct {
	ID                   int64   `json:"id"`
	EmpresaID            int64   `json:"empresa_id"`
	Codigo               string  `json:"codigo"`
	Nombre               string  `json:"nombre"`
	Categoria            string  `json:"categoria,omitempty"`
	Responsable          string  `json:"responsable,omitempty"`
	Telefono             string  `json:"telefono,omitempty"`
	Email                string  `json:"email,omitempty"`
	Direccion            string  `json:"direccion,omitempty"`
	Latitud              float64 `json:"latitud,omitempty"`
	Longitud             float64 `json:"longitud,omitempty"`
	TiempoPreparacionMin int     `json:"tiempo_preparacion_min,omitempty"`
	ComisionPorcentaje   float64 `json:"comision_porcentaje,omitempty"`
	AceptaPedidos        bool    `json:"acepta_pedidos"`
	Estado               string  `json:"estado,omitempty"`
	Pin                  string  `json:"pin,omitempty"`
	TokenSesion          string  `json:"token_sesion,omitempty"`
	TokenExpira          string  `json:"token_expira,omitempty"`
	FechaCreacion        string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion   string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador       string  `json:"usuario_creador,omitempty"`
	Observaciones        string  `json:"observaciones,omitempty"`
}

type EmpresaDomicilioCourier struct {
	ID                    int64   `json:"id"`
	EmpresaID             int64   `json:"empresa_id"`
	Codigo                string  `json:"codigo"`
	Nombre                string  `json:"nombre"`
	Documento             string  `json:"documento"`
	Telefono              string  `json:"telefono,omitempty"`
	Email                 string  `json:"email,omitempty"`
	VehiculoTipo          string  `json:"vehiculo_tipo,omitempty"`
	VehiculoPlaca         string  `json:"vehiculo_placa,omitempty"`
	ZonaBase              string  `json:"zona_base,omitempty"`
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

type EmpresaDomicilioMenuItem struct {
	ID                   int64   `json:"id"`
	EmpresaID            int64   `json:"empresa_id"`
	RestaurantID         int64   `json:"restaurant_id"`
	ServicioID           int64   `json:"servicio_id,omitempty"`
	RestaurantNombre     string  `json:"restaurant_nombre,omitempty"`
	Codigo               string  `json:"codigo,omitempty"`
	Nombre               string  `json:"nombre"`
	Descripcion          string  `json:"descripcion,omitempty"`
	Categoria            string  `json:"categoria,omitempty"`
	Precio               float64 `json:"precio"`
	ImagenURL            string  `json:"imagen_url,omitempty"`
	Disponible           bool    `json:"disponible"`
	TiempoPreparacionMin int     `json:"tiempo_preparacion_min,omitempty"`
	Orden                int     `json:"orden,omitempty"`
	FechaActualizacion   string  `json:"fecha_actualizacion,omitempty"`
}

type EmpresaDomicilioOrderItem struct {
	ID            int64   `json:"id"`
	EmpresaID     int64   `json:"empresa_id,omitempty"`
	OrderID       int64   `json:"order_id,omitempty"`
	MenuItemID    int64   `json:"menu_item_id"`
	ServicioID    int64   `json:"servicio_id,omitempty"`
	CarritoItemID int64   `json:"carrito_item_id,omitempty"`
	Nombre        string  `json:"nombre"`
	Cantidad      float64 `json:"cantidad"`
	PrecioUnit    float64 `json:"precio_unit"`
	Subtotal      float64 `json:"subtotal"`
	Notas         string  `json:"notas,omitempty"`
}

type EmpresaDomicilioOrder struct {
	ID                  int64                       `json:"id"`
	EmpresaID           int64                       `json:"empresa_id"`
	RestaurantID        int64                       `json:"restaurant_id"`
	ClienteID           int64                       `json:"cliente_id,omitempty"`
	CarritoID           int64                       `json:"carrito_id,omitempty"`
	RestaurantNombre    string                      `json:"restaurant_nombre,omitempty"`
	CourierID           int64                       `json:"courier_id,omitempty"`
	CourierNombre       string                      `json:"courier_nombre,omitempty"`
	CourierTelefono     string                      `json:"courier_telefono,omitempty"`
	CodigoPedido        string                      `json:"codigo_pedido"`
	CodigoEntrega       string                      `json:"codigo_entrega,omitempty"`
	TokenCliente        string                      `json:"token_cliente,omitempty"`
	ClienteNombre       string                      `json:"cliente_nombre"`
	ClienteTelefono     string                      `json:"cliente_telefono"`
	ClienteDireccion    string                      `json:"cliente_direccion"`
	ClienteLatitud      float64                     `json:"cliente_latitud,omitempty"`
	ClienteLongitud     float64                     `json:"cliente_longitud,omitempty"`
	MetodoPago          string                      `json:"metodo_pago,omitempty"`
	Estado              string                      `json:"estado"`
	Canal               string                      `json:"canal,omitempty"`
	NotasCliente        string                      `json:"notas_cliente,omitempty"`
	NotasInternas       string                      `json:"notas_internas,omitempty"`
	Subtotal            float64                     `json:"subtotal"`
	TarifaDomicilio     float64                     `json:"tarifa_domicilio"`
	Propina             float64                     `json:"propina"`
	Descuento           float64                     `json:"descuento"`
	Total               float64                     `json:"total"`
	DistanciaEstimadaKM float64                     `json:"distancia_estimada_km,omitempty"`
	TiempoEstimadoMin   float64                     `json:"tiempo_estimado_min,omitempty"`
	FechaPedido         string                      `json:"fecha_pedido,omitempty"`
	FechaConfirmacion   string                      `json:"fecha_confirmacion,omitempty"`
	FechaListo          string                      `json:"fecha_listo,omitempty"`
	FechaRecogida       string                      `json:"fecha_recogida,omitempty"`
	FechaEntrega        string                      `json:"fecha_entrega,omitempty"`
	FechaCancelacion    string                      `json:"fecha_cancelacion,omitempty"`
	Items               []EmpresaDomicilioOrderItem `json:"items,omitempty"`
}

type EmpresaDomicilioOffer struct {
	ID               int64   `json:"id"`
	EmpresaID        int64   `json:"empresa_id"`
	OrderID          int64   `json:"order_id"`
	CourierID        int64   `json:"courier_id"`
	CourierNombre    string  `json:"courier_nombre,omitempty"`
	DistanciaKM      float64 `json:"distancia_km,omitempty"`
	TiempoAproximado float64 `json:"tiempo_aproximado_min,omitempty"`
	Estado           string  `json:"estado"`
	FechaOferta      string  `json:"fecha_oferta,omitempty"`
	FechaRespuesta   string  `json:"fecha_respuesta,omitempty"`
	Observaciones    string  `json:"observaciones,omitempty"`
}

type EmpresaDomicilioTrackPoint struct {
	ID              int64   `json:"id"`
	EmpresaID       int64   `json:"empresa_id"`
	OrderID         int64   `json:"order_id,omitempty"`
	CourierID       int64   `json:"courier_id,omitempty"`
	ActorTipo       string  `json:"actor_tipo"`
	Latitud         float64 `json:"latitud"`
	Longitud        float64 `json:"longitud"`
	PrecisionMetros float64 `json:"precision_metros,omitempty"`
	VelocidadKMH    float64 `json:"velocidad_kmh,omitempty"`
	RumboGrados     float64 `json:"rumbo_grados,omitempty"`
	CapturadoEn     string  `json:"capturado_en,omitempty"`
}

type EmpresaDomiciliosDashboard struct {
	EmpresaID                int64                        `json:"empresa_id"`
	PedidosPendientes        int                          `json:"pedidos_pendientes"`
	PedidosPreparacion       int                          `json:"pedidos_preparacion"`
	PedidosRuta              int                          `json:"pedidos_ruta"`
	DomiciliariosOnline      int                          `json:"domiciliarios_online"`
	DomiciliariosDisponibles int                          `json:"domiciliarios_disponibles"`
	RestaurantesActivos      int                          `json:"restaurantes_activos"`
	VentasHoy                float64                      `json:"ventas_hoy"`
	Orders                   []EmpresaDomicilioOrder      `json:"orders"`
	Couriers                 []EmpresaDomicilioCourier    `json:"couriers"`
	Restaurants              []EmpresaDomicilioRestaurant `json:"restaurants"`
	Offers                   []EmpresaDomicilioOffer      `json:"offers"`
}

func EnsureEmpresaDomiciliosSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_domicilios_config (
			empresa_id BIGINT PRIMARY KEY,
			nombre_sistema TEXT DEFAULT 'Domicilios',
			nombre_portal TEXT DEFAULT 'Pide a domicilio',
			moneda TEXT DEFAULT 'COP',
			radio_cobertura_km NUMERIC(10,2) DEFAULT 8,
			radio_asignacion_km NUMERIC(10,2) DEFAULT 6,
			tarifa_base NUMERIC(12,2) DEFAULT 3500,
			tarifa_km NUMERIC(12,2) DEFAULT 950,
			comision_porcentaje NUMERIC(8,2) DEFAULT 12,
			tiempo_preparacion_default_min INTEGER DEFAULT 20,
			domiciliarios_por_ronda INTEGER DEFAULT 5,
			auto_asignar INTEGER DEFAULT 1,
			permitir_pedidos_publicos INTEGER DEFAULT 1,
			exigir_codigo_entrega INTEGER DEFAULT 1,
			latitud_base NUMERIC(12,8) DEFAULT 4.711,
			longitud_base NUMERIC(12,8) DEFAULT -74.0721,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_domicilios_restaurantes (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			categoria TEXT,
			responsable TEXT,
			telefono TEXT,
			email TEXT,
			direccion TEXT,
			latitud NUMERIC(12,8) DEFAULT 0,
			longitud NUMERIC(12,8) DEFAULT 0,
			tiempo_preparacion_min INTEGER DEFAULT 20,
			comision_porcentaje NUMERIC(8,2) DEFAULT 0,
			acepta_pedidos INTEGER DEFAULT 1,
			estado TEXT DEFAULT 'activo',
			pin_hash TEXT,
			pin_salt TEXT,
			token_sesion TEXT,
			token_expira TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			observaciones TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_domicilios_rest_codigo ON empresa_domicilios_restaurantes(empresa_id, codigo)`,
		`CREATE TABLE IF NOT EXISTS empresa_domicilios_couriers (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			documento TEXT NOT NULL,
			telefono TEXT,
			email TEXT,
			vehiculo_tipo TEXT,
			vehiculo_placa TEXT,
			zona_base TEXT,
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
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_domicilios_courier_codigo ON empresa_domicilios_couriers(empresa_id, codigo)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_domicilios_courier_doc ON empresa_domicilios_couriers(empresa_id, documento)`,
		`CREATE TABLE IF NOT EXISTS empresa_domicilios_menu_items (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			restaurant_id BIGINT NOT NULL,
			servicio_id BIGINT,
			codigo TEXT,
			nombre TEXT NOT NULL,
			descripcion TEXT,
			categoria TEXT,
			precio NUMERIC(12,2) NOT NULL DEFAULT 0,
			imagen_url TEXT,
			disponible INTEGER DEFAULT 1,
			tiempo_preparacion_min INTEGER DEFAULT 0,
			orden INTEGER DEFAULT 0,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_domicilios_menu_rest ON empresa_domicilios_menu_items(empresa_id, restaurant_id, disponible)`,
		`CREATE TABLE IF NOT EXISTS empresa_domicilios_orders (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			restaurant_id BIGINT NOT NULL,
			cliente_id BIGINT,
			carrito_id BIGINT,
			courier_id BIGINT,
			codigo_pedido TEXT NOT NULL,
			codigo_entrega TEXT,
			token_cliente TEXT,
			cliente_nombre TEXT NOT NULL,
			cliente_telefono TEXT NOT NULL,
			cliente_direccion TEXT NOT NULL,
			cliente_latitud NUMERIC(12,8) DEFAULT 0,
			cliente_longitud NUMERIC(12,8) DEFAULT 0,
			metodo_pago TEXT DEFAULT 'efectivo',
			estado TEXT DEFAULT 'nuevo',
			canal TEXT DEFAULT 'web',
			notas_cliente TEXT,
			notas_internas TEXT,
			subtotal NUMERIC(12,2) DEFAULT 0,
			tarifa_domicilio NUMERIC(12,2) DEFAULT 0,
			propina NUMERIC(12,2) DEFAULT 0,
			descuento NUMERIC(12,2) DEFAULT 0,
			total NUMERIC(12,2) DEFAULT 0,
			distancia_estimada_km NUMERIC(10,2) DEFAULT 0,
			tiempo_estimado_min NUMERIC(10,2) DEFAULT 0,
			fecha_pedido TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_confirmacion TEXT,
			fecha_listo TEXT,
			fecha_recogida TEXT,
			fecha_entrega TEXT,
			fecha_cancelacion TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_domicilios_orders_estado ON empresa_domicilios_orders(empresa_id, estado, fecha_pedido DESC)`,
		`CREATE TABLE IF NOT EXISTS empresa_domicilios_order_items (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			order_id BIGINT NOT NULL,
			menu_item_id BIGINT,
			servicio_id BIGINT,
			carrito_item_id BIGINT,
			nombre TEXT NOT NULL,
			cantidad NUMERIC(12,2) NOT NULL DEFAULT 1,
			precio_unit NUMERIC(12,2) NOT NULL DEFAULT 0,
			subtotal NUMERIC(12,2) NOT NULL DEFAULT 0,
			notas TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_domicilios_offers (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			order_id BIGINT NOT NULL,
			courier_id BIGINT NOT NULL,
			distancia_km NUMERIC(10,2) DEFAULT 0,
			tiempo_aproximado_min NUMERIC(10,2) DEFAULT 0,
			estado TEXT DEFAULT 'pendiente',
			fecha_oferta TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_respuesta TEXT,
			observaciones TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_domicilios_offers_courier ON empresa_domicilios_offers(empresa_id, courier_id, estado, fecha_oferta DESC)`,
		`CREATE TABLE IF NOT EXISTS empresa_domicilios_tracking (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			order_id BIGINT,
			courier_id BIGINT,
			actor_tipo TEXT NOT NULL,
			latitud NUMERIC(12,8) NOT NULL,
			longitud NUMERIC(12,8) NOT NULL,
			precision_metros NUMERIC(10,2) DEFAULT 0,
			velocidad_kmh NUMERIC(10,2) DEFAULT 0,
			rumbo_grados NUMERIC(10,2) DEFAULT 0,
			capturado_en TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_domicilios_tracking_order ON empresa_domicilios_tracking(empresa_id, order_id, capturado_en DESC)`,
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
		{"empresa_domicilios_menu_items", "servicio_id", "BIGINT"},
		{"empresa_domicilios_orders", "cliente_id", "BIGINT"},
		{"empresa_domicilios_orders", "carrito_id", "BIGINT"},
		{"empresa_domicilios_order_items", "servicio_id", "BIGINT"},
		{"empresa_domicilios_order_items", "carrito_item_id", "BIGINT"},
	}
	for _, col := range extraColumns {
		if err := ensureColumnIfMissing(dbConn, col.table, col.column, col.def); err != nil {
			return err
		}
	}
	for _, stmt := range []string{
		`CREATE INDEX IF NOT EXISTS ix_empresa_domicilios_orders_carrito ON empresa_domicilios_orders(empresa_id, carrito_id)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_domicilios_items_servicio ON empresa_domicilios_order_items(empresa_id, servicio_id)`,
	} {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func defaultEmpresaDomiciliosConfig(empresaID int64) EmpresaDomiciliosConfig {
	return EmpresaDomiciliosConfig{
		EmpresaID: empresaID, NombreSistema: "Domicilios", NombrePortal: "Pide a domicilio", Moneda: "COP",
		RadioCoberturaKM: 8, RadioAsignacionKM: 6, TarifaBase: 3500, TarifaKM: 950, ComisionPorcentaje: 12,
		TiempoPreparacionDefaultMin: 20, DomiciliariosPorRonda: 5, AutoAsignar: true, PermitirPedidosPublicos: true,
		ExigirCodigoEntrega: true, LatitudBase: 4.711, LongitudBase: -74.0721,
	}
}

func GetEmpresaDomiciliosConfig(dbConn *sql.DB, empresaID int64) (EmpresaDomiciliosConfig, error) {
	if err := EnsureEmpresaDomiciliosSchema(dbConn); err != nil {
		return EmpresaDomiciliosConfig{}, err
	}
	cfg := defaultEmpresaDomiciliosConfig(empresaID)
	var auto, publico, codigo int
	err := QueryRowCompat(dbConn, `SELECT empresa_id, COALESCE(nombre_sistema,''), COALESCE(nombre_portal,''), COALESCE(moneda,'COP'), COALESCE(radio_cobertura_km,8), COALESCE(radio_asignacion_km,6), COALESCE(tarifa_base,3500), COALESCE(tarifa_km,950), COALESCE(comision_porcentaje,12), COALESCE(tiempo_preparacion_default_min,20), COALESCE(domiciliarios_por_ronda,5), COALESCE(auto_asignar,1), COALESCE(permitir_pedidos_publicos,1), COALESCE(exigir_codigo_entrega,1), COALESCE(latitud_base,4.711), COALESCE(longitud_base,-74.0721), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'') FROM empresa_domicilios_config WHERE empresa_id = ?`, empresaID).Scan(
		&cfg.EmpresaID, &cfg.NombreSistema, &cfg.NombrePortal, &cfg.Moneda, &cfg.RadioCoberturaKM, &cfg.RadioAsignacionKM, &cfg.TarifaBase, &cfg.TarifaKM, &cfg.ComisionPorcentaje, &cfg.TiempoPreparacionDefaultMin, &cfg.DomiciliariosPorRonda, &auto, &publico, &codigo, &cfg.LatitudBase, &cfg.LongitudBase, &cfg.FechaActualizacion, &cfg.UsuarioCreador,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return cfg, nil
		}
		return EmpresaDomiciliosConfig{}, err
	}
	cfg.AutoAsignar, cfg.PermitirPedidosPublicos, cfg.ExigirCodigoEntrega = auto > 0, publico > 0, codigo > 0
	return cfg, nil
}

func UpsertEmpresaDomiciliosConfig(dbConn *sql.DB, cfg EmpresaDomiciliosConfig) error {
	if err := EnsureEmpresaDomiciliosSchema(dbConn); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.NombreSistema) == "" {
		cfg.NombreSistema = "Domicilios"
	}
	if strings.TrimSpace(cfg.NombrePortal) == "" {
		cfg.NombrePortal = "Pide a domicilio"
	}
	if strings.TrimSpace(cfg.Moneda) == "" {
		cfg.Moneda = "COP"
	}
	if cfg.RadioCoberturaKM <= 0 {
		cfg.RadioCoberturaKM = 8
	}
	if cfg.RadioAsignacionKM <= 0 {
		cfg.RadioAsignacionKM = 6
	}
	if cfg.TarifaBase < 0 {
		cfg.TarifaBase = 0
	}
	if cfg.TarifaKM < 0 {
		cfg.TarifaKM = 0
	}
	if cfg.TiempoPreparacionDefaultMin <= 0 {
		cfg.TiempoPreparacionDefaultMin = 20
	}
	if cfg.DomiciliariosPorRonda <= 0 {
		cfg.DomiciliariosPorRonda = 5
	}
	_, err := ExecCompat(dbConn, `INSERT INTO empresa_domicilios_config (empresa_id,nombre_sistema,nombre_portal,moneda,radio_cobertura_km,radio_asignacion_km,tarifa_base,tarifa_km,comision_porcentaje,tiempo_preparacion_default_min,domiciliarios_por_ronda,auto_asignar,permitir_pedidos_publicos,exigir_codigo_entrega,latitud_base,longitud_base,fecha_actualizacion,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (empresa_id) DO UPDATE SET nombre_sistema=EXCLUDED.nombre_sistema,nombre_portal=EXCLUDED.nombre_portal,moneda=EXCLUDED.moneda,radio_cobertura_km=EXCLUDED.radio_cobertura_km,radio_asignacion_km=EXCLUDED.radio_asignacion_km,tarifa_base=EXCLUDED.tarifa_base,tarifa_km=EXCLUDED.tarifa_km,comision_porcentaje=EXCLUDED.comision_porcentaje,tiempo_preparacion_default_min=EXCLUDED.tiempo_preparacion_default_min,domiciliarios_por_ronda=EXCLUDED.domiciliarios_por_ronda,auto_asignar=EXCLUDED.auto_asignar,permitir_pedidos_publicos=EXCLUDED.permitir_pedidos_publicos,exigir_codigo_entrega=EXCLUDED.exigir_codigo_entrega,latitud_base=EXCLUDED.latitud_base,longitud_base=EXCLUDED.longitud_base,fecha_actualizacion=EXCLUDED.fecha_actualizacion,usuario_creador=EXCLUDED.usuario_creador`,
		cfg.EmpresaID, strings.TrimSpace(cfg.NombreSistema), strings.TrimSpace(cfg.NombrePortal), strings.ToUpper(strings.TrimSpace(cfg.Moneda)), cfg.RadioCoberturaKM, cfg.RadioAsignacionKM, cfg.TarifaBase, cfg.TarifaKM, cfg.ComisionPorcentaje, cfg.TiempoPreparacionDefaultMin, cfg.DomiciliariosPorRonda, domicilioBoolToInt(cfg.AutoAsignar), domicilioBoolToInt(cfg.PermitirPedidosPublicos), domicilioBoolToInt(cfg.ExigirCodigoEntrega), cfg.LatitudBase, cfg.LongitudBase, nowDomicilio(), strings.TrimSpace(cfg.UsuarioCreador))
	return err
}

func ListDomicilioRestaurants(dbConn *sql.DB, empresaID int64, onlyActive bool) ([]EmpresaDomicilioRestaurant, error) {
	if err := EnsureEmpresaDomiciliosSchema(dbConn); err != nil {
		return nil, err
	}
	q := `SELECT id,empresa_id,COALESCE(codigo,''),COALESCE(nombre,''),COALESCE(categoria,''),COALESCE(responsable,''),COALESCE(telefono,''),COALESCE(email,''),COALESCE(direccion,''),COALESCE(latitud,0),COALESCE(longitud,0),COALESCE(tiempo_preparacion_min,20),COALESCE(comision_porcentaje,0),COALESCE(acepta_pedidos,1),COALESCE(estado,'activo'),COALESCE(token_sesion,''),COALESCE(token_expira,''),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,''),COALESCE(usuario_creador,''),COALESCE(observaciones,'') FROM empresa_domicilios_restaurantes WHERE empresa_id=?`
	if onlyActive {
		q += ` AND estado='activo' AND acepta_pedidos=1`
	}
	q += ` ORDER BY acepta_pedidos DESC, nombre ASC`
	rows, err := ExecQueryCompat(dbConn, q, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaDomicilioRestaurant
	for rows.Next() {
		var x EmpresaDomicilioRestaurant
		var acepta int
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Codigo, &x.Nombre, &x.Categoria, &x.Responsable, &x.Telefono, &x.Email, &x.Direccion, &x.Latitud, &x.Longitud, &x.TiempoPreparacionMin, &x.ComisionPorcentaje, &acepta, &x.Estado, &x.TokenSesion, &x.TokenExpira, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador, &x.Observaciones); err != nil {
			return nil, err
		}
		x.AceptaPedidos = acepta > 0
		out = append(out, x)
	}
	return out, rows.Err()
}

func CreateDomicilioRestaurant(dbConn *sql.DB, x EmpresaDomicilioRestaurant) (int64, error) {
	if err := EnsureEmpresaDomiciliosSchema(dbConn); err != nil {
		return 0, err
	}
	x.Codigo, x.Nombre = strings.ToUpper(strings.TrimSpace(x.Codigo)), strings.TrimSpace(x.Nombre)
	if x.Codigo == "" {
		x.Codigo = fmt.Sprintf("REST-%d", time.Now().Unix()%100000)
	}
	if x.Nombre == "" {
		return 0, fmt.Errorf("nombre del restaurante es obligatorio")
	}
	if x.TiempoPreparacionMin <= 0 {
		x.TiempoPreparacionMin = 20
	}
	salt, hash := hashDomicilioPin(x.Pin)
	return insertSQLCompat(dbConn, `INSERT INTO empresa_domicilios_restaurantes (empresa_id,codigo,nombre,categoria,responsable,telefono,email,direccion,latitud,longitud,tiempo_preparacion_min,comision_porcentaje,acepta_pedidos,estado,pin_hash,pin_salt,fecha_creacion,fecha_actualizacion,usuario_creador,observaciones) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		x.EmpresaID, x.Codigo, x.Nombre, strings.TrimSpace(x.Categoria), strings.TrimSpace(x.Responsable), strings.TrimSpace(x.Telefono), strings.TrimSpace(x.Email), strings.TrimSpace(x.Direccion), x.Latitud, x.Longitud, x.TiempoPreparacionMin, x.ComisionPorcentaje, domicilioBoolToInt(defaultTrueDomicilio(x.AceptaPedidos)), firstDomicilioState(x.Estado, "activo"), hash, salt, nowDomicilio(), nowDomicilio(), strings.TrimSpace(x.UsuarioCreador), strings.TrimSpace(x.Observaciones))
}

func UpdateDomicilioRestaurant(dbConn *sql.DB, x EmpresaDomicilioRestaurant) error {
	if err := EnsureEmpresaDomiciliosSchema(dbConn); err != nil {
		return err
	}
	if x.ID <= 0 || x.EmpresaID <= 0 || strings.TrimSpace(x.Nombre) == "" {
		return fmt.Errorf("id, empresa_id y nombre son obligatorios")
	}
	q := `UPDATE empresa_domicilios_restaurantes SET codigo=?,nombre=?,categoria=?,responsable=?,telefono=?,email=?,direccion=?,latitud=?,longitud=?,tiempo_preparacion_min=?,comision_porcentaje=?,acepta_pedidos=?,estado=?,observaciones=?,fecha_actualizacion=?`
	args := []interface{}{strings.ToUpper(strings.TrimSpace(x.Codigo)), strings.TrimSpace(x.Nombre), strings.TrimSpace(x.Categoria), strings.TrimSpace(x.Responsable), strings.TrimSpace(x.Telefono), strings.TrimSpace(x.Email), strings.TrimSpace(x.Direccion), x.Latitud, x.Longitud, x.TiempoPreparacionMin, x.ComisionPorcentaje, domicilioBoolToInt(x.AceptaPedidos), firstDomicilioState(x.Estado, "activo"), strings.TrimSpace(x.Observaciones), nowDomicilio()}
	if strings.TrimSpace(x.Pin) != "" {
		salt, hash := hashDomicilioPin(x.Pin)
		q += `, pin_hash=?, pin_salt=?`
		args = append(args, hash, salt)
	}
	q += ` WHERE id=? AND empresa_id=?`
	args = append(args, x.ID, x.EmpresaID)
	_, err := ExecCompat(dbConn, q, args...)
	return err
}

func ListDomicilioCouriers(dbConn *sql.DB, empresaID int64, onlyOnline bool) ([]EmpresaDomicilioCourier, error) {
	if err := EnsureEmpresaDomiciliosSchema(dbConn); err != nil {
		return nil, err
	}
	q := `SELECT id,empresa_id,COALESCE(codigo,''),COALESCE(nombre,''),COALESCE(documento,''),COALESCE(telefono,''),COALESCE(email,''),COALESCE(vehiculo_tipo,''),COALESCE(vehiculo_placa,''),COALESCE(zona_base,''),COALESCE(token_sesion,''),COALESCE(token_expira,''),COALESCE(ultima_latitud,0),COALESCE(ultima_longitud,0),COALESCE(ultima_precision_metros,0),COALESCE(ultima_velocidad_kmh,0),COALESCE(ultimo_reporte_en,''),COALESCE(online,0),COALESCE(disponible,1),COALESCE(estado,'activo'),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,''),COALESCE(usuario_creador,''),COALESCE(observaciones,'') FROM empresa_domicilios_couriers WHERE empresa_id=?`
	if onlyOnline {
		q += ` AND online=1 AND estado='activo'`
	}
	q += ` ORDER BY online DESC, disponible DESC, nombre ASC`
	rows, err := ExecQueryCompat(dbConn, q, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaDomicilioCourier
	for rows.Next() {
		var x EmpresaDomicilioCourier
		var online, disponible int
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Codigo, &x.Nombre, &x.Documento, &x.Telefono, &x.Email, &x.VehiculoTipo, &x.VehiculoPlaca, &x.ZonaBase, &x.TokenSesion, &x.TokenExpira, &x.UltimaLatitud, &x.UltimaLongitud, &x.UltimaPrecisionMetros, &x.UltimaVelocidadKMH, &x.UltimoReporteEn, &online, &disponible, &x.Estado, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador, &x.Observaciones); err != nil {
			return nil, err
		}
		x.Online, x.Disponible = online > 0, disponible > 0
		out = append(out, x)
	}
	return out, rows.Err()
}

func CreateDomicilioCourier(dbConn *sql.DB, x EmpresaDomicilioCourier) (int64, error) {
	if err := EnsureEmpresaDomiciliosSchema(dbConn); err != nil {
		return 0, err
	}
	x.Codigo, x.Nombre, x.Documento = strings.ToUpper(strings.TrimSpace(x.Codigo)), strings.TrimSpace(x.Nombre), strings.TrimSpace(x.Documento)
	if x.Codigo == "" {
		x.Codigo = fmt.Sprintf("DOM-%d", time.Now().Unix()%100000)
	}
	if x.Nombre == "" || x.Documento == "" {
		return 0, fmt.Errorf("nombre y documento son obligatorios")
	}
	salt, hash := hashDomicilioPin(x.Pin)
	return insertSQLCompat(dbConn, `INSERT INTO empresa_domicilios_couriers (empresa_id,codigo,nombre,documento,telefono,email,vehiculo_tipo,vehiculo_placa,zona_base,pin_hash,pin_salt,online,disponible,estado,fecha_creacion,fecha_actualizacion,usuario_creador,observaciones) VALUES (?,?,?,?,?,?,?,?,?,?,?,0,1,?,?,?,?,?)`,
		x.EmpresaID, x.Codigo, x.Nombre, x.Documento, strings.TrimSpace(x.Telefono), strings.TrimSpace(x.Email), strings.TrimSpace(x.VehiculoTipo), strings.ToUpper(strings.TrimSpace(x.VehiculoPlaca)), strings.TrimSpace(x.ZonaBase), hash, salt, firstDomicilioState(x.Estado, "activo"), nowDomicilio(), nowDomicilio(), strings.TrimSpace(x.UsuarioCreador), strings.TrimSpace(x.Observaciones))
}

func UpdateDomicilioCourier(dbConn *sql.DB, x EmpresaDomicilioCourier) error {
	if err := EnsureEmpresaDomiciliosSchema(dbConn); err != nil {
		return err
	}
	if x.ID <= 0 || x.EmpresaID <= 0 || strings.TrimSpace(x.Nombre) == "" || strings.TrimSpace(x.Documento) == "" {
		return fmt.Errorf("id, nombre y documento son obligatorios")
	}
	q := `UPDATE empresa_domicilios_couriers SET codigo=?,nombre=?,documento=?,telefono=?,email=?,vehiculo_tipo=?,vehiculo_placa=?,zona_base=?,estado=?,observaciones=?,fecha_actualizacion=?`
	args := []interface{}{strings.ToUpper(strings.TrimSpace(x.Codigo)), strings.TrimSpace(x.Nombre), strings.TrimSpace(x.Documento), strings.TrimSpace(x.Telefono), strings.TrimSpace(x.Email), strings.TrimSpace(x.VehiculoTipo), strings.ToUpper(strings.TrimSpace(x.VehiculoPlaca)), strings.TrimSpace(x.ZonaBase), firstDomicilioState(x.Estado, "activo"), strings.TrimSpace(x.Observaciones), nowDomicilio()}
	if strings.TrimSpace(x.Pin) != "" {
		salt, hash := hashDomicilioPin(x.Pin)
		q += `, pin_hash=?, pin_salt=?`
		args = append(args, hash, salt)
	}
	q += ` WHERE id=? AND empresa_id=?`
	args = append(args, x.ID, x.EmpresaID)
	_, err := ExecCompat(dbConn, q, args...)
	return err
}

func ListDomicilioMenuItems(dbConn *sql.DB, empresaID, restaurantID int64, onlyAvailable bool) ([]EmpresaDomicilioMenuItem, error) {
	if err := EnsureEmpresaDomiciliosSchema(dbConn); err != nil {
		return nil, err
	}
	q := `SELECT m.id,m.empresa_id,m.restaurant_id,COALESCE(m.servicio_id,0),COALESCE(r.nombre,''),COALESCE(m.codigo,''),COALESCE(m.nombre,''),COALESCE(m.descripcion,''),COALESCE(m.categoria,''),COALESCE(m.precio,0),COALESCE(m.imagen_url,''),COALESCE(m.disponible,1),COALESCE(m.tiempo_preparacion_min,0),COALESCE(m.orden,0),COALESCE(m.fecha_actualizacion,'') FROM empresa_domicilios_menu_items m LEFT JOIN empresa_domicilios_restaurantes r ON r.id=m.restaurant_id AND r.empresa_id=m.empresa_id WHERE m.empresa_id=?`
	args := []interface{}{empresaID}
	if restaurantID > 0 {
		q += ` AND m.restaurant_id=?`
		args = append(args, restaurantID)
	}
	if onlyAvailable {
		q += ` AND m.disponible=1`
	}
	q += ` ORDER BY r.nombre ASC, m.categoria ASC, m.orden ASC, m.nombre ASC`
	rows, err := ExecQueryCompat(dbConn, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaDomicilioMenuItem
	for rows.Next() {
		var x EmpresaDomicilioMenuItem
		var disp int
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.RestaurantID, &x.ServicioID, &x.RestaurantNombre, &x.Codigo, &x.Nombre, &x.Descripcion, &x.Categoria, &x.Precio, &x.ImagenURL, &disp, &x.TiempoPreparacionMin, &x.Orden, &x.FechaActualizacion); err != nil {
			return nil, err
		}
		x.Disponible = disp > 0
		out = append(out, x)
	}
	return out, rows.Err()
}

func UpsertDomicilioMenuItem(dbConn *sql.DB, x EmpresaDomicilioMenuItem) (int64, error) {
	if err := EnsureEmpresaDomiciliosSchema(dbConn); err != nil {
		return 0, err
	}
	if x.EmpresaID <= 0 || x.RestaurantID <= 0 || strings.TrimSpace(x.Nombre) == "" {
		return 0, fmt.Errorf("restaurante y nombre son obligatorios")
	}
	servicioID, err := ensureDomicilioMenuServicio(dbConn, x, "domicilios")
	if err != nil {
		return 0, err
	}
	x.ServicioID = servicioID
	if x.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_domicilios_menu_items SET restaurant_id=?,servicio_id=?,codigo=?,nombre=?,descripcion=?,categoria=?,precio=?,imagen_url=?,disponible=?,tiempo_preparacion_min=?,orden=?,fecha_actualizacion=? WHERE id=? AND empresa_id=?`,
			x.RestaurantID, nullableID(x.ServicioID), strings.TrimSpace(x.Codigo), strings.TrimSpace(x.Nombre), strings.TrimSpace(x.Descripcion), strings.TrimSpace(x.Categoria), x.Precio, strings.TrimSpace(x.ImagenURL), domicilioBoolToInt(x.Disponible), x.TiempoPreparacionMin, x.Orden, nowDomicilio(), x.ID, x.EmpresaID)
		return x.ID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_domicilios_menu_items (empresa_id,restaurant_id,servicio_id,codigo,nombre,descripcion,categoria,precio,imagen_url,disponible,tiempo_preparacion_min,orden,fecha_creacion,fecha_actualizacion) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		x.EmpresaID, x.RestaurantID, nullableID(x.ServicioID), strings.TrimSpace(x.Codigo), strings.TrimSpace(x.Nombre), strings.TrimSpace(x.Descripcion), strings.TrimSpace(x.Categoria), x.Precio, strings.TrimSpace(x.ImagenURL), domicilioBoolToInt(defaultTrueDomicilio(x.Disponible)), x.TiempoPreparacionMin, x.Orden, nowDomicilio(), nowDomicilio())
}

func BuildEmpresaDomiciliosDashboard(dbConn *sql.DB, empresaID int64) (EmpresaDomiciliosDashboard, error) {
	if err := EnsureEmpresaDomiciliosSchema(dbConn); err != nil {
		return EmpresaDomiciliosDashboard{}, err
	}
	orders, err := ListDomicilioOrders(dbConn, empresaID, "", 80)
	if err != nil {
		return EmpresaDomiciliosDashboard{}, err
	}
	couriers, err := ListDomicilioCouriers(dbConn, empresaID, false)
	if err != nil {
		return EmpresaDomiciliosDashboard{}, err
	}
	rests, err := ListDomicilioRestaurants(dbConn, empresaID, false)
	if err != nil {
		return EmpresaDomiciliosDashboard{}, err
	}
	offers, _ := ListDomicilioOffers(dbConn, empresaID, 80)
	var dash EmpresaDomiciliosDashboard
	dash.EmpresaID, dash.Orders, dash.Couriers, dash.Restaurants, dash.Offers = empresaID, orders, couriers, rests, offers
	for _, o := range orders {
		switch strings.ToLower(o.Estado) {
		case "nuevo", "confirmado":
			dash.PedidosPendientes++
		case "preparando", "listo":
			dash.PedidosPreparacion++
		case "asignado", "recogido", "en_camino":
			dash.PedidosRuta++
		case "entregado":
			if strings.HasPrefix(o.FechaEntrega, time.Now().Format("2006-01-02")) {
				dash.VentasHoy += o.Total
			}
		}
	}
	for _, c := range couriers {
		if c.Online {
			dash.DomiciliariosOnline++
		}
		if c.Online && c.Disponible {
			dash.DomiciliariosDisponibles++
		}
	}
	for _, r := range rests {
		if r.Estado == "activo" && r.AceptaPedidos {
			dash.RestaurantesActivos++
		}
	}
	return dash, nil
}

func ListDomicilioOrders(dbConn *sql.DB, empresaID int64, estado string, limit int) ([]EmpresaDomicilioOrder, error) {
	if err := EnsureEmpresaDomiciliosSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 300 {
		limit = 100
	}
	q := domicilioOrderSelect() + ` WHERE o.empresa_id=?`
	args := []interface{}{empresaID}
	if strings.TrimSpace(estado) != "" {
		q += ` AND o.estado=?`
		args = append(args, strings.TrimSpace(estado))
	}
	q += fmt.Sprintf(` ORDER BY o.fecha_pedido DESC LIMIT %d`, limit)
	rows, err := ExecQueryCompat(dbConn, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaDomicilioOrder
	for rows.Next() {
		o, err := scanDomicilioOrder(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func GetDomicilioOrderByID(dbConn *sql.DB, empresaID, orderID int64) (EmpresaDomicilioOrder, error) {
	if err := EnsureEmpresaDomiciliosSchema(dbConn); err != nil {
		return EmpresaDomicilioOrder{}, err
	}
	row := QueryRowCompat(dbConn, domicilioOrderSelect()+` WHERE o.empresa_id=? AND o.id=?`, empresaID, orderID)
	o, err := scanDomicilioOrder(row)
	if err != nil {
		return EmpresaDomicilioOrder{}, err
	}
	o.Items, _ = ListDomicilioOrderItems(dbConn, empresaID, orderID)
	return o, nil
}

func GetDomicilioOrderByCustomerToken(dbConn *sql.DB, empresaID, orderID int64, token string) (EmpresaDomicilioOrder, error) {
	if strings.TrimSpace(token) == "" {
		return EmpresaDomicilioOrder{}, sql.ErrNoRows
	}
	row := QueryRowCompat(dbConn, domicilioOrderSelect()+` WHERE o.empresa_id=? AND o.id=? AND o.token_cliente=?`, empresaID, orderID, strings.TrimSpace(token))
	o, err := scanDomicilioOrder(row)
	if err != nil {
		return EmpresaDomicilioOrder{}, err
	}
	o.Items, _ = ListDomicilioOrderItems(dbConn, empresaID, orderID)
	return o, nil
}

func CreateDomicilioOrder(dbConn *sql.DB, order EmpresaDomicilioOrder) (EmpresaDomicilioOrder, error) {
	if err := EnsureEmpresaDomiciliosSchema(dbConn); err != nil {
		return EmpresaDomicilioOrder{}, err
	}
	cfg, _ := GetEmpresaDomiciliosConfig(dbConn, order.EmpresaID)
	if !cfg.PermitirPedidosPublicos && order.Canal == "web" {
		return EmpresaDomicilioOrder{}, fmt.Errorf("los pedidos publicos estan deshabilitados")
	}
	if order.RestaurantID <= 0 || strings.TrimSpace(order.ClienteNombre) == "" || strings.TrimSpace(order.ClienteTelefono) == "" || strings.TrimSpace(order.ClienteDireccion) == "" {
		return EmpresaDomicilioOrder{}, fmt.Errorf("restaurante, cliente, telefono y direccion son obligatorios")
	}
	if len(order.Items) == 0 {
		return EmpresaDomicilioOrder{}, fmt.Errorf("agrega al menos un producto")
	}
	rests, err := ListDomicilioRestaurants(dbConn, order.EmpresaID, true)
	if err != nil {
		return EmpresaDomicilioOrder{}, err
	}
	var rest *EmpresaDomicilioRestaurant
	for i := range rests {
		if rests[i].ID == order.RestaurantID {
			rest = &rests[i]
			break
		}
	}
	if rest == nil {
		return EmpresaDomicilioOrder{}, fmt.Errorf("restaurante no disponible")
	}
	menu, err := ListDomicilioMenuItems(dbConn, order.EmpresaID, order.RestaurantID, true)
	if err != nil {
		return EmpresaDomicilioOrder{}, err
	}
	menuByID := map[int64]EmpresaDomicilioMenuItem{}
	for _, m := range menu {
		menuByID[m.ID] = m
	}
	var subtotal float64
	items := make([]EmpresaDomicilioOrderItem, 0, len(order.Items))
	for _, raw := range order.Items {
		m, ok := menuByID[raw.MenuItemID]
		if !ok {
			return EmpresaDomicilioOrder{}, fmt.Errorf("producto no disponible: %d", raw.MenuItemID)
		}
		qty := raw.Cantidad
		if qty <= 0 {
			qty = 1
		}
		sub := roundDomicilio(qty * m.Precio)
		subtotal += sub
		servicioID, err := ensureDomicilioMenuServicio(dbConn, m, "domicilios")
		if err != nil {
			return EmpresaDomicilioOrder{}, err
		}
		if servicioID > 0 && m.ServicioID <= 0 {
			_, _ = ExecCompat(dbConn, `UPDATE empresa_domicilios_menu_items SET servicio_id=?, fecha_actualizacion=? WHERE empresa_id=? AND id=?`, nullableID(servicioID), nowDomicilio(), order.EmpresaID, m.ID)
		}
		items = append(items, EmpresaDomicilioOrderItem{MenuItemID: m.ID, ServicioID: servicioID, Nombre: m.Nombre, Cantidad: qty, PrecioUnit: m.Precio, Subtotal: sub, Notas: strings.TrimSpace(raw.Notas)})
	}
	dist := haversineDomicilio(rest.Latitud, rest.Longitud, order.ClienteLatitud, order.ClienteLongitud)
	if dist <= 0 {
		dist = order.DistanciaEstimadaKM
	}
	deliveryFee := roundDomicilio(cfg.TarifaBase + math.Max(0, dist)*cfg.TarifaKM)
	order.Subtotal = roundDomicilio(subtotal)
	order.TarifaDomicilio = deliveryFee
	order.Propina = math.Max(0, order.Propina)
	order.Total = roundDomicilio(order.Subtotal + order.TarifaDomicilio + order.Propina - math.Max(0, order.Descuento))
	order.DistanciaEstimadaKM = roundDomicilio(dist)
	order.TiempoEstimadoMin = math.Ceil(float64(rest.TiempoPreparacionMin) + dist*5 + 8)
	order.CodigoPedido = fmt.Sprintf("DOM-%06d", time.Now().UnixNano()%1000000)
	order.CodigoEntrega = randomDigitsDomicilio(4)
	order.TokenCliente = newDomicilioToken()
	if strings.TrimSpace(order.Estado) == "" {
		order.Estado = "nuevo"
	}
	if metodo := NormalizeMetodoPagoCarrito(order.MetodoPago); metodo != "" {
		order.MetodoPago = metodo
	} else {
		order.MetodoPago = "efectivo"
	}
	clienteID, err := ensureDomicilioClienteCore(dbConn, order, "domicilios")
	if err != nil {
		return EmpresaDomicilioOrder{}, err
	}
	order.ClienteID = clienteID
	order.ID, err = insertSQLCompat(dbConn, `INSERT INTO empresa_domicilios_orders (empresa_id,restaurant_id,cliente_id,codigo_pedido,codigo_entrega,token_cliente,cliente_nombre,cliente_telefono,cliente_direccion,cliente_latitud,cliente_longitud,metodo_pago,estado,canal,notas_cliente,subtotal,tarifa_domicilio,propina,descuento,total,distancia_estimada_km,tiempo_estimado_min,fecha_pedido) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		order.EmpresaID, order.RestaurantID, nullableID(order.ClienteID), order.CodigoPedido, order.CodigoEntrega, order.TokenCliente, strings.TrimSpace(order.ClienteNombre), strings.TrimSpace(order.ClienteTelefono), strings.TrimSpace(order.ClienteDireccion), order.ClienteLatitud, order.ClienteLongitud, order.MetodoPago, order.Estado, firstDomicilioState(order.Canal, "web"), strings.TrimSpace(order.NotasCliente), order.Subtotal, order.TarifaDomicilio, order.Propina, order.Descuento, order.Total, order.DistanciaEstimadaKM, order.TiempoEstimadoMin, nowDomicilio())
	if err != nil {
		return EmpresaDomicilioOrder{}, err
	}
	for _, it := range items {
		_, err = ExecCompat(dbConn, `INSERT INTO empresa_domicilios_order_items (empresa_id,order_id,menu_item_id,servicio_id,nombre,cantidad,precio_unit,subtotal,notas) VALUES (?,?,?,?,?,?,?,?,?)`, order.EmpresaID, order.ID, it.MenuItemID, nullableID(it.ServicioID), it.Nombre, it.Cantidad, it.PrecioUnit, it.Subtotal, it.Notas)
		if err != nil {
			return EmpresaDomicilioOrder{}, err
		}
	}
	if cfg.AutoAsignar {
		_, _ = DispatchDomicilioOrder(dbConn, order.EmpresaID, order.ID, 0)
	}
	return GetDomicilioOrderByID(dbConn, order.EmpresaID, order.ID)
}

func UpdateDomicilioOrderState(dbConn *sql.DB, empresaID, orderID, actorID int64, actorTipo, state, notes, codigo string) (EmpresaDomicilioOrder, error) {
	if err := EnsureEmpresaDomiciliosSchema(dbConn); err != nil {
		return EmpresaDomicilioOrder{}, err
	}
	state = strings.ToLower(strings.TrimSpace(state))
	if state == "" {
		return EmpresaDomicilioOrder{}, fmt.Errorf("estado obligatorio")
	}
	if state == "entregado" {
		cfg, _ := GetEmpresaDomiciliosConfig(dbConn, empresaID)
		if cfg.ExigirCodigoEntrega {
			var expected string
			if err := QueryRowCompat(dbConn, `SELECT COALESCE(codigo_entrega,'') FROM empresa_domicilios_orders WHERE empresa_id=? AND id=?`, empresaID, orderID).Scan(&expected); err != nil {
				return EmpresaDomicilioOrder{}, err
			}
			if strings.TrimSpace(codigo) != "" && strings.TrimSpace(expected) != strings.TrimSpace(codigo) {
				return EmpresaDomicilioOrder{}, fmt.Errorf("codigo de entrega incorrecto")
			}
		}
	}
	now := nowDomicilio()
	extra := ""
	switch state {
	case "confirmado":
		extra = `, fecha_confirmacion='` + now + `'`
	case "listo":
		extra = `, fecha_listo='` + now + `'`
	case "recogido":
		extra = `, fecha_recogida='` + now + `'`
	case "entregado":
		extra = `, fecha_entrega='` + now + `'`
	case "cancelado":
		extra = `, fecha_cancelacion='` + now + `'`
	}
	if strings.TrimSpace(notes) != "" {
		extra += `, notas_internas=?`
	}
	args := []interface{}{state}
	if strings.TrimSpace(notes) != "" {
		args = append(args, strings.TrimSpace(notes))
	}
	args = append(args, orderID, empresaID)
	_, err := ExecCompat(dbConn, `UPDATE empresa_domicilios_orders SET estado=?`+extra+` WHERE id=? AND empresa_id=?`, args...)
	if err != nil {
		return EmpresaDomicilioOrder{}, err
	}
	if state == "entregado" || state == "cancelado" {
		_, _ = ExecCompat(dbConn, `UPDATE empresa_domicilios_couriers SET disponible=1 WHERE empresa_id=? AND id=(SELECT courier_id FROM empresa_domicilios_orders WHERE empresa_id=? AND id=?)`, empresaID, empresaID, orderID)
	}
	if state == "entregado" {
		order, err := GetDomicilioOrderByID(dbConn, empresaID, orderID)
		if err != nil {
			return EmpresaDomicilioOrder{}, err
		}
		if order.CarritoID <= 0 && order.Total > 0 {
			carritoID, clienteID, err := createDomicilioOrderCarrito(dbConn, order, "domicilios")
			if err != nil {
				return EmpresaDomicilioOrder{}, err
			}
			_, err = ExecCompat(dbConn, `UPDATE empresa_domicilios_orders SET cliente_id=?, carrito_id=? WHERE empresa_id=? AND id=?`, nullableID(clienteID), nullableID(carritoID), empresaID, orderID)
			if err != nil {
				return EmpresaDomicilioOrder{}, err
			}
		}
	}
	return GetDomicilioOrderByID(dbConn, empresaID, orderID)
}

func DispatchDomicilioOrder(dbConn *sql.DB, empresaID, orderID int64, maxCouriers int) ([]EmpresaDomicilioOffer, error) {
	if err := EnsureEmpresaDomiciliosSchema(dbConn); err != nil {
		return nil, err
	}
	cfg, _ := GetEmpresaDomiciliosConfig(dbConn, empresaID)
	if maxCouriers <= 0 {
		maxCouriers = cfg.DomiciliariosPorRonda
	}
	order, err := GetDomicilioOrderByID(dbConn, empresaID, orderID)
	if err != nil {
		return nil, err
	}
	var rest EmpresaDomicilioRestaurant
	if err := QueryRowCompat(dbConn, `SELECT COALESCE(latitud,0), COALESCE(longitud,0) FROM empresa_domicilios_restaurantes WHERE empresa_id=? AND id=?`, empresaID, order.RestaurantID).Scan(&rest.Latitud, &rest.Longitud); err != nil {
		return nil, err
	}
	couriers, err := ListDomicilioCouriers(dbConn, empresaID, true)
	if err != nil {
		return nil, err
	}
	type cand struct {
		c EmpresaDomicilioCourier
		d float64
	}
	var candidates []cand
	for _, c := range couriers {
		if !c.Disponible || c.UltimaLatitud == 0 || c.UltimaLongitud == 0 {
			continue
		}
		d := haversineDomicilio(rest.Latitud, rest.Longitud, c.UltimaLatitud, c.UltimaLongitud)
		if d <= cfg.RadioAsignacionKM {
			candidates = append(candidates, cand{c, d})
		}
	}
	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].d < candidates[i].d {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}
	_, _ = ExecCompat(dbConn, `UPDATE empresa_domicilios_orders SET estado=CASE WHEN estado='nuevo' THEN 'confirmado' ELSE estado END WHERE empresa_id=? AND id=?`, empresaID, orderID)
	offers := []EmpresaDomicilioOffer{}
	for i, c := range candidates {
		if i >= maxCouriers {
			break
		}
		id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_domicilios_offers (empresa_id,order_id,courier_id,distancia_km,tiempo_aproximado_min,estado,fecha_oferta) VALUES (?,?,?,?,?,?,?)`, empresaID, orderID, c.c.ID, roundDomicilio(c.d), math.Ceil(c.d*4+4), "pendiente", nowDomicilio())
		if err != nil {
			return offers, err
		}
		offers = append(offers, EmpresaDomicilioOffer{ID: id, EmpresaID: empresaID, OrderID: orderID, CourierID: c.c.ID, CourierNombre: c.c.Nombre, DistanciaKM: roundDomicilio(c.d), TiempoAproximado: math.Ceil(c.d*4 + 4), Estado: "pendiente", FechaOferta: nowDomicilio()})
	}
	return offers, nil
}

func ListDomicilioOffers(dbConn *sql.DB, empresaID int64, limit int) ([]EmpresaDomicilioOffer, error) {
	if limit <= 0 || limit > 300 {
		limit = 100
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT o.id,o.empresa_id,o.order_id,o.courier_id,COALESCE(c.nombre,''),COALESCE(o.distancia_km,0),COALESCE(o.tiempo_aproximado_min,0),COALESCE(o.estado,''),COALESCE(o.fecha_oferta,''),COALESCE(o.fecha_respuesta,''),COALESCE(o.observaciones,'') FROM empresa_domicilios_offers o LEFT JOIN empresa_domicilios_couriers c ON c.id=o.courier_id AND c.empresa_id=o.empresa_id WHERE o.empresa_id=? ORDER BY o.fecha_oferta DESC LIMIT %d`, limit), empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaDomicilioOffer
	for rows.Next() {
		var x EmpresaDomicilioOffer
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.OrderID, &x.CourierID, &x.CourierNombre, &x.DistanciaKM, &x.TiempoAproximado, &x.Estado, &x.FechaOferta, &x.FechaRespuesta, &x.Observaciones); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func ListDomicilioOffersForCourier(dbConn *sql.DB, empresaID, courierID int64) ([]EmpresaDomicilioOffer, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT o.id,o.empresa_id,o.order_id,o.courier_id,COALESCE(c.nombre,''),COALESCE(o.distancia_km,0),COALESCE(o.tiempo_aproximado_min,0),COALESCE(o.estado,''),COALESCE(o.fecha_oferta,''),COALESCE(o.fecha_respuesta,''),COALESCE(o.observaciones,'') FROM empresa_domicilios_offers o LEFT JOIN empresa_domicilios_couriers c ON c.id=o.courier_id AND c.empresa_id=o.empresa_id WHERE o.empresa_id=? AND o.courier_id=? AND o.estado='pendiente' ORDER BY o.fecha_oferta DESC LIMIT 30`, empresaID, courierID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaDomicilioOffer
	for rows.Next() {
		var x EmpresaDomicilioOffer
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.OrderID, &x.CourierID, &x.CourierNombre, &x.DistanciaKM, &x.TiempoAproximado, &x.Estado, &x.FechaOferta, &x.FechaRespuesta, &x.Observaciones); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func RespondDomicilioOffer(dbConn *sql.DB, empresaID, offerID, courierID int64, accept bool, obs string) (EmpresaDomicilioOrder, error) {
	if err := EnsureEmpresaDomiciliosSchema(dbConn); err != nil {
		return EmpresaDomicilioOrder{}, err
	}
	var orderID int64
	var state string
	if err := QueryRowCompat(dbConn, `SELECT order_id, COALESCE(estado,'') FROM empresa_domicilios_offers WHERE empresa_id=? AND id=? AND courier_id=?`, empresaID, offerID, courierID).Scan(&orderID, &state); err != nil {
		return EmpresaDomicilioOrder{}, err
	}
	if state != "pendiente" {
		return EmpresaDomicilioOrder{}, ErrDomicilioOfferUnavailable
	}
	now := nowDomicilio()
	if !accept {
		_, err := ExecCompat(dbConn, `UPDATE empresa_domicilios_offers SET estado='rechazada', fecha_respuesta=?, observaciones=? WHERE empresa_id=? AND id=? AND courier_id=?`, now, strings.TrimSpace(obs), empresaID, offerID, courierID)
		if err != nil {
			return EmpresaDomicilioOrder{}, err
		}
		return GetDomicilioOrderByID(dbConn, empresaID, orderID)
	}
	_, err := ExecCompat(dbConn, `UPDATE empresa_domicilios_offers SET estado='aceptada', fecha_respuesta=?, observaciones=? WHERE empresa_id=? AND id=? AND courier_id=?`, now, strings.TrimSpace(obs), empresaID, offerID, courierID)
	if err != nil {
		return EmpresaDomicilioOrder{}, err
	}
	_, _ = ExecCompat(dbConn, `UPDATE empresa_domicilios_offers SET estado='expirada', fecha_respuesta=? WHERE empresa_id=? AND order_id=? AND id<>? AND estado='pendiente'`, now, empresaID, orderID, offerID)
	_, err = ExecCompat(dbConn, `UPDATE empresa_domicilios_orders SET courier_id=?, estado='asignado', fecha_confirmacion=COALESCE(fecha_confirmacion, ?) WHERE empresa_id=? AND id=?`, courierID, now, empresaID, orderID)
	if err != nil {
		return EmpresaDomicilioOrder{}, err
	}
	_, _ = ExecCompat(dbConn, `UPDATE empresa_domicilios_couriers SET disponible=0 WHERE empresa_id=? AND id=?`, empresaID, courierID)
	return GetDomicilioOrderByID(dbConn, empresaID, orderID)
}

func DomicilioCourierLogin(dbConn *sql.DB, empresaID int64, documento, pin string) (EmpresaDomicilioCourier, error) {
	var x EmpresaDomicilioCourier
	var hash, salt string
	var online, disponible int
	err := QueryRowCompat(dbConn, `SELECT id,empresa_id,COALESCE(codigo,''),COALESCE(nombre,''),COALESCE(documento,''),COALESCE(telefono,''),COALESCE(email,''),COALESCE(vehiculo_tipo,''),COALESCE(vehiculo_placa,''),COALESCE(zona_base,''),COALESCE(pin_hash,''),COALESCE(pin_salt,''),COALESCE(online,0),COALESCE(disponible,1),COALESCE(estado,'activo') FROM empresa_domicilios_couriers WHERE empresa_id=? AND documento=?`, empresaID, strings.TrimSpace(documento)).Scan(&x.ID, &x.EmpresaID, &x.Codigo, &x.Nombre, &x.Documento, &x.Telefono, &x.Email, &x.VehiculoTipo, &x.VehiculoPlaca, &x.ZonaBase, &hash, &salt, &online, &disponible, &x.Estado)
	if err != nil || strings.ToLower(x.Estado) != "activo" || !verifyDomicilioPin(pin, salt, hash) {
		return EmpresaDomicilioCourier{}, ErrDomicilioCourierAuthInvalid
	}
	x.TokenSesion, x.TokenExpira, x.Online, x.Disponible = newDomicilioToken(), time.Now().Add(24*time.Hour).Format("2006-01-02 15:04:05"), online > 0, disponible > 0
	_, err = ExecCompat(dbConn, `UPDATE empresa_domicilios_couriers SET token_sesion=?, token_expira=?, fecha_actualizacion=? WHERE empresa_id=? AND id=?`, x.TokenSesion, x.TokenExpira, nowDomicilio(), empresaID, x.ID)
	return x, err
}

func DomicilioRestaurantLogin(dbConn *sql.DB, empresaID int64, codigo, pin string) (EmpresaDomicilioRestaurant, error) {
	var x EmpresaDomicilioRestaurant
	var hash, salt string
	var acepta int
	err := QueryRowCompat(dbConn, `SELECT id,empresa_id,COALESCE(codigo,''),COALESCE(nombre,''),COALESCE(categoria,''),COALESCE(responsable,''),COALESCE(telefono,''),COALESCE(email,''),COALESCE(direccion,''),COALESCE(latitud,0),COALESCE(longitud,0),COALESCE(tiempo_preparacion_min,20),COALESCE(comision_porcentaje,0),COALESCE(acepta_pedidos,1),COALESCE(estado,'activo'),COALESCE(pin_hash,''),COALESCE(pin_salt,'') FROM empresa_domicilios_restaurantes WHERE empresa_id=? AND codigo=?`, empresaID, strings.ToUpper(strings.TrimSpace(codigo))).Scan(&x.ID, &x.EmpresaID, &x.Codigo, &x.Nombre, &x.Categoria, &x.Responsable, &x.Telefono, &x.Email, &x.Direccion, &x.Latitud, &x.Longitud, &x.TiempoPreparacionMin, &x.ComisionPorcentaje, &acepta, &x.Estado, &hash, &salt)
	if err != nil || strings.ToLower(x.Estado) != "activo" || !verifyDomicilioPin(pin, salt, hash) {
		return EmpresaDomicilioRestaurant{}, ErrDomicilioRestaurantAuthInvalid
	}
	x.TokenSesion, x.TokenExpira, x.AceptaPedidos = newDomicilioToken(), time.Now().Add(24*time.Hour).Format("2006-01-02 15:04:05"), acepta > 0
	_, err = ExecCompat(dbConn, `UPDATE empresa_domicilios_restaurantes SET token_sesion=?, token_expira=?, fecha_actualizacion=? WHERE empresa_id=? AND id=?`, x.TokenSesion, x.TokenExpira, nowDomicilio(), empresaID, x.ID)
	return x, err
}

func ResolveDomicilioCourierByToken(dbConn *sql.DB, empresaID int64, token string) (EmpresaDomicilioCourier, error) {
	var x EmpresaDomicilioCourier
	var online, disponible int
	err := QueryRowCompat(dbConn, `SELECT id,empresa_id,COALESCE(codigo,''),COALESCE(nombre,''),COALESCE(documento,''),COALESCE(telefono,''),COALESCE(email,''),COALESCE(vehiculo_tipo,''),COALESCE(vehiculo_placa,''),COALESCE(zona_base,''),COALESCE(token_sesion,''),COALESCE(token_expira,''),COALESCE(ultima_latitud,0),COALESCE(ultima_longitud,0),COALESCE(online,0),COALESCE(disponible,1),COALESCE(estado,'activo') FROM empresa_domicilios_couriers WHERE empresa_id=? AND token_sesion=?`, empresaID, strings.TrimSpace(token)).Scan(&x.ID, &x.EmpresaID, &x.Codigo, &x.Nombre, &x.Documento, &x.Telefono, &x.Email, &x.VehiculoTipo, &x.VehiculoPlaca, &x.ZonaBase, &x.TokenSesion, &x.TokenExpira, &x.UltimaLatitud, &x.UltimaLongitud, &online, &disponible, &x.Estado)
	if err != nil {
		return EmpresaDomicilioCourier{}, err
	}
	if exp, err := time.Parse("2006-01-02 15:04:05", x.TokenExpira); err == nil && time.Now().After(exp) {
		return EmpresaDomicilioCourier{}, sql.ErrNoRows
	}
	x.Online, x.Disponible = online > 0, disponible > 0
	return x, nil
}

func ResolveDomicilioRestaurantByToken(dbConn *sql.DB, empresaID int64, token string) (EmpresaDomicilioRestaurant, error) {
	var x EmpresaDomicilioRestaurant
	var acepta int
	err := QueryRowCompat(dbConn, `SELECT id,empresa_id,COALESCE(codigo,''),COALESCE(nombre,''),COALESCE(acepta_pedidos,1),COALESCE(estado,'activo'),COALESCE(token_sesion,''),COALESCE(token_expira,'') FROM empresa_domicilios_restaurantes WHERE empresa_id=? AND token_sesion=?`, empresaID, strings.TrimSpace(token)).Scan(&x.ID, &x.EmpresaID, &x.Codigo, &x.Nombre, &acepta, &x.Estado, &x.TokenSesion, &x.TokenExpira)
	if err != nil {
		return EmpresaDomicilioRestaurant{}, err
	}
	x.AceptaPedidos = acepta > 0
	return x, nil
}

func UpdateDomicilioCourierPresence(dbConn *sql.DB, empresaID, courierID int64, online, disponible bool) error {
	_, err := ExecCompat(dbConn, `UPDATE empresa_domicilios_couriers SET online=?, disponible=?, fecha_actualizacion=? WHERE empresa_id=? AND id=?`, domicilioBoolToInt(online), domicilioBoolToInt(disponible), nowDomicilio(), empresaID, courierID)
	return err
}

func UpdateDomicilioCourierLocation(dbConn *sql.DB, empresaID, courierID, orderID int64, p EmpresaDomicilioTrackPoint) error {
	if p.Latitud == 0 && p.Longitud == 0 {
		return fmt.Errorf("latitud y longitud son obligatorias")
	}
	now := nowDomicilio()
	_, err := ExecCompat(dbConn, `UPDATE empresa_domicilios_couriers SET ultima_latitud=?, ultima_longitud=?, ultima_precision_metros=?, ultima_velocidad_kmh=?, ultimo_reporte_en=?, online=1, fecha_actualizacion=? WHERE empresa_id=? AND id=?`, p.Latitud, p.Longitud, p.PrecisionMetros, p.VelocidadKMH, now, now, empresaID, courierID)
	if err != nil {
		return err
	}
	_, err = ExecCompat(dbConn, `INSERT INTO empresa_domicilios_tracking (empresa_id,order_id,courier_id,actor_tipo,latitud,longitud,precision_metros,velocidad_kmh,rumbo_grados,capturado_en) VALUES (?,?,?,?,?,?,?,?,?,?)`, empresaID, orderID, courierID, "domiciliario", p.Latitud, p.Longitud, p.PrecisionMetros, p.VelocidadKMH, p.RumboGrados, now)
	return err
}

func ListDomicilioTracking(dbConn *sql.DB, empresaID, orderID int64, limit int) ([]EmpresaDomicilioTrackPoint, error) {
	if limit <= 0 || limit > 1000 {
		limit = 500
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,COALESCE(order_id,0),COALESCE(courier_id,0),COALESCE(actor_tipo,''),COALESCE(latitud,0),COALESCE(longitud,0),COALESCE(precision_metros,0),COALESCE(velocidad_kmh,0),COALESCE(rumbo_grados,0),COALESCE(capturado_en,'') FROM empresa_domicilios_tracking WHERE empresa_id=? AND order_id=? ORDER BY capturado_en DESC LIMIT %d`, limit), empresaID, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaDomicilioTrackPoint
	for rows.Next() {
		var x EmpresaDomicilioTrackPoint
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.OrderID, &x.CourierID, &x.ActorTipo, &x.Latitud, &x.Longitud, &x.PrecisionMetros, &x.VelocidadKMH, &x.RumboGrados, &x.CapturadoEn); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func ListDomicilioOrderItems(dbConn *sql.DB, empresaID, orderID int64) ([]EmpresaDomicilioOrderItem, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,order_id,COALESCE(menu_item_id,0),COALESCE(servicio_id,0),COALESCE(carrito_item_id,0),COALESCE(nombre,''),COALESCE(cantidad,0),COALESCE(precio_unit,0),COALESCE(subtotal,0),COALESCE(notas,'') FROM empresa_domicilios_order_items WHERE empresa_id=? AND order_id=? ORDER BY id`, empresaID, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaDomicilioOrderItem
	for rows.Next() {
		var x EmpresaDomicilioOrderItem
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.OrderID, &x.MenuItemID, &x.ServicioID, &x.CarritoItemID, &x.Nombre, &x.Cantidad, &x.PrecioUnit, &x.Subtotal, &x.Notas); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func ListDomicilioOrdersForRestaurant(dbConn *sql.DB, empresaID, restaurantID int64) ([]EmpresaDomicilioOrder, error) {
	rows, err := ExecQueryCompat(dbConn, domicilioOrderSelect()+` WHERE o.empresa_id=? AND o.restaurant_id=? AND o.estado NOT IN ('entregado','cancelado') ORDER BY o.fecha_pedido DESC LIMIT 80`, empresaID, restaurantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaDomicilioOrder
	for rows.Next() {
		o, err := scanDomicilioOrder(rows)
		if err != nil {
			return nil, err
		}
		o.Items, _ = ListDomicilioOrderItems(dbConn, empresaID, o.ID)
		out = append(out, o)
	}
	return out, rows.Err()
}

func ListDomicilioOrdersForCourier(dbConn *sql.DB, empresaID, courierID int64) ([]EmpresaDomicilioOrder, error) {
	rows, err := ExecQueryCompat(dbConn, domicilioOrderSelect()+` WHERE o.empresa_id=? AND o.courier_id=? AND o.estado NOT IN ('entregado','cancelado') ORDER BY o.fecha_pedido DESC LIMIT 30`, empresaID, courierID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaDomicilioOrder
	for rows.Next() {
		o, err := scanDomicilioOrder(rows)
		if err != nil {
			return nil, err
		}
		o.Items, _ = ListDomicilioOrderItems(dbConn, empresaID, o.ID)
		out = append(out, o)
	}
	return out, rows.Err()
}

func SeedEmpresaDomiciliosDemo(dbConn *sql.DB, empresaID int64, user string) error {
	rests, _ := ListDomicilioRestaurants(dbConn, empresaID, false)
	if len(rests) > 0 {
		return nil
	}
	rid, err := CreateDomicilioRestaurant(dbConn, EmpresaDomicilioRestaurant{EmpresaID: empresaID, Codigo: "REST-DEMO", Nombre: "La Cocina Central", Categoria: "Comida rapida", Responsable: "Administrador", Telefono: "3000000000", Direccion: "Zona centro", Latitud: 4.711, Longitud: -74.0721, TiempoPreparacionMin: 18, ComisionPorcentaje: 12, AceptaPedidos: true, Pin: "1234", UsuarioCreador: user})
	if err != nil {
		return err
	}
	_, _ = UpsertDomicilioMenuItem(dbConn, EmpresaDomicilioMenuItem{EmpresaID: empresaID, RestaurantID: rid, Codigo: "HAM-01", Nombre: "Hamburguesa artesanal", Descripcion: "Carne, queso, vegetales y salsa de casa", Categoria: "Hamburguesas", Precio: 22000, Disponible: true, Orden: 1})
	_, _ = UpsertDomicilioMenuItem(dbConn, EmpresaDomicilioMenuItem{EmpresaID: empresaID, RestaurantID: rid, Codigo: "BOWL-01", Nombre: "Bowl ejecutivo", Descripcion: "Arroz, pollo, aguacate y vegetales", Categoria: "Almuerzos", Precio: 26000, Disponible: true, Orden: 2})
	_, _ = UpsertDomicilioMenuItem(dbConn, EmpresaDomicilioMenuItem{EmpresaID: empresaID, RestaurantID: rid, Codigo: "JUGO-01", Nombre: "Jugo natural", Descripcion: "Fruta de temporada", Categoria: "Bebidas", Precio: 7000, Disponible: true, Orden: 3})
	_, _ = CreateDomicilioCourier(dbConn, EmpresaDomicilioCourier{EmpresaID: empresaID, Codigo: "DOM-001", Nombre: "Carlos Domicilios", Documento: "1001", Telefono: "3110000001", VehiculoTipo: "Moto", VehiculoPlaca: "DOM123", ZonaBase: "Centro", Pin: "1234", UsuarioCreador: user})
	_, _ = CreateDomicilioCourier(dbConn, EmpresaDomicilioCourier{EmpresaID: empresaID, Codigo: "DOM-002", Nombre: "Mariana Express", Documento: "1002", Telefono: "3110000002", VehiculoTipo: "Bicicleta", ZonaBase: "Norte", Pin: "1234", UsuarioCreador: user})
	return nil
}

type domicilioScanner interface {
	Scan(dest ...interface{}) error
}

func domicilioOrderSelect() string {
	return `SELECT o.id,o.empresa_id,o.restaurant_id,COALESCE(o.cliente_id,0),COALESCE(o.carrito_id,0),COALESCE(r.nombre,''),COALESCE(o.courier_id,0),COALESCE(c.nombre,''),COALESCE(c.telefono,''),COALESCE(o.codigo_pedido,''),COALESCE(o.codigo_entrega,''),COALESCE(o.token_cliente,''),COALESCE(o.cliente_nombre,''),COALESCE(o.cliente_telefono,''),COALESCE(o.cliente_direccion,''),COALESCE(o.cliente_latitud,0),COALESCE(o.cliente_longitud,0),COALESCE(o.metodo_pago,''),COALESCE(o.estado,''),COALESCE(o.canal,''),COALESCE(o.notas_cliente,''),COALESCE(o.notas_internas,''),COALESCE(o.subtotal,0),COALESCE(o.tarifa_domicilio,0),COALESCE(o.propina,0),COALESCE(o.descuento,0),COALESCE(o.total,0),COALESCE(o.distancia_estimada_km,0),COALESCE(o.tiempo_estimado_min,0),COALESCE(o.fecha_pedido,''),COALESCE(o.fecha_confirmacion,''),COALESCE(o.fecha_listo,''),COALESCE(o.fecha_recogida,''),COALESCE(o.fecha_entrega,''),COALESCE(o.fecha_cancelacion,'') FROM empresa_domicilios_orders o LEFT JOIN empresa_domicilios_restaurantes r ON r.id=o.restaurant_id AND r.empresa_id=o.empresa_id LEFT JOIN empresa_domicilios_couriers c ON c.id=o.courier_id AND c.empresa_id=o.empresa_id`
}

func scanDomicilioOrder(s domicilioScanner) (EmpresaDomicilioOrder, error) {
	var o EmpresaDomicilioOrder
	err := s.Scan(&o.ID, &o.EmpresaID, &o.RestaurantID, &o.ClienteID, &o.CarritoID, &o.RestaurantNombre, &o.CourierID, &o.CourierNombre, &o.CourierTelefono, &o.CodigoPedido, &o.CodigoEntrega, &o.TokenCliente, &o.ClienteNombre, &o.ClienteTelefono, &o.ClienteDireccion, &o.ClienteLatitud, &o.ClienteLongitud, &o.MetodoPago, &o.Estado, &o.Canal, &o.NotasCliente, &o.NotasInternas, &o.Subtotal, &o.TarifaDomicilio, &o.Propina, &o.Descuento, &o.Total, &o.DistanciaEstimadaKM, &o.TiempoEstimadoMin, &o.FechaPedido, &o.FechaConfirmacion, &o.FechaListo, &o.FechaRecogida, &o.FechaEntrega, &o.FechaCancelacion)
	return o, err
}

func domicilioCoreCode(prefix string, parts ...string) string {
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

func ensureDomicilioClienteCore(dbConn *sql.DB, order EmpresaDomicilioOrder, usuario string) (int64, error) {
	if order.ClienteID > 0 {
		return order.ClienteID, nil
	}
	if strings.TrimSpace(order.ClienteNombre) == "" && strings.TrimSpace(order.ClienteTelefono) == "" {
		return 0, nil
	}
	if err := EnsureEmpresaClientesSchema(dbConn); err != nil {
		return 0, err
	}
	if telefonoNorm := normalizeClienteTelefonoValue(order.ClienteTelefono); telefonoNorm != "" {
		query := fmt.Sprintf(`SELECT id FROM clientes WHERE empresa_id = ? AND %s = ? LIMIT 1`, clienteTelefonoSQLExpr("telefono"))
		if id, err := findClienteDuplicateID(dbConn, query, order.EmpresaID, telefonoNorm); err != nil {
			return 0, err
		} else if id > 0 {
			return id, nil
		}
	}
	nombre := strings.TrimSpace(order.ClienteNombre)
	if nombre == "" {
		nombre = "Cliente domicilios " + strings.TrimSpace(order.ClienteTelefono)
	}
	id, err := CreateCliente(dbConn, Cliente{
		EmpresaID:         order.EmpresaID,
		TipoDocumento:     "OTRO",
		NumeroDocumento:   domicilioCoreCode("DOM-CLI", order.ClienteTelefono, nombre),
		TipoPersona:       "natural",
		NombreRazonSocial: nombre,
		NombreComercial:   nombre,
		Telefono:          strings.TrimSpace(order.ClienteTelefono),
		Direccion:         strings.TrimSpace(order.ClienteDireccion),
		Pais:              "CO",
		UsuarioCreador:    strings.TrimSpace(usuario),
		Estado:            "activo",
		Observaciones:     "Cliente creado/sincronizado desde domicilios.",
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

func ensureDomicilioStaticServicio(dbConn *sql.DB, empresaID int64, code, nombre, descripcion string, precio float64, usuario string) (int64, error) {
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		return 0, err
	}
	code = strings.TrimSpace(code)
	var servicioID int64
	err := QueryRowCompat(dbConn, `SELECT id FROM servicios WHERE empresa_id=? AND codigo=? LIMIT 1`, empresaID, code).Scan(&servicioID)
	if err == nil {
		_, _ = ExecCompat(dbConn, `UPDATE servicios SET nombre=?, descripcion=?, categoria='domicilios', precio=?, estado='activo', fecha_actualizacion=? WHERE empresa_id=? AND id=?`, strings.TrimSpace(nombre), strings.TrimSpace(descripcion), precio, nowDomicilio(), empresaID, servicioID)
		return servicioID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	return CreateServicio(dbConn, Servicio{
		EmpresaID:      empresaID,
		Codigo:         code,
		Nombre:         strings.TrimSpace(nombre),
		Descripcion:    strings.TrimSpace(descripcion),
		Categoria:      "domicilios",
		Precio:         precio,
		Estado:         "activo",
		UsuarioCreador: strings.TrimSpace(usuario),
		Observaciones:  "Servicio sincronizado desde domicilios.",
	})
}

func ensureDomicilioMenuServicio(dbConn *sql.DB, item EmpresaDomicilioMenuItem, usuario string) (int64, error) {
	if item.ServicioID > 0 {
		return item.ServicioID, nil
	}
	codePart := strings.TrimSpace(item.Codigo)
	if codePart == "" && item.ID > 0 {
		codePart = fmt.Sprintf("%d", item.ID)
	}
	if codePart == "" {
		codePart = item.Nombre
	}
	code := domicilioCoreCode("DOM-MENU", fmt.Sprintf("%d", item.RestaurantID), codePart)
	servicioID, err := ensureDomicilioStaticServicio(dbConn, item.EmpresaID, code, item.Nombre, item.Descripcion, item.Precio, usuario)
	if err != nil {
		return 0, err
	}
	if item.ID > 0 && servicioID > 0 {
		_, _ = ExecCompat(dbConn, `UPDATE empresa_domicilios_menu_items SET servicio_id=?, fecha_actualizacion=? WHERE empresa_id=? AND id=?`, nullableID(servicioID), nowDomicilio(), item.EmpresaID, item.ID)
	}
	return servicioID, nil
}

func ensureDomicilioOrderItemServicio(dbConn *sql.DB, order EmpresaDomicilioOrder, item EmpresaDomicilioOrderItem, usuario string) (int64, error) {
	if item.ServicioID > 0 {
		return item.ServicioID, nil
	}
	if item.MenuItemID > 0 {
		var menu EmpresaDomicilioMenuItem
		err := QueryRowCompat(dbConn, `SELECT id,empresa_id,restaurant_id,COALESCE(servicio_id,0),COALESCE(codigo,''),COALESCE(nombre,''),COALESCE(descripcion,''),COALESCE(categoria,''),COALESCE(precio,0),COALESCE(imagen_url,''),COALESCE(tiempo_preparacion_min,0) FROM empresa_domicilios_menu_items WHERE empresa_id=? AND id=? LIMIT 1`, order.EmpresaID, item.MenuItemID).
			Scan(&menu.ID, &menu.EmpresaID, &menu.RestaurantID, &menu.ServicioID, &menu.Codigo, &menu.Nombre, &menu.Descripcion, &menu.Categoria, &menu.Precio, &menu.ImagenURL, &menu.TiempoPreparacionMin)
		if err == nil {
			return ensureDomicilioMenuServicio(dbConn, menu, usuario)
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return 0, err
		}
	}
	menu := EmpresaDomicilioMenuItem{
		EmpresaID:    order.EmpresaID,
		RestaurantID: order.RestaurantID,
		Codigo:       fmt.Sprintf("ORDER-%d-ITEM-%d", order.ID, item.ID),
		Nombre:       item.Nombre,
		Descripcion:  "Item vendido desde pedido de domicilios.",
		Categoria:    "domicilios",
		Precio:       item.PrecioUnit,
	}
	return ensureDomicilioMenuServicio(dbConn, menu, usuario)
}

func createDomicilioOrderCarrito(dbConn *sql.DB, order EmpresaDomicilioOrder, usuario string) (int64, int64, error) {
	if order.Total <= 0 {
		return order.CarritoID, order.ClienteID, nil
	}
	if err := EnsureEmpresaCarritosSchema(dbConn); err != nil {
		return 0, 0, err
	}
	clienteID, err := ensureDomicilioClienteCore(dbConn, order, usuario)
	if err != nil {
		return 0, 0, err
	}
	order.ClienteID = clienteID
	if len(order.Items) == 0 {
		order.Items, _ = ListDomicilioOrderItems(dbConn, order.EmpresaID, order.ID)
	}
	metodo := NormalizeMetodoPagoCarrito(order.MetodoPago)
	if metodo == "" {
		metodo = "efectivo"
	}
	cfg, _ := GetEmpresaDomiciliosConfig(dbConn, order.EmpresaID)
	moneda := strings.TrimSpace(strings.ToUpper(cfg.Moneda))
	if moneda == "" {
		moneda = "COP"
	}
	referenciaExterna := fmt.Sprintf("domicilios:order:%d:%s", order.ID, order.CodigoPedido)
	var carritoExistente int64
	err = QueryRowCompat(dbConn, `SELECT id FROM carritos_compras WHERE empresa_id=? AND referencia_externa=? LIMIT 1`, order.EmpresaID, referenciaExterna).Scan(&carritoExistente)
	if err == nil && carritoExistente > 0 {
		return carritoExistente, clienteID, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, 0, err
	}
	carritoID, err := CreateCarritoCompra(dbConn, CarritoCompra{
		EmpresaID:         order.EmpresaID,
		Codigo:            domicilioCoreCode("DOM-PEDIDO", fmt.Sprintf("%d", order.ID), order.CodigoPedido),
		Nombre:            "Domicilios - " + strings.TrimSpace(order.ClienteNombre),
		CanalVenta:        "domicilios",
		ClienteID:         clienteID,
		EstadoCarrito:     "abierto",
		Moneda:            moneda,
		ReferenciaExterna: referenciaExterna,
		MetodoPago:        metodo,
		ReferenciaPago:    order.CodigoPedido,
		UsuarioCreador:    strings.TrimSpace(usuario),
		Observaciones:     "Venta central generada desde pedido entregado de domicilios.",
	})
	if err != nil {
		return 0, 0, err
	}
	for _, it := range order.Items {
		if it.Cantidad <= 0 {
			continue
		}
		servicioID, err := ensureDomicilioOrderItemServicio(dbConn, order, it, usuario)
		if err != nil {
			return 0, 0, err
		}
		itemID, err := CreateCarritoCompraItem(dbConn, CarritoCompraItem{
			EmpresaID:          order.EmpresaID,
			CarritoID:          carritoID,
			TipoItem:           "servicio",
			ReferenciaID:       servicioID,
			CodigoItem:         domicilioCoreCode("DOM-ITEM", order.CodigoPedido, fmt.Sprintf("%d", it.ID)),
			Descripcion:        strings.TrimSpace(it.Nombre),
			UnidadMedida:       "servicio",
			Cantidad:           it.Cantidad,
			PrecioUnitario:     it.PrecioUnit,
			ImpuestoPorcentaje: 0,
			UsuarioCreador:     strings.TrimSpace(usuario),
			Estado:             "activo",
			Observaciones:      strings.TrimSpace(it.Notas),
		})
		if err != nil {
			return 0, 0, err
		}
		_, _ = ExecCompat(dbConn, `UPDATE empresa_domicilios_order_items SET servicio_id=?, carrito_item_id=? WHERE empresa_id=? AND id=?`, nullableID(servicioID), nullableID(itemID), order.EmpresaID, it.ID)
	}
	if order.TarifaDomicilio > 0 {
		servicioID, err := ensureDomicilioStaticServicio(dbConn, order.EmpresaID, "DOM-TARIFA-DOMICILIO", "Tarifa de domicilio", "Servicio central para tarifa de entrega a domicilio.", order.TarifaDomicilio, usuario)
		if err != nil {
			return 0, 0, err
		}
		if _, err := CreateCarritoCompraItem(dbConn, CarritoCompraItem{
			EmpresaID:          order.EmpresaID,
			CarritoID:          carritoID,
			TipoItem:           "servicio",
			ReferenciaID:       servicioID,
			CodigoItem:         domicilioCoreCode("DOM-FEE", order.CodigoPedido),
			Descripcion:        "Tarifa de domicilio",
			UnidadMedida:       "servicio",
			Cantidad:           1,
			PrecioUnitario:     order.TarifaDomicilio,
			ImpuestoPorcentaje: 0,
			UsuarioCreador:     strings.TrimSpace(usuario),
			Estado:             "activo",
		}); err != nil {
			return 0, 0, err
		}
	}
	if order.Propina > 0 {
		servicioID, err := ensureDomicilioStaticServicio(dbConn, order.EmpresaID, "DOM-PROPINA", "Propina domicilios", "Servicio central para propinas de domicilios.", order.Propina, usuario)
		if err != nil {
			return 0, 0, err
		}
		if _, err := CreateCarritoCompraItem(dbConn, CarritoCompraItem{
			EmpresaID:          order.EmpresaID,
			CarritoID:          carritoID,
			TipoItem:           "servicio",
			ReferenciaID:       servicioID,
			CodigoItem:         domicilioCoreCode("DOM-TIP", order.CodigoPedido),
			Descripcion:        "Propina domicilios",
			UnidadMedida:       "servicio",
			Cantidad:           1,
			PrecioUnitario:     order.Propina,
			ImpuestoPorcentaje: 0,
			UsuarioCreador:     strings.TrimSpace(usuario),
			Estado:             "activo",
		}); err != nil {
			return 0, 0, err
		}
	}
	descuentoTipo := ""
	if order.Descuento > 0 {
		descuentoTipo = "manual"
	}
	if err := PayCarritoStationSession(dbConn, order.EmpresaID, carritoID, metodo, order.CodigoPedido, descuentoTipo, "", order.Descuento, 0, order.Total, 0, 0, "", "", 0, strings.TrimSpace(usuario)); err != nil {
		return 0, 0, err
	}
	return carritoID, clienteID, nil
}

func hashDomicilioPin(pin string) (string, string) {
	pin = strings.TrimSpace(pin)
	if pin == "" {
		pin = "1234"
	}
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	salt := hex.EncodeToString(buf)
	sum := sha256.Sum256([]byte(salt + ":" + pin))
	return salt, hex.EncodeToString(sum[:])
}

func verifyDomicilioPin(pin, salt, hash string) bool {
	if strings.TrimSpace(pin) == "" || salt == "" || hash == "" {
		return false
	}
	sum := sha256.Sum256([]byte(salt + ":" + strings.TrimSpace(pin)))
	return hex.EncodeToString(sum[:]) == hash
}

func newDomicilioToken() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
func randomDigitsDomicilio(n int) string {
	if n <= 0 {
		n = 4
	}
	b := make([]byte, n)
	_, _ = rand.Read(b)
	out := make([]byte, n)
	for i := range b {
		out[i] = byte('0' + int(b[i])%10)
	}
	return string(out)
}
func nowDomicilio() string { return time.Now().Format("2006-01-02 15:04:05") }
func domicilioBoolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
func defaultTrueDomicilio(v bool) bool { return v }
func firstDomicilioState(v, fallback string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return fallback
	}
	return v
}
func roundDomicilio(v float64) float64 { return math.Round(v*100) / 100 }
func haversineDomicilio(lat1, lon1, lat2, lon2 float64) float64 {
	if lat1 == 0 && lon1 == 0 || lat2 == 0 && lon2 == 0 {
		return 0
	}
	const r = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return r * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
