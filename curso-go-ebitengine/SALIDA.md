# SALIDA — curso «Ebitengine: haz un juego en Go y entiende el bucle de fotogramas»

> Generado por `ejecutar_todo.py`. **No se escribe a mano.** Capturas en `capturas/`.

- fecha: `2026-10-05 18:10:55Z` · todo compilado y abierto en Docker (Debian 12) sobre Xvfb, con Mesa por software
- `go version go1.26.8 linux/amd64`
- `github.com/hajimehoshi/ebiten/v2 v2.10.4`
- ⚠️ en una pantalla virtual no hay refresco de monitor: el vsync no limita los fotogramas

## Leccion 1 — El bucle: Update y Draw no son lo mismo

`bash correr.sh 01_bucle; VSYNC=no bash correr.sh 01_bucle 01_bucle-sin-vsync` · codigo de salida `0`

```
  captura: capturas/01_bucle.png
  vsync sí                             5.0 s · Update:   294 ( 58.6/s) · Draw:  2291 (456.9/s)
  captura: capturas/01_bucle-sin-vsync.png
  vsync no                             5.0 s · Update:   296 ( 59.2/s) · Draw:  2612 (522.1/s)
```

## Leccion 2 — La trampa: mover el personaje en Draw

`HZ=60 bash correr.sh 02_mover_en_draw 02_mover-60hz; HZ=144 bash correr.sh 02_mover_en_draw 02_mover-144hz` · codigo de salida `0`

```
  captura: capturas/02_mover-60hz.png
  monitor simulado de 60 Hz            3.0 s · Update:   172 ( 57.1/s) · Draw:   126 ( 41.8/s)
    movido en Draw:    126 px  · movido en Update:  172 px
  captura: capturas/02_mover-144hz.png
  monitor simulado de 144 Hz           3.0 s · Update:   175 ( 58.3/s) · Draw:   276 ( 91.9/s)
    movido en Draw:    276 px  · movido en Update:  175 px
```

## Leccion 3 — Tiempo fijo: la velocidad por segundo, no por tick

`TPS=60 bash correr.sh 03_tps; TPS=30 bash correr.sh 03_tps` · codigo de salida `0`

```
  SetTPS(60)                           3.0 s · Update:   175 ( 58.1/s) · Draw:  1353 (448.8/s)
    «1 px por tick»:  175 px  ·  «60 px por segundo»:  175 px  (3 s: deberían ser 180)
  SetTPS(30)                           3.0 s · Update:    92 ( 30.5/s) · Draw:  1357 (450.3/s)
    «1 px por tick»:   92 px  ·  «60 px por segundo»:  184 px  (3 s: deberían ser 180)
```

## Leccion 4 — Lo que tarda el dibujado no frena la lógica; lo que tarda la lógica, sí

`DRAW=0 UPDATE=0 bash correr.sh 04_lento; DRAW=50 UPDATE=0 bash correr.sh 04_lento; DRAW=0 UPDATE=30 bash correr.sh 04_lento` · codigo de salida `0`

```
  Draw +0s · Update +0s                3.0 s · Update:   175 ( 58.0/s) · Draw:  1355 (449.2/s)
  Draw +50ms · Update +0s              3.0 s · Update:   172 ( 56.6/s) · Draw:    50 ( 16.4/s)
  Draw +0s · Update +30ms              3.0 s · Update:    88 ( 29.3/s) · Draw:    17 (  5.7/s)
```

## Leccion 5 — La entrada: pulsada no es «se acaba de pulsar»

`bash correr.sh 05_entrada` · codigo de salida `0`

```
  dos pulsaciones de espacio           5.0 s · Update:   295 ( 58.8/s) · Draw:  2347 (468.0/s)
    disparos con IsKeyPressed: 62  ·  con inpututil.IsKeyJustPressed: 2
```

## Leccion 6 — Probar la lógica sin ventana

`go test -v -count=1 ./mundo` · codigo de salida `0`

```
=== RUN   TestPartidaDeterminista
    mundo_test.go:22: semilla 42: sobrevive 2146 ticks (35.8 s de juego) · calculado en 50µs
    mundo_test.go:23: semilla 42 otra vez: 2146 ticks · semilla 7: 5490 ticks
--- PASS: TestPartidaDeterminista (0.00s)
=== RUN   TestChocar
--- PASS: TestChocar (0.00s)
PASS
ok  	curso-ebiten/mundo	0.004s
```

## Leccion 7 — El proyecto: esquiva

`DEMO=1 bash correr.sh 07_esquiva` · codigo de salida `0`

```
  captura: capturas/07_esquiva.png
  esquiva, piloto automático, semilla 42   8.0 s · Update:   474 ( 59.2/s) · Draw:  3774 (471.3/s)
    ticks de juego: 474 · vivo: true · bloques en pantalla: 6
```

