# Pruebas ejecutadas — gRPC contra REST en Go — vídeo 119

> Salida **literal** de esta máquina. La genera `investigacion/pruebas_grpc.py` y se
> vuelve a generar sola. ⛔ Lo que no esté aquí no puede salir en pantalla (canon:
> ninguna cifra inventada, y lo que se muestra como código se ejecuta antes).

**Entorno citable en el vídeo:**

```console
$ go version; protoc --version; protoc-gen-go --version; protoc-gen-go-grpc --version
go version go1.26.7 linux/amd64
libprotoc 3.21.12
protoc-gen-go v1.36.6
protoc-gen-go-grpc 1.5.1
```

---

## 1 · El contrato: de un `.proto` salen cliente y servidor

**La pregunta:** ¿Qué genera `protoc` exactamente, y qué te da eso que un JSON no te da?

`banco/proto/pedidos.proto`

```protobuf
syntax = "proto3";

package pedidos;
option go_package = "banco/pb";

// El contrato. Esto NO es documentacion: de aqui salen el cliente y el servidor.
message Pedido {
  int64  id       = 1;
  string cliente  = 2;
  string articulo = 3;
  int32  unidades = 4;
  int64  centimos = 5;   // dinero en centimos, nunca en float
  string estado   = 6;
  string creado   = 7;
}

message Peticion  { int32 cuantos = 1; }
message Respuesta { repeated Pedido pedidos = 1; }

service Pedidos {
  rpc Listar(Peticion) returns (Respuesta);
  // El mismo dato, pero saliendo a medida que se produce.
  rpc ListarEnStreaming(Peticion) returns (stream Pedido);
}
```

```console
$ cd banco && protoc --go_out=. --go_opt=module=banco --go-grpc_out=. --go-grpc_opt=module=banco proto/pedidos.proto && protoc --go_out=. --go_opt=module=banco proto/v2/pedidos.proto && protoc --go_out=. --go_opt=module=banco proto/malo/pedidos.proto && ls -l pb/ && go mod tidy 2>&1 | tail -3 && go build ./...
total 20
-rw-r--r-- 1 root root 8200 Aug 24 11:56 pedidos.pb.go
-rw-r--r-- 1 root root 6205 Aug 24 11:56 pedidos_grpc.pb.go
[codigo de salida: 0]
```

> `protoc` no escribe documentación: escribe **código que compila**. Los tipos del cliente y los del servidor salen del mismo fichero, así que no pueden desincronizarse sin que alguien se entere.

---

## 2 · Lo que el compilador caza y en REST no cazaría nadie

**La pregunta:** Si el servidor no cumple el contrato, ¿cuándo te enteras?

`roto/main.go`

```go
package main

// El contrato lo comprueba el COMPILADOR. Este servidor se deja un metodo del `.proto` sin
// implementar y ademas cambia el tipo que devuelve el otro. En REST esto es un 500 en produccion
// un martes; aqui es un error de compilacion antes de que exista el binario.

import (
	"context"

	"banco/pb"

	"google.golang.org/grpc"
)

type servicioIncompleto struct{} // ⛔ a proposito: SIN pb.UnimplementedPedidosServer

// Y la firma esta mal: el .proto dice que Listar devuelve Respuesta, no Pedido.
func (servicioIncompleto) Listar(context.Context, *pb.Peticion) (*pb.Pedido, error) {
	return &pb.Pedido{Id: 1}, nil
}

func main() {
	s := grpc.NewServer()
	pb.RegisterPedidosServer(s, servicioIncompleto{})
}
```

```console
$ cd roto && cp ../banco/go.mod . && sed -i 's/^module banco/module roto/' go.mod && echo 'replace banco => ../banco' >> go.mod && go mod edit -require=banco@v0.0.0 && go mod tidy >/dev/null 2>&1; go build ./...    # ⚠️ falla A PROPÓSITO
# roto
./main.go:24:30: cannot use servicioIncompleto{} (value of struct type servicioIncompleto) as pb.PedidosServer value in argument to pb.RegisterPedidosServer: servicioIncompleto does not implement pb.PedidosServer (wrong type for method Listar)
		have Listar(context.Context, *pb.Peticion) (*pb.Pedido, error)
		want Listar(context.Context, *pb.Peticion) (*pb.Respuesta, error)
[codigo de salida: 1]
```

