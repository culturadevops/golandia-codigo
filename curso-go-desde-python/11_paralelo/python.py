import time
from concurrent.futures import ThreadPoolExecutor
N, LAT = 30, 0.2
def chequear(i): time.sleep(LAT); return i % 7 != 0
t = time.perf_counter()
for i in range(N): chequear(i)
sec = time.perf_counter() - t
t = time.perf_counter()
with ThreadPoolExecutor(max_workers=N) as ex: list(ex.map(chequear, range(N)))
par = time.perf_counter() - t
print("== en Python ==")
print(f"  uno detras de otro        {sec*1000:6.0f} ms")
print(f"  hilos, todos a la vez     {par*1000:6.0f} ms")
print("  Para ESPERAR (red, disco), los hilos de Python tambien sirven.")
