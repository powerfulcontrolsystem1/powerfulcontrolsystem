package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaActivosFijosNIIFiscalHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "" {
			action = "dashboard"
		}
		usuario := strings.TrimSpace(adminEmailFromRequest(r))

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "dashboard":
				writeActivosFijosDashboard(w, dbEmp, empresaID, strings.TrimSpace(r.URL.Query().Get("periodo")))
				return
			case "activos", "libro":
				rows, err := dbpkg.ListEmpresaActivosFijos(dbEmp, empresaID, strings.TrimSpace(r.URL.Query().Get("estado")))
				if err != nil {
					http.Error(w, "No se pudieron listar activos fijos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "depreciaciones":
				rows, err := dbpkg.ListEmpresaActivosDepreciacion(dbEmp, empresaID, strings.TrimSpace(r.URL.Query().Get("periodo")), 1000)
				if err != nil {
					http.Error(w, "No se pudieron listar depreciaciones", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "eventos":
				rows, err := dbpkg.ListEmpresaActivosEventos(dbEmp, empresaID, int64Query(r, "activo_id"), 500)
				if err != nil {
					http.Error(w, "No se pudieron listar eventos de activos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			}
		case http.MethodPost, http.MethodPut:
			switch action {
			case "activo":
				var payload dbpkg.EmpresaActivoFijo
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.CreateEmpresaActivoFijo(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "depreciacion":
				var payload struct {
					Periodo string `json:"periodo"`
				}
				_ = json.NewDecoder(r.Body).Decode(&payload)
				periodo := strings.TrimSpace(payload.Periodo)
				if periodo == "" {
					periodo = strings.TrimSpace(r.URL.Query().Get("periodo"))
				}
				rows, err := dbpkg.GenerarEmpresaActivosDepreciacion(dbEmp, empresaID, periodo, usuario)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "depreciaciones": rows})
				return
			case "evento":
				var payload dbpkg.EmpresaActivoEvento
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.RegistrarEmpresaActivoEvento(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "seed_demo":
				if err := seedEmpresaActivosFijosDemo(dbEmp, empresaID, usuario); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			}
		}
		http.Error(w, "accion o metodo no soportado", http.StatusMethodNotAllowed)
	}
}

func writeActivosFijosDashboard(w http.ResponseWriter, dbEmp *sql.DB, empresaID int64, periodo string) {
	if periodo == "" {
		periodo = time.Now().Format("2006-01")
	}
	resumen, err := dbpkg.BuildEmpresaActivosFijosAvanzadoResumen(dbEmp, empresaID, periodo)
	if err != nil {
		http.Error(w, "No se pudo consultar resumen de activos fijos", http.StatusInternalServerError)
		return
	}
	activos, err := dbpkg.ListEmpresaActivosFijos(dbEmp, empresaID, "")
	if err != nil {
		http.Error(w, "No se pudo consultar libro de activos", http.StatusInternalServerError)
		return
	}
	deps, _ := dbpkg.ListEmpresaActivosDepreciacion(dbEmp, empresaID, periodo, 500)
	eventos, _ := dbpkg.ListEmpresaActivosEventos(dbEmp, empresaID, 0, 100)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"empresa_id":      empresaID,
		"periodo":         periodo,
		"resumen":         resumen,
		"activos":         activos,
		"depreciaciones":  deps,
		"eventos":         eventos,
		"alertas":         buildActivosFijosAlertas(activos),
		"por_categoria":   groupActivosFijosByCategoria(activos),
		"por_responsable": groupActivosFijosByResponsable(activos),
	})
}

func seedEmpresaActivosFijosDemo(dbEmp *sql.DB, empresaID int64, usuario string) error {
	existentes, err := dbpkg.ListEmpresaActivosFijos(dbEmp, empresaID, "")
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, item := range existentes {
		seen[strings.ToUpper(strings.TrimSpace(item.Codigo))] = true
	}
	rows := []dbpkg.EmpresaActivoFijo{
		{Codigo: "AF-NIIF-001", Nombre: "Servidor principal POS", Categoria: "equipo_computo", Serial: "SRV-PCS-2026", Placa: "PCS-0001", FechaCompra: "2026-01-10", Costo: 9800000, ValorResidual: 500000, VidaUtilMeses: 60, VidaUtilFiscalMeses: 36, MetodoDepreciacion: "linea_recta", MetodoDepreciacionFiscal: "linea_recta", CuentaActivo: "152805", CuentaDepreciacion: "159205", CuentaGasto: "516020", CuentaDeterioro: "519995", Ubicacion: "Centro de datos", Responsable: "Administrador TI", CentroCosto: "Administracion", Proveedor: "Proveedor tecnologico", ValorAsegurado: 9800000, Poliza: "POL-AF-001", MantenimientoCadaDias: 90},
		{Codigo: "AF-NIIF-002", Nombre: "Mobiliario recepcion", Categoria: "muebles_enseres", Serial: "MOB-REC", Placa: "PCS-0002", FechaCompra: "2026-02-05", Costo: 4200000, ValorResidual: 200000, VidaUtilMeses: 120, VidaUtilFiscalMeses: 120, CuentaActivo: "152405", CuentaDepreciacion: "159240", CuentaGasto: "516010", Ubicacion: "Recepcion", Responsable: "Jefe de sede", CentroCosto: "Ventas", Proveedor: "Muebles corporativos", ValorRazonable: 4100000},
		{Codigo: "AF-NIIF-003", Nombre: "Licencia ERP y seguridad", Categoria: "intangible", Serial: "LIC-ERP-2026", FechaCompra: "2026-03-01", Costo: 6500000, ValorResidual: 0, VidaUtilMeses: 36, VidaUtilFiscalMeses: 36, CuentaActivo: "163505", CuentaDepreciacion: "169805", CuentaGasto: "516025", Ubicacion: "Nube", Responsable: "Gerencia", CentroCosto: "Administracion", Proveedor: "Software corporativo"},
	}
	for _, row := range rows {
		if seen[strings.ToUpper(strings.TrimSpace(row.Codigo))] {
			continue
		}
		row.EmpresaID = empresaID
		row.UsuarioCreador = usuario
		if _, err := dbpkg.CreateEmpresaActivoFijo(dbEmp, row); err != nil {
			return err
		}
	}
	_, err = dbpkg.GenerarEmpresaActivosDepreciacion(dbEmp, empresaID, time.Now().Format("2006-01"), usuario)
	return err
}

func buildActivosFijosAlertas(rows []dbpkg.EmpresaActivoFijo) []string {
	alertas := []string{}
	today := time.Now().Format("2006-01-02")
	for _, row := range rows {
		if row.Estado == "activo" && strings.TrimSpace(row.ProximoMantenimiento) != "" && row.ProximoMantenimiento <= today {
			alertas = append(alertas, row.Codigo+" requiere mantenimiento")
		}
		if row.Estado == "activo" && row.DiferenciaNIIFFiscal != 0 {
			alertas = append(alertas, row.Codigo+" tiene diferencia NIIF/fiscal")
		}
		if row.Estado == "activo" && row.ValorAsegurado > 0 && row.ValorAsegurado < row.ValorLibros {
			alertas = append(alertas, row.Codigo+" tiene valor asegurado inferior al valor en libros")
		}
	}
	if len(alertas) == 0 {
		alertas = append(alertas, "Sin alertas criticas de activos fijos.")
	}
	return alertas
}

func groupActivosFijosByCategoria(rows []dbpkg.EmpresaActivoFijo) map[string]float64 {
	out := map[string]float64{}
	for _, row := range rows {
		if row.Estado != "activo" {
			continue
		}
		key := strings.TrimSpace(row.Categoria)
		if key == "" {
			key = "sin_categoria"
		}
		out[key] += row.ValorLibros
	}
	return out
}

func groupActivosFijosByResponsable(rows []dbpkg.EmpresaActivoFijo) map[string]int {
	out := map[string]int{}
	for _, row := range rows {
		if row.Estado != "activo" {
			continue
		}
		key := strings.TrimSpace(row.Responsable)
		if key == "" {
			key = "sin_responsable"
		}
		out[key]++
	}
	return out
}
