// conexiones: muestra cada conexión TCP nueva de la máquina, leída del kernel con eBPF.
package main

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -tags linux conexiones conexiones.c -- -I/usr/include/x86_64-linux-gnu

import (
	"bytes"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/netip"
	"os"
	"sort"
	"time"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
)

type evento struct {
	Pid      uint32
	Comm     [16]byte
	Saddr    [4]byte
	Daddr    [4]byte
	Sport    uint16
	Dport    uint16
	Saliente uint8
	_        [1]byte
}

func main() {
	dur := flag.Duration("durante", 5*time.Second, "cuánto tiempo escuchar")
	flag.Parse()

	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatal(err)
	}
	var objs conexionesObjects
	if err := loadConexionesObjects(&objs, nil); err != nil {
		log.Fatalf("cargar el programa eBPF: %v", err)
	}
	defer objs.Close()

	tp, err := link.Tracepoint("sock", "inet_sock_set_state", objs.Conexion, nil)
	if err != nil {
		log.Fatalf("enganchar el tracepoint: %v", err)
	}
	defer tp.Close()

	rd, err := ringbuf.NewReader(objs.Eventos)
	if err != nil {
		log.Fatal(err)
	}
	defer rd.Close()
	fmt.Fprintf(os.Stderr, "escuchando %v…\n", *dur)
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
		dir := "entrante"
		if e.Saliente == 1 {
			dir = "saliente"
		}
		comm := string(bytes.TrimRight(e.Comm[:], "\x00"))
		if e.Saliente == 1 { // al pasar a SYN_SENT el puerto de origen aún no está asignado
			fmt.Printf("%-8s %-10s pid %-6d %s → %s:%d\n", dir, comm, e.Pid,
				netip.AddrFrom4(e.Saddr), netip.AddrFrom4(e.Daddr), e.Dport)
		} else { // fuera del contexto del proceso: pid y comm no son de la conexión
			fmt.Printf("%-8s %s:%d → %s:%d\n", dir,
				netip.AddrFrom4(e.Saddr), e.Sport, netip.AddrFrom4(e.Daddr), e.Dport)
		}
	}

	// El mapa que el programa eBPF fue llenando en el kernel, leído desde Go.
	type par struct {
		puerto uint16
		n      uint64
	}
	var pares []par
	var k uint16
	var v uint64
	it := objs.PorPuerto.Iterate()
	for it.Next(&k, &v) {
		pares = append(pares, par{k, v})
	}
	sort.Slice(pares, func(i, j int) bool { return pares[i].puerto < pares[j].puerto })
	fmt.Println("--- conexiones por puerto del servicio (mapa del kernel)")
	for _, p := range pares {
		fmt.Printf("puerto %-5d %d\n", p.puerto, p.n)
	}
}
