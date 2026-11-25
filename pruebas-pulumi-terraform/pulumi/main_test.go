package main

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// mocks: Pulumi no llama a AWS; cada recurso "se crea" aquí y lo apuntamos
type mocks struct {
	mu    sync.Mutex
	tipos map[string]int
	pol   string
}

func (m *mocks) NewResource(a pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tipos[a.TypeToken]++
	out := a.Inputs.Copy()
	out["arn"] = resource.NewStringProperty("arn:aws:mock:::" + a.Name)
	if a.TypeToken == "aws:iam/policy:Policy" {
		m.pol = a.Inputs["policy"].StringValue()
	}
	return a.Name + "_id", out, nil
}

func (m *mocks) Call(a pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return a.Args, nil
}

func correr(t *testing.T, entorno string) *mocks {
	m := &mocks{tipos: map[string]int{}}
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		return crear(ctx, Ajustes{Entorno: entorno, Colas: []string{"pedidos", "pagos", "avisos"},
			Clave: pulumi.ToSecret(pulumi.String("x")).(pulumi.StringOutput)})
	}, pulumi.WithMocks("g680", entorno, m))
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestProdLlevaVersionado(t *testing.T) {
	m := correr(t, "prod")
	if n := m.tipos["aws:s3/bucketVersioning:BucketVersioning"]; n != 1 {
		t.Fatalf("prod: %d versionados, quería 1", n)
	}
	if n := m.tipos["aws:sqs/queue:Queue"]; n != 3 {
		t.Fatalf("%d colas, quería 3", n)
	}
}

func TestDevSinVersionado(t *testing.T) {
	m := correr(t, "dev")
	if n := m.tipos["aws:s3/bucketVersioning:BucketVersioning"]; n != 0 {
		t.Fatalf("dev: %d versionados, quería 0", n)
	}
}

func TestLaPoliticaNombraCadaCola(t *testing.T) {
	m := correr(t, "prod")
	var doc struct{ Statement []struct{ Resource []string } }
	if err := json.Unmarshal([]byte(m.pol), &doc); err != nil {
		t.Fatal(err)
	}
	if n := len(doc.Statement[0].Resource); n != 3 {
		t.Fatalf("la política nombra %d colas, quería 3", n)
	}
}
