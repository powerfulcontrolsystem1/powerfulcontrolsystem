package db

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"
)

type EmpresaAlquilerConfig struct {
	EmpresaID                int64   `json:"empresa_id"`
	NombreSistema            string  `json:"nombre_sistema"`
	Moneda                   string  `json:"moneda"`
	PermitirReservas         bool    `json:"permitir_reservas"`
	PermitirGPS              bool    `json:"permitir_gps"`
	RequerirDeposito         bool    `json:"requerir_deposito"`
	PermitirKilometraje      bool    `json:"permitir_kilometraje"`
	RequerirChecklist        bool    `json:"requerir_checklist"`
	PermitirEntregaDomicilio bool    `json:"permitir_entrega_domicilio"`
	AlertarVencimientoHoras  int     `json:"alertar_vencimiento_horas"`
	DepositoBaseSugerido     float64 `json:"deposito_base_sugerido"`
	FechaActualizacion       string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador           string  `json:"usuario_creador,omitempty"`
}

type EmpresaAlquilerCategoria struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Codigo             string `json:"codigo"`
	Nombre             string `json:"nombre"`
	TipoActivo         string `json:"tipo_activo,omitempty"`
	Descripcion        string `json:"descripcion,omitempty"`
	Estado             string `json:"estado,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
}

type EmpresaAlquilerActivo struct {
	ID                   int64   `json:"id"`
	EmpresaID            int64   `json:"empresa_id"`
	Codigo               string  `json:"codigo"`
	Nombre               string  `json:"nombre"`
	CategoriaID          int64   `json:"categoria_id,omitempty"`
	CategoriaNombre      string  `json:"categoria_nombre,omitempty"`
	TipoActivo           string  `json:"tipo_activo,omitempty"`
	Marca                string  `json:"marca,omitempty"`
	Modelo               string  `json:"modelo,omitempty"`
	Serie                string  `json:"serie,omitempty"`
	Placa                string  `json:"placa,omitempty"`
	Sede                 string  `json:"sede,omitempty"`
	Estado               string  `json:"estado,omitempty"`
	ValorReposicion      float64 `json:"valor_reposicion"`
	CostoBaseHora        float64 `json:"costo_base_hora"`
	DepositoSugerido     float64 `json:"deposito_sugerido"`
	UsaGPS               bool    `json:"usa_gps"`
	RequiereChecklist    bool    `json:"requiere_checklist"`
	RequiereLicencia     bool    `json:"requiere_licencia"`
	UrlFoto              string  `json:"url_foto,omitempty"`
	LatitudActual        float64 `json:"latitud_actual,omitempty"`
	LongitudActual       float64 `json:"longitud_actual,omitempty"`
	FechaUltimaUbicacion string  `json:"fecha_ultima_ubicacion,omitempty"`
	Notas                string  `json:"notas,omitempty"`
	FechaCreacion        string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion   string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador       string  `json:"usuario_creador,omitempty"`
}

type EmpresaAlquilerTarifa struct {
	ID                  int64   `json:"id"`
	EmpresaID           int64   `json:"empresa_id"`
	Codigo              string  `json:"codigo"`
	Nombre              string  `json:"nombre"`
	CategoriaID         int64   `json:"categoria_id,omitempty"`
	CategoriaNombre     string  `json:"categoria_nombre,omitempty"`
	ModalidadCobro      string  `json:"modalidad_cobro,omitempty"`
	PrecioBase          float64 `json:"precio_base"`
	PrecioHora          float64 `json:"precio_hora"`
	PrecioDia           float64 `json:"precio_dia"`
	PrecioSemana        float64 `json:"precio_semana"`
	PrecioMes           float64 `json:"precio_mes"`
	KilometrosIncluidos float64 `json:"kilometros_incluidos"`
	DepositoMinimo      float64 `json:"deposito_minimo"`
	Estado              string  `json:"estado,omitempty"`
	FechaCreacion       string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion  string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador      string  `json:"usuario_creador,omitempty"`
}

type EmpresaAlquilerContrato struct {
	ID                   int64   `json:"id"`
	EmpresaID            int64   `json:"empresa_id"`
	Codigo               string  `json:"codigo"`
	TipoRegistro         string  `json:"tipo_registro,omitempty"`
	ActivoID             int64   `json:"activo_id,omitempty"`
	ActivoNombre         string  `json:"activo_nombre,omitempty"`
	CategoriaNombre      string  `json:"categoria_nombre,omitempty"`
	ClienteNombre        string  `json:"cliente_nombre"`
	ClienteDocumento     string  `json:"cliente_documento,omitempty"`
	ClienteTelefono      string  `json:"cliente_telefono,omitempty"`
	ClienteEmail         string  `json:"cliente_email,omitempty"`
	ResponsableEmpresa   string  `json:"responsable_empresa,omitempty"`
	TarifaID             int64   `json:"tarifa_id,omitempty"`
	TarifaNombre         string  `json:"tarifa_nombre,omitempty"`
	ModalidadCobro       string  `json:"modalidad_cobro,omitempty"`
	FechaReserva         string  `json:"fecha_reserva,omitempty"`
	FechaInicio          string  `json:"fecha_inicio,omitempty"`
	FechaFinPrevista     string  `json:"fecha_fin_prevista,omitempty"`
	FechaEntregaReal     string  `json:"fecha_entrega_real,omitempty"`
	FechaDevolucionReal  string  `json:"fecha_devolucion_real,omitempty"`
	Estado               string  `json:"estado,omitempty"`
	Cantidad             int     `json:"cantidad"`
	HorasPlaneadas       float64 `json:"horas_planeadas"`
	DiasPlaneados        float64 `json:"dias_planeados"`
	KilometrosIncluidos  float64 `json:"kilometros_incluidos"`
	Deposito             float64 `json:"deposito"`
	ValorBase            float64 `json:"valor_base"`
	Descuento            float64 `json:"descuento"`
	Impuestos            float64 `json:"impuestos"`
	Total                float64 `json:"total"`
	SaldoPendiente       float64 `json:"saldo_pendiente"`
	OrigenEntrega        string  `json:"origen_entrega,omitempty"`
	DestinoDevolucion    string  `json:"destino_devolucion,omitempty"`
	Observaciones        string  `json:"observaciones,omitempty"`
	RequiereGarantia     bool    `json:"requiere_garantia"`
	GpsTrackingActivo    bool    `json:"gps_tracking_activo"`
	LatitudActual        float64 `json:"latitud_actual,omitempty"`
	LongitudActual       float64 `json:"longitud_actual,omitempty"`
	FechaUltimaUbicacion string  `json:"fecha_ultima_ubicacion,omitempty"`
	FechaCreacion        string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion   string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador       string  `json:"usuario_creador,omitempty"`
}

type EmpresaAlquilerMantenimiento struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	ActivoID           int64   `json:"activo_id"`
	ActivoNombre       string  `json:"activo_nombre,omitempty"`
	Tipo               string  `json:"tipo,omitempty"`
	Prioridad          string  `json:"prioridad,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	FechaProgramada    string  `json:"fecha_programada,omitempty"`
	FechaCierre        string  `json:"fecha_cierre,omitempty"`
	Proveedor          string  `json:"proveedor,omitempty"`
	CostoEstimado      float64 `json:"costo_estimado"`
	CostoReal          float64 `json:"costo_real"`
	Descripcion        string  `json:"descripcion,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
}

type EmpresaAlquilerUbicacion struct {
	ID              int64   `json:"id"`
	EmpresaID       int64   `json:"empresa_id"`
	ActivoID        int64   `json:"activo_id,omitempty"`
	ActivoNombre    string  `json:"activo_nombre,omitempty"`
	ContratoID      int64   `json:"contrato_id,omitempty"`
	ContratoCodigo  string  `json:"contrato_codigo,omitempty"`
	Latitud         float64 `json:"latitud"`
	Longitud        float64 `json:"longitud"`
	Velocidad       float64 `json:"velocidad,omitempty"`
	PrecisionMetros float64 `json:"precision_metros,omitempty"`
	Fuente          string  `json:"fuente,omitempty"`
	Referencia      string  `json:"referencia,omitempty"`
	FechaRegistro   string  `json:"fecha_registro,omitempty"`
	UsuarioCreador  string  `json:"usuario_creador,omitempty"`
}

type EmpresaAlquilerResumenGrupo struct {
	Clave       string  `json:"clave"`
	Etiqueta    string  `json:"etiqueta"`
	Cantidad    int     `json:"cantidad"`
	Monto       float64 `json:"monto"`
	Utilizacion float64 `json:"utilizacion"`
}

type EmpresaAlquilerDashboard struct {
	EmpresaID              int64                         `json:"empresa_id"`
	ActivosDisponibles     int                           `json:"activos_disponibles"`
	ActivosAlquilados      int                           `json:"activos_alquilados"`
	ReservasPendientes     int                           `json:"reservas_pendientes"`
	ContratosVencidos      int                           `json:"contratos_vencidos"`
	DevolucionesHoy        int                           `json:"devoluciones_hoy"`
	MantenimientosAbiertos int                           `json:"mantenimientos_abiertos"`
	IngresosMes            float64                       `json:"ingresos_mes"`
	DepositosRetenidos     float64                       `json:"depositos_retenidos"`
	UtilizacionPromedio    float64                       `json:"utilizacion_promedio"`
	ProximosVencimientos   []EmpresaAlquilerContrato     `json:"proximos_vencimientos"`
	ActivosEnRiesgo        []EmpresaAlquilerActivo       `json:"activos_en_riesgo"`
	IngresosPorLinea       []EmpresaAlquilerResumenGrupo `json:"ingresos_por_linea"`
	IngresosPorSede        []EmpresaAlquilerResumenGrupo `json:"ingresos_por_sede"`
}

var (
	empresaAlquileresSchemaEnsured sync.Map
	empresaAlquileresSchemaMu      sync.Mutex
)

func EnsureEmpresaAlquileresSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("dbConn es obligatorio")
	}
	cacheKey := fmt.Sprintf("%p", dbConn)
	if _, ok := empresaAlquileresSchemaEnsured.Load(cacheKey); ok {
		return nil
	}
	empresaAlquileresSchemaMu.Lock()
	defer empresaAlquileresSchemaMu.Unlock()
	if _, ok := empresaAlquileresSchemaEnsured.Load(cacheKey); ok {
		return nil
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_alquileres_config (
			empresa_id BIGINT PRIMARY KEY,
			nombre_sistema TEXT DEFAULT 'Alquileres y contratos',
			moneda TEXT DEFAULT 'COP',
			permitir_reservas INTEGER DEFAULT 1,
			permitir_gps INTEGER DEFAULT 0,
			requerir_deposito INTEGER DEFAULT 1,
			permitir_kilometraje INTEGER DEFAULT 0,
			requerir_checklist INTEGER DEFAULT 1,
			permitir_entrega_domicilio INTEGER DEFAULT 0,
			alertar_vencimiento_horas INTEGER DEFAULT 12,
			deposito_base_sugerido NUMERIC(14,2) DEFAULT 0,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_alquileres_categorias (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			tipo_activo TEXT DEFAULT 'equipo',
			descripcion TEXT,
			estado TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_alquileres_categorias_codigo ON empresa_alquileres_categorias(empresa_id, codigo)`,
		`CREATE TABLE IF NOT EXISTS empresa_alquileres_activos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			categoria_id BIGINT,
			tipo_activo TEXT DEFAULT 'equipo',
			marca TEXT,
			modelo TEXT,
			serie TEXT,
			placa TEXT,
			sede TEXT DEFAULT 'principal',
			estado TEXT DEFAULT 'disponible',
			valor_reposicion NUMERIC(14,2) DEFAULT 0,
			costo_base_hora NUMERIC(14,2) DEFAULT 0,
			deposito_sugerido NUMERIC(14,2) DEFAULT 0,
			usa_gps INTEGER DEFAULT 0,
			requiere_checklist INTEGER DEFAULT 0,
			requiere_licencia INTEGER DEFAULT 0,
			url_foto TEXT,
			latitud_actual NUMERIC(12,8) DEFAULT 0,
			longitud_actual NUMERIC(12,8) DEFAULT 0,
			fecha_ultima_ubicacion TEXT,
			notas TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_alquileres_activos_codigo ON empresa_alquileres_activos(empresa_id, codigo)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_alquileres_activos_estado ON empresa_alquileres_activos(empresa_id, estado, sede, id DESC)`,
		`CREATE TABLE IF NOT EXISTS empresa_alquileres_tarifas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			categoria_id BIGINT,
			modalidad_cobro TEXT DEFAULT 'dia',
			precio_base NUMERIC(14,2) DEFAULT 0,
			precio_hora NUMERIC(14,2) DEFAULT 0,
			precio_dia NUMERIC(14,2) DEFAULT 0,
			precio_semana NUMERIC(14,2) DEFAULT 0,
			precio_mes NUMERIC(14,2) DEFAULT 0,
			kilometros_incluidos NUMERIC(14,2) DEFAULT 0,
			deposito_minimo NUMERIC(14,2) DEFAULT 0,
			estado TEXT DEFAULT 'activa',
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_alquileres_tarifas_codigo ON empresa_alquileres_tarifas(empresa_id, codigo)`,
		`CREATE TABLE IF NOT EXISTS empresa_alquileres_contratos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			codigo TEXT NOT NULL,
			tipo_registro TEXT DEFAULT 'alquiler',
			activo_id BIGINT NOT NULL,
			cliente_nombre TEXT NOT NULL,
			cliente_documento TEXT,
			cliente_telefono TEXT,
			cliente_email TEXT,
			responsable_empresa TEXT,
			tarifa_id BIGINT,
			modalidad_cobro TEXT DEFAULT 'dia',
			fecha_reserva TEXT,
			fecha_inicio TEXT,
			fecha_fin_prevista TEXT,
			fecha_entrega_real TEXT,
			fecha_devolucion_real TEXT,
			estado TEXT DEFAULT 'reservado',
			cantidad INTEGER DEFAULT 1,
			horas_planeadas NUMERIC(14,2) DEFAULT 0,
			dias_planeados NUMERIC(14,2) DEFAULT 0,
			kilometros_incluidos NUMERIC(14,2) DEFAULT 0,
			deposito NUMERIC(14,2) DEFAULT 0,
			valor_base NUMERIC(14,2) DEFAULT 0,
			descuento NUMERIC(14,2) DEFAULT 0,
			impuestos NUMERIC(14,2) DEFAULT 0,
			total NUMERIC(14,2) DEFAULT 0,
			saldo_pendiente NUMERIC(14,2) DEFAULT 0,
			origen_entrega TEXT,
			destino_devolucion TEXT,
			observaciones TEXT,
			requiere_garantia INTEGER DEFAULT 0,
			gps_tracking_activo INTEGER DEFAULT 0,
			latitud_actual NUMERIC(12,8) DEFAULT 0,
			longitud_actual NUMERIC(12,8) DEFAULT 0,
			fecha_ultima_ubicacion TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_alquileres_contratos_codigo ON empresa_alquileres_contratos(empresa_id, codigo)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_alquileres_contratos_estado ON empresa_alquileres_contratos(empresa_id, estado, id DESC)`,
		`CREATE TABLE IF NOT EXISTS empresa_alquileres_mantenimientos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			activo_id BIGINT NOT NULL,
			tipo TEXT DEFAULT 'preventivo',
			prioridad TEXT DEFAULT 'media',
			estado TEXT DEFAULT 'abierto',
			fecha_programada TEXT,
			fecha_cierre TEXT,
			proveedor TEXT,
			costo_estimado NUMERIC(14,2) DEFAULT 0,
			costo_real NUMERIC(14,2) DEFAULT 0,
			descripcion TEXT,
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_alquileres_mantenimientos_estado ON empresa_alquileres_mantenimientos(empresa_id, estado, id DESC)`,
		`CREATE TABLE IF NOT EXISTS empresa_alquileres_ubicaciones (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			activo_id BIGINT,
			contrato_id BIGINT,
			latitud NUMERIC(12,8) NOT NULL,
			longitud NUMERIC(12,8) NOT NULL,
			velocidad NUMERIC(12,2) DEFAULT 0,
			precision_metros NUMERIC(12,2) DEFAULT 0,
			fuente TEXT DEFAULT 'manual',
			referencia TEXT,
			fecha_registro TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_alquileres_ubicaciones_activo ON empresa_alquileres_ubicaciones(empresa_id, activo_id, id DESC)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_alquileres_ubicaciones_contrato ON empresa_alquileres_ubicaciones(empresa_id, contrato_id, id DESC)`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	empresaAlquileresSchemaEnsured.Store(cacheKey, true)
	return nil
}

