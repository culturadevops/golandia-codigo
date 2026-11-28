# Banco de «Hugo (Go) contra Zola (Rust)» (golandia)

Código del vídeo [«Hugo (Go) contra Zola (Rust), lo medí»](https://youtu.be/4W0jmww5wBw) del canal Golandia, de la serie «Go contra Rust: herramientas populares».

| archivo | qué es |
|---|---|
| `pruebas_hugo_zola.py` | el banco |
| `grb_comun.py` | lo común de la serie (candado, Docker con timeout) |
| `Dockerfile*`, configs y scripts | las imágenes y la configuración de cada herramienta |
| `SALIDA-corrida-*.md` | resultados (las marcadas PARCIAL o DESCARTADA no cuentan; el nombre dice por qué) |

Ejecutar (necesita Docker; toma el candado `_logs/banco.lock` en la raíz del repo): `python pruebas_hugo_zola.py <corrida>`
