package main

// El servicio Lambda. Y este es distinto a todos los anteriores, porque Lambda no es UNA API:
// son DOS, y hablan con gente distinta.
//
//	1) el plano de datos     ← lo llama TU codigo, con el SDK de AWS
//	   POST /2015-03-31/functions/saludar/invocations
//
//	2) la API del runtime    ← la llama la FUNCION, desde dentro de su contenedor
//	   GET  /2018-06-01/runtime/invocation/next
//	   POST /2018-06-01/runtime/invocation/{id}/response
//
// La segunda es la que casi nadie ha visto, y es la que explica Lambda entera. Tu funcion no
// "recibe" peticiones: las va a BUSCAR, en un bucle, a una direccion HTTP que Lambda le pone en
// la variable de entorno AWS_LAMBDA_RUNTIME_API. Eso es todo el misterio. `aws-lambda-go`, la
// biblioteca oficial, es un bucle alrededor de ese GET.
//
// 🥇 Y por eso esto se puede emular de verdad: si nosotros contestamos a esa API, la biblioteca
// OFICIAL de Amazon se conecta a nubecita sin cambiar una linea de codigo. No hay parche, no
// hay «modo de pruebas»: es el mismo binario que subirias a produccion.
//
// Lo que se implementa:
//
//	CreateFunction · ListFunctions · GetFunction · DeleteFunction · Invoke
//	API del runtime: next · response · error · init/error
//
// ⛔ Lo que NO: subir un zip de verdad (el codigo vive en el contenedor del runtime), capas,
// versiones y alias, disparadores, VPC, ni el registro en CloudWatch.

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"
)

type ServicioLambda struct {
	fn       *Funciones
	lan      *Lanzador     // puede ser nil: entonces el runtime lo levantas tu, como en el episodio 5
	disp     *Disparadores // puede ser nil: entonces nadie invoca la funcion sola
	registro func(formato string, args ...any)
}

func NuevoLambda(f *Funciones, registro func(string, ...any)) *ServicioLambda {
	if registro == nil {
		registro = func(string, ...any) {}
	}
	return &ServicioLambda{fn: f, registro: registro}
}

// ConLanzador le da al servicio la capacidad de levantar los contenedores el solo. Sin esto,
// nubecita reparte trabajo pero no ejecuta nada por su cuenta.
func (s *ServicioLambda) ConLanzador(l *Lanzador) *ServicioLambda {
	s.lan = l
	return s
}

// ConDisparadores enchufa los mapeos de origen de eventos (episodio 8).
func (s *ServicioLambda) ConDisparadores(d *Disparadores) *ServicioLambda {
	s.disp = d
	return s
}

// Rutas devuelve el mux de Lambda. ⚠️ Es un mux SUYO, no el de S3: S3 se queda con la raiz
// entera con comodines, y meter aqui dentro `/2015-03-31/functions` hace que el mux de Go se
// niegue a arrancar. Ver `enrutador.go`.
func (s *ServicioLambda) Rutas() *http.ServeMux {
	mux := http.NewServeMux()
	// El plano de datos: lo que llama el SDK de AWS.
	mux.HandleFunc("POST /2015-03-31/functions", s.crear)
	mux.HandleFunc("GET /2015-03-31/functions", s.listar)
	mux.HandleFunc("GET /2015-03-31/functions/{nombre}", s.obtener)
	mux.HandleFunc("DELETE /2015-03-31/functions/{nombre}", s.borrar)
	mux.HandleFunc("POST /2015-03-31/functions/{nombre}/invocations", s.invocar)

	// La API del runtime: lo que llama la funcion desde dentro.
	//
	// ⚠️ En Amazon esta API NO lleva el nombre de la funcion en la ruta, porque cada
	// contenedor sirve a UNA funcion y ya lo sabe. Aqui una sola nubecita atiende a varias, y
	// hace falta distinguirlas. El truco sale gratis: `aws-lambda-go` construye la URL
	// pegando AWS_LAMBDA_RUNTIME_API delante, tal cual, asi que basta con arrancar el
	// contenedor con
	//
	//	AWS_LAMBDA_RUNTIME_API=nubecita:4566/f/saludar
	//
	// y la biblioteca oficial pide a /f/saludar/2018-06-01/runtime/... sin enterarse.
	mux.HandleFunc("GET /f/{nombre}/2018-06-01/runtime/invocation/next", s.siguiente)
	mux.HandleFunc("POST /f/{nombre}/2018-06-01/runtime/invocation/{id}/response", s.respuesta)
	mux.HandleFunc("POST /f/{nombre}/2018-06-01/runtime/invocation/{id}/error", s.errorDeFuncion)
	mux.HandleFunc("POST /f/{nombre}/2018-06-01/runtime/init/error", s.errorDeArranque)

	// Y los disparadores, que tambien son parte de la API de Lambda.
	s.rutasDisparadores(mux)
	return mux
}

