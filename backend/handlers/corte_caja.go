package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
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

type corteCajaProductoVendido struct {
	Fecha       string  `json:"fecha"`
	Producto    string  `json:"producto"`
	Cantidad    float64 `json:"cantidad"`
	Tipo        string  `json:"tipo"`
	Referencia  string  `json:"referencia"`
	VentaCodigo string  `json:"venta_codigo"`
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
	NumeroFacturas         int64   `json:"numero_facturas"`
	VentasTotal            float64 `json:"ventas_total"`
	VentasAnuladasCantidad int64   `json:"ventas_anuladas_cantidad"`
	VentasAnuladasTotal    float64 `json:"ventas_anuladas_total"`
	DevolucionesTotal      float64 `json:"devoluciones_total"`
	EfectivoVentas         float64 `json:"efectivo_ventas"`
	DebitoVentas           float64 `json:"debito_ventas"`
	CreditoVentas          float64 `json:"credito_ventas"`
	TarjetasVentas         float64 `json:"tarjetas_ventas"`
	TransferenciasVentas   float64 `json:"transferencias_ventas"`
	OtrosMediosVentas      float64 `json:"otros_medios_ventas"`
	IngresosCantidad       int64   `json:"ingresos_cantidad"`
	EgresosCantidad        int64   `json:"egresos_cantidad"`
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
	OK                 bool                                 `json:"ok"`
	Resumen            corteCajaResumen                     `json:"resumen"`
	Ventas             []corteCajaVenta                     `json:"ventas"`
	Anulaciones        []corteCajaVenta                     `json:"anulaciones"`
	Movimientos        []corteCajaMovimientoGrupo           `json:"movimientos"`
	ItemsPorTipo       []corteCajaTipoItem                  `json:"items_por_tipo"`
	ProductosVendidos  []corteCajaProductoVendido           `json:"productos_vendidos"`
	SensoresSinFactura []corteCajaSensorAlerta              `json:"sensores_sin_factura"`
	Reportes           []string                             `json:"reportes"`
	Secciones          map[string]bool                      `json:"secciones"`
	CajaActual         *corteCajaActualContext              `json:"caja_actual,omitempty"`
	Configuracion      *dbpkg.EmpresaCorteCajaConfiguracion `json:"configuracion_reporte,omitempty"`
	GeneradoEn         string                               `json:"generado_en"`
	Advertencias       []string                             `json:"advertencias,omitempty"`
}

type corteCajaActualContext struct {
	ID              int64  `json:"id"`
	CajaCodigo      string `json:"caja_codigo"`
	Turno           string `json:"turno"`
	SucursalID      int64  `json:"sucursal_id"`
	Usuario         string `json:"usuario"`
	FechaApertura   string `json:"fecha_apertura"`
	FechaOperacion  string `json:"fecha_operacion"`
	EstadoCierre    string `json:"estado_cierre"`
	SoloUsuarioCaja bool   `json:"solo_usuario_caja"`
}

type corteCajaCerrarPayload struct {
	Desde            string   `json:"desde"`
	Hasta            string   `json:"hasta"`
	Usuario          string   `json:"usuario"`
	AperturaEfectivo float64  `json:"apertura_efectivo"`
	CajaFisica       float64  `json:"caja_fisica"`
	CierreCajaID     int64    `json:"cierre_caja_id"`
	CajaCodigo       string   `json:"caja_codigo"`
	Turno            string   `json:"turno"`
	Reportes         []string `json:"reportes"`
	ImprimirAuto     bool     `json:"imprimir_auto"`
	Observaciones    string   `json:"observaciones"`
	UmbralIncidencia float64  `json:"umbral_incidencia"`
	UsuarioOperacion string   `json:"usuario_operacion"`
	FechaOperacion   string   `json:"fecha_operacion"`
}

func EmpresaCorteCajaHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
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
		configuracion, configErr := dbpkg.GetEmpresaCorteCajaConfiguracion(dbEmp, empresaID)
		if configErr != nil {
			http.Error(w, "No se pudo cargar la configuracion del corte de caja", http.StatusInternalServerError)
			return
		}
		reportes := parseCorteCajaReportesWithConfig(r.URL.Query().Get("reportes"), configuracion)

		if r.Method == http.MethodPost {
			var payload corteCajaCerrarPayload
			if r.Body != nil {
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && err != io.EOF {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
			}
			if strings.TrimSpace(payload.Desde) != "" {
				desde = normalizeCorteCajaDateTime(payload.Desde, now)
			}
			if strings.TrimSpace(payload.Hasta) != "" {
				hasta = normalizeCorteCajaDateTime(payload.Hasta, now)
			}
			if strings.TrimSpace(payload.Usuario) != "" {
				usuario = strings.TrimSpace(payload.Usuario)
			}
			if payload.AperturaEfectivo >= 0 {
				apertura = payload.AperturaEfectivo
			}
			if len(payload.Reportes) > 0 {
				reportes = normalizeCorteCajaReportes(payload.Reportes)
			}
			resp, cierreID, err := cerrarCorteCaja(dbEmp, empresaID, desde, hasta, usuario, apertura, payload, reportes, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			resp.Configuracion = configuracion
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":            true,
				"cierre_id":     cierreID,
				"reporte":       resp,
				"imprimir_auto": payload.ImprimirAuto,
				"reportes":      reportes,
			})
			return
		}

		if corteCajaSoloUsuarioCajaActual(r) {
			resp, err := buildCorteCajaUsuarioCajaActualReport(dbEmp, r, empresaID, desde, hasta, apertura)
			if err != nil {
				http.Error(w, "No se pudo generar el corte de caja del usuario actual", http.StatusInternalServerError)
				return
			}
			resp.Configuracion = configuracion
			applyCorteCajaReportSelection(resp, reportes)
			writeJSON(w, http.StatusOK, resp)
			return
		}

		cacheKey := buildCorteCajaCacheKeyFromRequest(r, empresaID, desde, hasta, usuario, apertura, reportes, configuracion)
		if cached := getCorteCajaCache(cacheKey); cached != nil {
			writeJSON(w, http.StatusOK, cached)
			return
		}

		resp, err := buildCorteCajaReport(dbEmp, empresaID, desde, hasta, usuario, apertura, strings.TrimSpace(r.URL.Query().Get("caja_codigo")), parseCorteCajaInt64(r.URL.Query().Get("cierre_caja_id")))
		if err != nil {
			http.Error(w, "No se pudo generar el corte de caja", http.StatusInternalServerError)
			return
		}
		resp.Configuracion = configuracion
		applyCorteCajaReportSelection(resp, reportes)
		storeCorteCajaCache(cacheKey, resp)
		writeJSON(w, http.StatusOK, resp)
	}
}

func EmpresaCorteCajaConfiguracionHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			cfg, err := dbpkg.GetEmpresaCorteCajaConfiguracion(dbEmp, empresaID)
			if err != nil {
				http.Error(w, "No se pudo cargar la configuracion del reporte", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":            true,
				"configuracion": cfg,
				"reportes":      dbpkg.EmpresaCorteCajaReportesDesdeConfiguracion(cfg),
			})
			return

		case http.MethodPost, http.MethodPut:
			if strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("action")), "restaurar") {
				cfg := dbpkg.DefaultEmpresaCorteCajaConfiguracion(empresaID)
				cfg.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				if _, err := dbpkg.UpsertEmpresaCorteCajaConfiguracion(dbEmp, cfg); err != nil {
					http.Error(w, "No se pudo restaurar la configuracion del reporte", http.StatusInternalServerError)
					return
				}
				saved, _ := dbpkg.GetEmpresaCorteCajaConfiguracion(dbEmp, empresaID)
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":            true,
					"configuracion": saved,
					"reportes":      dbpkg.EmpresaCorteCajaReportesDesdeConfiguracion(saved),
				})
				return
			}

			var payload dbpkg.EmpresaCorteCajaConfiguracion
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			payload.EmpresaID = empresaID
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			if payload.Estado == "" {
				payload.Estado = "activo"
			}
			if _, err := dbpkg.UpsertEmpresaCorteCajaConfiguracion(dbEmp, payload); err != nil {
				http.Error(w, "No se pudo guardar la configuracion del reporte", http.StatusInternalServerError)
				return
			}
			saved, _ := dbpkg.GetEmpresaCorteCajaConfiguracion(dbEmp, empresaID)
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":            true,
				"configuracion": saved,
				"reportes":      dbpkg.EmpresaCorteCajaReportesDesdeConfiguracion(saved),
			})
			return

		default:
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
	}
}

func buildCorteCajaCacheKeyFromRequest(r *http.Request, empresaID int64, desde, hasta, usuario string, apertura float64, reportes []string, cfg *dbpkg.EmpresaCorteCajaConfiguracion) string {
	rawDesde := ""
	rawHasta := ""
	rawCajaCodigo := ""
	rawCierreCajaID := ""
	if r != nil && r.URL != nil {
		rawDesde = strings.TrimSpace(r.URL.Query().Get("desde"))
		rawHasta = strings.TrimSpace(r.URL.Query().Get("hasta"))
		rawCajaCodigo = strings.TrimSpace(r.URL.Query().Get("caja_codigo"))
		rawCierreCajaID = strings.TrimSpace(r.URL.Query().Get("cierre_caja_id"))
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
		strings.ToUpper(rawCajaCodigo),
		rawCierreCajaID,
		strings.Join(normalizeCorteCajaReportes(reportes), ","),
		corteCajaConfigCacheFingerprint(cfg),
	}, "|")
}

func corteCajaConfigCacheFingerprint(cfg *dbpkg.EmpresaCorteCajaConfiguracion) string {
	if cfg == nil {
		return "default"
	}
	values := []bool{
		cfg.MostrarResumen, cfg.MostrarNumeroFacturas, cfg.MostrarTotalVentas,
		cfg.MostrarEfectivo, cfg.MostrarDebito, cfg.MostrarCredito, cfg.MostrarTransferencias, cfg.MostrarOtrosMedios,
		cfg.MostrarIngresos, cfg.MostrarEgresos, cfg.MostrarAnulaciones, cfg.MostrarDevoluciones,
		cfg.MostrarCajaEsperada, cfg.MostrarDiferenciaCaja, cfg.MostrarVentasDetalle, cfg.MostrarMovimientos,
		cfg.MostrarItems, cfg.MostrarSensoresPuertas, cfg.MostrarAuditoria,
	}
	var b strings.Builder
	for _, value := range values {
		if value {
			b.WriteByte('1')
		} else {
			b.WriteByte('0')
		}
	}
	b.WriteByte(':')
	b.WriteString(strings.ToLower(strings.TrimSpace(cfg.FormatoImpresion)))
	return b.String()
}

