#!/bin/bash
# Demo del vídeo 563 (golandia): webhook de despliegue en Go. Corre DENTRO de golang:1.27.2 (ver correr.sh).
# Todo pasa en un directorio de mktemp: repositorio «origin», clon de trabajo y clon de despliegue.
set -u
SECRETO="un-secreto-largo-de-demo"
T=$(mktemp -d)
cd "$T" || exit 1
git config --global user.email demo@golandia.dev
git config --global user.name demo
git config --global init.defaultBranch main
export GOFLAGS=-buildvcs=false

(cd /pruebas/receptor && go build -o "$T/receptor" .) || exit 1

git init -q --bare origin.git
git clone -q origin.git trabajo 2>/dev/null
cd trabajo
printf 'module app\n\ngo 1.27\n' > go.mod
sed 's/VERSION/v1/' /pruebas/app/main.go > main.go
git add . && git commit -qm v1 && git push -q origin main
cd "$T"
git clone -q origin.git despliegue

nueva_version() {  # commit + push de la versión $1; deja el sha en $SHA
  cd "$T/trabajo"
  sed "s/VERSION/$1/" /pruebas/app/main.go > main.go
  git commit -qam "$1" && git push -q origin main
  SHA=$(git rev-parse HEAD)
  cd "$T"
}

payload() {  # payload mínimo de un push: $1 = ref, $2 = sha
  printf '{"ref":"%s","after":"%s","repository":{"full_name":"golandia/demo"}}' "$1" "$2"
}

firma() { "$T/receptor" firmar "$SECRETO" < "$1"; }

enviar() {  # $1 = id de entrega, $2 = fichero del cuerpo, $3 = firma ("" = sin cabecera)
  local h=() code
  [ -n "$3" ] && h=(-H "X-Hub-Signature-256: $3")
  code=$(curl -s -m 5 -o "$T/resp-$1" -w '%{http_code}' -X POST http://127.0.0.1:9000/webhook \
    -H "X-GitHub-Event: push" -H "X-GitHub-Delivery: $1" -H "Content-Type: application/json" "${h[@]}" \
    --data-binary @"$2")
  echo "HTTP $code · $(tr -d '\n' < "$T/resp-$1")"
}

app() { curl -s -m 1 http://127.0.0.1:8080/ || echo "(no contesta)"; }

espera_version() {  # hasta 30 s
  for _ in $(seq 300); do [ "$(app)" = "hola desde $1" ] && return 0; sleep 0.1; done
  echo "($1 no llegó en 30 s)"
  return 1
}

echo "== 0 · el ejemplo de la documentación de GitHub"
printf 'Hello, World!' | "$T/receptor" firmar "It's a Secret to Everybody"; echo

echo
echo "== 1 · arranque: el receptor despliega lo que hay en main"
WEBHOOK_SECRET=$SECRETO "$T/receptor" -repo "$T/despliegue" > "$T/receptor.log" 2>&1 &
RECEPTOR=$!
espera_version v1
echo "app: $(app)"

echo
echo "== 2 · push firmado a main"
nueva_version v2
payload refs/heads/main "$SHA" > "$T/p2.json"
FIRMA=$(firma "$T/p2.json")
echo "firma: $FIRMA"
enviar entrega-0002 "$T/p2.json" "$FIRMA"
espera_version v2
echo "app: $(app)"

echo
echo "== 3 · firma hecha con otro secreto"
nueva_version v3-malicioso
payload refs/heads/main "$SHA" > "$T/p3.json"
enviar entrega-0003 "$T/p3.json" "$("$T/receptor" firmar "adivinado" < "$T/p3.json")"
echo "== 4 · sin firma"
enviar entrega-0004 "$T/p3.json" ""
echo "== 5 · firma buena del cuerpo de antes, cuerpo cambiado"
sed 's/golandia\/demo/atacante\/demo/' "$T/p2.json" > "$T/p5.json"
enviar entrega-0005 "$T/p5.json" "$FIRMA"
echo "== 6 · la misma entrega otra vez (reenvío)"
enviar entrega-0002 "$T/p2.json" "$FIRMA"
echo "== 7 · push a otra rama"
payload refs/heads/pruebas "$SHA" > "$T/p7.json"
enviar entrega-0007 "$T/p7.json" "$(firma "$T/p7.json")"
sleep 1
echo "app sigue en: $(app)"
# el v3-malicioso quedó en origin; se tapa con una versión buena
nueva_version v4
payload refs/heads/main "$SHA" > "$T/p8.json"
enviar entrega-0008 "$T/p8.json" "$(firma "$T/p8.json")"
espera_version v4
echo "app: $(app)"

echo
echo "== 8 · dos pushes casi a la vez (con cola)"
nueva_version v5; payload refs/heads/main "$SHA" > "$T/a.json"
nueva_version v6; payload refs/heads/main "$SHA" > "$T/b.json"
enviar entrega-0009 "$T/a.json" "$(firma "$T/a.json")" & P1=$!
enviar entrega-0010 "$T/b.json" "$(firma "$T/b.json")" & P2=$!
wait $P1 $P2
espera_version v6
sleep 2
echo "app: $(app)"

echo
echo "== 9 · cuánto tiempo no contesta la app durante un despliegue"
nueva_version v7; payload refs/heads/main "$SHA" > "$T/c.json"
( ok=0; fallo=0
  for _ in $(seq 400); do
    if curl -s -m 0.2 -o /dev/null http://127.0.0.1:8080/; then ok=$((ok+1)); else fallo=$((fallo+1)); fi
    :
  done
  echo "sondeo: $ok respuestas, $fallo fallos" ) & SONDA=$!
sleep 0.3
enviar entrega-0011 "$T/c.json" "$(firma "$T/c.json")"
wait $SONDA
echo "app: $(app)"

echo
echo "== log del receptor (con cola)"
cat "$T/receptor.log"
kill "$RECEPTOR" 2>/dev/null; wait "$RECEPTOR" 2>/dev/null
pkill -f "$T/despliegue/app" 2>/dev/null; sleep 0.5

echo
echo "== 10 · dos pushes a la vez SIN cola (cada uno en su goroutine)"
WEBHOOK_SECRET=$SECRETO "$T/receptor" -repo "$T/despliegue" -sin-cola > "$T/receptor2.log" 2>&1 &
RECEPTOR=$!
espera_version v7
for i in 1 2 3; do
  nueva_version "v8-$i-a"; payload refs/heads/main "$SHA" > "$T/x$i.json"
  nueva_version "v8-$i-b"; payload refs/heads/main "$SHA" > "$T/y$i.json"
  enviar "sin-cola-$i-a" "$T/x$i.json" "$(firma "$T/x$i.json")" & P1=$!
  enviar "sin-cola-$i-b" "$T/y$i.json" "$(firma "$T/y$i.json")" & P2=$!
  wait $P1 $P2
  sleep 4
done
echo "app: $(app)"
echo "== log del receptor (sin cola)"
cat "$T/receptor2.log"
kill "$RECEPTOR" 2>/dev/null
pkill -f "$T/despliegue/app" 2>/dev/null
echo FIN
