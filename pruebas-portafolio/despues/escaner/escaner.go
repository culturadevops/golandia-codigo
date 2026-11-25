// Package escaner escanea puertos TCP con una concurrencia limitada.
//
// Es el escáner del vídeo 239 («Escribí un escáner de puertos en Go…»), con la lógica sacada de main para poder
// probarla, un contexto para cancelarlo y los errores clasificados. Úsalo solo contra máquinas tuyas o con permiso.
package escaner

import (
	"context"
	"errors"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Resultado es lo que devuelve un escaneo.
type Resultado struct {
	Abiertos []int          // puertos que aceptaron la conexión, ordenados
	Errores  map[string]int // por tipo: rechazado, timeout, sin-descriptores…
	Duracion time.Duration
}

// Escanear prueba los puertos de desde a hasta en host, con como mucho conc conexiones a la vez.
func Escanear(ctx context.Context, host string, desde, hasta, conc int, espera time.Duration) (Resultado, error) {
	if conc < 1 {
		return Resultado{}, errors.New("la concurrencia tiene que ser al menos 1")
	}
	if desde < 1 || hasta > 65535 || desde > hasta {
		return Resultado{}, errors.New("rango de puertos inválido")
	}
	var (
		mu  sync.Mutex
		wg  sync.WaitGroup
		res = Resultado{Errores: map[string]int{}}
		sem = make(chan struct{}, conc)
		d   = net.Dialer{Timeout: espera}
	)
	inicio := time.Now()
	for p := desde; p <= hasta; p++ {
		select {
		case <-ctx.Done():
			wg.Wait()
			return res, ctx.Err()
		case sem <- struct{}{}:
		}
		wg.Add(1)
		go func(puerto int) {
			defer wg.Done()
			defer func() { <-sem }()
			c, err := d.DialContext(ctx, "tcp", net.JoinHostPort(host, strconv.Itoa(puerto)))
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				res.Errores[Clasificar(err)]++
				return
			}
			_ = c.Close()
			res.Abiertos = append(res.Abiertos, puerto)
		}(p)
	}
	wg.Wait()
	sort.Ints(res.Abiertos)
	res.Duracion = time.Since(inicio)
	return res, nil
}

// Clasificar reduce un error de conexión a una etiqueta corta.
func Clasificar(err error) string {
	s := err.Error()
	switch {
	case strings.Contains(s, "refused"):
		return "rechazado"
	case strings.Contains(s, "timeout"), strings.Contains(s, "deadline"):
		return "timeout"
	case strings.Contains(s, "too many open files"):
		return "sin-descriptores"
	default:
		return "otro"
	}
}
