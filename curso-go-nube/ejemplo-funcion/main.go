// Una funcion de Lambda de verdad, escrita con `aws-lambda-go`, la biblioteca OFICIAL de
// Amazon. Ni una linea distinta de lo que subirias a produccion.
//
// Lo unico que cambia para que hable con nubecita en vez de con Amazon es una variable de
// entorno, la misma que le pone Lambda cuando arranca tu contenedor de verdad:
//
//	AWS_LAMBDA_RUNTIME_API=nubecita:4566/f/saludar
//
// Y eso es todo. No hay parche, no hay bandera de pruebas, no hay «modo local».
//
// Desde el episodio 8 esta funcion entiende TRES formas de evento, porque asi es como se ven de
// verdad los disparadores:
//
//	{"nombre":"Ada"}                        una invocacion normal, la tuya
//	{"Records":[{"eventSource":"aws:s3"}]}  S3 avisando de que subiste algo
//	{"Records":[{"eventSource":"aws:sqs"}]} un lote sacado de una cola
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

type Entrada struct {
	Nombre   string `json:"nombre"`
	TardarMs int    `json:"tardarMs"`
	Fallar   bool   `json:"fallar"`
}

type Salida struct {
	Saludo    string `json:"saludo"`
	Frio      bool   `json:"frio"`
	Invocacio int    `json:"invocacion"`
	PeticionD string `json:"peticion"`
	QuedanMs  int64  `json:"quedanMs"`
	Contenedo string `json:"contenedor"`
}

// 🥇 Estas dos variables son de PAQUETE, no de la invocacion. Sobreviven entre llamadas mientras
// el contenedor siga vivo, y por eso el arranque en frio se paga UNA vez. Es tambien el motivo
// de que guardar estado en una global de tu funcion sea una trampa: a veces esta y a veces no,
// y no puedes saber cual de las dos te va a tocar.
var (
	frio  = true
	veces = 0
)

// ── los eventos de los disparadores ──────────────────────────────────────

type registroEvento struct {
	EventSource string `json:"eventSource"`
	EventName   string `json:"eventName"`

	// de SQS
	MessageID string `json:"messageId"`
	Body      string `json:"body"`

	// de S3
	S3 struct {
		Bucket struct {
			Name string `json:"name"`
		} `json:"bucket"`
		Object struct {
			Key  string `json:"key"`
			Size int    `json:"size"`
		} `json:"object"`
	} `json:"s3"`
}

type evento struct {
	Records []registroEvento `json:"Records"`
}

func manejar(ctx context.Context, crudo json.RawMessage) (any, error) {
	eraFrio := frio
	frio = false
	veces++

	// ¿Es un disparador? Se mira si trae `Records`. Es lo mismo que hace cualquier funcion que
	// atiende a mas de una fuente, y por eso conviene que cada una tenga la suya.
	var ev evento
	if err := json.Unmarshal(crudo, &ev); err == nil && len(ev.Records) > 0 {
		switch ev.Records[0].EventSource {
		case "aws:s3":
			return deS3(ev)
		case "aws:sqs":
			return deSQS(ev)
		}
	}
	return normal(ctx, crudo, eraFrio)
}

// deS3 reacciona a un objeto nuevo. Y hace algo que se ve: escribe OTRO objeto.
//
// ⚠️ Escribe con un PUT de HTTP a secas, sin SDK y sin firmar, porque nubecita arranca sin pedir
// firmas. Contra Amazon esto no valdria — y ese es justo el limite honesto de la demo.
func deS3(ev evento) (any, error) {
	base := os.Getenv("NUBECITA")
	hechos := []string{}
	for _, r := range ev.Records {
		log.Printf("[S3] %s %s/%s (%d bytes)", r.EventName, r.S3.Bucket.Name, r.S3.Object.Key,
			r.S3.Object.Size)
		hechos = append(hechos, r.S3.Object.Key)
		if base == "" {
			continue
		}
		nota, _ := json.Marshal(map[string]any{
			"original": r.S3.Object.Key,
			"bytes":    r.S3.Object.Size,
			"evento":   r.EventName,
			"cuando":   time.Now().UTC().Format(time.RFC3339),
		})
		destino := base + "/" + r.S3.Bucket.Name + "/procesados/" +
			strings.ReplaceAll(r.S3.Object.Key, "/", "_") + ".json"
		if err := ponerObjeto(destino, nota); err != nil {
			log.Printf("[S3] no pude escribir %s: %v", destino, err)
		} else {
			log.Printf("[S3] escrito %s", destino)
		}
	}
	return map[string]any{"procesados": hechos}, nil
}

// deSQS atiende un LOTE. Y aqui esta lo que hay que entender de un consumidor de Lambda:
//
// 🥇 NO se borra ningun mensaje. Eso lo hace el disparador, y solo si esta funcion va bien. Lo
// unico que puede hacer la funcion es DECIR CUALES le fallaron, y eso son los `batchItemFailures`.
// Sin esa lista, un fallo de uno reintenta los diez.
func deSQS(ev evento) (any, error) {
	fallos := []map[string]string{}
	for _, r := range ev.Records {
		if strings.Contains(r.Body, "malo") {
			log.Printf("[SQS] %s = %q -> lo marco como fallido", r.MessageID, r.Body)
			fallos = append(fallos, map[string]string{"itemIdentifier": r.MessageID})
			continue
		}
		log.Printf("[SQS] %s = %q -> procesado", r.MessageID, r.Body)
	}
	return map[string]any{"batchItemFailures": fallos}, nil
}

func normal(ctx context.Context, crudo json.RawMessage, eraFrio bool) (any, error) {
	var e Entrada
	_ = json.Unmarshal(crudo, &e)

	if e.TardarMs > 0 {
		time.Sleep(time.Duration(e.TardarMs) * time.Millisecond)
	}
	if e.Fallar {
		return nil, errors.New("me han pedido que falle, y he fallado")
	}

	// De aqui salen el identificador de la peticion y el plazo. Los pone la API del runtime en
	// unas cabeceras, y `aws-lambda-go` los mete en el contexto.
	peticion := ""
	if lc, ok := lambdacontext.FromContext(ctx); ok {
		peticion = lc.AwsRequestID
	}
	var quedan int64
	if limite, hay := ctx.Deadline(); hay {
		quedan = time.Until(limite).Milliseconds()
	}

	nombre := e.Nombre
	if nombre == "" {
		nombre = "mundo"
	}
	maquina, _ := os.Hostname()
	return Salida{
		Saludo: "hola, " + nombre, Frio: eraFrio, Invocacio: veces,
		PeticionD: peticion, QuedanMs: quedan, Contenedo: maquina,
	}, nil
}

func ponerObjeto(url string, datos []byte) error {
	pet, err := http.NewRequest(http.MethodPut, url, strings.NewReader(string(datos)))
	if err != nil {
		return err
	}
	pet.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(pet)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return fmt.Errorf("%s contesto %d", url, res.StatusCode)
	}
	return nil
}

func main() {
	lambda.Start(manejar)
}