// Prefijos es lo que hay que mandarle a `Raiz` para que estas rutas lleguen aqui y no a S3.
func (s *ServicioLambda) Prefijos() []string { return []string{"/2015-03-31/", "/f/"} }

// ── el plano de datos ────────────────────────────────────────────────────

type cuerpoCrear struct {
	FunctionName string `json:"FunctionName"`
	Runtime      string `json:"Runtime"`
	Handler      string `json:"Handler"`
	MemorySize   int32  `json:"MemorySize"`
	Timeout      int32  `json:"Timeout"`
	PackageType  string `json:"PackageType"`
	Code         struct {
		ImageUri string `json:"ImageUri"`
	} `json:"Code"`
	Environment struct {
		Variables map[string]string `json:"Variables"`
	} `json:"Environment"`
}

type fichaFuncion struct {
	FunctionName string `json:"FunctionName"`
	FunctionArn  string `json:"FunctionArn"`
	Runtime      string `json:"Runtime"`
	Handler      string `json:"Handler"`
	MemorySize   int32  `json:"MemorySize"`
	Timeout      int32  `json:"Timeout"`
	State        string `json:"State"`
	PackageType  string `json:"PackageType,omitempty"`
	LastModified string `json:"LastModified"`
	Version      string `json:"Version"`
}

func paqueteDe(f *Funcion) string {
	if f.Imagen != "" {
		return "Image"
	}
	return "Zip"
}

func arnDe(nombre string) string {
	return "arn:aws:lambda:us-east-1:000000000000:function:" + nombre
}

func ficha(f *Funcion) fichaFuncion {
	return fichaFuncion{
		FunctionName: f.Nombre, FunctionArn: arnDe(f.Nombre),
		Runtime: f.Runtime, Handler: f.Manejador,
		MemorySize: f.Memoria, Timeout: f.Segundos,
		State:        "Active",
		PackageType:  paqueteDe(f),
		LastModified: f.Creada.Format("2006-01-02T15:04:05.000-0700"),
		Version:      "$LATEST",
	}
}

func (s *ServicioLambda) crear(w http.ResponseWriter, r *http.Request) {
	var c cuerpoCrear
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		s.fallo(w, 400, "InvalidParameterValueException", err.Error())
		return
	}
	if c.MemorySize == 0 {
		c.MemorySize = 128
	}
	if c.Timeout == 0 {
		c.Timeout = 3 // el mismo valor por defecto que Amazon, y el que mas sustos da
	}
	// Lambda admite dos formas de empaquetar: un zip o una IMAGEN de contenedor. Aqui la
	// segunda es la natural, porque el codigo de la funcion vive en su contenedor y nubecita
	// no ejecuta nada: solo reparte trabajo.
	if c.PackageType == "" {
		c.PackageType = "Zip"
	}
	if c.PackageType == "Image" && c.Runtime == "" {
		c.Runtime = "imagen:" + c.Code.ImageUri
	}
	s.registro("POST /2015-03-31/functions [CreateFunction %s]", c.FunctionName)

	f, err := s.fn.Crear(c.FunctionName, c.Runtime, c.Handler, c.MemorySize, c.Timeout)
	if errors.Is(err, errYaExisteFuncion) {
		s.fallo(w, 409, "ResourceConflictException", "ya existe una funcion "+c.FunctionName)
		return
	}
	// La imagen es lo que el lanzador necesita para poder ejecutarla. Se guarda al crear y ya
	// no cambia: aqui no hay versiones.
	f.Imagen = c.Code.ImageUri
	f.Entorno = c.Environment.Variables
	s.json(w, 201, ficha(f))
}

func (s *ServicioLambda) listar(w http.ResponseWriter, r *http.Request) {
	s.registro("GET /2015-03-31/functions [ListFunctions]")
	fichas := []fichaFuncion{}
	for _, f := range s.fn.Listar() {
		fichas = append(fichas, ficha(f))
	}
	s.json(w, 200, map[string]any{"Functions": fichas})
}

func (s *ServicioLambda) obtener(w http.ResponseWriter, r *http.Request) {
	nombre := r.PathValue("nombre")
	s.registro("GET /2015-03-31/functions/%s [GetFunction]", nombre)
	f, err := s.fn.Buscar(nombre)
	if err != nil {
		s.noExiste(w, nombre)
		return
	}
	// GetFunction contesta con dos bloques; el SDK espera esta forma exacta.
	s.json(w, 200, map[string]any{
		"Configuration": ficha(f),
		"Code":          map[string]string{"RepositoryType": "Nubecita", "Location": ""},
	})
}

