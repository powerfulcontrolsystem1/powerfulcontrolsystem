package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

var (
	// ErrStockInsuficiente se usa cuando una salida/traslado excede la existencia disponible.
	ErrStockInsuficiente = errors.New("stock insuficiente")
)

const (
	inventarioPoliticaCostoPromedio = "promedio"
	inventarioPoliticaCostoPEPS     = "peps"
	comboCostoVariacionMaximaPct    = 35.0
)

// Bodega representa una ubicación de inventario dentro de una empresa.
type Bodega struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Codigo             string `json:"codigo,omitempty"`
	Nombre             string `json:"nombre"`
	Ubicacion          string `json:"ubicacion,omitempty"`
	Responsable        string `json:"responsable,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// CategoriaProducto representa una categoría de productos dentro de una empresa.
type CategoriaProducto struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Codigo             string `json:"codigo,omitempty"`
	Nombre             string `json:"nombre"`
	Descripcion        string `json:"descripcion,omitempty"`
	ColorHex           string `json:"color_hex,omitempty"`
	Orden              int64  `json:"orden,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// Producto representa un producto asociado a una empresa.
type Producto struct {
	ID                   int64   `json:"id"`
	EmpresaID            int64   `json:"empresa_id"`
	BodegaPrincipalID    int64   `json:"bodega_principal_id,omitempty"`
	ProveedorPrincipalID int64   `json:"proveedor_principal_id,omitempty"`
	CategoriaID          int64   `json:"categoria_id,omitempty"`
	SKU                  string  `json:"sku,omitempty"`
	CodigoBarras         string  `json:"codigo_barras,omitempty"`
	Nombre               string  `json:"nombre"`
	Descripcion          string  `json:"descripcion,omitempty"`
	Categoria            string  `json:"categoria,omitempty"`
	Marca                string  `json:"marca,omitempty"`
	UnidadMedida         string  `json:"unidad_medida,omitempty"`
	Costo                float64 `json:"costo"`
	Precio               float64 `json:"precio"`
	ImpuestoPorcentaje   float64 `json:"impuesto_porcentaje"`
	StockMinimo          float64 `json:"stock_minimo"`
	StockMaximo          float64 `json:"stock_maximo"`
	StockTotal           float64 `json:"stock_total"`
	ImagenURL            string  `json:"imagen_url,omitempty"`
	FechaCreacion        string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion   string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador       string  `json:"usuario_creador,omitempty"`
	Estado               string  `json:"estado,omitempty"`
	Observaciones        string  `json:"observaciones,omitempty"`
}

// ComboProducto representa un producto compuesto (combo) que se vende a precio unico.
type ComboProducto struct {
	ID                 int64                  `json:"id"`
	EmpresaID          int64                  `json:"empresa_id"`
	Codigo             string                 `json:"codigo,omitempty"`
	Nombre             string                 `json:"nombre"`
	Descripcion        string                 `json:"descripcion,omitempty"`
	UnidadMedida       string                 `json:"unidad_medida,omitempty"`
	Precio             float64                `json:"precio"`
	ImpuestoPorcentaje float64                `json:"impuesto_porcentaje"`
	RecetaVersion      int64                  `json:"receta_version,omitempty"`
	CostoTeorico       float64                `json:"costo_teorico,omitempty"`
	CostoReal          float64                `json:"costo_real,omitempty"`
	VariacionCosto     float64                `json:"variacion_costo,omitempty"`
	VariacionCostoPct  float64                `json:"variacion_costo_porcentaje,omitempty"`
	IngredientesCount  int64                  `json:"ingredientes_count,omitempty"`
	FechaCreacion      string                 `json:"fecha_creacion,omitempty"`
	FechaActualizacion string                 `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string                 `json:"usuario_creador,omitempty"`
	Estado             string                 `json:"estado,omitempty"`
	Observaciones      string                 `json:"observaciones,omitempty"`
	Ingredientes       []ComboProductoDetalle `json:"ingredientes,omitempty"`
}

// ComboProductoDetalle representa la receta/ingredientes de un combo.
type ComboProductoDetalle struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	ComboID            int64   `json:"combo_id"`
	ProductoID         int64   `json:"producto_id"`
	ProductoNombre     string  `json:"producto_nombre,omitempty"`
	ProductoSKU        string  `json:"producto_sku,omitempty"`
	ProductoCodigo     string  `json:"producto_codigo_barras,omitempty"`
	Cantidad           float64 `json:"cantidad"`
	UnidadMedida       string  `json:"unidad_medida,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// Proveedor representa un proveedor comercial de la empresa.
type Proveedor struct {
	ID                    int64   `json:"id"`
	EmpresaID             int64   `json:"empresa_id"`
	Codigo                string  `json:"codigo,omitempty"`
	Nombre                string  `json:"nombre"`
	Documento             string  `json:"documento,omitempty"`
	Contacto              string  `json:"contacto,omitempty"`
	Telefono              string  `json:"telefono,omitempty"`
	Email                 string  `json:"email,omitempty"`
	Direccion             string  `json:"direccion,omitempty"`
	CatalogoReferencia    string  `json:"catalogo_referencia,omitempty"`
	PrecioBaseReferencial float64 `json:"precio_base_referencial"`
	DescuentoPorcentaje   float64 `json:"descuento_porcentaje"`
	PlazoPagoDias         int64   `json:"plazo_pago_dias,omitempty"`
	CondicionEntrega      string  `json:"condicion_entrega,omitempty"`
	FechaCreacion         string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion    string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador        string  `json:"usuario_creador,omitempty"`
	Estado                string  `json:"estado,omitempty"`
	Observaciones         string  `json:"observaciones,omitempty"`
}

// Servicio representa un servicio vendible como si fuese un producto.
type Servicio struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	Codigo             string  `json:"codigo,omitempty"`
	Nombre             string  `json:"nombre"`
	Descripcion        string  `json:"descripcion,omitempty"`
	Categoria          string  `json:"categoria,omitempty"`
	DuracionMinutos    int     `json:"duracion_minutos,omitempty"`
	CostoReferencial   float64 `json:"costo_referencial"`
	Precio             float64 `json:"precio"`
	ImpuestoPorcentaje float64 `json:"impuesto_porcentaje"`
	ImagenURL          string  `json:"imagen_url,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// ProductoPrecioHistorial representa cambios de precio/costo/impuesto de productos.
type ProductoPrecioHistorial struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	ProductoID         int64   `json:"producto_id"`
	ProductoNombre     string  `json:"producto_nombre,omitempty"`
	CostoAnterior      float64 `json:"costo_anterior"`
	CostoNuevo         float64 `json:"costo_nuevo"`
	PrecioAnterior     float64 `json:"precio_anterior"`
	PrecioNuevo        float64 `json:"precio_nuevo"`
	ImpuestoAnterior   float64 `json:"impuesto_anterior"`
	ImpuestoNuevo      float64 `json:"impuesto_nuevo"`
	Motivo             string  `json:"motivo,omitempty"`
	Referencia         string  `json:"referencia,omitempty"`
	FechaCambio        string  `json:"fecha_cambio,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// InventarioExistencia representa el stock por producto y bodega.
type InventarioExistencia struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	ProductoID         int64   `json:"producto_id"`
	ProductoNombre     string  `json:"producto_nombre,omitempty"`
	BodegaID           int64   `json:"bodega_id"`
	BodegaNombre       string  `json:"bodega_nombre,omitempty"`
	Cantidad           float64 `json:"cantidad"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// InventarioAlertaQuiebre representa una alerta de quiebre/bajo minimo por producto y bodega.
type InventarioAlertaQuiebre struct {
	EmpresaID          int64   `json:"empresa_id"`
	ProductoID         int64   `json:"producto_id"`
	ProductoNombre     string  `json:"producto_nombre"`
	BodegaID           int64   `json:"bodega_id"`
	BodegaNombre       string  `json:"bodega_nombre"`
	Cantidad           float64 `json:"cantidad"`
	StockMinimo        float64 `json:"stock_minimo"`
	StockMaximo        float64 `json:"stock_maximo"`
	EstadoStock        string  `json:"estado_stock"`
	Deficit            float64 `json:"deficit"`
	SugeridoReposicion float64 `json:"sugerido_reposicion"`
}

// InventarioAlertaOperativa representa alertas proactivas de quiebre y sobrestock.
type InventarioAlertaOperativa struct {
	EmpresaID          int64   `json:"empresa_id"`
	ProductoID         int64   `json:"producto_id"`
	ProductoNombre     string  `json:"producto_nombre"`
	BodegaID           int64   `json:"bodega_id"`
	BodegaNombre       string  `json:"bodega_nombre"`
	Cantidad           float64 `json:"cantidad"`
	StockMinimo        float64 `json:"stock_minimo"`
	StockMaximo        float64 `json:"stock_maximo"`
	EstadoStock        string  `json:"estado_stock"`
	Deficit            float64 `json:"deficit"`
	Exceso             float64 `json:"exceso"`
	SugeridoReposicion float64 `json:"sugerido_reposicion"`
	NivelAlerta        string  `json:"nivel_alerta"`
	AccionSugerida     string  `json:"accion_sugerida"`
}

// EmpresaInventarioConfiguracion representa reglas operativas de inventario por empresa.
type EmpresaInventarioConfiguracion struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	PoliticaCosto      string `json:"politica_costo"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// InventarioConteoCiclico representa un conteo ciclico con trazabilidad de ajuste.
type InventarioConteoCiclico struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	ProductoID         int64   `json:"producto_id"`
	ProductoNombre     string  `json:"producto_nombre,omitempty"`
	BodegaID           int64   `json:"bodega_id"`
	BodegaNombre       string  `json:"bodega_nombre,omitempty"`
	CantidadSistema    float64 `json:"cantidad_sistema"`
	CantidadContada    float64 `json:"cantidad_contada"`
	Variacion          float64 `json:"variacion"`
	TipoAjuste         string  `json:"tipo_ajuste"`
	MovimientoID       int64   `json:"movimiento_id,omitempty"`
	Referencia         string  `json:"referencia,omitempty"`
	FechaConteo        string  `json:"fecha_conteo,omitempty"`
	UsuarioRevisor     string  `json:"usuario_revisor,omitempty"`
	EstadoConteo       string  `json:"estado_conteo,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// InventarioResumen representa indicadores operativos de inventario por empresa.
type InventarioResumen struct {
	EmpresaID              int64   `json:"empresa_id"`
	TotalExistencias       float64 `json:"total_existencias"`
	ProductosConExistencia int64   `json:"productos_con_existencia"`
	BodegasConStock        int64   `json:"bodegas_con_stock"`
	AlertasTotal           int64   `json:"alertas_total"`
	AlertasSinStock        int64   `json:"alertas_sin_stock"`
	AlertasBajoMinimo      int64   `json:"alertas_bajo_minimo"`
	DeficitTotal           float64 `json:"deficit_total"`
	MovimientosTotal       int64   `json:"movimientos_total"`
	MovimientosEntrada     int64   `json:"movimientos_entrada"`
	MovimientosSalida      int64   `json:"movimientos_salida"`
	MovimientosTraslado    int64   `json:"movimientos_traslado"`
	MovimientosAjuste      int64   `json:"movimientos_ajuste"`
	UltimoMovimiento       string  `json:"ultimo_movimiento,omitempty"`
	PeriodoDesde           string  `json:"periodo_desde,omitempty"`
	PeriodoHasta           string  `json:"periodo_hasta,omitempty"`
}

// InventarioTendenciaDia representa el comportamiento diario del inventario.
type InventarioTendenciaDia struct {
	Fecha      string  `json:"fecha"`
	Entradas   float64 `json:"entradas"`
	Salidas    float64 `json:"salidas"`
	AjusteNeto float64 `json:"ajuste_neto"`
	Traslados  float64 `json:"traslados"`
	Eventos    int64   `json:"eventos"`
	Neto       float64 `json:"neto"`
}

// InventarioBalanceBodega representa el balance operativo por bodega en un rango.
type InventarioBalanceBodega struct {
	BodegaID         int64   `json:"bodega_id"`
	BodegaNombre     string  `json:"bodega_nombre"`
	Entradas         float64 `json:"entradas"`
	Salidas          float64 `json:"salidas"`
	TrasladosEntrada float64 `json:"traslados_entrada"`
	TrasladosSalida  float64 `json:"traslados_salida"`
	TrasladoNeto     float64 `json:"traslado_neto"`
	Eventos          int64   `json:"eventos"`
	Neto             float64 `json:"neto"`
}

// InventarioProyeccionQuiebre representa una alerta preventiva de agotamiento.
type InventarioProyeccionQuiebre struct {
	EmpresaID            int64   `json:"empresa_id"`
	ProductoID           int64   `json:"producto_id"`
	ProductoNombre       string  `json:"producto_nombre"`
	BodegaID             int64   `json:"bodega_id"`
	BodegaNombre         string  `json:"bodega_nombre"`
	StockActual          float64 `json:"stock_actual"`
	StockMinimo          float64 `json:"stock_minimo"`
	StockMaximo          float64 `json:"stock_maximo"`
	Deficit              float64 `json:"deficit"`
	SalidaPromedioDiaria float64 `json:"salida_promedio_diaria"`
	DiasCobertura        float64 `json:"dias_cobertura"`
	EstadoProyeccion     string  `json:"estado_proyeccion"`
	SugeridoReposicion   float64 `json:"sugerido_reposicion"`
	DiasVentana          int     `json:"dias_ventana"`
}

// InventarioPlanReposicionItem representa una sugerencia de compra preventiva por proveedor.
type InventarioPlanReposicionItem struct {
	EmpresaID          int64   `json:"empresa_id"`
	ProveedorID        int64   `json:"proveedor_id"`
	ProveedorNombre    string  `json:"proveedor_nombre"`
	ProductoID         int64   `json:"producto_id"`
	ProductoNombre     string  `json:"producto_nombre"`
	BodegaID           int64   `json:"bodega_id"`
	BodegaNombre       string  `json:"bodega_nombre"`
	EstadoProyeccion   string  `json:"estado_proyeccion"`
	DiasCobertura      float64 `json:"dias_cobertura"`
	SugeridoReposicion float64 `json:"sugerido_reposicion"`
	CostoUnitarioRef   float64 `json:"costo_unitario_ref"`
	CostoEstimado      float64 `json:"costo_estimado"`
	DiasVentana        int     `json:"dias_ventana"`
}

// InventarioPlanReposicionProveedorResumen representa el consolidado de compra por proveedor.
type InventarioPlanReposicionProveedorResumen struct {
	EmpresaID        int64   `json:"empresa_id"`
	ProveedorID      int64   `json:"proveedor_id"`
	ProveedorNombre  string  `json:"proveedor_nombre"`
	Items            int64   `json:"items"`
	ProductosUnicos  int64   `json:"productos_unicos"`
	CantidadTotal    float64 `json:"cantidad_total"`
	CostoTotal       float64 `json:"costo_total"`
	QuiebreInminente int64   `json:"quiebre_inminente"`
	BajoMinimo       int64   `json:"bajo_minimo"`
	RiesgoAlto       int64   `json:"riesgo_alto"`
	RiesgoMedio      int64   `json:"riesgo_medio"`
	DiasVentana      int     `json:"dias_ventana"`
}

// InventarioPlanReposicionBorradorItem representa una linea sugerida en un borrador de orden de compra.
type InventarioPlanReposicionBorradorItem struct {
	EmpresaID        int64   `json:"empresa_id"`
	ProveedorID      int64   `json:"proveedor_id"`
	ProveedorNombre  string  `json:"proveedor_nombre"`
	ProductoID       int64   `json:"producto_id"`
	ProductoNombre   string  `json:"producto_nombre"`
	BodegaID         int64   `json:"bodega_id"`
	BodegaNombre     string  `json:"bodega_nombre"`
	EstadoProyeccion string  `json:"estado_proyeccion"`
	DiasCobertura    float64 `json:"dias_cobertura"`
	CantidadSugerida float64 `json:"cantidad_sugerida"`
	CostoUnitarioRef float64 `json:"costo_unitario_ref"`
	CostoEstimado    float64 `json:"costo_estimado"`
}

// InventarioPlanReposicionBorradorCompra representa un borrador consolidado de orden de compra por proveedor.
type InventarioPlanReposicionBorradorCompra struct {
	EmpresaID        int64                                  `json:"empresa_id"`
	ProveedorID      int64                                  `json:"proveedor_id"`
	ProveedorNombre  string                                 `json:"proveedor_nombre"`
	FechaDocumento   string                                 `json:"fecha_documento"`
	CodigoBorrador   string                                 `json:"codigo_borrador"`
	DiasVentana      int                                    `json:"dias_ventana"`
	Items            []InventarioPlanReposicionBorradorItem `json:"items"`
	TotalItems       int64                                  `json:"total_items"`
	ProductosUnicos  int64                                  `json:"productos_unicos"`
	CantidadTotal    float64                                `json:"cantidad_total"`
	CostoTotal       float64                                `json:"costo_total"`
	QuiebreInminente int64                                  `json:"quiebre_inminente"`
	BajoMinimo       int64                                  `json:"bajo_minimo"`
	RiesgoAlto       int64                                  `json:"riesgo_alto"`
	RiesgoMedio      int64                                  `json:"riesgo_medio"`
}

// InventarioPlanReposicionOrdenEmitida representa el resultado de emitir una OC desde un borrador de reposicion.
type InventarioPlanReposicionOrdenEmitida struct {
	EmpresaID       int64   `json:"empresa_id"`
	ProveedorID     int64   `json:"proveedor_id"`
	ProveedorNombre string  `json:"proveedor_nombre"`
	DocumentoCodigo string  `json:"documento_codigo"`
	Accion          string  `json:"accion"`
	EstadoAnterior  string  `json:"estado_anterior"`
	EstadoNuevo     string  `json:"estado_nuevo"`
	Evento          string  `json:"evento"`
	PeriodoContable string  `json:"periodo_contable"`
	Moneda          string  `json:"moneda"`
	FechaDocumento  string  `json:"fecha_documento"`
	TotalItems      int64   `json:"total_items"`
	ProductosUnicos int64   `json:"productos_unicos"`
	CantidadTotal   float64 `json:"cantidad_total"`
	CostoTotal      float64 `json:"costo_total"`
	EntidadID       int64   `json:"entidad_id"`
}

// InventarioPlanReposicionOrdenEstadoActualizado representa el resultado de transicionar la OC emitida.
type InventarioPlanReposicionOrdenEstadoActualizado struct {
	EmpresaID       int64   `json:"empresa_id"`
	ProveedorID     int64   `json:"proveedor_id"`
	DocumentoCodigo string  `json:"documento_codigo"`
	Accion          string  `json:"accion"`
	EstadoAnterior  string  `json:"estado_anterior"`
	EstadoNuevo     string  `json:"estado_nuevo"`
	Evento          string  `json:"evento"`
	PeriodoContable string  `json:"periodo_contable"`
	Moneda          string  `json:"moneda"`
	MontoTotal      float64 `json:"monto_total"`
	EntidadID       int64   `json:"entidad_id"`
}

// InventarioMovimiento representa un evento de inventario (entrada/salida/traslado/ajuste).
type InventarioMovimiento struct {
	ID                  int64   `json:"id"`
	EmpresaID           int64   `json:"empresa_id"`
	ProductoID          int64   `json:"producto_id"`
	ProductoNombre      string  `json:"producto_nombre,omitempty"`
	BodegaOrigenID      int64   `json:"bodega_origen_id,omitempty"`
	BodegaOrigenNombre  string  `json:"bodega_origen_nombre,omitempty"`
	BodegaDestinoID     int64   `json:"bodega_destino_id,omitempty"`
	BodegaDestinoNombre string  `json:"bodega_destino_nombre,omitempty"`
	Tipo                string  `json:"tipo"`
	Cantidad            float64 `json:"cantidad"`
	CostoUnitario       float64 `json:"costo_unitario"`
	Referencia          string  `json:"referencia,omitempty"`
	FechaMovimiento     string  `json:"fecha_movimiento,omitempty"`
	FechaCreacion       string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion  string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador      string  `json:"usuario_creador,omitempty"`
	Estado              string  `json:"estado,omitempty"`
	Observaciones       string  `json:"observaciones,omitempty"`
}

// EnsureEmpresaProductosSchema crea y migra las tablas del módulo de productos en empresas.db.
func EnsureEmpresaProductosSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS bodegas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT,
			nombre TEXT NOT NULL,
			ubicacion TEXT,
			responsable TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS categorias_productos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT,
			nombre TEXT NOT NULL,
			descripcion TEXT,
			color_hex TEXT,
			orden INTEGER DEFAULT 0,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS productos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			bodega_principal_id INTEGER,
			proveedor_principal_id INTEGER,
			categoria_id INTEGER,
			sku TEXT,
			codigo_barras TEXT,
			nombre TEXT NOT NULL,
			descripcion TEXT,
			categoria TEXT,
			marca TEXT,
			unidad_medida TEXT DEFAULT 'unidad',
			costo REAL DEFAULT 0,
			precio REAL DEFAULT 0,
			impuesto_porcentaje REAL DEFAULT 0,
			stock_minimo REAL DEFAULT 0,
			stock_maximo REAL DEFAULT 0,
			imagen_url TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS proveedores (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT,
			nombre TEXT NOT NULL,
			documento TEXT,
			contacto TEXT,
			telefono TEXT,
			email TEXT,
			direccion TEXT,
			catalogo_referencia TEXT,
			precio_base_referencial REAL DEFAULT 0,
			descuento_porcentaje REAL DEFAULT 0,
			plazo_pago_dias INTEGER DEFAULT 0,
			condicion_entrega TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS servicios (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT,
			nombre TEXT NOT NULL,
			descripcion TEXT,
			categoria TEXT,
			duracion_minutos INTEGER DEFAULT 0,
			costo_referencial REAL DEFAULT 0,
			precio REAL DEFAULT 0,
			impuesto_porcentaje REAL DEFAULT 0,
			imagen_url TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS combos_productos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT,
			nombre TEXT NOT NULL,
			descripcion TEXT,
			unidad_medida TEXT DEFAULT 'combo',
			precio REAL DEFAULT 0,
			impuesto_porcentaje REAL DEFAULT 0,
			receta_version INTEGER DEFAULT 1,
			costo_teorico REAL DEFAULT 0,
			costo_real REAL DEFAULT 0,
			variacion_costo REAL DEFAULT 0,
			variacion_costo_porcentaje REAL DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS combos_productos_detalle (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			empresa_id INTEGER NOT NULL,
			combo_id INTEGER NOT NULL,
			producto_id INTEGER NOT NULL,
			cantidad REAL NOT NULL DEFAULT 0,
			unidad_medida TEXT DEFAULT 'unidad'
		);`,
		`CREATE TABLE IF NOT EXISTS combos_productos_versiones (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			empresa_id INTEGER NOT NULL,
			combo_id INTEGER NOT NULL,
			receta_version INTEGER NOT NULL,
			ingredientes_json TEXT DEFAULT '[]',
			costo_teorico REAL DEFAULT 0,
			costo_real REAL DEFAULT 0,
			variacion_costo REAL DEFAULT 0,
			variacion_costo_porcentaje REAL DEFAULT 0,
			motivo TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS producto_precios_historial (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			producto_id INTEGER NOT NULL,
			costo_anterior REAL DEFAULT 0,
			costo_nuevo REAL DEFAULT 0,
			precio_anterior REAL DEFAULT 0,
			precio_nuevo REAL DEFAULT 0,
			impuesto_anterior REAL DEFAULT 0,
			impuesto_nuevo REAL DEFAULT 0,
			motivo TEXT,
			referencia TEXT,
			fecha_cambio TEXT DEFAULT (datetime('now','localtime')),
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS inventario_existencias (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			producto_id INTEGER NOT NULL,
			bodega_id INTEGER NOT NULL,
			cantidad REAL NOT NULL DEFAULT 0,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS inventario_movimientos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			producto_id INTEGER NOT NULL,
			bodega_origen_id INTEGER,
			bodega_destino_id INTEGER,
			tipo TEXT NOT NULL,
			cantidad REAL NOT NULL,
			costo_unitario REAL DEFAULT 0,
			referencia TEXT,
			fecha_movimiento TEXT DEFAULT (datetime('now','localtime')),
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS empresa_inventario_configuracion (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL UNIQUE,
			politica_costo TEXT NOT NULL DEFAULT 'promedio',
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS inventario_costos_lotes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			producto_id INTEGER NOT NULL,
			bodega_id INTEGER NOT NULL,
			cantidad_disponible REAL NOT NULL DEFAULT 0,
			costo_unitario REAL NOT NULL DEFAULT 0,
			referencia TEXT,
			fecha_lote TEXT DEFAULT (datetime('now','localtime')),
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS inventario_conteos_ciclicos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			producto_id INTEGER NOT NULL,
			bodega_id INTEGER NOT NULL,
			cantidad_sistema REAL NOT NULL DEFAULT 0,
			cantidad_contada REAL NOT NULL DEFAULT 0,
			variacion REAL NOT NULL DEFAULT 0,
			tipo_ajuste TEXT NOT NULL DEFAULT 'sin_ajuste',
			movimiento_id INTEGER,
			referencia TEXT,
			fecha_conteo TEXT DEFAULT (datetime('now','localtime')),
			usuario_revisor TEXT,
			estado_conteo TEXT DEFAULT 'sin_diferencia',
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_bodegas_empresa_codigo ON bodegas(empresa_id, codigo);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_bodegas_empresa_nombre ON bodegas(empresa_id, nombre);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_categorias_productos_empresa_codigo ON categorias_productos(empresa_id, codigo);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_categorias_productos_empresa_nombre ON categorias_productos(empresa_id, nombre);`,
		`CREATE INDEX IF NOT EXISTS ix_categorias_productos_empresa_orden ON categorias_productos(empresa_id, orden, nombre);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_productos_empresa_sku ON productos(empresa_id, sku);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_productos_empresa_barras ON productos(empresa_id, codigo_barras);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_proveedores_empresa_codigo ON proveedores(empresa_id, codigo);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_proveedores_empresa_nombre ON proveedores(empresa_id, nombre);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_servicios_empresa_codigo ON servicios(empresa_id, codigo);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_combos_empresa_codigo ON combos_productos(empresa_id, codigo);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_combos_empresa_nombre ON combos_productos(empresa_id, nombre);`,
		`CREATE INDEX IF NOT EXISTS ix_combos_empresa_estado ON combos_productos(empresa_id, estado, nombre);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_combos_detalle_combo_producto ON combos_productos_detalle(empresa_id, combo_id, producto_id);`,
		`CREATE INDEX IF NOT EXISTS ix_combos_detalle_empresa_producto ON combos_productos_detalle(empresa_id, producto_id);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_combos_versiones_empresa_combo_version ON combos_productos_versiones(empresa_id, combo_id, receta_version);`,
		`CREATE INDEX IF NOT EXISTS ix_combos_versiones_empresa_combo_fecha ON combos_productos_versiones(empresa_id, combo_id, fecha_creacion);`,
		`CREATE INDEX IF NOT EXISTS ix_historial_precios_empresa_producto ON producto_precios_historial(empresa_id, producto_id);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_existencias_empresa_prod_bodega ON inventario_existencias(empresa_id, producto_id, bodega_id);`,
		`CREATE INDEX IF NOT EXISTS ix_existencias_empresa_bodega ON inventario_existencias(empresa_id, bodega_id);`,
		`CREATE INDEX IF NOT EXISTS ix_movimientos_empresa_producto ON inventario_movimientos(empresa_id, producto_id);`,
		`CREATE INDEX IF NOT EXISTS ix_costos_lotes_empresa_producto_bodega ON inventario_costos_lotes(empresa_id, producto_id, bodega_id, fecha_lote, id);`,
		`CREATE INDEX IF NOT EXISTS ix_costos_lotes_empresa_estado ON inventario_costos_lotes(empresa_id, estado, producto_id);`,
		`CREATE INDEX IF NOT EXISTS ix_conteos_ciclicos_empresa_fecha ON inventario_conteos_ciclicos(empresa_id, fecha_conteo);`,
		`CREATE INDEX IF NOT EXISTS ix_conteos_ciclicos_empresa_producto_bodega ON inventario_conteos_ciclicos(empresa_id, producto_id, bodega_id);`,
	}

	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	// Migraciones seguras para instalaciones previas.
	if err := ensureColumnIfMissing(dbConn, "bodegas", "codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "bodegas", "ubicacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "bodegas", "responsable", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "bodegas", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "bodegas", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "bodegas", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "bodegas", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "categorias_productos", "codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "categorias_productos", "nombre", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "categorias_productos", "descripcion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "categorias_productos", "color_hex", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "categorias_productos", "orden", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "categorias_productos", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "categorias_productos", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "categorias_productos", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "categorias_productos", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "productos", "bodega_principal_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "proveedor_principal_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "categoria_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "codigo_barras", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "categoria", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "marca", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "unidad_medida", "TEXT DEFAULT 'unidad'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "costo", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "precio", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "impuesto_porcentaje", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "stock_minimo", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "stock_maximo", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "imagen_url", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "productos", "observaciones", "TEXT"); err != nil {
		return err
	}
	if _, err := dbConn.Exec(`CREATE INDEX IF NOT EXISTS ix_productos_empresa_categoria_id ON productos(empresa_id, categoria_id);`); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "proveedores", "codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "proveedores", "documento", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "proveedores", "contacto", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "proveedores", "telefono", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "proveedores", "email", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "proveedores", "direccion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "proveedores", "catalogo_referencia", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "proveedores", "precio_base_referencial", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "proveedores", "descuento_porcentaje", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "proveedores", "plazo_pago_dias", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "proveedores", "condicion_entrega", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "proveedores", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "proveedores", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "proveedores", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "proveedores", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "servicios", "codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "servicios", "categoria", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "servicios", "duracion_minutos", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "servicios", "costo_referencial", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "servicios", "precio", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "servicios", "impuesto_porcentaje", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "servicios", "imagen_url", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "servicios", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "servicios", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "servicios", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "servicios", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "combos_productos", "fecha_creacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos", "fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos", "observaciones", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos", "empresa_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos", "codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos", "nombre", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos", "descripcion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos", "unidad_medida", "TEXT DEFAULT 'combo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos", "precio", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos", "impuesto_porcentaje", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos", "receta_version", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos", "costo_teorico", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos", "costo_real", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos", "variacion_costo", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos", "variacion_costo_porcentaje", "REAL DEFAULT 0"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "combos_productos_detalle", "fecha_creacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos_detalle", "fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos_detalle", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos_detalle", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos_detalle", "observaciones", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos_detalle", "empresa_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos_detalle", "combo_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos_detalle", "producto_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos_detalle", "cantidad", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos_detalle", "unidad_medida", "TEXT DEFAULT 'unidad'"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "combos_productos_versiones", "fecha_creacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos_versiones", "fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos_versiones", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos_versiones", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos_versiones", "observaciones", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos_versiones", "empresa_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos_versiones", "combo_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos_versiones", "receta_version", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos_versiones", "ingredientes_json", "TEXT DEFAULT '[]'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos_versiones", "costo_teorico", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos_versiones", "costo_real", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos_versiones", "variacion_costo", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos_versiones", "variacion_costo_porcentaje", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "combos_productos_versiones", "motivo", "TEXT"); err != nil {
		return err
	}
	if _, err := dbConn.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_combos_versiones_empresa_combo_version ON combos_productos_versiones(empresa_id, combo_id, receta_version);`); err != nil {
		return err
	}
	if _, err := dbConn.Exec(`CREATE INDEX IF NOT EXISTS ix_combos_versiones_empresa_combo_fecha ON combos_productos_versiones(empresa_id, combo_id, fecha_creacion);`); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "costo_anterior", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "costo_nuevo", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "precio_anterior", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "precio_nuevo", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "impuesto_anterior", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "impuesto_nuevo", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "motivo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "referencia", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "fecha_cambio", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "producto_precios_historial", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "inventario_existencias", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "inventario_existencias", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "inventario_existencias", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "inventario_existencias", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "inventario_movimientos", "fecha_movimiento", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "inventario_movimientos", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "inventario_movimientos", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "inventario_movimientos", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "inventario_movimientos", "observaciones", "TEXT"); err != nil {
		return err
	}
	if err := seedCategoriasProductosFromLegacy(dbConn); err != nil {
		return err
	}
	if err := backfillProductoCategoriaIDs(dbConn); err != nil {
		return err
	}

	return nil
}

func ensureColumnIfMissing(dbConn *sql.DB, table, column, columnDef string) error {
	lowerColumn := strings.ToLower(column)
	if isPostgresDialect() {
		rows, err := querySQLCompat(dbConn, `
			SELECT column_name
			FROM information_schema.columns
			WHERE table_schema = ANY (current_schemas(false))
			  AND table_name = ?
		`, table)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return err
			}
			if strings.ToLower(name) == lowerColumn {
				return nil
			}
		}

		columnDef = normalizeColumnDefForDialect(columnDef)
		_, err = execSQLCompat(dbConn, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", table, column+" "+columnDef))
		return err
	}

	rows, err := querySQLCompat(dbConn, fmt.Sprintf("PRAGMA table_info(%s);", table))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return err
		}
		if strings.ToLower(name) == lowerColumn {
			return nil
		}
	}
	_, err = execSQLCompat(dbConn, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", table, column+" "+columnDef))
	return err
}

// CreateBodega inserta una nueva bodega para una empresa.
func CreateBodega(dbConn *sql.DB, b Bodega) (int64, error) {
	res, err := dbConn.Exec(`INSERT INTO bodegas (empresa_id, codigo, nombre, ubicacion, responsable, usuario_creador, estado, observaciones, fecha_creacion, fecha_actualizacion)
		VALUES (?, ?, ?, ?, ?, ?, COALESCE(NULLIF(?, ''), 'activo'), ?, datetime('now','localtime'), datetime('now','localtime'))`,
		b.EmpresaID, strings.TrimSpace(b.Codigo), strings.TrimSpace(b.Nombre), strings.TrimSpace(b.Ubicacion), strings.TrimSpace(b.Responsable), strings.TrimSpace(b.UsuarioCreador), strings.TrimSpace(b.Estado), strings.TrimSpace(b.Observaciones))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetBodegasByEmpresa lista bodegas por empresa.
func GetBodegasByEmpresa(dbConn *sql.DB, empresaID int64, incluirInactivas bool) ([]Bodega, error) {
	query := `SELECT id, empresa_id, codigo, nombre, ubicacion, responsable, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
		FROM bodegas
		WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	if !incluirInactivas {
		query += ` AND estado = 'activo'`
	}
	query += ` ORDER BY nombre ASC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Bodega, 0)
	for rows.Next() {
		var b Bodega
		var codigo, ubicacion, responsable, fechaCreacion, fechaAct, usuario, estado, obs sql.NullString
		if err := rows.Scan(&b.ID, &b.EmpresaID, &codigo, &b.Nombre, &ubicacion, &responsable, &fechaCreacion, &fechaAct, &usuario, &estado, &obs); err != nil {
			return nil, err
		}
		if codigo.Valid {
			b.Codigo = codigo.String
		}
		if ubicacion.Valid {
			b.Ubicacion = ubicacion.String
		}
		if responsable.Valid {
			b.Responsable = responsable.String
		}
		if fechaCreacion.Valid {
			b.FechaCreacion = fechaCreacion.String
		}
		if fechaAct.Valid {
			b.FechaActualizacion = fechaAct.String
		}
		if usuario.Valid {
			b.UsuarioCreador = usuario.String
		}
		if estado.Valid {
			b.Estado = estado.String
		}
		if obs.Valid {
			b.Observaciones = obs.String
		}
		out = append(out, b)
	}
	return out, nil
}

// UpdateBodega actualiza los datos editables de una bodega.
func UpdateBodega(dbConn *sql.DB, b Bodega) error {
	_, err := dbConn.Exec(`UPDATE bodegas
		SET codigo = ?, nombre = ?, ubicacion = ?, responsable = ?, observaciones = ?, fecha_actualizacion = datetime('now','localtime')
		WHERE id = ? AND empresa_id = ?`,
		strings.TrimSpace(b.Codigo), strings.TrimSpace(b.Nombre), strings.TrimSpace(b.Ubicacion), strings.TrimSpace(b.Responsable), strings.TrimSpace(b.Observaciones), b.ID, b.EmpresaID)
	return err
}

// DeleteBodega elimina una bodega de la empresa.
func DeleteBodega(dbConn *sql.DB, empresaID, bodegaID int64) error {
	_, err := dbConn.Exec("DELETE FROM bodegas WHERE id = ? AND empresa_id = ?", bodegaID, empresaID)
	return err
}

// SetBodegaEstado activa/desactiva una bodega.
func SetBodegaEstado(dbConn *sql.DB, empresaID, bodegaID int64, estado string) error {
	_, err := dbConn.Exec("UPDATE bodegas SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE id = ? AND empresa_id = ?", strings.TrimSpace(estado), bodegaID, empresaID)
	return err
}

// CreateCategoriaProducto inserta una categoría de producto para una empresa.
func CreateCategoriaProducto(dbConn *sql.DB, c CategoriaProducto) (int64, error) {
	res, err := dbConn.Exec(`INSERT INTO categorias_productos (
		empresa_id, codigo, nombre, descripcion, color_hex, orden,
		usuario_creador, estado, observaciones, fecha_creacion, fecha_actualizacion
	) VALUES (?, NULLIF(?, ''), ?, ?, ?, ?, ?, COALESCE(NULLIF(?, ''), 'activo'), ?, datetime('now','localtime'), datetime('now','localtime'))`,
		c.EmpresaID, strings.TrimSpace(c.Codigo), strings.TrimSpace(c.Nombre), strings.TrimSpace(c.Descripcion), strings.TrimSpace(c.ColorHex), c.Orden,
		strings.TrimSpace(c.UsuarioCreador), strings.TrimSpace(c.Estado), strings.TrimSpace(c.Observaciones))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetCategoriasProductoByEmpresa lista categorías de producto por empresa.
func GetCategoriasProductoByEmpresa(dbConn *sql.DB, empresaID int64, incluirInactivas bool, filtro string) ([]CategoriaProducto, error) {
	query := `SELECT id, empresa_id, codigo, nombre, descripcion, color_hex, orden,
		fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
	FROM categorias_productos
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	if !incluirInactivas {
		query += ` AND estado = 'activo'`
	}
	if strings.TrimSpace(filtro) != "" {
		like := "%" + strings.TrimSpace(filtro) + "%"
		query += ` AND (nombre LIKE ? OR codigo LIKE ? OR descripcion LIKE ?)`
		args = append(args, like, like, like)
	}
	query += ` ORDER BY orden ASC, nombre ASC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]CategoriaProducto, 0)
	for rows.Next() {
		var c CategoriaProducto
		var codigo, descripcion, colorHex, fechaCreacion, fechaActualizacion, usuario, estado, observaciones sql.NullString
		if err := rows.Scan(&c.ID, &c.EmpresaID, &codigo, &c.Nombre, &descripcion, &colorHex, &c.Orden, &fechaCreacion, &fechaActualizacion, &usuario, &estado, &observaciones); err != nil {
			return nil, err
		}
		if codigo.Valid {
			c.Codigo = codigo.String
		}
		if descripcion.Valid {
			c.Descripcion = descripcion.String
		}
		if colorHex.Valid {
			c.ColorHex = colorHex.String
		}
		if fechaCreacion.Valid {
			c.FechaCreacion = fechaCreacion.String
		}
		if fechaActualizacion.Valid {
			c.FechaActualizacion = fechaActualizacion.String
		}
		if usuario.Valid {
			c.UsuarioCreador = usuario.String
		}
		if estado.Valid {
			c.Estado = estado.String
		}
		if observaciones.Valid {
			c.Observaciones = observaciones.String
		}
		out = append(out, c)
	}
	return out, nil
}

// UpdateCategoriaProducto actualiza una categoría y sincroniza el nombre en productos asociados.
func UpdateCategoriaProducto(dbConn *sql.DB, c CategoriaProducto) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE categorias_productos
		SET codigo = NULLIF(?, ''), nombre = ?, descripcion = ?, color_hex = ?, orden = ?, observaciones = ?, fecha_actualizacion = datetime('now','localtime')
		WHERE id = ? AND empresa_id = ?`,
		strings.TrimSpace(c.Codigo), strings.TrimSpace(c.Nombre), strings.TrimSpace(c.Descripcion), strings.TrimSpace(c.ColorHex), c.Orden, strings.TrimSpace(c.Observaciones), c.ID, c.EmpresaID); err != nil {
		return err
	}

	if _, err := tx.Exec(`UPDATE productos
		SET categoria = ?, fecha_actualizacion = datetime('now','localtime')
		WHERE empresa_id = ? AND categoria_id = ?`, strings.TrimSpace(c.Nombre), c.EmpresaID, c.ID); err != nil {
		return err
	}

	return tx.Commit()
}

