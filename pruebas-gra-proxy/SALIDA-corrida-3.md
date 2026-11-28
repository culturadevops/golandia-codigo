# Proxy inverso: Caddy y Traefik (Go) contra Pingap y Sōzu (Rust) · corrida 3

> Corrida del 2026-11-27 09:08, con el candado `_logs/banco.lock` tomado y sin `fab producir` vivo. Regenerable con `python investigacion/pruebas_gra_proxy.py 3`. Orden: caddy → traefik → pingap → sozu. Cada carga: wrk -t6 -d20s.

## 0 · Versiones, imágenes y configuración

- **caddy** (Go) `caddy:2.11.9` · imagen 23.7 MiB · config 7 líneas (Caddyfile) · `v2.11.9 h1:bhZtcRKQOq0fuLvEgCdPYvUxXyGn9kzjG4lh2332DTE=`
- **traefik** (Go) `traefik:v3.8.0` · imagen 61.3 MiB · config 19 líneas (traefik.yml, traefik-dinamico.yml) · `Version: 3.8.0 Codename: salers Go version: go1.27.4 Built: 2026-11-23T14:33:45Z OS/Arch: linux/amd64`
- **pingap** (Rust) `vicanso/pingap:0.16.2` · imagen 41.6 MiB · config 13 líneas (pingap.toml) · `pingap 0.16.2 (5c8c3130127020e48f37896c7d5c46e8ec3442fb, tls=openssl)`
- **sozu** (Rust) `clevercloud/sozu:2.3.7` · imagen 49.1 MiB · config 21 líneas (sozu.toml) · `sozu 2.3.7 () +jemallocator +crypto-ring -crypto-aws-lc-rs -crypto-openssl -fips -simd -splice -opentelemetry -tolerant-http1-parser -logs-debug -logs-trace -un`
- backend nginx:1.30.6-alpine (4 CPU) · cliente wrk (Alpine 3.24, 6 hilos, 6 CPU) · ajenos vivos: [] · CPU de Windows: 29 %

## 1 · ¿La misma respuesta a través de cada uno?

- directo al backend: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254
- caddy: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254 · ✅ igual (respondió a los 0.4 s de arrancar)
- traefik: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254 · ✅ igual (respondió a los 0.4 s de arrancar)
- pingap: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254 · ✅ igual (respondió a los 0.4 s de arrancar)
- sozu: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254 · ✅ igual (respondió a los 0.4 s de arrancar)

## 2 · Línea base: wrk directo al backend (sin proxy)

- /chico c=64: 883867 pet/s · p50 0.06 ms · p99 0.46 ms · códigos {'2xx': 17701454, 'no-2xx': 0} · errores de socket: connect 0, read 0, write 0, timeout 51
- /chico c=256: 935094 pet/s · p50 0.23 ms · p99 0.75 ms · códigos {'2xx': 18783225, 'no-2xx': 0} · errores de socket: connect 0, read 0, write 0, timeout 236
- /grande c=64: 82128 pet/s · p50 0.66 ms · p99 1.59 ms · códigos {'2xx': 1645477, 'no-2xx': 0} · errores de socket: connect 0, read 0, write 0, timeout 59

## 3 · A tope, cada proxy (2 CPU, 1 GiB)

### caddy · memoria en reposo 10.8 MiB

