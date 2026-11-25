# Salida ejecutada — vídeo 226 «Go en Windows: qué se rompe al cruzar»

> Corrida del 2026-09-09. **El MISMO `main.go`**, compilado y ejecutado dos veces:
> nativo en Windows 11 (go1.26.5) y en `golang:alpine` dentro de un contenedor (go1.26.5).
> Salidas completas en `SALIDA-windows.txt` y `SALIDA-linux.txt`.

```
=== windows/amd64 · go go1.26.5 ===          === linux/amd64 · go go1.26.5 ===
```

---

## 🥇 LA TABLA, línea por línea

| | Windows | Linux |
|---|---|---|
| `filepath.Separator` | `\` | `/` |
| `filepath.Join("datos","informes","2026.csv")` | `datos\informes\2026.csv` | `datos/informes/2026.csv` |
| `path.Join(...)` (el de URLs) | `datos/informes/2026.csv` | `datos/informes/2026.csv` |
| **¿`filepath.Join` == `path.Join`?** | **false** | **true** |
| ¿abre un fichero con barra inclinada? | **sí** | sí |
| **¿abre `go.mod` escrito `GO.MOD`?** | **sí** | **no: no such file or directory** |
| crear una ruta de ~300 caracteres | sí, 330 car. | sí, 301 car. |
| `filepath.VolumeName(cwd)` | `C:` | *(vacío)* |
| `os.PathListSeparator` | `;` | `:` |
| directorios en PATH | 36 | 9 |
| **pedí `0600`, `Stat` dice** | **`-rw-rw-rw-`** | **`-rw-------`** |
| `os.Getuid()` | **-1** | 0 |
| ¿existe `/proc/self/status`? | **false** | true |
| `LookPath("cmd")` | `C:\Windows\system32\cmd.exe` | no está |
| nombre del binario | `gowin.exe` | `gowin` |
| `Signal(0)` a mí mismo | **`RECHAZADA: not supported by windows`** | aceptada |

---

## 🥇 1 · LA TRAMPA SILENCIOSA DE LAS RUTAS

En Linux, `filepath.Join` y `path.Join` dan **exactamente lo mismo**. En Windows no. Y como en
Linux dan lo mismo, el código escrito allí mezcla los dos paquetes sin que nadie se entere.

Y hay una segunda capa que lo hace peor: **en Windows, abrir `./go.mod` con barra inclinada
funciona**. Medido: `¿abre con barra inclinada?: sí`. Así que el programa no falla al abrir
ficheros. Falla cuando **compara** rutas, cuando las guarda en una base de datos, o cuando las
manda por la red: ahí una cadena tiene `\` y la otra `/`, y no coinciden.

> Un fallo que no aparece al abrir ficheros, solo al compararlos. Por eso escuece tanto.

## 🥇 1b · LAS MAYÚSCULAS, Y VA EN LA DIRECCIÓN QUE MUERDE

```
Windows:   ¿abre go.mod escrito GO.MOD?:  sí
Linux:     ¿abre go.mod escrito GO.MOD?:  no: open ./GO.MOD: no such file or directory
```

Fíjate en el sentido: **lo que perdona es Windows**. Escribes el código en Windows, abres
`Config.json` llamándolo `config.json`, funciona, pasan las pruebas, se hace la revisión. Y el día
que eso se despliega en un Linux, el fichero deja de existir.

Es el caso contrario a todo lo demás de este vídeo, y por eso es el que más veces llega a
producción: no falla en la máquina donde se escribió.

## ⚖️ 1c · Y UNA CREENCIA QUE LA MEDICIÓN DESMIENTE

El límite de 260 caracteres por ruta en Windows es de lo primero que sale cuando se habla de esto.
Aquí **no apareció**:

```
Windows:   crear una ruta de ~300 caracteres:  sí, 330 caracteres
Linux:     crear una ruta de ~300 caracteres:  sí, 301 caracteres
```

330 caracteres, creados y escritos sin un error. En Windows 11 con rutas largas habilitadas, y con
Go usando el prefijo de ruta extendida por debajo, ese tope ya no se toca en un caso normal. Se
dice tal cual: **medido aquí, no pasó**. En un Windows más viejo o con la opción desactivada puede
ser otra historia, y eso no lo he probado.

## 🥇 2 · LOS PERMISOS QUE PIDES NO SON LOS QUE TE DAN

```
Windows:   pedí 0600, Stat dice   -rw-rw-rw-
Linux:     pedí 0600, Stat dice   -rw-------
```

Le pides a Go que cree el fichero privado y en Windows sale con permiso de lectura y escritura
para todos. **El modo que pasas se ignora casi por completo**: Windows no usa permisos de estilo
Unix, usa listas de control de acceso, y la traducción es aproximada.

Si tu código escribe un fichero con credenciales y confía en el `0600`, en Windows esa confianza
no está fundada. Y `os.Getuid()` devolviendo **-1** es la otra cara: ahí no hay concepto de
usuario numérico que consultar.

## 🥇 3 · NO HAY `/proc`

```
Windows:   ¿existe /proc/self/status?: false
           y el error es: GetFileAttributesEx /proc/self/status:
                          El sistema no puede encontrar la ruta especificada.
```

Casi todo lo que se enseña de «Go de sistema» sale de leer `/proc`: memoria del proceso, hilos,
descriptores abiertos, línea de órdenes. Nada de eso existe al otro lado, y hay que ir a la API
del sistema. Fíjate además en el mensaje de error: ya no es un error de POSIX, es una función de
Windows la que se queja.

## 🥇 4 · LAS SEÑALES NO SE MANDAN

```
Windows:   Signal(0) a mí mismo: RECHAZADA: not supported by windows
Linux:     Signal(0) a mí mismo: aceptada
```

⚠️ Y lo que hace esto especialmente engañoso: **`syscall.SIGTERM` EXISTE en Windows dentro de Go**
y compila sin decir nada. Que la constante exista no significa que se pueda usar; `Process.Signal`
la rechaza en tiempo de ejecución. Recibir una interrupción sí funciona (`signal.Notify` con
`os.Interrupt`); mandarla, no.

Y lo que ni siquiera compila: `SIGUSR1` y `SIGHUP` no existen en Windows. El código que las usa
tiene que vivir en un fichero `_linux.go` y desaparecer al otro lado.

> ⚠️ Nota de método: se manda la señal **0**, que es la sonda que no hace nada. Con `SIGTERM` o con
> la interrupción, en Linux el programa se moriría ahí mismo y no habría salida que enseñar.

## 5 · Lo que NO cambió, y también hay que decirlo

- **Los saltos de línea**: escribí 8 bytes con `\n` y leí 8 bytes en los dos sitios, sin retorno de
  carro. Go no traduce nada: la conversión de finales de línea es cosa de Git o del editor, no del
  lenguaje. Es una creencia muy extendida y es falsa.
- `os.TempDir()` funciona en los dos, devolviendo lo que toca en cada sistema.
- `exec.LookPath` funciona en los dos y resuelve la extensión sola: por eso es la forma correcta,
  y escribir el nombre del ejecutable a pelo no lo es.

---

## Reproducir

```
# Windows nativo
go run .

# Linux, en contenedor
docker run --rm -v "$PWD:/src" -w /src golang:alpine \
  sh -c 'cp -r /src /tmp/w && cd /tmp/w && go run .'
```
