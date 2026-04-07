package handlers

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type empresaGraficoSerieVentas struct {
	Fecha  string  `json:"fecha"`
	Ventas int64   `json:"ventas"`
	Total  float64 `json:"total"`
}

type empresaGraficoSerieFinanzas struct {
	Fecha    string  `json:"fecha"`
	Ingresos float64 `json:"ingresos"`
	Egresos  float64 `json:"egresos"`
	Balance  float64 `json:"balance"`
}

type empresaGraficoSerieCompras struct {
	Fecha       string  `json:"fecha"`
	Movimientos int64   `json:"movimientos"`
	Costo       float64 `json:"costo"`
}

type empresaGraficoSerieAsistencia struct {
	Fecha     string  `json:"fecha"`
	Registros int64   `json:"registros"`
	Presentes int64   `json:"presentes"`
	Ausentes  int64   `json:"ausentes"`
	Tardanzas int64   `json:"tardanzas"`
	Horas     float64 `json:"horas"`
}

type empresaGraficoRankingProducto struct {
	Producto string  `json:"producto"`
	Cantidad float64 `json:"cantidad"`
	Total    float64 `json:"total"`
}

type empresaGraficoRankingCliente struct {
	Cliente string  `json:"cliente"`
	Ventas  int64   `json:"ventas"`
	Total   float64 `json:"total"`
}

type empresaGraficoDistribucionItem struct {
	Key   string  `json:"key"`
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

type empresaGraficosSeries struct {
	Ventas     []empresaGraficoSerieVentas     `json:"ventas"`
	Finanzas   []empresaGraficoSerieFinanzas   `json:"finanzas"`
	Compras    []empresaGraficoSerieCompras    `json:"compras"`
	Asistencia []empresaGraficoSerieAsistencia `json:"asistencia"`
}

type empresaGraficosRankings struct {
	TopProductos []empresaGraficoRankingProducto `json:"top_productos"`
	TopClientes  []empresaGraficoRankingCliente  `json:"top_clientes"`
}

type empresaGraficosDistribuciones struct {
	StockEstado      []empresaGraficoDistribucionItem `json:"stock_estado"`
	AsistenciaEstado []empresaGraficoDistribucionItem `json:"asistencia_estado"`
}

type empresaGraficosPanelResponse struct {
	EmpresaID      int64                               `json:"empresa_id"`
	Desde          string                              `json:"desde"`
	Hasta          string                              `json:"hasta"`
	GeneradoEn     string                              `json:"generado_en"`
	Tablero        dbpkg.EmpresaReportesTableroResumen `json:"tablero"`
	Series         empresaGraficosSeries               `json:"series"`
	Rankings       empresaGraficosRankings             `json:"rankings"`
	Distribuciones empresaGraficosDistribuciones       `json:"distribuciones"`
	Filtros        empresaGraficosFiltrosAplicados     `json:"filtros"`
	Comparativo    *empresaGraficosComparativo         `json:"comparativo,omitempty"`
	Cache          empresaGraficosCacheEstado          `json:"cache"`
}

type empresaGraficosFiltroCobertura struct {
	Sucursal []string `json:"sucursal,omitempty"`
	Estacion []string `json:"estacion,omitempty"`
	Segmento []string `json:"segmento,omitempty"`
}

type empresaGraficosFiltrosAplicados struct {
	SucursalID int64                          `json:"sucursal_id,omitempty"`
	EstacionID int64                          `json:"estacion_id,omitempty"`
	Segmento   string                         `json:"segmento,omitempty"`
	Cobertura  empresaGraficosFiltroCobertura `json:"cobertura"`
}

type empresaGraficosComparativoMetrica struct {
	Actual       float64 `json:"actual"`
	Anterior     float64 `json:"anterior"`
	Variacion    float64 `json:"variacion"`
	VariacionPct float64 `json:"variacion_pct"`
}

type empresaGraficosComparativo struct {
	Desde           string                                       `json:"desde"`
	Hasta           string                                       `json:"hasta"`
	ReferenciaDesde string                                       `json:"referencia_desde"`
	ReferenciaHasta string                                       `json:"referencia_hasta"`
	Metricas        map[string]empresaGraficosComparativoMetrica `json:"metricas"`
}

type empresaGraficosCacheEstado struct {
	Hit bool `json:"hit"`
}

type empresaGraficosBuildOptions struct {
	SucursalID    int64
	EstacionID    int64
	Segmento      string
	Comparar      bool
	CompararDesde string
	CompararHasta string
}

func (o empresaGraficosBuildOptions) hasFilters() bool {
	return o.SucursalID > 0 || o.EstacionID > 0 || strings.TrimSpace(o.Segmento) != ""
}

type empresaGraficosMetricsSnapshot struct {
	VentasCerradas      int64
	IngresosVentas      float64
	TicketPromedio      float64
	ComprasMovimientos  int64
	ComprasCosto        float64
	MovimientosIngresos int64
	MovimientosEgresos  int64
	Ingresos            float64
	Egresos             float64
	Balance             float64
	AsistenciaRegistros int64
}

type empresaGraficosPanelCacheEntry struct {
	expiresAt time.Time
	panel     empresaGraficosPanelResponse
}

type empresaGraficosPanelCache struct {
	mu         sync.RWMutex
	entries    map[string]empresaGraficosPanelCacheEntry
	ttl        time.Duration
	maxEntries int
}

func newEmpresaGraficosPanelCache(ttl time.Duration, maxEntries int) *empresaGraficosPanelCache {
	if ttl <= 0 {
		ttl = 90 * time.Second
	}
	if maxEntries <= 0 {
		maxEntries = 128
	}
	return &empresaGraficosPanelCache{
		entries:    make(map[string]empresaGraficosPanelCacheEntry),
		ttl:        ttl,
		maxEntries: maxEntries,
	}
}

func (c *empresaGraficosPanelCache) Get(key string) (empresaGraficosPanelResponse, bool) {
	if c == nil {
		return empresaGraficosPanelResponse{}, false
	}
	now := time.Now()
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok {
		return empresaGraficosPanelResponse{}, false
	}
	if now.After(entry.expiresAt) {
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()
		return empresaGraficosPanelResponse{}, false
	}
	return entry.panel, true
}

func (c *empresaGraficosPanelCache) Set(key string, panel empresaGraficosPanelResponse) {
	if c == nil || strings.TrimSpace(key) == "" {
		return
	}
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, entry := range c.entries {
		if now.After(entry.expiresAt) {
			delete(c.entries, k)
		}
	}

	if len(c.entries) >= c.maxEntries {
		var oldestKey string
		var oldest time.Time
		for k, entry := range c.entries {
			if oldestKey == "" || entry.expiresAt.Before(oldest) {
				oldestKey = k
				oldest = entry.expiresAt
			}
		}
		if oldestKey != "" {
			delete(c.entries, oldestKey)
		}
	}

	c.entries[key] = empresaGraficosPanelCacheEntry{
		expiresAt: now.Add(c.ttl),
		panel:     panel,
	}
}

// EmpresaGraficosEstadisticasHandler expone datasets listos para visualizacion grafica por empresa.
func EmpresaGraficosEstadisticasHandler(dbEmp *sql.DB) http.HandlerFunc {
	panelCache := newEmpresaGraficosPanelCache(90*time.Second, 128)
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
			action = "panel"
		}

		desde := strings.TrimSpace(r.URL.Query().Get("desde"))
		hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))

		maxPoints, err := parseIntQueryOptional(r, "max_points")
		if err != nil {
			http.Error(w, "max_points invalido", http.StatusBadRequest)
			return
		}
		if maxPoints <= 0 {
			maxPoints = 45
		}
		if maxPoints > 365 {
			maxPoints = 365
		}

		topN, err := parseIntQueryOptional(r, "top")
		if err != nil {
			http.Error(w, "top invalido", http.StatusBadRequest)
			return
		}
		if topN <= 0 {
			topN = 10
		}
		if topN > 50 {
			topN = 50
		}

		sucursalID, err := graficosParseInt64QueryOptional(r, "sucursal_id")
		if err != nil {
			http.Error(w, "sucursal_id invalido", http.StatusBadRequest)
			return
		}
		estacionID, err := graficosParseInt64QueryOptional(r, "estacion_id")
		if err != nil {
			http.Error(w, "estacion_id invalido", http.StatusBadRequest)
			return
		}
		segmento := graficosNormalizeSegmento(r.URL.Query().Get("segmento"))

		comparar := queryBool(r, "comparar") || queryBool(r, "comparativo")
		compararDesde := reportesNormalizeDatePart(strings.TrimSpace(reportesFirstNonBlank(
			r.URL.Query().Get("comparar_desde"),
			r.URL.Query().Get("comparativo_desde"),
			r.URL.Query().Get("compare_desde"),
		)))
		compararHasta := reportesNormalizeDatePart(strings.TrimSpace(reportesFirstNonBlank(
			r.URL.Query().Get("comparar_hasta"),
			r.URL.Query().Get("comparativo_hasta"),
			r.URL.Query().Get("compare_hasta"),
		)))
		if (compararDesde != "" && compararHasta == "") || (compararDesde == "" && compararHasta != "") {
			http.Error(w, "comparar_desde y comparar_hasta deben enviarse juntos", http.StatusBadRequest)
			return
		}
		if comparar && (reportesNormalizeDatePart(desde) == "" || reportesNormalizeDatePart(hasta) == "") {
			http.Error(w, "para comparar se requieren desde y hasta en formato YYYY-MM-DD", http.StatusBadRequest)
			return
		}

		buildOptions := empresaGraficosBuildOptions{
			SucursalID:    sucursalID,
			EstacionID:    estacionID,
			Segmento:      segmento,
			Comparar:      comparar,
			CompararDesde: compararDesde,
			CompararHasta: compararHasta,
		}

		includeInactive := queryBool(r, "include_inactive")
		builder := &reportesBuilder{
			db:              dbEmp,
			empresaID:       empresaID,
			desde:           desde,
			hasta:           hasta,
			maxRows:         2000,
			includeInactive: includeInactive,
			itemsCache:      make(map[int64][]dbpkg.CarritoCompraItem),
		}

		skipCache := queryBool(r, "skip_cache") || queryBool(r, "refresh")
		cacheKey := graficosBuildCacheKey(builder, maxPoints, topN, buildOptions)
		if !skipCache {
			if cachedPanel, ok := panelCache.Get(cacheKey); ok {
				cachedPanel.Cache.Hit = true
				serveGraficosAction(w, action, cachedPanel, r)
				return
			}
		}

		panel, err := buildEmpresaGraficosPanel(dbEmp, builder, maxPoints, topN, buildOptions)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		panel.Cache.Hit = false

		if !skipCache {
			panelCache.Set(cacheKey, panel)
		}

		serveGraficosAction(w, action, panel, r)
	}
}

