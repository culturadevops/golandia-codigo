# Telegraf (Go) contra Vector (Rust) · corrida 1

> Corrida del 2026-11-26 21:57, con el candado `_logs/banco.lock` tomado y sin `fab producir` vivo. Regenerable con `python investigacion/pruebas_gra_telegraf_vector.py 1`. Orden: telegraf → vector.

## 0 · Versiones y tamaños

- `Telegraf 1.40.4 (git: HEAD@c1b3d7f7)`
- `vector 0.62.0 (x86_64-unknown-linux-gnu b78d987 2026-11-16 16:23:48.988985759)`
- `/usr/local/bin/telegraf 334692512`
- `/usr/bin/vector 166583768`
- imagen oficial `telegraf:1.40.4`: 174.3 MiB
- imagen oficial `timberio/vector:0.62.0-debian`: 91.4 MiB
- `telegraf-json.conf`: 29 líneas sin comentarios
- `vector-json.yaml`: 23 líneas sin comentarios
- `telegraf-nginx.conf`: 17 líneas sin comentarios
- `vector-nginx.yaml`: 20 líneas sin comentarios
- contenedores ajenos vivos al empezar: ninguno · CPU de Windows: 9 %

## A · logs JSON, 1.000.000 de líneas → 900.000

### telegraf · repetición 1
```
RESULTADO lineas=900000 seg_total=4.31 seg_primera=0.25 cpu_s=7.48 pico_mib=228.1 final_mib=216.3 hilos=9
HUELLA 20c5de82055a filas=900000 suma_col2=200769000
MUESTRA {"fields":{"duracion_ms":7,"estado":200,"ruta":"/v1/pedidos/1"},"name":"log","tags":{"entorno":"prod","nivel":"warn","path":"/datos/eventos.json","servicio":"api-1"},"timestamp":1795651201}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

### vector · repetición 1
```
RESULTADO lineas=900000 seg_total=31.24 seg_primera=0.25 cpu_s=23.59 pico_mib=102.5 final_mib=101.4 hilos=9
HUELLA 20c5de82055a filas=900000 suma_col2=200769000
MUESTRA {"duracion_ms":7,"entorno":"prod","estado":200,"nivel":"warn","ruta":"/v1/pedidos/1","servicio":"api-1","ts":"2026-11-26T00:00:01Z"}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

### telegraf · repetición 2
```
RESULTADO lineas=900000 seg_total=4.32 seg_primera=0.25 cpu_s=7.25 pico_mib=225.1 final_mib=214.9 hilos=9
HUELLA 20c5de82055a filas=900000 suma_col2=200769000
MUESTRA {"fields":{"duracion_ms":7,"estado":200,"ruta":"/v1/pedidos/1"},"name":"log","tags":{"entorno":"prod","nivel":"warn","path":"/datos/eventos.json","servicio":"api-1"},"timestamp":1795651201}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

### vector · repetición 2
```
RESULTADO lineas=900000 seg_total=30.71 seg_primera=0.25 cpu_s=22.56 pico_mib=101.6 final_mib=96.2 hilos=10
HUELLA 20c5de82055a filas=900000 suma_col2=200769000
MUESTRA {"duracion_ms":7,"entorno":"prod","estado":200,"nivel":"warn","ruta":"/v1/pedidos/1","servicio":"api-1","ts":"2026-11-26T00:00:01Z"}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

**¿Hicieron el mismo trabajo?** huellas: telegraf 20c5de82055a, vector 20c5de82055a → ✅ iguales

## B · log de acceso de nginx, 1.000.000 de líneas

### telegraf · repetición 1
```
RESULTADO lineas=1000000 seg_total=377.17 seg_primera=1.27 cpu_s=394.61 pico_mib=212.4 final_mib=211.4 hilos=9
HUELLA 910926d8f09a filas=1000000 suma_col2=24999500000
MUESTRA {"fields":{"agent":"Mozilla/5.0 (X11; Linux x86_64) banco/1","auth":"-","client_ip":"10.1.0.1","http_version":1.1,"ident":"-","referrer":"https://golandia.dev/ref/1","request":"/api/v1/recurso/1?pagina=1","resp_bytes":31},"name":"acceso","tags":{"path":"/datos/access.log","resp_code":"200","verb":"POST"},"timestamp":1795651201}
FILA 10.1.0.1	31	200	/api/v1/recurso/1?pagina=1	POST
FILA 10.2.0.2	62	200	/api/v1/recurso/2?pagina=2	GET
```

### vector · repetición 1
```
RESULTADO lineas=1000000 seg_total=36.08 seg_primera=0.25 cpu_s=29.43 pico_mib=96.9 final_mib=96.0 hilos=9
HUELLA 910926d8f09a filas=1000000 suma_col2=24999500000
MUESTRA {"agent":"Mozilla/5.0 (X11; Linux x86_64) banco/1","client":"10.1.0.1","referer":"https://golandia.dev/ref/1","request":"POST /api/v1/recurso/1?pagina=1 HTTP/1.1","size":31,"status":200,"timestamp":"2026-11-26T00:00:01Z"}
FILA 10.1.0.1	31	200	/api/v1/recurso/1?pagina=1	POST
FILA 10.2.0.2	62	200	/api/v1/recurso/2?pagina=2	GET
```

