package db

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNormalizeInventarioTendenciaFechaPostgresYCompat(t *testing.T) {
	want := "2026-08-26"
	values := []interface{}{
		time.Date(2026, time.August, 26, 0, 0, 0, 0, time.UTC),
		"2026-08-26T00:00:00Z",
		"2026-08-26 00:00:00+00",
		[]byte("2026-08-26"),
	}
	for _, value := range values {
		got, err := normalizeInventarioTendenciaFecha(value)
		if err != nil {
			t.Fatalf("normalize %T: %v", value, err)
		}
		if got != want {
			t.Fatalf("normalize %T=%q, want %q", value, got, want)
		}
	}
	if _, err := normalizeInventarioTendenciaFecha("fecha-invalida"); err == nil {
		t.Fatal("invalid daily inventory date must fail closed")
	}
}

func TestCreateBodegaUsaTimestampPostgres(t *testing.T) {
	raw, err := os.ReadFile("productos.go")
	if err != nil {
		t.Fatalf("read productos.go: %v", err)
	}
	src := string(raw)
	start := strings.Index(src, "func CreateBodega(")
	if start < 0 {
		t.Fatal("no se encontro CreateBodega")
	}
	end := strings.Index(src[start:], "// GetBodegasByEmpresa")
	if end < 0 {
		t.Fatal("no se encontro limite de CreateBodega")
	}
	body := src[start : start+end]
	if strings.Contains(body, "pcs_ts(") {
		t.Fatalf("CreateBodega no debe usar pcs_ts() en runtime PostgreSQL: %s", body)
	}
	if !strings.Contains(body, "sqlNowExpr()") {
		t.Fatalf("CreateBodega debe usar sqlNowExpr() para fecha_creacion/fecha_actualizacion: %s", body)
	}
}

func TestEnsureEmpresaBodega1DefaultEsIdempotenteYSinStockDemo(t *testing.T) {
	raw, err := os.ReadFile("productos.go")
	if err != nil {
		t.Fatalf("read productos.go: %v", err)
	}
	src := string(raw)
	start := strings.Index(src, "func EnsureEmpresaBodega1(")
	if start < 0 {
		t.Fatal("no se encontro EnsureEmpresaBodega1")
	}
	end := strings.Index(src[start:], "// GetBodegasByEmpresa")
	if end < 0 {
		t.Fatal("no se encontro limite de EnsureEmpresaBodega1")
	}
	body := src[start : start+end]
	for _, required := range []string{`"Bodega 1"`, "getEmpresaBodegaIDByNombre", "SetBodegaEstado", "CreateBodega"} {
		if !strings.Contains(body, required) {
			t.Fatalf("EnsureEmpresaBodega1 debe conservar %s para idempotencia: %s", required, body)
		}
	}
	for _, forbidden := range []string{"inventario_existencias", "CreateProducto", "InsertProducto", "upsertExistencia"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("EnsureEmpresaBodega1 no debe crear stock/productos demo; encontro %s en: %s", forbidden, body)
		}
	}
	if !strings.Contains(src, "func ApplyDefaultBodega1ToExistingEmpresas(") {
		t.Fatal("debe existir backfill independiente para empresas existentes")
	}
	if !strings.Contains(src, "20260607_bodega_1_default") {
		t.Fatal("el backfill debe tener version nueva para ejecutarse aunque migraciones anteriores ya esten aplicadas")
	}
}