func defaultEmpresaAlquilerConfig(empresaID int64) EmpresaAlquilerConfig {
	return EmpresaAlquilerConfig{
		EmpresaID:                empresaID,
		NombreSistema:            "Alquileres y contratos",
		Moneda:                   "COP",
		PermitirReservas:         true,
		PermitirGPS:              false,
		RequerirDeposito:         true,
		PermitirKilometraje:      false,
		RequerirChecklist:        true,
		PermitirEntregaDomicilio: false,
		AlertarVencimientoHoras:  12,
	}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func normalizeAlquilerEstado(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "reservado", "reserva":
		return "reservado"
	case "entregado", "en_curso", "activo":
		return "en_curso"
	case "devuelto", "cerrado":
		return "devuelto"
	case "cancelado", "cancelada":
		return "cancelado"
	case "vencido":
		return "vencido"
	default:
		return "reservado"
	}
}

func normalizeActivoEstado(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "disponible", "alquilado", "mantenimiento", "fuera_de_servicio", "reservado":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return "disponible"
	}
}

func normalizeModalidadCobro(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "hora", "dia", "semana", "mes", "kilometro", "evento":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return "dia"
	}
}

func GetEmpresaAlquilerConfig(dbConn *sql.DB, empresaID int64) (EmpresaAlquilerConfig, error) {
	if err := EnsureEmpresaAlquileresSchema(dbConn); err != nil {
		return EmpresaAlquilerConfig{}, err
	}
	cfg := defaultEmpresaAlquilerConfig(empresaID)
	var permitirReservas, permitirGPS, requerirDeposito, permitirKilometraje, requerirChecklist, permitirEntrega int
	err := QueryRowCompat(dbConn, `SELECT empresa_id, COALESCE(nombre_sistema,''), COALESCE(moneda,'COP'), COALESCE(permitir_reservas,1), COALESCE(permitir_gps,0), COALESCE(requerir_deposito,1), COALESCE(permitir_kilometraje,0), COALESCE(requerir_checklist,1), COALESCE(permitir_entrega_domicilio,0), COALESCE(alertar_vencimiento_horas,12), COALESCE(deposito_base_sugerido,0), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'') FROM empresa_alquileres_config WHERE empresa_id = ?`, empresaID).Scan(
		&cfg.EmpresaID, &cfg.NombreSistema, &cfg.Moneda, &permitirReservas, &permitirGPS, &requerirDeposito, &permitirKilometraje, &requerirChecklist, &permitirEntrega, &cfg.AlertarVencimientoHoras, &cfg.DepositoBaseSugerido, &cfg.FechaActualizacion, &cfg.UsuarioCreador,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return cfg, nil
		}
		return EmpresaAlquilerConfig{}, err
	}
	cfg.PermitirReservas = permitirReservas > 0
	cfg.PermitirGPS = permitirGPS > 0
	cfg.RequerirDeposito = requerirDeposito > 0
	cfg.PermitirKilometraje = permitirKilometraje > 0
	cfg.RequerirChecklist = requerirChecklist > 0
	cfg.PermitirEntregaDomicilio = permitirEntrega > 0
	return cfg, nil
}

