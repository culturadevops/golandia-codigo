import tempfile, os
d = tempfile.mkdtemp()
print("== en Python ==")
for i in range(3):
    with open(os.path.join(d, f"f{i}.txt"), "w") as f:
        pass
    print(f"  vuelta {i}: cerrado al salir del with -> {f.closed}")
