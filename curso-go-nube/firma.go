package main

// 🥇 SigV4: qué significa de verdad FIRMAR una petición.
//
// Casi todo el mundo usa esto todos los días sin haberlo mirado nunca por dentro, porque el SDK
// lo hace solo. Y cuando falla, el mensaje es siempre el mismo y nunca dice por qué:
//
//	SignatureDoesNotMatch
//
// Lo primero, y lo que más cuesta creer: **firmar no es cifrar**. La petición viaja tal cual, en
// claro, con su cuerpo legible. Lo que la firma demuestra son otras dos cosas:
//
//	1. QUIÉN la manda, sin que la contraseña viaje nunca por el cable;
//	2. que NADIE la ha tocado por el camino — ni el cuerpo, ni las cabeceras firmadas, ni la ruta.
//
// ## Cómo funciona, en cuatro pasos
//
//	1. PETICIÓN CANÓNICA   se reescribe la petición en un formato exacto y sin ambigüedades
//	2. CADENA A FIRMAR     la fecha, el ámbito, y el resumen de esa petición canónica
//	3. CLAVE DE FIRMA      una cadena de HMAC: secreto -> fecha -> región -> servicio -> aws4
//	4. FIRMA               HMAC de la cadena a firmar con esa clave
//
// El servidor rehace los cuatro pasos con SU copia del secreto y compara. Si sale lo mismo, la
// petición es auténtica. Si no, `SignatureDoesNotMatch` — y el motivo suele estar en el paso 1.
//
// ⚠️ POR QUÉ EL PASO 1 ES EL QUE DUELE: para que salga la misma firma, el servidor tiene que
// reconstruir la petición EXACTAMENTE igual que el cliente. Una barra de más en la ruta, un
// parámetro de consulta en otro orden, una cabecera con un espacio de más… y la firma cambia
// entera. El error no dice «te sobra una barra»: dice que la firma no coincide.

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	algoritmo   = "AWS4-HMAC-SHA256"
	formatoAmz  = "20060102T150405Z"
	formatoDia  = "20060102"
	sinFirmar   = "UNSIGNED-PAYLOAD"
	ventanaVida = 15 * time.Minute // lo mismo que usa AWS: una firma vieja no vale
)

// Verificador comprueba la firma de cada petición. Con `claves` vacío no comprueba nada, que es
// el modo de los episodios 1 a 3: un doble de pruebas no necesita autenticar a nadie.
type Verificador struct {
	claves   map[string]string // identificador de acceso -> secreto
	registro func(string, ...any)
	ahora    func() time.Time // inyectable: sin esto no se puede probar que una firma CADUCA
}

func NuevoVerificador(claves map[string]string, registro func(string, ...any)) *Verificador {
	if registro == nil {
		registro = func(string, ...any) {}
	}
	return &Verificador{claves: claves, registro: registro, ahora: time.Now}
}

// Envolver mete la comprobación delante de lo que ya había. Si no hay claves configuradas, se
// aparta del camino: devuelve el mismo manejador.
func (v *Verificador) Envolver(siguiente http.Handler) http.Handler {
	if len(v.claves) == 0 {
		return siguiente
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ⚠️ Hay que leer el cuerpo para poder resumirlo... y volver a dejarlo puesto, o el
		// servicio de detrás se encuentra un cuerpo vacío. Es el fallo clásico de escribir un
		// middleware que mira el cuerpo.
		cuerpo, err := io.ReadAll(r.Body)
		if err != nil {
			v.rechazar(w, "IncompleteBody", err.Error())
			return
		}
		r.Body = io.NopCloser(strings.NewReader(string(cuerpo)))

		if motivo := v.Comprobar(r, cuerpo); motivo != "" {
			v.registro("  -> 403 SignatureDoesNotMatch (%s)", motivo)
			v.rechazar(w, "SignatureDoesNotMatch",
				"The request signature we calculated does not match the signature you provided. "+
					"nubecita dice por qué: "+motivo)
			return
		}
		siguiente.ServeHTTP(w, r)
	})
}

