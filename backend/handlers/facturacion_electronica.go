package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/internal/platform/valueutil"
)

func parseTruthy(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "si", "yes", "activo":
		return true
	default:
		return false
	}
}

func firstPositiveFloat64(values ...float64) float64 {
	return valueutil.FirstPositive(values...)
}

func completarClientePayloadFacturacion(dbEmp *sql.DB, empresaID int64, payload *facturacionOperacionPayload, doc dbpkg.EmpresaDocumentoFacturacion) {
	if dbEmp == nil || payload == nil || empresaID <= 0 {
		return
	}
	clienteID := payload.ClienteID
	if clienteID <= 0 {
		clienteID = payload.EntidadID
	}
	if clienteID <= 0 {
		clienteID = doc.EntidadRelacionadaID
	}
	if clienteID <= 0 {
		return
	}
	cliente, err := dbpkg.GetClienteByID(dbEmp, empresaID, clienteID)
	if err != nil || cliente == nil {
		return
	}
	payload.ClienteID = clienteID
	if payload.EntidadID <= 0 {
		payload.EntidadID = clienteID
	}
	if strings.TrimSpace(payload.ClienteNombre) == "" {
		payload.ClienteNombre = strings.TrimSpace(cliente.NombreRazonSocial)
	}
	if strings.TrimSpace(payload.ClienteNumeroDocumento) == "" {
		payload.ClienteNumeroDocumento = strings.TrimSpace(cliente.NumeroDocumento)
	}
	if strings.TrimSpace(payload.ClienteTipoDocumento) == "" {
		payload.ClienteTipoDocumento = strings.TrimSpace(cliente.TipoDocumento)
	}
	if strings.TrimSpace(payload.ClienteEmail) == "" {
		payload.ClienteEmail = strings.TrimSpace(cliente.Email)
	}
	if strings.TrimSpace(payload.ClienteTelefono) == "" {
		payload.ClienteTelefono = strings.TrimSpace(cliente.Telefono)
	}
	if strings.TrimSpace(payload.ClienteDireccion) == "" {
		payload.ClienteDireccion = strings.TrimSpace(cliente.Direccion)
	}
}

type facturacionOperacionPayload struct {
	EmpresaID                 int64   `json:"empresa_id"`
	EntidadID                 int64   `json:"entidad_id"`
	ClienteID                 int64   `json:"cliente_id"`
	TipoDocumento             string  `json:"tipo_documento"`
	ClienteEmail              string  `json:"cliente_email"`
	ClienteNombre             string  `json:"cliente_nombre"`
	ClienteTipoDocumento      string  `json:"cliente_tipo_documento"`
	ClienteNumeroDocumento    string  `json:"cliente_numero_documento"`
	ClienteTelefono           string  `json:"cliente_telefono"`
	ClienteDireccion          string  `json:"cliente_direccion"`
	PaisCodigo                string  `json:"pais_codigo"`
	DocumentoCodigo           string  `json:"documento_codigo"`
	EstadoActual              string  `json:"estado_actual"`
	FormaPago                 string  `json:"forma_pago"`
	MetodoPago                string  `json:"metodo_pago"`
	Subtotal                  float64 `json:"subtotal"`
	BaseGravable              float64 `json:"base_gravable"`
	IVA                       float64 `json:"iva"`
	Impuestos                 float64 `json:"impuestos"`
	RetencionFuente           float64 `json:"retencion_fuente"`
	RetencionICA              float64 `json:"retencion_ica"`
	RetencionIVA              float64 `json:"retencion_iva"`
	TotalRetenciones          float64 `json:"total_retenciones"`
	TotalNeto                 float64 `json:"total_neto"`
	MontoTotal                float64 `json:"monto_total"`
	Moneda                    string  `json:"moneda"`
	PeriodoContable           string  `json:"periodo_contable"`
	Observaciones             string  `json:"observaciones"`
	PermitirModoOffline       bool    `json:"permitir_modo_offline"`
	ConfirmarModoOffline      bool    `json:"confirmar_modo_offline"`
	OrigenModoOffline         string  `json:"origen_modo_offline"`
	MensajeConfirmacionDIAN   string  `json:"mensaje_confirmacion_dian"`
	ReferenciaDocumentoCodigo string  `json:"referencia_documento_codigo"`
	ReferenciaCUFE            string  `json:"referencia_cufe"`
	ReferenciaFechaEmision    string  `json:"referencia_fecha_emision"`
	CodigoCorreccion          string  `json:"codigo_correccion"`
	DescripcionCorreccion     string  `json:"descripcion_correccion"`
}

const nominaElectronicaReenvioConfirmacion = "REENVIAR NOMINA ELECTRONICA DIAN"

func validateNominaElectronicaManualRetryConfirmation(tipoDocumento, mensaje string) error {
	documentType := normalizeFacturacionDocumentoElectronicoTipo(tipoDocumento)
	if !facturacionDocumentoEsFamiliaNomina(documentType) {
		return nil
	}
	if documentType != "nomina_electronica" {
		return errors.New("la nota de ajuste de nómina electrónica aún no dispone de un adaptador DIAN seguro para retransmisión")
	}
	if strings.TrimSpace(mensaje) != nominaElectronicaReenvioConfirmacion {
		return fmt.Errorf("la nómina electrónica exige la confirmación exacta %q antes de retransmitir", nominaElectronicaReenvioConfirmacion)
	}
	return nil
}

func facturacionDocumentoEsFamiliaNomina(tipoDocumento string) bool {
	switch normalizeFacturacionDocumentoElectronicoTipo(tipoDocumento) {
	case "nomina_electronica", "nota_ajuste_nomina_electronica":
		return true
	default:
		return false
	}
}

// facturacionNominaReadScope keeps payroll data behind both the electronic
// invoicing read permission enforced by the outer route and the payroll read
// permission. Generic listings silently omit payroll when only the second
// permission is missing; an explicit payroll request fails closed.
func facturacionNominaReadScope(w http.ResponseWriter, r *http.Request, dbEmp, dbSuper *sql.DB, requestedType string) (bool, bool) {
	requestedType = strings.TrimSpace(requestedType)
	if requestedType != "" && !facturacionDocumentoEsFamiliaNomina(requestedType) {
		return true, true
	}
	allowed, status, message, err := inspectEmpresaAdditionalModulePermission(r, dbEmp, dbSuper, permModuleNominaSueldos, permActionRead, "linkNominaSueldos")
	if err != nil {
		tenant, _ := TenantContextFromRequest(r)
		log.Printf("[authz] payroll read scope empresa_id=%d error: %v", tenant.EmpresaID, err)
	}
	if allowed {
		return true, true
	}
	if requestedType == "" && status == http.StatusForbidden {
		return false, true
	}
	http.Error(w, message, status)
	return false, false
}

func decodeFacturacionOperacionPayload(reader io.Reader, payload *facturacionOperacionPayload) error {
	if reader == nil || payload == nil {
		return nil
	}
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(payload); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("JSON contiene datos adicionales")
	}
	return nil
}

// facturacionBindAuthorizedEmpresaID makes the tenant selected by the
// authorization wrapper authoritative. A JSON body must never redirect a
// fiscal operation to another company, even when the client omits or lies
// about Content-Type.
func facturacionBindAuthorizedEmpresaID(r *http.Request, payloadEmpresaID *int64) error {
	if r == nil || payloadEmpresaID == nil {
		return fmt.Errorf("empresa_id es obligatorio")
	}
	authorizedID := parseEmpresaIDFromContext(r)
	if authorizedID <= 0 {
		queryID, err := parseInt64QueryOptional(r, "empresa_id")
		if err == nil && queryID > 0 {
			authorizedID = queryID
		}
	}
	if authorizedID <= 0 {
		return fmt.Errorf("empresa_id es obligatorio")
	}
	if *payloadEmpresaID > 0 && *payloadEmpresaID != authorizedID {
		return fmt.Errorf("empresa_id no coincide con el contexto de empresa")
	}
	*payloadEmpresaID = authorizedID
	return nil
}

type facturaEmailResultado struct {
	Intentado             bool   `json:"intentado"`
	Enviado               bool   `json:"enviado"`
	AutomaticoDesactivado bool   `json:"automatico_desactivado,omitempty"`
	Destinatario          string `json:"destinatario,omitempty"`
	ClienteID             int64  `json:"cliente_id,omitempty"`
	OrigenDestinatario    string `json:"origen_destinatario,omitempty"`
	Error                 string `json:"error,omitempty"`
}

type facturacionIntegracionResultado struct {
	Aplica                      bool   `json:"aplica"`
	Accion                      string `json:"accion"`
	PaisCodigo                  string `json:"pais_codigo,omitempty"`
	Proveedor                   string `json:"proveedor,omitempty"`
	Ambiente                    string `json:"ambiente,omitempty"`
	EstadoEnvio                 string `json:"estado_envio"`
	Intentos                    int64  `json:"intentos"`
	MaxIntentos                 int64  `json:"max_intentos"`
	ProximoIntento              string `json:"proximo_intento,omitempty"`
	ContingenciaActiva          bool   `json:"contingencia_activa"`
	ReferenciaExterna           string `json:"referencia_externa,omitempty"`
	Error                       string `json:"error,omitempty"`
	OfflineDisponible           bool   `json:"offline_disponible,omitempty"`
	OfflineConfirmado           bool   `json:"offline_confirmado,omitempty"`
	RequiereConfirmacionOffline bool   `json:"requiere_confirmacion_offline,omitempty"`
	ConexionEstado              string `json:"conexion_estado,omitempty"`
	ConexionMensaje             string `json:"conexion_mensaje,omitempty"`
	AccionRecomendada           string `json:"accion_recomendada,omitempty"`
	Advertencia                 string `json:"advertencia,omitempty"`
}

type facturacionProveedorDispatchResult struct {
	Success             bool
	Pending             bool
	FinalFailure        bool
	ReferenciaExterna   string
	RespuestaJSON       string
	ArtifactWarning     string
	Error               string
	ConnectivityFailure bool
	HTTPStatus          int
}

func saveFacturacionFiscalArtifact(ctx context.Context, dbEmp *sql.DB, doc dbpkg.EmpresaDocumentoFacturacion, tipoArtefacto, extension, mimeType string, content []byte) (*dbpkg.EmpresaFacturacionArtefacto, error) {
	if len(content) == 0 {
		return nil, fmt.Errorf("contenido fiscal vacio")
	}
	old, oldErr := dbpkg.GetEmpresaFacturacionArtefactoByTypeContext(ctx, dbEmp, doc.EmpresaID, doc.TipoDocumento, doc.DocumentoCodigo, tipoArtefacto)
	if oldErr != nil && !errors.Is(oldErr, sql.ErrNoRows) {
		return nil, oldErr
	}
	name, path, written, err := saveEmpresaPrivateUpload(doc.EmpresaID, "facturacion_electronica", extension, bytes.NewReader(content), 50<<20)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(content)
	item, err := dbpkg.UpsertEmpresaFacturacionArtefactoContext(ctx, dbEmp, dbpkg.EmpresaFacturacionArtefacto{
		EmpresaID: doc.EmpresaID, TipoDocumento: doc.TipoDocumento, DocumentoCodigo: doc.DocumentoCodigo,
		TipoArtefacto: tipoArtefacto, StorageRef: name, SHA256: hex.EncodeToString(hash[:]), MimeType: mimeType,
		TamanoBytes: written, Estado: "activo",
	})
	if err != nil {
		if removeErr := os.Remove(path); removeErr != nil {
			log.Printf("warning: no se pudo retirar artefacto fiscal huerfano empresa_id=%d", doc.EmpresaID)
		}
		return nil, err
	}
	if old != nil && old.StorageRef != "" && old.StorageRef != name {
		if oldPath, resolveErr := resolveEmpresaPrivateFile(doc.EmpresaID, "facturacion_electronica", old.StorageRef); resolveErr == nil {
			if removeErr := os.Remove(oldPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
				log.Printf("warning: no se pudo retirar version previa de artefacto fiscal empresa_id=%d", doc.EmpresaID)
			}
		}
	}
	return item, nil
}

func loadFacturacionFiscalArtifact(ctx context.Context, dbEmp *sql.DB, doc dbpkg.EmpresaDocumentoFacturacion, tipoArtefacto string) ([]byte, error) {
	item, err := dbpkg.GetEmpresaFacturacionArtefactoByTypeContext(ctx, dbEmp, doc.EmpresaID, doc.TipoDocumento, doc.DocumentoCodigo, tipoArtefacto)
	if err != nil {
		return nil, err
	}
	return loadFacturacionFiscalArtifactItem(item)
}

func loadFacturacionFiscalArtifactItem(item *dbpkg.EmpresaFacturacionArtefacto) ([]byte, error) {
	if item == nil || item.EmpresaID <= 0 || strings.TrimSpace(item.StorageRef) == "" {
		return nil, fmt.Errorf("metadatos del artefacto fiscal incompletos")
	}
	path, err := resolveEmpresaPrivateFile(item.EmpresaID, "facturacion_electronica", item.StorageRef)
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(path) // #nosec G304 -- metadata is tenant-scoped and resolved below the private root.
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(content)
	if !strings.EqualFold(hex.EncodeToString(hash[:]), item.SHA256) {
		return nil, fmt.Errorf("integridad SHA-256 del artefacto fiscal no coincide")
	}
	return content, nil
}

func facturacionSafeDispatchJSON(response map[string]interface{}) string {
	safe, _ := facturacionSanitizeDispatchValue(response, 0).(map[string]interface{})
	raw, err := json.Marshal(safe)
	if err != nil {
		return `{"ok":false,"error":"respuesta fiscal no serializable"}`
	}
	return string(raw)
}

func facturacionSanitizeDispatchValue(value interface{}, depth int) interface{} {
	if depth > 10 {
		return "[omitido_por_profundidad]"
	}
	switch typed := value.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(typed))
		for key, child := range typed {
			lowerKey := strings.ToLower(strings.TrimSpace(key))
			if lowerKey == "" || lowerKey == "raw_response" || lowerKey == "raw_xml" || lowerKey == "request_resumen" || lowerKey == "xml_firmado" || lowerKey == "xml_ubl_base" || lowerKey == "private_key_pem" || strings.Contains(lowerKey, "password") || strings.Contains(lowerKey, "certificado_clave") || strings.Contains(lowerKey, "software_pin") || strings.Contains(lowerKey, "token") || strings.Contains(lowerKey, "llave_tecnica") {
				continue
			}
			out[key] = facturacionSanitizeDispatchValue(child, depth+1)
		}
		return out
	case []interface{}:
		out := make([]interface{}, 0, len(typed))
		for _, child := range typed {
			out = append(out, facturacionSanitizeDispatchValue(child, depth+1))
		}
		return out
	case string:
		return facturacionTruncate(typed, 4000)
	default:
		return value
	}
}

func buildFacturaElectronicaRepresentationPDF(doc dbpkg.EmpresaDocumentoFacturacion, payload facturacionOperacionPayload) []byte {
	var content bytes.Buffer
	pdfText(&content, "F2", 18, 54, 790, "REPRESENTACION GRAFICA - DOCUMENTO ELECTRONICO")
	pdfLine(&content, "q 0 0 0 RG 1 w 46 770 m 548 770 l S Q")
	rows := []string{
		"Tipo: " + strings.ReplaceAll(strings.ToUpper(doc.TipoDocumento), "_", " "),
		"Numero legal: " + facturacionFirstNonBlank(doc.NumeroLegal, doc.DocumentoCodigo),
		"Documento interno: " + doc.DocumentoCodigo,
		"Fecha fiscal: " + doc.FechaDocumento,
		"Cliente: " + facturacionFirstNonBlank(payload.ClienteNombre, "No registrado"),
		"Identificacion cliente: " + facturacionFirstNonBlank(payload.ClienteNumeroDocumento, "No registrada"),
		fmt.Sprintf("Total: %.2f %s", doc.MontoTotal, facturacionFirstNonBlank(doc.Moneda, "COP")),
		"CUFE/CUDE/CUDS: " + facturacionFirstNonBlank(doc.CodigoValidacion, "Pendiente de acuse"),
		"Estado fiscal: aceptado/enviado segun acuse conservado por PCS",
	}
	y := 735
	for _, row := range rows {
		for _, line := range wrapPDFText(row, 88) {
			pdfText(&content, "F1", 10, 58, y, line)
			y -= 16
		}
		y -= 4
	}
	pdfLine(&content, "q 0 0 0 RG 0.8 w 46 52 m 548 52 l S Q")
	pdfText(&content, "F1", 8, 54, 38, "La fuente fiscal autentica es el XML firmado y el acuse del proveedor conservados por empresa.")
	return assembleSimplePDF(content.Bytes())
}

type facturacionDianOfflineSettings struct {
	Enabled           bool   `json:"modo_offline_dian_activo"`
	AskBeforeContinue bool   `json:"modo_offline_preguntar"`
	AutoRetry         bool   `json:"modo_offline_auto_reintentar"`
	ContingencyType   string `json:"dian_contingencia_tipo"`
}

func isISODateYYYYMMDD(v string) bool {
	v = strings.TrimSpace(v)
	if len(v) != 10 {
		return false
	}
	for i := 0; i < len(v); i += 1 {
		if i == 4 || i == 7 {
			if v[i] != '-' {
				return false
			}
			continue
		}
		if v[i] < '0' || v[i] > '9' {
			return false
		}
	}
	return true
}

