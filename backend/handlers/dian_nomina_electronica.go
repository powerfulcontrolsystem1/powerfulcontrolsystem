package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/internal/platform/valueutil"
)

const (
	nominaElectronicaNamespace       = "dian:gov:co:facturaelectronica:NominaIndividual"
	nominaElectronicaAjusteNamespace = "dian:gov:co:facturaelectronica:NominaIndividualDeAjuste"
	nominaElectronicaVersion         = "V1.0: Documento Soporte de Pago de Nómina Electrónica"
	nominaElectronicaTipoXML         = "102"
)

func buildDIANNominaCUNE(numero, fecha, hora string, devengados, deducciones, comprobante interface{}, empleadorNIT, trabajadorDocumento, softwarePIN, ambiente string) string {
	return strings.ToLower(buildDIANSHA384Hex(
		strings.TrimSpace(numero),
		strings.TrimSpace(fecha),
		strings.TrimSpace(hora),
		dianNominaMoney(devengados),
		dianNominaMoney(deducciones),
		dianNominaMoney(comprobante),
		dianOnlyDigits(empleadorNIT),
		strings.TrimSpace(trabajadorDocumento),
		nominaElectronicaTipoXML,
		strings.TrimSpace(softwarePIN),
		strings.TrimSpace(ambiente),
	))
}

func dianNominaMoney(value interface{}) string {
	number := ventasAnyToFloat64(value)
	// El anexo de nomina exige truncar, no redondear, a dos decimales. El
	// epsilon solo neutraliza representaciones binarias inmediatamente por
	// debajo de un centavo que ya existe en la fuente NUMERIC(18,2).
	if number >= 0 {
		number = math.Trunc((number+1e-9)*100) / 100
	} else {
		number = math.Trunc((number-1e-9)*100) / 100
	}
	return fmt.Sprintf("%.2f", number)
}

func dianNominaBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func dianNominaOptionalAttr(name, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return " " + name + `="` + escapeXML(value) + `"`
}

func dianNominaParteAttrs(part dbpkg.EmpresaNominaDIANParte) string {
	var out strings.Builder
	out.WriteString(dianNominaOptionalAttr("RazonSocial", part.RazonSocial))
	out.WriteString(dianNominaOptionalAttr("PrimerApellido", part.PrimerApellido))
	out.WriteString(dianNominaOptionalAttr("SegundoApellido", part.SegundoApellido))
	out.WriteString(dianNominaOptionalAttr("PrimerNombre", part.PrimerNombre))
	out.WriteString(dianNominaOptionalAttr("OtrosNombres", part.OtrosNombres))
	out.WriteString(` NIT="` + escapeXML(dianOnlyDigits(part.NIT)) + `"`)
	out.WriteString(` DV="` + escapeXML(dianOnlyDigits(part.DV)) + `"`)
	return out.String()
}

type dianNominaEmissionData struct {
	IssueDate, IssueTime, Ambiente, CUNE, SoftwareSC, QRURL string
}

func prepareDIANNominaEmission(source *dbpkg.EmpresaNominaDIANFuente, softwarePIN string) (*dianNominaEmissionData, error) {
	if source == nil {
		return nil, errors.New("fuente fiscal de nómina electrónica no disponible")
	}
	if blockers := dbpkg.ValidateEmpresaNominaDIANFuente(source); len(blockers) > 0 {
		return nil, fmt.Errorf("fuente fiscal de nómina electrónica incompleta: %s", strings.Join(blockers, " | "))
	}
	if strings.TrimSpace(source.NumeroLegal) == "" || source.Consecutivo <= 0 || strings.TrimSpace(source.Prefijo) == "" {
		return nil, errors.New("la numeración interna de nómina electrónica no está reservada")
	}
	emissionTime, err := time.Parse(time.RFC3339, strings.TrimSpace(source.FechaEmisionLegal))
	if err != nil {
		return nil, errors.New("fecha_emision_legal de nómina electrónica inválida")
	}
	softwarePIN = strings.TrimSpace(softwarePIN)
	if softwarePIN == "" {
		return nil, errors.New("software PIN DIAN no disponible para nómina electrónica")
	}
	out := &dianNominaEmissionData{IssueDate: emissionTime.Format("2006-01-02"), IssueTime: emissionTime.Format("15:04:05-07:00"), Ambiente: "2"}
	qrBase := "https://catalogo-vpfe-hab.dian.gov.co/document/searchqr?documentkey="
	if strings.EqualFold(strings.TrimSpace(source.TipoAmbiente), "produccion") {
		out.Ambiente = "1"
		qrBase = "https://catalogo-vpfe.dian.gov.co/document/searchqr?documentkey="
	} else if !strings.EqualFold(strings.TrimSpace(source.TipoAmbiente), "habilitacion") {
		return nil, errors.New("ambiente DIAN de nómina electrónica inválido")
	}
	out.CUNE = buildDIANNominaCUNE(source.NumeroLegal, out.IssueDate, out.IssueTime, source.Devengados.Total,
		source.Deducciones.Total, source.ComprobanteTotal, source.Empleador.NIT, source.Trabajador.NumeroDocumento,
		softwarePIN, out.Ambiente)
	out.SoftwareSC = strings.ToLower(buildDIANSHA384Hex(source.SoftwareID, softwarePIN, source.NumeroLegal))
	out.QRURL = qrBase + out.CUNE
	return out, nil
}

func dianNominaPeriodAttrs(source *dbpkg.EmpresaNominaDIANFuente, issueDate string) string {
	retirement := dianNominaOptionalAttr("FechaRetiro", source.FechaRetiro)
	return ` FechaIngreso="` + escapeXML(source.FechaIngreso) + `"` + retirement +
		fmt.Sprintf(` FechaLiquidacionInicio="%s" FechaLiquidacionFin="%s" TiempoLaborado="%d" FechaGen="%s"`,
			escapeXML(source.PeriodoDesde), escapeXML(source.PeriodoHasta), source.TiempoLaborado, escapeXML(issueDate))
}

func dianNominaPartiesXML(source *dbpkg.EmpresaNominaDIANFuente, softwareSC string) (string, string, string) {
	provider := "<ProveedorXML" + dianNominaParteAttrs(source.ProveedorXML) + ` SoftwareID="` + escapeXML(source.SoftwareID) +
		`" SoftwareSC="` + escapeXML(softwareSC) + `"/>`
	employer := "<Empleador" + dianNominaParteAttrs(source.Empleador) + ` Pais="` + escapeXML(source.Empleador.Pais) +
		`" DepartamentoEstado="` + escapeXML(source.Empleador.Departamento) + `" MunicipioCiudad="` +
		escapeXML(source.Empleador.Municipio) + `" Direccion="` + escapeXML(source.Empleador.Direccion) + `"/>`
	worker := source.Trabajador
	workerXML := `<Trabajador TipoTrabajador="` + escapeXML(worker.TipoTrabajador) + `" SubTipoTrabajador="` + escapeXML(worker.SubTipoTrabajador) +
		`" AltoRiesgoPension="` + dianNominaBool(worker.AltoRiesgoPension) + `" TipoDocumento="` + escapeXML(worker.TipoDocumento) +
		`" NumeroDocumento="` + escapeXML(worker.NumeroDocumento) + `" PrimerApellido="` + escapeXML(worker.PrimerApellido) +
		`" SegundoApellido="` + escapeXML(worker.SegundoApellido) + `" PrimerNombre="` + escapeXML(worker.PrimerNombre) + `"` +
		dianNominaOptionalAttr("OtrosNombres", worker.OtrosNombres) + ` LugarTrabajoPais="` + escapeXML(worker.LugarTrabajoPais) +
		`" LugarTrabajoDepartamentoEstado="` + escapeXML(worker.LugarTrabajoDepartamento) + `" LugarTrabajoMunicipioCiudad="` +
		escapeXML(worker.LugarTrabajoMunicipio) + `" LugarTrabajoDireccion="` + escapeXML(worker.LugarTrabajoDireccion) +
		`" SalarioIntegral="` + dianNominaBool(worker.SalarioIntegral) + `" TipoContrato="` + escapeXML(worker.TipoContrato) +
		`" Sueldo="` + dianNominaMoney(worker.Sueldo) + `"` + dianNominaOptionalAttr("CodigoTrabajador", worker.CodigoTrabajador) + `/>`
	return provider, employer, workerXML
}

