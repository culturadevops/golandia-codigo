// cliente_ssh · se conecta a `nubepanel -ssh` como lo haría `ssh -p 23234 127.0.0.1`, pide un
// terminal de 80x14, espera, pulsa q y enseña lo que le llegó por la conexión.
//
// Lo que llega por ssh no es texto: son los bytes que pintan un terminal (mover el cursor,
// borrar, colores). Para poder fotografiarlo se pasa por un terminal virtual (x/vt), que es lo
// mismo que hace tu emulador de terminal con ellos.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/x/vt"
	"golang.org/x/crypto/ssh"

	"curso-tui/comun"
)

func main() {
	dir := flag.String("a", "127.0.0.1:23234", "dirección del servidor")
	clave := flag.String("clave", "", "clave privada con la que entrar; vacío = SIN NINGUNA credencial")
	flag.Parse()
	var auth []ssh.AuthMethod // vacío: ni contraseña ni clave
	if *clave != "" {
		pem, err := os.ReadFile(*clave)
		if err != nil {
			fmt.Println("  no leo la clave:", err)
			os.Exit(2)
		}
		firmante, err := ssh.ParsePrivateKey(pem)
		if err != nil {
			fmt.Println("  clave rota:", err)
			os.Exit(2)
		}
		auth = []ssh.AuthMethod{ssh.PublicKeys(firmante)}
	}
	cfg := &ssh.ClientConfig{User: "admin", Auth: auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), Timeout: 5 * time.Second} // solo para la demo
	if *clave == "" {
		fmt.Println("  entrando como admin con: ninguna credencial")
	} else {
		fmt.Println("  entrando como admin con: la clave", *clave)
	}
	var c *ssh.Client
	var err error
	for i := 0; i < 50; i++ {
		if c, err = ssh.Dial("tcp", *dir, cfg); err == nil || strings.Contains(err.Error(), "unable to authenticate") {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		fmt.Println("  rechazado:", err)
		os.Exit(3)
	}
	defer c.Close()

	sesion, _ := c.NewSession()
	sesion.RequestPty("xterm-256color", 14, 80, ssh.TerminalModes{})
	entrada, _ := sesion.StdinPipe()
	salida, _ := sesion.StdoutPipe()
	pantalla := vt.NewSafeEmulator(80, 14)
	go io.Copy(pantalla, salida)  // lo que manda el servidor, al terminal virtual
	go io.Copy(entrada, pantalla) // y sus respuestas a las preguntas del programa, de vuelta
	if err := sesion.Shell(); err != nil {
		fmt.Println("  no abre sesión:", err)
		os.Exit(2)
	}

	time.Sleep(1500 * time.Millisecond)
	comun.Pantalla("lo que ve quien hace ssh", strings.TrimRight(pantalla.String(), "\n "))
	entrada.Write([]byte("q"))
	fin := make(chan error, 1)
	go func() { fin <- sesion.Wait() }()
	select {
	case err := <-fin:
		fmt.Printf("  q -> la sesión se cerró (err = %v)\n", err)
	case <-time.After(3 * time.Second):
		fmt.Println("  q -> la sesión NO se cerró")
	}
}
