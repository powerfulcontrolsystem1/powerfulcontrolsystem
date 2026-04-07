# Diagrama de flujo de procesos

Fecha: 2026-04-07

```mermaid
flowchart TD
    A[Administrador crea usuario de empresa] --> B[Generar token de confirmacion]
    B --> C[Enviar correo SMTP]
    C -->|Falla| D[Rollback: usuario no se registra]
    C -->|Exito| E[Usuario abre enlace de confirmacion]
    E --> F[Correo confirmado: estado activo]
    F --> G[Usuario abre login_usuario]
    G --> H{Tiene contrasena?}
    H -->|No| I[Crear contrasena primer ingreso]
    I --> J[Guardar hash/salt]
    H -->|Si| K[Autenticacion email+password]
    J --> K
    K --> L[Sesion iniciada]

    SA0[Super administrador abre configuracion avanzada] --> SA1[Gestionar credencial IA de Google Gemini]
    SA1 --> SA2[Guardar en endpoint super de configuracion IA]
    SA2 --> SA3[Registrar cuenta Google autenticada que actualiza credenciales]
    SA3 --> AI1

    L --> L0[Abrir administrar_empresa]
    L0 --> AUTHZ[Middleware permisos valida rol y empresa_id por modulo]
    L0 --> L1[Cargar Inicio por defecto en Panel de la Empresa]
    L1 --> L1A[Abrir Centro de Ayuda]
    L1A --> L1B[Consultar menu interno y APIs operativas]
    L1 --> L11[Usuario navega a una subpagina del panel]
    L11 --> L12[Guardar subpagina actual en sessionStorage]
    L12 --> L13[Al presionar F5, restaurar la misma subpagina]

    L --> S0[Configurar colores de estado del carrito en Facturación electrónica]
    S0 --> S1[Guardar color activo/inactivo en la sección Configuración avanzada integrada]
    S1 --> S2[Sincronizar estaciones con carritos en estado inactivo/cerrado]
    S2 --> S3[Abrir módulo estaciones]
    S3 --> S3A[Opcional: usar boton Inactivar carritos de estaciones]
    S3A --> S3B[Aplicar desactivar + cerrar en carritos EST-empresa-*]
    S3B --> S4[Tarjetas inician inactivas]
    S3 --> S4[Tarjetas inician inactivas]
    S4 --> S5[Usuario selecciona estación]
    S5 --> S6[Activar carrito de estación y registrar activado_en]
    S6 --> S7[Tarjeta activa muestra color configurado y fecha/hora de entrada]
    S7 --> S7A[Motor de tarifa diaria recalcula deuda por dias segun check-in/check-out]
    S7A --> S8[Finalizar compra en carrito de estación]
    S8 --> S9[Marcar carrito inactivo/cerrado]
    S9 --> S10[Tarjeta vuelve a estado inactivo y oculta fecha/hora]

    AUTHZ --> M
    AUTHZ --> N
    AUTHZ --> F0
    AUTHZ --> T
    AUTHZ --> G0
    AUTHZ --> Z
    AUTHZ --> AUA0{Cambio critico de permisos en modulo seguridad?}
    AUA0 -->|No| AU0[Ejecutar accion critica autorizada]
    AUA0 -->|Si| AUA1{Existe evidencia de aprobacion trazable?}
    AUA1 -->|No| AUF0[Intento critico denegado por permisos o alcance]
    AUA1 -->|Si| AU0
    AU0 --> AU1[Registrar auditoria no bloqueante]
    AU1 --> AU10[Adjuntar metadata de aprobacion y codigo]
    AU10 --> AU2
    AUF0 --> AUF1[Registrar auditoria de denegacion no bloqueante]
    AUF1 --> AU2
    AU2[Persistir empresa_auditoria_eventos]
    AU2 --> AU3[Consultar auditoria en panel empresa]
    AU3 --> AU31[Aplicar filtros avanzados por codigo_http recurso_id metodo_http recurso endpoint y search]
    AU31 --> AU313[Resolver busqueda full-text FTS con fallback LIKE]
    AU313 --> AU311[Paginar resultados con limit offset y total]
    AU311 --> AU312[Inspeccionar detalle JSON por evento]
    AU312 --> AU32[Exportar trazabilidad filtrada en CSV y JSON]
    AU312 --> AU321[Exportar forense action=export_forense en JSON o CSV]
    AU321 --> AU322[Calcular hash_registro y hash_cadena por orden cronologico]
    AU322 --> AU323[Emitir hash_global para cadena de custodia basica]
    AU3 --> AU33[Resolver severidad del evento por modulo resultado y codigo_http]
    AU33 --> AU34[Aplicar politica de retencion por modulo y severidad]
    AU34 --> AU4
    AU3 --> AU4[Aplicar retencion manual por dias]
    AU2 --> AU5[Worker programado de retencion]
    AU5 --> AU6[Purgar eventos expirados por fecha_expiracion]

    L --> M[Crear cliente de venta]
    L --> N[Crear bodega y proveedor]
    L --> CM0[Entrar al modulo compras dedicado]
    CM0 --> CM1[Registrar documento en /api/empresa/compras/documentos]
    CM1 --> CP0
    N --> CP0[Ejecutar accion compras emitir_orden/solicitar_aprobacion/aprobar/rechazar/recepcionar_parcial/recepcionar/contabilizar/validar_documentos]
    CP0 --> CP01{Documento requiere aprobacion multinivel?}
    CP01 -->|No| CP1
    CP01 -->|Si| CP02[Solicitar aprobacion y mover a pendiente_aprobacion]
    CP02 --> CP03[Aprobar por nivel o rechazar documento]
    CP03 --> CP1
    CP1{Transicion valida segun estado_actual?}
    CP1 -->|No| CP2[Responder 409 y no registrar evento]
    CP1 -->|Si| CP3[Persistir documento canonico de compra]
    CP3 --> CP31{Accion recepcionar_parcial_compra?}
    CP31 -->|Si| CP32[Persistir recepcion_detalle_json y recepcion_resumen_json]
    CP31 -->|No| CP33
    CP32 --> CP33{Accion validar_documentos?}
    CP33 -->|Si| CP34[Validar proveedor-factura-entrada y actualizar validacion_documental_estado]
    CP33 -->|No| CP35[Registrar evento con entidad_id canonico]
    CP34 --> CP35
    L --> VC0[Entrar al modulo ventas extendidas]
    VC0 --> VC1[Crear o actualizar cotizacion comercial]
    VC1 --> VC2[Convertir cotizacion a pedido action=convertir_pedido]
    VC2 --> VC3[Persistir pedido con referencia cotizacion_id]
    VC3 --> VC4[Convertir a documento final action=convertir_documento_final]
    VC4 --> VC5[Persistir documento final en empresa_facturacion_documentos]
    VC5 --> VC6[Consultar embudo comercial action=embudo]
    VC6 --> VC7[Monitorear SLA y alertas de vencimiento por etapa]
    VC7 --> VC8[Exportar embudo en JSON CSV TXT XLS y PDF]
    L --> CB0[Entrar a modulo combos_productos]
    CB0 --> CB1[Definir combo con precio unico y receta de ingredientes]
    CB1 --> CB2[Guardar combo en /api/empresa/combos_productos]
    CB2 --> CB3[Agregar item tipo combo en carrito_de_compras]
    CB3 --> CB4[Descontar inventario por ingredientes del combo]
    CB4 --> Q
    L --> CD0[Entrar a modulo codigos_de_descuento]
    CD0 --> CD1[Crear codigo con valor y fecha de vencimiento]
    CD1 --> CD2[Guardar codigo en /api/empresa/codigos_de_descuento]
    CD2 --> CD3[Usar codigo en cierre de carrito de estacion]
    CD3 --> R
    L --> PR0[Entrar al modulo propinas]
    PR0 --> PR1[Configurar habilitacion porcentaje modo y aplicacion automatica]
    PR1 --> PR2[Guardar configuracion en /api/empresa/propinas]
    PR2 --> PR21[Configurar reglas fiscales por pais regimen e impuesto]
    PR21 --> PR22[Registrar ajuste manual auditado action=ajuste_manual]
    PR22 --> PR23[Conciliar manualmente por cierre action=conciliacion_cierre]
    PR23 --> PR3[Consultar reporte por rango usuario y modo]
    PR3 --> PR4[Visualizar resumen acumulado por usuario y movimientos]
    L --> CO0[Entrar al modulo comisiones]
    CO0 --> CO1[Configurar comision por servicio filtro y aplicacion automatica]
    CO1 --> CO2[Guardar configuracion en /api/empresa/comisiones]
    CO2 --> CO21[Configurar escalas y topes por rol/servicio]
    CO21 --> CO22[Registrar ajuste manual action=ajuste_manual]
    CO22 --> CO23[Aprobar o rechazar ajuste action=aprobar_ajuste/rechazar_ajuste]
    CO23 --> CO24[Consultar resumen para nomina action=resumen_liquidacion]
    CO24 --> CO3[Consultar reporte por rango y lavador]
    CO3 --> CO4[Visualizar acumulado por lavador y movimientos]
    L --> N0[Administrar categorias de productos por empresa]
    N0 --> N1[Crear/editar/activar categorias]
    N1 --> N2[Asignar categoria al producto desde selector]
    N2 --> O
    M --> O[Crear carrito]
    N --> O
    O --> P[Agregar items al carrito]
    O --> O1[Configurar lector de barras por empresa]
    O1 --> P0[Escanear codigo de barras o SKU]
    P0 --> P
    P --> P1[Descontar inventario por item producto agregado]
    P1 --> Q[Calcular totales]
    Q --> R[Pagar carrito]
    R --> R0[Validar metodo de pago efectivo/tarjeta/transferencia/codigo_descuento/mixto]
    R0 --> R0A[Consultar configuracion operativa de cobro por empresa y rol]
    R0A --> R0B{Metodo de pago habilitado para el rol?}
    R0B -->|No| R0C[Rechazar cierre de cobro por politica operativa]
    R0B -->|Si| R1[Validar referencia para tarjeta o transferencia y vigencia/usos de codigo]
    R1 --> R11[Consultar configuracion de propinas y politica operativa del rol]
    R11 --> R12{Propina habilitada y aplicada en el cobro?}
    R12 -->|Si| R13[Calcular monto de propina y total final]
    R12 -->|No| R14[Conservar total sin propina]
    R13 --> R15[Validar total_pagado contra total final]
    R14 --> R15
    R15 --> RC0[Consultar configuracion de comisiones por servicio y politica operativa]
    RC0 --> RC1{Comisiones habilitadas y automaticas?}
    RC1 -->|Si| RC2[Filtrar items de servicio segun criterio de lavado]
    RC2 --> RC21[Resolver escala por rol/servicio y aplicar tope]
    RC21 --> RC3[Calcular comision por item y asignar lavador]
    RC3 --> RC4[Registrar movimientos de comision por servicio]
    RC1 -->|No| RC5[Omitir registro automatico de comision]
    RC4 --> R2[Conservar descuento de inventario al cerrar venta]
    RC5 --> R2
    R2 --> S[Cerrar carrito y guardar resumen de pago]
    S --> S11[Registrar evento contable de venta]
    S11 --> RP0[Entrar al modulo reportes]
    RP0 --> RF0[Consultar tablero financiero-operativo action=tablero]
    RF0 --> RF1[Renderizar KPI operativos financieros y contables]
    RF1 --> RF2[Renderizar estado de resultados y balance general]
    RF2 --> RF3[Exportar tablero unificado CSV JSON por rango]
    RP0 --> RPD0[Consultar catalogo de datasets action=catalogo]
    RPD0 --> RPD1[Seleccionar dataset empresarial operativo o contable]
    RPD1 --> RPD2[Cargar dataset action=dataset]
    RPD2 --> RPD3[Renderizar tabla profesional y resumen del dataset]
    RPD3 --> RPD4[Exportar dataset JSON CSV TXT XLS]
    RPD3 --> RPD5[Exportar suite completa action=export format=json]
    RP0 --> RP1[Filtrar ventas cerradas por rango de fechas]
    RP1 --> RP2[Calcular KPIs: ventas, ingresos y ticket promedio]
    RP2 --> RP3[Construir reporte de ventas por fecha]
    RP3 --> RP31[Construir reporte de productos por fecha]
    RP31 --> RP32[Construir reporte de compras de productos por fecha]
    RP32 --> RP33[Consultar inventario actual por bodega y KPI bajo minimo]
    RP33 --> RP4[Consultar configuracion de impresion vigente]
    RP4 --> RP5[Visualizar vista previa de formato POS/Carta]

    L --> OC0[Entrar a configuracion operativa de cobro]
    OC0 --> OC1[Guardar regla base por empresa]
    OC1 --> OC2[Guardar override por rol action=rol]
    OC2 --> OC3[Guardar politica contextual action=politica]
    OC3 --> OC4[Simular reglas por contexto action=simular]
    OC4 --> OC5[Consultar historial y aplicar rollback action=historial/rollback]

    L --> CA0[Entrar a modulo calculadora por empresa]
    CA0 --> CA1[Consultar configuracion e integrar referencias action=config/referencias]
    CA1 --> CA2[Registrar operacion con etiquetas y asociaciones cliente/documento]
    CA2 --> CA3[Filtrar historial por rango y usuario]
    CA3 --> CA4[Exportar historial action=export en JSON CSV TXT XLS PDF]
    CA3 --> CA5[Limpiar historial por filtros action=limpiar]

    L --> CR0[Entrar a modulo creditos]
    CR0 --> CR0A[Menu valida visibilidad de linkCreditos segun permisos finanzas]
    CR0A --> CR0B[Cargar vista creditos.html con contexto empresa_id]
    CR0B --> CR0C[Consultar cartera y resumen inicial en panel]
    CR0C --> CR0D[Consultar limites por cliente action=limites_cliente o limite_cliente]
    CR0D --> CR0E[Gestionar limite por cliente action=upsert_limite_cliente o eliminar_limite_cliente]
    CR0E --> CR0F[Registrar auditoria no bloqueante de limites]
    CR0F --> CR1[Crear credito por cliente con cupo plazo y tasas]
    CR1 --> CR1A[Validar limite de cliente por saldo total y creditos activos]
    CR1A -->|No cumple| CR1B[Rechazar alta o edicion de credito por limite excedido]
    CR1B --> CR8
    CR1A -->|Cumple| CR2[Generar cuotas automaticas y plan de amortizacion base]
    CR2 --> CR3[Registrar abono total o parcial action=abono]
    CR3 --> CR4[Aplicar pago a cuotas pendientes y actualizar saldo]
    CR4 --> CR5[Calcular dias de mora y clasificacion de cartera]
    CR5 --> CR6[Consultar estado de cuenta action=estado_cuenta]
    CR6 --> CR7[Consultar resumen cartera action=resumen_cartera]
    CR7 --> CR71[Gestionar estado de credito y acciones de abono desde tabla]
    CR71 --> CR72[Consultar alertas action=alertas_mora con dias_proximos y top]
    CR72 --> CR73[Renderizar proximos a vencer vencidos y ranking avanzado]
    CR73 --> CR74[Exportar reporte de morosidad action=reporte tipo=morosidad]
    CR74 --> CR75[Registrar evento contable de abono action=abono modulo creditos]
    CR75 --> CR76[Clasificar canal de pago caja bancos o pasarela y persistir payload contable]
    CR76 --> CR77{Politica procesar_asientos habilitada?}
    CR77 -->|Si| CR78[Procesar asientos pendientes por politica asientos_limit max_reintentos]
    CR77 -->|No| CR8[Exportar reporte de cartera en JSON CSV TXT XLS PDF]
    CR78 --> CR8[Exportar reporte de cartera en JSON CSV TXT XLS PDF]
    CR4 --> CR79[Solicitar workflow de reverso o refinanciacion]
    CR79 --> CR80[Registrar solicitud en empresa_creditos_workflow con estado pendiente_aprobacion]
    CR80 --> CR80A{Rol puede decidir el tipo de workflow?}
    CR80A -->|No| CR80B[Denegar accion y registrar auditoria por permiso fino]
    CR80B --> CR8
    CR80A -->|Si| CR81{Aprobacion multinivel completada?}
    CR81 -->|No| CR82[Conservar pendiente y acumular historial de aprobaciones]
    CR82 --> CR8
    CR81 -->|Si| CR83[Ejecutar workflow]
    CR83 --> CR84{Tipo workflow}
    CR84 -->|reverso_abono| CR85[Crear movimiento reverso y recalcular saldo/cuotas]
    CR84 -->|refinanciacion| CR86[Inactivar cuotas pendientes y regenerar nuevo plan]
    CR85 --> CR87[Marcar workflow ejecutada y persistir resultado_json]
    CR86 --> CR87[Marcar workflow ejecutada y persistir resultado_json]
    CR87 --> CR87A[Registrar auditoria ampliada de solicitud y decision]
    CR87A --> CR8

    L --> BK0[Entrar a modulo backups empresariales]
    BK0 --> BK1[Menu valida visibilidad de linkBackups segun permisos de seguridad]
    BK1 --> BK2[Cargar vista backups.html con contexto empresa_id]
    BK2 --> BK3[Crear snapshot action=crear con filtros include/exclude tablas]
    BK3 --> BK4[Persistir metadata y snapshot_json en empresa_backups]
    BK4 --> BK5[Listar historial action=listar con busqueda y paginacion]
    BK5 --> BK6[Consultar detalle action=detalle include_snapshot]
    BK6 --> BK7[Exportar backup action=export en JSON CSV TXT XLS PDF]
    BK7 --> BK8{Se solicita restauracion?}
    BK8 -->|No| BK9[Conservar estado operativo y trazabilidad del backup]
    BK8 -->|Si| BK10[Validar accion de aprobacion para restaurar]
    BK10 --> BK11[Restaurar filas por tabla y registrar empresa_backups_restauraciones]
    BK11 --> BK12[Actualizar restaurado_en/restaurado_por y publicar resumen]
    BK12 --> BK13[Activar o desactivar snapshot action=activar/desactivar]

    L --> GX0[Entrar a modulo graficos_estadisticas]
    GX0 --> GX1[Consultar action=panel en /api/empresa/graficos_estadisticas]
    GX1 --> GX11[Aplicar filtros avanzados por sucursal_id estacion_id y segmento]
    GX11 --> GX12[Resolver cache de panel y estado cache_hit o cache_miss]
    GX12 --> GX13{Comparar periodo habilitado?}
    GX13 -->|Si| GX14[Calcular comparativo automatico o personalizado por rango]
    GX13 -->|No| GX2
    GX14 --> GX2
    GX2 --> GX21[Compactar series por buckets para rangos largos]
    GX2 --> GX3[Visualizar distribuciones de stock y asistencia]
    GX3 --> GX4[Analizar rankings top productos y top clientes]
    GX4 --> GX5[Ajustar rango top N y refresco sin cache para seguimiento directivo]

    L --> F0[Entrar al modulo finanzas]
    F0 --> F1[Consultar configuracion financiera por empresa]
    F1 --> F2[Definir categorias, prefijos, formato y plan de cuentas contable]
    F2 --> F201[Consultar plantillas contables action=plantillas]
    F201 --> F202[Aplicar plantilla por tipo de empresa action=aplicar_plantilla]
    F202 --> F203[Persistir cuentas y metadatos de plantilla]
    F2 --> F21[Configurar destino externo: generico, SIIGO, World Office o Alegra]
    F21 --> F22[Gestionar periodo contable: abierto/cerrado]
    F22 --> F228[Solicitar cerrar/reabrir periodo con evidencia de autorizacion]
    F228 --> F229[Validar autorizado_por motivo_autorizacion y evidencia_autorizacion]
    F229 --> F221
    F229 --> F222
    F22 --> F221[Validar estado de cierre action=validar_cierre_periodo]
    F22 --> F222[Conciliar CxC/CxP contra pagos reales action=conciliar_pagos]
    F222 --> F223[Cruzar cartera con empresa_finanzas_movimientos]
    F223 --> F224[Actualizar saldo estado_cartera y trazabilidad de conciliacion]
    F221 --> F225{Periodo cerrado?}
    F225 -->|Si| F226[Bloquear crear editar estado o eliminar en CxC/CxP]
    F225 -->|No| F227[Permitir operacion contable]
    F2 --> I210[Gestionar lotes y series action=validar_disponibilidad]
    I210 --> I211{Lote vencido en venta o reserva?}
    I211 -->|Si| I212[Marcar lote vencido y bloquear venta/reserva automaticamente]
    I211 -->|No| I213[Ejecutar operacion reserva venta o liberacion]
    I213 --> I214[Registrar movimiento en inventario_lotes_series_movimientos]
    N --> DP0[Registrar devolucion a proveedor]
    DP0 --> DP1[Contabilizar devolucion action=contabilizar]
    DP1 --> DP2[Crear movimiento financiero ingreso]
    DP2 --> DP3[Registrar evento contable devolucion_proveedor_contabilizada]
    DP3 --> DP4[Procesar asientos y marcar devolucion como contabilizada]
    L --> RHH0[Entrar a modulo RRHH vacaciones y licencias]
    RHH0 --> RHH1[Registrar solicitud de vacacion o licencia]
    RHH1 --> RHH2[Calcular saldo/acumulado action=resumen_saldo]
    RHH2 --> RHH3[Iniciar aprobacion action=solicitar_aprobacion]
    RHH3 --> RHH4[Ejecutar aprobacion por niveles action=aprobar]
    RHH4 --> RHH5{Aprobacion final alcanzada?}
    RHH5 -->|No| RHH4
    RHH5 -->|Si| RHH6[Persistir snapshot de saldo y aprobadores]
    RHH6 --> RHH7[Vincular a nomina action=vincular_nomina]
    RHH7 --> RHH8[Marcar novedad contabilizada con periodo y liquidacion]
    L --> P230[Entrar al modulo produccion]
    P230 --> P231[Consultar plan de capacidad action=plan_capacidad]
    P231 --> P232[Consolidar cantidad programada/producida por orden y por dia]
    P232 --> P233[Comparar contra meta diaria y calcular desviacion]
    P233 --> P234[Generar alertas de atraso o sobrecapacidad]
    L --> LG230[Entrar al modulo logistica]
    LG230 --> LG231[Consultar seguimiento action=seguimiento_hitos]
    LG231 --> LG232[Evaluar hitos fecha_programada fecha_salida fecha_entrega]
    LG232 --> LG233[Calcular cumplimiento SLA y alertas de incumplimiento]
    LG233 --> LG234[Priorizar envios criticos por alerta]
    L --> DOC240[Entrar al modulo documental]
    DOC240 --> DOC241[Consultar repositorio action=repositorio]
    DOC241 --> DOC242[Validar acceso por rol y modulo action=acceso]
    DOC242 --> DOC243[Versionar documento action=versionar]
    DOC243 --> DOC244[Consultar historial versionado action=versiones]
    L --> INT240[Entrar al modulo integraciones]
    INT240 --> INT241[Rotar referencia segura action=rotar_credencial]
    INT241 --> INT242[Ejecutar health_check o sync_manual]
    INT242 --> INT243[Monitorear conectores action=monitoreo]
    INT243 --> INT244[Generar alertas por endpoint conectividad latencia y sincronizacion]
    F22 --> C0[Gestionar cierre de caja por sucursal y caja]
    C0 --> C1[Abrir caja con base inicial de efectivo]
    C1 --> C2[Registrar ingresos, egresos y retiros del turno]
    C2 --> C3[Calcular caja teorica]
    C3 --> C4[Capturar arqueo de caja fisica]
    C4 --> C5{Diferencia supera umbral?}
    C5 -->|Si| C6[Marcar incidencia para revision]
    C5 -->|No| C7[Cerrar caja]
    C6 --> C7
    C7 --> C71[Conciliar propinas del turno contra cierre de caja]
    C71 --> C8[Aprobar cierre por rol autorizado]
    F22 --> F3[Registrar ingreso o egreso con comprobante]
    F3 --> F31[Calcular total bruto, retenciones y total neto]
    F31 --> F4[Filtrar movimientos y calcular balance]
    F4 --> FA0[Consultar eventos contables pendientes]
    FA0 --> FA00[Worker automatico por intervalo y politica]
    FA00 --> FA1
    FA0 --> FA1[Ejecutar action=procesar_asientos por lote]
    FA1 --> FA2[Persistir empresa_asientos_contables con hash_idempotencia]
    FA2 --> FA3[Marcar evento procesado o error con intentos]
    FA3 --> FA4[Consultar action=conciliacion_periodo]
    FA4 --> FA5[Comparar eventos procesados vs asientos por periodo]
    FA5 --> FA6[Clasificar estado conciliado o con alertas]
    FA6 --> F41
    F4 --> FB0[Importar extractos bancarios action=importar_extractos_bancarios]
    FB0 --> FB1[Persistir extractos en empresa_finanzas_bancos_movimientos por hash idempotente]
    FB1 --> FB2[Ejecutar conciliacion automatica action=conciliar_bancaria_auto]
    FB2 --> FB3[Emparejar extractos vs movimientos internos por referencia monto y fecha]
    FB3 --> FB4[Clasificar extractos en pendiente conciliado o con_desviacion]
    FB4 --> FB5[Consultar tablero action=conciliacion_bancaria]
    FB5 --> FB6[Exportar desviaciones en JSON CSV TXT XLS y PDF]
    F4 --> F41[Navegar pestañas: Todos, Ingresos o Egresos]
    F41 --> F42[Exportar libro filtrado]
    F42 --> F421[Excel CSV y PDF]
    F42 --> F43[Generar asientos con cuentas parametrizadas por empresa y categoria]
    F43 --> F44[Incluir perfil contable y proyeccion ERP en JSON]
    F42 --> F45[Generar plantilla dedicada SIIGO CSV]
    F42 --> F46[Generar balance de prueba CSV]
    F42 --> F47[Generar estado de resultados CSV]
    F42 --> F48[Generar libro diario, libro mayor y balance general CSV]
    F42 --> F5[Imprimir comprobante financiero Carta/POS]

    L --> AI0[Entrar al modulo chat con inteligencia artificial]
    AI0 --> AI1[Cargar Gemini empresarial para la empresa seleccionada]
    AI1 --> AI15[Resolver cuenta Google autenticada y modelo preferido]
    AI15 --> AI16[Registrar preferencia Gemini para la cuenta Google]
    AI16 --> AI2[Validar alcance por empresa_id y limite diario free-tier]
    AI2 --> AI3[Construir contexto empresarial desde la base de datos]
    AI3 --> AI4[Consultar proveedor IA externo Google Gemini]
    AI4 --> AI5[Registrar pregunta/respuesta y acumular uso diario]
    AI5 --> AI6[Mostrar respuesta o sugerir upgrade de plan]

    L --> T[Administrador abre modulo chat_y_tareas]
    T --> U[Crear conversacion por empresa]
    U --> V[Agregar participantes de la empresa]
    V --> W[Intercambiar mensajes y adjuntos foto/voz]
    W --> X[Crear tareas vinculadas a la conversacion]
    X --> Y[Actualizar avance: pendiente/en_progreso/completada]

    L --> G0[Entrar a modulo ubicacion_gps]
    G0 --> G1[Registrar dispositivos GPS por empresa]
    G1 --> G2[Mostrar dispositivos en mapa OpenStreetMap]
    G2 --> G3[Iniciar tracking automatico cada 10 segundos]
    G3 --> G4[Guardar punto en empresa_gps_recorridos]
    G4 --> G5[Actualizar ultima posicion en empresa_gps_dispositivos]
    G5 --> G6[Visualizar recorrido historico por dispositivo]

    L --> AS0[Entrar a modulo asistencia de empleados]
    AS0 --> AS1[Registrar asistencia diaria por empleado]
    AS1 --> AS2[Marcar entrada del turno]
    AS2 --> AS3[Marcar salida del turno]
    AS3 --> AS4[Calcular horas trabajadas y consolidar novedad]
    AS4 --> AS5[Consultar asistencia por rango y estado]

    L --> VH0[Entrar a modulo registro de vehiculos]
    VH0 --> VH1[Registrar ingreso con patente tipo y motivo]
    VH1 --> VH2[Guardar registro en /api/empresa/vehiculos_registro]
    VH2 --> VH3[Consultar registros por patente estado y rango]
    VH3 --> VH4[Marcar salida del vehiculo cuando abandona la empresa]
    VH4 --> VH5[Actualizar estado_registro en_empresa o retirado]

    L --> RH0[Entrar a modulo reservas por estacion]
    RH0 --> RH1[Consultar disponibilidad por fecha de entrada y salida]
    RH1 --> RH2[Crear reserva pendiente_pago para estacion disponible]
    RH2 --> RH3[Editar reserva pendiente o cambiar estacion]
    RH3 --> RH4[Confirmar pago o cancelar reserva]
    RH4 --> RH5[Actualizar disponibilidad segun estado final de reserva]

    L --> TPM0[Entrar a modulo tarifas por minutos]
    TPM0 --> TPM1[Seleccionar estacion y rango de dias operativos]
    TPM1 --> TPM2[Definir tarifa base y bloque adicional por minutos]
    TPM2 --> TPM3[Guardar regla en /api/empresa/tarifas_por_minutos]
    TPM3 --> TPM4[Simular cobro segun minutos consumidos y dia de semana]

    L --> TPD0[Entrar a modulo tarifas por dia]
    TPD0 --> TPD1[Seleccionar estacion y servicio hotelero]
    TPD1 --> TPD2[Configurar valor por dia y horarios check-in/check-out]
    TPD2 --> TPD3[Definir aplicacion automatica en carritos activos]
    TPD3 --> TPD4[Guardar regla en /api/empresa/tarifas_por_dia]
    TPD4 --> TPD5[Simular deuda por rango de fechas activado_en-fecha_corte]

    L --> Z[Entrar a modulo facturacion electronica]
    Z --> Z1[Detectar pais automaticamente tz/lang/config empresa]
    Z1 --> Z2[Mostrar bandera del pais detectado en menu flotante]
    Z2 --> Z3[Configurar parametros FE por pais CO/PA/EC]
    Z3 --> Z31[Si no existe configuracion FE, prellenar desde configuracion avanzada]
    Z31 --> Z4[Guardar configuracion por empresa y pais]
    Z4 --> Z5[Ejecutar accion transaccional emitir/anular/nota_credito]
    Z5 --> Z6{Transicion valida segun estado_actual?}
    Z6 -->|No| Z7[Responder 409 y no registrar evento]
    Z6 -->|Si| Z8[Persistir documento canonico de facturacion]
    Z8 --> Z81[Despachar integracion fiscal por pais/proveedor]
    Z81 --> Z82{Envio fiscal exitoso?}
    Z82 -->|Si| Z83[Actualizar cola FE a enviado y registrar referencia externa]
    Z82 -->|No| Z84[Registrar fallo e incrementar intentos en cola FE]
    Z84 --> Z85{Supera max_intentos?}
    Z85 -->|Si| Z86[Activar contingencia FE estado_envio=contingencia]
    Z85 -->|No| Z87[Programar proximo_intento para reintento automatico]
    Z83 --> Z8A{Accion emitir factura_electronica?}
    Z86 --> Z8A
    Z87 --> Z8A
    Z8A -->|No| Z9
    Z8A -->|Si| Z8B[Resolver destinatario cliente por cliente_id o cliente_email]
    Z8B --> Z8C{SMTP configurado y envio exitoso?}
    Z8C -->|Si| Z8D[Responder factura_email enviado=true]
    Z8C -->|No| Z8E[Responder factura_email con error sin bloquear emision]
    Z8D --> Z9[Registrar evento con entidad_id canonico]
    Z8E --> Z9

    L --> ZQ0[Entrar a modulo facturas electronicas]
    ZQ0 --> ZQ1[Aplicar filtros por cliente documento estado y fecha]
    ZQ1 --> ZQ2[Consultar /api/empresa/facturacion_electronica action=documentos]
    ZQ2 --> ZQ3[Visualizar listado y detalle documental]
    ZQ3 --> ZQ4[Reenviar correo de factura action=reenviar_correo]
    ZQ3 --> ZQ5[Imprimir factura desde vista de detalle]

    L --> ZR0[Operar cola FE por empresa]
    ZR0 --> ZR1[Consultar reintentos action=reintentos]
    ZR1 --> ZR2[Procesar cola vencida action=procesar_reintentos]
    ZR2 --> ZR3[Reconciliar estados action=reconciliacion/reconciliar_estados]

    L --> ZD0[Entrar a modulo DIAN Colombia]
    ZD0 --> ZD1[Configurar NIT software y credenciales ref seguras]
    ZD1 --> ZD2[Firmar XML action=firmar_xml_real]
    ZD2 --> ZD3[Enviar documento action=enviar_documento_real]
    ZD3 --> ZD4{DIAN responde acuse aceptado?}
    ZD4 -->|Si| ZD5[Actualizar estado_dian aceptado]
    ZD4 -->|No| ZD6[Marcar contingencia y registrar observacion]
    ZD5 --> ZD7[Consultar acuse action=consultar_acuse_real]
    ZD6 --> ZD8[Ejecutar reconexion action=reconexion_dian]
    ZD8 --> ZD9{Conectividad restablecida?}
    ZD9 -->|Si| ZD10[Estado reconectado y opcion de reenvio]
    ZD9 -->|No| ZD11[Mantener contingencia y monitoreo]
```

