// Lección 4: configuración por entorno. Lo que pasa cuando no se valida.
//
//	go run ./lecciones/04_config
package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"tareas/internal/config"
)

func main() {
	fmt.Println("1) os.Getenv sin validar, con PUERTO sin definir:")
	puerto := os.Getenv("PUERTO")
	ln, err := net.Listen("tcp", ":"+puerto)
	fmt.Printf("  net.Listen(\":\"+%q) -> escuchando en %s, err = %v\n", puerto, ln.Addr(), err)
	ln.Close()

	fmt.Println("\n2) TIEMPO_PETICION=30 (sin unidad), con el error ignorado:")
	d, _ := time.ParseDuration("30")
	fmt.Printf("  plazo = %v  -> sin plazo: una petición colgada espera para siempre\n", d)

	dsn := "postgres://app:S3cr3ta-de-verdad@bd:5432/tareas"
	fmt.Println("\n3) registrar la configuración al arrancar:")
	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: sinHora}))
	fmt.Print("  string:      ")
	l.Info("arrancando", "bd", dsn)
	fmt.Print("  config.DSN:  ")
	l.Info("arrancando", "bd", config.DSN(dsn))

	fmt.Println("\n4) config.Cargar con el entorno mal en cuatro sitios:")
	malo := map[string]string{"PUERTO": "80a", "TIEMPO_PETICION": "30", "NIVEL_LOG": "todo"}
	_, err = config.Cargar(func(k string) string { return malo[k] })
	fmt.Println(err)
}

func sinHora(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		return slog.Attr{}
	}
	return a
}
