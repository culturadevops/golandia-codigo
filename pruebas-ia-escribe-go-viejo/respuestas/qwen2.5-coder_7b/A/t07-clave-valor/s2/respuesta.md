```go
package tarea

func ParClaveValor(s string) (clave, valor string, ok bool) {
    idx := strings.Index(s, "=")
    if idx == -1 {
        return "", "", false
    }
    return s[:idx], s[idx+1:], true
}
```