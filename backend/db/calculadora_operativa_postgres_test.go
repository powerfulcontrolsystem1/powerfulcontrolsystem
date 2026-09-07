package db

import (
	"os"
	"strings"
	"testing"
)

func TestCalculadoraOperativaUsesPostgresReturningAndRequestContext(t *testing.T) {
	raw, err := os.ReadFile("calculadora_operativa.go")
	if err != nil {
		t.Fatalf("read calculadora_operativa.go: %v", err)
	}
	src := string(raw)

	for _, fn := range []string{
		"func UpsertEmpresaCalculadoraConfiguracionContext(",
		"func CreateEmpresaCalculadoraOperacionContext(",
	} {
		body := extractConfiguracionOperativaFunctionForTest(t, src, fn)
		if strings.Contains(body, "LastInsertId(") {
			t.Fatalf("%s no debe depender de LastInsertId en PostgreSQL", fn)
		}
		if !strings.Contains(body, "RETURNING id") {
			t.Fatalf("%s debe recuperar el id mediante RETURNING id", fn)
		}
		if !strings.Contains(body, "QueryRowContext(ctx") {
			t.Fatalf("%s debe propagar el contexto a PostgreSQL", fn)
		}
	}

	getBody := extractConfiguracionOperativaFunctionForTest(t, src, "func GetEmpresaCalculadoraConfiguracionContext(")
	if !strings.Contains(getBody, "QueryRowContext(ctx") {
		t.Fatal("la lectura de configuración debe ser cancelable")
	}
	if strings.Contains(getBody, "UpsertEmpresaCalculadoraConfiguracion(") {
		t.Fatal("una lectura GET no debe crear configuración en la base de datos")
	}
}
