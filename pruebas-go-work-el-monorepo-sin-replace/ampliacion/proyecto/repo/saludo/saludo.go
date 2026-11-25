package saludo

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func Hola(n string) string { return "Hola, " + cases.Title(language.Spanish).String(n) }
