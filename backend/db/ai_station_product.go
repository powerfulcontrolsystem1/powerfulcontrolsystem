package db

import (
	"database/sql"
	"fmt"
)

// Authority and actor are deliberately supplied separately by the server.
type EmpresaAIStationProductPlan struct {
	EstacionID int64   `json:"estacion_id"`
	CarritoID  int64   `json:"carrito_id"`
	ProductoID int64   `json:"producto_id"`
	Cantidad   int64   `json:"cantidad"`
	Precio     float64 `json:"precio"`
	Impuesto   float64 `json:"impuesto"`
	ActivadoEn string  `json:"activado_en"`
}

func AddEmpresaAIStationProduct(conn *sql.DB, empresaID int64, plan EmpresaAIStationProductPlan, user string) (int64, error) {
	if empresaID <= 0 || user == "" || plan.EstacionID <= 0 || plan.CarritoID <= 0 || plan.ProductoID <= 0 || plan.Cantidad < 1 || plan.Cantidad > 99 || plan.ActivadoEn == "" {
		return 0, fmt.Errorf("plan de consumo inválido")
	}
	product, err := GetProductoByID(conn, empresaID, plan.ProductoID)
	if err != nil || product == nil {
		return 0, fmt.Errorf("producto no disponible")
	}
	item := CarritoCompraItem{EmpresaID: empresaID, CarritoID: plan.CarritoID, TipoItem: "producto", ReferenciaID: product.ID, CodigoItem: product.SKU, Descripcion: product.Nombre, UnidadMedida: product.UnidadMedida, Cantidad: float64(plan.Cantidad), PrecioUnitario: plan.Precio, ImpuestoPorcentaje: plan.Impuesto, ImpuestoCodigo: "IVA", UsuarioCreador: user, Observaciones: "Consumo confirmado desde Agente PCS"}
	return createCarritoCompraItemGuarded(conn, item, func(tx *sql.Tx) error {
		var activated string
		// Lock the exact account/session. A paid or reused room cannot receive a stale proposal.
		err := tx.QueryRow(`SELECT COALESCE(activado_en::text,'') FROM carritos_compras WHERE empresa_id=$1 AND id=$2 AND codigo=$3 AND estado='activo' AND estado_carrito='abierto' AND estado_venta='venta_abierta' FOR UPDATE`, empresaID, plan.CarritoID, fmt.Sprintf("EST-%d-%d", empresaID, plan.EstacionID)).Scan(&activated)
		if err != nil || activated != plan.ActivadoEn {
			return fmt.Errorf("cuenta cerrada o sesión modificada")
		}
		var price, tax float64
		var unit string
		err = tx.QueryRow(`SELECT precio,COALESCE(impuesto_porcentaje,0),COALESCE(unidad_medida,'') FROM productos WHERE empresa_id=$1 AND id=$2 AND estado='activo' FOR SHARE`, empresaID, plan.ProductoID).Scan(&price, &tax, &unit)
		if err != nil || price != plan.Precio || tax != plan.Impuesto || unit != product.UnidadMedida {
			return fmt.Errorf("producto modificado; prepara otra propuesta")
		}
		return nil
	})
}
