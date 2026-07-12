package handlers

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type empresaFinanzasImportacionBancariaPayload struct {
	EmpresaID       int64                                     `json:"empresa_id"`
	CuentaBancaria  string                                    `json:"cuenta_bancaria"`
	BancoNombre     string                                    `json:"banco_nombre"`
	Origen          string                                    `json:"origen"`
	AutoConciliar   bool                                      `json:"auto_conciliar"`
	ToleranciaDias  int                                       `json:"tolerancia_dias"`
	ToleranciaMonto float64                                   `json:"tolerancia_monto"`
	Limit           int                                       `json:"limit"`
	Movimientos     []dbpkg.EmpresaFinanzasMovimientoBancario `json:"movimientos"`
}

type empresaFinanzasPeriodoAutorizacionPayload struct {
	AutorizadoPor         string `json:"autorizado_por"`
	MotivoAutorizacion    string `json:"motivo_autorizacion"`
	Motivo                string `json:"motivo"`
	EvidenciaAutorizacion string `json:"evidencia_autorizacion"`
	CodigoAutorizacion    string `json:"codigo_autorizacion"`
}

var empresaComprobanteAllowedExt = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".webp": true,
	".pdf":  true,
	".txt":  true,
	".csv":  true,
	".doc":  true,
	".docx": true,
	".xls":  true,
	".xlsx": true,
}

func sanitizeComprobanteBaseName(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" {
		return "archivo"
	}
	var b strings.Builder
	b.Grow(len(v))
	for _, r := range v {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	name := strings.Trim(b.String(), "_")
	if name == "" {
		return "archivo"
	}
	return name
}

func saveEmpresaComprobanteUpload(file io.Reader, originalFilename string, empresaID int64, modulo, baseName string) (string, string, string, error) {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(originalFilename)))
	if !empresaComprobanteAllowedExt[ext] {
		return "", "", "", fmt.Errorf("extension de comprobante no permitida")
	}

	webRoot := resolveWebRootDir()
	dir := filepath.Join(webRoot, "uploads", "comprobantes", fmt.Sprintf("empresa_%d", empresaID), sanitizeComprobanteBaseName(modulo))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", "", "", err
	}

	fileName := sanitizeComprobanteBaseName(baseName) + "_" + strconv.FormatInt(time.Now().UnixNano(), 10) + ext
	absPath := filepath.Join(dir, fileName)
	// #nosec G304 -- path is normalized and constrained to a server-controlled root before this operation.
	out, err := os.Create(absPath)
	if err != nil {
		return "", "", "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return "", "", "", err
	}

	fileURL := "/uploads/comprobantes/empresa_" + strconv.FormatInt(empresaID, 10) + "/" + sanitizeComprobanteBaseName(modulo) + "/" + fileName
	return fileURL, fileName, absPath, nil
}

func sanitizeFinanzasPeriodoAutorizacionValue(v string) string {
	v = strings.TrimSpace(v)
	v = strings.ReplaceAll(v, "|", "/")
	v = strings.ReplaceAll(v, "\n", " ")
	v = strings.ReplaceAll(v, "\r", " ")
	return strings.Join(strings.Fields(v), " ")
}

func parseEmpresaFinanzasPeriodoAutorizacionPayload(r *http.Request, fallbackUsuario string) (empresaFinanzasPeriodoAutorizacionPayload, error) {
	var payload empresaFinanzasPeriodoAutorizacionPayload
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && err != io.EOF {
			return payload, errors.New("JSON invalido para autorizacion de periodo")
		}
	}

	if strings.TrimSpace(payload.AutorizadoPor) == "" {
		payload.AutorizadoPor = strings.TrimSpace(r.URL.Query().Get("autorizado_por"))
	}
	if strings.TrimSpace(payload.MotivoAutorizacion) == "" {
		payload.MotivoAutorizacion = strings.TrimSpace(r.URL.Query().Get("motivo_autorizacion"))
	}
	if strings.TrimSpace(payload.MotivoAutorizacion) == "" {
		payload.MotivoAutorizacion = strings.TrimSpace(payload.Motivo)
	}
	if strings.TrimSpace(payload.MotivoAutorizacion) == "" {
		payload.MotivoAutorizacion = strings.TrimSpace(r.URL.Query().Get("motivo"))
	}
	if strings.TrimSpace(payload.EvidenciaAutorizacion) == "" {
		payload.EvidenciaAutorizacion = strings.TrimSpace(r.URL.Query().Get("evidencia_autorizacion"))
	}
	if strings.TrimSpace(payload.CodigoAutorizacion) == "" {
		payload.CodigoAutorizacion = strings.TrimSpace(r.URL.Query().Get("codigo_autorizacion"))
	}

	payload.AutorizadoPor = sanitizeFinanzasPeriodoAutorizacionValue(payload.AutorizadoPor)
	payload.MotivoAutorizacion = sanitizeFinanzasPeriodoAutorizacionValue(payload.MotivoAutorizacion)
	payload.EvidenciaAutorizacion = sanitizeFinanzasPeriodoAutorizacionValue(payload.EvidenciaAutorizacion)
	payload.CodigoAutorizacion = sanitizeFinanzasPeriodoAutorizacionValue(payload.CodigoAutorizacion)

	if payload.AutorizadoPor == "" {
		payload.AutorizadoPor = sanitizeFinanzasPeriodoAutorizacionValue(fallbackUsuario)
	}
	if payload.AutorizadoPor == "" {
		return payload, errors.New("autorizado_por es obligatorio para cerrar o reabrir periodos")
	}
	if payload.MotivoAutorizacion == "" {
		return payload, errors.New("motivo_autorizacion es obligatorio para cerrar o reabrir periodos")
	}
	if payload.EvidenciaAutorizacion == "" {
		return payload, errors.New("evidencia_autorizacion es obligatoria para cerrar o reabrir periodos")
	}

	return payload, nil
}

