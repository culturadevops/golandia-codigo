# SALIDA — ¿Go o Pyronaut? (golandia; pedido en un comentario del 848)

> Corrida del 2026-11-23 07:05. Regenerable con `python correr.py`. ⛔ Ninguna cifra de blog: todo salió de aquí. El banco del 848 (misma lógica, mismo servidor Go, mismas condiciones) con Pyronaut añadido.

- Pyronaut (`pyronaut --version` en la etapa de construcción):

```
Pyronaut: 0.1.2
Micronaut Core: 5.2.22
Micronaut Platform: 5.2.5
GraalPy: 25.4.4.1.1
GraalPy Interpreter: graalpy3.13-25.4.4
Native Image JDK: 25
```
- JVM de la imagen: `java version "25.0.4.1.1" 2026-08-18 LTS`
- Resto, sin tocar desde el 848: `golandia-govs-go:1` (go1.27.2, net/http, FROM scratch), `golandia-rust-py-fastapi:1` (Python 3.13.17, fastapi 0.162.0 + uvicorn 0.54.0, 2 procesos), `golandia-rust-py-robyn:1` (robyn 0.88.0, 2 procesos).
- Servidores con 2 CPU y 1g. Pyronaut: 1 proceso (la JVM), rutas `def` → ejecutor «io» de Micronaut, pool de contextos de GraalPy por defecto. `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen de Pyronaut desde cero (`docker build --no-cache`)

0 s. Incluye `pip install pyronaut` y `pyronaut setup`, que DESCARGA el GraalVM JDK, GraalPy y las dependencias de Maven (en el 848 las imágenes de Python tardaron 8,7-17,7 s y la de Go 4,0 s, con la imagen base ya descargada).

## 2 · Lo que despliegas

| imagen | tamaño (docker image inspect .Size, el mismo dato que el 848) |
|---|---|
| pyronaut (JVM) | 491.5 MiB |
| fastapi | 61.1 MiB |
| robyn | 54.1 MiB |
| go | 2.4 MiB |

Por dentro (`ls -l /app/banco.jar; du -sh /opt/graalvm`):

```
-rw------- 1 root root 97560169 Nov 23 12:01 /app/banco.jar
901M	/opt/graalvm
```

## 0 · ¿Hacen lo mismo? (antes de medir)

El cuerpo de /trabajo?n=20000, /suma?n=200000 y /suma?n=1, comparado byte a byte con go (net/http) quitando solo "lang" (go o python) y "ns" (el reloj del cálculo).

- go (net/http): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591, 2220 + 31 + 16 bytes, application/json → idéntico
- fastapi + uvicorn: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591, 2220 + 31 + 16 bytes, application/json → idéntico
- robyn: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591, 2220 + 31 + 16 bytes, application/json → idéntico
- pyronaut (JVM): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591, 2220 + 31 + 16 bytes, application/json → idéntico