func dianNominaPaymentXML(source *dbpkg.EmpresaNominaDIANFuente) string {
	payment := source.Pago
	out := `<Pago Forma="1" Metodo="` + escapeXML(payment.Metodo) + `"` + dianNominaOptionalAttr("Banco", payment.Banco) +
		dianNominaOptionalAttr("TipoCuenta", payment.TipoCuenta) + dianNominaOptionalAttr("NumeroCuenta", payment.NumeroCuenta) + `/><FechasPagos>`
	for _, paymentDate := range source.FechasPago {
		out += "<FechaPago>" + escapeXML(valueutil.TrimmedPrefix(paymentDate, 10)) + "</FechaPago>"
	}
	return out + "</FechasPagos>"
}

func dianNominaEarningsXML(dev dbpkg.EmpresaNominaDIANDevengados) string {
	out := `<Devengados><Basico DiasTrabajados="` + strconv.Itoa(dev.DiasTrabajados) + `" SueldoTrabajado="` + dianNominaMoney(dev.SueldoTrabajado) + `"/>`
	if dev.AuxilioTransporte > 0 {
		out += `<Transporte AuxilioTransporte="` + dianNominaMoney(dev.AuxilioTransporte) + `"/>`
	}
	if dev.BonificacionSalarial > 0 {
		out += `<Bonificaciones><Bonificacion BonificacionS="` + dianNominaMoney(dev.BonificacionSalarial) + `"/></Bonificaciones>`
	}
	if dev.Comisiones > 0 {
		out += `<Comisiones><Comision>` + dianNominaMoney(dev.Comisiones) + `</Comision></Comisiones>`
	}
	return out + `</Devengados>`
}

func dianNominaDeductionsXML(ded dbpkg.EmpresaNominaDIANDeducciones) string {
	out := `<Deducciones><Salud Porcentaje="` + dianNominaMoney(ded.PorcentajeSalud) + `" Deduccion="` + dianNominaMoney(ded.Salud) + `"/>` +
		`<FondoPension Porcentaje="` + dianNominaMoney(ded.PorcentajePension) + `" Deduccion="` + dianNominaMoney(ded.Pension) + `"/>`
	if ded.FondoSolidario > 0 || ded.PorcentajeFondoSolidario > 0 {
		out += `<FondoSP Porcentaje="` + dianNominaMoney(ded.PorcentajeFondoSolidario) + `" DeduccionSP="` + dianNominaMoney(ded.FondoSolidario) + `"/>`
	}
	if ded.DeduccionFija > 0 || ded.OtrasDeducciones > 0 {
		out += "<OtrasDeducciones>"
		for _, amount := range []float64{ded.DeduccionFija, ded.OtrasDeducciones} {
			if amount > 0 {
				out += "<OtraDeduccion>" + dianNominaMoney(amount) + "</OtraDeduccion>"
			}
		}
		out += "</OtrasDeducciones>"
	}
	return out + `</Deducciones>`
}

func buildDIANNominaIndividualXML(source *dbpkg.EmpresaNominaDIANFuente, emission *dianNominaEmissionData) string {
	provider, employer, workerXML := dianNominaPartiesXML(source, emission.SoftwareSC)
	worker := source.Trabajador
	dev, ded := source.Devengados, source.Deducciones
	return `<?xml version="1.0" encoding="UTF-8" standalone="no"?>` +
		`<NominaIndividual xmlns="` + nominaElectronicaNamespace + `" xmlns:ds="http://www.w3.org/2000/09/xmldsig#" xmlns:ext="urn:oasis:names:specification:ubl:schema:xsd:CommonExtensionComponents-2" xmlns:xades="http://uri.etsi.org/01903/v1.3.2#" xmlns:xades141="http://uri.etsi.org/01903/v1.4.1#" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" SchemaLocation="" xsi:schemaLocation="` + nominaElectronicaNamespace + ` NominaIndividualElectronicaXSD.xsd">` +
		`<ext:UBLExtensions><ext:UBLExtension><ext:ExtensionContent></ext:ExtensionContent></ext:UBLExtension></ext:UBLExtensions>` +
		`<Periodo` + dianNominaPeriodAttrs(source, emission.IssueDate) + `/>` +
		`<NumeroSecuenciaXML` + dianNominaOptionalAttr("CodigoTrabajador", worker.CodigoTrabajador) + ` Prefijo="` + escapeXML(source.Prefijo) +
		`" Consecutivo="` + strconv.FormatInt(source.Consecutivo, 10) + `" Numero="` + escapeXML(source.NumeroLegal) + `"/>` +
		`<LugarGeneracionXML Pais="` + escapeXML(source.Empleador.Pais) + `" DepartamentoEstado="` + escapeXML(source.Empleador.Departamento) +
		`" MunicipioCiudad="` + escapeXML(source.Empleador.Municipio) + `" Idioma="es"/>` + provider +
		`<CodigoQR>` + escapeXML(emission.QRURL) + `</CodigoQR><InformacionGeneral Version="` + escapeXML(nominaElectronicaVersion) +
		`" Ambiente="` + emission.Ambiente + `" TipoXML="` + nominaElectronicaTipoXML + `" CUNE="` + escapeXML(emission.CUNE) +
		`" EncripCUNE="CUNE-SHA384" FechaGen="` + escapeXML(emission.IssueDate) + `" HoraGen="` + escapeXML(emission.IssueTime) +
		`" PeriodoNomina="` + strconv.Itoa(source.PeriodoNomina) + `" TipoMoneda="COP"/>` + employer + workerXML +
		dianNominaPaymentXML(source) + dianNominaEarningsXML(dev) + dianNominaDeductionsXML(ded) +
		`<DevengadosTotal>` + dianNominaMoney(dev.Total) + `</DevengadosTotal><DeduccionesTotal>` + dianNominaMoney(ded.Total) +
		`</DeduccionesTotal><ComprobanteTotal>` + dianNominaMoney(source.ComprobanteTotal) + `</ComprobanteTotal></NominaIndividual>`
}

func generateDIANNominaIndividualDesdeFuente(source *dbpkg.EmpresaNominaDIANFuente, softwarePIN string) (map[string]interface{}, int, error) {
	emission, err := prepareDIANNominaEmission(source, softwarePIN)
	if err != nil {
		return nil, http.StatusUnprocessableEntity, err
	}
	return map[string]interface{}{
		"ok": true, "empresa_id": source.EmpresaID, "documento_tipo": "nomina_electronica",
		"documento_codigo": fmt.Sprintf("NE-NOMINA-%d", source.LiquidacionID), "numero_legal": source.NumeroLegal,
		"cune": emission.CUNE, "uuid": emission.CUNE, "uuid_scheme": "CUNE-SHA384", "profile_execution_id": emission.Ambiente,
		"software_security_code": "[calculado]", "xml_ubl_base": buildDIANNominaIndividualXML(source, emission),
		"estado_preparacion": "pre_envio_validable", "soap_operacion": "SendNominaSync",
	}, http.StatusOK, nil
}

type dianNominaXMLInspection struct {
	Root                    xml.Name
	RootSchemaLocation      string
	RootSchemaLocationSet   bool
	XSISchemaLocation       string
	NumeroSecuencia         map[string]string
	InformacionGeneral      map[string]string
	ProveedorXML            map[string]string
	Empleador               map[string]string
	Trabajador              map[string]string
	Texts                   map[string][]string
	SignaturePresent        bool
	X509Present             bool
	SecretNamedFieldPresent bool
}

func dianNominaXMLAttributes(start xml.StartElement) map[string]string {
	out := make(map[string]string, len(start.Attr))
	for _, attr := range start.Attr {
		out[attr.Name.Local] = strings.TrimSpace(attr.Value)
	}
	return out
}

