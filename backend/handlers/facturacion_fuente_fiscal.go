package handlers

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	facturacionFuenteFiscalEsquema = "pcs.facturacion.fuente_fiscal"
	facturacionFuenteFiscalVersion = 1
)

// facturacionFuenteFiscalSnapshot is the immutable, secret-free source used to
// build an electronic sale document. It intentionally preserves missing master
// data as empty values plus blockers; it must never manufacture fiscal data.
type facturacionFuenteFiscalSnapshot struct {
	Esquema     string                             `json:"esquema"`
	Version     int                                `json:"version"`
	EmpresaID   int64                              `json:"empresa_id"`
	Documento   facturacionFuenteFiscalDocumento   `json:"documento"`
	Referencia  *facturacionFuenteFiscalReferencia `json:"referencia,omitempty"`
	Carrito     facturacionFuenteFiscalCarrito     `json:"carrito"`
	Emisor      facturacionFuenteFiscalParte       `json:"emisor"`
	Cliente     facturacionFuenteFiscalParte       `json:"cliente"`
	Lineas      []facturacionFuenteFiscalLinea     `json:"lineas"`
	Totales     facturacionFuenteFiscalTotales     `json:"totales"`
	Pago        facturacionFuenteFiscalPago        `json:"pago"`
	Bloqueantes []string                           `json:"bloqueantes"`
}

type facturacionFuenteFiscalReferencia struct {
	TipoDocumento    string `json:"tipo_documento"`
	DocumentoCodigo  string `json:"documento_codigo"`
	NumeroLegal      string `json:"numero_legal"`
	CodigoValidacion string `json:"codigo_validacion"`
	FechaEmision     string `json:"fecha_emision"`
}

type facturacionFuenteFiscalDocumento struct {
	TipoOrigen    string  `json:"tipo_origen"`
	CodigoOrigen  string  `json:"codigo_origen"`
	TipoDestino   string  `json:"tipo_destino"`
	CodigoDestino string  `json:"codigo_destino"`
	Fecha         string  `json:"fecha"`
	Moneda        string  `json:"moneda"`
	MontoTotal    float64 `json:"monto_total"`
}

type facturacionFuenteFiscalCarrito struct {
	ID              int64   `json:"id"`
	Codigo          string  `json:"codigo"`
	PagadoEn        string  `json:"pagado_en"`
	ClienteID       int64   `json:"cliente_id"`
	Subtotal        float64 `json:"subtotal"`
	DescuentoTotal  float64 `json:"descuento_total"`
	ImpuestoTotal   float64 `json:"impuesto_total"`
	Total           float64 `json:"total"`
	DevolucionTotal float64 `json:"devolucion_total"`
}

type facturacionFuenteFiscalParte struct {
	ID                        int64  `json:"id,omitempty"`
	TipoDocumento             string `json:"tipo_documento"`
	NumeroDocumento           string `json:"numero_documento"`
	DigitoVerificacion        string `json:"digito_verificacion"`
	TipoPersona               string `json:"tipo_persona"`
	NombreRazonSocial         string `json:"nombre_razon_social"`
	NombreComercial           string `json:"nombre_comercial"`
	RegimenFiscal             string `json:"regimen_fiscal"`
	ResponsabilidadTributaria string `json:"responsabilidad_tributaria"`
	Email                     string `json:"email"`
	Telefono                  string `json:"telefono"`
	Direccion                 string `json:"direccion"`
	PaisCodigo                string `json:"pais_codigo"`
	Departamento              string `json:"departamento"`
	DepartamentoCodigoDANE    string `json:"departamento_codigo_dane"`
	Municipio                 string `json:"municipio"`
	MunicipioCodigoDANE       string `json:"municipio_codigo_dane"`
	CodigoPostal              string `json:"codigo_postal"`
	ResponsabilidadesRUTJSON  string `json:"responsabilidades_rut_json,omitempty"`
	ObligacionesFiscalesJSON  string `json:"obligaciones_fiscales_json,omitempty"`
}

type facturacionFuenteFiscalLinea struct {
	Numero                int     `json:"numero"`
	ItemID                int64   `json:"item_id"`
	TipoItem              string  `json:"tipo_item"`
	ReferenciaID          int64   `json:"referencia_id,omitempty"`
	CodigoItem            string  `json:"codigo_item"`
	Descripcion           string  `json:"descripcion"`
	UnidadMedida          string  `json:"unidad_medida"`
	Cantidad              float64 `json:"cantidad"`
	PrecioUnitario        float64 `json:"precio_unitario"`
	DescuentoPorcentaje   float64 `json:"descuento_porcentaje"`
	ValorDescuento        float64 `json:"valor_descuento"`
	BaseGravable          float64 `json:"base_gravable"`
	ImpuestoCodigo        string  `json:"impuesto_codigo"`
	ImpuestoPorcentaje    float64 `json:"impuesto_porcentaje"`
	ValorImpuesto         float64 `json:"valor_impuesto"`
	SubtotalLinea         float64 `json:"subtotal_linea"`
	TotalLinea            float64 `json:"total_linea"`
	TratamientoTributario string  `json:"tratamiento_tributario"`
}

type facturacionFuenteFiscalTotales struct {
	BrutoLineas          float64 `json:"bruto_lineas"`
	DescuentoLineas      float64 `json:"descuento_lineas"`
	BaseGravableLineas   float64 `json:"base_gravable_lineas"`
	ImpuestoLineas       float64 `json:"impuesto_lineas"`
	TotalLineas          float64 `json:"total_lineas"`
	SubtotalCarrito      float64 `json:"subtotal_carrito"`
	DescuentoCarrito     float64 `json:"descuento_carrito"`
	ImpuestoCarrito      float64 `json:"impuesto_carrito"`
	TotalCarrito         float64 `json:"total_carrito"`
	TotalDocumentoOrigen float64 `json:"total_documento_origen"`
}

type facturacionFuenteFiscalPago struct {
	Metodo     string `json:"metodo"`
	Referencia string `json:"referencia"`
}

func buildFacturacionFuenteFiscalSnapshot(carrito *dbpkg.CarritoCompra, items []dbpkg.CarritoCompraItem, cfg *dbpkg.EmpresaConfiguracionAvanzada, cliente *dbpkg.Cliente, doc dbpkg.EmpresaDocumentoFacturacion) (*facturacionFuenteFiscalSnapshot, error) {
	if carrito == nil || carrito.EmpresaID <= 0 || carrito.ID <= 0 {
		return nil, fmt.Errorf("carrito empresarial invalido para fuente fiscal")
	}
	if doc.EmpresaID != carrito.EmpresaID || !strings.EqualFold(strings.TrimSpace(doc.TipoDocumento), "comprobante_pago") || strings.TrimSpace(doc.DocumentoCodigo) == "" {
		return nil, fmt.Errorf("documento de venta no pertenece al carrito empresarial")
	}
	if cfg != nil && cfg.EmpresaID != carrito.EmpresaID {
		return nil, fmt.Errorf("configuracion fiscal pertenece a otra empresa")
	}
	if doc.EntidadRelacionadaID > 0 && carrito.ClienteID > 0 && doc.EntidadRelacionadaID != carrito.ClienteID {
		return nil, fmt.Errorf("cliente del documento no coincide con el carrito")
	}
	if cliente != nil && (cliente.EmpresaID != carrito.EmpresaID || (carrito.ClienteID > 0 && cliente.ID != carrito.ClienteID)) {
		return nil, fmt.Errorf("cliente pertenece a otra empresa o venta")
	}
	if !facturacionFuenteFiscalNumerosValidos(doc.MontoTotal, carrito.Subtotal, carrito.DescuentoTotal, carrito.ImpuestoTotal, carrito.Total, carrito.DevolucionTotal) {
		return nil, fmt.Errorf("cabecera fiscal contiene valores no finitos")
	}

	moneda := strings.ToUpper(strings.TrimSpace(doc.Moneda))
	if moneda == "" {
		moneda = strings.ToUpper(strings.TrimSpace(carrito.Moneda))
	}
	snapshot := &facturacionFuenteFiscalSnapshot{
		Esquema:   facturacionFuenteFiscalEsquema,
		Version:   facturacionFuenteFiscalVersion,
		EmpresaID: carrito.EmpresaID,
		Documento: facturacionFuenteFiscalDocumento{
			TipoOrigen:    "comprobante_pago",
			CodigoOrigen:  strings.TrimSpace(doc.DocumentoCodigo),
			TipoDestino:   "factura_electronica",
			CodigoDestino: buildVentaDocumentoCodigoFromBase(doc.DocumentoCodigo, "factura_electronica"),
			Fecha:         strings.TrimSpace(doc.FechaDocumento),
			Moneda:        moneda,
			MontoTotal:    doc.MontoTotal,
		},
		Carrito: facturacionFuenteFiscalCarrito{
			ID: carrito.ID, Codigo: strings.TrimSpace(carrito.Codigo), PagadoEn: strings.TrimSpace(carrito.PagadoEn),
			ClienteID: carrito.ClienteID, Subtotal: carrito.Subtotal, DescuentoTotal: carrito.DescuentoTotal,
			ImpuestoTotal: carrito.ImpuestoTotal, Total: carrito.Total, DevolucionTotal: carrito.DevolucionTotal,
		},
		Lineas: make([]facturacionFuenteFiscalLinea, 0, len(items)),
		Pago: facturacionFuenteFiscalPago{
			Metodo: strings.TrimSpace(carrito.MetodoPago), Referencia: strings.TrimSpace(carrito.ReferenciaPago),
		},
	}

	snapshot.Emisor, snapshot.Cliente = facturacionFuenteFiscalPartes(cfg, cliente)
	if err := facturacionFuenteFiscalAgregarLineas(snapshot, carrito, items); err != nil {
		return nil, err
	}
	facturacionFuenteFiscalFinalizarTotales(snapshot, carrito, doc)
	facturacionFuenteFiscalCompletarBloqueantes(snapshot, cfg, cliente)
	return snapshot, nil
}

func facturacionFuenteFiscalPartes(cfg *dbpkg.EmpresaConfiguracionAvanzada, cliente *dbpkg.Cliente) (facturacionFuenteFiscalParte, facturacionFuenteFiscalParte) {
	emisor := facturacionFuenteFiscalParte{}
	adquiriente := facturacionFuenteFiscalParte{}
	if cfg != nil {
		emisor = facturacionFuenteFiscalParte{
			TipoDocumento: strings.TrimSpace(cfg.TipoDocumentoEmisor), NumeroDocumento: strings.TrimSpace(cfg.NIT),
			DigitoVerificacion: strings.TrimSpace(cfg.DigitoVerificacion), TipoPersona: strings.TrimSpace(cfg.TipoPersonaFiscal),
			NombreRazonSocial: strings.TrimSpace(cfg.RazonSocial), NombreComercial: strings.TrimSpace(cfg.NombreComercial),
			RegimenFiscal: strings.TrimSpace(cfg.RegimenFiscal), ResponsabilidadTributaria: strings.TrimSpace(cfg.ResponsabilidadTributaria),
			Email: strings.TrimSpace(cfg.EmailFacturacion), Telefono: strings.TrimSpace(cfg.TelefonoFacturacion),
			Direccion: strings.TrimSpace(cfg.DireccionFiscal), PaisCodigo: strings.ToUpper(strings.TrimSpace(cfg.PaisCodigo)),
			Departamento: strings.TrimSpace(cfg.Departamento), DepartamentoCodigoDANE: strings.TrimSpace(cfg.DepartamentoCodigoDANE),
			Municipio: strings.TrimSpace(cfg.Municipio), MunicipioCodigoDANE: strings.TrimSpace(cfg.MunicipioCodigoDANE), CodigoPostal: strings.TrimSpace(cfg.CodigoPostal),
			ResponsabilidadesRUTJSON: strings.TrimSpace(cfg.ResponsabilidadesRUTJSON), ObligacionesFiscalesJSON: strings.TrimSpace(cfg.ObligacionesFiscalesJSON),
		}
	}
	if cliente != nil {
		adquiriente = facturacionFuenteFiscalParte{
			ID: cliente.ID, TipoDocumento: strings.TrimSpace(cliente.TipoDocumento), NumeroDocumento: strings.TrimSpace(cliente.NumeroDocumento),
			DigitoVerificacion: strings.TrimSpace(cliente.DigitoVerificacion), TipoPersona: strings.TrimSpace(cliente.TipoPersona),
			NombreRazonSocial: strings.TrimSpace(cliente.NombreRazonSocial), NombreComercial: strings.TrimSpace(cliente.NombreComercial),
			RegimenFiscal: strings.TrimSpace(cliente.RegimenFiscal), ResponsabilidadTributaria: strings.TrimSpace(cliente.ResponsabilidadTributaria),
			Email: strings.TrimSpace(cliente.Email), Telefono: strings.TrimSpace(cliente.Telefono), Direccion: strings.TrimSpace(cliente.Direccion),
			PaisCodigo: strings.ToUpper(strings.TrimSpace(cliente.Pais)), Departamento: strings.TrimSpace(cliente.Departamento),
			DepartamentoCodigoDANE: strings.TrimSpace(cliente.DepartamentoCodigoDANE), Municipio: strings.TrimSpace(cliente.Municipio), MunicipioCodigoDANE: strings.TrimSpace(cliente.MunicipioCodigoDANE), CodigoPostal: strings.TrimSpace(cliente.CodigoPostal),
		}
	}
	return emisor, adquiriente
}

