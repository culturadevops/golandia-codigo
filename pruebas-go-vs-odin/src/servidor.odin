// El servidor de «¿Go o Odin?» (golandia 905). Mismo contrato que toda la serie «¿Go o X? Lo medí»: /suma (CPU) y
// /trabajo (20.000 objetos, mapa, orden y JSON armado a mano), con el mismo resultado que el servidor de Go
// (recursos/pruebas-go-java-revancha/go/main.go).
//
// Odin no trae servidor HTTP en su librería (core:net son sockets). Se usa odin-http (laytan), el que va camino de
// core, que reparte las conexiones entre hilos con io_uring. Variable HILOS: cuántos hilos atienden (si no, lo que
// decide odin-http: el número de núcleos que ve).
//
// Odin no tiene recolector de basura: todo lo de una petición va al temp_allocator de la conexión, que odin-http
// libera de golpe al terminar la petición.
package main

import "core:fmt"
import "core:net"
import "core:os"
import "core:slice"
import "core:strconv"
import "core:strings"
import "core:time"

import http "../odin-http"

Item :: struct {
	id:     int,
	nombre: string,
	valor:  int,
}

arranque: time.Tick
// NOMBRES=fmt arma «item-N» con fmt.tprintf (lo cómodo en Odin); por defecto, con strconv.itoa + concatenate, que es lo
// que hace el servidor de Go (strconv.Itoa + concatenación). Así se ve cuánto cuesta fmt.
usar_fmt: bool

num :: proc(req: ^http.Request) -> int {
	v, ok := http.query_get(req.url, "n")
	if !ok {return 0}
	n, _ := strconv.parse_int(v)
	return n
}

suma :: proc(n: int) -> string {
	t0 := time.tick_now()
	s: u64 = 0
	for i in 1 ..= n {
		s += u64(i) * u64(i) % 1000003
	}
	return fmt.tprintf(`{{"lang":"odin","n":%d,"suma":%d,"ns":%d}}`, n, s, i64(time.tick_since(t0)))
}

trabajo :: proc(n: int) -> string {
	t0 := time.tick_now()
	m := make(map[string]^Item, n, context.temp_allocator)
	for i in 0 ..< n {
		it := new(Item, context.temp_allocator)
		nombre: string
		if usar_fmt {
			nombre = fmt.tprintf("item-%d", i)
		} else {
			buf: [24]u8
			nombre = strings.concatenate({"item-", strconv.itoa(buf[:], i)}, context.temp_allocator)
		}
		it^ = Item{i, nombre, (i * 7919) % 1000}
		m[it.nombre] = it
	}
	lista := make([dynamic]^Item, 0, len(m), context.temp_allocator)
	total := 0
	for _, it in m {
		append(&lista, it)
		total += it.valor
	}
	slice.sort_by(lista[:], proc(a, b: ^Item) -> bool {
		if a.valor != b.valor {return a.valor < b.valor}
		return a.id < b.id
	})
	b := strings.builder_make(context.temp_allocator)
	fmt.sbprintf(&b, `{{"lang":"odin","n":%d,"total":%d,"primeros":[`, n, total)
	for i in 0 ..< min(50, len(lista)) {
		if i > 0 {strings.write_byte(&b, ',')}
		it := lista[i]
		fmt.sbprintf(&b, `{{"id":%d,"nombre":"%s","valor":%d}}`, it.id, it.nombre, it.valor)
	}
	fmt.sbprintf(&b, `],"ns":%d}}`, i64(time.tick_since(t0)))
	return strings.to_string(b)
}

manejar :: proc(h: ^http.Handler, req: ^http.Request, res: ^http.Response) {
	http.headers_set_content_type(&res.headers, "application/json")
	switch req.url.path {
	case "/suma":
		http.respond_plain(res, suma(num(req)))
	case "/trabajo":
		http.respond_plain(res, trabajo(num(req)))
	case:
		http.respond(res, http.Status.Not_Found)
	}
}

main :: proc() {
	arranque = time.tick_now()
	opts := http.Default_Server_Opts
	if v, ok := os.lookup_env("NOMBRES", context.temp_allocator); ok && v == "fmt" {
		usar_fmt = true
	}
	if v, ok := os.lookup_env("HILOS", context.temp_allocator); ok {
		opts.thread_count, _ = strconv.parse_int(v)
	}
	s: http.Server
	h := http.Handler{handle = manejar}
	fmt.eprintfln("LISTO_EN_MS %.3f NUCLEOS %d HILOS %d NOMBRES %s", time.duration_milliseconds(time.tick_since(arranque)),
		os.get_processor_core_count(), opts.thread_count, "fmt" if usar_fmt else "strconv")
	err := http.listen_and_serve(&s, h, net.Endpoint{address = net.IP4_Any, port = 8080}, opts)
	fmt.assertf(err == nil, "el servidor paró: %v", err)
}