// EmpresaFacturacionElectronicaHandler gestiona configuración FE por empresa y país.
func EmpresaFacturacionElectronicaHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "configuracion_documentos_dian" {
				if err := dbpkg.EmpresaDIANDocumentosConfiguracionSchemaReady(dbEmp); err != nil {
					http.Error(w, "La configuracion DIAN por documento aun no esta disponible", http.StatusServiceUnavailable)
					return
				}
				items, err := dbpkg.ListEmpresaDIANDocumentosConfiguracionContext(r.Context(), dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar la configuracion DIAN por documento", http.StatusInternalServerError)
					return
				}
				canReadNomina, proceed := facturacionNominaReadScope(w, r, dbEmp, dbSuper, "")
				if !proceed {
					return
				}
				if !canReadNomina {
					filtered := make([]dbpkg.EmpresaDIANDocumentoConfiguracion, 0, len(items))
					for _, item := range items {
						if !facturacionDocumentoEsFamiliaNomina(item.TipoDocumento) {
							filtered = append(filtered, item)
						}
					}
					items = filtered
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "empresa_id": empresaID, "items": items})
				return
			}
			if action == "artefactos" {
				tipoDocumento := normalizeFacturacionDocumentoElectronicoTipo(r.URL.Query().Get("tipo_documento"))
				if tipoDocumento == "" {
					tipoDocumento = "factura_electronica"
				}
				if _, proceed := facturacionNominaReadScope(w, r, dbEmp, dbSuper, tipoDocumento); !proceed {
					return
				}
				documentoCodigo := strings.TrimSpace(r.URL.Query().Get("documento_codigo"))
				items, err := dbpkg.ListEmpresaFacturacionArtefactosContext(r.Context(), dbEmp, empresaID, tipoDocumento, documentoCodigo)
				if err != nil {
					http.Error(w, "No se pudieron consultar los artefactos fiscales", http.StatusBadRequest)
					return
				}
				out := make([]map[string]interface{}, 0, len(items))
				for _, item := range items {
					out = append(out, map[string]interface{}{
						"id": item.ID, "empresa_id": item.EmpresaID, "tipo_documento": item.TipoDocumento,
						"documento_codigo": item.DocumentoCodigo, "tipo_artefacto": item.TipoArtefacto,
						"sha256": item.SHA256, "mime_type": item.MimeType, "tamano_bytes": item.TamanoBytes,
						"fecha_actualizacion": item.FechaActualizacion,
						"download_url":        fmt.Sprintf("/api/empresa/facturacion_electronica?empresa_id=%d&action=descargar_artefacto&id=%d", empresaID, item.ID),
					})
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "empresa_id": empresaID, "items": out})
				return
			}
			if action == "descargar_artefacto" {
				id, err := parseInt64QueryOptional(r, "id")
				if err != nil || id <= 0 {
					http.Error(w, "id de artefacto invalido", http.StatusBadRequest)
					return
				}
				item, err := dbpkg.GetEmpresaFacturacionArtefactoByIDContext(r.Context(), dbEmp, empresaID, id)
				if err != nil {
					http.Error(w, "artefacto fiscal no disponible", http.StatusNotFound)
					return
				}
				if _, proceed := facturacionNominaReadScope(w, r, dbEmp, dbSuper, item.TipoDocumento); !proceed {
					return
				}
				ext := ".bin"
				switch item.TipoArtefacto {
				case "xml_firmado":
					ext = ".xml"
				case "respuesta_proveedor":
					ext = ".json"
				case "representacion_pdf":
					ext = ".pdf"
				}
				name := facturaElectronicaAttachmentBaseName(item.DocumentoCodigo, item.DocumentoCodigo) + "-" + item.TipoArtefacto + ext
				content, err := loadFacturacionFiscalArtifactItem(item)
				if err != nil {
					http.Error(w, "artefacto fiscal no disponible o con integridad invalida", http.StatusUnprocessableEntity)
					return
				}
				w.Header().Set("Cache-Control", "no-store")
				w.Header().Set("X-Content-Type-Options", "nosniff")
				w.Header().Set("Content-Type", item.MimeType)
				w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name))
				http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(content))
				return
			}
			if action == "documentos" {
				tipoDocumento := strings.TrimSpace(r.URL.Query().Get("tipo_documento"))
				canReadNomina, proceed := facturacionNominaReadScope(w, r, dbEmp, dbSuper, tipoDocumento)
				if !proceed {
					return
				}
				limit, err := parseIntQueryOptional(r, "limit")
				if err != nil {
					http.Error(w, "limit invalido", http.StatusBadRequest)
					return
				}
				offset, err := parseIntQueryOptional(r, "offset")
				if err != nil {
					http.Error(w, "offset invalido", http.StatusBadRequest)
					return
				}

				fechaDesde := strings.TrimSpace(r.URL.Query().Get("fecha_desde"))
				if fechaDesde != "" && !isISODateYYYYMMDD(fechaDesde) {
					http.Error(w, "fecha_desde invalida (use YYYY-MM-DD)", http.StatusBadRequest)
					return
				}
				fechaHasta := strings.TrimSpace(r.URL.Query().Get("fecha_hasta"))
				if fechaHasta != "" && !isISODateYYYYMMDD(fechaHasta) {
					http.Error(w, "fecha_hasta invalida (use YYYY-MM-DD)", http.StatusBadRequest)
					return
				}
				cajeroQuery := strings.TrimSpace(r.URL.Query().Get("cajero"))
				if normalizePermissionRole(effectiveAdminRoleFromRequest(r)) == "cajero" {
					cajeroQuery = strings.TrimSpace(adminEmailFromRequest(r))
				}

				excludedTypes := make([]string, 0, 2)
				if tipoDocumento == "" && !canReadNomina {
					excludedTypes = append(excludedTypes, "nomina_electronica", "nota_ajuste_nomina_electronica")
				}
				items, err := dbpkg.ListEmpresaDocumentosFacturacionByEmpresaContext(r.Context(), dbEmp, dbpkg.EmpresaDocumentoFacturacionListFilter{
					EmpresaID:             empresaID,
					TipoDocumento:         tipoDocumento,
					ExcluirTiposDocumento: excludedTypes,
					EstadoDocumento:       strings.TrimSpace(r.URL.Query().Get("estado_documento")),
					IncludeInactive:       parseTruthy(r.URL.Query().Get("include_inactive")) || parseTruthy(r.URL.Query().Get("incluir_inactivas")),
					ClienteQuery:          strings.TrimSpace(r.URL.Query().Get("cliente")),
					DocumentoQuery:        strings.TrimSpace(r.URL.Query().Get("documento")),
					CajeroQuery:           cajeroQuery,
					FechaDesde:            fechaDesde,
					FechaHasta:            fechaHasta,
					Query:                 strings.TrimSpace(r.URL.Query().Get("q")),
					Limit:                 limit,
					Offset:                offset,
				})
				if err != nil {
					http.Error(w, "No se pudo listar documentos de facturacion", http.StatusInternalServerError)
					return
				}
				fuentes, err := dbpkg.ListEmpresaFacturacionFuenteFiscalRefsContext(r.Context(), dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo verificar la trazabilidad fiscal de los documentos", http.StatusInternalServerError)
					return
				}
				facturacionMarcarDisponibilidadFuenteFiscal(items, fuentes)

				writeJSON(w, http.StatusOK, map[string]interface{}{
					"empresa_id": empresaID,
					"items":      items,
				})
				return
			}

			if action == "reintentos" {
				tipoDocumento := strings.TrimSpace(r.URL.Query().Get("tipo_documento"))
				canReadNomina, proceed := facturacionNominaReadScope(w, r, dbEmp, dbSuper, tipoDocumento)
				if !proceed {
					return
				}
				limit, err := parseIntQueryOptional(r, "limit")
				if err != nil {
					http.Error(w, "limit invalido", http.StatusBadRequest)
					return
				}
				offset, err := parseIntQueryOptional(r, "offset")
				if err != nil {
					http.Error(w, "offset invalido", http.StatusBadRequest)
					return
				}

				excludedTypes := make([]string, 0, 2)
				if tipoDocumento == "" && !canReadNomina {
					excludedTypes = append(excludedTypes, "nomina_electronica", "nota_ajuste_nomina_electronica")
				}
				items, err := dbpkg.ListFacturacionElectronicaRetriesByEmpresaContext(r.Context(), dbEmp, empresaID, dbpkg.FacturacionElectronicaRetryFilter{
					TipoDocumento:         tipoDocumento,
					ExcluirTiposDocumento: excludedTypes,
					EstadoEnvio:           strings.TrimSpace(r.URL.Query().Get("estado_envio")),
					DocumentoQuery:        strings.TrimSpace(comprasFirstNonBlank(r.URL.Query().Get("q"), r.URL.Query().Get("documento"))),
					SoloVencidos:          parseTruthy(comprasFirstNonBlank(r.URL.Query().Get("solo_vencidos"), r.URL.Query().Get("vencidos"))),
					IncludeInactive:       parseTruthy(r.URL.Query().Get("include_inactive")) || parseTruthy(r.URL.Query().Get("incluir_inactivas")),
					Limit:                 limit,
					Offset:                offset,
				})
				if err != nil {
					http.Error(w, "No se pudo listar cola de reintentos FE", http.StatusInternalServerError)
					return
				}

				writeJSON(w, http.StatusOK, map[string]interface{}{
					"empresa_id": empresaID,
					"items":      items,
				})
				return
			}

			if action == "reconciliacion" || action == "reconciliar_estados" {
				resumen, err := buildFacturacionReconciliacion(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo calcular reconciliacion FE", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, resumen)
				return
			}

			if action == "estado_conexion_dian" || action == "estado_conexion" {
				paisCodigo := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("pais_codigo")))
				if paisCodigo == "" {
					paisCodigo = "CO"
				}
				cfg, err := dbpkg.GetFacturacionElectronicaPaisConfigContext(r.Context(), dbEmp, empresaID, paisCodigo)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "No se pudo consultar conectividad DIAN", http.StatusInternalServerError)
					return
				}
				status := facturacionDIANConnectionStatus(dbEmp, empresaID, paisCodigo, cfg)
				// Consultar conectividad nunca debe transmitir documentos. Los
				// reintentos automaticos pertenecen al worker durable y el reintento
				// manual conserva un POST autorizado separado.
				status["reintentos_automaticos"] = "pcs-worker"
				writeJSON(w, http.StatusOK, status)
				return
			}

			if action == "catalogo_dian_colombia" || action == "documentos_dian_colombia" {
				cfg, err := dbpkg.GetFacturacionElectronicaPaisConfigContext(r.Context(), dbEmp, empresaID, "CO")
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "No se pudo consultar catalogo DIAN Colombia", http.StatusInternalServerError)
					return
				}
				extra := map[string]interface{}{}
				if cfg != nil {
					extra = facturacionTryParseJSONMap(cfg.CamposPaisJSON)
				}
				documentosActivos := facturacionStringListFromAny(extra["documentos_soportados"])
				documentosActivos = facturacionFiltrarDocumentosDianOperativos(documentosActivos)
				if len(documentosActivos) == 0 {
					documentosActivos = dbpkg.DefaultFacturacionDianDocumentosSoportados()
				}
				obligacionesActivas := facturacionStringListFromAny(extra["documentos_contadores_colombia"])
				if len(obligacionesActivas) == 0 {
					obligacionesActivas = []string{"declaraciones_tributarias", "informacion_exogena", "certificados_retencion", "conciliacion_fiscal"}
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":                             true,
					"empresa_id":                     empresaID,
					"pais_codigo":                    "CO",
					"documentos":                     dbpkg.ListFacturacionDianDocumentosElectronicos(),
					"documentos_soportados":          documentosActivos,
					"obligaciones_contador":          dbpkg.ListFacturacionDianObligacionesContadores(),
					"documentos_contadores_colombia": obligacionesActivas,
					"fuentes":                        dbpkg.ListFacturacionDianFuentesNormativas(),
					"nota":                           "El catalogo separa documentos electronicos del SFE y obligaciones contables/tributarias que preparan contadores.",
				})
				return
			}

			paisCodigo := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("pais_codigo")))
			incluirInactivas := parseTruthy(r.URL.Query().Get("incluir_inactivas"))

			if paisCodigo != "" {
				cfg, err := dbpkg.GetFacturacionElectronicaPaisConfigContext(r.Context(), dbEmp, empresaID, paisCodigo)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "No se pudo consultar la configuración de facturación electrónica", http.StatusInternalServerError)
					return
				}
				if cfg == nil {
					http.Error(w, "No se pudo resolver la configuración", http.StatusInternalServerError)
					return
				}
				// A requested profile must be returned exactly as requested. Falling
				// back to browser/company detection here used to turn a new country
				// into Colombia before its configuration could be saved.
				writeJSON(w, http.StatusOK, cfg)
				return
			}

			items, err := dbpkg.ListFacturacionElectronicaPaisConfigsContext(r.Context(), dbEmp, empresaID, incluirInactivas)
			if err != nil {
				http.Error(w, "No se pudo listar la configuración de facturación electrónica", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"empresa_id": empresaID,
				"items":      items,
			})
			return

		case http.MethodPost, http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "emitir_documento_soporte" {
				handleEmitirDocumentoSoporte(w, r, dbEmp, dbSuper)
				return
			}
			if action == "emitir_nomina_electronica" {
				handleEmitirNominaElectronica(w, r, dbEmp, dbSuper)
				return
			}
			if action == "configuracion_documentos_dian" {
				empresaID, err := parseEmpresaIDQuery(r)
				if err != nil {
					http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
					return
				}
				if err := dbpkg.EmpresaDIANDocumentosConfiguracionSchemaReady(dbEmp); err != nil {
					http.Error(w, "La configuracion DIAN por documento aun no esta disponible", http.StatusServiceUnavailable)
					return
				}
				var payload dbpkg.EmpresaDIANDocumentoConfiguracion
				decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
				decoder.DisallowUnknownFields()
				if err := decoder.Decode(&payload); err != nil {
					http.Error(w, "Configuracion DIAN por documento invalida", http.StatusBadRequest)
					return
				}
				if err := decoder.Decode(&struct{}{}); err != io.EOF {
					http.Error(w, "Configuracion DIAN por documento invalida", http.StatusBadRequest)
					return
				}
				if err := facturacionValidarConfiguracionDIANDocumento(payload.TipoDocumento, payload.Estado); err != nil {
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				normalizedDocumentType := normalizeFacturacionDocumentoElectronicoTipo(payload.TipoDocumento)
				if facturacionDocumentoEsFamiliaNomina(normalizedDocumentType) && !requireEmpresaAdditionalModulePermission(w, r, dbEmp, dbSuper, permModuleNominaSueldos, permActionApprove, "linkNominaSueldos") {
					return
				}
				if normalizedDocumentType == "documento_soporte" || normalizedDocumentType == "nomina_electronica" {
					existing, existingErr := dbpkg.GetEmpresaDIANDocumentoConfiguracionContext(r.Context(), dbEmp, empresaID, normalizedDocumentType)
					if existingErr != nil && !errors.Is(existingErr, sql.ErrNoRows) {
						http.Error(w, "No se pudo comprobar el consecutivo DIAN vigente", http.StatusInternalServerError)
						return
					}
					if existing != nil && existing.ConsecutivoActual > payload.ConsecutivoActual {
						payload.ConsecutivoActual = existing.ConsecutivoActual
					}
					switch strings.ToLower(strings.TrimSpace(payload.Estado)) {
					case "habilitacion", "activo":
						if normalizedDocumentType == "nomina_electronica" {
							if err := dbpkg.ValidateEmpresaNominaElectronicaConfigForEmission(payload); err != nil {
								http.Error(w, err.Error(), http.StatusConflict)
								return
							}
							break
						}
						if err := dbpkg.ValidateEmpresaDocumentoSoporteConfigForEmission(payload, time.Now()); err != nil {
							http.Error(w, err.Error(), http.StatusConflict)
							return
						}
					}
				}
				id, err := dbpkg.UpsertEmpresaDIANDocumentoConfiguracionContext(r.Context(), dbEmp, payload)
				if err != nil {
					http.Error(w, "No se pudo guardar la configuracion DIAN por documento", http.StatusBadRequest)
					return
				}
				item, err := dbpkg.GetEmpresaDIANDocumentoConfiguracionContext(r.Context(), dbEmp, empresaID, payload.TipoDocumento)
				if err != nil {
					http.Error(w, "La configuracion DIAN fue guardada pero no se pudo consultar", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id, "empresa_id": empresaID, "item": item})
				return
			}
			if action == "procesar_reintentos" {
				empresaID, err := parseInt64QueryOptional(r, "empresa_id")
				if err != nil {
					http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
					return
				}
				limit, err := parseIntQueryOptional(r, "limit")
				if err != nil {
					http.Error(w, "limit invalido", http.StatusBadRequest)
					return
				}
				processed, err := processFacturacionRetryQueueContextWithScope(r.Context(), dbEmp, dbSuper, empresaID, limit, strings.TrimSpace(adminEmailFromRequest(r)), false)
				if err != nil {
					http.Error(w, "No se pudo procesar cola de reintentos FE", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, processed)
				return
			}

			if action == "reconciliar_estados" || action == "reconciliar_aceptados_local" {
				empresaID, err := parseInt64QueryOptional(r, "empresa_id")
				if err != nil {
					http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
					return
				}
				aplicar := parseTruthy(comprasFirstNonBlank(r.URL.Query().Get("aplicar"), r.URL.Query().Get("sync"), r.URL.Query().Get("apply")))
				soloAcusesAceptados := action == "reconciliar_aceptados_local" || parseTruthy(r.URL.Query().Get("solo_acuses_aceptados"))
				resumen, err := reconcileFacturacionEstados(dbEmp, empresaID, aplicar, soloAcusesAceptados, strings.TrimSpace(adminEmailFromRequest(r)))
				if err != nil {
					http.Error(w, "No se pudo reconciliar estados FE", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, resumen)
				return
			}

			if action == "facturar_desde_venta" {
				var payload facturacionOperacionPayload
				if r.Body != nil {
					if err := decodeFacturacionOperacionPayload(http.MaxBytesReader(w, r.Body, 64<<10), &payload); err != nil {
						http.Error(w, "JSON de operación de facturación inválido", http.StatusBadRequest)
						return
					}
				}
				if err := facturacionBindAuthorizedEmpresaID(r, &payload.EmpresaID); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if strings.TrimSpace(payload.DocumentoCodigo) == "" {
					payload.DocumentoCodigo = strings.TrimSpace(r.URL.Query().Get("documento_codigo"))
				}
				if strings.TrimSpace(payload.DocumentoCodigo) == "" {
					http.Error(w, "documento_codigo es obligatorio", http.StatusBadRequest)
					return
				}
				if strings.TrimSpace(payload.TipoDocumento) == "" {
					payload.TipoDocumento = strings.TrimSpace(r.URL.Query().Get("tipo_documento"))
				}
				if strings.TrimSpace(payload.TipoDocumento) == "" {
					payload.TipoDocumento = "comprobante_pago"
				}

				ventaDoc, err := dbpkg.GetEmpresaDocumentoFacturacionByCodigoContext(r.Context(), dbEmp, payload.EmpresaID, payload.TipoDocumento, payload.DocumentoCodigo)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "venta no encontrada", http.StatusNotFound)
						return
					}
					http.Error(w, "No se pudo consultar la venta", http.StatusInternalServerError)
					return
				}
				if strings.TrimSpace(strings.ToLower(ventaDoc.TipoDocumento)) != "comprobante_pago" {
					http.Error(w, "solo se puede facturar electronicamente una venta/comprobante", http.StatusConflict)
					return
				}
				if strings.TrimSpace(strings.ToLower(ventaDoc.EstadoDocumento)) != "emitida" {
					http.Error(w, "la venta debe estar emitida para generar la factura electronica", http.StatusConflict)
					return
				}
				clienteID := payload.ClienteID
				if clienteID <= 0 && payload.EntidadID > 0 {
					clienteID = payload.EntidadID
				}
				if clienteID <= 0 && strings.TrimSpace(payload.ClienteNombre) != "" && strings.TrimSpace(payload.ClienteNumeroDocumento) != "" {
					if email := strings.TrimSpace(payload.ClienteEmail); email != "" {
						if _, err := mail.ParseAddress(email); err != nil {
							http.Error(w, "cliente_email invalido", http.StatusBadRequest)
							return
						}
					}
					tipoDocCliente := strings.ToUpper(strings.TrimSpace(payload.ClienteTipoDocumento))
					if tipoDocCliente == "" {
						tipoDocCliente = "CC"
					}
					newClienteID, err := dbpkg.CreateCliente(dbEmp, dbpkg.Cliente{
						EmpresaID:         payload.EmpresaID,
						TipoDocumento:     tipoDocCliente,
						NumeroDocumento:   strings.TrimSpace(payload.ClienteNumeroDocumento),
						TipoPersona:       "natural",
						NombreRazonSocial: strings.TrimSpace(payload.ClienteNombre),
						Email:             strings.TrimSpace(payload.ClienteEmail),
						Telefono:          strings.TrimSpace(payload.ClienteTelefono),
						Direccion:         strings.TrimSpace(payload.ClienteDireccion),
						Pais:              "CO",
						UsuarioCreador:    strings.TrimSpace(adminEmailFromRequest(r)),
						Observaciones:     "creado desde facturacion electronica de venta",
					})
					if err != nil {
						http.Error(w, "No se pudo crear el cliente para facturar la venta: "+err.Error(), http.StatusBadRequest)
						return
					}
					clienteID = newClienteID
				}
				if clienteID > 0 {
					if _, err := dbpkg.GetClienteByID(dbEmp, payload.EmpresaID, clienteID); err != nil {
						if errors.Is(err, sql.ErrNoRows) {
							http.Error(w, "cliente_id no pertenece a esta empresa", http.StatusNotFound)
							return
						}
						http.Error(w, "No se pudo validar el cliente", http.StatusInternalServerError)
						return
					}
					updatedVentaDoc, err := dbpkg.UpdateEmpresaDocumentoFacturacionClienteContext(r.Context(), dbEmp, payload.EmpresaID, ventaDoc.TipoDocumento, ventaDoc.DocumentoCodigo, clienteID)
					if err != nil {
						http.Error(w, "No se pudo asociar el cliente a la venta", http.StatusInternalServerError)
						return
					}
					ventaDoc = updatedVentaDoc
				}

				resultado, err := registrarFacturaElectronicaDesdeDocumentoVentaContext(
					r.Context(),
					dbEmp,
					dbSuper,
					ventaDoc,
					strings.TrimSpace(adminEmailFromRequest(r)),
					"factura electronica generada manualmente desde la bandeja de ventas",
				)
				if err != nil {
					http.Error(w, "No se pudo generar la factura electronica desde la venta", http.StatusInternalServerError)
					return
				}

				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":               true,
					"accion":           "facturar_desde_venta",
					"empresa_id":       payload.EmpresaID,
					"venta_origen":     ventaDoc,
					"factura_generada": resultado,
				})
				return
			}

			if action == "reenviar_correo" || action == "enviar_correo" {
				var payload facturacionOperacionPayload
				if r.Body != nil {
					if err := decodeFacturacionOperacionPayload(http.MaxBytesReader(w, r.Body, 64<<10), &payload); err != nil {
						http.Error(w, "JSON de operación de facturación inválido", http.StatusBadRequest)
						return
					}
				}
				if err := facturacionBindAuthorizedEmpresaID(r, &payload.EmpresaID); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if strings.TrimSpace(payload.DocumentoCodigo) == "" {
					payload.DocumentoCodigo = strings.TrimSpace(r.URL.Query().Get("documento_codigo"))
				}
				if strings.TrimSpace(payload.DocumentoCodigo) == "" {
					http.Error(w, "documento_codigo es obligatorio", http.StatusBadRequest)
					return
				}
				if strings.TrimSpace(payload.TipoDocumento) == "" {
					payload.TipoDocumento = strings.TrimSpace(r.URL.Query().Get("tipo_documento"))
				}
				if facturacionDocumentoEsFamiliaNomina(payload.TipoDocumento) {
					if !requireEmpresaAdditionalModulePermission(w, r, dbEmp, dbSuper, permModuleNominaSueldos, permActionApprove, "linkNominaSueldos") {
						return
					}
					writeJSON(w, http.StatusConflict, map[string]interface{}{
						"ok": false, "bloqueado": true, "codigo": "distribucion_nomina_dedicada_requerida",
						"error": "la nómina electrónica no puede enviarse con la representación gráfica de una factura; use el flujo dedicado de nómina",
					})
					return
				}
				doc, err := dbpkg.GetEmpresaDocumentoFacturacionByCodigoContext(r.Context(), dbEmp, payload.EmpresaID, payload.TipoDocumento, payload.DocumentoCodigo)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "documento no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, "No se pudo consultar el documento", http.StatusInternalServerError)
					return
				}
				if payload.ClienteID <= 0 && payload.EntidadID <= 0 && doc.EntidadRelacionadaID > 0 {
					payload.ClienteID = doc.EntidadRelacionadaID
					payload.EntidadID = doc.EntidadRelacionadaID
				}
				if strings.EqualFold(doc.PaisCodigo, "CO") && strings.EqualFold(doc.AmbienteFE, "produccion") {
					retry, retryErr := dbpkg.GetFacturacionElectronicaRetryByDocumentoContext(r.Context(), dbEmp, payload.EmpresaID, doc.TipoDocumento, doc.DocumentoCodigo)
					if retryErr != nil || retry == nil || normalizeFacturacionEstadoEnvio(retry.EstadoEnvio) != "aceptado" {
						http.Error(w, "el correo fiscal solo puede enviarse despues de la aceptacion DIAN", http.StatusConflict)
						return
					}
				}

				resultado := enviarFacturaElectronicaAlCliente(dbEmp, dbSuper, payload, *doc)
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":               true,
					"accion":           "reenviar_correo",
					"empresa_id":       payload.EmpresaID,
					"tipo_documento":   doc.TipoDocumento,
					"documento_codigo": doc.DocumentoCodigo,
					"factura_email":    resultado,
				})
				return
			}
			if action == "reenviar_dian" || action == "reintentar_dian" || action == "enviar_dian" {
				var payload facturacionOperacionPayload
				if r.Body != nil {
					if err := decodeFacturacionOperacionPayload(http.MaxBytesReader(w, r.Body, 64<<10), &payload); err != nil {
						http.Error(w, "JSON de operación de facturación inválido", http.StatusBadRequest)
						return
					}
				}
				if err := facturacionBindAuthorizedEmpresaID(r, &payload.EmpresaID); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if strings.TrimSpace(payload.DocumentoCodigo) == "" {
					payload.DocumentoCodigo = strings.TrimSpace(r.URL.Query().Get("documento_codigo"))
				}
				if strings.TrimSpace(payload.DocumentoCodigo) == "" {
					http.Error(w, "documento_codigo es obligatorio", http.StatusBadRequest)
					return
				}
				if strings.TrimSpace(payload.TipoDocumento) == "" {
					payload.TipoDocumento = strings.TrimSpace(r.URL.Query().Get("tipo_documento"))
				}
				documentoTipo := normalizeFacturacionDocumentoElectronicoTipo(payload.TipoDocumento)
				if documentoTipo == "" {
					documentoTipo = "factura_electronica"
				}
				doc, err := dbpkg.GetEmpresaDocumentoFacturacionByCodigoContext(r.Context(), dbEmp, payload.EmpresaID, documentoTipo, payload.DocumentoCodigo)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "documento no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, "No se pudo consultar el documento", http.StatusInternalServerError)
					return
				}
				if facturacionDocumentoEsFamiliaNomina(doc.TipoDocumento) {
					if !requireEmpresaAdditionalModulePermission(w, r, dbEmp, dbSuper, permModuleNominaSueldos, permActionApprove, "linkNominaSueldos") {
						return
					}
					if err := validateNominaElectronicaManualRetryConfirmation(doc.TipoDocumento, payload.MensajeConfirmacionDIAN); err != nil {
						writeJSON(w, http.StatusConflict, map[string]interface{}{
							"ok": false, "bloqueado": true, "codigo": "confirmacion_reenvio_nomina_requerida",
							"error": err.Error(), "mensaje_confirmacion_requerido": nominaElectronicaReenvioConfirmacion,
						})
						return
					}
				}
				merged := facturacionBuildOperacionPayloadFromDocumento(*doc)
				if strings.TrimSpace(payload.ClienteNombre) != "" {
					merged.ClienteNombre = payload.ClienteNombre
				}
				if strings.TrimSpace(payload.ClienteNumeroDocumento) != "" {
					merged.ClienteNumeroDocumento = payload.ClienteNumeroDocumento
				}
				if strings.TrimSpace(payload.ClienteTipoDocumento) != "" {
					merged.ClienteTipoDocumento = payload.ClienteTipoDocumento
				}
				if strings.TrimSpace(payload.ClienteEmail) != "" {
					merged.ClienteEmail = payload.ClienteEmail
				}
				if strings.TrimSpace(payload.ClienteTelefono) != "" {
					merged.ClienteTelefono = payload.ClienteTelefono
				}
				if strings.TrimSpace(payload.ClienteDireccion) != "" {
					merged.ClienteDireccion = payload.ClienteDireccion
				}
				manualRetryCtx := context.WithValue(r.Context(), facturacionManualTerminalRetryContextKey{}, true)
				resultado, retryItem, err := processFacturacionIntegracionForDocumentoContext(manualRetryCtx, dbEmp, merged, *doc, "emitir", strings.TrimSpace(adminEmailFromRequest(r)), dbSuper)
				if err != nil {
					http.Error(w, "No se pudo reintentar envio DIAN", http.StatusInternalServerError)
					return
				}
				var facturaAnulada *dbpkg.EmpresaDocumentoFacturacion
				if facturacionIntegracionAceptada(resultado) && strings.EqualFold(strings.TrimSpace(doc.TipoDocumento), "nota_credito") {
					facturaAnulada, err = finalizarFacturaAnuladaPorNotaCredito(dbEmp, *doc, strings.TrimSpace(adminEmailFromRequest(r)))
					if err != nil {
						http.Error(w, "La nota credito fue aceptada, pero no se pudo finalizar la anulacion local", http.StatusInternalServerError)
						return
					}
				}
				if refreshed, refreshErr := dbpkg.GetEmpresaDocumentoFacturacionByCodigoContext(r.Context(), dbEmp, payload.EmpresaID, documentoTipo, payload.DocumentoCodigo); refreshErr == nil && refreshed != nil {
					doc = refreshed
				}
				resp := map[string]interface{}{
					"ok":                 resultado.EstadoEnvio == "enviado" || resultado.EstadoEnvio == "aceptado" || strings.TrimSpace(resultado.ReferenciaExterna) != "",
					"accion":             action,
					"empresa_id":         payload.EmpresaID,
					"tipo_documento":     doc.TipoDocumento,
					"documento_codigo":   doc.DocumentoCodigo,
					"numero_legal":       doc.NumeroLegal,
					"codigo_validacion":  doc.CodigoValidacion,
					"integracion_fiscal": resultado,
				}
				if facturaAnulada != nil {
					resp["factura_anulada"] = facturaAnulada
					resp["anulacion_confirmada_dian"] = true
				}
				if retryItem != nil {
					resp["cola_reintentos"] = retryItem
				}
				writeJSON(w, http.StatusOK, resp)
				return
			}
			if action == "anular_venta" || action == "anular_comprobante" {
				var payload facturacionOperacionPayload
				if r.Body != nil {
					if err := decodeFacturacionOperacionPayload(http.MaxBytesReader(w, r.Body, 64<<10), &payload); err != nil {
						http.Error(w, "JSON de operación de facturación inválido", http.StatusBadRequest)
						return
					}
				}
				if err := facturacionBindAuthorizedEmpresaID(r, &payload.EmpresaID); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if strings.TrimSpace(payload.DocumentoCodigo) == "" {
					payload.DocumentoCodigo = strings.TrimSpace(r.URL.Query().Get("documento_codigo"))
				}
				if strings.TrimSpace(payload.DocumentoCodigo) == "" {
					http.Error(w, "documento_codigo es obligatorio", http.StatusBadRequest)
					return
				}
				motivo := strings.TrimSpace(payload.Observaciones)
				if motivo == "" {
					motivo = strings.TrimSpace(r.URL.Query().Get("motivo"))
				}
				if len(motivo) < 10 {
					http.Error(w, "motivo de anulacion es obligatorio y debe tener minimo 10 caracteres", http.StatusBadRequest)
					return
				}
				venta, err := dbpkg.GetEmpresaDocumentoFacturacionByCodigoContext(r.Context(), dbEmp, payload.EmpresaID, "comprobante_pago", payload.DocumentoCodigo)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "venta no encontrada", http.StatusNotFound)
						return
					}
					http.Error(w, "No se pudo consultar la venta", http.StatusInternalServerError)
					return
				}
				if strings.ToLower(strings.TrimSpace(venta.EstadoDocumento)) == "anulada" || strings.ToLower(strings.TrimSpace(venta.Estado)) == "inactivo" {
					http.Error(w, "la venta ya esta anulada", http.StatusConflict)
					return
				}
				usuario := strings.TrimSpace(adminEmailFromRequest(r))
				ventaAnulada, err := dbpkg.UpsertEmpresaDocumentoFacturacionContext(r.Context(), dbEmp, dbpkg.EmpresaDocumentoFacturacion{
					EmpresaID:            venta.EmpresaID,
					TipoDocumento:        venta.TipoDocumento,
					DocumentoCodigo:      venta.DocumentoCodigo,
					EstadoDocumento:      "anulada",
					EstadoAnterior:       venta.EstadoDocumento,
					EventoUltimo:         "venta_anulada",
					PeriodoContable:      venta.PeriodoContable,
					MontoTotal:           venta.MontoTotal,
					Moneda:               venta.Moneda,
					NumeroLegal:          venta.NumeroLegal,
					CodigoValidacion:     venta.CodigoValidacion,
					PaisCodigo:           venta.PaisCodigo,
					AmbienteFE:           venta.AmbienteFE,
					FechaDocumento:       venta.FechaDocumento,
					EntidadRelacionadaID: venta.EntidadRelacionadaID,
					UsuarioCreador:       usuario,
					Observaciones:        strings.TrimSpace(venta.Observaciones + "\nVenta anulada. Motivo: " + motivo),
				})
				if err != nil {
					http.Error(w, "No se pudo anular la venta", http.StatusInternalServerError)
					return
				}
				registrarEventoContableNoBloqueante(dbEmp, r, "facturacion", dbpkg.EmpresaEventoContable{
					EmpresaID:       venta.EmpresaID,
					Modulo:          "ventas",
					Evento:          "venta_anulada",
					Entidad:         "comprobante_pago",
					EntidadID:       ventaAnulada.ID,
					DocumentoTipo:   venta.TipoDocumento,
					DocumentoCodigo: venta.DocumentoCodigo,
					PeriodoContable: venta.PeriodoContable,
					MontoTotal:      venta.MontoTotal,
					Moneda:          venta.Moneda,
					Origen:          "api_facturacion_electronica",
					Observaciones:   motivo,
				}, map[string]interface{}{
					"documento_codigo": venta.DocumentoCodigo,
					"empresa_id":       venta.EmpresaID,
				})
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":               true,
					"accion":           "anular_venta",
					"empresa_id":       venta.EmpresaID,
					"venta_original":   ventaAnulada,
					"documento_codigo": venta.DocumentoCodigo,
				})
				return
			}
			if action == "anular_factura_nota_credito" || action == "anular_factura" {
				if !facturacionDocumentoElectronicoDIANComercialSoportado("nota_credito") {
					writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
						"ok":     false,
						"codigo": "nota_credito_dian_no_implementada",
						"error":  "la anulacion fiscal permanece bloqueada hasta contar con fuente inmutable de ajuste y referencia DIAN aceptada",
					})
					return
				}
				var payload facturacionOperacionPayload
				if r.Body != nil {
					if err := decodeFacturacionOperacionPayload(http.MaxBytesReader(w, r.Body, 64<<10), &payload); err != nil {
						http.Error(w, "JSON de operación de facturación inválido", http.StatusBadRequest)
						return
					}
				}
				if err := facturacionBindAuthorizedEmpresaID(r, &payload.EmpresaID); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if strings.TrimSpace(payload.DocumentoCodigo) == "" {
					payload.DocumentoCodigo = strings.TrimSpace(r.URL.Query().Get("documento_codigo"))
				}
				if strings.TrimSpace(payload.DocumentoCodigo) == "" {
					http.Error(w, "documento_codigo es obligatorio", http.StatusBadRequest)
					return
				}
				motivo := strings.TrimSpace(payload.Observaciones)
				if motivo == "" {
					motivo = strings.TrimSpace(r.URL.Query().Get("motivo"))
				}
				if len(motivo) < 10 {
					http.Error(w, "motivo de anulacion es obligatorio y debe tener minimo 10 caracteres", http.StatusBadRequest)
					return
				}
				lockedContext, releaseFacturaLock, facturaLocked, lockErr := acquireFacturacionDocumentAdvisoryLock(
					r.Context(), dbEmp, payload.EmpresaID, "factura_electronica", payload.DocumentoCodigo,
				)
				if lockErr != nil {
					http.Error(w, "No se pudo reservar la factura para la anulacion fiscal", http.StatusInternalServerError)
					return
				}
				if !facturaLocked {
					http.Error(w, "la factura ya tiene una anulacion fiscal en proceso", http.StatusConflict)
					return
				}
				defer releaseFacturaLock()
				r = r.WithContext(lockedContext)
				factura, err := dbpkg.GetEmpresaDocumentoFacturacionByCodigoContext(r.Context(), dbEmp, payload.EmpresaID, "factura_electronica", payload.DocumentoCodigo)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "factura no encontrada", http.StatusNotFound)
						return
					}
					http.Error(w, "No se pudo consultar la factura", http.StatusInternalServerError)
					return
				}
				if strings.ToLower(strings.TrimSpace(factura.EstadoDocumento)) != "emitida" {
					http.Error(w, "solo se puede anular una factura electronica emitida", http.StatusConflict)
					return
				}
				hidratarCUFEOficialFactura(dbEmp, factura)
				if !facturacionCodigoSHA384Valido(factura.CodigoValidacion) {
					http.Error(w, "la factura no tiene un CUFE oficial DIAN valido para relacionar la nota credito", http.StatusConflict)
					return
				}
				if _, err := loadFacturacionFuenteFiscalParaDocumento(r.Context(), dbEmp, *factura); err != nil {
					writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
						"ok": false, "codigo": "fuente_fiscal_factura_no_disponible",
						"error": "La factura aceptada no conserva una fuente fiscal inmutable valida para derivar la nota credito.",
					})
					return
				}
				existentes, err := dbpkg.ListEmpresaDocumentosFacturacionByEmpresaContext(r.Context(), dbEmp, dbpkg.EmpresaDocumentoFacturacionListFilter{
					EmpresaID:       payload.EmpresaID,
					TipoDocumento:   "nota_credito",
					DocumentoQuery:  factura.DocumentoCodigo,
					IncludeInactive: true,
					Limit:           20,
				})
				if err != nil {
					http.Error(w, "No se pudo validar notas credito existentes", http.StatusInternalServerError)
					return
				}
				var nota *dbpkg.EmpresaDocumentoFacturacion
				notaYaExistia := false
				for _, nc := range existentes {
					if strings.EqualFold(facturacionNotaCreditoFacturaOrigen(nc.Observaciones), factura.DocumentoCodigo) {
						existing := nc.EmpresaDocumentoFacturacion
						nota = &existing
						notaYaExistia = true
						break
					}
				}
				colombiaLoc, locErr := time.LoadLocation("America/Bogota")
				if locErr != nil {
					colombiaLoc = time.FixedZone("COT", -5*60*60)
				}
				nowLocal := time.Now().In(colombiaLoc)
				usuario := strings.TrimSpace(adminEmailFromRequest(r))
				if nota == nil {
					nowCode := nowLocal.Format("20060102150405")
					notaCodigo := "NC-" + strings.TrimSpace(factura.DocumentoCodigo) + "-" + nowCode
					observaciones := fmt.Sprintf("%s%s\nAnulacion total de factura %s. Numero legal original: %s. CUFE/CUDE original: %s. Motivo: %s", facturacionNotaCreditoFacturaOrigenMarker, factura.DocumentoCodigo, factura.DocumentoCodigo, factura.NumeroLegal, factura.CodigoValidacion, motivo)
					notaPayload := dbpkg.EmpresaDocumentoFacturacion{
						EmpresaID:            factura.EmpresaID,
						TipoDocumento:        "nota_credito",
						DocumentoCodigo:      notaCodigo,
						EstadoDocumento:      "pendiente_emision",
						EstadoAnterior:       "borrador",
						EventoUltimo:         "nota_credito_pendiente_dian",
						PeriodoContable:      factura.PeriodoContable,
						MontoTotal:           factura.MontoTotal,
						Moneda:               factura.Moneda,
						PaisCodigo:           factura.PaisCodigo,
						AmbienteFE:           factura.AmbienteFE,
						FechaDocumento:       nowLocal.Format("2006-01-02"),
						EntidadRelacionadaID: factura.EntidadRelacionadaID,
						UsuarioCreador:       usuario,
						Observaciones:        observaciones,
					}
					nota, err = dbpkg.UpsertEmpresaDocumentoFacturacionContext(r.Context(), dbEmp, notaPayload)
					if err != nil {
						http.Error(w, "No se pudo crear la nota credito de anulacion", http.StatusInternalServerError)
						return
					}
				}
				nota, err = asegurarNumeroLegalNotaCredito(dbEmp, *nota)
				if err != nil {
					http.Error(w, "No se pudo asignar el consecutivo interno de la nota credito", http.StatusInternalServerError)
					return
				}
				if _, _, err := ensureFacturacionFuenteFiscalNotaCreditoTotal(r.Context(), dbEmp, *nota, *factura); err != nil {
					log.Printf("[facturacion_electronica] fuente nota credito empresa_id=%d factura=%s nota=%s error=%v", factura.EmpresaID, factura.DocumentoCodigo, nota.DocumentoCodigo, err)
					writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
						"ok": false, "codigo": "fuente_fiscal_nota_credito_invalida",
						"error": "No se pudo preparar la fuente fiscal inmutable de la nota credito desde la factura aceptada.",
					})
					return
				}
				notaOperacion := facturacionBuildOperacionPayloadFromDocumento(*nota)
				notaOperacion.ReferenciaDocumentoCodigo = strings.TrimSpace(factura.NumeroLegal)
				if notaOperacion.ReferenciaDocumentoCodigo == "" {
					notaOperacion.ReferenciaDocumentoCodigo = strings.TrimSpace(factura.DocumentoCodigo)
				}
				notaOperacion.ReferenciaCUFE = strings.TrimSpace(factura.CodigoValidacion)
				notaOperacion.ReferenciaFechaEmision = strings.TrimSpace(factura.FechaDocumento)
				notaOperacion.CodigoCorreccion = "2"
				notaOperacion.DescripcionCorreccion = "Anulación de factura electrónica"
				completarClientePayloadFacturacion(dbEmp, factura.EmpresaID, &notaOperacion, *factura)
				integracionFiscal, retryRegistro, integErr := processFacturacionIntegracionForDocumentoContext(r.Context(), dbEmp, notaOperacion, *nota, "nota_credito", usuario, dbSuper)
				if integErr != nil {
					log.Printf("[facturacion_electronica] error nota credito anulacion empresa_id=%d factura=%s nota=%s err=%v", factura.EmpresaID, factura.DocumentoCodigo, nota.DocumentoCodigo, integErr)
					http.Error(w, "No se pudo completar de forma segura la integracion de la nota credito", http.StatusInternalServerError)
					return
				}
				if refreshed, refreshErr := dbpkg.GetEmpresaDocumentoFacturacionByCodigoContext(r.Context(), dbEmp, nota.EmpresaID, nota.TipoDocumento, nota.DocumentoCodigo); refreshErr == nil && refreshed != nil {
					nota = refreshed
				}
				facturaResultado := factura
				anulacionConfirmada := facturacionIntegracionAceptada(integracionFiscal)
				if anulacionConfirmada {
					facturaResultado, err = finalizarFacturaAnuladaPorNotaCredito(dbEmp, *nota, usuario)
					if err != nil {
						http.Error(w, "La nota credito fue aceptada, pero no se pudo finalizar la anulacion local", http.StatusInternalServerError)
						return
					}
					registrarEventoContableNoBloqueante(dbEmp, r, "facturacion", dbpkg.EmpresaEventoContable{
						EmpresaID:       factura.EmpresaID,
						Modulo:          "facturacion",
						Evento:          "factura_anulada_por_nota_credito",
						Entidad:         "factura_electronica",
						EntidadID:       facturaResultado.ID,
						DocumentoTipo:   factura.TipoDocumento,
						DocumentoCodigo: factura.DocumentoCodigo,
						PeriodoContable: factura.PeriodoContable,
						MontoTotal:      factura.MontoTotal,
						Moneda:          factura.Moneda,
						Origen:          "api_facturacion_electronica",
						Observaciones:   motivo,
					}, map[string]interface{}{
						"nota_credito_codigo": nota.DocumentoCodigo,
						"factura_codigo":      factura.DocumentoCodigo,
						"numero_legal":        factura.NumeroLegal,
						"codigo_validacion":   factura.CodigoValidacion,
						"empresa_id":          factura.EmpresaID,
					})
				}
				resp := map[string]interface{}{
					"ok":                          true,
					"accion":                      "anular_factura_nota_credito",
					"empresa_id":                  factura.EmpresaID,
					"factura_original":            facturaResultado,
					"nota_credito":                nota,
					"nota_credito_ya_existia":     notaYaExistia,
					"integracion_fiscal":          integracionFiscal,
					"anulacion_confirmada_dian":   anulacionConfirmada,
					"anulacion_pendiente_dian":    !anulacionConfirmada,
					"regla_anulacion_electronica": "la factura solo cambia a anulada cuando la nota credito queda aceptada por DIAN",
				}
				if retryRegistro != nil {
					resp["cola_reintentos"] = retryRegistro
				}
				status := http.StatusOK
				if !anulacionConfirmada {
					status = http.StatusAccepted
				}
				writeJSON(w, status, resp)
				return
			}
			if !facturacionActionIsPaisConfig(action) && facturacionActionRequiresFiscalIntegration(action) {
				var payload facturacionOperacionPayload
				if r.Body != nil {
					if err := decodeFacturacionOperacionPayload(http.MaxBytesReader(w, r.Body, 64<<10), &payload); err != nil {
						http.Error(w, "JSON de operación de facturación inválido", http.StatusBadRequest)
						return
					}
				}
				if err := facturacionBindAuthorizedEmpresaID(r, &payload.EmpresaID); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if strings.TrimSpace(payload.DocumentoCodigo) == "" {
					payload.DocumentoCodigo = strings.TrimSpace(r.URL.Query().Get("documento_codigo"))
				}
				if strings.TrimSpace(payload.DocumentoCodigo) == "" {
					http.Error(w, "documento_codigo es obligatorio para la accion", http.StatusBadRequest)
					return
				}

				// Current state is server-owned. A client cannot invent it to skip a
				// transition; persisted state is applied below when present.
				payload.EstadoActual = ""

				if payload.ClienteID <= 0 && payload.EntidadID > 0 {
					payload.ClienteID = payload.EntidadID
				}
				if payload.EntidadID <= 0 && payload.ClienteID > 0 {
					payload.EntidadID = payload.ClienteID
				}

				documentoTipo := normalizeFacturacionDocumentoElectronicoTipo(payload.TipoDocumento)
				entidad := facturacionDocumentoEntidad(documentoTipo)
				actionNormalized := normalizeDocumentoState(action)
				if fromAction := facturacionDocumentoTipoFromAction(actionNormalized); fromAction != "" {
					documentoTipo = fromAction
					entidad = facturacionDocumentoEntidad(documentoTipo)
				}
				if !facturacionDocumentoElectronicoPermitido(documentoTipo) {
					http.Error(w, "tipo_documento electronico no soportado", http.StatusBadRequest)
					return
				}
				if !facturacionDocumentoElectronicoDIANCreacionGenericaSoportada(documentoTipo) {
					codigo := "tipo_documento_dian_no_implementado"
					motivo := facturacionDocumentoElectronicoBloqueoMotivo(documentoTipo)
					if documentoTipo == "factura_electronica" {
						codigo = "emision_factura_libre_bloqueada"
						motivo = "la factura electronica comercial solo se genera desde una venta pagada mediante action=facturar_desde_venta; no se reservo consecutivo ni se envio informacion fiscal"
					}
					writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
						"ok":             false,
						"codigo":         codigo,
						"tipo_documento": documentoTipo,
						"bloqueado":      true,
						"motivo":         motivo,
					})
					return
				}
				if actionNormalized == "anular" && documentoTipo == "factura_electronica" {
					http.Error(w, "una factura electronica se anula fiscalmente mediante action=anular_factura_nota_credito", http.StatusConflict)
					return
				}
				lockedContext, releaseDocumentLock, documentLocked, lockErr := acquireFacturacionDocumentAdvisoryLock(
					r.Context(), dbEmp, payload.EmpresaID, documentoTipo, payload.DocumentoCodigo,
				)
				if lockErr != nil {
					http.Error(w, "No se pudo reservar el documento para la operacion fiscal", http.StatusInternalServerError)
					return
				}
				if !documentLocked {
					http.Error(w, "el documento ya tiene una operacion fiscal en proceso", http.StatusConflict)
					return
				}
				defer releaseDocumentLock()
				r = r.WithContext(lockedContext)

				docExistente, err := dbpkg.GetEmpresaDocumentoFacturacionByCodigoContext(r.Context(), dbEmp, payload.EmpresaID, documentoTipo, payload.DocumentoCodigo)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "No se pudo consultar el estado documental de facturacion", http.StatusInternalServerError)
					return
				}
				if docExistente != nil {
					payload.EstadoActual = docExistente.EstadoDocumento
				}

				transition, err := resolveFacturacionTransitionForDocument(action, payload.EstadoActual, documentoTipo)
				if err != nil {
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}

				if block, preflightErr := facturacionOfflineDianPreflight(dbEmp, payload); preflightErr != nil {
					http.Error(w, "No se pudo validar conectividad DIAN", http.StatusInternalServerError)
					return
				} else if block != nil {
					writeJSON(w, http.StatusConflict, block)
					return
				}

				var legalDoc *dbpkg.FacturacionDocumentoLegal
				if transition.Accion == "emitir" && documentoTipo == "factura_electronica" {
					paisCodigo := strings.TrimSpace(payload.PaisCodigo)
					if paisCodigo == "" {
						paisCodigo = strings.TrimSpace(r.URL.Query().Get("pais_codigo"))
					}
					legalDoc, err = dbpkg.PrepareFacturacionDocumentoLegalContext(r.Context(), dbEmp, payload.EmpresaID, paisCodigo, payload.DocumentoCodigo, payload.MontoTotal, payload.Moneda)
					if err != nil {
						http.Error(w, "cumplimiento normativo: "+err.Error(), http.StatusUnprocessableEntity)
						return
					}
				}

				evento := transition.Evento
				docPayload := dbpkg.EmpresaDocumentoFacturacion{
					EmpresaID:            payload.EmpresaID,
					TipoDocumento:        documentoTipo,
					DocumentoCodigo:      payload.DocumentoCodigo,
					EstadoDocumento:      transition.EstadoNuevo,
					EstadoAnterior:       transition.EstadoAnterior,
					EventoUltimo:         evento,
					PeriodoContable:      payload.PeriodoContable,
					MontoTotal:           payload.MontoTotal,
					Moneda:               payload.Moneda,
					EntidadRelacionadaID: payload.EntidadID,
					UsuarioCreador:       strings.TrimSpace(adminEmailFromRequest(r)),
					Observaciones:        payload.Observaciones,
				}
				if legalDoc != nil {
					docPayload.NumeroLegal = legalDoc.NumeroLegal
					docPayload.CodigoValidacion = legalDoc.CodigoValidacion
					docPayload.PaisCodigo = legalDoc.PaisCodigo
					docPayload.AmbienteFE = legalDoc.Ambiente
					docPayload.FechaDocumento = legalDoc.FechaEmisionLegal
				}
				if docExistente != nil {
					if docPayload.NumeroLegal == "" {
						docPayload.NumeroLegal = docExistente.NumeroLegal
					}
					if docPayload.CodigoValidacion == "" {
						docPayload.CodigoValidacion = docExistente.CodigoValidacion
					}
					if docPayload.PaisCodigo == "" {
						docPayload.PaisCodigo = docExistente.PaisCodigo
					}
					if docPayload.AmbienteFE == "" {
						docPayload.AmbienteFE = docExistente.AmbienteFE
					}
					if docPayload.FechaDocumento == "" {
						docPayload.FechaDocumento = docExistente.FechaDocumento
					}
				}

				docPersistido, err := dbpkg.UpsertEmpresaDocumentoFacturacionContext(r.Context(), dbEmp, docPayload)
				if err != nil {
					http.Error(w, "No se pudo persistir el documento transaccional", http.StatusInternalServerError)
					return
				}

				registrarEventoContableNoBloqueante(dbEmp, r, "facturacion", dbpkg.EmpresaEventoContable{
					EmpresaID:       payload.EmpresaID,
					Modulo:          "facturacion",
					Evento:          evento,
					Entidad:         entidad,
					EntidadID:       docPersistido.ID,
					DocumentoTipo:   documentoTipo,
					DocumentoCodigo: strings.TrimSpace(payload.DocumentoCodigo),
					PeriodoContable: strings.TrimSpace(payload.PeriodoContable),
					MontoTotal:      payload.MontoTotal,
					Moneda:          strings.ToUpper(strings.TrimSpace(payload.Moneda)),
					Origen:          "api_facturacion_electronica",
					Observaciones:   strings.TrimSpace(payload.Observaciones),
				}, map[string]interface{}{
					"accion":            transition.Accion,
					"estado_anterior":   transition.EstadoAnterior,
					"estado_nuevo":      transition.EstadoNuevo,
					"entidad_id":        docPersistido.ID,
					"documento_codigo":  strings.TrimSpace(payload.DocumentoCodigo),
					"numero_legal":      docPersistido.NumeroLegal,
					"codigo_validacion": docPersistido.CodigoValidacion,
					"pais_codigo":       docPersistido.PaisCodigo,
					"ambiente_fe":       docPersistido.AmbienteFE,
					"periodo_contable":  strings.TrimSpace(payload.PeriodoContable),
					"forma_pago":        strings.TrimSpace(payload.FormaPago),
					"metodo_pago":       strings.TrimSpace(payload.MetodoPago),
					"subtotal":          payload.Subtotal,
					"base_gravable":     payload.BaseGravable,
					"iva":               firstPositiveFloat64(payload.IVA, payload.Impuestos),
					"impuestos":         firstPositiveFloat64(payload.Impuestos, payload.IVA),
					"retencion_fuente":  payload.RetencionFuente,
					"retencion_ica":     payload.RetencionICA,
					"retencion_iva":     payload.RetencionIVA,
					"total_retenciones": firstPositiveFloat64(payload.TotalRetenciones, payload.RetencionFuente+payload.RetencionICA+payload.RetencionIVA),
					"total_neto":        payload.TotalNeto,
					"cliente_id":        payload.ClienteID,
					"cliente_nombre":    strings.TrimSpace(payload.ClienteNombre),
					"cliente_documento": strings.TrimSpace(payload.ClienteNumeroDocumento),
					"empresa_id":        payload.EmpresaID,
				})

				integracionFiscal := facturacionIntegracionResultado{
					Aplica:             false,
					Accion:             transition.Accion,
					EstadoEnvio:        "no_aplica",
					ContingenciaActiva: false,
				}
				var retryRegistro *dbpkg.FacturacionElectronicaRetryItem
				var integracionErr error

				if facturacionActionRequiresFiscalIntegration(transition.Accion) {
					resultadoIntegracion, retryItem, integErr := processFacturacionIntegracionForDocumentoContext(
						r.Context(),
						dbEmp,
						payload,
						*docPersistido,
						transition.Accion,
						strings.TrimSpace(adminEmailFromRequest(r)),
						dbSuper,
					)
					if integErr != nil {
						log.Printf("[facturacion_electronica] error integracion fiscal empresa_id=%d documento=%s accion=%s err=%v", payload.EmpresaID, payload.DocumentoCodigo, transition.Accion, integErr)
					}
					integracionErr = integErr
					integracionFiscal = resultadoIntegracion
					retryRegistro = retryItem

					eventoIntegracion := ""
					switch integracionFiscal.EstadoEnvio {
					case "enviado":
						eventoIntegracion = "factura_integracion_enviada"
					case "fallido":
						eventoIntegracion = "factura_integracion_fallida"
					case "contingencia":
						eventoIntegracion = "factura_contingencia_activada"
					}

					if eventoIntegracion != "" {
						registrarEventoContableNoBloqueante(dbEmp, r, "facturacion", dbpkg.EmpresaEventoContable{
							EmpresaID:       payload.EmpresaID,
							Modulo:          "facturacion",
							Evento:          eventoIntegracion,
							Entidad:         entidad,
							EntidadID:       docPersistido.ID,
							DocumentoTipo:   docPersistido.TipoDocumento,
							DocumentoCodigo: docPersistido.DocumentoCodigo,
							PeriodoContable: docPersistido.PeriodoContable,
							MontoTotal:      docPersistido.MontoTotal,
							Moneda:          docPersistido.Moneda,
							Origen:          "api_facturacion_electronica",
							Observaciones:   strings.TrimSpace(integracionFiscal.Error),
						}, map[string]interface{}{
							"accion":              transition.Accion,
							"estado_envio":        integracionFiscal.EstadoEnvio,
							"intentos":            integracionFiscal.Intentos,
							"max_intentos":        integracionFiscal.MaxIntentos,
							"contingencia_activa": integracionFiscal.ContingenciaActiva,
							"proximo_intento":     integracionFiscal.ProximoIntento,
							"referencia_externa":  integracionFiscal.ReferenciaExterna,
							"documento_codigo":    docPersistido.DocumentoCodigo,
							"codigo_validacion":   docPersistido.CodigoValidacion,
							"numero_legal":        docPersistido.NumeroLegal,
							"empresa_id":          payload.EmpresaID,
						})
					}
				}
				operacionFiscalConfirmada := !facturacionActionRequiresFiscalIntegration(transition.Accion) || facturacionIntegracionAceptada(integracionFiscal)
				if !operacionFiscalConfirmada {
					motivo := strings.TrimSpace(integracionFiscal.Error)
					if motivo == "" && integracionErr != nil {
						motivo = integracionErr.Error()
					}
					if motivo == "" {
						motivo = "acuse fiscal DIAN/proveedor aun no confirmado"
					}
					docPendiente := *docPersistido
					docPendiente.EstadoAnterior = strings.TrimSpace(docPersistido.EstadoDocumento)
					docPendiente.EstadoDocumento = "pendiente_emision"
					docPendiente.EventoUltimo = "integracion_fiscal_pendiente"
					docPendiente.Observaciones = strings.TrimSpace(docPendiente.Observaciones + ". Pendiente fiscal: " + facturacionTruncate(motivo, 240))
					updated, updateErr := dbpkg.UpsertEmpresaDocumentoFacturacionContext(r.Context(), dbEmp, docPendiente)
					if updateErr != nil {
						http.Error(w, "La integracion fiscal no fue confirmada y no se pudo persistir el estado pendiente", http.StatusInternalServerError)
						return
					}
					if updated != nil {
						docPersistido = updated
					}
				}
				if refreshed, refreshErr := dbpkg.GetEmpresaDocumentoFacturacionByCodigoContext(r.Context(), dbEmp, payload.EmpresaID, docPersistido.TipoDocumento, docPersistido.DocumentoCodigo); refreshErr == nil && refreshed != nil {
					docPersistido = refreshed
				}

				resp := map[string]interface{}{
					"ok":                 operacionFiscalConfirmada,
					"accion":             transition.Accion,
					"evento":             evento,
					"estado_anterior":    transition.EstadoAnterior,
					"estado_nuevo":       docPersistido.EstadoDocumento,
					"entidad_id":         docPersistido.ID,
					"documento_codigo":   strings.TrimSpace(payload.DocumentoCodigo),
					"numero_legal":       docPersistido.NumeroLegal,
					"codigo_validacion":  docPersistido.CodigoValidacion,
					"pais_codigo":        docPersistido.PaisCodigo,
					"ambiente_fe":        docPersistido.AmbienteFE,
					"integracion_fiscal": integracionFiscal,
				}
				if retryRegistro != nil {
					resp["cola_reintentos"] = retryRegistro
				}
				if !operacionFiscalConfirmada {
					resp["pendiente_emision"] = true
					resp["warning"] = "La operacion local fue registrada, pero no existe aceptacion fiscal concluyente; el documento sigue pendiente de emision."
				}
				if legalDoc != nil {
					resp["cumplimiento_normativo"] = map[string]interface{}{
						"validado":               true,
						"prefijo_factura":        legalDoc.PrefijoFactura,
						"resolucion_numero":      legalDoc.ResolucionNumero,
						"consecutivo_asignado":   legalDoc.ConsecutivoAsignado,
						"fecha_emision_legal":    legalDoc.FechaEmisionLegal,
						"resolucion_fecha_desde": legalDoc.ResolucionFechaDesde,
						"resolucion_fecha_hasta": legalDoc.ResolucionFechaHasta,
					}
				}
				if transition.Accion == "emitir" && documentoTipo == "factura_electronica" {
					if facturacionAutoEmailClienteEnabled(dbEmp, payload.EmpresaID, payload.PaisCodigo) {
						if strings.EqualFold(docPersistido.PaisCodigo, "CO") && strings.EqualFold(docPersistido.AmbienteFE, "produccion") && integracionFiscal.EstadoEnvio != "aceptado" {
							resp["factura_email"] = facturaEmailResultado{Intentado: false, Enviado: false, Error: "correo fiscal pendiente hasta recibir aceptacion DIAN"}
						} else {
							resp["factura_email"] = enviarFacturaElectronicaAlCliente(dbEmp, dbSuper, payload, *docPersistido)
						}
					} else {
						resp["factura_email"] = facturaEmailAutoDisabledResultado(payload)
					}
				}
				status := http.StatusOK
				if !operacionFiscalConfirmada {
					status = http.StatusAccepted
				}
				writeJSON(w, status, resp)
				return
			}

			var payload dbpkg.FacturacionElectronicaPaisConfig
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}

			if err := facturacionBindAuthorizedEmpresaID(r, &payload.EmpresaID); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.PaisCodigo) == "" {
				http.Error(w, "pais_codigo es obligatorio", http.StatusBadRequest)
				return
			}
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))

			id, err := dbpkg.UpsertFacturacionElectronicaPaisConfigContext(r.Context(), dbEmp, payload)
			if err != nil {
				http.Error(w, "No se pudo guardar la configuración de facturación electrónica", http.StatusBadRequest)
				return
			}
			cfg, err := dbpkg.GetFacturacionElectronicaPaisConfigContext(r.Context(), dbEmp, payload.EmpresaID, payload.PaisCodigo)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "No se pudo recuperar la configuración guardada", http.StatusInternalServerError)
				return
			}
			monedaEvento := strings.ToUpper(strings.TrimSpace(payload.MonedaCodigo))
			if monedaEvento == "" && cfg != nil {
				monedaEvento = strings.ToUpper(strings.TrimSpace(cfg.MonedaCodigo))
			}
			ambienteEvento := strings.ToLower(strings.TrimSpace(payload.Ambiente))
			proveedorEvento := strings.TrimSpace(payload.Proveedor)
			estadoEvento := strings.ToLower(strings.TrimSpace(payload.Estado))
			if cfg != nil {
				ambienteEvento = strings.ToLower(strings.TrimSpace(cfg.Ambiente))
				proveedorEvento = strings.TrimSpace(cfg.Proveedor)
				estadoEvento = strings.ToLower(strings.TrimSpace(cfg.Estado))
			}
			registrarEventoContableNoBloqueante(dbEmp, r, "facturacion", dbpkg.EmpresaEventoContable{
				EmpresaID:       payload.EmpresaID,
				Modulo:          "facturacion",
				Evento:          "configuracion_facturacion_actualizada",
				Entidad:         "facturacion_electronica_pais",
				EntidadID:       id,
				DocumentoTipo:   "facturacion_pais",
				DocumentoCodigo: strings.ToUpper(strings.TrimSpace(payload.PaisCodigo)),
				Moneda:          monedaEvento,
				Origen:          "api_facturacion_electronica",
				Observaciones:   "configuracion de facturacion electronica actualizada",
			}, map[string]interface{}{
				"pais_codigo": strings.ToUpper(strings.TrimSpace(payload.PaisCodigo)),
				"ambiente":    ambienteEvento,
				"proveedor":   proveedorEvento,
				"estado":      estadoEvento,
			})
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":            true,
				"id":            id,
				"configuracion": cfg,
			})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

