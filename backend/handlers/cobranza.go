package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const cobranzaClienteNotificationType = "cobranza_cliente"

type empresaCobranzaRunResult struct {
	EmpresaID int64 `json:"empresa_id"`
	Evaluadas int   `json:"evaluadas"`
	Enviadas  int   `json:"enviadas"`
	Omitidas  int   `json:"omitidas"`
	Errores   int   `json:"errores"`
	DryRun    bool  `json:"dry_run"`
}

func EmpresaCobranzaHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "" {
			action = "dashboard"
		}
		usuario := strings.TrimSpace(adminEmailFromRequest(r))
		if usuario == "" {
			usuario = "sistema"
		}

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "configuracion", "config":
				cfg, err := dbpkg.GetEmpresaCobranzaConfiguracion(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar configuracion de cobranza", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, cfg)
				return
			case "dashboard":
				row, err := dbpkg.BuildEmpresaCobranzaDashboard(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar gestion de cobranza", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "cuentas":
				rows, err := dbpkg.ListEmpresaCobranzaCuentas(dbEmp, empresaID, dbpkg.EmpresaCobranzaCuentaFiltro{
					Estado:  r.URL.Query().Get("estado"),
					Query:   r.URL.Query().Get("q"),
					MoraMin: intQuery(r, "mora_min"),
					Limit:   300,
				})
				if err != nil {
					http.Error(w, "No se pudieron listar cuentas por cobrar", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "plantillas":
				rows, err := dbpkg.ListEmpresaCobranzaPlantillas(dbEmp, empresaID, 300)
				if err != nil {
					http.Error(w, "No se pudieron listar plantillas de cobranza", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "campanas":
				rows, err := dbpkg.ListEmpresaCobranzaCampanas(dbEmp, empresaID, 300)
				if err != nil {
					http.Error(w, "No se pudieron listar campanas de cobranza", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "gestiones":
				rows, err := dbpkg.ListEmpresaCobranzaGestiones(dbEmp, empresaID, 300)
				if err != nil {
					http.Error(w, "No se pudieron listar gestiones de cobranza", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "promesas":
				rows, err := dbpkg.ListEmpresaCobranzaPromesas(dbEmp, empresaID, r.URL.Query().Get("estado"), 300)
				if err != nil {
					http.Error(w, "No se pudieron listar promesas de pago", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			}

		case http.MethodPost, http.MethodPut:
			switch action {
			case "configuracion", "config":
				var payload dbpkg.EmpresaCobranzaConfiguracion
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.Usuario = usuario
				cfg, err := dbpkg.UpsertEmpresaCobranzaConfiguracion(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "configuracion": cfg})
				return
			case "ejecutar_recordatorios", "enviar_recordatorios":
				cfg, err := dbpkg.GetEmpresaCobranzaConfiguracion(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar configuracion", http.StatusInternalServerError)
					return
				}
				result := processEmpresaCobranzaRecordatorios(dbEmp, dbSuper, cfg, queryBool(r, "dry_run"), usuario)
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": result.Errores == 0, "resultado": result})
				return
			case "plantilla":
				var payload dbpkg.EmpresaCobranzaPlantilla
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.Usuario = usuario
				id, err := dbpkg.UpsertEmpresaCobranzaPlantilla(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			case "campana":
				var payload dbpkg.EmpresaCobranzaCampana
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.Usuario = usuario
				id, err := dbpkg.UpsertEmpresaCobranzaCampana(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			case "gestion", "simular_envio":
				var payload dbpkg.EmpresaCobranzaGestion
				_ = json.NewDecoder(r.Body).Decode(&payload)
				payload.EmpresaID = empresaID
				payload.Usuario = usuario
				if action == "simular_envio" {
					payload.Resultado = "enviado_simulado"
					if payload.Mensaje == "" {
						payload.Mensaje = "Mensaje de cobranza programado desde simulacion interna. No se envio por proveedor externo."
					}
				}
				row, promesa, err := dbpkg.CreateEmpresaCobranzaGestion(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "gestion": row, "promesa": promesa})
				return
			case "promesa":
				var payload dbpkg.EmpresaCobranzaPromesa
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.Usuario = usuario
				id, err := dbpkg.UpsertEmpresaCobranzaPromesa(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			case "marcar_promesa":
				var payload struct {
					ID            int64  `json:"id"`
					Estado        string `json:"estado"`
					Observaciones string `json:"observaciones"`
				}
				_ = json.NewDecoder(r.Body).Decode(&payload)
				if payload.ID <= 0 {
					payload.ID = int64Query(r, "id")
				}
				if payload.Estado == "" {
					payload.Estado = r.URL.Query().Get("estado")
				}
				row, err := dbpkg.UpdateEmpresaCobranzaPromesaEstado(dbEmp, empresaID, payload.ID, payload.Estado, usuario, payload.Observaciones)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "promesa": row})
				return
			case "seed_demo":
				if err := dbpkg.SeedEmpresaCobranzaDemo(dbEmp, empresaID, usuario); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			}
		}
		http.Error(w, fmt.Sprintf("Metodo o accion no permitida: %s", action), http.StatusMethodNotAllowed)
	}
}

func processEmpresaCobranzaRecordatorios(dbEmp, dbSuper *sql.DB, cfg dbpkg.EmpresaCobranzaConfiguracion, dryRun bool, actor string) empresaCobranzaRunResult {
	result := empresaCobranzaRunResult{EmpresaID: cfg.EmpresaID, DryRun: dryRun}
	if cfg.EmpresaID <= 0 || (!cfg.EmailActivo && !cfg.WhatsAppActivo) {
		return result
	}
	cuentas, err := dbpkg.ListEmpresaCobranzaCuentas(dbEmp, cfg.EmpresaID, dbpkg.EmpresaCobranzaCuentaFiltro{Limit: 500})
	if err != nil {
		result.Errores++
		return result
	}
	now := time.Now().In(time.Local)
	cutoff := now.AddDate(0, 0, cfg.DiasAntes).Format("2006-01-02")
	empresa, _ := dbpkg.GetEmpresaByScopeID(dbEmp, cfg.EmpresaID)
	empresaNombre := "la empresa"
	if empresa != nil && strings.TrimSpace(empresa.Nombre) != "" {
		empresaNombre = strings.TrimSpace(empresa.Nombre)
	}
	for _, cuenta := range cuentas {
		if cuenta.Saldo <= 0 || strings.TrimSpace(cuenta.FechaVencimiento) == "" || strings.TrimSpace(cuenta.FechaVencimiento) > cutoff {
			continue
		}
		result.Evaluadas++
		var email, telefono string
		if cuenta.ClienteID > 0 {
			_ = dbpkg.QueryRowCompat(dbEmp, `SELECT COALESCE(email,''),COALESCE(telefono,'') FROM clientes WHERE empresa_id=? AND id=? LIMIT 1`, cfg.EmpresaID, cuenta.ClienteID).Scan(&email, &telefono)
		}
		message := renderCobranzaMessage(cfg.Mensaje, cuenta, empresaNombre)
		subject := renderCobranzaMessage(cfg.Asunto, cuenta, empresaNombre)
		bucket := fmt.Sprintf("%d", (now.Unix()/86400)/int64(cfg.FrecuenciaDias))
		if cfg.EmailActivo {
			result.processCobranzaChannel(dbEmp, dbSuper, cfg, cuenta, "email", strings.TrimSpace(email), subject, message, bucket, actor, dryRun)
		}
		if cfg.WhatsAppActivo {
			result.processCobranzaChannel(dbEmp, dbSuper, cfg, cuenta, "whatsapp", strings.TrimSpace(telefono), subject, message, bucket, actor, dryRun)
		}
	}
	if !dryRun {
		_ = dbpkg.MarkEmpresaCobranzaUltimaEjecucion(dbEmp, cfg.EmpresaID, now.Format("2006-01-02 15:04:05"))
	}
	return result
}

func (result *empresaCobranzaRunResult) processCobranzaChannel(dbEmp, dbSuper *sql.DB, cfg dbpkg.EmpresaCobranzaConfiguracion, cuenta dbpkg.EmpresaCobranzaCuenta, canal, destino, subject, message, bucket, actor string, dryRun bool) {
	if strings.TrimSpace(destino) == "" {
		result.Omitidas++
		return
	}
	dedupe := fmt.Sprintf("%d:%s:%s", cuenta.ID, canal, bucket)
	if dryRun {
		result.Enviadas++
		return
	}
	inserted, recordErr := dbpkg.RegisterEmpresaCobranzaEnvio(dbEmp, cfg.EmpresaID, cuenta.ID, canal, dedupe, destino, "procesando", "", actor)
	if recordErr != nil {
		result.Errores++
		return
	}
	if !inserted {
		result.Omitidas++
		return
	}
	status := "sent"
	var sendErr error
	if canal == "email" {
		sendErr = sendPCSSystemEmail(dbSuper, destino, cuenta.ClienteNombre, subject, message, "", cobranzaClienteNotificationType, fmt.Sprintf(`{"empresa_id":%d,"cuenta_id":%d}`, cfg.EmpresaID, cuenta.ID), actor)
	} else {
		status, sendErr = sendPCSWhatsAppNotification(dbSuper, cobranzaClienteNotificationType, destino, subject+"\n\n"+message, fmt.Sprintf(`{"empresa_id":%d,"cuenta_id":%d}`, cfg.EmpresaID, cuenta.ID), actor)
	}
	if sendErr != nil {
		status = "error"
	}
	if err := dbpkg.UpdateEmpresaCobranzaEnvioResultado(dbEmp, cfg.EmpresaID, dedupe, status, errorText(sendErr)); err != nil {
		result.Errores++
		return
	}
	gestionResult := "enviado"
	if sendErr != nil {
		gestionResult = "fallido"
		result.Errores++
	} else {
		result.Enviadas++
	}
	_, _, _ = dbpkg.CreateEmpresaCobranzaGestion(dbEmp, dbpkg.EmpresaCobranzaGestion{EmpresaID: cfg.EmpresaID, CuentaID: cuenta.ID, ClienteID: cuenta.ClienteID, ClienteNombre: cuenta.ClienteNombre, DocumentoCodigo: cuenta.DocumentoCodigo, Canal: canal, Resultado: gestionResult, Mensaje: message, Contacto: destino, Usuario: actor, Estado: "activo", Observaciones: "Recordatorio automatico de cobranza"})
}

func renderCobranzaMessage(template string, cuenta dbpkg.EmpresaCobranzaCuenta, empresaNombre string) string {
	return strings.NewReplacer("{{cliente}}", strings.TrimSpace(cuenta.ClienteNombre), "{{documento}}", firstNonEmptyWhatsApp(cuenta.DocumentoCodigo, cuenta.Codigo), "{{saldo}}", fmt.Sprintf("%.0f %s", cuenta.Saldo, firstNonEmptyWhatsApp(cuenta.Moneda, "COP")), "{{vencimiento}}", strings.TrimSpace(cuenta.FechaVencimiento), "{{empresa}}", strings.TrimSpace(empresaNombre)).Replace(strings.TrimSpace(template))
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func StartEmpresaCobranzaRecordatoriosWorker(dbEmp, dbSuper *sql.DB, interval time.Duration, stop <-chan struct{}) {
	if interval <= 0 {
		interval = 12 * time.Hour
	}
	run := func() {
		cfgs, err := dbpkg.ListEmpresaCobranzaConfiguracionesActivas(dbEmp)
		if err != nil {
			log.Printf("[cobranza] worker configuraciones: %v", err)
			return
		}
		for _, cfg := range cfgs {
			if !cobranzaWorkerDue(cfg, time.Now().In(time.Local)) {
				continue
			}
			result := processEmpresaCobranzaRecordatorios(dbEmp, dbSuper, cfg, false, "sistema:cobranza")
			log.Printf("[cobranza] empresa=%d evaluadas=%d enviadas=%d omitidas=%d errores=%d", cfg.EmpresaID, result.Evaluadas, result.Enviadas, result.Omitidas, result.Errores)
		}
	}
	timer := time.NewTimer(2 * time.Minute)
	defer timer.Stop()
	select {
	case <-timer.C:
		run()
	case <-stop:
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			run()
		case <-stop:
			return
		}
	}
}

func cobranzaWorkerDue(cfg dbpkg.EmpresaCobranzaConfiguracion, now time.Time) bool {
	configured, err := time.Parse("15:04", strings.TrimSpace(cfg.HoraLocal))
	if err != nil {
		configured, _ = time.Parse("15:04", "09:00")
	}
	if now.Hour() != configured.Hour() {
		return false
	}
	last, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(cfg.UltimaEjecucion), now.Location())
	return err != nil || last.Format("2006-01-02") != now.Format("2006-01-02")
}
