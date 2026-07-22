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

type agenteInternetFiscalProposal struct {
	Campo         string      `json:"campo"`
	CampoConfig   string      `json:"campo_config,omitempty"`
	Actual        string      `json:"actual"`
	Sugerido      string      `json:"sugerido"`
	ValorSugerido interface{} `json:"valor_sugerido,omitempty"`
	Fuente        string      `json:"fuente"`
	FuenteURL     string      `json:"fuente_url,omitempty"`
	Vigencia      string      `json:"vigencia,omitempty"`
	RequiereOK    bool        `json:"requiere_confirmacion"`
	Observacion   string      `json:"observacion"`
}

func EmpresaAgenteInternetNominaHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return empresaAgenteInternetFiscalHandler(dbEmp, dbSuper, "nomina")
}

func EmpresaAgenteInternetImpuestosHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return empresaAgenteInternetFiscalHandler(dbEmp, dbSuper, "impuestos")
}

func empresaAgenteInternetFiscalHandler(dbEmp, dbSuper *sql.DB, modulo string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodGet {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		empresaID, err := parseInt64QueryOptional(r, "empresa_id")
		if err != nil || empresaID <= 0 {
			http.Error(w, "empresa_id invalido", http.StatusBadRequest)
			return
		}
		usage, limits, err := reserveAgenteInternetLightUsage(dbEmp, dbSuper, empresaID, adminEmailFromRequest(r))
		if err != nil {
			writeAgenteInternetFiscalPublicError(w, r, http.StatusTooManyRequests, err, usage, limits)
			return
		}
		country := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("pais")))
		if country == "" {
			country = "CO"
		}
		proposals := buildAgenteInternetFiscalProposals(modulo, country, nil)
		if strings.EqualFold(modulo, "nomina") {
			cfg, cfgErr := dbpkg.GetEmpresaNominaConfiguracion(dbEmp, empresaID)
			if cfgErr != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "No se pudo consultar la configuracion actual de nomina"})
				return
			}
			proposals = buildAgenteInternetFiscalProposals(modulo, country, cfg)
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":          true,
			"modulo":      modulo,
			"pais":        country,
			"modelo":      "openai:gpt-5.4-mini",
			"agent":       "agente_internet",
			"proposals":   proposals,
			"usage":       usage,
			"limits":      limits,
			"aplica_auto": false,
			"message":     "Propuesta generada para revision humana. No se aplicaron cambios.",
		})
	}
}

func writeAgenteInternetFiscalPublicError(w http.ResponseWriter, r *http.Request, status int, err error, usage dbpkg.EmpresaAgenteUsoDiario, limits map[string]int64) {
	requestID := resolveAuditoriaRequestID(r)
	log.Printf("[agente_internet_fiscal] operation=reservar_cupo request_id=%s error_type=%T", requestID, err)
	payload := map[string]any{
		"ok":     false,
		"code":   "agente_internet_usage_unavailable",
		"error":  "No se pudo reservar el cupo del agente.",
		"usage":  usage,
		"limits": limits,
	}
	if requestID != "" {
		payload["request_id"] = requestID
	}
	writeJSON(w, status, payload)
}

func reserveAgenteInternetLightUsage(dbEmp, dbSuper *sql.DB, empresaID int64, user string) (dbpkg.EmpresaAgenteUsoDiario, map[string]int64, error) {
	limits := getAgentCompanyLimits(dbSuper)
	today := time.Now().Format("2006-01-02")
	usage, err := dbpkg.GetEmpresaAgenteUsoDiario(dbEmp, empresaID, today)
	if err != nil {
		return usage, limits, err
	}
	if usage.ConsultasLigeras >= limits["consultas_ligeras_diarias"] {
		return usage, limits, fmt.Errorf("limite diario de consultas ligeras del agente alcanzado")
	}
	if err := dbpkg.AddEmpresaAgenteUsoDiario(dbEmp, dbpkg.EmpresaAgenteUsoDiario{
		EmpresaID:        empresaID,
		FechaUso:         today,
		ConsultasLigeras: 1,
		SegundosUsados:   1,
		UsuarioCreador:   user,
	}); err != nil {
		return usage, limits, err
	}
	usage, _ = dbpkg.GetEmpresaAgenteUsoDiario(dbEmp, empresaID, today)
	return usage, limits, nil
}

