package handlers

import (
	"database/sql"
	"net/http"
	"sort"
	"strconv"
	"strings"
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
}

// EmpresaGraficosEstadisticasHandler expone datasets listos para visualizacion grafica por empresa.
func EmpresaGraficosEstadisticasHandler(dbEmp *sql.DB) http.HandlerFunc {
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

		panel, err := buildEmpresaGraficosPanel(dbEmp, builder, maxPoints, topN)
		if err != nil {
			http.Error(w, "No se pudo construir el panel de graficos y estadisticas", http.StatusInternalServerError)
			return
		}

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
			})
			return

		case "rankings", "ranking":
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"empresa_id":  panel.EmpresaID,
				"desde":       panel.Desde,
				"hasta":       panel.Hasta,
				"generado_en": panel.GeneradoEn,
				"rankings":    panel.Rankings,
			})
			return

		case "distribuciones", "distributions", "distribucion":
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"empresa_id":     panel.EmpresaID,
				"desde":          panel.Desde,
				"hasta":          panel.Hasta,
				"generado_en":    panel.GeneradoEn,
				"distribuciones": panel.Distribuciones,
			})
			return

		case "catalogo", "catalog":
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"empresa_id": panel.EmpresaID,
				"actions": []map[string]string{
					{"action": "panel", "description": "tablero consolidado con series, rankings y distribuciones"},
					{"action": "serie", "description": "serie puntual (ventas, finanzas, compras, asistencia)"},
					{"action": "rankings", "description": "top productos y top clientes"},
					{"action": "distribuciones", "description": "distribuciones de stock y asistencia"},
				},
				"series": []string{"ventas", "finanzas", "compras", "asistencia"},
			})
			return

		default:
			http.Error(w, "action invalida (use panel, serie, rankings, distribuciones o catalogo)", http.StatusBadRequest)
			return
		}
	}
}

func buildEmpresaGraficosPanel(dbEmp *sql.DB, builder *reportesBuilder, maxPoints, topN int) (empresaGraficosPanelResponse, error) {
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
	}

	ventasDataset, err := builder.buildDataset(reporteDatasetOperativoVentasDetalle)
	if err != nil {
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
	topProductosDataset, err := builder.buildDataset(reporteDatasetOperativoTopProductos)
	if err != nil {
		return empresaGraficosPanelResponse{}, err
	}
	topClientesDataset, err := builder.buildDataset(reporteDatasetOperativoTopClientes)
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

	panel.Series.Ventas = buildGraficoVentasSerie(ventasDataset.Rows, maxPoints)
	panel.Series.Finanzas = buildGraficoFinanzasSerie(finanzasDataset.Rows, maxPoints)
	panel.Series.Compras = buildGraficoComprasSerie(comprasDataset.Rows, maxPoints)
	if len(panel.Series.Compras) == 0 {
		panel.Series.Compras = buildGraficoComprasSerieDesdeFinanzas(finanzasDataset.Rows, maxPoints)
	}
	panel.Series.Asistencia = buildGraficoAsistenciaSerie(asistencias, maxPoints)

	panel.Rankings.TopProductos = buildGraficoTopProductos(topProductosDataset.Rows, topN)
	panel.Rankings.TopClientes = buildGraficoTopClientes(topClientesDataset.Rows, topN)

	panel.Distribuciones.StockEstado = buildGraficoStockDistribucion(inventarioDataset.Rows)
	panel.Distribuciones.AsistenciaEstado = buildGraficoAsistenciaDistribucion(asistencias)

	return panel, nil
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
	if len(series) > maxPoints {
		series = series[len(series)-maxPoints:]
	}
	return series
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
	if len(series) > maxPoints {
		series = series[len(series)-maxPoints:]
	}
	return series
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
	if len(series) > maxPoints {
		series = series[len(series)-maxPoints:]
	}
	return series
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
	if len(series) > maxPoints {
		series = series[len(series)-maxPoints:]
	}
	return series
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
	if len(series) > maxPoints {
		series = series[len(series)-maxPoints:]
	}
	return series
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
