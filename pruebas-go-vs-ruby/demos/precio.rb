$stdout.sync = true
# Un descuento que solo se aplica los lunes. El error está en la rama del lunes.
def precio_final(precio, dia)
  if dia == "lunes"
    precio - "10"      # error: resta un texto a un número
  else
    precio
  end
end

puts "martes: #{precio_final(100, "martes")}"
puts "miércoles: #{precio_final(100, "miércoles")}"
puts "lunes: #{precio_final(100, "lunes")}"