func serveGraficosAction(w http.ResponseWriter, action string, panel empresaGraficosPanelResponse, r *http.Request) {
	switch action {
	case "panel", "dashboard", "tablero":
		writeJSON(w, http.StatusOK, panel)
		return

	case "serie", "series":
		serie := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("serie")))
		if serie == "" {
			http.Error(w, "serie es obligatoria (ventas, finanzas, compras, asistencia)", http.StatusBadRequest)
			return
		}
		var data interface{}
		switch serie {
		case "ventas":
			data = panel.Series.Ventas
		case "finanzas":
			data = panel.Series.Finanzas
		case "compras":
			data = panel.Series.Compras
		case "asistencia":
			data = panel.Series.Asistencia
		default:
			http.Error(w, "serie invalida (use ventas, finanzas, compras o asistencia)", http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"empresa_id":  panel.EmpresaID,
			"desde":       panel.Desde,
			"hasta":       panel.Hasta,
			"generado_en": panel.GeneradoEn,
			"serie":       serie,
			"data":        data,
			"filtros":     panel.Filtros,
			"comparativo": panel.Comparativo,
			"cache":       panel.Cache,
		})
		return

	case "rankings", "ranking":
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"empresa_id":  panel.EmpresaID,
			"desde":       panel.Desde,
			"hasta":       panel.Hasta,
			"generado_en": panel.GeneradoEn,
			"rankings":    panel.Rankings,
			"filtros":     panel.Filtros,
			"comparativo": panel.Comparativo,
			"cache":       panel.Cache,
		})
		return

	case "distribuciones", "distributions", "distribucion":
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"empresa_id":     panel.EmpresaID,
			"desde":          panel.Desde,
			"hasta":          panel.Hasta,
			"generado_en":    panel.GeneradoEn,
			"distribuciones": panel.Distribuciones,
			"filtros":        panel.Filtros,
			"comparativo":    panel.Comparativo,
			"cache":          panel.Cache,
		})
		return

	case "catalogo", "catalog":
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"empresa_id": panel.EmpresaID,
			"actions": []map[string]string{
				{"action": "panel", "description": "tablero consolidado con series, rankings, distribuciones, filtros y comparativos"},
				{"action": "serie", "description": "serie puntual (ventas, finanzas, compras, asistencia)"},
				{"action": "rankings", "description": "top productos y top clientes"},
				{"action": "distribuciones", "description": "distribuciones de stock y asistencia"},
			},
			"series":  []string{"ventas", "finanzas", "compras", "asistencia"},
			"filters": []string{"sucursal_id", "estacion_id", "segmento"},
			"flags":   []string{"comparar", "skip_cache", "refresh"},
		})
		return

	default:
		http.Error(w, "action invalida (use panel, serie, rankings, distribuciones o catalogo)", http.StatusBadRequest)
		return
	}
}

