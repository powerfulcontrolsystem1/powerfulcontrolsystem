# Plantilla profesional de onboarding DIAN por empresa (modelo SaaS multiempresa)

Fecha: 2026-04-09

## 1) Objetivo
Estandarizar la configuracion de facturacion electronica DIAN para cada empresa en un modelo SaaS donde:
- El software DIAN puede ser compartido.
- La identidad tributaria y la firma digital son por empresa.

## 2) Regla operativa clave
- Compartido (plataforma): `Software ID` y `Software PIN`.
- Por empresa (obligatorio):
  - `nit`
  - `token_emisor_ref`
  - `certificado_clave_ref`
  - `prefijo`, `resolucion_numero`, `rango_desde`, `rango_hasta`, `consecutivo_actual`

## 3) Checklist de onboarding por empresa
1. Crear/actualizar la configuracion DIAN base de la empresa.
2. Activar modo SaaS con `usar_software_compartido=1`.
3. Configurar token de emisor por empresa (`token_emisor_ref`).
4. Subir firma PEM por empresa (`action=subir_firma`) o registrar referencia segura en `certificado_clave_ref`.
5. Ejecutar `action=checklist` y `action=validar`.
6. Ejecutar `action=validar_credenciales`.
7. Ejecutar `action=enviar_set_pruebas` con `simular=true`.
8. Ejecutar `action=enviar_set_pruebas` real (`simular=false`) cuando validaciones esten en verde.

## 4) Payload base recomendado por empresa
```json
{
  "empresa_id": 101,
  "codigo": "DIAN-EMP-101",
  "nit": "900123456",
  "digito_verificacion": "7",
  "razon_social": "Empresa Demo 101 SAS",
  "tipo_ambiente": "habilitacion",
  "usar_software_compartido": 1,
  "software_id_compartido_ref": "env:DIAN_SHARED_SOFTWARE_ID",
  "software_pin_compartido_ref": "env:DIAN_SHARED_SOFTWARE_PIN",
  "test_set_id": "SET-EMPRESA-101",
  "prefijo": "SETP",
  "resolucion_numero": "1876000000101",
  "resolucion_fecha_desde": "2026-01-01",
  "resolucion_fecha_hasta": "2030-01-01",
  "rango_desde": 1000,
  "rango_hasta": 999999,
  "consecutivo_actual": 1000,
  "url_dian": "https://vpfe-hab.dian.gov.co/WcfDianCustomerServices.svc?wsdl",
  "token_emisor_ref": "env:DIAN_TOKEN_EMP_101"
}
```

## 5) Endpoints de apoyo por empresa
- CRUD DIAN: `GET/POST/PUT/DELETE /api/empresa/facturacion_electronica/dian`
- Guia: `GET /api/empresa/facturacion_electronica/dian?action=guia_onboarding&empresa_id={id}`
- Checklist: `GET /api/empresa/facturacion_electronica/dian?action=checklist&empresa_id={id}`
- Validacion: `GET /api/empresa/facturacion_electronica/dian?action=validar&empresa_id={id}`
- Validacion de credenciales: `POST /api/empresa/facturacion_electronica/dian?action=validar_credenciales`
- Subir firma: `POST /api/empresa/facturacion_electronica/dian?action=subir_firma` (multipart)
- Set pruebas: `POST /api/empresa/facturacion_electronica/dian?action=enviar_set_pruebas`

## 6) Ejemplos operativos (cURL)

### 6.1 Validar credenciales por empresa
```bash
curl -s -X POST "http://localhost:8080/api/empresa/facturacion_electronica/dian?action=validar_credenciales" \
  -H "Content-Type: application/json" \
  -d '{"empresa_id":101}'
```

### 6.2 Subir firma PEM por empresa
```bash
curl -s -X POST "http://localhost:8080/api/empresa/facturacion_electronica/dian?action=subir_firma" \
  -F "empresa_id=101" \
  -F "archivo_firma=@empresa_101_key.pem"
```

### 6.3 Simular set de pruebas
```bash
curl -s -X POST "http://localhost:8080/api/empresa/facturacion_electronica/dian?action=enviar_set_pruebas" \
  -H "Content-Type: application/json" \
  -d '{"empresa_id":101,"facturas_electronicas":30,"notas_debito":10,"notas_credito":10,"total_documentos":50,"simular":true}'
```

## 7) Plantilla de evidencia por empresa

Empresa ID: ________
NIT: ________
Responsable: ________
Fecha onboarding: ________

Checklist:
- [ ] Configuracion base DIAN guardada
- [ ] Modo SaaS activado (software compartido)
- [ ] Token por empresa configurado
- [ ] Firma por empresa cargada/validada
- [ ] Checklist/validar en verde
- [ ] Validar credenciales en verde
- [ ] Set simulado completado
- [ ] Set real completado

Observaciones:
- _______________________________________
- _______________________________________

## 8) Seguridad minima obligatoria
- No guardar secretos en texto plano dentro del repositorio.
- Preferir referencias `env:` o `file:` en rutas protegidas.
- No reutilizar token o firma entre empresas.
- Mantener trazabilidad por `empresa_id`, documento y periodo.
