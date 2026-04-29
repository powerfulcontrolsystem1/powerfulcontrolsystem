package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const empresaPreconfigMarker = "[preconfiguracion_tipo_empresa]"

type empresaPreconfigApplyResult struct {
	Aplicada          bool     `json:"aplicada"`
	TipoEmpresaID     int64    `json:"tipo_empresa_id,omitempty"`
	TipoEmpresaNombre string   `json:"tipo_empresa_nombre,omitempty"`
	EstacionesCreadas int      `json:"estaciones_creadas"`
	ProductosCreados  int      `json:"productos_creados"`
	ProductosError    []string `json:"productos_error,omitempty"`
	CarritosSync      any      `json:"carritos_sync,omitempty"`
	Mensaje           string   `json:"mensaje,omitempty"`
}

func applyEmpresaTipoPreconfiguracion(dbEmp, dbSuper *sql.DB, empresaID, tipoEmpresaID int64, tipoEmpresaNombre, usuario string) (*empresaPreconfigApplyResult, error) {
	result := &empresaPreconfigApplyResult{
		TipoEmpresaID:     tipoEmpresaID,
		TipoEmpresaNombre: strings.TrimSpace(tipoEmpresaNombre),
	}
	if empresaID <= 0 || dbEmp == nil || dbSuper == nil {
		return result, nil
	}

	preconfig, err := dbpkg.ResolveTipoEmpresaPreconfiguracion(dbSuper, tipoEmpresaID, tipoEmpresaNombre)
	if err != nil {
		return result, err
	}
	if preconfig == nil || !preconfig.Enabled {
		return result, nil
	}
	template, err := dbpkg.ParseTipoEmpresaPreconfigTemplate(preconfig.ConfigJSON)
	if err != nil {
		return result, fmt.Errorf("plantilla de preconfiguracion invalida: %w", err)
	}
	if template.Estaciones.Cantidad == 0 && len(template.Productos) == 0 {
		return result, nil
	}

	if err := dbpkg.EnsureEmpresaEstacionPrefsSchema(dbEmp); err != nil {
		return result, err
	}
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		return result, err
	}

	productIDs := make([]int64, 0, len(template.Productos))
	productSKUs := make([]string, 0, len(template.Productos))
	if template.Estaciones.Cantidad > 0 {
		rawConfig, estaciones := buildEmpresaEstacionesPreconfig(template.Estaciones)
		if _, err := dbpkg.UpsertEmpresaEstacionPref(dbEmp, dbpkg.EmpresaEstacionPref{
			EmpresaID:      empresaID,
			EstacionID:     0,
			Clave:          "estaciones_config",
			Valor:          rawConfig,
			UsuarioCreador: usuario,
			Estado:         "activo",
			Observaciones:  empresaPreconfigMarker + " estaciones iniciales",
		}); err != nil {
			return result, err
		}
		result.EstacionesCreadas = estaciones
		if syncResult, syncErr := dbpkg.SyncEmpresaEstacionCarritos(dbEmp, empresaID, rawConfig, usuario); syncErr != nil {
			log.Printf("[empresa_preconfiguracion] sync carritos empresa_id=%d error: %v", empresaID, syncErr)
		} else {
			result.CarritosSync = syncResult
		}
	}

	for _, p := range template.Productos {
		productoID, err := dbpkg.CreateProducto(dbEmp, dbpkg.Producto{
			EmpresaID:          empresaID,
			SKU:                strings.TrimSpace(p.SKU),
			Nombre:             strings.TrimSpace(p.Nombre),
			Descripcion:        strings.TrimSpace(p.Descripcion),
			Categoria:          strings.TrimSpace(p.Categoria),
			UnidadMedida:       strings.TrimSpace(p.UnidadMedida),
			Costo:              p.Costo,
			Precio:             p.Precio,
			ImpuestoPorcentaje: p.ImpuestoPorcentaje,
			StockMinimo:        p.StockMinimo,
			UsuarioCreador:     usuario,
			Estado:             "activo",
			Observaciones:      empresaPreconfigMarker + " producto guia de " + strings.TrimSpace(preconfig.Nombre),
		}, p.StockInicial, p.ReferenciaInventario)
		if err != nil {
			result.ProductosError = append(result.ProductosError, fmt.Sprintf("%s: %v", strings.TrimSpace(p.Nombre), err))
			log.Printf("[empresa_preconfiguracion] crear producto empresa_id=%d sku=%q error: %v", empresaID, p.SKU, err)
			continue
		}
		productIDs = append(productIDs, productoID)
		productSKUs = append(productSKUs, strings.TrimSpace(p.SKU))
		result.ProductosCreados++
	}

	metaRaw, _ := json.Marshal(map[string]any{
		"tipo_empresa_id":     tipoEmpresaID,
		"tipo_empresa_nombre": strings.TrimSpace(tipoEmpresaNombre),
		"preconfiguracion_id": preconfig.ID,
		"preconfiguracion":    strings.TrimSpace(preconfig.Nombre),
		"aplicada_en":         time.Now().Format(time.RFC3339),
		"estaciones_creadas":  result.EstacionesCreadas,
		"productos_creados":   result.ProductosCreados,
		"producto_ids":        productIDs,
		"producto_skus":       productSKUs,
	})
	_, _ = dbpkg.UpsertEmpresaEstacionPref(dbEmp, dbpkg.EmpresaEstacionPref{
		EmpresaID:      empresaID,
		EstacionID:     0,
		Clave:          "preconfiguracion_tipo_empresa_aplicada",
		Valor:          string(metaRaw),
		UsuarioCreador: usuario,
		Estado:         "activo",
		Observaciones:  empresaPreconfigMarker + " marcador de limpieza",
	})

	result.Aplicada = true
	result.Mensaje = "Empresa creada con preconfiguracion inicial. Puedes conservarla o eliminar la configuracion guia."
	return result, nil
}

