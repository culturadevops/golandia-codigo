package main

// Las pruebas del lanzador. Aquí no hay Docker de verdad: hay un demonio de mentira que habla
// su mismo protocolo, porque el protocolo es HTTP y eso se puede fingir en veinte líneas.
//
// Y esa es justo la gracia de que la API de Docker sea HTTP: se puede probar sin Docker.
//
//	go test ./... -count=1 -race -run Lanzador

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	lambdasdk "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// dockerFalso es un demonio de Docker de mentira. Anota lo que le piden y contesta lo que
// contestaría el de verdad.
type dockerFalso struct {
	mu       sync.Mutex
	creados  []map[string]any
	vivos    map[string]bool
	borrados []string
	fallar   bool // para probar qué pasa cuando el demonio dice que no
	sec      int
	srv      *httptest.Server
}

func nuevoDockerFalso(t *testing.T) *dockerFalso {
	t.Helper()
	d := &dockerFalso{vivos: map[string]bool{}}
	mux := http.NewServeMux()

	mux.HandleFunc("POST /{version}/containers/create", func(w http.ResponseWriter, r *http.Request) {
		d.mu.Lock()
		defer d.mu.Unlock()
		if d.fallar {
			w.WriteHeader(404)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"message": "No such image: funcion-saludar:1"})
			return
		}
		var cuerpo map[string]any
		_ = json.NewDecoder(r.Body).Decode(&cuerpo)
		cuerpo["_nombre"] = r.URL.Query().Get("name")
		d.sec++
		id := strings.Repeat("a", 8) + string(rune('0'+d.sec)) + strings.Repeat("f", 55)
		cuerpo["_id"] = id
		d.creados = append(d.creados, cuerpo)
		d.vivos[id] = false
		w.WriteHeader(201)
		_ = json.NewEncoder(w).Encode(map[string]any{"Id": id, "Warnings": []string{}})
	})

	mux.HandleFunc("POST /{version}/containers/{id}/start", func(w http.ResponseWriter, r *http.Request) {
		d.mu.Lock()
		defer d.mu.Unlock()
		d.vivos[r.PathValue("id")] = true
		w.WriteHeader(204)
	})

	mux.HandleFunc("DELETE /{version}/containers/{id}", func(w http.ResponseWriter, r *http.Request) {
		d.mu.Lock()
		defer d.mu.Unlock()
		id := r.PathValue("id")
		delete(d.vivos, id)
		d.borrados = append(d.borrados, id)
		w.WriteHeader(204)
	})

	d.srv = httptest.NewServer(mux)
	t.Cleanup(d.srv.Close)
	return d
}

func (d *dockerFalso) cuantosCreados() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.creados)
}

func (d *dockerFalso) cuantosBorrados() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.borrados)
}

func (d *dockerFalso) vivosAhora() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	n := 0
	for _, arrancado := range d.vivos {
		if arrancado {
			n++
		}
	}
	return n
}

// nubeConLanzador monta la nubecita entera con el lanzador apuntando al demonio de mentira.
func nubeConLanzador(t *testing.T, d *dockerFalso, max int,
	ocioso time.Duration) (*lambdasdk.Client, *Funciones, *Lanzador) {
	t.Helper()
	fns := NuevasFunciones()
	lan := NuevoLanzador(d.srv.Client(), d.srv.URL, "demo", "nubecita:4566", max, ocioso, nil)

	mux := http.NewServeMux()
	NuevoS3(NuevoAlmacen(), nil).Rutas(mux)
	lam := NuevoLambda(fns, nil).ConLanzador(lan)
	rutas := lam.Rutas()
	porRuta := map[string]http.Handler{}
	for _, p := range lam.Prefijos() {
		porRuta[p] = rutas
	}
	srv := httptest.NewServer(Raiz(mux, nil, porRuta, nil))
	t.Cleanup(srv.Close)

	c := lambdasdk.New(lambdasdk.Options{
		Region:       "us-east-1",
		Credentials:  credencialesFijas("prueba", "prueba"),
		BaseEndpoint: aws.String(srv.URL),
	})
	return c, fns, lan
}

