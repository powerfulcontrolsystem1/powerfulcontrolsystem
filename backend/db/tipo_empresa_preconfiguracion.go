package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// TipoEmpresaPreconfiguracion define la plantilla que se aplica al crear una empresa por tipo.
type TipoEmpresaPreconfiguracion struct {
	ID                 int64  `json:"id"`
	TipoEmpresaID      int64  `json:"tipo_empresa_id"`
	TipoEmpresaNombre  string `json:"tipo_empresa_nombre,omitempty"`
	Enabled            bool   `json:"enabled"`
	Nombre             string `json:"nombre"`
	Descripcion        string `json:"descripcion,omitempty"`
	ConfigJSON         string `json:"config_json"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
}

type TipoEmpresaPreconfigTemplate struct {
	Estaciones TipoEmpresaPreconfigEstaciones  `json:"estaciones"`
	Operacion  TipoEmpresaPreconfigOperacion   `json:"operacion,omitempty"`
	Productos  []TipoEmpresaPreconfigProducto  `json:"productos"`
	Usuarios   []TipoEmpresaPreconfigUsuario   `json:"usuarios,omitempty"`
	Tarifas    TipoEmpresaPreconfigTarifas     `json:"tarifas,omitempty"`
	Modulos    TipoEmpresaPreconfigModulos     `json:"modulos,omitempty"`
	Asistente  TipoEmpresaPreconfigAsistenteIA `json:"asistente_ia,omitempty"`
	TareasGuia []TipoEmpresaPreconfigTareaGuia `json:"tareas_guia,omitempty"`
}

type TipoEmpresaPreconfigEstaciones struct {
	Enabled     bool   `json:"enabled"`
	Cantidad    int    `json:"cantidad"`
	Prefijo     string `json:"prefijo"`
	CardSize    string `json:"card_size"`
	CajaEnabled bool   `json:"caja_enabled"`
}

type TipoEmpresaPreconfigOperacion struct {
	TipoNegocio            string   `json:"tipo_negocio,omitempty"`
	NombreEstacionSingular string   `json:"nombre_estacion_singular,omitempty"`
	NombreEstacionPlural   string   `json:"nombre_estacion_plural,omitempty"`
	UsaEstaciones          bool     `json:"usa_estaciones"`
	VentaDirectaEnabled    bool     `json:"venta_directa_enabled"`
	VentaDirectaNombre     string   `json:"venta_directa_nombre,omitempty"`
	ComisionesEnabled      bool     `json:"comisiones_enabled"`
	ComisionRol            string   `json:"comision_rol,omitempty"`
	ComisionFiltro         string   `json:"comision_filtro,omitempty"`
	ComisionPorcentaje     float64  `json:"comision_porcentaje,omitempty"`
	RolesOperativos        []string `json:"roles_operativos,omitempty"`
}

type TipoEmpresaPreconfigProducto struct {
	SKU                  string  `json:"sku"`
	Nombre               string  `json:"nombre"`
	Categoria            string  `json:"categoria,omitempty"`
	Descripcion          string  `json:"descripcion,omitempty"`
	UnidadMedida         string  `json:"unidad_medida,omitempty"`
	Costo                float64 `json:"costo"`
	Precio               float64 `json:"precio"`
	ImpuestoPorcentaje   float64 `json:"impuesto_porcentaje"`
	StockMinimo          float64 `json:"stock_minimo"`
	StockInicial         float64 `json:"stock_inicial"`
	ReferenciaInventario string  `json:"referencia_inventario,omitempty"`
}

type TipoEmpresaPreconfigUsuario struct {
	Nombre        string `json:"nombre"`
	Email         string `json:"email,omitempty"`
	Rol           string `json:"rol"`
	Observaciones string `json:"observaciones,omitempty"`
}

type TipoEmpresaPreconfigAsistenteIA struct {
	Enabled       bool     `json:"enabled"`
	Rol           string   `json:"rol,omitempty"`
	Instrucciones []string `json:"instrucciones,omitempty"`
}

type TipoEmpresaPreconfigTareaGuia struct {
	Modulo      string `json:"modulo"`
	Titulo      string `json:"titulo"`
	Descripcion string `json:"descripcion,omitempty"`
}

type TipoEmpresaPreconfigTarifas struct {
	PorMinutos []TipoEmpresaPreconfigTarifaPorMinutos `json:"por_minutos,omitempty"`
	PorDia     []TipoEmpresaPreconfigTarifaPorDia     `json:"por_dia,omitempty"`
	Motel      []TipoEmpresaPreconfigTarifaMotel      `json:"motel,omitempty"`
}

type TipoEmpresaPreconfigTarifaPorMinutos struct {
	EstacionNumero    int     `json:"estacion_numero,omitempty"`
	DiaSemanaDesde    int     `json:"dia_semana_desde"`
	DiaSemanaHasta    int     `json:"dia_semana_hasta"`
	MinutosBase       int     `json:"minutos_base"`
	ValorBase         float64 `json:"valor_base"`
	MinutosExtra      int     `json:"minutos_extra"`
	ValorExtra        float64 `json:"valor_extra"`
	CobrarPorFraccion bool    `json:"cobrar_por_fraccion"`
	Moneda            string  `json:"moneda,omitempty"`
	Prioridad         int     `json:"prioridad,omitempty"`
	Observaciones     string  `json:"observaciones,omitempty"`
}

type TipoEmpresaPreconfigTarifaPorDia struct {
	EstacionNumero         int     `json:"estacion_numero,omitempty"`
	NombreTarifa           string  `json:"nombre_tarifa,omitempty"`
	ServicioNombre         string  `json:"servicio_nombre,omitempty"`
	ValorDia               float64 `json:"valor_dia"`
	PersonasDesde          int     `json:"personas_desde"`
	PersonasHasta          int     `json:"personas_hasta"`
	HoraCheckIn            string  `json:"hora_check_in,omitempty"`
	HoraCheckOut           string  `json:"hora_check_out,omitempty"`
	Moneda                 string  `json:"moneda,omitempty"`
	Prioridad              int     `json:"prioridad,omitempty"`
	AplicarAutomaticamente bool    `json:"aplicar_automaticamente"`
	Observaciones          string  `json:"observaciones,omitempty"`
}

type TipoEmpresaPreconfigTarifaMotel struct {
	EstacionNumero      int     `json:"estacion_numero,omitempty"`
	NombrePlan          string  `json:"nombre_plan"`
	TipoPlan            string  `json:"tipo_plan"`
	CategoriaHabitacion string  `json:"categoria_habitacion,omitempty"`
	DiaSemanaDesde      int     `json:"dia_semana_desde"`
	DiaSemanaHasta      int     `json:"dia_semana_hasta"`
	HoraInicio          string  `json:"hora_inicio,omitempty"`
	HoraFin             string  `json:"hora_fin,omitempty"`
	MinutosIncluidos    int     `json:"minutos_incluidos"`
	ValorBase           float64 `json:"valor_base"`
	MinutosExtra        int     `json:"minutos_extra"`
	ValorExtra          float64 `json:"valor_extra"`
	CobrarPorFraccion   bool    `json:"cobrar_por_fraccion"`
	ToleranciaMinutos   int     `json:"tolerancia_minutos"`
	Moneda              string  `json:"moneda,omitempty"`
	Prioridad           int     `json:"prioridad,omitempty"`
	AplicarAutomatico   bool    `json:"aplicar_automaticamente"`
	Observaciones       string  `json:"observaciones,omitempty"`
}

type TipoEmpresaPreconfigModulos struct {
	TurnosAtencion   *TipoEmpresaPreconfigTurnosAtencion   `json:"turnos_atencion,omitempty"`
	Gimnasio         *TipoEmpresaPreconfigGimnasio         `json:"gimnasio,omitempty"`
	Odontologia      *TipoEmpresaPreconfigOdontologia      `json:"odontologia,omitempty"`
	Vehiculos        *TipoEmpresaPreconfigVehiculos        `json:"vehiculos,omitempty"`
	ControlElectrico *TipoEmpresaPreconfigControlElectrico `json:"control_electrico,omitempty"`
	HojaVida         []TipoEmpresaPreconfigHojaVida        `json:"hoja_vida,omitempty"`
}

type TipoEmpresaPreconfigTurnosAtencion struct {
	NombreSistema             string                                      `json:"nombre_sistema,omitempty"`
	NombrePantalla            string                                      `json:"nombre_pantalla,omitempty"`
	PrefijoGeneral            string                                      `json:"prefijo_general,omitempty"`
	TiempoLlamadoSegundos     int                                         `json:"tiempo_llamado_segundos,omitempty"`
	PermitirEmisionPublica    bool                                        `json:"permitir_emision_publica"`
	MostrarTicketsCompletados bool                                        `json:"mostrar_tickets_completados"`
	Servicios                 []TipoEmpresaPreconfigTurnoAtencionServicio `json:"servicios,omitempty"`
	Puestos                   []TipoEmpresaPreconfigTurnoAtencionPuesto   `json:"puestos,omitempty"`
}

type TipoEmpresaPreconfigTurnoAtencionServicio struct {
	Codigo      string `json:"codigo"`
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion,omitempty"`
	Prefijo     string `json:"prefijo"`
	Prioridad   int    `json:"prioridad,omitempty"`
	Color       string `json:"color,omitempty"`
}

type TipoEmpresaPreconfigTurnoAtencionPuesto struct {
	Codigo              string `json:"codigo"`
	Nombre              string `json:"nombre"`
	Area                string `json:"area,omitempty"`
	Ubicacion           string `json:"ubicacion,omitempty"`
	ServiciosPermitidos string `json:"servicios_permitidos,omitempty"`
}

type TipoEmpresaPreconfigGimnasio struct {
	Planes       []TipoEmpresaPreconfigGimnasioPlan       `json:"planes,omitempty"`
	Entrenadores []TipoEmpresaPreconfigGimnasioEntrenador `json:"entrenadores,omitempty"`
	Clases       []TipoEmpresaPreconfigGimnasioClase      `json:"clases,omitempty"`
	Socios       []TipoEmpresaPreconfigGimnasioSocio      `json:"socios,omitempty"`
}

type TipoEmpresaPreconfigGimnasioPlan struct {
	Nombre                 string  `json:"nombre"`
	Descripcion            string  `json:"descripcion,omitempty"`
	Precio                 float64 `json:"precio"`
	DuracionDias           int     `json:"duracion_dias"`
	ClasesIncluidas        int     `json:"clases_incluidas"`
	AccesoIlimitado        bool    `json:"acceso_ilimitado"`
	SesionesPersonalizadas int     `json:"sesiones_personalizadas"`
}

type TipoEmpresaPreconfigGimnasioEntrenador struct {
	NombreCompleto  string `json:"nombre_completo"`
	Especialidad    string `json:"especialidad,omitempty"`
	Telefono        string `json:"telefono,omitempty"`
	Email           string `json:"email,omitempty"`
	Certificaciones string `json:"certificaciones,omitempty"`
	Disponibilidad  string `json:"disponibilidad,omitempty"`
	Observaciones   string `json:"observaciones,omitempty"`
}

type TipoEmpresaPreconfigGimnasioClase struct {
	Nombre          string  `json:"nombre"`
	Categoria       string  `json:"categoria,omitempty"`
	EntrenadorIndex int     `json:"entrenador_index,omitempty"`
	Sede            string  `json:"sede,omitempty"`
	Canal           string  `json:"canal,omitempty"`
	Cupos           int     `json:"cupos"`
	DuracionMinutos int     `json:"duracion_minutos"`
	Precio          float64 `json:"precio"`
	Descripcion     string  `json:"descripcion,omitempty"`
}

type TipoEmpresaPreconfigGimnasioSocio struct {
	Codigo         string `json:"codigo,omitempty"`
	NombreCompleto string `json:"nombre_completo"`
	Documento      string `json:"documento,omitempty"`
	Telefono       string `json:"telefono,omitempty"`
	Email          string `json:"email,omitempty"`
	Objetivo       string `json:"objetivo,omitempty"`
	PlanIndex      int    `json:"plan_index,omitempty"`
	Observaciones  string `json:"observaciones,omitempty"`
}

type TipoEmpresaPreconfigOdontologia struct {
	Pacientes     []TipoEmpresaPreconfigOdontoPaciente    `json:"pacientes,omitempty"`
	Profesionales []TipoEmpresaPreconfigOdontoProfesional `json:"profesionales,omitempty"`
	Consultorios  []TipoEmpresaPreconfigOdontoConsultorio `json:"consultorios,omitempty"`
	Tratamientos  []TipoEmpresaPreconfigOdontoTratamiento `json:"tratamientos,omitempty"`
}

type TipoEmpresaPreconfigOdontoPaciente struct {
	Codigo         string  `json:"codigo,omitempty"`
	NombreCompleto string  `json:"nombre_completo"`
	Documento      string  `json:"documento,omitempty"`
	Telefono       string  `json:"telefono,omitempty"`
	Email          string  `json:"email,omitempty"`
	Aseguradora    string  `json:"aseguradora,omitempty"`
	Alergias       string  `json:"alergias,omitempty"`
	RiesgoMedico   string  `json:"riesgo_medico,omitempty"`
	Saldo          float64 `json:"saldo,omitempty"`
	Observaciones  string  `json:"observaciones,omitempty"`
}

type TipoEmpresaPreconfigOdontoProfesional struct {
	NombreCompleto      string `json:"nombre_completo"`
	Especialidad        string `json:"especialidad,omitempty"`
	RegistroProfesional string `json:"registro_profesional,omitempty"`
	Telefono            string `json:"telefono,omitempty"`
	Email               string `json:"email,omitempty"`
	ColorAgenda         string `json:"color_agenda,omitempty"`
	Observaciones       string `json:"observaciones,omitempty"`
}

type TipoEmpresaPreconfigOdontoConsultorio struct {
	Nombre        string `json:"nombre"`
	Sede          string `json:"sede,omitempty"`
	Sillon        string `json:"sillon,omitempty"`
	Observaciones string `json:"observaciones,omitempty"`
}

type TipoEmpresaPreconfigOdontoTratamiento struct {
	PacienteIndex    int     `json:"paciente_index,omitempty"`
	ProfesionalIndex int     `json:"profesional_index,omitempty"`
	Nombre           string  `json:"nombre"`
	Categoria        string  `json:"categoria,omitempty"`
	Piezas           string  `json:"piezas,omitempty"`
	SesionesTotal    int     `json:"sesiones_total"`
	CostoEstimado    float64 `json:"costo_estimado"`
	Observaciones    string  `json:"observaciones,omitempty"`
}

type TipoEmpresaPreconfigVehiculos struct {
	PaisCodigo            string                                 `json:"pais_codigo,omitempty"`
	EvitarDuplicadoActivo bool                                   `json:"evitar_duplicado_activo"`
	Registros             []TipoEmpresaPreconfigVehiculoRegistro `json:"registros,omitempty"`
}

type TipoEmpresaPreconfigVehiculoRegistro struct {
	Patente              string `json:"patente"`
	TipoVehiculo         string `json:"tipo_vehiculo,omitempty"`
	Marca                string `json:"marca,omitempty"`
	Modelo               string `json:"modelo,omitempty"`
	Color                string `json:"color,omitempty"`
	PropietarioNombre    string `json:"propietario_nombre,omitempty"`
	PropietarioDocumento string `json:"propietario_documento,omitempty"`
	ConductorNombre      string `json:"conductor_nombre,omitempty"`
	MotivoIngreso        string `json:"motivo_ingreso,omitempty"`
	Observaciones        string `json:"observaciones,omitempty"`
}

