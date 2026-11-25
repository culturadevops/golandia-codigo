# Los N archivos más grandes de un árbol, sin guardar la lista entera.
source /work/repo/comun.sh
c 'ops-grandes /usr 5'
c 'find /usr -type f -printf "%s %p\n" | sort -rn | head -5'
c 'find /usr -type f | wc -l'