func crearConImagen(t *testing.T, c *lambdasdk.Client, nombre, imagen string, segundos int32) {
	t.Helper()
	_, err := c.CreateFunction(context.Background(), &lambdasdk.CreateFunctionInput{
		FunctionName: aws.String(nombre),
		PackageType:  types.PackageTypeImage,
		Role:         aws.String("arn:aws:iam::000000000000:role/nada"),
		Timeout:      aws.Int32(segundos),
		Code:         &types.FunctionCode{ImageUri: aws.String(imagen)},
	})
	if err != nil {
		t.Fatalf("CreateFunction: %v", err)
	}
}

// 🥇 La prueba de fondo: invocar una función parada levanta un contenedor. Nadie hizo
// `docker run`.
func TestLanzadorInvocarLevantaUnContenedor(t *testing.T) {
	d := nuevoDockerFalso(t)
	c, fns, lan := nubeConLanzador(t, d, 4, 0)
	crearConImagen(t, c, "saludar", "funcion-saludar:1", 2)

	// El runtime "aparece" en cuanto Docker dice que arrancó: aquí lo simulamos.
	go func() {
		for d.vivosAhora() == 0 {
			time.Sleep(5 * time.Millisecond)
		}
		fn, err := fns.Buscar("saludar")
		if err != nil {
			return
		}
		inv, err := fns.Pedir(context.Background(), fn)
		if err == nil {
			fns.Contestar(inv.ID, Resultado{Cuerpo: []byte(`"hola"`)})
		}
	}()

	res, err := c.Invoke(context.Background(), &lambdasdk.InvokeInput{
		FunctionName: aws.String("saludar"), Payload: []byte(`{}`)})
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	if string(res.Payload) != `"hola"` {
		t.Fatalf("respuesta = %s", res.Payload)
	}
	if d.cuantosCreados() != 1 {
		t.Fatalf("creo %d contenedores, quiero 1", d.cuantosCreados())
	}
	if lan.Cuenta("saludar") != 1 {
		t.Fatalf("el lanzador cree que hay %d", lan.Cuenta("saludar"))
	}
}

// 🥇 Y lo que se le pide a Docker es exactamente esto: la imagen de la función, la red, y LA
// VARIABLE DE ENTORNO que hace que la biblioteca oficial venga a pedirnos trabajo.
func TestLanzadorLoQueLePideADocker(t *testing.T) {
	d := nuevoDockerFalso(t)
	c, fns, _ := nubeConLanzador(t, d, 4, 0)
	crearConImagen(t, c, "saludar", "funcion-saludar:1", 1)

	go func() {
		for d.vivosAhora() == 0 {
			time.Sleep(5 * time.Millisecond)
		}
		fn, _ := fns.Buscar("saludar")
		if inv, err := fns.Pedir(context.Background(), fn); err == nil {
			fns.Contestar(inv.ID, Resultado{Cuerpo: []byte(`"ok"`)})
		}
	}()
	if _, err := c.Invoke(context.Background(), &lambdasdk.InvokeInput{
		FunctionName: aws.String("saludar"), Payload: []byte(`{}`)}); err != nil {
		t.Fatalf("Invoke: %v", err)
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.creados) != 1 {
		t.Fatalf("creo %d contenedores", len(d.creados))
	}
	p := d.creados[0]
	if p["Image"] != "funcion-saludar:1" {
		t.Fatalf("imagen = %v", p["Image"])
	}
	env, _ := p["Env"].([]any)
	if len(env) != 1 || env[0] != "AWS_LAMBDA_RUNTIME_API=nubecita:4566/f/saludar" {
		t.Fatalf("el entorno del contenedor es %v, y sin esa variable el runtime no sabe a "+
			"quien pedirle trabajo", env)
	}
	host, _ := p["HostConfig"].(map[string]any)
	if host["NetworkMode"] != "demo" {
		t.Fatalf("red = %v: fuera de la red, el runtime no nos alcanza", host["NetworkMode"])
	}
	etiquetas, _ := p["Labels"].(map[string]any)
	if etiquetas[etiqueta] != "saludar" {
		t.Fatalf("sin la etiqueta %q no sabriamos cual es nuestro: %v", etiqueta, etiquetas)
	}
	if nombre, _ := p["_nombre"].(string); !strings.HasPrefix(nombre, "nubecita-saludar-") {
		t.Fatalf("nombre = %q", nombre)
	}
}

