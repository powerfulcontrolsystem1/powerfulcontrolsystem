package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type empresaCalculadoraOperacionPayload struct {
	EmpresaID       int64                  `json:"empresa_id"`
	Expresion       string                 `json:"expresion"`
	Resultado       string                 `json:"resultado"`
	Etiquetas       []string               `json:"etiquetas"`
	EtiquetasTexto  string                 `json:"etiquetas_texto"`
	ClienteID       int64                  `json:"cliente_id"`
	ClienteNombre   string                 `json:"cliente_nombre"`
	DocumentoTipo   string                 `json:"documento_tipo"`
	DocumentoCodigo string                 `json:"documento_codigo"`
	CarritoID       int64                  `json:"carrito_id"`
	CotizacionID    int64                  `json:"cotizacion_id"`
	FechaOperacion  string                 `json:"fecha_operacion"`
	Metadata        map[string]interface{} `json:"metadata"`
	MetadataJSON    string                 `json:"metadata_json"`
	Observaciones   string                 `json:"observaciones"`
}

type empresaCalculadoraConfigPayload struct {
	EmpresaID            int64  `json:"empresa_id"`
	IntegrarCarritos     *bool  `json:"integrar_carritos"`
	IntegrarCotizaciones *bool  `json:"integrar_cotizaciones"`
	Estado               string `json:"estado"`
	Observaciones        string `json:"observaciones"`
}

type empresaCalculadoraClearPayload struct {
	EmpresaID int64  `json:"empresa_id"`
	Desde     string `json:"desde"`
	Hasta     string `json:"hasta"`
	Usuario   string `json:"usuario"`
	ClienteID int64  `json:"cliente_id"`
	Etiqueta  string `json:"etiqueta"`
}

func EmpresaCalculadoraHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		action := calcEmpresaNormalizeAction(r.URL.Query().Get("action"))
		if action == "" {
			action = calcEmpresaDefaultActionByMethod(r.Method)
		}

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "config", "configuracion":
				calcEmpresaHandleGetConfig(w, r, dbEmp)
				return
			case "referencias", "refs":
				calcEmpresaHandleReferences(w, r, dbEmp)
				return
			case "export", "exportar":
				calcEmpresaHandleExport(w, r, dbEmp)
				return
			default:
				calcEmpresaHandleList(w, r, dbEmp)
				return
			}
		case http.MethodPost:
			switch action {
			case "config", "configuracion":
				calcEmpresaHandleUpsertConfig(w, r, dbEmp)
				return
			case "limpiar", "clear", "clear_history":
				calcEmpresaHandleClear(w, r, dbEmp)
				return
			default:
				calcEmpresaHandleCreateOperacion(w, r, dbEmp)
				return
			}
		case http.MethodPut, http.MethodPatch:
			switch action {
			case "config", "configuracion":
				calcEmpresaHandleUpsertConfig(w, r, dbEmp)
				return
			case "activar", "desactivar":
				calcEmpresaHandleToggleOperacion(w, r, dbEmp, action)
				return
			case "limpiar", "clear", "clear_history":
				calcEmpresaHandleClear(w, r, dbEmp)
				return
			default:
				http.Error(w, "action invalida", http.StatusBadRequest)
				return
			}
		case http.MethodDelete:
			if action == "" || action == "eliminar" || action == "desactivar" {
				calcEmpresaHandleToggleOperacion(w, r, dbEmp, "desactivar")
				return
			}
			if action == "limpiar" || action == "clear" || action == "clear_history" {
				calcEmpresaHandleClear(w, r, dbEmp)
				return
			}
			http.Error(w, "action invalida", http.StatusBadRequest)
			return
		default:
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
	}
}

func calcEmpresaHandleGetConfig(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cfg, err := dbpkg.GetEmpresaCalculadoraConfiguracion(dbEmp, empresaID)
	if err != nil {
		http.Error(w, "No se pudo consultar configuracion de calculadora", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":            true,
		"empresa_id":    empresaID,
		"configuracion": cfg,
	})
}