func inspectDIANNominaXML(xmlPayload string) (*dianNominaXMLInspection, error) {
	decoder := xml.NewDecoder(strings.NewReader(strings.TrimSpace(xmlPayload)))
	inspection := &dianNominaXMLInspection{Texts: map[string][]string{}}
	stack := make([]xml.Name, 0, 24)
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch node := token.(type) {
		case xml.StartElement:
			if len(stack) == 0 {
				inspection.Root = node.Name
				for _, attr := range node.Attr {
					if attr.Name.Space == "" && attr.Name.Local == "SchemaLocation" {
						inspection.RootSchemaLocationSet = true
						inspection.RootSchemaLocation = strings.TrimSpace(attr.Value)
					}
					if attr.Name.Space == "http://www.w3.org/2001/XMLSchema-instance" && attr.Name.Local == "schemaLocation" {
						inspection.XSISchemaLocation = strings.TrimSpace(attr.Value)
					}
				}
			}
			attrs := dianNominaXMLAttributes(node)
			switch node.Name.Local {
			case "NumeroSecuenciaXML":
				inspection.NumeroSecuencia = attrs
			case "InformacionGeneral":
				inspection.InformacionGeneral = attrs
			case "ProveedorXML":
				inspection.ProveedorXML = attrs
			case "Empleador":
				inspection.Empleador = attrs
			case "Trabajador":
				inspection.Trabajador = attrs
			case "Signature":
				if node.Name.Space == "http://www.w3.org/2000/09/xmldsig#" {
					inspection.SignaturePresent = true
				}
			case "X509Certificate":
				inspection.X509Present = true
			}
			for name := range attrs {
				lower := strings.ToLower(strings.TrimSpace(name))
				if lower == "pin" || lower == "softwarepin" || lower == "software_pin" {
					inspection.SecretNamedFieldPresent = true
				}
			}
			lowerName := strings.ToLower(strings.TrimSpace(node.Name.Local))
			if lowerName == "pin" || lowerName == "softwarepin" || lowerName == "software_pin" {
				inspection.SecretNamedFieldPresent = true
			}
			stack = append(stack, node.Name)
		case xml.CharData:
			if len(stack) == 0 {
				continue
			}
			value := strings.TrimSpace(string(node))
			if value != "" {
				name := stack[len(stack)-1].Local
				inspection.Texts[name] = append(inspection.Texts[name], value)
			}
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		}
	}
	if inspection.Root.Local == "" {
		return nil, fmt.Errorf("XML de nómina electrónica vacío")
	}
	return inspection, nil
}

func dianNominaInspectionText(inspection *dianNominaXMLInspection, name string) string {
	if inspection == nil {
		return ""
	}
	values := inspection.Texts[name]
	if len(values) == 0 {
		return ""
	}
	return strings.TrimSpace(values[0])
}

func dianNominaStringSlicesEqual(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if strings.TrimSpace(left[i]) != strings.TrimSpace(right[i]) {
			return false
		}
	}
	return true
}

type dianNominaPreflightState struct {
	EmpresaID               int64
	Stage, Ambiente         string
	SoftwareID, SoftwarePIN string
	CredentialError         error
	Source                  *dbpkg.EmpresaNominaDIANFuente
	Issues, Warnings        []map[string]interface{}
	Checks                  map[string]interface{}
}

func (state *dianNominaPreflightState) addError(code, field, message, sourceName string) {
	dianAppendValidationIssue(&state.Issues, &state.Warnings, code, "error", field, message, sourceName)
}

func newDIANNominaPreflightState(cfg map[string]interface{}, empresaID int64, source *dbpkg.EmpresaNominaDIANFuente, stage string) *dianNominaPreflightState {
	stage = strings.ToLower(strings.TrimSpace(stage))
	if stage == "" {
		stage = "validacion_manual"
	}
	softwareID, softwarePIN, _, credentialErr := resolveDIANSoftwareCredentials(cfg, nil, empresaID)
	return &dianNominaPreflightState{
		EmpresaID: empresaID, Stage: stage, Ambiente: chooseDIANAmbiente(cfg), Source: source,
		SoftwareID: softwareID, SoftwarePIN: softwarePIN, CredentialError: credentialErr, Issues: make([]map[string]interface{}, 0),
		Warnings: make([]map[string]interface{}, 0), Checks: map[string]interface{}{},
	}
}

func validateDIANNominaPreflightSourceConfig(state *dianNominaPreflightState) {
	if state.EmpresaID <= 0 {
		state.addError("DIAN-NOM-EMPRESA", "empresa_id", "empresa_id es obligatorio", "preflight_nomina")
	}
	if state.Source == nil || state.Source.EmpresaID != state.EmpresaID {
		state.addError("DIAN-NOM-FUENTE", "fuente_fiscal", "la fuente fiscal de nómina no pertenece a la empresa", "fuente_fiscal_inmutable")
	} else {
		blockers := dbpkg.ValidateEmpresaNominaDIANFuente(state.Source)
		state.Checks["fuente_fiscal"] = map[string]interface{}{"ok": len(blockers) == 0, "bloqueos": blockers, "liquidacion_id": state.Source.LiquidacionID}
		for _, blocker := range blockers {
			state.addError("DIAN-NOM-FUENTE", "fuente_fiscal", blocker, "fuente_fiscal_inmutable")
		}
	}
	if state.Ambiente != "habilitacion" && state.Ambiente != "produccion" {
		state.addError("DIAN-NOM-AMBIENTE", "tipo_ambiente", "ambiente DIAN debe ser habilitación o producción", "configuracion_dian")
	}
	if state.Source != nil && !strings.EqualFold(strings.TrimSpace(state.Source.TipoAmbiente), state.Ambiente) {
		state.addError("DIAN-NOM-AMBIENTE-FUENTE", "tipo_ambiente", "el ambiente DIAN no coincide con la instantánea fiscal reservada", "fuente_fiscal_inmutable")
	}
	if state.CredentialError != nil || strings.TrimSpace(state.SoftwareID) == "" || strings.TrimSpace(state.SoftwarePIN) == "" {
		state.addError("DIAN-NOM-SOFTWARE", "software_id", "credenciales del software DIAN de nómina no disponibles", "configuracion_dian")
	}
	if state.Source != nil && state.SoftwareID != "" && !strings.EqualFold(strings.TrimSpace(state.Source.SoftwareID), strings.TrimSpace(state.SoftwareID)) {
		state.addError("DIAN-NOM-SOFTWARE-FUENTE", "software_id", "el Software ID cambió después de reservar la fuente fiscal", "fuente_fiscal_inmutable")
	}
}

func validateDIANNominaPreflightTransport(state *dianNominaPreflightState, cfg map[string]interface{}) {
	if state.Stage != "reserva" && state.Stage != "pre_envio" && state.Stage != "envio_real" {
		return
	}
	endpoint := dianConfiguredEndpoint(cfg, nil)
	if endpoint == "" {
		state.addError("DIAN-NOM-ENDPOINT", "url_dian", "endpoint DIAN de nómina no configurado o no autorizado", "configuracion_dian")
	} else {
		normalized := strings.ToLower(normalizeDIANSOAPEndpoint(endpoint))
		if state.Ambiente == "habilitacion" && !strings.Contains(normalized, "vpfe-hab.dian.gov.co") {
			state.addError("DIAN-NOM-ENDPOINT-AMBIENTE", "url_dian", "el preflight de habilitación exige el endpoint oficial vpfe-hab de DIAN", "configuracion_dian")
		}
		if state.Ambiente == "produccion" && strings.Contains(normalized, "vpfe-hab.dian.gov.co") {
			state.addError("DIAN-NOM-ENDPOINT-AMBIENTE", "url_dian", "producción no puede usar el endpoint de habilitación DIAN", "configuracion_dian")
		}
	}
	keyRef := strings.TrimSpace(genericStringValue(cfg["certificado_clave_ref"]))
	if keyRef == "" {
		state.addError("DIAN-NOM-FIRMA", "certificado_clave_ref", "llave privada de firma obligatoria antes del envío", "firma_digital")
	} else if _, err := parseDIANRSAPrivateKey(keyRef, state.EmpresaID); err != nil {
		state.addError("DIAN-NOM-FIRMA", "certificado_clave_ref", "llave privada de firma inválida", "firma_digital")
	}
	certRef := strings.TrimSpace(genericStringValue(cfg["certificado_url"]))
	if certRef == "" {
		state.addError("DIAN-NOM-CERTIFICADO", "certificado_url", "certificado X.509 obligatorio antes del envío", "firma_digital")
	} else if cert, err := parseDIANCertificate(certRef, state.EmpresaID); err != nil {
		state.addError("DIAN-NOM-CERTIFICADO", "certificado_url", "certificado X.509 inválido", "firma_digital")
	} else if now := time.Now(); now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
		state.addError("DIAN-NOM-CERTIFICADO-VIGENCIA", "certificado_url", "certificado X.509 fuera de vigencia", "firma_digital")
	}
}

