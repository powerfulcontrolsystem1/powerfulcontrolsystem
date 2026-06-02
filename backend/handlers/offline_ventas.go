package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type offlineVentaSyncRequest struct {
	EmpresaID      int64                 `json:"empresa_id"`
	SyncKey        string                `json:"sync_key"`
	Ventas         []offlineVentaPayload `json:"ventas"`
	FechaOffline   string                `json:"fecha_offline"`
	UsuarioEmail   string                `json:"usuario_email"`
	UsuarioRol     string                `json:"usuario_rol"`
	CarritoID      int64                 `json:"carrito_id"`
	CarritoCodigo  string                `json:"carrito_codigo"`
	EstacionID     int64                 `json:"estacion_id"`
	EstacionNombre string                `json:"estacion_nombre"`
	Modo           string                `json:"modo"`
	Carrito        offlineCarritoPayload `json:"carrito"`
	Items          []offlineItemPayload  `json:"items"`
	Pago           offlinePagoPayload    `json:"pago"`
}

type offlineVentaPayload struct {
	EmpresaID      int64                 `json:"empresa_id"`
	SyncKey        string                `json:"sync_key"`
	FechaOffline   string                `json:"fecha_offline"`
	UsuarioEmail   string                `json:"usuario_email"`
	UsuarioRol     string                `json:"usuario_rol"`
	CarritoID      int64                 `json:"carrito_id"`
	CarritoCodigo  string                `json:"carrito_codigo"`
	EstacionID     int64                 `json:"estacion_id"`
	EstacionNombre string                `json:"estacion_nombre"`
	Modo           string                `json:"modo"`
	Carrito        offlineCarritoPayload `json:"carrito"`
	Items          []offlineItemPayload  `json:"items"`
	Pago           offlinePagoPayload    `json:"pago"`
}

type offlineCarritoPayload struct {
	ID                int64   `json:"id"`
	Codigo            string  `json:"codigo"`
	Nombre            string  `json:"nombre"`
	CanalVenta        string  `json:"canal_venta"`
	ClienteID         int64   `json:"cliente_id"`
	Moneda            string  `json:"moneda"`
	ReferenciaExterna string  `json:"referencia_externa"`
	Total             float64 `json:"total"`
	Observaciones     string  `json:"observaciones"`
}

type offlineItemPayload struct {
	TipoItem            string  `json:"tipo_item"`
	ReferenciaID        int64   `json:"referencia_id"`
	CodigoItem          string  `json:"codigo_item"`
	Descripcion         string  `json:"descripcion"`
	UnidadMedida        string  `json:"unidad_medida"`
	Cantidad            float64 `json:"cantidad"`
	PrecioUnitario      float64 `json:"precio_unitario"`
	DescuentoPorcentaje float64 `json:"descuento_porcentaje"`
	ImpuestoPorcentaje  float64 `json:"impuesto_porcentaje"`
	ImpuestoCodigo      string  `json:"impuesto_codigo"`
	Observaciones       string  `json:"observaciones"`
}

type offlinePagoPayload struct {
	MetodoPago      string                    `json:"metodo_pago"`
	ReferenciaPago  string                    `json:"referencia_pago"`
	PagosMixtos     []carritoPagoMixtoEntrada `json:"pagos_mixtos"`
	DescuentoTipo   string                    `json:"descuento_tipo"`
	DescuentoCodigo string                    `json:"descuento_codigo"`
	CodigoDescuento string                    `json:"codigo_descuento"`
	DescuentoValor  float64                   `json:"descuento_valor"`
	DevolucionTotal float64                   `json:"devolucion_total"`
	AbonosTotal     float64                   `json:"abonos_total"`
	TotalPagado     float64                   `json:"total_pagado"`
	AplicarPropina  bool                      `json:"aplicar_propina"`
	CierreCajaID    int64                     `json:"cierre_caja_id"`
	CajaCodigo      string                    `json:"caja_codigo"`
	CajaTurno       string                    `json:"caja_turno"`
	CajaSucursalID  int64                     `json:"caja_sucursal_id"`
}

