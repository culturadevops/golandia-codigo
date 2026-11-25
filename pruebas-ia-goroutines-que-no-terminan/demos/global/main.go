// Demo 705: el límite que dicen las notas de Go 1.27. La misma fuga, una vez con un canal local
// (el perfil la ve) y otra con un canal en una variable global (no la ve: sigue alcanzable).
package main

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"runtime/pprof"
	"time"
)

var global = make(chan int)

func fugaLocal() {
	ch := make(chan int)
	go func() { ch <- 1 }() // nadie lee nunca
}

func fugaGlobal() {
	go func() { global <- 1 }() // nadie lee nunca, pero global sigue alcanzable
}

func total() string {
	time.Sleep(100 * time.Millisecond)
	runtime.GC()
	var b bytes.Buffer
	pprof.Lookup("goroutineleak").WriteTo(&b, 1)
	return regexp.MustCompile(`goroutineleak profile: total \d+`).FindString(b.String())
}

func main() {
	fugaLocal()
	fmt.Println("canal local:  ", total(), "| goroutines vivas:", runtime.NumGoroutine())
	fugaGlobal()
	fmt.Println("canal global: ", total(), "| goroutines vivas:", runtime.NumGoroutine())
	os.Exit(0)
}
