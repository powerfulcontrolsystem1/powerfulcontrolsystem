package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	ColombiaDefaultsVersion              = "CO-2026-06"
	ColombiaSalarioMinimoMensual2026     = 1750905
	ColombiaAuxilioTransporteMensual2026 = 249095
	colombiaDefaultsMigrationVersion     = "20260607_colombia_impuestos_nomina_defaults"
	colombiaDefaultsPrefClave            = "preconfiguracion_colombia_fiscal_nomina"
	colombiaDefaultsMigrationDescription = "Preconfigura impuestos Colombia y parametros base de nomina en empresas existentes"
	colombiaDefaultsUsuarioSistema       = "sistema.preconfiguracion_colombia"
	colombiaDefaultsObservacionPreprod   = "[preproduccion_2026-06-07] defaults Colombia impuestos y nomina; revisar con contador antes de produccion"
)

type EmpresaColombiaDefaultsResult struct {
	EmpresaID             int64    `json:"empresa_id"`
	Version               string   `json:"version"`
	Impuestos             int      `json:"impuestos"`
	NominaConfiguracionID int64    `json:"nomina_configuracion_id"`
	ConceptosNomina       int      `json:"conceptos_nomina"`
	MarkerID              int64    `json:"marker_id"`
	Errores               []string `json:"errores,omitempty"`
}

type EmpresasColombiaDefaultsBackfillResult struct {
	Version   string                          `json:"version"`
	Empresas  int                             `json:"empresas"`
	Aplicadas int                             `json:"aplicadas"`
	Errores   []string                        `json:"errores,omitempty"`
	Items     []EmpresaColombiaDefaultsResult `json:"items,omitempty"`
}

func EmpresaImpuestosCatalogoBase(pais string) []EmpresaImpuestoConfig {
	pais = strings.ToUpper(strings.TrimSpace(pais))
	switch pais {
	case "EC":
		return []EmpresaImpuestoConfig{
			{PaisCodigo: "EC", Codigo: "IVA", Nombre: "IVA tarifa general", Tipo: "impuesto", TasaPorcentaje: 15, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "EC", Codigo: "IVA_0", Nombre: "IVA 0% / exento", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "EC", Codigo: "ICE", Nombre: "ICE consumos especiales", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "EC", Codigo: "RET_IVA", Nombre: "Retencion IVA segun SRI", Tipo: "retencion", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "compras"},
			{PaisCodigo: "EC", Codigo: "RET_IR", Nombre: "Retencion IR segun tabla SRI", Tipo: "retencion", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "compras"},
		}
	case "PA":
		return []EmpresaImpuestoConfig{
			{PaisCodigo: "PA", Codigo: "ITBMS_7", Nombre: "ITBMS 7% tasa general", Tipo: "impuesto", TasaPorcentaje: 7, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "PA", Codigo: "ITBMS_10", Nombre: "ITBMS 10% rubros especiales", Tipo: "impuesto", TasaPorcentaje: 10, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "PA", Codigo: "ITBMS_15", Nombre: "ITBMS 15% rubros especiales", Tipo: "impuesto", TasaPorcentaje: 15, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "PA", Codigo: "ISC", Nombre: "ISC selectivo al consumo", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "PA", Codigo: "RET_ITBMS", Nombre: "Retencion ITBMS segun condicion", Tipo: "retencion", TasaPorcentaje: 50, Habilitado: 0, AplicaEn: "compras"},
		}
	case "CR":
		return []EmpresaImpuestoConfig{
			{PaisCodigo: "CR", Codigo: "IVA_13", Nombre: "IVA 13% tarifa general", Tipo: "impuesto", TasaPorcentaje: 13, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "CR", Codigo: "IVA_4", Nombre: "IVA 4% tarifa reducida", Tipo: "impuesto", TasaPorcentaje: 4, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "CR", Codigo: "IVA_2", Nombre: "IVA 2% tarifa reducida", Tipo: "impuesto", TasaPorcentaje: 2, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "CR", Codigo: "IVA_1", Nombre: "IVA 1% tarifa reducida", Tipo: "impuesto", TasaPorcentaje: 1, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "CR", Codigo: "EXENTO", Nombre: "Exento / no sujeto", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 1, AplicaEn: "ventas"},
		}
	case "AR":
		return []EmpresaImpuestoConfig{
			{PaisCodigo: "AR", Codigo: "IVA_21", Nombre: "IVA 21% tarifa general", Tipo: "impuesto", TasaPorcentaje: 21, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "AR", Codigo: "IVA_105", Nombre: "IVA 10.5% tarifa reducida", Tipo: "impuesto", TasaPorcentaje: 10.5, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "AR", Codigo: "IVA_27", Nombre: "IVA 27% tarifa diferencial", Tipo: "impuesto", TasaPorcentaje: 27, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "AR", Codigo: "EXENTO", Nombre: "Exento / no gravado", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "AR", Codigo: "RET_GAN", Nombre: "Retencion ganancias segun regimen", Tipo: "retencion", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "compras"},
			{PaisCodigo: "AR", Codigo: "IIBB", Nombre: "Ingresos brutos jurisdiccional", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "ventas"},
		}
	case "VE":
		return []EmpresaImpuestoConfig{
			{PaisCodigo: "VE", Codigo: "IVA_16", Nombre: "IVA 16% tarifa general", Tipo: "impuesto", TasaPorcentaje: 16, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "VE", Codigo: "IVA_8", Nombre: "IVA 8% tarifa reducida", Tipo: "impuesto", TasaPorcentaje: 8, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "VE", Codigo: "IVA_31", Nombre: "IVA adicional 31% rubros especiales", Tipo: "impuesto", TasaPorcentaje: 31, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "VE", Codigo: "EXENTO", Nombre: "Exento / no sujeto", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "VE", Codigo: "IGTF", Nombre: "IGTF segun medio de pago", Tipo: "impuesto", TasaPorcentaje: 3, Habilitado: 0, AplicaEn: "ventas"},
		}
	default:
		return EmpresaImpuestosCatalogoColombia()
	}
}

