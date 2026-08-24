package db

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestEmpresaCarteraCXPEdadRango(t *testing.T) {
	cases := []struct {
		name      string
		venc      string
		corte     string
		wantRango string
	}{
		{name: "por vencer", venc: "2026-06-20", corte: "2026-06-12", wantRango: "por_vencer"},
		{name: "vencido 30", venc: "2026-05-20", corte: "2026-06-12", wantRango: "0_30"},
		{name: "vencido 60", venc: "2026-04-20", corte: "2026-06-12", wantRango: "31_60"},
		{name: "vencido 90", venc: "2026-03-20", corte: "2026-06-12", wantRango: "61_90"},
		{name: "vencido 180", venc: "2026-01-20", corte: "2026-06-12", wantRango: "91_180"},
		{name: "vencido mayor", venc: "2025-01-20", corte: "2026-06-12", wantRango: "181_mas"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := empresaCarteraCXPEdadRango(tc.venc, tc.corte)
			if got != tc.wantRango {
				t.Fatalf("rango inesperado got=%s want=%s", got, tc.wantRango)
			}
		})
	}
}

func TestNormalizeEmpresaCarteraCXPEstado(t *testing.T) {
	if got := normalizeEmpresaCarteraCXPEstado("pendiente", 0, "2026-06-12"); got != "pagado" {
		t.Fatalf("saldo cero debe quedar pagado, got=%s", got)
	}
	if got := normalizeEmpresaCarteraCXPEstado("anulado", 100, "2026-06-12"); got != "anulado" {
		t.Fatalf("estado anulado debe conservarse, got=%s", got)
	}
}

func TestManualElectronicDocumentsCannotClaimDIANAcceptance(t *testing.T) {
	nomina := &EmpresaNominaElectronica{EstadoDIAN: "enviado", CUNE: "CUNE-FALSO", RespuestaDIAN: "aceptado", JSONPayload: `{"ok":true}`}
	normalizeEmpresaNominaElectronicaDraft(nomina)
	if nomina.EstadoDIAN != "borrador" || nomina.CUNE != "" || nomina.RespuestaDIAN != "" || nomina.JSONPayload != "" {
		t.Fatalf("nomina manual no quedo como borrador seguro: %+v", nomina)
	}
	soporte := &EmpresaDocumentoSoporteElectronico{EstadoDIAN: "validado", CUDS: "CUDS-FALSO", RespuestaDIAN: "aceptado", JSONPayload: `{"ok":true}`}
	normalizeEmpresaDocumentoSoporteDraft(soporte)
	if soporte.EstadoDIAN != "borrador" || soporte.CUDS != "" || soporte.RespuestaDIAN != "" || soporte.JSONPayload != "" {
		t.Fatalf("soporte manual no quedo como borrador seguro: %+v", soporte)
	}
}

func TestGetEmpresaDocumentoSoporteByIDContextIsTenantScoped(t *testing.T) {
	raw, err := os.ReadFile("contabilidad_colombia_avanzada.go")
	if err != nil {
		t.Fatalf("read document-support data access: %v", err)
	}
	if !strings.Contains(string(raw), "WHERE empresa_id = ? AND id = ?") {
		t.Fatal("document-support lookup must filter by empresa_id and id")
	}
}

func TestDIANManualDocumentsDraftMigrationIsCatalogued(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version != "20260823-002-dian-manual-document-drafts-v1" {
			continue
		}
		if migration.Body != empresaDIANManualDocumentDraftsFingerprint || migration.Apply == nil {
			t.Fatal("la migracion de borradores DIAN debe ser inmutable y ejecutable")
		}
		return
	}
	t.Fatal("no se encontro la migracion de borradores DIAN")
}

func TestDIANManualDocumentsDraftMigrationPostgres(t *testing.T) {
	dsn := os.Getenv("PCS_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("PCS_TEST_POSTGRES_DSN is not configured")
	}
	dbConn, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer dbConn.Close()
	tx, err := dbConn.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	for _, statement := range []string{
		`CREATE TEMP TABLE empresa_contabilidad_nomina_electronica (
			estado_dian TEXT, cune TEXT, respuesta_dian TEXT, json_payload TEXT,
			fecha_actualizacion TIMESTAMPTZ, total NUMERIC
		) ON COMMIT DROP`,
		`CREATE TEMP TABLE empresa_contabilidad_documentos_soporte (
			estado_dian TEXT, cuds TEXT, respuesta_dian TEXT, json_payload TEXT,
			fecha_actualizacion TIMESTAMPTZ, total NUMERIC
		) ON COMMIT DROP`,
		`INSERT INTO empresa_contabilidad_nomina_electronica (estado_dian, cune, respuesta_dian, json_payload, total)
			VALUES ('enviado', 'NO-VERIFICADO', 'aceptado', '{"fake":true}', 1250000)`,
		`INSERT INTO empresa_contabilidad_documentos_soporte (estado_dian, cuds, respuesta_dian, json_payload, total)
			VALUES ('validado', 'NO-VERIFICADO', 'aceptado', '{"fake":true}', 45000)`,
	} {
		if _, err := tx.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	if err := applyEmpresaDIANManualDocumentDraftsTx(context.Background(), tx); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"empresa_contabilidad_nomina_electronica", "empresa_contabilidad_documentos_soporte"} {
		var estado string
		var cuneOrCUDS, respuesta, payload sql.NullString
		var total string
		column := "cune"
		if table == "empresa_contabilidad_documentos_soporte" {
			column = "cuds"
		}
		if err := tx.QueryRow(`SELECT estado_dian, `+column+`, respuesta_dian, json_payload, total::TEXT FROM `+table).Scan(&estado, &cuneOrCUDS, &respuesta, &payload, &total); err != nil {
			t.Fatal(err)
		}
		if estado != "borrador" || cuneOrCUDS.Valid || respuesta.Valid || payload.Valid || total == "" {
			t.Fatalf("%s no quedo limpio y contablemente preservado: estado=%s id=%+v respuesta=%+v payload=%+v total=%s", table, estado, cuneOrCUDS, respuesta, payload, total)
		}
	}
}