func corteCajaSoloUsuarioCajaActual(r *http.Request) bool {
	if r == nil {
		return false
	}
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	if action == "mi_caja_actual" || action == "mis_movimientos_caja" || action == "ultimos_movimientos" {
		return true
	}
	return queryBool(r, "solo_usuario_actual") || queryBool(r, "mi_caja_actual")
}

func buildCorteCajaUsuarioCajaActualReport(dbEmp *sql.DB, r *http.Request, empresaID int64, desde, hasta string, apertura float64) (*corteCajaResponse, error) {
	usuarioActual := strings.TrimSpace(adminEmailFromRequest(r))
	cajaCodigo := strings.TrimSpace(r.URL.Query().Get("caja_codigo"))
	turno := strings.TrimSpace(r.URL.Query().Get("turno"))
	sucursalID := parseCorteCajaInt64(r.URL.Query().Get("sucursal_id"))
	cierreCajaID := parseCorteCajaInt64(r.URL.Query().Get("cierre_caja_id"))

	cierre, err := getCorteCajaAbiertaUsuarioActual(dbEmp, empresaID, usuarioActual, cajaCodigo, turno, sucursalID, cierreCajaID)
	if err != nil {
		if err == sql.ErrNoRows {
			resp := emptyCorteCajaUsuarioCajaActualResponse(empresaID, desde, hasta, usuarioActual, cajaCodigo)
			resp.Advertencias = append(resp.Advertencias, "No hay una caja abierta para el usuario actual con el filtro solicitado.")
			return resp, nil
		}
		return nil, err
	}

	cajaCodigo = strings.TrimSpace(cierre.CajaCodigo)
	if strings.TrimSpace(cierre.FechaApertura) != "" {
		desde = strings.TrimSpace(cierre.FechaApertura)
	} else if strings.TrimSpace(cierre.FechaOperacion) != "" {
		desde = strings.TrimSpace(cierre.FechaOperacion) + " 00:00:00"
	}
	if strings.TrimSpace(hasta) == "" {
		hasta = time.Now().Format("2006-01-02 15:04:05")
	}
	apertura = cierre.AperturaMonto

	resp, err := buildCorteCajaReport(dbEmp, empresaID, desde, hasta, usuarioActual, apertura, cajaCodigo, cierre.ID)
	if err != nil {
		return nil, err
	}
	resp.CajaActual = &corteCajaActualContext{
		ID:              cierre.ID,
		CajaCodigo:      cajaCodigo,
		Turno:           strings.TrimSpace(cierre.Turno),
		SucursalID:      cierre.SucursalID,
		Usuario:         usuarioActual,
		FechaApertura:   strings.TrimSpace(cierre.FechaApertura),
		FechaOperacion:  strings.TrimSpace(cierre.FechaOperacion),
		EstadoCierre:    strings.TrimSpace(cierre.EstadoCierre),
		SoloUsuarioCaja: true,
	}
	return resp, nil
}

func emptyCorteCajaUsuarioCajaActualResponse(empresaID int64, desde, hasta, usuario, cajaCodigo string) *corteCajaResponse {
	return &corteCajaResponse{
		OK:                true,
		Ventas:            []corteCajaVenta{},
		Anulaciones:       []corteCajaVenta{},
		Movimientos:       []corteCajaMovimientoGrupo{},
		ItemsPorTipo:      []corteCajaTipoItem{},
		ProductosVendidos: []corteCajaProductoVendido{},
		GeneradoEn:        time.Now().Format("2006-01-02 15:04:05"),
		Resumen: corteCajaResumen{
			EmpresaID:     empresaID,
			Desde:         desde,
			Hasta:         hasta,
			UsuarioFiltro: usuario,
			Moneda:        "COP",
		},
		CajaActual: &corteCajaActualContext{
			CajaCodigo:      strings.TrimSpace(cajaCodigo),
			Usuario:         strings.TrimSpace(usuario),
			SoloUsuarioCaja: true,
		},
	}
}

