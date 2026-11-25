// Lección 1 · ¿En qué cuenta y en qué región estoy?
//
// La trampa: pedir la lista de instancias en la región equivocada NO es un error. Devuelve
// cero, y cero parece «no hay nada». La primera línea de toda herramienta de infraestructura
// es preguntar a STS quién eres, y decir en voz alta la cuenta y la región.
package main

import (
	"context"
	"fmt"
	"log"

	"curso-aws/comun"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func main() {
	ctx := context.Background()
	comun.Reset()

	comun.Titulo("1.1 · Lo primero: quien soy")
	cfg := comun.Config("us-east-1")
	id, err := sts.NewFromConfig(cfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("  cuenta:  ", aws.ToString(id.Account))
	fmt.Println("  quien:   ", aws.ToString(id.Arn))
	fmt.Println("  region:  ", cfg.Region)
	fmt.Println("  endpoint:", comun.Endpoint(), "  <- el emulador; sin esta variable, AWS de verdad")

	comun.Titulo("1.2 · Tres servidores en us-east-1")
	este := comun.EC2("us-east-1")
	comun.Crear(este, 3, "Name", "web", "env", "prod")
	for _, r := range []string{"us-east-1", "us-west-2", "eu-west-1"} {
		out, err := comun.EC2(r).DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
		n := 0
		for _, rv := range out.Reservations {
			n += len(rv.Instances)
		}
		fmt.Printf("  %-10s -> %d instancias   (err = %v)\n", r, n, err)
	}
	fmt.Println("  Cero no es «no hay»: es «no hay AQUI».")

	comun.Titulo("1.3 · La salvaguarda: si la cuenta no es la esperada, no toco nada")
	esperada := "123456789012"
	for _, cuenta := range []string{aws.ToString(id.Account), "987654321098"} {
		if cuenta != esperada {
			fmt.Printf("  cuenta %s -> ⛔ esperaba %s: no hago nada\n", cuenta, esperada)
			continue
		}
		fmt.Printf("  cuenta %s -> ok, sigo\n", cuenta)
	}
}
