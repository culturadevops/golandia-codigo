// duraciones: cuánto dura cada conexión TCP saliente, medido en el kernel con eBPF.
package main

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -tags linux duraciones duraciones.c -- -I/usr/include/x86_64-linux-gnu

import (
	"bytes"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/netip"
	"os"
	"syscall"
	"time"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
)

type evento struct {
	Pid        uint32
	Comm       [16]byte
	Daddr      [4]byte
	Sport      uint16
	Dport      uint16
	_          [4]byte // relleno: el u64 siguiente va alineado a 8
	ConectarNs uint64
	TotalNs    uint64
}

func main() {
	dur := flag.Duration("durante", 5*time.Second, "cuánto tiempo escuchar")
	pid := flag.Uint("pid", 0, "solo este proceso (0 = todos)")
	sinNs := flag.Bool("sin-ns", false, "no traducir el PID a nuestro espacio de PIDs (para ver el problema)")
	flag.Parse()

	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatal(err)
	}
	spec, err := loadDuraciones()
	if err != nil {
		log.Fatal(err)
	}
	// Las constantes se escriben en el programa ANTES de cargarlo en el kernel.
	var st syscall.Stat_t
	if err := syscall.Stat("/proc/self/ns/pid", &st); err != nil {
		log.Fatal(err)
	}
	consts := map[string]any{"filtro_pid": uint32(*pid), "ns_dev": uint64(st.Dev), "ns_ino": st.Ino}
	if *sinNs {
		consts["ns_dev"], consts["ns_ino"] = uint64(0), uint64(0)
	}
	for nombre, valor := range consts {
		if err := spec.Variables[nombre].Set(valor); err != nil {
			log.Fatalf("poner %s: %v", nombre, err)
		}
	}
	var objs duracionesObjects
	if err := spec.LoadAndAssign(&objs, nil); err != nil {
		log.Fatalf("cargar el programa eBPF: %v", err)
	}
	defer objs.Close()

	tp, err := link.Tracepoint("sock", "inet_sock_set_state", objs.Duracion, nil)
	if err != nil {
		log.Fatalf("enganchar el tracepoint: %v", err)
	}
	defer tp.Close()

	rd, err := ringbuf.NewReader(objs.Eventos)
	if err != nil {
		log.Fatal(err)
	}
	defer rd.Close()
	fmt.Fprintf(os.Stderr, "escuchando %v (pid %d)…\n", *dur, *pid)
	go func() { time.Sleep(*dur); rd.Close() }()

	for {
		rec, err := rd.Read()
		if errors.Is(err, ringbuf.ErrClosed) {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		var e evento
		if err := binary.Read(bytes.NewReader(rec.RawSample), binary.LittleEndian, &e); err != nil {
			log.Fatal(err)
		}
		comm := string(bytes.TrimRight(e.Comm[:], "\x00"))
		fmt.Printf("%-8s pid %-6d → %s:%-5d conectar %-9v total %v\n", comm, e.Pid,
			netip.AddrFrom4(e.Daddr), e.Dport,
			time.Duration(e.ConectarNs).Round(time.Microsecond),
			time.Duration(e.TotalNs).Round(time.Millisecond))
	}
}
