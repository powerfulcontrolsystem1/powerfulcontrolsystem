package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStationStaffUISeparatesWaiterTipsAndCommissionist(t *testing.T) {
	files := map[string][]string{
		filepath.Join("..", "..", "web", "administrar_empresa", "configuracion_de_estaciones.html"): {
			"mesero_asignado", "comisionista_asignado", "mostrar_comisionista", "conservar_ultimo_comisionista",
		},
		filepath.Join("..", "..", "web", "administrar_empresa", "carrito_de_compras.html"): {
			"usuario_comisionista", "carrito.comisionista_ultimo", "rememberCurrentStationCommissionist", "Comisionista",
		},
		filepath.Join("..", "..", "web", "administrar_empresa", "propinas.html"): {
			"Propinas por mesero", "separado del documento fiscal",
		},
		filepath.Join("..", "..", "web", "administrar_empresa", "comisiones.html"): {
			"cfgIncluirProductos", "Incluir productos vendidos o aplicados", "usuario_comisionista",
		},
	}
	for path, markers := range files {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, marker := range markers {
			if !strings.Contains(string(raw), marker) {
				t.Fatalf("%s no contiene %q", path, marker)
			}
		}
	}
}

func TestPaymentUsesServerScopedStationStaff(t *testing.T) {
	raw, err := os.ReadFile("carritos_compras.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, marker := range []string{
		"loadCarritoStationStaffConfig", "ResolveEmpresaUsuarioByReference", "meseroOperacion",
		"comisionistaOperacion", "carrito.comisionista_ultimo", "empresaID, estacionID",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("payment station staff contract is missing %q", marker)
		}
	}
}
