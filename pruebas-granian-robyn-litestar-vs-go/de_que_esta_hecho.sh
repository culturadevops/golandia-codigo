# ¿Qué lleva Rust y qué no? Se baja cada paquete de PyPI (sin dependencias) y se mira:
#  · la etiqueta del wheel: py3-none-any = Python puro; cp313-…-manylinux = trae código nativo compilado;
#  · dentro del .so nativo, las rutas que deja el compilador de Rust (/rustc/, .cargo/registry) o el nombre del fuente en C.
# Uso: docker run --rm -v "$PWD:/b" python:3.13.17-slim sh /b/de_que_esta_hecho.sh   → salida en SALIDA-lenguajes.md
cd /tmp
pip download -q --no-deps litestar==2.25.0 msgspec==0.22.0 granian==2.8.7 robyn==0.88.0 fastapi==0.162.0 uvicorn==0.54.0 2>/dev/null
echo "## Wheels (PyPI)"; ls *.whl
mkdir -p x
for w in msgspec-*.whl granian-*.whl robyn-*.whl; do python -m zipfile -e "$w" x; done
echo; echo "## Dentro del código nativo"
for f in x/granian/_granian*.so x/robyn/robyn*.so x/msgspec/_core*.so; do
  echo "$(basename "$f"): rutas /rustc/=$(grep -c -a '/rustc/' "$f") · .cargo/registry=$(grep -c -a '.cargo/registry' "$f") · _core.c=$(grep -c -a '_core.c' "$f")"
done