// EmpresaOfflineVentasHandler recibe ventas capturadas sin internet y las aplica al carrito unificado.
func EmpresaOfflineVentasHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			limit, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("limit")))
			rows, err := dbpkg.ListEmpresaVentasOfflineSync(dbEmp, empresaID, limit)
			if err != nil {
				http.Error(w, "no se pudo listar la bitacora offline: "+err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, rows)
		case http.MethodPost:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action != "" && action != "sync" {
				http.Error(w, "accion no soportada", http.StatusBadRequest)
				return
			}
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if !carritoOfflineFacturacionHabilitada(dbEmp, empresaID) {
				http.Error(w, "la facturacion offline esta desactivada para esta empresa", http.StatusForbidden)
				return
			}
			var input offlineVentaSyncRequest
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				http.Error(w, "payload invalido: "+err.Error(), http.StatusBadRequest)
				return
			}
			ventas := input.Ventas
			if len(ventas) == 0 {
				ventas = []offlineVentaPayload{{
					EmpresaID:      input.EmpresaID,
					SyncKey:        input.SyncKey,
					FechaOffline:   input.FechaOffline,
					UsuarioEmail:   input.UsuarioEmail,
					UsuarioRol:     input.UsuarioRol,
					CarritoID:      input.CarritoID,
					CarritoCodigo:  input.CarritoCodigo,
					EstacionID:     input.EstacionID,
					EstacionNombre: input.EstacionNombre,
					Modo:           input.Modo,
					Carrito:        input.Carrito,
					Items:          input.Items,
					Pago:           input.Pago,
				}}
			}
			results := make([]map[string]interface{}, 0, len(ventas))
			status := http.StatusOK
			for _, venta := range ventas {
				venta.EmpresaID = empresaID
				result, err := syncOfflineVenta(r, dbEmp, dbSuper, empresaID, venta)
				if err != nil {
					status = http.StatusConflict
					result = map[string]interface{}{
						"ok":        false,
						"sync_key":  normalizeOfflineSyncKey(venta.SyncKey),
						"error":     err.Error(),
						"pendiente": true,
					}
				}
				results = append(results, result)
			}
			writeJSON(w, status, map[string]interface{}{
				"ok":      status == http.StatusOK,
				"results": results,
			})
		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
		}
	}
}

