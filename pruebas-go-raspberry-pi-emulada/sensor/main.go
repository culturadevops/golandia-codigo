// sensor: el servicio de ejemplo para la Raspberry Pi. Un LED que parpadea en un pin GPIO y un endpoint /estado.
//
//	-gpio=sim   el pin es un gpiotest.Pin de periph.io: un pin FALSO, en memoria (lo que se usa en las pruebas)
//	-gpio=real  intenta abrir GPIO17 con los drivers de periph.io/x/host (solo funciona en la placa de verdad)
//	-calculo N  no arranca el servicio: hace N vueltas de un cálculo y dice cuánto tardó (para comparar emulado y nativo)
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/gpio/gpiotest"
	"periph.io/x/host/v3"
)

func leer(ruta string) string {
	b, err := os.ReadFile(ruta)
	if err != nil {
		return "(no existe: " + ruta + ")"
	}
	return strings.TrimRight(string(b), "\x00\n")
}

func goarm() string {
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, s := range bi.Settings {
			if s.Key == "GOARM" {
				return s.Value
			}
		}
	}
	return ""
}

func calculo(n int) {
	t0 := time.Now()
	s := 0
	for i := 1; i <= n; i++ {
		s += (i * i) % 1000003
	}
	fmt.Printf("calculo n=%d suma=%d en %v (GOARCH=%s)\n", n, s, time.Since(t0).Round(time.Millisecond), runtime.GOARCH)
}

func main() {
	modo := flag.String("gpio", "sim", "sim o real")
	n := flag.Int("calculo", 0, "vueltas del cálculo (0 = arrancar el servicio)")
	flag.Parse()
	if *n > 0 {
		calculo(*n)
		return
	}
	var pin gpio.PinIO
	switch *modo {
	case "sim":
		pin = &gpiotest.Pin{N: "GPIO17", Num: 17}
		log.Printf("GPIO SIMULADO: gpiotest.Pin %s (en memoria, no toca ningún hardware)", pin)
	case "real":
		estado, err := host.Init()
		if err != nil {
			log.Fatalf("host.Init: %v", err)
		}
		log.Printf("drivers cargados: %d · omitidos: %d · con error: %d", len(estado.Loaded), len(estado.Skipped), len(estado.Failed))
		for _, f := range estado.Failed {
			log.Printf("  falló %s: %v", f.D, f.Err)
		}
		pin = gpioreg.ByName("GPIO17")
		if pin == nil {
			log.Fatalf("GPIO17 no existe en esta máquina: no hay driver de Raspberry Pi cargado")
		}
	}
	var encendido atomic.Bool
	go func() {
		for {
			v := !encendido.Load()
			if err := pin.Out(gpio.Level(v)); err != nil {
				log.Printf("pin.Out: %v", err)
			}
			encendido.Store(v)
			time.Sleep(time.Second)
		}
	}()
	http.HandleFunc("/estado", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"goos": runtime.GOOS, "goarch": runtime.GOARCH, "goarm": goarm(),
			"pin": pin.String(), "nivel": pin.Read().String(), "led_encendido": encendido.Load(),
			"modelo":      leer("/proc/device-tree/model"),
			"temperatura": leer("/sys/class/thermal/thermal_zone0/temp"),
			"pid":         os.Getpid(),
		})
	})
	log.Printf("sensor escuchando en :8080 (pid %d, %s/%s)", os.Getpid(), runtime.GOOS, runtime.GOARCH)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