> No hay binario. En una API REST este mismo error es un 500 en producción un martes por la tarde.

---

## 3 · Bytes por el cable, contados en el socket

**La pregunta:** ¿Cuánto ahorra protobuf de verdad, y a partir de qué tamaño se nota?

`banco/contar/contar.go`

```go
package contar

// Cuenta los bytes que pasan POR EL SOCKET, no los del cuerpo serializado.
//
// Es la unica forma honesta de comparar: si contaramos solo el JSON contra solo el protobuf,
// dejariamos fuera las cabeceras de HTTP/1.1 y el marco de HTTP/2 con su HPACK, que es
// justamente donde gRPC se gasta parte de lo que ahorra.

import (
	"net"
	"sync/atomic"
)

type Contador struct {
	Enviados  atomic.Int64
	Recibidos atomic.Int64
}

func (c *Contador) Cero() { c.Enviados.Store(0); c.Recibidos.Store(0) }

func (c *Contador) Total() int64 { return c.Enviados.Load() + c.Recibidos.Load() }

type marcada struct {
	net.Conn
	c *Contador
}

func (m marcada) Read(p []byte) (int, error) {
	n, err := m.Conn.Read(p)
	m.c.Recibidos.Add(int64(n))
	return n, err
}

func (m marcada) Write(p []byte) (int, error) {
	n, err := m.Conn.Write(p)
	m.c.Enviados.Add(int64(n))
	return n, err
}

func (c *Contador) Marcar(base net.Conn) net.Conn { return marcada{base, c} }
```

`banco/cmd/bytes/main.go`