func UpsertEmpresaAlquilerConfig(dbConn *sql.DB, cfg EmpresaAlquilerConfig) error {
	if err := EnsureEmpresaAlquileresSchema(dbConn); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.NombreSistema) == "" {
		cfg.NombreSistema = "Alquileres y contratos"
	}
	if strings.TrimSpace(cfg.Moneda) == "" {
		cfg.Moneda = "COP"
	}
	if cfg.AlertarVencimientoHoras <= 0 {
		cfg.AlertarVencimientoHoras = 12
	}
	_, err := ExecCompat(dbConn, `INSERT INTO empresa_alquileres_config (empresa_id, nombre_sistema, moneda, permitir_reservas, permitir_gps, requerir_deposito, permitir_kilometraje, requerir_checklist, permitir_entrega_domicilio, alertar_vencimiento_horas, deposito_base_sugerido, fecha_actualizacion, usuario_creador)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?)
		ON CONFLICT (empresa_id) DO UPDATE SET
			nombre_sistema = EXCLUDED.nombre_sistema,
			moneda = EXCLUDED.moneda,
			permitir_reservas = EXCLUDED.permitir_reservas,
			permitir_gps = EXCLUDED.permitir_gps,
			requerir_deposito = EXCLUDED.requerir_deposito,
			permitir_kilometraje = EXCLUDED.permitir_kilometraje,
			requerir_checklist = EXCLUDED.requerir_checklist,
			permitir_entrega_domicilio = EXCLUDED.permitir_entrega_domicilio,
			alertar_vencimiento_horas = EXCLUDED.alertar_vencimiento_horas,
			deposito_base_sugerido = EXCLUDED.deposito_base_sugerido,
			fecha_actualizacion = CURRENT_TIMESTAMP,
			usuario_creador = EXCLUDED.usuario_creador`,
		cfg.EmpresaID, cfg.NombreSistema, strings.ToUpper(strings.TrimSpace(cfg.Moneda)), boolInt(cfg.PermitirReservas), boolInt(cfg.PermitirGPS), boolInt(cfg.RequerirDeposito), boolInt(cfg.PermitirKilometraje), boolInt(cfg.RequerirChecklist), boolInt(cfg.PermitirEntregaDomicilio), cfg.AlertarVencimientoHoras, cfg.DepositoBaseSugerido, strings.TrimSpace(cfg.UsuarioCreador))
	return err
}

