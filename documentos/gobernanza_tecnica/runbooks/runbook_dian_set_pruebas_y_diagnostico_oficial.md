# Runbook: DIAN set de pruebas y diagnostico oficial

Estado: Vigente. Responsable: QA/operación. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- El set de habilitación es un envío real al ambiente autorizado, no un dry-run. Se exige el objetivo exacto del portal y fuente comercial genuina fuera del set.
- La nómina no reutiliza el PDF/correo de factura ni su habilitación. Los diagnósticos locales no acreditan acuse oficial.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

Fecha: 2026-04-18

La matriz por familia del [contrato fiscal](../contratos/contrato_facturacion_electronica_y_documentos_transaccionales.md) define el alcance actual. Los folios y acuses fechados anteriores son evidencia histórica; no reutilizar numeración de un ejemplo.

## Sintomas cubiertos

- la empresa no logra pasar de onboarding DIAN a pruebas operativas.
- `diagnostico_oficial` devuelve brechas de alistamiento o dependencias faltantes.
- `enviar_set_pruebas` falla por rango, consecutivos o configuracion incompleta.
- la firma base o la firma XAdES de prueba no generan salida util.
- `enviar_documento_real` o `consultar_acuse_real` responden con error operativo o sin trazabilidad suficiente.
- DIAN responde `Regla 90` o el usuario no encuentra inicialmente una factura en el portal.

## Alcance

Aplica al endpoint base de Colombia bajo `/api/empresa/facturacion_electronica/dian` y a su uso desde el contexto autenticado de empresa. Este runbook cubre la base operativa ya implementada para onboarding, checklist, validacion, firma base, diagnostico, `SendTestSetAsync` y `GetStatusZip` reales contra el endpoint SOAP/WCF oficial DIAN.

## Fuentes de evidencia

- `backend/handlers/modulos_faltantes.go`
- configuracion DIAN de la empresa en la base operativa
- referencias secretas DIAN (`env:`, `file:`, `base64:`)
- respuesta JSON de `guia_onboarding`, `checklist`, `validar`, `diagnostico_oficial`
- XML base generado por `generar_xml_ubl_base`
- resultado del despachador interno de firma; las actions directas
  `firmar_xml_real` y `firmar_xml_xades_base` permanecen bloqueadas
- respuesta de `enviar_set_pruebas` y `consultar_acuse_real`; la action directa
  `enviar_documento_real` no es una ruta comercial permitida

## Verificaciones iniciales

1. Confirmar `empresa_id` y que el usuario este dentro del alcance autenticado correcto.
2. Consultar `action=guia_onboarding` y `action=checklist` para verificar el punto exacto del proceso.
3. Ejecutar `action=validar` o `action=diagnostico_oficial` antes de intentar envios reales.
4. Confirmar que los secretos DIAN no esten incrustados en texto plano y que las referencias `env:`, `file:` o `base64:` resuelvan valores no vacios.
5. Verificar que el software configurado sea `compartido` o `empresa` segun el escenario esperado, sin asumir que uno reemplaza el token o la firma por empresa.
6. Validar el rango y consecutivos antes de correr `enviar_set_pruebas`.
7. Para software propio o proveedor tecnologico, abrir `Facturacion electronica > Pasar test DIAN`, cargar el objetivo exacto mostrado por el portal DIAN y guardar modo de operacion, fechas, rango, totales requeridos y minimos aceptados. No usar 60/20/20, 30/10/10 ni otro valor historico como sustituto del objetivo asignado.
8. En produccion, confirmar el siguiente consecutivo contra `empresa_dian_configuracion`, `empresa_configuracion_avanzada`, `empresa_facturacion_documentos`, `facturacion_electronica_reintentos` y portal DIAN cuando haya duda.
9. Confirmar el tipo documental. Solo `factura_electronica` puede entrar al
   adaptador comercial actual. Un 422
   `tipo_documento_dian_no_implementado` es un bloqueo correcto, no una
   contingencia ni un caso reintentable.

## Causas probables

- credenciales DIAN incompletas o inconsistentes.
- certificado o clave de firma no cargados correctamente.
- referencias secretas vacias, mal escritas o no resolubles.
- rango de set de pruebas agotado o consecutivos fuera de rango.
- confusion entre XML base generado y envio oficial completo.
- expectativa incorrecta sobre el alcance actual del modulo, asumiendo que las
  familias catalogadas ya tienen emision productiva.

## Acciones de recuperacion

1. Releer la salida de `guia_onboarding` y `checklist` para identificar el prerequisito exacto faltante antes de reintentar.
2. Ejecutar `validar_credenciales` si el problema apunta a token, software ID, prefijo, ambiente o datos tributarios incompletos.
3. Repetir `subir_firma` si el certificado o la clave se cargaron en formato incorrecto o quedaron asociados a referencias no validas.
4. Para un set de habilitación, `generar_xml_ubl_base` solo admite el marcador
   explícito `set_habilitacion=true`. La firma y el envío libres están
   bloqueados; usar `pruebas_dian` o el flujo documental canónico. Una factura
   comercial debe provenir de `fuente_fiscal_json` y nunca de ese fixture.
