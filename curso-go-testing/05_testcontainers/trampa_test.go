//go:build integracion && trampa

//	go test -tags 'integracion trampa' -v -run 'TestHora|TestTresContenedores' ./05_testcontainers
//	go test -tags 'integracion trampa' -v -run TestReuso ./05_testcontainers   (dos veces)
package contenedores

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"

	"curso-testing/tienda"
	"curso-testing/tienda/memoria"
)

// La hora que guardas no es la hora que lees: Postgres guarda microsegundos.
// Contra el fake, el mismo test pasa.
func TestHoraIdaYVuelta(t *testing.T) {
	ahora := time.Date(2026, 10, 28, 12, 0, 0, 123456789, time.UTC) // lo que da time.Now()
	for _, c := range []struct {
		nombre string
		r      tienda.Repositorio
	}{{"fake en memoria", memoria.Nuevo()}, {"Postgres", repo(t)}} {
		p := &tienda.Pedido{Cliente: "ana", Estado: "pagado", Creado: ahora,
			Items: []tienda.Item{{SKU: "TAZA", Cantidad: 1, Precio: 100}}}
		c.r.Guardar(t.Context(), p)
		got, _ := c.r.Buscar(t.Context(), p.ID)
		t.Logf("%-16s guardé %s  leí %s  iguales=%v", c.nombre,
			ahora.Format("15:04:05.000000000"), got.Creado.UTC().Format("15:04:05.000000000"),
			got.Creado.Equal(ahora))
		if !got.Creado.Equal(ahora) {
			t.Errorf("%s: la hora no vuelve igual", c.nombre)
		}
	}
}

// Lo que cuesta un contenedor por test: tres arranques seguidos.
func TestTresContenedores(t *testing.T) {
	for i := range 3 {
		t0 := time.Now()
		c, err := arrancar(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("contenedor %d: %.2f s", i+1, time.Since(t0).Seconds())
		testcontainers.TerminateContainer(c)
	}
}

// Reutilizar el contenedor ENTRE ejecuciones de go test, por nombre. La primera vez arranca;
// la segunda lo encuentra corriendo.
//
// A propósito NO se termina aquí (eso es reutilizarlo). Con Ryuk, se borra solo al acabar la
// sesión; sin Ryuk, se queda: bórralo tú, con su volumen:
//
//	docker rm -f -v curso-go-testing-pg-reuso
func TestReuso(t *testing.T) {
	t0 := time.Now()
	c, err := arrancar(context.Background(), testcontainers.WithReuseByName("curso-go-testing-pg-reuso"))
	if err != nil {
		t.Fatal(err)
	}
	id := c.GetContainerID()
	t.Logf("WithReuseByName: %.2f s · contenedor %s", time.Since(t0).Seconds(), id[:12])
}
