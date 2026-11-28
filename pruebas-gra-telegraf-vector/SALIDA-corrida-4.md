# Telegraf (Go) contra Vector (Rust) · corrida 4

> Corrida del 2026-11-27 08:48, con el candado `_logs/banco.lock` tomado y sin `fab producir` vivo. Regenerable con `python investigacion/pruebas_gra_telegraf_vector.py 4`. Orden: vector → telegraf.

## 0 · Versiones y tamaños

- `Telegraf 1.40.4 (git: HEAD@c1b3d7f7)`
- `vector 0.62.0 (x86_64-unknown-linux-gnu b78d987 2026-11-16 16:23:48.988985759)`
- `/usr/local/bin/telegraf 334692512 bytes = 319.2 MiB`
- `/usr/bin/vector 166583768 bytes = 158.9 MiB`
- imagen oficial `telegraf:1.40.4`: 174.3 MiB
- imagen oficial `timberio/vector:0.62.0-debian`: 91.4 MiB
- `telegraf-json.conf`: 29 líneas sin comentarios
- `vector-json.yaml`: 23 líneas sin comentarios
- `telegraf-json-stdout.conf`: 29 líneas sin comentarios
- `vector-json-stdout.yaml`: 22 líneas sin comentarios
- `telegraf-nginx.conf`: 17 líneas sin comentarios
- `vector-nginx.yaml`: 20 líneas sin comentarios
- contenedores ajenos vivos al empezar: ninguno · CPU de Windows: 20 %

## A · logs JSON, 1.000.000 de líneas → 900.000

### vector · repetición 1
```
RESULTADO lineas=900000 seg_total=29.98 seg_primera=0.25 cpu_s=25.77 pico_mib=104.1 final_mib=102.3 hilos=9
HUELLA 20c5de82055a filas=900000 suma_col2=200769000
MUESTRA {"duracion_ms":7,"entorno":"prod","estado":200,"nivel":"warn","ruta":"/v1/pedidos/1","servicio":"api-1","ts":"2026-11-26T00:00:01Z"}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

### telegraf · repetición 1
```
RESULTADO lineas=900000 seg_total=5.10 seg_primera=0.25 cpu_s=8.39 pico_mib=237.0 final_mib=219.0 hilos=9
HUELLA 20c5de82055a filas=900000 suma_col2=200769000
MUESTRA {"fields":{"duracion_ms":7,"estado":200,"ruta":"/v1/pedidos/1"},"name":"log","tags":{"entorno":"prod","nivel":"warn","path":"/datos/eventos.json","servicio":"api-1"},"timestamp":1795651201}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

### vector · repetición 2
```
RESULTADO lineas=900000 seg_total=32.56 seg_primera=0.25 cpu_s=27.58 pico_mib=102.6 final_mib=100.5 hilos=9
HUELLA 20c5de82055a filas=900000 suma_col2=200769000
MUESTRA {"duracion_ms":7,"entorno":"prod","estado":200,"nivel":"warn","ruta":"/v1/pedidos/1","servicio":"api-1","ts":"2026-11-26T00:00:01Z"}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

### telegraf · repetición 2
```
RESULTADO lineas=900000 seg_total=5.09 seg_primera=0.25 cpu_s=8.66 pico_mib=240.5 final_mib=211.3 hilos=9
HUELLA 20c5de82055a filas=900000 suma_col2=200769000
MUESTRA {"fields":{"duracion_ms":7,"estado":200,"ruta":"/v1/pedidos/1"},"name":"log","tags":{"entorno":"prod","nivel":"warn","path":"/datos/eventos.json","servicio":"api-1"},"timestamp":1795651201}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

**¿Hicieron el mismo trabajo?** huellas: vector 20c5de82055a, telegraf 20c5de82055a → ✅ iguales

## A2 · lo mismo que A, con la salida por stdout (redirigida a un archivo)

