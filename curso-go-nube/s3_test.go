package main

// Estas pruebas NO hablan con nuestro servidor a mano: usan el SDK de AWS para Go de verdad,
// el mismo `github.com/aws/aws-sdk-go-v2/service/s3` que usarias contra Amazon. Es la unica
// forma honesta de decir "esto es compatible": si el SDK lo entiende, es compatible.
//
//	go test ./... -race -v

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// nube levanta una nubecita en un puerto libre y devuelve un cliente de S3 apuntado a ella.
//
// Las DOS lineas que hacen que el SDK hable con nosotros y no con Amazon:
//
//	o.BaseEndpoint = aws.String(srv.URL)   // ve aqui en vez de a s3.amazonaws.com
//	o.UsePathStyle = true                  // /bucket/clave, no bucket.s3.amazonaws.com/clave
//
// Lo segundo no es un capricho: el estilo por defecto mete el bucket en el NOMBRE DEL HOST, y
// contra localhost eso no resuelve. Es el error numero uno al apuntar el SDK a un emulador.
func nube(t *testing.T) *s3.Client {
	t.Helper()
	mux := http.NewServeMux()
	NuevoS3(NuevoAlmacen(), nil).Rutas(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return clienteS3(srv.URL)
}

// clienteS3 esta separado porque lo reutiliza la prueba de que S3 y DynamoDB conviven en el
// mismo puerto (`dynamo_test.go`).
func clienteS3(url string) *s3.Client {
	return s3.New(s3.Options{
		Region:       "us-east-1",
		Credentials:  credentials.NewStaticCredentialsProvider("prueba", "prueba", ""),
		BaseEndpoint: aws.String(url),
		UsePathStyle: true,
	})
}

func TestCicloCompleto(t *testing.T) {
	ctx := context.Background()
	c := nube(t)

	if _, err := c.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String("fotos")}); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}

	cuerpo := []byte("hola nubecita")
	sub, err := c.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String("fotos"), Key: aws.String("gato.txt"),
		Body: bytes.NewReader(cuerpo), ContentType: aws.String("text/plain"),
	})
	if err != nil {
		t.Fatalf("PutObject: %v", err)
	}

	// El ETag tiene que ser el md5 del contenido, entre comillas, igual que en S3 de verdad.
	suma := md5.Sum(cuerpo)
	quiero := `"` + hex.EncodeToString(suma[:]) + `"`
	if aws.ToString(sub.ETag) != quiero {
		t.Errorf("ETag = %q, quiero %q", aws.ToString(sub.ETag), quiero)
	}

	baja, err := c.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String("fotos"), Key: aws.String("gato.txt"),
	})
	if err != nil {
		t.Fatalf("GetObject: %v", err)
	}
	defer baja.Body.Close()
	leido, _ := io.ReadAll(baja.Body)
	if !bytes.Equal(leido, cuerpo) {
		t.Errorf("bajo %q, subi %q", leido, cuerpo)
	}
	if aws.ToString(baja.ContentType) != "text/plain" {
		t.Errorf("ContentType = %q", aws.ToString(baja.ContentType))
	}
	if aws.ToInt64(baja.ContentLength) != int64(len(cuerpo)) {
		t.Errorf("ContentLength = %d, quiero %d", aws.ToInt64(baja.ContentLength), len(cuerpo))
	}
}

// El error tiene que llegar al SDK como el TIPO correcto, no como un 404 pelado. Si esto
// falla, el codigo del usuario no puede distinguir "no existe" de "se cayo la red".
func TestClaveQueNoExisteLlegaComoNoSuchKey(t *testing.T) {
	ctx := context.Background()
	c := nube(t)
	if _, err := c.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String("vacio")}); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}

	_, err := c.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String("vacio"), Key: aws.String("no-esta"),
	})
	if err == nil {
		t.Fatal("esperaba un error y no hubo ninguno")
	}
	var noEsta *types.NoSuchKey
	if !errors.As(err, &noEsta) {
		t.Fatalf("el SDK no lo vio como NoSuchKey: %v", err)
	}
}

func TestBucketQueNoExisteLlegaComoNoSuchBucket(t *testing.T) {
	ctx := context.Background()
	c := nube(t)
	_, err := c.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: aws.String("fantasma")})
	var noEsta *types.NoSuchBucket
	if !errors.As(err, &noEsta) {
		t.Fatalf("el SDK no lo vio como NoSuchBucket: %v", err)
	}
}