type TipoEmpresaPreconfigControlElectrico struct {
	Habilitado         bool                                       `json:"habilitado"`
	RaspberryIP        string                                     `json:"raspberry_ip,omitempty"`
	RaspberryPort      int                                        `json:"raspberry_port,omitempty"`
	APIPath            string                                     `json:"api_path,omitempty"`
	TimeoutMS          int                                        `json:"timeout_ms,omitempty"`
	AutoSyncEstaciones bool                                       `json:"auto_sync_estaciones"`
	FailSafeOnError    bool                                       `json:"fail_safe_on_error"`
	Raspberries        []TipoEmpresaPreconfigControlRaspberry     `json:"raspberries,omitempty"`
	Reles              []TipoEmpresaPreconfigControlElectricoRele `json:"reles,omitempty"`
}

type TipoEmpresaPreconfigControlRaspberry struct {
	Codigo        string `json:"codigo"`
	Nombre        string `json:"nombre"`
	RaspberryIP   string `json:"raspberry_ip"`
	RaspberryPort int    `json:"raspberry_port,omitempty"`
	APIPath       string `json:"api_path,omitempty"`
	TimeoutMS     int    `json:"timeout_ms,omitempty"`
	Observaciones string `json:"observaciones,omitempty"`
}

type TipoEmpresaPreconfigControlElectricoRele struct {
	RaspberryCodigo        string `json:"raspberry_codigo,omitempty"`
	EstacionNumero         int    `json:"estacion_numero,omitempty"`
	SalidaCodigo           string `json:"salida_codigo"`
	TipoCarga              string `json:"tipo_carga,omitempty"`
	GPIOPin                int    `json:"gpio_pin"`
	RelayName              string `json:"relay_name"`
	ActiveHigh             bool   `json:"active_high"`
	PulsoMS                int    `json:"pulso_ms,omitempty"`
	Modo                   string `json:"modo,omitempty"`
	ProgramacionHabilitada bool   `json:"programacion_habilitada"`
	HoraEncendido          string `json:"hora_encendido,omitempty"`
	HoraApagado            string `json:"hora_apagado,omitempty"`
	ProgramacionDias       string `json:"programacion_dias,omitempty"`
	ProgramacionTimezone   string `json:"programacion_timezone,omitempty"`
	ImagenURL              string `json:"imagen_url,omitempty"`
	Observaciones          string `json:"observaciones,omitempty"`
}

type TipoEmpresaPreconfigHojaVida struct {
	TipoEntidad     string                               `json:"tipo_entidad"`
	Codigo          string                               `json:"codigo"`
	Nombre          string                               `json:"nombre"`
	ClienteNombre   string                               `json:"cliente_nombre,omitempty"`
	Identificacion  string                               `json:"identificacion,omitempty"`
	Marca           string                               `json:"marca,omitempty"`
	Modelo          string                               `json:"modelo,omitempty"`
	Serie           string                               `json:"serie,omitempty"`
	Color           string                               `json:"color,omitempty"`
	EstadoOperativo string                               `json:"estado_operativo,omitempty"`
	Metadata        map[string]any                       `json:"metadata,omitempty"`
	Observaciones   string                               `json:"observaciones,omitempty"`
	Eventos         []TipoEmpresaPreconfigHojaVidaEvento `json:"eventos,omitempty"`
	Alertas         []TipoEmpresaPreconfigHojaVidaAlerta `json:"alertas,omitempty"`
}

type TipoEmpresaPreconfigHojaVidaEvento struct {
	TipoEvento          string  `json:"tipo_evento,omitempty"`
	Titulo              string  `json:"titulo"`
	Descripcion         string  `json:"descripcion,omitempty"`
	Costo               float64 `json:"costo,omitempty"`
	Responsable         string  `json:"responsable,omitempty"`
	DocumentoReferencia string  `json:"documento_referencia,omitempty"`
	Recurrente          bool    `json:"recurrente"`
	RecurrenciaDias     int64   `json:"recurrencia_dias,omitempty"`
	Observaciones       string  `json:"observaciones,omitempty"`
}

type TipoEmpresaPreconfigHojaVidaAlerta struct {
	Titulo        string `json:"titulo"`
	Descripcion   string `json:"descripcion,omitempty"`
	Prioridad     string `json:"prioridad,omitempty"`
	Responsable   string `json:"responsable,omitempty"`
	Observaciones string `json:"observaciones,omitempty"`
}

type TipoEmpresaPreconfigSeedItem struct {
	TipoEmpresaID     int64  `json:"tipo_empresa_id"`
	TipoEmpresaNombre string `json:"tipo_empresa_nombre"`
	PreconfigID       int64  `json:"preconfig_id,omitempty"`
	Accion            string `json:"accion"`
	Nombre            string `json:"nombre,omitempty"`
	Enabled           bool   `json:"enabled"`
	Error             string `json:"error,omitempty"`
}

type TipoEmpresaPreconfigSeedResult struct {
	TotalTipos             int                            `json:"total_tipos"`
	Creadas                int                            `json:"creadas"`
	Actualizadas           int                            `json:"actualizadas"`
	Omitidas               int                            `json:"omitidas"`
	Errores                int                            `json:"errores"`
	RolesCreados           int                            `json:"roles_creados"`
	RolesActualizados      int                            `json:"roles_actualizados"`
	PermisosConfigurados   int                            `json:"permisos_configurados"`
	PermisosPersonalizados int                            `json:"permisos_personalizados"`
	Items                  []TipoEmpresaPreconfigSeedItem `json:"items"`
}

// EnsureTipoEmpresaPreconfiguracionSchema crea/migra la tabla de plantillas por tipo de empresa.
func EnsureTipoEmpresaPreconfiguracionSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS tipo_empresa_preconfiguraciones (
			id BIGSERIAL PRIMARY KEY,
			tipo_empresa_id BIGINT NOT NULL UNIQUE,
			enabled INTEGER DEFAULT 0,
			nombre TEXT,
			descripcion TEXT,
			config_json TEXT,
			fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo'
		);`,
		`CREATE INDEX IF NOT EXISTS ix_tipo_empresa_preconfiguraciones_tipo ON tipo_empresa_preconfiguraciones(tipo_empresa_id);`,
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
		{"enabled", "INTEGER DEFAULT 0"},
		{"nombre", "TEXT"},
		{"descripcion", "TEXT"},
		{"config_json", "TEXT"},
		{"fecha_actualizacion", "TEXT"},
		{"usuario_creador", "TEXT"},
		{"estado", "TEXT DEFAULT 'activo'"},
	} {
		if err := ensureColumnIfMissing(dbConn, "tipo_empresa_preconfiguraciones", col.name, col.def); err != nil {
			return err
		}
	}
	return nil
}

func scanTipoEmpresaPreconfiguracion(row scanner) (*TipoEmpresaPreconfiguracion, error) {
	var item TipoEmpresaPreconfiguracion
	var enabled int
	if err := row.Scan(
		&item.ID,
		&item.TipoEmpresaID,
		&item.TipoEmpresaNombre,
		&enabled,
		&item.Nombre,
		&item.Descripcion,
		&item.ConfigJSON,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
	); err != nil {
		return nil, err
	}
	item.Enabled = enabled == 1
	return &item, nil
}

type scanner interface {
	Scan(dest ...interface{}) error
}

// ListTipoEmpresaPreconfiguraciones devuelve las plantillas guardadas.
func ListTipoEmpresaPreconfiguraciones(dbConn *sql.DB) ([]TipoEmpresaPreconfiguracion, error) {
	if err := EnsureTipoEmpresaPreconfiguracionSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		p.id, p.tipo_empresa_id, COALESCE(t.nombre, ''), COALESCE(p.enabled, 0),
		COALESCE(p.nombre, ''), COALESCE(p.descripcion, ''), COALESCE(p.config_json, ''),
		COALESCE(CAST(p.fecha_creacion AS TEXT), ''), COALESCE(CAST(p.fecha_actualizacion AS TEXT), ''),
		COALESCE(p.usuario_creador, ''), COALESCE(NULLIF(TRIM(p.estado), ''), 'activo')
	FROM tipo_empresa_preconfiguraciones p
	LEFT JOIN tipos_de_empresas t ON t.id = p.tipo_empresa_id
	ORDER BY p.tipo_empresa_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]TipoEmpresaPreconfiguracion, 0)
	for rows.Next() {
		item, err := scanTipoEmpresaPreconfiguracion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

// GetTipoEmpresaPreconfiguracionByTipoID devuelve una plantilla por tipo, o nil si no existe.
func GetTipoEmpresaPreconfiguracionByTipoID(dbConn *sql.DB, tipoEmpresaID int64) (*TipoEmpresaPreconfiguracion, error) {
	if tipoEmpresaID <= 0 {
		return nil, nil
	}
	if err := EnsureTipoEmpresaPreconfiguracionSchema(dbConn); err != nil {
		return nil, err
	}
	row := queryRowSQLCompat(dbConn, `SELECT
		p.id, p.tipo_empresa_id, COALESCE(t.nombre, ''), COALESCE(p.enabled, 0),
		COALESCE(p.nombre, ''), COALESCE(p.descripcion, ''), COALESCE(p.config_json, ''),
		COALESCE(CAST(p.fecha_creacion AS TEXT), ''), COALESCE(CAST(p.fecha_actualizacion AS TEXT), ''),
		COALESCE(p.usuario_creador, ''), COALESCE(NULLIF(TRIM(p.estado), ''), 'activo')
	FROM tipo_empresa_preconfiguraciones p
	LEFT JOIN tipos_de_empresas t ON t.id = p.tipo_empresa_id
	WHERE p.tipo_empresa_id = ? LIMIT 1`, tipoEmpresaID)
	item, err := scanTipoEmpresaPreconfiguracion(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return item, err
}

// UpsertTipoEmpresaPreconfiguracion crea o actualiza una plantilla por tipo de empresa.
func UpsertTipoEmpresaPreconfiguracion(dbConn *sql.DB, item TipoEmpresaPreconfiguracion) (int64, error) {
	if item.TipoEmpresaID <= 0 {
		return 0, errors.New("tipo_empresa_id invalido")
	}
	if err := EnsureTipoEmpresaPreconfiguracionSchema(dbConn); err != nil {
		return 0, err
	}
	enabled := 0
	if item.Enabled {
		enabled = 1
	}
	item.Nombre = strings.TrimSpace(item.Nombre)
	if item.Nombre == "" {
		item.Nombre = "Preconfiguracion inicial"
	}
	item.Estado = strings.ToLower(strings.TrimSpace(item.Estado))
	if item.Estado == "" {
		item.Estado = "activo"
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO tipo_empresa_preconfiguraciones (
		tipo_empresa_id, enabled, nombre, descripcion, config_json,
		fecha_creacion, fecha_actualizacion, usuario_creador, estado
	) VALUES (
		?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?
	) ON CONFLICT(tipo_empresa_id) DO UPDATE SET
		enabled = excluded.enabled,
		nombre = excluded.nombre,
		descripcion = excluded.descripcion,
		config_json = excluded.config_json,
		fecha_actualizacion = CURRENT_TIMESTAMP,
		usuario_creador = CASE WHEN trim(excluded.usuario_creador) <> '' THEN excluded.usuario_creador ELSE tipo_empresa_preconfiguraciones.usuario_creador END,
		estado = COALESCE(NULLIF(TRIM(excluded.estado), ''), 'activo')
	RETURNING id`,
		item.TipoEmpresaID,
		enabled,
		item.Nombre,
		strings.TrimSpace(item.Descripcion),
		strings.TrimSpace(item.ConfigJSON),
		strings.TrimSpace(item.UsuarioCreador),
		item.Estado,
	)
	if err != nil {
		return 0, fmt.Errorf("upsert tipo empresa preconfiguracion: %w", err)
	}
	return id, nil
}

// SeedDefaultTipoEmpresaPreconfiguraciones registra plantillas base para todos
// los tipos existentes. Por defecto respeta configuraciones ya personalizadas.
func SeedDefaultTipoEmpresaPreconfiguraciones(dbConn *sql.DB, usuario string, overwrite bool) (*TipoEmpresaPreconfigSeedResult, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if err := EnsureTipoEmpresaPreconfiguracionSchema(dbConn); err != nil {
		return nil, err
	}
	if err := EnsureCanonicalTiposEmpresaPreconfigurables(dbConn); err != nil {
		return nil, err
	}
	tipos, err := GetTiposEmpresas(dbConn)
	if err != nil {
		return nil, err
	}
	if !overwrite {
		if skip, err := canSkipDefaultTipoEmpresaSeed(dbConn, len(tipos)); err == nil && skip {
			return &TipoEmpresaPreconfigSeedResult{
				TotalTipos: len(tipos),
				Omitidas:   len(tipos),
				Items:      make([]TipoEmpresaPreconfigSeedItem, 0, len(tipos)),
			}, nil
		}
	}
	saved, err := ListTipoEmpresaPreconfiguraciones(dbConn)
	if err != nil {
		return nil, err
	}
	byTipo := make(map[int64]TipoEmpresaPreconfiguracion, len(saved))
	for _, item := range saved {
		byTipo[item.TipoEmpresaID] = item
	}
	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = "sistema.preconfiguracion"
	}
	result := &TipoEmpresaPreconfigSeedResult{
		TotalTipos: len(tipos),
		Items:      make([]TipoEmpresaPreconfigSeedItem, 0, len(tipos)),
	}
	for _, tipo := range tipos {
		item := TipoEmpresaPreconfigSeedItem{
			TipoEmpresaID:     tipo.ID,
			TipoEmpresaNombre: strings.TrimSpace(tipo.Nombre),
		}
		if existing, ok := byTipo[tipo.ID]; ok && !overwrite {
			item.PreconfigID = existing.ID
			item.Accion = "omitida"
			item.Nombre = existing.Nombre
			item.Enabled = existing.Enabled
			result.Omitidas++
			result.Items = append(result.Items, item)
			continue
		}

		preconfig := DefaultTipoEmpresaPreconfiguracion(tipo.ID, tipo.Nombre)
		preconfig.UsuarioCreador = usuario
		preconfig.Estado = "activo"
		id, err := UpsertTipoEmpresaPreconfiguracion(dbConn, preconfig)
		if err != nil {
			item.Accion = "error"
			item.Nombre = preconfig.Nombre
			item.Enabled = preconfig.Enabled
			item.Error = err.Error()
			result.Errores++
			result.Items = append(result.Items, item)
			continue
		}
		item.PreconfigID = id
		item.Nombre = preconfig.Nombre
		item.Enabled = preconfig.Enabled
		if _, existed := byTipo[tipo.ID]; existed {
			item.Accion = "actualizada"
			result.Actualizadas++
		} else {
			item.Accion = "creada"
			result.Creadas++
		}
		result.Items = append(result.Items, item)
	}
	rolesCreados, rolesActualizados, permisosConfigurados, permisosPersonalizados, err := EnsureDefaultRolesForTipoEmpresaPreconfiguraciones(dbConn, usuario)
	if err != nil {
		return nil, err
	}
	result.RolesCreados = rolesCreados
	result.RolesActualizados = rolesActualizados
	result.PermisosConfigurados = permisosConfigurados
	result.PermisosPersonalizados = permisosPersonalizados
	return result, nil
}

