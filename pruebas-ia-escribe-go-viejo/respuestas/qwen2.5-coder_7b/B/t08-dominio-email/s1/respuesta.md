```go
package tarea

func Dominio(email string) string {
    idx := strings.Index(email, "@")
    if idx == -1 {
        return ""
    }
    return email[idx+1:]
}
```