func facturacionFuenteFiscalAgregarLineas(snapshot *facturacionFuenteFiscalSnapshot, carrito *dbpkg.CarritoCompra, items []dbpkg.CarritoCompraItem) error {
	orderedItems := append([]dbpkg.CarritoCompraItem(nil), items...)
	sort.SliceStable(orderedItems, func(i, j int) bool {
		if orderedItems[i].ID != orderedItems[j].ID {
			return orderedItems[i].ID < orderedItems[j].ID
		}
		return strings.TrimSpace(orderedItems[i].CodigoItem) < strings.TrimSpace(orderedItems[j].CodigoItem)
	})
	for index, item := range orderedItems {
		if item.EmpresaID != carrito.EmpresaID || item.CarritoID != carrito.ID {
			return fmt.Errorf("linea de carrito pertenece a otra empresa o carrito")
		}
		if !facturacionFuenteFiscalNumerosValidos(item.Cantidad, item.PrecioUnitario, item.DescuentoPorcentaje, item.ValorDescuento, item.BaseGravable, item.ImpuestoPorcentaje, item.ValorImpuesto, item.SubtotalLinea, item.TotalLinea) {
			return fmt.Errorf("linea fiscal %d contiene valores no finitos", item.ID)
		}
		linea := facturacionFuenteFiscalLinea{
			Numero: index + 1, ItemID: item.ID, TipoItem: strings.TrimSpace(item.TipoItem), ReferenciaID: item.ReferenciaID,
			CodigoItem: strings.TrimSpace(item.CodigoItem), Descripcion: strings.TrimSpace(item.Descripcion), UnidadMedida: strings.TrimSpace(item.UnidadMedida),
			Cantidad: item.Cantidad, PrecioUnitario: item.PrecioUnitario, DescuentoPorcentaje: item.DescuentoPorcentaje,
			ValorDescuento: item.ValorDescuento, BaseGravable: item.BaseGravable, ImpuestoCodigo: strings.TrimSpace(item.ImpuestoCodigo),
			ImpuestoPorcentaje: item.ImpuestoPorcentaje, ValorImpuesto: item.ValorImpuesto, SubtotalLinea: item.SubtotalLinea,
			TotalLinea: item.TotalLinea,
		}
		_, _, tratamiento, _, tratamientoOK := dianFuenteFiscalTaxTreatment(linea.ImpuestoCodigo, linea.ImpuestoPorcentaje)
		if tratamientoOK {
			linea.TratamientoTributario = tratamiento
		}
		snapshot.Lineas = append(snapshot.Lineas, linea)
		snapshot.Totales.BrutoLineas += item.Cantidad * item.PrecioUnitario
		snapshot.Totales.DescuentoLineas += item.ValorDescuento
		snapshot.Totales.BaseGravableLineas += item.BaseGravable
		snapshot.Totales.ImpuestoLineas += item.ValorImpuesto
		snapshot.Totales.TotalLineas += item.TotalLinea
		if linea.Descripcion == "" {
			snapshot.Bloqueantes = append(snapshot.Bloqueantes, fmt.Sprintf("lineas.%d.descripcion_faltante", linea.Numero))
		}
		if linea.CodigoItem == "" {
			snapshot.Bloqueantes = append(snapshot.Bloqueantes, fmt.Sprintf("lineas.%d.codigo_item_faltante", linea.Numero))
		}
		if linea.UnidadMedida == "" {
			snapshot.Bloqueantes = append(snapshot.Bloqueantes, fmt.Sprintf("lineas.%d.unidad_medida_faltante", linea.Numero))
		}
		if linea.Cantidad <= 0 || linea.PrecioUnitario < 0 || linea.TotalLinea < 0 {
			snapshot.Bloqueantes = append(snapshot.Bloqueantes, fmt.Sprintf("lineas.%d.valores_invalidos", linea.Numero))
		}
		bruto := facturacionFuenteFiscalRound(linea.Cantidad * linea.PrecioUnitario)
		if !facturacionFuenteFiscalClose(bruto-linea.ValorDescuento, linea.BaseGravable) ||
			!facturacionFuenteFiscalClose(linea.BaseGravable+linea.ValorImpuesto, linea.TotalLinea) {
			snapshot.Bloqueantes = append(snapshot.Bloqueantes, fmt.Sprintf("lineas.%d.totales_no_conciliados", linea.Numero))
		}
		if linea.DescuentoPorcentaje < 0 || linea.DescuentoPorcentaje > 100 ||
			!facturacionFuenteFiscalClose(bruto*linea.DescuentoPorcentaje/100, linea.ValorDescuento) {
			snapshot.Bloqueantes = append(snapshot.Bloqueantes, fmt.Sprintf("lineas.%d.descuento_no_conciliado", linea.Numero))
		}
		if linea.ImpuestoPorcentaje > 0 && !facturacionFuenteFiscalClose(linea.BaseGravable*linea.ImpuestoPorcentaje/100, linea.ValorImpuesto) {
			snapshot.Bloqueantes = append(snapshot.Bloqueantes, fmt.Sprintf("lineas.%d.impuesto_no_conciliado", linea.Numero))
		}
		if linea.ImpuestoPorcentaje > 0 && linea.ImpuestoCodigo == "" {
			snapshot.Bloqueantes = append(snapshot.Bloqueantes, fmt.Sprintf("lineas.%d.impuesto_codigo_faltante", linea.Numero))
		}
		if !tratamientoOK {
			snapshot.Bloqueantes = append(snapshot.Bloqueantes, fmt.Sprintf("lineas.%d.tratamiento_tributario_faltante", linea.Numero))
		}
		if (tratamiento == "exento" || tratamiento == "excluido") && !facturacionFuenteFiscalClose(linea.ValorImpuesto, 0) {
			snapshot.Bloqueantes = append(snapshot.Bloqueantes, fmt.Sprintf("lineas.%d.impuesto_no_conciliado", linea.Numero))
		}
	}
	return nil
}

func facturacionFuenteFiscalFinalizarTotales(snapshot *facturacionFuenteFiscalSnapshot, carrito *dbpkg.CarritoCompra, doc dbpkg.EmpresaDocumentoFacturacion) {
	snapshot.Totales.BrutoLineas = facturacionFuenteFiscalRound(snapshot.Totales.BrutoLineas)
	snapshot.Totales.DescuentoLineas = facturacionFuenteFiscalRound(snapshot.Totales.DescuentoLineas)
	snapshot.Totales.BaseGravableLineas = facturacionFuenteFiscalRound(snapshot.Totales.BaseGravableLineas)
	snapshot.Totales.ImpuestoLineas = facturacionFuenteFiscalRound(snapshot.Totales.ImpuestoLineas)
	snapshot.Totales.TotalLineas = facturacionFuenteFiscalRound(snapshot.Totales.TotalLineas)
	snapshot.Totales.SubtotalCarrito = carrito.Subtotal
	snapshot.Totales.DescuentoCarrito = carrito.DescuentoTotal
	snapshot.Totales.ImpuestoCarrito = carrito.ImpuestoTotal
	snapshot.Totales.TotalCarrito = carrito.Total
	snapshot.Totales.TotalDocumentoOrigen = doc.MontoTotal
}