func canSkipDefaultTipoEmpresaSeed(dbConn *sql.DB, totalTipos int) (bool, error) {
	if dbConn == nil || totalTipos <= 0 {
		return false, nil
	}
	requiredTables := []string{
		"tipo_empresa_preconfiguraciones",
		"roles_de_usuario",
		"roles_de_usuario_permisos",
	}
	for _, tableName := range requiredTables {
		ok, err := tableExists(dbConn, tableName)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}

	var preconfigCount int
	if err := queryRowSQLCompat(dbConn, `SELECT COUNT(1) FROM tipo_empresa_preconfiguraciones WHERE COALESCE(NULLIF(TRIM(estado), ''), 'activo') <> 'inactivo'`).Scan(&preconfigCount); err != nil {
		return false, err
	}
	if preconfigCount < totalTipos {
		return false, nil
	}

	var rolesCount int
	if err := queryRowSQLCompat(dbConn, `SELECT COUNT(1) FROM roles_de_usuario WHERE COALESCE(NULLIF(TRIM(estado), ''), 'activo') <> 'inactivo'`).Scan(&rolesCount); err != nil {
		return false, err
	}
	if rolesCount <= 0 {
		return false, nil
	}

	var permisosCount int
	if err := queryRowSQLCompat(dbConn, `SELECT COUNT(1) FROM roles_de_usuario_permisos WHERE COALESCE(NULLIF(TRIM(estado), ''), 'activo') <> 'inactivo'`).Scan(&permisosCount); err != nil {
		return false, err
	}
	return permisosCount > 0, nil
}

// EnsureCanonicalTiposEmpresaPreconfigurables registra tipos operativos base
// cuando aun no existen, para que la preconfiguracion no dependa de captura manual.
func EnsureCanonicalTiposEmpresaPreconfigurables(dbConn *sql.DB) error {
	tipos, err := GetTiposEmpresas(dbConn)
	if err != nil {
		return err
	}
	canonicos := []struct {
		nombre        string
		observaciones string
		matches       func(string) bool
	}{
		{"Restaurante", "Mesas, cocina, pedidos y venta directa.", isTipoEmpresaRestaurante},
		{"Motel", "Habitaciones por turnos, minibar, tarifas y recepcion.", isTipoEmpresaMotel},
		{"Hotel", "Habitaciones por noche, reservas, consumos y recepcion.", isTipoEmpresaHotel},
		{"Bar", "Mesas, barra, bebidas, eventos y caja.", isTipoEmpresaBar},
		{"Salon de belleza", "Sillas, estilistas, agenda, servicios y comisiones.", isTipoEmpresaSalonBelleza},
		{"Lavadero de autos", "Bahias, lavado, vehiculos, tiempos y comisiones.", isTipoEmpresaLavaderoAutos},
		{"Pymes", "Empresa general con venta directa, productos, servicios y caja.", isTipoEmpresaPyme},
		{"Punto de venta", "Una estacion principal y venta directa por mostrador.", isTipoEmpresaPuntoVenta},
		{"Taller mecanico", "Bahias, tecnicos, ordenes de servicio y comisiones.", isTipoEmpresaTaller},
		{"Gimnasio", "Planes, socios, entrenadores, clases, accesos y pagos.", isTipoEmpresaGimnasio},
		{"Odontologia", "Pacientes, profesionales, consultorios, agenda, tratamientos y presupuestos.", isTipoEmpresaOdontologia},
		{"Manejo de turnos", "Servicios, puestos, emision publica y pantalla de llamados.", isTipoEmpresaTurnos},
		{"Vehiculos y flotas", "Registro, permanencia, hoja de vida, mantenimientos y alertas de vehiculos.", isTipoEmpresaVehiculos},
		{"Tecnico independiente", "Sin estaciones; venta directa, agenda y servicios.", isTipoEmpresaIndependiente},
		{"Redes sociales", "Clientes, paquetes, tareas, contenidos y reportes.", isTipoEmpresaRedesSociales},
		{"Sensores y monitoreo", "Accesos, sensores, instalaciones y monitoreo.", isTipoEmpresaSensores},
	}
	for _, canonical := range canonicos {
		exists := false
		for _, tipo := range tipos {
			if canonical.matches(tipo.Nombre) {
				exists = true
				break
			}
		}
		if exists {
			continue
		}
		if _, err := CreateTipoEmpresa(dbConn, canonical.nombre, canonical.observaciones); err != nil {
			return fmt.Errorf("crear tipo de empresa %q: %w", canonical.nombre, err)
		}
		tipos = append(tipos, TipoEmpresa{Nombre: canonical.nombre, Observaciones: canonical.observaciones, Estado: "activo"})
	}
	return nil
}

// DefaultTipoEmpresaPreconfiguracion entrega una plantilla profesional sugerida para tipos conocidos.
func DefaultTipoEmpresaPreconfiguracion(tipoEmpresaID int64, tipoNombre string) TipoEmpresaPreconfiguracion {
	template := DefaultTipoEmpresaPreconfigTemplate(tipoNombre)
	raw, _ := json.Marshal(template)
	enabled := len(template.Productos) > 0 || template.Estaciones.Cantidad > 0 || len(template.Usuarios) > 0 || template.Asistente.Enabled || len(template.TareasGuia) > 0
	nombre := defaultTipoEmpresaPreconfigNombre(tipoNombre)
	return TipoEmpresaPreconfiguracion{
		TipoEmpresaID:     tipoEmpresaID,
		TipoEmpresaNombre: strings.TrimSpace(tipoNombre),
		Enabled:           enabled,
		Nombre:            nombre,
		Descripcion:       "Plantilla inicial aplicada automaticamente al crear empresas nuevas de este tipo.",
		ConfigJSON:        string(raw),
		Estado:            "activo",
	}
}

// ResolveTipoEmpresaPreconfiguracion devuelve la configuracion guardada o la sugerida por defecto.
func ResolveTipoEmpresaPreconfiguracion(dbConn *sql.DB, tipoEmpresaID int64, tipoNombre string) (*TipoEmpresaPreconfiguracion, error) {
	if tipoEmpresaID > 0 {
		saved, err := GetTipoEmpresaPreconfiguracionByTipoID(dbConn, tipoEmpresaID)
		if err != nil {
			return nil, err
		}
		if saved != nil && strings.ToLower(strings.TrimSpace(saved.Estado)) != "inactivo" {
			if strings.TrimSpace(saved.TipoEmpresaNombre) == "" {
				saved.TipoEmpresaNombre = tipoNombre
			}
			return saved, nil
		}
	}
	def := DefaultTipoEmpresaPreconfiguracion(tipoEmpresaID, tipoNombre)
	return &def, nil
}

