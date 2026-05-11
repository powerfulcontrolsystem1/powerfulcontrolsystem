package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

type EmpresaParqueaderoConfig struct {
	EmpresaID              int64   `json:"empresa_id"`
	Nombre                 string  `json:"nombre"`
	PrefijoTicket          string  `json:"prefijo_ticket"`
	Moneda                 string  `json:"moneda"`
	MinutosTolerancia      int     `json:"minutos_tolerancia"`
	MinutosBase            int     `json:"minutos_base"`
	TarifaBase             float64 `json:"tarifa_base"`
	MinutosFraccion        int     `json:"minutos_fraccion"`
	TarifaFraccion         float64 `json:"tarifa_fraccion"`
	TarifaDiaMax           float64 `json:"tarifa_dia_max"`
	CobrarFraccionCompleta bool    `json:"cobrar_fraccion_completa"`
	IVAIncluido            bool    `json:"iva_incluido"`
	IVAPorcentaje          float64 `json:"iva_porcentaje"`
	SalidaRequiereQR       bool    `json:"salida_requiere_qr"`
	ImprimirTicketEntrada  bool    `json:"imprimir_ticket_entrada"`
	ImprimirReciboSalida   bool    `json:"imprimir_recibo_salida"`
	FechaActualizacion     string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador         string  `json:"usuario_creador,omitempty"`
}

type EmpresaParqueaderoTicket struct {
	ID               int64   `json:"id"`
	EmpresaID        int64   `json:"empresa_id"`
	CodigoTicket     string  `json:"codigo_ticket"`
	Placa            string  `json:"placa"`
	TipoVehiculo     string  `json:"tipo_vehiculo"`
	ClienteID        int64   `json:"cliente_id,omitempty"`
	ServicioID       int64   `json:"servicio_id,omitempty"`
	CarritoID        int64   `json:"carrito_id,omitempty"`
	CarritoItemID    int64   `json:"carrito_item_id,omitempty"`
	ClienteNombre    string  `json:"cliente_nombre,omitempty"`
	ClienteDocumento string  `json:"cliente_documento,omitempty"`
	Estado           string  `json:"estado"`
	FechaEntrada     string  `json:"fecha_entrada"`
	FechaSalida      string  `json:"fecha_salida,omitempty"`
	MinutosCobrados  int     `json:"minutos_cobrados"`
	Subtotal         float64 `json:"subtotal"`
	Impuestos        float64 `json:"impuestos"`
	Total            float64 `json:"total"`
	MetodoPago       string  `json:"metodo_pago,omitempty"`
	QRToken          string  `json:"qr_token,omitempty"`
	QRPayload        string  `json:"qr_payload,omitempty"`
	Observaciones    string  `json:"observaciones,omitempty"`
	UsuarioCreador   string  `json:"usuario_creador,omitempty"`
	UsuarioCierre    string  `json:"usuario_cierre,omitempty"`
}

type EmpresaParqueaderoCobro struct {
	EmpresaID       int64   `json:"empresa_id"`
	TicketID        int64   `json:"ticket_id"`
	CodigoTicket    string  `json:"codigo_ticket"`
	Placa           string  `json:"placa"`
	FechaEntrada    string  `json:"fecha_entrada"`
	FechaSalida     string  `json:"fecha_salida"`
	MinutosReales   int     `json:"minutos_reales"`
	MinutosCobrados int     `json:"minutos_cobrados"`
	Subtotal        float64 `json:"subtotal"`
	Impuestos       float64 `json:"impuestos"`
	Total           float64 `json:"total"`
	Moneda          string  `json:"moneda"`
	Detalle         string  `json:"detalle"`
}

type EmpresaParqueaderoDashboard struct {
	EmpresaID        int64                      `json:"empresa_id"`
	Abiertos         int                        `json:"abiertos"`
	SalidosHoy       int                        `json:"salidos_hoy"`
	AnuladosHoy      int                        `json:"anulados_hoy"`
	IngresosHoy      float64                    `json:"ingresos_hoy"`
	TicketsAbiertos  []EmpresaParqueaderoTicket `json:"tickets_abiertos"`
	SalidasRecientes []EmpresaParqueaderoTicket `json:"salidas_recientes"`
	Config           EmpresaParqueaderoConfig   `json:"config"`
}

type EmpresaParqueaderoIntegracionNucleoResumen struct {
	EmpresaID              int64    `json:"empresa_id"`
	TicketsSincronizados   int      `json:"tickets_sincronizados"`
	TicketsPendientes      int      `json:"tickets_pendientes"`
	ServiciosSincronizados int      `json:"servicios_sincronizados"`
	Errores                []string `json:"errores,omitempty"`
	EstadoIntegracion      string   `json:"estado_integracion"`
	VisibleOperativo       bool     `json:"visible_operativo"`
	RequiereRevisionDatos  bool     `json:"requiere_revision_datos"`
}

func EnsureEmpresaParqueaderoSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_parqueadero_config (
			empresa_id BIGINT PRIMARY KEY,
			nombre TEXT DEFAULT 'Parqueadero',
			prefijo_ticket TEXT DEFAULT 'PK',
			moneda TEXT DEFAULT 'COP',
			minutos_tolerancia INTEGER DEFAULT 10,
			minutos_base INTEGER DEFAULT 60,
			tarifa_base NUMERIC(14,2) DEFAULT 4000,
			minutos_fraccion INTEGER DEFAULT 15,
			tarifa_fraccion NUMERIC(14,2) DEFAULT 1000,
			tarifa_dia_max NUMERIC(14,2) DEFAULT 25000,
			cobrar_fraccion_completa INTEGER DEFAULT 1,
			iva_incluido INTEGER DEFAULT 1,
			iva_porcentaje NUMERIC(7,2) DEFAULT 0,
			salida_requiere_qr INTEGER DEFAULT 1,
			imprimir_ticket_entrada INTEGER DEFAULT 1,
			imprimir_recibo_salida INTEGER DEFAULT 1,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_parqueadero_tickets (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			codigo_ticket TEXT NOT NULL,
			placa TEXT NOT NULL,
			tipo_vehiculo TEXT DEFAULT 'carro',
			cliente_id BIGINT,
			servicio_id BIGINT,
			carrito_id BIGINT,
			carrito_item_id BIGINT,
			cliente_nombre TEXT,
			cliente_documento TEXT,
			estado TEXT DEFAULT 'abierto',
			fecha_entrada TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_salida TEXT,
			minutos_cobrados INTEGER DEFAULT 0,
			subtotal NUMERIC(14,2) DEFAULT 0,
			impuestos NUMERIC(14,2) DEFAULT 0,
			total NUMERIC(14,2) DEFAULT 0,
			metodo_pago TEXT,
			qr_token TEXT,
			qr_payload TEXT,
			observaciones TEXT,
			usuario_creador TEXT,
			usuario_cierre TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_parqueadero_ticket_empresa_codigo ON empresa_parqueadero_tickets(empresa_id, codigo_ticket)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_parqueadero_ticket_empresa_token ON empresa_parqueadero_tickets(empresa_id, qr_token)`,
		`CREATE INDEX IF NOT EXISTS ix_parqueadero_ticket_empresa_estado ON empresa_parqueadero_tickets(empresa_id, estado, id DESC)`,
		`CREATE INDEX IF NOT EXISTS ix_parqueadero_ticket_empresa_placa ON empresa_parqueadero_tickets(empresa_id, placa, estado)`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	columns := []struct {
		name string
		def  string
	}{
		{"cliente_id", "BIGINT"},
		{"servicio_id", "BIGINT"},
		{"carrito_id", "BIGINT"},
		{"carrito_item_id", "BIGINT"},
	}
	for _, column := range columns {
		if err := ensureColumnIfMissing(dbConn, "empresa_parqueadero_tickets", column.name, column.def); err != nil {
			return err
		}
	}
	if _, err := ExecCompat(dbConn, `CREATE INDEX IF NOT EXISTS ix_parqueadero_ticket_empresa_carrito ON empresa_parqueadero_tickets(empresa_id, carrito_id)`); err != nil {
		return err
	}
	return nil
}

func defaultParqueaderoConfig(empresaID int64) EmpresaParqueaderoConfig {
	return EmpresaParqueaderoConfig{
		EmpresaID:              empresaID,
		Nombre:                 "Parqueadero",
		PrefijoTicket:          "PK",
		Moneda:                 "COP",
		MinutosTolerancia:      10,
		MinutosBase:            60,
		TarifaBase:             4000,
		MinutosFraccion:        15,
		TarifaFraccion:         1000,
		TarifaDiaMax:           25000,
		CobrarFraccionCompleta: true,
		IVAIncluido:            true,
		IVAPorcentaje:          0,
		SalidaRequiereQR:       true,
		ImprimirTicketEntrada:  true,
		ImprimirReciboSalida:   true,
	}
}

func GetEmpresaParqueaderoConfig(dbConn *sql.DB, empresaID int64) (EmpresaParqueaderoConfig, error) {
	if err := EnsureEmpresaParqueaderoSchema(dbConn); err != nil {
		return EmpresaParqueaderoConfig{}, err
	}
	cfg := defaultParqueaderoConfig(empresaID)
	var cobrar, ivaIncluido, qr, printEntrada, printSalida int
	err := QueryRowCompat(dbConn, `SELECT empresa_id, COALESCE(nombre,''), COALESCE(prefijo_ticket,'PK'), COALESCE(moneda,'COP'), COALESCE(minutos_tolerancia,10), COALESCE(minutos_base,60), COALESCE(tarifa_base,0), COALESCE(minutos_fraccion,15), COALESCE(tarifa_fraccion,0), COALESCE(tarifa_dia_max,0), COALESCE(cobrar_fraccion_completa,1), COALESCE(iva_incluido,1), COALESCE(iva_porcentaje,0), COALESCE(salida_requiere_qr,1), COALESCE(imprimir_ticket_entrada,1), COALESCE(imprimir_recibo_salida,1), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'') FROM empresa_parqueadero_config WHERE empresa_id = ?`, empresaID).Scan(
		&cfg.EmpresaID, &cfg.Nombre, &cfg.PrefijoTicket, &cfg.Moneda, &cfg.MinutosTolerancia, &cfg.MinutosBase, &cfg.TarifaBase, &cfg.MinutosFraccion, &cfg.TarifaFraccion, &cfg.TarifaDiaMax, &cobrar, &ivaIncluido, &cfg.IVAPorcentaje, &qr, &printEntrada, &printSalida, &cfg.FechaActualizacion, &cfg.UsuarioCreador,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return cfg, nil
		}
		return EmpresaParqueaderoConfig{}, err
	}
	cfg.CobrarFraccionCompleta = cobrar > 0
	cfg.IVAIncluido = ivaIncluido > 0
	cfg.SalidaRequiereQR = qr > 0
	cfg.ImprimirTicketEntrada = printEntrada > 0
	cfg.ImprimirReciboSalida = printSalida > 0
	return normalizeParqueaderoConfig(cfg), nil
}

func UpsertEmpresaParqueaderoConfig(dbConn *sql.DB, cfg EmpresaParqueaderoConfig) error {
	if err := EnsureEmpresaParqueaderoSchema(dbConn); err != nil {
		return err
	}
	cfg = normalizeParqueaderoConfig(cfg)
	_, err := ExecCompat(dbConn, `INSERT INTO empresa_parqueadero_config (
		empresa_id, nombre, prefijo_ticket, moneda, minutos_tolerancia, minutos_base, tarifa_base, minutos_fraccion, tarifa_fraccion, tarifa_dia_max, cobrar_fraccion_completa, iva_incluido, iva_porcentaje, salida_requiere_qr, imprimir_ticket_entrada, imprimir_recibo_salida, fecha_actualizacion, usuario_creador
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?)
	ON CONFLICT (empresa_id) DO UPDATE SET
		nombre = EXCLUDED.nombre,
		prefijo_ticket = EXCLUDED.prefijo_ticket,
		moneda = EXCLUDED.moneda,
		minutos_tolerancia = EXCLUDED.minutos_tolerancia,
		minutos_base = EXCLUDED.minutos_base,
		tarifa_base = EXCLUDED.tarifa_base,
		minutos_fraccion = EXCLUDED.minutos_fraccion,
		tarifa_fraccion = EXCLUDED.tarifa_fraccion,
		tarifa_dia_max = EXCLUDED.tarifa_dia_max,
		cobrar_fraccion_completa = EXCLUDED.cobrar_fraccion_completa,
		iva_incluido = EXCLUDED.iva_incluido,
		iva_porcentaje = EXCLUDED.iva_porcentaje,
		salida_requiere_qr = EXCLUDED.salida_requiere_qr,
		imprimir_ticket_entrada = EXCLUDED.imprimir_ticket_entrada,
		imprimir_recibo_salida = EXCLUDED.imprimir_recibo_salida,
		fecha_actualizacion = CURRENT_TIMESTAMP,
		usuario_creador = EXCLUDED.usuario_creador`,
		cfg.EmpresaID, cfg.Nombre, cfg.PrefijoTicket, cfg.Moneda, cfg.MinutosTolerancia, cfg.MinutosBase, cfg.TarifaBase, cfg.MinutosFraccion, cfg.TarifaFraccion, cfg.TarifaDiaMax, parqueaderoBoolInt(cfg.CobrarFraccionCompleta), parqueaderoBoolInt(cfg.IVAIncluido), cfg.IVAPorcentaje, parqueaderoBoolInt(cfg.SalidaRequiereQR), parqueaderoBoolInt(cfg.ImprimirTicketEntrada), parqueaderoBoolInt(cfg.ImprimirReciboSalida), cfg.UsuarioCreador)
	return err
}

func CreateEmpresaParqueaderoTicket(dbConn *sql.DB, item EmpresaParqueaderoTicket, publicBaseURL string) (EmpresaParqueaderoTicket, error) {
	if err := EnsureEmpresaParqueaderoSchema(dbConn); err != nil {
		return EmpresaParqueaderoTicket{}, err
	}
	if strings.TrimSpace(item.Placa) == "" {
		return EmpresaParqueaderoTicket{}, errors.New("placa es obligatoria")
	}
	open, err := GetEmpresaParqueaderoTicketAbiertoPorPlaca(dbConn, item.EmpresaID, item.Placa)
	if err != nil {
		return EmpresaParqueaderoTicket{}, err
	}
	if open != nil {
		return EmpresaParqueaderoTicket{}, fmt.Errorf("la placa %s ya tiene un ticket abierto", strings.ToUpper(strings.TrimSpace(item.Placa)))
	}
	cfg, err := GetEmpresaParqueaderoConfig(dbConn, item.EmpresaID)
	if err != nil {
		return EmpresaParqueaderoTicket{}, err
	}
	item.Placa = normalizePlate(item.Placa)
	item.TipoVehiculo = normalizeVehicleType(item.TipoVehiculo)
	item.Estado = "abierto"
	item.CodigoTicket, err = nextParqueaderoTicketCode(dbConn, item.EmpresaID, cfg.PrefijoTicket)
	if err != nil {
		return EmpresaParqueaderoTicket{}, err
	}
	item.QRToken = generateParqueaderoToken()
	item.QRPayload = BuildEmpresaParqueaderoQRPayload(publicBaseURL, item.EmpresaID, item.QRToken)
	id, err := insertParqueaderoTicket(dbConn, item)
	if err != nil {
		return EmpresaParqueaderoTicket{}, err
	}
	return GetEmpresaParqueaderoTicketByID(dbConn, item.EmpresaID, id)
}

func ListEmpresaParqueaderoTickets(dbConn *sql.DB, empresaID int64, estado string, limit int) ([]EmpresaParqueaderoTicket, error) {
	if err := EnsureEmpresaParqueaderoSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	args := []interface{}{empresaID}
	where := "empresa_id = ?"
	if strings.TrimSpace(estado) != "" {
		where += " AND LOWER(COALESCE(estado,'')) = ?"
		args = append(args, strings.ToLower(strings.TrimSpace(estado)))
	}
	args = append(args, limit)
	rows, err := ExecQueryCompat(dbConn, `SELECT id, empresa_id, codigo_ticket, placa, COALESCE(tipo_vehiculo,''), COALESCE(cliente_id,0), COALESCE(servicio_id,0), COALESCE(carrito_id,0), COALESCE(carrito_item_id,0), COALESCE(cliente_nombre,''), COALESCE(cliente_documento,''), COALESCE(estado,''), COALESCE(fecha_entrada,''), COALESCE(fecha_salida,''), COALESCE(minutos_cobrados,0), COALESCE(subtotal,0), COALESCE(impuestos,0), COALESCE(total,0), COALESCE(metodo_pago,''), COALESCE(qr_token,''), COALESCE(qr_payload,''), COALESCE(observaciones,''), COALESCE(usuario_creador,''), COALESCE(usuario_cierre,'') FROM empresa_parqueadero_tickets WHERE `+where+` ORDER BY id DESC LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EmpresaParqueaderoTicket, 0)
	for rows.Next() {
		var item EmpresaParqueaderoTicket
		if err := scanParqueaderoTicket(rows, &item); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func GetEmpresaParqueaderoTicketByID(dbConn *sql.DB, empresaID, ticketID int64) (EmpresaParqueaderoTicket, error) {
	var item EmpresaParqueaderoTicket
	err := QueryRowCompat(dbConn, `SELECT id, empresa_id, codigo_ticket, placa, COALESCE(tipo_vehiculo,''), COALESCE(cliente_id,0), COALESCE(servicio_id,0), COALESCE(carrito_id,0), COALESCE(carrito_item_id,0), COALESCE(cliente_nombre,''), COALESCE(cliente_documento,''), COALESCE(estado,''), COALESCE(fecha_entrada,''), COALESCE(fecha_salida,''), COALESCE(minutos_cobrados,0), COALESCE(subtotal,0), COALESCE(impuestos,0), COALESCE(total,0), COALESCE(metodo_pago,''), COALESCE(qr_token,''), COALESCE(qr_payload,''), COALESCE(observaciones,''), COALESCE(usuario_creador,''), COALESCE(usuario_cierre,'') FROM empresa_parqueadero_tickets WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, ticketID).Scan(
		&item.ID, &item.EmpresaID, &item.CodigoTicket, &item.Placa, &item.TipoVehiculo, &item.ClienteID, &item.ServicioID, &item.CarritoID, &item.CarritoItemID, &item.ClienteNombre, &item.ClienteDocumento, &item.Estado, &item.FechaEntrada, &item.FechaSalida, &item.MinutosCobrados, &item.Subtotal, &item.Impuestos, &item.Total, &item.MetodoPago, &item.QRToken, &item.QRPayload, &item.Observaciones, &item.UsuarioCreador, &item.UsuarioCierre,
	)
	return item, err
}

func GetEmpresaParqueaderoTicketByToken(dbConn *sql.DB, empresaID int64, token string) (EmpresaParqueaderoTicket, error) {
	var item EmpresaParqueaderoTicket
	err := QueryRowCompat(dbConn, `SELECT id, empresa_id, codigo_ticket, placa, COALESCE(tipo_vehiculo,''), COALESCE(cliente_id,0), COALESCE(servicio_id,0), COALESCE(carrito_id,0), COALESCE(carrito_item_id,0), COALESCE(cliente_nombre,''), COALESCE(cliente_documento,''), COALESCE(estado,''), COALESCE(fecha_entrada,''), COALESCE(fecha_salida,''), COALESCE(minutos_cobrados,0), COALESCE(subtotal,0), COALESCE(impuestos,0), COALESCE(total,0), COALESCE(metodo_pago,''), COALESCE(qr_token,''), COALESCE(qr_payload,''), COALESCE(observaciones,''), COALESCE(usuario_creador,''), COALESCE(usuario_cierre,'') FROM empresa_parqueadero_tickets WHERE empresa_id = ? AND qr_token = ? LIMIT 1`, empresaID, strings.TrimSpace(token)).Scan(
		&item.ID, &item.EmpresaID, &item.CodigoTicket, &item.Placa, &item.TipoVehiculo, &item.ClienteID, &item.ServicioID, &item.CarritoID, &item.CarritoItemID, &item.ClienteNombre, &item.ClienteDocumento, &item.Estado, &item.FechaEntrada, &item.FechaSalida, &item.MinutosCobrados, &item.Subtotal, &item.Impuestos, &item.Total, &item.MetodoPago, &item.QRToken, &item.QRPayload, &item.Observaciones, &item.UsuarioCreador, &item.UsuarioCierre,
	)
	return item, err
}

func GetEmpresaParqueaderoTicketAbiertoPorPlaca(dbConn *sql.DB, empresaID int64, placa string) (*EmpresaParqueaderoTicket, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT id, empresa_id, codigo_ticket, placa, COALESCE(tipo_vehiculo,''), COALESCE(cliente_id,0), COALESCE(servicio_id,0), COALESCE(carrito_id,0), COALESCE(carrito_item_id,0), COALESCE(cliente_nombre,''), COALESCE(cliente_documento,''), COALESCE(estado,''), COALESCE(fecha_entrada,''), COALESCE(fecha_salida,''), COALESCE(minutos_cobrados,0), COALESCE(subtotal,0), COALESCE(impuestos,0), COALESCE(total,0), COALESCE(metodo_pago,''), COALESCE(qr_token,''), COALESCE(qr_payload,''), COALESCE(observaciones,''), COALESCE(usuario_creador,''), COALESCE(usuario_cierre,'') FROM empresa_parqueadero_tickets WHERE empresa_id = ? AND placa = ? AND LOWER(COALESCE(estado,'')) = 'abierto' ORDER BY id DESC LIMIT 1`, empresaID, normalizePlate(placa))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		var item EmpresaParqueaderoTicket
		if err := scanParqueaderoTicket(rows, &item); err != nil {
			return nil, err
		}
		return &item, nil
	}
	return nil, rows.Err()
}

