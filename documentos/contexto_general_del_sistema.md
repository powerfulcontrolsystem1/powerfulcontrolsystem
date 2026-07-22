# Contexto general del sistema

Estado: vigente. Ultima actualizacion: 2026-07-16.

## Actualizacion 2026-07-21 - Plan 105 pendiente de aprobacion

- `documentos/plan_105.md` consolida la auditoria general y es la hoja de ruta
  vigente para cerrar produccion. Su veredicto inicial es `NO-GO`.
- El plan no esta autorizado para ejecucion hasta aprobacion expresa del usuario.
  Debe comenzar por fijar el candidato y el alcance, sin usar `rs` ni tocar
  proveedores o datos reales durante la planificacion.
- Los validadores locales aprobados no sustituyen migraciones efimeras,
  aislamiento A/B, restore, race/carga, staging ni pruebas reales de proveedor.
- `documentos/matriz_alcance_piloto_plan_105.md` es el borrador de alcance que
  debe aprobarse antes de habilitar trafico comercial; no representa una
  activacion automatica de modulos.

## Actualizacion 2026-07-18 - contrato explicito del migrador

- El proceso `migrate` en produccion exige declarar
  `PCS_RUNTIME_SCHEMA_BOOTSTRAP`; si falta, el arranque falla cerrado en vez
  de activar por defecto el bootstrap legado.
- El servicio `migrate` de Compose propaga esa variable de forma obligatoria.
  `api` y `worker` mantienen `PCS_RUNTIME_SCHEMA_BOOTSTRAP=0` fijo.
- Esto reduce el riesgo de una configuracion incompleta, pero no sustituye la
  extraccion de los 154 `Ensure*` ni el ensayo de staging requerido antes de
  retirar definitivamente el puente legado.
- `sync_to_vps` completa automaticamente esa variable operativa en el archivo
  privado de Compose del VPS cuando falta, conserva valores explicitos validos
  y rechaza valores desconocidos antes de reconstruir servicios.
- En modo Docker, ese bootstrap conserva los secretos exclusivamente en el VPS
  y se limita a variables operativas necesarias para Compose; no imprime ni
  reemplaza credenciales privadas.

## Actualizacion 2026-07-16 - Plan 104: endurecimiento previo verificable

- `documentos/plan_104.md` es la matriz vigente de gates para produccion
  general. Su avance interno no sustituye staging, restauracion comprobada ni
  pruebas autorizadas con proveedores.
- La API y worker no pueden reactivar bootstrap heredado en produccion; el
  migrador puede ensayar su omision de forma explicita. Varios flujos ERP y de
  documentos ahora verifican el esquema antes de atender solicitudes, sin DDL
  desde HTTP.
- El catalogo compatibilidad `Ensure*` queda congelado por manifiesto generado
  con huella por paso; cualquier ajuste futuro debe llegar como migracion nueva
  y no como cambio silencioso del bootstrap.
- `documentos/arquitectura/matriz_rutas_multiempresa.md` se genera desde los
  registros HTTP y debe mantenerse vigente; sus wrappers no sustituyen las
  pruebas negativas de tenant sobre datos, archivos, cache y jobs.
- `documentos/arquitectura/inventario_runtime_ensure.md` lista el DDL heredado
  invocado fuera de `pcs-migrate`. Se usa para priorizar su extraccion y no
  debe confundirse con una autorizacion para habilitar bootstrap en produccion.
- Los flujos de licencia, DIAN, carrito y corte de caja verifican sus tablas
  de facturacion, configuracion avanzada, ventas y finanzas sin ejecutar
  `Ensure*` desde HTTP. Despues del primer lote del Plan 105, el inventario
  vigente conserva 110 llamadas legacy fuera del migrador, 37 de ellas en
  trafico HTTP; no debe aumentar sin una migracion de extraccion revisada.
- CSP usa origenes cerrados configurables y los errores 4xx tecnicos se
  redaccionan. Aun se mantiene `unsafe-inline` como compatibilidad transitoria
  hasta completar inventario y pruebas visuales de scripts.

## Actualizacion 2026-07-16 - Plan 103: cierre operativo verificable

- La recoleccion de metricas deja de vivir en la API y se agenda como trabajo
  durable del worker; su tabla queda versionada por `pcs-migrate`.
- `documentos/plan_103_cierre_produccion.md` concilia los hallazgos que ya
  estaban resueltos con los gates externos que siguen siendo obligatorios.
  Debe consultarse antes de declarar una liberacion general o replicas.

## Actualizacion 2026-07-16 - Plan 102: cierre tecnico controlado

