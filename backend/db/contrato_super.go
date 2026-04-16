package db

import (
	"database/sql"
	"errors"
	"strings"
)

type SuperContractVersion struct {
	ID                 int64  `json:"id"`
	Version            int    `json:"version"`
	Titulo             string `json:"titulo"`
	Resumen            string `json:"resumen"`
	Contenido          string `json:"contenido"`
	NotaAceptacion     string `json:"nota_aceptacion"`
	ResumenCambio      string `json:"resumen_cambio"`
	FechaCreacion      string `json:"fecha_creacion"`
	FechaActualizacion string `json:"fecha_actualizacion"`
	UsuarioCreador     string `json:"usuario_creador"`
	Estado             string `json:"estado"`
	Observaciones      string `json:"observaciones,omitempty"`
}

type SuperContractAcceptance struct {
	Acepta bool   `json:"acepta"`
	Version int   `json:"version"`
	Fecha   string `json:"fecha"`
}

func defaultSuperContractVersion() SuperContractVersion {
	return SuperContractVersion{
		Version: 1,
		Titulo:  "Contrato de uso de Powerful Control System",
		Resumen: "Contrato marco de uso, licenciamiento, seguridad y tratamiento operativo de datos para la plataforma POS multiempresa Powerful Control System.",
		Contenido: strings.TrimSpace(`1. Objeto
Powerful Control System presta una plataforma SaaS multiempresa para la gestion operativa, comercial, administrativa y documental de empresas que operan sobre la solucion.

2. Alcance del servicio
El servicio incluye acceso a modulos habilitados por licencia, panel de super administracion, configuraciones globales, herramientas de operacion por empresa e integraciones publicadas por el proveedor segun el plan contratado.

3. Registro y acceso
La persona que ingresa como administrador declara que suministra informacion veraz, que tiene facultades suficientes para actuar por cuenta propia o en nombre de la empresa y que protegera sus credenciales, sesiones y accesos autorizados.

4. Uso autorizado
El usuario se compromete a utilizar la plataforma de forma licita, sin realizar fraude, suplantacion, abuso de recursos, pruebas destructivas, distribucion de malware ni cargas automatizadas que comprometan la disponibilidad del servicio.

5. Responsabilidad sobre datos
Cada empresa es responsable de la veracidad, legalidad, respaldo funcional y autorizaciones necesarias sobre los datos que registra, importa, exporta o procesa dentro de la plataforma.

6. Seguridad y credenciales
Las claves, tokens, certificados, medios de pago y demas credenciales sensibles deben mantenerse bajo control del cliente y del proveedor segun su ambito. El usuario debe reportar accesos no autorizados o incidentes de seguridad tan pronto como los detecte.

7. Licencias, planes y pagos
El acceso a modulos, limites operativos y funcionalidades depende de la licencia vigente. Los valores, vigencias, renovaciones y medios de pago aceptados se rigen por la configuracion comercial y por los comprobantes emitidos para cada transaccion.

8. Disponibilidad y soporte
El proveedor procurara mantener la operacion del servicio y atender incidentes razonables de soporte, pero no garantiza continuidad absoluta ni ausencia total de interrupciones derivadas de mantenimiento, conectividad, terceros o fuerza mayor.

9. Integraciones de terceros
Las integraciones con Google, SMTP, pasarelas de pago, servicios tributarios o terceros externos dependen de credenciales validas, politicas de esos proveedores y disponibilidad de sus APIs. El proveedor no responde por cambios unilaterales de terceros ajenos a la plataforma.

10. Facturacion, cumplimiento y obligaciones regulatorias
Cada empresa es responsable de validar sus obligaciones tributarias, laborales, contables, de habeas data y de cualquier otra regulacion aplicable a su actividad. La plataforma es una herramienta de apoyo operativo y no reemplaza asesoria legal, contable o tributaria especializada.

11. Backups y continuidad operativa
La plataforma puede ofrecer funciones de respaldo y exportacion, pero cada empresa debe definir y ejecutar su propia politica de respaldo, validacion y conservacion documental segun su riesgo operativo y regulatorio.

12. Propiedad intelectual
El software, interfaces, codigo, diseno, flujos y componentes propios de Powerful Control System son titularidad del proveedor o de sus licenciantes. Este contrato no transfiere propiedad intelectual distinta al derecho limitado de uso conforme a la licencia vigente.

13. Suspensiones y terminacion
El proveedor podra suspender accesos ante incumplimientos graves, uso abusivo, riesgo de seguridad, mora o requerimientos legales. El cliente podra terminar el uso conforme a sus obligaciones pendientes y a los procedimientos de cierre aplicables.

14. Limitacion de responsabilidad
Salvo disposicion legal imperativa en contrario, el proveedor no respondera por lucro cesante, dano indirecto, perdida de oportunidad o fallos originados por uso indebido, datos incorrectos, terceros, conectividad, energia o decisiones operativas del cliente.

15. Notificaciones
Las notificaciones operativas, de seguridad o contractuales podran realizarse a traves del correo registrado, avisos en la plataforma o canales de contacto definidos por el proveedor.

16. Actualizaciones del contrato
El proveedor puede publicar nuevas versiones de este contrato por cambios legales, operativos, de seguridad o de alcance del servicio. Cuando exista una actualizacion material, la cuenta administrativa debera aceptar la version vigente para continuar el acceso administrativo.`),
		NotaAceptacion: "Al marcar la casilla y continuar, declaras que leiste y aceptas la version vigente de este contrato con facultad suficiente para actuar en nombre propio o de la empresa que administras.",
		ResumenCambio:  "Version base inicial del contrato operativo.",
		UsuarioCreador: "sistema",
		Estado:         "activo",
	}
}

