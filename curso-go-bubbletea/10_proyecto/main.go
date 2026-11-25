// nubepanel · lección 10, el proyecto: un panel de terminal para nubecita.
//
//	go run ./10_proyecto                                  # en tu terminal, contra 127.0.0.1:4566
//	go run ./10_proyecto -endpoint http://otra:4566
//	go run ./10_proyecto -ssh 127.0.0.1:23234             # servido por ssh: ssh -p 23234 127.0.0.1
//
// Con -ssh, cada conexión recibe SU propio tea.Program, con su modelo y su tamaño de ventana.
// Nadie instala nada en su máquina: el binario vive en el servidor.
package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"

	tea "charm.land/bubbletea/v2"
	"charm.land/ssh"
	"charm.land/wish/v2"
	"charm.land/wish/v2/activeterm"
	bm "charm.land/wish/v2/bubbletea"

	"curso-tui/10_proyecto/panel"
)

func main() {
	endpoint := flag.String("endpoint", "http://127.0.0.1:4566", "nubecita, LocalStack, moto o AWS")
	porSSH := flag.String("ssh", "", "dirección en la que servir el panel por ssh (p. ej. 127.0.0.1:23234)")
	clave := flag.String("clave", ".ssh/nubepanel_ed25519", "clave de host del servidor ssh (se crea si no existe)")
	autorizadas := flag.String("autorizadas", "", "fichero authorized_keys: solo entran esas claves")
	flag.Parse()
	fuente := panel.NuevaAWS(*endpoint)

	if *porSSH == "" {
		if _, err := tea.NewProgram(panel.Nuevo(fuente)).Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	// ⚠️ Sin ningún método de autenticación, wish deja entrar a CUALQUIERA, sin credenciales.
	// Por eso, sin -autorizadas, solo se acepta escuchar en 127.0.0.1.
	opciones := []ssh.Option{wish.WithAddress(*porSSH), wish.WithHostKeyPath(*clave)}
	if *autorizadas != "" {
		opciones = append(opciones, wish.WithAuthorizedKeys(*autorizadas))
	} else if host(*porSSH) != "127.0.0.1" && host(*porSSH) != "localhost" {
		fmt.Fprintln(os.Stderr, "sin -autorizadas entraría cualquiera: solo se permite 127.0.0.1")
		os.Exit(2)
	}
	s, err := wish.NewServer(append(opciones,
		wish.WithMiddleware(
			bm.Middleware(func(ssh.Session) (tea.Model, []tea.ProgramOption) {
				return panel.Nuevo(fuente), nil
			}),
			activeterm.Middleware(), // el último se ejecuta primero: sin terminal, fuera
		))...,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ln, err := net.Listen("tcp", *porSSH)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("nubepanel por ssh en %s -> ssh -p %s %s\n", *porSSH, puerto(*porSSH), host(*porSSH))
	go func() {
		if err := s.Serve(ln); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			fmt.Fprintln(os.Stderr, err)
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	s.Close()
}

func host(a string) string   { h, _, _ := net.SplitHostPort(a); return h }
func puerto(a string) string { _, p, _ := net.SplitHostPort(a); return p }
