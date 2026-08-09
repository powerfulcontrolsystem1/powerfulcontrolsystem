# P109-009 - Aislamiento A/B con identidad empresarial no global

Fecha: 2026-08-09

Candidato desplegado solo en staging: `34cfd852`

Empresa autorizada: PCS (`empresa_id=12`)

Empresa cruzada: `empresa_id=7`

## Objetivo

Comprobar con una identidad empresarial real, sin privilegio global, que el
backend aplica el alcance de empresa y el permiso efectivo aunque el cliente
altere `empresa_id` en la URL.

## Preparación y autenticación

- Se reutilizó una cuenta temporal de QA previamente inactiva.
- El rol se cambió por el flujo oficial de administración y la cuenta se activó
  por la API oficial.
- La contraseña histórica fue rechazada con HTTP 401.
- La recuperación oficial respondió HTTP 200, el mensaje llegó al buzón y el
  enlace permitió establecer una contraseña efímera e iniciar sesión.
- No se registran correo, contraseña, token ni contenido privado en esta
  evidencia.

## Matriz A/B

| Módulo | PCS 12 | Empresa 7 | Fuga PCS |
|---|---|---|---|
| Soportes de compras e IA | Carga y lista operativa | `forbidden: empresa_id fuera del alcance del usuario autenticado` | No |
| Finanzas | Carga operativa | `forbidden` | No |
| Contabilidad | Carga operativa | `forbidden` | No |
| Facturación electrónica DIAN | Carga operativa | `forbidden` | No |
| Documentos | Carga operativa | `forbidden` | No |

La variante cruzada se inspeccionó visualmente. No mostró filas, importes,
documentos, proveedores ni datos identificables de PCS.

## Permiso efectivo y no mutación

Con la misma identidad se intentó enviar `SCI-0011` a papelera desde el diálogo
visible, con motivo. El backend respondió:

`forbidden: rol sin permiso para la accion solicitada`

La verificación posterior por el endpoint oficial conservó `SCI-0011` como
registro `activo` y workflow `rechazado`. No se creó CxP, pago, movimiento ni
evento de papelera.

## Limpieza

- La cuenta temporal se desactivó por la API oficial: HTTP 204.
- La sesión emitida antes de la desactivación devolvió `unauthorized` al repetir
  la carga.
- Se cerró la pestaña del correo y las sesiones de prueba con contenido privado.
- Producción no recibió despliegue ni mutación.

## Resultado

**PASS** para aislamiento de lectura A/B y denegación por permiso efectivo con
identidad no global en los cinco módulos cubiertos. Esto elimina el hueco A/B
de P109-009, pero la fase permanece parcial por la migración CSP `unsafe-inline`
y las demás pruebas dinámicas pendientes del mismo candidato.