func CreateEmpresaAlquilerCategoria(dbConn *sql.DB, item EmpresaAlquilerCategoria) (int64, error) {
	if err := EnsureEmpresaAlquileresSchema(dbConn); err != nil {
		return 0, err
	}
	item.Codigo = strings.ToUpper(strings.TrimSpace(item.Codigo))
	item.Nombre = strings.TrimSpace(item.Nombre)
	if item.Codigo == "" || item.Nombre == "" {
		return 0, fmt.Errorf("codigo y nombre son obligatorios")
	}
	if strings.TrimSpace(item.TipoActivo) == "" {
		item.TipoActivo = "equipo"
	}
	if strings.TrimSpace(item.Estado) == "" {
		item.Estado = "activo"
	}
	if item.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_alquileres_categorias SET codigo=?, nombre=?, tipo_activo=?, descripcion=?, estado=?, fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=? WHERE empresa_id=? AND id=?`, item.Codigo, item.Nombre, item.TipoActivo, strings.TrimSpace(item.Descripcion), item.Estado, strings.TrimSpace(item.UsuarioCreador), item.EmpresaID, item.ID)
		return item.ID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_alquileres_categorias (empresa_id, codigo, nombre, tipo_activo, descripcion, estado, fecha_creacion, fecha_actualizacion, usuario_creador) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?)`, item.EmpresaID, item.Codigo, item.Nombre, item.TipoActivo, strings.TrimSpace(item.Descripcion), item.Estado, strings.TrimSpace(item.UsuarioCreador))
}

