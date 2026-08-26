package db

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// EmpresaDocumentoSoporteConfiguracionSnapshot is the immutable, secret-free
// numbering configuration associated with one reserved legal number.
type EmpresaDocumentoSoporteConfiguracionSnapshot struct {
	TipoDocumento        string `json:"tipo_documento"`
	TipoAmbiente         string `json:"tipo_ambiente"`
	ModoOperacionCodigo  string `json:"modo_operacion_codigo,omitempty"`
	Prefijo              string `json:"prefijo"`
	ResolucionNumero     string `json:"resolucion_numero,omitempty"`
	ResolucionFechaDesde string `json:"resolucion_fecha_desde,omitempty"`
	ResolucionFechaHasta string `json:"resolucion_fecha_hasta,omitempty"`
	RangoDesde           int64  `json:"rango_desde"`
	RangoHasta           int64  `json:"rango_hasta"`
	ConsecutivoAsignado  int64  `json:"consecutivo_asignado"`
	URLDIANOverride      string `json:"url_dian_override,omitempty"`
}

// ReserveEmpresaDocumentoSoporteNumeroContext reserves exactly one number for
// a tenant-owned draft. The document row and its dedicated DIAN configuration
// are locked in the same transaction, making concurrent retries idempotent.
func ReserveEmpresaDocumentoSoporteNumeroContext(ctx context.Context, dbConn *sql.DB, empresaID, documentoSoporteID int64, emissionTime time.Time) (*EmpresaDocumentoSoporteElectronico, *EmpresaDocumentoSoporteConfiguracionSnapshot, error) {
	if dbConn == nil {
		return nil, nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 || documentoSoporteID <= 0 {
		return nil, nil, errors.New("empresa_id y documento_soporte_id son obligatorios")
	}
	if emissionTime.IsZero() {
		emissionTime = time.Now()
	}
	emissionTime = emissionTime.In(facturacionColombiaLocation())

	tx, err := dbConn.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	document, err := lockEmpresaDocumentoSoporteDraftTx(tx, empresaID, documentoSoporteID)
	if err != nil {
		return nil, nil, err
	}

	if strings.TrimSpace(document.NumeroLegal) != "" {
		var snapshot EmpresaDocumentoSoporteConfiguracionSnapshot
		if err := json.Unmarshal([]byte(document.ConfiguracionDIANJSON), &snapshot); err != nil || snapshot.TipoDocumento != "documento_soporte" || snapshot.ConsecutivoAsignado <= 0 {
			return nil, nil, errors.New("documento soporte reservado sin instantanea DIAN valida; requiere conciliacion manual")
		}
		if err := tx.Commit(); err != nil {
			return nil, nil, err
		}
		return document, &snapshot, nil
	}
	if strings.TrimSpace(document.LineasJSON) == "" || strings.TrimSpace(document.LineasJSON) == "[]" || document.Total <= 0 {
		return nil, nil, errors.New("documento soporte no tiene fuente estructurada completa")
	}

	config, err := lockEmpresaDocumentoSoporteConfigTx(tx, empresaID, emissionTime)
	if err != nil {
		return nil, nil, err
	}

	assigned := config.ConsecutivoActual
	number := strings.ToUpper(strings.TrimSpace(config.Prefijo)) + fmt.Sprintf("%d", assigned)
	snapshot := &EmpresaDocumentoSoporteConfiguracionSnapshot{
		TipoDocumento:        "documento_soporte",
		TipoAmbiente:         strings.ToLower(strings.TrimSpace(config.TipoAmbiente)),
		ModoOperacionCodigo:  strings.TrimSpace(config.ModoOperacionCodigo),
		Prefijo:              strings.ToUpper(strings.TrimSpace(config.Prefijo)),
		ResolucionNumero:     strings.TrimSpace(config.ResolucionNumero),
		ResolucionFechaDesde: strings.TrimSpace(config.ResolucionFechaDesde),
		ResolucionFechaHasta: strings.TrimSpace(config.ResolucionFechaHasta),
		RangoDesde:           config.RangoDesde,
		RangoHasta:           config.RangoHasta,
		ConsecutivoAsignado:  assigned,
		URLDIANOverride:      strings.TrimSpace(config.URLDIANOverride),
	}
	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return nil, nil, fmt.Errorf("serializar configuracion DIAN de documento soporte: %w", err)
	}
	if _, err := execTxSQLCompat(tx, `UPDATE empresa_dian_documentos_configuracion
		SET consecutivo_actual = ?, fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE empresa_id = ? AND tipo_documento = 'documento_soporte'`, assigned+1, empresaID); err != nil {
		return nil, nil, err
	}
	if _, err := execTxSQLCompat(tx, `UPDATE empresa_contabilidad_documentos_soporte
		SET numero_legal = ?, fecha_emision_legal = ?, configuracion_dian_json = ?::jsonb,
			estado_dian = 'preparado', fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE empresa_id = ? AND id = ? AND numero_legal = ''`,
		number, emissionTime.Format(time.RFC3339), string(snapshotJSON), empresaID, documentoSoporteID); err != nil {
		return nil, nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}
	document.NumeroLegal = number
	document.FechaEmisionLegal = emissionTime.Format(time.RFC3339)
	document.ConfiguracionDIANJSON = string(snapshotJSON)
	document.EstadoDIAN = "preparado"
	return document, snapshot, nil
}

