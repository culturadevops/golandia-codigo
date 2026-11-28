# Proxy inverso: Caddy y Traefik (Go) contra Pingap y Sōzu (Rust) · corrida 1 (HUMO)

> Corrida del 2026-11-27 08:08, con el candado `_logs/banco.lock` tomado y sin `fab producir` vivo. Regenerable con `python investigacion/pruebas_gra_proxy.py 1`. Orden: caddy → traefik → pingap → sozu. Cada carga: wrk -t6 -d5s.

## 0 · Versiones, imágenes y configuración

- **caddy** (Go) `caddy:2.11.9` · imagen 23.7 MiB · config 7 líneas (Caddyfile) · `v2.11.9 h1:bhZtcRKQOq0fuLvEgCdPYvUxXyGn9kzjG4lh2332DTE=`
- **traefik** (Go) `traefik:v3.8.0` · imagen 61.3 MiB · config 19 líneas (traefik.yml, traefik-dinamico.yml) · `Version: 3.8.0 Codename: salers Go version: go1.27.4 Built: 2026-11-23T14:33:45Z OS/Arch: linux/amd64`
- **pingap** (Rust) `vicanso/pingap:0.16.2` · imagen 41.6 MiB · config 13 líneas (pingap.toml) · `pingap 0.16.2 (5c8c3130127020e48f37896c7d5c46e8ec3442fb, tls=openssl)`
- **sozu** (Rust) `clevercloud/sozu:2.3.7` · imagen 49.1 MiB · config 21 líneas (sozu.toml) · `sozu 2.3.7 () +jemallocator +crypto-ring -crypto-aws-lc-rs -crypto-openssl -fips -simd -splice -opentelemetry -tolerant-http1-parser -logs-debug -logs-trace -un`
- backend nginx:1.30.6-alpine (4 CPU) · cliente wrk (Alpine 3.24, 6 hilos, 6 CPU) · ajenos vivos: [] · CPU de Windows: 3 %

## 1 · ¿La misma respuesta a través de cada uno?

- directo al backend: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254
- caddy: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254 · ✅ igual (respondió a los 0.3 s de arrancar)
- traefik: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254 · ✅ igual (respondió a los 0.3 s de arrancar)
- pingap: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254 · ✅ igual (respondió a los 0.3 s de arrancar)
- sozu: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254 · ✅ igual (respondió a los 0.3 s de arrancar)

## 2 · Línea base: wrk directo al backend (sin proxy)

- /chico c=64: 1039414 pet/s · p50 0.05 ms · p99 0.13 ms · códigos {'2xx': 5242420, 'no-2xx': 0} · errores de socket: connect 0, read 0, write 0, timeout 53
- /chico c=256: 1046871 pet/s · p50 0.22 ms · p99 0.44 ms · códigos {'2xx': 5329475, 'no-2xx': 0}
- /grande c=64: 96954 pet/s · p50 0.59 ms · p99 0.95 ms · códigos {'2xx': 494440, 'no-2xx': 0}

## 3 · A tope, cada proxy (2 CPU, 1 GiB)

### caddy · memoria en reposo 10.7 MiB