func ListEmpresaAlquilerCategorias(dbConn *sql.DB, empresaID int64) ([]EmpresaAlquilerCategoria, error) {
	if err := EnsureEmpresaAlquileresSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT id, empresa_id, COALESCE(codigo,''), COALESCE(nombre,''), COALESCE(tipo_activo,''), COALESCE(descripcion,''), COALESCE(estado,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'') FROM empresa_alquileres_categorias WHERE empresa_id=? ORDER BY nombre`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EmpresaAlquilerCategoria, 0)
	for rows.Next() {
		var item EmpresaAlquilerCategoria
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.Codigo, &item.Nombre, &item.TipoActivo, &item.Descripcion, &item.Estado, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func CreateEmpresaAlquilerActivo(dbConn *sql.DB, item EmpresaAlquilerActivo) (int64, error) {
	if err := EnsureEmpresaAlquileresSchema(dbConn); err != nil {
		return 0, err
	}
	item.Codigo = strings.ToUpper(strings.TrimSpace(item.Codigo))
	item.Nombre = strings.TrimSpace(item.Nombre)
	if item.Codigo == "" || item.Nombre == "" {
		return 0, fmt.Errorf("codigo y nombre son obligatorios")
	}
	item.Estado = normalizeActivoEstado(item.Estado)
	if strings.TrimSpace(item.TipoActivo) == "" {
		item.TipoActivo = "equipo"
	}
	if strings.TrimSpace(item.Sede) == "" {
		item.Sede = "principal"
	}
	if item.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_alquileres_activos SET codigo=?, nombre=?, categoria_id=?, tipo_activo=?, marca=?, modelo=?, serie=?, placa=?, sede=?, estado=?, valor_reposicion=?, costo_base_hora=?, deposito_sugerido=?, usa_gps=?, requiere_checklist=?, requiere_licencia=?, url_foto=?, notas=?, fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=? WHERE empresa_id=? AND id=?`,
			item.Codigo, item.Nombre, item.CategoriaID, item.TipoActivo, strings.TrimSpace(item.Marca), strings.TrimSpace(item.Modelo), strings.TrimSpace(item.Serie), strings.TrimSpace(item.Placa), item.Sede, item.Estado, item.ValorReposicion, item.CostoBaseHora, item.DepositoSugerido, boolInt(item.UsaGPS), boolInt(item.RequiereChecklist), boolInt(item.RequiereLicencia), strings.TrimSpace(item.UrlFoto), strings.TrimSpace(item.Notas), strings.TrimSpace(item.UsuarioCreador), item.EmpresaID, item.ID)
		return item.ID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_alquileres_activos (empresa_id, codigo, nombre, categoria_id, tipo_activo, marca, modelo, serie, placa, sede, estado, valor_reposicion, costo_base_hora, deposito_sugerido, usa_gps, requiere_checklist, requiere_licencia, url_foto, notas, fecha_creacion, fecha_actualizacion, usuario_creador) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?)`,
		item.EmpresaID, item.Codigo, item.Nombre, item.CategoriaID, item.TipoActivo, strings.TrimSpace(item.Marca), strings.TrimSpace(item.Modelo), strings.TrimSpace(item.Serie), strings.TrimSpace(item.Placa), item.Sede, item.Estado, item.ValorReposicion, item.CostoBaseHora, item.DepositoSugerido, boolInt(item.UsaGPS), boolInt(item.RequiereChecklist), boolInt(item.RequiereLicencia), strings.TrimSpace(item.UrlFoto), strings.TrimSpace(item.Notas), strings.TrimSpace(item.UsuarioCreador))
}

func ListEmpresaAlquilerActivos(dbConn *sql.DB, empresaID int64) ([]EmpresaAlquilerActivo, error) {
	if err := EnsureEmpresaAlquileresSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT a.id, a.empresa_id, COALESCE(a.codigo,''), COALESCE(a.nombre,''), COALESCE(a.categoria_id,0), COALESCE(c.nombre,''), COALESCE(a.tipo_activo,''), COALESCE(a.marca,''), COALESCE(a.modelo,''), COALESCE(a.serie,''), COALESCE(a.placa,''), COALESCE(a.sede,''), COALESCE(a.estado,''), COALESCE(a.valor_reposicion,0), COALESCE(a.costo_base_hora,0), COALESCE(a.deposito_sugerido,0), COALESCE(a.usa_gps,0), COALESCE(a.requiere_checklist,0), COALESCE(a.requiere_licencia,0), COALESCE(a.url_foto,''), COALESCE(a.latitud_actual,0), COALESCE(a.longitud_actual,0), COALESCE(a.fecha_ultima_ubicacion,''), COALESCE(a.notas,''), COALESCE(a.fecha_creacion,''), COALESCE(a.fecha_actualizacion,''), COALESCE(a.usuario_creador,'')
		FROM empresa_alquileres_activos a
		LEFT JOIN empresa_alquileres_categorias c ON c.id = a.categoria_id
		WHERE a.empresa_id=? ORDER BY a.id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EmpresaAlquilerActivo, 0)
	for rows.Next() {
		var item EmpresaAlquilerActivo
		var usaGPS, checklist, licencia int
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.Codigo, &item.Nombre, &item.CategoriaID, &item.CategoriaNombre, &item.TipoActivo, &item.Marca, &item.Modelo, &item.Serie, &item.Placa, &item.Sede, &item.Estado, &item.ValorReposicion, &item.CostoBaseHora, &item.DepositoSugerido, &usaGPS, &checklist, &licencia, &item.UrlFoto, &item.LatitudActual, &item.LongitudActual, &item.FechaUltimaUbicacion, &item.Notas, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		item.UsaGPS = usaGPS > 0
		item.RequiereChecklist = checklist > 0
		item.RequiereLicencia = licencia > 0
		out = append(out, item)
	}
	return out, rows.Err()
}

func CreateEmpresaAlquilerTarifa(dbConn *sql.DB, item EmpresaAlquilerTarifa) (int64, error) {
	if err := EnsureEmpresaAlquileresSchema(dbConn); err != nil {
		return 0, err
	}
	item.Codigo = strings.ToUpper(strings.TrimSpace(item.Codigo))
	item.Nombre = strings.TrimSpace(item.Nombre)
	if item.Codigo == "" || item.Nombre == "" {
		return 0, fmt.Errorf("codigo y nombre son obligatorios")
	}
	item.ModalidadCobro = normalizeModalidadCobro(item.ModalidadCobro)
	if strings.TrimSpace(item.Estado) == "" {
		item.Estado = "activa"
	}
	if item.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_alquileres_tarifas SET codigo=?, nombre=?, categoria_id=?, modalidad_cobro=?, precio_base=?, precio_hora=?, precio_dia=?, precio_semana=?, precio_mes=?, kilometros_incluidos=?, deposito_minimo=?, estado=?, fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=? WHERE empresa_id=? AND id=?`, item.Codigo, item.Nombre, item.CategoriaID, item.ModalidadCobro, item.PrecioBase, item.PrecioHora, item.PrecioDia, item.PrecioSemana, item.PrecioMes, item.KilometrosIncluidos, item.DepositoMinimo, item.Estado, strings.TrimSpace(item.UsuarioCreador), item.EmpresaID, item.ID)
		return item.ID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_alquileres_tarifas (empresa_id, codigo, nombre, categoria_id, modalidad_cobro, precio_base, precio_hora, precio_dia, precio_semana, precio_mes, kilometros_incluidos, deposito_minimo, estado, fecha_creacion, fecha_actualizacion, usuario_creador) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?)`, item.EmpresaID, item.Codigo, item.Nombre, item.CategoriaID, item.ModalidadCobro, item.PrecioBase, item.PrecioHora, item.PrecioDia, item.PrecioSemana, item.PrecioMes, item.KilometrosIncluidos, item.DepositoMinimo, item.Estado, strings.TrimSpace(item.UsuarioCreador))
}

func ListEmpresaAlquilerTarifas(dbConn *sql.DB, empresaID int64) ([]EmpresaAlquilerTarifa, error) {
	if err := EnsureEmpresaAlquileresSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT t.id, t.empresa_id, COALESCE(t.codigo,''), COALESCE(t.nombre,''), COALESCE(t.categoria_id,0), COALESCE(c.nombre,''), COALESCE(t.modalidad_cobro,''), COALESCE(t.precio_base,0), COALESCE(t.precio_hora,0), COALESCE(t.precio_dia,0), COALESCE(t.precio_semana,0), COALESCE(t.precio_mes,0), COALESCE(t.kilometros_incluidos,0), COALESCE(t.deposito_minimo,0), COALESCE(t.estado,''), COALESCE(t.fecha_creacion,''), COALESCE(t.fecha_actualizacion,''), COALESCE(t.usuario_creador,'')
			FROM empresa_alquileres_tarifas t
			LEFT JOIN empresa_alquileres_categorias c ON c.id = t.categoria_id
			WHERE t.empresa_id=? ORDER BY t.id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EmpresaAlquilerTarifa, 0)
	for rows.Next() {
		var item EmpresaAlquilerTarifa
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.Codigo, &item.Nombre, &item.CategoriaID, &item.CategoriaNombre, &item.ModalidadCobro, &item.PrecioBase, &item.PrecioHora, &item.PrecioDia, &item.PrecioSemana, &item.PrecioMes, &item.KilometrosIncluidos, &item.DepositoMinimo, &item.Estado, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func calculateContratoTotal(item *EmpresaAlquilerContrato) {
	bruto := item.ValorBase + item.Impuestos - item.Descuento
	if bruto < 0 {
		bruto = 0
	}
	item.Total = bruto
	if item.SaldoPendiente <= 0 {
		item.SaldoPendiente = bruto
	}
}

func CreateEmpresaAlquilerContrato(dbConn *sql.DB, item EmpresaAlquilerContrato) (int64, error) {
	if err := EnsureEmpresaAlquileresSchema(dbConn); err != nil {
		return 0, err
	}
	item.Codigo = strings.ToUpper(strings.TrimSpace(item.Codigo))
	item.ClienteNombre = strings.TrimSpace(item.ClienteNombre)
	if item.Codigo == "" || item.ClienteNombre == "" || item.ActivoID <= 0 {
		return 0, fmt.Errorf("codigo, cliente y activo son obligatorios")
	}
	item.TipoRegistro = strings.TrimSpace(item.TipoRegistro)
	if item.TipoRegistro == "" {
		item.TipoRegistro = "alquiler"
	}
	item.ModalidadCobro = normalizeModalidadCobro(item.ModalidadCobro)
	item.Estado = normalizeAlquilerEstado(item.Estado)
	if item.Cantidad <= 0 {
		item.Cantidad = 1
	}
	calculateContratoTotal(&item)
	if item.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_alquileres_contratos SET codigo=?, tipo_registro=?, activo_id=?, cliente_nombre=?, cliente_documento=?, cliente_telefono=?, cliente_email=?, responsable_empresa=?, tarifa_id=?, modalidad_cobro=?, fecha_reserva=?, fecha_inicio=?, fecha_fin_prevista=?, estado=?, cantidad=?, horas_planeadas=?, dias_planeados=?, kilometros_incluidos=?, deposito=?, valor_base=?, descuento=?, impuestos=?, total=?, saldo_pendiente=?, origen_entrega=?, destino_devolucion=?, observaciones=?, requiere_garantia=?, gps_tracking_activo=?, fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=? WHERE empresa_id=? AND id=?`,
			item.Codigo, item.TipoRegistro, item.ActivoID, item.ClienteNombre, strings.TrimSpace(item.ClienteDocumento), strings.TrimSpace(item.ClienteTelefono), strings.TrimSpace(item.ClienteEmail), strings.TrimSpace(item.ResponsableEmpresa), item.TarifaID, item.ModalidadCobro, strings.TrimSpace(item.FechaReserva), strings.TrimSpace(item.FechaInicio), strings.TrimSpace(item.FechaFinPrevista), item.Estado, item.Cantidad, item.HorasPlaneadas, item.DiasPlaneados, item.KilometrosIncluidos, item.Deposito, item.ValorBase, item.Descuento, item.Impuestos, item.Total, item.SaldoPendiente, strings.TrimSpace(item.OrigenEntrega), strings.TrimSpace(item.DestinoDevolucion), strings.TrimSpace(item.Observaciones), boolInt(item.RequiereGarantia), boolInt(item.GpsTrackingActivo), strings.TrimSpace(item.UsuarioCreador), item.EmpresaID, item.ID)
		if err != nil {
			return 0, err
		}
		_ = syncActivoEstadoByContrato(dbConn, item.EmpresaID, item.ID, item.Estado)
		return item.ID, nil
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_alquileres_contratos (empresa_id, codigo, tipo_registro, activo_id, cliente_nombre, cliente_documento, cliente_telefono, cliente_email, responsable_empresa, tarifa_id, modalidad_cobro, fecha_reserva, fecha_inicio, fecha_fin_prevista, estado, cantidad, horas_planeadas, dias_planeados, kilometros_incluidos, deposito, valor_base, descuento, impuestos, total, saldo_pendiente, origen_entrega, destino_devolucion, observaciones, requiere_garantia, gps_tracking_activo, fecha_creacion, fecha_actualizacion, usuario_creador) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?)`,
		item.EmpresaID, item.Codigo, item.TipoRegistro, item.ActivoID, item.ClienteNombre, strings.TrimSpace(item.ClienteDocumento), strings.TrimSpace(item.ClienteTelefono), strings.TrimSpace(item.ClienteEmail), strings.TrimSpace(item.ResponsableEmpresa), item.TarifaID, item.ModalidadCobro, strings.TrimSpace(item.FechaReserva), strings.TrimSpace(item.FechaInicio), strings.TrimSpace(item.FechaFinPrevista), item.Estado, item.Cantidad, item.HorasPlaneadas, item.DiasPlaneados, item.KilometrosIncluidos, item.Deposito, item.ValorBase, item.Descuento, item.Impuestos, item.Total, item.SaldoPendiente, strings.TrimSpace(item.OrigenEntrega), strings.TrimSpace(item.DestinoDevolucion), strings.TrimSpace(item.Observaciones), boolInt(item.RequiereGarantia), boolInt(item.GpsTrackingActivo), strings.TrimSpace(item.UsuarioCreador))
	if err != nil {
		return 0, err
	}
	_ = syncActivoEstadoByContrato(dbConn, item.EmpresaID, id, item.Estado)
	return id, nil
}

func syncActivoEstadoByContrato(dbConn *sql.DB, empresaID, contratoID int64, estado string) error {
	var activoID int64
	if err := QueryRowCompat(dbConn, `SELECT COALESCE(activo_id,0) FROM empresa_alquileres_contratos WHERE empresa_id=? AND id=?`, empresaID, contratoID).Scan(&activoID); err != nil {
		return err
	}
	if activoID <= 0 {
		return nil
	}
	activoEstado := "disponible"
	switch normalizeAlquilerEstado(estado) {
	case "reservado":
		activoEstado = "reservado"
	case "en_curso":
		activoEstado = "alquilado"
	case "devuelto":
		activoEstado = "disponible"
	case "cancelado":
		activoEstado = "disponible"
	case "vencido":
		activoEstado = "alquilado"
	}
	_, err := ExecCompat(dbConn, `UPDATE empresa_alquileres_activos SET estado=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, activoEstado, empresaID, activoID)
	return err
}