```go
package main

// ¿Cuantos bytes viajan de verdad? Cuatro combinaciones y cuatro tamaños de respuesta.
//
// Metodo: una llamada de calentamiento (que paga el saludo TLS-menos, el SETTINGS de HTTP/2 y
// la tabla HPACK), se pone el contador a cero, y se miden 20 llamadas. Se divide.
// ⚠️ Sin el calentamiento, gRPC sale MUCHO peor de lo que es: el coste de abrir la conexion
// HTTP/2 se repartia entre 1 llamada en vez de entre las miles que hace una app de verdad.

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"

	"banco/contar"
	"banco/datos"
	"banco/pb"
	"banco/servidores"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	gz "google.golang.org/grpc/encoding/gzip"
)

const VUELTAS = 20

func clienteREST(comp bool, c *contar.Contador) *http.Client {
	return &http.Client{Transport: &http.Transport{
		DisableCompression: !comp,
		DialContext: func(ctx context.Context, red, dir string) (net.Conn, error) {
			cn, err := (&net.Dialer{}).DialContext(ctx, red, dir)
			if err != nil {
				return nil, err
			}
			return c.Marcar(cn), nil
		},
	}}
}

func pideREST(cl *http.Client, dir string, n int) int {
	r, err := cl.Get(fmt.Sprintf("http://%s/pedidos?cuantos=%d", dir, n))
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()
	var xs []datos.Pedido
	var cuerpo io.Reader = r.Body
	if r.Header.Get("Content-Encoding") == "gzip" {
		z, err := gzip.NewReader(r.Body)
		if err != nil {
			panic(err)
		}
		defer z.Close()
		cuerpo = z
	}
	if err := json.NewDecoder(cuerpo).Decode(&xs); err != nil {
		panic(err)
	}
	return len(xs)
}

func main() {
	dirREST, paraR := servidores.REST()
	defer paraR()
	dirGRPC, paraG := servidores.GRPC()
	defer paraG()

	var cR, cRz, cG, cGz contar.Contador
	rest := clienteREST(false, &cR)
	restz := clienteREST(true, &cRz)

	marca := func(c *contar.Contador) grpc.DialOption {
		return grpc.WithContextDialer(func(ctx context.Context, dir string) (net.Conn, error) {
			cn, err := (&net.Dialer{}).DialContext(ctx, "tcp", dir)
			if err != nil {
				return nil, err
			}
			return c.Marcar(cn), nil
		})
	}
	abre := func(c *contar.Contador) pb.PedidosClient {
		cc, err := grpc.NewClient(dirGRPC, grpc.WithTransportCredentials(insecure.NewCredentials()), marca(c))
		if err != nil {
			panic(err)
		}
		return pb.NewPedidosClient(cc)
	}
	gr, grz := abre(&cG), abre(&cGz)

	fmt.Printf("%8s | %14s | %14s | %14s | %14s | %s\n",
		"pedidos", "REST/JSON", "REST+gzip", "gRPC/proto", "gRPC+gzip", "JSON / gRPC")
	fmt.Println("---------|----------------|----------------|----------------|----------------|------------")

	for _, n := range []int{1, 10, 100, 1000} {
		ctx := context.Background()
		// calentamiento
		pideREST(rest, dirREST, n)
		pideREST(restz, dirREST, n)
		gr.Listar(ctx, &pb.Peticion{Cuantos: int32(n)})
		grz.Listar(ctx, &pb.Peticion{Cuantos: int32(n)}, grpc.UseCompressor(gz.Name))

		cR.Cero()
		cRz.Cero()
		cG.Cero()
		cGz.Cero()
		for i := 0; i < VUELTAS; i++ {
			pideREST(rest, dirREST, n)
			pideREST(restz, dirREST, n)
			if _, err := gr.Listar(ctx, &pb.Peticion{Cuantos: int32(n)}); err != nil {
				panic(err)
			}
			if _, err := grz.Listar(ctx, &pb.Peticion{Cuantos: int32(n)},
				grpc.UseCompressor(gz.Name)); err != nil {
				panic(err)
			}
		}
		a := float64(cR.Total()) / VUELTAS
		b := float64(cRz.Total()) / VUELTAS
		c := float64(cG.Total()) / VUELTAS
		d := float64(cGz.Total()) / VUELTAS
		fmt.Printf("%8d | %11.0f B | %11.0f B | %11.0f B | %11.0f B | %10.2fx\n", n, a, b, c, d, a/c)
	}

	fmt.Println()
	fmt.Println("Bytes por llamada, contados en el socket (cabeceras incluidas), media de", VUELTAS, "llamadas")
	fmt.Println("sobre una conexion ya abierta. La ultima columna es cuantas veces mas gasta REST/JSON")
	fmt.Println("que gRPC sin comprimir.")
}
```

```console
$ cd banco && go run ./cmd/bytes
 pedidos |      REST/JSON |      REST+gzip |     gRPC/proto |      gRPC+gzip | JSON / gRPC
---------|----------------|----------------|----------------|----------------|------------
       1 |         349 B |         378 B |         237 B |         272 B |       1.47x
      10 |        1675 B |         582 B |         895 B |         506 B |       1.87x
     100 |       15019 B |        2228 B |        7488 B |        2097 B |       2.01x
    1000 |      148380 B |       17488 B |       73494 B |       16649 B |       2.02x

Bytes por llamada, contados en el socket (cabeceras incluidas), media de 20 llamadas
sobre una conexion ya abierta. La ultima columna es cuantas veces mas gasta REST/JSON
que gRPC sin comprimir.
[codigo de salida: 0]
```

---

## 4 · Latencia: p50 y p99 con 2.000 llamadas

**La pregunta:** ¿Es más rápido, o solo más pequeño?

`banco/cmd/latencia/main.go`

