# P108-015 - Correo corporativo e identidad visual

Fecha: 2026-07-28  
Entorno observado: VPS accesible por `https://powerfulcontrolsystem.com`  
Empresa autorizada: Powerful Control System  
Rol: super administrador autenticado  
Candidato local: rama `codex/p108-email-brand-avatar`, basada en `ddc0976d`

## Alcance

- correo saliente por Mailu propio;
- logo oficial del computador PCS dentro del cuerpo;
- avatar de dominio por BIMI;
- SPF, DKIM y DMARC;
- acciones de prueba y configuración del panel super;
- evidencia visual en navegador interno.

No se registran contraseñas, cookies, tokens, clave DKIM ni cabeceras privadas.

## Criterio

El flujo solo se aprueba cuando el mismo candidato desplegado:

1. permite enviar la prueba con CSRF válido;
2. entrega el correo real;
3. muestra el logo oficial inline;
4. acredita SPF, DKIM y DMARC alineados;
5. demuestra el avatar BIMI en clientes compatibles;
6. dispone de VMC o CMC si el receptor lo exige.

