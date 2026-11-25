// Lección 3 · Filtros y etiquetas.
//
// La trampa: las claves y los valores de las etiquetas distinguen mayúsculas. Un filtro que
// no casa con nada no es un error: es una lista vacía. Y «no tiene la etiqueta» no se puede
// pedir con un filtro de EC2: hay que traerlas todas y mirarlo en Go.
package main

import (
	"fmt"

	"curso-aws/comun"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func contar(c *ec2.Client, fs ...types.Filter) int {
	return len(comun.Todas(c, &ec2.DescribeInstancesInput{Filters: fs}))
}

func f(nombre string, valores ...string) types.Filter {
	return types.Filter{Name: aws.String(nombre), Values: valores}
}

func main() {
	comun.Reset()
	c := comun.EC2("us-east-1")
	comun.Crear(c, 3, "Name", "web", "env", "prod", "dueno", "pagos")
	comun.Crear(c, 2, "Name", "api", "env", "dev", "dueno", "pagos")
	comun.Crear(c, 1, "Name", "prueba-juan", "Env", "dev") // alguien la etiquetó a mano
	fmt.Println("  6 instancias: 3 prod, 2 dev, y 1 que alguien etiqueto a mano (Env, sin dueno)")

	comun.Titulo("3.1 · El mismo filtro, escrito de tres maneras")
	fmt.Println("  tag:env = prod  ->", contar(c, f("tag:env", "prod")))
	fmt.Println("  tag:Env = prod  ->", contar(c, f("tag:Env", "prod")))
	fmt.Println("  tag:env = Prod  ->", contar(c, f("tag:env", "Prod")))

	comun.Titulo("3.2 · Cuantas son de dev?")
	fmt.Println("  tag:env = dev   ->", contar(c, f("tag:env", "dev")), "   (la de Env no sale)")

	comun.Titulo("3.3 · Varios filtros = Y; varios valores = O")
	fmt.Println("  env=dev,prod                   ->", contar(c, f("tag:env", "dev", "prod")))
	fmt.Println("  env=prod Y dueno=pagos         ->", contar(c, f("tag:env", "prod"), f("tag:dueno", "pagos")))

	comun.Titulo("3.4 · «Sin dueno» no es un filtro: se mira en Go")
	for _, i := range comun.Todas(c, &ec2.DescribeInstancesInput{}) {
		if comun.Etiqueta(i.Tags, "dueno") == "-" {
			fmt.Printf("  %s  Name=%s  -> sin etiqueta dueno\n", aws.ToString(i.InstanceId), comun.Etiqueta(i.Tags, "Name"))
		}
	}
}
