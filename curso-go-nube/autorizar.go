package main

// La comprobacion de permisos. La segunda mitad del episodio 10.
//
// Y lo primero que descubres al escribir esto es que la parte dificil NO es decidir si puedes.
// Eso son veinte lineas y estan en `permisos.go`. La parte dificil es saber QUE ESTAS PIDIENDO.
//
// 🥇 Porque una politica habla de `s3:PutObject` sobre `arn:aws:s3:::fotos/gato.png`, y lo que
// llega por el cable es esto:
//
//	PUT /fotos/gato.png
//
// Alguien tiene que traducir lo segundo a lo primero. Y esa traduccion es DISTINTA en cada
// servicio, porque cada servicio habla su propio protocolo: uno lo dice en la ruta, otro en una
// cabecera, y otro dentro del cuerpo.
//
// ⚠️ Y de ahi sale el problema que ya salio en el episodio 4: para saber sobre QUE recurso pides
// algo, a veces hay que LEER EL CUERPO. Y si lo lees, tienes que volver a dejarlo puesto o el
// servicio de detras se encuentra una peticion vacia.

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Autorizador struct {
	permisos *Permisos
	registro func(formato string, args ...any)
	cuenta   string
	region   string
}

func NuevoAutorizador(p *Permisos, registro func(string, ...any)) *Autorizador {
	if registro == nil {
		registro = func(string, ...any) {}
	}
	return &Autorizador{permisos: p, registro: registro,
		cuenta: "000000000000", region: "us-east-1"}
}

// Envolver comprueba los permisos de cada peticion.
//
// ⚠️ Va DESPUES de la firma y ANTES del reparto por servicio. Ese orden no es casual: primero
// hay que saber quien eres (la firma), luego que puedes (esto), y solo entonces hacerlo. Al
// reves no tiene sentido: no puedes mirar los permisos de alguien que todavia no sabes quien es.
func (a *Autorizador) Envolver(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a == nil || a.permisos == nil {
			siguiente.ServeHTTP(w, r)
			return
		}
		acceso := accesoDeLaFirma(r)
		ident, hay := a.permisos.Buscar(acceso)
		if !hay {
			// Sin politicas para esta clave no se prohibe nada: es el modo de los nueve
			// episodios anteriores, y tiene que seguir funcionando igual.
			siguiente.ServeHTTP(w, r)
			return
		}

		servicio, accion, recurso, cuerpo := a.quePide(r)
		// El cuerpo se vuelve a dejar puesto SIEMPRE, se haya leido o no.
		if cuerpo != nil {
			r.Body = io.NopCloser(bytes.NewReader(cuerpo))
		}
		if accion == "" {
			// No se ha podido traducir la peticion. No se inventa un permiso: se deja pasar y
			// se dice en voz alta, que es mas honesto que negar por no entender.
			a.registro("  -> ⚠️ no se que permiso pide esta peticion: la dejo pasar")
			siguiente.ServeHTTP(w, r)
			return
		}

		v := Evaluar(ident.Politicas, accion, recurso)
		if !v.Permitido {
			a.registro("  -> 403 %s NO puede %s sobre %s (%s)", ident.Nombre, accion, recurso,
				v.Motivo)
			a.negar(w, servicio, ident, accion, recurso, v)
			return
		}
		a.registro("  -> permitido: %s puede %s sobre %s", ident.Nombre, accion, recurso)
		siguiente.ServeHTTP(w, r)
	})
}

// accesoDeLaFirma saca el identificador de acceso de la cabecera de autorizacion.
//
//	Credential=prueba/20260825/us-east-1/s3/aws4_request
//
// ⚠️ Y esto vale AUNQUE la comprobacion de firma este apagada, porque el cliente firma siempre.
// Aqui solo se lee quien DICE ser. Comprobar que es verdad es el episodio 4, y son dos cosas
// distintas a proposito: se puede tener permisos sin comprobar firmas, y al reves.
func accesoDeLaFirma(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	i := strings.Index(auth, "Credential=")
	if i < 0 {
		return ""
	}
	cred := auth[i+len("Credential="):]
	if j := strings.IndexAny(cred, ", "); j >= 0 {
		cred = cred[:j]
	}
	if j := strings.Index(cred, "/"); j >= 0 {
		return cred[:j]
	}
	return cred
}

// ── la traduccion: de una peticion HTTP a un permiso ─────────────────────

