// mini: un runtime de contenedores de juguete, para ver qué hace Docker por debajo.
// Solo Linux y como root:
//
//	mini run [-hostname h] [-mem 64M] [-pids 20] [-red] <rootfs> <cmd> [args...]
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const raizCG = "/sys/fs/cgroup"

func main() {
	if len(os.Args) < 2 {
		uso()
	}
	switch os.Args[1] {
	case "run":
		os.Exit(run(os.Args[2:]))
	case "hijo": // nos llamamos a nosotros mismos, ya dentro de los namespaces nuevos
		hijo(os.Args[2], os.Args[3], os.Args[4:])
	default:
		uso()
	}
}

func uso() {
	fmt.Fprintln(os.Stderr, "uso: mini run [-hostname h] [-mem 64M] [-pids 20] [-red] <rootfs> <cmd> [args...]")
	os.Exit(2)
}

func run(args []string) int {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	host := fs.String("hostname", "mini", "hostname del contenedor")
	mem := fs.String("mem", "", "memory.max, por ejemplo 64M (vacío: sin límite)")
	pids := fs.String("pids", "", "pids.max, por ejemplo 20 (vacío: sin límite)")
	red := fs.Bool("red", false, "namespace de red propio (solo loopback, apagada)")
	fs.Parse(args)
	if fs.NArg() < 2 {
		uso()
	}
	rootfs, _ := filepath.Abs(fs.Arg(0))

	// 1 · cgroup v2: una carpeta en /sys/fs/cgroup; los límites son archivos
	cg := filepath.Join(raizCG, fmt.Sprintf("mini-%d", os.Getpid()))
	must(os.Mkdir(cg, 0o755))
	defer borrar(cg)
	if *mem != "" {
		escribir(cg, "memory.max", *mem)
		escribir(cg, "memory.swap.max", "0") // sin esto, lo que sobra se va a swap y nadie muere
	}
	if *pids != "" {
		escribir(cg, "pids.max", *pids)
	}
	fd, err := syscall.Open(cg, syscall.O_DIRECTORY|syscall.O_RDONLY, 0)
	must(err)
	defer syscall.Close(fd)

	// 2 · namespaces: el hijo nace con hostname, PIDs y montajes propios
	flags := syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS
	if *red {
		flags |= syscall.CLONE_NEWNET
	}
	cmd := exec.Command("/proc/self/exe", append([]string{"hijo", *host, rootfs}, fs.Args()[1:]...)...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:  uintptr(flags),
		UseCgroupFD: true, // clone3 con CLONE_INTO_CGROUP: nace ya dentro del cgroup
		CgroupFD:    fd,
	}
	err = cmd.Run()
	return informe(cg, cmd.ProcessState, err)
}

func hijo(host, rootfs string, argv []string) {
	must(syscall.Sethostname([]byte(host)))
	// que nada de lo que montemos aquí se propague al anfitrión
	must(syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""))
	// pivot_root exige que la raíz nueva sea un punto de montaje
	must(syscall.Mount(rootfs, rootfs, "", syscall.MS_BIND|syscall.MS_REC, ""))
	vieja := filepath.Join(rootfs, ".vieja")
	must(os.MkdirAll(vieja, 0o700))
	must(syscall.PivotRoot(rootfs, vieja))
	must(os.Chdir("/"))
	// soltamos la raíz vieja: el disco del anfitrión desaparece
	must(syscall.Unmount("/.vieja", syscall.MNT_DETACH))
	must(os.Remove("/.vieja"))
	// un /proc nuevo: solo enseña los procesos de este namespace
	must(syscall.Mount("proc", "/proc", "proc", syscall.MS_NOSUID|syscall.MS_NOEXEC|syscall.MS_NODEV, ""))
	// un /dev mínimo, como el de Docker: tmpfs y tres dispositivos
	must(syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID, "mode=755"))
	for nombre, menor := range map[string]int{"null": 3, "zero": 5, "urandom": 9} {
		must(syscall.Mknod("/dev/"+nombre, syscall.S_IFCHR|0o666, 1<<8|menor))
	}
	env := []string{"PATH=/usr/sbin:/usr/bin:/sbin:/bin", "HOME=/root", "HOSTNAME=" + host}
	os.Setenv("PATH", env[0][5:])
	ruta, err := exec.LookPath(argv[0])
	must(err)
	must(syscall.Exec(ruta, argv, env)) // el comando pasa a ser el PID 1
}

func informe(cg string, st *os.ProcessState, err error) int {
	codigo := 0
	var ee *exec.ExitError
	switch {
	case st == nil:
		fmt.Fprintln(os.Stderr, "[mini] no arrancó:", err)
		return 1
	case st.Sys().(syscall.WaitStatus).Signaled():
		fmt.Fprintf(os.Stderr, "[mini] el proceso murió por la señal %v\n", st.Sys().(syscall.WaitStatus).Signal())
		codigo = 137
	case errors.As(err, &ee):
		codigo = ee.ExitCode()
	}
	fmt.Fprintf(os.Stderr, "[mini] memory.peak=%s oom_kill=%s pids.peak=%s pids.max_rechazos=%s\n",
		leer(cg, "memory.peak"), campo(cg, "memory.events", "oom_kill"), leer(cg, "pids.peak"), campo(cg, "pids.events", "max"))
	return codigo
}

func escribir(cg, archivo, valor string) {
	must(os.WriteFile(filepath.Join(cg, archivo), []byte(valor), 0o644))
}

func leer(cg, archivo string) string {
	b, _ := os.ReadFile(filepath.Join(cg, archivo))
	return strings.TrimSpace(string(b))
}

func campo(cg, archivo, clave string) string {
	for _, l := range strings.Split(leer(cg, archivo), "\n") {
		if k, v, _ := strings.Cut(l, " "); k == clave {
			return v
		}
	}
	return "?"
}

// borrar quita el cgroup; si quedan procesos muriendo, espera un poco
func borrar(cg string) {
	for i := 0; i < 50 && os.Remove(cg) != nil; i++ {
		time.Sleep(20 * time.Millisecond)
	}
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "[mini]", err)
		os.Exit(1)
	}
}
