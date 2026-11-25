// Package tienda es el servicio pequeño sobre el que se prueba todo el curso: pedidos con
// lógica de negocio (descuento, envío), un repositorio (Postgres de verdad o un fake en
// memoria) y un cliente HTTP a una pasarela de cobros externa.
package tienda

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Item struct {
	SKU      string `json:"sku"`
	Cantidad int    `json:"cantidad"`
	Precio   Monto  `json:"precio_centavos"`
}

type Pedido struct {
	ID      int64     `json:"id"`
	Cliente string    `json:"cliente"`
	Items   []Item    `json:"items"`
	Total   Monto     `json:"total_centavos"`
	Estado  string    `json:"estado"`
	CobroID string    `json:"cobro_id"`
	Creado  time.Time `json:"creado"`
}

const (
	EnvioGratisDesde Monto = 50000 // 500,00
	CostoEnvio       Monto = 4900  // 49,00
)

var (
	ErrPedidoVacio = errors.New("el pedido no tiene items")
	ErrCantidad    = errors.New("cantidad inválida")
	ErrNoExiste    = errors.New("el pedido no existe")
)

func Validar(items []Item) error {
	if len(items) == 0 {
		return ErrPedidoVacio
	}
	for _, it := range items {
		if it.Cantidad <= 0 {
			return fmt.Errorf("%s: %w (%d)", it.SKU, ErrCantidad, it.Cantidad)
		}
	}
	return nil
}

func Subtotal(items []Item) Monto {
	var s Monto
	for _, it := range items {
		s += it.Precio * Monto(it.Cantidad)
	}
	return s
}

// Envio: gratis DESDE 500,00 (500,00 incluido).
func Envio(sub Monto) Monto {
	if sub >= EnvioGratisDesde {
		return 0
	}
	return CostoEnvio
}

// Descuento: 10 % en el primer pedido de cada cliente.
func Descuento(sub Monto, primero bool) Monto {
	if !primero {
		return 0
	}
	return sub / 10
}

func Total(items []Item, primero bool) Monto {
	sub := Subtotal(items)
	return sub - Descuento(sub, primero) + Envio(sub)
}

// Repositorio guarda pedidos. Guardar con ID 0 inserta y asigna el ID; con ID, actualiza.
type Repositorio interface {
	Guardar(ctx context.Context, p *Pedido) error
	Buscar(ctx context.Context, id int64) (*Pedido, error)
	PorCliente(ctx context.Context, cliente string) ([]Pedido, error)
}

// Cobrador es la pasarela de pagos externa (en producción, cobros.Cliente por HTTP).
type Cobrador interface {
	Cobrar(ctx context.Context, cliente string, m Monto) (string, error)
}

type Servicio struct {
	Repo   Repositorio
	Cobros Cobrador
	Ahora  func() time.Time
}

func (s *Servicio) ahora() time.Time {
	if s.Ahora != nil {
		return s.Ahora()
	}
	// Postgres guarda microsegundos: se trunca aquí para que lo guardado y lo leído coincidan
	// (lección 5).
	return time.Now().UTC().Truncate(time.Microsecond)
}

// Crear valida, calcula el total, cobra y guarda el pedido como pagado.
func (s *Servicio) Crear(ctx context.Context, cliente string, items []Item) (*Pedido, error) {
	if err := Validar(items); err != nil {
		return nil, err
	}
	previos, err := s.Repo.PorCliente(ctx, cliente)
	if err != nil {
		return nil, fmt.Errorf("leyendo pedidos de %s: %w", cliente, err)
	}
	total := Total(items, len(previos) == 0)
	id, err := s.Cobros.Cobrar(ctx, cliente, total)
	if err != nil {
		return nil, fmt.Errorf("cobrando %s a %s: %w", total, cliente, err)
	}
	p := &Pedido{Cliente: cliente, Items: items, Total: total, Estado: "pagado",
		CobroID: id, Creado: s.ahora()}
	if err := s.Repo.Guardar(ctx, p); err != nil {
		return nil, fmt.Errorf("guardando pedido: %w", err)
	}
	return p, nil
}

// Recibo es el texto que se manda al cliente. Lo prueban los golden files (lección 6).
func Recibo(p *Pedido) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Pedido #%d · %s\n", p.ID, p.Cliente)
	fmt.Fprintf(&b, "Fecha: %s\n", p.Creado.UTC().Format("2006-01-02 15:04 MST"))
	b.WriteString(strings.Repeat("-", 40) + "\n")
	for _, it := range p.Items {
		linea := it.Precio * Monto(it.Cantidad)
		fmt.Fprintf(&b, "%-10s %3d x %10s %12s\n", it.SKU, it.Cantidad, it.Precio, linea)
	}
	b.WriteString(strings.Repeat("-", 40) + "\n")
	sub := Subtotal(p.Items)
	fmt.Fprintf(&b, "%-26s %13s\n", "Subtotal", sub)
	if d := sub + Envio(sub) - p.Total; d != 0 {
		fmt.Fprintf(&b, "%-26s %13s\n", "Descuento primer pedido", -d)
	}
	if e := Envio(sub); e == 0 {
		fmt.Fprintf(&b, "%-26s %13s\n", "Envío", "gratis")
	} else {
		fmt.Fprintf(&b, "%-26s %13s\n", "Envío", e)
	}
	fmt.Fprintf(&b, "%-26s %13s\n", "TOTAL", p.Total)
	cobro := p.CobroID
	if cobro == "" {
		cobro = "ninguno"
	}
	fmt.Fprintf(&b, "Estado: %s · cobro: %s\n", p.Estado, cobro)
	return b.String()
}
