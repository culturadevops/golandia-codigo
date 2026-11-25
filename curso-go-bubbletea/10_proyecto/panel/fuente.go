// Package panel: nubepanel, un panel de terminal para nubecita (o para AWS de verdad).
//
// Este fichero es lo único que sabe de AWS. El modelo (modelo.go) solo conoce la interfaz
// Fuente: así se prueba sin red, igual que en la lección 9.
package panel

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// Recurso es una fila del panel.
type Recurso struct {
	Nombre string
	Cuenta int    // objetos, mensajes o elementos
	Clave  string // lo que hace falta para actuar sobre él (la URL de una cola)
}

// Fuente es de donde salen los datos.
type Fuente interface {
	Buckets(ctx context.Context) ([]Recurso, error)
	Colas(ctx context.Context) ([]Recurso, error)
	Tablas(ctx context.Context) ([]Recurso, error)
	Purgar(ctx context.Context, cola Recurso) error
}

// AWS habla con cualquier endpoint compatible: nubecita, LocalStack, moto o Amazon.
type AWS struct {
	s3  *s3.Client
	sqs *sqs.Client
	ddb *dynamodb.Client
}

// NuevaAWS apunta los tres clientes al endpoint dado.
func NuevaAWS(endpoint string) *AWS {
	cfg := aws.Config{
		Region:       "us-east-1",
		Credentials:  credentials.NewStaticCredentialsProvider("prueba", "prueba", ""),
		BaseEndpoint: aws.String(endpoint),
		// El SDK reintenta 3 veces con espera. En un panel que ya se refresca solo cada 2 s,
		// eso solo retrasa el aviso de «sin conexión»: un intento y que lo diga.
		RetryMaxAttempts: 1,
	}
	return &AWS{
		s3:  s3.NewFromConfig(cfg, func(o *s3.Options) { o.UsePathStyle = true }),
		sqs: sqs.NewFromConfig(cfg),
		ddb: dynamodb.NewFromConfig(cfg),
	}
}

func (a *AWS) Buckets(ctx context.Context) ([]Recurso, error) {
	r, err := a.s3.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}
	var out []Recurso
	for _, b := range r.Buckets {
		n := 0
		p := s3.NewListObjectsV2Paginator(a.s3, &s3.ListObjectsV2Input{Bucket: b.Name})
		for p.HasMorePages() {
			pg, err := p.NextPage(ctx)
			if err != nil {
				return nil, err
			}
			n += len(pg.Contents)
		}
		out = append(out, Recurso{Nombre: aws.ToString(b.Name), Cuenta: n})
	}
	return out, nil
}

func (a *AWS) Colas(ctx context.Context) ([]Recurso, error) {
	r, err := a.sqs.ListQueues(ctx, &sqs.ListQueuesInput{})
	if err != nil {
		return nil, err
	}
	var out []Recurso
	for _, u := range r.QueueUrls {
		at, err := a.sqs.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
			QueueUrl:       aws.String(u),
			AttributeNames: []sqstypes.QueueAttributeName{sqstypes.QueueAttributeNameApproximateNumberOfMessages},
		})
		if err != nil {
			return nil, err
		}
		n, _ := strconv.Atoi(at.Attributes[string(sqstypes.QueueAttributeNameApproximateNumberOfMessages)])
		out = append(out, Recurso{Nombre: u[strings.LastIndex(u, "/")+1:], Cuenta: n, Clave: u})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Nombre < out[j].Nombre })
	return out, nil
}

func (a *AWS) Tablas(ctx context.Context) ([]Recurso, error) {
	r, err := a.ddb.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		return nil, err
	}
	var out []Recurso
	for _, t := range r.TableNames {
		d, err := a.ddb.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(t)})
		if err != nil {
			return nil, err
		}
		out = append(out, Recurso{Nombre: t, Cuenta: int(aws.ToInt64(d.Table.ItemCount))})
	}
	return out, nil
}

func (a *AWS) Purgar(ctx context.Context, cola Recurso) error {
	if cola.Clave == "" {
		return fmt.Errorf("%s no es una cola", cola.Nombre)
	}
	_, err := a.sqs.PurgeQueue(ctx, &sqs.PurgeQueueInput{QueueUrl: aws.String(cola.Clave)})
	return err
}
