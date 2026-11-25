# banco-go-o-x

Código de estos vídeos del canal Golandia:

- [¿Go, Bun o Node? Lo medí en 2026 (y la trampa de los núcleos)](https://youtu.be/o7Ihsz9izcY)
- [¿Go o Java en 2026? Lo medí, y el que gana no es el que crees](https://youtu.be/8mCKytldeG4)
- [¿Go o Java? La revancha: GraalVM, jlink y una carga de verdad (medido)](https://youtu.be/BTjUtsPLrH8)
- [¿Go o Rust? Lo medí: velocidad, memoria y lo que cuesta escribirlo (axum vs net/http)](https://youtu.be/7spxAC8JTDw)
- [¿Go o C#? Lo medí con .NET 10 (también compilado nativo)](https://youtu.be/EFcnH4gahlE)
- [¿Go o Kotlin? Lo medí, y el JIT da la sorpresa](https://youtu.be/G1DfJ-VL-hI)
- [¿Go o Zig? Lo medí: más chico, más rápido… y más caro de escribir](https://youtu.be/_I6uG3utnm4)
- [¿Go o Elixir? Lo medí: lo que gana cada uno](https://youtu.be/ejEN_Vk4hnE)
- [¿Go o C? Lo medí: velocidad, memoria y los errores que C deja pasar (2026)](https://youtu.be/bQO4KryhBmo)
- [¿Go o C++? Lo medí: C++ calcula más rápido… y su servidor atendió menos](https://youtu.be/krzdSCnEd7E)
- [¿Go o PHP? Lo medí con PHP 8.5 y FrankenPHP (y el modo worker no ganó)](https://youtu.be/xaqTZ6v9Po8)
- [¿Go o Swift en el servidor? Lo medí, y esta vez Go perdió](https://youtu.be/zWaYcJ8XJwA)
- [¿Go o Crystal? Lo medí: el lenguaje con cara de Ruby contra Go (2026)](https://youtu.be/T35lhKysTgc)
- [¿Go o Dart? Lo medí: dart compile contra go build (2026)](https://youtu.be/aTzYHR-Nrts)
- [¿Gin + GORM o Serverpod? Lo medí: la misma API con Postgres en Go y en Dart](https://youtu.be/fsoKeGqW-I0)
- [¿Go o TypeScript con Deno? Lo medí: deno compile, permisos y rendimiento (2026)](https://youtu.be/y1O-jvNLStg)
- [¿Go o Gleam? Lo medí: tipos en la máquina de Erlang (2026)](https://youtu.be/jnS3qjzCGwI)
- [¿Go o Kotlin Native? Lo medí: arranca como Go, pero con objetos (2026)](https://youtu.be/buUkfJ9uzlQ)
- [¿Go o Mojo? Lo medí: rápido para calcular, pero sin servidor web (2026)](https://youtu.be/dbbpy_H6LLc)
- [¿Go o OCaml 5? Lo medí: dominios en varios núcleos contra gorrutinas (2026)](https://youtu.be/Ap4SJfWCKkU)
- [¿Go o Odin? Lo medí: el C moderno sin recolector contra Go (2026)](https://youtu.be/bqnLK1d-JzM)
- [¿Go o Python? Lo medí: velocidad, memoria y FastAPI (de 3 a 134 veces)](https://youtu.be/G7Ghf7UNhiQ)
- [¿Go o Ruby? Lo medí con Ruby 4, YJIT y ZJIT (2026)](https://youtu.be/5K0stB0oSQU)
- [¿Go o Scala? Lo medí en la JVM: velocidad, memoria y compilación (2026)](https://youtu.be/27-Odq2kaRc)
- [¿Go o Scala Native? Lo medí: arranca como Go, pero con objetos (2026)](https://youtu.be/jP3u1IDPAP4)
- [¿Go o TypeScript compilado a nativo? Bun AOT y scriptc, lo medí (2026)](https://youtu.be/94D-EpJk5jw)
- [¿Go o Zig? La REVANCHA: un comentario tenía razón (medido)](https://youtu.be/gi2NjDjEwUE)

## Qué es

El banco común de la serie «¿Go o X?»: los scripts que compilan y miden Go contra cada lenguaje (arranque, memoria, tamaño del binario, peticiones por segundo, cómputo) dentro de contenedores Docker. Cada script escribe sus resultados en la carpeta `pruebas-go-vs-<lenguaje>` correspondiente (en este repositorio, al lado). Los resultados medidos están en el `SALIDA*.md` de cada una.

## Cómo ejecutarlo

```bash
python pruebas_go_vs.py
python pruebas_go_vs_caliente.py
python pruebas_go_vs_deno.py
python pruebas_go_vs_lote_2026_11.py
python pruebas_go_vs_nativo_2026_11.py
python pruebas_go_vs_x.py
python pruebas_go_vs_x_2026_11_23.py
python pruebas_go_vs_x_computo.py
```

Casi todo corre dentro de Docker para que el resultado no dependa de tu máquina. Algunos scripts se escribieron para la estructura de carpetas original (rutas como `<repo>` o `<proyecto>`): ajústalas a la tuya.
