# SALIDA — monitor de uptime en Go (golandia)

> Corrida del 2026-10-21 09:47. golang:1.27.2, solo librería estándar, contenedor con --cpus 2. Tiempos redondeados a 100 ms.
> El «discord» es un servidor falso local que imita el webhook (POST {"content": ...} → 204). No se usa ningún webhook real.

```
$ go version
go version go1.27.2 linux/amd64

$ ./monitor secuencial
  en fila (for):            3.0 s
  una goroutine por sitio: 0.3 s

$ ./monitor ingenuo
  sitios: web (bien) · api (se cae del 1 al 3) · pagos (500 una sí, una no) · lento (tarda 3 s)
  cada 200 ms · timeout 500 ms · umbral 3 · solo cambios: false
  0.0 s  discord ← 🔴 pagos falló: HTTP 500
  0.4 s  discord ← 🔴 pagos falló: HTTP 500
  0.5 s  discord ← 🔴 lento falló: no respondió a tiempo
  0.8 s  discord ← 🔴 pagos falló: HTTP 500
  1.0 s  ── api se apaga
  1.0 s  discord ← 🔴 lento falló: no respondió a tiempo
  1.0 s  discord ← 🔴 api falló: conexión rechazada
  1.2 s  discord ← 🔴 api falló: conexión rechazada
  1.2 s  discord ← 🔴 pagos falló: HTTP 500
  1.4 s  discord ← 🔴 api falló: conexión rechazada
  1.5 s  discord ← 🔴 lento falló: no respondió a tiempo
  1.6 s  discord ← 🔴 pagos falló: HTTP 500
  1.6 s  discord ← 🔴 api falló: conexión rechazada
  1.8 s  discord ← 🔴 api falló: conexión rechazada
  2.0 s  discord ← 🔴 api falló: conexión rechazada
  2.0 s  discord ← 🔴 pagos falló: HTTP 500
  2.0 s  discord ← 🔴 lento falló: no respondió a tiempo
  2.2 s  discord ← 🔴 api falló: conexión rechazada
  2.4 s  discord ← 🔴 api falló: conexión rechazada
  2.4 s  discord ← 🔴 pagos falló: HTTP 500
  2.5 s  discord ← 🔴 lento falló: no respondió a tiempo
  2.6 s  discord ← 🔴 api falló: conexión rechazada
  2.8 s  discord ← 🔴 pagos falló: HTTP 500
  2.8 s  discord ← 🔴 api falló: conexión rechazada
  3.0 s  ── api vuelve
  3.0 s  discord ← 🔴 lento falló: no respondió a tiempo
  3.2 s  discord ← 🔴 pagos falló: HTTP 500
  3.5 s  discord ← 🔴 lento falló: no respondió a tiempo
  3.6 s  discord ← 🔴 pagos falló: HTTP 500
  4.0 s  discord ← 🔴 pagos falló: HTTP 500
  4.0 s  discord ← 🔴 lento falló: no respondió a tiempo
  4.4 s  discord ← 🔴 pagos falló: HTTP 500
  4.5 s  discord ← 🔴 lento falló: no respondió a tiempo
  4.8 s  discord ← 🔴 pagos falló: HTTP 500
  avisos en 5 s: 32  (web 0 · api 10 · pagos 13 · lento 9)

$ ./monitor estados
  sitios: web (bien) · api (se cae del 1 al 3) · pagos (500 una sí, una no) · lento (tarda 3 s)
  cada 200 ms · timeout 500 ms · umbral 3 · solo cambios: true
  1.0 s  ── api se apaga
  1.4 s  discord ← 🔴 api CAÍDO tras 3 fallos seguidos: conexión rechazada
  1.5 s  discord ← 🔴 lento CAÍDO tras 3 fallos seguidos: no respondió a tiempo
  3.0 s  ── api vuelve
  3.0 s  discord ← 🟢 api volvió
  avisos en 5 s: 3  (web 0 · api 2 · pagos 0 · lento 1)

$ ./monitor sin-timeout
  http.Client{} sin Timeout · lento tarda 3 s en contestar 200
  3.0 s  lento: 200 OK en 3.0 s
  6.0 s  lento: 200 OK en 3.0 s
  avisos en 7 s: 0  (web 0 · api 0 · pagos 0 · lento 0)

$ ./monitor rondas
  por rondas, sin timeout    → web revisado  2 veces en 5 s
  por rondas, timeout 500 ms → web revisado  8 veces en 5 s
  por sitio,  timeout 500 ms → web revisado 25 veces en 5 s

$ DISCORD_WEBHOOK_URL= ./monitor entorno; echo "salida: $?"
falta DISCORD_WEBHOOK_URL
salida: 1
```
