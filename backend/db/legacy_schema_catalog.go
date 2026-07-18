package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

const legacySchemaBaselineVersion = "20260716-000-legacy-schema-v1"

type legacySchemaStep struct {
	Name  string
	Apply func(*sql.DB) error
}

// legacyEmpresaSchemaCatalog is the reviewed compatibility bridge for modules
// that still expose Ensure* functions. It runs only in the migration process;
// API and worker DDL is blocked by runtime_schema_guard.go.
var legacyEmpresaSchemaCatalog = []legacySchemaStep{
	{"EnsureEmpresaUsuariosAuthSchema", EnsureEmpresaUsuariosAuthSchema},
	{"EnsureEmpresaBuzonSchema", EnsureEmpresaBuzonSchema},
	{"EnsureEmpresaCarritosSchema", EnsureEmpresaCarritosSchema},
	{"EnsureEmpresaClientesSchema", EnsureEmpresaClientesSchema},
	{"EnsureEmpresaProductosSchema", EnsureEmpresaProductosSchema},
	{"EnsureEmpresaFinanzasSchema", EnsureEmpresaFinanzasSchema},
	{"EnsureEmpresaEventosContablesSchema", EnsureEmpresaEventosContablesSchema},
	{"EnsureEmpresaFacturacionElectronicaSchema", EnsureEmpresaFacturacionElectronicaSchema},
	{"EnsureEmpresaAuditoriaSchema", EnsureEmpresaAuditoriaSchema},
	{"EnsureEmpresaConfiguracionGeneralSchema", EnsureEmpresaConfiguracionGeneralSchema},
	{"EnsureEmpresaConfiguracionOperativaSchema", EnsureEmpresaConfiguracionOperativaSchema},
	{"EnsureEmpresaConfiguracionAvanzadaSchema", EnsureEmpresaConfiguracionAvanzadaSchema},
	{"EnsureEmpresaEstacionPrefsSchema", EnsureEmpresaEstacionPrefsSchema},
	{"EnsureEmpresaEstacionColumnPreferencesSchema", EnsureEmpresaEstacionColumnPreferencesSchema},
	{"EnsureEmpresaEstacionAseoSchema", EnsureEmpresaEstacionAseoSchema},
	{"EnsureEmpresaDatafonosSchema", EnsureEmpresaDatafonosSchema},
	{"EnsureEmpresasComprasSchema", EnsureEmpresasComprasSchema},
	{"EnsureEmpresaComprasAvanzadasSchema", EnsureEmpresaComprasAvanzadasSchema},
	{"EnsureEmpresaInventarioAvanzadoSchema", EnsureEmpresaInventarioAvanzadoSchema},
	{"EnsureEmpresaCodigosDescuentoSchema", EnsureEmpresaCodigosDescuentoSchema},
	{"EnsureEmpresaCreditosSchema", EnsureEmpresaCreditosSchema},
	{"EnsureEmpresaCobranzaSchema", EnsureEmpresaCobranzaSchema},
	{"EnsureEmpresaPropinasSchema", EnsureEmpresaPropinasSchema},
	{"EnsureEmpresaComisionesServicioSchema", EnsureEmpresaComisionesServicioSchema},
	{"EnsureEmpresaCorteCajaConfiguracionSchema", EnsureEmpresaCorteCajaConfiguracionSchema},
	{"EnsureEmpresaImpresorasSchema", EnsureEmpresaImpresorasSchema},
	{"EnsureEmpresaVentasOfflineSchema", EnsureEmpresaVentasOfflineSchema},
	{"EnsureEmpresaDocumentosTransaccionalesSchema", EnsureEmpresaDocumentosTransaccionalesSchema},
	{"EnsureEmpresaBackupsSchema", EnsureEmpresaBackupsSchema},
	{"EnsureEmpresaReportesProgramacionSchema", EnsureEmpresaReportesProgramacionSchema},
	{"EnsureEmpresaAIOpenAIProviderSchema", EnsureEmpresaAIOpenAIProviderSchema},
	{"EnsureEmpresaAIEnterpriseSchema", EnsureEmpresaAIEnterpriseSchema},
	{"EnsureEmpresaAIChatSchema", EnsureEmpresaAIChatSchema},
	{"EnsureEmpresaChatTareasSchema", EnsureEmpresaChatTareasSchema},
	{"EnsureEmpresaAgentesUsoSchema", EnsureEmpresaAgentesUsoSchema},
	{"EnsureEmpresaImpuestosSchema", EnsureEmpresaImpuestosSchema},
	{"EnsureEmpresaNominaSchema", EnsureEmpresaNominaSchema},
	{"EnsureEmpresaNominaColombiaAvanzadaSchema", EnsureEmpresaNominaColombiaAvanzadaSchema},
	{"EnsureEmpresaContabilidadColombiaSchema", EnsureEmpresaContabilidadColombiaSchema},
	{"EnsureEmpresaContabilidadColombiaAvanzadaSchema", EnsureEmpresaContabilidadColombiaAvanzadaSchema},
	{"EnsureEmpresaCentrosCostoSchema", EnsureEmpresaCentrosCostoSchema},
	{"EnsureEmpresaCierreFiscalSchema", EnsureEmpresaCierreFiscalSchema},
	{"EnsureEmpresaDeclaracionesTributariasSchema", EnsureEmpresaDeclaracionesTributariasSchema},
	{"EnsureEmpresaTesoreriaPresupuestoSchema", EnsureEmpresaTesoreriaPresupuestoSchema},
	{"EnsureEmpresaImportacionesCosteoSchema", EnsureEmpresaImportacionesCosteoSchema},
	{"EnsureEmpresaAIUConstruccionSchema", EnsureEmpresaAIUConstruccionSchema},
	{"EnsureCatalogoLegalPaisSchema", EnsureCatalogoLegalPaisSchema},
	{"EnsureEmpresaModulosColombiaSchema", EnsureEmpresaModulosColombiaSchema},
	{"EnsureEmpresaModulosFaltantesSchema", EnsureEmpresaModulosFaltantesSchema},
	{"EnsureEmpresaPortalContadorSchema", EnsureEmpresaPortalContadorSchema},
	{"EnsureEmpresaPortalTercerosCertificadosSchema", EnsureEmpresaPortalTercerosCertificadosSchema},
	{"EnsureEmpresaSoportesComprasIASchema", EnsureEmpresaSoportesComprasIASchema},
	{"EnsureEmpresaReservasHotelSchema", EnsureEmpresaReservasHotelSchema},
	{"EnsureEmpresaTarifasMotelSchema", EnsureEmpresaTarifasMotelSchema},
	{"EnsureEmpresaTarifasPorMinutosSchema", EnsureEmpresaTarifasPorMinutosSchema},
	{"EnsureEmpresaTarifasPorMinutosConfiguracionSchema", EnsureEmpresaTarifasPorMinutosConfiguracionSchema},
	{"EnsureEmpresaTarifasPorDiaSchema", EnsureEmpresaTarifasPorDiaSchema},
	{"EnsureHotelTarjetasAccesoSchema", EnsureHotelTarjetasAccesoSchema},
	{"EnsureEmpresaAlquileresSchema", EnsureEmpresaAlquileresSchema},
	{"EnsureEmpresaApartamentosTuristicosSchema", EnsureEmpresaApartamentosTuristicosSchema},
	{"EnsureEmpresaParqueaderoSchema", EnsureEmpresaParqueaderoSchema},
	{"EnsureEmpresaPropiedadHorizontalSchema", EnsureEmpresaPropiedadHorizontalSchema},
	{"EnsureEmpresaDomiciliosSchema", EnsureEmpresaDomiciliosSchema},
	{"EnsureEmpresaGimnasioSchema", EnsureEmpresaGimnasioSchema},
	{"EnsureEmpresaOdontologiaSchema", EnsureEmpresaOdontologiaSchema},
	{"EnsureEmpresaTaxiSystemSchema", EnsureEmpresaTaxiSystemSchema},
	{"EnsureEmpresaTurnosAtencionSchema", EnsureEmpresaTurnosAtencionSchema},
	{"EnsureEmpresaAsistenciaSchema", EnsureEmpresaAsistenciaSchema},
	{"EnsureEmpresaHojaVidaOperativaSchema", EnsureEmpresaHojaVidaOperativaSchema},
	{"EnsureEmpresaVehiculosRegistroSchema", EnsureEmpresaVehiculosRegistroSchema},
	{"EnsureEmpresaUbicacionGPSSchema", EnsureEmpresaUbicacionGPSSchema},
	{"EnsureEmpresaSensorPuertasSchema", EnsureEmpresaSensorPuertasSchema},
	{"EnsureEmpresaControlElectricoSchema", EnsureEmpresaControlElectricoSchema},
	{"EnsureEmpresaEnergiaSolarSchema", EnsureEmpresaEnergiaSolarSchema},
	{"EnsureEmpresaCamarasSchema", EnsureEmpresaCamarasSchema},
	{"EnsureEmpresaGrafologiaSchema", EnsureEmpresaGrafologiaSchema},
	{"EnsureEmpresaCarnetsSchema", EnsureEmpresaCarnetsSchema},
	{"EnsureEmpresaProduccionMRPSchema", EnsureEmpresaProduccionMRPSchema},
	{"EnsureEmpresaWMSSchema", EnsureEmpresaWMSSchema},
	{"EnsureEmpresaCRMVentasAvanzadasSchema", EnsureEmpresaCRMVentasAvanzadasSchema},
	{"EnsureEmpresaSoporteRemotoSchema", EnsureEmpresaSoporteRemotoSchema},
	{"EnsureEmpresaRappiSchema", EnsureEmpresaRappiSchema},
	{"EnsureEmpresaPublicacionesRedSocialSchema", EnsureEmpresaPublicacionesRedSocialSchema},
	{"EnsureEmpresaRedSocialInteraccionesSchema", EnsureEmpresaRedSocialInteraccionesSchema},
	{"EnsureEmpresaVentaPublicaSchema", EnsureEmpresaVentaPublicaSchema},
	{"EnsureVentaPublicaSchema", EnsureVentaPublicaSchema},
	{"EnsureEmpresaNextcloudSchema", EnsureEmpresaNextcloudSchema},
	{"EnsureEmpresasScopeReferences", EnsureEmpresasScopeReferences},
	{"EnsureEmpresaPermisosFinosSchema", EnsureEmpresaPermisosFinosSchema},
	{"EnsureEstacionVIPCodigosSchema", EnsureEstacionVIPCodigosSchema},
	{"EnsureEmpresaCalculadoraSchema", EnsureEmpresaCalculadoraSchema},
}

