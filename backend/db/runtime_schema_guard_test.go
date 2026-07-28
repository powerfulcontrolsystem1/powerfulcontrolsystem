package db

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeDDLGuardBlocksProductionAPI(t *testing.T) {
	t.Setenv("PCS_ENV", "production")
	t.Setenv("PCS_RUNTIME_ROLE", "api")
	if !runtimeDDLBlocked("CREATE TABLE should_not_run (id BIGINT)") {
		t.Fatal("production API must block DDL")
	}
	if runtimeDDLBlocked("UPDATE empresas SET estado = 'activo'") {
		t.Fatal("business DML must remain available")
	}
}

func TestRuntimeDDLGuardAllowsMigrator(t *testing.T) {
	t.Setenv("PCS_ENV", "production")
	t.Setenv("PCS_RUNTIME_ROLE", "migrate")
	if runtimeDDLBlocked("ALTER TABLE empresas ADD COLUMN ejemplo TEXT") {
		t.Fatal("migration role must be allowed to execute DDL")
	}
}

func TestSchemaStatementLoopsUseRuntimeDDLGuard(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		raw, err := os.ReadFile(entry.Name())
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(raw), "dbConn.Exec(stmt)") {
			t.Fatalf("%s bypasses runtime DDL guard for a schema statement loop", entry.Name())
		}
	}
}

func TestEnsureFunctionsDoNotBypassRuntimeDDLGuard(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	files := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(files, entry.Name(), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range parsed.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Body == nil || !strings.HasPrefix(strings.ToLower(function.Name.Name), "ensure") {
				continue
			}
			ast.Inspect(function.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || selector.Sel.Name != "Exec" {
					return true
				}
				receiver, ok := selector.X.(*ast.Ident)
				if ok && receiver.Name == "dbConn" {
					t.Errorf("%s:%d %s bypasses runtime DDL guard", entry.Name(), files.Position(call.Pos()).Line, function.Name.Name)
				}
				return true
			})
		}
	}
}
