package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmpresaRoutesUsePermissionWrappers(t *testing.T) {
	files := []string{"main.go"}
	walkErr := filepath.WalkDir("handlers", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		files = append(files, filepath.ToSlash(path))
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk handlers files: %v", walkErr)
	}

	allowedWrappers := []string{
		"WithEmpresaVentasPermissions(",
		"WithEmpresaInventarioPermissions(",
		"WithEmpresaFinanzasPermissions(",
		"WithEmpresaContabilidadColombiaPermissions(",
		"WithEmpresaContabilidadColombiaAvanzadaPermissions(",
		"WithEmpresaActivosFijosNIIFPermissions(",
		"WithEmpresaDeclaracionesTributariasPermissions(",
		"WithEmpresaCentrosCostoPermissions(",
		"WithEmpresaCierreFiscalPermissions(",
		"WithEmpresaClientesPermissions(",
		"WithEmpresaCRMUnificadoPermissions(",
		"WithEmpresaComprasPermissions(",
		"WithEmpresaSoportesComprasIAPermissions(",
		"WithEmpresaFacturacionPermissions(",
		"WithEmpresaFacturacionEcuadorPermissions(",
		"WithEmpresaFacturacionPanamaPermissions(",
		"WithEmpresaSeguridadPermissions(",
		"WithEmpresaVentaPublicaPermissions(",
		"WithEmpresaReservasHotelPermissions(",
		"WithEmpresaChatTareasPermissions(",
		"WithEmpresaGimnasioPermissions(",
		"WithEmpresaTaxiSystemPermissions(",
		"WithEmpresaDomiciliosPermissions(",
		"WithEmpresaParqueaderoPermissions(",
		"WithEmpresaApartamentosTuristicosPermissions(",
		"WithEmpresaPropiedadHorizontalPermissions(",
		"WithEmpresaAlquileresPermissions(",
		"WithEmpresaOdontologiaPermissions(",
		"WithEmpresaTurnosAtencionPermissions(",
		"WithEmpresaControlElectricoPermissions(",
		"WithEmpresaEnergiaSolarPermissions(",
		"WithEmpresaGrafologiaPermissions(",
		"WithEmpresaCarnetsPermissions(",
		"WithEmpresaHorariosTrabajadoresPermissions(",
		"WithEmpresaAsistenciaEmpleadosPermissions(",
		"WithEmpresaVehiculosRegistroPermissions(",
		"WithEmpresaHojaVidaOperativaPermissions(",
		"WithEmpresaUbicacionGPSPermissions(",
		"WithEmpresaProduccionMRPPermissions(",
		"WithEmpresaWMSPermissions(",
		"WithEmpresaTesoreriaPresupuestoPermissions(",
		"WithEmpresaNominaSueldosPermissions(",
		"WithEmpresaImportacionesCosteoPermissions(",
		"WithEmpresaAIUConstruccionPermissions(",
		"WithEmpresaCobranzaPermissions(",
		"WithEmpresaReportesPermissions(",
		"WithEmpresaPortalContadorPermissions(",
		"WithEmpresaPortalTercerosPermissions(",
		"WithEmpresaBancosPagosPermissions(",
		"WithEmpresaGestionDocumentalPermissions(",
		"WithEmpresaCumplimientoKYCPermissions(",
		"WithEmpresaContratosObligacionesPermissions(",
		"WithEmpresaCalidadProcesosPermissions(",
		"WithEmpresaDrogueriaFarmaciaPermissions(",
		"WithEmpresaModuloVerticalPermissions(",
		"WithEmpresaAuditoriaPermissions(",
		"WithEmpresaBackupsPermissions(",
		"WithEmpresaDocumentosOnlyOfficePermissions(",
		"WithEmpresaPublicScope(",
		"WithEmpresaSelfServicePermissions(",
	}
	allowedPublicPaths := map[string]bool{
		"/api/empresa/usuarios/login":                           true,
		"/api/empresa/usuarios/establecer_password":             true,
		"/api/empresa/usuarios/recuperar_invitacion":            true,
		"/api/empresa/usuarios/solicitar_recuperacion_password": true,
		"/api/empresa/usuarios/restablecer_password":            true,
		"/api/empresa/usuarios/cambiar_password":                true,
		"/api/public/certificados_tributarios":                  true,
	}

	violations := make([]string, 0)
	for _, filePath := range files {
		raw, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("read %s: %v", filePath, err)
		}
		lines := strings.Split(string(raw), "\n")
		for i, line := range lines {
			if !strings.Contains(line, "http.HandleFunc(\"/api/empresa/") {
				continue
			}

			if !containsAny(line, allowedWrappers) {
				violations = append(violations, fmt.Sprintf("%s:%d route registration without permission wrapper: %s", filepath.ToSlash(filePath), i+1, strings.TrimSpace(line)))
				continue
			}

			routePath := extractEmpresaRoutePath(line)
			if strings.Contains(line, "WithEmpresaPublicScope(") && !allowedPublicPaths[routePath] {
				violations = append(violations, fmt.Sprintf("%s:%d unexpected public scope for route %s", filepath.ToSlash(filePath), i+1, routePath))
			}
		}
	}

	if len(violations) > 0 {
		t.Fatalf("empresa routes wrapper policy violations:\n%s", strings.Join(violations, "\n"))
	}
}

func containsAny(line string, values []string) bool {
	for _, value := range values {
		if strings.Contains(line, value) {
			return true
		}
	}
	return false
}

func extractEmpresaRoutePath(line string) string {
	const token = "http.HandleFunc(\""
	start := strings.Index(line, token)
	if start < 0 {
		return ""
	}
	rest := line[start+len(token):]
	end := strings.Index(rest, "\"")
	if end < 0 {
		return ""
	}
	return strings.TrimSpace(rest[:end])
}
