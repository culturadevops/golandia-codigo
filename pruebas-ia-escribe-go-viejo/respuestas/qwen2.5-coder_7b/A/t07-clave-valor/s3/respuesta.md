```go
package tarea

func ParClaveValor(s string) (clave, valor string, ok bool) {
    pos := s.indexOf("=")
    if pos == -1 {
        return "", "", false
    }
    clave = s[:pos]
    valor = s[pos+1:]
    return clave, valor, true
}
```