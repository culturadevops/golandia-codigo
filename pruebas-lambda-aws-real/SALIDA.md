# Pruebas ejecutadas — Go en AWS Lambda de verdad (la revancha del 121)

> Salida **literal**. La generan `banco.py` (construir · medir · memoria · limpiar) e `informe.py`, a partir
> de `_datos/*.json`. ⛔ Lo que no esté aquí no puede salir en pantalla. El número de cuenta va tachado.

**Dónde:** AWS Lambda, `us-east-1`, arquitectura `x86_64`, **512 MB**, timeout 15 s. Funciones `gl-lambda-*` (todas borradas al final: §11). Última medida: 2026-11-18T15:43:59+00:00.

**El mismo handler en todos** (el del 121): `n` vueltas de `i*i % 1000003` y, si se pide, esperar `io_ms`.
Python y Node son **literalmente** los archivos del 121 (Node suma con `BigInt`, como allí).
Java gestionado y Java nativo usan el **mismo** `Manejador.java`; el nativo lleva además un bucle del
Runtime API escrito a mano (`Bootstrap.java`, sin librería). Go usa `aws-lambda-go`, Rust `lambda_runtime`,
.NET `Amazon.Lambda.RuntimeSupport` con Native AOT.

**Cómo se fuerza un arranque en frío:** antes de cada invocación fría se cambia una variable de entorno de la
función (`FRIO=1, 2, 3…`); eso retira los entornos viejos. En Java + SnapStart, además, se publica una
versión nueva por cada arranque (SnapStart solo actúa sobre versiones publicadas). Las cifras de `Init`,
`Restore`, `Duration` y `Billed Duration` son las de la línea `REPORT` que escribe Lambda.

⚠️ **«Punta a punta»** es lo que tarda `invoke` medido desde esta máquina: **incluye la red de mi casa**
hasta us-east-1. Sirve para comparar entre lenguajes, no como cifra absoluta.

---

## 0 · Versiones

| variante | runtime de Lambda | versión del runtime (INIT_START) | compilado con |
|---|---|---|---|
| Go | `provided.al2023` | `provided:al2023.AmazonLinux2023.v228` | `golang:1.27.2` |
| Rust | `provided.al2023` | `provided:al2023.AmazonLinux2023.v228` | `rust:1.99.1-trixie` |
| .NET Native AOT | `provided.al2023` | `provided:al2023.AmazonLinux2023.v228` | `public.ecr.aws/sam/build-dotnet10:latest-x86_64` |
| Java GraalVM native | `provided.al2023` | `provided:al2023.AmazonLinux2023.v228` | `ghcr.io/graalvm/native-image-community:25` |
| Java (JVM) | `java25` | `java:25.mainline.v89` | `eclipse-temurin:25-jdk` |
| Java + SnapStart | `java25` | `java:25.mainline.v89` | `eclipse-temurin:25-jdk` |
| Python | `python3.14` | `python:3.14.mainline.v118` | — (se sube el fuente) |
| Node | `nodejs24.x` | `nodejs:24.mainline.v133` | — (se sube el fuente) |

Lo que imprimieron las herramientas al compilar:

```
# go
go version go1.27.2 linux/amd64
require github.com/aws/aws-lambda-go v1.54.0
# rust
rustc 1.99.1 (58cf8f9c3 2026-10-13)
cargo 1.99.1 (5f94df478 2026-08-27)
   Compiling lambda_runtime v1.5.0
   Compiling bootstrap v0.1.0 (/tmp/w)
    Finished `release` profile [optimized] target(s) in 6.88s
name = "lambda_runtime"
version = "1.5.0"
--
name = "serde"
version = "1.0.229"
--
name = "tokio"
version = "1.54.1"
# dotnet-aot
10.0.404
PRETTY_NAME="Amazon Linux 2023.12.20261109"
  Funcion -> /tmp/w/bin/Release/net10.0/linux-x64/bootstrap.dll
  Generating native code
  Funcion -> /tmp/pub/
   > Amazon.Lambda.Core                                 *             3.4.0
   > Amazon.Lambda.RuntimeSupport                       *             2.3.1
   > Amazon.Lambda.Serialization.SystemTextJson         *             3.0.1
# java-native
openjdk version "25.0.2" 2026-01-20
OpenJDK Runtime Environment GraalVM CE 25.0.2+10.1 (build 25.0.2+10-jvmci-b01)
OpenJDK 64-Bit Server VM GraalVM CE 25.0.2+10.1 (build 25.0.2+10-jvmci-b01, mixed mode, sharing)
native-image 25.0.2 2026-01-20
GraalVM Runtime Environment GraalVM CE 25.0.2+10.1 (build 25.0.2+10-jvmci-b01)
Substrate VM GraalVM CE 25.0.2+10.1 (build 25.0.2+10, serial gc)
Build artifacts:
 /out/bootstrap (executable)
========================================================================================================================
Finished generating 'bootstrap' in 13.6s.
# java
openjdk version "25.0.5" 2026-10-20 LTS
OpenJDK Runtime Environment Temurin-25.0.5+8 (build 25.0.5+8-LTS)
OpenJDK 64-Bit Server VM Temurin-25.0.5+8 (build 25.0.5+8-LTS, mixed mode, sharing)
```