- /chico c=64: **43301 pet/s** · p50 1.19 ms · p99 4.63 ms · CPU media 194.6 % (máx 200.3) · memoria máx 26.8 MiB · hilos máx 7 · códigos {'2xx': 870337, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 870337 respuestas (870337 2xx); nginx atendió 870392 por 1865 conexiones nuevas del proxy (✅)
- /chico c=256: **5656 pet/s** · p50 28.16 ms · p99 1320.00 ms · CPU media 196.9 % (máx 238.5) · memoria máx 65.6 MiB · hilos máx 33 · códigos {'2xx': 112938, 'no-2xx': 691} · ⚠️ {'no-2xx': 691}
  - ¿llegó todo al backend? wrk contó 113629 respuestas (112938 2xx); nginx atendió 113089 por 28267 conexiones nuevas del proxy (✅)
- /grande c=64: **21075 pet/s** · p50 2.59 ms · p99 8.14 ms · CPU media 195.5 % (máx 201.3) · memoria máx 29.1 MiB · hilos máx 8 · códigos {'2xx': 422516, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 422516 respuestas (422516 2xx); nginx atendió 422570 por 371 conexiones nuevas del proxy (✅)

### traefik · memoria en reposo 17.9 MiB

- /chico c=64: **59242 pet/s** · p50 0.89 ms · p99 4.80 ms · CPU media 196.4 % (máx 202.1) · memoria máx 46.7 MiB · hilos máx 10 · códigos {'2xx': 1190780, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 1190780 respuestas (1190780 2xx); nginx atendió 1190833 por 60 conexiones nuevas del proxy (✅)
- /chico c=256: **51149 pet/s** · p50 4.49 ms · p99 13.03 ms · CPU media 196.0 % (máx 202.3) · memoria máx 95.8 MiB · hilos máx 8 · códigos {'2xx': 1025710, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 1025710 respuestas (1025710 2xx); nginx atendió 1025956 por 259 conexiones nuevas del proxy (✅)
- /grande c=64: **23856 pet/s** · p50 2.32 ms · p99 7.41 ms · CPU media 196.9 % (máx 202.9) · memoria máx 56.8 MiB · hilos máx 7 · códigos {'2xx': 479497, 'no-2xx': 0} · errores de socket: connect 0, read 0, write 0, timeout 2
  - ¿llegó todo al backend? wrk contó 479497 respuestas (479497 2xx); nginx atendió 479552 por 69 conexiones nuevas del proxy (✅)

### pingap · memoria en reposo 8.3 MiB

- /chico c=64: **72969 pet/s** · p50 0.79 ms · p99 1.72 ms · CPU media 187.5 % (máx 192.0) · memoria máx 15.3 MiB · hilos máx 10 · códigos {'2xx': 1460304, 'no-2xx': 0} · errores de socket: connect 0, read 0, write 0, timeout 28
  - ¿llegó todo al backend? wrk contó 1460304 respuestas (1460304 2xx); nginx atendió 1460357 por 59 conexiones nuevas del proxy (✅)
- /chico c=256: **58850 pet/s** · p50 4.10 ms · p99 8.48 ms · CPU media 184.7 % (máx 190.9) · memoria máx 26.1 MiB · hilos máx 10 · códigos {'2xx': 1180846, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 1180846 respuestas (1180846 2xx); nginx atendió 1181090 por 247 conexiones nuevas del proxy (✅)
- /grande c=64: **17989 pet/s** · p50 3.25 ms · p99 5.84 ms · CPU media 193.3 % (máx 198.1) · memoria máx 20.3 MiB · hilos máx 10 · códigos {'2xx': 360353, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 360353 respuestas (360353 2xx); nginx atendió 360408 por 58 conexiones nuevas del proxy (✅)

### sozu · memoria en reposo 6.3 MiB

- /chico c=64: **116863 pet/s** · p50 0.39 ms · p99 1.75 ms · CPU media 197.7 % (máx 203.2) · memoria máx 9.7 MiB · hilos máx 3 · códigos {'2xx': 2341139, 'no-2xx': 0} · errores de socket: connect 0, read 0, write 0, timeout 58
  - ¿llegó todo al backend? wrk contó 2341139 respuestas (2341139 2xx); nginx atendió 2341189 por 60 conexiones nuevas del proxy (✅)
- /chico c=256: **112893 pet/s** · p50 1.83 ms · p99 5.07 ms · CPU media 197.5 % (máx 203.2) · memoria máx 15.6 MiB · hilos máx 3 · códigos {'2xx': 2265711, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 2265711 respuestas (2265711 2xx); nginx atendió 2265958 por 252 conexiones nuevas del proxy (✅)
- /grande c=64: **33441 pet/s** · p50 1.63 ms · p99 3.92 ms · CPU media 197.8 % (máx 203.1) · memoria máx 14.8 MiB · hilos máx 3 · códigos {'2xx': 672153, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 672153 respuestas (672153 2xx); nginx atendió 672206 por 60 conexiones nuevas del proxy (✅)

## 3b · Caddy con keepalive_idle_conns_per_host 1024 (por defecto 32)

- caddy-ka /chico c=64: **39581 pet/s** · p50 1.27 ms · p99 5.61 ms · CPU media 194.4 % · memoria máx 25.0 MiB · códigos {'2xx': 795549, 'no-2xx': 0} · nginx atendió 795604 por 61 conexiones nuevas
- caddy-ka /chico c=256: **28050 pet/s** · p50 8.57 ms · p99 19.33 ms · CPU media 194.8 % · memoria máx 56.9 MiB · códigos {'2xx': 563326, 'no-2xx': 0} · nginx atendió 563575 por 277 conexiones nuevas
- caddy-ka /grande c=64: **17288 pet/s** · p50 3.17 ms · p99 10.08 ms · CPU media 195.4 % · memoria máx 29.9 MiB · códigos {'2xx': 346395, 'no-2xx': 0} · nginx atendió 346450 por 72 conexiones nuevas

## 4 · La tabla

| proxy | lenguaje | carga | pet/s | p99 ms | CPU media % | memoria máx MiB | reposo MiB |
|---|---|---|---:|---:|---:|---:|---:|
| caddy | Go | /chico c=64 | 43301 | 4.63 | 194.6 | 26.8 | 10.8 |
| caddy | Go | /chico c=256 | 5656 | 1320.00 | 196.9 | 65.6 | 10.8 |
| caddy | Go | /grande c=64 | 21075 | 8.14 | 195.5 | 29.1 | 10.8 |
| traefik | Go | /chico c=64 | 59242 | 4.80 | 196.4 | 46.7 | 17.9 |
| traefik | Go | /chico c=256 | 51149 | 13.03 | 196.0 | 95.8 | 17.9 |
| traefik | Go | /grande c=64 | 23856 | 7.41 | 196.9 | 56.8 | 17.9 |
| pingap | Rust | /chico c=64 | 72969 | 1.72 | 187.5 | 15.3 | 8.3 |
| pingap | Rust | /chico c=256 | 58850 | 8.48 | 184.7 | 26.1 | 8.3 |
| pingap | Rust | /grande c=64 | 17989 | 5.84 | 193.3 | 20.3 | 8.3 |
| sozu | Rust | /chico c=64 | 116863 | 1.75 | 197.7 | 9.7 | 6.3 |
| sozu | Rust | /chico c=256 | 112893 | 5.07 | 197.5 | 15.6 | 6.3 |
| sozu | Rust | /grande c=64 | 33441 | 3.92 | 197.8 | 14.8 | 6.3 |

> Terminó 2026-11-27 09:16. CPU de Windows al final: 51 %.
