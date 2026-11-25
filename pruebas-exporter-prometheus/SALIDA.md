# SALIDA — Prometheus exporter en Go (golandia 679, 2026-10-26)

Ejecutado el 2026-10-26 con `python correr.py` (todo en Docker, red `g679`, contenedores `g679-*`, borrados al terminar).
Versiones: Docker 29.6.2 · golang:1.27.2 · github.com/prometheus/client_golang v1.25.0 · prom/prometheus:v3.15.1 ·
prom/alertmanager:v0.34.0 · prom/blackbox-exporter:v0.29.0.
Servidores de prueba (`servidores/main.go`): generan su propia CA al arrancar; `:9441` vence en 5 días, `:9442` venció hace 1 día,
`:9443` vence en 90 días, `:9444` acepta TCP y nunca contesta el handshake (el «lento»). No se toca Internet.
Configuración usada (Prometheus, reglas, Alertmanager, blackbox): `salidas/config/`. Salidas sin editar: `salidas/*.txt`.

## 0 · Compilar

```
$ go build -trimpath -ldflags='-s -w' ./cmd/...
go version go1.27.2 linux/amd64
github.com/prometheus/client_golang v1.25.0
total 25480
-rwxr-xr-x 1 root root 9244832 Oct 26 15:38 certexp
-rwxr-xr-x 1 root root 9293984 Oct 26 15:38 gauge-global
-rwxr-xr-x 1 root root 7543255 Oct 26 15:38 servidores
```

## 1 · Los servidores de prueba

```
$ docker logs g679-srv
2026/10/26 15:38:24 :9441  vence 2026-10-31T15:38:24Z
2026/10/26 15:38:24 :9442  vence 2026-10-25T15:38:24Z
2026/10/26 15:38:24 :9443  vence 2027-01-24T15:38:24Z
2026/10/26 15:38:24 :9444  lento (no contesta el handshake)
```

## 2 · /metrics del exporter (Collector, timeout 2 s por host)

```
$ curl -s localhost:9679/metrics | grep ^certexp   (tardó 2.02 s)
certexp_not_after_timestamp_seconds{objetivo="g679-srv:9441"} 1.793461104e+09
certexp_not_after_timestamp_seconds{objetivo="g679-srv:9442"} 1.792942704e+09
certexp_not_after_timestamp_seconds{objetivo="g679-srv:9443"} 1.800805104e+09
certexp_probe_duration_seconds{objetivo="g679-srv:9441"} 0.001252784
certexp_probe_duration_seconds{objetivo="g679-srv:9442"} 0.001307184
certexp_probe_duration_seconds{objetivo="g679-srv:9443"} 0.001248804
certexp_probe_duration_seconds{objetivo="g679-srv:9444"} 2.000822174
certexp_probe_success{objetivo="g679-srv:9441"} 1
certexp_probe_success{objetivo="g679-srv:9442"} 1
certexp_probe_success{objetivo="g679-srv:9443"} 1
certexp_probe_success{objetivo="g679-srv:9444"} 0
```

## 3 · Timeouts: 3 hosts + 2 lentos

```
objetivos: 3 sanos/por vencer + 2 lentos (g679-srv:9444,g679-srv2:9444)
-timeout 0 (sin límite)    scrape 15.02 s · sin respuesta (TimeoutError: timed out)
-timeout 2s -secuencial    scrape  4.02 s · 3 sondas bien, 13 líneas
-timeout 2s en paralelo    scrape  2.03 s · 3 sondas bien, 13 líneas
```

Con `-timeout 0` el cliente (urllib, 15 s) se rinde sin respuesta: un solo host colgado deja sin /metrics a TODOS. Secuencial = 2 lentos × 2 s = 4,02 s; en paralelo = 2,03 s.

## 4 · Gauge global frente a Collector cuando el host se cae

