// Lección 1 · slog: registros con estructura, la clave que falta y el secreto que se cuela.
package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"curso-obs/comun"
)

type Usuario struct {
	Nombre string
	Email  string
	Clave  string
}

// UsuarioSeguro es el mismo usuario, pero decide él qué sale en un registro.
type UsuarioSeguro Usuario

func (u UsuarioSeguro) LogValue() slog.Value {
	return slog.GroupValue(slog.String("nombre", u.Nombre), slog.String("clave", "[oculta]"))
}

func main() {
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{} // sin la hora, para que la salida se pueda comparar
			}
			return a
		},
	}))
	linea := func() string {
		s := strings.TrimSpace(buf.String())
		buf.Reset()
		return s
	}

	comun.Titulo("1.1 · Un texto armado con Sprintf contra atributos")
	log.Info(fmt.Sprintf("pedido %d pagado por %s", 42, "ana"))
	fmt.Println("  " + linea())
	log.Info("pedido pagado", "pedido", 42, "usuario", "ana")
	fmt.Println("  " + linea())
	fmt.Println("  el segundo se puede filtrar por usuario=\"ana\"; el primero, solo con una expresión regular")

	comun.Titulo("1.2 · La trampa: se olvida una clave")
	// La misma llamada que 01_slog/malo/malo.go; aquí va en un slice para que el vet del repo pase.
	sinClave := []any{"pedido", 42, "tarjeta caducada"}
	log.Info("pago rechazado", sinClave...)
	fmt.Println("  " + linea())
	fmt.Println("  compila, se ejecuta y no da error: el valor huérfano sale como !BADKEY")
	out, err := exec.Command("go", "vet", "./01_slog/malo").CombinedOutput()
	codigo := 0
	if e, ok := err.(*exec.ExitError); ok {
		codigo = e.ExitCode()
	}
	fmt.Println("  pero go vet sí lo ve:")
	for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if strings.Contains(l, "malo.go") {
			fmt.Println("    " + l[strings.Index(l, "malo.go"):])
		}
	}
	fmt.Printf("    (go vet, código de salida %d)\n", codigo)

	comun.Titulo("1.3 · La otra trampa: registrar el struct entero")
	u := Usuario{Nombre: "ana", Email: "ana@ejemplo.com", Clave: "hunter2"}
	log.Info("inicio de sesión", "usuario", u)
	l := linea()
	fmt.Println("  " + l)
	fmt.Printf("  ¿sale la clave?: %v\n", strings.Contains(l, "hunter2"))

	comun.Titulo("1.4 · Lo correcto: que el tipo decida qué se registra (slog.LogValuer)")
	log.Info("inicio de sesión", "usuario", UsuarioSeguro(u))
	l = linea()
	fmt.Println("  " + l)
	fmt.Printf("  ¿sale la clave?: %v\n", strings.Contains(l, "hunter2"))
}
