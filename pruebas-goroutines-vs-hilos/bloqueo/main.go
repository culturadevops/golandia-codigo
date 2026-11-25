// Goroutines que se bloquean DENTRO de una llamada al sistema (nanosleep directo, sin pasar por el runtime).
// Ahí el hilo se queda atrapado con ella, y el runtime tiene que crear otro hilo para seguir.
// Comparación: las mismas goroutines esperando con time.Sleep, que sí gestiona el runtime.
package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

func hilos() int {
	b, _ := os.ReadFile("/proc/self/status")
	for _, l := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(l, "Threads:") {
			n, _ := strconv.Atoi(strings.Fields(l)[1])
			return n
		}
	}
	return -1
}

func main() {
	modo := os.Args[1]
	n, _ := strconv.Atoi(os.Args[2])
	var wg sync.WaitGroup
	var ln net.Listener
	if modo == "red" { // un servidor local que acepta y nunca escribe: cada goroutine espera datos de la red
		ln, _ = net.Listen("tcp", "127.0.0.1:0")
		go func() {
			for {
				c, err := ln.Accept()
				if err != nil {
					return
				}
				defer c.Close()
			}
		}()
	}
	for i := 0; i < n; i++ {
		wg.Go(func() {
			if modo == "red" {
				c, err := net.Dial("tcp", ln.Addr().String())
				if err != nil {
					fmt.Println(err)
					return
				}
				c.SetReadDeadline(time.Now().Add(2 * time.Second))
				c.Read(make([]byte, 1))
				c.Close()
			} else if modo == "syscall" {
				ts := syscall.Timespec{Sec: 2}
				syscall.Syscall(syscall.SYS_NANOSLEEP, uintptr(unsafe.Pointer(&ts)), 0, 0)
			} else {
				time.Sleep(2 * time.Second)
			}
		})
	}
	time.Sleep(time.Second) // a mitad de la espera
	fmt.Printf("%-8s goroutines=%d hilos_del_SO=%d\n", modo, n, hilos())
	wg.Wait()
}
