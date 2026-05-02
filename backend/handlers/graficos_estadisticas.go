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

type empresaGraficosSaludArea struct {
	Key       string `json:"key"`
	Label     string `json:"label"`
	Score     int    `json:"score"`
	Status    string `json:"status"`
	Resumen   string `json:"resumen"`
	Prioridad bool   `json:"prioridad"`
}

type empresaGraficosSaludEjecutiva struct {
	GlobalScore int                        `json:"global_score"`
	Status      string                     `json:"status"`
	Resumen     string                     `json:"resumen"`
	Areas       []empresaGraficosSaludArea `json:"areas"`
	Prioridades []string                   `json:"prioridades"`
}

type empresaGraficosObjetivoMetrica struct {
	Key             string  `json:"key"`
	Label           string  `json:"label"`
	Unidad          string  `json:"unidad"`
	Actual          float64 `json:"actual"`
	Objetivo        float64 `json:"objetivo"`
	Brecha          float64 `json:"brecha"`
	CumplimientoPct float64 `json:"cumplimiento_pct"`
	Status          string  `json:"status"`
	Descripcion     string  `json:"descripcion"`
}

type empresaGraficosObjetivos struct {
	Base     string                           `json:"base"`
	Resumen  string                           `json:"resumen"`
	Metricas []empresaGraficosObjetivoMetrica `json:"metricas"`
}

type empresaGraficosPresupuestoPartida struct {
	Key           string  `json:"key"`
	Label         string  `json:"label"`
	Unidad        string  `json:"unidad"`
	Direccion     string  `json:"direccion"`
	Ejecutado     float64 `json:"ejecutado"`
	Presupuesto   float64 `json:"presupuesto"`
	Desviacion    float64 `json:"desviacion"`
	DesviacionPct float64 `json:"desviacion_pct"`
	Status        string  `json:"status"`
	Nota          string  `json:"nota"`
}

type empresaGraficosPresupuesto struct {
	Base     string                              `json:"base"`
	Resumen  string                              `json:"resumen"`
	Partidas []empresaGraficosPresupuestoPartida `json:"partidas"`
}

type empresaGraficosSemaforoPredictivo struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Severity string `json:"severity"`
	Signal   string `json:"signal"`
	Resumen  string `json:"resumen"`
	Accion   string `json:"accion"`
}

type empresaGraficosRentabilidadItem struct {
	Key             string  `json:"key"`
	Label           string  `json:"label"`
	Ingresos        float64 `json:"ingresos"`
	CostoEstimado   float64 `json:"costo_estimado"`
	MargenEstimado  float64 `json:"margen_estimado"`
	RentabilidadPct float64 `json:"rentabilidad_pct"`
	Volumen         float64 `json:"volumen"`
}

