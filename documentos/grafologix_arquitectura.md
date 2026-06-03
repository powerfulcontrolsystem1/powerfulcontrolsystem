# GRAFOLOGIX - Arquitectura Fase 1

Fecha: 2026-06-01
Estado: Fase 1 y Fase 2 implementadas

## Objetivo

GRAFOLOGIX permite a una empresa subir o capturar una fotografia de texto
manuscrito, ajustar la imagen en navegador y generar un informe grafológico
heuristico con metricas visuales, interpretacion orientativa y exportacion HTML
imprimible, PDF, Word compatible, JSON, CSV y TXT.

Cada manuscrito puede quedar asociado a un cliente central de la empresa. La UI
busca clientes por nombre o documento, permite crear uno nuevo con el endpoint
general de clientes y guarda descripcion/caracteristicas de la persona dentro
del informe.

El modulo no debe usarse como diagnostico psicologico, medico, juridico ni como
decision automatizada de seleccion de personal. La interpretacion es una lectura
heuristica sobre rasgos graficos.

## Estructura implementada

```text
backend/internal/grafologia/
  analyzer.go        motor matematico en Go puro
  preprocess.go      artefactos de preprocesamiento visual
  report.go          informe HTML imprimible blanco y negro
  analyzer_test.go   prueba con manuscrito sintetico
backend/db/grafologia.go
backend/handlers/grafologia.go
web/administrar_empresa/grafologia.html
web/js/grafologia.js
documentos/grafologix_arquitectura.md
```

La estructura solicitada para un proyecto independiente se adapta al monolito
actual asi:

| Carpeta objetivo | Ubicacion real |
| --- | --- |
| `/cmd/server` | `backend/main.go` |
| `/internal/grafologia` | `backend/internal/grafologia` |
| `/internal/imageprocessor` | `backend/internal/grafologia/analyzer.go` |
| `/internal/analyzer` | `backend/internal/grafologia/analyzer.go` |
| `/internal/reports` | `backend/internal/grafologia/report.go` |
| `/internal/database` | `backend/db/grafologia.go` |
| `/internal/handlers` | `backend/handlers/grafologia.go` |
| `/web/static` | `web/js/grafologia.js` y estilos de la pagina |
| `/web/templates` | HTML imprimible generado por Go |
| `/docker` | `deploy/` del proyecto; Tesseract/OpenCV quedan como paquetes opcionales |

## Esquema PostgreSQL

Tabla principal en `pcs_empresas`:

```sql
empresa_grafologia_analisis (
  id BIGSERIAL PRIMARY KEY,
  empresa_id INTEGER NOT NULL,
  cliente_id INTEGER DEFAULT 0,
  cliente_nombre TEXT,
  cliente_documento TEXT,
  persona_descripcion TEXT,
  persona_caracteristicas TEXT,
  titulo TEXT,
  archivo_nombre TEXT,
  imagen_url TEXT,
  imagen_mime TEXT,
  ocr_texto TEXT,
  ocr_motor TEXT DEFAULT 'go_heuristico',
  estado TEXT DEFAULT 'completado',
  resumen TEXT,
  metricas_json TEXT,
  interpretacion_json TEXT,
  preprocesamiento_json TEXT,
  reporte_html TEXT,
  confianza_global REAL DEFAULT 0,
  usuario_creador TEXT,
  fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
  fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP
)
```

Indices:

- `ix_grafologia_analisis_empresa(empresa_id, fecha_creacion DESC)`
- `ix_grafologia_analisis_estado(empresa_id, estado)`
- `ix_grafologia_analisis_cliente(empresa_id, cliente_id)`

Las imagenes se guardan en:

```text
web/uploads/empresas/empresa_{empresa_id}/imagenes/grafologia/
```

## API REST

Endpoint protegido por `WithEmpresaGrafologiaPermissions`:

```http
GET  /api/empresa/grafologia?empresa_id={id}&action=dashboard
GET  /api/empresa/grafologia?empresa_id={id}&action=catalogo
GET  /api/empresa/grafologia?empresa_id={id}&action=analisis&id={analisis_id}
GET  /api/empresa/grafologia?empresa_id={id}&action=reporte&id={analisis_id}&format=html
GET  /api/empresa/grafologia?empresa_id={id}&action=reporte&id={analisis_id}&format=json
GET  /api/empresa/grafologia?empresa_id={id}&action=reporte&id={analisis_id}&format=pdf
GET  /api/empresa/grafologia?empresa_id={id}&action=reporte&id={analisis_id}&format=doc
GET  /api/empresa/grafologia?empresa_id={id}&action=reporte&id={analisis_id}&format=csv
GET  /api/empresa/grafologia?empresa_id={id}&action=reporte&id={analisis_id}&format=txt
POST /api/empresa/grafologia?empresa_id={id}&action=analizar
POST /api/empresa/grafologia?empresa_id={id}&action=analizar_ia
```

`POST analizar` usa `multipart/form-data`:

- `imagen`: archivo de imagen.
- `titulo`: titulo del informe.
- `ocr_texto`: texto OCR opcional si se pega manualmente o viene de otro motor.
- `cliente_id`: cliente central asociado, validado por `empresa_id`.
- `persona_descripcion`: contexto de la persona asociada al manuscrito.
- `persona_caracteristicas`: caracteristicas registradas por el operador.

`POST analizar_ia` usa el mismo `multipart/form-data`, envia la imagen ajustada
al modelo configurado `openai:gpt-5.5` del modulo Chat IA y devuelve un informe
orientativo en texto. No crea dependencias nuevas: reutiliza el cliente OpenAI
existente, los checks globales de IA, el limite diario GPT-5.5 por empresa y el
registro de uso/auditoria de consultas IA. El cliente asociado se valida por
`empresa_id` igual que en el analisis local.

