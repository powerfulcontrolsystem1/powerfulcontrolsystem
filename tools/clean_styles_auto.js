const fs = require('fs');
const path = require('path');

const filesToClean = [
  "web/administrar_empresa/administrar_productos.html",
  "web/administrar_empresa/asistencia_empleados.html",
  "web/administrar_empresa/auditoria.html",
  "web/administrar_empresa/buscar_producto_botones.html",
  "web/administrar_empresa/carrito_de_compras.html",
  "web/administrar_empresa/chat_con_inteligencia_artificial.html",
  "web/administrar_empresa/chat_y_tareas.html",
  "web/administrar_empresa/codigos_de_descuento.html",
  "web/administrar_empresa/combos_productos.html",
  "web/administrar_empresa/configuracion.html",
  "web/administrar_empresa/configuracion_de_estaciones.html",
  "web/administrar_empresa/creditos.html",
  "web/administrar_empresa/estaciones.html",
  "web/administrar_empresa/estacion_ia_pedidos.html",
  "web/administrar_empresa/facturacion_electronica.html",
  "web/administrar_empresa/facturas_electronicas.html",
  "web/administrar_empresa/finanzas.html",
  "web/administrar_empresa/graficos_estadisticas.html",
  "web/administrar_empresa/historial_productos.html",
  "web/administrar_empresa/publicar_red_social.html",
  "web/administrar_empresa/reportes.html",
  "web/administrar_empresa/reservas_hotel.html",
  "web/administrar_empresa/soporte_remoto.html",
  "web/administrar_empresa/tarifas_por_dia.html",
  "web/administrar_empresa/tarifas_por_minutos.html",
  "web/administrar_empresa/ubicacion_gps.html",
  "web/administrar_empresa/vehiculos_registro.html",
  "web/js/configuracion_de_la_cuenta.js",
  "web/js/login.js",
  "web/js/login_usuario.js",
  "web/js/registrar_contrasena_usuario_de_google.js",
  "web/js/registrar_nuevo_usuario_administrador.js",
  "web/js/super_reportes_globales.js",
  "web/js/super_seguridad.js",
  "web/Juegos/n64/index.html",
  "web/Juegos/menu_juegos.html",
  "web/Juegos/patito_volando.html",
  "web/super/administrar_base_de_datos.html",
  "web/super/chat_con_ia_global.html",
  "web/super/contrato.html",
  "web/super/errores.html",
  "web/super/formato_para_emviar_email.html",
  "web/super/metricas_de_trafico_general.html",
  "web/super/permisos_rol.html",
  "web/super/roles_de_usuario.html",
  "web/super/seguridad.html",
  "web/super/servidores.html",
  "web/super/asesor_comercial.html",
  "web/super/venta_digital.html",
  "web/contrato.html",
  "web/index.html",
  "web/login.html",
  "web/mantenimiento.html",
  "web/pagar_licencia.html",
  "web/red_social_comercial.html",
  "web/soporte_remoto_acceso.html"
];

for (const file of filesToClean) {
  const filePath = path.join(__dirname, '..', file);
  if (!fs.existsSync(filePath)) continue;

  let content = fs.readFileSync(filePath, 'utf-8');
  let originalContent = content;

  if (file.endsWith('.html')) {
    // Basic replacements for known inline styles. We'll strip `style="color: ..."` and similar visually-breaking ones.
    content = content.replace(/style="[^"]*"/g, (match) => {
        // Strip colors
        if (/color|background|rgba|rgb|#[A-Fa-f0-9]/i.test(match)) {
            let replaced = match.replace(/(?:color|background|background-color)\s*:[^;"]*;?/gi, '');
            // If it's just style="", remove it entirely
            if (replaced === 'style=""' || replaced === 'style=" "') return '';
            return replaced;
        }
        return match;
    });

    // Also just aggressively remove 'style="..."' if instructed to heavily centralize,
    // but a targeted approach is safer.
  } else if (file.endsWith('.js')) {
    // e.g. style.color = '#ef5350' => className / logic fix. 
    // This is hard to perfectly script without breaking logic, so I'll write an echo script first.
  }

  if (content !== originalContent) {
    fs.writeFileSync(filePath, content, 'utf-8');
    console.log(`Updated ${file}`);
  }
}
