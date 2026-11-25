package main

// Las pruebas de Lambda. Quien INVOCA es el SDK oficial de AWS; quien EJECUTA es el runtime
// casero de aquí abajo, que existe para enseñar una cosa: un runtime de Lambda son veinte
// líneas. `aws-lambda-go` hace exactamente esto y poco más.
//
//	go test ./... -count=1 -race -run Lambda

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	lambdasdk "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

func nubeConLambda(t *testing.T) (*lambdasdk.Client, string) {
	t.Helper()
	mux := http.NewServeMux()
	NuevoS3(NuevoAlmacen(), nil).Rutas(mux)
	lam := NuevoLambda(NuevasFunciones(), nil)
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
	return c, srv.URL
}

func crearFuncion(t *testing.T, c *lambdasdk.Client, nombre string, segundos int32) {
	t.Helper()
	_, err := c.CreateFunction(context.Background(), &lambdasdk.CreateFunctionInput{
		FunctionName: aws.String(nombre),
		Runtime:      types.RuntimeProvidedal2023,
		Handler:      aws.String("bootstrap"),
		Role:         aws.String("arn:aws:iam::000000000000:role/nada"),
		Timeout:      aws.Int32(segundos),
		Code:         &types.FunctionCode{ZipFile: []byte("no hay zip: el codigo esta en el runtime")},
	})
	if err != nil {
		t.Fatalf("CreateFunction: %v", err)
	}
}

// 🥇 runtimeCasero es un runtime de Lambda ENTERO. Pide trabajo, lo hace, devuelve el
// resultado, y vuelta a empezar. No hay nada más. Cuando ves esto, `aws-lambda-go` deja de
// parecer magia y pasa a ser un bucle con dos llamadas HTTP.
func runtimeCasero(t *testing.T, base, funcion string,
	manejador func([]byte) ([]byte, error)) (parar func()) {
	t.Helper()
	ctx, cancelar := context.WithCancel(context.Background())
	fin := make(chan struct{})

	go func() {
		defer close(fin)
		api := base + "/f/" + funcion + "/2018-06-01/runtime"
		for {
			// 1. «¿Hay algo para mí?». Esta petición se queda abierta hasta que lo hay.
			pide, _ := http.NewRequestWithContext(ctx, http.MethodGet, api+"/invocation/next", nil)
			res, err := http.DefaultClient.Do(pide)
			if err != nil {
				return // nos han parado, o se cayó el servidor
			}
			id := res.Header.Get("Lambda-Runtime-Aws-Request-Id")
			evento, _ := io.ReadAll(res.Body)
			res.Body.Close()

			// 2. Tu código.
			salida, err := manejador(evento)
			ruta := "/response"
			if err != nil {
				ruta = "/error"
				salida, _ = json.Marshal(map[string]string{
					"errorType": "Error", "errorMessage": err.Error()})
			}

			// 3. Devolver el resultado.
			da, _ := http.NewRequestWithContext(ctx, http.MethodPost,
				api+"/invocation/"+id+ruta, bytes.NewReader(salida))
			r2, err := http.DefaultClient.Do(da)
			if err != nil {
				return
			}
			r2.Body.Close()
		}
	}()

	return func() { cancelar(); <-fin }
}

func esperarUnPoco() { time.Sleep(50 * time.Millisecond) }

// La prueba de fondo: el SDK oficial invoca, nuestro servidor hace de Lambda, y el runtime
// contesta. Nadie de los tres sabe que los otros dos son de mentira.
func TestLambdaInvocarDeVerdad(t *testing.T) {
	c, base := nubeConLambda(t)
	crearFuncion(t, c, "saludar", 5)

	parar := runtimeCasero(t, base, "saludar", func(evento []byte) ([]byte, error) {
		var e struct {
			Nombre string `json:"nombre"`
		}
		_ = json.Unmarshal(evento, &e)
		return json.Marshal(map[string]string{"saludo": "hola, " + e.Nombre})
	})
	defer parar()
	esperarUnPoco()

	res, err := c.Invoke(context.Background(), &lambdasdk.InvokeInput{
		FunctionName: aws.String("saludar"),
		Payload:      []byte(`{"nombre":"Ada"}`),
	})
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("codigo = %d, quiero 200", res.StatusCode)
	}
	if res.FunctionError != nil {
		t.Fatalf("la funcion dio error: %s", string(res.Payload))
	}
	if got := string(res.Payload); !strings.Contains(got, `"hola, Ada"`) {
		t.Fatalf("respuesta = %s", got)
	}
}

