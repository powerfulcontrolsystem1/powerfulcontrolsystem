package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"sort"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	superEmailTemplateKeyAdminConfirmation              = "admin_confirmation"
	superEmailTemplateKeyAdminScopedInvitation          = "admin_scoped_invitation"
	superEmailTemplateKeyAdminPortfolioDelegated        = "admin_portfolio_delegated"
	superEmailTemplateKeyEmpresaConfirmation            = "empresa_user_confirmation"
	superEmailTemplateKeyEmpresaAdminShareInvite        = "empresa_admin_share_invitation"
	superEmailTemplateKeyLicenciaActivation             = "licencia_activation_payment"
	superEmailTemplateKeyLicenciaSoftwarePDF            = "licencia_software_pdf"
	superEmailTemplateKeyLicenciaPaymentRejected        = "licencia_payment_rejected"
	superEmailTemplateKeyLicenciaExpiryWarning          = "licencia_expiry_warning"
	superEmailTemplateKeyLicenciaEmpresaDeletionWarning = "licencia_empresa_eliminacion_warning"
	superEmailTemplateKeyAdminPasswordRecovery          = "admin_password_recovery"
	superEmailTemplateKeyEmpresaPasswordRecovery        = "empresa_user_password_recovery" // #nosec G101 -- identificador de plantilla, no credencial.
	superEmailTemplateKeyServerRestartAlert             = "server_restart_alert"
)

type superEmailTemplateDefinition struct {
	Key             string
	Label           string
	Category        string
	Description     string
	Recommended     bool
	Variables       []string
	DefaultSubject  string
	DefaultBodyText string
	DefaultBodyHTML string
}

