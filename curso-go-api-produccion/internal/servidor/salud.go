package servidor

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"tareas/internal/web"
)

// BaseDeDatos es lo que la readiness necesita: contestar, y decir la versión del esquema.
type BaseDeDatos interface {
	Ping(ctx context.Context) error
	Version(ctx context.Context) (int64, error)
}

// Salud contesta las dos sondas. Son distintas a propósito:
//
//	/salud/vivo  ¿el proceso responde?  Nunca mira la base: si la base cae, reiniciar la
//	             API no la arregla, y con esta sonda mirándola se reiniciarían todas.
//	/salud/listo ¿le mando tráfico?     Base con plazo de 1 s, esquema al día, y no apagándose.
type Salud struct {
	BD       BaseDeDatos
	Esperada int64 // versión del esquema que necesita este binario
	Plazo    time.Duration
	apagando atomic.Bool
}

// Apagando hace que listo conteste 503: el balanceador deja de mandar peticiones nuevas.
func (s *Salud) Apagando() { s.apagando.Store(true) }

func (s *Salud) vivo(w http.ResponseWriter, _ *http.Request) {
	web.JSON(w, http.StatusOK, map[string]string{"estado": "vivo"})
}

func (s *Salud) listo(w http.ResponseWriter, r *http.Request) {
	estado, motivo := http.StatusOK, ""
	plazo := s.Plazo
	if plazo == 0 {
		plazo = time.Second
	}
	ctx, cancel := context.WithTimeout(r.Context(), plazo)
	defer cancel()
	v, err := s.revisar(ctx)
	if err != nil {
		estado, motivo = http.StatusServiceUnavailable, err.Error()
	}
	web.JSON(w, estado, map[string]any{"listo": estado == http.StatusOK, "esquema": v,
		"esperado": s.Esperada, "motivo": motivo})
}

func (s *Salud) revisar(ctx context.Context) (int64, error) {
	if s.apagando.Load() {
		return 0, fmt.Errorf("apagándose")
	}
	if err := s.BD.Ping(ctx); err != nil {
		return 0, fmt.Errorf("base de datos: %w", err)
	}
	v, err := s.BD.Version(ctx)
	if err != nil {
		return 0, fmt.Errorf("versión del esquema: %w", err)
	}
	if v < s.Esperada {
		return v, fmt.Errorf("esquema %d, el binario espera %d: falta migrar", v, s.Esperada)
	}
	return v, nil
}
