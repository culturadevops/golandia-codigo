// encoding/json con struct tags: usa reflexión por dentro.
package main

import (
	"encoding/json"
	"fmt"
)

type Pedido struct {
	ID      int      `json:"id"`
	Cliente string   `json:"cliente"`
	Total   float64  `json:"total"`
	Items   []string `json:"items"`
}

func main() {
	b, err := json.Marshal(Pedido{ID: 42, Cliente: "Ana", Total: 99.5, Items: []string{"café", "pan"}})
	fmt.Println(string(b), err)
	var p Pedido
	err = json.Unmarshal([]byte(`{"id":7,"cliente":"Luis","total":10,"items":["té"]}`), &p)
	fmt.Printf("%+v %v\n", p, err)
}
