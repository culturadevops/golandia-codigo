# Curso de Go desde cero 2027 — Golandia (gratis, en YouTube)

Código de estos vídeos del canal Golandia:

- [Curso de Go desde cero 2027 · Parte 1 de 6: instalar, variables, tipos y fmt](https://youtu.be/-urmwDq9f6Y)
- [Curso de Go desde cero 2027 · Parte 2 de 6: if, for, switch, funciones, errores y defer](https://youtu.be/yO5LeZFJ_3o)
- [Curso de Go desde cero 2027 · Parte 3 de 6: arrays, slices y maps](https://youtu.be/45Xx0VfaOuY)
- [Curso de Go desde cero 2027 · Parte 4 de 6: structs, punteros e interfaces](https://youtu.be/TfURDZc0X5w)
- [Curso de Go desde cero 2027 · Parte 5 de 6: errores, paquetes, módulos y tests](https://youtu.be/Dbd1kITcoSE)
- [Curso de Go desde cero 2027 · Parte 6 de 6: concurrencia, goroutines y proyecto final](https://youtu.be/3TO-bo-JA6w)

---

Todo el código de las 6 partes del curso, ejecutable. Cada lección es una carpeta con sus `.go` y un
`demo.sh` que crea un módulo de Go de verdad, compila y ejecuta. Los errores de compilación son a propósito:
cada lección enseña lo que dice el compilador.

| parte | tema |
|---|---|
| `parte1/` | Instalar, `go mod init`, `go run` y `go build`, variables, tipos, valores cero, constantes |
| `parte2/` | if, for, switch, funciones, varios retornos, errores como valores y defer |
| `parte3/` | Arrays, slices, append, len y cap, maps |
| `parte4/` | Structs, métodos, punteros, interfaces y composición |
| `parte5/` | Errores con `%w`, `errors.Is` y `errors.As`, paquetes, módulos y `go test` |
| `parte6/` | Goroutines, channels, WaitGroup, `-race`, Mutex y el proyecto final |

    python ejecutar_todo.py          # todo en Docker (golang:1.27.2) y regenera SALIDA.md
    python ejecutar_todo.py parte3   # solo una parte, por pantalla

Sin Docker, con Go instalado (go.dev/dl): copia cualquier `.go` a una carpeta, `go mod init demo` y `go run .`.
