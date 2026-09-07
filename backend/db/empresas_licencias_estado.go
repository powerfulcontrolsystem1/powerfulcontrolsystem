package db

import (
	"database/sql"
	"strings"
)

// EmpresaLicenciaEstadoFilter controla la lectura global de empresas por estado de licencia.
type EmpresaLicenciaEstadoFilter struct {
	Query          string
	LicenciaFiltro string
	Limit          int
}

// EmpresaLicenciaEstadoSummary resume el estado de licenciamiento de las empresas.
type EmpresaLicenciaEstadoSummary struct {
	Total              int64 `json:"total"`
	ConLicenciaActiva  int64 `json:"con_licencia_activa"`
	SinLicenciaActiva  int64 `json:"sin_licencia_activa"`
	ConLicencia15Dias  int64 `json:"con_licencia_15_dias"`
	ConLicenciaVencida int64 `json:"con_licencia_vencida"`
}

// EmpresaLicenciaEstado expone datos de empresa con el ultimo/actual estado de licencia.
type EmpresaLicenciaEstado struct {
	Empresa
	LicenciaEstado      string `json:"licencia_estado"`
	LicenciaActiva      bool   `json:"licencia_activa"`
	Licencia15Dias      bool   `json:"licencia_15_dias"`
	LicenciaVencida     bool   `json:"licencia_vencida"`
	LicenciaID          int64  `json:"licencia_id,omitempty"`
	LicenciaNombre      string `json:"licencia_nombre,omitempty"`
	LicenciaFechaInicio string `json:"licencia_fecha_inicio,omitempty"`
	LicenciaFechaFin    string `json:"licencia_fecha_fin,omitempty"`
	LicenciaEstadoRaw   string `json:"licencia_estado_raw,omitempty"`
	LicenciasActivas    int64  `json:"licencias_activas"`
	LicenciasVencidas   int64  `json:"licencias_vencidas"`
}

type empresaLicenciaResumen struct {
	selected licenciaEmpresaResumenRow
	hasRow   bool
	active   int64
	expired  int64
	has15    bool
}

type licenciaEmpresaResumenRow struct {
	id          int64
	empresaID   int64
	nombre      string
	duracion    int64
	fechaInicio string
	fechaFin    string
	estado      string
	activa      bool
	vencida     bool
}

// ListEmpresasLicenciaEstado lista empresas con un resumen read-only de licencias.
func ListEmpresasLicenciaEstado(dbEmp, dbSuper *sql.DB, filter EmpresaLicenciaEstadoFilter) ([]EmpresaLicenciaEstado, EmpresaLicenciaEstadoSummary, error) {
	empresas, err := GetEmpresas(dbEmp)
	if err != nil {
		return nil, EmpresaLicenciaEstadoSummary{}, err
	}
	licencias, err := queryLicenciasResumenPorEmpresa(dbSuper)
	if err != nil {
		return nil, EmpresaLicenciaEstadoSummary{}, err
	}

	filter.Query = strings.ToLower(strings.TrimSpace(filter.Query))
	filter.LicenciaFiltro = normalizeEmpresaLicenciaFiltro(filter.LicenciaFiltro)
	if filter.Limit <= 0 {
		filter.Limit = 500
	}
	if filter.Limit > 2000 {
		filter.Limit = 2000
	}

	var summary EmpresaLicenciaEstadoSummary
	items := make([]EmpresaLicenciaEstado, 0, len(empresas))
	for _, empresa := range empresas {
		item := buildEmpresaLicenciaEstado(empresa, licencias[empresaKey(empresa)])
		summary.Total++
		if item.LicenciaActiva {
			summary.ConLicenciaActiva++
		} else {
			summary.SinLicenciaActiva++
		}
		if item.Licencia15Dias {
			summary.ConLicencia15Dias++
		}
		if item.LicenciaVencida {
			summary.ConLicenciaVencida++
		}
		if !empresaLicenciaMatchesFilter(item, filter) {
			continue
		}
		if len(items) < filter.Limit {
			items = append(items, item)
		}
	}
	return items, summary, nil
}