func normalizeSuperContractText(raw, fallback string) string {
	value := strings.ReplaceAll(raw, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	value = strings.TrimSpace(value)
	if value == "" {
		return strings.TrimSpace(fallback)
	}
	return value
}

func normalizeSuperContractVersion(doc SuperContractVersion) SuperContractVersion {
	defaults := defaultSuperContractVersion()
	doc.Titulo = normalizeSuperContractText(doc.Titulo, defaults.Titulo)
	doc.Resumen = normalizeSuperContractText(doc.Resumen, defaults.Resumen)
	doc.Contenido = normalizeSuperContractText(doc.Contenido, defaults.Contenido)
	doc.NotaAceptacion = normalizeSuperContractText(doc.NotaAceptacion, defaults.NotaAceptacion)
	doc.ResumenCambio = normalizeSuperContractText(doc.ResumenCambio, "Actualizacion del contrato operativo.")
	doc.UsuarioCreador = normalizeSuperContractText(doc.UsuarioCreador, defaults.UsuarioCreador)
	doc.Estado = normalizeSuperContractText(doc.Estado, defaults.Estado)
	doc.Observaciones = normalizeSuperContractText(doc.Observaciones, "")
	if doc.Version <= 0 {
		doc.Version = defaults.Version
	}
	return doc
}

func superContractEquivalent(current, incoming SuperContractVersion) bool {
	return strings.TrimSpace(current.Titulo) == strings.TrimSpace(incoming.Titulo) &&
		strings.TrimSpace(current.Resumen) == strings.TrimSpace(incoming.Resumen) &&
		strings.TrimSpace(current.Contenido) == strings.TrimSpace(incoming.Contenido) &&
		strings.TrimSpace(current.NotaAceptacion) == strings.TrimSpace(incoming.NotaAceptacion)
}

func EnsureSuperContractSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is required")
	}

	if isPostgresDialect() {
		if _, err := execSQLCompat(dbConn, `CREATE TABLE IF NOT EXISTS super_contrato_versiones (
			id BIGSERIAL PRIMARY KEY,
			version INTEGER NOT NULL UNIQUE,
			titulo TEXT NOT NULL,
			resumen TEXT,
			contenido TEXT NOT NULL,
			nota_aceptacion TEXT,
			resumen_cambio TEXT,
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		)`); err != nil {
			return err
		}
	} else {
		if _, err := execSQLCompat(dbConn, `CREATE TABLE IF NOT EXISTS super_contrato_versiones (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			version INTEGER NOT NULL UNIQUE,
			titulo TEXT NOT NULL,
			resumen TEXT,
			contenido TEXT NOT NULL,
			nota_aceptacion TEXT,
			resumen_cambio TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		)`); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "super_contrato_versiones", "version", "INTEGER NOT NULL DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_contrato_versiones", "titulo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_contrato_versiones", "resumen", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_contrato_versiones", "contenido", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_contrato_versiones", "nota_aceptacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_contrato_versiones", "resumen_cambio", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_contrato_versiones", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_contrato_versiones", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_contrato_versiones", "observaciones", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_contrato_versiones", "fecha_creacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "super_contrato_versiones", "fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}

	if _, err := execSQLCompat(dbConn, "CREATE UNIQUE INDEX IF NOT EXISTS ux_super_contrato_versiones_version ON super_contrato_versiones(version)"); err != nil {
		return err
	}

	hasAdmins, err := tableExists(dbConn, "administradores")
	if err != nil {
		return err
	}
	if hasAdmins {
		if err := ensureColumnIfMissing(dbConn, "administradores", "acepta_contrato", "INTEGER DEFAULT 0"); err != nil {
			return err
		}
		if err := ensureColumnIfMissing(dbConn, "administradores", "contrato_version_aceptada", "INTEGER DEFAULT 0"); err != nil {
			return err
		}
		if err := ensureColumnIfMissing(dbConn, "administradores", "fecha_acepta_contrato", "TEXT"); err != nil {
			return err
		}
	}

	return nil
}

