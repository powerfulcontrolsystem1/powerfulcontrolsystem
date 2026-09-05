# Auditoría de facturación electrónica: seguridad y separación fiscal

Fecha: 2026-09-05. Empresa de QA: PCS, empresa_id 12.
Resultado: **NO-GO para declarar todo el módulo 100 % listo en todos los países**.
Esta revisión entrega correcciones; no es una certificación legal ni un pentest
exhaustivo. No se ejecutaron nuevas emisiones, anulaciones, reenvíos o cambios
de configuración fiscal en producción.

## Hallazgos corregidos en el candidato

1. **Alta — herencia fiscal entre países.** Al no existir un perfil guardado,
   `hydrateFacturacionFromEmpresaConfig` copiaba NIT, resolución, prefijo y
   ambiente de la empresa sin comprobar el país. El QA autenticado mostró
   datos colombianos y Producción en las pantallas de Ecuador y Panamá.
   Ahora solo se heredan datos cuando el país fuente coincide con el destino.
   No se borran ni migran perfiles ya guardados: necesitan revisión explícita.
2. **Alta — aceptación fiscal simulada.** El despacho permitía proveedores
   manual/interno/local y `mock://ok` con resultado exitoso. Se rechazan en
   producción antes de leer secretos o realizar transporte.
3. **Alta — conector genérico sin contrato fiscal.** Un HTTP 2xx bastaba para
   marcar éxito y podía enviar datos fiscales a una URL genérica. La ruta de
   producción queda limitada al adaptador Colombia/DIAN explícito; otros
   proveedores/países fallan cerrados. La función HTTP auxiliar conserva
   pendiente, nunca aceptación, ante una entrega 2xx.
4. **Alta — reserva colombiana para otros países.** La reserva legal común
   utilizaba `empresa_configuracion_avanzada` y su consecutivo global. Se
   bloquean jurisdicciones distintas de CO antes de iniciar transacción.
   No se consume numeración para simular soporte internacional.
5. **Defensa adicional de aislamiento.** El último paso del transporte exige
   coincidencia de empresa y país entre configuración, documento y solicitud;
   documento y configuración en producción y configuración activa.
   Los códigos de país desconocidos se rechazan antes del fallback a CO.
6. **Claridad operativa.** Ecuador y Panamá muestran que son perfiles de datos,
   no adaptadores operativos. Sus checklists exponen
   `emision_habilitada=false`; completar campos no certifica ni activa emisión.

No hay dependencias nuevas, cambios en go.mod ni migraciones. Se mantienen
permisos y licencias existentes; no se abre ninguna familia documental.

## QA visual autenticado de la versión pública

Se inspeccionaron capturas renderizadas y estados accesibles del navegador.
Las pantallas públicas siguen mostrando la versión desplegada, no este parche.

| Superficie | Evidencia observada | Límite |
| --- | --- | --- |
| DIAN principal / RUT | Configuración cargada 100 %, producción, secretos enmascarados; botón de carga RUT y aplicación bloqueada sin archivo; certificado y resolución mostrados vigentes | No se subió ni envió el RUT a IA ni se guardó configuración |
| Nómina | Configuración separada pendiente; preflight ejecutado sin reservar consecutivos; estado pendiente_liquidaciones, documentos 0, neto 0 | No hay fuente mensual genuina para emitir |
| Documento soporte | Configuración separada inexistente, formulario estructurado, bandeja vacía y mensaje sin compra real | No se inventó compra, vendedor, rango ni autorización |
| Ecuador / Panamá | Formularios independientes, requisitos locales pendientes, herencia incorrecta de datos CO confirmada visualmente | Corrección local pendiente de desplegar y revalidar |
| Factura 1PCS8 | Visible en bandeja de agosto; diálogo bloquea vacío, solo frase, solo motivo y frase minúscula; habilita ANULAR + motivo válido; limpiar vuelve a bloquear | Se canceló sin confirmar anulación |
| Notas crédito | NC12000000113 visible y Anular deshabilitado; facturas ya anuladas también bloqueadas | No hubo nueva nota crédito |
| Artefactos 1PCS8 | Botón informó descarga de 3 archivos fiscales iniciada | No se acreditó contenido, firma, integridad ni representación de esas descargas |
| Centro DIAN | Producción local activa; factura libre, NC/ND libres, POS y RADIAN bloqueados; historial con aceptados y pendientes; diagnóstico oficial HTTP 200 | No se procesó cola ni se reconsultaron pendientes |

Los avisos nuevos EC/PA se visualizaron también en una vista local sin API;
esa comprobación acredita el texto renderizado, no el funcionamiento del
backend corregido contra datos productivos ni el diseño responsive completo.

No se acreditó en esta ejecución el rechazo real de accesos con un usuario de
privilegios mínimos o de otra empresa. La sesión administrativa de PCS no
demuestra por sí sola aislamiento multiempresa ni privacidad de nómina.

