# -*- coding: utf-8 -*-
"""FastAPI servido con uvicorn (uvloop + httptools). Dos formas de escribir la ruta, elegidas con MODO:
  MODO=def    (por defecto) `def`: FastAPI la manda a un hilo del threadpool, lo que pide su documentación para código que bloquea.
  MODO=async  `async def`: corre en el bucle de eventos; un cálculo largo lo bloquea mientras dura.
Es el mismo servidor del banco «¿Python o X?» (canales/golandia/recursos/pruebas-go-vs-python/app/servidor.py) con la lógica
sacada a logica.py."""
import logica  # noqa: I001

import os
from contextlib import asynccontextmanager

from fastapi import FastAPI, Response

JSON = "application/json"
MODO = os.environ.get("MODO", "def")


@asynccontextmanager
async def vida(_app):
    logica.listo(f"fastapi-{MODO}")
    yield


app = FastAPI(lifespan=vida)

if MODO == "async":
    @app.get("/suma")
    async def suma(n: int = 0):
        return Response(logica.suma(n), media_type=JSON)

    @app.get("/trabajo")
    async def trabajo(n: int = 0):
        return Response(logica.trabajo(n), media_type=JSON)
else:
    @app.get("/suma")
    def suma(n: int = 0):
        return Response(logica.suma(n), media_type=JSON)

    @app.get("/trabajo")
    def trabajo(n: int = 0):
        return Response(logica.trabajo(n), media_type=JSON)