---

## 1 · ¿Devuelven lo mismo? (antes de medir nada)

Referencia calculada aparte: suma de `i*i % 1000003` para i = 1…200.000 = **99863963591**.

| variante | n=200000 → suma | n=1000 → suma | io_ms=100 → io_ms | ¿igual? |
|---|---:|---:|---:|---|
| Go | 99863963591 | 333833500 | 100 | ✅ |
| Rust | 99863963591 | 333833500 | 100 | ✅ |
| .NET Native AOT | 99863963591 | 333833500 | 100 | ✅ |
| Java GraalVM native | 99863963591 | 333833500 | 100 | ✅ |
| Java (JVM) | 99863963591 | 333833500 | 100 | ✅ |
| Java + SnapStart | 99863963591 | 333833500 | 100 | ✅ |
| Python | 99863963591 | 333833500 | 100 | ✅ |
| Node | 99863963591 | 333833500 | 100 | ✅ |

> Sumas distintas para n=1000 entre todas las variantes: **1** (1 = todas iguales).

Respuesta literal de cada una (n=200000):

```
{"lang":"go","n":200000,"suma":99863963591,"io_ms":0}
{"lang":"rust","n":200000,"suma":99863963591,"io_ms":0}
{"lang":"dotnet-aot","n":200000,"suma":99863963591,"io_ms":0}
{"lang":"java-native","n":200000,"suma":99863963591,"io_ms":0}
{"lang":"java","n":200000,"suma":99863963591,"io_ms":0}
{"lang":"java","n":200000,"suma":99863963591,"io_ms":0}
{"lang":"python","n":200000,"suma":99863963591,"io_ms":0}
{"lang":"node","n":200000,"suma":99863963591,"io_ms":0}
```

---

## 2 · Lo que se sube: el .zip

`CodeSize` es lo que devuelve Lambda al crear la función. «Compilar» es el `docker run` entero, en un
contenedor nuevo, con la imagen ya bajada pero **bajando las dependencias** (no hay caché), como en un CI
limpio. Dos rondas (`python banco.py cronometrar`, con el candado de la fábrica tomado: nadie más compilaba).

| variante | sin comprimir | .zip (CodeSize) | compilar, ronda 1 | ronda 2 |
|---|---:|---:|---:|---:|
| Go | 6.807.712 B (6.49 MiB) | 2.878.957 B (2.75 MiB) | 5.2 s | 4.7 s |
| Rust | 1.796.648 B (1.71 MiB) | 787.255 B (768.8 KiB) | 15.5 s | 12.9 s |
| .NET Native AOT | 7.786.288 B (7.43 MiB) | 3.645.357 B (3.48 MiB) | 17.0 s | 14.0 s |
| Java GraalVM native | 17.370.184 B (16.57 MiB) | 5.975.390 B (5.70 MiB) | 15.8 s | 14.4 s |
| Java (JVM) | 1.505 B (1.5 KiB) | 982 B (1.0 KiB) | 0.8 s | 0.6 s |
| Python | 313 B (0.3 KiB) | 294 B (0.3 KiB) | — | — |
| Node | 354 B (0.3 KiB) | 338 B (0.3 KiB) | — | — |

> Java + SnapStart sube el mismo .zip que Java (JVM).

---

## 3 · Arranque en frío — 25 entornos nuevos por variante

«Init» es `Init Duration` del REPORT; en SnapStart no hay Init: es `Restore Duration` (restaurar la
instantánea). «Duración» es la de esa primera invocación (200.000 vueltas). Con 25 muestras, el p99 es el máximo.

