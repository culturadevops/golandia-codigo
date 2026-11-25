# SALIDA — Autoheal en Go (golandia, 2026-10-21)

Ejecutado el 2026-10-21 en Docker 29.6.2 (API 1.55 en el daemon; el vigía fija /v1.44), golang:1.27.2, contenedores nginx:1.30.5-alpine con --init.
Healthcheck de la app: `wget -q -T 1 -O /dev/null http://127.0.0.1/ || exit 1`, interval 2s, timeout 2s, retries 3.
«Colgar» = `kill -STOP $(pidof nginx)` dentro del contenedor: el proceso sigue vivo pero no responde.
Los tiempos de la izquierda son del observador (cada 0,5 s, desde el cuelgue). El log del vigía tiene su propio reloj (arranca ~2 s antes).
Generado con `python correr.py` (salidas/*.txt, sin editar).

## 1 · Sin vigía: --restart unless-stopped no hace nada

```
$ docker run -d --name gh-app --restart unless-stopped --init --health-cmd wget -q -T 1 -O /dev/null http://127.0.0.1/ || exit 1 … nginx:1.30.5-alpine
$ (cuelgo nginx: kill -STOP)
   0.2s  gh-app    running  healthy    RestartCount=0  arrancó 14:43:49
   9.4s  gh-app    running  unhealthy  RestartCount=0  arrancó 14:43:49
  60.6s  fin de la observación · running unhealthy 0
```

## 2 · Vigía en modo eventos (GET /events)

```
$ vigia -modo eventos -etiqueta golandia.autoheal=demo
$ (cuelgo nginx: kill -STOP)
   0.2s  gh-app    running  healthy    RestartCount=0  arrancó 14:45:18
   8.8s  gh-app    running  unhealthy  RestartCount=0  arrancó 14:45:18
  11.1s  gh-app    running  starting   RestartCount=0  arrancó 14:45:34
  13.5s  gh-app    running  healthy    RestartCount=0  arrancó 14:45:34
  13.5s  sano otra vez
--- log del vigía (su reloj empieza cuando arranca él, ~2 s antes del cuelgue) ---
    0.0s  vigía en modo eventos · etiqueta golandia.autoheal=demo
   10.8s  evento: gh-app · health_status: unhealthy
   10.8s  reinicio gh-app
   13.1s  gh-app reiniciado (la llamada tardó 2.3s)
   15.2s  evento: gh-app · health_status: healthy
```

## 3 · Vigía en modo sondeo cada 10 s (GET /containers/json)

```
$ vigia -modo sondeo -cada 10s -etiqueta golandia.autoheal=demo
$ (cuelgo nginx: kill -STOP)
   0.2s  gh-app    running  healthy    RestartCount=0  arrancó 14:45:37
   8.8s  gh-app    running  unhealthy  RestartCount=0  arrancó 14:45:37
  20.6s  gh-app    running  starting   RestartCount=0  arrancó 14:46:02
  22.9s  gh-app    running  healthy    RestartCount=0  arrancó 14:46:02
  23.0s  sano otra vez
--- log del vigía (su reloj empieza cuando arranca él, ~2 s antes del cuelgue) ---
    0.0s  vigía en modo sondeo · etiqueta golandia.autoheal=demo
   20.0s  reinicio gh-app
   22.4s  gh-app reiniciado (la llamada tardó 2.4s)
```

## 4 · Contenedor roto de verdad: el freno (máx. 3 en 2 min)

```
$ gh-roto: nginx escucha en el 8080 y el healthcheck pide el 80 → enfermo siempre
$ vigia -modo eventos -max 3 -ventana 2m
   0.0s  gh-roto   running  starting   RestartCount=0  arrancó 14:46:06
   6.0s  gh-roto   running  unhealthy  RestartCount=0  arrancó 14:46:06
   6.5s  gh-roto   running  starting   RestartCount=0  arrancó 14:46:12
  12.5s  gh-roto   running  unhealthy  RestartCount=0  arrancó 14:46:12
  13.0s  gh-roto   running  starting   RestartCount=0  arrancó 14:46:19
  19.1s  gh-roto   exited   unhealthy  RestartCount=0  arrancó 14:46:19
  19.6s  gh-roto   running  starting   RestartCount=0  arrancó 14:46:25
  25.5s  gh-roto   running  unhealthy  RestartCount=0  arrancó 14:46:25
  70.3s  fin · running unhealthy 0
--- log del vigía ---
    0.0s  vigía en modo eventos · etiqueta golandia.autoheal=demo
    5.9s  evento: gh-roto · health_status: unhealthy
    5.9s  reinicio gh-roto
    6.2s  gh-roto reiniciado (la llamada tardó 0.3s)
   12.4s  evento: gh-roto · health_status: unhealthy
   12.4s  reinicio gh-roto
   12.7s  gh-roto reiniciado (la llamada tardó 0.4s)
   18.8s  evento: gh-roto · health_status: unhealthy
   18.8s  reinicio gh-roto
   19.2s  gh-roto reiniciado (la llamada tardó 0.3s)
   25.3s  evento: gh-roto · health_status: unhealthy
   25.3s  ⛔ gh-roto: 3 reinicios en 2m0s y sigue enfermo · lo dejo y aviso
```

## 5 · Solo toca lo que lleva la etiqueta

```
$ gh-app (con la etiqueta) y gh-ajeno (sin ella): cuelgo los dos
   0.2s  gh-app    running  healthy    RestartCount=0  arrancó 14:47:17
   0.3s  gh-ajeno  running  healthy    RestartCount=0  arrancó 14:47:18
   9.0s  gh-app    running  unhealthy  RestartCount=0  arrancó 14:47:17
   9.1s  gh-ajeno  running  unhealthy  RestartCount=0  arrancó 14:47:18
  11.4s  gh-app    exited   unhealthy  RestartCount=0  arrancó 14:47:17
  12.0s  gh-app    running  starting   RestartCount=0  arrancó 14:47:34
  14.4s  gh-app    running  healthy    RestartCount=0  arrancó 14:47:34
  40.3s  fin · gh-app running healthy 0 · gh-ajeno running unhealthy 0
--- log del vigía ---
    0.0s  vigía en modo eventos · etiqueta golandia.autoheal=demo
   10.5s  evento: gh-app · health_status: unhealthy
   10.5s  reinicio gh-app
   13.7s  gh-app reiniciado (la llamada tardó 3.2s)
   15.8s  evento: gh-app · health_status: healthy
```

## 6 · Binario: net/http contra el SDK oficial (github.com/moby/moby/client v0.6.2)

```
vigia      6725792 bytes con -s -w · módulos 1  · paquetes 201   (todo el vigía: eventos, sondeo, freno, apagado limpio)
vigia-sdk  7852192 bytes con -s -w · módulos 49 · paquetes 266   (solo listar enfermos y reiniciarlos)
sin -s -w: vigia 9903997 bytes · vigia-sdk 11482549 bytes
el SDK trae, entre otros: go.opentelemetry.io/otel, otelhttp, containerd/errdefs, opencontainers/image-spec, distribution/reference
```

Prueba del vigia-sdk contra un gh-app colgado:

```
unhealthy 2026-10-21T14:49:20.428890669Z
$ vigia-sdk golandia.autoheal=demo
reinicio /gh-app
healthy 2026-10-21T14:49:46.512207184Z
```

## Lo que sale

- Sin vigía, 60 s después del cuelgue el contenedor sigue `running unhealthy` con RestartCount=0: la política de reinicio solo actúa si el proceso termina.
- Con eventos: unhealthy a los 8,8 s, el vigía reinicia en el mismo instante y está healthy a los 13,5 s.
- Con sondeo cada 10 s: unhealthy a los 8,8 s, healthy a los 22,9 s (el reinicio esperó al siguiente sondeo).
- El reinicio de la app colgada tarda 2,3-3,2 s: SIGTERM no la despierta (está en SIGSTOP), Docker espera la gracia (t=2) y manda SIGKILL. La app sana (gh-roto) tarda 0,3-0,4 s.
- RestartCount sigue en 0 después de los reinicios por API: solo cuenta los de la política de Docker. StartedAt sí cambia.
- El freno: gh-roto se reinicia 3 veces (5,9 / 12,4 / 18,8 s) y al 4.º unhealthy (25,3 s) el vigía lo deja y avisa.
- gh-ajeno, sin la etiqueta, se queda unhealthy: el vigía no lo toca.