func facturaEmailAutoDisabledResultado(payload facturacionOperacionPayload) facturaEmailResultado {
	clienteID := payload.ClienteID
	if clienteID <= 0 {
		clienteID = payload.EntidadID
	}
	return facturaEmailResultado{
		Intentado:             false,
		Enviado:               false,
		AutomaticoDesactivado: true,
		ClienteID:             clienteID,
		OrigenDestinatario:    "configuracion",
	}
}

// EmpresaFacturacionElectronicaPanamaHandler gestiona el perfil independiente Panama/DGI.
func EmpresaFacturacionElectronicaPanamaHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch r.Method {
		case http.MethodGet:
			cfg, err := dbpkg.GetFacturacionElectronicaPaisConfigContext(r.Context(), dbEmp, empresaID, "PA")
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "No se pudo cargar facturacion electronica Panama", http.StatusInternalServerError)
				return
			}
			checklist := dbpkg.BuildFacturacionPanamaChecklist(cfg)
			if action == "checklist" || action == "validar" {
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":            checklist.Ok,
					"empresa_id":    empresaID,
					"pais_codigo":   "PA",
					"checklist":     checklist,
					"configuracion": cfg,
				})
				return
			}
			if action == "guia_onboarding" {
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":          true,
					"empresa_id":  empresaID,
					"pais_codigo": "PA",
					"pasos": []map[string]string{
						{"clave": "registro_sfep", "titulo": "Registrarse en SFEP/e-Tax2.0", "detalle": "Completar la solicitud de factura electronica ante DGI Panama."},
						{"clave": "modalidad", "titulo": "Elegir modalidad", "detalle": "Seleccionar Facturador Gratuito o Proveedor Autorizado Calificado (PAC)."},
						{"clave": "firma", "titulo": "Configurar firma electronica", "detalle": "Registrar certificado o referencia segura para firmar documentos electronicos."},
						{"clave": "pruebas", "titulo": "Validar ambiente de pruebas", "detalle": "Probar emision, CAFE/CUFE/QR y respuesta del PAC o facturador antes de produccion."},
					},
					"fuentes": checklist.Fuentes,
				})
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":            true,
				"empresa_id":    empresaID,
				"pais_codigo":   "PA",
				"configuracion": cfg,
				"checklist":     checklist,
				"vista":         dbpkg.FacturacionPaisVistaFor("PA"),
			})
			return
		case http.MethodPost, http.MethodPut:
			var payload dbpkg.FacturacionElectronicaPaisConfig
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			payload.EmpresaID = empresaID
			payload.PaisCodigo = "PA"
			payload.PaisNombre = "Panama"
			payload.BanderaPais = "PA"
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			if strings.TrimSpace(payload.MonedaCodigo) == "" {
				payload.MonedaCodigo = "PAB"
			}
			id, err := dbpkg.UpsertFacturacionElectronicaPaisConfigContext(r.Context(), dbEmp, payload)
			if err != nil {
				http.Error(w, "No se pudo guardar facturacion electronica Panama", http.StatusBadRequest)
				return
			}
			cfg, err := dbpkg.GetFacturacionElectronicaPaisConfigContext(r.Context(), dbEmp, empresaID, "PA")
			if err != nil {
				http.Error(w, "No se pudo recuperar facturacion electronica Panama", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":            true,
				"id":            id,
				"empresa_id":    empresaID,
				"pais_codigo":   "PA",
				"configuracion": cfg,
				"checklist":     dbpkg.BuildFacturacionPanamaChecklist(cfg),
			})
			return
		default:
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
	}
}