```go
package main

// ¿Y en tiempo? 2.000 llamadas seguidas de cada uno, con 100 pedidos de respuesta.
// Se ordenan y se sacan p50 y p99: la media esconde justo lo que duele.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"banco/datos"
	"banco/pb"
	"banco/servidores"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	LLAMADAS = 2000
	CUANTOS  = 100
)

func percentil(xs []time.Duration, p float64) time.Duration {
	i := int(float64(len(xs)-1) * p)
	return xs[i]
}

func resumen(nombre string, xs []time.Duration) {
	sort.Slice(xs, func(i, j int) bool { return xs[i] < xs[j] })
	var suma time.Duration
	for _, d := range xs {
		suma += d
	}
	fmt.Printf("%-12s  p50 %8.3f ms   p99 %8.3f ms   media %8.3f ms   min %8.3f ms\n",
		nombre,
		float64(percentil(xs, .50).Microseconds())/1000,
		float64(percentil(xs, .99).Microseconds())/1000,
		float64(suma.Microseconds())/1000/float64(len(xs)),
		float64(xs[0].Microseconds())/1000)
}

func main() {
	dirREST, paraR := servidores.REST()
	defer paraR()
	dirGRPC, paraG := servidores.GRPC()
	defer paraG()

	cl := &http.Client{Transport: &http.Transport{DisableCompression: true, MaxIdleConnsPerHost: 4}}
	cc, err := grpc.NewClient(dirGRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	gr := pb.NewPedidosClient(cc)
	ctx := context.Background()
	url := fmt.Sprintf("http://%s/pedidos?cuantos=%d", dirREST, CUANTOS)

	unaREST := func() {
		r, err := cl.Get(url)
		if err != nil {
			panic(err)
		}
		var xs []datos.Pedido
		json.NewDecoder(r.Body).Decode(&xs)
		r.Body.Close()
	}
	unaGRPC := func() {
		if _, err := gr.Listar(ctx, &pb.Peticion{Cuantos: CUANTOS}); err != nil {
			panic(err)
		}
	}

	for i := 0; i < 200; i++ { // calentamiento
		unaREST()
		unaGRPC()
	}

	tr := make([]time.Duration, LLAMADAS)
	tg := make([]time.Duration, LLAMADAS)
	for i := 0; i < LLAMADAS; i++ {
		t := time.Now()
		unaREST()
		tr[i] = time.Since(t)
		t = time.Now()
		unaGRPC()
		tg[i] = time.Since(t)
	}

	fmt.Printf("%d llamadas de cada uno, %d pedidos por respuesta, mismo proceso y localhost\n\n",
		LLAMADAS, CUANTOS)
	resumen("REST/JSON", tr)
	resumen("gRPC/proto", tg)
}
```

```console
$ cd banco && go run ./cmd/latencia
2000 llamadas de cada uno, 100 pedidos por respuesta, mismo proceso y localhost

REST/JSON     p50    0.308 ms   p99    0.972 ms   media    0.376 ms   min    0.222 ms
gRPC/proto    p50    0.156 ms   p99    0.646 ms   media    0.198 ms   min    0.085 ms
[codigo de salida: 0]
```

---

## 5 · Streaming: cuándo llega el PRIMER registro

**La pregunta:** Con 10.000 registros, ¿qué nota el usuario?

`banco/cmd/streaming/main.go`

