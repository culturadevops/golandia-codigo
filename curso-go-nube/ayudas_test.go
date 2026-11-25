package main

// Ayudas de las pruebas de firma: clientes de HTTP que se meten EN MEDIO y toquetean la
// petición DESPUÉS de que el SDK la haya firmado. Es la forma de representar a alguien que
// intercepta el tráfico, y lo que demuestra para qué sirve firmar.

import (
	"bytes"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// servidorCon levanta una nubecita entera sobre unos almacenes que le das tú. Lo necesitan las
// pruebas de persistencia: hay que poder llenar unos, guardar, y levantar OTRA nubecita sobre
// los almacenes cargados del fichero.
func servidorCon(t *testing.T, a *Almacen, ts *Tablas, cs *Colas, f *Funciones) string {
	t.Helper()
	mux := http.NewServeMux()
	NuevoS3(a, nil).Rutas(mux)
	lam := NuevoLambda(f, nil)
	rutas := lam.Rutas()
	porRuta := map[string]http.Handler{}
	for _, p := range lam.Prefijos() {
		porRuta[p] = rutas
	}
	srv := httptest.NewServer(Raiz(mux, map[string]http.Handler{
		"DynamoDB_":  NuevoDynamo(ts, nil),
		"AmazonSQS.": NuevoSQS(cs, nil),
	}, porRuta, nil))
	t.Cleanup(srv.Close)
	return srv.URL
}

func clienteSQS(url string) *sqs.Client {
	return sqs.New(sqs.Options{
		Region:       "us-east-1",
		Credentials:  credentials.NewStaticCredentialsProvider("prueba", "prueba", ""),
		BaseEndpoint: aws.String(url),
	})
}

func credencialesFijas(acceso, secreto string) aws.CredentialsProvider {
	return credentials.NewStaticCredentialsProvider(acceso, secreto, "")
}

func hexDe(b []byte) string { return hex.EncodeToString(b) }

// clienteQueToquetea cambia el CUERPO después de la firma, dejando las cabeceras intactas.
type clienteQueToquetea struct{}

func (c *clienteQueToquetea) Do(r *http.Request) (*http.Response, error) {
	nuevo := []byte("transferir 10000 euros")
	r.Body = io.NopCloser(bytes.NewReader(nuevo))
	r.ContentLength = int64(len(nuevo))
	return http.DefaultClient.Do(r)
}

// clienteQueCambiaLaRuta manda la petición firmada a OTRO sitio.
type clienteQueCambiaLaRuta struct{}

func (c *clienteQueCambiaLaRuta) Do(r *http.Request) (*http.Response, error) {
	r.URL.Path = "/otro-bucket"
	return http.DefaultClient.Do(r)
}