func reserveEmpresaAgentAdvancedUsage(dbEmp, dbSuper *sql.DB, empresaID int64, user string) (dbpkg.EmpresaAgenteUsoDiario, map[string]int64, error) {
	limits := getAgentCompanyLimits(dbSuper)
	today := time.Now().Format("2006-01-02")
	usage, err := dbpkg.GetEmpresaAgenteUsoDiario(dbEmp, empresaID, today)
	if err != nil {
		return usage, limits, err
	}
	if usage.ConsultasAvanzadas >= limits["consultas_avanzadas_diarias"] {
		return usage, limits, fmt.Errorf("limite diario de consultas avanzadas del agente alcanzado")
	}
	if err := dbpkg.AddEmpresaAgenteUsoDiario(dbEmp, dbpkg.EmpresaAgenteUsoDiario{
		EmpresaID:          empresaID,
		FechaUso:           today,
		ConsultasAvanzadas: 1,
		SegundosUsados:     5,
		UsuarioCreador:     user,
	}); err != nil {
		return usage, limits, err
	}
	usage, _ = dbpkg.GetEmpresaAgenteUsoDiario(dbEmp, empresaID, today)
	return usage, limits, nil
}

func buildAgenteInternetFiscalProposals(modulo, country string, nominaCfg *dbpkg.EmpresaNominaConfiguracion) []agenteInternetFiscalProposal {
	if strings.EqualFold(modulo, "nomina") {
		if !strings.EqualFold(country, "CO") {
			return []agenteInternetFiscalProposal{{Campo: "pais_normativo", Actual: "configuracion actual de la empresa", Sugerido: country, Fuente: "Autoridad laboral del pais", RequiereOK: true, Observacion: "No hay un catalogo oficial verificado para este pais; no se aplican cambios."}}
		}
		cfg := dbpkg.EmpresaNominaConfiguracion{}
		if nominaCfg != nil {
			cfg = *nominaCfg
		}
		money := func(v float64) string { return fmt.Sprintf("%.2f", v) }
		return []agenteInternetFiscalProposal{
			{Campo: "Salario minimo mensual", CampoConfig: "salario_minimo_mensual", Actual: money(cfg.SalarioMinimoMensual), Sugerido: "1750905", ValorSugerido: 1750905, Fuente: "Decreto 1469 de 2025 - Presidencia", FuenteURL: "https://www.presidencia.gov.co/Documents/251230-Decreto-1469-MinTrabajo.pdf", Vigencia: "2026", RequiereOK: true, Observacion: "Valor oficial para Colombia durante 2026."},
			{Campo: "Auxilio de transporte mensual", CampoConfig: "auxilio_transporte_legal_mensual", Actual: money(cfg.AuxilioTransporteLegalMensual), Sugerido: "249095", ValorSugerido: 249095, Fuente: "Decreto 1470 de 2025 - Presidencia", FuenteURL: "https://www.presidencia.gov.co/Documents/251230-Decreto-1470-MinTrabajo.pdf", Vigencia: "2026", RequiereOK: true, Observacion: "Aplica a quienes cumplan los requisitos legales."},
			{Campo: "Horas ordinarias por semana", CampoConfig: "horas_ordinarias_semana", Actual: money(cfg.HorasOrdinariasSemana), Sugerido: "42", ValorSugerido: 42, Fuente: "Ley 2101 de 2021 y Ley 2466 de 2025", FuenteURL: "https://www.funcionpublica.gov.co/eva/gestornormativo/norma.php?i=260676", Vigencia: "Desde julio de 2026", RequiereOK: true, Observacion: "Jornada maxima ordinaria general; valida excepciones del sector."},
			{Campo: "Divisor de hora ordinaria", CampoConfig: "divisor_hora_ordinaria", Actual: money(cfg.DivisorHoraOrdinaria), Sugerido: "210", ValorSugerido: 210, Fuente: "Calculo derivado de jornada semanal de 42 horas", FuenteURL: "https://www.funcionpublica.gov.co/eva/gestornormativo/norma.php?i=166506", Vigencia: "Desde julio de 2026", RequiereOK: true, Observacion: "42 horas por 30 dias dividido entre 6 dias laborales."},
			{Campo: "Inicio de jornada nocturna", CampoConfig: "hora_nocturna_desde", Actual: cfg.HoraNocturnaDesde, Sugerido: "19:00:00", ValorSugerido: "19:00:00", Fuente: "Ley 2466 de 2025, articulo 10", FuenteURL: "https://www.funcionpublica.gov.co/eva/gestornormativo/norma.php?i=260676", Vigencia: "Desde diciembre de 2025", RequiereOK: true, Observacion: "La jornada nocturna general inicia a las 7:00 p. m."},
			{Campo: "Recargo por descanso obligatorio", CampoConfig: "recargo_dominical_diurno_porcentaje", Actual: money(cfg.RecargoDominicalDiurnoPorcentaje), Sugerido: "90", ValorSugerido: 90, Fuente: "Ley 2466 de 2025, articulo 14", FuenteURL: "https://www.funcionpublica.gov.co/eva/gestornormativo/norma.php?i=260676", Vigencia: "Desde 2026-07-01", RequiereOK: true, Observacion: "Incremento gradual vigente a la fecha de la consulta."},
			{Campo: "Recargo nocturno en descanso obligatorio", CampoConfig: "recargo_dominical_nocturno_porcentaje", Actual: money(cfg.RecargoDominicalNocturnoPorcentaje), Sugerido: "125", ValorSugerido: 125, Fuente: "Ley 2466 de 2025, articulos 10 y 14", FuenteURL: "https://www.funcionpublica.gov.co/eva/gestornormativo/norma.php?i=260676", Vigencia: "Desde 2026-07-01", RequiereOK: true, Observacion: "Suma del recargo por descanso obligatorio vigente y el nocturno."},
			{Campo: "Hora extra diurna en descanso obligatorio", CampoConfig: "hora_extra_dominical_diurna_porcentaje", Actual: money(cfg.HoraExtraDominicalDiurnaPorcentaje), Sugerido: "115", ValorSugerido: 115, Fuente: "Ley 2466 de 2025 y Codigo Sustantivo del Trabajo", FuenteURL: "https://www.funcionpublica.gov.co/eva/gestornormativo/norma.php?i=260676", Vigencia: "Desde 2026-07-01", RequiereOK: true, Observacion: "Suma del recargo de descanso obligatorio y hora extra diurna."},
			{Campo: "Hora extra nocturna en descanso obligatorio", CampoConfig: "hora_extra_dominical_nocturna_porcentaje", Actual: money(cfg.HoraExtraDominicalNocturnaPorcentaje), Sugerido: "165", ValorSugerido: 165, Fuente: "Ley 2466 de 2025 y Codigo Sustantivo del Trabajo", FuenteURL: "https://www.funcionpublica.gov.co/eva/gestornormativo/norma.php?i=260676", Vigencia: "Desde 2026-07-01", RequiereOK: true, Observacion: "Suma del recargo de descanso obligatorio y hora extra nocturna."},
		}
	}
	return []agenteInternetFiscalProposal{
		{Campo: "pais_fiscal", Actual: "configuracion actual de la empresa", Sugerido: country, Fuente: "agente_internet: fuentes oficiales pendientes de revision", RequiereOK: true, Observacion: "El agente prepara comparacion antes de cambiar impuestos."},
		{Campo: "catalogo_impuestos", Actual: "impuestos activos", Sugerido: "revisar IVA/retenciones/calendario tributario y tasas vigentes", Fuente: "DIAN o autoridad fiscal del pais detectado", RequiereOK: true, Observacion: "La propuesta queda para aprobacion humana."},
	}
}

func decodeAgenteInternetPayload(r *http.Request, target any) {
	if r == nil || r.Body == nil || target == nil {
		return
	}
	_ = json.NewDecoder(r.Body).Decode(target)
}
