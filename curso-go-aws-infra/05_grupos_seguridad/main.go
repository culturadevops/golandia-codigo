// Lección 5 · Grupos de seguridad, y los errores de la API.
//
// La trampa: tratar todos los errores igual. Añadir una regla que ya existe devuelve un error
// (InvalidPermission.Duplicate) que para un script idempotente significa «ya está hecho».
// Para distinguirlo hay que sacar el CÓDIGO del error, no comparar el texto.
package main

import (
	"context"
	"fmt"

	"curso-aws/comun"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func regla(puerto int32, cidr string) []types.IpPermission {
	return []types.IpPermission{{IpProtocol: aws.String("tcp"), FromPort: aws.Int32(puerto),
		ToPort: aws.Int32(puerto), IpRanges: []types.IpRange{{CidrIp: aws.String(cidr)}}}}
}

// abrir es idempotente: si la regla ya estaba, no es un fallo.
func abrir(c *ec2.Client, sg string, puerto int32, cidr string) error {
	_, err := c.AuthorizeSecurityGroupIngress(context.Background(),
		&ec2.AuthorizeSecurityGroupIngressInput{GroupId: &sg, IpPermissions: regla(puerto, cidr)})
	if comun.Codigo(err) == "InvalidPermission.Duplicate" {
		return nil
	}
	return err
}

func main() {
	ctx := context.Background()
	comun.Reset()
	c := comun.EC2("us-east-1")
	vpc, _ := c.CreateVpc(ctx, &ec2.CreateVpcInput{CidrBlock: aws.String("10.0.0.0/16")})
	crear := func(n string) string {
		g, _ := c.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{GroupName: aws.String(n),
			Description: aws.String(n), VpcId: vpc.Vpc.VpcId})
		return aws.ToString(g.GroupId)
	}
	web, bastion, db := crear("web"), crear("bastion"), crear("db")

	comun.Titulo("5.1 · Abrir 443 al mundo en web, dos veces")
	_, err := c.AuthorizeSecurityGroupIngress(ctx, &ec2.AuthorizeSecurityGroupIngressInput{GroupId: &web, IpPermissions: regla(443, "0.0.0.0/0")})
	fmt.Println("  primera vez ->", err)
	_, err = c.AuthorizeSecurityGroupIngress(ctx, &ec2.AuthorizeSecurityGroupIngressInput{GroupId: &web, IpPermissions: regla(443, "0.0.0.0/0")})
	fmt.Println("  segunda vez ->", err)

	comun.Titulo("5.2 · El codigo del error, con errors.As y smithy.APIError")
	fmt.Println("  comun.Codigo(err) ->", comun.Codigo(err))
	fmt.Println("  abrir() idempotente, segunda vez ->", abrir(c, web, 443, "0.0.0.0/0"))

	comun.Titulo("5.3 · Auditoria: puertos de administracion abiertos al mundo")
	abrir(c, bastion, 22, "0.0.0.0/0") // «era solo un momento, para probar»
	abrir(c, db, 5432, "10.0.0.0/16")  // bien: solo desde la red interna
	abrir(c, db, 3389, "0.0.0.0/0")    // alguien entró por escritorio remoto
	peligrosos := map[int32]string{22: "ssh", 3389: "rdp", 5432: "postgres", 3306: "mysql"}
	out, _ := c.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{})
	for _, g := range out.SecurityGroups {
		for _, p := range g.IpPermissions {
			for _, r := range p.IpRanges {
				if aws.ToString(r.CidrIp) != "0.0.0.0/0" {
					continue
				}
				if nombre, ok := peligrosos[aws.ToInt32(p.FromPort)]; ok {
					fmt.Printf("  ⛔ %-8s %-22s puerto %-5d (%s) abierto a 0.0.0.0/0\n",
						aws.ToString(g.GroupName), aws.ToString(g.GroupId), aws.ToInt32(p.FromPort), nombre)
				}
			}
		}
	}

	comun.Titulo("5.4 · Cerrar")
	_, err = c.RevokeSecurityGroupIngress(ctx, &ec2.RevokeSecurityGroupIngressInput{GroupId: &bastion, IpPermissions: regla(22, "0.0.0.0/0")})
	fmt.Println("  revocar 22 en bastion ->", err)
}