| variante | fríos confirmados | init/restore mediana | p90 | p99 | mín | duración 1ª | facturada mediana | init + duración mediana | punta a punta mediana |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| Go | 25/25 | 71.7 ms | 94.6 ms | 100.8 ms | 57.1 ms | 2.1 ms | 75 ms | 74.5 ms | 321 ms |
| Rust | 25/25 | 27.7 ms | 33.2 ms | 70.3 ms | 23.8 ms | 1.6 ms | 30 ms | 29.4 ms | 268 ms |
| .NET Native AOT | 25/25 | 87.9 ms | 127.9 ms | 157.8 ms | 67.3 ms | 2.6 ms | 91 ms | 90.4 ms | 339 ms |
| Java GraalVM native | 25/25 | 125.6 ms | 153.1 ms | 203.9 ms | 87.6 ms | 5.5 ms | 133 ms | 132.1 ms | 412 ms |
| Java (JVM) | 25/25 | 302.1 ms | 363.7 ms | 485.6 ms | 257.2 ms | 152.2 ms | 450 ms | 449.9 ms | 697 ms |
| Java + SnapStart | 25/25 | 641.8 ms | 776.6 ms | 850.4 ms | 458.7 ms | 401.6 ms | 563 ms | 1041.6 ms | 1277 ms |
| Python | 25/25 | 106.2 ms | 129.3 ms | 160.4 ms | 82.6 ms | 105.3 ms | 213 ms | 212.8 ms | 439 ms |
| Node | 25/25 | 146.0 ms | 191.4 ms | 230.0 ms | 111.0 ms | 122.5 ms | 269 ms | 268.3 ms | 504 ms |

Ordenadas por «init + duración» (lo que espera la primera petición dentro de AWS):

```
Rust                       29.4 ms   (init/restore 27.7 + duración 1.6)
Go                         74.5 ms   (init/restore 71.7 + duración 2.1)
.NET Native AOT            90.4 ms   (init/restore 87.9 + duración 2.6)
Java GraalVM native       132.1 ms   (init/restore 125.6 + duración 5.5)
Python                    212.8 ms   (init/restore 106.2 + duración 105.3)
Node                      268.3 ms   (init/restore 146.0 + duración 122.5)
Java (JVM)                449.9 ms   (init/restore 302.1 + duración 152.2)
Java + SnapStart         1041.6 ms   (init/restore 641.8 + duración 401.6)
```

¿Se cobra el arranque? La «facturada» de la invocación fría contra init + duración:

| variante | init/restore mediana | duración mediana | facturada mediana | Billed Restore Duration mediana |
|---|---:|---:|---:|---:|
| Go | 71.7 ms | 2.1 ms | 75 ms | — |
| Rust | 27.7 ms | 1.6 ms | 30 ms | — |
| .NET Native AOT | 87.9 ms | 2.6 ms | 91 ms | — |
| Java GraalVM native | 125.6 ms | 5.5 ms | 133 ms | — |
| Java (JVM) | 302.1 ms | 152.2 ms | 450 ms | — |
| Java + SnapStart | 641.8 ms | 401.6 ms | 563 ms | 142 ms |
| Python | 106.2 ms | 105.3 ms | 213 ms | — |
| Node | 146.0 ms | 122.5 ms | 269 ms | — |

**Comprobación aparte: ¿es el tamaño del .zip?** El MISMO binario de Rust, con un archivo de relleno aleatorio
(2.091.702 B, incompresible) para que el .zip pese lo que el de Go. 25 arranques en frío:

| variante | .zip (CodeSize) | init mediana | p90 | mín |
|---|---:|---:|---:|---:|
| Rust | 787.255 B | 27.7 ms | 33.2 ms | 23.8 ms |
| Rust + relleno | 2.879.695 B | 28.2 ms | 36.9 ms | 23.7 ms |
| Go | 2.878.957 B | 71.7 ms | 94.6 ms | 57.1 ms |

> Con el mismo tamaño de .zip, Rust sigue arrancando como Rust: el peso del paquete NO explica el init de Go.
> Por qué Go tarda más en inicializarse no se midió aquí.

**Ráfaga: 20 invocaciones a la vez.** Cada una retiene su entorno 1,5 s (`io_ms=1500`), así que casi
todas necesitan un entorno nuevo, y nacen todos al mismo tiempo. En Java + SnapStart va contra la última versión
publicada, cuya instantánea ya se había restaurado antes (no recién publicada, como en la tabla de arriba).

