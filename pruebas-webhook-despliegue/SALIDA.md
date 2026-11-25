# SALIDA — webhook de despliegue en Go (golandia 563)

Ejecutado el 2026-10-21 con `bash correr.sh` (golang:1.27.2, --cpus 2, todo dentro de mktemp). Salida LITERAL de la segunda corrida; los tiempos varían unos ms entre corridas.

```
== 0 · el ejemplo de la documentación de GitHub
sha256=757107ea0eb2509fc211221cce984b8a37570b6d7586c22c46f4379c8b043e17

== 1 · arranque: el receptor despliega lo que hay en main
app: hola desde v1

== 2 · push firmado a main
firma: sha256=e9fb5b76fec5b78fb211b1faca9ef4e91e2d1cabb049c12dd51603dd0e78a0d1
HTTP 202 · despliegue #1 en cola
app: hola desde v2

== 3 · firma hecha con otro secreto
HTTP 401 · firma inválida
== 4 · sin firma
HTTP 401 · firma inválida
== 5 · firma buena del cuerpo de antes, cuerpo cambiado
HTTP 401 · firma inválida
== 6 · la misma entrega otra vez (reenvío)
HTTP 409 · entrega repetida
== 7 · push a otra rama
HTTP 200 · ignorado: refs/heads/pruebas no es refs/heads/main
app sigue en: hola desde v2
HTTP 202 · despliegue #2 en cola
app: hola desde v4

== 8 · dos pushes casi a la vez (con cola)
HTTP 202 · despliegue #3 en cola
HTTP 202 · despliegue #4 en cola
app: hola desde v6

== 9 · cuánto tiempo no contesta la app durante un despliegue
HTTP 202 · despliegue #5 en cola
sondeo: 399 respuestas, 1 fallos
app: hola desde v7

== log del receptor (con cola)
     0 ms  receptor escuchando en :9000 (sin-cola=false)
     0 ms  despliegue #0: empieza
   327 ms  despliegue #0: listo en 327 ms · 7191371 · la app dice: hola desde v1
   394 ms  aceptado: despliegue #1 (335e2b3)
   394 ms  despliegue #1: empieza
   645 ms      la app no contestó durante 4.9 ms (de parar la vieja a que contesta la nueva)
   645 ms  despliegue #1: listo en 251 ms · 335e2b3 · la app dice: hola desde v2
   758 ms  rechazado: firma inválida ("sha256=e8d89408f3bfe…")
   765 ms  rechazado: firma inválida ("")
   773 ms  rechazado: firma inválida ("sha256=e9fb5b76fec5b…")
   778 ms  rechazado: entrega repetida entrega-0002
   788 ms  ignorado: push a refs/heads/pruebas
  1828 ms  aceptado: despliegue #2 (6cb0edc)
  1828 ms  despliegue #2: empieza
  2065 ms      la app no contestó durante 4.1 ms (de parar la vieja a que contesta la nueva)
  2065 ms  despliegue #2: listo en 237 ms · 6cb0edc · la app dice: hola desde v4
  2201 ms  aceptado: despliegue #3 (53567b6)
  2201 ms  despliegue #3: empieza
  2201 ms  aceptado: despliegue #4 (c03e262)
  2438 ms      la app no contestó durante 4.1 ms (de parar la vieja a que contesta la nueva)
  2438 ms  despliegue #3: listo en 236 ms · c03e262 · la app dice: hola desde v6
  2438 ms  despliegue #4: empieza
  2445 ms  despliegue #4: nada que hacer, c03e262 ya está desplegado
  4870 ms  aceptado: despliegue #5 (6b7c254)
  4870 ms  despliegue #5: empieza
  5152 ms      la app no contestó durante 4.4 ms (de parar la vieja a que contesta la nueva)
  5152 ms  despliegue #5: listo en 281 ms · 6b7c254 · la app dice: hola desde v7

== 10 · dos pushes a la vez SIN cola (cada uno en su goroutine)
HTTP 202 · despliegue #1 en cola
HTTP 202 · despliegue #2 en cola
HTTP 202 · despliegue #3 en cola
HTTP 202 · despliegue #4 en cola
HTTP 202 · despliegue #6 en cola
HTTP 202 · despliegue #5 en cola
app: hola desde v8-2-b
== log del receptor (sin cola)
     0 ms  receptor escuchando en :9000 (sin-cola=true)
     0 ms  despliegue #0: empieza
   223 ms  despliegue #0: listo en 223 ms · 6b7c254 · la app dice: hola desde v7
   372 ms  aceptado: despliegue #1 (c488edf)
   372 ms  despliegue #1: empieza
   372 ms  aceptado: despliegue #2 (0aee3be)
   372 ms  despliegue #2: empieza
   385 ms  despliegue #2 FALLÓ en git pull: error: cannot lock ref 'refs/remotes/origin/main': is at 0aee3becd1b8ef4c4c596d51178ccc3e2a29bbb4 but expected 6b7c254ccbe013ad3c690aaa04d723c9d2ac661f
   613 ms      la app no contestó durante 6.5 ms (de parar la vieja a que contesta la nueva)
   613 ms  despliegue #1: listo en 240 ms · 0aee3be · la app dice: hola desde v8-1-b
  4417 ms  aceptado: despliegue #4 (2ac6e04)
  4417 ms  despliegue #3: empieza
  4417 ms  aceptado: despliegue #3 (54170d8)
  4417 ms  despliegue #4: empieza
  4429 ms  despliegue #3 FALLÓ en git pull: error: cannot lock ref 'refs/remotes/origin/main': is at 54170d884cdf247fe871966d30f2bcc1697a0864 but expected 0aee3becd1b8ef4c4c596d51178ccc3e2a29bbb4
  4661 ms      la app no contestó durante 4.7 ms (de parar la vieja a que contesta la nueva)
  4662 ms  despliegue #4: listo en 244 ms · 54170d8 · la app dice: hola desde v8-2-b
  8465 ms  aceptado: despliegue #5 (a8b9da0)
  8465 ms  despliegue #6: empieza
  8465 ms  aceptado: despliegue #6 (cc6b791)
  8465 ms  despliegue #5: empieza
  8474 ms  despliegue #6 FALLÓ en git pull: fatal: Cannot fast-forward to multiple branches.
  8474 ms  despliegue #5 FALLÓ en git pull: error: cannot lock ref 'refs/remotes/origin/main': is at a8b9da05e213b5b9f7a312b0fb5e849c81d8c8f2 but expected 54170d884cdf247fe871966d30f2bcc1697a0864
FIN
```

