package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	paginaPrincipalConfigKey          = "super.pagina_principal.cards.v1"
	paginaPrincipalConfigUpdatedByKey = "super.pagina_principal.cards.v1.updated_by"
	informacionModulosConfigKey       = "super.informacion_modulos.v1"
	informacionModulosUpdatedByKey    = "super.informacion_modulos.v1.updated_by"
	noticiasPortalConfigKey           = "super.noticias_portal.v1"
	noticiasPortalUpdatedByKey        = "super.noticias_portal.v1.updated_by"
	paginaPrincipalDefaultCardLimit   = 12
	informacionModulosDefaultLimit    = 24
	informacionModulosFeatureLimit    = 32
	noticiasPortalDefaultLimit        = 40
)

const (
	paginaPrincipalVisualSizeSmall  = "pequeno"
	paginaPrincipalVisualSizeMedium = "mediano"
	paginaPrincipalVisualSizeLarge  = "grande"
)

const (
	paginaPrincipalCardTypeInfoPhoto = "info_foto"
	paginaPrincipalCardTypeBanner    = "banner"
)

type paginaPrincipalCard struct {
	TipoTarjeta       string   `json:"tipo_tarjeta,omitempty"`
	Titulo            string   `json:"titulo"`
	Descripcion       string   `json:"descripcion"`
	ImagenURL         string   `json:"imagen_url"`
	ImagenSecundaria  string   `json:"imagen_secundaria_url,omitempty"`
	Enlace            string   `json:"enlace"`
	YouTubeURL        string   `json:"youtube_url,omitempty"`
	DetalleEtiqueta   string   `json:"detalle_etiqueta"`
	DetalleTitular    string   `json:"detalle_titular"`
	DetalleParrafoUno string   `json:"detalle_parrafo_uno"`
	DetalleParrafoDos string   `json:"detalle_parrafo_dos"`
	DetallePuntos     []string `json:"detalle_puntos"`
}

type paginaPrincipalBannerCard struct {
	ImagenURL string `json:"imagen_url"`
	Enlace    string `json:"enlace,omitempty"`
}

type paginaPrincipalVisualSettings struct {
	IndexCardSize   string `json:"index_card_size"`
	IndexTextSize   string `json:"index_text_size"`
	LandingCardSize string `json:"landing_card_size"`
	LandingTextSize string `json:"landing_text_size"`
}

type paginaPrincipalConfig struct {
	Cantidad int                           `json:"cantidad"`
	Tarjetas []paginaPrincipalCard         `json:"tarjetas"`
	Estilos  paginaPrincipalVisualSettings `json:"estilos"`
}

type informacionModuloItem struct {
	Titulo          string   `json:"titulo"`
	IconoURL        string   `json:"icono_url"`
	Caracteristicas []string `json:"caracteristicas"`
}

type informacionModulosConfig struct {
	Titulo  string                  `json:"titulo"`
	Modulos []informacionModuloItem `json:"modulos"`
}

type noticiaPortalItem struct {
	Titulo       string   `json:"titulo"`
	Resumen      string   `json:"resumen"`
	Contenido    string   `json:"contenido"`
	Categoria    string   `json:"categoria"`
	Fecha        string   `json:"fecha"`
	ImagenURL    string   `json:"imagen_url"`
	FuenteNombre string   `json:"fuente_nombre"`
	FuenteURL    string   `json:"fuente_url"`
	Etiquetas    []string `json:"etiquetas"`
	Destacada    bool     `json:"destacada"`
	Activa       bool     `json:"activa"`
}

type noticiasPortalConfig struct {
	Titulo       string              `json:"titulo"`
	Subtitulo    string              `json:"subtitulo"`
	PortadaURL   string              `json:"portada_url"`
	PerfilURL    string              `json:"perfil_url"`
	NombrePagina string              `json:"nombre_pagina"`
	Usuario      string              `json:"usuario"`
	Descripcion  string              `json:"descripcion"`
	Noticias     []noticiaPortalItem `json:"noticias"`
}

func paginaPrincipalDefaultWhatsAppContactNumber() string {
	return "573043306506"
}

func normalizePortalWhatsAppContactNumber(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	builder := strings.Builder{}
	for idx, r := range trimmed {
		if r >= '0' && r <= '9' {
			builder.WriteRune(r)
			continue
		}
		if r == '+' && idx == 0 {
			continue
		}
	}
	normalized := builder.String()
	if len(normalized) < 10 || len(normalized) > 15 {
		return ""
	}
	return normalized
}

func paginaPrincipalLoadWhatsAppContactNumber(dbSuper *sql.DB) string {
	configured, _, _, _, err := dbpkg.GetConfigEntry(dbSuper, "portal.whatsapp_contact_number")
	if err != nil {
		return paginaPrincipalDefaultWhatsAppContactNumber()
	}
	normalized := normalizePortalWhatsAppContactNumber(configured)
	if normalized == "" {
		return paginaPrincipalDefaultWhatsAppContactNumber()
	}
	return normalized
}

func paginaPrincipalDefaultVisualSettings() paginaPrincipalVisualSettings {
	return paginaPrincipalVisualSettings{
		IndexCardSize:   paginaPrincipalVisualSizeMedium,
		IndexTextSize:   paginaPrincipalVisualSizeMedium,
		LandingCardSize: paginaPrincipalVisualSizeMedium,
		LandingTextSize: paginaPrincipalVisualSizeMedium,
	}
}

