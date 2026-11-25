# Lo cargan los demo.sh. Los comandos ops-* ya están compilados (correr.sh los compila una vez)
# y en el PATH.
#   c 'comando'    muestra «$ comando» tal cual, lo ejecuta y enseña el código de salida
#   tmp            crea un directorio temporal y entra en él (nada se toca fuera de ahí)
#   tmpshm         lo mismo, pero en /dev/shm (un tmpfs de 256 MiB: el espacio se mide sin ruido)
export LC_ALL=C.UTF-8
c() {
	echo "\$ $1"
	(eval "$1") 2>&1
	echo "[codigo de salida: $?]"
}
tmp() { cd "$(mktemp -d /tmp/ops.XXXX)" || exit 1; }
tmpshm() { cd "$(mktemp -d /dev/shm/ops.XXXX)" || exit 1; }
