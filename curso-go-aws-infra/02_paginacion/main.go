// Lección 2 · Inventario: reservas, instancias y páginas.
//
// Dos trampas que no dan error:
//  1. DescribeInstances no devuelve instancias: devuelve RESERVAS, y cada una lleva dentro
//     las instancias que se lanzaron juntas. Contar reservas es contar mal.
//  2. La respuesta viene por páginas. Si no sigues el NextToken, tienes la primera y nada más.
package main

import (
	"context"
	"fmt"

	"curso-aws/comun"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func main() {
	ctx := context.Background()
	comun.Reset()
	c := comun.EC2("us-east-1")
	comun.Crear(c, 3, "Name", "web", "env", "prod") // una sola llamada, tres instancias
	for i := 1; i <= 9; i++ {
		comun.Crear(c, 1, "Name", fmt.Sprintf("worker-%02d", i), "env", "prod")
	}
	fmt.Println("  lanzadas: 1 llamada con 3 instancias + 9 llamadas con 1 = 12 instancias")

	comun.Titulo("2.1 · Contar reservas no es contar instancias")
	out, _ := c.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
	n := 0
	for _, r := range out.Reservations {
		n += len(r.Instances)
	}
	fmt.Printf("  len(out.Reservations)          -> %d   <- lo que cuenta mucha gente\n", len(out.Reservations))
	fmt.Printf("  instancias dentro de ellas     -> %d\n", n)

	comun.Titulo("2.2 · Una sola pagina (MaxResults=5, el minimo que admite la API)")
	pg, err := c.DescribeInstances(ctx, &ec2.DescribeInstancesInput{MaxResults: aws.Int32(5)})
	n = 0
	for _, r := range pg.Reservations {
		n += len(r.Instances)
	}
	fmt.Printf("  instancias: %d de 12   err = %v   NextToken vacio? %v\n", n, err, pg.NextToken == nil)
	fmt.Println("  Ni un error. Solo un token que hay que seguir.")

	comun.Titulo("2.3 · Con el paginador del SDK")
	todas := comun.Todas(c, &ec2.DescribeInstancesInput{MaxResults: aws.Int32(5)})
	fmt.Printf("  instancias: %d de 12\n", len(todas))
}