func CalcularEmpresaParqueaderoCobro(dbConn *sql.DB, empresaID, ticketID int64, salida time.Time) (EmpresaParqueaderoCobro, EmpresaParqueaderoTicket, error) {
	cfg, err := GetEmpresaParqueaderoConfig(dbConn, empresaID)
	if err != nil {
		return EmpresaParqueaderoCobro{}, EmpresaParqueaderoTicket{}, err
	}
	ticket, err := GetEmpresaParqueaderoTicketByID(dbConn, empresaID, ticketID)
	if err != nil {
		return EmpresaParqueaderoCobro{}, EmpresaParqueaderoTicket{}, err
	}
	cobro, err := calcularParqueaderoCobro(cfg, ticket, salida)
	return cobro, ticket, err
}

func parqueaderoCoreCode(prefix string, parts ...string) string {
	var b strings.Builder
	for _, part := range parts {
		for _, r := range strings.ToUpper(strings.TrimSpace(part)) {
			if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				b.WriteRune(r)
				continue
			}
			if b.Len() > 0 && b.String()[b.Len()-1] != '-' {
				b.WriteRune('-')
			}
		}
		if b.Len() > 0 && b.String()[b.Len()-1] != '-' {
			b.WriteRune('-')
		}
	}
	code := strings.Trim(b.String(), "-")
	if code == "" {
		code = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	if len(code) > 42 {
		code = code[:42]
	}
	return strings.Trim(strings.ToUpper(strings.TrimSpace(prefix)), "-") + "-" + strings.Trim(code, "-")
}

