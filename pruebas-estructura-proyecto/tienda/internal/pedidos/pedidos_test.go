package pedidos

import "testing"

type falso struct{ n int }

func (f *falso) Guardar(Pedido) (int, error) { f.n++; return f.n, nil }

func TestCrear(t *testing.T) {
	s := Servicio{A: &falso{}}
	if _, err := s.Crear("mouse", 0); err == nil {
		t.Fatal("esperaba error con cantidad 0")
	}
	p, err := s.Crear("mouse", 2)
	if err != nil || p.ID != 1 {
		t.Fatalf("p=%v err=%v", p, err)
	}
}
