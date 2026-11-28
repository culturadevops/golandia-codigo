# Proxy inverso: Caddy y Traefik (Go) contra Pingap y Sōzu (Rust) · corrida 2

> Corrida del 2026-11-27 08:19, con el candado `_logs/banco.lock` tomado y sin `fab producir` vivo. Regenerable con `python investigacion/pruebas_gra_proxy.py 2`. Orden: sozu → pingap → traefik → caddy. Cada carga: wrk -t6 -d20s.

## 0 · Versiones, imágenes y configuración

- **caddy** (Go) `caddy:2.11.9` · imagen 23.7 MiB · config 7 líneas (Caddyfile) · `v2.11.9 h1:bhZtcRKQOq0fuLvEgCdPYvUxXyGn9kzjG4lh2332DTE=`
- **traefik** (Go) `traefik:v3.8.0` · imagen 61.3 MiB · config 19 líneas (traefik.yml, traefik-dinamico.yml) · `Version: 3.8.0 Codename: salers Go version: go1.27.4 Built: 2026-11-23T14:33:45Z OS/Arch: linux/amd64`
- **pingap** (Rust) `vicanso/pingap:0.16.2` · imagen 41.6 MiB · config 13 líneas (pingap.toml) · `pingap 0.16.2 (5c8c3130127020e48f37896c7d5c46e8ec3442fb, tls=openssl)`
- **sozu** (Rust) `clevercloud/sozu:2.3.7` · imagen 49.1 MiB · config 21 líneas (sozu.toml) · `sozu 2.3.7 () +jemallocator +crypto-ring -crypto-aws-lc-rs -crypto-openssl -fips -simd -splice -opentelemetry -tolerant-http1-parser -logs-debug -logs-trace -un`
- backend nginx:1.30.6-alpine (4 CPU) · cliente wrk (Alpine 3.24, 6 hilos, 6 CPU) · ajenos vivos: [] · CPU de Windows: 7 %

## 1 · ¿La misma respuesta a través de cada uno?

- directo al backend: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254
- sozu: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254 · ✅ igual (respondió a los 0.4 s de arrancar)
- pingap: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254 · ✅ igual (respondió a los 0.4 s de arrancar)
- traefik: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254 · ✅ igual (respondió a los 0.4 s de arrancar)
- caddy: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254 · ✅ igual (respondió a los 0.3 s de arrancar)

## 2 · Línea base: wrk directo al backend (sin proxy)

- /chico c=64: 887140 pet/s · p50 0.05 ms · p99 0.42 ms · códigos {'2xx': 17831335, 'no-2xx': 0}
- /chico c=256: 995333 pet/s · p50 0.20 ms · p99 0.96 ms · códigos {'2xx': 19990535, 'no-2xx': 0} · errores de socket: connect 0, read 0, write 0, timeout 231
- /grande c=64: 85655 pet/s · p50 0.61 ms · p99 5.36 ms · códigos {'2xx': 1716793, 'no-2xx': 0} · errores de socket: connect 0, read 0, write 0, timeout 60

## 3 · A tope, cada proxy (2 CPU, 1 GiB)

### sozu · memoria en reposo 5.7 MiB

- /chico c=64: **150452 pet/s** · p50 0.35 ms · p99 2.05 ms · CPU media 197.8 % (máx 203.2) · memoria máx 9.6 MiB · hilos máx 3 · códigos {'2xx': 3023913, 'no-2xx': 0} · errores de socket: connect 0, read 0, write 0, timeout 54
  - ¿llegó todo al backend? wrk contó 3023913 respuestas (3023913 2xx); nginx atendió 3023964 por 60 conexiones nuevas del proxy (✅)