func ensureEmpresaParqueaderoCliente(dbConn *sql.DB, ticket EmpresaParqueaderoTicket) (int64, error) {
	if ticket.ClienteID > 0 {
		return ticket.ClienteID, nil
	}
	nombre := strings.TrimSpace(ticket.ClienteNombre)
	documento := strings.TrimSpace(ticket.ClienteDocumento)
	if nombre == "" && documento == "" {
		return 0, nil
	}
	if err := EnsureEmpresaClientesSchema(dbConn); err != nil {
		return 0, err
	}
	normalizedDoc := normalizeClienteDocumentoValue(documento)
	if normalizedDoc != "" {
		query := fmt.Sprintf(`SELECT id FROM clientes WHERE empresa_id = ? AND %s = ? LIMIT 1`, clienteDocumentoSQLExpr("numero_documento"))
		if id, err := findClienteDuplicateID(dbConn, query, ticket.EmpresaID, normalizedDoc); err != nil {
			return 0, err
		} else if id > 0 {
			return id, nil
		}
	}
	tipoDocumento := "CC"
	numeroDocumento := documento
	if numeroDocumento == "" {
		tipoDocumento = "OTRO"
		numeroDocumento = parqueaderoCoreCode("PK-PLACA", ticket.Placa)
	}
	if nombre == "" {
		nombre = "Cliente parqueadero " + normalizePlate(ticket.Placa)
	}
	id, err := CreateCliente(dbConn, Cliente{
		EmpresaID:         ticket.EmpresaID,
		TipoDocumento:     tipoDocumento,
		NumeroDocumento:   numeroDocumento,
		TipoPersona:       "natural",
		NombreRazonSocial: nombre,
		NombreComercial:   nombre,
		Pais:              "CO",
		UsuarioCreador:    ticket.UsuarioCierre,
		Estado:            "activo",
		Observaciones:     "Cliente creado/sincronizado desde parqueadero.",
	})
	if err != nil {
		var dup *ClienteDuplicadoError
		if errors.As(err, &dup) && dup.ClienteID > 0 {
			return dup.ClienteID, nil
		}
		return 0, err
	}
	return id, nil
}

