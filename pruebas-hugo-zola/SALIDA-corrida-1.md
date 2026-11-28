# Hugo (Go) contra Zola (Rust) · corrida 1

> Corrida del 2026-11-27 09:22. Cada bloque, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_hugo_zola.py 1`.

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
| 10.000 | 1g | hugo | primera | 5 | 3.82 | 3.66–4.67 | 990.5 | 6.42 | [10024] | ['0'] |
| 10.000 | 1g | hugo | repetida | 5 | 3.44 | 3.07–4.58 | 1026.6 | 5.48 | [10024] | ['0'] |
| 10.000 | 1g | zola | primera | 5 | 1.58 | 1.58–1.63 | 291.2 | 2.76 | [10024] | ['0'] |
| 10.000 | 1g | zola | repetida | 5 | 1.83 | 1.77–1.86 | 291.1 | 3.01 | [10024] | ['0'] |
| 10.000 | 1g | hugo [GOMEMLIMIT=700MiB] | primera | 3 | 4.41 | 4.34–4.69 | 732.1 | 8.4 | [10024] | ['0'] |
| 10.000 | 1g | hugo [HUGO_MEMORYLIMIT=1] | primera | 3 | 3.78 | 3.65–4.34 | 1011.9 | 6.44 | [10024] | ['0'] |
| 30.000 | 1g | hugo | primera | 1 | 28.04 | 28.04–28.04 | 1041.8 | 21.33 | [16346] | ['137'] |
| 30.000 | 1g | hugo [GOMEMLIMIT=700MiB] | primera | 3 | 32.29 | 31.97–32.77 | 1051.8 | 57.98 | [30024] | ['0'] |
| 30.000 | 1g | hugo [HUGO_MEMORYLIMIT=1] | primera | 3 | 62.83 | 57.93–66.13 | 1040.0 | 45.45 | [30024] | ['0'] |
| 30.000 | 1g | zola | primera | 3 | 5.49 | 5.49–5.98 | 838.7 | 9.3 | [30024] | ['0'] |
| 10.000 | 4g | hugo | primera | 3 | 3.11 | 3.09–3.32 | 1012.3 | 5.77 | [10024] | ['0'] |
| 10.000 | 4g | hugo | repetida | 3 | 2.52 | 2.50–2.55 | 1037.9 | 4.69 | [10024] | ['0'] |
| 10.000 | 4g | zola | primera | 3 | 1.62 | 1.61–1.66 | 291.3 | 2.79 | [10024] | ['0'] |
| 10.000 | 4g | zola | repetida | 3 | 1.82 | 1.80–1.89 | 290.9 | 2.98 | [10024] | ['0'] |
| 30.000 | 4g | hugo | primera | 3 | 10.02 | 9.75–10.15 | 2464.2 | 18.76 | [30024] | ['0'] |
| 30.000 | 4g | hugo | repetida | 3 | 8.19 | 8.05–8.40 | 2467.6 | 15.43 | [30024] | ['0'] |
| 30.000 | 4g | zola | primera | 3 | 5.11 | 5.04–5.16 | 838.5 | 8.73 | [30024] | ['0'] |
| 30.000 | 4g | zola | repetida | 3 | 5.89 | 5.85–6.45 | 838.2 | 9.53 | [30024] | ['0'] |

## C · Servidor de desarrollo con 1 GiB (hugo server / zola serve): listo, RAM y 10 ediciones de página + 3 de plantilla

Tiempo medido desde fuera: desde que se guarda el archivo hasta que la página servida por HTTP ya trae el cambio (sondeo cada 5 ms). Hugo con sus opciones por defecto (Fast Render Mode activo).

| páginas | herramienta | listo para servir (s) | RAM al estar listo MiB | editar página s (mediana · mín–máx) | editar plantilla s (mediana · mín–máx) | RAM al final MiB (máx.) |
|---|---|---|---|---|---|---|
| 10.000 | hugo | 4.508 | 1015.6 | 0.67 · 0.46–1.28 (10/10) | 0.42 · 0.21–0.52 (3/3) | 1034.6 (1045.5) |
| 10.000 | zola | 0.816 | 342.1 | 1.99 · 1.73–2.04 (10/10) | 1.98 · 1.97–2.01 (3/3) | 644.2 (644.2) |

```
{"herr": "hugo", "que": "listo", "seg": 4.508, "VmHWM": 1016.6, "VmRSS": 1015.6}
{"herr": "hugo", "que": "editar-pagina", "k": 0, "seg": 1.277}
{"herr": "hugo", "que": "editar-pagina", "k": 1, "seg": 0.556}
{"herr": "hugo", "que": "editar-pagina", "k": 2, "seg": 0.459}
{"herr": "hugo", "que": "editar-pagina", "k": 3, "seg": 0.55}
{"herr": "hugo", "que": "editar-pagina", "k": 4, "seg": 0.592}
{"herr": "hugo", "que": "editar-pagina", "k": 5, "seg": 0.733}
{"herr": "hugo", "que": "editar-pagina", "k": 6, "seg": 0.852}
{"herr": "hugo", "que": "editar-pagina", "k": 7, "seg": 0.863}
{"herr": "hugo", "que": "editar-pagina", "k": 8, "seg": 0.615}
{"herr": "hugo", "que": "editar-pagina", "k": 9, "seg": 0.855}
{"herr": "hugo", "que": "editar-plantilla", "k": 0, "seg": 0.214}
{"herr": "hugo", "que": "editar-plantilla", "k": 1, "seg": 0.418}
{"herr": "hugo", "que": "editar-plantilla", "k": 2, "seg": 0.519}
{"herr": "hugo", "que": "memoria-al-final", "VmHWM": 1045.5, "VmRSS": 1034.6}
{"herr": "zola", "que": "listo", "seg": 0.816, "VmHWM": 342.1, "VmRSS": 342.1}
{"herr": "zola", "que": "editar-pagina", "k": 0, "seg": 1.912}
{"herr": "zola", "que": "editar-pagina", "k": 1, "seg": 1.787}
{"herr": "zola", "que": "editar-pagina", "k": 2, "seg": 1.734}
{"herr": "zola", "que": "editar-pagina", "k": 3, "seg": 1.783}
{"herr": "zola", "que": "editar-pagina", "k": 4, "seg": 1.991}
{"herr": "zola", "que": "editar-pagina", "k": 5, "seg": 2.044}
{"herr": "zola", "que": "editar-pagina", "k": 6, "seg": 2.024}
{"herr": "zola", "que": "editar-pagina", "k": 7, "seg": 1.984}
{"herr": "zola", "que": "editar-pagina", "k": 8, "seg": 1.998}
{"herr": "zola", "que": "editar-pagina", "k": 9, "seg": 2.037}
{"herr": "zola", "que": "editar-plantilla", "k": 0, "seg": 1.982}
{"herr": "zola", "que": "editar-plantilla", "k": 1, "seg": 2.006}
{"herr": "zola", "que": "editar-plantilla", "k": 2, "seg": 1.971}
{"herr": "zola", "que": "memoria-al-final", "VmHWM": 644.2, "VmRSS": 644.2}
```

FIN