func calcEmpresaHandleUpsertConfig(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var payload empresaCalculadoraConfigPayload
	if err := calcEmpresaDecodeJSONBody(r, &payload); err != nil {
		http.Error(w, "body JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID > 0 && payload.EmpresaID != empresaID {
		http.Error(w, "empresa_id no coincide con el contexto", http.StatusBadRequest)
		return
	}

	current, err := dbpkg.GetEmpresaCalculadoraConfiguracion(dbEmp, empresaID)
	if err != nil {
		http.Error(w, "No se pudo consultar configuracion actual", http.StatusInternalServerError)
		return
	}

	if payload.IntegrarCarritos != nil {
		current.IntegrarCarritos = *payload.IntegrarCarritos
	}
	if payload.IntegrarCotizaciones != nil {
		current.IntegrarCotizaciones = *payload.IntegrarCotizaciones
	}
	if strings.TrimSpace(payload.Estado) != "" {
		current.Estado = strings.TrimSpace(payload.Estado)
	}
	if strings.TrimSpace(payload.Observaciones) != "" {
		current.Observaciones = strings.TrimSpace(payload.Observaciones)
	}
	current.EmpresaID = empresaID
	current.UsuarioCreador = calcEmpresaUsuarioFromRequest(r)

	if _, err := dbpkg.UpsertEmpresaCalculadoraConfiguracion(dbEmp, *current); err != nil {
		http.Error(w, "No se pudo actualizar configuracion de calculadora", http.StatusInternalServerError)
		return
	}

	updated, err := dbpkg.GetEmpresaCalculadoraConfiguracion(dbEmp, empresaID)
	if err != nil {
		http.Error(w, "No se pudo consultar configuracion actualizada", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":            true,
		"empresa_id":    empresaID,
		"configuracion": updated,
	})
}

func calcEmpresaHandleCreateOperacion(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var payload empresaCalculadoraOperacionPayload
	if err := calcEmpresaDecodeJSONBody(r, &payload); err != nil {
		http.Error(w, "body JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID > 0 && payload.EmpresaID != empresaID {
		http.Error(w, "empresa_id no coincide con el contexto", http.StatusBadRequest)
		return
	}

	cfg, err := dbpkg.GetEmpresaCalculadoraConfiguracion(dbEmp, empresaID)
	if err != nil {
		http.Error(w, "No se pudo consultar configuracion de calculadora", http.StatusInternalServerError)
		return
	}

	metadataJSON, err := calcEmpresaBuildMetadataJSON(payload.MetadataJSON, payload.Metadata)
	if err != nil {
		http.Error(w, "metadata_json invalido", http.StatusBadRequest)
		return
	}

	etiquetas := payload.Etiquetas
	if len(etiquetas) == 0 && strings.TrimSpace(payload.EtiquetasTexto) != "" {
		etiquetas = strings.Split(payload.EtiquetasTexto, ",")
	}

	if payload.ClienteID > 0 {
		cliente, cliErr := dbpkg.GetClienteByID(dbEmp, empresaID, payload.ClienteID)
		if cliErr != nil {
			if errors.Is(cliErr, sql.ErrNoRows) {
				http.Error(w, "cliente_id no encontrado", http.StatusNotFound)
				return
			}
			http.Error(w, "No se pudo validar cliente_id", http.StatusInternalServerError)
			return
		}
		if strings.TrimSpace(payload.ClienteNombre) == "" {
			payload.ClienteNombre = strings.TrimSpace(cliente.NombreRazonSocial)
		}
	}

	if payload.CarritoID > 0 {
		if cfg == nil || !cfg.IntegrarCarritos {
			http.Error(w, "integracion con carritos deshabilitada en configuracion", http.StatusBadRequest)
			return
		}
		carrito, cartErr := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, payload.CarritoID)
		if cartErr != nil {
			if errors.Is(cartErr, sql.ErrNoRows) {
				http.Error(w, "carrito_id no encontrado", http.StatusNotFound)
				return
			}
			http.Error(w, "No se pudo validar carrito_id", http.StatusInternalServerError)
			return
		}
		if payload.ClienteID <= 0 && carrito.ClienteID > 0 {
			payload.ClienteID = carrito.ClienteID
		}
		if strings.TrimSpace(payload.ClienteNombre) == "" {
			payload.ClienteNombre = strings.TrimSpace(carrito.ClienteNombre)
		}
		if strings.TrimSpace(payload.DocumentoTipo) == "" {
			payload.DocumentoTipo = "carrito"
		}
		if strings.TrimSpace(payload.DocumentoCodigo) == "" {
			payload.DocumentoCodigo = strings.TrimSpace(carrito.Codigo)
		}
	}

	if payload.CotizacionID > 0 {
		if cfg == nil || !cfg.IntegrarCotizaciones {
			http.Error(w, "integracion con cotizaciones deshabilitada en configuracion", http.StatusBadRequest)
			return
		}
		row, rowErr := dbpkg.GetEmpresaGenericRowByID(dbEmp, "empresa_cotizaciones_venta", empresaID, payload.CotizacionID)
		if rowErr != nil {
			if errors.Is(rowErr, sql.ErrNoRows) {
				http.Error(w, "cotizacion_id no encontrado", http.StatusNotFound)
				return
			}
			http.Error(w, "No se pudo validar cotizacion_id", http.StatusInternalServerError)
			return
		}
		if payload.ClienteID <= 0 {
			payload.ClienteID = calcEmpresaMapInt64(row, "cliente_id")
		}
		if strings.TrimSpace(payload.ClienteNombre) == "" {
			payload.ClienteNombre = calcEmpresaMapString(row, "cliente_nombre")
		}
		if strings.TrimSpace(payload.DocumentoTipo) == "" {
			payload.DocumentoTipo = "cotizacion"
		}
		if strings.TrimSpace(payload.DocumentoCodigo) == "" {
			payload.DocumentoCodigo = calcEmpresaMapString(row, "codigo")
		}
	}

	op := dbpkg.EmpresaCalculadoraOperacion{
		EmpresaID:       empresaID,
		Expresion:       strings.TrimSpace(payload.Expresion),
		Resultado:       strings.TrimSpace(payload.Resultado),
		Etiquetas:       etiquetas,
		ClienteID:       payload.ClienteID,
		ClienteNombre:   strings.TrimSpace(payload.ClienteNombre),
		DocumentoTipo:   strings.TrimSpace(payload.DocumentoTipo),
		DocumentoCodigo: strings.TrimSpace(payload.DocumentoCodigo),
		CarritoID:       payload.CarritoID,
		CotizacionID:    payload.CotizacionID,
		FechaOperacion:  strings.TrimSpace(payload.FechaOperacion),
		MetadataJSON:    metadataJSON,
		UsuarioCreador:  calcEmpresaUsuarioFromRequest(r),
		Estado:          "activo",
		Observaciones:   strings.TrimSpace(payload.Observaciones),
	}

	id, err := dbpkg.CreateEmpresaCalculadoraOperacion(dbEmp, op)
	if err != nil {
		http.Error(w, "No se pudo registrar operacion en calculadora", http.StatusInternalServerError)
		return
	}

	created, err := dbpkg.GetEmpresaCalculadoraOperacionByID(dbEmp, empresaID, id)
	if err != nil {
		http.Error(w, "operacion creada pero no se pudo consultar", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"operacion":  created,
	})
}

func calcEmpresaHandleToggleOperacion(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, action string) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	operacionID, err := parseInt64QueryOptional(r, "id")
	if err != nil || operacionID <= 0 {
		http.Error(w, "id es obligatorio", http.StatusBadRequest)
		return
	}
	state := "activo"
	if action == "desactivar" {
		state = "inactivo"
	}
	if err := dbpkg.SetEmpresaCalculadoraOperacionEstadoByID(dbEmp, empresaID, operacionID, state); err != nil {
		http.Error(w, "No se pudo actualizar estado de la operacion", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":             true,
		"empresa_id":     empresaID,
		"id":             operacionID,
		"estado":         state,
		"actualizado":    true,
		"actualizado_en": time.Now().Format("2006-01-02 15:04:05"),
	})
}