// DeleteCategoriaProducto elimina una categoría si no está asociada a productos.
func DeleteCategoriaProducto(dbConn *sql.DB, empresaID, categoriaID int64) error {
	var totalProductos int64
	if err := dbConn.QueryRow(`SELECT COUNT(1) FROM productos WHERE empresa_id = ? AND categoria_id = ?`, empresaID, categoriaID).Scan(&totalProductos); err != nil {
		return err
	}
	if totalProductos > 0 {
		return fmt.Errorf("no se puede eliminar la categoría porque está asociada a productos")
	}
	_, err := dbConn.Exec(`DELETE FROM categorias_productos WHERE id = ? AND empresa_id = ?`, categoriaID, empresaID)
	return err
}

// SetCategoriaProductoEstado activa/desactiva una categoría de producto.
func SetCategoriaProductoEstado(dbConn *sql.DB, empresaID, categoriaID int64, estado string) error {
	_, err := dbConn.Exec(`UPDATE categorias_productos SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE id = ? AND empresa_id = ?`, strings.TrimSpace(estado), categoriaID, empresaID)
	return err
}

func validateProductoStockThresholds(stockMinimo, stockMaximo float64) error {
	if stockMinimo < 0 || stockMaximo < 0 {
		return fmt.Errorf("stock_minimo y stock_maximo no pueden ser negativos")
	}
	if stockMaximo > 0 && stockMinimo > stockMaximo {
		return fmt.Errorf("stock_minimo no puede ser mayor que stock_maximo")
	}
	return nil
}

// CreateProducto crea un producto y opcionalmente su stock inicial.
func CreateProducto(dbConn *sql.DB, p Producto, stockInicial float64, referenciaInicial string) (int64, error) {
	if err := validateProductoStockThresholds(p.StockMinimo, p.StockMaximo); err != nil {
		return 0, err
	}

	tx, err := dbConn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if p.BodegaPrincipalID > 0 {
		if err := validateBodegaEmpresaTx(tx, p.EmpresaID, p.BodegaPrincipalID); err != nil {
			return 0, err
		}
	}
	if p.ProveedorPrincipalID > 0 {
		if err := validateProveedorEmpresaTx(tx, p.EmpresaID, p.ProveedorPrincipalID); err != nil {
			return 0, err
		}
	}
	if p.CategoriaID > 0 {
		categoriaNombre, err := resolveCategoriaProductoTx(tx, p.EmpresaID, p.CategoriaID)
		if err != nil {
			return 0, err
		}
		p.Categoria = categoriaNombre
	}

	res, err := tx.Exec(`INSERT INTO productos (
		empresa_id, bodega_principal_id, proveedor_principal_id, categoria_id, sku, codigo_barras, nombre, descripcion, categoria, marca, unidad_medida,
		costo, precio, impuesto_porcentaje, stock_minimo, stock_maximo, imagen_url,
		usuario_creador, estado, observaciones, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, COALESCE(NULLIF(?, ''), 'activo'), ?, datetime('now','localtime'), datetime('now','localtime'))`,
		p.EmpresaID, nullableInt64(p.BodegaPrincipalID), nullableInt64(p.ProveedorPrincipalID), nullableInt64(p.CategoriaID), strings.TrimSpace(p.SKU), strings.TrimSpace(p.CodigoBarras), strings.TrimSpace(p.Nombre), strings.TrimSpace(p.Descripcion), strings.TrimSpace(p.Categoria), strings.TrimSpace(p.Marca), defaultUnidad(p.UnidadMedida),
		p.Costo, p.Precio, p.ImpuestoPorcentaje, p.StockMinimo, p.StockMaximo, strings.TrimSpace(p.ImagenURL),
		strings.TrimSpace(p.UsuarioCreador), strings.TrimSpace(p.Estado), strings.TrimSpace(p.Observaciones))
	if err != nil {
		return 0, err
	}
	productoID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if stockInicial > 0 && p.BodegaPrincipalID > 0 {
		if err := upsertExistenciaTx(tx, p.EmpresaID, productoID, p.BodegaPrincipalID, stockInicial, p.UsuarioCreador, "entrada inicial"); err != nil {
			return 0, err
		}
		if err := insertMovimientoTx(tx, InventarioMovimiento{
			EmpresaID:       p.EmpresaID,
			ProductoID:      productoID,
			BodegaDestinoID: p.BodegaPrincipalID,
			Tipo:            "entrada",
			Cantidad:        stockInicial,
			CostoUnitario:   p.Costo,
			Referencia:      strings.TrimSpace(referenciaInicial),
			UsuarioCreador:  strings.TrimSpace(p.UsuarioCreador),
			Estado:          "activo",
			Observaciones:   "stock inicial al crear producto",
		}); err != nil {
			return 0, err
		}
	}

	if err := insertProductoPrecioHistorialTx(tx, ProductoPrecioHistorial{
		EmpresaID:        p.EmpresaID,
		ProductoID:       productoID,
		CostoAnterior:    0,
		CostoNuevo:       p.Costo,
		PrecioAnterior:   0,
		PrecioNuevo:      p.Precio,
		ImpuestoAnterior: 0,
		ImpuestoNuevo:    p.ImpuestoPorcentaje,
		Motivo:           "creacion_producto",
		Referencia:       strings.TrimSpace(referenciaInicial),
		UsuarioCreador:   strings.TrimSpace(p.UsuarioCreador),
		Estado:           "activo",
		Observaciones:    "historial inicial de precio",
	}); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return productoID, nil
}

