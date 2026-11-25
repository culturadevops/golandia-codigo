# Seguir un log como tail -f mientras lo rotan (mv) y luego lo truncan (copytruncate).
source /work/repo/comun.sh
tmp
cat > escritor.sh <<'FIN'
# 60 líneas, una cada 50 ms. En la 20 se rota con mv; en la 40, con copytruncate.
for i in $(seq 1 60); do
	echo "linea $i" >> app.log
	sleep 0.05
	if [ "$i" = 20 ]; then mv app.log app.log.1; fi
	if [ "$i" = 40 ]; then cp app.log app.log.2; : > app.log; fi
done
FIN
c 'cat escritor.sh'
for modo in --ingenuo ""; do
	rm -f app.log app.log.1 app.log.2
	touch app.log
	echo "\$ ops-seguir app.log 5 $modo &   y a la vez:  bash escritor.sh"
	ops-seguir app.log 5 $modo &
	S=$!
	sleep 0.3
	bash escritor.sh
	wait $S
	echo "[codigo de salida: $?]"
	c 'wc -l app.log.1 app.log.2 app.log'
done
