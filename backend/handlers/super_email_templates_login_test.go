package handlers

import (
	"strings"
	"testing"
)

func TestEmpresaConfirmationTemplateIsBusinessAgnostic(t *testing.T) {
	def, ok := getSuperEmailTemplateDefinition(superEmailTemplateKeyEmpresaConfirmation)
	if !ok {
		t.Fatal("plantilla de invitacion empresarial no encontrada")
	}
	combined := strings.ToLower(def.DefaultBodyText + "\n" + def.DefaultBodyHTML)
	if strings.Contains(combined, "sistema de motel") {
		t.Fatal("la invitacion empresarial no debe fijar una vertical de negocio")
	}
	if !strings.Contains(combined, "plataforma") {
		t.Fatal("la invitacion debe identificar la plataforma de forma generica")
	}
}

func TestNormalizeLegacyEmpresaConfirmationTemplate(t *testing.T) {
	textBody, htmlBody := normalizeLegacyEmpresaConfirmationTemplate(
		superEmailTemplateKeyEmpresaConfirmation,
		"te ha invitado a registrarte al sistema de motel Powerful Control System",
		"te ha invitado a registrarte al sistema de motel <strong>Powerful Control System</strong>",
	)
	if strings.Contains(strings.ToLower(textBody+htmlBody), "sistema de motel") {
		t.Fatal("la plantilla legada configurada debe normalizarse al renderizar")
	}
	if !strings.Contains(textBody, "plataforma Powerful Control System") ||
		!strings.Contains(htmlBody, "plataforma <strong>Powerful Control System</strong>") {
		t.Fatal("la normalizacion debe conservar el nombre de la plataforma")
	}
}
