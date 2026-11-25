package velocidad

import (
	"encoding/json"
	"testing"
)

var datos = Datos()

func BenchmarkUnmarshal_v1API(b *testing.B) {
	b.SetBytes(int64(len(datos)))
	for b.Loop() {
		var v []Item
		if err := json.Unmarshal(datos, &v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_v1API(b *testing.B) {
	var v []Item
	json.Unmarshal(datos, &v)
	for b.Loop() {
		if _, err := json.Marshal(v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshalSinMapas_v1API(b *testing.B) {
	var v []Item
	json.Unmarshal(datos, &v)
	w := Quitar(v)
	for b.Loop() {
		if _, err := json.Marshal(w); err != nil {
			b.Fatal(err)
		}
	}
}
