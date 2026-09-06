# Revisión semántica y cierre del lote documental

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-06.

## Resultado y alcance

Se resolvió la disposición de las 138 referencias pendientes inventariadas al
iniciar la continuación (el conteo inicial anterior era 137; apareció un informe
adicional). La [matriz por archivo](matriz_revision_semantica_2026-09-06.json)
conserva decisión, observaciones, fuentes y hash normalizado al corte.

- revisado_con_limites: 108.
- contrato_maquina_con_guia: 1.
- historico_con_sucesor: 28.
- evidencia_fechada: 1.

Revisar no significa aprobar todo el código: se contrastaron contratos y puntos
de implementación, se corrigieron afirmaciones concretas y se delimitaron las
capacidades. Los planes, resultados antiguos y textos acumulados se clasificaron
por propósito/fecha y se preservaron como antecedentes; no se verificó cada frase
histórica como comportamiento actual. Las migraciones funcionales, proveedores,
hardware y producción no se ejecutaron en esta revisión documental.

## Cambios principales

- Las entradas de proyecto, módulos, código, BD, permisos y flujos ahora remiten
  a fuentes temáticas; el detalle anterior queda en `documentos/historico/2026-09-05/`.
- Se revisaron guías funcionales, contratos, runbooks, API, perfiles de agentes
  y planes. Los perfiles ya no exigen delegación automática.
- Checkout de licencias describe el fallback Epayco `classic_js` comprobado
  en handler e interfaz, en lugar del formulario legacy. Se separa de venta pública.
- El contrato fiscal usa una matriz por familia en lugar de cronologías
  incompatibles; el runbook sigue fuente genuina, permiso y acuse individual.
- Se corrigieron catálogo de trece plantillas, modelo global del chat,
  alcance real de cobranza, periodos fiscales, MRP y servicios externos.
- Instalación/Docker/Nginx se consolidan en documentos; se retiraron recetas
  con datos personales/folios reales, puertos SSH contradictorios, un Compose
  Nextcloud inexistente y el supuesto adaptador `object`.
- Se mantiene el contrato móvil YAML como manual, distinto del OpenAPI generado.
  Los manifiestos/visores técnicos existentes conservan sus rutas.

## Brechas concretas pendientes de ingeniería o aceptación

| ID | Hallazgo contrastado | Fuente / criterio de cierre |
| --- | --- | --- |
| SEM-01 | Soportes IA busca GPT-5.5 dentro del catálogo global; otro modelo publicado puede bloquear extracción | [Guía](../soportes_compras_ia.md): unificar política de modelo y probar configuración/caso negativo sin invocar proveedor real |
| SEM-02 | Cierre fiscal tiene validación expuesta, sin conexión automática a todas las mutaciones | [Guía](../cierre_fiscal.md): definir operaciones protegidas y probar cada rechazo antes de afirmar bloqueo transversal |
| SEM-03 | AIU registra documento local emitida; no acredita fuente/transporte/aceptación DIAN | [Guía](../aiu_construccion.md): reconciliar con flujo fiscal desde fuente genuina y estados explícitos |
| SEM-04 | MRP estima necesidades con stock de planificación; no representa disponibilidad real/Kardex por ese cálculo | [Guía](../produccion_mrp.md): integrar/contrastar stock y probar concurrencia antes de usarlo como disponibilidad |
| SEM-05 | Conectores bancarios/API son sondas; timestamps de sync no prueban conciliación y algunos GET persisten | [Contrato](contratos/contrato_integraciones_bancarias_y_conectores_externos.md): separar lectura/acción y documentar adaptador de negocio real |
| SEM-06 | Versionado documental puede dejar warning al historizar; firma declarada no es verificación criptográfica | [Contrato](contratos/contrato_repositorio_documental_y_firmas_externas.md): probar atomicidad/versiones y definir evidencia de firma requerida |
| SEM-07 | OpenAPI móvil mantiene varias operaciones con schemas resumidos o ausentes | [Guía](../api/mobile_api_v1.md): completar request/response y pruebas de compatibilidad antes de generar cliente estricto |
| SEM-08 | No todos los consumidores IA adoptan el modelo global; documentos dinámicos tiene modelo propio | [Contrato](contratos/contrato_documentos_dinamicos_ia_exportacion.md): decidir política explícita y probar consumidores |
| SEM-09 | Stack Nextcloud heredado y servicios opcionales requieren inventario/restore propios | [Guía](../nextcloud_empresarial.md): verificar runtime real y migración autorizada; no inferir retiro por un archivo de decommission |
| SEM-10 | Voz sin token permite acceso y voice no selecciona modelo | [Servicio](../../services/voice_stream_server/README.md): configurar red/autenticación y definir selección antes de publicarlo |
| SEM-11 | Plantillas sectoriales, KYC y catálogos de dispositivos no prueban cumplimiento ni interoperabilidad completa | [Módulos](../modulos_empresariales_colombia.md), [solar](../energia_solar.md): casos reales autorizados por capacidad/adaptador |
| SEM-12 | Correcciones de seguridad locales no están acreditadas en el entorno publicado | [Cierre local](../cierre_reparaciones_produccion_seguridad_2026-09-05.md), [QA publicada](../evidencia_qa_autenticada_pcs_2026-09-06.md): candidato exacto y aceptación de staging antes de promover |

Además continúan DOC-03/05/06/07/09/10 del [registro de riesgos](riesgos_y_brechas.md):
trazabilidad exhaustiva de requisitos/pruebas, capacidad y recuperación medidas,
responsables nominales, privacidad/retención aprobadas, correspondencia de
inventarios y evaluación normativa completa. No se inventaron aprobaciones,
SLA firmados ni certificaciones ISO/ASVS.

## Verificación documental

Comandos desde la raíz: `node --test tools/docs_catalog.test.mjs`,
`node tools/docs_catalog.mjs --write`, `node tools/docs_catalog.mjs --check`
y `git -c core.safecrlf=false diff --check`.

Resultado del corte: seis pruebas del validador aprobadas; catálogo de 502
documentos, sin hallazgos ni referencias pendientes de clasificar; comprobación
de catálogo sin diferencias y revisión de espacios del diff sin errores. La
matriz conserva las 138 decisiones del lote original. Esto verifica
estructura/enlaces/catálogo; no sustituye las pruebas funcionales de los módulos.

## Mantenimiento

Actualizar la fuente del tema y su evidencia, evitando volver a acumular
cronologías en contextos/mapas. Usar el [portal](../README.md) como entrada.
Los hashes de esta matriz describen el corte de revisión; una edición posterior
legítima puede cambiarlos y debe dejar su trazabilidad habitual.
