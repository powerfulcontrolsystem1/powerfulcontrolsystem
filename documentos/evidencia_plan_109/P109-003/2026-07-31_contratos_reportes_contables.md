# P109-003 - Contratos base de reportes contables

Fecha: 2026-07-31
Entorno: suite focalizada local del candidato de trabajo.

## Ejecución

Se ejecutó desde `backend`:

```text
go test ./handlers ./db -run 'TestReportSpecIARechazaInyeccionesYReferenciasFueraDelContrato|TestEmpresaReportesProgramacionSchemaReadyRejectsNilDatabase|TestEmpresaPermisos' -count=1
```

Resultado: `handlers` PASS y `db` PASS.

La cobertura comprobó que el ReportSpec fuera de contrato o con referencias
inyectadas se rechaza, que la programación de reportes no acepta una base nula
y que el contrato de permisos mantiene la protección esperada.

## Límite de la evidencia

Este PASS no concilia saldos ni certifica impuestos, exportaciones, cierres,
anulaciones o UAT de contador. P109-003 permanece parcial hasta ejecutar los
flujos oficiales sobre staging y su conciliación documentada.