func facturacionFuenteFiscalCompletarBloqueantes(snapshot *facturacionFuenteFiscalSnapshot, cfg *dbpkg.EmpresaConfiguracionAvanzada, cliente *dbpkg.Cliente) {
	if snapshot == nil {
		return
	}
	if cfg == nil || snapshot.Emisor.NumeroDocumento == "" {
		snapshot.Bloqueantes = append(snapshot.Bloqueantes, "emisor.numero_documento_faltante")
	}
	if cfg == nil || snapshot.Emisor.NombreRazonSocial == "" {
		snapshot.Bloqueantes = append(snapshot.Bloqueantes, "emisor.nombre_razon_social_faltante")
	}
	if snapshot.Emisor.DigitoVerificacion != "" {
		if expected, ok := calculateColombianNITDV(snapshot.Emisor.NumeroDocumento); !ok || strconv.Itoa(expected) != dianOnlyDigits(snapshot.Emisor.DigitoVerificacion) {
			snapshot.Bloqueantes = append(snapshot.Bloqueantes, "emisor.digito_verificacion_invalido")
		}
	}
	if snapshot.Emisor.TipoDocumento == "" {
		snapshot.Bloqueantes = append(snapshot.Bloqueantes, "emisor.tipo_documento_faltante")
	}
	if snapshot.Emisor.Direccion == "" {
		snapshot.Bloqueantes = append(snapshot.Bloqueantes, "emisor.direccion_faltante")
	}
	if snapshot.Emisor.Departamento == "" {
		snapshot.Bloqueantes = append(snapshot.Bloqueantes, "emisor.departamento_faltante")
	}
	if snapshot.Emisor.Municipio == "" {
		snapshot.Bloqueantes = append(snapshot.Bloqueantes, "emisor.municipio_faltante")
	}
	if snapshot.Emisor.ResponsabilidadTributaria == "" {
		snapshot.Bloqueantes = append(snapshot.Bloqueantes, "emisor.responsabilidad_tributaria_faltante")
	}
	if !strings.EqualFold(snapshot.Emisor.PaisCodigo, "CO") {
		snapshot.Bloqueantes = append(snapshot.Bloqueantes, "emisor.pais_codigo_colombia_requerido")
	}
	if !facturacionFuenteFiscalDANEDepartamentoValido(snapshot.Emisor.DepartamentoCodigoDANE) {
		snapshot.Bloqueantes = append(snapshot.Bloqueantes, "emisor.departamento_codigo_dane_faltante")
	}
	if !facturacionFuenteFiscalDANEMunicipioValido(snapshot.Emisor.MunicipioCodigoDANE, snapshot.Emisor.DepartamentoCodigoDANE) {
		snapshot.Bloqueantes = append(snapshot.Bloqueantes, "emisor.municipio_codigo_dane_faltante")
	}
	if cliente == nil || snapshot.Cliente.ID <= 0 {
		snapshot.Bloqueantes = append(snapshot.Bloqueantes, "cliente.no_asignado_o_no_encontrado")
	} else {
		if snapshot.Cliente.NumeroDocumento == "" {
			snapshot.Bloqueantes = append(snapshot.Bloqueantes, "cliente.numero_documento_faltante")
		}
		if snapshot.Cliente.NombreRazonSocial == "" {
			snapshot.Bloqueantes = append(snapshot.Bloqueantes, "cliente.nombre_razon_social_faltante")
		}
		if snapshot.Cliente.TipoDocumento == "" {
			snapshot.Bloqueantes = append(snapshot.Bloqueantes, "cliente.tipo_documento_faltante")
		}
		if dianCustomerDocumentSchemeName(snapshot.Cliente.TipoDocumento, snapshot.Cliente.NumeroDocumento) == "31" && snapshot.Cliente.DigitoVerificacion != "" {
			if expected, ok := calculateColombianNITDV(snapshot.Cliente.NumeroDocumento); !ok || strconv.Itoa(expected) != dianOnlyDigits(snapshot.Cliente.DigitoVerificacion) {
				snapshot.Bloqueantes = append(snapshot.Bloqueantes, "cliente.digito_verificacion_invalido")
			}
		}
		if snapshot.Cliente.Direccion == "" {
			snapshot.Bloqueantes = append(snapshot.Bloqueantes, "cliente.direccion_faltante")
		}
		if snapshot.Cliente.Departamento == "" {
			snapshot.Bloqueantes = append(snapshot.Bloqueantes, "cliente.departamento_faltante")
		}
		if snapshot.Cliente.Municipio == "" {
			snapshot.Bloqueantes = append(snapshot.Bloqueantes, "cliente.municipio_faltante")
		}
		if snapshot.Cliente.ResponsabilidadTributaria == "" {
			snapshot.Bloqueantes = append(snapshot.Bloqueantes, "cliente.responsabilidad_tributaria_faltante")
		}
		if !strings.EqualFold(snapshot.Cliente.PaisCodigo, "CO") {
			snapshot.Bloqueantes = append(snapshot.Bloqueantes, "cliente.pais_codigo_colombia_requerido")
		}
		if !facturacionFuenteFiscalDANEDepartamentoValido(snapshot.Cliente.DepartamentoCodigoDANE) {
			snapshot.Bloqueantes = append(snapshot.Bloqueantes, "cliente.departamento_codigo_dane_faltante")
		}
		if !facturacionFuenteFiscalDANEMunicipioValido(snapshot.Cliente.MunicipioCodigoDANE, snapshot.Cliente.DepartamentoCodigoDANE) {
			snapshot.Bloqueantes = append(snapshot.Bloqueantes, "cliente.municipio_codigo_dane_faltante")
		}
	}
	if len(snapshot.Lineas) == 0 {
		snapshot.Bloqueantes = append(snapshot.Bloqueantes, "lineas.vacias")
	}
	if !facturacionFuenteFiscalClose(snapshot.Totales.DescuentoLineas, snapshot.Totales.DescuentoCarrito) {
		snapshot.Bloqueantes = append(snapshot.Bloqueantes, "totales.descuento_global_sin_asignar")
	}
	if !facturacionFuenteFiscalClose(snapshot.Totales.ImpuestoLineas, snapshot.Totales.ImpuestoCarrito) {
		snapshot.Bloqueantes = append(snapshot.Bloqueantes, "totales.impuestos_no_conciliados")
	}
	if !facturacionFuenteFiscalClose(snapshot.Totales.TotalLineas, snapshot.Totales.TotalDocumentoOrigen) {
		snapshot.Bloqueantes = append(snapshot.Bloqueantes, "totales.lineas_no_concilian_documento")
	}
	if !facturacionFuenteFiscalClose(snapshot.Totales.TotalCarrito, snapshot.Totales.TotalDocumentoOrigen) {
		snapshot.Bloqueantes = append(snapshot.Bloqueantes, "totales.carrito_no_concilia_documento")
	}
	snapshot.Bloqueantes = facturacionFuenteFiscalUniqueSorted(snapshot.Bloqueantes)
}

func facturacionFuenteFiscalDANEDepartamentoValido(value string) bool {
	value = strings.TrimSpace(value)
	return len(value) == 2 && dianOnlyDigits(value) == value
}

func facturacionFuenteFiscalDANEMunicipioValido(value, departamento string) bool {
	value = strings.TrimSpace(value)
	departamento = strings.TrimSpace(departamento)
	return len(value) == 5 && dianOnlyDigits(value) == value && facturacionFuenteFiscalDANEDepartamentoValido(departamento) && strings.HasPrefix(value, departamento)
}

func facturacionFuenteFiscalNumerosValidos(values ...float64) bool {
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	return true
}

func facturacionFuenteFiscalRound(value float64) float64 {
	return math.Round(value*100) / 100
}

func facturacionFuenteFiscalClose(a, b float64) bool {
	return math.Abs(a-b) <= 0.01
}

