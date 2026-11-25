# -*- coding: utf-8 -*-
import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).resolve().parent))
from probar import episodio

P1 = '''package main

import (
	"errors"
	"fmt"
)

var ErrNoEncontrado = errors.New("no encontrado")

func buscarBien(id int) error {
	return fmt.Errorf("buscando usuario %d: %w", id, ErrNoEncontrado) // %w
}

func buscarMal(id int) error {
	return fmt.Errorf("buscando usuario %d: %v", id, ErrNoEncontrado) // %v
}

func main() {
	bien, mal := buscarBien(7), buscarMal(7)

	fmt.Println("con %w ->", bien)
	fmt.Println("con %v ->", mal)
	fmt.Println("   los dos mensajes son IDENTICOS. Esa es la trampa.")
	fmt.Println()
	fmt.Println("errors.Is(bien, ErrNoEncontrado):", errors.Is(bien, ErrNoEncontrado))
	fmt.Println("errors.Is(mal,  ErrNoEncontrado):", errors.Is(mal, ErrNoEncontrado), " <-- la cadena esta rota")
	fmt.Println()
	fmt.Println("Unwrap(bien):", errors.Unwrap(bien))
	fmt.Println("Unwrap(mal): ", errors.Unwrap(mal))
}
'''

P2 = '''package main

import (
	"errors"
	"fmt"
)

type ErrHTTP struct {
	Codigo int
	URL    string
}

func (e *ErrHTTP) Error() string { return fmt.Sprintf("%s devolvio %d", e.URL, e.Codigo) }

func llamar() error {
	return fmt.Errorf("sincronizando: %w", &ErrHTTP{Codigo: 429, URL: "/api/pedidos"})
}

func main() {
	err := llamar()
	fmt.Println("el error:", err)

	// Is compara IDENTIDAD. As extrae el TIPO para poder leerle los campos.
	var e *ErrHTTP
	if errors.As(err, &e) {
		fmt.Println("errors.As lo saco. Codigo =", e.Codigo, "URL =", e.URL)
		if e.Codigo == 429 {
			fmt.Println("-> puedo decidir: esperar y reintentar")
		}
	}
	fmt.Println()
	fmt.Println("Esa es la diferencia en una frase:")
	fmt.Println("  Is  = es ESTE error?           (necesitas comparar)")
	fmt.Println("  As  = es de ESTE TIPO, damelo. (necesitas los datos)")
}
'''

P3 = '''package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

func main() {
	// Un servidor que tarda 2 segundos a proposito.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(2 * time.Second):
			fmt.Fprintln(w, "por fin")
		case <-r.Context().Done():
			fmt.Println("   [servidor] el cliente se fue:", r.Context().Err())
		}
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel() // <-- SIEMPRE. Sin esto se fuga el temporizador.

	req, _ := http.NewRequestWithContext(ctx, "GET", srv.URL, nil)
	inicio := time.Now()
	_, err := http.DefaultClient.Do(req)
	fmt.Printf("volvio en %v\\n", time.Since(inicio).Round(10*time.Millisecond))
	fmt.Println("error:", err)
	fmt.Println()
	fmt.Println("ctx.Err():                          ", ctx.Err())
	fmt.Println("errors.Is(ctx.Err(), DeadlineExceeded):", ctx.Err() == context.DeadlineExceeded)
	time.Sleep(100 * time.Millisecond) // que de tiempo a ver el mensaje del servidor
}
'''

P4 = '''package main

import (
	"context"
	"fmt"
	"time"
)

func trabajo() {
	// El descarte: ni se guarda el cancel ni se llama. Es el caso real.
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	<-ctx.Done()
	fmt.Println("listo")
}

func main() { trabajo() }
'''

P5 = '''package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"io"
)

// El UNICO sitio donde recover tiene sentido: el borde.
func recuperar(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				fmt.Println("   [middleware] atrapado:", p)
				http.Error(w, "error interno", http.StatusInternalServerError)
			}
		}()
		siguiente.ServeHTTP(w, r)
	})
}

func main() {
	roto := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m map[string]int
		m["boom"] = 1 // panic: escribir en un mapa nil (el del episodio 2)
	})

	srv := httptest.NewServer(recuperar(roto))
	defer srv.Close()

	res, err := http.Get(srv.URL)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer res.Body.Close()
	cuerpo, _ := io.ReadAll(res.Body)
	fmt.Println("el cliente recibio:", res.StatusCode, string(cuerpo))
	fmt.Println("y el servidor SIGUE VIVO. Sin el recover, se caia entero.")
}
'''

episodio(
    "07-errores-y-context",
    "Episodio 7/8 — errores (Is/As, %w) y context",
    [
        {"nombre": "1. %w contra %v: el mismo mensaje, distinto errors.Is",
         "pregunta": "Envuelves con %v en vez de %w. Que se rompe?",
         "archivos": {"p01/main.go": P1},
         "correr": [(["go", "run", "./p01"], False)],
         "nota": "El visual del episodio: los dos mensajes IDENTICOS en pantalla y debajo true / false.\nPor eso el bug sobrevive tanto en codigo real: no se ve leyendo el log."},
        {"nombre": "2. errors.Is o errors.As: la respuesta en una frase",
         "pregunta": "Cuando uno y cuando el otro?",
         "archivos": {"p02/main.go": P2},
         "correr": [(["go", "run", "./p02"], False)]},
        {"nombre": "3. context.WithTimeout cortando una peticion lenta",
         "pregunta": "Que es un context y por que va de primer parametro?",
         "archivos": {"p03/main.go": P3},
         "correr": [(["go", "run", "./p03"], False)],
         "nota": "Lo importante en pantalla: la cancelacion VIAJA. El servidor se entera de que el\ncliente se fue. Eso es lo que no hace un simple timeout."},
        {"nombre": "4. El defer cancel() que falta, y quien lo caza",
         "pregunta": "Que pasa si te olvidas del cancel?",
         "archivos": {"p04/main.go": P4},
         "correr": [(["go", "vet", "./p04"], True)],
         "nota": "Otra vez `go vet` y no el compilador. La regla: en cuanto escribes WithTimeout o\nWithCancel, la linea siguiente es el defer cancel()."},
        {"nombre": "5. recover: solo en el borde, nunca en una libreria",
         "pregunta": "Cuando se usa panic y cuando recover?",
         "archivos": {"p05/main.go": P5},
         "correr": [(["go", "run", "./p05"], False)],
         "nota": "panic para lo IMPOSIBLE (un programador se equivoco), error para lo ESPERABLE (el\nmundo falla). Y recover solo en el borde, como aqui."},
    ])
