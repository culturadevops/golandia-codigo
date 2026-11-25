// Demos de io.Reader para el vídeo: un Reader propio, la cadena de Readers, la memoria de leer todo vs en streaming,
// y la trampa de Read (lee menos de lo que pides).
package main

import (
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"testing/iotest"
)

// Mayusculas es un io.Reader propio: envuelve otro Reader y pasa a mayúsculas lo que lee (solo ASCII).
type Mayusculas struct{ r io.Reader }

func (m Mayusculas) Read(p []byte) (int, error) {
	n, err := m.r.Read(p)
	for i := 0; i < n; i++ {
		if 'a' <= p[i] && p[i] <= 'z' {
			p[i] -= 'a' - 'A'
		}
	}
	return n, err
}

// Contador cuenta los bytes que pasan por él.
type Contador struct {
	r io.Reader
	n int64
}

func (c *Contador) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}

func pico() string {
	b, _ := os.ReadFile("/proc/self/status")
	for _, l := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(l, "VmHWM:") {
			kb, _ := strconv.Atoi(strings.Fields(l)[1])
			return fmt.Sprintf("%.1f MB", float64(kb)/1024)
		}
	}
	return "?"
}

func main() {
	switch os.Args[1] {
	case "propio":
		io.Copy(os.Stdout, Mayusculas{strings.NewReader("hola, golandia\n")})
	case "generar": // archivo de prueba: N MB de líneas de texto
		mb, _ := strconv.Atoi(os.Args[3])
		f, _ := os.Create(os.Args[2])
		w := bufio.NewWriter(f)
		linea := "golandia io.Reader 0123456789 abcdefghijklmnopqrstuvwxyz\n"
		for i := 0; i < mb*1024*1024/len(linea); i++ {
			w.WriteString(linea)
		}
		w.Flush()
		f.Close()
	case "todo": // lee el archivo entero a memoria y luego calcula el hash
		datos, _ := os.ReadFile(os.Args[2])
		fmt.Printf("sha256=%x bytes=%d pico_de_memoria=%s\n", sha256.Sum256(datos), len(datos), pico())
	case "stream": // lo mismo, pasando el archivo por el hash en trozos
		f, _ := os.Open(os.Args[2])
		h := sha256.New()
		n, _ := io.Copy(h, f)
		fmt.Printf("sha256=%x bytes=%d pico_de_memoria=%s\n", h.Sum(nil), n, pico())
	case "cadena": // archivo -> gzip -> (hash + contador) en una sola pasada
		f, _ := os.Open(os.Args[2])
		pr, pw := io.Pipe()
		go func() {
			gz := gzip.NewWriter(pw)
			io.Copy(gz, f)
			gz.Close()
			pw.Close()
		}()
		c := &Contador{r: pr}
		h := sha256.New()
		sal, _ := os.Create(os.Args[2] + ".gz")
		io.Copy(sal, io.TeeReader(c, h))
		sal.Close()
		fmt.Printf("comprimido=%d bytes sha256_del_gz=%x pico_de_memoria=%s\n", c.n, h.Sum(nil)[:8], pico())
	case "lineas": // contar líneas con bufio.Scanner, sin cargar el archivo
		f, _ := os.Open(os.Args[2])
		sc := bufio.NewScanner(f)
		n := 0
		for sc.Scan() {
			n++
		}
		fmt.Printf("lineas=%d err=%v pico_de_memoria=%s\n", n, sc.Err(), pico())
	case "linea-larga": // una línea de 100 KB: más que el máximo por defecto de bufio.Scanner
		texto := "corta\n" + strings.Repeat("x", 100*1024) + "\nfinal\n"
		sc := bufio.NewScanner(strings.NewReader(texto))
		n := 0
		for sc.Scan() {
			n++
		}
		fmt.Printf("sin Buffer:   líneas leídas=%d err=%v\n", n, sc.Err())
		sc = bufio.NewScanner(strings.NewReader(texto))
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		n = 0
		for sc.Scan() {
			n++
		}
		fmt.Printf("con Buffer 1MB: líneas leídas=%d err=%v\n", n, sc.Err())
	case "limite":
		cuerpo := strings.NewReader(strings.Repeat("A", 5000))
		datos, err := io.ReadAll(io.LimitReader(cuerpo, 1024))
		fmt.Printf("LimitReader(1024) sobre 5000 bytes: leí %d err=%v\n", len(datos), err)
	case "bucle": // el mismo Reader leído con el bucle mal escrito (mira err primero) y con el bien escrito
		buf := make([]byte, 4)
		var mal []byte
		r := iotest.DataErrReader(strings.NewReader("hola, golandia"))
		for {
			n, err := r.Read(buf)
			if err != nil {
				break
			}
			mal = append(mal, buf[:n]...)
		}
		var bien []byte
		r = iotest.DataErrReader(strings.NewReader("hola, golandia"))
		for {
			n, err := r.Read(buf)
			bien = append(bien, buf[:n]...)
			if err == io.EOF {
				break
			}
			if err != nil {
				panic(err)
			}
		}
		fmt.Printf("err primero: %q\n", mal)
		fmt.Printf("n primero:   %q\n", bien)
	case "trampa":
		buf := make([]byte, 8)
		r := iotest.HalfReader(strings.NewReader("0123456789"))
		n, err := r.Read(buf)
		fmt.Printf("Read:     pedí 8, llegaron %d: %q err=%v\n", n, buf[:n], err)
		r = iotest.HalfReader(strings.NewReader("0123456789"))
		n, err = io.ReadFull(r, buf)
		fmt.Printf("ReadFull: pedí 8, llegaron %d: %q err=%v\n", n, buf[:n], err)
		r = iotest.DataErrReader(strings.NewReader("fin"))
		n, err = r.Read(buf)
		fmt.Printf("último Read: n=%d %q err=%v   <- datos Y EOF a la vez\n", n, buf[:n], err)
	}
}
