package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func PublicEstacionVIPHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch r.Method {
		case http.MethodGet:
			if action == "" || action == "info" || action == "productos" {
				handleVIPInfo(w, r, dbEmp)
				return
			}
			http.Error(w, "action invalida", http.StatusBadRequest)
			return
		case http.MethodPost:
			if action == "" || action == "agregar" || action == "agregar_item" {
				handleVIPAddItem(w, r, dbEmp)
				return
			}
			http.Error(w, "action invalida", http.StatusBadRequest)
			return
		default:
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
	}
}

func handleVIPInfo(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	codigo := strings.TrimSpace(r.URL.Query().Get("codigo"))
	if codigo == "" {
		http.Error(w, "codigo es obligatorio", http.StatusBadRequest)
		return
	}
	vip, err := dbpkg.GetVIPByCodigo(dbEmp, codigo)
	if err != nil {
		http.Error(w, "No se pudo validar codigo", http.StatusInternalServerError)
		return
	}
	if vip == nil || strings.ToLower(strings.TrimSpace(vip.Estado)) != "activo" {
		http.Error(w, "codigo invalido o vencido", http.StatusUnauthorized)
		return
	}
	if strings.TrimSpace(vip.ExpiraEn) != "" {
		if ts, perr := time.Parse("2006-01-02 15:04:05", strings.TrimSpace(vip.ExpiraEn)); perr == nil {
			if time.Now().After(ts) {
				_ = dbpkg.InvalidateVIPCodesForCarrito(dbEmp, vip.EmpresaID, vip.CarritoID, "codigo_vencido")
				http.Error(w, "codigo vencido", http.StatusUnauthorized)
				return
			}
		}
	}

	carrito, err := dbpkg.GetCarritoCompraByID(dbEmp, vip.EmpresaID, vip.CarritoID)
	if err != nil {
		http.Error(w, "No se pudo validar carrito", http.StatusInternalServerError)
		return
	}
	if carrito == nil || isCarritoVentaPagada(carrito) {
		_ = dbpkg.InvalidateVIPCodesForCarrito(dbEmp, vip.EmpresaID, vip.CarritoID, "carrito_pagado")
		http.Error(w, "carrito ya pagado", http.StatusUnauthorized)
		return
	}

	productos, err := dbpkg.GetProductosByEmpresa(dbEmp, vip.EmpresaID, "", "activo", 0, 0, 250, 0)
	if err != nil {
		http.Error(w, "No se pudieron listar productos", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok": true,
		"vip": map[string]interface{}{
			"empresa_id":  vip.EmpresaID,
			"estacion_id": vip.EstacionID,
			"carrito_id":  vip.CarritoID,
			"expira_en":   vip.ExpiraEn,
		},
		"carrito": carrito,
		"productos": productos,
	})
}

func handleVIPAddItem(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	var payload struct {
		Codigo    string  `json:"codigo"`
		ProductoID int64  `json:"producto_id"`
		Cantidad  float64 `json:"cantidad"`
		Nota      string  `json:"nota"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	payload.Codigo = strings.TrimSpace(payload.Codigo)
	if payload.Codigo == "" || payload.ProductoID <= 0 {
		http.Error(w, "codigo y producto_id son obligatorios", http.StatusBadRequest)
		return
	}
	if payload.Cantidad <= 0 {
		payload.Cantidad = 1
	}
	if payload.Cantidad > 999 {
		payload.Cantidad = 999
	}

	vip, err := dbpkg.GetVIPByCodigo(dbEmp, payload.Codigo)
	if err != nil || vip == nil || strings.ToLower(strings.TrimSpace(vip.Estado)) != "activo" {
		http.Error(w, "codigo invalido o vencido", http.StatusUnauthorized)
		return
	}

	carrito, err := dbpkg.GetCarritoCompraByID(dbEmp, vip.EmpresaID, vip.CarritoID)
	if err != nil || carrito == nil {
		http.Error(w, "carrito no disponible", http.StatusUnauthorized)
		return
	}
	if isCarritoVentaPagada(carrito) {
		_ = dbpkg.InvalidateVIPCodesForCarrito(dbEmp, vip.EmpresaID, vip.CarritoID, "carrito_pagado")
		http.Error(w, "carrito ya pagado", http.StatusUnauthorized)
		return
	}

	producto, err := dbpkg.GetProductoByID(dbEmp, vip.EmpresaID, payload.ProductoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "producto no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo validar producto", http.StatusInternalServerError)
		return
	}
	if strings.ToLower(strings.TrimSpace(producto.Estado)) != "activo" {
		http.Error(w, "producto inactivo", http.StatusConflict)
		return
	}

	obs := strings.TrimSpace(payload.Nota)
	if obs != "" {
		obs = "VIP: " + obs
	}

	itemID, err := dbpkg.CreateCarritoCompraItem(dbEmp, dbpkg.CarritoCompraItem{
		EmpresaID:      vip.EmpresaID,
		CarritoID:      vip.CarritoID,
		TipoItem:       "producto",
		ReferenciaID:   producto.ID,
		CodigoItem:     strings.TrimSpace(producto.SKU),
		Descripcion:    strings.TrimSpace(producto.Nombre),
		UnidadMedida:   "unidad",
		Cantidad:       payload.Cantidad,
		PrecioUnitario: producto.Precio,
		ImpuestoCodigo: "IVA",
		UsuarioCreador: "cliente_vip",
		Observaciones:  obs,
		Estado:         "activo",
	})
	if err != nil {
		http.Error(w, "No se pudo agregar el producto al carrito", http.StatusInternalServerError)
		return
	}

	// registrar evento métrica para que la estación pueda verlo como alerta (si se implementa polling).
	estacionID, estacionCodigo, estacionNombre := dbpkg.ResolveCarritoStationIdentity(carrito)
	_, _ = dbpkg.RecordCarritoStationMetric(dbEmp, dbpkg.CarritoStationMetricInput{
		EmpresaID:           vip.EmpresaID,
		CarritoID:           vip.CarritoID,
		EstacionID:          estacionID,
		EstacionCodigo:      estacionCodigo,
		EstacionNombre:      estacionNombre,
		EventoOperacion:     "pedido_vip",
		MetodoPago:          "",
		Moneda:              carrito.Moneda,
		MontoTotal:          0,
		MontoPagado:         0,
		ReferenciaOperacion: "vip:" + strconv.FormatInt(itemID, 10),
		FechaEvento:         time.Now().Format("2006-01-02 15:04:05"),
		UsuarioCreador:      "cliente_vip",
		Observaciones:       fmt.Sprintf("pedido_vip producto_id=%d x %.2f", producto.ID, payload.Cantidad),
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
		"item_id":  itemID,
		"carrito_id": vip.CarritoID,
	})
}

