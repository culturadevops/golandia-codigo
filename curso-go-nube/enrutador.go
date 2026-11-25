package main

// El enrutador de la raiz: como UN puerto sirve VARIOS servicios de AWS.
//
// Aqui hay una sorpresa que solo se ve montando esto. Los servicios de AWS no comparten
// protocolo:
//
//	S3         el metodo y la RUTA dicen que quieres.   PUT /fotos/gato.png   -> respuesta XML
//	DynamoDB   siempre POST a "/", cuerpo JSON, y la operacion en una CABECERA:
//	           X-Amz-Target: DynamoDB_20120810.PutItem
//	SQS        igual que DynamoDB, con otro prefijo:
//	           X-Amz-Target: AmazonSQS.SendMessage
//
// Son diseños distintos dentro de la misma nube, y el motivo es historico: S3 es de 2006 y se
// penso como almacen de objetos sobre HTTP; DynamoDB es de 2012 y usa el estilo "JSON-RPC" que
// Amazon adopto despues. SQS es MAS viejo que S3 (2006 tambien, con el protocolo Query de toda
// la vida), pero Amazon lo migro a JSON, asi que hoy habla como DynamoDB.
//
// Asi que para servir varios en el mismo puerto no hace falta magia: se mira la cabecera.
//
// 🥇 Y con Lambda aparecio un tercer caso que rompio el esquema. Lambda va por RUTA, como S3:
//
//	Lambda     POST /2015-03-31/functions/saludar/invocations
//
// Y ahi no vale meterlo en el mismo mux que S3, porque S3 se queda con la raiz ENTERA con sus
// comodines (`/{bucket}/{clave...}`). No es que se lie: es que el mux de Go ni siquiera
// arranca. Suelta un panico al registrar la ruta, diciendo que `GET /2015-03-31/functions`
// atiende mas metodos que `HEAD /{bucket}/{clave...}` —porque un GET registrado contesta
// tambien a HEAD— pero tiene la ruta mas concreta, y no sabe cual gana.
//
// O sea que el que va por rutas concretas tiene que ir DELANTE, en su propio mux, no al lado.
//
// 🥇 Y con SNS aparecio un CUARTO protocolo, que es el mas viejo de todos:
//
//	SNS        POST a "/" con un FORMULARIO en el cuerpo y la operacion en un campo:
//	           Action=Publish&TopicArn=...&Message=hola
//
// Es el protocolo «Query» de siempre, el que tenia SQS antes de que Amazon lo migrara a JSON.
// SNS se quedo donde estaba. Asi que en la misma nube conviven cuatro formas distintas de decir
// «haz esto», y ninguna es un error de nadie: son cuatro epocas.
//
// ⚠️ Para repartirlo hace falta un criterio POSITIVO otra vez, y el bueno aqui NO es el cuerpo
// —leerlo aqui se lo quitaria a quien viene detras— sino el TIPO DE CONTENIDO. SNS manda un
// formulario; DynamoDB y SQS mandan JSON. Se distingue sin tocar el cuerpo.

import (
	"net/http"
	"strings"
)

// Raiz devuelve el manejador que reparte entre servicios.
//
// `porRuta` va del principio de la RUTA al servicio, y se mira primero. `porObjetivo` va del
// prefijo de `X-Amz-Target` al servicio. `formulario` atiende los POST a "/" que traen un
// formulario en vez de JSON, que hoy es solo SNS. Se pasa como mapa y no con un
// `switch` a mano para que añadir el servicio siguiente sea una linea en `main`, no una
// modificacion aqui dentro.
//
// ⚠️ El criterio tiene que ser POSITIVO: se pregunta por lo que identifica a cada servicio, no
// por "lo que no es S3". Un criterio por descarte funciona con dos y se rompe callado con el
// tercero — que es exactamente lo que acaba de entrar.
func Raiz(s3 http.Handler, porObjetivo map[string]http.Handler,
	porRuta map[string]http.Handler, formulario http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Primero los que se reconocen por el principio de la ruta: si no van antes, S3 se
		// los come con sus comodines.
		for prefijo, servicio := range porRuta {
			if strings.HasPrefix(r.URL.Path, prefijo) {
				servicio.ServeHTTP(w, r)
				return
			}
		}
		// El protocolo viejo: un formulario en el cuerpo. Va antes que S3 porque S3 se queda
		// con la raiz, y va despues de las rutas porque es mas general que ellas.
		if formulario != nil && r.Method == http.MethodPost && r.URL.Path == "/" &&
			strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
			formulario.ServeHTTP(w, r)
			return
		}

		objetivo := r.Header.Get("X-Amz-Target")
		for prefijo, servicio := range porObjetivo {
			if strings.HasPrefix(objetivo, prefijo) {
				servicio.ServeHTTP(w, r)
				return
			}
		}
		s3.ServeHTTP(w, r)
	})
}
