package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

type facturacionContingenciaRequest struct {
	EmpresaID            int64  `json:"empresa_id"`
	Action               string `json:"action"`
	Tipo                 string `json:"tipo"`
	ContingenciaID       int64  `json:"contingencia_id"`
	Confirmacion         string `json:"confirmacion"`
	Motivo               string `json:"motivo"`
	EvidenciaReferencia  string `json:"evidencia_referencia"`
	PaisCodigo           string `json:"pais_codigo"`
	Prefijo              string `json:"prefijo"`
	ResolucionNumero     string `json:"resolucion_numero"`
	FechaDesde           string `json:"fecha_desde"`
	FechaHasta           string `json:"fecha_hasta"`
	RangoDesde           int64  `json:"rango_desde"`
	RangoHasta           int64  `json:"rango_hasta"`
	ProximoNumero        int64  `json:"proximo_numero"`
	Estado               string `json:"estado"`
	Observaciones        string `json:"observaciones"`
	CarritoID            int64  `json:"carrito_id"`
	NumeroPapel          string `json:"numero_papel"`
	FechaExpedicionPapel string `json:"fecha_expedicion_papel"`
}

func EmpresaFacturacionContingenciasHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := dbpkg.EmpresaFacturacionContingenciasSchemaReady(dbEmp); err != nil {
			http.Error(w, "El flujo de contingencia fiscal requiere ejecutar la migracion vigente", http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeFacturacionContingenciaState(w, dbEmp, empresaID)
		case http.MethodPost:
			r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
			decoder := json.NewDecoder(r.Body)
			decoder.DisallowUnknownFields()
			var input facturacionContingenciaRequest
			if err := decoder.Decode(&input); err != nil {
				http.Error(w, "solicitud de contingencia invalida", http.StatusBadRequest)
				return
			}
			if input.EmpresaID <= 0 {
				http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
				return
			}
			usuario := strings.TrimSpace(adminEmailFromRequest(r))
			action := strings.ToLower(strings.TrimSpace(input.Action))
			var err error
			switch action {
			case "guardar_talonario":
				_, err = dbpkg.UpsertEmpresaFacturacionContingenciaConfiguracion(dbEmp, dbpkg.EmpresaFacturacionContingenciaConfiguracion{
					EmpresaID: input.EmpresaID, PaisCodigo: input.PaisCodigo, Prefijo: input.Prefijo,
					ResolucionNumero: input.ResolucionNumero, FechaDesde: input.FechaDesde, FechaHasta: input.FechaHasta,
					RangoDesde: input.RangoDesde, RangoHasta: input.RangoHasta, ProximoNumero: input.ProximoNumero,
					Estado: input.Estado, UsuarioCreador: usuario, Observaciones: input.Observaciones,
				})
			case "abrir":
				if input.Confirmacion != "ACTIVAR CONTINGENCIA" {
					err = fmt.Errorf("escribe ACTIVAR CONTINGENCIA exactamente")
					break
				}
				if strings.EqualFold(strings.TrimSpace(input.Tipo), dbpkg.FacturacionContingenciaFallaDIAN) {
					cfg, cfgErr := dbpkg.GetFacturacionElectronicaPaisConfig(dbEmp, input.EmpresaID, "CO")
					if cfgErr != nil || cfg == nil || !strings.EqualFold(cfg.Ambiente, "produccion") || !strings.EqualFold(cfg.Proveedor, "dian") || !strings.EqualFold(cfg.Estado, "activo") {
						err = fmt.Errorf("la contingencia DIAN exige integracion Colombia/DIAN activa en produccion")
						break
					}
				}
				_, err = dbpkg.OpenEmpresaFacturacionContingencia(dbEmp, input.EmpresaID, input.Tipo, input.Motivo, input.EvidenciaReferencia, usuario, input.Observaciones)
			case "recuperar":
				if input.Confirmacion != "RECUPERAR SERVICIO" {
					err = fmt.Errorf("escribe RECUPERAR SERVICIO exactamente")
					break
				}
				err = dbpkg.RecoverEmpresaFacturacionContingencia(dbEmp, input.EmpresaID, input.ContingenciaID, usuario)
			case "registrar_talonario":
				if input.Confirmacion != "REGISTRAR TALONARIO" {
					err = fmt.Errorf("escribe REGISTRAR TALONARIO exactamente")
					break
				}
				carrito, cartErr := dbpkg.GetCarritoCompraByID(dbEmp, input.EmpresaID, input.CarritoID)
				if cartErr != nil || !isCarritoVentaPagada(carrito) {
					err = fmt.Errorf("el carrito no existe, pertenece a otra empresa o no esta pagado")
					break
				}
				comprobanteCodigo := buildVentaDocumentoCodigo(carrito, "comprobante_pago")
				fuente, sourceErr := loadFacturacionFuenteFiscalSnapshot(r.Context(), dbEmp, input.EmpresaID, "comprobante_pago", comprobanteCodigo)
				if sourceErr != nil || fuente == nil || fuente.Carrito.ID != input.CarritoID {
					err = fmt.Errorf("el comprobante no tiene una fuente fiscal inmutable del mismo carrito")
					break
				}
				_, err = dbpkg.RegisterEmpresaFacturacionTalonarioSale(dbEmp, input.EmpresaID, input.ContingenciaID, input.CarritoID, comprobanteCodigo, input.NumeroPapel, input.FechaExpedicionPapel, usuario)
			case "cerrar":
				if input.Confirmacion != "CERRAR CONTINGENCIA" {
					err = fmt.Errorf("escribe CERRAR CONTINGENCIA exactamente")
					break
				}
				err = dbpkg.CloseEmpresaFacturacionContingencia(dbEmp, input.EmpresaID, input.ContingenciaID, usuario)
			default:
				err = fmt.Errorf("accion de contingencia no soportada")
			}
			if err != nil {
				status := http.StatusConflict
				if errors.Is(err, sql.ErrNoRows) {
					status = http.StatusNotFound
				}
				http.Error(w, err.Error(), status)
				return
			}
			metadata, _ := json.Marshal(map[string]interface{}{"tipo": strings.TrimSpace(input.Tipo), "contingencia_id": input.ContingenciaID})
			_, _ = dbpkg.CreateEmpresaAuditoriaEvento(dbEmp, dbpkg.EmpresaAuditoriaEvento{
				EmpresaID: input.EmpresaID, Modulo: "facturacion", Accion: "contingencia_" + action,
				Recurso: "contingencia_fiscal", RecursoID: input.ContingenciaID, MetodoHTTP: r.Method,
				Endpoint: r.URL.Path, Resultado: "ok", CodigoHTTP: http.StatusOK, MetadataJSON: string(metadata),
				UsuarioCreador: usuario, Estado: "activo", Observaciones: "operacion de contingencia fiscal auditada",
			})
			writeFacturacionContingenciaState(w, dbEmp, input.EmpresaID)
		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		}
	}
}

