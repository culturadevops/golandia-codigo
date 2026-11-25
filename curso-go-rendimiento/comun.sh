# Funciones comunes de las lecciones. Cada comandos.sh la carga con: . ./comun.sh
# Se ejecuta dentro del contenedor golang (ver ejecutar_todo.py), desde la raíz del repo.
set -u

# Los perfiles, binarios y resultados van a un directorio temporal, nunca al repositorio.
T=${CURSO_TMP:-/tmp/curso-go-rendimiento}
mkdir -p "$T"

# paso "comando": lo imprime como en una terminal y lo ejecuta.
paso() {
	printf '$ %s\n' "$*"
	eval "$*" 2>&1
	echo
}

# Los benchmarks del servicio, 10 repeticiones de cada versión. Varias lecciones los usan:
# se corren una vez y se reutilizan.
bench_resumen() {
	[ -s "$T/resumen.txt" ] ||
		go test ./resumen -run '^$' -bench Resumen -benchmem -count 10 >"$T/resumen.txt"
}