| variante | entornos nuevos | init/restore mediana | p90 | máx. | mín. |
|---|---:|---:|---:|---:|---:|
| Go | 18/20 | 71.3 ms | 85.1 ms | 109.1 ms | 58.9 ms |
| Rust | 19/20 | 28.9 ms | 36.8 ms | 37.1 ms | 24.7 ms |
| .NET Native AOT | 19/20 | 89.2 ms | 99.2 ms | 118.7 ms | 69.7 ms |
| Java GraalVM native | 18/20 | 117.5 ms | 139.6 ms | 148.2 ms | 95.4 ms |
| Java (JVM) | 19/20 | 278.3 ms | 403.1 ms | 414.0 ms | 243.1 ms |
| Java + SnapStart | 19/20 | 412.6 ms | 477.6 ms | 554.0 ms | 266.6 ms |
| Python | 19/20 | 124.5 ms | 162.4 ms | 182.7 ms | 82.3 ms |
| Node | 18/20 | 137.8 ms | 192.0 ms | 193.8 ms | 108.2 ms |

Una línea REPORT de cada variante, tal cual (la primera invocación fría):

```
# Go
REPORT RequestId: 97db4cfc-6d51-4298-b31e-e0480deb3d98  Duration: 2.68 ms  Billed Duration: 62 ms  Memory Size: 512 MB  Max Memory Used: 19 MB  Init Duration: 58.47 ms
# Rust
REPORT RequestId: 3054f58c-30d1-4e2d-be92-d17383d65e46  Duration: 1.46 ms  Billed Duration: 29 ms  Memory Size: 512 MB  Max Memory Used: 14 MB  Init Duration: 27.12 ms
# .NET Native AOT
REPORT RequestId: 988fccdd-cc62-4d61-90f6-ca38fc248d2e  Duration: 2.84 ms  Billed Duration: 104 ms  Memory Size: 512 MB  Max Memory Used: 21 MB  Init Duration: 100.58 ms
# Java GraalVM native
REPORT RequestId: c10c63e5-a4b4-42c6-b4c1-fc5d65dc8989  Duration: 5.45 ms  Billed Duration: 128 ms  Memory Size: 512 MB  Max Memory Used: 26 MB  Init Duration: 121.90 ms
# Java (JVM)
REPORT RequestId: 66c72375-914d-41f3-ae7e-f05d17e2d0b2  Duration: 175.62 ms  Billed Duration: 540 ms  Memory Size: 512 MB  Max Memory Used: 87 MB  Init Duration: 363.68 ms
# Java + SnapStart
REPORT RequestId: 93c385c2-5530-4c6b-82c8-6b4987ad5aae  Duration: 323.18 ms  Billed Duration: 440 ms  Memory Size: 512 MB  Max Memory Used: 95 MB  Restore Duration: 462.47 ms  Billed Restore Duration: 116 ms
# Python
REPORT RequestId: 35e3b2b5-c217-4189-b15a-f03d9a576b07  Duration: 99.12 ms  Billed Duration: 204 ms  Memory Size: 512 MB  Max Memory Used: 34 MB  Init Duration: 104.28 ms
# Node
REPORT RequestId: 861acf51-d467-4424-bf03-625736003fb9  Duration: 146.18 ms  Billed Duration: 377 ms  Memory Size: 512 MB  Max Memory Used: 82 MB  Init Duration: 229.99 ms
```

<details><summary>todas las medidas de arranque, una a una (ms)</summary>