func paginaPrincipalDefaultConfig() paginaPrincipalConfig {
	cards := []paginaPrincipalCard{
		{
			TipoTarjeta:       paginaPrincipalCardTypeInfoPhoto,
			Titulo:            "Punto de venta",
			Descripcion:       "Solucion completa para ventas rapidas y facturacion electronica.",
			ImagenURL:         "/img/punto_venta.png",
			ImagenSecundaria:  "/img/sistema punto de venta.png",
			Enlace:            "/administrar_empresa.html?module=punto_venta",
			DetalleEtiqueta:   "Retail y mostrador",
			DetalleTitular:    "Vende rapido, factura mejor y controla la caja desde una sola pantalla.",
			DetalleParrafoUno: "La solucion de Punto de Venta esta pensada para negocios que necesitan registrar ventas agiles sin perder trazabilidad. Cada movimiento puede quedar asociado a empresa, usuario, estacion, cliente, metodo de pago y documento emitido, lo que facilita operar mostrador, caja rapida o atencion general desde una misma interfaz.",
			DetalleParrafoDos: "Ademas de cobrar, el sistema ayuda a controlar inventario, descuentos, precios, clientes frecuentes, cierres de caja y reportes operativos. Esto permite que el comercio no solo venda mas rapido, sino que tambien mantenga orden financiero y visibilidad real sobre lo que esta ocurriendo en el negocio.",
			DetallePuntos: []string{
				"Facturacion electronica y documentos de venta con trazabilidad por empresa.",
				"Carritos de compra rapidos con productos, servicios, recetas y descuentos controlados.",
				"Cierres de caja, metodos de pago y conciliacion operativa para supervision diaria.",
				"Inventario sincronizado, historial de ventas y reportes para decisiones comerciales.",
			},
		},
		{
			TipoTarjeta:       paginaPrincipalCardTypeInfoPhoto,
			Titulo:            "Motel",
			Descripcion:       "Gestion por tiempo de servicio y facturacion tarifada por estancia.",
			ImagenURL:         "/img/motel.png",
			ImagenSecundaria:  "/img/sistema punto de venta.png",
			Enlace:            "/administrar_empresa.html?module=motel",
			DetalleEtiqueta:   "Operacion por estancias",
			DetalleTitular:    "Controla estaciones, tiempos de ocupacion y consumos sin perder detalle.",
			DetalleParrafoUno: "El sistema para Motel esta orientado a negocios donde el cobro depende del tiempo de uso, la disponibilidad de estaciones y los consumos agregados durante la estancia. Permite conocer que estaciones estan libres, ocupadas, reservadas o listas para limpieza, facilitando la operacion en tiempo real.",
			DetalleParrafoDos: "La plataforma combina tarifas por minutos, tarifas por bloques, cargos adicionales, consumos de minibar o servicios y seguimiento de pagos por estacion. Esto mejora la rotacion, reduce errores de cobro y entrega una trazabilidad clara para auditoria interna y control administrativo.",
			DetallePuntos: []string{
				"Tarifas por tiempo, reglas por bloques y calculo automatico del valor a cobrar.",
				"Control de estaciones con estados operativos y consumos asociados a la estancia.",
				"Carritos simultaneos por estacion con aislamiento por empresa.",
				"Reportes de ocupacion, ingresos por turno y seguimiento detallado de servicios.",
			},
		},
		{
			TipoTarjeta:       paginaPrincipalCardTypeInfoPhoto,
			Titulo:            "Restaurante",
			Descripcion:       "Gestion de mesas, pedidos y facturacion para restaurantes.",
			ImagenURL:         "/img/restaurante.png",
			ImagenSecundaria:  "/img/sistema punto de venta.png",
			Enlace:            "/administrar_empresa.html?module=restaurante",
			DetalleEtiqueta:   "Mesas y cocina",
			DetalleTitular:    "Administra mesas, pedidos, cocina y cobro final con flujo continuo.",
			DetalleParrafoUno: "La solucion para Restaurante ayuda a organizar la atencion desde que el cliente se sienta hasta que se factura la cuenta. El personal puede trabajar por mesas, estaciones o usuarios, tomar pedidos rapidamente y mantener control sobre productos, tiempos de despacho y consumos acumulados.",
			DetalleParrafoDos: "Tambien facilita dividir cuentas, manejar propinas, emitir facturas, enviar ordenes a cocina o barra y consultar reportes por turno. Con esto, el restaurante gana velocidad en servicio, reduce reprocesos y mejora la coordinacion entre salon, caja y produccion.",
			DetallePuntos: []string{
				"Gestion de mesas, pedidos abiertos y consumos acumulados por cliente o grupo.",
				"Impresion o resolucion de comandos para cocina, barra o puntos de preparacion.",
				"Propinas, descuentos y metodos de pago integrados al cierre de la cuenta.",
				"Indicadores de ventas, rotacion de mesas y control operativo por turno.",
			},
		},
		{
			TipoTarjeta:       paginaPrincipalCardTypeInfoPhoto,
			Titulo:            "Control por sensor",
			Descripcion:       "Integracion y alertas con sensores para control de accesos.",
			ImagenURL:         "/img/sensor.png",
			ImagenSecundaria:  "/img/sistema punto de venta.png",
			Enlace:            "/administrar_empresa.html?module=sensor",
			DetalleEtiqueta:   "Monitoreo y automatizacion",
			DetalleTitular:    "Conecta eventos fisicos con alertas, estados y control operativo centralizado.",
			DetalleParrafoUno: "La solucion de Control por Sensor esta diseñada para operaciones donde un evento fisico debe producir una accion o una evidencia digital. Puede servir para accesos, aperturas, cierres, confirmaciones de paso, sensores de puerta o estados que deban registrarse automaticamente para soporte operativo o seguridad.",
			DetalleParrafoDos: "En lugar de depender solo de verificaciones manuales, el sistema centraliza senales, alertas y trazabilidad para que supervisores y administradores sepan que ocurrio, cuando ocurrio y en que punto operativo sucedio. Esto mejora la reaccion, la auditoria y la consistencia de la operacion diaria.",
			DetallePuntos: []string{
				"Registro de eventos de sensores con relacion a estaciones o puntos de control.",
				"Alertas operativas y visibilidad de estados recientes para soporte inmediato.",
				"Trazabilidad de accesos, aperturas o incidencias con evidencia temporal.",
				"Integracion con flujos operativos que requieren validacion fisica o automatizada.",
			},
		},
		{
			TipoTarjeta:       paginaPrincipalCardTypeInfoPhoto,
			Titulo:            "Hotel",
			Descripcion:       "Reservas, ocupacion, tarifas por dia, check-in, servicios, estaciones y reportes ejecutivos.",
			ImagenURL:         "/img/hotel-logo.svg",
			ImagenSecundaria:  "/img/sistema punto de venta.png",
			Enlace:            "/administrar_empresa.html?module=configuracion",
			DetalleEtiqueta:   "Reservas y hospedaje",
			DetalleTitular:    "Gestiona reservas, check-in, check-out y facturacion hotelera en un mismo flujo.",
			DetalleParrafoUno: "La solucion para Hotel permite administrar estaciones, reservas futuras, ocupacion actual y cargos adicionales dentro de una operacion unificada. De esta forma, recepcion y administracion pueden trabajar con informacion consistente sobre disponibilidad, tiempos de entrada y salida y consumos por huesped.",
			DetalleParrafoDos: "El sistema ayuda a estructurar el ciclo completo del hospedaje: reserva, confirmacion, asignacion de estacion, facturacion, cargos por servicios y control posterior. Esto lo convierte en una herramienta util tanto para pequeños hoteles como para operaciones que necesitan mayor orden en caja, reportes y servicio al cliente.",
			DetallePuntos: []string{
				"Control de reservas por rango de fechas y disponibilidad real de estaciones.",
				"Check-in y check-out con cargos diarios, servicios adicionales y seguimiento por huesped.",
				"Facturacion, reportes de ocupacion y trazabilidad por empresa y periodo.",
				"Soporte para operaciones multiusuario con historial claro de movimientos y cobros.",
			},
		},
		{
			TipoTarjeta:      paginaPrincipalCardTypeInfoPhoto,
			Titulo:           "Clientes y CRM",
			Descripcion:      "Leads, clientes, historial comercial, pipeline, campanas, seguimientos, cotizaciones y conversion a ventas centrales por empresa.",
			ImagenURL:        "/img/customer.svg",
			ImagenSecundaria: "/img/analytics-color.svg",
			Enlace:           "/administrar_empresa.html?module=crm_unificado",
		},
		{
			TipoTarjeta:      paginaPrincipalCardTypeInfoPhoto,
			Titulo:           "Drogueria y farmacia",
			Descripcion:      "Plantilla sanitaria integrada al nucleo: lotes, INVIMA, vencimientos, formulas, dispensacion, inventario, ventas, pagos y facturacion.",
			ImagenURL:        "/img/shield-license-color.svg",
			ImagenSecundaria: "/img/report.svg",
			Enlace:           "/administrar_empresa.html?module=drogueria_farmacia",
		},
		{
			TipoTarjeta:      paginaPrincipalCardTypeInfoPhoto,
			Titulo:           "Alquileres de activos",
			Descripcion:      "Contratos, garantias, checklist, devoluciones, mantenimiento y venta central conectada a clientes y servicios.",
			ImagenURL:        "/img/company-briefcase-color.svg",
			ImagenSecundaria: "/img/warehouse-color.svg",
			Enlace:           "/administrar_empresa.html?module=alquileres",
		},
		{
			TipoTarjeta:      paginaPrincipalCardTypeInfoPhoto,
			Titulo:           "Logistica WMS",
			Descripcion:      "Ubicaciones internas, conteos, picking, packing, despachos, rutas y bitacora conectados al inventario central.",
			ImagenURL:        "/img/warehouse-color.svg",
			ImagenSecundaria: "/img/analytics-color.svg",
			Enlace:           "/administrar_empresa.html?module=logistica_wms",
		},
		{
			TipoTarjeta:      paginaPrincipalCardTypeInfoPhoto,
			Titulo:           "Bancos y pagos masivos",
			Descripcion:      "Cuentas bancarias, conciliacion, extractos, lotes de pagos, aprobaciones, caja y tesoreria conectados al centro financiero.",
			ImagenURL:        "/img/money.svg",
			ImagenSecundaria: "/img/report.svg",
			Enlace:           "/administrar_empresa.html?module=bancos_pagos",
		},
		{
			TipoTarjeta:      paginaPrincipalCardTypeInfoPhoto,
			Titulo:           "Gestion documental y contratos",
			Descripcion:      "Expedientes, documentos, contratos, obligaciones, vencimientos, firma externa/manual, aprobaciones y trazabilidad por empresa.",
			ImagenURL:        "/img/report.svg",
			ImagenSecundaria: "/img/company-briefcase-color.svg",
			Enlace:           "/administrar_empresa.html?module=gestion_documental",
		},
		{
			TipoTarjeta:      paginaPrincipalCardTypeInfoPhoto,
			Titulo:           "Facturacion electronica Colombia",
			Descripcion:      "Factura electronica, notas, documento soporte, nomina electronica, POS electronico, impuestos Colombia y trazabilidad DIAN.",
			ImagenURL:        "/img/invoice.svg",
			ImagenSecundaria: "/img/taxes.svg",
			Enlace:           "/administrar_empresa.html?module=facturacion",
		},
	}
	return paginaPrincipalConfig{
		Cantidad: len(cards),
		Tarjetas: cards,
		Estilos:  paginaPrincipalDefaultVisualSettings(),
	}
}