func buildEmpresaGraficosPanel(dbEmp *sql.DB, builder *reportesBuilder, maxPoints, topN int, options empresaGraficosBuildOptions) (empresaGraficosPanelResponse, error) {
	tablero, err := builder.getTableroResumen()
	if err != nil {
		return empresaGraficosPanelResponse{}, err
	}

	panel := empresaGraficosPanelResponse{
		EmpresaID:      builder.empresaID,
		Desde:          builder.desde,
		Hasta:          builder.hasta,
		GeneradoEn:     time.Now().Format("2006-01-02 15:04:05"),
		Tablero:        *tablero,
		Series:         empresaGraficosSeries{},
		Rankings:       empresaGraficosRankings{},
		Distribuciones: empresaGraficosDistribuciones{},
		Filtros:        empresaGraficosFiltrosAplicados{},
		Cache:          empresaGraficosCacheEstado{Hit: false},
	}

	ventas, err := builder.getVentasCerradasFiltradas()
	if err != nil {
		return empresaGraficosPanelResponse{}, err
	}
	if err := builder.ensureItemsForCarritos(ventas); err != nil {
		return empresaGraficosPanelResponse{}, err
	}

	finanzasDataset, err := builder.buildDataset(reporteDatasetContableMovFin)
	if err != nil {
		return empresaGraficosPanelResponse{}, err
	}
	comprasDataset, err := builder.buildDataset(reporteDatasetOperativoCompras)
	if err != nil {
		return empresaGraficosPanelResponse{}, err
	}
	inventarioDataset, err := builder.buildDataset(reporteDatasetOperativoInventario)
	if err != nil {
		return empresaGraficosPanelResponse{}, err
	}

	asistencias, err := dbpkg.ListEmpresaAsistenciaEmpleados(dbEmp, builder.empresaID, builder.includeInactive, builder.desde, builder.hasta, "", "", 2000)
	if err != nil {
		return empresaGraficosPanelResponse{}, err
	}

	filterCtx, err := newEmpresaGraficosFilterContext(dbEmp, builder, options, ventas)
	if err != nil {
		return empresaGraficosPanelResponse{}, err
	}

	filteredVentas, err := filterCtx.filterVentas(ventas)
	if err != nil {
		return empresaGraficosPanelResponse{}, err
	}
	filteredFinanzas := filterCtx.filterRowsByDate(finanzasDataset.Rows, "fecha_movimiento", "fecha")
	filteredCompras := filterCtx.filterRowsByDate(comprasDataset.Rows, "fecha", "fecha_documento", "ultima_fecha_documento", "fecha_movimiento")
	filteredAsistencias := filterCtx.filterAsistencias(asistencias)

	panel.Series.Ventas = buildGraficoVentasSerieDesdeCarritos(filteredVentas, maxPoints)
	panel.Series.Finanzas = buildGraficoFinanzasSerie(filteredFinanzas, maxPoints)
	panel.Series.Compras = buildGraficoComprasSerie(filteredCompras, maxPoints)
	if len(panel.Series.Compras) == 0 {
		panel.Series.Compras = buildGraficoComprasSerieDesdeFinanzas(filteredFinanzas, maxPoints)
	}
	panel.Series.Asistencia = buildGraficoAsistenciaSerie(filteredAsistencias, maxPoints)

	panel.Rankings.TopProductos = buildGraficoTopProductosDesdeVentas(filteredVentas, builder.itemsCache, topN)
	panel.Rankings.TopClientes = buildGraficoTopClientesDesdeVentas(filteredVentas, topN)

	panel.Distribuciones.StockEstado = buildGraficoStockDistribucion(inventarioDataset.Rows)
	panel.Distribuciones.AsistenciaEstado = buildGraficoAsistenciaDistribucion(filteredAsistencias)

	currentSnapshot := graficosBuildMetricsSnapshot(filteredVentas, filteredFinanzas, filteredCompras, filteredAsistencias)
	if options.hasFilters() {
		graficosApplySnapshotToTablero(&panel.Tablero, currentSnapshot)
	}

	panel.Filtros = filterCtx.appliedFilters()

	if options.Comparar {
		referenciaDesde, referenciaHasta, err := graficosResolveComparativoRange(builder.desde, builder.hasta, options.CompararDesde, options.CompararHasta)
		if err != nil {
			return empresaGraficosPanelResponse{}, err
		}
		referenceBuilder := &reportesBuilder{
			db:              dbEmp,
			empresaID:       builder.empresaID,
			desde:           referenciaDesde,
			hasta:           referenciaHasta,
			maxRows:         builder.maxRows,
			includeInactive: builder.includeInactive,
			itemsCache:      make(map[int64][]dbpkg.CarritoCompraItem),
		}
		referenceSnapshot, err := graficosBuildMetricsSnapshotForRange(dbEmp, referenceBuilder, options)
		if err != nil {
			return empresaGraficosPanelResponse{}, err
		}
		panel.Comparativo = graficosBuildComparativo(
			currentSnapshot,
			referenceSnapshot,
			reportesNormalizeDatePart(builder.desde),
			reportesNormalizeDatePart(builder.hasta),
			referenciaDesde,
			referenciaHasta,
		)
	}

	return panel, nil
}

type empresaGraficosFilterContext struct {
	builder          *reportesBuilder
	options          empresaGraficosBuildOptions
	sucursalDateSet  map[string]struct{}
	carritoEstacion  map[int64]int64
	clienteSegmentos map[int64]string
	cobertura        empresaGraficosFiltroCobertura
}

func newEmpresaGraficosFilterContext(dbEmp *sql.DB, builder *reportesBuilder, options empresaGraficosBuildOptions, ventas []dbpkg.CarritoCompra) (*empresaGraficosFilterContext, error) {
	ctx := &empresaGraficosFilterContext{
		builder:          builder,
		options:          options,
		sucursalDateSet:  make(map[string]struct{}),
		carritoEstacion:  make(map[int64]int64),
		clienteSegmentos: make(map[int64]string),
		cobertura:        empresaGraficosFiltroCobertura{},
	}

	if options.SucursalID > 0 {
		dateSet, err := graficosLoadSucursalDateSet(dbEmp, builder, options.SucursalID)
		if err != nil {
			return nil, err
		}
		ctx.sucursalDateSet = dateSet
		ctx.cobertura.Sucursal = []string{"ventas", "finanzas", "compras", "asistencia", "rankings"}
	}

	if options.EstacionID > 0 {
		idx, err := graficosLoadCarritoEstacionIndex(dbEmp, builder.empresaID)
		if err != nil {
			return nil, err
		}
		ctx.carritoEstacion = idx
		ctx.cobertura.Estacion = []string{"ventas", "rankings"}
	}

	if options.Segmento != "" {
		segments, err := graficosBuildClienteSegmentMap(dbEmp, builder.empresaID, ventas)
		if err != nil {
			return nil, err
		}
		ctx.clienteSegmentos = segments
		ctx.cobertura.Segmento = []string{"ventas", "rankings"}
	}

	return ctx, nil
}

func (c *empresaGraficosFilterContext) appliedFilters() empresaGraficosFiltrosAplicados {
	if c == nil {
		return empresaGraficosFiltrosAplicados{}
	}
	return empresaGraficosFiltrosAplicados{
		SucursalID: c.options.SucursalID,
		EstacionID: c.options.EstacionID,
		Segmento:   c.options.Segmento,
		Cobertura:  c.cobertura,
	}
}

func (c *empresaGraficosFilterContext) filterVentas(ventas []dbpkg.CarritoCompra) ([]dbpkg.CarritoCompra, error) {
	if c == nil {
		return ventas, nil
	}
	out := make([]dbpkg.CarritoCompra, 0, len(ventas))
	for _, venta := range ventas {
		if c.options.SucursalID > 0 {
			fechaVenta := reportesNormalizeDatePart(reportesFirstNonBlank(venta.PagadoEn, venta.FechaActualizacion, venta.FechaCreacion))
			if !graficosDateAllowed(fechaVenta, c.sucursalDateSet) {
				continue
			}
		}
		if c.options.EstacionID > 0 {
			estacionID := c.resolveVentaEstacionID(venta)
			if estacionID != c.options.EstacionID {
				continue
			}
		}
		if c.options.Segmento != "" {
			segmentoVenta := c.resolveVentaSegmento(venta)
			if segmentoVenta != c.options.Segmento {
				continue
			}
		}
		out = append(out, venta)
	}
	return out, nil
}

