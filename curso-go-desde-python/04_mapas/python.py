srv = {"web-01": 1, "web-02": 2, "web-03": 3, "db-01": 4, "cache-01": 5}
print("== en Python ==")
for i in range(1, 4):
    print(f"  vuelta {i}: {' '.join(srv)}")
print("  siempre el mismo orden: el de insercion")
