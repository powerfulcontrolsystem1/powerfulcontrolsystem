# P109-000 - SHA actual `3676cc02` por digest en staging

El workflow `31319557537` aprobó construcción única, Trivy, SBOM, publicación y
Compose para `3676cc02b61eca5cc7d13f21c857712006ec065a`.

Las imágenes API, migrador, worker y frontend se promovieron exclusivamente a
staging mediante `vps-staging-digest-up.sh`, sin recompilar. Después del
reinicio controlado, API, worker y frontend quedaron saludables; migrador
terminó con `0`; `/health` devolvió `ok` y `/ready` devolvió `ready`.

Las imágenes productivas se verificaron por lectura y conservaron sus tags
locales previos, todas saludables; `/health` productivo continuó `ok`. No se
creó PR ni se modificó producción. La fase continúa parcial hasta revisión y
fusión a `main`, más las demás compuertas del plan.
