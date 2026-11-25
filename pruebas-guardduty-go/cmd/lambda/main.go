// El handler de Lambda. Se compila (GOOS=linux GOARCH=arm64, runtime provided.al2023) pero en
// este repo no se despliega: la lógica se prueba en alerta/ con go test.
package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/culturadevops/pruebas-guardduty-go/alerta"
)

func handler(ctx context.Context, ev events.CloudWatchEvent) error {
	d, err := alerta.Decidir(ev)
	if err != nil {
		return err
	}
	if d.Accion == alerta.Ignorar {
		return nil
	}
	m, err := alerta.Mensaje(d)
	if err != nil {
		return err
	}
	// Aquí iría la publicación en SNS, Slack o lo que uses. Lo dejamos en el log.
	fmt.Println(string(m))
	return nil
}

func main() { lambda.Start(handler) }