// El delimitador es lo que hace que S3 PAREZCA que tiene carpetas. No las tiene.
func TestListarConDelimitadorAgrupaComoSiHubieraCarpetas(t *testing.T) {
	ctx := context.Background()
	c := nube(t)
	if _, err := c.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String("archivo")}); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}
	for _, k := range []string{"2025/enero.txt", "2025/febrero.txt", "2026/enero.txt", "raiz.txt"} {
		if _, err := c.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String("archivo"), Key: aws.String(k), Body: bytes.NewReader([]byte("x")),
		}); err != nil {
			t.Fatalf("PutObject %s: %v", k, err)
		}
	}

	res, err := c.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String("archivo"), Delimiter: aws.String("/"),
	})
	if err != nil {
		t.Fatalf("ListObjectsV2: %v", err)
	}
	// Cuatro objetos, pero con delimitador se ven DOS "carpetas" y UN fichero suelto.
	if len(res.CommonPrefixes) != 2 {
		t.Errorf("prefijos comunes = %d, quiero 2", len(res.CommonPrefixes))
	}
	if len(res.Contents) != 1 {
		t.Errorf("objetos en la raiz = %d, quiero 1", len(res.Contents))
	}
	if aws.ToInt32(res.KeyCount) != 3 {
		t.Errorf("KeyCount = %d, quiero 3", aws.ToInt32(res.KeyCount))
	}

	// Y sin delimitador, los cuatro planos: no hay carpetas, hay claves con barras dentro.
	todo, err := c.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: aws.String("archivo")})
	if err != nil {
		t.Fatalf("ListObjectsV2 sin delimitador: %v", err)
	}
	if len(todo.Contents) != 4 {
		t.Errorf("sin delimitador = %d objetos, quiero 4", len(todo.Contents))
	}
}

func TestNoSeBorraUnBucketConCosasDentro(t *testing.T) {
	ctx := context.Background()
	c := nube(t)
	if _, err := c.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String("lleno")}); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}
	if _, err := c.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String("lleno"), Key: aws.String("algo"), Body: bytes.NewReader([]byte("x")),
	}); err != nil {
		t.Fatalf("PutObject: %v", err)
	}
	if _, err := c.DeleteBucket(ctx, &s3.DeleteBucketInput{Bucket: aws.String("lleno")}); err == nil {
		t.Fatal("borro un bucket con cosas dentro, y S3 no deja")
	}

	// Se vacia y ahora si.
	if _, err := c.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String("lleno"), Key: aws.String("algo"),
	}); err != nil {
		t.Fatalf("DeleteObject: %v", err)
	}
	if _, err := c.DeleteBucket(ctx, &s3.DeleteBucketInput{Bucket: aws.String("lleno")}); err != nil {
		t.Fatalf("DeleteBucket: %v", err)
	}
}

// Borrar en S3 es idempotente: borrar algo que ya no esta NO es un error. Si lo fuera, un
// reintento de red despues de un borrado bueno fallaria.
func TestBorrarDosVecesNoEsUnError(t *testing.T) {
	ctx := context.Background()
	c := nube(t)
	if _, err := c.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String("b")}); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}
	for i := range 2 {
		if _, err := c.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String("b"), Key: aws.String("nunca-existio"),
		}); err != nil {
			t.Fatalf("DeleteObject numero %d: %v", i+1, err)
		}
	}
}

func TestListarBuckets(t *testing.T) {
	ctx := context.Background()
	c := nube(t)
	for _, n := range []string{"zeta", "alfa"} {
		if _, err := c.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(n)}); err != nil {
			t.Fatalf("CreateBucket %s: %v", n, err)
		}
	}
	res, err := c.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		t.Fatalf("ListBuckets: %v", err)
	}
	if len(res.Buckets) != 2 {
		t.Fatalf("buckets = %d, quiero 2", len(res.Buckets))
	}
	// Ordenados: si no ordenas, el mapa de Go los devuelve en orden distinto cada vez.
	if aws.ToString(res.Buckets[0].Name) != "alfa" {
		t.Errorf("el primero es %q, quiero alfa", aws.ToString(res.Buckets[0].Name))
	}
}
