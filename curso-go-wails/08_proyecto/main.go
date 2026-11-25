// Lección 8 · El proyecto: un inspector de colas para nubecita, en una ventana.
// Junta lo del curso: el slice que nunca va nil (lección 2), los errores como texto (3), los
// eventos con saludo de «Listo» (4) y el contexto de OnStartup (5).
package main

import (
	"context"
	"embed"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"curso-wails/comun"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var web embed.FS

type Cola struct {
	Nombre   string `json:"nombre"`
	Mensajes int    `json:"mensajes"`
}

type App struct {
	comun.Chivato
	ctx context.Context
	sqs *sqs.Client
}

// Colas nunca devuelve nil: si no hay ninguna, devuelve [] (lección 2).
func (a *App) Colas() ([]Cola, error) {
	ctx, cancel := context.WithTimeout(a.ctx, 3*time.Second)
	defer cancel()
	out := []Cola{}
	r, err := a.sqs.ListQueues(ctx, &sqs.ListQueuesInput{})
	if err != nil {
		return out, err // el texto le llega a la ventana (lección 3)
	}
	for _, u := range r.QueueUrls {
		at, err := a.sqs.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{QueueUrl: aws.String(u),
			AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameApproximateNumberOfMessages}})
		if err != nil {
			return out, err
		}
		n, _ := strconv.Atoi(at.Attributes[string(types.QueueAttributeNameApproximateNumberOfMessages)])
		out = append(out, Cola{Nombre: u[strings.LastIndex(u, "/")+1:], Mensajes: n})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Nombre < out[j].Nombre })
	return out, nil
}

// Listo: la ventana ya escucha; empieza el refresco por eventos (lección 4).
func (a *App) Listo() {
	go func() {
		for {
			colas, err := a.Colas()
			if err != nil {
				runtime.EventsEmit(a.ctx, "error", err.Error())
			} else {
				runtime.EventsEmit(a.ctx, "colas", colas)
			}
			time.Sleep(time.Second)
		}
	}()
}

func main() {
	ep := os.Getenv("ENDPOINT")
	a := &App{sqs: sqs.New(sqs.Options{Region: "us-east-1", BaseEndpoint: aws.String(ep), RetryMaxAttempts: 1,
		Credentials: credentials.NewStaticCredentialsProvider("prueba", "prueba", "")})}
	wails.Run(&options.App{
		Title: "nubecita · colas", Width: 900, Height: 640,
		AssetServer: &assetserver.Options{Assets: web},
		OnStartup:   func(ctx context.Context) { a.ctx = ctx }, // lección 5
		Bind:        []interface{}{a},
	})
}
