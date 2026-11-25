$stdout.sync = true
# El cálculo de /suma (200.000 vueltas), 400 veces: en serie, con 2 hilos y con 2 Ractors. Antes, una pasada de calentamiento
# (para que YJIT ya haya compilado suma y no se mida su compilación en la primera columna).
def suma(n)
  s = 0
  i = 1
  while i <= n
    s += (i * i) % 1000003
    i += 1
  end
  s
end

def reloj
  t = Process.clock_gettime(Process::CLOCK_MONOTONIC)
  yield
  ((Process.clock_gettime(Process::CLOCK_MONOTONIC) - t) * 1000).round
end

N, VECES = 200_000, 400
puts "ruby #{RUBY_VERSION} · núcleos visibles: #{Etc.nprocessors}" if require("etc")
VECES.times { suma(N) } # calentamiento, no se mide
serie = reloj { VECES.times { suma(N) } }
puts "en serie:        #{serie} ms"
hilos = reloj { 2.times.map { Thread.new { (VECES / 2).times { suma(N) } } }.each(&:join) }
puts "2 hilos:         #{hilos} ms"
ractors = reloj do
  rs = 2.times.map { Ractor.new(N, VECES / 2) { |n, v| v.times { suma(n) }; :ok } }
  rs.each(&:value)
end
puts "2 Ractors:       #{ractors} ms"
