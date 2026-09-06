# Incorporación de ingeniería y entorno de desarrollo

Estado: Vigente. Responsable: Ingeniería y QA/operación. Revisión documental: 2026-09-05.

## Resultado esperado

Un ingeniero puede localizar una función desde página hasta handler, autorización, datos y prueba; preparar validación local aislada y entregar un cambio revisable. No necesita leer cronológicamente todos los planes ni recibir secretos por chat.

## Herramientas y fuentes

- Git y PowerShell en el entorno Windows del proyecto; comandos Go desde `backend/`.
- Versión Go definida en [go.mod](../../backend/go.mod) y CI. En este corte: lenguaje 1.26.0 y toolchain 1.26.6; comprobar el archivo al preparar el entorno.
- Node para herramientas `.mjs`; CI usa Node 22. No se incorpora un bundler al frontend.
- PostgreSQL aislado para integración; Docker/Compose para los procedimientos de staging y contenedores ya existentes. No se utiliza un motor alternativo en tests operativos.
- Dependencias existentes se resuelven por los manifiestos del repositorio; nuevas instalaciones o cambios de dependencias se deciden por su alcance.

## Recorrido inicial

1. Leer [contexto](../contexto_general_del_sistema.md), [arquitectura](../arquitectura/descripcion_arquitectura.md) y [estado](../estado_actual.md).
2. Elegir un módulo en el [mapa](../mapa_modulos.md). Localizar su página, ruta en `backend/main.go`, wrapper, handler, tabla y test.
3. Leer [decisiones](../decisiones_tecnicas.md), [checklist multiempresa](../checklist_seguridad_endpoint_multiempresa.md) y contrato específico.
4. Confirmar con operación qué entorno y datos están autorizados antes de iniciar procesos de aplicación.

## Inspección y compilación local

Los siguientes comandos no arrancan servicios ni aplican migraciones; el toolchain puede resolver dependencias de red. Ejecutar desde la raíz y usar la configuración de Go aprobada:

```powershell
git status --short
git rev-parse HEAD
go version
node --version
node tools/docs_catalog.mjs --check
Set-Location backend
go build ./...
Set-Location ..
```

Para pruebas enfocadas usar la sección `Pruebas Go` de [comandos](../comandos_codex.md) y la [estrategia QA](../calidad/estrategia_verificacion.md). Revisar antes si el test necesita DSN, schema restaurado o proveedor. No imprimir entorno completo, archivos `.env`, cookies ni credenciales.

## Preparar ejecución aislada

1. Revisar [manual de instalación](../manual_de_instalacion.md), [Docker/VPS](../docker_vps_operacion.md) y [staging](../gobernanza_tecnica/runbooks/runbook_staging_ci_e2e.md). Los ejemplos históricos requieren cotejo con scripts actuales.
2. Proveer dos bases PostgreSQL del entorno de prueba y almacenamiento privado exclusivo. Para staging usar Compose de plataforma más su override y el script documentado; no usar solo el override como stack completo.
3. Cargar secretos por el mecanismo privado aprobado. Ver [configuración](configuracion_y_entornos.md); no copiar configuración productiva ni crear credenciales de ejemplo válidas.
4. Mantener aislados correo, pagos, DIAN, IA y dispositivos hasta que el caso de prueba autorice sus efectos. El worker puede despachar jobs al iniciar.
5. Ejecutar migración y arranque únicamente conforme al runbook y entorno autorizado; comprobar DSN/volúmenes/puertos por identidad segura, sin divulgar sus valores privados.
6. Validar readiness y luego login, tenant, permisos y flujo del módulo. Abrir una página desde un servidor estático no prueba backend, autenticación ni datos PCS.

## Primera entrega y transferencia

Completar un cambio pequeño con requisito, contrato, implementación y test; pedir revisión según CODEOWNERS. Entregar comandos/resultados, limitaciones, pasos de reproducción segura y documentación actualizada. Los accesos, contactos y guardia pertenecen al directorio privado de la organización, pendiente de asignación nominal; no se inventan en este manual.
