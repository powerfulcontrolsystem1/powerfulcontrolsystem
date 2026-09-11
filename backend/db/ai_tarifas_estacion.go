package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

// EmpresaAITarifaMinutosPlan es el contrato cerrado que el chat IA puede proponer.
// Empresa, usuario y confirmacion siempre se derivan de la sesion autenticada.
type EmpresaAITarifaMinutosPlan struct {
	EstacionID          int64                         `json:"estacion_id"`
	NombreEstacion      string                        `json:"nombre_estacion"`
	ConservarExistentes bool                          `json:"conservar_existentes"`
	Tarifas             []EmpresaAITarifaMinutosRegla `json:"tarifas"`
}

type EmpresaAITarifaMinutosRegla struct {
	DiaSemanaDesde    int     `json:"dia_semana_desde"`
	DiaSemanaHasta    int     `json:"dia_semana_hasta"`
	MinutosBase       int     `json:"minutos_base"`
	ValorBase         float64 `json:"valor_base"`
	MinutosExtra      int     `json:"minutos_extra"`
	ValorExtra        float64 `json:"valor_extra"`
	CobrarPorFraccion bool    `json:"cobrar_por_fraccion"`
	Prioridad         int     `json:"prioridad"`
}

func NormalizeEmpresaAITarifaMinutosPlan(plan *EmpresaAITarifaMinutosPlan) error {
	if plan == nil || plan.EstacionID <= 0 {
		return fmt.Errorf("estacion_id es obligatorio")
	}
	plan.NombreEstacion = strings.TrimSpace(plan.NombreEstacion)
	if len([]rune(plan.NombreEstacion)) > 120 {
		return fmt.Errorf("nombre de estacion invalido")
	}
	if len(plan.Tarifas) == 0 || len(plan.Tarifas) > 14 {
		return fmt.Errorf("debe indicar entre una y catorce tarifas")
	}
	seen := map[string]bool{}
	coveredDays := map[int]bool{}
	for i := range plan.Tarifas {
		rate := &plan.Tarifas[i]
		if rate.DiaSemanaDesde < 1 || rate.DiaSemanaDesde > 7 || rate.DiaSemanaHasta < 1 || rate.DiaSemanaHasta > 7 {
			return fmt.Errorf("rango de dias invalido")
		}
		if rate.MinutosBase <= 0 || rate.MinutosBase > 43200 || rate.ValorBase <= 0 || rate.ValorBase > 1000000000 {
			return fmt.Errorf("tarifa base invalida")
		}
		if rate.MinutosExtra < 0 || rate.MinutosExtra > 43200 || rate.ValorExtra < 0 || rate.ValorExtra > 1000000000 {
			return fmt.Errorf("tarifa extra invalida")
		}
		if (rate.MinutosExtra == 0) != (rate.ValorExtra == 0) {
			return fmt.Errorf("minutos y valor extra deben configurarse juntos")
		}
		if rate.Prioridad <= 0 {
			rate.Prioridad = 1
		}
		key := fmt.Sprintf("%d:%d", rate.DiaSemanaDesde, rate.DiaSemanaHasta)
		if seen[key] {
			return fmt.Errorf("rango de dias duplicado")
		}
		seen[key] = true
		for day := 1; day <= 7; day++ {
			if !diaSemanaInRange(day, rate.DiaSemanaDesde, rate.DiaSemanaHasta) {
				continue
			}
			if coveredDays[day] {
				return fmt.Errorf("los rangos de dias no pueden superponerse")
			}
			coveredDays[day] = true
		}
	}
	return nil
}