func informacionModulosDefaultConfig() informacionModulosConfig {
	return informacionModulosConfig{
		Titulo: "Modulos y caracteristicas principales",
		Modulos: []informacionModuloItem{
			{Titulo: "Inventario profesional", IconoURL: "/img/warehouse-color.svg", Caracteristicas: []string{"Productos", "Servicios", "Recetas", "Categorias", "Bodegas", "Kardex", "Traslados", "Compras", "Proveedores", "Control de existencias"}},
			{Titulo: "Ventas POS", IconoURL: "/img/punto_venta.png", Caracteristicas: []string{"Venta directa", "Carritos por estacion", "Pagos mixtos", "Abonos", "Descuentos", "Codigos promocionales", "Caja por usuario", "Varias cajas simultaneas"}},
			{Titulo: "Pagos y hardware", IconoURL: "/img/money.svg", Caracteristicas: []string{"Manejo de datafonos", "Pagos QR", "Cajon monedero", "Impresoras POS", "Impresoras por area o producto", "Facturacion offline"}},
			{Titulo: "Documentos electronicos", IconoURL: "/img/invoice.svg", Caracteristicas: []string{
				"Documentos electronicos DIAN Colombia",
				"Activa los documentos del Sistema de Facturacion Electronica que la empresa usara. Las obligaciones de contador se separan para no mezclarlas con documentos UBL de venta.",
				"Documentos y eventos del SFE",
				"Factura electronica de venta - Venta: venta de bienes o servicios validada previamente por DIAN.",
				"Nota credito electronica - Ajustes de venta: disminuye, corrige o reversa valores de una factura electronica.",
				"Nota debito electronica - Ajustes de venta: aumenta o corrige valores de una factura electronica.",
				"Reporte de factura de talonario o papel por contingencia - Contingencia: reporte para validacion posterior cuando hubo inconveniente tecnologico del facturador.",
				"Documento soporte en adquisiciones a no obligados - Compras: soporta costos, deducciones o impuestos descontables en compras a sujetos no obligados a facturar.",
				"Nota de ajuste del documento soporte - Compras: ajusta o corrige un documento soporte de adquisiciones.",
			}},
			{Titulo: "Finanzas y cumplimiento", IconoURL: "/img/taxes.svg", Caracteristicas: []string{"Impuestos", "Bancos", "Ingresos", "Egresos", "Tesoreria", "Presupuesto", "Reportes", "Modulo del contador", "Certificados tributarios", "Informacion exogena"}},
			{Titulo: "Operacion por estaciones", IconoURL: "/img/hotel-logo.svg", Caracteristicas: []string{"Estaciones", "Mesas", "Habitaciones", "Zonas o bahias configurables", "Control de estados", "Turnos", "Reservas", "Alertas de tiempo", "Aseo", "Cierre/corte de caja"}},
			{Titulo: "Automatizacion e IA", IconoURL: "/img/gpt.svg", Caracteristicas: []string{"Integracion con IA", "Documentos inteligentes", "OCR de compras", "Soporte operativo", "Reportes asistidos", "Acciones confirmables"}},
			{Titulo: "GRAFOLOGIX", IconoURL: "/img/analytics-color.svg", Caracteristicas: []string{"Carga de manuscritos", "OCR libre con Tesseract", "Medidas de inclinacion", "Espacios entre palabras y letras", "Metricas de margenes", "Reporte PDF", "Exportacion HTML, JSON, CSV y TXT", "Analisis complementario con IA configurada"}},
			{Titulo: "Energia solar", IconoURL: "/img/solar-energy.svg", Caracteristicas: []string{"Monitoreo de paneles", "Controladoras Victron", "SMA Sunny Portal", "SolarEdge Monitoring", "Baterias Powerwall, BYD, Pylontech y Enphase", "Alertas por correo", "Lecturas por gateway local", "BMS y estado de salud"}},
			{Titulo: "Camaras y DVR", IconoURL: "/img/shield-security-color.svg", Caracteristicas: []string{"Registro de camaras por empresa", "DVR/NVR por canal", "Visores HLS, WebRTC, MJPEG o iframe", "Estaciones tipo camara", "Carga antes o despues de estaciones", "Acceso por permisos", "Monitoreo operativo"}},
			{Titulo: "Domotica y control fisico", IconoURL: "/img/sensor.png", Caracteristicas: []string{"Domotica por estacion", "Manejo de sensores", "Puertas", "Aparatos", "Permanencia", "Acceso", "Vehiculos", "Parqueaderos", "Trazabilidad operativa"}},
			{Titulo: "Gestion empresarial", IconoURL: "/img/company-briefcase-color.svg", Caracteristicas: []string{"Clientes", "CRM", "Usuarios", "Roles", "Permisos", "Licencias", "Auditoria", "Backups", "Comunicaciones", "Soporte", "Chat y tareas"}},
			{Titulo: "Plantillas listas", IconoURL: "/img/analytics-color.svg", Caracteristicas: []string{"Hotel", "Motel", "Restaurante", "Gimnasio", "Odontologia", "Propiedad horizontal", "Domicilios", "Taxi", "Carta QR", "Venta publica", "Red social comercial"}},
		},
	}
}

