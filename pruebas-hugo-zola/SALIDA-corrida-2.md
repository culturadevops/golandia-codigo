# Hugo (Go) contra Zola (Rust) · corrida 2

> Corrida del 2026-11-27 09:38. Cada bloque, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_hugo_zola.py 2`.

- Máquina (docker info): 24 CPU · 16487469056 B · 6.18.33.2-microsoft-standard-WSL2 · 29.6.2
- Contenedor: `--cpus 2 --memory 1g` (y 4g donde se dice), imagen gl-grb-hugo:1 (debian:trixie-slim + binarios oficiales)

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

## A · ¿Generan lo mismo? (10.000 con 1 GiB; 30.000 con 4 GiB, porque con 1 GiB Hugo no termina: ver B)

- 10.000 páginas (--memory 1g): hugo rc 0 · 10.024 HTML · 10.026 archivos | zola rc 0 · 10.024 HTML · 10.026 archivos
  - solo en hugo: [] · solo en zola: []
  - texto de 25 páginas comparado (sin etiquetas HTML): distinto en 0 []
  - ejemplo: «Sección 1: Texto tipo método Valor puerto cola go estructura canal binario imagen compilador puerto go vida disco vida página. Orden carga etiqueta etiqueta map…»
- 30.000 páginas (--memory 4g): hugo rc 0 · 30.024 HTML · 30.026 archivos | zola rc 0 · 30.024 HTML · 30.028 archivos
  - solo en hugo: [] · solo en zola: ['sitemap1.xml', 'sitemap2.xml']
  - texto de 25 páginas comparado (sin etiquetas HTML): distinto en 0 []
  - ejemplo: «Sección 1: Texto tipo método Valor puerto cola go estructura canal binario imagen compilador puerto go vida disco vida página. Orden carga etiqueta etiqueta map…»
  - archivos que no son HTML y no coinciden (no paran el banco): ['sitemap1.xml', 'sitemap2.xml']

## B · Build completo («primera» = sin public/, «repetida» = otra vez con public/ ya escrito)

Variantes de Hugo: HUGO_MEMORYLIMIT (GB; la documentación dice que por defecto es el 25 % de la memoria del sistema, y dentro del contenedor el sistema son 16 GB) y GOMEMLIMIT (el límite blando del recolector de Go).

| páginas | memoria | herramienta | build | reps | s (mediana) | s (mín–máx) | RAM máx. MiB | CPU s (u+s) | HTML | rc |
|---|---|---|---|---|---|---|---|---|---|---|
| 10.000 | 1g | hugo | primera | 5 | 3.41 | 3.28–3.77 | 1004.7 | 6.12 | [10024] | ['0'] |
| 10.000 | 1g | hugo | repetida | 5 | 2.92 | 2.80–5.35 | 1002.8 | 4.99 | [10024] | ['0'] |
| 10.000 | 1g | zola | primera | 5 | 1.63 | 1.57–1.67 | 290.9 | 2.81 | [10024] | ['0'] |
| 10.000 | 1g | zola | repetida | 5 | 1.82 | 1.80–1.83 | 290.8 | 2.96 | [10024] | ['0'] |
| 10.000 | 1g | hugo [GOMEMLIMIT=700MiB] | primera | 3 | 4.40 | 4.40–4.78 | 731.7 | 8.37 | [10024] | ['0'] |
| 10.000 | 1g | hugo [HUGO_MEMORYLIMIT=1] | primera | 3 | 3.72 | 3.35–4.08 | 1009.9 | 6.31 | [10024] | ['0'] |
| 30.000 | 1g | hugo | primera | 1 | 29.50 | 29.50–29.50 | 1041.3 | 22.1 | [17060] | ['137'] |
| 30.000 | 1g | hugo [GOMEMLIMIT=700MiB] | primera | 3 | 32.77 | 32.72–33.65 | 1047.6 | 58.68 | [30024] | ['0'] |
| 30.000 | 1g | hugo [HUGO_MEMORYLIMIT=1] | primera | 3 | 64.10 | 63.58–64.19 | 1040.6 | 45.24 | [30024] | ['0'] |
| 30.000 | 1g | zola | primera | 3 | 5.44 | 5.25–5.76 | 838.6 | 9.17 | [30024] | ['0'] |
| 10.000 | 4g | hugo | primera | 3 | 3.17 | 3.14–3.53 | 1015.2 | 5.88 | [10024] | ['0'] |
| 10.000 | 4g | hugo | repetida | 3 | 2.58 | 2.50–2.63 | 1042.8 | 4.83 | [10024] | ['0'] |
| 10.000 | 4g | zola | primera | 3 | 1.64 | 1.62–1.65 | 291.0 | 2.81 | [10024] | ['0'] |
| 10.000 | 4g | zola | repetida | 3 | 1.82 | 1.82–1.90 | 291.1 | 2.96 | [10024] | ['0'] |
| 30.000 | 4g | hugo | primera | 3 | 10.44 | 10.10–10.69 | 2462.9 | 19.65 | [30024] | ['0'] |
| 30.000 | 4g | hugo | repetida | 3 | 8.17 | 8.16–8.17 | 2468.5 | 15.32 | [30024] | ['0'] |
| 30.000 | 4g | zola | primera | 3 | 5.16 | 5.10–5.26 | 838.8 | 8.81 | [30024] | ['0'] |
| 30.000 | 4g | zola | repetida | 3 | 5.83 | 5.72–5.83 | 837.4 | 9.45 | [30024] | ['0'] |

## C · Servidor de desarrollo con 1 GiB (hugo server / zola serve): listo, RAM y 10 ediciones de página + 3 de plantilla

Tiempo medido desde fuera: desde que se guarda el archivo hasta que la página servida por HTTP ya trae el cambio (sondeo cada 5 ms). Hugo con sus opciones por defecto (Fast Render Mode activo).

| páginas | herramienta | listo para servir (s) | RAM al estar listo MiB | editar página s (mediana · mín–máx) | editar plantilla s (mediana · mín–máx) | RAM al final MiB (máx.) |
|---|---|---|---|---|---|---|
| 10.000 | hugo | 4.411 | 1009.4 | 0.87 · 0.51–1.13 (10/10) | 0.46 · 0.36–0.50 (3/3) | 1032.1 (1049.7) |
| 10.000 | zola | 0.798 | 342.4 | 1.96 · 1.76–2.07 (10/10) | 1.96 · 1.94–2.06 (3/3) | 641.1 (641.1) |

```
{"herr": "hugo", "que": "listo", "seg": 4.411, "VmHWM": 1021.0, "VmRSS": 1009.4}
{"herr": "hugo", "que": "editar-pagina", "k": 0, "seg": 0.569}
{"herr": "hugo", "que": "editar-pagina", "k": 1, "seg": 0.819}
{"herr": "hugo", "que": "editar-pagina", "k": 2, "seg": 0.542}
{"herr": "hugo", "que": "editar-pagina", "k": 3, "seg": 0.513}
{"herr": "hugo", "que": "editar-pagina", "k": 4, "seg": 0.682}
{"herr": "hugo", "que": "editar-pagina", "k": 5, "seg": 1.043}
{"herr": "hugo", "que": "editar-pagina", "k": 6, "seg": 0.93}
{"herr": "hugo", "que": "editar-pagina", "k": 7, "seg": 1.015}
{"herr": "hugo", "que": "editar-pagina", "k": 8, "seg": 1.131}
{"herr": "hugo", "que": "editar-pagina", "k": 9, "seg": 1.064}
{"herr": "hugo", "que": "editar-plantilla", "k": 0, "seg": 0.356}
{"herr": "hugo", "que": "editar-plantilla", "k": 1, "seg": 0.457}
{"herr": "hugo", "que": "editar-plantilla", "k": 2, "seg": 0.503}
{"herr": "hugo", "que": "memoria-al-final", "VmHWM": 1049.7, "VmRSS": 1032.1}
{"herr": "zola", "que": "listo", "seg": 0.798, "VmHWM": 342.4, "VmRSS": 342.4}
{"herr": "zola", "que": "editar-pagina", "k": 0, "seg": 1.91}
{"herr": "zola", "que": "editar-pagina", "k": 1, "seg": 1.756}
{"herr": "zola", "que": "editar-pagina", "k": 2, "seg": 1.768}
{"herr": "zola", "que": "editar-pagina", "k": 3, "seg": 1.822}
{"herr": "zola", "que": "editar-pagina", "k": 4, "seg": 2.026}
{"herr": "zola", "que": "editar-pagina", "k": 5, "seg": 2.036}
{"herr": "zola", "que": "editar-pagina", "k": 6, "seg": 1.944}
{"herr": "zola", "que": "editar-pagina", "k": 7, "seg": 2.029}
{"herr": "zola", "que": "editar-pagina", "k": 8, "seg": 1.972}
{"herr": "zola", "que": "editar-pagina", "k": 9, "seg": 2.074}
{"herr": "zola", "que": "editar-plantilla", "k": 0, "seg": 1.956}
{"herr": "zola", "que": "editar-plantilla", "k": 1, "seg": 2.063}
{"herr": "zola", "que": "editar-plantilla", "k": 2, "seg": 1.939}
{"herr": "zola", "que": "memoria-al-final", "VmHWM": 641.1, "VmRSS": 641.1}
```

FIN