func getCorteCajaAbiertaUsuarioActual(dbEmp *sql.DB, empresaID int64, usuario, cajaCodigo, turno string, sucursalID, cierreCajaID int64) (*dbpkg.EmpresaCierreCaja, error) {
	if err := dbpkg.EnsureEmpresaFinanzasSchema(dbEmp); err != nil {
		return nil, err
	}
	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = "sistema"
	}
	query := `SELECT
		id,
		empresa_id,
		COALESCE(sucursal_id, 0),
		COALESCE(caja_codigo, ''),
		COALESCE(turno, 'general'),
		COALESCE(fecha_operacion, ''),
		COALESCE(fecha_apertura, ''),
		COALESCE(fecha_cierre, ''),
		COALESCE(estado_cierre, 'abierto'),
		COALESCE(apertura_monto, 0),
		COALESCE(ingresos_efectivo, 0),
		COALESCE(egresos_efectivo, 0),
		COALESCE(retiros_efectivo, 0),
		COALESCE(caja_teorica, 0),
		COALESCE(caja_fisica, 0),
		COALESCE(diferencia_caja, 0),
		COALESCE(moneda, 'COP'),
		COALESCE(cerrado_por, ''),
		COALESCE(aprobado_por, ''),
		COALESCE(aprobado_en, ''),
		COALESCE(tiene_incidencia, 0),
		COALESCE(umbral_incidencia, 0),
		COALESCE(propinas_movimientos, 0),
		COALESCE(propinas_total, 0),
		COALESCE(propinas_ajustes, 0),
		COALESCE(propinas_impuesto, 0),
		COALESCE(propinas_neto, 0),
		COALESCE(propinas_conciliado_en, ''),
		COALESCE(propinas_conciliado_por, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_cierres_caja
	WHERE empresa_id = ?
	  AND LOWER(COALESCE(estado_cierre, 'abierto')) = 'abierto'
	  AND LOWER(COALESCE(estado, 'activo')) = 'activo'
	  AND LOWER(COALESCE(usuario_creador, '')) = LOWER(?)`
	args := []interface{}{empresaID, usuario}
	if cierreCajaID > 0 {
		query += ` AND id = ?`
		args = append(args, cierreCajaID)
	} else {
		cajaCodigo = strings.ToUpper(strings.TrimSpace(cajaCodigo))
		if cajaCodigo != "" {
			query += ` AND UPPER(COALESCE(caja_codigo, '')) = ?`
			args = append(args, cajaCodigo)
		}
		if strings.TrimSpace(turno) != "" {
			query += ` AND LOWER(COALESCE(turno, 'general')) = ?`
			args = append(args, strings.ToLower(strings.TrimSpace(turno)))
		}
		if sucursalID > 0 {
			query += ` AND COALESCE(sucursal_id, 0) = ?`
			args = append(args, sucursalID)
		}
	}
	query += ` ORDER BY COALESCE(fecha_apertura, fecha_creacion, '') DESC, id DESC LIMIT 1`

	rows, err := dbpkg.ExecQueryCompat(dbEmp, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	var item dbpkg.EmpresaCierreCaja
	var incidencia int
	if err := rows.Scan(
		&item.ID,
		&item.EmpresaID,
		&item.SucursalID,
		&item.CajaCodigo,
		&item.Turno,
		&item.FechaOperacion,
		&item.FechaApertura,
		&item.FechaCierre,
		&item.EstadoCierre,
		&item.AperturaMonto,
		&item.IngresosEfectivo,
		&item.EgresosEfectivo,
		&item.RetirosEfectivo,
		&item.CajaTeorica,
		&item.CajaFisica,
		&item.DiferenciaCaja,
		&item.Moneda,
		&item.CerradoPor,
		&item.AprobadoPor,
		&item.AprobadoEn,
		&incidencia,
		&item.UmbralIncidencia,
		&item.PropinasMovimientos,
		&item.PropinasTotal,
		&item.PropinasAjustes,
		&item.PropinasImpuesto,
		&item.PropinasNeto,
		&item.PropinasConciliadoEn,
		&item.PropinasConciliadoPor,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	); err != nil {
		return nil, err
	}
	item.TieneIncidencia = incidencia == 1
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &item, nil
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

func parseCorteCajaReportes(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return defaultCorteCajaReportes()
	}
	return normalizeCorteCajaReportes(strings.Split(raw, ","))
}

func parseCorteCajaReportesWithConfig(raw string, cfg *dbpkg.EmpresaCorteCajaConfiguracion) []string {
	if strings.TrimSpace(raw) == "" {
		return normalizeCorteCajaReportes(dbpkg.EmpresaCorteCajaReportesDesdeConfiguracion(cfg))
	}
	return normalizeCorteCajaReportes(strings.Split(raw, ","))
}

func normalizeCorteCajaReportes(items []string) []string {
	allowed := map[string]string{
		"resumen":       "resumen",
		"ejecutivo":     "resumen",
		"movimientos":   "movimientos",
		"ventas":        "ventas",
		"anulaciones":   "anulaciones",
		"items":         "items",
		"productos":     "items",
		"servicios":     "items",
		"sensores":      "sensores",
		"habitaciones":  "sensores",
		"auditoria":     "auditoria",
		"observaciones": "auditoria",
	}
	seen := map[string]bool{}
	out := []string{}
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item))
		if normalized, ok := allowed[key]; ok && !seen[normalized] {
			seen[normalized] = true
			out = append(out, normalized)
		}
	}
	if len(out) == 0 {
		return defaultCorteCajaReportes()
	}
	return out
}

func defaultCorteCajaReportes() []string {
	return []string{"resumen", "movimientos", "ventas", "anulaciones", "items", "sensores", "auditoria"}
}

func applyCorteCajaReportSelection(resp *corteCajaResponse, reportes []string) {
	if resp == nil {
		return
	}
	reportes = normalizeCorteCajaReportes(reportes)
	resp.Reportes = reportes
	resp.Secciones = map[string]bool{}
	for _, reporte := range reportes {
		resp.Secciones[reporte] = true
	}
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

func parseCorteCajaInt64(raw string) int64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || v < 0 {
		return 0
	}
	return v
}