// GetProductosByEmpresa lista productos con filtros y stock total.
func GetProductosByEmpresa(dbConn *sql.DB, empresaID int64, filtro, estado string, bodegaID, categoriaID int64, limit, offset int) ([]Producto, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	query := `SELECT
		p.id, p.empresa_id, p.bodega_principal_id, p.proveedor_principal_id, p.categoria_id, p.sku, p.codigo_barras, p.nombre, p.descripcion, COALESCE(NULLIF(cp.nombre, ''), p.categoria), p.marca, p.unidad_medida,
		p.costo, p.precio, p.impuesto_porcentaje, p.stock_minimo, p.stock_maximo, p.imagen_url,
		p.fecha_creacion, p.fecha_actualizacion, p.usuario_creador, p.estado, p.observaciones,
		COALESCE(SUM(e.cantidad), 0) AS stock_total
	FROM productos p
	LEFT JOIN categorias_productos cp ON cp.empresa_id = p.empresa_id AND cp.id = p.categoria_id
	LEFT JOIN inventario_existencias e ON e.empresa_id = p.empresa_id AND e.producto_id = p.id
	WHERE p.empresa_id = ?`

	args := []interface{}{empresaID}
	if strings.TrimSpace(estado) != "" {
		query += " AND p.estado = ?"
		args = append(args, strings.TrimSpace(estado))
	}
	if strings.TrimSpace(filtro) != "" {
		like := "%" + strings.TrimSpace(filtro) + "%"
		query += " AND (p.nombre LIKE ? OR p.sku LIKE ? OR p.codigo_barras LIKE ? OR p.marca LIKE ? OR p.categoria LIKE ? OR cp.nombre LIKE ?)"
		args = append(args, like, like, like, like, like, like)
	}
	if bodegaID > 0 {
		query += ` AND EXISTS (
			SELECT 1 FROM inventario_existencias ex
			WHERE ex.empresa_id = p.empresa_id AND ex.producto_id = p.id AND ex.bodega_id = ?
		)`
		args = append(args, bodegaID)
	}
	if categoriaID > 0 {
		query += ` AND p.categoria_id = ?`
		args = append(args, categoriaID)
	}

	query += ` GROUP BY p.id ORDER BY p.id DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Producto, 0)
	for rows.Next() {
		var p Producto
		var bodegaPrincipalID sql.NullInt64
		var proveedorPrincipalID sql.NullInt64
		var categoriaIDVal sql.NullInt64
		var sku, codigoBarras, desc, categoria, marca, unidad, imagenURL, fechaCre, fechaAct, usuario, estadoVal, obs sql.NullString
		if err := rows.Scan(
			&p.ID, &p.EmpresaID, &bodegaPrincipalID, &proveedorPrincipalID, &categoriaIDVal, &sku, &codigoBarras, &p.Nombre, &desc, &categoria, &marca, &unidad,
			&p.Costo, &p.Precio, &p.ImpuestoPorcentaje, &p.StockMinimo, &p.StockMaximo, &imagenURL,
			&fechaCre, &fechaAct, &usuario, &estadoVal, &obs,
			&p.StockTotal,
		); err != nil {
			return nil, err
		}
		if bodegaPrincipalID.Valid {
			p.BodegaPrincipalID = bodegaPrincipalID.Int64
		}
		if proveedorPrincipalID.Valid {
			p.ProveedorPrincipalID = proveedorPrincipalID.Int64
		}
		if categoriaIDVal.Valid {
			p.CategoriaID = categoriaIDVal.Int64
		}
		if sku.Valid {
			p.SKU = sku.String
		}
		if codigoBarras.Valid {
			p.CodigoBarras = codigoBarras.String
		}
		if desc.Valid {
			p.Descripcion = desc.String
		}
		if categoria.Valid {
			p.Categoria = categoria.String
		}
		if marca.Valid {
			p.Marca = marca.String
		}
		if unidad.Valid {
			p.UnidadMedida = unidad.String
		}
		if imagenURL.Valid {
			p.ImagenURL = imagenURL.String
		}
		if fechaCre.Valid {
			p.FechaCreacion = fechaCre.String
		}
		if fechaAct.Valid {
			p.FechaActualizacion = fechaAct.String
		}
		if usuario.Valid {
			p.UsuarioCreador = usuario.String
		}
		if estadoVal.Valid {
			p.Estado = estadoVal.String
		}
		if obs.Valid {
			p.Observaciones = obs.String
		}
		out = append(out, p)
	}

	return out, nil
}

// GetProductoByID devuelve un producto específico por empresa.
func GetProductoByID(dbConn *sql.DB, empresaID, productoID int64) (*Producto, error) {
	query := `SELECT p.id, p.empresa_id, p.bodega_principal_id, p.proveedor_principal_id, p.categoria_id, p.sku, p.codigo_barras, p.nombre, p.descripcion, COALESCE(NULLIF(cp.nombre, ''), p.categoria), p.marca, p.unidad_medida,
		p.costo, p.precio, p.impuesto_porcentaje, p.stock_minimo, p.stock_maximo, p.imagen_url,
		p.fecha_creacion, p.fecha_actualizacion, p.usuario_creador, p.estado, p.observaciones
	FROM productos p
	LEFT JOIN categorias_productos cp ON cp.empresa_id = p.empresa_id AND cp.id = p.categoria_id
	WHERE p.empresa_id = ? AND p.id = ? LIMIT 1`

	row := dbConn.QueryRow(query, empresaID, productoID)
	var p Producto
	var bodegaPrincipalID sql.NullInt64
	var proveedorPrincipalID sql.NullInt64
	var categoriaID sql.NullInt64
	var sku, codigoBarras, desc, categoria, marca, unidad, imagenURL, fechaCre, fechaAct, usuario, estadoVal, obs sql.NullString
	if err := row.Scan(
		&p.ID, &p.EmpresaID, &bodegaPrincipalID, &proveedorPrincipalID, &categoriaID, &sku, &codigoBarras, &p.Nombre, &desc, &categoria, &marca, &unidad,
		&p.Costo, &p.Precio, &p.ImpuestoPorcentaje, &p.StockMinimo, &p.StockMaximo, &imagenURL,
		&fechaCre, &fechaAct, &usuario, &estadoVal, &obs,
	); err != nil {
		return nil, err
	}
	if bodegaPrincipalID.Valid {
		p.BodegaPrincipalID = bodegaPrincipalID.Int64
	}
	if proveedorPrincipalID.Valid {
		p.ProveedorPrincipalID = proveedorPrincipalID.Int64
	}
	if categoriaID.Valid {
		p.CategoriaID = categoriaID.Int64
	}
	if sku.Valid {
		p.SKU = sku.String
	}
	if codigoBarras.Valid {
		p.CodigoBarras = codigoBarras.String
	}
	if desc.Valid {
		p.Descripcion = desc.String
	}
	if categoria.Valid {
		p.Categoria = categoria.String
	}
	if marca.Valid {
		p.Marca = marca.String
	}
	if unidad.Valid {
		p.UnidadMedida = unidad.String
	}
	if imagenURL.Valid {
		p.ImagenURL = imagenURL.String
	}
	if fechaCre.Valid {
		p.FechaCreacion = fechaCre.String
	}
	if fechaAct.Valid {
		p.FechaActualizacion = fechaAct.String
	}
	if usuario.Valid {
		p.UsuarioCreador = usuario.String
	}
	if estadoVal.Valid {
		p.Estado = estadoVal.String
	}
	if obs.Valid {
		p.Observaciones = obs.String
	}
	return &p, nil
}

// UpdateProducto actualiza un producto de la empresa.
func UpdateProducto(dbConn *sql.DB, p Producto, motivoCambio, referenciaCambio string) error {
	if err := validateProductoStockThresholds(p.StockMinimo, p.StockMaximo); err != nil {
		return err
	}

	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if p.BodegaPrincipalID > 0 {
		if err := validateBodegaEmpresaTx(tx, p.EmpresaID, p.BodegaPrincipalID); err != nil {
			return err
		}
	}
	if p.ProveedorPrincipalID > 0 {
		if err := validateProveedorEmpresaTx(tx, p.EmpresaID, p.ProveedorPrincipalID); err != nil {
			return err
		}
	}
	if p.CategoriaID > 0 {
		categoriaNombre, err := resolveCategoriaProductoTx(tx, p.EmpresaID, p.CategoriaID)
		if err != nil {
			return err
		}
		p.Categoria = categoriaNombre
	}

	var costoAnterior, precioAnterior, impuestoAnterior float64
	if err := tx.QueryRow(`SELECT costo, precio, impuesto_porcentaje FROM productos WHERE id = ? AND empresa_id = ? LIMIT 1`, p.ID, p.EmpresaID).Scan(&costoAnterior, &precioAnterior, &impuestoAnterior); err != nil {
		return err
	}

	if _, err := tx.Exec(`UPDATE productos
		SET bodega_principal_id = ?, proveedor_principal_id = ?, categoria_id = ?, sku = NULLIF(?, ''), codigo_barras = NULLIF(?, ''), nombre = ?, descripcion = ?, categoria = ?, marca = ?, unidad_medida = ?,
			costo = ?, precio = ?, impuesto_porcentaje = ?, stock_minimo = ?, stock_maximo = ?, imagen_url = ?, observaciones = ?, fecha_actualizacion = datetime('now','localtime')
		WHERE id = ? AND empresa_id = ?`,
		nullableInt64(p.BodegaPrincipalID), nullableInt64(p.ProveedorPrincipalID), nullableInt64(p.CategoriaID), strings.TrimSpace(p.SKU), strings.TrimSpace(p.CodigoBarras), strings.TrimSpace(p.Nombre), strings.TrimSpace(p.Descripcion), strings.TrimSpace(p.Categoria), strings.TrimSpace(p.Marca), defaultUnidad(p.UnidadMedida),
		p.Costo, p.Precio, p.ImpuestoPorcentaje, p.StockMinimo, p.StockMaximo, strings.TrimSpace(p.ImagenURL), strings.TrimSpace(p.Observaciones), p.ID, p.EmpresaID); err != nil {
		return err
	}

	if costoAnterior != p.Costo || precioAnterior != p.Precio || impuestoAnterior != p.ImpuestoPorcentaje {
		if strings.TrimSpace(motivoCambio) == "" {
			motivoCambio = "actualizacion_precio"
		}
		if err := insertProductoPrecioHistorialTx(tx, ProductoPrecioHistorial{
			EmpresaID:        p.EmpresaID,
			ProductoID:       p.ID,
			CostoAnterior:    costoAnterior,
			CostoNuevo:       p.Costo,
			PrecioAnterior:   precioAnterior,
			PrecioNuevo:      p.Precio,
			ImpuestoAnterior: impuestoAnterior,
			ImpuestoNuevo:    p.ImpuestoPorcentaje,
			Motivo:           strings.TrimSpace(motivoCambio),
			Referencia:       strings.TrimSpace(referenciaCambio),
			UsuarioCreador:   strings.TrimSpace(p.UsuarioCreador),
			Estado:           "activo",
			Observaciones:    "historial automático por cambio de precio",
		}); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// DeleteProducto elimina un producto y su data de inventario asociada.
func DeleteProducto(dbConn *sql.DB, empresaID, productoID int64) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM inventario_movimientos WHERE empresa_id = ? AND producto_id = ?", empresaID, productoID); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM inventario_existencias WHERE empresa_id = ? AND producto_id = ?", empresaID, productoID); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM productos WHERE empresa_id = ? AND id = ?", empresaID, productoID); err != nil {
		return err
	}
	return tx.Commit()
}

// SetProductoEstado activa/desactiva un producto.
func SetProductoEstado(dbConn *sql.DB, empresaID, productoID int64, estado string) error {
	_, err := dbConn.Exec("UPDATE productos SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE id = ? AND empresa_id = ?", strings.TrimSpace(estado), productoID, empresaID)
	return err
}

// UpdateProductoImagen actualiza la URL de imagen/logo de un producto.
func UpdateProductoImagen(dbConn *sql.DB, empresaID, productoID int64, imagenURL string) error {
	_, err := dbConn.Exec("UPDATE productos SET imagen_url = ?, fecha_actualizacion = datetime('now','localtime') WHERE id = ? AND empresa_id = ?", strings.TrimSpace(imagenURL), productoID, empresaID)
	return err
}

// GetExistenciasByEmpresa devuelve stock por bodega y producto.
func GetExistenciasByEmpresa(dbConn *sql.DB, empresaID, productoID, bodegaID int64, limit int, offset int) ([]InventarioExistencia, error) {
	if limit <= 0 || limit > 1000 {
		limit = 300
	}
	if offset < 0 {
		offset = 0
	}

	query := `SELECT e.id, e.empresa_id, e.producto_id, p.nombre, e.bodega_id, b.nombre, e.cantidad,
		e.fecha_creacion, e.fecha_actualizacion, e.usuario_creador, e.estado, e.observaciones
	FROM inventario_existencias e
	JOIN productos p ON p.id = e.producto_id AND p.empresa_id = e.empresa_id
	JOIN bodegas b ON b.id = e.bodega_id AND b.empresa_id = e.empresa_id
	WHERE e.empresa_id = ?`
	args := []interface{}{empresaID}
	if productoID > 0 {
		query += " AND e.producto_id = ?"
		args = append(args, productoID)
	}
	if bodegaID > 0 {
		query += " AND e.bodega_id = ?"
		args = append(args, bodegaID)
	}
	query += " ORDER BY p.nombre ASC, b.nombre ASC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]InventarioExistencia, 0)
	for rows.Next() {
		var e InventarioExistencia
		var fechaCre, fechaAct, usuario, estado, obs sql.NullString
		if err := rows.Scan(&e.ID, &e.EmpresaID, &e.ProductoID, &e.ProductoNombre, &e.BodegaID, &e.BodegaNombre, &e.Cantidad, &fechaCre, &fechaAct, &usuario, &estado, &obs); err != nil {
			return nil, err
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

func normalizeInventarioPoliticaCosto(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case inventarioPoliticaCostoPEPS, "fifo":
		return inventarioPoliticaCostoPEPS
	default:
		return inventarioPoliticaCostoPromedio
	}
}

// GetEmpresaInventarioConfiguracion obtiene la configuracion operativa de inventario por empresa.
func GetEmpresaInventarioConfiguracion(dbConn *sql.DB, empresaID int64) (EmpresaInventarioConfiguracion, error) {
	conf := EmpresaInventarioConfiguracion{
		EmpresaID:     empresaID,
		PoliticaCosto: inventarioPoliticaCostoPromedio,
		Estado:        "activo",
	}

	row := dbConn.QueryRow(`SELECT id, empresa_id, politica_costo, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
		FROM empresa_inventario_configuracion
		WHERE empresa_id = ?
		LIMIT 1`, empresaID)

	var politicaCosto sql.NullString
	var fechaCre, fechaAct, usuario, estado, obs sql.NullString
	err := row.Scan(&conf.ID, &conf.EmpresaID, &politicaCosto, &fechaCre, &fechaAct, &usuario, &estado, &obs)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return conf, nil
		}
		return conf, err
	}

	conf.PoliticaCosto = normalizeInventarioPoliticaCosto(politicaCosto.String)
	if fechaCre.Valid {
		conf.FechaCreacion = fechaCre.String
	}
	if fechaAct.Valid {
		conf.FechaActualizacion = fechaAct.String
	}
	if usuario.Valid {
		conf.UsuarioCreador = usuario.String
	}
	if estado.Valid {
		conf.Estado = estado.String
	}
	if obs.Valid {
		conf.Observaciones = obs.String
	}

	return conf, nil
}

// UpsertEmpresaInventarioConfiguracion crea o actualiza la politica de costo por empresa.
func UpsertEmpresaInventarioConfiguracion(dbConn *sql.DB, conf EmpresaInventarioConfiguracion) (EmpresaInventarioConfiguracion, error) {
	if conf.EmpresaID <= 0 {
		return EmpresaInventarioConfiguracion{}, fmt.Errorf("empresa_id invalido")
	}

	conf.PoliticaCosto = normalizeInventarioPoliticaCosto(conf.PoliticaCosto)
	if strings.TrimSpace(conf.Estado) == "" {
		conf.Estado = "activo"
	}

	_, err := dbConn.Exec(`INSERT INTO empresa_inventario_configuracion (
		empresa_id,
		politica_costo,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, datetime('now','localtime'), datetime('now','localtime'), ?, COALESCE(NULLIF(?, ''), 'activo'), ?)
	ON CONFLICT(empresa_id) DO UPDATE SET
		politica_costo = excluded.politica_costo,
		fecha_actualizacion = datetime('now','localtime'),
		usuario_creador = CASE
			WHEN TRIM(COALESCE(excluded.usuario_creador, '')) <> '' THEN excluded.usuario_creador
			ELSE empresa_inventario_configuracion.usuario_creador
		END,
		estado = COALESCE(NULLIF(excluded.estado, ''), empresa_inventario_configuracion.estado),
		observaciones = excluded.observaciones`,
		conf.EmpresaID,
		conf.PoliticaCosto,
		strings.TrimSpace(conf.UsuarioCreador),
		strings.TrimSpace(conf.Estado),
		strings.TrimSpace(conf.Observaciones),
	)
	if err != nil {
		return EmpresaInventarioConfiguracion{}, err
	}

	return GetEmpresaInventarioConfiguracion(dbConn, conf.EmpresaID)
}

// GetAlertasQuiebreByEmpresa devuelve alertas de quiebre/bajo minimo por producto y bodega.
func GetAlertasQuiebreByEmpresa(dbConn *sql.DB, empresaID, productoID, bodegaID int64, limit int, offset int) ([]InventarioAlertaQuiebre, error) {
	if limit <= 0 || limit > 1000 {
		limit = 300
	}
	if offset < 0 {
		offset = 0
	}

	query := `SELECT
		e.empresa_id,
		e.producto_id,
		p.nombre,
		e.bodega_id,
		b.nombre,
		e.cantidad,
		COALESCE(p.stock_minimo, 0),
		COALESCE(p.stock_maximo, 0)
	FROM inventario_existencias e
	JOIN productos p ON p.id = e.producto_id AND p.empresa_id = e.empresa_id
	JOIN bodegas b ON b.id = e.bodega_id AND b.empresa_id = e.empresa_id
	WHERE e.empresa_id = ?
		AND LOWER(COALESCE(e.estado, 'activo')) = 'activo'
		AND LOWER(COALESCE(p.estado, 'activo')) = 'activo'
		AND (
			COALESCE(e.cantidad, 0) <= 0
			OR (COALESCE(p.stock_minimo, 0) > 0 AND COALESCE(e.cantidad, 0) <= COALESCE(p.stock_minimo, 0))
		)`
	args := []interface{}{empresaID}

	if productoID > 0 {
		query += " AND e.producto_id = ?"
		args = append(args, productoID)
	}
	if bodegaID > 0 {
		query += " AND e.bodega_id = ?"
		args = append(args, bodegaID)
	}

	query += ` ORDER BY
		CASE WHEN COALESCE(e.cantidad, 0) <= 0 THEN 0 ELSE 1 END ASC,
		(COALESCE(p.stock_minimo, 0) - COALESCE(e.cantidad, 0)) DESC,
		p.nombre ASC,
		b.nombre ASC
		LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]InventarioAlertaQuiebre, 0)
	for rows.Next() {
		var a InventarioAlertaQuiebre
		if err := rows.Scan(
			&a.EmpresaID,
			&a.ProductoID,
			&a.ProductoNombre,
			&a.BodegaID,
			&a.BodegaNombre,
			&a.Cantidad,
			&a.StockMinimo,
			&a.StockMaximo,
		); err != nil {
			return nil, err
		}

		a.EstadoStock = "bajo_minimo"
		if a.Cantidad <= 0 {
			a.EstadoStock = "sin_stock"
		}

		if a.Cantidad < a.StockMinimo {
			a.Deficit = a.StockMinimo - a.Cantidad
		}
		if a.Deficit < 0 {
			a.Deficit = 0
		}
		a.SugeridoReposicion = a.Deficit

		out = append(out, a)
	}

	return out, nil
}

