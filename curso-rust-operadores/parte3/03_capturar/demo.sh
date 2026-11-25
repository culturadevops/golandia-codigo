# Capturar la salida de un comando a un archivo, con rotación por tamaño y su código de salida.
source /work/repo/comun.sh
tmp
cat > ruidoso.sh <<'FIN'
for i in $(seq 1 400); do
	echo "procesando lote $i"
	if [ $((i % 50)) = 0 ]; then echo "aviso: lote $i lento" >&2; fi
done
echo "fallo al subir el resultado" >&2
exit 3
FIN
c 'bash ruidoso.sh > solo-stdout.log; echo "código: $?"'
c 'grep -c fallo solo-stdout.log'
c 'bash ruidoso.sh 2>&1 | tee todo.log > /dev/null; echo "código: $?"'
c 'ops-capturar app.log 2048 3 -- bash ruidoso.sh'
c 'ls -l app.log*'
c 'cat app.log app.log.1 app.log.2 app.log.3 | wc -l'
c 'head -1 app.log.3'
c 'tail -2 app.log'