- API, worker y migrador quedan separados por rol: solo `pcs-migrate` puede
  aplicar DDL de plataforma y el guard de runtime bloquea mutaciones de esquema
  en API/worker de produccion. El catalogo heredado se registra una vez en el
  ledger, con checksum y lock, antes de desactivar el bootstrap historico.
- Los cron de negocio salen de `main.go` y se programan como jobs durables del
  worker; `commerce.sale-paid` se publica en la outbox dentro de la transaccion
  de cobro. La caja usa bloqueo transaccional y las rutas empresariales nuevas
  reciben un `TenantContext` validado, nunca un tenant confiado desde el cliente.
- `/ready` verifica ambas bases, migraciones y almacenamiento privado. Con mas
  de una API, produccion exige storage declarado `shared` u `object`; no se
  permite escalar horizontalmente usando archivos locales por accidente.
- La liberacion reproducible usa imagenes separadas API/worker/migrador y el
  override `docker-compose.release.yml` exige digests. Aun faltan evidencia de
  staging, restauracion ensayada, Object Storage externo y pruebas autorizadas
  de proveedores antes de declarar produccion general lista.

## Actualizacion 2026-07-16 - Plan final de produccion

- `documentos/plan_final_para_produccion.md` es la hoja de ruta vigente para
  convertir la fundacion actual de migrador, worker, cola, outbox y API movil
  en una plataforma operable con replicas. Debe consultarse antes de proponer
  cambios transversales de arquitectura, despliegue, PostgreSQL, jobs o API
  movil.
- No confundir los binarios ya creados con una separacion terminada: la API aun
  conserva bootstrap historico y timers; el worker y la outbox requieren los
  handlers y dispatcher definidos en el plan. Produccion general no queda
  autorizada hasta cumplir los hitos y gates de staging del documento.

## Actualizacion 2026-07-16 - Migrador y trabajo durable

- `pcs-migrate` usa un ledger con checksum, corrida auditable y advisory lock
  transaccional. Solo las migraciones catalogadas pueden modificar el esquema
  nuevo de plataforma; una diferencia de checksum bloquea el arranque de
  migracion.
- La cola, outbox e idempotencia movil de plataforma son verificadas en API y
  worker sin emitir DDL. El bootstrap heredado permanece activo mientras se
  traslada su inventario clasificado: no establecer
  `PCS_RUNTIME_SCHEMA_BOOTSTRAP=0` sin ensayo de staging documentado.
- El worker tiene leases, recuperacion de tareas vencidas y registro tipado;
  los handlers de correo, DIAN, pagos, documentos y reportes siguen como
  trabajo pendiente y no deben declararse migrados por esta base tecnica.
- El healthcheck del worker es interno al contenedor (`127.0.0.1:8082` por
  defecto), sin ruta publica. Docker comprueba `/ready`, que exige ultimo batch
  correcto y PostgreSQL disponible.

## Actualizacion 2026-07-13 - Flujos moviles POS v1

- `/api/v1/` ya cubre carrito, items, cobro, sincronizacion offline,
  documentos fiscales y buzon/notificaciones, ademas de identidad, productos y
  clientes. Las escrituras moviles usan `Idempotency-Key` persistente por
  empresa y no duplican pagos, documentos ni mensajes cuando la red reintenta.
- La capa v1 reutiliza los handlers operativos existentes; no crea una segunda
  regla de caja, inventario, impuestos o DIAN. Ver
  `documentos/api/mobile_api_v1.md` y su contrato OpenAPI.

## Actualizacion 2026-07-13 - Plan 101 de arquitectura

- PCS mantiene un monolito modular en Go y PostgreSQL: los limites se aplican
  por modulo, repositorio y wrapper multiempresa sin duplicar reglas de caja,
  inventario, impuestos o DIAN. La guia y los gates de preproduccion estan en
  `documentos/plan_101_arquitectura_modular.md`.

## Actualizacion 2026-07-13 - IA propia y base movil

- El proveedor OpenAI propio es opcional por empresa, cifrado por tenant y sin
  exposicion de la clave. Solo permite superar la cuota PCS para OpenAI; la
  seguridad, permisos, auditoria y los limites de OpenAI permanecen activos.
- La base movil versionada se publica bajo `/api/v1/` con JSON uniforme. Su
  contrato y plan de migracion estan en `documentos/api/mobile_api_v1.md`.

## Regla obligatoria para agentes