func buildEmpresaFinanzasPeriodoAutorizacionObservaciones(action string, autorizacion empresaFinanzasPeriodoAutorizacionPayload, ejecutadoPor string) string {
	parts := []string{
		"accion=" + sanitizeFinanzasPeriodoAutorizacionValue(action),
		"autorizado_por=" + sanitizeFinanzasPeriodoAutorizacionValue(autorizacion.AutorizadoPor),
		"motivo=" + sanitizeFinanzasPeriodoAutorizacionValue(autorizacion.MotivoAutorizacion),
		"evidencia=" + sanitizeFinanzasPeriodoAutorizacionValue(autorizacion.EvidenciaAutorizacion),
	}
	if strings.TrimSpace(autorizacion.CodigoAutorizacion) != "" {
		parts = append(parts, "codigo_autorizacion="+sanitizeFinanzasPeriodoAutorizacionValue(autorizacion.CodigoAutorizacion))
	}
	if strings.TrimSpace(ejecutadoPor) != "" {
		parts = append(parts, "ejecutado_por="+sanitizeFinanzasPeriodoAutorizacionValue(ejecutadoPor))
	}
	return strings.Join(parts, " | ")
}

func buildEmpresaConciliacionBancariaDataset(resumen dbpkg.EmpresaConciliacionBancariaResumen) empresaReporteDataset {
	ds := empresaReporteDataset{
		Key:         "contable_conciliacion_bancaria",
		Title:       "Conciliacion bancaria por periodo",
		Level:       "contable",
		Description: "Resumen de extractos bancarios, emparejamiento automatico y desviaciones financieras por periodo",
		EmpresaID:   resumen.EmpresaID,
		Desde:       resumen.Desde,
		Hasta:       resumen.Hasta,
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		Columns: []string{
			"periodo_contable",
			"extractos_total",
			"extractos_conciliados",
			"extractos_pendientes",
			"extractos_con_desviacion",
			"extractos_monto_total",
			"extractos_monto_conciliado",
			"movimientos_internos_total",
			"movimientos_internos_monto",
			"desfase_registros",
			"desfase_monto",
			"estado_conciliacion",
			"ultimo_extracto",
			"ultima_conciliacion",
		},
		Rows:    make([]map[string]interface{}, 0, len(resumen.Filas)),
		Summary: make(map[string]interface{}),
	}

	for _, row := range resumen.Filas {
		ds.Rows = append(ds.Rows, map[string]interface{}{
			"periodo_contable":           row.PeriodoContable,
			"extractos_total":            row.ExtractosTotal,
			"extractos_conciliados":      row.ExtractosConciliados,
			"extractos_pendientes":       row.ExtractosPendientes,
			"extractos_con_desviacion":   row.ExtractosConDesviacion,
			"extractos_monto_total":      row.ExtractosMontoTotal,
			"extractos_monto_conciliado": row.ExtractosMontoConciliado,
			"movimientos_internos_total": row.MovimientosInternosTotal,
			"movimientos_internos_monto": row.MovimientosInternosMonto,
			"desfase_registros":          row.DesfaseRegistros,
			"desfase_monto":              row.DesfaseMonto,
			"estado_conciliacion":        row.EstadoConciliacion,
			"ultimo_extracto":            row.UltimoExtracto,
			"ultima_conciliacion":        row.UltimaConciliacion,
		})
	}
	ds.RowCount = len(ds.Rows)
	ds.Summary["total_periodos"] = resumen.TotalPeriodos
	ds.Summary["periodos_conciliados"] = resumen.PeriodosConciliados
	ds.Summary["periodos_con_pendientes"] = resumen.PeriodosConPendientes
	ds.Summary["periodos_con_descuadre"] = resumen.PeriodosConDescuadre
	ds.Summary["periodos_sin_movimientos"] = resumen.PeriodosSinMovimientos

	return ds
}

func buildTableroResumenExportPayload(resumen *dbpkg.EmpresaReportesTableroResumen) map[string]interface{} {
	if resumen == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"empresa_id":        resumen.EmpresaID,
		"desde":             resumen.Desde,
		"hasta":             resumen.Hasta,
		"generado_en":       resumen.GeneradoEn,
		"operativo":         resumen.Operativo,
		"financiero":        resumen.Financiero,
		"contable":          resumen.Contable,
		"estado_resultados": resumen.EstadoResultados,
		"balance_general":   resumen.BalanceGeneral,
	}
}

func formatTableroMetricFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}

func buildTableroResumenCSVRows(resumen *dbpkg.EmpresaReportesTableroResumen) [][]string {
	rows := [][]string{}
	if resumen == nil {
		return rows
	}

	empresaID := strconv.FormatInt(resumen.EmpresaID, 10)
	addRow := func(bloque, metrica, valor string) {
		rows = append(rows, []string{empresaID, resumen.Desde, resumen.Hasta, resumen.GeneradoEn, bloque, metrica, valor})
	}

	addRow("operativo", "ventas_cerradas", strconv.FormatInt(resumen.Operativo.VentasCerradas, 10))
	addRow("operativo", "ventas_hoy", strconv.FormatInt(resumen.Operativo.VentasHoy, 10))
	addRow("operativo", "ingresos_ventas", formatTableroMetricFloat(resumen.Operativo.IngresosVentas))
	addRow("operativo", "ticket_promedio", formatTableroMetricFloat(resumen.Operativo.TicketPromedio))
	addRow("operativo", "clientes_activos", strconv.FormatInt(resumen.Operativo.ClientesActivos, 10))
	addRow("operativo", "productos_activos", strconv.FormatInt(resumen.Operativo.ProductosActivos, 10))
	addRow("operativo", "productos_bajo_minimo", strconv.FormatInt(resumen.Operativo.ProductosBajoMinimo, 10))
	addRow("operativo", "compras_movimientos", strconv.FormatInt(resumen.Operativo.ComprasMovimientos, 10))
	addRow("operativo", "compras_costo", formatTableroMetricFloat(resumen.Operativo.ComprasCosto))

	addRow("financiero", "movimientos_ingresos", strconv.FormatInt(resumen.Financiero.MovimientosIngresos, 10))
	addRow("financiero", "movimientos_egresos", strconv.FormatInt(resumen.Financiero.MovimientosEgresos, 10))
	addRow("financiero", "ingresos", formatTableroMetricFloat(resumen.Financiero.Ingresos))
	addRow("financiero", "egresos", formatTableroMetricFloat(resumen.Financiero.Egresos))
	addRow("financiero", "balance", formatTableroMetricFloat(resumen.Financiero.Balance))
	addRow("financiero", "periodos_abiertos", strconv.FormatInt(resumen.Financiero.PeriodosAbiertos, 10))
	addRow("financiero", "periodos_cerrados", strconv.FormatInt(resumen.Financiero.PeriodosCerrados, 10))

	addRow("contable", "eventos_pendientes", strconv.FormatInt(resumen.Contable.EventosPendientes, 10))
	addRow("contable", "eventos_procesados", strconv.FormatInt(resumen.Contable.EventosProcesados, 10))
	addRow("contable", "eventos_total", strconv.FormatInt(resumen.Contable.EventosTotal, 10))
	addRow("contable", "eventos_monto_total", formatTableroMetricFloat(resumen.Contable.EventosMontoTotal))
	addRow("contable", "asientos_generados", strconv.FormatInt(resumen.Contable.AsientosGenerados, 10))
	addRow("contable", "asientos_monto_total", formatTableroMetricFloat(resumen.Contable.AsientosMontoTotal))
	addRow("contable", "documentos_facturacion_activos", strconv.FormatInt(resumen.Contable.DocumentosFacturacionActivos, 10))
	addRow("contable", "documentos_compras_activos", strconv.FormatInt(resumen.Contable.DocumentosComprasActivos, 10))

	addRow("estado_resultados", "ingresos", formatTableroMetricFloat(resumen.EstadoResultados.Ingresos))
	addRow("estado_resultados", "gastos", formatTableroMetricFloat(resumen.EstadoResultados.Gastos))
	addRow("estado_resultados", "utilidad_operacional", formatTableroMetricFloat(resumen.EstadoResultados.UtilidadOperacional))

	addRow("balance_general", "activos", formatTableroMetricFloat(resumen.BalanceGeneral.Activos))
	addRow("balance_general", "pasivos", formatTableroMetricFloat(resumen.BalanceGeneral.Pasivos))
	addRow("balance_general", "patrimonio", formatTableroMetricFloat(resumen.BalanceGeneral.Patrimonio))
	addRow("balance_general", "resultado_ejercicio", formatTableroMetricFloat(resumen.BalanceGeneral.ResultadoEjercicio))
	addRow("balance_general", "cuadre", formatTableroMetricFloat(resumen.BalanceGeneral.Cuadre))

	return rows
}

