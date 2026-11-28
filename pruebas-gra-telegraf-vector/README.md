# Banco de «Telegraf (Go) contra Vector (Rust)» (golandia)

Código del vídeo [«Telegraf (Go) contra Vector (Rust), lo medí»](https://youtu.be/MASx0qWCxCQ) del canal Golandia, de la serie «Go contra Rust: herramientas populares».

| archivo | qué es |
|---|---|
| `pruebas_gra_telegraf_vector.py` | el banco |
| `pruebas_gra_comun.py` | lo común de la serie (candado, Docker con timeout) |
| `Dockerfile*`, configs y scripts | las imágenes y la configuración de cada herramienta |
| `SALIDA-corrida-*.md` | resultados (las marcadas PARCIAL o DESCARTADA no cuentan; el nombre dice por qué) |

Ejecutar (necesita Docker; toma el candado `_logs/banco.lock` en la raíz del repo): `python pruebas_gra_telegraf_vector.py <corrida>`