func DefaultTipoEmpresaPreconfigTemplate(tipoNombre string) TipoEmpresaPreconfigTemplate {
	if isTipoEmpresaMotel(tipoNombre) {
		return withPreconfigOperacion(newDefaultTipoEmpresaPreconfigTemplate("MOTEL", "Habitacion", 10, []TipoEmpresaPreconfigProducto{
			productoPreconfig("DEMO-MOTEL-001", "Habitacion sencilla", "Habitaciones", "Servicio base por turno", 18000, 45000, 0),
			productoPreconfig("DEMO-MOTEL-002", "Habitacion doble", "Habitaciones", "Servicio doble por turno", 25000, 65000, 0),
			productoPreconfig("DEMO-MOTEL-003", "Suite jacuzzi", "Habitaciones", "Servicio premium por turno", 42000, 110000, 0),
			productoPreconfig("DEMO-MOTEL-004", "Hora adicional", "Adicionales", "Tiempo adicional de permanencia", 6000, 15000, 0),
			productoPreconfig("DEMO-MOTEL-005", "Minibar gaseosa", "Minibar", "Bebida de minibar", 2500, 6000, 8),
			productoPreconfig("DEMO-MOTEL-006", "Kit aseo personal", "Minibar", "Kit de aseo para huesped", 5000, 12000, 8),
		}, []TipoEmpresaPreconfigUsuario{
			usuarioPreconfig("Recepcion principal", "recepcion", "Gestiona ingresos, salidas y disponibilidad."),
			usuarioPreconfig("Caja turno", "caja", "Registra cobros y cierres de turno."),
			usuarioPreconfig("Limpieza habitaciones", "operacion", "Actualiza estados de limpieza y alistamiento."),
		}, "Asistente operativo para recepcion, turnos, limpieza, tarifas y facturacion."), operacionPreconfig("motel", "Habitacion", "Habitaciones", true, false, false, "", "", 0, []string{"recepcion", "caja", "operacion"}))
	}
	if isTipoEmpresaHotel(tipoNombre) {
		return withPreconfigOperacion(newDefaultTipoEmpresaPreconfigTemplate("HOTEL", "Habitacion", 12, []TipoEmpresaPreconfigProducto{
			productoPreconfig("DEMO-HOTEL-001", "Noche habitacion sencilla", "Alojamiento", "Hospedaje por noche", 45000, 95000, 0),
			productoPreconfig("DEMO-HOTEL-002", "Noche habitacion doble", "Alojamiento", "Hospedaje doble por noche", 65000, 145000, 0),
			productoPreconfig("DEMO-HOTEL-003", "Desayuno huesped", "Restaurante", "Desayuno servido a huesped", 8000, 18000, 10),
			productoPreconfig("DEMO-HOTEL-004", "Lavanderia por kilo", "Servicios", "Servicio de lavanderia", 3500, 9000, 0),
			productoPreconfig("DEMO-HOTEL-005", "Late checkout", "Adicionales", "Salida extendida", 15000, 35000, 0),
		}, []TipoEmpresaPreconfigUsuario{
			usuarioPreconfig("Recepcion hotel", "recepcion", "Atiende reservas, check-in y check-out."),
			usuarioPreconfig("Caja hotel", "caja", "Controla pagos, anticipos y facturacion."),
			usuarioPreconfig("Ama de llaves", "operacion", "Coordina limpieza y disponibilidad."),
		}, "Asistente guia para reservas, ocupacion, consumos, pagos y cierre diario."), operacionPreconfig("hotel", "Habitacion", "Habitaciones", true, false, false, "", "", 0, []string{"recepcion", "caja", "operacion"}))
	}
	if isTipoEmpresaBar(tipoNombre) {
		return withPreconfigOperacion(newDefaultTipoEmpresaPreconfigTemplate("BAR", "Mesa", 10, []TipoEmpresaPreconfigProducto{
			productoPreconfig("DEMO-BAR-001", "Cerveza nacional", "Bebidas", "Botella o lata nacional", 3000, 7000, 24),
			productoPreconfig("DEMO-BAR-002", "Coctel de la casa", "Cocteles", "Preparacion estandar del bar", 9000, 22000, 6),
			productoPreconfig("DEMO-BAR-003", "Gaseosa personal", "Bebidas", "Bebida sin alcohol", 2200, 5000, 18),
			productoPreconfig("DEMO-BAR-004", "Picada para compartir", "Comidas", "Picada de mesa", 18000, 42000, 3),
			productoPreconfig("DEMO-BAR-005", "Cover evento", "Servicios", "Ingreso por evento", 0, 15000, 0),
		}, []TipoEmpresaPreconfigUsuario{
			usuarioPreconfig("Mesero turno", "mesero", "Toma pedidos y atiende mesas."),
			usuarioPreconfig("Barra principal", "barra", "Prepara bebidas y controla inventario."),
			usuarioPreconfig("Caja bar", "caja", "Cobra cuentas y cierra turno."),
		}, "Asistente de pedidos, mesas, inventario de bebidas, promociones y cierre de caja."), operacionPreconfig("bar", "Mesa", "Mesas", true, false, false, "", "", 0, []string{"mesero", "barra", "caja"}))
	}
	if isTipoEmpresaSalonBelleza(tipoNombre) {
		return withPreconfigOperacion(newDefaultTipoEmpresaPreconfigTemplate("BELLEZA", "Silla", 6, []TipoEmpresaPreconfigProducto{
			productoPreconfig("DEMO-BELLEZA-001", "Corte dama", "Peluqueria", "Servicio de corte para dama", 12000, 30000, 0),
			productoPreconfig("DEMO-BELLEZA-002", "Corte caballero", "Peluqueria", "Servicio de corte para caballero", 8000, 22000, 0),
			productoPreconfig("DEMO-BELLEZA-003", "Manicure tradicional", "Unas", "Servicio de manicure", 9000, 25000, 0),
			productoPreconfig("DEMO-BELLEZA-004", "Tinte raiz", "Color", "Aplicacion de tinte en raiz", 35000, 85000, 0),
			productoPreconfig("DEMO-BELLEZA-005", "Tratamiento capilar", "Tratamientos", "Hidratacion o reparacion", 18000, 45000, 0),
		}, []TipoEmpresaPreconfigUsuario{
			usuarioPreconfig("Recepcion salon", "recepcion", "Agenda citas y recibe clientes."),
			usuarioPreconfig("Estilista principal", "estilista", "Atiende servicios de belleza y gana comision por servicio."),
			usuarioPreconfig("Manicurista", "estilista", "Atiende servicios de unas y gana comision por servicio."),
			usuarioPreconfig("Caja salon", "caja", "Registra pagos y paquetes."),
		}, "Asistente para agenda, servicios, recordatorios, inventario de insumos y ventas."), operacionPreconfig("salon_belleza", "Silla", "Sillas", true, true, true, "estilista", "servicio", 35, []string{"recepcion", "estilista", "caja"}))
	}
	if isTipoEmpresaLavaderoAutos(tipoNombre) {
		return withPreconfigOperacion(newDefaultTipoEmpresaPreconfigTemplate("LAVADERO", "Bahia", 6, []TipoEmpresaPreconfigProducto{
			productoPreconfig("DEMO-LAV-001", "Lavado basico carro", "Lavado", "Lavado exterior basico", 8000, 22000, 0),
			productoPreconfig("DEMO-LAV-002", "Lavado premium carro", "Lavado", "Exterior, interior y aspirado", 15000, 38000, 0),
			productoPreconfig("DEMO-LAV-003", "Lavado camioneta", "Lavado", "Servicio para camioneta", 18000, 45000, 0),
			productoPreconfig("DEMO-LAV-004", "Lavado de motor", "Servicios", "Lavado tecnico de motor", 12000, 30000, 0),
			productoPreconfig("DEMO-LAV-005", "Encerado", "Servicios", "Aplicacion de cera", 14000, 35000, 0),
		}, []TipoEmpresaPreconfigUsuario{
			usuarioPreconfig("Recepcion vehiculos", "recepcion", "Recibe vehiculos y asigna bahias."),
			usuarioPreconfig("Operario lavado", "operacion", "Actualiza estados de lavado."),
			usuarioPreconfig("Caja lavadero", "caja", "Cobra servicios y controla turnos."),
		}, "Asistente para turnos, bahias, servicios por vehiculo, tiempos y facturacion."), operacionPreconfig("lavadero_autos", "Bahia", "Bahias", true, true, true, "operacion", "lavado", 20, []string{"recepcion", "operacion", "caja"}))
	}
	if isTipoEmpresaRestaurante(tipoNombre) {
		return withPreconfigOperacion(newDefaultTipoEmpresaPreconfigTemplate("REST", "Mesa", 8, []TipoEmpresaPreconfigProducto{
			productoPreconfig("DEMO-REST-001", "Hamburguesa clasica", "Comidas", "Producto guia de cocina", 9000, 18000, 5),
			productoPreconfig("DEMO-REST-002", "Perro caliente", "Comidas", "Producto guia de cocina", 6000, 12000, 5),
			productoPreconfig("DEMO-REST-003", "Gaseosa personal", "Bebidas", "Bebida personal", 2200, 4000, 12),
			productoPreconfig("DEMO-REST-004", "Agua botella", "Bebidas", "Agua embotellada", 1800, 3500, 12),
			productoPreconfig("DEMO-REST-005", "Menu del dia", "Almuerzos", "Menu diario guia", 12000, 22000, 3),
			productoPreconfig("DEMO-REST-006", "Cafe", "Bebidas calientes", "Cafe preparado", 1200, 3500, 10),
		}, []TipoEmpresaPreconfigUsuario{
			usuarioPreconfig("Mesero principal", "mesero", "Toma pedidos y atiende mesas."),
			usuarioPreconfig("Cocina", "operacion", "Gestiona preparacion y despacho."),
			usuarioPreconfig("Caja restaurante", "caja", "Cobra cuentas y cierres."),
		}, "Asistente para pedidos, mesas, cocina, inventario, descuentos y facturacion."), operacionPreconfig("restaurante", "Mesa", "Mesas", true, true, false, "", "", 0, []string{"mesero", "operacion", "caja"}))
	}
	if isTipoEmpresaPyme(tipoNombre) {
		return withPreconfigOperacion(newDefaultTipoEmpresaPreconfigTemplate("PYME", "Punto de venta", 2, []TipoEmpresaPreconfigProducto{
			productoPreconfig("DEMO-PYME-001", "Producto comercial guia", "General", "Producto base para venta directa o mostrador", 7000, 16000, 8),
			productoPreconfig("DEMO-PYME-002", "Servicio profesional guia", "Servicios", "Servicio configurable para la pyme", 0, 60000, 0),
			productoPreconfig("DEMO-PYME-003", "Paquete mensual guia", "Paquetes", "Paquete de servicio recurrente", 0, 180000, 0),
			productoPreconfig("DEMO-PYME-004", "Entrega local", "Servicios", "Cargo guia de entrega o domicilio", 0, 8000, 0),
		}, []TipoEmpresaPreconfigUsuario{
			usuarioPreconfig("Administrador pyme", "administrador", "Configura empresa, usuarios, reportes y parametros generales."),
			usuarioPreconfig("Vendedor pyme", "vendedor", "Registra ventas, clientes y cotizaciones."),
			usuarioPreconfig("Caja pyme", "caja", "Controla pagos, descuentos y cierres diarios."),
		}, "Asistente para configuracion empresarial, venta directa, caja, clientes, documentos y reportes."), operacionPreconfig("pyme", "Punto de venta", "Puntos de venta", true, true, false, "", "", 0, []string{"administrador", "vendedor", "caja"}))
	}
	if isTipoEmpresaPuntoVenta(tipoNombre) {
		return withPreconfigOperacion(newDefaultTipoEmpresaPreconfigTemplate("PV", "Punto de venta", 1, []TipoEmpresaPreconfigProducto{
			productoPreconfig("DEMO-PV-001", "Producto general A", "General", "Producto de inventario inicial", 5000, 10000, 10),
			productoPreconfig("DEMO-PV-002", "Producto general B", "General", "Producto de inventario inicial", 8000, 16000, 10),
			productoPreconfig("DEMO-PV-003", "Servicio domicilio", "Servicios", "Cargo por domicilio", 0, 5000, 0),
			productoPreconfig("DEMO-PV-004", "Bolsa", "Empaque", "Empaque opcional", 100, 300, 50),
		}, []TipoEmpresaPreconfigUsuario{
			usuarioPreconfig("Vendedor mostrador", "vendedor", "Registra ventas y clientes."),
			usuarioPreconfig("Caja principal", "caja", "Controla pagos y cierre."),
			usuarioPreconfig("Administrador inventario", "administrador", "Ajusta inventario y precios."),
		}, "Asistente para ventas, inventario, alertas de stock, descuentos y reportes."), operacionPreconfig("punto_venta", "Punto de venta", "Puntos de venta", true, true, false, "", "", 0, []string{"vendedor", "caja", "administrador"}))
	}
	if isTipoEmpresaTaller(tipoNombre) {
		return withPreconfigOperacion(newDefaultTipoEmpresaPreconfigTemplate("TALLER", "Bahia", 5, []TipoEmpresaPreconfigProducto{
			productoPreconfig("DEMO-TALLER-001", "Revision general", "Diagnostico", "Revision inicial del vehiculo", 12000, 30000, 0),
			productoPreconfig("DEMO-TALLER-002", "Cambio de aceite", "Mantenimiento", "Mano de obra cambio de aceite", 10000, 25000, 0),
			productoPreconfig("DEMO-TALLER-003", "Alineacion", "Servicios", "Servicio de alineacion", 22000, 55000, 0),
			productoPreconfig("DEMO-TALLER-004", "Filtro de aceite", "Repuestos", "Repuesto guia", 12000, 26000, 4),
		}, []TipoEmpresaPreconfigUsuario{
			usuarioPreconfig("Recepcion taller", "recepcion", "Recibe vehiculos y ordenes."),
			usuarioPreconfig("Tecnico taller", "tecnico", "Ejecuta servicios, reporta avances y gana comision por servicio."),
			usuarioPreconfig("Caja taller", "caja", "Cobra ordenes y repuestos."),
		}, "Asistente para ordenes de servicio, repuestos, tiempos, diagnosticos y cobros."), operacionPreconfig("taller", "Bahia", "Bahias", true, true, true, "tecnico", "servicio", 25, []string{"recepcion", "tecnico", "caja"}))
	}
	if isTipoEmpresaGimnasio(tipoNombre) {
		return withPreconfigOperacion(newDefaultTipoEmpresaPreconfigTemplate("GYM", "Zona", 4, []TipoEmpresaPreconfigProducto{
			productoPreconfig("DEMO-GYM-001", "Mensualidad general", "Planes", "Plan mensual de acceso general", 0, 95000, 0),
			productoPreconfig("DEMO-GYM-002", "Clase personalizada", "Entrenamiento", "Sesion individual con entrenador", 0, 45000, 0),
			productoPreconfig("DEMO-GYM-003", "Dia de entrenamiento", "Planes", "Ingreso por dia", 0, 15000, 0),
			productoPreconfig("DEMO-GYM-004", "Bebida hidratante", "Tienda", "Bebida de tienda fitness", 2500, 6000, 10),
			productoPreconfig("DEMO-GYM-005", "Proteina porcion", "Tienda", "Porcion individual de proteina", 4500, 12000, 6),
		}, []TipoEmpresaPreconfigUsuario{
			usuarioPreconfig("Recepcion gimnasio", "recepcion", "Gestiona socios, accesos, pagos y renovaciones."),
			usuarioPreconfig("Entrenador principal", "entrenador", "Programa clases, valoraciones y sesiones personalizadas."),
			usuarioPreconfig("Caja gimnasio", "caja", "Registra pagos, ventas de tienda y cierres."),
		}, "Asistente para planes, socios, clases, accesos, renovaciones, pagos y alertas de vencimiento."), operacionPreconfig("gimnasio", "Zona", "Zonas", true, true, true, "entrenador", "entrenamiento", 25, []string{"recepcion", "entrenador", "caja"}))
	}
	if isTipoEmpresaOdontologia(tipoNombre) {
		return withPreconfigOperacion(newDefaultTipoEmpresaPreconfigTemplate("ODONTO", "Consultorio", 3, []TipoEmpresaPreconfigProducto{
			productoPreconfig("DEMO-ODONTO-001", "Consulta odontologica", "Consulta", "Valoracion inicial del paciente", 0, 60000, 0),
			productoPreconfig("DEMO-ODONTO-002", "Limpieza dental", "Higiene oral", "Profilaxis y control", 12000, 90000, 0),
			productoPreconfig("DEMO-ODONTO-003", "Resina simple", "Operatoria", "Restauracion de una superficie", 18000, 120000, 0),
			productoPreconfig("DEMO-ODONTO-004", "Radiografia periapical", "Imagenologia", "Radiografia de apoyo diagnostico", 8000, 35000, 0),
			productoPreconfig("DEMO-ODONTO-005", "Kit higiene oral", "Productos", "Cepillo, seda y crema dental", 12000, 28000, 8),
		}, []TipoEmpresaPreconfigUsuario{
			usuarioPreconfig("Recepcion odontologia", "recepcion", "Agenda pacientes, confirma citas y recauda anticipos."),
			usuarioPreconfig("Odontologo general", "odontologo", "Atiende historias, odontogramas, tratamientos y presupuestos."),
			usuarioPreconfig("Auxiliar consultorio", "operacion", "Prepara consultorio, apoya procedimientos e inventario."),
			usuarioPreconfig("Caja odontologia", "caja", "Registra pagos, abonos y cartera de pacientes."),
		}, "Asistente clinico-administrativo para agenda, pacientes, historias, tratamientos, presupuestos y cartera."), operacionPreconfig("odontologia", "Consultorio", "Consultorios", true, true, true, "odontologo", "tratamiento", 30, []string{"recepcion", "odontologo", "operacion", "caja"}))
	}
	if isTipoEmpresaTurnos(tipoNombre) {
		return withPreconfigOperacion(newDefaultTipoEmpresaPreconfigTemplate("TURNOS", "Puesto", 4, []TipoEmpresaPreconfigProducto{
			productoPreconfig("DEMO-TURNOS-001", "Atencion general", "Servicios", "Servicio guia de atencion al cliente", 0, 0, 0),
			productoPreconfig("DEMO-TURNOS-002", "Tramite prioritario", "Servicios", "Servicio guia para filas preferenciales", 0, 0, 0),
			productoPreconfig("DEMO-TURNOS-003", "Certificado o documento", "Servicios", "Documento de tramite con cobro opcional", 0, 12000, 0),
		}, []TipoEmpresaPreconfigUsuario{
			usuarioPreconfig("Coordinador turnos", "supervisor", "Configura servicios, puestos y pantalla publica."),
			usuarioPreconfig("Asesor modulo", "operacion", "Llama turnos y registra atenciones."),
			usuarioPreconfig("Caja tramites", "caja", "Cobra tramites o documentos cuando aplique."),
		}, "Asistente para configurar filas, servicios, puestos, tiempos de espera y pantalla publica."), operacionPreconfig("turnos_atencion", "Puesto", "Puestos", true, true, false, "", "", 0, []string{"supervisor", "operacion", "caja"}))
	}
	if isTipoEmpresaVehiculos(tipoNombre) {
		return withPreconfigOperacion(newDefaultTipoEmpresaPreconfigTemplate("VEH", "Bahia", 4, []TipoEmpresaPreconfigProducto{
			productoPreconfig("DEMO-VEH-001", "Revision de ingreso", "Control vehicular", "Inspeccion de ingreso o recepcion", 0, 25000, 0),
			productoPreconfig("DEMO-VEH-002", "Mantenimiento preventivo", "Mantenimiento", "Servicio preventivo de vehiculo", 18000, 85000, 0),
			productoPreconfig("DEMO-VEH-003", "Lavado operativo", "Servicios", "Lavado o alistamiento de vehiculo", 12000, 30000, 0),
			productoPreconfig("DEMO-VEH-004", "Cambio aceite guia", "Mantenimiento", "Cambio de aceite con insumo de ejemplo", 45000, 95000, 4),
		}, []TipoEmpresaPreconfigUsuario{
			usuarioPreconfig("Recepcion vehicular", "recepcion", "Registra ingresos, placas, propietarios y permanencia."),
			usuarioPreconfig("Tecnico flota", "tecnico", "Actualiza hoja de vida, mantenimientos y alertas."),
			usuarioPreconfig("Caja vehiculos", "caja", "Cobra servicios y reportes de permanencia."),
		}, "Asistente para registro vehicular, hoja de vida, mantenimientos, alertas y cobros."), operacionPreconfig("vehiculos", "Bahia", "Bahias", true, true, true, "tecnico", "mantenimiento", 20, []string{"recepcion", "tecnico", "caja"}))
	}
	if isTipoEmpresaIndependiente(tipoNombre) {
		return withPreconfigOperacion(newDefaultTipoEmpresaPreconfigTemplate("IND", "Venta directa", 0, []TipoEmpresaPreconfigProducto{
			productoPreconfig("DEMO-IND-001", "Consulta inicial", "Servicios", "Servicio profesional inicial", 0, 50000, 0),
			productoPreconfig("DEMO-IND-002", "Servicio especializado", "Servicios", "Servicio principal del profesional", 0, 120000, 0),
			productoPreconfig("DEMO-IND-003", "Paquete mensual", "Paquetes", "Plan mensual de acompanamiento", 0, 350000, 0),
		}, []TipoEmpresaPreconfigUsuario{
			usuarioPreconfig("Administrador profesional", "administrador", "Configura agenda, clientes y servicios."),
			usuarioPreconfig("Asistente administrativo", "operacion", "Ayuda con agenda, cobros y seguimiento."),
			usuarioPreconfig("Caja profesional", "caja", "Registra cobros, comprobantes y cartera simple."),
		}, "Asistente para agenda, clientes, cobros, recordatorios y tareas administrativas."), operacionPreconfig("independiente", "Venta directa", "Ventas directas", false, true, false, "", "", 0, []string{"administrador", "operacion", "caja"}))
	}
	if isTipoEmpresaRedesSociales(tipoNombre) {
		return withPreconfigOperacion(newDefaultTipoEmpresaPreconfigTemplate("SOCIAL", "Cliente", 4, []TipoEmpresaPreconfigProducto{
			productoPreconfig("DEMO-SOCIAL-001", "Plan publicaciones basico", "Marketing", "Gestion de publicaciones basicas", 0, 180000, 0),
			productoPreconfig("DEMO-SOCIAL-002", "Campana pauta", "Publicidad", "Gestion inicial de pauta", 0, 300000, 0),
			productoPreconfig("DEMO-SOCIAL-003", "Diseno pieza grafica", "Diseno", "Pieza individual para redes", 0, 45000, 0),
			productoPreconfig("DEMO-SOCIAL-004", "Reporte mensual", "Reportes", "Informe mensual de gestion", 0, 90000, 0),
		}, []TipoEmpresaPreconfigUsuario{
			usuarioPreconfig("Community manager", "operacion", "Gestiona canales, tareas y publicaciones."),
			usuarioPreconfig("Asesor comercial", "vendedor", "Cotiza planes y atiende clientes."),
			usuarioPreconfig("Caja servicios", "caja", "Registra cobros de servicios y planes."),
		}, "Asistente para tareas de clientes, contenidos, cotizaciones, reportes y seguimiento comercial."), operacionPreconfig("servicios_digitales", "Cliente", "Clientes", true, true, false, "", "", 0, []string{"operacion", "vendedor", "caja"}))
	}
	if isTipoEmpresaSensores(tipoNombre) {
		return withPreconfigOperacion(newDefaultTipoEmpresaPreconfigTemplate("SENSOR", "Acceso", 4, []TipoEmpresaPreconfigProducto{
			productoPreconfig("DEMO-SENSOR-001", "Instalacion sensor", "Instalacion", "Servicio de instalacion inicial", 25000, 80000, 0),
			productoPreconfig("DEMO-SENSOR-002", "Mantenimiento sensor", "Mantenimiento", "Revision tecnica programada", 15000, 45000, 0),
			productoPreconfig("DEMO-SENSOR-003", "Sensor magnetico", "Dispositivos", "Dispositivo guia de inventario", 18000, 42000, 5),
			productoPreconfig("DEMO-SENSOR-004", "Monitoreo mensual", "Servicios", "Servicio mensual de monitoreo", 0, 65000, 0),
		}, []TipoEmpresaPreconfigUsuario{
			usuarioPreconfig("Tecnico instalador", "tecnico", "Instala y revisa sensores."),
			usuarioPreconfig("Monitoreo", "operacion", "Revisa eventos y alertas."),
			usuarioPreconfig("Caja sensores", "caja", "Registra pagos y contratos."),
		}, "Asistente para instalaciones, alertas, mantenimientos, contratos y seguimiento tecnico."), operacionPreconfig("sensores", "Acceso", "Accesos", true, true, false, "", "", 0, []string{"tecnico", "operacion", "caja"}))
	}
	return withPreconfigOperacion(newDefaultTipoEmpresaPreconfigTemplate("GEN", "Estacion", 4, []TipoEmpresaPreconfigProducto{
		productoPreconfig("DEMO-GEN-001", "Producto guia", "General", "Producto inicial de ejemplo", 5000, 12000, 5),
		productoPreconfig("DEMO-GEN-002", "Servicio guia", "Servicios", "Servicio inicial de ejemplo", 0, 25000, 0),
		productoPreconfig("DEMO-GEN-003", "Paquete guia", "Paquetes", "Paquete inicial de ejemplo", 0, 75000, 0),
	}, []TipoEmpresaPreconfigUsuario{
		usuarioPreconfig("Administrador operativo", "administrador", "Configura la empresa y revisa reportes."),
		usuarioPreconfig("Vendedor operativo", "vendedor", "Atiende clientes y registra ventas."),
		usuarioPreconfig("Caja principal", "caja", "Registra ventas y pagos."),
	}, "Asistente guia para configuracion inicial, ventas, auditoria, reportes y tareas diarias."), operacionPreconfig("general", "Estacion", "Estaciones", true, true, false, "", "", 0, []string{"administrador", "vendedor", "caja"}))
}