func facturacionFuenteFiscalUniqueSorted(values []string) []string {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			set[value] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for value := range set {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func marshalFacturacionFuenteFiscal(snapshot *facturacionFuenteFiscalSnapshot) ([]byte, error) {
	if snapshot == nil || snapshot.EmpresaID <= 0 || snapshot.Esquema != facturacionFuenteFiscalEsquema || snapshot.Version != facturacionFuenteFiscalVersion {
		return nil, fmt.Errorf("fuente fiscal invalida")
	}
	tipoOrigen := normalizeFacturacionDocumentoElectronicoTipo(snapshot.Documento.TipoOrigen)
	if strings.TrimSpace(snapshot.Documento.CodigoOrigen) == "" || (tipoOrigen != "comprobante_pago" && tipoOrigen != "nota_credito") {
		return nil, fmt.Errorf("origen de fuente fiscal invalido")
	}
	if tipoOrigen == "nota_credito" {
		if snapshot.Referencia == nil || normalizeFacturacionDocumentoElectronicoTipo(snapshot.Referencia.TipoDocumento) != "factura_electronica" ||
			strings.TrimSpace(snapshot.Referencia.NumeroLegal) == "" || !facturacionCodigoSHA384Valido(snapshot.Referencia.CodigoValidacion) ||
			strings.TrimSpace(snapshot.Referencia.FechaEmision) == "" {
			return nil, fmt.Errorf("referencia fiscal de nota credito invalida")
		}
	}
	if !facturacionFuenteFiscalNumerosValidos(snapshot.Documento.MontoTotal, snapshot.Totales.TotalDocumentoOrigen) {
		return nil, fmt.Errorf("fuente fiscal contiene totales no finitos")
	}
	return json.Marshal(snapshot)
}

func saveFacturacionFuenteFiscalSnapshot(ctx context.Context, dbEmp *sql.DB, doc dbpkg.EmpresaDocumentoFacturacion, snapshot *facturacionFuenteFiscalSnapshot) (*dbpkg.EmpresaFacturacionArtefacto, error) {
	if dbEmp == nil || snapshot == nil || doc.EmpresaID <= 0 || snapshot.EmpresaID != doc.EmpresaID {
		return nil, fmt.Errorf("fuente fiscal empresarial invalida")
	}
	if !strings.EqualFold(strings.TrimSpace(doc.TipoDocumento), snapshot.Documento.TipoOrigen) || !strings.EqualFold(strings.TrimSpace(doc.DocumentoCodigo), snapshot.Documento.CodigoOrigen) {
		return nil, fmt.Errorf("fuente fiscal no corresponde al documento origen")
	}
	raw, err := marshalFacturacionFuenteFiscal(snapshot)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(raw)
	hashHex := hex.EncodeToString(hash[:])
	existing, existingErr := dbpkg.GetEmpresaFacturacionArtefactoByTypeContext(ctx, dbEmp, doc.EmpresaID, doc.TipoDocumento, doc.DocumentoCodigo, dbpkg.EmpresaFacturacionArtefactoTipoFuenteFiscalJSON)
	if existingErr == nil && existing != nil {
		if !strings.EqualFold(existing.SHA256, hashHex) || existing.TamanoBytes != int64(len(raw)) {
			return nil, dbpkg.ErrEmpresaFacturacionFuenteFiscalInmutable
		}
		if _, err := loadFacturacionFuenteFiscalSnapshot(ctx, dbEmp, doc.EmpresaID, doc.TipoDocumento, doc.DocumentoCodigo); err != nil {
			return nil, err
		}
		return existing, nil
	}
	if existingErr != nil && !errors.Is(existingErr, sql.ErrNoRows) {
		return nil, existingErr
	}

	name, path, written, err := saveEmpresaPrivateUpload(doc.EmpresaID, "facturacion_electronica", ".json", strings.NewReader(string(raw)), 4<<20)
	if err != nil {
		return nil, err
	}
	item, err := dbpkg.InsertEmpresaFacturacionFuenteFiscalContext(ctx, dbEmp, dbpkg.EmpresaFacturacionArtefacto{
		EmpresaID: doc.EmpresaID, TipoDocumento: doc.TipoDocumento, DocumentoCodigo: doc.DocumentoCodigo,
		TipoArtefacto: dbpkg.EmpresaFacturacionArtefactoTipoFuenteFiscalJSON, StorageRef: name, SHA256: hashHex,
		MimeType: "application/json", TamanoBytes: written, Estado: "activo",
	})
	if err != nil {
		_ = os.Remove(path)
		return nil, err
	}
	if item.StorageRef != name {
		_ = os.Remove(path)
	}
	return item, nil
}

func loadFacturacionFuenteFiscalSnapshot(ctx context.Context, dbEmp *sql.DB, empresaID int64, tipoDocumento, documentoCodigo string) (*facturacionFuenteFiscalSnapshot, error) {
	doc := dbpkg.EmpresaDocumentoFacturacion{EmpresaID: empresaID, TipoDocumento: tipoDocumento, DocumentoCodigo: documentoCodigo}
	raw, err := loadFacturacionFiscalArtifact(ctx, dbEmp, doc, dbpkg.EmpresaFacturacionArtefactoTipoFuenteFiscalJSON)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	var snapshot facturacionFuenteFiscalSnapshot
	if err := decoder.Decode(&snapshot); err != nil {
		return nil, fmt.Errorf("decodificar fuente fiscal: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("fuente fiscal contiene datos adicionales")
	}
	if snapshot.Esquema != facturacionFuenteFiscalEsquema || snapshot.Version != facturacionFuenteFiscalVersion || snapshot.EmpresaID != empresaID {
		return nil, fmt.Errorf("fuente fiscal no pertenece a la empresa o version esperada")
	}
	if !strings.EqualFold(snapshot.Documento.TipoOrigen, strings.TrimSpace(tipoDocumento)) || !strings.EqualFold(snapshot.Documento.CodigoOrigen, strings.TrimSpace(documentoCodigo)) {
		return nil, fmt.Errorf("fuente fiscal no corresponde al documento solicitado")
	}
	return &snapshot, nil
}

func loadFacturacionFuenteFiscalParaDocumento(ctx context.Context, dbEmp *sql.DB, doc dbpkg.EmpresaDocumentoFacturacion) (*facturacionFuenteFiscalSnapshot, error) {
	tipo := strings.ToLower(strings.TrimSpace(doc.TipoDocumento))
	codigo := strings.TrimSpace(doc.DocumentoCodigo)
	switch tipo {
	case "comprobante_pago":
		return loadFacturacionFuenteFiscalSnapshot(ctx, dbEmp, doc.EmpresaID, tipo, codigo)
	case "factura_electronica":
		codigoOrigen := buildVentaDocumentoCodigoFromBase(codigo, "comprobante_pago")
		snapshot, err := loadFacturacionFuenteFiscalSnapshot(ctx, dbEmp, doc.EmpresaID, "comprobante_pago", codigoOrigen)
		if err != nil {
			return nil, err
		}
		if !strings.EqualFold(snapshot.Documento.CodigoDestino, codigo) {
			return nil, fmt.Errorf("fuente fiscal no corresponde a la factura solicitada")
		}
		return snapshot, nil
	case "nota_credito":
		return loadFacturacionFuenteFiscalSnapshot(ctx, dbEmp, doc.EmpresaID, tipo, codigo)
	default:
		return nil, fmt.Errorf("tipo documental sin fuente fiscal de venta")
	}
}

type dianFuenteFiscalTaxGroup struct {
	Codigo     string
	Nombre     string
	Porcentaje float64
	Base       float64
	Impuesto   float64
}

type dianFuenteFiscalUBLContext struct {
	DocumentoTipo      string
	DocumentoCodigo    string
	EmisorNIT          string
	Prefijo            string
	ResolucionNumero   string
	ResolucionDesde    string
	ResolucionHasta    string
	LlaveTecnica       string
	RangoDesde         int64
	RangoHasta         int64
	SoftwareID         string
	SoftwarePIN        string
	IssueDate          string
	IssueTime          string
	Moneda             string
	ProfileExecutionID string
}

// generateDIANUBLDesdeFuenteFiscal is the only commercial UBL generator. It
// accepts a server-loaded immutable snapshot, never request-provided lines or
// parties. The older payload generator remains limited to explicit DIAN
// habilitation fixtures and must not be used by the commercial dispatcher.
func generateDIANUBLDesdeFuenteFiscal(cfg map[string]interface{}, empresaID int64, payload map[string]interface{}, snapshot *facturacionFuenteFiscalSnapshot) (map[string]interface{}, int, error) {
	if normalizeFacturacionDocumentoElectronicoTipo(genericStringValue(payload["documento_tipo"])) == "nota_credito" {
		return generateDIANUBLNotaCreditoDesdeFuenteFiscal(cfg, empresaID, payload, snapshot)
	}
	prepared, status, err := prepareDIANUBLDesdeFuenteFiscal(cfg, empresaID, payload, snapshot)
	if err != nil {
		return nil, status, err
	}
	documentoTipo := prepared.DocumentoTipo
	documentoCodigo := prepared.DocumentoCodigo
	emisorNIT := prepared.EmisorNIT
	prefijo := prepared.Prefijo
	resolucionNumero := prepared.ResolucionNumero
	resolucionDesde := prepared.ResolucionDesde
	resolucionHasta := prepared.ResolucionHasta
	llaveTecnica := prepared.LlaveTecnica
	rangoDesde := prepared.RangoDesde
	rangoHasta := prepared.RangoHasta
	softwareID := prepared.SoftwareID
	softwarePIN := prepared.SoftwarePIN
	issueDate := prepared.IssueDate
	issueTime := prepared.IssueTime
	moneda := prepared.Moneda
	profileExecutionID := prepared.ProfileExecutionID

	taxGroups, taxByCode, err := dianFuenteFiscalTaxGroups(snapshot.Lineas)
	if err != nil {
		return nil, http.StatusUnprocessableEntity, err
	}
	lineExtension := facturacionFuenteFiscalRound(snapshot.Totales.BaseGravableLineas)
	taxInclusive := facturacionFuenteFiscalRound(lineExtension + snapshot.Totales.ImpuestoLineas)
	total := facturacionFuenteFiscalRound(snapshot.Totales.TotalDocumentoOrigen)
	if !facturacionFuenteFiscalClose(taxInclusive, total) {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("base mas impuestos no concilia con el total de la fuente fiscal")
	}
	cuFE := buildDIANCUFEFacturaVenta(
		documentoCodigo,
		issueDate,
		issueTime,
		fmt.Sprintf("%.2f", lineExtension),
		fmt.Sprintf("%.2f", taxByCode["01"]),
		fmt.Sprintf("%.2f", taxByCode["04"]),
		fmt.Sprintf("%.2f", taxByCode["03"]),
		fmt.Sprintf("%.2f", total),
		emisorNIT,
		dianNormalizeCustomerDocumentNumber(snapshot.Cliente.NumeroDocumento, snapshot.Cliente.TipoDocumento),
		llaveTecnica,
		profileExecutionID,
	)
	softwareSecurityCode := buildDIANSHA384Hex(softwareID, softwarePIN, documentoCodigo)
	qrURL := "https://catalogo-vpfe-hab.dian.gov.co/Document/FindDocument?documentKey=" + strings.ToLower(cuFE)
	if profileExecutionID == "1" {
		qrURL = "https://catalogo-vpfe.dian.gov.co/Document/FindDocument?documentKey=" + strings.ToLower(cuFE)
	}

	invoiceControl := fmt.Sprintf(`<sts:InvoiceControl><sts:InvoiceAuthorization>%s</sts:InvoiceAuthorization><sts:AuthorizationPeriod><cbc:StartDate>%s</cbc:StartDate><cbc:EndDate>%s</cbc:EndDate></sts:AuthorizationPeriod><sts:AuthorizedInvoices><sts:Prefix>%s</sts:Prefix><sts:From>%d</sts:From><sts:To>%d</sts:To></sts:AuthorizedInvoices></sts:InvoiceControl>`,
		escapeXML(resolucionNumero), escapeXML(resolucionDesde), escapeXML(resolucionHasta), escapeXML(prefijo), rangoDesde, rangoHasta)
	dianExtensions := fmt.Sprintf(`<ext:UBLExtensions><ext:UBLExtension><ext:ExtensionContent><sts:DianExtensions>%s<sts:InvoiceSource><cbc:IdentificationCode listAgencyID="6" listAgencyName="United Nations Economic Commission for Europe" listSchemeURI="urn:oasis:names:specification:ubl:codelist:gc:CountryIdentificationCode-2.1">CO</cbc:IdentificationCode></sts:InvoiceSource><sts:SoftwareProvider><sts:ProviderID schemeAgencyID="195" schemeAgencyName="%s" schemeID="%s" schemeName="31">%s</sts:ProviderID><sts:SoftwareID schemeAgencyID="195" schemeAgencyName="%s">%s</sts:SoftwareID></sts:SoftwareProvider><sts:SoftwareSecurityCode schemeAgencyID="195" schemeAgencyName="%s">%s</sts:SoftwareSecurityCode><sts:AuthorizationProvider><sts:AuthorizationProviderID schemeAgencyID="195" schemeAgencyName="%s" schemeID="4" schemeName="31">800197268</sts:AuthorizationProviderID></sts:AuthorizationProvider><sts:QRCode>NroFactura=%s&#10;NitFacturador=%s&#10;NitAdquiriente=%s&#10;FechaFactura=%s&#10;ValorTotalFactura=%s&#10;CUFE=%s&#10;URL=%s</sts:QRCode></sts:DianExtensions></ext:ExtensionContent></ext:UBLExtension><ext:UBLExtension><ext:ExtensionContent></ext:ExtensionContent></ext:UBLExtension></ext:UBLExtensions>`,
		invoiceControl,
		escapeXML(dianAgencyName), escapeXML(dianCompanyIDSchemeID(emisorNIT, snapshot.Emisor.DigitoVerificacion)), escapeXML(emisorNIT),
		escapeXML(dianAgencyName), escapeXML(softwareID), escapeXML(dianAgencyName), escapeXML(softwareSecurityCode), escapeXML(dianAgencyName),
		escapeXML(documentoCodigo), escapeXML(emisorNIT), escapeXML(dianNormalizeCustomerDocumentNumber(snapshot.Cliente.NumeroDocumento, snapshot.Cliente.TipoDocumento)),
		escapeXML(issueDate), escapeXML(fmt.Sprintf("%.2f", total)), escapeXML(strings.ToLower(cuFE)), escapeXML(qrURL),
	)
	header := fmt.Sprintf(`<cbc:UBLVersionID>UBL 2.1</cbc:UBLVersionID><cbc:CustomizationID>01</cbc:CustomizationID><cbc:ProfileID>DIAN 2.1</cbc:ProfileID><cbc:ProfileExecutionID>%s</cbc:ProfileExecutionID><cbc:ID>%s</cbc:ID><cbc:UUID schemeID="%s" schemeName="CUFE-SHA384">%s</cbc:UUID><cbc:IssueDate>%s</cbc:IssueDate><cbc:IssueTime>%s</cbc:IssueTime><cbc:DueDate>%s</cbc:DueDate><cbc:InvoiceTypeCode>01</cbc:InvoiceTypeCode><cbc:DocumentCurrencyCode listAgencyID="6" listAgencyName="United Nations Economic Commission for Europe" listID="ISO 4217 Alpha">%s</cbc:DocumentCurrencyCode><cbc:LineCountNumeric>%d</cbc:LineCountNumeric>`,
		escapeXML(profileExecutionID), escapeXML(documentoCodigo), escapeXML(profileExecutionID), escapeXML(strings.ToLower(cuFE)), escapeXML(issueDate), escapeXML(issueTime), escapeXML(issueDate), escapeXML(moneda), len(snapshot.Lineas))
	supplierParty, err := dianFuenteFiscalSupplierPartyXML(snapshot.Emisor, prefijo)
	if err != nil {
		return nil, http.StatusUnprocessableEntity, err
	}
	customerParty, err := dianFuenteFiscalCustomerPartyXML(snapshot.Cliente)
	if err != nil {
		return nil, http.StatusUnprocessableEntity, err
	}
	paymentMeans, err := dianFuenteFiscalPaymentMeansXML(snapshot.Pago, issueDate)
	if err != nil {
		return nil, http.StatusUnprocessableEntity, err
	}
	linesXML, err := dianFuenteFiscalLinesXML(snapshot.Lineas, moneda)
	if err != nil {
		return nil, http.StatusUnprocessableEntity, err
	}
	taxesXML := dianFuenteFiscalTaxTotalsXML(taxGroups, moneda)
	monetaryXML := dianFuenteFiscalMonetaryTotalXML(snapshot, moneda)
	xmlPayload := `<?xml version="1.0" encoding="UTF-8" standalone="no"?>` +
		`<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2" xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2" xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2" xmlns:ds="http://www.w3.org/2000/09/xmldsig#" xmlns:ext="urn:oasis:names:specification:ubl:schema:xsd:CommonExtensionComponents-2" xmlns:sts="dian:gov:co:facturaelectronica:Structures-2-1" xmlns:xades="http://uri.etsi.org/01903/v1.3.2#" xmlns:xades141="http://uri.etsi.org/01903/v1.4.1#" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2 http://docs.oasis-open.org/ubl/os-UBL-2.1/xsd/maindoc/UBL-Invoice-2.1.xsd">` +
		dianExtensions + header + supplierParty + customerParty + paymentMeans + taxesXML + monetaryXML + linesXML + `</Invoice>`

	return map[string]interface{}{
		"ok": true, "empresa_id": empresaID, "documento_codigo": documentoCodigo, "documento_tipo": documentoTipo,
		"ubl_version": "UBL 2.1", "profile_execution_id": profileExecutionID, "customization_id": "01",
		"uuid_scheme": "CUFE-SHA384", "uuid": strings.ToLower(cuFE), "software_security_code": "[calculado]",
		"xml_ubl_base": xmlPayload, "estado_preparacion": "pre_envio_validable",
		"fuente_fiscal": map[string]interface{}{"tipo": snapshot.Documento.TipoOrigen, "codigo": snapshot.Documento.CodigoOrigen, "lineas": len(snapshot.Lineas)},
	}, http.StatusOK, nil
}

func validateDIANFuenteFiscalNotaCredito(empresaID int64, payload map[string]interface{}, snapshot *facturacionFuenteFiscalSnapshot) (string, error) {
	if snapshot == nil || snapshot.EmpresaID != empresaID || snapshot.Esquema != facturacionFuenteFiscalEsquema || snapshot.Version != facturacionFuenteFiscalVersion {
		return "", fmt.Errorf("fuente fiscal inmutable de nota credito invalida o de otra empresa")
	}
	if len(snapshot.Bloqueantes) > 0 || snapshot.Referencia == nil || normalizeFacturacionDocumentoElectronicoTipo(snapshot.Documento.TipoOrigen) != "nota_credito" ||
		normalizeFacturacionDocumentoElectronicoTipo(snapshot.Documento.TipoDestino) != "nota_credito" || normalizeFacturacionDocumentoElectronicoTipo(snapshot.Referencia.TipoDocumento) != "factura_electronica" {
		return "", fmt.Errorf("fuente fiscal de nota credito o referencia de factura incompleta")
	}
	if !strings.EqualFold(strings.TrimSpace(snapshot.Documento.CodigoDestino), strings.TrimSpace(genericStringValue(payload["documento_codigo"]))) {
		return "", fmt.Errorf("fuente fiscal no corresponde a la nota credito solicitada")
	}
	numeroLegal := strings.ReplaceAll(strings.TrimSpace(genericStringValue(payload["numero_legal"])), " ", "")
	if numeroLegal == "" || len(snapshot.Lineas) == 0 || snapshot.Totales.TotalDocumentoOrigen <= 0 {
		return "", fmt.Errorf("numero legal, lineas y total son obligatorios para nota credito")
	}
	if !facturacionCodigoSHA384Valido(snapshot.Referencia.CodigoValidacion) || strings.TrimSpace(snapshot.Referencia.NumeroLegal) == "" || strings.TrimSpace(snapshot.Referencia.FechaEmision) == "" {
		return "", fmt.Errorf("CUFE, numero o fecha de factura original invalidos")
	}
	if !strings.EqualFold(snapshot.Emisor.PaisCodigo, "CO") || !strings.EqualFold(snapshot.Cliente.PaisCodigo, "CO") ||
		!facturacionFuenteFiscalDANEDepartamentoValido(snapshot.Emisor.DepartamentoCodigoDANE) ||
		!facturacionFuenteFiscalDANEMunicipioValido(snapshot.Emisor.MunicipioCodigoDANE, snapshot.Emisor.DepartamentoCodigoDANE) ||
		!facturacionFuenteFiscalDANEDepartamentoValido(snapshot.Cliente.DepartamentoCodigoDANE) ||
		!facturacionFuenteFiscalDANEMunicipioValido(snapshot.Cliente.MunicipioCodigoDANE, snapshot.Cliente.DepartamentoCodigoDANE) {
		return "", fmt.Errorf("fuente fiscal de nota credito con pais o codigos DANE invalidos")
	}
	return numeroLegal, nil
}

func generateDIANUBLNotaCreditoDesdeFuenteFiscal(cfg map[string]interface{}, empresaID int64, payload map[string]interface{}, snapshot *facturacionFuenteFiscalSnapshot) (map[string]interface{}, int, error) {
	numeroLegal, err := validateDIANFuenteFiscalNotaCredito(empresaID, payload, snapshot)
	if err != nil {
		return nil, http.StatusUnprocessableEntity, err
	}
	emisorNIT := dianOnlyDigits(snapshot.Emisor.NumeroDocumento)
	if cfgNIT := dianOnlyDigits(genericStringValue(cfg["nit"])); cfgNIT == "" || cfgNIT != emisorNIT {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("NIT DIAN no coincide con el emisor de la nota credito")
	}
	softwareID, softwarePIN, _, credErr := resolveDIANSoftwareCredentials(cfg, nil, empresaID)
	if credErr != nil || strings.TrimSpace(softwareID) == "" || strings.TrimSpace(softwarePIN) == "" {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("credenciales de software DIAN no disponibles")
	}
	fechaEmision := strings.TrimSpace(genericStringValue(payload["fecha_emision"]))
	if fechaEmision == "" {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("fecha de emision firmada es obligatoria")
	}
	issueDate, issueTime := dianIssuepcs_ts(fechaEmision)
	moneda := strings.ToUpper(strings.TrimSpace(snapshot.Documento.Moneda))
	if moneda == "" {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("moneda de nota credito obligatoria")
	}
	profileExecutionID := "2"
	if chooseDIANAmbiente(cfg) == "produccion" {
		profileExecutionID = "1"
	}
	taxGroups, taxByCode, err := dianFuenteFiscalTaxGroups(snapshot.Lineas)
	if err != nil {
		return nil, http.StatusUnprocessableEntity, err
	}
	lineExtension := facturacionFuenteFiscalRound(snapshot.Totales.BaseGravableLineas)
	total := facturacionFuenteFiscalRound(snapshot.Totales.TotalDocumentoOrigen)
	if !facturacionFuenteFiscalClose(lineExtension+snapshot.Totales.ImpuestoLineas, total) {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("base mas impuestos no concilia con el total de la nota credito")
	}
	clienteDocumento := dianNormalizeCustomerDocumentNumber(snapshot.Cliente.NumeroDocumento, snapshot.Cliente.TipoDocumento)
	cude := buildDIANCUFEFacturaVenta(numeroLegal, issueDate, issueTime,
		fmt.Sprintf("%.2f", lineExtension), fmt.Sprintf("%.2f", taxByCode["01"]), fmt.Sprintf("%.2f", taxByCode["04"]),
		fmt.Sprintf("%.2f", taxByCode["03"]), fmt.Sprintf("%.2f", total), emisorNIT, clienteDocumento, softwarePIN, profileExecutionID)
	softwareSecurityCode := buildDIANSHA384Hex(softwareID, softwarePIN, numeroLegal)
	qrURL := "https://catalogo-vpfe-hab.dian.gov.co/Document/FindDocument?documentKey=" + strings.ToLower(cude)
	if profileExecutionID == "1" {
		qrURL = "https://catalogo-vpfe.dian.gov.co/Document/FindDocument?documentKey=" + strings.ToLower(cude)
	}
	dianExtensions := fmt.Sprintf(`<ext:UBLExtensions><ext:UBLExtension><ext:ExtensionContent><sts:DianExtensions><sts:InvoiceSource><cbc:IdentificationCode listAgencyID="6" listAgencyName="United Nations Economic Commission for Europe" listSchemeURI="urn:oasis:names:specification:ubl:codelist:gc:CountryIdentificationCode-2.1">CO</cbc:IdentificationCode></sts:InvoiceSource><sts:SoftwareProvider><sts:ProviderID schemeAgencyID="195" schemeAgencyName="%s" schemeID="%s" schemeName="31">%s</sts:ProviderID><sts:SoftwareID schemeAgencyID="195" schemeAgencyName="%s">%s</sts:SoftwareID></sts:SoftwareProvider><sts:SoftwareSecurityCode schemeAgencyID="195" schemeAgencyName="%s">%s</sts:SoftwareSecurityCode><sts:AuthorizationProvider><sts:AuthorizationProviderID schemeAgencyID="195" schemeAgencyName="%s" schemeID="4" schemeName="31">800197268</sts:AuthorizationProviderID></sts:AuthorizationProvider><sts:QRCode>NroNota=%s&#10;NitFacturador=%s&#10;NitAdquiriente=%s&#10;FechaNota=%s&#10;ValorTotalNota=%s&#10;CUDE=%s&#10;URL=%s</sts:QRCode></sts:DianExtensions></ext:ExtensionContent></ext:UBLExtension><ext:UBLExtension><ext:ExtensionContent></ext:ExtensionContent></ext:UBLExtension></ext:UBLExtensions>`,
		escapeXML(dianAgencyName), escapeXML(dianCompanyIDSchemeID(emisorNIT, snapshot.Emisor.DigitoVerificacion)), escapeXML(emisorNIT),
		escapeXML(dianAgencyName), escapeXML(softwareID), escapeXML(dianAgencyName), escapeXML(softwareSecurityCode), escapeXML(dianAgencyName),
		escapeXML(numeroLegal), escapeXML(emisorNIT), escapeXML(clienteDocumento), escapeXML(issueDate), escapeXML(fmt.Sprintf("%.2f", total)), escapeXML(strings.ToLower(cude)), escapeXML(qrURL))
	header := fmt.Sprintf(`<cbc:UBLVersionID>UBL 2.1</cbc:UBLVersionID><cbc:CustomizationID>20</cbc:CustomizationID><cbc:ProfileID>%s</cbc:ProfileID><cbc:ProfileExecutionID>%s</cbc:ProfileExecutionID><cbc:ID>%s</cbc:ID><cbc:UUID schemeID="%s" schemeName="CUDE-SHA384">%s</cbc:UUID><cbc:IssueDate>%s</cbc:IssueDate><cbc:IssueTime>%s</cbc:IssueTime><cbc:CreditNoteTypeCode>91</cbc:CreditNoteTypeCode><cbc:Note>Anulacion total de factura electronica</cbc:Note><cbc:DocumentCurrencyCode listAgencyID="6" listAgencyName="United Nations Economic Commission for Europe" listID="ISO 4217 Alpha">%s</cbc:DocumentCurrencyCode><cbc:LineCountNumeric>%d</cbc:LineCountNumeric>`,
		escapeXML(dianDocumentProfileID("CreditNote")), escapeXML(profileExecutionID), escapeXML(numeroLegal), escapeXML(profileExecutionID), escapeXML(strings.ToLower(cude)), escapeXML(issueDate), escapeXML(issueTime), escapeXML(moneda), len(snapshot.Lineas))
	correctionCode := dianFirstNonBlank(genericStringValue(payload["codigo_correccion"]), "2")
	correctionDescription := dianFirstNonBlank(genericStringValue(payload["descripcion_correccion"]), "Anulacion de factura electronica")
	referenceDate, _ := dianIssuepcs_ts(snapshot.Referencia.FechaEmision)
	references := fmt.Sprintf(`<cac:DiscrepancyResponse><cbc:ReferenceID>%s</cbc:ReferenceID><cbc:ResponseCode>%s</cbc:ResponseCode><cbc:Description>%s</cbc:Description></cac:DiscrepancyResponse><cac:BillingReference><cac:InvoiceDocumentReference><cbc:ID>%s</cbc:ID><cbc:UUID schemeName="CUFE-SHA384">%s</cbc:UUID><cbc:IssueDate>%s</cbc:IssueDate></cac:InvoiceDocumentReference></cac:BillingReference>`,
		escapeXML(snapshot.Referencia.NumeroLegal), escapeXML(correctionCode), escapeXML(correctionDescription), escapeXML(snapshot.Referencia.NumeroLegal), escapeXML(strings.ToLower(snapshot.Referencia.CodigoValidacion)), escapeXML(referenceDate))
	prefix := dianNotePrefix(numeroLegal, "NC")
	supplierParty, err := dianFuenteFiscalSupplierPartyXML(snapshot.Emisor, prefix)
	if err != nil {
		return nil, http.StatusUnprocessableEntity, err
	}
	customerParty, err := dianFuenteFiscalCustomerPartyXML(snapshot.Cliente)
	if err != nil {
		return nil, http.StatusUnprocessableEntity, err
	}
	paymentMeans, err := dianFuenteFiscalPaymentMeansXML(snapshot.Pago, issueDate)
	if err != nil {
		return nil, http.StatusUnprocessableEntity, err
	}
	linesXML, err := dianFuenteFiscalAdjustmentLinesXML(snapshot.Lineas, moneda, "CreditNoteLine", "CreditedQuantity")
	if err != nil {
		return nil, http.StatusUnprocessableEntity, err
	}
	xmlPayload := `<?xml version="1.0" encoding="UTF-8" standalone="no"?>` +
		`<CreditNote xmlns="urn:oasis:names:specification:ubl:schema:xsd:CreditNote-2" xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2" xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2" xmlns:ds="http://www.w3.org/2000/09/xmldsig#" xmlns:ext="urn:oasis:names:specification:ubl:schema:xsd:CommonExtensionComponents-2" xmlns:sts="dian:gov:co:facturaelectronica:Structures-2-1" xmlns:xades="http://uri.etsi.org/01903/v1.3.2#" xmlns:xades141="http://uri.etsi.org/01903/v1.4.1#" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="urn:oasis:names:specification:ubl:schema:xsd:CreditNote-2 http://docs.oasis-open.org/ubl/os-UBL-2.1/xsd/maindoc/UBL-CreditNote-2.1.xsd">` +
		dianExtensions + header + references + supplierParty + customerParty + paymentMeans + dianFuenteFiscalTaxTotalsXML(taxGroups, moneda) + dianFuenteFiscalMonetaryTotalXML(snapshot, moneda) + linesXML + `</CreditNote>`
	return map[string]interface{}{
		"ok": true, "empresa_id": empresaID, "documento_codigo": numeroLegal, "documento_tipo": "nota_credito",
		"ubl_version": "UBL 2.1", "profile_execution_id": profileExecutionID, "customization_id": "20",
		"uuid_scheme": "CUDE-SHA384", "uuid": strings.ToLower(cude), "software_security_code": "[calculado]",
		"xml_ubl_base": xmlPayload, "estado_preparacion": "pre_envio_validable",
		"fuente_fiscal": map[string]interface{}{"tipo": snapshot.Documento.TipoOrigen, "codigo": snapshot.Documento.CodigoOrigen, "lineas": len(snapshot.Lineas)},
	}, http.StatusOK, nil
}

func prepareDIANUBLDesdeFuenteFiscal(cfg map[string]interface{}, empresaID int64, payload map[string]interface{}, snapshot *facturacionFuenteFiscalSnapshot) (*dianFuenteFiscalUBLContext, int, error) {
	if snapshot == nil || snapshot.EmpresaID != empresaID || snapshot.Esquema != facturacionFuenteFiscalEsquema || snapshot.Version != facturacionFuenteFiscalVersion {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("fuente fiscal inmutable invalida o de otra empresa")
	}
	if len(snapshot.Bloqueantes) > 0 {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("fuente fiscal incompleta: %s", strings.Join(snapshot.Bloqueantes, ", "))
	}
	documentoTipo := normalizeFacturacionDocumentoElectronicoTipo(genericStringValue(payload["documento_tipo"]))
	if documentoTipo != "factura_electronica" || snapshot.Documento.TipoDestino != "factura_electronica" {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("la fuente fiscal de venta solo soporta factura electronica; notas requieren su propia fuente de ajuste")
	}
	if !strings.EqualFold(strings.TrimSpace(snapshot.Documento.CodigoDestino), strings.TrimSpace(genericStringValue(payload["documento_codigo"]))) {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("la fuente fiscal no corresponde al documento comercial solicitado")
	}
	documentoCodigo := strings.ReplaceAll(strings.TrimSpace(genericStringValue(payload["numero_legal"])), " ", "")
	if documentoCodigo == "" {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("numero legal reservado es obligatorio para generar UBL")
	}
	if len(snapshot.Lineas) == 0 || snapshot.Totales.TotalDocumentoOrigen <= 0 {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("fuente fiscal sin lineas o total valido")
	}
	if !strings.EqualFold(snapshot.Emisor.PaisCodigo, "CO") || !strings.EqualFold(snapshot.Cliente.PaisCodigo, "CO") ||
		!facturacionFuenteFiscalDANEDepartamentoValido(snapshot.Emisor.DepartamentoCodigoDANE) ||
		!facturacionFuenteFiscalDANEMunicipioValido(snapshot.Emisor.MunicipioCodigoDANE, snapshot.Emisor.DepartamentoCodigoDANE) ||
		!facturacionFuenteFiscalDANEDepartamentoValido(snapshot.Cliente.DepartamentoCodigoDANE) ||
		!facturacionFuenteFiscalDANEMunicipioValido(snapshot.Cliente.MunicipioCodigoDANE, snapshot.Cliente.DepartamentoCodigoDANE) {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("fuente fiscal con pais o codigos DANE invalidos")
	}
	for _, line := range snapshot.Lineas {
		if strings.TrimSpace(line.CodigoItem) == "" || strings.TrimSpace(line.Descripcion) == "" {
			return nil, http.StatusUnprocessableEntity, fmt.Errorf("linea %d sin identificador o descripcion fiscal", line.Numero)
		}
	}
	if payloadTotal := ventasAnyToFloat64(genericStringValue(payload["total"])); payloadTotal > 0 && !facturacionFuenteFiscalClose(payloadTotal, snapshot.Totales.TotalDocumentoOrigen) {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("total del documento no coincide con la fuente fiscal")
	}

	emisorNIT := dianOnlyDigits(snapshot.Emisor.NumeroDocumento)
	if cfgNIT := dianOnlyDigits(genericStringValue(cfg["nit"])); cfgNIT == "" || cfgNIT != emisorNIT {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("NIT DIAN no coincide con el emisor de la fuente fiscal")
	}
	prefijo := strings.TrimSpace(genericStringValue(cfg["prefijo"]))
	resolucionNumero := strings.TrimSpace(genericStringValue(cfg["resolucion_numero"]))
	resolucionDesde := strings.TrimSpace(genericStringValue(cfg["resolucion_fecha_desde"]))
	resolucionHasta := strings.TrimSpace(genericStringValue(cfg["resolucion_fecha_hasta"]))
	llaveTecnica := dianFirstNonBlank(genericStringValue(cfg["llave_tecnica"]), genericStringValue(cfg["clave_tecnica"]), genericStringValue(cfg["technical_key"]))
	if prefijo == "" || resolucionNumero == "" || resolucionDesde == "" || resolucionHasta == "" || llaveTecnica == "" {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("resolucion, prefijo, vigencia y llave tecnica DIAN son obligatorios")
	}
	if !strings.HasPrefix(strings.ToUpper(documentoCodigo), strings.ToUpper(prefijo)) {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("numero legal no pertenece al prefijo DIAN configurado")
	}
	rangoDesde := anyToInt64(cfg["rango_desde"])
	rangoHasta := anyToInt64(cfg["rango_hasta"])
	if rangoDesde <= 0 || rangoHasta < rangoDesde {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("rango autorizado DIAN invalido")
	}

	softwareID, softwarePIN, _, credErr := resolveDIANSoftwareCredentials(cfg, nil, empresaID)
	if credErr != nil || strings.TrimSpace(softwareID) == "" || strings.TrimSpace(softwarePIN) == "" {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("credenciales de software DIAN no disponibles")
	}
	fechaEmision := strings.TrimSpace(genericStringValue(payload["fecha_emision"]))
	if fechaEmision == "" {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("fecha de emision firmada es obligatoria")
	}
	issueDate, issueTime := dianIssuepcs_ts(fechaEmision)
	moneda := strings.ToUpper(strings.TrimSpace(snapshot.Documento.Moneda))
	if moneda == "" {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("moneda de fuente fiscal es obligatoria")
	}
	profileExecutionID := "2"
	if chooseDIANAmbiente(cfg) == "produccion" {
		profileExecutionID = "1"
	}
	return &dianFuenteFiscalUBLContext{
		DocumentoTipo: documentoTipo, DocumentoCodigo: documentoCodigo, EmisorNIT: emisorNIT,
		Prefijo: prefijo, ResolucionNumero: resolucionNumero, ResolucionDesde: resolucionDesde,
		ResolucionHasta: resolucionHasta, LlaveTecnica: llaveTecnica, RangoDesde: rangoDesde, RangoHasta: rangoHasta,
		SoftwareID: softwareID, SoftwarePIN: softwarePIN, IssueDate: issueDate, IssueTime: issueTime,
		Moneda: moneda, ProfileExecutionID: profileExecutionID,
	}, http.StatusOK, nil
}

func dianFuenteFiscalTaxCode(raw string) (string, string, bool) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "01", "IVA":
		return "01", "IVA", true
	case "04", "INC", "ICO", "CONSUMO":
		return "04", "INC", true
	case "03", "ICA":
		return "03", "ICA", true
	default:
		return "", "", false
	}
}

// dianFuenteFiscalTaxTreatment interprets the existing item tax code without
// manufacturing a tax rate. DIAN reports exempt IVA items with Percent=0 and
// omits TaxTotal for excluded items. A zero IVA rate is therefore explicit
// enough to mean exempt, while excluded items must use an explicit code.
func dianFuenteFiscalTaxTreatment(raw string, percentage float64) (code, name, treatment string, reportTax bool, ok bool) {
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	if percentage < 0 {
		return "", "", "", false, false
	}
	switch normalized {
	case "EXCLUIDO", "IVA_EXCLUIDO":
		if !facturacionFuenteFiscalClose(percentage, 0) {
			return "", "", "", false, false
		}
		return "", "", "excluido", false, true
	case "EXENTO", "IVA_EXENTO":
		if !facturacionFuenteFiscalClose(percentage, 0) {
			return "", "", "", false, false
		}
		return "01", "IVA", "exento", true, true
	}

	code, name, ok = dianFuenteFiscalTaxCode(normalized)
	if !ok {
		return "", "", "", false, false
	}
	if facturacionFuenteFiscalClose(percentage, 0) {
		if code != "01" {
			return "", "", "", false, false
		}
		return code, name, "exento", true, true
	}
	return code, name, "gravado", true, true
}

func dianFuenteFiscalTaxGroups(lines []facturacionFuenteFiscalLinea) ([]dianFuenteFiscalTaxGroup, map[string]float64, error) {
	groups := map[string]*dianFuenteFiscalTaxGroup{}
	byCode := map[string]float64{"01": 0, "04": 0, "03": 0}
	for _, line := range lines {
		code, name, _, reportTax, ok := dianFuenteFiscalTaxTreatment(line.ImpuestoCodigo, line.ImpuestoPorcentaje)
		if !ok {
			return nil, nil, fmt.Errorf("linea %d sin codigo y porcentaje tributario DIAN soportado", line.Numero)
		}
		if !reportTax {
			continue
		}
		key := fmt.Sprintf("%s|%.6f", code, line.ImpuestoPorcentaje)
		group := groups[key]
		if group == nil {
			group = &dianFuenteFiscalTaxGroup{Codigo: code, Nombre: name, Porcentaje: line.ImpuestoPorcentaje}
			groups[key] = group
		}
		group.Base += line.BaseGravable
		group.Impuesto += line.ValorImpuesto
		byCode[code] += line.ValorImpuesto
	}
	out := make([]dianFuenteFiscalTaxGroup, 0, len(groups))
	for _, group := range groups {
		group.Base = facturacionFuenteFiscalRound(group.Base)
		group.Impuesto = facturacionFuenteFiscalRound(group.Impuesto)
		out = append(out, *group)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Codigo != out[j].Codigo {
			return out[i].Codigo < out[j].Codigo
		}
		return out[i].Porcentaje < out[j].Porcentaje
	})
	for code, value := range byCode {
		byCode[code] = facturacionFuenteFiscalRound(value)
	}
	return out, byCode, nil
}