func buildTableroResumenCSVContent(resumen *dbpkg.EmpresaReportesTableroResumen) (string, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	if err := writer.Write([]string{"empresa_id", "desde", "hasta", "generado_en", "bloque", "metrica", "valor"}); err != nil {
		return "", err
	}
	for _, row := range buildTableroResumenCSVRows(resumen) {
		if err := writer.Write(row); err != nil {
			return "", err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}
	return builder.String(), nil
}

func normalizarCajaMovimientoFinanzas(dbEmp *sql.DB, payload *dbpkg.EmpresaFinanzasMovimiento, usuario string) error {
	if payload == nil || payload.EmpresaID <= 0 {
		return nil
	}
	cierreID := payload.CierreCajaID
	cajaCodigo := strings.TrimSpace(payload.CajaCodigo)
	if cierreID <= 0 && cajaCodigo == "" {
		return nil
	}
	cierre, err := dbpkg.GetEmpresaCierreCajaAbiertaUsuario(dbEmp, payload.EmpresaID, cierreID, cajaCodigo, payload.CajaTurno, payload.CajaSucursalID, usuario)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("la caja seleccionada no esta abierta o activa")
		}
		return err
	}
	payload.CierreCajaID = cierre.ID
	payload.CajaCodigo = cierre.CajaCodigo
	payload.CajaTurno = cierre.Turno
	payload.CajaSucursalID = cierre.SucursalID
	return nil
}

func validarPermisoRolMovimientoFinanzasManual(dbEmp *sql.DB, r *http.Request, empresaID int64, tipoMovimiento string) error {
	role := normalizePermissionRole(fmt.Sprint(r.Context().Value("adminRoleEfectivo")))
	if role != "cajero" {
		return nil
	}
	permisos, err := dbpkg.GetEmpresaConfiguracionOperativaPermisos(dbEmp, empresaID, role)
	if err != nil {
		return fmt.Errorf("no se pudo validar la configuracion operativa del rol: %w", err)
	}
	if !permisos.PermiteMovimientoFinancieroManual(tipoMovimiento) {
		return fmt.Errorf("el rol cajero no tiene habilitado registrar %s manuales; activalo en Configuracion > Impresoras y caja > Configuracion operativa de cobro > Override por rol", strings.ToLower(strings.TrimSpace(tipoMovimiento)))
	}
	return nil
}

func validarCupoCajasLicencia(dbEmp *sql.DB, dbSuper *sql.DB, empresaID int64, excludeCierreID int64) (int, int, error) {
	if empresaID <= 0 {
		return 0, 0, fmt.Errorf("empresa_id es obligatorio")
	}
	maxCajas := 0
	if cfg, err := dbpkg.GetEmpresaConfiguracionGeneral(dbEmp, empresaID); err == nil && cfg != nil {
		if !cfg.CajaActiva {
			return maxCajas, 0, fmt.Errorf("la caja esta desactivada en la configuracion de la empresa")
		}
		if !cfg.CajasSimultaneasHabilitadas {
			maxCajas = 1
		} else if cfg.MaxCajasSimultaneasEmpresa > 0 {
			maxCajas = int(cfg.MaxCajasSimultaneasEmpresa)
		}
	} else if err != nil {
		return 0, 0, err
	}
	abiertas, err := dbpkg.CountEmpresaCierresCajaAbiertosExcepto(dbEmp, empresaID, excludeCierreID)
	if err != nil {
		return maxCajas, abiertas, err
	}
	if maxCajas > 0 && abiertas >= maxCajas {
		return maxCajas, abiertas, fmt.Errorf("la configuracion actual permite maximo %d caja(s) abiertas simultaneamente; cierre una caja antes de abrir otra o ajuste la configuracion de la empresa", maxCajas)
	}
	return maxCajas, abiertas, nil
}