func validateDIANNominaXMLSource(state *dianNominaPreflightState, inspection *dianNominaXMLInspection) {
	if state.Source == nil {
		return
	}
	source := state.Source
	emissionTime, err := time.Parse(time.RFC3339, strings.TrimSpace(source.FechaEmisionLegal))
	if err != nil {
		state.addError("DIAN-NOM-FECHA", "fecha_emision_legal", "fecha de emisión fiscal reservada inválida", "fuente_fiscal_inmutable")
		return
	}
	issueDate, issueTime, environmentCode := emissionTime.Format("2006-01-02"), emissionTime.Format("15:04:05-07:00"), "2"
	if state.Ambiente == "produccion" {
		environmentCode = "1"
	}
	expectedCUNE := buildDIANNominaCUNE(source.NumeroLegal, issueDate, issueTime, source.Devengados.Total, source.Deducciones.Total,
		source.ComprobanteTotal, source.Empleador.NIT, source.Trabajador.NumeroDocumento, state.SoftwarePIN, environmentCode)
	expectedSoftwareSC := strings.ToLower(buildDIANSHA384Hex(state.SoftwareID, state.SoftwarePIN, source.NumeroLegal))
	state.Checks["cune"] = map[string]interface{}{"esperado": expectedCUNE, "xml": strings.ToLower(inspection.InformacionGeneral["CUNE"]), "ok": strings.EqualFold(inspection.InformacionGeneral["CUNE"], expectedCUNE)}
	fields := map[string][2]string{
		"NumeroSecuenciaXML.Numero": {inspection.NumeroSecuencia["Numero"], source.NumeroLegal}, "NumeroSecuenciaXML.Prefijo": {inspection.NumeroSecuencia["Prefijo"], source.Prefijo},
		"NumeroSecuenciaXML.Consecutivo": {inspection.NumeroSecuencia["Consecutivo"], strconv.FormatInt(source.Consecutivo, 10)}, "InformacionGeneral.Version": {inspection.InformacionGeneral["Version"], nominaElectronicaVersion},
		"InformacionGeneral.Ambiente": {inspection.InformacionGeneral["Ambiente"], environmentCode}, "InformacionGeneral.TipoXML": {inspection.InformacionGeneral["TipoXML"], nominaElectronicaTipoXML},
		"InformacionGeneral.CUNE": {strings.ToLower(inspection.InformacionGeneral["CUNE"]), expectedCUNE}, "InformacionGeneral.EncripCUNE": {inspection.InformacionGeneral["EncripCUNE"], "CUNE-SHA384"},
		"InformacionGeneral.FechaGen": {inspection.InformacionGeneral["FechaGen"], issueDate}, "InformacionGeneral.HoraGen": {inspection.InformacionGeneral["HoraGen"], issueTime},
		"InformacionGeneral.PeriodoNomina": {inspection.InformacionGeneral["PeriodoNomina"], strconv.Itoa(source.PeriodoNomina)}, "ProveedorXML.SoftwareID": {inspection.ProveedorXML["SoftwareID"], source.SoftwareID},
		"ProveedorXML.SoftwareSC": {strings.ToLower(inspection.ProveedorXML["SoftwareSC"]), expectedSoftwareSC}, "Empleador.NIT": {dianOnlyDigits(inspection.Empleador["NIT"]), dianOnlyDigits(source.Empleador.NIT)},
		"Trabajador.NumeroDocumento": {inspection.Trabajador["NumeroDocumento"], source.Trabajador.NumeroDocumento}, "DevengadosTotal": {dianNominaInspectionText(inspection, "DevengadosTotal"), dianNominaMoney(source.Devengados.Total)},
		"DeduccionesTotal": {dianNominaInspectionText(inspection, "DeduccionesTotal"), dianNominaMoney(source.Deducciones.Total)}, "ComprobanteTotal": {dianNominaInspectionText(inspection, "ComprobanteTotal"), dianNominaMoney(source.ComprobanteTotal)},
	}
	for field, pair := range fields {
		if pair[0] != pair[1] {
			state.addError("DIAN-NOM-CONTENIDO", field, field+" no coincide con la fuente fiscal sellada", "fuente_fiscal_inmutable")
		}
	}
	paymentDates := make([]string, 0, len(inspection.Texts["FechaPago"]))
	for _, paymentDate := range inspection.Texts["FechaPago"] {
		paymentDates = append(paymentDates, valueutil.TrimmedPrefix(paymentDate, 10))
	}
	if !dianNominaStringSlicesEqual(paymentDates, source.FechasPago) {
		state.addError("DIAN-NOM-CONTENIDO", "FechasPagos", "las fechas de pago XML no coinciden con la fuente fiscal sellada", "fuente_fiscal_inmutable")
	}
}

func validateDIANNominaPreflightXML(state *dianNominaPreflightState, xmlPayload string) {
	xmlChecks := map[string]interface{}{"presente": strings.TrimSpace(xmlPayload) != ""}
	state.Checks["xml_nomina"] = xmlChecks
	if strings.TrimSpace(xmlPayload) == "" {
		if state.Stage == "pre_envio" || state.Stage == "envio_real" {
			state.addError("DIAN-NOM-XML", "xml_firmado", "XML firmado de nómina obligatorio antes del envío", "anexo_nomina_dian")
		}
		return
	}
	inspection, err := inspectDIANNominaXML(xmlPayload)
	if err != nil {
		state.addError("DIAN-NOM-XML-FORMATO", "xml_firmado", "XML de nómina no está bien formado", "anexo_nomina_dian")
		return
	}
	xmlChecks["root"], xmlChecks["namespace"] = inspection.Root.Local, inspection.Root.Space
	xmlChecks["signature_presente"], xmlChecks["x509_presente"] = inspection.SignaturePresent, inspection.X509Present
	xmlChecks["schema_location_presente"] = inspection.RootSchemaLocationSet
	if inspection.Root.Local != "NominaIndividual" || inspection.Root.Space != nominaElectronicaNamespace {
		state.addError("DIAN-NOM-ROOT", "xml_firmado", "raíz o namespace de NominaIndividual inválido", "xsd_nomina_dian")
	}
	if !inspection.RootSchemaLocationSet || inspection.XSISchemaLocation != nominaElectronicaNamespace+" NominaIndividualElectronicaXSD.xsd" {
		state.addError("DIAN-NOM-SCHEMA", "schemaLocation", "referencia de esquema de NominaIndividual inválida", "xsd_nomina_dian")
	}
	if inspection.SecretNamedFieldPresent {
		state.addError("DIAN-NOM-PIN", "xml_firmado", "el XML no puede exponer el PIN del software DIAN", "seguridad_secretos")
	}
	validateDIANNominaXMLSource(state, inspection)
	if (state.Stage == "pre_envio" || state.Stage == "envio_real") && !inspection.SignaturePresent {
		state.addError("DIAN-NOM-FIRMA-XML", "xml_firmado", "XML de nómina debe incluir ds:Signature", "firma_digital")
	}
	if (state.Stage == "pre_envio" || state.Stage == "envio_real") && !inspection.X509Present {
		state.addError("DIAN-NOM-X509-XML", "xml_firmado", "XML de nómina debe incluir X509Certificate", "firma_digital")
	}
}

