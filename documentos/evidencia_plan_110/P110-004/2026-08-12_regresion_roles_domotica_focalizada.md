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