func newDefaultTipoEmpresaPreconfigTemplate(prefix, stationPrefix string, stationCount int, productos []TipoEmpresaPreconfigProducto, usuarios []TipoEmpresaPreconfigUsuario, iaRol string) TipoEmpresaPreconfigTemplate {
	return NormalizeTipoEmpresaPreconfigTemplate(TipoEmpresaPreconfigTemplate{
		Estaciones: TipoEmpresaPreconfigEstaciones{
			Enabled:     stationCount > 0,
			Cantidad:    stationCount,
			Prefijo:     stationPrefix,
			CardSize:    "medium",
			CajaEnabled: true,
		},
		Operacion: operacionPreconfig(strings.ToLower(prefix), stationPrefix, pluralizeTipoEmpresaStationName(stationPrefix), stationCount > 0, false, false, "", "", 0, nil),
		Productos: productos,
		Usuarios:  usuarios,
		Asistente: TipoEmpresaPreconfigAsistenteIA{
			Enabled: true,
			Rol:     iaRol,
			Instrucciones: []string{
				"Usa la auditoria y la configuracion de la empresa como contexto principal antes de guiar al usuario.",
				"Explica el siguiente paso con instrucciones cortas y accionables segun el modulo donde este trabajando el usuario.",
				"Sugiere revisar productos, estaciones, usuarios, facturacion y reportes antes de operar en produccion.",
				"No bloquees la operacion si el servicio de IA no responde; deja siempre continuar con el flujo normal.",
			},
		},
		TareasGuia: []TipoEmpresaPreconfigTareaGuia{
			{Modulo: "Configuracion", Titulo: "Revisar datos de la empresa", Descripcion: "Completar NIT, direccion, telefonos, regimen, resoluciones y preferencias operativas."},
			{Modulo: "Estaciones", Titulo: "Validar nombres y capacidad", Descripcion: fmt.Sprintf("Ajustar %s, cantidad, tarjeta de caja y vista movil antes de abrir operacion.", stationPrefix)},
			{Modulo: "Productos", Titulo: "Ajustar precios e inventario", Descripcion: "Cambiar costos, precios, stock minimo, categorias e impuestos segun la operacion real."},
			{Modulo: "Usuarios", Titulo: "Convertir usuarios guia en usuarios reales", Descripcion: "Invitar colaboradores con correo real, rol correcto y permisos finos."},
			{Modulo: "IA", Titulo: "Usar el asistente como guia", Descripcion: "Pedirle pasos de configuracion, revision de auditoria, reportes y ayuda operativa diaria."},
		},
	})
}

func withPreconfigOperacion(template TipoEmpresaPreconfigTemplate, operacion TipoEmpresaPreconfigOperacion) TipoEmpresaPreconfigTemplate {
	template.Operacion = operacion
	return enrichTipoEmpresaPreconfigTemplate(NormalizeTipoEmpresaPreconfigTemplate(template))
}

func enrichTipoEmpresaPreconfigTemplate(template TipoEmpresaPreconfigTemplate) TipoEmpresaPreconfigTemplate {
	switch template.Operacion.TipoNegocio {
	case "motel":
		if len(template.Tarifas.Motel) == 0 {
			template.Tarifas.Motel = defaultMotelTarifasPreconfig()
		}
		if template.Modulos.ControlElectrico == nil {
			template.Modulos.ControlElectrico = defaultControlElectricoPreconfig("motel")
		}
	case "hotel":
		if len(template.Tarifas.PorDia) == 0 {
			template.Tarifas.PorDia = defaultHotelTarifasPorDiaPreconfig()
		}
		if template.Modulos.ControlElectrico == nil {
			template.Modulos.ControlElectrico = defaultControlElectricoPreconfig("hotel")
		}
	case "lavadero_autos":
		if len(template.Tarifas.PorMinutos) == 0 {
			template.Tarifas.PorMinutos = defaultLavaderoTarifasPorMinutosPreconfig()
		}
		if template.Modulos.Vehiculos == nil {
			template.Modulos.Vehiculos = defaultVehiculosPreconfig("lavadero")
		}
		if len(template.Modulos.HojaVida) == 0 {
			template.Modulos.HojaVida = defaultHojaVidaVehiculosPreconfig("lavadero")
		}
	case "taller":
		if template.Modulos.Vehiculos == nil {
			template.Modulos.Vehiculos = defaultVehiculosPreconfig("taller")
		}
		if len(template.Modulos.HojaVida) == 0 {
			template.Modulos.HojaVida = defaultHojaVidaVehiculosPreconfig("taller")
		}
	case "gimnasio":
		if template.Modulos.Gimnasio == nil {
			template.Modulos.Gimnasio = defaultGimnasioPreconfig()
		}
	case "odontologia":
		if template.Modulos.Odontologia == nil {
			template.Modulos.Odontologia = defaultOdontologiaPreconfig()
		}
	case "turnos_atencion":
		if template.Modulos.TurnosAtencion == nil {
			template.Modulos.TurnosAtencion = defaultTurnosAtencionPreconfig()
		}
	case "vehiculos":
		if template.Modulos.Vehiculos == nil {
			template.Modulos.Vehiculos = defaultVehiculosPreconfig("flota")
		}
		if len(template.Modulos.HojaVida) == 0 {
			template.Modulos.HojaVida = defaultHojaVidaVehiculosPreconfig("flota")
		}
	}
	return NormalizeTipoEmpresaPreconfigTemplate(template)
}

func defaultMotelTarifasPreconfig() []TipoEmpresaPreconfigTarifaMotel {
	return []TipoEmpresaPreconfigTarifaMotel{
		{EstacionNumero: 1, NombrePlan: "Express sencillo", TipoPlan: "express", CategoriaHabitacion: "Sencilla", DiaSemanaDesde: 1, DiaSemanaHasta: 5, HoraInicio: "00:00", HoraFin: "23:59", MinutosIncluidos: 180, ValorBase: 45000, MinutosExtra: 30, ValorExtra: 8000, CobrarPorFraccion: true, ToleranciaMinutos: 10, Moneda: "COP", Prioridad: 1, AplicarAutomatico: true, Observaciones: "Tarifa guia lunes a viernes."},
		{EstacionNumero: 2, NombrePlan: "Suite jacuzzi", TipoPlan: "suite", CategoriaHabitacion: "Suite", DiaSemanaDesde: 1, DiaSemanaHasta: 7, HoraInicio: "00:00", HoraFin: "23:59", MinutosIncluidos: 240, ValorBase: 110000, MinutosExtra: 30, ValorExtra: 15000, CobrarPorFraccion: true, ToleranciaMinutos: 10, Moneda: "COP", Prioridad: 2, AplicarAutomatico: true, Observaciones: "Tarifa premium de ejemplo."},
		{EstacionNumero: 1, NombrePlan: "Amanecida fin de semana", TipoPlan: "amanecida", CategoriaHabitacion: "Todas", DiaSemanaDesde: 6, DiaSemanaHasta: 7, HoraInicio: "20:00", HoraFin: "08:00", MinutosIncluidos: 720, ValorBase: 150000, MinutosExtra: 60, ValorExtra: 20000, CobrarPorFraccion: true, ToleranciaMinutos: 15, Moneda: "COP", Prioridad: 3, AplicarAutomatico: true, Observaciones: "Plan nocturno guia para fines de semana."},
	}
}

func defaultHotelTarifasPorDiaPreconfig() []TipoEmpresaPreconfigTarifaPorDia {
	return []TipoEmpresaPreconfigTarifaPorDia{
		{EstacionNumero: 1, NombreTarifa: "Sencilla noche", ServicioNombre: "hospedaje", ValorDia: 95000, PersonasDesde: 1, PersonasHasta: 1, HoraCheckIn: "15:00", HoraCheckOut: "12:00", Moneda: "COP", Prioridad: 1, AplicarAutomaticamente: true, Observaciones: "Tarifa guia habitacion sencilla."},
		{EstacionNumero: 2, NombreTarifa: "Doble noche", ServicioNombre: "hospedaje", ValorDia: 145000, PersonasDesde: 1, PersonasHasta: 2, HoraCheckIn: "15:00", HoraCheckOut: "12:00", Moneda: "COP", Prioridad: 2, AplicarAutomaticamente: true, Observaciones: "Tarifa guia habitacion doble."},
		{EstacionNumero: 3, NombreTarifa: "Familiar noche", ServicioNombre: "hospedaje", ValorDia: 220000, PersonasDesde: 3, PersonasHasta: 5, HoraCheckIn: "15:00", HoraCheckOut: "12:00", Moneda: "COP", Prioridad: 3, AplicarAutomaticamente: true, Observaciones: "Tarifa guia familiar."},
	}
}

func defaultControlElectricoPreconfig(contexto string) *TipoEmpresaPreconfigControlElectrico {
	cfg := &TipoEmpresaPreconfigControlElectrico{
		Habilitado:         false,
		RaspberryIP:        "192.168.1.50",
		RaspberryPort:      8081,
		APIPath:            "/api/gpio/relay",
		TimeoutMS:          2500,
		AutoSyncEstaciones: true,
		FailSafeOnError:    false,
		Raspberries: []TipoEmpresaPreconfigControlRaspberry{
			{Codigo: "principal", Nombre: "Control electrico principal", RaspberryIP: "192.168.1.50", RaspberryPort: 8081, APIPath: "/api/gpio/relay", TimeoutMS: 2500, Observaciones: "Raspberry Pi guia para pruebas de control electrico."},
		},
	}
	if contexto == "hotel" {
		cfg.Reles = []TipoEmpresaPreconfigControlElectricoRele{
			{RaspberryCodigo: "principal", EstacionNumero: 1, SalidaCodigo: "luces", TipoCarga: "luces", GPIOPin: 2, RelayName: "Luces habitacion 1", ActiveHigh: true, Modo: "seguimiento_estacion", Observaciones: "Aparato guia para encendido de luces por ocupacion."},
			{RaspberryCodigo: "principal", EstacionNumero: 1, SalidaCodigo: "aire", TipoCarga: "aire_acondicionado", GPIOPin: 3, RelayName: "Aire habitacion 1", ActiveHigh: true, Modo: "manual", ProgramacionHabilitada: true, HoraEncendido: "15:00", HoraApagado: "12:00", ProgramacionDias: "todos", ProgramacionTimezone: "America/Bogota", Observaciones: "Aparato guia para aire acondicionado con horario hotelero."},
			{RaspberryCodigo: "principal", EstacionNumero: 2, SalidaCodigo: "tomacorrientes", TipoCarga: "energia", GPIOPin: 4, RelayName: "Energia habitacion 2", ActiveHigh: true, Modo: "seguimiento_estacion", Observaciones: "Aparato guia para energia de habitacion."},
		}
		return cfg
	}
	cfg.Reles = []TipoEmpresaPreconfigControlElectricoRele{
		{RaspberryCodigo: "principal", EstacionNumero: 1, SalidaCodigo: "luces", TipoCarga: "luces", GPIOPin: 2, RelayName: "Luces habitacion 1", ActiveHigh: true, Modo: "seguimiento_estacion", Observaciones: "Aparato guia para luces por ingreso y salida."},
		{RaspberryCodigo: "principal", EstacionNumero: 1, SalidaCodigo: "jacuzzi", TipoCarga: "jacuzzi", GPIOPin: 3, RelayName: "Jacuzzi habitacion 1", ActiveHigh: true, Modo: "manual", PulsoMS: 0, Observaciones: "Aparato guia para jacuzzi o hidromasaje."},
		{RaspberryCodigo: "principal", EstacionNumero: 2, SalidaCodigo: "ambiente", TipoCarga: "luces_ambiente", GPIOPin: 4, RelayName: "Luces ambiente habitacion 2", ActiveHigh: true, Modo: "seguimiento_estacion", Observaciones: "Aparato guia para luces decorativas."},
		{RaspberryCodigo: "principal", EstacionNumero: 2, SalidaCodigo: "aire", TipoCarga: "aire_acondicionado", GPIOPin: 5, RelayName: "Aire habitacion 2", ActiveHigh: true, Modo: "seguimiento_estacion", Observaciones: "Aparato guia para aire acondicionado."},
	}
	return cfg
}

func defaultLavaderoTarifasPorMinutosPreconfig() []TipoEmpresaPreconfigTarifaPorMinutos {
	return []TipoEmpresaPreconfigTarifaPorMinutos{
		{EstacionNumero: 1, DiaSemanaDesde: 1, DiaSemanaHasta: 5, MinutosBase: 45, ValorBase: 22000, MinutosExtra: 15, ValorExtra: 5000, CobrarPorFraccion: true, Moneda: "COP", Prioridad: 1, Observaciones: "Lavado basico por bahia entre semana."},
		{EstacionNumero: 2, DiaSemanaDesde: 1, DiaSemanaHasta: 7, MinutosBase: 75, ValorBase: 38000, MinutosExtra: 15, ValorExtra: 7000, CobrarPorFraccion: true, Moneda: "COP", Prioridad: 2, Observaciones: "Lavado premium por bahia."},
	}
}

