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
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

// AdminCreatePaginaHandler crea una pagina publica para la empresa (requiere autenticacion y scope de empresa)
func AdminCreatePaginaHandler(db *sql.DB, webDir string) http.HandlerFunc {
	type req struct {
		EmpresaID   int64  `json:"empresa_id"`
		Slug        string `json:"slug"`
		Titulo      string `json:"titulo"`
		Descripcion string `json:"descripcion"`
		VideoURL    string `json:"video_url"`
		Activo      bool   `json:"activo"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var payload req
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		slug := strings.TrimSpace(payload.Slug)
		if slug == "" || payload.EmpresaID <= 0 || strings.TrimSpace(payload.Titulo) == "" {
			http.Error(w, "missing required fields", http.StatusBadRequest)
			return
		}
		id, err := dbpkg.CreatePaginaPublica(db, payload.EmpresaID, slug, payload.Titulo, payload.Descripcion, payload.VideoURL, payload.Activo)
		if err != nil {
			http.Error(w, fmt.Sprintf("db error: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
	}
}

// AdminCreateProductoHandler crea un producto en una pagina publica
func AdminCreateProductoHandler(db *sql.DB, webDir string) http.HandlerFunc {
	type req struct {
		PaginaID    int64  `json:"pagina_id"`
		Nombre      string `json:"nombre"`
		Descripcion string `json:"descripcion"`
		PrecioCents int64  `json:"precio_cents"`
		Moneda      string `json:"moneda"`
		Stock       *int   `json:"stock"`
		SKU         string `json:"sku"`
		YoutubeURL  string `json:"youtube_url"`
		Activo      bool   `json:"activo"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var payload req
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		if payload.PaginaID <= 0 || strings.TrimSpace(payload.Nombre) == "" {
			http.Error(w, "missing required fields", http.StatusBadRequest)
			return
		}
		stock := sql.NullInt64{}
		if payload.Stock != nil {
			stock.Int64 = int64(*payload.Stock)
			stock.Valid = true
		}
		id, err := dbpkg.CreateProductoPublico(db, payload.PaginaID, payload.Nombre, payload.Descripcion, payload.PrecioCents, payload.Moneda, stock, payload.SKU, payload.YoutubeURL, payload.Activo)
		if err != nil {
			http.Error(w, fmt.Sprintf("db error: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
	}
}

// UploadProductImageHandler sube una imagen para un producto y la guarda en web dir bajo empresa/productos
func UploadProductImageHandler(db *sql.DB, webDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB
			http.Error(w, "invalid multipart form", http.StatusBadRequest)
			return
		}
		productoIDStr := r.FormValue("producto_id")
		empresaIDStr := r.FormValue("empresa_id")
		var empresaID int64
		if empresaIDStr != "" {
			v, err := strconv.ParseInt(empresaIDStr, 10, 64)
			if err == nil {
				empresaID = v
			}
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "file required", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Ensure path exists: web/empresa_publica/<empresa_id>/productos/
		baseDir := filepath.Join(webDir, "empresa_publica")
		if empresaID > 0 {
			baseDir = filepath.Join(baseDir, strconv.FormatInt(empresaID, 10))
		} else {
			baseDir = filepath.Join(baseDir, "misc")
		}
		prodDir := filepath.Join(baseDir, "productos")
		if err := os.MkdirAll(prodDir, 0755); err != nil {
			http.Error(w, "failed create dir", http.StatusInternalServerError)
			return
		}

		// sanitize filename
		fname := filepath.Base(header.Filename)
		dstPath := filepath.Join(prodDir, fname)
		out, err := os.Create(dstPath)
		if err != nil {
			http.Error(w, "failed to create file", http.StatusInternalServerError)
			return
		}
		defer out.Close()
		if _, err := io.Copy(out, file); err != nil {
			http.Error(w, "failed to write file", http.StatusInternalServerError)
			return
		}

		// Save DB reference if producto_id provided
		if productoIDStr != "" {
			if pid, err := strconv.ParseInt(productoIDStr, 10, 64); err == nil {
				rel := strings.TrimPrefix(dstPath, webDir)
				rel = strings.ReplaceAll(rel, "\\", "/")
				if !strings.HasPrefix(rel, "/") {
					rel = "/" + rel
				}
				if err := dbpkg.AddImagenProductoPublico(db, pid, rel, 0); err != nil {
					// Non-fatal: log
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"url": strings.ReplaceAll(strings.TrimPrefix(dstPath, webDir), "\\", "/")})
	}
}

// PublicListPaginaHandler lista paginas publicas por empresa o por slug
func PublicListPaginaHandler(db *sql.DB, webDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaIDStr := r.URL.Query().Get("empresa_id")
		slug := r.URL.Query().Get("slug")
		if empresaIDStr == "" && slug == "" {
			http.Error(w, "empresa_id or slug required", http.StatusBadRequest)
			return
		}
		var resp interface{}
		var err error
		if slug != "" {
			resp, err = dbpkg.GetPaginaPublicaBySlug(db, slug)
		} else {
			eid, _ := strconv.ParseInt(empresaIDStr, 10, 64)
			resp, err = dbpkg.ListPaginasPublicasByEmpresa(db, eid)
		}
		if err != nil {
			if err == sql.ErrNoRows {
				http.NotFound(w, r)
				return
			}
			http.Error(w, fmt.Sprintf("db error: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

type empresaVentaPublicaConfigPayload struct {
	EmpresaID           int64  `json:"empresa_id"`
	EmpresaSlug         string `json:"empresa_slug"`
	NombreTienda        string `json:"nombre_tienda"`
	DescripcionTienda   string `json:"descripcion_tienda"`
	LogoURL             string `json:"logo_url"`
	BannerURL           string `json:"banner_url"`
	ColorPrimario       string `json:"color_primario"`
	TemaVisual          string `json:"tema_visual"`
	Moneda              string `json:"moneda"`
	DominioPublico      string `json:"dominio_publico"`
	MostrarStock        *bool  `json:"mostrar_stock"`
	PedidosRestauranteActivo        *bool  `json:"pedidos_restaurante_activo"`
	PedidosRegistroOpcionalCliente  *bool  `json:"pedidos_registro_opcional_cliente"`
	PedidosPermitirRecogerEnTienda  *bool  `json:"pedidos_permitir_recoger_en_tienda"`
	PedidosPermitirDomicilio        *bool  `json:"pedidos_permitir_domicilio"`
	PedidosTrackingDomiciliario     *bool  `json:"pedidos_tracking_domiciliario"`
	PedidosDespachoAutomatico       *bool  `json:"pedidos_despacho_automatico"`
	PedidosNombreSistema            string `json:"pedidos_nombre_sistema"`
	PedidosTiempoPreparacionMinutos int    `json:"pedidos_tiempo_preparacion_minutos"`
	WompiActivo         *bool  `json:"wompi_activo"`
	WompiMode           string `json:"wompi_mode"`
	WompiPublicKey      string `json:"wompi_public_key"`
	WompiPrivateKeyRef  string `json:"wompi_private_key_ref"`
	WompiIntegrityRef   string `json:"wompi_integrity_key_ref"`
	WompiEventKeyRef    string `json:"wompi_event_key_ref"`
	EpaycoActivo        *bool  `json:"epayco_activo"`
	EpaycoMode          string `json:"epayco_mode"`
	EpaycoPublicKey     string `json:"epayco_public_key"`
	EpaycoPrivateKeyRef string `json:"epayco_private_key_ref"`
	EpaycoCustomerID    string `json:"epayco_customer_id"`
	Observaciones       string `json:"observaciones"`
}

type empresaVentaPublicaItemPayload struct {
	EmpresaID      int64   `json:"empresa_id"`
	ID             int64   `json:"id"`
	PaginaID       int64   `json:"pagina_id"`
	ProductoID     int64   `json:"producto_id"`
	CodigoPublico  string  `json:"codigo_publico"`
	Nombre         string  `json:"nombre"`
	Descripcion    string  `json:"descripcion"`
	Precio         float64 `json:"precio"`
	Moneda         string  `json:"moneda"`
	ImagenURL      string  `json:"imagen_url"`
	StockPublicado float64 `json:"stock_publicado"`
	OrdenVisual    int     `json:"orden_visual"`
	Destacado      bool    `json:"destacado"`
	Observaciones  string  `json:"observaciones"`
}

type empresaVentaPublicaPaginaPayload struct {
	EmpresaID     int64  `json:"empresa_id"`
	ID            int64  `json:"id"`
	Slug          string `json:"slug"`
	Nombre        string `json:"nombre"`
	Descripcion   string `json:"descripcion"`
	BannerURL     string `json:"banner_url"`
	OrdenVisual   int    `json:"orden_visual"`
	Estado        string `json:"estado"`
	Observaciones string `json:"observaciones"`
}

type ventaPublicaPagoItemPayload struct {
	ItemID    int64   `json:"item_id"`
	Cantidad  float64 `json:"cantidad"`
	Nombre    string  `json:"nombre,omitempty"`
	Precio    float64 `json:"precio,omitempty"`
	Subtotal  float64 `json:"subtotal,omitempty"`
	ImagenURL string  `json:"imagen_url,omitempty"`
}

type ventaPublicaCrearPagoPayload struct {
	EmpresaID         int64                         `json:"empresa_id"`
	EmpresaSlug       string                        `json:"empresa_slug"`
	MetodoPago        string                        `json:"metodo_pago"`
	CompradorNombre   string                        `json:"comprador_nombre"`
	CompradorEmail    string                        `json:"comprador_email"`
	CompradorTelefono string                        `json:"comprador_telefono"`
	AcceptTerms       bool                          `json:"accept_terms"`
	Items             []ventaPublicaPagoItemPayload `json:"items"`
}

type ventaPublicaCrearPedidoPayload struct {
	EmpresaID                    int64                         `json:"empresa_id"`
	EmpresaSlug                  string                        `json:"empresa_slug"`
	CompradorNombre              string                        `json:"comprador_nombre"`
	CompradorTelefono            string                        `json:"comprador_telefono"`
	CompradorEmail               string                        `json:"comprador_email"`
	CanalEntrega                 string                        `json:"canal_entrega"`
	DireccionEntrega             string                        `json:"direccion_entrega"`
	NotasEntrega                 string                        `json:"notas_entrega"`
	ClienteComparteUbicacion     bool                          `json:"cliente_comparte_ubicacion"`
	EntregaLatitud               float64                       `json:"entrega_latitud"`
	EntregaLongitud              float64                       `json:"entrega_longitud"`
	Items                        []ventaPublicaPagoItemPayload `json:"items"`
}

func empresaVentaPublicaNormalizeAction(raw string) string {
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
	case "paginas", "pages":
		return "paginas"
	case "ordenes", "orders":
		return "ordenes"
	case "pedido_estado", "order_state", "estado_pedido":
		return "pedido_estado"
	case "subir_imagen", "upload_image", "imagen":
		return "subir_imagen"
	default:
		return ""
	}
}

func ventaPublicaNormalizeAction(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "catalogo", "catalog", "tienda", "store":
		return "catalogo"
	case "crear_pago", "pagar", "create_payment", "payment":
		return "crear_pago"
	case "estado_pago", "status", "payment_status":
		return "estado_pago"
	case "crear_pedido", "pedido", "restaurant_order":
		return "crear_pedido"
	case "estado_pedido", "order_status", "tracking":
		return "estado_pedido"
	default:
		return ""
	}
}

func ventaPublicaResolveCredential(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	resolved, err := resolveDIANSecretValue(raw)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(resolved), nil
}

func ventaPublicaMapWompiStatus(raw string) string {
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

func ventaPublicaMapEpaycoStatus(raw string) string {
	status := strings.ToUpper(strings.TrimSpace(raw))
	switch status {
	case "APPROVED", "ACCEPTED", "SUCCESS":
		return "aprobado"
	case "DECLINED", "VOIDED", "ERROR", "FAILED", "REJECTED", "CANCELLED":
		return "rechazado"
	default:
		return "pendiente"
	}
}

func ventaPublicaNormalizeMetodoPago(raw string) string {
	method := strings.ToLower(strings.TrimSpace(raw))
	switch method {
	case "", "wompi", "wompi_nequi", "nequi":
		return "wompi_nequi"
	case "epayco", "epayco_checkout", "smart_checkout":
		return "epayco"
	default:
		return ""
	}
}

func ventaPublicaProviderFromMetodo(raw string) string {
	method := ventaPublicaNormalizeMetodoPago(raw)
	if method == "epayco" {
		return "epayco"
	}
	return "wompi"
}

func ventaPublicaResolveBaseURL(r *http.Request) string {
	proto := strings.TrimSpace(ventaPublicaFirstHeaderValue(r.Header.Get("X-Forwarded-Proto")))
	if proto == "" {
		if r.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}
	host := strings.TrimSpace(ventaPublicaFirstHeaderValue(r.Header.Get("X-Forwarded-Host")))
	if host == "" {
		host = strings.TrimSpace(r.Host)
	}
	if host == "" {
		host = "localhost"
	}
	return proto + "://" + host
}

func ventaPublicaResolveOrderContextFromReference(reference string) (int64, string) {
	reference = strings.TrimSpace(reference)
	if reference == "" {
		return 0, ""
	}
	parts := strings.Split(reference, "|")
	if len(parts) < 3 {
		return 0, dbpkg.TryParseOrderCodeFromReference(reference)
	}
	if strings.TrimSpace(parts[0]) != "VP" {
		return 0, dbpkg.TryParseOrderCodeFromReference(reference)
	}
	empresaID, _ := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
	return empresaID, strings.TrimSpace(parts[2])
}

func processVentaPublicaPaymentStatusUpdate(dbEmp *sql.DB, provider, transactionID, reference, status, rawPayload string) (bool, error) {
	if dbEmp == nil {
		return false, nil
	}
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		provider = "wompi"
	}

	resolvedEmpresaID, orderCode := ventaPublicaResolveOrderContextFromReference(reference)
	var order dbpkg.EmpresaVentaPublicaOrder
	var err error
	if resolvedEmpresaID > 0 && orderCode != "" {
		order, err = dbpkg.GetEmpresaVentaPublicaOrderByCodigo(dbEmp, resolvedEmpresaID, orderCode)
	} else {
		order, err = dbpkg.FindEmpresaVentaPublicaOrderByTransactionOrReference(dbEmp, transactionID, reference)
		if err == nil {
			resolvedEmpresaID = order.EmpresaID
			orderCode = order.CodigoOrden
		}
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	statusLocal := ventaPublicaMapWompiStatus(status)
	if provider == "epayco" {
		statusLocal = ventaPublicaMapEpaycoStatus(status)
	}
	pagadoEn := ""
	if statusLocal == "aprobado" {
		pagadoEn = time.Now().In(time.Local).Format("2006-01-02 15:04:05")
	}
	referenciaToSave := strings.TrimSpace(order.ReferenciaExterna)
	if referenciaToSave == "" {
		referenciaToSave = strings.TrimSpace(reference)
	}
	transactionToSave := strings.TrimSpace(transactionID)
	if transactionToSave == "" {
		transactionToSave = strings.TrimSpace(order.TransactionID)
	}
	if err := dbpkg.UpdateEmpresaVentaPublicaOrderPayment(dbEmp, resolvedEmpresaID, orderCode, statusLocal, transactionToSave, referenciaToSave, rawPayload, pagadoEn, provider+"_webhook"); err != nil {
		return true, err
	}
	return true, nil
}

func ventaPublicaSlugFromRequest(r *http.Request) string {
	raw := strings.TrimSpace(r.URL.Query().Get("empresa_slug"))
	if raw != "" {
		return dbpkg.NormalizeEmpresaPublicSlug(raw)
	}
	path := strings.TrimSpace(r.URL.Path)
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 2 && strings.EqualFold(parts[len(parts)-1], "venta_publica.html") {
		candidate := strings.TrimSpace(parts[len(parts)-2])
		if candidate != "" {
			return dbpkg.NormalizeEmpresaPublicSlug(candidate)
		}
	}
	return ResolveVentaPublicaSlugFromHost(r)
}

func ventaPublicaFirstHeaderValue(raw string) string {
	parts := strings.Split(raw, ",")
	if len(parts) == 0 {
		return strings.TrimSpace(raw)
	}
	return strings.TrimSpace(parts[0])
}

func ventaPublicaHostWithoutPort(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	host, _, err := net.SplitHostPort(trimmed)
	if err == nil {
		return strings.Trim(host, "[]")
	}
	return strings.Trim(trimmed, "[]")
}

func ventaPublicaNormalizeBaseDomain(raw string) string {
	candidate := strings.ToLower(strings.TrimSpace(raw))
	candidate = strings.TrimPrefix(candidate, "http://")
	candidate = strings.TrimPrefix(candidate, "https://")
	if idx := strings.Index(candidate, "/"); idx >= 0 {
		candidate = candidate[:idx]
	}
	candidate = ventaPublicaHostWithoutPort(candidate)
	candidate = strings.Trim(candidate, ".")
	return candidate
}

func ventaPublicaBaseDomains() []string {
	seen := map[string]bool{}
	out := make([]string, 0, 4)
	raw := strings.TrimSpace(os.Getenv("VENTA_PUBLICA_BASE_DOMAINS"))
	for _, part := range strings.Split(raw, ",") {
		base := ventaPublicaNormalizeBaseDomain(part)
		if base == "" || seen[base] {
			continue
		}
		seen[base] = true
		out = append(out, base)
	}
	if len(out) == 0 {
		out = append(out, "powerfulcontrolsystem.com")
	}
	return out
}

// ResolveVentaPublicaSlugFromHost resuelve slug desde Host/X-Forwarded-Host en subdominios tipo empresa1.midominio.com.
func ResolveVentaPublicaSlugFromHost(r *http.Request) string {
	rawHost := ventaPublicaFirstHeaderValue(r.Header.Get("X-Forwarded-Host"))
	if rawHost == "" {
		rawHost = strings.TrimSpace(r.Host)
	}
	host := strings.ToLower(strings.Trim(ventaPublicaHostWithoutPort(rawHost), "."))
	if host == "" {
		return ""
	}
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return ""
	}
	if net.ParseIP(host) != nil {
		return ""
	}

	for _, baseDomain := range ventaPublicaBaseDomains() {
		if host == baseDomain || host == "www."+baseDomain {
			continue
		}
		suffix := "." + baseDomain
		if !strings.HasSuffix(host, suffix) {
			continue
		}
		label := strings.Trim(strings.TrimSuffix(host, suffix), ".")
		if label == "" || strings.Contains(label, ".") {
			continue
		}
		if ok, _ := regexp.MatchString(`^[a-z0-9-]+$`, label); !ok {
			continue
		}
		normalized := dbpkg.NormalizeEmpresaPublicSlug(label)
		if normalized != "" {
			return normalized
		}
	}

	return ""
}

func sanitizeVentaPublicaConfigForPublic(cfg dbpkg.EmpresaVentaPublicaConfig) map[string]interface{} {
	paymentMethods := make([]string, 0, 2)
	if cfg.WompiActivo {
		paymentMethods = append(paymentMethods, "wompi_nequi")
	}
	if cfg.EpaycoActivo {
		paymentMethods = append(paymentMethods, "epayco")
	}
	return map[string]interface{}{
		"empresa_id":         cfg.EmpresaID,
		"empresa_slug":       cfg.EmpresaSlug,
		"nombre_tienda":      cfg.NombreTienda,
		"descripcion_tienda": cfg.DescripcionTienda,
		"logo_url":           cfg.LogoURL,
		"banner_url":         cfg.BannerURL,
		"color_primario":     cfg.ColorPrimario,
		"tema_visual":        cfg.TemaVisual,
		"moneda":             cfg.Moneda,
		"dominio_publico":    cfg.DominioPublico,
		"mostrar_stock":      cfg.MostrarStock,
		"pedidos_restaurante_activo": cfg.PedidosRestauranteActivo,
		"pedidos_registro_opcional_cliente": cfg.PedidosRegistroOpcionalCliente,
		"pedidos_permitir_recoger_en_tienda": cfg.PedidosPermitirRecogerEnTienda,
		"pedidos_permitir_domicilio": cfg.PedidosPermitirDomicilio,
		"pedidos_tracking_domiciliario": cfg.PedidosTrackingDomiciliario,
		"pedidos_despacho_automatico": cfg.PedidosDespachoAutomatico,
		"pedidos_nombre_sistema": cfg.PedidosNombreSistema,
		"pedidos_tiempo_preparacion_minutos": cfg.PedidosTiempoPreparacionMinutos,
		"wompi_activo":       cfg.WompiActivo,
		"wompi_mode":         cfg.WompiMode,
		"epayco_activo":      cfg.EpaycoActivo,
		"epayco_mode":        cfg.EpaycoMode,
		"payment_methods":    paymentMethods,
	}
}

func validateVentaPublicaSecureRefIfProvided(raw string, field string) error {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil
	}
	if _, err := validateIntegracionCredentialReference(value); err != nil {
		return fmt.Errorf("%s invalido: %w", field, err)
	}
	return nil
}

// EmpresaVentaPublicaHandler gestiona catalogo/configuracion de venta publica por empresa.
func EmpresaVentaPublicaHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		action := empresaVentaPublicaNormalizeAction(r.URL.Query().Get("action"))
		if action == "" {
			http.Error(w, "action invalida", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "catalogo":
				handleEmpresaVentaPublicaList(w, r, dbEmp)
			case "detalle":
				handleEmpresaVentaPublicaDetail(w, r, dbEmp)
			case "config":
				handleEmpresaVentaPublicaConfigGet(w, r, dbEmp)
			case "paginas":
				handleEmpresaVentaPublicaPages(w, r, dbEmp)
			case "ordenes":
				handleEmpresaVentaPublicaOrders(w, r, dbEmp)
			default:
				http.Error(w, "action invalida", http.StatusBadRequest)
			}
		case http.MethodPost:
			switch action {
			case "crear", "catalogo":
				handleEmpresaVentaPublicaCreate(w, r, dbEmp)
			case "config":
				handleEmpresaVentaPublicaConfigUpsert(w, r, dbEmp)
			case "paginas":
				handleEmpresaVentaPublicaPageUpsert(w, r, dbEmp)
			case "subir_imagen":
				handleEmpresaVentaPublicaUploadImage(w, r, dbEmp)
			case "pedido_estado":
				handleEmpresaVentaPublicaOrderState(w, r, dbEmp)
			default:
				http.Error(w, "action invalida", http.StatusBadRequest)
			}
		case http.MethodPut, http.MethodPatch:
			switch action {
			case "actualizar", "catalogo":
				handleEmpresaVentaPublicaUpdate(w, r, dbEmp)
			case "activar", "desactivar":
				handleEmpresaVentaPublicaToggle(w, r, dbEmp, action)
			case "config":
				handleEmpresaVentaPublicaConfigUpsert(w, r, dbEmp)
			case "paginas":
				handleEmpresaVentaPublicaPageUpsert(w, r, dbEmp)
			default:
				http.Error(w, "action invalida", http.StatusBadRequest)
			}
		case http.MethodDelete:
			if action == "paginas" {
				handleEmpresaVentaPublicaPageToggle(w, r, dbEmp, "desactivar")
			} else {
				handleEmpresaVentaPublicaToggle(w, r, dbEmp, "desactivar")
			}
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func handleEmpresaVentaPublicaPages(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rows, err := dbpkg.ListEmpresaVentaPublicaPaginas(dbEmp, empresaID, queryBool(r, "include_inactive"))
	if err != nil {
		http.Error(w, "No se pudieron listar paginas de venta publica", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"rows":       rows,
	})
}

func handleEmpresaVentaPublicaPageUpsert(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var payload empresaVentaPublicaPaginaPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID > 0 && payload.EmpresaID != empresaID {
		http.Error(w, "empresa_id no coincide con contexto", http.StatusBadRequest)
		return
	}
	if payload.ID <= 0 {
		if qID, qErr := parseInt64QueryOptional(r, "id"); qErr == nil && qID > 0 {
			payload.ID = qID
		}
	}
	id, err := dbpkg.UpsertEmpresaVentaPublicaPagina(dbEmp, dbpkg.EmpresaVentaPublicaPagina{
		ID:             payload.ID,
		EmpresaID:      empresaID,
		Slug:           payload.Slug,
		Nombre:         payload.Nombre,
		Descripcion:    payload.Descripcion,
		BannerURL:      payload.BannerURL,
		OrdenVisual:    payload.OrdenVisual,
		UsuarioCreador: adminEmailFromRequest(r),
		Estado:         firstNonEmptyString(payload.Estado, "activo"),
		Observaciones:  payload.Observaciones,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	page, err := dbpkg.GetEmpresaVentaPublicaPaginaByID(dbEmp, empresaID, id)
	if err != nil {
		http.Error(w, "pagina guardada pero no se pudo consultar", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"page":       page,
	})
}

func handleEmpresaVentaPublicaPageToggle(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, action string) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pageID, err := parseInt64QueryOptional(r, "id")
	if err != nil || pageID <= 0 {
		http.Error(w, "id es obligatorio", http.StatusBadRequest)
		return
	}
	estado := "activo"
	if action == "desactivar" {
		estado = "inactivo"
	}
	if err := dbpkg.SetEmpresaVentaPublicaPaginaEstadoByID(dbEmp, empresaID, pageID, estado); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "pagina no encontrada", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo actualizar pagina", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "empresa_id": empresaID, "id": pageID, "estado": estado})
}

func handleEmpresaVentaPublicaList(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
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
	paginaID, err := parseInt64QueryOptional(r, "pagina_id")
	if err != nil {
		http.Error(w, "pagina_id invalido", http.StatusBadRequest)
		return
	}

	rows, total, err := dbpkg.ListEmpresaVentaPublicaItems(dbEmp, empresaID, dbpkg.EmpresaVentaPublicaItemsFilter{
		IncludeInactive: queryBool(r, "include_inactive"),
		PaginaID:        paginaID,
		PaginaSlug:      strings.TrimSpace(r.URL.Query().Get("pagina_slug")),
		Q:               strings.TrimSpace(r.URL.Query().Get("q")),
		Limit:           limit,
		Offset:          offset,
	})
	if err != nil {
		http.Error(w, "No se pudo listar catalogo publico", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"total":      total,
		"rows":       rows,
	})
}

func handleEmpresaVentaPublicaDetail(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	itemID, err := parseInt64QueryOptional(r, "id")
	if err != nil || itemID <= 0 {
		http.Error(w, "id es obligatorio", http.StatusBadRequest)
		return
	}
	item, err := dbpkg.GetEmpresaVentaPublicaItemByID(dbEmp, empresaID, itemID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "item no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo consultar item", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"item":       item,
	})
}

func handleEmpresaVentaPublicaCreate(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var payload empresaVentaPublicaItemPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID > 0 && payload.EmpresaID != empresaID {
		http.Error(w, "empresa_id no coincide con contexto", http.StatusBadRequest)
		return
	}
	id, err := dbpkg.CreateEmpresaVentaPublicaItem(dbEmp, dbpkg.EmpresaVentaPublicaItem{
		EmpresaID:      empresaID,
		PaginaID:       payload.PaginaID,
		ProductoID:     payload.ProductoID,
		CodigoPublico:  payload.CodigoPublico,
		Nombre:         payload.Nombre,
		Descripcion:    payload.Descripcion,
		Precio:         payload.Precio,
		Moneda:         payload.Moneda,
		ImagenURL:      payload.ImagenURL,
		StockPublicado: payload.StockPublicado,
		OrdenVisual:    payload.OrdenVisual,
		Destacado:      payload.Destacado,
		UsuarioCreador: adminEmailFromRequest(r),
		Estado:         "activo",
		Observaciones:  payload.Observaciones,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	item, err := dbpkg.GetEmpresaVentaPublicaItemByID(dbEmp, empresaID, id)
	if err != nil {
		http.Error(w, "item creado pero no se pudo consultar", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"item":       item,
	})
}

func handleEmpresaVentaPublicaUpdate(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var payload empresaVentaPublicaItemPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID > 0 && payload.EmpresaID != empresaID {
		http.Error(w, "empresa_id no coincide con contexto", http.StatusBadRequest)
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

	err = dbpkg.UpdateEmpresaVentaPublicaItem(dbEmp, dbpkg.EmpresaVentaPublicaItem{
		ID:             payload.ID,
		EmpresaID:      empresaID,
		PaginaID:       payload.PaginaID,
		ProductoID:     payload.ProductoID,
		CodigoPublico:  payload.CodigoPublico,
		Nombre:         payload.Nombre,
		Descripcion:    payload.Descripcion,
		Precio:         payload.Precio,
		Moneda:         payload.Moneda,
		ImagenURL:      payload.ImagenURL,
		StockPublicado: payload.StockPublicado,
		OrdenVisual:    payload.OrdenVisual,
		Destacado:      payload.Destacado,
		UsuarioCreador: adminEmailFromRequest(r),
		Estado:         "activo",
		Observaciones:  payload.Observaciones,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "item no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	item, err := dbpkg.GetEmpresaVentaPublicaItemByID(dbEmp, empresaID, payload.ID)
	if err != nil {
		http.Error(w, "item actualizado pero no se pudo consultar", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"item":       item,
	})
}

func handleEmpresaVentaPublicaToggle(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, action string) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	itemID, err := parseInt64QueryOptional(r, "id")
	if err != nil || itemID <= 0 {
		http.Error(w, "id es obligatorio", http.StatusBadRequest)
		return
	}
	estado := "activo"
	if action == "desactivar" {
		estado = "inactivo"
	}
	if err := dbpkg.SetEmpresaVentaPublicaItemEstadoByID(dbEmp, empresaID, itemID, estado); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "item no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo actualizar estado", http.StatusInternalServerError)
		return
	}
	item, _ := dbpkg.GetEmpresaVentaPublicaItemByID(dbEmp, empresaID, itemID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"item":       item,
	})
}

func handleEmpresaVentaPublicaConfigGet(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cfg, err := dbpkg.GetEmpresaVentaPublicaConfig(dbEmp, empresaID)
	if err != nil {
		http.Error(w, "No se pudo consultar configuracion", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":          true,
		"empresa_id":  empresaID,
		"config":      cfg,
		"public_path": "/" + cfg.EmpresaSlug + "/venta_publica.html",
	})
}

func handleEmpresaVentaPublicaConfigUpsert(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var payload empresaVentaPublicaConfigPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID > 0 && payload.EmpresaID != empresaID {
		http.Error(w, "empresa_id no coincide con contexto", http.StatusBadRequest)
		return
	}
	if err := validateVentaPublicaSecureRefIfProvided(payload.WompiPrivateKeyRef, "wompi_private_key_ref"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateVentaPublicaSecureRefIfProvided(payload.WompiIntegrityRef, "wompi_integrity_key_ref"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateVentaPublicaSecureRefIfProvided(payload.WompiEventKeyRef, "wompi_event_key_ref"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateVentaPublicaSecureRefIfProvided(payload.EpaycoPrivateKeyRef, "epayco_private_key_ref"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mostrarStock := true
	if payload.MostrarStock != nil {
		mostrarStock = *payload.MostrarStock
	}
	pedidosRestauranteActivo := false
	if payload.PedidosRestauranteActivo != nil {
		pedidosRestauranteActivo = *payload.PedidosRestauranteActivo
	}
	pedidosRegistroOpcional := true
	if payload.PedidosRegistroOpcionalCliente != nil {
		pedidosRegistroOpcional = *payload.PedidosRegistroOpcionalCliente
	}
	pedidosPermitirRecoger := true
	if payload.PedidosPermitirRecogerEnTienda != nil {
		pedidosPermitirRecoger = *payload.PedidosPermitirRecogerEnTienda
	}
	pedidosPermitirDomicilio := true
	if payload.PedidosPermitirDomicilio != nil {
		pedidosPermitirDomicilio = *payload.PedidosPermitirDomicilio
	}
	pedidosTracking := true
	if payload.PedidosTrackingDomiciliario != nil {
		pedidosTracking = *payload.PedidosTrackingDomiciliario
	}
	pedidosDespacho := true
	if payload.PedidosDespachoAutomatico != nil {
		pedidosDespacho = *payload.PedidosDespachoAutomatico
	}
	wompiActivo := false
	if payload.WompiActivo != nil {
		wompiActivo = *payload.WompiActivo
	}
	epaycoActivo := false
	if payload.EpaycoActivo != nil {
		epaycoActivo = *payload.EpaycoActivo
	}
	if wompiActivo {
		if strings.TrimSpace(payload.WompiPublicKey) == "" {
			http.Error(w, "wompi_public_key es obligatoria cuando wompi_activo=1", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(payload.WompiPrivateKeyRef) == "" || strings.TrimSpace(payload.WompiIntegrityRef) == "" {
			http.Error(w, "wompi_private_key_ref y wompi_integrity_key_ref son obligatorias cuando wompi_activo=1", http.StatusBadRequest)
			return
		}
	}
	if epaycoActivo {
		if strings.TrimSpace(payload.EpaycoPublicKey) == "" {
			http.Error(w, "epayco_public_key es obligatoria cuando epayco_activo=1", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(payload.EpaycoPrivateKeyRef) == "" {
			http.Error(w, "epayco_private_key_ref es obligatoria cuando epayco_activo=1", http.StatusBadRequest)
			return
		}
	}

	_, err = dbpkg.UpsertEmpresaVentaPublicaConfig(dbEmp, dbpkg.EmpresaVentaPublicaConfig{
		EmpresaID:           empresaID,
		EmpresaSlug:         payload.EmpresaSlug,
		NombreTienda:        payload.NombreTienda,
		DescripcionTienda:   payload.DescripcionTienda,
		LogoURL:             payload.LogoURL,
		BannerURL:           payload.BannerURL,
		ColorPrimario:       payload.ColorPrimario,
		TemaVisual:          payload.TemaVisual,
		Moneda:              payload.Moneda,
		DominioPublico:      payload.DominioPublico,
		MostrarStock:        mostrarStock,
		PedidosRestauranteActivo:        pedidosRestauranteActivo,
		PedidosRegistroOpcionalCliente:  pedidosRegistroOpcional,
		PedidosPermitirRecogerEnTienda:  pedidosPermitirRecoger,
		PedidosPermitirDomicilio:        pedidosPermitirDomicilio,
		PedidosTrackingDomiciliario:     pedidosTracking,
		PedidosDespachoAutomatico:       pedidosDespacho,
		PedidosNombreSistema:            payload.PedidosNombreSistema,
		PedidosTiempoPreparacionMinutos: payload.PedidosTiempoPreparacionMinutos,
		WompiActivo:         wompiActivo,
		WompiMode:           payload.WompiMode,
		WompiPublicKey:      payload.WompiPublicKey,
		WompiPrivateKeyRef:  payload.WompiPrivateKeyRef,
		WompiIntegrityRef:   payload.WompiIntegrityRef,
		WompiEventKeyRef:    payload.WompiEventKeyRef,
		EpaycoActivo:        epaycoActivo,
		EpaycoMode:          payload.EpaycoMode,
		EpaycoPublicKey:     payload.EpaycoPublicKey,
		EpaycoPrivateKeyRef: payload.EpaycoPrivateKeyRef,
		EpaycoCustomerID:    payload.EpaycoCustomerID,
		UsuarioCreador:      adminEmailFromRequest(r),
		Estado:              "activo",
		Observaciones:       payload.Observaciones,
	})
	if err != nil {
		http.Error(w, "No se pudo guardar configuracion: "+err.Error(), http.StatusBadRequest)
		return
	}

	cfg, err := dbpkg.GetEmpresaVentaPublicaConfig(dbEmp, empresaID)
	if err != nil {
		http.Error(w, "Configuracion guardada, pero no se pudo consultar", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":          true,
		"empresa_id":  empresaID,
		"config":      cfg,
		"public_path": "/" + cfg.EmpresaSlug + "/venta_publica.html",
	})
}

func handleEmpresaVentaPublicaOrders(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
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
	rows, total, err := dbpkg.ListEmpresaVentaPublicaOrders(dbEmp, empresaID, dbpkg.EmpresaVentaPublicaOrdersFilter{
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
		"ok":         true,
		"empresa_id": empresaID,
		"total":      total,
		"rows":       rows,
	})
}

func handleEmpresaVentaPublicaOrderState(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var payload struct {
		CodigoOrden   string `json:"codigo_orden"`
		EstadoPedido  string `json:"estado_pedido"`
		Observaciones string `json:"observaciones"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	order, err := dbpkg.GetEmpresaVentaPublicaOrderByCodigo(dbEmp, empresaID, payload.CodigoOrden)
	if err != nil {
		http.Error(w, "orden no encontrada", http.StatusNotFound)
		return
	}
	cfg, err := dbpkg.GetEmpresaVentaPublicaConfig(dbEmp, empresaID)
	if err != nil {
		http.Error(w, "No se pudo cargar configuracion", http.StatusInternalServerError)
		return
	}
	estadoPedido := payload.EstadoPedido
	taxiRequestID := order.TaxiRequestID
	trackingToken := strings.TrimSpace(order.TrackingToken)
	if order.TipoOrden == "pedido_restaurante" && estadoPedido == "entregado_al_mensajero" && order.CanalEntrega == "domicilio" && cfg.PedidosTrackingDomiciliario && cfg.PedidosDespachoAutomatico {
		if taxiRequestID <= 0 {
			if trackingToken == "" {
				sum := sha256.Sum256([]byte(fmt.Sprintf("%d|%s|%d", empresaID, order.CodigoOrden, time.Now().UnixNano())))
				trackingToken = hex.EncodeToString(sum[:])[:24]
			}
			req, err := dbpkg.CreateTaxiRequest(dbEmp, dbpkg.EmpresaTaxiRequest{
				EmpresaID:                empresaID,
				ClienteNombre:            firstNonEmptyString(order.CompradorNombre, "Cliente"),
				ClienteTelefono:          order.CompradorTelefono,
				RecogerTexto:             firstNonEmptyString(cfg.NombreTienda, "Restaurante"),
				RecogerLatitud:           0,
				RecogerLongitud:          0,
				DestinoTexto:             firstNonEmptyString(order.DireccionEntrega, "Entrega a domicilio"),
				DestinoLatitud:           order.EntregaLatitud,
				DestinoLongitud:          order.EntregaLongitud,
				ComparteUbicacionCliente: order.ClienteComparteUbicacion,
				MetodoSolicitud:          "pedido_restaurante",
				Canal:                    "venta_publica",
				Notas:                    "Entrega de pedido " + order.CodigoOrden,
			})
			if err == nil {
				taxiRequestID = req.ID
				_ = dbpkg.UpdateEmpresaVentaPublicaOrderTracking(dbEmp, empresaID, order.CodigoOrden, taxiRequestID, trackingToken)
			}
		}
	}
	if err := dbpkg.UpdateEmpresaVentaPublicaOrderOperationalState(dbEmp, empresaID, order.CodigoOrden, estadoPedido, payload.Observaciones, taxiRequestID); err != nil {
		http.Error(w, "No se pudo actualizar el pedido", http.StatusBadRequest)
		return
	}
	updated, err := dbpkg.GetEmpresaVentaPublicaOrderByCodigo(dbEmp, empresaID, order.CodigoOrden)
	if err != nil {
		http.Error(w, "pedido actualizado pero no se pudo consultar", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "order": updated})
}

func handleEmpresaVentaPublicaUploadImage(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	if err := r.ParseMultipartForm(12 << 20); err != nil {
		http.Error(w, "invalid multipart payload", http.StatusBadRequest)
		return
	}
	empresaID, err := parseInt64Form(r, "empresa_id")
	if err != nil || empresaID <= 0 {
		http.Error(w, "empresa_id required", http.StatusBadRequest)
		return
	}
	w.Header().Set("X-Empresa-ID", strconv.FormatInt(empresaID, 10))

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
	dir := filepath.Join(webRoot, "uploads", "venta_publica", fmt.Sprintf("empresa_%d", empresaID))
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

	imageURL := "/uploads/venta_publica/empresa_" + strconv.FormatInt(empresaID, 10) + "/" + fileName
	if itemID > 0 {
		item, err := dbpkg.GetEmpresaVentaPublicaItemByID(dbEmp, empresaID, itemID)
		if err == nil {
			item.ImagenURL = imageURL
			item.UsuarioCreador = adminEmailFromRequest(r)
			_ = dbpkg.UpdateEmpresaVentaPublicaItem(dbEmp, item)
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"saved":      true,
		"empresa_id": empresaID,
		"item_id":    itemID,
		"image_url":  imageURL,
	})
}

// PublicVentaPublicaHandler expone catalogo y pagos para clientes finales.
func PublicVentaPublicaHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		action := ventaPublicaNormalizeAction(r.URL.Query().Get("action"))
		if action == "" {
			http.Error(w, "action invalida", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "catalogo":
				handleVentaPublicaCatalogoPublico(w, r, dbEmp)
			case "estado_pago":
				handleVentaPublicaEstadoPagoPublico(w, r, dbEmp)
			case "estado_pedido":
				handleVentaPublicaEstadoPedidoPublico(w, r, dbEmp)
			default:
				http.Error(w, "action invalida", http.StatusBadRequest)
			}
		case http.MethodPost:
			switch action {
			case "crear_pago":
				handleVentaPublicaCrearPagoPublico(w, r, dbEmp)
			case "crear_pedido":
				handleVentaPublicaCrearPedidoPublico(w, r, dbEmp)
			default:
				http.Error(w, "action invalida", http.StatusBadRequest)
			}
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func ensureMotelCalipsoPublicShowcase(dbEmp *sql.DB, empresaID int64, cfg *dbpkg.EmpresaVentaPublicaConfig) {
	if dbEmp == nil || cfg == nil || empresaID <= 0 {
		return
	}
	if dbpkg.NormalizeEmpresaPublicSlug(cfg.EmpresaSlug) != "motel-calipso" {
		return
	}

	pageID := int64(0)
	pages, err := dbpkg.ListEmpresaVentaPublicaPaginas(dbEmp, empresaID, true)
	if err == nil {
		for _, page := range pages {
			slug := dbpkg.NormalizeEmpresaPublicSlug(page.Slug)
			name := strings.ToLower(strings.TrimSpace(page.Nombre))
			if slug == "experiencias-calipso" || name == "experiencias calipso" {
				pageID = page.ID
				break
			}
		}
	}
	if pageID <= 0 {
		pageID, err = dbpkg.UpsertEmpresaVentaPublicaPagina(dbEmp, dbpkg.EmpresaVentaPublicaPagina{
			EmpresaID:      empresaID,
			Slug:           "experiencias-calipso",
			Nombre:         "Experiencias Calipso",
			Descripcion:    "Servicios y ambientaciones especiales para estadias privadas en Motel Calipso.",
			OrdenVisual:    1,
			UsuarioCreador: "sistema",
			Estado:         "activo",
			Observaciones:  "Provision automatica de vitrina publica Motel Calipso",
		})
		if err != nil {
			return
		}
	}

	items, _, err := dbpkg.ListEmpresaVentaPublicaItems(dbEmp, empresaID, dbpkg.EmpresaVentaPublicaItemsFilter{
		IncludeInactive: true,
		Limit:           200,
		Offset:          0,
	})
	if err != nil {
		return
	}

	var existing *dbpkg.EmpresaVentaPublicaItem
	for i := range items {
		name := strings.ToLower(strings.TrimSpace(items[i].Nombre))
		code := strings.ToUpper(strings.TrimSpace(items[i].CodigoPublico))
		if code == "CALIPSO-DECORACION-HABITACION" || name == "decoracion de la habitacion" {
			existing = &items[i]
			break
		}
	}

	if existing != nil {
		existing.PaginaID = pageID
		existing.CodigoPublico = "CALIPSO-DECORACION-HABITACION"
		existing.Nombre = "Decoracion de la habitacion"
		existing.Descripcion = "Ambientacion especial de la habitacion con arreglo romantico y detalle visual para una experiencia privada."
		existing.Precio = 120000
		existing.Moneda = "COP"
		existing.ImagenURL = "/img/motel_calipso_decoracion_habitacion.jpg"
		existing.StockPublicado = 20
		existing.OrdenVisual = 1
		existing.Destacado = true
		existing.UsuarioCreador = "sistema"
		existing.Estado = "activo"
		existing.Observaciones = "Provision automatica de vitrina publica Motel Calipso"
		_ = dbpkg.UpdateEmpresaVentaPublicaItem(dbEmp, *existing)
		return
	}

	_, _ = dbpkg.CreateEmpresaVentaPublicaItem(dbEmp, dbpkg.EmpresaVentaPublicaItem{
		EmpresaID:      empresaID,
		PaginaID:       pageID,
		CodigoPublico:  "CALIPSO-DECORACION-HABITACION",
		Nombre:         "Decoracion de la habitacion",
		Descripcion:    "Ambientacion especial de la habitacion con arreglo romantico y detalle visual para una experiencia privada.",
		Precio:         120000,
		Moneda:         "COP",
		ImagenURL:      "/img/motel_calipso_decoracion_habitacion.jpg",
		StockPublicado: 20,
		OrdenVisual:    1,
		Destacado:      true,
		UsuarioCreador: "sistema",
		Estado:         "activo",
		Observaciones:  "Provision automatica de vitrina publica Motel Calipso",
	})
}

func handleVentaPublicaCatalogoPublico(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, _ := parseInt64QueryOptional(r, "empresa_id")
	slug := ventaPublicaSlugFromRequest(r)
	resolvedEmpresaID, err := dbpkg.ResolveVentaPublicaEmpresaIDFromAny(dbEmp, empresaID, slug)
	if err != nil {
		http.Error(w, "empresa no encontrada", http.StatusNotFound)
		return
	}
	cfg, err := dbpkg.GetEmpresaVentaPublicaConfig(dbEmp, resolvedEmpresaID)
	if err != nil {
		http.Error(w, "No se pudo cargar configuracion publica", http.StatusInternalServerError)
		return
	}
	pages, err := dbpkg.ListEmpresaVentaPublicaPaginas(dbEmp, resolvedEmpresaID, false)
	if err != nil {
		http.Error(w, "No se pudieron cargar paginas publicas", http.StatusInternalServerError)
		return
	}
	pageSlug := strings.TrimSpace(r.URL.Query().Get("pagina_slug"))
	if pageSlug == "" {
		pageSlug = strings.TrimSpace(r.URL.Query().Get("page"))
	}
	var currentPage interface{} = nil
	var pageID int64
	if pageSlug != "" {
		page, err := dbpkg.GetEmpresaVentaPublicaPaginaBySlug(dbEmp, resolvedEmpresaID, pageSlug)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "pagina publica no encontrada", http.StatusNotFound)
				return
			}
			http.Error(w, "No se pudo consultar pagina publica", http.StatusInternalServerError)
			return
		}
		currentPage = page
		pageID = page.ID
	}
	items, _, err := dbpkg.ListEmpresaVentaPublicaItems(dbEmp, resolvedEmpresaID, dbpkg.EmpresaVentaPublicaItemsFilter{
		IncludeInactive: false,
		PaginaID:        pageID,
		Q:               strings.TrimSpace(r.URL.Query().Get("q")),
		Sort:            strings.TrimSpace(r.URL.Query().Get("sort")),
		Limit:           firstPositiveInt(limitFromRequest(r), 18),
		Offset:          offsetFromRequest(r),
	})
	if err != nil {
		http.Error(w, "No se pudo cargar catalogo publico", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":           true,
		"empresa_id":   resolvedEmpresaID,
		"empresa_slug": cfg.EmpresaSlug,
		"tienda":       sanitizeVentaPublicaConfigForPublic(cfg),
		"pages":        pages,
		"page":         currentPage,
		"items":        items,
		"paths": map[string]string{
			"local":      "/venta_publica.html?empresa_slug=" + cfg.EmpresaSlug,
			"produccion": "/" + cfg.EmpresaSlug + "/venta_publica.html",
		},
	})
}

func limitFromRequest(r *http.Request) int {
	limit, err := parseIntQueryOptional(r, "limit")
	if err != nil || limit <= 0 {
		return 0
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func offsetFromRequest(r *http.Request) int {
	offset, err := parseIntQueryOptional(r, "offset")
	if err != nil || offset < 0 {
		return 0
	}
	return offset
}

func firstPositiveInt(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func handleVentaPublicaCrearPagoPublico(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	var payload ventaPublicaCrearPagoPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	resolvedEmpresaID, err := dbpkg.ResolveVentaPublicaEmpresaIDFromAny(dbEmp, payload.EmpresaID, payload.EmpresaSlug)
	if err != nil {
		http.Error(w, "empresa no encontrada", http.StatusNotFound)
		return
	}
	cfg, err := dbpkg.GetEmpresaVentaPublicaConfig(dbEmp, resolvedEmpresaID)
	if err != nil {
		http.Error(w, "No se pudo cargar configuracion de tienda", http.StatusInternalServerError)
		return
	}
	if len(payload.Items) == 0 {
		http.Error(w, "items es obligatorio", http.StatusBadRequest)
		return
	}
	metodoPago := ventaPublicaNormalizeMetodoPago(payload.MetodoPago)
	if metodoPago == "" {
		http.Error(w, "metodo_pago invalido", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.CompradorNombre) == "" {
		http.Error(w, "comprador_nombre es obligatorio", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.CompradorEmail) == "" {
		http.Error(w, "comprador_email es obligatorio", http.StatusBadRequest)
		return
	}
	if !payload.AcceptTerms {
		http.Error(w, "debes aceptar terminos para continuar", http.StatusBadRequest)
		return
	}

	rows, _, err := dbpkg.ListEmpresaVentaPublicaItems(dbEmp, resolvedEmpresaID, dbpkg.EmpresaVentaPublicaItemsFilter{IncludeInactive: false, Limit: 500})
	if err != nil {
		http.Error(w, "No se pudo validar catalogo", http.StatusInternalServerError)
		return
	}
	itemsMap := make(map[int64]dbpkg.EmpresaVentaPublicaItem, len(rows))
	for _, row := range rows {
		itemsMap[row.ID] = row
	}

	subtotal := 0.0
	persistedItems := make([]map[string]interface{}, 0, len(payload.Items))
	responseItems := make([]ventaPublicaPagoItemPayload, 0, len(payload.Items))
	for _, reqItem := range payload.Items {
		if reqItem.ItemID <= 0 || reqItem.Cantidad <= 0 {
			continue
		}
		catalogItem, ok := itemsMap[reqItem.ItemID]
		if !ok {
			http.Error(w, fmt.Sprintf("item_id %d no disponible", reqItem.ItemID), http.StatusBadRequest)
			return
		}
		lineSubtotal := catalogItem.Precio * reqItem.Cantidad
		subtotal += lineSubtotal

		persistedItems = append(persistedItems, map[string]interface{}{
			"item_id":        catalogItem.ID,
			"producto_id":    catalogItem.ProductoID,
			"codigo_publico": catalogItem.CodigoPublico,
			"nombre":         catalogItem.Nombre,
			"descripcion":    catalogItem.Descripcion,
			"precio":         catalogItem.Precio,
			"cantidad":       reqItem.Cantidad,
			"subtotal":       lineSubtotal,
			"moneda":         cfg.Moneda,
			"imagen_url":     catalogItem.ImagenURL,
		})
		responseItems = append(responseItems, ventaPublicaPagoItemPayload{
			ItemID:    catalogItem.ID,
			Cantidad:  reqItem.Cantidad,
			Nombre:    catalogItem.Nombre,
			Precio:    catalogItem.Precio,
			Subtotal:  lineSubtotal,
			ImagenURL: catalogItem.ImagenURL,
		})
	}

	if len(responseItems) == 0 || subtotal <= 0 {
		http.Error(w, "no hay items validos para pago", http.StatusBadRequest)
		return
	}

	orderCode := fmt.Sprintf("VP-ORD-%d-%d", resolvedEmpresaID, time.Now().UnixNano())
	orderID, err := dbpkg.CreateEmpresaVentaPublicaOrder(dbEmp, dbpkg.EmpresaVentaPublicaOrder{
		EmpresaID:         resolvedEmpresaID,
		CodigoOrden:       orderCode,
		CompradorNombre:   strings.TrimSpace(payload.CompradorNombre),
		CompradorEmail:    strings.TrimSpace(payload.CompradorEmail),
		CompradorTelefono: strings.TrimSpace(payload.CompradorTelefono),
		Moneda:            cfg.Moneda,
		Subtotal:          subtotal,
		DescuentoTotal:    0,
		ImpuestoTotal:     0,
		Total:             subtotal,
		MetodoPago:        metodoPago,
		EstadoPago:        "pendiente",
		ItemsJSON:         dbpkg.EncodeEmpresaVentaPublicaOrderItems(persistedItems),
		UsuarioCreador:    "publico",
		Estado:            "activo",
	})
	if err != nil {
		http.Error(w, "No se pudo crear orden", http.StatusInternalServerError)
		return
	}

	provider := ventaPublicaProviderFromMetodo(metodoPago)
	phone := strings.TrimSpace(payload.CompradorTelefono)
	reference := dbpkg.BuildVentaPublicaOrderReference(resolvedEmpresaID, orderCode)
	if provider == "wompi" {
		if !cfg.WompiActivo {
			writeJSON(w, http.StatusPreconditionFailed, map[string]interface{}{
				"ok":          false,
				"empresa_id":  resolvedEmpresaID,
				"order_id":    orderID,
				"order_code":  orderCode,
				"provider":    "wompi",
				"metodo_pago": metodoPago,
				"error":       "wompi no esta activo para esta tienda",
			})
			return
		}

		publicKey, err := ventaPublicaResolveCredential(cfg.WompiPublicKey)
		if err != nil {
			http.Error(w, "No se pudo resolver wompi_public_key", http.StatusInternalServerError)
			return
		}
		privateKey, err := ventaPublicaResolveCredential(cfg.WompiPrivateKeyRef)
		if err != nil {
			http.Error(w, "No se pudo resolver wompi_private_key_ref", http.StatusInternalServerError)
			return
		}
		integrityKey, err := ventaPublicaResolveCredential(cfg.WompiIntegrityRef)
		if err != nil {
			http.Error(w, "No se pudo resolver wompi_integrity_key_ref", http.StatusInternalServerError)
			return
		}
		if strings.TrimSpace(publicKey) == "" || strings.TrimSpace(privateKey) == "" || strings.TrimSpace(integrityKey) == "" {
			http.Error(w, "configuracion wompi incompleta para la tienda", http.StatusPreconditionFailed)
			return
		}
		if ok, _ := regexp.MatchString(`^3\d{9}$`, phone); !ok {
			http.Error(w, "comprador_telefono invalido (debe ser telefono Nequi 10 digitos en CO)", http.StatusBadRequest)
			return
		}

		mode := normalizeWompiMode(cfg.WompiMode)
		if mode == "" {
			mode = "sandbox"
		}
		baseURL := wompiBaseURLFromMode(mode)
		acceptanceToken, personalToken, acceptancePermalink, personalPermalink, ferr := fetchWompiAcceptanceInfo(baseURL, publicKey)
		if ferr != nil {
			http.Error(w, "No se pudo consultar terminos de Wompi: "+ferr.Error(), http.StatusBadGateway)
			return
		}
		if strings.TrimSpace(acceptanceToken) == "" || strings.TrimSpace(personalToken) == "" {
			http.Error(w, "Wompi no devolvio tokens de aceptacion", http.StatusBadGateway)
			return
		}

		amountInCents := int64(math.Round(subtotal * 100))
		if amountInCents <= 0 {
			http.Error(w, "monto total invalido", http.StatusBadRequest)
			return
		}
		signatureSource := fmt.Sprintf("%s%dCOP%s", reference, amountInCents, integrityKey)
		signatureHash := sha256.Sum256([]byte(signatureSource))
		signature := hex.EncodeToString(signatureHash[:])
		redirectURL := fmt.Sprintf("%s/venta_publica.html?empresa_slug=%s&order_code=%s&provider=wompi&status=pending", strings.TrimRight(ventaPublicaResolveBaseURL(r), "/"), url.QueryEscape(cfg.EmpresaSlug), url.QueryEscape(orderCode))

		reqBody := map[string]interface{}{
			"acceptance_token":     acceptanceToken,
			"accept_personal_auth": personalToken,
			"amount_in_cents":      amountInCents,
			"currency":             "COP",
			"customer_email":       strings.TrimSpace(payload.CompradorEmail),
			"reference":            reference,
			"signature":            signature,
			"redirect_url":         redirectURL,
			"payment_method": map[string]interface{}{
				"type":         "NEQUI",
				"phone_number": phone,
			},
		}

		bodyBytes, _ := json.Marshal(reqBody)
		apiURL := strings.TrimRight(baseURL, "/") + "/transactions"
		req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(bodyBytes))
		if err != nil {
			http.Error(w, "No se pudo preparar solicitud Wompi", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+privateKey)

		client := &http.Client{Timeout: 25 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "No se pudo crear transaccion Wompi: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		if resp.StatusCode >= 400 {
			_ = dbpkg.UpdateEmpresaVentaPublicaOrderPayment(dbEmp, resolvedEmpresaID, orderCode, "error", "", reference, string(respBody), "", "wompi_error")
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
		if strings.TrimSpace(transactionID) == "" || transactionID == "<nil>" {
			http.Error(w, "Wompi no devolvio transaction id", http.StatusBadGateway)
			return
		}
		if statusWompi == "" || statusWompi == "<nil>" {
			statusWompi = "PENDING"
		}
		statusLocal := ventaPublicaMapWompiStatus(statusWompi)
		pagadoEn := ""
		if statusLocal == "aprobado" {
			pagadoEn = time.Now().In(time.Local).Format("2006-01-02 15:04:05")
		}
		if err := dbpkg.UpdateEmpresaVentaPublicaOrderPayment(dbEmp, resolvedEmpresaID, orderCode, statusLocal, transactionID, reference, string(respBody), pagadoEn, ""); err != nil {
			http.Error(w, "No se pudo actualizar orden", http.StatusInternalServerError)
			return
		}
		order, _ := dbpkg.GetEmpresaVentaPublicaOrderByCodigo(dbEmp, resolvedEmpresaID, orderCode)

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":                      true,
			"empresa_id":              resolvedEmpresaID,
			"empresa_slug":            cfg.EmpresaSlug,
			"order":                   order,
			"items":                   responseItems,
			"provider":                "wompi",
			"metodo_pago":             metodoPago,
			"payment_method":          "NEQUI",
			"mode":                    mode,
			"transaction_id":          transactionID,
			"reference":               reference,
			"status":                  statusWompi,
			"status_local":            statusLocal,
			"acceptance_permalink":    acceptancePermalink,
			"personal_data_permalink": personalPermalink,
			"data":                    data,
		})
		return
	}

	if !cfg.EpaycoActivo {
		writeJSON(w, http.StatusPreconditionFailed, map[string]interface{}{
			"ok":          false,
			"empresa_id":  resolvedEmpresaID,
			"order_id":    orderID,
			"order_code":  orderCode,
			"provider":    "epayco",
			"metodo_pago": metodoPago,
			"error":       "epayco no esta activo para esta tienda",
		})
		return
	}

	epaycoPublicKey, err := ventaPublicaResolveCredential(cfg.EpaycoPublicKey)
	if err != nil {
		http.Error(w, "No se pudo resolver epayco_public_key", http.StatusInternalServerError)
		return
	}
	epaycoPrivateKey, err := ventaPublicaResolveCredential(cfg.EpaycoPrivateKeyRef)
	if err != nil {
		http.Error(w, "No se pudo resolver epayco_private_key_ref", http.StatusInternalServerError)
		return
	}
	if strings.TrimSpace(epaycoPublicKey) == "" || strings.TrimSpace(epaycoPrivateKey) == "" {
		http.Error(w, "configuracion epayco incompleta para la tienda", http.StatusPreconditionFailed)
		return
	}
	mode := normalizeEpaycoMode(cfg.EpaycoMode)
	if mode == "" {
		mode = "sandbox"
	}
	baseURL := strings.TrimRight(ventaPublicaResolveBaseURL(r), "/")
	responseURL := baseURL + "/venta_publica.html?empresa_slug=" + url.QueryEscape(cfg.EmpresaSlug) + "&order_code=" + url.QueryEscape(orderCode) + "&provider=epayco&status=pending&reference=" + url.QueryEscape(reference)
	confirmationURL := baseURL + "/epayco/webhook"
	sessionPayload := map[string]interface{}{
		"checkout_version": "2",
		"name":             strings.TrimSpace(cfg.NombreTienda),
		"description":      "Compra publica " + strings.TrimSpace(orderCode),
		"currency":         "COP",
		"amount":           subtotal,
		"lang":             "ES",
		"invoice":          reference,
		"country":          "CO",
		"taxBase":          0,
		"tax":              0,
		"taxIco":           0,
		"response":         responseURL,
		"confirmation":     confirmationURL,
		"method":           "POST",
		"extras": map[string]interface{}{
			"extra1": orderCode,
			"extra2": strconv.FormatInt(resolvedEmpresaID, 10),
			"extra3": reference,
			"extra4": "venta_publica",
		},
	}
	billing := map[string]interface{}{"email": strings.TrimSpace(payload.CompradorEmail)}
	if phone != "" {
		billing["phone"] = phone
	}
	sessionPayload["billing"] = billing

	apifyToken, loginRaw, err := fetchEpaycoApifyToken(epaycoPublicKey, epaycoPrivateKey)
	if err != nil {
		http.Error(w, "No se pudo autenticar con Epayco Smart Checkout: "+err.Error(), http.StatusBadGateway)
		return
	}
	sessionID, sessionRaw, err := createEpaycoSmartCheckoutSession(apifyToken, sessionPayload)
	if err != nil {
		http.Error(w, "No se pudo crear sesion Smart Checkout de Epayco: "+err.Error(), http.StatusBadGateway)
		return
	}
	rawMap := map[string]interface{}{
		"provider":         "epayco",
		"mode":             mode,
		"reference":        reference,
		"order_code":       orderCode,
		"empresa_id":       resolvedEmpresaID,
		"customer_email":   strings.TrimSpace(payload.CompradorEmail),
		"checkout_type":    "standard",
		"checkout_script":  epaycoSmartCheckoutScriptURL,
		"session_id":       sessionID,
		"response":         responseURL,
		"confirmation":     confirmationURL,
		"integration_flow": "smart_checkout_v2",
		"apify_login_raw":  loginRaw,
		"session_raw":      sessionRaw,
	}
	rawBytes, _ := json.Marshal(rawMap)
	if err := dbpkg.UpdateEmpresaVentaPublicaOrderPayment(dbEmp, resolvedEmpresaID, orderCode, "pendiente", reference, reference, string(rawBytes), "", "epayco_pending"); err != nil {
		http.Error(w, "No se pudo actualizar orden Epayco", http.StatusInternalServerError)
		return
	}
	order, _ := dbpkg.GetEmpresaVentaPublicaOrderByCodigo(dbEmp, resolvedEmpresaID, orderCode)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                  true,
		"empresa_id":          resolvedEmpresaID,
		"empresa_slug":        cfg.EmpresaSlug,
		"order":               order,
		"items":               responseItems,
		"provider":            "epayco",
		"metodo_pago":         metodoPago,
		"mode":                mode,
		"transaction_id":      reference,
		"reference":           reference,
		"status":              "PENDING",
		"status_local":        "pendiente",
		"session_id":          sessionID,
		"checkout_session_id": sessionID,
		"checkout_type":       "standard",
		"checkout_script_url": epaycoSmartCheckoutScriptURL,
		"data": map[string]interface{}{
			"id":         reference,
			"reference":  reference,
			"sessionId":  sessionID,
			"type":       "standard",
			"script_url": epaycoSmartCheckoutScriptURL,
		},
	})
}

func handleVentaPublicaCrearPedidoPublico(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	var payload ventaPublicaCrearPedidoPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	resolvedEmpresaID, err := dbpkg.ResolveVentaPublicaEmpresaIDFromAny(dbEmp, payload.EmpresaID, payload.EmpresaSlug)
	if err != nil {
		http.Error(w, "empresa no encontrada", http.StatusNotFound)
		return
	}
	cfg, err := dbpkg.GetEmpresaVentaPublicaConfig(dbEmp, resolvedEmpresaID)
	if err != nil {
		http.Error(w, "No se pudo cargar la configuracion", http.StatusInternalServerError)
		return
	}
	if !cfg.PedidosRestauranteActivo {
		http.Error(w, "Los pedidos de restaurante no estan activos en esta empresa", http.StatusForbidden)
		return
	}
	if len(payload.Items) == 0 {
		http.Error(w, "items es obligatorio", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.CompradorNombre) == "" || strings.TrimSpace(payload.CompradorTelefono) == "" {
		http.Error(w, "comprador_nombre y comprador_telefono son obligatorios", http.StatusBadRequest)
		return
	}
	canalEntrega := strings.ToLower(strings.TrimSpace(payload.CanalEntrega))
	if canalEntrega == "" || canalEntrega == "delivery" {
		canalEntrega = "domicilio"
	}
	if canalEntrega == "retiro" || canalEntrega == "pickup" {
		canalEntrega = "recoger"
	}
	if canalEntrega == "recoger" && !cfg.PedidosPermitirRecogerEnTienda {
		http.Error(w, "La empresa no tiene activo recoger en tienda", http.StatusBadRequest)
		return
	}
	if canalEntrega != "recoger" && !cfg.PedidosPermitirDomicilio {
		http.Error(w, "La empresa no tiene activo domicilio", http.StatusBadRequest)
		return
	}
	if canalEntrega != "recoger" && strings.TrimSpace(payload.DireccionEntrega) == "" {
		http.Error(w, "direccion_entrega es obligatoria para domicilio", http.StatusBadRequest)
		return
	}

	rows, _, err := dbpkg.ListEmpresaVentaPublicaItems(dbEmp, resolvedEmpresaID, dbpkg.EmpresaVentaPublicaItemsFilter{IncludeInactive: false, Limit: 500})
	if err != nil {
		http.Error(w, "No se pudo validar el catalogo", http.StatusInternalServerError)
		return
	}
	itemsMap := make(map[int64]dbpkg.EmpresaVentaPublicaItem, len(rows))
	for _, row := range rows {
		itemsMap[row.ID] = row
	}

	subtotal := 0.0
	persistedItems := make([]map[string]interface{}, 0, len(payload.Items))
	for _, reqItem := range payload.Items {
		if reqItem.ItemID <= 0 || reqItem.Cantidad <= 0 {
			continue
		}
		catalogItem, ok := itemsMap[reqItem.ItemID]
		if !ok {
			http.Error(w, fmt.Sprintf("item_id %d no disponible", reqItem.ItemID), http.StatusBadRequest)
			return
		}
		lineSubtotal := catalogItem.Precio * reqItem.Cantidad
		subtotal += lineSubtotal
		persistedItems = append(persistedItems, map[string]interface{}{
			"item_id":        catalogItem.ID,
			"producto_id":    catalogItem.ProductoID,
			"codigo_publico": catalogItem.CodigoPublico,
			"nombre":         catalogItem.Nombre,
			"precio":         catalogItem.Precio,
			"cantidad":       reqItem.Cantidad,
			"subtotal":       lineSubtotal,
			"moneda":         cfg.Moneda,
			"imagen_url":     catalogItem.ImagenURL,
		})
	}
	if len(persistedItems) == 0 || subtotal <= 0 {
		http.Error(w, "No hay productos validos para el pedido", http.StatusBadRequest)
		return
	}

	orderCode := fmt.Sprintf("VP-FOOD-%d-%d", resolvedEmpresaID, time.Now().UnixNano())
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%d|%d", orderCode, resolvedEmpresaID, time.Now().UnixNano())))
	trackingToken := hex.EncodeToString(sum[:])
	if len(trackingToken) > 24 {
		trackingToken = trackingToken[:24]
	}
	orderID, err := dbpkg.CreateEmpresaVentaPublicaOrder(dbEmp, dbpkg.EmpresaVentaPublicaOrder{
		EmpresaID:                resolvedEmpresaID,
		CodigoOrden:              orderCode,
		TipoOrden:                "pedido_restaurante",
		CompradorNombre:          strings.TrimSpace(payload.CompradorNombre),
		CompradorEmail:           strings.TrimSpace(payload.CompradorEmail),
		CompradorTelefono:        strings.TrimSpace(payload.CompradorTelefono),
		Moneda:                   cfg.Moneda,
		Subtotal:                 subtotal,
		Total:                    subtotal,
		MetodoPago:               "pedido_restaurante",
		EstadoPago:               "pendiente",
		EstadoPedido:             "recibido",
		CanalEntrega:             canalEntrega,
		DireccionEntrega:         strings.TrimSpace(payload.DireccionEntrega),
		NotasEntrega:             strings.TrimSpace(payload.NotasEntrega),
		ClienteComparteUbicacion: payload.ClienteComparteUbicacion,
		EntregaLatitud:           payload.EntregaLatitud,
		EntregaLongitud:          payload.EntregaLongitud,
		TrackingToken:            trackingToken,
		ItemsJSON:                dbpkg.EncodeEmpresaVentaPublicaOrderItems(persistedItems),
		UsuarioCreador:           "publico",
		Estado:                   "activo",
		Observaciones:            "pedido_restaurante_web",
	})
	if err != nil {
		http.Error(w, "No se pudo crear el pedido", http.StatusInternalServerError)
		return
	}
	_ = orderID
	order, err := dbpkg.GetEmpresaVentaPublicaOrderByCodigo(dbEmp, resolvedEmpresaID, orderCode)
	if err != nil {
		http.Error(w, "Pedido creado, pero no se pudo consultar", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"ok":            true,
		"empresa_id":    resolvedEmpresaID,
		"empresa_slug":  cfg.EmpresaSlug,
		"order":         order,
		"tracking_token": trackingToken,
	})
}

func handleVentaPublicaEstadoPedidoPublico(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, _ := parseInt64QueryOptional(r, "empresa_id")
	slug := ventaPublicaSlugFromRequest(r)
	resolvedEmpresaID, err := dbpkg.ResolveVentaPublicaEmpresaIDFromAny(dbEmp, empresaID, slug)
	if err != nil {
		http.Error(w, "empresa no encontrada", http.StatusNotFound)
		return
	}
	orderCode := strings.TrimSpace(r.URL.Query().Get("order_code"))
	trackingToken := strings.TrimSpace(r.URL.Query().Get("tracking_token"))
	var order dbpkg.EmpresaVentaPublicaOrder
	if trackingToken != "" {
		order, err = dbpkg.GetEmpresaVentaPublicaOrderByTrackingToken(dbEmp, resolvedEmpresaID, trackingToken)
	} else {
		order, err = dbpkg.GetEmpresaVentaPublicaOrderByCodigo(dbEmp, resolvedEmpresaID, orderCode)
	}
	if err != nil {
		http.Error(w, "pedido no encontrado", http.StatusNotFound)
		return
	}
	resp := map[string]interface{}{
		"ok":      true,
		"order":   order,
		"tracking": map[string]interface{}{"enabled": false},
	}
	if order.TaxiRequestID > 0 {
		if req, err := dbpkg.GetTaxiRequestByID(dbEmp, resolvedEmpresaID, order.TaxiRequestID); err == nil {
			route, _ := dbpkg.ListTaxiRoutePoints(dbEmp, resolvedEmpresaID, order.TaxiRequestID, 300)
			resp["tracking"] = map[string]interface{}{
				"enabled": true,
				"request": req,
				"route":   route,
			}
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleVentaPublicaEstadoPagoPublico(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, _ := parseInt64QueryOptional(r, "empresa_id")
	slug := ventaPublicaSlugFromRequest(r)
	resolvedEmpresaID, err := dbpkg.ResolveVentaPublicaEmpresaIDFromAny(dbEmp, empresaID, slug)
	if err != nil {
		http.Error(w, "empresa no encontrada", http.StatusNotFound)
		return
	}
	orderCode := strings.TrimSpace(r.URL.Query().Get("order_code"))
	if orderCode == "" {
		orderCode = strings.TrimSpace(r.URL.Query().Get("codigo_orden"))
	}
	if orderCode == "" {
		http.Error(w, "order_code es obligatorio", http.StatusBadRequest)
		return
	}

	order, err := dbpkg.GetEmpresaVentaPublicaOrderByCodigo(dbEmp, resolvedEmpresaID, orderCode)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "orden no encontrada", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo consultar orden", http.StatusInternalServerError)
		return
	}

	transactionID := strings.TrimSpace(r.URL.Query().Get("transaction_id"))
	if transactionID == "" {
		transactionID = strings.TrimSpace(order.TransactionID)
	}
	cfg, err := dbpkg.GetEmpresaVentaPublicaConfig(dbEmp, resolvedEmpresaID)
	if err != nil {
		http.Error(w, "No se pudo cargar configuracion de tienda", http.StatusInternalServerError)
		return
	}

	statusLocal := strings.TrimSpace(order.EstadoPago)
	statusGateway := ""
	data := map[string]interface{}{}
	metodoPago := ventaPublicaNormalizeMetodoPago(r.URL.Query().Get("metodo_pago"))
	if metodoPago == "" {
		metodoPago = ventaPublicaNormalizeMetodoPago(order.MetodoPago)
	}
	provider := ventaPublicaProviderFromMetodo(metodoPago)

	if provider == "wompi" && cfg.WompiActivo && transactionID != "" {
		publicKey, _ := ventaPublicaResolveCredential(cfg.WompiPublicKey)
		privateKey, _ := ventaPublicaResolveCredential(cfg.WompiPrivateKeyRef)
		if strings.TrimSpace(publicKey) != "" {
			mode := normalizeWompiMode(cfg.WompiMode)
			if mode == "" {
				mode = "sandbox"
			}
			baseURL := wompiBaseURLFromMode(mode)
			statusURL := strings.TrimRight(baseURL, "/") + "/transactions/" + url.PathEscape(transactionID)

			fetchStatus := func(authKey string) ([]byte, int, error) {
				req, err := http.NewRequest(http.MethodGet, statusURL, nil)
				if err != nil {
					return nil, 0, err
				}
				req.Header.Set("Authorization", "Bearer "+authKey)
				client := &http.Client{Timeout: 20 * time.Second}
				resp, err := client.Do(req)
				if err != nil {
					return nil, 0, err
				}
				defer resp.Body.Close()
				body, _ := io.ReadAll(resp.Body)
				return body, resp.StatusCode, nil
			}

			respBody, statusCode, ferr := fetchStatus(publicKey)
			if ferr == nil {
				if statusCode >= 400 && strings.TrimSpace(privateKey) != "" {
					if body2, code2, err2 := fetchStatus(privateKey); err2 == nil {
						respBody = body2
						statusCode = code2
					}
				}
				if statusCode < 400 {
					var wompiResp map[string]interface{}
					if err := json.Unmarshal(respBody, &wompiResp); err == nil {
						data, _ = wompiResp["data"].(map[string]interface{})
						statusGateway = strings.ToUpper(strings.TrimSpace(fmt.Sprint(data["status"])))
						if statusGateway == "" || statusGateway == "<nil>" {
							statusGateway = "PENDING"
						}
						statusLocal = ventaPublicaMapWompiStatus(statusGateway)
						pagadoEn := ""
						if statusLocal == "aprobado" {
							pagadoEn = time.Now().In(time.Local).Format("2006-01-02 15:04:05")
						}
						_ = dbpkg.UpdateEmpresaVentaPublicaOrderPayment(dbEmp, resolvedEmpresaID, orderCode, statusLocal, transactionID, order.ReferenciaExterna, string(respBody), pagadoEn, "status_check")
						order, _ = dbpkg.GetEmpresaVentaPublicaOrderByCodigo(dbEmp, resolvedEmpresaID, orderCode)
					}
				}
			}
		}
	}
	if provider == "epayco" && cfg.EpaycoActivo {
		recordReference := strings.TrimSpace(order.ReferenciaExterna)
		if recordReference == "" {
			recordReference = strings.TrimSpace(r.URL.Query().Get("reference"))
		}
		if recordReference != "" {
			validationURL := "https://secure.epayco.co/validation/v1/reference/" + url.PathEscape(recordReference)
			req, err := http.NewRequest(http.MethodGet, validationURL, nil)
			if err == nil {
				client := &http.Client{Timeout: 15 * time.Second}
				resp, reqErr := client.Do(req)
				if reqErr == nil {
					defer resp.Body.Close()
					respBody, _ := io.ReadAll(resp.Body)
					if resp.StatusCode < 400 {
						if err := json.Unmarshal(respBody, &data); err == nil {
							statusGateway = parseEpaycoPaymentStatus(data)
							if statusGateway == "ERROR" && shouldPreservePendingEpaycoStatus(strings.ToUpper(strings.TrimSpace(order.EstadoPago)), data) {
								statusGateway = "PENDING"
							}
							if strings.TrimSpace(statusGateway) == "" {
								statusGateway = "PENDING"
							}
							statusLocal = ventaPublicaMapEpaycoStatus(statusGateway)
							gatewayTx := firstNonEmptyString(
								strings.TrimSpace(pickEpaycoField(data, "x_transaction_id", "transaction_id", "id")),
								transactionID,
								order.TransactionID,
							)
							pagadoEn := ""
							if statusLocal == "aprobado" {
								pagadoEn = time.Now().In(time.Local).Format("2006-01-02 15:04:05")
							}
							_ = dbpkg.UpdateEmpresaVentaPublicaOrderPayment(dbEmp, resolvedEmpresaID, orderCode, statusLocal, gatewayTx, recordReference, string(respBody), pagadoEn, "status_check")
							order, _ = dbpkg.GetEmpresaVentaPublicaOrderByCodigo(dbEmp, resolvedEmpresaID, orderCode)
						}
					}
				}
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":           true,
		"empresa_id":   resolvedEmpresaID,
		"empresa_slug": cfg.EmpresaSlug,
		"provider":     provider,
		"metodo_pago":  metodoPago,
		"order":        order,
		"status":       statusGateway,
		"status_local": statusLocal,
		"data":         data,
	})
}
