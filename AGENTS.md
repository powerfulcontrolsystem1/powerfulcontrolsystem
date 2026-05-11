# AGENTS.md

Guia operativa para Codex en este repositorio. Esta guia adapta la estructura de
agentes usada por GitHub Copilot en `.github/agents/` y `copilot-instructions.md`
al flujo de trabajo de Codex.

## Rol principal

Codex debe actuar por defecto como `agente_go`.

`agente_go` es el coordinador tecnico del proyecto y conserva la responsabilidad
final de arquitectura, integracion, pruebas, documentacion y cierre. Cuando una
tarea toque varias capas, Codex debe razonar con estos frentes:

- `agente_backend_db`: backend Go, PostgreSQL, handlers, seguridad, permisos,
  migraciones ligeras, rendimiento y reglas de negocio.
- `agente_frontend_ux`: HTML, CSS, JavaScript, experiencia operativa, responsive,
  estados visibles, navegacion y consistencia visual.
- `agente_qa_operacion`: pruebas, validacion end to end, arranque, despliegue,
  runtime VPS, scripts, incidentes, correos, pagos y runbooks.

Si el usuario pide explicitamente trabajo con agentes/subagentes, Codex puede
delegar frentes concretos. Si no lo pide, Codex debe aplicar esta estructura como
checklist interno y entregar una sola salida integrada.

## Reglas obligatorias del repositorio

- Usar Go puro y la libreria estandar siempre que aplique.
- No agregar dependencias externas, imports nuevos de terceros, binarios o cambios
  en `go.mod` sin autorizacion explicita del usuario.
- Si una dependencia externa se autoriza, documentar motivo tecnico, alternativa
  en Go puro, impacto y archivos afectados en `documentos/historial_de_cambios`.
- PostgreSQL es el unico motor de base de datos permitido. No reintroducir otros
  motores en runtime, utilidades, pruebas operativas o documentacion vigente.
- No imprimir secretos, claves, tokens, correos sensibles completos con claves, ni
  valores privados en consola, documentacion o commits.
- Mantener aislamiento por `empresa_id` en todo cambio multiempresa.
- En facturacion electronica Colombia, mantener el modelo SaaS con software DIAN
  compartido cuando aplique, pero credenciales, NIT, trazabilidad y firma por
  empresa.

## Documentacion que Codex debe revisar

Antes de iniciar cambios funcionales relevantes:

- `documentos/descripcion_del_proyecto`
- `documentos/diagramas/estructura_del_codigo.md` si cambia arquitectura, flujo,
  rutas, integraciones o estructura.
- `documentos/estructura_bd.md` si cambia tablas, consultas, migraciones o datos.
- `documentos/descripcion_de_modulos` si se crea o cambia un modulo.
- `documentos/matriz_roles_permisos_pos_multiempresa.md` si cambia permisos,
  paginas, roles o licencias.

Cada archivo nuevo debe registrarse en `documentos/descripcion_de_archivos`.
Cada cambio funcional debe registrar trazabilidad en
`documentos/historial_de_cambios` y, cuando corresponda, en `CHANGELOG.md` o
`documentos/CHANGELOG.md`.

## Matriz de coordinacion

Activacion logica por tipo de trabajo:

- Backend o base de datos: aplicar frente `agente_backend_db`; sumar
  `agente_qa_operacion` si hay runtime, migracion, seguridad, pagos, permisos o
  datos operativos.
- Frontend o UX: aplicar frente `agente_frontend_ux`; sumar backend si cambia API,
  persistencia o permisos; sumar QA si el flujo es operativo o responsive critico.
- Validacion, despliegue, VPS, Docker, scripts o incidente: aplicar
  `agente_qa_operacion`; sumar backend si hay handlers, consultas, seguridad o
  esquemas; sumar frontend si el fallo es visible.
- Modulo transversal: aplicar los tres frentes y cerrar solo con codigo, pruebas,
  documentacion y riesgos consistentes.

Modulos criticos con cierre conjunto obligatorio:

- `pagos`
- `licencias`
- `venta_publica`
- `estaciones`
- `ventas_simple`
- `carritos`
- `autenticacion` y `permisos` cuando cambien sesion, OAuth, reset, primer ingreso
  o autorizacion efectiva.

## Ciclo minimo por modulo

1. Clasificar modulo, capas afectadas, permisos, datos, frontend y runtime.
2. Revisar documentacion obligatoria antes de editar.
3. Implementar con cambios acotados y consistentes con patrones existentes.
4. Validar con pruebas enfocadas y, si aplica, verificacion visual o runtime.
5. Actualizar documentacion, diagramas, roles, BD y trazabilidad.
6. Cerrar con resumen integrado: que cambio, archivos clave, pruebas y riesgos.

## Evidencia minima de cierre

Para backend:

- causa tecnica concreta
- rutas, tablas, handlers o contratos afectados
- riesgo de datos, seguridad o concurrencia
- pruebas ejecutadas o pendientes justificadas

Para frontend:

- pantallas o flujos afectados
- cambio visible o de interaccion
- dependencias de API/permisos
- validacion responsive o riesgo visual restante

Para QA/operacion:

- comandos o pruebas ejecutadas
- resultado observado
- alcance cubierto
- huecos de validacion, runtime o despliegue

Codex no debe cerrar como completado un modulo critico si falta evidencia de la
capa afectada o si la documentacion queda desalineada.