// quePide traduce la peticion. Devuelve tambien el cuerpo si tuvo que leerlo, para que quien
// llama lo vuelva a dejar puesto.
func (a *Autorizador) quePide(r *http.Request) (servicio, accion, recurso string, cuerpo []byte) {
	// Lambda y sus rutas.
	if strings.HasPrefix(r.URL.Path, "/2015-03-31/") || strings.HasPrefix(r.URL.Path, "/f/") {
		return "lambda", a.deLambda(r), a.arnLambda(r), nil
	}
	// SNS: un formulario.
	if r.Method == http.MethodPost && r.URL.Path == "/" &&
		strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		crudo, _ := io.ReadAll(r.Body)
		form, _ := url.ParseQuery(string(crudo))
		accion := form.Get("Action")
		if accion == "" {
			return "sns", "", "", crudo
		}
		destino := form.Get("TopicArn")
		if destino == "" && form.Get("Name") != "" {
			destino = "arn:aws:sns:" + a.region + ":" + a.cuenta + ":" + form.Get("Name")
		}
		if destino == "" {
			destino = "*"
		}
		return "sns", "sns:" + accion, destino, crudo
	}
	// DynamoDB y SQS: la operacion en la cabecera, el recurso en el cuerpo.
	if objetivo := r.Header.Get("X-Amz-Target"); objetivo != "" {
		crudo, _ := io.ReadAll(r.Body)
		op := objetivo
		if i := strings.LastIndex(objetivo, "."); i >= 0 {
			op = objetivo[i+1:]
		}
		switch {
		case strings.HasPrefix(objetivo, "DynamoDB_"):
			var c struct {
				TableName string `json:"TableName"`
			}
			_ = json.Unmarshal(crudo, &c)
			return "dynamodb", "dynamodb:" + op,
				"arn:aws:dynamodb:" + a.region + ":" + a.cuenta + ":table/" + c.TableName, crudo
		case strings.HasPrefix(objetivo, "AmazonSQS."):
			var c struct {
				QueueURL  string `json:"QueueUrl"`
				QueueName string `json:"QueueName"`
			}
			_ = json.Unmarshal(crudo, &c)
			nombre := c.QueueName
			if nombre == "" && c.QueueURL != "" {
				nombre = nombreDeURL(c.QueueURL)
			}
			return "sqs", "sqs:" + op,
				"arn:aws:sqs:" + a.region + ":" + a.cuenta + ":" + nombre, crudo
		case strings.HasPrefix(objetivo, "Logs_"):
			// ⚠️ Mismo protocolo y campo en MINUSCULA: `logGroupName`, no `LogGroupName`. Si
			// copias la estructura de al lado cambiando el nombre del campo, esto se queda
			// vacio y el permiso pasa a ser sobre "*" sin que salte nada.
			var c struct {
				LogGroupName string `json:"logGroupName"`
			}
			_ = json.Unmarshal(crudo, &c)
			// Y el ARN de un grupo acaba en ":*". No es un comodin que ponemos nosotros: es
			// parte del nombre del recurso, y una politica escrita sin el no encaja.
			recurso := "arn:aws:logs:" + a.region + ":" + a.cuenta + ":log-group:" +
				c.LogGroupName + ":*"
			if c.LogGroupName == "" {
				recurso = "*"
			}
			return "logs", "logs:" + op, recurso, crudo
		}
		return "", "", "", crudo
	}
	// Y lo que queda es S3, que lo dice todo con el metodo y la ruta.
	return "s3", a.deS3(r), a.arnS3(r), nil
}

// deS3 traduce metodo + ruta a un permiso de S3.
//
// ⚠️ Y aqui esta la primera sorpresa: el MISMO metodo y la MISMA forma de ruta son operaciones
// distintas segun cuantos tramos tenga. `PUT /fotos` crea un bucket; `PUT /fotos/gato.png` sube
// un objeto. Un permiso para lo uno no es un permiso para lo otro, y por eso los ARN de S3
// tienen dos formas.
func (a *Autorizador) deS3(r *http.Request) string {
	partes := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/"), "/", 2)
	bucket := partes[0]
	hayClave := len(partes) > 1 && partes[1] != ""

	switch r.Method {
	case http.MethodGet, http.MethodHead:
		switch {
		case bucket == "":
			return "s3:ListAllMyBuckets"
		case r.URL.Query().Has("notification"):
			return "s3:GetBucketNotification"
		case hayClave:
			return "s3:GetObject"
		default:
			return "s3:ListBucket"
		}
	case http.MethodPut:
		switch {
		case r.URL.Query().Has("notification"):
			return "s3:PutBucketNotification"
		case hayClave:
			return "s3:PutObject"
		default:
			return "s3:CreateBucket"
		}
	case http.MethodDelete:
		if hayClave {
			return "s3:DeleteObject"
		}
		return "s3:DeleteBucket"
	}
	return ""
}

