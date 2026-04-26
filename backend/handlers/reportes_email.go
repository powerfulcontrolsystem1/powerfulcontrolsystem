package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"net/http"
	"database/sql"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func reportesBuildExportBytes(ds empresaReporteDataset, format string) (string, string, []byte, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "excel" || format == "tsv" {
		format = "xls"
	}
	switch format {
	case "json":
		b, err := jsonMarshalDataset(ds)
		if err != nil {
			return "", "", nil, err
		}
		return reportesBuildFileName(ds.Key, ds.EmpresaID, "json"), "application/json", b, nil
	case "csv":
		content, err := reportesDatasetCSVContent(ds)
		if err != nil {
			return "", "", nil, fmt.Errorf("no se pudo generar CSV")
		}
		return reportesBuildFileName(ds.Key, ds.EmpresaID, "csv"), "text/csv; charset=utf-8", []byte(content), nil
	case "txt":
		content := reportesDatasetTXTContent(ds)
		return reportesBuildFileName(ds.Key, ds.EmpresaID, "txt"), "text/plain; charset=utf-8", []byte(content), nil
	case "xls":
		content := "\ufeff" + reportesDatasetTSVContent(ds)
		return reportesBuildFileName(ds.Key, ds.EmpresaID, "xls"), "application/vnd.ms-excel; charset=utf-8", []byte(content), nil
	case "pdf":
		content := reportesDatasetPDFContent(ds)
		return reportesBuildFileName(ds.Key, ds.EmpresaID, "pdf"), "application/pdf", content, nil
	default:
		return "", "", nil, fmt.Errorf("format invalido (use json, csv, txt, xls o pdf)")
	}
}

func jsonMarshalDataset(ds empresaReporteDataset) ([]byte, error) {
	// Evitamos import cíclico: json ya está en reportes.go; reutilizamos encoding/json aquí vía helper en ese archivo.
	// (Implementación local simple)
	return json.Marshal(ds)
}

func sendReportesEmailWithAttachment(r *http.Request, dbSuper *sql.DB, empresaID int64, toEmail, subject, bodyText, attachmentName, attachmentContentType string, attachment []byte, metadataJSON string) error {
	toEmail = strings.TrimSpace(toEmail)
	if toEmail == "" {
		return fmt.Errorf("email requerido")
	}
	if _, err := mail.ParseAddress(toEmail); err != nil {
		return fmt.Errorf("correo destino invalido: %w", err)
	}
	if strings.TrimSpace(subject) == "" {
		subject = "Reporte del sistema"
	}
	if strings.TrimSpace(bodyText) == "" {
		bodyText = "Adjunto encontrarás el reporte solicitado."
	}
	if strings.TrimSpace(attachmentName) == "" || len(attachment) == 0 {
		return fmt.Errorf("adjunto invalido")
	}
	if strings.TrimSpace(attachmentContentType) == "" {
		attachmentContentType = "application/octet-stream"
	}

	if isEmpresaUsuarioMailTestMode(dbSuper) {
		return captureEmpresaUsuarioMailNotification(dbSuper, "reportes_export_email", empresaID, toEmail, subject, bodyText, attachmentName, metadataJSON, adminEmailFromRequest(r))
	}

	smtpEmail, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_email")
	if err != nil {
		return err
	}
	smtpPass, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_app_password")
	if err != nil {
		return err
	}
	smtpEmail = strings.TrimSpace(smtpEmail)
	smtpPass = strings.TrimSpace(smtpPass)
	if smtpEmail == "" || smtpPass == "" {
		return fmt.Errorf("smtp gmail no configurado")
	}
	smtpHost, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_host")
	smtpPort, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_port")
	fromName, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_from_name")
	smtpHost = strings.TrimSpace(smtpHost)
	if smtpHost == "" {
		smtpHost = "smtp.gmail.com"
	}
	smtpPort = strings.TrimSpace(smtpPort)
	if smtpPort == "" {
		smtpPort = "587"
	}
	fromName = strings.TrimSpace(fromName)
	if fromName == "" {
		fromName = "Powerful Control System"
	}
	hostForAuth := smtpHost
	if strings.Contains(smtpHost, ":") {
		if h, _, splitErr := net.SplitHostPort(smtpHost); splitErr == nil && strings.TrimSpace(h) != "" {
			hostForAuth = h
		}
	}
	addr := smtpHost
	if !strings.Contains(addr, ":") {
		addr = net.JoinHostPort(smtpHost, smtpPort)
	}

	from := (&mail.Address{Name: fromName, Address: smtpEmail}).String()
	to := (&mail.Address{Name: "", Address: toEmail}).String()
	boundary := "pcs-reportes-" + fmt.Sprint(time.Now().UnixNano())

	var msg bytes.Buffer
	msg.WriteString("From: " + from + "\r\n")
	msg.WriteString("To: " + to + "\r\n")
	msg.WriteString("Subject: " + subject + "\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: multipart/mixed; boundary=" + boundary + "\r\n\r\n")

	// Texto
	msg.WriteString("--" + boundary + "\r\n")
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
	msg.WriteString(bodyText + "\r\n\r\n")

	// Adjunto
	encoded := base64.StdEncoding.EncodeToString(attachment)
	msg.WriteString("--" + boundary + "\r\n")
	msg.WriteString("Content-Type: " + attachmentContentType + "\r\n")
	msg.WriteString("Content-Transfer-Encoding: base64\r\n")
	msg.WriteString("Content-Disposition: attachment; filename=\"" + strings.ReplaceAll(attachmentName, "\"", "") + "\"\r\n\r\n")
	// líneas de 76 chars para base64
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		msg.WriteString(encoded[i:end] + "\r\n")
	}
	msg.WriteString("\r\n--" + boundary + "--\r\n")

	auth := smtp.PlainAuth("", smtpEmail, smtpPass, hostForAuth)
	return smtp.SendMail(addr, auth, smtpEmail, []string{toEmail}, msg.Bytes())
}

func reportesDefaultEmailSubject(prefix, datasetTitle string, empresaLabel string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "Reporte"
	}
	datasetTitle = strings.TrimSpace(datasetTitle)
	empresaLabel = strings.TrimSpace(empresaLabel)
	if datasetTitle == "" {
		datasetTitle = "dataset"
	}
	if empresaLabel != "" {
		return prefix + ": " + datasetTitle + " (" + empresaLabel + ")"
	}
	return prefix + ": " + datasetTitle
}

func reportesResolveEmpresaLabel(dbSuper *sql.DB, empresaID int64) string {
	if empresaID <= 0 || dbSuper == nil {
		return ""
	}
	empresa, err := dbpkg.GetEmpresaByScopeID(dbSuper, empresaID)
	if err != nil || empresa == nil {
		return ""
	}
	name := strings.TrimSpace(empresa.Nombre)
	if name == "" {
		return fmt.Sprintf("Empresa #%d", empresaID)
	}
	return name
}

