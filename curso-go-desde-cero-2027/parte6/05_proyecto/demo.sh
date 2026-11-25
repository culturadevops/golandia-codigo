source /work/repo/comun.sh
ver main.go; proyecto p6_huellas main.go
cd /tmp/p/p6_huellas
mkdir -p datos
for i in $(seq -w 1 16); do yes "línea $i de golandia" | head -c 64M > datos/f$i.txt; done
echo '$ ls datos | wc -l && du -sh datos'; ls datos | wc -l; du -sh datos | cut -f1
echo '$ nproc'; nproc
go build -o huellas . 2>&1
echo '$ ./huellas datos/f01.txt datos/f02.txt datos/no-existe.txt'
./huellas datos/f01.txt datos/f02.txt datos/no-existe.txt 2>&1 | sed -E 's/ · [0-9]+ ms$/ · … ms/'; echo "[codigo de salida: ${PIPESTATUS[0]}]"
echo '$ ./huellas'
./huellas 2>&1; echo "[codigo de salida: $?]"
./huellas -workers 1 datos/* >/dev/null 2>&1   # calentar la caché de disco
for w in 1 2 4 8 16; do
	echo "\$ ./huellas -workers $w datos/* > /dev/null   (3 veces)"
	for r in 1 2 3; do ./huellas -workers $w datos/* 2>&1 >/dev/null; done
done
echo '$ ./huellas -workers 1 datos/* | head -3'
./huellas -workers 1 datos/* 2>/dev/null | head -3
echo '$ diff <(./huellas -workers 1 datos/*) <(./huellas -workers 16 datos/*) && echo iguales'
diff <(./huellas -workers 1 datos/* 2>/dev/null) <(./huellas -workers 16 datos/* 2>/dev/null) && echo iguales
