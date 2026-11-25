# -*- coding: utf-8 -*-
"""NATS en Go: cuándo sustituye a Kafka — banco ejecutable (golandia, vídeo 385). Escribe SALIDA.md.

Levanta en una red de Docker un servidor NATS (nats:2.15, con JetStream) y un Kafka de un nodo
(apache/kafka:4.1.1, KRaft, configuración por defecto de la imagen), y mide lo que se puede medir en
esta máquina: tamaño de imagen, tiempo hasta estar listo y memoria en reposo. Después ejecuta los
programas de proyecto/ (nats.go) y un `nats bench` desde nats-box.

    python correr.py
"""
import subprocess
import sys
import time
from pathlib import Path

AQUI = Path(__file__).resolve().parent
sys.path.insert(0, str(AQUI.parent))
import banco_go  # noqa: E402

RED = "fab-nats-red"


def d(*args, check=False):
    r = subprocess.run(["docker", *args], capture_output=True, text=True, encoding="utf-8", errors="replace")
    return (r.stdout + r.stderr).strip()


class BancoRed(banco_go.Banco):
    def sh(self, orden, mostrar=None, guardar=True, env=None, timeout=1800):
        cmd = ["docker", "run", "--rm", "--network", RED, "-e", "NATS_URL=nats://fab-nats:4222", "-e", "KAFKA=fab-kafka:9092",
               "-v", f"{self.proyecto}:/w", "-w", "/w",
               "-v", "fab-go-mod:/go/pkg/mod", "-v", "fab-go-cache:/root/.cache/go-build",
               banco_go.IMAGEN, "bash", "-c", orden]
        r = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8", errors="replace", timeout=timeout)
        txt = (r.stdout + r.stderr).strip()
        if guardar:
            self.anotar(mostrar or orden, txt)
        return txt


def listo(contenedor, texto, limite=120):
    t0 = time.time()
    while time.time() - t0 < limite:
        if texto in d("logs", contenedor):
            return time.time() - t0
        time.sleep(0.2)
    return None


def memoria(contenedor):
    return d("stats", "--no-stream", "--format", "{{.MemUsage}}", contenedor).split(" / ")[0]


