package db

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestLegacySchemaCatalogIsExplicitAndUnique(t *testing.T) {
	t.Parallel()
	seen := map[string]string{}
	for target, steps := range map[string][]legacySchemaStep{"empresas": legacyEmpresaSchemaCatalog, "super": legacySuperSchemaCatalog} {
		for _, step := range steps {
			if step.Name == "" || step.Apply == nil {
				t.Fatalf("invalid %s migration step: %#v", target, step)
			}
			if previous := seen[step.Name]; previous != "" {
				t.Fatalf("migration step %s repeated in %s and %s", step.Name, previous, target)
			}
			seen[step.Name] = target
		}
	}
	if len(legacyEmpresaSchemaCatalog) < 80 || len(legacySuperSchemaCatalog) < 25 {
		t.Fatalf("legacy schema inventory unexpectedly small: empresas=%d super=%d", len(legacyEmpresaSchemaCatalog), len(legacySuperSchemaCatalog))
	}
}

func TestLegacySchemaCatalogManifestCoversEveryCatalogStep(t *testing.T) {
	t.Parallel()
	if err := ValidateLegacySchemaCatalogManifest(); err != nil {
		t.Fatalf("legacy schema catalog manifest: %v", err)
	}
	if legacySchemaCatalogSourceFingerprint == "" {
		t.Fatal("legacy schema catalog fingerprint must be generated")
	}
}

func TestLegacySchemaCatalogPreparesMigrationLedgerBeforeBaselineRead(t *testing.T) {
	t.Parallel()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "legacy_schema_catalog.go", nil, 0)
	if err != nil {
		t.Fatalf("parse legacy catalog: %v", err)
	}
	var apply *ast.FuncDecl
	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if ok && fn.Name.Name == "ApplyLegacySchemaCatalog" {
			apply = fn
			break
		}
	}
	if apply == nil {
		t.Fatal("ApplyLegacySchemaCatalog is missing")
	}
	ledgerPrepared := false
	baselineRead := false
	ast.Inspect(apply.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		name, ok := call.Fun.(*ast.Ident)
		if !ok {
			return true
		}
		switch name.Name {
		case "EnsureSchemaMigrationsTable":
			ledgerPrepared = true
		case "LegacySchemaBaselineApplied":
			baselineRead = true
			if !ledgerPrepared {
				t.Error("baseline must not be read before the migration ledger is prepared")
			}
		}
		return true
	})
	if !ledgerPrepared || !baselineRead {
		t.Fatalf("expected ledger preparation and baseline read, got prepared=%t baseline=%t", ledgerPrepared, baselineRead)
	}
}

func TestLegacySchemaCatalogCoversEveryDatabaseEnsureSchemaFunction(t *testing.T) {
	registered := make(map[string]bool, len(legacyEmpresaSchemaCatalog)+len(legacySuperSchemaCatalog))
	for _, step := range append(append([]legacySchemaStep{}, legacyEmpresaSchemaCatalog...), legacySuperSchemaCatalog...) {
		registered[step.Name] = true
	}
	// These schemas are owned by versioned platform migrations rather than the
	// compatibility catalog, but they must remain part of the same coverage gate.
	registered["EnsureAsyncJobsSchema"] = true
	registered["EnsureMobileAPIIdempotencySchema"] = true
	registered["EnsureOutboxSchema"] = true

	packages, err := parser.ParseDir(token.NewFileSet(), ".", nil, 0)
	if err != nil {
		t.Fatalf("parse db package: %v", err)
	}
	pkg := packages["db"]
	if pkg == nil {
		t.Fatal("db package was not parsed")
	}
	for _, file := range pkg.Files {
		for _, declaration := range file.Decls {
			fn, ok := declaration.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || !strings.HasPrefix(fn.Name.Name, "Ensure") || !strings.HasSuffix(fn.Name.Name, "Schema") {
				continue
			}
			if !isSingleDatabaseEnsureSignature(fn.Type) {
				continue
			}
			if !registered[fn.Name.Name] {
				t.Errorf("%s must be registered in the pcs-migrate legacy catalog", fn.Name.Name)
			}
		}
	}
}

func isSingleDatabaseEnsureSignature(fn *ast.FuncType) bool {
	if fn == nil || fn.Params == nil || fn.Results == nil || len(fn.Params.List) != 1 || len(fn.Results.List) != 1 {
		return false
	}
	pointer, ok := fn.Params.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	selector, ok := pointer.X.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "DB" {
		return false
	}
	packageName, ok := selector.X.(*ast.Ident)
	if !ok || packageName.Name != "sql" {
		return false
	}
	result, ok := fn.Results.List[0].Type.(*ast.Ident)
	return ok && result.Name == "error"
}