Todo agente que reciba una consulta sobre este repositorio debe leer primero
este documento. Debe abrir despues
`documentos/contexto_especifico_del_sistema.md` y seguir sus enlaces solo para
el area solicitada. Esta secuencia reduce contexto innecesario y no sustituye
las revisiones de seguridad, datos, permisos o despliegue exigidas por el tema.

## Que es PCS

Powerful Control System (PCS) es una plataforma SaaS POS multiempresa para
empresas colombianas y otros paises configurables. Integra ventas POS, carritos,
inventario, compras, clientes, finanzas, nomina, reportes, facturacion
electronica, licencias, correo corporativo, canales digitales, automatizacion e
IA empresarial.

La empresa interna `Powerful Control System` participa como una empresa normal
del sistema para vender licencias y emitir sus documentos, pero no reemplaza el
rol reservado de super administrador.

## Arquitectura y tecnologias

- Backend: Go y libreria estandar; no agregar dependencias ni modificar
  `go.mod` sin autorizacion expresa.
- Datos: PostgreSQL es el unico motor permitido.
- Frontend: HTML, CSS y JavaScript estatico; no migrar a frameworks o bundlers
  sin autorizacion.
- Runtime: Docker en VPS principal. El Nextcloud empresarial es un servicio del
  VPS principal y se separa del VPS2, que permanece auxiliar e independiente.
- Despliegue: el flujo canonico es `scripts/rs.ps1`; consultar antes
  `documentos/comandos_codex.md`.

## Reglas no negociables

1. Todo dato, consulta, mutacion, archivo, permiso y auditoria empresarial se
   aisla por `empresa_id` y se valida en backend.
2. Los menus y controles visuales no son seguridad. Cada endpoint aplica sesion,
   rol, permiso y licencia cuando corresponda.
3. No imprimir ni versionar secretos, contrasenas, tokens, certificados, DSN o
   datos privados.
4. No destruir datos, empresas, licencias, usuarios, configuraciones o backups
   sin instruccion explicita, alcance claro y trazabilidad.
5. Las acciones criticas deben ser idempotentes y auditables: pagos, ventas,
   documentos, caja, licencias, usuarios, archivos y configuraciones.
6. Mantener UI real para PC y celular, UTF-8 correcto, controles accesibles e
   impresion independiente de la apariencia clara u oscura.
7. No existe modulo de juegos ni emuladores en PCS. No reintroducir rutas,
   assets, datos, contenedores ni menues de entretenimiento.

## Operacion y roles

- `super_administrador`: gobierno global, configuraciones e infraestructura.
- `admin_empresa` y administradores compartidos: administracion de la empresa
  autorizada.
- Roles operativos, como `cajero`, reciben solo los permisos de su matriz.
- La cuenta reservada de super administrador se debe preservar; no debe tratarse
  como usuario operativo de una empresa.

## Flujos criticos

- Venta y carrito: una venta puede ser sola o emitir factura electronica segun
  configuracion y permisos. Inventario, pagos, caja, documento e impresion deben
  conservar consistencia.
- Facturacion DIAN Colombia: PCS usa modalidad `Software propio`; cada empresa
  conserva NIT, firma, credenciales y trazabilidad propia. Una respuesta de
  recepcion no equivale a aceptacion final.
- Pagos y licencias: Epayco y Wompi se procesan idempotentemente. Las licencias
  limitan documentos emitidos, no el numero de cajas.
- Correo y WhatsApp: canales configurables y auditables, con secretos cifrados
  o referencias seguras. Mailu se administra mediante API interna autenticada,
  no mediante socket Docker del backend. No prometer entrega real sin pruebas
  del proveedor y DNS.
- IA y agentes: se limitan por empresa, respetan roles, lista cerrada de
  herramientas, propuestas temporales e idempotencia. Las escrituras requieren
  confirmacion independiente; ningun modelo recibe capacidad directa de SQL,
  HTTP arbitrario ni seleccion de empresa.

## Forma de trabajar

1. Identificar modulo, rutas, tablas, permisos, empresa y riesgo.
2. Abrir desde el contexto especifico solo la documentacion requerida.
3. Aplicar cambios acotados, sin duplicar reglas ni crear persistencia local
   cuando existe backend.
4. Probar en proporcion al riesgo: sintaxis, tests Go, endpoints, responsive y
   flujo real cuando el usuario lo solicite.
5. Actualizar documentacion, trazabilidad y ejecutar `rs` solo cuando se haya
   solicitado o sea parte del cierre aprobado.

## Indice de ampliacion

Abrir `documentos/contexto_especifico_del_sistema.md`. Ese documento enlaza los
mapas de modulos, flujos, datos, seguridad, comandos, decisiones, integraciones
y runbooks vigentes.
