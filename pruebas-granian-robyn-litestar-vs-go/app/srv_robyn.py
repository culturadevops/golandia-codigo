# -*- coding: utf-8 -*-
"""Robyn: el framework con el runtime en Rust. La MISMA lógica (logica.py) que FastAPI, Litestar y Go.
Rutas `def` (síncronas), como `def` en FastAPI y `sync_to_thread=True` en Litestar: el cálculo es CPU y bloquea.
Procesos y workers se pasan por la línea de órdenes (--processes / --workers), ver arrancar.sh."""
import logica  # noqa: I001

from robyn import Response, Robyn

JSON = {"Content-Type": "application/json"}
app = Robyn(__file__)


def n_de(request):
    return int(request.query_params.get("n", "0") or 0)


@app.get("/suma")
def suma(request):
    return Response(status_code=200, headers=JSON, description=logica.suma(n_de(request)))


@app.get("/trabajo")
def trabajo(request):
    return Response(status_code=200, headers=JSON, description=logica.trabajo(n_de(request)))


@app.startup_handler
def _listo():
    logica.listo("robyn")


if __name__ == "__main__":
    app.start(host="0.0.0.0", port=8080)