func TestTransferirProductoEntreBodegasUsaSQLCompatPostgres(t *testing.T) {
	raw, err := os.ReadFile("productos.go")
	if err != nil {
		t.Fatalf("read productos.go: %v", err)
	}
	src := string(raw)
	start := strings.Index(src, "func TransferirProductoEntreBodegas(")
	if start < 0 {
		t.Fatal("no se encontro TransferirProductoEntreBodegas")
	}
	end := strings.Index(src[start:], "// GetMovimientosByEmpresa")
	if end < 0 {
		t.Fatal("no se encontro limite de TransferirProductoEntreBodegas")
	}
	body := src[start : start+end]
	for _, forbidden := range []string{"tx.QueryRow(", "tx.Exec("} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("TransferirProductoEntreBodegas debe usar wrappers SQLCompat en runtime PostgreSQL; encontro %s en: %s", forbidden, body)
		}
	}
	for _, required := range []string{"queryRowTxSQLCompat", "decrementarExistenciaTx", "insertMovimientoTx"} {
		if !strings.Contains(body, required) {
			t.Fatalf("TransferirProductoEntreBodegas debe conservar %s en la transaccion: %s", required, body)
		}
	}
	helperStart := strings.Index(src, "func decrementarExistenciaTx(")
	if helperStart < 0 {
		t.Fatal("no se encontro decrementarExistenciaTx")
	}
	helperEnd := strings.Index(src[helperStart:], "\nfunc ")
	if helperEnd < 0 || !strings.Contains(src[helperStart:helperStart+helperEnd], "execTxSQLCompat") {
		t.Fatal("decrementarExistenciaTx debe conservar SQLCompat para PostgreSQL")
	}
}

func TestProveedorCodigoOpcionalNoBloqueaMultiplesProveedores(t *testing.T) {
	raw, err := os.ReadFile("productos.go")
	if err != nil {
		t.Fatalf("read productos.go: %v", err)
	}
	src := string(raw)

	createStart := strings.Index(src, "func CreateProveedor(")
	if createStart < 0 {
		t.Fatal("no se encontro CreateProveedor")
	}
	createEnd := strings.Index(src[createStart:], "// GetProveedoresByEmpresa")
	if createEnd < 0 {
		t.Fatal("no se encontro limite de CreateProveedor")
	}
	createBody := src[createStart : createStart+createEnd]
	if !strings.Contains(createBody, "VALUES (?, NULLIF(?, ''), ?") {
		t.Fatalf("CreateProveedor debe persistir codigo vacio como NULL para no colisionar con el indice unico: %s", createBody)
	}

	updateStart := strings.Index(src, "func UpdateProveedor(")
	if updateStart < 0 {
		t.Fatal("no se encontro UpdateProveedor")
	}
	updateEnd := strings.Index(src[updateStart:], "func validateProveedorCondiciones")
	if updateEnd < 0 {
		t.Fatal("no se encontro limite de UpdateProveedor")
	}
	updateBody := src[updateStart : updateStart+updateEnd]
	if !strings.Contains(updateBody, "SET codigo = NULLIF(?, '')") {
		t.Fatalf("UpdateProveedor debe normalizar codigo vacio a NULL: %s", updateBody)
	}
}

func TestServicioCodigoOpcionalYValoresProductivos(t *testing.T) {
	raw, err := os.ReadFile("productos.go")
	if err != nil {
		t.Fatalf("read productos.go: %v", err)
	}
	src := string(raw)

	createStart := strings.Index(src, "func CreateServicio(")
	createEnd := strings.Index(src[createStart:], "func validateServicioPayload(")
	if createStart < 0 || createEnd < 0 {
		t.Fatal("no se encontro el contrato de CreateServicio")
	}
	if body := src[createStart : createStart+createEnd]; !strings.Contains(body, "VALUES (?, NULLIF(?, ''), ?") {
		t.Fatalf("CreateServicio debe persistir codigo vacio como NULL: %s", body)
	}

	updateStart := strings.Index(src, "func UpdateServicio(")
	updateEnd := strings.Index(src[updateStart:], "// DeleteServicio")
	if updateStart < 0 || updateEnd < 0 {
		t.Fatal("no se encontro el contrato de UpdateServicio")
	}
	if body := src[updateStart : updateStart+updateEnd]; !strings.Contains(body, "SET codigo = NULLIF(?, '')") {
		t.Fatalf("UpdateServicio debe normalizar codigo vacio a NULL: %s", body)
	}

	valid := Servicio{EmpresaID: 12, Nombre: "Servicio QA", DuracionMinutos: 30, CostoReferencial: 1000, Precio: 2000, ImpuestoPorcentaje: 19}
	if err := validateServicioPayload(valid); err != nil {
		t.Fatalf("servicio valido rechazado: %v", err)
	}
	invalid := []Servicio{
		{EmpresaID: 12, Nombre: "", Precio: 100},
		{EmpresaID: 12, Nombre: "Duracion", DuracionMinutos: -1},
		{EmpresaID: 12, Nombre: "Costo", CostoReferencial: -1},
		{EmpresaID: 12, Nombre: "Precio", Precio: -1},
		{EmpresaID: 12, Nombre: "Impuesto", ImpuestoPorcentaje: 101},
	}
	for _, candidate := range invalid {
		if err := validateServicioPayload(candidate); !errors.Is(err, ErrProductosDatosInvalidos) {
			t.Errorf("servicio invalido no produjo error publico tipado: %+v err=%v", candidate, err)
		}
	}
}

