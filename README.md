# Powerful Control System

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-05.


PCS es un POS/ERP SaaS multiempresa con backend Go, PostgreSQL y frontend HTML, CSS y JavaScript. Integra operación comercial, administración empresarial y servicios externos con controles de acceso por empresa.

## Empezar

1. Leer el [contexto general](documentos/contexto_general_del_sistema.md): alcance, reglas y límites actuales.
2. Abrir el [portal documental](documentos/README.md) y la [guía de incorporación](documentos/desarrollo/incorporacion.md).
3. Ubicar el módulo en el [mapa técnico](documentos/mapa_modulos.md) antes de modificar código.
4. Seguir [CONTRIBUTING](CONTRIBUTING.md) para implementar, probar y entregar.

Los agentes leen además [AGENTS.md](AGENTS.md) y el [contexto operativo para IA](documentos/contexto_codex.md).

## Estructura

| Directorio | Responsabilidad |
| --- | --- |
| `backend/` | API Go, reglas, acceso PostgreSQL, worker y migrador |
| `web/` | Portal, paneles, páginas empresariales y recursos estáticos |
| `deploy/` | Docker, configuración de despliegue y operación |
| `scripts/`, `tools/` | Orquestación y validadores |
| `documentos/` | Fuentes técnicas, contratos, procedimientos e historia |
| `.github/` | CI, revisión y coordinación |

## Estado y seguridad

La existencia de una pantalla o adaptador no acredita su disponibilidad comercial. El [estado de entrega](documentos/estado_actual.md) remite a evidencia fechada y mantiene el NO-GO general informado por la auditoría del 2026-09-05. Esta revisión documental no verifica el VPS ni despliega código.

Reportar vulnerabilidades por el canal privado de [SECURITY.md](SECURITY.md). Nunca incluir credenciales, datos empresariales o certificados en documentación, ejemplos ni incidencias públicas.

La organización documental adopta referencias internacionales con [alcance y brechas explícitos](documentos/gobernanza_tecnica/marco_documental.md); no constituye certificación ISO, auditoría legal ni promesa de SLA.
