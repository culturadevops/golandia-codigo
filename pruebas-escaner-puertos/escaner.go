// escaner.go — un escáner de puertos en Go, con la concurrencia como parámetro.
//
// ⛔ ÉTICO: esto se ejecuta contra un objetivo que levanto yo, en mi propia red de contenedores,
//    con puertos que yo mismo he abierto. Escanear una máquina ajena sin permiso por escrito es
//    otra cosa, y no es lo que hace este banco.
//
// La tesis que se mide: lanzar una goroutine por puerto (65.535) no es «más paralelo», es peor.
// Y lo peor no es el tiempo: es que EMPIEZA A FALLAR PUERTOS que están abiertos.
package main

import (
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("uso: escaner <host> <concurrencia> <timeout_ms> [hasta_puerto]")
		os.Exit(2)
	}
	host := os.Args[1]
	conc, _ := strconv.Atoi(os.Args[2])
	espera, _ := strconv.Atoi(os.Args[3])
	hasta := 65535
	if len(os.Args) > 4 {
		hasta, _ = strconv.Atoi(os.Args[4])
	}

	var (
		mu       sync.Mutex
		abiertos []int
		errores  = map[string]int{}
	)

	// El semáforo: un canal con capacidad. Cada goroutine tiene que meter una ficha
	// antes de trabajar y sacarla al terminar. Con conc=hasta, no frena nada.
	sem := make(chan struct{}, conc)
	var wg sync.WaitGroup

	inicio := time.Now()
	for p := 1; p <= hasta; p++ {
		wg.Add(1)
		go func(puerto int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			dir := net.JoinHostPort(host, strconv.Itoa(puerto))
			c, err := net.DialTimeout("tcp", dir, time.Duration(espera)*time.Millisecond)
			if err != nil {
				mu.Lock()
				errores[clasifica(err)]++
				mu.Unlock()
				return
			}
			c.Close()
			mu.Lock()
			abiertos = append(abiertos, puerto)
			mu.Unlock()
		}(p)
	}
	wg.Wait()
	dur := time.Since(inicio)

	sort.Ints(abiertos)
	fmt.Printf("concurrencia=%-6d timeout=%-5dms  ->  %d abiertos  en %v\n",
		conc, espera, len(abiertos), dur.Round(time.Millisecond))
	fmt.Printf("   puertos: %v\n", abiertos)

	// ⚠️ Los errores NO van a /dev/null: clasificarlos es la mitad del hallazgo.
	tipos := make([]string, 0, len(errores))
	for k := range errores {
		tipos = append(tipos, k)
	}
	sort.Strings(tipos)
	fmt.Print("   errores: ")
	for _, t := range tipos {
		fmt.Printf("%s=%d  ", t, errores[t])
	}
	fmt.Println()
}

// clasifica reduce el error a una etiqueta corta. «rechazado» es la respuesta NORMAL de un
// puerto cerrado; cualquier otra cosa es el sistema diciéndote que no puede más.
func clasifica(err error) string {
	s := err.Error()
	switch {
	case contiene(s, "refused"):
		return "rechazado"
	case contiene(s, "timeout"), contiene(s, "i/o timeout"), contiene(s, "deadline"):
		return "TIMEOUT"
	case contiene(s, "too many open files"), contiene(s, "cannot allocate"):
		return "SIN-DESCRIPTORES"
	case contiene(s, "cannot assign requested address"):
		return "SIN-PUERTOS-LOCALES"
	default:
		return "otro:" + s
	}
}

func contiene(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
