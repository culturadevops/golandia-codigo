## 6 · /suma?n=1 a tope — vuelta 1

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go (net/http) | 86158 | 87436 | 0.1 ms | 0.5 ms | 12 MiB |
| fastapi + uvicorn | 12137 | 13035 | 1.1 ms | 3.0 ms | 110 MiB |
| fastapi + granian | 11204 | 11484 | 1.1 ms | 17.7 ms | 85 MiB |
| litestar + uvicorn | 15075 | 15091 | 0.9 ms | 3.3 ms | 93 MiB |
| litestar + granian | 13506 | 13447 | 0.9 ms | 17.4 ms | 74 MiB |
| robyn | 63903 | 59502 | 0.3 ms | 0.8 ms | 37 MiB |

## 6 · /suma?n=1 a tope — vuelta 2

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| fastapi + granian | 11158 | 11056 | 1.1 ms | 17.6 ms | 85 MiB |
| litestar + uvicorn | 14783 | 14782 | 1.0 ms | 3.4 ms | 92 MiB |
| litestar + granian | 13185 | 13240 | 1.0 ms | 17.8 ms | 74 MiB |
| robyn | 64614 | 63185 | 0.2 ms | 0.7 ms | 38 MiB |
| go (net/http) | 87089 | 87134 | 0.1 ms | 0.5 ms | 12 MiB |
| fastapi + uvicorn | 11949 | 13048 | 1.1 ms | 2.9 ms | 101 MiB |

## 6 · /suma?n=1 a tope — vuelta 3

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| litestar + granian | 13470 | 13483 | 0.9 ms | 17.7 ms | 75 MiB |
| robyn | 61984 | 50408 | 0.3 ms | 0.9 ms | 38 MiB |
| go (net/http) | 85414 | 86512 | 0.1 ms | 0.5 ms | 12 MiB |
| fastapi + uvicorn | 12250 | 13482 | 1.1 ms | 2.9 ms | 101 MiB |
| fastapi + granian | 11573 | 11031 | 1.1 ms | 18.4 ms | 85 MiB |
| litestar + uvicorn | 14050 | 14011 | 1.0 ms | 3.8 ms | 92 MiB |

## Mediana de las 3 vueltas (templado, pet/s)

| variante | vueltas | mediana |
|---|---|---|
| go (net/http) | 87436, 87134, 86512 | 87134 |
| fastapi + uvicorn | 13035, 13048, 13482 | 13048 |
| fastapi + granian | 11484, 11056, 11031 | 11056 |
| litestar + uvicorn | 15091, 14782, 14011 | 14782 |
| litestar + granian | 13447, 13240, 13483 | 13447 |
| robyn | 59502, 63185, 50408 | 59502 |
