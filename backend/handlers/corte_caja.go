package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type corteCajaCacheEntry struct {
	Response *corteCajaResponse
	CachedAt time.Time
}

var (
	corteCajaCacheMu sync.Mutex
	corteCajaCache   = map[string]corteCajaCacheEntry{}
)

const corteCajaCacheTTL = 10 * time.Second

type corteCajaVenta struct {
	ID             int64   `json:"id"`
	Codigo         string  `json:"codigo"`
	Nombre         string  `json:"nombre"`
	FechaPago      string  `json:"fecha_pago"`
	MetodoPago     string  `json:"metodo_pago"`
	Moneda         string  `json:"moneda"`
	Total          float64 `json:"total"`
	TotalPagado    float64 `json:"total_pagado"`
	Devolucion     float64 `json:"devolucion"`
	Cajero         string  `json:"cajero"`
	EstacionID     int64   `json:"estacion_id"`
	EstacionCodigo string  `json:"estacion_codigo"`
	EstacionNombre string  `json:"estacion_nombre"`
}

type corteCajaMovimientoGrupo struct {
	Tipo      string  `json:"tipo"`
	Metodo    string  `json:"metodo"`
	Categoria string  `json:"categoria"`
	Usuario   string  `json:"usuario"`
	Cantidad  int64   `json:"cantidad"`
	Total     float64 `json:"total"`
}

type corteCajaTipoItem struct {
	Tipo     string  `json:"tipo"`
	Cantidad float64 `json:"cantidad"`
	Total    float64 `json:"total"`
}

type corteCajaSensorAlerta struct {
	EstacionID     int64  `json:"estacion_id"`
	EstacionNombre string `json:"estacion_nombre"`
	DeviceID       string `json:"device_id"`
	EstadoSensor   string `json:"estado_sensor"`
	UltimaLectura  string `json:"ultima_lectura"`
	Motivo         string `json:"motivo"`
}

type corteCajaResumen struct {
	EmpresaID              int64   `json:"empresa_id"`
	EmpresaNombre          string  `json:"empresa_nombre"`
	EmpresaTipo            string  `json:"empresa_tipo"`
	Desde                  string  `json:"desde"`
	Hasta                  string  `json:"hasta"`
	UsuarioFiltro          string  `json:"usuario_filtro"`
	Moneda                 string  `json:"moneda"`
	AperturaEfectivo       float64 `json:"apertura_efectivo"`
	VentasCantidad         int64   `json:"ventas_cantidad"`
	VentasTotal            float64 `json:"ventas_total"`
	DevolucionesTotal      float64 `json:"devoluciones_total"`
	EfectivoVentas         float64 `json:"efectivo_ventas"`
	TarjetasVentas         float64 `json:"tarjetas_ventas"`
	TransferenciasVentas   float64 `json:"transferencias_ventas"`
	OtrosMediosVentas      float64 `json:"otros_medios_ventas"`
	IngresosEfectivo       float64 `json:"ingresos_efectivo"`
	EgresosEfectivo        float64 `json:"egresos_efectivo"`
	IngresosFinancieros    float64 `json:"ingresos_financieros"`
	EgresosFinancieros     float64 `json:"egresos_financieros"`
	EfectivoEsperadoCaja   float64 `json:"efectivo_esperado_caja"`
	TotalProductos         float64 `json:"total_productos"`
	TotalServicios         float64 `json:"total_servicios"`
	TotalOtrosItems        float64 `json:"total_otros_items"`
	HabitacionesSinFactura int64   `json:"habitaciones_sin_factura"`
}

type corteCajaResponse struct {
	OK                 bool                       `json:"ok"`
	Resumen            corteCajaResumen           `json:"resumen"`
	Ventas             []corteCajaVenta           `json:"ventas"`
	Movimientos        []corteCajaMovimientoGrupo `json:"movimientos"`
	ItemsPorTipo       []corteCajaTipoItem        `json:"items_por_tipo"`
	SensoresSinFactura []corteCajaSensorAlerta    `json:"sensores_sin_factura"`
	GeneradoEn         string                     `json:"generado_en"`
	Advertencias       []string                   `json:"advertencias,omitempty"`
}

func EmpresaCorteCajaHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		now := time.Now()
		desde := normalizeCorteCajaDateTime(r.URL.Query().Get("desde"), time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()))
		hasta := normalizeCorteCajaDateTime(r.URL.Query().Get("hasta"), now)
		usuario := strings.TrimSpace(r.URL.Query().Get("usuario"))
		apertura := parseCorteCajaFloat(r.URL.Query().Get("apertura_efectivo"))

		cacheKey := buildCorteCajaCacheKeyFromRequest(r, empresaID, desde, hasta, usuario, apertura)
		if cached := getCorteCajaCache(cacheKey); cached != nil {
			writeJSON(w, http.StatusOK, cached)
			return
		}

		resp, err := buildCorteCajaReport(dbEmp, empresaID, desde, hasta, usuario, apertura)
		if err != nil {
			http.Error(w, "No se pudo generar el corte de caja", http.StatusInternalServerError)
			return
		}
		storeCorteCajaCache(cacheKey, resp)
		writeJSON(w, http.StatusOK, resp)
	}
}

func buildCorteCajaCacheKeyFromRequest(r *http.Request, empresaID int64, desde, hasta, usuario string, apertura float64) string {
	rawDesde := ""
	rawHasta := ""
	if r != nil && r.URL != nil {
		rawDesde = strings.TrimSpace(r.URL.Query().Get("desde"))
		rawHasta = strings.TrimSpace(r.URL.Query().Get("hasta"))
	}
	keyDesde := strings.TrimSpace(desde)
	keyHasta := strings.TrimSpace(hasta)
	if rawDesde == "" {
		keyDesde = "default_desde"
	}
	if rawHasta == "" {
		keyHasta = "default_hasta"
	}
	return strings.Join([]string{
		strconv.FormatInt(empresaID, 10),
		keyDesde,
		keyHasta,
		strings.ToLower(strings.TrimSpace(usuario)),
		strconv.FormatFloat(apertura, 'f', 2, 64),
	}, "|")
}

func getCorteCajaCache(key string) *corteCajaResponse {
	if strings.TrimSpace(key) == "" {
		return nil
	}
	corteCajaCacheMu.Lock()
	defer corteCajaCacheMu.Unlock()
	entry, ok := corteCajaCache[key]
	if !ok {
		return nil
	}
	if entry.Response == nil || time.Since(entry.CachedAt) > corteCajaCacheTTL {
		delete(corteCajaCache, key)
		return nil
	}
	return entry.Response
}

func storeCorteCajaCache(key string, resp *corteCajaResponse) {
	if strings.TrimSpace(key) == "" || resp == nil {
		return
	}
	corteCajaCacheMu.Lock()
	corteCajaCache[key] = corteCajaCacheEntry{
		Response: resp,
		CachedAt: time.Now(),
	}
	corteCajaCacheMu.Unlock()
}

func normalizeCorteCajaDateTime(raw string, fallback time.Time) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return fallback.Format("2006-01-02 15:04:05")
	}
	value = strings.ReplaceAll(value, "T", " ")
	if len(value) == len("2006-01-02") {
		return value + " 00:00:00"
	}
	if len(value) == len("2006-01-02 15:04") {
		return value + ":00"
	}
	return value
}

func parseCorteCajaFloat(raw string) float64 {
	raw = strings.TrimSpace(strings.ReplaceAll(raw, ",", "."))
	if raw == "" {
		return 0
	}
	out, _ := strconv.ParseFloat(raw, 64)
	if out < 0 {
		return 0
	}
	return out
}

