// supresion: crea un detector y las reglas de supresión de los ejemplos de la documentación de
// GuardDuty, con aws-sdk-go-v2. Se corre contra un emulador (moto) con AWS_ENDPOINT_URL: no hace
// falta una cuenta de AWS.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"maps"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"

	"github.com/culturadevops/pruebas-guardduty-go/supresion"
)

func main() {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}
	gd := guardduty.NewFromConfig(cfg)

	// 1. El detector: uno por cuenta y por región. Todo cuelga de su ID.
	det, err := gd.CreateDetector(ctx, &guardduty.CreateDetectorInput{Enable: aws.Bool(true)})
	if err != nil {
		log.Fatal("CreateDetector: ", err)
	}
	id := aws.ToString(det.DetectorId)
	fmt.Printf("detector: %s (%d caracteres)\n", id, len(id))

	lista, err := gd.ListDetectors(ctx, &guardduty.ListDetectorsInput{})
	if err != nil {
		log.Fatal("ListDetectors: ", err)
	}
	fmt.Printf("detectores en esta región: %d\n\n", len(lista.DetectorIds))

	// 2. Las reglas. Las dos primeras son los ejemplos de la documentación; las otras dos, las
	//    que la regla de oro prohíbe, y nunca llegan a la API.
	reglas := []supresion.Regla{
		{Nombre: "escaner-de-vulnerabilidades", Descripcion: "el escáner genera port scan",
			Tipo:    "Recon:EC2/Portscan",
			Recurso: map[string]string{"resource.instanceDetails.imageId": "ami-999999999"}},
		{Nombre: "bastion-ssh", Descripcion: "el bastión recibe fuerza bruta: es su trabajo",
			Tipo:    "UnauthorizedAccess:EC2/SSHBruteForce",
			Recurso: map[string]string{"resource.instanceDetails.tags.value": "devops"}},
		{Nombre: "todo-el-portscan", Tipo: "Recon:EC2/Portscan"},
		{Nombre: "todo-kubernetes", Tipo: "*:Kubernetes/*",
			Recurso: map[string]string{"resource.resourceType": "EKSCluster"}},
	}
	var creadas []string
	for i, r := range reglas {
		in, err := r.Entrada(id, int32(i+1))
		if err != nil {
			fmt.Printf("✗ rechazada antes de llamar a la API · %v\n", err)
			continue
		}
		out, err := gd.CreateFilter(ctx, in)
		if err != nil {
			log.Fatal("CreateFilter: ", err)
		}
		fmt.Printf("✓ creada: %s\n", aws.ToString(out.Name))
		creadas = append(creadas, aws.ToString(out.Name))
	}

	// 3. Leerlas de vuelta, como las ve GuardDuty. (ListFilters también existe en la API, pero
	//    el emulador no lo implementa: se piden una por una con GetFilter.)
	fmt.Println()
	for _, nombre := range creadas {
		f, err := gd.GetFilter(ctx, &guardduty.GetFilterInput{DetectorId: det.DetectorId,
			FilterName: aws.String(nombre)})
		if err != nil {
			log.Fatal("GetFilter: ", err)
		}
		fmt.Printf("%s · acción=%s · rango=%d\n", nombre, f.Action, aws.ToInt32(f.Rank))
		claves := slices.Sorted(maps.Keys(f.FindingCriteria.Criterion))
		for _, k := range claves {
			fmt.Printf("    %s = %v\n", k, f.FindingCriteria.Criterion[k].Equals)
		}
	}

	// 4. Borrar una, y pedirla otra vez.
	if _, err := gd.DeleteFilter(ctx, &guardduty.DeleteFilterInput{DetectorId: det.DetectorId,
		FilterName: aws.String("bastion-ssh")}); err != nil {
		log.Fatal("DeleteFilter: ", err)
	}
	_, err = gd.GetFilter(ctx, &guardduty.GetFilterInput{DetectorId: det.DetectorId,
		FilterName: aws.String("bastion-ssh")})
	fmt.Printf("\ntras borrar bastion-ssh, GetFilter: %v\n", err)

	// 5. Lo que el emulador no hace. En AWS, ListFindings devolvería los IDs.
	_, err = gd.ListFindings(ctx, &guardduty.ListFindingsInput{DetectorId: det.DetectorId})
	var e interface{ HTTPStatusCode() int }
	if errors.As(err, &e) {
		fmt.Printf("\nListFindings: HTTP %d · el emulador no implementa los findings\n", e.HTTPStatusCode())
	} else {
		fmt.Printf("\nListFindings: %v\n", err)
	}
}
