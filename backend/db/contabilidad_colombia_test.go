package db

import (
	"strings"
	"testing"
)

func TestValidateEmpresaContabilidadConfigProfessionalRules(t *testing.T) {
	valid := EmpresaContabilidadConfig{EmpresaID: 7, NombreSistema: "Contabilidad", Moneda: "COP", PeriodoActual: "2026-05"}
	if err := ValidateEmpresaContabilidadConfig(valid); err != nil {
		t.Fatalf("config valida rechazada: %v", err)
	}
	invalid := valid
	invalid.Moneda = "PESOS"
	if err := ValidateEmpresaContabilidadConfig(invalid); err == nil || !strings.Contains(err.Error(), "moneda") {
		t.Fatalf("moneda invalida no fue rechazada: %v", err)
	}
	invalid = valid
	invalid.PeriodoActual = "2026/05"
	if err := ValidateEmpresaContabilidadConfig(invalid); err == nil || !strings.Contains(err.Error(), "periodo") {
		t.Fatalf("periodo invalido no fue rechazado: %v", err)
	}
}

func TestValidateEmpresaContabilidadCuentaProfessionalRules(t *testing.T) {
	valid := EmpresaContabilidadCuenta{EmpresaID: 7, Codigo: "110505", Nombre: "Caja general", Naturaleza: "debito", TipoCuenta: "auxiliar", CuentaPadre: "1105", Estado: "activo"}
	if err := ValidateEmpresaContabilidadCuenta(valid); err != nil {
		t.Fatalf("cuenta valida rechazada: %v", err)
	}
	invalid := valid
	invalid.Codigo = "11A"
	if err := ValidateEmpresaContabilidadCuenta(invalid); err == nil || !strings.Contains(err.Error(), "PUC") {
		t.Fatalf("codigo invalido no fue rechazado: %v", err)
	}
	invalid = valid
	invalid.CuentaPadre = "110505"
	if err := ValidateEmpresaContabilidadCuenta(invalid); err == nil || !strings.Contains(err.Error(), "padre") {
		t.Fatalf("cuenta padre circular no fue rechazada: %v", err)
	}
}

func TestValidateEmpresaContabilidadComprobanteProfessionalRules(t *testing.T) {
	valid := EmpresaContabilidadComprobante{
		EmpresaID:        7,
		Concepto:         "Ajuste contable",
		FechaComprobante: "2026-05-07",
		PeriodoContable:  "2026-05",
		Estado:           "contabilizado",
		Lineas:           []EmpresaContabilidadAsientoLinea{{CuentaCodigo: "110505", Debito: 100}, {CuentaCodigo: "413595", Credito: 100}},
	}
	if err := ValidateEmpresaContabilidadComprobante(valid); err != nil {
		t.Fatalf("comprobante valido rechazado: %v", err)
	}
	invalid := valid
	invalid.PeriodoContable = "2026-04"
	if err := ValidateEmpresaContabilidadComprobante(invalid); err == nil || !strings.Contains(err.Error(), "coincide") {
		t.Fatalf("periodo que no coincide no fue rechazado: %v", err)
	}
	invalid = valid
	invalid.Estado = "cerrado"
	if err := ValidateEmpresaContabilidadComprobante(invalid); err == nil || !strings.Contains(err.Error(), "estado") {
		t.Fatalf("estado invalido no fue rechazado: %v", err)
	}
}

func TestValidateEmpresaContabilidadTerceroAndImpuesto(t *testing.T) {
	if err := ValidateEmpresaContabilidadTercero(EmpresaContabilidadTercero{EmpresaID: 7, Documento: "900123456", Nombre: "Proveedor SAS", Email: "conta@example.com", Estado: "activo"}); err != nil {
		t.Fatalf("tercero valido rechazado: %v", err)
	}
	if err := ValidateEmpresaContabilidadTercero(EmpresaContabilidadTercero{EmpresaID: 7, Documento: "900123456", Nombre: "Proveedor SAS", Email: "mal", Estado: "activo"}); err == nil {
		t.Fatalf("email invalido no fue rechazado")
	}
	if err := ValidateEmpresaContabilidadImpuesto(EmpresaContabilidadImpuesto{EmpresaID: 7, Codigo: "IVA19", Nombre: "IVA", Tipo: "iva", Porcentaje: 19, CuentaDebito: "240810", CuentaCredito: "240805", Estado: "activo"}); err != nil {
		t.Fatalf("impuesto valido rechazado: %v", err)
	}
	if err := ValidateEmpresaContabilidadImpuesto(EmpresaContabilidadImpuesto{EmpresaID: 7, Codigo: "IVA999", Nombre: "IVA", Tipo: "iva", Porcentaje: 120, Estado: "activo"}); err == nil {
		t.Fatalf("porcentaje invalido no fue rechazado")
	}
}
