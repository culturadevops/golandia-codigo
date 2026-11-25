// Lección 9 · CDK en Go: lo que escribe UNA línea.
//
// awsec2.NewVpc(pila, "Vpc", nil) parece una red. Es una red con dos zonas, subredes públicas
// y privadas, y un NAT gateway en cada zona para que las privadas salgan a internet. Los NAT
// gateways se cobran por hora aunque no pase ni un byte. cdk synth te lo enseña ANTES de pagar.
package main

import (
	"fmt"

	"curso-aws/comun"
)

// Precio público (aws.amazon.com/vpc/pricing, ejemplo de US East): 0,045 USD por hora y por
// NAT gateway, más 0,045 USD por GB procesado. Un mes = 730 horas.
const usdHoraNAT = 0.045

func informe(pila string) {
	total, tipos := comun.Recursos(pila)
	fmt.Printf("  %d recursos de AWS:\n", total)
	nat := 0
	for _, t := range tipos {
		fmt.Printf("    %2d  %s\n", t[1], t[0])
		if t[0] == "AWS::EC2::NatGateway" {
			nat = t[1].(int)
		}
	}
	fmt.Printf("  NAT gateways: %d  ->  %d x 0.045 USD/h x 730 h = %.2f USD al mes, sin trafico\n",
		nat, nat, float64(nat)*usdHoraNAT*730)
}

func main() {
	comun.Titulo("9.1 · awsec2.NewVpc(pila, \"Vpc\", nil)")
	fmt.Println(comun.CDK("synth", "-q"))
	informe("Red")

	comun.Titulo("9.2 · La misma VPC, sin NAT: subredes publicas y aisladas")
	fmt.Println(comun.CDK("synth", "-q", "-c", "nat=0"))
	informe("Red")
}
