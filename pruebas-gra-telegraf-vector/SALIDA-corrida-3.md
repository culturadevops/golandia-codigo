# Telegraf (Go) contra Vector (Rust) · corrida 3

> Corrida del 2026-11-27 08:27, con el candado `_logs/banco.lock` tomado y sin `fab producir` vivo. Regenerable con `python investigacion/pruebas_gra_telegraf_vector.py 3`. Orden: telegraf → vector.

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
- contenedores ajenos vivos al empezar: ninguno · CPU de Windows: 11 %

## A · logs JSON, 1.000.000 de líneas → 900.000

### telegraf · repetición 1
```
RESULTADO lineas=900000 seg_total=5.10 seg_primera=0.25 cpu_s=8.16 pico_mib=237.3 final_mib=216.1 hilos=9
HUELLA 20c5de82055a filas=900000 suma_col2=200769000
MUESTRA {"fields":{"duracion_ms":7,"estado":200,"ruta":"/v1/pedidos/1"},"name":"log","tags":{"entorno":"prod","nivel":"warn","path":"/datos/eventos.json","servicio":"api-1"},"timestamp":1795651201}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

### vector · repetición 1
```
RESULTADO lineas=900000 seg_total=43.46 seg_primera=0.25 cpu_s=31.75 pico_mib=101.1 final_mib=101.1 hilos=10
HUELLA 20c5de82055a filas=900000 suma_col2=200769000
MUESTRA {"duracion_ms":7,"entorno":"prod","estado":200,"nivel":"warn","ruta":"/v1/pedidos/1","servicio":"api-1","ts":"2026-11-26T00:00:01Z"}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

### telegraf · repetición 2
```
RESULTADO lineas=900000 seg_total=4.31 seg_primera=0.25 cpu_s=7.35 pico_mib=222.8 final_mib=215.9 hilos=9
HUELLA 20c5de82055a filas=900000 suma_col2=200769000
MUESTRA {"fields":{"duracion_ms":7,"estado":200,"ruta":"/v1/pedidos/1"},"name":"log","tags":{"entorno":"prod","nivel":"warn","path":"/datos/eventos.json","servicio":"api-1"},"timestamp":1795651201}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

### vector · repetición 2
```
RESULTADO lineas=900000 seg_total=26.87 seg_primera=0.25 cpu_s=21.19 pico_mib=103.6 final_mib=103.6 hilos=9
HUELLA 20c5de82055a filas=900000 suma_col2=200769000
MUESTRA {"duracion_ms":7,"entorno":"prod","estado":200,"nivel":"warn","ruta":"/v1/pedidos/1","servicio":"api-1","ts":"2026-11-26T00:00:01Z"}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

**¿Hicieron el mismo trabajo?** huellas: telegraf 20c5de82055a, vector 20c5de82055a → ✅ iguales

## A2 · lo mismo que A, con la salida por stdout (redirigida a un archivo)

### telegraf · repetición 1
```
RESULTADO lineas=900000 seg_total=4.31 seg_primera=0.25 cpu_s=7.42 pico_mib=222.8 final_mib=216.9 hilos=9
HUELLA 20c5de82055a filas=900000 suma_col2=200769000
MUESTRA {"fields":{"duracion_ms":7,"estado":200,"ruta":"/v1/pedidos/1"},"name":"log","tags":{"entorno":"prod","nivel":"warn","path":"/datos/eventos.json","servicio":"api-1"},"timestamp":1795651201}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

### vector · repetición 1
```
RESULTADO lineas=900000 seg_total=24.83 seg_primera=0.25 cpu_s=17.70 pico_mib=103.9 final_mib=103.3 hilos=10
HUELLA 20c5de82055a filas=900000 suma_col2=200769000
MUESTRA {"duracion_ms":7,"entorno":"prod","estado":200,"nivel":"warn","ruta":"/v1/pedidos/1","servicio":"api-1","ts":"2026-11-26T00:00:01Z"}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

