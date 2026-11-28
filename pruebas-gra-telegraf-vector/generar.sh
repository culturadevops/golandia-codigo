#!/bin/bash
# generar.sh <N> — escribe /base/eventos.json (N líneas JSON, 1 de cada 10 con nivel=debug) y /base/access.log (N líneas
# del formato «combined» de nginx). Deterministas: mismas líneas en cada corrida.
set -e
N=$1
awk -v n=$N 'BEGIN{
  split("info warn error info info info info info info debug", niv, " ");
  for (i=1;i<=n;i++) {
    s=i%86400; printf "{\"ts\":\"2026-11-26T%02d:%02d:%02dZ\",\"nivel\":\"%s\",\"servicio\":\"api-%d\",\"ruta\":\"/v1/pedidos/%d\",\"estado\":%d,\"ms\":%d}\n",
      int(s/3600), int(s/60)%60, s%60, niv[(i%10)+1], i%7, i, (i%13==0?500:200), (i*7)%997
  }}' > /base/eventos.json
awk -v n=$N 'BEGIN{
  split("GET POST GET GET PUT GET DELETE GET", m, " ");
  for (i=1;i<=n;i++) {
    s=i%86400; printf "10.%d.%d.%d - - [26/Nov/2026:%02d:%02d:%02d +0000] \"%s /api/v1/recurso/%d?pagina=%d HTTP/1.1\" %d %d \"https://golandia.dev/ref/%d\" \"Mozilla/5.0 (X11; Linux x86_64) banco/%d\"\n",
      i%250, (i/250)%250, i%199, int(s/3600), int(s/60)%60, s%60, m[(i%8)+1], i, i%50, (i%17==0?404:200), (i*31)%50000, i%100, i%9
  }}' > /base/access.log
wc -l /base/eventos.json /base/access.log; ls -la /base
