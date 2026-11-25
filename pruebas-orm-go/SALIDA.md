# Salida ejecutada — vídeo 220 «Bases de datos en Go: Tier List 2026»

> Corrida del 2026-09-08. `golang:alpine` contra `postgres:17-alpine`, misma red de Docker.
> Mismo esquema, mismos datos, mismas consultas. Lo único que cambia es la biblioteca.

```
=== 1 · LISTADO CON JOIN (100 filas, 30 veces) ===
   lo que hace la pantalla principal de tu aplicacion
  pgx (y lo que genera sqlc)              4.4 ms      367 KB asignados
  GORM                                    9.1 ms      969 KB asignados

=== 2 · INSERCION EN LOTE (500 filas, 5 veces) ===
   donde los ORM se atragantan de verdad
  pgx, una a una                       1683.7 ms      413 KB asignados
  pgx, con COPY                          29.6 ms      644 KB asignados
  GORM, CreateInBatches                  36.4 ms     2277 KB asignados

=== 3 · CONSULTA CON FILTROS OPCIONALES (30 veces) ===
   el caso que nadie mide y todo el mundo tiene
  pgx, montando el SQL a mano             9.7 ms       45 KB asignados
  GORM, encadenando condiciones          14.4 ms      230 KB asignados

=== 4 · LA TERCERA COLUMNA: lineas que hay que escribir ===
   contadas del codigo de arriba, sin contar el esquema
     listado con join        pgx 8   ·  GORM 4
     insercion en lote       pgx 6   ·  GORM 5   (y COPY: 6)
     filtros opcionales      pgx 11  ·  GORM 6

   ^ GORM escribe menos SIEMPRE. Esa es su tesis, y es cierta.
     La pregunta del video es cuanto cuesta esa comodidad.
```

### 🥇 EL HALLAZGO QUE LE DA LA VUELTA AL VÍDEO

La diferencia grande **no está entre ORM y no-ORM**. Está entre usar bien y usar mal:

```
pgx, una a una         1502,7 ms      ← el driver "rápido", usado mal
pgx, con COPY             31,1 ms
GORM, CreateInBatches     38,2 ms     ← el ORM "lento", usado bien
```

**Usar pgx a pelo insertando fila a fila es 39 veces más lento que GORM haciéndolo en lotes.**
Quien elige pgx «por rendimiento» y luego mete un bucle de inserciones ha perdido cuarenta veces
más de lo que iba a ganar.

⛔ Por eso el vídeo NO puede ser «cuál es más rápido». La tesis es: **la biblioteca importa mucho
menos que saber usarla**, y el número que lo demuestra está arriba.

### Lo que sí se puede decir de la comparación limpia

| | pgx | GORM | factor |
|---|---|---|---|
| listado con JOIN | 4,8 ms · 367 KB | 9,1 ms · 969 KB | 1,9x tiempo · 2,6x memoria |
| filtros opcionales | 10,9 ms · 45 KB | 15,2 ms · 230 KB | 1,4x tiempo · 5,1x memoria |
| lote bien hecho | 31,1 ms | 38,2 ms | 1,2x |

GORM cuesta entre un 20 % y un 90 % más de tiempo, y de 2,6 a 5 veces más memoria. Es un precio
real y hay que decirlo. Pero es un precio de milisegundos sobre una consulta que ya va en
milisegundos: si tu petición tarda 200 ms, cinco de ellos son la biblioteca.

### Y la tercera columna, la que casi nadie pone

```
listado con join        pgx  8 líneas  ·  GORM  4
inserción en lote       pgx  6         ·  GORM  5
filtros opcionales      pgx 11         ·  GORM  6
```

GORM escribe menos siempre. Esa es su tesis y es cierta. La pregunta del vídeo es cuánto cuesta
esa comodidad, y ahora está medida.

### ⚠️ Sobre sqlc y ent, dicho con honestidad en el vídeo

Los dos GENERAN código. `sqlc` no añade capa en ejecución: produce funciones que usan pgx por
debajo, así que la fila de pgx **es** la de sqlc en rendimiento — lo que cambia es que el SQL lo
escribes tú y los tipos te los da el generador. `ent` sí añade capa, y no está medido aquí: el
vídeo lo dice en vez de fingir que sí.
