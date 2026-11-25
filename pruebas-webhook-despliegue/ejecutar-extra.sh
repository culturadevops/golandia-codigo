#!/bin/bash
# Extra del vídeo 563: el evento ping y el content type de formulario. Corre DENTRO de golang:1.27.2:
#   MSYS_NO_PATHCONV=1 docker run --rm -v "$(pwd -W)":/pruebas golang:1.27.2 bash /pruebas/ejecutar-extra.sh
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
SHA=$(git rev-parse HEAD)
cd "$T"
git clone -q origin.git despliegue
WEBHOOK_SECRET=$SECRETO "$T/receptor" -repo "$T/despliegue" > "$T/receptor.log" 2>&1 &
RECEPTOR=$!
for _ in $(seq 300); do curl -s -m 1 http://127.0.0.1:8080/ > /dev/null && break; sleep 0.1; done

enviar() {  # $1 = evento, $2 = entrega, $3 = content type, $4 = fichero del cuerpo
  local code
  code=$(curl -s -m 5 -o "$T/resp-$2" -w '%{http_code}' -X POST http://127.0.0.1:9000/webhook \
    -H "X-GitHub-Event: $1" -H "X-GitHub-Delivery: $2" -H "Content-Type: $3" \
    -H "X-Hub-Signature-256: $("$T/receptor" firmar "$SECRETO" < "$4")" --data-binary @"$4")
  echo "HTTP $code · $(tr -d '\n' < "$T/resp-$2")"
}

echo "== 11 · el ping que GitHub manda al crear el webhook"
printf '{"zen":"Keep it logically awesome.","hook_id":1}' > "$T/ping.json"
enviar ping entrega-ping application/json "$T/ping.json"

echo "== 12 · push con content type de formulario (application/x-www-form-urlencoded)"
printf '{"ref":"refs/heads/main","after":"%s"}' "$SHA" > "$T/p.json"
printf 'payload=%s' "$(python3 -c 'import sys,urllib.parse;print(urllib.parse.quote(open(sys.argv[1]).read()),end="")' "$T/p.json" 2>/dev/null || sed 's/{/%7B/g;s/}/%7D/g;s/"/%22/g;s/:/%3A/g;s/,/%2C/g;s/\//%2F/g' "$T/p.json")" > "$T/form.txt"
head -c 60 "$T/form.txt"; echo "…"
enviar push entrega-form application/x-www-form-urlencoded "$T/form.txt"

kill "$RECEPTOR" 2>/dev/null
pkill -f "$T/despliegue/app" 2>/dev/null
echo FIN
