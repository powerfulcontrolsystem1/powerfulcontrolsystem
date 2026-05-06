# Propiedad Horizontal / Administracion de Copropiedades

Fecha: 2026-05-06

## Objetivo

Modulo empresarial para administrar conjuntos residenciales, edificios, condominios, copropiedades mixtas y centros comerciales bajo el alcance de una empresa. No reemplaza apartamentos turisticos: este modulo maneja administracion de copropiedad, cuotas, residentes, cartera, recaudos, PQR y asambleas.

## Alcance funcional

- Configuracion de copropiedad: nombre, NIT, tipo, direccion, ciudad, administrador, contacto, interes de mora, dias de gracia, facturacion electronica y portal de residente.
- Unidades privadas y comunes: torre, piso, tipo, area, coeficiente, cuota base, parqueadero, deposito y estado.
- Propietarios, residentes, arrendatarios y apoderados por unidad.
- Cargos y cartera: cuotas de administracion, cuotas extraordinarias, multas, intereses, reservas y otros conceptos.
- Recaudos: metodo de pago, referencia, valor pagado y afectacion automatica del saldo del cargo.
- PQR y mantenimiento: peticiones, quejas, reclamos, solicitudes, responsables, prioridad y fecha limite.
- Asambleas: programacion, tipo, quorum objetivo, quorum actual, acta y estado.
- Dashboard con unidades, ocupacion, residentes activos, cartera pendiente, recaudo del mes, PQR pendientes y asambleas abiertas.
- Datos demo para acelerar pruebas y puesta en marcha.

## Seguridad y permisos

- Clave de modulo/licencia: `propiedad_horizontal`.
- Pagina empresarial: `web/administrar_empresa/propiedad_horizontal.html`.
- Endpoint protegido: `/api/empresa/propiedad_horizontal`.
- Wrapper: `WithEmpresaPropiedadHorizontalPermissions`.
- Todas las tablas usan `empresa_id`; no hay rutas publicas ni mezcla de informacion entre empresas.

## Tablas

- `empresa_propiedad_horizontal_config`
- `empresa_propiedad_horizontal_unidades`
- `empresa_propiedad_horizontal_personas`
- `empresa_propiedad_horizontal_cargos`
- `empresa_propiedad_horizontal_recaudos`
- `empresa_propiedad_horizontal_pqrs`
- `empresa_propiedad_horizontal_asambleas`

## Integraciones previstas

- Facturacion electronica para cuotas cuando la empresa lo active.
- Finanzas/cartera para conciliacion de recaudos.
- Reportes para cartera por unidad, recaudos, paz y salvo, PQR y asambleas.
- Portal de residentes como capa futura de autoservicio externo.
