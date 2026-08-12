# P110-004 — regresión focalizada de permisos y domótica

Fecha: 2026-08-12  
Ámbito: candidato local equivalente al código de staging; no se enviaron
órdenes GPIO ni se modificaron datos empresariales.

## Verificación

Se ejecutaron las pruebas enfocadas de `handlers` y `db` para permisos de
empresa, invitaciones y límites de cajero, sesión, aislamiento de CxP/IA y
contratos de control eléctrico. También aprobó `go vet ./handlers ./db`.

```text
go test ./handlers ... PASS
go test ./db ... PASS
go vet ./handlers ./db PASS
```

La cobertura confirma que el contrato de rol no se eleva por correo, conserva
las restricciones de cajero y que las rutas de domótica mantienen empresa y
estación como contexto de servidor. No sustituye una identidad real por cada
rol ni una prueba física de relés.

## Límite

P110-004 sigue parcial hasta ejecutar la matriz mutante con identidades y
cajas reales del candidato final; la validación GPIO requiere hardware PCS
registrado y supervisión humana.

## Revisión visual autenticada posterior

En PCS/staging se actualizaron desde la interfaz las vistas de Finanzas y
Domótica, sin acciones mutantes. Finanzas mostró tablas de movimientos y
conciliación por período con dos períodos conciliados, cero pendientes, cero
errores y desfase monetario cero en la vista actual. Los controles de cartera,
extractos, exportación, procesamiento y estados quedaron visibles y etiquetados.

Domótica confirmó el contexto de empresa, módulo activo y estado seguro sin
hardware: cero Raspberry activas, cero estaciones, cero aparatos, consumo 0 W y
sin último evento. No se pulsaron agenda, sincronización de túnel ni relés.
Esto confirma que no se puede declarar validación GPIO física hasta registrar y
supervisar hardware real.
