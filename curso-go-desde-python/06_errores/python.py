print("== en Python ==")
try:
    int("cuatro")
except ValueError as e:
    print(f"  int('cuatro') -> ValueError: {e}")
print("  la excepcion te para en el sitio, aunque no la hayas pedido")