func (c *empresaGraficosFilterContext) filterRowsByDate(rows []map[string]interface{}, dateFields ...string) []map[string]interface{} {
	if c == nil || c.options.SucursalID <= 0 {
		return rows
	}
	if len(c.sucursalDateSet) == 0 {
		return []map[string]interface{}{}
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		dateValue := ""
		for _, field := range dateFields {
			dateValue = reportesNormalizeDatePart(graficoString(row[field]))
			if dateValue != "" {
				break
			}
		}
		if graficosDateAllowed(dateValue, c.sucursalDateSet) {
			out = append(out, row)
		}
	}
	return out
}

func (c *empresaGraficosFilterContext) filterAsistencias(items []dbpkg.EmpresaAsistenciaEmpleado) []dbpkg.EmpresaAsistenciaEmpleado {
	if c == nil || c.options.SucursalID <= 0 {
		return items
	}
	if len(c.sucursalDateSet) == 0 {
		return []dbpkg.EmpresaAsistenciaEmpleado{}
	}
	out := make([]dbpkg.EmpresaAsistenciaEmpleado, 0, len(items))
	for _, item := range items {
		fecha := reportesNormalizeDatePart(item.FechaAsistencia)
		if graficosDateAllowed(fecha, c.sucursalDateSet) {
			out = append(out, item)
		}
	}
	return out
}

func (c *empresaGraficosFilterContext) resolveVentaEstacionID(venta dbpkg.CarritoCompra) int64 {
	if c == nil {
		return 0
	}
	if id, ok := c.carritoEstacion[venta.ID]; ok && id > 0 {
		return id
	}
	return graficosInferEstacionID(venta.Codigo, venta.Nombre, venta.ReferenciaExterna)
}

func (c *empresaGraficosFilterContext) resolveVentaSegmento(venta dbpkg.CarritoCompra) string {
	if c == nil || venta.ClienteID <= 0 {
		return ""
	}
	if segmento, ok := c.clienteSegmentos[venta.ClienteID]; ok {
		return segmento
	}
	return ""
}

func graficosLoadSucursalDateSet(dbEmp *sql.DB, builder *reportesBuilder, sucursalID int64) (map[string]struct{}, error) {
	set := make(map[string]struct{})
	if sucursalID <= 0 {
		return set, nil
	}

	cierres, err := dbpkg.ListEmpresaCierresCaja(dbEmp, builder.empresaID, dbpkg.EmpresaCierreCajaFilter{
		SucursalID:      sucursalID,
		Desde:           builder.desde,
		Hasta:           builder.hasta,
		IncludeInactive: builder.includeInactive,
		Limit:           2000,
	})
	if err != nil {
		if graficosIsNoSuchTableErr(err) {
			return set, nil
		}
		return nil, err
	}
	for _, cierre := range cierres {
		fecha := reportesNormalizeDatePart(cierre.FechaOperacion)
		if fecha == "" {
			fecha = reportesNormalizeDatePart(reportesFirstNonBlank(cierre.FechaApertura, cierre.FechaCierre))
		}
		if fecha != "" {
			set[fecha] = struct{}{}
		}
	}
	return set, nil
}

func graficosLoadCarritoEstacionIndex(dbEmp *sql.DB, empresaID int64) (map[int64]int64, error) {
	out := make(map[int64]int64)
	rows, err := dbEmp.Query(`SELECT
		COALESCE(carrito_id, 0),
		COALESCE(estacion_id, 0),
		COALESCE(estacion_codigo, ''),
		COALESCE(estacion_nombre, ''),
		COALESCE(referencia_operacion, '')
	FROM empresa_ventas_estacion_metricas
	WHERE empresa_id = ?
		AND LOWER(COALESCE(estado, 'activo')) = 'activo'
	ORDER BY id DESC`, empresaID)
	if err != nil {
		if graficosIsNoSuchTableErr(err) {
			return out, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var carritoID int64
		var estacionID int64
		var estacionCodigo string
		var estacionNombre string
		var referencia string
		if err := rows.Scan(&carritoID, &estacionID, &estacionCodigo, &estacionNombre, &referencia); err != nil {
			return nil, err
		}
		if carritoID <= 0 {
			continue
		}
		if _, exists := out[carritoID]; exists {
			continue
		}
		if estacionID <= 0 {
			estacionID = graficosInferEstacionID(estacionCodigo, estacionNombre, referencia)
		}
		if estacionID > 0 {
			out[carritoID] = estacionID
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func graficosBuildClienteSegmentMap(dbEmp *sql.DB, empresaID int64, ventas []dbpkg.CarritoCompra) (map[int64]string, error) {
	ids := make(map[int64]struct{})
	for _, venta := range ventas {
		if venta.ClienteID > 0 {
			ids[venta.ClienteID] = struct{}{}
		}
	}

	out := make(map[int64]string, len(ids))
	for clienteID := range ids {
		perfil, err := dbpkg.GetClientePerfilComercialByEmpresa(dbEmp, empresaID, clienteID)
		if err != nil {
			if err == sql.ErrNoRows || graficosIsNoSuchTableErr(err) {
				out[clienteID] = "nuevo"
				continue
			}
			return nil, err
		}
		segmento := graficosNormalizeSegmento(perfil.Segmento)
		if segmento == "" {
			segmento = "nuevo"
		}
		out[clienteID] = segmento
	}
	return out, nil
}

func graficosBuildCacheKey(builder *reportesBuilder, maxPoints, topN int, options empresaGraficosBuildOptions) string {
	return fmt.Sprintf(
		"e:%d|d:%s|h:%s|mp:%d|top:%d|ia:%t|suc:%d|est:%d|seg:%s|cmp:%t|cd:%s|ch:%s",
		builder.empresaID,
		reportesNormalizeDatePart(builder.desde),
		reportesNormalizeDatePart(builder.hasta),
		maxPoints,
		topN,
		builder.includeInactive,
		options.SucursalID,
		options.EstacionID,
		options.Segmento,
		options.Comparar,
		reportesNormalizeDatePart(options.CompararDesde),
		reportesNormalizeDatePart(options.CompararHasta),
	)
}

func graficosNormalizeSegmento(raw string) string {
	segmento := strings.ToLower(strings.TrimSpace(raw))
	segmento = strings.ReplaceAll(segmento, " ", "_")
	return segmento
}

func graficosParseInt64QueryOptional(r *http.Request, key string) (int64, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return 0, nil
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, err
	}
	if v < 0 {
		return 0, fmt.Errorf("%s invalido", key)
	}
	return v, nil
}

func graficosDateAllowed(dateValue string, set map[string]struct{}) bool {
	if len(set) == 0 {
		return false
	}
	_, ok := set[dateValue]
	return ok
}

func graficosInferEstacionID(values ...string) int64 {
	for _, value := range values {
		if id := graficosExtractEstacionID(value); id > 0 {
			return id
		}
	}
	return 0
}

func graficosExtractEstacionID(raw string) int64 {
	value := strings.ToUpper(strings.TrimSpace(raw))
	if value == "" {
		return 0
	}

	markers := []string{"ESTACION_", "ESTACION-", "EST_", "EST-", "MESA_", "MESA-", "HABITACION_", "HABITACION-", "ROOM_", "ROOM-"}
	for _, marker := range markers {
		idx := strings.LastIndex(value, marker)
		if idx < 0 {
			continue
		}
		if id := graficosParseLeadingInt(value[idx+len(marker):]); id > 0 {
			return id
		}
	}

	return graficosParseLeadingInt(value)
}

func graficosParseLeadingInt(raw string) int64 {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		return 0
	}
	digits := strings.Builder{}
	started := false
	for _, r := range clean {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
			started = true
			continue
		}
		if started {
			break
		}
	}
	if digits.Len() == 0 {
		return 0
	}
	n, err := strconv.ParseInt(digits.String(), 10, 64)
	if err != nil {
		return 0
	}
	return n
}

func graficosIsNoSuchTableErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "no such table")
}

