# Pruebas ejecutadas — ¿Go es el mejor lenguaje para AWS Lambda?

> Salida **literal** de esta máquina. La genera `investigacion/pruebas_lambda.py` y se
> vuelve a generar sola. ⛔ Lo que no esté aquí no puede salir en pantalla.

**Entorno citable en el vídeo:** `go version go1.26.5 windows/amd64` · `Docker version 29.6.2, build dfc4efb` · `github.com/aws/aws-lambda-go v1.54.0`

**Imágenes oficiales de AWS usadas** (traen dentro el Runtime Interface Emulator):
- `public.ecr.aws/lambda/provided:al2023`
- `public.ecr.aws/lambda/python:3.13`
- `public.ecr.aws/lambda/nodejs:22`

⚠️ **Qué mide esto y qué NO.** El emulador arranca el *runtime* de Lambda, no la microVM
de Firecracker. Mide **el arranque del lenguaje y de tu código** —la parte que depende del
lenguaje—, no lo que AWS tarda en encontrarte una microVM, que es igual para los tres.

El handler es el mismo trabajo en los tres lenguajes: `n` vueltas de `i*i % 1000003`, y
opcionalmente esperar `io_ms` milisegundos.

---

## 1 · Lo que se sube: el paquete

| compilado con | arquitectura | binario | .zip que se sube |
|---|---|---:|---:|
| `plano` | amd64 | 11.977.735 B (11.42 MiB) | 6.405.133 B (6.11 MiB) |
| `plano` | arm64 | 11.145.456 B (10.63 MiB) | 5.853.783 B (5.58 MiB) |
| `lambda.norpc` | amd64 | 9.110.061 B (8.69 MiB) | 5.005.513 B (4.77 MiB) |
| `lambda.norpc` | arm64 | 8.455.290 B (8.06 MiB) | 4.569.135 B (4.36 MiB) |
| `lambda.norpc + -s -w` | amd64 | 6.303.906 B (6.01 MiB) | 2.620.249 B (2.50 MiB) |
| `lambda.norpc + -s -w` | arm64 | 5.832.866 B (5.56 MiB) | 2.359.320 B (2.25 MiB) |

Y lo que se sube en los otros dos, donde el runtime lo pone AWS:

| lenguaje | fuente | .zip que se sube |
|---|---:|---:|
| python | 313 B | 294 B |
| node | 354 B | 338 B |

> 📉 Del binario plano de amd64 (11.42 MiB) al de arm64 con `lambda.norpc` y `-s -w` (5.56 MiB): **−51.3 %**.

---

## 2 · La otra forma de desplegar: la imagen

⚠️ Son **dos números distintos** y confundirlos es publicar una cifra falsa: lo que te
descargas (capas comprimidas del registro) no es lo que ocupa en tu disco (descomprimida).
La columna de descarga se comprobó resolviendo a mano el manifiesto de amd64 con
`docker manifest inspect`: sale el mismo byte.

| lenguaje | qué | referencia | descarga | en disco |
|---|---|---|---:|---:|
| go | base oficial | `public.ecr.aws/lambda/provided:al2023` | 41.13 MiB | 172MB |
| go | con el handler | `golandia-lambda-go:1` | 43.69 MiB | 181MB |
| python | base oficial | `public.ecr.aws/lambda/python:3.13` | 192.39 MiB | 793MB |
| python | con el handler | `golandia-lambda-py:1` | 192.39 MiB | 793MB |
| node | base oficial | `public.ecr.aws/lambda/nodejs:22` | 131.69 MiB | 597MB |
| node | con el handler | `golandia-lambda-node:1` | 131.69 MiB | 597MB |

---

## 3 · Arranque en frío — 8 contenedores nuevos por lenguaje

`INIT REPORT(durationMs: …)` del propio emulador: la inicialización que Lambda factura
aparte. «1ª respuesta» es lo que tarda el cliente desde que el contenedor existe.

| lenguaje | init mediana | init mínimo | init máximo | 1ª respuesta mediana |
|---|---:|---:|---:|---:|
| go | 2.8 ms | 2.5 ms | 7.5 ms | 6.0 ms |
| python | 21.5 ms | 20.4 ms | 57.0 ms | 32.3 ms |
| node | 32.3 ms | 30.0 ms | 108.5 ms | 45.4 ms |

