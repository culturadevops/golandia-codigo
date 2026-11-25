package main

// Las pruebas de la firma. Y aquí hay una gracia: quien FIRMA es el SDK oficial de AWS, y quien
// COMPRUEBA es nuestro código. O sea que si estas pruebas pasan, nuestra comprobación entiende
// exactamente lo que produce Amazon — que es la única forma honesta de decirlo.
//
//	go test ./... -count=1 -v -run Firma

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithy "github.com/aws/smithy-go"
)

// nubeFirmada levanta la nubecita CON comprobación de firma, y devuelve además el verificador
// para poder moverle el reloj (y así probar que una firma caduca sin esperar quince minutos).
func nubeFirmada(t *testing.T, acceso, secreto string) (*s3.Client, *Verificador, *httptest.Server) {
	t.Helper()
	mux := http.NewServeMux()
	NuevoS3(NuevoAlmacen(), nil).Rutas(mux)
	v := NuevoVerificador(map[string]string{acceso: secreto}, nil)
	srv := httptest.NewServer(v.Envolver(Raiz(mux, map[string]http.Handler{
		"DynamoDB_":  NuevoDynamo(NuevasTablas(), nil),
		"AmazonSQS.": NuevoSQS(NuevasColas(), nil),
	}, nil, nil)))
	t.Cleanup(srv.Close)
	return clienteS3(srv.URL), v, srv
}

func codigoDeError(err error) string {
	var api smithy.APIError
	if errors.As(err, &api) {
		return api.ErrorCode()
	}
	return ""
}

// 🥇 La prueba de fondo: el SDK oficial firma, nosotros comprobamos, y encaja.
func TestFirmaDelSDKOficialEncaja(t *testing.T) {
	c, _, _ := nubeFirmada(t, "prueba", "prueba")
	if _, err := c.CreateBucket(context.Background(), &s3.CreateBucketInput{
		Bucket: aws.String("fotos")}); err != nil {
		t.Fatalf("el SDK firmó y nuestra comprobación la rechazó: %v", err)
	}
}

// Con el secreto equivocado, la firma sale distinta. Y fíjate en que el secreto NUNCA viaja:
// lo que no coincide es el resultado de una cadena de HMAC.
func TestConElSecretoEquivocadoNoPasa(t *testing.T) {
	_, _, srv := nubeFirmada(t, "prueba", "el-de-verdad")
	malo := s3.New(s3.Options{
		Region:       "us-east-1",
		Credentials:  credencialesFijas("prueba", "otro-secreto"),
		BaseEndpoint: aws.String(srv.URL),
		UsePathStyle: true,
	})
	_, err := malo.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err == nil {
		t.Fatal("aceptó una firma hecha con otro secreto")
	}
	if c := codigoDeError(err); c != "SignatureDoesNotMatch" {
		t.Fatalf("el error llegó como %q y esperaba SignatureDoesNotMatch: %v", c, err)
	}
}

// Una clave de acceso que no conocemos se rechaza antes de calcular nada.
func TestClaveDeAccesoDesconocida(t *testing.T) {
	_, _, srv := nubeFirmada(t, "prueba", "prueba")
	otro := s3.New(s3.Options{
		Region:       "us-east-1",
		Credentials:  credencialesFijas("no-te-conozco", "prueba"),
		BaseEndpoint: aws.String(srv.URL),
		UsePathStyle: true,
	})
	_, err := otro.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err == nil || !strings.Contains(err.Error(), "no conozco la clave de acceso") {
		t.Fatalf("esperaba que dijera que no conoce la clave: %v", err)
	}
}