```
go           init/restore  58.5  76.1  70.6  85.2  100.8  66.4  71.0  78.6  78.7  77.7  66.1  81.0  71.7  71.3  96.6  94.6  65.0  73.4  80.4  69.6  69.9  74.1  68.2  57.1  71.2
             punta a punta 310  362  304  325  394  312  321  321  307  346  301  337  693  315  358  369  293  300  368  302  323  317  345  310  346
rust         init/restore  27.1  27.8  70.3  23.8  25.9  31.6  27.7  25.7  27.9  27.1  27.5  30.5  25.6  33.2  26.9  27.1  27.2  28.5  29.1  27.5  55.8  27.8  30.3  27.6  27.8
             punta a punta 276  258  335  267  271  273  263  252  254  256  265  259  248  287  386  251  268  271  279  277  418  260  261  276  293
dotnet-aot   init/restore  100.6  78.9  109.4  85.7  157.8  133.4  80.6  92.1  87.9  88.4  100.4  90.4  67.3  120.9  84.2  96.0  127.9  80.1  79.9  81.5  74.3  82.0  74.1  91.9  87.6
             punta a punta 346  317  386  327  515  426  403  350  330  453  347  334  309  425  313  339  416  380  303  318  322  319  294  349  324
java-native  init/restore  121.9  106.0  113.1  119.5  164.2  87.6  130.5  126.8  129.1  115.8  142.5  203.9  120.8  119.0  113.9  151.2  125.6  142.2  134.3  106.1  137.2  108.5  153.1  137.9  110.1
             punta a punta 373  391  366  450  445  323  418  405  412  424  518  484  396  403  398  433  499  417  428  373  417  385  413  402  367
java         init/restore  363.7  267.3  300.9  452.3  342.2  314.0  301.3  257.2  295.3  285.3  267.7  295.1  337.2  257.9  321.6  315.7  268.8  485.6  305.5  302.1  307.4  282.7  313.1  288.1  321.8
             punta a punta 843  672  697  970  782  706  684  804  684  686  667  717  741  655  709  719  674  856  721  696  663  653  696  680  716
java-snap    init/restore  462.5  458.7  619.6  703.0  684.3  602.4  732.8  689.1  675.0  631.5  822.1  661.2  641.8  776.6  850.4  633.0  620.5  622.1  601.0  632.2  627.3  682.0  668.8  605.4  685.9
             punta a punta 992  1070  1225  1327  1487  1249  1363  1227  1277  1321  1810  1305  1218  1390  1511  1303  1310  1182  1263  1272  1304  1269  1304  1239  1257
python       init/restore  104.3  115.5  102.6  86.7  99.3  119.4  105.7  123.6  106.2  82.6  116.4  109.8  159.7  160.4  107.3  97.7  97.5  85.8  129.3  99.0  103.7  115.4  109.8  93.2  114.6
             punta a punta 426  456  410  420  430  454  631  473  448  396  457  430  525  551  427  426  389  419  519  443  413  460  439  418  446
node         init/restore  230.0  111.0  188.5  191.4  138.5  139.1  131.4  132.9  163.2  141.8  166.3  153.3  202.9  128.3  148.1  154.1  135.3  139.9  136.4  137.2  146.0  152.3  165.3  168.2  142.6
             punta a punta 645  448  623  566  471  473  471  471  536  498  515  512  612  445  528  610  479  465  510  482  481  504  536  553  484
```

</details>

---

## 4 · En caliente — 200 invocaciones con 200.000 vueltas de CPU

| variante | duración p50 | duración p99 | facturada p50 | facturada p99 | memoria máx. usada | punta a punta p50 | p99 | invocaciones con arranque |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Go | 1.49 ms | 3.35 ms | 2 ms | 4 ms | 22 MB | 107 ms | 127 ms | 0 |
| Rust | 1.13 ms | 3.49 ms | 2 ms | 4 ms | 14 MB | 110 ms | 141 ms | 0 |
| .NET Native AOT | 1.51 ms | 3.72 ms | 2 ms | 4 ms | 24 MB | 107 ms | 128 ms | 0 |
| Java GraalVM native | 1.40 ms | 2.47 ms | 2 ms | 3 ms | 43 MB | 113 ms | 127 ms | 0 |
| Java (JVM) | 3.87 ms | 6.74 ms | 4 ms | 7 ms | 87 MB | 114 ms | 140 ms | 0 |
| Java + SnapStart | 1.63 ms | 6.51 ms | 2 ms | 7 ms | 96 MB | 111 ms | 135 ms | 0 |
| Python | 94.43 ms | 107.71 ms | 95 ms | 108 ms | 34 MB | 201 ms | 233 ms | 0 |
| Node | 58.48 ms | 71.64 ms | 59 ms | 72 ms | 84 MB | 162 ms | 219 ms | 0 |

```
Rust                   p50     1.13 ms   facturada     2 ms   memoria 14 MB
Java GraalVM native    p50     1.40 ms   facturada     2 ms   memoria 43 MB
Go                     p50     1.49 ms   facturada     2 ms   memoria 22 MB
.NET Native AOT        p50     1.51 ms   facturada     2 ms   memoria 24 MB
Java + SnapStart       p50     1.63 ms   facturada     2 ms   memoria 96 MB
Java (JVM)             p50     3.87 ms   facturada     4 ms   memoria 87 MB
Node                   p50    58.48 ms   facturada    59 ms   memoria 84 MB
Python                 p50    94.43 ms   facturada    95 ms   memoria 34 MB
```

