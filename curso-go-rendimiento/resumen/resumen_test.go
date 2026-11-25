package resumen

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"curso-go-rendimiento/datos"
)

// 5.000 líneas: unos 1,1 MB de registro por petición.
var registro = datos.Generar(5000)

func TestTodasIguales(t *testing.T) {
	v1, err := ResumenV1(registro)
	if err != nil {
		t.Fatal(err)
	}
	for nombre, f := range map[string]func([]byte) ([]byte, error){
		"V2": ResumenV2, "V3": ResumenV3, "V4": ResumenV4,
	} {
		got, err := f(registro)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, v1) {
			t.Errorf("%s no devuelve lo mismo que V1", nombre)
		}
	}
	var b bytes.Buffer
	if err := EscribirV5(&b, registro); err != nil || !bytes.Equal(b.Bytes(), v1) {
		t.Errorf("V5 no devuelve lo mismo que V1 (err=%v)", err)
	}
}

// La prueba de siempre: una petición, se compara la respuesta. V4 la pasa.
func TestHandlerV4(t *testing.T) {
	esperado, _ := ResumenV1(registro)
	for range 3 {
		rec := httptest.NewRecorder()
		HandlerV4(rec, httptest.NewRequest("POST", "/resumen", bytes.NewReader(registro)))
		if rec.Code != http.StatusOK || !bytes.Equal(rec.Body.Bytes(), esperado) {
			t.Fatalf("respuesta distinta (código %d)", rec.Code)
		}
	}
}

func BenchmarkResumen(b *testing.B) {
	for _, v := range []struct {
		nombre string
		f      func([]byte) ([]byte, error)
	}{{"V1", ResumenV1}, {"V2", ResumenV2}, {"V3", ResumenV3}} {
		b.Run("version="+v.nombre, func(b *testing.B) {
			b.SetBytes(int64(len(registro)))
			for b.Loop() {
				if _, err := v.f(registro); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// El servicio entero, con la lectura del cuerpo y la escritura de la respuesta.
func BenchmarkHandler(b *testing.B) {
	for _, v := range []struct {
		nombre string
		h      http.HandlerFunc
	}{{"V3", HandlerV3}, {"V5", HandlerV5}} {
		b.Run("version="+v.nombre, func(b *testing.B) {
			b.SetBytes(int64(len(registro)))
			for b.Loop() {
				rec := httptest.NewRecorder()
				v.h(rec, httptest.NewRequest("POST", "/resumen", bytes.NewReader(registro)))
				if rec.Code != http.StatusOK {
					b.Fatal(rec.Code)
				}
			}
		})
	}
}
