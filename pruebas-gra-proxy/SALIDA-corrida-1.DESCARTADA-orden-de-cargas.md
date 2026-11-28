# Proxy inverso: Caddy y Traefik (Go) contra Pingap y Sōzu (Rust) · corrida 1

> Corrida del 2026-11-27 08:10, con el candado `_logs/banco.lock` tomado y sin `fab producir` vivo. Regenerable con `python investigacion/pruebas_gra_proxy.py 1`. Orden: caddy → traefik → pingap → sozu. Cada carga: wrk -t6 -d20s.

## 0 · Versiones, imágenes y configuración

- **caddy** (Go) `caddy:2.11.9` · imagen 23.7 MiB · config 7 líneas (Caddyfile) · `v2.11.9 h1:bhZtcRKQOq0fuLvEgCdPYvUxXyGn9kzjG4lh2332DTE=`
- **traefik** (Go) `traefik:v3.8.0` · imagen 61.3 MiB · config 19 líneas (traefik.yml, traefik-dinamico.yml) · `Version: 3.8.0 Codename: salers Go version: go1.27.4 Built: 2026-11-23T14:33:45Z OS/Arch: linux/amd64`
- **pingap** (Rust) `vicanso/pingap:0.16.2` · imagen 41.6 MiB · config 13 líneas (pingap.toml) · `pingap 0.16.2 (5c8c3130127020e48f37896c7d5c46e8ec3442fb, tls=openssl)`
- **sozu** (Rust) `clevercloud/sozu:2.3.7` · imagen 49.1 MiB · config 21 líneas (sozu.toml) · `sozu 2.3.7 () +jemallocator +crypto-ring -crypto-aws-lc-rs -crypto-openssl -fips -simd -splice -opentelemetry -tolerant-http1-parser -logs-debug -logs-trace -un`
- backend nginx:1.30.6-alpine (4 CPU) · cliente wrk (Alpine 3.24, 6 hilos, 6 CPU) · ajenos vivos: [] · CPU de Windows: 1 %

## 1 · ¿La misma respuesta a través de cada uno?

- directo al backend: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254
- caddy: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254 · ✅ igual (respondió a los 0.3 s de arrancar)
- traefik: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254 · ✅ igual (respondió a los 0.3 s de arrancar)
- pingap: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254 · ✅ igual (respondió a los 0.3 s de arrancar)
- sozu: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254 · ✅ igual (respondió a los 0.3 s de arrancar)

## 2 · Línea base: wrk directo al backend (sin proxy)

- /chico c=64: 1050702 pet/s · p50 0.05 ms · p99 0.15 ms · códigos {'2xx': 21118670, 'no-2xx': 0}
- /chico c=256: 1058034 pet/s · p50 0.22 ms · p99 0.47 ms · códigos {'2xx': 21260792, 'no-2xx': 0} · errores de socket: connect 0, read 0, write 0, timeout 236
- /grande c=64: 96120 pet/s · p50 0.60 ms · p99 0.98 ms · códigos {'2xx': 1926713, 'no-2xx': 0} · errores de socket: connect 0, read 0, write 0, timeout 59

## 3 · A tope, cada proxy (2 CPU, 1 GiB)

### caddy · memoria en reposo 10.8 MiB

- /chico c=64: **49990 pet/s** · p50 1.04 ms · p99 3.64 ms · CPU media 194.7 % (máx 200.1) · memoria máx 25.3 MiB · hilos máx 7 · códigos {'2xx': 1001560, 'no-2xx': 0} · errores de socket: connect 0, read 0, write 0, timeout 30
  - ¿llegó todo al backend? wrk contó 1001560 respuestas (1001560 2xx); nginx atendió 1001615 por 1738 conexiones nuevas del proxy (✅)
- /chico c=256: **5968 pet/s** · p50 26.85 ms · p99 1640.00 ms · CPU media 196.0 % (máx 232.5) · memoria máx 62.7 MiB · hilos máx 31 · códigos {'2xx': 118149, 'no-2xx': 1633} · ⚠️ {'no-2xx': 1633}
  - ¿llegó todo al backend? wrk contó 119782 respuestas (118149 2xx); nginx atendió 118308 por 26511 conexiones nuevas del proxy (✅)
- /grande c=64: **4072 pet/s** · p50 4.55 ms · p99 733.04 ms · CPU media 197.2 % (máx 243.7) · memoria máx 55.0 MiB · hilos máx 31 · códigos {'2xx': 78488, 'no-2xx': 3153} · ⚠️ {'no-2xx': 3153}
  - ¿llegó todo al backend? wrk contó 81641 respuestas (78488 2xx); nginx atendió 78541 por 91 conexiones nuevas del proxy (✅)

### traefik · memoria en reposo 17.3 MiB

- /chico c=64: **69865 pet/s** · p50 0.77 ms · p99 3.87 ms · CPU media 196.4 % (máx 202.2) · memoria máx 47.5 MiB · hilos máx 7 · códigos {'2xx': 1399057, 'no-2xx': 0} · errores de socket: connect 0, read 0, write 0, timeout 42
  - ¿llegó todo al backend? wrk contó 1399057 respuestas (1399057 2xx); nginx atendió 1399111 por 61 conexiones nuevas del proxy (✅)
