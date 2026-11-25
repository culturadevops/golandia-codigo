source /work/repo/comun.sh
ver a_carrera.go; proyecto p6_race a_carrera.go; correr p6_race; correr p6_race; correr p6_race
cd /tmp/p/p6_race
echo '$ go run -race .'
go run -race . > /tmp/race.txt 2>&1; rc=$?
sed -E 's/0x[0-9a-f]+/0x…/g' /tmp/race.txt | head -13; echo "..."; tail -3 /tmp/race.txt; echo "[codigo de salida: $rc]"
ver b_mutex.go; proyecto p6_race b_mutex.go; correr p6_race
cd /tmp/p/p6_race
echo '$ go run -race .'
go run -race . 2>&1; echo "[codigo de salida: $?]"
