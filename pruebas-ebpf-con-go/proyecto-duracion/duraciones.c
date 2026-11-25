//go:build ignore

// Programa eBPF: cuánto dura cada conexión TCP saliente, medido en el kernel.
// Mismo tracepoint que conexiones.c; ahora se mira la conexión de principio a fin.
#include <linux/types.h>
#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

struct inet_sock_set_state_args {
	__u64 common;
	const void *skaddr;
	int oldstate;
	int newstate;
	__u16 sport;
	__u16 dport;
	__u16 family;
	__u16 protocol;
	__u8 saddr[4];
	__u8 daddr[4];
	__u8 saddr_v6[16];
	__u8 daddr_v6[16];
};

#define AF_INET 2
#define AF_INET6 10
#define TCP_ESTABLISHED 1
#define TCP_SYN_SENT 2
#define TCP_CLOSE 7

// Las pone Go ANTES de cargar el programa. filtro_pid 0 = todos los procesos.
volatile const __u32 filtro_pid = 0;
// El espacio de PIDs de quien carga el programa (stat de /proc/self/ns/pid): sin esto, el kernel
// da el PID del espacio raíz, que no es el que ves con ps dentro de un contenedor.
volatile const __u64 ns_dev = 0;
volatile const __u64 ns_ino = 0;

static __always_inline __u32 pid_visible(void) {
	struct bpf_pidns_info ns = {};
	if (ns_ino && bpf_get_ns_current_pid_tgid(ns_dev, ns_ino, &ns, sizeof(ns)) == 0)
		return ns.tgid;
	return bpf_get_current_pid_tgid() >> 32; // fuera de nuestro espacio: el del kernel
}

struct inicio {
	__u64 ts_syn;   // al pasar a SYN_SENT
	__u64 ts_est;   // al pasar a ESTABLISHED
	__u32 pid;
	__u8 comm[16];
};

struct evento {
	__u32 pid;
	__u8 comm[16];
	__u8 daddr[4];
	__u16 sport;
	__u16 dport;
	__u64 conectar_ns; // SYN_SENT -> ESTABLISHED
	__u64 total_ns;    // SYN_SENT -> CLOSE
};

struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 4096);
	__type(key, __u64);            // la dirección del socket en el kernel
	__type(value, struct inicio);
} vivas SEC(".maps");

struct {
	__uint(type, BPF_MAP_TYPE_RINGBUF);
	__uint(max_entries, 1 << 16);
} eventos SEC(".maps");

SEC("tracepoint/sock/inet_sock_set_state")
int duracion(struct inet_sock_set_state_args *a) {
	if (a->family != AF_INET && a->family != AF_INET6)
		return 0;
	__u64 sk = (__u64)a->skaddr;
	int viejo = a->oldstate, nuevo = a->newstate;

	if (viejo == TCP_CLOSE && nuevo == TCP_SYN_SENT) {
		// Aquí sí estamos en el proceso que conecta: es donde se filtra.
		__u32 pid = pid_visible();
		if (filtro_pid && pid != filtro_pid)
			return 0;
		struct inicio i = {};
		i.ts_syn = bpf_ktime_get_ns();
		i.pid = pid;
		bpf_get_current_comm(&i.comm, sizeof(i.comm));
		bpf_map_update_elem(&vivas, &sk, &i, BPF_ANY);
		return 0;
	}

	struct inicio *i = bpf_map_lookup_elem(&vivas, &sk);
	if (!i)
		return 0; // no la vimos nacer (o no es del proceso filtrado)

	if (nuevo == TCP_ESTABLISHED) {
		i->ts_est = bpf_ktime_get_ns();
		return 0;
	}
	if (nuevo != TCP_CLOSE)
		return 0;

	struct evento *e = bpf_ringbuf_reserve(&eventos, sizeof(*e), 0);
	if (e) {
		__u64 ahora = bpf_ktime_get_ns();
		e->pid = i->pid;
		__builtin_memcpy(e->comm, i->comm, sizeof(e->comm));
		__builtin_memcpy(e->daddr, a->daddr, 4);
		__u16 sport = a->sport, dport = a->dport;
		e->sport = sport;
		e->dport = dport;
		e->conectar_ns = i->ts_est ? i->ts_est - i->ts_syn : 0;
		e->total_ns = ahora - i->ts_syn;
		bpf_ringbuf_submit(e, 0);
	}
	bpf_map_delete_elem(&vivas, &sk);
	return 0;
}

char LICENSE[] SEC("license") = "GPL";
