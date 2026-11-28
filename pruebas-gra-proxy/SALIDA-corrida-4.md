# Proxy inverso: Caddy y Traefik (Go) contra Pingap y Sōzu (Rust) · corrida 4

> Corrida del 2026-11-27 09:43, con el candado `_logs/banco.lock` tomado y sin `fab producir` vivo. Regenerable con `python investigacion/pruebas_gra_proxy.py 4`. Orden: sozu → pingap → traefik → caddy. Cada carga: wrk -t6 -d20s.

## 0 · Versiones, imágenes y configuración

- **caddy** (Go) `caddy:2.11.9` · imagen 23.7 MiB · config 7 líneas (Caddyfile) · `v2.11.9 h1:bhZtcRKQOq0fuLvEgCdPYvUxXyGn9kzjG4lh2332DTE=`
- **traefik** (Go) `traefik:v3.8.0` · imagen 61.3 MiB · config 19 líneas (traefik.yml, traefik-dinamico.yml) · `Version: 3.8.0 Codename: salers Go version: go1.27.4 Built: 2026-11-23T14:33:45Z OS/Arch: linux/amd64`
- **pingap** (Rust) `vicanso/pingap:0.16.2` · imagen 41.6 MiB · config 13 líneas (pingap.toml) · `pingap 0.16.2 (5c8c3130127020e48f37896c7d5c46e8ec3442fb, tls=openssl)`
- **sozu** (Rust) `clevercloud/sozu:2.3.7` · imagen 49.1 MiB · config 21 líneas (sozu.toml) · `sozu 2.3.7 () +jemallocator +crypto-ring -crypto-aws-lc-rs -crypto-openssl -fips -simd -splice -opentelemetry -tolerant-http1-parser -logs-debug -logs-trace -un`
- backend nginx:1.30.6-alpine (4 CPU) · cliente wrk (Alpine 3.24, 6 hilos, 6 CPU) · ajenos vivos: [] · CPU de Windows: 36 %

## 1 · ¿La misma respuesta a través de cada uno?

- directo al backend: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254
- sozu: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254 · ✅ igual (respondió a los 0.4 s de arrancar)
- pingap: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254 · ✅ igual (respondió a los 0.3 s de arrancar)
- traefik: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254 · ✅ igual (respondió a los 0.4 s de arrancar)
- caddy: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254 · ✅ igual (respondió a los 0.4 s de arrancar)

## 2 · Línea base: wrk directo al backend (sin proxy)

- /chico c=64: 811597 pet/s · p50 0.06 ms · p99 0.60 ms · códigos {'2xx': 16312233, 'no-2xx': 0} · errores de socket: connect 0, read 0, write 0, timeout 59
- /chico c=256: 834036 pet/s · p50 0.24 ms · p99 1.05 ms · códigos {'2xx': 16756541, 'no-2xx': 0}
- /grande c=64: 70751 pet/s · p50 0.74 ms · p99 1.98 ms · códigos {'2xx': 1422080, 'no-2xx': 0}

## 3 · A tope, cada proxy (2 CPU, 1 GiB)

### sozu · memoria en reposo 5.5 MiB

- /chico c=64: **129978 pet/s** · p50 0.38 ms · p99 1.35 ms · CPU media 197.8 % (máx 203.7) · memoria máx 9.7 MiB · hilos máx 3 · códigos {'2xx': 2605462, 'no-2xx': 0} · errores de socket: connect 0, read 0, write 0, timeout 60
  - ¿llegó todo al backend? wrk contó 2605462 respuestas (2605462 2xx); nginx atendió 2605516 por 60 conexiones nuevas del proxy (✅)