---

## 5 · El matiz del 121, en AWS — 100 invocaciones esperando 100 ms de E/S (sin CPU)

| variante | duración p50 | duración p99 | facturada p50 | punta a punta p50 |
|---|---:|---:|---:|---:|
| Go | 101.65 ms | 104.01 ms | 102 ms | 208 ms |
| Rust | 102.68 ms | 105.31 ms | 103 ms | 210 ms |
| .NET Native AOT | 101.42 ms | 104.01 ms | 102 ms | 206 ms |
| Java GraalVM native | 101.79 ms | 103.69 ms | 102 ms | 211 ms |
| Java (JVM) | 101.69 ms | 103.62 ms | 102 ms | 213 ms |
| Java + SnapStart | 101.67 ms | 102.43 ms | 102 ms | 210 ms |
| Python | 101.91 ms | 102.47 ms | 102 ms | 205 ms |
| Node | 101.83 ms | 102.94 ms | 102 ms | 206 ms |

> Entre la mejor y la peor p50 esperando 100 ms: **1.26 ms**.

---

## 6 · La CPU la reparte la memoria — el mismo trabajo con 128, 512 y 1024 MB

Duración p50 de 50 invocaciones calientes (200.000 vueltas) con cada memoria; la de 512 MB es la de §4.

| variante | 128 MB | 512 MB | 1024 MB | init a 128 MB | init a 1024 MB |
|---|---:|---:|---:|---:|---:|
| Go | 1.32 ms | 1.49 ms | 1.31 ms | 66.1 ms | 61.4 ms |
| Rust | 1.42 ms | 1.13 ms | 1.49 ms | 43.0 ms | 30.0 ms |
| .NET Native AOT | 2.10 ms | 1.51 ms | 1.54 ms | 105.5 ms | 82.8 ms |
| Java GraalVM native | 1.75 ms | 1.40 ms | 1.48 ms | 112.5 ms | 108.5 ms |
| Java (JVM) | 17.29 ms | 3.87 ms | 4.17 ms | 308.3 ms | 305.3 ms |
| Python | 434.75 ms | 94.43 ms | 50.99 ms | 81.9 ms | 114.4 ms |
| Node | 288.67 ms | 58.48 ms | 33.73 ms | 152.1 ms | 155.8 ms |

> «init a 128/1024 MB» es UNA sola medida (el primer arranque tras cambiar la memoria), no una mediana.

---

## 7 · El 121 (emulador RIE en mi PC) contra AWS de verdad

Los números del 121 salen de `recursos/pruebas-lambda/SALIDA.json` (8 arranques, 200 invocaciones; Python 3.13 y
Node 22 allí, Python 3.14 y Node 24 aquí). La duración es la del REPORT en los dos.

| lenguaje | init 121 (emulador) | init AWS | × | caliente p50 121 | caliente p50 AWS | × | facturada p50 121 | facturada p50 AWS |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| Go | 2.8 ms | 71.7 ms | 25.8 | 1.42 ms | 1.49 ms | 1.1 | 0.43 ms | 2 ms |
| Python | 21.5 ms | 106.2 ms | 4.9 | 9.77 ms | 94.43 ms | 9.7 | 8.31 ms | 95 ms |
| Node | 32.3 ms | 146.0 ms | 4.5 | 7.24 ms | 58.48 ms | 8.1 | 5.75 ms | 59 ms |

> ⚠️ «caliente p50 121» es lo que el 121 llamó «p50» (cronometrado desde el cliente, en local); la facturada del 121
> era la duración del REPORT del emulador. Aquí la duración es la del REPORT de Lambda con 512 MB.

> Orden de arranque en el 121: Go < Python < Node. En AWS: Go < Python < Node.

---

## 8 · La factura

⛔ Los precios **no se miden: se copian de AWS**. Lo nuestro es la aritmética y la duración facturada medida arriba.

**Fuente:** https://aws.amazon.com/lambda/pricing/ (us-east-1, x86, consultado 2026-11-18)

