# Runbook: DIAN set de pruebas y diagnostico oficial

Fecha: 2026-04-18
Estado: vigente

Actualizacion 2026-06-06: `pruebas_dian` y `enviar_set_pruebas` ya no aceptan
simulacion. El cierre operativo del set requiere envio real al ambiente de
habilitacion, `ZipKey` y acuse final consultado con `GetStatusZip`.

## Sintomas cubiertos

- la empresa no logra pasar de onboarding DIAN a pruebas operativas.
- `diagnostico_oficial` devuelve brechas de alistamiento o dependencias faltantes.
- `enviar_set_pruebas` falla por rango, consecutivos o configuracion incompleta.
- la firma base o la firma XAdES de prueba no generan salida util.
- `enviar_documento_real` o `consultar_acuse_real` responden con error operativo o sin trazabilidad suficiente.

## Alcance

Aplica al endpoint base de Colombia bajo `/api/empresa/facturacion_electronica/dian` y a su uso desde el contexto autenticado de empresa. Este runbook cubre la base operativa ya implementada para onboarding, checklist, validacion, firma base, diagnostico y set de pruebas.

No cubre una integracion SOAP/WSDL oficial completa con DIAN. El propio backend deja esa brecha explicita en el reporte de alistamiento oficial.

## Fuentes de evidencia

- `backend/handlers/modulos_faltantes.go`
- configuracion DIAN de la empresa en la base operativa
- referencias secretas DIAN (`env:`, `file:`, `base64:`)
- respuesta JSON de `guia_onboarding`, `checklist`, `validar`, `diagnostico_oficial`
- XML base generado por `generar_xml_ubl_base`
- resultado de `firmar_xml_real` y `firmar_xml_xades_base`
- respuesta de `enviar_set_pruebas`, `enviar_documento_real` y `consultar_acuse_real`

## Verificaciones iniciales

1. Confirmar `empresa_id` y que el usuario este dentro del alcance autenticado correcto.
2. Consultar `action=guia_onboarding` y `action=checklist` para verificar el punto exacto del proceso.
3. Ejecutar `action=validar` o `action=diagnostico_oficial` antes de intentar envios reales.
4. Confirmar que los secretos DIAN no esten incrustados en texto plano y que las referencias `env:`, `file:` o `base64:` resuelvan valores no vacios.
5. Verificar que el software configurado sea `compartido` o `empresa` segun el escenario esperado, sin asumir que uno reemplaza el token o la firma por empresa.
6. Validar el rango y consecutivos antes de correr `enviar_set_pruebas`.
7. Para software propio o proveedor tecnologico, abrir `Facturacion electronica > Pasar test DIAN`, cargar el objetivo exacto mostrado por el portal DIAN y guardar modo de operacion, fechas, rango, totales requeridos y minimos aceptados. La base historica 60/20/20 solo sirve como respaldo si la empresa aun no tiene datos del portal.

## Causas probables

- credenciales DIAN incompletas o inconsistentes.
- certificado o clave de firma no cargados correctamente.
- referencias secretas vacias, mal escritas o no resolubles.
- rango de set de pruebas agotado o consecutivos fuera de rango.
- confusion entre XML base generado y envio oficial completo.
- expectativa incorrecta sobre el alcance actual del modulo, asumiendo que ya cubre transporte oficial DIAN de extremo a extremo.

## Acciones de recuperacion

1. Releer la salida de `guia_onboarding` y `checklist` para identificar el prerequisito exacto faltante antes de reintentar.
2. Ejecutar `validar_credenciales` si el problema apunta a token, software ID, prefijo, ambiente o datos tributarios incompletos.
3. Repetir `subir_firma` si el certificado o la clave se cargaron en formato incorrecto o quedaron asociados a referencias no validas.
4. Generar `generar_xml_ubl_base` y, si hace falta evidencia de firma, correr `firmar_xml_xades_base` para verificar que la salida base exista antes de intentar envios.
5. Usar `diagnostico_oficial` para distinguir entre una falla de configuracion local y una brecha del transporte oficial aun no implementado.
6. Si el error es de rango, corregir consecutivos o ampliar el tramo disponible antes de repetir `enviar_set_pruebas`.
7. Si el objetivo guardado no coincide con el portal, actualizarlo antes de repetir el set; los botones manuales pueden usar totales 1/0/0 para verificar recepcion por tipo sin consumir un lote completo.
8. Si el problema ocurre en `enviar_documento_real` o `consultar_acuse_real`, registrar la respuesta exacta y verificar primero que no se trate de una limitacion conocida del transporte oficial pendiente.
9. Si la empresa usa software `compartido`, confirmar que las referencias compartidas existan y que la empresa aun provea sus propios secretos exigidos por el flujo real.

## Validacion posterior

- `diagnostico_oficial` refleja menos brechas o deja claramente separada la brecha del transporte oficial pendiente.
- `generar_xml_ubl_base` produce una salida reutilizable.
- `firmar_xml_xades_base` o `firmar_xml_real` generan evidencia consistente de firma base.
- `enviar_set_pruebas` responde sin conflicto de rango, envia documentos reales y consulta `GetStatusZip` cuando DIAN devuelve `ZipKey`.
- el equipo entiende si el bloqueo restante es de datos/configuracion o del alcance aun no completado de la integracion oficial.

## Limites vigentes del modulo

1. El backend ofrece una base operativa util para onboarding, validacion, diagnostico, firma base y pruebas reales de habilitacion; la simulacion queda bloqueada en el set DIAN automatico.
2. El backend tiene sobres SOAP/WSDL base para envio y consulta, pero no debe prometer aceptacion fiscal sin acuse real DIAN/proveedor de la empresa.
3. El correo automatico actual envia resumen fiscal; el adjunto XML/PDF certificado por documento queda como brecha hasta persistir artefactos fiscales definitivos.
4. Cualquier incidencia debe clasificarse explicitamente en una de estas dos categorias:
   - error de configuracion o datos de la empresa
   - brecha de implementacion del transporte oficial

## Contrato relacionado

- `documentos/gobernanza_tecnica/contratos/contrato_facturacion_electronica_y_documentos_transaccionales.md`

## ADRs relacionados

- `ADR-0001-frontera-multiempresa-empresa-id.md`
- `ADR-0002-postgresql-runtime-canonico-vps.md`