// GetAlertasOperativasByEmpresa devuelve alertas proactivas de quiebre y sobrestock por producto/bodega.
func GetAlertasOperativasByEmpresa(dbConn *sql.DB, empresaID, productoID, bodegaID int64, limit int, offset int) ([]InventarioAlertaOperativa, error) {
	if limit <= 0 || limit > 1000 {
		limit = 300
	}
	if offset < 0 {
		offset = 0
	}

	query := `SELECT
		e.empresa_id,
		e.producto_id,
		p.nombre,
		e.bodega_id,
		b.nombre,
		COALESCE(e.cantidad, 0),
		COALESCE(p.stock_minimo, 0),
		COALESCE(p.stock_maximo, 0)
	FROM inventario_existencias e
	JOIN productos p ON p.id = e.producto_id AND p.empresa_id = e.empresa_id
	JOIN bodegas b ON b.id = e.bodega_id AND b.empresa_id = e.empresa_id
	WHERE e.empresa_id = ?
		AND LOWER(COALESCE(e.estado, 'activo')) = 'activo'
		AND LOWER(COALESCE(p.estado, 'activo')) = 'activo'
		AND LOWER(COALESCE(b.estado, 'activo')) = 'activo'
		AND (
			COALESCE(e.cantidad, 0) <= 0
			OR (COALESCE(p.stock_minimo, 0) > 0 AND COALESCE(e.cantidad, 0) <= COALESCE(p.stock_minimo, 0))
			OR (COALESCE(p.stock_maximo, 0) > 0 AND COALESCE(e.cantidad, 0) >= COALESCE(p.stock_maximo, 0))
		)`
	args := []interface{}{empresaID}

	if productoID > 0 {
		query += " AND e.producto_id = ?"
		args = append(args, productoID)
	}
	if bodegaID > 0 {
		query += " AND e.bodega_id = ?"
		args = append(args, bodegaID)
	}

	query += ` ORDER BY p.nombre ASC, b.nombre ASC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]InventarioAlertaOperativa, 0)
	for rows.Next() {
		var a InventarioAlertaOperativa
		if err := rows.Scan(
			&a.EmpresaID,
			&a.ProductoID,
			&a.ProductoNombre,
			&a.BodegaID,
			&a.BodegaNombre,
			&a.Cantidad,
			&a.StockMinimo,
			&a.StockMaximo,
		); err != nil {
			return nil, err
		}

		a.EstadoStock = "estable"
		a.NivelAlerta = "baja"
		a.AccionSugerida = "mantener_monitoreo"

		switch {
		case a.Cantidad <= 0:
			a.EstadoStock = "sin_stock"
			a.NivelAlerta = "critica"
			a.AccionSugerida = "reponer_urgente"
		case a.StockMinimo > 0 && a.Cantidad <= a.StockMinimo:
			a.EstadoStock = "bajo_minimo"
			a.NivelAlerta = "alta"
			a.AccionSugerida = "reposicion_preventiva"
		case a.StockMaximo > 0 && a.Cantidad >= a.StockMaximo:
			a.EstadoStock = "sobrestock"
			a.NivelAlerta = "media"
			a.AccionSugerida = "pausar_compra_promocionar_salida"
		}

		if a.Cantidad < a.StockMinimo {
			a.Deficit = a.StockMinimo - a.Cantidad
			a.SugeridoReposicion = a.Deficit
		}
		if a.StockMaximo > 0 && a.Cantidad > a.StockMaximo {
			a.Exceso = a.Cantidad - a.StockMaximo
		}

		out = append(out, a)
	}

	severidad := func(estado string) int {
		switch estado {
		case "sin_stock":
			return 0
		case "bajo_minimo":
			return 1
		case "sobrestock":
			return 2
		default:
			return 3
		}
	}

	sort.Slice(out, func(i, j int) bool {
		si := severidad(out[i].EstadoStock)
		sj := severidad(out[j].EstadoStock)
		if si != sj {
			return si < sj
		}
		if out[i].Deficit != out[j].Deficit {
			return out[i].Deficit > out[j].Deficit
		}
		if out[i].Exceso != out[j].Exceso {
			return out[i].Exceso > out[j].Exceso
		}
		if out[i].ProductoNombre != out[j].ProductoNombre {
			return out[i].ProductoNombre < out[j].ProductoNombre
		}
		return out[i].BodegaNombre < out[j].BodegaNombre
	})

	return out, nil
}

// GetInventarioResumenByEmpresa devuelve KPI operativos del inventario por empresa.
func GetInventarioResumenByEmpresa(dbConn *sql.DB, empresaID int64, desde, hasta string) (InventarioResumen, error) {
	resumen := InventarioResumen{
		EmpresaID:    empresaID,
		PeriodoDesde: strings.TrimSpace(desde),
		PeriodoHasta: strings.TrimSpace(hasta),
	}

	err := dbConn.QueryRow(`SELECT
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(e.estado, 'activo')) = 'activo' THEN COALESCE(e.cantidad, 0) ELSE 0 END), 0),
		COALESCE(COUNT(DISTINCT CASE WHEN LOWER(COALESCE(e.estado, 'activo')) = 'activo' THEN e.producto_id END), 0),
		COALESCE(COUNT(DISTINCT CASE WHEN LOWER(COALESCE(e.estado, 'activo')) = 'activo' THEN e.bodega_id END), 0)
	FROM inventario_existencias e
	WHERE e.empresa_id = ?`, empresaID).Scan(
		&resumen.TotalExistencias,
		&resumen.ProductosConExistencia,
		&resumen.BodegasConStock,
	)
	if err != nil {
		return resumen, err
	}

	err = dbConn.QueryRow(`SELECT
		COALESCE(SUM(CASE WHEN COALESCE(e.cantidad, 0) <= 0 THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE
			WHEN COALESCE(e.cantidad, 0) > 0
				AND COALESCE(p.stock_minimo, 0) > 0
				AND COALESCE(e.cantidad, 0) <= COALESCE(p.stock_minimo, 0)
			THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE
			WHEN COALESCE(p.stock_minimo, 0) > COALESCE(e.cantidad, 0)
			THEN COALESCE(p.stock_minimo, 0) - COALESCE(e.cantidad, 0)
			ELSE 0 END), 0)
	FROM inventario_existencias e
	JOIN productos p ON p.id = e.producto_id AND p.empresa_id = e.empresa_id
	WHERE e.empresa_id = ?
		AND LOWER(COALESCE(e.estado, 'activo')) = 'activo'
		AND LOWER(COALESCE(p.estado, 'activo')) = 'activo'
		AND (
			COALESCE(e.cantidad, 0) <= 0
			OR (COALESCE(p.stock_minimo, 0) > 0 AND COALESCE(e.cantidad, 0) <= COALESCE(p.stock_minimo, 0))
		)`, empresaID).Scan(
		&resumen.AlertasSinStock,
		&resumen.AlertasBajoMinimo,
		&resumen.DeficitTotal,
	)
	if err != nil {
		return resumen, err
	}
	resumen.AlertasTotal = resumen.AlertasSinStock + resumen.AlertasBajoMinimo

	query := `SELECT
		COALESCE(COUNT(*), 0),
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(m.tipo, '')) IN ('entrada', 'compra', 'devolucion', 'ajuste_positivo') THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(m.tipo, '')) IN ('salida', 'perdida', 'ajuste_negativo') THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(m.tipo, '')) = 'traslado' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(m.tipo, '')) IN ('ajuste_positivo', 'ajuste_negativo') THEN 1 ELSE 0 END), 0),
		MAX(COALESCE(m.fecha_movimiento, m.fecha_creacion))
	FROM inventario_movimientos m
	WHERE m.empresa_id = ?
		AND LOWER(COALESCE(m.estado, 'activo')) = 'activo'`
	args := []interface{}{empresaID}
	if resumen.PeriodoDesde != "" {
		query += " AND date(COALESCE(m.fecha_movimiento, m.fecha_creacion)) >= date(?)"
		args = append(args, resumen.PeriodoDesde)
	}
	if resumen.PeriodoHasta != "" {
		query += " AND date(COALESCE(m.fecha_movimiento, m.fecha_creacion)) <= date(?)"
		args = append(args, resumen.PeriodoHasta)
	}

	var ultimoMovimiento sql.NullString
	err = dbConn.QueryRow(query, args...).Scan(
		&resumen.MovimientosTotal,
		&resumen.MovimientosEntrada,
		&resumen.MovimientosSalida,
		&resumen.MovimientosTraslado,
		&resumen.MovimientosAjuste,
		&ultimoMovimiento,
	)
	if err != nil {
		return resumen, err
	}
	if ultimoMovimiento.Valid {
		resumen.UltimoMovimiento = ultimoMovimiento.String
	}

	return resumen, nil
}

// GetInventarioTendenciaByEmpresa devuelve una serie diaria de entradas/salidas de inventario.
func GetInventarioTendenciaByEmpresa(dbConn *sql.DB, empresaID, bodegaID int64, desde, hasta string, dias int) ([]InventarioTendenciaDia, error) {
	desde = strings.TrimSpace(desde)
	hasta = strings.TrimSpace(hasta)
	if dias <= 0 {
		dias = 7
	}
	if dias > 120 {
		dias = 120
	}

	now := time.Now()
	if hasta == "" {
		hasta = now.Format("2006-01-02")
	}
	if desde == "" {
		parsedHasta, err := time.Parse("2006-01-02", hasta)
		if err != nil {
			parsedHasta = now
			hasta = parsedHasta.Format("2006-01-02")
		}
		desde = parsedHasta.AddDate(0, 0, -dias+1).Format("2006-01-02")
	}
	if desde > hasta {
		desde, hasta = hasta, desde
	}

	query := `WITH RECURSIVE dias(fecha) AS (
		SELECT date(?)
		UNION ALL
		SELECT date(fecha, '+1 day') FROM dias WHERE fecha < date(?)
	)
	SELECT
		d.fecha,
		COALESCE(SUM(CASE
			WHEN LOWER(COALESCE(m.tipo, '')) IN ('entrada', 'compra', 'devolucion', 'ajuste_positivo') THEN COALESCE(m.cantidad, 0)
			ELSE 0 END), 0),
		COALESCE(SUM(CASE
			WHEN LOWER(COALESCE(m.tipo, '')) IN ('salida', 'perdida', 'ajuste_negativo') THEN COALESCE(m.cantidad, 0)
			ELSE 0 END), 0),
		COALESCE(SUM(CASE
			WHEN LOWER(COALESCE(m.tipo, '')) = 'ajuste_positivo' THEN COALESCE(m.cantidad, 0)
			WHEN LOWER(COALESCE(m.tipo, '')) = 'ajuste_negativo' THEN -COALESCE(m.cantidad, 0)
			ELSE 0 END), 0),
		COALESCE(SUM(CASE
			WHEN LOWER(COALESCE(m.tipo, '')) IN ('traslado', 'cambio_producto') THEN COALESCE(m.cantidad, 0)
			ELSE 0 END), 0),
		COALESCE(COUNT(m.id), 0)
	FROM dias d
	LEFT JOIN inventario_movimientos m
		ON m.empresa_id = ?
		AND LOWER(COALESCE(m.estado, 'activo')) = 'activo'
		AND date(COALESCE(m.fecha_movimiento, m.fecha_creacion)) = d.fecha`
	args := []interface{}{desde, hasta, empresaID}
	if bodegaID > 0 {
		query += " AND (m.bodega_origen_id = ? OR m.bodega_destino_id = ?)"
		args = append(args, bodegaID, bodegaID)
	}
	query += `
	GROUP BY d.fecha
	ORDER BY d.fecha ASC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]InventarioTendenciaDia, 0)
	for rows.Next() {
		var row InventarioTendenciaDia
		if err := rows.Scan(&row.Fecha, &row.Entradas, &row.Salidas, &row.AjusteNeto, &row.Traslados, &row.Eventos); err != nil {
			return nil, err
		}
		row.Neto = row.Entradas - row.Salidas
		out = append(out, row)
	}

	return out, nil
}

