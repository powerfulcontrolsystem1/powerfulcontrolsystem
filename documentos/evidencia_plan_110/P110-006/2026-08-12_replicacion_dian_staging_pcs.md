# P110-006 — Réplica DIAN aislada de PCS en staging

Fecha: 2026-08-12  
Ámbito: empresa PCS (`empresa_id=12`) entre principal y staging. Producción no
fue modificada y no se emitieron documentos fiscales.

## Resultado verificable

- El instalador `deploy/scripts/vps-stage-dian-from-primary.sh` validó en
  principal una única configuración DIAN y la legibilidad del par privado de
  firma desde la identidad real del backend, sin imprimir material sensible.
- La réplica se creó solamente porque staging no tenía fila previa para la
  empresa. La clave y el certificado quedaron en almacenamiento privado de
  staging, con permisos del usuario del backend; la comprobación criptográfica
  confirmó que ambos siguen siendo el mismo par.
- La auditoría independiente de paridad informó principal completo y staging
  completo, con ambas referencias de archivo legibles. No expuso valores,
  contenido ni rutas en la evidencia.
- La comprobación de salvaguardas devolvió cuatro indicadores afirmativos:
  ambiente `habilitacion`, emisión local desactivada, estado `pendiente` y
  consecutivo reiniciado a cero.

## Salvaguardas y rollback

- El script exige `EMPRESA_ID` positivo, una confirmación explícita, una única
  fila fuente y ausencia de configuración destino; no sobrescribe staging.
- Fuerza el WSDL oficial de habilitación, no copia la activación local de
  producción y no realiza una llamada de emisión.
- Ante cualquier error posterior a la copia, elimina los archivos privados
  parciales. Los intentos de validación del instalador se detuvieron antes del
  `INSERT` y su limpieza fue comprobada antes del reintento final.
- El rollback de esta preparación es eliminar exclusivamente la fila y archivos
  de staging mediante un procedimiento revisado; no se usa SQL sobre documentos
  fiscales ni se toca principal.

## Siguiente control

Ejecutar bajo sesión autenticada de PCS las acciones no emisoras
`validar_credenciales`, `checklist` y `diagnostico_oficial`. Solo si estas
pruebas muestran habilitación coherente se agenda el set oficial DIAN y sus
acuse(s), que siguen siendo evidencia externa pendiente.