Resultado esperado:
- Flujo completo desde onboarding de usuario hasta cierre de venta con carrito pagado.
- Flujo colaborativo interno por empresa para comunicacion operativa y seguimiento de tareas.
- Al abrir el Panel de la Empresa, la subpagina inicial predeterminada es Inicio.
- En `super/licencias_resumen`, el conteo refleja solo licencias activas asignadas a empresa.
- En `seleccionar_empresa`, la seccion de licencias se filtra por empresas creadas por el usuario autenticado.
- El sistema detecta país para facturación electrónica y muestra su bandera en el menú flotante.
- En `estaciones`, los carritos de estación inician inactivos, se activan al seleccionar la estación y vuelven a inactivos al finalizar la compra.
- En `estaciones`, la tarjeta activa muestra fecha y hora de entrada (`activado_en`), y las inactivas no muestran esa marca.
- En `configuracion_de_estaciones`, existe accion manual para forzar inactivacion/cierre masivo de carritos de estaciones.
- En `administrar_empresa`, `super_administrador` y `seleccionar_empresa`, al recargar con F5 se restaura la subpagina/vista que estaba abierta.
- En `administrar_productos`, el catálogo de `categorias_productos` permite filtrar y asignar categorías de forma consistente por `empresa_id`.
- En `ubicacion_gps`, cada dispositivo puede registrar su recorrido automaticamente cada 10 segundos y visualizarse sobre mapa de codigo abierto.
- En `asistencia_empleados`, el sistema registra entrada/salida por empleado y calcula horas trabajadas por jornada.
- En `asistencia_empleados`, la consulta por rango/estado permite control operativo diario y trazabilidad de novedades.
- En `vehiculos_registro`, la empresa puede registrar ingresos por patente, conductor y motivo operativo por `empresa_id`.
- En `vehiculos_registro`, la operacion `marcar_salida` actualiza fecha de salida y estado del vehiculo a `retirado` para trazabilidad de permanencia.
- En `reservas_hotel`, la empresa puede consultar disponibilidad por rango y crear reservas por estacion con datos de cliente, monto y fechas de entrada/salida.
- En `reservas_hotel`, el flujo operativo permite editar reservas pendientes, confirmar pago, cancelar y activar/desactivar registros segun politica operativa.
- En `reservas_hotel`, el backend evita solapamientos de rango para la misma estacion y expira reservas `pendiente_pago` fuera de vigencia.
- En `facturacion_electronica`, la accion `emitir` intenta envio automatico al correo del cliente y reporta el resultado en `factura_email` sin bloquear la emision legal.
- En `facturacion_electronica`, cada accion transaccional (`emitir`, `anular`, `nota_credito`) retorna `integracion_fiscal` con estado de envio, intentos, contingencia y referencia externa.
- En `facturacion_electronica`, existe cola de reintentos por documento (`action=reintentos`, `action=procesar_reintentos`) para reenvio fiscal controlado por `proximo_intento`.
- En `facturacion_electronica`, la reconciliacion (`action=reconciliacion`/`action=reconciliar_estados`) cruza documentos canonicos vs cola FE y permite aplicar sincronizacion.
- En `facturacion_electronica/dian`, la accion `firmar_xml_real` aplica firma digital RSA-SHA256 sobre XML de factura usando referencia segura de llave (`certificado_clave_ref`).
- En `facturacion_electronica/dian`, las acciones `enviar_documento_real` y `consultar_acuse_real` permiten despacho y trazabilidad de acuse DIAN por documento/cufe.
- En `facturacion_electronica/dian`, `reconexion_dian` permite salir de contingencia al recuperar conectividad y opcionalmente ejecutar reenvio controlado.
- En `inventario/lotes_series`, las operaciones `reservar`, `vender` y `liberar_reserva` registran trazabilidad por lote en `inventario_lotes_series_movimientos`.
- En `inventario/lotes_series`, cuando un lote esta vencido el backend bloquea automaticamente venta/reserva y marca `estado_lote=vencido` con `bloqueado_venta=1`.
- En `compras/devoluciones_proveedor`, la accion `contabilizar` genera movimiento financiero, evento contable y actualiza la devolucion a estado `contabilizada`.
- En `produccion/ordenes`, la accion `action=plan_capacidad` consolida carga planificada vs producida, calcula desviacion contra meta diaria y expone alertas por atraso/sobrecapacidad.
- En `logistica/envios`, la accion `action=seguimiento_hitos` evalua hitos de programacion/salida/entrega, mide cumplimiento SLA y publica alertas de incumplimiento.
- En `facturas_electronicas`, la empresa puede buscar por cliente/documento/rango de fechas, ver detalle documental, reenviar factura por correo e imprimir comprobante.
- En `combos_productos`, el usuario puede crear combos con precio unico y receta de ingredientes por empresa.
- En `carrito_de_compras`, al agregar un item `tipo_item=combo`, el sistema descuenta inventario por ingrediente y revierte correctamente en inactivacion/eliminacion del item.
- En `codigos_de_descuento`, el usuario puede crear codigos promocionales con generacion automatica, fecha de vencimiento y limite de usos por empresa.
- En `carrito_de_compras`, el cierre de venta valida en backend los metodos `efectivo`, `tarjeta_credito`, `tarjeta_debito`, `transferencia_bancaria`, `mixto` y `codigo_descuento`, exigiendo referencia para tarjetas y transferencias bancarias.
- En `carrito_de_compras`, antes de cerrar la venta se resuelve la configuracion operativa de cobro por `empresa_id + rol + contexto` (`canal_venta`, `sucursal_id`, `turno`), bloqueando metodos no habilitados y desactivando propina/comision cuando la politica contextual lo define.
- En `configuracion_operativa`, el administrador puede publicar reglas base, overrides por rol y politicas contextuales; ademas puede ejecutar simulaciones previas y restaurar snapshots mediante rollback operativo.
- En `calculadora`, la empresa puede registrar operaciones con etiquetas y asociaciones a cliente/documento/carrito/cotizacion, consultar historial por rango/usuario y exportar en `PDF`, `XLS`, `CSV`, `JSON` y `TXT` manteniendo estructura y totales consistentes.
- En `propinas`, la empresa puede activar/desactivar porcentaje de propina, definir modo `por_usuario` o `universal` y consultar reporte de acumulados y movimientos.
- En `carrito_de_compras`, al cerrar venta en estacion, el backend calcula el total final con propina (si aplica), valida `total_pagado` contra ese total y registra el movimiento de propina.
- En `graficos_estadisticas`, el usuario visualiza series por dia de ventas, finanzas, compras y asistencia por `empresa_id`.
- En `graficos_estadisticas`, el panel muestra distribuciones y rankings para priorizar decisiones operativas/comerciales por rango.
- En `reportes`, el usuario consulta ventas cerradas por rango, indicadores clave y top comerciales, con validacion visual de formato de impresion POS/Carta.
- En `carrito_de_compras`, al agregar items de tipo producto se descuenta inventario y, al cerrar la venta, el descuento se mantiene aplicado.
- En `reportes`, se dispone de reportes profesionales por rango de fechas: ventas, productos y compras de productos.
- En `reportes`, el tablero financiero-operativo puede exportarse en formato unificado `CSV/JSON` por rango, incluyendo `estado_resultados` y `balance_general`.
- En `reportes`, existe un centro profesional de datasets por empresa con selector de nivel (`empresarial`, `operativo`, `contable`) y vista tabular dinamica.
- En `reportes`, los datasets se exportan en `JSON`, `CSV`, `TXT` y `XLS`, y la suite consolidada se exporta en `JSON`.
- En `reportes`, el dataset `operativo_cadena_cumplimiento` incorpora metas y desviaciones por dominio (`meta_cumplimiento_pct`, `desviacion_meta_pct`, `estado_meta`) y resumen global de brecha (`meta_global_pct`, `desviacion_meta_global_pct`).
- En `ayuda`, existe un menu interno con accesos rapidos y una seccion de APIs principales para operacion diaria.
- En `configuracion`, las opciones del lector de barras se gestionan por empresa y aplican al flujo operativo del carrito.
- En `reportes`, se agrega tabla de inventario actual por bodega y KPI de productos bajo minimo.
- En `finanzas`, cada empresa administra ingresos y egresos con configuracion propia, comprobantes y soporte de impresion.
- En `finanzas`, el flujo de caja operativo permite apertura, arqueo, cierre y aprobacion de caja por `sucursal_id` y `caja_codigo`.
- En `finanzas`, la interfaz operativa de cierres de caja permite ejecutar acciones de ciclo desde la tabla (cerrar, reabrir, aprobar, anular), junto con activacion/desactivacion y filtros por estado/rango.
- En `finanzas/cierres_caja`, la validacion UAT por rol confirma autorizacion esperada: `admin_empresa` aprueba, `cajero` y `supervisor_sucursal` no aprueban bajo politica financiera actual.
- En `finanzas`, el libro financiero se consulta por pestañas (`Todos`, `Ingresos`, `Egresos`) y puede exportarse por rango a Excel (CSV), PDF y JSON contable para integración externa.
- En `finanzas/asientos_contables`, la API permite consultar asientos canonicos (`GET`) y ejecutar procesamiento manual por lotes (`POST/PUT action=procesar_asientos`) con `max_reintentos` opcional.
- En backend, un worker automatico procesa eventos contables pendientes por lotes con politica configurable de intervalo, tamaño de lote y limite de reintentos.
- En `finanzas/asientos_contables`, la API expone `GET action=conciliacion_periodo` para comparar por periodo los eventos contables vs asientos canonicos y detectar pendientes, errores y descuadres.
- En `administrar_empresa/finanzas`, existe vista de conciliacion por periodo con filtros de rango/periodo y KPIs de estado de conciliacion.
- En `finanzas/movimientos`, la API permite importar extractos bancarios (`POST action=importar_extractos_bancarios`) con idempotencia por `hash_movimiento`.
- En `finanzas/movimientos`, la API permite conciliacion bancaria automatica (`PUT action=conciliar_bancaria_auto`) con tolerancia configurable por dias y monto.
- En `finanzas/movimientos`, la API expone `GET action=conciliacion_bancaria` y `GET action=conciliacion_bancaria_export` para tablero de desviaciones por periodo y exportacion `JSON/CSV/TXT/XLS/PDF`.
- En `auditoria/eventos`, la API permite consultar trazabilidad por filtros (`GET`) con `search` full-text (FTS con fallback), aplicar retencion manual (`PUT/POST action=retener|purgar`) y ejecutar exportacion forense (`action=export_forense`) en `json/csv`.
- En `administrar_empresa/auditoria`, la UI permite consultar eventos con filtros de modulo/accion/usuario/request/rango y filtros avanzados por `codigo_http`/`recurso_id`/`search`, exportar resultados operativos a CSV/JSON y ejecutar purga manual por dias de retencion.
- En `auditoria/eventos?action=export_forense`, la salida incluye cadena de custodia basica con `hash_registro`, `hash_cadena` y `hash_global` para validacion de integridad.
- En backend, un worker periodico elimina eventos expirados de auditoria usando `fecha_expiracion` (con fallback por `retencion_dias` para registros legacy).
- En `finanzas`, el JSON contable usa cuentas parametrizadas por empresa/categoria e incluye perfil de referencia para ERP destino.
- En `finanzas`, existe plantilla dedicada SIIGO en CSV y exportaciones de `balance de prueba` y `estado de resultados` para trabajo contable/directivo.
- En `finanzas`, los movimientos quedan asociados a `periodo_contable`; al cerrar un periodo se bloquean edición, activación/desactivación y eliminación hasta reabrir.
- En `finanzas`, cada movimiento calcula total bruto, retenciones (`fuente`, `ICA`, `IVA`) y total neto antes de persistir/exportar.
- En `finanzas`, también se exportan `libro diario`, `libro mayor` y `balance general` en CSV.
- En `chat_con_inteligencia_artificial`, el alcance de consultas queda restringido por `empresa_id` y validacion del usuario autenticado.
- En `chat_con_inteligencia_artificial`, el sistema controla limite free-tier diario por `empresa/proveedor/modelo` y muestra opcion de upgrade cuando aplica.
- En `chat_con_inteligencia_artificial`, cada consulta/respuesta queda auditada junto con metrica de tokens para trazabilidad operativa.
- En rutas criticas de `ventas`, `inventario`, `finanzas`, `clientes`, `compras/proveedores`, `facturacion` y `seguridad/usuarios`, el middleware valida rol y alcance de `empresa_id` antes de ejecutar operaciones sensibles.
- En rutas operativas de `chat_tareas`, `ubicacion_gps` y `productos/imagen`, el middleware aplica politicas por modulo para mantener control uniforme de acceso.
- En `carritos_compra`, cada lectura de carrito expone `estado_venta` estandarizado (`venta_abierta`, `venta_cerrada`, `venta_pagada`, `venta_suspendida`) para normalizar decisiones operativas y reportes.
- En acciones de carrito (`activar_estacion`, `pagar_estacion`, `activar/desactivar`, `cerrar/reabrir`), la API responde `estado_venta` para trazabilidad inmediata del ciclo de venta.
- En acciones de ciclo de venta, el backend bloquea transiciones no permitidas (doble pago, reabrir pagada, activar estacion pagada sin `reset_items=1`) con `409`, y usa `404` cuando el carrito no existe.
- En cierre y cambios operativos de venta, se registra un evento contable (`empresa_eventos_contables`) para habilitar integracion contable por modulo.
- En `reportes`, la vista consume `GET /api/empresa/finanzas/movimientos?action=tablero` para mostrar KPI financieros y contables junto a los KPI operativos.
- En `reportes`, el tablero incluye `estado_resultados` y `balance_general` con base en asientos canonicos procesados por periodo.
- En `finanzas`, la accion `procesar_asientos` requiere permiso de aprobacion (`A`) en el middleware de roles.
- En middleware de permisos por empresa, toda accion critica autorizada (`C/U/D/A`) registra auditoria no bloqueante con modulo, accion, recurso, resultado HTTP y metadatos de trazabilidad.
- En acciones criticas de `ventas`, `compras` y `facturacion`, la auditoria automatica conserva metadata de negocio (`carrito_id`, `proveedor_id`, `entidad_id`, `documento_codigo`) para trazabilidad operacional.
- En `facturacion_electronica`, al guardar configuracion FE por pais se registra evento contable de modulo `facturacion` para trazabilidad de parametrizacion fiscal.
- En `facturacion_electronica`, acciones transaccionales (`emitir`, `anular`, `nota_credito`) registran eventos `factura_emitida`, `factura_anulada` y `nota_credito_emitida`.
- En `proveedores`, las operaciones de alta, actualizacion, activacion/desactivacion y eliminacion registran eventos del modulo `compras`.
- En `proveedores`, acciones transaccionales (`emitir_orden`, `recepcionar_compra`, `contabilizar_compra`) registran eventos de orden y ciclo contable de compra.
- En `administrar_empresa/compras`, el modulo dedicado permite crear, consultar y ejecutar ciclo documental de compras sobre `/api/empresa/compras/documentos`.
- En `administrar_empresa/compras`, el ciclo documental de compras permite solicitud de aprobacion, aprobacion multinivel y rechazo por accion explicita.
- En `administrar_empresa/compras`, la recepcion parcial por item persiste detalle/resumen (`recepcion_detalle_json`, `recepcion_resumen_json`) y consolida recepcion total cuando no quedan pendientes.
- En `administrar_empresa/compras`, la validacion documental cruza proveedor-factura-entrada y actualiza estado documental de validacion antes de contabilizacion.
- En acciones transaccionales de `facturacion_electronica` y `proveedores`, el backend valida `estado_actual` y responde `409` cuando la transicion no corresponde al ciclo documental.
- En transacciones documentales validas, la API devuelve `estado_anterior` y `estado_nuevo`, y los persiste en el payload del evento contable para auditoria.
- En transacciones de facturacion y compras, `empresa_eventos_contables.entidad_id` corresponde al ID canonico persistido en `empresa_facturacion_documentos` o `empresa_compras_documentos`.
- En `finanzas`, el alta de movimientos y el cierre/reapertura de periodos registran eventos contables del modulo `finanzas`.
- La estrategia de asientos contables se define sobre consumo por lotes de eventos pendientes con idempotencia por referencia canonica documental y marcacion de procesamiento por resultado.
- En pruebas de seguridad de endpoints protegidos, se valida extraccion de `empresa_id` desde `multipart/form-data` para `chat_tareas/mensajes/adjunto` y denegacion por rol en `ubicacion_gps/dispositivos`.
- En `super/configuracion_avanzada`, la tarjeta IA permite guardar credenciales de 5 modelos populares y registrar la cuenta Google del administrador que realiza el cambio.