func EmpresaImpuestosCatalogoColombia() []EmpresaImpuestoConfig {
	return []EmpresaImpuestoConfig{
		{PaisCodigo: "CO", Codigo: "IVA", Nombre: "IVA tarifa general 19%", Tipo: "impuesto", TasaPorcentaje: 19, Habilitado: 1, AplicaEn: "ventas"},
		{PaisCodigo: "CO", Codigo: "IVA_0", Nombre: "IVA 0% / exento / excluido", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 1, AplicaEn: "ventas"},
		{PaisCodigo: "CO", Codigo: "INC_8", Nombre: "Impuesto nacional al consumo 8%", Tipo: "impuesto", TasaPorcentaje: 8, Habilitado: 0, AplicaEn: "ventas"},
		{PaisCodigo: "CO", Codigo: "ICA", Nombre: "ICA municipal variable", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "ventas"},
		{PaisCodigo: "CO", Codigo: "RETEFUENTE", Nombre: "Retencion en la fuente renta", Tipo: "retencion", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "compras"},
		{PaisCodigo: "CO", Codigo: "RETEIVA", Nombre: "Retencion a titulo de IVA 15%", Tipo: "retencion", TasaPorcentaje: 15, Habilitado: 0, AplicaEn: "compras"},
		{PaisCodigo: "CO", Codigo: "RETEICA", Nombre: "Retencion a titulo de ICA", Tipo: "retencion", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "compras"},
	}
}