func lockEmpresaDocumentoSoporteDraftTx(tx *sql.Tx, empresaID, documentoSoporteID int64) (*EmpresaDocumentoSoporteElectronico, error) {
	var document EmpresaDocumentoSoporteElectronico
	err := scanEmpresaDocumentoSoporte(queryRowTxSQLCompat(tx, `SELECT `+empresaDocumentoSoporteSelectColumns+`
		FROM empresa_contabilidad_documentos_soporte
		WHERE empresa_id = ? AND id = ?
		FOR UPDATE`, empresaID, documentoSoporteID), &document)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("documento soporte no encontrado para la empresa")
	}
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func lockEmpresaDocumentoSoporteConfigTx(tx *sql.Tx, empresaID int64, emissionTime time.Time) (*EmpresaDIANDocumentoConfiguracion, error) {
	var config EmpresaDIANDocumentoConfiguracion
	err := queryRowTxSQLCompat(tx, `SELECT
		id, empresa_id, tipo_documento, estado, tipo_ambiente,
		COALESCE(modo_operacion_codigo, ''), COALESCE(test_set_id, ''),
		COALESCE(prefijo, ''), COALESCE(resolucion_numero, ''),
		COALESCE(resolucion_fecha_desde::TEXT, ''), COALESCE(resolucion_fecha_hasta::TEXT, ''),
		COALESCE(rango_desde, 0), COALESCE(rango_hasta, 0), COALESCE(consecutivo_actual, 0),
		COALESCE(url_dian_override, ''), COALESCE(observaciones, ''),
		COALESCE(usuario_creador, ''), COALESCE(fecha_creacion::TEXT, ''), COALESCE(fecha_actualizacion::TEXT, '')
	FROM empresa_dian_documentos_configuracion
	WHERE empresa_id = ? AND tipo_documento = 'documento_soporte'
	FOR UPDATE`, empresaID).Scan(
		&config.ID, &config.EmpresaID, &config.TipoDocumento, &config.Estado, &config.TipoAmbiente,
		&config.ModoOperacionCodigo, &config.TestSetID, &config.Prefijo, &config.ResolucionNumero,
		&config.ResolucionFechaDesde, &config.ResolucionFechaHasta, &config.RangoDesde,
		&config.RangoHasta, &config.ConsecutivoActual, &config.URLDIANOverride,
		&config.Observaciones, &config.UsuarioCreador, &config.FechaCreacion, &config.FechaActualizacion)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("no existe configuracion DIAN separada para documento soporte")
	}
	if err != nil {
		return nil, err
	}
	if err := ValidateEmpresaDocumentoSoporteConfigForEmission(config, emissionTime); err != nil {
		return nil, err
	}
	return &config, nil
}