func validateDIANNominaDocumentPreflight(cfg map[string]interface{}, empresaID int64, source *dbpkg.EmpresaNominaDIANFuente, xmlPayload, stage string) map[string]interface{} {
	state := newDIANNominaPreflightState(cfg, empresaID, source, stage)
	validateDIANNominaPreflightSourceConfig(state)
	validateDIANNominaPreflightTransport(state, cfg)
	validateDIANNominaPreflightXML(state, xmlPayload)
	ok := len(state.Issues) == 0
	return map[string]interface{}{
		"ok": ok, "bloqueado": !ok, "empresa_id": empresaID, "etapa": state.Stage, "ambiente": state.Ambiente,
		"issues": state.Issues, "warnings": state.Warnings, "checks": state.Checks,
		"fuente_normativa": []string{"DIAN - Anexo técnico Documento Soporte de Pago de Nómina Electrónica 1.0", "DIAN - WSDL SendNominaSync"},
		"total_errores":    len(state.Issues), "total_advertencias": len(state.Warnings), "validacion_en_sistema": true,
	}
}

func nominaElectronicaCodigoInterno(liquidacionID int64) string {
	return fmt.Sprintf("NE-NOMINA-%d", liquidacionID)
}

func nominaElectronicaLiquidacionDesdeCodigo(raw string) (int64, error) {
	code := strings.ToUpper(strings.TrimSpace(raw))
	if !strings.HasPrefix(code, "NE-NOMINA-") {
		return 0, errors.New("código interno de nómina electrónica inválido")
	}
	id, err := strconv.ParseInt(strings.TrimPrefix(code, "NE-NOMINA-"), 10, 64)
	if err != nil || id <= 0 || nominaElectronicaCodigoInterno(id) != code {
		return 0, errors.New("código interno de nómina electrónica inválido")
	}
	return id, nil
}

func nominaElectronicaMergeDIANConfig(base map[string]interface{}, snapshot *dbpkg.EmpresaNominaDIANConfiguracionSnapshot) map[string]interface{} {
	out := make(map[string]interface{}, len(base)+8)
	for key, value := range base {
		out[key] = value
	}
	if snapshot == nil {
		return out
	}
	out["tipo_ambiente"] = snapshot.TipoAmbiente
	out["prefijo"] = snapshot.Prefijo
	out["consecutivo_actual"] = snapshot.ConsecutivoAsignado
	out["modo_operacion_codigo"] = snapshot.ModoOperacionCodigo
	out["test_set_id"] = snapshot.TestSetID
	// Nómina usa numeración interna del empleador, no resolución ni rango de
	// factura de venta. Vaciar esos campos evita que validaciones comerciales se
	// mezclen con la familia documental de nómina.
	out["resolucion_numero"] = ""
	out["resolucion_fecha_desde"] = ""
	out["resolucion_fecha_hasta"] = ""
	out["rango_desde"] = int64(0)
	out["rango_hasta"] = int64(0)
	if strings.TrimSpace(snapshot.URLDIANOverride) != "" {
		out["url_dian"] = strings.TrimSpace(snapshot.URLDIANOverride)
	}
	return out
}

func validateNominaElectronicaEmployerDIANConfig(cfg map[string]interface{}, source *dbpkg.EmpresaNominaDIANFuente) error {
	if source == nil {
		return errors.New("fuente fiscal de nómina electrónica no disponible")
	}
	configuredNIT := dianOnlyDigits(genericStringValue(cfg["nit"]))
	configuredDV := strings.TrimSpace(genericStringValue(cfg["digito_verificacion"]))
	if configuredNIT == "" || len(configuredDV) != 1 || configuredDV[0] < '0' || configuredDV[0] > '9' {
		return errors.New("la configuración DIAN principal no tiene NIT y DV completos")
	}
	if configuredNIT != dianOnlyDigits(source.Empleador.NIT) || configuredDV != strings.TrimSpace(source.Empleador.DV) {
		return errors.New("el NIT/DV de la configuración DIAN principal no coincide con la identidad fiscal del empleador")
	}
	return nil
}

func loadNominaElectronicaParaDispatch(ctx context.Context, dbEmp *sql.DB, doc dbpkg.EmpresaDocumentoFacturacion) (*dbpkg.EmpresaNominaElectronica, *dbpkg.EmpresaNominaDIANFuente, *dbpkg.EmpresaNominaDIANConfiguracionSnapshot, error) {
	if dbEmp == nil || doc.EmpresaID <= 0 || normalizeFacturacionDocumentoElectronicoTipo(doc.TipoDocumento) != "nomina_electronica" {
		return nil, nil, nil, errors.New("documento fiscal de nómina inválido")
	}
	liquidacionID, err := nominaElectronicaLiquidacionDesdeCodigo(doc.DocumentoCodigo)
	if err != nil {
		return nil, nil, nil, err
	}
	row, err := dbpkg.GetEmpresaNominaElectronicaByLiquidacionContext(ctx, dbEmp, doc.EmpresaID, liquidacionID)
	if err != nil {
		return nil, nil, nil, err
	}
	source, snapshot, err := dbpkg.DecodeEmpresaNominaDIANSnapshots(row.FuenteFiscalJSON, row.ConfiguracionDIANJSON)
	if err != nil {
		return nil, nil, nil, err
	}
	if source.EmpresaID != doc.EmpresaID || source.LiquidacionID != liquidacionID || source.NominaID != row.ID ||
		source.EmpleadoNominaID != row.EmpleadoNominaID || source.Consecutivo != snapshot.ConsecutivoAsignado ||
		!strings.EqualFold(source.Prefijo, snapshot.Prefijo) || !strings.EqualFold(source.NumeroLegal, row.NumeroLegal) ||
		!strings.EqualFold(row.NumeroLegal, doc.NumeroLegal) || strings.TrimSpace(source.FechaEmisionLegal) == "" ||
		!strings.EqualFold(strings.TrimSpace(source.TipoAmbiente), strings.TrimSpace(snapshot.TipoAmbiente)) {
		return nil, nil, nil, errors.New("trazabilidad de nómina electrónica no coincide con la fuente fiscal reservada")
	}
	if !row.FuenteFiscalSellada {
		return nil, nil, nil, errors.New("la fuente fiscal de nómina electrónica aún no está sellada")
	}
	if math.Abs(row.Devengados-source.Devengados.Total) > 0.01 || math.Abs(row.Deducciones-source.Deducciones.Total) > 0.01 ||
		math.Abs(row.Total-source.ComprobanteTotal) > 0.01 || math.Abs(doc.MontoTotal-source.ComprobanteTotal) > 0.01 {
		return nil, nil, nil, errors.New("totales de nómina electrónica no coinciden con su fuente fiscal")
	}
	return row, source, snapshot, nil
}

func nominaElectronicaCUNEDesdeXML(xmlPayload string) string {
	inspection, err := inspectDIANNominaXML(xmlPayload)
	if err != nil {
		return ""
	}
	cune := strings.ToLower(strings.TrimSpace(inspection.InformacionGeneral["CUNE"]))
	if !valueutil.IsHexLength(cune, 96) {
		return ""
	}
	return cune
}

func updateNominaElectronicaDispatchMirror(dbEmp *sql.DB, doc dbpkg.EmpresaDocumentoFacturacion, row *dbpkg.EmpresaNominaElectronica, estado string, ok bool, envioResp map[string]interface{}, xmlFirmado, safeJSON string, registrarIntento bool) error {
	if row == nil {
		return nil
	}
	state := documentoSoporteEstadoDIANMirror(estado)
	if !ok && strings.TrimSpace(estado) == "" {
		state = "fallido"
	}
	cune := facturacionCUFEOficialDesdeMap(envioResp)
	if cune == "" {
		cune = nominaElectronicaCUNEDesdeXML(xmlFirmado)
	}
	return dbpkg.UpdateEmpresaNominaDIANResultContext(context.Background(), dbEmp, doc.EmpresaID, row.ID, state, cune, safeJSON, true, registrarIntento)
}