## Flujo operativo

1. El usuario abre `Administrar empresa > Analisis y control > GRAFOLOGIX`.
2. Busca el cliente central por nombre/documento o lo crea desde la tarjeta
   `Cliente asociado`.
3. Registra descripcion y caracteristicas autorizadas de la persona.
4. Sube una foto, arrastra un archivo o toma una fotografia desde la camara.
5. Ajusta brillo, contraste, recorte central o recorte automatico por tinta.
6. El navegador envia el canvas final como PNG al backend.
7. El backend valida que `cliente_id` pertenezca a la misma empresa y guarda la
   imagen aislada por `empresa_id`.
8. Si `GRAFOLOGIA_TESSERACT_ENABLED=1`, intenta OCR con Tesseract CLI.
9. El motor Go calcula metricas de imagen y rasgos heurísticos.
10. El preprocesador genera escala de grises, binarizacion, bordes y overlay de
   lineas/margenes.
11. Se guarda el resultado en PostgreSQL con snapshot del cliente/persona.
12. La UI muestra metricas, barras de interpretacion, preprocesamiento visual e
    historial.
13. El informe se abre en HTML imprimible, PDF real, Word compatible, JSON,
    CSV o TXT.
14. Opcionalmente, el operador presiona `Analizar con GPT-5.5`; el navegador
    envia el mismo canvas a `action=analizar_ia` y muestra una respuesta
    complementaria separada del informe local.

## Algoritmos Fase 1

El motor usa Go puro y libreria estandar:

- decodificacion JPEG/PNG/GIF;
- conversion a escala de grises;
- umbral Otsu;
- densidad y oscuridad de tinta;
- segmentacion de lineas por proyeccion horizontal;
- estimacion de palabras/letras por corridas verticales;
- inclinacion por regresion de centroides;
- presion por oscuridad y densidad;
- tamano por altura promedio de bandas;
- espaciado por huecos entre lineas y grupos;
- continuidad por corridas horizontales;
- direccion de lineas por centroides izquierda/derecha;
- margenes por caja envolvente;
- regularidad por variacion de alturas y espacios;
- forma de letras por vecindad de pixeles y cambios angulosos.

Cada metrica incluye un bloque `details` en `metricas_json` para que los
reportes muestren las medidas usadas por el motor: angulo de inclinacion,
pendiente, altura promedio/minima/maxima de renglones, separacion entre
letras/componentes, palabras y lineas, continuidad, direccion de linea base,
margenes izquierdo/derecho/superior/inferior, densidad de tinta, umbral Otsu,
regularidad y forma de letras. Estos detalles se imprimen en HTML, PDF
multipagina, Word, TXT y CSV.

## Preprocesamiento Fase 2

El backend genera artefactos PNG por analisis:

- `grayscale`: imagen normalizada a escala de grises.
- `binary`: tinta/fondo por umbral Otsu.
- `edges`: bordes por gradiente tipo Sobel.
- `lines`: imagen original con bandas de lineas y caja envolvente de tinta.

Los archivos se guardan en:

```text
web/uploads/empresas/empresa_{empresa_id}/imagenes/grafologia/procesado/
```

El JSON `preprocesamiento_json` guarda URLs, calidad visual, umbral, lineas y
caja de tinta.

## OCR, zoom y OpenCV

El backend no agrega dependencias Go nuevas. En Docker/VPS el OCR libre queda
habilitado con Tesseract CLI dentro de la imagen del backend:

```env
GRAFOLOGIA_TESSERACT_ENABLED=1
GRAFOLOGIA_TESSERACT_BIN=tesseract
GRAFOLOGIA_TESSERACT_LANG=spa+eng
```

La pantalla de carga permite ampliar, reducir y restablecer la vista previa de la
imagen entre 50% y 300%. El zoom se aplica al recorte visible que se envia al
motor, junto con brillo, contraste, recorte central y auto perspectiva.

OpenCV queda recomendado como sidecar o herramienta CLI futura para mejorar
perspectiva, Canny, Hough y deskew avanzado sin acoplar bindings externos al
backend Go.

## Seguridad multiempresa

- El endpoint exige `empresa_id`, sesion, alcance, licencia y permiso de modulo.
- Toda consulta filtra por `empresa_id`.
- Si se envia `cliente_id`, el backend verifica `clientes.empresa_id + id`
  antes de guardar el analisis; un cliente de otra empresa se rechaza.
- La accion `analizar_ia` respeta la configuracion global/empresarial del Chat
  IA y bloquea llamadas cuando GPT-5.5 esta desactivado o sin cupo diario.
- Los archivos cargados se guardan en carpeta de la empresa.
- No se aceptan tipos que no sean `image/*`.
- Los reportes solo cargan `id` asociado al mismo `empresa_id`.

## Pruebas

Comandos usados:

```powershell
go test ./internal/grafologia -count=1
go test ./internal/grafologia ./db ./handlers -run "Grafologia|EmpresaRoutesUsePermissionWrappers" -count=1
go test . -run TestEmpresaRoutesUsePermissionWrappers -count=1
node --check web/js/grafologia.js
```

## Pendientes Fase 2

- Dockerfile/VPS con paquetes opcionales `tesseract-ocr`, `tesseract-ocr-spa` y
  utilidades OpenCV CLI si se decide instalar en imagen.
- PDF avanzado con logo, graficas y estilos completos. La fase actual ya entrega
  PDF real multipagina con Go estandar y conserva HTML para el detalle visual.
- Correccion de perspectiva completa por cuadrilateros.
- Panel administrativo super para activar/desactivar el modulo por licencia.
- Dataset y motor entrenable futuro para separar analisis geométrico de reglas
  grafológicas.