// Si ya hay un runtime esperando trabajo, NO se levanta otro. Es la mitad de la regla.
func TestLanzadorSiYaHayUnoLibreNoLevantaOtro(t *testing.T) {
	d := nuevoDockerFalso(t)
	c, fns, _ := nubeConLanzador(t, d, 4, 0)
	crearConImagen(t, c, "saludar", "funcion-saludar:1", 3)
	fn, err := fns.Buscar("saludar")
	if err != nil {
		t.Fatalf("Buscar: %v", err)
	}

	// Un runtime "levantado a mano" que ya está pidiendo trabajo, en bucle.
	ctx, parar := context.WithCancel(context.Background())
	listo := make(chan struct{})
	go func() {
		defer close(listo)
		for {
			inv, err := fns.Pedir(ctx, fn)
			if err != nil {
				return
			}
			fns.Contestar(inv.ID, Resultado{Cuerpo: []byte(`"a mano"`)})
		}
	}()
	// Esperar a que de verdad esté esperando en el canal.
	for fn.Libres() == 0 {
		time.Sleep(2 * time.Millisecond)
	}

	for i := 0; i < 3; i++ {
		if _, err := c.Invoke(context.Background(), &lambdasdk.InvokeInput{
			FunctionName: aws.String("saludar"), Payload: []byte(`{}`)}); err != nil {
			t.Fatalf("Invoke: %v", err)
		}
	}
	parar()
	<-listo

	if d.cuantosCreados() != 0 {
		t.Fatalf("levanto %d contenedores habiendo uno libre", d.cuantosCreados())
	}
}

// 🥇 El tope. Una ráfaga con la función parada levanta contenedores, pero no infinitos.
func TestLanzadorNoPasaDelTope(t *testing.T) {
	d := nuevoDockerFalso(t)
	c, _, lan := nubeConLanzador(t, d, 2, 0)
	crearConImagen(t, c, "lenta", "funcion-saludar:1", 1)

	// Nadie contesta: las invocaciones vencen en 1 s. Lo que se mide es cuántos contenedores
	// se pidieron.
	var g sync.WaitGroup
	for i := 0; i < 6; i++ {
		g.Add(1)
		go func() {
			defer g.Done()
			_, _ = c.Invoke(context.Background(), &lambdasdk.InvokeInput{
				FunctionName: aws.String("lenta"), Payload: []byte(`{}`)})
		}()
	}
	g.Wait()

	if n := d.cuantosCreados(); n != 2 {
		t.Fatalf("seis invocaciones levantaron %d contenedores y el tope era 2", n)
	}
	if lan.Cuenta("lenta") != 2 {
		t.Fatalf("el lanzador cuenta %d", lan.Cuenta("lenta"))
	}
}

// 🥇 Lo que hace que el arranque en frío VUELVA: una función que nadie llama se queda sin
// contenedores.
func TestLanzadorRecogeLoQueLlevaRatoSinUsarse(t *testing.T) {
	d := nuevoDockerFalso(t)
	c, fns, lan := nubeConLanzador(t, d, 4, 50*time.Millisecond)
	crearConImagen(t, c, "saludar", "funcion-saludar:1", 1)

	go func() {
		for d.vivosAhora() == 0 {
			time.Sleep(5 * time.Millisecond)
		}
		fn, _ := fns.Buscar("saludar")
		if inv, err := fns.Pedir(context.Background(), fn); err == nil {
			fns.Contestar(inv.ID, Resultado{Cuerpo: []byte(`"ok"`)})
		}
	}()
	if _, err := c.Invoke(context.Background(), &lambdasdk.InvokeInput{
		FunctionName: aws.String("saludar"), Payload: []byte(`{}`)}); err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	if lan.Cuenta("saludar") != 1 {
		t.Fatal("no levanto nada")
	}

	time.Sleep(120 * time.Millisecond) // más del rato ocioso
	lan.Recoger(context.Background(), fns.Listar())

	if lan.Cuenta("saludar") != 0 {
		t.Fatalf("quedan %d contenedores de una funcion que nadie llama", lan.Cuenta("saludar"))
	}
	if d.cuantosBorrados() != 1 {
		t.Fatalf("borro %d contenedores", d.cuantosBorrados())
	}
	if d.vivosAhora() != 0 {
		t.Fatal("el contenedor sigue vivo en Docker")
	}
}

