// Lección 7 · DryRun: el error que significa «adelante».
//
// EC2 deja preguntar «¿podría hacer esto?» sin hacerlo: DryRun = true. La respuesta llega
// SIEMPRE como error. DryRunOperation quiere decir «sí, podrías»; UnauthorizedOperation,
// «no tienes permiso». Un script que trata todo error como fallo nunca sabrá que podía.
package main

import (
	"context"
	"fmt"

	"curso-aws/comun"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// podria pregunta a EC2 si se podrían parar estas instancias, sin pararlas.
func podria(c *ec2.Client, ids []string) (bool, error) {
	_, err := c.StopInstances(context.Background(), &ec2.StopInstancesInput{InstanceIds: ids, DryRun: aws.Bool(true)})
	switch comun.Codigo(err) {
	case "DryRunOperation":
		return true, nil
	case "UnauthorizedOperation":
		return false, nil
	}
	return false, err
}

func main() {
	comun.Reset()
	c := comun.EC2("us-east-1")
	var ids []string
	for _, i := range comun.Crear(c, 2, "Name", "web", "env", "dev", "dueno", "pagos") {
		ids = append(ids, aws.ToString(i.InstanceId))
	}

	comun.Titulo("7.1 · StopInstances con DryRun = true")
	_, err := c.StopInstances(context.Background(), &ec2.StopInstancesInput{InstanceIds: ids, DryRun: aws.Bool(true)})
	fmt.Println("  err ->", err)
	fmt.Println("  codigo ->", comun.Codigo(err))

	comun.Titulo("7.2 · Leerlo bien")
	ok, err := podria(c, ids)
	fmt.Printf("  podria parar %d instancias? %v   (err = %v)\n", len(ids), ok, err)
	for _, i := range comun.Todas(c, &ec2.DescribeInstancesInput{InstanceIds: ids}) {
		fmt.Printf("  %s sigue %s\n", aws.ToString(i.InstanceId), i.State.Name)
	}

	comun.Titulo("7.3 · Lo que DryRun NO comprueba")
	fmt.Println("  Que las instancias sean las que tu querias. Eso lo decide tu filtro (leccion 8).")
}