<details><summary>las medidas, una a una</summary>

```
go       init(ms)  7.5  2.7  2.5  2.7  3.5  2.9  2.8  2.7
python   init(ms)  57.0  22.1  24.9  20.9  20.4  20.9  22.4  20.9
node     init(ms)  108.5  32.3  30.8  33.7  30.0  34.3  32.2  30.1
```

</details>

---

## 4 · Latencia templada — 200 invocaciones con 200.000 vueltas de CPU

| lenguaje | p50 | p90 | p99 | duración facturada p50 |
|---|---:|---:|---:|---:|
| go | 1.43 ms | 16.09 ms | 25.83 ms | 0.43 ms |
| python | 9.78 ms | 24.03 ms | 33.39 ms | 8.33 ms |
| node | 7.24 ms | 29.58 ms | 32.09 ms | 5.75 ms |

---

## 5 · El matiz honesto — el mismo handler esperando 100 ms de E/S

Aquí no hay trabajo de CPU: solo espera, que es lo que hace tu Lambda de verdad mientras
DynamoDB o una API ajena le contestan.

| lenguaje | p50 | p99 |
|---|---:|---:|
| go | 102.68 ms | 126.11 ms |
| python | 102.92 ms | 125.38 ms |
| node | 102.87 ms | 127.23 ms |

> 🥇 Entre el mejor y el peor hay **0.24 ms** sobre 100 ms de espera. Cuando tu Lambda espera, el lenguaje no es tu problema.

---

## 6 · La factura

⛔ Los precios **no se miden aquí: se copian de AWS**, y la fuente va escrita para que se
pueda comprobar. Lo que es nuestro es la aritmética y la duración facturada medida arriba.

**Fuente de los precios:** https://aws.amazon.com/lambda/pricing/ (us-east-1, consultado 2026-08-25)

- 0.20 $ por millón de peticiones
- 0.0000166667 $ por GB-segundo (x86)
- siempre gratis al mes: 1.000.000 peticiones y 400.000 GB-s
- Lambda redondea la duración **hacia arriba al milisegundo**

| invocaciones/mes | memoria | lenguaje | p50 facturada medida | ms cobrados | GB-s | peticiones | cómputo | **total/mes** |
|---:|---:|---|---:|---:|---:|---:|---:|---:|
| 10.000.000 | 128 MB | go | 0.43 ms | 1 ms | 1.250 | 1.80 $ | 0.00 $ | **1.80 $** |
| 10.000.000 | 128 MB | python | 8.33 ms | 9 ms | 11.250 | 1.80 $ | 0.00 $ | **1.80 $** |
| 10.000.000 | 128 MB | node | 5.75 ms | 6 ms | 7.500 | 1.80 $ | 0.00 $ | **1.80 $** |
| 1.000.000.000 | 128 MB | go | 0.43 ms | 1 ms | 125.000 | 199.80 $ | 0.00 $ | **199.80 $** |
| 1.000.000.000 | 128 MB | python | 8.33 ms | 9 ms | 1.125.000 | 199.80 $ | 12.08 $ | **211.88 $** |
| 1.000.000.000 | 128 MB | node | 5.75 ms | 6 ms | 750.000 | 199.80 $ | 5.83 $ | **205.63 $** |
| 1.000.000.000 | 1024 MB | go | 0.43 ms | 1 ms | 1.000.000 | 199.80 $ | 10.00 $ | **209.80 $** |
| 1.000.000.000 | 1024 MB | python | 8.33 ms | 9 ms | 9.000.000 | 199.80 $ | 143.33 $ | **343.13 $** |
| 1.000.000.000 | 1024 MB | node | 5.75 ms | 6 ms | 6.000.000 | 199.80 $ | 93.33 $ | **293.13 $** |

> Con 10.000.000 invocaciones al mes a 128 MB, entre el mejor (go) y el peor (go) hay **0.00 $ al mes**.

> Con 1.000.000.000 invocaciones al mes a 128 MB, entre el mejor (go) y el peor (python) hay **12.08 $ al mes**.

> Con 1.000.000.000 invocaciones al mes a 1024 MB, entre el mejor (go) y el peor (python) hay **133.33 $ al mes**.
