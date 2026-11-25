// Más diferencias v1/v2 ejecutadas: time.Duration, null sobre un valor ya lleno, claves desconocidas, texto en un campo int.
package main

import (
	jsonv1 "encoding/json"
	"encoding/json/v2"
	"fmt"
	"time"
)

type T struct {
	D time.Duration `json:"d"`
	N int           `json:"n"`
	M map[string]int `json:"m"`
}

func main() {
	b, err := jsonv1.Marshal(T{D: 1500 * time.Millisecond})
	fmt.Println("v1 dur", string(b), err)
	b, err = json.Marshal(T{D: 1500 * time.Millisecond})
	fmt.Println("v2 dur", string(b), err)
	var a, c T
	a = T{N: 5, M: map[string]int{"x": 1}}
	err = jsonv1.Unmarshal([]byte(`{"n":null,"m":{"y":2}}`), &a)
	fmt.Printf("v1 null/merge %+v %v\n", a, err)
	c = T{N: 5, M: map[string]int{"x": 1}}
	err = json.Unmarshal([]byte(`{"n":null,"m":{"y":2}}`), &c)
	fmt.Printf("v2 null/merge %+v %v\n", c, err)
	var u T
	err = jsonv1.Unmarshal([]byte(`{"n":1,"extra":true}`), &u)
	fmt.Println("v1 desconocido", err)
	err = json.Unmarshal([]byte(`{"n":1,"extra":true}`), &u, json.RejectUnknownMembers(true))
	fmt.Println("v2 RejectUnknown", err)
	var x struct{ N int `json:"n"` }
	err = jsonv1.Unmarshal([]byte(`{"n":"7"}`), &x)
	fmt.Println("v1 string->int", x, err)
	err = json.Unmarshal([]byte(`{"n":"7"}`), &x)
	fmt.Println("v2 string->int", x, err)
}
