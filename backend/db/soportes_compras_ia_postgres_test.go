package db

import (
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestSoporteComprasIAPapeleraPostgres(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("PCS_TEST_POSTGRES_DSN"))
	if dsn == "" {
		t.Skip("PCS_TEST_POSTGRES_DSN is not configured")
	}
	dbConn, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer dbConn.Close()
	dbConn.SetMaxOpenConns(1)
	if err := createTempSoportesComprasIAPapeleraTables(dbConn); err != nil {
		t.Fatal(err)
	}

	insert := func(empresaID int64, codigo, workflow, record, hash, document string, convertedID int64) int64 {
		t.Helper()
		var id int64
		err := dbConn.QueryRow(`INSERT INTO empresa_soportes_compras_ia
			(empresa_id,codigo,estado_soporte,estado,archivo_hash,documento_numero,convertido_id)
			VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`, empresaID, codigo, workflow, record, hash, document, convertedID).Scan(&id)
		if err != nil {
			t.Fatal(err)
		}
		return id
	}

	company12ID := insert(12, "SCI-PG-12", "en_revision", "activo", "hash-a", "FV-A", 0)
	_ = insert(53, "SCI-PG-53", "radicado", "activo", "hash-a", "FV-A", 0)
	deleted, err := UpdateEmpresaSoporteComprasIARegistroEstado(dbConn, 12, company12ID, "eliminado", "qa@local", "prueba postgres")
	if err != nil || deleted.Estado != "eliminado" {
		t.Fatalf("delete tenant 12: row=%#v err=%v", deleted, err)
	}
	if _, err := GetEmpresaSoporteComprasIAActivo(dbConn, 12, company12ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("deleted support must not be active: %v", err)
	}
	if _, err := UpdateEmpresaSoporteComprasIARegistroEstado(dbConn, 53, company12ID, "activo", "qa@local", "cross tenant"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("cross-tenant restore must return no rows: %v", err)
	}
	restored, err := UpdateEmpresaSoporteComprasIARegistroEstado(dbConn, 12, company12ID, "activo", "qa@local", "recuperacion postgres")
	if err != nil || restored.Estado != "activo" {
		t.Fatalf("restore tenant 12: row=%#v err=%v", restored, err)
	}

	if _, err := UpdateEmpresaSoporteComprasIARegistroEstado(dbConn, 12, company12ID, "eliminado", "qa@local", "preparar duplicado"); err != nil {
		t.Fatal(err)
	}
	if _, err := dbConn.Exec(`UPDATE empresa_soportes_compras_ia SET fecha_actualizacion=(CURRENT_TIMESTAMP-INTERVAL '120 days')::text WHERE empresa_id=12 AND id=$1`, company12ID); err != nil {
		t.Fatal(err)
	}
	retentionRows, err := ListEmpresaSoportesComprasIARetencion(dbConn, 12, 90, 20)
	if err != nil || len(retentionRows) != 1 || retentionRows[0].ID != company12ID {
		t.Fatalf("tenant retention preview: rows=%#v err=%v", retentionRows, err)
	}
	duplicateID := insert(12, "SCI-PG-DUP", "radicado", "activo", "hash-a", "FV-A", 0)
	if _, err := UpdateEmpresaSoporteComprasIARegistroEstado(dbConn, 12, company12ID, "activo", "qa@local", "debe bloquear duplicado"); err == nil || !strings.Contains(err.Error(), "soporte activo") {
		t.Fatalf("restore with active duplicate must fail: %v", err)
	}
	if _, err := UpdateEmpresaSoporteComprasIARegistroEstado(dbConn, 12, duplicateID, "eliminado", "qa@local", "retirar duplicado"); err != nil {
		t.Fatal(err)
	}
	if _, err := UpdateEmpresaSoporteComprasIARegistroEstado(dbConn, 12, company12ID, "activo", "qa@local", "recuperar sin duplicado"); err != nil {
		t.Fatal(err)
	}

	accountedID := insert(12, "SCI-PG-CXP", "contabilizado", "activo", "hash-cxp", "FV-CXP", 99)
	if _, err := UpdateEmpresaSoporteComprasIARegistroEstado(dbConn, 12, accountedID, "eliminado", "qa@local", "no permitido"); err == nil || !strings.Contains(err.Error(), "trazabilidad contable") {
		t.Fatalf("accounted support deletion must fail: %v", err)
	}
	var eventCount int
	if err := dbConn.QueryRow(`SELECT COUNT(*) FROM empresa_soportes_compras_ia_eventos WHERE empresa_id=12 AND soporte_id=$1 AND evento IN ('eliminar','restaurar')`, company12ID).Scan(&eventCount); err != nil {
		t.Fatal(err)
	}
	if eventCount != 4 {
		t.Fatalf("audited transitions = %d, want 4", eventCount)
	}

	purgeID := insert(12, "SCI-PG-PURGE", "en_revision", "eliminado", "hash-purge", "FV-PURGE", 0)
	if _, err := dbConn.Exec(`UPDATE empresa_soportes_compras_ia SET archivo_url='private://soportes_compras_ia/empresa_12/purge.pdf', fecha_actualizacion=(CURRENT_TIMESTAMP-INTERVAL '120 days')::text WHERE empresa_id=12 AND id=$1`, purgeID); err != nil {
		t.Fatal(err)
	}
	if _, err := PurgeEmpresaSoporteComprasIA(dbConn, 53, purgeID, 90, "SCI-PG-PURGE", "qa@local", "cross tenant"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("cross-tenant purge must return no rows: %v", err)
	}
	if _, err := PurgeEmpresaSoporteComprasIA(dbConn, 12, purgeID, 90, "WRONG", "qa@local", "wrong confirmation"); err == nil {
		t.Fatal("purge accepted wrong confirmation")
	}
	purged, err := PurgeEmpresaSoporteComprasIA(dbConn, 12, purgeID, 90, "SCI-PG-PURGE", "qa@local", "retention passed")
	if err != nil || purged.Estado != "purgado" || purged.ArchivoURL != "" {
		t.Fatalf("purge tombstone: row=%#v err=%v", purged, err)
	}
	if _, err := UpdateEmpresaSoporteComprasIARegistroEstado(dbConn, 12, purgeID, "activo", "qa@local", "must not restore"); err == nil {
		t.Fatal("purged support was restored")
	}
}

