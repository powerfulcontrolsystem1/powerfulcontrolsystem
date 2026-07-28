# Riesgos P108-015

1. Gmail no garantiza el avatar de un BIMI autoafirmado. Falta contratar VMC o
   CMC, publicar el PEM por HTTPS y completar `a=` en DNS.
2. Algunos clientes no soportan BIMI y pueden mostrar una inicial aunque el
   dominio y el mensaje estén correctamente configurados.
3. El arreglo CSRF aún debe integrarse y desplegarse antes de repetir la prueba
   real. La evidencia local no sustituye esa aceptación.
4. El endpoint público de salud no informa SHA/digest; debe incorporarse una
   prueba de identidad del artefacto antes del cierre productivo.