func ensureEmpresaParqueaderoServicio(dbConn *sql.DB, cfg EmpresaParqueaderoConfig, ticket EmpresaParqueaderoTicket) (int64, error) {
	if ticket.ServicioID > 0 {
		return ticket.ServicioID, nil
	}
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		return 0, err
	}
	tipo := normalizeVehicleType(ticket.TipoVehiculo)
	code := parqueaderoCoreCode("PK-SERV", tipo)
	var servicioID int64
	err := queryRowSQLCompat(dbConn, `SELECT id FROM servicios WHERE empresa_id=? AND codigo=? LIMIT 1`, ticket.EmpresaID, code).Scan(&servicioID)
	if err == sql.ErrNoRows {
		servicioID, err = CreateServicio(dbConn, Servicio{
			EmpresaID:       ticket.EmpresaID,
			Codigo:          code,
			Nombre:          "Parqueadero " + tipo,
			Descripcion:     "Servicio vendible para cobro de parqueadero por tiempo.",
			Categoria:       "parqueadero",
			DuracionMinutos: cfg.MinutosBase,
			Precio:          cfg.TarifaBase,
			Estado:          "activo",
			UsuarioCreador:  ticket.UsuarioCierre,
			Observaciones:   "Servicio sincronizado desde tickets de parqueadero.",
		})
	}
	return servicioID, err
}

func createEmpresaParqueaderoCarrito(dbConn *sql.DB, cfg EmpresaParqueaderoConfig, ticket EmpresaParqueaderoTicket, cobro EmpresaParqueaderoCobro, metodoPago string) (int64, int64, int64, error) {
	if cobro.Total <= 0 {
		return 0, 0, ticket.ServicioID, nil
	}
	if err := EnsureEmpresaCarritosSchema(dbConn); err != nil {
		return 0, 0, 0, err
	}
	clienteID, err := ensureEmpresaParqueaderoCliente(dbConn, ticket)
	if err != nil {
		return 0, 0, 0, err
	}
	ticket.ClienteID = clienteID
	servicioID, err := ensureEmpresaParqueaderoServicio(dbConn, cfg, ticket)
	if err != nil {
		return 0, 0, 0, err
	}
	metodo := NormalizeMetodoPagoCarrito(metodoPago)
	if metodo == "" {
		metodo = "efectivo"
	}
	carritoID, err := CreateCarritoCompra(dbConn, CarritoCompra{
		EmpresaID:         ticket.EmpresaID,
		Codigo:            parqueaderoCoreCode("PK-TICKET", fmt.Sprintf("%d", ticket.ID), ticket.CodigoTicket),
		Nombre:            "Parqueadero - " + ticket.Placa,
		CanalVenta:        "parqueadero",
		ClienteID:         clienteID,
		EstadoCarrito:     "abierto",
		Moneda:            cfg.Moneda,
		ReferenciaExterna: fmt.Sprintf("parqueadero:ticket:%d:%s", ticket.ID, ticket.CodigoTicket),
		MetodoPago:        metodo,
		ReferenciaPago:    ticket.CodigoTicket,
		UsuarioCreador:    ticket.UsuarioCierre,
		Observaciones:     "Venta central generada desde ticket de parqueadero.",
	})
	if err != nil {
		return 0, 0, 0, err
	}
	itemID, err := CreateCarritoCompraItem(dbConn, CarritoCompraItem{
		EmpresaID:          ticket.EmpresaID,
		CarritoID:          carritoID,
		TipoItem:           "servicio",
		ReferenciaID:       servicioID,
		CodigoItem:         parqueaderoCoreCode("PK-ITEM", ticket.CodigoTicket),
		Descripcion:        fmt.Sprintf("Parqueadero %s - %s min", ticket.Placa, fmt.Sprintf("%d", cobro.MinutosCobrados)),
		UnidadMedida:       "servicio",
		Cantidad:           1,
		PrecioUnitario:     cobro.Total,
		ImpuestoPorcentaje: 0,
		UsuarioCreador:     ticket.UsuarioCierre,
		Estado:             "activo",
		Observaciones:      cobro.Detalle,
	})
	if err != nil {
		return 0, 0, 0, err
	}
	if err := PayCarritoStationSession(dbConn, ticket.EmpresaID, carritoID, metodo, ticket.CodigoTicket, "", "", 0, 0, cobro.Total, 0, ticket.UsuarioCierre); err != nil {
		return 0, 0, 0, err
	}
	return carritoID, itemID, servicioID, nil
}