func defaultTurnosAtencionPreconfig() *TipoEmpresaPreconfigTurnosAtencion {
	return &TipoEmpresaPreconfigTurnosAtencion{
		NombreSistema:             "Turnos de atencion",
		NombrePantalla:            "Pantalla de llamados",
		PrefijoGeneral:            "T",
		TiempoLlamadoSegundos:     20,
		PermitirEmisionPublica:    true,
		MostrarTicketsCompletados: true,
		Servicios: []TipoEmpresaPreconfigTurnoAtencionServicio{
			{Codigo: "GEN", Nombre: "Atencion general", Descripcion: "Fila general de servicio.", Prefijo: "G", Prioridad: 10, Color: "#2563eb"},
			{Codigo: "PRI", Nombre: "Prioritario", Descripcion: "Adultos mayores, discapacidad o casos preferenciales.", Prefijo: "P", Prioridad: 1, Color: "#dc2626"},
			{Codigo: "CAJ", Nombre: "Caja y pagos", Descripcion: "Pagos, recaudos y facturacion.", Prefijo: "C", Prioridad: 20, Color: "#16a34a"},
		},
		Puestos: []TipoEmpresaPreconfigTurnoAtencionPuesto{
			{Codigo: "P1", Nombre: "Puesto 1", Area: "Atencion", Ubicacion: "Modulo principal", ServiciosPermitidos: "GEN,PRI"},
			{Codigo: "P2", Nombre: "Puesto 2", Area: "Atencion", Ubicacion: "Modulo secundario", ServiciosPermitidos: "GEN"},
			{Codigo: "CJ1", Nombre: "Caja 1", Area: "Caja", Ubicacion: "Caja principal", ServiciosPermitidos: "CAJ"},
		},
	}
}

func defaultGimnasioPreconfig() *TipoEmpresaPreconfigGimnasio {
	return &TipoEmpresaPreconfigGimnasio{
		Planes: []TipoEmpresaPreconfigGimnasioPlan{
			{Nombre: "Plan mensual basico", Descripcion: "Acceso general en horario regular.", Precio: 95000, DuracionDias: 30, ClasesIncluidas: 4, AccesoIlimitado: false},
			{Nombre: "Plan ilimitado", Descripcion: "Acceso ilimitado y clases grupales.", Precio: 145000, DuracionDias: 30, ClasesIncluidas: 999, AccesoIlimitado: true},
			{Nombre: "Plan personalizado", Descripcion: "Entrenamiento con sesiones personalizadas.", Precio: 260000, DuracionDias: 30, ClasesIncluidas: 8, AccesoIlimitado: true, SesionesPersonalizadas: 4},
		},
		Entrenadores: []TipoEmpresaPreconfigGimnasioEntrenador{
			{NombreCompleto: "Entrenador guia funcional", Especialidad: "Funcional y fuerza", Email: "entrenador.funcional@preconfig.local", Disponibilidad: "Lunes a sabado 06:00-14:00", Certificaciones: "Entrenamiento funcional"},
			{NombreCompleto: "Entrenadora guia bienestar", Especialidad: "Yoga y movilidad", Email: "entrenadora.bienestar@preconfig.local", Disponibilidad: "Lunes a viernes 16:00-21:00", Certificaciones: "Yoga, movilidad"},
		},
		Clases: []TipoEmpresaPreconfigGimnasioClase{
			{Nombre: "Funcional manana", Categoria: "Funcional", EntrenadorIndex: 1, Sede: "principal", Canal: "presencial", Cupos: 20, DuracionMinutos: 60, Precio: 0, Descripcion: "Clase guia incluida en planes."},
			{Nombre: "Yoga movilidad", Categoria: "Bienestar", EntrenadorIndex: 2, Sede: "principal", Canal: "presencial", Cupos: 15, DuracionMinutos: 50, Precio: 25000, Descripcion: "Clase abierta con cobro individual opcional."},
		},
		Socios: []TipoEmpresaPreconfigGimnasioSocio{
			{Codigo: "GYM-DEMO-001", NombreCompleto: "Socio Demo Activo", Documento: "100000001", Telefono: "3000000001", Email: "socio.activo@preconfig.local", Objetivo: "Bajar grasa y mejorar condicion", PlanIndex: 1, Observaciones: "Socio guia para validar accesos y renovaciones."},
			{Codigo: "GYM-DEMO-002", NombreCompleto: "Socia Demo Personalizado", Documento: "100000002", Telefono: "3000000002", Email: "socia.personal@preconfig.local", Objetivo: "Ganar fuerza", PlanIndex: 3, Observaciones: "Socio guia con plan personalizado."},
		},
	}
}

func defaultOdontologiaPreconfig() *TipoEmpresaPreconfigOdontologia {
	return &TipoEmpresaPreconfigOdontologia{
		Pacientes: []TipoEmpresaPreconfigOdontoPaciente{
			{Codigo: "PAC-DEMO-001", NombreCompleto: "Paciente Demo Control", Documento: "100000101", Telefono: "3000000101", Email: "paciente.control@preconfig.local", Aseguradora: "Particular", Alergias: "Sin alergias reportadas", RiesgoMedico: "Bajo", Observaciones: "Paciente guia para historia clinica y odontograma."},
			{Codigo: "PAC-DEMO-002", NombreCompleto: "Paciente Demo Tratamiento", Documento: "100000102", Telefono: "3000000102", Email: "paciente.tratamiento@preconfig.local", Aseguradora: "EPS guia", RiesgoMedico: "Medio", Saldo: 180000, Observaciones: "Paciente guia con presupuesto y saldo."},
		},
		Profesionales: []TipoEmpresaPreconfigOdontoProfesional{
			{NombreCompleto: "Odontologo General Demo", Especialidad: "Odontologia general", RegistroProfesional: "OD-DEMO-001", Email: "odontologo.general@preconfig.local", ColorAgenda: "#0ea5e9"},
			{NombreCompleto: "Ortodoncista Demo", Especialidad: "Ortodoncia", RegistroProfesional: "OD-DEMO-002", Email: "ortodoncia@preconfig.local", ColorAgenda: "#7c3aed"},
		},
		Consultorios: []TipoEmpresaPreconfigOdontoConsultorio{
			{Nombre: "Consultorio 1", Sede: "principal", Sillon: "Sillon A", Observaciones: "Consultorio guia principal."},
			{Nombre: "Consultorio 2", Sede: "principal", Sillon: "Sillon B", Observaciones: "Consultorio guia de apoyo."},
		},
		Tratamientos: []TipoEmpresaPreconfigOdontoTratamiento{
			{PacienteIndex: 2, ProfesionalIndex: 1, Nombre: "Plan limpieza y resina", Categoria: "Operatoria", Piezas: "16, 26", SesionesTotal: 2, CostoEstimado: 210000, Observaciones: "Tratamiento guia para presupuestos y seguimiento."},
			{PacienteIndex: 1, ProfesionalIndex: 2, Nombre: "Valoracion ortodoncia", Categoria: "Ortodoncia", SesionesTotal: 1, CostoEstimado: 80000, Observaciones: "Tratamiento guia de valoracion."},
		},
	}
}

func defaultVehiculosPreconfig(contexto string) *TipoEmpresaPreconfigVehiculos {
	motivo := "Ingreso operativo"
	if contexto == "taller" {
		motivo = "Orden de servicio"
	}
	if contexto == "lavadero" {
		motivo = "Lavado y alistamiento"
	}
	return &TipoEmpresaPreconfigVehiculos{
		PaisCodigo:            "CO",
		EvitarDuplicadoActivo: true,
		Registros: []TipoEmpresaPreconfigVehiculoRegistro{
			{Patente: "ABC123", TipoVehiculo: "automovil", Marca: "Renault", Modelo: "Logan", Color: "Blanco", PropietarioNombre: "Cliente Vehiculo Demo", PropietarioDocumento: "900000001", ConductorNombre: "Conductor Demo", MotivoIngreso: motivo, Observaciones: "Registro vehicular guia."},
			{Patente: "XYZ987", TipoVehiculo: "camioneta", Marca: "Toyota", Modelo: "Hilux", Color: "Gris", PropietarioNombre: "Empresa Flota Demo", PropietarioDocumento: "900000002", ConductorNombre: "Operador Demo", MotivoIngreso: motivo, Observaciones: "Registro guia para permanencia y reportes."},
		},
	}
}

func defaultHojaVidaVehiculosPreconfig(contexto string) []TipoEmpresaPreconfigHojaVida {
	tipoEvento := "mantenimiento"
	eventoTitulo := "Revision preventiva inicial"
	if contexto == "lavadero" {
		tipoEvento = "servicio"
		eventoTitulo = "Lavado premium inicial"
	}
	return []TipoEmpresaPreconfigHojaVida{
		{
			TipoEntidad: "vehiculo", Codigo: "HV-VEH-001", Nombre: "Renault Logan ABC123", ClienteNombre: "Cliente Vehiculo Demo", Identificacion: "ABC123", Marca: "Renault", Modelo: "Logan", Color: "Blanco", EstadoOperativo: "activo",
			Metadata:      map[string]any{"placa": "ABC123", "kilometraje": 45600, "tipo_combustible": "gasolina"},
			Observaciones: "Hoja de vida vehicular guia para mantenimientos, servicios y alertas.",
			Eventos: []TipoEmpresaPreconfigHojaVidaEvento{
				{TipoEvento: tipoEvento, Titulo: eventoTitulo, Descripcion: "Evento guia creado por la preconfiguracion.", Costo: 85000, Responsable: "Tecnico flota", Recurrente: true, RecurrenciaDias: 90},
			},
			Alertas: []TipoEmpresaPreconfigHojaVidaAlerta{
				{Titulo: "Proximo mantenimiento", Descripcion: "Validar aceite, frenos, llantas y documentos.", Prioridad: "media", Responsable: "Tecnico flota"},
			},
		},
		{
			TipoEntidad: "vehiculo", Codigo: "HV-VEH-002", Nombre: "Toyota Hilux XYZ987", ClienteNombre: "Empresa Flota Demo", Identificacion: "XYZ987", Marca: "Toyota", Modelo: "Hilux", Color: "Gris", EstadoOperativo: "activo",
			Metadata:      map[string]any{"placa": "XYZ987", "kilometraje": 78200, "tipo_combustible": "diesel"},
			Observaciones: "Hoja de vida guia para flota o vehiculo de cliente.",
			Eventos: []TipoEmpresaPreconfigHojaVidaEvento{
				{TipoEvento: "documento", Titulo: "Revision SOAT y tecnomecanica", Descripcion: "Control documental guia.", Responsable: "Recepcion vehicular", Recurrente: true, RecurrenciaDias: 365},
			},
			Alertas: []TipoEmpresaPreconfigHojaVidaAlerta{
				{Titulo: "Renovar documentos", Descripcion: "Revisar vencimientos de SOAT, tecnomecanica y polizas.", Prioridad: "alta", Responsable: "Recepcion vehicular"},
			},
		},
	}
}

func operacionPreconfig(tipoNegocio, singular, plural string, usaEstaciones, ventaDirecta, comisiones bool, comisionRol, comisionFiltro string, comisionPorcentaje float64, roles []string) TipoEmpresaPreconfigOperacion {
	return TipoEmpresaPreconfigOperacion{
		TipoNegocio:            tipoNegocio,
		NombreEstacionSingular: singular,
		NombreEstacionPlural:   plural,
		UsaEstaciones:          usaEstaciones,
		VentaDirectaEnabled:    ventaDirecta,
		VentaDirectaNombre:     "Venta directa",
		ComisionesEnabled:      comisiones,
		ComisionRol:            comisionRol,
		ComisionFiltro:         comisionFiltro,
		ComisionPorcentaje:     comisionPorcentaje,
		RolesOperativos:        roles,
	}
}

// EnsureDefaultRolesForTipoEmpresaPreconfiguraciones crea roles base por tipo de empresa
// y solo inicializa permisos cuando el rol no tiene una matriz personalizada.
func EnsureDefaultRolesForTipoEmpresaPreconfiguraciones(dbConn *sql.DB, usuario string) (rolesCreados, rolesActualizados, permisosConfigurados, permisosPersonalizados int, err error) {
	if dbConn == nil {
		err = errors.New("db connection is nil")
		return
	}
	if err = EnsureRolesDeUsuarioSchema(dbConn); err != nil {
		return
	}
	if err = EnsureRolesPermisosSchema(dbConn); err != nil {
		return
	}
	tipos, err := GetTiposEmpresas(dbConn)
	if err != nil {
		return
	}
	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = "sistema.preconfiguracion.roles"
	}
	for _, tipo := range tipos {
		preconfig, resolveErr := ResolveTipoEmpresaPreconfiguracion(dbConn, tipo.ID, tipo.Nombre)
		if resolveErr != nil {
			err = resolveErr
			return
		}
		if preconfig == nil {
			continue
		}
		template, parseErr := ParseTipoEmpresaPreconfigTemplate(preconfig.ConfigJSON)
		if parseErr != nil {
			err = parseErr
			return
		}
		for _, rolNombre := range rolesFromTipoEmpresaPreconfigTemplate(template) {
			rolID, created, upsertErr := UpsertRolDeUsuarioByTipoNombre(dbConn, tipo.ID, rolNombre, descripcionRolPreconfig(rolNombre, tipo.Nombre), usuario)
			if upsertErr != nil {
				err = upsertErr
				return
			}
			if created {
				rolesCreados++
			} else {
				rolesActualizados++
			}
			existentes, listErr := ListRolPermisosModuloByRolID(dbConn, rolID)
			if listErr != nil {
				err = listErr
				return
			}
			if len(existentes) > 0 {
				permisosPersonalizados++
				continue
			}
			if replaceErr := ReplaceRolPermisosDeUsuario(dbConn, rolID, permisosModuloPreconfigRol(rolID, rolNombre), nil, usuario); replaceErr != nil {
				err = replaceErr
				return
			}
			permisosConfigurados++
		}
	}
	return
}

func rolesFromTipoEmpresaPreconfigTemplate(template TipoEmpresaPreconfigTemplate) []string {
	roles := make([]string, 0, len(template.Usuarios)+len(template.Operacion.RolesOperativos)+3)
	add := func(value string) {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			return
		}
		for _, existing := range roles {
			if existing == value {
				return
			}
		}
		roles = append(roles, value)
	}
	add("administrador")
	for _, usuario := range template.Usuarios {
		add(usuario.Rol)
	}
	for _, rol := range template.Operacion.RolesOperativos {
		add(rol)
	}
	if template.Operacion.VentaDirectaEnabled {
		add("caja")
	}
	return roles
}

func permisosModuloPreconfigRol(rolID int64, rolNombre string) []RolPermisoModulo {
	rol := strings.ToLower(strings.TrimSpace(rolNombre))
	allActions := []string{"R", "C", "U", "D", "A"}
	readCreateUpdate := []string{"R", "C", "U"}
	readCreate := []string{"R", "C"}
	readOnly := []string{"R"}
	permisos := make([]RolPermisoModulo, 0, 32)
	add := func(modulo string, acciones []string) {
		for _, accion := range acciones {
			permisos = append(permisos, RolPermisoModulo{RolID: rolID, Modulo: modulo, Accion: accion, Permitido: true})
		}
	}
	switch rol {
	case "administrador", "admin", "admin_empresa", "supervisor":
		for _, modulo := range []string{"ventas", "inventario", "finanzas", "clientes", "compras", "facturacion", "seguridad"} {
			add(modulo, allActions)
		}
	case "caja", "cajero":
		add("ventas", []string{"R", "C", "U", "A"})
		add("finanzas", readCreate)
		add("clientes", readCreateUpdate)
		add("facturacion", readCreate)
		add("inventario", readOnly)
	case "recepcion":
		add("ventas", readCreateUpdate)
		add("clientes", readCreateUpdate)
		add("finanzas", readCreate)
		add("facturacion", readCreate)
		add("inventario", readOnly)
	case "mesero", "barra", "vendedor", "operacion", "tecnico", "estilista", "entrenador", "odontologo":
		add("ventas", readCreateUpdate)
		add("clientes", readCreate)
		add("inventario", readOnly)
		add("facturacion", readOnly)
	case "compras":
		add("compras", readCreateUpdate)
		add("inventario", readCreateUpdate)
		add("finanzas", readOnly)
	case "auditor":
		for _, modulo := range []string{"ventas", "inventario", "finanzas", "clientes", "compras", "facturacion", "seguridad"} {
			add(modulo, readOnly)
		}
	default:
		add("ventas", readCreateUpdate)
		add("clientes", readCreate)
		add("inventario", readOnly)
	}
	return permisos
}