// EmpresaFinanzasMovimientosHandler gestiona CRUD de ingresos/egresos por empresa.
func EmpresaFinanzasMovimientosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "conciliacion_bancaria" || action == "desviaciones_financieras" || action == "desviaciones_periodo" || action == "conciliacion_bancaria_export" || action == "desviaciones_financieras_export" {
				limit, err := parseIntQueryOptional(r, "limit")
				if err != nil {
					http.Error(w, "limit invalido", http.StatusBadRequest)
					return
				}
				resumen, err := dbpkg.GetEmpresaConciliacionBancariaPorPeriodo(dbEmp, empresaID, dbpkg.EmpresaConciliacionBancariaFilter{
					Desde:           strings.TrimSpace(r.URL.Query().Get("desde")),
					Hasta:           strings.TrimSpace(r.URL.Query().Get("hasta")),
					PeriodoContable: strings.TrimSpace(r.URL.Query().Get("periodo")),
					IncludeInactive: queryBool(r, "include_inactive"),
					Limit:           limit,
				})
				if err != nil {
					http.Error(w, "No se pudo construir la conciliacion bancaria", http.StatusInternalServerError)
					return
				}

				if action == "conciliacion_bancaria_export" || action == "desviaciones_financieras_export" {
					format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
					if format == "" {
						format = "json"
					}
					if format == "json" {
						fileName := "conciliacion_bancaria_empresa_" + strconv.FormatInt(empresaID, 10) + "_" + time.Now().Format("20060102_150405") + ".json"
						w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
						writeJSON(w, http.StatusOK, resumen)
						return
					}
					ds := buildEmpresaConciliacionBancariaDataset(resumen)
					if err := writeReportesDatasetExport(w, ds, format); err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
					return
				}

				writeJSON(w, http.StatusOK, resumen)
				return
			}
			if action == "extractos_bancarios" || action == "movimientos_bancarios" {
				limit, err := parseIntQueryOptional(r, "limit")
				if err != nil {
					http.Error(w, "limit invalido", http.StatusBadRequest)
					return
				}
				rows, err := dbpkg.ListEmpresaFinanzasMovimientosBancarios(dbEmp, empresaID, dbpkg.EmpresaFinanzasMovimientoBancarioFilter{
					Desde:              strings.TrimSpace(r.URL.Query().Get("desde")),
					Hasta:              strings.TrimSpace(r.URL.Query().Get("hasta")),
					PeriodoContable:    strings.TrimSpace(r.URL.Query().Get("periodo")),
					EstadoConciliacion: strings.TrimSpace(r.URL.Query().Get("estado_conciliacion")),
					IncludeInactive:    queryBool(r, "include_inactive"),
					Limit:              limit,
				})
				if err != nil {
					http.Error(w, "No se pudieron listar los extractos bancarios", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			}
			if action == "tablero_export" || action == "tablero_exportar" || action == "export_tablero" {
				desde := strings.TrimSpace(r.URL.Query().Get("desde"))
				hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
				resumen, err := dbpkg.GetEmpresaReportesTableroResumen(dbEmp, empresaID, desde, hasta)
				if err != nil {
					http.Error(w, "No se pudo construir el tablero de reportes", http.StatusInternalServerError)
					return
				}

				format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
				if format == "" {
					format = "json"
				}
				fileNameBase := "tablero_empresa_" + strconv.FormatInt(empresaID, 10) + "_" + time.Now().Format("20060102_150405")

				switch format {
				case "json":
					w.Header().Set("Content-Disposition", "attachment; filename=\""+fileNameBase+".json\"")
					writeJSON(w, http.StatusOK, buildTableroResumenExportPayload(resumen))
					return
				case "csv":
					content, err := buildTableroResumenCSVContent(resumen)
					if err != nil {
						http.Error(w, "No se pudo generar la exportacion CSV del tablero", http.StatusInternalServerError)
						return
					}
					w.Header().Set("Content-Type", "text/csv; charset=utf-8")
					w.Header().Set("Content-Disposition", "attachment; filename=\""+fileNameBase+".csv\"")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(content))
					return
				default:
					http.Error(w, "format invalido (use csv o json)", http.StatusBadRequest)
					return
				}
			}
			if action == "tablero" || action == "dashboard" || action == "resumen_kpi" {
				desde := strings.TrimSpace(r.URL.Query().Get("desde"))
				hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
				resumen, err := dbpkg.GetEmpresaReportesTableroResumen(dbEmp, empresaID, desde, hasta)
				if err != nil {
					http.Error(w, "No se pudo construir el tablero de reportes", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, resumen)
				return
			}
			includeInactive := queryBool(r, "include_inactive")
			tipo := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("tipo")))
			desde := strings.TrimSpace(r.URL.Query().Get("desde"))
			hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
			periodo := strings.TrimSpace(r.URL.Query().Get("periodo"))
			q := strings.TrimSpace(r.URL.Query().Get("q"))
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
			rows, err := dbpkg.ListEmpresaFinanzasMovimientos(dbEmp, empresaID, dbpkg.EmpresaFinanzasMovimientoFilter{
				Tipo:            tipo,
				Desde:           desde,
				Hasta:           hasta,
				Periodo:         periodo,
				Q:               q,
				CierreCajaID:    cierreCajaID,
				CajaCodigo:      strings.TrimSpace(r.URL.Query().Get("caja_codigo")),
				IncludeInactive: includeInactive,
				Limit:           limit,
			})
			if err != nil {
				http.Error(w, "No se pudieron listar los movimientos financieros", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, rows)
			return

		case http.MethodPost:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "importar_extractos_bancarios" || action == "importar_bancario" || action == "conciliacion_bancaria_importar" {
				var payload empresaFinanzasImportacionBancariaPayload
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
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
				if len(payload.Movimientos) == 0 {
					http.Error(w, "movimientos es obligatorio", http.StatusBadRequest)
					return
				}

				usuarioOperacion := strings.TrimSpace(adminEmailFromRequest(r))
				if usuarioOperacion == "" {
					usuarioOperacion = "sistema"
				}
				for i := range payload.Movimientos {
					if payload.Movimientos[i].EmpresaID <= 0 {
						payload.Movimientos[i].EmpresaID = payload.EmpresaID
					}
					if strings.TrimSpace(payload.Movimientos[i].CuentaBancaria) == "" {
						payload.Movimientos[i].CuentaBancaria = strings.TrimSpace(payload.CuentaBancaria)
					}
					if strings.TrimSpace(payload.Movimientos[i].BancoNombre) == "" {
						payload.Movimientos[i].BancoNombre = strings.TrimSpace(payload.BancoNombre)
					}
					if strings.TrimSpace(payload.Movimientos[i].Origen) == "" {
						payload.Movimientos[i].Origen = strings.TrimSpace(payload.Origen)
					}
					payload.Movimientos[i].UsuarioCreador = usuarioOperacion
				}

				importacion, err := dbpkg.UpsertEmpresaFinanzasMovimientosBancarios(dbEmp, payload.EmpresaID, payload.Movimientos)
				if err != nil {
					http.Error(w, "No se pudieron importar los extractos bancarios", http.StatusBadRequest)
					return
				}

				response := map[string]interface{}{
					"ok":          true,
					"importacion": importacion,
				}

				autoConciliar := payload.AutoConciliar || queryBool(r, "auto_conciliar")
				if autoConciliar {
					limit := payload.Limit
					if qLimit, qErr := parseIntQueryOptional(r, "limit"); qErr == nil && qLimit > 0 {
						limit = qLimit
					} else if qErr != nil {
						http.Error(w, "limit invalido", http.StatusBadRequest)
						return
					}

					toleranciaDias := payload.ToleranciaDias
					if raw := strings.TrimSpace(r.URL.Query().Get("tolerancia_dias")); raw != "" {
						v, err := strconv.Atoi(raw)
						if err != nil {
							http.Error(w, "tolerancia_dias invalido", http.StatusBadRequest)
							return
						}
						toleranciaDias = v
					}

					toleranciaMonto := payload.ToleranciaMonto
					if raw := strings.TrimSpace(r.URL.Query().Get("tolerancia_monto")); raw != "" {
						v, err := strconv.ParseFloat(raw, 64)
						if err != nil {
							http.Error(w, "tolerancia_monto invalido", http.StatusBadRequest)
							return
						}
						toleranciaMonto = v
					}

					resultado, err := dbpkg.ConciliarEmpresaMovimientosBancariosAutomatico(dbEmp, payload.EmpresaID, dbpkg.EmpresaConciliacionBancariaAutoConfig{
						Desde:           strings.TrimSpace(r.URL.Query().Get("desde")),
						Hasta:           strings.TrimSpace(r.URL.Query().Get("hasta")),
						PeriodoContable: strings.TrimSpace(r.URL.Query().Get("periodo")),
						ToleranciaDias:  toleranciaDias,
						ToleranciaMonto: toleranciaMonto,
						Limit:           limit,
						Usuario:         usuarioOperacion,
					})
					if err != nil {
						http.Error(w, "Importacion realizada, pero la conciliacion automatica fallo", http.StatusInternalServerError)
						return
					}
					response["conciliacion_automatica"] = resultado
				}

				writeJSON(w, http.StatusCreated, response)
				return
			}

			var payload dbpkg.EmpresaFinanzasMovimiento
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 {
				if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
					payload.EmpresaID = empresaID
				}
			}
			if err := validarPermisoRolMovimientoFinanzasManual(dbEmp, r, payload.EmpresaID, payload.TipoMovimiento); err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			if err := normalizarCajaMovimientoFinanzas(dbEmp, &payload, payload.UsuarioCreador); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			id, err := dbpkg.CreateEmpresaFinanzasMovimiento(dbEmp, payload)
			if err != nil {
				if errors.Is(err, dbpkg.ErrPeriodoFinancieroCerrado) {
					http.Error(w, "el periodo contable del movimiento esta cerrado", http.StatusConflict)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			evento := "movimiento_ingreso_registrado"
			if strings.EqualFold(strings.TrimSpace(payload.TipoMovimiento), "egreso") {
				evento = "movimiento_egreso_registrado"
			}
			montoEvento := payload.Total
			if montoEvento <= 0 {
				montoEvento = payload.Monto
			}
			registrarEventoContableNoBloqueante(dbEmp, r, "finanzas", dbpkg.EmpresaEventoContable{
				EmpresaID:       payload.EmpresaID,
				Modulo:          "finanzas",
				Evento:          evento,
				Entidad:         "finanzas_movimiento",
				EntidadID:       id,
				DocumentoTipo:   strings.TrimSpace(payload.TipoComprobante),
				DocumentoCodigo: strings.TrimSpace(payload.Codigo),
				PeriodoContable: strings.TrimSpace(payload.PeriodoContable),
				MontoTotal:      montoEvento,
				Moneda:          strings.TrimSpace(payload.Moneda),
				Origen:          "api_finanzas_movimientos",
				Observaciones:   "movimiento financiero registrado desde API",
			}, map[string]interface{}{
				"tipo_movimiento":  strings.ToLower(strings.TrimSpace(payload.TipoMovimiento)),
				"concepto":         strings.TrimSpace(payload.Concepto),
				"categoria":        strings.TrimSpace(payload.Categoria),
				"subcategoria":     strings.TrimSpace(payload.Subcategoria),
				"periodo_contable": strings.TrimSpace(payload.PeriodoContable),
				"metodo_pago":      strings.TrimSpace(payload.MetodoPago),
				"subtotal":         firstPositiveFloat64(payload.Monto, payload.Total),
				"base_gravable":    firstPositiveFloat64(payload.Monto, payload.Total),
				"iva":              payload.Impuesto,
				"impuestos":        payload.Impuesto,
				"retencion_fuente": payload.RetencionFuente,
				"retencion_ica":    payload.RetencionICA,
				"retencion_iva":    payload.RetencionIVA,
				"total_retenciones": firstPositiveFloat64(
					payload.TotalRetenciones,
					payload.RetencionFuente+payload.RetencionICA+payload.RetencionIVA,
				),
				"total_neto":         payload.TotalNeto,
				"tercero_nombre":     strings.TrimSpace(payload.TerceroNombre),
				"tercero_documento":  strings.TrimSpace(payload.TerceroDocumento),
				"referencia_externa": strings.TrimSpace(payload.ReferenciaExterna),
				"empresa_id":         payload.EmpresaID,
				"cierre_caja_id":     payload.CierreCajaID,
				"caja_codigo":        strings.TrimSpace(payload.CajaCodigo),
			})
			if payload.CierreCajaID > 0 && strings.EqualFold(strings.TrimSpace(payload.MetodoPago), "efectivo") {
				montoCaja := payload.TotalNeto
				if montoCaja <= 0 {
					montoCaja = payload.Total
				}
				if montoCaja <= 0 {
					montoCaja = payload.Monto
				}
				if err := dbpkg.RegistrarMovimientoEfectivoCierreCaja(dbEmp, payload.EmpresaID, payload.CierreCajaID, payload.TipoMovimiento, montoCaja); err != nil {
					http.Error(w, "movimiento registrado, pero no se pudo actualizar la caja abierta", http.StatusInternalServerError)
					return
				}
			}
			writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			return

		case http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "conciliar_bancaria_auto" || action == "conciliar_bancos" || action == "conciliar_bancaria_automatica" {
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

				toleranciaDias, err := parseIntQueryOptional(r, "tolerancia_dias")
				if err != nil {
					http.Error(w, "tolerancia_dias invalido", http.StatusBadRequest)
					return
				}

				toleranciaMonto, err := parseFloat64QueryOptional(r, "tolerancia_monto")
				if err != nil {
					http.Error(w, "tolerancia_monto invalido", http.StatusBadRequest)
					return
				}

				resultado, err := dbpkg.ConciliarEmpresaMovimientosBancariosAutomatico(dbEmp, empresaID, dbpkg.EmpresaConciliacionBancariaAutoConfig{
					Desde:           strings.TrimSpace(r.URL.Query().Get("desde")),
					Hasta:           strings.TrimSpace(r.URL.Query().Get("hasta")),
					PeriodoContable: strings.TrimSpace(r.URL.Query().Get("periodo")),
					ToleranciaDias:  toleranciaDias,
					ToleranciaMonto: toleranciaMonto,
					Limit:           limit,
					Usuario:         strings.TrimSpace(adminEmailFromRequest(r)),
				})
				if err != nil {
					http.Error(w, "No se pudo ejecutar la conciliacion bancaria automatica", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, resultado)
				return
			}
			if action == "activar" || action == "desactivar" || action == "anular" {
				empresaID, err := parseEmpresaIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				id, err := parseInt64Query(r, "id")
				if err != nil {
					http.Error(w, "id es obligatorio", http.StatusBadRequest)
					return
				}
				estado := "activo"
				if action == "desactivar" {
					estado = "inactivo"
				}
				if action == "anular" {
					estado = "anulado"
				}
				tipoMovimiento, err := dbpkg.GetEmpresaFinanzasMovimientoTipo(dbEmp, empresaID, id)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "movimiento no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, "No se pudo validar el tipo de movimiento", http.StatusInternalServerError)
					return
				}
				if err := validarPermisoRolMovimientoFinanzasManual(dbEmp, r, empresaID, tipoMovimiento); err != nil {
					http.Error(w, err.Error(), http.StatusForbidden)
					return
				}
				if err := dbpkg.SetEmpresaFinanzasMovimientoEstado(dbEmp, empresaID, id, estado); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "movimiento no encontrado", http.StatusNotFound)
						return
					}
					if errors.Is(err, dbpkg.ErrPeriodoFinancieroCerrado) {
						http.Error(w, "el periodo contable del movimiento esta cerrado", http.StatusConflict)
						return
					}
					http.Error(w, "No se pudo actualizar el estado", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": estado})
				return
			}

			var payload dbpkg.EmpresaFinanzasMovimiento
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 || payload.ID <= 0 {
				http.Error(w, "id y empresa_id son obligatorios", http.StatusBadRequest)
				return
			}
			if err := validarPermisoRolMovimientoFinanzasManual(dbEmp, r, payload.EmpresaID, payload.TipoMovimiento); err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
			if payload.UsuarioCreador == "" {
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			}
			if err := normalizarCajaMovimientoFinanzas(dbEmp, &payload, payload.UsuarioCreador); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := dbpkg.UpdateEmpresaFinanzasMovimiento(dbEmp, payload); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "movimiento no encontrado", http.StatusNotFound)
					return
				}
				if errors.Is(err, dbpkg.ErrPeriodoFinancieroCerrado) {
					http.Error(w, "el periodo contable del movimiento esta cerrado", http.StatusConflict)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return

		case http.MethodDelete:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			id, err := parseInt64Query(r, "id")
			if err != nil {
				http.Error(w, "id es obligatorio", http.StatusBadRequest)
				return
			}
			if err := dbpkg.DeleteEmpresaFinanzasMovimiento(dbEmp, empresaID, id); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "movimiento no encontrado", http.StatusNotFound)
					return
				}
				if errors.Is(err, dbpkg.ErrPeriodoFinancieroCerrado) {
					http.Error(w, "el periodo contable del movimiento esta cerrado", http.StatusConflict)
					return
				}
				http.Error(w, "No se pudo eliminar el movimiento", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

// EmpresaFinanzasMovimientoComprobanteUploadHandler carga un comprobante físico por empresa para un movimiento financiero.
func EmpresaFinanzasMovimientoComprobanteUploadHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseMultipartForm(20 << 20); err != nil {
			http.Error(w, "payload multipart invalido", http.StatusBadRequest)
			return
		}

		empresaID, err := parseInt64Form(r, "empresa_id")
		if err != nil || empresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}
		movimientoID, err := parseInt64Form(r, "movimiento_id")
		if err != nil || movimientoID <= 0 {
			http.Error(w, "movimiento_id es obligatorio", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("archivo")
		if err != nil {
			file, header, err = r.FormFile("comprobante")
		}
		if err != nil {
			http.Error(w, "archivo es obligatorio", http.StatusBadRequest)
			return
		}
		defer file.Close()

		fileURL, fileName, absPath, err := saveEmpresaComprobanteUpload(file, header.Filename, empresaID, "finanzas", fmt.Sprintf("movimiento_%d", movimientoID))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := dbpkg.UpdateEmpresaFinanzasMovimientoComprobante(dbEmp, empresaID, movimientoID, fileURL); err != nil {
			_ = os.Remove(absPath)
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "movimiento no encontrado", http.StatusNotFound)
				return
			}
			http.Error(w, "No se pudo guardar el comprobante", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusCreated, map[string]interface{}{
			"ok":                         true,
			"empresa_id":                 empresaID,
			"movimiento_id":              movimientoID,
			"comprobante_url":            fileURL,
			"comprobante_nombre_archivo": fileName,
		})
	}
}

