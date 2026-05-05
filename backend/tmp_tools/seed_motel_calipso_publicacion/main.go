package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

type result struct {
	OK             bool     `json:"ok"`
	EmpresaID      int64    `json:"empresa_id"`
	EmpresaNombre  string   `json:"empresa_nombre"`
	ConfigID       int64    `json:"config_id"`
	Slug           string   `json:"slug"`
	Paginas        []int64  `json:"paginas"`
	Items          []int64  `json:"items"`
	Publicaciones  []int    `json:"publicaciones"`
	URLs           []string `json:"urls"`
	Verificaciones []string `json:"verificaciones"`
	Advertencias   []string `json:"advertencias,omitempty"`
}

type seedPage struct {
	Slug        string
	Nombre      string
	Descripcion string
	BannerURL   string
	Orden       int
}

type seedItem struct {
	PageSlug    string
	Code        string
	Nombre      string
	Descripcion string
	Precio      float64
	ImagenURL   string
	Stock       float64
	Orden       int
	Destacado   bool
}

func loadEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		raw := strings.TrimSpace(line)
		if raw == "" || strings.HasPrefix(raw, "#") || !strings.Contains(raw, "=") {
			continue
		}
		parts := strings.SplitN(raw, "=", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
		if key != "" && os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
}

func tunnelCandidate(raw string) string {
	if strings.TrimSpace(os.Getenv("DB_VPS_TUNNEL_ENABLED")) != "1" {
		return raw
	}
	port := strings.TrimSpace(os.Getenv("DB_VPS_LOCAL_PORT"))
	if port == "" {
		port = "15432"
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	u.Host = "127.0.0.1:" + port
	return u.String()
}

func openDB(rawDSN string) (*sql.DB, error) {
	if strings.TrimSpace(rawDSN) == "" {
		return nil, errors.New("DB_EMPRESAS_DSN no definido")
	}
	candidates := []string{rawDSN}
	rewritten := tunnelCandidate(rawDSN)
	if rewritten != rawDSN {
		candidates = append(candidates, rewritten)
	}
	var lastErr error
	for _, dsn := range candidates {
		db, err := sql.Open(dbpkg.PostgresCompatDriverName(), dsn)
		if err != nil {
			lastErr = err
			continue
		}
		if err = db.Ping(); err == nil {
			return db, nil
		}
		lastErr = err
		_ = db.Close()
	}
	return nil, lastErr
}

func resolveEmpresa(db *sql.DB, preferredID int64) (int64, string, error) {
	var nombre string
	if preferredID > 0 {
		err := db.QueryRow(`SELECT COALESCE(nombre, '') FROM empresas WHERE id=$1 OR COALESCE(empresa_id, 0)=$1 LIMIT 1`, preferredID).Scan(&nombre)
		if err == nil {
			return preferredID, nombre, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return 0, "", err
		}
	}
	var id int64
	err := db.QueryRow(`SELECT COALESCE(NULLIF(empresa_id, 0), id), COALESCE(nombre, '') FROM empresas WHERE lower(COALESCE(nombre, '')) LIKE '%calipso%' ORDER BY id LIMIT 1`).Scan(&id, &nombre)
	if err != nil {
		return 0, "", err
	}
	return id, nombre, nil
}

func ensureItem(db *sql.DB, empresaID int64, pageID int64, item seedItem, usuario string) (int64, error) {
	existing, _, err := dbpkg.ListEmpresaVentaPublicaItems(db, empresaID, dbpkg.EmpresaVentaPublicaItemsFilter{
		IncludeInactive: true,
		PaginaID:        pageID,
		Limit:           500,
	})
	if err != nil {
		return 0, err
	}
	codePrefix := strings.ToLower(strings.TrimSpace(item.Code))
	name := strings.ToLower(strings.TrimSpace(item.Nombre))
	for _, current := range existing {
		code := strings.ToLower(strings.TrimSpace(current.CodigoPublico))
		code = strings.TrimSuffix(code, fmt.Sprintf("-p%d", pageID))
		if code == codePrefix || strings.ToLower(strings.TrimSpace(current.Nombre)) == name {
			current.PaginaID = pageID
			current.CodigoPublico = item.Code
			current.Nombre = item.Nombre
			current.Descripcion = item.Descripcion
			current.Precio = item.Precio
			current.Moneda = "COP"
			current.ImagenURL = item.ImagenURL
			current.StockPublicado = item.Stock
			current.OrdenVisual = item.Orden
			current.Destacado = item.Destacado
			current.UsuarioCreador = usuario
			current.Estado = "activo"
			current.Observaciones = "Publicado para carta publica, QR y venta publica de Motel Calipso"
			return current.ID, dbpkg.UpdateEmpresaVentaPublicaItem(db, current)
		}
	}
	return dbpkg.CreateEmpresaVentaPublicaItem(db, dbpkg.EmpresaVentaPublicaItem{
		EmpresaID:      empresaID,
		PaginaID:       pageID,
		CodigoPublico:  item.Code,
		Nombre:         item.Nombre,
		Descripcion:    item.Descripcion,
		Precio:         item.Precio,
		Moneda:         "COP",
		ImagenURL:      item.ImagenURL,
		StockPublicado: item.Stock,
		OrdenVisual:    item.Orden,
		Destacado:      item.Destacado,
		UsuarioCreador: usuario,
		Estado:         "activo",
		Observaciones:  "Publicado para carta publica, QR y venta publica de Motel Calipso",
	})
}

func ensurePublicacion(db *sql.DB, empresaID int64, nombre, descripcion, fotoURL string) (int, error) {
	var id int
	err := db.QueryRow(`SELECT id FROM empresa_publicaciones_red_social WHERE empresa_id=$1 AND lower(nombre)=lower($2) LIMIT 1`, empresaID, nombre).Scan(&id)
	pub := &dbpkg.PublicacionRedSocial{
		ID:          id,
		EmpresaID:   int(empresaID),
		Nombre:      nombre,
		Descripcion: descripcion,
		FotoURL:     fotoURL,
		YoutubeURL:  "",
		Estado:      "activo",
	}
	if err == nil && id > 0 {
		return id, dbpkg.UpdatePublicacionRedSocial(db, pub)
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	if err := dbpkg.InsertPublicacionRedSocial(db, pub); err != nil {
		return 0, err
	}
	return pub.ID, nil
}

func main() {
	empresaFlag := flag.Int64("empresa_id", 7, "empresa_id Motel Calipso")
	usuario := flag.String("usuario", "codex_publicacion", "usuario creador")
	dsnFlag := flag.String("dsn", "", "DSN empresas")
	flag.Parse()

	loadEnv(".env.local")
	if strings.TrimSpace(os.Getenv("DB_DIALECT")) == "" {
		_ = os.Setenv("DB_DIALECT", "postgres")
	}
	dsn := strings.TrimSpace(*dsnFlag)
	if dsn == "" {
		dsn = strings.TrimSpace(os.Getenv("DB_EMPRESAS_DSN"))
	}
	db, err := openDB(dsn)
	if err != nil {
		log.Fatalf("conexion DB empresas: %v", err)
	}
	defer db.Close()

	if err := dbpkg.EnsurePostgresRuntimeCompat(db); err != nil {
		log.Fatalf("EnsurePostgresRuntimeCompat: %v", err)
	}
	if err := dbpkg.EnsureEmpresaVentaPublicaSchema(db); err != nil {
		log.Fatalf("EnsureEmpresaVentaPublicaSchema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaRedSocialInteraccionesSchema(db); err != nil {
		log.Fatalf("EnsureEmpresaRedSocialInteraccionesSchema: %v", err)
	}

	empresaID, empresaNombre, err := resolveEmpresa(db, *empresaFlag)
	if err != nil {
		log.Fatalf("resolveEmpresa: %v", err)
	}

	cfg := dbpkg.EmpresaVentaPublicaConfig{
		EmpresaID:                       empresaID,
		EmpresaSlug:                     "motel-calipso",
		NombreTienda:                    "Motel Calipso",
		DescripcionTienda:               "Carta publica, productos, servicios, experiencias y venta online de Motel Calipso.",
		LogoURL:                         "/img/motel.png",
		BannerURL:                       "/img/motel_calipso_decoracion_habitacion.jpg",
		ColorPrimario:                   "#0f766e",
		TemaVisual:                      "catalogo_publico",
		Moneda:                          "COP",
		MostrarStock:                    true,
		PedidosRegistroOpcionalCliente:  true,
		PedidosPermitirRecogerEnTienda:  true,
		PedidosPermitirDomicilio:        true,
		PedidosTrackingDomiciliario:     true,
		PedidosDespachoAutomatico:       true,
		PedidosNombreSistema:            "Pedidos Motel Calipso",
		PedidosTiempoPreparacionMinutos: 20,
		WompiMode:                       "sandbox",
		EpaycoMode:                      "sandbox",
		UsuarioCreador:                  *usuario,
		Estado:                          "activo",
		Observaciones:                   "Publicacion activa para POS, red social, venta publica, carta publica y QR imprimible.",
	}
	if existing, err := dbpkg.GetEmpresaVentaPublicaConfig(db, empresaID); err == nil {
		if strings.TrimSpace(existing.LogoURL) != "" {
			cfg.LogoURL = existing.LogoURL
		}
		if strings.TrimSpace(existing.BannerURL) != "" {
			cfg.BannerURL = existing.BannerURL
		}
		cfg.DominioPublico = existing.DominioPublico
		cfg.WompiActivo = existing.WompiActivo
		cfg.WompiMode = existing.WompiMode
		cfg.WompiPublicKey = existing.WompiPublicKey
		cfg.WompiPrivateKeyRef = existing.WompiPrivateKeyRef
		cfg.WompiIntegrityRef = existing.WompiIntegrityRef
		cfg.WompiEventKeyRef = existing.WompiEventKeyRef
		cfg.EpaycoActivo = existing.EpaycoActivo
		cfg.EpaycoMode = existing.EpaycoMode
		cfg.EpaycoPublicKey = existing.EpaycoPublicKey
		cfg.EpaycoPrivateKeyRef = existing.EpaycoPrivateKeyRef
		cfg.EpaycoCustomerID = existing.EpaycoCustomerID
	}
	configID, err := dbpkg.UpsertEmpresaVentaPublicaConfig(db, cfg)
	if err != nil {
		log.Fatalf("UpsertEmpresaVentaPublicaConfig: %v", err)
	}

	pages := []seedPage{
		{
			Slug:        "experiencias-calipso",
			Nombre:      "Experiencias Calipso",
			Descripcion: "Paquetes, decoraciones y servicios especiales disponibles para venta publica de Motel Calipso.",
			BannerURL:   "/img/motel_calipso_decoracion_habitacion.jpg",
			Orden:       1,
		},
		{
			Slug:        "carta-productos-precios",
			Nombre:      "Carta publica de productos y precios",
			Descripcion: "Productos, minibar, amenities y servicios visibles para clientes externos mediante carta publica y QR.",
			BannerURL:   "/img/motel.png",
			Orden:       2,
		},
		{
			Slug:        "pos-motel-calipso",
			Nombre:      "POS Motel Calipso",
			Descripcion: "Pagina publica para promocionar la operacion POS y los servicios de venta conectados a Motel Calipso.",
			BannerURL:   "/img/logo punto_venta.png",
			Orden:       3,
		},
	}
	pageIDs := map[string]int64{}
	out := result{
		OK:            true,
		EmpresaID:     empresaID,
		EmpresaNombre: empresaNombre,
		ConfigID:      configID,
		Slug:          cfg.EmpresaSlug,
	}
	for _, page := range pages {
		id, err := dbpkg.UpsertEmpresaVentaPublicaPagina(db, dbpkg.EmpresaVentaPublicaPagina{
			EmpresaID:      empresaID,
			Slug:           page.Slug,
			Nombre:         page.Nombre,
			Descripcion:    page.Descripcion,
			BannerURL:      page.BannerURL,
			OrdenVisual:    page.Orden,
			UsuarioCreador: *usuario,
			Estado:         "activo",
			Observaciones:  "Pagina publicada para Motel Calipso",
		})
		if err != nil {
			log.Fatalf("UpsertEmpresaVentaPublicaPagina %s: %v", page.Slug, err)
		}
		pageIDs[page.Slug] = id
		out.Paginas = append(out.Paginas, id)
	}

	items := []seedItem{
		{"experiencias-calipso", "CALIPSO-DECORACION-HABITACION", "Decoracion de la habitacion", "Ambientacion romantica con flores, velas LED y montaje especial antes del ingreso.", 120000, "/img/motel_calipso_decoracion_habitacion.jpg", 20, 1, true},
		{"experiencias-calipso", "CALIPSO-NOCHE-ROMANTICA", "Noche romantica Calipso", "Paquete de estadia privada con ambientacion especial, minibar basico y atencion prioritaria.", 260000, "/img/motel_calipso_decoracion_habitacion.jpg", 10, 2, true},
		{"carta-productos-precios", "CALIPSO-BEBIDAS-SNACKS", "Combo bebidas y snacks", "Seleccion de bebidas personales y snacks premium para habitacion.", 45000, "/img/motel.png", 50, 1, true},
		{"carta-productos-precios", "CALIPSO-KIT-ASEO", "Kit de aseo premium", "Kit de aseo personal para reposicion o compra directa desde la carta publica.", 28000, "/img/motel.png", 30, 2, false},
		{"carta-productos-precios", "CALIPSO-DESAYUNO-HABITACION", "Desayuno en habitacion", "Servicio de desayuno sencillo entregado en la habitacion.", 38000, "/img/hotel-logo.svg", 25, 3, false},
		{"pos-motel-calipso", "CALIPSO-POS-VENTA", "POS Motel Calipso", "Operacion POS conectada a productos, carta publica, pagos y red social comercial.", 0, "/img/logo punto_venta.png", 1, 1, true},
	}
	for _, item := range items {
		pageID := pageIDs[item.PageSlug]
		id, err := ensureItem(db, empresaID, pageID, item, *usuario)
		if err != nil {
			log.Fatalf("ensureItem %s: %v", item.Code, err)
		}
		out.Items = append(out.Items, id)
	}

	publicSale := "/motel-calipso/venta_publica.html"
	publicCatalog := "/motel-calipso/visualizar_productos_y_precios_publico.html"
	localCatalog := "/visualizar_productos_y_precios_publico.html?empresa_slug=motel-calipso"
	localSale := "/venta_publica.html?empresa_slug=motel-calipso"
	redSocial := "/red_social_comercial.html"

	posts := []struct {
		Nombre      string
		Descripcion string
		FotoURL     string
	}{
		{
			Nombre:      "POS y carta publica de Motel Calipso",
			Descripcion: "Motel Calipso ya tiene publicado su POS comercial con venta publica, carta de productos y QR imprimible.\nVenta publica: " + publicSale + "\nCarta y precios: " + publicCatalog + "\nEn pruebas locales: " + localSale + " y " + localCatalog,
			FotoURL:     "/img/logo punto_venta.png",
		},
		{
			Nombre:      "Experiencias Calipso disponibles en linea",
			Descripcion: "Decoraciones, paquetes romanticos, minibar y servicios publicados para consulta externa desde la carta publica de Motel Calipso.\nAbre la carta: " + publicCatalog + "\nEl QR se genera y exporta desde Administrar empresa > Carta publica.",
			FotoURL:     "/img/motel_calipso_decoracion_habitacion.jpg",
		},
	}
	for _, post := range posts {
		id, err := ensurePublicacion(db, empresaID, post.Nombre, post.Descripcion, post.FotoURL)
		if err != nil {
			log.Fatalf("ensurePublicacion %s: %v", post.Nombre, err)
		}
		out.Publicaciones = append(out.Publicaciones, id)
	}

	out.URLs = []string{publicSale, publicCatalog, localSale, localCatalog, redSocial}
	var pagesCount, itemsCount, postsCount int
	_ = db.QueryRow(`SELECT COUNT(*) FROM empresa_venta_publica_paginas WHERE empresa_id=$1 AND COALESCE(estado,'activo')='activo'`, empresaID).Scan(&pagesCount)
	_ = db.QueryRow(`SELECT COUNT(*) FROM empresa_venta_publica_items WHERE empresa_id=$1 AND COALESCE(estado,'activo')='activo'`, empresaID).Scan(&itemsCount)
	_ = db.QueryRow(`SELECT COUNT(*) FROM empresa_publicaciones_red_social WHERE empresa_id=$1 AND COALESCE(estado,'activo')='activo'`, empresaID).Scan(&postsCount)
	out.Verificaciones = append(out.Verificaciones,
		fmt.Sprintf("paginas_activas=%d", pagesCount),
		fmt.Sprintf("items_publicos_activos=%d", itemsCount),
		fmt.Sprintf("publicaciones_red_social_activas=%d", postsCount),
		"qr_admin_exporta_png_svg_pdf=/administrar_empresa/carta_productos_publica.html",
	)
	if strings.TrimSpace(cfg.DominioPublico) == "" {
		out.Advertencias = append(out.Advertencias, "No hay dominio_publico personalizado; se publican rutas por slug y enlaces locales con empresa_slug.")
	}

	b, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(b))
}
