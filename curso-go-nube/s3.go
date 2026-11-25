package main

// El servicio S3. Aqui esta TODO lo que hace falta para que el SDK de AWS y la CLI de AWS
// hablen con esto creyendo que es Amazon.
//
// Lo importante del episodio 1, y lo que casi nadie sabe: **la API de S3 no es un protocolo
// raro, es HTTP a secas**. El bucket es el primer trozo de la ruta, la clave del objeto es el
// resto, y las respuestas son XML. Ni gRPC, ni protobuf, ni una libreria de Amazon: net/http.
//
// Y el enrutado es el de la biblioteca estandar, sin framework. Desde Go 1.22 el ServeMux
// entiende el metodo y los comodines en el patron, que es exactamente lo que hace falta:
//
//	PUT /{bucket}            -> crear bucket
//	PUT /{bucket}/{clave...} -> subir objeto      ({clave...} se traga las barras
//	GET /{bucket}            -> listar             de "fotos/2026/gato.png")

import (
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// El espacio de nombres que S3 pone en todas sus respuestas. Si te lo saltas, algunos clientes
// se lo tragan y otros no: el SDK de Go si, la CLI tambien, pero no todos.
const nsS3 = "http://s3.amazonaws.com/doc/2006-03-01/"

// formatoISO es el que usa S3 en los listados: hasta milisegundos y con Z. Con el RFC3339 de
// Go a secas sale con nanosegundos y algunos analizadores se quejan.
const formatoISO = "2006-01-02T15:04:05.000Z"

type ServicioS3 struct {
	almacen *Almacen
	// notif puede ser nil: entonces S3 no avisa a nadie, que es como estuvo siete episodios.
	notif *Notificador
	// registro es opcional: cuando esta puesto, cada peticion se imprime. Es LA razon por la
	// que un emulador propio gana a uno de caja: ves el HTTP crudo que manda el SDK.
	registro func(formato string, args ...any)
}

func NuevoS3(a *Almacen, registro func(string, ...any)) *ServicioS3 {
	if registro == nil {
		registro = func(string, ...any) {}
	}
	return &ServicioS3{almacen: a, registro: registro}
}

// ConNotificador le da a S3 la capacidad de avisar cuando pasa algo (episodio 8).
func (s *ServicioS3) ConNotificador(n *Notificador) *ServicioS3 {
	s.notif = n
	return s
}

// avisa dispara la notificacion si hay a quien avisar. Se llama SIEMPRE despues de haber
// contestado al almacen, nunca con el candado en la mano.
func (s *ServicioS3) avisa(bucket, clave, etag string, tam int, evento string) {
	if s.notif == nil {
		return
	}
	s.notif.Aviso(bucket, clave, etag, tam, evento)
}

func (s *ServicioS3) Rutas(mux *http.ServeMux) {
	mux.HandleFunc("GET /", s.listarBuckets)
	// ⚠️ El mux de Go NO mira la cadena de consulta, y S3 mete ahi sus «subrecursos»:
	// `?notification`, `?acl`, `?versioning`... Asi que la misma ruta y el mismo metodo son
	// dos operaciones distintas, y hay que mirarlo a mano dentro del manejador.
	mux.HandleFunc("PUT /{bucket}", s.putBucket)
	mux.HandleFunc("DELETE /{bucket}", s.borrarBucket)
	mux.HandleFunc("HEAD /{bucket}", s.cabezaBucket)
	mux.HandleFunc("GET /{bucket}", s.listarObjetos)
	mux.HandleFunc("PUT /{bucket}/{clave...}", s.ponerObjeto)
	mux.HandleFunc("GET /{bucket}/{clave...}", s.obtenerObjeto)
	mux.HandleFunc("HEAD /{bucket}/{clave...}", s.obtenerObjeto)
	mux.HandleFunc("DELETE /{bucket}/{clave...}", s.borrarObjeto)
}

// ── los errores ──────────────────────────────────────────────────────────
// S3 no devuelve un 404 pelado: devuelve un XML con un Code que el SDK convierte en un tipo de
// error de Go (NoSuchBucket, NoSuchKey...). Si aqui contestas un 404 vacio, el SDK del usuario
// recibe "unexpected response" y no puede distinguir "no existe" de "se rompio la red".

type errorXML struct {
	XMLName  xml.Name `xml:"Error"`
	Code     string   `xml:"Code"`
	Message  string   `xml:"Message"`
	Resource string   `xml:"Resource,omitempty"`
}

func (s *ServicioS3) fallo(w http.ResponseWriter, r *http.Request, estado int, codigo, mensaje string) {
	s.registro("  -> %d %s", estado, codigo)
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(estado)
	_, _ = io.WriteString(w, xml.Header)
	_ = xml.NewEncoder(w).Encode(errorXML{Code: codigo, Message: mensaje, Resource: r.URL.Path})
}

func (s *ServicioS3) responderXML(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(estado)
	_, _ = io.WriteString(w, xml.Header)
	_ = xml.NewEncoder(w).Encode(v)
}

// ── buckets ──────────────────────────────────────────────────────────────

type bucketXML struct {
	Nombre string `xml:"Name"`
	Creado string `xml:"CreationDate"`
}

type listaBucketsXML struct {
	XMLName xml.Name `xml:"ListAllMyBucketsResult"`
	NS      string   `xml:"xmlns,attr"`
	Owner   struct {
		ID          string `xml:"ID"`
		DisplayName string `xml:"DisplayName"`
	} `xml:"Owner"`
	Buckets struct {
		Bucket []bucketXML `xml:"Bucket"`
	} `xml:"Buckets"`
}

func (s *ServicioS3) listarBuckets(w http.ResponseWriter, r *http.Request) {
	s.registro("%s %s", r.Method, r.URL.RequestURI())
	var res listaBucketsXML
	res.NS = nsS3
	res.Owner.ID = "nubecita"
	res.Owner.DisplayName = "nubecita"
	for _, b := range s.almacen.Buckets() {
		res.Buckets.Bucket = append(res.Buckets.Bucket, bucketXML{
			Nombre: b.Nombre, Creado: b.Creado.Format(formatoISO),
		})
	}
	s.registro("  -> 200 %d bucket(s)", len(res.Buckets.Bucket))
	s.responderXML(w, http.StatusOK, res)
}

// putBucket reparte entre «crear el bucket» y «configurar sus notificaciones», que en S3
// comparten metodo y ruta y solo se distinguen por la consulta.
func (s *ServicioS3) putBucket(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has("notification") {
		s.ponerNotificacion(w, r)
		return
	}
	s.crearBucket(w, r)
}

// ponerNotificacion guarda a quien hay que avisar cuando pasa algo en este bucket.
//
// ⚠️ Y fijate donde vive esto: en el BUCKET. Buscas en tu funcion quien la llama y no lo
// encuentras, porque esta escrito al otro lado.
func (s *ServicioS3) ponerNotificacion(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	s.registro("PUT /%s?notification", bucket)
	if s.notif == nil {
		s.fallo(w, r, http.StatusNotImplemented, "NotImplemented",
			"esta nubecita arranco sin disparadores")
		return
	}
	if !s.almacen.HayBucket(bucket) {
		s.fallo(w, r, http.StatusNotFound, "NoSuchBucket", "The specified bucket does not exist")
		return
	}
	var cfg notificacionXML
	if err := xml.NewDecoder(r.Body).Decode(&cfg); err != nil {
		s.fallo(w, r, http.StatusBadRequest, "MalformedXML", err.Error())
		return
	}
	s.notif.Poner(bucket, cfg)
	s.registro("  -> 200 %d destino(s) a Lambda, %d a colas", len(cfg.Lambdas), len(cfg.Colas))
	w.WriteHeader(http.StatusOK)
}

// obtenerNotificacion devuelve lo configurado. El SDK lo llama para comprobar, y sin esto
// `get-bucket-notification-configuration` da un error raro en vez de una lista vacia.
func (s *ServicioS3) obtenerNotificacion(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	s.registro("GET /%s?notification", bucket)
	cfg := notificacionXML{}
	if s.notif != nil {
		cfg = s.notif.Config(bucket)
	}
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(xml.Header))
	_ = xml.NewEncoder(w).Encode(cfg)
}