```
== con g679-srv2 vivo
[gauge-global]
  certexp_not_after_timestamp_seconds{objetivo="g679-srv2:9441"} 1.793461104e+09
  certexp_not_after_timestamp_seconds{objetivo="g679-srv2:9443"} 1.800805104e+09
[certexp (Collector)]
  certexp_not_after_timestamp_seconds{objetivo="g679-srv2:9441"} 1.793461104e+09
  certexp_not_after_timestamp_seconds{objetivo="g679-srv2:9443"} 1.800805104e+09
  certexp_probe_duration_seconds{objetivo="g679-srv2:9441"} 0.001317409
  certexp_probe_duration_seconds{objetivo="g679-srv2:9443"} 0.001356411
  certexp_probe_success{objetivo="g679-srv2:9441"} 1
  certexp_probe_success{objetivo="g679-srv2:9443"} 1
== docker stop g679-srv2 · 5.0 s después
[gauge-global]
  certexp_not_after_timestamp_seconds{objetivo="g679-srv2:9441"} 1.793461104e+09
  certexp_not_after_timestamp_seconds{objetivo="g679-srv2:9443"} 1.800805104e+09
[certexp (Collector)]
  certexp_probe_duration_seconds{objetivo="g679-srv2:9441"} 2.000499285
  certexp_probe_duration_seconds{objetivo="g679-srv2:9443"} 2.000573903
  certexp_probe_success{objetivo="g679-srv2:9441"} 0
  certexp_probe_success{objetivo="g679-srv2:9443"} 0
(gauge-global con promhttp.Handler() expone además go_* y process_*: 132 líneas en total)
```

5 s después de `docker stop`, el Gauge global sigue diciendo que el certificado vence el mismo día (valor viejo, para siempre: el código ingenuo solo hace `continue` en el error). El Collector ya no expone la fecha y `certexp_probe_success` = 0.

## 5 · Alertas en Alertmanager (regla «vence en menos de 14 días»)

```
Prometheus v3.15.1 · Alertmanager v0.34.0 · scrape cada 15 s · for: 30s
segundos desde que arrancó Prometheus hasta que la alerta llegó a Alertmanager:
   61.1 s  CertificadoPorVencer   certexp                  g679-srv:9441 vence en 5.0 días
   61.1 s  SondaTLSFalla          certexp-etiqueta-error   no pude leer el certificado de g679-srv:9999
   61.1 s  SondaTLSFalla          certexp-etiqueta-error   no pude leer el certificado de g679-srv:9444
   61.1 s  SondaTLSFalla          certexp                  no pude leer el certificado de g679-srv:9444
   61.1 s  CertificadoPorVencer   certexp                  g679-srv:9442 vence en -1.0 días
```

Las 5 alertas llegan a Alertmanager a los 61,1 s de arrancar Prometheus (scrape 15 s + evaluación 15 s + `for: 30s`). El vencido (-1,0 días) también dispara «por vencer» porque -1 < 14.

## 6 · Prometheus: días restantes, up y scrape_duration_seconds

```
$ (certexp_not_after_timestamp_seconds - time()) / 86400
  {objetivo="g679-srv:9441"}  5.00
  {objetivo="g679-srv:9442"}  -1.00
  {objetivo="g679-srv:9443"}  90.00
$ certexp_probe_success{job="certexp"}
  {objetivo="g679-srv:9441"}  1
  {objetivo="g679-srv:9442"}  1
  {objetivo="g679-srv:9443"}  1
  {objetivo="g679-srv:9444"}  0
$ certexp_probe_duration_seconds{job="certexp"}
  {objetivo="g679-srv:9441"}  0.001
  {objetivo="g679-srv:9442"}  0.001
  {objetivo="g679-srv:9443"}  0.001
  {objetivo="g679-srv:9444"}  2.002
$ up · scrape_duration_seconds
  certexp                up=1  scrape_duration_seconds=2.00
  certexp-sin-timeout    up=0  scrape_duration_seconds=10.00
  lastError de certexp-sin-timeout: Get "http://g679-cx-sin:9679/metrics": context deadline exceeded
```