- 0.20 $ por millón de peticiones
- 0.0000166667 $ por GB-segundo (x86)
- capa gratuita al mes: 1.000.000 de peticiones y 400.000 GB-s (aquí NO se descuenta)
- SnapStart: «SnapStart pricing does not apply to supported Java managed runtimes»

Un millón de invocaciones **en caliente** con 512 MB, con la facturada p50 de §4; y lo que cuesta un millón
de **arranques en frío**, con la facturada mediana de §3 (que ya incluye el init cuando Lambda lo cobra):

| variante | facturada p50 caliente | $ por millón en caliente | facturada fría mediana | $ por millón de arranques en frío |
|---|---:|---:|---:|---:|
| Go | 2 ms | 0.22 $ | 75 ms | 0.83 $ |
| Rust | 2 ms | 0.22 $ | 30 ms | 0.45 $ |
| .NET Native AOT | 2 ms | 0.22 $ | 91 ms | 0.96 $ |
| Java GraalVM native | 2 ms | 0.22 $ | 133 ms | 1.31 $ |
| Java + SnapStart | 2 ms | 0.22 $ | 563 ms | 4.89 $ |
| Java (JVM) | 4 ms | 0.23 $ | 450 ms | 3.95 $ |
| Node | 59 ms | 0.69 $ | 269 ms | 2.44 $ |
| Python | 95 ms | 0.99 $ | 213 ms | 1.98 $ |

> Solo peticiones, sin cómputo: 0.20 $ por millón.

La misma cuenta en caliente con 128 y 1024 MB (facturada p50 de §6):

| variante | 128 MB: facturada p50 | $ / millón | 512 MB: $ / millón | 1024 MB: facturada p50 | $ / millón |
|---|---:|---:|---:|---:|---:|
| Go | 2 ms | 0.20 $ | 0.22 $ | 2 ms | 0.23 $ |
| Rust | 2 ms | 0.20 $ | 0.22 $ | 2 ms | 0.23 $ |
| .NET Native AOT | 3 ms | 0.21 $ | 0.22 $ | 2 ms | 0.23 $ |
| Java GraalVM native | 2 ms | 0.20 $ | 0.22 $ | 2 ms | 0.23 $ |
| Java (JVM) | 18 ms | 0.24 $ | 0.23 $ | 5 ms | 0.28 $ |
| Python | 435 ms | 1.11 $ | 0.99 $ | 51 ms | 1.05 $ |
| Node | 289 ms | 0.80 $ | 0.69 $ | 34 ms | 0.77 $ |

**La factura del 121, rehecha con lo que factura AWS de verdad.** El 121 calculó con la duración del emulador en mi PC.
Mismos precios, mismas invocaciones; aquí con la facturada p50 medida en Lambda con esa memoria:

| lenguaje | invocaciones/mes | memoria | facturada 121 | total/mes 121 | facturada AWS | total/mes AWS |
|---|---:|---:|---:|---:|---:|---:|
| Go | 10.000.000 | 128 MB | 1 ms | 1.80 $ | 2 ms | 1.80 $ |
| Python | 10.000.000 | 128 MB | 9 ms | 1.80 $ | 435 ms | 4.20 $ |
| Node | 10.000.000 | 128 MB | 6 ms | 1.80 $ | 289 ms | 1.80 $ |
| Go | 1.000.000.000 | 128 MB | 1 ms | 199.80 $ | 2 ms | 199.80 $ |
| Python | 1.000.000.000 | 128 MB | 9 ms | 211.88 $ | 435 ms | 1099.39 $ |
| Node | 1.000.000.000 | 128 MB | 6 ms | 205.63 $ | 289 ms | 795.22 $ |
| Go | 1.000.000.000 | 1024 MB | 1 ms | 209.80 $ | 2 ms | 226.47 $ |
| Python | 1.000.000.000 | 1024 MB | 9 ms | 343.13 $ | 51 ms | 1043.14 $ |
| Node | 1.000.000.000 | 1024 MB | 6 ms | 293.13 $ | 34 ms | 759.80 $ |

> Esta tabla, como la del 121, SÍ descuenta la capa gratuita (1 M de peticiones y 400.000 GB-s al mes).
> 10.000.000 invocaciones a 128 MB, Python menos Go: el 121 decía **0.00 $**; con AWS de verdad, **2.40 $**.
> 1.000.000.000 invocaciones a 128 MB, Python menos Go: el 121 decía **12.08 $**; con AWS de verdad, **899.59 $**.
> 1.000.000.000 invocaciones a 1024 MB, Python menos Go: el 121 decía **133.33 $**; con AWS de verdad, **816.67 $**.


