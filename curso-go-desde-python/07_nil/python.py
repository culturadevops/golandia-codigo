def validar(ok):
    return None if ok else ValueError("puerto invalido")
err = validar(True)
print("== en Python ==")
print(f"  err is None -> {err is None}   (None es None, sin tipo escondido)")