El job `certexp-sin-timeout` (exporter con `-timeout 0` y un host lento) queda `up=0` a los 10,00 s (scrape_timeout): Prometheus pierde TODAS sus métricas, también las de los hosts sanos.

## 7 · Cardinalidad: el texto del error como etiqueta

```
job certexp-etiqueta-error: objetivos g679-srv:9443 (sano), :9444 (lento), :9999 (cerrado) · scrape cada 5 s
  ahora: {objetivo="g679-srv:9444", error="context deadline exceeded"}
  ahora: {objetivo="g679-srv:9999", error="dial tcp 192.168.208.2:9999: connect: connection refused"}
  series distintas en 1m: certexp_probe_error=2 · certexp_probe_success=3
```

En esta corrida, 2 series de error (los textos fueron estables). Pero el texto lleva la IP del host (`192.168.208.2`): si cambia la IP o el mensaje, es otra serie. Etiquetas = valores que controlas (la lista de objetivos).

## 8 · blackbox_exporter oficial con los mismos hosts

```
blackbox_exporter v0.29.0 · módulo tls_connect (tcp + tls, ca_file con nuestra CA)
  g679-srv:9441    probe_success=1  días=5.00  probe_duration_seconds=0.022
  g679-srv:9442    probe_success=0  días=sin serie  probe_duration_seconds=0.003
  g679-srv:9443    probe_success=1  días=90.00  probe_duration_seconds=0.004
  g679-srv:9444    probe_success=0  días=sin serie  probe_duration_seconds=2.001
$ curl 'localhost:9115/probe?module=tls_connect&target=g679-srv:9442&debug=true' | grep -i -E 'error|expir'
  time=2026-10-26T15:40:27.532Z level=ERROR source=tcp.go:92 msg="Error dialing TCP" module=tls_connect target=g679-srv:9442 err="tls: failed to verify certificate: x509: certificate has expired or is not yet valid: current time 2026-10-26T15:40:27Z is after 2026-10-25T15:38:24Z"
  time=2026-10-26T15:40:27.532Z level=ERROR source=handler.go:140 msg="Probe failed" module=tls_connect target=g679-srv:9442 duration_seconds=0.003030717
$ curl 'localhost:9115/probe?module=tls_sin_verificar&target=g679-srv:9442'  (insecure_skip_verify: true)
  probe_ssl_earliest_cert_expiry 1.792942704e+09
  probe_success 1
  una sonda tls_connect sana expone 13 series: probe_dns_lookup_time_seconds, probe_duration_seconds, probe_failed_due_to_bytes, probe_failed_due_to_regex, probe_ip_addr_hash, probe_ip_protocol, probe_ssl_earliest_cert_expiry, probe_ssl_last_chain_expiry_timestamp_seconds, probe_ssl_last_chain_info, probe_success, probe_timeout_seconds, probe_tls_cipher_info, probe_tls_version_info
```

Con la cadena validada (`ca_file`), el certificado vencido da `probe_success=0` y NO hay serie `probe_ssl_earliest_cert_expiry`: una alerta «< 14 días» sobre esa métrica se queda callada justo cuando ya venció; hace falta además una alerta sobre probe_success. Con `insecure_skip_verify: true` sí expone la fecha del vencido.

## Reglas de alerta usadas

```yaml
groups:
  - name: certificados
    rules:
      - alert: CertificadoPorVencer
        expr: (certexp_not_after_timestamp_seconds - time()) / 86400 < 14
        for: 30s
        labels: {severity: aviso}
        annotations:
          summary: "{{ $labels.objetivo }} vence en {{ $value | printf \"%.1f\" }} días"
      - alert: SondaTLSFalla
        expr: certexp_probe_success == 0
        for: 30s
        labels: {severity: aviso}
        annotations:
          summary: "no pude leer el certificado de {{ $labels.objetivo }}"
```
