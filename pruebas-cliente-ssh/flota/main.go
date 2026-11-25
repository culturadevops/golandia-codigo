// flota: ejecuta un comando por SSH en muchos servidores a la vez y devuelve la salida ordenada por host.
// Vídeo 682 de Golandia («Cliente SSH en Go»). Solo usa golang.org/x/crypto/ssh.
package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

type resultado struct {
	host   string
	salida string
	errout string
	codigo int // -1: no llegó a ejecutarse o no terminó
	err    error
	dur    time.Duration
}

var (
	fHosts    = flag.String("hosts", "hosts.txt", "un host por línea")
	fCmd      = flag.String("cmd", "revisar", "comando a ejecutar")
	fUsuario  = flag.String("u", "admin", "usuario")
	fClave    = flag.String("clave", "id_ed25519", "clave privada")
	fAgente   = flag.Bool("agente", false, "firmar con el agente (SSH_AUTH_SOCK) en vez de leer la clave")
	fKnown    = flag.String("known", "known_hosts", "archivo known_hosts")
	fInseguro = flag.Bool("inseguro", false, "ssh.InsecureIgnoreHostKey: acepta cualquier servidor")
	fP        = flag.Int("p", 10, "cuántos hosts a la vez (1 = en fila)")
	fConectar = flag.Duration("conectar", 3*time.Second, "tope para TCP + handshake")
	fTimeout  = flag.Duration("timeout", 10*time.Second, "tope total por host")
	fIngenuo  = flag.Bool("ingenuo", false, "ssh.Dial con config.Timeout y nada más")
	fLlegada  = flag.Bool("llegada", false, "imprimir cada host al terminar, sin ordenar")
	fPty      = flag.Bool("pty", false, "pedir una terminal (RequestPty) antes del comando")
)

func main() {
	flag.Parse()
	hosts, err := leerHosts(*fHosts)
	salirSi(err)
	cfg, err := configurar()
	salirSi(err)

	inicio := time.Now()
	res := enParalelo(hosts, cfg)
	if !*fLlegada {
		for _, r := range res {
			imprimir(r)
		}
	}
	os.Exit(resumen(res, time.Since(inicio)))
}

func configurar() (*ssh.ClientConfig, error) {
	auth, err := autenticacion()
	if err != nil {
		return nil, err
	}
	cb, err := verificador()
	if err != nil {
		return nil, err
	}
	return &ssh.ClientConfig{
		User:            *fUsuario,
		Auth:            []ssh.AuthMethod{auth},
		HostKeyCallback: cb,
		Timeout:         *fConectar, // ojo: solo cubre la conexión TCP, no el handshake
	}, nil
}

func autenticacion() (ssh.AuthMethod, error) {
	if *fAgente {
		sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
		if err != nil {
			return nil, err
		}
		return ssh.PublicKeysCallback(agent.NewClient(sock).Signers), nil
	}
	pem, err := os.ReadFile(*fClave)
	if err != nil {
		return nil, err
	}
	firmante, err := ssh.ParsePrivateKey(pem)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(firmante), nil
}

func verificador() (ssh.HostKeyCallback, error) {
	if *fInseguro {
		return ssh.InsecureIgnoreHostKey(), nil
	}
	return knownhosts.New(*fKnown)
}

func enParalelo(hosts []string, cfg *ssh.ClientConfig) []resultado {
	res := make([]resultado, len(hosts))
	sem := make(chan struct{}, *fP)
	var wg sync.WaitGroup
	for i, h := range hosts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}        // pido turno
			defer func() { <-sem }() // lo devuelvo
			res[i] = uno(h, cfg)
			if *fLlegada {
				imprimir(res[i])
			}
		}()
	}
	wg.Wait()
	return res
}

