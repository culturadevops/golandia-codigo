# El servidor de la revancha (recursos/pruebas-go-java-revancha/go/main.go) en Ruby, para «¿Go o Ruby?» (golandia 804).
# El mismo contrato: /suma (CPU) y /trabajo (objetos, Hash, orden y JSON armado a mano). Rack puro sobre Puma, sin Rails.
ARRANQUE = Process.clock_gettime(Process::CLOCK_MONOTONIC)

Item = Struct.new(:id, :nombre, :valor)

def ahora_ns = Process.clock_gettime(Process::CLOCK_MONOTONIC, :nanosecond)

def suma(n)
  t0 = ahora_ns
  s = 0
  i = 1
  while i <= n
    s += (i * i) % 1000003
    i += 1
  end
  %({"lang":"ruby","n":#{n},"suma":#{s},"ns":#{ahora_ns - t0}})
end

def trabajo(n)
  t0 = ahora_ns
  m = {}
  n.times do |i|
    it = Item.new(i, "item-#{i}", (i * 7919) % 1000)
    m[it.nombre] = it
  end
  lista = m.values
  total = lista.sum(&:valor)
  lista.sort_by! { |it| [it.valor, it.id] }
  partes = lista.first(50).map { |it| %({"id":#{it.id},"nombre":"#{it.nombre}","valor":#{it.valor}}) }
  %({"lang":"ruby","n":#{n},"total":#{total},"primeros":[#{partes.join(",")}],"ns":#{ahora_ns - t0}})
end

JSON_H = { "content-type" => "application/json" }.freeze

app = lambda do |env|
  q = Rack::Utils.parse_query(env["QUERY_STRING"])
  n = q["n"].to_i
  case env["PATH_INFO"]
  when "/suma" then [200, JSON_H, [suma(n)]]
  when "/trabajo" then [200, JSON_H, [trabajo(n)]]
  else [404, JSON_H, ["{}"]]
  end
end

warn format("ARRANQUE_CONFIG_MS %.3f ruby %s yjit=%s zjit=%s", (Process.clock_gettime(Process::CLOCK_MONOTONIC) - ARRANQUE) * 1000,
            RUBY_VERSION, defined?(RubyVM::YJIT) && RubyVM::YJIT.enabled?, defined?(RubyVM::ZJIT) && RubyVM::ZJIT.enabled?)
run app