type empresaGraficosRentabilidad struct {
	BaseCosto string                            `json:"base_costo"`
	Linea     []empresaGraficosRentabilidadItem `json:"linea"`
	Sede      []empresaGraficosRentabilidadItem `json:"sede"`
	Canal     []empresaGraficosRentabilidadItem `json:"canal"`
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
	Salud          empresaGraficosSaludEjecutiva       `json:"salud"`
	Objetivos      empresaGraficosObjetivos            `json:"objetivos"`
	Presupuesto    empresaGraficosPresupuesto          `json:"presupuesto"`
	Predictivos    []empresaGraficosSemaforoPredictivo `json:"predictivos"`
	Rentabilidad   empresaGraficosRentabilidad         `json:"rentabilidad"`
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

type empresaGraficosRentabilidadAgg struct {
	label   string
	revenue float64
	volume  float64
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
	panel.Salud = graficosBuildExecutiveHealth(panel.Tablero, panel.Comparativo)
	panel.Objetivos = graficosBuildObjetivosVsReal(panel.Tablero, panel.Comparativo)
	panel.Presupuesto = graficosBuildPresupuestoVsEjecucion(panel.Tablero, panel.Comparativo)
	panel.Predictivos = graficosBuildPredictiveSignals(panel.Tablero, panel.Comparativo, panel.Objetivos, panel.Presupuesto, panel.Salud)
	panel.Rentabilidad = graficosBuildRentabilidad(dbEmp, builder.empresaID, filteredVentas, builder.itemsCache, filterCtx.carritoEstacion, panel.Tablero)

	return panel, nil
}

func graficosBuildObjetivosVsReal(tablero dbpkg.EmpresaReportesTableroResumen, comparativo *empresaGraficosComparativo) empresaGraficosObjetivos {
	base := "Objetivos sugeridos desde el comportamiento actual del período."
	if comparativo != nil {
		base = "Objetivos sugeridos usando el período de referencia comparativa como línea base."
	}

	buildMetric := func(key, label, unidad string, actual float64, growthPct float64, fallbackFloor float64, descripcion string) empresaGraficosObjetivoMetrica {
		baseValue := actual
		if comparativo != nil {
			if metric, ok := comparativo.Metricas[key]; ok && metric.Anterior > 0 {
				baseValue = metric.Anterior
			}
		}
		objetivo := reportesRound(baseValue * (1 + growthPct))
		if objetivo <= 0 {
			objetivo = reportesRound(math.Max(actual, fallbackFloor))
		}
		cumplimiento := 100.0
		if objetivo > 0 {
			cumplimiento = reportesRound((actual / objetivo) * 100)
		}
		return empresaGraficosObjetivoMetrica{
			Key:             key,
			Label:           label,
			Unidad:          unidad,
			Actual:          reportesRound(actual),
			Objetivo:        objetivo,
			Brecha:          reportesRound(actual - objetivo),
			CumplimientoPct: cumplimiento,
			Status:          graficosMetricComplianceStatus(cumplimiento),
			Descripcion:     descripcion,
		}
	}

	metricas := []empresaGraficosObjetivoMetrica{
		buildMetric("ingresos_ventas", "Ingresos comerciales", "money", tablero.Operativo.IngresosVentas, 0.08, 0, "Compara la venta real contra una meta comercial sugerida para el mismo frente de ingresos."),
		buildMetric("ventas_cerradas", "Ventas cerradas", "count", float64(tablero.Operativo.VentasCerradas), 0.05, 1, "Evalúa volumen de cierres frente a una meta operativa sugerida."),
		buildMetric("ticket_promedio", "Ticket promedio", "money", tablero.Operativo.TicketPromedio, 0.04, 0, "Mide si el valor promedio por venta ya acompaña el crecimiento buscado."),
		buildMetric("utilidad_operacional", "Utilidad operacional", "money", tablero.EstadoResultados.UtilidadOperacional, 0.10, 0, "Contrasta el margen real con el nivel sugerido para sostener rentabilidad."),
	}

	resumen := "Las metas sugeridas ayudan a leer avance comercial y margen sin depender todavía de una tabla manual de objetivos."
	return empresaGraficosObjetivos{
		Base:     base,
		Resumen:  resumen,
		Metricas: metricas,
	}
}

func graficosBuildPresupuestoVsEjecucion(tablero dbpkg.EmpresaReportesTableroResumen, comparativo *empresaGraficosComparativo) empresaGraficosPresupuesto {
	actualIngresos := math.Max(tablero.Operativo.IngresosVentas, 0)
	actualCompras := math.Max(tablero.Operativo.ComprasCosto, 0)
	actualEgresos := math.Max(tablero.Financiero.Egresos, 0)
	actualUtilidad := tablero.EstadoResultados.UtilidadOperacional

	refIngresos := actualIngresos
	if comparativo != nil {
		if metric, ok := comparativo.Metricas["ingresos_ventas"]; ok && metric.Anterior > 0 {
			refIngresos = metric.Anterior
		}
	}
	if refIngresos <= 0 {
		refIngresos = actualIngresos
	}
	if refIngresos <= 0 {
		refIngresos = 1
	}

	purchaseRatio := graficosClampFloat(actualCompras/refIngresos, 0.18, 0.78)
	opexRatio := graficosClampFloat(actualEgresos/refIngresos, 0.08, 0.55)

	presupuestoIngresos := reportesRound(refIngresos * 1.06)
	if presupuestoIngresos <= 0 {
		presupuestoIngresos = reportesRound(actualIngresos)
	}
	presupuestoCompras := reportesRound(presupuestoIngresos * graficosClampFloat(purchaseRatio*0.97, 0.16, 0.74))
	presupuestoEgresos := reportesRound(presupuestoIngresos * graficosClampFloat(opexRatio*0.96, 0.07, 0.5))
	presupuestoUtilidad := reportesRound(math.Max(presupuestoIngresos-presupuestoCompras-presupuestoEgresos, 0))

	partidas := []empresaGraficosPresupuestoPartida{
		graficosBuildBudgetItem("ingresos", "Ingresos presupuestados", "money", "higher_better", actualIngresos, presupuestoIngresos, "La ejecución comercial idealmente debe alcanzar o superar la meta de ingresos."),
		graficosBuildBudgetItem("compras", "Costo de abastecimiento", "money", "lower_better", actualCompras, presupuestoCompras, "Compara compras y recepción frente al nivel sugerido para proteger margen."),
		graficosBuildBudgetItem("egresos", "Caja operacional", "money", "lower_better", actualEgresos, presupuestoEgresos, "Vigila si los egresos ya se están yendo por encima del ritmo sostenible."),
		graficosBuildBudgetItem("utilidad", "Utilidad esperada", "money", "higher_better", actualUtilidad, presupuestoUtilidad, "Resume si el período ya está materializando el margen que se esperaba producir."),
	}

	return empresaGraficosPresupuesto{
		Base:     "Presupuesto sugerido desde histórico reciente y ratios reales de operación.",
		Resumen:  "Esta vista traduce históricos y ratios de costo en una ejecución esperada para ventas, compras, caja y margen.",
		Partidas: partidas,
	}
}

func graficosBuildBudgetItem(key, label, unidad, direccion string, ejecutado, presupuesto float64, nota string) empresaGraficosPresupuestoPartida {
	desviacion := reportesRound(ejecutado - presupuesto)
	desviacionPct := 0.0
	if presupuesto != 0 {
		desviacionPct = reportesRound((desviacion / math.Abs(presupuesto)) * 100)
	}
	return empresaGraficosPresupuestoPartida{
		Key:           key,
		Label:         label,
		Unidad:        unidad,
		Direccion:     direccion,
		Ejecutado:     reportesRound(ejecutado),
		Presupuesto:   reportesRound(presupuesto),
		Desviacion:    desviacion,
		DesviacionPct: desviacionPct,
		Status:        graficosBudgetStatus(direccion, ejecutado, presupuesto),
		Nota:          nota,
	}
}

func graficosBuildPredictiveSignals(tablero dbpkg.EmpresaReportesTableroResumen, comparativo *empresaGraficosComparativo, objetivos empresaGraficosObjetivos, presupuesto empresaGraficosPresupuesto, salud empresaGraficosSaludEjecutiva) []empresaGraficosSemaforoPredictivo {
	out := make([]empresaGraficosSemaforoPredictivo, 0, 5)

	if tablero.Financiero.Balance < 0 || tablero.EstadoResultados.UtilidadOperacional < 0 {
		out = append(out, empresaGraficosSemaforoPredictivo{
			Key:      "caja_margen",
			Label:    "Presión sobre caja y margen",
			Severity: "critico",
			Signal:   "Balance o utilidad ya están en rojo.",
			Resumen:  "La empresa ya entró en zona de presión financiera y puede deteriorar liquidez si mantiene el mismo ritmo.",
			Accion:   "Revisar egresos, compras y precios antes del siguiente cierre.",
		})
	} else {
		out = append(out, empresaGraficosSemaforoPredictivo{
			Key:      "caja_margen",
			Label:    "Presión sobre caja y margen",
			Severity: "estable",
			Signal:   "Caja y utilidad siguen respirando.",
			Resumen:  "No hay presión crítica inmediata en liquidez o margen, aunque conviene vigilar variaciones semanales.",
			Accion:   "Mantener control de egresos y seguir comparando contra el período anterior.",
		})
	}

	commercialSeverity := "estable"
	commercialSignal := "La demanda sostiene ritmo suficiente."
	commercialResumen := "El frente comercial no muestra enfriamiento severo con la información actual."
	if comparativo != nil {
		if metric, ok := comparativo.Metricas["ingresos_ventas"]; ok && metric.VariacionPct <= -12 {
			commercialSeverity = "atencion"
			commercialSignal = "Los ingresos caen frente al período de referencia."
			commercialResumen = "Si la tendencia se mantiene, la empresa puede cerrar el siguiente período con menor volumen y menor absorción de costos."
		}
	}
	for _, metrica := range objetivos.Metricas {
		if metrica.Key == "ingresos_ventas" && metrica.CumplimientoPct < 85 {
			commercialSeverity = "critico"
			commercialSignal = "Los ingresos van por debajo de la meta sugerida."
			commercialResumen = "La velocidad comercial actual no alcanza para cerrar el período en el nivel objetivo."
			break
		}
	}
	out = append(out, empresaGraficosSemaforoPredictivo{
		Key:      "demanda",
		Label:    "Tracción comercial",
		Severity: commercialSeverity,
		Signal:   commercialSignal,
		Resumen:  commercialResumen,
		Accion:   "Activar campañas, revisar ticket promedio y priorizar recuperación de clientes.",
	})

	inventorySeverity := "solido"
	if tablero.Operativo.ProductosBajoMinimo > 0 {
		inventorySeverity = "atencion"
	}
	if tablero.Operativo.ProductosBajoMinimo >= 8 {
		inventorySeverity = "critico"
	}
	out = append(out, empresaGraficosSemaforoPredictivo{
		Key:      "inventario",
		Label:    "Continuidad operativa",
		Severity: inventorySeverity,
		Signal:   fmt.Sprintf("%d productos bajo mínimo.", tablero.Operativo.ProductosBajoMinimo),
		Resumen:  "El semáforo proyecta riesgo de quiebre comercial cuando el inventario crítico ya empieza a comprimirse.",
		Accion:   "Reponer primero las referencias de mayor margen o mayor rotación.",
	})

	controlSeverity := "estable"
	if tablero.Contable.EventosPendientes > 0 || tablero.Financiero.PeriodosAbiertos > 1 {
		controlSeverity = "atencion"
	}
	if math.Abs(tablero.BalanceGeneral.Cuadre) > 1 || tablero.Contable.EventosPendientes > 10 {
		controlSeverity = "critico"
	}
	out = append(out, empresaGraficosSemaforoPredictivo{
		Key:      "control",
		Label:    "Gobierno y trazabilidad",
		Severity: controlSeverity,
		Signal:   fmt.Sprintf("%d eventos pendientes y %d períodos abiertos.", tablero.Contable.EventosPendientes, tablero.Financiero.PeriodosAbiertos),
		Resumen:  "Las señales de control anticipan fricción en cierres, conciliaciones y lectura gerencial si se siguen acumulando pendientes.",
		Accion:   "Cerrar períodos, procesar eventos y corregir diferencias de balance oportunamente.",
	})

	healthSeverity := graficosNormalizeSeverity(salud.Status)
	if healthSeverity == "solido" {
		healthSeverity = "estable"
	}
	out = append(out, empresaGraficosSemaforoPredictivo{
		Key:      "salud_global",
		Label:    "Salud ejecutiva consolidada",
		Severity: healthSeverity,
		Signal:   fmt.Sprintf("Salud global %d/100.", salud.GlobalScore),
		Resumen:  salud.Resumen,
		Accion:   "Intervenir primero las prioridades listadas para evitar que las alertas se vuelvan estructurales.",
	})

	return out
}

func graficosBuildRentabilidad(dbEmp *sql.DB, empresaID int64, ventas []dbpkg.CarritoCompra, itemsCache map[int64][]dbpkg.CarritoCompraItem, carritoEstacion map[int64]int64, tablero dbpkg.EmpresaReportesTableroResumen) empresaGraficosRentabilidad {
	costRatio := 0.42
	if tablero.Operativo.IngresosVentas > 0 {
		costRatio = graficosClampFloat(tablero.Operativo.ComprasCosto/tablero.Operativo.IngresosVentas, 0.18, 0.78)
	}
	opexRatio := 0.18
	if tablero.Operativo.IngresosVentas > 0 {
		opexRatio = graficosClampFloat(tablero.Financiero.Egresos/tablero.Operativo.IngresosVentas, 0.06, 0.45)
	}

	lineas := make(map[string]*empresaGraficosRentabilidadAgg)
	canales := make(map[string]*empresaGraficosRentabilidadAgg)
	sedes := make(map[string]*empresaGraficosRentabilidadAgg)

	stationLabels := graficosLoadEstacionLabels(dbEmp, empresaID)

	for _, venta := range ventas {
		totalVenta := reportesVentaTotal(venta)
		canalKey, canalLabel := graficosNormalizeCanalLabel(venta.CanalVenta)
		canalAgg := canales[canalKey]
		if canalAgg == nil {
			canalAgg = &empresaGraficosRentabilidadAgg{label: canalLabel}
			canales[canalKey] = canalAgg
		}
		canalAgg.revenue += totalVenta
		canalAgg.volume++

		estacionID := carritoEstacion[venta.ID]
		sedeKey, sedeLabel := graficosSedeLabel(estacionID, stationLabels)
		sedeAgg := sedes[sedeKey]
		if sedeAgg == nil {
			sedeAgg = &empresaGraficosRentabilidadAgg{label: sedeLabel}
			sedes[sedeKey] = sedeAgg
		}
		sedeAgg.revenue += totalVenta
		sedeAgg.volume++

		for _, item := range itemsCache[venta.ID] {
			if strings.EqualFold(strings.TrimSpace(item.Estado), "inactivo") {
				continue
			}
			lineKey, lineLabel := graficosLineaItem(item)
			lineAgg := lineas[lineKey]
			if lineAgg == nil {
				lineAgg = &empresaGraficosRentabilidadAgg{label: lineLabel}
				lineas[lineKey] = lineAgg
			}
			lineAgg.revenue += item.TotalLinea
			lineAgg.volume += math.Max(item.Cantidad, 1)
		}
	}

	return empresaGraficosRentabilidad{
		BaseCosto: "Rentabilidad estimada con mezcla de costo de abastecimiento y carga operativa real del período.",
		Linea:     graficosFinalizeRentabilidad(lineas, func(key string) float64 { return graficosLineaCostoRatio(key, costRatio, opexRatio) }, 6),
		Sede:      graficosFinalizeRentabilidad(sedes, func(key string) float64 { return graficosSedeCostoRatio(key, costRatio, opexRatio) }, 6),
		Canal:     graficosFinalizeRentabilidad(canales, func(key string) float64 { return graficosCanalCostoRatio(key, costRatio, opexRatio) }, 6),
	}
}

func graficosMetricComplianceStatus(cumplimiento float64) string {
	switch {
	case cumplimiento >= 105:
		return "solido"
	case cumplimiento >= 95:
		return "estable"
	case cumplimiento >= 80:
		return "atencion"
	default:
		return "critico"
	}
}

func graficosBudgetStatus(direccion string, ejecutado, presupuesto float64) string {
	if strings.EqualFold(strings.TrimSpace(direccion), "lower_better") {
		if ejecutado <= presupuesto {
			return "solido"
		}
		if presupuesto <= 0 {
			return "atencion"
		}
		excesoPct := ((ejecutado - presupuesto) / math.Abs(presupuesto)) * 100
		if excesoPct <= 10 {
			return "atencion"
		}
		return "critico"
	}

	if presupuesto <= 0 {
		if ejecutado > 0 {
			return "solido"
		}
		return "estable"
	}
	cumplimiento := (ejecutado / presupuesto) * 100
	return graficosMetricComplianceStatus(cumplimiento)
}

func graficosNormalizeSeverity(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "solido":
		return "solido"
	case "estable":
		return "estable"
	case "atencion":
		return "atencion"
	case "critico":
		return "critico"
	default:
		return "estable"
	}
}