func uno(host string, cfg *ssh.ClientConfig) (r resultado) {
	inicio := time.Now()
	r = resultado{host: host, codigo: -1}
	defer func() { r.dur = time.Since(inicio) }()

	ctx, cancel := context.WithTimeout(context.Background(), *fTimeout)
	defer cancel()
	if *fIngenuo {
		ctx = context.Background()
	}
	cli, err := conectar(ctx, host, cfg)
	if err != nil {
		r.err = err
		return r
	}
	defer cli.Close()
	r.salida, r.errout, r.codigo, r.err = correr(ctx, cli, *fCmd)
	return r
}

func conectar(ctx context.Context, host string, cfg *ssh.ClientConfig) (*ssh.Client, error) {
	addr := net.JoinHostPort(host, "22")
	if *fIngenuo {
		return ssh.Dial("tcp", addr, cfg)
	}
	d := net.Dialer{Timeout: *fConectar}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	conn.SetDeadline(time.Now().Add(*fConectar)) // el handshake también tiene tope
	c, chans, reqs, err := ssh.NewClientConn(conn, addr, cfg)
	if err != nil {
		conn.Close()
		return nil, err
	}
	conn.SetDeadline(time.Time{}) // ya conectado: manda el contexto
	return ssh.NewClient(c, chans, reqs), nil
}

func correr(ctx context.Context, cli *ssh.Client, cmd string) (string, string, int, error) {
	s, err := cli.NewSession()
	if err != nil {
		return "", "", -1, err
	}
	defer s.Close()
	if *fPty {
		if err := s.RequestPty("xterm", 24, 80, ssh.TerminalModes{}); err != nil {
			return "", "", -1, err
		}
	}
	var out, errb bytes.Buffer
	s.Stdout, s.Stderr = &out, &errb

	hecho := make(chan error, 1)
	go func() { hecho <- s.Run(cmd) }()
	select {
	case err = <-hecho:
	case <-ctx.Done():
		cli.Close() // corta la conexión: así el host lento no retiene a nadie
		return out.String(), errb.String(), -1, fmt.Errorf("el comando no terminó en %s", *fTimeout)
	}
	var salida *ssh.ExitError
	if errors.As(err, &salida) {
		return out.String(), errb.String(), salida.ExitStatus(), nil
	}
	if err != nil {
		return out.String(), errb.String(), -1, err
	}
	return out.String(), errb.String(), 0, nil
}

var mu sync.Mutex

func imprimir(r resultado) {
	mu.Lock()
	defer mu.Unlock()
	switch {
	case r.err != nil:
		fmt.Printf("%-5s  ERR   %5.1fs  %v\n", r.host, r.dur.Seconds(), r.err)
	default:
		fmt.Printf("%-5s  %3d   %5.1fs  %s\n", r.host, r.codigo, r.dur.Seconds(), primera(r.salida))
		if r.errout != "" {
			fmt.Printf("%-5s  stderr        %s\n", r.host, primera(r.errout))
		}
	}
}

func resumen(res []resultado, total time.Duration) int {
	var ok, codigo, caidos int
	for _, r := range res {
		switch {
		case r.err != nil:
			caidos++
		case r.codigo != 0:
			codigo++
		default:
			ok++
		}
	}
	fmt.Printf("%d hosts · %d ok · %d con código distinto de 0 · %d sin respuesta · %.1fs en total\n",
		len(res), ok, codigo, caidos, total.Seconds())
	switch {
	case caidos > 0:
		return 2
	case codigo > 0:
		return 1
	}
	return 0
}

// primera junta las líneas en una sola (con pty llegan con \r\n).
func primera(s string) string {
	s = strings.ReplaceAll(strings.TrimRight(s, "\r\n"), "\r", "")
	return strings.ReplaceAll(s, "\n", " | ")
}

func leerHosts(ruta string) ([]string, error) {
	f, err := os.Open(ruta)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var hosts []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if h := strings.TrimSpace(sc.Text()); h != "" && !strings.HasPrefix(h, "#") {
			hosts = append(hosts, h)
		}
	}
	return hosts, sc.Err()
}

func salirSi(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "flota:", err)
		os.Exit(3)
	}
}
