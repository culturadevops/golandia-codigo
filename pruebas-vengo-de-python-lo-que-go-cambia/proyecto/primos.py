# Cuenta primos hasta N, repartido en 8 trozos: en fila, con 8 hilos y con 8 procesos.
import sys, time
from concurrent.futures import ThreadPoolExecutor, ProcessPoolExecutor

N, TROZOS = 2_000_000, 8

def es_primo(n):
    if n < 2: return False
    i = 2
    while i * i <= n:
        if n % i == 0: return False
        i += 1
    return True

def contar(rango):
    a, b = rango
    return sum(1 for n in range(a, b) if es_primo(n))

def trozos():
    paso = N // TROZOS
    return [(i * paso, (i + 1) * paso) for i in range(TROZOS)]

def medir(nombre, f):
    t = time.perf_counter(); total = f(); s = time.perf_counter() - t
    print(f"{nombre:<12} {total} primos · {s:.2f} s")

if __name__ == "__main__":
    medir("en fila", lambda: sum(map(contar, trozos())))
    with ThreadPoolExecutor(TROZOS) as ex:
        medir("8 hilos", lambda: sum(ex.map(contar, trozos())))
    with ProcessPoolExecutor(TROZOS) as ex:
        medir("8 procesos", lambda: sum(ex.map(contar, trozos())))
