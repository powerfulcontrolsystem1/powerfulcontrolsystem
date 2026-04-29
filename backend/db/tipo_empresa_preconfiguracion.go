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
	Productos  []TipoEmpresaPreconfigProducto  `json:"productos"`
	Usuarios   []TipoEmpresaPreconfigUsuario   `json:"usuarios,omitempty"`
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
		COALESCE(p.fecha_creacion, ''), COALESCE(p.fecha_actualizacion, ''),
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
		COALESCE(p.fecha_creacion, ''), COALESCE(p.fecha_actualizacion, ''),
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
		return newDefaultTipoEmpresaPreconfigTemplate("MOTEL", "Habitacion", 10, []TipoEmpresaPreconfigProducto{
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
		}, "Asistente operativo para recepcion, turnos, limpieza, tarifas y facturacion.")
	}
	if isTipoEmpresaHotel(tipoNombre) {
		return newDefaultTipoEmpresaPreconfigTemplate("HOTEL", "Habitacion", 12, []TipoEmpresaPreconfigProducto{
			productoPreconfig("DEMO-HOTEL-001", "Noche habitacion sencilla", "Alojamiento", "Hospedaje por noche", 45000, 95000, 0),
			productoPreconfig("DEMO-HOTEL-002", "Noche habitacion doble", "Alojamiento", "Hospedaje doble por noche", 65000, 145000, 0),
			productoPreconfig("DEMO-HOTEL-003", "Desayuno huesped", "Restaurante", "Desayuno servido a huesped", 8000, 18000, 10),
			productoPreconfig("DEMO-HOTEL-004", "Lavanderia por kilo", "Servicios", "Servicio de lavanderia", 3500, 9000, 0),
			productoPreconfig("DEMO-HOTEL-005", "Late checkout", "Adicionales", "Salida extendida", 15000, 35000, 0),
		}, []TipoEmpresaPreconfigUsuario{
			usuarioPreconfig("Recepcion hotel", "recepcion", "Atiende reservas, check-in y check-out."),
			usuarioPreconfig("Caja hotel", "caja", "Controla pagos, anticipos y facturacion."),
			usuarioPreconfig("Ama de llaves", "operacion", "Coordina limpieza y disponibilidad."),
		}, "Asistente guia para reservas, ocupacion, consumos, pagos y cierre diario.")
	}
	if isTipoEmpresaBar(tipoNombre) {
		return newDefaultTipoEmpresaPreconfigTemplate("BAR", "Mesa", 10, []TipoEmpresaPreconfigProducto{
			productoPreconfig("DEMO-BAR-001", "Cerveza nacional", "Bebidas", "Botella o lata nacional", 3000, 7000, 24),
			productoPreconfig("DEMO-BAR-002", "Coctel de la casa", "Cocteles", "Preparacion estandar del bar", 9000, 22000, 6),
			productoPreconfig("DEMO-BAR-003", "Gaseosa personal", "Bebidas", "Bebida sin alcohol", 2200, 5000, 18),
			productoPreconfig("DEMO-BAR-004", "Picada para compartir", "Comidas", "Picada de mesa", 18000, 42000, 3),
			productoPreconfig("DEMO-BAR-005", "Cover evento", "Servicios", "Ingreso por evento", 0, 15000, 0),
		}, []TipoEmpresaPreconfigUsuario{
			usuarioPreconfig("Mesero turno", "mesero", "Toma pedidos y atiende mesas."),
			usuarioPreconfig("Barra principal", "barra", "Prepara bebidas y controla inventario."),
			usuarioPreconfig("Caja bar", "caja", "Cobra cuentas y cierra turno."),
		}, "Asistente de pedidos, mesas, inventario de bebidas, promociones y cierre de caja.")
	}
	if isTipoEmpresaSalonBelleza(tipoNombre) {
		return newDefaultTipoEmpresaPreconfigTemplate("BELLEZA", "Silla", 6, []TipoEmpresaPreconfigProducto{
			productoPreconfig("DEMO-BELLEZA-001", "Corte dama", "Peluqueria", "Servicio de corte para dama", 12000, 30000, 0),
			productoPreconfig("DEMO-BELLEZA-002", "Corte caballero", "Peluqueria", "Servicio de corte para caballero", 8000, 22000, 0),
			productoPreconfig("DEMO-BELLEZA-003", "Manicure tradicional", "Unas", "Servicio de manicure", 9000, 25000, 0),
			productoPreconfig("DEMO-BELLEZA-004", "Tinte raiz", "Color", "Aplicacion de tinte en raiz", 35000, 85000, 0),
			productoPreconfig("DEMO-BELLEZA-005", "Tratamiento capilar", "Tratamientos", "Hidratacion o reparacion", 18000, 45000, 0),
		}, []TipoEmpresaPreconfigUsuario{
			usuarioPreconfig("Recepcion salon", "recepcion", "Agenda citas y recibe clientes."),
			usuarioPreconfig("Estilista principal", "operacion", "Atiende servicios de belleza."),
			usuarioPreconfig("Caja salon", "caja", "Registra pagos y paquetes."),
		}, "Asistente para agenda, servicios, recordatorios, inventario de insumos y ventas.")
	}
	if isTipoEmpresaLavaderoAutos(tipoNombre) {
		return newDefaultTipoEmpresaPreconfigTemplate("LAVADERO", "Bahia", 6, []TipoEmpresaPreconfigProducto{
			productoPreconfig("DEMO-LAV-001", "Lavado basico carro", "Lavado", "Lavado exterior basico", 8000, 22000, 0),
			productoPreconfig("DEMO-LAV-002", "Lavado premium carro", "Lavado", "Exterior, interior y aspirado", 15000, 38000, 0),
			productoPreconfig("DEMO-LAV-003", "Lavado camioneta", "Lavado", "Servicio para camioneta", 18000, 45000, 0),
			productoPreconfig("DEMO-LAV-004", "Lavado de motor", "Servicios", "Lavado tecnico de motor", 12000, 30000, 0),
			productoPreconfig("DEMO-LAV-005", "Encerado", "Servicios", "Aplicacion de cera", 14000, 35000, 0),
		}, []TipoEmpresaPreconfigUsuario{
			usuarioPreconfig("Recepcion vehiculos", "recepcion", "Recibe vehiculos y asigna bahias."),
			usuarioPreconfig("Operario lavado", "operacion", "Actualiza estados de lavado."),
			usuarioPreconfig("Caja lavadero", "caja", "Cobra servicios y controla turnos."),
		}, "Asistente para turnos, bahias, servicios por vehiculo, tiempos y facturacion.")
	}
	if isTipoEmpresaRestaurante(tipoNombre) {
		return newDefaultTipoEmpresaPreconfigTemplate("REST", "Mesa", 8, []TipoEmpresaPreconfigProducto{
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
		}, "Asistente para pedidos, mesas, cocina, inventario, descuentos y facturacion.")
	}
	if isTipoEmpresaPuntoVenta(tipoNombre) {
		return newDefaultTipoEmpresaPreconfigTemplate("PV", "Caja", 3, []TipoEmpresaPreconfigProducto{
			productoPreconfig("DEMO-PV-001", "Producto general A", "General", "Producto de inventario inicial", 5000, 10000, 10),
			productoPreconfig("DEMO-PV-002", "Producto general B", "General", "Producto de inventario inicial", 8000, 16000, 10),
			productoPreconfig("DEMO-PV-003", "Servicio domicilio", "Servicios", "Cargo por domicilio", 0, 5000, 0),
			productoPreconfig("DEMO-PV-004", "Bolsa", "Empaque", "Empaque opcional", 100, 300, 50),
		}, []TipoEmpresaPreconfigUsuario{
			usuarioPreconfig("Vendedor mostrador", "vendedor", "Registra ventas y clientes."),
			usuarioPreconfig("Caja principal", "caja", "Controla pagos y cierre."),
			usuarioPreconfig("Administrador inventario", "administrador", "Ajusta inventario y precios."),
		}, "Asistente para ventas, inventario, alertas de stock, descuentos y reportes.")
	}
	if isTipoEmpresaTaller(tipoNombre) {
		return newDefaultTipoEmpresaPreconfigTemplate("TALLER", "Bahia", 5, []TipoEmpresaPreconfigProducto{
			productoPreconfig("DEMO-TALLER-001", "Revision general", "Diagnostico", "Revision inicial del vehiculo", 12000, 30000, 0),
			productoPreconfig("DEMO-TALLER-002", "Cambio de aceite", "Mantenimiento", "Mano de obra cambio de aceite", 10000, 25000, 0),
			productoPreconfig("DEMO-TALLER-003", "Alineacion", "Servicios", "Servicio de alineacion", 22000, 55000, 0),
			productoPreconfig("DEMO-TALLER-004", "Filtro de aceite", "Repuestos", "Repuesto guia", 12000, 26000, 4),
		}, []TipoEmpresaPreconfigUsuario{
			usuarioPreconfig("Recepcion taller", "recepcion", "Recibe vehiculos y ordenes."),
			usuarioPreconfig("Tecnico taller", "operacion", "Ejecuta servicios y reporta avances."),
			usuarioPreconfig("Caja taller", "caja", "Cobra ordenes y repuestos."),
		}, "Asistente para ordenes de servicio, repuestos, tiempos, diagnosticos y cobros.")
	}
	if isTipoEmpresaIndependiente(tipoNombre) {
		return newDefaultTipoEmpresaPreconfigTemplate("IND", "Agenda", 3, []TipoEmpresaPreconfigProducto{
			productoPreconfig("DEMO-IND-001", "Consulta inicial", "Servicios", "Servicio profesional inicial", 0, 50000, 0),
			productoPreconfig("DEMO-IND-002", "Servicio especializado", "Servicios", "Servicio principal del profesional", 0, 120000, 0),
			productoPreconfig("DEMO-IND-003", "Paquete mensual", "Paquetes", "Plan mensual de acompanamiento", 0, 350000, 0),
		}, []TipoEmpresaPreconfigUsuario{
			usuarioPreconfig("Administrador profesional", "administrador", "Configura agenda, clientes y servicios."),
			usuarioPreconfig("Asistente administrativo", "operacion", "Ayuda con agenda, cobros y seguimiento."),
		}, "Asistente para agenda, clientes, cobros, recordatorios y tareas administrativas.")
	}
	if isTipoEmpresaRedesSociales(tipoNombre) {
		return newDefaultTipoEmpresaPreconfigTemplate("SOCIAL", "Canal", 4, []TipoEmpresaPreconfigProducto{
			productoPreconfig("DEMO-SOCIAL-001", "Plan publicaciones basico", "Marketing", "Gestion de publicaciones basicas", 0, 180000, 0),
			productoPreconfig("DEMO-SOCIAL-002", "Campana pauta", "Publicidad", "Gestion inicial de pauta", 0, 300000, 0),
			productoPreconfig("DEMO-SOCIAL-003", "Diseno pieza grafica", "Diseno", "Pieza individual para redes", 0, 45000, 0),
			productoPreconfig("DEMO-SOCIAL-004", "Reporte mensual", "Reportes", "Informe mensual de gestion", 0, 90000, 0),
		}, []TipoEmpresaPreconfigUsuario{
			usuarioPreconfig("Community manager", "operacion", "Gestiona canales, tareas y publicaciones."),
			usuarioPreconfig("Asesor comercial", "vendedor", "Cotiza planes y atiende clientes."),
			usuarioPreconfig("Caja servicios", "caja", "Registra cobros de servicios y planes."),
		}, "Asistente para tareas de clientes, contenidos, cotizaciones, reportes y seguimiento comercial.")
	}
	if isTipoEmpresaSensores(tipoNombre) {
		return newDefaultTipoEmpresaPreconfigTemplate("SENSOR", "Acceso", 4, []TipoEmpresaPreconfigProducto{
			productoPreconfig("DEMO-SENSOR-001", "Instalacion sensor", "Instalacion", "Servicio de instalacion inicial", 25000, 80000, 0),
			productoPreconfig("DEMO-SENSOR-002", "Mantenimiento sensor", "Mantenimiento", "Revision tecnica programada", 15000, 45000, 0),
			productoPreconfig("DEMO-SENSOR-003", "Sensor magnetico", "Dispositivos", "Dispositivo guia de inventario", 18000, 42000, 5),
			productoPreconfig("DEMO-SENSOR-004", "Monitoreo mensual", "Servicios", "Servicio mensual de monitoreo", 0, 65000, 0),
		}, []TipoEmpresaPreconfigUsuario{
			usuarioPreconfig("Tecnico instalador", "operacion", "Instala y revisa sensores."),
			usuarioPreconfig("Monitoreo", "operacion", "Revisa eventos y alertas."),
			usuarioPreconfig("Caja sensores", "caja", "Registra pagos y contratos."),
		}, "Asistente para instalaciones, alertas, mantenimientos, contratos y seguimiento tecnico.")
	}
	return newDefaultTipoEmpresaPreconfigTemplate("GEN", "Estacion", 4, []TipoEmpresaPreconfigProducto{
		productoPreconfig("DEMO-GEN-001", "Producto guia", "General", "Producto inicial de ejemplo", 5000, 12000, 5),
		productoPreconfig("DEMO-GEN-002", "Servicio guia", "Servicios", "Servicio inicial de ejemplo", 0, 25000, 0),
	}, []TipoEmpresaPreconfigUsuario{
		usuarioPreconfig("Administrador operativo", "administrador", "Configura la empresa y revisa reportes."),
		usuarioPreconfig("Caja principal", "caja", "Registra ventas y pagos."),
	}, "Asistente guia para configuracion inicial, ventas, auditoria, reportes y tareas diarias.")
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
	if template.Estaciones.Cantidad < 0 {
		template.Estaciones.Cantidad = 0
	}
	if template.Estaciones.Cantidad > 200 {
		template.Estaciones.Cantidad = 200
	}
	template.Estaciones.Prefijo = strings.TrimSpace(template.Estaciones.Prefijo)
	if template.Estaciones.Prefijo == "" {
		template.Estaciones.Prefijo = "Estacion"
	}
	template.Estaciones.CardSize = strings.ToLower(strings.TrimSpace(template.Estaciones.CardSize))
	if template.Estaciones.CardSize == "" {
		template.Estaciones.CardSize = "medium"
	}
	if !template.Estaciones.Enabled {
		template.Estaciones.Cantidad = 0
	}

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
	return template
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

func isTipoEmpresaPuntoVenta(tipoNombre string) bool {
	return tipoEmpresaNameContains(tipoNombre, "tienda", "punto de venta", "retail", "minimercado", "supermercado", "miscelanea", "miscelanea", "almacen", "almacen")
}

func isTipoEmpresaTaller(tipoNombre string) bool {
	return tipoEmpresaNameContains(tipoNombre, "taller", "mecanica", "mecanica", "serviteca")
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
	case isTipoEmpresaPuntoVenta(tipoNombre):
		return "Punto de venta guia"
	case isTipoEmpresaTaller(tipoNombre):
		return "Taller con bahias guia"
	case isTipoEmpresaIndependiente(tipoNombre):
		return "Independiente con agenda guia"
	case isTipoEmpresaRedesSociales(tipoNombre):
		return "Redes sociales con canales guia"
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
