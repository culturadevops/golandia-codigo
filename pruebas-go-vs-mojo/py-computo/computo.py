# El cómputo de «¿Go o Mojo?» (golandia 904) en Python 3.13, el lenguaje del que Mojo dice ser pariente: las mismas
# dos funciones del servidor de Go de la serie, con el mismo arnés que computo.mojo y go-computo/main.go.
#   python computo.py una | medir | (nada)
import sys
import time


class Item:
    __slots__ = ("id", "nombre", "valor")

    def __init__(self, id, nombre, valor):
        self.id, self.nombre, self.valor = id, nombre, valor


def suma(n):
    t0 = time.perf_counter_ns()
    s = 0
    for i in range(1, n + 1):
        s += i * i % 1000003
    return '{"lang":"python","n":%d,"suma":%d,"ns":%d}' % (n, s, time.perf_counter_ns() - t0)


def trabajo(n):
    t0 = time.perf_counter_ns()
    m = {}
    for i in range(n):
        it = Item(i, "item-" + str(i), (i * 7919) % 1000)
        m[it.nombre] = it
    lista = []
    total = 0
    for it in m.values():
        lista.append(it)
        total += it.valor
    lista.sort(key=lambda it: (it.valor, it.id))
    b = ['{"lang":"python","n":', str(n), ',"total":', str(total), ',"primeros":[']
    for i in range(min(50, len(lista))):
        if i > 0:
            b.append(",")
        it = lista[i]
        b.append('{"id":%d,"nombre":"%s","valor":%d}' % (it.id, it.nombre, it.valor))
    b.append('],"ns":%d}' % (time.perf_counter_ns() - t0))
    return "".join(b)


def medir(nombre, f, n):
    ns = []
    for i in range(350):
        t0 = time.perf_counter_ns()
        f(n)
        dt = time.perf_counter_ns() - t0
        if i >= 50:
            ns.append(str(dt))
    print(nombre, " ".join(ns))


if __name__ == "__main__":
    if sys.argv[1:] == ["una"]:
        print(suma(200000))
        print(trabajo(20000))
    elif sys.argv[1:] == ["medir"]:
        medir("SUMA", suma, 200000)
        medir("TRABAJO", trabajo, 20000)