func graficosClampFloat(value, minVal, maxVal float64) float64 {
	if value < minVal {
		return minVal
	}
	if value > maxVal {
		return maxVal
	}
	return value
}

func graficosLoadEstacionLabels(dbEmp *sql.DB, empresaID int64) map[int64]string {
	out := make(map[int64]string)
	rows, err := dbEmp.Query(`SELECT
		COALESCE(estacion_id, 0),
		COALESCE(estacion_nombre, ''),
		COALESCE(estacion_codigo, '')
	FROM empresa_ventas_estacion_metricas
	WHERE empresa_id = ?
		AND LOWER(COALESCE(estado, 'activo')) = 'activo'
		AND COALESCE(estacion_id, 0) > 0
	ORDER BY id DESC`, empresaID)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var nombre string
		var codigo string
		if err := rows.Scan(&id, &nombre, &codigo); err != nil {
			return out
		}
		if id <= 0 {
			continue
		}
		if _, ok := out[id]; ok {
			continue
		}
		out[id] = reportesFirstNonBlank(strings.TrimSpace(nombre), strings.TrimSpace(codigo), fmt.Sprintf("Estación %d", id))
	}
	return out
}

func graficosLineaItem(item dbpkg.CarritoCompraItem) (string, string) {
	tipo := strings.ToLower(strings.TrimSpace(item.TipoItem))
	switch tipo {
	case "producto":
		return "producto", "Productos"
	case "servicio":
		return "servicio", "Servicios"
	case "combo", "combo_producto":
		return "combo", "Combos"
	case "tarifa_por_dia", "tarifa_dia", "estadia", "habitacion":
		return "tarifa_dia", "Tarifas por día"
	case "tarifa_por_minutos", "tarifa_minutos", "minutos":
		return "tarifa_minutos", "Tarifas por minutos"
	default:
		if tipo == "" {
			desc := strings.ToLower(strings.TrimSpace(item.Descripcion))
			switch {
			case strings.Contains(desc, "tarifa") && strings.Contains(desc, "dia"):
				return "tarifa_dia", "Tarifas por día"
			case strings.Contains(desc, "tarifa") && strings.Contains(desc, "minuto"):
				return "tarifa_minutos", "Tarifas por minutos"
			}
		}
		return reportesFirstNonBlank(tipo, "otros"), reportesFirstNonBlank(strings.Title(tipo), "Otros")
	}
}

