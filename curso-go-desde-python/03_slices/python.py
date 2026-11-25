hosts = ["web-01", "web-02", "web-03", "web-04"]
primeros = hosts[:2]
primeros[0] = "CAMBIADO"
print("== en Python ==")
print(f"  hosts    = {hosts}")
print(f"  primeros = {primeros}")
print("  el corte hosts[:2] es una COPIA")
