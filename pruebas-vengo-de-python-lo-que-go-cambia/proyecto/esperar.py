# Esperar a la red: 50 peticiones a servidores que tardan 50-300 ms. En fila y con 50 hilos.
import time, urllib.request
from concurrent.futures import ThreadPoolExecutor

URLS = [f"http://127.0.0.1:{p}/salud" for p in range(9001, 9051)]

def pedir(u):
    with urllib.request.urlopen(u, timeout=2) as r:
        return r.status == 200

def medir(nombre, f):
    t = time.perf_counter(); ok = f(); s = time.perf_counter() - t
    print(f"{nombre:<10} {ok} bien · {s:.2f} s")

medir("en fila", lambda: sum(map(pedir, URLS)))
with ThreadPoolExecutor(50) as ex:
    medir("50 hilos", lambda: sum(ex.map(pedir, URLS)))
