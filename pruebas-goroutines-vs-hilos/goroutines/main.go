// Arranca N goroutines que se quedan esperando, y mide cuánto tarda y cuánta memoria e hilos del sistema usa el proceso.
package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// estado lee una línea de /proc/self/status (VmRSS en kB, Threads).
func estado(campo string) int {
	b, _ := os.ReadFile("/proc/self/status")
	for _, l := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(l, campo+":") {
			f := strings.Fields(l)
			n, _ := strconv.Atoi(f[1])
			return n
		}
	}
	return -1
}

func main() {
	n, _ := strconv.Atoi(os.Args[1])
	rssAntes := estado("VmRSS")
	var arrancadas, terminadas sync.WaitGroup
	espera := make(chan struct{})
	inicio := time.Now()
	for i := 0; i < n; i++ {
		arrancadas.Add(1)
		terminadas.Add(1)
		go func() {
			arrancadas.Done()
			<-espera
			terminadas.Done()
		}()
	}
	arrancadas.Wait()
	creadas := time.Since(inicio)
	rss := estado("VmRSS") - rssAntes
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("goroutines n=%d crear=%.1fms memoria=%.1fMB porcada=%.2fKB hilos_del_SO=%d pila_en_uso=%.1fMB GOMAXPROCS=%d\n",
		n, float64(creadas.Microseconds())/1000, float64(rss)/1024, float64(rss)/float64(n), estado("Threads"),
		float64(m.StackInuse)/1024/1024, runtime.GOMAXPROCS(0))
	close(espera)
	terminadas.Wait()
}
