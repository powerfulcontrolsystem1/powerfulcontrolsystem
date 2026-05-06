package db

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

type EmpresaCierreFiscalPolitica struct {
	ID                           int64  `json:"id"`
	EmpresaID                    int64  `json:"empresa_id"`
	Modulo                       string `json:"modulo"`
	Nombre                       string `json:"nombre"`
	BloqueoAutomatico            bool   `json:"bloqueo_automatico"`
	DiasEdicionRetroactiva       int    `json:"dias_edicion_retroactiva"`
	RequiereAprobacionReapertura bool   `json:"requiere_aprobacion_reapertura"`
	PermiteExcepciones           bool   `json:"permite_excepciones"`
	NotificarCambiosPostCierre   bool   `json:"notificar_cambios_post_cierre"`
	Estado                       string `json:"estado"`
	Observaciones                string `json:"observaciones,omitempty"`
	FechaCreacion                string `json:"fecha_creacion,omitempty"`
	FechaActualizacion           string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador               string `json:"usuario_creador,omitempty"`
}

type EmpresaCierreFiscalPeriodo struct {
	ID                  int64  `json:"id"`
	EmpresaID           int64  `json:"empresa_id"`
	Periodo             string `json:"periodo"`
	FechaDesde          string `json:"fecha_desde"`
	FechaHasta          string `json:"fecha_hasta"`
	TipoCierre          string `json:"tipo_cierre"`
	EstadoPeriodo       string `json:"estado_periodo"`
	BloqueaVentas       bool   `json:"bloquea_ventas"`
	BloqueaCompras      bool   `json:"bloquea_compras"`
	BloqueaCaja         bool   `json:"bloquea_caja"`
	BloqueaInventario   bool   `json:"bloquea_inventario"`
	BloqueaContabilidad bool   `json:"bloquea_contabilidad"`
	BloqueaFacturacion  bool   `json:"bloquea_facturacion"`
	CerradoPor          string `json:"cerrado_por,omitempty"`
	FechaCierre         string `json:"fecha_cierre,omitempty"`
	ReabiertoPor        string `json:"reabierto_por,omitempty"`
	FechaReapertura     string `json:"fecha_reapertura,omitempty"`
	Motivo              string `json:"motivo,omitempty"`
	Observaciones       string `json:"observaciones,omitempty"`
	FechaCreacion       string `json:"fecha_creacion,omitempty"`
	FechaActualizacion  string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador      string `json:"usuario_creador,omitempty"`
}

type EmpresaCierreFiscalExcepcion struct {
	ID              int64  `json:"id"`
	EmpresaID       int64  `json:"empresa_id"`
	PeriodoID       int64  `json:"periodo_id,omitempty"`
	Periodo         string `json:"periodo"`
	Modulo          string `json:"modulo"`
	AccionPermitida string `json:"accion_permitida"`
	DocumentoTipo   string `json:"documento_tipo,omitempty"`
	DocumentoID     int64  `json:"documento_id,omitempty"`
	FechaOperacion  string `json:"fecha_operacion,omitempty"`
	Motivo          string `json:"motivo"`
	AprobadoPor     string `json:"aprobado_por,omitempty"`
	ExpiraEn        string `json:"expira_en,omitempty"`
	Usada           bool   `json:"usada"`
	Estado          string `json:"estado"`
	FechaCreacion   string `json:"fecha_creacion,omitempty"`
	UsuarioCreador  string `json:"usuario_creador,omitempty"`
}

type EmpresaCierreFiscalEvento struct {
	ID             int64  `json:"id"`
	EmpresaID      int64  `json:"empresa_id"`
	PeriodoID      int64  `json:"periodo_id,omitempty"`
	Periodo        string `json:"periodo,omitempty"`
	Modulo         string `json:"modulo"`
	Accion         string `json:"accion"`
	Resultado      string `json:"resultado"`
	DocumentoTipo  string `json:"documento_tipo,omitempty"`
	DocumentoID    int64  `json:"documento_id,omitempty"`
	FechaOperacion string `json:"fecha_operacion,omitempty"`
	Motivo         string `json:"motivo,omitempty"`
	Usuario        string `json:"usuario,omitempty"`
	FechaCreacion  string `json:"fecha_creacion,omitempty"`
}

type EmpresaCierreFiscalValidacion struct {
	Permitido       bool                          `json:"permitido"`
	RequiereMotivo  bool                          `json:"requiere_motivo"`
	RequiereAprobar bool                          `json:"requiere_aprobar"`
	Razon           string                        `json:"razon,omitempty"`
	Periodo         *EmpresaCierreFiscalPeriodo   `json:"periodo,omitempty"`
	Politica        *EmpresaCierreFiscalPolitica  `json:"politica,omitempty"`
	Excepcion       *EmpresaCierreFiscalExcepcion `json:"excepcion,omitempty"`
}

type EmpresaCierreFiscalDashboard struct {
	EmpresaID               int64                          `json:"empresa_id"`
	PeriodoActual           string                         `json:"periodo_actual"`
	PeriodosAbiertos        int                            `json:"periodos_abiertos"`
	PeriodosRevision        int                            `json:"periodos_revision"`
	PeriodosCerrados        int                            `json:"periodos_cerrados"`
	PeriodosBloqueados      int                            `json:"periodos_bloqueados"`
	ExcepcionesActivas      int                            `json:"excepciones_activas"`
	EventosBloqueados30Dias int                            `json:"eventos_bloqueados_30_dias"`
	Politicas               []EmpresaCierreFiscalPolitica  `json:"politicas"`
	Periodos                []EmpresaCierreFiscalPeriodo   `json:"periodos"`
	Excepciones             []EmpresaCierreFiscalExcepcion `json:"excepciones"`
	EventosRecientes        []EmpresaCierreFiscalEvento    `json:"eventos_recientes"`
	Alertas                 []string                       `json:"alertas"`
}

func EnsureEmpresaCierreFiscalSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_cierre_fiscal_politicas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			modulo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			bloqueo_automatico INTEGER DEFAULT 1,
			dias_edicion_retroactiva INTEGER DEFAULT 30,
			requiere_aprobacion_reapertura INTEGER DEFAULT 1,
			permite_excepciones INTEGER DEFAULT 1,
			notificar_cambios_post_cierre INTEGER DEFAULT 1,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_cierre_fiscal_politica_empresa_modulo ON empresa_cierre_fiscal_politicas(empresa_id,modulo)`,
		`CREATE TABLE IF NOT EXISTS empresa_cierre_fiscal_periodos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			periodo TEXT NOT NULL,
			fecha_desde TEXT NOT NULL,
			fecha_hasta TEXT NOT NULL,
			tipo_cierre TEXT DEFAULT 'mensual',
			estado_periodo TEXT DEFAULT 'abierto',
			bloquea_ventas INTEGER DEFAULT 1,
			bloquea_compras INTEGER DEFAULT 1,
			bloquea_caja INTEGER DEFAULT 1,
			bloquea_inventario INTEGER DEFAULT 1,
			bloquea_contabilidad INTEGER DEFAULT 1,
			bloquea_facturacion INTEGER DEFAULT 1,
			cerrado_por TEXT,
			fecha_cierre TEXT,
			reabierto_por TEXT,
			fecha_reapertura TEXT,
			motivo TEXT,
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_cierre_fiscal_periodo_empresa ON empresa_cierre_fiscal_periodos(empresa_id,periodo)`,
		`CREATE INDEX IF NOT EXISTS ix_cierre_fiscal_periodo_fechas ON empresa_cierre_fiscal_periodos(empresa_id,fecha_desde,fecha_hasta,estado_periodo)`,
		`CREATE TABLE IF NOT EXISTS empresa_cierre_fiscal_excepciones (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			periodo_id BIGINT DEFAULT 0,
			periodo TEXT NOT NULL,
			modulo TEXT DEFAULT 'todos',
			accion_permitida TEXT DEFAULT 'actualizar',
			documento_tipo TEXT,
			documento_id BIGINT DEFAULT 0,
			fecha_operacion TEXT,
			motivo TEXT NOT NULL,
			aprobado_por TEXT,
			expira_en TEXT,
			usada INTEGER DEFAULT 0,
			estado TEXT DEFAULT 'activa',
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_cierre_fiscal_excepciones_empresa ON empresa_cierre_fiscal_excepciones(empresa_id,periodo,modulo,estado)`,
		`CREATE TABLE IF NOT EXISTS empresa_cierre_fiscal_eventos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			periodo_id BIGINT DEFAULT 0,
			periodo TEXT,
			modulo TEXT DEFAULT 'general',
			accion TEXT DEFAULT 'validar',
			resultado TEXT DEFAULT 'permitido',
			documento_tipo TEXT,
			documento_id BIGINT DEFAULT 0,
			fecha_operacion TEXT,
			motivo TEXT,
			usuario TEXT,
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT))
		)`,
		`CREATE INDEX IF NOT EXISTS ix_cierre_fiscal_eventos_empresa ON empresa_cierre_fiscal_eventos(empresa_id,fecha_creacion DESC,resultado)`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func UpsertEmpresaCierreFiscalPolitica(dbConn *sql.DB, item EmpresaCierreFiscalPolitica) (int64, error) {
	if err := EnsureEmpresaCierreFiscalSchema(dbConn); err != nil {
		return 0, err
	}
	item = normalizeCierreFiscalPolitica(item)
	if item.EmpresaID <= 0 || item.Modulo == "" || item.Nombre == "" {
		return 0, errors.New("empresa_id, modulo y nombre son obligatorios")
	}
	var id int64
	err := QueryRowCompat(dbConn, `INSERT INTO empresa_cierre_fiscal_politicas (empresa_id,modulo,nombre,bloqueo_automatico,dias_edicion_retroactiva,requiere_aprobacion_reapertura,permite_excepciones,notificar_cambios_post_cierre,estado,observaciones,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (empresa_id,modulo) DO UPDATE SET nombre=EXCLUDED.nombre,bloqueo_automatico=EXCLUDED.bloqueo_automatico,dias_edicion_retroactiva=EXCLUDED.dias_edicion_retroactiva,requiere_aprobacion_reapertura=EXCLUDED.requiere_aprobacion_reapertura,permite_excepciones=EXCLUDED.permite_excepciones,notificar_cambios_post_cierre=EXCLUDED.notificar_cambios_post_cierre,estado=EXCLUDED.estado,observaciones=EXCLUDED.observaciones,fecha_actualizacion=CAST(CURRENT_TIMESTAMP AS TEXT),usuario_creador=EXCLUDED.usuario_creador
		RETURNING id`,
		item.EmpresaID, item.Modulo, item.Nombre, boolIntCierreFiscal(item.BloqueoAutomatico), item.DiasEdicionRetroactiva, boolIntCierreFiscal(item.RequiereAprobacionReapertura), boolIntCierreFiscal(item.PermiteExcepciones), boolIntCierreFiscal(item.NotificarCambiosPostCierre), item.Estado, item.Observaciones, item.UsuarioCreador).Scan(&id)
	return id, err
}

func ListEmpresaCierreFiscalPoliticas(dbConn *sql.DB, empresaID int64) ([]EmpresaCierreFiscalPolitica, error) {
	if err := EnsureEmpresaCierreFiscalSchema(dbConn); err != nil {
		return nil, err
	}
	if err := SeedEmpresaCierreFiscalPoliticasBase(dbConn, empresaID, "sistema"); err != nil {
		return nil, err
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,COALESCE(modulo,''),COALESCE(nombre,''),COALESCE(bloqueo_automatico,1),COALESCE(dias_edicion_retroactiva,30),COALESCE(requiere_aprobacion_reapertura,1),COALESCE(permite_excepciones,1),COALESCE(notificar_cambios_post_cierre,1),COALESCE(estado,'activo'),COALESCE(observaciones,''),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,''),COALESCE(usuario_creador,'') FROM empresa_cierre_fiscal_politicas WHERE empresa_id=? ORDER BY modulo`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCierreFiscalPolitica{}
	for rows.Next() {
		var x EmpresaCierreFiscalPolitica
		var bloqueo, reapertura, excepciones, notificar int
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Modulo, &x.Nombre, &bloqueo, &x.DiasEdicionRetroactiva, &reapertura, &excepciones, &notificar, &x.Estado, &x.Observaciones, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		x.BloqueoAutomatico = bloqueo > 0
		x.RequiereAprobacionReapertura = reapertura > 0
		x.PermiteExcepciones = excepciones > 0
		x.NotificarCambiosPostCierre = notificar > 0
		out = append(out, x)
	}
	return out, rows.Err()
}