func CerrarEmpresaParqueaderoTicket(dbConn *sql.DB, empresaID, ticketID int64, metodoPago, usuario string) (EmpresaParqueaderoTicket, EmpresaParqueaderoCobro, error) {
	cobro, ticket, err := CalcularEmpresaParqueaderoCobro(dbConn, empresaID, ticketID, time.Now())
	if err != nil {
		return EmpresaParqueaderoTicket{}, EmpresaParqueaderoCobro{}, err
	}
	if strings.ToLower(strings.TrimSpace(ticket.Estado)) != "abierto" {
		return EmpresaParqueaderoTicket{}, EmpresaParqueaderoCobro{}, errors.New("el ticket no esta abierto")
	}
	if strings.TrimSpace(metodoPago) == "" {
		metodoPago = "efectivo"
	}
	metodoPago = NormalizeMetodoPagoCarrito(metodoPago)
	if metodoPago == "" {
		metodoPago = "efectivo"
	}
	cfg, err := GetEmpresaParqueaderoConfig(dbConn, empresaID)
	if err != nil {
		return EmpresaParqueaderoTicket{}, EmpresaParqueaderoCobro{}, err
	}
	ticket.UsuarioCierre = strings.TrimSpace(usuario)
	carritoID, itemID, servicioID, err := createEmpresaParqueaderoCarrito(dbConn, cfg, ticket, cobro, metodoPago)
	if err != nil {
		return EmpresaParqueaderoTicket{}, EmpresaParqueaderoCobro{}, err
	}
	clienteID, err := ensureEmpresaParqueaderoCliente(dbConn, ticket)
	if err != nil {
		return EmpresaParqueaderoTicket{}, EmpresaParqueaderoCobro{}, err
	}
	_, err = ExecCompat(dbConn, `UPDATE empresa_parqueadero_tickets SET estado = 'salido', fecha_salida = ?, minutos_cobrados = ?, subtotal = ?, impuestos = ?, total = ?, metodo_pago = ?, cliente_id = ?, servicio_id = ?, carrito_id = ?, carrito_item_id = ?, usuario_cierre = ? WHERE empresa_id = ? AND id = ?`,
		cobro.FechaSalida, cobro.MinutosCobrados, cobro.Subtotal, cobro.Impuestos, cobro.Total, metodoPago, nullableInt64(clienteID), nullableInt64(servicioID), nullableInt64(carritoID), nullableInt64(itemID), strings.TrimSpace(usuario), empresaID, ticketID)
	if err != nil {
		return EmpresaParqueaderoTicket{}, EmpresaParqueaderoCobro{}, err
	}
	closed, err := GetEmpresaParqueaderoTicketByID(dbConn, empresaID, ticketID)
	return closed, cobro, err
}