func graficosResolveComparativoRange(actualDesde, actualHasta, manualDesde, manualHasta string) (string, string, error) {
	baseDesde := reportesNormalizeDatePart(actualDesde)
	baseHasta := reportesNormalizeDatePart(actualHasta)
	refDesde := reportesNormalizeDatePart(manualDesde)
	refHasta := reportesNormalizeDatePart(manualHasta)

	if refDesde != "" || refHasta != "" {
		if refDesde == "" || refHasta == "" {
			return "", "", fmt.Errorf("comparar_desde y comparar_hasta deben enviarse juntos")
		}
		if refDesde > refHasta {
			return "", "", fmt.Errorf("comparar_desde no puede ser mayor que comparar_hasta")
		}
		return refDesde, refHasta, nil
	}

	if baseDesde == "" || baseHasta == "" {
		return "", "", fmt.Errorf("no se pudo resolver el rango comparativo")
	}
	if baseDesde > baseHasta {
		return "", "", fmt.Errorf("desde no puede ser mayor que hasta")
	}

	start, err := time.ParseInLocation("2006-01-02", baseDesde, time.Local)
	if err != nil {
		return "", "", fmt.Errorf("desde invalido para comparativo")
	}
	end, err := time.ParseInLocation("2006-01-02", baseHasta, time.Local)
	if err != nil {
		return "", "", fmt.Errorf("hasta invalido para comparativo")
	}
	days := int(end.Sub(start).Hours()/24) + 1
	if days <= 0 {
		days = 1
	}

	prevEnd := start.AddDate(0, 0, -1)
	prevStart := prevEnd.AddDate(0, 0, -(days - 1))
	return prevStart.Format("2006-01-02"), prevEnd.Format("2006-01-02"), nil
}

func graficosBuildComparativo(actual, anterior empresaGraficosMetricsSnapshot, desde, hasta, referenciaDesde, referenciaHasta string) *empresaGraficosComparativo {
	return &empresaGraficosComparativo{
		Desde:           desde,
		Hasta:           hasta,
		ReferenciaDesde: referenciaDesde,
		ReferenciaHasta: referenciaHasta,
		Metricas: map[string]empresaGraficosComparativoMetrica{
			"ventas_cerradas":      graficosBuildComparativoMetrica(float64(actual.VentasCerradas), float64(anterior.VentasCerradas)),
			"ingresos_ventas":      graficosBuildComparativoMetrica(actual.IngresosVentas, anterior.IngresosVentas),
			"ticket_promedio":      graficosBuildComparativoMetrica(actual.TicketPromedio, anterior.TicketPromedio),
			"compras_movimientos":  graficosBuildComparativoMetrica(float64(actual.ComprasMovimientos), float64(anterior.ComprasMovimientos)),
			"compras_costo":        graficosBuildComparativoMetrica(actual.ComprasCosto, anterior.ComprasCosto),
			"ingresos_financieros": graficosBuildComparativoMetrica(actual.Ingresos, anterior.Ingresos),
			"egresos_financieros":  graficosBuildComparativoMetrica(actual.Egresos, anterior.Egresos),
			"balance_financiero":   graficosBuildComparativoMetrica(actual.Balance, anterior.Balance),
			"asistencia_registros": graficosBuildComparativoMetrica(float64(actual.AsistenciaRegistros), float64(anterior.AsistenciaRegistros)),
		},
	}
}

func graficosBuildComparativoMetrica(actual, anterior float64) empresaGraficosComparativoMetrica {
	variacion := reportesRound(actual - anterior)
	variacionPct := 0.0
	if anterior == 0 {
		if actual != 0 {
			variacionPct = 100
		}
	} else {
		variacionPct = reportesRound((variacion / math.Abs(anterior)) * 100)
	}
	return empresaGraficosComparativoMetrica{
		Actual:       reportesRound(actual),
		Anterior:     reportesRound(anterior),
		Variacion:    variacion,
		VariacionPct: variacionPct,
	}
}

func graficosBuildMetricsSnapshot(ventas []dbpkg.CarritoCompra, finanzasRows, comprasRows []map[string]interface{}, asistencias []dbpkg.EmpresaAsistenciaEmpleado) empresaGraficosMetricsSnapshot {
	snapshot := empresaGraficosMetricsSnapshot{}

	snapshot.VentasCerradas = int64(len(ventas))
	for _, venta := range ventas {
		snapshot.IngresosVentas = reportesRound(snapshot.IngresosVentas + reportesVentaTotal(venta))
	}
	if snapshot.VentasCerradas > 0 {
		snapshot.TicketPromedio = reportesRound(snapshot.IngresosVentas / float64(snapshot.VentasCerradas))
	}

	for _, row := range finanzasRows {
		tipo := strings.ToLower(strings.TrimSpace(graficoString(row["tipo_movimiento"])))
		monto := graficoFloat(row["total_neto"])
		if monto == 0 {
			monto = graficoFloat(row["monto"])
		}
		switch tipo {
		case "ingreso":
			snapshot.MovimientosIngresos++
			snapshot.Ingresos = reportesRound(snapshot.Ingresos + monto)
		case "egreso":
			snapshot.MovimientosEgresos++
			snapshot.Egresos = reportesRound(snapshot.Egresos + monto)
		}
	}
	snapshot.Balance = reportesRound(snapshot.Ingresos - snapshot.Egresos)

	for _, row := range comprasRows {
		movimientos := int64(graficoFloat(row["movimientos"]))
		if movimientos <= 0 {
			movimientos = int64(graficoFloat(row["documentos"]))
		}
		if movimientos <= 0 {
			movimientos = int64(graficoFloat(row["ordenes_emitidas"]))
		}
		if movimientos <= 0 {
			movimientos = 1
		}
		snapshot.ComprasMovimientos += movimientos

		costo := graficoFloat(row["costo_total"])
		if costo == 0 {
			costo = graficoFloat(row["monto_ordenado"])
		}
		if costo == 0 {
			costo = graficoFloat(row["monto_recepcionado"])
		}
		if costo == 0 {
			costo = graficoFloat(row["monto_contabilizado"])
		}
		if costo == 0 {
			costo = graficoFloat(row["total_neto"])
		}
		if costo == 0 {
			costo = graficoFloat(row["monto"])
		}
		snapshot.ComprasCosto = reportesRound(snapshot.ComprasCosto + costo)
	}

	snapshot.AsistenciaRegistros = int64(len(asistencias))
	return snapshot
}