// Y no recoge lo que se acaba de usar: eso sería apagar la luz con alguien dentro.
func TestLanzadorNoRecogeLoRecienUsado(t *testing.T) {
	d := nuevoDockerFalso(t)
	c, fns, lan := nubeConLanzador(t, d, 4, time.Hour)
	crearConImagen(t, c, "saludar", "funcion-saludar:1", 1)

	go func() {
		for d.vivosAhora() == 0 {
			time.Sleep(5 * time.Millisecond)
		}
		fn, _ := fns.Buscar("saludar")
		if inv, err := fns.Pedir(context.Background(), fn); err == nil {
			fns.Contestar(inv.ID, Resultado{Cuerpo: []byte(`"ok"`)})
		}
	}()
	_, _ = c.Invoke(context.Background(), &lambdasdk.InvokeInput{
		FunctionName: aws.String("saludar"), Payload: []byte(`{}`)})

	lan.Recoger(context.Background(), fns.Listar())
	if lan.Cuenta("saludar") != 1 {
		t.Fatal("apago un contenedor que se acababa de usar")
	}
}

// Borrar la función apaga sus contenedores. Si no, quedarían runtimes pidiéndole trabajo a algo
// que ya no existe.
func TestLanzadorBorrarLaFuncionApagaSusContenedores(t *testing.T) {
	d := nuevoDockerFalso(t)
	c, fns, lan := nubeConLanzador(t, d, 4, 0)
	crearConImagen(t, c, "saludar", "funcion-saludar:1", 1)

	go func() {
		for d.vivosAhora() == 0 {
			time.Sleep(5 * time.Millisecond)
		}
		fn, _ := fns.Buscar("saludar")
		if inv, err := fns.Pedir(context.Background(), fn); err == nil {
			fns.Contestar(inv.ID, Resultado{Cuerpo: []byte(`"ok"`)})
		}
	}()
	_, _ = c.Invoke(context.Background(), &lambdasdk.InvokeInput{
		FunctionName: aws.String("saludar"), Payload: []byte(`{}`)})
	if lan.Cuenta("saludar") != 1 {
		t.Fatal("no levanto nada")
	}

	if _, err := c.DeleteFunction(context.Background(), &lambdasdk.DeleteFunctionInput{
		FunctionName: aws.String("saludar")}); err != nil {
		t.Fatalf("DeleteFunction: %v", err)
	}
	if lan.Cuenta("saludar") != 0 {
		t.Fatalf("quedan %d contenedores de una funcion borrada", lan.Cuenta("saludar"))
	}
	if d.vivosAhora() != 0 {
		t.Fatal("el contenedor sigue vivo despues de borrar la funcion")
	}
}