### telegraf · repetición 2
```
RESULTADO lineas=900000 seg_total=4.06 seg_primera=0.25 cpu_s=7.06 pico_mib=226.6 final_mib=216.1 hilos=9
HUELLA 20c5de82055a filas=900000 suma_col2=200769000
MUESTRA {"fields":{"duracion_ms":7,"estado":200,"ruta":"/v1/pedidos/1"},"name":"log","tags":{"entorno":"prod","nivel":"warn","path":"/datos/eventos.json","servicio":"api-1"},"timestamp":1795651201}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

### vector · repetición 2
```
RESULTADO lineas=900000 seg_total=24.08 seg_primera=0.25 cpu_s=17.19 pico_mib=102.0 final_mib=101.2 hilos=9
HUELLA 20c5de82055a filas=900000 suma_col2=200769000
MUESTRA {"duracion_ms":7,"entorno":"prod","estado":200,"nivel":"warn","ruta":"/v1/pedidos/1","servicio":"api-1","ts":"2026-11-26T00:00:01Z"}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

**¿Hicieron el mismo trabajo?** huellas: telegraf 20c5de82055a, vector 20c5de82055a → ✅ iguales

## B · log de acceso de nginx, 1.000.000 de líneas

### telegraf · repetición 1
```
RESULTADO lineas=1000000 seg_total=423.36 seg_primera=1.26 cpu_s=447.91 pico_mib=214.4 final_mib=214.1 hilos=10
HUELLA 910926d8f09a filas=1000000 suma_col2=24999500000
MUESTRA {"fields":{"agent":"Mozilla/5.0 (X11; Linux x86_64) banco/1","auth":"-","client_ip":"10.1.0.1","http_version":1.1,"ident":"-","referrer":"https://golandia.dev/ref/1","request":"/api/v1/recurso/1?pagina=1","resp_bytes":31},"name":"acceso","tags":{"path":"/datos/access.log","resp_code":"200","verb":"POST"},"timestamp":1795651201}
FILA 10.1.0.1	31	200	/api/v1/recurso/1?pagina=1	POST
FILA 10.2.0.2	62	200	/api/v1/recurso/2?pagina=2	GET
```

### vector · repetición 1
```
RESULTADO lineas=1000000 seg_total=36.65 seg_primera=0.25 cpu_s=31.85 pico_mib=95.9 final_mib=95.3 hilos=10
HUELLA 910926d8f09a filas=1000000 suma_col2=24999500000
MUESTRA {"agent":"Mozilla/5.0 (X11; Linux x86_64) banco/1","client":"10.1.0.1","referer":"https://golandia.dev/ref/1","request":"POST /api/v1/recurso/1?pagina=1 HTTP/1.1","size":31,"status":200,"timestamp":"2026-11-26T00:00:01Z"}
FILA 10.1.0.1	31	200	/api/v1/recurso/1?pagina=1	POST
FILA 10.2.0.2	62	200	/api/v1/recurso/2?pagina=2	GET
```

### telegraf · repetición 2
```
RESULTADO lineas=1000000 seg_total=399.41 seg_primera=1.27 cpu_s=420.26 pico_mib=213.3 final_mib=213.3 hilos=9
HUELLA 910926d8f09a filas=1000000 suma_col2=24999500000
MUESTRA {"fields":{"agent":"Mozilla/5.0 (X11; Linux x86_64) banco/1","auth":"-","client_ip":"10.1.0.1","http_version":1.1,"ident":"-","referrer":"https://golandia.dev/ref/1","request":"/api/v1/recurso/1?pagina=1","resp_bytes":31},"name":"acceso","tags":{"path":"/datos/access.log","resp_code":"200","verb":"POST"},"timestamp":1795651201}
FILA 10.1.0.1	31	200	/api/v1/recurso/1?pagina=1	POST
FILA 10.2.0.2	62	200	/api/v1/recurso/2?pagina=2	GET
```

