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
	TarifasCreadas    int      `json:"tarifas_creadas"`
	ModulosCreados    int      `json:"modulos_creados"`
	VentaDirecta      bool     `json:"venta_directa"`
	Comisiones        bool     `json:"comisiones"`
	AdaptacionNucleo  any      `json:"adaptacion_nucleo,omitempty"`
	ProductosError    []string `json:"productos_error,omitempty"`
	UsuariosError     []string `json:"usuarios_error,omitempty"`
	TarifasError      []string `json:"tarifas_error,omitempty"`
	ModulosError      []string `json:"modulos_error,omitempty"`
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
	if template.Estaciones.Cantidad == 0 && len(template.Productos) == 0 && len(template.Usuarios) == 0 && !template.Asistente.Enabled && len(template.TareasGuia) == 0 && tipoEmpresaPreconfigTarifasEmpty(template.Tarifas) && tipoEmpresaPreconfigModulosEmpty(template.Modulos) {
		return result, nil
	}

	if err := dbpkg.EnsureEmpresaEstacionPrefsSchema(dbEmp); err != nil {
		return result, err
	}
	if pref, prefErr := dbpkg.GetEmpresaEstacionPref(dbEmp, empresaID, 0, "preconfiguracion_tipo_empresa_aplicada"); prefErr == nil && pref != nil && strings.TrimSpace(pref.Valor) != "" {
		var marker struct {
			TipoEmpresaID int64 `json:"tipo_empresa_id"`
		}
		if json.Unmarshal([]byte(pref.Valor), &marker) == nil && marker.TipoEmpresaID == tipoEmpresaID {
			_ = applyEmpresaPreconfigOperacion(dbEmp, empresaID, template.Operacion, usuario)
			_ = applyEmpresaPreconfigAdaptacionNucleo(dbEmp, empresaID, template.AdaptacionNucleo, usuario)
			result.Aplicada = true
			result.AdaptacionNucleo = template.AdaptacionNucleo
			result.Mensaje = "La preconfiguracion de este tipo de empresa ya estaba aplicada; se conserva sin duplicar datos guia."
			return result, nil
		}
	}
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		return result, err
	}
	if len(template.Usuarios) > 0 {
		if err := dbpkg.EnsureEmpresaUsuariosAuthSchema(dbEmp); err != nil {
			return result, err
		}
	}
	if err := applyEmpresaPreconfigOperacion(dbEmp, empresaID, template.Operacion, usuario); err != nil {
		return result, err
	}
	if err := applyEmpresaPreconfigAdaptacionNucleo(dbEmp, empresaID, template.AdaptacionNucleo, usuario); err != nil {
		return result, err
	}
	result.VentaDirecta = template.Operacion.VentaDirectaEnabled
	result.Comisiones = template.Operacion.ComisionesEnabled
	result.AdaptacionNucleo = template.AdaptacionNucleo

	productIDs := make([]int64, 0, len(template.Productos))
	productSKUs := make([]string, 0, len(template.Productos))
	userIDs := make([]int64, 0, len(template.Usuarios))
	userEmails := make([]string, 0, len(template.Usuarios))
	if template.Estaciones.Cantidad > 0 {
		rawConfig, estaciones := buildEmpresaEstacionesPreconfig(template.Estaciones, template.AdaptacionNucleo)
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

	tarifasCreadas, tarifasErr := applyEmpresaPreconfigTarifas(dbEmp, empresaID, template.Estaciones, template.Tarifas, usuario)
	result.TarifasCreadas = tarifasCreadas
	result.TarifasError = append(result.TarifasError, tarifasErr...)

	modulosCreados, modulosErr := applyEmpresaPreconfigModulos(dbEmp, empresaID, template.Estaciones, template.Modulos, usuario)
	result.ModulosCreados = modulosCreados
	result.ModulosError = append(result.ModulosError, modulosErr...)

	assistantRaw, _ := json.Marshal(map[string]any{
		"tipo_empresa_id":     tipoEmpresaID,
		"tipo_empresa_nombre": strings.TrimSpace(tipoEmpresaNombre),
		"preconfiguracion":    strings.TrimSpace(preconfig.Nombre),
		"operacion":           template.Operacion,
		"adaptacion_nucleo":   template.AdaptacionNucleo,
		"asistente_ia":        template.Asistente,
		"tareas_guia":         template.TareasGuia,
		"usuarios_guia":       template.Usuarios,
		"estaciones":          template.Estaciones,
		"tarifas":             template.Tarifas,
		"modulos":             template.Modulos,
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
		"operacion":           template.Operacion,
		"adaptacion_nucleo":   template.AdaptacionNucleo,
		"aplicada_en":         time.Now().Format(time.RFC3339),
		"estaciones_creadas":  result.EstacionesCreadas,
		"productos_creados":   result.ProductosCreados,
		"usuarios_creados":    result.UsuariosCreados,
		"tarifas_creadas":     result.TarifasCreadas,
		"modulos_creados":     result.ModulosCreados,
		"producto_ids":        productIDs,
		"producto_skus":       productSKUs,
		"usuario_ids":         userIDs,
		"usuario_emails":      userEmails,
		"asistente_ia":        template.Asistente,
		"tareas_guia":         template.TareasGuia,
		"tarifas":             template.Tarifas,
		"modulos":             template.Modulos,
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

func applyEmpresaTipoPreconfiguracionFromLicencia(dbEmp, dbSuper *sql.DB, empresaID, licenciaID int64, usuario string) (*empresaPreconfigApplyResult, error) {
	if dbEmp == nil {
		dbEmp = dbpkg.GetDB()
	}
	if empresaID <= 0 || dbEmp == nil || dbSuper == nil {
		return nil, nil
	}

	lic, _ := dbpkg.GetLicenciaByID(dbSuper, licenciaID)
	empresa, _ := dbpkg.GetEmpresaByScopeID(dbEmp, empresaID)
	tipoID := int64(0)
	tipoNombre := ""
	if empresa != nil {
		tipoID = empresa.TipoID
		tipoNombre = strings.TrimSpace(empresa.TipoNombre)
	}
	if lic != nil && lic.TipoID > 0 {
		if tipoID > 0 && tipoID != lic.TipoID && lic.EsAdicional != 1 {
			return nil, fmt.Errorf("la licencia pertenece a otro tipo de empresa")
		}
		tipoID = lic.TipoID
	}
	if tipoID <= 0 && tipoNombre == "" {
		return nil, nil
	}
	result, err := applyEmpresaTipoPreconfiguracion(dbEmp, dbSuper, empresaID, tipoID, tipoNombre, usuario)
	if err != nil {
		return result, err
	}
	return result, nil
}

func applyEmpresaPreconfigOperacion(dbEmp *sql.DB, empresaID int64, operacion dbpkg.TipoEmpresaPreconfigOperacion, usuario string) error {
	if empresaID <= 0 || dbEmp == nil {
		return nil
	}
	rawOperacion, _ := json.Marshal(map[string]any{
		"tipo_negocio":              strings.TrimSpace(operacion.TipoNegocio),
		"nombre_estacion_singular":  strings.TrimSpace(operacion.NombreEstacionSingular),
		"nombre_estacion_plural":    strings.TrimSpace(operacion.NombreEstacionPlural),
		"usa_estaciones":            operacion.UsaEstaciones,
		"venta_directa_enabled":     operacion.VentaDirectaEnabled,
		"venta_directa_nombre":      strings.TrimSpace(operacion.VentaDirectaNombre),
		"venta_directa_url":         "/administrar_empresa/carrito_de_compras.html?modo=venta_directa&perm_page=linkVentaDirecta",
		"carrito_rapido_url":        "/administrar_empresa/carrito_de_compras.html?modo=venta_directa&perm_page=linkVentaDirecta",
		"comisiones_enabled":        operacion.ComisionesEnabled,
		"comision_rol":              strings.TrimSpace(operacion.ComisionRol),
		"comision_filtro":           strings.TrimSpace(operacion.ComisionFiltro),
		"comision_porcentaje":       operacion.ComisionPorcentaje,
		"roles_operativos":          operacion.RolesOperativos,
		"preconfiguracion_aplicada": true,
		"fecha_actualizacion":       time.Now().Format(time.RFC3339),
	})
	if _, err := dbpkg.UpsertEmpresaEstacionPref(dbEmp, dbpkg.EmpresaEstacionPref{
		EmpresaID:      empresaID,
		EstacionID:     0,
		Clave:          "preconfiguracion_tipo_empresa_operacion",
		Valor:          string(rawOperacion),
		UsuarioCreador: usuario,
		Estado:         "activo",
		Observaciones:  empresaPreconfigMarker + " reglas operativas por tipo de empresa",
	}); err != nil {
		return err
	}

	if operacion.VentaDirectaEnabled {
		rawVentaDirecta, _ := json.Marshal(map[string]any{
			"enabled":        true,
			"nombre":         strings.TrimSpace(defaultString(operacion.VentaDirectaNombre, "Venta directa")),
			"url":            "/administrar_empresa/carrito_de_compras.html?modo=venta_directa&perm_page=linkVentaDirecta",
			"carrito_url":    "/administrar_empresa/carrito_de_compras.html?modo=venta_directa&perm_page=linkVentaDirecta",
			"modo":           "venta_directa",
			"crear_carrito":  true,
			"usa_estaciones": operacion.UsaEstaciones,
			"tipo_negocio":   strings.TrimSpace(operacion.TipoNegocio),
		})
		if _, err := dbpkg.UpsertEmpresaEstacionPref(dbEmp, dbpkg.EmpresaEstacionPref{
			EmpresaID:      empresaID,
			EstacionID:     0,
			Clave:          "venta_directa_config",
			Valor:          string(rawVentaDirecta),
			UsuarioCreador: usuario,
			Estado:         "activo",
			Observaciones:  empresaPreconfigMarker + " carrito rapido para venta directa",
		}); err != nil {
			return err
		}
		if err := ensureEmpresaPreconfigVentaDirectaCarrito(dbEmp, empresaID, usuario); err != nil {
			log.Printf("[empresa_preconfiguracion] venta directa carrito empresa_id=%d error: %v", empresaID, err)
		}
	}

	if err := dbpkg.EnsureEmpresaConfiguracionOperativaSchema(dbEmp); err != nil {
		return err
	}
	if _, err := dbpkg.UpsertEmpresaConfiguracionOperativa(dbEmp, dbpkg.EmpresaConfiguracionOperativa{
		EmpresaID:                       empresaID,
		MetodoPagoEfectivo:              true,
		MetodoPagoTarjetaCredito:        true,
		MetodoPagoTarjetaDebito:         true,
		MetodoPagoTransferenciaBancaria: true,
		MetodoPagoMixto:                 true,
		MetodoPagoCodigoDescuento:       true,
		HabilitarPropinas:               false,
		HabilitarComisiones:             operacion.ComisionesEnabled,
		UsuarioCreador:                  usuario,
		Estado:                          "activo",
		Observaciones:                   empresaPreconfigMarker + " configuracion operativa inicial por tipo de empresa",
	}); err != nil {
		return err
	}
	for _, role := range operacion.RolesOperativos {
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}
		if _, err := dbpkg.UpsertEmpresaConfiguracionOperativaRol(dbEmp, dbpkg.EmpresaConfiguracionOperativaRol{
			EmpresaID:                       empresaID,
			Rol:                             role,
			MetodoPagoEfectivo:              true,
			MetodoPagoTarjetaCredito:        true,
			MetodoPagoTarjetaDebito:         true,
			MetodoPagoTransferenciaBancaria: true,
			MetodoPagoMixto:                 true,
			MetodoPagoCodigoDescuento:       true,
			HabilitarPropinas:               false,
			HabilitarComisiones:             operacion.ComisionesEnabled && strings.EqualFold(role, operacion.ComisionRol),
			UsuarioCreador:                  usuario,
			Estado:                          "activo",
			Observaciones:                   empresaPreconfigMarker + " rol operativo inicial",
		}); err != nil {
			return err
		}
	}

	if operacion.ComisionesEnabled {
		if err := dbpkg.EnsureEmpresaComisionesServicioSchema(dbEmp); err != nil {
			return err
		}
		porcentaje := operacion.ComisionPorcentaje
		if porcentaje <= 0 {
			porcentaje = 10
		}
		filtro := strings.TrimSpace(operacion.ComisionFiltro)
		if filtro == "" {
			filtro = "servicio"
		}
		if _, err := dbpkg.UpsertEmpresaComisionesServicioConfiguracion(dbEmp, dbpkg.EmpresaComisionesServicioConfiguracion{
			EmpresaID:              empresaID,
			HabilitarComisiones:    true,
			PorcentajeComision:     porcentaje,
			FiltroServicio:         filtro,
			AplicarAutomaticamente: true,
			UsuarioCreador:         usuario,
			Estado:                 "activo",
			Observaciones:          empresaPreconfigMarker + " comisiones automaticas por tipo de empresa",
		}); err != nil {
			return err
		}
	}
	return nil
}

func applyEmpresaPreconfigAdaptacionNucleo(dbEmp *sql.DB, empresaID int64, adaptacion dbpkg.TipoEmpresaPreconfigAdaptacionNucleo, usuario string) error {
	if empresaID <= 0 || dbEmp == nil {
		return nil
	}
	raw, _ := json.Marshal(map[string]any{
		"fuente_unica":                          adaptacion.FuenteUnica,
		"usuarios_desde_nucleo":                 adaptacion.UsuariosDesdeNucleo,
		"productos_servicios_desde_nucleo":      adaptacion.ProductosServiciosDesdeNucleo,
		"estaciones_como_recursos_configurados": adaptacion.EstacionesComoRecursosConfigurados,
		"entidad_estacion_singular":             strings.TrimSpace(adaptacion.EntidadEstacionSingular),
		"entidad_estacion_plural":               strings.TrimSpace(adaptacion.EntidadEstacionPlural),
		"usuarios_operativos":                   adaptacion.UsuariosOperativos,
		"productos_servicios_guia":              adaptacion.ProductosServiciosGuia,
		"estaciones_guia":                       adaptacion.EstacionesGuia,
		"reglas":                                adaptacion.Reglas,
		"usuarios_url":                          "/administrar_empresa/administrar_usuarios.html",
		"productos_servicios_url":               "/administrar_empresa/administrar_productos_menu.html",
		"estaciones_url":                        "/administrar_empresa/estaciones.html",
		"configuracion_estaciones_url":          "/administrar_empresa/configuracion_de_estaciones.html",
		"preconfiguracion_aplicada":             true,
		"fecha_actualizacion":                   time.Now().Format(time.RFC3339),
	})
	_, err := dbpkg.UpsertEmpresaEstacionPref(dbEmp, dbpkg.EmpresaEstacionPref{
		EmpresaID:      empresaID,
		EstacionID:     0,
		Clave:          "preconfiguracion_tipo_empresa_adaptacion_nucleo",
		Valor:          string(raw),
		UsuarioCreador: usuario,
		Estado:         "activo",
		Observaciones:  empresaPreconfigMarker + " usuarios, productos/servicios y estaciones como nucleo configurable",
	})
	return err
}

func ensureEmpresaPreconfigVentaDirectaCarrito(dbEmp *sql.DB, empresaID int64, usuario string) error {
	if err := dbpkg.EnsureEmpresaCarritosSchema(dbEmp); err != nil {
		return err
	}
	code := fmt.Sprintf("VENTA-DIRECTA-%d", empresaID)
	existing, err := dbpkg.GetCarritoCompraByCodigo(dbEmp, empresaID, code)
	if err == nil && existing != nil {
		return nil
	}
	if err != nil && err != sql.ErrNoRows {
		log.Printf("[empresa_preconfiguracion] buscar carrito venta directa empresa_id=%d codigo=%s error: %v", empresaID, code, err)
	}
	_, err = dbpkg.CreateCarritoCompra(dbEmp, dbpkg.CarritoCompra{
		EmpresaID:         empresaID,
		Codigo:            code,
		Nombre:            "Venta directa",
		CanalVenta:        "mostrador",
		EstadoCarrito:     "abierto",
		ReferenciaExterna: "VENTA_DIRECTA",
		UsuarioCreador:    usuario,
		Estado:            "activo",
		Observaciones:     empresaPreconfigMarker + " carrito rapido generado para venta directa",
	})
	return err
}

func defaultString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	return fallback
}

func pluralizeStationName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "Estaciones"
	}
	lower := strings.ToLower(value)
	if strings.HasSuffix(lower, "s") {
		return value
	}
	if strings.HasSuffix(lower, "z") {
		return value[:len(value)-1] + "ces"
	}
	if strings.HasSuffix(lower, "ion") || strings.HasSuffix(lower, "ad") || strings.HasSuffix(lower, "ed") {
		return value + "es"
	}
	return value + "s"
}

