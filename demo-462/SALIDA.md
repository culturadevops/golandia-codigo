$ grep modernc go.mod
require modernc.org/sqlite v1.39.0

$ ./login ana gato123
entrada: usuario="ana" clave="gato123"
A) fmt.Sprintf
  consulta: SELECT nombre FROM usuarios WHERE usuario = 'ana' AND clave = 'gato123'
  resultado: ENTRASTE como Ana
B) placeholder ?
  resultado: ENTRASTE como Ana

$ ./login ana nolase
entrada: usuario="ana" clave="nolase"
A) fmt.Sprintf
  consulta: SELECT nombre FROM usuarios WHERE usuario = 'ana' AND clave = 'nolase'
  resultado: acceso denegado ( sql: no rows in result set )
B) placeholder ?
  resultado: acceso denegado ( sql: no rows in result set )

$ ./login admin "' OR '1'='1"
entrada: usuario="admin" clave="' OR '1'='1"
A) fmt.Sprintf
  consulta: SELECT nombre FROM usuarios WHERE usuario = 'admin' AND clave = '' OR '1'='1'
  resultado: ENTRASTE como Administración
B) placeholder ?
  resultado: acceso denegado ( sql: no rows in result set )

$ ./login "admin' --" cualquiera
entrada: usuario="admin' --" clave="cualquiera"
A) fmt.Sprintf
  consulta: SELECT nombre FROM usuarios WHERE usuario = 'admin' --' AND clave = 'cualquiera'
  resultado: ENTRASTE como Administración
B) placeholder ?
  resultado: acceso denegado ( sql: no rows in result set )
