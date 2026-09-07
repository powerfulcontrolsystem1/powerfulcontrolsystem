package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaCodigosDescuentoHandler gestiona CRUD de codigos de descuento por empresa.
func EmpresaCodigosDescuentoHandler(dbEmp *sql.DB, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "validar" {
				codigo := strings.TrimSpace(r.URL.Query().Get("codigo"))
				monto, _ := parseFloat64QueryOptional(r, "monto")
				carritoID, _ := parseInt64QueryOptional(r, "carrito_id")
				clienteID, _ := parseInt64QueryOptional(r, "cliente_id")
				canalVenta := strings.TrimSpace(r.URL.Query().Get("canal_venta"))
				aplicado, err := dbpkg.ResolveCodigoDescuentoParaMontoConContexto(
					dbEmp,
					empresaID,
					codigo,
					monto,
					dbpkg.CodigoDescuentoContexto{
						CarritoID:  carritoID,
						ClienteID:  clienteID,
						CanalVenta: canalVenta,
					},
				)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "aplicado": aplicado})
				return
			}
			if action == "redenciones" {
				codigoID, _ := parseInt64QueryOptional(r, "codigo_id")
				estado := strings.TrimSpace(r.URL.Query().Get("estado_redencion"))
				limit, _ := parseIntQueryOptional(r, "limit")
				rows, err := dbpkg.ListCodigoDescuentoRedencionesByEmpresa(dbEmp, empresaID, codigoID, estado, limit)
				if err != nil {
					log.Printf("[codigos_descuento] redenciones empresa_id=%d codigo_id=%d error: %v", empresaID, codigoID, err)
					http.Error(w, "No se pudieron listar las redenciones", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			}

			if id, _ := parseInt64QueryOptional(r, "id"); id > 0 {
				item, err := dbpkg.GetCodigoDescuentoByID(dbEmp, empresaID, id)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "codigo de descuento no encontrado", http.StatusNotFound)
						return
					}
					log.Printf("[codigos_descuento] get empresa_id=%d id=%d error: %v", empresaID, id, err)
					http.Error(w, "No se pudo consultar el codigo de descuento", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, item)
				return
			}

			q := strings.TrimSpace(r.URL.Query().Get("q"))
			estado := strings.TrimSpace(r.URL.Query().Get("estado"))
			includeInactive := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("include_inactive")), "1") ||
				strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("include_inactive")), "true")
			limit, _ := parseIntQueryOptional(r, "limit")
			offset, _ := parseIntQueryOptional(r, "offset")

			rows, err := dbpkg.GetCodigosDescuentoByEmpresa(dbEmp, empresaID, q, estado, includeInactive, limit, offset)
			if err != nil {
				log.Printf("[codigos_descuento] list empresa_id=%d error: %v", empresaID, err)
				http.Error(w, "No se pudieron listar los codigos de descuento", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, rows)
			return

		case http.MethodPost:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "enviar_correo" || action == "enviar_email" {
				empresaCodigosDescuentoEnviarCorreo(w, r, dbEmp, dbSuper)
				return
			}

			var payload dbpkg.CodigoDescuento
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))

			id, err := dbpkg.CreateCodigoDescuento(dbEmp, payload)
			if err != nil {
				status := codigoDescuentoWriteStatus(err)
				log.Printf("[codigos_descuento] create empresa_id=%d codigo=%q error: %v", payload.EmpresaID, payload.Codigo, err)
				http.Error(w, err.Error(), status)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			return

		case http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "activar" || action == "desactivar" {
				empresaID, err := parseEmpresaIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				id, err := parseInt64Query(r, "id")
				if err != nil {
					http.Error(w, "id es obligatorio", http.StatusBadRequest)
					return
				}
				estado := "inactivo"
				if action == "activar" {
					estado = "activo"
				}
				if err := dbpkg.SetCodigoDescuentoEstado(dbEmp, empresaID, id, estado); err != nil {
					status := codigoDescuentoWriteStatus(err)
					log.Printf("[codigos_descuento] set estado empresa_id=%d id=%d estado=%s error: %v", empresaID, id, estado, err)
					http.Error(w, err.Error(), status)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": estado})
				return
			}

			var payload dbpkg.CodigoDescuento
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.ID <= 0 {
				http.Error(w, "id es obligatorio", http.StatusBadRequest)
				return
			}
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			if err := dbpkg.UpdateCodigoDescuento(dbEmp, payload); err != nil {
				status := codigoDescuentoWriteStatus(err)
				log.Printf("[codigos_descuento] update empresa_id=%d id=%d error: %v", payload.EmpresaID, payload.ID, err)
				http.Error(w, err.Error(), status)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return

		case http.MethodDelete:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			id, err := parseInt64Query(r, "id")
			if err != nil {
				http.Error(w, "id es obligatorio", http.StatusBadRequest)
				return
			}
			if err := dbpkg.DeleteCodigoDescuento(dbEmp, empresaID, id); err != nil {
				status := codigoDescuentoWriteStatus(err)
				log.Printf("[codigos_descuento] delete empresa_id=%d id=%d error: %v", empresaID, id, err)
				http.Error(w, err.Error(), status)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

type codigoDescuentoEmailPayload struct {
	EmpresaID            int64  `json:"empresa_id"`
	ID                   int64  `json:"id"`
	Email                string `json:"email"`
	NombreDestinatario   string `json:"nombre_destinatario"`
	MensajePersonalizado string `json:"mensaje"`
}

func empresaCodigosDescuentoEnviarCorreo(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, dbSuper *sql.DB) {
	if dbSuper == nil {
		http.Error(w, "configuracion SMTP no disponible", http.StatusInternalServerError)
		return
	}

	var payload codigoDescuentoEmailPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	empresaID := payload.EmpresaID
	if empresaID <= 0 {
		if id, err := parseEmpresaIDQuery(r); err == nil {
			empresaID = id
		}
	}
	if empresaID <= 0 {
		http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
		return
	}

	codigoID := payload.ID
	if codigoID <= 0 {
		if id, err := parseInt64QueryOptional(r, "id"); err == nil {
			codigoID = id
		}
	}
	if codigoID <= 0 {
		http.Error(w, "id es obligatorio", http.StatusBadRequest)
		return
	}

	toEmail := strings.ToLower(strings.TrimSpace(payload.Email))
	if toEmail == "" {
		http.Error(w, "email es obligatorio", http.StatusBadRequest)
		return
	}
	if _, err := mail.ParseAddress(toEmail); err != nil {
		http.Error(w, "correo destino invalido", http.StatusBadRequest)
		return
	}

	item, err := dbpkg.GetCodigoDescuentoByID(dbEmp, empresaID, codigoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "codigo de descuento no encontrado", http.StatusNotFound)
			return
		}
		log.Printf("[codigos_descuento] enviar_correo get empresa_id=%d id=%d error: %v", empresaID, codigoID, err)
		http.Error(w, "No se pudo consultar el codigo de descuento", http.StatusInternalServerError)
		return
	}

	empresaNombre := fmt.Sprintf("Empresa %d", empresaID)
	if empresa, err := dbpkg.GetEmpresaByScopeID(dbSuper, empresaID); err == nil && empresa != nil && strings.TrimSpace(empresa.Nombre) != "" {
		empresaNombre = strings.TrimSpace(empresa.Nombre)
	}

	subject, body := buildCodigoDescuentoEmailContent(r, dbSuper, empresaID, empresaNombre, strings.TrimSpace(payload.NombreDestinatario), strings.TrimSpace(payload.MensajePersonalizado), item)
	if err := sendCodigoDescuentoEmail(r, dbSuper, empresaID, toEmail, subject, body, item.Codigo); err != nil {
		log.Printf("[codigos_descuento] enviar_correo empresa_id=%d id=%d email=%q error: %v", empresaID, codigoID, redactEmailForLog(toEmail), err)
		http.Error(w, "No se pudo enviar el correo. Intenta nuevamente o contacta al administrador.", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":           true,
		"email_sent":   true,
		"destinatario": toEmail,
		"codigo":       item.Codigo,
	})
}

func buildCodigoDescuentoEmailContent(r *http.Request, dbSuper *sql.DB, empresaID int64, empresaNombre, nombreDestinatario, mensaje string, item *dbpkg.CodigoDescuento) (string, string) {
	codigo := strings.TrimSpace(item.Codigo)
	if codigo == "" {
		codigo = "CODIGO"
	}
	nombreDestinatario = strings.TrimSpace(nombreDestinatario)
	if nombreDestinatario == "" {
		nombreDestinatario = "cliente"
	}
	empresaNombre = strings.TrimSpace(empresaNombre)
	if empresaNombre == "" {
		empresaNombre = fmt.Sprintf("Empresa %d", empresaID)
	}

	baseURL := strings.TrimRight(resolveBaseURLForConfirmation(r, dbSuper), "/")
	enlace := baseURL + "/administrar_empresa/codigos_de_descuento.html?empresa_id=" + strconv.FormatInt(empresaID, 10) + "&q=" + codigo

	valor := codigoDescuentoEmailValor(item)
	minimo := codigoDescuentoEmailMoney(item.MontoMinimoCompra, item.Moneda)
	if item.MontoMinimoCompra <= 0 {
		minimo = "Sin minimo de compra"
	}
	vence := strings.TrimSpace(item.FechaVencimiento)
	if vence == "" {
		vence = "Sin fecha de vencimiento configurada"
	}
	usos := fmt.Sprintf("%d disponibles de %d", codigoDescuentoMaxInt64(0, item.UsosMaximos-item.UsosActuales), item.UsosMaximos)
	if item.UsosMaximos <= 0 {
		usos = "Uso sujeto a disponibilidad"
	}

	subject := "Codigo de descuento " + codigo + " - " + empresaNombre
	lines := []string{
		"Hola " + nombreDestinatario + ",",
		"",
		empresaNombre + " te comparte este codigo de descuento:",
		"",
		"Codigo: " + codigo,
		"Descuento: " + valor,
		"Compra minima: " + minimo,
		"Vigencia: " + vence,
		"Usos: " + usos,
	}
	if canal := strings.TrimSpace(item.CanalVenta); canal != "" && !strings.EqualFold(canal, "todos") {
		lines = append(lines, "Canal: "+canal)
	}
	if segmento := strings.TrimSpace(item.SegmentoCliente); segmento != "" && !strings.EqualFold(segmento, "todos") {
		lines = append(lines, "Segmento: "+segmento)
	}
	if mensaje != "" {
		lines = append(lines, "", mensaje)
	}
	lines = append(lines,
		"",
		"Consulta o valida el codigo desde:",
		enlace,
		"",
		"Este codigo esta sujeto a vigencia, disponibilidad y reglas comerciales de la empresa.",
	)
	return subject, strings.Join(lines, "\n")
}

func codigoDescuentoEmailValor(item *dbpkg.CodigoDescuento) string {
	if item == nil {
		return ""
	}
	if strings.EqualFold(strings.TrimSpace(item.TipoDescuento), "porcentaje") {
		return fmt.Sprintf("%.2f%%", item.Valor)
	}
	return codigoDescuentoEmailMoney(item.Valor, item.Moneda)
}

func codigoDescuentoEmailMoney(valor float64, moneda string) string {
	moneda = strings.ToUpper(strings.TrimSpace(moneda))
	if moneda == "" {
		moneda = "COP"
	}
	return fmt.Sprintf("%s %.2f", moneda, valor)
}

func sendCodigoDescuentoEmail(r *http.Request, dbSuper *sql.DB, empresaID int64, toEmail, subject, body, codigo string) error {
	if isEmpresaUsuarioMailTestMode(dbSuper) {
		metadataJSON := fmt.Sprintf(`{"codigo":%q,"mail_mode":"test"}`, strings.TrimSpace(codigo))
		return captureEmpresaUsuarioMailNotification(dbSuper, "codigo_descuento_enviado", empresaID, toEmail, subject, body, strings.TrimSpace(codigo), metadataJSON, adminEmailFromRequest(r))
	}
	return sendEmpresaUsuarioMailuPlain(dbSuper, toEmail, subject, body)
}

func codigoDescuentoMaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func codigoDescuentoWriteStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if errors.Is(err, sql.ErrNoRows) {
		return http.StatusNotFound
	}
	lower := strings.ToLower(strings.TrimSpace(err.Error()))
	if strings.Contains(lower, "obligatorio") ||
		strings.Contains(lower, "invalido") ||
		strings.Contains(lower, "vencido") ||
		strings.Contains(lower, "sin usos") ||
		strings.Contains(lower, "monto minimo") ||
		strings.Contains(lower, "segmento") ||
		strings.Contains(lower, "canal") ||
		strings.Contains(lower, "horario") ||
		strings.Contains(lower, "dia actual") ||
		strings.Contains(lower, "antifraude") ||
		strings.Contains(lower, "ya fue aplicado") ||
		strings.Contains(lower, "limite de uso por cliente") ||
		strings.Contains(lower, "debe") {
		return http.StatusBadRequest
	}
	if strings.Contains(lower, "duplic") || strings.Contains(lower, "conflict") {
		return http.StatusConflict
	}
	return http.StatusInternalServerError
}

func parseFloat64QueryOptional(r *http.Request, key string) (float64, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return 0, nil
	}
	return strconv.ParseFloat(raw, 64)
}
