// nubecita — una nube de mentira, escrita en Go, para probar tu codigo sin pagar ni pedir
// permiso a nadie.
//
// Episodio 1: S3.  ·  Episodio 2: DynamoDB.  ·  Episodio 3: SQS.  ·  Episodio 4: SigV4.
// Episodio 5: Lambda, con la API del runtime de verdad.  ·  Episodio 6: persistencia.
// Episodio 7: el lanzador — nubecita levanta ella sola los contenedores de tus funciones.
// Episodio 8: los disparadores — S3 y SQS invocan tus funciones sin que tu llames.
// Episodio 9: SNS — un tema reparte a N colas, y el mensaje llega envuelto.
// Episodio 10: los permisos — la otra mitad del episodio 4: quien eres, y que puedes.
// El SDK de AWS para Go y la CLI de AWS hablan con esto sin saber que no es Amazon. Lo unico
// que cambia en tu codigo es el endpoint.
//
//	go run .                       # escucha en 127.0.0.1:4566
//	go run . -puerto 9000 -callado
//
// El puerto 4566 es el de LocalStack a proposito: si ya tenias algo apuntando ahi, no tocas
// nada.
//
// ⛔ Esto NO es para produccion y no lo va a ser nunca. No hay firmas, no hay permisos, no hay
// disco. Es un doble de pruebas, y un doble de pruebas vale justo por lo que NO hace.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	puerto := flag.Int("puerto", 4566, "puerto en el que escuchar")
	// ⛔ Dentro de un contenedor hay que escuchar en 0.0.0.0: con 127.0.0.1 el proceso solo se
	// oye a si mismo y el `-p` de docker publica un puerto al que no contesta nadie. Es el
	// fallo numero uno al meter un servidor en un contenedor, y no da ningun error.
	escuchar := flag.String("escuchar", "127.0.0.1", "interfaz en la que escuchar")
	callado := flag.Bool("callado", false, "no imprimir cada peticion")
	// ⛔ Apagado por defecto: un doble de pruebas no necesita autenticar a nadie, y pedir
	// firmas por defecto solo haria mas dificil empezar. Con -firmar se comprueban de verdad.
	firmar := flag.Bool("firmar", false, "comprobar la firma SigV4 de cada peticion")
	acceso := flag.String("acceso", "prueba", "identificador de acceso admitido")
	secreto := flag.String("secreto", "prueba", "el secreto de ese identificador")
	// Persistencia. Son DOS banderas y hacen cosas distintas a proposito:
	//   -datos    lee al arrancar Y escribe cuando algo cambia  (maquina de desarrollo)
	//   -semilla  lee al arrancar y NO escribe nunca            (pruebas reproducibles)
	// Sin ninguna de las dos, nubecita sigue siendo lo que era: memoria y nada mas.
	datos := flag.String("datos", "", "fichero de estado: se lee al arrancar y se va guardando")
	semilla := flag.String("semilla", "", "fichero de estado que se lee al arrancar y NO se escribe")
	cada := flag.Duration("cada", 200*time.Millisecond, "cuanto se agrupan los cambios antes de guardar")
	// El lanzador. Apagado por defecto: hace falta el socket de Docker, y eso no se enciende
	// solo. Con -docker, nubecita levanta los contenedores de tus funciones cuando invocas.
	socket := flag.String("docker", "", "socket del demonio de Docker (p.ej. /var/run/docker.sock)")
	red := flag.String("red", "bridge", "red de Docker donde levantar los runtimes")
	yo := flag.String("yo", "", "como me ven los runtimes (p.ej. nubecita:4566)")
	maxRuntimes := flag.Int("max", 4, "tope de contenedores por funcion")
	ocioso := flag.Duration("ocioso", 60*time.Second, "apagar los runtimes tras este rato sin invocar")
	sondeo := flag.Duration("sondeo", 200*time.Millisecond, "cada cuanto mira un disparador su cola vacia")
	// Los permisos. Apagados por defecto, como las firmas y por el mismo motivo: un doble de
	// pruebas no tiene que ponerte trabas hasta que tu se lo pidas.
	permisos := flag.String("permisos", "", "fichero JSON con las politicas por clave de acceso")
	flag.Parse()

	if *datos != "" && *semilla != "" {
		log.Fatal("-datos y -semilla a la vez no: decide si esto guarda o no guarda")
	}

	registro := log.New(os.Stdout, "", 0).Printf
	if *callado {
		registro = func(string, ...any) {}
	}

	// S3 se enruta por METODO y RUTA; DynamoDB y SQS van siempre a POST "/" con la operacion
	// en una cabecera. `Raiz` mira esa cabecera y reparte. Por eso caben en el mismo puerto, y
	// añadir el siguiente servicio es UNA linea en este mapa.
	dir := fmt.Sprintf("%s:%d", *escuchar, *puerto)
	almacen, tablas, colas, funciones := NuevoAlmacen(), NuevasTablas(), NuevasColas(), NuevasFunciones()
	temas := NuevosTemas()
	registros := NuevosRegistros()

	mux := http.NewServeMux()
	s3 := NuevoS3(almacen, registro)
	s3.Rutas(mux)
	porObjetivo := map[string]http.Handler{
		"DynamoDB_":  NuevoDynamo(tablas, registro),
		"AmazonSQS.": NuevoSQS(colas, registro),
		// CloudWatch Logs: el mismo protocolo, otro prefijo, y los campos en minuscula.
		objetivoLogs: NuevoLogs(registros, registro),
	}

	// SNS no va ni por ruta ni por cabecera: manda un formulario. Se reparte por el tipo de
	// contenido, que es el unico criterio positivo que no obliga a leer el cuerpo.
	sns := NuevoSNS(temas, colas, registro)

	// Lambda no va por X-Amz-Target: va por RUTA, como S3. Y por eso necesita su propio mux y
	// entrar ANTES que S3, que se queda con la raiz entera. Son dos APIs en una: la que llama
	// tu codigo y la que llama la funcion desde dentro de su contenedor.
	// ⛔ Montar el socket de Docker aqui dentro le da a este proceso control del demonio
	// ENTERO. Es lo que hace falta para levantar contenedores, es lo que hace LocalStack, y es
	// justo por eso que esto no se pone en una maquina compartida.
	var lanzador *Lanzador
	if *socket != "" {
		if *yo == "" {
			log.Fatal("-docker sin -yo: los runtimes tienen que saber a que direccion pedirme trabajo")
		}
		lanzador = NuevoLanzador(ClienteDocker(*socket), "http://docker", *red, *yo,
			*maxRuntimes, *ocioso, registro).ConRegistros(registros)
	}

	// Los disparadores juntan las tres piezas: sacan de las colas, escuchan a S3, e invocan.
	disparadores := NuevosDisparadores(funciones, lanzador, colas, *sondeo, registro)
	s3.ConNotificador(NuevoNotificador(disparadores, colas, registro))

	lam := NuevoLambda(funciones, registro).ConLanzador(lanzador).ConDisparadores(disparadores)
	rutasDeLambda := lam.Rutas()
	porRuta := map[string]http.Handler{}
	for _, p := range lam.Prefijos() {
		porRuta[p] = rutasDeLambda
	}

	// Cargar, si hay de donde. Si el fichero no existe se arranca vacio, que la primera vez es
	// lo normal. Si existe y NO se entiende, se para: cargar la mitad de un estado es peor que
	// no cargarlo, porque el fallo aparece tres pruebas mas tarde.
	fuente := *datos
	if fuente == "" {
		fuente = *semilla
	}
	if fuente != "" {
		inst, err := Cargar(fuente)
		if err != nil {
			log.Fatalf("no puedo leer %s: %v", fuente, err)
		}
		if err := Aplicar(inst, almacen, tablas, colas, funciones); err != nil {
			log.Fatalf("%s no se puede aplicar: %v", fuente, err)
		}
		if inst == nil {
			log.Printf("%s no existe todavia: arranco vacio", fuente)
		} else {
			log.Printf("estado cargado de %s (%d bucket(s), %d tabla(s), %d cola(s), %d funcion(es))",
				fuente, len(inst.Buckets), len(inst.Tablas), len(inst.Colas), len(inst.Funciones))
		}
	}

	// ⚠️ El orden importa y no es capricho: primero saber QUIEN eres (la firma), luego QUE
	// puedes (esto), y solo entonces hacerlo. Mirar los permisos de alguien que todavia no
	// sabes quien es no significa nada.
	claves := map[string]string{}
	if *firmar {
		claves[*acceso] = *secreto
	}
	var autorizador *Autorizador
	if *permisos != "" {
		crudo, err := os.ReadFile(*permisos)
		if err != nil {
			log.Fatalf("no puedo leer %s: %v", *permisos, err)
		}
		cargadas, err := CargarPoliticas(crudo)
		if err != nil {
			log.Fatalf("%s: %v", *permisos, err)
		}
		perm := NuevosPermisos(registro)
		for acceso, v := range cargadas {
			perm.Poner(acceso, v.Nombre, []Politica{v.Politica})
			// El mismo fichero da de alta la clave para la firma: identidad y permisos son
			// dos cosas del mismo usuario.
			if v.Secreto != "" {
				claves[acceso] = v.Secreto
			}
			log.Printf("permisos: %s (%s) con %d sentencia(s)", acceso, v.Nombre,
				len(v.Politica.Statement))
		}
		autorizador = NuevoAutorizador(perm, registro)
	}
	if !*firmar {
		// Sin -firmar no se comprueba ninguna, aunque el fichero traiga secretos.
		claves = nil
	}
	verificador := NuevoVerificador(claves, registro)

	manejador := Raiz(mux, porObjetivo, porRuta, sns)
	if autorizador != nil {
		manejador = autorizador.Envolver(manejador)
	}
	manejador = verificador.Envolver(manejador)

	// La guarda va por FUERA de todo: marca «puede que algo haya cambiado» al terminar cada
	// peticion, y quien escribe decide si de verdad cambio. Con -semilla no hay guarda: se lee
	// y se olvida.
	var guarda *Guarda
	if *datos != "" {
		guarda = NuevaGuarda(*datos, *cada,
			func() Instantanea { return Reunir(almacen, tablas, colas, funciones) }, registro)
		manejador = guarda.Envolver(manejador)
	}

	srv := &http.Server{
		Addr: dir,
		// La comprobacion de firma envuelve TODO, antes del reparto por servicio: una peticion
		// mal firmada no debe llegar ni a saber a que servicio iba.
		Handler:           manejador,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Apagado limpio. Un emulador se arranca y se mata cien veces al dia desde un script o
	// desde un Makefile: si no cierras bien, la siguiente corrida se encuentra el puerto
	// ocupado y el fallo aparece en OTRA prueba, que es lo peor que puede pasar.
	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer parar()

	// El lanzador necesita saber cuanto vive el proceso: seguir la salida de un contenedor es
	// una conexion abierta, y esa no puede depender del contexto de una peticion.
	lanzador.ConVida(ctx)

	if guarda != nil {
		guarda.Arrancar(ctx)
	}
	// El vigilante apaga lo que lleva un rato sin usarse. Es lo que hace que el arranque en
	// frio VUELVA: nadie te guarda una maquina encendida por si acaso.
	lanzador.Vigilar(ctx, 5*time.Second, func() []*Funcion { return funciones.Listar() })

	go func() {
		firma := "sin comprobar firmas"
		if *firmar {
			firma = "comprobando firmas SigV4"
		}
		motor := "sin levantar runtimes"
		if lanzador != nil {
			motor = fmt.Sprintf("levantando runtimes en la red %s (tope %d, ocioso %s)",
				*red, *maxRuntimes, *ocioso)
		}
		puerta := "sin permisos"
		if autorizador != nil {
			puerta = "comprobando permisos de " + *permisos
		}
		disco := "sin guardar nada"
		switch {
		case *datos != "":
			disco = "guardando en " + *datos
		case *semilla != "":
			disco = "sembrado desde " + *semilla + ", sin guardar"
		}
		log.Printf("nubecita · S3, DynamoDB, SQS, SNS, Logs y Lambda en http://%s · %s · %s · %s · %s "+
			"· Ctrl+C para parar", dir, firma, puerta, disco, motor)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("no puedo escuchar en %s: %v", dir, err)
		}
	}()

	<-ctx.Done()
	log.Println("cerrando...")
	cierre, cancelar := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelar()
	if err := srv.Shutdown(cierre); err != nil {
		log.Printf("cierre forzado: %v", err)
	}
	// ⚠️ Y se espera a que la guarda haya escrito lo ultimo. Si el proceso saliera aqui sin
	// esperar, el ultimo cambio se quedaria en memoria — y lo descubririas manana.
	if guarda != nil {
		guarda.Esperar()
	}
	// Primero los disparadores: si se apagaran los contenedores antes, un bucle podria sacar
	// mensajes de la cola y no tener a quien darselos.
	disparadores.Apagar()
	// Y se apagan los contenedores que levantamos. Dejarlos vivos seria dejar runtimes
	// preguntandole por trabajo a un servidor que ya no existe.
	if lanzador != nil {
		limpieza, cancelar := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancelar()
		lanzador.Apagar(limpieza, "")
		a, b := lanzador.Cifras()
		log.Printf("runtimes levantados: %d · apagados: %d", a, b)
	}
}
