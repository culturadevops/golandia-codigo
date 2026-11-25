# -*- coding: utf-8 -*-
"""Demo del vídeo «Prometheus exporter en Go» (golandia 679).

    python correr.py

Todo corre en Docker, en la red g679, con contenedores g679-*; nunca toca otros. La configuración de Prometheus,
las reglas y la CA viven en un directorio temporal (tempfile.mkdtemp) que se archiva al final en salidas/config/.
Cada escenario deja su salida en salidas/<escenario>.txt; SALIDA.md se escribe a mano a partir de ellas.
"""
import json
import os
import shutil
import subprocess
import sys
import tempfile
import time
import urllib.parse
import urllib.request
from pathlib import Path

AQUI = Path(__file__).parent
SAL = AQUI / "salidas"
SAL.mkdir(exist_ok=True)
BIN = AQUI / "bin"
ENV = {**os.environ, "MSYS_NO_PATHCONV": "1"}
GO = "golang:1.27.2"
PROM = "prom/prometheus:v3.15.1"
AM = "prom/alertmanager:v0.34.0"
BB = "prom/blackbox-exporter:v0.29.0"
RED_ = "g679"
TMP = Path(tempfile.mkdtemp(prefix="g679-"))
NOMBRES = ["g679-srv", "g679-srv2", "g679-cx", "g679-cx-sin", "g679-cx-sec", "g679-cx-err", "g679-cx-par", "g679-cx-2", "g679-gauge", "g679-prom", "g679-am", "g679-bb"]
OBJ = "g679-srv:9441,g679-srv:9442,g679-srv:9443"
LENTOS = "g679-srv:9444,g679-srv2:9444"


def w(p):
    return subprocess.run(["cygpath", "-w", str(p)], capture_output=True, text=True).stdout.strip() or str(p)


def sh(*a, check=True):
    r = subprocess.run(list(a), capture_output=True, text=True, env=ENV, encoding="utf-8", errors="replace")
    if check and r.returncode != 0:
        raise SystemExit(f"falló {' '.join(a)}:\n{r.stdout}{r.stderr}")
    return r.stdout.strip()


def limpiar():
    for n in NOMBRES:
        sh("docker", "rm", "-f", n, check=False)


def get(url, timeout=30):
    t = time.monotonic()
    try:
        with urllib.request.urlopen(url, timeout=timeout) as r:
            return r.read().decode(), time.monotonic() - t, None
    except Exception as e:  # noqa: BLE001
        return "", time.monotonic() - t, e


def q(expr):
    txt, _, err = get("http://127.0.0.1:29679/api/v1/query?" + urllib.parse.urlencode({"query": expr}))
    if err:
        raise SystemExit(f"query {expr}: {err}")
    return json.loads(txt)["data"]["result"]


def metricas(txt, prefijo="certexp_"):
    return [l for l in txt.splitlines() if l.startswith(prefijo)]


def compilar():
    BIN.mkdir(exist_ok=True)
    s = sh("docker", "run", "--rm", "-v", f"{w(AQUI)}:/trabajo", "-v", "g679-gomod:/go/pkg/mod", "-e", "CGO_ENABLED=0", GO, "sh", "-c",
           "cd /trabajo/exporter && go vet ./... && go build -trimpath -ldflags='-s -w' -o /trabajo/bin/ ./cmd/... && "
           "cd /trabajo/servidores && go build -trimpath -o /trabajo/bin/servidores . && go version && "
           "cd /trabajo/exporter && go list -m github.com/prometheus/client_golang && ls -l /trabajo/bin")
    return s


def correr_bin(nombre, *args, puerto=None):
    a = ["docker", "run", "-d", "--name", nombre, "--network", RED_, "-v", f"{w(BIN)}:/b:ro", "-v", f"{w(TMP)}:/compartido"]
    if puerto:
        a += ["-p", f"127.0.0.1:{puerto}:9679"]
    sh(*a, GO, *args)


def esperar(url, s=30):
    fin = time.monotonic() + s
    while time.monotonic() < fin:
        if get(url, 2)[2] is None:
            return
        time.sleep(0.5)
    raise SystemExit(f"no responde {url}")


def escribir(nombre, lineas):
    (SAL / f"{nombre}.txt").write_text("\n".join(lineas) + "\n", encoding="utf-8")
    print(f"--- {nombre}\n" + "\n".join(lineas))