// 🥇 EL ARRANQUE EN FRÍO, MEDIDO. El invocador llega ANTES que el runtime, y se queda esperando
// en el canal. Eso es todo lo que es un arranque en frío: no es una penalización que te cobran,
// es que al otro lado todavía no hay nadie.
func TestLambdaElArranqueEnFrioEsEsperar(t *testing.T) {
	c, base := nubeConLambda(t)
	crearFuncion(t, c, "tarde", 5)

	// El runtime tarda 250 ms en aparecer: es el contenedor arrancando.
	//
	// ⚠️ El «parar» viaja por un canal y no por una variable compartida. La primera version de
	// esta prueba lo guardaba en una `var parar func()` desde la gorrutina, y el detector de
	// carreras la tumbo: dos hilos tocando la misma variable sin nada en medio. Una prueba con
	// una carrera dentro no demuestra nada, aunque pase.
	paradas := make(chan func(), 1)
	go func() {
		time.Sleep(250 * time.Millisecond)
		paradas <- runtimeCasero(t, base, "tarde", func([]byte) ([]byte, error) {
			return []byte(`"ya estoy"`), nil
		})
	}()

	empezo := time.Now()
	res, err := c.Invoke(context.Background(), &lambdasdk.InvokeInput{
		FunctionName: aws.String("tarde"),
		Payload:      []byte(`{}`),
	})
	tardo := time.Since(empezo)
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	if string(res.Payload) != `"ya estoy"` {
		t.Fatalf("respuesta = %s", res.Payload)
	}
	if tardo < 200*time.Millisecond {
		t.Fatalf("contesto en %v: no espero al runtime, asi que no lo esperaba de verdad", tardo)
	}
	t.Logf("el invocador espero %v a que apareciera el runtime", tardo.Round(time.Millisecond))
	(<-paradas)()
}

