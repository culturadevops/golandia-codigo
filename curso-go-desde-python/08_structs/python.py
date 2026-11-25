class Servidor:
    def __init__(self, nombre, replicas):
        self.nombre, self.replicas = nombre, replicas
    def escalar(self, n):
        self.replicas = n
s = Servidor("api", 1); s.escalar(5)
a = Servidor("web", 2); b = a; b.replicas = 9
print("== en Python ==")
print(f"  tras s.escalar(5) -> replicas = {s.replicas}   (self es el objeto de verdad)")
print(f"  a.replicas={a.replicas}  b.replicas={b.replicas}   (b = a NO copia)")