// GetInventarioBalanceBodegasByEmpresa devuelve balance de entradas/salidas por bodega en un rango.
func GetInventarioBalanceBodegasByEmpresa(dbConn *sql.DB, empresaID, bodegaID int64, desde, hasta string, dias int) ([]InventarioBalanceBodega, error) {
	desde = strings.TrimSpace(desde)
	hasta = strings.TrimSpace(hasta)
	if dias <= 0 {
		dias = 7
	}
	if dias > 120 {
		dias = 120
	}

	now := time.Now()
	if hasta == "" {
		hasta = now.Format("2006-01-02")
	}
	if desde == "" {
		parsedHasta, err := time.Parse("2006-01-02", hasta)
		if err != nil {
			parsedHasta = now
			hasta = parsedHasta.Format("2006-01-02")
		}
		desde = parsedHasta.AddDate(0, 0, -dias+1).Format("2006-01-02")
	}
	if desde > hasta {
		desde, hasta = hasta, desde
	}

	query := `WITH mov_base AS (
		SELECT m.id, m.bodega_origen_id, m.bodega_destino_id, LOWER(COALESCE(m.tipo, '')) AS tipo, COALESCE(m.cantidad, 0) AS cantidad
		FROM inventario_movimientos m
		WHERE m.empresa_id = ?
			AND LOWER(COALESCE(m.estado, 'activo')) = 'activo'
			AND date(COALESCE(m.fecha_movimiento, m.fecha_creacion)) >= date(?)
			AND date(COALESCE(m.fecha_movimiento, m.fecha_creacion)) <= date(?)`
	args := []interface{}{empresaID, desde, hasta}
	if bodegaID > 0 {
		query += " AND (m.bodega_origen_id = ? OR m.bodega_destino_id = ?)"
		args = append(args, bodegaID, bodegaID)
	}
	query += `
	), mov_expand AS (
		SELECT
			mb.bodega_destino_id AS bodega_id,
			CASE WHEN mb.tipo IN ('entrada', 'compra', 'devolucion', 'ajuste_positivo', 'ajuste_entrada') THEN mb.cantidad ELSE 0 END AS entradas,
			0 AS salidas,
			CASE WHEN mb.tipo IN ('traslado', 'cambio_producto') THEN mb.cantidad ELSE 0 END AS traslados_entrada,
			0 AS traslados_salida,
			CASE
				WHEN mb.tipo IN ('entrada', 'compra', 'devolucion', 'ajuste_positivo', 'ajuste_entrada', 'traslado', 'cambio_producto') THEN mb.cantidad
				ELSE 0
			END AS neto,
			CASE
				WHEN mb.tipo IN ('entrada', 'compra', 'devolucion', 'ajuste_positivo', 'ajuste_entrada', 'traslado', 'cambio_producto') THEN 1
				ELSE 0
			END AS eventos
		FROM mov_base mb
		WHERE mb.bodega_destino_id IS NOT NULL
		UNION ALL
		SELECT
			mb.bodega_origen_id AS bodega_id,
			0 AS entradas,
			CASE WHEN mb.tipo IN ('salida', 'perdida', 'ajuste_negativo', 'ajuste_salida') THEN mb.cantidad ELSE 0 END AS salidas,
			0 AS traslados_entrada,
			CASE WHEN mb.tipo IN ('traslado', 'cambio_producto') THEN mb.cantidad ELSE 0 END AS traslados_salida,
			CASE
				WHEN mb.tipo IN ('salida', 'perdida', 'ajuste_negativo', 'ajuste_salida', 'traslado', 'cambio_producto') THEN -mb.cantidad
				ELSE 0
			END AS neto,
			CASE
				WHEN mb.tipo IN ('salida', 'perdida', 'ajuste_negativo', 'ajuste_salida', 'traslado', 'cambio_producto') THEN 1
				ELSE 0
			END AS eventos
		FROM mov_base mb
		WHERE mb.bodega_origen_id IS NOT NULL
	)
	SELECT
		b.id,
		b.nombre,
		COALESCE(SUM(me.entradas), 0),
		COALESCE(SUM(me.salidas), 0),
		COALESCE(SUM(me.traslados_entrada), 0),
		COALESCE(SUM(me.traslados_salida), 0),
		COALESCE(SUM(me.eventos), 0),
		COALESCE(SUM(me.neto), 0)
	FROM bodegas b
	LEFT JOIN mov_expand me ON me.bodega_id = b.id
	WHERE b.empresa_id = ?
		AND LOWER(COALESCE(b.estado, 'activo')) = 'activo'`
	args = append(args, empresaID)
	if bodegaID > 0 {
		query += " AND b.id = ?"
		args = append(args, bodegaID)
	}
	query += `
	GROUP BY b.id, b.nombre
	ORDER BY ABS(COALESCE(SUM(me.neto), 0)) DESC, b.nombre ASC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]InventarioBalanceBodega, 0)
	for rows.Next() {
		var row InventarioBalanceBodega
		if err := rows.Scan(
			&row.BodegaID,
			&row.BodegaNombre,
			&row.Entradas,
			&row.Salidas,
			&row.TrasladosEntrada,
			&row.TrasladosSalida,
			&row.Eventos,
			&row.Neto,
		); err != nil {
			return nil, err
		}
		row.TrasladoNeto = row.TrasladosEntrada - row.TrasladosSalida
		out = append(out, row)
	}

	return out, nil
}

// GetInventarioProyeccionQuiebreByEmpresa estima cobertura y riesgo de quiebre por producto/bodega.
func GetInventarioProyeccionQuiebreByEmpresa(dbConn *sql.DB, empresaID, bodegaID int64, diasVentana int, limit int, offset int) ([]InventarioProyeccionQuiebre, error) {
	if diasVentana <= 0 {
		diasVentana = 30
	}
	if diasVentana > 180 {
		diasVentana = 180
	}
	if limit <= 0 || limit > 500 {
		limit = 80
	}
	if offset < 0 {
		offset = 0
	}

	periodExpr := fmt.Sprintf("-%d day", diasVentana-1)
	query := `SELECT
		e.empresa_id,
		e.producto_id,
		p.nombre,
		e.bodega_id,
		b.nombre,
		COALESCE(e.cantidad, 0),
		COALESCE(p.stock_minimo, 0),
		COALESCE(p.stock_maximo, 0),
		COALESCE((
			SELECT SUM(COALESCE(m.cantidad, 0))
			FROM inventario_movimientos m
			WHERE m.empresa_id = e.empresa_id
				AND m.producto_id = e.producto_id
				AND m.bodega_origen_id = e.bodega_id
				AND LOWER(COALESCE(m.estado, 'activo')) = 'activo'
				AND LOWER(COALESCE(m.tipo, '')) IN ('salida', 'perdida', 'ajuste_negativo', 'ajuste_salida', 'traslado', 'cambio_producto')
				AND date(COALESCE(m.fecha_movimiento, m.fecha_creacion)) >= date('now', ?)
		), 0) / CAST(? AS REAL) AS salida_promedio
	FROM inventario_existencias e
	JOIN productos p ON p.id = e.producto_id AND p.empresa_id = e.empresa_id
	JOIN bodegas b ON b.id = e.bodega_id AND b.empresa_id = e.empresa_id
	WHERE e.empresa_id = ?
		AND LOWER(COALESCE(e.estado, 'activo')) = 'activo'
		AND LOWER(COALESCE(p.estado, 'activo')) = 'activo'
		AND LOWER(COALESCE(b.estado, 'activo')) = 'activo'`
	args := []interface{}{periodExpr, diasVentana, empresaID}
	if bodegaID > 0 {
		query += " AND e.bodega_id = ?"
		args = append(args, bodegaID)
	}
	query += " ORDER BY p.nombre ASC, b.nombre ASC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]InventarioProyeccionQuiebre, 0)
	for rows.Next() {
		var row InventarioProyeccionQuiebre
		if err := rows.Scan(
			&row.EmpresaID,
			&row.ProductoID,
			&row.ProductoNombre,
			&row.BodegaID,
			&row.BodegaNombre,
			&row.StockActual,
			&row.StockMinimo,
			&row.StockMaximo,
			&row.SalidaPromedioDiaria,
		); err != nil {
			return nil, err
		}
		row.DiasVentana = diasVentana
		if row.StockMinimo > row.StockActual {
			row.Deficit = row.StockMinimo - row.StockActual
		}

		if row.SalidaPromedioDiaria > 0 {
			row.DiasCobertura = row.StockActual / row.SalidaPromedioDiaria
		} else if row.StockActual > 0 {
			row.DiasCobertura = -1
		}

		targetStock := row.StockMaximo
		if targetStock <= 0 {
			if row.StockMinimo > 0 {
				targetStock = row.StockMinimo * 2
			} else {
				targetStock = row.StockActual
			}
		}
		if targetStock > row.StockActual {
			row.SugeridoReposicion = targetStock - row.StockActual
		}

		switch {
		case row.StockActual <= 0:
			row.EstadoProyeccion = "quiebre_inminente"
		case row.SalidaPromedioDiaria > 0 && row.DiasCobertura > 0 && row.DiasCobertura <= 3:
			row.EstadoProyeccion = "quiebre_inminente"
		case row.StockMinimo > 0 && row.StockActual <= row.StockMinimo:
			row.EstadoProyeccion = "bajo_minimo"
		case row.SalidaPromedioDiaria <= 0:
			row.EstadoProyeccion = "sin_consumo_reciente"
		case row.DiasCobertura <= 7:
			row.EstadoProyeccion = "riesgo_alto"
		case row.DiasCobertura <= 14:
			row.EstadoProyeccion = "riesgo_medio"
		default:
			row.EstadoProyeccion = "estable"
		}

		out = append(out, row)
	}

	severity := func(estado string) int {
		switch estado {
		case "quiebre_inminente":
			return 0
		case "bajo_minimo":
			return 1
		case "riesgo_alto":
			return 2
		case "riesgo_medio":
			return 3
		case "sin_consumo_reciente":
			return 4
		default:
			return 5
		}
	}
	normalizeDias := func(d float64) float64 {
		if d < 0 {
			return 999999
		}
		return d
	}

	sort.Slice(out, func(i, j int) bool {
		si := severity(out[i].EstadoProyeccion)
		sj := severity(out[j].EstadoProyeccion)
		if si != sj {
			return si < sj
		}
		di := normalizeDias(out[i].DiasCobertura)
		dj := normalizeDias(out[j].DiasCobertura)
		if di != dj {
			return di < dj
		}
		if out[i].Deficit != out[j].Deficit {
			return out[i].Deficit > out[j].Deficit
		}
		if out[i].ProductoNombre != out[j].ProductoNombre {
			return out[i].ProductoNombre < out[j].ProductoNombre
		}
		return out[i].BodegaNombre < out[j].BodegaNombre
	})

	return out, nil
}

// GetInventarioPlanReposicionByEmpresa consolida una propuesta preventiva de compra por proveedor.
func GetInventarioPlanReposicionByEmpresa(dbConn *sql.DB, empresaID, bodegaID int64, diasVentana int, soloRiesgo bool, limit int, offset int) ([]InventarioPlanReposicionItem, error) {
	if diasVentana <= 0 {
		diasVentana = 30
	}
	if diasVentana > 180 {
		diasVentana = 180
	}
	if limit <= 0 || limit > 500 {
		limit = 80
	}
	if offset < 0 {
		offset = 0
	}

	baseRows, err := GetInventarioProyeccionQuiebreByEmpresa(dbConn, empresaID, bodegaID, diasVentana, 500, 0)
	if err != nil {
		return nil, err
	}
	if len(baseRows) == 0 {
		return []InventarioPlanReposicionItem{}, nil
	}

	productIDSet := make(map[int64]struct{})
	for _, row := range baseRows {
		if row.ProductoID > 0 {
			productIDSet[row.ProductoID] = struct{}{}
		}
	}
	if len(productIDSet) == 0 {
		return []InventarioPlanReposicionItem{}, nil
	}

	productIDs := make([]int64, 0, len(productIDSet))
	for id := range productIDSet {
		productIDs = append(productIDs, id)
	}
	sort.Slice(productIDs, func(i, j int) bool {
		return productIDs[i] < productIDs[j]
	})

	placeholders := make([]string, 0, len(productIDs))
	args := make([]interface{}, 0, 1+len(productIDs))
	args = append(args, empresaID)
	for _, id := range productIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}

	metaQuery := `SELECT
		p.id,
		COALESCE(p.proveedor_principal_id, 0),
		COALESCE(pr.nombre, ''),
		COALESCE(p.costo, 0)
	FROM productos p
	LEFT JOIN proveedores pr ON pr.id = p.proveedor_principal_id AND pr.empresa_id = p.empresa_id
	WHERE p.empresa_id = ?
		AND p.id IN (` + strings.Join(placeholders, ",") + `)`

	metaRows, err := dbConn.Query(metaQuery, args...)
	if err != nil {
		return nil, err
	}
	defer metaRows.Close()

	type productMeta struct {
		ProveedorID     int64
		ProveedorNombre string
		Costo           float64
	}
	metaByProduct := make(map[int64]productMeta)
	for metaRows.Next() {
		var productID int64
		var meta productMeta
		if err := metaRows.Scan(&productID, &meta.ProveedorID, &meta.ProveedorNombre, &meta.Costo); err != nil {
			return nil, err
		}
		metaByProduct[productID] = meta
	}

	isRiesgo := func(estado string) bool {
		switch estado {
		case "quiebre_inminente", "bajo_minimo", "riesgo_alto", "riesgo_medio":
			return true
		default:
			return false
		}
	}
	severity := func(estado string) int {
		switch estado {
		case "quiebre_inminente":
			return 0
		case "bajo_minimo":
			return 1
		case "riesgo_alto":
			return 2
		case "riesgo_medio":
			return 3
		case "sin_consumo_reciente":
			return 4
		default:
			return 5
		}
	}

	out := make([]InventarioPlanReposicionItem, 0, len(baseRows))
	for _, row := range baseRows {
		if row.SugeridoReposicion <= 0 {
			continue
		}
		estado := strings.TrimSpace(row.EstadoProyeccion)
		if soloRiesgo && !isRiesgo(estado) {
			continue
		}

		meta := metaByProduct[row.ProductoID]
		proveedorNombre := strings.TrimSpace(meta.ProveedorNombre)
		if proveedorNombre == "" {
			if meta.ProveedorID > 0 {
				proveedorNombre = fmt.Sprintf("Proveedor #%d", meta.ProveedorID)
			} else {
				proveedorNombre = "Sin proveedor principal"
			}
		}

		item := InventarioPlanReposicionItem{
			EmpresaID:          row.EmpresaID,
			ProveedorID:        meta.ProveedorID,
			ProveedorNombre:    proveedorNombre,
			ProductoID:         row.ProductoID,
			ProductoNombre:     row.ProductoNombre,
			BodegaID:           row.BodegaID,
			BodegaNombre:       row.BodegaNombre,
			EstadoProyeccion:   estado,
			DiasCobertura:      row.DiasCobertura,
			SugeridoReposicion: row.SugeridoReposicion,
			CostoUnitarioRef:   meta.Costo,
			DiasVentana:        diasVentana,
		}
		item.CostoEstimado = item.SugeridoReposicion * item.CostoUnitarioRef
		out = append(out, item)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].ProveedorNombre != out[j].ProveedorNombre {
			return out[i].ProveedorNombre < out[j].ProveedorNombre
		}
		si := severity(out[i].EstadoProyeccion)
		sj := severity(out[j].EstadoProyeccion)
		if si != sj {
			return si < sj
		}
		if out[i].CostoEstimado != out[j].CostoEstimado {
			return out[i].CostoEstimado > out[j].CostoEstimado
		}
		if out[i].ProductoNombre != out[j].ProductoNombre {
			return out[i].ProductoNombre < out[j].ProductoNombre
		}
		return out[i].BodegaNombre < out[j].BodegaNombre
	})

	if offset >= len(out) {
		return []InventarioPlanReposicionItem{}, nil
	}
	end := offset + limit
	if end > len(out) {
		end = len(out)
	}

	return out[offset:end], nil
}

// GetInventarioPlanReposicionResumenByEmpresa consolida el plan preventivo agrupado por proveedor.
func GetInventarioPlanReposicionResumenByEmpresa(dbConn *sql.DB, empresaID, bodegaID int64, diasVentana int, soloRiesgo bool, limit int, offset int) ([]InventarioPlanReposicionProveedorResumen, error) {
	if limit <= 0 || limit > 500 {
		limit = 80
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := GetInventarioPlanReposicionByEmpresa(dbConn, empresaID, bodegaID, diasVentana, soloRiesgo, 500, 0)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []InventarioPlanReposicionProveedorResumen{}, nil
	}

	type providerAgg struct {
		summary      InventarioPlanReposicionProveedorResumen
		productIDSet map[int64]struct{}
	}
	aggByProvider := make(map[string]*providerAgg)

	for _, row := range rows {
		providerName := strings.TrimSpace(row.ProveedorNombre)
		if providerName == "" {
			if row.ProveedorID > 0 {
				providerName = fmt.Sprintf("Proveedor #%d", row.ProveedorID)
			} else {
				providerName = "Sin proveedor principal"
			}
		}
		key := fmt.Sprintf("%d|%s", row.ProveedorID, providerName)

		agg := aggByProvider[key]
		if agg == nil {
			agg = &providerAgg{
				summary: InventarioPlanReposicionProveedorResumen{
					EmpresaID:       row.EmpresaID,
					ProveedorID:     row.ProveedorID,
					ProveedorNombre: providerName,
					DiasVentana:     row.DiasVentana,
				},
				productIDSet: make(map[int64]struct{}),
			}
			aggByProvider[key] = agg
		}

		agg.summary.Items++
		agg.summary.CantidadTotal += row.SugeridoReposicion
		agg.summary.CostoTotal += row.CostoEstimado
		agg.productIDSet[row.ProductoID] = struct{}{}

		switch row.EstadoProyeccion {
		case "quiebre_inminente":
			agg.summary.QuiebreInminente++
		case "bajo_minimo":
			agg.summary.BajoMinimo++
		case "riesgo_alto":
			agg.summary.RiesgoAlto++
		case "riesgo_medio":
			agg.summary.RiesgoMedio++
		}
	}

	out := make([]InventarioPlanReposicionProveedorResumen, 0, len(aggByProvider))
	for _, agg := range aggByProvider {
		agg.summary.ProductosUnicos = int64(len(agg.productIDSet))
		out = append(out, agg.summary)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].CostoTotal != out[j].CostoTotal {
			return out[i].CostoTotal > out[j].CostoTotal
		}
		if out[i].QuiebreInminente != out[j].QuiebreInminente {
			return out[i].QuiebreInminente > out[j].QuiebreInminente
		}
		return out[i].ProveedorNombre < out[j].ProveedorNombre
	})

	if offset >= len(out) {
		return []InventarioPlanReposicionProveedorResumen{}, nil
	}
	end := offset + limit
	if end > len(out) {
		end = len(out)
	}

	return out[offset:end], nil
}

// GetInventarioPlanReposicionBorradorByEmpresa construye un borrador de orden de compra para un proveedor.
func GetInventarioPlanReposicionBorradorByEmpresa(dbConn *sql.DB, empresaID, proveedorID, bodegaID int64, diasVentana int, soloRiesgo bool) (InventarioPlanReposicionBorradorCompra, error) {
	if empresaID <= 0 {
		return InventarioPlanReposicionBorradorCompra{}, fmt.Errorf("empresa_id invalido")
	}
	if proveedorID <= 0 {
		return InventarioPlanReposicionBorradorCompra{}, fmt.Errorf("proveedor_id invalido")
	}
	if diasVentana <= 0 {
		diasVentana = 30
	}
	if diasVentana > 180 {
		diasVentana = 180
	}

	nowLocal := time.Now()
	out := InventarioPlanReposicionBorradorCompra{
		EmpresaID:       empresaID,
		ProveedorID:     proveedorID,
		ProveedorNombre: fmt.Sprintf("Proveedor #%d", proveedorID),
		FechaDocumento:  nowLocal.Format("2006-01-02"),
		CodigoBorrador:  fmt.Sprintf("BORR-OC-%d-%s", proveedorID, nowLocal.Format("20060102150405")),
		DiasVentana:     diasVentana,
		Items:           []InventarioPlanReposicionBorradorItem{},
	}

	var providerName string
	if err := dbConn.QueryRow("SELECT COALESCE(nombre, '') FROM proveedores WHERE empresa_id = ? AND id = ?", empresaID, proveedorID).Scan(&providerName); err == nil {
		providerName = strings.TrimSpace(providerName)
		if providerName != "" {
			out.ProveedorNombre = providerName
		}
	} else if err != nil && err != sql.ErrNoRows {
		return InventarioPlanReposicionBorradorCompra{}, err
	}

	rows, err := GetInventarioPlanReposicionByEmpresa(dbConn, empresaID, bodegaID, diasVentana, soloRiesgo, 500, 0)
	if err != nil {
		return InventarioPlanReposicionBorradorCompra{}, err
	}
	if len(rows) == 0 {
		return out, nil
	}

	severity := func(estado string) int {
		switch estado {
		case "quiebre_inminente":
			return 0
		case "bajo_minimo":
			return 1
		case "riesgo_alto":
			return 2
		case "riesgo_medio":
			return 3
		default:
			return 4
		}
	}

	productIDSet := make(map[int64]struct{})
	for _, row := range rows {
		if row.ProveedorID != proveedorID {
			continue
		}
		if row.SugeridoReposicion <= 0 {
			continue
		}
		if strings.TrimSpace(row.ProveedorNombre) != "" {
			out.ProveedorNombre = strings.TrimSpace(row.ProveedorNombre)
		}

		item := InventarioPlanReposicionBorradorItem{
			EmpresaID:        row.EmpresaID,
			ProveedorID:      row.ProveedorID,
			ProveedorNombre:  out.ProveedorNombre,
			ProductoID:       row.ProductoID,
			ProductoNombre:   row.ProductoNombre,
			BodegaID:         row.BodegaID,
			BodegaNombre:     row.BodegaNombre,
			EstadoProyeccion: row.EstadoProyeccion,
			DiasCobertura:    row.DiasCobertura,
			CantidadSugerida: row.SugeridoReposicion,
			CostoUnitarioRef: row.CostoUnitarioRef,
			CostoEstimado:    row.CostoEstimado,
		}

		out.Items = append(out.Items, item)
		out.TotalItems++
		out.CantidadTotal += item.CantidadSugerida
		out.CostoTotal += item.CostoEstimado
		productIDSet[item.ProductoID] = struct{}{}

		switch item.EstadoProyeccion {
		case "quiebre_inminente":
			out.QuiebreInminente++
		case "bajo_minimo":
			out.BajoMinimo++
		case "riesgo_alto":
			out.RiesgoAlto++
		case "riesgo_medio":
			out.RiesgoMedio++
		}
	}

	if len(out.Items) == 0 {
		return out, nil
	}

	sort.Slice(out.Items, func(i, j int) bool {
		si := severity(out.Items[i].EstadoProyeccion)
		sj := severity(out.Items[j].EstadoProyeccion)
		if si != sj {
			return si < sj
		}
		if out.Items[i].CostoEstimado != out.Items[j].CostoEstimado {
			return out.Items[i].CostoEstimado > out.Items[j].CostoEstimado
		}
		if out.Items[i].ProductoNombre != out.Items[j].ProductoNombre {
			return out.Items[i].ProductoNombre < out.Items[j].ProductoNombre
		}
		return out.Items[i].BodegaNombre < out.Items[j].BodegaNombre
	})

	out.ProductosUnicos = int64(len(productIDSet))
	return out, nil
}

// EmitirOrdenCompraDesdePlanReposicionBorrador emite una orden de compra tomando como base el borrador preventivo.
func EmitirOrdenCompraDesdePlanReposicionBorrador(dbConn *sql.DB, empresaID, proveedorID, bodegaID int64, diasVentana int, soloRiesgo bool, documentoCodigo, periodoContable, moneda, usuario, observaciones string) (InventarioPlanReposicionOrdenEmitida, error) {
	borrador, err := GetInventarioPlanReposicionBorradorByEmpresa(dbConn, empresaID, proveedorID, bodegaID, diasVentana, soloRiesgo)
	if err != nil {
		return InventarioPlanReposicionOrdenEmitida{}, err
	}
	if len(borrador.Items) == 0 {
		return InventarioPlanReposicionOrdenEmitida{}, fmt.Errorf("no hay items sugeridos para emitir orden")
	}

	codigo := strings.ToUpper(strings.TrimSpace(documentoCodigo))
	if codigo == "" {
		codigo = strings.ToUpper(strings.TrimSpace(borrador.CodigoBorrador))
		if strings.HasPrefix(codigo, "BORR-") {
			codigo = strings.TrimPrefix(codigo, "BORR-")
		}
	}
	if codigo == "" {
		codigo = fmt.Sprintf("OC-%d-%s", proveedorID, time.Now().Format("20060102150405"))
	}

	estadoAnterior := "borrador"
	estadoNuevo := "emitida"
	evento := "orden_compra_emitida"

	docPersistido, err := UpsertEmpresaDocumentoCompra(dbConn, EmpresaDocumentoCompra{
		EmpresaID:            empresaID,
		ProveedorID:          proveedorID,
		TipoDocumento:        "orden_compra",
		DocumentoCodigo:      codigo,
		EstadoDocumento:      estadoNuevo,
		EstadoAnterior:       estadoAnterior,
		EventoUltimo:         evento,
		PeriodoContable:      periodoContable,
		MontoTotal:           borrador.CostoTotal,
		Moneda:               moneda,
		FechaDocumento:       borrador.FechaDocumento,
		EntidadRelacionadaID: proveedorID,
		UsuarioCreador:       strings.TrimSpace(usuario),
		Estado:               "activo",
		Observaciones:        strings.TrimSpace(observaciones),
	})
	if err != nil {
		return InventarioPlanReposicionOrdenEmitida{}, err
	}

	out := InventarioPlanReposicionOrdenEmitida{
		EmpresaID:       empresaID,
		ProveedorID:     proveedorID,
		ProveedorNombre: borrador.ProveedorNombre,
		DocumentoCodigo: docPersistido.DocumentoCodigo,
		Accion:          "emitir_orden",
		EstadoAnterior:  estadoAnterior,
		EstadoNuevo:     estadoNuevo,
		Evento:          evento,
		PeriodoContable: docPersistido.PeriodoContable,
		Moneda:          docPersistido.Moneda,
		FechaDocumento:  docPersistido.FechaDocumento,
		TotalItems:      borrador.TotalItems,
		ProductosUnicos: borrador.ProductosUnicos,
		CantidadTotal:   borrador.CantidadTotal,
		CostoTotal:      borrador.CostoTotal,
		EntidadID:       docPersistido.ID,
	}

	return out, nil
}

// ActualizarEstadoOrdenCompraDesdeReposicion actualiza la OC emitida en su ciclo documental (recepcionar/contabilizar).
func ActualizarEstadoOrdenCompraDesdeReposicion(dbConn *sql.DB, empresaID, proveedorID int64, documentoCodigo, accion, estadoActual, periodoContable, observaciones, usuario string) (InventarioPlanReposicionOrdenEstadoActualizado, error) {
	if empresaID <= 0 {
		return InventarioPlanReposicionOrdenEstadoActualizado{}, fmt.Errorf("empresa_id invalido")
	}
	if proveedorID <= 0 {
		return InventarioPlanReposicionOrdenEstadoActualizado{}, fmt.Errorf("proveedor_id invalido")
	}
	codigo := strings.ToUpper(strings.TrimSpace(documentoCodigo))
	if codigo == "" {
		return InventarioPlanReposicionOrdenEstadoActualizado{}, fmt.Errorf("documento_codigo es obligatorio")
	}

	docActual, err := GetEmpresaDocumentoCompraByCodigo(dbConn, empresaID, "orden_compra", codigo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return InventarioPlanReposicionOrdenEstadoActualizado{}, fmt.Errorf("documento no encontrado para codigo %s", codigo)
		}
		return InventarioPlanReposicionOrdenEstadoActualizado{}, err
	}

	normalize := func(raw string) string {
		v := strings.ToLower(strings.TrimSpace(raw))
		v = strings.ReplaceAll(v, "-", "_")
		v = strings.ReplaceAll(v, " ", "_")
		return v
	}

	action := normalize(accion)
	if action == "" {
		return InventarioPlanReposicionOrdenEstadoActualizado{}, fmt.Errorf("accion es obligatoria")
	}

	resolvedEstadoActual := normalize(estadoActual)
	if resolvedEstadoActual == "" {
		resolvedEstadoActual = normalize(docActual.EstadoDocumento)
	}
	if resolvedEstadoActual == "" {
		resolvedEstadoActual = "emitida"
	}

	transition := struct {
		accionCanonica string
		evento         string
		estadoNuevo    string
		allowedPrev    map[string]struct{}
	}{
		accionCanonica: "",
		evento:         "",
		estadoNuevo:    "",
		allowedPrev:    map[string]struct{}{},
	}

	switch action {
	case "recepcionar", "recepcionar_compra":
		transition.accionCanonica = "recepcionar_compra"
		transition.evento = "compra_recepcionada"
		transition.estadoNuevo = "recepcionada"
		transition.allowedPrev = map[string]struct{}{"emitida": {}, "recepcion_parcial": {}}
	case "contabilizar", "contabilizar_compra":
		transition.accionCanonica = "contabilizar_compra"
		transition.evento = "compra_contabilizada"
		transition.estadoNuevo = "contabilizada"
		transition.allowedPrev = map[string]struct{}{"recepcionada": {}}
	default:
		return InventarioPlanReposicionOrdenEstadoActualizado{}, fmt.Errorf("accion no soportada para compras reposicion: %s", strings.TrimSpace(accion))
	}

	if _, ok := transition.allowedPrev[resolvedEstadoActual]; !ok {
		keys := make([]string, 0, len(transition.allowedPrev))
		for k := range transition.allowedPrev {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		return InventarioPlanReposicionOrdenEstadoActualizado{}, fmt.Errorf("transicion invalida para compras reposicion: accion=%s requiere estado_actual en [%s], recibido=%s", transition.accionCanonica, strings.Join(keys, ","), resolvedEstadoActual)
	}

	docPersistido, err := UpsertEmpresaDocumentoCompra(dbConn, EmpresaDocumentoCompra{
		EmpresaID:            empresaID,
		ProveedorID:          proveedorID,
		TipoDocumento:        "orden_compra",
		DocumentoCodigo:      codigo,
		EstadoDocumento:      transition.estadoNuevo,
		EstadoAnterior:       resolvedEstadoActual,
		EventoUltimo:         transition.evento,
		PeriodoContable:      firstNonBlank(periodoContable, docActual.PeriodoContable),
		MontoTotal:           docActual.MontoTotal,
		Moneda:               firstNonBlank(docActual.Moneda, "COP"),
		FechaDocumento:       docActual.FechaDocumento,
		EntidadRelacionadaID: proveedorID,
		UsuarioCreador:       strings.TrimSpace(usuario),
		Estado:               "activo",
		Observaciones:        strings.TrimSpace(observaciones),
	})
	if err != nil {
		return InventarioPlanReposicionOrdenEstadoActualizado{}, err
	}

	out := InventarioPlanReposicionOrdenEstadoActualizado{
		EmpresaID:       empresaID,
		ProveedorID:     proveedorID,
		DocumentoCodigo: docPersistido.DocumentoCodigo,
		Accion:          transition.accionCanonica,
		EstadoAnterior:  resolvedEstadoActual,
		EstadoNuevo:     transition.estadoNuevo,
		Evento:          transition.evento,
		PeriodoContable: docPersistido.PeriodoContable,
		Moneda:          docPersistido.Moneda,
		MontoTotal:      docPersistido.MontoTotal,
		EntidadID:       docPersistido.ID,
	}

	return out, nil
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// TransferirProductoEntreBodegas mueve unidades entre bodegas de una empresa y registra movimiento.
func TransferirProductoEntreBodegas(dbConn *sql.DB, empresaID, productoID, bodegaOrigenID, bodegaDestinoID int64, cantidad float64, referencia, usuario, observaciones string) error {
	if cantidad <= 0 {
		return fmt.Errorf("cantidad debe ser mayor a 0")
	}
	if bodegaOrigenID == bodegaDestinoID {
		return fmt.Errorf("la bodega origen y destino no pueden ser iguales")
	}

	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := validateBodegaEmpresaTx(tx, empresaID, bodegaOrigenID); err != nil {
		return err
	}
	if err := validateBodegaEmpresaTx(tx, empresaID, bodegaDestinoID); err != nil {
		return err
	}
	if err := validateProductoEmpresaTx(tx, empresaID, productoID); err != nil {
		return err
	}

	politicaCosto, err := getInventarioPoliticaCostoTx(tx, empresaID)
	if err != nil {
		return err
	}

	var costoUnitario float64
	if err := tx.QueryRow("SELECT costo FROM productos WHERE empresa_id = ? AND id = ?", empresaID, productoID).Scan(&costoUnitario); err != nil {
		return err
	}

	var stockOrigen float64
	err = tx.QueryRow("SELECT cantidad FROM inventario_existencias WHERE empresa_id = ? AND producto_id = ? AND bodega_id = ?", empresaID, productoID, bodegaOrigenID).Scan(&stockOrigen)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrStockInsuficiente
		}
		return err
	}
	if stockOrigen < cantidad {
		return ErrStockInsuficiente
	}

	costoTransferido, err := transferirCostoLotesEntreBodegasTx(tx, empresaID, productoID, bodegaOrigenID, bodegaDestinoID, cantidad, strings.TrimSpace(referencia), strings.TrimSpace(usuario))
	if err != nil {
		if errors.Is(err, ErrStockInsuficiente) {
			return ErrStockInsuficiente
		}
		return err
	}
	if costoTransferido > 0 {
		costoUnitario = costoTransferido
	}

	if _, err := tx.Exec(`UPDATE inventario_existencias
		SET cantidad = cantidad - ?, fecha_actualizacion = datetime('now','localtime')
		WHERE empresa_id = ? AND producto_id = ? AND bodega_id = ?`, cantidad, empresaID, productoID, bodegaOrigenID); err != nil {
		return err
	}

	if err := upsertExistenciaTx(tx, empresaID, productoID, bodegaDestinoID, cantidad, usuario, "traslado entre bodegas"); err != nil {
		return err
	}

	if err := insertMovimientoTx(tx, InventarioMovimiento{
		EmpresaID:       empresaID,
		ProductoID:      productoID,
		BodegaOrigenID:  bodegaOrigenID,
		BodegaDestinoID: bodegaDestinoID,
		Tipo:            "traslado",
		Cantidad:        cantidad,
		CostoUnitario:   costoUnitario,
		Referencia:      strings.TrimSpace(referencia),
		UsuarioCreador:  strings.TrimSpace(usuario),
		Estado:          "activo",
		Observaciones:   strings.TrimSpace(observaciones),
	}); err != nil {
		return err
	}
	if politicaCosto == inventarioPoliticaCostoPromedio {
		if err := recalcularCostoPromedioProductoTx(tx, empresaID, productoID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetMovimientosByEmpresa devuelve historial de movimientos de inventario (kardex) con filtros operativos.
func GetMovimientosByEmpresa(dbConn *sql.DB, empresaID, productoID, bodegaID int64, tipo, desde, hasta string, limit int, offset int) ([]InventarioMovimiento, error) {
	if limit <= 0 || limit > 1000 {
		limit = 300
	}
	if offset < 0 {
		offset = 0
	}

	query := `SELECT m.id, m.empresa_id, m.producto_id, p.nombre,
		m.bodega_origen_id, bo.nombre,
		m.bodega_destino_id, bd.nombre,
		m.tipo, m.cantidad, m.costo_unitario, m.referencia, m.fecha_movimiento,
		m.fecha_creacion, m.fecha_actualizacion, m.usuario_creador, m.estado, m.observaciones
	FROM inventario_movimientos m
	JOIN productos p ON p.id = m.producto_id AND p.empresa_id = m.empresa_id
	LEFT JOIN bodegas bo ON bo.id = m.bodega_origen_id AND bo.empresa_id = m.empresa_id
	LEFT JOIN bodegas bd ON bd.id = m.bodega_destino_id AND bd.empresa_id = m.empresa_id
	WHERE m.empresa_id = ?`
	args := []interface{}{empresaID}
	if productoID > 0 {
		query += " AND m.producto_id = ?"
		args = append(args, productoID)
	}
	if bodegaID > 0 {
		query += " AND (m.bodega_origen_id = ? OR m.bodega_destino_id = ?)"
		args = append(args, bodegaID, bodegaID)
	}
	if strings.TrimSpace(tipo) != "" {
		query += " AND LOWER(COALESCE(m.tipo, '')) = LOWER(?)"
		args = append(args, strings.TrimSpace(tipo))
	}
	if strings.TrimSpace(desde) != "" {
		query += " AND date(COALESCE(m.fecha_movimiento, m.fecha_creacion)) >= date(?)"
		args = append(args, strings.TrimSpace(desde))
	}
	if strings.TrimSpace(hasta) != "" {
		query += " AND date(COALESCE(m.fecha_movimiento, m.fecha_creacion)) <= date(?)"
		args = append(args, strings.TrimSpace(hasta))
	}

	query += " ORDER BY datetime(COALESCE(m.fecha_movimiento, m.fecha_creacion)) DESC, m.id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]InventarioMovimiento, 0)
	for rows.Next() {
		var m InventarioMovimiento
		var bodegaOrigenID, bodegaDestinoID sql.NullInt64
		var bodegaOrigenNombre, bodegaDestinoNombre sql.NullString
		var referencia, fechaMov, fechaCre, fechaAct, usuario, estado, obs sql.NullString
		if err := rows.Scan(
			&m.ID, &m.EmpresaID, &m.ProductoID, &m.ProductoNombre,
			&bodegaOrigenID, &bodegaOrigenNombre,
			&bodegaDestinoID, &bodegaDestinoNombre,
			&m.Tipo, &m.Cantidad, &m.CostoUnitario, &referencia, &fechaMov,
			&fechaCre, &fechaAct, &usuario, &estado, &obs,
		); err != nil {
			return nil, err
		}
		if bodegaOrigenID.Valid {
			m.BodegaOrigenID = bodegaOrigenID.Int64
		}
		if bodegaDestinoID.Valid {
			m.BodegaDestinoID = bodegaDestinoID.Int64
		}
		if bodegaOrigenNombre.Valid {
			m.BodegaOrigenNombre = bodegaOrigenNombre.String
		}
		if bodegaDestinoNombre.Valid {
			m.BodegaDestinoNombre = bodegaDestinoNombre.String
		}
		if referencia.Valid {
			m.Referencia = referencia.String
		}
		if fechaMov.Valid {
			m.FechaMovimiento = fechaMov.String
		}
		if fechaCre.Valid {
			m.FechaCreacion = fechaCre.String
		}
		if fechaAct.Valid {
			m.FechaActualizacion = fechaAct.String
		}
		if usuario.Valid {
			m.UsuarioCreador = usuario.String
		}
		if estado.Valid {
			m.Estado = estado.String
		}
		if obs.Valid {
			m.Observaciones = obs.String
		}
		out = append(out, m)
	}

	return out, nil
}

// CreateProveedor inserta un proveedor para la empresa.
func CreateProveedor(dbConn *sql.DB, p Proveedor) (int64, error) {
	if err := validateProveedorCondiciones(p); err != nil {
		return 0, err
	}

	res, err := dbConn.Exec(`INSERT INTO proveedores (
		empresa_id, codigo, nombre, documento, contacto, telefono, email, direccion, catalogo_referencia,
		precio_base_referencial, descuento_porcentaje, plazo_pago_dias, condicion_entrega,
		usuario_creador, estado, observaciones, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, COALESCE(NULLIF(?, ''), 'activo'), ?, datetime('now','localtime'), datetime('now','localtime'))`,
		p.EmpresaID,
		strings.TrimSpace(p.Codigo),
		strings.TrimSpace(p.Nombre),
		strings.TrimSpace(p.Documento),
		strings.TrimSpace(p.Contacto),
		strings.TrimSpace(p.Telefono),
		strings.TrimSpace(p.Email),
		strings.TrimSpace(p.Direccion),
		strings.TrimSpace(p.CatalogoReferencia),
		p.PrecioBaseReferencial,
		p.DescuentoPorcentaje,
		p.PlazoPagoDias,
		strings.TrimSpace(p.CondicionEntrega),
		strings.TrimSpace(p.UsuarioCreador),
		strings.TrimSpace(p.Estado),
		strings.TrimSpace(p.Observaciones))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetProveedoresByEmpresa lista proveedores por empresa.
func GetProveedoresByEmpresa(dbConn *sql.DB, empresaID int64, incluirInactivos bool) ([]Proveedor, error) {
	query := `SELECT id, empresa_id, codigo, nombre, documento, contacto, telefono, email, direccion,
		COALESCE(precio_base_referencial, 0), COALESCE(descuento_porcentaje, 0), COALESCE(plazo_pago_dias, 0),
		catalogo_referencia, condicion_entrega,
		fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
	FROM proveedores WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	if !incluirInactivos {
		query += ` AND estado = 'activo'`
	}
	query += ` ORDER BY nombre ASC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Proveedor, 0)
	for rows.Next() {
		var p Proveedor
		var codigo, documento, contacto, telefono, email, direccion, catalogoReferencia, condicionEntrega, fechaCre, fechaAct, usuario, estado, obs sql.NullString
		if err := rows.Scan(&p.ID, &p.EmpresaID, &codigo, &p.Nombre, &documento, &contacto, &telefono, &email, &direccion, &p.PrecioBaseReferencial, &p.DescuentoPorcentaje, &p.PlazoPagoDias, &catalogoReferencia, &condicionEntrega, &fechaCre, &fechaAct, &usuario, &estado, &obs); err != nil {
			return nil, err
		}
		if codigo.Valid {
			p.Codigo = codigo.String
		}
		if documento.Valid {
			p.Documento = documento.String
		}
		if contacto.Valid {
			p.Contacto = contacto.String
		}
		if telefono.Valid {
			p.Telefono = telefono.String
		}
		if email.Valid {
			p.Email = email.String
		}
		if direccion.Valid {
			p.Direccion = direccion.String
		}
		if catalogoReferencia.Valid {
			p.CatalogoReferencia = catalogoReferencia.String
		}
		if condicionEntrega.Valid {
			p.CondicionEntrega = condicionEntrega.String
		}
		if fechaCre.Valid {
			p.FechaCreacion = fechaCre.String
		}
		if fechaAct.Valid {
			p.FechaActualizacion = fechaAct.String
		}
		if usuario.Valid {
			p.UsuarioCreador = usuario.String
		}
		if estado.Valid {
			p.Estado = estado.String
		}
		if obs.Valid {
			p.Observaciones = obs.String
		}
		out = append(out, p)
	}
	return out, nil
}

// UpdateProveedor actualiza proveedor.
func UpdateProveedor(dbConn *sql.DB, p Proveedor) error {
	if err := validateProveedorCondiciones(p); err != nil {
		return err
	}

	_, err := dbConn.Exec(`UPDATE proveedores
		SET codigo = ?, nombre = ?, documento = ?, contacto = ?, telefono = ?, email = ?, direccion = ?,
			catalogo_referencia = ?, precio_base_referencial = ?, descuento_porcentaje = ?, plazo_pago_dias = ?, condicion_entrega = ?,
			observaciones = ?, fecha_actualizacion = datetime('now','localtime')
		WHERE id = ? AND empresa_id = ?`,
		strings.TrimSpace(p.Codigo), strings.TrimSpace(p.Nombre), strings.TrimSpace(p.Documento), strings.TrimSpace(p.Contacto), strings.TrimSpace(p.Telefono), strings.TrimSpace(p.Email), strings.TrimSpace(p.Direccion), strings.TrimSpace(p.CatalogoReferencia), p.PrecioBaseReferencial, p.DescuentoPorcentaje, p.PlazoPagoDias, strings.TrimSpace(p.CondicionEntrega), strings.TrimSpace(p.Observaciones), p.ID, p.EmpresaID)
	return err
}

func validateProveedorCondiciones(p Proveedor) error {
	if p.PrecioBaseReferencial < 0 {
		return fmt.Errorf("precio_base_referencial no puede ser negativo")
	}
	if p.DescuentoPorcentaje < 0 || p.DescuentoPorcentaje > 100 {
		return fmt.Errorf("descuento_porcentaje debe estar entre 0 y 100")
	}
	if p.PlazoPagoDias < 0 {
		return fmt.Errorf("plazo_pago_dias no puede ser negativo")
	}
	return nil
}

// DeleteProveedor elimina proveedor.
func DeleteProveedor(dbConn *sql.DB, empresaID, proveedorID int64) error {
	_, err := dbConn.Exec(`DELETE FROM proveedores WHERE id = ? AND empresa_id = ?`, proveedorID, empresaID)
	return err
}

// SetProveedorEstado activa/desactiva proveedor.
func SetProveedorEstado(dbConn *sql.DB, empresaID, proveedorID int64, estado string) error {
	_, err := dbConn.Exec(`UPDATE proveedores SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE id = ? AND empresa_id = ?`, strings.TrimSpace(estado), proveedorID, empresaID)
	return err
}

func defaultComboUnidad(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" {
		return "combo"
	}
	return v
}

func normalizeComboEstado(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" {
		return "activo"
	}
	return v
}

func validateComboProductoPayload(c ComboProducto) error {
	if c.EmpresaID <= 0 {
		return fmt.Errorf("empresa_id es obligatorio")
	}
	if strings.TrimSpace(c.Nombre) == "" {
		return fmt.Errorf("nombre es obligatorio")
	}
	if c.Precio < 0 {
		return fmt.Errorf("precio invalido")
	}
	if c.ImpuestoPorcentaje < 0 || c.ImpuestoPorcentaje > 100 {
		return fmt.Errorf("impuesto_porcentaje debe estar entre 0 y 100")
	}
	return nil
}

func validateComboIngredientePayload(i ComboProductoDetalle) error {
	if i.ProductoID <= 0 {
		return fmt.Errorf("producto_id es obligatorio")
	}
	if i.Cantidad <= 0 {
		return fmt.Errorf("cantidad debe ser mayor a cero")
	}
	return nil
}

func normalizeComboIngredientesInput(ingredientes []ComboProductoDetalle) ([]ComboProductoDetalle, error) {
	if len(ingredientes) == 0 {
		return nil, fmt.Errorf("el combo debe tener al menos un ingrediente")
	}

	merged := make(map[int64]ComboProductoDetalle)
	for _, raw := range ingredientes {
		if err := validateComboIngredientePayload(raw); err != nil {
			return nil, err
		}
		item, exists := merged[raw.ProductoID]
		if !exists {
			merged[raw.ProductoID] = ComboProductoDetalle{
				ProductoID:    raw.ProductoID,
				Cantidad:      raw.Cantidad,
				UnidadMedida:  strings.TrimSpace(raw.UnidadMedida),
				Estado:        normalizeComboEstado(raw.Estado),
				Observaciones: strings.TrimSpace(raw.Observaciones),
			}
			continue
		}
		item.Cantidad = round2(item.Cantidad + raw.Cantidad)
		if strings.TrimSpace(item.UnidadMedida) == "" {
			item.UnidadMedida = strings.TrimSpace(raw.UnidadMedida)
		}
		if strings.TrimSpace(item.Observaciones) == "" {
			item.Observaciones = strings.TrimSpace(raw.Observaciones)
		}
		if item.Estado == "" {
			item.Estado = normalizeComboEstado(raw.Estado)
		}
		merged[raw.ProductoID] = item
	}

	keys := make([]int64, 0, len(merged))
	for productID := range merged {
		keys = append(keys, productID)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	out := make([]ComboProductoDetalle, 0, len(keys))
	for _, productID := range keys {
		item := merged[productID]
		if item.Cantidad <= 0 {
			continue
		}
		item.Cantidad = round2(item.Cantidad)
		item.Estado = normalizeComboEstado(item.Estado)
		out = append(out, item)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("el combo debe tener al menos un ingrediente")
	}
	return out, nil
}

type comboCostoMetrics struct {
	CostoTeorico      float64
	CostoReal         float64
	VariacionCosto    float64
	VariacionCostoPct float64
}

func normalizeComboUnidadDetalle(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" {
		return "unidad"
	}
	return v
}

func comboIngredientesEquivalent(a, b []ComboProductoDetalle) bool {
	na, errA := normalizeComboIngredientesInput(a)
	nb, errB := normalizeComboIngredientesInput(b)
	if errA != nil || errB != nil {
		return false
	}
	if len(na) != len(nb) {
		return false
	}
	for idx := range na {
		if na[idx].ProductoID != nb[idx].ProductoID {
			return false
		}
		if round2(na[idx].Cantidad) != round2(nb[idx].Cantidad) {
			return false
		}
		if normalizeComboUnidadDetalle(na[idx].UnidadMedida) != normalizeComboUnidadDetalle(nb[idx].UnidadMedida) {
			return false
		}
	}
	return true
}

func buildComboIngredientesValidadosTx(tx *sql.Tx, empresaID int64, ingredientes []ComboProductoDetalle) ([]ComboProductoDetalle, comboCostoMetrics, error) {
	metrics := comboCostoMetrics{}
	normalized, err := normalizeComboIngredientesInput(ingredientes)
	if err != nil {
		return nil, metrics, err
	}

	politicaCosto, err := getInventarioPoliticaCostoTx(tx, empresaID)
	if err != nil {
		return nil, metrics, err
	}
	usarCostoLotes := politicaCosto == inventarioPoliticaCostoPEPS || politicaCosto == inventarioPoliticaCostoPromedio

	prepared := make([]ComboProductoDetalle, 0, len(normalized))
	for _, it := range normalized {
		var unidadProducto sql.NullString
		var costoProducto float64
		err := tx.QueryRow(`SELECT COALESCE(unidad_medida, 'unidad'), COALESCE(costo, 0)
			FROM productos
			WHERE empresa_id = ? AND id = ? AND COALESCE(estado, 'activo') = 'activo'
			LIMIT 1`, empresaID, it.ProductoID).Scan(&unidadProducto, &costoProducto)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, metrics, fmt.Errorf("producto %d no existe o esta inactivo", it.ProductoID)
			}
			return nil, metrics, err
		}

		unidad := strings.TrimSpace(it.UnidadMedida)
		if unidad == "" && unidadProducto.Valid {
			unidad = strings.TrimSpace(unidadProducto.String)
		}
		if unidad == "" {
			unidad = "unidad"
		}

		costoTeoricoUnit := round2(math.Max(costoProducto, 0))
		costoRealUnit := costoTeoricoUnit
		if usarCostoLotes {
			promedioLotes, err := calcularCostoPromedioDisponibleTx(tx, empresaID, it.ProductoID, 0)
			if err != nil {
				return nil, metrics, err
			}
			if promedioLotes > 0 {
				costoRealUnit = promedioLotes
			}
		}

		metrics.CostoTeorico += it.Cantidad * costoTeoricoUnit
		metrics.CostoReal += it.Cantidad * costoRealUnit

		it.UnidadMedida = normalizeComboUnidadDetalle(unidad)
		it.Estado = normalizeComboEstado(it.Estado)
		prepared = append(prepared, it)
	}

	metrics.CostoTeorico = round2(metrics.CostoTeorico)
	metrics.CostoReal = round2(metrics.CostoReal)
	metrics.VariacionCosto = round2(metrics.CostoReal - metrics.CostoTeorico)
	if metrics.CostoTeorico > 0 {
		metrics.VariacionCostoPct = round2((metrics.VariacionCosto / metrics.CostoTeorico) * 100)
	}

	return prepared, metrics, nil
}

func validateComboCostos(c ComboProducto, metrics comboCostoMetrics) error {
	precioCombo := round2(c.Precio)
	if metrics.CostoReal > 0 && precioCombo+0.0001 < metrics.CostoReal {
		return fmt.Errorf("el precio del combo (%.2f) no cubre el costo real (%.2f)", precioCombo, metrics.CostoReal)
	}
	if metrics.CostoTeorico > 0 && math.Abs(metrics.VariacionCostoPct) > comboCostoVariacionMaximaPct {
		return fmt.Errorf("la variacion de costo teorico vs real (%.2f%%) supera el maximo permitido (%.2f%%)", math.Abs(metrics.VariacionCostoPct), comboCostoVariacionMaximaPct)
	}
	return nil
}

func getComboProductoCurrentStateTx(tx *sql.Tx, empresaID, comboID int64) (int64, comboCostoMetrics, error) {
	metrics := comboCostoMetrics{}
	var version sql.NullInt64
	err := tx.QueryRow(`SELECT
		COALESCE(receta_version, 1),
		COALESCE(costo_teorico, 0),
		COALESCE(costo_real, 0),
		COALESCE(variacion_costo, 0),
		COALESCE(variacion_costo_porcentaje, 0)
	FROM combos_productos
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, comboID).Scan(
		&version,
		&metrics.CostoTeorico,
		&metrics.CostoReal,
		&metrics.VariacionCosto,
		&metrics.VariacionCostoPct,
	)
	if err != nil {
		return 0, metrics, err
	}
	if !version.Valid || version.Int64 <= 0 {
		return 1, metrics, nil
	}
	return version.Int64, metrics, nil
}