type superEmailTemplateItem struct {
	Key         string   `json:"key"`
	Label       string   `json:"label"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Recommended bool     `json:"recommended"`
	Variables   []string `json:"variables"`
	Subject     string   `json:"subject"`
	BodyText    string   `json:"body_text"`
	BodyHTML    string   `json:"body_html"`
	UpdatedAt   string   `json:"updated_at,omitempty"`
}

var superEmailTemplateDefinitions = []superEmailTemplateDefinition{
	{
		Key:             superEmailTemplateKeyAdminConfirmation,
		Label:           "Confirmación de correo administrativo",
		Category:        "confirmacion",
		Description:     "Correo que recibe el administrador cuando debe confirmar su cuenta del panel.",
		Variables:       []string{"name", "confirm_url", "login_url"},
		DefaultSubject:  "Confirma tu correo - Powerful Control System",
		DefaultBodyText: "Hola {{name}},\n\nPara activar tu cuenta, haz clic en el siguiente enlace:\n{{confirm_url}}\n\nDespués de confirmar, inicia sesión aquí:\n{{login_url}}\n\nSi no solicitaste esta cuenta, ignora este mensaje.\n",
		DefaultBodyHTML: "<html><body><p>Hola {{name}},</p><p>Para activar tu cuenta, haz clic en el siguiente enlace:</p><p><a href=\"{{confirm_url}}\">Confirmar correo</a></p><p>Después de confirmar, inicia sesión <a href=\"{{login_url}}\">aquí</a>.</p><p>Si no solicitaste esta cuenta, ignora este mensaje.</p></body></html>",
	},
	{
		Key:             superEmailTemplateKeyAdminScopedInvitation,
		Label:           "Invitacion a administrador delegado",
		Category:        "administracion",
		Description:     "Correo enviado cuando un administrador principal agrega otro administrador para operar sus empresas.",
		Variables:       []string{"name", "invited_by_name", "register_url", "login_url"},
		DefaultSubject:  "Te invitaron a administrar empresas en Powerful Control System",
		DefaultBodyText: "Hola {{name}},\n\n{{invited_by_name}} te invito a registrarte como administrador en Powerful Control System.\n\nPara aceptar la invitacion y crear tu clave, abre este enlace:\n{{register_url}}\n\nDespues de registrarte podras iniciar sesion aqui:\n{{login_url}}\n\nSi no esperabas esta invitacion, ignora este mensaje.\n",
		DefaultBodyHTML: "<html><body><p>Hola {{name}},</p><p><strong>{{invited_by_name}}</strong> te invito a registrarte como administrador en <strong>Powerful Control System</strong>.</p><p>Para aceptar la invitacion y crear tu clave, abre este enlace:</p><p><a href=\"{{register_url}}\">Aceptar invitacion y registrarme</a></p><p>Despues de registrarte podras iniciar sesion <a href=\"{{login_url}}\">aqui</a>.</p><p>Si no esperabas esta invitacion, ignora este mensaje.</p></body></html>",
	},
	{
		Key:             superEmailTemplateKeyAdminPortfolioDelegated,
		Label:           "Aviso de empresas compartidas a administrador registrado",
		Category:        "administracion",
		Description:     "Correo enviado cuando un administrador registrado recibe acceso al portafolio de empresas de otro administrador.",
		Variables:       []string{"name", "invited_by_name", "login_url"},
		DefaultSubject:  "Ahora puedes ver empresas compartidas en Powerful Control System",
		DefaultBodyText: "Hola {{name}},\n\n{{invited_by_name}} te compartio sus empresas en Powerful Control System.\n\nComo tu cuenta ya esta registrada, no debes crear otra cuenta. Inicia sesion y veras tus empresas mas las empresas compartidas:\n{{login_url}}\n\nSi no esperabas este acceso, informa al administrador que te invito.\n",
		DefaultBodyHTML: "<html><body><p>Hola {{name}},</p><p><strong>{{invited_by_name}}</strong> te compartio sus empresas en <strong>Powerful Control System</strong>.</p><p>Como tu cuenta ya esta registrada, no debes crear otra cuenta. Inicia sesion y veras tus empresas mas las empresas compartidas:</p><p><a href=\"{{login_url}}\">Abrir Powerful Control System</a></p><p>Si no esperabas este acceso, informa al administrador que te invito.</p></body></html>",
	},
	{
		Key:             superEmailTemplateKeyEmpresaConfirmation,
		Label:           "Confirmación de correo de usuario empresa",
		Category:        "confirmacion",
		Description:     "Invitación y confirmación para usuarios creados dentro de una empresa.",
		Variables:       []string{"name", "company_name", "confirm_url", "login_url", "admin_message", "admin_message_block_text", "admin_message_block_html"},
		DefaultSubject:  "Confirma tu correo - Powerful Control System",
		DefaultBodyText: "Hola {{name}},\n\nEl administrador de la empresa {{company_name}} te ha invitado a registrarte al sistema de motel Powerful Control System.\n\n{{admin_message_block_text}}Tu cuenta fue creada y necesita confirmar el correo para quedar habilitada.\nHaz clic en este enlace:\n{{confirm_url}}\n\nDespués de confirmar, inicia sesión aquí:\n{{login_url}}\n\nSi no solicitaste esta cuenta, ignora este mensaje.\n",
		DefaultBodyHTML: "<html><body><p>Hola {{name}},</p><p>El administrador de la empresa <strong>{{company_name}}</strong> te ha invitado a registrarte al sistema de motel <strong>Powerful Control System</strong>.</p>{{admin_message_block_html}}<p>Tu cuenta fue creada y necesita confirmar el correo para quedar habilitada.</p><p><a href=\"{{confirm_url}}\">Confirmar correo</a></p><p>Después de confirmar, inicia sesión <a href=\"{{login_url}}\">aquí</a>.</p><p>Si no solicitaste esta cuenta, ignora este mensaje.</p></body></html>",
	},
	{
		Key:             superEmailTemplateKeyEmpresaAdminShareInvite,
		Label:           "Invitación para compartir empresa entre administradores",
		Category:        "administracion",
		Description:     "Correo enviado a un administrador registrado para darle acceso compartido a una empresa.",
		Variables:       []string{"name", "company_name", "invited_by_name", "accept_url", "login_url", "admin_message", "admin_message_block_text", "admin_message_block_html"},
		DefaultSubject:  "Te compartieron una empresa en Powerful Control System",
		DefaultBodyText: "Hola {{name}},\n\n{{invited_by_name}} te compartió acceso administrativo a la empresa {{company_name}} en Powerful Control System.\n\n{{admin_message_block_text}}Para aceptar el acceso, haz clic en este enlace:\n{{accept_url}}\n\nEl enlace acepta la invitación directamente y abre seleccionar empresa con tu cuenta administrativa.\n\nSi prefieres entrar manualmente al login, usa:\n{{login_url}}\n\nSi no esperabas esta invitación, ignora este mensaje.\n",
		DefaultBodyHTML: "<html><body><p>Hola {{name}},</p><p><strong>{{invited_by_name}}</strong> te compartió acceso administrativo a la empresa <strong>{{company_name}}</strong> en Powerful Control System.</p>{{admin_message_block_html}}<p>Para aceptar el acceso, haz clic en este enlace:</p><p><a href=\"{{accept_url}}\">Aceptar acceso compartido</a></p><p>El enlace acepta la invitación directamente y abre seleccionar empresa con tu cuenta administrativa.</p><p>Si prefieres entrar manualmente al login, usa <a href=\"{{login_url}}\">este acceso</a>.</p><p>Si no esperabas esta invitación, ignora este mensaje.</p></body></html>",
	},
	{
		Key:             superEmailTemplateKeyLicenciaActivation,
		Label:           "Pago de licencia aprobado",
		Category:        "licencias",
		Description:     "Notificación enviada cuando una licencia queda activa tras un pago aprobado.",
		Variables:       []string{"company_name", "license_name", "provider", "reference", "start_date_line", "end_date_line", "reference_line", "license_name_line", "amount_paid_line", "discount_code_line", "discount_value_line", "original_value_line", "asesor_id_line", "amount_paid_line_html", "discount_code_line_html", "discount_value_line_html", "original_value_line_html", "asesor_id_line_html"},
		DefaultSubject:  "Tu licencia ya quedó activa",
		DefaultBodyText: "Hola,\n\nTu pago fue confirmado correctamente y la licencia ya quedó activa en Powerful Control System.\n\nEmpresa: {{company_name}}\n{{license_name_line}}{{start_date_line}}{{end_date_line}}{{reference_line}}Pasarela: {{provider}}\n{{amount_paid_line}}{{discount_code_line}}{{discount_value_line}}{{original_value_line}}{{asesor_id_line}}\n\nYa puedes ingresar al sistema y continuar con la operación normal de tu empresa.\n\nSi no reconoces este movimiento o necesitas ayuda, responde este correo.\n\nPowerful Control System\n",
		DefaultBodyHTML: "<html><body><p>Hola,</p><p>Tu pago fue confirmado correctamente y la licencia ya quedó activa en <strong>Powerful Control System</strong>.</p><p><strong>Empresa:</strong> {{company_name}}<br/>{{license_name_line}}{{start_date_line}}{{end_date_line}}{{reference_line}}<strong>Pasarela:</strong> {{provider}}<br/>{{amount_paid_line_html}}{{discount_code_line_html}}{{discount_value_line_html}}{{original_value_line_html}}{{asesor_id_line_html}}</p><p>Ya puedes ingresar al sistema y continuar con la operación normal de tu empresa.</p><p>Si no reconoces este movimiento o necesitas ayuda, responde este correo.</p><p>Powerful Control System</p></body></html>",
	},
	{
		Key:             superEmailTemplateKeyLicenciaSoftwarePDF,
		Label:           "PDF de licencia del software",
		Category:        "licencias",
		Description:     "Texto legal y operativo que se adjunta en PDF cuando una licencia queda activa.",
		Variables:       []string{"company_name", "company_nit", "license_name", "license_code", "provider", "reference", "issue_date", "start_date", "end_date", "amount_paid", "system_name"},
		DefaultSubject:  "Licencia de software Powerful Control System - {{company_name}}",
		DefaultBodyText: "LICENCIA DE USO DEL SOFTWARE {{system_name}}\n\nPowerful Control System concede a {{company_name}} una licencia de uso del sistema {{system_name}} bajo modalidad SaaS multiempresa, para operar los modulos contratados segun el plan activo.\n\nDatos de la licencia\nEmpresa: {{company_name}}\nNIT/Documento: {{company_nit}}\nPlan: {{license_name}}\nCodigo de licencia: {{license_code}}\nFecha de emision: {{issue_date}}\nVigencia: {{start_date}} hasta {{end_date}}\nReferencia de pago: {{reference}}\nPasarela: {{provider}}\nValor pagado: {{amount_paid}}\n\nCondiciones principales\n1. La licencia es personal para la empresa indicada, no exclusiva y no transferible sin autorizacion de Powerful Control System.\n2. El uso del software queda sujeto al plan contratado, a la vigencia pagada y a las politicas de seguridad, respaldo y soporte del sistema.\n3. La empresa usuaria es responsable de la veracidad de su informacion, de sus usuarios internos y del cumplimiento legal de sus operaciones.\n4. Powerful Control System podra suspender accesos ante uso indebido, fraude, incumplimiento de pago o riesgos de seguridad.\n5. Este documento sirve como constancia de activacion de la licencia del software.\n\nPowerful Control System\nSistema de facturacion electronica con domotica integrada\n",
		DefaultBodyHTML: "",
	},
	{
		Key:             superEmailTemplateKeyLicenciaPaymentRejected,
		Label:           "Pago de licencia rechazado",
		Category:        "licencias",
		Description:     "Notificación enviada cuando la pasarela reporta un pago rechazado/declinado y se ofrece reintentar.",
		Variables:       []string{"company_name", "license_name", "provider", "reference", "status", "retry_url", "reference_line", "license_name_line"},
		DefaultSubject:  "Tu pago fue rechazado (puedes reintentar)",
		DefaultBodyText: "Hola,\n\nLa pasarela reportó que el pago de tu licencia no se completó.\n\nEmpresa: {{company_name}}\n{{license_name_line}}{{reference_line}}Pasarela: {{provider}}\nEstado: {{status}}\n\nPuedes reintentar el pago más tarde desde este enlace:\n{{retry_url}}\n\nSi necesitas ayuda, responde este correo.\n\nPowerful Control System\n",
		DefaultBodyHTML: "<html><body><p>Hola,</p><p>La pasarela reportó que el pago de tu licencia no se completó.</p><p><strong>Empresa:</strong> {{company_name}}<br/>{{license_name_line}}{{reference_line}}<strong>Pasarela:</strong> {{provider}}<br/><strong>Estado:</strong> {{status}}</p><p><a href=\"{{retry_url}}\" style=\"display:inline-block;padding:12px 18px;background:#0f4c81;color:#ffffff;text-decoration:none;border-radius:10px;font-weight:700;\">Reintentar pago</a></p><p>Si el botón no abre correctamente, usa este enlace:</p><p><a href=\"{{retry_url}}\">{{retry_url}}</a></p><p>Si necesitas ayuda, responde este correo.</p><p>Powerful Control System</p></body></html>",
	},
	{
		Key:             superEmailTemplateKeyLicenciaExpiryWarning,
		Label:           "Licencia proxima a vencer",
		Category:        "licencias",
		Description:     "Aviso automatico al administrador de empresa cuando una licencia base o adicional esta cerca de vencer.",
		Recommended:     true,
		Variables:       []string{"name", "company_name", "license_name", "license_type", "days_remaining", "notice_threshold", "end_date", "renew_url"},
		DefaultSubject:  "Tu licencia esta proxima a vencer",
		DefaultBodyText: "Hola {{name}},\n\nLa licencia {{license_name}} de la empresa {{company_name}} esta proxima a vencer.\n\nTipo: {{license_type}}\nFecha de vencimiento: {{end_date}}\nDias restantes: {{days_remaining}}\n\nPuedes revisar o renovar la licencia desde este enlace:\n{{renew_url}}\n\nSi ya realizaste el pago, puedes ignorar este aviso.\n\nPowerful Control System\n",
		DefaultBodyHTML: "<html><body><p>Hola {{name}},</p><p>La licencia <strong>{{license_name}}</strong> de la empresa <strong>{{company_name}}</strong> esta proxima a vencer.</p><p><strong>Tipo:</strong> {{license_type}}<br/><strong>Fecha de vencimiento:</strong> {{end_date}}<br/><strong>Dias restantes:</strong> {{days_remaining}}</p><p><a href=\"{{renew_url}}\" style=\"display:inline-block;padding:12px 18px;background:#0f4c81;color:#ffffff;text-decoration:none;border-radius:10px;font-weight:700;\">Revisar licencia</a></p><p>Si el boton no abre correctamente, usa este enlace:</p><p><a href=\"{{renew_url}}\">{{renew_url}}</a></p><p>Si ya realizaste el pago, puedes ignorar este aviso.</p><p>Powerful Control System</p></body></html>",
	},
	{
		Key:             superEmailTemplateKeyLicenciaEmpresaDeletionWarning,
		Label:           "Preaviso de eliminacion por licencia vencida",
		Category:        "licencias",
		Description:     "Aviso al administrador antes de eliminar una empresa inactiva con licencia base vencida y periodo de retencion cumplido.",
		Recommended:     true,
		Variables:       []string{"name", "company_name", "last_license_end", "deletion_date", "retention_days", "notice_days", "renew_url", "support_url"},
		DefaultSubject:  "Tu empresa sera eliminada si no renuevas la licencia",
		DefaultBodyText: "Hola {{name}},\n\nLa empresa {{company_name}} tiene la licencia base vencida desde {{last_license_end}} y se encuentra sin operacion activa en Powerful Control System.\n\nSegun la politica de retencion configurada, el sistema espera {{retention_days}} dias y envia este preaviso {{notice_days}} dia(s) antes de la eliminacion programada.\n\nFecha programada de eliminacion: {{deletion_date}}\n\nPara conservar la empresa y sus datos, renueva o regulariza la licencia antes de esa fecha:\n{{renew_url}}\n\nAyuda y soporte:\n{{support_url}}\n\nSi ya renovaste la licencia o la empresa debe conservarse, responde este correo o contacta al soporte antes de la fecha indicada.\n\nPowerful Control System\n",
		DefaultBodyHTML: "<html><body><p>Hola {{name}},</p><p>La empresa <strong>{{company_name}}</strong> tiene la licencia base vencida desde <strong>{{last_license_end}}</strong> y se encuentra sin operacion activa en Powerful Control System.</p><p>Segun la politica de retencion configurada, el sistema espera <strong>{{retention_days}} dias</strong> y envia este preaviso <strong>{{notice_days}} dia(s)</strong> antes de la eliminacion programada.</p><p><strong>Fecha programada de eliminacion:</strong> {{deletion_date}}</p><p><a href=\"{{renew_url}}\" style=\"display:inline-block;padding:12px 18px;background:#7f1d1d;color:#ffffff;text-decoration:none;border-radius:10px;font-weight:700;\">Renovar licencia y conservar empresa</a></p><p>Si el boton no abre correctamente, usa este enlace:</p><p><a href=\"{{renew_url}}\">{{renew_url}}</a></p><p>Ayuda y soporte: <a href=\"{{support_url}}\">{{support_url}}</a></p><p>Si ya renovaste la licencia o la empresa debe conservarse, responde este correo o contacta al soporte antes de la fecha indicada.</p><p>Powerful Control System</p></body></html>",
	},
	{
		Key:             superEmailTemplateKeyAdminPasswordRecovery,
		Label:           "Recuperación de contraseña administrativa",
		Category:        "recomendadas",
		Description:     "Correo con enlace directo para restablecer contraseña del panel administrativo.",
		Recommended:     true,
		Variables:       []string{"name", "token", "reset_url"},
		DefaultSubject:  "Recuperacion de contraseña - Powerful Control System",
		DefaultBodyText: "Hola {{name}},\n\nRecibimos una solicitud para restablecer tu contraseña. Abre este enlace para definir una nueva clave:\n{{reset_url}}\n\nSi no solicitaste este cambio, ignora este mensaje.\n",
		DefaultBodyHTML: "<html><body><p>Hola {{name}},</p><p>Recibimos una solicitud para restablecer tu contraseña.</p><p><a href=\"{{reset_url}}\" style=\"display:inline-block;padding:12px 20px;background:#0f4c81;color:#ffffff;text-decoration:none;border-radius:8px;font-weight:700;\">Cambiar contraseña</a></p><p>Si el botón no abre correctamente, usa este enlace:</p><p><a href=\"{{reset_url}}\">{{reset_url}}</a></p><p>Si no solicitaste este cambio, ignora este mensaje.</p></body></html>",
	},
	{
		Key:             superEmailTemplateKeyEmpresaPasswordRecovery,
		Label:           "Recuperación de contraseña de usuario empresa",
		Category:        "recomendadas",
		Description:     "Correo con token para restablecer contraseña de un usuario interno de empresa.",
		Recommended:     true,
		Variables:       []string{"name", "token", "reset_url"},
		DefaultSubject:  "Recuperacion de contraseña - Powerful Control System",
		DefaultBodyText: "Hola {{name}},\n\nRecibimos una solicitud para restablecer tu contraseña.\nToken de recuperación (vigencia limitada):\n{{token}}\n\nAbre el login de usuario y usa el token para completar el restablecimiento:\n{{reset_url}}\n\nSi no solicitaste este cambio, ignora este mensaje.\n",
		DefaultBodyHTML: "<html><body><p>Hola {{name}},</p><p>Recibimos una solicitud para restablecer tu contraseña.</p><p><strong>Token de recuperación (vigencia limitada):</strong><br/>{{token}}</p><p><a href=\"{{reset_url}}\" style=\"display:inline-block;padding:12px 20px;background:#0f4c81;color:#ffffff;text-decoration:none;border-radius:8px;font-weight:700;\">Cambiar contraseña</a></p><p>Si el botón no abre correctamente, usa este enlace:</p><p><a href=\"{{reset_url}}\">{{reset_url}}</a></p><p>Si no solicitaste este cambio, ignora este mensaje.</p></body></html>",
	},
	{
		Key:             superEmailTemplateKeyServerRestartAlert,
		Label:           "Alerta de inicio o reinicio del servidor",
		Category:        "recomendadas",
		Description:     "Notificación operativa cuando el backend detecta inicio o reinicio inesperado.",
		Recommended:     true,
		Variables:       []string{"hostname", "event_date", "listen_addr_line", "reason", "unexpected_restart", "detail_line", "previous_status_block", "previous_start_block", "previous_stop_block", "previous_stop_reason_block"},
		DefaultSubject:  "[PCS] Inicio de servidor detectado ({{hostname}})",
		DefaultBodyText: "Inicio de servidor detectado.\n\nFecha evento: {{event_date}}\nHost: {{hostname}}\n{{listen_addr_line}}Motivo: {{reason}}\nReinicio inesperado: {{unexpected_restart}}\n{{detail_line}}{{previous_status_block}}{{previous_start_block}}{{previous_stop_block}}{{previous_stop_reason_block}}\nMensaje generado automaticamente por el backend PCS.",
	},
}

func init() {
	for i := range superEmailTemplateDefinitions {
		if superEmailTemplateDefinitions[i].Key != superEmailTemplateKeyLicenciaActivation {
			continue
		}
		superEmailTemplateDefinitions[i].Description = "Correo de bienvenida enviado cuando una licencia queda activa tras un pago aprobado. La factura electronica puede adjuntarse en PDF; la licencia del software se descarga desde Administrar empresa."
		superEmailTemplateDefinitions[i].Variables = []string{"company_name", "license_name", "provider", "reference", "license_download_url", "start_date_line", "end_date_line", "reference_line", "license_name_line", "amount_paid_line", "discount_code_line", "discount_value_line", "original_value_line", "asesor_id_line", "amount_paid_line_html", "discount_code_line_html", "discount_value_line_html", "original_value_line_html", "asesor_id_line_html"}
		superEmailTemplateDefinitions[i].DefaultSubject = "Bienvenido a Powerful Control System"
		superEmailTemplateDefinitions[i].DefaultBodyText = "Hola,\n\nBienvenido a Powerful Control System. Tu pago fue confirmado correctamente y tu licencia ya quedo activa.\n\nEmpresa: {{company_name}}\n{{license_name_line}}{{start_date_line}}{{end_date_line}}{{reference_line}}Pasarela: {{provider}}\n{{amount_paid_line}}{{discount_code_line}}{{discount_value_line}}{{original_value_line}}{{asesor_id_line}}\n\nAdjuntamos la factura electronica en PDF cuando la emision automatica esta habilitada. La licencia del software ya no se envia por correo; puedes descargarla desde Administrar empresa > Licencia del sistema:\n{{license_download_url}}\n\nYa puedes ingresar al sistema y continuar con la operacion normal de tu empresa.\n\nSi no reconoces este movimiento o necesitas ayuda, responde este correo.\n\nPowerful Control System\n"
		superEmailTemplateDefinitions[i].DefaultBodyHTML = "<html><body><p>Hola,</p><p>Bienvenido a <strong>Powerful Control System</strong>. Tu pago fue confirmado correctamente y tu licencia ya quedo activa.</p><p><strong>Empresa:</strong> {{company_name}}<br/>{{license_name_line}}{{start_date_line}}{{end_date_line}}{{reference_line}}<strong>Pasarela:</strong> {{provider}}<br/>{{amount_paid_line_html}}{{discount_code_line_html}}{{discount_value_line_html}}{{original_value_line_html}}{{asesor_id_line_html}}</p><p>Adjuntamos la factura electronica en PDF cuando la emision automatica esta habilitada.</p><p>La licencia del software ya no se envia por correo; puedes descargarla desde <strong>Administrar empresa &gt; Licencia del sistema</strong>: <a href=\"{{license_download_url}}\">descargar licencia</a>.</p><p>Ya puedes ingresar al sistema y continuar con la operacion normal de tu empresa.</p><p>Si no reconoces este movimiento o necesitas ayuda, responde este correo.</p><p>Powerful Control System</p></body></html>"
		break
	}
}

func superEmailTemplateConfigKey(key, field string) string {
	return "super.email_templates." + strings.TrimSpace(key) + "." + strings.TrimSpace(field)
}

func getSuperEmailTemplateDefinition(key string) (superEmailTemplateDefinition, bool) {
	for _, item := range superEmailTemplateDefinitions {
		if item.Key == strings.TrimSpace(key) {
			return item, true
		}
	}
	return superEmailTemplateDefinition{}, false
}

func isMissingConfigTableError(err error) bool {
	if err == nil {
		return false
	}
	low := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(low, "no such table: configuraciones") ||
		(strings.Contains(low, "relation") && strings.Contains(low, "configuraciones") && strings.Contains(low, "does not exist"))
}

func getSuperEmailTemplateConfigEntry(dbSuper *sql.DB, key string) (string, string, error) {
	if dbSuper == nil {
		return "", "", nil
	}
	value, _, _, updatedAt, err := dbpkg.GetConfigEntry(dbSuper, key)
	if err != nil {
		if isMissingConfigTableError(err) {
			return "", "", nil
		}
		return "", "", err
	}
	return value, updatedAt, nil
}

func listSuperEmailTemplates(dbSuper *sql.DB) ([]superEmailTemplateItem, error) {
	items := make([]superEmailTemplateItem, 0, len(superEmailTemplateDefinitions))
	for _, def := range superEmailTemplateDefinitions {
		subject, subjectUpdated, err := getSuperEmailTemplateConfigEntry(dbSuper, superEmailTemplateConfigKey(def.Key, "subject"))
		if err != nil {
			return nil, err
		}
		bodyText, bodyTextUpdated, err := getSuperEmailTemplateConfigEntry(dbSuper, superEmailTemplateConfigKey(def.Key, "body_text"))
		if err != nil {
			return nil, err
		}
		bodyHTML, bodyHTMLUpdated, err := getSuperEmailTemplateConfigEntry(dbSuper, superEmailTemplateConfigKey(def.Key, "body_html"))
		if err != nil {
			return nil, err
		}
		updatedAt := latestNonEmptyString(bodyHTMLUpdated, bodyTextUpdated, subjectUpdated)
		if strings.TrimSpace(subject) == "" {
			subject = def.DefaultSubject
		}
		if strings.TrimSpace(bodyText) == "" {
			bodyText = def.DefaultBodyText
		}
		if strings.TrimSpace(bodyHTML) == "" {
			bodyHTML = def.DefaultBodyHTML
		}
		items = append(items, superEmailTemplateItem{
			Key:         def.Key,
			Label:       def.Label,
			Category:    def.Category,
			Description: def.Description,
			Recommended: def.Recommended,
			Variables:   append([]string(nil), def.Variables...),
			Subject:     subject,
			BodyText:    bodyText,
			BodyHTML:    bodyHTML,
			UpdatedAt:   updatedAt,
		})
	}
	return items, nil
}

func latestNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func applySuperEmailTemplate(dbSuper *sql.DB, key string, values map[string]string) (string, string, string, error) {
	def, ok := getSuperEmailTemplateDefinition(key)
	if !ok {
		return "", "", "", fmt.Errorf("plantilla de correo no soportada: %s", key)
	}
	subject, _, err := getSuperEmailTemplateConfigEntry(dbSuper, superEmailTemplateConfigKey(def.Key, "subject"))
	if err != nil {
		return "", "", "", err
	}
	bodyText, _, err := getSuperEmailTemplateConfigEntry(dbSuper, superEmailTemplateConfigKey(def.Key, "body_text"))
	if err != nil {
		return "", "", "", err
	}
	bodyHTML, bodyHTMLUpdated, err := getSuperEmailTemplateConfigEntry(dbSuper, superEmailTemplateConfigKey(def.Key, "body_html"))
	if err != nil {
		return "", "", "", err
	}
	if strings.TrimSpace(subject) == "" {
		subject = def.DefaultSubject
	}
	if strings.TrimSpace(bodyText) == "" {
		bodyText = def.DefaultBodyText
	}
	bodyHTMLConfigured := strings.TrimSpace(bodyHTMLUpdated) != ""
	if strings.TrimSpace(bodyHTML) == "" && !bodyHTMLConfigured {
		bodyHTML = def.DefaultBodyHTML
	}
	subject = replaceTemplateVariables(subject, values)
	bodyText = replaceTemplateVariables(bodyText, values)
	bodyHTML = replaceTemplateVariables(bodyHTML, values)
	if strings.TrimSpace(bodyHTML) == "" && strings.TrimSpace(bodyText) != "" {
		bodyHTML = plainTextEmailToHTML(bodyText)
	}
	return subject, bodyText, bodyHTML, nil
}

func replaceTemplateVariables(input string, values map[string]string) string {
	output := input
	if len(values) == 0 {
		return output
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		output = strings.ReplaceAll(output, "{{"+key+"}}", values[key])
	}
	return output
}

func plainTextEmailToHTML(input string) string {
	safe := html.EscapeString(strings.ReplaceAll(strings.ReplaceAll(input, "\r\n", "\n"), "\r", "\n"))
	safe = strings.ReplaceAll(safe, "\n\n", "</p><p>")
	safe = strings.ReplaceAll(safe, "\n", "<br/>")
	return "<html><body><p>" + safe + "</p></body></html>"
}

func templateLine(label, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return label + value + "\n"
}

func templateLineHTML(label, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return "<strong>" + html.EscapeString(label) + "</strong> " + html.EscapeString(value) + "<br/>"
}

func templateParagraphText(title, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return title + "\n" + value + "\n\n"
}

func templateParagraphHTML(title, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return "<p><strong>" + html.EscapeString(title) + "</strong><br/>" + strings.ReplaceAll(html.EscapeString(value), "\n", "<br/>") + "</p>"
}

func SuperEmailTemplatesHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper); !ok {
			return
		}
		switch r.Method {
		case http.MethodGet:
			items, err := listSuperEmailTemplates(dbSuper)
			if err != nil {
				http.Error(w, "failed to read email templates: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			encodeJSONResponse(w, map[string]interface{}{"templates": items})
			return

		case http.MethodPut, http.MethodPost:
			var payload struct {
				Templates []superEmailTemplateItem `json:"templates"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
				return
			}
			if len(payload.Templates) == 0 {
				http.Error(w, "templates requeridas", http.StatusBadRequest)
				return
			}
			for _, item := range payload.Templates {
				def, ok := getSuperEmailTemplateDefinition(item.Key)
				if !ok {
					http.Error(w, "plantilla no soportada: "+item.Key, http.StatusBadRequest)
					return
				}
				subject := strings.TrimSpace(item.Subject)
				bodyText := strings.TrimSpace(item.BodyText)
				bodyHTML := strings.TrimSpace(item.BodyHTML)
				if subject == "" {
					subject = def.DefaultSubject
				}
				if bodyText == "" {
					bodyText = def.DefaultBodyText
				}
				if err := dbpkg.SetConfigValue(dbSuper, superEmailTemplateConfigKey(def.Key, "subject"), subject, false); err != nil {
					http.Error(w, "failed to save template subject: "+err.Error(), http.StatusInternalServerError)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, superEmailTemplateConfigKey(def.Key, "body_text"), bodyText, false); err != nil {
					http.Error(w, "failed to save template body_text: "+err.Error(), http.StatusInternalServerError)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, superEmailTemplateConfigKey(def.Key, "body_html"), bodyHTML, false); err != nil {
					http.Error(w, "failed to save template body_html: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			items, err := listSuperEmailTemplates(dbSuper)
			if err != nil {
				http.Error(w, "failed to reload email templates: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			encodeJSONResponse(w, map[string]interface{}{"saved": true, "templates": items})
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}
