package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type seedProduct struct {
	SKU, Nombre, Categoria string
	Costo, Precio          float64
}
type seedCombo struct {
	Codigo, Nombre, Descripcion string
	Precio                      float64
	Ingredientes                map[string]float64
}

func must(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}

func firstProductBySKU(db *sql.DB, empresaID int64, sku string) (*dbpkg.Producto, error) {
	rows, err := dbpkg.GetProductosByEmpresa(db, empresaID, sku, "", 0, 0, 50, 0)
	if err != nil {
		return nil, err
	}
	for _, p := range rows {
		if strings.EqualFold(strings.TrimSpace(p.SKU), strings.TrimSpace(sku)) {
			pp := p
			return &pp, nil
		}
	}
	return nil, sql.ErrNoRows
}

func ensureBodega(db *sql.DB, empresaID int64, usuario string) (int64, error) {
	bodegas, err := dbpkg.GetBodegasByEmpresa(db, empresaID, true)
	if err != nil {
		return 0, err
	}
	for _, b := range bodegas {
		if strings.EqualFold(b.Codigo, "CALIPSO-MINIBAR") || strings.Contains(strings.ToLower(b.Nombre), "calipso") {
			return b.ID, nil
		}
	}
	return dbpkg.CreateBodega(db, dbpkg.Bodega{EmpresaID: empresaID, Codigo: "CALIPSO-MINIBAR", Nombre: "Bodega Minibar Motel Calipso", Ubicacion: "Operativa", Responsable: usuario, UsuarioCreador: usuario, Estado: "activo", Observaciones: "seed combos motel calipso"})
}

func ensureStock(db *sql.DB, empresaID, productoID, bodegaID int64, minimo float64, usuario string) error {
	ex, err := dbpkg.GetExistenciasByEmpresa(db, empresaID, productoID, bodegaID, 10, 0)
	if err != nil {
		return err
	}
	actual := 0.0
	if len(ex) > 0 {
		actual = ex[0].Cantidad
	}
	if actual >= minimo {
		return nil
	}
	return dbpkg.RegistrarMovimientoInventario(db, empresaID, productoID, bodegaID, "ajuste_positivo", minimo-actual, fmt.Sprintf("SEED-CALIPSO-%d", time.Now().UnixNano()), usuario, "stock inicial para combos de venta")
}

func comboExists(db *sql.DB, empresaID int64, codigo string) (bool, error) {
	rows, err := dbpkg.GetCombosProductosByEmpresa(db, empresaID, codigo, "", true, 50, 0)
	if err != nil {
		return false, err
	}
	for _, c := range rows {
		if strings.EqualFold(strings.TrimSpace(c.Codigo), strings.TrimSpace(codigo)) {
			return true, nil
		}
	}
	return false, nil
}