func noticiasPortalDefaultConfig() noticiasPortalConfig {
	return noticiasPortalConfig{
		Titulo:       "Noticias",
		Subtitulo:    "Actualidad tributaria, facturacion electronica y novedades de Powerful Control System.",
		PortadaURL:   "/img/sistema punto de venta.png",
		PerfilURL:    "/img/pwa-icon-192.png",
		NombrePagina: "Powerful Control System Noticias",
		Usuario:      "@powerfulcontrolsystem",
		Descripcion:  "Pagina informativa para seguir cambios de facturacion electronica, impuestos, tecnologia POS y operacion empresarial.",
		Noticias: []noticiaPortalItem{
			{
				Titulo:       "DIAN publica doctrina 2026 sobre obligacion de facturar y documentos equivalentes",
				Resumen:      "La DIAN reitero en doctrina reciente que la obligacion formal de facturar aplica cuando se venden bienes o se prestan servicios, aun cuando existan situaciones especiales sobre la calidad tributaria del sujeto.",
				Contenido:    "El Concepto 005716 int. 514 de 2026 recuerda que la factura electronica de venta con validacion previa hace parte del sistema de facturacion colombiano y que la obligacion de expedir factura o documento equivalente debe revisarse frente a las reglas vigentes de la DIAN. Para empresas que operan POS, esta noticia refuerza la importancia de conservar configuracion de numeracion, cliente, impuestos, documentos electronicos y soportes al dia.",
				Categoria:    "DIAN / Facturacion electronica",
				Fecha:        "2026-04-14",
				ImagenURL:    "/img/invoice.svg",
				FuenteNombre: "DIAN - Concepto 005716 int. 514 de 2026",
				FuenteURL:    "https://www.dian.gov.co/Contribuyentes-Plus/Documents/21-CONCEPTO-005716-int-514-14042026.pdf",
				Etiquetas:    []string{"DIAN", "Factura electronica", "Documento equivalente", "Colombia"},
				Destacada:    true,
				Activa:       true,
			},
			{
				Titulo:       "DIAN fortalece controles sobre expedicion de factura electronica",
				Resumen:      "La entidad anuncio seguimiento a establecimientos con irregularidades en facturacion electronica y revision de soportes operativos.",
				Contenido:    "La DIAN informo una estrategia de verificacion y seguimiento a obligados a facturar electronicamente. Para los negocios, esto confirma la necesidad de operar con trazabilidad de ventas, egresos, proveedores, soportes e informes por turno, especialmente cuando hay varias cajas y usuarios simultaneos.",
				Categoria:    "DIAN / Control fiscal",
				Fecha:        "2025-02-13",
				ImagenURL:    "/img/shield-license-color.svg",
				FuenteNombre: "DIAN - Comunicado de Prensa No. 014",
				FuenteURL:    "https://www.dian.gov.co/scripts/HPublishing?id=3813&type=NewsPdf",
				Etiquetas:    []string{"Control", "Facturacion electronica", "Soportes", "Fiscalizacion"},
				Destacada:    false,
				Activa:       true,
			},
		},
	}
}

func paginaPrincipalNormalizeVisualSize(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "small", "pequeno":
		return paginaPrincipalVisualSizeSmall
	case "medium", "mediano":
		return paginaPrincipalVisualSizeMedium
	case "large", "grande":
		return paginaPrincipalVisualSizeLarge
	default:
		return paginaPrincipalVisualSizeMedium
	}
}

func paginaPrincipalNormalizeVisualSettings(raw paginaPrincipalVisualSettings) paginaPrincipalVisualSettings {
	defaults := paginaPrincipalDefaultVisualSettings()
	indexCardSize := paginaPrincipalNormalizeVisualSize(raw.IndexCardSize)
	if indexCardSize == "" {
		indexCardSize = defaults.IndexCardSize
	}
	indexTextSize := paginaPrincipalNormalizeVisualSize(raw.IndexTextSize)
	if indexTextSize == "" {
		indexTextSize = defaults.IndexTextSize
	}
	landingCardSize := paginaPrincipalNormalizeVisualSize(raw.LandingCardSize)
	if landingCardSize == "" {
		landingCardSize = defaults.LandingCardSize
	}
	landingTextSize := paginaPrincipalNormalizeVisualSize(raw.LandingTextSize)
	if landingTextSize == "" {
		landingTextSize = defaults.LandingTextSize
	}
	return paginaPrincipalVisualSettings{
		IndexCardSize:   indexCardSize,
		IndexTextSize:   indexTextSize,
		LandingCardSize: landingCardSize,
		LandingTextSize: landingTextSize,
	}
}

func paginaPrincipalNormalizeCardType(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "banner", "baner", "tarjeta_banner", "tarjeta_baner":
		return paginaPrincipalCardTypeBanner
	case "info_foto", "informacion_foto", "info_photo", "tarjeta_informacion_mas_foto":
		return paginaPrincipalCardTypeInfoPhoto
	default:
		return paginaPrincipalCardTypeInfoPhoto
	}
}

func paginaPrincipalNormalizeImageURL(raw, fallback string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		value = strings.TrimSpace(fallback)
	}
	if value == "" {
		value = "/img/punto_venta.png"
	}
	if strings.Contains(value, "..") {
		return "/img/punto_venta.png"
	}
	if !strings.HasPrefix(value, "/img/") {
		return "/img/punto_venta.png"
	}
	return value
}

func paginaPrincipalNormalizeBannerCards(raw []paginaPrincipalBannerCard) []paginaPrincipalBannerCard {
	source := raw
	if len(source) == 0 {
		return []paginaPrincipalBannerCard{}
	}
	out := make([]paginaPrincipalBannerCard, 0, len(source))
	for _, it := range source {
		img := paginaPrincipalNormalizeImageURL(it.ImagenURL, "/img/baner_ia.png")
		out = append(out, paginaPrincipalBannerCard{
			ImagenURL: img,
			Enlace:    paginaPrincipalNormalizeLink(it.Enlace, ""),
		})
	}
	return out
}

func paginaPrincipalNormalizeLink(raw, fallback string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		value = strings.TrimSpace(fallback)
	}
	if value == "" {
		return "/login.html"
	}
	lower := strings.ToLower(value)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return value
	}
	if strings.HasPrefix(value, "/") {
		return value
	}
	return "/" + strings.TrimLeft(value, "/")
}

