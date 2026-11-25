# El servidor de la revancha (recursos/pruebas-go-java-revancha/go/main.go) en Python, con FastAPI sobre uvicorn.
# El mismo contrato: /suma (CPU), /trabajo (objetos, dict, orden y JSON armado a mano), /espera.
# Para «¿Go o Python?» (2026-10-24): la consulta que más se busca es «golang vs python speed / performance» y «golang vs fastapi».
import asyncio
import os
import sys
import time

ARRANQUE = time.perf_counter()

from fastapi import FastAPI, Response  # noqa: E402

app = FastAPI()
JSON = "application/json"


@app.get("/suma")
def suma(n: int = 0):
    t0 = time.perf_counter_ns()
    s = 0
    for i in range(1, n + 1):
        s += (i * i) % 1000003
    return Response(f'{{"lang":"python","n":{n},"suma":{s},"ns":{time.perf_counter_ns() - t0}}}', media_type=JSON)


class Item:
    __slots__ = ("id", "nombre", "valor")

    def __init__(self, i, nombre, valor):
        self.id, self.nombre, self.valor = i, nombre, valor


@app.get("/trabajo")
def trabajo(n: int = 0):
    t0 = time.perf_counter_ns()
    m = {}
    for i in range(n):
        it = Item(i, "item-" + str(i), (i * 7919) % 1000)
        m[it.nombre] = it
    lista = list(m.values())
    total = sum(it.valor for it in lista)
    lista.sort(key=lambda it: (it.valor, it.id))
    partes = [f'{{"id":{it.id},"nombre":"{it.nombre}","valor":{it.valor}}}' for it in lista[:50]]
    cuerpo = (f'{{"lang":"python","n":{n},"total":{total},"primeros":[' + ",".join(partes)
              + f'],"ns":{time.perf_counter_ns() - t0}}}')
    return Response(cuerpo, media_type=JSON)


@app.get("/espera")
async def espera(ms: int = 0):
    await asyncio.sleep(ms / 1000)
    return Response(f'{{"lang":"python","n":{ms},"suma":0,"ns":0}}', media_type=JSON)


@app.on_event("startup")
async def listo():
    print(f"LISTO_EN_MS {(time.perf_counter() - ARRANQUE) * 1000:.3f} python {sys.version.split()[0]} pid={os.getpid()}",
          file=sys.stderr, flush=True)