func (s *ServicioS3) crearBucket(w http.ResponseWriter, r *http.Request) {
	nombre := r.PathValue("bucket")
	s.registro("%s %s", r.Method, r.URL.RequestURI())
	// El cuerpo trae el LocationConstraint cuando la region no es us-east-1. Da igual el
	// contenido, pero hay que LEERLO Y CERRARLO o la conexion no se puede reutilizar.
	_, _ = io.Copy(io.Discard, r.Body)
	if !s.almacen.CrearBucket(nombre) {
		s.fallo(w, r, http.StatusConflict, "BucketAlreadyOwnedByYou",
			"Your previous request to create the named bucket succeeded and you already own it.")
		return
	}
	s.registro("  -> 200 creado")
	w.Header().Set("Location", "/"+nombre)
	w.WriteHeader(http.StatusOK)
}

func (s *ServicioS3) borrarBucket(w http.ResponseWriter, r *http.Request) {
	nombre := r.PathValue("bucket")
	s.registro("%s %s", r.Method, r.URL.RequestURI())
	existia, vacio := s.almacen.BorrarBucket(nombre)
	if !existia {
		s.fallo(w, r, http.StatusNotFound, "NoSuchBucket", "The specified bucket does not exist")
		return
	}
	if !vacio {
		s.fallo(w, r, http.StatusConflict, "BucketNotEmpty",
			"The bucket you tried to delete is not empty")
		return
	}
	s.registro("  -> 204 borrado")
	w.WriteHeader(http.StatusNoContent)
}

