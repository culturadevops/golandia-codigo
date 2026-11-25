// Package pedido: el flujo de un pedido (reservar, cobrar, facturar, enviar) como workflow de Temporal.
//
// Los «efectos» (lo que en la vida real sería otra API: el stock, la pasarela de pago, el PDF de la factura)
// se apuntan en /datos/efectos.log, un volumen compartido por todos los workers. Así se cuenta cuántas veces
// se ejecutó de verdad cada paso, aunque el proceso que lo hizo esté muerto.
package pedido

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const Cola = "pedidos"

const Paginas = 6 // la factura tiene 6 páginas; cada una tarda 1 s

// Version la fija cada worker con la variable VERSION:
//
//	"1": el flujo original
//	"2": alguien agrega un paso (Validar) antes de cobrar, sin más
//	"3": el mismo paso, pero detrás de workflow.GetVersion
var Version = "1"

// Efecto apunta un efecto «externo» en el volumen compartido.
func Efecto(quien, que string) {
	f, err := os.OpenFile("/datos/efectos.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	fmt.Fprintf(f, "%s %s %s\n", time.Now().UTC().Format("15:04:05.000"), quien, que)
}

func quien() string { return os.Getenv("WORKER") }

func log(ctx context.Context, f string, a ...any) {
	i := activity.GetInfo(ctx)
	fmt.Printf("%s [%s] %s (intento %d) %s\n", time.Now().UTC().Format("15:04:05.000"), quien(), i.ActivityType.Name,
		i.Attempt, fmt.Sprintf(f, a...))
}

func Reservar(ctx context.Context, id string) error {
	log(ctx, "reservo el stock de %s", id)
	time.Sleep(time.Second)
	Efecto(quien(), "reservar "+id)
	return nil
}

func Cobrar(ctx context.Context, id string) error {
	i := activity.GetInfo(ctx)
	if i.Attempt < 3 { // la pasarela falla las dos primeras veces
		log(ctx, "la pasarela no responde")
		return fmt.Errorf("pasarela de pago: tiempo agotado")
	}
	log(ctx, "cobro %s", id)
	Efecto(quien(), "cobrar "+id)
	return nil
}

func Validar(ctx context.Context, id string) error {
	log(ctx, "valido la dirección de %s", id)
	Efecto(quien(), "validar "+id)
	return nil
}

// Facturar genera 6 páginas y avisa (heartbeat) con la última terminada. Si el worker muere, el siguiente
// intento lee ese número y sigue desde la página siguiente.
func Facturar(ctx context.Context, id string) error {
	desde := 1
	if activity.HasHeartbeatDetails(ctx) {
		var hecha int
		if err := activity.GetHeartbeatDetails(ctx, &hecha); err == nil {
			desde = hecha + 1
		}
	}
	log(ctx, "empiezo en la página %d de %d", desde, Paginas)
	for p := desde; p <= Paginas; p++ {
		time.Sleep(time.Second)
		Efecto(quien(), fmt.Sprintf("factura %s página %d", id, p))
		log(ctx, "página %d de %d", p, Paginas)
		activity.RecordHeartbeat(ctx, p)
	}
	return nil
}

func Enviar(ctx context.Context, id string) error {
	log(ctx, "envío %s", id)
	Efecto(quien(), "enviar "+id)
	return nil
}

// Pedido es el workflow. Su código se REPRODUCE desde el historial cada vez que un worker lo retoma:
// por eso tiene que ser determinista (nada de time.Now, rand ni goroutines propias aquí dentro).
func Pedido(ctx workflow.Context, id string) (string, error) {
	estado := "empezando"
	if err := workflow.SetQueryHandler(ctx, "estado", func() (string, error) { return estado, nil }); err != nil {
		return "", err
	}
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		HeartbeatTimeout:    3 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2,
			MaximumInterval:    10 * time.Second,
		},
	})
	estado = "reservando"
	if err := workflow.ExecuteActivity(ctx, Reservar, id).Get(ctx, nil); err != nil {
		return "", err
	}
	switch Version {
	case "2":
		estado = "validando"
		if err := workflow.ExecuteActivity(ctx, Validar, id).Get(ctx, nil); err != nil {
			return "", err
		}
	case "3":
		if workflow.GetVersion(ctx, "validar-direccion", workflow.DefaultVersion, 1) == 1 {
			estado = "validando"
			if err := workflow.ExecuteActivity(ctx, Validar, id).Get(ctx, nil); err != nil {
				return "", err
			}
		}
	}
	estado = "cobrando"
	if err := workflow.ExecuteActivity(ctx, Cobrar, id).Get(ctx, nil); err != nil {
		return "", err
	}
	estado = "facturando"
	if err := workflow.ExecuteActivity(ctx, Facturar, id).Get(ctx, nil); err != nil {
		return "", err
	}
	estado = "enviando"
	if err := workflow.ExecuteActivity(ctx, Enviar, id).Get(ctx, nil); err != nil {
		return "", err
	}
	estado = "terminado"
	return "pedido " + id + " enviado", nil
}
