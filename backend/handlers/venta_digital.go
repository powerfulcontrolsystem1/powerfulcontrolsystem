package handlers

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const ventaDigitalMailTipoEntrega = "venta_digital_entrega"

type superVentaDigitalConfigPayload struct {
	NombreTienda      string `json:"nombre_tienda"`
	DescripcionTienda string `json:"descripcion_tienda"`
	LogoURL           string `json:"logo_url"`
	BannerURL         string `json:"banner_url"`
	ColorPrimario     string `json:"color_primario"`
	Moneda            string `json:"moneda"`
	WompiActivo       *bool  `json:"wompi_activo"`
	Observaciones     string `json:"observaciones"`
}

type superVentaDigitalItemPayload struct {
	ID                      int64   `json:"id"`
	CodigoPublico           string  `json:"codigo_publico"`
	Nombre                  string  `json:"nombre"`
	Descripcion             string  `json:"descripcion"`
	Precio                  float64 `json:"precio"`
	Moneda                  string  `json:"moneda"`
	ImagenURL               string  `json:"imagen_url"`
	LicenciaCodigo          string  `json:"licencia_codigo"`
	InstruccionesArchivoURL string  `json:"instrucciones_archivo_url"`
	OrdenVisual             int     `json:"orden_visual"`
	Destacado               bool    `json:"destacado"`
	Observaciones           string  `json:"observaciones"`
}

type ventaDigitalCrearPagoPayload struct {
	ItemID            int64  `json:"item_id"`
	CompradorNombre   string `json:"comprador_nombre"`
	CompradorEmail    string `json:"comprador_email"`
	CompradorTelefono string `json:"comprador_telefono"`
	AcceptTerms       bool   `json:"accept_terms"`
}

func superVentaDigitalNormalizeAction(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "listar", "list", "catalogo", "catalog":
		return "catalogo"
	case "detalle", "item", "get":
		return "detalle"
	case "crear", "create":
		return "crear"
	case "actualizar", "update", "editar", "edit":
		return "actualizar"
	case "activar":
		return "activar"
	case "desactivar", "delete", "eliminar":
		return "desactivar"
	case "config", "configuracion", "settings":
		return "config"
	case "ordenes", "orders":
		return "ordenes"
	case "subir_imagen", "upload_image", "imagen":
		return "subir_imagen"
	case "subir_instrucciones", "upload_instructions", "instrucciones":
		return "subir_instrucciones"
	default:
		return ""
	}
}

func ventaDigitalNormalizeAction(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "catalogo", "catalog", "tienda", "store":
		return "catalogo"
	case "crear_pago", "pagar", "create_payment", "payment":
		return "crear_pago"
	case "estado_pago", "status", "payment_status":
		return "estado_pago"
	default:
		return ""
	}
}

func ventaDigitalMapWompiStatus(raw string) string {
	status := strings.ToUpper(strings.TrimSpace(raw))
	switch status {
	case "APPROVED":
		return "aprobado"
	case "DECLINED", "VOIDED", "ERROR":
		return "rechazado"
	case "PENDING":
		return "pendiente"
	default:
		return "pendiente"
	}
}

func sanitizeVentaDigitalConfigForPublic(cfg dbpkg.SuperVentaDigitalConfig) map[string]interface{} {
	return map[string]interface{}{
		"nombre_tienda":      cfg.NombreTienda,
		"descripcion_tienda": cfg.DescripcionTienda,
		"logo_url":           cfg.LogoURL,
		"banner_url":         cfg.BannerURL,
		"color_primario":     cfg.ColorPrimario,
		"moneda":             cfg.Moneda,
		"wompi_activo":       cfg.WompiActivo,
	}
}

func sanitizeVentaDigitalItemsForPublic(items []dbpkg.SuperVentaDigitalItem) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]interface{}{
			"id":             item.ID,
			"codigo_publico": item.CodigoPublico,
			"nombre":         item.Nombre,
			"descripcion":    item.Descripcion,
			"precio":         item.Precio,
			"moneda":         item.Moneda,
			"imagen_url":     item.ImagenURL,
			"orden_visual":   item.OrdenVisual,
			"destacado":      item.Destacado,
		})
	}
	return out
}

func sanitizeVentaDigitalOrderForPublic(order dbpkg.SuperVentaDigitalOrder) map[string]interface{} {
	return map[string]interface{}{
		"id":                  order.ID,
		"codigo_orden":        order.CodigoOrden,
		"item_id":             order.ItemID,
		"item_nombre":         order.ItemNombre,
		"item_precio":         order.ItemPrecio,
		"item_moneda":         order.ItemMoneda,
		"comprador_nombre":    order.CompradorNombre,
		"comprador_email":     order.CompradorEmail,
		"comprador_telefono":  order.CompradorTelefono,
		"metodo_pago":         order.MetodoPago,
		"estado_pago":         order.EstadoPago,
		"transaction_id":      order.TransactionID,
		"referencia_externa":  order.ReferenciaExterna,
		"pagado_en":           order.PagadoEn,
		"correo_entregado":    order.CorreoEntregado,
		"correo_entregado_en": order.CorreoEntregadoEn,
		"fecha_creacion":      order.FechaCreacion,
	}
}

