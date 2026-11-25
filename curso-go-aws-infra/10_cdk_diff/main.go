// Lección 10 · cdk diff: cambiar un nombre no es renombrar.
//
// El id de construcción ("Registros") forma parte del nombre lógico del recurso en la
// plantilla. Si lo cambias, para CloudFormation es OTRO recurso: crea uno nuevo y quita el
// viejo. Con un bucket, que por defecto lleva DeletionPolicy: Retain, el viejo no se borra:
// se queda huérfano, fuera de la pila, con sus datos y su factura. cdk diff lo avisa.
package main

import (
	"fmt"

	"curso-aws/comun"
)

func main() {
	comun.Reset()

	comun.Titulo("10.1 · Desplegar la pila Datos, con el bucket «Registros»")
	fmt.Println(comun.CDK("deploy", "Datos", "--method=direct", "--require-approval", "never"))
	fmt.Println("  buckets que existen:", comun.Buckets())

	comun.Titulo("10.2 · En el codigo, el id pasa a ser «Logs». Antes de desplegar: cdk diff")
	fmt.Println(comun.CDK("diff", "Datos", "-c", "id=Logs", "--method=template"))

	comun.Titulo("10.3 · Por que «orphan»: lo que dice la plantilla del bucket")
	r := comun.Plantilla("Datos")["Resources"].(map[string]any)
	for id, v := range r {
		m := v.(map[string]any)
		if m["Type"] == "AWS::S3::Bucket" {
			fmt.Printf("  %s  DeletionPolicy=%v  UpdateReplacePolicy=%v\n", id, m["DeletionPolicy"], m["UpdateReplacePolicy"])
		}
	}
}