func buildCorteCajaReport(dbEmp *sql.DB, empresaID int64, desde, hasta, usuario string, apertura float64) (*corteCajaResponse, error) {
	if err := dbpkg.EnsureEmpresaCarritosSchema(dbEmp); err != nil {
		return nil, err
	}
	if err := dbpkg.EnsureEmpresaFinanzasSchema(dbEmp); err != nil {
		return nil, err
	}

	resp := &corteCajaResponse{
		OK:                 true,
		Ventas:             []corteCajaVenta{},
		Movimientos:        []corteCajaMovimientoGrupo{},
		ItemsPorTipo:       []corteCajaTipoItem{},
		SensoresSinFactura: []corteCajaSensorAlerta{},
		GeneradoEn:         time.Now().Format("2006-01-02 15:04:05"),
	}

	empresa, err := dbpkg.GetEmpresaByScopeID(dbEmp, empresaID)
	if err == nil && empresa != nil {
		resp.Resumen.EmpresaNombre = strings.TrimSpace(empresa.Nombre)
		resp.Resumen.EmpresaTipo = strings.TrimSpace(empresa.TipoNombre)
	}
	resp.Resumen.EmpresaID = empresaID
	resp.Resumen.Desde = desde
	resp.Resumen.Hasta = hasta
	resp.Resumen.UsuarioFiltro = usuario
	resp.Resumen.Moneda = "COP"
	resp.Resumen.AperturaEfectivo = apertura

	ventas, err := listCorteCajaVentas(dbEmp, empresaID, desde, hasta, usuario)
	if err != nil {
		return nil, err
	}
	resp.Ventas = ventas

	soldStationIDs := make(map[int64]bool)
	for _, venta := range ventas {
		resp.Resumen.VentasCantidad++
		resp.Resumen.VentasTotal += venta.TotalPagado
		if venta.TotalPagado <= 0 {
			resp.Resumen.VentasTotal += venta.Total
		}
		resp.Resumen.DevolucionesTotal += venta.Devolucion
		if strings.TrimSpace(venta.Moneda) != "" {
			resp.Resumen.Moneda = venta.Moneda
		}
		if venta.EstacionID > 0 {
			soldStationIDs[venta.EstacionID] = true
		}
		amount := venta.TotalPagado
		if amount <= 0 {
			amount = venta.Total
		}
		switch normalizeCorteCajaMetodo(venta.MetodoPago) {
		case "efectivo":
			resp.Resumen.EfectivoVentas += amount
		case "tarjeta":
			resp.Resumen.TarjetasVentas += amount
		case "transferencia":
			resp.Resumen.TransferenciasVentas += amount
		default:
			resp.Resumen.OtrosMediosVentas += amount
		}
	}

	items, err := listCorteCajaItemsPorTipo(dbEmp, empresaID, desde, hasta, usuario)
	if err != nil {
		return nil, err
	}
	resp.ItemsPorTipo = items
	for _, item := range items {
		switch normalizeCorteCajaTipoItem(item.Tipo) {
		case "producto", "combo":
			resp.Resumen.TotalProductos += item.Total
		case "servicio":
			resp.Resumen.TotalServicios += item.Total
		default:
			resp.Resumen.TotalOtrosItems += item.Total
		}
	}

	movimientos, err := listCorteCajaMovimientos(dbEmp, empresaID, desde, hasta, usuario)
	if err != nil {
		return nil, err
	}
	resp.Movimientos = movimientos
	for _, mov := range movimientos {
		tipo := strings.ToLower(strings.TrimSpace(mov.Tipo))
		metodo := normalizeCorteCajaMetodo(mov.Metodo)
		if tipo == "ingreso" {
			resp.Resumen.IngresosFinancieros += mov.Total
			if metodo == "efectivo" {
				resp.Resumen.IngresosEfectivo += mov.Total
			}
		}
		if tipo == "egreso" {
			resp.Resumen.EgresosFinancieros += mov.Total
			if metodo == "efectivo" {
				resp.Resumen.EgresosEfectivo += mov.Total
			}
		}
	}
	resp.Resumen.EfectivoEsperadoCaja = resp.Resumen.AperturaEfectivo + resp.Resumen.EfectivoVentas + resp.Resumen.IngresosEfectivo - resp.Resumen.EgresosEfectivo

	if strings.Contains(strings.ToLower(resp.Resumen.EmpresaTipo), "motel") {
		stationNames, _ := getCorteCajaStationNames(dbEmp, empresaID)
		alertas, alertErr := listCorteCajaSensoresSinFactura(dbEmp, empresaID, soldStationIDs, stationNames)
		if alertErr != nil {
			resp.Advertencias = append(resp.Advertencias, "No se pudieron evaluar sensores de puerta.")
		} else {
			resp.SensoresSinFactura = alertas
			resp.Resumen.HabitacionesSinFactura = int64(len(alertas))
		}
	}

	return resp, nil
}