func cerrarCorteCaja(dbEmp *sql.DB, empresaID int64, desde, hasta, usuario string, apertura float64, payload corteCajaCerrarPayload, reportes []string, r *http.Request) (*corteCajaResponse, int64, error) {
	if dbEmp == nil {
		return nil, 0, sql.ErrConnDone
	}
	if err := dbpkg.EnsureEmpresaFinanzasSchema(dbEmp); err != nil {
		return nil, 0, err
	}
	reportes = normalizeCorteCajaReportes(reportes)
	resp, err := buildCorteCajaReport(dbEmp, empresaID, desde, hasta, usuario, apertura, payload.CajaCodigo, payload.CierreCajaID)
	if err != nil {
		return nil, 0, err
	}
	applyCorteCajaReportSelection(resp, reportes)

	cajaCodigo := strings.TrimSpace(payload.CajaCodigo)
	if cajaCodigo == "" {
		cajaCodigo = "CAJA-PRINCIPAL"
	}
	turno := strings.TrimSpace(payload.Turno)
	if turno == "" {
		turno = "general"
	}
	usuarioOperacion := strings.TrimSpace(payload.UsuarioOperacion)
	if usuarioOperacion == "" && r != nil {
		usuarioOperacion = strings.TrimSpace(adminEmailFromRequest(r))
	}
	if usuarioOperacion == "" {
		usuarioOperacion = "sistema"
	}
	fechaOperacion := strings.TrimSpace(payload.FechaOperacion)
	if fechaOperacion == "" {
		fechaOperacion = strings.TrimSpace(desde)
		if len(fechaOperacion) >= 10 {
			fechaOperacion = fechaOperacion[:10]
		}
	}
	if fechaOperacion == "" {
		fechaOperacion = time.Now().Format("2006-01-02")
	}

	meta, _ := json.Marshal(map[string]interface{}{
		"origen":                 "corte_de_caja",
		"reportes":               reportes,
		"ventas_cantidad":        resp.Resumen.VentasCantidad,
		"ventas_total":           resp.Resumen.VentasTotal,
		"anulaciones_cantidad":   resp.Resumen.VentasAnuladasCantidad,
		"anulaciones_total":      resp.Resumen.VentasAnuladasTotal,
		"habitaciones_alertadas": resp.Resumen.HabitacionesSinFactura,
		"observaciones":          strings.TrimSpace(payload.Observaciones),
	})
	ingresosEfectivo := resp.Resumen.EfectivoVentas + resp.Resumen.IngresosEfectivo
	egresosEfectivo := resp.Resumen.EgresosEfectivo
	if payload.CierreCajaID > 0 {
		cajaFisica := payload.CajaFisica
		if err := dbpkg.SetEmpresaCierreCajaEstado(dbEmp, empresaID, payload.CierreCajaID, "cerrado", &cajaFisica, usuarioOperacion, string(meta)); err != nil {
			return nil, 0, err
		}
		return resp, payload.CierreCajaID, nil
	}
	cierreID, err := dbpkg.CreateEmpresaCierreCaja(dbEmp, dbpkg.EmpresaCierreCaja{
		EmpresaID:        empresaID,
		CajaCodigo:       cajaCodigo,
		Turno:            turno,
		FechaOperacion:   fechaOperacion,
		FechaApertura:    desde,
		FechaCierre:      hasta,
		EstadoCierre:     "cerrado",
		AperturaMonto:    apertura,
		IngresosEfectivo: ingresosEfectivo,
		EgresosEfectivo:  egresosEfectivo,
		RetirosEfectivo:  0,
		CajaFisica:       payload.CajaFisica,
		Moneda:           resp.Resumen.Moneda,
		CerradoPor:       usuarioOperacion,
		UsuarioCreador:   usuarioOperacion,
		Estado:           "activo",
		Observaciones:    string(meta),
		UmbralIncidencia: payload.UmbralIncidencia,
	})
	if err != nil {
		return nil, 0, err
	}
	return resp, cierreID, nil
}

