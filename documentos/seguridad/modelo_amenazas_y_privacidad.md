# Modelo de amenazas y protección de datos

Estado: Vigente. Responsable: Seguridad e ingeniería. Revisión documental: 2026-09-05.

## Alcance y límites

Modelo inicial de amenazas de PCS elaborado desde arquitectura, reglas y auditoría local. No es un pentest, análisis legal ni declaración de controles efectivos. [SECURITY](../../SECURITY.md) define reporte privado; el [checklist multiempresa](../checklist_seguridad_endpoint_multiempresa.md) define revisión por operación.

Activos: sesiones y credenciales; datos comerciales/personales; caja y pagos; fuentes y certificados fiscales; archivos privados; configuraciones; jobs/outbox; backups y auditoría. Actores: usuario legítimo con permisos limitados, otro tenant, visitante, cliente/dispositivo comprometido, proveedor no confiable, operador con privilegios y contenido malicioso procesado por IA.

## Fronteras y escenarios

| ID | Frontera y amenaza | Control requerido | Verificación y límite actual |
| --- | --- | --- | --- |
| TH-01 | Navegador → API: suplantar sesión, rol o empresa | Sesión validada, permisos/licencia, tenant del servidor | PCS-REQ-001/002; probar objetos hijos, no solo wrapper |
| TH-02 | Handler → PostgreSQL: relacionar registro de otra empresa o alterar autor/estado | Consultas parametrizadas, ownership de IDs, campos permitidos y transacción | Auditoría general registra defectos de referencias; requieren cierre por módulo |
| TH-03 | Público → pagos: webhook falso, replay o monto alterado | Firma/autenticidad, referencia, moneda/importe y deduplicación persistente | Contrato checkout y pruebas PostgreSQL/proveedor aislado |
| TH-04 | API/worker → fiscal: país/familia incorrectos o aceptación simulada | Fuente sellada, adaptador específico, acuse oficial y reconciliación | Auditoría fiscal local; no acredita despliegue ni aceptación nueva |
| TH-05 | Archivo/URL → servidor: traversal, SSRF, malware o contenido activo | Autorización, límites, validación de destino y archivo, storage privado | Tests negativos y configuración real de escaneo/red aún deben verificarse |
| TH-06 | Prompt/adjunto → herramienta IA: ampliar permisos o extraer datos | Herramientas cerradas, minimización, tenant, confirmación e idempotencia | No confiar instrucciones dentro del contenido ni respuestas del modelo |
| TH-07 | API → colas: repetición, job ajeno o efecto perdido | Identidad de operación, tenant, lease, reintentos y auditoría | Prueba caída tras efecto y antes del ack |
| TH-08 | Operador/runtime → esquema/secretos: privilegio excesivo o exposición | Migrador separado, runtime mínimo, cifrado y redacción | Verificar privilegios reales; declarar variables no basta |
| TH-09 | Tenant → capacidad compartida: abuso, consumo de cuota o agotamiento | Límites por identidad validada, pool y carga controlados | Cierre local registra cuota durable tras autenticación; falta acreditar comportamiento del candidato bajo carga |
| TH-10 | Dispositivo → control físico: empresa falsa, comando repetido o estado ambiguo | Credenciales de túnel, pertenencia, reserva e identidad de comando | [Domótica](../domotica_raspberry_tunnel.md); confirmación física autorizada |
| TH-11 | Backup/exporte → destinatario: fuga o restauración inconsistente | Control de acceso, cifrado, retención, integridad y restore aislado | Ensayo medido y recuperación de claves; no copiar datos en informes |

## Clasificación y minimización

| Categoría | Uso y acceso esperado | Requisito documental |
| --- | --- | --- |
| Público | Catálogo publicado explícitamente | Separar de contacto/datos privados y revisar contenido comercial |
| Interno técnico | Arquitectura y procedimientos | Sin hosts privados, secretos ni material de clientes |
| Confidencial empresarial | Ventas, inventario, documentos y configuración | Empresa/rol, finalidad, exportación y retención aprobadas |
| Restringido personal/fiscal | Nómina, Vida, identidad, certificados y fuentes | Mínimo acceso por propósito/usuario, auditoría y recuperación controlada |
| Secreto | Claves, tokens, contraseñas y DSN | Fuera de Git/logs, referencia privada y procedimiento de rotación |

Un repositorio accesible a terceros no es un almacén de evidencia sensible. Usar datos ficticios de estructura válida, sin inventar hechos comerciales/fiscales reales. Guardar evidencia privada en el sistema autorizado y publicar solo su identificador/resumen minimizado.

## Registro de tratamiento por módulo

Antes de incorporar una categoría personal, documentar finalidad, categorías y origen, interesados, responsable/encargados, consumidores/proveedores, ubicación y transferencias, base aplicable validada, acceso, retención, solicitudes del titular, borrado/anonimización y backups. El [gobierno de datos](../arquitectura/gobierno_datos.md) define el contrato por entidad.

Falta completar el registro organizacional, plazos legales, acuerdos de proveedores y evaluación normativa aplicable con responsables competentes. No asumir que cifrado, una política escrita o aislamiento por empresa resuelven todos los requisitos de privacidad.

## Cambios, vulnerabilidades e incidentes

Revisar este modelo cuando cambien actores, datos, autenticación, exposición pública, proveedores, herramientas IA, dispositivos, storage o privilegios. Vincular cada riesgo al requisito y prueba de mitigación. Mantener vulnerabilidades por el proceso de SECURITY y [continuidad](../operacion/incidentes_y_continuidad.md); no silenciar un hallazgo con exclusiones genéricas.