```go
package main

// 10.000 pedidos de golpe contra 10.000 en streaming.
// Lo que importa no es el total: es CUANDO LLEGA EL PRIMERO, que es lo que ve el usuario.

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"banco/datos"
	"banco/pb"
	"banco/servidores"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const CUANTOS = 10000

func main() {
	dirREST, paraR := servidores.REST()
	defer paraR()
	dirGRPC, paraG := servidores.GRPC()
	defer paraG()

	cl := &http.Client{Transport: &http.Transport{DisableCompression: true}}
	cc, err := grpc.NewClient(dirGRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	gr := pb.NewPedidosClient(cc)
	ctx := context.Background()

	// calentamiento
	r0, _ := cl.Get(fmt.Sprintf("http://%s/pedidos?cuantos=%d", dirREST, CUANTOS))
	io.Copy(io.Discard, r0.Body)
	r0.Body.Close()
	gr.Listar(ctx, &pb.Peticion{Cuantos: CUANTOS})

	// 1) REST: hay que esperar a que llegue el JSON entero antes de tener el primer pedido
	t := time.Now()
	r, err := cl.Get(fmt.Sprintf("http://%s/pedidos?cuantos=%d", dirREST, CUANTOS))
	if err != nil {
		panic(err)
	}
	var xs []datos.Pedido
	json.NewDecoder(r.Body).Decode(&xs)
	r.Body.Close()
	restTotal := time.Since(t)

	// 2) gRPC unario: mismo problema, el mensaje llega entero o no llega
	t = time.Now()
	resp, err := gr.Listar(ctx, &pb.Peticion{Cuantos: CUANTOS})
	if err != nil {
		panic(err)
	}
	unarioTotal := time.Since(t)

	// 3) gRPC en streaming: el primero llega mucho antes
	t = time.Now()
	st, err := gr.ListarEnStreaming(ctx, &pb.Peticion{Cuantos: CUANTOS})
	if err != nil {
		panic(err)
	}
	var primero time.Duration
	n := 0
	for {
		_, err := st.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		if n == 0 {
			primero = time.Since(t)
		}
		n++
	}
	streamTotal := time.Since(t)

	ms := func(d time.Duration) float64 { return float64(d.Microseconds()) / 1000 }
	fmt.Printf("%d pedidos\n\n", CUANTOS)
	fmt.Printf("%-26s %10s %14s\n", "", "total", "primer pedido")
	fmt.Printf("%-26s %9.1f ms %13.1f ms   (%d pedidos)\n", "REST/JSON de una vez", ms(restTotal), ms(restTotal), len(xs))
	fmt.Printf("%-26s %9.1f ms %13.1f ms   (%d pedidos)\n", "gRPC unario", ms(unarioTotal), ms(unarioTotal), len(resp.Pedidos))
	fmt.Printf("%-26s %9.1f ms %13.1f ms   (%d pedidos)\n", "gRPC en streaming", ms(streamTotal), ms(primero), n)
	fmt.Println()
	fmt.Printf("El primer pedido llega %.0fx antes en streaming que esperando el JSON entero.\n",
		ms(restTotal)/ms(primero))
}
```

```console
$ cd banco && go run ./cmd/streaming
10000 pedidos

                                total  primer pedido
REST/JSON de una vez            15.1 ms          15.1 ms   (10000 pedidos)
gRPC unario                      7.2 ms           7.2 ms   (10000 pedidos)
gRPC en streaming                9.3 ms           0.4 ms   (10000 pedidos)

El primer pedido llega 35x antes en streaming que esperando el JSON entero.
[codigo de salida: 0]
```

---

## 6 · ⛔ El navegador no habla gRPC

**La pregunta:** ¿Qué pasa si llamas al puerto de gRPC como llamaría un navegador?

`banco/cmd/navegador/main.go`

```go
package main

// ⛔ El «cuando NO» mas caro: un navegador NO puede llamar a un servidor gRPC.
// No es una limitacion de permisos ni de CORS: fetch() no deja controlar los marcos de HTTP/2
// que gRPC necesita. Aqui se hace exactamente lo que haria un navegador —una peticion HTTP/1.1
// normal y corriente— contra el puerto de gRPC, y se enseña lo que contesta.

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"banco/servidores"
)

func main() {
	dir, para := servidores.GRPC()
	defer para()

	cn, err := net.Dial("tcp", dir)
	if err != nil {
		panic(err)
	}
	defer cn.Close()

	peticion := "GET /pedidos.Pedidos/Listar HTTP/1.1\r\nHost: " + dir + "\r\nAccept: application/json\r\n\r\n"
	fmt.Println("Lo que manda un navegador:")
	fmt.Println("  " + strings.ReplaceAll(strings.TrimSpace(peticion), "\r\n", "\n  "))
	fmt.Println()

	cn.Write([]byte(peticion))
	cn.SetReadDeadline(time.Now().Add(3 * time.Second))
	lector := bufio.NewReader(cn)
	buf := make([]byte, 512)
	n, err := lector.Read(buf)

	fmt.Println("Lo que contesta el servidor gRPC:")
	if n > 0 {
		fmt.Printf("  %d bytes: %q\n", n, string(buf[:n]))
	} else {
		fmt.Println("  nada. Ni una cabecera, ni un 400, ni un 404.")
	}
	if err != nil {
		fmt.Printf("  error de lectura: %v\n", err)
	}
	fmt.Println()
	fmt.Println("Por eso hace falta un proxy (grpc-web / Envoy) para hablar con gRPC desde el navegador.")
}
```

