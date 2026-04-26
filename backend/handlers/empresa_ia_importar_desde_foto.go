package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaIAImportarDesdeFotoHandler aplica altas masivas sugeridas por IA (precios/productos o egresos),
// pero SOLO para administradores autenticados (Google) dentro del panel empresa.
//
// Se ejecuta bajo WithEmpresaSeguridadPermissions y además exige adminEmailFromRequest != "".
func EmpresaIAImportarDesdeFotoHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		adminEmail := strings.TrimSpace(adminEmailFromRequest(r))
		if adminEmail == "" {
			http.Error(w, "solo administrador puede ejecutar esta accion", http.StatusForbidden)
			return
		}

		var payload struct {
			EmpresaID int64 `json:"empresa_id"`
			Productos []struct {
				Nombre             string  `json:"nombre"`
				Precio             float64 `json:"precio"`
				SKU                string  `json:"sku"`
				CodigoBarras       string  `json:"codigo_barras"`
				Descripcion        string  `json:"descripcion"`
				CategoriaID        int64   `json:"categoria_id"`
				ImpuestoPorcentaje float64 `json:"impuesto_porcentaje"`
			} `json:"productos"`
			Egresos []struct {
				FechaMovimiento string  `json:"fecha_movimiento"`
				PeriodoContable string  `json:"periodo_contable"`
				Categoria       string  `json:"categoria"`
				Concepto        string  `json:"concepto"`
				Descripcion     string  `json:"descripcion"`
				MetodoPago      string  `json:"metodo_pago"`
				Moneda          string  `json:"moneda"`
				Total           float64 `json:"total"`
			} `json:"egresos"`
			Note string `json:"note"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "JSON invalido", http.StatusBadRequest)
			return
		}
		if payload.EmpresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}
		if len(payload.Productos) == 0 && len(payload.Egresos) == 0 {
			http.Error(w, "productos o egresos es obligatorio", http.StatusBadRequest)
			return
		}

		if len(payload.Productos) > 250 || len(payload.Egresos) > 250 {
			http.Error(w, "demasiados items (max 250 por tipo)", http.StatusBadRequest)
			return
		}

		createdProductos := make([]int64, 0, len(payload.Productos))
		for _, p := range payload.Productos {
			nombre := strings.TrimSpace(p.Nombre)
			if nombre == "" {
				http.Error(w, "producto.nombre es obligatorio", http.StatusBadRequest)
				return
			}
			if p.Precio < 0 {
				http.Error(w, "producto.precio invalido", http.StatusBadRequest)
				return
			}
			prod := dbpkg.Producto{
				EmpresaID:          payload.EmpresaID,
				CategoriaID:        p.CategoriaID,
				SKU:                strings.TrimSpace(p.SKU),
				CodigoBarras:       strings.TrimSpace(p.CodigoBarras),
				Nombre:             nombre,
				Descripcion:        strings.TrimSpace(p.Descripcion),
				Costo:              0,
				Precio:             p.Precio,
				ImpuestoPorcentaje: p.ImpuestoPorcentaje,
				Estado:             "activo",
				UsuarioCreador:     adminEmail,
				Observaciones:      "importado_desde_foto_por_ia",
			}
			id, err := dbpkg.CreateProducto(dbEmp, prod, 0, "")
			if err != nil {
				http.Error(w, "no se pudo crear producto: "+err.Error(), http.StatusBadRequest)
				return
			}
			createdProductos = append(createdProductos, id)
		}

		createdEgresos := make([]int64, 0, len(payload.Egresos))
		for _, e := range payload.Egresos {
			concepto := strings.TrimSpace(e.Concepto)
			if concepto == "" {
				http.Error(w, "egreso.concepto es obligatorio", http.StatusBadRequest)
				return
			}
			total := e.Total
			if total <= 0 {
				http.Error(w, "egreso.total es obligatorio y debe ser > 0", http.StatusBadRequest)
				return
			}
			fecha := strings.TrimSpace(e.FechaMovimiento)
			if fecha == "" {
				fecha = time.Now().Format("2006-01-02")
			}
			mov := dbpkg.EmpresaFinanzasMovimiento{
				EmpresaID:       payload.EmpresaID,
				TipoMovimiento:  "egreso",
				FechaMovimiento: fecha,
				PeriodoContable: strings.TrimSpace(e.PeriodoContable),
				Categoria:       strings.TrimSpace(e.Categoria),
				Concepto:        concepto,
				Descripcion:     strings.TrimSpace(e.Descripcion),
				MetodoPago:      strings.TrimSpace(e.MetodoPago),
				Moneda:          strings.TrimSpace(e.Moneda),
				Total:           total,
				UsuarioCreador:  adminEmail,
				Estado:          "activo",
				Observaciones:   "importado_desde_foto_por_ia",
			}
			id, err := dbpkg.CreateEmpresaFinanzasMovimiento(dbEmp, mov)
			if err != nil {
				if err == dbpkg.ErrPeriodoFinancieroCerrado {
					http.Error(w, "el periodo contable del movimiento esta cerrado", http.StatusConflict)
					return
				}
				http.Error(w, "no se pudo crear egreso: "+err.Error(), http.StatusBadRequest)
				return
			}
			createdEgresos = append(createdEgresos, id)
		}

		writeJSON(w, http.StatusCreated, map[string]any{
			"ok":         true,
			"empresa_id": payload.EmpresaID,
			"productos_creados": map[string]any{
				"count": len(createdProductos),
				"ids":   createdProductos,
			},
			"egresos_creados": map[string]any{
				"count": len(createdEgresos),
				"ids":   createdEgresos,
			},
			"note":   strings.TrimSpace(payload.Note),
			"source": "ia_importar_desde_foto",
			"admin":  adminEmail,
			"ts":     strconv.FormatInt(time.Now().Unix(), 10),
		})
	}
}
