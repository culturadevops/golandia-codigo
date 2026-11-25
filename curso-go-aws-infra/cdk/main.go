// La app CDK del curso: tres pilas, en Go.
//
//	Red    una VPC. Por defecto, como la crea CDK; con -c nat=0, sin NAT gateways.
//	Datos  un bucket. -c id=Logs cambia su id de construcción; -c borrar=1 le pone DESTROY.
//	Web    una red mínima y un grupo de seguridad con el 443 abierto.
//
// Sintetizador «legacy»: no necesita `cdk bootstrap`, así que se despliega igual en el
// emulador que en una cuenta recién creada (mientras la plantilla no lleve ficheros).
package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func contexto(app constructs.Construct, k, porDefecto string) string {
	if v, ok := app.Node().TryGetContext(jsii.String(k)).(string); ok && v != "" {
		return v
	}
	return porDefecto
}

func pila(app awscdk.App, nombre string) awscdk.Stack {
	return awscdk.NewStack(app, jsii.String(nombre), &awscdk.StackProps{
		Synthesizer: awscdk.NewLegacyStackSynthesizer(),
	})
}

func main() {
	defer jsii.Close()
	app := awscdk.NewApp(nil)

	// ── Red ──
	red := pila(app, "Red")
	if contexto(app, "nat", "") == "0" {
		awsec2.NewVpc(red, jsii.String("Vpc"), &awsec2.VpcProps{
			NatGateways: jsii.Number(0),
			SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
				{Name: jsii.String("publica"), SubnetType: awsec2.SubnetType_PUBLIC},
				{Name: jsii.String("aislada"), SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED},
			},
		})
	} else {
		awsec2.NewVpc(red, jsii.String("Vpc"), nil) // ⭐ una línea
	}

	// ── Datos ──
	datos := pila(app, "Datos")
	props := &awss3.BucketProps{}
	if contexto(app, "borrar", "") == "1" {
		props.RemovalPolicy = awscdk.RemovalPolicy_DESTROY
	}
	awss3.NewBucket(datos, jsii.String(contexto(app, "id", "Registros")), props)

	// ── Web ──
	web := pila(app, "Web")
	vpc := awsec2.NewVpc(web, jsii.String("Vpc"), &awsec2.VpcProps{
		MaxAzs: jsii.Number(1), NatGateways: jsii.Number(0),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{Name: jsii.String("publica"), SubnetType: awsec2.SubnetType_PUBLIC},
		},
	})
	sg := awsec2.NewSecurityGroup(web, jsii.String("Web"), &awsec2.SecurityGroupProps{
		Vpc: vpc, Description: jsii.String("solo https"), AllowAllOutbound: jsii.Bool(true),
	})
	sg.AddIngressRule(awsec2.Peer_AnyIpv4(), awsec2.Port_Tcp(jsii.Number(443)), jsii.String("https"), nil)

	app.Synth(nil)
}
