#!/bin/sh
# Corre dentro de python:3.13.18-slim. /w = la carpeta del vídeo, montada de solo lectura.
pip install -q --root-user-action=ignore ruff==0.17.3 mypy==2.4.1 >/dev/null 2>&1
python -V; ruff --version; mypy --version
revisa() {
  echo "=== $1|$2|ruff"; ruff check --no-cache --select E9,F,B --output-format concise "$3" 2>&1; echo "--- exit=$?"
  echo "=== $1|$2|mypy"; mypy --strict --no-incremental --cache-dir /tmp/mc "$3" 2>&1; echo "--- exit=$?"
}
cd /w/casos/python
revisa caso python precios.py
revisa caso python_bien precios_bien.py
echo "=== caso|python|run"; python precios.py 2>&1; echo "--- exit=$?"
echo "=== caso|python_bien|run"; python precios_bien.py 2>&1; echo "--- exit=$?"
for f in /w/ia/*/python/*.py; do
  cd "$(dirname "$f")"; revisa ia "${f#/w/}" "$(basename "$f")"
done
