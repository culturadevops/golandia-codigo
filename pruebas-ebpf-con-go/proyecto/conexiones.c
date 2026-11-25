//go:build ignore

// Programa eBPF: se engancha al tracepoint sock/inet_sock_set_state del kernel y avisa a Go de
// cada conexión TCP nueva (saliente o entrante), sin tocar ninguna aplicación.
#include <linux/types.h>
#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

// El formato sale de /sys/kernel/tracing/events/sock/inet_sock_set_state/format.
struct inet_sock_set_state_args {
	__u64 common;          // common_type, flags, preempt_count, pid: 8 bytes
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
#define TCP_SYN_RECV 3
#define TCP_CLOSE 7
#define TCP_NEW_SYN_RECV 12

struct evento {
	__u32 pid;
	__u8 comm[16];
	__u8 saddr[4];
	__u8 daddr[4];
	__u16 sport;
	__u16 dport;
	__u8 saliente;
};

struct {
	__uint(type, BPF_MAP_TYPE_RINGBUF);
	__uint(max_entries, 1 << 16);
} eventos SEC(".maps");

struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 1024);
	__type(key, __u16);   // puerto del servicio
	__type(value, __u64); // conexiones
} por_puerto SEC(".maps");

SEC("tracepoint/sock/inet_sock_set_state")
int conexion(struct inet_sock_set_state_args *a) {
	// Un servidor que escucha en «:8080» usa un socket de doble pila: sus conexiones IPv4 llegan
	// como AF_INET6 con la dirección mapeada, y el tracepoint rellena igualmente saddr/daddr.
	if (a->family != AF_INET && a->family != AF_INET6)
		return 0;
	if (*(__u32 *)a->saddr == 0) // IPv6 de verdad: fuera de este ejemplo
		return 0;
	__u8 saliente;
	if (a->oldstate == TCP_CLOSE && a->newstate == TCP_SYN_SENT)
		saliente = 1;
	else if ((a->oldstate == TCP_SYN_RECV || a->oldstate == TCP_NEW_SYN_RECV) && a->newstate == TCP_ESTABLISHED)
		saliente = 0;
	else
		return 0;

	// Primero se leen los dos campos a variables: elegir con un ternario directamente sobre el
	// contexto (saliente ? a->dport : a->sport) lo rechaza el verificador del kernel.
	__u16 sport = a->sport, dport = a->dport;
	__u16 puerto = saliente ? dport : sport; // el puerto del servicio
	__u64 uno = 1, *n = bpf_map_lookup_elem(&por_puerto, &puerto);
	if (n)
		__sync_fetch_and_add(n, 1);
	else
		bpf_map_update_elem(&por_puerto, &puerto, &uno, BPF_ANY);

	struct evento *e = bpf_ringbuf_reserve(&eventos, sizeof(*e), 0);
	if (!e)
		return 0;
	e->pid = bpf_get_current_pid_tgid() >> 32;
	bpf_get_current_comm(&e->comm, sizeof(e->comm));
	__builtin_memcpy(e->saddr, a->saddr, 4);
	__builtin_memcpy(e->daddr, a->daddr, 4);
	e->sport = sport;
	e->dport = dport;
	e->saliente = saliente;
	bpf_ringbuf_submit(e, 0);
	return 0;
}

char LICENSE[] SEC("license") = "GPL";
