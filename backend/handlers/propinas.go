package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaPropinasHandler gestiona configuracion, reportes y movimientos de propinas por empresa.
func EmpresaPropinasHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			switch action {
			case "", "config", "configuracion":
				cfg, err := dbpkg.GetEmpresaPropinasConfiguracion(dbEmp, empresaID)
				if err != nil {
					log.Printf("[propinas] get config empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo consultar la configuracion de propinas", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, cfg)
				return

			case "reporte", "resumen", "dashboard":
				limit, err := parseIntQueryOptional(r, "limit")
				if err != nil {
					http.Error(w, "limit invalido", http.StatusBadRequest)
					return
				}
				cierreCajaID, err := parseInt64QueryOptional(r, "cierre_caja_id")
				if err != nil {
					http.Error(w, "cierre_caja_id invalido", http.StatusBadRequest)
					return
				}
				report, err := dbpkg.GetEmpresaPropinasReporte(dbEmp, empresaID, dbpkg.EmpresaPropinaMovimientoFilter{
					Desde:            strings.TrimSpace(r.URL.Query().Get("desde")),
					Hasta:            strings.TrimSpace(r.URL.Query().Get("hasta")),
					ModoDistribucion: strings.TrimSpace(r.URL.Query().Get("modo")),
					Usuario:          strings.TrimSpace(r.URL.Query().Get("usuario")),
					OrigenMovimiento: strings.TrimSpace(r.URL.Query().Get("origen")),
					CierreCajaID:     cierreCajaID,
					SoloAjustes:      queryBool(r, "solo_ajustes"),
					IncludeInactive:  queryBool(r, "include_inactive"),
					Limit:            limit,
				})
				if err != nil {
					log.Printf("[propinas] get report empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo construir el reporte de propinas", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, report)
				return

			case "movimientos", "detalle":
				limit, err := parseIntQueryOptional(r, "limit")
				if err != nil {
					http.Error(w, "limit invalido", http.StatusBadRequest)
					return
				}
				cierreCajaID, err := parseInt64QueryOptional(r, "cierre_caja_id")
				if err != nil {
					http.Error(w, "cierre_caja_id invalido", http.StatusBadRequest)
					return
				}
				rows, err := dbpkg.ListEmpresaPropinaMovimientos(dbEmp, empresaID, dbpkg.EmpresaPropinaMovimientoFilter{
					Desde:            strings.TrimSpace(r.URL.Query().Get("desde")),
					Hasta:            strings.TrimSpace(r.URL.Query().Get("hasta")),
					ModoDistribucion: strings.TrimSpace(r.URL.Query().Get("modo")),
					Usuario:          strings.TrimSpace(r.URL.Query().Get("usuario")),
					OrigenMovimiento: strings.TrimSpace(r.URL.Query().Get("origen")),
					CierreCajaID:     cierreCajaID,
					SoloAjustes:      queryBool(r, "solo_ajustes"),
					IncludeInactive:  queryBool(r, "include_inactive"),
					Limit:            limit,
				})
				if err != nil {
					log.Printf("[propinas] list movimientos empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo listar movimientos de propinas", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return

			case "conciliacion_cierre", "conciliar_cierre", "conciliar":
				cierreCajaID, err := parseInt64QueryOptional(r, "cierre_caja_id")
				if err != nil || cierreCajaID <= 0 {
					http.Error(w, "cierre_caja_id es obligatorio", http.StatusBadRequest)
					return
				}
				conciliacion, err := dbpkg.ConciliarEmpresaPropinasConCierreCaja(dbEmp, empresaID, cierreCajaID, strings.TrimSpace(adminEmailFromRequest(r)))
				if err != nil {
					log.Printf("[propinas] conciliar cierre empresa_id=%d cierre_id=%d error: %v", empresaID, cierreCajaID, err)
					status := propinasWriteStatus(err)
					http.Error(w, err.Error(), status)
					return
				}
				writeJSON(w, http.StatusOK, conciliacion)
				return

			default:
				http.Error(w, "action invalida. Use: config, reporte, movimientos o conciliacion_cierre", http.StatusBadRequest)
				return
			}

		case http.MethodPost, http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "ajuste_manual" {
				var payload struct {
					EmpresaID         int64   `json:"empresa_id"`
					CarritoID         int64   `json:"carrito_id"`
					CierreCajaID      int64   `json:"cierre_caja_id"`
					VentaReferencia   string  `json:"venta_referencia"`
					UsuarioAsignado   string  `json:"usuario_asignado"`
					UsuarioAsignadoID int64   `json:"usuario_asignado_id"`
					ModoDistribucion  string  `json:"modo_distribucion"`
					MontoAjuste       float64 `json:"monto_ajuste"`
					MontoPropina      float64 `json:"monto_propina"`
					Motivo            string  `json:"motivo"`
					ReferenciaAjuste  string  `json:"referencia_ajuste"`
					Moneda            string  `json:"moneda"`
					BaseCobro         float64 `json:"base_cobro"`
					FiscalPais        string  `json:"fiscal_pais"`
					FiscalRegimen     string  `json:"fiscal_regimen"`
					FiscalTratamiento string  `json:"fiscal_tratamiento"`
					FiscalPorcentaje  float64 `json:"fiscal_porcentaje_impuesto"`
				}
				if r.Body != nil {
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
						http.Error(w, "JSON invalido", http.StatusBadRequest)
						return
					}
				}

				if payload.EmpresaID <= 0 {
					if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
						payload.EmpresaID = empresaID
					}
				}
				if payload.EmpresaID <= 0 {
					http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
					return
				}
				if payload.CierreCajaID <= 0 {
					if cierreID, err := parseInt64QueryOptional(r, "cierre_caja_id"); err == nil && cierreID > 0 {
						payload.CierreCajaID = cierreID
					}
				}

				montoAjuste := payload.MontoAjuste
				if math.Abs(montoAjuste) < 0.0001 {
					montoAjuste = payload.MontoPropina
				}
				if math.Abs(montoAjuste) < 0.0001 {
					http.Error(w, "monto_ajuste es obligatorio y debe ser diferente de cero", http.StatusBadRequest)
					return
				}

				motivo := strings.TrimSpace(payload.Motivo)
				if motivo == "" {
					http.Error(w, "motivo es obligatorio para ajuste manual", http.StatusBadRequest)
					return
				}

				usuarioOperacion := strings.TrimSpace(adminEmailFromRequest(r))
				mov := dbpkg.EmpresaPropinaMovimiento{
					EmpresaID:         payload.EmpresaID,
					CarritoID:         payload.CarritoID,
					CierreCajaID:      payload.CierreCajaID,
					VentaReferencia:   strings.TrimSpace(payload.VentaReferencia),
					UsuarioOrigen:     usuarioOperacion,
					UsuarioAsignado:   strings.TrimSpace(payload.UsuarioAsignado),
					UsuarioAsignadoID: payload.UsuarioAsignadoID,
					ModoDistribucion:  strings.TrimSpace(payload.ModoDistribucion),
					Moneda:            strings.TrimSpace(payload.Moneda),
					BaseCobro:         payload.BaseCobro,
					PorcentajePropina: 0,
					MontoPropina:      montoAjuste,
					ReferenciaAjuste:  strings.TrimSpace(payload.ReferenciaAjuste),
					FiscalPais:        strings.TrimSpace(payload.FiscalPais),
					FiscalRegimen:     strings.TrimSpace(payload.FiscalRegimen),
					FiscalTratamiento: strings.TrimSpace(payload.FiscalTratamiento),
					FiscalPorcentaje:  payload.FiscalPorcentaje,
					UsuarioCreador:    usuarioOperacion,
					Estado:            "activo",
					Observaciones:     motivo,
				}

				id, err := dbpkg.CreateEmpresaPropinaAjusteManual(dbEmp, mov)
				if err != nil {
					log.Printf("[propinas] ajuste manual empresa_id=%d error: %v", payload.EmpresaID, err)
					http.Error(w, err.Error(), propinasWriteStatus(err))
					return
				}

				var conciliacion *dbpkg.EmpresaPropinaConciliacionCierre
				if payload.CierreCajaID > 0 {
					conciliacion, err = dbpkg.ConciliarEmpresaPropinasConCierreCaja(dbEmp, payload.EmpresaID, payload.CierreCajaID, usuarioOperacion)
					if err != nil {
						log.Printf("[propinas] ajuste manual conciliar empresa_id=%d cierre_id=%d error: %v", payload.EmpresaID, payload.CierreCajaID, err)
						http.Error(w, "No se pudo conciliar propinas para el cierre indicado", http.StatusInternalServerError)
						return
					}
				}

				registrarAuditoriaPropinaAjusteNoBloqueante(dbEmp, r, payload.EmpresaID, id, payload.CierreCajaID, montoAjuste, motivo)

				resp := map[string]interface{}{
					"ok":             true,
					"id":             id,
					"empresa_id":     payload.EmpresaID,
					"cierre_caja_id": payload.CierreCajaID,
					"monto_ajuste":   montoAjuste,
					"motivo":         motivo,
				}
				if conciliacion != nil {
					resp["conciliacion"] = conciliacion
				}
				writeJSON(w, http.StatusOK, resp)
				return
			}

			var payload dbpkg.EmpresaPropinasConfiguracion
			if r.Body != nil {
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
			}

			if payload.EmpresaID <= 0 {
				if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
					payload.EmpresaID = empresaID
				}
			}
			if payload.EmpresaID <= 0 {
				http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
				return
			}
			if action == "activar" {
				payload.HabilitarPropina = true
			}
			if action == "desactivar" {
				payload.HabilitarPropina = false
			}
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))

			id, err := dbpkg.UpsertEmpresaPropinasConfiguracion(dbEmp, payload)
			if err != nil {
				log.Printf("[propinas] upsert config empresa_id=%d error: %v", payload.EmpresaID, err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			cfg, err := dbpkg.GetEmpresaPropinasConfiguracion(dbEmp, payload.EmpresaID)
			if err != nil {
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":            true,
				"id":            id,
				"configuracion": cfg,
			})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

func propinasWriteStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if errors.Is(err, sql.ErrNoRows) {
		return http.StatusNotFound
	}
	lower := strings.ToLower(strings.TrimSpace(err.Error()))
	if strings.Contains(lower, "obligatorio") ||
		strings.Contains(lower, "invalido") ||
		strings.Contains(lower, "debe") ||
		strings.Contains(lower, "anulado") {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

func registrarAuditoriaPropinaAjusteNoBloqueante(dbEmp *sql.DB, r *http.Request, empresaID, recursoID, cierreCajaID int64, monto float64, motivo string) {
	if dbEmp == nil || empresaID <= 0 {
		return
	}
	metadata, _ := json.Marshal(map[string]interface{}{
		"cierre_caja_id": cierreCajaID,
		"monto_ajuste":   monto,
		"motivo":         motivo,
	})
	usuario := strings.TrimSpace(adminEmailFromRequest(r))
	if usuario == "" {
		usuario = "sistema"
	}
	requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
	if requestID == "" {
		requestID = strings.TrimSpace(r.Header.Get("X-Request-Id"))
	}
	if _, err := dbpkg.CreateEmpresaAuditoriaEvento(dbEmp, dbpkg.EmpresaAuditoriaEvento{
		EmpresaID:      empresaID,
		Modulo:         "propinas",
		Accion:         "ajuste_manual",
		Recurso:        "empresa_propinas_movimientos",
		RecursoID:      recursoID,
		MetodoHTTP:     r.Method,
		Endpoint:       r.URL.Path,
		Resultado:      "ok",
		CodigoHTTP:     http.StatusOK,
		RequestID:      requestID,
		IPOrigen:       strings.TrimSpace(r.RemoteAddr),
		UserAgent:      strings.TrimSpace(r.UserAgent()),
		MetadataJSON:   string(metadata),
		UsuarioCreador: usuario,
		Observaciones:  strings.TrimSpace(motivo),
	}); err != nil {
		log.Printf("[propinas] auditoria ajuste manual empresa_id=%d recurso_id=%d error: %v", empresaID, recursoID, err)
	}
}