func graficosApplySnapshotToTablero(tablero *dbpkg.EmpresaReportesTableroResumen, snapshot empresaGraficosMetricsSnapshot) {
	if tablero == nil {
		return
	}
	tablero.Operativo.VentasCerradas = snapshot.VentasCerradas
	tablero.Operativo.IngresosVentas = reportesRound(snapshot.IngresosVentas)
	tablero.Operativo.TicketPromedio = reportesRound(snapshot.TicketPromedio)
	tablero.Operativo.ComprasMovimientos = snapshot.ComprasMovimientos
	tablero.Operativo.ComprasCosto = reportesRound(snapshot.ComprasCosto)

	tablero.Financiero.MovimientosIngresos = snapshot.MovimientosIngresos
	tablero.Financiero.MovimientosEgresos = snapshot.MovimientosEgresos
	tablero.Financiero.Ingresos = reportesRound(snapshot.Ingresos)
	tablero.Financiero.Egresos = reportesRound(snapshot.Egresos)
	tablero.Financiero.Balance = reportesRound(snapshot.Balance)
}

func graficosBuildMetricsSnapshotForRange(dbEmp *sql.DB, builder *reportesBuilder, options empresaGraficosBuildOptions) (empresaGraficosMetricsSnapshot, error) {
	ventas, err := builder.getVentasCerradasFiltradas()
	if err != nil {
		return empresaGraficosMetricsSnapshot{}, err
	}
	if err := builder.ensureItemsForCarritos(ventas); err != nil {
		return empresaGraficosMetricsSnapshot{}, err
	}

	finanzasDataset, err := builder.buildDataset(reporteDatasetContableMovFin)
	if err != nil {
		return empresaGraficosMetricsSnapshot{}, err
	}
	comprasDataset, err := builder.buildDataset(reporteDatasetOperativoCompras)
	if err != nil {
		return empresaGraficosMetricsSnapshot{}, err
	}
	asistencias, err := dbpkg.ListEmpresaAsistenciaEmpleados(dbEmp, builder.empresaID, builder.includeInactive, builder.desde, builder.hasta, "", "", 2000)
	if err != nil {
		return empresaGraficosMetricsSnapshot{}, err
	}

	ctx, err := newEmpresaGraficosFilterContext(dbEmp, builder, options, ventas)
	if err != nil {
		return empresaGraficosMetricsSnapshot{}, err
	}
	filteredVentas, err := ctx.filterVentas(ventas)
	if err != nil {
		return empresaGraficosMetricsSnapshot{}, err
	}
	filteredFinanzas := ctx.filterRowsByDate(finanzasDataset.Rows, "fecha_movimiento", "fecha")
	filteredCompras := ctx.filterRowsByDate(comprasDataset.Rows, "fecha", "fecha_documento", "ultima_fecha_documento", "fecha_movimiento")
	filteredAsistencias := ctx.filterAsistencias(asistencias)

	return graficosBuildMetricsSnapshot(filteredVentas, filteredFinanzas, filteredCompras, filteredAsistencias), nil
}