### vector · repetición 1
```
RESULTADO lineas=900000 seg_total=32.56 seg_primera=0.25 cpu_s=26.22 pico_mib=102.6 final_mib=102.6 hilos=10
HUELLA 20c5de82055a filas=900000 suma_col2=200769000
MUESTRA {"duracion_ms":7,"entorno":"prod","estado":200,"nivel":"warn","ruta":"/v1/pedidos/1","servicio":"api-1","ts":"2026-11-26T00:00:01Z"}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

### telegraf · repetición 1
```
RESULTADO lineas=900000 seg_total=5.08 seg_primera=0.25 cpu_s=8.65 pico_mib=228.6 final_mib=227.2 hilos=10
HUELLA 20c5de82055a filas=900000 suma_col2=200769000
MUESTRA {"fields":{"duracion_ms":7,"estado":200,"ruta":"/v1/pedidos/1"},"name":"log","tags":{"entorno":"prod","nivel":"warn","path":"/datos/eventos.json","servicio":"api-1"},"timestamp":1795651201}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

### vector · repetición 2
```
RESULTADO lineas=900000 seg_total=31.30 seg_primera=0.25 cpu_s=24.26 pico_mib=102.2 final_mib=102.2 hilos=9
HUELLA 20c5de82055a filas=900000 suma_col2=200769000
MUESTRA {"duracion_ms":7,"entorno":"prod","estado":200,"nivel":"warn","ruta":"/v1/pedidos/1","servicio":"api-1","ts":"2026-11-26T00:00:01Z"}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

### telegraf · repetición 2
```
RESULTADO lineas=900000 seg_total=5.34 seg_primera=0.25 cpu_s=8.74 pico_mib=227.6 final_mib=208.0 hilos=9
HUELLA 20c5de82055a filas=900000 suma_col2=200769000
MUESTRA {"fields":{"duracion_ms":7,"estado":200,"ruta":"/v1/pedidos/1"},"name":"log","tags":{"entorno":"prod","nivel":"warn","path":"/datos/eventos.json","servicio":"api-1"},"timestamp":1795651201}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

**¿Hicieron el mismo trabajo?** huellas: vector 20c5de82055a, telegraf 20c5de82055a → ✅ iguales

## B · log de acceso de nginx, 1.000.000 de líneas

### vector · repetición 1
```
RESULTADO lineas=1000000 seg_total=40.47 seg_primera=0.25 cpu_s=38.65 pico_mib=94.6 final_mib=94.6 hilos=10
HUELLA 910926d8f09a filas=1000000 suma_col2=24999500000
MUESTRA {"agent":"Mozilla/5.0 (X11; Linux x86_64) banco/1","client":"10.1.0.1","referer":"https://golandia.dev/ref/1","request":"POST /api/v1/recurso/1?pagina=1 HTTP/1.1","size":31,"status":200,"timestamp":"2026-11-26T00:00:01Z"}
FILA 10.1.0.1	31	200	/api/v1/recurso/1?pagina=1	POST
FILA 10.2.0.2	62	200	/api/v1/recurso/2?pagina=2	GET
```

### telegraf · repetición 1
```
RESULTADO lineas=1000000 seg_total=391.30 seg_primera=1.27 cpu_s=409.12 pico_mib=214.3 final_mib=214.3 hilos=9
HUELLA 910926d8f09a filas=1000000 suma_col2=24999500000
MUESTRA {"fields":{"agent":"Mozilla/5.0 (X11; Linux x86_64) banco/1","auth":"-","client_ip":"10.1.0.1","http_version":1.1,"ident":"-","referrer":"https://golandia.dev/ref/1","request":"/api/v1/recurso/1?pagina=1","resp_bytes":31},"name":"acceso","tags":{"path":"/datos/access.log","resp_code":"200","verb":"POST"},"timestamp":1795651201}
FILA 10.1.0.1	31	200	/api/v1/recurso/1?pagina=1	POST
FILA 10.2.0.2	62	200	/api/v1/recurso/2?pagina=2	GET
```

### vector · repetición 2
```
RESULTADO lineas=1000000 seg_total=28.15 seg_primera=0.25 cpu_s=23.56 pico_mib=98.4 final_mib=96.6 hilos=10
HUELLA 910926d8f09a filas=1000000 suma_col2=24999500000
MUESTRA {"agent":"Mozilla/5.0 (X11; Linux x86_64) banco/1","client":"10.1.0.1","referer":"https://golandia.dev/ref/1","request":"POST /api/v1/recurso/1?pagina=1 HTTP/1.1","size":31,"status":200,"timestamp":"2026-11-26T00:00:01Z"}
FILA 10.1.0.1	31	200	/api/v1/recurso/1?pagina=1	POST
FILA 10.2.0.2	62	200	/api/v1/recurso/2?pagina=2	GET
```