func ListEmpresaAlquilerContratos(dbConn *sql.DB, empresaID int64) ([]EmpresaAlquilerContrato, error) {
	if err := EnsureEmpresaAlquileresSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT c.id, c.empresa_id, COALESCE(c.codigo,''), COALESCE(c.tipo_registro,''), COALESCE(c.activo_id,0), COALESCE(a.nombre,''), COALESCE(cat.nombre,''), COALESCE(c.cliente_nombre,''), COALESCE(c.cliente_documento,''), COALESCE(c.cliente_telefono,''), COALESCE(c.cliente_email,''), COALESCE(c.responsable_empresa,''), COALESCE(c.tarifa_id,0), COALESCE(t.nombre,''), COALESCE(c.modalidad_cobro,''), COALESCE(c.fecha_reserva,''), COALESCE(c.fecha_inicio,''), COALESCE(c.fecha_fin_prevista,''), COALESCE(c.fecha_entrega_real,''), COALESCE(c.fecha_devolucion_real,''), COALESCE(c.estado,''), COALESCE(c.cantidad,1), COALESCE(c.horas_planeadas,0), COALESCE(c.dias_planeados,0), COALESCE(c.kilometros_incluidos,0), COALESCE(c.deposito,0), COALESCE(c.valor_base,0), COALESCE(c.descuento,0), COALESCE(c.impuestos,0), COALESCE(c.total,0), COALESCE(c.saldo_pendiente,0), COALESCE(c.origen_entrega,''), COALESCE(c.destino_devolucion,''), COALESCE(c.observaciones,''), COALESCE(c.requiere_garantia,0), COALESCE(c.gps_tracking_activo,0), COALESCE(c.latitud_actual,0), COALESCE(c.longitud_actual,0), COALESCE(c.fecha_ultima_ubicacion,''), COALESCE(c.fecha_creacion,''), COALESCE(c.fecha_actualizacion,''), COALESCE(c.usuario_creador,'')
			FROM empresa_alquileres_contratos c
			JOIN empresa_alquileres_activos a ON a.id = c.activo_id
			LEFT JOIN empresa_alquileres_categorias cat ON cat.id = a.categoria_id
			LEFT JOIN empresa_alquileres_tarifas t ON t.id = c.tarifa_id
			WHERE c.empresa_id=?
			ORDER BY c.id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EmpresaAlquilerContrato, 0)
	for rows.Next() {
		var item EmpresaAlquilerContrato
		var garantia, gps int
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.Codigo, &item.TipoRegistro, &item.ActivoID, &item.ActivoNombre, &item.CategoriaNombre, &item.ClienteNombre, &item.ClienteDocumento, &item.ClienteTelefono, &item.ClienteEmail, &item.ResponsableEmpresa, &item.TarifaID, &item.TarifaNombre, &item.ModalidadCobro, &item.FechaReserva, &item.FechaInicio, &item.FechaFinPrevista, &item.FechaEntregaReal, &item.FechaDevolucionReal, &item.Estado, &item.Cantidad, &item.HorasPlaneadas, &item.DiasPlaneados, &item.KilometrosIncluidos, &item.Deposito, &item.ValorBase, &item.Descuento, &item.Impuestos, &item.Total, &item.SaldoPendiente, &item.OrigenEntrega, &item.DestinoDevolucion, &item.Observaciones, &garantia, &gps, &item.LatitudActual, &item.LongitudActual, &item.FechaUltimaUbicacion, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		item.RequiereGarantia = garantia > 0
		item.GpsTrackingActivo = gps > 0
		out = append(out, item)
	}
	return out, rows.Err()
}

func UpdateEmpresaAlquilerContratoEstado(dbConn *sql.DB, empresaID, contratoID int64, estado, responsable, observaciones string) error {
	if err := EnsureEmpresaAlquileresSchema(dbConn); err != nil {
		return err
	}
	estado = normalizeAlquilerEstado(estado)
	setParts := []string{"estado = ?", "responsable_empresa = ?", "fecha_actualizacion = CURRENT_TIMESTAMP"}
	args := []interface{}{estado, strings.TrimSpace(responsable)}
	if strings.TrimSpace(observaciones) != "" {
		setParts = append(setParts, "observaciones = ?")
		args = append(args, strings.TrimSpace(observaciones))
	}
	if estado == "en_curso" {
		setParts = append(setParts, "fecha_entrega_real = COALESCE(NULLIF(fecha_entrega_real,''), CAST(CURRENT_TIMESTAMP AS TEXT))")
	}
	if estado == "devuelto" {
		setParts = append(setParts, "fecha_devolucion_real = CAST(CURRENT_TIMESTAMP AS TEXT)", "saldo_pendiente = 0")
	}
	args = append(args, empresaID, contratoID)
	_, err := ExecCompat(dbConn, `UPDATE empresa_alquileres_contratos SET `+strings.Join(setParts, ", ")+` WHERE empresa_id = ? AND id = ?`, args...)
	if err != nil {
		return err
	}
	return syncActivoEstadoByContrato(dbConn, empresaID, contratoID, estado)
}

