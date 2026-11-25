# -*- coding: utf-8 -*-
"""pruebas_grpc.py — mide de verdad lo que afirma el vídeo 119 (curso-grpc-go).

Ninguna cifra sale de un blog ni de la documentación de gRPC: se construyen los dos servicios
—el MISMO dato por REST/JSON y por gRPC/protobuf— y se miden aquí.

Los bytes se cuentan **en el socket**, envolviendo la `net.Conn` de los dos clientes. Así entra
todo lo que de verdad viaja: las cabeceras HTTP, el marco de HTTP/2 y la tabla HPACK. Contar solo
el cuerpo serializado sería hacer trampa a favor de gRPC.

    python canales/golandia/investigacion/pruebas_grpc.py
"""
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import correr_grpc as R  # noqa: E402

# ── el contrato ───────────────────────────────────────────────────────────────────────
PROTO = r'''syntax = "proto3";

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
'''

GOMOD = '''module banco

go 1.24

require (
	google.golang.org/grpc v1.71.0
	google.golang.org/protobuf v1.36.6
)
'''

CONTAR = r'''package contar

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
'''

DATOS = r'''package datos

// Datos DETERMINISTAS: sin aleatoriedad, para que dos corridas den exactamente los mismos bytes.

import "fmt"

type Pedido struct {
	ID       int64  `json:"id"`
	Cliente  string `json:"cliente"`
	Articulo string `json:"articulo"`
	Unidades int32  `json:"unidades"`
	Centimos int64  `json:"centimos"`
	Estado   string `json:"estado"`
	Creado   string `json:"creado"`
}

var estados = []string{"pendiente", "enviado", "entregado", "devuelto"}

func Uno(i int) Pedido {
	return Pedido{
		ID:       int64(100000 + i),
		Cliente:  fmt.Sprintf("cliente-%05d", i%997),
		Articulo: fmt.Sprintf("articulo-%04d", i%1301),
		Unidades: int32(1 + i%7),
		Centimos: int64(199 + (i*37)%9800),
		Estado:   estados[i%len(estados)],
		Creado:   fmt.Sprintf("2026-08-%02dT%02d:%02d:00Z", 1+i%28, i%24, i%60),
	}
}

func Muchos(n int) []Pedido {
	xs := make([]Pedido, n)
	for i := range xs {
		xs[i] = Uno(i)
	}
	return xs
}
'''

SERVIDORES = r'''package servidores

// Levanta LOS DOS servicios con el mismo dato detras. Cada programa de medida los arranca en su
// propio proceso y en puertos que pide el sistema, para que nada se pise.

import (
	"encoding/json"
	"net"
	"net/http"

	"banco/datos"
	"banco/pb"

	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip" // registra el compresor, no lo activa
	"google.golang.org/grpc/reflection"
)

// ── REST ──────────────────────────────────────────────────────────────────────────────
func REST() (string, func()) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/pedidos", func(w http.ResponseWriter, r *http.Request) {
		n := 100
		if v := r.URL.Query().Get("cuantos"); v != "" {
			n = 0
			for _, c := range v {
				n = n*10 + int(c-'0')
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(datos.Muchos(n))
	})
	// La compresion la decide el cliente con Accept-Encoding; el servidor la sirve si se la piden.
	s := &http.Server{Handler: gzipSiLaPiden(mux)}
	go s.Serve(lis)
	return lis.Addr().String(), func() { s.Close() }
}

// ── gRPC ──────────────────────────────────────────────────────────────────────────────
type servicio struct{ pb.UnimplementedPedidosServer }

func aPB(p datos.Pedido) *pb.Pedido {
	return &pb.Pedido{Id: p.ID, Cliente: p.Cliente, Articulo: p.Articulo,
		Unidades: p.Unidades, Centimos: p.Centimos, Estado: p.Estado, Creado: p.Creado}
}

func (servicio) Listar(_ context0, in *pb.Peticion) (*pb.Respuesta, error) {
	xs := datos.Muchos(int(in.Cuantos))
	out := make([]*pb.Pedido, len(xs))
	for i, p := range xs {
		out[i] = aPB(p)
	}
	return &pb.Respuesta{Pedidos: out}, nil
}

func (servicio) ListarEnStreaming(in *pb.Peticion, srv pb.Pedidos_ListarEnStreamingServer) error {
	for i := 0; i < int(in.Cuantos); i++ {
		if err := srv.Send(aPB(datos.Uno(i))); err != nil {
			return err
		}
	}
	return nil
}

func GRPC() (string, func()) { return GRPCEn("127.0.0.1:0") }

// Puerto fijo: `grpcurl` es otro proceso y necesita saber a donde llamar.
func GRPCEn(dir string) (string, func()) {
	lis, err := net.Listen("tcp", dir)
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	pb.RegisterPedidosServer(s, servicio{})
	reflection.Register(s) // sin esto, grpcurl no sabe que metodos hay
	go s.Serve(lis)
	return lis.Addr().String(), func() { s.Stop() }
}
'''