func buildEmpresaEstacionesPreconfig(estaciones dbpkg.TipoEmpresaPreconfigEstaciones) (string, int) {
	cantidad := estaciones.Cantidad
	if cantidad <= 0 {
		cantidad = 1
	}
	if cantidad > 200 {
		cantidad = 200
	}
	prefijo := strings.TrimSpace(estaciones.Prefijo)
	if prefijo == "" {
		prefijo = "Estacion"
	}
	items := make([]map[string]any, 0, cantidad)
	for i := 1; i <= cantidad; i++ {
		items = append(items, map[string]any{
			"id":                            i,
			"nombre":                        fmt.Sprintf("%s %d", prefijo, i),
			"mostrar_fecha_hora_inicio":     true,
			"mostrar_fecha_hora_fin_tarifa": true,
			"mostrar_total":                 true,
			"carrito": map[string]any{
				"usar_configuracion_global": true,
			},
		})
	}
	cardSize := strings.ToLower(strings.TrimSpace(estaciones.CardSize))
	if cardSize == "" {
		cardSize = "medium"
	}
	raw, _ := json.Marshal(map[string]any{
		"cantidad":           cantidad,
		"estaciones":         items,
		"card_size":          cardSize,
		"caja_enabled":       estaciones.CajaEnabled,
		"caja_placement":     "before",
		"youtube_enabled":    false,
		"notas_enabled":      false,
		"ia_pedidos_enabled": false,
		"carrito_ui_global":  map[string]any{"mostrar_imagen": true, "mostrar_precio": true},
		"station_card_ui":    map[string]any{"mostrar_cliente_nombre": true, "mostrar_tarifa_resumen": true, "mostrar_inicio": true, "mostrar_fin": true, "mostrar_total": true},
	})
	return string(raw), cantidad
}