### telegraf · repetición 2
```
RESULTADO lineas=1000000 seg_total=378.19 seg_primera=1.28 cpu_s=396.10 pico_mib=211.5 final_mib=210.3 hilos=9
HUELLA 910926d8f09a filas=1000000 suma_col2=24999500000
MUESTRA {"fields":{"agent":"Mozilla/5.0 (X11; Linux x86_64) banco/1","auth":"-","client_ip":"10.1.0.1","http_version":1.1,"ident":"-","referrer":"https://golandia.dev/ref/1","request":"/api/v1/recurso/1?pagina=1","resp_bytes":31},"name":"acceso","tags":{"path":"/datos/access.log","resp_code":"200","verb":"POST"},"timestamp":1795651201}
FILA 10.1.0.1	31	200	/api/v1/recurso/1?pagina=1	POST
FILA 10.2.0.2	62	200	/api/v1/recurso/2?pagina=2	GET
```

### vector · repetición 2
```
RESULTADO lineas=1000000 seg_total=28.15 seg_primera=0.25 cpu_s=23.18 pico_mib=98.9 final_mib=98.8 hilos=11
HUELLA 910926d8f09a filas=1000000 suma_col2=24999500000
MUESTRA {"agent":"Mozilla/5.0 (X11; Linux x86_64) banco/1","client":"10.1.0.1","referer":"https://golandia.dev/ref/1","request":"POST /api/v1/recurso/1?pagina=1 HTTP/1.1","size":31,"status":200,"timestamp":"2026-11-26T00:00:01Z"}
FILA 10.1.0.1	31	200	/api/v1/recurso/1?pagina=1	POST
FILA 10.2.0.2	62	200	/api/v1/recurso/2?pagina=2	GET
```

**¿Hicieron el mismo trabajo?** huellas: telegraf 910926d8f09a, vector 910926d8f09a → ✅ iguales

## La tabla (mediana de las repeticiones; eventos/s = líneas de entrada / segundos hasta la última línea escrita)

| pipeline | herramienta | segundos | eventos/s | CPU s | pico RAM MiB |
|---|---|---:|---:|---:|---:|
| json | telegraf | 4.31 | 231.750 | 7.37 | 226.6 |
| json | vector | 30.98 | 32.284 | 23.07 | 102.0 |
| nginx | telegraf | 377.68 | 2.648 | 395.36 | 211.9 |
| nginx | vector | 32.11 | 31.138 | 26.30 | 97.9 |

## La trampa: Telegraf con max_undelivered_lines por defecto (1000), pipeline A, 100.000 líneas → 90.000

### telegraf · json-defecto
```
RESULTADO lineas=90000 seg_total=90.18 seg_primera=1.27 cpu_s=1.10 pico_mib=190.4 final_mib=190.4 hilos=9
HUELLA fcdf91ad70fe filas=90000 suma_col2=20076900
MUESTRA {"fields":{"duracion_ms":7,"estado":200,"ruta":"/v1/pedidos/1"},"name":"log","tags":{"entorno":"prod","nivel":"warn","path":"/datos/eventos.json","servicio":"api-1"},"timestamp":1795651201}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

### telegraf · json
```
RESULTADO lineas=90000 seg_total=0.51 seg_primera=0.25 cpu_s=0.84 pico_mib=216.5 final_mib=201.8 hilos=9
HUELLA fcdf91ad70fe filas=90000 suma_col2=20076900
MUESTRA {"fields":{"duracion_ms":7,"estado":200,"ruta":"/v1/pedidos/1"},"name":"log","tags":{"entorno":"prod","nivel":"warn","path":"/datos/eventos.json","servicio":"api-1"},"timestamp":1795651201}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

### vector · json
```
RESULTADO lineas=90000 seg_total=2.79 seg_primera=0.25 cpu_s=1.94 pico_mib=101.0 final_mib=101.0 hilos=10
HUELLA fcdf91ad70fe filas=90000 suma_col2=20076900
MUESTRA {"duracion_ms":7,"entorno":"prod","estado":200,"nivel":"warn","ruta":"/v1/pedidos/1","servicio":"api-1","ts":"2026-11-26T00:00:01Z"}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

## Memoria en reposo (pipeline A cargado, sin archivo de entrada, 15 s)

- telegraf: `REPOSO rss_mib=151.4 cpu_s=0.08`
- vector: `REPOSO rss_mib=82.2 cpu_s=0.08`

> Terminó 2026-11-26 22:15. CPU de Windows al final: 4 %.
