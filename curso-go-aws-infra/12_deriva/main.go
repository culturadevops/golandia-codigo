// Lección 12 · La deriva que cdk diff no ve.
//
// cdk diff compara tu código con la ÚLTIMA PLANTILLA desplegada, no con lo que hay de verdad
// en la cuenta. Si alguien abre el puerto 22 a mano desde la consola, la plantilla no cambia,
// y cdk diff dice que no hay diferencias. Para ver la deriva hay que preguntar a la realidad.
package main

import (
	"context"
	"fmt"
	"sort"

	"curso-aws/comun"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// grupoDeLaPila busca el id real del grupo de seguridad que creó la pila.
func grupoDeLaPila() string {
	cf := cloudformation.NewFromConfig(comun.Config("us-east-1"))
	out, err := cf.DescribeStackResources(context.Background(), &cloudformation.DescribeStackResourcesInput{StackName: aws.String("Web")})
	if err != nil {
		panic(err)
	}
	for _, r := range out.StackResources {
		if aws.ToString(r.ResourceType) == "AWS::EC2::SecurityGroup" {
			return aws.ToString(r.PhysicalResourceId)
		}
	}
	return ""
}

// declaradas: las reglas de entrada que dice la plantilla.
func declaradas() []string {
	var out []string
	for _, v := range comun.Plantilla("Web")["Resources"].(map[string]any) {
		m := v.(map[string]any)
		if m["Type"] != "AWS::EC2::SecurityGroup" {
			continue
		}
		p := m["Properties"].(map[string]any)
		for _, r := range p["SecurityGroupIngress"].([]any) {
			rr := r.(map[string]any)
			out = append(out, fmt.Sprintf("tcp/%v desde %v", rr["FromPort"], rr["CidrIp"]))
		}
	}
	sort.Strings(out)
	return out
}

// reales: las reglas de entrada que tiene el grupo de verdad.
func reales(c *ec2.Client, sg string) []string {
	g, _ := c.DescribeSecurityGroups(context.Background(), &ec2.DescribeSecurityGroupsInput{GroupIds: []string{sg}})
	var out []string
	for _, p := range g.SecurityGroups[0].IpPermissions {
		for _, r := range p.IpRanges {
			out = append(out, fmt.Sprintf("%s/%d desde %s", aws.ToString(p.IpProtocol), aws.ToInt32(p.FromPort), aws.ToString(r.CidrIp)))
		}
	}
	sort.Strings(out)
	return out
}

func main() {
	comun.Reset()
	c := comun.EC2("us-east-1")

	comun.Titulo("12.1 · Desplegar la pila Web: un grupo con solo el 443")
	fmt.Println(comun.CDK("deploy", "Web", "--method=direct", "--require-approval", "never"))
	sg := grupoDeLaPila()

	comun.Titulo("12.2 · Alguien abre el 22 «un momento», a mano")
	_, err := c.AuthorizeSecurityGroupIngress(context.Background(), &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: &sg, IpPermissions: []types.IpPermission{{IpProtocol: aws.String("tcp"),
			FromPort: aws.Int32(22), ToPort: aws.Int32(22), IpRanges: []types.IpRange{{CidrIp: aws.String("0.0.0.0/0")}}}}})
	fmt.Println("  AuthorizeSecurityGroupIngress 22 0.0.0.0/0 ->", err)

	comun.Titulo("12.3 · cdk diff")
	fmt.Println(comun.CDK("diff", "Web", "--method=template"))

	comun.Titulo("12.4 · Preguntar a la realidad: plantilla contra cuenta")
	decl, real := declaradas(), reales(c, sg)
	fmt.Println("  declaradas:", decl)
	fmt.Println("  reales:    ", real)
	en := map[string]bool{}
	for _, d := range decl {
		en[d] = true
	}
	for _, r := range real {
		if !en[r] {
			fmt.Printf("  ⛔ DERIVA en %s: %s no esta en el codigo\n", sg, r)
		}
	}
}