func (s *ServicioLambda) borrar(w http.ResponseWriter, r *http.Request) {
	nombre := r.PathValue("nombre")
	s.registro("DELETE /2015-03-31/functions/%s [DeleteFunction]", nombre)
	if err := s.fn.Borrar(nombre); err != nil {
		s.noExiste(w, nombre)
		return
	}
	// Borrar la funcion sin apagar sus contenedores dejaria runtimes pidiendo trabajo a una
	// funcion que ya no existe. Nadie los pararia nunca.
	s.lan.Apagar(r.Context(), nombre)
	w.WriteHeader(http.StatusNoContent)
}

// invocar es el corazon del plano de datos.
//
// 🥇 Y aqui esta la trampa mas cara de Lambda: si TU CODIGO falla, esta llamada contesta
// 200 OK. Doscientos. El fallo no viaja en el codigo de estado, sino en una CABECERA,
// `X-Amz-Function-Error`, y el cuerpo trae el error en vez del resultado.
//
// El motivo tiene sentido: para Lambda, «tu funcion se ejecuto y devolvio un error» es una
// invocacion CORRECTA. La API hizo su trabajo. Pero si tu cliente solo mira el codigo HTTP —y
// casi todos los clientes escritos a mano solo miran eso— vas a tratar como exito cada fallo
// de tu funcion, y no lo vas a ver hasta que alguien pregunte por que faltan datos.
func (s *ServicioLambda) invocar(w http.ResponseWriter, r *http.Request) {
	nombre := r.PathValue("nombre")
	f, err := s.fn.Buscar(nombre)
	if err != nil {
		s.noExiste(w, nombre)
		return
	}
	evento, err := io.ReadAll(r.Body)
	if err != nil {
		s.fallo(w, 400, "InvalidRequestContentException", err.Error())
		return
	}
	if len(evento) == 0 {
		evento = []byte("{}")
	}
	tipo := r.Header.Get("X-Amz-Invocation-Type")
	if tipo == "" {
		tipo = "RequestResponse"
	}
	s.registro("POST /2015-03-31/functions/%s/invocations [%s] %d byte(s)",
		nombre, tipo, len(evento))

	// DryRun: comprobar que la funcion existe y no ejecutar nada.
	if tipo == "DryRun" {
		w.Header().Set("X-Amz-Executed-Version", "$LATEST")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// ⚠️ Event = asincrono. Amazon contesta 202 en cuanto ACEPTA el evento, sin esperar a que
	// nadie lo ejecute. Por eso una invocacion asincrona nunca te devuelve el resultado: no es
	// que se pierda, es que cuando contesta todavia no existe.
	if tipo == "Event" {
		go func() {
			ctx, cancelar := contextoConLimite(f)
			defer cancelar()
			s.lan.Asegurar(ctx, f)
			if _, err := s.fn.Invocar(ctx, f, evento); err != nil {
				s.registro("  -> el evento asincrono de %s no se pudo entregar: %v", nombre, err)
			}
		}()
		w.WriteHeader(http.StatusAccepted)
		return
	}

	// El contexto de la peticion HTTP manda: si el cliente cuelga, dejamos de esperar. Y
	// encima le ponemos el plazo de la funcion.
	ctx, cancelar := contextoConLimiteDe(r, f)
	defer cancelar()

	// 🥇 Aqui, y solo aqui, se decide si hace falta levantar un contenedor. Antes de la
	// invocacion, no despues: si no hay nadie escuchando, hay que traer a alguien.
	s.lan.Asegurar(ctx, f)

	res, err := s.fn.Invocar(ctx, f, evento)
	if err != nil {
		// El plazo se agoto y nadie contesto. En Amazon esto es el famoso «Task timed out».
		s.registro("  -> 200 sin respuesta del runtime (%v)", err)
		w.Header().Set("X-Amz-Function-Error", "Unhandled")
		s.json(w, 200, map[string]string{
			"errorType":    "Sandbox.Timedout",
			"errorMessage": "la funcion no contesto en " + strconv.Itoa(int(f.Segundos)) + " s",
		})
		return
	}

	if res.Fallo {
		s.registro("  -> 200 con X-Amz-Function-Error (la funcion devolvio un error)")
		w.Header().Set("X-Amz-Function-Error", "Unhandled")
	} else {
		s.registro("  -> 200 %d byte(s)", len(res.Cuerpo))
	}
	w.Header().Set("X-Amz-Executed-Version", "$LATEST")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(res.Cuerpo)
}

// ── la API del runtime ───────────────────────────────────────────────────

// siguiente es el sondeo largo: la funcion pregunta «¿hay algo para mi?» y esta peticion se
// queda ABIERTA, sin contestar, hasta que la hay.
//
// ⚠️ Sin tiempo limite a proposito. Una conexion que no se cierra parece un fallo, pero es el
// diseño: es mas barato tener una peticion HTTP parada que preguntar cada 100 ms para siempre.
// Por eso `aws-lambda-go` crea su cliente con Timeout: 0.
func (s *ServicioLambda) siguiente(w http.ResponseWriter, r *http.Request) {
	nombre := r.PathValue("nombre")
	f, err := s.fn.Buscar(nombre)
	if err != nil {
		s.noExiste(w, nombre)
		return
	}
	s.registro("GET /f/%s/.../invocation/next [el runtime pide trabajo]", nombre)

	inv, err := s.fn.Pedir(r.Context(), f)
	if err != nil {
		// El runtime se fue (lo pararon, se cayo). No es un error nuestro.
		return
	}

	// Estas cabeceras son el contrato. `aws-lambda-go` las lee para armar el `context.Context`
	// que le llega a tu manejador: de aqui salen `lambdacontext.FromContext(ctx)` y el
	// `ctx.Deadline()` que casi nadie mira y que es lo unico que te avisa de que te queda poco.
	h := w.Header()
	h.Set("Lambda-Runtime-Aws-Request-Id", inv.ID)
	h.Set("Lambda-Runtime-Deadline-Ms", strconv.FormatInt(inv.Limite.UnixMilli(), 10))
	h.Set("Lambda-Runtime-Invoked-Function-Arn", arnDe(nombre))
	h.Set("Lambda-Runtime-Trace-Id", "Root=1-00000000-000000000000000000000000")
	h.Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(inv.Evento)
	s.registro("  -> entregada la invocacion %s", inv.ID)
}

func (s *ServicioLambda) respuesta(w http.ResponseWriter, r *http.Request) {
	s.entregar(w, r, false)
}

func (s *ServicioLambda) errorDeFuncion(w http.ResponseWriter, r *http.Request) {
	s.entregar(w, r, true)
}

func (s *ServicioLambda) entregar(w http.ResponseWriter, r *http.Request, fallo bool) {
	id := r.PathValue("id")
	cuerpo, err := io.ReadAll(r.Body)
	if err != nil {
		s.fallo(w, 400, "InvalidRequestContentException", err.Error())
		return
	}
	que := "response"
	if fallo {
		que = "error"
	}
	s.registro("POST .../invocation/%s/%s %d byte(s)", id, que, len(cuerpo))

	if !s.fn.Contestar(id, Resultado{Cuerpo: cuerpo, Fallo: fallo}) {
		// Llego tarde: ya nadie esperaba. Amazon contesta 400 con este tipo, y el runtime lo
		// registra y sigue con la siguiente. No se reintenta.
		s.fallo(w, 400, "InvalidStateTransition", "esa invocacion ya no esta en vuelo")
		return
	}
	s.json(w, 202, map[string]string{"status": "OK"})
}

// errorDeArranque es lo que llama un runtime que no consiguio ni empezar: no pudo cargar tu
// manejador, le falta una variable de entorno, se rompio al importar. En Amazon esto es el
// `Runtime.InvalidEntrypoint` que sale antes de cualquier invocacion.
func (s *ServicioLambda) errorDeArranque(w http.ResponseWriter, r *http.Request) {
	cuerpo, _ := io.ReadAll(r.Body)
	s.registro("POST .../runtime/init/error [%s ni arranco] %s",
		r.PathValue("nombre"), string(cuerpo))
	s.json(w, 202, map[string]string{"status": "OK"})
}

// ── utilidades ───────────────────────────────────────────────────────────

// El plazo de la funcion, como contexto. Es LO MISMO que el `Timeout` que configuras en la
// consola de Amazon: no hay ningun otro mecanismo detras.
func contextoConLimite(f *Funcion) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(f.Segundos)*time.Second)
}

// Igual, pero colgando del contexto de la peticion: si el cliente cuelga, esto se cancela
// tambien. Se cumple el primero de los dos que venza.
func contextoConLimiteDe(r *http.Request, f *Funcion) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), time.Duration(f.Segundos)*time.Second)
}

func (s *ServicioLambda) json(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(v)
}

// fallo devuelve un error con la forma que el SDK sabe convertir en un tipo de error de Go.
// La cabecera X-Amzn-Errortype es la que decide cual: sin ella, el SDK solo ve «algo fue mal».
func (s *ServicioLambda) fallo(w http.ResponseWriter, estado int, tipo, mensaje string) {
	s.registro("  -> %d %s", estado, tipo)
	w.Header().Set("X-Amzn-Errortype", tipo)
	s.json(w, estado, map[string]string{"Type": "User", "message": mensaje})
}

func (s *ServicioLambda) noExiste(w http.ResponseWriter, nombre string) {
	s.fallo(w, 404, "ResourceNotFoundException",
		"Function not found: "+arnDe(nombre))
}
