# Reporte QA de modulos empresariales 2026-05-06

## Alcance

Revision transversal de los modulos recientes y de los accesos principales del sistema para validar permisos, paginas, botones/enlaces estaticos, APIs empresariales y documentacion. La prueba autenticada se hizo sobre Motel Calipso (`empresa_id=7`) con la cuenta super administradora indicada por el propietario del sistema, sin registrar la clave en la documentacion.

## Correcciones aplicadas

- `web/js/administrar_empresa.js`: la resolucion local de permisos ahora reconoce `administrador_total` con acceso total igual que `super_administrador`, alineando el menu empresarial con las reglas backend.
- `web/index.html`: la portada publica actualiza la descripcion de modulos para incluir Cobranza, Portal contador, Captura IA/OCR de compras y gastos, AIU construccion, Parqueaderos con ticket QR y Apartamentos turisticos en la seccion fija y tarjetas fallback.
- `backend/db/cobranza.go`: el dashboard de Gestion de cobranza valida esquema una sola vez por peticion y reutiliza consultas internas, evitando validaciones repetidas.
- `backend/db/soportes_compras_ia.go`: el dashboard de Captura inteligente de compras/gastos evita una segunda validacion de esquema al listar soportes recientes.
- `backend/db/portal_contador.go`: el dashboard de Portal contador evita validaciones repetidas de esquema en clientes, obligaciones, solicitudes y comunicaciones.
- `web/administrar_empresa/soportes_compras_ia.html`: los enlaces dinamicos de archivos radicados se validan antes de renderizarse para impedir protocolos no seguros.

## Pruebas automaticas

- `go test ./... -count=1` en `backend`: aprobado.
- `git diff --check`: aprobado; solo se reportaron advertencias normales de final de linea CRLF/LF.
- Auditoria estatica de enlaces HTML: sin enlaces privados rotos relevantes. Los hallazgos restantes son rutas backend esperadas (`/auth/google/login`, `/auth/logout`), la ruta publica legacy `/emulador/` y un enlace dinamico de archivo ya endurecido.
- Auditoria estatica de botones `onclick`: sin funciones inexistentes detectadas en las paginas HTML revisadas.
- Auditoria estatica de IDs en `web/administrar_empresa/*.html`: sin duplicados literales relevantes.

## Pruebas autenticadas con Motel Calipso

Sesion super administradora:

- `POST /super/api/administradores/login`: 200.
- `GET /api/empresa/permisos_contexto?empresa_id=7&include_matrix=1`: 200, rol efectivo `super_administrador`.

Paginas validadas con HTTP 200:

- `/super_administrador.html`
- `/administrar_empresa.html?id=7&empresa_id=7`
- `/administrar_empresa/compras_menu.html?empresa_id=7`
- `/administrar_empresa/soportes_compras_ia.html?empresa_id=7`
- `/administrar_empresa/cobranza.html?empresa_id=7`
- `/administrar_empresa/portal_contador.html?empresa_id=7`
- `/administrar_empresa/finanzas_menu.html?empresa_id=7`
- `/administrar_empresa/configuracion.html?empresa_id=7`
- `/administrar_empresa/taxi_system.html?empresa_id=7`
- `/administrar_empresa/domicilios.html?empresa_id=7`
- `/administrar_empresa/consultorio_odontologico.html?empresa_id=7`
- `/administrar_empresa/parqueadero.html?empresa_id=7`
- `/administrar_empresa/apartamentos_turisticos.html?empresa_id=7`
- `/administrar_empresa/aiu_construccion.html?empresa_id=7`

APIs validadas con HTTP 200:

- `/api/empresa/soportes_compras_ia?empresa_id=7&action=dashboard`
- `/api/empresa/cobranza?empresa_id=7&action=dashboard`
- `/api/empresa/portal_contador?empresa_id=7&action=dashboard`
- `/api/empresa/aiu_construccion?empresa_id=7&action=dashboard`
- `/api/empresa/parqueadero?empresa_id=7&action=dashboard`
- `/api/empresa/apartamentos_turisticos?empresa_id=7&action=dashboard`
- `/api/empresa/domicilios?empresa_id=7&action=dashboard`
- `/api/empresa/taxi_system?empresa_id=7&action=dashboard`
- `/api/empresa/odontologia?empresa_id=7&action=dashboard`
- `/api/empresa/configuracion_avanzada?empresa_id=7`
- `/api/empresa/impresoras?empresa_id=7`

## Observaciones operativas

- La cuenta indicada autentica correctamente como super administradora. El login operativo de usuario interno de empresa para `empresa_id=7` respondio 401, por lo que las pruebas se ejecutaron con sesion super administradora y permisos empresariales efectivos.
- La prueba interactiva de clicks con navegador embebido no pudo ejecutarse porque el runtime local de Node del plugin navegador devolvio `Acceso denegado`. Se sustituyo por validacion HTTP autenticada y auditorias estaticas de botones/enlaces.
- Algunas APIs siguen mostrando latencia de varios segundos por el costo de PostgreSQL remoto y validaciones de esquema/arranque. Los dashboards de los tres modulos nuevos fueron optimizados para evitar trabajo repetido dentro de una misma peticion.

## Estado final

Los modulos revisados quedan cargando con HTTP 200, con permisos empresariales coherentes, sin enlaces privados rotos relevantes y con pruebas Go completas aprobadas. Las mejoras pendientes recomendadas son de rendimiento transversal: cachear validaciones de esquema por proceso y separar migraciones de arranque de las peticiones normales.
