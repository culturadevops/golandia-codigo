# El servidor de «¿Go o Crystal?» (golandia 902). Mismo contrato que toda la serie «¿Go o X? Lo medí»:
# /suma (CPU) y /trabajo (20.000 objetos, mapa, orden y JSON armado a mano), con el mismo resultado que el
# servidor de Go (recursos/pruebas-go-java-revancha/go/main.go). Solo librería estándar: HTTP::Server.
#
# Crystal ejecuta por defecto UNA fibra a la vez (el contexto por defecto tiene paralelismo 1). Con la variable
# HILOS se agranda el contexto por defecto (Fiber::ExecutionContext.default.resize) para usar más núcleos.
require "http/server"

ARRANQUE = Time.instant

class Item
  getter id : Int32
  getter nombre : String
  getter valor : Int32

  def initialize(@id, @nombre, @valor)
  end
end

def suma(n : Int32) : String
  t0 = Time.instant
  s = 0_u64
  i = 1_u64
  while i <= n
    s &+= i &* i % 1000003_u64
    i += 1
  end
  ns = (Time.instant - t0).total_nanoseconds.to_i64
  %({"lang":"crystal","n":#{n},"suma":#{s},"ns":#{ns}})
end

def trabajo(n : Int32) : String
  t0 = Time.instant
  m = Hash(String, Item).new(initial_capacity: n)
  n.times do |i|
    it = Item.new(i, "item-#{i}", ((i.to_i64 * 7919) % 1000).to_i32)
    m[it.nombre] = it
  end
  lista = Array(Item).new(m.size)
  total = 0
  m.each_value do |it|
    lista << it
    total += it.valor
  end
  lista.sort! do |a, b|
    a.valor != b.valor ? a.valor <=> b.valor : a.id <=> b.id
  end
  String.build do |b|
    b << %({"lang":"crystal","n":) << n << %(,"total":) << total << %(,"primeros":[)
    {50, lista.size}.min.times do |i|
      b << ',' if i > 0
      it = lista[i]
      b << %({"id":) << it.id << %(,"nombre":") << it.nombre << %(","valor":) << it.valor << '}'
    end
    b << %(],"ns":) << (Time.instant - t0).total_nanoseconds.to_i64 << '}'
  end
end

def num(ctx) : Int32
  (ctx.request.query_params["n"]? || "0").to_i? || 0
end

if h = ENV["HILOS"]?
  Fiber::ExecutionContext.default.resize(h.to_i)
end

servidor = HTTP::Server.new do |ctx|
  ctx.response.content_type = "application/json"
  case ctx.request.path
  when "/suma"    then ctx.response.print suma(num(ctx))
  when "/trabajo" then ctx.response.print trabajo(num(ctx))
  else
    ctx.response.status_code = 404
  end
end
servidor.bind_tcp "0.0.0.0", 8080
STDERR.puts "LISTO_EN_MS #{(Time.instant - ARRANQUE).total_milliseconds.round(3)} NUCLEOS #{Fiber::ExecutionContext.default_workers_count} HILOS #{ENV["HILOS"]? || "1 (por defecto)"}"
servidor.listen