func CreateEmpresaAlquilerMantenimiento(dbConn *sql.DB, item EmpresaAlquilerMantenimiento) (int64, error) {
	if err := EnsureEmpresaAlquileresSchema(dbConn); err != nil {
		return 0, err
	}
	if item.ActivoID <= 0 {
		return 0, fmt.Errorf("activo_id es obligatorio")
	}
	if strings.TrimSpace(item.Tipo) == "" {
		item.Tipo = "preventivo"
	}
	if strings.TrimSpace(item.Prioridad) == "" {
		item.Prioridad = "media"
	}
	if strings.TrimSpace(item.Estado) == "" {
		item.Estado = "abierto"
	}
	if item.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_alquileres_mantenimientos SET activo_id=?, tipo=?, prioridad=?, estado=?, fecha_programada=?, fecha_cierre=?, proveedor=?, costo_estimado=?, costo_real=?, descripcion=?, observaciones=?, fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=? WHERE empresa_id=? AND id=?`, item.ActivoID, item.Tipo, item.Prioridad, item.Estado, strings.TrimSpace(item.FechaProgramada), strings.TrimSpace(item.FechaCierre), strings.TrimSpace(item.Proveedor), item.CostoEstimado, item.CostoReal, strings.TrimSpace(item.Descripcion), strings.TrimSpace(item.Observaciones), strings.TrimSpace(item.UsuarioCreador), item.EmpresaID, item.ID)
		if err == nil && strings.ToLower(strings.TrimSpace(item.Estado)) != "cerrado" {
			_, _ = ExecCompat(dbConn, `UPDATE empresa_alquileres_activos SET estado='mantenimiento', fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, item.EmpresaID, item.ActivoID)
		}
		return item.ID, err
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_alquileres_mantenimientos (empresa_id, activo_id, tipo, prioridad, estado, fecha_programada, fecha_cierre, proveedor, costo_estimado, costo_real, descripcion, observaciones, fecha_creacion, fecha_actualizacion, usuario_creador) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?)`, item.EmpresaID, item.ActivoID, item.Tipo, item.Prioridad, item.Estado, strings.TrimSpace(item.FechaProgramada), strings.TrimSpace(item.FechaCierre), strings.TrimSpace(item.Proveedor), item.CostoEstimado, item.CostoReal, strings.TrimSpace(item.Descripcion), strings.TrimSpace(item.Observaciones), strings.TrimSpace(item.UsuarioCreador))
	if err == nil && strings.ToLower(strings.TrimSpace(item.Estado)) != "cerrado" {
		_, _ = ExecCompat(dbConn, `UPDATE empresa_alquileres_activos SET estado='mantenimiento', fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, item.EmpresaID, item.ActivoID)
	}
	return id, err
}

