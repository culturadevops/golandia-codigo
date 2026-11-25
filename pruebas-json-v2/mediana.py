# Lee la salida de go test -bench (varias pasadas) y deja una línea por benchmark: la pasada mediana por ns/op.
import collections
import re
import sys

d = collections.defaultdict(list)
for l in sys.stdin:
    if not l.startswith("Benchmark"):
        continue
    f = l.split()
    n = re.sub(r"-\d+$", "", f[0])
    d[n].append({f[k + 1]: float(f[k]) for k in range(2, len(f) - 1) if f[k + 1].endswith("/op")})
for n in sorted(d):
    r = sorted(d[n], key=lambda v: v["ns/op"])[len(d[n]) // 2]
    print("  %-34s %6.2f ms/op %6.2f MB/op %7d allocs/op" % (n, r["ns/op"] / 1e6, r["B/op"] / 1e6, r["allocs/op"]))