// Si Docker dice que no —la imagen no existe—, la invocación NO se cae ahí: sigue su camino y
// vence como cualquier otra. Puede que haya un runtime tuyo a punto de llegar.
func TestLanzadorSiDockerDiceQueNoLaInvocacionSigue(t *testing.T) {
	d := nuevoDockerFalso(t)
	d.fallar = true
	c, _, lan := nubeConLanzador(t, d, 4, 0)
	crearConImagen(t, c, "saludar", "no-existe:1", 1)

	res, err := c.Invoke(context.Background(), &lambdasdk.InvokeInput{
		FunctionName: aws.String("saludar"), Payload: []byte(`{}`)})
	if err != nil {
		t.Fatalf("Invoke devolvio error de Go: %v", err)
	}
	if res.FunctionError == nil {
		t.Fatal("nadie la ejecuto y aun asi vino como exito")
	}
	if !strings.Contains(string(res.Payload), "Timedout") {
		t.Fatalf("el cuerpo no dice que vencio: %s", res.Payload)
	}
	if lan.Cuenta("saludar") != 0 {
		t.Fatal("apunto un contenedor que no llego a arrancar")
	}
}

// Una función sin imagen no levanta nada: es el modo del episodio 5, y tiene que seguir
// funcionando igual.
func TestLanzadorSinImagenNoLevantaNada(t *testing.T) {
	d := nuevoDockerFalso(t)
	c, fns, _ := nubeConLanzador(t, d, 4, 0)
	crearFuncion(t, c, "amano", 2)

	go func() {
		fn, _ := fns.Buscar("amano")
		if inv, err := fns.Pedir(context.Background(), fn); err == nil {
			fns.Contestar(inv.ID, Resultado{Cuerpo: []byte(`"a mano"`)})
		}
	}()
	res, err := c.Invoke(context.Background(), &lambdasdk.InvokeInput{
		FunctionName: aws.String("amano"), Payload: []byte(`{}`)})
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	if string(res.Payload) != `"a mano"` {
		t.Fatalf("respuesta = %s", res.Payload)
	}
	if d.cuantosCreados() != 0 {
		t.Fatalf("levanto %d contenedores para una funcion sin imagen", d.cuantosCreados())
	}
}

// Al cerrar, se apaga todo lo que levantamos. Dejarlo vivo sería dejar runtimes preguntándole
// por trabajo a un servidor que ya no existe.
func TestLanzadorApagarLosQuitaTodos(t *testing.T) {
	d := nuevoDockerFalso(t)
	c, _, lan := nubeConLanzador(t, d, 3, 0)
	crearConImagen(t, c, "lenta", "funcion-saludar:1", 1)

	var g sync.WaitGroup
	for i := 0; i < 3; i++ {
		g.Add(1)
		go func() {
			defer g.Done()
			_, _ = c.Invoke(context.Background(), &lambdasdk.InvokeInput{
				FunctionName: aws.String("lenta"), Payload: []byte(`{}`)})
		}()
	}
	g.Wait()
	if lan.Cuenta("lenta") != 3 {
		t.Fatalf("levanto %d y esperaba 3", lan.Cuenta("lenta"))
	}

	lan.Apagar(context.Background(), "")
	if lan.Cuenta("lenta") != 0 {
		t.Fatal("quedan contenedores apuntados despues de apagar")
	}
	if d.vivosAhora() != 0 {
		t.Fatalf("quedan %d vivos en Docker", d.vivosAhora())
	}
	arrancados, apagados := lan.Cifras()
	if arrancados != 3 || apagados != 3 {
		t.Fatalf("arrancados=%d apagados=%d", arrancados, apagados)
	}
}

// El cliente de Docker habla por un SOCKET, no por la red. Esto comprueba que la conexión se
// abre por donde debe.
func TestLanzadorElClienteHablaPorElSocket(t *testing.T) {
	c := ClienteDocker("/var/run/docker.sock")
	tr, ok := c.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("el transporte es %T", c.Transport)
	}
	if tr.DialContext == nil {
		t.Fatal("sin DialContext propio, esto se iria a la red y no al socket")
	}
	// Y el intento falla con «no existe ese fichero» y no con «no hay ruta al host»: es la
	// prueba de que ni siquiera intenta resolver un nombre.
	_, err := tr.DialContext(context.Background(), "tcp", "docker:80")
	if err == nil {
		t.Skip("hay un socket de Docker aqui: nada que comprobar")
	}
	if strings.Contains(err.Error(), "lookup") {
		t.Fatalf("intento resolver un nombre: %v", err)
	}
}