func listCorteCajaVentas(dbEmp *sql.DB, empresaID int64, desde, hasta, usuario string) ([]corteCajaVenta, error) {
	query := `SELECT
		c.id,
		COALESCE(c.codigo, ''),
		COALESCE(c.nombre, ''),
		COALESCE(c.pagado_en, ''),
		COALESCE(c.metodo_pago, 'efectivo'),
		COALESCE(c.moneda, 'COP'),
		COALESCE(c.total, 0),
		COALESCE(c.total_pagado, 0),
		COALESCE(c.devolucion_total, 0),
		COALESCE(m.usuario_creador, c.usuario_creador, ''),
		COALESCE(m.estacion_id, 0),
		COALESCE(m.estacion_codigo, ''),
		COALESCE(m.estacion_nombre, '')
	FROM carritos_compras c
	LEFT JOIN (
		SELECT carrito_id,
		       MAX(COALESCE(usuario_creador, '')) AS usuario_creador,
		       MAX(COALESCE(estacion_id, 0)) AS estacion_id,
		       MAX(COALESCE(estacion_codigo, '')) AS estacion_codigo,
		       MAX(COALESCE(estacion_nombre, '')) AS estacion_nombre
		  FROM empresa_ventas_estacion_metricas
		 WHERE empresa_id = ?
		   AND LOWER(COALESCE(evento_operacion, '')) = 'venta_pagada'
		 GROUP BY carrito_id
	) m ON m.carrito_id = c.id
	WHERE c.empresa_id = ?
	  AND COALESCE(c.pagado_en, '') <> ''
	  AND COALESCE(c.pagado_en, '') >= ?
	  AND COALESCE(c.pagado_en, '') <= ?
	  AND LOWER(COALESCE(c.estado_carrito, '')) = 'cerrado'
	  AND (? = '' OR LOWER(COALESCE(m.usuario_creador, c.usuario_creador, '')) = LOWER(?))
	ORDER BY c.pagado_en ASC, c.id ASC`
	rows, err := dbpkg.ExecQueryCompat(dbEmp, query, empresaID, empresaID, desde, hasta, usuario, usuario)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []corteCajaVenta{}
	for rows.Next() {
		var item corteCajaVenta
		if err := rows.Scan(&item.ID, &item.Codigo, &item.Nombre, &item.FechaPago, &item.MetodoPago, &item.Moneda, &item.Total, &item.TotalPagado, &item.Devolucion, &item.Cajero, &item.EstacionID, &item.EstacionCodigo, &item.EstacionNombre); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func listCorteCajaItemsPorTipo(dbEmp *sql.DB, empresaID int64, desde, hasta, usuario string) ([]corteCajaTipoItem, error) {
	query := `SELECT
		LOWER(COALESCE(i.tipo_item, 'producto')) AS tipo,
		COALESCE(SUM(COALESCE(i.cantidad, 0)), 0) AS cantidad,
		COALESCE(SUM(COALESCE(i.total_linea, 0)), 0) AS total
	FROM carrito_compra_items i
	JOIN carritos_compras c ON c.id = i.carrito_id AND c.empresa_id = i.empresa_id
	LEFT JOIN (
		SELECT carrito_id, MAX(COALESCE(usuario_creador, '')) AS usuario_creador
		  FROM empresa_ventas_estacion_metricas
		 WHERE empresa_id = ?
		   AND LOWER(COALESCE(evento_operacion, '')) = 'venta_pagada'
		 GROUP BY carrito_id
	) m ON m.carrito_id = c.id
	WHERE c.empresa_id = ?
	  AND COALESCE(c.pagado_en, '') <> ''
	  AND COALESCE(c.pagado_en, '') >= ?
	  AND COALESCE(c.pagado_en, '') <= ?
	  AND LOWER(COALESCE(c.estado_carrito, '')) = 'cerrado'
	  AND LOWER(COALESCE(i.estado, 'activo')) = 'activo'
	  AND (? = '' OR LOWER(COALESCE(m.usuario_creador, c.usuario_creador, '')) = LOWER(?))
	GROUP BY LOWER(COALESCE(i.tipo_item, 'producto'))
	ORDER BY total DESC`
	rows, err := dbpkg.ExecQueryCompat(dbEmp, query, empresaID, empresaID, desde, hasta, usuario, usuario)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []corteCajaTipoItem{}
	for rows.Next() {
		var item corteCajaTipoItem
		if err := rows.Scan(&item.Tipo, &item.Cantidad, &item.Total); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func listCorteCajaMovimientos(dbEmp *sql.DB, empresaID int64, desde, hasta, usuario string) ([]corteCajaMovimientoGrupo, error) {
	query := `SELECT
		LOWER(COALESCE(tipo_movimiento, '')),
		LOWER(COALESCE(metodo_pago, '')),
		COALESCE(categoria, ''),
		COALESCE(usuario_creador, ''),
		COUNT(1),
		COALESCE(SUM(CASE WHEN COALESCE(total_neto, 0) <> 0 THEN total_neto ELSE total END), 0)
	FROM empresa_finanzas_movimientos
	WHERE empresa_id = ?
	  AND COALESCE(fecha_movimiento, fecha_creacion, '') >= ?
	  AND COALESCE(fecha_movimiento, fecha_creacion, '') <= ?
	  AND LOWER(COALESCE(estado, 'activo')) = 'activo'
	  AND (? = '' OR LOWER(COALESCE(usuario_creador, '')) = LOWER(?))
	GROUP BY LOWER(COALESCE(tipo_movimiento, '')), LOWER(COALESCE(metodo_pago, '')), COALESCE(categoria, ''), COALESCE(usuario_creador, '')
	ORDER BY 1, 2, 3, 4`
	rows, err := dbpkg.ExecQueryCompat(dbEmp, query, empresaID, desde, hasta, usuario, usuario)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []corteCajaMovimientoGrupo{}
	for rows.Next() {
		var item corteCajaMovimientoGrupo
		if err := rows.Scan(&item.Tipo, &item.Metodo, &item.Categoria, &item.Usuario, &item.Cantidad, &item.Total); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func listCorteCajaSensoresSinFactura(dbEmp *sql.DB, empresaID int64, soldStationIDs map[int64]bool, stationNames map[int64]string) ([]corteCajaSensorAlerta, error) {
	if err := dbpkg.EnsureEmpresaSensorPuertasSchema(dbEmp); err != nil {
		return nil, err
	}
	rows, err := dbpkg.ExecQueryCompat(dbEmp, `SELECT COALESCE(estacion_id, 0), COALESCE(device_id, ''), COALESCE(last_state, ''), COALESCE(last_seen, '') FROM empresa_sensor_puertas_devices WHERE empresa_id = ? AND LOWER(COALESCE(estado, 'activo')) = 'activo' AND COALESCE(estacion_id, 0) > 0`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []corteCajaSensorAlerta{}
	for rows.Next() {
		var estacionID int64
		var deviceID, state, seen string
		if err := rows.Scan(&estacionID, &deviceID, &state, &seen); err != nil {
			return nil, err
		}
		if !isCorteCajaSensorOccupiedState(state) || soldStationIDs[estacionID] {
			continue
		}
		name := strings.TrimSpace(stationNames[estacionID])
		if name == "" {
			name = "Habitacion " + strconv.FormatInt(estacionID, 10)
		}
		out = append(out, corteCajaSensorAlerta{
			EstacionID:     estacionID,
			EstacionNombre: name,
			DeviceID:       deviceID,
			EstadoSensor:   state,
			UltimaLectura:  seen,
			Motivo:         "El sensor figura ocupado, pero no hay venta facturada para esta habitacion en el rango del corte.",
		})
	}
	return out, rows.Err()
}

func getCorteCajaStationNames(dbEmp *sql.DB, empresaID int64) (map[int64]string, error) {
	out := map[int64]string{}
	rows, err := dbpkg.ExecQueryCompat(dbEmp, `SELECT COALESCE(valor, '') FROM empresa_estacion_prefs WHERE empresa_id = ? AND estacion_id = 0 AND clave = 'estaciones_config' AND LOWER(COALESCE(estado, 'activo')) = 'activo' LIMIT 1`, empresaID)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	if !rows.Next() {
		return out, rows.Err()
	}
	var raw string
	if err := rows.Scan(&raw); err != nil {
		return out, err
	}
	var cfg struct {
		Estaciones []struct {
			ID     int64  `json:"id"`
			Nombre string `json:"nombre"`
		} `json:"estaciones"`
	}
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		var decoded string
		if err2 := json.Unmarshal([]byte(raw), &decoded); err2 == nil {
			_ = json.Unmarshal([]byte(decoded), &cfg)
		}
	}
	for _, st := range cfg.Estaciones {
		if st.ID > 0 && strings.TrimSpace(st.Nombre) != "" {
			out[st.ID] = strings.TrimSpace(st.Nombre)
		}
	}
	return out, nil
}

func normalizeCorteCajaMetodo(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "efectivo", "cash":
		return "efectivo"
	case "tarjeta", "tarjeta_credito", "tarjeta_debito", "credito", "debito", "card":
		return "tarjeta"
	case "transferencia", "transferencia_bancaria", "banco":
		return "transferencia"
	default:
		return v
	}
}

func normalizeCorteCajaTipoItem(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func isCorteCajaSensorOccupiedState(raw string) bool {
	v := strings.ToLower(strings.TrimSpace(raw))
	if v == "" {
		return false
	}
	freeStates := map[string]bool{
		"0": true, "false": true, "off": true, "libre": true, "disponible": true,
		"free": true, "abierta": true, "abierto": true, "open": true,
		"vacia": true, "vacio": true, "vacía": true, "vacío": true,
	}
	if freeStates[v] {
		return false
	}
	occupiedMarkers := []string{"ocup", "occupied", "busy", "cerrad", "closed", "on", "true", "1", "presencia"}
	for _, marker := range occupiedMarkers {
		if strings.Contains(v, marker) {
			return true
		}
	}
	return false
}
