package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type superReportesDatasetEmpresaItem struct {
	Empresa dbpkg.Empresa         `json:"empresa"`
	Dataset empresaReporteDataset `json:"dataset"`
}

type superReportesTableroEmpresaItem struct {
	Empresa dbpkg.Empresa                       `json:"empresa"`
	Tablero dbpkg.EmpresaReportesTableroResumen `json:"tablero"`
}

type superReportesTableroTotales struct {
	EmpresasSeleccionadas int64                                  `json:"empresas_seleccionadas"`
	EmpresasActivas       int64                                  `json:"empresas_activas"`
	Operativo             dbpkg.EmpresaReportesTableroOperativo  `json:"operativo"`
	Financiero            dbpkg.EmpresaReportesTableroFinanciero `json:"financiero"`
	Contable              dbpkg.EmpresaReportesTableroContable   `json:"contable"`
	EstadoResultados      dbpkg.EmpresaReportesEstadoResultados  `json:"estado_resultados"`
	BalanceGeneral        dbpkg.EmpresaReportesBalanceGeneral    `json:"balance_general"`
}

type superReportesTableroResponse struct {
	AdminEmail string                            `json:"admin_email"`
	Desde      string                            `json:"desde"`
	Hasta      string                            `json:"hasta"`
	GeneradoEn string                            `json:"generado_en"`
	Empresas   []dbpkg.Empresa                   `json:"empresas"`
	Totales    superReportesTableroTotales       `json:"totales"`
	PorEmpresa []superReportesTableroEmpresaItem `json:"por_empresa"`
}

type superReportesDatasetResponse struct {
	AdminEmail   string                            `json:"admin_email"`
	Modo         string                            `json:"modo"`
	DatasetKey   string                            `json:"dataset_key"`
	DatasetTitle string                            `json:"dataset_title"`
	Desde        string                            `json:"desde"`
	Hasta        string                            `json:"hasta"`
	GeneradoEn   string                            `json:"generado_en"`
	Empresas     []dbpkg.Empresa                   `json:"empresas"`
	Combinado    empresaReporteDataset             `json:"combinado"`
	Individuales []superReportesDatasetEmpresaItem `json:"individuales"`
}

