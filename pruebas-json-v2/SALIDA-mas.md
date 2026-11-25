# SALIDA — más diferencias entre encoding/json y encoding/json/v2 (golandia)

> Corrida del 2026-10-20 17:43. golang:1.27.2.

```
$ go run ./mas
v1 dur {"d":1500000000,"n":0,"m":null} <nil>
v2 dur {"d" json: cannot marshal from Go time.Duration within "/d": no default representation
v1 null/merge {D:0s N:5 M:map[x:1 y:2]} <nil>
v2 null/merge {D:0s N:0 M:map[x:1 y:2]} <nil>
v1 desconocido <nil>
v2 RejectUnknown json: cannot unmarshal JSON string into Go main.T: unknown object member name "extra"
v1 string->int {0} json: cannot unmarshal string into Go struct field .n of type int
v2 string->int {0} json: cannot unmarshal JSON string into Go int within "/n"
```
