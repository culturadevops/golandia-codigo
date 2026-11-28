# Proxy inverso: Caddy y Traefik (Go) contra Pingap y Sōzu (Rust) · corrida 1

> Corrida del 2026-11-26 22:15, con el candado `_logs/banco.lock` tomado y sin `fab producir` vivo. Regenerable con `python investigacion/pruebas_gra_proxy.py 1`. Orden: caddy → traefik → pingap → sozu. Cada carga: hey -z 20s.

## 0 · Versiones, imágenes y configuración

- **caddy** (Go) `caddy:2.11.9` · imagen 23.7 MiB · config 7 líneas (Caddyfile) · `v2.11.9 h1:bhZtcRKQOq0fuLvEgCdPYvUxXyGn9kzjG4lh2332DTE=`
- **traefik** (Go) `traefik:v3.8.0` · imagen 61.3 MiB · config 19 líneas (traefik.yml, traefik-dinamico.yml) · `Version: 3.8.0 Codename: salers Go version: go1.27.4 Built: 2026-11-23T14:33:45Z OS/Arch: linux/amd64`
- **pingap** (Rust) `vicanso/pingap:0.16.2` · imagen 41.6 MiB · config 13 líneas (pingap.toml) · `pingap 0.16.2 (5c8c3130127020e48f37896c7d5c46e8ec3442fb, tls=openssl)`
- **sozu** (Rust) `clevercloud/sozu:2.3.7` · imagen 49.1 MiB · config 21 líneas (sozu.toml) · `sozu 2.3.7 () +jemallocator +crypto-ring -crypto-aws-lc-rs -crypto-openssl -fips -simd -splice -opentelemetry -tolerant-http1-parser -logs-debug -logs-trace -un`
- backend nginx:1.30.6-alpine (4 CPU) · cliente hey (6 CPU) · ajenos vivos: [] · CPU de Windows: 2 %

## 1 · ¿La misma respuesta a través de cada uno?

- directo al backend: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254
- caddy: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254 · ✅ igual (respondió a los 0.3 s de arrancar)
- traefik: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254 · ✅ igual (respondió a los 0.3 s de arrancar)
- pingap: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254 · ✅ igual (respondió a los 0.3 s de arrancar)
- sozu: /chico 200 63 bytes · /grande 200 102400 bytes md5 14d92c271254 · ✅ igual (respondió a los 0.3 s de arrancar)

## 2 · Línea base: hey directo al backend (sin proxy)

- /chico c=64: 170989 pet/s · p50 0.10 ms · p99 1.30 ms · códigos {'200': '1000000'}
- /chico c=256: 159402 pet/s · p50 0.10 ms · p99 48.60 ms · códigos {'200': '1000000'}
- /grande c=64: 72803 pet/s · p50 0.90 ms · p99 1.60 ms · códigos {'200': '1000000'}

## 3 · A tope, cada proxy (2 CPU, 1 GiB)

### caddy · memoria en reposo 10.6 MiB

  - ¿llegó todo al backend? hey contó 932755 respuestas; nginx atendió 932755 por 4517 conexiones nuevas del proxy (✅)
- /chico c=64: **46634 pet/s** · p50 1.10 ms · p99 4.10 ms · CPU media 195.0 % (máx 200.8) · memoria máx 28.3 MiB · hilos máx 8 · códigos {'200': '932755'}
  - ¿llegó todo al backend? hey contó 100609 respuestas; nginx atendió 99244 por 23715 conexiones nuevas del proxy (⛔ NO cuadra)