func buildCorteCajaReport(dbEmp *sql.DB, empresaID int64, desde, hasta, usuario string, apertura float64, cajaCodigo string, cierreCajaID int64) (*corteCajaResponse, error) {
	startedAt := time.Now()
	defer func() {
		dbpkg.PerfLogf("[perf][corte] buildCorteCajaReport empresa=%d usuario=%q dur=%s", empresaID, usuario, time.Since(startedAt))
	}()
	if err := dbpkg.EnsureEmpresaCarritosSchema(dbEmp); err != nil {
		return nil, err
	}
	dbpkg.PerfLogf("[perf][corte] step ensure_carritos empresa=%d dur=%s", empresaID, time.Since(startedAt))
	if err := dbpkg.EnsureEmpresaFinanzasSchema(dbEmp); err != nil {
		return nil, err
	}
	dbpkg.PerfLogf("[perf][corte] step ensure_finanzas empresa=%d dur=%s", empresaID, time.Since(startedAt))

	resp := &corteCajaResponse{
		OK:                 true,
		Ventas:             []corteCajaVenta{},
		Anulaciones:        []corteCajaVenta{},
		Movimientos:        []corteCajaMovimientoGrupo{},
		ItemsPorTipo:       []corteCajaTipoItem{},
		ProductosVendidos:  []corteCajaProductoVendido{},
		SensoresSinFactura: []corteCajaSensorAlerta{},
		GeneradoEn:         time.Now().Format("2006-01-02 15:04:05"),
	}

	empresa, err := dbpkg.GetEmpresaByScopeID(dbEmp, empresaID)
	if err == nil && empresa != nil {
		resp.Resumen.EmpresaNombre = strings.TrimSpace(empresa.Nombre)
		resp.Resumen.EmpresaTipo = strings.TrimSpace(empresa.TipoNombre)
	}
	dbpkg.PerfLogf("[perf][corte] step empresa empresa=%d dur=%s", empresaID, time.Since(startedAt))
	resp.Resumen.EmpresaID = empresaID
	resp.Resumen.Desde = desde
	resp.Resumen.Hasta = hasta
	resp.Resumen.UsuarioFiltro = usuario
	resp.Resumen.Moneda = "COP"
	resp.Resumen.AperturaEfectivo = apertura

	ventas, err := listCorteCajaVentas(dbEmp, empresaID, desde, hasta, usuario, cajaCodigo, cierreCajaID)
	if err != nil {
		return nil, err
	}
	dbpkg.PerfLogf("[perf][corte] step ventas empresa=%d dur=%s", empresaID, time.Since(startedAt))
	resp.Ventas = ventas

	soldStationIDs := make(map[int64]bool)
	for _, venta := range ventas {
		resp.Resumen.VentasCantidad++
		resp.Resumen.NumeroFacturas++
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
		switch normalizeCorteCajaMetodoDetalle(venta.MetodoPago) {
		case "efectivo":
			resp.Resumen.EfectivoVentas += amount
		case "tarjeta_credito":
			resp.Resumen.CreditoVentas += amount
			resp.Resumen.TarjetasVentas += amount
		case "tarjeta_debito":
			resp.Resumen.DebitoVentas += amount
			resp.Resumen.TarjetasVentas += amount
		case "tarjeta":
			resp.Resumen.TarjetasVentas += amount
		case "transferencia":
			resp.Resumen.TransferenciasVentas += amount
		default:
			resp.Resumen.OtrosMediosVentas += amount
		}
	}

	anulaciones, err := listCorteCajaVentasAnuladas(dbEmp, empresaID, desde, hasta, usuario, cajaCodigo, cierreCajaID)
	if err != nil {
		return nil, err
	}
	dbpkg.PerfLogf("[perf][corte] step anulaciones empresa=%d dur=%s", empresaID, time.Since(startedAt))
	resp.Anulaciones = anulaciones
	for _, anulacion := range anulaciones {
		resp.Resumen.VentasAnuladasCantidad++
		amount := anulacion.Devolucion
		if amount <= 0 {
			amount = anulacion.TotalPagado
		}
		if amount <= 0 {
			amount = anulacion.Total
		}
		resp.Resumen.VentasAnuladasTotal += amount
	}

	items, err := listCorteCajaItemsPorTipo(dbEmp, empresaID, desde, hasta, usuario, cajaCodigo, cierreCajaID)
	if err != nil {
		return nil, err
	}
	dbpkg.PerfLogf("[perf][corte] step items empresa=%d dur=%s", empresaID, time.Since(startedAt))
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

	productosVendidos, err := listCorteCajaProductosVendidos(dbEmp, empresaID, desde, hasta, usuario, cajaCodigo, cierreCajaID, 50)
	if err != nil {
		return nil, err
	}
	resp.ProductosVendidos = productosVendidos

	movimientos, err := listCorteCajaMovimientos(dbEmp, empresaID, desde, hasta, usuario, cajaCodigo, cierreCajaID)
	if err != nil {
		return nil, err
	}
	dbpkg.PerfLogf("[perf][corte] step movimientos empresa=%d dur=%s", empresaID, time.Since(startedAt))
	resp.Movimientos = movimientos
	for _, mov := range movimientos {
		tipo := strings.ToLower(strings.TrimSpace(mov.Tipo))
		metodo := normalizeCorteCajaMetodo(mov.Metodo)
		if tipo == "ingreso" {
			resp.Resumen.IngresosCantidad += mov.Cantidad
			resp.Resumen.IngresosFinancieros += mov.Total
			if metodo == "efectivo" {
				resp.Resumen.IngresosEfectivo += mov.Total
			}
		}
		if tipo == "egreso" {
			resp.Resumen.EgresosCantidad += mov.Cantidad
			resp.Resumen.EgresosFinancieros += mov.Total
			if metodo == "efectivo" {
				resp.Resumen.EgresosEfectivo += mov.Total
			}
		}
	}
	resp.Resumen.EfectivoEsperadoCaja = resp.Resumen.AperturaEfectivo + resp.Resumen.EfectivoVentas + resp.Resumen.IngresosEfectivo - resp.Resumen.EgresosEfectivo

	if strings.Contains(strings.ToLower(resp.Resumen.EmpresaTipo), "motel") {
		stationNames, _ := getCorteCajaStationNames(dbEmp, empresaID)
		dbpkg.PerfLogf("[perf][corte] step station_names empresa=%d dur=%s", empresaID, time.Since(startedAt))
		alertas, alertErr := listCorteCajaSensoresSinFactura(dbEmp, empresaID, soldStationIDs, stationNames)
		if alertErr != nil {
			resp.Advertencias = append(resp.Advertencias, "No se pudieron evaluar sensores de puerta.")
		} else {
			resp.SensoresSinFactura = alertas
			resp.Resumen.HabitacionesSinFactura = int64(len(alertas))
		}
		dbpkg.PerfLogf("[perf][corte] step sensores empresa=%d dur=%s", empresaID, time.Since(startedAt))
	}

	return resp, nil
}

