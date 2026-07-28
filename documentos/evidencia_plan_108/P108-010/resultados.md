# P108-010 - Centro IA: permiso efectivo y validacion visual

Fecha: 2026-07-28
Ambiente: staging autorizado
Empresa: 12 (Powerful Control System)
Rol: super_administrador autenticado
Artefacto observado: candidato publicado `20f8b726`.

## Alcance y resultado

Se abrio `centro_ia_empresarial.html?empresa_id=12` con sesion autorizada.
La interfaz mostro el unico interruptor **Modo agente**, apagado por defecto,
el campo de consulta y las acciones **Diagnostico ERP** y **Analizar con IA**;
no mostro selector de agente.

La carga inicial devolvio visiblemente:

`forbidden: rol sin acceso a la funcionalidad solicitada`

Resultado: **bloqueo seguro esperado**. La empresa no tiene habilitada
explicitamente `linkCentroIAEmpresarial`; la capacidad IA se oculta por
defecto por empresa. No se alteraron permisos, datos de negocio,
conversaciones ni acciones IA para convertir la validacion en falso positivo.

## Comprobacion de implementacion

La ruta permanece detras de `WithEmpresaReportesPermissions`, que exige
sesion, empresa valida, licencia, accion de lectura y permiso de pagina.
Se descarto promover una alternativa que usara solo el wrapper de tenant IA,
porque aun no reemplaza todos los controles funcionales de rol/licencia.

## Pendiente para aprobar P108-010

1. Habilitar la pagina de forma explicita y revisable en staging mediante el
   flujo oficial de permisos, sin reemplazar el conjunto existente.
2. Repetir GET, cada accion IA visible, cancelacion, doble clic y degradacion.
3. Verificar memoria por usuario/empresa, auditoria, latencia, tokens/costo y
   rechazo A/B en el mismo digest candidato.

Estado de fase: **parcial; no aprobada**.
