go version go1.27.2 linux/amd64
$ go mod init mistareas && go get fyne.io/fyne/v2@v2.8.2
go: added fyne.io/fyne/v2 v2.8.2
$ go mod tidy   # sin esto, go build se queja de entradas que faltan en go.sum
go.sum: 78 líneas (Fyne y sus dependencias)
$ time go build -o mistareas .   # primera vez: compila también la parte en C de Fyne
compilado en 79 s
$ echo // cambio >> main.go && time go build -o mistareas .
recompilado en 1003 ms
$ ls -la mistareas
33232144 mistareas
$ go test -v -run TestUsarLaApp .
=== RUN   TestUsarLaApp
    captura_test.go:42: contador: 3 tareas
--- PASS: TestUsarLaApp (0.04s)
PASS
ok  	mistareas	0.048s
1-vacia.png
2-con-tareas.png