func loadOrGenerateSignedNominaXML(dbEmp *sql.DB, doc dbpkg.EmpresaDocumentoFacturacion, source *dbpkg.EmpresaNominaDIANFuente, dianCfg map[string]interface{}, softwarePIN string) (string, bool, bool, error) {
	if stored, err := loadFacturacionFiscalArtifact(context.Background(), dbEmp, doc, "xml_firmado"); err == nil {
		if xmlPayload := strings.TrimSpace(string(stored)); xmlPayload != "" {
			return xmlPayload, false, false, nil
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return "", false, false, fmt.Errorf("leer XML firmado de nómina: %w", err)
	}
	generated, _, err := generateDIANNominaIndividualDesdeFuente(source, softwarePIN)
	if err != nil {
		return "", false, true, fmt.Errorf("generar NominaIndividual DIAN: %w", err)
	}
	signPayload := map[string]interface{}{
		"empresa_id": doc.EmpresaID, "documento_tipo": "nomina_electronica",
		"documento_codigo": doc.DocumentoCodigo, "xml_ubl_base": genericStringValue(generated["xml_ubl_base"]),
	}
	signed, _, err := signDIANXMLXAdESBase(dianCfg, doc.EmpresaID, signPayload)
	if err != nil {
		return "", false, true, fmt.Errorf("firmar NominaIndividual DIAN: %w", err)
	}
	return strings.TrimSpace(genericStringValue(signed["xml_firmado"])), true, false, nil
}

func saveNominaDIANProviderResponse(dbEmp *sql.DB, doc dbpkg.EmpresaDocumentoFacturacion, envioResp map[string]interface{}) string {
	content := []byte(strings.TrimSpace(genericStringValue(envioResp["raw_response"])))
	extension, mimeType := ".xml", "application/xml"
	if len(content) == 0 {
		if encoded, err := json.Marshal(envioResp["respuesta_dian"]); err == nil {
			content = encoded
		}
		extension, mimeType = ".json", "application/json"
	}
	if len(content) == 0 || string(content) == "null" {
		return ""
	}
	if _, err := saveFacturacionFiscalArtifact(context.Background(), dbEmp, doc, "respuesta_proveedor", extension, mimeType, content); err != nil {
		return "DIAN respondió, pero no se pudo persistir el acuse privado de nómina: " + err.Error()
	}
	return ""
}

// dispatchNominaDIANOficial is the only production dispatch path for
// NominaIndividual. It deliberately does not call the invoice UBL generator or
// create an invoice-style PDF representation.
func dispatchNominaDIANOficial(dbEmp *sql.DB, payload facturacionOperacionPayload, doc dbpkg.EmpresaDocumentoFacturacion, apiBaseURL string) facturacionProveedorDispatchResult {
	row, source, snapshot, err := loadNominaElectronicaParaDispatch(context.Background(), dbEmp, doc)
	if err != nil {
		return facturacionProveedorDispatchResult{FinalFailure: true, Error: "fuente fiscal de nómina no disponible: " + err.Error()}
	}
	baseConfig, err := getEmpresaDIANConfig(dbEmp, doc.EmpresaID)
	if err != nil || len(baseConfig) == 0 {
		return facturacionProveedorDispatchResult{Error: "configuración DIAN principal no disponible para nómina"}
	}
	if err := validateNominaElectronicaEmployerDIANConfig(baseConfig, source); err != nil {
		return facturacionProveedorDispatchResult{FinalFailure: true, Error: err.Error()}
	}
	currentFamilyConfig, currentErr := dbpkg.GetEmpresaDIANDocumentoConfiguracionContext(context.Background(), dbEmp, doc.EmpresaID, "nomina_electronica")
	if currentErr != nil || currentFamilyConfig == nil || dbpkg.ValidateEmpresaNominaElectronicaConfigForEmission(*currentFamilyConfig) != nil {
		return facturacionProveedorDispatchResult{FinalFailure: true, Error: "configuración separada de nómina electrónica no está habilitada para emitir"}
	}
	if !strings.EqualFold(strings.TrimSpace(currentFamilyConfig.TipoAmbiente), "produccion") || !strings.EqualFold(strings.TrimSpace(currentFamilyConfig.Estado), "activo") {
		return facturacionProveedorDispatchResult{FinalFailure: true, Error: "SendNominaSync solo está habilitado para nómina activa en producción; la habilitación requiere su flujo de set de pruebas"}
	}
	if !strings.EqualFold(strings.TrimSpace(currentFamilyConfig.TipoAmbiente), strings.TrimSpace(snapshot.TipoAmbiente)) {
		return facturacionProveedorDispatchResult{FinalFailure: true, Error: "el ambiente de nómina cambió después de reservar el documento; requiere conciliación manual"}
	}
	dianCfg := nominaElectronicaMergeDIANConfig(baseConfig, snapshot)
	if strings.TrimSpace(snapshot.URLDIANOverride) == "" && strings.TrimSpace(apiBaseURL) != "" {
		dianCfg["url_dian"] = strings.TrimSpace(apiBaseURL)
	}
	softwareID, softwarePIN, _, credentialErr := resolveDIANSoftwareCredentials(dianCfg, nil, doc.EmpresaID)
	if credentialErr != nil || strings.TrimSpace(softwareID) == "" || strings.TrimSpace(softwarePIN) == "" {
		return facturacionProveedorDispatchResult{FinalFailure: true, Error: "credenciales del software DIAN de nómina no disponibles"}
	}
	if !strings.EqualFold(strings.TrimSpace(source.SoftwareID), strings.TrimSpace(softwareID)) {
		return facturacionProveedorDispatchResult{FinalFailure: true, Error: "Software ID de nómina cambió después de reservar la fuente fiscal"}
	}

	xmlFirmado, xmlNuevo, finalFailure, err := loadOrGenerateSignedNominaXML(dbEmp, doc, source, dianCfg, softwarePIN)
	if err != nil {
		return facturacionProveedorDispatchResult{FinalFailure: finalFailure, Error: err.Error()}
	}
	preflight := validateDIANNominaDocumentPreflight(dianCfg, doc.EmpresaID, source, xmlFirmado, "envio_real")
	if parseTruthy(genericStringValue(preflight["bloqueado"])) {
		raw, _ := json.Marshal(preflight)
		return facturacionProveedorDispatchResult{FinalFailure: true, Error: "validación preventiva de nómina DIAN no superada", RespuestaJSON: string(raw)}
	}
	if xmlNuevo {
		if _, err := saveFacturacionFiscalArtifact(context.Background(), dbEmp, doc, "xml_firmado", ".xml", "application/xml", []byte(xmlFirmado)); err != nil {
			return facturacionProveedorDispatchResult{Error: "persistir XML firmado de nómina antes del envío: " + err.Error()}
		}
	}
	cune := nominaElectronicaCUNEDesdeXML(xmlFirmado)
	envioPayload := map[string]interface{}{
		"empresa_id": doc.EmpresaID, "documento_codigo": doc.DocumentoCodigo, "documento_tipo": "nomina_electronica",
		"numero_legal": doc.NumeroLegal, "xml_firmado": xmlFirmado, "cune": cune, "cufe": cune,
		"total": dianNominaMoney(source.ComprobanteTotal), "fecha_emision": source.FechaEmisionLegal,
		"moneda": "COP", "usar_soap_dian": true, "soap_operacion": "SendNominaSync",
	}
	if endpoint := strings.TrimSpace(genericStringValue(dianCfg["url_dian"])); endpoint != "" {
		envioPayload["url_dian"] = endpoint
	}
	envioResp, _, sendErr := sendDIANDocumentoReal(dbEmp, dianCfg, doc.EmpresaID, envioPayload)
	if sendErr != nil {
		return facturacionProveedorDispatchResult{Error: sendErr.Error()}
	}
	safeJSON := facturacionSafeDispatchJSON(envioResp)
	artifactWarning := saveNominaDIANProviderResponse(dbEmp, doc, envioResp)
	estado := strings.ToLower(strings.TrimSpace(genericStringValue(envioResp["estado_dian"])))
	trackID := strings.TrimSpace(genericStringValue(envioResp["track_id"]))
	ok := parseTruthy(genericStringValue(envioResp["ok"])) || trackID != "" || estado == "enviado" || estado == "aceptado"
	if updateErr := updateNominaElectronicaDispatchMirror(dbEmp, doc, row, estado, ok, envioResp, xmlFirmado, safeJSON, true); updateErr != nil {
		return facturacionProveedorDispatchResult{Error: "DIAN respondió, pero no se pudo actualizar el espejo de nómina", RespuestaJSON: safeJSON, ArtifactWarning: artifactWarning}
	}
	if !ok {
		errMsg := dianFirstNonBlank(genericStringValue(envioResp["acuse_mensaje"]), genericStringValue(envioResp["error"]), "DIAN no aceptó la nómina electrónica")
		finalFailure := estado == "rechazado" || strings.EqualFold(genericStringValue(envioResp["acuse_estado"]), "rechazado")
		return facturacionProveedorDispatchResult{FinalFailure: finalFailure, Error: errMsg, RespuestaJSON: safeJSON, ArtifactWarning: artifactWarning, HTTPStatus: int(anyToInt64(envioResp["http_status"]))}
	}
	ref := trackID
	if ref == "" {
		ref = strings.TrimSpace(genericStringValue(envioResp["zip_key"]))
	}
	return facturacionProveedorDispatchResult{Success: true, Pending: estado != "aceptado", ReferenciaExterna: ref, RespuestaJSON: safeJSON, ArtifactWarning: artifactWarning, HTTPStatus: int(anyToInt64(envioResp["http_status"]))}
}

const nominaElectronicaConfirmacion = "EMITIR NOMINA ELECTRONICA DIAN"

type nominaElectronicaEmisionRequest struct {
	EmpresaID               int64  `json:"empresa_id"`
	LiquidacionID           int64  `json:"liquidacion_id"`
	ConfirmarEmision        bool   `json:"confirmar_emision"`
	MensajeConfirmacionDIAN string `json:"mensaje_confirmacion_dian"`
}

func decodeNominaElectronicaEmisionRequest(w http.ResponseWriter, r *http.Request) (nominaElectronicaEmisionRequest, bool) {
	var request nominaElectronicaEmisionRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Solicitud de emisión de nómina electrónica inválida", http.StatusBadRequest)
		return request, false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		http.Error(w, "Solicitud de emisión de nómina electrónica inválida", http.StatusBadRequest)
		return request, false
	}
	if err := facturacionBindAuthorizedEmpresaID(r, &request.EmpresaID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return request, false
	}
	if request.LiquidacionID <= 0 {
		http.Error(w, "liquidacion_id es obligatorio", http.StatusBadRequest)
		return request, false
	}
	if !request.ConfirmarEmision || strings.TrimSpace(request.MensajeConfirmacionDIAN) != nominaElectronicaConfirmacion {
		writeJSON(w, http.StatusConflict, map[string]interface{}{
			"ok": false, "bloqueado": true, "codigo": "confirmacion_dian_requerida",
			"mensaje_confirmacion_requerido": nominaElectronicaConfirmacion,
		})
		return request, false
	}
	return request, true
}