// EmpresaFinanzasConfiguracionHandler gestiona configuracion por empresa del modulo financiero.
func EmpresaFinanzasConfiguracionHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			cfg, err := dbpkg.GetEmpresaFinanzasConfiguracion(dbEmp, empresaID)
			if err != nil {
				http.Error(w, "No se pudo consultar la configuracion financiera", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, cfg)
			return

		case http.MethodPost, http.MethodPut:
			var payload dbpkg.EmpresaFinanzasConfiguracion
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
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
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			id, err := dbpkg.UpsertEmpresaFinanzasConfiguracion(dbEmp, payload)
			if err != nil {
				http.Error(w, "No se pudo guardar la configuracion financiera", http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

// EmpresaFinanzasPeriodosHandler gestiona periodos contables por empresa.
func EmpresaFinanzasPeriodosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			includeInactive := queryBool(r, "include_inactive")
			rows, err := dbpkg.ListEmpresaFinanzasPeriodos(dbEmp, empresaID, includeInactive)
			if err != nil {
				http.Error(w, "No se pudieron listar los periodos", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, rows)
			return

		case http.MethodPost:
			var payload dbpkg.EmpresaFinanzasPeriodo
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
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
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			id, err := dbpkg.UpsertEmpresaFinanzasPeriodo(dbEmp, payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
			return

		case http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "cerrar" || action == "reabrir" {
				empresaID, err := parseEmpresaIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				periodo := strings.TrimSpace(r.URL.Query().Get("periodo"))
				if periodo == "" {
					http.Error(w, "periodo es obligatorio", http.StatusBadRequest)
					return
				}
				estado := "cerrado"
				if action == "reabrir" {
					estado = "abierto"
				}
				ejecutadoPor := strings.TrimSpace(adminEmailFromRequest(r))
				autorizacion, err := parseEmpresaFinanzasPeriodoAutorizacionPayload(r, ejecutadoPor)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if strings.TrimSpace(ejecutadoPor) == "" {
					ejecutadoPor = autorizacion.AutorizadoPor
				}
				observaciones := buildEmpresaFinanzasPeriodoAutorizacionObservaciones(action, autorizacion, ejecutadoPor)

				if err := dbpkg.SetEmpresaFinanzasPeriodoEstado(dbEmp, empresaID, periodo, estado, ejecutadoPor, observaciones); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				evento := "periodo_contable_cerrado"
				if estado == "abierto" {
					evento = "periodo_contable_reabierto"
				}
				registrarEventoContableNoBloqueante(dbEmp, r, "finanzas", dbpkg.EmpresaEventoContable{
					EmpresaID:       empresaID,
					Modulo:          "finanzas",
					Evento:          evento,
					Entidad:         "finanzas_periodo",
					DocumentoTipo:   "periodo_contable",
					DocumentoCodigo: periodo,
					PeriodoContable: periodo,
					Origen:          "api_finanzas_periodos",
					Observaciones:   "actualizacion de estado de periodo contable con evidencia de autorizacion",
				}, map[string]interface{}{
					"periodo":                periodo,
					"estado":                 estado,
					"empresa_id":             empresaID,
					"policy_autorizacion":    true,
					"autorizado_por":         autorizacion.AutorizadoPor,
					"motivo_autorizacion":    autorizacion.MotivoAutorizacion,
					"evidencia_autorizacion": autorizacion.EvidenciaAutorizacion,
					"codigo_autorizacion":    autorizacion.CodigoAutorizacion,
					"ejecutado_por":          ejecutadoPor,
				})
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":      true,
					"periodo": periodo,
					"estado":  estado,
					"autorizacion": map[string]interface{}{
						"autorizado_por":         autorizacion.AutorizadoPor,
						"motivo_autorizacion":    autorizacion.MotivoAutorizacion,
						"evidencia_autorizacion": autorizacion.EvidenciaAutorizacion,
						"codigo_autorizacion":    autorizacion.CodigoAutorizacion,
						"ejecutado_por":          ejecutadoPor,
					},
				})
				return
			}

			var payload dbpkg.EmpresaFinanzasPeriodo
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
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
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			id, err := dbpkg.UpsertEmpresaFinanzasPeriodo(dbEmp, payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

// EmpresaFinanzasCierresCajaHandler gestiona apertura/arqueo/cierre de caja por empresa/sucursal.
func EmpresaFinanzasCierresCajaHandler(dbEmp *sql.DB, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			sucursalID, err := parseInt64QueryOptional(r, "sucursal_id")
			if err != nil {
				http.Error(w, "sucursal_id invalido", http.StatusBadRequest)
				return
			}
			limit, err := parseIntQueryOptional(r, "limit")
			if err != nil {
				http.Error(w, "limit invalido", http.StatusBadRequest)
				return
			}
			rows, err := dbpkg.ListEmpresaCierresCaja(dbEmp, empresaID, dbpkg.EmpresaCierreCajaFilter{
				SucursalID:      sucursalID,
				CajaCodigo:      strings.TrimSpace(r.URL.Query().Get("caja_codigo")),
				EstadoCierre:    strings.TrimSpace(r.URL.Query().Get("estado_cierre")),
				Desde:           strings.TrimSpace(r.URL.Query().Get("desde")),
				Hasta:           strings.TrimSpace(r.URL.Query().Get("hasta")),
				IncludeInactive: queryBool(r, "include_inactive"),
				Limit:           limit,
			})
			if err != nil {
				http.Error(w, "No se pudieron listar los cierres de caja", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, rows)
			return

		case http.MethodPost:
			var payload dbpkg.EmpresaCierreCaja
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 {
				if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
					payload.EmpresaID = empresaID
				}
			}
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			if payload.UsuarioCreador == "" {
				payload.UsuarioCreador = "sistema"
			}
			if strings.TrimSpace(payload.EstadoCierre) == "" || strings.EqualFold(strings.TrimSpace(payload.EstadoCierre), "abierto") {
				if strings.TrimSpace(payload.CajaCodigo) != "" {
					if existing, err := dbpkg.GetEmpresaCierreCajaAbiertaUsuario(dbEmp, payload.EmpresaID, 0, payload.CajaCodigo, payload.Turno, payload.SucursalID, payload.UsuarioCreador); err == nil && existing != nil {
						writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": existing.ID, "existente": true})
						return
					} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "No se pudo validar la caja abierta del usuario", http.StatusInternalServerError)
						return
					}
				}
				if _, _, err := validarCupoCajasLicencia(dbEmp, dbSuper, payload.EmpresaID, 0); err != nil {
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}
			}
			id, err := dbpkg.CreateEmpresaCierreCaja(dbEmp, payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			return

		case http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "cerrar" || action == "reabrir" || action == "aprobar" || action == "anular" || action == "activar" || action == "desactivar" {
				empresaID, err := parseEmpresaIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				id, err := parseInt64Query(r, "id")
				if err != nil {
					http.Error(w, "id es obligatorio", http.StatusBadRequest)
					return
				}

				if action == "activar" || action == "desactivar" {
					estado := "activo"
					if action == "desactivar" {
						estado = "inactivo"
					}
					if err := dbpkg.SetEmpresaCierreCajaRegistroEstado(dbEmp, empresaID, id, estado); err != nil {
						if errors.Is(err, sql.ErrNoRows) {
							http.Error(w, "cierre de caja no encontrado", http.StatusNotFound)
							return
						}
						http.Error(w, "No se pudo actualizar el estado del registro", http.StatusInternalServerError)
						return
					}
					writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": estado})
					return
				}

				estadoCierre := "cerrado"
				switch action {
				case "reabrir":
					estadoCierre = "abierto"
				case "aprobar":
					estadoCierre = "aprobado"
				case "anular":
					estadoCierre = "anulado"
				}
				if estadoCierre == "abierto" {
					if _, _, err := validarCupoCajasLicencia(dbEmp, dbSuper, empresaID, id); err != nil {
						http.Error(w, err.Error(), http.StatusConflict)
						return
					}
				}

				var cajaFisica *float64
				if raw := strings.TrimSpace(r.URL.Query().Get("caja_fisica")); raw != "" {
					v, err := strconv.ParseFloat(raw, 64)
					if err != nil {
						http.Error(w, "caja_fisica invalida", http.StatusBadRequest)
						return
					}
					if v < 0 {
						v = 0
					}
					cajaFisica = &v
				}

				usuarioOperacion := strings.TrimSpace(adminEmailFromRequest(r))

				if err := dbpkg.SetEmpresaCierreCajaEstado(
					dbEmp,
					empresaID,
					id,
					estadoCierre,
					cajaFisica,
					usuarioOperacion,
					strings.TrimSpace(r.URL.Query().Get("observaciones")),
				); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "cierre de caja no encontrado", http.StatusNotFound)
						return
					}
					if errors.Is(err, dbpkg.ErrCierreCajaTransicionInvalida) || errors.Is(err, dbpkg.ErrCierreCajaAprobadoBloqueado) {
						http.Error(w, err.Error(), http.StatusConflict)
						return
					}
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				resp := map[string]interface{}{"ok": true, "estado_cierre": estadoCierre}
				if estadoCierre == "cerrado" || estadoCierre == "aprobado" {
					conciliacion, err := dbpkg.ConciliarEmpresaPropinasConCierreCaja(dbEmp, empresaID, id, usuarioOperacion)
					if err != nil {
						http.Error(w, "No se pudo conciliar propinas para el cierre de caja", http.StatusInternalServerError)
						return
					}
					resp["conciliacion_propinas"] = conciliacion
				}
				writeJSON(w, http.StatusOK, resp)
				return
			}

			var payload dbpkg.EmpresaCierreCaja
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 || payload.ID <= 0 {
				http.Error(w, "id y empresa_id son obligatorios", http.StatusBadRequest)
				return
			}
			if strings.EqualFold(strings.TrimSpace(payload.EstadoCierre), "abierto") {
				if _, _, err := validarCupoCajasLicencia(dbEmp, dbSuper, payload.EmpresaID, payload.ID); err != nil {
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}
			}
			if payload.UsuarioCreador == "" {
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			}
			if err := dbpkg.UpdateEmpresaCierreCaja(dbEmp, payload); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "cierre de caja no encontrado", http.StatusNotFound)
					return
				}
				if errors.Is(err, dbpkg.ErrCierreCajaAprobadoBloqueado) {
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return

		case http.MethodDelete:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			id, err := parseInt64Query(r, "id")
			if err != nil {
				http.Error(w, "id es obligatorio", http.StatusBadRequest)
				return
			}
			if err := dbpkg.DeleteEmpresaCierreCaja(dbEmp, empresaID, id); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "cierre de caja no encontrado", http.StatusNotFound)
					return
				}
				if errors.Is(err, dbpkg.ErrCierreCajaAprobadoBloqueado) {
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}
				http.Error(w, "No se pudo eliminar el cierre de caja", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

// EmpresaFinanzasAsientosContablesHandler procesa eventos contables pendientes y consulta asientos canónicos.
func EmpresaFinanzasAsientosContablesHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			limit, err := parseIntQueryOptional(r, "limit")
			if err != nil {
				http.Error(w, "limit invalido", http.StatusBadRequest)
				return
			}
			if action == "conciliacion_periodo" || action == "conciliacion" || action == "conciliar" {
				resumen, err := dbpkg.GetEmpresaConciliacionContablePorPeriodo(dbEmp, empresaID, dbpkg.EmpresaConciliacionContableFilter{
					Desde:           strings.TrimSpace(r.URL.Query().Get("desde")),
					Hasta:           strings.TrimSpace(r.URL.Query().Get("hasta")),
					PeriodoContable: strings.TrimSpace(r.URL.Query().Get("periodo")),
					IncludeInactive: queryBool(r, "include_inactive"),
					Limit:           limit,
				})
				if err != nil {
					http.Error(w, "No se pudo construir la conciliacion contable", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, resumen)
				return
			}
			rows, err := dbpkg.ListEmpresaAsientosContables(dbEmp, empresaID, dbpkg.EmpresaAsientoContableFilter{
				Modulo:          strings.TrimSpace(r.URL.Query().Get("modulo")),
				Evento:          strings.TrimSpace(r.URL.Query().Get("evento")),
				PeriodoContable: strings.TrimSpace(r.URL.Query().Get("periodo")),
				Desde:           strings.TrimSpace(r.URL.Query().Get("desde")),
				Hasta:           strings.TrimSpace(r.URL.Query().Get("hasta")),
				IncludeInactive: queryBool(r, "include_inactive"),
				Limit:           limit,
			})
			if err != nil {
				http.Error(w, "No se pudieron listar los asientos contables", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, rows)
			return

		case http.MethodPut, http.MethodPost:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "" {
				action = "procesar_asientos"
			}
			if action != "procesar_asientos" && action != "procesar" {
				http.Error(w, "action invalida", http.StatusBadRequest)
				return
			}

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
			maxRetries, err := parseIntQueryOptional(r, "max_reintentos")
			if err != nil {
				http.Error(w, "max_reintentos invalido", http.StatusBadRequest)
				return
			}

			resultado, err := dbpkg.ProcessEmpresaEventosContablesPendientesConPolitica(dbEmp, empresaID, strings.TrimSpace(adminEmailFromRequest(r)), limit, maxRetries)
			if err != nil {
				http.Error(w, "No se pudieron procesar los eventos contables pendientes", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, resultado)
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}