def configs():
    (TMP / "prometheus.yml").write_text(f"""global:
  scrape_interval: 15s
  scrape_timeout: 10s
  evaluation_interval: 15s
rule_files: [/compartido/reglas.yml]
alerting:
  alertmanagers: [{{static_configs: [{{targets: ["g679-am:9093"]}}]}}]
scrape_configs:
  - job_name: certexp
    static_configs: [{{targets: ["g679-cx:9679"]}}]
  - job_name: certexp-sin-timeout
    static_configs: [{{targets: ["g679-cx-sin:9679"]}}]
  - job_name: certexp-etiqueta-error
    scrape_interval: 5s
    scrape_timeout: 4s
    static_configs: [{{targets: ["g679-cx-err:9679"]}}]
  - job_name: blackbox
    metrics_path: /probe
    params: {{module: [tls_connect]}}
    static_configs: [{{targets: ["g679-srv:9441", "g679-srv:9442", "g679-srv:9443", "g679-srv:9444"]}}]
    relabel_configs:
      - {{source_labels: [__address__], target_label: __param_target}}
      - {{source_labels: [__param_target], target_label: instance}}
      - {{target_label: __address__, replacement: "g679-bb:9115"}}
""", encoding="utf-8")
    (TMP / "reglas.yml").write_text("""groups:
  - name: certificados
    rules:
      - alert: CertificadoPorVencer
        expr: (certexp_not_after_timestamp_seconds - time()) / 86400 < 14
        for: 30s
        labels: {severity: aviso}
        annotations:
          summary: "{{ $labels.objetivo }} vence en {{ $value | printf \\"%.1f\\" }} días"
      - alert: SondaTLSFalla
        expr: certexp_probe_success == 0
        for: 30s
        labels: {severity: aviso}
        annotations:
          summary: "no pude leer el certificado de {{ $labels.objetivo }}"
""", encoding="utf-8")
    (TMP / "alertmanager.yml").write_text("""route:
  receiver: nadie
  group_wait: 5s
receivers:
  - name: nadie
""", encoding="utf-8")
    (TMP / "blackbox.yml").write_text("""modules:
  tls_connect:
    prober: tcp
    timeout: 2s
    tcp:
      tls: true
      tls_config:
        ca_file: /compartido/ca.pem
  tls_sin_verificar:
    prober: tcp
    timeout: 2s
    tcp:
      tls: true
      tls_config:
        insecure_skip_verify: true
""", encoding="utf-8")


