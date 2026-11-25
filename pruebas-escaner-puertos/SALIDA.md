# Salida ejecutada — vídeo «Escribí un escáner de puertos en Go»

> Corrida del 2026-09-10. Todo en contenedores de `golang:1.26-bookworm`, en una red propia.
> Reproducible con `bash demo.sh`.

## ⛔ ÉTICO, y se dice en el vídeo

Todo corre contra **un objetivo que levanto yo**, en mi propia red de contenedores, con puertos
que yo mismo abro. No se toca ninguna máquina ajena. Escanear una máquina que no es tuya sin
permiso por escrito es otra cosa, y no es lo que hace este banco.

## ⛔ Y HAY UNA VERDAD CONTRA LA QUE COMPARAR

El objetivo abre **15 puertos conocidos**:

```
OBJETIVO LISTO: 15 puertos abiertos de 15 pedidos
   [22 80 443 3306 5432 6379 8080 8443 9090 9200 11211 15672 27017 50000 61616]
```

Así que no se mide solo el tiempo: se mide **si el escáner los encuentra**. Un escáner rápido que
se deja puertos abiertos no es rápido, es inútil.

---

## ⛔ MI TESIS ERA FALSA, Y LA PRIMERA CORRIDA LO DIJO

Yo iba a demostrar que 65.535 goroutines se dejan puertos. En la primera corrida **no pasó**:

| concurrencia | abiertos | tiempo | errores |
|---|---|---|---|
| 100 | **15** ✅ | 378 ms | rechazado=65.520 |
| 500 | **15** ✅ | 384 ms | rechazado=65.520 |
| 2.000 | **15** ✅ | 490 ms | rechazado=65.520 |
| 10.000 | **15** ✅ | 606 ms | rechazado=65.520 |
| 65.535 | **15** ✅ | 714 ms | **TIMEOUT=3.647** · rechazado=61.873 |

Los encontró todos, siempre. Lo único que se degradó fue el tiempo: **378 ms con 100 goroutines,
714 ms con 65.535**. Casi el doble de lento por 655 veces más goroutines.

Podría haber publicado eso. Pero había un número que no cuadraba:

```
   descriptores de fichero del contenedor: 1048576
```

**Un millón de descriptores.** Una máquina normal tiene 1.024. Estaba midiendo en un sitio donde
el fallo que buscaba no puede ocurrir.

---

## 🥇 EL MISMO ESCÁNER, EN UNA MÁQUINA NORMAL (1.024 descriptores)

```
   descriptores: 1024
```

| concurrencia | abiertos | tiempo | errores |
|---|---|---|---|
| 500 | **15 de 15** ✅ | 586 ms | rechazado=65.520 |
| 10.000 | **8 de 15** ❌ | **367 ms** | SIN-DESCRIPTORES=27.718 · DNS caído=8.900 · rechazado=28.909 |
| 65.535 | **5 de 15** ❌ | **315 ms** | SIN-DESCRIPTORES=56.432 · rechazado=9.098 |

### Y esto es lo peor de todo

```
   500     ->  15 abiertos  en 586 ms
   65.535  ->   5 abiertos  en 315 ms
```

**El que se deja diez puertos es el que parece el mejor.** Es casi el doble de rápido en el
reloj. Si mides el tiempo y no la respuesta, eliges exactamente el peor.

Y no falla con un error: te devuelve una lista. Una lista corta, y tú tan tranquilo.

```
   con 500:     [22 80 443 3306 5432 6379 8080 8443 9090 9200 11211 15672 27017 50000 61616]
   con 65.535:  [22 80 443 3306 50000]
```

### El segundo modo de fallo, que no buscaba

Con 10.000 goroutines aparecieron **8.900 fallos de resolución de nombres**:

```
otro:dial tcp: lookup objetivo on 127.0.0.11:53: no such host = 8900
```

Diez mil resoluciones a la vez tumbaron el resolvedor. Así que la concurrencia sin freno no solo
agota tus descriptores: también revienta servicios de los que dependes sin pensarlo, y el error
que sale no menciona los descriptores por ninguna parte.

---

## Contra nmap, sobre el mismo objetivo

```
   nmap: 15 abiertos en 607 ms
```

Mi escáner con 500 goroutines encuentra lo mismo en 384 ms (con un millón de descriptores) o 586
ms (con 1.024). Es decir: **están a la par**, y eso solo en este caso concreto.

⚠️ Y hay que decir en voz alta qué NO hace mi escáner y sí hace nmap: identificar el servicio y
su versión, detectar el sistema operativo, ajustar los tiempos según lo que responde la red,
UDP, y no arrasar con la máquina de enfrente. Empatar en tiempo en un caso de laboratorio no es
empatar en nada más.

## El suelo: secuencial

```
   concurrencia=1  ->  3 abiertos (de los 3 que hay en ese rango)  en 330 ms
```

⚠️ Ese es de **2.000 puertos**, no de 65.535, porque si no no termina en un rato razonable. Para
los 65.535 serían unas 30 veces más. La concurrencia no es opcional aquí: el problema es esperar
a la red 65.535 veces.

---

## ⛔ Los dos fallos de este banco, los dos míos

1. **Medí donde el fallo no podía ocurrir.** Un contenedor de Docker trae un millón de
   descriptores por defecto. Con eso, el escáner salvaje no falla nunca, y mi conclusión habría
   sido «da igual la concurrencia, solo va más lento». Es falsa fuera del contenedor.
2. **Mi tesis de partida era la contraria a lo que salió primero.** La primera tabla dice que
   65.535 goroutines encuentran los 15 puertos. Está en este documento porque es verdad **en esas
   condiciones**, y porque enseñar solo la corrida que me da la razón es justo lo que hace que
   estas mediciones no valgan nada.

## ⛔ Lo que NO mide este banco

- **Una red de verdad.** Aquí el objetivo está a un salto, en la misma máquina. Con latencia real
  y pérdida de paquetes, los tiempos de espera mandan mucho más que la concurrencia.
- **Cortafuegos ni detección.** Un escaneo así, contra algo que no sea tuyo, se ve desde el otro
  lado y salta en cualquier sistema de detección medio decente.
- **UDP.** Solo TCP, y con una conexión completa, que es lo más ruidoso que hay.
- **El límite real de tu máquina**, que depende de tu sistema. 1.024 es lo típico; se puede subir,
  y subirlo sin entender lo que estás haciendo es otra forma de romper cosas.