func syncOfflineVenta(r *http.Request, dbEmp, dbSuper *sql.DB, empresaID int64, venta offlineVentaPayload) (map[string]interface{}, error) {
	syncKey := normalizeOfflineSyncKey(venta.SyncKey)
	if syncKey == "" {
		syncKey = normalizeOfflineSyncKey(fmt.Sprintf("OFF-%d-%d", empresaID, time.Now().UnixNano()))
	}
	usuario := strings.TrimSpace(adminEmailFromRequest(r))
	if usuario == "" {
		usuario = "sistema"
	}
	payloadBytes, _ := json.Marshal(venta)
	if err := dbpkg.UpsertEmpresaVentaOfflineSyncPending(dbEmp, empresaID, syncKey, string(payloadBytes), venta.FechaOffline, usuario, "venta capturada sin internet"); err != nil {
		return nil, err
	}
	existing, errExisting := dbpkg.GetEmpresaVentaOfflineSyncByKey(dbEmp, empresaID, syncKey)
	if errExisting == nil && existing != nil && strings.EqualFold(existing.EstadoSync, "sincronizado") {
		return map[string]interface{}{
			"ok":               true,
			"sync_key":         syncKey,
			"idempotente":      true,
			"carrito_id":       existing.CarritoID,
			"documento_codigo": existing.DocumentoCodigo,
			"estado_sync":      existing.EstadoSync,
		}, nil
	}
	if err := validateOfflineVentaSessionOwner(venta, usuario); err != nil {
		_ = dbpkg.MarkEmpresaVentaOfflineSyncResult(dbEmp, empresaID, syncKey, "error", 0, "", "", err.Error())
		return nil, err
	}
	if err := validateOfflineVentaCaja(venta.Pago); err != nil {
		_ = dbpkg.MarkEmpresaVentaOfflineSyncResult(dbEmp, empresaID, syncKey, "error", 0, "", "", err.Error())
		return nil, err
	}

	carrito, err := resolveOfflineCarrito(dbEmp, empresaID, venta, syncKey, usuario, existing)
	if err != nil {
		_ = dbpkg.MarkEmpresaVentaOfflineSyncResult(dbEmp, empresaID, syncKey, "error", 0, "", "", err.Error())
		return nil, err
	}
	if carrito.PagadoEn != "" || strings.EqualFold(carrito.EstadoCarrito, "cerrado") {
		result := map[string]interface{}{
			"ok":          true,
			"sync_key":    syncKey,
			"carrito_id":  carrito.ID,
			"estado_sync": "sincronizado",
			"message":     "carrito ya estaba pagado",
		}
		resultBytes, _ := json.Marshal(result)
		_ = dbpkg.MarkEmpresaVentaOfflineSyncResult(dbEmp, empresaID, syncKey, "sincronizado", carrito.ID, "", string(resultBytes), "")
		return result, nil
	}

	if carritoClienteObligatorioParaPago(dbEmp, empresaID) && carrito.ClienteID <= 0 {
		err := fmt.Errorf("cliente obligatorio: la venta offline no tiene cliente registrado")
		_ = dbpkg.MarkEmpresaVentaOfflineSyncResult(dbEmp, empresaID, syncKey, "error", carrito.ID, "", "", err.Error())
		return nil, err
	}

	pago := venta.Pago
	metodoPago := dbpkg.NormalizeMetodoPagoCarrito(pago.MetodoPago)
	if metodoPago == "" {
		metodoPago = "efectivo"
	}
	referenciaPago := strings.TrimSpace(pago.ReferenciaPago)
	if metodoPago == "mixto" && len(pago.PagosMixtos) > 0 {
		pagosMixtos, _, err := normalizePagosMixtosCarrito(pago.PagosMixtos, carrito.Moneda)
		if err != nil {
			_ = dbpkg.MarkEmpresaVentaOfflineSyncResult(dbEmp, empresaID, syncKey, "error", carrito.ID, "", "", err.Error())
			return nil, err
		}
		referenciaPago = buildReferenciaPagoMixto(pagosMixtos, carrito.Moneda)
	}

	descuentoTipo := strings.TrimSpace(strings.ToLower(pago.DescuentoTipo))
	descuentoCodigo := strings.TrimSpace(pago.DescuentoCodigo)
	if descuentoCodigo == "" {
		descuentoCodigo = strings.TrimSpace(pago.CodigoDescuento)
	}
	descuentoValor := pago.DescuentoValor
	if descuentoValor < 0 {
		descuentoValor = 0
	}
	codigoDescuentoID := int64(0)
	if descuentoCodigo != "" || metodoPago == "codigo_descuento" || descuentoTipo == "code" {
		aplicado, err := dbpkg.ResolveCodigoDescuentoParaMontoConContexto(dbEmp, empresaID, descuentoCodigo, carrito.Total, dbpkg.CodigoDescuentoContexto{
			CarritoID:  carrito.ID,
			ClienteID:  carrito.ClienteID,
			CanalVenta: carrito.CanalVenta,
		})
		if err != nil {
			_ = dbpkg.MarkEmpresaVentaOfflineSyncResult(dbEmp, empresaID, syncKey, "error", carrito.ID, "", "", err.Error())
			return nil, err
		}
		codigoDescuentoID = aplicado.ID
		descuentoTipo = "code"
		descuentoCodigo = aplicado.Codigo
		descuentoValor = aplicado.ValorAplicado
	}
	if descuentoValor > carrito.Total {
		descuentoValor = carrito.Total
	}
	descuentoValor = roundMoneyCarritoForMoneda(descuentoValor, carrito.Moneda)
	devolucionTotal := pago.DevolucionTotal
	if devolucionTotal < 0 {
		devolucionTotal = 0
	}
	devolucionTotal = roundMoneyCarritoForMoneda(devolucionTotal, carrito.Moneda)
	totalPagado := pago.TotalPagado
	if totalPagado <= 0 {
		totalPagado = carrito.Total - descuentoValor - devolucionTotal
	}
	if totalPagado < 0 {
		totalPagado = 0
	}
	totalPagado = roundMoneyCarritoForMoneda(totalPagado, carrito.Moneda)

	caja, _, err := openCajaCobroForCarrito(dbEmp, dbSuper, empresaID, pago.CajaCodigo, pago.CajaTurno, pago.CajaSucursalID, 0, carrito.Moneda, usuario)
	if err != nil {
		_ = dbpkg.MarkEmpresaVentaOfflineSyncResult(dbEmp, empresaID, syncKey, "error", carrito.ID, "", "", err.Error())
		return nil, err
	}

	if err := dbpkg.PayCarritoStationSession(dbEmp, empresaID, carrito.ID, metodoPago, referenciaPago, descuentoTipo, descuentoCodigo, descuentoValor, devolucionTotal, totalPagado, codigoDescuentoID, caja.ID, caja.CajaCodigo, caja.Turno, caja.SucursalID, usuario); err != nil {
		if errors.Is(err, dbpkg.ErrCarritoYaPagado) {
			if carritoActual, errActual := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, carrito.ID); errActual == nil && carritoActual != nil && offlineCarritoEstaPagado(carritoActual) {
				result := map[string]interface{}{
					"ok":          true,
					"sync_key":    syncKey,
					"idempotente": true,
					"carrito_id":  carritoActual.ID,
					"estado_sync": "sincronizado",
					"message":     "carrito ya estaba pagado; no se duplico la venta offline",
				}
				resultBytes, _ := json.Marshal(result)
				_ = dbpkg.MarkEmpresaVentaOfflineSyncResult(dbEmp, empresaID, syncKey, "sincronizado", carritoActual.ID, "", string(resultBytes), "")
				return result, nil
			}
		}
		_ = dbpkg.MarkEmpresaVentaOfflineSyncResult(dbEmp, empresaID, syncKey, "error", carrito.ID, "", "", err.Error())
		return nil, err
	}
	_ = dbpkg.InvalidateVIPCodesForCarrito(dbEmp, empresaID, carrito.ID, "venta_offline_sync")

	carritoPagado, err := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, carrito.ID)
	if err != nil {
		carritoPagado = carrito
	}
	estacionID, estacionCodigo, estacionNombre := dbpkg.ResolveCarritoStationIdentity(carritoPagado)
	if estacionID <= 0 && venta.EstacionID > 0 {
		estacionID = venta.EstacionID
		estacionCodigo = fmt.Sprintf("EST-%d", venta.EstacionID)
		estacionNombre = strings.TrimSpace(venta.EstacionNombre)
	}
	if _, errMetric := dbpkg.RecordCarritoStationMetric(dbEmp, dbpkg.CarritoStationMetricInput{
		EmpresaID:           empresaID,
		CarritoID:           carritoPagado.ID,
		EstacionID:          estacionID,
		EstacionCodigo:      estacionCodigo,
		EstacionNombre:      estacionNombre,
		EventoOperacion:     "venta_offline_sincronizada",
		MetodoPago:          metodoPago,
		Moneda:              carritoPagado.Moneda,
		MontoTotal:          carritoPagado.Total,
		MontoPagado:         totalPagado,
		DevolucionTotal:     devolucionTotal,
		ActivadoEn:          carritoPagado.ActivadoEn,
		PagadoEn:            carritoPagado.PagadoEn,
		ReferenciaOperacion: referenciaPago,
		CierreCajaID:        caja.ID,
		CajaCodigo:          caja.CajaCodigo,
		CajaTurno:           caja.Turno,
		CajaSucursalID:      caja.SucursalID,
		UsuarioCreador:      usuario,
		Observaciones:       "venta sincronizada despues de operar sin internet",
	}); errMetric != nil {
		log.Printf("[offline_ventas] metrica omitida empresa_id=%d carrito_id=%d error=%v", empresaID, carritoPagado.ID, errMetric)
	}
	if efectivo := offlineMontoEfectivoCaja(metodoPago, totalPagado, pago.PagosMixtos); efectivo > 0 {
		if err := dbpkg.RegistrarIngresoEfectivoCierreCaja(dbEmp, empresaID, caja.ID, efectivo); err != nil {
			log.Printf("[offline_ventas] efectivo caja omitido empresa_id=%d cierre_id=%d error=%v", empresaID, caja.ID, err)
		}
	}
	documentoVenta, errDocumento := registrarDocumentoVentaDesdeCarritoPagado(dbEmp, dbSuper, carritoPagado, totalPagado, usuario)
	if errDocumento != nil {
		log.Printf("[offline_ventas] documento venta empresa_id=%d carrito_id=%d error=%v", empresaID, carritoPagado.ID, errDocumento)
	}
	dispatchControlElectricoEstacionAsync(dbEmp, carritoPagado, false, usuario, "venta_offline_sync")

	documentoCodigo := ""
	if documentoVenta != nil {
		documentoCodigo = strings.TrimSpace(fmt.Sprint(documentoVenta["documento_codigo"]))
	}
	result := map[string]interface{}{
		"ok":              true,
		"sync_key":        syncKey,
		"estado_sync":     "sincronizado",
		"carrito_id":      carritoPagado.ID,
		"documento_venta": documentoVenta,
		"caja": map[string]interface{}{
			"id":          caja.ID,
			"caja_codigo": caja.CajaCodigo,
			"turno":       caja.Turno,
		},
	}
	resultBytes, _ := json.Marshal(result)
	_ = dbpkg.MarkEmpresaVentaOfflineSyncResult(dbEmp, empresaID, syncKey, "sincronizado", carritoPagado.ID, documentoCodigo, string(resultBytes), "")
	return result, nil
}

