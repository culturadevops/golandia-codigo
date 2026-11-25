// sembrar: crea tres colas en nubecita y les mete mensajes. Con -mas N, mete N en «pedidos».
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func main() {
	mas := flag.Int("mas", 0, "mensajes a añadir a «pedidos»")
	flag.Parse()
	c := sqs.New(sqs.Options{Region: "us-east-1", BaseEndpoint: aws.String(os.Getenv("ENDPOINT")),
		Credentials: credentials.NewStaticCredentialsProvider("prueba", "prueba", "")})
	ctx := context.Background()
	plan := map[string]int{"pedidos": 4, "correos": 0, "pedidos-muertos": 1}
	if *mas > 0 {
		plan = map[string]int{"pedidos": *mas}
	}
	for cola, n := range plan {
		u, err := c.CreateQueue(ctx, &sqs.CreateQueueInput{QueueName: aws.String(cola)})
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		for i := 0; i < n; i++ {
			c.SendMessage(ctx, &sqs.SendMessageInput{QueueUrl: u.QueueUrl, MessageBody: aws.String("pedido")})
		}
	}
	fmt.Println("  sembrado:", plan)
}
