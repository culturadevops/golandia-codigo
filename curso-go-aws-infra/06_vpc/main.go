// Lección 6 · Una red desde cero, y el orden para borrarla.
//
// Crear una VPC útil son seis llamadas: VPC, subred, internet gateway, engancharlo, tabla de
// rutas con su ruta por defecto, y asociarla a la subred. Borrarla es el mismo camino al revés:
// si empiezas por la VPC, AWS se niega (DependencyViolation) y no te dice qué sobra.
package main

import (
	"context"
	"fmt"

	"curso-aws/comun"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func paso(n int, que string, err error) {
	fmt.Printf("  %d. %-42s -> %s\n", n, que, comun.Codigo(err))
}

func main() {
	ctx := context.Background()
	comun.Reset()
	c := comun.EC2("us-east-1")

	comun.Titulo("6.1 · Crear, en orden")
	v, err := c.CreateVpc(ctx, &ec2.CreateVpcInput{CidrBlock: aws.String("10.20.0.0/16"),
		TagSpecifications: nil})
	paso(1, "CreateVpc 10.20.0.0/16", err)
	vpc := v.Vpc.VpcId
	s, err := c.CreateSubnet(ctx, &ec2.CreateSubnetInput{VpcId: vpc, CidrBlock: aws.String("10.20.1.0/24")})
	paso(2, "CreateSubnet 10.20.1.0/24", err)
	ig, err := c.CreateInternetGateway(ctx, &ec2.CreateInternetGatewayInput{})
	paso(3, "CreateInternetGateway", err)
	_, err = c.AttachInternetGateway(ctx, &ec2.AttachInternetGatewayInput{InternetGatewayId: ig.InternetGateway.InternetGatewayId, VpcId: vpc})
	paso(4, "AttachInternetGateway", err)
	rt, err := c.CreateRouteTable(ctx, &ec2.CreateRouteTableInput{VpcId: vpc})
	paso(5, "CreateRouteTable", err)
	_, err = c.CreateRoute(ctx, &ec2.CreateRouteInput{RouteTableId: rt.RouteTable.RouteTableId,
		DestinationCidrBlock: aws.String("0.0.0.0/0"), GatewayId: ig.InternetGateway.InternetGatewayId})
	paso(6, "CreateRoute 0.0.0.0/0 -> internet gateway", err)
	as, err := c.AssociateRouteTable(ctx, &ec2.AssociateRouteTableInput{RouteTableId: rt.RouteTable.RouteTableId, SubnetId: s.Subnet.SubnetId})
	paso(7, "AssociateRouteTable con la subred", err)

	comun.Titulo("6.2 · Borrar empezando por la VPC")
	_, err = c.DeleteVpc(ctx, &ec2.DeleteVpcInput{VpcId: vpc})
	fmt.Println("  DeleteVpc ->", comun.Codigo(err))
	fmt.Println("  ", err)

	comun.Titulo("6.3 · Borrar al reves de como se creo")
	_, err = c.DisassociateRouteTable(ctx, &ec2.DisassociateRouteTableInput{AssociationId: as.AssociationId})
	paso(7, "DisassociateRouteTable", err)
	_, err = c.DeleteRouteTable(ctx, &ec2.DeleteRouteTableInput{RouteTableId: rt.RouteTable.RouteTableId})
	paso(6, "DeleteRouteTable (con su ruta)", err)
	_, err = c.DetachInternetGateway(ctx, &ec2.DetachInternetGatewayInput{InternetGatewayId: ig.InternetGateway.InternetGatewayId, VpcId: vpc})
	paso(4, "DetachInternetGateway", err)
	_, err = c.DeleteInternetGateway(ctx, &ec2.DeleteInternetGatewayInput{InternetGatewayId: ig.InternetGateway.InternetGatewayId})
	paso(3, "DeleteInternetGateway", err)
	_, err = c.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{SubnetId: s.Subnet.SubnetId})
	paso(2, "DeleteSubnet", err)
	_, err = c.DeleteVpc(ctx, &ec2.DeleteVpcInput{VpcId: vpc})
	paso(1, "DeleteVpc", err)
	fmt.Println("\n  Seis recursos para una red minima. Esta cuenta es la que hace el CDK por ti (leccion 9).")
}
