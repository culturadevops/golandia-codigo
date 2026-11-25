# El cómputo de «¿Go o Mojo?» (golandia 904): las MISMAS dos funciones de toda la serie «¿Go o X? Lo medí»
# (/suma y /trabajo del servidor de Go), sin servidor web, porque Mojo 1.2 estable no trae TCP en su librería
# estándar (ver SALIDA-demos.md). Mismo método que la sección 8 del banco: 50 de calentamiento y 300 medidas.
from std.time import perf_counter_ns
from std.collections import Dict
from std.sys import argv


@fieldwise_init
struct Item(Copyable, Movable):
    var id: Int
    var nombre: String
    var valor: Int


def suma(n: Int) -> String:
    var t0 = perf_counter_ns()
    var s: UInt64 = 0
    for i in range(1, n + 1):
        var u = UInt64(i)
        s += u * u % 1000003
    var ns = perf_counter_ns() - t0
    return String('{"lang":"mojo","n":', n, ',"suma":', s, ',"ns":', ns, "}")


def trabajo(n: Int) raises -> String:
    var t0 = perf_counter_ns()
    var m = Dict[String, Item](capacity=n)
    for i in range(n):
        var nombre = String("item-", i)
        m[nombre] = Item(i, nombre, (i * 7919) % 1000)
    var lista = List[Item](capacity=len(m))
    var total = 0
    for v in m.values():
        lista.append(v.copy())
        total += v.valor

    def menor(a: Item, b: Item) -> Bool:
        if a.valor != b.valor:
            return a.valor < b.valor
        return a.id < b.id

    sort(lista, menor)
    var b = String('{"lang":"mojo","n":', n, ',"total":', total, ',"primeros":[')
    for i in range(min(50, len(lista))):
        if i > 0:
            b += ","
        b += String('{"id":', lista[i].id, ',"nombre":"', lista[i].nombre, '","valor":', lista[i].valor, "}")
    b += String('],"ns":', perf_counter_ns() - t0, "}")
    return b


def main() raises:
    # computo una    → una llamada de cada función (para comprobar que da lo mismo que Go)
    # computo medir  → 50 de calentamiento y 300 medidas de cada una; imprime los 300 tiempos en ns
    # computo        → nada (para medir el arranque del proceso)
    var args = argv()
    if len(args) > 1 and args[1] == "una":
        print(suma(200000))
        print(trabajo(20000))
    elif len(args) > 1 and args[1] == "medir":
        var linea = String("SUMA")
        for i in range(350):
            var t0 = perf_counter_ns()
            var r = suma(200000)
            var dt = perf_counter_ns() - t0
            if i >= 50:
                linea += String(" ", dt)
            if r.byte_length() == 0:
                print(r)
        print(linea)
        linea = String("TRABAJO")
        for i in range(350):
            var t0 = perf_counter_ns()
            var r = trabajo(20000)
            var dt = perf_counter_ns() - t0
            if i >= 50:
                linea += String(" ", dt)
            if r.byte_length() == 0:
                print(r)
        print(linea)