```console
$ cd banco && go run ./cmd/navegador
Lo que manda un navegador:
  GET /pedidos.Pedidos/Listar HTTP/1.1
  Host: 127.0.0.1:33371
  Accept: application/json

Lo que contesta el servidor gRPC:
  15 bytes: "\x00\x00\x06\x04\x00\x00\x00\x00\x00\x00\x05\x00\x00@\x00"

Por eso hace falta un proxy (grpc-web / Envoy) para hablar con gRPC desde el navegador.
[codigo de salida: 0]
```

---

## 7 · Depurar sin curl: `grpcurl` y la reflexión

**La pregunta:** Si no puedo hacer curl, ¿cómo miro qué hace el servicio?

`banco/cmd/servidor/main.go`

```go
package main

// Deja el servicio escuchando en un puerto fijo para poder interrogarlo desde fuera con grpcurl.

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"banco/servidores"
)

func main() {
	dir, para := servidores.GRPCEn("127.0.0.1:50051")
	defer para()
	fmt.Println("gRPC escuchando en", dir)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
```

```console
$ cd banco && go build -o /tmp/srv ./cmd/servidor && /tmp/srv & sleep 3; grpcurl -plaintext localhost:50051 list; echo; grpcurl -plaintext localhost:50051 describe pedidos.Pedidos; echo; grpcurl -plaintext -d '{"cuantos":2}' localhost:50051 pedidos.Pedidos/Listar; kill %1 2>/dev/null; true
gRPC escuchando en 127.0.0.1:50051
grpc.reflection.v1.ServerReflection
grpc.reflection.v1alpha.ServerReflection
pedidos.Pedidos

pedidos.Pedidos is a service:
service Pedidos {
  rpc Listar ( .pedidos.Peticion ) returns ( .pedidos.Respuesta );
  rpc ListarEnStreaming ( .pedidos.Peticion ) returns ( stream .pedidos.Pedido );
}

{
  "pedidos": [
    {
      "id": "100000",
      "cliente": "cliente-00000",
      "articulo": "articulo-0000",
      "unidades": 1,
      "centimos": "199",
      "estado": "pendiente",
      "creado": "2026-08-01T00:00:00Z"
    },
    {
      "id": "100001",
      "cliente": "cliente-00001",
      "articulo": "articulo-0001",
      "unidades": 2,
      "centimos": "236",
      "estado": "enviado",
      "creado": "2026-08-02T01:01:00Z"
    }
  ]
}
[codigo de salida: 0]
```

> El servicio se describe SOLO, porque el servidor tiene la reflexión activada (`reflection.Register`). ⚠️ Sin esa línea, `grpcurl` no sabe qué métodos hay y hay que pasarle el `.proto` a mano cada vez.

---

## 8 · El contrato cambia: ¿se cae el cliente viejo?

**La pregunta:** Si el servidor añade un campo y el cliente no se actualiza, ¿qué pasa?

`banco/proto/v2/pedidos.proto`

```protobuf
syntax = "proto3";

// El paquete se llama distinto SOLO porque las tres variantes se enlazan en el mismo binario
// de prueba y el registro de protobuf de Go no admite dos `pedidos.Pedido`. En el cable no
// viaja el nombre del paquete: solo los NUMEROS de campo y su tipo. La demostracion es valida.
package pedidosv2;
option go_package = "banco/pbv2";

// La MISMA version de antes MAS un campo nuevo, el 8. Nada mas cambia.
message Pedido {
  int64  id       = 1;
  string cliente  = 2;
  string articulo = 3;
  int32  unidades = 4;
  int64  centimos = 5;
  string estado   = 6;
  string creado   = 7;
  string almacen  = 8;   // <- lo nuevo
}
```

`banco/proto/malo/pedidos.proto`