### telegraf · repetición 2
```
RESULTADO lineas=1000000 seg_total=389.16 seg_primera=1.27 cpu_s=408.27 pico_mib=213.2 final_mib=213.2 hilos=9
HUELLA 910926d8f09a filas=1000000 suma_col2=24999500000
MUESTRA {"fields":{"agent":"Mozilla/5.0 (X11; Linux x86_64) banco/1","auth":"-","client_ip":"10.1.0.1","http_version":1.1,"ident":"-","referrer":"https://golandia.dev/ref/1","request":"/api/v1/recurso/1?pagina=1","resp_bytes":31},"name":"acceso","tags":{"path":"/datos/access.log","resp_code":"200","verb":"POST"},"timestamp":1795651201}
FILA 10.1.0.1	31	200	/api/v1/recurso/1?pagina=1	POST
FILA 10.2.0.2	62	200	/api/v1/recurso/2?pagina=2	GET
```

**¿Hicieron el mismo trabajo?** huellas: vector 910926d8f09a, telegraf 910926d8f09a → ✅ iguales

## La tabla (mediana de las repeticiones; eventos/s = líneas de entrada / segundos hasta la última línea escrita)

| pipeline | herramienta | segundos | eventos/s | CPU s | pico RAM MiB |
|---|---|---:|---:|---:|---:|
| json | telegraf | 5.09 | 196.271 | 8.53 | 238.8 |
| json | vector | 31.27 | 31.980 | 26.67 | 103.3 |
| json-stdout | telegraf | 5.21 | 191.939 | 8.70 | 228.1 |
| json-stdout | vector | 31.93 | 31.319 | 25.24 | 102.4 |
| nginx | telegraf | 390.23 | 2.563 | 408.69 | 213.8 |
| nginx | vector | 34.31 | 29.146 | 31.10 | 96.5 |

## La trampa: Telegraf con max_undelivered_lines por defecto (1000), pipeline A, 100.000 líneas → 90.000

### telegraf · json-defecto
```
RESULTADO lineas=90000 seg_total=90.25 seg_primera=1.27 cpu_s=1.11 pico_mib=188.9 final_mib=188.9 hilos=9
HUELLA fcdf91ad70fe filas=90000 suma_col2=20076900
MUESTRA {"fields":{"duracion_ms":7,"estado":200,"ruta":"/v1/pedidos/1"},"name":"log","tags":{"entorno":"prod","nivel":"warn","path":"/datos/eventos.json","servicio":"api-1"},"timestamp":1795651201}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

### telegraf · json
```
RESULTADO lineas=90000 seg_total=0.76 seg_primera=0.25 cpu_s=0.97 pico_mib=218.8 final_mib=210.4 hilos=9
HUELLA fcdf91ad70fe filas=90000 suma_col2=20076900
MUESTRA {"fields":{"duracion_ms":7,"estado":200,"ruta":"/v1/pedidos/1"},"name":"log","tags":{"entorno":"prod","nivel":"warn","path":"/datos/eventos.json","servicio":"api-1"},"timestamp":1795651201}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

### vector · json
```
RESULTADO lineas=90000 seg_total=3.56 seg_primera=0.25 cpu_s=2.69 pico_mib=101.5 final_mib=101.5 hilos=8
HUELLA fcdf91ad70fe filas=90000 suma_col2=20076900
MUESTRA {"duracion_ms":7,"entorno":"prod","estado":200,"nivel":"warn","ruta":"/v1/pedidos/1","servicio":"api-1","ts":"2026-11-26T00:00:01Z"}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

## Memoria en reposo (pipeline A cargado, sin archivo de entrada, 15 s)

- telegraf: `REPOSO rss_mib=151.5 cpu_s=0.08`
- vector: `REPOSO rss_mib=80.6 cpu_s=0.10`

> Terminó 2026-11-27 09:08. CPU de Windows al final: 35 %.
