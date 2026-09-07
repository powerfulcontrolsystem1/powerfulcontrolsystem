package handlers

import "testing"

func TestConfiguracionGuiadaNoAutoOpenWhenUserHidIt(t *testing.T) {
	state := &empresaConfiguracionGuiadaState{
		EmpresaID:           12,
		ConfiguradaAnterior: true,
		ResumenAnterior: map[string]interface{}{
			"estado":         "no_mostrar_mas",
			"no_mostrar_mas": true,
		},
	}

	if !isEmpresaConfiguracionGuiadaHidden(state) {
		t.Fatal("la configuracion guiada debe quedar oculta cuando la empresa guardo no_mostrar_mas")
	}
	if shouldAutoOpenEmpresaConfiguracionGuiada(state) {
		t.Fatal("la configuracion guiada no debe autoabrirse si la empresa guardo no_mostrar_mas")
	}
}

func TestConfiguracionGuiadaNoAutoOpenWhenPostponed(t *testing.T) {
	state := &empresaConfiguracionGuiadaState{
		EmpresaID:           12,
		ConfiguradaAnterior: true,
		ResumenAnterior: map[string]interface{}{
			"estado":    "pospuesta",
			"pospuesta": "true",
		},
	}

	if !isEmpresaConfiguracionGuiadaHidden(state) {
		t.Fatal("la configuracion guiada debe quedar oculta cuando esta pospuesta")
	}
	if shouldAutoOpenEmpresaConfiguracionGuiada(state) {
		t.Fatal("la configuracion guiada no debe autoabrirse si esta pospuesta")
	}
}

func TestConfiguracionGuiadaAutoOpenForNewCompanyWithoutSummary(t *testing.T) {
	state := &empresaConfiguracionGuiadaState{EmpresaID: 12}

	if isEmpresaConfiguracionGuiadaHidden(state) {
		t.Fatal("una empresa sin resumen previo no debe tratarse como oculta")
	}
	if !shouldAutoOpenEmpresaConfiguracionGuiada(state) {
		t.Fatal("una empresa nueva sin resumen previo debe autoabrir la configuracion guiada")
	}
}
