#!/bin/bash
# Bash en paralelo: 50 curl a la vez con xargs -P. También tarda lo que el más lento.
ini=$(date +%s%N)
caidos=$(xargs -P 50 -I{} curl -s -o /dev/null -m 2 -w '{} %{http_code}\n' http://{}/salud < hosts.txt | grep -vc ' 200$')
fin=$(date +%s%N)
echo "$caidos caídos · $(awk "BEGIN{printf \"%.1f\", ($fin-$ini)/1e9}") s en total"
