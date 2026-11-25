# -*- coding: utf-8 -*-
"""El contrato del banco «¿Go o X?» (canales/golandia/recursos/pruebas-go-vs-python/app/servidor.py), sin framework.

La MISMA lógica para todos los frameworks: cada servidor solo lee `n` a su manera y devuelve este texto tal cual, con
Content-Type application/json. Así la única diferencia entre ellos es el framework y su servidor.
  /suma?n=…     multiplicar y módulo en un bucle (CPU pura). Con n=1 lo que queda es el coste del framework.
  /trabajo?n=…  n objetos, a un dict, ordenarlos y armar el JSON a mano.
El campo "ns" es el reloj del propio cálculo (sin red ni framework) y es lo único que cambia entre respuestas.
"""
import sys
import time

ARRANQUE = time.perf_counter()  # se importa lo primero en cada servidor: el reloj de «listo» empieza aquí
LANG = "python"


def suma(n):
    t0 = time.perf_counter_ns()
    s = 0
    for i in range(1, n + 1):
        s += (i * i) % 1000003
    return f'{{"lang":"{LANG}","n":{n},"suma":{s},"ns":{time.perf_counter_ns() - t0}}}'


class Item:
    __slots__ = ("id", "nombre", "valor")

    def __init__(self, i, nombre, valor):
        self.id, self.nombre, self.valor = i, nombre, valor


def trabajo(n):
    t0 = time.perf_counter_ns()
    m = {}
    for i in range(n):
        it = Item(i, "item-" + str(i), (i * 7919) % 1000)
        m[it.nombre] = it
    lista = list(m.values())
    total = sum(it.valor for it in lista)
    lista.sort(key=lambda it: (it.valor, it.id))
    partes = [f'{{"id":{it.id},"nombre":"{it.nombre}","valor":{it.valor}}}' for it in lista[:50]]
    return (f'{{"lang":"{LANG}","n":{n},"total":{total},"primeros":[' + ",".join(partes)
            + f'],"ns":{time.perf_counter_ns() - t0}}}')


def listo(quien):
    """La marca que lee el banco (LISTO_EN_MS): desde que se empezó a importar hasta que la app puede atender."""
    import os
    print(f"LISTO_EN_MS {(time.perf_counter() - ARRANQUE) * 1000:.3f} {quien} python {sys.version.split()[0]} pid={os.getpid()}",
          file=sys.stderr, flush=True)
