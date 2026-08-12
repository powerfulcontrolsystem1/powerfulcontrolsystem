# P110-007 - ClamAV HTTP del candidato final

Fecha: 2026-08-11

Ambiente: staging aislado, empresa PCS (#12), administrador autorizado.

Producción: no modificada.

## Candidato

- SHA: `fd6a4a8a18b44100d423cdae07db4d957e69da97`.
- API: `sha256:6c587545048b4a6f4c8a7a9353f5eaa3cda87f6f33a914f7a5d9fd95e68c9504`.
- Migrador: `sha256:f8a4c65ee756b3af4c691010a6484969a96cfefc0d1d8057c95f7b67b4a4a660`.
- Worker: `sha256:55f6c59e08d9fa42bbf8300cddee9d05ce992a2b7ad474f7be4578d461c69578`.
- Frontend: `sha256:12d7a47c76656e9a7af0b800519357f68db04814662055dddb1ac3a3267331ff`.

## Prueba por flujo oficial

La sesión oficial de administrador se abrió sobre PCS en staging. La pantalla
de captura inteligente mostró el formulario de radicación, contexto empresarial
`#12` y el flujo de soportes visible. El selector de archivos del navegador
interno se interrumpió antes de entregar un archivo, sin envío ni persistencia;
por ello la carga se ejecutó por el mismo endpoint HTTP oficial con su sesión y
CSRF, sin imprimir cookies ni credenciales.

| Caso | HTTP | Persistencia | Resultado |
|---|---:|---|---|
| XML limpio | 200 | una fila nueva | PASS |
| Sonda EICAR | 422 | cero filas nuevas | PASS |
| ClamAV detenido, XML limpio | 503 | cero filas nuevas | PASS fail-closed |
| ClamAV saludable, XML limpio | 200 | una fila nueva | PASS recuperación |

Las dos comparaciones de filas se realizaron dentro de la base de staging solo
como conteos de la empresa #12. Los archivos temporales, cookies y respuestas
se eliminaron al finalizar.

## Límite

No sustituye DAST, pruebas A/B, sesiones, CSP ni la verificación del receptor
externo de alertas. P110-007 permanece **parcial** y el veredicto es **NO-GO**.
