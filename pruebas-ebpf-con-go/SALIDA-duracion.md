# SALIDA — eBPF con Go — duración y filtro por proceso (ejecutado)

Generado por `python correr_duracion.py`. **No se edita a mano.**

- fecha: 2026-10-11 02:38 UTC
- imagen: fab-go-ebpf:1.27.2 (Docker, linux/amd64)
- máquina: 24 · Intel(R) Core(TM) Ultra 7 270K Plus núcleos/CPU
- los tiempos son de ESTA máquina, dentro de Docker en Windows.

## Compilar

`$ go generate && go build -o duraciones .`

```
7088 duraciones_bpfel.o
```

## Cuánto dura cada conexión

`$ ./duraciones -durante 8s   # y curl a /, /?ms=300, /?ms=1200, a 1.1.1.1 y al puerto 9`

```
curl :9 código 7
escuchando 8s (pid 0)…
curl     pid 89     → 127.0.0.1:8080  conectar 34µs      total 0s
curl     pid 90     → 127.0.0.1:8080  conectar 16µs      total 301ms
curl     pid 91     → 127.0.0.1:8080  conectar 19µs      total 1.201s
curl     pid 92     → 1.1.1.1:443   conectar 7.524ms   total 46ms
curl     pid 93     → 127.0.0.1:9     conectar 0s        total 0s
```

## Filtrar por proceso: el PID que da el kernel, sin traducir

`$ ./duraciones -durante 6s -pid $ELEGIDO -sin-ns`

```
pid del curl elegido: 87
escuchando 6s (pid 87)…
```

`$ ./duraciones -sin-ns   # y un curl, con su pid según el contenedor`

```
pid del curl según el contenedor: 87
curl     pid 6249   → 127.0.0.1:8080  conectar 44µs      total 0s
```

## Lo mismo con --pid=host (Docker Desktop)

`$ docker run --pid=host ... ./duraciones -sin-ns   # y un curl`

```
pid del curl según el contenedor: 5926
curl     pid 6401   → 127.0.0.1:8080  conectar 26µs      total 0s
```

## Filtrar por proceso: con el espacio de PIDs propio

`$ ./duraciones -durante 6s -pid $ELEGIDO   # y, mientras, curl a 1.1.1.1 y otro al 8080`

```
pid del curl elegido: 86
escuchando 6s (pid 86)…
curl     pid 86     → 127.0.0.1:8080  conectar 32µs      total 101ms
curl     pid 86     → 127.0.0.1:8080  conectar 33µs      total 201ms
curl     pid 86     → 127.0.0.1:8080  conectar 19µs      total 301ms
curl     pid 86     → 127.0.0.1:8080  conectar 33µs      total 401ms
curl     pid 86     → 127.0.0.1:8080  conectar 31µs      total 501ms
```
