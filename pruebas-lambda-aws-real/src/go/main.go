// Handler de prueba: EXACTAMENTE el mismo trabajo en los tres lenguajes.
//   entrada  {"n": 200000, "io_ms": 0}
//   salida   {"lang":"go","n":200000,"suma":...,"io_ms":0}
package main

import (
	"context"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

type Peticion struct {
	N    int `json:"n"`
	IoMs int `json:"io_ms"`
}

type Respuesta struct {
	Lang string `json:"lang"`
	N    int    `json:"n"`
	Suma uint64 `json:"suma"`
	IoMs int    `json:"io_ms"`
}

func manejar(ctx context.Context, p Peticion) (Respuesta, error) {
	var suma uint64
	for i := 1; i <= p.N; i++ {
		suma += uint64(i) * uint64(i) % 1000003
	}
	if p.IoMs > 0 {
		time.Sleep(time.Duration(p.IoMs) * time.Millisecond)
	}
	return Respuesta{Lang: "go", N: p.N, Suma: suma, IoMs: p.IoMs}, nil
}

func main() {
	lambda.Start(manejar)
}