- /chico c=256: **112786 pet/s** · p50 1.95 ms · p99 4.68 ms · CPU media 196.8 % (máx 203.2) · memoria máx 14.9 MiB · hilos máx 3 · códigos {'2xx': 2263947, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 2263947 respuestas (2263947 2xx); nginx atendió 2264179 por 252 conexiones nuevas del proxy (✅)
- /grande c=64: **30383 pet/s** · p50 1.83 ms · p99 4.66 ms · CPU media 197.5 % (máx 203.1) · memoria máx 14.1 MiB · hilos máx 3 · códigos {'2xx': 610686, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 610686 respuestas (610686 2xx); nginx atendió 610738 por 60 conexiones nuevas del proxy (✅)

### pingap · memoria en reposo 8.0 MiB

- /chico c=64: **65837 pet/s** · p50 0.84 ms · p99 2.11 ms · CPU media 188.1 % (máx 192.6) · memoria máx 14.7 MiB · hilos máx 10 · códigos {'2xx': 1318549, 'no-2xx': 0} · errores de socket: connect 0, read 0, write 0, timeout 39
  - ¿llegó todo al backend? wrk contó 1318549 respuestas (1318549 2xx); nginx atendió 1318598 por 60 conexiones nuevas del proxy (✅)
- /chico c=256: **53151 pet/s** · p50 4.50 ms · p99 9.56 ms · CPU media 185.6 % (máx 192.3) · memoria máx 25.9 MiB · hilos máx 10 · códigos {'2xx': 1067495, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 1067495 respuestas (1067495 2xx); nginx atendió 1067740 por 251 conexiones nuevas del proxy (✅)
- /grande c=64: **23040 pet/s** · p50 2.49 ms · p99 5.06 ms · CPU media 194.2 % (máx 199.5) · memoria máx 20.1 MiB · hilos máx 10 · códigos {'2xx': 461583, 'no-2xx': 0} · errores de socket: connect 0, read 0, write 0, timeout 4
  - ¿llegó todo al backend? wrk contó 461583 respuestas (461583 2xx); nginx atendió 461638 por 60 conexiones nuevas del proxy (✅)

### traefik · memoria en reposo 17.6 MiB

- /chico c=64: **56330 pet/s** · p50 0.94 ms · p99 4.86 ms · CPU media 196.3 % (máx 202.3) · memoria máx 46.3 MiB · hilos máx 8 · códigos {'2xx': 1127900, 'no-2xx': 0} · errores de socket: connect 0, read 0, write 0, timeout 13
  - ¿llegó todo al backend? wrk contó 1127900 respuestas (1127900 2xx); nginx atendió 1127953 por 57 conexiones nuevas del proxy (✅)
- /chico c=256: **46850 pet/s** · p50 4.95 ms · p99 14.48 ms · CPU media 195.8 % (máx 202.6) · memoria máx 99.6 MiB · hilos máx 7 · códigos {'2xx': 941037, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 941037 respuestas (941037 2xx); nginx atendió 941282 por 287 conexiones nuevas del proxy (✅)
- /grande c=64: **21332 pet/s** · p50 2.56 ms · p99 8.61 ms · CPU media 196.9 % (máx 202.8) · memoria máx 51.7 MiB · hilos máx 8 · códigos {'2xx': 428772, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 428772 respuestas (428772 2xx); nginx atendió 428827 por 53 conexiones nuevas del proxy (✅)

### caddy · memoria en reposo 10.1 MiB

- /chico c=64: **39728 pet/s** · p50 1.29 ms · p99 5.25 ms · CPU media 194.7 % (máx 200.3) · memoria máx 26.0 MiB · hilos máx 8 · códigos {'2xx': 798523, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 798523 respuestas (798523 2xx); nginx atendió 798578 por 1638 conexiones nuevas del proxy (✅)
- /chico c=256: **5499 pet/s** · p50 30.63 ms · p99 1430.00 ms · CPU media 197.0 % (máx 225.9) · memoria máx 62.4 MiB · hilos máx 32 · códigos {'2xx': 110125, 'no-2xx': 389} · ⚠️ {'no-2xx': 389}
  - ¿llegó todo al backend? wrk contó 110514 respuestas (110125 2xx); nginx atendió 110284 por 28254 conexiones nuevas del proxy (✅)
- /grande c=64: **19016 pet/s** · p50 2.90 ms · p99 8.71 ms · CPU media 195.5 % (máx 202.0) · memoria máx 29.4 MiB · hilos máx 8 · códigos {'2xx': 380977, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 380977 respuestas (380977 2xx); nginx atendió 381035 por 421 conexiones nuevas del proxy (✅)

## 3b · Caddy con keepalive_idle_conns_per_host 1024 (por defecto 32)

- caddy-ka /chico c=64: **40400 pet/s** · p50 1.27 ms · p99 4.99 ms · CPU media 194.4 % · memoria máx 25.1 MiB · códigos {'2xx': 809453, 'no-2xx': 0} · nginx atendió 809508 por 57 conexiones nuevas
- caddy-ka /chico c=256: **29717 pet/s** · p50 8.03 ms · p99 18.70 ms · CPU media 195.0 % · memoria máx 57.0 MiB · códigos {'2xx': 596736, 'no-2xx': 0} · nginx atendió 596983 por 267 conexiones nuevas
- caddy-ka /grande c=64: **17913 pet/s** · p50 3.07 ms · p99 9.63 ms · CPU media 195.3 % · memoria máx 30.0 MiB · códigos {'2xx': 358740, 'no-2xx': 0} · nginx atendió 358796 por 65 conexiones nuevas

## 4 · La tabla

| proxy | lenguaje | carga | pet/s | p99 ms | CPU media % | memoria máx MiB | reposo MiB |
|---|---|---|---:|---:|---:|---:|---:|
| caddy | Go | /chico c=64 | 39728 | 5.25 | 194.7 | 26.0 | 10.1 |
| caddy | Go | /chico c=256 | 5499 | 1430.00 | 197.0 | 62.4 | 10.1 |
| caddy | Go | /grande c=64 | 19016 | 8.71 | 195.5 | 29.4 | 10.1 |
| traefik | Go | /chico c=64 | 56330 | 4.86 | 196.3 | 46.3 | 17.6 |
| traefik | Go | /chico c=256 | 46850 | 14.48 | 195.8 | 99.6 | 17.6 |
| traefik | Go | /grande c=64 | 21332 | 8.61 | 196.9 | 51.7 | 17.6 |
| pingap | Rust | /chico c=64 | 65837 | 2.11 | 188.1 | 14.7 | 8.0 |
| pingap | Rust | /chico c=256 | 53151 | 9.56 | 185.6 | 25.9 | 8.0 |
| pingap | Rust | /grande c=64 | 23040 | 5.06 | 194.2 | 20.1 | 8.0 |
| sozu | Rust | /chico c=64 | 129978 | 1.35 | 197.8 | 9.7 | 5.5 |
| sozu | Rust | /chico c=256 | 112786 | 4.68 | 196.8 | 14.9 | 5.5 |
| sozu | Rust | /grande c=64 | 30383 | 4.66 | 197.5 | 14.1 | 5.5 |

> Terminó 2026-11-27 09:51. CPU de Windows al final: 38 %.