var legacySuperSchemaCatalog = []legacySchemaStep{
	{"EnsureAdministradoresAuthSchema", EnsureAdministradoresAuthSchema},
	{"EnsurePaymentGatewaySchema", EnsurePaymentGatewaySchema},
	{"EnsureLicenciasSchema", EnsureLicenciasSchema},
	{"EnsureLicenciasGratisActivacionesSchema", EnsureLicenciasGratisActivacionesSchema},
	{"EnsureEmpresaLicenciasAdicionalesSchema", EnsureEmpresaLicenciasAdicionalesSchema},
	{"EnsureLicenciaEmpresaRetencionSchema", EnsureLicenciaEmpresaRetencionSchema},
	{"EnsureLicenciaVencimientoNotificacionesSchema", EnsureLicenciaVencimientoNotificacionesSchema},
	{"EnsureSuperAuditoriaSchema", EnsureSuperAuditoriaSchema},
	{"EnsureSuperAIChatSchema", EnsureSuperAIChatSchema},
	{"EnsureSuperContractSchema", EnsureSuperContractSchema},
	{"EnsureDefaultSuperContract", EnsureDefaultSuperContract},
	{"EnsureSuperCorreoNotificacionesPruebaSchema", EnsureSuperCorreoNotificacionesPruebaSchema},
	{"EnsureSuperAlertasSchema", EnsureSuperAlertasSchema},
	{"EnsureSuperErroresSistemaSchema", EnsureSuperErroresSistemaSchema},
	{"EnsureSuperCorreosMasivosSchema", EnsureSuperCorreosMasivosSchema},
	{"EnsureSuperMantenimientoAgentesSchema", EnsureSuperMantenimientoAgentesSchema},
	{"EnsureSuperServidorEventosSchema", EnsureSuperServidorEventosSchema},
	{"EnsureSuperVPSSnapshotSchema", EnsureSuperVPSSnapshotSchema},
	{"EnsureEmpresaEmailCorporativoSchema", EnsureEmpresaEmailCorporativoSchema},
	{"EnsureUsuarioConfiguracionSchema", EnsureUsuarioConfiguracionSchema},
	{"EnsureAdminPrincipalDelegacionesSchema", EnsureAdminPrincipalDelegacionesSchema},
	{"EnsureAdminEmpresaCompartidaSchema", EnsureAdminEmpresaCompartidaSchema},
	{"EnsureAsesorComercialSchema", EnsureAsesorComercialSchema},
	{"EnsureRolesDeUsuarioSchema", EnsureRolesDeUsuarioSchema},
	{"EnsureRolesPermisosSchema", EnsureRolesPermisosSchema},
	{"EnsureTipoEmpresaPreconfiguracionSchema", EnsureTipoEmpresaPreconfiguracionSchema},
	{"EnsureCanonicalTiposEmpresaPreconfigurables", EnsureCanonicalTiposEmpresaPreconfigurables},
	{"EnsureEnergiaSolarInTipoEmpresaPreconfiguraciones", EnsureEnergiaSolarInTipoEmpresaPreconfiguraciones},
	{"EnsureAyudaTicketsSchema", EnsureAyudaTicketsSchema},
	{"EnsureSuperVentaDigitalSchema", EnsureSuperVentaDigitalSchema},
	{"InitMetricsTable", InitMetricsTable},
}