func buildEmpresaEstacionesPreconfig(estaciones dbpkg.TipoEmpresaPreconfigEstaciones, adaptacion dbpkg.TipoEmpresaPreconfigAdaptacionNucleo) (string, int) {
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
	tipoRecurso := strings.TrimSpace(adaptacion.EntidadEstacionSingular)
	if tipoRecurso == "" {
		tipoRecurso = prefijo
	}
	tipoRecursoPlural := strings.TrimSpace(adaptacion.EntidadEstacionPlural)
	if tipoRecursoPlural == "" {
		tipoRecursoPlural = pluralizeStationName(tipoRecurso)
	}
	items := make([]map[string]any, 0, cantidad)
	for i := 1; i <= cantidad; i++ {
		items = append(items, map[string]any{
			"id":                            i,
			"nombre":                        fmt.Sprintf("%s %d", prefijo, i),
			"tipo_recurso":                  tipoRecurso,
			"representa_recurso_negocio":    true,
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
		"cantidad":                 cantidad,
		"estaciones":               items,
		"tipo_recurso":             tipoRecurso,
		"tipo_recurso_plural":      tipoRecursoPlural,
		"estacion_nombre_singular": tipoRecurso,
		"estacion_nombre_plural":   tipoRecursoPlural,
		"adaptacion_nucleo":        true,
		"card_size":                cardSize,
		"caja_enabled":             estaciones.CajaEnabled,
		"caja_placement":           "before",
		"youtube_enabled":          false,
		"notas_enabled":            false,
		"ia_pedidos_enabled":       false,
		"carrito_ui_global":        defaultEmpresaPreconfigCarritoUI(),
		"station_card_ui":          map[string]any{"mostrar_cliente_nombre": true, "mostrar_tarifa_resumen": true, "mostrar_inicio": true, "mostrar_fin": true, "mostrar_total": true},
	})
	return string(raw), cantidad
}

func defaultEmpresaPreconfigCarritoUI() map[string]any {
	return map[string]any{
		"modo_pantalla_tactil":                    false,
		"mostrar_boton_buscar_productos":          true,
		"mostrar_busqueda_catalogo":               true,
		"mostrar_codigo_manual_item":              true,
		"mostrar_observaciones_item":              true,
		"mostrar_selector_cliente":                true,
		"cliente_obligatorio_pago":                false,
		"cliente_general_nombre":                  "Cliente General",
		"mostrar_impuestos_item":                  true,
		"mostrar_lector_codigo_barras":            true,
		"mostrar_descuentos":                      false,
		"mostrar_propina":                         false,
		"mostrar_comision":                        false,
		"permitir_pago_mixto":                     true,
		"mostrar_resumen_totales_carrito":         true,
		"mostrar_desglose_cobro":                  false,
		"mostrar_resumen_productos":               true,
		"mostrar_boton_pagar":                     true,
		"mostrar_tarjetas_pago":                   true,
		"mostrar_tarjeta_lector_codigo":           true,
		"mostrar_tarjeta_items_carrito":           true,
		"mostrar_tarjeta_totales_detalles":        true,
		"mostrar_tarjeta_cobro_estados":           false,
		"mostrar_tarjeta_acciones_carrito":        true,
		"mostrar_control_electrico_carrito":       true,
		"mostrar_tarjeta_valores_pago":            true,
		"mostrar_tarjeta_comision":                false,
		"mostrar_tarjeta_vip_cliente":             true,
		"mostrar_boton_descuentos_carrito":        true,
		"mostrar_boton_cambiar_tarifa_carrito":    true,
		"mostrar_boton_control_electrico_carrito": true,
		"mostrar_boton_cancelar_carrito":          true,
		"mostrar_boton_taxi_carrito":              true,
		"mostrar_boton_clientes_carrito":          true,
		"mostrar_boton_abonos_carrito":            true,
		"mostrar_boton_vehiculo_carrito":          true,
		"mostrar_alerta_tiempo_carrito":           false,
		"alerta_tiempo_minutos":                   10,
		"alerta_tiempo_activa_default":            false,
		"facturacion_offline_habilitada":          false,
		"marcar_factura_offline_pendiente":        true,
		"pago_qr_habilitado":                      false,
		"pago_qr_proveedor":                       "breb",
		"pago_qr_llave":                           "",
		"pago_qr_comercio":                        "",
		"pago_qr_payload_oficial":                 "",
		"pago_qr_instrucciones":                   "",
		"pago_qr_cuentas":                         []any{},
		"mostrar_imagen":                          true,
		"mostrar_precio":                          true,
	}
}

// ApplyDefaultCarritoUIToExistingEmpresaPrefs normaliza las configuraciones antiguas
// para que las empresas existentes usen el mismo carrito simplificado por defecto.
func ApplyDefaultCarritoUIToExistingEmpresaPrefs(dbEmp *sql.DB) error {
	if dbEmp == nil {
		return nil
	}
	if err := dbpkg.EnsureEmpresaEstacionPrefsSchema(dbEmp); err != nil {
		return err
	}

	empresas, err := dbpkg.GetEmpresas(dbEmp)
	if err != nil {
		return err
	}
	labelPresets := make(map[int64]empresaStationLabelPreset, len(empresas))
	for _, empresa := range empresas {
		empresaID := empresa.EmpresaID
		if empresaID <= 0 {
			empresaID = empresa.ID
		}
		if empresaID > 0 {
			labelPresets[empresaID] = stationLabelsForTipoEmpresaName(empresa.TipoNombre)
		}
	}

	rows, err := dbpkg.ExecQueryCompat(dbEmp, `
		SELECT id, empresa_id, COALESCE(valor, '')
		FROM empresa_estacion_prefs
		WHERE estacion_id = 0
		  AND clave = 'estaciones_config'
		  AND LOWER(COALESCE(estado, 'activo')) = 'activo'
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	updated := 0
	created := 0
	skippedInvalid := 0
	seenEmpresas := make(map[int64]bool)
	for rows.Next() {
		var id, empresaID int64
		var raw string
		if err := rows.Scan(&id, &empresaID, &raw); err != nil {
			return err
		}
		seenEmpresas[empresaID] = true
		nextRaw, changed, err := applyDefaultCarritoUIPresetToConfig(raw, defaultEmpresaPreconfigCarritoUI(), labelPresets[empresaID])
		if err != nil {
			skippedInvalid++
			log.Printf("[empresa_preconfiguracion] carrito_ui empresa_id=%d pref_id=%d json invalido: %v", empresaID, id, err)
			continue
		}
		if !changed {
			continue
		}
		if _, err := dbpkg.ExecCompat(dbEmp, `
			UPDATE empresa_estacion_prefs
			SET valor = ?, fecha_actualizacion = CURRENT_TIMESTAMP, observaciones = ?
			WHERE id = ?
		`, nextRaw, "[migracion_carrito_default_2026-05-17] carrito simplificado aplicado a empresa existente", id); err != nil {
			return err
		}
		updated++
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, empresa := range empresas {
		empresaID := empresa.EmpresaID
		if empresaID <= 0 {
			empresaID = empresa.ID
		}
		if empresaID <= 0 || seenEmpresas[empresaID] {
			continue
		}
		estado := strings.ToLower(strings.TrimSpace(empresa.Estado))
		if estado == "inactivo" || estado == "eliminado" {
			continue
		}
		labels := labelPresets[empresaID]
		rawConfig, _ := buildEmpresaEstacionesPreconfig(dbpkg.TipoEmpresaPreconfigEstaciones{
			Cantidad:    1,
			Prefijo:     defaultString(labels.Singular, "Estacion"),
			CardSize:    "medium",
			CajaEnabled: false,
		}, dbpkg.TipoEmpresaPreconfigAdaptacionNucleo{
			EntidadEstacionSingular: labels.Singular,
			EntidadEstacionPlural:   labels.Plural,
		})
		if _, err := dbpkg.UpsertEmpresaEstacionPref(dbEmp, dbpkg.EmpresaEstacionPref{
			EmpresaID:      empresaID,
			EstacionID:     0,
			Clave:          "estaciones_config",
			Valor:          rawConfig,
			UsuarioCreador: "sistema",
			Estado:         "activo",
			Observaciones:  "[migracion_carrito_default_2026-05-17] carrito simplificado creado para empresa existente",
		}); err != nil {
			return err
		}
		created++
	}
	if updated > 0 || created > 0 || skippedInvalid > 0 {
		log.Printf("[empresa_preconfiguracion] carrito_ui defaults empresas existentes actualizadas=%d creadas=%d omitidas_json_invalido=%d", updated, created, skippedInvalid)
	}
	return nil
}

type empresaStationLabelPreset struct {
	Singular string
	Plural   string
}

func stationLabelsForTipoEmpresaName(tipoNombre string) empresaStationLabelPreset {
	template := dbpkg.DefaultTipoEmpresaPreconfigTemplate(tipoNombre)
	singular := defaultString(template.Operacion.NombreEstacionSingular, template.Estaciones.Prefijo)
	singular = defaultString(singular, "Estacion")
	plural := defaultString(template.Operacion.NombreEstacionPlural, pluralizeStationName(singular))
	return empresaStationLabelPreset{Singular: singular, Plural: plural}
}

func stringFromConfigMap(cfg map[string]any, key string) string {
	if cfg == nil {
		return ""
	}
	value, ok := cfg[key]
	if !ok || value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func shouldReplaceGenericStationLabel(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	return normalized == "" || normalized == "estacion" || normalized == "estaciones"
}

func applyStationLabelsToExistingConfig(cfg map[string]any, labels empresaStationLabelPreset) {
	singular := defaultString(labels.Singular, "Estacion")
	plural := defaultString(labels.Plural, pluralizeStationName(singular))
	if shouldReplaceGenericStationLabel(stringFromConfigMap(cfg, "estacion_nombre_singular")) {
		cfg["estacion_nombre_singular"] = singular
	}
	if shouldReplaceGenericStationLabel(stringFromConfigMap(cfg, "estacion_nombre_plural")) {
		cfg["estacion_nombre_plural"] = plural
	}
	if shouldReplaceGenericStationLabel(stringFromConfigMap(cfg, "tipo_recurso")) {
		cfg["tipo_recurso"] = singular
	}
	if shouldReplaceGenericStationLabel(stringFromConfigMap(cfg, "tipo_recurso_plural")) {
		cfg["tipo_recurso_plural"] = plural
	}
}

func applyDefaultCarritoUIPresetToConfig(raw string, preset map[string]any, labels empresaStationLabelPreset) (string, bool, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", false, nil
	}
	var cfg map[string]any
	if err := json.Unmarshal([]byte(trimmed), &cfg); err != nil {
		return "", false, err
	}
	before, err := json.Marshal(cfg)
	if err != nil {
		return "", false, err
	}

	applyStationLabelsToExistingConfig(cfg, labels)
	cfg["carrito_ui_global"] = copyAnyMap(preset)
	if items, ok := cfg["estaciones"].([]any); ok {
		for _, item := range items {
			station, ok := item.(map[string]any)
			if !ok {
				continue
			}
			carrito := copyAnyMap(asAnyMap(station["carrito"]))
			if len(carrito) == 0 {
				carrito = copyAnyMap(asAnyMap(station["carrito_ui"]))
			}
			if len(carrito) == 0 {
				carrito = make(map[string]any)
			}
			carrito["usar_configuracion_global"] = true
			carrito["configuracion"] = copyAnyMap(preset)
			station["carrito"] = carrito
		}
	}

	after, err := json.Marshal(cfg)
	if err != nil {
		return "", false, err
	}
	return string(after), string(before) != string(after), nil
}

func asAnyMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}

func copyAnyMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		if nested, ok := v.(map[string]any); ok {
			out[k] = copyAnyMap(nested)
			continue
		}
		out[k] = v
	}
	return out
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
	return fmt.Sprintf("%s.%d.%d@empresa.local", local, empresaID, idx+1)
}

func tipoEmpresaPreconfigTarifasEmpty(t dbpkg.TipoEmpresaPreconfigTarifas) bool {
	return len(t.PorMinutos) == 0 && len(t.PorDia) == 0 && len(t.Motel) == 0
}

func tipoEmpresaPreconfigModulosEmpty(m dbpkg.TipoEmpresaPreconfigModulos) bool {
	return m.TurnosAtencion == nil && m.Gimnasio == nil && m.Odontologia == nil && m.Vehiculos == nil && m.ControlElectrico == nil && len(m.HojaVida) == 0
}

func preconfigStationRef(estaciones dbpkg.TipoEmpresaPreconfigEstaciones, numero int) (int64, string, string) {
	if numero <= 0 {
		numero = 1
	}
	if estaciones.Cantidad > 0 && numero > estaciones.Cantidad {
		numero = estaciones.Cantidad
	}
	prefix := strings.TrimSpace(estaciones.Prefijo)
	if prefix == "" {
		prefix = "Estacion"
	}
	name := fmt.Sprintf("%s %d", prefix, numero)
	return int64(numero), fmt.Sprintf("%s-%03d", strings.ToUpper(strings.ReplaceAll(prefix, " ", "-")), numero), name
}

func applyEmpresaPreconfigTarifas(dbEmp *sql.DB, empresaID int64, estaciones dbpkg.TipoEmpresaPreconfigEstaciones, tarifas dbpkg.TipoEmpresaPreconfigTarifas, usuario string) (int, []string) {
	created := 0
	var errs []string
	if len(tarifas.PorMinutos) > 0 {
		if err := dbpkg.EnsureEmpresaTarifasPorMinutosSchema(dbEmp); err != nil {
			errs = append(errs, "tarifas por minutos: "+err.Error())
		} else {
			for _, item := range tarifas.PorMinutos {
				stationID, code, name := preconfigStationRef(estaciones, item.EstacionNumero)
				_, err := dbpkg.CreateEmpresaTarifaPorMinutos(dbEmp, dbpkg.EmpresaTarifaPorMinutos{
					EmpresaID:         empresaID,
					EstacionID:        stationID,
					EstacionCodigo:    code,
					EstacionNombre:    name,
					DiaSemanaDesde:    item.DiaSemanaDesde,
					DiaSemanaHasta:    item.DiaSemanaHasta,
					MinutosBase:       item.MinutosBase,
					ValorBase:         item.ValorBase,
					MinutosExtra:      item.MinutosExtra,
					ValorExtra:        item.ValorExtra,
					CobrarPorFraccion: item.CobrarPorFraccion,
					Moneda:            item.Moneda,
					Prioridad:         item.Prioridad,
					UsuarioCreador:    usuario,
					Estado:            "activo",
					Observaciones:     empresaPreconfigMarker + " " + strings.TrimSpace(item.Observaciones),
				})
				if err != nil {
					errs = append(errs, fmt.Sprintf("tarifa minutos %s: %v", name, err))
					continue
				}
				created++
			}
		}
	}
	if len(tarifas.PorDia) > 0 {
		if err := dbpkg.EnsureEmpresaTarifasPorDiaSchema(dbEmp); err != nil {
			errs = append(errs, "tarifas por dia: "+err.Error())
		} else {
			for _, item := range tarifas.PorDia {
				stationID, code, name := preconfigStationRef(estaciones, item.EstacionNumero)
				_, err := dbpkg.CreateEmpresaTarifaPorDia(dbEmp, dbpkg.EmpresaTarifaPorDia{
					EmpresaID:              empresaID,
					NombreTarifa:           item.NombreTarifa,
					EstacionID:             stationID,
					EstacionCodigo:         code,
					EstacionNombre:         name,
					ServicioNombre:         item.ServicioNombre,
					ValorDia:               item.ValorDia,
					PersonasDesde:          item.PersonasDesde,
					PersonasHasta:          item.PersonasHasta,
					HoraCheckIn:            item.HoraCheckIn,
					HoraCheckOut:           item.HoraCheckOut,
					Moneda:                 item.Moneda,
					Prioridad:              item.Prioridad,
					AplicarAutomaticamente: item.AplicarAutomaticamente,
					UsuarioCreador:         usuario,
					Estado:                 "activo",
					Observaciones:          empresaPreconfigMarker + " " + strings.TrimSpace(item.Observaciones),
				})
				if err != nil {
					errs = append(errs, fmt.Sprintf("tarifa dia %s: %v", name, err))
					continue
				}
				created++
			}
		}
	}
	if len(tarifas.Motel) > 0 {
		if err := dbpkg.EnsureEmpresaTarifasMotelSchema(dbEmp); err != nil {
			errs = append(errs, "tarifas motel: "+err.Error())
		} else {
			for _, item := range tarifas.Motel {
				stationID, code, name := preconfigStationRef(estaciones, item.EstacionNumero)
				_, err := dbpkg.CreateEmpresaTarifaMotel(dbEmp, dbpkg.EmpresaTarifaMotel{
					EmpresaID:           empresaID,
					EstacionID:          stationID,
					EstacionCodigo:      code,
					EstacionNombre:      name,
					NombrePlan:          item.NombrePlan,
					TipoPlan:            item.TipoPlan,
					CategoriaHabitacion: item.CategoriaHabitacion,
					DiaSemanaDesde:      item.DiaSemanaDesde,
					DiaSemanaHasta:      item.DiaSemanaHasta,
					HoraInicio:          item.HoraInicio,
					HoraFin:             item.HoraFin,
					MinutosIncluidos:    item.MinutosIncluidos,
					ValorBase:           item.ValorBase,
					MinutosExtra:        item.MinutosExtra,
					ValorExtra:          item.ValorExtra,
					CobrarPorFraccion:   item.CobrarPorFraccion,
					ToleranciaMinutos:   item.ToleranciaMinutos,
					Moneda:              item.Moneda,
					Prioridad:           item.Prioridad,
					AplicarAutomatico:   item.AplicarAutomatico,
					UsuarioCreador:      usuario,
					Estado:              "activo",
					Observaciones:       empresaPreconfigMarker + " " + strings.TrimSpace(item.Observaciones),
				})
				if err != nil {
					errs = append(errs, fmt.Sprintf("tarifa motel %s: %v", item.NombrePlan, err))
					continue
				}
				created++
			}
		}
	}
	return created, errs
}

func applyEmpresaPreconfigModulos(dbEmp *sql.DB, empresaID int64, estaciones dbpkg.TipoEmpresaPreconfigEstaciones, modulos dbpkg.TipoEmpresaPreconfigModulos, usuario string) (int, []string) {
	created := 0
	var errs []string
	if modulos.TurnosAtencion != nil {
		n, e := applyEmpresaPreconfigTurnos(dbEmp, empresaID, *modulos.TurnosAtencion, usuario)
		created += n
		errs = append(errs, e...)
	}
	if modulos.Gimnasio != nil {
		n, e := applyEmpresaPreconfigGimnasio(dbEmp, empresaID, *modulos.Gimnasio, usuario)
		created += n
		errs = append(errs, e...)
	}
	if modulos.Odontologia != nil {
		n, e := applyEmpresaPreconfigOdontologia(dbEmp, empresaID, *modulos.Odontologia, usuario)
		created += n
		errs = append(errs, e...)
	}
	if modulos.Vehiculos != nil {
		n, e := applyEmpresaPreconfigVehiculos(dbEmp, empresaID, *modulos.Vehiculos, usuario)
		created += n
		errs = append(errs, e...)
	}
	if modulos.ControlElectrico != nil {
		n, e := applyEmpresaPreconfigControlElectrico(dbEmp, empresaID, estaciones, *modulos.ControlElectrico, usuario)
		created += n
		errs = append(errs, e...)
	}
	if len(modulos.HojaVida) > 0 {
		n, e := applyEmpresaPreconfigHojaVida(dbEmp, empresaID, modulos.HojaVida, usuario)
		created += n
		errs = append(errs, e...)
	}
	return created, errs
}

func applyEmpresaPreconfigTurnos(dbEmp *sql.DB, empresaID int64, cfg dbpkg.TipoEmpresaPreconfigTurnosAtencion, usuario string) (int, []string) {
	created := 0
	var errs []string
	if err := dbpkg.UpsertEmpresaTurnoAtencionConfig(dbEmp, dbpkg.EmpresaTurnoAtencionConfig{
		EmpresaID:                 empresaID,
		NombreSistema:             cfg.NombreSistema,
		NombrePantalla:            cfg.NombrePantalla,
		PrefijoGeneral:            cfg.PrefijoGeneral,
		TiempoLlamadoSegundos:     cfg.TiempoLlamadoSegundos,
		PermitirEmisionPublica:    cfg.PermitirEmisionPublica,
		MostrarTicketsCompletados: cfg.MostrarTicketsCompletados,
		UsuarioCreador:            usuario,
	}); err != nil {
		errs = append(errs, "turnos config: "+err.Error())
	} else {
		created++
	}
	for _, svc := range cfg.Servicios {
		if _, err := dbpkg.CreateEmpresaTurnoAtencionServicio(dbEmp, dbpkg.EmpresaTurnoAtencionServicio{
			EmpresaID:      empresaID,
			Codigo:         svc.Codigo,
			Nombre:         svc.Nombre,
			Descripcion:    empresaPreconfigMarker + " " + svc.Descripcion,
			Prefijo:        svc.Prefijo,
			Prioridad:      svc.Prioridad,
			Color:          svc.Color,
			Estado:         "activo",
			UsuarioCreador: usuario,
		}); err != nil {
			errs = append(errs, fmt.Sprintf("turnos servicio %s: %v", svc.Codigo, err))
		} else {
			created++
		}
	}
	for _, puesto := range cfg.Puestos {
		if _, err := dbpkg.CreateEmpresaTurnoAtencionPuesto(dbEmp, dbpkg.EmpresaTurnoAtencionPuesto{
			EmpresaID:           empresaID,
			Codigo:              puesto.Codigo,
			Nombre:              puesto.Nombre,
			Area:                puesto.Area,
			Ubicacion:           puesto.Ubicacion,
			ServiciosPermitidos: puesto.ServiciosPermitidos,
			Estado:              "activo",
			UsuarioCreador:      usuario,
		}); err != nil {
			errs = append(errs, fmt.Sprintf("turnos puesto %s: %v", puesto.Codigo, err))
		} else {
			created++
		}
	}
	return created, errs
}

func applyEmpresaPreconfigGimnasio(dbEmp *sql.DB, empresaID int64, cfg dbpkg.TipoEmpresaPreconfigGimnasio, usuario string) (int, []string) {
	created := 0
	var errs []string
	planIDs := make([]int64, 0, len(cfg.Planes))
	entrenadorIDs := make([]int64, 0, len(cfg.Entrenadores))
	for _, plan := range cfg.Planes {
		id, err := dbpkg.CreateEmpresaGimnasioPlan(dbEmp, dbpkg.EmpresaGimnasioPlan{EmpresaID: empresaID, Nombre: plan.Nombre, Descripcion: empresaPreconfigMarker + " " + plan.Descripcion, Precio: plan.Precio, DuracionDias: plan.DuracionDias, ClasesIncluidas: plan.ClasesIncluidas, AccesoIlimitado: plan.AccesoIlimitado, SesionesPersonalizadas: plan.SesionesPersonalizadas, Estado: "activo", UsuarioCreador: usuario})
		if err != nil {
			errs = append(errs, fmt.Sprintf("gimnasio plan %s: %v", plan.Nombre, err))
			continue
		}
		planIDs = append(planIDs, id)
		created++
	}
	for _, entrenador := range cfg.Entrenadores {
		id, err := dbpkg.CreateEmpresaGimnasioEntrenador(dbEmp, dbpkg.EmpresaGimnasioEntrenador{EmpresaID: empresaID, NombreCompleto: entrenador.NombreCompleto, Especialidad: entrenador.Especialidad, Telefono: entrenador.Telefono, Email: entrenador.Email, Certificaciones: entrenador.Certificaciones, Estado: "activo", Disponibilidad: entrenador.Disponibilidad, Observaciones: empresaPreconfigMarker + " " + entrenador.Observaciones, UsuarioCreador: usuario})
		if err != nil {
			errs = append(errs, fmt.Sprintf("gimnasio entrenador %s: %v", entrenador.NombreCompleto, err))
			continue
		}
		entrenadorIDs = append(entrenadorIDs, id)
		created++
	}
	for _, clase := range cfg.Clases {
		var entrenadorID int64
		if clase.EntrenadorIndex > 0 && clase.EntrenadorIndex <= len(entrenadorIDs) {
			entrenadorID = entrenadorIDs[clase.EntrenadorIndex-1]
		}
		if _, err := dbpkg.CreateEmpresaGimnasioClase(dbEmp, dbpkg.EmpresaGimnasioClase{EmpresaID: empresaID, Nombre: clase.Nombre, Categoria: clase.Categoria, EntrenadorID: entrenadorID, Sede: clase.Sede, Canal: clase.Canal, Cupos: clase.Cupos, DuracionMinutos: clase.DuracionMinutos, Estado: "programada", Precio: clase.Precio, Descripcion: empresaPreconfigMarker + " " + clase.Descripcion, UsuarioCreador: usuario}); err != nil {
			errs = append(errs, fmt.Sprintf("gimnasio clase %s: %v", clase.Nombre, err))
		} else {
			created++
		}
	}
	for _, socio := range cfg.Socios {
		var planID int64
		if socio.PlanIndex > 0 && socio.PlanIndex <= len(planIDs) {
			planID = planIDs[socio.PlanIndex-1]
		}
		if _, err := dbpkg.CreateEmpresaGimnasioSocio(dbEmp, dbpkg.EmpresaGimnasioSocio{EmpresaID: empresaID, Codigo: socio.Codigo, NombreCompleto: socio.NombreCompleto, Documento: socio.Documento, Telefono: socio.Telefono, Email: socio.Email, Objetivo: socio.Objetivo, Estado: "activo", PlanID: planID, Observaciones: empresaPreconfigMarker + " " + socio.Observaciones, UsuarioCreador: usuario}); err != nil {
			errs = append(errs, fmt.Sprintf("gimnasio socio %s: %v", socio.NombreCompleto, err))
		} else {
			created++
		}
	}
	return created, errs
}

func applyEmpresaPreconfigOdontologia(dbEmp *sql.DB, empresaID int64, cfg dbpkg.TipoEmpresaPreconfigOdontologia, usuario string) (int, []string) {
	created := 0
	var errs []string
	pacienteIDs := make([]int64, 0, len(cfg.Pacientes))
	profesionalIDs := make([]int64, 0, len(cfg.Profesionales))
	for _, paciente := range cfg.Pacientes {
		id, err := dbpkg.CreateEmpresaOdontologiaPaciente(dbEmp, dbpkg.EmpresaOdontologiaPaciente{EmpresaID: empresaID, Codigo: paciente.Codigo, NombreCompleto: paciente.NombreCompleto, Documento: paciente.Documento, Telefono: paciente.Telefono, Email: paciente.Email, Aseguradora: paciente.Aseguradora, Alergias: paciente.Alergias, RiesgoMedico: paciente.RiesgoMedico, Saldo: paciente.Saldo, Estado: "activo", Observaciones: empresaPreconfigMarker + " " + paciente.Observaciones, UsuarioCreador: usuario})
		if err != nil {
			errs = append(errs, fmt.Sprintf("odontologia paciente %s: %v", paciente.NombreCompleto, err))
			continue
		}
		pacienteIDs = append(pacienteIDs, id)
		created++
	}
	for _, profesional := range cfg.Profesionales {
		id, err := dbpkg.CreateEmpresaOdontologiaProfesional(dbEmp, dbpkg.EmpresaOdontologiaProfesional{EmpresaID: empresaID, NombreCompleto: profesional.NombreCompleto, Especialidad: profesional.Especialidad, RegistroProfesional: profesional.RegistroProfesional, Telefono: profesional.Telefono, Email: profesional.Email, ColorAgenda: profesional.ColorAgenda, Estado: "activo", Observaciones: empresaPreconfigMarker + " " + profesional.Observaciones, UsuarioCreador: usuario})
		if err != nil {
			errs = append(errs, fmt.Sprintf("odontologia profesional %s: %v", profesional.NombreCompleto, err))
			continue
		}
		profesionalIDs = append(profesionalIDs, id)
		created++
	}
	for _, consultorio := range cfg.Consultorios {
		if _, err := dbpkg.CreateEmpresaOdontologiaConsultorio(dbEmp, dbpkg.EmpresaOdontologiaConsultorio{EmpresaID: empresaID, Nombre: consultorio.Nombre, Sede: consultorio.Sede, Sillon: consultorio.Sillon, Estado: "activo", Observaciones: empresaPreconfigMarker + " " + consultorio.Observaciones, UsuarioCreador: usuario}); err != nil {
			errs = append(errs, fmt.Sprintf("odontologia consultorio %s: %v", consultorio.Nombre, err))
		} else {
			created++
		}
	}
	for _, tratamiento := range cfg.Tratamientos {
		if len(pacienteIDs) == 0 {
			break
		}
		pacienteID := pacienteIDs[0]
		if tratamiento.PacienteIndex > 0 && tratamiento.PacienteIndex <= len(pacienteIDs) {
			pacienteID = pacienteIDs[tratamiento.PacienteIndex-1]
		}
		var profesionalID int64
		if tratamiento.ProfesionalIndex > 0 && tratamiento.ProfesionalIndex <= len(profesionalIDs) {
			profesionalID = profesionalIDs[tratamiento.ProfesionalIndex-1]
		}
		if _, err := dbpkg.CreateEmpresaOdontologiaTratamiento(dbEmp, dbpkg.EmpresaOdontologiaTratamiento{EmpresaID: empresaID, PacienteID: pacienteID, ProfesionalID: profesionalID, Nombre: tratamiento.Nombre, Categoria: tratamiento.Categoria, Piezas: tratamiento.Piezas, SesionesTotal: tratamiento.SesionesTotal, CostoEstimado: tratamiento.CostoEstimado, Estado: "planificado", Observaciones: empresaPreconfigMarker + " " + tratamiento.Observaciones, UsuarioCreador: usuario}); err != nil {
			errs = append(errs, fmt.Sprintf("odontologia tratamiento %s: %v", tratamiento.Nombre, err))
		} else {
			created++
		}
	}
	return created, errs
}

func applyEmpresaPreconfigVehiculos(dbEmp *sql.DB, empresaID int64, cfg dbpkg.TipoEmpresaPreconfigVehiculos, usuario string) (int, []string) {
	created := 0
	var errs []string
	if _, err := dbpkg.UpsertEmpresaVehiculosRegistroConfiguracion(dbEmp, dbpkg.EmpresaVehiculosRegistroConfiguracion{EmpresaID: empresaID, PaisCodigo: cfg.PaisCodigo, EvitarDuplicadoActivo: cfg.EvitarDuplicadoActivo, UsuarioCreador: usuario, Estado: "activo", Observaciones: empresaPreconfigMarker + " configuracion guia de vehiculos"}); err != nil {
		errs = append(errs, "vehiculos configuracion: "+err.Error())
	} else {
		created++
	}
	for _, item := range cfg.Registros {
		if _, err := dbpkg.CreateEmpresaVehiculoRegistro(dbEmp, dbpkg.EmpresaVehiculoRegistro{EmpresaID: empresaID, Patente: item.Patente, TipoVehiculo: item.TipoVehiculo, Marca: item.Marca, Modelo: item.Modelo, Color: item.Color, PropietarioNombre: item.PropietarioNombre, PropietarioDocumento: item.PropietarioDocumento, ConductorNombre: item.ConductorNombre, MotivoIngreso: item.MotivoIngreso, EstadoRegistro: "en_empresa", UsuarioCreador: usuario, Estado: "activo", Observaciones: empresaPreconfigMarker + " " + item.Observaciones}); err != nil {
			errs = append(errs, fmt.Sprintf("vehiculo %s: %v", item.Patente, err))
		} else {
			created++
		}
	}
	return created, errs
}

func applyEmpresaPreconfigControlElectrico(dbEmp *sql.DB, empresaID int64, estaciones dbpkg.TipoEmpresaPreconfigEstaciones, cfg dbpkg.TipoEmpresaPreconfigControlElectrico, usuario string) (int, []string) {
	created := 0
	var errs []string
	if err := dbpkg.EnsureEmpresaControlElectricoSchema(dbEmp); err != nil {
		return 0, []string{"control electrico: " + err.Error()}
	}
	if _, err := dbpkg.UpsertEmpresaControlElectricoConfig(dbEmp, &dbpkg.EmpresaControlElectricoConfig{
		EmpresaID:          empresaID,
		Habilitado:         cfg.Habilitado,
		RaspberryIP:        cfg.RaspberryIP,
		RaspberryPort:      cfg.RaspberryPort,
		APIPath:            cfg.APIPath,
		TimeoutMS:          cfg.TimeoutMS,
		AutoSyncEstaciones: cfg.AutoSyncEstaciones,
		FailSafeOnError:    cfg.FailSafeOnError,
		UsuarioCreador:     usuario,
		Estado:             "activo",
		Observaciones:      empresaPreconfigMarker + " configuracion guia de control electrico",
	}); err != nil {
		errs = append(errs, "control electrico config: "+err.Error())
	} else {
		created++
	}
	raspberryIDs := map[string]int64{}
	for _, item := range cfg.Raspberries {
		id, err := dbpkg.UpsertEmpresaControlElectricoRaspberry(dbEmp, &dbpkg.EmpresaControlElectricoRaspberry{
			EmpresaID:      empresaID,
			Codigo:         item.Codigo,
			Nombre:         item.Nombre,
			RaspberryIP:    item.RaspberryIP,
			RaspberryPort:  item.RaspberryPort,
			APIPath:        item.APIPath,
			TimeoutMS:      item.TimeoutMS,
			UsuarioCreador: usuario,
			Estado:         "activo",
			Observaciones:  empresaPreconfigMarker + " " + item.Observaciones,
		})
		if err != nil {
			errs = append(errs, fmt.Sprintf("control electrico raspberry %s: %v", item.Codigo, err))
			continue
		}
		raspberryIDs[strings.ToLower(strings.TrimSpace(item.Codigo))] = id
		created++
	}
	for _, item := range cfg.Reles {
		stationID, code, name := preconfigStationRef(estaciones, item.EstacionNumero)
		raspberryID := raspberryIDs[strings.ToLower(strings.TrimSpace(item.RaspberryCodigo))]
		id, err := dbpkg.UpsertEmpresaControlElectricoRele(dbEmp, &dbpkg.EmpresaControlElectricoRele{
			EmpresaID:              empresaID,
			RaspberryID:            raspberryID,
			EstacionID:             stationID,
			EstacionCodigo:         code,
			EstacionNombre:         name,
			SalidaCodigo:           item.SalidaCodigo,
			TipoCarga:              item.TipoCarga,
			GPIOPin:                item.GPIOPin,
			RelayName:              item.RelayName,
			ActiveHigh:             item.ActiveHigh,
			PulsoMS:                item.PulsoMS,
			Modo:                   item.Modo,
			ProgramacionHabilitada: item.ProgramacionHabilitada,
			HoraEncendido:          item.HoraEncendido,
			HoraApagado:            item.HoraApagado,
			ProgramacionDias:       item.ProgramacionDias,
			ProgramacionTimezone:   item.ProgramacionTimezone,
			ImagenURL:              item.ImagenURL,
			UsuarioCreador:         usuario,
			Estado:                 "activo",
			Observaciones:          empresaPreconfigMarker + " " + item.Observaciones,
		})
		if err != nil {
			errs = append(errs, fmt.Sprintf("control electrico rele %s estacion %d: %v", item.SalidaCodigo, item.EstacionNumero, err))
			continue
		}
		if id > 0 {
			created++
		}
	}
	return created, errs
}

func applyEmpresaPreconfigHojaVida(dbEmp *sql.DB, empresaID int64, hojas []dbpkg.TipoEmpresaPreconfigHojaVida, usuario string) (int, []string) {
	created := 0
	var errs []string
	if err := dbpkg.EnsureEmpresaHojaVidaOperativaSchema(dbEmp); err != nil {
		return 0, []string{"hoja de vida: " + err.Error()}
	}
	for _, item := range hojas {
		metaRaw := ""
		if len(item.Metadata) > 0 {
			raw, _ := json.Marshal(item.Metadata)
			metaRaw = string(raw)
		}
		entidadID, err := dbpkg.CreateEmpresaHojaVidaEntidad(dbEmp, dbpkg.EmpresaHojaVidaEntidad{EmpresaID: empresaID, TipoEntidad: item.TipoEntidad, Codigo: item.Codigo, Nombre: item.Nombre, ClienteNombre: item.ClienteNombre, Identificacion: item.Identificacion, Marca: item.Marca, Modelo: item.Modelo, Serie: item.Serie, Color: item.Color, EstadoOperativo: item.EstadoOperativo, MetadataJSON: metaRaw, UsuarioCreador: usuario, Estado: "activo", Observaciones: empresaPreconfigMarker + " " + item.Observaciones})
		if err != nil {
			errs = append(errs, fmt.Sprintf("hoja vida %s: %v", item.Nombre, err))
			continue
		}
		created++
		for _, evento := range item.Eventos {
			if _, err := dbpkg.CreateEmpresaHojaVidaEvento(dbEmp, dbpkg.EmpresaHojaVidaEvento{EmpresaID: empresaID, EntidadID: entidadID, TipoEvento: evento.TipoEvento, Titulo: evento.Titulo, Descripcion: evento.Descripcion, Costo: evento.Costo, Responsable: evento.Responsable, DocumentoReferencia: evento.DocumentoReferencia, Recurrente: evento.Recurrente, RecurrenciaDias: evento.RecurrenciaDias, UsuarioCreador: usuario, Estado: "activo", Observaciones: empresaPreconfigMarker + " " + evento.Observaciones}); err != nil {
				errs = append(errs, fmt.Sprintf("evento hoja vida %s: %v", evento.Titulo, err))
			} else {
				created++
			}
		}
		for _, alerta := range item.Alertas {
			if _, err := dbpkg.CreateEmpresaHojaVidaAlerta(dbEmp, dbpkg.EmpresaHojaVidaAlerta{EmpresaID: empresaID, EntidadID: entidadID, Titulo: alerta.Titulo, Descripcion: alerta.Descripcion, Prioridad: alerta.Prioridad, EstadoAlerta: "pendiente", Responsable: alerta.Responsable, UsuarioCreador: usuario, Estado: "activo", Observaciones: empresaPreconfigMarker + " " + alerta.Observaciones}); err != nil {
				errs = append(errs, fmt.Sprintf("alerta hoja vida %s: %v", alerta.Titulo, err))
			} else {
				created++
			}
		}
	}
	return created, errs
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
		"preconfiguracion_tipo_empresa_operacion",
		"venta_directa_config",
	})
	if err != nil {
		return nil, err
	}
	_ = clearEmpresaPreconfigOperacion(dbEmp, empresaID)
	return map[string]any{
		"productos_eliminados":    productosEliminados,
		"usuarios_eliminados":     usuariosEliminados,
		"preferencias_eliminadas": prefsEliminadas,
		"mensaje":                 "Preconfiguracion eliminada. La empresa quedo sin datos guia personalizados.",
	}, nil
}

func clearEmpresaPreconfigOperacion(dbEmp *sql.DB, empresaID int64) error {
	if dbEmp == nil || empresaID <= 0 {
		return nil
	}
	if err := dbpkg.EnsureEmpresaConfiguracionOperativaSchema(dbEmp); err == nil {
		_, _ = dbpkg.UpsertEmpresaConfiguracionOperativa(dbEmp, dbpkg.EmpresaConfiguracionOperativa{
			EmpresaID:                       empresaID,
			MetodoPagoEfectivo:              true,
			MetodoPagoTarjetaCredito:        true,
			MetodoPagoTarjetaDebito:         true,
			MetodoPagoTransferenciaBancaria: true,
			MetodoPagoMixto:                 true,
			MetodoPagoCodigoDescuento:       true,
			HabilitarPropinas:               true,
			HabilitarComisiones:             false,
			UsuarioCreador:                  "sistema.preconfiguracion",
			Estado:                          "activo",
			Observaciones:                   empresaPreconfigMarker + " limpieza de reglas operativas guia",
		})
	}
	if err := dbpkg.EnsureEmpresaComisionesServicioSchema(dbEmp); err == nil {
		_, _ = dbpkg.UpsertEmpresaComisionesServicioConfiguracion(dbEmp, dbpkg.EmpresaComisionesServicioConfiguracion{
			EmpresaID:              empresaID,
			HabilitarComisiones:    false,
			PorcentajeComision:     10,
			FiltroServicio:         "servicio",
			AplicarAutomaticamente: false,
			UsuarioCreador:         "sistema.preconfiguracion",
			Estado:                 "inactivo",
			Observaciones:          empresaPreconfigMarker + " limpieza de comisiones guia",
		})
	}
	return nil
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
				template, parseErr := dbpkg.ParseTipoEmpresaPreconfigTemplate(item.ConfigJSON)
				defaultTemplate, _ := dbpkg.ParseTipoEmpresaPreconfigTemplate(defaultItem.ConfigJSON)
				responseItem := map[string]any{
					"tipo_empresa":      tipo,
					"preconfig":         item,
					"template":          template,
					"default_preconfig": defaultItem,
					"default_template":  defaultTemplate,
					"es_default":        !exists,
				}
				if parseErr != nil {
					responseItem["template"] = defaultTemplate
					responseItem["config_error"] = parseErr.Error()
				}
				items = append(items, responseItem)
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "items": items})
			return
		case http.MethodPost, http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "seed_defaults" || action == "registrar_defaults" || action == "registrar_preconfiguraciones" {
				overwrite := parseEmpresaPreconfigBool(r.URL.Query().Get("overwrite"))
				result, err := dbpkg.SeedDefaultTipoEmpresaPreconfiguraciones(dbSuper, adminEmail, overwrite)
				if err != nil {
					http.Error(w, "no se pudieron registrar preconfiguraciones: "+err.Error(), http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]any{"ok": true, "result": result})
				return
			}

			var payload struct {
				TipoEmpresaID int64                                      `json:"tipo_empresa_id"`
				Enabled       bool                                       `json:"enabled"`
				Nombre        string                                     `json:"nombre"`
				Descripcion   string                                     `json:"descripcion"`
				Estaciones    dbpkg.TipoEmpresaPreconfigEstaciones       `json:"estaciones"`
				Operacion     dbpkg.TipoEmpresaPreconfigOperacion        `json:"operacion"`
				Adaptacion    dbpkg.TipoEmpresaPreconfigAdaptacionNucleo `json:"adaptacion_nucleo"`
				Productos     []dbpkg.TipoEmpresaPreconfigProducto       `json:"productos"`
				Usuarios      []dbpkg.TipoEmpresaPreconfigUsuario        `json:"usuarios"`
				Asistente     dbpkg.TipoEmpresaPreconfigAsistenteIA      `json:"asistente_ia"`
				TareasGuia    []dbpkg.TipoEmpresaPreconfigTareaGuia      `json:"tareas_guia"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.TipoEmpresaID <= 0 {
				http.Error(w, "tipo_empresa_id requerido", http.StatusBadRequest)
				return
			}
			payload.Nombre = strings.TrimSpace(payload.Nombre)
			if payload.Nombre == "" {
				payload.Nombre = "Preconfiguracion inicial"
			}
			if payload.Estaciones.Cantidad < 0 || payload.Estaciones.Cantidad > 200 {
				http.Error(w, "cantidad de estaciones debe estar entre 0 y 200", http.StatusBadRequest)
				return
			}
			configJSON, err := dbpkg.MarshalTipoEmpresaPreconfigTemplate(dbpkg.TipoEmpresaPreconfigTemplate{
				Estaciones:       payload.Estaciones,
				Operacion:        payload.Operacion,
				AdaptacionNucleo: payload.Adaptacion,
				Productos:        payload.Productos,
				Usuarios:         payload.Usuarios,
				Asistente:        payload.Asistente,
				TareasGuia:       payload.TareasGuia,
			})
			if err != nil {
				http.Error(w, "plantilla invalida: "+err.Error(), http.StatusBadRequest)
				return
			}
			id, err := dbpkg.UpsertTipoEmpresaPreconfiguracion(dbSuper, dbpkg.TipoEmpresaPreconfiguracion{
				TipoEmpresaID:  payload.TipoEmpresaID,
				Enabled:        payload.Enabled,
				Nombre:         payload.Nombre,
				Descripcion:    strings.TrimSpace(payload.Descripcion),
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

func parseEmpresaPreconfigBool(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "si", "sí", "yes", "on", "activo", "enabled":
		return true
	default:
		return false
	}
}
