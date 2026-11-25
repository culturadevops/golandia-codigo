// Package comun: la configuración del SDK, la salvaguarda del emulador y ayudas de las lecciones.
//
// ⛔ Las lecciones crean, paran y borran instancias. Por eso Config() se niega a devolver nada
// si no está apuntando a un emulador en tu máquina (AWS_ENDPOINT_URL en localhost). Contra una
// cuenta de verdad, este curso no arranca. Es la lección 1 aplicada al propio curso.
package comun

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
)

// Endpoint devuelve AWS_ENDPOINT_URL. El SDK v2 la lee solo: el código de las lecciones es
// el mismo que irá contra AWS de verdad.
func Endpoint() string { return os.Getenv("AWS_ENDPOINT_URL") }

func esLocal() bool {
	u, err := url.Parse(Endpoint())
	if err != nil || u.Host == "" {
		return false
	}
	h := u.Hostname()
	return h == "localhost" || h == "127.0.0.1"
}

// Config carga la configuración para una región, solo si apunta al emulador.
func Config(region string) aws.Config {
	if !esLocal() {
		log.Fatalf("⛔ AWS_ENDPOINT_URL=%q no es un emulador local. Este curso para y borra "+
			"instancias: no lo ejecuto contra una cuenta de verdad.", Endpoint())
	}
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}

// EC2 es un cliente de EC2 para una región.
func EC2(region string) *ec2.Client { return ec2.NewFromConfig(Config(region)) }

// Reset vacía el emulador (moto) para que cada lección empiece del mismo estado.
func Reset() {
	Config("us-east-1") // la misma salvaguarda: fuera del emulador, aquí se para
	r, err := http.Post(Endpoint()+"/moto-api/reset", "application/json", nil)
	if err != nil {
		log.Fatal(err)
	}
	r.Body.Close()
}

// Titulo imprime la cabecera de cada parte de una lección.
func Titulo(s string) { fmt.Printf("\n== %s ==\n", s) }

// AMI devuelve una imagen cualquiera del emulador.
func AMI(c *ec2.Client) string {
	out, err := c.DescribeImages(context.Background(), &ec2.DescribeImagesInput{
		Filters: []types.Filter{{Name: aws.String("name"), Values: []string{"amzn2-ami-*"}}},
	})
	if err != nil || len(out.Images) == 0 {
		log.Fatal("sin AMI: ", err)
	}
	return *out.Images[0].ImageId
}

// Tags convierte pares clave=valor en etiquetas.
func Tags(kv ...string) []types.Tag {
	var t []types.Tag
	for i := 0; i+1 < len(kv); i += 2 {
		t = append(t, types.Tag{Key: aws.String(kv[i]), Value: aws.String(kv[i+1])})
	}
	return t
}

// Crear lanza n instancias con etiquetas puestas EN la creación (lección 4).
func Crear(c *ec2.Client, n int32, kv ...string) []types.Instance {
	out, err := c.RunInstances(context.Background(), &ec2.RunInstancesInput{
		ImageId: aws.String(AMI(c)), InstanceType: types.InstanceTypeT3Micro,
		MinCount: aws.Int32(n), MaxCount: aws.Int32(n),
		TagSpecifications: []types.TagSpecification{{ResourceType: types.ResourceTypeInstance, Tags: Tags(kv...)}},
	})
	if err != nil {
		log.Fatal(err)
	}
	return out.Instances
}

// Etiqueta devuelve el valor de una etiqueta, o "-".
func Etiqueta(tags []types.Tag, k string) string {
	for _, t := range tags {
		if aws.ToString(t.Key) == k {
			return aws.ToString(t.Value)
		}
	}
	return "-"
}

// Todas recorre TODAS las páginas de DescribeInstances y devuelve las instancias (lección 2).
func Todas(c *ec2.Client, in *ec2.DescribeInstancesInput) []types.Instance {
	var out []types.Instance
	p := ec2.NewDescribeInstancesPaginator(c, in)
	for p.HasMorePages() {
		pg, err := p.NextPage(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		for _, r := range pg.Reservations {
			out = append(out, r.Instances...)
		}
	}
	return out
}

// Codigo saca el código de error de la API de AWS (lección 5).
func Codigo(err error) string {
	var ae smithy.APIError
	if errors.As(err, &ae) {
		return ae.ErrorCode()
	}
	if err == nil {
		return "<nil>"
	}
	return err.Error()
}
