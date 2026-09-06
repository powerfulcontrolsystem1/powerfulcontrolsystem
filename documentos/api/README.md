# Contratos e inventarios API

Estado: Vigente. Responsable: Ingeniería backend. Revisión documental: 2026-09-05.

## Qué consultar

| Fuente | Propósito | Límite |
| --- | --- | --- |
| [Contratos por flujo](../gobernanza_tecnica/contratos/README.md) | Autorización, estados, reglas y efectos del negocio | Verificar contrato y pruebas del módulo tocado |
| [API móvil v1](mobile_api_v1.md) y [OpenAPI móvil](openapi.mobile.v1.yaml) | Contrato específico del canal `/api/v1/` | No atribuir automáticamente capacidades móviles futuras |
| [OpenAPI generado](openapi.generated.yaml) | Descubrimiento de rutas en código | Métodos inferidos/inventario no equivalen a contrato completo |
| [Ayuda API](ayuda_apis.md) e [inventario móvil](inventario_api_movil.md) | Navegación y alcance | No prueban seguridad ni despliegue |
| [Matriz de rutas multiempresa](../arquitectura/matriz_rutas_multiempresa.md) | Inspección de registros y wrappers | Un wrapper no prueba pertenencia de cada ID |

## Contrato mínimo por operación

Documentar método/ruta/acción, versión y consumidores; autenticación; fuente del tenant; rol/permiso/licencia; IDs secundarios; parámetros y esquema; límites; ejemplos ficticios; respuestas/errores; paginación; idempotencia, concurrencia y transacción; efectos externos; auditoría y privacidad; compatibilidad; pruebas positivas y negativas; responsable y estado de aceptación.

Los errores no exponen detalles privados; documentar el código HTTP y la respuesta realmente implementados, sin imponer retrospectivamente un formato nuevo a rutas heredadas. No escribir un ejemplo exitoso para una operación fiscal bloqueada.

## Cambio y compatibilidad

Antes de modificar una operación, inventariar UI, móvil, worker, webhook y conectores que la consumen. Cambios incompatibles requieren estrategia explícita: versión nueva, transición o retirada controlada con consumidores y fecha. Nunca marcar una ruta retirada solo porque desapareció del menú.

Actualizar contrato y generador/overrides aplicables en la misma entrega; comparar inventario con `backend/main.go`. El [informe del 2026-09-05](../informe_general_produccion_seguridad_2026-09-05.md) registra drift de la matriz de rutas, pendiente de reconciliar con el candidato funcional.

## Verificación

Aplicar [checklist multiempresa](../checklist_seguridad_endpoint_multiempresa.md) y [requisitos](../requisitos/especificacion_y_trazabilidad.md). Los tests deben verificar denegación sin efectos, IDs ajenos, licencia, límites, reintentos y estado ante fallos de proveedor. No publicar tokens, cookies ni datos de cliente en ejemplos OpenAPI.