func EnsureDefaultSuperContract(dbConn *sql.DB) error {
	if err := EnsureSuperContractSchema(dbConn); err != nil {
		return err
	}

	var total int
	if err := queryRowSQLCompat(dbConn, "SELECT COUNT(1) FROM super_contrato_versiones").Scan(&total); err != nil {
		return err
	}
	if total > 0 {
		return nil
	}

	defaults := normalizeSuperContractVersion(defaultSuperContractVersion())
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, "INSERT INTO super_contrato_versiones (version, titulo, resumen, contenido, nota_aceptacion, resumen_cambio, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones) VALUES (?, ?, ?, ?, ?, ?, "+nowExpr+", "+nowExpr+", ?, ?, ?)", defaults.Version, defaults.Titulo, defaults.Resumen, defaults.Contenido, defaults.NotaAceptacion, defaults.ResumenCambio, defaults.UsuarioCreador, defaults.Estado, defaults.Observaciones)
	return err
}

func scanSuperContractVersion(scanner interface{ Scan(dest ...interface{}) error }) (*SuperContractVersion, error) {
	var item SuperContractVersion
	var resumen sql.NullString
	var nota sql.NullString
	var cambio sql.NullString
	var fechaCre sql.NullString
	var fechaAct sql.NullString
	var creador sql.NullString
	var estado sql.NullString
	var observaciones sql.NullString
	if err := scanner.Scan(&item.ID, &item.Version, &item.Titulo, &resumen, &item.Contenido, &nota, &cambio, &fechaCre, &fechaAct, &creador, &estado, &observaciones); err != nil {
		return nil, err
	}
	if resumen.Valid {
		item.Resumen = resumen.String
	}
	if nota.Valid {
		item.NotaAceptacion = nota.String
	}
	if cambio.Valid {
		item.ResumenCambio = cambio.String
	}
	if fechaCre.Valid {
		item.FechaCreacion = fechaCre.String
	}
	if fechaAct.Valid {
		item.FechaActualizacion = fechaAct.String
	}
	if creador.Valid {
		item.UsuarioCreador = creador.String
	}
	if estado.Valid {
		item.Estado = estado.String
	}
	if observaciones.Valid {
		item.Observaciones = observaciones.String
	}
	return &item, nil
}

func GetCurrentSuperContract(dbConn *sql.DB) (*SuperContractVersion, error) {
	if err := EnsureDefaultSuperContract(dbConn); err != nil {
		return nil, err
	}
	row := queryRowSQLCompat(dbConn, "SELECT id, version, titulo, resumen, contenido, nota_aceptacion, resumen_cambio, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones FROM super_contrato_versiones ORDER BY version DESC LIMIT 1")
	return scanSuperContractVersion(row)
}

