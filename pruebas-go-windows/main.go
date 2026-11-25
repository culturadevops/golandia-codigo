// main.go — el MISMO programa, ejecutado en Windows nativo y en un contenedor Linux.
//
// Vídeo 226 · «Go en Windows: qué se rompe al cruzar».
//
// ⛔ LA REGLA: no se cuenta de memoria. Este fichero se compila y se ejecuta en los dos sitios y
// se comparan las dos salidas, línea por línea. Lo que no salga en las dos corridas, no entra
// en el vídeo.
//
// Lo que se mide, y cada punto es una cosa que le rompe el código a alguien al cruzar:
//  1. separadores de ruta y qué pasa al mezclar `path` con `path/filepath`
//  2. la variable del sistema con los directorios, y su separador (que TAMBIÉN cambia)
//  3. saltos de línea al escribir y leer un fichero de texto
//  4. permisos: qué devuelve `Stat` en cada sitio
//  5. si /proc existe, que es de donde sale casi todo lo que se enseña de Go de sistema
//  6. señales: cuáles se pueden siquiera nombrar
//  7. cómo se llama un ejecutable, y qué pasa con el nombre del binario
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

func linea(k string, v any) { fmt.Printf("  %-34s %v\n", k+":", v) }

func main() {
	fmt.Printf("=== %s/%s · go %s ===\n", runtime.GOOS, runtime.GOARCH, runtime.Version())

	fmt.Println("\n-- 1 · RUTAS --")
	linea("filepath.Separator", string(filepath.Separator))
	linea("filepath.Join(a,b,c)", filepath.Join("datos", "informes", "2026.csv"))
	// ⚠️ `path` es SIEMPRE con barra inclinada: es para URLs y rutas de red, NO para el disco.
	// Mezclarlos es el fallo silencioso más común al cruzar, porque en Linux los dos dan lo mismo.
	linea("path.Join(a,b,c)  (¡URLs!)", path.Join("datos", "informes", "2026.csv"))
	linea("¿dan lo mismo?", filepath.Join("a", "b") == path.Join("a", "b"))
	linea("filepath.Abs(\"x\")", must(filepath.Abs("x")))
	linea("VolumeName del actual", filepath.VolumeName(must(os.Getwd())))
	// Y esta es la trampa fina: en Windows, la barra inclinada TAMBIÉN funciona al abrir
	// ficheros. Por eso el código mal escrito parece que va bien... hasta que compara cadenas.
	linea("¿abre con barra inclinada?", abre("./go.mod"))

	// ⚠️ MAYÚSCULAS: en Windows el sistema de ficheros no distingue may. de min., y en Linux sí.
	// Es de los que más rompen: el código escrito en Windows abre "Config.json" llamándolo
	// "config.json" y funciona; al desplegar en Linux, deja de existir.
	linea("¿abre go.mod escrito GO.MOD?", abre("./GO.MOD"))

	// ⚠️ LÍMITE DE LONGITUD: Windows arrastra un tope histórico de 260 caracteres por ruta.
	// Se mide creando una ruta larga de verdad, no citándolo de memoria.
	linea("crear una ruta de ~300 caracteres", rutaLarga())

	fmt.Println("\n-- 2 · LA VARIABLE DE DIRECTORIOS --")
	linea("os.PathListSeparator", string(os.PathListSeparator))
	rutas := filepath.SplitList(os.Getenv("PATH"))
	linea("cuántos directorios trae", len(rutas))
	if len(rutas) > 0 {
		linea("el primero", rutas[0])
	}

	fmt.Println("\n-- 3 · SALTOS DE LÍNEA --")
	f := filepath.Join(os.TempDir(), "prueba_saltos.txt")
	_ = os.WriteFile(f, []byte("uno\ndos\n"), 0o644)
	b, _ := os.ReadFile(f)
	linea("escribí 8 bytes, leí", len(b))
	linea("¿hay retorno de carro?", strings.Contains(string(b), "\r"))
	linea("os.TempDir()", os.TempDir())
	_ = os.Remove(f)

	fmt.Println("\n-- 4 · PERMISOS --")
	_ = os.WriteFile(f, []byte("x"), 0o600)
	st, err := os.Stat(f)
	if err == nil {
		linea("pedí 0600, Stat dice", st.Mode().Perm())
		linea("¿es de solo lectura?", st.Mode().Perm()&0o200 == 0)
	}
	_ = os.Remove(f)
	linea("os.Getuid()", os.Getuid()) // en Windows devuelve -1: no existe el concepto

	fmt.Println("\n-- 5 · /proc, DE DONDE SALE TODO LO QUE SE ENSEÑA --")
	_, err = os.Stat("/proc/self/status")
	linea("¿existe /proc/self/status?", err == nil)
	if err != nil {
		linea("y el error es", err)
	}

	fmt.Println("\n-- 6 · EJECUTAR OTRO PROGRAMA --")
	// `exec.LookPath` es la forma correcta y portable. Escribir "ls" a pelo no cruza.
	for _, orden := range []string{"go", "ls", "cmd", "sh"} {
		if p, e := exec.LookPath(orden); e == nil {
			linea("LookPath("+orden+")", p)
		} else {
			linea("LookPath("+orden+")", "no está")
		}
	}
	linea("nombre de mi binario", filepath.Base(must(os.Executable())))
	linea("¿acaba en .exe?", strings.HasSuffix(must(os.Executable()), ".exe"))

	fmt.Println("\n-- 7 · SEÑALES --")
	// ⚠️ Se manda la señal 0, que es la sonda clásica: no hace nada, solo pregunta si se puede.
	//    Con SIGTERM o con la interrupción, en Linux el programa se moriría aquí mismo y no
	//    habría salida que enseñar. La señal 0 mide lo mismo sin matar a nadie.
	yo, _ := os.FindProcess(os.Getpid())
	linea("Signal(0) a mí mismo", resultado(yo.Signal(syscall.Signal(0))))
	// Y esta es la parte que NO cruza ni compilando: en Windows no hay SIGUSR1 ni SIGHUP,
	// así que el código que las usa va en un fichero _linux.go y no existe al otro lado.
	linea("y SIGUSR1 / SIGHUP", "no existen en Windows: van en un fichero _linux.go")
}

func resultado(e error) string {
	if e == nil {
		return "aceptada (la señal 0 no hace nada: solo comprueba)"
	}
	return "RECHAZADA: " + e.Error()
}

func must(s string, e error) string {
	if e != nil {
		return "ERROR: " + e.Error()
	}
	return s
}

// Crea una ruta anidada de ~300 caracteres y dice si el sistema la aguanta.
func rutaLarga() string {
	base := filepath.Join(os.TempDir(), "largo_"+fmt.Sprint(os.Getpid()))
	p := base
	for i := 0; i < 8; i++ {
		p = filepath.Join(p, strings.Repeat("d", 34)) // 8 x 35 = 280 y pico
	}
	defer os.RemoveAll(base)
	if e := os.MkdirAll(p, 0o755); e != nil {
		return fmt.Sprintf("FALLA a los %d caracteres: %v", len(p), e)
	}
	f := filepath.Join(p, "x.txt")
	if e := os.WriteFile(f, []byte("x"), 0o644); e != nil {
		return fmt.Sprintf("directorio sí (%d car.), fichero NO: %v", len(p), e)
	}
	return fmt.Sprintf("sí, %d caracteres", len(f))
}

func abre(p string) any {
	f, e := os.Open(p)
	if e != nil {
		return "no: " + e.Error()
	}
	f.Close()
	return "sí"
}
