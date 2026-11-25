# Un error de tipeo en una rama que casi nunca se ejecuta.
import sys

def avisar(disco_lleno):
    if disco_lleno:
        enviar_alrta("disco lleno")   # ← la errata
    return "todo bien"

print(avisar(False))
print(avisar(True))
