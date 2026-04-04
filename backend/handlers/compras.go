package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaComprasDocumentosHandler gestiona ciclo documental general de compras (orden, recepcion y contabilizacion).
func EmpresaComprasDocumentosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			tipoDocumento := strings.TrimSpace(r.URL.Query().Get("tipo_documento"))
			estadoDocumento := strings.TrimSpace(r.URL.Query().Get("estado_documento"))
			proveedorID, err := parseInt64QueryOptional(r, "proveedor_id")
			if err != nil {
				http.Error(w, "proveedor_id invalido", http.StatusBadRequest)
				return
			}
			includeInactive := parseBoolQuery(r, "include_inactive")
			q := strings.TrimSpace(r.URL.Query().Get("q"))
			limit, err := parseIntQueryOptional(r, "limit")
			if err != nil {
				http.Error(w, "limit invalido", http.StatusBadRequest)
				return
			}
			offset, err := parseIntQueryOptional(r, "offset")
			if err != nil {
				http.Error(w, "offset invalido", http.StatusBadRequest)
				return
			}

			rows, err := dbpkg.ListEmpresaDocumentosCompraByEmpresa(dbEmp, empresaID, tipoDocumento, proveedorID, estadoDocumento, includeInactive, q, limit, offset)
			if err != nil {
				http.Error(w, "No se pudo listar documentos de compras", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(rows)
			return

		case http.MethodPost:
			var payload struct {
				EmpresaID       int64   `json:"empresa_id"`
				ProveedorID     int64   `json:"proveedor_id"`
				TipoDocumento   string  `json:"tipo_documento"`
				DocumentoCodigo string  `json:"documento_codigo"`
				EstadoDocumento string  `json:"estado_documento"`
				EstadoActual    string  `json:"estado_actual"`
				Accion          string  `json:"accion"`
				PeriodoContable string  `json:"periodo_contable"`
				MontoTotal      float64 `json:"monto_total"`
				Moneda          string  `json:"moneda"`
				FechaDocumento  string  `json:"fecha_documento"`
				Observaciones   string  `json:"observaciones"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}

			if payload.EmpresaID <= 0 {
				empresaID, err := parseInt64QueryOptional(r, "empresa_id")
				if err != nil {
					http.Error(w, "empresa_id invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
			}
			if payload.EmpresaID <= 0 {
				http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
				return
			}
			if payload.ProveedorID <= 0 {
				http.Error(w, "proveedor_id es obligatorio", http.StatusBadRequest)
				return
			}

			if strings.TrimSpace(payload.DocumentoCodigo) == "" {
				payload.DocumentoCodigo = fmt.Sprintf("OC-%d-%s", payload.ProveedorID, time.Now().Format("20060102150405"))
			}
			if strings.TrimSpace(payload.PeriodoContable) == "" {
				payload.PeriodoContable = time.Now().Format("2006-01")
			}
			if strings.TrimSpace(payload.FechaDocumento) == "" {
				payload.FechaDocumento = time.Now().Format("2006-01-02")
			}

			action := strings.ToLower(strings.TrimSpace(payload.Accion))
			if action == "" {
				action = strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			}
			if action == "" {
				action = "crear"
			}

			estadoDocumento := strings.TrimSpace(payload.EstadoDocumento)
			if estadoDocumento == "" {
				estadoDocumento = "borrador"
			}
			estadoAnterior := ""
			evento := "orden_compra_creada"
			accionResp := "crear"

			if action != "crear" && action != "guardar" {
				transition, err := resolveComprasTransition(action, comprasFirstNonBlank(payload.EstadoActual, estadoDocumento, "borrador"))
				if err != nil {
					errLower := strings.ToLower(err.Error())
					if strings.Contains(errLower, "transicion invalida") {
						http.Error(w, err.Error(), http.StatusConflict)
						return
					}
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				estadoDocumento = transition.EstadoNuevo
				estadoAnterior = transition.EstadoAnterior
				evento = transition.Evento
				accionResp = transition.Accion
			}

			docPersistido, err := dbpkg.UpsertEmpresaDocumentoCompra(dbEmp, dbpkg.EmpresaDocumentoCompra{
				EmpresaID:            payload.EmpresaID,
				ProveedorID:          payload.ProveedorID,
				TipoDocumento:        comprasFirstNonBlank(payload.TipoDocumento, "orden_compra"),
				DocumentoCodigo:      payload.DocumentoCodigo,
				EstadoDocumento:      estadoDocumento,
				EstadoAnterior:       estadoAnterior,
				EventoUltimo:         evento,
				PeriodoContable:      payload.PeriodoContable,
				MontoTotal:           payload.MontoTotal,
				Moneda:               payload.Moneda,
				FechaDocumento:       payload.FechaDocumento,
				EntidadRelacionadaID: payload.ProveedorID,
				UsuarioCreador:       adminEmailFromRequest(r),
				Estado:               "activo",
				Observaciones:        payload.Observaciones,
			})
			if err != nil {
				http.Error(w, "No se pudo guardar el documento de compras", http.StatusInternalServerError)
				return
			}

			registrarEventoContableNoBloqueante(dbEmp, r, "compras", dbpkg.EmpresaEventoContable{
				EmpresaID:       payload.EmpresaID,
				Modulo:          "compras",
				Evento:          evento,
				Entidad:         "orden_compra",
				EntidadID:       docPersistido.ID,
				DocumentoTipo:   docPersistido.TipoDocumento,
				DocumentoCodigo: docPersistido.DocumentoCodigo,
				PeriodoContable: docPersistido.PeriodoContable,
				MontoTotal:      docPersistido.MontoTotal,
				Moneda:          docPersistido.Moneda,
				Origen:          "api_compras_documentos",
				Observaciones:   strings.TrimSpace(payload.Observaciones),
			}, map[string]interface{}{
				"accion":           accionResp,
				"estado_anterior":  docPersistido.EstadoAnterior,
				"estado_nuevo":     docPersistido.EstadoDocumento,
				"entidad_id":       docPersistido.ID,
				"documento_codigo": docPersistido.DocumentoCodigo,
				"proveedor_id":     docPersistido.ProveedorID,
				"empresa_id":       docPersistido.EmpresaID,
			})

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":        true,
				"accion":    accionResp,
				"evento":    evento,
				"resultado": docPersistido,
			})
			return

		case http.MethodPut:
			var payload struct {
				EmpresaID       int64   `json:"empresa_id"`
				ProveedorID     int64   `json:"proveedor_id"`
				TipoDocumento   string  `json:"tipo_documento"`
				DocumentoCodigo string  `json:"documento_codigo"`
				EstadoActual    string  `json:"estado_actual"`
				EstadoDocumento string  `json:"estado_documento"`
				Accion          string  `json:"accion"`
				PeriodoContable string  `json:"periodo_contable"`
				MontoTotal      float64 `json:"monto_total"`
				Moneda          string  `json:"moneda"`
				FechaDocumento  string  `json:"fecha_documento"`
				Observaciones   string  `json:"observaciones"`
				Activo          *bool   `json:"activo"`
			}
			if r.Body != nil {
				_ = json.NewDecoder(r.Body).Decode(&payload)
			}

			if payload.EmpresaID <= 0 {
				empresaID, err := parseInt64QueryOptional(r, "empresa_id")
				if err != nil {
					http.Error(w, "empresa_id invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
			}
			if payload.EmpresaID <= 0 {
				http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
				return
			}

			if strings.TrimSpace(payload.DocumentoCodigo) == "" {
				payload.DocumentoCodigo = strings.TrimSpace(r.URL.Query().Get("documento_codigo"))
			}
			if strings.TrimSpace(payload.DocumentoCodigo) == "" {
				http.Error(w, "documento_codigo es obligatorio", http.StatusBadRequest)
				return
			}

			action := strings.ToLower(strings.TrimSpace(payload.Accion))
			if action == "" {
				action = strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			}
			if action == "" {
				action = "actualizar"
			}

			tipoDocumento := comprasFirstNonBlank(payload.TipoDocumento, "orden_compra")

			if action == "activar" {
				estado := "activo"
				if payload.Activo != nil {
					if !*payload.Activo {
						estado = "inactivo"
					}
				} else {
					activoRaw := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("activo")))
					if activoRaw == "0" || activoRaw == "false" || activoRaw == "no" {
						estado = "inactivo"
					}
				}

				if err := dbpkg.SetEmpresaDocumentoCompraEstadoByCodigo(dbEmp, payload.EmpresaID, tipoDocumento, payload.DocumentoCodigo, estado); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "documento no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, "No se pudo actualizar estado activo del documento", http.StatusInternalServerError)
					return
				}

				evento := "orden_compra_activada"
				if estado == "inactivo" {
					evento = "orden_compra_desactivada"
				}
				registrarEventoContableNoBloqueante(dbEmp, r, "compras", dbpkg.EmpresaEventoContable{
					EmpresaID:       payload.EmpresaID,
					Modulo:          "compras",
					Evento:          evento,
					Entidad:         "orden_compra",
					DocumentoTipo:   tipoDocumento,
					DocumentoCodigo: strings.ToUpper(strings.TrimSpace(payload.DocumentoCodigo)),
					Origen:          "api_compras_documentos",
					Observaciones:   strings.TrimSpace(payload.Observaciones),
				}, map[string]interface{}{
					"accion":           "activar",
					"estado":           estado,
					"documento_codigo": strings.ToUpper(strings.TrimSpace(payload.DocumentoCodigo)),
					"empresa_id":       payload.EmpresaID,
				})

				w.WriteHeader(http.StatusNoContent)
				return
			}

			docActual, err := dbpkg.GetEmpresaDocumentoCompraByCodigo(dbEmp, payload.EmpresaID, tipoDocumento, payload.DocumentoCodigo)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "documento no encontrado", http.StatusNotFound)
					return
				}
				http.Error(w, "No se pudo consultar el documento de compras", http.StatusInternalServerError)
				return
			}

			payload.EstadoActual = comprasFirstNonBlank(payload.EstadoActual, docActual.EstadoDocumento)
			payload.ProveedorID = comprasFirstNonBlankInt64(payload.ProveedorID, docActual.ProveedorID)

			estadoDocumento := docActual.EstadoDocumento
			estadoAnterior := docActual.EstadoAnterior
			evento := "orden_compra_actualizada"
			accionResp := "actualizar"

			if action != "actualizar" {
				transition, err := resolveComprasTransition(action, payload.EstadoActual)
				if err != nil {
					errLower := strings.ToLower(err.Error())
					switch {
					case strings.Contains(errLower, "transicion invalida"):
						http.Error(w, err.Error(), http.StatusConflict)
						return
					case strings.Contains(errLower, "accion no soportada"):
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					default:
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
				}
				estadoDocumento = transition.EstadoNuevo
				estadoAnterior = transition.EstadoAnterior
				evento = transition.Evento
				accionResp = transition.Accion
			}

			docPersistido, err := dbpkg.UpsertEmpresaDocumentoCompra(dbEmp, dbpkg.EmpresaDocumentoCompra{
				EmpresaID:            payload.EmpresaID,
				ProveedorID:          payload.ProveedorID,
				TipoDocumento:        docActual.TipoDocumento,
				DocumentoCodigo:      docActual.DocumentoCodigo,
				EstadoDocumento:      estadoDocumento,
				EstadoAnterior:       estadoAnterior,
				EventoUltimo:         evento,
				PeriodoContable:      comprasFirstNonBlank(payload.PeriodoContable, docActual.PeriodoContable),
				MontoTotal:           comprasFirstPositive(payload.MontoTotal, docActual.MontoTotal),
				Moneda:               comprasFirstNonBlank(payload.Moneda, docActual.Moneda),
				FechaDocumento:       comprasFirstNonBlank(payload.FechaDocumento, docActual.FechaDocumento),
				EntidadRelacionadaID: payload.ProveedorID,
				UsuarioCreador:       adminEmailFromRequest(r),
				Estado:               docActual.Estado,
				Observaciones:        comprasFirstNonBlank(payload.Observaciones, docActual.Observaciones),
			})
			if err != nil {
				http.Error(w, "No se pudo actualizar el documento de compras", http.StatusInternalServerError)
				return
			}

			registrarEventoContableNoBloqueante(dbEmp, r, "compras", dbpkg.EmpresaEventoContable{
				EmpresaID:       payload.EmpresaID,
				Modulo:          "compras",
				Evento:          evento,
				Entidad:         "orden_compra",
				EntidadID:       docPersistido.ID,
				DocumentoTipo:   docPersistido.TipoDocumento,
				DocumentoCodigo: docPersistido.DocumentoCodigo,
				PeriodoContable: docPersistido.PeriodoContable,
				MontoTotal:      docPersistido.MontoTotal,
				Moneda:          docPersistido.Moneda,
				Origen:          "api_compras_documentos",
				Observaciones:   strings.TrimSpace(payload.Observaciones),
			}, map[string]interface{}{
				"accion":           accionResp,
				"estado_anterior":  docPersistido.EstadoAnterior,
				"estado_nuevo":     docPersistido.EstadoDocumento,
				"entidad_id":       docPersistido.ID,
				"documento_codigo": docPersistido.DocumentoCodigo,
				"proveedor_id":     docPersistido.ProveedorID,
				"empresa_id":       docPersistido.EmpresaID,
			})

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":        true,
				"accion":    accionResp,
				"evento":    evento,
				"resultado": docPersistido,
			})
			return

		case http.MethodDelete:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			tipoDocumento := comprasFirstNonBlank(r.URL.Query().Get("tipo_documento"), "orden_compra")
			documentoCodigo := strings.TrimSpace(r.URL.Query().Get("documento_codigo"))
			if documentoCodigo == "" {
				http.Error(w, "documento_codigo es obligatorio", http.StatusBadRequest)
				return
			}

			if err := dbpkg.SetEmpresaDocumentoCompraEstadoByCodigo(dbEmp, empresaID, tipoDocumento, documentoCodigo, "inactivo"); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "documento no encontrado", http.StatusNotFound)
					return
				}
				http.Error(w, "No se pudo eliminar documento de compras", http.StatusInternalServerError)
				return
			}

			registrarEventoContableNoBloqueante(dbEmp, r, "compras", dbpkg.EmpresaEventoContable{
				EmpresaID:       empresaID,
				Modulo:          "compras",
				Evento:          "orden_compra_eliminada",
				Entidad:         "orden_compra",
				DocumentoTipo:   tipoDocumento,
				DocumentoCodigo: strings.ToUpper(strings.TrimSpace(documentoCodigo)),
				Origen:          "api_compras_documentos",
				Observaciones:   "eliminacion logica de orden de compra",
			}, map[string]interface{}{
				"accion":           "eliminar",
				"documento_codigo": strings.ToUpper(strings.TrimSpace(documentoCodigo)),
				"empresa_id":       empresaID,
			})

			w.WriteHeader(http.StatusNoContent)
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

func parseBoolQuery(r *http.Request, key string) bool {
	raw := strings.ToLower(strings.TrimSpace(r.URL.Query().Get(key)))
	switch raw {
	case "1", "true", "si", "yes":
		return true
	default:
		return false
	}
}

func comprasFirstNonBlank(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func comprasFirstPositive(values ...float64) float64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func comprasFirstNonBlankInt64(values ...int64) int64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}
