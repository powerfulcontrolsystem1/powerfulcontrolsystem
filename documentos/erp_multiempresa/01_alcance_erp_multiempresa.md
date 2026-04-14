# Documento de alcance ERP multiempresa

Fecha: 2026-04-14
Version: 1.0
Estado: listo para revision

## 1. Proposito

Definir el alcance oficial del proyecto ERP multiempresa, delimitando procesos empresariales cubiertos, reglas de negocio base, actores, integraciones y criterios de exito.

## 2. Contexto

El sistema opera como plataforma ERP/POS multiempresa con aislamiento estricto por `empresa_id`, gobierno de seguridad por rol y trazabilidad operativa/contable por documento, periodo y actor.

Fuentes consolidadas para este alcance:
- descripcion funcional general del proyecto
- arquitectura y diagramas de procesos
- matriz de roles y permisos
- matriz de aislamiento por endpoint
- matriz KPI operativos, financieros y contables
- estructura de base de datos y reglas de datos

## 3. Objetivos del alcance

1. Estandarizar el marco documental ERP del proyecto.
2. Unificar lenguaje de negocio entre producto, tecnologia y operacion.
3. Asegurar claridad en procesos empresariales end-to-end.
4. Formalizar reglas de negocio multiempresa y de interoperabilidad.
5. Definir base de evaluacion para implementacion, UAT y despliegue.

## 4. Alcance funcional incluido

### 4.1 Gobierno, seguridad y plataforma

- Gestion global de empresas, licencias, tipos de empresa y administradores.
- Autenticacion administrativa con OAuth Google y aceptacion contractual.
- Seguridad por rol/modulo/accion con control de alcance por empresa.
- Auditoria empresarial de acciones criticas con retencion y exportacion.
- Backups empresariales con restauracion trazable.

### 4.2 Proceso comercial (Lead-to-Cash)

- Gestion de clientes y segmentacion comercial.
- Cotizaciones, pedidos y conversion a documento final.
- Venta POS/carritos con cierre de pago por metodo habilitado.
- Venta publica por empresa con catalogo y pagos en pasarela.
- Facturacion electronica general y flujo DIAN para Colombia.

### 4.3 Proceso abastecimiento (Procure-to-Pay)

- Gestion de proveedores y condiciones comerciales.
- Ordenes de compra, recepcion parcial/total y contabilizacion.
- Validacion documental del ciclo de compras.

### 4.4 Proceso inventario y operaciones de piso

- Productos, servicios, categorias y bodegas.
- Existencias, movimientos, ajustes, transferencias y kardex.
- Alertas de quiebre, proyeccion de riesgo y plan de reposicion.
- Soporte de estaciones operativas (mesa/habitacion/punto de atencion).
- Multiples carritos/sesiones concurrentes por estacion y empresa.

### 4.5 Proceso financiero-contable (Record-to-Report)

- Movimientos financieros, periodos y cierres de caja.
- Eventos contables y asientos canonicos por procesamiento por lotes.
- Conciliacion contable y conciliacion bancaria.
- Reportes directivos y operativos multiformato.

### 4.6 Proceso cartera y credito

- Creditos por cliente, cuotas, abonos y estado de cuenta.
- Alertas y ranking de morosidad.
- Workflow de reversos/refinanciaciones con aprobacion.

### 4.7 Proceso talento humano (Hire-to-Retire)

- Asistencia de empleados.
- Nomina de sueldos.
- Vacaciones y licencias.

### 4.8 Soporte operacional y colaboracion

- Chat y tareas colaborativas usuario-admin con adjuntos.
- Soporte remoto por empresa (dispositivos y sesiones).
- Centro de ayuda operativo y modulo de calculadora empresarial.

## 5. Alcance tecnico incluido

- Backend en Go con handlers por dominio y capa de acceso a datos.
- Frontend web HTML/CSS/JS por panel de super y empresa.
- Operacion objetivo en PostgreSQL VPS con SQLite legado para contingencia/migracion.
- Integraciones externas de identidad, notificaciones, pagos, DIAN, mapas e IA.
- Instrumentacion de pruebas unitarias e integracion por dominio.

## 6. Fuera de alcance de esta version documental

1. Rediseno UX completo de todas las pantallas.
2. Reescritura del backend a microservicios.
3. Sustitucion total del stack frontend actual.
4. Migracion a otro proveedor cloud distinto del entorno ya operativo.
5. Cambios regulatorios pais-especificos fuera de los flujos ya modelados.

## 7. Actores y responsabilidades

- Super administrador: gobierno global, seguridad, configuraciones centrales.
- Admin empresa: operacion integral de su empresa.
- Supervisor sucursal: aprobaciones y control operativo.
- Cajero: operacion de venta y cobro.
- Inventario/compras/contabilidad: ejecucion especializada por dominio.
- Auditor: consulta de evidencia y trazabilidad.

## 8. Reglas de negocio de alcance (RB-ALC)

- RB-ALC-01: Toda operacion empresarial debe quedar aislada por `empresa_id`.
- RB-ALC-02: Los cambios criticos deben registrar trazabilidad auditable.
- RB-ALC-03: Las acciones de aprobacion/cierre requieren permiso explicito.
- RB-ALC-04: Estaciones deben permitir concurrencia multi-carrito por empresa.
- RB-ALC-05: Todo reporte debe exportarse al menos en PDF, XLS, CSV, JSON y TXT.
- RB-ALC-06: Integraciones contables deben conservar referencia por empresa, documento y periodo.
- RB-ALC-07: Secretos de integracion no se almacenan en texto plano.

## 9. Integraciones en alcance

- Identidad: OAuth Google para autenticacion administrativa.
- Correo: SMTP para confirmaciones, notificaciones y alertas operativas.
- Pagos: Wompi/Mercado Pago para operaciones de cobro en linea.
- Fiscal: DIAN para habilitacion, envio y acuse documental.
- Geoespacial: OpenStreetMap para modulos con mapa operativo.
- IA: proveedor configurable para asistencia empresarial.

## 10. Criterios de exito del alcance

1. Procesos empresariales clave documentados end-to-end.
2. Reglas de negocio multiempresa explicitadas y verificables.
3. Integraciones descritas con puntos de control operativos.
4. Entregables listos para revision de negocio, tecnologia y QA.
5. Coherencia con matrices de roles, entidades y KPI existentes.

## 11. Indicadores de aceptacion documental

- Cobertura de procesos: 100% de macroprocesos definidos en este documento.
- Cobertura de reglas: todas las reglas criticas RB-ALC trazables a modulos.
- Cobertura de integraciones: 100% de integraciones productivas registradas.
- Claridad de revision: estructura apta para comite funcional y tecnico.

## 12. Riesgos y mitigaciones

- Riesgo: dispersion documental entre multiples fuentes.
  - Mitigacion: uso de este alcance como entrada unica de alto nivel.
- Riesgo: desviacion entre operacion real y permisos definidos.
  - Mitigacion: validar siempre contra matriz de roles/permisos vigente.
- Riesgo: diferencias de motor de base durante transicion.
  - Mitigacion: mantener lineamientos de compatibilidad SQL y pruebas por dominio.

## 13. Entregables asociados

- 02_diseno_tecnico_erp_multiempresa.md
- 03_especificaciones_funcionales_erp_multiempresa.md
- 04_guia_implementacion_erp_multiempresa.md
