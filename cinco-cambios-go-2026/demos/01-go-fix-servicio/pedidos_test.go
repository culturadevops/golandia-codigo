package pedidos

import (
	"errors"
	"io/fs"
	"slices"
	"testing"
	"time"
)

func TestTodo(t *testing.T) {
	if got := Etiquetas("a, b ,c"); !slices.Equal(got, []string{"a", "b", "c"}) {
		t.Fatalf("Etiquetas = %q", got)
	}
	if c, ok := Cliente("cliente=ana"); c != "ana" || !ok {
		t.Fatalf("Cliente = %q %v", c, ok)
	}
	if !EsVIP([]string{"ana", "luis"}, "luis") || EsVIP(nil, "x") {
		t.Fatal("EsVIP")
	}
	dst := map[string]int{"a": 1}
	Copiar(dst, map[string]int{"b": 2})
	if len(dst) != 2 {
		t.Fatal("Copiar")
	}
	tot := []int{3, 1, 2}
	Ordenar(tot)
	if !slices.Equal(tot, []int{1, 2, 3}) {
		t.Fatal("Ordenar")
	}
	if r := Reintentos(3); r[2] != 400*time.Millisecond {
		t.Fatalf("Reintentos = %v", r)
	}
	res := Procesar([]Pedido{{Total: 1}, {Total: 2}}, func(p Pedido) int { return p.Total * 10 })
	if !slices.Equal(res, []int{10, 20}) || Procesados() != 2 {
		t.Fatalf("Procesar = %v %d", res, Procesados())
	}
	err := &fs.PathError{Op: "open", Path: "/datos/pedidos.csv", Err: fs.ErrNotExist}
	if r, ok := RutaFallida(errors.Join(errors.New("cargar"), err)); r != "/datos/pedidos.csv" || !ok {
		t.Fatal("RutaFallida")
	}
	if len(Campos(Pedido{ID: "x"})) != 2 {
		t.Fatal("Campos")
	}
}