func resolveVentaDigitalBaseURL(r *http.Request, dbSuper *sql.DB) string {
	if configured, err := getDecryptedConfigValue(dbSuper, "gmail.confirm_base_url"); err == nil {
		configured = strings.TrimSpace(configured)
		if configured != "" {
			return strings.TrimRight(configured, "/")
		}
	}

	if r == nil {
		return "http://localhost:8080"
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if xfProto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); xfProto != "" {
		scheme = xfProto
	}
	host := strings.TrimSpace(r.Host)
	if host == "" {
		host = "localhost:8080"
	}
	return scheme + "://" + host
}

func buildVentaDigitalDeliveryMailBody(order dbpkg.SuperVentaDigitalOrder, licenseCode string, instruccionesURL string) string {
	name := strings.TrimSpace(order.CompradorNombre)
	if name == "" {
		name = "cliente"
	}
	itemName := strings.TrimSpace(order.ItemNombre)
	if itemName == "" {
		itemName = "producto digital"
	}
	body := "Hola " + name + ",\r\n\r\n"
	body += "Tu pago fue aprobado para: " + itemName + ".\r\n"
	body += "Codigo de licencia: " + licenseCode + "\r\n"
	if strings.TrimSpace(instruccionesURL) != "" {
		body += "Archivo de instrucciones de instalacion: " + instruccionesURL + "\r\n"
	}
	body += "\r\nCodigo de orden: " + strings.TrimSpace(order.CodigoOrden) + "\r\n"
	body += "Si tienes dudas, contacta al super administrador.\r\n"
	body += "\r\nPowerful Control System\r\n"
	return body
}

func sendVentaDigitalDeliveryEmail(r *http.Request, dbSuper *sql.DB, order dbpkg.SuperVentaDigitalOrder) error {
	toEmail := strings.TrimSpace(order.CompradorEmail)
	if toEmail == "" {
		return fmt.Errorf("comprador_email no disponible para entrega")
	}
	if _, err := mail.ParseAddress(toEmail); err != nil {
		return fmt.Errorf("correo del comprador invalido: %w", err)
	}

	licenseCode := strings.TrimSpace(order.LicenciaCodigoEnviado)
	if licenseCode == "" {
		return fmt.Errorf("codigo de licencia no disponible para entrega")
	}

	baseURL := resolveVentaDigitalBaseURL(r, dbSuper)
	instructionsURL := dbpkg.BuildSuperVentaDigitalInstructionAbsoluteURL(baseURL, order.InstruccionesArchivoURL)
	subject := "Tu licencia digital - Powerful Control System"
	body := buildVentaDigitalDeliveryMailBody(order, licenseCode, instructionsURL)

	if isEmpresaUsuarioMailTestMode(dbSuper) {
		metadataJSON := fmt.Sprintf(`{"order_code":%q,"item_id":%d,"mail_mode":"test"}`, order.CodigoOrden, order.ItemID)
		if err := captureEmpresaUsuarioMailNotification(
			dbSuper,
			ventaDigitalMailTipoEntrega,
			0,
			toEmail,
			subject,
			body,
			order.CodigoOrden,
			metadataJSON,
			"sistema",
		); err != nil {
			return err
		}
		return nil
	}

	return sendEmpresaUsuarioMailuPlain(dbSuper, toEmail, subject, body)
}

func syncVentaDigitalDeliveryIfApproved(r *http.Request, dbSuper *sql.DB, orderCode string) (bool, string, error) {
	order, err := dbpkg.GetSuperVentaDigitalOrderByCodigo(dbSuper, orderCode)
	if err != nil {
		return false, "order_not_found", err
	}

	if !strings.EqualFold(strings.TrimSpace(order.EstadoPago), "aprobado") {
		return false, "payment_not_approved", nil
	}
	if order.CorreoEntregado {
		return true, "already_delivered", nil
	}

	if err := sendVentaDigitalDeliveryEmail(r, dbSuper, order); err != nil {
		errMsg := strings.TrimSpace(err.Error())
		if len(errMsg) > 300 {
			errMsg = errMsg[:300]
		}
		_ = dbpkg.SetSuperVentaDigitalOrderDelivery(dbSuper, order.CodigoOrden, false, "", errMsg, "", "")
		return false, "delivery_error", err
	}

	deliveredAt := time.Now().In(time.Local).Format("2006-01-02 15:04:05")
	if err := dbpkg.SetSuperVentaDigitalOrderDelivery(dbSuper, order.CodigoOrden, true, deliveredAt, "", order.LicenciaCodigoEnviado, order.InstruccionesArchivoURL); err != nil {
		return false, "delivery_mark_error", err
	}
	return true, "delivered", nil
}

