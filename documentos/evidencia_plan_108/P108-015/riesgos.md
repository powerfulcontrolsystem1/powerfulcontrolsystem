# Riesgos P108-015

1. Gmail no garantiza el avatar de un BIMI autoafirmado. Falta contratar VMC o
   CMC, publicar el PEM por HTTPS y completar `a=` en DNS.
2. Algunos clientes no soportan BIMI y pueden mostrar una inicial aunque el
   dominio y el mensaje estén correctamente configurados.
3. El arreglo CSRF fue probado con envío real en staging, pero la rama todavía
   requiere revisión e integración. El ajuste responsive final debe desplegarse
   y volver a capturarse.
4. El endpoint público de salud no informa SHA/digest; debe incorporarse una
   prueba de identidad del artefacto antes del cierre productivo.
