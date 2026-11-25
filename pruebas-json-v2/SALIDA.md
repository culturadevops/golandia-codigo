# SALIDA — encoding/json/v2 en Go 1.27 (golandia)

> Corrida del 2026-10-20 16:46. golang:1.27.2, contenedor con --cpus 2. Banco: 10.000 objetos (~1,6 MB), mediana de 5 pasadas. Notas de Go 1.27: fuente-go1.27.html.

```
$ go version
go version go1.27.2 linux/amd64

$ go run .
1 · Slice y map sin inicializar (nil):
  v1  {"nombre":"ana","roles":null,"extra":null}
  v2  {"nombre":"ana","roles":[],"extra":{},"edad":0}

2 · Mayúsculas distintas en la clave ({"NOMBRE":"ana"}):
  v1  nombre="ana"
  v2  nombre=""

3 · Clave repetida ({"nombre":"ana","nombre":"admin"}):
  v1  nombre="admin"
  v2  error: jsontext: duplicate object member name "nombre"

4 · UTF-8 inválido en el texto (byte 0xff):
  v1  nombre="an�a"
  v2  error: jsontext: invalid UTF-8 within "/nombre" after offset 13

5 · Opciones en la llamada (v2):
  Deterministic + WithIndent:
{
  "a": 2,
  "m": 3,
  "z": 1
}
  v2  {"nombre":"ana","roles":null,"extra":null,"edad":0}
      ^ con FormatNilSliceAsNull y FormatNilMapAsNull: igual que v1

6 · Leer directo de un io.Reader:
  UnmarshalRead -> [admin] err=<nil>

$ GOEXPERIMENT=nojsonv2 go test -bench . -benchmem -count 5   # encoding/json con su implementación vieja
  BenchmarkMarshalSinMapas_v1API       1.29 ms/op   1.04 MB/op       2 allocs/op
  BenchmarkMarshal_v1API               2.92 ms/op   2.93 MB/op   50002 allocs/op
  BenchmarkUnmarshal_v1API            13.55 ms/op   8.83 MB/op  150026 allocs/op

$ go test -bench . -benchmem -count 5   # Go 1.27 por defecto: encoding/json ya va sobre v2
  BenchmarkMarshalSinMapas_v1API       2.08 ms/op   1.06 MB/op       3 allocs/op
  BenchmarkMarshalSinMapas_v2API       2.19 ms/op   1.04 MB/op       3 allocs/op
  BenchmarkMarshal_v1API               4.69 ms/op   2.13 MB/op   40004 allocs/op
  BenchmarkMarshal_v2API               4.22 ms/op   1.84 MB/op   20003 allocs/op
  BenchmarkUnmarshal_v1API            10.03 ms/op   8.44 MB/op   83191 allocs/op
  BenchmarkUnmarshal_v2API             8.58 ms/op   8.44 MB/op   83191 allocs/op
```