func dianFuenteFiscalAddressXML(part facturacionFuenteFiscalParte) string {
	return fmt.Sprintf(`<cbc:ID>%s</cbc:ID><cbc:CityName>%s</cbc:CityName><cbc:CountrySubentity>%s</cbc:CountrySubentity><cbc:CountrySubentityCode>%s</cbc:CountrySubentityCode><cac:AddressLine><cbc:Line>%s</cbc:Line></cac:AddressLine><cac:Country><cbc:IdentificationCode>%s</cbc:IdentificationCode><cbc:Name languageID="es">Colombia</cbc:Name></cac:Country>`,
		escapeXML(part.MunicipioCodigoDANE), escapeXML(part.Municipio), escapeXML(part.Departamento), escapeXML(part.DepartamentoCodigoDANE), escapeXML(part.Direccion), escapeXML(part.PaisCodigo))
}

func dianFuenteFiscalSupplierPartyXML(part facturacionFuenteFiscalParte, prefijo string) (string, error) {
	nit := dianOnlyDigits(part.NumeroDocumento)
	if nit == "" || part.NombreRazonSocial == "" || part.ResponsabilidadTributaria == "" {
		return "", fmt.Errorf("emisor fiscal incompleto")
	}
	address := dianFuenteFiscalAddressXML(part)
	dv := dianCompanyIDSchemeID(nit, part.DigitoVerificacion)
	contact := ""
	if part.Email != "" || part.Telefono != "" {
		contact = fmt.Sprintf(`<cac:Contact><cbc:Telephone>%s</cbc:Telephone><cbc:ElectronicMail>%s</cbc:ElectronicMail></cac:Contact>`, escapeXML(part.Telefono), escapeXML(part.Email))
	}
	return fmt.Sprintf(`<cac:AccountingSupplierParty><cbc:AdditionalAccountID schemeAgencyID="195">1</cbc:AdditionalAccountID><cac:Party><cac:PartyName><cbc:Name>%s</cbc:Name></cac:PartyName><cac:PhysicalLocation><cac:Address>%s</cac:Address></cac:PhysicalLocation><cac:PartyTaxScheme><cbc:RegistrationName>%s</cbc:RegistrationName><cbc:CompanyID schemeAgencyID="195" schemeAgencyName="%s" schemeID="%s" schemeName="31">%s</cbc:CompanyID><cbc:TaxLevelCode listName="05">%s</cbc:TaxLevelCode><cac:RegistrationAddress>%s</cac:RegistrationAddress><cac:TaxScheme><cbc:ID>01</cbc:ID><cbc:Name>IVA</cbc:Name></cac:TaxScheme></cac:PartyTaxScheme><cac:PartyLegalEntity><cbc:RegistrationName>%s</cbc:RegistrationName><cbc:CompanyID schemeAgencyID="195" schemeAgencyName="%s" schemeID="%s" schemeName="31">%s</cbc:CompanyID><cac:CorporateRegistrationScheme><cbc:ID>%s</cbc:ID></cac:CorporateRegistrationScheme></cac:PartyLegalEntity>%s</cac:Party></cac:AccountingSupplierParty>`,
		escapeXML(part.NombreRazonSocial), address, escapeXML(part.NombreRazonSocial), escapeXML(dianAgencyName), escapeXML(dv), escapeXML(nit), escapeXML(part.ResponsabilidadTributaria), address, escapeXML(part.NombreRazonSocial), escapeXML(dianAgencyName), escapeXML(dv), escapeXML(nit), escapeXML(prefijo), contact), nil
}