func TestSoporteComprasIARetentionEligibility(t *testing.T) {
	now := time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC)
	for _, raw := range []string{"2026-05-01 10:00:00", "2026-05-01T10:00:00Z", "2026-05-01 10:00:00-05"} {
		if !soporteComprasIARetentionEligible(raw, 90, now) {
			t.Fatalf("old timestamp rejected: %q", raw)
		}
	}
	if soporteComprasIARetentionEligible("2026-08-01 10:00:00", 90, now) {
		t.Fatal("recent timestamp accepted")
	}
	if soporteComprasIARetentionEligible("invalid", 90, now) {
		t.Fatal("invalid timestamp accepted")
	}
}

func createTempSoportesComprasIAPapeleraTables(dbConn *sql.DB) error {
	if _, err := dbConn.Exec(`CREATE TEMP TABLE empresa_soportes_compras_ia (
		id BIGSERIAL PRIMARY KEY, empresa_id INTEGER NOT NULL, codigo TEXT NOT NULL,
		tipo_soporte TEXT DEFAULT 'gasto', estado_soporte TEXT DEFAULT 'radicado', origen TEXT DEFAULT 'manual',
		archivo_nombre TEXT, archivo_url TEXT, archivo_mime TEXT, archivo_hash TEXT,
		proveedor_id INTEGER DEFAULT 0, proveedor_nombre TEXT, proveedor_nit TEXT,
		documento_tipo TEXT DEFAULT 'factura_compra', documento_numero TEXT, fecha_documento TEXT, fecha_vencimiento TEXT,
		subtotal REAL DEFAULT 0, impuesto_iva REAL DEFAULT 0, retencion_fuente REAL DEFAULT 0, retencion_ica REAL DEFAULT 0,
		retencion_iva REAL DEFAULT 0, total REAL DEFAULT 0, moneda TEXT DEFAULT 'COP', categoria_contable TEXT, centro_costo TEXT,
		impacta_inventario INTEGER DEFAULT 0, confianza_ia REAL DEFAULT 0, modelo_ia TEXT DEFAULT 'openai:gpt-5.5',
		extraccion_json TEXT, respuesta_ia TEXT, duplicado_soporte_id INTEGER DEFAULT 0, requiere_revision_humana INTEGER DEFAULT 1,
		aprobado_por TEXT, fecha_aprobacion TEXT, convertido_tipo TEXT, convertido_id INTEGER DEFAULT 0,
		fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP, fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
		usuario_creador TEXT, estado TEXT DEFAULT 'activo', observaciones TEXT, UNIQUE(empresa_id,codigo)
	)`); err != nil {
		return err
	}
	_, err := dbConn.Exec(`CREATE TEMP TABLE empresa_soportes_compras_ia_eventos (
		id BIGSERIAL PRIMARY KEY, empresa_id INTEGER NOT NULL, soporte_id INTEGER NOT NULL,
		evento TEXT NOT NULL, estado_anterior TEXT, estado_nuevo TEXT, detalle_json TEXT,
		fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP, usuario_creador TEXT
	)`)
	return err
}
