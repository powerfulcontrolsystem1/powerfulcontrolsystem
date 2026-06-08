# Referencias DIAN Colombia

Esta carpeta documenta las referencias tecnicas oficiales usadas para reparar y
validar el flujo de facturacion electronica Colombia.

## Descarga local 2026-06-08

Los documentos oficiales se descargaron localmente en:

`documentos/referencias/dian/2026-06-08/`

Archivos fuente:

- `Anexo-Tecnico-Factura-Electronica-de-Venta-vr-1-9.pdf`
- `Caja-de-herramientas-FE-V19-V2026.zip`
- `Guia-Herramienta-para-el-Consumo-de-Web-Services.pdf`

La carpeta fechada esta ignorada por Git porque contiene PDFs grandes, ZIPs,
XSDs, Schematron y JARs de soporte publicados por DIAN. Es material de consulta
y validacion local, no dependencia runtime del backend PCS.

Para validacion local contra XSD, usar:

```powershell
.\scripts\validar_dian_xsd.ps1 -XmlPath ruta\documento.xml
```

