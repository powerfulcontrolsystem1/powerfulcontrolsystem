package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type agenteInternetFiscalProposal struct {
	Campo       string `json:"campo"`
	Actual      string `json:"actual"`
	Sugerido    string `json:"sugerido"`
	Fuente      string `json:"fuente"`
	RequiereOK  bool   `json:"requiere_confirmacion"`
	Observacion string `json:"observacion"`
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
			writeJSON(w, http.StatusTooManyRequests, map[string]any{"ok": false, "error": err.Error(), "usage": usage, "limits": limits})
			return
		}
		country := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("pais")))
		if country == "" {
			country = "CO"
		}
		proposals := buildAgenteInternetFiscalProposals(modulo, country)
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

func buildAgenteInternetFiscalProposals(modulo, country string) []agenteInternetFiscalProposal {
	if strings.EqualFold(modulo, "nomina") {
		return []agenteInternetFiscalProposal{
			{Campo: "pais_normativo", Actual: "configuracion actual de la empresa", Sugerido: country, Fuente: "agente_internet: fuentes oficiales pendientes de revision", RequiereOK: true, Observacion: "El agente prepara comparacion antes de actualizar parametros legales."},
			{Campo: "parametros_nomina", Actual: "valores guardados", Sugerido: "revisar salario minimo, auxilio, aportes, recargos y calendario laboral vigente", Fuente: "Ministerio de Trabajo/DIAN/UGPP segun pais", RequiereOK: true, Observacion: "No se modifica nada sin confirmacion del usuario."},
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
