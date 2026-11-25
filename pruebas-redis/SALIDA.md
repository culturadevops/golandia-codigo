# Salida ejecutada — vídeo 198 «Escribí Redis en Go»

> Corrida del 2026-09-05. Servidor: `redis.go` en `golang:alpine`.
> Cliente: **redis-cli 7.4.11 OFICIAL**, de la imagen `redis:7-alpine`, por red de Docker.
> ⚠️ En esta máquina App Control bloquea binarios sin firmar, así que el servidor NO puede
> correr en el host: todo va en contenedor.

## La promesa del título: el cliente oficial no nota la diferencia
```
redis-cli 7.4.11

PING            -> PONG
SET saludo hola -> OK
GET saludo      -> hola
EXISTS saludo   -> 1
GET no-existe   -> []  (nil)
INCR visitas    -> 1
INCR visitas    -> 2
DECR visitas    -> 1
SET s v EX 2    -> OK
TTL s           -> 2
(3 s despues)  GET s -> []   TTL s -> -2  (-2 = no existe)
DBSIZE          -> 4
```

## El benchmark: un hilo contra mutex

`redis-benchmark` oficial, `-t set,get,incr -n 20000 -c 50 -q`, mismo servidor, solo cambia `-modo`:

```
===== modo=mutex =====
 SET:  200.000,00 requests per second, p50=0.135 msec
 GET:  202.020,20 requests per second, p50=0.135 msec
 INCR: 198.019,80 requests per second, p50=0.135 msec

===== modo=hilo =====
 SET:  188.679,25 requests per second, p50=0.143 msec
 GET:  204.081,64 requests per second, p50=0.135 msec
 INCR: 198.019,80 requests per second, p50=0.135 msec
```

### ⛔ EL HALLAZGO DESMONTA LA HIPÓTESIS DEL ENCARGO, y el guion se cambia

El encargo daba por hecho que la versión de un solo hilo ganaría y que la del mutex se hundiría
por contención. **No pasa: empatan.** A 200.000 operaciones por segundo el candado no es el cuello
de botella — lo son la red y el parseo de RESP.

Y la contención está al máximo, no es que la prueba sea floja: sin `-r`, `redis-benchmark` machaca
**la misma clave** con 50 clientes. Peor caso posible para un mutex, y aun así no se nota.

**Cómo queda el vídeo:** el giro deja de ser «un hilo es más rápido» —que sería mentira— y pasa a
ser el que de verdad se aprende escribiéndolo: *esperaba que ganara el hilo único y no ganó, porque
el cuello de botella estaba en otro sitio*. Redis no es de un solo hilo por velocidad bruta a esta
escala: es por **simplicidad y atomicidad gratis**. Eso sí se sostiene y sí se puede demostrar.

## 🐛 El fallo que encontré escribiéndolo, y que se cuenta en el vídeo

Primera versión de `conCandado` en modo hilo: `peticiones <- f` y seguir. Con un canal sin buffer,
el envío termina cuando el dueño del mapa **recibe** la función, no cuando la **ejecuta**. Así que
`ejecutar` leía su variable de respuesta antes de que nadie la escribiera y contestaba la cadena
vacía. El `redis-cli` oficial **se quedaba colgado** esperando un RESP que nunca llegaba entero:
sin error, sin traza, sin nada en los registros. Se arregla haciendo que el que pide espere —que
es justo lo que hace Redis de verdad.
