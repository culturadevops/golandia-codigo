$ go test -run TestRango -v .
=== RUN   TestRango
--- PASS: TestRango (…s)
PASS
ok  	demo/rangos

$ go test -fuzz FuzzRango -fuzztime 60s .
(código de salida: 1)
--- FAIL: FuzzRango (0.18s)
    --- FAIL: FuzzRango (0.00s)
    Failing input written to testdata\fuzz\FuzzRango\771e938e4458e983
    To re-run:
FAIL
FAIL	demo/rangos	0.595s
        testing.go:1927: panic: runtime error: slice bounds out of range [:-1]

$ ls testdata/fuzz/FuzzRango/
771e938e…
$ cat testdata/fuzz/FuzzRango/*
go test fuzz v1
string("0")

$ go test -run FuzzRango .   # sin -fuzz: el caso guardado ya es un test
--- FAIL: FuzzRango (0.00s)
FAIL	demo/rangos
FAIL

$ # arreglo: strings.Cut y error si falta el guion
$ go test -fuzz FuzzRango -fuzztime 60s .
(código de salida: 0)
fuzz: elapsed: 1m0s, execs: 180573613
PASS
ok  	demo/rangos