- /chico c=256: **67171 pet/s** · p50 3.53 ms · p99 10.41 ms · CPU media 196.3 % (máx 202.6) · memoria máx 108.0 MiB · hilos máx 7 · códigos {'2xx': 1347015, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 1347015 respuestas (1347015 2xx); nginx atendió 1347262 por 233 conexiones nuevas del proxy (✅)
- /grande c=64: **29513 pet/s** · p50 1.92 ms · p99 5.45 ms · CPU media 197.3 % (máx 202.5) · memoria máx 102.7 MiB · hilos máx 7 · códigos {'2xx': 593208, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 593208 respuestas (593208 2xx); nginx atendió 593262 por 0 conexiones nuevas del proxy (✅)

### pingap · memoria en reposo 8.4 MiB

- /chico c=64: **80879 pet/s** · p50 0.74 ms · p99 1.19 ms · CPU media 186.9 % (máx 191.9) · memoria máx 14.8 MiB · hilos máx 10 · códigos {'2xx': 1625668, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 1625668 respuestas (1625668 2xx); nginx atendió 1625721 por 60 conexiones nuevas del proxy (✅)
- /chico c=256: **73210 pet/s** · p50 3.45 ms · p99 6.11 ms · CPU media 182.8 % (máx 189.7) · memoria máx 30.7 MiB · hilos máx 10 · códigos {'2xx': 1467063, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 1467063 respuestas (1467063 2xx); nginx atendió 1467308 por 188 conexiones nuevas del proxy (✅)
- /grande c=64: **35533 pet/s** · p50 1.70 ms · p99 2.59 ms · CPU media 191.8 % (máx 196.4) · memoria máx 36.0 MiB · hilos máx 10 · códigos {'2xx': 714172, 'no-2xx': 0} · errores de socket: connect 0, read 0, write 0, timeout 10
  - ¿llegó todo al backend? wrk contó 714172 respuestas (714172 2xx); nginx atendió 714226 por 3 conexiones nuevas del proxy (✅)

### sozu · memoria en reposo 5.5 MiB

- /chico c=64: **158765 pet/s** · p50 0.39 ms · p99 0.98 ms · CPU media 198.3 % (máx 203.2) · memoria máx 9.8 MiB · hilos máx 3 · códigos {'2xx': 3179972, 'no-2xx': 0} · errores de socket: connect 0, read 0, write 0, timeout 44
  - ¿llegó todo al backend? wrk contó 3179972 respuestas (3179972 2xx); nginx atendió 3180023 por 60 conexiones nuevas del proxy (✅)
- /chico c=256: **145049 pet/s** · p50 1.64 ms · p99 5.03 ms · CPU media 197.5 % (máx 203.2) · memoria máx 14.4 MiB · hilos máx 3 · códigos {'2xx': 2912583, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 2912583 respuestas (2912583 2xx); nginx atendió 2912828 por 252 conexiones nuevas del proxy (✅)
- /grande c=64: **39597 pet/s** · p50 1.58 ms · p99 2.15 ms · CPU media 198.0 % (máx 203.2) · memoria máx 17.5 MiB · hilos máx 3 · códigos {'2xx': 793389, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 793389 respuestas (793389 2xx); nginx atendió 793443 por 60 conexiones nuevas del proxy (✅)

## 3b · Caddy con keepalive_idle_conns_per_host 1024 (por defecto 32)

- caddy-ka /chico c=64: **52379 pet/s** · p50 1.00 ms · p99 3.50 ms · CPU media 194.5 % · memoria máx 25.2 MiB · códigos {'2xx': 1049183, 'no-2xx': 0} · nginx atendió 1049240 por 61 conexiones nuevas
- caddy-ka /chico c=256: **41517 pet/s** · p50 5.62 ms · p99 12.67 ms · CPU media 195.8 % · memoria máx 55.9 MiB · códigos {'2xx': 832504, 'no-2xx': 0} · nginx atendió 832752 por 187 conexiones nuevas

## 4 · La tabla

| proxy | lenguaje | carga | pet/s | p99 ms | CPU media % | memoria máx MiB | reposo MiB |
|---|---|---|---:|---:|---:|---:|---:|
| caddy | Go | /chico c=64 | 49990 | 3.64 | 194.7 | 25.3 | 10.8 |
| caddy | Go | /chico c=256 | 5968 | 1640.00 | 196.0 | 62.7 | 10.8 |
| caddy | Go | /grande c=64 | 4072 | 733.04 | 197.2 | 55.0 | 10.8 |
| traefik | Go | /chico c=64 | 69865 | 3.87 | 196.4 | 47.5 | 17.3 |
| traefik | Go | /chico c=256 | 67171 | 10.41 | 196.3 | 108.0 | 17.3 |
| traefik | Go | /grande c=64 | 29513 | 5.45 | 197.3 | 102.7 | 17.3 |
| pingap | Rust | /chico c=64 | 80879 | 1.19 | 186.9 | 14.8 | 8.4 |
| pingap | Rust | /chico c=256 | 73210 | 6.11 | 182.8 | 30.7 | 8.4 |
| pingap | Rust | /grande c=64 | 35533 | 2.59 | 191.8 | 36.0 | 8.4 |
| sozu | Rust | /chico c=64 | 158765 | 0.98 | 198.3 | 9.8 | 5.5 |
| sozu | Rust | /chico c=256 | 145049 | 5.03 | 197.5 | 14.4 | 5.5 |
| sozu | Rust | /grande c=64 | 39597 | 2.15 | 198.0 | 17.5 | 5.5 |

> Terminó 2026-11-27 08:17. CPU de Windows al final: 22 %.
