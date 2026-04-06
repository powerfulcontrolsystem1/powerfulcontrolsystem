package handlers

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type empresaReporteCatalogoItem struct {
	Key         string   `json:"key"`
	Title       string   `json:"title"`
	Level       string   `json:"level"`
	Description string   `json:"description"`
	Formats     []string `json:"formats"`
}

type empresaReporteDataset struct {
	Key         string                   `json:"key"`
	Title       string                   `json:"title"`
	Level       string                   `json:"level"`
	Description string                   `json:"description"`
	EmpresaID   int64                    `json:"empresa_id"`
	Desde       string                   `json:"desde"`
	Hasta       string                   `json:"hasta"`
	GeneratedAt string                   `json:"generated_at"`
	Columns     []string                 `json:"columns"`
	Rows        []map[string]interface{} `json:"rows"`
	RowCount    int                      `json:"row_count"`
	Summary     map[string]interface{}   `json:"summary,omitempty"`
}

type empresaReportesSuiteResponse struct {
	EmpresaID  int64                   `json:"empresa_id"`
	Desde      string                  `json:"desde"`
	Hasta      string                  `json:"hasta"`
	GeneradoEn string                  `json:"generado_en"`
	Tablero    interface{}             `json:"tablero"`
	Datasets   []empresaReporteDataset `json:"datasets"`
}

type reportesBuilder struct {
	db               *sql.DB
	empresaID        int64
	desde            string
	hasta            string
	cierreID         int64
	empleadoNominaID int64
	cajaCodigo       string
	turno            string
	usuario          string
	categoria        string
	metodoPago       string
	maxRows          int
	includeInactive  bool

	tableroCache *dbpkg.EmpresaReportesTableroResumen
	ventasCache  []dbpkg.CarritoCompra
	itemsCache   map[int64][]dbpkg.CarritoCompraItem
}

const (
	reporteDatasetEmpresarialTablero          = "empresarial_tablero"
	reporteDatasetContableEstadoResultados    = "contable_estado_resultados"
	reporteDatasetContableBalanceGeneral      = "contable_balance_general"
	reporteDatasetContableFlujoCaja           = "contable_flujo_caja"
	reporteDatasetOperativoModulos            = "operativo_modulos_resumen"
	reporteDatasetOperativoReservas           = "operativo_reservas_ocupacion"
	reporteDatasetOperativoTarifas            = "operativo_tarifas_ingresos"
	reporteDatasetOperativoTarifasComparativo = "operativo_tarifas_comparativo_estaciones"
	reporteDatasetOperativoCadena             = "operativo_cadena_cumplimiento"
	reporteDatasetOperativoVentasDetalle      = "operativo_ventas_detalle"
	reporteDatasetOperativoTurno              = "reporte_de_turno"
	reporteDatasetOperativoTopProductos       = "operativo_top_productos"
	reporteDatasetOperativoTopClientes        = "operativo_top_clientes"
	reporteDatasetOperativoClientesSegmentos  = "operativo_clientes_segmentacion_comercial"
	reporteDatasetOperativoInventario         = "operativo_inventario_bodega"
	reporteDatasetOperativoCompras            = "operativo_compras_movimientos"
	reporteDatasetOperativoPropinas           = "operativo_propinas_acumulado"
	reporteDatasetOperativoComisiones         = "operativo_comisiones_lavador"
	reporteDatasetOperativoFacturacion        = "operativo_facturacion_trazabilidad"
	reporteDatasetOperativoAuditoria          = "operativo_auditoria_acciones"
	reporteDatasetOperativoAsistenciaNomina   = "operativo_asistencia_nomina_auditoria"
	reporteDatasetOperativoVehiculos          = "operativo_vehiculos_permanencia"
	reporteDatasetContableMovFin              = "contable_movimientos_financieros"
	reporteDatasetContableEventos             = "contable_eventos_contables"
	reporteDatasetContableAsientos            = "contable_asientos_contables"
	reporteDatasetContableNomina              = "contable_nomina_liquidaciones"
)

