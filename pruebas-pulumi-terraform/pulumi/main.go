package main

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/s3"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/sqs"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ssm"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type Ajustes struct {
	Entorno string
	Colas   []string
	Clave   pulumi.StringOutput
}

func crear(ctx *pulumi.Context, a Ajustes) error {
	datos, err := s3.NewBucket(ctx, "datos", &s3.BucketArgs{
		Bucket: pulumi.String("g680-datos-" + a.Entorno),
	})
	if err != nil {
		return err
	}

	// La condición: el versionado solo en prod
	if a.Entorno == "prod" {
		_, err = s3.NewBucketVersioning(ctx, "datos", &s3.BucketVersioningArgs{
			Bucket: datos.ID(),
			VersioningConfiguration: &s3.BucketVersioningVersioningConfigurationArgs{
				Status: pulumi.String("Enabled"),
			},
		})
		if err != nil {
			return err
		}
	}

	// El bucle: una cola por nombre
	var arns []interface{}
	urls := pulumi.StringMap{}
	for _, n := range a.Colas {
		c, err := sqs.NewQueue(ctx, "cola-"+n, &sqs.QueueArgs{
			Name: pulumi.String("g680-" + n),
		})
		if err != nil {
			return err
		}
		arns = append(arns, c.Arn)
		urls[n] = c.Url
	}

	// Los ARN todavía no existen: la política se arma cuando lleguen
	doc := pulumi.All(append(arns, datos.Arn)...).ApplyT(func(v []interface{}) (string, error) {
		colas := v[:len(v)-1]
		b, err := json.Marshal(map[string]any{
			"Version": "2012-10-17",
			"Statement": []map[string]any{
				{"Effect": "Allow", "Action": []string{"sqs:SendMessage"}, "Resource": colas},
				{"Effect": "Allow", "Action": []string{"s3:GetObject", "s3:PutObject"},
					"Resource": []string{fmt.Sprintf("%s/*", v[len(v)-1])}},
			},
		})
		return string(b), err
	}).(pulumi.StringOutput)

	pol, err := iam.NewPolicy(ctx, "app", &iam.PolicyArgs{
		Name:   pulumi.String("g680-app-" + a.Entorno),
		Policy: doc,
	})
	if err != nil {
		return err
	}

	_, err = ssm.NewParameter(ctx, "clave", &ssm.ParameterArgs{
		Name:  pulumi.String("/g680/" + a.Entorno + "/clave"),
		Type:  pulumi.String("SecureString"),
		Value: a.Clave,
	})
	if err != nil {
		return err
	}

	ctx.Export("bucket", datos.Bucket)
	ctx.Export("colas", urls)
	ctx.Export("politica", pol.Arn)
	return nil
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		var colas []string
		cfg.RequireObject("colas", &colas)
		return crear(ctx, Ajustes{
			Entorno: cfg.Require("entorno"),
			Colas:   colas,
			Clave:   cfg.RequireSecret("clave"),
		})
	})
}
