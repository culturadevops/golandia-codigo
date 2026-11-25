go version go1.27.2 linux/amd64
$ go build -o bot ./bot && wc -c < bot
9753069
$ DISCORD_PUBLIC_KEY=<clave pública> ./bot &
$ go run ./simular   # firma cada interacción como Discord
# PING de Discord (type 1)
→ 200 en 807µs · {"type":1}
# misma petición con la firma alterada
→ 401 en 297µs · firma inválida
# /hola
→ 200 en 201µs · {"data":{"content":"¡Hola, golandia! Te respondo desde Go."},"type":4}
# /dado lados:20
→ 200 en 236µs · {"data":{"content":"🎲 Salió 5 (dado de 20)"},"type":4}
# /dado lados:1
→ 200 en 159µs · {"data":{"content":"Un dado necesita al menos 2 lados."},"type":4}
$ cat comandos.json   # lo que se registra en Discord (PUT /applications/<id>/commands, con el token del bot; NO se envía aquí)
[
  {"name": "hola", "description": "El bot te saluda", "type": 1},
  {"name": "dado", "description": "Tira un dado", "type": 1,
   "options": [{"name": "lados", "description": "Cuántos lados", "type": 4, "required": false}]}
]