func getComboProductoIngredientesTx(tx *sql.Tx, empresaID, comboID int64) ([]ComboProductoDetalle, error) {
	rows, err := tx.Query(`SELECT
		COALESCE(producto_id, 0),
		COALESCE(cantidad, 0),
		COALESCE(unidad_medida, 'unidad')
	FROM combos_productos_detalle
	WHERE empresa_id = ? AND combo_id = ? AND COALESCE(estado, 'activo') = 'activo'
	ORDER BY producto_id ASC, id ASC`, empresaID, comboID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ComboProductoDetalle, 0)
	for rows.Next() {
		var item ComboProductoDetalle
		if err := rows.Scan(&item.ProductoID, &item.Cantidad, &item.UnidadMedida); err != nil {
			return nil, err
		}
		item.UnidadMedida = normalizeComboUnidadDetalle(item.UnidadMedida)
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("el combo debe tener al menos un ingrediente")
	}
	return out, nil
}

func insertComboVersionSnapshotTx(tx *sql.Tx, empresaID, comboID, recetaVersion int64, ingredientes []ComboProductoDetalle, metrics comboCostoMetrics, usuario, motivo string) error {
	if recetaVersion <= 0 {
		recetaVersion = 1
	}
	payload, err := json.Marshal(ingredientes)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT OR IGNORE INTO combos_productos_versiones (
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones,
		empresa_id,
		combo_id,
		receta_version,
		ingredientes_json,
		costo_teorico,
		costo_real,
		variacion_costo,
		variacion_costo_porcentaje,
		motivo
	) VALUES (
		datetime('now','localtime'),
		datetime('now','localtime'),
		?,
		'activo',
		?,
		?,
		?,
		?,
		?,
		?,
		?,
		?,
		?,
		?
	)`,
		strings.TrimSpace(usuario),
		strings.TrimSpace(motivo),
		empresaID,
		comboID,
		recetaVersion,
		string(payload),
		round2(metrics.CostoTeorico),
		round2(metrics.CostoReal),
		round2(metrics.VariacionCosto),
		round2(metrics.VariacionCostoPct),
		strings.TrimSpace(motivo),
	)
	return err
}

func replaceComboIngredientesTx(tx *sql.Tx, empresaID, comboID int64, ingredientes []ComboProductoDetalle, usuario string) error {
	if len(ingredientes) == 0 {
		return fmt.Errorf("el combo debe tener al menos un ingrediente")
	}

	if _, err := tx.Exec(`DELETE FROM combos_productos_detalle WHERE empresa_id = ? AND combo_id = ?`, empresaID, comboID); err != nil {
		return err
	}

	for _, it := range ingredientes {
		if _, err := tx.Exec(`INSERT INTO combos_productos_detalle (
			fecha_creacion,
			fecha_actualizacion,
			usuario_creador,
			estado,
			observaciones,
			empresa_id,
			combo_id,
			producto_id,
			cantidad,
			unidad_medida
		) VALUES (
			datetime('now','localtime'),
			datetime('now','localtime'),
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			?
		)`,
			strings.TrimSpace(usuario),
			normalizeComboEstado(it.Estado),
			strings.TrimSpace(it.Observaciones),
			empresaID,
			comboID,
			it.ProductoID,
			round2(it.Cantidad),
			normalizeComboUnidadDetalle(it.UnidadMedida),
		); err != nil {
			return err
		}
	}

	return nil
}

func countActiveOpenCarritoItemsByComboTx(tx *sql.Tx, empresaID, comboID int64) (int64, error) {
	var total int64
	err := tx.QueryRow(`SELECT COUNT(1)
	FROM carrito_compra_items i
	JOIN carritos_compras c ON c.empresa_id = i.empresa_id AND c.id = i.carrito_id
	WHERE i.empresa_id = ?
	  AND COALESCE(LOWER(i.tipo_item), '') = 'combo'
	  AND COALESCE(i.referencia_id, 0) = ?
	  AND COALESCE(i.estado, 'activo') = 'activo'
	  AND COALESCE(c.estado_carrito, 'abierto') <> 'cerrado'`, empresaID, comboID).Scan(&total)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such table") {
			return 0, nil
		}
		return 0, err
	}
	return total, nil
}

// CreateComboProducto crea un combo y su receta de ingredientes por empresa.
func CreateComboProducto(dbConn *sql.DB, combo ComboProducto, ingredientes []ComboProductoDetalle) (int64, error) {
	if err := validateComboProductoPayload(combo); err != nil {
		return 0, err
	}

	combo.UnidadMedida = defaultComboUnidad(combo.UnidadMedida)
	combo.Estado = normalizeComboEstado(combo.Estado)

	tx, err := dbConn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	preparedIngredientes, metrics, err := buildComboIngredientesValidadosTx(tx, combo.EmpresaID, ingredientes)
	if err != nil {
		return 0, err
	}
	if err := validateComboCostos(combo, metrics); err != nil {
		return 0, err
	}

	res, err := tx.Exec(`INSERT INTO combos_productos (
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones,
		empresa_id,
		codigo,
		nombre,
		descripcion,
		unidad_medida,
		precio,
		impuesto_porcentaje,
		receta_version,
		costo_teorico,
		costo_real,
		variacion_costo,
		variacion_costo_porcentaje
	) VALUES (
		datetime('now','localtime'),
		datetime('now','localtime'),
		?,
		?,
		?,
		?,
		NULLIF(?, ''),
		?,
		?,
		?,
		?,
		?,
		1,
		?,
		?,
		?,
		?
	)`,
		strings.TrimSpace(combo.UsuarioCreador),
		combo.Estado,
		strings.TrimSpace(combo.Observaciones),
		combo.EmpresaID,
		strings.TrimSpace(combo.Codigo),
		strings.TrimSpace(combo.Nombre),
		strings.TrimSpace(combo.Descripcion),
		combo.UnidadMedida,
		round2(combo.Precio),
		round2(combo.ImpuestoPorcentaje),
		metrics.CostoTeorico,
		metrics.CostoReal,
		metrics.VariacionCosto,
		metrics.VariacionCostoPct,
	)
	if err != nil {
		return 0, err
	}

	comboID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err := replaceComboIngredientesTx(tx, combo.EmpresaID, comboID, preparedIngredientes, combo.UsuarioCreador); err != nil {
		return 0, err
	}

	if err := insertComboVersionSnapshotTx(tx, combo.EmpresaID, comboID, 1, preparedIngredientes, metrics, combo.UsuarioCreador, "creacion_combo"); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return comboID, nil
}

// GetCombosProductosByEmpresa lista combos por empresa con filtros opcionales.
func GetCombosProductosByEmpresa(dbConn *sql.DB, empresaID int64, filtro, estado string, incluirInactivos bool, limit, offset int) ([]ComboProducto, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	query := `SELECT
		c.id,
		c.empresa_id,
		COALESCE(c.codigo, ''),
		COALESCE(c.nombre, ''),
		COALESCE(c.descripcion, ''),
		COALESCE(c.unidad_medida, 'combo'),
		COALESCE(c.precio, 0),
		COALESCE(c.impuesto_porcentaje, 0),
		COALESCE(c.receta_version, 1),
		COALESCE(c.costo_teorico, 0),
		COALESCE(c.costo_real, 0),
		COALESCE(c.variacion_costo, 0),
		COALESCE(c.variacion_costo_porcentaje, 0),
		COALESCE(det.ingredientes_count, 0),
		COALESCE(c.fecha_creacion, ''),
		COALESCE(c.fecha_actualizacion, ''),
		COALESCE(c.usuario_creador, ''),
		COALESCE(c.estado, 'activo'),
		COALESCE(c.observaciones, '')
	FROM combos_productos c
	LEFT JOIN (
		SELECT empresa_id, combo_id, COUNT(1) AS ingredientes_count
		FROM combos_productos_detalle
		WHERE COALESCE(estado, 'activo') = 'activo'
		GROUP BY empresa_id, combo_id
	) det ON det.empresa_id = c.empresa_id AND det.combo_id = c.id
	WHERE c.empresa_id = ?`

	args := []interface{}{empresaID}
	if strings.TrimSpace(estado) != "" {
		query += ` AND COALESCE(c.estado, 'activo') = ?`
		args = append(args, strings.TrimSpace(estado))
	} else if !incluirInactivos {
		query += ` AND COALESCE(c.estado, 'activo') = 'activo'`
	}

	if strings.TrimSpace(filtro) != "" {
		like := "%" + strings.TrimSpace(filtro) + "%"
		query += ` AND (c.nombre LIKE ? OR c.codigo LIKE ? OR c.descripcion LIKE ?)`
		args = append(args, like, like, like)
	}

	query += ` ORDER BY c.id DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ComboProducto, 0)
	for rows.Next() {
		var c ComboProducto
		if err := rows.Scan(
			&c.ID,
			&c.EmpresaID,
			&c.Codigo,
			&c.Nombre,
			&c.Descripcion,
			&c.UnidadMedida,
			&c.Precio,
			&c.ImpuestoPorcentaje,
			&c.RecetaVersion,
			&c.CostoTeorico,
			&c.CostoReal,
			&c.VariacionCosto,
			&c.VariacionCostoPct,
			&c.IngredientesCount,
			&c.FechaCreacion,
			&c.FechaActualizacion,
			&c.UsuarioCreador,
			&c.Estado,
			&c.Observaciones,
		); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

// GetComboProductoByID obtiene un combo por empresa e incluye su receta.
func GetComboProductoByID(dbConn *sql.DB, empresaID, comboID int64) (*ComboProducto, error) {
	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(codigo, ''),
		COALESCE(nombre, ''),
		COALESCE(descripcion, ''),
		COALESCE(unidad_medida, 'combo'),
		COALESCE(precio, 0),
		COALESCE(impuesto_porcentaje, 0),
		COALESCE(receta_version, 1),
		COALESCE(costo_teorico, 0),
		COALESCE(costo_real, 0),
		COALESCE(variacion_costo, 0),
		COALESCE(variacion_costo_porcentaje, 0),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM combos_productos
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, comboID)

	var combo ComboProducto
	if err := row.Scan(
		&combo.ID,
		&combo.EmpresaID,
		&combo.Codigo,
		&combo.Nombre,
		&combo.Descripcion,
		&combo.UnidadMedida,
		&combo.Precio,
		&combo.ImpuestoPorcentaje,
		&combo.RecetaVersion,
		&combo.CostoTeorico,
		&combo.CostoReal,
		&combo.VariacionCosto,
		&combo.VariacionCostoPct,
		&combo.FechaCreacion,
		&combo.FechaActualizacion,
		&combo.UsuarioCreador,
		&combo.Estado,
		&combo.Observaciones,
	); err != nil {
		return nil, err
	}

	ingredientes, err := GetComboProductoIngredientes(dbConn, empresaID, comboID, true)
	if err != nil {
		return nil, err
	}
	combo.Ingredientes = ingredientes
	combo.IngredientesCount = int64(len(ingredientes))
	return &combo, nil
}