func writeFacturacionContingenciaState(w http.ResponseWriter, dbEmp *sql.DB, empresaID int64) {
	var config interface{}
	if item, err := dbpkg.GetEmpresaFacturacionContingenciaConfiguracion(dbEmp, empresaID); err == nil {
		config = item
	} else if !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "No se pudo consultar la autorizacion de talonario", http.StatusInternalServerError)
		return
	}
	items, err := dbpkg.ListEmpresaFacturacionContingencias(dbEmp, empresaID, 30)
	if err != nil {
		http.Error(w, "No se pudo consultar el historial de contingencias", http.StatusInternalServerError)
		return
	}
	documents, err := dbpkg.ListEmpresaFacturacionContingenciaDocumentos(dbEmp, empresaID, 100)
	if err != nil {
		http.Error(w, "No se pudo consultar el detalle de documentos de contingencia", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok": true, "empresa_id": empresaID, "pais_codigo": "CO", "configuracion_talonario": config,
		"contingencias": items, "documentos": documents,
		"reglas": map[string]interface{}{
			"offline_es_contingencia":                        false,
			"falla_facturador_requiere_talonario_autorizado": true,
			"transmision_tipo_03_requerida":                  true,
			"plazo_operativo_horas_despues_recuperacion":     48,
		},
	})
}