func dianFuenteFiscalCustomerPartyXML(part facturacionFuenteFiscalParte) (string, error) {
	document := dianNormalizeCustomerDocumentNumber(part.NumeroDocumento, part.TipoDocumento)
	if document == "" || part.NombreRazonSocial == "" || part.ResponsabilidadTributaria == "" {
		return "", fmt.Errorf("adquiriente fiscal incompleto")
	}
	schemeName := dianCustomerDocumentSchemeName(part.TipoDocumento, document)
	accountID := "2"
	if schemeName == "31" {
		accountID = "1"
	}
	companyAttrs := fmt.Sprintf(`schemeAgencyID="195" schemeAgencyName="%s" schemeName="%s"`, escapeXML(dianAgencyName), escapeXML(schemeName))
	if schemeName == "31" {
		companyAttrs += fmt.Sprintf(` schemeID="%s"`, escapeXML(dianCompanyIDSchemeID(document, part.DigitoVerificacion)))
	}
	address := dianFuenteFiscalAddressXML(part)
	contact := ""
	if part.Email != "" || part.Telefono != "" {
		contact = fmt.Sprintf(`<cac:Contact><cbc:Telephone>%s</cbc:Telephone><cbc:ElectronicMail>%s</cbc:ElectronicMail></cac:Contact>`, escapeXML(part.Telefono), escapeXML(part.Email))
	}
	person := ""
	if accountID == "2" {
		person = dianPersonNameXML(part.NombreRazonSocial)
	}
	taxLevelListName := "04"
	if accountID == "2" {
		taxLevelListName = "49"
	}
	return fmt.Sprintf(`<cac:AccountingCustomerParty><cbc:AdditionalAccountID>%s</cbc:AdditionalAccountID><cac:Party>%s<cac:PartyName><cbc:Name>%s</cbc:Name></cac:PartyName><cac:PhysicalLocation><cac:Address>%s</cac:Address></cac:PhysicalLocation><cac:PartyTaxScheme><cbc:RegistrationName>%s</cbc:RegistrationName><cbc:CompanyID %s>%s</cbc:CompanyID><cbc:TaxLevelCode listName="%s">%s</cbc:TaxLevelCode><cac:RegistrationAddress>%s</cac:RegistrationAddress><cac:TaxScheme><cbc:ID>01</cbc:ID><cbc:Name>IVA</cbc:Name></cac:TaxScheme></cac:PartyTaxScheme><cac:PartyLegalEntity><cbc:RegistrationName>%s</cbc:RegistrationName><cbc:CompanyID %s>%s</cbc:CompanyID></cac:PartyLegalEntity>%s%s</cac:Party></cac:AccountingCustomerParty>`,
		escapeXML(accountID), dianCustomerPartyIdentificationXML(document, schemeName), escapeXML(part.NombreRazonSocial), address, escapeXML(part.NombreRazonSocial), companyAttrs, escapeXML(document), escapeXML(taxLevelListName), escapeXML(part.ResponsabilidadTributaria), address, escapeXML(part.NombreRazonSocial), companyAttrs, escapeXML(document), contact, person), nil
}