func UpsertEmpresaCierreFiscalPeriodo(dbConn *sql.DB, item EmpresaCierreFiscalPeriodo) (int64, error) {
	if err := EnsureEmpresaCierreFiscalSchema(dbConn); err != nil {
		return 0, err
	}
	item = normalizeCierreFiscalPeriodoRow(item)
	if item.EmpresaID <= 0 || item.Periodo == "" || item.FechaDesde == "" || item.FechaHasta == "" {
		return 0, errors.New("empresa_id, periodo, fecha_desde y fecha_hasta son obligatorios")
	}
	var id int64
	err := QueryRowCompat(dbConn, `INSERT INTO empresa_cierre_fiscal_periodos (empresa_id,periodo,fecha_desde,fecha_hasta,tipo_cierre,estado_periodo,bloquea_ventas,bloquea_compras,bloquea_caja,bloquea_inventario,bloquea_contabilidad,bloquea_facturacion,cerrado_por,fecha_cierre,reabierto_por,fecha_reapertura,motivo,observaciones,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (empresa_id,periodo) DO UPDATE SET fecha_desde=EXCLUDED.fecha_desde,fecha_hasta=EXCLUDED.fecha_hasta,tipo_cierre=EXCLUDED.tipo_cierre,estado_periodo=EXCLUDED.estado_periodo,bloquea_ventas=EXCLUDED.bloquea_ventas,bloquea_compras=EXCLUDED.bloquea_compras,bloquea_caja=EXCLUDED.bloquea_caja,bloquea_inventario=EXCLUDED.bloquea_inventario,bloquea_contabilidad=EXCLUDED.bloquea_contabilidad,bloquea_facturacion=EXCLUDED.bloquea_facturacion,cerrado_por=EXCLUDED.cerrado_por,fecha_cierre=EXCLUDED.fecha_cierre,reabierto_por=EXCLUDED.reabierto_por,fecha_reapertura=EXCLUDED.fecha_reapertura,motivo=EXCLUDED.motivo,observaciones=EXCLUDED.observaciones,fecha_actualizacion=CAST(CURRENT_TIMESTAMP AS TEXT),usuario_creador=EXCLUDED.usuario_creador
		RETURNING id`,
		item.EmpresaID, item.Periodo, item.FechaDesde, item.FechaHasta, item.TipoCierre, item.EstadoPeriodo, boolIntCierreFiscal(item.BloqueaVentas), boolIntCierreFiscal(item.BloqueaCompras), boolIntCierreFiscal(item.BloqueaCaja), boolIntCierreFiscal(item.BloqueaInventario), boolIntCierreFiscal(item.BloqueaContabilidad), boolIntCierreFiscal(item.BloqueaFacturacion), item.CerradoPor, item.FechaCierre, item.ReabiertoPor, item.FechaReapertura, item.Motivo, item.Observaciones, item.UsuarioCreador).Scan(&id)
	return id, err
}

func CambiarEstadoEmpresaCierreFiscalPeriodo(dbConn *sql.DB, empresaID, periodoID int64, estado, usuario, motivo string) (EmpresaCierreFiscalPeriodo, error) {
	if empresaID <= 0 || periodoID <= 0 {
		return EmpresaCierreFiscalPeriodo{}, errors.New("empresa_id y periodo_id son obligatorios")
	}
	estado = normalizeCierreFiscalEstadoPeriodo(estado)
	motivo = strings.TrimSpace(motivo)
	if (estado == "cerrado" || estado == "bloqueado" || estado == "abierto") && motivo == "" {
		return EmpresaCierreFiscalPeriodo{}, errors.New("motivo es obligatorio para cerrar, bloquear o reabrir")
	}
	periodo, err := GetEmpresaCierreFiscalPeriodo(dbConn, empresaID, periodoID)
	if err != nil {
		return EmpresaCierreFiscalPeriodo{}, err
	}
	periodo.EstadoPeriodo = estado
	periodo.Motivo = motivo
	periodo.UsuarioCreador = usuario
	if estado == "cerrado" || estado == "bloqueado" {
		periodo.CerradoPor = usuario
		periodo.FechaCierre = time.Now().Format(time.RFC3339)
	}
	if estado == "abierto" {
		periodo.ReabiertoPor = usuario
		periodo.FechaReapertura = time.Now().Format(time.RFC3339)
	}
	if _, err := UpsertEmpresaCierreFiscalPeriodo(dbConn, periodo); err != nil {
		return EmpresaCierreFiscalPeriodo{}, err
	}
	_ = InsertEmpresaCierreFiscalEvento(dbConn, EmpresaCierreFiscalEvento{EmpresaID: empresaID, PeriodoID: periodo.ID, Periodo: periodo.Periodo, Modulo: "cierre_fiscal", Accion: estado, Resultado: "ok", Motivo: motivo, Usuario: usuario})
	return GetEmpresaCierreFiscalPeriodo(dbConn, empresaID, periodoID)
}

