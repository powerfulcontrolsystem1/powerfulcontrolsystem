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
	UsuariosCreados   int      `json:"usuarios_creados"`
	ProductosError    []string `json:"productos_error,omitempty"`
	UsuariosError     []string `json:"usuarios_error,omitempty"`
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
	if template.Estaciones.Cantidad == 0 && len(template.Productos) == 0 && len(template.Usuarios) == 0 && !template.Asistente.Enabled && len(template.TareasGuia) == 0 {
		return result, nil
	}

	if err := dbpkg.EnsureEmpresaEstacionPrefsSchema(dbEmp); err != nil {
		return result, err
	}
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		return result, err
	}
	if len(template.Usuarios) > 0 {
		if err := dbpkg.EnsureEmpresaUsuariosAuthSchema(dbEmp); err != nil {
			return result, err
		}
	}

	productIDs := make([]int64, 0, len(template.Productos))
	productSKUs := make([]string, 0, len(template.Productos))
	userIDs := make([]int64, 0, len(template.Usuarios))
	userEmails := make([]string, 0, len(template.Usuarios))
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

	for idx, u := range template.Usuarios {
		email := buildPreconfigUsuarioEmail(u, empresaID, idx)
		nombre := strings.TrimSpace(u.Nombre)
		rol := strings.TrimSpace(u.Rol)
		if nombre == "" {
			continue
		}
		if rol == "" {
			rol = "operacion"
		}
		observaciones := empresaPreconfigMarker + " usuario guia de " + strings.TrimSpace(preconfig.Nombre)
		if strings.TrimSpace(u.Observaciones) != "" {
			observaciones += " | " + strings.TrimSpace(u.Observaciones)
		}
		id, err := dbpkg.CreateEmpresaUsuario(
			dbEmp,
			empresaID,
			email,
			nombre,
			"",
			0,
			rol,
			observaciones,
			usuario,
			fmt.Sprintf("preconfig-%d-%d-%d", empresaID, idx+1, time.Now().UnixNano()),
			"",
		)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "unique") {
				result.UsuariosError = append(result.UsuariosError, fmt.Sprintf("%s: ya existe el correo guia %s", nombre, email))
				continue
			}
			result.UsuariosError = append(result.UsuariosError, fmt.Sprintf("%s: %v", nombre, err))
			log.Printf("[empresa_preconfiguracion] crear usuario empresa_id=%d email=%q error: %v", empresaID, email, err)
			continue
		}
		userIDs = append(userIDs, id)
		userEmails = append(userEmails, email)
		result.UsuariosCreados++
	}

	assistantRaw, _ := json.Marshal(map[string]any{
		"tipo_empresa_id":     tipoEmpresaID,
		"tipo_empresa_nombre": strings.TrimSpace(tipoEmpresaNombre),
		"preconfiguracion":    strings.TrimSpace(preconfig.Nombre),
		"asistente_ia":        template.Asistente,
		"tareas_guia":         template.TareasGuia,
		"usuarios_guia":       template.Usuarios,
		"estaciones":          template.Estaciones,
		"producto_skus":       productSKUs,
		"actualizado_en":      time.Now().Format(time.RFC3339),
	})
	_, _ = dbpkg.UpsertEmpresaEstacionPref(dbEmp, dbpkg.EmpresaEstacionPref{
		EmpresaID:      empresaID,
		EstacionID:     0,
		Clave:          "preconfiguracion_tipo_empresa_asistente_ia",
		Valor:          string(assistantRaw),
		UsuarioCreador: usuario,
		Estado:         "activo",
		Observaciones:  empresaPreconfigMarker + " contexto guia para IA",
	})

	metaRaw, _ := json.Marshal(map[string]any{
		"tipo_empresa_id":     tipoEmpresaID,
		"tipo_empresa_nombre": strings.TrimSpace(tipoEmpresaNombre),
		"preconfiguracion_id": preconfig.ID,
		"preconfiguracion":    strings.TrimSpace(preconfig.Nombre),
		"aplicada_en":         time.Now().Format(time.RFC3339),
		"estaciones_creadas":  result.EstacionesCreadas,
		"productos_creados":   result.ProductosCreados,
		"usuarios_creados":    result.UsuariosCreados,
		"producto_ids":        productIDs,
		"producto_skus":       productSKUs,
		"usuario_ids":         userIDs,
		"usuario_emails":      userEmails,
		"asistente_ia":        template.Asistente,
		"tareas_guia":         template.TareasGuia,
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

func buildPreconfigUsuarioEmail(u dbpkg.TipoEmpresaPreconfigUsuario, empresaID int64, idx int) string {
	email := strings.ToLower(strings.TrimSpace(u.Email))
	if email != "" {
		return email
	}
	base := strings.ToLower(strings.TrimSpace(u.Rol))
	if base == "" {
		base = strings.ToLower(strings.TrimSpace(u.Nombre))
	}
	if base == "" {
		base = fmt.Sprintf("usuario%d", idx+1)
	}
	var b strings.Builder
	for _, r := range base {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '.' || r == '_' || r == '-':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune('.')
		}
	}
	local := strings.Trim(b.String(), ".-_")
	if local == "" {
		local = fmt.Sprintf("usuario%d", idx+1)
	}
	return fmt.Sprintf("%s.%d.%d@preconfig.local", local, empresaID, idx+1)
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
	usuariosEliminados, err := dbpkg.DeleteEmpresaUsuariosPreconfiguracion(dbEmp, empresaID, empresaPreconfigMarker)
	if err != nil {
		return nil, err
	}
	prefsEliminadas, err := dbpkg.DeleteEmpresaEstacionPrefsByKeys(dbEmp, empresaID, 0, []string{
		"estaciones_config",
		"preconfiguracion_tipo_empresa_aplicada",
		"preconfiguracion_tipo_empresa_asistente_ia",
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"productos_eliminados":    productosEliminados,
		"usuarios_eliminados":     usuariosEliminados,
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
				defaultItem := dbpkg.DefaultTipoEmpresaPreconfiguracion(tipo.ID, tipo.Nombre)
				if !exists {
					item = defaultItem
				}
				template, _ := dbpkg.ParseTipoEmpresaPreconfigTemplate(item.ConfigJSON)
				defaultTemplate, _ := dbpkg.ParseTipoEmpresaPreconfigTemplate(defaultItem.ConfigJSON)
				items = append(items, map[string]any{
					"tipo_empresa":      tipo,
					"preconfig":         item,
					"template":          template,
					"default_preconfig": defaultItem,
					"default_template":  defaultTemplate,
					"es_default":        !exists,
				})
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "items": items})
			return
		case http.MethodPost, http.MethodPut:
			var payload struct {
				TipoEmpresaID int64                                 `json:"tipo_empresa_id"`
				Enabled       bool                                  `json:"enabled"`
				Nombre        string                                `json:"nombre"`
				Descripcion   string                                `json:"descripcion"`
				Estaciones    dbpkg.TipoEmpresaPreconfigEstaciones  `json:"estaciones"`
				Productos     []dbpkg.TipoEmpresaPreconfigProducto  `json:"productos"`
				Usuarios      []dbpkg.TipoEmpresaPreconfigUsuario   `json:"usuarios"`
				Asistente     dbpkg.TipoEmpresaPreconfigAsistenteIA `json:"asistente_ia"`
				TareasGuia    []dbpkg.TipoEmpresaPreconfigTareaGuia `json:"tareas_guia"`
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
				Usuarios:   payload.Usuarios,
				Asistente:  payload.Asistente,
				TareasGuia: payload.TareasGuia,
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