var reportesCatalogo = []empresaReporteCatalogoItem{
	{
		Key:         reporteDatasetEmpresarialTablero,
		Title:       "Tablero Empresarial Ejecutivo",
		Level:       "empresarial",
		Description: "KPI consolidados operativos, financieros y contables por empresa.",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
	{
		Key:         reporteDatasetContableEstadoResultados,
		Title:       "Estado de Resultados",
		Level:       "contable",
		Description: "Ingresos, gastos y utilidad operacional del rango.",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
	{
		Key:         reporteDatasetContableBalanceGeneral,
		Title:       "Balance General",
		Level:       "contable",
		Description: "Activos, pasivos, patrimonio y cuadre contable.",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
	{
		Key:         reporteDatasetContableFlujoCaja,
		Title:       "Flujo de Caja Diario",
		Level:       "contable",
		Description: "Flujo diario de ingresos y egresos con neto y saldo acumulado del periodo.",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
	{
		Key:         reporteDatasetOperativoModulos,
		Title:       "Resumen por Módulos",
		Level:       "operativo",
		Description: "Consolida estado por módulo (totales, activos, rango y último registro) para seguimiento integral.",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
	{
		Key:         reporteDatasetOperativoReservas,
		Title:       "Reservas - Ocupación y Cumplimiento",
		Level:       "operativo",
		Description: "Consolida reservas por estación con ocupación estimada y cumplimiento de confirmación por rango.",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
	{
		Key:         reporteDatasetOperativoTarifas,
		Title:       "Tarifas - Ingresos por Modelo",
		Level:       "operativo",
		Description: "Consolida ingresos de ventas por modelo de tarifa aplicado (día, minutos o sin modelo).",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
	{
		Key:         reporteDatasetOperativoTarifasComparativo,
		Title:       "Tarifas por Día - Comparativo Esperado vs Real por Estación",
		Level:       "operativo",
		Description: "Compara ingresos esperados (motor de tarifa diaria con prorrateo) frente al ingreso real cobrado por estación.",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
	{
		Key:         reporteDatasetOperativoCadena,
		Title:       "CRM/Producción/Logística - Conversión y Cumplimiento",
		Level:       "operativo",
		Description: "Consolida conversión comercial y cumplimiento operativo por módulo en el rango consultado.",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
	{
		Key:         reporteDatasetOperativoVentasDetalle,
		Title:       "Ventas Cerradas Detalle",
		Level:       "operativo",
		Description: "Detalle de ventas cerradas con cliente, canal e importe.",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
	{
		Key:         reporteDatasetOperativoTurno,
		Title:       "Reporte de Turno",
		Level:       "operativo",
		Description: "Turno por usuario/caja con activacion de carritos, ventas por tipo, gastos y efectivo esperado.",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
	{
		Key:         reporteDatasetOperativoTopProductos,
		Title:       "Top Productos Vendidos",
		Level:       "operativo",
		Description: "Ranking de productos por unidades e ingresos vendidos.",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
	{
		Key:         reporteDatasetOperativoTopClientes,
		Title:       "Top Clientes",
		Level:       "operativo",
		Description: "Ranking de clientes por ventas e ingresos del periodo.",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
	{
		Key:         reporteDatasetOperativoClientesSegmentos,
		Title:       "Clientes - Segmentacion Comercial Masiva",
		Level:       "operativo",
		Description: "Listado masivo de clientes con segmento comercial, metricas de compra y accion sugerida para campanas.",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
	{
		Key:         reporteDatasetOperativoInventario,
		Title:       "Inventario por Bodega",
		Level:       "operativo",
		Description: "Existencias por bodega con rotacion estimada, riesgo de quiebre y valorizacion de inventario.",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
	{
		Key:         reporteDatasetOperativoCompras,
		Title:       "Compras por Proveedor y Recepcion vs Orden",
		Level:       "operativo",
		Description: "Consolida documentos de compra por proveedor para medir costo ordenado, recepcionado y brecha operativa.",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
	{
		Key:         reporteDatasetOperativoPropinas,
		Title:       "Propinas - Acumulado por Usuario",
		Level:       "operativo",
		Description: "Consolida propinas por usuario y periodo con distribucion directa/universal y participacion sobre el total.",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
	{
		Key:         reporteDatasetOperativoComisiones,
		Title:       "Comisiones por Servicio - Acumulado por Lavador",
		Level:       "operativo",
		Description: "Consolida comisiones por lavador en el periodo con base de servicios, total de comision y participacion.",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
	{
		Key:         reporteDatasetOperativoFacturacion,
		Title:       "Facturacion Electronica - Documentos y Trazabilidad",
		Level:       "operativo",
		Description: "Consolida documentos por tipo para seguimiento de emision, anulacion y trazabilidad legal.",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
	{
		Key:         reporteDatasetOperativoAuditoria,
		Title:       "Auditoria Empresarial - Acciones Criticas",
		Level:       "operativo",
		Description: "Consolida auditoria por modulo/usuario con errores HTTP y accion principal del periodo.",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
	{
		Key:         reporteDatasetOperativoAsistenciaNomina,
		Title:       "Asistencia - Auditoria para Nomina",
		Level:       "operativo",
		Description: "Consolida asistencia por empleado para auditoria de nomina (horas, tardanzas, ausencias e inconsistencias).",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
	{
		Key:         reporteDatasetOperativoVehiculos,
		Title:       "Vehiculos - Permanencia y Tiempos de Estancia",
		Level:       "operativo",
		Description: "Consolida permanencia por vehiculo con tiempos de estancia por registro y estado operativo.",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
	{
		Key:         reporteDatasetContableMovFin,
		Title:       "Movimientos Financieros",
		Level:       "contable",
		Description: "Libro de ingresos/egresos con totales y netos.",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
	{
		Key:         reporteDatasetContableEventos,
		Title:       "Eventos Contables",
		Level:       "contable",
		Description: "Eventos contables por modulo/evento y estado de procesamiento.",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
	{
		Key:         reporteDatasetContableAsientos,
		Title:       "Asientos Contables",
		Level:       "contable",
		Description: "Asientos canónicos con débitos/créditos y diferencia.",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
	{
		Key:         reporteDatasetContableNomina,
		Title:       "Nomina Liquidaciones",
		Level:       "contable",
		Description: "Liquidaciones de nomina por periodo y empleado con totales devengado, deducciones y neto.",
		Formats:     []string{"json", "csv", "txt", "xls", "pdf"},
	},
}

// EmpresaReportesHandler centraliza reportes empresariales, operativos y contables por empresa.
func EmpresaReportesHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "" {
			action = "suite"
		}

		desde := strings.TrimSpace(r.URL.Query().Get("desde"))
		hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
		maxRows, err := parseIntQueryOptional(r, "max_rows")
		if err != nil {
			http.Error(w, "max_rows invalido", http.StatusBadRequest)
			return
		}
		if maxRows <= 0 {
			maxRows = 200
		}
		if maxRows > 1000 {
			maxRows = 1000
		}
		cierreID, err := parseInt64QueryOptional(r, "cierre_id")
		if err != nil {
			http.Error(w, "cierre_id invalido", http.StatusBadRequest)
			return
		}
		if cierreID < 0 {
			http.Error(w, "cierre_id invalido", http.StatusBadRequest)
			return
		}
		empleadoNominaID, err := parseInt64QueryOptional(r, "empleado_nomina_id")
		if err != nil {
			http.Error(w, "empleado_nomina_id invalido", http.StatusBadRequest)
			return
		}
		if empleadoNominaID < 0 {
			http.Error(w, "empleado_nomina_id invalido", http.StatusBadRequest)
			return
		}
		includeInactive := queryBool(r, "include_inactive")

		builder := &reportesBuilder{
			db:               dbEmp,
			empresaID:        empresaID,
			desde:            desde,
			hasta:            hasta,
			cierreID:         cierreID,
			empleadoNominaID: empleadoNominaID,
			cajaCodigo:       strings.TrimSpace(r.URL.Query().Get("caja_codigo")),
			turno:            strings.TrimSpace(r.URL.Query().Get("turno")),
			usuario:          strings.TrimSpace(r.URL.Query().Get("usuario")),
			categoria:        strings.TrimSpace(r.URL.Query().Get("categoria")),
			metodoPago:       strings.TrimSpace(r.URL.Query().Get("metodo_pago")),
			maxRows:          maxRows,
			includeInactive:  includeInactive,
			itemsCache:       make(map[int64][]dbpkg.CarritoCompraItem),
		}

		switch action {
		case "catalogo", "catalog", "datasets":
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"empresa_id": empresaID,
				"datasets":   reportesCatalogo,
			})
			return

		case "tablero", "dashboard":
			tablero, err := builder.getTableroResumen()
			if err != nil {
				http.Error(w, "No se pudo construir el tablero empresarial", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, tablero)
			return

		case "dataset":
			datasetKey := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("dataset")))
			if datasetKey == "" {
				http.Error(w, "dataset es obligatorio", http.StatusBadRequest)
				return
			}
			ds, err := builder.buildDataset(datasetKey)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, ds)
			return

		case "export", "exportar":
			format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
			if format == "" {
				format = "json"
			}
			datasetKey := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("dataset")))

			if datasetKey == "" {
				if format != "json" {
					http.Error(w, "dataset es obligatorio para formatos tabulares (csv, txt, xls o pdf)", http.StatusBadRequest)
					return
				}
				suite, err := builder.buildSuite()
				if err != nil {
					http.Error(w, "No se pudo construir la suite de reportes", http.StatusInternalServerError)
					return
				}
				fileName := reportesBuildFileName("suite_reportes", empresaID, "json")
				w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
				writeJSON(w, http.StatusOK, suite)
				return
			}

			ds, err := builder.buildDataset(datasetKey)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if err := writeReportesDatasetExport(w, ds, format); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			return

		case "suite", "resumen":
			suite, err := builder.buildSuite()
			if err != nil {
				http.Error(w, "No se pudo construir la suite de reportes", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, suite)
			return

		default:
			http.Error(w, "action invalida (use catalogo, suite, dataset, tablero o export)", http.StatusBadRequest)
			return
		}
	}
}

func (b *reportesBuilder) buildSuite() (empresaReportesSuiteResponse, error) {
	tablero, err := b.getTableroResumen()
	if err != nil {
		return empresaReportesSuiteResponse{}, err
	}

	suite := empresaReportesSuiteResponse{
		EmpresaID:  b.empresaID,
		Desde:      b.desde,
		Hasta:      b.hasta,
		GeneradoEn: time.Now().Format("2006-01-02 15:04:05"),
		Tablero:    tablero,
		Datasets:   make([]empresaReporteDataset, 0, len(reportesCatalogo)),
	}

	for _, item := range reportesCatalogo {
		ds, err := b.buildDataset(item.Key)
		if err != nil {
			return empresaReportesSuiteResponse{}, err
		}
		suite.Datasets = append(suite.Datasets, ds)
	}

	return suite, nil
}

func (b *reportesBuilder) buildDataset(key string) (empresaReporteDataset, error) {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case reporteDatasetEmpresarialTablero:
		return b.buildEmpresarialTableroDataset()
	case reporteDatasetContableEstadoResultados:
		return b.buildContableEstadoResultadosDataset()
	case reporteDatasetContableBalanceGeneral:
		return b.buildContableBalanceGeneralDataset()
	case reporteDatasetContableFlujoCaja:
		return b.buildContableFlujoCajaDataset()
	case reporteDatasetOperativoModulos:
		return b.buildOperativoModulosResumenDataset()
	case reporteDatasetOperativoReservas:
		return b.buildOperativoReservasOcupacionDataset()
	case reporteDatasetOperativoTarifas:
		return b.buildOperativoTarifasIngresosDataset()
	case reporteDatasetOperativoTarifasComparativo:
		return b.buildOperativoTarifasComparativoEstacionesDataset()
	case reporteDatasetOperativoCadena:
		return b.buildOperativoCadenaCumplimientoDataset()
	case reporteDatasetOperativoVentasDetalle:
		return b.buildOperativoVentasDetalleDataset()
	case reporteDatasetOperativoTurno:
		return b.buildOperativoTurnoDataset()
	case reporteDatasetOperativoTopProductos:
		return b.buildOperativoTopProductosDataset()
	case reporteDatasetOperativoTopClientes:
		return b.buildOperativoTopClientesDataset()
	case reporteDatasetOperativoClientesSegmentos:
		return b.buildOperativoClientesSegmentacionComercialDataset()
	case reporteDatasetOperativoInventario:
		return b.buildOperativoInventarioBodegaDataset()
	case reporteDatasetOperativoCompras:
		return b.buildOperativoComprasMovimientosDataset()
	case reporteDatasetOperativoPropinas:
		return b.buildOperativoPropinasAcumuladoDataset()
	case reporteDatasetOperativoComisiones:
		return b.buildOperativoComisionesLavadorDataset()
	case reporteDatasetOperativoFacturacion:
		return b.buildOperativoFacturacionTrazabilidadDataset()
	case reporteDatasetOperativoAuditoria:
		return b.buildOperativoAuditoriaAccionesDataset()
	case reporteDatasetOperativoAsistenciaNomina:
		return b.buildOperativoAsistenciaNominaAuditoriaDataset()
	case reporteDatasetOperativoVehiculos:
		return b.buildOperativoVehiculosPermanenciaDataset()
	case reporteDatasetContableMovFin:
		return b.buildContableMovimientosFinancierosDataset()
	case reporteDatasetContableEventos:
		return b.buildContableEventosDataset()
	case reporteDatasetContableAsientos:
		return b.buildContableAsientosDataset()
	case reporteDatasetContableNomina:
		return b.buildContableNominaLiquidacionesDataset()
	default:
		return empresaReporteDataset{}, fmt.Errorf("dataset no soportado")
	}
}

func (b *reportesBuilder) getTableroResumen() (*dbpkg.EmpresaReportesTableroResumen, error) {
	if b.tableroCache != nil {
		return b.tableroCache, nil
	}
	tablero, err := dbpkg.GetEmpresaReportesTableroResumen(b.db, b.empresaID, b.desde, b.hasta)
	if err != nil {
		return nil, err
	}
	b.tableroCache = tablero
	return tablero, nil
}

func (b *reportesBuilder) datasetMeta(key string) empresaReporteCatalogoItem {
	for _, item := range reportesCatalogo {
		if item.Key == key {
			return item
		}
	}
	return empresaReporteCatalogoItem{Key: key, Title: key, Level: "operativo"}
}

func (b *reportesBuilder) newDataset(key string, columns []string) empresaReporteDataset {
	meta := b.datasetMeta(key)
	return empresaReporteDataset{
		Key:         meta.Key,
		Title:       meta.Title,
		Level:       meta.Level,
		Description: meta.Description,
		EmpresaID:   b.empresaID,
		Desde:       b.desde,
		Hasta:       b.hasta,
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		Columns:     append([]string{}, columns...),
		Rows:        make([]map[string]interface{}, 0),
		Summary:     make(map[string]interface{}),
	}
}

func (b *reportesBuilder) buildEmpresarialTableroDataset() (empresaReporteDataset, error) {
	tablero, err := b.getTableroResumen()
	if err != nil {
		return empresaReporteDataset{}, err
	}
	ds := b.newDataset(reporteDatasetEmpresarialTablero, []string{
		"empresa_id",
		"desde",
		"hasta",
		"ventas_cerradas",
		"ventas_hoy",
		"ingresos_ventas",
		"ticket_promedio",
		"balance_financiero",
		"utilidad_operacional",
		"activos",
		"pasivos",
		"patrimonio",
		"cuadre_balance",
	})
	ds.Rows = append(ds.Rows, map[string]interface{}{
		"empresa_id":           tablero.EmpresaID,
		"desde":                tablero.Desde,
		"hasta":                tablero.Hasta,
		"ventas_cerradas":      tablero.Operativo.VentasCerradas,
		"ventas_hoy":           tablero.Operativo.VentasHoy,
		"ingresos_ventas":      tablero.Operativo.IngresosVentas,
		"ticket_promedio":      tablero.Operativo.TicketPromedio,
		"balance_financiero":   tablero.Financiero.Balance,
		"utilidad_operacional": tablero.EstadoResultados.UtilidadOperacional,
		"activos":              tablero.BalanceGeneral.Activos,
		"pasivos":              tablero.BalanceGeneral.Pasivos,
		"patrimonio":           tablero.BalanceGeneral.Patrimonio,
		"cuadre_balance":       tablero.BalanceGeneral.Cuadre,
	})
	ds.RowCount = len(ds.Rows)
	ds.Summary["nivel"] = "empresarial"
	ds.Summary["kpi_operativos"] = 9
	ds.Summary["kpi_contables"] = 5
	return ds, nil
}

func (b *reportesBuilder) buildContableEstadoResultadosDataset() (empresaReporteDataset, error) {
	tablero, err := b.getTableroResumen()
	if err != nil {
		return empresaReporteDataset{}, err
	}
	ds := b.newDataset(reporteDatasetContableEstadoResultados, []string{
		"empresa_id",
		"desde",
		"hasta",
		"ingresos",
		"gastos",
		"utilidad_operacional",
	})
	ds.Rows = append(ds.Rows, map[string]interface{}{
		"empresa_id":           tablero.EmpresaID,
		"desde":                tablero.Desde,
		"hasta":                tablero.Hasta,
		"ingresos":             tablero.EstadoResultados.Ingresos,
		"gastos":               tablero.EstadoResultados.Gastos,
		"utilidad_operacional": tablero.EstadoResultados.UtilidadOperacional,
	})
	ds.RowCount = 1
	ds.Summary["resultado"] = tablero.EstadoResultados.UtilidadOperacional
	return ds, nil
}

func (b *reportesBuilder) buildContableBalanceGeneralDataset() (empresaReporteDataset, error) {
	tablero, err := b.getTableroResumen()
	if err != nil {
		return empresaReporteDataset{}, err
	}
	ds := b.newDataset(reporteDatasetContableBalanceGeneral, []string{
		"empresa_id",
		"desde",
		"hasta",
		"activos",
		"pasivos",
		"patrimonio",
		"resultado_ejercicio",
		"cuadre",
	})
	ds.Rows = append(ds.Rows, map[string]interface{}{
		"empresa_id":          tablero.EmpresaID,
		"desde":               tablero.Desde,
		"hasta":               tablero.Hasta,
		"activos":             tablero.BalanceGeneral.Activos,
		"pasivos":             tablero.BalanceGeneral.Pasivos,
		"patrimonio":          tablero.BalanceGeneral.Patrimonio,
		"resultado_ejercicio": tablero.BalanceGeneral.ResultadoEjercicio,
		"cuadre":              tablero.BalanceGeneral.Cuadre,
	})
	ds.RowCount = 1
	ds.Summary["cuadre"] = tablero.BalanceGeneral.Cuadre
	return ds, nil
}

type reporteModuloResumenDef struct {
	ModuloKey  string
	Modulo     string
	Categoria  string
	Tabla      string
	FechaCampo string
}

var reporteModulosResumenDefs = []reporteModuloResumenDef{
	{ModuloKey: "usuarios_empresa", Modulo: "Usuarios de empresa", Categoria: "seguridad", Tabla: "users", FechaCampo: "fecha_creacion"},
	{ModuloKey: "clientes", Modulo: "Clientes", Categoria: "operativo", Tabla: "clientes", FechaCampo: "fecha_creacion"},
	{ModuloKey: "productos", Modulo: "Inventario - Productos", Categoria: "operativo", Tabla: "productos", FechaCampo: "fecha_creacion"},
	{ModuloKey: "servicios", Modulo: "Servicios", Categoria: "operativo", Tabla: "servicios", FechaCampo: "fecha_creacion"},
	{ModuloKey: "combos_productos", Modulo: "Combos de productos", Categoria: "operativo", Tabla: "combos_productos", FechaCampo: "fecha_creacion"},
	{ModuloKey: "codigos_descuento", Modulo: "Códigos de descuento", Categoria: "operativo", Tabla: "codigos_de_descuento", FechaCampo: "fecha_creacion"},
	{ModuloKey: "carritos", Modulo: "Carritos de compra", Categoria: "ventas", Tabla: "carritos_compras", FechaCampo: "fecha_creacion"},
	{ModuloKey: "reservas", Modulo: "Reservas por estación", Categoria: "ventas", Tabla: "reservas_hotel", FechaCampo: "fecha_creacion"},
	{ModuloKey: "tarifas_minutos", Modulo: "Tarifas por minutos", Categoria: "configuracion", Tabla: "empresa_tarifas_por_minutos", FechaCampo: "fecha_creacion"},
	{ModuloKey: "tarifas_dia", Modulo: "Tarifas por día", Categoria: "configuracion", Tabla: "empresa_tarifas_por_dia", FechaCampo: "fecha_creacion"},
	{ModuloKey: "propinas_movimientos", Modulo: "Propinas", Categoria: "finanzas", Tabla: "empresa_propinas_movimientos", FechaCampo: "fecha_movimiento"},
	{ModuloKey: "comisiones_movimientos", Modulo: "Comisiones por servicio", Categoria: "finanzas", Tabla: "empresa_comisiones_servicio_movimientos", FechaCampo: "fecha_movimiento"},
	{ModuloKey: "compras_documentos", Modulo: "Compras documentales", Categoria: "compras", Tabla: "empresa_compras_documentos", FechaCampo: "fecha_documento"},
	{ModuloKey: "facturacion_documentos", Modulo: "Facturación documental", Categoria: "facturacion", Tabla: "empresa_facturacion_documentos", FechaCampo: "fecha_documento"},
	{ModuloKey: "finanzas_movimientos", Modulo: "Finanzas - Movimientos", Categoria: "finanzas", Tabla: "empresa_finanzas_movimientos", FechaCampo: "fecha_movimiento"},
	{ModuloKey: "cierres_caja", Modulo: "Finanzas - Cierres de caja", Categoria: "finanzas", Tabla: "empresa_cierres_caja", FechaCampo: "fecha_operacion"},
	{ModuloKey: "eventos_contables", Modulo: "Eventos contables", Categoria: "contable", Tabla: "empresa_eventos_contables", FechaCampo: "fecha_evento"},
	{ModuloKey: "asientos_contables", Modulo: "Asientos contables", Categoria: "contable", Tabla: "empresa_asientos_contables", FechaCampo: "fecha_asiento"},
	{ModuloKey: "nomina_liquidaciones", Modulo: "Nómina - Liquidaciones", Categoria: "rrhh", Tabla: "empresa_nomina_liquidaciones", FechaCampo: "fecha_generacion"},
	{ModuloKey: "auditoria", Modulo: "Auditoría empresarial", Categoria: "seguridad", Tabla: "empresa_auditoria_eventos", FechaCampo: "fecha_evento"},
	{ModuloKey: "crm_leads", Modulo: "CRM - Leads", Categoria: "crm", Tabla: "crm_leads", FechaCampo: "fecha_creacion"},
	{ModuloKey: "produccion_ordenes", Modulo: "Producción - Órdenes", Categoria: "produccion", Tabla: "produccion_ordenes", FechaCampo: "fecha_creacion"},
	{ModuloKey: "logistica_envios", Modulo: "Logística - Envíos", Categoria: "logistica", Tabla: "logistica_envios", FechaCampo: "fecha_creacion"},
	{ModuloKey: "integraciones_apis", Modulo: "Integraciones API", Categoria: "integraciones", Tabla: "empresa_integraciones_apis", FechaCampo: "fecha_creacion"},
	{ModuloKey: "integraciones_bancos", Modulo: "Integraciones bancos", Categoria: "integraciones", Tabla: "empresa_integraciones_bancos", FechaCampo: "fecha_creacion"},
	{ModuloKey: "dian_configuracion", Modulo: "DIAN Colombia", Categoria: "facturacion", Tabla: "empresa_dian_configuracion", FechaCampo: "fecha_actualizacion"},
}

func (b *reportesBuilder) buildOperativoModulosResumenDataset() (empresaReporteDataset, error) {
	ds := b.newDataset(reporteDatasetOperativoModulos, []string{
		"modulo_key",
		"modulo",
		"categoria",
		"tabla",
		"registros_totales",
		"registros_activos",
		"registros_rango",
		"ultimo_registro",
		"nota",
	})

	hasDateFilter := strings.TrimSpace(b.desde) != "" || strings.TrimSpace(b.hasta) != ""

	modulosConTabla := 0
	modulosSinTabla := 0
	modulosConDatos := 0
	var registrosTotales int64
	var registrosActivos int64
	var registrosRango int64

	for _, mod := range reporteModulosResumenDefs {
		row := map[string]interface{}{
			"modulo_key":        mod.ModuloKey,
			"modulo":            mod.Modulo,
			"categoria":         mod.Categoria,
			"tabla":             mod.Tabla,
			"registros_totales": int64(0),
			"registros_activos": int64(0),
			"registros_rango":   int64(0),
			"ultimo_registro":   "",
			"nota":              "tabla_no_disponible",
		}

		tablaExiste, err := b.reportesTableExists(mod.Tabla)
		if err != nil {
			return empresaReporteDataset{}, err
		}
		if !tablaExiste {
			modulosSinTabla++
			ds.Rows = append(ds.Rows, row)
			continue
		}
		modulosConTabla++

		columnas, err := b.reportesTableColumns(mod.Tabla)
		if err != nil {
			return empresaReporteDataset{}, err
		}

		if !columnas["empresa_id"] {
			row["nota"] = "sin_columna_empresa_id"
			ds.Rows = append(ds.Rows, row)
			continue
		}

		total, err := b.reportesCountByEmpresa(mod.Tabla, "")
		if err != nil {
			return empresaReporteDataset{}, err
		}

		activos := total
		notas := make([]string, 0, 2)
		if columnas["estado"] {
			activos, err = b.reportesCountByEmpresa(mod.Tabla, "AND LOWER(COALESCE(estado, 'activo')) = 'activo'")
			if err != nil {
				return empresaReporteDataset{}, err
			}
		} else {
			notas = append(notas, "sin_columna_estado")
		}

		rango := total
		ultimoRegistro := ""
		if mod.FechaCampo != "" && columnas[strings.ToLower(strings.TrimSpace(mod.FechaCampo))] {
			ultimoRegistro, err = b.reportesMaxDateByEmpresa(mod.Tabla, mod.FechaCampo)
			if err != nil {
				return empresaReporteDataset{}, err
			}
			if hasDateFilter {
				filtroFecha, argsFecha := reportesBuildDateFilterClause(mod.FechaCampo, b.desde, b.hasta)
				rango, err = b.reportesCountByEmpresa(mod.Tabla, filtroFecha, argsFecha...)
				if err != nil {
					return empresaReporteDataset{}, err
				}
			}
		} else if hasDateFilter {
			rango = 0
			notas = append(notas, "sin_columna_fecha")
		}

		if len(notas) == 0 {
			notas = append(notas, "ok")
		}

		row["registros_totales"] = total
		row["registros_activos"] = activos
		row["registros_rango"] = rango
		row["ultimo_registro"] = ultimoRegistro
		row["nota"] = strings.Join(notas, ",")

		if total > 0 {
			modulosConDatos++
		}
		registrosTotales += total
		registrosActivos += activos
		registrosRango += rango

		ds.Rows = append(ds.Rows, row)
	}

	ds.RowCount = len(ds.Rows)
	ds.Summary["modulos_total"] = ds.RowCount
	ds.Summary["modulos_con_tabla"] = modulosConTabla
	ds.Summary["modulos_sin_tabla"] = modulosSinTabla
	ds.Summary["modulos_con_datos"] = modulosConDatos
	ds.Summary["registros_totales"] = registrosTotales
	ds.Summary["registros_activos"] = registrosActivos
	ds.Summary["registros_rango"] = registrosRango
	ds.Summary["rango_desde"] = reportesFirstNonBlank(strings.TrimSpace(b.desde), "sin_desde")
	ds.Summary["rango_hasta"] = reportesFirstNonBlank(strings.TrimSpace(b.hasta), "sin_hasta")

	return ds, nil
}

func (b *reportesBuilder) reportesTableExists(table string) (bool, error) {
	if !reportesSafeSQLIdentifier(table) {
		return false, fmt.Errorf("tabla invalida")
	}
	var count int
	err := b.db.QueryRow("SELECT COUNT(1) FROM sqlite_master WHERE type = 'table' AND name = ?", table).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (b *reportesBuilder) reportesTableColumns(table string) (map[string]bool, error) {
	if !reportesSafeSQLIdentifier(table) {
		return nil, fmt.Errorf("tabla invalida")
	}

	rows, err := b.db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name string
		var colType string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &colType, &notnull, &dflt, &pk); err != nil {
			return nil, err
		}
		cols[strings.ToLower(strings.TrimSpace(name))] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cols, nil
}

func (b *reportesBuilder) reportesCountByEmpresa(table string, extraWhere string, extraArgs ...interface{}) (int64, error) {
	if !reportesSafeSQLIdentifier(table) {
		return 0, fmt.Errorf("tabla invalida")
	}
	query := "SELECT COUNT(1) FROM " + table + " WHERE empresa_id = ?"
	if strings.TrimSpace(extraWhere) != "" {
		query += " " + strings.TrimSpace(extraWhere)
	}

	args := make([]interface{}, 0, 1+len(extraArgs))
	args = append(args, b.empresaID)
	args = append(args, extraArgs...)

	var total int64
	if err := b.db.QueryRow(query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (b *reportesBuilder) reportesSumByEmpresa(table string, sumColumn string, extraWhere string, extraArgs ...interface{}) (float64, error) {
	if !reportesSafeSQLIdentifier(table) || !reportesSafeSQLIdentifier(sumColumn) {
		return 0, fmt.Errorf("identificador invalido")
	}
	query := "SELECT COALESCE(SUM(COALESCE(" + sumColumn + ", 0)), 0) FROM " + table + " WHERE empresa_id = ?"
	if strings.TrimSpace(extraWhere) != "" {
		query += " " + strings.TrimSpace(extraWhere)
	}

	args := make([]interface{}, 0, 1+len(extraArgs))
	args = append(args, b.empresaID)
	args = append(args, extraArgs...)

	var total float64
	if err := b.db.QueryRow(query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return reportesRound(total), nil
}

func (b *reportesBuilder) reportesMaxDateByEmpresa(table string, dateColumn string) (string, error) {
	if !reportesSafeSQLIdentifier(table) || !reportesSafeSQLIdentifier(dateColumn) {
		return "", fmt.Errorf("identificador invalido")
	}
	query := "SELECT COALESCE(MAX(substr(COALESCE(" + dateColumn + ", ''), 1, 19)), '') FROM " + table + " WHERE empresa_id = ?"
	var value string
	if err := b.db.QueryRow(query, b.empresaID).Scan(&value); err != nil {
		return "", err
	}
	return strings.TrimSpace(value), nil
}

func reportesBuildDateFilterClause(dateColumn string, desde string, hasta string) (string, []interface{}) {
	dateColumn = strings.TrimSpace(dateColumn)
	if !reportesSafeSQLIdentifier(dateColumn) {
		return "", nil
	}
	desde = strings.TrimSpace(desde)
	hasta = strings.TrimSpace(hasta)
	if desde == "" && hasta == "" {
		return "", nil
	}

	dateExpr := "substr(COALESCE(" + dateColumn + ", ''), 1, 10)"
	parts := make([]string, 0, 2)
	args := make([]interface{}, 0, 2)
	if desde != "" {
		parts = append(parts, dateExpr+" >= ?")
		args = append(args, desde)
	}
	if hasta != "" {
		parts = append(parts, dateExpr+" <= ?")
		args = append(args, hasta)
	}

	if len(parts) == 0 {
		return "", nil
	}
	return "AND " + strings.Join(parts, " AND "), args
}

func reportesBuildStateFilterClause(stateColumn string, states []string) (string, []interface{}) {
	stateColumn = strings.TrimSpace(stateColumn)
	if !reportesSafeSQLIdentifier(stateColumn) {
		return "", nil
	}

	normalized := make([]string, 0, len(states))
	for _, state := range states {
		s := strings.ToLower(strings.TrimSpace(state))
		if s == "" {
			continue
		}
		normalized = append(normalized, s)
	}
	if len(normalized) == 0 {
		return "", nil
	}

	placeholders := make([]string, 0, len(normalized))
	args := make([]interface{}, 0, len(normalized))
	for _, state := range normalized {
		placeholders = append(placeholders, "?")
		args = append(args, state)
	}

	clause := "AND LOWER(COALESCE(" + stateColumn + ", '')) IN (" + strings.Join(placeholders, ", ") + ")"
	return clause, args
}

func reportesSafeSQLIdentifier(v string) bool {
	v = strings.TrimSpace(v)
	if v == "" {
		return false
	}
	for _, ch := range v {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' {
			continue
		}
		return false
	}
	return true
}

func (b *reportesBuilder) getVentasCerradasFiltradas() ([]dbpkg.CarritoCompra, error) {
	if b.ventasCache != nil {
		return b.ventasCache, nil
	}
	carritos, err := dbpkg.GetCarritosCompraByEmpresa(b.db, b.empresaID, b.includeInactive, "")
	if err != nil {
		return nil, err
	}
	filtered := make([]dbpkg.CarritoCompra, 0)
	for _, item := range carritos {
		estadoCarrito := strings.ToLower(strings.TrimSpace(item.EstadoCarrito))
		if estadoCarrito != "cerrado" {
			continue
		}
		fechaPago := reportesFirstNonBlank(item.PagadoEn, item.FechaActualizacion, item.FechaCreacion)
		if !reportesDateWithinRange(fechaPago, b.desde, b.hasta) {
			continue
		}
		filtered = append(filtered, item)
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		return reportesDateUnix(reportesFirstNonBlank(filtered[i].PagadoEn, filtered[i].FechaActualizacion, filtered[i].FechaCreacion)) >
			reportesDateUnix(reportesFirstNonBlank(filtered[j].PagadoEn, filtered[j].FechaActualizacion, filtered[j].FechaCreacion))
	})
	if len(filtered) > b.maxRows {
		filtered = filtered[:b.maxRows]
	}
	b.ventasCache = filtered
	return filtered, nil
}

func (b *reportesBuilder) ensureItemsForCarritos(carritos []dbpkg.CarritoCompra) error {
	for _, carrito := range carritos {
		carritoID := carrito.ID
		if _, ok := b.itemsCache[carritoID]; ok {
			continue
		}
		items, err := dbpkg.GetCarritoCompraItems(b.db, b.empresaID, carritoID, true)
		if err != nil {
			return err
		}
		b.itemsCache[carritoID] = items
	}
	return nil
}

func (b *reportesBuilder) buildOperativoVentasDetalleDataset() (empresaReporteDataset, error) {
	ventas, err := b.getVentasCerradasFiltradas()
	if err != nil {
		return empresaReporteDataset{}, err
	}
	ds := b.newDataset(reporteDatasetOperativoVentasDetalle, []string{
		"fecha_pago",
		"carrito_id",
		"carrito_codigo",
		"cliente",
		"canal_venta",
		"items",
		"total",
		"estado_venta",
	})
	totalVentas := 0.0
	for _, venta := range ventas {
		total := reportesVentaTotal(venta)
		totalVentas += total
		ds.Rows = append(ds.Rows, map[string]interface{}{
			"fecha_pago":     reportesFirstNonBlank(venta.PagadoEn, venta.FechaActualizacion, venta.FechaCreacion),
			"carrito_id":     venta.ID,
			"carrito_codigo": reportesFirstNonBlank(venta.Codigo, venta.Nombre),
			"cliente":        reportesFirstNonBlank(venta.ClienteNombre, "Sin cliente"),
			"canal_venta":    reportesFirstNonBlank(venta.CanalVenta, "mostrador"),
			"items":          venta.ItemCount,
			"total":          total,
			"estado_venta":   reportesFirstNonBlank(venta.EstadoVenta, venta.EstadoCarrito),
		})
	}
	ds.RowCount = len(ds.Rows)
	ds.Summary["ventas"] = ds.RowCount
	ds.Summary["ingresos"] = reportesRound(totalVentas)
	if ds.RowCount > 0 {
		ds.Summary["ticket_promedio"] = reportesRound(totalVentas / float64(ds.RowCount))
	}
	return ds, nil
}

func (b *reportesBuilder) buildOperativoTurnoDataset() (empresaReporteDataset, error) {
	ds := b.newDataset(reporteDatasetOperativoTurno, []string{
		"fecha_operacion",
		"caja_codigo",
		"turno",
		"usuario_turno",
		"carrito_id",
		"carrito_codigo",
		"cliente",
		"activado_en",
		"pagado_en",
		"metodo_pago",
		"total_productos",
		"total_servicios",
		"total_otros",
		"total_carrito",
	})

	usuarioFiltro := strings.ToLower(strings.TrimSpace(b.usuario))
	cajaFiltro := strings.ToUpper(strings.TrimSpace(b.cajaCodigo))
	turnoFiltro := strings.ToLower(strings.TrimSpace(b.turno))

	cierres, err := dbpkg.ListEmpresaCierresCaja(b.db, b.empresaID, dbpkg.EmpresaCierreCajaFilter{
		CajaCodigo:      cajaFiltro,
		Desde:           b.desde,
		Hasta:           b.hasta,
		IncludeInactive: b.includeInactive,
		Limit:           1000,
	})
	if err != nil {
		return empresaReporteDataset{}, err
	}

	cierresFiltrados := make([]dbpkg.EmpresaCierreCaja, 0, len(cierres))
	for _, cierre := range cierres {
		if turnoFiltro != "" && strings.ToLower(strings.TrimSpace(cierre.Turno)) != turnoFiltro {
			continue
		}
		if usuarioFiltro != "" && !reportesStringEqualsFoldAny(usuarioFiltro, cierre.UsuarioCreador, cierre.CerradoPor, cierre.AprobadoPor) {
			continue
		}
		cierresFiltrados = append(cierresFiltrados, cierre)
	}

	var cierreSeleccionado *dbpkg.EmpresaCierreCaja
	if b.cierreID > 0 {
		for i := range cierresFiltrados {
			if cierresFiltrados[i].ID == b.cierreID {
				cierreSeleccionado = &cierresFiltrados[i]
				break
			}
		}
		if cierreSeleccionado == nil {
			return empresaReporteDataset{}, fmt.Errorf("cierre_id no encontrado para los filtros aplicados")
		}
	}

	cierresAplicados := cierresFiltrados
	if cierreSeleccionado != nil {
		cierresAplicados = []dbpkg.EmpresaCierreCaja{*cierreSeleccionado}
	}

	rangoDesde := strings.TrimSpace(b.desde)
	rangoHasta := strings.TrimSpace(b.hasta)
	if cierreSeleccionado != nil {
		if rangoDesde == "" {
			rangoDesde = reportesNormalizeDatePart(reportesFirstNonBlank(cierreSeleccionado.FechaApertura, cierreSeleccionado.FechaOperacion))
		}
		if rangoHasta == "" {
			rangoHasta = reportesNormalizeDatePart(reportesFirstNonBlank(cierreSeleccionado.FechaCierre, cierreSeleccionado.FechaOperacion))
		}
	}

	if (rangoDesde == "" || rangoHasta == "") && len(cierresAplicados) > 0 {
		minDate := ""
		maxDate := ""
		for _, cierre := range cierresAplicados {
			fechaOperacion := reportesNormalizeDatePart(cierre.FechaOperacion)
			if fechaOperacion == "" {
				continue
			}
			if minDate == "" || fechaOperacion < minDate {
				minDate = fechaOperacion
			}
			if maxDate == "" || fechaOperacion > maxDate {
				maxDate = fechaOperacion
			}
		}
		if rangoDesde == "" {
			rangoDesde = minDate
		}
		if rangoHasta == "" {
			rangoHasta = maxDate
		}
	}

	if rangoDesde != "" {
		ds.Desde = rangoDesde
	}
	if rangoHasta != "" {
		ds.Hasta = rangoHasta
	}

	cierresPorFecha := make(map[string]dbpkg.EmpresaCierreCaja)
	fechasCierreSet := make(map[string]struct{})
	aperturaEfectivo := 0.0
	gastosCierres := 0.0
	efectivoTeoricoCierre := 0.0
	efectivoFisicoCierre := 0.0
	diferenciaCajaCierre := 0.0
	monedaTurno := "COP"
	for _, cierre := range cierresAplicados {
		fechaOperacion := reportesNormalizeDatePart(cierre.FechaOperacion)
		if fechaOperacion != "" {
			fechasCierreSet[fechaOperacion] = struct{}{}
			if _, ok := cierresPorFecha[fechaOperacion]; !ok {
				cierresPorFecha[fechaOperacion] = cierre
			}
		}
		aperturaEfectivo += cierre.AperturaMonto
		gastosCierres += cierre.EgresosEfectivo + cierre.RetirosEfectivo
		efectivoTeoricoCierre += cierre.CajaTeorica
		efectivoFisicoCierre += cierre.CajaFisica
		diferenciaCajaCierre += cierre.DiferenciaCaja
		if strings.TrimSpace(monedaTurno) == "" {
			monedaTurno = strings.TrimSpace(cierre.Moneda)
		}
	}

	aperturaSeleccionadaUnix := int64(0)
	cierreSeleccionadaUnix := int64(0)
	if cierreSeleccionado != nil {
		aperturaSeleccionadaUnix = reportesDateUnix(cierreSeleccionado.FechaApertura)
		cierreSeleccionadaUnix = reportesDateUnix(cierreSeleccionado.FechaCierre)
	}

	carritos, err := dbpkg.GetCarritosCompraByEmpresa(b.db, b.empresaID, true, "")
	if err != nil {
		return empresaReporteDataset{}, err
	}

	ventas := make([]dbpkg.CarritoCompra, 0)
	requiereRelacionTurno := b.cierreID > 0 || cajaFiltro != "" || turnoFiltro != ""
	for _, carrito := range carritos {
		if strings.ToLower(strings.TrimSpace(carrito.EstadoCarrito)) != "cerrado" {
			continue
		}
		fechaPago := reportesFirstNonBlank(carrito.PagadoEn, carrito.FechaActualizacion, carrito.FechaCreacion)
		if !reportesDateWithinRange(fechaPago, rangoDesde, rangoHasta) {
			continue
		}
		if usuarioFiltro != "" && strings.ToLower(strings.TrimSpace(carrito.UsuarioCreador)) != usuarioFiltro {
			continue
		}

		if requiereRelacionTurno {
			fechaOperacion := reportesNormalizeDatePart(fechaPago)
			if fechaOperacion == "" {
				continue
			}
			if _, ok := fechasCierreSet[fechaOperacion]; !ok {
				continue
			}
		}

		if cierreSeleccionado != nil {
			fechaPagoUnix := reportesDateUnix(fechaPago)
			if aperturaSeleccionadaUnix > 0 && fechaPagoUnix > 0 && fechaPagoUnix < aperturaSeleccionadaUnix {
				continue
			}
			if cierreSeleccionadaUnix > 0 && fechaPagoUnix > 0 && fechaPagoUnix > cierreSeleccionadaUnix {
				continue
			}
		}

		ventas = append(ventas, carrito)
	}

	sort.SliceStable(ventas, func(i, j int) bool {
		return reportesDateUnix(reportesFirstNonBlank(ventas[i].PagadoEn, ventas[i].FechaActualizacion, ventas[i].FechaCreacion)) >
			reportesDateUnix(reportesFirstNonBlank(ventas[j].PagadoEn, ventas[j].FechaActualizacion, ventas[j].FechaCreacion))
	})
	if len(ventas) > b.maxRows {
		ventas = ventas[:b.maxRows]
	}

	if err := b.ensureItemsForCarritos(ventas); err != nil {
		return empresaReporteDataset{}, err
	}

	totalProductos := 0.0
	totalServicios := 0.0
	totalOtros := 0.0
	totalVentas := 0.0
	ventasEfectivo := 0.0
	ventasNoEfectivo := 0.0

	for _, venta := range ventas {
		items := b.itemsCache[venta.ID]
		totalProductosCarrito := 0.0
		totalServiciosCarrito := 0.0
		totalOtrosCarrito := 0.0
		for _, item := range items {
			if !reportesEstadoActivo(item.Estado) {
				continue
			}
			totalLinea := reportesRound(item.TotalLinea)
			switch strings.ToLower(strings.TrimSpace(item.TipoItem)) {
			case "servicio":
				totalServiciosCarrito += totalLinea
			case "producto", "combo":
				totalProductosCarrito += totalLinea
			default:
				totalOtrosCarrito += totalLinea
			}
		}

		totalCarrito := reportesVentaTotal(venta)
		if reportesRound(totalProductosCarrito+totalServiciosCarrito+totalOtrosCarrito) == 0 && totalCarrito > 0 {
			totalOtrosCarrito = totalCarrito
		}

		totalProductos += totalProductosCarrito
		totalServicios += totalServiciosCarrito
		totalOtros += totalOtrosCarrito
		totalVentas += totalCarrito

		metodoPago := dbpkg.NormalizeMetodoPagoCarrito(venta.MetodoPago)
		if metodoPago == "" {
			metodoPago = strings.ToLower(strings.TrimSpace(venta.MetodoPago))
		}
		if metodoPago == "" {
			metodoPago = "efectivo"
		}
		if metodoPago == "efectivo" {
			ventasEfectivo += totalCarrito
		} else {
			ventasNoEfectivo += totalCarrito
		}

		fechaPago := reportesFirstNonBlank(venta.PagadoEn, venta.FechaActualizacion, venta.FechaCreacion)
		fechaOperacion := reportesNormalizeDatePart(fechaPago)
		cierreMeta, existeCierreMeta := cierresPorFecha[fechaOperacion]
		if cierreSeleccionado != nil {
			cierreMeta = *cierreSeleccionado
			existeCierreMeta = true
		}

		cajaCodigo := reportesFirstNonBlank(cajaFiltro, "-")
		turno := reportesFirstNonBlank(turnoFiltro, "-")
		usuarioTurno := reportesFirstNonBlank(venta.UsuarioCreador, "sistema")
		if existeCierreMeta {
			cajaCodigo = reportesFirstNonBlank(cierreMeta.CajaCodigo, cajaCodigo)
			turno = reportesFirstNonBlank(cierreMeta.Turno, turno)
			usuarioTurno = reportesFirstNonBlank(cierreMeta.UsuarioCreador, cierreMeta.CerradoPor, usuarioTurno)
		}

		ds.Rows = append(ds.Rows, map[string]interface{}{
			"fecha_operacion": fechaOperacion,
			"caja_codigo":     cajaCodigo,
			"turno":           turno,
			"usuario_turno":   usuarioTurno,
			"carrito_id":      venta.ID,
			"carrito_codigo":  reportesFirstNonBlank(venta.Codigo, venta.Nombre),
			"cliente":         reportesFirstNonBlank(venta.ClienteNombre, "Sin cliente"),
			"activado_en":     reportesFirstNonBlank(venta.ActivadoEn, venta.FechaCreacion),
			"pagado_en":       fechaPago,
			"metodo_pago":     metodoPago,
			"total_productos": reportesRound(totalProductosCarrito),
			"total_servicios": reportesRound(totalServiciosCarrito),
			"total_otros":     reportesRound(totalOtrosCarrito),
			"total_carrito":   reportesRound(totalCarrito),
		})
	}

	movimientosEgreso, err := dbpkg.ListEmpresaFinanzasMovimientos(b.db, b.empresaID, dbpkg.EmpresaFinanzasMovimientoFilter{
		Tipo:            "egreso",
		Desde:           rangoDesde,
		Hasta:           rangoHasta,
		IncludeInactive: b.includeInactive,
		Limit:           b.maxRows * 10,
	})
	if err != nil {
		return empresaReporteDataset{}, err
	}
	gastosMovimientos := 0.0
	for _, mov := range movimientosEgreso {
		if usuarioFiltro != "" && strings.ToLower(strings.TrimSpace(mov.UsuarioCreador)) != usuarioFiltro {
			continue
		}
		gastosMovimientos += reportesMovimientoTotalNeto(mov)
	}

	gastosTurno := gastosMovimientos
	if len(cierresAplicados) > 0 {
		gastosTurno = gastosCierres
	}
	efectivoCalculado := aperturaEfectivo + ventasEfectivo - gastosTurno
	efectivoDeberiaHaber := efectivoCalculado
	if len(cierresAplicados) > 0 {
		efectivoDeberiaHaber = efectivoTeoricoCierre
	}

	ds.RowCount = len(ds.Rows)
	ds.Summary["filtro_usuario"] = reportesFirstNonBlank(strings.TrimSpace(b.usuario), "todos")
	ds.Summary["filtro_caja_codigo"] = reportesFirstNonBlank(cajaFiltro, "todas")
	ds.Summary["filtro_turno"] = reportesFirstNonBlank(turnoFiltro, "todos")
	ds.Summary["filtro_cierre_id"] = b.cierreID
	ds.Summary["cierres_relacionados"] = len(cierresAplicados)
	ds.Summary["ventas_carritos"] = ds.RowCount
	ds.Summary["ventas_productos"] = reportesRound(totalProductos)
	ds.Summary["ventas_servicios"] = reportesRound(totalServicios)
	ds.Summary["ventas_otros"] = reportesRound(totalOtros)
	ds.Summary["ventas_totales"] = reportesRound(totalVentas)
	ds.Summary["ventas_efectivo"] = reportesRound(ventasEfectivo)
	ds.Summary["ventas_no_efectivo"] = reportesRound(ventasNoEfectivo)
	ds.Summary["gastos_movimientos"] = reportesRound(gastosMovimientos)
	ds.Summary["gastos_cierres_caja"] = reportesRound(gastosCierres)
	ds.Summary["gastos_turno"] = reportesRound(gastosTurno)
	ds.Summary["apertura_efectivo"] = reportesRound(aperturaEfectivo)
	ds.Summary["efectivo_calculado"] = reportesRound(efectivoCalculado)
	ds.Summary["efectivo_caja_teorica"] = reportesRound(efectivoTeoricoCierre)
	ds.Summary["efectivo_caja_fisica"] = reportesRound(efectivoFisicoCierre)
	ds.Summary["diferencia_caja"] = reportesRound(diferenciaCajaCierre)
	ds.Summary["efectivo_deberia_haber"] = reportesRound(efectivoDeberiaHaber)
	ds.Summary["moneda"] = reportesFirstNonBlank(strings.TrimSpace(monedaTurno), "COP")

	if cierreSeleccionado != nil {
		ds.Summary["cierre_estado"] = cierreSeleccionado.EstadoCierre
		ds.Summary["cierre_fecha_operacion"] = cierreSeleccionado.FechaOperacion
	}

	return ds, nil
}

type reporteReservaOcupacionAgg struct {
	EstacionID        int64
	EstacionCodigo    string
	EstacionNombre    string
	ReservasTotales   int64
	ReservasConfirm   int64
	ReservasPend      int64
	ReservasCanc      int64
	ReservasExp       int64
	HuespedesTotales  int64
	MontoTotal        float64
	MontoConfirmado   float64
	OcupacionDias     float64
	UltimaFechaInicio string
	ultimaFechaUnix   int64
}

func (b *reportesBuilder) buildOperativoReservasOcupacionDataset() (empresaReporteDataset, error) {
	ds := b.newDataset(reporteDatasetOperativoReservas, []string{
		"estacion_id",
		"estacion_codigo",
		"estacion_nombre",
		"reservas_totales",
		"reservas_confirmadas",
		"reservas_pendientes",
		"reservas_canceladas",
		"reservas_expiradas",
		"cumplimiento_pct",
		"ocupacion_pct",
		"huespedes_totales",
		"monto_total",
		"monto_confirmado",
		"ultima_fecha_entrada",
	})

	tablaExiste, err := b.reportesTableExists("reservas_hotel")
	if err != nil {
		return empresaReporteDataset{}, err
	}
	if !tablaExiste {
		ds.Summary["nota"] = "tabla_reservas_hotel_no_disponible"
		ds.Summary["estaciones_con_reservas"] = 0
		ds.Summary["reservas_totales"] = int64(0)
		return ds, nil
	}

	reservas, err := dbpkg.ListReservasHotelByEmpresa(b.db, b.empresaID, dbpkg.ReservaHotelFilter{
		FechaDesde: b.desde,
		FechaHasta: b.hasta,
		Limit:      1000,
		Offset:     0,
	})
	if err != nil {
		return empresaReporteDataset{}, err
	}

	periodoDesde, periodoHasta := reportesResolveReservaPeriodo(b.desde, b.hasta, reservas)
	periodoDias := reportesDaysInclusive(periodoDesde, periodoHasta)
	if periodoDias <= 0 {
		periodoDias = 1
	}

	aggregates := make(map[string]*reporteReservaOcupacionAgg)
	var totalReservas int64
	var totalConfirmadas int64
	var totalPendientes int64
	var totalCanceladas int64
	var totalExpiradas int64
	var totalHuespedes int64
	var ingresosPotenciales float64
	var ingresosConfirmados float64

	for _, reserva := range reservas {
		if !reportesEstadoActivo(reserva.Estado) {
			continue
		}

		key := strings.TrimSpace(reserva.EstacionCodigo)
		if key == "" {
			key = "estacion_" + strconv.FormatInt(reserva.EstacionID, 10)
		}
		current, ok := aggregates[key]
		if !ok {
			current = &reporteReservaOcupacionAgg{
				EstacionID:     reserva.EstacionID,
				EstacionCodigo: strings.TrimSpace(reserva.EstacionCodigo),
				EstacionNombre: reportesFirstNonBlank(strings.TrimSpace(reserva.EstacionNombre), strings.TrimSpace(reserva.EstacionCodigo), "Estacion "+strconv.FormatInt(reserva.EstacionID, 10)),
			}
			aggregates[key] = current
		}

		current.ReservasTotales++
		totalReservas++

		estadoReserva := strings.ToLower(strings.TrimSpace(reserva.EstadoReserva))
		switch estadoReserva {
		case "confirmada":
			current.ReservasConfirm++
			totalConfirmadas++
		case "cancelada":
			current.ReservasCanc++
			totalCanceladas++
		case "expirada":
			current.ReservasExp++
			totalExpiradas++
		default:
			current.ReservasPend++
			totalPendientes++
		}

		current.HuespedesTotales += reserva.CantidadHuespedes
		totalHuespedes += reserva.CantidadHuespedes

		current.MontoTotal += reportesRound(reserva.MontoTotal)
		ingresosPotenciales += reportesRound(reserva.MontoTotal)
		if estadoReserva == "confirmada" {
			current.MontoConfirmado += reportesRound(reserva.MontoTotal)
			ingresosConfirmados += reportesRound(reserva.MontoTotal)
		}

		ocupacionDias := reportesReservaOverlapDays(reserva.FechaEntrada, reserva.FechaSalida, periodoDesde, periodoHasta)
		if ocupacionDias > 0 {
			current.OcupacionDias += ocupacionDias
		}

		fechaInicio := reportesFirstNonBlank(reserva.FechaEntrada, reserva.FechaCreacion)
		fechaUnix := reportesDateUnix(fechaInicio)
		if fechaUnix > current.ultimaFechaUnix {
			current.ultimaFechaUnix = fechaUnix
			current.UltimaFechaInicio = fechaInicio
		}
	}

	rows := make([]*reporteReservaOcupacionAgg, 0, len(aggregates))
	for _, item := range aggregates {
		rows = append(rows, item)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].OcupacionDias == rows[j].OcupacionDias {
			if rows[i].ReservasTotales == rows[j].ReservasTotales {
				return rows[i].EstacionNombre < rows[j].EstacionNombre
			}
			return rows[i].ReservasTotales > rows[j].ReservasTotales
		}
		return rows[i].OcupacionDias > rows[j].OcupacionDias
	})

	ocupacionAcumulada := 0.0
	for _, row := range rows {
		cumplimiento := 0.0
		if row.ReservasTotales > 0 {
			cumplimiento = reportesRound((float64(row.ReservasConfirm) * 100.0) / float64(row.ReservasTotales))
		}
		ocupacionPct := 0.0
		if periodoDias > 0 {
			ocupacionPct = reportesRound((row.OcupacionDias * 100.0) / periodoDias)
			if ocupacionPct > 100 {
				ocupacionPct = 100
			}
		}
		ocupacionAcumulada += ocupacionPct

		ds.Rows = append(ds.Rows, map[string]interface{}{
			"estacion_id":          row.EstacionID,
			"estacion_codigo":      row.EstacionCodigo,
			"estacion_nombre":      row.EstacionNombre,
			"reservas_totales":     row.ReservasTotales,
			"reservas_confirmadas": row.ReservasConfirm,
			"reservas_pendientes":  row.ReservasPend,
			"reservas_canceladas":  row.ReservasCanc,
			"reservas_expiradas":   row.ReservasExp,
			"cumplimiento_pct":     cumplimiento,
			"ocupacion_pct":        ocupacionPct,
			"huespedes_totales":    row.HuespedesTotales,
			"monto_total":          reportesRound(row.MontoTotal),
			"monto_confirmado":     reportesRound(row.MontoConfirmado),
			"ultima_fecha_entrada": row.UltimaFechaInicio,
		})
	}

	ds.RowCount = len(ds.Rows)
	ds.Summary["periodo_desde"] = reportesFirstNonBlank(periodoDesde, "sin_desde")
	ds.Summary["periodo_hasta"] = reportesFirstNonBlank(periodoHasta, "sin_hasta")
	ds.Summary["periodo_dias"] = reportesRound(periodoDias)
	ds.Summary["estaciones_con_reservas"] = ds.RowCount
	ds.Summary["reservas_totales"] = totalReservas
	ds.Summary["reservas_confirmadas"] = totalConfirmadas
	ds.Summary["reservas_pendientes"] = totalPendientes
	ds.Summary["reservas_canceladas"] = totalCanceladas
	ds.Summary["reservas_expiradas"] = totalExpiradas
	ds.Summary["huespedes_totales"] = totalHuespedes
	ds.Summary["ingresos_potenciales"] = reportesRound(ingresosPotenciales)
	ds.Summary["ingresos_confirmados"] = reportesRound(ingresosConfirmados)

	cumplimientoGlobal := 0.0
	if totalReservas > 0 {
		cumplimientoGlobal = reportesRound((float64(totalConfirmadas) * 100.0) / float64(totalReservas))
	}
	ds.Summary["cumplimiento_global_pct"] = cumplimientoGlobal

	ocupacionPromedio := 0.0
	if ds.RowCount > 0 {
		ocupacionPromedio = reportesRound(ocupacionAcumulada / float64(ds.RowCount))
	}
	ds.Summary["ocupacion_promedio_pct"] = ocupacionPromedio

	return ds, nil
}

func (b *reportesBuilder) buildOperativoTarifasIngresosDataset() (empresaReporteDataset, error) {
	ds := b.newDataset(reporteDatasetOperativoTarifas, []string{
		"modelo_tarifa",
		"carritos_cerrados",
		"ingresos_totales",
		"ticket_promedio",
		"estaciones_distintas",
		"tarifas_configuradas",
	})

	ventas, err := b.getVentasCerradasFiltradas()
	if err != nil {
		return empresaReporteDataset{}, err
	}

	tarifasDiaByEstacion := make(map[int64]struct{})
	tarifasMinByEstacion := make(map[int64]struct{})

	existeTarifaDia, err := b.reportesTableExists("empresa_tarifas_por_dia")
	if err != nil {
		return empresaReporteDataset{}, err
	}
	if existeTarifaDia {
		tarifasDia, err := dbpkg.ListEmpresaTarifasPorDia(b.db, b.empresaID, dbpkg.EmpresaTarifaPorDiaFilter{IncludeInactive: false, Limit: 2000})
		if err != nil {
			return empresaReporteDataset{}, err
		}
		for _, tarifa := range tarifasDia {
			if tarifa.EstacionID > 0 && reportesEstadoActivo(tarifa.Estado) {
				tarifasDiaByEstacion[tarifa.EstacionID] = struct{}{}
			}
		}
	}

	existeTarifaMin, err := b.reportesTableExists("empresa_tarifas_por_minutos")
	if err != nil {
		return empresaReporteDataset{}, err
	}
	if existeTarifaMin {
		tarifasMin, err := dbpkg.ListEmpresaTarifasPorMinutos(b.db, b.empresaID, dbpkg.EmpresaTarifaPorMinutosFilter{IncludeInactive: false, Limit: 2000})
		if err != nil {
			return empresaReporteDataset{}, err
		}
		for _, tarifa := range tarifasMin {
			if tarifa.EstacionID > 0 && reportesEstadoActivo(tarifa.Estado) {
				tarifasMinByEstacion[tarifa.EstacionID] = struct{}{}
			}
		}
	}

	type agg struct {
		carritos   int64
		ingresos   float64
		estaciones map[int64]struct{}
	}
	aggByModel := map[string]*agg{
		"tarifa_por_dia":     {estaciones: make(map[int64]struct{})},
		"tarifa_por_minutos": {estaciones: make(map[int64]struct{})},
		"sin_modelo":         {estaciones: make(map[int64]struct{})},
	}

	var carritosConModelo int64
	for _, venta := range ventas {
		modelo := "sin_modelo"
		estacionID := reportesParseEstacionID(venta.ReferenciaExterna, venta.Codigo, b.empresaID)
		if estacionID > 0 {
			if _, ok := tarifasDiaByEstacion[estacionID]; ok {
				modelo = "tarifa_por_dia"
			} else if _, ok := tarifasMinByEstacion[estacionID]; ok {
				modelo = "tarifa_por_minutos"
			}
		}

		current := aggByModel[modelo]
		current.carritos++
		totalVenta := reportesVentaTotal(venta)
		current.ingresos += totalVenta
		if estacionID > 0 {
			current.estaciones[estacionID] = struct{}{}
		}
		if modelo != "sin_modelo" {
			carritosConModelo++
		}
	}

	models := []string{"tarifa_por_dia", "tarifa_por_minutos", "sin_modelo"}
	var ingresosTotal float64
	for _, model := range models {
		current := aggByModel[model]
		tarifasConfiguradas := 0
		switch model {
		case "tarifa_por_dia":
			tarifasConfiguradas = len(tarifasDiaByEstacion)
		case "tarifa_por_minutos":
			tarifasConfiguradas = len(tarifasMinByEstacion)
		}

		if current.carritos == 0 && tarifasConfiguradas == 0 && model != "sin_modelo" {
			continue
		}

		ticketPromedio := 0.0
		if current.carritos > 0 {
			ticketPromedio = reportesRound(current.ingresos / float64(current.carritos))
		}
		ingresosTotal += current.ingresos

		ds.Rows = append(ds.Rows, map[string]interface{}{
			"modelo_tarifa":        model,
			"carritos_cerrados":    current.carritos,
			"ingresos_totales":     reportesRound(current.ingresos),
			"ticket_promedio":      ticketPromedio,
			"estaciones_distintas": len(current.estaciones),
			"tarifas_configuradas": tarifasConfiguradas,
		})
	}

	ds.RowCount = len(ds.Rows)
	ds.Summary["carritos_cerrados_total"] = int64(len(ventas))
	ds.Summary["carritos_con_modelo"] = carritosConModelo
	ds.Summary["carritos_sin_modelo"] = int64(len(ventas)) - carritosConModelo
	ds.Summary["ingresos_total"] = reportesRound(ingresosTotal)
	ds.Summary["ingresos_tarifa_por_dia"] = reportesRound(aggByModel["tarifa_por_dia"].ingresos)
	ds.Summary["ingresos_tarifa_por_minutos"] = reportesRound(aggByModel["tarifa_por_minutos"].ingresos)
	ds.Summary["ingresos_sin_modelo"] = reportesRound(aggByModel["sin_modelo"].ingresos)
	ds.Summary["tarifas_por_dia_configuradas"] = len(tarifasDiaByEstacion)
	ds.Summary["tarifas_por_minutos_configuradas"] = len(tarifasMinByEstacion)

	cobertura := 0.0
	if len(ventas) > 0 {
		cobertura = reportesRound((float64(carritosConModelo) * 100.0) / float64(len(ventas)))
	}
	ds.Summary["cobertura_modelo_tarifa_pct"] = cobertura

	return ds, nil
}

func (b *reportesBuilder) buildOperativoTarifasComparativoEstacionesDataset() (empresaReporteDataset, error) {
	ds := b.newDataset(reporteDatasetOperativoTarifasComparativo, []string{
		"estacion_id",
		"estacion_codigo",
		"estacion_nombre",
		"ventas_cerradas",
		"tarifa_valor_dia",
		"ingreso_esperado",
		"ingreso_real",
		"desviacion_monto",
		"cumplimiento_pct",
		"dias_equivalentes_esperados",
		"minutos_prorrateo_fuera_ventana",
		"ventas_sin_base_calculo",
	})

	ventas, err := b.getVentasCerradasFiltradas()
	if err != nil {
		return empresaReporteDataset{}, err
	}

	existeTarifaDia, err := b.reportesTableExists("empresa_tarifas_por_dia")
	if err != nil {
		return empresaReporteDataset{}, err
	}
	if !existeTarifaDia {
		ds.Summary["nota"] = "tabla_empresa_tarifas_por_dia_no_disponible"
		ds.Summary["ventas_evaluadas"] = 0
		ds.Summary["estaciones_comparadas"] = 0
		return ds, nil
	}

	tarifasDia, err := dbpkg.ListEmpresaTarifasPorDia(b.db, b.empresaID, dbpkg.EmpresaTarifaPorDiaFilter{IncludeInactive: false, Limit: 2000})
	if err != nil {
		return empresaReporteDataset{}, err
	}
	tarifaByStation := make(map[int64]dbpkg.EmpresaTarifaPorDia)
	for _, tarifa := range tarifasDia {
		if tarifa.EstacionID <= 0 || !reportesEstadoActivo(tarifa.Estado) {
			continue
		}
		if _, exists := tarifaByStation[tarifa.EstacionID]; !exists {
			tarifaByStation[tarifa.EstacionID] = tarifa
		}
	}

	type tarifaComparativoAgg struct {
		estacionID       int64
		estacionCodigo   string
		estacionNombre   string
		tarifaValorDia   float64
		ventasCerradas   int64
		ingresoEsperado  float64
		ingresoReal      float64
		diasEquivalentes float64
		minutosProrrateo int64
		ventasSinBase    int64
	}

	aggByStation := make(map[int64]*tarifaComparativoAgg)
	for _, venta := range ventas {
		estacionID := reportesParseEstacionID(venta.ReferenciaExterna, venta.Codigo, b.empresaID)
		if estacionID <= 0 {
			continue
		}

		tarifa, hasTarifa := tarifaByStation[estacionID]
		if !hasTarifa {
			continue
		}

		row, exists := aggByStation[estacionID]
		if !exists {
			row = &tarifaComparativoAgg{
				estacionID:     estacionID,
				estacionCodigo: reportesFirstNonBlank(strings.TrimSpace(tarifa.EstacionCodigo), strings.TrimSpace(venta.Codigo), fmt.Sprintf("EST-%d-%d", b.empresaID, estacionID)),
				estacionNombre: reportesFirstNonBlank(strings.TrimSpace(tarifa.EstacionNombre), strings.TrimSpace(venta.Nombre), fmt.Sprintf("Estacion %d", estacionID)),
				tarifaValorDia: reportesRound(tarifa.ValorDia),
			}
			aggByStation[estacionID] = row
		}

		row.ventasCerradas++
		row.ingresoReal += reportesVentaTotal(venta)

		activadoRaw := reportesFirstNonBlank(venta.ActivadoEn, venta.FechaCreacion)
		corteRaw := reportesFirstNonBlank(venta.PagadoEn, venta.FechaActualizacion, venta.FechaCreacion)
		activadoAt, okInicio := reportesParseDateTime(activadoRaw)
		corteAt, okCorte := reportesParseDateTime(corteRaw)
		if !okInicio || !okCorte || corteAt.Before(activadoAt) {
			row.ventasSinBase++
			continue
		}

		detalle := dbpkg.CalcularDetalleTarifaPorDia(tarifa, activadoAt, corteAt)
		row.ingresoEsperado += reportesRound(detalle.MontoTotal)
		row.diasEquivalentes += reportesRound(detalle.DiasEquivalentes)
		row.minutosProrrateo += detalle.MinutosProrrateoFueraWindow
	}

	rows := make([]*tarifaComparativoAgg, 0, len(aggByStation))
	for _, row := range aggByStation {
		rows = append(rows, row)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		diffI := math.Abs(rows[i].ingresoReal - rows[i].ingresoEsperado)
		diffJ := math.Abs(rows[j].ingresoReal - rows[j].ingresoEsperado)
		if diffI == diffJ {
			return rows[i].estacionID < rows[j].estacionID
		}
		return diffI > diffJ
	})

	var ventasEvaluadas int64
	var totalEsperado float64
	var totalReal float64
	var totalDiasEquivalentes float64
	var totalMinutosProrrateo int64
	var totalVentasSinBase int64

	for _, row := range rows {
		ventasEvaluadas += row.ventasCerradas
		totalEsperado += reportesRound(row.ingresoEsperado)
		totalReal += reportesRound(row.ingresoReal)
		totalDiasEquivalentes += reportesRound(row.diasEquivalentes)
		totalMinutosProrrateo += row.minutosProrrateo
		totalVentasSinBase += row.ventasSinBase

		desviacion := reportesRound(row.ingresoReal - row.ingresoEsperado)
		cumplimiento := 0.0
		if row.ingresoEsperado > 0 {
			cumplimiento = reportesRound((row.ingresoReal * 100.0) / row.ingresoEsperado)
		}

		ds.Rows = append(ds.Rows, map[string]interface{}{
			"estacion_id":                     row.estacionID,
			"estacion_codigo":                 row.estacionCodigo,
			"estacion_nombre":                 row.estacionNombre,
			"ventas_cerradas":                 row.ventasCerradas,
			"tarifa_valor_dia":                reportesRound(row.tarifaValorDia),
			"ingreso_esperado":                reportesRound(row.ingresoEsperado),
			"ingreso_real":                    reportesRound(row.ingresoReal),
			"desviacion_monto":                desviacion,
			"cumplimiento_pct":                cumplimiento,
			"dias_equivalentes_esperados":     reportesRound(row.diasEquivalentes),
			"minutos_prorrateo_fuera_ventana": row.minutosProrrateo,
			"ventas_sin_base_calculo":         row.ventasSinBase,
		})
	}

	ds.RowCount = len(ds.Rows)
	ds.Summary["ventas_evaluadas"] = ventasEvaluadas
	ds.Summary["estaciones_comparadas"] = ds.RowCount
	ds.Summary["ingreso_esperado_total"] = reportesRound(totalEsperado)
	ds.Summary["ingreso_real_total"] = reportesRound(totalReal)
	ds.Summary["desviacion_total"] = reportesRound(totalReal - totalEsperado)
	ds.Summary["dias_equivalentes_total"] = reportesRound(totalDiasEquivalentes)
	ds.Summary["minutos_prorrateo_fuera_ventana_total"] = totalMinutosProrrateo
	ds.Summary["ventas_sin_base_calculo_total"] = totalVentasSinBase

	cumplimientoGlobal := 0.0
	if totalEsperado > 0 {
		cumplimientoGlobal = reportesRound((totalReal * 100.0) / totalEsperado)
	}
	ds.Summary["cumplimiento_global_pct"] = cumplimientoGlobal

	return ds, nil
}

type reporteCadenaCumplimientoDef struct {
	ModuloKey          string
	Modulo             string
	Tabla              string
	EstadoColumna      string
	FechaColumnas      []string
	MontoColumna       string
	EstadosFinalizados []string
	EstadosProceso     []string
}

var reporteCadenaCumplimientoDefs = []reporteCadenaCumplimientoDef{
	{
		ModuloKey:          "crm_leads",
		Modulo:             "CRM - Leads",
		Tabla:              "crm_leads",
		EstadoColumna:      "estado_lead",
		FechaColumnas:      []string{"fecha_creacion", "fecha_actualizacion"},
		MontoColumna:       "valor_potencial",
		EstadosFinalizados: []string{"ganado"},
		EstadosProceso:     []string{"contactado", "calificado", "propuesta", "negociacion", "reactivado"},
	},
	{
		ModuloKey:          "produccion_ordenes",
		Modulo:             "Producción - Órdenes",
		Tabla:              "produccion_ordenes",
		EstadoColumna:      "estado_orden",
		FechaColumnas:      []string{"fecha_programada", "fecha_inicio", "fecha_creacion"},
		MontoColumna:       "costo_real",
		EstadosFinalizados: []string{"entregado", "cerrado", "finalizada", "aplicada"},
		EstadosProceso:     []string{"confirmado", "en_preparacion", "despachado", "planificada", "en_proceso"},
	},
	{
		ModuloKey:          "logistica_envios",
		Modulo:             "Logística - Envíos",
		Tabla:              "logistica_envios",
		EstadoColumna:      "estado_envio",
		FechaColumnas:      []string{"fecha_programada", "fecha_salida", "fecha_creacion"},
		MontoColumna:       "costo_envio",
		EstadosFinalizados: []string{"entregado", "cerrado", "aplicado"},
		EstadosProceso:     []string{"programado", "en_ruta", "despachado", "en_preparacion"},
	},
}

func (b *reportesBuilder) buildOperativoCadenaCumplimientoDataset() (empresaReporteDataset, error) {
	ds := b.newDataset(reporteDatasetOperativoCadena, []string{
		"modulo_key",
		"modulo",
		"tabla",
		"registros_totales",
		"registros_activos",
		"registros_rango",
		"en_proceso",
		"finalizados",
		"cumplimiento_pct",
		"monto_referencia",
		"fecha_columna",
		"nota",
	})

	hasDateFilter := strings.TrimSpace(b.desde) != "" || strings.TrimSpace(b.hasta) != ""

	var registrosTotales int64
	var registrosRango int64
	var finalizadosTotal int64
	var enProcesoTotal int64
	var montoReferenciaTotal float64

	cumplimientoPorModulo := make(map[string]float64)

	for _, def := range reporteCadenaCumplimientoDefs {
		row := map[string]interface{}{
			"modulo_key":        def.ModuloKey,
			"modulo":            def.Modulo,
			"tabla":             def.Tabla,
			"registros_totales": int64(0),
			"registros_activos": int64(0),
			"registros_rango":   int64(0),
			"en_proceso":        int64(0),
			"finalizados":       int64(0),
			"cumplimiento_pct":  float64(0),
			"monto_referencia":  float64(0),
			"fecha_columna":     "",
			"nota":              "ok",
		}

		tablaExiste, err := b.reportesTableExists(def.Tabla)
		if err != nil {
			return empresaReporteDataset{}, err
		}
		if !tablaExiste {
			row["nota"] = "tabla_no_disponible"
			ds.Rows = append(ds.Rows, row)
			continue
		}

		columnas, err := b.reportesTableColumns(def.Tabla)
		if err != nil {
			return empresaReporteDataset{}, err
		}
		if !columnas["empresa_id"] {
			row["nota"] = "sin_columna_empresa_id"
			ds.Rows = append(ds.Rows, row)
			continue
		}

		total, err := b.reportesCountByEmpresa(def.Tabla, "")
		if err != nil {
			return empresaReporteDataset{}, err
		}
		activos := total
		if columnas["estado"] {
			activos, err = b.reportesCountByEmpresa(def.Tabla, "AND LOWER(COALESCE(estado, 'activo')) = 'activo'")
			if err != nil {
				return empresaReporteDataset{}, err
			}
		}

		fechaColumna := ""
		for _, candidate := range def.FechaColumnas {
			c := strings.ToLower(strings.TrimSpace(candidate))
			if c != "" && columnas[c] {
				fechaColumna = c
				break
			}
		}

		rangeClause := ""
		rangeArgs := make([]interface{}, 0)
		rango := total
		if hasDateFilter {
			if fechaColumna == "" {
				rango = 0
				row["nota"] = "sin_columna_fecha"
			} else {
				rangeClause, rangeArgs = reportesBuildDateFilterClause(fechaColumna, b.desde, b.hasta)
				rango, err = b.reportesCountByEmpresa(def.Tabla, rangeClause, rangeArgs...)
				if err != nil {
					return empresaReporteDataset{}, err
				}
			}
		}

		finalizados := int64(0)
		if def.EstadoColumna != "" && columnas[strings.ToLower(strings.TrimSpace(def.EstadoColumna))] {
			finalClause, finalArgs := reportesBuildStateFilterClause(def.EstadoColumna, def.EstadosFinalizados)
			extra := strings.TrimSpace(strings.Join([]string{rangeClause, finalClause}, " "))
			args := append([]interface{}{}, rangeArgs...)
			args = append(args, finalArgs...)
			finalizados, err = b.reportesCountByEmpresa(def.Tabla, extra, args...)
			if err != nil {
				return empresaReporteDataset{}, err
			}
		}

		enProceso := int64(0)
		if def.EstadoColumna != "" && columnas[strings.ToLower(strings.TrimSpace(def.EstadoColumna))] {
			procClause, procArgs := reportesBuildStateFilterClause(def.EstadoColumna, def.EstadosProceso)
			extra := strings.TrimSpace(strings.Join([]string{rangeClause, procClause}, " "))
			args := append([]interface{}{}, rangeArgs...)
			args = append(args, procArgs...)
			enProceso, err = b.reportesCountByEmpresa(def.Tabla, extra, args...)
			if err != nil {
				return empresaReporteDataset{}, err
			}
		}

		montoReferencia := 0.0
		if def.MontoColumna != "" && columnas[strings.ToLower(strings.TrimSpace(def.MontoColumna))] {
			montoReferencia, err = b.reportesSumByEmpresa(def.Tabla, def.MontoColumna, rangeClause, rangeArgs...)
			if err != nil {
				return empresaReporteDataset{}, err
			}
		}

		cumplimiento := 0.0
		if rango > 0 {
			cumplimiento = reportesRound((float64(finalizados) * 100.0) / float64(rango))
		}

		row["registros_totales"] = total
		row["registros_activos"] = activos
		row["registros_rango"] = rango
		row["en_proceso"] = enProceso
		row["finalizados"] = finalizados
		row["cumplimiento_pct"] = cumplimiento
		row["monto_referencia"] = reportesRound(montoReferencia)
		row["fecha_columna"] = fechaColumna

		ds.Rows = append(ds.Rows, row)

		registrosTotales += total
		registrosRango += rango
		finalizadosTotal += finalizados
		enProcesoTotal += enProceso
		montoReferenciaTotal += montoReferencia
		cumplimientoPorModulo[def.ModuloKey] = cumplimiento
	}

	ds.RowCount = len(ds.Rows)
	ds.Summary["modulos_total"] = ds.RowCount
	ds.Summary["registros_totales"] = registrosTotales
	ds.Summary["registros_rango"] = registrosRango
	ds.Summary["finalizados_totales"] = finalizadosTotal
	ds.Summary["en_proceso_totales"] = enProcesoTotal
	ds.Summary["monto_referencia_total"] = reportesRound(montoReferenciaTotal)
	ds.Summary["crm_conversion_pct"] = cumplimientoPorModulo["crm_leads"]
	ds.Summary["produccion_cumplimiento_pct"] = cumplimientoPorModulo["produccion_ordenes"]
	ds.Summary["logistica_cumplimiento_pct"] = cumplimientoPorModulo["logistica_envios"]
	ds.Summary["rango_desde"] = reportesFirstNonBlank(strings.TrimSpace(b.desde), "sin_desde")
	ds.Summary["rango_hasta"] = reportesFirstNonBlank(strings.TrimSpace(b.hasta), "sin_hasta")

	cumplimientoGlobal := 0.0
	if registrosRango > 0 {
		cumplimientoGlobal = reportesRound((float64(finalizadosTotal) * 100.0) / float64(registrosRango))
	}
	ds.Summary["cumplimiento_global_pct"] = cumplimientoGlobal

	return ds, nil
}

func (b *reportesBuilder) buildOperativoTopProductosDataset() (empresaReporteDataset, error) {
	ventas, err := b.getVentasCerradasFiltradas()
	if err != nil {
		return empresaReporteDataset{}, err
	}
	if err := b.ensureItemsForCarritos(ventas); err != nil {
		return empresaReporteDataset{}, err
	}
	type agg struct {
		nombre    string
		cantidad  float64
		total     float64
		ventasSet map[int64]struct{}
	}
	aggregates := make(map[string]*agg)

	for _, venta := range ventas {
		items := b.itemsCache[venta.ID]
		for _, it := range items {
			if !reportesEstadoActivo(it.Estado) {
				continue
			}
			key := strings.TrimSpace(it.Descripcion)
			if it.ReferenciaID > 0 {
				key = "producto_" + strconv.FormatInt(it.ReferenciaID, 10)
			}
			if key == "" {
				key = "item_" + strconv.FormatInt(it.ID, 10)
			}
			current, ok := aggregates[key]
			if !ok {
				current = &agg{
					nombre:    reportesFirstNonBlank(it.Descripcion, it.CodigoItem, key),
					ventasSet: make(map[int64]struct{}),
				}
				aggregates[key] = current
			}
			current.cantidad += it.Cantidad
			current.total += it.TotalLinea
			current.ventasSet[venta.ID] = struct{}{}
		}
	}
	type row struct {
		Producto string
		Cantidad float64
		Total    float64
		Ventas   int
	}
	rows := make([]row, 0, len(aggregates))
	for _, item := range aggregates {
		rows = append(rows, row{
			Producto: item.nombre,
			Cantidad: reportesRound(item.cantidad),
			Total:    reportesRound(item.total),
			Ventas:   len(item.ventasSet),
		})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Total == rows[j].Total {
			return rows[i].Cantidad > rows[j].Cantidad
		}
		return rows[i].Total > rows[j].Total
	})
	if len(rows) > b.maxRows {
		rows = rows[:b.maxRows]
	}

	ds := b.newDataset(reporteDatasetOperativoTopProductos, []string{
		"producto",
		"cantidad_vendida",
		"total_vendido",
		"ventas_relacionadas",
	})
	totalVendido := 0.0
	for _, item := range rows {
		totalVendido += item.Total
		ds.Rows = append(ds.Rows, map[string]interface{}{
			"producto":            item.Producto,
			"cantidad_vendida":    item.Cantidad,
			"total_vendido":       item.Total,
			"ventas_relacionadas": item.Ventas,
		})
	}
	ds.RowCount = len(ds.Rows)
	ds.Summary["productos"] = ds.RowCount
	ds.Summary["ingresos"] = reportesRound(totalVendido)
	return ds, nil
}

func (b *reportesBuilder) buildOperativoTopClientesDataset() (empresaReporteDataset, error) {
	ventas, err := b.getVentasCerradasFiltradas()
	if err != nil {
		return empresaReporteDataset{}, err
	}
	type agg struct {
		cliente string
		ventas  int64
		total   float64
	}
	clients := make(map[string]*agg)
	for _, venta := range ventas {
		cliente := reportesFirstNonBlank(venta.ClienteNombre, "Sin cliente")
		row, ok := clients[cliente]
		if !ok {
			row = &agg{cliente: cliente}
			clients[cliente] = row
		}
		row.ventas++
		row.total += reportesVentaTotal(venta)
	}
	rows := make([]agg, 0, len(clients))
	for _, item := range clients {
		rows = append(rows, agg{cliente: item.cliente, ventas: item.ventas, total: reportesRound(item.total)})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].total == rows[j].total {
			return rows[i].ventas > rows[j].ventas
		}
		return rows[i].total > rows[j].total
	})
	if len(rows) > b.maxRows {
		rows = rows[:b.maxRows]
	}
	ds := b.newDataset(reporteDatasetOperativoTopClientes, []string{
		"cliente",
		"ventas",
		"total_comprado",
	})
	totalComprado := 0.0
	for _, item := range rows {
		totalComprado += item.total
		ds.Rows = append(ds.Rows, map[string]interface{}{
			"cliente":        item.cliente,
			"ventas":         item.ventas,
			"total_comprado": item.total,
		})
	}
	ds.RowCount = len(ds.Rows)
	ds.Summary["clientes"] = ds.RowCount
	ds.Summary["total_comprado"] = reportesRound(totalComprado)
	return ds, nil
}

func reportesClienteSegmentoPrioridad(segmento string) int {
	switch strings.ToLower(strings.TrimSpace(segmento)) {
	case "estrategico":
		return 0
	case "frecuente":
		return 1
	case "activo":
		return 2
	case "nuevo":
		return 3
	case "inactivo":
		return 4
	default:
		return 99
	}
}

func reportesClienteAccionComercial(segmento string) string {
	switch strings.ToLower(strings.TrimSpace(segmento)) {
	case "estrategico":
		return "fidelizacion_vip"
	case "frecuente":
		return "upsell_crosssell"
	case "activo":
		return "reactivacion_temprana"
	case "inactivo":
		return "recuperacion"
	default:
		return "onboarding"
	}
}

func (b *reportesBuilder) buildOperativoClientesSegmentacionComercialDataset() (empresaReporteDataset, error) {
	clientes, err := dbpkg.GetClientesByEmpresa(b.db, b.empresaID, b.includeInactive, "")
	if err != nil {
		return empresaReporteDataset{}, err
	}

	type row struct {
		ClienteID       int64
		Documento       string
		Nombre          string
		NombreComercial string
		Email           string
		Telefono        string
		Segmento        string
		Accion          string
		Compras         int64
		MontoCompras    float64
		TicketPromedio  float64
		UltimaCompra    string
		DiasSinCompra   int
		Estado          string
	}

	rows := make([]row, 0, len(clientes))
	for _, cliente := range clientes {
		perfil, err := dbpkg.GetClientePerfilComercialByEmpresa(b.db, b.empresaID, cliente.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return empresaReporteDataset{}, err
		}

		segmento := strings.ToLower(strings.TrimSpace(perfil.Segmento))
		if segmento == "" {
			segmento = "nuevo"
		}

		documento := strings.TrimSpace(strings.TrimSpace(perfil.Cliente.TipoDocumento) + " " + strings.TrimSpace(perfil.Cliente.NumeroDocumento))
		rows = append(rows, row{
			ClienteID:       perfil.Cliente.ID,
			Documento:       documento,
			Nombre:          strings.TrimSpace(perfil.Cliente.NombreRazonSocial),
			NombreComercial: strings.TrimSpace(perfil.Cliente.NombreComercial),
			Email:           strings.TrimSpace(perfil.Cliente.Email),
			Telefono:        strings.TrimSpace(perfil.Cliente.Telefono),
			Segmento:        segmento,
			Accion:          reportesClienteAccionComercial(segmento),
			Compras:         perfil.NumeroCompras,
			MontoCompras:    reportesRound(perfil.MontoCompras),
			TicketPromedio:  reportesRound(perfil.TicketPromedio),
			UltimaCompra:    strings.TrimSpace(perfil.UltimaCompra),
			DiasSinCompra:   perfil.DiasSinCompra,
			Estado:          reportesFirstNonBlank(strings.TrimSpace(perfil.Cliente.Estado), "activo"),
		})
	}

	sort.SliceStable(rows, func(i, j int) bool {
		pi := reportesClienteSegmentoPrioridad(rows[i].Segmento)
		pj := reportesClienteSegmentoPrioridad(rows[j].Segmento)
		if pi != pj {
			return pi < pj
		}
		if rows[i].MontoCompras != rows[j].MontoCompras {
			return rows[i].MontoCompras > rows[j].MontoCompras
		}
		if rows[i].Compras != rows[j].Compras {
			return rows[i].Compras > rows[j].Compras
		}
		return rows[i].ClienteID < rows[j].ClienteID
	})

	if len(rows) > b.maxRows {
		rows = rows[:b.maxRows]
	}

	ds := b.newDataset(reporteDatasetOperativoClientesSegmentos, []string{
		"cliente_id",
		"documento",
		"nombre_razon_social",
		"nombre_comercial",
		"email",
		"telefono",
		"segmento",
		"accion_comercial_sugerida",
		"numero_compras",
		"monto_compras",
		"ticket_promedio",
		"ultima_compra",
		"dias_sin_compra",
		"estado",
	})

	segmentos := map[string]int64{
		"estrategico": 0,
		"frecuente":   0,
		"activo":      0,
		"nuevo":       0,
		"inactivo":    0,
	}
	totalCompras := int64(0)
	totalMonto := 0.0

	for _, item := range rows {
		totalCompras += item.Compras
		totalMonto += item.MontoCompras
		if _, ok := segmentos[item.Segmento]; ok {
			segmentos[item.Segmento]++
		}
		ds.Rows = append(ds.Rows, map[string]interface{}{
			"cliente_id":                item.ClienteID,
			"documento":                 item.Documento,
			"nombre_razon_social":       item.Nombre,
			"nombre_comercial":          item.NombreComercial,
			"email":                     item.Email,
			"telefono":                  item.Telefono,
			"segmento":                  item.Segmento,
			"accion_comercial_sugerida": item.Accion,
			"numero_compras":            item.Compras,
			"monto_compras":             item.MontoCompras,
			"ticket_promedio":           item.TicketPromedio,
			"ultima_compra":             item.UltimaCompra,
			"dias_sin_compra":           item.DiasSinCompra,
			"estado":                    item.Estado,
		})
	}

	ds.RowCount = len(ds.Rows)
	ds.Summary["clientes_considerados"] = len(clientes)
	ds.Summary["clientes_exportados"] = ds.RowCount
	ds.Summary["compras_totales"] = totalCompras
	ds.Summary["monto_compras_total"] = reportesRound(totalMonto)
	ds.Summary["include_inactive"] = b.includeInactive
	ds.Summary["segmento_estrategico"] = segmentos["estrategico"]
	ds.Summary["segmento_frecuente"] = segmentos["frecuente"]
	ds.Summary["segmento_activo"] = segmentos["activo"]
	ds.Summary["segmento_nuevo"] = segmentos["nuevo"]
	ds.Summary["segmento_inactivo"] = segmentos["inactivo"]

	return ds, nil
}

func (b *reportesBuilder) buildOperativoInventarioBodegaDataset() (empresaReporteDataset, error) {
	limit := b.maxRows * 5
	if limit < 500 {
		limit = 500
	}
	existencias, err := dbpkg.GetExistenciasByEmpresa(b.db, b.empresaID, 0, 0, limit, 0)
	if err != nil {
		return empresaReporteDataset{}, err
	}
	productos, err := dbpkg.GetProductosByEmpresa(b.db, b.empresaID, "", "", 0, 0, limit, 0)
	if err != nil {
		return empresaReporteDataset{}, err
	}
	resumen, err := dbpkg.GetInventarioResumenByEmpresa(b.db, b.empresaID, b.desde, b.hasta)
	if err != nil {
		return empresaReporteDataset{}, err
	}
	proyecciones, err := dbpkg.GetInventarioProyeccionQuiebreByEmpresa(b.db, b.empresaID, 0, 30, limit, 0)
	if err != nil {
		return empresaReporteDataset{}, err
	}

	type proyeccionInventario struct {
		Estado             string
		SalidaPromedio     float64
		DiasCobertura      float64
		SugeridoReposicion float64
	}
	proyeccionByKey := make(map[string]proyeccionInventario)
	for _, proy := range proyecciones {
		key := reportesInventarioKey(proy.ProductoID, proy.BodegaID)
		proyeccionByKey[key] = proyeccionInventario{
			Estado:             strings.TrimSpace(proy.EstadoProyeccion),
			SalidaPromedio:     proy.SalidaPromedioDiaria,
			DiasCobertura:      proy.DiasCobertura,
			SugeridoReposicion: proy.SugeridoReposicion,
		}
	}

	productoByID := make(map[int64]dbpkg.Producto)
	for _, p := range productos {
		productoByID[p.ID] = p
	}

	type row struct {
		Producto             string
		Bodega               string
		Cantidad             float64
		Minimo               float64
		Maximo               float64
		EstadoStock          string
		EstadoProyeccion     string
		SalidaPromedio       float64
		DiasCobertura        float64
		IndiceRotacion30d    float64
		SugeridoReposicion   float64
		ValorizacionCosto    float64
		ValorizacionVenta    float64
		PrioridadOrdenRiesgo int
	}
	rows := make([]row, 0, len(existencias))
	for _, ex := range existencias {
		prod := productoByID[ex.ProductoID]
		minimo := prod.StockMinimo
		maximo := prod.StockMaximo
		estadoStock, prioridadStock := reportesEstadoStock(ex.Cantidad, minimo, maximo)

		proy := proyeccionByKey[reportesInventarioKey(ex.ProductoID, ex.BodegaID)]
		estadoProyeccion := reportesFirstNonBlank(strings.TrimSpace(proy.Estado), "sin_datos")
		salidaPromedio := reportesRound(proy.SalidaPromedio)
		diasCobertura := reportesRound(proy.DiasCobertura)
		sugeridoReposicion := reportesRound(proy.SugeridoReposicion)

		indiceRotacion30d := 0.0
		if ex.Cantidad > 0 && proy.SalidaPromedio > 0 {
			indiceRotacion30d = reportesRound((proy.SalidaPromedio * 30.0) / ex.Cantidad)
		}

		valorizacionCosto := reportesRound(ex.Cantidad * prod.Costo)
		valorizacionVenta := reportesRound(ex.Cantidad * prod.Precio)
		rows = append(rows, row{
			Producto:             reportesFirstNonBlank(ex.ProductoNombre, prod.Nombre),
			Bodega:               reportesFirstNonBlank(ex.BodegaNombre, "Bodega #"+strconv.FormatInt(ex.BodegaID, 10)),
			Cantidad:             reportesRound(ex.Cantidad),
			Minimo:               reportesRound(minimo),
			Maximo:               reportesRound(maximo),
			EstadoStock:          estadoStock,
			EstadoProyeccion:     estadoProyeccion,
			SalidaPromedio:       salidaPromedio,
			DiasCobertura:        diasCobertura,
			IndiceRotacion30d:    indiceRotacion30d,
			SugeridoReposicion:   sugeridoReposicion,
			ValorizacionCosto:    valorizacionCosto,
			ValorizacionVenta:    valorizacionVenta,
			PrioridadOrdenRiesgo: reportesInventarioRiesgoPrioridad(estadoProyeccion, estadoStock, prioridadStock),
		})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].PrioridadOrdenRiesgo != rows[j].PrioridadOrdenRiesgo {
			return rows[i].PrioridadOrdenRiesgo < rows[j].PrioridadOrdenRiesgo
		}
		if rows[i].Producto != rows[j].Producto {
			return rows[i].Producto < rows[j].Producto
		}
		return rows[i].Bodega < rows[j].Bodega
	})
	if len(rows) > b.maxRows {
		rows = rows[:b.maxRows]
	}
	ds := b.newDataset(reporteDatasetOperativoInventario, []string{
		"producto",
		"bodega",
		"existencia",
		"stock_minimo",
		"stock_maximo",
		"estado_stock",
		"estado_proyeccion",
		"salida_promedio_diaria",
		"dias_cobertura",
		"indice_rotacion_30d",
		"sugerido_reposicion",
		"valorizacion_costo",
		"valorizacion_venta",
	})
	estadoCount := map[string]int{}
	estadoProyeccionCount := map[string]int{}
	valorizacionCostoTotal := 0.0
	valorizacionVentaTotal := 0.0
	salidaPromedioTotal := 0.0
	indiceRotacionTotal := 0.0
	coberturaTotal := 0.0
	coberturaConDato := 0
	for _, item := range rows {
		estadoCount[item.EstadoStock]++
		estadoProyeccionCount[item.EstadoProyeccion]++
		valorizacionCostoTotal += item.ValorizacionCosto
		valorizacionVentaTotal += item.ValorizacionVenta
		salidaPromedioTotal += item.SalidaPromedio
		indiceRotacionTotal += item.IndiceRotacion30d
		if item.DiasCobertura > 0 {
			coberturaTotal += item.DiasCobertura
			coberturaConDato++
		}
		ds.Rows = append(ds.Rows, map[string]interface{}{
			"producto":               item.Producto,
			"bodega":                 item.Bodega,
			"existencia":             item.Cantidad,
			"stock_minimo":           item.Minimo,
			"stock_maximo":           item.Maximo,
			"estado_stock":           item.EstadoStock,
			"estado_proyeccion":      item.EstadoProyeccion,
			"salida_promedio_diaria": item.SalidaPromedio,
			"dias_cobertura":         item.DiasCobertura,
			"indice_rotacion_30d":    item.IndiceRotacion30d,
			"sugerido_reposicion":    item.SugeridoReposicion,
			"valorizacion_costo":     item.ValorizacionCosto,
			"valorizacion_venta":     item.ValorizacionVenta,
		})
	}
	ds.RowCount = len(ds.Rows)
	ds.Summary["sin_stock"] = estadoCount["sin_stock"]
	ds.Summary["bajo_minimo"] = estadoCount["bajo_minimo"]
	ds.Summary["ok"] = estadoCount["ok"]
	ds.Summary["sobre_stock"] = estadoCount["sobre_stock"]
	ds.Summary["quiebre_inminente"] = estadoProyeccionCount["quiebre_inminente"]
	ds.Summary["bajo_minimo_proyeccion"] = estadoProyeccionCount["bajo_minimo"]
	ds.Summary["riesgo_alto"] = estadoProyeccionCount["riesgo_alto"]
	ds.Summary["riesgo_medio"] = estadoProyeccionCount["riesgo_medio"]
	ds.Summary["sin_consumo_reciente"] = estadoProyeccionCount["sin_consumo_reciente"]
	ds.Summary["estable"] = estadoProyeccionCount["estable"]
	ds.Summary["valorizacion_costo_total"] = reportesRound(valorizacionCostoTotal)
	ds.Summary["valorizacion_venta_total"] = reportesRound(valorizacionVentaTotal)
	ds.Summary["salida_promedio_diaria_total"] = reportesRound(salidaPromedioTotal)
	if ds.RowCount > 0 {
		ds.Summary["rotacion_promedio_30d"] = reportesRound(indiceRotacionTotal / float64(ds.RowCount))
	} else {
		ds.Summary["rotacion_promedio_30d"] = 0.0
	}
	if coberturaConDato > 0 {
		ds.Summary["cobertura_promedio_dias"] = reportesRound(coberturaTotal / float64(coberturaConDato))
	} else {
		ds.Summary["cobertura_promedio_dias"] = 0.0
	}
	ds.Summary["productos_con_existencia"] = resumen.ProductosConExistencia
	ds.Summary["bodegas_con_stock"] = resumen.BodegasConStock
	ds.Summary["alertas_total"] = resumen.AlertasTotal
	ds.Summary["alertas_sin_stock"] = resumen.AlertasSinStock
	ds.Summary["alertas_bajo_minimo"] = resumen.AlertasBajoMinimo
	ds.Summary["deficit_total"] = reportesRound(resumen.DeficitTotal)
	ds.Summary["movimientos_total"] = resumen.MovimientosTotal
	ds.Summary["movimientos_entrada"] = resumen.MovimientosEntrada
	ds.Summary["movimientos_salida"] = resumen.MovimientosSalida
	ds.Summary["movimientos_traslado"] = resumen.MovimientosTraslado
	ds.Summary["movimientos_ajuste"] = resumen.MovimientosAjuste
	ds.Summary["ultimo_movimiento"] = reportesFirstNonBlank(strings.TrimSpace(resumen.UltimoMovimiento), "sin_movimientos")
	ds.Summary["periodo_desde"] = reportesFirstNonBlank(strings.TrimSpace(resumen.PeriodoDesde), strings.TrimSpace(b.desde), "sin_desde")
	ds.Summary["periodo_hasta"] = reportesFirstNonBlank(strings.TrimSpace(resumen.PeriodoHasta), strings.TrimSpace(b.hasta), "sin_hasta")
	return ds, nil
}

func (b *reportesBuilder) buildOperativoComprasMovimientosDataset() (empresaReporteDataset, error) {
	type comprasProveedorAgg struct {
		ProveedorID          int64
		Proveedor            string
		Documentos           int
		OrdenesEmitidas      int
		Recepciones          int
		Contabilizaciones    int
		MontoOrdenado        float64
		MontoRecepcionado    float64
		MontoContabilizado   float64
		UltimaFechaDocumento string
		Moneda               string
	}

	normalizeEstadoDocumento := func(raw string) string {
		value := strings.ToLower(strings.TrimSpace(raw))
		value = strings.ReplaceAll(value, "-", "_")
		value = strings.ReplaceAll(value, " ", "_")
		return value
	}

	isEstadoOrdenEmitida := func(estado string) bool {
		switch normalizeEstadoDocumento(estado) {
		case "emitida", "recepcionada", "recepcion_parcial", "contabilizada":
			return true
		default:
			return false
		}
	}

	isEstadoRecepcion := func(estado string) bool {
		switch normalizeEstadoDocumento(estado) {
		case "recepcionada", "recepcion_parcial", "contabilizada":
			return true
		default:
			return false
		}
	}

	isEstadoContabilizada := func(estado string) bool {
		return normalizeEstadoDocumento(estado) == "contabilizada"
	}

	proveedores, err := dbpkg.GetProveedoresByEmpresa(b.db, b.empresaID, true)
	if err != nil {
		return empresaReporteDataset{}, err
	}
	proveedorNombreByID := make(map[int64]string, len(proveedores))
	for _, prov := range proveedores {
		nombre := strings.TrimSpace(prov.Nombre)
		if nombre == "" {
			nombre = "Proveedor #" + strconv.FormatInt(prov.ID, 10)
		}
		proveedorNombreByID[prov.ID] = nombre
	}

	const pageSize = 500
	maxDocs := b.maxRows * 25
	if maxDocs < pageSize {
		maxDocs = pageSize
	}
	if maxDocs > 5000 {
		maxDocs = 5000
	}

	documentos := make([]dbpkg.EmpresaDocumentoCompra, 0, maxDocs)
	for offset := 0; offset < maxDocs; {
		remaining := maxDocs - offset
		limit := pageSize
		if remaining < limit {
			limit = remaining
		}

		batch, listErr := dbpkg.ListEmpresaDocumentosCompraByEmpresa(
			b.db,
			b.empresaID,
			"orden_compra",
			0,
			"",
			b.includeInactive,
			"",
			limit,
			offset,
		)
		if listErr != nil {
			return empresaReporteDataset{}, listErr
		}

		documentos = append(documentos, batch...)
		if len(batch) < limit {
			break
		}
		offset += len(batch)
	}

	aggByProveedor := make(map[int64]*comprasProveedorAgg)
	globalMoneda := ""
	globalMonedaMixta := false

	totalDocumentos := 0
	totalOrdenesEmitidas := 0
	totalRecepciones := 0
	totalContabilizaciones := 0
	totalMontoOrdenado := 0.0
	totalMontoRecepcionado := 0.0
	totalMontoContabilizado := 0.0

	for _, doc := range documentos {
		fechaDocumento := reportesFirstNonBlank(doc.FechaDocumento, doc.FechaActualizacion, doc.FechaCreacion)
		if !reportesDateWithinRange(fechaDocumento, b.desde, b.hasta) {
			continue
		}

		proveedorID := doc.ProveedorID
		agg, ok := aggByProveedor[proveedorID]
		if !ok {
			nombreProveedor := strings.TrimSpace(proveedorNombreByID[proveedorID])
			if nombreProveedor == "" {
				if proveedorID > 0 {
					nombreProveedor = "Proveedor #" + strconv.FormatInt(proveedorID, 10)
				} else {
					nombreProveedor = "Sin proveedor"
				}
			}
			agg = &comprasProveedorAgg{
				ProveedorID: proveedorID,
				Proveedor:   nombreProveedor,
			}
			aggByProveedor[proveedorID] = agg
		}

		agg.Documentos++
		totalDocumentos++

		montoDocumento := reportesRound(doc.MontoTotal)
		if montoDocumento < 0 {
			montoDocumento = 0
		}

		estadoDoc := normalizeEstadoDocumento(doc.EstadoDocumento)
		if isEstadoOrdenEmitida(estadoDoc) {
			agg.OrdenesEmitidas++
			agg.MontoOrdenado += montoDocumento
			totalOrdenesEmitidas++
			totalMontoOrdenado += montoDocumento
		}
		if isEstadoRecepcion(estadoDoc) {
			agg.Recepciones++
			agg.MontoRecepcionado += montoDocumento
			totalRecepciones++
			totalMontoRecepcionado += montoDocumento
		}
		if isEstadoContabilizada(estadoDoc) {
			agg.Contabilizaciones++
			agg.MontoContabilizado += montoDocumento
			totalContabilizaciones++
			totalMontoContabilizado += montoDocumento
		}

		if reportesDateUnix(fechaDocumento) > reportesDateUnix(agg.UltimaFechaDocumento) {
			agg.UltimaFechaDocumento = fechaDocumento
		}

		monedaDocumento := strings.ToUpper(strings.TrimSpace(doc.Moneda))
		if monedaDocumento == "" {
			monedaDocumento = "COP"
		}
		if agg.Moneda == "" {
			agg.Moneda = monedaDocumento
		} else if agg.Moneda != monedaDocumento {
			agg.Moneda = "MIXTA"
		}
		if globalMoneda == "" {
			globalMoneda = monedaDocumento
		} else if globalMoneda != monedaDocumento {
			globalMonedaMixta = true
		}
	}

	aggregated := make([]*comprasProveedorAgg, 0, len(aggByProveedor))
	for _, item := range aggByProveedor {
		aggregated = append(aggregated, item)
	}
	sort.SliceStable(aggregated, func(i, j int) bool {
		if aggregated[i].MontoOrdenado == aggregated[j].MontoOrdenado {
			if aggregated[i].Recepciones == aggregated[j].Recepciones {
				return strings.ToLower(strings.TrimSpace(aggregated[i].Proveedor)) < strings.ToLower(strings.TrimSpace(aggregated[j].Proveedor))
			}
			return aggregated[i].Recepciones > aggregated[j].Recepciones
		}
		return aggregated[i].MontoOrdenado > aggregated[j].MontoOrdenado
	})

	proveedoresConOrden := 0
	proveedoresPendientesRecepcion := 0
	for _, item := range aggregated {
		if item.OrdenesEmitidas > 0 {
			proveedoresConOrden++
		}
		if item.OrdenesEmitidas > item.Recepciones {
			proveedoresPendientesRecepcion++
		}
	}

	ds := b.newDataset(reporteDatasetOperativoCompras, []string{
		"proveedor_id",
		"proveedor",
		"documentos",
		"ordenes_emitidas",
		"recepciones",
		"contabilizaciones",
		"monto_ordenado",
		"monto_recepcionado",
		"monto_contabilizado",
		"brecha_monto",
		"cumplimiento_recepcion_pct",
		"cumplimiento_monto_pct",
		"ticket_promedio_orden",
		"ultima_fecha_documento",
		"moneda",
	})

	shown := aggregated
	if len(shown) > b.maxRows {
		shown = shown[:b.maxRows]
	}
	for _, item := range shown {
		cumplimientoRecepcionPct := 0.0
		cumplimientoMontoPct := 0.0
		ticketPromedioOrden := 0.0
		if item.OrdenesEmitidas > 0 {
			cumplimientoRecepcionPct = reportesRound((float64(item.Recepciones) * 100.0) / float64(item.OrdenesEmitidas))
			ticketPromedioOrden = reportesRound(item.MontoOrdenado / float64(item.OrdenesEmitidas))
		}
		if item.MontoOrdenado > 0 {
			cumplimientoMontoPct = reportesRound((item.MontoRecepcionado * 100.0) / item.MontoOrdenado)
		}

		brechaMonto := reportesRound(item.MontoOrdenado - item.MontoRecepcionado)
		ds.Rows = append(ds.Rows, map[string]interface{}{
			"proveedor_id":               item.ProveedorID,
			"proveedor":                  item.Proveedor,
			"documentos":                 item.Documentos,
			"ordenes_emitidas":           item.OrdenesEmitidas,
			"recepciones":                item.Recepciones,
			"contabilizaciones":          item.Contabilizaciones,
			"monto_ordenado":             reportesRound(item.MontoOrdenado),
			"monto_recepcionado":         reportesRound(item.MontoRecepcionado),
			"monto_contabilizado":        reportesRound(item.MontoContabilizado),
			"brecha_monto":               brechaMonto,
			"cumplimiento_recepcion_pct": cumplimientoRecepcionPct,
			"cumplimiento_monto_pct":     cumplimientoMontoPct,
			"ticket_promedio_orden":      ticketPromedioOrden,
			"ultima_fecha_documento":     item.UltimaFechaDocumento,
			"moneda":                     reportesFirstNonBlank(strings.TrimSpace(item.Moneda), "COP"),
		})
	}

	totalBrechaMonto := reportesRound(totalMontoOrdenado - totalMontoRecepcionado)
	cumplimientoRecepcionGlobal := 0.0
	if totalOrdenesEmitidas > 0 {
		cumplimientoRecepcionGlobal = reportesRound((float64(totalRecepciones) * 100.0) / float64(totalOrdenesEmitidas))
	}
	cumplimientoMontoGlobal := 0.0
	if totalMontoOrdenado > 0 {
		cumplimientoMontoGlobal = reportesRound((totalMontoRecepcionado * 100.0) / totalMontoOrdenado)
	}
	monedaResumen := strings.TrimSpace(globalMoneda)
	if monedaResumen == "" {
		monedaResumen = "COP"
	}
	if globalMonedaMixta {
		monedaResumen = "MIXTA"
	}

	ds.RowCount = len(ds.Rows)
	ds.Summary["movimientos"] = totalDocumentos
	ds.Summary["documentos"] = totalDocumentos
	ds.Summary["proveedores_total"] = len(aggregated)
	ds.Summary["proveedores_listados"] = ds.RowCount
	ds.Summary["proveedores_con_orden"] = proveedoresConOrden
	ds.Summary["proveedores_pendientes_recepcion"] = proveedoresPendientesRecepcion
	ds.Summary["ordenes_emitidas"] = totalOrdenesEmitidas
	ds.Summary["recepciones"] = totalRecepciones
	ds.Summary["contabilizaciones"] = totalContabilizaciones
	ds.Summary["documentos_pendientes_recepcion"] = totalOrdenesEmitidas - totalRecepciones
	ds.Summary["monto_ordenado"] = reportesRound(totalMontoOrdenado)
	ds.Summary["monto_recepcionado"] = reportesRound(totalMontoRecepcionado)
	ds.Summary["monto_contabilizado"] = reportesRound(totalMontoContabilizado)
	ds.Summary["brecha_monto"] = totalBrechaMonto
	ds.Summary["costo_total"] = reportesRound(totalMontoOrdenado)
	ds.Summary["cantidad_total"] = 0.0
	ds.Summary["cumplimiento_recepcion_pct"] = cumplimientoRecepcionGlobal
	ds.Summary["cumplimiento_monto_pct"] = cumplimientoMontoGlobal
	ds.Summary["moneda"] = monedaResumen
	ds.Summary["periodo_desde"] = reportesFirstNonBlank(strings.TrimSpace(b.desde), "sin_desde")
	ds.Summary["periodo_hasta"] = reportesFirstNonBlank(strings.TrimSpace(b.hasta), "sin_hasta")
	return ds, nil
}

func (b *reportesBuilder) buildOperativoPropinasAcumuladoDataset() (empresaReporteDataset, error) {
	limit := b.maxRows * 20
	if limit < 300 {
		limit = 300
	}
	if limit > 5000 {
		limit = 5000
	}

	reporte, err := dbpkg.GetEmpresaPropinasReporte(b.db, b.empresaID, dbpkg.EmpresaPropinaMovimientoFilter{
		Desde:           b.desde,
		Hasta:           b.hasta,
		Usuario:         b.usuario,
		IncludeInactive: b.includeInactive,
		Limit:           limit,
	})
	if err != nil {
		return empresaReporteDataset{}, err
	}

	ds := b.newDataset(reporteDatasetOperativoPropinas, []string{
		"usuario_clave",
		"usuario",
		"es_usuario_activo",
		"movimientos",
		"base_cobro",
		"propina_por_usuario",
		"propina_universal",
		"propina_total",
		"participacion_pct",
	})

	if reporte == nil {
		ds.Summary["movimientos"] = 0
		ds.Summary["total_propinas"] = 0.0
		ds.Summary["periodo_desde"] = reportesFirstNonBlank(strings.TrimSpace(b.desde), "sin_desde")
		ds.Summary["periodo_hasta"] = reportesFirstNonBlank(strings.TrimSpace(b.hasta), "sin_hasta")
		ds.Summary["moneda"] = "COP"
		return ds, nil
	}

	type propinaStats struct {
		Movimientos int64
		BaseCobro   float64
		Propina     float64
	}

	statsByUsuario := make(map[string]*propinaStats)
	monedaResumen := ""
	monedaMixta := false
	for _, mov := range reporte.Movimientos {
		usuarioKey := strings.ToLower(strings.TrimSpace(reportesFirstNonBlank(mov.UsuarioAsignado, mov.UsuarioOrigen, mov.UsuarioCreador, "sistema")))
		if usuarioKey == "" {
			usuarioKey = "sistema"
		}
		stats := statsByUsuario[usuarioKey]
		if stats == nil {
			stats = &propinaStats{}
			statsByUsuario[usuarioKey] = stats
		}
		stats.Movimientos++
		stats.BaseCobro = reportesRound(stats.BaseCobro + mov.BaseCobro)
		stats.Propina = reportesRound(stats.Propina + mov.MontoPropina)

		moneda := strings.ToUpper(strings.TrimSpace(mov.Moneda))
		if moneda == "" {
			moneda = "COP"
		}
		if monedaResumen == "" {
			monedaResumen = moneda
		} else if monedaResumen != moneda {
			monedaMixta = true
		}
	}

	usuarios := append([]dbpkg.EmpresaPropinaUsuarioResumen{}, reporte.Usuarios...)
	if len(usuarios) == 0 && len(statsByUsuario) > 0 {
		type fallbackUsuario struct {
			Clave string
			Total float64
		}
		fallback := make([]fallbackUsuario, 0, len(statsByUsuario))
		for key, stats := range statsByUsuario {
			fallback = append(fallback, fallbackUsuario{Clave: key, Total: reportesRound(stats.Propina)})
		}
		sort.SliceStable(fallback, func(i, j int) bool {
			if fallback[i].Total == fallback[j].Total {
				return fallback[i].Clave < fallback[j].Clave
			}
			return fallback[i].Total > fallback[j].Total
		})
		for _, item := range fallback {
			usuarios = append(usuarios, dbpkg.EmpresaPropinaUsuarioResumen{
				UsuarioClave:      item.Clave,
				UsuarioEtiqueta:   item.Clave,
				EsUsuarioActivo:   false,
				PropinaPorUsuario: item.Total,
				PropinaUniversal:  0,
				PropinaTotal:      item.Total,
			})
		}
	}

	shown := usuarios
	if len(shown) > b.maxRows {
		shown = shown[:b.maxRows]
	}

	totalPropinas := reportesRound(reporte.Resumen.TotalPropinas)
	for _, usuario := range shown {
		usuarioClave := strings.ToLower(strings.TrimSpace(reportesFirstNonBlank(usuario.UsuarioClave, usuario.UsuarioEtiqueta, "sistema")))
		if usuarioClave == "" {
			usuarioClave = "sistema"
		}

		movimientos := int64(0)
		baseCobro := 0.0
		if stats, ok := statsByUsuario[usuarioClave]; ok && stats != nil {
			movimientos = stats.Movimientos
			baseCobro = reportesRound(stats.BaseCobro)
		}

		propinaPorUsuario := reportesRound(usuario.PropinaPorUsuario)
		propinaUniversal := reportesRound(usuario.PropinaUniversal)
		propinaTotal := reportesRound(usuario.PropinaTotal)
		if propinaTotal == 0 {
			propinaTotal = reportesRound(propinaPorUsuario + propinaUniversal)
		}
		if propinaTotal == 0 {
			if stats, ok := statsByUsuario[usuarioClave]; ok && stats != nil {
				propinaTotal = reportesRound(stats.Propina)
			}
		}

		participacion := 0.0
		if totalPropinas > 0 {
			participacion = reportesRound((propinaTotal * 100.0) / totalPropinas)
		}

		ds.Rows = append(ds.Rows, map[string]interface{}{
			"usuario_clave":       usuarioClave,
			"usuario":             reportesFirstNonBlank(strings.TrimSpace(usuario.UsuarioEtiqueta), usuarioClave),
			"es_usuario_activo":   usuario.EsUsuarioActivo,
			"movimientos":         movimientos,
			"base_cobro":          baseCobro,
			"propina_por_usuario": propinaPorUsuario,
			"propina_universal":   propinaUniversal,
			"propina_total":       propinaTotal,
			"participacion_pct":   participacion,
		})
	}

	if monedaResumen == "" {
		monedaResumen = "COP"
	}
	if monedaMixta {
		monedaResumen = "MIXTA"
	}

	cfg := reporte.Configuracion
	if cfg == nil {
		cfg = &dbpkg.EmpresaPropinasConfiguracion{
			EmpresaID:        b.empresaID,
			ModoDistribucion: dbpkg.EmpresaPropinaModoPorUsuario,
			Estado:           "activo",
		}
	}

	ds.RowCount = len(ds.Rows)
	ds.Summary["habilitar_propina"] = cfg.HabilitarPropina
	ds.Summary["porcentaje_propina_config"] = reportesRound(cfg.PorcentajePropina)
	ds.Summary["modo_distribucion"] = reportesFirstNonBlank(strings.TrimSpace(cfg.ModoDistribucion), dbpkg.EmpresaPropinaModoPorUsuario)
	ds.Summary["aplicar_automaticamente"] = cfg.AplicarAutomaticamente
	ds.Summary["usuarios_totales"] = len(usuarios)
	ds.Summary["usuarios_listados"] = ds.RowCount
	ds.Summary["usuarios_activos"] = reporte.Resumen.UsuariosActivos
	ds.Summary["movimientos"] = reporte.Resumen.CantidadMovimientos
	ds.Summary["movimientos_total"] = reporte.Resumen.CantidadMovimientos
	ds.Summary["total_base_cobro"] = reportesRound(reporte.Resumen.TotalBaseCobro)
	ds.Summary["total_propinas"] = totalPropinas
	ds.Summary["total_propinas_por_usuario"] = reportesRound(reporte.Resumen.TotalPropinasPorUsuario)
	ds.Summary["total_propinas_universal"] = reportesRound(reporte.Resumen.TotalPropinasUniversal)
	ds.Summary["cuota_universal_por_usuario"] = reportesRound(reporte.Resumen.CuotaUniversalPorUsuario)
	ds.Summary["filtro_usuario"] = reportesFirstNonBlank(strings.TrimSpace(b.usuario), "todos")
	ds.Summary["periodo_desde"] = reportesFirstNonBlank(strings.TrimSpace(reporte.Desde), strings.TrimSpace(b.desde), "sin_desde")
	ds.Summary["periodo_hasta"] = reportesFirstNonBlank(strings.TrimSpace(reporte.Hasta), strings.TrimSpace(b.hasta), "sin_hasta")
	ds.Summary["moneda"] = monedaResumen

	if totalPropinas > 0 {
		divisor := float64(reporte.Resumen.CantidadMovimientos)
		if divisor <= 0 {
			divisor = 1
		}
		ds.Summary["ticket_promedio_propina"] = reportesRound(totalPropinas / divisor)
	} else {
		ds.Summary["ticket_promedio_propina"] = 0.0
	}
	if ds.RowCount > 0 {
		ds.Summary["usuario_top"] = ds.Rows[0]["usuario"]
		ds.Summary["usuario_top_total"] = ds.Rows[0]["propina_total"]
		ds.Summary["usuario_top_participacion_pct"] = ds.Rows[0]["participacion_pct"]
	}

	return ds, nil
}

func (b *reportesBuilder) buildOperativoComisionesLavadorDataset() (empresaReporteDataset, error) {
	limit := b.maxRows * 20
	if limit < 300 {
		limit = 300
	}
	if limit > 5000 {
		limit = 5000
	}

	reporte, err := dbpkg.GetEmpresaComisionesServicioReporte(b.db, b.empresaID, dbpkg.EmpresaComisionServicioMovimientoFilter{
		Desde:           b.desde,
		Hasta:           b.hasta,
		UsuarioLavador:  b.usuario,
		ServicioFiltro:  b.categoria,
		IncludeInactive: b.includeInactive,
		Limit:           limit,
	})
	if err != nil {
		return empresaReporteDataset{}, err
	}

	ds := b.newDataset(reporteDatasetOperativoComisiones, []string{
		"usuario_lavador",
		"movimientos",
		"base_servicios",
		"monto_comision",
		"ticket_comision",
		"participacion_pct",
	})

	if reporte == nil {
		ds.Summary["movimientos_total"] = 0
		ds.Summary["total_comisiones"] = 0.0
		ds.Summary["periodo_desde"] = reportesFirstNonBlank(strings.TrimSpace(b.desde), "sin_desde")
		ds.Summary["periodo_hasta"] = reportesFirstNonBlank(strings.TrimSpace(b.hasta), "sin_hasta")
		ds.Summary["moneda"] = "COP"
		return ds, nil
	}

	monedaResumen := ""
	monedaMixta := false
	serviciosUnicos := map[string]struct{}{}
	for _, mov := range reporte.Movimientos {
		moneda := strings.ToUpper(strings.TrimSpace(mov.Moneda))
		if moneda == "" {
			moneda = "COP"
		}
		if monedaResumen == "" {
			monedaResumen = moneda
		} else if monedaResumen != moneda {
			monedaMixta = true
		}

		servicioKey := strings.ToLower(strings.TrimSpace(reportesFirstNonBlank(mov.ServicioCodigo, mov.ServicioNombre, mov.ServicioCategoria)))
		if servicioKey != "" {
			serviciosUnicos[servicioKey] = struct{}{}
		}
	}

	lavadores := append([]dbpkg.EmpresaComisionServicioLavadorResumen{}, reporte.Lavadores...)
	shown := lavadores
	if len(shown) > b.maxRows {
		shown = shown[:b.maxRows]
	}

	totalComision := reportesRound(reporte.Resumen.TotalComisiones)
	for _, lavador := range shown {
		movimientos := lavador.CantidadMovimientos
		montoComision := reportesRound(lavador.TotalComision)
		ticketComision := 0.0
		if movimientos > 0 {
			ticketComision = reportesRound(montoComision / float64(movimientos))
		}
		participacion := 0.0
		if totalComision > 0 {
			participacion = reportesRound((montoComision * 100.0) / totalComision)
		}

		ds.Rows = append(ds.Rows, map[string]interface{}{
			"usuario_lavador":   reportesFirstNonBlank(strings.TrimSpace(lavador.UsuarioLavador), "sistema"),
			"movimientos":       movimientos,
			"base_servicios":    reportesRound(lavador.TotalBaseServicios),
			"monto_comision":    montoComision,
			"ticket_comision":   ticketComision,
			"participacion_pct": participacion,
		})
	}

	if monedaResumen == "" {
		monedaResumen = "COP"
	}
	if monedaMixta {
		monedaResumen = "MIXTA"
	}

	cfg := reporte.Configuracion
	if cfg == nil {
		cfg = &dbpkg.EmpresaComisionesServicioConfiguracion{
			EmpresaID:      b.empresaID,
			FiltroServicio: "lavado",
			Estado:         "activo",
		}
	}

	ds.RowCount = len(ds.Rows)
	ds.Summary["habilitar_comisiones"] = cfg.HabilitarComisiones
	ds.Summary["porcentaje_comision_config"] = reportesRound(cfg.PorcentajeComision)
	ds.Summary["filtro_servicio_config"] = reportesFirstNonBlank(strings.TrimSpace(cfg.FiltroServicio), "lavado")
	ds.Summary["aplicar_automaticamente"] = cfg.AplicarAutomaticamente
	ds.Summary["lavadores_totales"] = len(lavadores)
	ds.Summary["lavadores_listados"] = ds.RowCount
	ds.Summary["lavadores_con_comision"] = reporte.Resumen.LavadoresConComision
	ds.Summary["movimientos_total"] = reporte.Resumen.CantidadMovimientos
	ds.Summary["total_base_servicios"] = reportesRound(reporte.Resumen.TotalBaseServicios)
	ds.Summary["total_comisiones"] = totalComision
	ds.Summary["servicios_unicos"] = len(serviciosUnicos)
	ds.Summary["filtro_usuario"] = reportesFirstNonBlank(strings.TrimSpace(b.usuario), "todos")
	ds.Summary["filtro_servicio"] = reportesFirstNonBlank(strings.TrimSpace(b.categoria), "todos")
	ds.Summary["periodo_desde"] = reportesFirstNonBlank(strings.TrimSpace(reporte.Desde), strings.TrimSpace(b.desde), "sin_desde")
	ds.Summary["periodo_hasta"] = reportesFirstNonBlank(strings.TrimSpace(reporte.Hasta), strings.TrimSpace(b.hasta), "sin_hasta")
	ds.Summary["moneda"] = monedaResumen

	if reporte.Resumen.CantidadMovimientos > 0 {
		ds.Summary["ticket_promedio_comision"] = reportesRound(totalComision / float64(reporte.Resumen.CantidadMovimientos))
	} else {
		ds.Summary["ticket_promedio_comision"] = 0.0
	}
	if ds.RowCount > 0 {
		ds.Summary["lavador_top"] = ds.Rows[0]["usuario_lavador"]
		ds.Summary["lavador_top_total_comision"] = ds.Rows[0]["monto_comision"]
		ds.Summary["lavador_top_participacion_pct"] = ds.Rows[0]["participacion_pct"]
	}

	return ds, nil
}

func (b *reportesBuilder) buildOperativoFacturacionTrazabilidadDataset() (empresaReporteDataset, error) {
	type facturacionAgg struct {
		TipoDocumento       string
		Documentos          int
		Emitidas            int
		Anuladas            int
		Pendientes          int
		NotasCredito        int
		MontoTotal          float64
		ConNumeroLegal      int
		ConCodigoValidacion int
		ConTrazabilidad     int
		UltimaFecha         string
		Moneda              string
	}

	normalizeEstado := func(value string) string {
		v := strings.ToLower(strings.TrimSpace(value))
		v = strings.ReplaceAll(v, "-", "_")
		v = strings.ReplaceAll(v, " ", "_")
		return v
	}

	isEmitida := func(estado string) bool {
		switch normalizeEstado(estado) {
		case "emitida", "emitido", "validada", "aprobada", "contabilizada":
			return true
		default:
			return false
		}
	}

	isAnulada := func(estado string) bool {
		switch normalizeEstado(estado) {
		case "anulada", "anulado", "cancelada", "cancelado", "rechazada":
			return true
		default:
			return false
		}
	}

	isPendiente := func(estado string) bool {
		switch normalizeEstado(estado) {
		case "borrador", "pendiente", "por_emitir", "en_proceso":
			return true
		default:
			return false
		}
	}

	const pageSize = 300
	maxDocs := b.maxRows * 40
	if maxDocs < 500 {
		maxDocs = 500
	}
	if maxDocs > 10000 {
		maxDocs = 10000
	}

	documentos := make([]dbpkg.EmpresaDocumentoFacturacionListado, 0, maxDocs)
	for offset := 0; offset < maxDocs; {
		remaining := maxDocs - offset
		limit := pageSize
		if remaining < limit {
			limit = remaining
		}

		batch, listErr := dbpkg.ListEmpresaDocumentosFacturacionByEmpresa(b.db, dbpkg.EmpresaDocumentoFacturacionListFilter{
			EmpresaID:       b.empresaID,
			IncludeInactive: b.includeInactive,
			FechaDesde:      strings.TrimSpace(b.desde),
			FechaHasta:      strings.TrimSpace(b.hasta),
			Limit:           limit,
			Offset:          offset,
		})
		if listErr != nil {
			return empresaReporteDataset{}, listErr
		}

		documentos = append(documentos, batch...)
		if len(batch) < limit {
			break
		}
		offset += len(batch)
	}

	aggByTipo := make(map[string]*facturacionAgg)
	monedaResumen := ""
	monedaMixta := false

	totalDocumentos := 0
	totalEmitidas := 0
	totalAnuladas := 0
	totalPendientes := 0
	totalNotasCredito := 0
	totalConNumeroLegal := 0
	totalConCodigoValidacion := 0
	totalConTrazabilidad := 0
	totalMonto := 0.0

	for _, doc := range documentos {
		fechaDocumento := reportesFirstNonBlank(doc.FechaDocumento, doc.FechaActualizacion, doc.FechaCreacion)
		if !reportesDateWithinRange(fechaDocumento, b.desde, b.hasta) {
			continue
		}

		tipoDocumento := strings.ToLower(strings.TrimSpace(doc.TipoDocumento))
		if tipoDocumento == "" {
			tipoDocumento = "factura_electronica"
		}

		agg, ok := aggByTipo[tipoDocumento]
		if !ok {
			agg = &facturacionAgg{TipoDocumento: tipoDocumento}
			aggByTipo[tipoDocumento] = agg
		}

		agg.Documentos++
		totalDocumentos++

		montoDocumento := reportesRound(doc.MontoTotal)
		if montoDocumento < 0 {
			montoDocumento = 0
		}
		agg.MontoTotal += montoDocumento
		totalMonto += montoDocumento

		estadoDocumento := normalizeEstado(doc.EstadoDocumento)
		if isEmitida(estadoDocumento) {
			agg.Emitidas++
			totalEmitidas++
		}
		if isAnulada(estadoDocumento) {
			agg.Anuladas++
			totalAnuladas++
		}
		if isPendiente(estadoDocumento) {
			agg.Pendientes++
			totalPendientes++
		}
		if strings.Contains(tipoDocumento, "nota_credito") {
			agg.NotasCredito++
			totalNotasCredito++
		}

		if strings.TrimSpace(doc.NumeroLegal) != "" {
			agg.ConNumeroLegal++
			totalConNumeroLegal++
		}
		if strings.TrimSpace(doc.CodigoValidacion) != "" {
			agg.ConCodigoValidacion++
			totalConCodigoValidacion++
		}
		if strings.TrimSpace(doc.NumeroLegal) != "" && strings.TrimSpace(doc.CodigoValidacion) != "" {
			agg.ConTrazabilidad++
			totalConTrazabilidad++
		}

		if reportesDateUnix(fechaDocumento) > reportesDateUnix(agg.UltimaFecha) {
			agg.UltimaFecha = fechaDocumento
		}

		monedaDocumento := strings.ToUpper(strings.TrimSpace(doc.Moneda))
		if monedaDocumento == "" {
			monedaDocumento = "COP"
		}
		if agg.Moneda == "" {
			agg.Moneda = monedaDocumento
		} else if agg.Moneda != monedaDocumento {
			agg.Moneda = "MIXTA"
		}

		if monedaResumen == "" {
			monedaResumen = monedaDocumento
		} else if monedaResumen != monedaDocumento {
			monedaMixta = true
		}
	}

	consolidado := make([]*facturacionAgg, 0, len(aggByTipo))
	for _, item := range aggByTipo {
		consolidado = append(consolidado, item)
	}
	sort.SliceStable(consolidado, func(i, j int) bool {
		if consolidado[i].MontoTotal == consolidado[j].MontoTotal {
			if consolidado[i].Documentos == consolidado[j].Documentos {
				return consolidado[i].TipoDocumento < consolidado[j].TipoDocumento
			}
			return consolidado[i].Documentos > consolidado[j].Documentos
		}
		return consolidado[i].MontoTotal > consolidado[j].MontoTotal
	})

	ds := b.newDataset(reporteDatasetOperativoFacturacion, []string{
		"tipo_documento",
		"documentos",
		"emitidas",
		"anuladas",
		"pendientes",
		"notas_credito",
		"monto_total",
		"con_numero_legal",
		"con_codigo_validacion",
		"con_trazabilidad",
		"trazabilidad_pct",
		"ultima_fecha_documento",
		"moneda",
	})

	shown := consolidado
	if len(shown) > b.maxRows {
		shown = shown[:b.maxRows]
	}
	for _, item := range shown {
		trazabilidadPct := 0.0
		if item.Documentos > 0 {
			trazabilidadPct = reportesRound((float64(item.ConTrazabilidad) * 100.0) / float64(item.Documentos))
		}

		ds.Rows = append(ds.Rows, map[string]interface{}{
			"tipo_documento":         item.TipoDocumento,
			"documentos":             item.Documentos,
			"emitidas":               item.Emitidas,
			"anuladas":               item.Anuladas,
			"pendientes":             item.Pendientes,
			"notas_credito":          item.NotasCredito,
			"monto_total":            reportesRound(item.MontoTotal),
			"con_numero_legal":       item.ConNumeroLegal,
			"con_codigo_validacion":  item.ConCodigoValidacion,
			"con_trazabilidad":       item.ConTrazabilidad,
			"trazabilidad_pct":       trazabilidadPct,
			"ultima_fecha_documento": item.UltimaFecha,
			"moneda":                 reportesFirstNonBlank(strings.TrimSpace(item.Moneda), "COP"),
		})
	}

	if monedaResumen == "" {
		monedaResumen = "COP"
	}
	if monedaMixta {
		monedaResumen = "MIXTA"
	}

	trazabilidadGlobal := 0.0
	if totalDocumentos > 0 {
		trazabilidadGlobal = reportesRound((float64(totalConTrazabilidad) * 100.0) / float64(totalDocumentos))
	}

	ds.RowCount = len(ds.Rows)
	ds.Summary["documentos"] = totalDocumentos
	ds.Summary["documentos_total"] = totalDocumentos
	ds.Summary["tipos_documento"] = len(consolidado)
	ds.Summary["tipos_documento_listados"] = ds.RowCount
	ds.Summary["documentos_emitidos"] = totalEmitidas
	ds.Summary["documentos_anulados"] = totalAnuladas
	ds.Summary["documentos_pendientes"] = totalPendientes
	ds.Summary["notas_credito"] = totalNotasCredito
	ds.Summary["monto_total"] = reportesRound(totalMonto)
	ds.Summary["con_numero_legal"] = totalConNumeroLegal
	ds.Summary["con_codigo_validacion"] = totalConCodigoValidacion
	ds.Summary["documentos_trazables"] = totalConTrazabilidad
	ds.Summary["trazabilidad_pct"] = trazabilidadGlobal
	ds.Summary["periodo_desde"] = reportesFirstNonBlank(strings.TrimSpace(b.desde), "sin_desde")
	ds.Summary["periodo_hasta"] = reportesFirstNonBlank(strings.TrimSpace(b.hasta), "sin_hasta")
	ds.Summary["moneda"] = monedaResumen
	if totalDocumentos > 0 {
		ds.Summary["monto_promedio_documento"] = reportesRound(totalMonto / float64(totalDocumentos))
	} else {
		ds.Summary["monto_promedio_documento"] = 0.0
	}

	return ds, nil
}

func (b *reportesBuilder) buildOperativoAuditoriaAccionesDataset() (empresaReporteDataset, error) {
	type auditoriaAgg struct {
		Modulo      string
		Usuario     string
		Eventos     int
		Errores     int
		HTTP4xx     int
		HTTP5xx     int
		UltimaFecha string
		Acciones    map[string]int
		Recursos    map[string]struct{}
	}

	const pageSize = 300
	maxEventos := b.maxRows * 40
	if maxEventos < 500 {
		maxEventos = 500
	}
	if maxEventos > 12000 {
		maxEventos = 12000
	}

	eventos := make([]dbpkg.EmpresaAuditoriaEvento, 0, maxEventos)
	for offset := 0; offset < maxEventos; {
		remaining := maxEventos - offset
		limit := pageSize
		if remaining < limit {
			limit = remaining
		}

		batch, listErr := dbpkg.ListEmpresaAuditoriaEventos(b.db, b.empresaID, dbpkg.EmpresaAuditoriaEventoFilter{
			Desde:           b.desde,
			Hasta:           b.hasta,
			UsuarioCreador:  b.usuario,
			IncludeInactive: b.includeInactive,
			Limit:           limit,
			Offset:          offset,
		})
		if listErr != nil {
			return empresaReporteDataset{}, listErr
		}

		eventos = append(eventos, batch...)
		if len(batch) < limit {
			break
		}
		offset += len(batch)
	}

	aggByPair := make(map[string]*auditoriaAgg)
	modulos := make(map[string]struct{})
	usuarios := make(map[string]struct{})
	requestIDs := make(map[string]struct{})

	actionsCriticasSet := map[string]struct{}{
		"emitir":              {},
		"anular":              {},
		"nota_credito":        {},
		"emitir_nota_credito": {},
		"eliminar":            {},
		"desactivar":          {},
		"contabilizar":        {},
		"contabilizar_compra": {},
		"recepcionar":         {},
		"recepcionar_compra":  {},
		"emitir_orden":        {},
		"aprobar":             {},
		"cerrar":              {},
	}

	totalErrores := 0
	totalHTTP4xx := 0
	totalHTTP5xx := 0
	totalAccionesCriticas := 0

	for _, ev := range eventos {
		modulo := reportesFirstNonBlank(strings.TrimSpace(ev.Modulo), "sin_modulo")
		usuario := reportesFirstNonBlank(strings.TrimSpace(ev.UsuarioCreador), "sistema")
		key := strings.ToLower(modulo) + "|" + strings.ToLower(usuario)

		agg := aggByPair[key]
		if agg == nil {
			agg = &auditoriaAgg{
				Modulo:   modulo,
				Usuario:  usuario,
				Acciones: make(map[string]int),
				Recursos: make(map[string]struct{}),
			}
			aggByPair[key] = agg
		}

		agg.Eventos++

		hasError := false
		resultado := strings.ToLower(strings.TrimSpace(ev.Resultado))
		if resultado != "" && resultado != "ok" {
			hasError = true
		}
		if ev.CodigoHTTP >= 400 {
			hasError = true
		}
		if hasError {
			agg.Errores++
			totalErrores++
		}
		if ev.CodigoHTTP >= 400 && ev.CodigoHTTP < 500 {
			agg.HTTP4xx++
			totalHTTP4xx++
		}
		if ev.CodigoHTTP >= 500 {
			agg.HTTP5xx++
			totalHTTP5xx++
		}

		accion := strings.ToLower(strings.TrimSpace(reportesFirstNonBlank(ev.Accion, "sin_accion")))
		agg.Acciones[accion]++
		if _, ok := actionsCriticasSet[accion]; ok {
			totalAccionesCriticas++
		}

		recurso := strings.ToLower(strings.TrimSpace(ev.Recurso))
		if recurso != "" || ev.RecursoID > 0 {
			recursoKey := recurso + "#" + strconv.FormatInt(ev.RecursoID, 10)
			agg.Recursos[recursoKey] = struct{}{}
		}

		if reportesDateUnix(ev.FechaEvento) > reportesDateUnix(agg.UltimaFecha) {
			agg.UltimaFecha = ev.FechaEvento
		}

		modulos[modulo] = struct{}{}
		usuarios[usuario] = struct{}{}
		if reqID := strings.TrimSpace(ev.RequestID); reqID != "" {
			requestIDs[reqID] = struct{}{}
		}
	}

	consolidado := make([]*auditoriaAgg, 0, len(aggByPair))
	for _, item := range aggByPair {
		consolidado = append(consolidado, item)
	}

	ds := b.newDataset(reporteDatasetOperativoAuditoria, []string{
		"modulo",
		"usuario",
		"eventos",
		"errores",
		"http_4xx",
		"http_5xx",
		"error_pct",
		"recursos_unicos",
		"accion_principal",
		"ultima_fecha_evento",
	})

	sort.SliceStable(consolidado, func(i, j int) bool {
		if consolidado[i].Eventos == consolidado[j].Eventos {
			if consolidado[i].Errores == consolidado[j].Errores {
				if strings.EqualFold(consolidado[i].Modulo, consolidado[j].Modulo) {
					return strings.ToLower(consolidado[i].Usuario) < strings.ToLower(consolidado[j].Usuario)
				}
				return strings.ToLower(consolidado[i].Modulo) < strings.ToLower(consolidado[j].Modulo)
			}
			return consolidado[i].Errores > consolidado[j].Errores
		}
		return consolidado[i].Eventos > consolidado[j].Eventos
	})

	for _, item := range consolidado {
		accionPrincipal := "sin_accion"
		conteoAccionPrincipal := 0
		for accion, total := range item.Acciones {
			if total > conteoAccionPrincipal || (total == conteoAccionPrincipal && accion < accionPrincipal) {
				accionPrincipal = accion
				conteoAccionPrincipal = total
			}
		}

		errorPct := 0.0
		if item.Eventos > 0 {
			errorPct = reportesRound((float64(item.Errores) * 100.0) / float64(item.Eventos))
		}

		ds.Rows = append(ds.Rows, map[string]interface{}{
			"modulo":              item.Modulo,
			"usuario":             item.Usuario,
			"eventos":             item.Eventos,
			"errores":             item.Errores,
			"http_4xx":            item.HTTP4xx,
			"http_5xx":            item.HTTP5xx,
			"error_pct":           errorPct,
			"recursos_unicos":     len(item.Recursos),
			"accion_principal":    accionPrincipal,
			"ultima_fecha_evento": item.UltimaFecha,
		})
	}

	if len(ds.Rows) > b.maxRows {
		ds.Rows = ds.Rows[:b.maxRows]
	}

	ds.RowCount = len(ds.Rows)
	ds.Summary["eventos_total"] = len(eventos)
	ds.Summary["modulos_total"] = len(modulos)
	ds.Summary["usuarios_total"] = len(usuarios)
	ds.Summary["parejas_modulo_usuario"] = len(consolidado)
	ds.Summary["parejas_listadas"] = ds.RowCount
	ds.Summary["errores_total"] = totalErrores
	ds.Summary["http_4xx_total"] = totalHTTP4xx
	ds.Summary["http_5xx_total"] = totalHTTP5xx
	ds.Summary["acciones_criticas_total"] = totalAccionesCriticas
	ds.Summary["requests_unicos"] = len(requestIDs)
	ds.Summary["filtro_usuario"] = reportesFirstNonBlank(strings.TrimSpace(b.usuario), "todos")
	ds.Summary["periodo_desde"] = reportesFirstNonBlank(strings.TrimSpace(b.desde), "sin_desde")
	ds.Summary["periodo_hasta"] = reportesFirstNonBlank(strings.TrimSpace(b.hasta), "sin_hasta")

	if len(eventos) > 0 {
		ds.Summary["error_global_pct"] = reportesRound((float64(totalErrores) * 100.0) / float64(len(eventos)))
	} else {
		ds.Summary["error_global_pct"] = 0.0
	}
	if ds.RowCount > 0 {
		ds.Summary["modulo_top"] = ds.Rows[0]["modulo"]
		ds.Summary["usuario_top"] = ds.Rows[0]["usuario"]
		ds.Summary["eventos_top"] = ds.Rows[0]["eventos"]
	}

	return ds, nil
}

func (b *reportesBuilder) buildOperativoAsistenciaNominaAuditoriaDataset() (empresaReporteDataset, error) {
	limit := b.maxRows * 25
	if limit < 500 {
		limit = 500
	}
	if limit > 5000 {
		limit = 5000
	}

	rows, err := dbpkg.ListEmpresaAsistenciaEmpleados(
		b.db,
		b.empresaID,
		b.includeInactive,
		b.desde,
		b.hasta,
		"",
		"",
		limit,
	)
	if err != nil {
		return empresaReporteDataset{}, err
	}

	type asistenciaAuditoriaAgg struct {
		empleadoCodigo     string
		empleadoNombre     string
		empleadoDocumento  string
		cargo              string
		turno              string
		registros          int64
		registrosCompletos int64
		entradasMarcadas   int64
		salidasMarcadas    int64
		horasTotales       float64
		minutosTardeTotal  int64
		tardanzas          int64
		ausencias          int64
		inconsistencias    int64
		novedades          int64
		dias               map[string]struct{}
		estados            map[string]int64
	}

	bucket := make(map[string]*asistenciaAuditoriaAgg)
	for _, item := range rows {
		key := strings.ToLower(strings.TrimSpace(reportesFirstNonBlank(item.EmpleadoDocumento, item.EmpleadoCodigo, item.EmpleadoNombre, strconv.FormatInt(item.ID, 10))))
		if key == "" {
			continue
		}

		agg, ok := bucket[key]
		if !ok {
			agg = &asistenciaAuditoriaAgg{
				empleadoCodigo:    strings.TrimSpace(item.EmpleadoCodigo),
				empleadoNombre:    strings.TrimSpace(item.EmpleadoNombre),
				empleadoDocumento: strings.TrimSpace(item.EmpleadoDocumento),
				cargo:             strings.TrimSpace(item.Cargo),
				turno:             strings.TrimSpace(item.Turno),
				dias:              make(map[string]struct{}),
				estados:           make(map[string]int64),
			}
			bucket[key] = agg
		}

		agg.registros++
		if fecha := strings.TrimSpace(item.FechaAsistencia); fecha != "" {
			agg.dias[fecha] = struct{}{}
		}
		if strings.TrimSpace(item.Cargo) != "" {
			agg.cargo = strings.TrimSpace(item.Cargo)
		}
		if strings.TrimSpace(item.Turno) != "" {
			agg.turno = strings.TrimSpace(item.Turno)
		}
		if strings.TrimSpace(item.EmpleadoCodigo) != "" {
			agg.empleadoCodigo = strings.TrimSpace(item.EmpleadoCodigo)
		}
		if strings.TrimSpace(item.EmpleadoDocumento) != "" {
			agg.empleadoDocumento = strings.TrimSpace(item.EmpleadoDocumento)
		}
		if strings.TrimSpace(item.EmpleadoNombre) != "" {
			agg.empleadoNombre = strings.TrimSpace(item.EmpleadoNombre)
		}

		estadoAsistencia := strings.ToLower(strings.TrimSpace(item.EstadoAsistencia))
		if estadoAsistencia == "" {
			estadoAsistencia = "pendiente"
		}
		agg.estados[estadoAsistencia]++

		horaEntrada := strings.TrimSpace(item.HoraEntrada)
		horaSalida := strings.TrimSpace(item.HoraSalida)
		tieneEntrada := horaEntrada != ""
		tieneSalida := horaSalida != ""
		if tieneEntrada {
			agg.entradasMarcadas++
		}
		if tieneSalida {
			agg.salidasMarcadas++
		}
		if tieneEntrada && tieneSalida {
			agg.registrosCompletos++
		}

		agg.horasTotales += item.HorasTrabajadas
		if item.MinutosTarde > 0 {
			agg.minutosTardeTotal += int64(item.MinutosTarde)
			agg.tardanzas++
		} else if estadoAsistencia == "tarde" {
			agg.tardanzas++
		}
		if estadoAsistencia == "ausente" {
			agg.ausencias++
		}
		if strings.TrimSpace(item.Novedad) != "" {
			agg.novedades++
		}

		if (tieneEntrada && !tieneSalida && estadoAsistencia != "ausente" && estadoAsistencia != "permiso" && estadoAsistencia != "incapacidad" && estadoAsistencia != "vacaciones") || (!tieneEntrada && tieneSalida) {
			agg.inconsistencias++
		}
	}

	ds := b.newDataset(reporteDatasetOperativoAsistenciaNomina, []string{
		"empleado_codigo",
		"empleado_nombre",
		"empleado_documento",
		"cargo",
		"turno",
		"dias_registrados",
		"registros_asistencia",
		"registros_completos",
		"entradas_marcadas",
		"salidas_marcadas",
		"horas_trabajadas_total",
		"minutos_tarde_total",
		"tardanzas",
		"ausencias",
		"novedades",
		"inconsistencias",
		"estado_dominante",
		"completitud_registro_pct",
	})

	list := make([]map[string]interface{}, 0, len(bucket))
	totalRegistros := int64(0)
	totalCompletos := int64(0)
	totalEntradas := int64(0)
	totalSalidas := int64(0)
	totalHoras := 0.0
	totalMinutosTarde := int64(0)
	totalTardanzas := int64(0)
	totalAusencias := int64(0)
	totalNovedades := int64(0)
	totalInconsistencias := int64(0)

	for _, agg := range bucket {
		estadoDominante := "pendiente"
		maxEstado := int64(0)
		for estado, cnt := range agg.estados {
			if cnt > maxEstado {
				maxEstado = cnt
				estadoDominante = estado
			}
		}

		completitud := 0.0
		if agg.registros > 0 {
			completitud = reportesRound((float64(agg.registrosCompletos) * 100.0) / float64(agg.registros))
		}

		row := map[string]interface{}{
			"empleado_codigo":          agg.empleadoCodigo,
			"empleado_nombre":          agg.empleadoNombre,
			"empleado_documento":       agg.empleadoDocumento,
			"cargo":                    agg.cargo,
			"turno":                    agg.turno,
			"dias_registrados":         len(agg.dias),
			"registros_asistencia":     agg.registros,
			"registros_completos":      agg.registrosCompletos,
			"entradas_marcadas":        agg.entradasMarcadas,
			"salidas_marcadas":         agg.salidasMarcadas,
			"horas_trabajadas_total":   reportesRound(agg.horasTotales),
			"minutos_tarde_total":      agg.minutosTardeTotal,
			"tardanzas":                agg.tardanzas,
			"ausencias":                agg.ausencias,
			"novedades":                agg.novedades,
			"inconsistencias":          agg.inconsistencias,
			"estado_dominante":         estadoDominante,
			"completitud_registro_pct": completitud,
		}
		list = append(list, row)

		totalRegistros += agg.registros
		totalCompletos += agg.registrosCompletos
		totalEntradas += agg.entradasMarcadas
		totalSalidas += agg.salidasMarcadas
		totalHoras += agg.horasTotales
		totalMinutosTarde += agg.minutosTardeTotal
		totalTardanzas += agg.tardanzas
		totalAusencias += agg.ausencias
		totalNovedades += agg.novedades
		totalInconsistencias += agg.inconsistencias
	}

	toFloat := func(v interface{}) float64 {
		switch value := v.(type) {
		case float64:
			return value
		case float32:
			return float64(value)
		case int:
			return float64(value)
		case int64:
			return float64(value)
		case int32:
			return float64(value)
		case json.Number:
			f, _ := value.Float64()
			return f
		case string:
			f, _ := strconv.ParseFloat(strings.TrimSpace(value), 64)
			return f
		default:
			return 0
		}
	}

	toInt := func(v interface{}) int64 {
		switch value := v.(type) {
		case int:
			return int64(value)
		case int32:
			return int64(value)
		case int64:
			return value
		case float64:
			return int64(value)
		case json.Number:
			i, _ := value.Int64()
			return i
		case string:
			i, _ := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
			return i
		default:
			return 0
		}
	}

	sort.Slice(list, func(i, j int) bool {
		horasI := toFloat(list[i]["horas_trabajadas_total"])
		horasJ := toFloat(list[j]["horas_trabajadas_total"])
		if horasI == horasJ {
			regI := toInt(list[i]["registros_asistencia"])
			regJ := toInt(list[j]["registros_asistencia"])
			if regI == regJ {
				nameI, _ := list[i]["empleado_nombre"].(string)
				nameJ, _ := list[j]["empleado_nombre"].(string)
				return strings.ToLower(strings.TrimSpace(nameI)) < strings.ToLower(strings.TrimSpace(nameJ))
			}
			return regI > regJ
		}
		return horasI > horasJ
	})

	if len(list) > b.maxRows {
		list = list[:b.maxRows]
	}

	ds.Rows = append(ds.Rows, list...)
	ds.RowCount = len(ds.Rows)
	ds.Summary["empleados_auditados"] = len(bucket)
	ds.Summary["filas_listadas"] = ds.RowCount
	ds.Summary["registros_total"] = totalRegistros
	ds.Summary["registros_completos_total"] = totalCompletos
	ds.Summary["entradas_marcadas_total"] = totalEntradas
	ds.Summary["salidas_marcadas_total"] = totalSalidas
	ds.Summary["horas_trabajadas_total"] = reportesRound(totalHoras)
	ds.Summary["minutos_tarde_total"] = totalMinutosTarde
	ds.Summary["tardanzas_total"] = totalTardanzas
	ds.Summary["ausencias_total"] = totalAusencias
	ds.Summary["novedades_total"] = totalNovedades
	ds.Summary["inconsistencias_total"] = totalInconsistencias
	ds.Summary["periodo_desde"] = reportesFirstNonBlank(strings.TrimSpace(b.desde), "sin_desde")
	ds.Summary["periodo_hasta"] = reportesFirstNonBlank(strings.TrimSpace(b.hasta), "sin_hasta")
	if totalRegistros > 0 {
		ds.Summary["completitud_global_pct"] = reportesRound((float64(totalCompletos) * 100.0) / float64(totalRegistros))
	} else {
		ds.Summary["completitud_global_pct"] = 0.0
	}

	return ds, nil
}

func (b *reportesBuilder) buildOperativoVehiculosPermanenciaDataset() (empresaReporteDataset, error) {
	limit := b.maxRows * 20
	if limit < 500 {
		limit = 500
	}
	if limit > 5000 {
		limit = 5000
	}

	rows, err := dbpkg.ListEmpresaVehiculosPermanenciaReporte(
		b.db,
		b.empresaID,
		b.includeInactive,
		b.desde,
		b.hasta,
		"",
		"",
		limit,
	)
	if err != nil {
		return empresaReporteDataset{}, err
	}

	ds := b.newDataset(reporteDatasetOperativoVehiculos, []string{
		"registro_id",
		"patente",
		"tipo_vehiculo",
		"conductor_nombre",
		"propietario_nombre",
		"fecha_ingreso",
		"fecha_salida",
		"estado_registro",
		"estado",
		"minutos_estadia",
		"horas_estadia",
		"dias_estadia",
	})

	totalMinutos := int64(0)
	vehiculosEnEmpresa := 0
	vehiculosRetirados := 0
	maxMinutos := int64(0)

	for _, item := range rows {
		if item.MinutosEstadia < 0 {
			item.MinutosEstadia = 0
		}
		totalMinutos += item.MinutosEstadia
		if item.MinutosEstadia > maxMinutos {
			maxMinutos = item.MinutosEstadia
		}

		estadoRegistro := strings.ToLower(strings.TrimSpace(item.EstadoRegistro))
		switch estadoRegistro {
		case "retirado":
			vehiculosRetirados++
		default:
			vehiculosEnEmpresa++
		}

		ds.Rows = append(ds.Rows, map[string]interface{}{
			"registro_id":        item.ID,
			"patente":            item.Patente,
			"tipo_vehiculo":      item.TipoVehiculo,
			"conductor_nombre":   item.ConductorNombre,
			"propietario_nombre": item.PropietarioNombre,
			"fecha_ingreso":      item.FechaIngreso,
			"fecha_salida":       item.FechaSalida,
			"estado_registro":    item.EstadoRegistro,
			"estado":             item.Estado,
			"minutos_estadia":    item.MinutosEstadia,
			"horas_estadia":      reportesRound(item.HorasEstadia),
			"dias_estadia":       reportesRound(item.DiasEstadia),
		})
	}

	if len(ds.Rows) > b.maxRows {
		ds.Rows = ds.Rows[:b.maxRows]
	}

	ds.RowCount = len(ds.Rows)
	ds.Summary["registros_total"] = len(rows)
	ds.Summary["registros_listados"] = ds.RowCount
	ds.Summary["vehiculos_en_empresa"] = vehiculosEnEmpresa
	ds.Summary["vehiculos_retirados"] = vehiculosRetirados
	ds.Summary["minutos_estadia_total"] = totalMinutos
	ds.Summary["horas_estadia_total"] = reportesRound(float64(totalMinutos) / 60.0)
	ds.Summary["dias_estadia_total"] = reportesRound(float64(totalMinutos) / 1440.0)
	ds.Summary["maximo_minutos_estadia"] = maxMinutos
	ds.Summary["periodo_desde"] = reportesFirstNonBlank(strings.TrimSpace(b.desde), "sin_desde")
	ds.Summary["periodo_hasta"] = reportesFirstNonBlank(strings.TrimSpace(b.hasta), "sin_hasta")
	if len(rows) > 0 {
		ds.Summary["promedio_minutos_estadia"] = reportesRound(float64(totalMinutos) / float64(len(rows)))
		ds.Summary["promedio_horas_estadia"] = reportesRound((float64(totalMinutos) / 60.0) / float64(len(rows)))
		ds.Summary["promedio_dias_estadia"] = reportesRound((float64(totalMinutos) / 1440.0) / float64(len(rows)))
	} else {
		ds.Summary["promedio_minutos_estadia"] = 0.0
		ds.Summary["promedio_horas_estadia"] = 0.0
		ds.Summary["promedio_dias_estadia"] = 0.0
	}

	return ds, nil
}

func (b *reportesBuilder) buildContableFlujoCajaDataset() (empresaReporteDataset, error) {
	limit := b.maxRows * 20
	if limit < 1000 {
		limit = 1000
	}
	movimientos, err := dbpkg.ListEmpresaFinanzasMovimientos(b.db, b.empresaID, dbpkg.EmpresaFinanzasMovimientoFilter{
		Desde:           b.desde,
		Hasta:           b.hasta,
		IncludeInactive: b.includeInactive,
		Limit:           limit,
	})
	if err != nil {
		return empresaReporteDataset{}, err
	}

	type flujoDia struct {
		ingresos    float64
		egresos     float64
		movimientos int
	}

	porFecha := make(map[string]*flujoDia)
	filtroCategoria := strings.ToLower(strings.TrimSpace(b.categoria))
	filtroMetodoPago := reportesNormalizeMetodoPagoFinanzas(b.metodoPago)
	moneda := ""
	for _, mov := range movimientos {
		if filtroCategoria != "" && strings.ToLower(strings.TrimSpace(mov.Categoria)) != filtroCategoria {
			continue
		}
		if filtroMetodoPago != "" && reportesNormalizeMetodoPagoFinanzas(mov.MetodoPago) != filtroMetodoPago {
			continue
		}

		fecha := reportesNormalizeDatePart(reportesFirstNonBlank(mov.FechaMovimiento, mov.FechaActualizacion, mov.FechaCreacion))
		if fecha == "" {
			continue
		}
		if strings.TrimSpace(moneda) == "" {
			moneda = strings.TrimSpace(mov.Moneda)
		}

		item, ok := porFecha[fecha]
		if !ok {
			item = &flujoDia{}
			porFecha[fecha] = item
		}

		netoMovimiento := reportesRound(reportesMovimientoTotalNeto(mov))
		tipo := strings.ToLower(strings.TrimSpace(mov.TipoMovimiento))
		switch tipo {
		case "ingreso":
			item.ingresos += netoMovimiento
		case "egreso":
			item.egresos += netoMovimiento
		default:
			if netoMovimiento >= 0 {
				item.ingresos += netoMovimiento
			} else {
				item.egresos += math.Abs(netoMovimiento)
			}
		}
		item.movimientos++
	}

	fechas := make([]string, 0, len(porFecha))
	for fecha := range porFecha {
		fechas = append(fechas, fecha)
	}
	sort.Strings(fechas)
	if len(fechas) > b.maxRows {
		fechas = fechas[len(fechas)-b.maxRows:]
	}

	ds := b.newDataset(reporteDatasetContableFlujoCaja, []string{
		"fecha",
		"ingresos",
		"egresos",
		"neto_dia",
		"saldo_acumulado",
		"movimientos",
	})

	saldoAcumulado := 0.0
	totalIngresos := 0.0
	totalEgresos := 0.0
	totalMovimientos := 0
	for _, fecha := range fechas {
		item := porFecha[fecha]
		ingresos := reportesRound(item.ingresos)
		egresos := reportesRound(item.egresos)
		netoDia := reportesRound(ingresos - egresos)
		saldoAcumulado = reportesRound(saldoAcumulado + netoDia)
		totalIngresos += ingresos
		totalEgresos += egresos
		totalMovimientos += item.movimientos

		ds.Rows = append(ds.Rows, map[string]interface{}{
			"fecha":           fecha,
			"ingresos":        ingresos,
			"egresos":         egresos,
			"neto_dia":        netoDia,
			"saldo_acumulado": saldoAcumulado,
			"movimientos":     item.movimientos,
		})
	}

	ds.RowCount = len(ds.Rows)
	ds.Summary["dias"] = ds.RowCount
	ds.Summary["movimientos_total"] = totalMovimientos
	ds.Summary["total_ingresos"] = reportesRound(totalIngresos)
	ds.Summary["total_egresos"] = reportesRound(totalEgresos)
	ds.Summary["neto_periodo"] = reportesRound(totalIngresos - totalEgresos)
	ds.Summary["saldo_final"] = reportesRound(saldoAcumulado)
	ds.Summary["filtro_categoria"] = reportesFirstNonBlank(strings.TrimSpace(b.categoria), "todas")
	ds.Summary["filtro_metodo_pago"] = reportesFirstNonBlank(reportesNormalizeMetodoPagoFinanzas(b.metodoPago), "todos")
	if ds.RowCount > 0 {
		ds.Summary["promedio_neto_dia"] = reportesRound((totalIngresos - totalEgresos) / float64(ds.RowCount))
	}
	ds.Summary["moneda"] = reportesFirstNonBlank(strings.ToUpper(strings.TrimSpace(moneda)), "COP")

	return ds, nil
}

func (b *reportesBuilder) buildContableMovimientosFinancierosDataset() (empresaReporteDataset, error) {
	rows, err := dbpkg.ListEmpresaFinanzasMovimientos(b.db, b.empresaID, dbpkg.EmpresaFinanzasMovimientoFilter{
		Desde:           b.desde,
		Hasta:           b.hasta,
		IncludeInactive: b.includeInactive,
		Limit:           b.maxRows,
	})
	if err != nil {
		return empresaReporteDataset{}, err
	}
	ds := b.newDataset(reporteDatasetContableMovFin, []string{
		"fecha_movimiento",
		"tipo_movimiento",
		"codigo",
		"categoria",
		"concepto",
		"metodo_pago",
		"monto",
		"total_neto",
		"estado",
	})
	ingresos := 0.0
	egresos := 0.0
	for _, mov := range rows {
		tipo := strings.ToLower(strings.TrimSpace(mov.TipoMovimiento))
		if tipo == "ingreso" {
			ingresos += mov.TotalNeto
		}
		if tipo == "egreso" {
			egresos += mov.TotalNeto
		}
		ds.Rows = append(ds.Rows, map[string]interface{}{
			"fecha_movimiento": mov.FechaMovimiento,
			"tipo_movimiento":  tipo,
			"codigo":           mov.Codigo,
			"categoria":        mov.Categoria,
			"concepto":         mov.Concepto,
			"metodo_pago":      mov.MetodoPago,
			"monto":            reportesRound(mov.Monto),
			"total_neto":       reportesRound(mov.TotalNeto),
			"estado":           mov.Estado,
		})
	}
	ds.RowCount = len(ds.Rows)
	ds.Summary["ingresos"] = reportesRound(ingresos)
	ds.Summary["egresos"] = reportesRound(egresos)
	ds.Summary["balance"] = reportesRound(ingresos - egresos)
	return ds, nil
}

func (b *reportesBuilder) buildContableEventosDataset() (empresaReporteDataset, error) {
	eventos, err := dbpkg.ListEmpresaEventosContables(b.db, b.empresaID, dbpkg.EmpresaEventoContableFilter{
		IncludeInactive: b.includeInactive,
		Limit:           b.maxRows,
	})
	if err != nil {
		return empresaReporteDataset{}, err
	}
	ds := b.newDataset(reporteDatasetContableEventos, []string{
		"fecha_evento",
		"modulo",
		"evento",
		"entidad",
		"documento_codigo",
		"periodo_contable",
		"monto_total",
		"procesado",
		"asiento_contable_id",
		"error_procesamiento",
	})
	pendientes := 0
	procesados := 0
	totalMonto := 0.0
	for _, ev := range eventos {
		if !reportesDateWithinRange(ev.FechaEvento, b.desde, b.hasta) {
			continue
		}
		if ev.Procesado {
			procesados++
		} else {
			pendientes++
		}
		totalMonto += ev.MontoTotal
		ds.Rows = append(ds.Rows, map[string]interface{}{
			"fecha_evento":        ev.FechaEvento,
			"modulo":              ev.Modulo,
			"evento":              ev.Evento,
			"entidad":             ev.Entidad,
			"documento_codigo":    ev.DocumentoCodigo,
			"periodo_contable":    ev.PeriodoContable,
			"monto_total":         reportesRound(ev.MontoTotal),
			"procesado":           ev.Procesado,
			"asiento_contable_id": ev.AsientoContableID,
			"error_procesamiento": ev.ErrorProcesamiento,
		})
	}
	ds.RowCount = len(ds.Rows)
	ds.Summary["procesados"] = procesados
	ds.Summary["pendientes"] = pendientes
	ds.Summary["monto_total"] = reportesRound(totalMonto)
	return ds, nil
}

func (b *reportesBuilder) buildContableAsientosDataset() (empresaReporteDataset, error) {
	asientos, err := dbpkg.ListEmpresaAsientosContables(b.db, b.empresaID, dbpkg.EmpresaAsientoContableFilter{
		Desde:           b.desde,
		Hasta:           b.hasta,
		IncludeInactive: b.includeInactive,
		Limit:           b.maxRows,
	})
	if err != nil {
		return empresaReporteDataset{}, err
	}
	ds := b.newDataset(reporteDatasetContableAsientos, []string{
		"fecha_asiento",
		"modulo",
		"evento",
		"documento_codigo",
		"periodo_contable",
		"total_debito",
		"total_credito",
		"diferencia",
		"estado",
	})
	totalDebito := 0.0
	totalCredito := 0.0
	for _, as := range asientos {
		totalDebito += as.TotalDebito
		totalCredito += as.TotalCredito
		ds.Rows = append(ds.Rows, map[string]interface{}{
			"fecha_asiento":    as.FechaAsiento,
			"modulo":           as.Modulo,
			"evento":           as.Evento,
			"documento_codigo": as.DocumentoCodigo,
			"periodo_contable": as.PeriodoContable,
			"total_debito":     reportesRound(as.TotalDebito),
			"total_credito":    reportesRound(as.TotalCredito),
			"diferencia":       reportesRound(as.Diferencia),
			"estado":           as.Estado,
		})
	}
	ds.RowCount = len(ds.Rows)
	ds.Summary["total_debito"] = reportesRound(totalDebito)
	ds.Summary["total_credito"] = reportesRound(totalCredito)
	ds.Summary["desfase"] = reportesRound(totalDebito - totalCredito)
	return ds, nil
}

func (b *reportesBuilder) buildContableNominaLiquidacionesDataset() (empresaReporteDataset, error) {
	rows, err := dbpkg.ListEmpresaNominaLiquidaciones(b.db, b.empresaID, dbpkg.EmpresaNominaLiquidacionFilter{
		PeriodoDesde:     b.desde,
		PeriodoHasta:     b.hasta,
		EmpleadoNominaID: b.empleadoNominaID,
		IncludeInactive:  b.includeInactive,
		Limit:            b.maxRows,
	})
	if err != nil {
		return empresaReporteDataset{}, err
	}

	ds := b.newDataset(reporteDatasetContableNomina, []string{
		"fecha_generacion",
		"periodo_desde",
		"periodo_hasta",
		"empleado_nomina_id",
		"empleado_nombre",
		"empleado_documento",
		"cargo",
		"dias_liquidados",
		"horas_asistencia_total",
		"horas_extra_totales",
		"devengado_total",
		"deduccion_total",
		"neto_pagar",
		"estado",
	})

	totalDevengado := 0.0
	totalDeduccion := 0.0
	totalNeto := 0.0

	for _, item := range rows {
		horasExtraTotales := reportesRound(
			item.HorasExtraDiurnas +
				item.HorasExtraNocturnas +
				item.HorasExtraDominicalesDiurnas +
				item.HorasExtraDominicalesNocturnas,
		)
		totalDevengado += item.DevengadoTotal
		totalDeduccion += item.DeduccionTotal
		totalNeto += item.NetoPagar

		ds.Rows = append(ds.Rows, map[string]interface{}{
			"fecha_generacion":       reportesFirstNonBlank(item.FechaGeneracion, item.FechaActualizacion, item.FechaCreacion),
			"periodo_desde":          item.PeriodoDesde,
			"periodo_hasta":          item.PeriodoHasta,
			"empleado_nomina_id":     item.EmpleadoNominaID,
			"empleado_nombre":        item.EmpleadoNombre,
			"empleado_documento":     item.EmpleadoDocumento,
			"cargo":                  item.Cargo,
			"dias_liquidados":        reportesRound(item.DiasLiquidados),
			"horas_asistencia_total": reportesRound(item.HorasAsistenciaTotal),
			"horas_extra_totales":    horasExtraTotales,
			"devengado_total":        reportesRound(item.DevengadoTotal),
			"deduccion_total":        reportesRound(item.DeduccionTotal),
			"neto_pagar":             reportesRound(item.NetoPagar),
			"estado":                 item.Estado,
		})
	}

	ds.RowCount = len(ds.Rows)
	ds.Summary["filtro_empleado_nomina_id"] = b.empleadoNominaID
	ds.Summary["liquidaciones"] = ds.RowCount
	ds.Summary["total_devengado"] = reportesRound(totalDevengado)
	ds.Summary["total_deduccion"] = reportesRound(totalDeduccion)
	ds.Summary["total_neto"] = reportesRound(totalNeto)
	if ds.RowCount > 0 {
		ds.Summary["promedio_neto"] = reportesRound(totalNeto / float64(ds.RowCount))
	}

	if cfg, cfgErr := dbpkg.GetEmpresaNominaConfiguracion(b.db, b.empresaID); cfgErr == nil && cfg != nil {
		ds.Summary["moneda"] = reportesFirstNonBlank(strings.ToUpper(strings.TrimSpace(cfg.Moneda)), "COP")
	} else {
		ds.Summary["moneda"] = "COP"
	}

	return ds, nil
}

func writeReportesDatasetExport(w http.ResponseWriter, ds empresaReporteDataset, format string) error {
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "excel" {
		format = "xls"
	}
	if format == "tsv" {
		format = "xls"
	}

	switch format {
	case "json":
		fileName := reportesBuildFileName(ds.Key, ds.EmpresaID, "json")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
		writeJSON(w, http.StatusOK, ds)
		return nil
	case "csv":
		content, err := reportesDatasetCSVContent(ds)
		if err != nil {
			return fmt.Errorf("no se pudo generar CSV")
		}
		fileName := reportesBuildFileName(ds.Key, ds.EmpresaID, "csv")
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(content))
		return nil
	case "txt":
		content := reportesDatasetTXTContent(ds)
		fileName := reportesBuildFileName(ds.Key, ds.EmpresaID, "txt")
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(content))
		return nil
	case "xls":
		content := reportesDatasetTSVContent(ds)
		fileName := reportesBuildFileName(ds.Key, ds.EmpresaID, "xls")
		w.Header().Set("Content-Type", "application/vnd.ms-excel; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("\ufeff" + content))
		return nil
	case "pdf":
		content := reportesDatasetPDFContent(ds)
		fileName := reportesBuildFileName(ds.Key, ds.EmpresaID, "pdf")
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
		return nil
	default:
		return fmt.Errorf("format invalido (use json, csv, txt, xls o pdf)")
	}
}

func reportesDatasetCSVContent(ds empresaReporteDataset) (string, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	if err := writer.Write(ds.Columns); err != nil {
		return "", err
	}
	for _, row := range ds.Rows {
		record := make([]string, len(ds.Columns))
		for i, col := range ds.Columns {
			record[i] = reportesStringValue(row[col])
		}
		if err := writer.Write(record); err != nil {
			return "", err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}
	return builder.String(), nil
}

func reportesDatasetTSVContent(ds empresaReporteDataset) string {
	var builder strings.Builder
	builder.WriteString(strings.Join(ds.Columns, "\t"))
	builder.WriteString("\n")
	for _, row := range ds.Rows {
		values := make([]string, len(ds.Columns))
		for i, col := range ds.Columns {
			values[i] = strings.ReplaceAll(reportesStringValue(row[col]), "\t", " ")
		}
		builder.WriteString(strings.Join(values, "\t"))
		builder.WriteString("\n")
	}
	return builder.String()
}

func reportesDatasetTXTContent(ds empresaReporteDataset) string {
	var builder strings.Builder
	builder.WriteString("Reporte: ")
	builder.WriteString(ds.Title)
	builder.WriteString("\n")
	builder.WriteString("Nivel: ")
	builder.WriteString(ds.Level)
	builder.WriteString("\n")
	builder.WriteString("Empresa: ")
	builder.WriteString(strconv.FormatInt(ds.EmpresaID, 10))
	builder.WriteString("\n")
	builder.WriteString("Rango: ")
	builder.WriteString(reportesFirstNonBlank(ds.Desde, "sin_desde"))
	builder.WriteString(" .. ")
	builder.WriteString(reportesFirstNonBlank(ds.Hasta, "sin_hasta"))
	builder.WriteString("\n")
	builder.WriteString("Generado: ")
	builder.WriteString(ds.GeneratedAt)
	builder.WriteString("\n\n")

	if len(ds.Summary) > 0 {
		keys := make([]string, 0, len(ds.Summary))
		for k := range ds.Summary {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		builder.WriteString("Resumen:\n")
		for _, k := range keys {
			builder.WriteString("- ")
			builder.WriteString(k)
			builder.WriteString(": ")
			builder.WriteString(reportesStringValue(ds.Summary[k]))
			builder.WriteString("\n")
		}
		builder.WriteString("\n")
	}

	builder.WriteString(reportesDatasetTSVContent(ds))
	return builder.String()
}

func reportesDatasetPDFContent(ds empresaReporteDataset) []byte {
	lines := reportesDatasetPDFLines(ds)
	if len(lines) > 46 {
		lines = append(lines[:45], "Salida truncada. Use CSV o JSON para detalle completo.")
	}

	var streamBuilder strings.Builder
	streamBuilder.WriteString("BT\n/F1 9 Tf\n13 TL\n50 760 Td\n")
	for idx, line := range lines {
		if idx > 0 {
			streamBuilder.WriteString("T*\n")
		}
		streamBuilder.WriteString("(")
		streamBuilder.WriteString(reportesEscapePDFText(line))
		streamBuilder.WriteString(") Tj\n")
	}
	streamBuilder.WriteString("ET\n")
	stream := streamBuilder.String()

	var pdf bytes.Buffer
	offsets := make([]int, 6)

	pdf.WriteString("%PDF-1.4\n")
	offsets[1] = pdf.Len()
	pdf.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")
	offsets[2] = pdf.Len()
	pdf.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")
	offsets[3] = pdf.Len()
	pdf.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>\nendobj\n")
	offsets[4] = pdf.Len()
	pdf.WriteString(fmt.Sprintf("4 0 obj\n<< /Length %d >>\nstream\n%sendstream\nendobj\n", len(stream), stream))
	offsets[5] = pdf.Len()
	pdf.WriteString("5 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n")

	startXRef := pdf.Len()
	pdf.WriteString("xref\n0 6\n")
	pdf.WriteString("0000000000 65535 f \n")
	for i := 1; i <= 5; i++ {
		pdf.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}
	pdf.WriteString("trailer\n<< /Size 6 /Root 1 0 R >>\n")
	pdf.WriteString(fmt.Sprintf("startxref\n%d\n%%%%EOF", startXRef))

	return pdf.Bytes()
}

func reportesDatasetPDFLines(ds empresaReporteDataset) []string {
	lines := []string{
		"Reporte: " + ds.Title,
		"Nivel: " + ds.Level,
		"Empresa: " + strconv.FormatInt(ds.EmpresaID, 10),
		"Rango: " + reportesFirstNonBlank(ds.Desde, "sin_desde") + " .. " + reportesFirstNonBlank(ds.Hasta, "sin_hasta"),
		"Generado: " + ds.GeneratedAt,
		"",
	}

	if len(ds.Summary) > 0 {
		keys := make([]string, 0, len(ds.Summary))
		for k := range ds.Summary {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		lines = append(lines, "Resumen:")
		for _, key := range keys {
			lines = append(lines, "- "+key+": "+reportesStringValue(ds.Summary[key]))
		}
		lines = append(lines, "")
	}

	if len(ds.Columns) > 0 {
		lines = append(lines, strings.Join(ds.Columns, " | "))
	}

	for _, row := range ds.Rows {
		values := make([]string, len(ds.Columns))
		for i, col := range ds.Columns {
			values[i] = reportesStringValue(row[col])
		}
		lines = append(lines, strings.Join(values, " | "))
	}

	if len(ds.Rows) == 0 {
		lines = append(lines, "Sin filas para el rango consultado.")
	}

	return lines
}

func reportesEscapePDFText(raw string) string {
	raw = strings.ReplaceAll(raw, "\\", "\\\\")
	raw = strings.ReplaceAll(raw, "(", "\\(")
	raw = strings.ReplaceAll(raw, ")", "\\)")

	var builder strings.Builder
	for _, r := range raw {
		switch {
		case r == '\n' || r == '\r' || r == '\t':
			builder.WriteByte(' ')
		case r < 32:
			continue
		case r > 126:
			builder.WriteByte('?')
		default:
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func reportesBuildFileName(base string, empresaID int64, ext string) string {
	safe := strings.ToLower(strings.TrimSpace(base))
	safe = strings.ReplaceAll(safe, " ", "_")
	if safe == "" {
		safe = "reporte"
	}
	stamp := time.Now().Format("20060102_150405")
	return safe + "_empresa_" + strconv.FormatInt(empresaID, 10) + "_" + stamp + "." + ext
}

func reportesStringValue(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case int32:
		return strconv.FormatInt(int64(t), 10)
	case float64:
		return strconv.FormatFloat(t, 'f', 2, 64)
	case float32:
		return strconv.FormatFloat(float64(t), 'f', 2, 64)
	default:
		raw, err := json.Marshal(t)
		if err != nil {
			return fmt.Sprintf("%v", t)
		}
		return string(raw)
	}
}

func reportesVentaTotal(venta dbpkg.CarritoCompra) float64 {
	if venta.TotalPagado > 0 {
		return reportesRound(venta.TotalPagado)
	}
	return reportesRound(venta.Total)
}

func reportesMovimientoTotalNeto(mov dbpkg.EmpresaFinanzasMovimiento) float64 {
	if mov.TotalNeto > 0 {
		return reportesRound(mov.TotalNeto)
	}
	if mov.Total > 0 {
		return reportesRound(mov.Total)
	}
	return reportesRound(mov.Monto)
}

func reportesStringEqualsFoldAny(expected string, values ...string) bool {
	expected = strings.ToLower(strings.TrimSpace(expected))
	if expected == "" {
		return false
	}
	for _, value := range values {
		if strings.ToLower(strings.TrimSpace(value)) == expected {
			return true
		}
	}
	return false
}

func reportesEstadoActivo(estado string) bool {
	e := strings.ToLower(strings.TrimSpace(estado))
	return e == "" || e == "activo"
}

func reportesRound(v float64) float64 {
	return math.Round(v*100) / 100
}

func reportesNormalizeMetodoPagoFinanzas(v string) string {
	n := strings.ToLower(strings.TrimSpace(v))
	switch n {
	case "":
		return ""
	case "transferencia":
		return "transferencia_bancaria"
	default:
		return n
	}
}

func reportesFirstNonBlank(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func reportesEstadoStock(cantidad, minimo, maximo float64) (string, int) {
	if cantidad <= 0 {
		return "sin_stock", 0
	}
	if minimo > 0 && cantidad <= minimo {
		return "bajo_minimo", 1
	}
	if maximo > 0 && cantidad >= maximo {
		return "sobre_stock", 3
	}
	return "ok", 2
}

func reportesInventarioKey(productoID, bodegaID int64) string {
	return strconv.FormatInt(productoID, 10) + ":" + strconv.FormatInt(bodegaID, 10)
}

func reportesInventarioRiesgoPrioridad(estadoProyeccion, estadoStock string, prioridadStock int) int {
	switch strings.ToLower(strings.TrimSpace(estadoProyeccion)) {
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
	case "estable":
		return 5
	}

	if strings.EqualFold(strings.TrimSpace(estadoStock), "sin_stock") {
		return 0
	}
	if strings.EqualFold(strings.TrimSpace(estadoStock), "bajo_minimo") {
		return 1
	}
	if prioridadStock < 0 {
		prioridadStock = 0
	}
	return 10 + prioridadStock
}

func reportesDateWithinRange(rawDateTime, desde, hasta string) bool {
	desde = strings.TrimSpace(desde)
	hasta = strings.TrimSpace(hasta)
	if desde == "" && hasta == "" {
		return true
	}
	datePart := reportesNormalizeDatePart(rawDateTime)
	if datePart == "" {
		return false
	}
	if desde != "" && datePart < desde {
		return false
	}
	if hasta != "" && datePart > hasta {
		return false
	}
	return true
}

func reportesNormalizeDatePart(rawDateTime string) string {
	value := strings.TrimSpace(rawDateTime)
	if value == "" {
		return ""
	}
	if len(value) >= 10 {
		candidate := value[:10]
		if _, err := time.Parse("2006-01-02", candidate); err == nil {
			return candidate
		}
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.Format("2006-01-02")
		}
	}
	return ""
}

func reportesDateUnix(rawDateTime string) int64 {
	value := strings.TrimSpace(rawDateTime)
	if value == "" {
		return 0
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.Unix()
		}
	}
	if part := reportesNormalizeDatePart(value); part != "" {
		if parsed, err := time.Parse("2006-01-02", part); err == nil {
			return parsed.Unix()
		}
	}
	return 0
}

func reportesResolveReservaPeriodo(desde, hasta string, reservas []dbpkg.ReservaHotel) (string, string) {
	periodoDesde := reportesNormalizeDatePart(desde)
	periodoHasta := reportesNormalizeDatePart(hasta)

	if periodoDesde == "" || periodoHasta == "" {
		minFecha := ""
		maxFecha := ""
		for _, reserva := range reservas {
			fechaInicio := reportesNormalizeDatePart(reserva.FechaEntrada)
			fechaFin := reportesNormalizeDatePart(reportesFirstNonBlank(reserva.FechaSalida, reserva.FechaEntrada))
			if fechaInicio != "" {
				if minFecha == "" || fechaInicio < minFecha {
					minFecha = fechaInicio
				}
			}
			if fechaFin != "" {
				if maxFecha == "" || fechaFin > maxFecha {
					maxFecha = fechaFin
				}
			}
		}
		if periodoDesde == "" {
			periodoDesde = minFecha
		}
		if periodoHasta == "" {
			periodoHasta = maxFecha
		}
	}

	if periodoDesde == "" && periodoHasta != "" {
		periodoDesde = periodoHasta
	}
	if periodoHasta == "" && periodoDesde != "" {
		periodoHasta = periodoDesde
	}
	if periodoDesde != "" && periodoHasta != "" && periodoDesde > periodoHasta {
		periodoDesde, periodoHasta = periodoHasta, periodoDesde
	}

	return periodoDesde, periodoHasta
}

func reportesDaysInclusive(desde, hasta string) float64 {
	desde = strings.TrimSpace(desde)
	hasta = strings.TrimSpace(hasta)
	if desde == "" || hasta == "" {
		return 0
	}
	start, errStart := time.Parse("2006-01-02", desde)
	end, errEnd := time.Parse("2006-01-02", hasta)
	if errStart != nil || errEnd != nil {
		return 0
	}
	if end.Before(start) {
		start, end = end, start
	}
	return end.Sub(start).Hours()/24 + 1
}

func reportesParseDateTime(raw string) (time.Time, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, false
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func reportesReservaOverlapDays(fechaEntrada, fechaSalida, periodoDesde, periodoHasta string) float64 {
	if strings.TrimSpace(periodoDesde) == "" || strings.TrimSpace(periodoHasta) == "" {
		return 0
	}

	periodoStart, errDesde := time.Parse("2006-01-02", periodoDesde)
	periodoHastaDate, errHasta := time.Parse("2006-01-02", periodoHasta)
	if errDesde != nil || errHasta != nil {
		return 0
	}
	periodoEnd := periodoHastaDate.Add(24 * time.Hour)

	entrada, okEntrada := reportesParseDateTime(fechaEntrada)
	salida, okSalida := reportesParseDateTime(fechaSalida)
	if !okEntrada || !okSalida {
		fecha := reportesNormalizeDatePart(reportesFirstNonBlank(fechaEntrada, fechaSalida))
		if fecha == "" {
			return 0
		}
		if !reportesDateWithinRange(fecha, periodoDesde, periodoHasta) {
			return 0
		}
		return 1
	}
	if !salida.After(entrada) {
		salida = entrada.Add(24 * time.Hour)
	}

	if salida.Before(periodoStart) || !entrada.Before(periodoEnd) {
		return 0
	}

	overlapStart := entrada
	if periodoStart.After(overlapStart) {
		overlapStart = periodoStart
	}
	overlapEnd := salida
	if periodoEnd.Before(overlapEnd) {
		overlapEnd = periodoEnd
	}
	if !overlapEnd.After(overlapStart) {
		return 0
	}

	days := overlapEnd.Sub(overlapStart).Hours() / 24
	if days < 0 {
		return 0
	}
	return reportesRound(days)
}

func reportesParseEstacionID(referenciaExterna, codigo string, empresaID int64) int64 {
	referencia := strings.ToUpper(strings.TrimSpace(referenciaExterna))
	if strings.HasPrefix(referencia, "ESTACION_") {
		if parsed, err := strconv.ParseInt(strings.TrimPrefix(referencia, "ESTACION_"), 10, 64); err == nil && parsed > 0 {
			return parsed
		}
	}
	prefix := strings.ToUpper(fmt.Sprintf("EST-%d-", empresaID))
	codigo = strings.ToUpper(strings.TrimSpace(codigo))
	if strings.HasPrefix(codigo, prefix) {
		if parsed, err := strconv.ParseInt(strings.TrimPrefix(codigo, prefix), 10, 64); err == nil && parsed > 0 {
			return parsed
		}
	}
	return 0
}
