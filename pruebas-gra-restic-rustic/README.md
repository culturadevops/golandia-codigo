# Banco de «restic (Go) contra rustic (Rust)» (golandia)

Código del vídeo [«restic (Go) contra rustic (Rust), lo medí»](https://youtu.be/2QQm41xmn30) del canal Golandia, de la serie «Go contra Rust: herramientas populares».

| archivo | qué es |
|---|---|
| `pruebas_gra_restic_rustic.py` | el banco |
| `pruebas_gra_comun.py` | lo común de la serie (candado, Docker con timeout) |
| `Dockerfile*`, configs y scripts | las imágenes y la configuración de cada herramienta |
| `SALIDA-corrida-*.md` | resultados (las marcadas PARCIAL o DESCARTADA no cuentan; el nombre dice por qué) |

Ejecutar (necesita Docker; toma el candado `_logs/banco.lock` en la raíz del repo): `python pruebas_gra_restic_rustic.py <corrida>`
