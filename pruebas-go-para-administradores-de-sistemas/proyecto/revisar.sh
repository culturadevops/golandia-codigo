#!/bin/bash
# La versión de toda la vida: uno detrás de otro, con curl y 2 segundos de tope.
ini=$(date +%s%N); caidos=0; total=0
while read -r h; do
  total=$((total+1))
  code=$(curl -s -o /dev/null -m 2 -w '%{http_code}' "http://$h/salud")
  if [ "$code" != "200" ]; then caidos=$((caidos+1)); echo "CAÍDO  $h"; fi
done < hosts.txt
fin=$(date +%s%N)
echo "$total servidores · $((total-caidos)) bien · $caidos caídos · $(awk "BEGIN{printf \"%.1f\", ($fin-$ini)/1e9}") s en total"
