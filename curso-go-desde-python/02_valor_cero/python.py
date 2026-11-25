puertos = {"web": 443, "db": 5432}
print("== en Python ==")
try:
    puertos["cache"]
except KeyError as e:
    print(f"  puertos['cache'] -> KeyError: {e}")
print(f"  puertos.get('cache') -> {puertos.get('cache')}")
