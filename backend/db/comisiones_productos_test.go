package db

import "testing"

func TestItemCumpleBaseComisionSeparatesServicesAndProducts(t *testing.T) {
	service := comisionServicioItemSnapshot{TipoItem: "servicio", ServicioNombre: "Corte de cabello"}
	product := comisionServicioItemSnapshot{TipoItem: "producto", Descripcion: "Tratamiento capilar"}

	onlyServices := EmpresaComisionesServicioConfiguracion{FiltroServicio: "corte", IncluirProductos: false}
	if !itemCumpleBaseComision(service, onlyServices) {
		t.Fatal("matching service must be commissionable")
	}
	if itemCumpleBaseComision(product, onlyServices) {
		t.Fatal("product must be excluded when incluir_productos is disabled")
	}

	withProducts := onlyServices
	withProducts.IncluirProductos = true
	if !itemCumpleBaseComision(product, withProducts) {
		t.Fatal("product must be included when incluir_productos is enabled")
	}
}