// 🥇 LA PRUEBA QUE EXPLICA PARA QUÉ SIRVE FIRMAR: alguien en el medio cambia el cuerpo DESPUÉS
// de que el SDK lo firmara. El cuerpo viaja en claro y se puede leer — firmar no es cifrar —
// pero en cuanto se toca, la firma deja de cuadrar.
func TestSiAlguienToquetealCuerpoLaFirmaSeCae(t *testing.T) {
	_, _, srv := nubeFirmada(t, "prueba", "prueba")
	entrometido := s3.New(s3.Options{
		Region:       "us-east-1",
		Credentials:  credencialesFijas("prueba", "prueba"),
		BaseEndpoint: aws.String(srv.URL),
		UsePathStyle: true,
		HTTPClient:   &clienteQueToquetea{},
	})
	_, err := entrometido.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String("fotos"), Key: aws.String("gato.txt"),
		Body: bytes.NewReader([]byte("transferir 10 euros")),
	})
	if err == nil {
		t.Fatal("aceptó una petición cuyo cuerpo cambió después de firmarse")
	}
	if c := codigoDeError(err); c != "SignatureDoesNotMatch" {
		t.Fatalf("el error llegó como %q: %v", c, err)
	}
}

// Y lo mismo con la ruta: cambiar el destino después de firmar tampoco cuela.
func TestSiAlguienCambiaLaRutaLaFirmaSeCae(t *testing.T) {
	_, _, srv := nubeFirmada(t, "prueba", "prueba")
	entrometido := s3.New(s3.Options{
		Region:       "us-east-1",
		Credentials:  credencialesFijas("prueba", "prueba"),
		BaseEndpoint: aws.String(srv.URL),
		UsePathStyle: true,
		HTTPClient:   &clienteQueCambiaLaRuta{},
	})
	_, err := entrometido.CreateBucket(context.Background(), &s3.CreateBucketInput{
		Bucket: aws.String("mio")})
	if err == nil {
		t.Fatal("aceptó una petición cuya ruta cambió después de firmarse")
	}
}

// 🥇 Una firma CADUCA. Sin esto, quien capture una sola petición firmada podría repetirla para
// siempre. Y se prueba sin esperar: el reloj del verificador es inyectable.
func TestUnaFirmaCaduca(t *testing.T) {
	c, v, _ := nubeFirmada(t, "prueba", "prueba")
	// El servidor cree que han pasado dos horas desde que se firmó.
	v.ahora = func() time.Time { return time.Now().UTC().Add(2 * time.Hour) }

	_, err := c.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err == nil {
		t.Fatal("aceptó una firma de hace dos horas")
	}
	if !strings.Contains(err.Error(), "caducó") {
		t.Fatalf("el error no dice que caducó: %v", err)
	}
}

// Sin cabecera de autorización, ni se mira qué pedías.
func TestSinFirmaNiSeMira(t *testing.T) {
	_, _, srv := nubeFirmada(t, "prueba", "prueba")
	res, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("código = %d, quiero 403", res.StatusCode)
	}
}

// Y sin claves configuradas, el verificador se aparta del camino: es el modo de los episodios
// 1 a 3, y tiene que seguir funcionando igual.
func TestSinClavesElVerificadorNoEstorba(t *testing.T) {
	mux := http.NewServeMux()
	NuevoS3(NuevoAlmacen(), nil).Rutas(mux)
	v := NuevoVerificador(nil, nil)
	srv := httptest.NewServer(v.Envolver(mux))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("código = %d, quiero 200: el verificador estorba cuando no debería",
			res.StatusCode)
	}
}

// La cadena de HMAC, comprobada contra el ejemplo que AWS publica en su documentación de SigV4.
// Es el único número de todo el proyecto que NO sale de esta máquina: sale de Amazon, y por eso
// sirve para saber si nuestra implementación es la de verdad o solo se parece.
func TestLaClaveDeFirmaCoincideConElEjemploDeAWS(t *testing.T) {
	// Ejemplo oficial: secreto wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY,
	// fecha 20150830, región us-east-1, servicio iam.
	k := claveDeFirma("wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY", "20150830", "us-east-1", "iam")
	quiero := "c4afb1cc5771d871763a393e44b703571b55cc28424d1a5e86da6ed3c154a4b9"
	if got := hexDe(k); got != quiero {
		t.Fatalf("clave de firma = %s\nquiero          = %s", got, quiero)
	}
}
