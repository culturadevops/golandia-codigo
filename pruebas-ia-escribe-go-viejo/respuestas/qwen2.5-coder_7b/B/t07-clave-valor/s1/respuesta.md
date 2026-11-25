```go
package tarea

func ParClaveValor(s string) (clave, valor string, ok bool) {
    index := strings.Index(s, "=")
    if index == -1 {
        return "", "", false
    }
    clave = s[:index]
    valor = s[index+1:]
    return clave, valor, true
}
```