func GetEmpresaCierreFiscalPeriodo(dbConn *sql.DB, empresaID, id int64) (EmpresaCierreFiscalPeriodo, error) {
	if err := EnsureEmpresaCierreFiscalSchema(dbConn); err != nil {
		return EmpresaCierreFiscalPeriodo{}, err
	}
	var x EmpresaCierreFiscalPeriodo
	var ventas, compras, caja, inventario, contabilidad, facturacion int
	err := QueryRowCompat(dbConn, cierreFiscalPeriodoSelectSQL()+` WHERE empresa_id=? AND id=?`, empresaID, id).Scan(&x.ID, &x.EmpresaID, &x.Periodo, &x.FechaDesde, &x.FechaHasta, &x.TipoCierre, &x.EstadoPeriodo, &ventas, &compras, &caja, &inventario, &contabilidad, &facturacion, &x.CerradoPor, &x.FechaCierre, &x.ReabiertoPor, &x.FechaReapertura, &x.Motivo, &x.Observaciones, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador)
	if err != nil {
		return EmpresaCierreFiscalPeriodo{}, err
	}
	x.BloqueaVentas, x.BloqueaCompras, x.BloqueaCaja = ventas > 0, compras > 0, caja > 0
	x.BloqueaInventario, x.BloqueaContabilidad, x.BloqueaFacturacion = inventario > 0, contabilidad > 0, facturacion > 0
	return x, nil
}