func graficosNormalizeCanalLabel(raw string) (string, string) {
	canal := strings.ToLower(strings.TrimSpace(raw))
	switch canal {
	case "", "mostrador":
		return "mostrador", "Mostrador"
	case "estacion", "estaciones":
		return "estacion", "Estación"
	case "vip", "portal_vip", "portal vip":
		return "vip", "Portal VIP"
	case "venta_publica", "publico", "publica":
		return "venta_publica", "Venta pública"
	case "crm", "comercial":
		return "crm", "CRM comercial"
	default:
		return canal, strings.Title(strings.ReplaceAll(canal, "_", " "))
	}
}

func graficosSedeLabel(estacionID int64, labels map[int64]string) (string, string) {
	if estacionID <= 0 {
		return "sin_sede", "Operación general"
	}
	if label, ok := labels[estacionID]; ok && strings.TrimSpace(label) != "" {
		return fmt.Sprintf("estacion_%d", estacionID), label
	}
	return fmt.Sprintf("estacion_%d", estacionID), fmt.Sprintf("Estación %d", estacionID)
}

func graficosLineaCostoRatio(key string, costRatio, opexRatio float64) float64 {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "producto":
		return graficosClampFloat(costRatio*0.9+opexRatio*0.35, 0.28, 0.9)
	case "servicio":
		return graficosClampFloat(costRatio*0.3+opexRatio*0.85, 0.16, 0.82)
	case "combo":
		return graficosClampFloat(costRatio*0.8+opexRatio*0.45, 0.24, 0.88)
	case "tarifa_dia", "tarifa_minutos":
		return graficosClampFloat(costRatio*0.18+opexRatio*0.82, 0.12, 0.76)
	default:
		return graficosClampFloat(costRatio*0.5+opexRatio*0.55, 0.18, 0.85)
	}
}

