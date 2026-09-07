package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProductosDefaultViewLoadsInventorySummary(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "web", "administrar_empresa", "administrar_productos.html"))
	if err != nil {
		t.Fatalf("read administrar_productos.html: %v", err)
	}
	source := string(raw)
	start := strings.Index(source, "if (viewMode === 'productos') {")
	if start < 0 {
		t.Fatal("default products view bootstrap block not found")
	}
	end := strings.Index(source[start:], "\n      }")
	if end < 0 {
		t.Fatal("default products view bootstrap block is incomplete")
	}
	block := source[start : start+end]
	if !strings.Contains(block, "await loadInventarioResumen();") {
		t.Fatal("default products view must load inventory KPI summary")
	}
}
