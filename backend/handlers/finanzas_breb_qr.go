package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type empresaBrebQRConfigPayload struct {
	EmpresaID                    int64                    `json:"empresa_id"`
	PagoQRHabilitado             bool                     `json:"pago_qr_habilitado"`
	MetodoPagoTransferenciaBreb  bool                     `json:"metodo_pago_transferencia_bre_b"`
	PermitirPagoMixto            bool                     `json:"permitir_pago_mixto"`
	BrebConciliacionAutomatica   bool                     `json:"breb_conciliacion_automatica"`
	BrebRequiereComprobante      bool                     `json:"breb_requiere_comprobante"`
	BrebReferenciaPrefijo        string                   `json:"breb_referencia_prefijo"`
	BrebWebhookURL               string                   `json:"breb_webhook_url"`
	BrebAlertaPagosPendientesMin int                      `json:"breb_alerta_pagos_pendientes_min"`
	BrebCuentaDefaultPorCaja     bool                     `json:"breb_cuenta_default_por_caja"`
	BrebInstruccionesOperacion   string                   `json:"breb_instrucciones_operacion"`
	PagoQRCuentas                []map[string]interface{} `json:"pago_qr_cuentas"`
}

type empresaBrebQRPago struct {
	Origen           string  `json:"origen"`
	ID               int64   `json:"id"`
	CarritoID        int64   `json:"carrito_id,omitempty"`
	Codigo           string  `json:"codigo,omitempty"`
	ReferenciaCaja   string  `json:"referencia_caja,omitempty"`
	CajaCodigo       string  `json:"caja_codigo,omitempty"`
	CajaTurno        string  `json:"caja_turno,omitempty"`
	MetodoPago       string  `json:"metodo_pago"`
	ReferenciaPago   string  `json:"referencia_pago"`
	Monto            float64 `json:"monto"`
	Moneda           string  `json:"moneda"`
	Fecha            string  `json:"fecha"`
	UsuarioCreador   string  `json:"usuario_creador"`
	Estado           string  `json:"estado"`
	EstadoConciliado string  `json:"estado_conciliacion,omitempty"`
	Observaciones    string  `json:"observaciones,omitempty"`
}

type empresaBrebQRRegistroManualPayload struct {
	EmpresaID          int64   `json:"empresa_id"`
	FechaMovimiento    string  `json:"fecha_movimiento"`
	CuentaBancaria     string  `json:"cuenta_bancaria"`
	BancoNombre        string  `json:"banco_nombre"`
	ReferenciaBancaria string  `json:"referencia_bancaria"`
	DocumentoCodigo    string  `json:"documento_codigo"`
	Monto              float64 `json:"monto"`
	Moneda             string  `json:"moneda"`
	EstadoConciliacion string  `json:"estado_conciliacion"`
	CajaCodigo         string  `json:"caja_codigo"`
	CajaTurno          string  `json:"caja_turno"`
	EstacionID         int64   `json:"estacion_id"`
	CarritoID          int64   `json:"carrito_id"`
	PagadorNombre      string  `json:"pagador_nombre"`
	PagadorDocumento   string  `json:"pagador_documento"`
	Observaciones      string  `json:"observaciones"`
}

const (
	brebQRCarritosOrderExpr = "pcs_ts(COALESCE(pagado_en, fecha_actualizacion, fecha_creacion, ''))"
	brebQRAbonosOrderExpr   = "pcs_ts(COALESCE(a.fecha_abono, a.fecha_creacion, ''))"
	brebQRBancosOrderExpr   = "pcs_ts(COALESCE(fecha_movimiento, fecha_creacion, ''))"
)

var brebQRReferencePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9 ._:/+\-]{2,78}[A-Za-z0-9]$`)

// EmpresaFinanzasBrebQRHandler centraliza configuracion y trazabilidad Bre-B QR por empresa.
func EmpresaFinanzasBrebQRHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := ensureEmpresaBrebQRSchemas(dbEmp); err != nil {
				log.Printf("[breb_qr] ensure schema empresa_id=%d error: %v", empresaID, err)
				http.Error(w, "No se pudo preparar el modulo Bre-B QR", http.StatusInternalServerError)
				return
			}
			limit, err := parseIntQueryOptional(r, "limit")
			if err != nil {
				http.Error(w, "limit invalido", http.StatusBadRequest)
				return
			}
			if limit <= 0 || limit > 200 {
				limit = 80
			}
			config, configErr := loadEmpresaBrebQRConfig(dbEmp, empresaID)
			if configErr != nil {
				log.Printf("[breb_qr] load config empresa_id=%d error: %v", empresaID, configErr)
				http.Error(w, "No se pudo leer configuracion Bre-B QR", http.StatusInternalServerError)
				return
			}
			pagos, pagosErr := listEmpresaBrebQRPagos(dbEmp, empresaID, limit)
			if pagosErr != nil {
				log.Printf("[breb_qr] list pagos empresa_id=%d error: %v", empresaID, pagosErr)
				http.Error(w, "No se pudo consultar registro de pagos Bre-B", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":         true,
				"empresa_id": empresaID,
				"config":     config,
				"resumen":    summarizeEmpresaBrebQRPagos(pagos),
				"pagos":      pagos,
			})
			return
		case http.MethodPut:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := ensureEmpresaBrebQRSchemas(dbEmp); err != nil {
				log.Printf("[breb_qr] ensure schema put empresa_id=%d error: %v", empresaID, err)
				http.Error(w, "No se pudo preparar el modulo Bre-B QR", http.StatusInternalServerError)
				return
			}
			var payload empresaBrebQRConfigPayload
			r.Body = http.MaxBytesReader(w, r.Body, 256<<10)
			decoder := json.NewDecoder(r.Body)
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			payload.EmpresaID = empresaID
			if payload.BrebAlertaPagosPendientesMin < 0 {
				payload.BrebAlertaPagosPendientesMin = 0
			}
			if payload.BrebReferenciaPrefijo == "" {
				payload.BrebReferenciaPrefijo = "BREB"
			}
			normalizedAccounts, err := validateAndNormalizeBrebQRCuentas(payload.PagoQRCuentas)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if payload.BrebConciliacionAutomatica || strings.TrimSpace(payload.BrebWebhookURL) != "" {
				http.Error(w, "la conciliacion automatica Bre-B no puede activarse sin un conector bancario firmado implementado en el servidor", http.StatusConflict)
				return
			}
			payload.PagoQRCuentas = normalizedAccounts
			if err := saveEmpresaBrebQRConfig(dbEmp, payload, adminEmailFromRequest(r)); err != nil {
				log.Printf("[breb_qr] save config empresa_id=%d error: %v", empresaID, err)
				http.Error(w, "No se pudo guardar configuracion Bre-B QR", http.StatusInternalServerError)
				return
			}
			registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, "breb_qr", "configuracion_guardada", "empresa_estacion_prefs", 0, http.StatusOK, map[string]interface{}{
				"pago_qr_habilitado":        payload.PagoQRHabilitado,
				"transferencia_bre_b":       payload.MetodoPagoTransferenciaBreb,
				"permitir_pago_mixto":       payload.PermitirPagoMixto,
				"conciliacion_automatica":   payload.BrebConciliacionAutomatica,
				"requiere_comprobante":      payload.BrebRequiereComprobante,
				"cuentas_receptoras":        len(normalizeBrebQRCuentas(payload.PagoQRCuentas)),
				"cuenta_default_por_caja":   payload.BrebCuentaDefaultPorCaja,
				"alerta_pendientes_minutos": payload.BrebAlertaPagosPendientesMin,
			}, "configuracion Bre-B QR actualizada por empresa")
			config, _ := loadEmpresaBrebQRConfig(dbEmp, empresaID)
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "empresa_id": empresaID, "config": config})
			return
		case http.MethodPost:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action != "registro_manual" {
				http.Error(w, "accion no permitida", http.StatusBadRequest)
				return
			}
			if err := ensureEmpresaBrebQRSchemas(dbEmp); err != nil {
				log.Printf("[breb_qr] ensure schema post empresa_id=%d error: %v", empresaID, err)
				http.Error(w, "No se pudo preparar el modulo Bre-B QR", http.StatusInternalServerError)
				return
			}
			var payload empresaBrebQRRegistroManualPayload
			r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
			decoder := json.NewDecoder(r.Body)
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			payload.EmpresaID = empresaID
			id, err := insertEmpresaBrebQRRegistroManual(dbEmp, payload, adminEmailFromRequest(r))
			if err != nil {
				log.Printf("[breb_qr] registro manual empresa_id=%d error: %v", empresaID, err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, "breb_qr", "pago_manual_registrado", "empresa_finanzas_bancos_movimientos", id, http.StatusOK, map[string]interface{}{
				"monto":               payload.Monto,
				"moneda":              strings.ToUpper(strings.TrimSpace(payload.Moneda)),
				"estado_conciliacion": strings.ToLower(strings.TrimSpace(payload.EstadoConciliacion)),
				"caja_codigo":         strings.TrimSpace(payload.CajaCodigo),
				"estacion_id":         payload.EstacionID,
				"carrito_id":          payload.CarritoID,
				"documento_codigo":    strings.TrimSpace(payload.DocumentoCodigo),
			}, "pago Bre-B QR registrado manualmente para conciliacion")
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id, "empresa_id": empresaID})
			return
		default:
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
	}
}

func ensureEmpresaBrebQRSchemas(dbEmp *sql.DB) error {
	if err := dbpkg.EmpresaEstacionPrefsSchemaReady(dbEmp); err != nil {
		return err
	}
	if err := dbpkg.EmpresaCarritosSchemaReady(dbEmp); err != nil {
		return err
	}
	if err := dbpkg.EmpresaFinanzasSchemaReady(dbEmp); err != nil {
		return err
	}
	return nil
}

func loadEmpresaBrebQRConfig(dbEmp *sql.DB, empresaID int64) (map[string]interface{}, error) {
	root, _, err := loadEmpresaEstacionesConfigMap(dbEmp, empresaID)
	if err != nil {
		return nil, err
	}
	global := getOrCreateMap(root, "carrito_ui_global")
	out := map[string]interface{}{
		"pago_qr_habilitado":                    boolFromInterface(global["pago_qr_habilitado"], false),
		"metodo_pago_transferencia_bre_b":       boolFromInterface(global["metodo_pago_transferencia_bre_b"], true),
		"permitir_pago_mixto":                   boolFromInterface(global["permitir_pago_mixto"], true),
		"breb_conciliacion_automatica":          false,
		"breb_requiere_comprobante":             boolFromInterface(global["breb_requiere_comprobante"], true),
		"breb_referencia_prefijo":               stringFromInterfaceDefault(global["breb_referencia_prefijo"], "BREB"),
		"breb_webhook_url":                      "",
		"breb_alerta_pagos_pendientes_min":      intFromInterfaceDefault(global["breb_alerta_pagos_pendientes_min"], 10),
		"breb_cuenta_default_por_caja":          boolFromInterface(global["breb_cuenta_default_por_caja"], false),
		"breb_instrucciones_operacion":          stringFromInterfaceDefault(global["breb_instrucciones_operacion"], "Verificar el abono en la app bancaria antes de cerrar el carrito."),
		"pago_qr_cuentas":                       normalizeBrebQRCuentas(global["pago_qr_cuentas"]),
		"pago_qr_proveedor_legacy":              stringFromInterfaceDefault(global["pago_qr_proveedor"], "breb"),
		"pago_qr_llave_legacy":                  stringFromInterfaceDefault(global["pago_qr_llave"], ""),
		"pago_qr_payload_oficial_legacy":        stringFromInterfaceDefault(global["pago_qr_payload_oficial"], ""),
		"origen_configuracion":                  "empresa_estacion_prefs.estaciones_config.carrito_ui_global",
		"requiere_integracion_bancaria_externa": true,
	}
	return out, nil
}

func saveEmpresaBrebQRConfig(dbEmp *sql.DB, payload empresaBrebQRConfigPayload, usuario string) error {
	root, raw, err := loadEmpresaEstacionesConfigMap(dbEmp, payload.EmpresaID)
	if err != nil {
		return err
	}
	_ = raw
	global := getOrCreateMap(root, "carrito_ui_global")
	global["pago_qr_habilitado"] = payload.PagoQRHabilitado
	global["metodo_pago_transferencia_bre_b"] = payload.MetodoPagoTransferenciaBreb
	global["permitir_pago_mixto"] = payload.PermitirPagoMixto
	global["breb_conciliacion_automatica"] = false
	global["breb_requiere_comprobante"] = payload.BrebRequiereComprobante
	global["breb_referencia_prefijo"] = strings.TrimSpace(payload.BrebReferenciaPrefijo)
	global["breb_webhook_url"] = ""
	global["breb_alerta_pagos_pendientes_min"] = payload.BrebAlertaPagosPendientesMin
	global["breb_cuenta_default_por_caja"] = payload.BrebCuentaDefaultPorCaja
	global["breb_instrucciones_operacion"] = strings.TrimSpace(payload.BrebInstruccionesOperacion)
	normalizedAccounts := normalizeBrebQRCuentas(payload.PagoQRCuentas)
	global["pago_qr_cuentas"] = normalizedAccounts
	if len(normalizedAccounts) > 0 {
		first := normalizedAccounts[0]
		global["pago_qr_proveedor"] = first["proveedor"]
		global["pago_qr_llave"] = first["llave"]
		global["pago_qr_comercio"] = first["comercio"]
		global["pago_qr_payload_oficial"] = first["payload_oficial"]
		global["pago_qr_instrucciones"] = first["instrucciones"]
	}
	blob, err := json.Marshal(root)
	if err != nil {
		return err
	}
	_, err = dbpkg.UpsertEmpresaEstacionPref(dbEmp, dbpkg.EmpresaEstacionPref{
		EmpresaID:      payload.EmpresaID,
		EstacionID:     0,
		Clave:          "estaciones_config",
		Valor:          string(blob),
		UsuarioCreador: strings.TrimSpace(usuario),
		Estado:         "activo",
		Observaciones:  "Configuracion Bre-B QR actualizada desde Finanzas",
	})
	return err
}

func loadEmpresaEstacionesConfigMap(dbEmp *sql.DB, empresaID int64) (map[string]interface{}, string, error) {
	pref, err := dbpkg.GetEmpresaEstacionPref(dbEmp, empresaID, 0, "estaciones_config")
	if err != nil {
		return nil, "", err
	}
	root := map[string]interface{}{}
	raw := ""
	if pref != nil {
		raw = strings.TrimSpace(pref.Valor)
	}
	if raw == "" {
		return root, raw, nil
	}
	var current interface{} = raw
	for i := 0; i < 3; i++ {
		asString, ok := current.(string)
		if !ok {
			break
		}
		var decoded interface{}
		if err := json.Unmarshal([]byte(strings.TrimSpace(asString)), &decoded); err != nil {
			return nil, raw, err
		}
		current = decoded
	}
	blob, err := json.Marshal(current)
	if err != nil {
		return nil, raw, err
	}
	if err := json.Unmarshal(blob, &root); err != nil {
		return nil, raw, err
	}
	return root, raw, nil
}

func getOrCreateMap(root map[string]interface{}, key string) map[string]interface{} {
	if root == nil {
		return map[string]interface{}{}
	}
	if existing, ok := root[key].(map[string]interface{}); ok {
		return existing
	}
	next := map[string]interface{}{}
	root[key] = next
	return next
}

func listEmpresaBrebQRPagos(dbEmp *sql.DB, empresaID int64, limit int) ([]empresaBrebQRPago, error) {
	pagos := make([]empresaBrebQRPago, 0)
	carritos, err := dbpkg.ExecQueryCompat(dbEmp, `SELECT
		id, COALESCE(codigo,''), COALESCE(referencia_externa,''), COALESCE(caja_codigo,''), COALESCE(caja_turno,''),
		COALESCE(metodo_pago,''), COALESCE(referencia_pago,''), COALESCE(NULLIF(total_pagado,0), total, 0),
		COALESCE(moneda,'COP'), COALESCE(pagado_en, fecha_actualizacion, fecha_creacion, ''), COALESCE(usuario_creador,''), COALESCE(estado_carrito,'')
	FROM carritos_compras
	WHERE empresa_id = ?
	  AND LOWER(COALESCE(estado_carrito,'')) = 'pagado'
	  AND (
		LOWER(COALESCE(metodo_pago,'')) = 'transferencia_bre_b'
		OR (LOWER(COALESCE(metodo_pago,'')) = 'mixto' AND LOWER(COALESCE(referencia_pago,'')) LIKE '%transferencia_bre_b%')
		OR LOWER(COALESCE(referencia_pago,'')) LIKE 'qr-breb-%'
		OR LOWER(COALESCE(referencia_pago,'')) LIKE 'breb-%'
	  )
	ORDER BY `+brebQRCarritosOrderExpr+` DESC, id DESC
	LIMIT ?`, empresaID, limit)
	if err != nil {
		return nil, err
	}
	defer carritos.Close()
	for carritos.Next() {
		var p empresaBrebQRPago
		p.Origen = "venta_carrito"
		if err := carritos.Scan(&p.ID, &p.Codigo, &p.ReferenciaCaja, &p.CajaCodigo, &p.CajaTurno, &p.MetodoPago, &p.ReferenciaPago, &p.Monto, &p.Moneda, &p.Fecha, &p.UsuarioCreador, &p.Estado); err != nil {
			return nil, err
		}
		p.CarritoID = p.ID
		pagos = append(pagos, p)
	}
	if err := carritos.Err(); err != nil {
		return nil, err
	}

	abonos, err := dbpkg.ExecQueryCompat(dbEmp, `SELECT
		a.id, a.carrito_id, COALESCE(c.codigo,''), COALESCE(c.referencia_externa,''), COALESCE(a.caja_codigo,''), COALESCE(a.caja_turno,''),
		COALESCE(a.metodo_pago,''), COALESCE(a.referencia_pago,''), COALESCE(a.monto,0), COALESCE(c.moneda,'COP'),
		COALESCE(a.fecha_abono, a.fecha_creacion, ''), COALESCE(a.usuario_creador,''), COALESCE(a.estado,'activo')
	FROM carrito_compra_abonos a
	JOIN carritos_compras c ON c.empresa_id = a.empresa_id AND c.id = a.carrito_id
	WHERE a.empresa_id = ?
	  AND LOWER(COALESCE(a.estado,'activo')) = 'activo'
	  AND LOWER(COALESCE(a.metodo_pago,'')) = 'transferencia_bre_b'
	ORDER BY `+brebQRAbonosOrderExpr+` DESC, a.id DESC
	LIMIT ?`, empresaID, limit)
	if err != nil {
		return nil, err
	}
	defer abonos.Close()
	for abonos.Next() {
		var p empresaBrebQRPago
		p.Origen = "abono_carrito"
		if err := abonos.Scan(&p.ID, &p.CarritoID, &p.Codigo, &p.ReferenciaCaja, &p.CajaCodigo, &p.CajaTurno, &p.MetodoPago, &p.ReferenciaPago, &p.Monto, &p.Moneda, &p.Fecha, &p.UsuarioCreador, &p.Estado); err != nil {
			return nil, err
		}
		pagos = append(pagos, p)
	}
	if err := abonos.Err(); err != nil {
		return nil, err
	}

	bancos, err := dbpkg.ExecQueryCompat(dbEmp, `SELECT
		id, COALESCE(cuenta_bancaria,''), COALESCE(banco_nombre,''), COALESCE(referencia_bancaria,''), COALESCE(documento_codigo,''),
		COALESCE(monto, total, 0), COALESCE(moneda,'COP'), COALESCE(fecha_movimiento, fecha_creacion, ''), COALESCE(usuario_creador,''),
		COALESCE(estado,'activo'), COALESCE(estado_conciliacion,'pendiente'), COALESCE(observaciones,'')
	FROM empresa_finanzas_bancos_movimientos
	WHERE empresa_id = ?
	  AND LOWER(COALESCE(estado,'activo')) = 'activo'
	  AND (LOWER(COALESCE(origen,'')) LIKE 'breb%' OR LOWER(COALESCE(descripcion,'')) LIKE '%bre-b%' OR LOWER(COALESCE(referencia_bancaria,'')) LIKE 'breb-%' OR LOWER(COALESCE(referencia_bancaria,'')) LIKE 'qr-breb-%')
	ORDER BY `+brebQRBancosOrderExpr+` DESC, id DESC
	LIMIT ?`, empresaID, limit)
	if err != nil {
		return nil, err
	}
	defer bancos.Close()
	for bancos.Next() {
		var cuenta, banco, documento string
		var p empresaBrebQRPago
		p.Origen = "registro_bancario"
		p.MetodoPago = "transferencia_bre_b"
		if err := bancos.Scan(&p.ID, &cuenta, &banco, &p.ReferenciaPago, &documento, &p.Monto, &p.Moneda, &p.Fecha, &p.UsuarioCreador, &p.Estado, &p.EstadoConciliado, &p.Observaciones); err != nil {
			return nil, err
		}
		p.Codigo = documento
		p.ReferenciaCaja = cuenta
		if banco != "" {
			p.Observaciones = strings.TrimSpace(strings.TrimSpace(p.Observaciones) + " " + banco)
		}
		pagos = append(pagos, p)
	}
	if err := bancos.Err(); err != nil {
		return nil, err
	}

	sort.SliceStable(pagos, func(i, j int) bool {
		return pagos[i].Fecha > pagos[j].Fecha
	})
	if len(pagos) > limit {
		pagos = pagos[:limit]
	}
	return pagos, nil
}

func summarizeEmpresaBrebQRPagos(pagos []empresaBrebQRPago) map[string]interface{} {
	resumen := map[string]interface{}{
		"total_registros":                       len(pagos),
		"total_monto":                           0.0,
		"ventas_carrito":                        0,
		"abonos_carrito":                        0,
		"registros_bancarios":                   0,
		"pendientes_conciliar":                  0,
		"confirmados_bancarios":                 0,
		"registrados_sin_confirmacion_bancaria": 0,
	}
	total := 0.0
	for _, p := range pagos {
		total += p.Monto
		switch p.Origen {
		case "venta_carrito":
			resumen["ventas_carrito"] = resumen["ventas_carrito"].(int) + 1
			resumen["registrados_sin_confirmacion_bancaria"] = resumen["registrados_sin_confirmacion_bancaria"].(int) + 1
		case "abono_carrito":
			resumen["abonos_carrito"] = resumen["abonos_carrito"].(int) + 1
			resumen["registrados_sin_confirmacion_bancaria"] = resumen["registrados_sin_confirmacion_bancaria"].(int) + 1
		case "registro_bancario":
			resumen["registros_bancarios"] = resumen["registros_bancarios"].(int) + 1
			if strings.EqualFold(strings.TrimSpace(p.EstadoConciliado), "conciliado") {
				resumen["confirmados_bancarios"] = resumen["confirmados_bancarios"].(int) + 1
			} else {
				resumen["pendientes_conciliar"] = resumen["pendientes_conciliar"].(int) + 1
			}
		}
	}
	resumen["total_monto"] = mathRound2(total)
	return resumen
}

func insertEmpresaBrebQRRegistroManual(dbEmp *sql.DB, payload empresaBrebQRRegistroManualPayload, usuario string) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	if payload.Monto <= 0 || math.IsNaN(payload.Monto) || math.IsInf(payload.Monto, 0) {
		return 0, fmt.Errorf("monto debe ser mayor a cero")
	}
	ref := strings.TrimSpace(payload.ReferenciaBancaria)
	if len(ref) < 4 || len(ref) > 80 || !brebQRReferencePattern.MatchString(ref) {
		return 0, fmt.Errorf("referencia_bancaria es obligatoria")
	}
	fecha, err := normalizeBrebQRMovementDate(payload.FechaMovimiento)
	if err != nil {
		return 0, err
	}
	moneda := strings.ToUpper(strings.TrimSpace(payload.Moneda))
	if moneda == "" {
		moneda = "COP"
	}
	if moneda != "COP" {
		return 0, fmt.Errorf("Bre-B Colombia solo admite moneda COP")
	}
	estadoConciliacion := strings.ToLower(strings.TrimSpace(payload.EstadoConciliacion))
	if estadoConciliacion != "" && estadoConciliacion != "pendiente" {
		return 0, fmt.Errorf("un registro manual debe iniciar pendiente de conciliacion bancaria")
	}
	estadoConciliacion = "pendiente"
	if payload.CarritoID > 0 {
		carrito, err := dbpkg.GetCarritoCompraByID(dbEmp, payload.EmpresaID, payload.CarritoID)
		if err != nil || carrito == nil {
			return 0, fmt.Errorf("el carrito indicado no pertenece a esta empresa")
		}
		expected := carrito.TotalPagado
		if expected <= 0 {
			expected = carrito.Total
		}
		if math.Abs(mathRound2(expected)-mathRound2(payload.Monto)) > 0.01 {
			return 0, fmt.Errorf("el monto no coincide con el total registrado para el carrito")
		}
		cartCurrency := strings.ToUpper(strings.TrimSpace(carrito.Moneda))
		if cartCurrency != "" && cartCurrency != "COP" {
			return 0, fmt.Errorf("la moneda del carrito no corresponde a Bre-B Colombia")
		}
	}
	hash := hashEmpresaBrebQRMovimiento(payload.EmpresaID, fecha, ref, payload.Monto, payload.CajaCodigo, payload.CarritoID)
	obs := map[string]interface{}{
		"caja_codigo":       strings.TrimSpace(payload.CajaCodigo),
		"caja_turno":        strings.TrimSpace(payload.CajaTurno),
		"estacion_id":       payload.EstacionID,
		"carrito_id":        payload.CarritoID,
		"pagador_nombre":    strings.TrimSpace(payload.PagadorNombre),
		"pagador_documento": strings.TrimSpace(payload.PagadorDocumento),
		"observaciones":     strings.TrimSpace(payload.Observaciones),
	}
	obsBlob, _ := json.Marshal(obs)
	id := int64(0)
	err = dbpkg.QueryRowCompat(dbEmp, `INSERT INTO empresa_finanzas_bancos_movimientos (
		empresa_id, periodo_contable, fecha_movimiento, fecha_valor, cuenta_bancaria, banco_nombre, tipo_movimiento,
		descripcion, referencia_bancaria, documento_codigo, moneda, monto, total, estado_conciliacion, origen,
		hash_movimiento, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
	) VALUES (
		?, '', ?, ?, ?, ?, 'ingreso',
		'Pago Bre-B QR registrado manualmente', ?, ?, ?, ?, ?, ?, 'breb_qr_manual',
		?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, 'activo', ?
	) ON CONFLICT(empresa_id, hash_movimiento) DO UPDATE SET
		estado_conciliacion = excluded.estado_conciliacion,
		fecha_actualizacion = CURRENT_TIMESTAMP,
		usuario_creador = excluded.usuario_creador,
		observaciones = excluded.observaciones
	RETURNING id`,
		payload.EmpresaID, fecha, fecha, strings.TrimSpace(payload.CuentaBancaria), strings.TrimSpace(payload.BancoNombre),
		ref, strings.TrimSpace(payload.DocumentoCodigo), moneda, payload.Monto, payload.Monto, estadoConciliacion, hash,
		strings.TrimSpace(usuario), string(obsBlob)).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("no se pudo registrar pago Bre-B")
	}
	return id, nil
}

func hashEmpresaBrebQRMovimiento(empresaID int64, fecha, referencia string, monto float64, caja string, carritoID int64) string {
	base := strconv.FormatInt(empresaID, 10) + "|" + strings.ToLower(strings.TrimSpace(referencia))
	sum := sha256.Sum256([]byte(base))
	return "breb_qr_" + hex.EncodeToString(sum[:])
}

func normalizeBrebQRCuentas(raw interface{}) []map[string]interface{} {
	rows, ok := raw.([]map[string]interface{})
	if !ok {
		if asList, ok := raw.([]interface{}); ok {
			rows = make([]map[string]interface{}, 0, len(asList))
			for _, item := range asList {
				if m, ok := item.(map[string]interface{}); ok {
					rows = append(rows, m)
				}
			}
		}
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for i, row := range rows {
		proveedor := strings.ToLower(stringFromInterfaceDefault(row["proveedor"], "breb"))
		if proveedor != "breb" && proveedor != "nequi" && proveedor != "otro" {
			proveedor = "breb"
		}
		item := map[string]interface{}{
			"id":              stringFromInterfaceDefault(row["id"], fmt.Sprintf("qr_account_%d", i+1)),
			"activa":          boolFromInterface(row["activa"], true),
			"nombre":          stringFromInterfaceDefault(row["nombre"], ""),
			"proveedor":       proveedor,
			"tipo_llave":      stringFromInterfaceDefault(row["tipo_llave"], ""),
			"llave":           stringFromInterfaceDefault(row["llave"], ""),
			"comercio":        stringFromInterfaceDefault(row["comercio"], ""),
			"caja_codigo":     stringFromInterfaceDefault(row["caja_codigo"], ""),
			"payload_oficial": stringFromInterfaceDefault(row["payload_oficial"], ""),
			"instrucciones":   stringFromInterfaceDefault(row["instrucciones"], ""),
			"cuenta_contable": stringFromInterfaceDefault(row["cuenta_contable"], ""),
			"banco_receptor":  stringFromInterfaceDefault(row["banco_receptor"], ""),
			"referencia_fija": stringFromInterfaceDefault(row["referencia_fija"], ""),
			"qr_tipo":         stringFromInterfaceDefault(row["qr_tipo"], "estatico"),
		}
		if strings.TrimSpace(fmt.Sprint(item["nombre"], item["llave"], item["payload_oficial"], item["caja_codigo"])) == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

func validateAndNormalizeBrebQRCuentas(raw []map[string]interface{}) ([]map[string]interface{}, error) {
	if len(raw) > 50 {
		return nil, fmt.Errorf("se permiten maximo 50 cuentas receptoras")
	}
	normalized := normalizeBrebQRCuentas(raw)
	for index, row := range normalized {
		provider := strings.ToLower(stringFromInterfaceDefault(row["proveedor"], "breb"))
		keyType := normalizeBrebQRKeyType(stringFromInterfaceDefault(row["tipo_llave"], ""))
		key := stringFromInterfaceDefault(row["llave"], "")
		payload := stringFromInterfaceDefault(row["payload_oficial"], "")
		qrType := strings.ToLower(stringFromInterfaceDefault(row["qr_tipo"], "estatico"))
		if qrType != "estatico" {
			return nil, fmt.Errorf("cuenta %d: el QR dinamico requiere API oficial del participante Bre-B; solo se admite payload estatico emitido por el proveedor", index+1)
		}
		if len(payload) > 4096 || strings.ContainsAny(payload, "\r\n") {
			return nil, fmt.Errorf("cuenta %d: payload QR invalido o demasiado largo", index+1)
		}
		if strings.Contains(payload, "{") || strings.Contains(payload, "}") {
			return nil, fmt.Errorf("cuenta %d: no se permiten plantillas en un payload QR oficial; usa el contenido estatico exacto", index+1)
		}
		if key != "" {
			if keyType == "" {
				keyType = inferBrebQRKeyType(key, provider)
			}
			if err := validateBrebQRKey(keyType, key); err != nil {
				return nil, fmt.Errorf("cuenta %d: %w", index+1, err)
			}
		}
		if boolFromInterface(row["activa"], true) && key == "" && payload == "" {
			return nil, fmt.Errorf("cuenta %d: una cuenta activa requiere llave o payload QR oficial", index+1)
		}
		row["proveedor"] = provider
		row["tipo_llave"] = keyType
		row["llave"] = key
		row["payload_oficial"] = payload
		row["qr_tipo"] = "estatico"
	}
	return normalized, nil
}

func normalizeBrebQRKeyType(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	v = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u").Replace(v)
	switch v {
	case "celular", "telefono", "movil", "nequi":
		return "celular"
	case "correo", "email", "correo_electronico":
		return "correo"
	case "documento", "nit", "cedula", "identificacion":
		return "documento"
	case "alfanumerica", "usuario", "alias":
		return "alfanumerica"
	case "comercio", "codigo_comercio":
		return "comercio"
	default:
		return ""
	}
}

func inferBrebQRKeyType(key, provider string) string {
	key = strings.TrimSpace(key)
	if strings.Contains(key, "@") && !strings.HasPrefix(key, "@") {
		return "correo"
	}
	if strings.HasPrefix(key, "@") {
		return "alfanumerica"
	}
	if len(key) == 10 && strings.HasPrefix(key, "00") {
		return "comercio"
	}
	if len(key) == 10 && strings.HasPrefix(key, "3") {
		return "celular"
	}
	if provider == "nequi" {
		return "celular"
	}
	return "documento"
}

func validateBrebQRKey(keyType, key string) error {
	key = strings.TrimSpace(key)
	switch keyType {
	case "celular":
		if len(key) != 10 || key[0] != '3' || !regexp.MustCompile(`^[0-9]{10}$`).MatchString(key) {
			return fmt.Errorf("la llave celular debe tener 10 digitos y comenzar por 3")
		}
	case "correo":
		parts := strings.Split(key, "@")
		if len(key) > 120 || len(parts) != 2 || parts[0] == "" || !strings.Contains(parts[1], ".") || strings.ContainsAny(key, " \t\r\n") {
			return fmt.Errorf("la llave de correo no tiene un formato valido")
		}
	case "documento":
		if len(key) > 18 || !regexp.MustCompile(`^[A-Za-z0-9]{1,18}$`).MatchString(key) {
			return fmt.Errorf("la llave de documento admite hasta 18 caracteres alfanumericos sin puntuacion")
		}
	case "alfanumerica":
		if len(key) < 5 || len(key) > 20 || !regexp.MustCompile(`^@[A-Za-z0-9._-]{4,19}$`).MatchString(key) {
			return fmt.Errorf("la llave alfanumerica debe iniciar por @ y tener entre 5 y 20 caracteres")
		}
	case "comercio":
		if !regexp.MustCompile(`^00[0-9]{8}$`).MatchString(key) {
			return fmt.Errorf("la llave de comercio debe tener 10 digitos y comenzar por 00")
		}
	default:
		return fmt.Errorf("tipo_llave no reconocido para Bre-B Colombia")
	}
	return nil
}

func normalizeBrebQRMovementDate(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Now().Format("2006-01-02 15:04:05"), nil
	}
	var parsed time.Time
	var err error
	for _, layout := range []string{"2006-01-02 15:04:05", "2006-01-02 15:04", time.RFC3339} {
		parsed, err = time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			break
		}
	}
	if err != nil || parsed.IsZero() {
		return "", fmt.Errorf("fecha_movimiento invalida")
	}
	if parsed.After(time.Now().Add(5 * time.Minute)) {
		return "", fmt.Errorf("fecha_movimiento no puede estar en el futuro")
	}
	return parsed.Format("2006-01-02 15:04:05"), nil
}

func boolFromInterface(value interface{}, fallback bool) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "1", "true", "si", "sí", "yes", "activo", "on":
			return true
		case "0", "false", "no", "inactivo", "off":
			return false
		}
	case float64:
		return v != 0
	case int:
		return v != 0
	case int64:
		return v != 0
	}
	return fallback
}

func stringFromInterfaceDefault(value interface{}, fallback string) string {
	if value == nil {
		return fallback
	}
	out := strings.TrimSpace(fmt.Sprint(value))
	if out == "" || strings.EqualFold(out, "<nil>") {
		return fallback
	}
	return out
}

func intFromInterfaceDefault(value interface{}, fallback int) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err == nil {
			return n
		}
	}
	return fallback
}

func mathRound2(value float64) float64 {
	return float64(int64(value*100+0.5)) / 100
}
