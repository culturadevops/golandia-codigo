package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

var _ = sql.ErrNoRows

// Un manejador realista: pide un dato y lo formatea.
// El formateo es el codigo «lento»; la peticion es la espera real.
func trabajoCPU(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "fila %d;", i)
	}
	return b.String()
}

func main() {
	// La «base de datos»: 40 ms por consulta, que es de lo normal tirando a bueno.
	bd := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(40 * time.Millisecond)
		w.Write([]byte("ok"))
	}))
	defer bd.Close()

	medir := func(nombre string, filas int, consultas int) {
		t := time.Now()
		for i := 0; i < consultas; i++ {
			res, _ := http.Get(bd.URL)
			res.Body.Close()
		}
		tBD := time.Since(t)
		t2 := time.Now()
		_ = trabajoCPU(filas)
		tCPU := time.Since(t2)
		fmt.Printf("  %-30s  base de datos: %-8v  CPU: %-10v  total: %v\n",
			nombre, tBD.Round(time.Millisecond),
			tCPU.Round(10*time.Microsecond),
			(tBD + tCPU).Round(time.Millisecond))
	}

	fmt.Println("Una peticion que formatea 5.000 filas:")
	medir("1 consulta", 5000, 1)
	fmt.Println()
	fmt.Println("La misma, con el clasico N+1 (una consulta por fila... 50 de ellas):")
	medir("50 consultas (N+1)", 5000, 50)
	fmt.Println()
	fmt.Println("Optimizar el formateo al doble de rapido ahorra microsegundos.")
	fmt.Println("Quitar el N+1 ahorra dos segundos.")
}