```protobuf
syntax = "proto3";

package pedidosmalo;   // ver la nota del v2: el nombre del paquete no viaja por el cable
option go_package = "banco/pbmalo";

// LO QUE NO SE HACE NUNCA: reutilizar un numero de campo para otra cosa.
// El 2 era `string cliente`. Aqui alguien lo aprovecho para meter un entero.
message Pedido {
  int64  id        = 1;
  int64  descuento = 2;  // <- ERA `string cliente`
  string articulo  = 3;
  int32  unidades  = 4;
  int64  centimos  = 5;
  string estado    = 6;
  string creado    = 7;
}
```

`banco/cmd/contrato/main.go`

```go
package main

// Que pasa cuando el contrato cambia y las dos partes no se actualizan a la vez.
// Es la pregunta que decide si puedes desplegar tus servicios por separado o no.

import (
	"fmt"

	"banco/pb"
	"banco/pbmalo"
	"banco/pbv2"

	"google.golang.org/protobuf/proto"
)

func main() {
	// El servidor NUEVO manda un pedido con el campo 8, que el cliente viejo no conoce.
	nuevo := &pbv2.Pedido{Id: 42, Cliente: "cliente-00042", Articulo: "articulo-0042",
		Unidades: 3, Centimos: 1999, Estado: "enviado", Creado: "2026-08-24T10:00:00Z",
		Almacen: "MADRID-3"}
	crudo, err := proto.Marshal(nuevo)
	if err != nil {
		panic(err)
	}
	fmt.Printf("1) el servidor NUEVO manda %d bytes, con el campo 8 (almacen=%q)\n\n",
		len(crudo), nuevo.Almacen)

	// El cliente VIEJO, generado del .proto de 7 campos, lo lee.
	viejo := &pb.Pedido{}
	if err := proto.Unmarshal(crudo, viejo); err != nil {
		fmt.Println("   el cliente viejo FALLO:", err)
	} else {
		fmt.Printf("2) el cliente VIEJO lo lee sin enterarse: id=%d cliente=%q estado=%q\n",
			viejo.Id, viejo.Cliente, viejo.Estado)
		fmt.Println("   no conoce el campo 8, y no le importa. Nadie se cae.")
	}

	// Y lo mejor: si ese cliente viejo lo vuelve a serializar, el campo que no entiende
	// NO SE PIERDE. protobuf guarda los campos desconocidos y los devuelve tal cual.
	otraVez, _ := proto.Marshal(viejo)
	rebote := &pbv2.Pedido{}
	proto.Unmarshal(otraVez, rebote)
	fmt.Printf("3) y si el viejo lo reenvia, el campo 8 SIGUE AHI: almacen=%q\n", rebote.Almacen)
	fmt.Println("   protobuf conserva los campos que no conoce. Por eso un servicio viejo")
	fmt.Println("   en medio no destroza los datos de uno nuevo.")

	// Y ahora lo que SI rompe: reutilizar un numero para otro tipo.
	fmt.Println()
	malo := &pbmalo.Pedido{}
	err = proto.Unmarshal(crudo, malo)
	fmt.Println("4) ahora alguien reutiliza el campo 2 (era string cliente) para un entero:")
	if err != nil {
		fmt.Printf("   proto.Unmarshal devuelve error: %v\n", err)
	} else {
		fmt.Printf("   NO da error. descuento=%d, y el nombre del cliente se perdio\n",
			malo.Descuento)
	}
	fmt.Println("   Este es el bug silencioso: anadir campos es gratis, reutilizar numeros no.")
}
```

```console
$ cd banco && go run ./cmd/contrato
1) el servidor NUEVO manda 78 bytes, con el campo 8 (almacen="MADRID-3")

2) el cliente VIEJO lo lee sin enterarse: id=42 cliente="cliente-00042" estado="enviado"
   no conoce el campo 8, y no le importa. Nadie se cae.
3) y si el viejo lo reenvia, el campo 8 SIGUE AHI: almacen="MADRID-3"
   protobuf conserva los campos que no conoce. Por eso un servicio viejo
   en medio no destroza los datos de uno nuevo.

4) ahora alguien reutiliza el campo 2 (era string cliente) para un entero:
   NO da error. descuento=0, y el nombre del cliente se perdio
   Este es el bug silencioso: anadir campos es gratis, reutilizar numeros no.
[codigo de salida: 0]
```

