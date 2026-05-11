package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

type EmpresaOdontologiaPaciente struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	ClienteID          int64   `json:"cliente_id,omitempty"`
	Codigo             string  `json:"codigo"`
	NombreCompleto     string  `json:"nombre_completo"`
	Documento          string  `json:"documento,omitempty"`
	Telefono           string  `json:"telefono,omitempty"`
	Email              string  `json:"email,omitempty"`
	FechaNacimiento    string  `json:"fecha_nacimiento,omitempty"`
	Genero             string  `json:"genero,omitempty"`
	Aseguradora        string  `json:"aseguradora,omitempty"`
	Alergias           string  `json:"alergias,omitempty"`
	RiesgoMedico       string  `json:"riesgo_medico,omitempty"`
	UltimaVisita       string  `json:"ultima_visita,omitempty"`
	Saldo              float64 `json:"saldo"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
}

type EmpresaOdontologiaProfesional struct {
	ID                  int64  `json:"id"`
	EmpresaID           int64  `json:"empresa_id"`
	NombreCompleto      string `json:"nombre_completo"`
	Especialidad        string `json:"especialidad,omitempty"`
	RegistroProfesional string `json:"registro_profesional,omitempty"`
	Telefono            string `json:"telefono,omitempty"`
	Email               string `json:"email,omitempty"`
	ColorAgenda         string `json:"color_agenda,omitempty"`
	Estado              string `json:"estado,omitempty"`
	Observaciones       string `json:"observaciones,omitempty"`
	FechaCreacion       string `json:"fecha_creacion,omitempty"`
	FechaActualizacion  string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador      string `json:"usuario_creador,omitempty"`
}

type EmpresaOdontologiaConsultorio struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Nombre             string `json:"nombre"`
	Sede               string `json:"sede,omitempty"`
	Sillon             string `json:"sillon,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
}

type EmpresaOdontologiaCita struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	PacienteID         int64  `json:"paciente_id"`
	PacienteNombre     string `json:"paciente_nombre,omitempty"`
	ProfesionalID      int64  `json:"profesional_id"`
	ProfesionalNombre  string `json:"profesional_nombre,omitempty"`
	ConsultorioID      int64  `json:"consultorio_id,omitempty"`
	ConsultorioNombre  string `json:"consultorio_nombre,omitempty"`
	FechaHora          string `json:"fecha_hora,omitempty"`
	DuracionMinutos    int    `json:"duracion_minutos"`
	Motivo             string `json:"motivo,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Canal              string `json:"canal,omitempty"`
	Prioridad          string `json:"prioridad,omitempty"`
	Aseguradora        string `json:"aseguradora,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
}

type EmpresaOdontologiaHistoria struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	PacienteID         int64  `json:"paciente_id"`
	PacienteNombre     string `json:"paciente_nombre,omitempty"`
	ProfesionalID      int64  `json:"profesional_id,omitempty"`
	ProfesionalNombre  string `json:"profesional_nombre,omitempty"`
	CitaID             int64  `json:"cita_id,omitempty"`
	FechaAtencion      string `json:"fecha_atencion,omitempty"`
	MotivoConsulta     string `json:"motivo_consulta,omitempty"`
	Diagnostico        string `json:"diagnostico,omitempty"`
	PlanTratamiento    string `json:"plan_tratamiento,omitempty"`
	Evolucion          string `json:"evolucion,omitempty"`
	Formula            string `json:"formula,omitempty"`
	Recomendaciones    string `json:"recomendaciones,omitempty"`
	ProximaCita        string `json:"proxima_cita,omitempty"`
	Estado             string `json:"estado,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
}

type EmpresaOdontologiaOdontograma struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	PacienteID         int64  `json:"paciente_id"`
	PacienteNombre     string `json:"paciente_nombre,omitempty"`
	ProfesionalID      int64  `json:"profesional_id,omitempty"`
	ProfesionalNombre  string `json:"profesional_nombre,omitempty"`
	FechaRegistro      string `json:"fecha_registro,omitempty"`
	PiezasJSON         string `json:"piezas_json,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
	Estado             string `json:"estado,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
}

type EmpresaOdontologiaTratamiento struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	ServicioID         int64   `json:"servicio_id,omitempty"`
	PacienteID         int64   `json:"paciente_id"`
	PacienteNombre     string  `json:"paciente_nombre,omitempty"`
	ProfesionalID      int64   `json:"profesional_id,omitempty"`
	ProfesionalNombre  string  `json:"profesional_nombre,omitempty"`
	Nombre             string  `json:"nombre"`
	Categoria          string  `json:"categoria,omitempty"`
	Piezas             string  `json:"piezas,omitempty"`
	SesionesTotal      int     `json:"sesiones_total"`
	SesionesRealizadas int     `json:"sesiones_realizadas"`
	CostoEstimado      float64 `json:"costo_estimado"`
	CostoReal          float64 `json:"costo_real"`
	FechaInicio        string  `json:"fecha_inicio,omitempty"`
	FechaFin           string  `json:"fecha_fin,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
}

type EmpresaOdontologiaPresupuesto struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	PacienteID         int64   `json:"paciente_id"`
	PacienteNombre     string  `json:"paciente_nombre,omitempty"`
	TratamientoID      int64   `json:"tratamiento_id,omitempty"`
	TratamientoNombre  string  `json:"tratamiento_nombre,omitempty"`
	Nombre             string  `json:"nombre"`
	ValorTotal         float64 `json:"valor_total"`
	CuotaInicial       float64 `json:"cuota_inicial"`
	Saldo              float64 `json:"saldo"`
	Estado             string  `json:"estado,omitempty"`
	VigenciaHasta      string  `json:"vigencia_hasta,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
}

type EmpresaOdontologiaPago struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	PacienteID         int64   `json:"paciente_id,omitempty"`
	PacienteNombre     string  `json:"paciente_nombre,omitempty"`
	ClienteID          int64   `json:"cliente_id,omitempty"`
	PresupuestoID      int64   `json:"presupuesto_id,omitempty"`
	PresupuestoNombre  string  `json:"presupuesto_nombre,omitempty"`
	ServicioID         int64   `json:"servicio_id,omitempty"`
	CarritoID          int64   `json:"carrito_id,omitempty"`
	CarritoItemID      int64   `json:"carrito_item_id,omitempty"`
	Concepto           string  `json:"concepto"`
	Monto              float64 `json:"monto"`
	MetodoPago         string  `json:"metodo_pago,omitempty"`
	Referencia         string  `json:"referencia,omitempty"`
	FechaPago          string  `json:"fecha_pago,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
}

type EmpresaOdontologiaDashboard struct {
	EmpresaID             int64                           `json:"empresa_id"`
	PacientesActivos      int                             `json:"pacientes_activos"`
	ProfesionalesActivos  int                             `json:"profesionales_activos"`
	CitasHoy              int                             `json:"citas_hoy"`
	CitasPendientes       int                             `json:"citas_pendientes"`
	TratamientosActivos   int                             `json:"tratamientos_activos"`
	PresupuestosVigentes  int                             `json:"presupuestos_vigentes"`
	RecaudoMes            float64                         `json:"recaudo_mes"`
	SaldoPendiente        float64                         `json:"saldo_pendiente"`
	AgendaHoy             []EmpresaOdontologiaCita        `json:"agenda_hoy"`
	TratamientosPrioridad []EmpresaOdontologiaTratamiento `json:"tratamientos_prioridad"`
}

type EmpresaOdontologiaIntegracionNucleoResumen struct {
	EmpresaID                 int64    `json:"empresa_id"`
	PacientesSincronizados    int      `json:"pacientes_sincronizados"`
	TratamientosSincronizados int      `json:"tratamientos_sincronizados"`
	PagosSincronizados        int      `json:"pagos_sincronizados"`
	PagosPendientes           int      `json:"pagos_pendientes"`
	Errores                   []string `json:"errores,omitempty"`
	EstadoIntegracion         string   `json:"estado_integracion"`
	VisibleOperativo          bool     `json:"visible_operativo"`
	RequiereRevisionDatos     bool     `json:"requiere_revision_datos"`
}

var (
	empresaOdontologiaSchemaEnsured sync.Map
	empresaOdontologiaSchemaMu      sync.Mutex
)

func EnsureEmpresaOdontologiaSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("dbConn es obligatorio")
	}
	cacheKey := fmt.Sprintf("%p", dbConn)
	if _, ok := empresaOdontologiaSchemaEnsured.Load(cacheKey); ok {
		return nil
	}
	empresaOdontologiaSchemaMu.Lock()
	defer empresaOdontologiaSchemaMu.Unlock()
	if _, ok := empresaOdontologiaSchemaEnsured.Load(cacheKey); ok {
		return nil
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_odontologia_pacientes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			cliente_id INTEGER,
			codigo TEXT,
			nombre_completo TEXT NOT NULL,
			documento TEXT,
			telefono TEXT,
			email TEXT,
			fecha_nacimiento TEXT,
			genero TEXT,
			aseguradora TEXT,
			alergias TEXT,
			riesgo_medico TEXT,
			ultima_visita TEXT,
			saldo REAL DEFAULT 0,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_odontologia_pacientes_empresa ON empresa_odontologia_pacientes(empresa_id, estado, id DESC);`,
		`CREATE TABLE IF NOT EXISTS empresa_odontologia_profesionales (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			nombre_completo TEXT NOT NULL,
			especialidad TEXT,
			registro_profesional TEXT,
			telefono TEXT,
			email TEXT,
			color_agenda TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_odontologia_profesionales_empresa ON empresa_odontologia_profesionales(empresa_id, estado, id DESC);`,
		`CREATE TABLE IF NOT EXISTS empresa_odontologia_consultorios (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			nombre TEXT NOT NULL,
			sede TEXT,
			sillon TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_odontologia_consultorios_empresa ON empresa_odontologia_consultorios(empresa_id, estado, id DESC);`,
		`CREATE TABLE IF NOT EXISTS empresa_odontologia_citas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			paciente_id INTEGER NOT NULL,
			profesional_id INTEGER NOT NULL,
			consultorio_id INTEGER DEFAULT 0,
			fecha_hora TEXT NOT NULL,
			duracion_minutos INTEGER DEFAULT 45,
			motivo TEXT,
			estado TEXT DEFAULT 'programada',
			canal TEXT,
			prioridad TEXT,
			aseguradora TEXT,
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_odontologia_citas_empresa ON empresa_odontologia_citas(empresa_id, estado, fecha_hora);`,
		`CREATE TABLE IF NOT EXISTS empresa_odontologia_historias (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			paciente_id INTEGER NOT NULL,
			profesional_id INTEGER DEFAULT 0,
			cita_id INTEGER DEFAULT 0,
			fecha_atencion TEXT,
			motivo_consulta TEXT,
			diagnostico TEXT,
			plan_tratamiento TEXT,
			evolucion TEXT,
			formula TEXT,
			recomendaciones TEXT,
			proxima_cita TEXT,
			estado TEXT DEFAULT 'cerrada',
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_odontologia_historias_empresa ON empresa_odontologia_historias(empresa_id, paciente_id, fecha_atencion DESC);`,
		`CREATE TABLE IF NOT EXISTS empresa_odontologia_odontogramas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			paciente_id INTEGER NOT NULL,
			profesional_id INTEGER DEFAULT 0,
			fecha_registro TEXT,
			piezas_json TEXT,
			observaciones TEXT,
			estado TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_odontologia_odontogramas_empresa ON empresa_odontologia_odontogramas(empresa_id, paciente_id, fecha_registro DESC);`,
		`CREATE TABLE IF NOT EXISTS empresa_odontologia_tratamientos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			servicio_id INTEGER,
			paciente_id INTEGER NOT NULL,
			profesional_id INTEGER DEFAULT 0,
			nombre TEXT NOT NULL,
			categoria TEXT,
			piezas TEXT,
			sesiones_total INTEGER DEFAULT 1,
			sesiones_realizadas INTEGER DEFAULT 0,
			costo_estimado REAL DEFAULT 0,
			costo_real REAL DEFAULT 0,
			fecha_inicio TEXT,
			fecha_fin TEXT,
			estado TEXT DEFAULT 'planificado',
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_odontologia_tratamientos_empresa ON empresa_odontologia_tratamientos(empresa_id, estado, id DESC);`,
		`CREATE TABLE IF NOT EXISTS empresa_odontologia_presupuestos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			paciente_id INTEGER NOT NULL,
			tratamiento_id INTEGER DEFAULT 0,
			nombre TEXT NOT NULL,
			valor_total REAL DEFAULT 0,
			cuota_inicial REAL DEFAULT 0,
			saldo REAL DEFAULT 0,
			estado TEXT DEFAULT 'vigente',
			vigencia_hasta TEXT,
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_odontologia_presupuestos_empresa ON empresa_odontologia_presupuestos(empresa_id, estado, id DESC);`,
		`CREATE TABLE IF NOT EXISTS empresa_odontologia_pagos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			paciente_id INTEGER DEFAULT 0,
			cliente_id INTEGER,
			presupuesto_id INTEGER DEFAULT 0,
			servicio_id INTEGER,
			carrito_id INTEGER,
			carrito_item_id INTEGER,
			concepto TEXT NOT NULL,
			monto REAL DEFAULT 0,
			metodo_pago TEXT,
			referencia TEXT,
			fecha_pago TEXT,
			estado TEXT DEFAULT 'aplicado',
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_odontologia_pagos_empresa ON empresa_odontologia_pagos(empresa_id, fecha_pago DESC, id DESC);`,
	}
	for _, stmt := range stmts {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	columnGroups := []struct {
		table   string
		columns []struct {
			name string
			def  string
		}
	}{
		{"empresa_odontologia_pacientes", []struct {
			name string
			def  string
		}{{"cliente_id", "INTEGER"}}},
		{"empresa_odontologia_tratamientos", []struct {
			name string
			def  string
		}{{"servicio_id", "INTEGER"}}},
		{"empresa_odontologia_pagos", []struct {
			name string
			def  string
		}{{"cliente_id", "INTEGER"}, {"servicio_id", "INTEGER"}, {"carrito_id", "INTEGER"}, {"carrito_item_id", "INTEGER"}}},
	}
	for _, group := range columnGroups {
		for _, column := range group.columns {
			if err := ensureColumnIfMissing(dbConn, group.table, column.name, column.def); err != nil {
				return err
			}
		}
	}
	postColumnIndexes := []string{
		`CREATE INDEX IF NOT EXISTS ix_empresa_odontologia_pacientes_cliente ON empresa_odontologia_pacientes(empresa_id, cliente_id);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_odontologia_tratamientos_servicio ON empresa_odontologia_tratamientos(empresa_id, servicio_id);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_odontologia_pagos_carrito ON empresa_odontologia_pagos(empresa_id, carrito_id);`,
	}
	for _, stmt := range postColumnIndexes {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	empresaOdontologiaSchemaEnsured.Store(cacheKey, true)
	return nil
}

func normalizeOdontoEstado(raw, fallback string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		value = fallback
	}
	if value == "" {
		value = "activo"
	}
	return value
}

func defaultOdontoCode(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().Unix())
}

func odontoCoreCode(prefix string, parts ...string) string {
	var b strings.Builder
	for _, part := range parts {
		clean := strings.ToUpper(strings.TrimSpace(part))
		for _, r := range clean {
			if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				b.WriteRune(r)
			} else if b.Len() > 0 {
				last := b.String()[b.Len()-1]
				if last != '-' {
					b.WriteRune('-')
				}
			}
		}
		if b.Len() > 0 {
			last := b.String()[b.Len()-1]
			if last != '-' {
				b.WriteRune('-')
			}
		}
	}
	code := strings.Trim(b.String(), "-")
	if code == "" {
		code = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	if len(code) > 42 {
		code = code[:42]
	}
	return strings.Trim(strings.ToUpper(strings.TrimSpace(prefix)), "-") + "-" + code
}

func findEmpresaOdontologiaClienteID(dbConn *sql.DB, paciente EmpresaOdontologiaPaciente) (int64, error) {
	if paciente.ClienteID > 0 {
		return paciente.ClienteID, nil
	}
	documento := normalizeClienteDocumentoValue(paciente.Documento)
	if documento != "" {
		query := fmt.Sprintf(`SELECT id FROM clientes WHERE empresa_id = ? AND %s = ? LIMIT 1`, clienteDocumentoSQLExpr("numero_documento"))
		return findClienteDuplicateID(dbConn, query, paciente.EmpresaID, documento)
	}
	if email := normalizeClienteEmailValue(paciente.Email); email != "" {
		return findClienteDuplicateID(dbConn, `SELECT id FROM clientes WHERE empresa_id = ? AND lower(trim(COALESCE(email, ''))) = ? LIMIT 1`, paciente.EmpresaID, email)
	}
	if telefono := normalizeClienteTelefonoValue(paciente.Telefono); telefono != "" {
		query := fmt.Sprintf(`SELECT id FROM clientes WHERE empresa_id = ? AND %s = ? LIMIT 1`, clienteTelefonoSQLExpr("telefono"))
		return findClienteDuplicateID(dbConn, query, paciente.EmpresaID, telefono)
	}
	if codigo := strings.TrimSpace(paciente.Codigo); codigo != "" {
		return findClienteDuplicateID(dbConn, `SELECT id FROM clientes WHERE empresa_id = ? AND tipo_documento = 'OTRO' AND numero_documento = ? LIMIT 1`, paciente.EmpresaID, "OD-"+codigo)
	}
	return 0, nil
}

func ensureEmpresaOdontologiaPacienteCliente(dbConn *sql.DB, paciente EmpresaOdontologiaPaciente) (int64, error) {
	if err := EnsureEmpresaClientesSchema(dbConn); err != nil {
		return 0, err
	}
	if id, err := findEmpresaOdontologiaClienteID(dbConn, paciente); err != nil {
		return 0, err
	} else if id > 0 {
		return id, nil
	}
	tipoDocumento := "CC"
	numeroDocumento := strings.TrimSpace(paciente.Documento)
	if numeroDocumento == "" {
		tipoDocumento = "OTRO"
		numeroDocumento = "OD-" + strings.TrimSpace(paciente.Codigo)
		if strings.TrimSpace(paciente.Codigo) == "" {
			numeroDocumento = odontoCoreCode("OD-PAC", paciente.NombreCompleto)
		}
	}
	id, err := CreateCliente(dbConn, Cliente{
		EmpresaID:         paciente.EmpresaID,
		TipoDocumento:     tipoDocumento,
		NumeroDocumento:   numeroDocumento,
		TipoPersona:       "natural",
		NombreRazonSocial: paciente.NombreCompleto,
		NombreComercial:   paciente.NombreCompleto,
		Email:             paciente.Email,
		Telefono:          paciente.Telefono,
		Pais:              "CO",
		UsuarioCreador:    paciente.UsuarioCreador,
		Estado:            "activo",
		Observaciones:     "Cliente creado/sincronizado desde odontologia.",
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

func getEmpresaOdontologiaPacienteByID(dbConn *sql.DB, empresaID, pacienteID int64) (*EmpresaOdontologiaPaciente, error) {
	var paciente EmpresaOdontologiaPaciente
	err := queryRowSQLCompat(dbConn, `SELECT id, empresa_id, COALESCE(cliente_id,0), COALESCE(codigo,''), COALESCE(nombre_completo,''), COALESCE(documento,''), COALESCE(telefono,''), COALESCE(email,''), COALESCE(usuario_creador,'')
		FROM empresa_odontologia_pacientes WHERE empresa_id=? AND id=? LIMIT 1`, empresaID, pacienteID).
		Scan(&paciente.ID, &paciente.EmpresaID, &paciente.ClienteID, &paciente.Codigo, &paciente.NombreCompleto, &paciente.Documento, &paciente.Telefono, &paciente.Email, &paciente.UsuarioCreador)
	if err != nil {
		return nil, err
	}
	return &paciente, nil
}

func syncEmpresaOdontologiaPacienteCliente(dbConn *sql.DB, paciente EmpresaOdontologiaPaciente) (int64, error) {
	clienteID, err := ensureEmpresaOdontologiaPacienteCliente(dbConn, paciente)
	if err != nil || clienteID <= 0 || paciente.ID <= 0 {
		return clienteID, err
	}
	_, err = execSQLCompat(dbConn, `UPDATE empresa_odontologia_pacientes SET cliente_id=?, fecha_actualizacion=datetime('now','localtime') WHERE empresa_id=? AND id=?`, clienteID, paciente.EmpresaID, paciente.ID)
	return clienteID, err
}

func syncEmpresaOdontologiaTratamientoServicio(dbConn *sql.DB, tratamiento EmpresaOdontologiaTratamiento) (int64, error) {
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		return 0, err
	}
	code := odontoCoreCode("OD-TRAT", fmt.Sprintf("%d", tratamiento.ID), tratamiento.Nombre)
	precio := tratamiento.CostoEstimado
	if precio <= 0 {
		precio = tratamiento.CostoReal
	}
	if tratamiento.ServicioID > 0 {
		if err := UpdateServicio(dbConn, Servicio{ID: tratamiento.ServicioID, EmpresaID: tratamiento.EmpresaID, Codigo: code, Nombre: tratamiento.Nombre, Descripcion: tratamiento.Observaciones, Categoria: "odontologia", DuracionMinutos: tratamiento.SesionesTotal * 45, Precio: precio, Estado: "activo", UsuarioCreador: tratamiento.UsuarioCreador, Observaciones: "Servicio vendible sincronizado desde tratamiento odontologico."}); err != nil {
			return 0, err
		}
		return tratamiento.ServicioID, nil
	}
	var servicioID int64
	err := queryRowSQLCompat(dbConn, `SELECT id FROM servicios WHERE empresa_id=? AND codigo=? LIMIT 1`, tratamiento.EmpresaID, code).Scan(&servicioID)
	if err == sql.ErrNoRows {
		servicioID, err = CreateServicio(dbConn, Servicio{EmpresaID: tratamiento.EmpresaID, Codigo: code, Nombre: tratamiento.Nombre, Descripcion: tratamiento.Observaciones, Categoria: "odontologia", DuracionMinutos: tratamiento.SesionesTotal * 45, Precio: precio, Estado: "activo", UsuarioCreador: tratamiento.UsuarioCreador, Observaciones: "Servicio vendible sincronizado desde tratamiento odontologico."})
	}
	if err != nil {
		return 0, err
	}
	if tratamiento.ID > 0 {
		_, err = execSQLCompat(dbConn, `UPDATE empresa_odontologia_tratamientos SET servicio_id=?, fecha_actualizacion=datetime('now','localtime') WHERE empresa_id=? AND id=?`, servicioID, tratamiento.EmpresaID, tratamiento.ID)
	}
	return servicioID, err
}

func getEmpresaOdontologiaTratamientoByID(dbConn *sql.DB, empresaID, tratamientoID int64) (*EmpresaOdontologiaTratamiento, error) {
	var t EmpresaOdontologiaTratamiento
	err := queryRowSQLCompat(dbConn, `SELECT id, empresa_id, COALESCE(servicio_id,0), COALESCE(paciente_id,0), COALESCE(profesional_id,0), COALESCE(nombre,''), COALESCE(categoria,''), COALESCE(sesiones_total,1), COALESCE(costo_estimado,0), COALESCE(costo_real,0), COALESCE(estado,'planificado'), COALESCE(observaciones,''), COALESCE(usuario_creador,'')
		FROM empresa_odontologia_tratamientos WHERE empresa_id=? AND id=? LIMIT 1`, empresaID, tratamientoID).
		Scan(&t.ID, &t.EmpresaID, &t.ServicioID, &t.PacienteID, &t.ProfesionalID, &t.Nombre, &t.Categoria, &t.SesionesTotal, &t.CostoEstimado, &t.CostoReal, &t.Estado, &t.Observaciones, &t.UsuarioCreador)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func ensureEmpresaOdontologiaPagoServicio(dbConn *sql.DB, pago EmpresaOdontologiaPago) (int64, error) {
	if pago.ServicioID > 0 {
		return pago.ServicioID, nil
	}
	if pago.PresupuestoID > 0 {
		var tratamientoID int64
		err := queryRowSQLCompat(dbConn, `SELECT COALESCE(tratamiento_id,0) FROM empresa_odontologia_presupuestos WHERE empresa_id=? AND id=? LIMIT 1`, pago.EmpresaID, pago.PresupuestoID).Scan(&tratamientoID)
		if err != nil {
			return 0, err
		}
		if tratamientoID > 0 {
			t, err := getEmpresaOdontologiaTratamientoByID(dbConn, pago.EmpresaID, tratamientoID)
			if err != nil {
				return 0, err
			}
			if t.UsuarioCreador == "" {
				t.UsuarioCreador = pago.UsuarioCreador
			}
			return syncEmpresaOdontologiaTratamientoServicio(dbConn, *t)
		}
	}
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		return 0, err
	}
	code := odontoCoreCode("OD-SERV", pago.Concepto)
	var servicioID int64
	err := queryRowSQLCompat(dbConn, `SELECT id FROM servicios WHERE empresa_id=? AND codigo=? LIMIT 1`, pago.EmpresaID, code).Scan(&servicioID)
	if err == sql.ErrNoRows {
		servicioID, err = CreateServicio(dbConn, Servicio{EmpresaID: pago.EmpresaID, Codigo: code, Nombre: pago.Concepto, Descripcion: "Servicio odontologico creado desde recaudo.", Categoria: "odontologia", Precio: pago.Monto, Estado: "activo", UsuarioCreador: pago.UsuarioCreador, Observaciones: "Servicio vendible sincronizado desde pagos de odontologia."})
	}
	return servicioID, err
}

func odontoPagoCarritoReferencia(pago EmpresaOdontologiaPago) string {
	if pago.ID > 0 {
		return fmt.Sprintf("odontologia:pago:%d", pago.ID)
	}
	referencia := strings.TrimSpace(pago.Referencia)
	if referencia != "" {
		return "odontologia:pago:" + odontoCoreCode("REF", referencia)
	}
	return fmt.Sprintf("odontologia:paciente:%d:%s:%.2f", pago.PacienteID, strings.TrimSpace(pago.FechaPago), pago.Monto)
}

func odontoPagoCarritoNombre(pago EmpresaOdontologiaPago) string {
	concepto := strings.TrimSpace(pago.Concepto)
	if concepto == "" {
		concepto = "Pago odontologico"
	}
	if pago.ID > 0 {
		return fmt.Sprintf("Odontologia - %s #%d", concepto, pago.ID)
	}
	if ref := strings.TrimSpace(pago.Referencia); ref != "" {
		return "Odontologia - " + concepto + " - " + odontoCoreCode("REF", ref)
	}
	return "Odontologia - " + concepto + " - " + odontoCoreCode("PAC", fmt.Sprintf("%d", pago.PacienteID), pago.FechaPago)
}

func createEmpresaOdontologiaPagoCarrito(dbConn *sql.DB, pago EmpresaOdontologiaPago) (int64, int64, error) {
	if pago.Estado == "anulado" {
		return 0, 0, nil
	}
	if err := EnsureEmpresaCarritosSchema(dbConn); err != nil {
		return 0, 0, err
	}
	metodo := NormalizeMetodoPagoCarrito(pago.MetodoPago)
	if metodo == "" {
		metodo = "efectivo"
	}
	referenciaExterna := odontoPagoCarritoReferencia(pago)
	var carritoExistente, itemExistente int64
	err := queryRowSQLCompat(dbConn, `SELECT id FROM carritos_compras WHERE empresa_id=? AND referencia_externa=? LIMIT 1`, pago.EmpresaID, referenciaExterna).Scan(&carritoExistente)
	if err == nil && carritoExistente > 0 {
		_ = queryRowSQLCompat(dbConn, `SELECT id FROM carrito_compra_items WHERE empresa_id=? AND carrito_id=? AND referencia_id=? AND tipo_item='servicio' LIMIT 1`, pago.EmpresaID, carritoExistente, pago.ServicioID).Scan(&itemExistente)
		return carritoExistente, itemExistente, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}
	carritoID, err := CreateCarritoCompra(dbConn, CarritoCompra{
		EmpresaID:         pago.EmpresaID,
		Codigo:            odontoCoreCode("OD-PAGO", fmt.Sprintf("%d", pago.PacienteID), fmt.Sprintf("%d", time.Now().UnixNano())),
		Nombre:            odontoPagoCarritoNombre(pago),
		CanalVenta:        "odontologia",
		ClienteID:         pago.ClienteID,
		EstadoCarrito:     "abierto",
		Moneda:            "COP",
		ReferenciaExterna: referenciaExterna,
		MetodoPago:        metodo,
		ReferenciaPago:    pago.Referencia,
		UsuarioCreador:    pago.UsuarioCreador,
		Observaciones:     "Venta central generada desde pago odontologico.",
	})
	if err != nil {
		return 0, 0, err
	}
	itemID, err := CreateCarritoCompraItem(dbConn, CarritoCompraItem{
		EmpresaID:          pago.EmpresaID,
		CarritoID:          carritoID,
		TipoItem:           "servicio",
		ReferenciaID:       pago.ServicioID,
		CodigoItem:         odontoCoreCode("OD-ITEM", pago.Concepto),
		Descripcion:        pago.Concepto,
		UnidadMedida:       "servicio",
		Cantidad:           1,
		PrecioUnitario:     pago.Monto,
		ImpuestoPorcentaje: 0,
		UsuarioCreador:     pago.UsuarioCreador,
		Estado:             "activo",
		Observaciones:      "Item central generado desde pago odontologico.",
	})
	if err != nil {
		return 0, 0, err
	}
	if err := PayCarritoStationSession(dbConn, pago.EmpresaID, carritoID, metodo, pago.Referencia, "", "", 0, 0, pago.Monto, 0, pago.UsuarioCreador); err != nil {
		return 0, 0, err
	}
	return carritoID, itemID, nil
}

func BuildEmpresaOdontologiaDashboard(dbConn *sql.DB, empresaID int64) (*EmpresaOdontologiaDashboard, error) {
	if err := EnsureEmpresaOdontologiaSchema(dbConn); err != nil {
		return nil, err
	}
	row := &EmpresaOdontologiaDashboard{EmpresaID: empresaID}
	_ = queryRowSQLCompat(dbConn, `SELECT COUNT(1) FROM empresa_odontologia_pacientes WHERE empresa_id = ? AND COALESCE(estado,'activo') = 'activo'`, empresaID).Scan(&row.PacientesActivos)
	_ = queryRowSQLCompat(dbConn, `SELECT COUNT(1) FROM empresa_odontologia_profesionales WHERE empresa_id = ? AND COALESCE(estado,'activo') = 'activo'`, empresaID).Scan(&row.ProfesionalesActivos)
	_ = queryRowSQLCompat(dbConn, `SELECT COUNT(1) FROM empresa_odontologia_citas WHERE empresa_id = ? AND substr(COALESCE(fecha_hora,''),1,10) = substr(datetime('now','localtime'),1,10)`, empresaID).Scan(&row.CitasHoy)
	_ = queryRowSQLCompat(dbConn, `SELECT COUNT(1) FROM empresa_odontologia_citas WHERE empresa_id = ? AND COALESCE(estado,'programada') IN ('programada','confirmada','en_sala')`, empresaID).Scan(&row.CitasPendientes)
	_ = queryRowSQLCompat(dbConn, `SELECT COUNT(1) FROM empresa_odontologia_tratamientos WHERE empresa_id = ? AND COALESCE(estado,'planificado') IN ('planificado','en_proceso')`, empresaID).Scan(&row.TratamientosActivos)
	_ = queryRowSQLCompat(dbConn, `SELECT COUNT(1) FROM empresa_odontologia_presupuestos WHERE empresa_id = ? AND COALESCE(estado,'vigente') IN ('vigente','parcial')`, empresaID).Scan(&row.PresupuestosVigentes)
	_ = queryRowSQLCompat(dbConn, `SELECT COALESCE(SUM(monto),0) FROM empresa_odontologia_pagos WHERE empresa_id = ? AND substr(COALESCE(fecha_pago,''),1,7) = substr(datetime('now','localtime'),1,7) AND COALESCE(estado,'aplicado') <> 'anulado'`, empresaID).Scan(&row.RecaudoMes)
	_ = queryRowSQLCompat(dbConn, `SELECT COALESCE(SUM(saldo),0) FROM empresa_odontologia_presupuestos WHERE empresa_id = ? AND COALESCE(estado,'vigente') IN ('vigente','parcial')`, empresaID).Scan(&row.SaldoPendiente)
	agenda, err := ListEmpresaOdontologiaCitasByFecha(dbConn, empresaID, time.Now().Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	row.AgendaHoy = agenda
	trats, err := ListEmpresaOdontologiaTratamientos(dbConn, empresaID)
	if err != nil {
		return nil, err
	}
	if len(trats) > 6 {
		trats = trats[:6]
	}
	row.TratamientosPrioridad = trats
	return row, nil
}

func ListEmpresaOdontologiaPacientes(dbConn *sql.DB, empresaID int64) ([]EmpresaOdontologiaPaciente, error) {
	if err := EnsureEmpresaOdontologiaSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT id, empresa_id, COALESCE(cliente_id,0), COALESCE(codigo,''), COALESCE(nombre_completo,''), COALESCE(documento,''), COALESCE(telefono,''), COALESCE(email,''), COALESCE(fecha_nacimiento,''), COALESCE(genero,''), COALESCE(aseguradora,''), COALESCE(alergias,''), COALESCE(riesgo_medico,''), COALESCE(ultima_visita,''), COALESCE(saldo,0), COALESCE(estado,'activo'), COALESCE(observaciones,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'') FROM empresa_odontologia_pacientes WHERE empresa_id = ? ORDER BY id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaOdontologiaPaciente{}
	for rows.Next() {
		var item EmpresaOdontologiaPaciente
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.ClienteID, &item.Codigo, &item.NombreCompleto, &item.Documento, &item.Telefono, &item.Email, &item.FechaNacimiento, &item.Genero, &item.Aseguradora, &item.Alergias, &item.RiesgoMedico, &item.UltimaVisita, &item.Saldo, &item.Estado, &item.Observaciones, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func CreateEmpresaOdontologiaPaciente(dbConn *sql.DB, payload EmpresaOdontologiaPaciente) (int64, error) {
	if err := EnsureEmpresaOdontologiaSchema(dbConn); err != nil {
		return 0, err
	}
	if payload.EmpresaID <= 0 || strings.TrimSpace(payload.NombreCompleto) == "" {
		return 0, fmt.Errorf("empresa_id y nombre_completo son obligatorios")
	}
	if strings.TrimSpace(payload.Codigo) == "" {
		payload.Codigo = defaultOdontoCode("PAC")
	}
	payload.Estado = normalizeOdontoEstado(payload.Estado, "activo")
	clienteID, err := ensureEmpresaOdontologiaPacienteCliente(dbConn, payload)
	if err != nil {
		return 0, err
	}
	payload.ClienteID = clienteID
	return insertSQLCompat(dbConn, `INSERT INTO empresa_odontologia_pacientes (empresa_id,cliente_id,codigo,nombre_completo,documento,telefono,email,fecha_nacimiento,genero,aseguradora,alergias,riesgo_medico,ultima_visita,saldo,estado,observaciones,fecha_creacion,fecha_actualizacion,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,datetime('now','localtime'),datetime('now','localtime'),?)`, payload.EmpresaID, nullableInt64(payload.ClienteID), strings.TrimSpace(payload.Codigo), strings.TrimSpace(payload.NombreCompleto), strings.TrimSpace(payload.Documento), strings.TrimSpace(payload.Telefono), strings.TrimSpace(payload.Email), strings.TrimSpace(payload.FechaNacimiento), strings.TrimSpace(payload.Genero), strings.TrimSpace(payload.Aseguradora), strings.TrimSpace(payload.Alergias), strings.TrimSpace(payload.RiesgoMedico), strings.TrimSpace(payload.UltimaVisita), payload.Saldo, payload.Estado, strings.TrimSpace(payload.Observaciones), strings.TrimSpace(payload.UsuarioCreador))
}

func SetEmpresaOdontologiaPacienteEstado(dbConn *sql.DB, empresaID, id int64, estado string) error {
	_, err := execSQLCompat(dbConn, `UPDATE empresa_odontologia_pacientes SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ? AND id = ?`, normalizeOdontoEstado(estado, "activo"), empresaID, id)
	return err
}

func ListEmpresaOdontologiaProfesionales(dbConn *sql.DB, empresaID int64) ([]EmpresaOdontologiaProfesional, error) {
	if err := EnsureEmpresaOdontologiaSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT id, empresa_id, COALESCE(nombre_completo,''), COALESCE(especialidad,''), COALESCE(registro_profesional,''), COALESCE(telefono,''), COALESCE(email,''), COALESCE(color_agenda,''), COALESCE(estado,'activo'), COALESCE(observaciones,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'') FROM empresa_odontologia_profesionales WHERE empresa_id = ? ORDER BY id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaOdontologiaProfesional{}
	for rows.Next() {
		var item EmpresaOdontologiaProfesional
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.NombreCompleto, &item.Especialidad, &item.RegistroProfesional, &item.Telefono, &item.Email, &item.ColorAgenda, &item.Estado, &item.Observaciones, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func CreateEmpresaOdontologiaProfesional(dbConn *sql.DB, payload EmpresaOdontologiaProfesional) (int64, error) {
	if err := EnsureEmpresaOdontologiaSchema(dbConn); err != nil {
		return 0, err
	}
	if payload.EmpresaID <= 0 || strings.TrimSpace(payload.NombreCompleto) == "" {
		return 0, fmt.Errorf("empresa_id y nombre_completo son obligatorios")
	}
	payload.Estado = normalizeOdontoEstado(payload.Estado, "activo")
	if strings.TrimSpace(payload.ColorAgenda) == "" {
		payload.ColorAgenda = "#0ea5e9"
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_odontologia_profesionales (empresa_id,nombre_completo,especialidad,registro_profesional,telefono,email,color_agenda,estado,observaciones,fecha_creacion,fecha_actualizacion,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,datetime('now','localtime'),datetime('now','localtime'),?)`, payload.EmpresaID, strings.TrimSpace(payload.NombreCompleto), strings.TrimSpace(payload.Especialidad), strings.TrimSpace(payload.RegistroProfesional), strings.TrimSpace(payload.Telefono), strings.TrimSpace(payload.Email), strings.TrimSpace(payload.ColorAgenda), payload.Estado, strings.TrimSpace(payload.Observaciones), strings.TrimSpace(payload.UsuarioCreador))
}

func SetEmpresaOdontologiaProfesionalEstado(dbConn *sql.DB, empresaID, id int64, estado string) error {
	_, err := execSQLCompat(dbConn, `UPDATE empresa_odontologia_profesionales SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ? AND id = ?`, normalizeOdontoEstado(estado, "activo"), empresaID, id)
	return err
}

func ListEmpresaOdontologiaConsultorios(dbConn *sql.DB, empresaID int64) ([]EmpresaOdontologiaConsultorio, error) {
	if err := EnsureEmpresaOdontologiaSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT id, empresa_id, COALESCE(nombre,''), COALESCE(sede,''), COALESCE(sillon,''), COALESCE(estado,'activo'), COALESCE(observaciones,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'') FROM empresa_odontologia_consultorios WHERE empresa_id = ? ORDER BY id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaOdontologiaConsultorio{}
	for rows.Next() {
		var item EmpresaOdontologiaConsultorio
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.Nombre, &item.Sede, &item.Sillon, &item.Estado, &item.Observaciones, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func CreateEmpresaOdontologiaConsultorio(dbConn *sql.DB, payload EmpresaOdontologiaConsultorio) (int64, error) {
	if err := EnsureEmpresaOdontologiaSchema(dbConn); err != nil {
		return 0, err
	}
	if payload.EmpresaID <= 0 || strings.TrimSpace(payload.Nombre) == "" {
		return 0, fmt.Errorf("empresa_id y nombre son obligatorios")
	}
	payload.Estado = normalizeOdontoEstado(payload.Estado, "activo")
	return insertSQLCompat(dbConn, `INSERT INTO empresa_odontologia_consultorios (empresa_id,nombre,sede,sillon,estado,observaciones,fecha_creacion,fecha_actualizacion,usuario_creador) VALUES (?,?,?,?,?,?,datetime('now','localtime'),datetime('now','localtime'),?)`, payload.EmpresaID, strings.TrimSpace(payload.Nombre), strings.TrimSpace(payload.Sede), strings.TrimSpace(payload.Sillon), payload.Estado, strings.TrimSpace(payload.Observaciones), strings.TrimSpace(payload.UsuarioCreador))
}

func SetEmpresaOdontologiaConsultorioEstado(dbConn *sql.DB, empresaID, id int64, estado string) error {
	_, err := execSQLCompat(dbConn, `UPDATE empresa_odontologia_consultorios SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ? AND id = ?`, normalizeOdontoEstado(estado, "activo"), empresaID, id)
	return err
}

func ListEmpresaOdontologiaCitasByFecha(dbConn *sql.DB, empresaID int64, fecha string) ([]EmpresaOdontologiaCita, error) {
	if err := EnsureEmpresaOdontologiaSchema(dbConn); err != nil {
		return nil, err
	}
	query := `SELECT c.id,c.empresa_id,c.paciente_id,COALESCE(p.nombre_completo,''),c.profesional_id,COALESCE(pr.nombre_completo,''),c.consultorio_id,COALESCE(co.nombre,''),COALESCE(c.fecha_hora,''),COALESCE(c.duracion_minutos,45),COALESCE(c.motivo,''),COALESCE(c.estado,'programada'),COALESCE(c.canal,''),COALESCE(c.prioridad,''),COALESCE(c.aseguradora,''),COALESCE(c.observaciones,''),COALESCE(c.fecha_creacion,''),COALESCE(c.fecha_actualizacion,''),COALESCE(c.usuario_creador,'') FROM empresa_odontologia_citas c LEFT JOIN empresa_odontologia_pacientes p ON p.id = c.paciente_id AND p.empresa_id = c.empresa_id LEFT JOIN empresa_odontologia_profesionales pr ON pr.id = c.profesional_id AND pr.empresa_id = c.empresa_id LEFT JOIN empresa_odontologia_consultorios co ON co.id = c.consultorio_id AND co.empresa_id = c.empresa_id WHERE c.empresa_id = ?`
	args := []interface{}{empresaID}
	if strings.TrimSpace(fecha) != "" {
		query += ` AND substr(COALESCE(c.fecha_hora,''),1,10) = ?`
		args = append(args, strings.TrimSpace(fecha))
	}
	query += ` ORDER BY c.fecha_hora ASC, c.id ASC`
	rows, err := querySQLCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaOdontologiaCita{}
	for rows.Next() {
		var item EmpresaOdontologiaCita
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.PacienteID, &item.PacienteNombre, &item.ProfesionalID, &item.ProfesionalNombre, &item.ConsultorioID, &item.ConsultorioNombre, &item.FechaHora, &item.DuracionMinutos, &item.Motivo, &item.Estado, &item.Canal, &item.Prioridad, &item.Aseguradora, &item.Observaciones, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func CreateEmpresaOdontologiaCita(dbConn *sql.DB, payload EmpresaOdontologiaCita) (int64, error) {
	if err := EnsureEmpresaOdontologiaSchema(dbConn); err != nil {
		return 0, err
	}
	if payload.EmpresaID <= 0 || payload.PacienteID <= 0 || payload.ProfesionalID <= 0 || strings.TrimSpace(payload.FechaHora) == "" {
		return 0, fmt.Errorf("empresa_id, paciente_id, profesional_id y fecha_hora son obligatorios")
	}
	if payload.DuracionMinutos <= 0 {
		payload.DuracionMinutos = 45
	}
	payload.Estado = normalizeOdontoEstado(payload.Estado, "programada")
	return insertSQLCompat(dbConn, `INSERT INTO empresa_odontologia_citas (empresa_id,paciente_id,profesional_id,consultorio_id,fecha_hora,duracion_minutos,motivo,estado,canal,prioridad,aseguradora,observaciones,fecha_creacion,fecha_actualizacion,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,datetime('now','localtime'),datetime('now','localtime'),?)`, payload.EmpresaID, payload.PacienteID, payload.ProfesionalID, payload.ConsultorioID, strings.TrimSpace(payload.FechaHora), payload.DuracionMinutos, strings.TrimSpace(payload.Motivo), payload.Estado, strings.TrimSpace(payload.Canal), strings.TrimSpace(payload.Prioridad), strings.TrimSpace(payload.Aseguradora), strings.TrimSpace(payload.Observaciones), strings.TrimSpace(payload.UsuarioCreador))
}

func SetEmpresaOdontologiaCitaEstado(dbConn *sql.DB, empresaID, id int64, estado string) error {
	_, err := execSQLCompat(dbConn, `UPDATE empresa_odontologia_citas SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ? AND id = ?`, normalizeOdontoEstado(estado, "programada"), empresaID, id)
	return err
}

func ListEmpresaOdontologiaHistorias(dbConn *sql.DB, empresaID int64) ([]EmpresaOdontologiaHistoria, error) {
	if err := EnsureEmpresaOdontologiaSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT h.id,h.empresa_id,h.paciente_id,COALESCE(p.nombre_completo,''),h.profesional_id,COALESCE(pr.nombre_completo,''),h.cita_id,COALESCE(h.fecha_atencion,''),COALESCE(h.motivo_consulta,''),COALESCE(h.diagnostico,''),COALESCE(h.plan_tratamiento,''),COALESCE(h.evolucion,''),COALESCE(h.formula,''),COALESCE(h.recomendaciones,''),COALESCE(h.proxima_cita,''),COALESCE(h.estado,'cerrada'),COALESCE(h.fecha_creacion,''),COALESCE(h.fecha_actualizacion,''),COALESCE(h.usuario_creador,'') FROM empresa_odontologia_historias h LEFT JOIN empresa_odontologia_pacientes p ON p.id = h.paciente_id AND p.empresa_id = h.empresa_id LEFT JOIN empresa_odontologia_profesionales pr ON pr.id = h.profesional_id AND pr.empresa_id = h.empresa_id WHERE h.empresa_id = ? ORDER BY h.fecha_atencion DESC, h.id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaOdontologiaHistoria{}
	for rows.Next() {
		var item EmpresaOdontologiaHistoria
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.PacienteID, &item.PacienteNombre, &item.ProfesionalID, &item.ProfesionalNombre, &item.CitaID, &item.FechaAtencion, &item.MotivoConsulta, &item.Diagnostico, &item.PlanTratamiento, &item.Evolucion, &item.Formula, &item.Recomendaciones, &item.ProximaCita, &item.Estado, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func CreateEmpresaOdontologiaHistoria(dbConn *sql.DB, payload EmpresaOdontologiaHistoria) (int64, error) {
	if err := EnsureEmpresaOdontologiaSchema(dbConn); err != nil {
		return 0, err
	}
	if payload.EmpresaID <= 0 || payload.PacienteID <= 0 {
		return 0, fmt.Errorf("empresa_id y paciente_id son obligatorios")
	}
	payload.Estado = normalizeOdontoEstado(payload.Estado, "cerrada")
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_odontologia_historias (empresa_id,paciente_id,profesional_id,cita_id,fecha_atencion,motivo_consulta,diagnostico,plan_tratamiento,evolucion,formula,recomendaciones,proxima_cita,estado,fecha_creacion,fecha_actualizacion,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,datetime('now','localtime'),datetime('now','localtime'),?)`, payload.EmpresaID, payload.PacienteID, payload.ProfesionalID, payload.CitaID, strings.TrimSpace(payload.FechaAtencion), strings.TrimSpace(payload.MotivoConsulta), strings.TrimSpace(payload.Diagnostico), strings.TrimSpace(payload.PlanTratamiento), strings.TrimSpace(payload.Evolucion), strings.TrimSpace(payload.Formula), strings.TrimSpace(payload.Recomendaciones), strings.TrimSpace(payload.ProximaCita), payload.Estado, strings.TrimSpace(payload.UsuarioCreador))
	if err == nil {
		_, _ = execSQLCompat(dbConn, `UPDATE empresa_odontologia_pacientes SET ultima_visita = ?, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ? AND id = ?`, strings.TrimSpace(payload.FechaAtencion), payload.EmpresaID, payload.PacienteID)
	}
	return id, err
}

func ListEmpresaOdontologiaOdontogramas(dbConn *sql.DB, empresaID int64) ([]EmpresaOdontologiaOdontograma, error) {
	if err := EnsureEmpresaOdontologiaSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT o.id,o.empresa_id,o.paciente_id,COALESCE(p.nombre_completo,''),o.profesional_id,COALESCE(pr.nombre_completo,''),COALESCE(o.fecha_registro,''),COALESCE(o.piezas_json,''),COALESCE(o.observaciones,''),COALESCE(o.estado,'activo'),COALESCE(o.fecha_creacion,''),COALESCE(o.fecha_actualizacion,''),COALESCE(o.usuario_creador,'') FROM empresa_odontologia_odontogramas o LEFT JOIN empresa_odontologia_pacientes p ON p.id = o.paciente_id AND p.empresa_id = o.empresa_id LEFT JOIN empresa_odontologia_profesionales pr ON pr.id = o.profesional_id AND pr.empresa_id = o.empresa_id WHERE o.empresa_id = ? ORDER BY o.fecha_registro DESC, o.id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaOdontologiaOdontograma{}
	for rows.Next() {
		var item EmpresaOdontologiaOdontograma
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.PacienteID, &item.PacienteNombre, &item.ProfesionalID, &item.ProfesionalNombre, &item.FechaRegistro, &item.PiezasJSON, &item.Observaciones, &item.Estado, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func CreateEmpresaOdontologiaOdontograma(dbConn *sql.DB, payload EmpresaOdontologiaOdontograma) (int64, error) {
	if err := EnsureEmpresaOdontologiaSchema(dbConn); err != nil {
		return 0, err
	}
	if payload.EmpresaID <= 0 || payload.PacienteID <= 0 {
		return 0, fmt.Errorf("empresa_id y paciente_id son obligatorios")
	}
	payload.Estado = normalizeOdontoEstado(payload.Estado, "activo")
	return insertSQLCompat(dbConn, `INSERT INTO empresa_odontologia_odontogramas (empresa_id,paciente_id,profesional_id,fecha_registro,piezas_json,observaciones,estado,fecha_creacion,fecha_actualizacion,usuario_creador) VALUES (?,?,?,?,?,?,?,?,datetime('now','localtime'),?)`, payload.EmpresaID, payload.PacienteID, payload.ProfesionalID, strings.TrimSpace(payload.FechaRegistro), strings.TrimSpace(payload.PiezasJSON), strings.TrimSpace(payload.Observaciones), payload.Estado, strings.TrimSpace(payload.UsuarioCreador))
}

func ListEmpresaOdontologiaTratamientos(dbConn *sql.DB, empresaID int64) ([]EmpresaOdontologiaTratamiento, error) {
	if err := EnsureEmpresaOdontologiaSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT t.id,t.empresa_id,COALESCE(t.servicio_id,0),t.paciente_id,COALESCE(p.nombre_completo,''),t.profesional_id,COALESCE(pr.nombre_completo,''),COALESCE(t.nombre,''),COALESCE(t.categoria,''),COALESCE(t.piezas,''),COALESCE(t.sesiones_total,1),COALESCE(t.sesiones_realizadas,0),COALESCE(t.costo_estimado,0),COALESCE(t.costo_real,0),COALESCE(t.fecha_inicio,''),COALESCE(t.fecha_fin,''),COALESCE(t.estado,'planificado'),COALESCE(t.observaciones,''),COALESCE(t.fecha_creacion,''),COALESCE(t.fecha_actualizacion,''),COALESCE(t.usuario_creador,'') FROM empresa_odontologia_tratamientos t LEFT JOIN empresa_odontologia_pacientes p ON p.id = t.paciente_id AND p.empresa_id = t.empresa_id LEFT JOIN empresa_odontologia_profesionales pr ON pr.id = t.profesional_id AND pr.empresa_id = t.empresa_id WHERE t.empresa_id = ? ORDER BY t.id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaOdontologiaTratamiento{}
	for rows.Next() {
		var item EmpresaOdontologiaTratamiento
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.ServicioID, &item.PacienteID, &item.PacienteNombre, &item.ProfesionalID, &item.ProfesionalNombre, &item.Nombre, &item.Categoria, &item.Piezas, &item.SesionesTotal, &item.SesionesRealizadas, &item.CostoEstimado, &item.CostoReal, &item.FechaInicio, &item.FechaFin, &item.Estado, &item.Observaciones, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func CreateEmpresaOdontologiaTratamiento(dbConn *sql.DB, payload EmpresaOdontologiaTratamiento) (int64, error) {
	if err := EnsureEmpresaOdontologiaSchema(dbConn); err != nil {
		return 0, err
	}
	if payload.EmpresaID <= 0 || payload.PacienteID <= 0 || strings.TrimSpace(payload.Nombre) == "" {
		return 0, fmt.Errorf("empresa_id, paciente_id y nombre son obligatorios")
	}
	if payload.SesionesTotal <= 0 {
		payload.SesionesTotal = 1
	}
	payload.Estado = normalizeOdontoEstado(payload.Estado, "planificado")
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_odontologia_tratamientos (empresa_id,paciente_id,profesional_id,nombre,categoria,piezas,sesiones_total,sesiones_realizadas,costo_estimado,costo_real,fecha_inicio,fecha_fin,estado,observaciones,fecha_creacion,fecha_actualizacion,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,datetime('now','localtime'),datetime('now','localtime'),?)`, payload.EmpresaID, payload.PacienteID, payload.ProfesionalID, strings.TrimSpace(payload.Nombre), strings.TrimSpace(payload.Categoria), strings.TrimSpace(payload.Piezas), payload.SesionesTotal, payload.SesionesRealizadas, payload.CostoEstimado, payload.CostoReal, strings.TrimSpace(payload.FechaInicio), strings.TrimSpace(payload.FechaFin), payload.Estado, strings.TrimSpace(payload.Observaciones), strings.TrimSpace(payload.UsuarioCreador))
	if err != nil {
		return 0, err
	}
	payload.ID = id
	if _, err := syncEmpresaOdontologiaTratamientoServicio(dbConn, payload); err != nil {
		return 0, err
	}
	return id, nil
}

func SetEmpresaOdontologiaTratamientoEstado(dbConn *sql.DB, empresaID, id int64, estado string) error {
	_, err := execSQLCompat(dbConn, `UPDATE empresa_odontologia_tratamientos SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ? AND id = ?`, normalizeOdontoEstado(estado, "planificado"), empresaID, id)
	return err
}

func ListEmpresaOdontologiaPresupuestos(dbConn *sql.DB, empresaID int64) ([]EmpresaOdontologiaPresupuesto, error) {
	if err := EnsureEmpresaOdontologiaSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT pr.id,pr.empresa_id,pr.paciente_id,COALESCE(pa.nombre_completo,''),pr.tratamiento_id,COALESCE(t.nombre,''),COALESCE(pr.nombre,''),COALESCE(pr.valor_total,0),COALESCE(pr.cuota_inicial,0),COALESCE(pr.saldo,0),COALESCE(pr.estado,'vigente'),COALESCE(pr.vigencia_hasta,''),COALESCE(pr.observaciones,''),COALESCE(pr.fecha_creacion,''),COALESCE(pr.fecha_actualizacion,''),COALESCE(pr.usuario_creador,'') FROM empresa_odontologia_presupuestos pr LEFT JOIN empresa_odontologia_pacientes pa ON pa.id = pr.paciente_id AND pa.empresa_id = pr.empresa_id LEFT JOIN empresa_odontologia_tratamientos t ON t.id = pr.tratamiento_id AND t.empresa_id = pr.empresa_id WHERE pr.empresa_id = ? ORDER BY pr.id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaOdontologiaPresupuesto{}
	for rows.Next() {
		var item EmpresaOdontologiaPresupuesto
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.PacienteID, &item.PacienteNombre, &item.TratamientoID, &item.TratamientoNombre, &item.Nombre, &item.ValorTotal, &item.CuotaInicial, &item.Saldo, &item.Estado, &item.VigenciaHasta, &item.Observaciones, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func CreateEmpresaOdontologiaPresupuesto(dbConn *sql.DB, payload EmpresaOdontologiaPresupuesto) (int64, error) {
	if err := EnsureEmpresaOdontologiaSchema(dbConn); err != nil {
		return 0, err
	}
	if payload.EmpresaID <= 0 || payload.PacienteID <= 0 || strings.TrimSpace(payload.Nombre) == "" {
		return 0, fmt.Errorf("empresa_id, paciente_id y nombre son obligatorios")
	}
	payload.Estado = normalizeOdontoEstado(payload.Estado, "vigente")
	saldo := payload.ValorTotal - payload.CuotaInicial
	if saldo < 0 {
		saldo = 0
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_odontologia_presupuestos (empresa_id,paciente_id,tratamiento_id,nombre,valor_total,cuota_inicial,saldo,estado,vigencia_hasta,observaciones,fecha_creacion,fecha_actualizacion,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,datetime('now','localtime'),datetime('now','localtime'),?)`, payload.EmpresaID, payload.PacienteID, payload.TratamientoID, strings.TrimSpace(payload.Nombre), payload.ValorTotal, payload.CuotaInicial, saldo, payload.Estado, strings.TrimSpace(payload.VigenciaHasta), strings.TrimSpace(payload.Observaciones), strings.TrimSpace(payload.UsuarioCreador))
}

func SetEmpresaOdontologiaPresupuestoEstado(dbConn *sql.DB, empresaID, id int64, estado string) error {
	_, err := execSQLCompat(dbConn, `UPDATE empresa_odontologia_presupuestos SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ? AND id = ?`, normalizeOdontoEstado(estado, "vigente"), empresaID, id)
	return err
}

func ListEmpresaOdontologiaPagos(dbConn *sql.DB, empresaID int64) ([]EmpresaOdontologiaPago, error) {
	if err := EnsureEmpresaOdontologiaSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT pg.id,pg.empresa_id,COALESCE(pg.paciente_id,0),COALESCE(pa.nombre_completo,''),COALESCE(pg.cliente_id,0),COALESCE(pg.presupuesto_id,0),COALESCE(pr.nombre,''),COALESCE(pg.servicio_id,0),COALESCE(pg.carrito_id,0),COALESCE(pg.carrito_item_id,0),COALESCE(pg.concepto,''),COALESCE(pg.monto,0),COALESCE(pg.metodo_pago,''),COALESCE(pg.referencia,''),COALESCE(pg.fecha_pago,''),COALESCE(pg.estado,'aplicado'),COALESCE(pg.observaciones,''),COALESCE(pg.fecha_creacion,''),COALESCE(pg.fecha_actualizacion,''),COALESCE(pg.usuario_creador,'') FROM empresa_odontologia_pagos pg LEFT JOIN empresa_odontologia_pacientes pa ON pa.id = pg.paciente_id AND pa.empresa_id = pg.empresa_id LEFT JOIN empresa_odontologia_presupuestos pr ON pr.id = pg.presupuesto_id AND pr.empresa_id = pg.empresa_id WHERE pg.empresa_id = ? ORDER BY pg.fecha_pago DESC, pg.id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaOdontologiaPago{}
	for rows.Next() {
		var item EmpresaOdontologiaPago
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.PacienteID, &item.PacienteNombre, &item.ClienteID, &item.PresupuestoID, &item.PresupuestoNombre, &item.ServicioID, &item.CarritoID, &item.CarritoItemID, &item.Concepto, &item.Monto, &item.MetodoPago, &item.Referencia, &item.FechaPago, &item.Estado, &item.Observaciones, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func CreateEmpresaOdontologiaPago(dbConn *sql.DB, payload EmpresaOdontologiaPago) (int64, error) {
	if err := EnsureEmpresaOdontologiaSchema(dbConn); err != nil {
		return 0, err
	}
	if payload.EmpresaID <= 0 || strings.TrimSpace(payload.Concepto) == "" || payload.Monto <= 0 {
		return 0, fmt.Errorf("empresa_id, concepto y monto son obligatorios")
	}
	payload.Estado = normalizeOdontoEstado(payload.Estado, "aplicado")
	if strings.TrimSpace(payload.FechaPago) == "" {
		payload.FechaPago = time.Now().Format("2006-01-02 15:04:05")
	}
	if metodo := NormalizeMetodoPagoCarrito(payload.MetodoPago); metodo != "" {
		payload.MetodoPago = metodo
	} else {
		payload.MetodoPago = "efectivo"
	}
	if payload.PacienteID <= 0 && payload.PresupuestoID > 0 {
		_ = queryRowSQLCompat(dbConn, `SELECT COALESCE(paciente_id,0) FROM empresa_odontologia_presupuestos WHERE empresa_id=? AND id=? LIMIT 1`, payload.EmpresaID, payload.PresupuestoID).Scan(&payload.PacienteID)
	}
	if payload.PacienteID > 0 {
		paciente, err := getEmpresaOdontologiaPacienteByID(dbConn, payload.EmpresaID, payload.PacienteID)
		if err != nil {
			return 0, err
		}
		if paciente.UsuarioCreador == "" {
			paciente.UsuarioCreador = payload.UsuarioCreador
		}
		clienteID, err := syncEmpresaOdontologiaPacienteCliente(dbConn, *paciente)
		if err != nil {
			return 0, err
		}
		payload.ClienteID = clienteID
	}
	servicioID, err := ensureEmpresaOdontologiaPagoServicio(dbConn, payload)
	if err != nil {
		return 0, err
	}
	payload.ServicioID = servicioID
	if payload.Estado != "anulado" {
		carritoID, itemID, err := createEmpresaOdontologiaPagoCarrito(dbConn, payload)
		if err != nil {
			return 0, err
		}
		payload.CarritoID = carritoID
		payload.CarritoItemID = itemID
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_odontologia_pagos (empresa_id,paciente_id,cliente_id,presupuesto_id,servicio_id,carrito_id,carrito_item_id,concepto,monto,metodo_pago,referencia,fecha_pago,estado,observaciones,fecha_creacion,fecha_actualizacion,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,datetime('now','localtime'),datetime('now','localtime'),?)`, payload.EmpresaID, nullableInt64(payload.PacienteID), nullableInt64(payload.ClienteID), nullableInt64(payload.PresupuestoID), nullableInt64(payload.ServicioID), nullableInt64(payload.CarritoID), nullableInt64(payload.CarritoItemID), strings.TrimSpace(payload.Concepto), payload.Monto, strings.TrimSpace(payload.MetodoPago), strings.TrimSpace(payload.Referencia), strings.TrimSpace(payload.FechaPago), payload.Estado, strings.TrimSpace(payload.Observaciones), strings.TrimSpace(payload.UsuarioCreador))
	if err != nil {
		return 0, err
	}
	if payload.PresupuestoID > 0 {
		_, _ = execSQLCompat(dbConn, `UPDATE empresa_odontologia_presupuestos SET saldo = CASE WHEN saldo - ? < 0 THEN 0 ELSE saldo - ? END, estado = CASE WHEN saldo - ? <= 0 THEN 'pagado' ELSE estado END, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ? AND id = ?`, payload.Monto, payload.Monto, payload.Monto, payload.EmpresaID, payload.PresupuestoID)
	}
	if payload.PacienteID > 0 {
		_, _ = execSQLCompat(dbConn, `UPDATE empresa_odontologia_pacientes SET saldo = CASE WHEN saldo - ? < 0 THEN 0 ELSE saldo - ? END, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ? AND id = ?`, payload.Monto, payload.Monto, payload.EmpresaID, payload.PacienteID)
	}
	return id, nil
}

func SyncEmpresaOdontologiaNucleo(dbConn *sql.DB, empresaID int64, usuario string) (*EmpresaOdontologiaIntegracionNucleoResumen, error) {
	if err := EnsureEmpresaOdontologiaSchema(dbConn); err != nil {
		return nil, err
	}
	resumen := &EmpresaOdontologiaIntegracionNucleoResumen{
		EmpresaID:         empresaID,
		EstadoIntegracion: "plantilla_integrada_nucleo",
		VisibleOperativo:  true,
	}
	pacientes, err := ListEmpresaOdontologiaPacientes(dbConn, empresaID)
	if err != nil {
		return nil, err
	}
	for _, paciente := range pacientes {
		if paciente.UsuarioCreador == "" {
			paciente.UsuarioCreador = usuario
		}
		if _, err := syncEmpresaOdontologiaPacienteCliente(dbConn, paciente); err != nil {
			resumen.Errores = append(resumen.Errores, fmt.Sprintf("paciente %d: %v", paciente.ID, err))
			continue
		}
		resumen.PacientesSincronizados++
	}
	tratamientos, err := ListEmpresaOdontologiaTratamientos(dbConn, empresaID)
	if err != nil {
		return nil, err
	}
	for _, tratamiento := range tratamientos {
		if tratamiento.UsuarioCreador == "" {
			tratamiento.UsuarioCreador = usuario
		}
		if _, err := syncEmpresaOdontologiaTratamientoServicio(dbConn, tratamiento); err != nil {
			resumen.Errores = append(resumen.Errores, fmt.Sprintf("tratamiento %d: %v", tratamiento.ID, err))
			continue
		}
		resumen.TratamientosSincronizados++
	}
	pagos, err := ListEmpresaOdontologiaPagos(dbConn, empresaID)
	if err != nil {
		return nil, err
	}
	for _, pago := range pagos {
		if pago.UsuarioCreador == "" {
			pago.UsuarioCreador = usuario
		}
		pago.Estado = normalizeOdontoEstado(pago.Estado, "aplicado")
		if pago.Estado == "anulado" || pago.CarritoID > 0 {
			resumen.PagosPendientes++
			continue
		}
		if pago.PacienteID <= 0 && pago.PresupuestoID > 0 {
			_ = queryRowSQLCompat(dbConn, `SELECT COALESCE(paciente_id,0) FROM empresa_odontologia_presupuestos WHERE empresa_id=? AND id=? LIMIT 1`, pago.EmpresaID, pago.PresupuestoID).Scan(&pago.PacienteID)
		}
		if pago.PacienteID > 0 {
			paciente, err := getEmpresaOdontologiaPacienteByID(dbConn, pago.EmpresaID, pago.PacienteID)
			if err != nil {
				resumen.Errores = append(resumen.Errores, fmt.Sprintf("pago %d paciente: %v", pago.ID, err))
				continue
			}
			if paciente.UsuarioCreador == "" {
				paciente.UsuarioCreador = usuario
			}
			clienteID, err := syncEmpresaOdontologiaPacienteCliente(dbConn, *paciente)
			if err != nil {
				resumen.Errores = append(resumen.Errores, fmt.Sprintf("pago %d cliente: %v", pago.ID, err))
				continue
			}
			pago.ClienteID = clienteID
		}
		if metodo := NormalizeMetodoPagoCarrito(pago.MetodoPago); metodo != "" {
			pago.MetodoPago = metodo
		} else {
			pago.MetodoPago = "efectivo"
		}
		servicioID, err := ensureEmpresaOdontologiaPagoServicio(dbConn, pago)
		if err != nil {
			resumen.Errores = append(resumen.Errores, fmt.Sprintf("pago %d servicio: %v", pago.ID, err))
			continue
		}
		pago.ServicioID = servicioID
		carritoID, itemID, err := createEmpresaOdontologiaPagoCarrito(dbConn, pago)
		if err != nil {
			resumen.Errores = append(resumen.Errores, fmt.Sprintf("pago %d carrito: %v", pago.ID, err))
			continue
		}
		_, err = execSQLCompat(dbConn, `UPDATE empresa_odontologia_pagos SET cliente_id=?, servicio_id=?, carrito_id=?, carrito_item_id=?, metodo_pago=?, fecha_actualizacion=datetime('now','localtime') WHERE empresa_id=? AND id=?`, nullableInt64(pago.ClienteID), nullableInt64(pago.ServicioID), nullableInt64(carritoID), nullableInt64(itemID), pago.MetodoPago, pago.EmpresaID, pago.ID)
		if err != nil {
			resumen.Errores = append(resumen.Errores, fmt.Sprintf("pago %d refs: %v", pago.ID, err))
			continue
		}
		resumen.PagosSincronizados++
	}
	if len(resumen.Errores) > 0 {
		resumen.EstadoIntegracion = "integrado_con_observaciones"
		resumen.RequiereRevisionDatos = true
	}
	return resumen, nil
}