type nominaElectronicaPreflightContext struct {
	mainConfig   map[string]interface{}
	familyConfig *dbpkg.EmpresaDIANDocumentoConfiguracion
	source       *dbpkg.EmpresaNominaDIANFuente
	snapshot     *dbpkg.EmpresaNominaDIANConfiguracionSnapshot
	preflight    map[string]interface{}
}

func loadNominaElectronicaPreflightContext(ctx context.Context, dbEmp *sql.DB, empresaID, liquidacionID int64, requireProduction bool) (*nominaElectronicaPreflightContext, error) {
	if dbEmp == nil || empresaID <= 0 || liquidacionID <= 0 {
		return nil, errors.New("empresa_id y liquidacion_id son obligatorios")
	}
	familyConfig, err := dbpkg.GetEmpresaDIANDocumentoConfiguracionContext(ctx, dbEmp, empresaID, "nomina_electronica")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no existe configuración DIAN separada para nómina electrónica")
		}
		return nil, err
	}
	if err := dbpkg.ValidateEmpresaNominaElectronicaConfigForEmission(*familyConfig); err != nil {
		return nil, err
	}
	if requireProduction && (!strings.EqualFold(strings.TrimSpace(familyConfig.TipoAmbiente), "produccion") || !strings.EqualFold(strings.TrimSpace(familyConfig.Estado), "activo")) {
		return nil, errors.New("la emisión SendNominaSync exige nómina electrónica activa en producción; habilitación debe completarse mediante el set de pruebas DIAN")
	}
	mainConfig, err := getEmpresaDIANConfig(dbEmp, empresaID)
	if err != nil || len(mainConfig) == 0 {
		return nil, errors.New("no existe configuración DIAN principal para firma y transporte")
	}
	runtimeConfig, err := dbpkg.GetFacturacionElectronicaPaisConfig(dbEmp, empresaID, "CO")
	if err != nil || runtimeConfig == nil {
		return nil, errors.New("no existe configuración operativa de facturación electrónica Colombia")
	}
	if requireProduction && (!strings.EqualFold(strings.TrimSpace(runtimeConfig.Ambiente), "produccion") || strings.EqualFold(strings.TrimSpace(runtimeConfig.Estado), "inactivo")) {
		return nil, errors.New("la integración fiscal Colombia debe estar activa en modo producción para usar el transporte real")
	}
	if !requireProduction && strings.EqualFold(strings.TrimSpace(runtimeConfig.Estado), "inactivo") {
		return nil, errors.New("la integración fiscal Colombia está inactiva; no se puede completar el preflight de nómina")
	}
	provider := strings.ToLower(strings.TrimSpace(runtimeConfig.Proveedor))
	if provider != "dian" && !strings.Contains(strings.ToLower(strings.TrimSpace(runtimeConfig.APIBaseURL)), "dian.gov.co") {
		return nil, errors.New("el proveedor fiscal Colombia no está configurado para DIAN real")
	}

	snapshot := &dbpkg.EmpresaNominaDIANConfiguracionSnapshot{
		TipoDocumento: "nomina_electronica", TipoAmbiente: familyConfig.TipoAmbiente,
		ModoOperacionCodigo: familyConfig.ModoOperacionCodigo, TestSetID: familyConfig.TestSetID, Prefijo: strings.ToUpper(strings.TrimSpace(familyConfig.Prefijo)),
		ConsecutivoAsignado: familyConfig.ConsecutivoActual, URLDIANOverride: strings.TrimSpace(familyConfig.URLDIANOverride),
	}
	mergedConfig := nominaElectronicaMergeDIANConfig(mainConfig, snapshot)
	if strings.TrimSpace(snapshot.URLDIANOverride) == "" && strings.TrimSpace(runtimeConfig.APIBaseURL) != "" {
		mergedConfig["url_dian"] = strings.TrimSpace(runtimeConfig.APIBaseURL)
	}
	softwareID, _, _, credentialErr := resolveDIANSoftwareCredentials(mergedConfig, nil, empresaID)
	if credentialErr != nil || strings.TrimSpace(softwareID) == "" {
		return nil, errors.New("Software ID/PIN DIAN de nómina no disponibles")
	}

	currentSource, blockers, err := dbpkg.LoadEmpresaNominaDIANFuenteContext(ctx, dbEmp, empresaID, liquidacionID, softwareID)
	if err != nil {
		return nil, err
	}
	if err := validateNominaElectronicaEmployerDIANConfig(mainConfig, currentSource); err != nil {
		return nil, err
	}
	if !dbpkg.EmpresaNominaDIANPeriodoCerrado(currentSource.PeriodoReporte, time.Now()) {
		return nil, errors.New("el mes de nómina aún no está cerrado; el documento DIAN debe acumular todos los pagos mensuales")
	}
	if len(blockers) > 0 {
		return nil, fmt.Errorf("preflight de nómina electrónica bloqueado: %s", strings.Join(blockers, " | "))
	}
	var source *dbpkg.EmpresaNominaDIANFuente
	existing, existingErr := dbpkg.GetEmpresaNominaElectronicaByEmpleadoPeriodoContext(ctx, dbEmp, empresaID, currentSource.EmpleadoNominaID, currentSource.PeriodoReporte)
	if existingErr == nil && strings.TrimSpace(existing.NumeroLegal) != "" {
		storedSource, storedSnapshot, decodeErr := dbpkg.DecodeEmpresaNominaDIANSnapshots(existing.FuenteFiscalJSON, existing.ConfiguracionDIANJSON)
		if decodeErr != nil {
			return nil, decodeErr
		}
		if storedSource.EmpresaID != empresaID || storedSource.LiquidacionID != existing.LiquidacionID || storedSource.NominaID != existing.ID ||
			storedSource.EmpleadoNominaID != currentSource.EmpleadoNominaID || storedSource.PeriodoReporte != currentSource.PeriodoReporte ||
			!dbpkg.EmpresaNominaDIANFuenteOperacionalCoincide(storedSource, currentSource) {
			return nil, errors.New("fuente fiscal reservada no coincide con la nómina mensual actual; requiere conciliación manual")
		}
		if !strings.EqualFold(storedSnapshot.TipoAmbiente, familyConfig.TipoAmbiente) {
			return nil, errors.New("el ambiente DIAN cambió después de reservar la nómina; requiere conciliación manual")
		}
		source = storedSource
		snapshot = storedSnapshot
		mergedConfig = nominaElectronicaMergeDIANConfig(mainConfig, snapshot)
		if strings.TrimSpace(snapshot.URLDIANOverride) == "" && strings.TrimSpace(runtimeConfig.APIBaseURL) != "" {
			mergedConfig["url_dian"] = strings.TrimSpace(runtimeConfig.APIBaseURL)
		}
	} else if existingErr != nil && !errors.Is(existingErr, sql.ErrNoRows) {
		return nil, existingErr
	} else {
		source = currentSource
		// Represent the number that would be reserved only for source/config
		// validation. Nothing is persisted or consumed in this preflight.
		source.Prefijo = snapshot.Prefijo
		source.Consecutivo = snapshot.ConsecutivoAsignado
		source.NumeroLegal = snapshot.Prefijo + strconv.FormatInt(snapshot.ConsecutivoAsignado, 10)
		source.TipoAmbiente = snapshot.TipoAmbiente
		source.FechaEmisionLegal = time.Now().In(dianColombiaLocation()).Format(time.RFC3339)
	}
	preflight := validateDIANNominaDocumentPreflight(mergedConfig, empresaID, source, "", "reserva")
	if parseTruthy(genericStringValue(preflight["bloqueado"])) {
		return &nominaElectronicaPreflightContext{mainConfig: mergedConfig, familyConfig: familyConfig, source: source, snapshot: snapshot, preflight: preflight}, errors.New("validación preventiva de nómina DIAN no superada")
	}
	return &nominaElectronicaPreflightContext{mainConfig: mergedConfig, familyConfig: familyConfig, source: source, snapshot: snapshot, preflight: preflight}, nil
}

