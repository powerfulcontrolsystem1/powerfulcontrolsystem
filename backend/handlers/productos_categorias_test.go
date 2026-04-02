package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestEmpresaCategoriasProductosHandlerCRUD(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_categorias_handler.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}

	h := EmpresaCategoriasProductosHandler(dbEmp)

	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/categorias_productos", strings.NewReader(`{"empresa_id":7,"nombre":"Bebidas","codigo":"CAT-BEB"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	h.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, createRR.Code, createRR.Body.String())
	}

	var createResp map[string]interface{}
	if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	categoriaID := int64(createResp["id"].(float64))
	if categoriaID <= 0 {
		t.Fatalf("invalid categoria id: %v", createResp)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/empresa/categorias_productos?empresa_id=7&include_inactive=1", nil)
	listRR := httptest.NewRecorder()
	h.ServeHTTP(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d body=%s", http.StatusOK, listRR.Code, listRR.Body.String())
	}
	var categorias []dbpkg.CategoriaProducto
	if err := json.Unmarshal(listRR.Body.Bytes(), &categorias); err != nil {
		t.Fatalf("decode categorias: %v", err)
	}
	if len(categorias) != 1 {
		t.Fatalf("expected 1 categoria, got %d", len(categorias))
	}

	updateBody := `{"id":` + strconv.FormatInt(categoriaID, 10) + `,"empresa_id":7,"nombre":"Bebidas frias","codigo":"CAT-BEB"}`
	updateReq := httptest.NewRequest(http.MethodPut, "/api/empresa/categorias_productos", strings.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateRR := httptest.NewRecorder()
	h.ServeHTTP(updateRR, updateReq)
	if updateRR.Code != http.StatusNoContent {
		t.Fatalf("expected update status %d, got %d body=%s", http.StatusNoContent, updateRR.Code, updateRR.Body.String())
	}

	toggleReq := httptest.NewRequest(http.MethodPut, "/api/empresa/categorias_productos?empresa_id=7&id="+strconv.FormatInt(categoriaID, 10)+"&action=activar&activo=0", nil)
	toggleRR := httptest.NewRecorder()
	h.ServeHTTP(toggleRR, toggleReq)
	if toggleRR.Code != http.StatusNoContent {
		t.Fatalf("expected toggle status %d, got %d body=%s", http.StatusNoContent, toggleRR.Code, toggleRR.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/empresa/categorias_productos?empresa_id=7&id="+strconv.FormatInt(categoriaID, 10), nil)
	deleteRR := httptest.NewRecorder()
	h.ServeHTTP(deleteRR, deleteReq)
	if deleteRR.Code != http.StatusNoContent {
		t.Fatalf("expected delete status %d, got %d body=%s", http.StatusNoContent, deleteRR.Code, deleteRR.Body.String())
	}
}

func TestEmpresaProductosHandlerFiltraPorCategoriaID(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_productos_categoria_filter.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}

	catA, err := dbpkg.CreateCategoriaProducto(dbEmp, dbpkg.CategoriaProducto{EmpresaID: 11, Nombre: "Tecnologia"})
	if err != nil {
		t.Fatalf("create categoria A: %v", err)
	}
	catB, err := dbpkg.CreateCategoriaProducto(dbEmp, dbpkg.CategoriaProducto{EmpresaID: 11, Nombre: "Hogar"})
	if err != nil {
		t.Fatalf("create categoria B: %v", err)
	}

	if _, err := dbpkg.CreateProducto(dbEmp, dbpkg.Producto{EmpresaID: 11, Nombre: "Laptop", CategoriaID: catA, Precio: 1000, Costo: 800}, 0, "TEST"); err != nil {
		t.Fatalf("create producto categoria A: %v", err)
	}
	if _, err := dbpkg.CreateProducto(dbEmp, dbpkg.Producto{EmpresaID: 11, Nombre: "Silla", CategoriaID: catB, Precio: 200, Costo: 120}, 0, "TEST"); err != nil {
		t.Fatalf("create producto categoria B: %v", err)
	}

	h := EmpresaProductosHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/productos?empresa_id=11&categoria_id="+strconv.FormatInt(catA, 10), nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var productos []dbpkg.Producto
	if err := json.Unmarshal(rr.Body.Bytes(), &productos); err != nil {
		t.Fatalf("decode productos: %v", err)
	}
	if len(productos) != 1 {
		t.Fatalf("expected 1 producto filtered, got %d", len(productos))
	}
	if productos[0].CategoriaID != catA {
		t.Fatalf("expected categoria_id=%d, got %d", catA, productos[0].CategoriaID)
	}
}