// 🥇 LA TRAMPA MÁS CARA DE LAMBDA: la función falla y la llamada contesta 200. El error viaja
// en una cabecera, no en el código HTTP. Si tu cliente solo mira el código, cada fallo de tu
// función te parece un éxito.
func TestLambdaSiLaFuncionFallaLlegaUn200(t *testing.T) {
	c, base := nubeConLambda(t)
	crearFuncion(t, c, "revienta", 5)

	parar := runtimeCasero(t, base, "revienta", func([]byte) ([]byte, error) {
		return nil, errors.New("no encuentro el registro 7")
	})
	defer parar()
	esperarUnPoco()

	res, err := c.Invoke(context.Background(), &lambdasdk.InvokeInput{
		FunctionName: aws.String("revienta"),
		Payload:      []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Invoke devolvio error de Go, y no deberia: %v", err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("codigo = %d, y Lambda contesta 200 aunque tu funcion falle", res.StatusCode)
	}
	if res.FunctionError == nil {
		t.Fatal("no vino X-Amz-Function-Error: el fallo de la funcion se ha perdido")
	}
	if !strings.Contains(string(res.Payload), "registro 7") {
		t.Fatalf("el cuerpo no trae el error de la funcion: %s", res.Payload)
	}
}

// Un runtime atiende DE UNA EN UNA. Esto no es un límite nuestro: es el modelo de Lambda, y es
// la razón de que «más concurrencia» en Lambda signifique más contenedores y no más hilos.
func TestLambdaUnRuntimeAtiendeDeUnaEnUna(t *testing.T) {
	c, base := nubeConLambda(t)
	crearFuncion(t, c, "lenta", 10)

	parar := runtimeCasero(t, base, "lenta", func([]byte) ([]byte, error) {
		time.Sleep(150 * time.Millisecond)
		return []byte(`"hecho"`), nil
	})
	defer parar()
	esperarUnPoco()

	empezo := time.Now()
	var g sync.WaitGroup
	for i := 0; i < 3; i++ {
		g.Add(1)
		go func() {
			defer g.Done()
			if _, err := c.Invoke(context.Background(), &lambdasdk.InvokeInput{
				FunctionName: aws.String("lenta"), Payload: []byte(`{}`)}); err != nil {
				t.Errorf("Invoke: %v", err)
			}
		}()
	}
	g.Wait()
	tardo := time.Since(empezo)

	// Tres invocaciones de 150 ms con UN runtime tardan ~450, no ~150.
	if tardo < 400*time.Millisecond {
		t.Fatalf("tres invocaciones tardaron %v: un solo runtime no puede hacerlas a la vez", tardo)
	}
	t.Logf("3 invocaciones · 1 runtime · %v", tardo.Round(time.Millisecond))
}

// Y con TRES runtimes, las mismas tres invocaciones caben en una tanda. Más contenedores, no
// más hilos: exactamente lo que hace Amazon cuando escala tu función.
func TestLambdaTresRuntimesVanALaVez(t *testing.T) {
	c, base := nubeConLambda(t)
	crearFuncion(t, c, "lenta", 10)

	for i := 0; i < 3; i++ {
		parar := runtimeCasero(t, base, "lenta", func([]byte) ([]byte, error) {
			time.Sleep(150 * time.Millisecond)
			return []byte(`"hecho"`), nil
		})
		defer parar()
	}
	esperarUnPoco()

	empezo := time.Now()
	var g sync.WaitGroup
	for i := 0; i < 3; i++ {
		g.Add(1)
		go func() {
			defer g.Done()
			if _, err := c.Invoke(context.Background(), &lambdasdk.InvokeInput{
				FunctionName: aws.String("lenta"), Payload: []byte(`{}`)}); err != nil {
				t.Errorf("Invoke: %v", err)
			}
		}()
	}
	g.Wait()
	tardo := time.Since(empezo)

	if tardo > 400*time.Millisecond {
		t.Fatalf("tres invocaciones con tres runtimes tardaron %v: no fueron en paralelo", tardo)
	}
	t.Logf("3 invocaciones · 3 runtimes · %v", tardo.Round(time.Millisecond))
}

// Si NADIE recoge la invocación, vence. Es el «Task timed out» de Amazon, y también llega
// con un 200 y la cabecera de error.
func TestLambdaSiNadieLaRecogeVence(t *testing.T) {
	c, _ := nubeConLambda(t)
	crearFuncion(t, c, "huerfana", 1) // un segundo de plazo

	empezo := time.Now()
	res, err := c.Invoke(context.Background(), &lambdasdk.InvokeInput{
		FunctionName: aws.String("huerfana"), Payload: []byte(`{}`)})
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	if res.FunctionError == nil {
		t.Fatal("nadie ejecuto nada y aun asi vino como exito")
	}
	if !strings.Contains(string(res.Payload), "Timedout") {
		t.Fatalf("el cuerpo no dice que vencio: %s", res.Payload)
	}
	if d := time.Since(empezo); d < 900*time.Millisecond {
		t.Fatalf("vencio en %v y el plazo era de 1 s: no espero el plazo", d)
	}
}

// La invocación asíncrona contesta 202 y NO trae resultado. No es que se pierda: es que cuando
// contesta todavía no existe.
func TestLambdaAsincronaContesta202(t *testing.T) {
	c, base := nubeConLambda(t)
	crearFuncion(t, c, "aparte", 5)

	visto := make(chan []byte, 1)
	parar := runtimeCasero(t, base, "aparte", func(e []byte) ([]byte, error) {
		visto <- e
		return []byte(`"da igual, nadie mira"`), nil
	})
	defer parar()
	esperarUnPoco()

	res, err := c.Invoke(context.Background(), &lambdasdk.InvokeInput{
		FunctionName:   aws.String("aparte"),
		InvocationType: types.InvocationTypeEvent,
		Payload:        []byte(`{"cual":"asincrona"}`),
	})
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	if res.StatusCode != 202 {
		t.Fatalf("codigo = %d, quiero 202", res.StatusCode)
	}
	if len(res.Payload) != 0 {
		t.Fatalf("una invocacion asincrona no devuelve resultado, y trajo %s", res.Payload)
	}
	select {
	case e := <-visto:
		if !strings.Contains(string(e), "asincrona") {
			t.Fatalf("al runtime le llego %s", e)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("el runtime nunca vio el evento asincrono")
	}
}

// El runtime recibe el identificador y el PLAZO. De esa cabecera sale el `ctx.Deadline()` que
// le llega a tu manejador, y que casi nadie mira.
func TestLambdaElRuntimeRecibeElPlazo(t *testing.T) {
	c, base := nubeConLambda(t)
	crearFuncion(t, c, "conplazo", 7)

	pedido := make(chan *http.Response, 1)
	go func() {
		res, err := http.Get(base + "/f/conplazo/2018-06-01/runtime/invocation/next")
		if err == nil {
			pedido <- res
		}
	}()
	go func() {
		esperarUnPoco()
		_, _ = c.Invoke(context.Background(), &lambdasdk.InvokeInput{
			FunctionName: aws.String("conplazo"), Payload: []byte(`{"a":1}`)})
	}()

	select {
	case res := <-pedido:
		defer res.Body.Close()
		if res.Header.Get("Lambda-Runtime-Aws-Request-Id") == "" {
			t.Fatal("sin identificador de peticion, el runtime no sabe a que contestar")
		}
		ms, err := strconv.ParseInt(res.Header.Get("Lambda-Runtime-Deadline-Ms"), 10, 64)
		if err != nil {
			t.Fatalf("el plazo no es un numero: %v", err)
		}
		queda := time.UnixMilli(ms).Sub(time.Now().UTC())
		if queda < 5*time.Second || queda > 8*time.Second {
			t.Fatalf("el plazo dice que quedan %v y la funcion tenia 7 s", queda)
		}
		if !strings.Contains(res.Header.Get("Lambda-Runtime-Invoked-Function-Arn"), "conplazo") {
			t.Fatal("el ARN invocado no menciona la funcion")
		}
		// Y se contesta, para que el invocador no se quede colgado los 7 s del plazo. Un
		// runtime que lee la invocacion y no contesta deja al de enfrente esperando: es
		// justo lo que pasa cuando tu funcion se queda pillada.
		id := res.Header.Get("Lambda-Runtime-Aws-Request-Id")
		fin, err := http.Post(base+"/f/conplazo/2018-06-01/runtime/invocation/"+id+"/response",
			"application/json", strings.NewReader(`"listo"`))
		if err != nil {
			t.Fatalf("contestar: %v", err)
		}
		fin.Body.Close()
	case <-time.After(3 * time.Second):
		t.Fatal("el runtime no recibio nada")
	}
}

// Contestar dos veces a la misma invocación no cuela: la segunda ya no está en vuelo.
func TestLambdaNoSePuedeContestarDosVeces(t *testing.T) {
	c, base := nubeConLambda(t)
	crearFuncion(t, c, "unavez", 5)

	ids := make(chan string, 1)
	go func() {
		res, err := http.Get(base + "/f/unavez/2018-06-01/runtime/invocation/next")
		if err != nil {
			return
		}
		defer res.Body.Close()
		_, _ = io.Copy(io.Discard, res.Body)
		ids <- res.Header.Get("Lambda-Runtime-Aws-Request-Id")
	}()
	go func() {
		esperarUnPoco()
		_, _ = c.Invoke(context.Background(), &lambdasdk.InvokeInput{
			FunctionName: aws.String("unavez"), Payload: []byte(`{}`)})
	}()

	id := <-ids
	url := base + "/f/unavez/2018-06-01/runtime/invocation/" + id + "/response"
	uno, err := http.Post(url, "application/json", strings.NewReader(`"primera"`))
	if err != nil {
		t.Fatalf("primera respuesta: %v", err)
	}
	uno.Body.Close()
	if uno.StatusCode != 202 {
		t.Fatalf("la primera respuesta dio %d, quiero 202", uno.StatusCode)
	}

	dos, err := http.Post(url, "application/json", strings.NewReader(`"segunda"`))
	if err != nil {
		t.Fatalf("segunda respuesta: %v", err)
	}
	dos.Body.Close()
	if dos.StatusCode != 400 {
		t.Fatalf("la segunda respuesta dio %d, y esa invocacion ya no estaba en vuelo",
			dos.StatusCode)
	}
}

// Invocar algo que no existe da el error tipado que el SDK sabe reconocer.
func TestLambdaFuncionQueNoExiste(t *testing.T) {
	c, _ := nubeConLambda(t)
	_, err := c.Invoke(context.Background(), &lambdasdk.InvokeInput{
		FunctionName: aws.String("fantasma"), Payload: []byte(`{}`)})
	if err == nil {
		t.Fatal("invoco una funcion que no existe y no se quejo")
	}
	var noHay *types.ResourceNotFoundException
	if !errors.As(err, &noHay) {
		t.Fatalf("el error no llego como ResourceNotFoundException: %v", err)
	}
}

// Crear, listar, mirar y borrar: el ciclo entero con el SDK oficial.
func TestLambdaCicloDeVidaDeLaFuncion(t *testing.T) {
	c, _ := nubeConLambda(t)
	crearFuncion(t, c, "una", 5)
	crearFuncion(t, c, "otra", 5)

	lista, err := c.ListFunctions(context.Background(), &lambdasdk.ListFunctionsInput{})
	if err != nil {
		t.Fatalf("ListFunctions: %v", err)
	}
	if len(lista.Functions) != 2 {
		t.Fatalf("hay %d funciones, quiero 2", len(lista.Functions))
	}

	det, err := c.GetFunction(context.Background(), &lambdasdk.GetFunctionInput{
		FunctionName: aws.String("una")})
	if err != nil {
		t.Fatalf("GetFunction: %v", err)
	}
	if aws.ToString(det.Configuration.FunctionArn) != arnDe("una") {
		t.Fatalf("ARN = %s", aws.ToString(det.Configuration.FunctionArn))
	}
	if det.Configuration.State != types.StateActive {
		t.Fatalf("estado = %s, quiero Active", det.Configuration.State)
	}

	if _, err := c.DeleteFunction(context.Background(), &lambdasdk.DeleteFunctionInput{
		FunctionName: aws.String("una")}); err != nil {
		t.Fatalf("DeleteFunction: %v", err)
	}
	lista, _ = c.ListFunctions(context.Background(), &lambdasdk.ListFunctionsInput{})
	if len(lista.Functions) != 1 {
		t.Fatalf("despues de borrar quedan %d", len(lista.Functions))
	}
}

// Crear dos veces la misma función tampoco cuela.
func TestLambdaNoSePuedeCrearDosVeces(t *testing.T) {
	c, _ := nubeConLambda(t)
	crearFuncion(t, c, "repe", 5)
	_, err := c.CreateFunction(context.Background(), &lambdasdk.CreateFunctionInput{
		FunctionName: aws.String("repe"),
		Runtime:      types.RuntimeProvidedal2023,
		Handler:      aws.String("bootstrap"),
		Role:         aws.String("arn:aws:iam::000000000000:role/nada"),
		Code:         &types.FunctionCode{ZipFile: []byte("x")},
	})
	if err == nil {
		t.Fatal("creo dos veces la misma funcion")
	}
	var choque *types.ResourceConflictException
	if !errors.As(err, &choque) {
		t.Fatalf("el error no llego como ResourceConflictException: %v", err)
	}
}