// SuperReportesGlobalesHandler devuelve reportes globales de las empresas creadas por el administrador autenticado.
func SuperReportesGlobalesHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminEmail := strings.TrimSpace(adminEmailFromRequest(r))
		if adminEmail == "" || strings.EqualFold(adminEmail, "sistema") {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}

		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "" {
			action = "tablero"
		}
		modo := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("modo")))
		if modo == "" {
			modo = "consolidado"
		}
		if modo != "consolidado" && modo != "individual" {
			http.Error(w, "modo invalido (use consolidado o individual)", http.StatusBadRequest)
			return
		}

		_, principalEmail, err := resolveRequesterAdminScope(dbSuper, r)
		if err != nil {
			http.Error(w, "no se pudo resolver el alcance del administrador", http.StatusInternalServerError)
			return
		}
		empresas, err := superReportesEmpresasPermitidas(dbEmp, dbSuper, adminEmail, principalEmail)
		if err != nil {
			http.Error(w, "no se pudieron cargar las empresas del administrador", http.StatusInternalServerError)
			return
		}
		empresasSeleccionadas, err := superReportesResolveEmpresasSeleccionadas(empresas, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch action {
		case "catalogo", "catalog", "datasets":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"admin_email":     adminEmail,
				"empresas":        empresas,
				"datasets":        reportesCatalogo,
				"default_dataset": reporteDatasetOperativoVentasDetalle,
				"default_mode":    "consolidado",
			})
			return
		case "tablero", "dashboard":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			resp, err := superReportesBuildTablero(dbEmp, adminEmail, empresasSeleccionadas, r)
			if err != nil {
				http.Error(w, "no se pudo construir el tablero global", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, resp)
			return
		case "dataset":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			datasetKey := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("dataset")))
			if datasetKey == "" {
				http.Error(w, "dataset es obligatorio", http.StatusBadRequest)
				return
			}
			resp, err := superReportesBuildDatasetResponse(dbEmp, adminEmail, empresasSeleccionadas, datasetKey, modo, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, resp)
			return
		case "export", "exportar":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			datasetKey := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("dataset")))
			if datasetKey == "" {
				http.Error(w, "dataset es obligatorio", http.StatusBadRequest)
				return
			}
			format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
			if format == "" {
				format = "json"
			}
			resp, err := superReportesBuildDatasetResponse(dbEmp, adminEmail, empresasSeleccionadas, datasetKey, modo, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if modo == "individual" && format == "json" {
				fileName := "reportes_globales_" + datasetKey + "_admin_" + superReportesSafeFileLabel(adminEmail) + "_" + time.Now().Format("20060102_150405") + ".json"
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
				_ = json.NewEncoder(w).Encode(resp)
				return
			}
			if err := writeReportesDatasetExport(w, resp.Combinado, format); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			return
		case "enviar_email", "email", "send_email":
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			var payload struct {
				ToEmail string `json:"to_email"`
				Format  string `json:"format"`
				Dataset string `json:"dataset"`
				Modo    string `json:"modo"`
				Desde   string `json:"desde"`
				Hasta   string `json:"hasta"`
				// opcional: filtrar a empresas específicas (mismo contrato que API)
				EmpresaID  int64  `json:"empresa_id,omitempty"`
				EmpresaIDs string `json:"empresa_ids,omitempty"`
				Subject    string `json:"subject,omitempty"`
				Message    string `json:"message,omitempty"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "json invalido", http.StatusBadRequest)
				return
			}
			// reconstruir query con el contrato existente para reutilizar validaciones.
			q := r.URL.Query()
			if strings.TrimSpace(payload.Dataset) != "" {
				q.Set("dataset", strings.TrimSpace(payload.Dataset))
			}
			if strings.TrimSpace(payload.Modo) != "" {
				q.Set("modo", strings.TrimSpace(payload.Modo))
			}
			if strings.TrimSpace(payload.Desde) != "" {
				q.Set("desde", strings.TrimSpace(payload.Desde))
			}
			if strings.TrimSpace(payload.Hasta) != "" {
				q.Set("hasta", strings.TrimSpace(payload.Hasta))
			}
			if payload.EmpresaID > 0 {
				q.Set("empresa_id", strconv.FormatInt(payload.EmpresaID, 10))
			}
			if strings.TrimSpace(payload.EmpresaIDs) != "" {
				q.Set("empresa_ids", strings.TrimSpace(payload.EmpresaIDs))
			}
			r.URL.RawQuery = q.Encode()

			datasetKey := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("dataset")))
			if datasetKey == "" {
				http.Error(w, "dataset es obligatorio", http.StatusBadRequest)
				return
			}
			format := strings.ToLower(strings.TrimSpace(payload.Format))
			if format == "" {
				format = "pdf"
			}
			modo := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("modo")))
			if modo == "" {
				modo = "consolidado"
			}
			if modo != "consolidado" && modo != "individual" {
				http.Error(w, "modo invalido (use consolidado o individual)", http.StatusBadRequest)
				return
			}
			resp, err := superReportesBuildDatasetResponse(dbEmp, adminEmail, empresasSeleccionadas, datasetKey, modo, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			empresaLabel := ""
			if len(empresasSeleccionadas) == 1 {
				empresaLabel = strings.TrimSpace(empresasSeleccionadas[0].Nombre)
			}
			subject := strings.TrimSpace(payload.Subject)
			if subject == "" {
				subject = reportesDefaultEmailSubject("Reporte global", strings.TrimSpace(resp.DatasetTitle), empresaLabel)
			}
			body := strings.TrimSpace(payload.Message)
			if body == "" {
				body = "Adjunto encontrarás el reporte solicitado desde Reportes globales."
			}

			var fileName, contentType string
			var content []byte
			if modo == "individual" && strings.ToLower(format) == "json" {
				raw, jerr := json.Marshal(resp)
				if jerr != nil {
					http.Error(w, "no se pudo serializar el reporte", http.StatusInternalServerError)
					return
				}
				fileName = "reportes_globales_" + datasetKey + "_admin_" + superReportesSafeFileLabel(adminEmail) + "_" + time.Now().Format("20060102_150405") + ".json"
				contentType = "application/json"
				content = raw
			} else {
				fileName, contentType, content, err = reportesBuildExportBytes(resp.Combinado, format)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			}
			metaJSON := fmt.Sprintf(`{"scope":"super_reportes_globales","modo":%q,"dataset":%q,"format":%q,"desde":%q,"hasta":%q,"empresa_ids":%q}`, modo, datasetKey, format, strings.TrimSpace(payload.Desde), strings.TrimSpace(payload.Hasta), strings.TrimSpace(payload.EmpresaIDs))
			if err := sendReportesEmailWithAttachment(r, dbSuper, 0, payload.ToEmail, subject, body, fileName, contentType, content, metaJSON); err != nil {
				http.Error(w, "no se pudo enviar el correo: "+err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":       true,
				"to_email": strings.TrimSpace(payload.ToEmail),
				"filename": fileName,
				"format":   format,
			})
			return
		default:
			http.Error(w, "action invalida (use catalogo, tablero, dataset o export)", http.StatusBadRequest)
			return
		}
	}
}

func superReportesEmpresasPermitidas(dbEmp, dbSuper *sql.DB, adminEmail, principalEmail string) ([]dbpkg.Empresa, error) {
	adminEmail = strings.ToLower(strings.TrimSpace(adminEmail))
	empresas, err := dbpkg.GetEmpresas(dbEmp)
	if err != nil {
		return nil, err
	}
	return superReportesFiltrarEmpresasPermitidas(dbSuper, adminEmail, principalEmail, empresas)
}

func superReportesFiltrarEmpresasPermitidas(dbSuper *sql.DB, adminEmail, principalEmail string, empresas []dbpkg.Empresa) ([]dbpkg.Empresa, error) {
	return decorateEmpresasByEffectiveAccess(dbSuper, adminEmail, principalEmail, empresas)
}

func superReportesResolveEmpresasSeleccionadas(empresas []dbpkg.Empresa, r *http.Request) ([]dbpkg.Empresa, error) {
	if len(empresas) == 0 {
		return []dbpkg.Empresa{}, nil
	}
	permitidas := make(map[int64]dbpkg.Empresa, len(empresas))
	for _, empresa := range empresas {
		permitidas[empresa.ID] = empresa
		if empresa.EmpresaID > 0 {
			permitidas[empresa.EmpresaID] = empresa
		}
	}

	raw := strings.TrimSpace(r.URL.Query().Get("empresa_ids"))
	if raw == "" {
		if single := strings.TrimSpace(r.URL.Query().Get("empresa_id")); single != "" {
			raw = single
		}
	}
	if raw == "" {
		return append([]dbpkg.Empresa(nil), empresas...), nil
	}

	parts := strings.Split(raw, ",")
	selected := make([]dbpkg.Empresa, 0, len(parts))
	seen := make(map[int64]struct{}, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.ParseInt(part, 10, 64)
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("empresa_ids invalido")
		}
		empresa, ok := permitidas[id]
		if !ok {
			return nil, fmt.Errorf("empresa_id fuera de alcance")
		}
		if _, exists := seen[empresa.ID]; exists {
			continue
		}
		seen[empresa.ID] = struct{}{}
		selected = append(selected, empresa)
	}
	if len(selected) == 0 {
		return []dbpkg.Empresa{}, nil
	}
	return selected, nil
}

func superReportesBuildTablero(dbEmp *sql.DB, adminEmail string, empresas []dbpkg.Empresa, r *http.Request) (superReportesTableroResponse, error) {
	desde := strings.TrimSpace(r.URL.Query().Get("desde"))
	hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
	resp := superReportesTableroResponse{
		AdminEmail: adminEmail,
		Desde:      desde,
		Hasta:      hasta,
		GeneradoEn: time.Now().Format("2006-01-02 15:04:05"),
		Empresas:   empresas,
		PorEmpresa: make([]superReportesTableroEmpresaItem, 0, len(empresas)),
	}

	for _, empresa := range empresas {
		builder := &reportesBuilder{db: dbEmp, empresaID: empresa.ID, desde: desde, hasta: hasta, maxRows: 200, itemsCache: make(map[int64][]dbpkg.CarritoCompraItem)}
		tablero, err := builder.getTableroResumen()
		if err != nil {
			return superReportesTableroResponse{}, err
		}
		if tablero == nil {
			continue
		}
		resp.PorEmpresa = append(resp.PorEmpresa, superReportesTableroEmpresaItem{Empresa: empresa, Tablero: *tablero})
		resp.Totales.EmpresasSeleccionadas++
		if reportesEstadoActivo(empresa.Estado) {
			resp.Totales.EmpresasActivas++
		}
		superReportesAccumulateTablero(&resp.Totales, tablero)
	}
	return resp, nil
}

func superReportesAccumulateTablero(dst *superReportesTableroTotales, src *dbpkg.EmpresaReportesTableroResumen) {
	if dst == nil || src == nil {
		return
	}
	dst.Operativo.VentasCerradas += src.Operativo.VentasCerradas
	dst.Operativo.VentasHoy += src.Operativo.VentasHoy
	dst.Operativo.IngresosVentas = reportesRound(dst.Operativo.IngresosVentas + src.Operativo.IngresosVentas)
	dst.Operativo.ClientesActivos += src.Operativo.ClientesActivos
	dst.Operativo.ProductosActivos += src.Operativo.ProductosActivos
	dst.Operativo.ProductosBajoMinimo += src.Operativo.ProductosBajoMinimo
	dst.Operativo.ComprasMovimientos += src.Operativo.ComprasMovimientos
	dst.Operativo.ComprasCosto = reportesRound(dst.Operativo.ComprasCosto + src.Operativo.ComprasCosto)
	if dst.Operativo.VentasCerradas > 0 {
		dst.Operativo.TicketPromedio = reportesRound(dst.Operativo.IngresosVentas / float64(dst.Operativo.VentasCerradas))
	}

	dst.Financiero.MovimientosIngresos += src.Financiero.MovimientosIngresos
	dst.Financiero.MovimientosEgresos += src.Financiero.MovimientosEgresos
	dst.Financiero.Ingresos = reportesRound(dst.Financiero.Ingresos + src.Financiero.Ingresos)
	dst.Financiero.Egresos = reportesRound(dst.Financiero.Egresos + src.Financiero.Egresos)
	dst.Financiero.Balance = reportesRound(dst.Financiero.Balance + src.Financiero.Balance)
	dst.Financiero.PeriodosAbiertos += src.Financiero.PeriodosAbiertos
	dst.Financiero.PeriodosCerrados += src.Financiero.PeriodosCerrados

	dst.Contable.EventosPendientes += src.Contable.EventosPendientes
	dst.Contable.EventosProcesados += src.Contable.EventosProcesados
	dst.Contable.EventosTotal += src.Contable.EventosTotal
	dst.Contable.EventosMontoTotal = reportesRound(dst.Contable.EventosMontoTotal + src.Contable.EventosMontoTotal)
	dst.Contable.AsientosGenerados += src.Contable.AsientosGenerados
	dst.Contable.AsientosMontoTotal = reportesRound(dst.Contable.AsientosMontoTotal + src.Contable.AsientosMontoTotal)
	dst.Contable.DocumentosFacturacionActivos += src.Contable.DocumentosFacturacionActivos
	dst.Contable.DocumentosComprasActivos += src.Contable.DocumentosComprasActivos

	dst.EstadoResultados.Ingresos = reportesRound(dst.EstadoResultados.Ingresos + src.EstadoResultados.Ingresos)
	dst.EstadoResultados.Gastos = reportesRound(dst.EstadoResultados.Gastos + src.EstadoResultados.Gastos)
	dst.EstadoResultados.UtilidadOperacional = reportesRound(dst.EstadoResultados.UtilidadOperacional + src.EstadoResultados.UtilidadOperacional)

	dst.BalanceGeneral.Activos = reportesRound(dst.BalanceGeneral.Activos + src.BalanceGeneral.Activos)
	dst.BalanceGeneral.Pasivos = reportesRound(dst.BalanceGeneral.Pasivos + src.BalanceGeneral.Pasivos)
	dst.BalanceGeneral.Patrimonio = reportesRound(dst.BalanceGeneral.Patrimonio + src.BalanceGeneral.Patrimonio)
	dst.BalanceGeneral.ResultadoEjercicio = reportesRound(dst.BalanceGeneral.ResultadoEjercicio + src.BalanceGeneral.ResultadoEjercicio)
	dst.BalanceGeneral.Cuadre = reportesRound(dst.BalanceGeneral.Cuadre + src.BalanceGeneral.Cuadre)
}

func superReportesBuildDatasetResponse(dbEmp *sql.DB, adminEmail string, empresas []dbpkg.Empresa, datasetKey, modo string, r *http.Request) (superReportesDatasetResponse, error) {
	desde := strings.TrimSpace(r.URL.Query().Get("desde"))
	hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
	maxRows, err := parseIntQueryOptional(r, "max_rows")
	if err != nil {
		return superReportesDatasetResponse{}, fmt.Errorf("max_rows invalido")
	}
	if maxRows <= 0 {
		maxRows = 300
	}
	if maxRows > 1000 {
		maxRows = 1000
	}
	cierreID, err := parseInt64QueryOptional(r, "cierre_id")
	if err != nil {
		return superReportesDatasetResponse{}, fmt.Errorf("cierre_id invalido")
	}
	empleadoNominaID, err := parseInt64QueryOptional(r, "empleado_nomina_id")
	if err != nil {
		return superReportesDatasetResponse{}, fmt.Errorf("empleado_nomina_id invalido")
	}
	includeInactive := queryBool(r, "include_inactive")

	resp := superReportesDatasetResponse{
		AdminEmail:   adminEmail,
		Modo:         modo,
		DatasetKey:   datasetKey,
		DatasetTitle: superReportesDatasetTitle(datasetKey),
		Desde:        desde,
		Hasta:        hasta,
		GeneradoEn:   time.Now().Format("2006-01-02 15:04:05"),
		Empresas:     empresas,
		Individuales: make([]superReportesDatasetEmpresaItem, 0, len(empresas)),
	}

	if err := dbpkg.EnsureEmpresaReportesProgramacionSchema(dbEmp); err != nil {
		return superReportesDatasetResponse{}, fmt.Errorf("no se pudo inicializar la programacion de reportes")
	}

	for _, empresa := range empresas {
		builder := &reportesBuilder{
			db:               dbEmp,
			empresaID:        empresa.ID,
			desde:            desde,
			hasta:            hasta,
			cierreID:         cierreID,
			empleadoNominaID: empleadoNominaID,
			cajaCodigo:       strings.TrimSpace(r.URL.Query().Get("caja_codigo")),
			turno:            strings.TrimSpace(r.URL.Query().Get("turno")),
			usuario:          strings.TrimSpace(r.URL.Query().Get("usuario")),
			categoria:        strings.TrimSpace(r.URL.Query().Get("categoria")),
			metodoPago:       strings.TrimSpace(r.URL.Query().Get("metodo_pago")),
			maxRows:          maxRows,
			includeInactive:  includeInactive,
			itemsCache:       make(map[int64][]dbpkg.CarritoCompraItem),
		}
		ds, err := builder.buildDataset(datasetKey)
		if err != nil {
			return superReportesDatasetResponse{}, err
		}
		resp.Individuales = append(resp.Individuales, superReportesDatasetEmpresaItem{Empresa: empresa, Dataset: ds})
	}
	resp.Combinado = superReportesMergeDatasets(datasetKey, desde, hasta, resp.Individuales)
	return resp, nil
}

func superReportesMergeDatasets(datasetKey, desde, hasta string, items []superReportesDatasetEmpresaItem) empresaReporteDataset {
	meta := empresaReporteCatalogoItem{Key: datasetKey, Title: datasetKey, Level: "operativo"}
	for _, item := range reportesCatalogo {
		if item.Key == datasetKey {
			meta = item
			break
		}
	}
	columns := []string{"empresa_id", "empresa_nombre", "empresa_nit"}
	seenCols := map[string]struct{}{"empresa_id": {}, "empresa_nombre": {}, "empresa_nit": {}}
	for _, item := range items {
		for _, col := range item.Dataset.Columns {
			if _, exists := seenCols[col]; exists {
				continue
			}
			seenCols[col] = struct{}{}
			columns = append(columns, col)
		}
	}

	merged := empresaReporteDataset{
		Key:         meta.Key,
		Title:       meta.Title + " - Global administrador",
		Level:       meta.Level,
		Description: meta.Description,
		EmpresaID:   0,
		Desde:       desde,
		Hasta:       hasta,
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		Columns:     columns,
		Rows:        make([]map[string]interface{}, 0),
		Summary: map[string]interface{}{
			"empresas_incluidas": len(items),
		},
	}

	for _, item := range items {
		for _, row := range item.Dataset.Rows {
			mergedRow := map[string]interface{}{
				"empresa_id":     item.Empresa.ID,
				"empresa_nombre": item.Empresa.Nombre,
				"empresa_nit":    item.Empresa.Nit,
			}
			for key, value := range row {
				mergedRow[key] = value
			}
			merged.Rows = append(merged.Rows, mergedRow)
		}
		superReportesMergeSummary(merged.Summary, item.Dataset.Summary)
	}
	merged.RowCount = len(merged.Rows)
	if merged.Summary == nil {
		merged.Summary = map[string]interface{}{}
	}
	merged.Summary["filas_totales"] = merged.RowCount
	return merged
}

func superReportesMergeSummary(dst map[string]interface{}, src map[string]interface{}) {
	if dst == nil || src == nil {
		return
	}
	for key, value := range src {
		switch typed := value.(type) {
		case int:
			dst[key] = superReportesToFloat(dst[key]) + float64(typed)
		case int32:
			dst[key] = superReportesToFloat(dst[key]) + float64(typed)
		case int64:
			dst[key] = superReportesToFloat(dst[key]) + float64(typed)
		case float32:
			dst[key] = reportesRound(superReportesToFloat(dst[key]) + float64(typed))
		case float64:
			dst[key] = reportesRound(superReportesToFloat(dst[key]) + typed)
		default:
			if _, exists := dst[key]; !exists {
				dst[key] = value
			}
		}
	}
}

func superReportesToFloat(v interface{}) float64 {
	switch typed := v.(type) {
	case int:
		return float64(typed)
	case int32:
		return float64(typed)
	case int64:
		return float64(typed)
	case float32:
		return float64(typed)
	case float64:
		return typed
	default:
		return 0
	}
}

func superReportesDatasetTitle(datasetKey string) string {
	for _, item := range reportesCatalogo {
		if item.Key == datasetKey {
			return item.Title
		}
	}
	return datasetKey
}

func superReportesSafeFileLabel(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "admin"
	}
	replacer := strings.NewReplacer("@", "_", ".", "_", " ", "_", "/", "_", "\\", "_")
	return replacer.Replace(value)
}