// arnS3 arma el ARN. Fijate en los tres dos puntos seguidos: S3 no lleva ni region ni cuenta,
// porque el nombre de un bucket es unico en el mundo entero.
func (a *Autorizador) arnS3(r *http.Request) string {
	camino := strings.TrimPrefix(r.URL.Path, "/")
	if camino == "" {
		return "*"
	}
	return "arn:aws:s3:::" + camino
}

func (a *Autorizador) deLambda(r *http.Request) string {
	p := r.URL.Path
	switch {
	case strings.HasPrefix(p, "/f/"):
		// La interfaz del runtime. No es una operacion tuya: es la funcion pidiendo trabajo.
		return "lambda:RuntimeAPI"
	case strings.HasSuffix(p, "/invocations"):
		return "lambda:InvokeFunction"
	case strings.Contains(p, "/event-source-mappings"):
		switch r.Method {
		case http.MethodPost:
			return "lambda:CreateEventSourceMapping"
		case http.MethodDelete:
			return "lambda:DeleteEventSourceMapping"
		default:
			return "lambda:ListEventSourceMappings"
		}
	}
	switch r.Method {
	case http.MethodPost:
		return "lambda:CreateFunction"
	case http.MethodDelete:
		return "lambda:DeleteFunction"
	case http.MethodGet:
		if strings.HasSuffix(p, "/functions") {
			return "lambda:ListFunctions"
		}
		return "lambda:GetFunction"
	}
	return ""
}

func (a *Autorizador) arnLambda(r *http.Request) string {
	p := strings.TrimPrefix(r.URL.Path, "/")
	tramos := strings.Split(p, "/")
	// /2015-03-31/functions/{nombre}[/invocations]
	if len(tramos) >= 3 && tramos[1] == "functions" {
		return arnDe(tramos[2])
	}
	if strings.HasPrefix(r.URL.Path, "/f/") && len(tramos) >= 2 {
		return arnDe(tramos[1])
	}
	return "*"
}

// ── el «no» ──────────────────────────────────────────────────────────────

// negar contesta que no. Y contesta CON LA FORMA DEL SERVICIO.
//
// ⚠️ Eso no es coquetear: si le devuelves a un cliente de S3 un JSON, el SDK no puede convertir
// eso en un error tipado y le llega a tu codigo un «respuesta inesperada» en vez de un
// `AccessDenied`. El error tiene que hablar el idioma del servicio al que ibas.
//
// 🥇 Y fijate en lo que dice el mensaje, porque es la diferencia con el episodio 4: aqui SI se
// puede decir que falta. «No puedes hacer X sobre Y». El servidor lo sabe, porque el permiso que
// falta esta escrito. Con una firma que no cuadra no puede: no sabe QUE esta mal, solo que no
// coincide. Por eso `SignatureDoesNotMatch` es opaco y `AccessDenied` no.
func (a *Autorizador) negar(w http.ResponseWriter, servicio string, ident *Identidad,
	accion, recurso string, v Veredicto) {
	usuario := "arn:aws:iam::" + a.cuenta + ":user/" + ident.Nombre
	mensaje := fmt.Sprintf("User: %s is not authorized to perform: %s on resource: %s",
		usuario, accion, recurso)
	if v.Motivo != "" {
		mensaje += " (" + v.Motivo + ")"
	}
	if v.Sid != "" {
		mensaje += " [sentencia: " + v.Sid + "]"
	}

	switch servicio {
	case "s3":
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(xml.Header))
		_ = xml.NewEncoder(w).Encode(errorXML{
			Code: "AccessDenied", Message: mensaje, Resource: recurso})
	case "sns":
		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(xml.Header))
		_, _ = fmt.Fprintf(w, `<ErrorResponse xmlns="http://sns.amazonaws.com/doc/2010-03-31/">`+
			`<Error><Type>Sender</Type><Code>AuthorizationError</Code><Message>%s</Message>`+
			`</Error><RequestId>00000000-0000-4000-8000-000000000000</RequestId>`+
			`</ErrorResponse>`, xmlEscapa(mensaje))
	case "lambda":
		w.Header().Set("X-Amzn-Errortype", "AccessDeniedException")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"Type": "User", "message": mensaje})
	case "logs":
		// La misma familia que DynamoDB y SQS, con su version de JSON. Y el nombre de la
		// excepcion de este servicio no es `AccessDeniedException`: la suya se llama asi.
		w.Header().Set("Content-Type", "application/x-amz-json-1.1")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"__type": "AccessDeniedException", "message": mensaje})
	default: // dynamodb y sqs
		tipo := "com.amazon.coral.service#AccessDeniedException"
		if servicio == "sqs" {
			tipo = tipoSQS + "AccessDeniedException"
		}
		w.Header().Set("Content-Type", "application/x-amz-json-1.0")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"__type": tipo, "message": mensaje})
	}
}