func ListEmpresaCierreFiscalPeriodos(dbConn *sql.DB, empresaID int64, estado string, limit int) ([]EmpresaCierreFiscalPeriodo, error) {
	if err := EnsureEmpresaCierreFiscalSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 120
	}
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if strings.TrimSpace(estado) != "" {
		where += " AND estado_periodo=?"
		args = append(args, normalizeCierreFiscalEstadoPeriodo(estado))
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(cierreFiscalPeriodoSelectSQL()+` WHERE %s ORDER BY periodo DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCierreFiscalPeriodo{}
	for rows.Next() {
		var x EmpresaCierreFiscalPeriodo
		var ventas, compras, caja, inventario, contabilidad, facturacion int
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Periodo, &x.FechaDesde, &x.FechaHasta, &x.TipoCierre, &x.EstadoPeriodo, &ventas, &compras, &caja, &inventario, &contabilidad, &facturacion, &x.CerradoPor, &x.FechaCierre, &x.ReabiertoPor, &x.FechaReapertura, &x.Motivo, &x.Observaciones, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		x.BloqueaVentas, x.BloqueaCompras, x.BloqueaCaja = ventas > 0, compras > 0, caja > 0
		x.BloqueaInventario, x.BloqueaContabilidad, x.BloqueaFacturacion = inventario > 0, contabilidad > 0, facturacion > 0
		out = append(out, x)
	}
	return out, rows.Err()
}

func CreateEmpresaCierreFiscalExcepcion(dbConn *sql.DB, item EmpresaCierreFiscalExcepcion) (int64, error) {
	if err := EnsureEmpresaCierreFiscalSchema(dbConn); err != nil {
		return 0, err
	}
	item = normalizeCierreFiscalExcepcion(item)
	if item.EmpresaID <= 0 || item.Periodo == "" || item.Motivo == "" {
		return 0, errors.New("empresa_id, periodo y motivo son obligatorios")
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_cierre_fiscal_excepciones (empresa_id,periodo_id,periodo,modulo,accion_permitida,documento_tipo,documento_id,fecha_operacion,motivo,aprobado_por,expira_en,usada,estado,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		item.EmpresaID, item.PeriodoID, item.Periodo, item.Modulo, item.AccionPermitida, item.DocumentoTipo, item.DocumentoID, item.FechaOperacion, item.Motivo, item.AprobadoPor, item.ExpiraEn, boolIntCierreFiscal(item.Usada), item.Estado, item.UsuarioCreador)
}

func ListEmpresaCierreFiscalExcepciones(dbConn *sql.DB, empresaID int64, periodo, estado string, limit int) ([]EmpresaCierreFiscalExcepcion, error) {
	if err := EnsureEmpresaCierreFiscalSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 120
	}
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if strings.TrimSpace(periodo) != "" {
		where += " AND periodo=?"
		args = append(args, normalizeCierreFiscalPeriodo(periodo))
	}
	if strings.TrimSpace(estado) != "" {
		where += " AND estado=?"
		args = append(args, normalizeCierreFiscalEstadoExcepcion(estado))
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,COALESCE(periodo_id,0),COALESCE(periodo,''),COALESCE(modulo,'todos'),COALESCE(accion_permitida,'actualizar'),COALESCE(documento_tipo,''),COALESCE(documento_id,0),COALESCE(fecha_operacion,''),COALESCE(motivo,''),COALESCE(aprobado_por,''),COALESCE(expira_en,''),COALESCE(usada,0),COALESCE(estado,'activa'),COALESCE(fecha_creacion,''),COALESCE(usuario_creador,'') FROM empresa_cierre_fiscal_excepciones WHERE %s ORDER BY id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCierreFiscalExcepcion{}
	for rows.Next() {
		var x EmpresaCierreFiscalExcepcion
		var usada int
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.PeriodoID, &x.Periodo, &x.Modulo, &x.AccionPermitida, &x.DocumentoTipo, &x.DocumentoID, &x.FechaOperacion, &x.Motivo, &x.AprobadoPor, &x.ExpiraEn, &usada, &x.Estado, &x.FechaCreacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		x.Usada = usada > 0
		out = append(out, x)
	}
	return out, rows.Err()
}

func ValidarEmpresaCierreFiscalOperacion(dbConn *sql.DB, empresaID int64, fechaOperacion, modulo, accion, documentoTipo string, documentoID int64, usuario string, registrar bool) (EmpresaCierreFiscalValidacion, error) {
	if err := EnsureEmpresaCierreFiscalSchema(dbConn); err != nil {
		return EmpresaCierreFiscalValidacion{}, err
	}
	fechaOperacion = strings.TrimSpace(fechaOperacion)
	if fechaOperacion == "" {
		fechaOperacion = time.Now().Format("2006-01-02")
	}
	modulo = normalizeCierreFiscalModulo(modulo)
	accion = normalizeCierreFiscalAccion(accion)
	result := EmpresaCierreFiscalValidacion{Permitido: true}
	politica := getCierreFiscalPoliticaForModulo(dbConn, empresaID, modulo)
	if politica != nil {
		result.Politica = politica
	}
	periodo, err := getCierreFiscalPeriodoByFecha(dbConn, empresaID, fechaOperacion)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return result, err
	}
	if periodo != nil {
		result.Periodo = periodo
		if cierreFiscalPeriodoBloqueaModulo(*periodo, modulo) && periodo.EstadoPeriodo != "abierto" {
			ex := findCierreFiscalExcepcion(dbConn, empresaID, periodo.Periodo, modulo, accion, documentoTipo, documentoID)
			if ex != nil && politica != nil && politica.PermiteExcepciones {
				result.Excepcion = ex
				result.RequiereMotivo = true
				result.Razon = "Operacion permitida por excepcion aprobada"
			} else {
				result.Permitido = false
				result.RequiereAprobar = true
				result.Razon = fmt.Sprintf("Periodo %s en estado %s bloquea %s", periodo.Periodo, periodo.EstadoPeriodo, modulo)
			}
		}
	}
	if result.Permitido && politica != nil && politica.BloqueoAutomatico && politica.DiasEdicionRetroactiva > 0 && olderThanCierreFiscalDays(fechaOperacion, politica.DiasEdicionRetroactiva) {
		result.Permitido = false
		result.RequiereAprobar = true
		result.Razon = fmt.Sprintf("Fecha supera %d dias de edicion retroactiva para %s", politica.DiasEdicionRetroactiva, modulo)
	}
	if registrar {
		res := "permitido"
		if !result.Permitido {
			res = "bloqueado"
		}
		periodoID := int64(0)
		periodoCode := normalizeCierreFiscalPeriodoFromFecha(fechaOperacion)
		if result.Periodo != nil {
			periodoID = result.Periodo.ID
			periodoCode = result.Periodo.Periodo
		}
		_ = InsertEmpresaCierreFiscalEvento(dbConn, EmpresaCierreFiscalEvento{EmpresaID: empresaID, PeriodoID: periodoID, Periodo: periodoCode, Modulo: modulo, Accion: accion, Resultado: res, DocumentoTipo: documentoTipo, DocumentoID: documentoID, FechaOperacion: fechaOperacion, Motivo: result.Razon, Usuario: usuario})
	}
	return result, nil
}

func InsertEmpresaCierreFiscalEvento(dbConn *sql.DB, item EmpresaCierreFiscalEvento) error {
	if item.EmpresaID <= 0 {
		return nil
	}
	item.Modulo = normalizeCierreFiscalModulo(item.Modulo)
	item.Accion = normalizeCierreFiscalAccion(item.Accion)
	item.Resultado = normalizeOneOfCierreFiscal(item.Resultado, "permitido", "permitido", "bloqueado", "ok", "error", "advertencia")
	_, err := ExecCompat(dbConn, `INSERT INTO empresa_cierre_fiscal_eventos (empresa_id,periodo_id,periodo,modulo,accion,resultado,documento_tipo,documento_id,fecha_operacion,motivo,usuario) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		item.EmpresaID, item.PeriodoID, item.Periodo, item.Modulo, item.Accion, item.Resultado, strings.TrimSpace(item.DocumentoTipo), item.DocumentoID, strings.TrimSpace(item.FechaOperacion), strings.TrimSpace(item.Motivo), strings.TrimSpace(item.Usuario))
	return err
}