### vector · repetición 2
```
RESULTADO lineas=1000000 seg_total=36.07 seg_primera=0.25 cpu_s=31.86 pico_mib=93.3 final_mib=93.3 hilos=10
HUELLA 910926d8f09a filas=1000000 suma_col2=24999500000
MUESTRA {"agent":"Mozilla/5.0 (X11; Linux x86_64) banco/1","client":"10.1.0.1","referer":"https://golandia.dev/ref/1","request":"POST /api/v1/recurso/1?pagina=1 HTTP/1.1","size":31,"status":200,"timestamp":"2026-11-26T00:00:01Z"}
FILA 10.1.0.1	31	200	/api/v1/recurso/1?pagina=1	POST
FILA 10.2.0.2	62	200	/api/v1/recurso/2?pagina=2	GET
```

**¿Hicieron el mismo trabajo?** huellas: telegraf 910926d8f09a, vector 910926d8f09a → ✅ iguales

## La tabla (mediana de las repeticiones; eventos/s = líneas de entrada / segundos hasta la última línea escrita)

| pipeline | herramienta | segundos | eventos/s | CPU s | pico RAM MiB |
|---|---|---:|---:|---:|---:|
| json | telegraf | 4.71 | 212.540 | 7.75 | 230.1 |
| json | vector | 35.16 | 28.437 | 26.47 | 102.3 |
| json-stdout | telegraf | 4.18 | 238.949 | 7.24 | 224.7 |
| json-stdout | vector | 24.45 | 40.891 | 17.45 | 103.0 |
| nginx | telegraf | 411.38 | 2.431 | 434.09 | 213.9 |
| nginx | vector | 36.36 | 27.503 | 31.86 | 94.6 |

## La trampa: Telegraf con max_undelivered_lines por defecto (1000), pipeline A, 100.000 líneas → 90.000

### telegraf · json-defecto
```
RESULTADO lineas=90000 seg_total=90.31 seg_primera=1.27 cpu_s=1.09 pico_mib=191.9 final_mib=191.9 hilos=9
HUELLA fcdf91ad70fe filas=90000 suma_col2=20076900
MUESTRA {"fields":{"duracion_ms":7,"estado":200,"ruta":"/v1/pedidos/1"},"name":"log","tags":{"entorno":"prod","nivel":"warn","path":"/datos/eventos.json","servicio":"api-1"},"timestamp":1795651201}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

### telegraf · json
```
RESULTADO lineas=90000 seg_total=0.76 seg_primera=0.25 cpu_s=0.98 pico_mib=214.7 final_mib=202.3 hilos=9
HUELLA fcdf91ad70fe filas=90000 suma_col2=20076900
MUESTRA {"fields":{"duracion_ms":7,"estado":200,"ruta":"/v1/pedidos/1"},"name":"log","tags":{"entorno":"prod","nivel":"warn","path":"/datos/eventos.json","servicio":"api-1"},"timestamp":1795651201}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

### vector · json
```
RESULTADO lineas=90000 seg_total=3.81 seg_primera=0.25 cpu_s=2.48 pico_mib=99.3 final_mib=99.3 hilos=9
HUELLA fcdf91ad70fe filas=90000 suma_col2=20076900
MUESTRA {"duracion_ms":7,"entorno":"prod","estado":200,"nivel":"warn","ruta":"/v1/pedidos/1","servicio":"api-1","ts":"2026-11-26T00:00:01Z"}
FILA 7	200	/v1/pedidos/1	warn	prod
FILA 14	200	/v1/pedidos/2	error	prod
```

## Memoria en reposo (pipeline A cargado, sin archivo de entrada, 15 s)

- telegraf: `REPOSO rss_mib=152.6 cpu_s=0.08`
- vector: `REPOSO rss_mib=81.1 cpu_s=0.12`

> Terminó 2026-11-27 08:47. CPU de Windows al final: 38 %.