func validateOfflineVentaSessionOwner(venta offlineVentaPayload, sessionEmail string) error {
	claimed := strings.ToLower(strings.TrimSpace(venta.UsuarioEmail))
	session := strings.ToLower(strings.TrimSpace(sessionEmail))
	if claimed == "" || session == "" || session == "sistema" {
		return nil
	}
	if claimed != session {
		return fmt.Errorf("esta venta offline pertenece al cajero %s; inicia sesion con ese usuario para sincronizarla", claimed)
	}
	return nil
}

func validateOfflineVentaCaja(pago offlinePagoPayload) error {
	if strings.TrimSpace(pago.CajaCodigo) == "" {
		return fmt.Errorf("la venta offline debe indicar el codigo de la caja abierta del cajero antes de sincronizar")
	}
	return nil
}

func offlineCarritoEstaPagado(carrito *dbpkg.CarritoCompra) bool {
	if carrito == nil {
		return false
	}
	return strings.TrimSpace(carrito.PagadoEn) != "" ||
		strings.EqualFold(carrito.EstadoCarrito, "cerrado") ||
		strings.EqualFold(carrito.EstadoCarrito, "pagado") ||
		strings.EqualFold(carrito.Estado, "inactivo")
}

func resolveOfflineCarrito(dbEmp *sql.DB, empresaID int64, venta offlineVentaPayload, syncKey, usuario string, existing *dbpkg.EmpresaVentaOfflineSync) (*dbpkg.CarritoCompra, error) {
	if existing != nil && existing.CarritoID > 0 {
		if carrito, err := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, existing.CarritoID); err == nil && carrito != nil {
			return carrito, nil
		}
	}
	carritoID := venta.CarritoID
	if carritoID <= 0 {
		carritoID = venta.Carrito.ID
	}
	if carritoID > 0 {
		if carrito, err := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, carritoID); err == nil && carrito != nil {
			if err := ensureOfflineItemsIfNeeded(dbEmp, empresaID, carrito.ID, venta.Items, usuario); err != nil {
				return nil, err
			}
			return dbpkg.GetCarritoCompraByID(dbEmp, empresaID, carrito.ID)
		} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}
	code := strings.TrimSpace(venta.CarritoCodigo)
	if code == "" {
		code = strings.TrimSpace(venta.Carrito.Codigo)
	}
	if code == "" {
		code = "OFF-" + syncKey
	}
	if carrito, err := dbpkg.GetCarritoCompraByCodigo(dbEmp, empresaID, code); err == nil && carrito != nil {
		if err := ensureOfflineItemsIfNeeded(dbEmp, empresaID, carrito.ID, venta.Items, usuario); err != nil {
			return nil, err
		}
		return dbpkg.GetCarritoCompraByID(dbEmp, empresaID, carrito.ID)
	}
	nombre := strings.TrimSpace(venta.Carrito.Nombre)
	if nombre == "" {
		nombre = "Venta offline " + syncKey
	}
	moneda := strings.TrimSpace(venta.Carrito.Moneda)
	if moneda == "" {
		moneda = "COP"
	}
	referencia := strings.TrimSpace(venta.Carrito.ReferenciaExterna)
	if referencia == "" {
		referencia = "OFFLINE:" + syncKey
	}
	newID, err := dbpkg.CreateCarritoCompra(dbEmp, dbpkg.CarritoCompra{
		EmpresaID:         empresaID,
		Codigo:            code,
		Nombre:            nombre,
		CanalVenta:        strings.TrimSpace(venta.Carrito.CanalVenta),
		ClienteID:         venta.Carrito.ClienteID,
		EstadoCarrito:     "abierto",
		Moneda:            moneda,
		ReferenciaExterna: referencia,
		UsuarioCreador:    usuario,
		Observaciones:     strings.TrimSpace("Sincronizado desde modo sin internet. " + venta.Carrito.Observaciones),
	})
	if err != nil {
		return nil, err
	}
	_ = dbpkg.MarkEmpresaVentaOfflineSyncResult(dbEmp, empresaID, syncKey, "procesando", newID, "", "", "")
	if err := ensureOfflineItemsIfNeeded(dbEmp, empresaID, newID, venta.Items, usuario); err != nil {
		return nil, err
	}
	return dbpkg.GetCarritoCompraByID(dbEmp, empresaID, newID)
}

