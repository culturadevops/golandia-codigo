# -*- coding: utf-8 -*-
"""Pyronaut 0.1.2 (Python sobre GraalPy + Micronaut/Netty). La MISMA lógica del banco 848 (logica.py, copiada tal cual):
cada ruta solo lee `n` y devuelve el texto de logica.py con Content-Type application/json.
Rutas síncronas (`def`): Pyronaut las manda al ejecutor «io» de Micronaut (docs 6.5.1), como FastAPI a su threadpool."""
from banco import logica  # noqa: I001

from micronaut.http.annotation import Get

logica.listo("pyronaut")


@Get(value="/suma", produces="application/json")
def suma(n: int = 0) -> str:
    return logica.suma(n)


@Get(value="/trabajo", produces="application/json")
def trabajo(n: int = 0) -> str:
    return logica.trabajo(n)