func GetSuperContractVersionByNumber(dbConn *sql.DB, version int) (*SuperContractVersion, error) {
	if err := EnsureDefaultSuperContract(dbConn); err != nil {
		return nil, err
	}
	row := queryRowSQLCompat(dbConn, "SELECT id, version, titulo, resumen, contenido, nota_aceptacion, resumen_cambio, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones FROM super_contrato_versiones WHERE version = ? LIMIT 1", version)
	return scanSuperContractVersion(row)
}

func ListSuperContractVersions(dbConn *sql.DB, limit int) ([]SuperContractVersion, error) {
	if err := EnsureDefaultSuperContract(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	rows, err := querySQLCompat(dbConn, "SELECT id, version, titulo, resumen, contenido, nota_aceptacion, resumen_cambio, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones FROM super_contrato_versiones ORDER BY version DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]SuperContractVersion, 0)
	for rows.Next() {
		item, err := scanSuperContractVersion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, nil
}

func SaveSuperContractVersion(dbConn *sql.DB, doc SuperContractVersion) (*SuperContractVersion, bool, error) {
	if err := EnsureSuperContractSchema(dbConn); err != nil {
		return nil, false, err
	}

	current, err := GetCurrentSuperContract(dbConn)
	if err != nil {
		return nil, false, err
	}

	normalized := normalizeSuperContractVersion(doc)
	if superContractEquivalent(*current, normalized) {
		return current, true, nil
	}

	normalized.Version = current.Version + 1
	nowExpr := sqlNowExpr()
	_, err = execSQLCompat(dbConn, "INSERT INTO super_contrato_versiones (version, titulo, resumen, contenido, nota_aceptacion, resumen_cambio, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones) VALUES (?, ?, ?, ?, ?, ?, "+nowExpr+", "+nowExpr+", ?, ?, ?)", normalized.Version, normalized.Titulo, normalized.Resumen, normalized.Contenido, normalized.NotaAceptacion, normalized.ResumenCambio, normalized.UsuarioCreador, normalized.Estado, normalized.Observaciones)
	if err != nil {
		return nil, false, err
	}

	saved, err := GetCurrentSuperContract(dbConn)
	if err != nil {
		return nil, false, err
	}
	return saved, false, nil
}

func GetAdministradorContratoAceptacion(dbConn *sql.DB, email string) (SuperContractAcceptance, error) {
	acceptance := SuperContractAcceptance{}
	if err := EnsureSuperContractSchema(dbConn); err != nil {
		return acceptance, err
	}

	row := queryRowSQLCompat(dbConn, "SELECT COALESCE(acepta_contrato, 0), COALESCE(contrato_version_aceptada, 0), COALESCE(fecha_acepta_contrato, '') FROM administradores WHERE LOWER(COALESCE(email,'')) = LOWER(?) LIMIT 1", strings.TrimSpace(email))
	var acepta sql.NullInt64
	var version sql.NullInt64
	var fecha sql.NullString
	if err := row.Scan(&acepta, &version, &fecha); err != nil {
		if err == sql.ErrNoRows {
			return acceptance, nil
		}
		return acceptance, err
	}
	acceptance.Acepta = acepta.Valid && acepta.Int64 == 1
	if version.Valid {
		acceptance.Version = int(version.Int64)
	}
	if fecha.Valid {
		acceptance.Fecha = strings.TrimSpace(fecha.String)
	}
	return acceptance, nil
}

func SetAdministradorContratoAceptado(dbConn *sql.DB, email string, version int) error {
	if err := EnsureSuperContractSchema(dbConn); err != nil {
		return err
	}
	if version < 0 {
		version = 0
	}
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, "UPDATE administradores SET acepta_contrato = 1, contrato_version_aceptada = ?, fecha_acepta_contrato = "+nowExpr+", fecha_actualizacion = "+nowExpr+" WHERE LOWER(COALESCE(email,'')) = LOWER(?)", version, strings.TrimSpace(email))
	return err
}