func SyncEmpresaParqueaderoNucleo(dbConn *sql.DB, empresaID int64, usuario string) (*EmpresaParqueaderoIntegracionNucleoResumen, error) {
	if err := EnsureEmpresaParqueaderoSchema(dbConn); err != nil {
		return nil, err
	}
	cfg, err := GetEmpresaParqueaderoConfig(dbConn, empresaID)
	if err != nil {
		return nil, err
	}
	resumen := &EmpresaParqueaderoIntegracionNucleoResumen{
		EmpresaID:         empresaID,
		EstadoIntegracion: "plantilla_integrada_nucleo",
		VisibleOperativo:  true,
	}
	tickets, err := ListEmpresaParqueaderoTickets(dbConn, empresaID, "salido", 500)
	if err != nil {
		return nil, err
	}
	for _, ticket := range tickets {
		if ticket.CarritoID > 0 || ticket.Total <= 0 {
			resumen.TicketsPendientes++
			continue
		}
		ticket.UsuarioCierre = usuario
		if strings.TrimSpace(ticket.MetodoPago) == "" {
			ticket.MetodoPago = "efectivo"
		}
		cobro := EmpresaParqueaderoCobro{
			EmpresaID:       ticket.EmpresaID,
			TicketID:        ticket.ID,
			CodigoTicket:    ticket.CodigoTicket,
			Placa:           ticket.Placa,
			FechaEntrada:    ticket.FechaEntrada,
			FechaSalida:     ticket.FechaSalida,
			MinutosCobrados: ticket.MinutosCobrados,
			Subtotal:        ticket.Subtotal,
			Impuestos:       ticket.Impuestos,
			Total:           ticket.Total,
			Moneda:          cfg.Moneda,
			Detalle:         "Sincronizacion historica de parqueadero.",
		}
		carritoID, itemID, servicioID, err := createEmpresaParqueaderoCarrito(dbConn, cfg, ticket, cobro, ticket.MetodoPago)
		if err != nil {
			resumen.Errores = append(resumen.Errores, fmt.Sprintf("ticket %d: %v", ticket.ID, err))
			continue
		}
		clienteID, err := ensureEmpresaParqueaderoCliente(dbConn, ticket)
		if err != nil {
			resumen.Errores = append(resumen.Errores, fmt.Sprintf("ticket %d cliente: %v", ticket.ID, err))
			continue
		}
		metodo := NormalizeMetodoPagoCarrito(ticket.MetodoPago)
		if metodo == "" {
			metodo = "efectivo"
		}
		_, err = ExecCompat(dbConn, `UPDATE empresa_parqueadero_tickets SET cliente_id=?, servicio_id=?, carrito_id=?, carrito_item_id=?, metodo_pago=?, usuario_cierre=COALESCE(NULLIF(usuario_cierre,''), ? ) WHERE empresa_id=? AND id=?`,
			nullableInt64(clienteID), nullableInt64(servicioID), nullableInt64(carritoID), nullableInt64(itemID), metodo, strings.TrimSpace(usuario), ticket.EmpresaID, ticket.ID)
		if err != nil {
			resumen.Errores = append(resumen.Errores, fmt.Sprintf("ticket %d refs: %v", ticket.ID, err))
			continue
		}
		resumen.TicketsSincronizados++
		if servicioID > 0 {
			resumen.ServiciosSincronizados++
		}
	}
	if len(resumen.Errores) > 0 {
		resumen.EstadoIntegracion = "integrado_con_observaciones"
		resumen.RequiereRevisionDatos = true
	}
	return resumen, nil
}

func AnularEmpresaParqueaderoTicket(dbConn *sql.DB, empresaID, ticketID int64, usuario, motivo string) error {
	_, err := ExecCompat(dbConn, `UPDATE empresa_parqueadero_tickets SET estado = 'anulado', fecha_salida = CURRENT_TIMESTAMP, observaciones = ?, usuario_cierre = ? WHERE empresa_id = ? AND id = ? AND LOWER(COALESCE(estado,'')) = 'abierto'`, strings.TrimSpace(motivo), strings.TrimSpace(usuario), empresaID, ticketID)
	return err
}

func BuildEmpresaParqueaderoDashboard(dbConn *sql.DB, empresaID int64) (EmpresaParqueaderoDashboard, error) {
	cfg, err := GetEmpresaParqueaderoConfig(dbConn, empresaID)
	if err != nil {
		return EmpresaParqueaderoDashboard{}, err
	}
	abiertos, err := ListEmpresaParqueaderoTickets(dbConn, empresaID, "abierto", 100)
	if err != nil {
		return EmpresaParqueaderoDashboard{}, err
	}
	recientes, err := ListEmpresaParqueaderoTickets(dbConn, empresaID, "salido", 20)
	if err != nil {
		return EmpresaParqueaderoDashboard{}, err
	}
	out := EmpresaParqueaderoDashboard{EmpresaID: empresaID, Abiertos: len(abiertos), TicketsAbiertos: abiertos, SalidasRecientes: recientes, Config: cfg}
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*), COALESCE(SUM(COALESCE(total,0)),0) FROM empresa_parqueadero_tickets WHERE empresa_id = ? AND LOWER(COALESCE(estado,'')) = 'salido' AND CAST(COALESCE(fecha_salida, fecha_entrada) AS TEXT) >= ?`, empresaID, time.Now().Format("2006-01-02")).Scan(&out.SalidosHoy, &out.IngresosHoy)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_parqueadero_tickets WHERE empresa_id = ? AND LOWER(COALESCE(estado,'')) = 'anulado' AND CAST(COALESCE(fecha_salida, fecha_entrada) AS TEXT) >= ?`, empresaID, time.Now().Format("2006-01-02")).Scan(&out.AnuladosHoy)
	return out, nil
}

func BuildEmpresaParqueaderoQRPayload(publicBaseURL string, empresaID int64, token string) string {
	base := strings.TrimSpace(publicBaseURL)
	if base == "" {
		base = "https://powerfulcontrolsystem.com/"
	}
	base = strings.TrimRight(base, "/")
	return fmt.Sprintf("%s/api/public/parqueadero?empresa_id=%d&action=validar_salida&token=%s", base, empresaID, strings.TrimSpace(token))
}

func insertParqueaderoTicket(dbConn *sql.DB, item EmpresaParqueaderoTicket) (int64, error) {
	var id int64
	err := QueryRowCompat(dbConn, `INSERT INTO empresa_parqueadero_tickets (empresa_id, codigo_ticket, placa, tipo_vehiculo, cliente_nombre, cliente_documento, estado, fecha_entrada, qr_token, qr_payload, observaciones, usuario_creador) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?, ?, ?, ?) RETURNING id`,
		item.EmpresaID, item.CodigoTicket, item.Placa, item.TipoVehiculo, strings.TrimSpace(item.ClienteNombre), strings.TrimSpace(item.ClienteDocumento), item.Estado, item.QRToken, item.QRPayload, strings.TrimSpace(item.Observaciones), strings.TrimSpace(item.UsuarioCreador)).Scan(&id)
	return id, err
}