5. Usar `diagnostico_oficial` para distinguir entre una falla de configuracion local y una brecha del transporte oficial aun no implementado.
6. Si el error es de rango, corregir consecutivos o ampliar el tramo disponible antes de repetir `enviar_set_pruebas`.
7. Si el objetivo guardado no coincide con el portal, actualizarlo antes de repetir el set. Una prueba manual consume folios y solo debe ejecutarse con alcance confirmado y trazabilidad del documento.
8. Si el problema ocurre en `consultar_acuse_real`, registrar la respuesta
   saneada y verificar el TrackId individual. Nunca propagar su rechazo al
   estado global de la empresa. `enviar_documento_real` directo debe permanecer
   bloqueado para emisiones comerciales.
9. Si la empresa usa software `compartido`, confirmar que las referencias compartidas existan y que la empresa aun provea sus propios secretos exigidos por el flujo real.
10. Si DIAN devuelve `Regla 90`, consultar primero el portal, CUFE/TrackId o historial de acuse original. No marcar el documento como aceptado solo por esa regla.
11. Si el portal muestra `Aprobado con notificacion`, registrar el documento como aprobado y conservar la notificacion como observacion; `RUT01` no bloqueo `1PCS3`.
12. Si un folio se consumió fuera del flujo documental, detener nuevas emisiones de ese rango y abrir reconciliación autorizada desde evidencia oficial; no editar contadores a ciegas.

## Validacion posterior

- `diagnostico_oficial` refleja menos brechas o deja claramente separada la brecha del transporte oficial pendiente.
- el set explícito genera su fixture UBL; la factura comercial genera líneas y
  partes desde la fuente fiscal privada e inmutable.
- el XML firmado pasa `scripts/validar_dian_xsd.ps1` y
  `scripts/validar_dian_firma.ps1` antes de usarlo como candidato de
  habilitación. El acuse DIAN sigue siendo la evidencia fiscal final.
- `enviar_set_pruebas` responde sin conflicto de rango, envia documentos reales y consulta `GetStatusZip` cuando DIAN devuelve `ZipKey`.
- la consulta de conectividad no procesa cola; el boton manual usa `POST` y `pcs-worker` procesa automaticamente los vencidos con bloqueo por empresa.
- cada documento enviado conserva XML firmado, acuse y representacion PDF privados con SHA-256; un TrackId pendiente se consulta sin regenerar el XML.
- el equipo entiende si el bloqueo restante es de datos/configuracion, transporte DIAN, portal DIAN o evidencia de acuse.
- Para PCS produccion, `GetNumberingRange` confirmo el rango `1PCS 1-100000` y
  la configuracion mostro siguiente folio 12 el 2026-08-24; debe reconfirmarse
  inmediatamente antes de cualquier envio real.

## Limites vigentes del modulo

1. El backend ofrece onboarding, diagnóstico, set real de habilitación y un
   candidato de factura Colombia por SOAP/WCF basado en fuente fiscal real. No
   se declara producción lista hasta validar migraciones PostgreSQL, Schematron,
   portal y acuse real de la empresa.
2. El backend no debe prometer aceptacion fiscal sin acuse real DIAN/proveedor, documento visible en portal DIAN o evidencia oficial equivalente.
3. El correo de Colombia produccion se envia despues de la aceptacion DIAN y adjunta el XML firmado y la representacion PDF persistidos. El XML y el acuse siguen siendo la evidencia fiscal autentica.
4. Cualquier incidencia debe clasificarse explicitamente en una de estas dos categorias:
   - error de configuracion o datos de la empresa
   - brecha de implementacion del transporte oficial
5. Factura, anulación total, soporte ordinario y nómina ordinaria tienen rutas específicas. Crédito parcial, débito, ajustes, equivalentes y RADIAN permanecen bloqueados; no enviar una familia por el adaptador de otra.

## Contrato relacionado

- `documentos/gobernanza_tecnica/contratos/contrato_facturacion_electronica_y_documentos_transaccionales.md`

## ADRs relacionados

- `ADR-0001-frontera-multiempresa-empresa-id.md`
- `ADR-0002-postgresql-runtime-canonico-vps.md`

## Fuentes y aceptación de la revisión

[contrato_facturacion_electronica_y_documentos_transaccionales.md](../contratos/contrato_facturacion_electronica_y_documentos_transaccionales.md), [modulos_faltantes.go](../../../backend/handlers/modulos_faltantes.go).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](../../requisitos/especificacion_y_trazabilidad.md)).