func dianFuenteFiscalPaymentMeansXML(payment facturacionFuenteFiscalPago, issueDate string) (string, error) {
	method := strings.ToLower(strings.TrimSpace(payment.Metodo))
	code := ""
	switch method {
	case "efectivo", "cash":
		code = "10"
	case "transferencia", "transferencia_bancaria", "transfer":
		code = "47"
	case "tarjeta", "tarjeta_credito", "tarjeta_debito", "card":
		code = "48"
	default:
		return "", fmt.Errorf("metodo de pago sin codigo DIAN soportado: %s", method)
	}
	return fmt.Sprintf(`<cac:PaymentMeans><cbc:ID>1</cbc:ID><cbc:PaymentMeansCode>%s</cbc:PaymentMeansCode><cbc:PaymentDueDate>%s</cbc:PaymentDueDate><cbc:PaymentID>%s</cbc:PaymentID></cac:PaymentMeans>`, escapeXML(code), escapeXML(issueDate), escapeXML(dianFirstNonBlank(payment.Referencia, "1"))), nil
}

func dianFuenteFiscalUnitCode(raw string) (string, bool) {
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	switch normalized {
	case "EA", "NIU", "UNIDAD", "UNIDADES", "UND":
		return "NIU", true
	case "HUR", "HORA", "HORAS":
		return "HUR", true
	case "KGM", "KG", "KILOGRAMO", "KILOGRAMOS":
		return "KGM", true
	case "MTR", "METRO", "METROS":
		return "MTR", true
	case "LTR", "LITRO", "LITROS":
		return "LTR", true
	default:
		return "", false
	}
}

func dianFuenteFiscalLinesXML(lines []facturacionFuenteFiscalLinea, currency string) (string, error) {
	return dianFuenteFiscalAdjustmentLinesXML(lines, currency, "InvoiceLine", "InvoicedQuantity")
}

func dianFuenteFiscalAdjustmentLinesXML(lines []facturacionFuenteFiscalLinea, currency, lineName, quantityName string) (string, error) {
	if lineName != "InvoiceLine" && lineName != "CreditNoteLine" && lineName != "DebitNoteLine" {
		return "", fmt.Errorf("tipo de linea UBL no soportado")
	}
	if quantityName != "InvoicedQuantity" && quantityName != "CreditedQuantity" && quantityName != "DebitedQuantity" {
		return "", fmt.Errorf("tipo de cantidad UBL no soportado")
	}
	var out strings.Builder
	for _, line := range lines {
		unit, ok := dianFuenteFiscalUnitCode(line.UnidadMedida)
		if !ok {
			return "", fmt.Errorf("linea %d con unidad de medida DIAN no soportada", line.Numero)
		}
		code, taxName, _, reportTax, ok := dianFuenteFiscalTaxTreatment(line.ImpuestoCodigo, line.ImpuestoPorcentaje)
		if !ok {
			return "", fmt.Errorf("linea %d con tratamiento tributario incompleto", line.Numero)
		}
		allowance := ""
		if line.ValorDescuento > 0 {
			allowance = fmt.Sprintf(`<cac:AllowanceCharge><cbc:ID>1</cbc:ID><cbc:ChargeIndicator>false</cbc:ChargeIndicator><cbc:AllowanceChargeReason>Descuento comercial</cbc:AllowanceChargeReason><cbc:MultiplierFactorNumeric>%.6f</cbc:MultiplierFactorNumeric><cbc:Amount currencyID="%s">%.2f</cbc:Amount><cbc:BaseAmount currencyID="%s">%.2f</cbc:BaseAmount></cac:AllowanceCharge>`, line.DescuentoPorcentaje, escapeXML(currency), line.ValorDescuento, escapeXML(currency), facturacionFuenteFiscalRound(line.Cantidad*line.PrecioUnitario))
		}
		tax := ""
		if reportTax {
			tax = fmt.Sprintf(`<cac:TaxTotal><cbc:TaxAmount currencyID="%s">%.2f</cbc:TaxAmount><cac:TaxSubtotal><cbc:TaxableAmount currencyID="%s">%.2f</cbc:TaxableAmount><cbc:TaxAmount currencyID="%s">%.2f</cbc:TaxAmount><cac:TaxCategory><cbc:Percent>%.2f</cbc:Percent><cac:TaxScheme><cbc:ID>%s</cbc:ID><cbc:Name>%s</cbc:Name></cac:TaxScheme></cac:TaxCategory></cac:TaxSubtotal></cac:TaxTotal>`, escapeXML(currency), line.ValorImpuesto, escapeXML(currency), line.BaseGravable, escapeXML(currency), line.ValorImpuesto, line.ImpuestoPorcentaje, escapeXML(code), escapeXML(taxName))
		}
		itemID := strings.TrimSpace(line.CodigoItem)
		if itemID == "" {
			return "", fmt.Errorf("linea %d sin codigo de producto o servicio", line.Numero)
		}
		out.WriteString(fmt.Sprintf(`<cac:%s><cbc:ID>%d</cbc:ID><cbc:%s unitCode="%s">%.6f</cbc:%s><cbc:LineExtensionAmount currencyID="%s">%.2f</cbc:LineExtensionAmount><cbc:FreeOfChargeIndicator>false</cbc:FreeOfChargeIndicator>%s%s<cac:Item><cbc:Description>%s</cbc:Description><cac:SellersItemIdentification><cbc:ID>%s</cbc:ID></cac:SellersItemIdentification></cac:Item><cac:Price><cbc:PriceAmount currencyID="%s">%.2f</cbc:PriceAmount><cbc:BaseQuantity unitCode="%s">1.000000</cbc:BaseQuantity></cac:Price></cac:%s>`,
			lineName, line.Numero, quantityName, escapeXML(unit), line.Cantidad, quantityName, escapeXML(currency), line.BaseGravable, allowance, tax, escapeXML(line.Descripcion), escapeXML(itemID), escapeXML(currency), line.PrecioUnitario, escapeXML(unit), lineName))
	}
	return out.String(), nil
}