def main():
    limpiar()
    sh("docker", "network", "create", RED_, check=False)
    escribir("compilar", ["$ go build -trimpath -ldflags='-s -w' ./cmd/...", *compilar().splitlines()])
    configs()
    sh("docker", "run", "--rm", "-v", f"{w(TMP)}:/compartido", "--entrypoint", "promtool", PROM, "check", "rules", "/compartido/reglas.yml")

    # Servidores: srv vive toda la demo; srv2 es el que «se cae» en el escenario del Gauge global.
    correr_bin("g679-srv", "/b/servidores", "-nombre", "g679-srv", "-ca", "/compartido/ca.pem")
    correr_bin("g679-srv2", "/b/servidores", "-nombre", "g679-srv2", "-ca", "/compartido/ca2.pem")
    time.sleep(2)
    escribir("servidores", ["$ docker logs g679-srv", *sh("docker", "logs", "g679-srv").splitlines(),
                            *subprocess.run(["docker", "logs", "g679-srv"], capture_output=True, text=True).stderr.strip().splitlines()])

    # ── 1 · /metrics del Collector ──
    correr_bin("g679-cx", "/b/certexp", "-objetivos", OBJ + ",g679-srv:9444", "-timeout", "2s", puerto=29680)
    esperar("http://127.0.0.1:29680/metrics")
    txt, dur, _ = get("http://127.0.0.1:29680/metrics")
    escribir("metrics", [f"$ curl -s localhost:9679/metrics | grep ^certexp   (tardó {dur:.2f} s)", *sorted(metricas(txt))])

    # ── 2 · Timeouts: sin límite, secuencial y en paralelo, con dos hosts lentos ──
    correr_bin("g679-cx-sin", "/b/certexp", "-objetivos", OBJ + "," + LENTOS, "-timeout", "0", puerto=29684)
    correr_bin("g679-cx-sec", "/b/certexp", "-objetivos", OBJ + "," + LENTOS, "-timeout", "2s", "-secuencial", puerto=29686)
    correr_bin("g679-cx-par", "/b/certexp", "-objetivos", OBJ + "," + LENTOS, "-timeout", "2s", puerto=29687)
    time.sleep(2)
    lin = [f"objetivos: 3 sanos/por vencer + 2 lentos ({LENTOS})"]
    for nombre, p in (("-timeout 0 (sin límite)", 29684), ("-timeout 2s -secuencial", 29686), ("-timeout 2s en paralelo", 29687)):
        txt, dur, err = get(f"http://127.0.0.1:{p}/metrics", timeout=15)
        ok = sum(1 for l in metricas(txt) if l.startswith("certexp_probe_success") and l.endswith(" 1"))
        lin.append(f"{nombre:26s} scrape {dur:5.2f} s · " + (f"sin respuesta ({type(err).__name__}: {err})" if err else f"{ok} sondas bien, {len(metricas(txt))} líneas"))
    escribir("timeouts", lin)

    # ── 3 · Gauge global frente a Collector cuando un host se cae ──
    correr_bin("g679-gauge", "/b/gauge-global", "-objetivos", "g679-srv2:9441,g679-srv2:9443", "-cada", "60s", puerto=29682)
    correr_bin("g679-cx-2", "/b/certexp", "-objetivos", "g679-srv2:9441,g679-srv2:9443", puerto=29688)
    esperar("http://127.0.0.1:29682/metrics")
    esperar("http://127.0.0.1:29688/metrics")
    time.sleep(1)
    lin = ["== con g679-srv2 vivo"]
    for nombre, p in (("gauge-global", 29682), ("certexp (Collector)", 29688)):
        lin += [f"[{nombre}]", *[f"  {l}" for l in sorted(metricas(get(f"http://127.0.0.1:{p}/metrics")[0]))]]
    sh("docker", "stop", "-t", "1", "g679-srv2")
    t0 = time.monotonic()
    time.sleep(5)
    lin.append(f"== docker stop g679-srv2 · {time.monotonic() - t0:.1f} s después")
    for nombre, p in (("gauge-global", 29682), ("certexp (Collector)", 29688)):
        lin += [f"[{nombre}]", *[f"  {l}" for l in sorted(metricas(get(f"http://127.0.0.1:{p}/metrics")[0]))]]
    total = len(get("http://127.0.0.1:29682/metrics")[0].splitlines())
    lin.append(f"(gauge-global con promhttp.Handler() expone además go_* y process_*: {total} líneas en total)")
    escribir("gauge-vs-collector", lin)

    # ── 4 · Prometheus + Alertmanager + blackbox ──
    correr_bin("g679-cx-err", "/b/certexp", "-objetivos", "g679-srv:9443,g679-srv:9444,g679-srv:9999", "-etiqueta-error")
    sh("docker", "run", "-d", "--name", "g679-am", "--network", RED_, "-p", "127.0.0.1:29681:9093", "-v", f"{w(TMP)}:/compartido",
       AM, "--config.file=/compartido/alertmanager.yml")
    sh("docker", "run", "-d", "--name", "g679-bb", "--network", RED_, "-p", "127.0.0.1:29683:9115", "-v", f"{w(TMP)}:/compartido",
       BB, "--config.file=/compartido/blackbox.yml")
    sh("docker", "run", "-d", "--name", "g679-prom", "--network", RED_, "-p", "127.0.0.1:29679:9090", "-v", f"{w(TMP)}:/compartido",
       PROM, "--config.file=/compartido/prometheus.yml")
    t0 = time.monotonic()
    esperar("http://127.0.0.1:29679/-/ready", 60)
    esperar("http://127.0.0.1:29681/-/ready", 60)

    # Esperar a que disparen las alertas en Alertmanager (for: 30s + evaluación cada 15 s).
    vistas = {}
    while time.monotonic() - t0 < 150:
        txt, _, err = get("http://127.0.0.1:29681/api/v2/alerts")
        if not err:
            for a in json.loads(txt):
                k = (a["labels"]["alertname"], a["labels"].get("objetivo", ""), a["labels"].get("job", ""))
                vistas.setdefault(k, (time.monotonic() - t0, a["annotations"].get("summary", "")))
        if len(vistas) >= 3 and time.monotonic() - t0 > 75:
            break
        time.sleep(1)
    lin = [f"Prometheus {PROM.split(':')[1]} · Alertmanager {AM.split(':')[1]} · scrape cada 15 s · for: 30s",
           "segundos desde que arrancó Prometheus hasta que la alerta llegó a Alertmanager:"]
    for (nom, obj, job), (t, s) in sorted(vistas.items(), key=lambda x: x[1][0]):
        lin.append(f"  {t:5.1f} s  {nom:22s} {job:24s} {s}")
    escribir("alertas", lin)

    lin = ["$ (certexp_not_after_timestamp_seconds - time()) / 86400"]
    for r in sorted(q('(certexp_not_after_timestamp_seconds{job="certexp"} - time()) / 86400'), key=lambda r: r["metric"]["objetivo"]):
        lin.append(f"  {{objetivo=\"{r['metric']['objetivo']}\"}}  {float(r['value'][1]):.2f}")
    lin.append("$ certexp_probe_success{job=\"certexp\"}")
    for r in sorted(q('certexp_probe_success{job="certexp"}'), key=lambda r: r["metric"]["objetivo"]):
        lin.append(f"  {{objetivo=\"{r['metric']['objetivo']}\"}}  {r['value'][1]}")
    lin.append("$ certexp_probe_duration_seconds{job=\"certexp\"}")
    for r in sorted(q('certexp_probe_duration_seconds{job="certexp"}'), key=lambda r: r["metric"]["objetivo"]):
        lin.append(f"  {{objetivo=\"{r['metric']['objetivo']}\"}}  {float(r['value'][1]):.3f}")
    lin.append("$ up · scrape_duration_seconds")
    for r in sorted(q('up{job=~"certexp|certexp-sin-timeout"}'), key=lambda r: r["metric"]["job"]):
        d = q(f'scrape_duration_seconds{{job="{r["metric"]["job"]}"}}')
        lin.append(f"  {r['metric']['job']:22s} up={r['value'][1]}  scrape_duration_seconds={float(d[0]['value'][1]):.2f}")
    tg = json.loads(get("http://127.0.0.1:29679/api/v1/targets")[0])["data"]["activeTargets"]
    for t in tg:
        if t["labels"]["job"] == "certexp-sin-timeout":
            lin.append(f"  lastError de certexp-sin-timeout: {t['lastError']}")
    escribir("prometheus", lin)

    # Cardinalidad: el texto del error en una etiqueta. Scrape cada 5 s; contamos series distintas.
    time.sleep(max(0, 65 - (time.monotonic() - t0)))
    lin = ["job certexp-etiqueta-error: objetivos g679-srv:9443 (sano), :9444 (lento), :9999 (cerrado) · scrape cada 5 s"]
    vals = q('certexp_probe_error')
    for r in vals:
        lin.append(f"  ahora: {{objetivo=\"{r['metric']['objetivo']}\", error=\"{r['metric']['error']}\"}}")
    for ventana in ("1m",):
        n = q(f'count(count_over_time(certexp_probe_error[{ventana}]))')
        n2 = q(f'count(count_over_time(certexp_probe_success{{job="certexp-etiqueta-error"}}[{ventana}]))')
        lin.append(f"  series distintas en {ventana}: certexp_probe_error={n[0]['value'][1] if n else 0} · certexp_probe_success={n2[0]['value'][1] if n2 else 0}")
    escribir("cardinalidad", lin)

    # blackbox_exporter: lo mismo con lo oficial
    lin = [f"blackbox_exporter {BB.split(':')[1]} · módulo tls_connect (tcp + tls, ca_file con nuestra CA)"]
    for r in sorted(q('probe_success{job="blackbox"}'), key=lambda r: r["metric"]["instance"]):
        inst = r["metric"]["instance"]
        e = q(f'(probe_ssl_earliest_cert_expiry{{instance="{inst}"}} - time()) / 86400')
        d = q(f'probe_duration_seconds{{instance="{inst}"}}')
        lin.append(f"  {inst:16s} probe_success={r['value'][1]}  días={'%.2f' % float(e[0]['value'][1]) if e else 'sin serie'}"
                   f"  probe_duration_seconds={float(d[0]['value'][1]):.3f}")
    txt, _, _ = get("http://127.0.0.1:29683/probe?module=tls_connect&target=g679-srv:9442&debug=true")
    lin.append("$ curl 'localhost:9115/probe?module=tls_connect&target=g679-srv:9442&debug=true' | grep -i -E 'error|expir'")
    lin += [f"  {l.strip()}" for l in txt.splitlines() if ("error" in l.lower() or "expir" in l.lower())][:6]
    txt, _, _ = get("http://127.0.0.1:29683/probe?module=tls_sin_verificar&target=g679-srv:9442")
    lin.append("$ curl 'localhost:9115/probe?module=tls_sin_verificar&target=g679-srv:9442'  (insecure_skip_verify: true)")
    lin += [f"  {l}" for l in txt.splitlines() if l.startswith(("probe_success", "probe_ssl_earliest"))]
    txt, _, _ = get("http://127.0.0.1:29683/probe?module=tls_connect&target=g679-srv:9441")
    lin.append(f"  una sonda tls_connect sana expone {sum(1 for l in txt.splitlines() if l and not l.startswith('#'))} series: "
               + ", ".join(sorted({l.split('{')[0].split(' ')[0] for l in txt.splitlines() if l and not l.startswith('#')})))
    escribir("blackbox", lin)

    # Archivar la configuración usada (no se borra el temporal hasta copiarlo).
    dst = SAL / "config"
    dst.mkdir(exist_ok=True)
    for f in ("prometheus.yml", "reglas.yml", "alertmanager.yml", "blackbox.yml"):
        shutil.copy(TMP / f, dst / f)
    escribir("versiones", [f"Docker {sh('docker', 'version', '--format', '{{.Server.Version}}')}", PROM, AM, BB, GO])

    limpiar()
    sh("docker", "network", "rm", RED_, check=False)
    shutil.rmtree(TMP, ignore_errors=True)


if __name__ == "__main__":
    try:
        main()
    except BaseException:
        limpiar()
        raise
