# P109-008 - Cuota empresarial para soportes CxP/IA

Fecha: 2026-08-08 02:00 America/Bogota  
Alcance: candidato local `00230e41` como base; sin despliegue ni mutacion de
PCS, staging o produccion.

## Cambio implementado

`/api/empresa/soportes_compras_ia?action=radicar` conserva el wrapper de
permisos del modulo y el `empresa_id` validado. Antes de guardar un adjunto
privado ahora:

1. suma el uso de `/uploads/empresas/<empresa>` y de
   `private_storage/soportes_compras_ia/empresa_<id>`;
2. aplica la configuracion global existente `empresa_storage` de limite,
   maximo por archivo, bloqueo y activacion;
3. rechaza con HTTP 507 y mensaje publico saneado si la carga excede la cuota;
4. mantiene el maximo estricto de soportes de 15 MiB y usa el menor limite si
   la configuracion corporativa define uno inferior.

El cliente no controla ruta, empresa, limite ni contador. No se cambiaron
tablas, rutas, permisos, dependencias ni archivos existentes.

## Pruebas locales

- cuota superada bloquea el adjunto de la misma empresa;
- llegar exactamente al limite permanece permitido, coherente con el adjunto
  de buzon;
- switches de cuota y bloqueo se respetan;
- maximo por archivo se aplica;
- los bytes privados de empresa 12 no incluyen los de empresa 7;
- `go test ./... -count=1`, `go vet ./...` y `git diff --check`: PASS.

## Limite

La evidencia es local. Falta construir/promover un nuevo candidato aislado y
probar HTTP 507, uso agregado y carreras entre replicas antes de considerar
resuelta la subcompuerta de cuota de P109-008.