func main() {
	empresaID := flag.Int64("empresa_id", 7, "empresa_id Motel Calipso")
	usuario := flag.String("usuario", "codex_seed", "usuario")
	flag.Parse()
	dsn := strings.TrimSpace(os.Getenv("DB_EMPRESAS_DSN"))
	if dsn == "" {
		log.Fatal("DB_EMPRESAS_DSN no definido")
	}
	db, err := sql.Open(dbpkg.PostgresCompatDriverName(), dsn)
	must(err, "sql.Open")
	defer db.Close()
	must(db.Ping(), "db.Ping")
	_ = dbpkg.EnsurePostgresRuntimeCompat(db)
	must(dbpkg.EnsureEmpresaProductosSchema(db), "EnsureEmpresaProductosSchema")
	must(dbpkg.EnsureEmpresaCarritosSchema(db), "EnsureEmpresaCarritosSchema")

	bodegaID, err := ensureBodega(db, *empresaID, *usuario)
	must(err, "ensureBodega")

	products := []seedProduct{
		{"CAL-AGUA-600", "Agua botella 600 ml", "Minibar", 1200, 3000},
		{"CAL-GASEOSA", "Gaseosa personal", "Minibar", 1800, 4500},
		{"CAL-SNACK", "Snack premium habitacion", "Minibar", 2500, 6500},
		{"CAL-KIT-ASEO", "Kit aseo personal", "Amenities", 3500, 9000},
		{"CAL-DECOR", "Decoracion romantica", "Experiencias", 15000, 35000},
		{"CAL-DESAYUNO", "Desayuno sencillo", "Room service", 8000, 18000},
		{"CAL-VINO", "Vino media botella", "Bebidas", 18000, 42000},
	}
	prodMap := map[string]dbpkg.Producto{}
	createdProducts := 0
	for _, sp := range products {
		p, err := firstProductBySKU(db, *empresaID, sp.SKU)
		if err == sql.ErrNoRows {
			id, err := dbpkg.CreateProducto(db, dbpkg.Producto{EmpresaID: *empresaID, BodegaPrincipalID: bodegaID, SKU: sp.SKU, CodigoBarras: sp.SKU, Nombre: sp.Nombre, Categoria: sp.Categoria, UnidadMedida: "unidad", Costo: sp.Costo, Precio: sp.Precio, ImpuestoPorcentaje: 19, StockMinimo: 5, StockMaximo: 1000, UsuarioCreador: *usuario, Estado: "activo", Observaciones: "seed combos motel calipso"}, 80, "stock inicial calipso")
			must(err, "CreateProducto "+sp.SKU)
			p, err = firstProductBySKU(db, *empresaID, sp.SKU)
			if err != nil {
				p = &dbpkg.Producto{ID: id, SKU: sp.SKU, Nombre: sp.Nombre}
			}
			createdProducts++
		} else {
			must(err, "firstProductBySKU "+sp.SKU)
			must(ensureStock(db, *empresaID, p.ID, bodegaID, 80, *usuario), "ensureStock "+sp.SKU)
		}
		prodMap[sp.SKU] = *p
	}

	combos := []seedCombo{
		{"CAL-COMBO-NOCHE", "Combo Noche Calipso", "Minibar para estadia corta: agua, gaseosa y snack.", 12500, map[string]float64{"CAL-AGUA-600": 1, "CAL-GASEOSA": 1, "CAL-SNACK": 1}},
		{"CAL-COMBO-ROMANTICO", "Combo Romantico Calipso", "Decoracion, vino y snack premium para experiencia especial.", 76000, map[string]float64{"CAL-DECOR": 1, "CAL-VINO": 1, "CAL-SNACK": 1}},
		{"CAL-COMBO-AMENITIES", "Combo Amenities Calipso", "Kit de aseo con bebida para reposicion rapida de habitacion.", 14500, map[string]float64{"CAL-KIT-ASEO": 1, "CAL-AGUA-600": 1}},
		{"CAL-COMBO-DESAYUNO", "Combo Desayuno Habitacion", "Desayuno sencillo con bebida fria para entrega en habitacion.", 22500, map[string]float64{"CAL-DESAYUNO": 1, "CAL-GASEOSA": 1}},
	}
	createdCombos := 0
	skippedCombos := 0
	for _, sc := range combos {
		exists, err := comboExists(db, *empresaID, sc.Codigo)
		must(err, "comboExists "+sc.Codigo)
		if exists {
			skippedCombos++
			continue
		}
		ingredientes := []dbpkg.ComboProductoDetalle{}
		for sku, qty := range sc.Ingredientes {
			p := prodMap[sku]
			ingredientes = append(ingredientes, dbpkg.ComboProductoDetalle{EmpresaID: *empresaID, ProductoID: p.ID, Cantidad: qty, UnidadMedida: "unidad", UsuarioCreador: *usuario, Estado: "activo", Observaciones: "receta seed motel calipso"})
		}
		_, err = dbpkg.CreateComboProducto(db, dbpkg.ComboProducto{EmpresaID: *empresaID, Codigo: sc.Codigo, Nombre: sc.Nombre, Descripcion: sc.Descripcion, UnidadMedida: "combo", Precio: sc.Precio, ImpuestoPorcentaje: 19, UsuarioCreador: *usuario, Estado: "activo", Observaciones: "seed profesional combos motel calipso"}, ingredientes)
		must(err, "CreateComboProducto "+sc.Codigo)
		createdCombos++
	}
	out := map[string]interface{}{"ok": true, "empresa_id": *empresaID, "bodega_id": bodegaID, "productos_creados": createdProducts, "combos_creados": createdCombos, "combos_existentes": skippedCombos}
	b, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(b))
}