## Seguridad y escalabilidad: alcance y pendientes

Las regresiones existentes de permisos DIAN, aislamiento de fuentes/artefactos,
privacidad de nómina, secretos y salida HTTP se incluyen en las pruebas Go.
Se añadieron regresiones para cruces empresa/país/ambiente, estado inactivo,
simulación, URL cuyo texto contiene dian.gov.co sin ser adaptador DIAN,
país desconocido, reserva extranjera y capacidades de checklist.

No se realizaron pruebas destructivas, carga contra DIAN ni intentos de
explotación sobre empresas ajenas. Pendientes para cierre integral:

- Pruebas negativas autenticadas con roles mínimos: IDOR por empresa e ID
  secundario, permisos de nómina/artefactos, CSRF y rutas de descarga.
- PostgreSQL de integración: verificar concurrencia, numeración e idempotencia
  con base aislada. Sin DSN de pruebas, los casos opcionales pueden omitirse;
  un go test verde no acredita PostgreSQL real.
- Ensayo medido de cola multiempresa, latencias, equidad entre empresas,
  reintentos sin duplicados, recuperación tras caída, restauración de backup
  y conservación verificable de XML/acuse.
- Comprobar inventario de claves, rotación, controles del VPS, alertas,
  retención y borrado, accesos administrativos y respuesta a incidentes.
- Revalidar después del despliegue los perfiles extranjeros: no deben heredar
  datos CO. Revisar perfiles ya persistidos sin alterarlos automáticamente.

## Alcance documental y legal

El anexo técnico de factura electrónica publicado por DIAN es 1.9. La
Resolución Única 000227 de 2025 compila reglas de generación/transmisión y
validación. El criterio de aceptación debe ser el acuse oficial, no HTTP 200,
un TrackId o el estado saludable del servidor.
Fuentes oficiales consultadas:
[documentación DIAN](https://micrositios.dian.gov.co/sistema-de-facturacion-electronica/documentacion-tecnica/),
[Resolución 000227](https://www.dian.gov.co/normatividad/Paginas/Resolucion-000227-del-23092025.aspx).

La seguridad técnica no sustituye obligaciones de tratamiento de datos,
finalidad, confidencialidad y controles humanos/administrativos de la Ley 1581.
Queda pendiente contrastar las políticas y contratos operativos de PCS con el
responsable jurídico/contable:
[SIC — Ley 1581](https://sedeelectronica.sic.gov.co/transparencia/normativa/ley-estatutaria-1581-de-2012).

Panamá tiene modalidades, afiliación y especificaciones propias; la presencia
de un formulario no equivale a interoperabilidad:
[DGI — factura electrónica](https://dgi.mef.gob.pa/_7facturaelectronica/felectronica).
La documentación SRI fue consultada en la investigación; la última reapertura
tuvo timeout, por lo que no se declara validación normativa integral de Ecuador.

Según el contrato del repositorio, siguen fuera del cierre completo: nota
crédito parcial, nota débito, ajustes de nómina/soporte, equivalentes/POS,
contingencia y eventos RADIAN que no cuenten con adaptador/fuente específicos.
Nómina ordinaria y soporte tienen implementación, pero no se puede acreditar
su primera aceptación real sin fuentes, configuración y acuses genuinos.

## Entrega y compuertas

Candidato aislado en `D:\pcs-fiscal-security-20260905`,
rama `codex/fiscal-country-security-20260905`, basado en
`460e6122c8416eaa86323eb5a641253fe0294ccf`.
El árbol principal tenía trabajo ajeno; no se incluye en esta entrega.

Pruebas iniciales en el árbol principal: `go test ./... -count=1` y
`go vet ./...` finalizaron con código 0; sintaxis de ambos HTML correcta.
Validación definitiva del candidato aislado:

- `scripts/profesional_preflight.ps1 -Full -RequireMigrationAudit -MigrationBaseRef origin/main`: código 0, batería Go completa y git diff --check verdes.
- Informe local: `documentos/reportes_profesionales/preflight_20260905_154150.md`.
- `go vet ./...`: código 0.
- Primer intento: inventario Ensure desactualizado por cambio de línea;
  regeneración mecánica, sin cambiar migraciones históricas.
- Segundo intento: la prueba SSRF esperaba llegar al bloqueo de endpoint
  usando un proveedor ahora no implementado. Se ajustó el fixture a DIAN
  para seguir probando el mismo bloqueo y conservar la verificación de cero
  solicitudes al servidor local.
- **Omisiones reales:** Docker Compose se omitió por Docker no disponible;
  `PCS_TEST_POSTGRES_DSN` no está configurado. No son pruebas de runtime ni
  PostgreSQL completadas, aunque el orquestador concluya con código 0.

No confundir estos resultados con integración, merge o despliegue.

La publicación por PR requiere revisión independiente y checks del mismo SHA.
No se debe desplegar desde el árbol sucio ni saltar protecciones.