// EmpresaFacturacionElectronicaEcuadorHandler gestiona el perfil independiente Ecuador/SRI.
func EmpresaFacturacionElectronicaEcuadorHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch r.Method {
		case http.MethodGet:
			cfg, err := dbpkg.GetFacturacionElectronicaPaisConfigContext(r.Context(), dbEmp, empresaID, "EC")
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "No se pudo cargar facturacion electronica Ecuador", http.StatusInternalServerError)
				return
			}
			checklist := dbpkg.BuildFacturacionEcuadorChecklist(cfg)
			if action == "checklist" || action == "validar" {
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":            checklist.Ok,
					"empresa_id":    empresaID,
					"pais_codigo":   "EC",
					"checklist":     checklist,
					"configuracion": cfg,
				})
				return
			}
			if action == "guia_onboarding" {
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":          true,
					"empresa_id":  empresaID,
					"pais_codigo": "EC",
					"pasos": []map[string]string{
						{"clave": "firma", "titulo": "Adquirir firma electronica", "detalle": "Mantener vigente un certificado de firma electronica tipo archivo para firmar XML."},
						{"clave": "ambiente_pruebas", "titulo": "Preparar ambiente de pruebas", "detalle": "Configurar ambiente SRI 1 para validar XML, firma y secuencias antes de produccion."},
						{"clave": "autorizacion", "titulo": "Confirmar autorizacion SRI", "detalle": "Verificar autorizacion de emision de comprobantes electronicos en SRI en Linea para produccion."},
						{"clave": "ride", "titulo": "Generar RIDE y notificar", "detalle": "Emitir XML autorizado, generar representacion impresa RIDE y notificar por correo al destinatario."},
					},
					"fuentes": checklist.Fuentes,
				})
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":            true,
				"empresa_id":    empresaID,
				"pais_codigo":   "EC",
				"configuracion": cfg,
				"checklist":     checklist,
				"vista":         dbpkg.FacturacionPaisVistaFor("EC"),
			})
			return
		case http.MethodPost, http.MethodPut:
			var payload dbpkg.FacturacionElectronicaPaisConfig
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			payload.EmpresaID = empresaID
			payload.PaisCodigo = "EC"
			payload.PaisNombre = "Ecuador"
			payload.BanderaPais = "EC"
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			if strings.TrimSpace(payload.MonedaCodigo) == "" {
				payload.MonedaCodigo = "USD"
			}
			id, err := dbpkg.UpsertFacturacionElectronicaPaisConfigContext(r.Context(), dbEmp, payload)
			if err != nil {
				http.Error(w, "No se pudo guardar facturacion electronica Ecuador", http.StatusBadRequest)
				return
			}
			cfg, err := dbpkg.GetFacturacionElectronicaPaisConfigContext(r.Context(), dbEmp, empresaID, "EC")
			if err != nil {
				http.Error(w, "No se pudo recuperar facturacion electronica Ecuador", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":            true,
				"id":            id,
				"empresa_id":    empresaID,
				"pais_codigo":   "EC",
				"configuracion": cfg,
				"checklist":     dbpkg.BuildFacturacionEcuadorChecklist(cfg),
			})
			return
		default:
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
	}
}

func facturacionAutoEmailClienteEnabled(dbEmp *sql.DB, empresaID int64, paisCodigo string) bool {
	if dbEmp == nil || empresaID <= 0 {
		return false
	}
	code := strings.TrimSpace(paisCodigo)
	if code == "" {
		code = "CO"
	}
	if cfg, err := dbpkg.GetFacturacionElectronicaPaisConfig(dbEmp, empresaID, code); err == nil && cfg != nil && cfg.EnviarFacturaEmailClienteAuto {
		return true
	}
	if cfg, err := dbpkg.GetEmpresaConfiguracionAvanzada(dbEmp, empresaID); err == nil && cfg != nil && cfg.EnviarFacturaElectronicaVenta {
		return true
	}
	return false
}

func enviarFacturaElectronicaAlCliente(dbEmp, dbSuper *sql.DB, payload facturacionOperacionPayload, doc dbpkg.EmpresaDocumentoFacturacion) facturaEmailResultado {
	emailCliente, nombreCliente, clienteID, origen, err := resolverDestinoCorreoFactura(dbEmp, payload)
	resultado := facturaEmailResultado{
		Intentado:          false,
		Enviado:            false,
		ClienteID:          clienteID,
		OrigenDestinatario: origen,
	}
	if err != nil {
		resultado.Error = err.Error()
		return resultado
	}
	if strings.TrimSpace(emailCliente) == "" {
		resultado.Error = "sin destinatario de cliente para envio automatico"
		return resultado
	}

	resultado.Intentado = true
	resultado.Destinatario = emailCliente
	if err := sendFacturaElectronicaEmail(dbEmp, dbSuper, emailCliente, nombreCliente, doc, payload); err != nil {
		resultado.Error = err.Error()
		log.Printf("[facturacion_electronica] envio correo fallido empresa_id=%d documento=%s destinatario=%s error=%v", payload.EmpresaID, payload.DocumentoCodigo, redactEmailForLog(emailCliente), err)
		return resultado
	}

	resultado.Enviado = true
	resultado.Error = ""
	return resultado
}

func resolverDestinoCorreoFactura(dbEmp *sql.DB, payload facturacionOperacionPayload) (string, string, int64, string, error) {
	clienteID := payload.ClienteID
	if clienteID <= 0 {
		clienteID = payload.EntidadID
	}

	emailCliente := strings.TrimSpace(payload.ClienteEmail)
	nombreCliente := strings.TrimSpace(payload.ClienteNombre)
	if emailCliente != "" {
		if _, err := mail.ParseAddress(emailCliente); err != nil {
			return "", nombreCliente, clienteID, "payload", fmt.Errorf("cliente_email invalido: %w", err)
		}
		if nombreCliente == "" {
			nombreCliente = "cliente"
		}
		return emailCliente, nombreCliente, clienteID, "payload", nil
	}

	if clienteID <= 0 {
		return "", nombreCliente, 0, "sin_cliente", nil
	}

	cliente, err := dbpkg.GetClienteByID(dbEmp, payload.EmpresaID, clienteID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nombreCliente, clienteID, "cliente_id", nil
		}
		return "", nombreCliente, clienteID, "cliente_id", err
	}

	emailCliente = strings.TrimSpace(cliente.Email)
	if nombreCliente == "" {
		nombreCliente = strings.TrimSpace(cliente.NombreRazonSocial)
	}
	if emailCliente == "" {
		return "", nombreCliente, clienteID, "cliente_id", nil
	}
	if _, err := mail.ParseAddress(emailCliente); err != nil {
		return "", nombreCliente, clienteID, "cliente_id", fmt.Errorf("email de cliente invalido en registro: %w", err)
	}
	if nombreCliente == "" {
		nombreCliente = "cliente"
	}

	return emailCliente, nombreCliente, clienteID, "cliente_id", nil
}

func sendFacturaElectronicaEmail(dbEmp, dbSuper *sql.DB, toEmail, toName string, doc dbpkg.EmpresaDocumentoFacturacion, payload facturacionOperacionPayload) error {
	if dbSuper == nil {
		return fmt.Errorf("configuracion de correo corporativo no disponible")
	}

	fromName, fromEmail := corporateSystemSenderAddress(dbSuper, "soporte")
	safeName := strings.TrimSpace(toName)
	if safeName == "" {
		safeName = "cliente"
	}
	numeroLegal := strings.TrimSpace(doc.NumeroLegal)
	if numeroLegal == "" {
		numeroLegal = strings.TrimSpace(doc.DocumentoCodigo)
	}
	codigoValidacion := strings.TrimSpace(doc.CodigoValidacion)
	monto := doc.MontoTotal
	if monto <= 0 {
		monto = payload.MontoTotal
	}
	moneda := strings.ToUpper(strings.TrimSpace(doc.Moneda))
	if moneda == "" {
		moneda = strings.ToUpper(strings.TrimSpace(payload.Moneda))
	}
	if moneda == "" {
		moneda = "COP"
	}

	documentLabel := "Factura electronica"
	introLine := "Tu factura electronica fue emitida correctamente."
	qrURL := facturaElectronicaDIANQRURL(doc)
	feDetail := "Pais FE: " + strings.ToUpper(strings.TrimSpace(doc.PaisCodigo)) + "\r\n" +
		"Ambiente FE: " + strings.TrimSpace(doc.AmbienteFE) + "\r\n"
	if strings.EqualFold(strings.TrimSpace(doc.TipoDocumento), "comprobante_pago") {
		documentLabel = "Comprobante de pago"
		introLine = "Tu comprobante de pago fue generado correctamente."
		feDetail = ""
	}

	subject := documentLabel + " emitido " + numeroLegal
	body := "Hola " + safeName + ",\r\n\r\n" +
		introLine + "\r\n" +
		"Documento: " + strings.TrimSpace(doc.DocumentoCodigo) + "\r\n" +
		"Numero legal: " + numeroLegal + "\r\n" +
		"Codigo de validacion: " + codigoValidacion + "\r\n" +
		"Total: " + fmt.Sprintf("%.2f", monto) + " " + moneda + "\r\n" +
		feDetail + "\r\n" +
		func() string {
			if qrURL == "" {
				return ""
			}
			return "Consulta DIAN / QR: " + qrURL + "\r\n\r\n"
		}() +
		"Gracias por tu compra.\r\n"

	bodyHTML := facturaElectronicaEmailHTML(safeName, introLine, documentLabel, numeroLegal, codigoValidacion, monto, moneda, doc, qrURL)
	baseName := facturaElectronicaAttachmentBaseName(numeroLegal, doc.DocumentoCodigo)
	boundaryMixed := "pcs_fe_mixed_" + time.Now().UTC().Format("20060102150405")
	boundaryAlt := "pcs_fe_alt_" + time.Now().UTC().Format("150405")
	artifactParts := ""
	if strings.EqualFold(strings.TrimSpace(doc.TipoDocumento), "factura_electronica") {
		requireFiscalArtifacts := strings.EqualFold(strings.TrimSpace(doc.PaisCodigo), "CO") && strings.EqualFold(strings.TrimSpace(doc.AmbienteFE), "produccion")
		xmlContent, xmlErr := loadFacturacionFiscalArtifact(context.Background(), dbEmp, doc, "xml_firmado")
		if xmlErr == nil && len(xmlContent) > 0 {
			artifactParts += facturaElectronicaEmailAttachmentPart(boundaryMixed, baseName+".xml", "application/xml", xmlContent)
		} else if requireFiscalArtifacts {
			return fmt.Errorf("XML fiscal firmado no disponible o con integridad invalida")
		}
		pdfContent, pdfErr := loadFacturacionFiscalArtifact(context.Background(), dbEmp, doc, "representacion_pdf")
		if pdfErr == nil && len(pdfContent) > 0 {
			artifactParts += facturaElectronicaEmailAttachmentPart(boundaryMixed, baseName+".pdf", "application/pdf", pdfContent)
		} else if requireFiscalArtifacts {
			return fmt.Errorf("representacion PDF fiscal no disponible o con integridad invalida")
		}
	}
	msg := "From: " + (&mail.Address{Name: fromName, Address: fromEmail}).String() + "\r\n" +
		"To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/mixed; boundary=\"" + boundaryMixed + "\"\r\n\r\n" +
		"--" + boundaryMixed + "\r\n" +
		"Content-Type: multipart/alternative; boundary=\"" + boundaryAlt + "\"\r\n\r\n" +
		"--" + boundaryAlt + "\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"Content-Transfer-Encoding: 8bit\r\n\r\n" +
		body + "\r\n" +
		"--" + boundaryAlt + "\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"Content-Transfer-Encoding: 8bit\r\n\r\n" +
		bodyHTML + "\r\n" +
		"--" + boundaryAlt + "--\r\n" +
		facturaElectronicaEmailAttachmentPart(boundaryMixed, baseName+".html", "text/html; charset=UTF-8", []byte(bodyHTML)) +
		facturaElectronicaEmailAttachmentPart(boundaryMixed, baseName+".txt", "text/plain; charset=UTF-8", []byte(body)) +
		artifactParts +
		"--" + boundaryMixed + "--\r\n"

	return sendEmpresaUsuarioMailuMessage(dbSuper, fromEmail, toEmail, []byte(msg))
}

// EmpresaFacturacionElectronicaPaisDetectadoHandler detecta automáticamente país FE.
func facturaElectronicaDIANQRURL(doc dbpkg.EmpresaDocumentoFacturacion) string {
	key := strings.TrimSpace(doc.CodigoValidacion)
	if key == "" {
		return ""
	}
	if strings.EqualFold(strings.TrimSpace(doc.PaisCodigo), "CO") && !facturacionCodigoSHA384Valido(key) {
		return ""
	}
	base := "https://catalogo-vpfe.dian.gov.co/document/searchqr?documentkey="
	ambiente := strings.ToLower(strings.TrimSpace(doc.AmbienteFE))
	if strings.Contains(ambiente, "hab") || strings.Contains(ambiente, "test") || ambiente == "2" {
		base = "https://catalogo-vpfe-hab.dian.gov.co/document/searchqr?documentkey="
	}
	return base + url.QueryEscape(key)
}

func facturaElectronicaEmailHTML(toName, introLine, documentLabel, numeroLegal, codigoValidacion string, monto float64, moneda string, doc dbpkg.EmpresaDocumentoFacturacion, qrURL string) string {
	qrBlock := ""
	if qrURL != "" {
		qrBlock = `<p><a href="` + html.EscapeString(qrURL) + `" style="display:inline-block;padding:12px 18px;background:#0f4c81;color:#ffffff;text-decoration:none;font-weight:700;">Verificar en DIAN / QR</a></p>` +
			`<p style="font-size:12px;color:#475569;word-break:break-all;">` + html.EscapeString(qrURL) + `</p>`
	}
	return `<!doctype html><html><body style="font-family:Arial,Helvetica,sans-serif;color:#111827;line-height:1.45;">` +
		`<h1 style="font-size:20px;margin:0 0 12px;">` + html.EscapeString(documentLabel) + `</h1>` +
		`<p>Hola ` + html.EscapeString(toName) + `,</p>` +
		`<p>` + html.EscapeString(introLine) + `</p>` +
		`<table style="border-collapse:collapse;width:100%;max-width:620px;font-size:14px;">` +
		facturaElectronicaEmailRow("Documento", doc.DocumentoCodigo) +
		facturaElectronicaEmailRow("Numero legal", numeroLegal) +
		facturaElectronicaEmailRow("Codigo de validacion", codigoValidacion) +
		facturaElectronicaEmailRow("Total", fmt.Sprintf("%.2f %s", monto, moneda)) +
		facturaElectronicaEmailRow("Pais FE", strings.ToUpper(strings.TrimSpace(doc.PaisCodigo))) +
		facturaElectronicaEmailRow("Ambiente FE", doc.AmbienteFE) +
		`</table>` + qrBlock +
		`<p style="font-size:12px;color:#475569;">Adjuntamos la representacion del documento. Cuando el acuse DIAN esta aceptado, el correo incluye tambien el XML firmado y la representacion PDF conservados por la empresa.</p>` +
		`</body></html>`
}

func facturaElectronicaEmailRow(label, value string) string {
	return `<tr><th style="text-align:left;border:1px solid #d1d5db;background:#f3f4f6;padding:8px;">` + html.EscapeString(label) + `</th><td style="border:1px solid #d1d5db;padding:8px;">` + html.EscapeString(strings.TrimSpace(value)) + `</td></tr>`
}

func facturaElectronicaAttachmentBaseName(numeroLegal, documentoCodigo string) string {
	raw := strings.TrimSpace(numeroLegal)
	if raw == "" {
		raw = strings.TrimSpace(documentoCodigo)
	}
	if raw == "" {
		raw = "factura_electronica"
	}
	var b strings.Builder
	for _, r := range raw {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "factura_electronica"
	}
	return out
}

func facturaElectronicaEmailAttachmentPart(boundary, filename, contentType string, content []byte) string {
	encoded := base64.StdEncoding.EncodeToString(content)
	var wrapped strings.Builder
	for len(encoded) > 76 {
		wrapped.WriteString(encoded[:76])
		wrapped.WriteString("\r\n")
		encoded = encoded[76:]
	}
	if encoded != "" {
		wrapped.WriteString(encoded)
		wrapped.WriteString("\r\n")
	}
	return "--" + boundary + "\r\n" +
		"Content-Type: " + contentType + "; name=\"" + filename + "\"\r\n" +
		"Content-Disposition: attachment; filename=\"" + filename + "\"\r\n" +
		"Content-Transfer-Encoding: base64\r\n\r\n" +
		wrapped.String()
}

func EmpresaFacturacionElectronicaPaisDetectadoHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}

		empresaID, err := parseInt64QueryOptional(r, "empresa_id")
		if err != nil {
			http.Error(w, "empresa_id inválido", http.StatusBadRequest)
			return
		}

		tz := strings.TrimSpace(r.URL.Query().Get("tz"))
		if tz == "" {
			tz = strings.TrimSpace(r.URL.Query().Get("timezone"))
		}
		lang := strings.TrimSpace(r.URL.Query().Get("lang"))
		if lang == "" {
			acceptLang := strings.TrimSpace(r.Header.Get("Accept-Language"))
			if idx := strings.Index(acceptLang, ","); idx > 0 {
				lang = strings.TrimSpace(acceptLang[:idx])
			} else {
				lang = acceptLang
			}
		}

		pais, source, err := dbpkg.DetectFacturacionPaisContext(r.Context(), dbEmp, empresaID, tz, lang)
		if err != nil {
			http.Error(w, "No se pudo detectar el país", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"empresa_id":  empresaID,
			"pais_codigo": pais.Codigo,
			"pais_nombre": pais.Nombre,
			"bandera":     pais.Bandera,
			"moneda":      pais.Moneda,
			"source":      source,
			"vista":       dbpkg.FacturacionPaisVistaFor(pais.Codigo),
		})
	}
}

// EmpresaFacturacionElectronicaPaisesDisponiblesHandler retorna catálogo de países FE soportados.
func EmpresaFacturacionElectronicaPaisesDisponiblesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"items": dbpkg.ListFacturacionPaisesConVista(),
		})
	}
}

func normalizeFacturacionEstadoEnvio(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "pendiente", "fallido", "fallido_terminal", "enviado", "aceptado", "reconciliado", "contingencia", "no_aplica":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return "pendiente"
	}
}

func facturacionDianDocumentoProcesadoAnteriormente(raw string) bool {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" {
		return false
	}
	return strings.Contains(raw, "regla: 90") && strings.Contains(raw, "documento procesado anteriormente")
}

func facturacionNowLocal() string {
	return time.Now().In(dianColombiaLocation()).Format("2006-01-02 15:04:05")
}

func facturacionNextRetryAt(intentos int64) string {
	if intentos < 0 {
		intentos = 0
	}
	minutes := int64(1)
	if intentos > 0 {
		minutes = 1 << intentos
	}
	if minutes > 120 {
		minutes = 120
	}
	return time.Now().In(dianColombiaLocation()).Add(time.Duration(minutes) * time.Minute).Format("2006-01-02 15:04:05")
}

func facturacionFirstNonBlank(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func normalizeFacturacionDocumentoElectronicoTipo(raw string) string {
	v := normalizeDocumentoState(raw)
	switch v {
	case "", "invoice", "factura", "factura_venta", "factura_de_venta", "factura_electronica_venta", "factura_electronica":
		return "factura_electronica"
	case "nota_credito", "nota_credito_ventas", "nota_credito_venta", "credit_note", "creditnote":
		return "nota_credito"
	case "nota_debito", "nota_debito_ventas", "nota_debito_venta", "debit_note", "debitnote":
		return "nota_debito"
	case "documento_soporte", "documento_soporte_electronico", "documento_soporte_adquisicion", "documento_soporte_adquisiciones", "soporte_compras":
		return "documento_soporte"
	case "nota_ajuste_documento_soporte", "nota_ajuste_soporte", "ajuste_documento_soporte":
		return "nota_ajuste_documento_soporte"
	case "nomina", "nomina_electronica", "documento_soporte_nomina", "documento_soporte_pago_nomina", "documento_soporte_de_pago_nomina", "documento_soporte_pago_nomina_electronica", "documento_soporte_de_pago_nomina_electronica":
		return "nomina_electronica"
	case "nota_ajuste_nomina", "nota_ajuste_nomina_electronica", "ajuste_nomina_electronica":
		return "nota_ajuste_nomina_electronica"
	case "pos", "pos_electronico", "tiquete_pos", "tiquete_maquina_registradora_pos", "tiquete_de_maquina_registradora_pos", "documento_equivalente", "documento_equivalente_pos", "documento_equivalente_electronico_pos":
		return "documento_equivalente_pos"
	case "nota_ajuste_documento_equivalente", "nota_ajuste_equivalente", "ajuste_documento_equivalente":
		return "nota_ajuste_documento_equivalente"
	case "factura_talonario", "factura_papel_contingencia", "talonario_contingencia", "factura_talonario_contingencia":
		return "factura_talonario_contingencia"
	case "eventos_radian", "evento_radian", "radian", "eventos_radian_recepcion":
		return "eventos_radian_recepcion"
	default:
		return v
	}
}

func facturacionDocumentoElectronicoPermitido(tipo string) bool {
	normalized := normalizeFacturacionDocumentoElectronicoTipo(tipo)
	for _, item := range dbpkg.ListFacturacionDianDocumentosElectronicos() {
		if item.Codigo == normalized {
			return true
		}
	}
	return false
}

// facturacionDocumentoElectronicoDIANUBLVentaSoportado limita el transporte
// implementado al anexo tecnico de factura electronica de venta. Documento
// soporte, nomina, documentos equivalentes y RADIAN tienen anexos, esquemas y
// servicios propios; nunca deben caer por defecto en un Invoice UBL 2.1.
func facturacionDocumentoElectronicoDIANUBLVentaSoportado(tipo string) bool {
	switch normalizeFacturacionDocumentoElectronicoTipo(tipo) {
	case "factura_electronica", "nota_credito", "nota_debito":
		return true
	default:
		return false
	}
}

// facturacionDocumentoElectronicoDIANTransporteSoportado lists only document
// families with a complete production generator and DIAN transport adapter.
// Documento soporte uses Invoice UBL 2.1 but a distinct Annex 1.1 profile.
func facturacionDocumentoElectronicoDIANTransporteSoportado(tipo string) bool {
	switch normalizeFacturacionDocumentoElectronicoTipo(tipo) {
	case "factura_electronica", "nota_credito", "nota_debito", "documento_soporte", "nomina_electronica":
		return true
	default:
		return false
	}
}

// facturacionDocumentoElectronicoDIANComercialSoportado is stricter than the
// habilitation UBL fixture catalog. Invoices use the immutable paid-cart source;
// credit notes are admitted only after the total-cancellation flow derives a
// separate immutable adjustment source from an accepted invoice. Generic or
// partial notes and every other family remain closed.
func facturacionDocumentoElectronicoDIANComercialSoportado(tipo string) bool {
	switch normalizeFacturacionDocumentoElectronicoTipo(tipo) {
	case "factura_electronica", "nota_credito", "documento_soporte":
		return true
	default:
		return false
	}
}

func facturacionDocumentoElectronicoDIANFinalizacionLocalSoportada(tipo string) bool {
	return facturacionDocumentoElectronicoDIANComercialSoportado(tipo) || normalizeFacturacionDocumentoElectronicoTipo(tipo) == "nomina_electronica"
}

func facturacionMarcarDisponibilidadFuenteFiscal(items []dbpkg.EmpresaDocumentoFacturacionListado, fuentes []dbpkg.EmpresaFacturacionFuenteFiscalRef) {
	disponibles := make(map[string]struct{}, len(fuentes))
	for _, fuente := range fuentes {
		tipo := normalizeFacturacionDocumentoElectronicoTipo(fuente.TipoDocumento)
		codigo := strings.ToUpper(strings.TrimSpace(fuente.DocumentoCodigo))
		if tipo != "" && codigo != "" {
			disponibles[tipo+"\x00"+codigo] = struct{}{}
		}
	}
	for idx := range items {
		item := &items[idx]
		tipoFuente := normalizeFacturacionDocumentoElectronicoTipo(item.TipoDocumento)
		codigoFuente := strings.TrimSpace(item.DocumentoCodigo)
		switch tipoFuente {
		case "factura_electronica":
			tipoFuente = "comprobante_pago"
			codigoFuente = buildVentaDocumentoCodigoFromBase(codigoFuente, "comprobante_pago")
		case "nota_credito", "comprobante_pago", "documento_soporte":
		default:
			item.FuenteFiscalDisponible = false
			continue
		}
		_, item.FuenteFiscalDisponible = disponibles[tipoFuente+"\x00"+strings.ToUpper(strings.TrimSpace(codigoFuente))]
	}
}

// The generic mutation endpoint accepts no commercial fiscal document. Invoices
// originate in paid sales and credit notes in the total-cancellation workflow;
// both resolve their immutable source server-side before numbering or dispatch.
func facturacionDocumentoElectronicoDIANCreacionGenericaSoportada(tipo string) bool {
	// Toda factura comercial debe nacer de una venta pagada con fuente fiscal
	// inmutable. Mantener este endpoint cerrado evita que un payload libre reserve
	// consecutivos antes de demostrar el origen operativo del documento.
	return false
}

// facturacionValidarConfiguracionDIANDocumento permits only preparation of a
// separate configuration. It deliberately refuses to mark a non-UBL family as
// active until its own generator, signature, transport and acknowledgement are
// available. Invoice/credit/debit continue using their established config.
func facturacionValidarConfiguracionDIANDocumento(tipo, estado string) error {
	normalized := normalizeFacturacionDocumentoElectronicoTipo(tipo)
	if !facturacionDocumentoElectronicoPermitido(normalized) {
		return fmt.Errorf("tipo_documento DIAN no reconocido")
	}
	if facturacionDocumentoElectronicoDIANUBLVentaSoportado(normalized) {
		return fmt.Errorf("la configuracion de factura y notas de venta usa su configuracion DIAN existente")
	}
	if normalized == "documento_soporte" || normalized == "nomina_electronica" {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(estado)) {
	case "", "inactivo", "configurando":
		return nil
	default:
		return fmt.Errorf("el documento DIAN aun no puede activarse: falta adaptador, firma, transporte y acuse propios")
	}
}

func facturacionDocumentoElectronicoBloqueoMotivo(tipo string) string {
	normalized := normalizeFacturacionDocumentoElectronicoTipo(tipo)
	if normalized == "" {
		normalized = "desconocido"
	}
	return "tipo_documento=" + normalized + " no dispone aun de un adaptador DIAN conforme a su anexo tecnico; no se genero XML, no se consumio consecutivo y no se envio informacion fiscal"
}

func facturacionFiltrarDocumentosDianOperativos(items []string) []string {
	out := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		normalized := normalizeFacturacionDocumentoElectronicoTipo(item)
		emisionOperativa := false
		for _, catalogItem := range dbpkg.ListFacturacionDianDocumentosElectronicos() {
			if catalogItem.Codigo == normalized && catalogItem.DisponibleEmision {
				emisionOperativa = true
				break
			}
		}
		if !emisionOperativa {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func facturacionDocumentoTipoFromAction(actionRaw string) string {
	action := normalizeDocumentoState(actionRaw)
	switch action {
	case "nota_credito", "emitir_nota_credito":
		return "nota_credito"
	case "nota_debito", "emitir_nota_debito":
		return "nota_debito"
	case "documento_soporte", "emitir_documento_soporte":
		return "documento_soporte"
	case "nomina_electronica", "emitir_nomina_electronica":
		return "nomina_electronica"
	case "documento_equivalente_pos", "emitir_documento_equivalente_pos":
		return "documento_equivalente_pos"
	default:
		if strings.HasPrefix(action, "emitir_") {
			action = strings.TrimPrefix(action, "emitir_")
		}
		docType := normalizeFacturacionDocumentoElectronicoTipo(action)
		if facturacionDocumentoElectronicoPermitido(docType) {
			return docType
		}
		return ""
	}
}

func facturacionDocumentoEntidad(tipo string) string {
	switch normalizeFacturacionDocumentoElectronicoTipo(tipo) {
	case "factura_electronica":
		return "factura_electronica"
	case "nota_credito":
		return "nota_credito"
	case "nota_debito":
		return "nota_debito"
	case "documento_soporte":
		return "documento_soporte"
	case "nomina_electronica":
		return "nomina_electronica"
	case "documento_equivalente_pos":
		return "documento_equivalente_pos"
	default:
		normalized := normalizeFacturacionDocumentoElectronicoTipo(tipo)
		if facturacionDocumentoElectronicoPermitido(normalized) {
			return normalized
		}
		return "documento_electronico"
	}
}

func facturacionActionRequiresFiscalIntegration(action string) bool {
	actionNormalized := normalizeDocumentoState(action)
	switch actionNormalized {
	case "emitir", "anular", "nota_credito", "emitir_nota_credito", "nota_debito", "emitir_nota_debito", "documento_soporte", "emitir_documento_soporte", "nomina_electronica", "emitir_nomina_electronica", "documento_equivalente_pos", "emitir_documento_equivalente_pos":
		return true
	default:
		if strings.HasPrefix(actionNormalized, "emitir_") {
			actionNormalized = strings.TrimPrefix(actionNormalized, "emitir_")
		}
		return facturacionDocumentoElectronicoPermitido(actionNormalized)
	}
}

func facturacionActionIsPaisConfig(action string) bool {
	switch normalizeDocumentoState(action) {
	case "", "config_pais", "guardar_config_pais", "configuracion_pais":
		return true
	default:
		return false
	}
}

func facturacionTryParseJSONMap(raw string) map[string]interface{} {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]interface{}{}
	}
	var out map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return map[string]interface{}{}
	}
	if out == nil {
		return map[string]interface{}{}
	}
	return out
}