func ApplyLegacySchemaCatalog(dbEmpresas, dbSuper *sql.DB) error {
	if dbEmpresas == nil || dbSuper == nil {
		return fmt.Errorf("legacy schema catalog requires both databases")
	}
	// A legacy installation can have the migration ledger without the fields
	// introduced by the immutable migrator. Normalize that ledger before the
	// read-only baseline check below; this function is called only by
	// PCS_RUNTIME_ROLE=migrate, never by API or worker processes.
	for _, target := range []struct {
		name string
		db   *sql.DB
	}{{"empresas", dbEmpresas}, {"superadministrador", dbSuper}} {
		if err := EnsureSchemaMigrationsTable(target.db); err != nil {
			return fmt.Errorf("prepare %s migration ledger: %w", target.name, err)
		}
	}
	if err := ValidateLegacySchemaCatalogManifest(); err != nil {
		return err
	}
	for _, target := range []struct {
		name  string
		db    *sql.DB
		steps []legacySchemaStep
	}{{"empresas", dbEmpresas, legacyEmpresaSchemaCatalog}, {"superadministrador", dbSuper, legacySuperSchemaCatalog}} {
		applied, err := LegacySchemaBaselineApplied(context.Background(), target.db)
		if err != nil {
			return fmt.Errorf("verify %s legacy baseline: %w", target.name, err)
		}
		if applied {
			continue
		}
		for _, step := range target.steps {
			if step.Apply == nil {
				return fmt.Errorf("%s migration step %s is nil", target.name, step.Name)
			}
			if err := step.Apply(target.db); err != nil {
				return fmt.Errorf("%s migration step %s: %w", target.name, step.Name, err)
			}
		}
	}
	return nil
}

