# Pruebas ejecutadas — Episodio 7/8 — errores (Is/As, %w) y context

> Salida **literal** de esta máquina. La genera `probar.py` y se vuelve a generar sola.
> ⛔ Lo que no esté aquí no puede salir en pantalla (canon: ninguna cifra inventada,
> y lo que se muestra como código se ejecuta antes).

**Entorno citable en el vídeo:** contenedor `golang:1.26` · `go version go1.26.7 linux/amd64`

---

## 1. %w contra %v: el mismo mensaje, distinto errors.Is

**La pregunta:** Envuelves con %v en vez de %w. Que se rompe?

`p01/main.go`

```go
package main

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
```

```console
$ go run ./p01
con %w -> buscando usuario 7: no encontrado
con %v -> buscando usuario 7: no encontrado
   los dos mensajes son IDENTICOS. Esa es la trampa.

errors.Is(bien, ErrNoEncontrado): true
errors.Is(mal,  ErrNoEncontrado): false  <-- la cadena esta rota

Unwrap(bien): no encontrado
Unwrap(mal):  <nil>
[codigo de salida: 0]
```

> El visual del episodio: los dos mensajes IDENTICOS en pantalla y debajo true / false.
> Por eso el bug sobrevive tanto en codigo real: no se ve leyendo el log.

---

## 2. errors.Is o errors.As: la respuesta en una frase

**La pregunta:** Cuando uno y cuando el otro?

`p02/main.go`

```go
package main

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
```

```console
$ go run ./p02
el error: sincronizando: /api/pedidos devolvio 429
errors.As lo saco. Codigo = 429 URL = /api/pedidos
-> puedo decidir: esperar y reintentar

Esa es la diferencia en una frase:
  Is  = es ESTE error?           (necesitas comparar)
  As  = es de ESTE TIPO, damelo. (necesitas los datos)
[codigo de salida: 0]
```

---

## 3. context.WithTimeout cortando una peticion lenta

**La pregunta:** Que es un context y por que va de primer parametro?

`p03/main.go`

```go
package main

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
	fmt.Printf("volvio en %v\n", time.Since(inicio).Round(10*time.Millisecond))
	fmt.Println("error:", err)
	fmt.Println()
	fmt.Println("ctx.Err():                          ", ctx.Err())
	fmt.Println("errors.Is(ctx.Err(), DeadlineExceeded):", ctx.Err() == context.DeadlineExceeded)
	time.Sleep(100 * time.Millisecond) // que de tiempo a ver el mensaje del servidor
}
```

```console
$ go run ./p03
volvio en 300ms
error: Get "http://127.0.0.1:39055": context deadline exceeded

ctx.Err():                           context deadline exceeded
errors.Is(ctx.Err(), DeadlineExceeded): true
   [servidor] el cliente se fue: context canceled
[codigo de salida: 0]
```

> Lo importante en pantalla: la cancelacion VIAJA. El servidor se entera de que el
> cliente se fue. Eso es lo que no hace un simple timeout.

---

## 4. El defer cancel() que falta, y quien lo caza

**La pregunta:** Que pasa si te olvidas del cancel?

`p04/main.go`

```go
package main

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
```

```console
$ go vet ./p04   # ⚠️ falla A PROPÓSITO
p04/main.go:11:7: the cancel function returned by context.WithTimeout should be called, not discarded, to avoid a context leak
[codigo de salida: 1]
```

> Otra vez `go vet` y no el compilador. La regla: en cuanto escribes WithTimeout o
> WithCancel, la linea siguiente es el defer cancel().

---

## 5. recover: solo en el borde, nunca en una libreria

**La pregunta:** Cuando se usa panic y cuando recover?

`p05/main.go`

```go
package main

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
```

```console
$ go run ./p05
   [middleware] atrapado: assignment to entry in nil map
el cliente recibio: 500 error interno

y el servidor SIGUE VIVO. Sin el recover, se caia entero.
[codigo de salida: 0]
```

> panic para lo IMPOSIBLE (un programador se equivoco), error para lo ESPERABLE (el
> mundo falla). Y recover solo en el borde, como aqui.
