$ go version
go1.26.5
$ ls . publico
.:
publico
secreto.txt

publico:
docs
leeme.txt

$ ./archivos leeme.txt
pide: "leeme.txt"
  A) os.Open:      lee: hola, soy público
  B) root.Open:    lee: hola, soy público
  C) IsLocal:      true

$ ./archivos docs/manual.txt
pide: "docs/manual.txt"
  A) os.Open:      lee: manual en docs
  B) root.Open:    lee: manual en docs
  C) IsLocal:      true

$ ./archivos ../secreto.txt
pide: "../secreto.txt"
  A) os.Open:      lee: CLAVE_FALSA_DE_DEMO=1234
  B) root.Open:    error: openat ../secreto.txt: path escapes from parent
  C) IsLocal:      false

$ ./archivos docs/../../secreto.txt
pide: "docs/../../secreto.txt"
  A) os.Open:      lee: CLAVE_FALSA_DE_DEMO=1234
  B) root.Open:    error: openat docs/../../secreto.txt: path escapes from parent
  C) IsLocal:      false

