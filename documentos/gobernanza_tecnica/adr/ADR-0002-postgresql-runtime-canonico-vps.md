# ADR-0002: PostgreSQL en VPS como runtime productivo canonico

Fecha: 2026-04-18
Estado: aceptada

## Contexto

El proyecto convivio historicamente con SQLite, pero la regla operativa vigente establece que la base de datos productiva corre en un VPS y el motor objetivo es PostgreSQL. Persistir decisiones nuevas basadas en SQLite como runtime principal genera errores futuros de compatibilidad, rendimiento y comportamiento.

## Decision

Se declara PostgreSQL en VPS como runtime productivo canonico del sistema.

SQLite queda limitado a:

- migracion historica,
- respaldos tecnicos puntuales,
- pruebas locales especificas cuando el modulo lo requiera.

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
