# SALIDA — demos de pruebas-go-vs-ruby (golandia)

> Corrida del 2026-11-05 08:17. Regenerable con `python investigacion/demos_go_vs_lote_2026_11.py ruby`. Cada bloque es la salida literal del guion de `demos/`.

## Ruby: el error de tipos que espera a que pase por ahí

Imagen `ruby:4.0.9-slim`, `--cpus 1`, guion `demos/tipos.sh`.

```
### ruby -wc precio.rb   (revisa la sintaxis)
Syntax OK
[código de salida: 0]
### ruby precio.rb
martes: 100
miércoles: 100
precio.rb:5:in 'Integer#-': String can't be coerced into Integer (TypeError)
	from precio.rb:5:in 'Object#precio_final'
	from precio.rb:13:in '<main>'
[código de salida: 1]
```

## Go: el mismo error

Imagen `golang:1.27.2`, `--cpus 1`, guion `demos/tipos_go.sh`.

```
### go build .
# demo
./main.go:8:10: invalid operation: precio - "10" (mismatched types int and untyped string)
[código de salida: 1]
```

## Líneas de cada servidor

Imagen `golang:1.27.2`, `--cpus 1`, guion `demos/lineas.sh`.

```
90 líneas  go/main.go
39 líneas  app/config.ru
```