// ConfigureEmpresaAITarifasMinutosStation aplica estacion y reglas en una transaccion.
func ConfigureEmpresaAITarifasMinutosStation(dbConn *sql.DB, empresaID int64, plan EmpresaAITarifaMinutosPlan, usuario string) ([]int64, error) {
	usuario = strings.TrimSpace(usuario)
	if empresaID <= 0 || usuario == "" {
		return nil, fmt.Errorf("contexto empresarial invalido")
	}
	if err := NormalizeEmpresaAITarifaMinutosPlan(&plan); err != nil {
		return nil, err
	}
	if err := EmpresaEstacionPrefsSchemaReady(dbConn); err != nil {
		return nil, err
	}
	if err := EmpresaTarifasPorMinutosSchemaReady(dbConn); err != nil {
		return nil, err
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Bloquea la preferencia mientras se combina el cambio propuesto. Sin este
	// lock, una edicion manual concurrente podria perderse al confirmar el chat.
	var rawConfig string
	if err := queryRowTxSQLCompat(tx, `SELECT COALESCE(valor, '')
		FROM empresa_estacion_prefs
		WHERE empresa_id=? AND estacion_id=0 AND clave='estaciones_config'
		FOR UPDATE`, empresaID).Scan(&rawConfig); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no existe configuracion de estaciones para la empresa")
		}
		return nil, err
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(rawConfig), &cfg); err != nil {
		return nil, fmt.Errorf("configuracion de estaciones invalida")
	}
	items, ok := cfg["estaciones"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no se encontro la estacion solicitada")
	}
	found := false
	stationName := plan.NombreEstacion
	for _, raw := range items {
		station, valid := raw.(map[string]interface{})
		if !valid {
			continue
		}
		id, _ := station["id"].(float64)
		if int64(id) != plan.EstacionID {
			continue
		}
		found = true
		if stationName == "" {
			stationName = strings.TrimSpace(fmt.Sprint(station["nombre"]))
		}
		if stationName != "" {
			station["nombre"] = stationName
		}
		station["tipo_operacion"] = "motel"
		break
	}
	if !found {
		return nil, fmt.Errorf("la estacion no pertenece a la empresa")
	}
	cfgJSON, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	if _, err := execTxSQLCompat(tx, `UPDATE empresa_estacion_prefs SET valor=?, usuario_creador=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND estacion_id=0 AND clave='estaciones_config'`, string(cfgJSON), usuario, empresaID); err != nil {
		return nil, err
	}
	if !plan.ConservarExistentes {
		if _, err := execTxSQLCompat(tx, `UPDATE empresa_tarifas_por_minutos SET estado='inactivo', fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND estacion_id=? AND COALESCE(estado,'activo')='activo'`, empresaID, plan.EstacionID); err != nil {
			return nil, err
		}
	}
	ids := make([]int64, 0, len(plan.Tarifas))
	stationCode := fmt.Sprintf("EST-%d-%d", empresaID, plan.EstacionID)
	for _, rate := range plan.Tarifas {
		id, err := insertTxSQLCompat(tx, `INSERT INTO empresa_tarifas_por_minutos (empresa_id,estacion_id,estacion_codigo,estacion_nombre,dia_semana_desde,dia_semana_hasta,minutos_base,valor_base,minutos_extra,valor_extra,cobrar_por_fraccion,moneda,prioridad,usuario_creador,estado,observaciones,fecha_creacion,fecha_actualizacion) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?, ?,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP) ON CONFLICT (empresa_id,estacion_id,dia_semana_desde,dia_semana_hasta) DO UPDATE SET estacion_codigo=EXCLUDED.estacion_codigo,estacion_nombre=EXCLUDED.estacion_nombre,minutos_base=EXCLUDED.minutos_base,valor_base=EXCLUDED.valor_base,minutos_extra=EXCLUDED.minutos_extra,valor_extra=EXCLUDED.valor_extra,cobrar_por_fraccion=EXCLUDED.cobrar_por_fraccion,moneda=EXCLUDED.moneda,prioridad=EXCLUDED.prioridad,usuario_creador=EXCLUDED.usuario_creador,estado='activo',observaciones=EXCLUDED.observaciones,fecha_actualizacion=CURRENT_TIMESTAMP`, empresaID, plan.EstacionID, stationCode, stationName, rate.DiaSemanaDesde, rate.DiaSemanaHasta, rate.MinutosBase, round2(rate.ValorBase), rate.MinutosExtra, round2(rate.ValorExtra), boolToInt(rate.CobrarPorFraccion), "COP", rate.Prioridad, usuario, "activo", "Configurada mediante propuesta IA confirmada")
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return ids, nil
}