func descripcionRolPreconfig(rolNombre, tipoEmpresaNombre string) string {
	rol := strings.ToLower(strings.TrimSpace(rolNombre))
	tipo := strings.TrimSpace(tipoEmpresaNombre)
	if tipo == "" {
		tipo = "este tipo de empresa"
	}
	switch rol {
	case "administrador", "admin", "admin_empresa":
		return "Rol administrador para configurar " + tipo + ", usuarios, permisos, reportes e integraciones."
	case "caja", "cajero":
		return "Rol de caja para ventas, cobros, cierres, descuentos y comprobantes."
	case "recepcion":
		return "Rol de recepcion para atender clientes, turnos, reservas, ingresos y salidas."
	case "mesero":
		return "Rol de mesero para tomar pedidos, gestionar mesas y entregar cuentas."
	case "barra":
		return "Rol de barra para preparar productos, controlar bebidas e inventario operativo."
	case "vendedor":
		return "Rol vendedor para atender clientes, cotizar y registrar ventas."
	case "operacion":
		return "Rol operativo para ejecutar servicios, actualizar estados y apoyar la atencion diaria."
	case "tecnico":
		return "Rol tecnico para diagnosticos, servicios, instalaciones y avances operativos."
	case "estilista":
		return "Rol de estilista para servicios de belleza, agenda y comisiones."
	case "entrenador":
		return "Rol de entrenador para socios, clases, asistencias, planes y seguimiento fisico."
	case "odontologo":
		return "Rol de odontologo para pacientes, agenda clinica, historias, tratamientos y presupuestos."
	case "compras":
		return "Rol de compras para proveedores, ordenes e inventario."
	case "auditor":
		return "Rol auditor para consultar trazabilidad, reportes y cumplimiento."
	default:
		return "Rol operativo guia para " + tipo + "."
	}
}

func pluralizeTipoEmpresaStationName(singular string) string {
	singular = strings.TrimSpace(singular)
	if singular == "" {
		return "Estaciones"
	}
	lower := strings.ToLower(singular)
	if strings.HasSuffix(lower, "s") {
		return singular
	}
	if strings.HasSuffix(lower, "z") {
		return singular[:len(singular)-1] + "ces"
	}
	if strings.HasSuffix(lower, "a") || strings.HasSuffix(lower, "e") || strings.HasSuffix(lower, "i") || strings.HasSuffix(lower, "o") || strings.HasSuffix(lower, "u") {
		return singular + "s"
	}
	return singular + "es"
}

func productoPreconfig(sku, nombre, categoria, descripcion string, costo, precio, stockMinimo float64) TipoEmpresaPreconfigProducto {
	return TipoEmpresaPreconfigProducto{
		SKU:                sku,
		Nombre:             nombre,
		Categoria:          categoria,
		Descripcion:        descripcion,
		UnidadMedida:       "unidad",
		Costo:              costo,
		Precio:             precio,
		ImpuestoPorcentaje: 0,
		StockMinimo:        stockMinimo,
		StockInicial:       stockMinimo,
	}
}

func usuarioPreconfig(nombre, rol, observaciones string) TipoEmpresaPreconfigUsuario {
	return TipoEmpresaPreconfigUsuario{Nombre: nombre, Rol: rol, Observaciones: observaciones}
}

func ParseTipoEmpresaPreconfigTemplate(raw string) (TipoEmpresaPreconfigTemplate, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return TipoEmpresaPreconfigTemplate{}, nil
	}
	var template TipoEmpresaPreconfigTemplate
	if err := json.Unmarshal([]byte(raw), &template); err != nil {
		return TipoEmpresaPreconfigTemplate{}, err
	}
	return NormalizeTipoEmpresaPreconfigTemplate(template), nil
}

