package main

// El almacen. Es lo unico que guarda estado, y por eso es lo unico que lleva candado.
//
// Decision de diseno del episodio 1: TODO en memoria. No hay disco, no hay base de datos.
// Cuando matas el proceso, se acabo. Eso no es una limitacion: es lo que hace que una prueba
// arranque limpia siempre, que es justo para lo que sirve un emulador.
//
// El candado es un RWMutex y no un Mutex porque el trafico real de un emulador de S3 es
// muchisimo mas de lectura que de escritura: listar y descargar. Con RWMutex, N lecturas
// pasan a la vez y solo las escrituras se serializan.

import (
	"crypto/md5"
	"encoding/hex"
	"sort"
	"strings"
	"sync"
	"time"
)

// Objeto es un fichero guardado dentro de un bucket.
type Objeto struct {
	Clave      string
	Datos      []byte
	Tipo       string
	ETag       string // el md5 entre comillas, que es lo que S3 devuelve
	Tam        int    // el tamano va APARTE de Datos: en un listado no se devuelve contenido
	Modificado time.Time
}

// Bucket es un contenedor de objetos. En S3 el nombre es unico en toda la cuenta.
type Bucket struct {
	Nombre  string
	Creado  time.Time
	objetos map[string]*Objeto
}

// Almacen guarda todos los buckets de esta instancia.
type Almacen struct {
	mu      sync.RWMutex
	buckets map[string]*Bucket
	reloj   func() time.Time // inyectable: las pruebas necesitan un reloj que no se mueva
}

func NuevoAlmacen() *Almacen {
	return &Almacen{buckets: map[string]*Bucket{}, reloj: time.Now}
}

// CrearBucket devuelve false si ya existia. S3 distingue los dos casos con codigos de error
// distintos, asi que el que llama necesita saberlo.
func (a *Almacen) CrearBucket(nombre string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ya := a.buckets[nombre]; ya {
		return false
	}
	a.buckets[nombre] = &Bucket{Nombre: nombre, Creado: a.reloj().UTC(), objetos: map[string]*Objeto{}}
	return true
}

func (a *Almacen) HayBucket(nombre string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	_, ok := a.buckets[nombre]
	return ok
}

// BorrarBucket: S3 se niega a borrar un bucket que no esta vacio, y ese "no" es
// deliberadamente ruidoso. Aqui se replica.
func (a *Almacen) BorrarBucket(nombre string) (existia, vacio bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	b, ok := a.buckets[nombre]
	if !ok {
		return false, false
	}
	if len(b.objetos) > 0 {
		return true, false
	}
	delete(a.buckets, nombre)
	return true, true
}

// Buckets devuelve la lista ORDENADA por nombre. S3 los devuelve ordenados, y un mapa de Go
// no tiene orden: si no ordenas, la salida cambia en cada llamada y las pruebas parpadean.
func (a *Almacen) Buckets() []Bucket {
	a.mu.RLock()
	defer a.mu.RUnlock()
	fuera := make([]Bucket, 0, len(a.buckets))
	for _, b := range a.buckets {
		fuera = append(fuera, Bucket{Nombre: b.Nombre, Creado: b.Creado})
	}
	sort.Slice(fuera, func(i, j int) bool { return fuera[i].Nombre < fuera[j].Nombre })
	return fuera
}

// Poner guarda un objeto y devuelve su ETag. El ETag de S3 para una subida simple es el md5
// del contenido, entre comillas: no lo inventamos, lo calculamos igual que el de verdad.
func (a *Almacen) Poner(bucket, clave string, datos []byte, tipo string) (string, bool) {
	suma := md5.Sum(datos)
	etag := `"` + hex.EncodeToString(suma[:]) + `"`
	a.mu.Lock()
	defer a.mu.Unlock()
	b, ok := a.buckets[bucket]
	if !ok {
		return "", false
	}
	b.objetos[clave] = &Objeto{
		Clave: clave, Datos: datos, Tipo: tipo, ETag: etag, Tam: len(datos),
		Modificado: a.reloj().UTC(),
	}
	return etag, true
}

// Obtener devuelve una COPIA de los datos. Devolver el slice interno dejaria que quien llama
// lo modificara por debajo del candado: seria una carrera de datos que el -race caza y que en
// produccion se ve como bytes corrompidos sin explicacion.
func (a *Almacen) Obtener(bucket, clave string) (*Objeto, bool, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	b, ok := a.buckets[bucket]
	if !ok {
		return nil, false, false
	}
	o, ok := b.objetos[clave]
	if !ok {
		return nil, true, false
	}
	copia := *o
	copia.Datos = append([]byte(nil), o.Datos...)
	return &copia, true, true
}

func (a *Almacen) Borrar(bucket, clave string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	b, ok := a.buckets[bucket]
	if !ok {
		return false
	}
	delete(b.objetos, clave)
	return true
}

// Listar implementa lo que de verdad hace ListObjectsV2: filtrar por prefijo, agrupar por
// delimitador y cortar en maxClaves.
//
// El delimitador es la parte que nadie espera: S3 NO tiene carpetas. Lo que parece una carpeta
// es esto — cuando pides delimiter="/" , todas las claves que comparten el trozo hasta la
// siguiente barra se colapsan en un unico "prefijo comun". Por eso `aws s3 ls` parece un `ls`.
func (a *Almacen) Listar(bucket, prefijo, delimitador string, maxClaves int) ([]Objeto, []string, bool, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	b, ok := a.buckets[bucket]
	if !ok {
		return nil, nil, false, false
	}
	claves := make([]string, 0, len(b.objetos))
	for k := range b.objetos {
		if strings.HasPrefix(k, prefijo) {
			claves = append(claves, k)
		}
	}
	sort.Strings(claves)

	var objetos []Objeto
	var comunes []string
	vistos := map[string]bool{}
	truncado := false
	for _, k := range claves {
		if delimitador != "" {
			resto := k[len(prefijo):]
			if i := strings.Index(resto, delimitador); i >= 0 {
				grupo := prefijo + resto[:i+len(delimitador)]
				if !vistos[grupo] {
					if len(objetos)+len(comunes) >= maxClaves {
						truncado = true
						break
					}
					vistos[grupo] = true
					comunes = append(comunes, grupo)
				}
				continue
			}
		}
		if len(objetos)+len(comunes) >= maxClaves {
			truncado = true
			break
		}
		o := b.objetos[k]
		objetos = append(objetos, Objeto{
			Clave: o.Clave, ETag: o.ETag, Modificado: o.Modificado, Tipo: o.Tipo,
			Tam:   o.Tam,
			Datos: nil, // el listado no lleva contenido: seria devolver el bucket entero
		})
	}
	return objetos, comunes, truncado, true
}

// ⚠️ Aqui NO hay un metodo Tamano(bucket, clave) a proposito. Tendria que coger el RLock, y
// llamarlo desde dentro de Listar —que ya lo tiene— es la receta clasica del interbloqueo en Go:
// si un escritor esta esperando entre los dos RLock, el segundo se queda parado para siempre.
// Por eso el tamano viaja DENTRO del Objeto y no se vuelve a preguntar.
