# -*- coding: utf-8 -*-
"""NATS en Go: cuándo sustituye a Kafka — segunda parte del banco (golandia, vídeo 385). Escribe SALIDA-extra.md.

Reentrega y límites de JetStream (proyecto/reentrega) y releer en Kafka con otro grupo (proyecto/relee).
Levanta su propio NATS y Kafka, igual que correr.py; SALIDA.md (primera parte) no se toca.

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
    with BancoRed(AQUI, "NATS frente a Kafka — reentrega y releer", salida="SALIDA-extra.md") as b:
        d("run", "-d", "--name", "fab-nats", "--network", RED, "nats:2.15", "-js")
        listo("fab-nats", "Server is ready")
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
        listo("fab-kafka", "Kafka Server started")
        K = "/opt/kafka/bin"
        d("exec", "fab-kafka", f"{K}/kafka-topics.sh", "--bootstrap-server", "localhost:9092", "--create",
          "--topic", "facturas", "--partitions", "3", "--replication-factor", "1")
        d("exec", "fab-kafka", f"{K}/kafka-topics.sh", "--bootstrap-server", "localhost:9092", "--create",
          "--topic", "lecturas", "--partitions", "1", "--replication-factor", "1")

        b.titulo("7 · JetStream: el mensaje que nadie confirma, y el stream con límite")
        b.sh("go run ./reentrega", mostrar="go run ./reentrega")

        b.titulo("7b · NATS básico: el suscriptor que no da abasto")
        b.sh("go run ./lento", mostrar="go run ./lento")

        b.titulo("7c · JetStream: el mismo mensaje dos veces, con Nats-Msg-Id")
        b.sh("go run ./dedup", mostrar="go run ./dedup")

        b.titulo("7d · El consumidor que se queda atrás: Kafka y JetStream")
        b.sh("go run ./atras", mostrar="go run ./atras")
        b.anotar("kafka-consumer-groups.sh --describe --group lento",
                 "\n".join(" ".join(l.split()) for l in d("exec", "fab-kafka", f"{K}/kafka-consumer-groups.sh",
                           "--bootstrap-server", "localhost:9092", "--describe", "--group", "lento").splitlines()
                           if l.strip() and not l.startswith("Consumer group")))

        b.titulo("8 · Kafka: leer no borra")
        b.sh("go run ./relee", mostrar="go run ./relee")
finally:
    d("rm", "-f", "fab-nats", "fab-kafka")
    d("network", "rm", RED)
