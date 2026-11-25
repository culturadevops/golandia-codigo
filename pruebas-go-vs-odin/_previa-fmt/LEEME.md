Primera corrida (2026-11-23 12:14), con el servidor anterior: «item-N» se armaba con fmt.tprintf, y el servidor de Go
lo arma con strconv.Itoa + concatenación. No era el mismo trabajo en /trabajo, así que se cambió el servidor (por defecto
strconv; con NOMBRES=fmt, la forma anterior, que pasa a ser una variante más) y se midió otra vez (SALIDA-corrida-2.md).

La repetición con el servidor definitivo («odin 1», 13:22) se colgó 4 h: gl-905-srv al 199 % de CPU y el hey sin volver;
el coordinador mató el contenedor a las 17:43 y Docker quedó respondiendo «EOF». No se sabe si fue Odin o el banco: no se
cuenta como hallazgo. Esta corrida previa se cita en el vídeo 905 solo donde nada cambió (/suma, que no arma nombres) y
como contraste de la variante con fmt (363 aquí, 366 en SALIDA-corrida-2.md).