func ApplyEmpresaColombiaDefaults(dbConn *sql.DB, empresaID int64, usuario string) (EmpresaColombiaDefaultsResult, error) {
	res := EmpresaColombiaDefaultsResult{EmpresaID: empresaID, Version: ColombiaDefaultsVersion}
	if dbConn == nil {
		return res, fmt.Errorf("db nil")
	}
	if empresaID <= 0 {
		return res, fmt.Errorf("empresa_id es obligatorio")
	}
	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = colombiaDefaultsUsuarioSistema
	}
	if err := EnsureEmpresaImpuestosSchema(dbConn); err != nil {
		return res, err
	}
	if err := EnsureEmpresaNominaSchema(dbConn); err != nil {
		return res, err
	}
	if err := EnsureEmpresaEstacionPrefsSchema(dbConn); err != nil {
		return res, err
	}

	for _, imp := range EmpresaImpuestosCatalogoColombia() {
		imp.EmpresaID = empresaID
		imp.UsuarioCreador = usuario
		imp.Estado = "activo"
		imp.Observaciones = colombiaDefaultsObservacionPreprod + " version=" + ColombiaDefaultsVersion
		if _, err := UpsertEmpresaImpuesto(dbConn, imp); err != nil {
			res.Errores = append(res.Errores, fmt.Sprintf("impuesto %s: %v", imp.Codigo, err))
			continue
		}
		res.Impuestos++
	}

	cfg := defaultEmpresaNominaConfiguracion(empresaID)
	cfg.UsuarioCreador = usuario
	cfg.Observaciones = colombiaDefaultsObservacionPreprod + " version=" + ColombiaDefaultsVersion
	id, err := UpsertEmpresaNominaConfiguracion(dbConn, cfg)
	if err != nil {
		res.Errores = append(res.Errores, "nomina_configuracion: "+err.Error())
	} else {
		res.NominaConfiguracionID = id
	}

	if err := SeedEmpresaNominaColombiaConceptosBase(dbConn, empresaID, usuario); err != nil {
		res.Errores = append(res.Errores, "nomina_conceptos: "+err.Error())
	} else {
		res.ConceptosNomina = len(nominaColombiaConceptosProfesionales(empresaID, usuario))
	}

	markerRaw, _ := json.Marshal(map[string]interface{}{
		"version":                    ColombiaDefaultsVersion,
		"salario_minimo_mensual":     ColombiaSalarioMinimoMensual2026,
		"auxilio_transporte_mensual": ColombiaAuxilioTransporteMensual2026,
		"impuestos":                  res.Impuestos,
		"conceptos_nomina":           res.ConceptosNomina,
		"observacion":                "preconfiguracion Colombia aplicada en preproduccion",
	})
	markerID, err := UpsertEmpresaEstacionPref(dbConn, EmpresaEstacionPref{
		EmpresaID:      empresaID,
		EstacionID:     0,
		Clave:          colombiaDefaultsPrefClave,
		Valor:          string(markerRaw),
		UsuarioCreador: usuario,
		Estado:         "activo",
		Observaciones:  colombiaDefaultsObservacionPreprod,
	})
	if err != nil {
		res.Errores = append(res.Errores, "marker: "+err.Error())
	} else {
		res.MarkerID = markerID
	}

	if len(res.Errores) > 0 {
		return res, fmt.Errorf("%s", strings.Join(res.Errores, "; "))
	}
	return res, nil
}

func ApplyColombiaDefaultsToExistingEmpresas(dbConn *sql.DB) (*EmpresasColombiaDefaultsBackfillResult, error) {
	res := &EmpresasColombiaDefaultsBackfillResult{Version: ColombiaDefaultsVersion}
	if dbConn == nil {
		return res, fmt.Errorf("db nil")
	}
	err := ApplySchemaMigration(dbConn, "empresas", colombiaDefaultsMigrationVersion, colombiaDefaultsMigrationDescription, func(tx *sql.DB) error {
		empresas, err := GetEmpresas(tx)
		if err != nil {
			return err
		}
		res.Empresas = len(empresas)
		for _, empresa := range empresas {
			empresaID := empresa.EmpresaID
			if empresaID <= 0 {
				empresaID = empresa.ID
			}
			estado := strings.ToLower(strings.TrimSpace(empresa.Estado))
			if empresaID <= 0 || estado == "eliminada" || estado == "eliminado" {
				continue
			}
			item, err := ApplyEmpresaColombiaDefaults(tx, empresaID, colombiaDefaultsUsuarioSistema)
			res.Items = append(res.Items, item)
			if err != nil {
				res.Errores = append(res.Errores, fmt.Sprintf("empresa_id=%d: %v", empresaID, err))
				continue
			}
			res.Aplicadas++
		}
		if len(res.Errores) > 0 {
			return fmt.Errorf("%s", strings.Join(res.Errores, "; "))
		}
		return nil
	})
	return res, err
}