func buildGraficoVentasSerieDesdeCarritos(ventas []dbpkg.CarritoCompra, maxPoints int) []empresaGraficoSerieVentas {
	byDate := make(map[string]*empresaGraficoSerieVentas)
	for _, venta := range ventas {
		fecha := reportesNormalizeDatePart(reportesFirstNonBlank(venta.PagadoEn, venta.FechaActualizacion, venta.FechaCreacion))
		if fecha == "" {
			continue
		}
		entry, ok := byDate[fecha]
		if !ok {
			entry = &empresaGraficoSerieVentas{Fecha: fecha}
			byDate[fecha] = entry
		}
		entry.Ventas++
		entry.Total = reportesRound(entry.Total + reportesVentaTotal(venta))
	}

	dates := make([]string, 0, len(byDate))
	for date := range byDate {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	series := make([]empresaGraficoSerieVentas, 0, len(dates))
	for _, date := range dates {
		item := byDate[date]
		series = append(series, empresaGraficoSerieVentas{
			Fecha:  item.Fecha,
			Ventas: item.Ventas,
			Total:  reportesRound(item.Total),
		})
	}
	return compactGraficoVentasSerie(series, maxPoints)
}

func buildGraficoTopProductosDesdeVentas(ventas []dbpkg.CarritoCompra, itemsCache map[int64][]dbpkg.CarritoCompraItem, topN int) []empresaGraficoRankingProducto {
	type agg struct {
		producto string
		cantidad float64
		total    float64
	}
	aggByProduct := make(map[string]*agg)
	for _, venta := range ventas {
		items := itemsCache[venta.ID]
		for _, item := range items {
			if strings.EqualFold(strings.TrimSpace(item.Estado), "inactivo") {
				continue
			}
			key := strings.TrimSpace(item.CodigoItem)
			if key == "" && item.ReferenciaID > 0 {
				key = "producto_" + strconv.FormatInt(item.ReferenciaID, 10)
			}
			if key == "" {
				key = "item_" + strconv.FormatInt(item.ID, 10)
			}
			entry, ok := aggByProduct[key]
			if !ok {
				entry = &agg{producto: reportesFirstNonBlank(item.Descripcion, item.CodigoItem, key)}
				aggByProduct[key] = entry
			}
			entry.cantidad += item.Cantidad
			entry.total += item.TotalLinea
		}
	}

	out := make([]empresaGraficoRankingProducto, 0, len(aggByProduct))
	for _, item := range aggByProduct {
		out = append(out, empresaGraficoRankingProducto{
			Producto: item.producto,
			Cantidad: reportesRound(item.cantidad),
			Total:    reportesRound(item.total),
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Total == out[j].Total {
			return out[i].Cantidad > out[j].Cantidad
		}
		return out[i].Total > out[j].Total
	})
	if len(out) > topN {
		out = out[:topN]
	}
	return out
}

func buildGraficoTopClientesDesdeVentas(ventas []dbpkg.CarritoCompra, topN int) []empresaGraficoRankingCliente {
	type agg struct {
		cliente string
		ventas  int64
		total   float64
	}
	aggByClient := make(map[string]*agg)
	for _, venta := range ventas {
		cliente := reportesFirstNonBlank(venta.ClienteNombre, "Sin cliente")
		entry, ok := aggByClient[cliente]
		if !ok {
			entry = &agg{cliente: cliente}
			aggByClient[cliente] = entry
		}
		entry.ventas++
		entry.total += reportesVentaTotal(venta)
	}

	out := make([]empresaGraficoRankingCliente, 0, len(aggByClient))
	for _, item := range aggByClient {
		out = append(out, empresaGraficoRankingCliente{
			Cliente: item.cliente,
			Ventas:  item.ventas,
			Total:   reportesRound(item.total),
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Total == out[j].Total {
			return out[i].Ventas > out[j].Ventas
		}
		return out[i].Total > out[j].Total
	})
	if len(out) > topN {
		out = out[:topN]
	}
	return out
}

func graficosBucketSize(length, maxPoints int) int {
	if maxPoints <= 0 || length <= maxPoints {
		return 1
	}
	bucket := int(math.Ceil(float64(length) / float64(maxPoints)))
	if bucket <= 0 {
		return 1
	}
	return bucket
}

func compactGraficoVentasSerie(series []empresaGraficoSerieVentas, maxPoints int) []empresaGraficoSerieVentas {
	bucketSize := graficosBucketSize(len(series), maxPoints)
	if bucketSize <= 1 {
		return series
	}
	out := make([]empresaGraficoSerieVentas, 0, maxPoints)
	for i := 0; i < len(series); i += bucketSize {
		end := i + bucketSize
		if end > len(series) {
			end = len(series)
		}
		chunk := series[i:end]
		agg := empresaGraficoSerieVentas{Fecha: chunk[0].Fecha}
		for _, item := range chunk {
			agg.Ventas += item.Ventas
			agg.Total = reportesRound(agg.Total + item.Total)
		}
		out = append(out, agg)
	}
	return out
}

func compactGraficoFinanzasSerie(series []empresaGraficoSerieFinanzas, maxPoints int) []empresaGraficoSerieFinanzas {
	bucketSize := graficosBucketSize(len(series), maxPoints)
	if bucketSize <= 1 {
		return series
	}
	out := make([]empresaGraficoSerieFinanzas, 0, maxPoints)
	for i := 0; i < len(series); i += bucketSize {
		end := i + bucketSize
		if end > len(series) {
			end = len(series)
		}
		chunk := series[i:end]
		agg := empresaGraficoSerieFinanzas{Fecha: chunk[0].Fecha}
		for _, item := range chunk {
			agg.Ingresos = reportesRound(agg.Ingresos + item.Ingresos)
			agg.Egresos = reportesRound(agg.Egresos + item.Egresos)
		}
		agg.Balance = reportesRound(agg.Ingresos - agg.Egresos)
		out = append(out, agg)
	}
	return out
}

func compactGraficoComprasSerie(series []empresaGraficoSerieCompras, maxPoints int) []empresaGraficoSerieCompras {
	bucketSize := graficosBucketSize(len(series), maxPoints)
	if bucketSize <= 1 {
		return series
	}
	out := make([]empresaGraficoSerieCompras, 0, maxPoints)
	for i := 0; i < len(series); i += bucketSize {
		end := i + bucketSize
		if end > len(series) {
			end = len(series)
		}
		chunk := series[i:end]
		agg := empresaGraficoSerieCompras{Fecha: chunk[0].Fecha}
		for _, item := range chunk {
			agg.Movimientos += item.Movimientos
			agg.Costo = reportesRound(agg.Costo + item.Costo)
		}
		out = append(out, agg)
	}
	return out
}

func compactGraficoAsistenciaSerie(series []empresaGraficoSerieAsistencia, maxPoints int) []empresaGraficoSerieAsistencia {
	bucketSize := graficosBucketSize(len(series), maxPoints)
	if bucketSize <= 1 {
		return series
	}
	out := make([]empresaGraficoSerieAsistencia, 0, maxPoints)
	for i := 0; i < len(series); i += bucketSize {
		end := i + bucketSize
		if end > len(series) {
			end = len(series)
		}
		chunk := series[i:end]
		agg := empresaGraficoSerieAsistencia{Fecha: chunk[0].Fecha}
		for _, item := range chunk {
			agg.Registros += item.Registros
			agg.Presentes += item.Presentes
			agg.Ausentes += item.Ausentes
			agg.Tardanzas += item.Tardanzas
			agg.Horas = reportesRound(agg.Horas + item.Horas)
		}
		out = append(out, agg)
	}
	return out
}

func buildGraficoVentasSerie(rows []map[string]interface{}, maxPoints int) []empresaGraficoSerieVentas {
	byDate := make(map[string]*empresaGraficoSerieVentas)
	for _, row := range rows {
		fecha := reportesNormalizeDatePart(graficoString(row["fecha_pago"]))
		if fecha == "" {
			continue
		}
		entry, ok := byDate[fecha]
		if !ok {
			entry = &empresaGraficoSerieVentas{Fecha: fecha}
			byDate[fecha] = entry
		}
		entry.Ventas++
		entry.Total = reportesRound(entry.Total + graficoFloat(row["total"]))
	}

	dates := make([]string, 0, len(byDate))
	for date := range byDate {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	series := make([]empresaGraficoSerieVentas, 0, len(dates))
	for _, date := range dates {
		item := byDate[date]
		series = append(series, empresaGraficoSerieVentas{
			Fecha:  item.Fecha,
			Ventas: item.Ventas,
			Total:  reportesRound(item.Total),
		})
	}
	return compactGraficoVentasSerie(series, maxPoints)
}

func buildGraficoFinanzasSerie(rows []map[string]interface{}, maxPoints int) []empresaGraficoSerieFinanzas {
	byDate := make(map[string]*empresaGraficoSerieFinanzas)
	for _, row := range rows {
		fecha := reportesNormalizeDatePart(graficoString(row["fecha_movimiento"]))
		if fecha == "" {
			continue
		}
		entry, ok := byDate[fecha]
		if !ok {
			entry = &empresaGraficoSerieFinanzas{Fecha: fecha}
			byDate[fecha] = entry
		}

		tipo := strings.ToLower(strings.TrimSpace(graficoString(row["tipo_movimiento"])))
		monto := graficoFloat(row["total_neto"])
		if monto == 0 {
			monto = graficoFloat(row["monto"])
		}
		switch tipo {
		case "ingreso":
			entry.Ingresos = reportesRound(entry.Ingresos + monto)
		case "egreso":
			entry.Egresos = reportesRound(entry.Egresos + monto)
		}
	}

	dates := make([]string, 0, len(byDate))
	for date := range byDate {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	series := make([]empresaGraficoSerieFinanzas, 0, len(dates))
	for _, date := range dates {
		item := byDate[date]
		item.Balance = reportesRound(item.Ingresos - item.Egresos)
		series = append(series, empresaGraficoSerieFinanzas{
			Fecha:    item.Fecha,
			Ingresos: reportesRound(item.Ingresos),
			Egresos:  reportesRound(item.Egresos),
			Balance:  reportesRound(item.Balance),
		})
	}
	return compactGraficoFinanzasSerie(series, maxPoints)
}

func buildGraficoComprasSerie(rows []map[string]interface{}, maxPoints int) []empresaGraficoSerieCompras {
	byDate := make(map[string]*empresaGraficoSerieCompras)
	for _, row := range rows {
		fecha := reportesNormalizeDatePart(reportesFirstNonBlank(
			graficoString(row["fecha"]),
			graficoString(row["fecha_documento"]),
			graficoString(row["ultima_fecha_documento"]),
			graficoString(row["fecha_movimiento"]),
		))
		if fecha == "" {
			continue
		}
		entry, ok := byDate[fecha]
		if !ok {
			entry = &empresaGraficoSerieCompras{Fecha: fecha}
			byDate[fecha] = entry
		}

		movimientos := int64(graficoFloat(row["movimientos"]))
		if movimientos <= 0 {
			movimientos = int64(graficoFloat(row["documentos"]))
		}
		if movimientos <= 0 {
			movimientos = int64(graficoFloat(row["ordenes_emitidas"]))
		}
		if movimientos <= 0 {
			movimientos = 1
		}
		entry.Movimientos += movimientos

		costo := graficoFloat(row["costo_total"])
		if costo == 0 {
			costo = graficoFloat(row["monto_ordenado"])
		}
		if costo == 0 {
			costo = graficoFloat(row["monto_recepcionado"])
		}
		if costo == 0 {
			costo = graficoFloat(row["monto_contabilizado"])
		}
		if costo == 0 {
			costo = graficoFloat(row["total_neto"])
		}
		if costo == 0 {
			costo = graficoFloat(row["monto"])
		}
		entry.Costo = reportesRound(entry.Costo + costo)
	}

	dates := make([]string, 0, len(byDate))
	for date := range byDate {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	series := make([]empresaGraficoSerieCompras, 0, len(dates))
	for _, date := range dates {
		item := byDate[date]
		series = append(series, empresaGraficoSerieCompras{
			Fecha:       item.Fecha,
			Movimientos: item.Movimientos,
			Costo:       reportesRound(item.Costo),
		})
	}
	return compactGraficoComprasSerie(series, maxPoints)
}

func buildGraficoComprasSerieDesdeFinanzas(rows []map[string]interface{}, maxPoints int) []empresaGraficoSerieCompras {
	byDate := make(map[string]*empresaGraficoSerieCompras)
	for _, row := range rows {
		tipo := strings.ToLower(strings.TrimSpace(graficoString(row["tipo_movimiento"])))
		categoria := strings.ToLower(strings.TrimSpace(graficoString(row["categoria"])))
		concepto := strings.ToLower(strings.TrimSpace(graficoString(row["concepto"])))
		if tipo != "egreso" && !strings.Contains(categoria, "compra") && !strings.Contains(concepto, "compra") {
			continue
		}

		fecha := reportesNormalizeDatePart(reportesFirstNonBlank(
			graficoString(row["fecha_movimiento"]),
			graficoString(row["fecha"]),
		))
		if fecha == "" {
			continue
		}

		entry, ok := byDate[fecha]
		if !ok {
			entry = &empresaGraficoSerieCompras{Fecha: fecha}
			byDate[fecha] = entry
		}

		entry.Movimientos += 1
		monto := graficoFloat(row["total_neto"])
		if monto == 0 {
			monto = graficoFloat(row["monto"])
		}
		if monto == 0 {
			monto = graficoFloat(row["total"])
		}
		entry.Costo = reportesRound(entry.Costo + monto)
	}

	dates := make([]string, 0, len(byDate))
	for date := range byDate {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	series := make([]empresaGraficoSerieCompras, 0, len(dates))
	for _, date := range dates {
		item := byDate[date]
		series = append(series, empresaGraficoSerieCompras{
			Fecha:       item.Fecha,
			Movimientos: item.Movimientos,
			Costo:       reportesRound(item.Costo),
		})
	}
	return compactGraficoComprasSerie(series, maxPoints)
}

func buildGraficoAsistenciaSerie(items []dbpkg.EmpresaAsistenciaEmpleado, maxPoints int) []empresaGraficoSerieAsistencia {
	byDate := make(map[string]*empresaGraficoSerieAsistencia)
	for _, item := range items {
		fecha := reportesNormalizeDatePart(item.FechaAsistencia)
		if fecha == "" {
			continue
		}
		entry, ok := byDate[fecha]
		if !ok {
			entry = &empresaGraficoSerieAsistencia{Fecha: fecha}
			byDate[fecha] = entry
		}
		entry.Registros++

		estado := strings.ToLower(strings.TrimSpace(item.EstadoAsistencia))
		if estado == "ausente" || estado == "falta" {
			entry.Ausentes++
		} else {
			entry.Presentes++
		}
		if item.MinutosTarde > 0 || estado == "tarde" || estado == "retardo" {
			entry.Tardanzas++
		}
		entry.Horas = reportesRound(entry.Horas + item.HorasTrabajadas)
	}

	dates := make([]string, 0, len(byDate))
	for date := range byDate {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	series := make([]empresaGraficoSerieAsistencia, 0, len(dates))
	for _, date := range dates {
		item := byDate[date]
		series = append(series, empresaGraficoSerieAsistencia{
			Fecha:     item.Fecha,
			Registros: item.Registros,
			Presentes: item.Presentes,
			Ausentes:  item.Ausentes,
			Tardanzas: item.Tardanzas,
			Horas:     reportesRound(item.Horas),
		})
	}
	return compactGraficoAsistenciaSerie(series, maxPoints)
}

func buildGraficoTopProductos(rows []map[string]interface{}, topN int) []empresaGraficoRankingProducto {
	out := make([]empresaGraficoRankingProducto, 0, len(rows))
	for _, row := range rows {
		out = append(out, empresaGraficoRankingProducto{
			Producto: graficoString(row["producto"]),
			Cantidad: reportesRound(graficoFloat(row["cantidad_vendida"])),
			Total:    reportesRound(graficoFloat(row["total_vendido"])),
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Total == out[j].Total {
			return out[i].Cantidad > out[j].Cantidad
		}
		return out[i].Total > out[j].Total
	})
	if len(out) > topN {
		out = out[:topN]
	}
	return out
}

func buildGraficoTopClientes(rows []map[string]interface{}, topN int) []empresaGraficoRankingCliente {
	out := make([]empresaGraficoRankingCliente, 0, len(rows))
	for _, row := range rows {
		out = append(out, empresaGraficoRankingCliente{
			Cliente: graficoString(row["cliente"]),
			Ventas:  int64(graficoFloat(row["ventas"])),
			Total:   reportesRound(graficoFloat(row["total_comprado"])),
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Total == out[j].Total {
			return out[i].Ventas > out[j].Ventas
		}
		return out[i].Total > out[j].Total
	})
	if len(out) > topN {
		out = out[:topN]
	}
	return out
}

func buildGraficoStockDistribucion(rows []map[string]interface{}) []empresaGraficoDistribucionItem {
	counter := map[string]float64{}
	for _, row := range rows {
		key := strings.ToLower(strings.TrimSpace(graficoString(row["estado_stock"])))
		if key == "" {
			continue
		}
		counter[key]++
	}
	labels := map[string]string{
		"sin_stock":   "Sin stock",
		"bajo_minimo": "Bajo mínimo",
		"ok":          "Stock estable",
		"sobre_stock": "Sobre stock",
	}
	return buildGraficoDistribucion(counter, labels)
}

func buildGraficoAsistenciaDistribucion(items []dbpkg.EmpresaAsistenciaEmpleado) []empresaGraficoDistribucionItem {
	counter := map[string]float64{}
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.EstadoAsistencia))
		if key == "" {
			key = "pendiente"
		}
		counter[key]++
	}
	labels := map[string]string{
		"presente":    "Presente",
		"ausente":     "Ausente",
		"tarde":       "Tarde",
		"permiso":     "Permiso",
		"incapacidad": "Incapacidad",
		"pendiente":   "Pendiente",
	}
	return buildGraficoDistribucion(counter, labels)
}

func buildGraficoDistribucion(counter map[string]float64, labels map[string]string) []empresaGraficoDistribucionItem {
	out := make([]empresaGraficoDistribucionItem, 0, len(counter))
	for key, value := range counter {
		if value <= 0 {
			continue
		}
		label := labels[key]
		if strings.TrimSpace(label) == "" {
			label = strings.ToUpper(strings.ReplaceAll(key, "_", " "))
		}
		out = append(out, empresaGraficoDistribucionItem{
			Key:   key,
			Label: label,
			Value: reportesRound(value),
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Value == out[j].Value {
			return out[i].Label < out[j].Label
		}
		return out[i].Value > out[j].Value
	})
	return out
}

func graficoString(raw interface{}) string {
	switch value := raw.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(value)
	default:
		return strings.TrimSpace(reportesStringValue(value))
	}
}

func graficoFloat(raw interface{}) float64 {
	switch value := raw.(type) {
	case nil:
		return 0
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
	case string:
		n, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return 0
		}
		return n
	default:
		n, err := strconv.ParseFloat(strings.TrimSpace(reportesStringValue(value)), 64)
		if err != nil {
			return 0
		}
		return n
	}
}