func graficosCanalCostoRatio(key string, costRatio, opexRatio float64) float64 {
	base := costRatio*0.55 + opexRatio*0.55
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "venta_publica", "vip":
		base -= 0.04
	case "crm":
		base -= 0.02
	case "estacion":
		base -= 0.01
	}
	return graficosClampFloat(base, 0.16, 0.85)
}

func graficosSedeCostoRatio(key string, costRatio, opexRatio float64) float64 {
	base := costRatio*0.52 + opexRatio*0.58
	if strings.EqualFold(strings.TrimSpace(key), "sin_sede") {
		base += 0.03
	}
	return graficosClampFloat(base, 0.16, 0.86)
}

func graficosFinalizeRentabilidad(source map[string]*empresaGraficosRentabilidadAgg, ratioFn func(string) float64, limit int) []empresaGraficosRentabilidadItem {
	out := make([]empresaGraficosRentabilidadItem, 0, len(source))
	for key, item := range source {
		if item == nil || item.revenue <= 0 {
			continue
		}
		ratio := ratioFn(key)
		costo := reportesRound(item.revenue * ratio)
		margen := reportesRound(item.revenue - costo)
		rentabilidad := 0.0
		if item.revenue > 0 {
			rentabilidad = reportesRound((margen / item.revenue) * 100)
		}
		out = append(out, empresaGraficosRentabilidadItem{
			Key:             key,
			Label:           item.label,
			Ingresos:        reportesRound(item.revenue),
			CostoEstimado:   costo,
			MargenEstimado:  margen,
			RentabilidadPct: rentabilidad,
			Volumen:         reportesRound(item.volume),
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].MargenEstimado == out[j].MargenEstimado {
			return out[i].Ingresos > out[j].Ingresos
		}
		return out[i].MargenEstimado > out[j].MargenEstimado
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

func graficosBuildExecutiveHealth(tablero dbpkg.EmpresaReportesTableroResumen, comparativo *empresaGraficosComparativo) empresaGraficosSaludEjecutiva {
	operativo := tablero.Operativo
	financiero := tablero.Financiero
	contable := tablero.Contable
	estado := tablero.EstadoResultados
	balance := tablero.BalanceGeneral

	areas := make([]empresaGraficosSaludArea, 0, 4)
	prioridades := make([]string, 0, 6)

	comercialScore := 45
	if operativo.VentasCerradas > 0 {
		comercialScore += 20
	}
	if operativo.TicketPromedio > 0 {
		comercialScore += 10
	}
	if operativo.ClientesActivos > 0 {
		comercialScore += 10
	}
	if comparativo != nil {
		if metric, ok := comparativo.Metricas["ingresos_ventas"]; ok && metric.VariacionPct > 0 {
			comercialScore += 10
		}
	}
	comercialScore = graficosClampScore(comercialScore)
	comercialArea := empresaGraficosSaludArea{
		Key:     "comercial",
		Label:   "Comercial",
		Score:   comercialScore,
		Status:  graficosHealthStatus(comercialScore),
		Resumen: fmt.Sprintf("%d ventas cerradas, ticket promedio %s y %d clientes activos en el rango.", operativo.VentasCerradas, graficosFmtMoneyShort(operativo.TicketPromedio), operativo.ClientesActivos),
	}
	if operativo.VentasCerradas == 0 || operativo.ClientesActivos == 0 {
		comercialArea.Prioridad = true
		prioridades = append(prioridades, "Activar ventas y recompra: el frente comercial todavía no muestra suficiente tracción en el período.")
	}
	areas = append(areas, comercialArea)

	financieroScore := 40
	if financiero.Balance >= 0 {
		financieroScore += 25
	} else {
		prioridades = append(prioridades, "Corregir balance financiero negativo antes de seguir expandiendo egresos o compras.")
	}
	if estado.UtilidadOperacional >= 0 {
		financieroScore += 20
	} else {
		prioridades = append(prioridades, "La utilidad operacional está en rojo; conviene revisar margen, precios y costos de inmediato.")
	}
	if financiero.Ingresos >= financiero.Egresos {
		financieroScore += 10
	}
	financieroScore = graficosClampScore(financieroScore)
	financieroArea := empresaGraficosSaludArea{
		Key:       "financiero",
		Label:     "Financiero",
		Score:     financieroScore,
		Status:    graficosHealthStatus(financieroScore),
		Resumen:   fmt.Sprintf("Balance %s, utilidad operacional %s e ingresos %s frente a egresos %s.", graficosFmtMoneyShort(financiero.Balance), graficosFmtMoneyShort(estado.UtilidadOperacional), graficosFmtMoneyShort(financiero.Ingresos), graficosFmtMoneyShort(financiero.Egresos)),
		Prioridad: financiero.Balance < 0 || estado.UtilidadOperacional < 0,
	}
	areas = append(areas, financieroArea)

	operativoScore := 55
	if operativo.ProductosBajoMinimo == 0 {
		operativoScore += 20
	} else {
		prioridades = append(prioridades, fmt.Sprintf("Reabastecer %d productos bajo mínimo para proteger continuidad de venta.", operativo.ProductosBajoMinimo))
	}
	if operativo.ProductosActivos > 0 {
		operativoScore += 10
	}
	if contable.EventosPendientes == 0 {
		operativoScore += 10
	}
	if financiero.PeriodosAbiertos <= 1 {
		operativoScore += 5
	}
	operativoScore = graficosClampScore(operativoScore)
	operativoArea := empresaGraficosSaludArea{
		Key:       "operativo",
		Label:     "Operativo",
		Score:     operativoScore,
		Status:    graficosHealthStatus(operativoScore),
		Resumen:   fmt.Sprintf("%d productos activos, %d bajo mínimo y %d eventos pendientes por procesar.", operativo.ProductosActivos, operativo.ProductosBajoMinimo, contable.EventosPendientes),
		Prioridad: operativo.ProductosBajoMinimo > 0 || contable.EventosPendientes > 0,
	}
	areas = append(areas, operativoArea)

	controlScore := 60
	if math.Abs(balance.Cuadre) <= 1 {
		controlScore += 20
	} else {
		prioridades = append(prioridades, "Revisar el cuadre contable: el balance aún muestra diferencias que deben resolverse.")
	}
	if contable.EventosPendientes == 0 {
		controlScore += 10
	}
	if financiero.PeriodosAbiertos <= 1 {
		controlScore += 10
	} else {
		prioridades = append(prioridades, fmt.Sprintf("Cerrar %d períodos abiertos para mejorar control y trazabilidad.", financiero.PeriodosAbiertos))
	}
	controlScore = graficosClampScore(controlScore)
	controlArea := empresaGraficosSaludArea{
		Key:       "control",
		Label:     "Control",
		Score:     controlScore,
		Status:    graficosHealthStatus(controlScore),
		Resumen:   fmt.Sprintf("Cuadre %s, %d períodos abiertos y %d eventos pendientes.", graficosFmtMoneyShort(balance.Cuadre), financiero.PeriodosAbiertos, contable.EventosPendientes),
		Prioridad: math.Abs(balance.Cuadre) > 1 || financiero.PeriodosAbiertos > 1,
	}
	areas = append(areas, controlArea)

	if len(prioridades) == 0 {
		prioridades = append(prioridades, "Mantener el ritmo actual: no se detectan desvíos críticos en ventas, control o rentabilidad para este período.")
	}
	if len(prioridades) > 5 {
		prioridades = prioridades[:5]
	}

	total := 0
	for _, area := range areas {
		total += area.Score
	}
	globalScore := 0
	if len(areas) > 0 {
		globalScore = int(math.Round(float64(total) / float64(len(areas))))
	}
	status := graficosHealthStatus(globalScore)
	resumen := "La empresa muestra una salud ejecutiva estable, con espacio para seguir afinando foco comercial, control y margen."
	switch status {
	case "critico":
		resumen = "La lectura ejecutiva exige intervención inmediata: margen, control o continuidad operativa están comprometidos."
	case "atencion":
		resumen = "La operación sigue en pie, pero hay señales que conviene intervenir pronto para no perder margen ni control."
	case "solido":
		resumen = "La empresa sostiene una lectura ejecutiva fuerte y controlada para el rango analizado."
	}

	return empresaGraficosSaludEjecutiva{
		GlobalScore: globalScore,
		Status:      status,
		Resumen:     resumen,
		Areas:       areas,
		Prioridades: prioridades,
	}
}

func graficosClampScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func graficosHealthStatus(score int) string {
	switch {
	case score >= 80:
		return "solido"
	case score >= 60:
		return "estable"
	case score >= 40:
		return "atencion"
	default:
		return "critico"
	}
}

func graficosFmtMoneyShort(value float64) string {
	return fmt.Sprintf("$%.0f", math.Round(value))
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
