(* El servidor de «¿Go o OCaml 5?» (golandia 903). Mismo contrato que toda la serie «¿Go o X? Lo medí»:
   /suma (CPU) y /trabajo (20.000 objetos, mapa, orden y JSON armado a mano), con el mismo resultado que el
   servidor de Go (recursos/pruebas-go-java-revancha/go/main.go).

   OCaml 5 trae multinúcleo de verdad con dominios (Domain). El HTTP va con cohttp-eio sobre Eio, que reparte las
   conexiones entre dominios con ~additional_domains. Variable DOMINIOS: cuántos dominios atienden (por defecto 1,
   que es lo que hace cohttp-eio si no le pasas nada). *)

let arranque = Mtime_clock.counter ()

type item = { id : int; nombre : string; valor : int }

let ns_desde c = Int64.to_string (Mtime.Span.to_uint64_ns (Mtime_clock.count c))

let suma n =
  let c = Mtime_clock.counter () in
  let s = ref 0 in
  for i = 1 to n do
    s := !s + (i * i mod 1000003)
  done;
  Printf.sprintf {|{"lang":"ocaml","n":%d,"suma":%d,"ns":%s}|} n !s (ns_desde c)

let trabajo n =
  let c = Mtime_clock.counter () in
  let m = Hashtbl.create n in
  for i = 0 to n - 1 do
    let it = { id = i; nombre = "item-" ^ string_of_int i; valor = i * 7919 mod 1000 } in
    Hashtbl.replace m it.nombre it
  done;
  let lista = Array.make (Hashtbl.length m) { id = 0; nombre = ""; valor = 0 } in
  let total = ref 0 and k = ref 0 in
  Hashtbl.iter (fun _ it -> lista.(!k) <- it; incr k; total := !total + it.valor) m;
  Array.sort (fun a b -> if a.valor <> b.valor then compare a.valor b.valor else compare a.id b.id) lista;
  let b = Buffer.create 4096 in
  Buffer.add_string b {|{"lang":"ocaml","n":|};
  Buffer.add_string b (string_of_int n);
  Buffer.add_string b {|,"total":|};
  Buffer.add_string b (string_of_int !total);
  Buffer.add_string b {|,"primeros":[|};
  for i = 0 to min 50 (Array.length lista) - 1 do
    if i > 0 then Buffer.add_char b ',';
    let it = lista.(i) in
    Buffer.add_string b {|{"id":|};
    Buffer.add_string b (string_of_int it.id);
    Buffer.add_string b {|,"nombre":"|};
    Buffer.add_string b it.nombre;
    Buffer.add_string b {|","valor":|};
    Buffer.add_string b (string_of_int it.valor);
    Buffer.add_char b '}'
  done;
  Buffer.add_string b {|],"ns":|};
  Buffer.add_string b (ns_desde c);
  Buffer.add_char b '}';
  Buffer.contents b

let headers = Http.Header.of_list [ ("content-type", "application/json") ]

let manejar _conn req _body =
  let uri = Uri.of_string (Http.Request.resource req) in
  let n = match Uri.get_query_param uri "n" with Some v -> (try int_of_string v with _ -> 0) | None -> 0 in
  match Uri.path uri with
  | "/suma" -> Cohttp_eio.Server.respond_string ~headers ~status:`OK ~body:(suma n) ()
  | "/trabajo" -> Cohttp_eio.Server.respond_string ~headers ~status:`OK ~body:(trabajo n) ()
  | _ -> Cohttp_eio.Server.respond_string ~status:`Not_found ~body:"" ()

let () =
  let dominios = match Sys.getenv_opt "DOMINIOS" with Some v -> int_of_string v | None -> 1 in
  Eio_main.run @@ fun env ->
  Eio.Switch.run @@ fun sw ->
  let socket =
    Eio.Net.listen env#net ~sw ~backlog:1024 ~reuse_addr:true (`Tcp (Eio.Net.Ipaddr.V4.any, 8080))
  in
  let srv = Cohttp_eio.Server.make ~callback:manejar () in
  Printf.eprintf "LISTO_EN_MS %.3f NUCLEOS %d DOMINIOS %d\n%!"
    (Mtime.Span.to_float_ns (Mtime_clock.count arranque) /. 1e6)
    (Domain.recommended_domain_count ()) dominios;
  Cohttp_eio.Server.run ~additional_domains:(env#domain_mgr, dominios - 1) ~on_error:(fun e -> Printf.eprintf "%s\n%!" (Printexc.to_string e)) socket srv