- /chico c=256: **145120 pet/s** · p50 1.59 ms · p99 6.27 ms · CPU media 197.4 % (máx 203.2) · memoria máx 15.1 MiB · hilos máx 3 · códigos {'2xx': 2914942, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 2914942 respuestas (2914942 2xx); nginx atendió 2915184 por 252 conexiones nuevas del proxy (✅)
- /grande c=64: **39751 pet/s** · p50 1.31 ms · p99 7.13 ms · CPU media 197.7 % (máx 203.1) · memoria máx 14.2 MiB · hilos máx 3 · códigos {'2xx': 798981, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 798981 respuestas (798981 2xx); nginx atendió 799038 por 60 conexiones nuevas del proxy (✅)

### pingap · memoria en reposo 8.0 MiB

- /chico c=64: **76152 pet/s** · p50 0.74 ms · p99 1.87 ms · CPU media 186.9 % (máx 191.3) · memoria máx 16.9 MiB · hilos máx 10 · códigos {'2xx': 1530633, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 1530633 respuestas (1530633 2xx); nginx atendió 1530687 por 60 conexiones nuevas del proxy (✅)
- /chico c=256: **70422 pet/s** · p50 3.47 ms · p99 9.52 ms · CPU media 183.6 % (máx 190.6) · memoria máx 27.8 MiB · hilos máx 10 · códigos {'2xx': 1414729, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 1414729 respuestas (1414729 2xx); nginx atendió 1414972 por 245 conexiones nuevas del proxy (✅)
- /grande c=64: **34563 pet/s** · p50 1.50 ms · p99 6.72 ms · CPU media 193.3 % (máx 199.6) · memoria máx 19.7 MiB · hilos máx 10 · códigos {'2xx': 694720, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 694720 respuestas (694720 2xx); nginx atendió 694771 por 60 conexiones nuevas del proxy (✅)

### traefik · memoria en reposo 18.7 MiB

- /chico c=64: **66698 pet/s** · p50 0.80 ms · p99 5.44 ms · CPU media 195.5 % (máx 202.0) · memoria máx 52.1 MiB · hilos máx 26 · códigos {'2xx': 1340619, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 1340619 respuestas (1340619 2xx); nginx atendió 1340673 por 72 conexiones nuevas del proxy (✅)
- /chico c=256: **59613 pet/s** · p50 3.76 ms · p99 15.45 ms · CPU media 196.5 % (máx 203.0) · memoria máx 96.1 MiB · hilos máx 8 · códigos {'2xx': 1197201, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 1197201 respuestas (1197201 2xx); nginx atendió 1197446 por 236 conexiones nuevas del proxy (✅)
- /grande c=64: **28133 pet/s** · p50 1.90 ms · p99 8.08 ms · CPU media 196.9 % (máx 202.6) · memoria máx 54.0 MiB · hilos máx 8 · códigos {'2xx': 563710, 'no-2xx': 0} · errores de socket: connect 0, read 0, write 0, timeout 6
  - ¿llegó todo al backend? wrk contó 563710 respuestas (563710 2xx); nginx atendió 563760 por 58 conexiones nuevas del proxy (✅)

### caddy · memoria en reposo 11.4 MiB

- /chico c=64: **49016 pet/s** · p50 1.06 ms · p99 4.89 ms · CPU media 194.5 % (máx 199.8) · memoria máx 25.6 MiB · hilos máx 8 · códigos {'2xx': 981274, 'no-2xx': 0} · errores de socket: connect 0, read 0, write 0, timeout 36
  - ¿llegó todo al backend? wrk contó 981274 respuestas (981274 2xx); nginx atendió 981328 por 1990 conexiones nuevas del proxy (✅)
- /chico c=256: **5267 pet/s** · p50 32.22 ms · p99 1670.00 ms · CPU media 196.6 % (máx 217.9) · memoria máx 62.4 MiB · hilos máx 32 · códigos {'2xx': 104593, 'no-2xx': 1246} · ⚠️ {'no-2xx': 1246}
  - ¿llegó todo al backend? wrk contó 105839 respuestas (104593 2xx); nginx atendió 104720 por 28263 conexiones nuevas del proxy (✅)
- /grande c=64: **24417 pet/s** · p50 2.19 ms · p99 8.51 ms · CPU media 195.1 % (máx 201.2) · memoria máx 29.6 MiB · hilos máx 8 · códigos {'2xx': 489282, 'no-2xx': 0} · errores de socket: connect 0, read 0, write 0, timeout 2
  - ¿llegó todo al backend? wrk contó 489282 respuestas (489282 2xx); nginx atendió 489338 por 376 conexiones nuevas del proxy (✅)

## 3b · Caddy con keepalive_idle_conns_per_host 1024 (por defecto 32)

- caddy-ka /chico c=64: **50561 pet/s** · p50 1.03 ms · p99 4.45 ms · CPU media 194.5 % · memoria máx 26.1 MiB · códigos {'2xx': 1012412, 'no-2xx': 0} · nginx atendió 1012464 por 71 conexiones nuevas
- caddy-ka /chico c=256: **43866 pet/s** · p50 5.18 ms · p99 16.91 ms · CPU media 194.7 % · memoria máx 56.3 MiB · códigos {'2xx': 880574, 'no-2xx': 0} · nginx atendió 880819 por 259 conexiones nuevas
- caddy-ka /grande c=64: **25191 pet/s** · p50 2.09 ms · p99 8.84 ms · CPU media 195.9 % · memoria máx 28.5 MiB · códigos {'2xx': 506337, 'no-2xx': 0} · nginx atendió 506393 por 57 conexiones nuevas

## 4 · La tabla

| proxy | lenguaje | carga | pet/s | p99 ms | CPU media % | memoria máx MiB | reposo MiB |
|---|---|---|---:|---:|---:|---:|---:|
| caddy | Go | /chico c=64 | 49016 | 4.89 | 194.5 | 25.6 | 11.4 |
| caddy | Go | /chico c=256 | 5267 | 1670.00 | 196.6 | 62.4 | 11.4 |
| caddy | Go | /grande c=64 | 24417 | 8.51 | 195.1 | 29.6 | 11.4 |
| traefik | Go | /chico c=64 | 66698 | 5.44 | 195.5 | 52.1 | 18.7 |
| traefik | Go | /chico c=256 | 59613 | 15.45 | 196.5 | 96.1 | 18.7 |
| traefik | Go | /grande c=64 | 28133 | 8.08 | 196.9 | 54.0 | 18.7 |
| pingap | Rust | /chico c=64 | 76152 | 1.87 | 186.9 | 16.9 | 8.0 |
| pingap | Rust | /chico c=256 | 70422 | 9.52 | 183.6 | 27.8 | 8.0 |
| pingap | Rust | /grande c=64 | 34563 | 6.72 | 193.3 | 19.7 | 8.0 |
| sozu | Rust | /chico c=64 | 150452 | 2.05 | 197.8 | 9.6 | 5.7 |
| sozu | Rust | /chico c=256 | 145120 | 6.27 | 197.4 | 15.1 | 5.7 |
| sozu | Rust | /grande c=64 | 39751 | 7.13 | 197.7 | 14.2 | 5.7 |

> Terminó 2026-11-27 08:27. CPU de Windows al final: 2 %.