func (s *ServicioS3) cabezaBucket(w http.ResponseWriter, r *http.Request) {
	nombre := r.PathValue("bucket")
	s.registro("%s %s", r.Method, r.URL.RequestURI())
	if !s.almacen.HayBucket(nombre) {
		// ⚠️ HeadBucket NO puede llevar cuerpo: es un HEAD. El codigo es toda la respuesta.
		s.registro("  -> 404")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	s.registro("  -> 200")
	w.WriteHeader(http.StatusOK)
}

// ── listado de objetos ───────────────────────────────────────────────────

type contenidoXML struct {
	Clave        string `xml:"Key"`
	Modificado   string `xml:"LastModified"`
	ETag         string `xml:"ETag"`
	Tam          int    `xml:"Size"`
	StorageClass string `xml:"StorageClass"`
}

type prefijoComunXML struct {
	Prefijo string `xml:"Prefix"`
}

type listaObjetosXML struct {
	XMLName        xml.Name          `xml:"ListBucketResult"`
	NS             string            `xml:"xmlns,attr"`
	Nombre         string            `xml:"Name"`
	Prefijo        string            `xml:"Prefix"`
	Delimitador    string            `xml:"Delimiter,omitempty"`
	MaxClaves      int               `xml:"MaxKeys"`
	NumClaves      int               `xml:"KeyCount"`
	Truncado       bool              `xml:"IsTruncated"`
	TipoCodificado string            `xml:"EncodingType,omitempty"`
	Contenido      []contenidoXML    `xml:"Contents"`
	PrefijosComun  []prefijoComunXML `xml:"CommonPrefixes"`
}

func (s *ServicioS3) listarObjetos(w http.ResponseWriter, r *http.Request) {
	// Mismo caso que en el PUT: `?notification` es otra operacion con la misma ruta.
	if r.URL.Query().Has("notification") {
		s.obtenerNotificacion(w, r)
		return
	}
	bucket := r.PathValue("bucket")
	q := r.URL.Query()
	s.registro("%s %s", r.Method, r.URL.RequestURI())

	prefijo, delim := q.Get("prefix"), q.Get("delimiter")
	maxClaves := 1000
	if v := q.Get("max-keys"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			maxClaves = n
		}
	}
	objetos, comunes, truncado, hay := s.almacen.Listar(bucket, prefijo, delim, maxClaves)
	if !hay {
		s.fallo(w, r, http.StatusNotFound, "NoSuchBucket", "The specified bucket does not exist")
		return
	}

	// encoding-type=url: la CLI SIEMPRE lo pide. Si lo ignoras, una clave con un espacio o un
	// acento vuelve mal y el cliente muestra basura. Es un detalle de dos lineas que separa
	// "casi funciona" de "funciona".
	codificar := strings.EqualFold(q.Get("encoding-type"), "url")
	esc := func(v string) string {
		if codificar {
			return url.QueryEscape(v)
		}
		return v
	}

	res := listaObjetosXML{
		NS: nsS3, Nombre: bucket, Prefijo: esc(prefijo), Delimitador: esc(delim),
		MaxClaves: maxClaves, Truncado: truncado,
	}
	if codificar {
		res.TipoCodificado = "url"
	}
	for _, o := range objetos {
		res.Contenido = append(res.Contenido, contenidoXML{
			Clave: esc(o.Clave), Modificado: o.Modificado.Format(formatoISO),
			ETag: o.ETag, Tam: o.Tam, StorageClass: "STANDARD",
		})
	}
	for _, p := range comunes {
		res.PrefijosComun = append(res.PrefijosComun, prefijoComunXML{Prefijo: esc(p)})
	}
	res.NumClaves = len(res.Contenido) + len(res.PrefijosComun)
	s.registro("  -> 200 %d objeto(s), %d prefijo(s)", len(res.Contenido), len(res.PrefijosComun))
	s.responderXML(w, http.StatusOK, res)
}

