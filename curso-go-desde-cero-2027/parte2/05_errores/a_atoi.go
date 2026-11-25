package main

import (
	"fmt"
	"strconv"
)

func main() {
	for _, s := range []string{"42", "4x2", "99999999999999999999", ""} {
		n, err := strconv.Atoi(s)
		if err != nil {
			fmt.Printf("%-22q error: %v\n", s, err)
			continue
		}
		fmt.Printf("%-22q ok: %d\n", s, n)
	}
}