func calcEmpresaHandleList(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filter, err := calcEmpresaBuildFilterFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rows, total, err := dbpkg.ListEmpresaCalculadoraOperaciones(dbEmp, empresaID, filter)
	if err != nil {
		http.Error(w, "No se pudo consultar historial de calculadora", http.StatusInternalServerError)
		return
	}

	limit, offset := calcEmpresaNormalizeLimitOffset(filter.Limit, filter.Offset)
	w.Header().Set("X-Total-Count", strconv.FormatInt(total, 10))
	w.Header().Set("X-Page-Limit", strconv.Itoa(limit))
	w.Header().Set("X-Page-Offset", strconv.Itoa(offset))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":           true,
		"empresa_id":   empresaID,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
		"rows":         rows,
		"filtros":      calcEmpresaFilterMap(filter),
		"generated_at": time.Now().Format("2006-01-02 15:04:05"),
	})
}

func calcEmpresaHandleClear(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filter, err := calcEmpresaBuildFilterFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var payload empresaCalculadoraClearPayload
	if (r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch || r.Method == http.MethodDelete) && r.Body != nil {
		if err := calcEmpresaDecodeJSONBody(r, &payload); err != nil {
			http.Error(w, "body JSON invalido", http.StatusBadRequest)
			return
		}
	}
	if payload.EmpresaID > 0 && payload.EmpresaID != empresaID {
		http.Error(w, "empresa_id no coincide con el contexto", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(payload.Desde) != "" {
		filter.Desde = strings.TrimSpace(payload.Desde)
	}
	if strings.TrimSpace(payload.Hasta) != "" {
		filter.Hasta = strings.TrimSpace(payload.Hasta)
	}
	if strings.TrimSpace(payload.Usuario) != "" {
		filter.UsuarioCreador = strings.TrimSpace(payload.Usuario)
	}
	if payload.ClienteID > 0 {
		filter.ClienteID = payload.ClienteID
	}
	if strings.TrimSpace(payload.Etiqueta) != "" {
		filter.Etiqueta = strings.TrimSpace(payload.Etiqueta)
	}

	removed, err := dbpkg.SetEmpresaCalculadoraOperacionesEstado(dbEmp, empresaID, filter, "inactivo")
	if err != nil {
		http.Error(w, "No se pudo limpiar historial de calculadora", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":             true,
		"empresa_id":     empresaID,
		"desactivados":   removed,
		"filtros":        calcEmpresaFilterMap(filter),
		"actualizado_en": time.Now().Format("2006-01-02 15:04:05"),
	})
}

func calcEmpresaHandleExport(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	filter, err := calcEmpresaBuildFilterFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	limit, offset := calcEmpresaNormalizeLimitOffset(filter.Limit, filter.Offset)
	filter.Limit = limit
	filter.Offset = offset

	rows, total, err := dbpkg.ListEmpresaCalculadoraOperaciones(dbEmp, empresaID, filter)
	if err != nil {
		http.Error(w, "No se pudo consultar historial para exportacion", http.StatusInternalServerError)
		return
	}

	sumResult := 0.0
	numericCount := 0
	exportRows := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		if v, ok := calcEmpresaTryParseNumeric(row.Resultado); ok {
			sumResult += v
			numericCount++
		}
		exportRows = append(exportRows, map[string]interface{}{
			"id":               row.ID,
			"empresa_id":       row.EmpresaID,
			"fecha_operacion":  row.FechaOperacion,
			"expresion":        row.Expresion,
			"resultado":        row.Resultado,
			"etiquetas":        strings.Join(row.Etiquetas, ", "),
			"cliente_id":       row.ClienteID,
			"cliente_nombre":   row.ClienteNombre,
			"documento_tipo":   row.DocumentoTipo,
			"documento_codigo": row.DocumentoCodigo,
			"carrito_id":       row.CarritoID,
			"cotizacion_id":    row.CotizacionID,
			"usuario_creador":  row.UsuarioCreador,
			"estado":           row.Estado,
			"observaciones":    row.Observaciones,
		})
	}

	ds := empresaReporteDataset{
		Key:         "operativo_calculadora_historial",
		Title:       "Historial de Calculadora Empresarial",
		Level:       "operativo",
		Description: "Operaciones registradas por la calculadora de empresa con trazabilidad por usuario, cliente y documento.",
		EmpresaID:   empresaID,
		Desde:       filter.Desde,
		Hasta:       filter.Hasta,
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		Columns: []string{
			"id",
			"empresa_id",
			"fecha_operacion",
			"expresion",
			"resultado",
			"etiquetas",
			"cliente_id",
			"cliente_nombre",
			"documento_tipo",
			"documento_codigo",
			"carrito_id",
			"cotizacion_id",
			"usuario_creador",
			"estado",
			"observaciones",
		},
		Rows:     exportRows,
		RowCount: len(exportRows),
		Summary: map[string]interface{}{
			"total_coincidencias":            total,
			"sumatoria_resultados_numericos": math.Round(sumResult*100) / 100,
			"resultados_numericos":           numericCount,
			"filtros":                        calcEmpresaFilterMap(filter),
		},
	}

	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	if format == "" {
		format = "json"
	}
	if err := writeReportesDatasetExport(w, ds, format); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func calcEmpresaHandleReferences(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	limit, err := parseIntQueryOptional(r, "limit")
	if err != nil {
		http.Error(w, "limit invalido", http.StatusBadRequest)
		return
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))

	cfg, err := dbpkg.GetEmpresaCalculadoraConfiguracion(dbEmp, empresaID)
	if err != nil {
		http.Error(w, "No se pudo consultar configuracion de calculadora", http.StatusInternalServerError)
		return
	}

	clientes, err := dbpkg.GetClientesByEmpresa(dbEmp, empresaID, false, q)
	if err != nil {
		http.Error(w, "No se pudo consultar clientes", http.StatusInternalServerError)
		return
	}
	if len(clientes) > limit {
		clientes = clientes[:limit]
	}

	clientesRows := make([]map[string]interface{}, 0, len(clientes))
	for _, cli := range clientes {
		clientesRows = append(clientesRows, map[string]interface{}{
			"id":                  cli.ID,
			"nombre_razon_social": cli.NombreRazonSocial,
			"numero_documento":    cli.NumeroDocumento,
			"tipo_documento":      cli.TipoDocumento,
			"email":               cli.Email,
		})
	}

	carritosRows := make([]map[string]interface{}, 0)
	if cfg != nil && cfg.IntegrarCarritos {
		carritos, carritosErr := dbpkg.GetCarritosCompraByEmpresa(dbEmp, empresaID, false, q)
		if carritosErr != nil {
			http.Error(w, "No se pudo consultar carritos", http.StatusInternalServerError)
			return
		}
		if len(carritos) > limit {
			carritos = carritos[:limit]
		}
		for _, item := range carritos {
			carritosRows = append(carritosRows, map[string]interface{}{
				"id":             item.ID,
				"codigo":         item.Codigo,
				"nombre":         item.Nombre,
				"cliente_id":     item.ClienteID,
				"cliente_nombre": item.ClienteNombre,
				"total":          item.Total,
				"estado_venta":   item.EstadoVenta,
				"fecha_creacion": item.FechaCreacion,
			})
		}
	}

	cotizacionesRows := make([]map[string]interface{}, 0)
	if cfg != nil && cfg.IntegrarCotizaciones {
		items, listErr := dbpkg.ListEmpresaGenericRows(dbEmp, "empresa_cotizaciones_venta", empresaID, dbpkg.EmpresaGenericListFilter{
			IncludeInactive: false,
			Q:               q,
			Limit:           limit,
			Offset:          0,
			SearchColumns:   []string{"codigo", "cliente_nombre", "estado_documento", "notas"},
		})
		if listErr != nil {
			http.Error(w, "No se pudo consultar cotizaciones", http.StatusInternalServerError)
			return
		}
		for _, item := range items {
			cotizacionesRows = append(cotizacionesRows, map[string]interface{}{
				"id":               calcEmpresaMapInt64(item, "id"),
				"codigo":           calcEmpresaMapString(item, "codigo"),
				"cliente_id":       calcEmpresaMapInt64(item, "cliente_id"),
				"cliente_nombre":   calcEmpresaMapString(item, "cliente_nombre"),
				"estado_documento": calcEmpresaMapString(item, "estado_documento"),
				"fecha_documento":  calcEmpresaMapString(item, "fecha_documento"),
				"total":            calcEmpresaMapFloat(item, "total"),
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":            true,
		"empresa_id":    empresaID,
		"q":             q,
		"limit":         limit,
		"configuracion": cfg,
		"clientes":      clientesRows,
		"carritos":      carritosRows,
		"cotizaciones":  cotizacionesRows,
	})
}

func calcEmpresaBuildMetadataJSON(raw string, metadata map[string]interface{}) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw != "" {
		if !json.Valid([]byte(raw)) {
			return "", errors.New("metadata_json invalido")
		}
		return raw, nil
	}
	if metadata == nil {
		return "{}", nil
	}
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func calcEmpresaBuildFilterFromRequest(r *http.Request) (dbpkg.EmpresaCalculadoraOperacionFilter, error) {
	limit, err := parseIntQueryOptional(r, "limit")
	if err != nil {
		return dbpkg.EmpresaCalculadoraOperacionFilter{}, fmt.Errorf("limit invalido")
	}
	offset, err := parseIntQueryOptional(r, "offset")
	if err != nil {
		return dbpkg.EmpresaCalculadoraOperacionFilter{}, fmt.Errorf("offset invalido")
	}
	if limit < 0 {
		return dbpkg.EmpresaCalculadoraOperacionFilter{}, fmt.Errorf("limit invalido")
	}
	if offset < 0 {
		return dbpkg.EmpresaCalculadoraOperacionFilter{}, fmt.Errorf("offset invalido")
	}
	clienteID, err := parseInt64QueryOptional(r, "cliente_id")
	if err != nil {
		return dbpkg.EmpresaCalculadoraOperacionFilter{}, fmt.Errorf("cliente_id invalido")
	}
	if clienteID < 0 {
		return dbpkg.EmpresaCalculadoraOperacionFilter{}, fmt.Errorf("cliente_id invalido")
	}
	desde, err := calcEmpresaNormalizeDateValue(r.URL.Query().Get("desde"))
	if err != nil {
		return dbpkg.EmpresaCalculadoraOperacionFilter{}, fmt.Errorf("desde invalido")
	}
	hasta, err := calcEmpresaNormalizeDateValue(r.URL.Query().Get("hasta"))
	if err != nil {
		return dbpkg.EmpresaCalculadoraOperacionFilter{}, fmt.Errorf("hasta invalido")
	}
	if desde != "" && hasta != "" {
		desdeDate, _ := time.ParseInLocation("2006-01-02", desde, time.Local)
		hastaDate, _ := time.ParseInLocation("2006-01-02", hasta, time.Local)
		if desdeDate.After(hastaDate) {
			return dbpkg.EmpresaCalculadoraOperacionFilter{}, fmt.Errorf("rango de fechas invalido")
		}
	}

	filter := dbpkg.EmpresaCalculadoraOperacionFilter{
		Desde:           desde,
		Hasta:           hasta,
		UsuarioCreador:  strings.TrimSpace(r.URL.Query().Get("usuario")),
		ClienteID:       clienteID,
		Etiqueta:        strings.TrimSpace(r.URL.Query().Get("etiqueta")),
		IncludeInactive: queryBool(r, "include_inactive"),
		Limit:           limit,
		Offset:          offset,
	}

	if strings.TrimSpace(filter.UsuarioCreador) == "" {
		filter.UsuarioCreador = strings.TrimSpace(r.URL.Query().Get("usuario_creador"))
	}

	return filter, nil
}

func calcEmpresaDecodeJSONBody(r *http.Request, target interface{}) error {
	if r == nil || r.Body == nil {
		return nil
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
	return nil
}

func calcEmpresaNormalizeAction(raw string) string {
	action := strings.ToLower(strings.TrimSpace(raw))
	action = strings.ReplaceAll(action, "-", "_")
	action = strings.ReplaceAll(action, " ", "_")
	return action
}

func calcEmpresaDefaultActionByMethod(method string) string {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case http.MethodGet:
		return "historial"
	case http.MethodPost:
		return "registrar"
	case http.MethodPut, http.MethodPatch:
		return "config"
	case http.MethodDelete:
		return "desactivar"
	default:
		return ""
	}
}

func calcEmpresaUsuarioFromRequest(r *http.Request) string {
	if r == nil {
		return "sistema"
	}
	if v := strings.TrimSpace(adminEmailFromRequest(r)); v != "" {
		return strings.ToLower(v)
	}
	if v := strings.TrimSpace(r.Header.Get("X-Usuario-Email")); v != "" {
		return strings.ToLower(v)
	}
	if v := strings.TrimSpace(r.Header.Get("X-User-Email")); v != "" {
		return strings.ToLower(v)
	}
	return "sistema"
}

func calcEmpresaNormalizeDateValue(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil
	}
	layouts := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed.Format("2006-01-02"), nil
		}
	}
	return "", fmt.Errorf("fecha invalida")
}

