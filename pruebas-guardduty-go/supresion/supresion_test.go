package supresion

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/guardduty/types"
)

func TestValidar(t *testing.T) {
	casos := []struct {
		nombre string
		r      Regla
		quiere error
	}{
		{"estrecha", Regla{Tipo: "Recon:EC2/Portscan",
			Recurso: map[string]string{"resource.instanceDetails.imageId": "ami-999999999"}}, nil},
		{"sin tipo", Regla{Recurso: map[string]string{"resource.instanceDetails.imageId": "ami-1"}}, ErrSinTipo},
		{"solo tipo", Regla{Tipo: "Recon:EC2/Portscan"}, ErrSinRecurso},
		{"comodín", Regla{Tipo: "*:Kubernetes/*",
			Recurso: map[string]string{"resource.resourceType": "EKSCluster"}}, ErrComodin},
	}
	for _, c := range casos {
		if err := c.r.Validar(); !errors.Is(err, c.quiere) {
			t.Errorf("%s: got %v, quiere %v", c.nombre, err, c.quiere)
		}
	}
}

func TestEntradaArchiva(t *testing.T) {
	r := Regla{Nombre: "bastion", Tipo: "UnauthorizedAccess:EC2/SSHBruteForce",
		Recurso: map[string]string{"resource.instanceDetails.tags.value": "devops"}}
	in, err := r.Entrada("det", 1)
	if err != nil {
		t.Fatal(err)
	}
	if in.Action != types.FilterActionArchive {
		t.Errorf("acción %q, quiere ARCHIVE", in.Action)
	}
	if n := len(in.FindingCriteria.Criterion); n != 2 {
		t.Errorf("%d criterios, quiere 2 (tipo + recurso)", n)
	}
}
