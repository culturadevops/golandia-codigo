# SALIDA — repeticiones del monitor (golandia)

> Corrida del 2026-10-21 09:48. golang:1.27.2, --cpus 2, --network none. 5 repeticiones de cada modo.

```
ingenuo #1:   avisos en 5 s: 33  (web 0 · api 11 · pagos 13 · lento 9)
ingenuo #2:   avisos en 5 s: 33  (web 0 · api 11 · pagos 13 · lento 9)
ingenuo #3:   avisos en 5 s: 32  (web 0 · api 10 · pagos 13 · lento 9)
ingenuo #4:   avisos en 5 s: 33  (web 0 · api 11 · pagos 13 · lento 9)
ingenuo #5:   avisos en 5 s: 33  (web 0 · api 11 · pagos 13 · lento 9)
estados #1:  1.5 s discord ← 🔴 lento CAÍDO tras 3 fallos seguidos: no respondió a tiempo| 1.6 s discord ← 🔴 api CAÍDO tras 3 fallos seguidos: conexión rechazada| 3.2 s discord ← 🟢 api volvió| avisos en 5 s: 3 (web 0 · api 2 · pagos 0 · lento 1)|
estados #2:  1.4 s discord ← 🔴 api CAÍDO tras 3 fallos seguidos: conexión rechazada| 1.5 s discord ← 🔴 lento CAÍDO tras 3 fallos seguidos: no respondió a tiempo| 3.2 s discord ← 🟢 api volvió| avisos en 5 s: 3 (web 0 · api 2 · pagos 0 · lento 1)|
estados #3:  1.4 s discord ← 🔴 api CAÍDO tras 3 fallos seguidos: conexión rechazada| 1.5 s discord ← 🔴 lento CAÍDO tras 3 fallos seguidos: no respondió a tiempo| 3.0 s discord ← 🟢 api volvió| avisos en 5 s: 3 (web 0 · api 2 · pagos 0 · lento 1)|
estados #4:  1.4 s discord ← 🔴 api CAÍDO tras 3 fallos seguidos: conexión rechazada| 1.5 s discord ← 🔴 lento CAÍDO tras 3 fallos seguidos: no respondió a tiempo| 3.2 s discord ← 🟢 api volvió| avisos en 5 s: 3 (web 0 · api 2 · pagos 0 · lento 1)|
estados #5:  1.4 s discord ← 🔴 api CAÍDO tras 3 fallos seguidos: conexión rechazada| 1.5 s discord ← 🔴 lento CAÍDO tras 3 fallos seguidos: no respondió a tiempo| 3.2 s discord ← 🟢 api volvió| avisos en 5 s: 3 (web 0 · api 2 · pagos 0 · lento 1)|
```
