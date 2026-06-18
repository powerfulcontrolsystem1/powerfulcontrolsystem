package handlers

import (
	"database/sql"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
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
		if h, _, splitErr := net.SplitHostPort(smtpHost); splitErr == nil {
			hostForAuth = h
		}
	}
	addr := smtpHost
	if !strings.Contains(addr, ":") {
		addr = smtpHost + ":" + smtpPort
	}
	if strings.TrimSpace(toName) == "" {
		toName = toEmail
	}
	from := (&mail.Address{Name: fromName, Address: smtpEmail}).String()
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
	auth := smtp.PlainAuth("", smtpEmail, smtpPass, hostForAuth)
	return smtp.SendMail(addr, auth, smtpEmail, []string{toEmail}, []byte(msg))
}