func clearEmpresaTipoPreconfiguracion(dbEmp *sql.DB, empresaID int64) (map[string]any, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id invalido")
	}
	if err := dbpkg.EnsureEmpresaEstacionPrefsSchema(dbEmp); err != nil {
		return nil, err
	}
	productosEliminados, err := dbpkg.DeleteProductosPreconfiguracion(dbEmp, empresaID)
	if err != nil {
		return nil, err
	}
	prefsEliminadas, err := dbpkg.DeleteEmpresaEstacionPrefsByKeys(dbEmp, empresaID, 0, []string{
		"estaciones_config",
		"preconfiguracion_tipo_empresa_aplicada",
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"productos_eliminados":    productosEliminados,
		"preferencias_eliminadas": prefsEliminadas,
		"mensaje":                 "Preconfiguracion eliminada. La empresa quedo sin datos guia personalizados.",
	}, nil
}

// SuperTipoEmpresaPreconfiguracionHandler administra plantillas iniciales por tipo de empresa.
func SuperTipoEmpresaPreconfiguracionHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminEmail, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper)
		if !ok {
			return
		}
		if dbSuper == nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "db_super no disponible"})
			return
		}

		switch r.Method {
		case http.MethodGet:
			tipos, err := dbpkg.GetTiposEmpresas(dbSuper)
			if err != nil {
				http.Error(w, "failed to query tipos_de_empresas: "+err.Error(), http.StatusInternalServerError)
				return
			}
			saved, err := dbpkg.ListTipoEmpresaPreconfiguraciones(dbSuper)
			if err != nil {
				http.Error(w, "failed to query preconfiguraciones: "+err.Error(), http.StatusInternalServerError)
				return
			}
			byTipo := map[int64]dbpkg.TipoEmpresaPreconfiguracion{}
			for _, item := range saved {
				byTipo[item.TipoEmpresaID] = item
			}
			items := make([]map[string]any, 0, len(tipos))
			for _, tipo := range tipos {
				item, exists := byTipo[tipo.ID]
				if !exists {
					item = dbpkg.DefaultTipoEmpresaPreconfiguracion(tipo.ID, tipo.Nombre)
				}
				template, _ := dbpkg.ParseTipoEmpresaPreconfigTemplate(item.ConfigJSON)
				items = append(items, map[string]any{
					"tipo_empresa": tipo,
					"preconfig":    item,
					"template":     template,
					"es_default":   !exists,
				})
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "items": items})
			return
		case http.MethodPost, http.MethodPut:
			var payload struct {
				TipoEmpresaID int64                                `json:"tipo_empresa_id"`
				Enabled       bool                                 `json:"enabled"`
				Nombre        string                               `json:"nombre"`
				Descripcion   string                               `json:"descripcion"`
				Estaciones    dbpkg.TipoEmpresaPreconfigEstaciones `json:"estaciones"`
				Productos     []dbpkg.TipoEmpresaPreconfigProducto `json:"productos"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.TipoEmpresaID <= 0 {
				http.Error(w, "tipo_empresa_id requerido", http.StatusBadRequest)
				return
			}
			configJSON, err := dbpkg.MarshalTipoEmpresaPreconfigTemplate(dbpkg.TipoEmpresaPreconfigTemplate{
				Estaciones: payload.Estaciones,
				Productos:  payload.Productos,
			})
			if err != nil {
				http.Error(w, "plantilla invalida: "+err.Error(), http.StatusBadRequest)
				return
			}
			id, err := dbpkg.UpsertTipoEmpresaPreconfiguracion(dbSuper, dbpkg.TipoEmpresaPreconfiguracion{
				TipoEmpresaID:  payload.TipoEmpresaID,
				Enabled:        payload.Enabled,
				Nombre:         payload.Nombre,
				Descripcion:    payload.Descripcion,
				ConfigJSON:     configJSON,
				UsuarioCreador: adminEmail,
				Estado:         "activo",
			})
			if err != nil {
				http.Error(w, "no se pudo guardar preconfiguracion: "+err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "id": id})
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}