// ValidateEmpresaDocumentoSoporteConfigForEmission is read-only and can be
// used by HTTP preflight before a legal number is consumed.
func ValidateEmpresaDocumentoSoporteConfigForEmission(config EmpresaDIANDocumentoConfiguracion, emissionTime time.Time) error {
	environment := strings.ToLower(strings.TrimSpace(config.TipoAmbiente))
	state := strings.ToLower(strings.TrimSpace(config.Estado))
	switch environment {
	case "habilitacion":
		if state != "habilitacion" {
			return errors.New("la configuracion de documento soporte debe estar en estado habilitacion")
		}
	case "produccion":
		if state != "activo" {
			return errors.New("la configuracion de documento soporte debe estar activa para produccion")
		}
	default:
		return errors.New("ambiente DIAN de documento soporte invalido")
	}
	resolution := strings.TrimSpace(config.ResolucionNumero)
	if len(resolution) != 14 || !documentoSoporteASCIIDigits(resolution) {
		return errors.New("la autorizacion DIAN de documento soporte debe tener exactamente 14 digitos")
	}
	prefix := strings.ToUpper(strings.TrimSpace(config.Prefijo))
	if len(prefix) > 4 || !documentoSoportePrefixValid(prefix) {
		return errors.New("el prefijo DIAN de documento soporte debe ser alfanumerico y tener maximo 4 caracteres; puede quedar vacio")
	}
	if config.RangoDesde <= 0 || config.RangoHasta < config.RangoDesde || config.RangoHasta > 999999999 {
		return errors.New("rango DIAN de documento soporte invalido; debe estar entre 1 y 999999999")
	}
	if config.ConsecutivoActual < config.RangoDesde || config.ConsecutivoActual > config.RangoHasta {
		return errors.New("consecutivo DIAN de documento soporte fuera del rango autorizado")
	}
	start := strings.TrimSpace(config.ResolucionFechaDesde)
	end := strings.TrimSpace(config.ResolucionFechaHasta)
	if start == "" || end == "" {
		return errors.New("la vigencia completa de la autorizacion DIAN de documento soporte es obligatoria")
	}
	if _, err := time.Parse("2006-01-02", start); err != nil {
		return errors.New("fecha inicial de autorizacion DIAN de documento soporte invalida")
	}
	if _, err := time.Parse("2006-01-02", end); err != nil || end < start {
		return errors.New("fecha final de autorizacion DIAN de documento soporte invalida")
	}
	today := emissionTime.In(facturacionColombiaLocation()).Format("2006-01-02")
	if today < start {
		return errors.New("la numeracion DIAN de documento soporte aun no inicia vigencia")
	}
	if today > end {
		return errors.New("la numeracion DIAN de documento soporte esta vencida")
	}
	return nil
}

func documentoSoporteASCIIDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func documentoSoportePrefixValid(value string) bool {
	for _, char := range value {
		if (char < 'A' || char > 'Z') && (char < '0' || char > '9') {
			return false
		}
	}
	return true
}

// UpdateEmpresaDocumentoSoporteDIANResultContext mirrors only the provider
// result into the accounting record. CUDS is accepted exclusively from the
// signed XML or DIAN acknowledgment processed by the dedicated adapter.
func UpdateEmpresaDocumentoSoporteDIANResultContext(ctx context.Context, dbConn *sql.DB, empresaID, documentoSoporteID int64, estado, cuds, respuesta string, fuenteSellada bool) error {
	if dbConn == nil || empresaID <= 0 || documentoSoporteID <= 0 {
		return errors.New("empresa_id y documento_soporte_id son obligatorios")
	}
	estado = strings.ToLower(strings.TrimSpace(estado))
	switch estado {
	case "preparado", "pendiente", "enviado", "aceptado", "rechazado", "fallido", "contingencia":
	default:
		return errors.New("estado DIAN de documento soporte invalido")
	}
	cuds = strings.ToLower(strings.TrimSpace(cuds))
	if cuds != "" && !empresaDocumentoSoporteCUDSValido(cuds) {
		return errors.New("CUDS de documento soporte invalido")
	}
	result, err := execSQLCompatContext(ctx, dbConn, `UPDATE empresa_contabilidad_documentos_soporte
		SET estado_dian = ?, cuds = COALESCE(NULLIF(?, ''), cuds),
			respuesta_dian = COALESCE(NULLIF(?, ''), respuesta_dian), fuente_fiscal_sellada = ?,
			fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE empresa_id = ? AND id = ?`, estado, strings.TrimSpace(cuds), strings.TrimSpace(respuesta), fuenteSellada, empresaID, documentoSoporteID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return errors.New("documento soporte no encontrado para la empresa")
	}
	return nil
}

func empresaDocumentoSoporteCUDSValido(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) != 96 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
