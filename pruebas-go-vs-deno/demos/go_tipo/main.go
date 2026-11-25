// El mismo error de tipo en Go: Get devuelve string, no int.
package main

import (
	"fmt"
	"net/url"
)

func main() {
	q, _ := url.ParseQuery("n=5")
	var n int = q.Get("n")
	fmt.Println("n + 1 =", n+1)
}