func handleEmitirNominaElectronica(w http.ResponseWriter, r *http.Request, dbEmp, dbSuper *sql.DB) {
	if !requireEmpresaAdditionalModulePermission(w, r, dbEmp, dbSuper, permModuleNominaSueldos, permActionApprove, "linkNominaSueldos") {
		return
	}
	request, ok := decodeNominaElectronicaEmisionRequest(w, r)
	if !ok {
		return
	}
	preflightContext, preflightErr := loadNominaElectronicaPreflightContext(r.Context(), dbEmp, request.EmpresaID, request.LiquidacionID, true)
	if preflightErr != nil {
		response := map[string]interface{}{
			"ok": false, "bloqueado": true, "codigo": "preflight_nomina_electronica", "error": preflightErr.Error(),
		}
		if preflightContext != nil {
			response["preflight"] = preflightContext.preflight
		}
		writeJSON(w, http.StatusUnprocessableEntity, response)
		return
	}
	usuario := strings.TrimSpace(adminEmailFromRequest(r))
	code := nominaElectronicaCodigoInterno(preflightContext.source.LiquidacionID)
	lockContext, release, locked, lockErr := acquireFacturacionDocumentAdvisoryLock(r.Context(), dbEmp, request.EmpresaID, "nomina_electronica", code)
	if lockErr != nil {
		http.Error(w, "No se pudo reservar la nómina electrónica", http.StatusInternalServerError)
		return
	}
	if !locked {
		http.Error(w, "la nómina electrónica ya tiene una emisión en proceso", http.StatusConflict)
		return
	}
	defer release()

	softwareID, _, _, credentialErr := resolveDIANSoftwareCredentials(preflightContext.mainConfig, nil, request.EmpresaID)
	if credentialErr != nil || strings.TrimSpace(softwareID) == "" {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{"ok": false, "bloqueado": true, "codigo": "software_dian_nomina", "error": "Software ID/PIN DIAN de nómina no disponibles"})
		return
	}
	reservation, source, snapshot, reserveErr := dbpkg.ReserveEmpresaNominaElectronicaNumeroContext(lockContext, dbEmp, request.EmpresaID, request.LiquidacionID, softwareID, time.Now(), usuario)
	if reserveErr != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{"ok": false, "bloqueado": true, "codigo": "reserva_numeracion_nomina_electronica", "error": reserveErr.Error()})
		return
	}
	documentState := "pendiente_emision"
	if strings.EqualFold(reservation.EstadoDIAN, "aceptado") {
		documentState = "emitida"
	}
	doc := dbpkg.EmpresaDocumentoFacturacion{
		EmpresaID: request.EmpresaID, TipoDocumento: "nomina_electronica", DocumentoCodigo: code,
		NumeroLegal: reservation.NumeroLegal, PaisCodigo: "CO", AmbienteFE: snapshot.TipoAmbiente,
		EstadoDocumento: documentState, EstadoAnterior: "borrador", EventoUltimo: "nomina_electronica_preparada",
		PeriodoContable: source.PeriodoDesde, MontoTotal: source.ComprobanteTotal, Moneda: "COP",
		FechaDocumento: source.FechaEmisionLegal, EntidadRelacionadaID: source.EmpleadoNominaID,
		UsuarioCreador: usuario, Estado: "activo", Observaciones: "nómina electrónica mensual desde liquidaciones pagadas ids=" + fmt.Sprint(source.LiquidacionIDs),
	}
	persisted, persistErr := dbpkg.UpsertEmpresaDocumentoFacturacionContext(lockContext, dbEmp, doc)
	if persistErr != nil {
		http.Error(w, "No se pudo persistir el documento fiscal de nómina reservado", http.StatusInternalServerError)
		return
	}
	if err := dbpkg.UpdateEmpresaNominaDIANResultContext(lockContext, dbEmp, request.EmpresaID, reservation.NominaID, "preparado", reservation.CUNE, "", true, false); err != nil {
		http.Error(w, "La nómina quedó numerada, pero no se pudo sellar su fuente fiscal; no se transmitió a DIAN", http.StatusInternalServerError)
		return
	}

	payload := facturacionBuildOperacionPayloadFromDocumento(*persisted)
	integration, retry, integrationErr := processFacturacionIntegracionForDocumentoContext(lockContext, dbEmp, payload, *persisted, "emitir_nomina_electronica", usuario, dbSuper)
	if integrationErr != nil {
		http.Error(w, "No se pudo completar la integración DIAN de nómina electrónica", http.StatusInternalServerError)
		return
	}
	warnings := make([]string, 0, 2)
	if refreshed, refreshErr := dbpkg.GetEmpresaDocumentoFacturacionByCodigoContext(lockContext, dbEmp, request.EmpresaID, "nomina_electronica", code); refreshErr == nil && refreshed != nil {
		persisted = refreshed
	} else {
		warnings = append(warnings, "La integración terminó, pero no se pudo refrescar el documento fiscal en la respuesta.")
	}
	payrollMirror, mirrorErr := dbpkg.GetEmpresaNominaElectronicaByLiquidacionContext(lockContext, dbEmp, request.EmpresaID, source.LiquidacionID)
	if mirrorErr != nil {
		warnings = append(warnings, "La integración terminó, pero no se pudo refrescar el espejo de nómina en la respuesta.")
	}
	status := http.StatusAccepted
	if integration.EstadoEnvio == "aceptado" {
		status = http.StatusOK
	}
	writeJSON(w, status, map[string]interface{}{
		"ok":     integration.EstadoEnvio == "aceptado" || integration.EstadoEnvio == "enviado" || integration.EstadoEnvio == "pendiente",
		"accion": "emitir_nomina_electronica", "empresa_id": request.EmpresaID,
		"nomina_electronica": payrollMirror, "documento_fiscal": persisted, "integracion_fiscal": integration,
		"cola_reintentos": retry, "fuente_fiscal_sellada": true, "preflight_reserva": preflightContext.preflight,
		"advertencias_persistencia": warnings,
	})
}