func paginaPrincipalNormalizeYouTubeURL(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	lower := strings.ToLower(value)
	if strings.HasPrefix(lower, "https://youtu.be/") ||
		strings.HasPrefix(lower, "https://www.youtube.com/") ||
		strings.HasPrefix(lower, "https://youtube.com/") ||
		strings.HasPrefix(lower, "http://youtu.be/") ||
		strings.HasPrefix(lower, "http://www.youtube.com/") ||
		strings.HasPrefix(lower, "http://youtube.com/") {
		// Normalizar a https para evitar mixed-content en el portal.
		value = strings.TrimSpace(value)
		if strings.HasPrefix(strings.ToLower(value), "http://") {
			value = "https://" + strings.TrimPrefix(value, "http://")
		}
		return value
	}
	return ""
}

func paginaPrincipalNormalizeText(raw, fallback string) string {
	value := strings.TrimSpace(raw)
	if value != "" {
		return value
	}
	return strings.TrimSpace(fallback)
}

func paginaPrincipalNormalizePoints(raw, fallback []string) []string {
	normalized := make([]string, 0, len(raw))
	for _, item := range raw {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, trimmed)
	}
	if len(normalized) > 0 {
		return normalized
	}
	for _, item := range fallback {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, trimmed)
	}
	return normalized
}

func paginaPrincipalNormalizeConfig(cfg paginaPrincipalConfig) paginaPrincipalConfig {
	defaults := paginaPrincipalDefaultConfig()
	if cfg.Cantidad <= 0 {
		cfg.Cantidad = len(cfg.Tarjetas)
	}
	if cfg.Cantidad <= 0 {
		cfg.Cantidad = defaults.Cantidad
	}
	if cfg.Cantidad > paginaPrincipalDefaultCardLimit {
		cfg.Cantidad = paginaPrincipalDefaultCardLimit
	}

	normalized := make([]paginaPrincipalCard, 0, cfg.Cantidad)
	for i := 0; i < cfg.Cantidad; i++ {
		base := defaults.Tarjetas[i%len(defaults.Tarjetas)]
		var current paginaPrincipalCard
		if i < len(cfg.Tarjetas) {
			current = cfg.Tarjetas[i]
		}
		title := strings.TrimSpace(current.Titulo)
		if title == "" {
			title = base.Titulo
		}
		description := strings.TrimSpace(current.Descripcion)
		if description == "" {
			description = base.Descripcion
		}
		normalized = append(normalized, paginaPrincipalCard{
			TipoTarjeta:       paginaPrincipalNormalizeCardType(current.TipoTarjeta),
			Titulo:            title,
			Descripcion:       description,
			ImagenURL:         paginaPrincipalNormalizeImageURL(current.ImagenURL, base.ImagenURL),
			ImagenSecundaria:  paginaPrincipalNormalizeImageURL(current.ImagenSecundaria, base.ImagenSecundaria),
			Enlace:            paginaPrincipalNormalizeLink(current.Enlace, base.Enlace),
			YouTubeURL:        paginaPrincipalNormalizeYouTubeURL(current.YouTubeURL),
			DetalleEtiqueta:   paginaPrincipalNormalizeText(current.DetalleEtiqueta, base.DetalleEtiqueta),
			DetalleTitular:    paginaPrincipalNormalizeText(current.DetalleTitular, base.DetalleTitular),
			DetalleParrafoUno: paginaPrincipalNormalizeText(current.DetalleParrafoUno, base.DetalleParrafoUno),
			DetalleParrafoDos: paginaPrincipalNormalizeText(current.DetalleParrafoDos, base.DetalleParrafoDos),
			DetallePuntos:     paginaPrincipalNormalizePoints(current.DetallePuntos, base.DetallePuntos),
		})
	}

	return paginaPrincipalConfig{
		Cantidad: cfg.Cantidad,
		Tarjetas: normalized,
		Estilos:  paginaPrincipalNormalizeVisualSettings(cfg.Estilos),
	}
}

func informacionModulosNormalizeFeatures(raw []string, fallback []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, value)
		if len(out) >= informacionModulosFeatureLimit {
			break
		}
	}
	if len(out) > 0 {
		return out
	}
	for _, item := range fallback {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		out = append(out, value)
		if len(out) >= informacionModulosFeatureLimit {
			break
		}
	}
	if len(out) == 0 {
		out = []string{"Caracteristica principal"}
	}
	return out
}

func informacionModulosNormalizeConfig(cfg informacionModulosConfig) informacionModulosConfig {
	defaults := informacionModulosDefaultConfig()
	title := strings.TrimSpace(cfg.Titulo)
	if title == "" {
		title = defaults.Titulo
	}

	source := cfg.Modulos
	if len(source) == 0 {
		source = defaults.Modulos
	}
	source = informacionModulosEnsureDefaultHighlights(source, defaults.Modulos)
	documentosBase := informacionModulosDocumentosElectronicosDefault(defaults.Modulos)

	modules := make([]informacionModuloItem, 0, len(source))
	for i, item := range source {
		base := defaults.Modulos[i%len(defaults.Modulos)]
		moduleTitle := strings.TrimSpace(item.Titulo)
		if moduleTitle == "" {
			moduleTitle = base.Titulo
		}
		if informacionModulosIsDocumentosElectronicosTitle(moduleTitle) {
			base = documentosBase
		}
		features := informacionModulosNormalizeFeatures(item.Caracteristicas, base.Caracteristicas)
		if informacionModulosIsDocumentosElectronicosTitle(moduleTitle) && !informacionModulosHasDocumentosDianColombia(features) {
			features = informacionModulosNormalizeFeatures(base.Caracteristicas, base.Caracteristicas)
		}
		modules = append(modules, informacionModuloItem{
			Titulo:          moduleTitle,
			IconoURL:        paginaPrincipalNormalizeImageURL(item.IconoURL, base.IconoURL),
			Caracteristicas: features,
		})
	}

	return informacionModulosConfig{Titulo: title, Modulos: modules}
}

func informacionModulosDocumentosElectronicosDefault(defaults []informacionModuloItem) informacionModuloItem {
	for _, item := range defaults {
		if informacionModulosIsDocumentosElectronicosTitle(item.Titulo) {
			return item
		}
	}
	return informacionModuloItem{
		Titulo:          "Documentos electronicos",
		IconoURL:        "/img/invoice.svg",
		Caracteristicas: []string{"Documentos electronicos DIAN Colombia"},
	}
}

func informacionModulosIsDocumentosElectronicosTitle(title string) bool {
	key := strings.ToLower(strings.TrimSpace(title))
	key = strings.ReplaceAll(key, "é", "e")
	return key == "documentos electronicos" || key == "documentos electrónicos"
}

func informacionModulosHasDocumentosDianColombia(features []string) bool {
	for _, item := range features {
		key := strings.ToLower(strings.TrimSpace(item))
		key = strings.ReplaceAll(key, "ó", "o")
		if strings.Contains(key, "documentos electronicos dian colombia") {
			return true
		}
	}
	return false
}

