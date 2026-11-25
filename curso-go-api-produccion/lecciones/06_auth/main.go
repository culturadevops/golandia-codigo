// Lección 6: autenticación con sesiones en Postgres. Lo medido:
// el costo de bcrypt, el tiempo que delata qué correos existen, la tarea ajena y la sesión.
//
//	APP_BD_URL=postgres://app:app@localhost:5432/tareas go run ./lecciones/06_auth
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"tareas/internal/bd"
	"tareas/internal/cuentas"
	"tareas/internal/tareas"
)

func mediana(n int, f func()) time.Duration {
	ts := make([]time.Duration, n)
	for i := range ts {
		t0 := time.Now()
		f()
		ts[i] = time.Since(t0)
	}
	slices.Sort(ts)
	return ts[n/2]
}

func ms(d time.Duration) string { return fmt.Sprintf("%.1f ms", float64(d.Microseconds())/1000) }

func main() {
	ctx := context.Background()
	pool, err := bd.Abrir(ctx, os.Getenv("APP_BD_URL"))
	if err != nil {
		panic(err)
	}
	defer pool.Close()
	sufijo := time.Now().Format("150405")

	fmt.Println("1) lo que tarda bcrypt en cada costo (mediana de 3):")
	for costo := 10; costo <= 14; costo++ {
		d := mediana(3, func() { _, _ = bcrypt.GenerateFromPassword([]byte("una clave larga"), costo) })
		fmt.Printf("  costo %d -> %s\n", costo, ms(d))
	}

	c := cuentas.Nuevo(pool, 12)
	ana := "ana" + sufijo + "@ejemplo.com"
	idAna, err := c.Registrar(ctx, ana, "una clave larga")
	if err != nil {
		panic(err)
	}

	fmt.Println("\n2) iniciar sesión mal, 15 veces cada caso (mediana):")
	nadie := "nadie" + sufijo + "@ejemplo.com"
	fmt.Printf("  ingenuo   correo que NO existe: %s\n", ms(mediana(15, func() { _ = loginIngenuo(ctx, pool, nadie, "x") })))
	fmt.Printf("  ingenuo   correo que SÍ existe: %s\n", ms(mediana(15, func() { _ = loginIngenuo(ctx, pool, ana, "x") })))
	fmt.Printf("  plantilla correo que NO existe: %s\n", ms(mediana(15, func() { _, _, _ = c.IniciarSesion(ctx, nadie, "x") })))
	fmt.Printf("  plantilla correo que SÍ existe: %s\n", ms(mediana(15, func() { _, _, _ = c.IniciarSesion(ctx, ana, "x") })))

	fmt.Println("\n3) Beto pide la tarea de Ana por su id:")
	idBeto, _ := c.Registrar(ctx, "beto"+sufijo+"@ejemplo.com", "otra clave larga")
	t, _ := tareas.Postgres{BD: pool}.Crear(ctx, idAna, tareas.Nueva{Titulo: "pagar la renta", Prioridad: 1})
	var titulo string
	_ = pool.QueryRow(ctx, `SELECT titulo FROM tareas WHERE id = $1`, t.ID).Scan(&titulo)
	fmt.Printf("  ingenuo   WHERE id = $1                  -> %q\n", titulo)
	_, err = tareas.Postgres{BD: pool}.Obtener(ctx, idBeto, t.ID)
	fmt.Printf("  plantilla WHERE id = $1 AND usuario_id = $2 -> %v\n", err)

	fmt.Println("\n4) la sesión:")
	token, _, _ := c.IniciarSesion(ctx, ana, "una clave larga")
	var huella []byte
	_ = pool.QueryRow(ctx, `SELECT token_hash FROM sesiones WHERE usuario_id = $1`, idAna).Scan(&huella)
	fmt.Printf("  token al cliente: %s… (%d caracteres)\n", token[:12], len(token))
	fmt.Printf("  en la tabla:      %x… (SHA-256, %d bytes)\n", huella[:6], len(huella))
	d := mediana(200, func() { _, _ = c.Usuario(ctx, token) })
	fmt.Printf("  buscar la sesión en cada petición: %s (mediana de 200)\n", ms(d))
	_ = c.Cerrar(ctx, token)
	_, err = c.Usuario(ctx, token)
	fmt.Printf("  tras cerrar sesión, el mismo token -> %v\n", err)
}

// loginIngenuo devuelve en cuanto no encuentra el correo: no gasta el bcrypt.
func loginIngenuo(ctx context.Context, pool *pgxpool.Pool, email, clave string) error {
	var h []byte
	err := pool.QueryRow(ctx, `SELECT clave_hash FROM usuarios WHERE lower(email) = $1`, email).Scan(&h)
	if errors.Is(err, pgx.ErrNoRows) {
		return errors.New("correo o clave incorrectos")
	}
	return bcrypt.CompareHashAndPassword(h, []byte(clave))
}
