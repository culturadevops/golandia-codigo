// simulador: 50 «servidores» en 127.0.0.1:9001-9050 con /salud. Cada uno tarda entre 50 y
// 300 ms en responder; aquí ninguno se cuelga: solo se mide esperar. Era el simulador del vídeo 388, sin
// los colgados. Solo existe para el banco.
package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	colgados := map[int]bool{}
	for p := 9001; p <= 9050; p++ {
		retardo := time.Duration(50+(p*37)%250) * time.Millisecond
		colgado := colgados[p]
		mux := http.NewServeMux()
		mux.HandleFunc("/salud", func(w http.ResponseWriter, r *http.Request) {
			if colgado {
				<-r.Context().Done()
				return
			}
			time.Sleep(retardo)
			fmt.Fprintln(w, "ok")
		})
		go http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", p), mux)
	}
	select {}
}