// ── objetos ──────────────────────────────────────────────────────────────

func (s *ServicioS3) ponerObjeto(w http.ResponseWriter, r *http.Request) {
	bucket, clave := r.PathValue("bucket"), r.PathValue("clave")
	s.registro("%s %s", r.Method, r.URL.RequestURI())
	datos, err := io.ReadAll(r.Body)
	if err != nil {
		s.fallo(w, r, http.StatusBadRequest, "IncompleteBody", err.Error())
		return
	}
	tipo := r.Header.Get("Content-Type")
	if tipo == "" {
		tipo = "binary/octet-stream"
	}
	etag, hay := s.almacen.Poner(bucket, clave, datos, tipo)
	if !hay {
		s.fallo(w, r, http.StatusNotFound, "NoSuchBucket", "The specified bucket does not exist")
		return
	}
	s.registro("  -> 200 %d byte(s) etag=%s", len(datos), etag)
	w.Header().Set("ETag", etag)
	w.WriteHeader(http.StatusOK)
	// El aviso sale DESPUES de contestar. Tu PutObject ya ha terminado bien pase lo que pase
	// con la notificacion: son dos cosas distintas.
	s.avisa(bucket, clave, etag, len(datos), "s3:ObjectCreated:Put")
}

func (s *ServicioS3) obtenerObjeto(w http.ResponseWriter, r *http.Request) {
	bucket, clave := r.PathValue("bucket"), r.PathValue("clave")
	s.registro("%s %s", r.Method, r.URL.RequestURI())
	o, hayBucket, hayObjeto := s.almacen.Obtener(bucket, clave)
	if !hayBucket {
		s.fallo(w, r, http.StatusNotFound, "NoSuchBucket", "The specified bucket does not exist")
		return
	}
	if !hayObjeto {
		s.fallo(w, r, http.StatusNotFound, "NoSuchKey", "The specified key does not exist.")
		return
	}
	w.Header().Set("Content-Type", o.Tipo)
	w.Header().Set("Content-Length", strconv.Itoa(len(o.Datos)))
	w.Header().Set("ETag", o.ETag)
	w.Header().Set("Last-Modified", o.Modificado.Format(http.TimeFormat))
	w.Header().Set("Accept-Ranges", "bytes")
	w.WriteHeader(http.StatusOK)
	// En un HEAD, net/http tira el cuerpo el solo. No hace falta ramificar.
	s.registro("  -> 200 %d byte(s)", len(o.Datos))
	_, _ = w.Write(o.Datos)
}

func (s *ServicioS3) borrarObjeto(w http.ResponseWriter, r *http.Request) {
	bucket, clave := r.PathValue("bucket"), r.PathValue("clave")
	s.registro("%s %s", r.Method, r.URL.RequestURI())
	if !s.almacen.Borrar(bucket, clave) {
		s.fallo(w, r, http.StatusNotFound, "NoSuchBucket", "The specified bucket does not exist")
		return
	}
	// ⚠️ S3 devuelve 204 aunque la clave NO existiera. Borrar es idempotente a proposito: si
	// devolvieras 404, un reintento de red fallaria despues de haber borrado bien.
	s.registro("  -> 204")
	w.WriteHeader(http.StatusNoContent)
	s.avisa(bucket, clave, "", 0, "s3:ObjectRemoved:Delete")
}