func queryLicenciasResumenPorEmpresa(dbConn *sql.DB) (map[int64]*empresaLicenciaResumen, error) {
	out := make(map[int64]*empresaLicenciaResumen)
	if dbConn == nil {
		return out, nil
	}
	activePredicate := licenciaActivePredicate("")
	expiredPredicate := licenciaExpiredPredicate("")
	rows, err := querySQLCompat(dbConn, `
		SELECT id, COALESCE(empresa_id, 0), COALESCE(nombre, ''), COALESCE(duracion_dias, 0),
		       COALESCE(CAST(fecha_inicio AS TEXT), ''), COALESCE(CAST(fecha_fin AS TEXT), ''), COALESCE(estado, ''),
		       CASE WHEN `+activePredicate+` THEN 1 ELSE 0 END,
		       CASE WHEN `+expiredPredicate+` THEN 1 ELSE 0 END
		FROM licencias
		WHERE COALESCE(empresa_id, 0) > 0
		ORDER BY COALESCE(empresa_id, 0) ASC,
		         CASE WHEN `+activePredicate+` THEN 0 ELSE 1 END,
		         COALESCE(CAST(fecha_fin AS TEXT), '') DESC,
		         id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var row licenciaEmpresaResumenRow
		var activa, vencida int
		if err := rows.Scan(&row.id, &row.empresaID, &row.nombre, &row.duracion, &row.fechaInicio, &row.fechaFin, &row.estado, &activa, &vencida); err != nil {
			return nil, err
		}
		if row.empresaID <= 0 {
			continue
		}
		row.activa = activa == 1
		row.vencida = vencida == 1
		resume := out[row.empresaID]
		if resume == nil {
			resume = &empresaLicenciaResumen{}
			out[row.empresaID] = resume
		}
		if row.activa {
			resume.active++
		}
		if row.vencida {
			resume.expired++
		}
		if row.duracion == 15 {
			resume.has15 = true
		}
		if !resume.hasRow || (row.activa && !resume.selected.activa) {
			resume.selected = row
			resume.hasRow = true
		}
	}
	return out, rows.Err()
}

func buildEmpresaLicenciaEstado(empresa Empresa, resumen *empresaLicenciaResumen) EmpresaLicenciaEstado {
	item := EmpresaLicenciaEstado{Empresa: empresa, LicenciaEstado: "sin_licencia"}
	if resumen == nil || !resumen.hasRow {
		return item
	}
	row := resumen.selected
	item.LicenciaID = row.id
	item.LicenciaNombre = row.nombre
	item.LicenciaFechaInicio = row.fechaInicio
	item.LicenciaFechaFin = row.fechaFin
	item.LicenciaEstadoRaw = row.estado
	item.LicenciasActivas = resumen.active
	item.LicenciasVencidas = resumen.expired
	item.LicenciaActiva = resumen.active > 0
	item.Licencia15Dias = resumen.has15
	item.LicenciaVencida = resumen.active == 0 && resumen.expired > 0
	if item.LicenciaActiva {
		item.LicenciaEstado = "activa"
	} else if item.LicenciaVencida {
		item.LicenciaEstado = "vencida"
	}
	return item
}

func empresaKey(empresa Empresa) int64 {
	if empresa.EmpresaID > 0 {
		return empresa.EmpresaID
	}
	return empresa.ID
}

func normalizeEmpresaLicenciaFiltro(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "-", "_")
	switch value {
	case "activa", "licencia_activa", "activas":
		return "activa"
	case "sin_activa", "sin_licencia_activa", "sin_licencia", "inactiva":
		return "sin_activa"
	case "15", "15_dias", "licencia_15", "licencia_15_dias", "prueba_15":
		return "15_dias"
	case "vencida", "licencia_vencida", "vencidas":
		return "vencida"
	default:
		return ""
	}
}

func empresaLicenciaMatchesFilter(item EmpresaLicenciaEstado, filter EmpresaLicenciaEstadoFilter) bool {
	if filter.Query != "" {
		haystack := strings.ToLower(strings.Join([]string{
			item.Nombre,
			item.Nit,
			item.TipoNombre,
			item.UsuarioCreador,
			item.LicenciaNombre,
		}, " "))
		if !strings.Contains(haystack, filter.Query) {
			return false
		}
	}
	switch filter.LicenciaFiltro {
	case "activa":
		return item.LicenciaActiva
	case "sin_activa":
		return !item.LicenciaActiva
	case "15_dias":
		return item.Licencia15Dias
	case "vencida":
		return item.LicenciaVencida
	default:
		return true
	}
}
