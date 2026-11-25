package main

import (
	"fmt"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func main() { fmt.Println(cases.Upper(language.Spanish).String("listo")) }
