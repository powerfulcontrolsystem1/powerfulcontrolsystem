package handlers

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
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
	db              *sql.DB
	empresaID       int64
	desde           string
	hasta           string
	maxRows         int
	includeInactive bool

	tableroCache *dbpkg.EmpresaReportesTableroResumen
	ventasCache  []dbpkg.CarritoCompra
	itemsCache   map[int64][]dbpkg.CarritoCompraItem
}

const (
	reporteDatasetEmpresarialTablero       = "empresarial_tablero"
	reporteDatasetContableEstadoResultados = "contable_estado_resultados"
	reporteDatasetContableBalanceGeneral   = "contable_balance_general"
	reporteDatasetOperativoVentasDetalle   = "operativo_ventas_detalle"
	reporteDatasetOperativoTopProductos    = "operativo_top_productos"
	reporteDatasetOperativoTopClientes     = "operativo_top_clientes"
	reporteDatasetOperativoInventario      = "operativo_inventario_bodega"
	reporteDatasetOperativoCompras         = "operativo_compras_movimientos"
	reporteDatasetContableMovFin           = "contable_movimientos_financieros"
	reporteDatasetContableEventos          = "contable_eventos_contables"
	reporteDatasetContableAsientos         = "contable_asientos_contables"
)

var reportesCatalogo = []empresaReporteCatalogoItem{
	{
		Key:         reporteDatasetEmpresarialTablero,
		Title:       "Tablero Empresarial Ejecutivo",
		Level:       "empresarial",
		Description: "KPI consolidados operativos, financieros y contables por empresa.",
		Formats:     []string{"json", "csv", "txt", "xls"},
	},
	{
		Key:         reporteDatasetContableEstadoResultados,
		Title:       "Estado de Resultados",
		Level:       "contable",
		Description: "Ingresos, gastos y utilidad operacional del rango.",
		Formats:     []string{"json", "csv", "txt", "xls"},
	},
	{
		Key:         reporteDatasetContableBalanceGeneral,
		Title:       "Balance General",
		Level:       "contable",
		Description: "Activos, pasivos, patrimonio y cuadre contable.",
		Formats:     []string{"json", "csv", "txt", "xls"},
	},
	{
		Key:         reporteDatasetOperativoVentasDetalle,
		Title:       "Ventas Cerradas Detalle",
		Level:       "operativo",
		Description: "Detalle de ventas cerradas con cliente, canal e importe.",
		Formats:     []string{"json", "csv", "txt", "xls"},
	},
	{
		Key:         reporteDatasetOperativoTopProductos,
		Title:       "Top Productos Vendidos",
		Level:       "operativo",
		Description: "Ranking de productos por unidades e ingresos vendidos.",
		Formats:     []string{"json", "csv", "txt", "xls"},
	},
	{
		Key:         reporteDatasetOperativoTopClientes,
		Title:       "Top Clientes",
		Level:       "operativo",
		Description: "Ranking de clientes por ventas e ingresos del periodo.",
		Formats:     []string{"json", "csv", "txt", "xls"},
	},
	{
		Key:         reporteDatasetOperativoInventario,
		Title:       "Inventario por Bodega",
		Level:       "operativo",
		Description: "Existencias por bodega y estado de stock.",
		Formats:     []string{"json", "csv", "txt", "xls"},
	},
	{
		Key:         reporteDatasetOperativoCompras,
		Title:       "Compras y Movimientos de Inventario",
		Level:       "operativo",
		Description: "Movimientos de compra (entradas/ajustes positivos) con costos.",
		Formats:     []string{"json", "csv", "txt", "xls"},
	},
	{
		Key:         reporteDatasetContableMovFin,
		Title:       "Movimientos Financieros",
		Level:       "contable",
		Description: "Libro de ingresos/egresos con totales y netos.",
		Formats:     []string{"json", "csv", "txt", "xls"},
	},
	{
		Key:         reporteDatasetContableEventos,
		Title:       "Eventos Contables",
		Level:       "contable",
		Description: "Eventos contables por modulo/evento y estado de procesamiento.",
		Formats:     []string{"json", "csv", "txt", "xls"},
	},
	{
		Key:         reporteDatasetContableAsientos,
		Title:       "Asientos Contables",
		Level:       "contable",
		Description: "Asientos canónicos con débitos/créditos y diferencia.",
		Formats:     []string{"json", "csv", "txt", "xls"},
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
		includeInactive := queryBool(r, "include_inactive")

		builder := &reportesBuilder{
			db:              dbEmp,
			empresaID:       empresaID,
			desde:           desde,
			hasta:           hasta,
			maxRows:         maxRows,
			includeInactive: includeInactive,
			itemsCache:      make(map[int64][]dbpkg.CarritoCompraItem),
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
					http.Error(w, "dataset es obligatorio para formatos tabulares (csv, txt, xls)", http.StatusBadRequest)
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
	case reporteDatasetOperativoVentasDetalle:
		return b.buildOperativoVentasDetalleDataset()
	case reporteDatasetOperativoTopProductos:
		return b.buildOperativoTopProductosDataset()
	case reporteDatasetOperativoTopClientes:
		return b.buildOperativoTopClientesDataset()
	case reporteDatasetOperativoInventario:
		return b.buildOperativoInventarioBodegaDataset()
	case reporteDatasetOperativoCompras:
		return b.buildOperativoComprasMovimientosDataset()
	case reporteDatasetContableMovFin:
		return b.buildContableMovimientosFinancierosDataset()
	case reporteDatasetContableEventos:
		return b.buildContableEventosDataset()
	case reporteDatasetContableAsientos:
		return b.buildContableAsientosDataset()
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
	productoByID := make(map[int64]dbpkg.Producto)
	for _, p := range productos {
		productoByID[p.ID] = p
	}

	type row struct {
		Producto  string
		Bodega    string
		Cantidad  float64
		Minimo    float64
		Maximo    float64
		Estado    string
		Prioridad int
	}
	rows := make([]row, 0, len(existencias))
	for _, ex := range existencias {
		prod := productoByID[ex.ProductoID]
		minimo := prod.StockMinimo
		maximo := prod.StockMaximo
		estado, prioridad := reportesEstadoStock(ex.Cantidad, minimo, maximo)
		rows = append(rows, row{
			Producto:  reportesFirstNonBlank(ex.ProductoNombre, prod.Nombre),
			Bodega:    reportesFirstNonBlank(ex.BodegaNombre, "Bodega #"+strconv.FormatInt(ex.BodegaID, 10)),
			Cantidad:  reportesRound(ex.Cantidad),
			Minimo:    reportesRound(minimo),
			Maximo:    reportesRound(maximo),
			Estado:    estado,
			Prioridad: prioridad,
		})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Prioridad != rows[j].Prioridad {
			return rows[i].Prioridad < rows[j].Prioridad
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
	})
	estadoCount := map[string]int{}
	for _, item := range rows {
		estadoCount[item.Estado]++
		ds.Rows = append(ds.Rows, map[string]interface{}{
			"producto":     item.Producto,
			"bodega":       item.Bodega,
			"existencia":   item.Cantidad,
			"stock_minimo": item.Minimo,
			"stock_maximo": item.Maximo,
			"estado_stock": item.Estado,
		})
	}
	ds.RowCount = len(ds.Rows)
	ds.Summary["sin_stock"] = estadoCount["sin_stock"]
	ds.Summary["bajo_minimo"] = estadoCount["bajo_minimo"]
	ds.Summary["ok"] = estadoCount["ok"]
	ds.Summary["sobre_stock"] = estadoCount["sobre_stock"]
	return ds, nil
}

func (b *reportesBuilder) buildOperativoComprasMovimientosDataset() (empresaReporteDataset, error) {
	limit := b.maxRows * 5
	if limit < 500 {
		limit = 500
	}
	movimientos, err := dbpkg.GetMovimientosByEmpresa(b.db, b.empresaID, 0, 0, "", b.desde, b.hasta, limit, 0)
	if err != nil {
		return empresaReporteDataset{}, err
	}
	rows := make([]dbpkg.InventarioMovimiento, 0)
	for _, mov := range movimientos {
		tipo := strings.ToLower(strings.TrimSpace(mov.Tipo))
		switch tipo {
		case "entrada", "ajuste_entrada", "ajuste_positivo", "compra":
			rows = append(rows, mov)
		}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		return reportesDateUnix(reportesFirstNonBlank(rows[i].FechaMovimiento, rows[i].FechaActualizacion, rows[i].FechaCreacion)) >
			reportesDateUnix(reportesFirstNonBlank(rows[j].FechaMovimiento, rows[j].FechaActualizacion, rows[j].FechaCreacion))
	})
	if len(rows) > b.maxRows {
		rows = rows[:b.maxRows]
	}

	ds := b.newDataset(reporteDatasetOperativoCompras, []string{
		"fecha",
		"producto",
		"bodega_origen",
		"bodega_destino",
		"tipo",
		"cantidad",
		"costo_unitario",
		"costo_total",
		"referencia",
	})
	totalCantidad := 0.0
	totalCosto := 0.0
	for _, mov := range rows {
		costoTotal := mov.Cantidad * mov.CostoUnitario
		totalCantidad += mov.Cantidad
		totalCosto += costoTotal
		ds.Rows = append(ds.Rows, map[string]interface{}{
			"fecha":          reportesFirstNonBlank(mov.FechaMovimiento, mov.FechaActualizacion, mov.FechaCreacion),
			"producto":       reportesFirstNonBlank(mov.ProductoNombre, "Producto #"+strconv.FormatInt(mov.ProductoID, 10)),
			"bodega_origen":  reportesFirstNonBlank(mov.BodegaOrigenNombre, "-"),
			"bodega_destino": reportesFirstNonBlank(mov.BodegaDestinoNombre, "-"),
			"tipo":           strings.ToLower(strings.TrimSpace(mov.Tipo)),
			"cantidad":       reportesRound(mov.Cantidad),
			"costo_unitario": reportesRound(mov.CostoUnitario),
			"costo_total":    reportesRound(costoTotal),
			"referencia":     mov.Referencia,
		})
	}
	ds.RowCount = len(ds.Rows)
	ds.Summary["movimientos"] = ds.RowCount
	ds.Summary["cantidad_total"] = reportesRound(totalCantidad)
	ds.Summary["costo_total"] = reportesRound(totalCosto)
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
	default:
		return fmt.Errorf("format invalido (use json, csv, txt o xls)")
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

func reportesEstadoActivo(estado string) bool {
	e := strings.ToLower(strings.TrimSpace(estado))
	return e == "" || e == "activo"
}

func reportesRound(v float64) float64 {
	return math.Round(v*100) / 100
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
