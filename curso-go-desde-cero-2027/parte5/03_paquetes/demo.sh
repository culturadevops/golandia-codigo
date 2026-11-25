source /work/repo/comun.sh
montar() {
	rm -rf /tmp/p/tiendita; mkdir -p /tmp/p/tiendita/tienda
	cp "$AQUI_LECCION/$1" /tmp/p/tiendita/main.go
	cp "$AQUI_LECCION/tienda/tienda.go" /tmp/p/tiendita/tienda/
	(cd /tmp/p/tiendita && go mod init tiendita >/dev/null 2>&1)
}
ver tienda/tienda.go
montar a_main.go
echo '$ find . -name "*.go" -o -name go.mod'; (cd /tmp/p/tiendita && find . -name "*.go" -o -name go.mod | sort)
ver a_main.go; correr tiendita
ver b_main.go; montar b_main.go; correr tiendita