func informacionModulosEnsureDefaultHighlights(source, defaults []informacionModuloItem) []informacionModuloItem {
	if len(source) == 0 {
		return source
	}
	seen := map[string]bool{}
	for _, item := range source {
		seen[strings.ToLower(strings.TrimSpace(item.Titulo))] = true
	}
	for _, item := range defaults {
		key := strings.ToLower(strings.TrimSpace(item.Titulo))
		switch key {
		case "grafologix", "camaras y dvr", "energia solar":
			if !seen[key] && len(source) < informacionModulosDefaultLimit {
				source = append(source, item)
				seen[key] = true
			}
		}
	}
	if len(source) > informacionModulosDefaultLimit {
		source = source[:informacionModulosDefaultLimit]
	}
	return source
}

func informacionModulosLoadConfig(dbSuper *sql.DB) (informacionModulosConfig, string, string, error) {
	cfg := informacionModulosDefaultConfig()
	stored, _, _, updatedAt, err := dbpkg.GetConfigEntry(dbSuper, informacionModulosConfigKey)
	if err != nil {
		return cfg, "", "", err
	}
	updatedBy, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, informacionModulosUpdatedByKey)
	if strings.TrimSpace(stored) == "" {
		return cfg, "", strings.TrimSpace(updatedBy), nil
	}
	var decoded informacionModulosConfig
	if err := json.Unmarshal([]byte(stored), &decoded); err != nil {
		log.Printf("[informacion_modulos] invalid config JSON, fallback defaults: %v", err)
		return cfg, strings.TrimSpace(updatedAt), strings.TrimSpace(updatedBy), nil
	}
	return informacionModulosNormalizeConfig(decoded), strings.TrimSpace(updatedAt), strings.TrimSpace(updatedBy), nil
}

func informacionModulosSaveConfig(dbSuper *sql.DB, cfg informacionModulosConfig, updatedBy string) (informacionModulosConfig, error) {
	normalized := informacionModulosNormalizeConfig(cfg)
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return normalized, err
	}
	if err := dbpkg.SetConfigValue(dbSuper, informacionModulosConfigKey, string(encoded), false); err != nil {
		return normalized, err
	}
	actor := strings.TrimSpace(updatedBy)
	if actor == "" {
		actor = "sistema"
	}
	if err := dbpkg.SetConfigValue(dbSuper, informacionModulosUpdatedByKey, actor, false); err != nil {
		return normalized, err
	}
	return normalized, nil
}

func noticiasPortalNormalizeText(raw string, fallback string, max int) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		value = strings.TrimSpace(fallback)
	}
	value = strings.Join(strings.Fields(value), " ")
	if max > 0 && len([]rune(value)) > max {
		runes := []rune(value)
		value = strings.TrimSpace(string(runes[:max]))
	}
	return value
}

func noticiasPortalNormalizeMultiline(raw string, fallback string, max int) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		value = strings.TrimSpace(fallback)
	}
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	lines := strings.Split(value, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		clean := strings.Join(strings.Fields(line), " ")
		if clean != "" {
			out = append(out, clean)
		}
	}
	value = strings.Join(out, "\n\n")
	if max > 0 && len([]rune(value)) > max {
		runes := []rune(value)
		value = strings.TrimSpace(string(runes[:max]))
	}
	return value
}

func noticiasPortalNormalizeURL(raw string, fallback string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		value = strings.TrimSpace(fallback)
	}
	if value == "" {
		return ""
	}
	if strings.HasPrefix(value, "/") {
		if strings.Contains(value, "..") || strings.Contains(value, "\\") {
			return strings.TrimSpace(fallback)
		}
		return value
	}
	parsed, err := urlParseForNoticias(value)
	if err != nil || parsed == "" {
		return strings.TrimSpace(fallback)
	}
	return parsed
}

func urlParseForNoticias(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return trimmed, nil
	}
	return "", fmt.Errorf("url invalida")
}

func noticiasPortalNormalizeDate(raw string, fallback string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		value = strings.TrimSpace(fallback)
	}
	if value == "" {
		return time.Now().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", value); err == nil {
		return value
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.Format("2006-01-02")
	}
	return strings.TrimSpace(fallback)
}

func noticiasPortalNormalizeTags(raw []string, fallback []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		value := noticiasPortalNormalizeText(item, "", 42)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, value)
		if len(out) >= 8 {
			break
		}
	}
	if len(out) > 0 {
		return out
	}
	for _, item := range fallback {
		value := noticiasPortalNormalizeText(item, "", 42)
		if value == "" {
			continue
		}
		out = append(out, value)
		if len(out) >= 8 {
			break
		}
	}
	if len(out) == 0 {
		out = []string{"Noticia"}
	}
	return out
}

func noticiasPortalNormalizeConfig(cfg noticiasPortalConfig) noticiasPortalConfig {
	defaults := noticiasPortalDefaultConfig()
	source := cfg.Noticias
	if len(source) == 0 {
		source = defaults.Noticias
	}
	if len(source) > noticiasPortalDefaultLimit {
		source = source[:noticiasPortalDefaultLimit]
	}
	items := make([]noticiaPortalItem, 0, len(source))
	for i, item := range source {
		base := defaults.Noticias[i%len(defaults.Noticias)]
		active := item.Activa
		if !active && strings.TrimSpace(item.Titulo) == "" && strings.TrimSpace(item.Contenido) == "" {
			active = base.Activa
		}
		items = append(items, noticiaPortalItem{
			Titulo:       noticiasPortalNormalizeText(item.Titulo, base.Titulo, 180),
			Resumen:      noticiasPortalNormalizeText(item.Resumen, base.Resumen, 360),
			Contenido:    noticiasPortalNormalizeMultiline(item.Contenido, base.Contenido, 4000),
			Categoria:    noticiasPortalNormalizeText(item.Categoria, base.Categoria, 80),
			Fecha:        noticiasPortalNormalizeDate(item.Fecha, base.Fecha),
			ImagenURL:    noticiasPortalNormalizeURL(item.ImagenURL, base.ImagenURL),
			FuenteNombre: noticiasPortalNormalizeText(item.FuenteNombre, base.FuenteNombre, 120),
			FuenteURL:    noticiasPortalNormalizeURL(item.FuenteURL, base.FuenteURL),
			Etiquetas:    noticiasPortalNormalizeTags(item.Etiquetas, base.Etiquetas),
			Destacada:    item.Destacada,
			Activa:       active,
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Destacada != items[j].Destacada {
			return items[i].Destacada
		}
		return items[i].Fecha > items[j].Fecha
	})
	return noticiasPortalConfig{
		Titulo:       noticiasPortalNormalizeText(cfg.Titulo, defaults.Titulo, 90),
		Subtitulo:    noticiasPortalNormalizeText(cfg.Subtitulo, defaults.Subtitulo, 220),
		PortadaURL:   noticiasPortalNormalizeURL(cfg.PortadaURL, defaults.PortadaURL),
		PerfilURL:    noticiasPortalNormalizeURL(cfg.PerfilURL, defaults.PerfilURL),
		NombrePagina: noticiasPortalNormalizeText(cfg.NombrePagina, defaults.NombrePagina, 90),
		Usuario:      noticiasPortalNormalizeText(cfg.Usuario, defaults.Usuario, 60),
		Descripcion:  noticiasPortalNormalizeText(cfg.Descripcion, defaults.Descripcion, 260),
		Noticias:     items,
	}
}

func noticiasPortalLoadConfig(dbSuper *sql.DB) (noticiasPortalConfig, string, string, error) {
	cfg := noticiasPortalDefaultConfig()
	stored, _, _, updatedAt, err := dbpkg.GetConfigEntry(dbSuper, noticiasPortalConfigKey)
	if err != nil {
		return cfg, "", "", err
	}
	updatedBy, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, noticiasPortalUpdatedByKey)
	if strings.TrimSpace(stored) == "" {
		return cfg, "", strings.TrimSpace(updatedBy), nil
	}
	var decoded noticiasPortalConfig
	if err := json.Unmarshal([]byte(stored), &decoded); err != nil {
		log.Printf("[noticias_portal] invalid config JSON, fallback defaults: %v", err)
		return cfg, strings.TrimSpace(updatedAt), strings.TrimSpace(updatedBy), nil
	}
	return noticiasPortalNormalizeConfig(decoded), strings.TrimSpace(updatedAt), strings.TrimSpace(updatedBy), nil
}

