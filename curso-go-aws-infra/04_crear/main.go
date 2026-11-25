// Lección 4 · Crear sin dejar huérfanos ni duplicados.
//
// Tres trampas:
//  1. Etiquetar DESPUÉS de crear: si el script muere entre las dos llamadas, queda una
//     instancia sin etiquetas que nadie sabe de quién es. Las etiquetas van EN RunInstances.
//  2. Reintentar tras un timeout: el primer intento sí llegó, y ahora hay dos.
//  3. Actuar sobre una instancia recién creada: está «pending», no «running».
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"curso-aws/comun"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

var errTimeout = errors.New("timeout esperando la respuesta") // simulado

// crearMal: crea y luego etiqueta. Si «muere» entre medias, la instancia queda sin nombre.
func crearMal(c *ec2.Client, nombre string, muere bool) error {
	out, err := c.RunInstances(context.Background(), &ec2.RunInstancesInput{
		ImageId: aws.String(comun.AMI(c)), MinCount: aws.Int32(1), MaxCount: aws.Int32(1)})
	if err != nil {
		return err
	}
	if muere {
		return errTimeout
	}
	_, err = c.CreateTags(context.Background(), &ec2.CreateTagsInput{
		Resources: []string{*out.Instances[0].InstanceId}, Tags: comun.Tags("Name", nombre)})
	return err
}

// crearBien busca por nombre antes de crear, y etiqueta en la misma llamada.
func crearBien(c *ec2.Client, nombre string) (string, bool) {
	ya := comun.Todas(c, &ec2.DescribeInstancesInput{Filters: []types.Filter{
		{Name: aws.String("tag:Name"), Values: []string{nombre}},
		{Name: aws.String("instance-state-name"), Values: []string{"pending", "running", "stopping", "stopped"}},
	}})
	if len(ya) > 0 {
		return aws.ToString(ya[0].InstanceId), false
	}
	i := comun.Crear(c, 1, "Name", nombre, "dueno", "pagos")
	return aws.ToString(i[0].InstanceId), true
}

func nombres(c *ec2.Client) {
	for _, i := range comun.Todas(c, &ec2.DescribeInstancesInput{}) {
		fmt.Printf("    %s  Name=%s\n", aws.ToString(i.InstanceId), comun.Etiqueta(i.Tags, "Name"))
	}
}

func main() {
	comun.Reset()
	c := comun.EC2("us-east-1")

	comun.Titulo("4.1 · Crear y DESPUES etiquetar; el primer intento «da timeout»")
	fmt.Println("  intento 1 ->", crearMal(c, "web-nueva", true))
	fmt.Println("  intento 2 ->", crearMal(c, "web-nueva", false))
	fmt.Println("  en la cuenta:")
	nombres(c)

	comun.Titulo("4.2 · Buscar por nombre antes de crear, etiquetas EN la creacion")
	comun.Reset()
	for i := 1; i <= 2; i++ {
		id, nueva := crearBien(c, "web-nueva")
		if nueva {
			fmt.Printf("  intento %d -> creada %s\n", i, id)
		} else {
			fmt.Printf("  intento %d -> ya existe %s: no creo nada\n", i, id)
		}
	}
	fmt.Println("  en la cuenta:")
	nombres(c)

	comun.Titulo("4.3 · Recien creada no es lista")
	i := comun.Crear(c, 1, "Name", "db-nueva", "dueno", "datos")
	id := aws.ToString(i[0].InstanceId)
	fmt.Println("  estado al volver de RunInstances:", i[0].State.Name)
	t0 := time.Now()
	w := ec2.NewInstanceRunningWaiter(c)
	err := w.Wait(context.Background(), &ec2.DescribeInstancesInput{InstanceIds: []string{id}}, 5*time.Minute)
	d := comun.Todas(c, &ec2.DescribeInstancesInput{InstanceIds: []string{id}})
	fmt.Printf("  tras NewInstanceRunningWaiter:    %s   (err = %v, %.1f s en el emulador)\n",
		d[0].State.Name, err, time.Since(t0).Seconds())
}