func MarshalTipoEmpresaPreconfigTemplate(template TipoEmpresaPreconfigTemplate) (string, error) {
	normalized := NormalizeTipoEmpresaPreconfigTemplate(template)
	raw, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func NormalizeTipoEmpresaPreconfigTemplate(template TipoEmpresaPreconfigTemplate) TipoEmpresaPreconfigTemplate {
	template.Operacion.TipoNegocio = strings.ToLower(strings.TrimSpace(template.Operacion.TipoNegocio))
	template.Operacion.NombreEstacionSingular = strings.TrimSpace(template.Operacion.NombreEstacionSingular)
	template.Operacion.NombreEstacionPlural = strings.TrimSpace(template.Operacion.NombreEstacionPlural)
	template.Operacion.VentaDirectaNombre = strings.TrimSpace(template.Operacion.VentaDirectaNombre)
	template.Operacion.ComisionRol = strings.ToLower(strings.TrimSpace(template.Operacion.ComisionRol))
	template.Operacion.ComisionFiltro = strings.ToLower(strings.TrimSpace(template.Operacion.ComisionFiltro))
	if template.Operacion.VentaDirectaNombre == "" {
		template.Operacion.VentaDirectaNombre = "Venta directa"
	}
	if template.Operacion.ComisionPorcentaje < 0 {
		template.Operacion.ComisionPorcentaje = 0
	}
	if template.Operacion.ComisionPorcentaje > 100 {
		template.Operacion.ComisionPorcentaje = 100
	}
	if template.Estaciones.Cantidad < 0 {
		template.Estaciones.Cantidad = 0
	}
	if template.Estaciones.Cantidad > 200 {
		template.Estaciones.Cantidad = 200
	}
	template.Estaciones.Prefijo = strings.TrimSpace(template.Estaciones.Prefijo)
	if template.Estaciones.Prefijo == "" {
		template.Estaciones.Prefijo = template.Operacion.NombreEstacionSingular
	}
	if template.Estaciones.Prefijo == "" {
		template.Estaciones.Prefijo = "Estacion"
	}
	if template.Operacion.NombreEstacionSingular == "" {
		template.Operacion.NombreEstacionSingular = template.Estaciones.Prefijo
	}
	if template.Operacion.NombreEstacionPlural == "" {
		template.Operacion.NombreEstacionPlural = pluralizeTipoEmpresaStationName(template.Operacion.NombreEstacionSingular)
	}
	if !template.Operacion.UsaEstaciones && template.Estaciones.Enabled && template.Estaciones.Cantidad > 0 {
		template.Operacion.UsaEstaciones = true
	}
	if !template.Operacion.UsaEstaciones {
		template.Estaciones.Enabled = false
		template.Estaciones.Cantidad = 0
	}
	template.Estaciones.CardSize = strings.ToLower(strings.TrimSpace(template.Estaciones.CardSize))
	if template.Estaciones.CardSize == "" {
		template.Estaciones.CardSize = "medium"
	}
	if !template.Estaciones.Enabled {
		template.Estaciones.Cantidad = 0
	}
	roles := make([]string, 0, len(template.Operacion.RolesOperativos))
	seenRole := map[string]bool{}
	for _, role := range template.Operacion.RolesOperativos {
		role = strings.ToLower(strings.TrimSpace(role))
		if role == "" || seenRole[role] {
			continue
		}
		seenRole[role] = true
		roles = append(roles, role)
	}
	template.Operacion.RolesOperativos = roles

	productos := make([]TipoEmpresaPreconfigProducto, 0, len(template.Productos))
	seenSKU := map[string]bool{}
	for idx, p := range template.Productos {
		p.Nombre = strings.TrimSpace(p.Nombre)
		if p.Nombre == "" {
			continue
		}
		p.SKU = strings.ToUpper(strings.TrimSpace(p.SKU))
		if p.SKU == "" {
			p.SKU = fmt.Sprintf("DEMO-%03d", idx+1)
		}
		if seenSKU[p.SKU] {
			continue
		}
		seenSKU[p.SKU] = true
		p.Categoria = strings.TrimSpace(p.Categoria)
		p.Descripcion = strings.TrimSpace(p.Descripcion)
		p.UnidadMedida = strings.TrimSpace(p.UnidadMedida)
		if p.UnidadMedida == "" {
			p.UnidadMedida = "unidad"
		}
		if p.Precio < 0 {
			p.Precio = 0
		}
		if p.Costo < 0 {
			p.Costo = 0
		}
		if p.StockMinimo < 0 {
			p.StockMinimo = 0
		}
		if p.StockInicial < 0 {
			p.StockInicial = 0
		}
		productos = append(productos, p)
	}
	template.Productos = productos
	usuarios := make([]TipoEmpresaPreconfigUsuario, 0, len(template.Usuarios))
	seenEmail := map[string]bool{}
	for _, u := range template.Usuarios {
		u.Nombre = strings.TrimSpace(u.Nombre)
		u.Rol = strings.ToLower(strings.TrimSpace(u.Rol))
		u.Email = strings.ToLower(strings.TrimSpace(u.Email))
		u.Observaciones = strings.TrimSpace(u.Observaciones)
		if u.Nombre == "" {
			continue
		}
		if u.Rol == "" {
			u.Rol = "operacion"
		}
		if u.Email != "" {
			if seenEmail[u.Email] {
				continue
			}
			seenEmail[u.Email] = true
		}
		usuarios = append(usuarios, u)
	}
	template.Usuarios = usuarios
	template.Asistente.Rol = strings.TrimSpace(template.Asistente.Rol)
	if template.Asistente.Enabled && template.Asistente.Rol == "" {
		template.Asistente.Rol = "Asistente guia para configuracion inicial, operacion diaria, auditoria y reportes."
	}
	instrucciones := make([]string, 0, len(template.Asistente.Instrucciones))
	seenInstruction := map[string]bool{}
	for _, instruction := range template.Asistente.Instrucciones {
		instruction = strings.TrimSpace(instruction)
		key := strings.ToLower(instruction)
		if instruction == "" || seenInstruction[key] {
			continue
		}
		seenInstruction[key] = true
		instrucciones = append(instrucciones, instruction)
	}
	template.Asistente.Instrucciones = instrucciones
	tareas := make([]TipoEmpresaPreconfigTareaGuia, 0, len(template.TareasGuia))
	seenTask := map[string]bool{}
	for _, task := range template.TareasGuia {
		task.Modulo = strings.TrimSpace(task.Modulo)
		task.Titulo = strings.TrimSpace(task.Titulo)
		task.Descripcion = strings.TrimSpace(task.Descripcion)
		if task.Titulo == "" {
			continue
		}
		if task.Modulo == "" {
			task.Modulo = "General"
		}
		key := strings.ToLower(task.Modulo + "|" + task.Titulo)
		if seenTask[key] {
			continue
		}
		seenTask[key] = true
		tareas = append(tareas, task)
	}
	template.TareasGuia = tareas
	template.Tarifas = normalizeTipoEmpresaPreconfigTarifas(template.Tarifas, template.Estaciones.Cantidad)
	template.Modulos = normalizeTipoEmpresaPreconfigModulos(template.Modulos)
	return template
}

func normalizeTipoEmpresaPreconfigTarifas(tarifas TipoEmpresaPreconfigTarifas, estaciones int) TipoEmpresaPreconfigTarifas {
	clampStation := func(v int) int {
		if v <= 0 {
			return 1
		}
		if estaciones > 0 && v > estaciones {
			return estaciones
		}
		return v
	}
	porMinutos := make([]TipoEmpresaPreconfigTarifaPorMinutos, 0, len(tarifas.PorMinutos))
	for _, item := range tarifas.PorMinutos {
		if item.ValorBase < 0 || item.ValorExtra < 0 {
			continue
		}
		item.EstacionNumero = clampStation(item.EstacionNumero)
		if item.DiaSemanaDesde <= 0 {
			item.DiaSemanaDesde = 1
		}
		if item.DiaSemanaHasta <= 0 {
			item.DiaSemanaHasta = 7
		}
		if item.MinutosBase <= 0 {
			item.MinutosBase = 60
		}
		if item.MinutosExtra <= 0 {
			item.MinutosExtra = 30
		}
		item.Moneda = strings.ToUpper(strings.TrimSpace(item.Moneda))
		if item.Moneda == "" {
			item.Moneda = "COP"
		}
		if item.Prioridad <= 0 {
			item.Prioridad = len(porMinutos) + 1
		}
		item.Observaciones = strings.TrimSpace(item.Observaciones)
		porMinutos = append(porMinutos, item)
	}
	tarifas.PorMinutos = porMinutos

	porDia := make([]TipoEmpresaPreconfigTarifaPorDia, 0, len(tarifas.PorDia))
	for _, item := range tarifas.PorDia {
		if item.ValorDia <= 0 {
			continue
		}
		item.EstacionNumero = clampStation(item.EstacionNumero)
		item.NombreTarifa = strings.TrimSpace(item.NombreTarifa)
		item.ServicioNombre = strings.TrimSpace(item.ServicioNombre)
		if item.ServicioNombre == "" {
			item.ServicioNombre = "hospedaje"
		}
		if item.PersonasDesde <= 0 {
			item.PersonasDesde = 1
		}
		if item.PersonasHasta > 0 && item.PersonasHasta < item.PersonasDesde {
			item.PersonasHasta = item.PersonasDesde
		}
		item.HoraCheckIn = strings.TrimSpace(item.HoraCheckIn)
		if item.HoraCheckIn == "" {
			item.HoraCheckIn = "15:00"
		}
		item.HoraCheckOut = strings.TrimSpace(item.HoraCheckOut)
		if item.HoraCheckOut == "" {
			item.HoraCheckOut = "12:00"
		}
		item.Moneda = strings.ToUpper(strings.TrimSpace(item.Moneda))
		if item.Moneda == "" {
			item.Moneda = "COP"
		}
		if item.Prioridad <= 0 {
			item.Prioridad = len(porDia) + 1
		}
		item.Observaciones = strings.TrimSpace(item.Observaciones)
		porDia = append(porDia, item)
	}
	tarifas.PorDia = porDia

	motel := make([]TipoEmpresaPreconfigTarifaMotel, 0, len(tarifas.Motel))
	for _, item := range tarifas.Motel {
		item.NombrePlan = strings.TrimSpace(item.NombrePlan)
		if item.NombrePlan == "" || item.ValorBase < 0 || item.ValorExtra < 0 {
			continue
		}
		item.EstacionNumero = clampStation(item.EstacionNumero)
		item.TipoPlan = strings.ToLower(strings.TrimSpace(item.TipoPlan))
		if item.TipoPlan == "" {
			item.TipoPlan = "express"
		}
		item.CategoriaHabitacion = strings.TrimSpace(item.CategoriaHabitacion)
		if item.DiaSemanaDesde <= 0 {
			item.DiaSemanaDesde = 1
		}
		if item.DiaSemanaHasta <= 0 {
			item.DiaSemanaHasta = 7
		}
		if item.MinutosIncluidos <= 0 {
			item.MinutosIncluidos = 180
		}
		if item.MinutosExtra <= 0 {
			item.MinutosExtra = 30
		}
		item.HoraInicio = strings.TrimSpace(item.HoraInicio)
		if item.HoraInicio == "" {
			item.HoraInicio = "00:00"
		}
		item.HoraFin = strings.TrimSpace(item.HoraFin)
		if item.HoraFin == "" {
			item.HoraFin = "23:59"
		}
		item.Moneda = strings.ToUpper(strings.TrimSpace(item.Moneda))
		if item.Moneda == "" {
			item.Moneda = "COP"
		}
		if item.Prioridad <= 0 {
			item.Prioridad = len(motel) + 1
		}
		item.Observaciones = strings.TrimSpace(item.Observaciones)
		motel = append(motel, item)
	}
	tarifas.Motel = motel
	return tarifas
}

func normalizeTipoEmpresaPreconfigModulos(modulos TipoEmpresaPreconfigModulos) TipoEmpresaPreconfigModulos {
	if modulos.TurnosAtencion != nil {
		cfg := *modulos.TurnosAtencion
		cfg.NombreSistema = strings.TrimSpace(cfg.NombreSistema)
		if cfg.NombreSistema == "" {
			cfg.NombreSistema = "Turnos de atencion"
		}
		cfg.NombrePantalla = strings.TrimSpace(cfg.NombrePantalla)
		if cfg.NombrePantalla == "" {
			cfg.NombrePantalla = "Pantalla de llamados"
		}
		cfg.PrefijoGeneral = strings.ToUpper(strings.TrimSpace(cfg.PrefijoGeneral))
		if cfg.PrefijoGeneral == "" {
			cfg.PrefijoGeneral = "T"
		}
		if cfg.TiempoLlamadoSegundos <= 0 {
			cfg.TiempoLlamadoSegundos = 20
		}
		servicios := make([]TipoEmpresaPreconfigTurnoAtencionServicio, 0, len(cfg.Servicios))
		for _, svc := range cfg.Servicios {
			svc.Codigo = strings.ToUpper(strings.TrimSpace(svc.Codigo))
			svc.Nombre = strings.TrimSpace(svc.Nombre)
			if svc.Codigo == "" || svc.Nombre == "" {
				continue
			}
			svc.Prefijo = strings.ToUpper(strings.TrimSpace(svc.Prefijo))
			if svc.Prefijo == "" {
				svc.Prefijo = svc.Codigo
			}
			svc.Descripcion = strings.TrimSpace(svc.Descripcion)
			svc.Color = strings.TrimSpace(svc.Color)
			if svc.Color == "" {
				svc.Color = "#2563eb"
			}
			if svc.Prioridad <= 0 {
				svc.Prioridad = len(servicios) + 1
			}
			servicios = append(servicios, svc)
		}
		cfg.Servicios = servicios
		puestos := make([]TipoEmpresaPreconfigTurnoAtencionPuesto, 0, len(cfg.Puestos))
		for _, puesto := range cfg.Puestos {
			puesto.Codigo = strings.ToUpper(strings.TrimSpace(puesto.Codigo))
			puesto.Nombre = strings.TrimSpace(puesto.Nombre)
			if puesto.Codigo == "" || puesto.Nombre == "" {
				continue
			}
			puesto.Area = strings.TrimSpace(puesto.Area)
			puesto.Ubicacion = strings.TrimSpace(puesto.Ubicacion)
			puesto.ServiciosPermitidos = strings.ToUpper(strings.TrimSpace(puesto.ServiciosPermitidos))
			puestos = append(puestos, puesto)
		}
		cfg.Puestos = puestos
		modulos.TurnosAtencion = &cfg
	}
	if modulos.Vehiculos != nil {
		cfg := *modulos.Vehiculos
		cfg.PaisCodigo = strings.ToUpper(strings.TrimSpace(cfg.PaisCodigo))
		if cfg.PaisCodigo == "" {
			cfg.PaisCodigo = "CO"
		}
		registros := make([]TipoEmpresaPreconfigVehiculoRegistro, 0, len(cfg.Registros))
		for _, item := range cfg.Registros {
			item.Patente = strings.ToUpper(strings.TrimSpace(item.Patente))
			if item.Patente == "" {
				continue
			}
			item.TipoVehiculo = strings.TrimSpace(item.TipoVehiculo)
			item.Marca = strings.TrimSpace(item.Marca)
			item.Modelo = strings.TrimSpace(item.Modelo)
			item.Color = strings.TrimSpace(item.Color)
			item.PropietarioNombre = strings.TrimSpace(item.PropietarioNombre)
			item.PropietarioDocumento = strings.TrimSpace(item.PropietarioDocumento)
			item.ConductorNombre = strings.TrimSpace(item.ConductorNombre)
			item.MotivoIngreso = strings.TrimSpace(item.MotivoIngreso)
			item.Observaciones = strings.TrimSpace(item.Observaciones)
			registros = append(registros, item)
		}
		cfg.Registros = registros
		modulos.Vehiculos = &cfg
	}
	if modulos.ControlElectrico != nil {
		cfg := *modulos.ControlElectrico
		cfg.RaspberryIP = strings.TrimSpace(cfg.RaspberryIP)
		if cfg.RaspberryIP == "" {
			cfg.RaspberryIP = "192.168.1.50"
		}
		if cfg.RaspberryPort <= 0 {
			cfg.RaspberryPort = 8081
		}
		cfg.APIPath = strings.TrimSpace(cfg.APIPath)
		if cfg.APIPath == "" {
			cfg.APIPath = "/api/gpio/relay"
		}
		if cfg.TimeoutMS <= 0 {
			cfg.TimeoutMS = 2500
		}
		raspberries := make([]TipoEmpresaPreconfigControlRaspberry, 0, len(cfg.Raspberries))
		for _, item := range cfg.Raspberries {
			item.Codigo = strings.ToLower(strings.TrimSpace(item.Codigo))
			if item.Codigo == "" {
				item.Codigo = "principal"
			}
			item.Nombre = strings.TrimSpace(item.Nombre)
			if item.Nombre == "" {
				item.Nombre = "Control electrico " + item.Codigo
			}
			item.RaspberryIP = strings.TrimSpace(item.RaspberryIP)
			if item.RaspberryIP == "" {
				item.RaspberryIP = cfg.RaspberryIP
			}
			if item.RaspberryPort <= 0 {
				item.RaspberryPort = cfg.RaspberryPort
			}
			item.APIPath = strings.TrimSpace(item.APIPath)
			if item.APIPath == "" {
				item.APIPath = cfg.APIPath
			}
			if item.TimeoutMS <= 0 {
				item.TimeoutMS = cfg.TimeoutMS
			}
			item.Observaciones = strings.TrimSpace(item.Observaciones)
			raspberries = append(raspberries, item)
		}
		if len(raspberries) == 0 {
			raspberries = append(raspberries, TipoEmpresaPreconfigControlRaspberry{Codigo: "principal", Nombre: "Control electrico principal", RaspberryIP: cfg.RaspberryIP, RaspberryPort: cfg.RaspberryPort, APIPath: cfg.APIPath, TimeoutMS: cfg.TimeoutMS})
		}
		cfg.Raspberries = raspberries
		reles := make([]TipoEmpresaPreconfigControlElectricoRele, 0, len(cfg.Reles))
		for _, item := range cfg.Reles {
			item.RaspberryCodigo = strings.ToLower(strings.TrimSpace(item.RaspberryCodigo))
			if item.RaspberryCodigo == "" {
				item.RaspberryCodigo = "principal"
			}
			if item.EstacionNumero <= 0 {
				item.EstacionNumero = 1
			}
			item.SalidaCodigo = strings.ToLower(strings.TrimSpace(item.SalidaCodigo))
			item.TipoCarga = strings.ToLower(strings.TrimSpace(item.TipoCarga))
			item.RelayName = strings.TrimSpace(item.RelayName)
			if item.SalidaCodigo == "" && item.TipoCarga == "" && item.RelayName == "" {
				continue
			}
			if item.GPIOPin < 0 {
				item.GPIOPin = 0
			}
			if item.GPIOPin > 27 {
				item.GPIOPin = 27
			}
			if item.Modo == "" {
				item.Modo = "seguimiento_estacion"
			}
			item.ProgramacionDias = strings.TrimSpace(item.ProgramacionDias)
			if item.ProgramacionDias == "" {
				item.ProgramacionDias = "todos"
			}
			item.ProgramacionTimezone = strings.TrimSpace(item.ProgramacionTimezone)
			if item.ProgramacionTimezone == "" {
				item.ProgramacionTimezone = "America/Bogota"
			}
			item.Observaciones = strings.TrimSpace(item.Observaciones)
			reles = append(reles, item)
		}
		cfg.Reles = reles
		modulos.ControlElectrico = &cfg
	}
	hojas := make([]TipoEmpresaPreconfigHojaVida, 0, len(modulos.HojaVida))
	for _, item := range modulos.HojaVida {
		item.TipoEntidad = strings.TrimSpace(item.TipoEntidad)
		if item.TipoEntidad == "" {
			item.TipoEntidad = "activo"
		}
		item.Codigo = strings.TrimSpace(item.Codigo)
		item.Nombre = strings.TrimSpace(item.Nombre)
		if item.Nombre == "" {
			continue
		}
		item.ClienteNombre = strings.TrimSpace(item.ClienteNombre)
		item.Identificacion = strings.TrimSpace(item.Identificacion)
		item.Marca = strings.TrimSpace(item.Marca)
		item.Modelo = strings.TrimSpace(item.Modelo)
		item.Serie = strings.TrimSpace(item.Serie)
		item.Color = strings.TrimSpace(item.Color)
		item.EstadoOperativo = strings.TrimSpace(item.EstadoOperativo)
		if item.EstadoOperativo == "" {
			item.EstadoOperativo = "activo"
		}
		item.Observaciones = strings.TrimSpace(item.Observaciones)
		hojas = append(hojas, item)
	}
	modulos.HojaVida = hojas
	return modulos
}

func isTipoEmpresaRestaurante(tipoNombre string) bool {
	return tipoEmpresaNameContains(tipoNombre, "restaurante", "restaurant", "comida", "cafeteria", "cafeteria", "panaderia", "panaderia")
}

func isTipoEmpresaMotel(tipoNombre string) bool {
	return tipoEmpresaNameContains(tipoNombre, "motel", "residencia")
}

func isTipoEmpresaHotel(tipoNombre string) bool {
	return tipoEmpresaNameContains(tipoNombre, "hotel", "hostal", "hospedaje")
}

func isTipoEmpresaBar(tipoNombre string) bool {
	return tipoEmpresaNameContains(tipoNombre, "bar", "discoteca", "cantina", "licorera")
}

func isTipoEmpresaSalonBelleza(tipoNombre string) bool {
	return tipoEmpresaNameContains(tipoNombre, "salon de belleza", "salon de belleza", "belleza", "peluqueria", "peluqueria", "spa", "barberia", "barberia")
}

func isTipoEmpresaLavaderoAutos(tipoNombre string) bool {
	return tipoEmpresaNameContains(tipoNombre, "lavadero", "autolavado", "lavado de autos", "car wash")
}

func isTipoEmpresaPyme(tipoNombre string) bool {
	return tipoEmpresaNameContains(tipoNombre, "pyme", "pymes", "microempresa", "empresa general", "negocio general")
}

func isTipoEmpresaPuntoVenta(tipoNombre string) bool {
	return tipoEmpresaNameContains(tipoNombre, "tienda", "punto de venta", "retail", "minimercado", "supermercado", "miscelanea", "miscelanea", "almacen", "almacen")
}

func isTipoEmpresaTaller(tipoNombre string) bool {
	return tipoEmpresaNameContains(tipoNombre, "taller", "mecanica", "mecanica", "serviteca")
}

func isTipoEmpresaGimnasio(tipoNombre string) bool {
	return tipoEmpresaNameContains(tipoNombre, "gimnasio", "gym", "fitness", "centro deportivo", "entrenamiento")
}

func isTipoEmpresaOdontologia(tipoNombre string) bool {
	return tipoEmpresaNameContains(tipoNombre, "odontologia", "odontologico", "odontologica", "dentista", "dental", "consultorio dental")
}

func isTipoEmpresaTurnos(tipoNombre string) bool {
	return tipoEmpresaNameContains(tipoNombre, "turnos", "manejo de turnos", "filas", "atencion al cliente", "pantalla de llamados")
}

func isTipoEmpresaVehiculos(tipoNombre string) bool {
	return tipoEmpresaNameContains(tipoNombre, "vehiculo", "vehiculos", "flota", "flotas", "parqueadero", "estacionamiento", "logistica vehicular")
}

func isTipoEmpresaIndependiente(tipoNombre string) bool {
	return tipoEmpresaNameContains(tipoNombre, "independiente", "profesional", "consultor", "freelance")
}

func isTipoEmpresaRedesSociales(tipoNombre string) bool {
	return tipoEmpresaNameContains(tipoNombre, "redes sociales", "social media", "marketing", "agencia digital")
}

func isTipoEmpresaSensores(tipoNombre string) bool {
	return tipoEmpresaNameContains(tipoNombre, "sensor", "sensores", "acceso", "monitoreo", "alarma")
}

func defaultTipoEmpresaPreconfigNombre(tipoNombre string) string {
	switch {
	case isTipoEmpresaMotel(tipoNombre):
		return "Motel con habitaciones guia"
	case isTipoEmpresaHotel(tipoNombre):
		return "Hotel con habitaciones guia"
	case isTipoEmpresaBar(tipoNombre):
		return "Bar con mesas guia"
	case isTipoEmpresaSalonBelleza(tipoNombre):
		return "Salon de belleza con sillas guia"
	case isTipoEmpresaLavaderoAutos(tipoNombre):
		return "Lavadero de autos con bahias guia"
	case isTipoEmpresaRestaurante(tipoNombre):
		return "Restaurante con mesas guia"
	case isTipoEmpresaPyme(tipoNombre):
		return "Pyme con venta directa guia"
	case isTipoEmpresaPuntoVenta(tipoNombre):
		return "Punto de venta guia"
	case isTipoEmpresaTaller(tipoNombre):
		return "Taller con bahias guia"
	case isTipoEmpresaGimnasio(tipoNombre):
		return "Gimnasio con socios, planes y clases guia"
	case isTipoEmpresaOdontologia(tipoNombre):
		return "Odontologia con pacientes y agenda guia"
	case isTipoEmpresaTurnos(tipoNombre):
		return "Manejo de turnos con servicios guia"
	case isTipoEmpresaVehiculos(tipoNombre):
		return "Vehiculos y flotas con hoja de vida guia"
	case isTipoEmpresaIndependiente(tipoNombre):
		return "Independiente con venta directa guia"
	case isTipoEmpresaRedesSociales(tipoNombre):
		return "Redes sociales con clientes guia"
	case isTipoEmpresaSensores(tipoNombre):
		return "Sensores y accesos guia"
	default:
		return "Preconfiguracion inicial guia"
	}
}

func tipoEmpresaNameContains(tipoNombre string, tokens ...string) bool {
	n := normalizeTipoEmpresaName(tipoNombre)
	if n == "" {
		return false
	}
	for _, token := range tokens {
		token = normalizeTipoEmpresaName(token)
		if token != "" && strings.Contains(n, token) {
			return true
		}
	}
	return false
}

func normalizeTipoEmpresaName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(
		"á", "a", "à", "a", "ä", "a",
		"é", "e", "è", "e", "ë", "e",
		"í", "i", "ì", "i", "ï", "i",
		"ó", "o", "ò", "o", "ö", "o",
		"ú", "u", "ù", "u", "ü", "u",
		"ñ", "n",
	)
	return replacer.Replace(value)
}