func processVentaDigitalPaymentStatusUpdate(r *http.Request, dbSuper *sql.DB, transactionID, reference, providerStatus, providerPayload string) (bool, bool, string, error) {
	order, found, err := dbpkg.FindSuperVentaDigitalOrderByPaymentContext(dbSuper, transactionID, reference)
	if err != nil {
		return false, false, "lookup_error", err
	}
	if !found {
		return false, false, "not_found", nil
	}

	localStatus := ventaDigitalMapWompiStatus(providerStatus)
	pagadoEn := ""
	if localStatus == "aprobado" {
		pagadoEn = time.Now().In(time.Local).Format("2006-01-02 15:04:05")
	}
	if err := dbpkg.UpdateSuperVentaDigitalOrderPayment(dbSuper, order.CodigoOrden, localStatus, transactionID, reference, providerPayload, pagadoEn, "status_sync"); err != nil {
		return true, false, "update_error", err
	}

	if localStatus != "aprobado" {
		return true, false, "payment_not_approved", nil
	}

	delivered, stage, err := syncVentaDigitalDeliveryIfApproved(r, dbSuper, order.CodigoOrden)
	if err != nil {
		return true, false, stage, err
	}
	return true, delivered, stage, nil
}

// SuperVentaDigitalHandler gestiona configuracion, catalogo y ordenes desde panel super.
func SuperVentaDigitalHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper); !ok {
			return
		}
		action := superVentaDigitalNormalizeAction(r.URL.Query().Get("action"))
		if action == "" {
			http.Error(w, "action invalida", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "catalogo":
				handleSuperVentaDigitalList(w, r, dbSuper)
			case "detalle":
				handleSuperVentaDigitalDetail(w, r, dbSuper)
			case "config":
				handleSuperVentaDigitalConfigGet(w, r, dbSuper)
			case "ordenes":
				handleSuperVentaDigitalOrders(w, r, dbSuper)
			default:
				http.Error(w, "action invalida", http.StatusBadRequest)
			}
		case http.MethodPost:
			switch action {
			case "crear", "catalogo":
				handleSuperVentaDigitalCreate(w, r, dbSuper)
			case "config":
				handleSuperVentaDigitalConfigUpsert(w, r, dbSuper)
			case "subir_imagen":
				handleSuperVentaDigitalUploadImage(w, r, dbSuper)
			case "subir_instrucciones":
				handleSuperVentaDigitalUploadInstructions(w, r, dbSuper)
			default:
				http.Error(w, "action invalida", http.StatusBadRequest)
			}
		case http.MethodPut, http.MethodPatch:
			switch action {
			case "actualizar", "catalogo":
				handleSuperVentaDigitalUpdate(w, r, dbSuper)
			case "activar", "desactivar":
				handleSuperVentaDigitalToggle(w, r, dbSuper, action)
			case "config":
				handleSuperVentaDigitalConfigUpsert(w, r, dbSuper)
			default:
				http.Error(w, "action invalida", http.StatusBadRequest)
			}
		case http.MethodDelete:
			handleSuperVentaDigitalToggle(w, r, dbSuper, "desactivar")
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func handleSuperVentaDigitalList(w http.ResponseWriter, r *http.Request, dbSuper *sql.DB) {
	limit, err := parseIntQueryOptional(r, "limit")
	if err != nil {
		http.Error(w, "limit invalido", http.StatusBadRequest)
		return
	}
	offset, err := parseIntQueryOptional(r, "offset")
	if err != nil {
		http.Error(w, "offset invalido", http.StatusBadRequest)
		return
	}

	rows, total, err := dbpkg.ListSuperVentaDigitalItems(dbSuper, dbpkg.SuperVentaDigitalItemsFilter{
		IncludeInactive: queryBool(r, "include_inactive"),
		Q:               strings.TrimSpace(r.URL.Query().Get("q")),
		Limit:           limit,
		Offset:          offset,
	})
	if err != nil {
		http.Error(w, "No se pudo listar catalogo digital", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"total": total,
		"rows":  rows,
	})
}

func handleSuperVentaDigitalDetail(w http.ResponseWriter, r *http.Request, dbSuper *sql.DB) {
	itemID, err := parseInt64QueryOptional(r, "id")
	if err != nil || itemID <= 0 {
		http.Error(w, "id es obligatorio", http.StatusBadRequest)
		return
	}
	item, err := dbpkg.GetSuperVentaDigitalItemByID(dbSuper, itemID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "item no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo consultar item", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":   true,
		"item": item,
	})
}