// ValidateLegacySchemaCatalogManifest ensures every compatibility step belongs
// to the generated, source-fingerprinted manifest. It is intentionally a
// fail-closed guard: legacy Ensure* functions are frozen once their manifest
// migration is released; later schema changes belong in new migrations.
func ValidateLegacySchemaCatalogManifest() error {
	if strings.TrimSpace(legacySchemaCatalogSourceFingerprint) == "" {
		return fmt.Errorf("legacy schema catalog fingerprint is empty")
	}
	seen := make(map[string]struct{}, len(legacyEmpresaSchemaCatalog)+len(legacySuperSchemaCatalog))
	for target, steps := range map[string][]legacySchemaStep{
		MigrationTargetEmpresas: legacyEmpresaSchemaCatalog,
		MigrationTargetSuper:    legacySuperSchemaCatalog,
	} {
		for _, step := range steps {
			name := strings.TrimSpace(step.Name)
			if name == "" || step.Apply == nil {
				return fmt.Errorf("legacy %s catalog contains an invalid step", target)
			}
			if _, duplicate := seen[name]; duplicate {
				return fmt.Errorf("legacy schema catalog repeats step %s", name)
			}
			if strings.TrimSpace(legacySchemaCatalogStepSourceFingerprints[name]) == "" {
				return fmt.Errorf("legacy %s step %s is missing its source fingerprint", target, name)
			}
			seen[name] = struct{}{}
		}
	}
	if len(seen) != len(legacySchemaCatalogStepSourceFingerprints) {
		return fmt.Errorf("legacy schema catalog fingerprint manifest is out of sync: catalog=%d manifest=%d", len(seen), len(legacySchemaCatalogStepSourceFingerprints))
	}
	return nil
}

// LegacySchemaBaselineApplied is a read-only gate used by the migration
// process to avoid replaying the historical mutable bootstrap after baseline.
func LegacySchemaBaselineApplied(ctx context.Context, dbConn *sql.DB) (bool, error) {
	var table sql.NullString
	if err := dbConn.QueryRowContext(ctx, `SELECT to_regclass('schema_migrations')`).Scan(&table); err != nil {
		return false, err
	}
	if !table.Valid || strings.TrimSpace(table.String) == "" {
		return false, nil
	}
	var count int
	err := dbConn.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations
		WHERE scope = $1 AND version = $2 AND state = 'applied'`, platformMigrationScope, legacySchemaBaselineVersion).Scan(&count)
	return count == 1, err
}

func LegacySchemaCatalogCounts() (empresas, super int) {
	return len(legacyEmpresaSchemaCatalog), len(legacySuperSchemaCatalog)
}