## Primera corrida (misma demo, antes de medir el hueco del reinicio)

- Sección 0: la misma firma `sha256=757107ea…043e17`.
- Sección 8 con cola: #3 desplegó v6 directamente (el pull ya traía los dos commits) y #4 «nada que hacer».
- Sección 9: el sondeo vio 400 respuestas y 0 fallos (con sleep de 10 ms entre peticiones).
- Sección 10 sin cola: ronda 1, #1 falló en git pull (cannot lock ref) y #2 desplegó v8-1-b; ronda 2, #3 y #4 fallaron (cannot lock ref / Cannot fast-forward to multiple branches); ronda 3, #5 y #6 fallaron igual. La app quedó en v8-1-b con main en v8-3-b, y GitHub habría recibido 202 en las seis entregas.

## Resumen de las dos corridas

- Sin cola: 6 de 6 rondas con al menos un despliegue fallido; en 3 de 6 fallaron los dos y la app quedó atrás de main.
- Con cola: 0 fallos; el segundo push de una pareja encuentra el commit ya desplegado.
- Hueco del reinicio (de parar la vieja a que contesta la nueva): 4,1 a 6,5 ms en esta corrida.

## Extra: ping y content type de formulario (`ejecutar-extra.sh`, misma imagen, 2026-10-21)

```
== 11 · el ping que GitHub manda al crear el webhook
HTTP 200 · pong
== 12 · push con content type de formulario (application/x-www-form-urlencoded)
payload=%7B%22ref%22%3A%22refs/heads/main%22%2C%22after%22%3…
HTTP 400 · json inválido
FIN
```