func listCorteCajaVentas(dbEmp *sql.DB, empresaID int64, desde, hasta, usuario, cajaCodigo string, cierreCajaID int64) ([]corteCajaVenta, error) {
	startedAt := time.Now()
	defer func() {
		dbpkg.PerfLogf("[perf][corte] listCorteCajaVentas empresa=%d usuario=%q dur=%s", empresaID, usuario, time.Since(startedAt))
	}()
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
	  AND (? = 0 OR COALESCE(c.cierre_caja_id, 0) = ?)
	  AND (? = '' OR UPPER(COALESCE(c.caja_codigo, '')) = ?)
	ORDER BY c.pagado_en ASC, c.id ASC`
	cajaCodigo = strings.ToUpper(strings.TrimSpace(cajaCodigo))
	rows, err := dbpkg.ExecQueryCompat(dbEmp, query, empresaID, empresaID, desde, hasta, usuario, usuario, cierreCajaID, cierreCajaID, cajaCodigo, cajaCodigo)
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

func listCorteCajaVentasAnuladas(dbEmp *sql.DB, empresaID int64, desde, hasta, usuario, cajaCodigo string, cierreCajaID int64) ([]corteCajaVenta, error) {
	startedAt := time.Now()
	defer func() {
		dbpkg.PerfLogf("[perf][corte] listCorteCajaVentasAnuladas empresa=%d usuario=%q dur=%s", empresaID, usuario, time.Since(startedAt))
	}()
	query := `SELECT
		c.id,
		COALESCE(c.codigo, ''),
		COALESCE(c.nombre, ''),
		COALESCE(m.fecha_evento, c.fecha_actualizacion, c.pagado_en, ''),
		COALESCE(c.metodo_pago, 'efectivo'),
		COALESCE(c.moneda, 'COP'),
		COALESCE(c.total, 0),
		COALESCE(m.monto_anulado, c.total_pagado, 0),
		COALESCE(c.devolucion_total, 0),
		COALESCE(m.usuario_creador, c.usuario_creador, ''),
		COALESCE(m.estacion_id, 0),
		COALESCE(m.estacion_codigo, ''),
		COALESCE(m.estacion_nombre, '')
	FROM carritos_compras c
	LEFT JOIN (
		SELECT carrito_id,
		       MAX(COALESCE(fecha_evento, '')) AS fecha_evento,
		       MAX(COALESCE(usuario_creador, '')) AS usuario_creador,
		       MAX(COALESCE(estacion_id, 0)) AS estacion_id,
		       MAX(COALESCE(estacion_codigo, '')) AS estacion_codigo,
		       MAX(COALESCE(estacion_nombre, '')) AS estacion_nombre,
		       MAX(COALESCE(monto_anulado, 0)) AS monto_anulado
		  FROM empresa_ventas_estacion_metricas
		 WHERE empresa_id = ?
		   AND LOWER(COALESCE(evento_operacion, '')) = 'venta_anulada'
		 GROUP BY carrito_id
	) m ON m.carrito_id = c.id
	WHERE c.empresa_id = ?
	  AND LOWER(COALESCE(c.estado_carrito, '')) IN ('anulado', 'anulada')
	  AND COALESCE(m.fecha_evento, c.fecha_actualizacion, c.pagado_en, '') >= ?
	  AND COALESCE(m.fecha_evento, c.fecha_actualizacion, c.pagado_en, '') <= ?
	  AND (? = '' OR LOWER(COALESCE(m.usuario_creador, c.usuario_creador, '')) = LOWER(?))
	  AND (? = 0 OR COALESCE(c.cierre_caja_id, 0) = ?)
	  AND (? = '' OR UPPER(COALESCE(c.caja_codigo, '')) = ?)
	ORDER BY COALESCE(m.fecha_evento, c.fecha_actualizacion, c.pagado_en, '') ASC, c.id ASC`
	cajaCodigo = strings.ToUpper(strings.TrimSpace(cajaCodigo))
	rows, err := dbpkg.ExecQueryCompat(dbEmp, query, empresaID, empresaID, desde, hasta, usuario, usuario, cierreCajaID, cierreCajaID, cajaCodigo, cajaCodigo)
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

func listCorteCajaItemsPorTipo(dbEmp *sql.DB, empresaID int64, desde, hasta, usuario, cajaCodigo string, cierreCajaID int64) ([]corteCajaTipoItem, error) {
	startedAt := time.Now()
	defer func() {
		dbpkg.PerfLogf("[perf][corte] listCorteCajaItemsPorTipo empresa=%d usuario=%q dur=%s", empresaID, usuario, time.Since(startedAt))
	}()
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
	  AND (? = 0 OR COALESCE(c.cierre_caja_id, 0) = ?)
	  AND (? = '' OR UPPER(COALESCE(c.caja_codigo, '')) = ?)
	GROUP BY LOWER(COALESCE(i.tipo_item, 'producto'))
	ORDER BY total DESC`
	cajaCodigo = strings.ToUpper(strings.TrimSpace(cajaCodigo))
	rows, err := dbpkg.ExecQueryCompat(dbEmp, query, empresaID, empresaID, desde, hasta, usuario, usuario, cierreCajaID, cierreCajaID, cajaCodigo, cajaCodigo)
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

func listCorteCajaProductosVendidos(dbEmp *sql.DB, empresaID int64, desde, hasta, usuario, cajaCodigo string, cierreCajaID int64, limit int) ([]corteCajaProductoVendido, error) {
	startedAt := time.Now()
	defer func() {
		dbpkg.PerfLogf("[perf][corte] listCorteCajaProductosVendidos empresa=%d usuario=%q dur=%s", empresaID, usuario, time.Since(startedAt))
	}()
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	query := `SELECT
		COALESCE(c.pagado_en, i.fecha_creacion, ''),
		COALESCE(i.descripcion, ''),
		COALESCE(i.cantidad, 0),
		LOWER(COALESCE(i.tipo_item, 'producto')),
		COALESCE(i.codigo_item, ''),
		COALESCE(c.codigo, '')
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
	  AND (? = '' OR LOWER(COALESCE(m.usuario_creador, c.usuario_creador, i.usuario_creador, '')) = LOWER(?))
	  AND (? = 0 OR COALESCE(c.cierre_caja_id, 0) = ?)
	  AND (? = '' OR UPPER(COALESCE(c.caja_codigo, '')) = ?)
	ORDER BY COALESCE(c.pagado_en, i.fecha_creacion, '') DESC, i.id DESC
	LIMIT ?`
	cajaCodigo = strings.ToUpper(strings.TrimSpace(cajaCodigo))
	rows, err := dbpkg.ExecQueryCompat(dbEmp, query, empresaID, empresaID, desde, hasta, usuario, usuario, cierreCajaID, cierreCajaID, cajaCodigo, cajaCodigo, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []corteCajaProductoVendido{}
	for rows.Next() {
		var item corteCajaProductoVendido
		if err := rows.Scan(&item.Fecha, &item.Producto, &item.Cantidad, &item.Tipo, &item.Referencia, &item.VentaCodigo); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func listCorteCajaMovimientos(dbEmp *sql.DB, empresaID int64, desde, hasta, usuario, cajaCodigo string, cierreCajaID int64) ([]corteCajaMovimientoGrupo, error) {
	startedAt := time.Now()
	defer func() {
		dbpkg.PerfLogf("[perf][corte] listCorteCajaMovimientos empresa=%d usuario=%q dur=%s", empresaID, usuario, time.Since(startedAt))
	}()
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
	  AND (? = 0 OR COALESCE(cierre_caja_id, 0) = ?)
	  AND (? = '' OR UPPER(COALESCE(caja_codigo, '')) = ?)
	GROUP BY LOWER(COALESCE(tipo_movimiento, '')), LOWER(COALESCE(metodo_pago, '')), COALESCE(categoria, ''), COALESCE(usuario_creador, '')
	ORDER BY 1, 2, 3, 4`
	cajaCodigo = strings.ToUpper(strings.TrimSpace(cajaCodigo))
	rows, err := dbpkg.ExecQueryCompat(dbEmp, query, empresaID, desde, hasta, usuario, usuario, cierreCajaID, cierreCajaID, cajaCodigo, cajaCodigo)
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
	startedAt := time.Now()
	defer func() {
		dbpkg.PerfLogf("[perf][corte] listCorteCajaSensoresSinFactura empresa=%d dur=%s", empresaID, time.Since(startedAt))
	}()
	if err := dbpkg.EnsureEmpresaSensorPuertasSchema(dbEmp); err != nil {
		return nil, err
	}
	query := `SELECT COALESCE(estacion_id, 0), COALESCE(device_id, ''), COALESCE(last_state, ''), COALESCE(last_seen, '')
		FROM empresa_sensor_puertas_devices
		WHERE empresa_id = ?
		  AND ` + corteCajaSensorEstadoActivoClause() + `
		  AND estacion_id > 0
		  AND ` + corteCajaSensorOccupiedClause()
	queryStarted := time.Now()
	rows, err := dbpkg.ExecQueryCompat(dbEmp, query, empresaID)
	if err != nil {
		return nil, err
	}
	dbpkg.PerfLogf("[perf][corte] sensores query empresa=%d dur=%s", empresaID, time.Since(queryStarted))
	defer rows.Close()
	out := []corteCajaSensorAlerta{}
	var scanned int
	scanStarted := time.Now()
	for rows.Next() {
		var estacionID int64
		var deviceID, state, seen string
		if err := rows.Scan(&estacionID, &deviceID, &state, &seen); err != nil {
			return nil, err
		}
		scanned++
		if soldStationIDs[estacionID] {
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
	dbpkg.PerfLogf("[perf][corte] sensores scan empresa=%d scanned=%d out=%d dur=%s", empresaID, scanned, len(out), time.Since(scanStarted))
	return out, rows.Err()
}

func corteCajaSensorEstadoActivoClause() string {
	return "LOWER(COALESCE(estado, 'activo')) = 'activo'"
}

func corteCajaSensorOccupiedClause() string {
	return `(COALESCE(last_state, '') <> ''
		AND lower(COALESCE(last_state, '')) NOT IN ('0', 'false', 'off', 'libre', 'disponible', 'free', 'abierta', 'abierto', 'open', 'vacia', 'vacio', 'vacía', 'vacío')
		AND (
			lower(COALESCE(last_state, '')) LIKE '%ocup%'
			OR lower(COALESCE(last_state, '')) LIKE '%occupied%'
			OR lower(COALESCE(last_state, '')) LIKE '%busy%'
			OR lower(COALESCE(last_state, '')) LIKE '%cerrad%'
			OR lower(COALESCE(last_state, '')) LIKE '%closed%'
			OR lower(COALESCE(last_state, '')) LIKE '%presencia%'
			OR lower(COALESCE(last_state, '')) = 'on'
			OR lower(COALESCE(last_state, '')) = 'true'
			OR lower(COALESCE(last_state, '')) = '1'
		))`
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

func normalizeCorteCajaMetodoDetalle(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	v = strings.ReplaceAll(v, " ", "_")
	v = strings.ReplaceAll(v, "-", "_")
	switch v {
	case "efectivo", "cash":
		return "efectivo"
	case "tarjeta_credito", "credito", "credit", "credit_card", "tc":
		return "tarjeta_credito"
	case "tarjeta_debito", "debito", "debit", "debit_card", "td":
		return "tarjeta_debito"
	case "tarjeta", "card":
		return "tarjeta"
	case "transferencia", "transferencia_bancaria", "banco":
		return "transferencia"
	default:
		return normalizeCorteCajaMetodo(raw)
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