func dianFuenteFiscalTaxTotalsXML(groups []dianFuenteFiscalTaxGroup, currency string) string {
	var out strings.Builder
	for _, group := range groups {
		out.WriteString(fmt.Sprintf(`<cac:TaxTotal><cbc:TaxAmount currencyID="%s">%.2f</cbc:TaxAmount><cac:TaxSubtotal><cbc:TaxableAmount currencyID="%s">%.2f</cbc:TaxableAmount><cbc:TaxAmount currencyID="%s">%.2f</cbc:TaxAmount><cac:TaxCategory><cbc:Percent>%.2f</cbc:Percent><cac:TaxScheme><cbc:ID>%s</cbc:ID><cbc:Name>%s</cbc:Name></cac:TaxScheme></cac:TaxCategory></cac:TaxSubtotal></cac:TaxTotal>`,
			escapeXML(currency), group.Impuesto, escapeXML(currency), group.Base, escapeXML(currency), group.Impuesto, group.Porcentaje, escapeXML(group.Codigo), escapeXML(group.Nombre)))
	}
	return out.String()
}

func dianFuenteFiscalMonetaryTotalXML(snapshot *facturacionFuenteFiscalSnapshot, currency string) string {
	lineExtension := facturacionFuenteFiscalRound(snapshot.Totales.BaseGravableLineas)
	taxInclusive := facturacionFuenteFiscalRound(lineExtension + snapshot.Totales.ImpuestoLineas)
	// Los descuentos de esta fuente se expresan dentro de cada InvoiceLine. No
	// se repiten aqui como descuento de cabecera porque se restarian dos veces.
	return fmt.Sprintf(`<cac:LegalMonetaryTotal><cbc:LineExtensionAmount currencyID="%s">%.2f</cbc:LineExtensionAmount><cbc:TaxExclusiveAmount currencyID="%s">%.2f</cbc:TaxExclusiveAmount><cbc:TaxInclusiveAmount currencyID="%s">%.2f</cbc:TaxInclusiveAmount><cbc:AllowanceTotalAmount currencyID="%s">0.00</cbc:AllowanceTotalAmount><cbc:ChargeTotalAmount currencyID="%s">0.00</cbc:ChargeTotalAmount><cbc:PrepaidAmount currencyID="%s">0.00</cbc:PrepaidAmount><cbc:PayableAmount currencyID="%s">%.2f</cbc:PayableAmount></cac:LegalMonetaryTotal>`,
		escapeXML(currency), lineExtension, escapeXML(currency), lineExtension, escapeXML(currency), taxInclusive, escapeXML(currency), escapeXML(currency), escapeXML(currency), escapeXML(currency), snapshot.Totales.TotalDocumentoOrigen)
}

func buildFacturacionFuenteFiscalNotaCreditoTotal(original *facturacionFuenteFiscalSnapshot, nota, factura dbpkg.EmpresaDocumentoFacturacion, fechaReferencia string) (*facturacionFuenteFiscalSnapshot, error) {
	if original == nil || original.EmpresaID <= 0 || original.EmpresaID != nota.EmpresaID || original.EmpresaID != factura.EmpresaID {
		return nil, fmt.Errorf("fuente fiscal y documentos de nota credito no pertenecen a la misma empresa")
	}
	if normalizeFacturacionDocumentoElectronicoTipo(nota.TipoDocumento) != "nota_credito" || normalizeFacturacionDocumentoElectronicoTipo(factura.TipoDocumento) != "factura_electronica" {
		return nil, fmt.Errorf("tipos documentales invalidos para nota credito")
	}
	if !strings.EqualFold(strings.TrimSpace(original.Documento.CodigoDestino), strings.TrimSpace(factura.DocumentoCodigo)) ||
		strings.TrimSpace(nota.DocumentoCodigo) == "" || strings.TrimSpace(nota.NumeroLegal) == "" || strings.TrimSpace(factura.NumeroLegal) == "" ||
		!facturacionCodigoSHA384Valido(factura.CodigoValidacion) || strings.TrimSpace(fechaReferencia) == "" {
		return nil, fmt.Errorf("factura aceptada o referencia fiscal incompleta para nota credito")
	}
	if !facturacionFuenteFiscalClose(nota.MontoTotal, original.Totales.TotalDocumentoOrigen) || !facturacionFuenteFiscalClose(factura.MontoTotal, original.Totales.TotalDocumentoOrigen) {
		return nil, fmt.Errorf("total de nota credito no coincide con la fuente fiscal de la factura")
	}

	snapshot := *original
	snapshot.Lineas = append([]facturacionFuenteFiscalLinea(nil), original.Lineas...)
	snapshot.Bloqueantes = append([]string(nil), original.Bloqueantes...)
	snapshot.Documento = facturacionFuenteFiscalDocumento{
		TipoOrigen: "nota_credito", CodigoOrigen: strings.TrimSpace(nota.DocumentoCodigo),
		TipoDestino: "nota_credito", CodigoDestino: strings.TrimSpace(nota.DocumentoCodigo),
		Fecha: strings.TrimSpace(nota.FechaDocumento), Moneda: strings.ToUpper(strings.TrimSpace(nota.Moneda)),
		MontoTotal: nota.MontoTotal,
	}
	if snapshot.Documento.Fecha == "" {
		snapshot.Documento.Fecha = facturacionNowLocal()
	}
	if snapshot.Documento.Moneda == "" {
		snapshot.Documento.Moneda = strings.ToUpper(strings.TrimSpace(original.Documento.Moneda))
	}
	snapshot.Referencia = &facturacionFuenteFiscalReferencia{
		TipoDocumento: "factura_electronica", DocumentoCodigo: strings.TrimSpace(factura.DocumentoCodigo),
		NumeroLegal: strings.TrimSpace(factura.NumeroLegal), CodigoValidacion: strings.ToLower(strings.TrimSpace(factura.CodigoValidacion)),
		FechaEmision: strings.TrimSpace(fechaReferencia),
	}
	return &snapshot, nil
}

func ensureFacturacionFuenteFiscalNotaCreditoTotal(ctx context.Context, dbEmp *sql.DB, nota, factura dbpkg.EmpresaDocumentoFacturacion) (*facturacionFuenteFiscalSnapshot, *dbpkg.EmpresaFacturacionArtefacto, error) {
	if existing, err := loadFacturacionFuenteFiscalSnapshot(ctx, dbEmp, nota.EmpresaID, "nota_credito", nota.DocumentoCodigo); err == nil {
		artifact, artifactErr := dbpkg.GetEmpresaFacturacionArtefactoByTypeContext(ctx, dbEmp, nota.EmpresaID, "nota_credito", nota.DocumentoCodigo, dbpkg.EmpresaFacturacionArtefactoTipoFuenteFiscalJSON)
		return existing, artifact, artifactErr
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, err
	}
	original, err := loadFacturacionFuenteFiscalParaDocumento(ctx, dbEmp, factura)
	if err != nil {
		return nil, nil, fmt.Errorf("cargar fuente fiscal de factura original: %w", err)
	}
	fechaReferencia := strings.TrimSpace(factura.FechaDocumento)
	if xmlFirmado, xmlErr := loadFacturacionFiscalArtifact(ctx, dbEmp, factura, "xml_firmado"); xmlErr == nil {
		if fechaXML := facturacionDIANFechaEmisionDesdeXML(string(xmlFirmado)); len(fechaXML) >= 10 {
			fechaReferencia = fechaXML[:10]
		}
	} else if !errors.Is(xmlErr, sql.ErrNoRows) {
		return nil, nil, fmt.Errorf("leer XML de factura original: %w", xmlErr)
	}
	snapshot, err := buildFacturacionFuenteFiscalNotaCreditoTotal(original, nota, factura, fechaReferencia)
	if err != nil {
		return nil, nil, err
	}
	artifact, err := saveFacturacionFuenteFiscalSnapshot(ctx, dbEmp, nota, snapshot)
	if err != nil {
		return nil, nil, err
	}
	return snapshot, artifact, nil
}

func ensureFacturacionFuenteFiscalDesdeCarrito(ctx context.Context, dbEmp *sql.DB, carrito *dbpkg.CarritoCompra, cfg *dbpkg.EmpresaConfiguracionAvanzada, doc dbpkg.EmpresaDocumentoFacturacion) (*facturacionFuenteFiscalSnapshot, *dbpkg.EmpresaFacturacionArtefacto, error) {
	if existing, err := loadFacturacionFuenteFiscalSnapshot(ctx, dbEmp, doc.EmpresaID, doc.TipoDocumento, doc.DocumentoCodigo); err == nil {
		artifact, artifactErr := dbpkg.GetEmpresaFacturacionArtefactoByTypeContext(ctx, dbEmp, doc.EmpresaID, doc.TipoDocumento, doc.DocumentoCodigo, dbpkg.EmpresaFacturacionArtefactoTipoFuenteFiscalJSON)
		return existing, artifact, artifactErr
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, err
	}
	items, err := dbpkg.GetCarritoCompraItems(dbEmp, carrito.EmpresaID, carrito.ID, false)
	if err != nil {
		return nil, nil, fmt.Errorf("cargar lineas reales del carrito: %w", err)
	}
	var cliente *dbpkg.Cliente
	if carrito.ClienteID > 0 {
		cliente, err = dbpkg.GetClienteByID(dbEmp, carrito.EmpresaID, carrito.ClienteID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, nil, fmt.Errorf("cargar cliente real de la venta: %w", err)
		}
	}
	snapshot, err := buildFacturacionFuenteFiscalSnapshot(carrito, items, cfg, cliente, doc)
	if err != nil {
		return nil, nil, err
	}
	if len(snapshot.Bloqueantes) > 0 {
		// Never seal an incomplete fiscal source: doing so would make a later
		// correction of required master data impossible. Invoice payments are
		// blocked by the preflight before reaching this point; ordinary receipts
		// may still close without claiming that a fiscal source was persisted.
		return snapshot, nil, nil
	}
	artifact, err := saveFacturacionFuenteFiscalSnapshot(ctx, dbEmp, doc, snapshot)
	if err != nil {
		return nil, nil, err
	}
	return snapshot, artifact, nil
}

// facturacionFuenteFiscalPreflightBloqueantes validates the same server-owned
// source before a cart is paid. It does not persist files, reserve numbering,
// sign XML or contact DIAN.
func facturacionFuenteFiscalPreflightBloqueantes(dbEmp *sql.DB, carrito *dbpkg.CarritoCompra, cfg *dbpkg.EmpresaConfiguracionAvanzada) ([]string, error) {
	if dbEmp == nil || carrito == nil || carrito.EmpresaID <= 0 || carrito.ID <= 0 {
		return nil, fmt.Errorf("carrito invalido para prevalidacion fiscal")
	}
	items, err := dbpkg.GetCarritoCompraItems(dbEmp, carrito.EmpresaID, carrito.ID, false)
	if err != nil {
		return nil, fmt.Errorf("cargar lineas para prevalidacion fiscal: %w", err)
	}
	var cliente *dbpkg.Cliente
	if carrito.ClienteID > 0 {
		cliente, err = dbpkg.GetClienteByID(dbEmp, carrito.EmpresaID, carrito.ClienteID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("cargar cliente para prevalidacion fiscal: %w", err)
		}
	}
	moneda := strings.TrimSpace(carrito.Moneda)
	if moneda == "" {
		moneda = "COP"
	}
	doc := dbpkg.EmpresaDocumentoFacturacion{
		EmpresaID: carrito.EmpresaID, TipoDocumento: "comprobante_pago",
		DocumentoCodigo: fmt.Sprintf("CP-PREFLIGHT-CRT-%d", carrito.ID),
		MontoTotal:      carrito.Total, Moneda: moneda, FechaDocumento: strings.TrimSpace(carrito.PagadoEn),
		EntidadRelacionadaID: carrito.ClienteID,
	}
	snapshot, err := buildFacturacionFuenteFiscalSnapshot(carrito, items, cfg, cliente, doc)
	if err != nil {
		return nil, err
	}
	return append([]string(nil), snapshot.Bloqueantes...), nil
}
