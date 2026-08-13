package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

// RunEmpresaControlElectricoConnectivityScheduled evalua desconexiones desde el
// worker durable. La alerta se reclama por empresa, Raspberry y ultimo
// heartbeat, de modo que varias replicas no repitan la misma notificacion.
func RunEmpresaControlElectricoConnectivityScheduled(dbEmp, dbSuper *sql.DB) error {
	if dbEmp == nil {
		return fmt.Errorf("base empresarial no disponible")
	}
	candidates, err := dbpkg.ListEmpresaControlElectricoDisconnectCandidates(dbEmp)
	if err != nil {
		return err
	}
	var firstErr error
	for _, candidate := range candidates {
		if err := dbpkg.MarkEmpresaControlElectricoDisconnectAlerted(dbEmp, candidate); err != nil {
			continue
		}
		empresa, empresaErr := dbpkg.GetEmpresaByScopeID(dbEmp, candidate.EmpresaID)
		ownerEmail := ""
		empresaNombre := fmt.Sprintf("Empresa %d", candidate.EmpresaID)
		if empresaErr == nil && empresa != nil {
			ownerEmail = strings.ToLower(strings.TrimSpace(empresa.UsuarioCreador))
			empresaNombre = firstNonEmpty(strings.TrimSpace(empresa.Nombre), empresaNombre)
		}
		alertEmail := strings.ToLower(strings.TrimSpace(candidate.AlertEmail))
		recipientEmail := ownerEmail
		if !strings.Contains(recipientEmail, "@") {
			recipientEmail = alertEmail
		}
		title := "Raspberry desconectada en " + empresaNombre
		deviceName := firstNonEmpty(strings.TrimSpace(candidate.RaspberryNombre), "Raspberry "+fmt.Sprint(candidate.RaspberryID))
		message := fmt.Sprintf("%s no reporta actividad desde %s. PCS espero %d minutos para descartar un reinicio o una perdida breve de red. Revise energia, Internet y el servicio pcs-domotica-agent.", deviceName, candidate.LastSeen, candidate.GraceMinutes)
		if strings.Contains(recipientEmail, "@") {
			actor, actorErr := dbpkg.ResolveEmpresaBuzonActor(dbEmp, dbSuper, candidate.EmpresaID, recipientEmail)
			if actorErr != nil {
				actor = dbpkg.EmpresaBuzonActor{Tipo: "admin", Ref: recipientEmail, Email: recipientEmail, Nombre: recipientEmail, Rol: "administrador"}
			}
			_, buzErr := dbpkg.CreateEmpresaBuzonMensaje(dbEmp, dbpkg.EmpresaBuzonMensaje{
				EmpresaID: candidate.EmpresaID, DestinatarioTipo: actor.Tipo, DestinatarioRef: actor.Ref,
				DestinatarioEmail: actor.Email, DestinatarioNombre: actor.Nombre,
				RemitenteTipo: "sistema", RemitenteRef: "domotica_monitor", RemitenteNombre: "Domotica PCS",
				Titulo: title, Mensaje: message, Tipo: "domotica_desconexion", Prioridad: "alta", Modulo: "control_electrico",
				ReferenciaTipo: "raspberry", ReferenciaID: candidate.RaspberryID,
				EnlaceURL:      fmt.Sprintf("/administrar_empresa/control_electrico.html?pagina=raspberry&empresa_id=%d", candidate.EmpresaID),
				UsuarioCreador: "pcs-worker",
			})
			if buzErr != nil {
				log.Printf("[domotica_monitor] buzon empresa_id=%d raspberry_id=%d error: %v", candidate.EmpresaID, candidate.RaspberryID, buzErr)
				if firstErr == nil {
					firstErr = buzErr
				}
			}
		}
		emailResult := "no_configurado"
		emailError := ""
		if strings.Contains(alertEmail, "@") {
			emailResult = "enviado"
			metadata := fmt.Sprintf(`{"empresa_id":%d,"raspberry_id":%d}`, candidate.EmpresaID, candidate.RaspberryID)
			if sendErr := sendPCSSystemEmail(dbSuper, alertEmail, empresaNombre, title, message, "", "domotica_desconexion", metadata, "pcs-worker"); sendErr != nil {
				emailResult = "error"
				emailError = "No se pudo enviar el correo configurado"
				log.Printf("[domotica_monitor] email empresa_id=%d raspberry_id=%d error: %v", candidate.EmpresaID, candidate.RaspberryID, sendErr)
			}
		}
		_, _ = dbpkg.InsertEmpresaControlElectricoEvento(dbEmp, dbpkg.EmpresaControlElectricoEvento{
			EmpresaID: candidate.EmpresaID, RaspberryID: candidate.RaspberryID,
			Comando: "alerta_desconexion", EstadoObjetivo: "desconectado", Resultado: emailResult,
			Error: emailError, Actor: "pcs-worker", Origen: "monitor_conectividad",
		})
	}
	return firstErr
}
