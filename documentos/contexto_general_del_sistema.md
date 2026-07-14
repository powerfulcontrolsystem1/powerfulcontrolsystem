# Contexto general del sistema

Estado: vigente. Ultima actualizacion: 2026-07-13.

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