func handleSuperVentaDigitalCreate(w http.ResponseWriter, r *http.Request, dbSuper *sql.DB) {
	var payload superVentaDigitalItemPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	id, err := dbpkg.CreateSuperVentaDigitalItem(dbSuper, dbpkg.SuperVentaDigitalItem{
		CodigoPublico:           payload.CodigoPublico,
		Nombre:                  payload.Nombre,
		Descripcion:             payload.Descripcion,
		Precio:                  payload.Precio,
		Moneda:                  payload.Moneda,
		ImagenURL:               payload.ImagenURL,
		LicenciaCodigo:          payload.LicenciaCodigo,
		InstruccionesArchivoURL: payload.InstruccionesArchivoURL,
		OrdenVisual:             payload.OrdenVisual,
		Destacado:               payload.Destacado,
		UsuarioCreador:          adminEmailFromRequest(r),
		Estado:                  "activo",
		Observaciones:           payload.Observaciones,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	item, err := dbpkg.GetSuperVentaDigitalItemByID(dbSuper, id)
	if err != nil {
		http.Error(w, "item creado pero no se pudo consultar", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"ok":   true,
		"item": item,
	})
}

func handleSuperVentaDigitalUpdate(w http.ResponseWriter, r *http.Request, dbSuper *sql.DB) {
	var payload superVentaDigitalItemPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.ID <= 0 {
		if qID, qErr := parseInt64QueryOptional(r, "id"); qErr == nil && qID > 0 {
			payload.ID = qID
		}
	}
	if payload.ID <= 0 {
		http.Error(w, "id es obligatorio", http.StatusBadRequest)
		return
	}

	err := dbpkg.UpdateSuperVentaDigitalItem(dbSuper, dbpkg.SuperVentaDigitalItem{
		ID:                      payload.ID,
		CodigoPublico:           payload.CodigoPublico,
		Nombre:                  payload.Nombre,
		Descripcion:             payload.Descripcion,
		Precio:                  payload.Precio,
		Moneda:                  payload.Moneda,
		ImagenURL:               payload.ImagenURL,
		LicenciaCodigo:          payload.LicenciaCodigo,
		InstruccionesArchivoURL: payload.InstruccionesArchivoURL,
		OrdenVisual:             payload.OrdenVisual,
		Destacado:               payload.Destacado,
		UsuarioCreador:          adminEmailFromRequest(r),
		Estado:                  "activo",
		Observaciones:           payload.Observaciones,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "item no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	item, err := dbpkg.GetSuperVentaDigitalItemByID(dbSuper, payload.ID)
	if err != nil {
		http.Error(w, "item actualizado pero no se pudo consultar", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":   true,
		"item": item,
	})
}

func handleSuperVentaDigitalToggle(w http.ResponseWriter, r *http.Request, dbSuper *sql.DB, action string) {
	itemID, err := parseInt64QueryOptional(r, "id")
	if err != nil || itemID <= 0 {
		http.Error(w, "id es obligatorio", http.StatusBadRequest)
		return
	}
	estado := "activo"
	if action == "desactivar" {
		estado = "inactivo"
	}
	if err := dbpkg.SetSuperVentaDigitalItemEstadoByID(dbSuper, itemID, estado); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "item no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo actualizar estado", http.StatusInternalServerError)
		return
	}
	item, _ := dbpkg.GetSuperVentaDigitalItemByID(dbSuper, itemID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":   true,
		"item": item,
	})
}

func handleSuperVentaDigitalConfigGet(w http.ResponseWriter, _ *http.Request, dbSuper *sql.DB) {
	cfg, err := dbpkg.GetSuperVentaDigitalConfig(dbSuper)
	if err != nil {
		http.Error(w, "No se pudo consultar configuracion", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"config": cfg,
	})
}

func handleSuperVentaDigitalConfigUpsert(w http.ResponseWriter, r *http.Request, dbSuper *sql.DB) {
	var payload superVentaDigitalConfigPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	wompiActivo := true
	if payload.WompiActivo != nil {
		wompiActivo = *payload.WompiActivo
	}

	_, err := dbpkg.UpsertSuperVentaDigitalConfig(dbSuper, dbpkg.SuperVentaDigitalConfig{
		NombreTienda:      payload.NombreTienda,
		DescripcionTienda: payload.DescripcionTienda,
		LogoURL:           payload.LogoURL,
		BannerURL:         payload.BannerURL,
		ColorPrimario:     payload.ColorPrimario,
		Moneda:            payload.Moneda,
		WompiActivo:       wompiActivo,
		UsuarioCreador:    adminEmailFromRequest(r),
		Estado:            "activo",
		Observaciones:     payload.Observaciones,
	})
	if err != nil {
		http.Error(w, "No se pudo guardar configuracion: "+err.Error(), http.StatusBadRequest)
		return
	}

	cfg, err := dbpkg.GetSuperVentaDigitalConfig(dbSuper)
	if err != nil {
		http.Error(w, "Configuracion guardada, pero no se pudo consultar", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"config": cfg,
	})
}

func handleSuperVentaDigitalOrders(w http.ResponseWriter, r *http.Request, dbSuper *sql.DB) {
	limit, err := parseIntQueryOptional(r, "limit")
	if err != nil {
		http.Error(w, "limit invalido", http.StatusBadRequest)
		return
	}
	offset, err := parseIntQueryOptional(r, "offset")
	if err != nil {
		http.Error(w, "offset invalido", http.StatusBadRequest)
		return
	}
	rows, total, err := dbpkg.ListSuperVentaDigitalOrders(dbSuper, dbpkg.SuperVentaDigitalOrdersFilter{
		IncludeInactive: queryBool(r, "include_inactive"),
		EstadoPago:      strings.TrimSpace(r.URL.Query().Get("estado_pago")),
		Q:               strings.TrimSpace(r.URL.Query().Get("q")),
		Limit:           limit,
		Offset:          offset,
	})
	if err != nil {
		http.Error(w, "No se pudo consultar ordenes", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"total": total,
		"rows":  rows,
	})
}

func handleSuperVentaDigitalUploadImage(w http.ResponseWriter, r *http.Request, dbSuper *sql.DB) {
	if err := r.ParseMultipartForm(12 << 20); err != nil {
		http.Error(w, "invalid multipart payload", http.StatusBadRequest)
		return
	}

	itemID, _ := parseInt64Form(r, "item_id")
	file, header, err := r.FormFile("imagen")
	if err != nil {
		http.Error(w, "imagen required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true, ".svg": true}
	if !allowed[ext] {
		http.Error(w, "image extension not allowed", http.StatusBadRequest)
		return
	}

	webRoot := resolveWebRootDir()
	dir := filepath.Join(webRoot, "uploads", "venta_digital", "imagenes")
	// #nosec G301 -- imagen publica de venta digital servida por Nginx.
	if err := os.MkdirAll(dir, 0o755); err != nil {
		http.Error(w, "failed to prepare upload directory", http.StatusInternalServerError)
		return
	}

	prefix := "item"
	if itemID > 0 {
		prefix = fmt.Sprintf("item_%d", itemID)
	}
	fileName := fmt.Sprintf("%s_%d%s", prefix, time.Now().UnixNano(), ext)
	absPath := filepath.Join(dir, fileName)
	out, err := os.Create(absPath)
	if err != nil {
		http.Error(w, "failed to create image file", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, "failed to save image file", http.StatusInternalServerError)
		return
	}

	imageURL := "/uploads/venta_digital/imagenes/" + fileName
	if itemID > 0 {
		item, err := dbpkg.GetSuperVentaDigitalItemByID(dbSuper, itemID)
		if err == nil {
			item.ImagenURL = imageURL
			item.UsuarioCreador = adminEmailFromRequest(r)
			_ = dbpkg.UpdateSuperVentaDigitalItem(dbSuper, item)
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"saved":     true,
		"item_id":   itemID,
		"image_url": imageURL,
	})
}

func handleSuperVentaDigitalUploadInstructions(w http.ResponseWriter, r *http.Request, dbSuper *sql.DB) {
	if err := r.ParseMultipartForm(24 << 20); err != nil {
		http.Error(w, "invalid multipart payload", http.StatusBadRequest)
		return
	}

	itemID, _ := parseInt64Form(r, "item_id")
	file, header, err := r.FormFile("archivo")
	if err != nil {
		http.Error(w, "archivo required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{
		".txt": true, ".pdf": true, ".md": true, ".json": true,
		".zip": true, ".rar": true, ".7z": true,
		".doc": true, ".docx": true,
	}
	if !allowed[ext] {
		http.Error(w, "file extension not allowed", http.StatusBadRequest)
		return
	}

	webRoot := resolveWebRootDir()
	dir := filepath.Join(webRoot, "uploads", "venta_digital", "instrucciones")
	// #nosec G301 -- recurso publico descargable servido por Nginx.
	if err := os.MkdirAll(dir, 0o755); err != nil {
		http.Error(w, "failed to prepare upload directory", http.StatusInternalServerError)
		return
	}

	prefix := "instrucciones"
	if itemID > 0 {
		prefix = fmt.Sprintf("item_%d", itemID)
	}
	fileName := fmt.Sprintf("%s_%d%s", prefix, time.Now().UnixNano(), ext)
	absPath := filepath.Join(dir, fileName)
	out, err := os.Create(absPath)
	if err != nil {
		http.Error(w, "failed to create file", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}

	fileURL := "/uploads/venta_digital/instrucciones/" + fileName
	if itemID > 0 {
		item, err := dbpkg.GetSuperVentaDigitalItemByID(dbSuper, itemID)
		if err == nil {
			item.InstruccionesArchivoURL = fileURL
			item.UsuarioCreador = adminEmailFromRequest(r)
			_ = dbpkg.UpdateSuperVentaDigitalItem(dbSuper, item)
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"saved":                     true,
		"item_id":                   itemID,
		"instrucciones_archivo_url": fileURL,
	})
}

// PublicVentaDigitalHandler expone catalogo y pagos para clientes finales.
func PublicVentaDigitalHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		action := ventaDigitalNormalizeAction(r.URL.Query().Get("action"))
		if action == "" {
			http.Error(w, "action invalida", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "catalogo":
				handleVentaDigitalCatalogoPublico(w, r, dbSuper)
			case "estado_pago":
				handleVentaDigitalEstadoPagoPublico(w, r, dbSuper)
			default:
				http.Error(w, "action invalida", http.StatusBadRequest)
			}
		case http.MethodPost:
			switch action {
			case "crear_pago":
				handleVentaDigitalCrearPagoPublico(w, r, dbSuper)
			default:
				http.Error(w, "action invalida", http.StatusBadRequest)
			}
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func handleVentaDigitalCatalogoPublico(w http.ResponseWriter, _ *http.Request, dbSuper *sql.DB) {
	cfg, err := dbpkg.GetSuperVentaDigitalConfig(dbSuper)
	if err != nil {
		http.Error(w, "No se pudo cargar configuracion de tienda", http.StatusInternalServerError)
		return
	}
	items, err := dbpkg.ListSuperVentaDigitalItemsPublic(dbSuper)
	if err != nil {
		http.Error(w, "No se pudo cargar catalogo digital", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"tienda": sanitizeVentaDigitalConfigForPublic(cfg),
		"items":  sanitizeVentaDigitalItemsForPublic(items),
	})
}

func handleVentaDigitalCrearPagoPublico(w http.ResponseWriter, r *http.Request, dbSuper *sql.DB) {
	var payload ventaDigitalCrearPagoPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.ItemID <= 0 {
		http.Error(w, "item_id es obligatorio", http.StatusBadRequest)
		return
	}

	buyerName := strings.TrimSpace(payload.CompradorNombre)
	if buyerName == "" {
		http.Error(w, "comprador_nombre es obligatorio", http.StatusBadRequest)
		return
	}
	buyerEmail := strings.TrimSpace(payload.CompradorEmail)
	if buyerEmail == "" {
		http.Error(w, "comprador_email es obligatorio", http.StatusBadRequest)
		return
	}
	if _, err := mail.ParseAddress(buyerEmail); err != nil {
		http.Error(w, "comprador_email invalido", http.StatusBadRequest)
		return
	}
	buyerPhone := strings.TrimSpace(payload.CompradorTelefono)
	if ok, _ := regexp.MatchString(`^3\d{9}$`, buyerPhone); !ok {
		http.Error(w, "comprador_telefono invalido (10 digitos en CO para Nequi)", http.StatusBadRequest)
		return
	}

	item, err := dbpkg.GetSuperVentaDigitalItemByID(dbSuper, payload.ItemID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "item no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo validar item", http.StatusInternalServerError)
		return
	}
	if strings.EqualFold(strings.TrimSpace(item.Estado), "inactivo") {
		http.Error(w, "item inactivo", http.StatusBadRequest)
		return
	}
	if item.Precio <= 0 {
		http.Error(w, "item sin precio valido", http.StatusBadRequest)
		return
	}

	orderCode := fmt.Sprintf("VD-ORD-%d", time.Now().UnixNano())
	orderID, err := dbpkg.CreateSuperVentaDigitalOrder(dbSuper, dbpkg.SuperVentaDigitalOrder{
		CodigoOrden:             orderCode,
		ItemID:                  item.ID,
		ItemNombre:              item.Nombre,
		ItemPrecio:              item.Precio,
		ItemMoneda:              item.Moneda,
		CompradorNombre:         buyerName,
		CompradorEmail:          buyerEmail,
		CompradorTelefono:       buyerPhone,
		MetodoPago:              "wompi_nequi",
		EstadoPago:              "pendiente",
		LicenciaCodigoEnviado:   item.LicenciaCodigo,
		InstruccionesArchivoURL: item.InstruccionesArchivoURL,
		UsuarioCreador:          "publico",
		Estado:                  "activo",
	})
	if err != nil {
		http.Error(w, "No se pudo crear orden", http.StatusInternalServerError)
		return
	}

	cfg, err := dbpkg.GetSuperVentaDigitalConfig(dbSuper)
	if err != nil {
		http.Error(w, "No se pudo cargar configuracion de venta digital", http.StatusInternalServerError)
		return
	}
	if !cfg.WompiActivo {
		writeJSON(w, http.StatusPreconditionFailed, map[string]interface{}{
			"ok":         false,
			"order_id":   orderID,
			"order_code": orderCode,
			"error":      "wompi no esta activo para venta digital",
		})
		return
	}

	publicKey, err := getDecryptedConfigValue(dbSuper, "wompi.public_key")
	if err != nil {
		http.Error(w, "No se pudo leer wompi.public_key", http.StatusInternalServerError)
		return
	}
	privateKey, err := getDecryptedConfigValue(dbSuper, "wompi.private_key")
	if err != nil {
		http.Error(w, "No se pudo leer wompi.private_key", http.StatusInternalServerError)
		return
	}
	integrityKey, err := getDecryptedConfigValue(dbSuper, "wompi.integrity_key")
	if err != nil {
		http.Error(w, "No se pudo leer wompi.integrity_key", http.StatusInternalServerError)
		return
	}
	publicKey = strings.TrimSpace(publicKey)
	privateKey = strings.TrimSpace(privateKey)
	integrityKey = strings.TrimSpace(integrityKey)
	if publicKey == "" || privateKey == "" || integrityKey == "" {
		http.Error(w, "Wompi no configurado: faltan llaves (public/private/integrity)", http.StatusPreconditionFailed)
		return
	}

	mode, _ := resolveWompiMode(dbSuper, publicKey, privateKey)
	baseURL := wompiBaseURLFromMode(mode)
	acceptanceToken, personalToken, acceptancePermalink, personalPermalink, ferr := fetchWompiAcceptanceInfo(baseURL, publicKey)
	if ferr != nil {
		http.Error(w, "No se pudo consultar terminos de Wompi: "+ferr.Error(), http.StatusBadGateway)
		return
	}
	if !payload.AcceptTerms {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"ok":                      false,
			"error":                   "Debes aceptar terminos de Wompi para continuar",
			"order_id":                orderID,
			"order_code":              orderCode,
			"acceptance_permalink":    acceptancePermalink,
			"personal_data_permalink": personalPermalink,
		})
		return
	}
	if strings.TrimSpace(acceptanceToken) == "" || strings.TrimSpace(personalToken) == "" {
		http.Error(w, "Wompi no devolvio tokens de aceptacion", http.StatusBadGateway)
		return
	}

	amountInCents := int64(math.Round(item.Precio * 100))
	if amountInCents <= 0 {
		http.Error(w, "monto invalido para pago", http.StatusBadRequest)
		return
	}

	reference := dbpkg.BuildSuperVentaDigitalOrderReference(orderCode)
	signatureSource := fmt.Sprintf("%s%dCOP%s", reference, amountInCents, integrityKey)
	signatureHash := sha256.Sum256([]byte(signatureSource))
	signature := hex.EncodeToString(signatureHash[:])

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	redirectURL := fmt.Sprintf("%s://%s/venta_digital.html?order_code=%s&status=pending", scheme, r.Host, url.QueryEscape(orderCode))

	reqBody := map[string]interface{}{
		"acceptance_token":     acceptanceToken,
		"accept_personal_auth": personalToken,
		"amount_in_cents":      amountInCents,
		"currency":             "COP",
		"customer_email":       buyerEmail,
		"reference":            reference,
		"signature":            signature,
		"redirect_url":         redirectURL,
		"payment_method": map[string]interface{}{
			"type":         "NEQUI",
			"phone_number": buyerPhone,
		},
	}

	bodyBytes, _ := json.Marshal(reqBody)
	apiURL := strings.TrimRight(baseURL, "/") + "/transactions"
	request, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		http.Error(w, "No se pudo preparar solicitud Wompi", http.StatusInternalServerError)
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+privateKey)

	client := &http.Client{Timeout: 25 * time.Second}
	resp, err := client.Do(request)
	if err != nil {
		http.Error(w, "No se pudo crear transaccion Wompi: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		_ = dbpkg.UpdateSuperVentaDigitalOrderPayment(dbSuper, orderCode, "error", "", reference, string(respBody), "", "wompi_error")
		http.Error(w, "Wompi API error: "+string(respBody), http.StatusBadGateway)
		return
	}

	var wompiResp map[string]interface{}
	if err := json.Unmarshal(respBody, &wompiResp); err != nil {
		http.Error(w, "respuesta Wompi invalida", http.StatusInternalServerError)
		return
	}
	data, _ := wompiResp["data"].(map[string]interface{})
	transactionID := strings.TrimSpace(fmt.Sprint(data["id"]))
	statusWompi := strings.ToUpper(strings.TrimSpace(fmt.Sprint(data["status"])))
	if transactionID == "" || transactionID == "<nil>" {
		http.Error(w, "Wompi no devolvio transaction id", http.StatusBadGateway)
		return
	}
	if statusWompi == "" || statusWompi == "<nil>" {
		statusWompi = "PENDING"
	}

	_, delivered, deliveryStage, processErr := processVentaDigitalPaymentStatusUpdate(r, dbSuper, transactionID, reference, statusWompi, string(respBody))
	if processErr != nil {
		http.Error(w, "No se pudo actualizar orden: "+processErr.Error(), http.StatusInternalServerError)
		return
	}
	order, _ := dbpkg.GetSuperVentaDigitalOrderByCodigo(dbSuper, orderCode)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                      true,
		"order":                   sanitizeVentaDigitalOrderForPublic(order),
		"provider":                "wompi",
		"payment_method":          "NEQUI",
		"mode":                    mode,
		"transaction_id":          transactionID,
		"reference":               reference,
		"status":                  statusWompi,
		"status_local":            order.EstadoPago,
		"delivery_sent":           delivered,
		"delivery_stage":          deliveryStage,
		"acceptance_permalink":    acceptancePermalink,
		"personal_data_permalink": personalPermalink,
		"data":                    data,
	})
}

func handleVentaDigitalEstadoPagoPublico(w http.ResponseWriter, r *http.Request, dbSuper *sql.DB) {
	orderCode := strings.TrimSpace(r.URL.Query().Get("order_code"))
	if orderCode == "" {
		orderCode = strings.TrimSpace(r.URL.Query().Get("codigo_orden"))
	}
	if orderCode == "" {
		http.Error(w, "order_code es obligatorio", http.StatusBadRequest)
		return
	}

	order, err := dbpkg.GetSuperVentaDigitalOrderByCodigo(dbSuper, orderCode)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "orden no encontrada", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo consultar orden", http.StatusInternalServerError)
		return
	}

	cfg, err := dbpkg.GetSuperVentaDigitalConfig(dbSuper)
	if err != nil {
		http.Error(w, "No se pudo cargar configuracion", http.StatusInternalServerError)
		return
	}

	transactionID := strings.TrimSpace(r.URL.Query().Get("transaction_id"))
	if transactionID == "" {
		transactionID = strings.TrimSpace(order.TransactionID)
	}

	statusWompi := ""
	data := map[string]interface{}{}
	deliverySent := order.CorreoEntregado
	deliveryStage := "not_processed"

	if cfg.WompiActivo && transactionID != "" {
		publicKey, _ := getDecryptedConfigValue(dbSuper, "wompi.public_key")
		privateKey, _ := getDecryptedConfigValue(dbSuper, "wompi.private_key")
		publicKey = strings.TrimSpace(publicKey)
		privateKey = strings.TrimSpace(privateKey)
		if publicKey != "" {
			mode := "sandbox"
			if resolvedMode, _ := resolveWompiMode(dbSuper, publicKey, privateKey); resolvedMode != "" {
				mode = resolvedMode
			}
			baseURL := wompiBaseURLFromMode(mode)
			statusURL := strings.TrimRight(baseURL, "/") + "/transactions/" + url.PathEscape(transactionID)

			fetchStatus := func(authKey string) ([]byte, int, error) {
				req, reqErr := http.NewRequest(http.MethodGet, statusURL, nil)
				if reqErr != nil {
					return nil, 0, reqErr
				}
				req.Header.Set("Authorization", "Bearer "+authKey)
				client := &http.Client{Timeout: 20 * time.Second}
				resp, callErr := client.Do(req)
				if callErr != nil {
					return nil, 0, callErr
				}
				defer resp.Body.Close()
				body, _ := io.ReadAll(resp.Body)
				return body, resp.StatusCode, nil
			}

			respBody, statusCode, ferr := fetchStatus(publicKey)
			if ferr == nil {
				if statusCode >= 400 && privateKey != "" {
					if body2, code2, err2 := fetchStatus(privateKey); err2 == nil {
						respBody = body2
						statusCode = code2
					}
				}
				if statusCode < 400 {
					var wompiResp map[string]interface{}
					if err := json.Unmarshal(respBody, &wompiResp); err == nil {
						data, _ = wompiResp["data"].(map[string]interface{})
						statusWompi = strings.ToUpper(strings.TrimSpace(fmt.Sprint(data["status"])))
						if statusWompi == "" || statusWompi == "<nil>" {
							statusWompi = "PENDING"
						}
						_, deliverySent, deliveryStage, _ = processVentaDigitalPaymentStatusUpdate(r, dbSuper, transactionID, order.ReferenciaExterna, statusWompi, string(respBody))
						order, _ = dbpkg.GetSuperVentaDigitalOrderByCodigo(dbSuper, orderCode)
					}
				}
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":             true,
		"order":          sanitizeVentaDigitalOrderForPublic(order),
		"status":         statusWompi,
		"status_local":   order.EstadoPago,
		"delivery_sent":  deliverySent,
		"delivery_stage": deliveryStage,
		"data":           data,
	})
}