GZIPSRV = r'''package servidores

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type escritorGzip struct {
	http.ResponseWriter
	z *gzip.Writer
}

func (e escritorGzip) Write(p []byte) (int, error) { return e.z.Write(p) }

func gzipSiLaPiden(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			siguiente.ServeHTTP(w, r)
			return
		}
		z := gzip.NewWriter(w)
		defer z.Close()
		w.Header().Set("Content-Encoding", "gzip")
		siguiente.ServeHTTP(escritorGzip{w, z}, r)
	})
}
'''

CONTEXTO = r'''package servidores

import "context"

// Alias local para que la firma de arriba se lea de un vistazo en pantalla.
type context0 = context.Context
'''

# ── programa 1: los bytes por el cable ────────────────────────────────────────────────
BYTES = r'''package main

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
'''

# ── programa 2: latencia ──────────────────────────────────────────────────────────────
LATENCIA = r'''package main

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
'''

# ── programa 3: streaming ─────────────────────────────────────────────────────────────
STREAMING = r'''package main

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
'''

# ── programa 4: el navegador ──────────────────────────────────────────────────────────
NAVEGADOR = r'''package main

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
'''

SERVIDOR = r'''package main

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
'''

CONTRATO_ROTO = r'''package main

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
'''


# -- prueba 8: la evolucion del contrato -----------------------------------------------
PROTO_V2 = r'''syntax = "proto3";

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
'''

PROTO_MALO = r'''syntax = "proto3";

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
'''

CONTRATO = r'''package main

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
'''

# -- prueba 9: los plazos viajan --------------------------------------------------------
PLAZOS = r'''package main

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
'''