func (v *Verificador) rechazar(w http.ResponseWriter, codigo, mensaje string) {
	// Se contesta en XML porque el que más se usa aquí es S3. Los servicios de estilo JSON
	// aceptan igualmente el 403; lo que importa es el código.
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusForbidden)
	_, _ = io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>`+
		"<Error><Code>"+codigo+"</Code><Message>"+mensaje+"</Message></Error>")
}

// Comprobar devuelve "" si la firma es buena, y si no, EL MOTIVO. Eso es lo único que esta
// implementación hace mejor que la de verdad: decir por qué.
func (v *Verificador) Comprobar(r *http.Request, cuerpo []byte) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "no viene la cabecera Authorization"
	}
	if !strings.HasPrefix(auth, algoritmo+" ") {
		return "el algoritmo no es " + algoritmo
	}
	cred, firmadas, firma := "", "", ""
	for _, trozo := range strings.Split(strings.TrimPrefix(auth, algoritmo+" "), ",") {
		trozo = strings.TrimSpace(trozo)
		switch {
		case strings.HasPrefix(trozo, "Credential="):
			cred = strings.TrimPrefix(trozo, "Credential=")
		case strings.HasPrefix(trozo, "SignedHeaders="):
			firmadas = strings.TrimPrefix(trozo, "SignedHeaders=")
		case strings.HasPrefix(trozo, "Signature="):
			firma = strings.TrimPrefix(trozo, "Signature=")
		}
	}
	if cred == "" || firmadas == "" || firma == "" {
		return "la cabecera Authorization está incompleta"
	}
	partes := strings.Split(cred, "/")
	if len(partes) != 5 {
		return "la credencial no tiene los cinco tramos: clave/fecha/region/servicio/aws4_request"
	}
	acceso, dia, region, servicio := partes[0], partes[1], partes[2], partes[3]
	secreto, hay := v.claves[acceso]
	if !hay {
		return "no conozco la clave de acceso " + acceso
	}

	// La fecha va en una cabecera propia, y hay que usar ESA, no la del reloj del servidor.
	fechaAmz := r.Header.Get("X-Amz-Date")
	if fechaAmz == "" {
		fechaAmz = r.Header.Get("Date")
	}
	t, err := time.Parse(formatoAmz, fechaAmz)
	if err != nil {
		return "la cabecera X-Amz-Date no se entiende: " + fechaAmz
	}
	// 🥇 La ventana de validez: una firma caduca. Sin esto, quien capture UNA petición firmada
	// puede repetirla para siempre. Es lo que impide el ataque de repetición.
	if d := v.ahora().UTC().Sub(t); d > ventanaVida || d < -ventanaVida {
		return fmt.Sprintf("la firma es de hace %s: caducó (la ventana es de %s)",
			d.Round(time.Second), ventanaVida)
	}
	if t.UTC().Format(formatoDia) != dia {
		return "el día de la credencial no coincide con X-Amz-Date"
	}

	canonica, motivo := peticionCanonica(r, cuerpo, firmadas)
	if motivo != "" {
		return motivo
	}
	ambito := dia + "/" + region + "/" + servicio + "/aws4_request"
	aFirmar := strings.Join([]string{
		algoritmo, fechaAmz, ambito, resumen([]byte(canonica)),
	}, "\n")

	mia := hex.EncodeToString(hmacSHA256(claveDeFirma(secreto, dia, region, servicio),
		[]byte(aFirmar)))
	// ⚠️ hmac.Equal y NO `==`. Comparar cadenas con == sale antes en cuanto encuentra una
	// diferencia, y ese "antes" se puede MEDIR: es un canal por tiempos que deja adivinar la
	// firma byte a byte. Aquí no importa mucho; en un servidor de verdad, sí.
	if !hmac.Equal([]byte(mia), []byte(firma)) {
		return "la firma no coincide (revisa la petición canónica: ruta, consulta o cabeceras)"
	}

	// 🥇 Y falta un paso que casi nadie menciona, y sin el cual todo lo anterior no protege el
	// cuerpo: la firma cubre el RESUMEN del cuerpo, no el cuerpo. Si alguien cambia el cuerpo y
	// deja la cabecera del resumen como estaba, la firma SIGUE CUADRANDO. Hay que comprobar
	// además que el cuerpo hace ese resumen.
	if h := r.Header.Get("X-Amz-Content-Sha256"); h != "" && !especial(h) {
		if resumen(cuerpo) != h {
			return "el cuerpo no coincide con x-amz-content-sha256: alguien lo tocó " +
				"después de firmar"
		}
	}
	return ""
}

// especial dice si el valor de x-amz-content-sha256 no es un resumen, sino una palabra clave.
//
// ⚠️ `UNSIGNED-PAYLOAD` significa literalmente «no firmo el cuerpo». S3 lo admite cuando vas por
// TLS, porque entonces confías en el cifrado del canal para que nadie lo toque. Ahí la firma
// autentica QUIÉN eres, pero NO protege el contenido: eso lo hace el TLS.
func especial(h string) bool {
	return h == sinFirmar || strings.HasPrefix(h, "STREAMING-")
}

// peticionCanonica reescribe la petición en el formato exacto que manda SigV4. Es el paso 1, y
// el que se lleva el 90 % de los dolores de cabeza.
func peticionCanonica(r *http.Request, cuerpo []byte, firmadas string) (string, string) {
	ruta := r.URL.EscapedPath()
	if ruta == "" {
		ruta = "/"
	}

	// La consulta va ORDENADA por nombre y con cada trozo codificado. Si el cliente manda
	// `b=2&a=1`, aquí tiene que quedar `a=1&b=2`, o la firma no sale.
	q := r.URL.Query()
	nombres := make([]string, 0, len(q))
	for k := range q {
		nombres = append(nombres, k)
	}
	sort.Strings(nombres)
	var trozos []string
	for _, k := range nombres {
		vs := append([]string(nil), q[k]...)
		sort.Strings(vs)
		for _, val := range vs {
			trozos = append(trozos, url.QueryEscape(k)+"="+url.QueryEscape(val))
		}
	}
	consulta := strings.Join(trozos, "&")

	// Solo las cabeceras que el cliente DECLARÓ firmadas entran aquí. Por eso un proxy puede
	// añadir cabeceras por el camino sin romper la firma: si no están en la lista, no cuentan.
	lista := strings.Split(firmadas, ";")
	sort.Strings(lista)
	var cab strings.Builder
	for _, nombre := range lista {
		valor := r.Header.Get(nombre)
		if nombre == "host" {
			valor = r.Host // el Host NO está en r.Header: Go lo saca aparte
		}
		if valor == "" && nombre != "host" {
			return "", "la cabecera firmada " + nombre + " no viene en la petición"
		}
		cab.WriteString(strings.ToLower(nombre) + ":" + strings.TrimSpace(valor) + "\n")
	}

	// El resumen del cuerpo. En S3 viene en una cabecera —y puede decir literalmente «sin
	// firmar»—; en los servicios de estilo JSON se calcula.
	hash := r.Header.Get("X-Amz-Content-Sha256")
	if hash == "" {
		hash = resumen(cuerpo)
	}

	return strings.Join([]string{
		r.Method, ruta, consulta, cab.String(), firmadas, hash,
	}, "\n"), ""
}

// claveDeFirma es la cadena de HMAC que hace que el secreto NO viaje nunca, ni siquiera
// derivado de una sola vez: cada eslabón acota para qué sirve la clave resultante.
//
//	AWS4+secreto -> fecha -> región -> servicio -> aws4_request
//
// Consecuencia práctica: una firma de hoy no vale mañana, y una de S3 no vale para DynamoDB.
// Está acotada por diseño.
func claveDeFirma(secreto, dia, region, servicio string) []byte {
	k := hmacSHA256([]byte("AWS4"+secreto), []byte(dia))
	k = hmacSHA256(k, []byte(region))
	k = hmacSHA256(k, []byte(servicio))
	return hmacSHA256(k, []byte("aws4_request"))
}

func hmacSHA256(clave, datos []byte) []byte {
	m := hmac.New(sha256.New, clave)
	m.Write(datos)
	return m.Sum(nil)
}

func resumen(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}
