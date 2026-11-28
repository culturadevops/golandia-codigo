# Banco de «lazygit (Go) contra gitui (Rust)» (golandia)

Código del vídeo [«lazygit (Go) contra gitui (Rust), lo medí»](https://youtu.be/q9SluTP6igE) del canal Golandia, de la serie «Go contra Rust: herramientas populares».

| archivo | qué es |
|---|---|
| `pruebas_lazygit_gitui.py` | el banco |
| `grb_comun.py` | lo común de la serie (candado, Docker con timeout) |
| `Dockerfile*`, configs y scripts | las imágenes y la configuración de cada herramienta |
| `SALIDA-corrida-*.md` | resultados (las marcadas PARCIAL o DESCARTADA no cuentan; el nombre dice por qué) |

Ejecutar (necesita Docker; toma el candado `_logs/banco.lock` en la raíz del repo): `python pruebas_lazygit_gitui.py <corrida>`
