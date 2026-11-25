# SALIDA — Go para administradores de sistemas (ejecutado)

Generado por `python correr.py`. **No se edita a mano.**

- fecha: 2026-10-12 01:32 UTC
- imagen: golang:1.27.2 (Docker, linux/amd64)
- máquina: 24 · Intel(R) Core(TM) Ultra 7 270K Plus núcleos/CPU
- los tiempos son de ESTA máquina, dentro de Docker en Windows.

## La versión de Go

`$ go version`

```
go version go1.27.2 linux/amd64
```

## 1 · Bash + curl, uno detrás de otro

`$ bash revisar.sh`

```
CAÍDO  127.0.0.1:9013
CAÍDO  127.0.0.1:9027
CAÍDO  127.0.0.1:9041
50 servidores · 47 bien · 3 caídos · 14.4 s en total
```

## 1b · Bash también puede ir en paralelo

`$ bash paralelo.sh   # xargs -P 50 curl -m 2`

```
3 caídos · 2.0 s en total
```

## 2 · Go, todos a la vez

`$ go run ./cmd/revisar`

```
CAÍDO  127.0.0.1:9013  (sin respuesta)
CAÍDO  127.0.0.1:9027  (sin respuesta)
CAÍDO  127.0.0.1:9041  (sin respuesta)
50 servidores · 47 bien · 3 caídos · 2.0 s en total
exit status 1
```

## 2b · Como mucho 10 a la vez

`$ ./revisar -max 10`

```
CAÍDO  127.0.0.1:9013  (sin respuesta)
CAÍDO  127.0.0.1:9027  (sin respuesta)
CAÍDO  127.0.0.1:9041  (sin respuesta)
50 servidores · 47 bien · 3 caídos · 2.6 s en total
```

## 2c · Para otra herramienta, y para cron

`$ ./revisar -json 2>/dev/null | head -c 160`

```
[{"host":"127.0.0.1:9001","ok":true,"ms":112.066},{"host":"127.0.0.1:9002","ok":true,"ms":149.905},{"host":"127.0.0.1:9003","ok":true,"ms":187.298},{"host":"127 …
```

`$ ./revisar > /dev/null 2>&1; echo "código de salida: $?"`

```
código de salida: 1
```

## 3 · El binario

`$ CGO_ENABLED=0 go build -o revisar ./cmd/revisar && ls -lh revisar && ldd revisar`

```
9.5M revisar
	not a dynamic executable
```

`$ go build -ldflags='-s -w' -o revisar-min ./cmd/revisar && ls -lh revisar-min`

```
6.5M revisar-min
```

## 4 · Compilación cruzada

`$ GOOS=linux GOARCH=arm64 go build -o revisar-arm64 ./cmd/revisar && GOOS=windows GOARCH=amd64 go build -o revisar.exe ./cmd/revisar && ls -lh revisar-arm64 revisar.exe`

```
8.8M revisar-arm64
9.6M revisar.exe
```

`$ go tool dist list | wc -l`

```
47
```

## 5 · En un servidor sin Go

`$ docker run httpd:2.4-alpine  →  which go; ./revisar`

```
go: no está instalado
CAÍDO  127.0.0.1:9013  (sin respuesta)
CAÍDO  127.0.0.1:9027  (sin respuesta)
CAÍDO  127.0.0.1:9041  (sin respuesta)
50 servidores · 47 bien · 3 caídos · 2.0 s en total
```
