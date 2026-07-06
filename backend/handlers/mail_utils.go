package handlers

import (
	"database/sql"
	"fmt"
	"net/mail"
	"strings"
)

func sendPCSSystemEmail(dbSuper *sql.DB, toEmail, toName, subject, textBody, htmlBody, notificationType, metadataJSON, actorEmail string) error {
	toEmail = strings.ToLower(strings.TrimSpace(toEmail))
	if _, err := mail.ParseAddress(toEmail); err != nil {
		return fmt.Errorf("correo destino invalido: %w", err)
	}
	if strings.TrimSpace(subject) == "" {
		return fmt.Errorf("asunto es obligatorio")
	}
	if strings.TrimSpace(textBody) == "" {
		return fmt.Errorf("cuerpo es obligatorio")
	}
	if notificationType == "" {
		notificationType = "sistema"
	}
	if isEmpresaUsuarioMailTestMode(dbSuper) {
		sendPCSWhatsAppForEmailRecipient(dbSuper, notificationType, toEmail, subject, textBody, metadataJSON, actorEmail)
		return captureEmpresaUsuarioMailNotification(dbSuper, notificationType, 0, toEmail, subject, textBody, "", metadataJSON, actorEmail)
	}
	if !isPCSEmailEventEnabled(dbSuper, notificationType) {
		sendPCSWhatsAppForEmailRecipient(dbSuper, notificationType, toEmail, subject, textBody, metadataJSON, actorEmail)
		return nil
	}

	if strings.TrimSpace(htmlBody) == "" {
		htmlBody = "<html><body><pre style=\"font-family:Arial,sans-serif;white-space:pre-wrap\">" + htmlEscape(textBody) + "</pre></body></html>"
	}
	fromName, fromEmail := corporateSystemSenderAddress(dbSuper, "soporte")
	msg := buildEmpresaUsuarioMultipartMessage(dbSuper, "https://powerfulcontrolsystem.com", fromName, fromEmail, toEmail, subject, textBody, htmlBody)
	err := sendEmpresaUsuarioMailuMessage(dbSuper, fromEmail, toEmail, []byte(msg))
	if err == nil {
		sendPCSWhatsAppForEmailRecipient(dbSuper, notificationType, toEmail, subject, textBody, metadataJSON, actorEmail)
	}
	return err
}

func isEmpresaUsuarioMailConfigError(err error) bool {
	if err == nil {
		return false
	}
	if isEmpresaUsuarioMailSecretDecryptError(err) {
		return true
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(msg, "gmail.smtp_") ||
		strings.Contains(msg, "smtp gmail no configurado") ||
		strings.Contains(msg, "mailu") ||
		strings.Contains(msg, "no configurado")
}
