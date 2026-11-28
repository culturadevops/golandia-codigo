# Hugo (Go) contra Zola (Rust) · corrida 1

> Corrida del 2026-11-26 21:56. Cada bloque, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_hugo_zola.py 1`.

- Máquina (docker info): 24 CPU · 16487464960 B · 6.18.33.2-microsoft-standard-WSL2 · 29.6.2
- Contenedor: `--cpus 2 --memory 1g`, imagen gl-grb-hugo:1 (debian:trixie-slim + binarios oficiales)

## 0 · Versiones, binarios y sitios

```
hugo v0.170.1-c8b03de9564cf846a3986e879727d638bb96ab33 linux/amd64 BuildDate=2026-11-17T16:38:13Z VendorInfo=gohugoio
zola 0.23.6
-rwxr-xr-x 1 root root 61145248 Nov 17 16:41 /usr/local/bin/hugo
-rwxr-xr-x 1 1001 1001 50794544 Sep 12 22:02 /usr/local/bin/zola
8f5132b3522412d04e395e0b25f6d68613ad272a873e54a2b3ebf664873024a4  zola-v0.23.6-x86_64-unknown-linux-gnu.tar.gz
24
10000: 10001 md hugo, 10001 md zola, 40M
30000: 30001 md hugo, 30001 md zola, 119M
```

## A · ¿Generan lo mismo?

- 10.000 páginas: hugo rc 0 · 10.024 HTML · 10.026 archivos | zola rc 0 · 10.024 HTML · 10.026 archivos
  - solo en hugo: [] · solo en zola: []
  - texto de 25 páginas comparado (sin etiquetas HTML): distinto en 0 []
  - ejemplo: «Sección 1: Texto tipo método Valor puerto cola go estructura canal binario imagen compilador puerto go vida disco vida página. Orden carga etiqueta etiqueta map…»
- 30.000 páginas: hugo rc 137 · 17.207 HTML · 17.209 archivos | zola rc 0 · 30.024 HTML · 30.028 archivos
  - solo en hugo: [] · solo en zola: ['posts/nota-017174/index.html', 'posts/nota-017190/index.html', 'posts/nota-017191/index.html', 'posts/nota-017193/index.html', 'posts/nota-017196/index.html', 'posts/nota-017198/index.html', 'posts/nota-017199/index.html', 'posts/nota-017200/index.html', 'posts/nota-017201/index.html', 'posts/nota-017202/index.html']
  - texto de 26 páginas comparado (sin etiquetas HTML): distinto en 0 []
  - ejemplo: «Sección 1: Texto tipo método Valor puerto cola go estructura canal binario imagen compilador puerto go vida disco vida página. Orden carga etiqueta etiqueta map…»

⛔ No generan lo mismo (Command terminated by signal 9
26.54 1066928 11.66 10.14
 ). Paro.