- /chico c=64: **51063 pet/s** · p50 1.01 ms · p99 3.81 ms · CPU media 187.7 % (máx 200.3) · memoria máx 25.2 MiB · hilos máx 8 · códigos {'2xx': 260404, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 260404 respuestas (260404 2xx); nginx atendió 260457 por 499 conexiones nuevas del proxy (✅)
- /chico c=256: **13762 pet/s** · p50 12.34 ms · p99 90.42 ms · CPU media 188.4 % (máx 201.8) · memoria máx 61.8 MiB · hilos máx 9 · códigos {'2xx': 69886, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 69886 respuestas (69886 2xx); nginx atendió 70136 por 21133 conexiones nuevas del proxy (✅)
- /grande c=64: **25615 pet/s** · p50 2.12 ms · p99 6.26 ms · CPU media 182.0 % (máx 201.8) · memoria máx 54.0 MiB · hilos máx 14 · códigos {'2xx': 130625, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 130625 respuestas (130625 2xx); nginx atendió 130678 por 105 conexiones nuevas del proxy (✅)

### traefik · memoria en reposo 18.1 MiB

- /chico c=64: **69141 pet/s** · p50 0.78 ms · p99 4.36 ms · CPU media 190.3 % (máx 201.5) · memoria máx 48.5 MiB · hilos máx 7 · códigos {'2xx': 352613, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 352613 respuestas (352613 2xx); nginx atendió 352665 por 64 conexiones nuevas del proxy (✅)
- /chico c=256: **65943 pet/s** · p50 3.54 ms · p99 11.43 ms · CPU media 189.6 % (máx 203.0) · memoria máx 106.1 MiB · hilos máx 7 · códigos {'2xx': 334328, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 334328 respuestas (334328 2xx); nginx atendió 334575 por 252 conexiones nuevas del proxy (✅)
- /grande c=64: **30649 pet/s** · p50 1.83 ms · p99 5.39 ms · CPU media 192.6 % (máx 201.3) · memoria máx 97.9 MiB · hilos máx 7 · códigos {'2xx': 156307, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 156307 respuestas (156307 2xx); nginx atendió 156361 por 0 conexiones nuevas del proxy (✅)

### pingap · memoria en reposo 8.0 MiB

- /chico c=64: **79882 pet/s** · p50 0.74 ms · p99 1.23 ms · CPU media 180.4 % (máx 190.3) · memoria máx 14.0 MiB · hilos máx 10 · códigos {'2xx': 407348, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 407348 respuestas (407348 2xx); nginx atendió 407402 por 58 conexiones nuevas del proxy (✅)
- /chico c=256: **72265 pet/s** · p50 3.46 ms · p99 7.89 ms · CPU media 175.7 % (máx 188.2) · memoria máx 29.4 MiB · hilos máx 10 · códigos {'2xx': 368502, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 368502 respuestas (368502 2xx); nginx atendió 368747 por 187 conexiones nuevas del proxy (✅)
- /grande c=64: **35296 pet/s** · p50 1.69 ms · p99 2.81 ms · CPU media 187.2 % (máx 196.6) · memoria máx 35.2 MiB · hilos máx 10 · códigos {'2xx': 180010, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 180010 respuestas (180010 2xx); nginx atendió 180065 por 0 conexiones nuevas del proxy (✅)

### sozu · memoria en reposo 5.3 MiB

- /chico c=64: **162368 pet/s** · p50 0.36 ms · p99 0.58 ms · CPU media 191.5 % (máx 203.2) · memoria máx 9.2 MiB · hilos máx 3 · códigos {'2xx': 828031, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 828031 respuestas (828031 2xx); nginx atendió 828081 por 60 conexiones nuevas del proxy (✅)
- /chico c=256: **151501 pet/s** · p50 1.62 ms · p99 2.08 ms · CPU media 189.3 % (máx 203.2) · memoria máx 14.1 MiB · hilos máx 3 · códigos {'2xx': 768581, 'no-2xx': 0}
  - ¿llegó todo al backend? wrk contó 768581 respuestas (768581 2xx); nginx atendió 768814 por 252 conexiones nuevas del proxy (✅)
- /grande c=64: **39200 pet/s** · p50 1.58 ms · p99 2.20 ms · CPU media 192.9 % (máx 203.1) · memoria máx 17.2 MiB · hilos máx 3 · códigos {'2xx': 197139, 'no-2xx': 0} · errores de socket: connect 0, read 0, write 0, timeout 22
  - ¿llegó todo al backend? wrk contó 197139 respuestas (197139 2xx); nginx atendió 197194 por 60 conexiones nuevas del proxy (✅)

## 3b · Caddy con keepalive_idle_conns_per_host 1024 (por defecto 32)

- caddy-ka /chico c=64: **51567 pet/s** · p50 1.01 ms · p99 3.52 ms · CPU media 189.5 % · memoria máx 25.0 MiB · códigos {'2xx': 262998, 'no-2xx': 0} · nginx atendió 263050 por 53 conexiones nuevas
- caddy-ka /chico c=256: **46105 pet/s** · p50 4.91 ms · p99 14.32 ms · CPU media 187.8 % · memoria máx 56.1 MiB · códigos {'2xx': 233479, 'no-2xx': 0} · nginx atendió 233725 por 251 conexiones nuevas

## 4 · La tabla

| proxy | lenguaje | carga | pet/s | p99 ms | CPU media % | memoria máx MiB | reposo MiB |
|---|---|---|---:|---:|---:|---:|---:|
| caddy | Go | /chico c=64 | 51063 | 3.81 | 187.7 | 25.2 | 10.7 |
| caddy | Go | /chico c=256 | 13762 | 90.42 | 188.4 | 61.8 | 10.7 |
| caddy | Go | /grande c=64 | 25615 | 6.26 | 182.0 | 54.0 | 10.7 |
| traefik | Go | /chico c=64 | 69141 | 4.36 | 190.3 | 48.5 | 18.1 |
| traefik | Go | /chico c=256 | 65943 | 11.43 | 189.6 | 106.1 | 18.1 |
| traefik | Go | /grande c=64 | 30649 | 5.39 | 192.6 | 97.9 | 18.1 |
| pingap | Rust | /chico c=64 | 79882 | 1.23 | 180.4 | 14.0 | 8.0 |
| pingap | Rust | /chico c=256 | 72265 | 7.89 | 175.7 | 29.4 | 8.0 |
| pingap | Rust | /grande c=64 | 35296 | 2.81 | 187.2 | 35.2 | 8.0 |
| sozu | Rust | /chico c=64 | 162368 | 0.58 | 191.5 | 9.2 | 5.3 |
| sozu | Rust | /chico c=256 | 151501 | 2.08 | 189.3 | 14.1 | 5.3 |
| sozu | Rust | /grande c=64 | 39200 | 2.20 | 192.9 | 17.2 | 5.3 |

> Terminó 2026-11-27 08:10. CPU de Windows al final: 2 %.
