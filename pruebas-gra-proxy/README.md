# Banco de «Proxy inverso: Caddy y Traefik (Go) contra Pingap y Sōzu (Rust)» (golandia)

Código del vídeo [«Proxy inverso: Caddy y Traefik (Go) contra Pingap y Sōzu (Rust), lo medí»](https://youtu.be/cov3H4rElgY) del canal Golandia, de la serie «Go contra Rust: herramientas populares».

| archivo | qué es |
|---|---|
| `pruebas_gra_proxy.py` | el banco |
| `pruebas_gra_comun.py` | lo común de la serie (candado, Docker con timeout) |
| `muestreo.py` | muestreo de CPU y memoria durante la carga |
| `Dockerfile*`, configs y scripts | las imágenes y la configuración de cada herramienta |
| `SALIDA-corrida-*.md` | resultados (las marcadas PARCIAL o DESCARTADA no cuentan; el nombre dice por qué) |

Ejecutar (necesita Docker; toma el candado `_logs/banco.lock` en la raíz del repo): `python pruebas_gra_proxy.py <corrida>`
