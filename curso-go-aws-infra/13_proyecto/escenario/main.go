// escenario le prepara a inventario una cuenta con lo que se encuentra de verdad: una
// instancia sin dueño, un bastión abierto al mundo, un volumen suelto, una IP reservada sin
// usar y un NAT gateway. En us-west-2, todo en orden.
package main

import (
	"context"
	"fmt"

	"curso-aws/comun"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func main() {
	ctx := context.Background()
	comun.Reset()
	c := comun.EC2("us-east-1")
	comun.Crear(c, 3, "Name", "web", "env", "prod", "dueno", "pagos")
	comun.Crear(c, 1, "Name", "prueba-juan", "env", "dev")

	vpc, _ := c.CreateVpc(ctx, &ec2.CreateVpcInput{CidrBlock: aws.String("10.0.0.0/16")})
	sub, _ := c.CreateSubnet(ctx, &ec2.CreateSubnetInput{VpcId: vpc.Vpc.VpcId, CidrBlock: aws.String("10.0.1.0/24")})
	g, _ := c.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{GroupName: aws.String("bastion"),
		Description: aws.String("bastion"), VpcId: vpc.Vpc.VpcId})
	c.AuthorizeSecurityGroupIngress(ctx, &ec2.AuthorizeSecurityGroupIngressInput{GroupId: g.GroupId,
		IpPermissions: []types.IpPermission{{IpProtocol: aws.String("tcp"), FromPort: aws.Int32(22), ToPort: aws.Int32(22),
			IpRanges: []types.IpRange{{CidrIp: aws.String("0.0.0.0/0")}}}}})

	c.CreateVolume(ctx, &ec2.CreateVolumeInput{AvailabilityZone: aws.String("us-east-1a"), Size: aws.Int32(100)})
	c.AllocateAddress(ctx, &ec2.AllocateAddressInput{Domain: types.DomainTypeVpc})
	eip, _ := c.AllocateAddress(ctx, &ec2.AllocateAddressInput{Domain: types.DomainTypeVpc})
	c.CreateNatGateway(ctx, &ec2.CreateNatGatewayInput{SubnetId: sub.Subnet.SubnetId, AllocationId: eip.AllocationId})

	oeste := comun.EC2("us-west-2")
	comun.Crear(oeste, 2, "Name", "replica", "env", "prod", "dueno", "datos")
	fmt.Println("escenario listo")
}