d("rm", "-f", "fab-nats", "fab-kafka")
d("network", "rm", RED)
d("network", "create", RED)
try:
    with BancoRed(AQUI, "NATS frente a Kafka") as b:
        b.titulo("Las imágenes")
        b.anotar("docker images", "\n".join(
            l for l in d("images", "--format", "{{.Repository}}:{{.Tag}}  {{.Size}}").splitlines()
            if l.startswith(("nats:2.15", "apache/kafka:4.1.1"))))

        b.titulo("Arrancar NATS (con JetStream)")
        d("run", "-d", "--name", "fab-nats", "--network", RED, "nats:2.15", "-js")
        t = listo("fab-nats", "Server is ready")
        b.anotar("docker run -d --name fab-nats nats:2.15 -js",
                 f"listo en {t:.1f} s (desde que el contenedor existe)" if t else "no arrancó")
        b.anotar("docker logs fab-nats | grep -E 'Version|JetStream|ready'",
                 "\n".join(l.split("] ", 1)[-1] for l in d("logs", "fab-nats").splitlines()
                           if any(k in l for k in ("Version:", "Starting JetStream", "Server is ready"))))

        b.titulo("Arrancar Kafka (KRaft, un nodo)")
        # Un nodo KRaft igual que el de la imagen, pero anunciándose como fab-kafka:9092 para que los
        # clientes de otros contenedores lleguen. (Cualquier KAFKA_* sustituye la config por defecto.)
        kenv = {"KAFKA_NODE_ID": "1", "KAFKA_PROCESS_ROLES": "broker,controller",
                "KAFKA_LISTENERS": "PLAINTEXT://:9092,CONTROLLER://:9093",
                "KAFKA_ADVERTISED_LISTENERS": "PLAINTEXT://fab-kafka:9092",
                "KAFKA_CONTROLLER_LISTENER_NAMES": "CONTROLLER",
                "KAFKA_LISTENER_SECURITY_PROTOCOL_MAP": "CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT",
                "KAFKA_CONTROLLER_QUORUM_VOTERS": "1@localhost:9093",
                "KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR": "1", "KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR": "1",
                "KAFKA_TRANSACTION_STATE_LOG_MIN_ISR": "1", "KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS": "0"}
        args = []
        for k, v in kenv.items():
            args += ["-e", f"{k}={v}"]
        d("run", "-d", "--name", "fab-kafka", "--network", RED, *args, "apache/kafka:4.1.1")
        t = listo("fab-kafka", "Kafka Server started")
        b.anotar("docker run -d --name fab-kafka apache/kafka:4.1.1",
                 f"listo en {t:.1f} s (desde que el contenedor existe)" if t else "no arrancó en 120 s")

        b.titulo("Memoria en reposo, 20 s después")
        time.sleep(20)
        b.anotar("docker stats --no-stream fab-nats fab-kafka",
                 f"fab-nats   {memoria('fab-nats')}\nfab-kafka  {memoria('fab-kafka')}")

        b.titulo("1 · Publicar y suscribirse (NATS básico)")
        b.sh("wc -l pubsub/main.go | awk '{print $1, \"líneas\"}' && go run ./pubsub", mostrar="go run ./pubsub")

        b.titulo("2 · JetStream: guardar y leer después")
        b.sh("go run ./jetstream", mostrar="go run ./jetstream")
        b.sh("go run ./jetstream 2>&1 | tail -4", mostrar="go run ./jetstream   # otra vez: el consumidor ya leyó los 3 primeros")

        b.titulo("3 · Petición y respuesta")
        b.sh("go run ./peticion", mostrar="go run ./peticion")

        b.titulo("3b · Un grupo de cola: repartir el trabajo")
        b.sh("go run ./cola", mostrar="go run ./cola")

        b.titulo("3c · El almacén clave-valor de JetStream")
        b.sh("go run ./kv", mostrar="go run ./kv")

        b.titulo("4 · nats bench: publicar 1 millón de mensajes de 128 bytes")
        r = subprocess.run(["docker", "run", "--rm", "--network", RED, "natsio/nats-box:latest", "nats", "bench",
                            "pub", "bench.prueba", "--msgs", "1000000", "--size", "128", "--server", "nats://fab-nats:4222",
                            "--no-progress"], capture_output=True, text=True, encoding="utf-8", errors="replace")
        b.anotar("nats bench pub bench.prueba --msgs 1000000 --size 128",
                 "\n".join(l for l in (r.stdout + r.stderr).splitlines() if l.strip()))

        b.titulo("4b · nats bench con JetStream: lo mismo, pero guardado en disco")
        r = subprocess.run(["docker", "run", "--rm", "--network", RED, "natsio/nats-box:latest", "nats", "bench",
                            "js", "pub", "async", "bench.js", "--create", "--storage", "file", "--msgs", "1000000",
                            "--size", "128", "--server", "nats://fab-nats:4222", "--no-progress"],
                           capture_output=True, text=True, encoding="utf-8", errors="replace")
        b.anotar("nats bench js pub async bench.js --create --storage file --msgs 1000000 --size 128",
                 "\n".join(l for l in (r.stdout + r.stderr).splitlines() if "stats:" in l) or (r.stdout + r.stderr))

        b.titulo("5 · Kafka: un tema con 3 particiones")
        K = "/opt/kafka/bin"
        b.anotar("kafka-topics.sh --create --topic pedidos --partitions 3",
                 d("exec", "fab-kafka", f"{K}/kafka-topics.sh", "--bootstrap-server", "localhost:9092",
                   "--create", "--topic", "pedidos", "--partitions", "3", "--replication-factor", "1"))
        b.anotar("kafka-topics.sh --describe --topic pedidos",
                 "\n".join(l.strip() for l in d("exec", "fab-kafka", f"{K}/kafka-topics.sh", "--bootstrap-server",
                           "localhost:9092", "--describe", "--topic", "pedidos").splitlines()))

        b.titulo("5b · Kafka desde Go (franz-go): la misma clave, la misma partición")
        b.sh("wc -l kafka/main.go | awk '{print $1, \"líneas\"}' && go run ./kafka", mostrar="go run ./kafka")

        b.titulo("6 · kafka-producer-perf-test: 1 millón de mensajes de 128 bytes")
        d("exec", "fab-kafka", f"{K}/kafka-topics.sh", "--bootstrap-server", "localhost:9092", "--create",
          "--topic", "bench", "--partitions", "1", "--replication-factor", "1")
        out = d("exec", "fab-kafka", f"{K}/kafka-producer-perf-test.sh", "--topic", "bench", "--num-records",
                "1000000", "--record-size", "128", "--throughput", "-1", "--producer-props",
                "bootstrap.servers=localhost:9092", "acks=1")
        b.anotar("kafka-producer-perf-test.sh --topic bench --num-records 1000000 --record-size 128 --throughput -1",
                 ([l for l in out.splitlines() if "records sent" in l] or ["(sin salida)"])[-1])

        # Sin «memoria después de todo»: el docker exec del perf-test arranca una JVM DENTRO del
        # contenedor de Kafka y la cuenta en su memoria. La comparación limpia es la de reposo.
finally:
    d("rm", "-f", "fab-nats", "fab-kafka")
    d("network", "rm", RED)
