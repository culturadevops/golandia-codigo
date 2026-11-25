// esterminal dice, por stderr, si su stdout es una terminal según dos formas de mirarlo.
package main

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

func main() {
	st, _ := os.Stdout.Stat()
	fmt.Fprintf(os.Stderr, "  os.ModeCharDevice: %-5t   term.IsTerminal: %t\n",
		st.Mode()&os.ModeCharDevice != 0, term.IsTerminal(int(os.Stdout.Fd())))
}