> Añadir un campo es gratis y los campos desconocidos **se conservan**. Reutilizar un número de campo es el bug silencioso de protobuf.

---

## 9 · El plazo del cliente, ¿llega al servidor?

**La pregunta:** El cliente da 200 ms. ¿Lo sabe el servidor ANTES de ponerse a trabajar?

`banco/cmd/plazos/main.go`

```go
package main

// El cliente dice «tienes 200 ms». Se entera el servidor?
//
// En HTTP el servidor se entera de que el cliente COLGO, pero no de cuanto tiempo le quedaba.
// En gRPC el plazo viaja en una cabecera y el servidor puede leerlo ANTES de empezar a trabajar.
// La diferencia es la que hay entre abortar a mitad y no llegar a empezar.

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"banco/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type lento struct{ pb.UnimplementedPedidosServer }

func (lento) Listar(ctx context.Context, _ *pb.Peticion) (*pb.Respuesta, error) {
	if plazo, hay := ctx.Deadline(); hay {
		fmt.Printf("   [servidor gRPC] me dicen que tengo %.0f ms\n",
			float64(time.Until(plazo).Microseconds())/1000)
	} else {
		fmt.Println("   [servidor gRPC] no me han dicho ningun plazo")
	}
	select {
	case <-time.After(2 * time.Second):
		return &pb.Respuesta{}, nil
	case <-ctx.Done():
		fmt.Println("   [servidor gRPC] me lo cancelan, paro:", ctx.Err())
		return nil, ctx.Err()
	}
}

func main() {
	lis, _ := net.Listen("tcp", "127.0.0.1:0")
	s := grpc.NewServer()
	pb.RegisterPedidosServer(s, lento{})
	go s.Serve(lis)
	defer s.Stop()

	cc, err := grpc.NewClient(lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	fmt.Println("gRPC - el cliente da 200 ms:")
	t := time.Now()
	_, err = pb.NewPedidosClient(cc).Listar(ctx, &pb.Peticion{Cuantos: 1})
	fmt.Printf("   [cliente] corta a los %.0f ms con %v\n",
		float64(time.Since(t).Microseconds())/1000, status.Code(err))
	time.Sleep(200 * time.Millisecond)

	fmt.Println()
	fmt.Println("HTTP - el mismo cliente con 200 ms de plazo:")
	lh, _ := net.Listen("tcp", "127.0.0.1:0")
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if plazo, hay := r.Context().Deadline(); hay {
			fmt.Printf("   [servidor HTTP] me dicen que tengo %v\n", time.Until(plazo))
		} else {
			fmt.Println("   [servidor HTTP] no se cuanto tiempo tengo: la peticion NO trae plazo")
		}
		select {
		case <-time.After(2 * time.Second):
			w.Write([]byte("tarde"))
		case <-r.Context().Done():
			fmt.Println("   [servidor HTTP] me entero de que colgo, pero ya habia empezado:",
				r.Context().Err())
		}
	})}
	go srv.Serve(lh)
	defer srv.Close()

	cl := &http.Client{Timeout: 200 * time.Millisecond}
	t = time.Now()
	_, _ = cl.Get("http://" + lh.Addr().String() + "/pedidos")
	fmt.Printf("   [cliente] corta a los %.0f ms\n", float64(time.Since(t).Microseconds())/1000)
	time.Sleep(400 * time.Millisecond) // que de tiempo a imprimirse lo del servidor
}
```

```console
$ cd banco && go run ./cmd/plazos
gRPC - el cliente da 200 ms:
   [servidor gRPC] me dicen que tengo 199 ms
   [servidor gRPC] me lo cancelan, paro: context deadline exceeded
   [cliente] corta a los 201 ms con DeadlineExceeded

HTTP - el mismo cliente con 200 ms de plazo:
   [servidor HTTP] no se cuanto tiempo tengo: la peticion NO trae plazo
   [cliente] corta a los 201 ms
   [servidor HTTP] me entero de que colgo, pero ya habia empezado: context canceled
[codigo de salida: 0]
```