func ListEmpresaAlquilerMantenimientos(dbConn *sql.DB, empresaID int64) ([]EmpresaAlquilerMantenimiento, error) {
	if err := EnsureEmpresaAlquileresSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT m.id, m.empresa_id, m.activo_id, COALESCE(a.nombre,''), COALESCE(m.tipo,''), COALESCE(m.prioridad,''), COALESCE(m.estado,''), COALESCE(m.fecha_programada,''), COALESCE(m.fecha_cierre,''), COALESCE(m.proveedor,''), COALESCE(m.costo_estimado,0), COALESCE(m.costo_real,0), COALESCE(m.descripcion,''), COALESCE(m.observaciones,''), COALESCE(m.fecha_creacion,''), COALESCE(m.fecha_actualizacion,''), COALESCE(m.usuario_creador,'')
			FROM empresa_alquileres_mantenimientos m
			JOIN empresa_alquileres_activos a ON a.id = m.activo_id
			WHERE m.empresa_id=? ORDER BY m.id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EmpresaAlquilerMantenimiento, 0)
	for rows.Next() {
		var item EmpresaAlquilerMantenimiento
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.ActivoID, &item.ActivoNombre, &item.Tipo, &item.Prioridad, &item.Estado, &item.FechaProgramada, &item.FechaCierre, &item.Proveedor, &item.CostoEstimado, &item.CostoReal, &item.Descripcion, &item.Observaciones, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func CreateEmpresaAlquilerUbicacion(dbConn *sql.DB, item EmpresaAlquilerUbicacion) (int64, error) {
	if err := EnsureEmpresaAlquileresSchema(dbConn); err != nil {
		return 0, err
	}
	if item.Latitud == 0 && item.Longitud == 0 {
		return 0, fmt.Errorf("latitud y longitud son obligatorias")
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_alquileres_ubicaciones (empresa_id, activo_id, contrato_id, latitud, longitud, velocidad, precision_metros, fuente, referencia, fecha_registro, usuario_creador) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?)`, item.EmpresaID, item.ActivoID, item.ContratoID, item.Latitud, item.Longitud, item.Velocidad, item.PrecisionMetros, strings.TrimSpace(item.Fuente), strings.TrimSpace(item.Referencia), strings.TrimSpace(item.UsuarioCreador))
	if err != nil {
		return 0, err
	}
	if item.ActivoID > 0 {
		_, _ = ExecCompat(dbConn, `UPDATE empresa_alquileres_activos SET latitud_actual=?, longitud_actual=?, fecha_ultima_ubicacion=CAST(CURRENT_TIMESTAMP AS TEXT), fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, item.Latitud, item.Longitud, item.EmpresaID, item.ActivoID)
	}
	if item.ContratoID > 0 {
		_, _ = ExecCompat(dbConn, `UPDATE empresa_alquileres_contratos SET latitud_actual=?, longitud_actual=?, fecha_ultima_ubicacion=CAST(CURRENT_TIMESTAMP AS TEXT), fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, item.Latitud, item.Longitud, item.EmpresaID, item.ContratoID)
	}
	return id, nil
}

func ListEmpresaAlquilerUbicaciones(dbConn *sql.DB, empresaID int64, contratoID int64) ([]EmpresaAlquilerUbicacion, error) {
	if err := EnsureEmpresaAlquileresSchema(dbConn); err != nil {
		return nil, err
	}
	query := `SELECT u.id, u.empresa_id, COALESCE(u.activo_id,0), COALESCE(a.nombre,''), COALESCE(u.contrato_id,0), COALESCE(c.codigo,''), COALESCE(u.latitud,0), COALESCE(u.longitud,0), COALESCE(u.velocidad,0), COALESCE(u.precision_metros,0), COALESCE(u.fuente,''), COALESCE(u.referencia,''), COALESCE(u.fecha_registro,''), COALESCE(u.usuario_creador,'')
		FROM empresa_alquileres_ubicaciones u
		LEFT JOIN empresa_alquileres_activos a ON a.id = u.activo_id
		LEFT JOIN empresa_alquileres_contratos c ON c.id = u.contrato_id
		WHERE u.empresa_id=?`
	args := []interface{}{empresaID}
	if contratoID > 0 {
		query += ` AND u.contrato_id=?`
		args = append(args, contratoID)
	}
	query += ` ORDER BY u.id DESC LIMIT 300`
	rows, err := ExecQueryCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EmpresaAlquilerUbicacion, 0)
	for rows.Next() {
		var item EmpresaAlquilerUbicacion
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.ActivoID, &item.ActivoNombre, &item.ContratoID, &item.ContratoCodigo, &item.Latitud, &item.Longitud, &item.Velocidad, &item.PrecisionMetros, &item.Fuente, &item.Referencia, &item.FechaRegistro, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func BuildEmpresaAlquilerDashboard(dbConn *sql.DB, empresaID int64) (EmpresaAlquilerDashboard, error) {
	if err := EnsureEmpresaAlquileresSchema(dbConn); err != nil {
		return EmpresaAlquilerDashboard{}, err
	}
	row := EmpresaAlquilerDashboard{EmpresaID: empresaID}
	_ = QueryRowCompat(dbConn, `SELECT
		COALESCE(SUM(CASE WHEN estado='disponible' THEN 1 ELSE 0 END),0),
		COALESCE(SUM(CASE WHEN estado='alquilado' THEN 1 ELSE 0 END),0)
		FROM empresa_alquileres_activos WHERE empresa_id=?`, empresaID).Scan(&row.ActivosDisponibles, &row.ActivosAlquilados)
	_ = QueryRowCompat(dbConn, `SELECT
		COALESCE(SUM(CASE WHEN estado='reservado' THEN 1 ELSE 0 END),0),
		COALESCE(SUM(CASE WHEN estado='vencido' THEN 1 ELSE 0 END),0),
		COALESCE(SUM(CASE WHEN estado='devuelto' AND CAST(fecha_devolucion_real AS DATE)=CURRENT_DATE THEN 1 ELSE 0 END),0),
		COALESCE(SUM(CASE WHEN estado IN ('reservado','en_curso') THEN deposito ELSE 0 END),0),
		COALESCE(SUM(CASE WHEN CAST(fecha_creacion AS DATE) >= date_trunc('month', CURRENT_DATE) THEN total ELSE 0 END),0)
		FROM empresa_alquileres_contratos WHERE empresa_id=?`, empresaID).Scan(&row.ReservasPendientes, &row.ContratosVencidos, &row.DevolucionesHoy, &row.DepositosRetenidos, &row.IngresosMes)
	_ = QueryRowCompat(dbConn, `SELECT COALESCE(COUNT(*),0) FROM empresa_alquileres_mantenimientos WHERE empresa_id=? AND estado NOT IN ('cerrado','cancelado')`, empresaID).Scan(&row.MantenimientosAbiertos)

	contratos, _ := ListEmpresaAlquilerContratos(dbConn, empresaID)
	activos, _ := ListEmpresaAlquilerActivos(dbConn, empresaID)
	proximos := make([]EmpresaAlquilerContrato, 0)
	lineMap := map[string]*EmpresaAlquilerResumenGrupo{}
	sedeMap := map[string]*EmpresaAlquilerResumenGrupo{}
	activosRiesgo := make([]EmpresaAlquilerActivo, 0)
	activosMapa := map[int64]EmpresaAlquilerActivo{}
	for _, a := range activos {
		activosMapa[a.ID] = a
		if a.Estado == "mantenimiento" || (a.UsaGPS && a.LatitudActual == 0 && a.LongitudActual == 0) {
			activosRiesgo = append(activosRiesgo, a)
		}
	}
	activosActivos := 0
	activosOcupados := 0
	for _, a := range activos {
		activosActivos++
		if a.Estado == "alquilado" || a.Estado == "reservado" {
			activosOcupados++
		}
	}
	if activosActivos > 0 {
		row.UtilizacionPromedio = float64(activosOcupados) * 100 / float64(activosActivos)
	}
	for _, c := range contratos {
		if c.Estado == "reservado" || c.Estado == "en_curso" || c.Estado == "vencido" {
			proximos = append(proximos, c)
		}
		lineKey := strings.TrimSpace(c.CategoriaNombre)
		if lineKey == "" {
			lineKey = "Sin categoría"
		}
		if _, ok := lineMap[lineKey]; !ok {
			lineMap[lineKey] = &EmpresaAlquilerResumenGrupo{Clave: lineKey, Etiqueta: lineKey}
		}
		lineMap[lineKey].Cantidad++
		lineMap[lineKey].Monto += c.Total
		sedeKey := strings.TrimSpace(activosMapa[c.ActivoID].Sede)
		if sedeKey == "" {
			sedeKey = "principal"
		}
		if _, ok := sedeMap[sedeKey]; !ok {
			sedeMap[sedeKey] = &EmpresaAlquilerResumenGrupo{Clave: sedeKey, Etiqueta: sedeKey}
		}
		sedeMap[sedeKey].Cantidad++
		sedeMap[sedeKey].Monto += c.Total
	}
	row.ProximosVencimientos = limitAlquilerContratos(proximos, 8)
	row.ActivosEnRiesgo = limitAlquilerActivos(activosRiesgo, 8)
	row.IngresosPorLinea = flattenAlquilerResumen(lineMap)
	row.IngresosPorSede = flattenAlquilerResumen(sedeMap)
	return row, nil
}

func flattenAlquilerResumen(m map[string]*EmpresaAlquilerResumenGrupo) []EmpresaAlquilerResumenGrupo {
	out := make([]EmpresaAlquilerResumenGrupo, 0, len(m))
	for _, item := range m {
		out = append(out, *item)
	}
	return out
}

func limitAlquilerContratos(items []EmpresaAlquilerContrato, max int) []EmpresaAlquilerContrato {
	if len(items) <= max {
		return items
	}
	return items[:max]
}

func limitAlquilerActivos(items []EmpresaAlquilerActivo, max int) []EmpresaAlquilerActivo {
	if len(items) <= max {
		return items
	}
	return items[:max]
}

func SeedEmpresaAlquilerDemoData(dbConn *sql.DB, empresaID int64, usuario string) error {
	if err := EnsureEmpresaAlquileresSchema(dbConn); err != nil {
		return err
	}
	cats, err := ListEmpresaAlquilerCategorias(dbConn, empresaID)
	if err != nil {
		return err
	}
	if len(cats) > 0 {
		return nil
	}
	categoriaID, err := CreateEmpresaAlquilerCategoria(dbConn, EmpresaAlquilerCategoria{EmpresaID: empresaID, Codigo: "EQP", Nombre: "Equipos y herramientas", TipoActivo: "herramienta", UsuarioCreador: usuario})
	if err != nil {
		return err
	}
	tarifaID, err := CreateEmpresaAlquilerTarifa(dbConn, EmpresaAlquilerTarifa{EmpresaID: empresaID, Codigo: "DIA-STD", Nombre: "Tarifa diaria estándar", CategoriaID: categoriaID, ModalidadCobro: "dia", PrecioDia: 85000, DepositoMinimo: 120000, Estado: "activa", UsuarioCreador: usuario})
	if err != nil {
		return err
	}
	activoID, err := CreateEmpresaAlquilerActivo(dbConn, EmpresaAlquilerActivo{EmpresaID: empresaID, Codigo: "ALQ-001", Nombre: "Martillo demoledor Bosch", CategoriaID: categoriaID, TipoActivo: "herramienta", Marca: "Bosch", Modelo: "GSH 11", Sede: "principal", Estado: "disponible", ValorReposicion: 4200000, DepositoSugerido: 180000, CostoBaseHora: 12000, RequiereChecklist: true, UsuarioCreador: usuario})
	if err != nil {
		return err
	}
	_, err = CreateEmpresaAlquilerContrato(dbConn, EmpresaAlquilerContrato{EmpresaID: empresaID, Codigo: fmt.Sprintf("ALQ-%d", time.Now().Unix()%1000000), TipoRegistro: "reserva", ActivoID: activoID, ClienteNombre: "Cliente demo", ClienteTelefono: "3000000000", TarifaID: tarifaID, ModalidadCobro: "dia", FechaReserva: time.Now().Format("2006-01-02 15:04:05"), FechaInicio: time.Now().Format("2006-01-02 15:04:05"), FechaFinPrevista: time.Now().Add(48 * time.Hour).Format("2006-01-02 15:04:05"), Estado: "reservado", DiasPlaneados: 2, Deposito: 180000, ValorBase: 170000, Impuestos: 32300, UsuarioCreador: usuario})
	return err
}
