# -*- coding: utf-8 -*-
"""Litestar servido con uvicorn. Rutas síncronas con sync_to_thread=True: lo que pide su documentación para código que bloquea
(van al threadpool, como `def` en FastAPI)."""
import logica  # noqa: I001

from litestar import Litestar, Response, get

JSON = "application/json"


@get("/suma", sync_to_thread=True)
def suma(n: int = 0) -> Response:
    return Response(logica.suma(n), media_type=JSON)


@get("/trabajo", sync_to_thread=True)
def trabajo(n: int = 0) -> Response:
    return Response(logica.trabajo(n), media_type=JSON)


app = Litestar(route_handlers=[suma, trabajo], on_startup=[lambda: logica.listo("litestar")])