func noticiasPortalSaveConfig(dbSuper *sql.DB, cfg noticiasPortalConfig, updatedBy string) (noticiasPortalConfig, error) {
	normalized := noticiasPortalNormalizeConfig(cfg)
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return normalized, err
	}
	if err := dbpkg.SetConfigValue(dbSuper, noticiasPortalConfigKey, string(encoded), false); err != nil {
		return normalized, err
	}
	actor := strings.TrimSpace(updatedBy)
	if actor == "" {
		actor = "sistema"
	}
	if err := dbpkg.SetConfigValue(dbSuper, noticiasPortalUpdatedByKey, actor, false); err != nil {
		return normalized, err
	}
	return normalized, nil
}

func paginaPrincipalLoadConfig(dbSuper *sql.DB) (paginaPrincipalConfig, string, string, error) {
	cfg := paginaPrincipalDefaultConfig()
	stored, _, _, updatedAt, err := dbpkg.GetConfigEntry(dbSuper, paginaPrincipalConfigKey)
	if err != nil {
		return cfg, "", "", err
	}
	updatedBy, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, paginaPrincipalConfigUpdatedByKey)
	if strings.TrimSpace(stored) == "" {
		return cfg, "", strings.TrimSpace(updatedBy), nil
	}

	var decoded paginaPrincipalConfig
	if err := json.Unmarshal([]byte(stored), &decoded); err != nil {
		log.Printf("[pagina_principal] invalid config JSON, fallback defaults: %v", err)
		return cfg, strings.TrimSpace(updatedAt), strings.TrimSpace(updatedBy), nil
	}

	return paginaPrincipalNormalizeConfig(decoded), strings.TrimSpace(updatedAt), strings.TrimSpace(updatedBy), nil
}

func paginaPrincipalSaveConfig(dbSuper *sql.DB, cfg paginaPrincipalConfig, updatedBy string) error {
	normalized := paginaPrincipalNormalizeConfig(cfg)
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return err
	}
	if err := dbpkg.SetConfigValue(dbSuper, paginaPrincipalConfigKey, string(encoded), false); err != nil {
		return err
	}
	actor := strings.TrimSpace(updatedBy)
	if actor == "" {
		actor = "sistema"
	}
	if err := dbpkg.SetConfigValue(dbSuper, paginaPrincipalConfigUpdatedByKey, actor, false); err != nil {
		return err
	}
	return nil
}

func paginaPrincipalListImageURLs(webDir string) ([]string, error) {
	imgDir := filepath.Join(strings.TrimSpace(webDir), "img")
	entries, err := os.ReadDir(imgDir)
	if err != nil {
		return nil, err
	}
	images := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		switch ext {
		case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg":
			images = append(images, "/img/"+entry.Name())
		}
	}
	sort.Slice(images, func(i, j int) bool {
		return strings.ToLower(images[i]) < strings.ToLower(images[j])
	})
	return images, nil
}

func paginaPrincipalSanitizeUploadName(raw string) string {
	base := strings.ToLower(strings.TrimSpace(filepath.Base(raw)))
	if base == "" || base == "." {
		base = "imagen"
	}
	var b strings.Builder
	for _, r := range base {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			b.WriteRune(r)
			continue
		}
		b.WriteRune('-')
	}
	clean := strings.Trim(b.String(), ".-_")
	if clean == "" {
		clean = "imagen"
	}
	return clean
}

func paginaPrincipalAllowedUploadExt(ext string) bool {
	switch strings.ToLower(strings.TrimSpace(ext)) {
	case ".png", ".jpg", ".jpeg", ".webp", ".gif":
		return true
	default:
		return false
	}
}

func paginaPrincipalUploadImage(w http.ResponseWriter, r *http.Request, webDir string) {
	r.Body = http.MaxBytesReader(w, r.Body, 8<<20)
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		http.Error(w, "imagen invalida o demasiado grande", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("imagen")
	if err != nil {
		http.Error(w, "imagen requerida", http.StatusBadRequest)
		return
	}
	defer file.Close()

	original := paginaPrincipalSanitizeUploadName(header.Filename)
	ext := strings.ToLower(filepath.Ext(original))
	if !paginaPrincipalAllowedUploadExt(ext) {
		http.Error(w, "formato no permitido; usa png, jpg, jpeg, webp o gif", http.StatusBadRequest)
		return
	}
	nameWithoutExt := strings.TrimSuffix(original, ext)
	if nameWithoutExt == "" {
		nameWithoutExt = "imagen"
	}
	finalName := fmt.Sprintf("pagina_principal_%d_%s%s", time.Now().UnixNano(), nameWithoutExt, ext)
	imgDir := filepath.Join(strings.TrimSpace(webDir), "img")
	if err := os.MkdirAll(imgDir, 0755); err != nil {
		http.Error(w, "no se pudo preparar carpeta de imagenes: "+err.Error(), http.StatusInternalServerError)
		return
	}
	destPath := filepath.Join(imgDir, finalName)
	dest, err := os.OpenFile(destPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		http.Error(w, "no se pudo guardar imagen: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dest.Close()
	if _, err := io.Copy(dest, file); err != nil {
		http.Error(w, "no se pudo escribir imagen: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":  true,
		"url": "/img/" + finalName,
	})
}

func paginaPrincipalRoleIsSuper(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "super_administrador", "superadministrador", "superadmin", "super":
		return true
	default:
		return false
	}
}

func paginaPrincipalRequireSuperAdmin(w http.ResponseWriter, r *http.Request, dbSuper *sql.DB) (string, bool) {
	cookie, err := r.Cookie("session_token")
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return "", false
	}
	session, err := dbpkg.GetSessionByToken(dbSuper, cookie.Value)
	if err != nil || session == nil {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return "", false
	}
	admin, err := dbpkg.GetAdminByEmail(dbSuper, strings.TrimSpace(session.AdminEmail))
	if err != nil {
		http.Error(w, "failed to resolve admin session", http.StatusInternalServerError)
		return "", false
	}
	if admin == nil {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return "", false
	}
	if !paginaPrincipalRoleIsSuper(admin.Role) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return "", false
	}
	return strings.TrimSpace(admin.Email), true
}