func nextParqueaderoTicketCode(dbConn *sql.DB, empresaID int64, prefix string) (string, error) {
	prefix = strings.ToUpper(strings.TrimSpace(prefix))
	if prefix == "" {
		prefix = "PK"
	}
	day := time.Now().Format("20060102")
	like := prefix + "-" + day + "-%"
	var count int
	if err := QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_parqueadero_tickets WHERE empresa_id = ? AND codigo_ticket LIKE ?`, empresaID, like).Scan(&count); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s-%04d", prefix, day, count+1), nil
}

func calcularParqueaderoCobro(cfg EmpresaParqueaderoConfig, ticket EmpresaParqueaderoTicket, salida time.Time) (EmpresaParqueaderoCobro, error) {
	entrada, err := parseParqueaderoTime(ticket.FechaEntrada)
	if err != nil {
		return EmpresaParqueaderoCobro{}, err
	}
	if salida.IsZero() {
		salida = time.Now()
	}
	minutosReales := int(math.Ceil(salida.Sub(entrada).Minutes()))
	if minutosReales < 0 {
		minutosReales = 0
	}
	cobrables := minutosReales - cfg.MinutosTolerancia
	if cobrables < 0 {
		cobrables = 0
	}
	subtotal := 0.0
	detalle := "Dentro de tolerancia"
	if cobrables > 0 {
		subtotal = cfg.TarifaBase
		detalle = fmt.Sprintf("Base %d min", cfg.MinutosBase)
		if cobrables > cfg.MinutosBase {
			extra := cobrables - cfg.MinutosBase
			fracciones := float64(extra) / float64(cfg.MinutosFraccion)
			if cfg.CobrarFraccionCompleta {
				fracciones = math.Ceil(fracciones)
			}
			if fracciones < 0 {
				fracciones = 0
			}
			subtotal += fracciones * cfg.TarifaFraccion
			detalle = fmt.Sprintf("Base %d min + %.2f fracciones de %d min", cfg.MinutosBase, fracciones, cfg.MinutosFraccion)
		}
		if cfg.TarifaDiaMax > 0 {
			days := math.Ceil(float64(cobrables) / 1440.0)
			capValue := days * cfg.TarifaDiaMax
			if capValue > 0 && subtotal > capValue {
				subtotal = capValue
				detalle = fmt.Sprintf("Tope diario aplicado (%.0f dia/s)", days)
			}
		}
	}
	impuestos := 0.0
	if cfg.IVAPorcentaje > 0 && !cfg.IVAIncluido {
		impuestos = subtotal * cfg.IVAPorcentaje / 100
	}
	total := roundMoney(subtotal + impuestos)
	return EmpresaParqueaderoCobro{
		EmpresaID:       ticket.EmpresaID,
		TicketID:        ticket.ID,
		CodigoTicket:    ticket.CodigoTicket,
		Placa:           ticket.Placa,
		FechaEntrada:    ticket.FechaEntrada,
		FechaSalida:     salida.Format("2006-01-02 15:04:05"),
		MinutosReales:   minutosReales,
		MinutosCobrados: cobrables,
		Subtotal:        roundMoney(subtotal),
		Impuestos:       roundMoney(impuestos),
		Total:           total,
		Moneda:          cfg.Moneda,
		Detalle:         detalle,
	}, nil
}

func normalizeParqueaderoConfig(cfg EmpresaParqueaderoConfig) EmpresaParqueaderoConfig {
	cfg.Nombre = strings.TrimSpace(cfg.Nombre)
	if cfg.Nombre == "" {
		cfg.Nombre = "Parqueadero"
	}
	cfg.PrefijoTicket = strings.ToUpper(strings.TrimSpace(cfg.PrefijoTicket))
	if cfg.PrefijoTicket == "" {
		cfg.PrefijoTicket = "PK"
	}
	cfg.Moneda = strings.ToUpper(strings.TrimSpace(cfg.Moneda))
	if cfg.Moneda == "" {
		cfg.Moneda = "COP"
	}
	if cfg.MinutosTolerancia < 0 {
		cfg.MinutosTolerancia = 0
	}
	if cfg.MinutosBase <= 0 {
		cfg.MinutosBase = 60
	}
	if cfg.MinutosFraccion <= 0 {
		cfg.MinutosFraccion = 15
	}
	if cfg.TarifaBase < 0 {
		cfg.TarifaBase = 0
	}
	if cfg.TarifaFraccion < 0 {
		cfg.TarifaFraccion = 0
	}
	if cfg.TarifaDiaMax < 0 {
		cfg.TarifaDiaMax = 0
	}
	if cfg.IVAPorcentaje < 0 {
		cfg.IVAPorcentaje = 0
	}
	return cfg
}

func normalizePlate(raw string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(raw), " ", ""))
}

func normalizeVehicleType(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "moto", "camioneta", "camion", "bicicleta", "carro":
		return v
	default:
		return "carro"
	}
}

func parqueaderoBoolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func roundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}

func generateParqueaderoToken() string {
	buf := make([]byte, 18)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("pk-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}

func parseParqueaderoTime(raw string) (time.Time, error) {
	value := strings.TrimSpace(raw)
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999-07",
		"2006-01-02 15:04:05.999999Z07:00",
		"2006-01-02 15:04:05.999999",
		"2006-01-02 15:04:05-07",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("fecha de entrada invalida: %s", value)
}

type parqueaderoScanner interface {
	Scan(dest ...interface{}) error
}

func scanParqueaderoTicket(row parqueaderoScanner, item *EmpresaParqueaderoTicket) error {
	return row.Scan(&item.ID, &item.EmpresaID, &item.CodigoTicket, &item.Placa, &item.TipoVehiculo, &item.ClienteID, &item.ServicioID, &item.CarritoID, &item.CarritoItemID, &item.ClienteNombre, &item.ClienteDocumento, &item.Estado, &item.FechaEntrada, &item.FechaSalida, &item.MinutosCobrados, &item.Subtotal, &item.Impuestos, &item.Total, &item.MetodoPago, &item.QRToken, &item.QRPayload, &item.Observaciones, &item.UsuarioCreador, &item.UsuarioCierre)
}
