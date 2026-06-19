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
	if isEmpresaUsuarioMailTestMode(dbSuper) {
		if notificationType == "" {
			notificationType = "sistema"
		}
		return captureEmpresaUsuarioMailNotification(dbSuper, notificationType, 0, toEmail, subject, textBody, "", metadataJSON, actorEmail)
	}

	fromName, fromEmail := corporateSystemSenderAddress(dbSuper, "soporte")
	if strings.TrimSpace(toName) == "" {
		toName = toEmail
	}
	from := (&mail.Address{Name: fromName, Address: fromEmail}).String()
	to := (&mail.Address{Name: toName, Address: toEmail}).String()
	boundary := "pcs-system-mail"
	if strings.TrimSpace(htmlBody) == "" {
		htmlBody = "<html><body><pre style=\"font-family:Arial,sans-serif;white-space:pre-wrap\">" + htmlEscape(textBody) + "</pre></body></html>"
	}
	msg := "From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/alternative; boundary=" + boundary + "\r\n\r\n" +
		"--" + boundary + "\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" + textBody + "\r\n" +
		"--" + boundary + "\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n" + htmlBody + "\r\n" +
		"--" + boundary + "--\r\n"
	return sendEmpresaUsuarioMailuMessage(dbSuper, fromEmail, toEmail, []byte(msg))
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
