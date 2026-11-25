// Lección 8 · El filtro vacío que selecciona TODO.
//
// La trampa más cara del curso. Un script para parar las máquinas de desarrollo por la noche.
// El entorno llega por una variable. Si esa variable viene vacía, el filtro no se añade, y
// DescribeInstances SIN filtros devuelve todas las instancias de la región. Producción incluida.
package main

import (
	"context"
	"errors"
	"fmt"

	"curso-aws/comun"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func seleccionar(c *ec2.Client, entorno string) []types.Instance {
	var filtros []types.Filter
	if entorno != "" {
		filtros = append(filtros, types.Filter{Name: aws.String("tag:env"), Values: []string{entorno}})
	}
	filtros = append(filtros, types.Filter{Name: aws.String("instance-state-name"), Values: []string{"running"}})
	return comun.Todas(c, &ec2.DescribeInstancesInput{Filters: filtros})
}

func pararMal(c *ec2.Client, entorno string) {
	is := seleccionar(c, entorno)
	var ids []string
	for _, i := range is {
		ids = append(ids, aws.ToString(i.InstanceId))
		fmt.Printf("    parando %s  Name=%-8s env=%s\n", aws.ToString(i.InstanceId), comun.Etiqueta(i.Tags, "Name"), comun.Etiqueta(i.Tags, "env"))
	}
	c.StopInstances(context.Background(), &ec2.StopInstancesInput{InstanceIds: ids})
}

var errSinFiltro = errors.New("sin entorno no paro nada: diria que TODAS")

// pararBien: sin entorno no hace nada, comprueba lo seleccionado y, por defecto, solo enseña el plan.
func pararBien(c *ec2.Client, entorno string, aplicar bool) error {
	if entorno == "" {
		return errSinFiltro
	}
	is := seleccionar(c, entorno)
	for _, i := range is {
		if e := comun.Etiqueta(i.Tags, "env"); e != entorno {
			return fmt.Errorf("%s tiene env=%s, no %s: aborto", aws.ToString(i.InstanceId), e, entorno)
		}
	}
	accion := "PLAN (sin -aplicar)"
	if aplicar {
		accion = "parando"
	}
	for _, i := range is {
		fmt.Printf("    %s %s  Name=%s\n", accion, aws.ToString(i.InstanceId), comun.Etiqueta(i.Tags, "Name"))
	}
	return nil
}

func porEntorno(c *ec2.Client) {
	cuenta := map[string]map[string]int{}
	for _, i := range comun.Todas(c, &ec2.DescribeInstancesInput{}) {
		e := comun.Etiqueta(i.Tags, "env")
		if cuenta[e] == nil {
			cuenta[e] = map[string]int{}
		}
		cuenta[e][string(i.State.Name)]++
	}
	for _, e := range []string{"prod", "dev"} {
		fmt.Printf("    %-5s running=%d stopped=%d\n", e, cuenta[e]["running"], cuenta[e]["stopped"]+cuenta[e]["stopping"])
	}
}

func escenario(c *ec2.Client) {
	comun.Reset()
	comun.Crear(c, 3, "Name", "pagos", "env", "prod", "dueno", "pagos")
	comun.Crear(c, 3, "Name", "pruebas", "env", "dev", "dueno", "pagos")
	w := ec2.NewInstanceRunningWaiter(c)
	w.Wait(context.Background(), &ec2.DescribeInstancesInput{}, 60e9)
}

func main() {
	c := comun.EC2("us-east-1")

	comun.Titulo("8.1 · parar(ENTORNO) con ENTORNO=dev")
	escenario(c)
	pararMal(c, "dev")
	porEntorno(c)

	comun.Titulo("8.2 · El mismo script, y la variable llega vacia")
	escenario(c)
	pararMal(c, "")
	porEntorno(c)

	comun.Titulo("8.3 · Con tres protecciones")
	escenario(c)
	fmt.Println("  entorno vacio ->", pararBien(c, "", true))
	fmt.Println("  entorno dev, sin -aplicar:")
	pararBien(c, "dev", false)
	porEntorno(c)
}