func ListEmpresaCierreFiscalEventos(dbConn *sql.DB, empresaID int64, limit int) ([]EmpresaCierreFiscalEvento, error) {
	if err := EnsureEmpresaCierreFiscalSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 300 {
		limit = 80
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,COALESCE(periodo_id,0),COALESCE(periodo,''),COALESCE(modulo,'general'),COALESCE(accion,'validar'),COALESCE(resultado,'permitido'),COALESCE(documento_tipo,''),COALESCE(documento_id,0),COALESCE(fecha_operacion,''),COALESCE(motivo,''),COALESCE(usuario,''),COALESCE(fecha_creacion,'') FROM empresa_cierre_fiscal_eventos WHERE empresa_id=? ORDER BY id DESC LIMIT %d`, limit), empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCierreFiscalEvento{}
	for rows.Next() {
		var x EmpresaCierreFiscalEvento
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.PeriodoID, &x.Periodo, &x.Modulo, &x.Accion, &x.Resultado, &x.DocumentoTipo, &x.DocumentoID, &x.FechaOperacion, &x.Motivo, &x.Usuario, &x.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func BuildEmpresaCierreFiscalDashboard(dbConn *sql.DB, empresaID int64) (EmpresaCierreFiscalDashboard, error) {
	if err := EnsureEmpresaCierreFiscalSchema(dbConn); err != nil {
		return EmpresaCierreFiscalDashboard{}, err
	}
	if err := SeedEmpresaCierreFiscalPoliticasBase(dbConn, empresaID, "sistema"); err != nil {
		return EmpresaCierreFiscalDashboard{}, err
	}
	politicas, _ := ListEmpresaCierreFiscalPoliticas(dbConn, empresaID)
	periodos, _ := ListEmpresaCierreFiscalPeriodos(dbConn, empresaID, "", 80)
	excepciones, _ := ListEmpresaCierreFiscalExcepciones(dbConn, empresaID, "", "activa", 80)
	eventos, _ := ListEmpresaCierreFiscalEventos(dbConn, empresaID, 80)
	out := EmpresaCierreFiscalDashboard{EmpresaID: empresaID, PeriodoActual: time.Now().Format("2006-01"), Politicas: politicas, Periodos: periodos, Excepciones: excepciones, EventosRecientes: eventos}
	for _, p := range periodos {
		switch p.EstadoPeriodo {
		case "abierto":
			out.PeriodosAbiertos++
		case "en_revision":
			out.PeriodosRevision++
		case "cerrado":
			out.PeriodosCerrados++
		case "bloqueado":
			out.PeriodosBloqueados++
		}
	}
	for _, ex := range excepciones {
		if ex.Estado == "activa" && !ex.Usada {
			out.ExcepcionesActivas++
		}
	}
	cut := time.Now().AddDate(0, 0, -30)
	for _, ev := range eventos {
		if ev.Resultado == "bloqueado" {
			t, _ := time.Parse(time.RFC3339, ev.FechaCreacion)
			if t.IsZero() {
				t, _ = time.Parse("2006-01-02 15:04:05", ev.FechaCreacion)
			}
			if t.IsZero() || t.After(cut) {
				out.EventosBloqueados30Dias++
			}
		}
	}
	if out.PeriodosAbiertos == 0 {
		out.Alertas = append(out.Alertas, "No hay periodos abiertos registrados para operar el calendario fiscal.")
	}
	if out.EventosBloqueados30Dias > 0 {
		out.Alertas = append(out.Alertas, fmt.Sprintf("%d operaciones fueron bloqueadas en los ultimos 30 dias.", out.EventosBloqueados30Dias))
	}
	if len(out.Alertas) == 0 {
		out.Alertas = append(out.Alertas, "Politicas de cierre fiscal sin alertas criticas.")
	}
	return out, nil
}

func SeedEmpresaCierreFiscalDemo(dbConn *sql.DB, empresaID int64, usuario string) error {
	if err := SeedEmpresaCierreFiscalPoliticasBase(dbConn, empresaID, usuario); err != nil {
		return err
	}
	now := time.Now()
	prev := now.AddDate(0, -1, 0)
	curPeriod := now.Format("2006-01")
	prevPeriod := prev.Format("2006-01")
	_, err := UpsertEmpresaCierreFiscalPeriodo(dbConn, EmpresaCierreFiscalPeriodo{EmpresaID: empresaID, Periodo: prevPeriod, FechaDesde: prevPeriod + "-01", FechaHasta: lastDayOfCierreFiscalMonth(prev), TipoCierre: "mensual", EstadoPeriodo: "cerrado", Motivo: "Cierre demo de periodo anterior", Observaciones: "Periodo cerrado para pruebas de bloqueo fiscal.", UsuarioCreador: usuario, CerradoPor: usuario, FechaCierre: now.Format(time.RFC3339), BloqueaVentas: true, BloqueaCompras: true, BloqueaCaja: true, BloqueaInventario: true, BloqueaContabilidad: true, BloqueaFacturacion: true})
	if err != nil {
		return err
	}
	id, err := UpsertEmpresaCierreFiscalPeriodo(dbConn, EmpresaCierreFiscalPeriodo{EmpresaID: empresaID, Periodo: curPeriod, FechaDesde: curPeriod + "-01", FechaHasta: lastDayOfCierreFiscalMonth(now), TipoCierre: "mensual", EstadoPeriodo: "abierto", Motivo: "Periodo operativo actual", UsuarioCreador: usuario, BloqueaVentas: true, BloqueaCompras: true, BloqueaCaja: true, BloqueaInventario: true, BloqueaContabilidad: true, BloqueaFacturacion: true})
	if err != nil {
		return err
	}
	_, _ = CreateEmpresaCierreFiscalExcepcion(dbConn, EmpresaCierreFiscalExcepcion{EmpresaID: empresaID, PeriodoID: id, Periodo: curPeriod, Modulo: "contabilidad", AccionPermitida: "actualizar", Motivo: "Ajuste contable demo aprobado antes del cierre definitivo.", AprobadoPor: usuario, ExpiraEn: now.AddDate(0, 0, 7).Format("2006-01-02"), Estado: "activa", UsuarioCreador: usuario})
	return nil
}

func SeedEmpresaCierreFiscalPoliticasBase(dbConn *sql.DB, empresaID int64, usuario string) error {
	defaults := []EmpresaCierreFiscalPolitica{
		{EmpresaID: empresaID, Modulo: "ventas", Nombre: "Ventas y POS", DiasEdicionRetroactiva: 15, BloqueoAutomatico: true, RequiereAprobacionReapertura: true, PermiteExcepciones: true, NotificarCambiosPostCierre: true, Estado: "activo", UsuarioCreador: usuario},
		{EmpresaID: empresaID, Modulo: "compras", Nombre: "Compras y gastos", DiasEdicionRetroactiva: 30, BloqueoAutomatico: true, RequiereAprobacionReapertura: true, PermiteExcepciones: true, NotificarCambiosPostCierre: true, Estado: "activo", UsuarioCreador: usuario},
		{EmpresaID: empresaID, Modulo: "caja", Nombre: "Caja y tesoreria", DiasEdicionRetroactiva: 7, BloqueoAutomatico: true, RequiereAprobacionReapertura: true, PermiteExcepciones: true, NotificarCambiosPostCierre: true, Estado: "activo", UsuarioCreador: usuario},
		{EmpresaID: empresaID, Modulo: "inventario", Nombre: "Inventario y costos", DiasEdicionRetroactiva: 30, BloqueoAutomatico: true, RequiereAprobacionReapertura: true, PermiteExcepciones: true, NotificarCambiosPostCierre: true, Estado: "activo", UsuarioCreador: usuario},
		{EmpresaID: empresaID, Modulo: "contabilidad", Nombre: "Contabilidad y libros", DiasEdicionRetroactiva: 30, BloqueoAutomatico: true, RequiereAprobacionReapertura: true, PermiteExcepciones: true, NotificarCambiosPostCierre: true, Estado: "activo", UsuarioCreador: usuario},
		{EmpresaID: empresaID, Modulo: "facturacion", Nombre: "Facturacion electronica", DiasEdicionRetroactiva: 5, BloqueoAutomatico: true, RequiereAprobacionReapertura: true, PermiteExcepciones: false, NotificarCambiosPostCierre: true, Estado: "activo", UsuarioCreador: usuario},
	}
	for _, item := range defaults {
		if _, err := UpsertEmpresaCierreFiscalPolitica(dbConn, item); err != nil {
			return err
		}
	}
	return nil
}

func cierreFiscalPeriodoSelectSQL() string {
	return `SELECT id,empresa_id,COALESCE(periodo,''),COALESCE(fecha_desde,''),COALESCE(fecha_hasta,''),COALESCE(tipo_cierre,'mensual'),COALESCE(estado_periodo,'abierto'),COALESCE(bloquea_ventas,1),COALESCE(bloquea_compras,1),COALESCE(bloquea_caja,1),COALESCE(bloquea_inventario,1),COALESCE(bloquea_contabilidad,1),COALESCE(bloquea_facturacion,1),COALESCE(cerrado_por,''),COALESCE(fecha_cierre,''),COALESCE(reabierto_por,''),COALESCE(fecha_reapertura,''),COALESCE(motivo,''),COALESCE(observaciones,''),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,''),COALESCE(usuario_creador,'') FROM empresa_cierre_fiscal_periodos`
}

func getCierreFiscalPeriodoByFecha(dbConn *sql.DB, empresaID int64, fecha string) (*EmpresaCierreFiscalPeriodo, error) {
	rows, err := ListEmpresaCierreFiscalPeriodos(dbConn, empresaID, "", 240)
	if err != nil {
		return nil, err
	}
	for _, p := range rows {
		if p.FechaDesde <= fecha && p.FechaHasta >= fecha {
			cp := p
			return &cp, nil
		}
	}
	return nil, sql.ErrNoRows
}

func getCierreFiscalPoliticaForModulo(dbConn *sql.DB, empresaID int64, modulo string) *EmpresaCierreFiscalPolitica {
	rows, err := ListEmpresaCierreFiscalPoliticas(dbConn, empresaID)
	if err != nil {
		return nil
	}
	modulo = normalizeCierreFiscalModulo(modulo)
	for _, p := range rows {
		if p.Modulo == modulo {
			cp := p
			return &cp
		}
	}
	return nil
}

func findCierreFiscalExcepcion(dbConn *sql.DB, empresaID int64, periodo, modulo, accion, documentoTipo string, documentoID int64) *EmpresaCierreFiscalExcepcion {
	rows, err := ListEmpresaCierreFiscalExcepciones(dbConn, empresaID, periodo, "activa", 100)
	if err != nil {
		return nil
	}
	modulo, accion = normalizeCierreFiscalModulo(modulo), normalizeCierreFiscalAccion(accion)
	documentoTipo = strings.TrimSpace(documentoTipo)
	now := time.Now().Format("2006-01-02")
	for _, ex := range rows {
		if ex.Usada || ex.Estado != "activa" {
			continue
		}
		if ex.ExpiraEn != "" && ex.ExpiraEn < now {
			continue
		}
		if ex.Modulo != "todos" && ex.Modulo != modulo {
			continue
		}
		if ex.AccionPermitida != "todas" && ex.AccionPermitida != accion {
			continue
		}
		if ex.DocumentoTipo != "" && !strings.EqualFold(ex.DocumentoTipo, documentoTipo) {
			continue
		}
		if ex.DocumentoID > 0 && ex.DocumentoID != documentoID {
			continue
		}
		cp := ex
		return &cp
	}
	return nil
}

func cierreFiscalPeriodoBloqueaModulo(p EmpresaCierreFiscalPeriodo, modulo string) bool {
	switch normalizeCierreFiscalModulo(modulo) {
	case "ventas":
		return p.BloqueaVentas
	case "compras":
		return p.BloqueaCompras
	case "caja", "tesoreria":
		return p.BloqueaCaja
	case "inventario":
		return p.BloqueaInventario
	case "contabilidad":
		return p.BloqueaContabilidad
	case "facturacion":
		return p.BloqueaFacturacion
	default:
		return p.BloqueaContabilidad || p.BloqueaVentas || p.BloqueaCompras
	}
}

func normalizeCierreFiscalPolitica(x EmpresaCierreFiscalPolitica) EmpresaCierreFiscalPolitica {
	x.Modulo = normalizeCierreFiscalModulo(x.Modulo)
	x.Nombre = strings.TrimSpace(x.Nombre)
	if x.Nombre == "" {
		x.Nombre = strings.Title(strings.ReplaceAll(x.Modulo, "_", " "))
	}
	if x.DiasEdicionRetroactiva < 0 {
		x.DiasEdicionRetroactiva = 0
	}
	if x.DiasEdicionRetroactiva > 3650 {
		x.DiasEdicionRetroactiva = 3650
	}
	x.Estado = normalizeOneOfCierreFiscal(x.Estado, "activo", "activo", "inactivo")
	x.Observaciones = strings.TrimSpace(x.Observaciones)
	x.UsuarioCreador = strings.TrimSpace(x.UsuarioCreador)
	return x
}

func normalizeCierreFiscalPeriodoRow(x EmpresaCierreFiscalPeriodo) EmpresaCierreFiscalPeriodo {
	x.Periodo = normalizeCierreFiscalPeriodo(x.Periodo)
	if x.Periodo == "" {
		x.Periodo = normalizeCierreFiscalPeriodoFromFecha(x.FechaDesde)
	}
	if x.FechaDesde == "" && x.Periodo != "" {
		x.FechaDesde = x.Periodo + "-01"
	}
	if x.FechaHasta == "" && x.Periodo != "" {
		t, _ := time.Parse("2006-01", x.Periodo)
		x.FechaHasta = lastDayOfCierreFiscalMonth(t)
	}
	x.TipoCierre = normalizeOneOfCierreFiscal(x.TipoCierre, "mensual", "mensual", "bimestral", "trimestral", "anual", "manual")
	x.EstadoPeriodo = normalizeCierreFiscalEstadoPeriodo(x.EstadoPeriodo)
	if !x.BloqueaVentas && !x.BloqueaCompras && !x.BloqueaCaja && !x.BloqueaInventario && !x.BloqueaContabilidad && !x.BloqueaFacturacion {
		x.BloqueaVentas, x.BloqueaCompras, x.BloqueaCaja = true, true, true
		x.BloqueaInventario, x.BloqueaContabilidad, x.BloqueaFacturacion = true, true, true
	}
	x.CerradoPor = strings.TrimSpace(x.CerradoPor)
	x.ReabiertoPor = strings.TrimSpace(x.ReabiertoPor)
	x.Motivo = strings.TrimSpace(x.Motivo)
	x.Observaciones = strings.TrimSpace(x.Observaciones)
	x.UsuarioCreador = strings.TrimSpace(x.UsuarioCreador)
	return x
}

func normalizeCierreFiscalExcepcion(x EmpresaCierreFiscalExcepcion) EmpresaCierreFiscalExcepcion {
	x.Periodo = normalizeCierreFiscalPeriodo(x.Periodo)
	x.Modulo = normalizeCierreFiscalModulo(x.Modulo)
	if x.Modulo == "" {
		x.Modulo = "todos"
	}
	x.AccionPermitida = normalizeCierreFiscalAccion(x.AccionPermitida)
	if x.AccionPermitida == "" {
		x.AccionPermitida = "actualizar"
	}
	x.DocumentoTipo = strings.TrimSpace(x.DocumentoTipo)
	x.FechaOperacion = strings.TrimSpace(x.FechaOperacion)
	x.Motivo = strings.TrimSpace(x.Motivo)
	x.AprobadoPor = strings.TrimSpace(x.AprobadoPor)
	x.ExpiraEn = strings.TrimSpace(x.ExpiraEn)
	x.Estado = normalizeCierreFiscalEstadoExcepcion(x.Estado)
	x.UsuarioCreador = strings.TrimSpace(x.UsuarioCreador)
	return x
}

func normalizeCierreFiscalModulo(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	v = strings.ReplaceAll(v, " ", "_")
	switch v {
	case "pos", "venta", "ventas_pos":
		return "ventas"
	case "tesoreria_presupuesto", "finanzas":
		return "caja"
	case "contabilidad_colombia", "contabilidad_colombia_avanzada":
		return "contabilidad"
	case "facturacion_electronica":
		return "facturacion"
	case "":
		return "general"
	default:
		return v
	}
}

func normalizeCierreFiscalAccion(v string) string {
	return normalizeOneOfCierreFiscal(v, "actualizar", "crear", "actualizar", "eliminar", "anular", "aprobar", "emitir", "contabilizar", "reabrir", "todas", "validar")
}

func normalizeCierreFiscalEstadoPeriodo(v string) string {
	return normalizeOneOfCierreFiscal(v, "abierto", "abierto", "en_revision", "cerrado", "bloqueado")
}

func normalizeCierreFiscalEstadoExcepcion(v string) string {
	return normalizeOneOfCierreFiscal(v, "activa", "activa", "usada", "revocada", "vencida")
}

func normalizeOneOfCierreFiscal(v, fallback string, allowed ...string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	v = strings.ReplaceAll(v, " ", "_")
	if v == "" {
		return fallback
	}
	for _, item := range allowed {
		if v == item {
			return v
		}
	}
	return fallback
}

func normalizeCierreFiscalPeriodo(v string) string {
	v = strings.TrimSpace(v)
	if len(v) >= 7 {
		return v[:7]
	}
	return v
}

func normalizeCierreFiscalPeriodoFromFecha(v string) string {
	v = strings.TrimSpace(v)
	if len(v) >= 7 {
		return v[:7]
	}
	return time.Now().Format("2006-01")
}

func olderThanCierreFiscalDays(fecha string, days int) bool {
	if days <= 0 {
		return false
	}
	t, err := time.Parse("2006-01-02", strings.TrimSpace(fecha)[:int(math.Min(float64(len(strings.TrimSpace(fecha))), 10))])
	if err != nil {
		return false
	}
	cut := time.Now().AddDate(0, 0, -days)
	return t.Before(cut)
}

func lastDayOfCierreFiscalMonth(t time.Time) string {
	if t.IsZero() {
		t = time.Now()
	}
	firstNext := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, time.Local)
	return firstNext.AddDate(0, 0, -1).Format("2006-01-02")
}

func boolIntCierreFiscal(v bool) int {
	if v {
		return 1
	}
	return 0
}
