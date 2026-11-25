// Lección 8: la condición de carrera que las pruebas no ven sin -race.
package carrera

import (
	"bytes"
	"net/http/httptest"
	"sync"
	"testing"

	"curso-go-rendimiento/datos"
	"curso-go-rendimiento/resumen"
)

var registro = datos.Generar(5000)

// Lección 8. Cuarenta peticiones a la vez contra el mismo handler, como en producción.
// Sin -race, casi siempre pasa. Con -race, el detector ve la carrera aunque los bytes
// salgan bien esa vez.
func probarConcurrente(t *testing.T, h func(*httptest.ResponseRecorder, []byte)) {
	esperado, _ := resumen.ResumenV1(registro)
	var wg sync.WaitGroup
	var mu sync.Mutex
	malas := 0
	// Una goroutine por petición, como hace net/http con cada conexión.
	for range 40 {
		wg.Go(func() {
			rec := httptest.NewRecorder()
			h(rec, registro)
			if !bytes.Equal(rec.Body.Bytes(), esperado) {
				mu.Lock()
				malas++
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	if malas > 0 {
		t.Fatalf("%d de 40 respuestas corruptas", malas)
	}
}

func TestHandlerV4Concurrente(t *testing.T) {
	probarConcurrente(t, func(rec *httptest.ResponseRecorder, c []byte) {
		resumen.HandlerV4(rec, httptest.NewRequest("POST", "/resumen", bytes.NewReader(c)))
	})
}

func TestHandlerV5Concurrente(t *testing.T) {
	probarConcurrente(t, func(rec *httptest.ResponseRecorder, c []byte) {
		resumen.HandlerV5(rec, httptest.NewRequest("POST", "/resumen", bytes.NewReader(c)))
	})
}