def main():
    entorno = None
    v = R.correr(R.RAIZ, "go version; protoc --version; protoc-gen-go --version; "
                         "protoc-gen-go-grpc --version", shell=True)
    entorno = f"$ go version; protoc --version; protoc-gen-go --version; protoc-gen-go-grpc --version\n{v['salida']}"

    archivos = {
        "banco/go.mod": GOMOD,
        "banco/proto/pedidos.proto": PROTO,
        "banco/contar/contar.go": CONTAR,
        "banco/datos/datos.go": DATOS,
        "banco/servidores/servidores.go": SERVIDORES,
        "banco/servidores/gzip.go": GZIPSRV,
        "banco/servidores/contexto.go": CONTEXTO,
        "banco/cmd/bytes/main.go": BYTES,
        "banco/cmd/latencia/main.go": LATENCIA,
        "banco/cmd/streaming/main.go": STREAMING,
        "banco/cmd/navegador/main.go": NAVEGADOR,
        "banco/cmd/servidor/main.go": SERVIDOR,
        "roto/main.go": CONTRATO_ROTO,
        "banco/proto/v2/pedidos.proto": PROTO_V2,
        "banco/proto/malo/pedidos.proto": PROTO_MALO,
        "banco/cmd/contrato/main.go": CONTRATO,
        "banco/cmd/plazos/main.go": PLAZOS,
    }

    generar = ("cd banco && protoc --go_out=. --go_opt=module=banco "
               "--go-grpc_out=. --go-grpc_opt=module=banco proto/pedidos.proto "
               "&& protoc --go_out=. --go_opt=module=banco proto/v2/pedidos.proto "
               "&& protoc --go_out=. --go_opt=module=banco proto/malo/pedidos.proto "
               "&& ls -l pb/ && go mod tidy 2>&1 | tail -3 && go build ./...")

    pruebas = [
        dict(nombre="1 · El contrato: de un `.proto` salen cliente y servidor",
             pregunta="¿Qué genera `protoc` exactamente, y qué te da eso que un JSON no te da?",
             archivos=archivos,
             muestra=["banco/proto/pedidos.proto"],
             correr=[(generar, False, True)],
             nota="`protoc` no escribe documentación: escribe **código que compila**. Los tipos "
                  "del cliente y los del servidor salen del mismo fichero, así que no pueden "
                  "desincronizarse sin que alguien se entere."),

        dict(nombre="2 · Lo que el compilador caza y en REST no cazaría nadie",
             pregunta="Si el servidor no cumple el contrato, ¿cuándo te enteras?",
             muestra=["roto/main.go"],
             correr=[("cd roto && cp ../banco/go.mod . && sed -i 's/^module banco/module roto/' go.mod "
                      "&& echo 'replace banco => ../banco' >> go.mod "
                      "&& go mod edit -require=banco@v0.0.0 && go mod tidy >/dev/null 2>&1; "
                      "go build ./... ", True, True)],
             nota="No hay binario. En una API REST este mismo error es un 500 en producción un "
                  "martes por la tarde."),

        dict(nombre="3 · Bytes por el cable, contados en el socket",
             pregunta="¿Cuánto ahorra protobuf de verdad, y a partir de qué tamaño se nota?",
             muestra=["banco/contar/contar.go", "banco/cmd/bytes/main.go"],
             correr=[("cd banco && go run ./cmd/bytes", False, True)]),

        dict(nombre="4 · Latencia: p50 y p99 con 2.000 llamadas",
             pregunta="¿Es más rápido, o solo más pequeño?",
             muestra=["banco/cmd/latencia/main.go"],
             correr=[("cd banco && go run ./cmd/latencia", False, True)]),

        dict(nombre="5 · Streaming: cuándo llega el PRIMER registro",
             pregunta="Con 10.000 registros, ¿qué nota el usuario?",
             muestra=["banco/cmd/streaming/main.go"],
             correr=[("cd banco && go run ./cmd/streaming", False, True)]),

        dict(nombre="6 · ⛔ El navegador no habla gRPC",
             pregunta="¿Qué pasa si llamas al puerto de gRPC como llamaría un navegador?",
             muestra=["banco/cmd/navegador/main.go"],
             correr=[("cd banco && go run ./cmd/navegador", False, True)]),

        dict(nombre="7 · Depurar sin curl: `grpcurl` y la reflexión",
             pregunta="Si no puedo hacer curl, ¿cómo miro qué hace el servicio?",
             muestra=["banco/cmd/servidor/main.go"],
             correr=[("cd banco && go build -o /tmp/srv ./cmd/servidor && /tmp/srv & "
                      "sleep 3; "
                      "grpcurl -plaintext localhost:50051 list; "
                      "echo; grpcurl -plaintext localhost:50051 describe pedidos.Pedidos; "
                      "echo; grpcurl -plaintext -d '{\"cuantos\":2}' "
                      "localhost:50051 pedidos.Pedidos/Listar; "
                      "kill %1 2>/dev/null; true", False, True)],
             nota="El servicio se describe SOLO, porque el servidor tiene la reflexión activada "
                  "(`reflection.Register`). ⚠️ Sin esa línea, `grpcurl` no sabe qué métodos hay y "
                  "hay que pasarle el `.proto` a mano cada vez."),
    ]

    pruebas += [
        dict(nombre="8 · El contrato cambia: ¿se cae el cliente viejo?",
             pregunta="Si el servidor añade un campo y el cliente no se actualiza, ¿qué pasa?",
             muestra=["banco/proto/v2/pedidos.proto", "banco/proto/malo/pedidos.proto",
                      "banco/cmd/contrato/main.go"],
             correr=[("cd banco && go run ./cmd/contrato", False, True)],
             nota="Añadir un campo es gratis y los campos desconocidos **se conservan**. "
                  "Reutilizar un número de campo es el bug silencioso de protobuf."),

        dict(nombre="9 · El plazo del cliente, ¿llega al servidor?",
             pregunta="El cliente da 200 ms. ¿Lo sabe el servidor ANTES de ponerse a trabajar?",
             muestra=["banco/cmd/plazos/main.go"],
             correr=[("cd banco && go run ./cmd/plazos", False, True)]),
    ]

    R.bloque("gRPC contra REST en Go — vídeo 119", pruebas, entorno=entorno)


if __name__ == "__main__":
    main()
