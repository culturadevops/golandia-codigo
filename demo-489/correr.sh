#!/usr/bin/env bash
# Demo del vídeo 489 «¿Cómo hago un bot de Discord en Go?» — se EJECUTA y su salida va a SALIDA.md.
# Sin token de Discord: se genera un par de claves Ed25519 de PRUEBA, el bot recibe la pública (como en producción) y
# simular/ firma las interacciones con la privada, como haría Discord. Todo en un contenedor golang:1.27.2.
set -u
export MSYS_NO_PATHCONV=1
D="$(cd "$(dirname "$0")" && pwd)"
SEMILLA=$(python -c "import os;print(os.urandom(32).hex())")
docker run --rm -v "$(cygpath -w "$D")":/src -w /src -e SEMILLA_PRUEBA="$SEMILLA" golang:1.27.2 sh -c '
  go version
  PUB=$(cat > /tmp/pub.go <<EOG
package main
import ("crypto/ed25519";"encoding/hex";"fmt";"os")
func main(){ s,_ := hex.DecodeString(os.Getenv("SEMILLA_PRUEBA")); fmt.Print(hex.EncodeToString(ed25519.NewKeyFromSeed(s).Public().(ed25519.PublicKey))) }
EOG
  go run /tmp/pub.go)
  echo "\$ go build -o bot ./bot && wc -c < bot"
  CGO_ENABLED=0 go build -o /tmp/bot ./bot && wc -c < /tmp/bot
  echo "\$ DISCORD_PUBLIC_KEY=<clave pública> ./bot &"
  DISCORD_PUBLIC_KEY=$PUB /tmp/bot 2>&1 | sed "s/^[0-9\/: ]*//" &
  sleep 1
  echo "\$ go run ./simular   # firma cada interacción como Discord"
  go run ./simular
  echo "\$ cat comandos.json   # lo que se registra en Discord (PUT /applications/<id>/commands, con el token del bot; NO se envía aquí)"
  cat comandos.json
'
