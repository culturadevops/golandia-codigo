// Lección 9: pprof en un servicio vivo, y por qué no exponerlo.
//
// Importar net/http/pprof registra /debug/pprof/ en http.DefaultServeMux. Si tu servicio
// público sirve DefaultServeMux (ListenAndServe con nil), los perfiles quedan en internet.
//
//	go run ./09_vivo -modo expuesto -token sk_demo   # pprof en el puerto público (MAL)
//	go run ./09_vivo -modo privado  -token sk_demo   # pprof solo en 127.0.0.1:6060
package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof" // registra /debug/pprof/* en http.DefaultServeMux
	"time"

	"curso-go-rendimiento/resumen"
)

func main() {
	modo := flag.String("modo", "privado", "expuesto o privado")
	version := flag.String("version", "V1", "V1 o V5: qué handler atiende /resumen")
	_ = flag.String("token", "", "token de la API de facturación (de mentira)")
	flag.Parse()

	h := resumen.HandlerV1
	if *version == "V5" {
		h = resumen.HandlerV5
	}

	switch *modo {
	case "expuesto":
		// El error de siempre: la ruta en el mux por defecto y nil en ListenAndServe.
		http.HandleFunc("/resumen", h)
		log.Println("servicio en 0.0.0.0:8080 (y pprof con él)")
		log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))

	case "privado":
		// El servicio con su propio mux: no conoce /debug/pprof/.
		publico := http.NewServeMux()
		publico.HandleFunc("/resumen", h)
		// pprof, en otro servidor y solo en la interfaz local. Para verlo desde fuera:
		// kubectl port-forward, un túnel SSH... nunca un puerto abierto.
		go func() {
			log.Println("pprof en 127.0.0.1:6060")
			log.Fatal(http.ListenAndServe("127.0.0.1:6060", http.DefaultServeMux))
		}()
		srv := &http.Server{Addr: "0.0.0.0:8080", Handler: publico, ReadHeaderTimeout: 5 * time.Second}
		log.Println("servicio en 0.0.0.0:8080")
		log.Fatal(srv.ListenAndServe())
	}
}