// GetComboProductoIngredientes lista la receta de un combo.
func GetComboProductoIngredientes(dbConn *sql.DB, empresaID, comboID int64, incluirInactivos bool) ([]ComboProductoDetalle, error) {
	query := `SELECT
		d.id,
		d.empresa_id,
		d.combo_id,
		d.producto_id,
		COALESCE(p.nombre, ''),
		COALESCE(p.sku, ''),
		COALESCE(p.codigo_barras, ''),
		COALESCE(d.cantidad, 0),
		COALESCE(d.unidad_medida, 'unidad'),
		COALESCE(d.fecha_creacion, ''),
		COALESCE(d.fecha_actualizacion, ''),
		COALESCE(d.usuario_creador, ''),
		COALESCE(d.estado, 'activo'),
		COALESCE(d.observaciones, '')
	FROM combos_productos_detalle d
	LEFT JOIN productos p ON p.empresa_id = d.empresa_id AND p.id = d.producto_id
	WHERE d.empresa_id = ? AND d.combo_id = ?`
	args := []interface{}{empresaID, comboID}
	if !incluirInactivos {
		query += ` AND COALESCE(d.estado, 'activo') = 'activo'`
	}
	query += ` ORDER BY d.id ASC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ComboProductoDetalle, 0)
	for rows.Next() {
		var d ComboProductoDetalle
		if err := rows.Scan(
			&d.ID,
			&d.EmpresaID,
			&d.ComboID,
			&d.ProductoID,
			&d.ProductoNombre,
			&d.ProductoSKU,
			&d.ProductoCodigo,
			&d.Cantidad,
			&d.UnidadMedida,
			&d.FechaCreacion,
			&d.FechaActualizacion,
			&d.UsuarioCreador,
			&d.Estado,
			&d.Observaciones,
		); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}

// UpdateComboProducto actualiza cabecera y receta de un combo.
func UpdateComboProducto(dbConn *sql.DB, combo ComboProducto, ingredientes []ComboProductoDetalle) error {
	if combo.ID <= 0 {
		return fmt.Errorf("id es obligatorio")
	}
	if err := validateComboProductoPayload(combo); err != nil {
		return err
	}

	combo.UnidadMedida = defaultComboUnidad(combo.UnidadMedida)
	combo.Estado = normalizeComboEstado(combo.Estado)

	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	openItems, err := countActiveOpenCarritoItemsByComboTx(tx, combo.EmpresaID, combo.ID)
	if err != nil {
		return err
	}
	if openItems > 0 {
		return fmt.Errorf("no se puede modificar el combo mientras tenga items activos en carritos abiertos")
	}

	currentVersion, currentMetrics, err := getComboProductoCurrentStateTx(tx, combo.EmpresaID, combo.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sql.ErrNoRows
		}
		return err
	}
	if currentVersion <= 0 {
		currentVersion = 1
	}

	currentIngredientes, err := getComboProductoIngredientesTx(tx, combo.EmpresaID, combo.ID)
	if err != nil {
		return err
	}

	preparedIngredientes, metrics, err := buildComboIngredientesValidadosTx(tx, combo.EmpresaID, ingredientes)
	if err != nil {
		return err
	}
	if err := validateComboCostos(combo, metrics); err != nil {
		return err
	}

	recipeChanged := !comboIngredientesEquivalent(currentIngredientes, preparedIngredientes)
	nextVersion := currentVersion
	if recipeChanged {
		if err := insertComboVersionSnapshotTx(tx, combo.EmpresaID, combo.ID, currentVersion, currentIngredientes, currentMetrics, combo.UsuarioCreador, "snapshot_pre_actualizacion"); err != nil {
			return err
		}
		nextVersion = currentVersion + 1
	}

	res, err := tx.Exec(`UPDATE combos_productos SET
		codigo = NULLIF(?, ''),
		nombre = ?,
		descripcion = ?,
		unidad_medida = ?,
		precio = ?,
		impuesto_porcentaje = ?,
		receta_version = ?,
		costo_teorico = ?,
		costo_real = ?,
		variacion_costo = ?,
		variacion_costo_porcentaje = ?,
		estado = ?,
		observaciones = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE id = ? AND empresa_id = ?`,
		strings.TrimSpace(combo.Codigo),
		strings.TrimSpace(combo.Nombre),
		strings.TrimSpace(combo.Descripcion),
		combo.UnidadMedida,
		round2(combo.Precio),
		round2(combo.ImpuestoPorcentaje),
		nextVersion,
		metrics.CostoTeorico,
		metrics.CostoReal,
		metrics.VariacionCosto,
		metrics.VariacionCostoPct,
		combo.Estado,
		strings.TrimSpace(combo.Observaciones),
		combo.ID,
		combo.EmpresaID,
	)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}

	if err := replaceComboIngredientesTx(tx, combo.EmpresaID, combo.ID, preparedIngredientes, combo.UsuarioCreador); err != nil {
		return err
	}

	if recipeChanged {
		if err := insertComboVersionSnapshotTx(tx, combo.EmpresaID, combo.ID, nextVersion, preparedIngredientes, metrics, combo.UsuarioCreador, "actualizacion_receta"); err != nil {
			return err
		}
	}

	if !recipeChanged && currentVersion == 1 {
		if err := insertComboVersionSnapshotTx(tx, combo.EmpresaID, combo.ID, currentVersion, preparedIngredientes, metrics, combo.UsuarioCreador, "normalizacion_version_inicial"); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// DeleteComboProducto elimina un combo y su receta.
func DeleteComboProducto(dbConn *sql.DB, empresaID, comboID int64) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	openItems, err := countActiveOpenCarritoItemsByComboTx(tx, empresaID, comboID)
	if err != nil {
		return err
	}
	if openItems > 0 {
		return fmt.Errorf("no se puede eliminar el combo mientras tenga items activos en carritos abiertos")
	}

	if _, err := tx.Exec(`DELETE FROM combos_productos_detalle WHERE empresa_id = ? AND combo_id = ?`, empresaID, comboID); err != nil {
		return err
	}
	res, err := tx.Exec(`DELETE FROM combos_productos WHERE empresa_id = ? AND id = ?`, empresaID, comboID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}

	return tx.Commit()
}

// SetComboProductoEstado activa/desactiva un combo por empresa.
func SetComboProductoEstado(dbConn *sql.DB, empresaID, comboID int64, estado string) error {
	nuevoEstado := normalizeComboEstado(estado)

	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if nuevoEstado == "inactivo" {
		openItems, err := countActiveOpenCarritoItemsByComboTx(tx, empresaID, comboID)
		if err != nil {
			return err
		}
		if openItems > 0 {
			return fmt.Errorf("no se puede desactivar el combo mientras tenga items activos en carritos abiertos")
		}
	}

	res, err := tx.Exec(`UPDATE combos_productos
	SET estado = ?, fecha_actualizacion = datetime('now','localtime')
	WHERE id = ? AND empresa_id = ?`, nuevoEstado, comboID, empresaID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}

	return tx.Commit()
}

// CreateServicio crea un servicio comercial por empresa.
func CreateServicio(dbConn *sql.DB, s Servicio) (int64, error) {
	res, err := dbConn.Exec(`INSERT INTO servicios (
		empresa_id, codigo, nombre, descripcion, categoria, duracion_minutos, costo_referencial, precio, impuesto_porcentaje, imagen_url,
		usuario_creador, estado, observaciones, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, COALESCE(NULLIF(?, ''), 'activo'), ?, datetime('now','localtime'), datetime('now','localtime'))`,
		s.EmpresaID, strings.TrimSpace(s.Codigo), strings.TrimSpace(s.Nombre), strings.TrimSpace(s.Descripcion), strings.TrimSpace(s.Categoria), s.DuracionMinutos, s.CostoReferencial, s.Precio, s.ImpuestoPorcentaje, strings.TrimSpace(s.ImagenURL), strings.TrimSpace(s.UsuarioCreador), strings.TrimSpace(s.Estado), strings.TrimSpace(s.Observaciones))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetServiciosByEmpresa lista servicios por empresa.
func GetServiciosByEmpresa(dbConn *sql.DB, empresaID int64, filtro, estado string, limit, offset int) ([]Servicio, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	query := `SELECT id, empresa_id, codigo, nombre, descripcion, categoria, duracion_minutos, costo_referencial, precio, impuesto_porcentaje, imagen_url,
		fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
	FROM servicios WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	if strings.TrimSpace(estado) != "" {
		query += ` AND estado = ?`
		args = append(args, strings.TrimSpace(estado))
	}
	if strings.TrimSpace(filtro) != "" {
		like := "%" + strings.TrimSpace(filtro) + "%"
		query += ` AND (nombre LIKE ? OR codigo LIKE ? OR categoria LIKE ?)`
		args = append(args, like, like, like)
	}
	query += ` ORDER BY id DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Servicio, 0)
	for rows.Next() {
		var s Servicio
		var codigo, desc, categoria, imagen, fechaCre, fechaAct, usuario, estadoVal, obs sql.NullString
		if err := rows.Scan(&s.ID, &s.EmpresaID, &codigo, &s.Nombre, &desc, &categoria, &s.DuracionMinutos, &s.CostoReferencial, &s.Precio, &s.ImpuestoPorcentaje, &imagen, &fechaCre, &fechaAct, &usuario, &estadoVal, &obs); err != nil {
			return nil, err
		}
		if codigo.Valid {
			s.Codigo = codigo.String
		}
		if desc.Valid {
			s.Descripcion = desc.String
		}
		if categoria.Valid {
			s.Categoria = categoria.String
		}
		if imagen.Valid {
			s.ImagenURL = imagen.String
		}
		if fechaCre.Valid {
			s.FechaCreacion = fechaCre.String
		}
		if fechaAct.Valid {
			s.FechaActualizacion = fechaAct.String
		}
		if usuario.Valid {
			s.UsuarioCreador = usuario.String
		}
		if estadoVal.Valid {
			s.Estado = estadoVal.String
		}
		if obs.Valid {
			s.Observaciones = obs.String
		}
		out = append(out, s)
	}
	return out, nil
}

// UpdateServicio actualiza servicio.
func UpdateServicio(dbConn *sql.DB, s Servicio) error {
	_, err := dbConn.Exec(`UPDATE servicios
		SET codigo = ?, nombre = ?, descripcion = ?, categoria = ?, duracion_minutos = ?, costo_referencial = ?, precio = ?, impuesto_porcentaje = ?, imagen_url = ?, observaciones = ?, fecha_actualizacion = datetime('now','localtime')
		WHERE id = ? AND empresa_id = ?`,
		strings.TrimSpace(s.Codigo), strings.TrimSpace(s.Nombre), strings.TrimSpace(s.Descripcion), strings.TrimSpace(s.Categoria), s.DuracionMinutos, s.CostoReferencial, s.Precio, s.ImpuestoPorcentaje, strings.TrimSpace(s.ImagenURL), strings.TrimSpace(s.Observaciones), s.ID, s.EmpresaID)
	return err
}

// DeleteServicio elimina servicio.
func DeleteServicio(dbConn *sql.DB, empresaID, servicioID int64) error {
	_, err := dbConn.Exec(`DELETE FROM servicios WHERE id = ? AND empresa_id = ?`, servicioID, empresaID)
	return err
}

// SetServicioEstado activa/desactiva servicio.
func SetServicioEstado(dbConn *sql.DB, empresaID, servicioID int64, estado string) error {
	_, err := dbConn.Exec(`UPDATE servicios SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE id = ? AND empresa_id = ?`, strings.TrimSpace(estado), servicioID, empresaID)
	return err
}

// GetProductoPrecioHistorialByEmpresa lista historial de precios de productos.
func GetProductoPrecioHistorialByEmpresa(dbConn *sql.DB, empresaID, productoID int64, limit, offset int) ([]ProductoPrecioHistorial, error) {
	if limit <= 0 || limit > 1000 {
		limit = 300
	}
	if offset < 0 {
		offset = 0
	}

	query := `SELECT h.id, h.empresa_id, h.producto_id, p.nombre,
		h.costo_anterior, h.costo_nuevo, h.precio_anterior, h.precio_nuevo, h.impuesto_anterior, h.impuesto_nuevo,
		h.motivo, h.referencia, h.fecha_cambio, h.fecha_creacion, h.fecha_actualizacion, h.usuario_creador, h.estado, h.observaciones
	FROM producto_precios_historial h
	JOIN productos p ON p.id = h.producto_id AND p.empresa_id = h.empresa_id
	WHERE h.empresa_id = ?`
	args := []interface{}{empresaID}
	if productoID > 0 {
		query += ` AND h.producto_id = ?`
		args = append(args, productoID)
	}
	query += ` ORDER BY h.id DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ProductoPrecioHistorial, 0)
	for rows.Next() {
		var h ProductoPrecioHistorial
		var motivo, referencia, fechaCambio, fechaCre, fechaAct, usuario, estadoVal, obs sql.NullString
		if err := rows.Scan(&h.ID, &h.EmpresaID, &h.ProductoID, &h.ProductoNombre, &h.CostoAnterior, &h.CostoNuevo, &h.PrecioAnterior, &h.PrecioNuevo, &h.ImpuestoAnterior, &h.ImpuestoNuevo, &motivo, &referencia, &fechaCambio, &fechaCre, &fechaAct, &usuario, &estadoVal, &obs); err != nil {
			return nil, err
		}
		if motivo.Valid {
			h.Motivo = motivo.String
		}
		if referencia.Valid {
			h.Referencia = referencia.String
		}
		if fechaCambio.Valid {
			h.FechaCambio = fechaCambio.String
		}
		if fechaCre.Valid {
			h.FechaCreacion = fechaCre.String
		}
		if fechaAct.Valid {
			h.FechaActualizacion = fechaAct.String
		}
		if usuario.Valid {
			h.UsuarioCreador = usuario.String
		}
		if estadoVal.Valid {
			h.Estado = estadoVal.String
		}
		if obs.Valid {
			h.Observaciones = obs.String
		}
		out = append(out, h)
	}
	return out, nil
}

// RegistrarMovimientoInventario registra entradas, salidas, devoluciones y pérdidas con impacto en stock.
func RegistrarMovimientoInventario(dbConn *sql.DB, empresaID, productoID, bodegaID int64, tipo string, cantidad float64, referencia, usuario, observaciones string) error {
	tipo = strings.ToLower(strings.TrimSpace(tipo))
	referencia = strings.TrimSpace(referencia)
	usuario = strings.TrimSpace(usuario)
	observaciones = strings.TrimSpace(observaciones)
	if cantidad <= 0 {
		return fmt.Errorf("cantidad debe ser mayor a 0")
	}
	if tipo == "" {
		return fmt.Errorf("tipo de movimiento requerido")
	}

	incoming := tipo == "entrada" || tipo == "devolucion" || tipo == "ajuste_entrada" || tipo == "ajuste_positivo"
	outgoing := tipo == "salida" || tipo == "perdida" || tipo == "ajuste_salida" || tipo == "ajuste_negativo"
	if !incoming && !outgoing {
		return fmt.Errorf("tipo de movimiento no soportado: %s", tipo)
	}

	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := validateBodegaEmpresaTx(tx, empresaID, bodegaID); err != nil {
		return err
	}
	if err := validateProductoEmpresaTx(tx, empresaID, productoID); err != nil {
		return err
	}

	politicaCosto, err := getInventarioPoliticaCostoTx(tx, empresaID)
	if err != nil {
		return err
	}

	var costoUnitario float64
	if err := tx.QueryRow(`SELECT COALESCE(costo, 0) FROM productos WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, productoID).Scan(&costoUnitario); err != nil {
		return err
	}
	if costoUnitario < 0 {
		costoUnitario = 0
	}

	if incoming {
		costoEntrada := costoUnitario
		if costoEntrada <= 0 {
			promedio, err := calcularCostoPromedioDisponibleTx(tx, empresaID, productoID, bodegaID)
			if err != nil {
				return err
			}
			if promedio > 0 {
				costoEntrada = promedio
			}
		}

		if err := upsertExistenciaTx(tx, empresaID, productoID, bodegaID, cantidad, usuario, observaciones); err != nil {
			return err
		}
		if err := registerCostoLoteTx(tx, empresaID, productoID, bodegaID, cantidad, costoEntrada, referencia, usuario, observaciones); err != nil {
			return err
		}
		if politicaCosto == inventarioPoliticaCostoPromedio {
			if err := recalcularCostoPromedioProductoTx(tx, empresaID, productoID); err != nil {
				return err
			}
		}
		if err := insertMovimientoTx(tx, InventarioMovimiento{
			EmpresaID:       empresaID,
			ProductoID:      productoID,
			BodegaDestinoID: bodegaID,
			Tipo:            tipo,
			Cantidad:        cantidad,
			CostoUnitario:   costoEntrada,
			Referencia:      referencia,
			UsuarioCreador:  usuario,
			Estado:          "activo",
			Observaciones:   observaciones,
		}); err != nil {
			return err
		}
		return tx.Commit()
	}

	var stockActual float64
	err = tx.QueryRow(`SELECT cantidad FROM inventario_existencias WHERE empresa_id = ? AND producto_id = ? AND bodega_id = ?`, empresaID, productoID, bodegaID).Scan(&stockActual)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrStockInsuficiente
		}
		return err
	}
	if stockActual < cantidad {
		return ErrStockInsuficiente
	}

	if err := ensureCostoLotesSeedTx(tx, empresaID, productoID, bodegaID, usuario); err != nil {
		return err
	}

	costoSalida := costoUnitario
	if politicaCosto == inventarioPoliticaCostoPromedio {
		promedio, err := calcularCostoPromedioDisponibleTx(tx, empresaID, productoID, bodegaID)
		if err != nil {
			return err
		}
		if promedio > 0 {
			costoSalida = promedio
		}
		if _, err := consumirCostoLotesPEPSTx(tx, empresaID, productoID, bodegaID, cantidad); err != nil && !errors.Is(err, ErrStockInsuficiente) {
			return err
		}
	} else {
		costoPEPS, err := consumirCostoLotesPEPSTx(tx, empresaID, productoID, bodegaID, cantidad)
		if err != nil {
			if !errors.Is(err, ErrStockInsuficiente) {
				return err
			}
		} else if costoPEPS > 0 {
			costoSalida = costoPEPS
		}
	}

	if _, err := tx.Exec(`UPDATE inventario_existencias SET cantidad = cantidad - ?, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ? AND producto_id = ? AND bodega_id = ?`, cantidad, empresaID, productoID, bodegaID); err != nil {
		return err
	}
	if err := insertMovimientoTx(tx, InventarioMovimiento{
		EmpresaID:      empresaID,
		ProductoID:     productoID,
		BodegaOrigenID: bodegaID,
		Tipo:           tipo,
		Cantidad:       cantidad,
		CostoUnitario:  costoSalida,
		Referencia:     referencia,
		UsuarioCreador: usuario,
		Estado:         "activo",
		Observaciones:  observaciones,
	}); err != nil {
		return err
	}
	if politicaCosto == inventarioPoliticaCostoPromedio {
		if err := recalcularCostoPromedioProductoTx(tx, empresaID, productoID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// RegistrarConteoCiclicoInventario registra conteo ciclico y aplica ajuste auditado cuando hay diferencia.
func RegistrarConteoCiclicoInventario(dbConn *sql.DB, conteo InventarioConteoCiclico) (InventarioConteoCiclico, error) {
	if conteo.EmpresaID <= 0 {
		return InventarioConteoCiclico{}, fmt.Errorf("empresa_id invalido")
	}
	if conteo.ProductoID <= 0 {
		return InventarioConteoCiclico{}, fmt.Errorf("producto_id invalido")
	}
	if conteo.BodegaID <= 0 {
		return InventarioConteoCiclico{}, fmt.Errorf("bodega_id invalido")
	}
	if conteo.CantidadContada < 0 {
		return InventarioConteoCiclico{}, fmt.Errorf("cantidad_contada no puede ser negativa")
	}

	conteo.Referencia = strings.TrimSpace(conteo.Referencia)
	if conteo.Referencia == "" {
		conteo.Referencia = fmt.Sprintf("CONTEO-CICLICO-%d", time.Now().Unix())
	}
	conteo.UsuarioRevisor = strings.TrimSpace(conteo.UsuarioRevisor)
	if conteo.UsuarioRevisor == "" {
		conteo.UsuarioRevisor = strings.TrimSpace(conteo.UsuarioCreador)
	}

	var exists int
	if err := dbConn.QueryRow(`SELECT 1 FROM productos WHERE empresa_id = ? AND id = ? LIMIT 1`, conteo.EmpresaID, conteo.ProductoID).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return InventarioConteoCiclico{}, fmt.Errorf("producto %d no pertenece a la empresa %d", conteo.ProductoID, conteo.EmpresaID)
		}
		return InventarioConteoCiclico{}, err
	}
	if err := dbConn.QueryRow(`SELECT 1 FROM bodegas WHERE empresa_id = ? AND id = ? LIMIT 1`, conteo.EmpresaID, conteo.BodegaID).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return InventarioConteoCiclico{}, fmt.Errorf("bodega %d no pertenece a la empresa %d", conteo.BodegaID, conteo.EmpresaID)
		}
		return InventarioConteoCiclico{}, err
	}

	stockSistema := 0.0
	err := dbConn.QueryRow(`SELECT COALESCE(cantidad, 0)
		FROM inventario_existencias
		WHERE empresa_id = ? AND producto_id = ? AND bodega_id = ?
		LIMIT 1`, conteo.EmpresaID, conteo.ProductoID, conteo.BodegaID).Scan(&stockSistema)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return InventarioConteoCiclico{}, err
	}

	conteo.CantidadSistema = round2(stockSistema)
	conteo.CantidadContada = round2(conteo.CantidadContada)
	conteo.Variacion = round2(conteo.CantidadContada - conteo.CantidadSistema)
	conteo.TipoAjuste = "sin_ajuste"
	conteo.EstadoConteo = "sin_diferencia"
	conteo.Estado = "activo"

	movimientoID := int64(0)
	if conteo.Variacion != 0 {
		tipoMovimiento := "ajuste_positivo"
		if conteo.Variacion < 0 {
			tipoMovimiento = "ajuste_negativo"
		}
		conteo.TipoAjuste = tipoMovimiento
		conteo.EstadoConteo = "ajustado"

		referenciaAjuste := fmt.Sprintf("%s|AJ-%d", conteo.Referencia, time.Now().UnixNano())
		observacionAjuste := strings.TrimSpace(strings.Join([]string{conteo.Observaciones, "Ajuste automatico por conteo ciclico."}, " "))
		if err := RegistrarMovimientoInventario(
			dbConn,
			conteo.EmpresaID,
			conteo.ProductoID,
			conteo.BodegaID,
			tipoMovimiento,
			round2(math.Abs(conteo.Variacion)),
			referenciaAjuste,
			conteo.UsuarioRevisor,
			observacionAjuste,
		); err != nil {
			return InventarioConteoCiclico{}, err
		}
		_ = dbConn.QueryRow(`SELECT id
			FROM inventario_movimientos
			WHERE empresa_id = ?
				AND producto_id = ?
				AND referencia = ?
			ORDER BY id DESC
			LIMIT 1`, conteo.EmpresaID, conteo.ProductoID, referenciaAjuste).Scan(&movimientoID)
	}

	res, err := dbConn.Exec(`INSERT INTO inventario_conteos_ciclicos (
		empresa_id,
		producto_id,
		bodega_id,
		cantidad_sistema,
		cantidad_contada,
		variacion,
		tipo_ajuste,
		movimiento_id,
		referencia,
		fecha_conteo,
		usuario_revisor,
		estado_conteo,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), ?, ?, datetime('now','localtime'), datetime('now','localtime'), ?, COALESCE(NULLIF(?, ''), 'activo'), ?)`,
		conteo.EmpresaID,
		conteo.ProductoID,
		conteo.BodegaID,
		conteo.CantidadSistema,
		conteo.CantidadContada,
		conteo.Variacion,
		conteo.TipoAjuste,
		nullableInt64(movimientoID),
		conteo.Referencia,
		conteo.UsuarioRevisor,
		conteo.EstadoConteo,
		conteo.UsuarioRevisor,
		conteo.Estado,
		strings.TrimSpace(conteo.Observaciones),
	)
	if err != nil {
		return InventarioConteoCiclico{}, err
	}

	conteoID, err := res.LastInsertId()
	if err != nil {
		return InventarioConteoCiclico{}, err
	}

	return getInventarioConteoCiclicoByID(dbConn, conteo.EmpresaID, conteoID)
}

func getInventarioConteoCiclicoByID(dbConn *sql.DB, empresaID, conteoID int64) (InventarioConteoCiclico, error) {
	row := dbConn.QueryRow(`SELECT
		c.id,
		c.empresa_id,
		c.producto_id,
		COALESCE(p.nombre, ''),
		c.bodega_id,
		COALESCE(b.nombre, ''),
		COALESCE(c.cantidad_sistema, 0),
		COALESCE(c.cantidad_contada, 0),
		COALESCE(c.variacion, 0),
		COALESCE(c.tipo_ajuste, 'sin_ajuste'),
		COALESCE(c.movimiento_id, 0),
		COALESCE(c.referencia, ''),
		COALESCE(c.fecha_conteo, ''),
		COALESCE(c.usuario_revisor, ''),
		COALESCE(c.estado_conteo, 'sin_diferencia'),
		COALESCE(c.fecha_creacion, ''),
		COALESCE(c.fecha_actualizacion, ''),
		COALESCE(c.usuario_creador, ''),
		COALESCE(c.estado, 'activo'),
		COALESCE(c.observaciones, '')
	FROM inventario_conteos_ciclicos c
	LEFT JOIN productos p ON p.empresa_id = c.empresa_id AND p.id = c.producto_id
	LEFT JOIN bodegas b ON b.empresa_id = c.empresa_id AND b.id = c.bodega_id
	WHERE c.empresa_id = ? AND c.id = ?
	LIMIT 1`, empresaID, conteoID)

	var out InventarioConteoCiclico
	if err := row.Scan(
		&out.ID,
		&out.EmpresaID,
		&out.ProductoID,
		&out.ProductoNombre,
		&out.BodegaID,
		&out.BodegaNombre,
		&out.CantidadSistema,
		&out.CantidadContada,
		&out.Variacion,
		&out.TipoAjuste,
		&out.MovimientoID,
		&out.Referencia,
		&out.FechaConteo,
		&out.UsuarioRevisor,
		&out.EstadoConteo,
		&out.FechaCreacion,
		&out.FechaActualizacion,
		&out.UsuarioCreador,
		&out.Estado,
		&out.Observaciones,
	); err != nil {
		return InventarioConteoCiclico{}, err
	}

	return out, nil
}

// GetInventarioConteosCiclicosByEmpresa lista conteos ciclicos auditados por empresa.
func GetInventarioConteosCiclicosByEmpresa(dbConn *sql.DB, empresaID, productoID, bodegaID int64, estadoConteo, desde, hasta string, limit, offset int) ([]InventarioConteoCiclico, error) {
	if limit <= 0 || limit > 500 {
		limit = 120
	}
	if offset < 0 {
		offset = 0
	}

	query := `SELECT
		c.id,
		c.empresa_id,
		c.producto_id,
		COALESCE(p.nombre, ''),
		c.bodega_id,
		COALESCE(b.nombre, ''),
		COALESCE(c.cantidad_sistema, 0),
		COALESCE(c.cantidad_contada, 0),
		COALESCE(c.variacion, 0),
		COALESCE(c.tipo_ajuste, 'sin_ajuste'),
		COALESCE(c.movimiento_id, 0),
		COALESCE(c.referencia, ''),
		COALESCE(c.fecha_conteo, ''),
		COALESCE(c.usuario_revisor, ''),
		COALESCE(c.estado_conteo, 'sin_diferencia'),
		COALESCE(c.fecha_creacion, ''),
		COALESCE(c.fecha_actualizacion, ''),
		COALESCE(c.usuario_creador, ''),
		COALESCE(c.estado, 'activo'),
		COALESCE(c.observaciones, '')
	FROM inventario_conteos_ciclicos c
	LEFT JOIN productos p ON p.empresa_id = c.empresa_id AND p.id = c.producto_id
	LEFT JOIN bodegas b ON b.empresa_id = c.empresa_id AND b.id = c.bodega_id
	WHERE c.empresa_id = ?`
	args := []interface{}{empresaID}

	if productoID > 0 {
		query += ` AND c.producto_id = ?`
		args = append(args, productoID)
	}
	if bodegaID > 0 {
		query += ` AND c.bodega_id = ?`
		args = append(args, bodegaID)
	}
	if strings.TrimSpace(estadoConteo) != "" {
		query += ` AND LOWER(COALESCE(c.estado_conteo, '')) = LOWER(?)`
		args = append(args, strings.TrimSpace(estadoConteo))
	}
	if strings.TrimSpace(desde) != "" {
		query += ` AND date(COALESCE(c.fecha_conteo, c.fecha_creacion)) >= date(?)`
		args = append(args, strings.TrimSpace(desde))
	}
	if strings.TrimSpace(hasta) != "" {
		query += ` AND date(COALESCE(c.fecha_conteo, c.fecha_creacion)) <= date(?)`
		args = append(args, strings.TrimSpace(hasta))
	}

	query += ` ORDER BY datetime(COALESCE(c.fecha_conteo, c.fecha_creacion)) DESC, c.id DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]InventarioConteoCiclico, 0)
	for rows.Next() {
		var row InventarioConteoCiclico
		if err := rows.Scan(
			&row.ID,
			&row.EmpresaID,
			&row.ProductoID,
			&row.ProductoNombre,
			&row.BodegaID,
			&row.BodegaNombre,
			&row.CantidadSistema,
			&row.CantidadContada,
			&row.Variacion,
			&row.TipoAjuste,
			&row.MovimientoID,
			&row.Referencia,
			&row.FechaConteo,
			&row.UsuarioRevisor,
			&row.EstadoConteo,
			&row.FechaCreacion,
			&row.FechaActualizacion,
			&row.UsuarioCreador,
			&row.Estado,
			&row.Observaciones,
		); err != nil {
			return nil, err
		}
		out = append(out, row)
	}

	return out, nil
}