func ensureOfflineItemsIfNeeded(dbEmp *sql.DB, empresaID, carritoID int64, items []offlineItemPayload, usuario string) error {
	current, err := dbpkg.GetCarritoCompraItems(dbEmp, empresaID, carritoID, false)
	if err != nil {
		return err
	}
	if len(current) > 0 || len(items) == 0 {
		return nil
	}
	for _, item := range items {
		if strings.TrimSpace(item.Descripcion) == "" || item.Cantidad <= 0 {
			continue
		}
		_, err := dbpkg.CreateCarritoCompraItem(dbEmp, dbpkg.CarritoCompraItem{
			EmpresaID:           empresaID,
			CarritoID:           carritoID,
			TipoItem:            item.TipoItem,
			ReferenciaID:        item.ReferenciaID,
			CodigoItem:          item.CodigoItem,
			Descripcion:         item.Descripcion,
			UnidadMedida:        item.UnidadMedida,
			Cantidad:            item.Cantidad,
			PrecioUnitario:      item.PrecioUnitario,
			DescuentoPorcentaje: item.DescuentoPorcentaje,
			ImpuestoPorcentaje:  item.ImpuestoPorcentaje,
			ImpuestoCodigo:      item.ImpuestoCodigo,
			UsuarioCreador:      usuario,
			Estado:              "activo",
			Observaciones:       strings.TrimSpace("capturado offline. " + item.Observaciones),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func carritoOfflineFacturacionHabilitada(dbEmp *sql.DB, empresaID int64) bool {
	if dbEmp == nil || empresaID <= 0 {
		return false
	}
	pref, err := dbpkg.GetEmpresaEstacionPref(dbEmp, empresaID, 0, "estaciones_config")
	if err != nil || pref == nil || strings.TrimSpace(pref.Valor) == "" {
		return false
	}
	root := carritoParseConfigJSON(pref.Valor)
	if root == nil {
		return false
	}
	for _, key := range []string{"carrito_ui_global", "carrito", "carrito_configuracion_global"} {
		cfg, ok := root[key].(map[string]interface{})
		if !ok {
			continue
		}
		if carritoBoolFromConfigValue(cfg["facturacion_offline_habilitada"]) || carritoBoolFromConfigValue(cfg["permitir_facturacion_offline"]) {
			return true
		}
	}
	return false
}

func normalizeOfflineSyncKey(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	var b strings.Builder
	for _, r := range value {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	out := strings.Trim(b.String(), "-_")
	if len(out) > 90 {
		out = out[:90]
	}
	return out
}

func offlineMontoEfectivoCaja(metodoPago string, total float64, mixtos []carritoPagoMixtoEntrada) float64 {
	if metodoPago == "efectivo" {
		return total
	}
	if metodoPago != "mixto" {
		return 0
	}
	sum := 0.0
	for _, item := range mixtos {
		if dbpkg.NormalizeMetodoPagoCarrito(item.Metodo) == "efectivo" && item.Monto > 0 {
			sum += item.Monto
		}
	}
	return sum
}
