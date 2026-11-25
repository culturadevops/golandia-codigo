package main

import (
	"fmt"
	"os/exec"
)

func main() {
	out, err := exec.Command("uname", "-s").Output()
	fmt.Printf("%q %v\n", out, err)
}
