// Go 1.27: el GODEBUG asynctimerchan se eliminó; los canales de time son siempre sin búfer.
package main

import (
	"fmt"
	"time"
)

func main() {
	t := time.NewTimer(10 * time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	fmt.Println("cap(t.C) =", cap(t.C))
}