func facturacionAnyToBool(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case float64:
		return t > 0
	case int:
		return t > 0
	case int64:
		return t > 0
	case string:
		s := strings.ToLower(strings.TrimSpace(t))
		return s == "1" || s == "true" || s == "si" || s == "yes" || s == "on"
	default:
		return false
	}
}

func facturacionStringListFromAny(v interface{}) []string {
	out := []string{}
	seen := map[string]struct{}{}
	appendOne := func(raw string) {
		item := strings.TrimSpace(raw)
		if item == "" {
			return
		}
		if _, ok := seen[item]; ok {
			return
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	switch t := v.(type) {
	case []string:
		for _, item := range t {
			appendOne(item)
		}
	case []interface{}:
		for _, item := range t {
			appendOne(fmt.Sprintf("%v", item))
		}
	case string:
		for _, item := range strings.Split(t, ",") {
			appendOne(item)
		}
	}
	return out
}

func facturacionDianOfflineSettingsFromConfig(cfg *dbpkg.FacturacionElectronicaPaisConfig) facturacionDianOfflineSettings {
	settings := facturacionDianOfflineSettings{
		Enabled:           false,
		AskBeforeContinue: false,
		AutoRetry:         false,
		ContingencyType:   "servicio_dian",
	}
	if cfg == nil {
		return settings
	}
	extra := facturacionTryParseJSONMap(cfg.CamposPaisJSON)
	if raw := strings.TrimSpace(fmt.Sprintf("%v", extra["dian_contingencia_tipo"])); raw != "" && raw != "<nil>" {
		settings.ContingencyType = strings.ToLower(raw)
	}
	if settings.ContingencyType == "" {
		settings.ContingencyType = "servicio_dian"
	}
	return settings
}

func facturacionIsConnectivityHTTPStatus(status int) bool {
	switch status {
	case http.StatusRequestTimeout, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func facturacionConnectivityMessage(base string) string {
	base = strings.TrimSpace(base)
	if base == "" {
		base = "no hay internet o no se detecta el servidor de la DIAN/proveedor"
	}
	return base
}

func facturacionTruncate(raw string, max int) string {
	raw = strings.TrimSpace(raw)
	if max <= 0 || len(raw) <= max {
		return raw
	}
	return strings.TrimSpace(raw[:max])
}

func facturacionExtractReferenciaExterna(raw string) string {
	m := facturacionTryParseJSONMap(raw)
	keys := []string{"referencia_externa", "external_reference", "reference", "id", "uuid", "codigo", "tracking_id"}
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s := strings.TrimSpace(fmt.Sprintf("%v", v)); s != "" && s != "<nil>" {
				return s
			}
		}
	}
	return ""
}

func dispatchFacturacionProveedorHTTP(url string, payload map[string]interface{}) facturacionProveedorDispatchResult {
	url = normalizePublicIntegracionEndpoint(url)
	if url == "" {
		return facturacionProveedorDispatchResult{Success: false, Error: "endpoint de proveedor fiscal no permitido", ConnectivityFailure: true}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return facturacionProveedorDispatchResult{Success: false, Error: "no se pudo serializar request de integracion"}
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return facturacionProveedorDispatchResult{Success: false, Error: "no se pudo construir request de integracion"}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := publicOutboundHTTPClientForEndpoint(8*time.Second, url)
	resp, err := client.Do(req)
	if err != nil {
		if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
			return facturacionProveedorDispatchResult{Success: false, Error: "timeout de comunicacion con proveedor fiscal", ConnectivityFailure: true}
		}
		return facturacionProveedorDispatchResult{Success: false, Error: "fallo de comunicacion con proveedor fiscal", ConnectivityFailure: true}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	rawResp := strings.TrimSpace(string(respBody))
	if rawResp == "" {
		rawResp = "{}"
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		ref := facturacionExtractReferenciaExterna(rawResp)
		if ref == "" {
			ref = strings.TrimSpace(resp.Header.Get("X-Referencia-Externa"))
		}
		return facturacionProveedorDispatchResult{
			Success:           true,
			Pending:           true, // HTTP delivery alone is not fiscal acceptance.
			ReferenciaExterna: ref,
			RespuestaJSON:     rawResp,
		}
	}

	statusMsg := fmt.Sprintf("proveedor fiscal respondio HTTP %d", resp.StatusCode)
	if rawResp != "" && rawResp != "{}" {
		statusMsg += ": " + facturacionTruncate(rawResp, 280)
	}
	return facturacionProveedorDispatchResult{
		Success:             false,
		RespuestaJSON:       rawResp,
		Error:               statusMsg,
		ConnectivityFailure: facturacionIsConnectivityHTTPStatus(resp.StatusCode),
		HTTPStatus:          resp.StatusCode,
	}
}

func facturacionDIANLegacySignedXMLNeedsManualRegeneration(xmlPayload, documentoTipo string) bool {
	xmlPayload = strings.TrimSpace(xmlPayload)
	if xmlPayload == "" {
		return false
	}
	expectedRoot, _, _, _, _, _, _, _ := dianDocumentKind(documentoTipo)
	profileKind := expectedRoot
	if normalizeFacturacionDocumentoElectronicoTipo(documentoTipo) == "documento_soporte" {
		// Documento soporte usa el elemento UBL Invoice, aunque su familia
		// logica se conserve separada para ProfileID, CUDS y transporte.
		expectedRoot = "Invoice"
		profileKind = "DocumentoSoporte"
	}
	if expectedRoot == "" {
		return false
	}
	values, actualRoot, err := parseDIANXMLTextValues(xmlPayload)
	if err != nil || actualRoot != expectedRoot {
		return false
	}
	if dianXMLFirst(values, "ProfileID") != dianDocumentProfileID(profileKind) {
		return true
	}
	return !strings.Contains(xmlPayload, "<cac:StandardItemIdentification>")
}

func prepareFacturacionDIANDispatchSource(dbEmp *sql.DB, doc dbpkg.EmpresaDocumentoFacturacion, payload facturacionOperacionPayload, documentoTipo string, dianCfg map[string]interface{}) (facturacionOperacionPayload, map[string]interface{}, *facturacionFuenteFiscalSnapshot, *dbpkg.EmpresaDocumentoSoporteElectronico, error) {
	fuenteFiscal, err := loadFacturacionFuenteFiscalParaDocumento(context.Background(), dbEmp, doc)
	if err != nil {
		return payload, dianCfg, nil, nil, fmt.Errorf("fuente fiscal real no disponible: %w", err)
	}
	if documentoTipo != "documento_soporte" {
		return payload, dianCfg, fuenteFiscal, nil, nil
	}
	soporteContable, soporteConfig, err := loadDocumentoSoporteParaDispatch(context.Background(), dbEmp, doc)
	if err != nil {
		return payload, dianCfg, nil, nil, fmt.Errorf("trazabilidad de documento soporte no disponible: %w", err)
	}
	payload.ClienteNombre = fuenteFiscal.Cliente.NombreRazonSocial
	payload.ClienteNumeroDocumento = fuenteFiscal.Cliente.NumeroDocumento
	payload.ClienteTipoDocumento = fuenteFiscal.Cliente.TipoDocumento
	payload.ClienteEmail = fuenteFiscal.Cliente.Email
	payload.ClienteTelefono = fuenteFiscal.Cliente.Telefono
	payload.ClienteDireccion = fuenteFiscal.Cliente.Direccion
	return payload, documentoSoporteMergeDIANConfig(dianCfg, soporteConfig), fuenteFiscal, soporteContable, nil
}

func updateDocumentoSoporteDispatchMirror(dbEmp *sql.DB, doc dbpkg.EmpresaDocumentoFacturacion, soporte *dbpkg.EmpresaDocumentoSoporteElectronico, estado string, ok bool, envioResp map[string]interface{}, xmlFirmado, safeJSON string) error {
	if soporte == nil {
		return nil
	}
	cuds := facturacionCUFEOficialDesdeMap(envioResp)
	if cuds == "" {
		cuds = facturacionCodigoValidacionDesdeXML(xmlFirmado)
	}
	mirrorState := documentoSoporteEstadoDIANMirror(estado)
	if !ok && estado == "" {
		mirrorState = "fallido"
	}
	return dbpkg.UpdateEmpresaDocumentoSoporteDIANResultContext(context.Background(), dbEmp, doc.EmpresaID, soporte.ID, mirrorState, cuds, safeJSON, true)
}

func facturacionDIANDispatchTotals(doc dbpkg.EmpresaDocumentoFacturacion, payload facturacionOperacionPayload, soporte *dbpkg.EmpresaDocumentoSoporteElectronico) (float64, float64) {
	if soporte != nil {
		return soporte.Total, soporte.IVA
	}
	return firstPositiveFloat64(doc.MontoTotal, payload.MontoTotal, payload.TotalNeto), firstPositiveFloat64(payload.Impuestos, payload.IVA)
}

func dispatchFacturacionDIANOficial(dbEmp *sql.DB, payload facturacionOperacionPayload, doc dbpkg.EmpresaDocumentoFacturacion, accion, apiBaseURL string, allowLegacySignedXMLRegeneration bool) facturacionProveedorDispatchResult {
	if dbEmp == nil || doc.EmpresaID <= 0 {
		return facturacionProveedorDispatchResult{Success: false, Error: "conexion o empresa invalida para DIAN oficial"}
	}
	completarClientePayloadFacturacion(dbEmp, doc.EmpresaID, &payload, doc)
	documentoTipo := normalizeFacturacionDocumentoElectronicoTipo(facturacionFirstNonBlank(doc.TipoDocumento, payload.TipoDocumento))
	if documentoTipo == "nomina_electronica" {
		return dispatchNominaDIANOficial(dbEmp, payload, doc, apiBaseURL)
	}
	dianCfg, err := getEmpresaDIANConfig(dbEmp, doc.EmpresaID)
	if err != nil || len(dianCfg) == 0 {
		return facturacionProveedorDispatchResult{Success: false, Error: "configuracion DIAN Colombia no disponible"}
	}
	if documentoTipo == "" {
		documentoTipo = "factura_electronica"
	}
	payload, dianCfg, fuenteFiscal, soporteContable, sourceErr := prepareFacturacionDIANDispatchSource(dbEmp, doc, payload, documentoTipo, dianCfg)
	if sourceErr != nil {
		return facturacionProveedorDispatchResult{FinalFailure: true, Success: false, Error: sourceErr.Error()}
	}
	total, impuesto := facturacionDIANDispatchTotals(doc, payload, soporteContable)
	// DIAN valida que la fecha de generacion del UBL coincida con la fecha de la
	// firma XAdES. Una venta pendiente puede reenviarse en un dia posterior a su
	// registro comercial; usar esa fecha historica aqui provoca FAD09e. La fecha
	// comercial permanece en el documento, mientras que la emision fiscal se
	// genera en el instante de la firma.
	fechaEmision := facturacionDIANFechaEmisionFirmada(time.Now())
	moneda := strings.ToUpper(strings.TrimSpace(facturacionFirstNonBlank(doc.Moneda, payload.Moneda, "COP")))
	docPayload := map[string]interface{}{
		"empresa_id":                  doc.EmpresaID,
		"documento_codigo":            strings.TrimSpace(doc.DocumentoCodigo),
		"documento_tipo":              documentoTipo,
		"fecha_emision":               fechaEmision,
		"total":                       fmt.Sprintf("%.2f", total),
		"impuesto_total":              fmt.Sprintf("%.2f", impuesto),
		"moneda":                      moneda,
		"cliente_nombre":              strings.TrimSpace(payload.ClienteNombre),
		"cliente_nit":                 strings.TrimSpace(payload.ClienteNumeroDocumento),
		"cliente_tipo_documento":      strings.TrimSpace(payload.ClienteTipoDocumento),
		"cliente_email":               strings.TrimSpace(payload.ClienteEmail),
		"cliente_telefono":            strings.TrimSpace(payload.ClienteTelefono),
		"cliente_direccion":           strings.TrimSpace(payload.ClienteDireccion),
		"numero_legal":                strings.TrimSpace(doc.NumeroLegal),
		"codigo_validacion":           strings.TrimSpace(doc.CodigoValidacion),
		"referencia_documento_codigo": strings.TrimSpace(payload.ReferenciaDocumentoCodigo),
		"referencia_cufe":             strings.TrimSpace(payload.ReferenciaCUFE),
		"referencia_fecha_emision":    strings.TrimSpace(payload.ReferenciaFechaEmision),
		"codigo_correccion":           strings.TrimSpace(payload.CodigoCorreccion),
		"descripcion_correccion":      strings.TrimSpace(payload.DescripcionCorreccion),
		"resolucion_numero":           strings.TrimSpace(genericStringValue(dianCfg["resolucion_numero"])),
		"prefijo":                     strings.TrimSpace(genericStringValue(dianCfg["prefijo"])),
		"usar_soap_dian":              true,
		"accion_facturacion":          strings.ToLower(strings.TrimSpace(accion)),
	}
	if documentoTipo == "documento_soporte" {
		docPayload["soap_operacion"] = "SendBillSync"
	}
	effectiveEndpoint := strings.TrimSpace(apiBaseURL)
	if documentoTipo == "documento_soporte" {
		effectiveEndpoint = strings.TrimSpace(genericStringValue(dianCfg["url_dian"]))
	}
	if endpoint := effectiveEndpoint; endpoint != "" {
		docPayload["url_dian"] = endpoint
	}
	xmlFirmado := ""
	xmlNuevo := false
	legacyXMLRegenerated := false
	if storedXML, loadErr := loadFacturacionFiscalArtifact(context.Background(), dbEmp, doc, "xml_firmado"); loadErr == nil {
		xmlFirmado = strings.TrimSpace(string(storedXML))
		if _, _, sourceErr := generateDIANUBLBase(dianCfg, doc.EmpresaID, docPayload, fuenteFiscal); sourceErr != nil {
			return facturacionProveedorDispatchResult{Success: false, Error: "fuente fiscal no valida para reintentar XML firmado: " + sourceErr.Error()}
		}
		if allowLegacySignedXMLRegeneration && facturacionDIANLegacySignedXMLNeedsManualRegeneration(xmlFirmado, documentoTipo) {
			// Solo la accion manual autorizada puede reemplazar un XML rechazado y
			// obsoleto. La fuente fiscal sigue siendo la misma e inmutable; se
			// renuevan fecha, UBL y firma para superar el preflight vigente.
			fechaEmision = facturacionDIANFechaEmisionFirmada(time.Now())
			docPayload["fecha_emision"] = fechaEmision
			ublResp, _, generateErr := generateDIANUBLBase(dianCfg, doc.EmpresaID, docPayload, fuenteFiscal)
			if generateErr != nil {
				return facturacionProveedorDispatchResult{Success: false, Error: "regenerar XML UBL DIAN obsoleto: " + generateErr.Error()}
			}
			docPayload["xml_ubl_base"] = genericStringValue(ublResp["xml_ubl_base"])
			signResp, _, signErr := signDIANXMLXAdESBase(dianCfg, doc.EmpresaID, docPayload)
			if signErr != nil {
				return facturacionProveedorDispatchResult{Success: false, Error: "firmar XML DIAN regenerado: " + signErr.Error()}
			}
			xmlFirmado = genericStringValue(signResp["xml_firmado"])
			xmlNuevo = true
			legacyXMLRegenerated = true
		} else if storedFecha := facturacionDIANFechaEmisionDesdeXML(xmlFirmado); storedFecha != "" {
			fechaEmision = storedFecha
			docPayload["fecha_emision"] = storedFecha
		}
	} else if !errors.Is(loadErr, sql.ErrNoRows) {
		return facturacionProveedorDispatchResult{Success: false, Error: "leer XML fiscal persistido: " + loadErr.Error()}
	}
	if xmlFirmado == "" {
		xmlNuevo = true
		ublResp, _, err := generateDIANUBLBase(dianCfg, doc.EmpresaID, docPayload, fuenteFiscal)
		if err != nil {
			return facturacionProveedorDispatchResult{Success: false, Error: "generar XML UBL DIAN: " + err.Error()}
		}
		docPayload["xml_ubl_base"] = genericStringValue(ublResp["xml_ubl_base"])
		signResp, _, err := signDIANXMLXAdESBase(dianCfg, doc.EmpresaID, docPayload)
		if err != nil {
			return facturacionProveedorDispatchResult{Success: false, Error: "firmar XML DIAN: " + err.Error()}
		}
		xmlFirmado = genericStringValue(signResp["xml_firmado"])
	}
	preflight := validateDIANDocumentPreflight(dianCfg, doc.EmpresaID, docPayload, xmlFirmado, "envio_real")
	if parseTruthy(genericStringValue(preflight["bloqueado"])) {
		raw, _ := json.Marshal(preflight)
		return facturacionProveedorDispatchResult{Success: false, Error: "validacion preventiva DIAN no superada", RespuestaJSON: string(raw)}
	}
	if xmlNuevo {
		if _, err := saveFacturacionFiscalArtifact(context.Background(), dbEmp, doc, "xml_firmado", ".xml", "application/xml", []byte(xmlFirmado)); err != nil {
			return facturacionProveedorDispatchResult{Success: false, Error: "persistir XML firmado antes del envio: " + err.Error()}
		}
	}
	envioPayload := map[string]interface{}{
		"empresa_id":             doc.EmpresaID,
		"documento_codigo":       strings.TrimSpace(doc.DocumentoCodigo),
		"documento_tipo":         documentoTipo,
		"numero_legal":           strings.TrimSpace(doc.NumeroLegal),
		"xml_firmado":            xmlFirmado,
		"total":                  fmt.Sprintf("%.2f", total),
		"impuesto_total":         fmt.Sprintf("%.2f", impuesto),
		"fecha_emision":          fechaEmision,
		"moneda":                 moneda,
		"cliente_nombre":         strings.TrimSpace(payload.ClienteNombre),
		"cliente_nit":            strings.TrimSpace(payload.ClienteNumeroDocumento),
		"cliente_tipo_documento": strings.TrimSpace(payload.ClienteTipoDocumento),
		"cliente_email":          strings.TrimSpace(payload.ClienteEmail),
		"cliente_telefono":       strings.TrimSpace(payload.ClienteTelefono),
		"cliente_direccion":      strings.TrimSpace(payload.ClienteDireccion),
		"codigo_validacion":      strings.TrimSpace(doc.CodigoValidacion),
		"resolucion_numero":      strings.TrimSpace(genericStringValue(dianCfg["resolucion_numero"])),
		"prefijo":                strings.TrimSpace(genericStringValue(dianCfg["prefijo"])),
		"usar_soap_dian":         true,
	}
	if documentoTipo == "documento_soporte" {
		envioPayload["soap_operacion"] = "SendBillSync"
	}
	if endpoint := effectiveEndpoint; endpoint != "" {
		envioPayload["url_dian"] = endpoint
	}
	envioResp, _, err := sendDIANDocumentoReal(dbEmp, dianCfg, doc.EmpresaID, envioPayload)
	if err != nil {
		return facturacionProveedorDispatchResult{Success: false, Error: err.Error()}
	}
	safeJSON := facturacionSafeDispatchJSON(envioResp)
	artifactWarning := ""
	if legacyXMLRegenerated {
		artifactWarning = "XML firmado obsoleto regenerado desde la fuente fiscal inmutable por reenvio manual."
	}
	providerContent := []byte(strings.TrimSpace(genericStringValue(envioResp["raw_response"])))
	providerExtension := ".xml"
	providerMime := "application/xml"
	if len(providerContent) == 0 {
		if encoded, encodeErr := json.Marshal(envioResp["respuesta_dian"]); encodeErr == nil {
			providerContent = encoded
		}
		providerExtension = ".json"
		providerMime = "application/json"
	}
	if len(providerContent) > 0 && string(providerContent) != "null" {
		if _, saveErr := saveFacturacionFiscalArtifact(context.Background(), dbEmp, doc, "respuesta_proveedor", providerExtension, providerMime, providerContent); saveErr != nil {
			artifactWarning = "DIAN respondio, pero no se pudo persistir el acuse privado: " + saveErr.Error()
		}
	}
	estado := strings.ToLower(strings.TrimSpace(genericStringValue(envioResp["estado_dian"])))
	trackID := strings.TrimSpace(genericStringValue(envioResp["track_id"]))
	ok := parseTruthy(genericStringValue(envioResp["ok"])) || trackID != "" || estado == "enviado" || estado == "aceptado"
	if updateErr := updateDocumentoSoporteDispatchMirror(dbEmp, doc, soporteContable, estado, ok, envioResp, xmlFirmado, safeJSON); updateErr != nil {
		return facturacionProveedorDispatchResult{Success: false, Error: "DIAN respondio, pero no se pudo actualizar el documento soporte contable", RespuestaJSON: safeJSON}
	}
	if !ok {
		errMsg := dianFirstNonBlank(genericStringValue(envioResp["acuse_mensaje"]), genericStringValue(envioResp["error"]), genericStringValue(envioResp["mensaje_recepcion"]), "DIAN no acepto el documento")
		if facturacionDianDocumentoProcesadoAnteriormente(errMsg) {
			return facturacionProveedorDispatchResult{Success: false, FinalFailure: true, Error: "documento procesado anteriormente por DIAN; consulte el acuse original antes de marcarlo como aceptado", RespuestaJSON: safeJSON, ArtifactWarning: artifactWarning, HTTPStatus: int(anyToInt64(envioResp["http_status"]))}
		}
		finalFailure := estado == "rechazado" || strings.EqualFold(genericStringValue(envioResp["acuse_estado"]), "rechazado")
		return facturacionProveedorDispatchResult{Success: false, FinalFailure: finalFailure, Error: errMsg, RespuestaJSON: safeJSON, ArtifactWarning: artifactWarning, HTTPStatus: int(anyToInt64(envioResp["http_status"]))}
	}
	ref := trackID
	if ref == "" {
		ref = strings.TrimSpace(genericStringValue(envioResp["zip_key"]))
	}
	if ref == "" {
		ref = strings.TrimSpace(genericStringValue(envioResp["referencia_externa"]))
	}
	if cufe := facturacionCUFEOficialDesdeMap(envioResp); cufe != "" {
		doc.CodigoValidacion = cufe
	}
	if pdf := buildFacturaElectronicaRepresentationPDF(doc, payload); len(pdf) > 0 {
		if _, saveErr := saveFacturacionFiscalArtifact(context.Background(), dbEmp, doc, "representacion_pdf", ".pdf", "application/pdf", pdf); saveErr != nil {
			artifactWarning = strings.TrimSpace(facturacionFirstNonBlank(artifactWarning, "No se pudo persistir la representacion PDF: "+saveErr.Error()))
		}
	}
	return facturacionProveedorDispatchResult{Success: true, Pending: estado != "aceptado", ReferenciaExterna: ref, RespuestaJSON: safeJSON, ArtifactWarning: artifactWarning, HTTPStatus: int(anyToInt64(envioResp["http_status"]))}
}

func facturacionDIANFechaEmisionFirmada(now time.Time) string {
	return now.In(dianColombiaLocation()).Format("2006-01-02T15:04:05-07:00")
}

func facturacionDIANFechaEmisionDesdeXML(xmlFirmado string) string {
	values, _, err := parseDIANXMLTextValues(xmlFirmado)
	if err != nil {
		return ""
	}
	issueDate := strings.TrimSpace(dianXMLFirst(values, "IssueDate"))
	issueTime := strings.TrimSpace(dianXMLFirst(values, "IssueTime"))
	if issueDate == "" {
		return ""
	}
	if issueTime == "" {
		return issueDate
	}
	return issueDate + "T" + strings.TrimPrefix(issueTime, "T")
}

func facturacionCodigoValidacionDesdeXML(xmlFirmado string) string {
	values, _, err := parseDIANXMLTextValues(xmlFirmado)
	if err != nil {
		return ""
	}
	value := strings.TrimSpace(dianXMLFirst(values, "UUID"))
	if !facturacionCodigoSHA384Valido(value) {
		return ""
	}
	return strings.ToLower(value)
}

func dispatchFacturacionDIANAcusePendiente(dbEmp *sql.DB, doc dbpkg.EmpresaDocumentoFacturacion, payload facturacionOperacionPayload, trackID string) facturacionProveedorDispatchResult {
	dianCfg, err := getEmpresaDIANConfig(dbEmp, doc.EmpresaID)
	if err != nil || len(dianCfg) == 0 {
		return facturacionProveedorDispatchResult{Error: "configuracion DIAN no disponible para consultar acuse"}
	}
	var soporteContable *dbpkg.EmpresaDocumentoSoporteElectronico
	var nominaContable *dbpkg.EmpresaNominaElectronica
	if normalizeFacturacionDocumentoElectronicoTipo(doc.TipoDocumento) == "documento_soporte" {
		var soporteConfig *dbpkg.EmpresaDocumentoSoporteConfiguracionSnapshot
		soporteContable, soporteConfig, err = loadDocumentoSoporteParaDispatch(context.Background(), dbEmp, doc)
		if err != nil {
			return facturacionProveedorDispatchResult{Error: "trazabilidad de documento soporte no disponible: " + err.Error()}
		}
		dianCfg = documentoSoporteMergeDIANConfig(dianCfg, soporteConfig)
	}
	if normalizeFacturacionDocumentoElectronicoTipo(doc.TipoDocumento) == "nomina_electronica" {
		var nominaConfig *dbpkg.EmpresaNominaDIANConfiguracionSnapshot
		nominaContable, _, nominaConfig, err = loadNominaElectronicaParaDispatch(context.Background(), dbEmp, doc)
		if err != nil {
			return facturacionProveedorDispatchResult{Error: "trazabilidad de nómina electrónica no disponible: " + err.Error()}
		}
		dianCfg = nominaElectronicaMergeDIANConfig(dianCfg, nominaConfig)
	}
	endpoint := normalizeDIANSOAPEndpoint(dianConfiguredEndpoint(dianCfg, nil))
	if endpoint == "" {
		return facturacionProveedorDispatchResult{Error: "endpoint DIAN no disponible para consultar acuse"}
	}
	token := ""
	if resolved, resolveErr := dianResolveOptionalReference(genericStringValue(dianCfg["token_emisor_ref"]), doc.EmpresaID); resolveErr == nil {
		token = resolved
	}
	response, _, err := consultarDIANStatusZipSOAP(dbEmp, dianCfg, doc.EmpresaID, endpoint, trackID, token)
	if err != nil {
		return facturacionProveedorDispatchResult{Error: err.Error(), ConnectivityFailure: true, ReferenciaExterna: trackID}
	}
	safeJSON := facturacionSafeDispatchJSON(response)
	artifactWarning := ""
	providerContent := []byte(strings.TrimSpace(genericStringValue(response["raw_response"])))
	if len(providerContent) > 0 {
		if _, saveErr := saveFacturacionFiscalArtifact(context.Background(), dbEmp, doc, "respuesta_proveedor", ".xml", "application/xml", providerContent); saveErr != nil {
			artifactWarning = "No se pudo actualizar el acuse privado DIAN: " + saveErr.Error()
		}
	}
	estado := strings.ToLower(strings.TrimSpace(genericStringValue(response["estado_dian"])))
	acuse := strings.ToLower(strings.TrimSpace(genericStringValue(response["acuse_estado"])))
	if soporteContable != nil {
		soporteEstado := documentoSoporteEstadoDIANMirror(dianFirstNonBlank(estado, acuse))
		if updateErr := dbpkg.UpdateEmpresaDocumentoSoporteDIANResultContext(context.Background(), dbEmp, doc.EmpresaID, soporteContable.ID, soporteEstado, facturacionCUFEOficialDesdeMap(response), safeJSON, true); updateErr != nil {
			return facturacionProveedorDispatchResult{Error: "acuse DIAN recibido, pero no se pudo actualizar el documento soporte", ReferenciaExterna: trackID}
		}
	}
	if nominaContable != nil {
		nominaEstado := documentoSoporteEstadoDIANMirror(dianFirstNonBlank(estado, acuse))
		if updateErr := dbpkg.UpdateEmpresaNominaDIANResultContext(context.Background(), dbEmp, doc.EmpresaID, nominaContable.ID, nominaEstado, facturacionCUFEOficialDesdeMap(response), safeJSON, true, false); updateErr != nil {
			return facturacionProveedorDispatchResult{Error: "acuse DIAN recibido, pero no se pudo actualizar la nómina electrónica", ReferenciaExterna: trackID}
		}
	}
	if estado == "aceptado" || acuse == "aceptado" {
		if cufe := facturacionCUFEOficialDesdeMap(response); cufe != "" {
			doc.CodigoValidacion = cufe
		}
		if nominaContable == nil {
			if pdf := buildFacturaElectronicaRepresentationPDF(doc, payload); len(pdf) > 0 {
				if _, saveErr := saveFacturacionFiscalArtifact(context.Background(), dbEmp, doc, "representacion_pdf", ".pdf", "application/pdf", pdf); saveErr != nil && artifactWarning == "" {
					artifactWarning = "No se pudo actualizar la representacion PDF: " + saveErr.Error()
				}
			}
		}
		return facturacionProveedorDispatchResult{Success: true, ReferenciaExterna: trackID, RespuestaJSON: safeJSON, ArtifactWarning: artifactWarning}
	}
	if estado == "rechazado" || acuse == "rechazado" {
		return facturacionProveedorDispatchResult{FinalFailure: true, ReferenciaExterna: trackID, RespuestaJSON: safeJSON, ArtifactWarning: artifactWarning, Error: dianFirstNonBlank(genericStringValue(response["acuse_mensaje"]), "DIAN rechazo el documento")}
	}
	if estado == "contingencia" || acuse == "contingencia" {
		return facturacionProveedorDispatchResult{ReferenciaExterna: trackID, RespuestaJSON: safeJSON, ArtifactWarning: artifactWarning, Error: dianFirstNonBlank(genericStringValue(response["error"]), genericStringValue(response["acuse_mensaje"]), "no fue posible consultar el acuse DIAN"), ConnectivityFailure: true}
	}
	return facturacionProveedorDispatchResult{Success: true, Pending: true, ReferenciaExterna: trackID, RespuestaJSON: safeJSON, ArtifactWarning: artifactWarning}
}

func dispatchFacturacionProveedor(dbEmp *sql.DB, cfg *dbpkg.FacturacionElectronicaPaisConfig, payload facturacionOperacionPayload, doc dbpkg.EmpresaDocumentoFacturacion, accion string, allowLegacySignedXMLRegeneration bool) facturacionProveedorDispatchResult {
	// Recheck the immutable document boundary at the final transport hop. A
	// caller or queued job must never dispatch with another tenant's settings.
	if cfg == nil || doc.EmpresaID <= 0 || cfg.EmpresaID != doc.EmpresaID ||
		(payload.EmpresaID != 0 && payload.EmpresaID != doc.EmpresaID) {
		return facturacionProveedorDispatchResult{FinalFailure: true, Error: "configuracion, documento y solicitud fiscal deben pertenecer a la misma empresa"}
	}
	if strings.TrimSpace(doc.PaisCodigo) == "" || !strings.EqualFold(strings.TrimSpace(cfg.PaisCodigo), strings.TrimSpace(doc.PaisCodigo)) ||
		(strings.TrimSpace(payload.PaisCodigo) != "" && !strings.EqualFold(strings.TrimSpace(payload.PaisCodigo), strings.TrimSpace(doc.PaisCodigo))) {
		return facturacionProveedorDispatchResult{FinalFailure: true, Error: "pais fiscal del documento no coincide con la configuracion o solicitud"}
	}
	if !strings.EqualFold(strings.TrimSpace(doc.AmbienteFE), "produccion") {
		return facturacionProveedorDispatchResult{FinalFailure: true, Error: "un documento de pruebas no puede transmitirse en produccion"}
	}
	if !strings.EqualFold(strings.TrimSpace(cfg.Estado), "activo") {
		return facturacionProveedorDispatchResult{FinalFailure: true, Error: "configuracion fiscal inactiva: transmision bloqueada"}
	}
	proveedor := "manual"
	ambiente := "sandbox"
	apiBaseURL := ""
	camposPaisJSON := "{}"
	paisCodigo := ""

	if cfg != nil {
		if strings.TrimSpace(cfg.Proveedor) != "" {
			proveedor = strings.ToLower(strings.TrimSpace(cfg.Proveedor))
		}
		if strings.TrimSpace(cfg.Ambiente) != "" {
			ambiente = strings.ToLower(strings.TrimSpace(cfg.Ambiente))
		}
		apiBaseURL = strings.TrimSpace(cfg.APIBaseURL)
		camposPaisJSON = strings.TrimSpace(cfg.CamposPaisJSON)
		paisCodigo = strings.ToUpper(strings.TrimSpace(cfg.PaisCodigo))
	}

	if paisCodigo == "CO" && !facturacionDocumentoElectronicoDIANComercialSoportado(doc.TipoDocumento) && normalizeFacturacionDocumentoElectronicoTipo(doc.TipoDocumento) != "nomina_electronica" {
		return facturacionProveedorDispatchResult{
			FinalFailure: true,
			Error:        facturacionDocumentoElectronicoBloqueoMotivo(doc.TipoDocumento),
		}
	}

	if ambiente != "produccion" {
		return facturacionProveedorDispatchResult{Success: false, Error: "integracion fiscal no aplica fuera de produccion"}
	}
	if proveedor == "manual" || proveedor == "interno" || proveedor == "local" || strings.HasPrefix(strings.ToLower(apiBaseURL), "mock:") {
		return facturacionProveedorDispatchResult{FinalFailure: true, Error: "produccion requiere un proveedor fiscal real; no se permiten proveedores locales ni simulados"}
	}

	camposPais := facturacionTryParseJSONMap(camposPaisJSON)
	if facturacionAnyToBool(camposPais["force_fail"]) || facturacionAnyToBool(camposPais["simular_error"]) {
		return facturacionProveedorDispatchResult{Success: false, Error: "simulacion de fallo de proveedor fiscal"}
	}

	if paisCodigo == "CO" && proveedor == "dian" {
		return dispatchFacturacionDIANOficial(dbEmp, payload, doc, accion, apiBaseURL, allowLegacySignedXMLRegeneration)
	}

	// Country configuration is not an implemented fiscal adapter. Do not send
	// private fiscal data to an arbitrary JSON endpoint or treat its 2xx as an
	// authority acknowledgement. Each adapter must implement both contracts.
	return facturacionProveedorDispatchResult{FinalFailure: true, Error: "emision bloqueada: falta un adaptador fiscal especifico con validacion de acuse para este pais y proveedor"}
}

func facturacionProveedorConnectionStatus(cfg *dbpkg.FacturacionElectronicaPaisConfig) map[string]interface{} {
	settings := facturacionDianOfflineSettingsFromConfig(cfg)
	out := map[string]interface{}{
		"ok":                            true,
		"online":                        false,
		"estado_conexion":               "sin_configuracion",
		"mensaje":                       "configuracion FE no disponible",
		"modo_offline_dian_activo":      settings.Enabled,
		"modo_offline_preguntar":        settings.AskBeforeContinue,
		"modo_offline_auto_reintentar":  settings.AutoRetry,
		"dian_contingencia_tipo":        settings.ContingencyType,
		"accion_recomendada":            "bloquear_facturacion_electronica",
		"requiere_confirmacion_offline": false,
	}
	if cfg == nil {
		return out
	}

	paisCodigo := strings.ToUpper(strings.TrimSpace(cfg.PaisCodigo))
	ambiente := strings.ToLower(strings.TrimSpace(cfg.Ambiente))
	proveedor := strings.ToLower(strings.TrimSpace(cfg.Proveedor))
	apiBaseURL := strings.TrimSpace(cfg.APIBaseURL)
	out["pais_codigo"] = paisCodigo
	out["proveedor"] = strings.TrimSpace(cfg.Proveedor)
	out["ambiente"] = ambiente

	if paisCodigo != "CO" {
		out["estado_conexion"] = "sin_adaptador_fiscal"
		out["mensaje"] = "configuracion por pais disponible; emision bloqueada hasta implementar y validar su adaptador fiscal"
		return out
	}
	if !strings.EqualFold(strings.TrimSpace(cfg.Estado), "activo") {
		out["estado_conexion"] = "configuracion_inactiva"
		out["mensaje"] = "configuracion fiscal inactiva: transmision bloqueada"
		return out
	}
	if ambiente != "produccion" {
		out["online"] = true
		out["estado_conexion"] = "no_aplica"
		out["mensaje"] = "la integracion DIAN no aplica fuera de produccion o esta inactiva"
		out["accion_recomendada"] = "continuar_online"
		return out
	}
	if proveedor == "" || proveedor == "manual" || proveedor == "interno" || proveedor == "local" {
		out["online"] = false
		out["estado_conexion"] = "sin_proveedor_dian"
		out["mensaje"] = "proveedor DIAN real no configurado para Colombia en produccion"
		out["accion_recomendada"] = "bloquear_facturacion_electronica"
		return out
	}
	if proveedor != "dian" {
		out["estado_conexion"] = "sin_adaptador_fiscal"
		out["mensaje"] = "proveedor sin adaptador fiscal validado"
		return out
	}
	if strings.HasPrefix(strings.ToLower(apiBaseURL), "mock://") {
		out["estado_conexion"] = "proveedor_simulado_bloqueado"
		out["mensaje"] = "un proveedor simulado no habilita conectividad fiscal en produccion"
	} else if apiBaseURL == "" {
		out["estado_conexion"] = "sin_endpoint"
		out["mensaje"] = "api_base_url no configurado para proveedor DIAN"
	} else {
		endpoint := normalizePublicIntegracionEndpoint(strings.TrimRight(apiBaseURL, "/"))
		if endpoint == "" {
			out["estado_conexion"] = "sin_endpoint"
			out["mensaje"] = "api_base_url invalido o no permitido para proveedor DIAN"
			out["accion_recomendada"] = "bloquear_facturacion_electronica"
			return out
		}
		req, err := http.NewRequest(http.MethodHead, endpoint, nil)
		if err != nil {
			out["estado_conexion"] = "sin_endpoint"
			out["mensaje"] = "api_base_url invalido para proveedor DIAN"
		} else {
			req.Header.Set("Accept", "application/json")
			resp, err := publicOutboundHTTPClientForEndpoint(4*time.Second, endpoint).Do(req)
			if err != nil {
				if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
					out["mensaje"] = "timeout al detectar servidor DIAN/proveedor"
				} else {
					out["mensaje"] = "no se detecta internet o servidor DIAN/proveedor"
				}
				out["estado_conexion"] = "offline"
			} else {
				defer resp.Body.Close()
				out["http_status"] = resp.StatusCode
				if resp.StatusCode < 500 || resp.StatusCode == http.StatusMethodNotAllowed {
					out["online"] = true
					out["estado_conexion"] = "online"
					out["mensaje"] = "servidor DIAN/proveedor detectado"
					out["accion_recomendada"] = "continuar_online"
					return out
				}
				out["estado_conexion"] = "offline"
				out["mensaje"] = fmt.Sprintf("servidor DIAN/proveedor respondio HTTP %d", resp.StatusCode)
			}
		}
	}

	out["accion_recomendada"] = "bloquear_facturacion_electronica"
	return out
}

func facturacionDIANConnectionStatus(dbEmp *sql.DB, empresaID int64, paisCodigo string, cfg *dbpkg.FacturacionElectronicaPaisConfig) map[string]interface{} {
	status := facturacionProveedorConnectionStatus(cfg)
	if strings.ToUpper(strings.TrimSpace(paisCodigo)) != "CO" || empresaID <= 0 || dbEmp == nil ||
		cfg == nil || cfg.EmpresaID != empresaID || !strings.EqualFold(strings.TrimSpace(cfg.PaisCodigo), "CO") ||
		!strings.EqualFold(strings.TrimSpace(cfg.Proveedor), "dian") || !strings.EqualFold(strings.TrimSpace(cfg.Estado), "activo") {
		return status
	}

	dianCfg, err := getEmpresaDIANConfig(dbEmp, empresaID)
	if err != nil || len(dianCfg) == 0 {
		return status
	}
	endpoint := normalizeDIANSOAPEndpoint(genericStringValue(dianCfg["url_dian"]))
	if endpoint == "" {
		return status
	}

	httpStatus, reachable, latencyMS, message := runIntegracionProbe(endpoint)
	status["ok"] = true
	status["online"] = reachable
	status["estado_conexion"] = map[bool]string{true: "online", false: "offline"}[reachable]
	status["mensaje"] = message
	status["endpoint"] = endpoint
	status["http_status"] = httpStatus
	status["latency_ms"] = latencyMS
	status["proveedor"] = "DIAN"
	status["transporte"] = "soap_dian"
	status["ambiente"] = chooseDIANAmbiente(dianCfg)
	status["estado_dian"] = genericStringValue(dianCfg["estado_dian"])
	status["test_set_id_configurado"] = strings.TrimSpace(genericStringValue(dianCfg["test_set_id"])) != ""
	if reachable {
		status["accion_recomendada"] = "continuar_online"
	} else {
		status["accion_recomendada"] = "revisar_endpoint_dian"
	}
	return status
}

func facturacionOfflineDianPreflight(dbEmp *sql.DB, payload facturacionOperacionPayload) (map[string]interface{}, error) {
	if dbEmp == nil || payload.EmpresaID <= 0 {
		return nil, nil
	}
	paisCodigo := strings.ToUpper(strings.TrimSpace(payload.PaisCodigo))
	if paisCodigo == "" {
		paisDetectado, _, detectErr := dbpkg.DetectFacturacionPais(dbEmp, payload.EmpresaID, "", "")
		if detectErr == nil {
			paisCodigo = strings.ToUpper(strings.TrimSpace(paisDetectado.Codigo))
		}
	}
	if paisCodigo == "" {
		paisCodigo = "CO"
	}
	if paisCodigo != "CO" {
		return nil, nil
	}
	cfg, err := dbpkg.GetFacturacionElectronicaPaisConfig(dbEmp, payload.EmpresaID, paisCodigo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	status := facturacionProveedorConnectionStatus(cfg)
	online, _ := status["online"].(bool)
	if online {
		return nil, nil
	}
	status["ok"] = false
	status["bloqueado"] = true
	status["requiere_confirmacion_offline"] = false
	status["modo_offline_dian_activo"] = false
	status["error"] = "DIAN/proveedor no disponible; se requiere conexion activa para facturar"
	return status, nil
}

func processFacturacionIntegracionForDocumento(dbEmp *sql.DB, payload facturacionOperacionPayload, doc dbpkg.EmpresaDocumentoFacturacion, accion, usuario string, dbSuperOpt ...*sql.DB) (facturacionIntegracionResultado, *dbpkg.FacturacionElectronicaRetryItem, error) {
	return processFacturacionIntegracionForDocumentoContext(context.Background(), dbEmp, payload, doc, accion, usuario, dbSuperOpt...)
}

func facturacionDocumentoAdvisoryLockKey(empresaID int64, tipoDocumento, documentoCodigo string) int64 {
	raw := fmt.Sprintf("facturacion-electronica|%d|%s|%s", empresaID, strings.ToLower(strings.TrimSpace(tipoDocumento)), strings.ToUpper(strings.TrimSpace(documentoCodigo)))
	// Signed representation of the FNV-1a 64-bit offset basis. Keeping the
	// accumulator signed preserves the same PostgreSQL advisory-lock key while
	// avoiding an unchecked uint64 -> int64 conversion.
	hash := int64(-3750763034362895579)
	for i := 0; i < len(raw); i++ {
		hash ^= int64(raw[i])
		hash *= 1099511628211
	}
	return hash
}

type facturacionDocumentLockContextKey struct{}

// facturacionManualTerminalRetryContextKey solo se activa desde la accion
// empresarial explicita de reenvio. El worker y las emisiones automaticas no
// deben revivir una cola terminal por su cuenta.
type facturacionManualTerminalRetryContextKey struct{}

func facturacionManualTerminalRetryAllowed(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	allowed, _ := ctx.Value(facturacionManualTerminalRetryContextKey{}).(bool)
	return allowed
}

func facturacionPrepareManualTerminalRetry(retry *dbpkg.FacturacionElectronicaRetryItem, doc dbpkg.EmpresaDocumentoFacturacion, usuario string) (*dbpkg.FacturacionElectronicaRetryItem, bool, error) {
	if retry == nil || normalizeFacturacionEstadoEnvio(retry.EstadoEnvio) != "fallido_terminal" {
		return retry, false, nil
	}
	if strings.TrimSpace(retry.ReferenciaExterna) != "" ||
		strings.TrimSpace(retry.CodigoValidacion) != "" ||
		strings.TrimSpace(doc.CodigoValidacion) != "" {
		return retry, false, fmt.Errorf("el documento tiene referencia fiscal; consulta el acuse antes de reenviar")
	}

	reactivated := *retry
	previousAttempts := reactivated.Intentos
	previousError := facturacionTruncate(reactivated.UltimoError, 240)
	reactivated.EstadoEnvio = "fallido"
	reactivated.Intentos = 0
	reactivated.ProximoIntento = ""
	reactivated.FechaUltimoIntento = ""
	reactivated.UltimoError = ""
	reactivated.RespuestaProveedor = ""
	reactivated.ContingenciaActiva = false
	reactivated.FechaContingencia = ""
	reactivated.ReferenciaExterna = ""
	reactivated.Estado = "activo"
	if strings.TrimSpace(usuario) != "" {
		reactivated.UsuarioCreador = strings.TrimSpace(usuario)
	}
	auditLine := fmt.Sprintf("[%s] reactivacion manual DIAN; intentos previos=%d", facturacionNowLocal(), previousAttempts)
	if previousError != "" {
		auditLine += "; error previo=" + previousError
	}
	reactivated.Observaciones = strings.TrimSpace(strings.TrimSpace(reactivated.Observaciones) + "\n" + auditLine)
	return &reactivated, true, nil
}

func acquireFacturacionDocumentAdvisoryLock(ctx context.Context, dbEmp *sql.DB, empresaID int64, tipoDocumento, documentoCodigo string) (context.Context, func(), bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	lockKey := facturacionDocumentoAdvisoryLockKey(empresaID, tipoDocumento, documentoCodigo)
	if heldKey, ok := ctx.Value(facturacionDocumentLockContextKey{}).(int64); ok && heldKey == lockKey {
		return ctx, func() {}, true, nil
	}
	lockConn, err := dbEmp.Conn(ctx)
	if err != nil {
		return ctx, func() {}, false, err
	}
	var locked bool
	if err := lockConn.QueryRowContext(ctx, `SELECT pg_try_advisory_lock($1::bigint)`, lockKey).Scan(&locked); err != nil {
		_ = lockConn.Close()
		return ctx, func() {}, false, err
	}
	if !locked {
		_ = lockConn.Close()
		return ctx, func() {}, false, nil
	}
	var releaseOnce sync.Once
	release := func() {
		releaseOnce.Do(func() {
			releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			var unlocked bool
			if err := lockConn.QueryRowContext(releaseCtx, `SELECT pg_advisory_unlock($1::bigint)`, lockKey).Scan(&unlocked); err != nil || !unlocked {
				log.Printf("warning: no se pudo liberar bloqueo de documento FE para empresa_id=%d", empresaID)
			}
			_ = lockConn.Close()
		})
	}
	return context.WithValue(ctx, facturacionDocumentLockContextKey{}, lockKey), release, true, nil
}

func processFacturacionIntegracionForDocumentoContext(ctx context.Context, dbEmp *sql.DB, payload facturacionOperacionPayload, doc dbpkg.EmpresaDocumentoFacturacion, accion, usuario string, dbSuperOpt ...*sql.DB) (facturacionIntegracionResultado, *dbpkg.FacturacionElectronicaRetryItem, error) {
	resultado := facturacionIntegracionResultado{
		Aplica:             false,
		Accion:             strings.ToLower(strings.TrimSpace(accion)),
		EstadoEnvio:        "no_aplica",
		ContingenciaActiva: false,
		MaxIntentos:        5,
	}

	if dbEmp == nil {
		resultado.Error = "conexion de base de datos no disponible"
		return resultado, nil, fmt.Errorf("base de datos de empresa no disponible")
	}

	if doc.EmpresaID <= 0 {
		doc.EmpresaID = payload.EmpresaID
	}
	if doc.EmpresaID <= 0 {
		resultado.Error = "empresa_id es obligatorio"
		return resultado, nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if strings.TrimSpace(doc.DocumentoCodigo) == "" {
		doc.DocumentoCodigo = strings.TrimSpace(payload.DocumentoCodigo)
	}
	if strings.TrimSpace(doc.DocumentoCodigo) == "" {
		resultado.Error = "documento_codigo es obligatorio"
		return resultado, nil, fmt.Errorf("documento_codigo es obligatorio")
	}
	if strings.TrimSpace(doc.TipoDocumento) == "" {
		doc.TipoDocumento = strings.TrimSpace(payload.TipoDocumento)
	}
	if strings.EqualFold(strings.TrimSpace(doc.TipoDocumento), "nota_credito") {
		var ensureErr error
		if ensured, err := asegurarNumeroLegalNotaCredito(dbEmp, doc); err != nil {
			ensureErr = err
		} else if ensured != nil {
			doc = *ensured
		}
		if ensureErr != nil {
			resultado.Error = "no se pudo asegurar el consecutivo interno de la nota credito"
			return resultado, nil, ensureErr
		}
		hidratarReferenciaNotaCredito(dbEmp, doc, &payload)
	}
	if strings.TrimSpace(doc.TipoDocumento) == "" {
		doc.TipoDocumento = "factura_electronica"
	}
	if ctx == nil {
		ctx = context.Background()
	}
	documentLockKey := facturacionDocumentoAdvisoryLockKey(doc.EmpresaID, doc.TipoDocumento, doc.DocumentoCodigo)
	if documentLockKey == 0 {
		resultado.Error = "identidad de documento invalida para integracion fiscal"
		return resultado, nil, fmt.Errorf("clave de bloqueo fiscal invalida")
	}
	lockedContext, releaseDocumentLock, documentLocked, lockErr := acquireFacturacionDocumentAdvisoryLock(ctx, dbEmp, doc.EmpresaID, doc.TipoDocumento, doc.DocumentoCodigo)
	if lockErr != nil {
		resultado.Error = "no se pudo reservar el documento para integracion fiscal"
		return resultado, nil, lockErr
	}
	if !documentLocked {
		resultado.Error = "el documento ya tiene una integracion fiscal en proceso"
		return resultado, nil, fmt.Errorf("integracion fiscal concurrente para empresa_id=%d", doc.EmpresaID)
	}
	defer releaseDocumentLock()
	ctx = lockedContext
	if strings.TrimSpace(usuario) == "" {
		usuario = "sistema_facturacion"
	}
	var dbSuper *sql.DB
	if len(dbSuperOpt) > 0 {
		dbSuper = dbSuperOpt[0]
	}

	paisCodigo := strings.ToUpper(strings.TrimSpace(facturacionFirstNonBlank(doc.PaisCodigo, payload.PaisCodigo)))
	if paisCodigo == "" {
		paisDetectado, _, detectErr := dbpkg.DetectFacturacionPais(dbEmp, doc.EmpresaID, "", "")
		if detectErr == nil {
			paisCodigo = strings.ToUpper(strings.TrimSpace(paisDetectado.Codigo))
		}
	}
	if paisCodigo == "" {
		paisCodigo = "CO"
	}
	resultado.PaisCodigo = paisCodigo

	cfg, cfgErr := dbpkg.GetFacturacionElectronicaPaisConfig(dbEmp, doc.EmpresaID, paisCodigo)
	if cfgErr != nil && !errors.Is(cfgErr, sql.ErrNoRows) {
		resultado.Error = "no se pudo cargar configuracion FE"
		return resultado, nil, cfgErr
	}
	if cfg == nil {
		resultado.Error = "configuracion FE no disponible"
		return resultado, nil, nil
	}
	offlineSettings := facturacionDianOfflineSettingsFromConfig(cfg)
	offlineAplicaDIAN := paisCodigo == "CO"
	if offlineAplicaDIAN {
		resultado.OfflineDisponible = offlineSettings.Enabled
		resultado.OfflineConfirmado = false
		resultado.ConexionEstado = "online"
	}

	resultado.Proveedor = strings.TrimSpace(cfg.Proveedor)
	if resultado.Proveedor == "" {
		resultado.Proveedor = "manual"
	}
	resultado.Ambiente = strings.ToLower(strings.TrimSpace(cfg.Ambiente))
	if resultado.Ambiente != "produccion" {
		resultado.Ambiente = "sandbox"
	}

	retryActual, retryErr := dbpkg.GetFacturacionElectronicaRetryByDocumento(dbEmp, doc.EmpresaID, doc.TipoDocumento, doc.DocumentoCodigo)
	if retryErr != nil && !errors.Is(retryErr, sql.ErrNoRows) {
		resultado.Error = "no se pudo consultar cola de reintentos FE"
		return resultado, nil, retryErr
	}
	if retryActual != nil && (normalizeFacturacionEstadoEnvio(retryActual.EstadoEnvio) == "aceptado" || normalizeFacturacionEstadoEnvio(retryActual.EstadoEnvio) == "reconciliado") {
		docAceptado, changed := facturacionDocumentoAceptadoDIAN(doc, retryActual.RespuestaProveedor)
		if retryAceptado, retryChanged := facturacionRetryAceptadoConCodigoValidacion(retryActual, docAceptado.CodigoValidacion); retryChanged {
			retryActual, retryErr = dbpkg.UpsertFacturacionElectronicaRetryContext(ctx, dbEmp, retryAceptado)
			if retryErr != nil {
				resultado.Error = "el documento ya fue aceptado por DIAN, pero no se pudo completar su CUFE/CUDE en la cola fiscal"
				return resultado, nil, retryErr
			}
		}
		if changed {
			if _, updateErr := dbpkg.UpsertEmpresaDocumentoFacturacionContext(ctx, dbEmp, docAceptado); updateErr != nil {
				resultado.Error = "el documento ya fue aceptado por DIAN, pero no se pudo finalizar su estado local"
				return resultado, retryActual, updateErr
			}
		}
		resultado.Aplica = true
		resultado.EstadoEnvio = "aceptado"
		resultado.Intentos = retryActual.Intentos
		resultado.MaxIntentos = retryActual.MaxIntentos
		resultado.ReferenciaExterna = strings.TrimSpace(retryActual.ReferenciaExterna)
		resultado.ConexionEstado = "online"
		resultado.ConexionMensaje = "documento ya aceptado; no se reenvio"
		return resultado, retryActual, nil
	}
	if retryActual != nil && normalizeFacturacionEstadoEnvio(retryActual.EstadoEnvio) == "fallido_terminal" {
		if facturacionManualTerminalRetryAllowed(ctx) {
			reactivated, changed, reactivateErr := facturacionPrepareManualTerminalRetry(retryActual, doc, usuario)
			if reactivateErr != nil {
				resultado.Aplica = true
				resultado.EstadoEnvio = "fallido_terminal"
				resultado.Intentos = retryActual.Intentos
				resultado.MaxIntentos = retryActual.MaxIntentos
				resultado.Error = reactivateErr.Error()
				resultado.ConexionMensaje = "reenvio bloqueado hasta consultar la referencia fiscal existente"
				return resultado, retryActual, nil
			}
			if changed {
				retryActual, retryErr = dbpkg.UpsertFacturacionElectronicaRetryContext(ctx, dbEmp, *reactivated)
				if retryErr != nil {
					resultado.Error = "no se pudo reactivar la cola fiscal para reenvio manual"
					return resultado, nil, retryErr
				}
			}
		}
	}
	if retryActual != nil && normalizeFacturacionEstadoEnvio(retryActual.EstadoEnvio) == "fallido_terminal" {
		resultado.Aplica = true
		resultado.EstadoEnvio = "fallido_terminal"
		resultado.Intentos = retryActual.Intentos
		resultado.MaxIntentos = retryActual.MaxIntentos
		resultado.Error = strings.TrimSpace(retryActual.UltimoError)
		resultado.ConexionMensaje = "reintentos fiscales agotados; requiere revision manual"
		return resultado, retryActual, nil
	}

	retryPayload := dbpkg.FacturacionElectronicaRetryItem{
		EmpresaID:         doc.EmpresaID,
		TipoDocumento:     doc.TipoDocumento,
		DocumentoCodigo:   doc.DocumentoCodigo,
		PaisCodigo:        paisCodigo,
		Proveedor:         resultado.Proveedor,
		Ambiente:          resultado.Ambiente,
		MaxIntentos:       5,
		NumeroLegal:       strings.TrimSpace(doc.NumeroLegal),
		CodigoValidacion:  strings.TrimSpace(doc.CodigoValidacion),
		FechaEmisionLegal: strings.TrimSpace(doc.FechaDocumento),
		UsuarioCreador:    strings.TrimSpace(usuario),
		Estado:            "activo",
		Observaciones:     strings.TrimSpace(doc.Observaciones),
	}
	if retryActual != nil {
		retryPayload.ID = retryActual.ID
		retryPayload.Intentos = retryActual.Intentos
		retryPayload.MaxIntentos = retryActual.MaxIntentos
		retryPayload.ProximoIntento = strings.TrimSpace(retryActual.ProximoIntento)
		retryPayload.ReferenciaExterna = strings.TrimSpace(retryActual.ReferenciaExterna)
		retryPayload.FechaContingencia = strings.TrimSpace(retryActual.FechaContingencia)
		retryPayload.ContingenciaActiva = retryActual.ContingenciaActiva
		if strings.TrimSpace(retryActual.FechaEmisionLegal) != "" {
			retryPayload.FechaEmisionLegal = strings.TrimSpace(retryActual.FechaEmisionLegal)
		}
	}
	if retryPayload.MaxIntentos <= 0 {
		retryPayload.MaxIntentos = 5
	}
	resultado.Intentos = retryPayload.Intentos
	resultado.MaxIntentos = retryPayload.MaxIntentos

	aplicaIntegracion := resultado.Ambiente == "produccion" && strings.ToLower(strings.TrimSpace(cfg.Estado)) != "inactivo"
	if !aplicaIntegracion {
		retryPayload.EstadoEnvio = "no_aplica"
		retryPayload.Estado = "inactivo"
		retryPayload.ProximoIntento = ""
		retryPayload.FechaUltimoIntento = facturacionNowLocal()
		retryPayload.UltimoError = ""
		retryPayload.RespuestaProveedor = ""
		retryPayload.ContingenciaActiva = false
		retryPayload.FechaContingencia = ""
		retryPayload.Observaciones = strings.TrimSpace(facturacionFirstNonBlank(retryPayload.Observaciones, "integracion no aplica para ambiente/configuracion actual"))

		persistido, err := dbpkg.UpsertFacturacionElectronicaRetry(dbEmp, retryPayload)
		if err != nil {
			resultado.Error = "no se pudo actualizar cola FE no_aplica"
			return resultado, nil, err
		}
		resultado.EstadoEnvio = "no_aplica"
		resultado.Intentos = persistido.Intentos
		resultado.MaxIntentos = persistido.MaxIntentos
		resultado.ProximoIntento = strings.TrimSpace(persistido.ProximoIntento)
		resultado.ContingenciaActiva = persistido.ContingenciaActiva
		resultado.ReferenciaExterna = strings.TrimSpace(persistido.ReferenciaExterna)
		return resultado, persistido, nil
	}

	resultado.Aplica = true
	if paisCodigo == "CO" && (retryActual == nil || strings.TrimSpace(retryActual.FechaEmisionLegal) == "") {
		// Mantiene trazabilidad de la fecha fiscal efectiva sin reescribir la
		// fecha comercial almacenada en empresa_facturacion_documentos.
		retryPayload.FechaEmisionLegal = facturacionNowLocal()
	}
	var dispatch facturacionProveedorDispatchResult
	if paisCodigo == "CO" && normalizeFacturacionEstadoEnvio(retryPayload.EstadoEnvio) == "pendiente" && strings.TrimSpace(retryPayload.ReferenciaExterna) != "" {
		dispatch = dispatchFacturacionDIANAcusePendiente(dbEmp, doc, payload, strings.TrimSpace(retryPayload.ReferenciaExterna))
	} else {
		dispatch = dispatchFacturacionProveedor(dbEmp, cfg, payload, doc, accion, facturacionManualTerminalRetryAllowed(ctx))
	}
	now := facturacionNowLocal()
	retryPayload.Intentos = retryPayload.Intentos + 1
	retryPayload.FechaUltimoIntento = now
	retryPayload.RespuestaProveedor = strings.TrimSpace(dispatch.RespuestaJSON)
	retryPayload.UsuarioCreador = strings.TrimSpace(usuario)
	retryPayload.Estado = "activo"
	resultado.Intentos = retryPayload.Intentos
	resultado.MaxIntentos = retryPayload.MaxIntentos
	resultado.Advertencia = strings.TrimSpace(dispatch.ArtifactWarning)

	if dispatch.Success {
		estadoExito := "enviado"
		var dispatchMap map[string]interface{}
		if strings.TrimSpace(dispatch.RespuestaJSON) != "" && json.Unmarshal([]byte(dispatch.RespuestaJSON), &dispatchMap) == nil {
			estadoDIAN := strings.ToLower(strings.TrimSpace(genericStringValue(dispatchMap["estado_dian"])))
			acuseDIAN := strings.ToLower(strings.TrimSpace(genericStringValue(dispatchMap["acuse_estado"])))
			if estadoDIAN == "aceptado" || acuseDIAN == "aceptado" {
				estadoExito = "aceptado"
			}
			if estadoExito == "aceptado" {
				codigoValidacion := facturacionCUFEOficialDesdeMap(dispatchMap)
				if codigoValidacion != "" {
					doc.CodigoValidacion = codigoValidacion
					retryPayload.CodigoValidacion = codigoValidacion
				}
			}
		}
		if dispatch.Pending && estadoExito != "aceptado" {
			if retryPayload.Intentos >= retryPayload.MaxIntentos {
				estadoExito = "fallido_terminal"
				retryPayload.ProximoIntento = ""
				retryPayload.UltimoError = "acuse fiscal no concluyente despues de agotar los reintentos"
			} else {
				estadoExito = "pendiente"
				retryPayload.ProximoIntento = facturacionNextRetryAt(retryPayload.Intentos)
			}
		} else {
			retryPayload.ProximoIntento = ""
		}
		retryPayload.EstadoEnvio = estadoExito
		if estadoExito != "fallido_terminal" {
			retryPayload.UltimoError = ""
		}
		retryPayload.ContingenciaActiva = false
		retryPayload.FechaContingencia = ""
		retryPayload.ReferenciaExterna = strings.TrimSpace(dispatch.ReferenciaExterna)
		resultado.EstadoEnvio = estadoExito
		resultado.ReferenciaExterna = retryPayload.ReferenciaExterna
		resultado.Error = strings.TrimSpace(retryPayload.UltimoError)
		resultado.ConexionEstado = "online"
		if estadoExito == "fallido_terminal" {
			resultado.ConexionMensaje = "acuse no concluyente y reintentos agotados; requiere revision manual"
		} else {
			resultado.ConexionMensaje = "servidor DIAN/proveedor disponible"
		}
	} else {
		retryPayload.UltimoError = strings.TrimSpace(dispatch.Error)
		if retryPayload.UltimoError == "" {
			retryPayload.UltimoError = "fallo de integracion fiscal"
		}
		if dispatch.FinalFailure {
			retryPayload.EstadoEnvio = "fallido_terminal"
			retryPayload.Intentos = retryPayload.MaxIntentos
			retryPayload.ProximoIntento = ""
			retryPayload.ContingenciaActiva = false
			retryPayload.FechaContingencia = ""
			resultado.EstadoEnvio = "fallido_terminal"
			resultado.Intentos = retryPayload.Intentos
		} else if retryPayload.Intentos >= retryPayload.MaxIntentos {
			retryPayload.EstadoEnvio = "fallido_terminal"
			retryPayload.ContingenciaActiva = false
			retryPayload.FechaContingencia = ""
			retryPayload.ProximoIntento = ""
			resultado.EstadoEnvio = "fallido_terminal"
			resultado.ContingenciaActiva = false
		} else if offlineAplicaDIAN && dispatch.ConnectivityFailure {
			resultado.ConexionEstado = "offline"
			resultado.ConexionMensaje = facturacionConnectivityMessage(dispatch.Error)
			retryPayload.ProximoIntento = facturacionNextRetryAt(retryPayload.Intentos)
			contingenciaActiva, contingenciaErr := dbpkg.GetActiveEmpresaFacturacionContingencia(dbEmp, doc.EmpresaID, dbpkg.FacturacionContingenciaFallaDIAN)
			if contingenciaErr == nil && contingenciaActiva != nil {
				retryPayload.EstadoEnvio = "contingencia"
				retryPayload.ContingenciaActiva = true
				retryPayload.FechaContingencia = strings.TrimSpace(contingenciaActiva.FechaInicio)
				retryPayload.UltimoError = "Documento firmado y conservado durante contingencia DIAN declarada; transmision pendiente"
				resultado.EstadoEnvio = "contingencia"
				resultado.ContingenciaActiva = true
				resultado.AccionRecomendada = "reintentar_despues_recuperacion"
			} else {
				retryPayload.EstadoEnvio = "fallido"
				retryPayload.ContingenciaActiva = false
				retryPayload.FechaContingencia = ""
				retryPayload.UltimoError = "No hay conexion activa con DIAN/proveedor y no existe una contingencia fiscal declarada"
				resultado.EstadoEnvio = "fallido"
				resultado.AccionRecomendada = "declarar_incidente_o_revisar_conexion"
			}
			resultado.ProximoIntento = retryPayload.ProximoIntento
			resultado.RequiereConfirmacionOffline = false
		} else {
			retryPayload.EstadoEnvio = "fallido"
			retryPayload.ContingenciaActiva = false
			retryPayload.FechaContingencia = ""
			retryPayload.ProximoIntento = facturacionNextRetryAt(retryPayload.Intentos)
			resultado.EstadoEnvio = "fallido"
			resultado.ProximoIntento = retryPayload.ProximoIntento
		}
		resultado.Error = retryPayload.UltimoError
	}

	persistido, err := dbpkg.UpsertFacturacionElectronicaRetry(dbEmp, retryPayload)
	if err != nil {
		resultado.Error = "no se pudo persistir estado de integracion FE"
		return resultado, nil, err
	}
	if normalizeFacturacionEstadoEnvio(persistido.EstadoEnvio) == "contingencia" {
		if _, linkErr := dbpkg.RegisterEmpresaFacturacionContingenciaDocumento(dbEmp, doc.EmpresaID, dbpkg.FacturacionContingenciaFallaDIAN, doc.TipoDocumento, doc.DocumentoCodigo, persistido.ID); linkErr != nil {
			resultado.Error = "el documento quedo pendiente, pero no se pudo registrar en el incidente de contingencia"
			return resultado, persistido, linkErr
		}
	} else if normalizeFacturacionEstadoEnvio(persistido.EstadoEnvio) == "aceptado" || normalizeFacturacionEstadoEnvio(persistido.EstadoEnvio) == "reconciliado" {
		_ = dbpkg.SetEmpresaFacturacionContingenciaDocumentoEstado(dbEmp, doc.EmpresaID, doc.TipoDocumento, doc.DocumentoCodigo, "aceptado")
	}
	if estadoPersistido := normalizeFacturacionEstadoEnvio(persistido.EstadoEnvio); estadoPersistido == "fallido" || estadoPersistido == "fallido_terminal" {
		notificarFalloFacturacionElectronica(dbEmp, dbSuper, persistido, resultado, doc, usuario)
	} else if estadoPersistido == "aceptado" || estadoPersistido == "reconciliado" {
		docAceptado, changed := facturacionDocumentoAceptadoDIAN(doc, persistido.RespuestaProveedor)
		if changed {
			if _, updateErr := dbpkg.UpsertEmpresaDocumentoFacturacionContext(ctx, dbEmp, docAceptado); updateErr != nil {
				resultado.Error = "DIAN acepto el documento, pero no se pudo finalizar su estado local"
				return resultado, persistido, updateErr
			}
		}
	}

	resultado.EstadoEnvio = normalizeFacturacionEstadoEnvio(persistido.EstadoEnvio)
	resultado.Intentos = persistido.Intentos
	resultado.MaxIntentos = persistido.MaxIntentos
	resultado.ProximoIntento = strings.TrimSpace(persistido.ProximoIntento)
	resultado.ContingenciaActiva = persistido.ContingenciaActiva
	if strings.TrimSpace(resultado.ReferenciaExterna) == "" {
		resultado.ReferenciaExterna = strings.TrimSpace(persistido.ReferenciaExterna)
	}
	if strings.TrimSpace(resultado.Error) == "" {
		resultado.Error = strings.TrimSpace(persistido.UltimoError)
	}

	return resultado, persistido, nil
}

func asegurarNumeroLegalNotaCredito(dbEmp *sql.DB, nota dbpkg.EmpresaDocumentoFacturacion) (*dbpkg.EmpresaDocumentoFacturacion, error) {
	if dbEmp == nil || nota.EmpresaID <= 0 || nota.ID <= 0 {
		return nil, fmt.Errorf("nota credito persistida y empresa son obligatorias")
	}
	if strings.TrimSpace(nota.NumeroLegal) != "" {
		return &nota, nil
	}
	// Las notas usan un consecutivo interno independiente de la resolucion de
	// facturas. El prefijo NC seguido solo de digitos permite que CAJ50 compare
	// de forma inequivoca el prefijo con cbc:ID y conserva aislamiento empresarial.
	nota.NumeroLegal = fmt.Sprintf("NC%d%09d", nota.EmpresaID, nota.ID)
	return dbpkg.UpsertEmpresaDocumentoFacturacion(dbEmp, nota)
}

func hidratarReferenciaNotaCredito(dbEmp *sql.DB, nota dbpkg.EmpresaDocumentoFacturacion, payload *facturacionOperacionPayload) {
	if dbEmp == nil || payload == nil || nota.EmpresaID <= 0 {
		return
	}
	facturaCodigo := facturacionNotaCreditoFacturaOrigen(nota.Observaciones)
	if facturaCodigo == "" {
		return
	}
	factura, err := dbpkg.GetEmpresaDocumentoFacturacionByCodigo(dbEmp, nota.EmpresaID, "factura_electronica", facturaCodigo)
	if err != nil || factura == nil {
		return
	}
	hidratarCUFEOficialFactura(dbEmp, factura)
	if strings.TrimSpace(payload.ReferenciaDocumentoCodigo) == "" {
		payload.ReferenciaDocumentoCodigo = strings.TrimSpace(factura.NumeroLegal)
		if payload.ReferenciaDocumentoCodigo == "" {
			payload.ReferenciaDocumentoCodigo = strings.TrimSpace(factura.DocumentoCodigo)
		}
	}
	if strings.TrimSpace(payload.ReferenciaCUFE) == "" {
		payload.ReferenciaCUFE = strings.TrimSpace(factura.CodigoValidacion)
	}
	if strings.TrimSpace(payload.ReferenciaFechaEmision) == "" {
		payload.ReferenciaFechaEmision = strings.TrimSpace(factura.FechaDocumento)
	}
	if strings.TrimSpace(payload.CodigoCorreccion) == "" {
		payload.CodigoCorreccion = "2"
	}
	if strings.TrimSpace(payload.DescripcionCorreccion) == "" {
		payload.DescripcionCorreccion = "Anulación de factura electrónica"
	}
}

func notificarFalloFacturacionElectronica(dbEmp, dbSuper *sql.DB, retry *dbpkg.FacturacionElectronicaRetryItem, resultado facturacionIntegracionResultado, doc dbpkg.EmpresaDocumentoFacturacion, usuario string) {
	if dbEmp == nil || retry == nil || retry.EmpresaID <= 0 {
		return
	}
	ownerEmail := getEmpresaOwnerEmail(dbEmp, retry.EmpresaID)
	if ownerEmail == "" {
		ownerEmail = strings.ToLower(strings.TrimSpace(usuario))
	}
	if ownerEmail == "" || ownerEmail == "sistema" || ownerEmail == "sistema_facturacion" {
		return
	}
	actor, err := dbpkg.ResolveEmpresaBuzonActor(dbEmp, dbSuper, retry.EmpresaID, ownerEmail)
	if err != nil {
		actor = dbpkg.EmpresaBuzonActor{Tipo: "admin", Ref: ownerEmail, Email: ownerEmail, Nombre: ownerEmail, Rol: "administrador"}
	}
	errorText := strings.TrimSpace(firstNonEmptyString(resultado.Error, retry.UltimoError, "DIAN/proveedor rechazo o no confirmo el documento electronico"))
	errorVisible := dianUserVisibleError(errorText)
	causa := dianErrorUserHelp(errorText)
	mensaje := "La facturacion electronica requiere revision.\n\n" +
		"Documento: " + strings.TrimSpace(retry.TipoDocumento) + " " + strings.TrimSpace(retry.DocumentoCodigo) + "\n" +
		"Numero legal: " + strings.TrimSpace(firstNonEmptyString(retry.NumeroLegal, doc.NumeroLegal)) + "\n" +
		"Estado: " + strings.TrimSpace(retry.EstadoEnvio) + "\n" +
		"Error DIAN: " + errorVisible + "\n\n" +
		"Que hacer: " + causa + "\n\n" +
		"Abra Facturacion electronica > Pruebas DIAN para ver la consola, corregir configuracion y reenviar."
	_, _ = dbpkg.CreateEmpresaBuzonMensaje(dbEmp, dbpkg.EmpresaBuzonMensaje{
		EmpresaID:          retry.EmpresaID,
		DestinatarioTipo:   actor.Tipo,
		DestinatarioRef:    actor.Ref,
		DestinatarioEmail:  actor.Email,
		DestinatarioNombre: actor.Nombre,
		RemitenteTipo:      "sistema",
		RemitenteRef:       "facturacion_electronica",
		RemitenteNombre:    "Facturacion electronica PCS",
		Titulo:             "Error DIAN en facturacion electronica",
		Mensaje:            mensaje,
		Tipo:               "alerta_facturacion_electronica",
		Prioridad:          "alta",
		Modulo:             "facturacion_electronica",
		ReferenciaTipo:     strings.TrimSpace(retry.TipoDocumento),
		ReferenciaID:       retry.ID,
		EnlaceURL:          fmt.Sprintf("/administrar_empresa/facturacion_electronica_pruebas_dian.html?empresa_id=%d", retry.EmpresaID),
		UsuarioCreador:     usuario,
	})
}

func dianErrorUserHelp(raw string) string {
	lower := strings.ToLower(strings.TrimSpace(raw))
	switch {
	case strings.Contains(lower, "permission denied") || strings.Contains(lower, "acceso denegado") || strings.Contains(lower, "operation not permitted"):
		return "La clave privada de firma no puede ser leida por el servicio. Un administrador debe restaurar acceso solo para el usuario del backend o cargar nuevamente la firma desde PCS; no cambie ni comparta la clave en el buzon."
	case strings.Contains(lower, "fab05c") || (strings.Contains(lower, "identificador del software") && strings.Contains(lower, "rango")):
		return "Asocie en el portal DIAN el prefijo/rango de numeracion al software correcto y verifique Software ID, prefijo, resolucion y rango en PCS."
	case strings.Contains(lower, "fad06") || strings.Contains(lower, "cufe"):
		return "Consulte de nuevo la clave tecnica con GetNumberingRange y revise prefijo, numero legal, fecha/hora, impuestos y totales."
	case strings.Contains(lower, "fad05") || strings.Contains(lower, "rango de numeracion") || strings.Contains(lower, "resolucion"):
		return "Revise que la resolucion DIAN este vigente, asociada al software, con prefijo y consecutivo dentro del rango autorizado."
	case strings.Contains(lower, "fad10") || strings.Contains(lower, "softwaresecuritycode"):
		return "Revise Software ID, PIN tecnico y el numero completo de la factura usado para calcular el codigo de seguridad."
	case strings.Contains(lower, "fak61") || strings.Contains(lower, "party") || strings.Contains(lower, "cliente"):
		return "Corrija los datos del cliente: tipo de persona, tipo y numero de documento, municipio, direccion y regimen tributario."
	case strings.Contains(lower, "ze02") || strings.Contains(lower, "signature") || strings.Contains(lower, "firma"):
		return "Revise el certificado digital P12, su clave, vigencia y que corresponda al NIT emisor."
	case strings.Contains(lower, "vencid") || strings.Contains(lower, "expired"):
		return "Renueve certificado digital o resolucion vencida, cargue los nuevos datos en PCS y vuelva a probar DIAN."
	case strings.Contains(lower, "90") && strings.Contains(lower, "procesado anteriormente"):
		return "El documento ya fue procesado por DIAN. Consulte el CUFE/TrackId antes de reenviar para evitar duplicados."
	default:
		return "Lea el mensaje exacto de DIAN en la consola, valide configuracion, certificado, resolucion, rango, cliente y reintente el envio."
	}
}

// dianUserVisibleError keeps operational notifications useful without exposing
// filesystem paths, certificate material, tokens, or provider internals.
func dianUserVisibleError(raw string) string {
	clean := strings.TrimSpace(raw)
	lower := strings.ToLower(clean)
	if strings.Contains(lower, "permission denied") || strings.Contains(lower, "acceso denegado") || strings.Contains(lower, "operation not permitted") {
		return "No se pudo acceder a la clave privada de firma del certificado DIAN."
	}
	if dianErrorMayExposeInternalDetail(clean) {
		return "No se pudo preparar la firma o comunicacion DIAN. Revise la consola de Pruebas DIAN con un administrador autorizado."
	}
	return dianTruncate(clean, 240)
}

func dianErrorMayExposeInternalDetail(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	for _, marker := range []string{
		"sql:", "pq:", "postgres", "database", "dsn", "password", "token", "secret",
		"certificate", "x509", "dial tcp", "connection refused", "no such file",
		"permission denied", "stack trace", "traceback", "panic", "file:",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return strings.Contains(lower, "/app/") || strings.Contains(lower, "../") || strings.Contains(lower, "c:\\")
}

func facturacionBuildOperacionPayloadFromDocumento(doc dbpkg.EmpresaDocumentoFacturacion) facturacionOperacionPayload {
	return facturacionOperacionPayload{
		EmpresaID:       doc.EmpresaID,
		EntidadID:       doc.EntidadRelacionadaID,
		ClienteID:       doc.EntidadRelacionadaID,
		TipoDocumento:   strings.TrimSpace(doc.TipoDocumento),
		PaisCodigo:      strings.TrimSpace(doc.PaisCodigo),
		DocumentoCodigo: strings.TrimSpace(doc.DocumentoCodigo),
		EstadoActual:    strings.TrimSpace(doc.EstadoDocumento),
		MontoTotal:      doc.MontoTotal,
		Moneda:          strings.TrimSpace(doc.Moneda),
		PeriodoContable: strings.TrimSpace(doc.PeriodoContable),
		Observaciones:   strings.TrimSpace(doc.Observaciones),
	}
}

const facturacionNotaCreditoFacturaOrigenMarker = "FACTURA_ORIGEN=" // #nosec G101 -- prefijo documental publico, no es una credencial.

func facturacionIntegracionAceptada(resultado facturacionIntegracionResultado) bool {
	return strings.EqualFold(strings.TrimSpace(resultado.EstadoEnvio), "aceptado")
}

// facturacionDocumentoAceptadoDIAN completa la transicion local que quedo
// pendiente mientras no existia un acuse fiscal concluyente. Es deliberadamente
// idempotente: una factura ya emitida o una nota credito ya finalizada no cambia
// de estado ni pierde su ultimo evento.
func facturacionDocumentoAceptadoDIAN(doc dbpkg.EmpresaDocumentoFacturacion, respuestaProveedor string) (dbpkg.EmpresaDocumentoFacturacion, bool) {
	if !facturacionDocumentoElectronicoDIANFinalizacionLocalSoportada(doc.TipoDocumento) {
		return doc, false
	}

	changed := false
	if cufe := facturacionCUFEOficialDesdeRespuesta(respuestaProveedor); cufe != "" && !strings.EqualFold(strings.TrimSpace(doc.CodigoValidacion), cufe) {
		doc.CodigoValidacion = cufe
		changed = true
	}
	if !facturacionCodigoSHA384Valido(doc.CodigoValidacion) {
		return doc, changed
	}
	if strings.EqualFold(strings.TrimSpace(doc.EstadoDocumento), "pendiente_emision") {
		doc.EstadoAnterior = strings.TrimSpace(doc.EstadoDocumento)
		doc.EstadoDocumento = "emitida"
		doc.EventoUltimo = "integracion_fiscal_aceptada"
		changed = true
	}
	return doc, changed
}

func facturacionRetryAceptadoConCodigoValidacion(retry *dbpkg.FacturacionElectronicaRetryItem, codigoValidacion string) (dbpkg.FacturacionElectronicaRetryItem, bool) {
	if retry == nil {
		return dbpkg.FacturacionElectronicaRetryItem{}, false
	}
	estado := normalizeFacturacionEstadoEnvio(retry.EstadoEnvio)
	if (estado != "aceptado" && estado != "reconciliado") || !facturacionCodigoSHA384Valido(codigoValidacion) {
		return *retry, false
	}
	codigoValidacion = strings.ToLower(strings.TrimSpace(codigoValidacion))
	if strings.EqualFold(strings.TrimSpace(retry.CodigoValidacion), codigoValidacion) {
		return *retry, false
	}
	updated := *retry
	updated.CodigoValidacion = codigoValidacion
	return updated, true
}

func facturacionCodigoSHA384Valido(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) != 96 {
		return false
	}
	for _, char := range value {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') && (char < 'A' || char > 'F') {
			return false
		}
	}
	return true
}

func facturacionCUFEOficialDesdeRespuesta(raw string) string {
	var respuesta map[string]interface{}
	if json.Unmarshal([]byte(strings.TrimSpace(raw)), &respuesta) != nil {
		return ""
	}
	return facturacionCUFEOficialDesdeMap(respuesta)
}

func facturacionCUFEOficialDesdeMap(respuesta map[string]interface{}) string {
	if len(respuesta) == 0 {
		return ""
	}
	candidatos := []string{
		genericStringValue(respuesta["cufe"]),
		genericStringValue(respuesta["cude"]),
		genericStringValue(respuesta["cuds"]),
		genericStringValue(respuesta["cune"]),
		genericStringValue(respuesta["codigo_validacion"]),
		genericStringValue(respuesta["xml_document_key"]),
	}
	if respuestaDIAN, ok := respuesta["respuesta_dian"].(map[string]interface{}); ok {
		candidatos = append(candidatos,
			genericStringValue(respuestaDIAN["cufe"]),
			genericStringValue(respuestaDIAN["cude"]),
			genericStringValue(respuestaDIAN["cuds"]),
			genericStringValue(respuestaDIAN["cune"]),
			genericStringValue(respuestaDIAN["xml_document_key"]),
			genericStringValue(respuestaDIAN["document_key"]),
		)
	}
	for _, candidato := range candidatos {
		if facturacionCodigoSHA384Valido(candidato) {
			return strings.ToLower(strings.TrimSpace(candidato))
		}
	}
	return ""
}

func hidratarCUFEOficialFactura(dbEmp *sql.DB, factura *dbpkg.EmpresaDocumentoFacturacion) {
	if dbEmp == nil || factura == nil || factura.EmpresaID <= 0 || facturacionCodigoSHA384Valido(factura.CodigoValidacion) {
		return
	}
	retry, err := dbpkg.GetFacturacionElectronicaRetryByDocumento(dbEmp, factura.EmpresaID, "factura_electronica", factura.DocumentoCodigo)
	if err != nil || retry == nil || !strings.EqualFold(strings.TrimSpace(retry.EstadoEnvio), "aceptado") {
		return
	}
	cufe := facturacionCUFEOficialDesdeRespuesta(retry.RespuestaProveedor)
	if cufe == "" {
		return
	}
	factura.CodigoValidacion = cufe
	if actualizado, updateErr := dbpkg.UpsertEmpresaDocumentoFacturacion(dbEmp, *factura); updateErr == nil && actualizado != nil {
		*factura = *actualizado
	}
}

func facturacionNotaCreditoFacturaOrigen(observaciones string) string {
	for _, line := range strings.Split(observaciones, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, facturacionNotaCreditoFacturaOrigenMarker) {
			continue
		}
		codigo := strings.TrimSpace(strings.TrimPrefix(line, facturacionNotaCreditoFacturaOrigenMarker))
		if codigo != "" && len(codigo) <= 180 {
			return codigo
		}
	}
	return ""
}

func finalizarFacturaAnuladaPorNotaCredito(dbEmp *sql.DB, nota dbpkg.EmpresaDocumentoFacturacion, usuario string) (*dbpkg.EmpresaDocumentoFacturacion, error) {
	if dbEmp == nil || nota.EmpresaID <= 0 {
		return nil, fmt.Errorf("empresa y conexion son obligatorias")
	}
	refreshed, refreshErr := dbpkg.GetEmpresaDocumentoFacturacionByCodigo(dbEmp, nota.EmpresaID, nota.TipoDocumento, nota.DocumentoCodigo)
	if refreshErr != nil {
		return nil, fmt.Errorf("consultar nota credito aceptada: %w", refreshErr)
	}
	if refreshed != nil {
		nota = *refreshed
	}
	retry, retryErr := dbpkg.GetFacturacionElectronicaRetryByDocumento(dbEmp, nota.EmpresaID, nota.TipoDocumento, nota.DocumentoCodigo)
	if retryErr != nil {
		return nil, fmt.Errorf("consultar aceptacion de nota credito: %w", retryErr)
	}
	if !facturacionNotaCreditoAceptadaParaAnulacion(nota, retry) {
		return nil, fmt.Errorf("la nota credito no tiene estado, CUDE y acuse DIAN aceptados")
	}
	facturaCodigo := facturacionNotaCreditoFacturaOrigen(nota.Observaciones)
	if facturaCodigo == "" {
		return nil, nil
	}
	factura, err := dbpkg.GetEmpresaDocumentoFacturacionByCodigo(dbEmp, nota.EmpresaID, "factura_electronica", facturaCodigo)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(strings.TrimSpace(factura.EstadoDocumento), "anulada") {
		return factura, nil
	}
	if !strings.EqualFold(strings.TrimSpace(factura.EstadoDocumento), "emitida") {
		return nil, fmt.Errorf("la factura relacionada no esta emitida")
	}
	fuenteNota, fuenteErr := loadFacturacionFuenteFiscalParaDocumento(context.Background(), dbEmp, nota)
	if fuenteErr != nil {
		return nil, fmt.Errorf("la nota credito aceptada no conserva una fuente fiscal inmutable valida: %w", fuenteErr)
	}
	if !facturacionNotaCreditoFuenteValidaParaAnulacion(fuenteNota, nota, *factura) {
		return nil, fmt.Errorf("la fuente fiscal de la nota credito no coincide con la factura relacionada")
	}
	return dbpkg.UpsertEmpresaDocumentoFacturacion(dbEmp, dbpkg.EmpresaDocumentoFacturacion{
		EmpresaID:            factura.EmpresaID,
		TipoDocumento:        factura.TipoDocumento,
		DocumentoCodigo:      factura.DocumentoCodigo,
		EstadoDocumento:      "anulada",
		EstadoAnterior:       factura.EstadoDocumento,
		EventoUltimo:         "factura_anulada_por_nota_credito_aceptada",
		PeriodoContable:      factura.PeriodoContable,
		MontoTotal:           factura.MontoTotal,
		Moneda:               factura.Moneda,
		NumeroLegal:          factura.NumeroLegal,
		CodigoValidacion:     factura.CodigoValidacion,
		PaisCodigo:           factura.PaisCodigo,
		AmbienteFE:           factura.AmbienteFE,
		FechaDocumento:       factura.FechaDocumento,
		EntidadRelacionadaID: factura.EntidadRelacionadaID,
		UsuarioCreador:       strings.TrimSpace(usuario),
		Observaciones:        strings.TrimSpace(factura.Observaciones + "\nAnulada por nota credito DIAN aceptada " + nota.DocumentoCodigo + "."),
	})
}

func facturacionNotaCreditoAceptadaParaAnulacion(nota dbpkg.EmpresaDocumentoFacturacion, retry *dbpkg.FacturacionElectronicaRetryItem) bool {
	if !strings.EqualFold(strings.TrimSpace(nota.TipoDocumento), "nota_credito") ||
		!strings.EqualFold(strings.TrimSpace(nota.EstadoDocumento), "emitida") ||
		!facturacionCodigoSHA384Valido(nota.CodigoValidacion) || retry == nil {
		return false
	}
	estadoRetry := normalizeFacturacionEstadoEnvio(retry.EstadoEnvio)
	if estadoRetry != "aceptado" && estadoRetry != "reconciliado" {
		return false
	}
	if !facturacionCodigoSHA384Valido(retry.CodigoValidacion) {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(nota.CodigoValidacion), strings.TrimSpace(retry.CodigoValidacion))
}

func facturacionNotaCreditoFuenteValidaParaAnulacion(fuente *facturacionFuenteFiscalSnapshot, nota, factura dbpkg.EmpresaDocumentoFacturacion) bool {
	if fuente == nil || fuente.Referencia == nil ||
		!strings.EqualFold(strings.TrimSpace(fuente.Documento.TipoOrigen), "nota_credito") ||
		!strings.EqualFold(strings.TrimSpace(fuente.Documento.CodigoOrigen), strings.TrimSpace(nota.DocumentoCodigo)) ||
		!strings.EqualFold(strings.TrimSpace(fuente.Referencia.TipoDocumento), "factura_electronica") ||
		!strings.EqualFold(strings.TrimSpace(fuente.Referencia.DocumentoCodigo), strings.TrimSpace(factura.DocumentoCodigo)) ||
		!strings.EqualFold(strings.TrimSpace(fuente.Referencia.NumeroLegal), strings.TrimSpace(factura.NumeroLegal)) ||
		!facturacionCodigoSHA384Valido(fuente.Referencia.CodigoValidacion) {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(fuente.Referencia.CodigoValidacion), strings.TrimSpace(factura.CodigoValidacion))
}

func facturacionDeriveAccionByDocumento(doc dbpkg.EmpresaDocumentoFacturacion) string {
	tipo := strings.ToLower(strings.TrimSpace(doc.TipoDocumento))
	estado := strings.ToLower(strings.TrimSpace(doc.EstadoDocumento))
	switch normalizeFacturacionDocumentoElectronicoTipo(tipo) {
	case "nota_credito":
		return "nota_credito"
	case "nota_debito":
		return "nota_debito"
	case "documento_soporte":
		return "documento_soporte"
	case "nomina_electronica":
		return "nomina_electronica"
	case "documento_equivalente_pos":
		return "documento_equivalente_pos"
	default:
		if facturacionDocumentoElectronicoPermitido(tipo) {
			return normalizeFacturacionDocumentoElectronicoTipo(tipo)
		}
	}
	if estado == "anulada" {
		return "anular"
	}
	return "emitir"
}

const facturacionRetryAdvisoryNamespace int64 = 0x46455254 // FERT

func processFacturacionRetryQueue(dbEmp *sql.DB, empresaID int64, limit int, usuario string) (map[string]interface{}, error) {
	return processFacturacionRetryQueueContext(context.Background(), dbEmp, nil, empresaID, limit, usuario)
}

func processFacturacionRetryQueueContext(ctx context.Context, dbEmp, dbSuper *sql.DB, empresaID int64, limit int, usuario string) (map[string]interface{}, error) {
	return processFacturacionRetryQueueContextWithScope(ctx, dbEmp, dbSuper, empresaID, limit, usuario, true)
}

func facturacionRetryQueueDocumentAllowed(tipoDocumento string, includeNomina bool) bool {
	return includeNomina || !facturacionDocumentoEsFamiliaNomina(tipoDocumento)
}

func facturacionRetryQueueExcludedTypes(includeNomina bool) []string {
	if includeNomina {
		return nil
	}
	return []string{"nomina_electronica", "nota_ajuste_nomina_electronica"}
}

func processFacturacionRetryQueueContextWithScope(ctx context.Context, dbEmp, dbSuper *sql.DB, empresaID int64, limit int, usuario string, includeNomina bool) (map[string]interface{}, error) {
	if dbEmp == nil {
		return nil, fmt.Errorf("base de datos de empresa no disponible")
	}
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if strings.TrimSpace(usuario) == "" {
		usuario = "sistema_facturacion"
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	if ctx == nil {
		ctx = context.Background()
	}

	lockKey := (facturacionRetryAdvisoryNamespace << 32) | (empresaID & 0xffffffff)
	conn, err := dbEmp.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	locked := false
	if err := conn.QueryRowContext(ctx, `SELECT pg_try_advisory_lock($1::bigint)`, lockKey).Scan(&locked); err != nil {
		return nil, err
	}
	if !locked {
		return map[string]interface{}{
			"ok": true, "empresa_id": empresaID, "procesados": 0,
			"omitido": true, "motivo": "cola_en_proceso",
		}, nil
	}
	defer func() {
		releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var unlocked bool
		if err := conn.QueryRowContext(releaseCtx, `SELECT pg_advisory_unlock($1::bigint)`, lockKey).Scan(&unlocked); err != nil || !unlocked {
			log.Printf("warning: no se pudo liberar bloqueo de reintentos FE para empresa_id=%d", empresaID)
		}
	}()

	items, err := dbpkg.ClaimFacturacionElectronicaRetriesByEmpresaWithScope(dbEmp, empresaID, limit, usuario, includeNomina)
	if err != nil {
		return nil, err
	}

	resumenItems := make([]map[string]interface{}, 0, len(items))
	procesados := 0
	enviados := 0
	fallidos := 0
	contingencia := 0
	noAplica := 0
	erroresInternos := 0

	for _, retryItem := range items {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		detail := map[string]interface{}{
			"tipo_documento":   retryItem.TipoDocumento,
			"documento_codigo": retryItem.DocumentoCodigo,
			"estado_anterior":  retryItem.EstadoEnvio,
		}
		if !facturacionRetryQueueDocumentAllowed(retryItem.TipoDocumento, includeNomina) {
			continue
		}

		doc, docErr := dbpkg.GetEmpresaDocumentoFacturacionByCodigo(dbEmp, empresaID, retryItem.TipoDocumento, retryItem.DocumentoCodigo)
		if docErr != nil {
			if errors.Is(docErr, sql.ErrNoRows) {
				actualizado := retryItem
				if actualizado.MaxIntentos <= 0 {
					actualizado.MaxIntentos = 5
				}
				actualizado.Intentos = actualizado.Intentos + 1
				actualizado.FechaUltimoIntento = facturacionNowLocal()
				actualizado.UltimoError = "documento transaccional no encontrado para reintento"
				actualizado.UsuarioCreador = usuario
				actualizado.Estado = "activo"
				if actualizado.Intentos >= actualizado.MaxIntentos {
					actualizado.EstadoEnvio = "fallido"
					actualizado.ContingenciaActiva = false
					actualizado.FechaContingencia = ""
					actualizado.ProximoIntento = ""
				} else {
					actualizado.EstadoEnvio = "fallido"
					actualizado.ContingenciaActiva = false
					actualizado.FechaContingencia = ""
					actualizado.ProximoIntento = facturacionNextRetryAt(actualizado.Intentos)
				}
				persistido, upErr := dbpkg.UpsertFacturacionElectronicaRetry(dbEmp, actualizado)
				if upErr != nil {
					erroresInternos += 1
					detail["error"] = "no se pudo actualizar retry para documento inexistente"
				} else {
					detail["estado_nuevo"] = persistido.EstadoEnvio
					detail["intentos"] = persistido.Intentos
					detail["ultimo_error"] = persistido.UltimoError
					fallidos += 1
					procesados += 1
				}
				resumenItems = append(resumenItems, detail)
				if releaseErr := dbpkg.ReleaseFacturacionElectronicaRetryClaim(dbEmp, empresaID, retryItem.ID, retryItem.LeaseToken); releaseErr != nil {
					erroresInternos += 1
				}
				continue
			}
			erroresInternos += 1
			detail["error"] = "no se pudo consultar documento para reintento"
			resumenItems = append(resumenItems, detail)
			if releaseErr := dbpkg.ReleaseFacturacionElectronicaRetryClaim(dbEmp, empresaID, retryItem.ID, retryItem.LeaseToken); releaseErr != nil {
				erroresInternos += 1
			}
			continue
		}

		payload := facturacionBuildOperacionPayloadFromDocumento(*doc)
		accion := facturacionDeriveAccionByDocumento(*doc)
		resultado, persistido, procErr := processFacturacionIntegracionForDocumentoContext(ctx, dbEmp, payload, *doc, accion, usuario, dbSuper)
		if procErr != nil {
			erroresInternos += 1
			detail["error"] = procErr.Error()
		} else {
			detail["estado_nuevo"] = resultado.EstadoEnvio
			detail["intentos"] = resultado.Intentos
			detail["max_intentos"] = resultado.MaxIntentos
			detail["referencia_externa"] = resultado.ReferenciaExterna
			detail["proximo_intento"] = resultado.ProximoIntento
			detail["contingencia_activa"] = resultado.ContingenciaActiva
			detail["error_integracion"] = resultado.Error
			if persistido != nil {
				detail["cola_reintentos"] = persistido
			}
			if resultado.EstadoEnvio == "aceptado" && strings.EqualFold(doc.TipoDocumento, "factura_electronica") && facturacionAutoEmailClienteEnabled(dbEmp, empresaID, doc.PaisCodigo) {
				if refreshed, refreshErr := dbpkg.GetEmpresaDocumentoFacturacionByCodigo(dbEmp, empresaID, doc.TipoDocumento, doc.DocumentoCodigo); refreshErr == nil && refreshed != nil {
					doc = refreshed
				}
				detail["factura_email"] = enviarFacturaElectronicaAlCliente(dbEmp, dbSuper, payload, *doc)
			}

			if facturacionIntegracionAceptada(resultado) && strings.EqualFold(strings.TrimSpace(doc.TipoDocumento), "nota_credito") {
				if facturaAnulada, finalizeErr := finalizarFacturaAnuladaPorNotaCredito(dbEmp, *doc, usuario); finalizeErr != nil {
					erroresInternos += 1
					detail["error_finalizacion_anulacion"] = "no se pudo finalizar factura relacionada"
				} else if facturaAnulada != nil {
					detail["factura_anulada_codigo"] = facturaAnulada.DocumentoCodigo
				}
			}
			switch resultado.EstadoEnvio {
			case "aceptado", "enviado", "reconciliado":
				enviados += 1
			case "contingencia":
				contingencia += 1
			case "no_aplica":
				noAplica += 1
			default:
				fallidos += 1
			}
			procesados += 1
		}
		resumenItems = append(resumenItems, detail)
		if releaseErr := dbpkg.ReleaseFacturacionElectronicaRetryClaim(dbEmp, empresaID, retryItem.ID, retryItem.LeaseToken); releaseErr != nil {
			erroresInternos += 1
		}
	}

	return map[string]interface{}{
		"ok":               true,
		"empresa_id":       empresaID,
		"limit":            limit,
		"en_cola":          len(items),
		"procesados":       procesados,
		"enviados":         enviados,
		"fallidos":         fallidos,
		"contingencia":     contingencia,
		"no_aplica":        noAplica,
		"errores_internos": erroresInternos,
		"items":            resumenItems,
	}, nil
}

// RunFacturacionElectronicaRetriesScheduled procesa, desde pcs-worker, las
// empresas que tengan reintentos fiscales vencidos. El bloqueo por empresa
// evita duplicar transmisiones con una operacion manual concurrente.
func RunFacturacionElectronicaRetriesScheduled(ctx context.Context, dbEmp, dbSuper *sql.DB, tenantLimit, documentLimit int) error {
	return RunFacturacionElectronicaRetriesScheduledShard(ctx, dbEmp, dbSuper, tenantLimit, documentLimit, 0, 1)
}

// RunFacturacionElectronicaRetriesScheduledShard processes one disjoint tenant
// shard. A failure in one company is recorded but does not block the remaining
// companies selected for the same cycle.
func RunFacturacionElectronicaRetriesScheduledShard(ctx context.Context, dbEmp, dbSuper *sql.DB, tenantLimit, documentLimit, shardIndex, shardCount int) error {
	if dbEmp == nil {
		return fmt.Errorf("base de datos de empresa no disponible")
	}
	empresaIDs, err := dbpkg.ListFacturacionElectronicaRetryEmpresaIDsDueShardContext(ctx, dbEmp, tenantLimit, shardIndex, shardCount)
	if err != nil {
		return err
	}
	var firstErr error
	failures := 0
	for _, empresaID := range empresaIDs {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		_, processErr := processFacturacionRetryQueueContext(ctx, dbEmp, dbSuper, empresaID, documentLimit, "sistema.pcs-worker")
		if markErr := dbpkg.MarkQueueTenantServedContext(ctx, dbEmp, dbpkg.QueueLaneFiscal, empresaID); markErr != nil {
			log.Printf("[facturacion_queue] no se pudo rotar empresa_id=%d shard=%d/%d error_type=%T", empresaID, shardIndex, shardCount, markErr)
		}
		if processErr != nil {
			failures++
			if firstErr == nil {
				firstErr = fmt.Errorf("empresa_id %d: %w", empresaID, processErr)
			}
			log.Printf("[facturacion_queue] empresa fallida empresa_id=%d shard=%d/%d error_type=%T", empresaID, shardIndex, shardCount, processErr)
		}
	}
	if failures > 0 {
		return fmt.Errorf("cola fiscal shard %d/%d termino con %d empresa(s) fallidas; primera: %w", shardIndex, shardCount, failures, firstErr)
	}
	return nil
}

func buildFacturacionReconciliacion(dbEmp *sql.DB, empresaID int64) (map[string]interface{}, error) {
	return reconcileFacturacionEstados(dbEmp, empresaID, false, false, "")
}

func listFacturacionDocumentosForReconciliacion(dbEmp *sql.DB, empresaID int64) ([]dbpkg.EmpresaDocumentoFacturacionListado, error) {
	return dbpkg.ListEmpresaDocumentosFacturacionByEmpresa(dbEmp, dbpkg.EmpresaDocumentoFacturacionListFilter{
		EmpresaID:       empresaID,
		IncludeInactive: false,
		Limit:           1000,
		Offset:          0,
	})
}

type facturacionReconciliacionAceptadaResultado struct {
	Atendida        bool
	Conciliada      bool
	Pendiente       bool
	Procesada       bool
	ReparacionLocal bool
	ErroresInternos int
	Inconsistencia  map[string]interface{}
}

func reconciliarFacturacionAceptacionLocal(dbEmp *sql.DB, doc dbpkg.EmpresaDocumentoFacturacionListado, retryItem *dbpkg.FacturacionElectronicaRetryItem, aplicar bool, usuario string) facturacionReconciliacionAceptadaResultado {
	resultado := facturacionReconciliacionAceptadaResultado{}
	if retryItem == nil {
		return resultado
	}
	estadoRetry := normalizeFacturacionEstadoEnvio(retryItem.EstadoEnvio)
	if estadoRetry != "aceptado" && estadoRetry != "reconciliado" {
		return resultado
	}
	resultado.Atendida = true

	docAceptado, docChanged := facturacionDocumentoAceptadoDIAN(doc.EmpresaDocumentoFacturacion, retryItem.RespuestaProveedor)
	retryAceptado, retryChanged := facturacionRetryAceptadoConCodigoValidacion(retryItem, docAceptado.CodigoValidacion)
	if !facturacionCodigoSHA384Valido(docAceptado.CodigoValidacion) {
		resultado.Pendiente = true
		resultado.Inconsistencia = map[string]interface{}{
			"tipo_documento":   doc.TipoDocumento,
			"documento_codigo": doc.DocumentoCodigo,
			"problema":         "acuse_aceptado_sin_cufe_cude_oficial",
		}
		return resultado
	}
	if docChanged || retryChanged {
		resultado.Inconsistencia = map[string]interface{}{
			"tipo_documento":      doc.TipoDocumento,
			"documento_codigo":    doc.DocumentoCodigo,
			"problema":            "aceptacion_local_pendiente",
			"reparar_documento":   docChanged,
			"reparar_cola_fiscal": retryChanged,
		}
	}
	if aplicar {
		if retryChanged {
			persistido, updateErr := dbpkg.UpsertFacturacionElectronicaRetry(dbEmp, retryAceptado)
			if updateErr != nil {
				resultado.ErroresInternos++
				return resultado
			}
			retryItem = persistido
		}
		if docChanged {
			persistido, updateErr := dbpkg.UpsertEmpresaDocumentoFacturacion(dbEmp, docAceptado)
			if updateErr != nil {
				resultado.ErroresInternos++
				return resultado
			}
			if persistido != nil {
				doc.EmpresaDocumentoFacturacion = *persistido
			}
		}
		if docChanged || retryChanged {
			resultado.ReparacionLocal = true
			resultado.Procesada = true
		}
		if strings.EqualFold(strings.TrimSpace(doc.TipoDocumento), "nota_credito") {
			if _, finalizeErr := finalizarFacturaAnuladaPorNotaCredito(dbEmp, doc.EmpresaDocumentoFacturacion, usuario); finalizeErr != nil {
				resultado.ErroresInternos++
			}
		}
	}
	resultado.Conciliada = true
	return resultado
}

func reconcileFacturacionEstados(dbEmp *sql.DB, empresaID int64, aplicar, soloAcusesAceptados bool, usuario string) (map[string]interface{}, error) {
	if dbEmp == nil {
		return nil, fmt.Errorf("base de datos de empresa no disponible")
	}
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if strings.TrimSpace(usuario) == "" {
		usuario = "sistema_facturacion"
	}

	documentos, err := listFacturacionDocumentosForReconciliacion(dbEmp, empresaID)
	if err != nil {
		return nil, err
	}

	inconsistencias := make([]map[string]interface{}, 0)
	documentosEvaluados := 0
	conciliados := 0
	pendientes := 0
	noAplica := 0
	procesados := 0
	enviados := 0
	fallidos := 0
	contingencia := 0
	erroresInternos := 0
	reparacionesLocales := 0
	omitidosNoAceptados := 0

	for _, doc := range documentos {
		tipo := strings.ToLower(strings.TrimSpace(doc.TipoDocumento))
		estadoDocumento := strings.ToLower(strings.TrimSpace(doc.EstadoDocumento))
		if !facturacionDocumentoElectronicoDIANComercialSoportado(tipo) {
			continue
		}
		if estadoDocumento != "emitida" && estadoDocumento != "anulada" && estadoDocumento != "pendiente_emision" {
			continue
		}

		documentosEvaluados += 1
		retryItem, retryErr := dbpkg.GetFacturacionElectronicaRetryByDocumento(dbEmp, empresaID, doc.TipoDocumento, doc.DocumentoCodigo)
		if retryErr != nil && !errors.Is(retryErr, sql.ErrNoRows) {
			erroresInternos += 1
			inconsistencias = append(inconsistencias, map[string]interface{}{
				"tipo_documento":   doc.TipoDocumento,
				"documento_codigo": doc.DocumentoCodigo,
				"problema":         "error_consulta_retry",
				"detalle":          retryErr.Error(),
			})
			continue
		}

		estadoRetry := "sin_cola"
		if retryItem != nil {
			estadoRetry = normalizeFacturacionEstadoEnvio(retryItem.EstadoEnvio)
		}

		aceptada := reconciliarFacturacionAceptacionLocal(dbEmp, doc, retryItem, aplicar, usuario)
		if aceptada.Atendida {
			if aceptada.Inconsistencia != nil {
				inconsistencias = append(inconsistencias, aceptada.Inconsistencia)
			}
			if aceptada.Pendiente {
				pendientes++
			}
			if aceptada.Conciliada {
				conciliados++
			}
			if aceptada.Procesada {
				procesados++
			}
			if aceptada.ReparacionLocal {
				reparacionesLocales++
			}
			erroresInternos += aceptada.ErroresInternos
			continue
		}
		// El modo operativo de cierre local se limita a colas que ya contienen un
		// acuse aceptado/reconciliado. Debe terminar antes de cualquier camino que
		// consulte o despache nuevamente el XML de documentos pendientes.
		if soloAcusesAceptados {
			omitidosNoAceptados++
			continue
		}
		// "enviado" ya tiene una transmisión registrada, pero todavía no un
		// acuse concluyente. La reconciliación no debe volver a despachar ese XML;
		// su seguimiento corresponde al TrackId/acuse o a la cola dedicada.
		if estadoRetry == "enviado" {
			conciliados++
			continue
		}
		if estadoDocumento == "pendiente_emision" {
			pendientes += 1
			inconsistencias = append(inconsistencias, map[string]interface{}{
				"tipo_documento":   doc.TipoDocumento,
				"documento_codigo": doc.DocumentoCodigo,
				"estado_documento": doc.EstadoDocumento,
				"estado_retry":     estadoRetry,
				"problema":         "documento_pendiente_sin_aceptacion",
			})
			continue
		}
		if estadoRetry == "no_aplica" {
			noAplica += 1
			continue
		}

		if strings.ToLower(strings.TrimSpace(doc.AmbienteFE)) == "sandbox" && retryItem == nil {
			noAplica += 1
			continue
		}

		pendientes += 1
		item := map[string]interface{}{
			"tipo_documento":   doc.TipoDocumento,
			"documento_codigo": doc.DocumentoCodigo,
			"estado_documento": doc.EstadoDocumento,
			"estado_retry":     estadoRetry,
			"pais_codigo":      doc.PaisCodigo,
			"ambiente_fe":      doc.AmbienteFE,
		}

		if aplicar {
			payload := facturacionBuildOperacionPayloadFromDocumento(doc.EmpresaDocumentoFacturacion)
			accion := facturacionDeriveAccionByDocumento(doc.EmpresaDocumentoFacturacion)
			resultado, persistido, procErr := processFacturacionIntegracionForDocumento(dbEmp, payload, doc.EmpresaDocumentoFacturacion, accion, usuario)
			if procErr != nil {
				erroresInternos += 1
				item["error"] = procErr.Error()
			} else {
				item["estado_reconciliado"] = resultado.EstadoEnvio
				item["intentos"] = resultado.Intentos
				item["max_intentos"] = resultado.MaxIntentos
				item["proximo_intento"] = resultado.ProximoIntento
				item["contingencia_activa"] = resultado.ContingenciaActiva
				item["referencia_externa"] = resultado.ReferenciaExterna
				item["error_integracion"] = resultado.Error
				if persistido != nil {
					item["cola_reintentos"] = persistido
				}

				if facturacionIntegracionAceptada(resultado) && strings.EqualFold(strings.TrimSpace(doc.TipoDocumento), "nota_credito") {
					if facturaAnulada, finalizeErr := finalizarFacturaAnuladaPorNotaCredito(dbEmp, doc.EmpresaDocumentoFacturacion, usuario); finalizeErr != nil {
						erroresInternos += 1
						item["error_finalizacion_anulacion"] = "no se pudo finalizar factura relacionada"
					} else if facturaAnulada != nil {
						item["factura_anulada_codigo"] = facturaAnulada.DocumentoCodigo
					}
				}
				switch resultado.EstadoEnvio {
				case "aceptado", "enviado", "reconciliado":
					enviados += 1
				case "contingencia":
					contingencia += 1
				case "no_aplica":
					noAplica += 1
				default:
					fallidos += 1
				}
				procesados += 1
			}
		}

		inconsistencias = append(inconsistencias, item)
	}

	return map[string]interface{}{
		"ok":                         true,
		"empresa_id":                 empresaID,
		"aplicar":                    aplicar,
		"solo_acuses_aceptados":      soloAcusesAceptados,
		"transmision_xml_habilitada": aplicar && !soloAcusesAceptados,
		"timestamp":                  facturacionNowLocal(),
		"documentos_evaluados":       documentosEvaluados,
		"documentos_conciliados":     conciliados,
		"pendientes_reconciliacion":  pendientes,
		"documentos_no_aplica":       noAplica,
		"procesados":                 procesados,
		"reparaciones_locales":       reparacionesLocales,
		"omitidos_no_aceptados":      omitidosNoAceptados,
		"enviados":                   enviados,
		"fallidos":                   fallidos,
		"contingencia":               contingencia,
		"errores_internos":           erroresInternos,
		"inconsistencias":            inconsistencias,
	}, nil
}
