// Lección 11 · cdk destroy no borra todo: RemovalPolicy.
//
// Para lo que guarda datos (buckets, tablas, volúmenes) CDK elige por defecto RETAIN: al
// destruir la pila, el recurso se queda. Es una buena protección y una mala sorpresa: la pila
// dice «destroyed» y el bucket sigue ahí, sin pila que lo recuerde.
package main

import (
	"fmt"

	"curso-aws/comun"
)

func politica() {
	for id, v := range comun.Plantilla("Datos")["Resources"].(map[string]any) {
		m := v.(map[string]any)
		if m["Type"] == "AWS::S3::Bucket" {
			fmt.Printf("  plantilla: %s  DeletionPolicy=%v\n", id, m["DeletionPolicy"])
		}
	}
}

func main() {
	comun.Reset()

	comun.Titulo("11.1 · Por defecto: RETAIN")
	fmt.Println(comun.CDK("deploy", "Datos", "--method=direct", "--require-approval", "never"))
	politica()
	fmt.Println(comun.CDK("destroy", "Datos", "-f"))
	fmt.Println("  buckets que siguen existiendo:", comun.Buckets())

	comun.Titulo("11.2 · Con RemovalPolicy_DESTROY (-c borrar=1)")
	comun.Reset()
	fmt.Println(comun.CDK("deploy", "Datos", "-c", "borrar=1", "--method=direct", "--require-approval", "never"))
	politica()
	fmt.Println(comun.CDK("destroy", "Datos", "-c", "borrar=1", "-f"))
	fmt.Println("  buckets que siguen existiendo:", comun.Buckets())
	fmt.Println("  (con datos dentro, S3 se niega a borrar el bucket: haria falta AutoDeleteObjects)")
}