func TestProductoRechazaValoresEconomicosYStockInvalidos(t *testing.T) {
	valid := Producto{Nombre: "Producto QA", Costo: 100, Precio: 200, ImpuestoPorcentaje: 19, StockMinimo: 1, StockMaximo: 10}
	if err := validateProductoValores(valid, 2); err != nil {
		t.Fatalf("producto valido rechazado: %v", err)
	}
	tests := []struct {
		name  string
		value Producto
		stock float64
	}{
		{name: "sin nombre", value: Producto{Precio: 1}},
		{name: "costo negativo", value: Producto{Nombre: "P", Costo: -1}},
		{name: "precio negativo", value: Producto{Nombre: "P", Precio: -1}},
		{name: "impuesto mayor a cien", value: Producto{Nombre: "P", ImpuestoPorcentaje: 101}},
		{name: "stock inicial negativo", value: Producto{Nombre: "P"}, stock: -1},
		{name: "umbrales invertidos", value: Producto{Nombre: "P", StockMinimo: 5, StockMaximo: 2}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := validateProductoValores(test.value, test.stock); !errors.Is(err, ErrProductosDatosInvalidos) {
				t.Fatalf("error=%v, want ErrProductosDatosInvalidos", err)
			}
		})
	}
}

func TestProductosInventarioEImpresorasNoDescartanRowsAffected(t *testing.T) {
	for _, path := range []string{"productos.go", "inventario_avanzado.go", "empresa_impresoras.go"} {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, forbidden := range []string{
			`affected, _ := res.RowsAffected()`,
			`affected, _ = res.RowsAffected()`,
			`if affected, _ := res.RowsAffected()`,
		} {
			if strings.Contains(string(raw), forbidden) {
				t.Errorf("%s descarta un error de RowsAffected mediante %q", path, forbidden)
			}
		}
	}
}

func TestReferenciasInventarioAjenasUsanErrorSeguroTipado(t *testing.T) {
	raw, err := os.ReadFile("productos.go")
	if err != nil {
		t.Fatalf("read productos.go: %v", err)
	}
	src := string(raw)
	for _, required := range []string{
		`ErrInventarioEntidadNoDisponible = errors.New("entidad de inventario no disponible para la empresa")`,
		`fmt.Errorf("%w: producto", ErrInventarioEntidadNoDisponible)`,
		`fmt.Errorf("%w: bodega", ErrInventarioEntidadNoDisponible)`,
		`fmt.Errorf("%w: proveedor", ErrInventarioEntidadNoDisponible)`,
		`fmt.Errorf("%w: categoria", ErrInventarioEntidadNoDisponible)`,
		`fmt.Errorf("%w: servicio", ErrInventarioEntidadNoDisponible)`,
	} {
		if !strings.Contains(src, required) {
			t.Errorf("productos.go debe conservar el contrato de ownership seguro %q", required)
		}
	}
	for _, leaked := range []string{
		`fmt.Errorf("producto %d no pertenece a la empresa %d"`,
		`fmt.Errorf("bodega %d no pertenece a la empresa %d"`,
		`fmt.Errorf("proveedor %d no pertenece a la empresa %d"`,
		`fmt.Errorf("categoria %d no pertenece a la empresa %d"`,
		`fmt.Errorf("servicio %d no pertenece a la empresa %d"`,
	} {
		if strings.Contains(src, leaked) {
			t.Errorf("productos.go no debe filtrar IDs multiempresa mediante %q", leaked)
		}
	}
}
