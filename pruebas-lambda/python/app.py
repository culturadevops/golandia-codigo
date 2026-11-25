import time


def handler(evento, contexto):
    n = int(evento.get("n", 0))
    io_ms = int(evento.get("io_ms", 0))
    suma = 0
    for i in range(1, n + 1):
        suma += i * i % 1000003
    if io_ms > 0:
        time.sleep(io_ms / 1000.0)
    return {"lang": "python", "n": n, "suma": suma, "io_ms": io_ms}
