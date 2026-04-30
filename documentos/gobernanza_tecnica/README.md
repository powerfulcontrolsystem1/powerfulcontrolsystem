# Gobernanza tecnica del proyecto

Fecha: 2026-04-30
Estado: vigente

Este paquete instala la capa de gobernanza tecnica formal del proyecto para reducir errores de implementacion, mejorar la calidad de las decisiones, facilitar el trabajo de Copilot y sostener la escalabilidad del sistema.

## Objetivos

- Convertir decisiones arquitectonicas implicitas en decisiones trazables.
- Definir contratos tecnicos para flujos criticos antes de que se rompan por cambios laterales.
- Estandarizar cambios seguros por modulo, tipo de riesgo y evidencia minima.
- Transformar incidentes repetidos en runbooks operativos reutilizables.
- Reducir diferencias entre documentacion funcional, arquitectura, base de datos y comportamiento real.

## Componentes del paquete

- `plan_implementacion_gobernanza_tecnica.md`: roadmap por fases y entregables.
- `estandares_de_cambio_seguro.md`: reglas obligatorias para modificar codigo y documentacion tecnica.
- `adr/`: decisiones arquitectonicas registradas y aceptadas.
- `contratos/`: catalogo y contratos tecnicos por flujo critico.
- `runbooks/`: procedimientos operativos y de diagnostico.
- `runbooks/checklist_evidencia_documental_para_qa_y_soporte.md`: checklist corta para validar evidencia documental en QA y soporte.

## Avance actual 2026-04-30

- Contratos vigentes para checkout de licencias, estaciones, autenticacion, venta publica, permisos de api empresa, facturacion electronica y reportes multiformato.
- Runbooks vigentes para checkout de licencias, estaciones, arranque PostgreSQL por tunel local, DIAN set de pruebas y alertas de reinicio con Gmail SMTP.
- El contrato y runbook de checkout de licencias ya documentan el fallback clasico de Epayco firmado por POST a `https://secure.payco.co/checkout.php`, evitando redirecciones GET que producen XML `AccessDenied`.
- La documentacion canonica incorpora la secretaria IA, el robot/chat configurable por empresa y la regla de fallback a texto o voz de navegador cuando el servicio de voz falla.
- Se incorpora el flujo de empresas compartidas: consulta de administradores, revocacion desde ambos lados y trazabilidad en base super.
- Se incorpora el flujo de documentos dinamicos con IA mediante `/generate` y `/download`, con HTML/template como estructura intermedia y exportes PDF, DOCX, XLSX, HTML, TXT y JSON.
- La gobernanza ya distingue explicitamente entre capacidades operativas actuales y capacidades futuras no implementadas, especialmente en DIAN e integraciones sensibles.
- Quedan incorporados ademas el contrato de soporte remoto multiempresa y los runbooks de contingencia para reportes programados y para sesiones/dispositivos de soporte remoto.
- Se incorpora el frente financiero de periodos contables y conciliacion bancaria con contrato tecnico y runbook operativo alineados al backend real.
- Se incorpora tambien el frente de interoperabilidad documental y el runbook de contingencias para conectores e integraciones bancarias.
- Se incorpora ahora el contrato tecnico de integraciones externas y el runbook de reconciliacion documental fiscal y contable externa.
- Se incorpora ahora el contrato tecnico del repositorio documental con versionado, acceso por rol y firmas externas, junto con su runbook operativo de diagnostico.
- Se endurecen ahora las reglas de reconciliacion entre repositorio documental, interoperabilidad fiscal/contable y exportes regulatorios para evitar tomar reportes o PDFs como evidencia aislada.
- Se incorpora una checklist operativa breve para QA y soporte, pensada para revisar evidencia documental sin recorrer todo el paquete de contratos.

## Orden de uso

1. Leer `plan_implementacion_gobernanza_tecnica.md` para entender el orden de adopcion.
2. Aplicar `estandares_de_cambio_seguro.md` antes de tocar codigo o arquitectura.
3. Consultar el ADR aplicable si el cambio afecta frontera multiempresa, runtime, persistencia o integraciones.
4. Consultar el contrato tecnico si el cambio toca rutas, payloads, estados o side effects de un flujo critico.
5. Consultar el runbook si el cambio responde a una falla ya conocida o a un incidente de produccion.

## Criterios para crear artefactos nuevos

- Crear un ADR cuando la decision:
  - afecte varios modulos,
  - cambie el runtime oficial,
  - defina una frontera de seguridad o datos,
  - imponga compatibilidad hacia atras.
- Crear un contrato tecnico cuando el flujo:
  - sea publico,
  - active efectos persistentes,
  - dependa de integraciones externas,
  - tenga alto riesgo multiempresa.
- Crear un runbook cuando exista:
  - un incidente repetible,
  - un diagnostico no trivial,
  - una recuperacion operativa sensible,
  - un flujo donde soporte o QA puedan cometer errores por falta de pasos claros.

## Fuentes canonicas relacionadas

- Vision de negocio y modulo: `documentos/descripcion_del_proyecto`
- Arquitectura y mapa tecnico: `documentos/diagramas/estructura_del_codigo.md`
- Esquema fisico: `documentos/estructura_bd.md`
- Evolucion modular: `documentos/descripcion_de_modulos`
- Permisos y wrappers: `documentos/matriz_roles_permisos_pos_multiempresa.md`
