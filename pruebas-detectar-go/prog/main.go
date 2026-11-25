// Un programa corriente, con las librerías que usa cualquier herramienta de red:
// habla HTTP, maneja JSON y hace TLS. Se elige así porque es lo que delata al binario.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type respuesta struct {
	Estado string `json:"estado"`
	Host   string `json:"host"`
}

func main() {
	host, _ := os.Hostname()
	r := respuesta{Estado: "vivo", Host: host}
	b, _ := json.Marshal(r)
	fmt.Println(string(b))
	http.DefaultClient.Timeout = 0
}
