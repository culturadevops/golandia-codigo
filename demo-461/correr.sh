#!/usr/bin/env bash
# Demo del vídeo 461 «Fuzzing en Go» — se EJECUTA y su salida va a SALIDA.md (regla: la terminal del vídeo se ejecuta).
#   1) una función que interpreta rangos «10-20», con tests de tabla en verde;
#   2) go test -fuzz encuentra en segundos una entrada que la hace entrar en pánico;
#   3) el caso queda guardado en testdata/fuzz y pasa a ser un test más;
#   4) el arreglo, y el fuzzer otra vez sin encontrar nada.
# Trabaja dentro de un directorio temporal: no toca nada fuera.
set -u
T=$(mktemp -d)
cd "$T"
go mod init demo/rangos >/dev/null 2>&1

cat > rango.go <<'EOF'
package rangos

import (
	"strconv"
	"strings"
)

// Rango interpreta "10-20" y devuelve 10 y 20.
func Rango(s string) (int, int, error) {
	i := strings.Index(s, "-")
	desde, err := strconv.Atoi(s[:i])
	if err != nil {
		return 0, 0, err
	}
	hasta, err := strconv.Atoi(s[i+1:])
	if err != nil {
		return 0, 0, err
	}
	return desde, hasta, nil
}
EOF

cat > rango_test.go <<'EOF'
package rangos

import "testing"

func TestRango(t *testing.T) {
	casos := []struct {
		in          string
		desde, hasta int
	}{
		{"10-20", 10, 20},
		{"1-2", 1, 2},
		{"0-100", 0, 100},
	}
	for _, c := range casos {
		d, h, err := Rango(c.in)
		if err != nil || d != c.desde || h != c.hasta {
			t.Errorf("Rango(%q) = %d, %d, %v", c.in, d, h, err)
		}
	}
}

func FuzzRango(f *testing.F) {
	f.Add("10-20")
	f.Fuzz(func(t *testing.T, s string) {
		Rango(s) // no debe entrar en pánico con NINGUNA entrada
	})
}
EOF

echo '$ go test -run TestRango -v .'
go test -run TestRango -v . 2>&1 | grep -E "^(=== RUN|--- |PASS|FAIL|ok)" | sed 's/[0-9.]*s)$/…s)/; s/\t[0-9.]*s$//'
echo
echo '$ go test -fuzz FuzzRango -fuzztime 60s .'
go test -fuzz FuzzRango -fuzztime 60s . > fuzz1.txt 2>&1
echo "(código de salida: $?)"
grep -E "^(--- FAIL|    --- FAIL|        panic|panic:|    Failing input|    To re-run|        go test|FAIL|ok)" fuzz1.txt | head -8
grep -m1 -E "runtime error" fuzz1.txt
echo
echo '$ ls testdata/fuzz/FuzzRango/'
ls testdata/fuzz/FuzzRango/ | sed 's/^\(.\{8\}\).*/\1…/'
echo '$ cat testdata/fuzz/FuzzRango/*'
cat testdata/fuzz/FuzzRango/*
echo
echo '$ go test -run FuzzRango .   # sin -fuzz: el caso guardado ya es un test'
go test -run FuzzRango . 2>&1 | grep -E "^(--- FAIL|FAIL|ok)" | sed 's/\t[0-9.]*s$//'
echo

# el arreglo: si no hay guion, error en vez de pánico
cat > rango.go <<'EOF'
package rangos

import (
	"errors"
	"strconv"
	"strings"
)

// Rango interpreta "10-20" y devuelve 10 y 20.
func Rango(s string) (int, int, error) {
	a, b, ok := strings.Cut(s, "-")
	if !ok {
		return 0, 0, errors.New("falta el guion")
	}
	desde, err := strconv.Atoi(a)
	if err != nil {
		return 0, 0, err
	}
	hasta, err := strconv.Atoi(b)
	if err != nil {
		return 0, 0, err
	}
	return desde, hasta, nil
}
EOF
echo '$ # arreglo: strings.Cut y error si falta el guion'
echo '$ go test -fuzz FuzzRango -fuzztime 60s .'
go test -fuzz FuzzRango -fuzztime 60s . > fuzz2.txt 2>&1
echo "(código de salida: $?)"
grep -E "^fuzz: elapsed" fuzz2.txt | tail -1 | sed 's/fuzz: elapsed: \([0-9a-z]*\), execs: \([0-9]*\) .*/fuzz: elapsed: \1, execs: \2/'
grep -E "^(PASS|ok|FAIL)" fuzz2.txt | sed 's/\t[0-9.]*s$//'
cd / && rm -rf "$T"