func calcEmpresaNormalizeLimitOffset(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func calcEmpresaFilterMap(filter dbpkg.EmpresaCalculadoraOperacionFilter) map[string]interface{} {
	return map[string]interface{}{
		"desde":            filter.Desde,
		"hasta":            filter.Hasta,
		"usuario_creador":  filter.UsuarioCreador,
		"cliente_id":       filter.ClienteID,
		"etiqueta":         filter.Etiqueta,
		"include_inactive": filter.IncludeInactive,
		"limit":            filter.Limit,
		"offset":           filter.Offset,
	}
}

func calcEmpresaTryParseNumeric(raw string) (float64, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, false
	}
	value = strings.ReplaceAll(value, ",", "")
	n, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, false
	}
	if math.IsNaN(n) || math.IsInf(n, 0) {
		return 0, false
	}
	return n, true
}

func calcEmpresaMapString(row map[string]interface{}, key string) string {
	v, ok := row[key]
	if !ok || v == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprintf("%v", v))
}

func calcEmpresaMapInt64(row map[string]interface{}, key string) int64 {
	v, ok := row[key]
	if !ok || v == nil {
		return 0
	}
	s := strings.TrimSpace(fmt.Sprintf("%v", v))
	if s == "" {
		return 0
	}
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return int64(f)
	}
	return 0
}

func calcEmpresaMapFloat(row map[string]interface{}, key string) float64 {
	v, ok := row[key]
	if !ok || v == nil {
		return 0
	}
	s := strings.TrimSpace(fmt.Sprintf("%v", v))
	if s == "" {
		return 0
	}
	s = strings.ReplaceAll(s, ",", "")
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}
