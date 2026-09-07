# ADR-0002: PostgreSQL en VPS como runtime productivo canonico

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Se elimina la excepción antigua que permitía pruebas con otro motor, incompatible con AGENTS.md.
- El túnel a VPS no convierte una base real en entorno de pruebas: usar PostgreSQL aislado para integración.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

Fecha: 2026-04-18

## Contexto

El proyecto convivio historicamente con motor legado retirado, pero la regla operativa vigente establece que la base de datos productiva corre en un VPS y el motor objetivo es PostgreSQL. Persistir decisiones nuevas basadas en motor legado retirado como runtime principal genera errores futuros de compatibilidad, rendimiento y comportamiento.

## Decision

Se declara PostgreSQL en VPS como runtime productivo canonico del sistema.

Aclaración 2026-09-05: PostgreSQL es el único motor permitido también en utilidades y pruebas operativas. Los respaldos históricos de motores retirados son archivos de antecedente, no runtime autorizado.

## Consecuencias

### Positivas

- las decisiones nuevas se alinean con el entorno real del sistema.
- obliga a diseñar inserciones, secuencias, migraciones y consultas pensando en PostgreSQL.
- reduce errores por diferencias de autoincremento, sintaxis o runtime local.

### Costos

- toda implementacion nueva debe validarse con compatibilidad PostgreSQL.
- las rutas de saneamiento legacy deben seguir existiendo donde el sistema arrastre esquemas anteriores.

## Aplicacion inmediata

- no introducir nuevas implementaciones que dependan de `LastInsertId` como comportamiento principal.
- documentar y validar secuencias, defaults y compatibilidad de insercion en tablas transaccionales.
- tratar `.env.local` y tuneles a VPS como parte del runtime real de desarrollo.

## Fuentes y aceptación de la revisión

[AGENTS.md](../../../AGENTS.md), [gobierno_datos.md](../../arquitectura/gobierno_datos.md).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](../../requisitos/especificacion_y_trazabilidad.md)).
