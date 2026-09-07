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

## Nomina electronica 2026-08-26

Referencias oficiales consultadas:

- [Documento soporte de pago de nomina electronica](https://www.dian.gov.co/impuestos/Paginas/Sistema-de-Factura-Electronica/Documento-Soporte-de-Pago-de-Nomina-Electronica.aspx).
- [Documentacion tecnica](https://micrositios.dian.gov.co/sistema-de-facturacion-electronica/documentacion-tecnica-soporte-de-pago-nomina-electronica/).
- [Caja de herramientas Nomina Electronica V1.0](https://www.dian.gov.co/impuestos/factura-electronica/Documents/Caja-de-Herramientas-Nomina-Electronica-V1-0.zip).
- [ABECE del documento soporte de pago de nomina electronica](https://www.dian.gov.co/impuestos/factura-electronica/Documents/Abece-documento-soporte-de-pago-nomina-electronica.pdf).
- [Memorias DIAN de nomina electronica](https://micrositios.dian.gov.co/sistema-de-facturacion-electronica/memorias-documento-soporte-de-pago-de-nomina-electronica/).

La caja oficial contiene, entre otros, `XSD/NominaIndividualElectronicaXSDV1.0.6.xsd`
y los esquemas comunes bajo `Schemes/`. El ZIP/XSD no se versiona ni se vuelve
una dependencia runtime. La regresion opcional genera y firma el XML durante la
prueba y valida esa salida contra la copia oficial indicada por variables de
entorno:

```powershell
$env:PCS_DIAN_NOMINA_XSD = 'C:\ruta\XSD\NominaIndividualElectronicaXSDV1.0.6.xsd'
$env:PCS_DIAN_NOMINA_SCHEMES = 'C:\ruta\Schemes'
$env:PCS_PYTHON = 'C:\ruta\python.exe'
go test ./handlers -run '^TestDIANNominaIndividualXSDOficialOpcional$' -count=1
```

Esta prueba demuestra conformidad XSD del fixture firmado; no sustituye
habilitacion, acuse real DIAN, validaciones normativas adicionales ni QA de una
fuente empresarial genuina.

