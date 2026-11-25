#!/bin/sh
# Corre dentro de node:22.23.4-slim. /w = la carpeta del vídeo, montada de solo lectura.
mkdir /p && cd /p && cp /w/casos/ts_config/* . && npm i --silent --no-audit --no-fund >/dev/null 2>&1
node -v; npx tsc -v; npx eslint -v
proyecto() { t=$(mktemp -d); cp /p/tsconfig.json /p/eslint.config.mjs /p/package.json "$t"; ln -s /p/node_modules "$t/node_modules"; cp "$1" "$t/main.ts"; cd "$t"; }
for d in ts ts_bien; do
  proyecto /w/casos/$d/usuarios.ts
  echo "=== caso|$d|tsc"; npx tsc -p . > /tmp/o 2>&1; e=$?; sed 's/^main.ts/usuarios.ts/' /tmp/o; echo "--- exit=$e"
  echo "=== caso|$d|eslint"; npx eslint main.ts > /tmp/o 2>&1; e=$?; sed "s#^$t/main.ts#usuarios.ts#" /tmp/o; echo "--- exit=$e"
done
for f in /w/ia/*/ts/*.ts; do
  proyecto "$f"; rel=${f#/w/}
  echo "=== ia|$rel|tsc"; npx tsc -p . 2>&1; echo "--- exit=$?"
  echo "=== ia|$rel|eslint"; npx eslint --format json main.ts 2>&1; echo; echo "--- exit=$?"
done