---

## 9 · Solo Python, Node, Java y .NET (para la versión de Cultura DevOps, sin Go)

Mismas medidas que arriba, separadas. Nada nuevo: es un subconjunto de §2-§8.

| variante | .zip | init/restore mediana | p90 | p99 | init + duración | caliente p50 | caliente p99 | facturada p50 | memoria | $ / millón caliente |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| Python | 0.3 KiB | 106.2 ms | 129.3 ms | 160.4 ms | 212.8 ms | 94.43 ms | 107.71 ms | 95 ms | 34 MB | 0.99 $ |
| Node | 0.3 KiB | 146.0 ms | 191.4 ms | 230.0 ms | 268.3 ms | 58.48 ms | 71.64 ms | 59 ms | 84 MB | 0.69 $ |
| Java (JVM) | 1.0 KiB | 302.1 ms | 363.7 ms | 485.6 ms | 449.9 ms | 3.87 ms | 6.74 ms | 4 ms | 87 MB | 0.23 $ |
| Java + SnapStart | 1.0 KiB | 641.8 ms | 776.6 ms | 850.4 ms | 1041.6 ms | 1.63 ms | 6.51 ms | 2 ms | 96 MB | 0.22 $ |
| .NET Native AOT | 3.48 MiB | 87.9 ms | 127.9 ms | 157.8 ms | 90.4 ms | 1.51 ms | 3.72 ms | 2 ms | 24 MB | 0.22 $ |

---

## 10 · Lo que NO se midió

- arm64 (Graviton): todo es x86_64.
- SnapStart con «priming» (ejecutar código antes de la instantánea): el Java + SnapStart de aquí es el handler tal cual.
- Imágenes de contenedor: todo se subió como .zip.
- Concurrencia dentro de un mismo entorno: salvo la ráfaga de §3, las invocaciones de cada función van de una en una.
- Llamadas reales a otros servicios (DynamoDB, S3): la E/S es una espera de 100 ms.
- Los primeros arranques tras publicar una versión de SnapStart pueden ir más lentos mientras AWS calienta su caché de
  instantáneas; aquí cada arranque es de una versión recién publicada, o sea, el peor caso.

---

## 11 · Limpieza: no queda nada `gl-lambda-*` en la cuenta

Hecho el 2026-11-18T15:44:55+00:00 por `python banco.py limpiar`:

```
lambda delete-function gl-lambda-go
lambda delete-function gl-lambda-rust
lambda delete-function gl-lambda-dotnet-aot
lambda delete-function gl-lambda-java-native
lambda delete-function gl-lambda-java
lambda delete-function gl-lambda-java-snap
lambda delete-function gl-lambda-python
lambda delete-function gl-lambda-node
lambda delete-function gl-lambda-rust-relleno
iam detach-role-policy gl-lambda-rol AWSLambdaBasicExecutionRole
iam delete-role gl-lambda-rol
logs delete-log-group /aws/lambda/gl-lambda-dotnet-aot
logs delete-log-group /aws/lambda/gl-lambda-go
logs delete-log-group /aws/lambda/gl-lambda-java
logs delete-log-group /aws/lambda/gl-lambda-java-native
logs delete-log-group /aws/lambda/gl-lambda-java-snap
logs delete-log-group /aws/lambda/gl-lambda-node
logs delete-log-group /aws/lambda/gl-lambda-python
logs delete-log-group /aws/lambda/gl-lambda-rust
logs delete-log-group /aws/lambda/gl-lambda-rust-relleno
```

Comprobación posterior, mirando donde vive cada cosa:

```
lambda list-functions (nombres que empiezan por gl-lambda-) → 0 []
iam get-role gl-lambda-rol → NoSuchEntity
iam list-roles (nombres que empiezan por gl-lambda-) → 0 []
logs describe-log-groups --log-group-name-prefix /aws/lambda/gl-lambda- → 0 []
fallos al borrar → 0 []
```

Comprobación repetida el 2026-11-18T15:47:17+00:00 (`python banco.py verificar`, sin borrar nada):

```
lambda list-functions (gl-lambda-*) → 0 []
iam get-role gl-lambda-rol → NoSuchEntity
iam list-roles (gl-lambda-*) → 0 []
logs describe-log-groups /aws/lambda/gl-lambda- → 0 []
```