// RegistrarCambioProducto registra cambio de un producto por otro afectando existencias y movimientos.
func RegistrarCambioProducto(dbConn *sql.DB, empresaID, productoOrigenID, productoDestinoID, bodegaID int64, cantidad float64, referencia, usuario, observaciones string) error {
	if cantidad <= 0 {
		return fmt.Errorf("cantidad debe ser mayor a 0")
	}
	if productoOrigenID <= 0 || productoDestinoID <= 0 {
		return fmt.Errorf("producto_origen_id y producto_destino_id son requeridos")
	}

	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := validateBodegaEmpresaTx(tx, empresaID, bodegaID); err != nil {
		return err
	}
	if err := validateProductoEmpresaTx(tx, empresaID, productoOrigenID); err != nil {
		return err
	}
	if err := validateProductoEmpresaTx(tx, empresaID, productoDestinoID); err != nil {
		return err
	}

	politicaCosto, err := getInventarioPoliticaCostoTx(tx, empresaID)
	if err != nil {
		return err
	}

	var costoOrigen float64
	if err := tx.QueryRow(`SELECT COALESCE(costo, 0) FROM productos WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, productoOrigenID).Scan(&costoOrigen); err != nil {
		return err
	}
	var costoDestino float64
	if err := tx.QueryRow(`SELECT COALESCE(costo, 0) FROM productos WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, productoDestinoID).Scan(&costoDestino); err != nil {
		return err
	}
	if costoOrigen < 0 {
		costoOrigen = 0
	}
	if costoDestino < 0 {
		costoDestino = 0
	}

	var stockOrigen float64
	err = tx.QueryRow(`SELECT cantidad FROM inventario_existencias WHERE empresa_id = ? AND producto_id = ? AND bodega_id = ?`, empresaID, productoOrigenID, bodegaID).Scan(&stockOrigen)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrStockInsuficiente
		}
		return err
	}
	if stockOrigen < cantidad {
		return ErrStockInsuficiente
	}

	if err := ensureCostoLotesSeedTx(tx, empresaID, productoOrigenID, bodegaID, strings.TrimSpace(usuario)); err != nil {
		return err
	}
	costoConsumoOrigen, err := consumirCostoLotesPEPSTx(tx, empresaID, productoOrigenID, bodegaID, cantidad)
	if err != nil {
		if !errors.Is(err, ErrStockInsuficiente) {
			return err
		}
		costoConsumoOrigen = costoOrigen
	}
	if costoConsumoOrigen > 0 {
		costoOrigen = costoConsumoOrigen
	}

	if _, err := tx.Exec(`UPDATE inventario_existencias SET cantidad = cantidad - ?, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ? AND producto_id = ? AND bodega_id = ?`, cantidad, empresaID, productoOrigenID, bodegaID); err != nil {
		return err
	}
	if err := upsertExistenciaTx(tx, empresaID, productoDestinoID, bodegaID, cantidad, usuario, "cambio de producto"); err != nil {
		return err
	}
	if err := registerCostoLoteTx(tx, empresaID, productoDestinoID, bodegaID, cantidad, costoDestino, strings.TrimSpace(referencia), strings.TrimSpace(usuario), "ingreso por cambio de producto"); err != nil {
		return err
	}

	if err := insertMovimientoTx(tx, InventarioMovimiento{
		EmpresaID:      empresaID,
		ProductoID:     productoOrigenID,
		BodegaOrigenID: bodegaID,
		Tipo:           "cambio_salida",
		Cantidad:       cantidad,
		CostoUnitario:  costoOrigen,
		Referencia:     strings.TrimSpace(referencia),
		UsuarioCreador: strings.TrimSpace(usuario),
		Estado:         "activo",
		Observaciones:  strings.TrimSpace(observaciones),
	}); err != nil {
		return err
	}
	if err := insertMovimientoTx(tx, InventarioMovimiento{
		EmpresaID:       empresaID,
		ProductoID:      productoDestinoID,
		BodegaDestinoID: bodegaID,
		Tipo:            "cambio_entrada",
		Cantidad:        cantidad,
		CostoUnitario:   costoDestino,
		Referencia:      strings.TrimSpace(referencia),
		UsuarioCreador:  strings.TrimSpace(usuario),
		Estado:          "activo",
		Observaciones:   strings.TrimSpace(observaciones),
	}); err != nil {
		return err
	}
	if politicaCosto == inventarioPoliticaCostoPromedio {
		if err := recalcularCostoPromedioProductoTx(tx, empresaID, productoOrigenID); err != nil {
			return err
		}
		if err := recalcularCostoPromedioProductoTx(tx, empresaID, productoDestinoID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func seedCategoriasProductosFromLegacy(dbConn *sql.DB) error {
	rows, err := dbConn.Query(`SELECT DISTINCT empresa_id, TRIM(categoria)
		FROM productos
		WHERE TRIM(COALESCE(categoria, '')) <> ''`)
	if err != nil {
		return err
	}
	type categoriaSeed struct {
		empresaID int64
		nombre    string
	}
	seeds := make([]categoriaSeed, 0)

	for rows.Next() {
		var empresaID int64
		var nombre string
		if err := rows.Scan(&empresaID, &nombre); err != nil {
			rows.Close()
			return err
		}
		nombre = strings.TrimSpace(nombre)
		if empresaID <= 0 || nombre == "" {
			continue
		}
		seeds = append(seeds, categoriaSeed{empresaID: empresaID, nombre: nombre})
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}

	for _, seed := range seeds {
		if _, err := dbConn.Exec(`INSERT OR IGNORE INTO categorias_productos (
			empresa_id, nombre, descripcion, usuario_creador, estado, observaciones, fecha_creacion, fecha_actualizacion
		) VALUES (?, ?, ?, 'migracion', 'activo', ?, datetime('now','localtime'), datetime('now','localtime'))`,
			seed.empresaID, seed.nombre, "categoria migrada desde productos.categoria", "registro automático por migración"); err != nil {
			return err
		}
	}
	return nil
}

func backfillProductoCategoriaIDs(dbConn *sql.DB) error {
	if _, err := dbConn.Exec(`UPDATE productos
		SET categoria_id = (
			SELECT cp.id
			FROM categorias_productos cp
			WHERE cp.empresa_id = productos.empresa_id
				AND LOWER(TRIM(cp.nombre)) = LOWER(TRIM(productos.categoria))
			LIMIT 1
		)
		WHERE (categoria_id IS NULL OR categoria_id <= 0)
			AND TRIM(COALESCE(categoria, '')) <> ''`); err != nil {
		return err
	}

	if _, err := dbConn.Exec(`UPDATE productos
		SET categoria = (
			SELECT cp.nombre
			FROM categorias_productos cp
			WHERE cp.empresa_id = productos.empresa_id
				AND cp.id = productos.categoria_id
			LIMIT 1
		)
		WHERE categoria_id IS NOT NULL AND categoria_id > 0`); err != nil {
		return err
	}
	return nil
}

func resolveCategoriaProductoTx(tx *sql.Tx, empresaID, categoriaID int64) (string, error) {
	var nombre string
	if err := tx.QueryRow(`SELECT nombre FROM categorias_productos WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, categoriaID).Scan(&nombre); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("categoria %d no pertenece a la empresa %d", categoriaID, empresaID)
		}
		return "", err
	}
	nombre = strings.TrimSpace(nombre)
	if nombre == "" {
		return "", fmt.Errorf("la categoría seleccionada no tiene nombre")
	}
	return nombre, nil
}

func insertProductoPrecioHistorialTx(tx *sql.Tx, h ProductoPrecioHistorial) error {
	_, err := tx.Exec(`INSERT INTO producto_precios_historial (
		empresa_id, producto_id, costo_anterior, costo_nuevo, precio_anterior, precio_nuevo, impuesto_anterior, impuesto_nuevo,
		motivo, referencia, fecha_cambio, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'), datetime('now','localtime'), ?, COALESCE(NULLIF(?, ''), 'activo'), ?)`,
		h.EmpresaID, h.ProductoID, h.CostoAnterior, h.CostoNuevo, h.PrecioAnterior, h.PrecioNuevo, h.ImpuestoAnterior, h.ImpuestoNuevo,
		strings.TrimSpace(h.Motivo), strings.TrimSpace(h.Referencia), strings.TrimSpace(h.UsuarioCreador), strings.TrimSpace(h.Estado), strings.TrimSpace(h.Observaciones))
	return err
}

func getInventarioPoliticaCostoTx(tx *sql.Tx, empresaID int64) (string, error) {
	var politica sql.NullString
	err := tx.QueryRow(`SELECT politica_costo FROM empresa_inventario_configuracion WHERE empresa_id = ? LIMIT 1`, empresaID).Scan(&politica)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return inventarioPoliticaCostoPromedio, nil
		}
		return "", err
	}
	return normalizeInventarioPoliticaCosto(politica.String), nil
}

func calcularCostoPromedioDisponibleTx(tx *sql.Tx, empresaID, productoID, bodegaID int64) (float64, error) {
	query := `SELECT COALESCE(SUM(cantidad_disponible * costo_unitario), 0), COALESCE(SUM(cantidad_disponible), 0)
		FROM inventario_costos_lotes
		WHERE empresa_id = ?
			AND producto_id = ?
			AND LOWER(COALESCE(estado, 'activo')) = 'activo'
			AND COALESCE(cantidad_disponible, 0) > 0`
	args := []interface{}{empresaID, productoID}
	if bodegaID > 0 {
		query += ` AND bodega_id = ?`
		args = append(args, bodegaID)
	}

	var totalCosto float64
	var totalCantidad float64
	if err := tx.QueryRow(query, args...).Scan(&totalCosto, &totalCantidad); err != nil {
		return 0, err
	}
	if totalCantidad <= 0 {
		return 0, nil
	}
	return round2(totalCosto / totalCantidad), nil
}

func recalcularCostoPromedioProductoTx(tx *sql.Tx, empresaID, productoID int64) error {
	promedio, err := calcularCostoPromedioDisponibleTx(tx, empresaID, productoID, 0)
	if err != nil {
		return err
	}
	if promedio <= 0 {
		return nil
	}
	_, err = tx.Exec(`UPDATE productos
		SET costo = ?, fecha_actualizacion = datetime('now','localtime')
		WHERE empresa_id = ? AND id = ?`, promedio, empresaID, productoID)
	return err
}

func ensureCostoLotesSeedTx(tx *sql.Tx, empresaID, productoID, bodegaID int64, usuario string) error {
	var existenciaActual float64
	err := tx.QueryRow(`SELECT COALESCE(cantidad, 0)
		FROM inventario_existencias
		WHERE empresa_id = ? AND producto_id = ? AND bodega_id = ?
		LIMIT 1`, empresaID, productoID, bodegaID).Scan(&existenciaActual)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	if existenciaActual <= 0 {
		return nil
	}

	var lotesActuales float64
	if err := tx.QueryRow(`SELECT COALESCE(SUM(cantidad_disponible), 0)
		FROM inventario_costos_lotes
		WHERE empresa_id = ?
			AND producto_id = ?
			AND bodega_id = ?
			AND LOWER(COALESCE(estado, 'activo')) = 'activo'`, empresaID, productoID, bodegaID).Scan(&lotesActuales); err != nil {
		return err
	}

	deltaSeed := round2(existenciaActual - lotesActuales)
	if deltaSeed <= 0 {
		return nil
	}

	var costoBase float64
	if err := tx.QueryRow(`SELECT COALESCE(costo, 0) FROM productos WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, productoID).Scan(&costoBase); err != nil {
		return err
	}
	if costoBase < 0 {
		costoBase = 0
	}

	return registerCostoLoteTx(tx, empresaID, productoID, bodegaID, deltaSeed, costoBase, "SEED-STOCK-LEGACY", usuario, "regularizacion de costos para stock historico")
}

func registerCostoLoteTx(tx *sql.Tx, empresaID, productoID, bodegaID int64, cantidad, costoUnitario float64, referencia, usuario, observaciones string) error {
	if cantidad <= 0 {
		return nil
	}
	if costoUnitario < 0 {
		costoUnitario = 0
	}
	_, err := tx.Exec(`INSERT INTO inventario_costos_lotes (
		empresa_id,
		producto_id,
		bodega_id,
		cantidad_disponible,
		costo_unitario,
		referencia,
		fecha_lote,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'), datetime('now','localtime'), ?, 'activo', ?)`,
		empresaID,
		productoID,
		bodegaID,
		cantidad,
		round2(costoUnitario),
		strings.TrimSpace(referencia),
		strings.TrimSpace(usuario),
		strings.TrimSpace(observaciones),
	)
	return err
}

func consumirCostoLotesPEPSTx(tx *sql.Tx, empresaID, productoID, bodegaID int64, cantidad float64) (float64, error) {
	if cantidad <= 0 {
		return 0, nil
	}

	rows, err := tx.Query(`SELECT id, COALESCE(cantidad_disponible, 0), COALESCE(costo_unitario, 0)
		FROM inventario_costos_lotes
		WHERE empresa_id = ?
			AND producto_id = ?
			AND bodega_id = ?
			AND LOWER(COALESCE(estado, 'activo')) = 'activo'
			AND COALESCE(cantidad_disponible, 0) > 0
		ORDER BY datetime(COALESCE(fecha_lote, fecha_creacion)) ASC, id ASC`, empresaID, productoID, bodegaID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type loteConsumo struct {
		id         int64
		cantidad   float64
		costo      float64
		consumir   float64
		saldoFinal float64
		debeCerrar bool
	}
	consumos := make([]loteConsumo, 0)
	restante := cantidad
	totalCosto := 0.0

	for rows.Next() {
		if restante <= 0 {
			break
		}
		var lote loteConsumo
		if err := rows.Scan(&lote.id, &lote.cantidad, &lote.costo); err != nil {
			return 0, err
		}
		if lote.cantidad <= 0 {
			continue
		}

		lote.consumir = restante
		if lote.consumir > lote.cantidad {
			lote.consumir = lote.cantidad
		}
		lote.saldoFinal = round2(lote.cantidad - lote.consumir)
		lote.debeCerrar = lote.saldoFinal <= 0

		restante = round2(restante - lote.consumir)
		totalCosto += lote.consumir * lote.costo
		consumos = append(consumos, lote)
	}

	if restante > 0 {
		return 0, ErrStockInsuficiente
	}

	for _, lote := range consumos {
		if lote.debeCerrar {
			if _, err := tx.Exec(`UPDATE inventario_costos_lotes
				SET cantidad_disponible = 0,
					estado = 'consumido',
					fecha_actualizacion = datetime('now','localtime')
				WHERE id = ?`, lote.id); err != nil {
				return 0, err
			}
			continue
		}
		if _, err := tx.Exec(`UPDATE inventario_costos_lotes
			SET cantidad_disponible = ?, fecha_actualizacion = datetime('now','localtime')
			WHERE id = ?`, lote.saldoFinal, lote.id); err != nil {
			return 0, err
		}
	}

	if cantidad <= 0 {
		return 0, nil
	}
	return round2(totalCosto / cantidad), nil
}

func transferirCostoLotesEntreBodegasTx(tx *sql.Tx, empresaID, productoID, bodegaOrigenID, bodegaDestinoID int64, cantidad float64, referencia, usuario string) (float64, error) {
	if err := ensureCostoLotesSeedTx(tx, empresaID, productoID, bodegaOrigenID, usuario); err != nil {
		return 0, err
	}
	costoTransferido, err := consumirCostoLotesPEPSTx(tx, empresaID, productoID, bodegaOrigenID, cantidad)
	if err != nil {
		return 0, err
	}
	if costoTransferido < 0 {
		costoTransferido = 0
	}
	if err := registerCostoLoteTx(tx, empresaID, productoID, bodegaDestinoID, cantidad, costoTransferido, referencia, usuario, "traslado recibido entre bodegas"); err != nil {
		return 0, err
	}
	return costoTransferido, nil
}

func validateProductoEmpresaTx(tx *sql.Tx, empresaID, productoID int64) error {
	var exists int
	if err := tx.QueryRow(`SELECT 1 FROM productos WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, productoID).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("producto %d no pertenece a la empresa %d", productoID, empresaID)
		}
		return err
	}
	return nil
}

func validateBodegaEmpresaTx(tx *sql.Tx, empresaID, bodegaID int64) error {
	var exists int
	if err := tx.QueryRow("SELECT 1 FROM bodegas WHERE empresa_id = ? AND id = ? LIMIT 1", empresaID, bodegaID).Scan(&exists); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("bodega %d no pertenece a la empresa %d", bodegaID, empresaID)
		}
		return err
	}
	return nil
}

func validateProveedorEmpresaTx(tx *sql.Tx, empresaID, proveedorID int64) error {
	var exists int
	if err := tx.QueryRow("SELECT 1 FROM proveedores WHERE empresa_id = ? AND id = ? LIMIT 1", empresaID, proveedorID).Scan(&exists); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("proveedor %d no pertenece a la empresa %d", proveedorID, empresaID)
		}
		return err
	}
	return nil
}

func upsertExistenciaTx(tx *sql.Tx, empresaID, productoID, bodegaID int64, delta float64, usuario, observaciones string) error {
	res, err := tx.Exec(`UPDATE inventario_existencias
		SET cantidad = cantidad + ?, fecha_actualizacion = datetime('now','localtime')
		WHERE empresa_id = ? AND producto_id = ? AND bodega_id = ?`,
		delta, empresaID, productoID, bodegaID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected > 0 {
		return nil
	}
	_, err = tx.Exec(`INSERT INTO inventario_existencias (
		empresa_id, producto_id, bodega_id, cantidad, usuario_creador, estado, observaciones, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, 'activo', ?, datetime('now','localtime'), datetime('now','localtime'))`,
		empresaID, productoID, bodegaID, delta, strings.TrimSpace(usuario), strings.TrimSpace(observaciones))
	return err
}

func insertMovimientoTx(tx *sql.Tx, m InventarioMovimiento) error {
	_, err := tx.Exec(`INSERT INTO inventario_movimientos (
		empresa_id, producto_id, bodega_origen_id, bodega_destino_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento,
		usuario_creador, estado, observaciones, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), ?, COALESCE(NULLIF(?, ''), 'activo'), ?, datetime('now','localtime'), datetime('now','localtime'))`,
		m.EmpresaID, m.ProductoID, nullableInt64(m.BodegaOrigenID), nullableInt64(m.BodegaDestinoID), strings.TrimSpace(m.Tipo), m.Cantidad, m.CostoUnitario, strings.TrimSpace(m.Referencia), strings.TrimSpace(m.UsuarioCreador), strings.TrimSpace(m.Estado), strings.TrimSpace(m.Observaciones))
	return err
}

func nullableInt64(v int64) interface{} {
	if v <= 0 {
		return nil
	}
	return v
}

func defaultUnidad(u string) string {
	u = strings.TrimSpace(u)
	if u == "" {
		return "unidad"
	}
	return u
}