// SuperInformacionModulosHandler administra la lista editable de modulos principales del portal.
func SuperInformacionModulosHandler(dbSuper *sql.DB, webDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminEmail, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper)
		if !ok {
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "" {
			action = "config"
		}

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "config", "get", "listar":
				cfg, updatedAt, updatedBy, err := informacionModulosLoadConfig(dbSuper)
				if err != nil {
					http.Error(w, "failed to read informacion de modulos: "+err.Error(), http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":          true,
					"config":      cfg,
					"updated_at":  updatedAt,
					"updated_by":  updatedBy,
					"admin_email": adminEmail,
				})
				return
			case "defaults", "predeterminados":
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":     true,
					"config": informacionModulosDefaultConfig(),
				})
				return
			case "imagenes", "images", "listar_imagenes":
				images, err := paginaPrincipalListImageURLs(webDir)
				if err != nil {
					http.Error(w, "failed to list images: "+err.Error(), http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "imagenes": images, "total": len(images)})
				return
			default:
				http.Error(w, "action not supported", http.StatusBadRequest)
				return
			}

		case http.MethodPut, http.MethodPost:
			if action != "config" && action != "save" && action != "guardar" {
				http.Error(w, "action not supported", http.StatusBadRequest)
				return
			}
			var payload informacionModulosConfig
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
				return
			}
			normalized, err := informacionModulosSaveConfig(dbSuper, payload, adminEmail)
			if err != nil {
				http.Error(w, "failed to save informacion de modulos: "+err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":         true,
				"saved":      true,
				"config":     normalized,
				"updated_by": adminEmail,
			})
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

// PublicInformacionModulosHandler expone los modulos principales configurados para el index publico.
func PublicInformacionModulosHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		cfg, updatedAt, _, err := informacionModulosLoadConfig(dbSuper)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "failed to read informacion de modulos: "+err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":         true,
			"titulo":     cfg.Titulo,
			"modulos":    cfg.Modulos,
			"updated_at": updatedAt,
		})
	}
}

// SuperNoticiasPortalHandler administra la pagina publica de noticias.
func SuperNoticiasPortalHandler(dbSuper *sql.DB, webDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminEmail, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper)
		if !ok {
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "" {
			action = "config"
		}
		switch r.Method {
		case http.MethodGet:
			switch action {
			case "config", "get", "listar":
				cfg, updatedAt, updatedBy, err := noticiasPortalLoadConfig(dbSuper)
				if err != nil {
					http.Error(w, "failed to read noticias: "+err.Error(), http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":          true,
					"config":      cfg,
					"updated_at":  updatedAt,
					"updated_by":  updatedBy,
					"admin_email": adminEmail,
				})
				return
			case "defaults", "predeterminados":
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "config": noticiasPortalDefaultConfig()})
				return
			case "imagenes", "images", "listar_imagenes":
				images, err := paginaPrincipalListImageURLs(webDir)
				if err != nil {
					http.Error(w, "failed to list images: "+err.Error(), http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "imagenes": images, "total": len(images)})
				return
			default:
				http.Error(w, "action not supported", http.StatusBadRequest)
				return
			}
		case http.MethodPut, http.MethodPost:
			if action == "upload_image" || action == "subir_imagen" {
				if r.Method != http.MethodPost {
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
					return
				}
				paginaPrincipalUploadImage(w, r, webDir)
				return
			}
			if action != "config" && action != "save" && action != "guardar" {
				http.Error(w, "action not supported", http.StatusBadRequest)
				return
			}
			var payload noticiasPortalConfig
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
				return
			}
			normalized, err := noticiasPortalSaveConfig(dbSuper, payload, adminEmail)
			if err != nil {
				http.Error(w, "failed to save noticias: "+err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":         true,
				"saved":      true,
				"config":     normalized,
				"updated_by": adminEmail,
			})
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

// PublicNoticiasPortalHandler expone la pagina de noticias para el menu flotante.
func PublicNoticiasPortalHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		cfg, updatedAt, _, err := noticiasPortalLoadConfig(dbSuper)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "failed to read noticias: "+err.Error(), http.StatusInternalServerError)
			return
		}
		publicNews := make([]noticiaPortalItem, 0, len(cfg.Noticias))
		for _, item := range cfg.Noticias {
			if item.Activa {
				publicNews = append(publicNews, item)
			}
		}
		cfg.Noticias = publicNews
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":         true,
			"config":     cfg,
			"updated_at": updatedAt,
		})
	}
}

// SuperPaginaPrincipalHandler administra las tarjetas configurables del portal principal y su landing descriptiva.
func SuperPaginaPrincipalHandler(dbSuper *sql.DB, webDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminEmail, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper)
		if !ok {
			return
		}

		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "" {
			action = "config"
		}

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "config", "get", "listar":
				cfg, updatedAt, updatedBy, err := paginaPrincipalLoadConfig(dbSuper)
				if err != nil {
					http.Error(w, "failed to read pagina principal config: "+err.Error(), http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":          true,
					"config":      cfg,
					"updated_at":  updatedAt,
					"updated_by":  updatedBy,
					"admin_email": adminEmail,
				})
				return
			case "imagenes", "images", "listar_imagenes":
				images, err := paginaPrincipalListImageURLs(webDir)
				if err != nil {
					http.Error(w, "failed to list images: "+err.Error(), http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":       true,
					"imagenes": images,
					"total":    len(images),
				})
				return
			default:
				http.Error(w, "action not supported", http.StatusBadRequest)
				return
			}

		case http.MethodPut, http.MethodPost:
			if action == "upload_image" || action == "subir_imagen" {
				if r.Method != http.MethodPost {
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
					return
				}
				paginaPrincipalUploadImage(w, r, webDir)
				return
			}
			if action != "config" && action != "save" && action != "guardar" {
				http.Error(w, "action not supported", http.StatusBadRequest)
				return
			}
			var payload paginaPrincipalConfig
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
				return
			}
			if payload.Cantidad <= 0 {
				http.Error(w, "cantidad must be greater than 0", http.StatusBadRequest)
				return
			}
			normalized := paginaPrincipalNormalizeConfig(payload)
			if err := paginaPrincipalSaveConfig(dbSuper, normalized, adminEmail); err != nil {
				http.Error(w, "failed to save pagina principal config: "+err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"saved":      true,
				"config":     normalized,
				"updated_by": adminEmail,
			})
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

// PublicPaginaPrincipalHandler expone tarjetas del portal para visualizacion publica del index y la landing descriptiva.
func PublicPaginaPrincipalHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		cfg, updatedAt, _, err := paginaPrincipalLoadConfig(dbSuper)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "failed to read pagina principal config: "+err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":                      true,
			"cantidad":                cfg.Cantidad,
			"tarjetas":                cfg.Tarjetas,
			"estilos":                 cfg.Estilos,
			"whatsapp_contact_number": paginaPrincipalLoadWhatsAppContactNumber(dbSuper),
			"updated_at":              updatedAt,
		})
	}
}
