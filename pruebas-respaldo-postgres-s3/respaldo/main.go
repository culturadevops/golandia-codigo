// respaldo: un job de Go que respalda Postgres en S3 sin cargar el respaldo en memoria, aplica una
// retención y verifica restaurando. Demo del vídeo 561 de golandia (canal: «Respaldo de Postgres a S3 con Go»).
//
//	respaldo respaldar  -fecha 2026-10-21T0300 [-modo fc|fc+gzip|z0+gzip] [-base tienda]
//	respaldo ingenuo    -fecha ...   (lee todo el pg_dump en memoria y luego sube: para comparar)
//	respaldo sin-wait   -fecha ...   (no mira cómo terminó pg_dump: el error de verdad)
//	respaldo listar     [-base tienda]
//	respaldo retener    [-base tienda] [-ultimos 3] [-dias 7] [-borrar]
//	respaldo restaurar  -clave respaldos/tienda/2026-10-21T0300.dump.gz
//	respaldo verificar
//	respaldo bucket     (crea el bucket de la demo)
//
// pg_dump y pg_restore leen PGHOST/PGUSER/PGPASSWORD del entorno (PGHOST_DESTINO para restaurar).
// S3: AWS_ENDPOINT_URL (moto en la demo), BUCKET.
package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const formatoFecha = "2006-01-02T1504"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "uso: respaldo respaldar|ingenuo|sin-wait|listar|retener|restaurar|verificar")
		os.Exit(2)
	}
	fs := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	fecha := fs.String("fecha", "", "marca del respaldo, AAAA-MM-DDTHHMM")
	modo := fs.String("modo", "fc", "fc | fc+gzip | z0+gzip")
	base := fs.String("base", "tienda", "base de datos")
	clave := fs.String("clave", "", "objeto de S3 a restaurar")
	ultimos := fs.Int("ultimos", 3, "retención: los N más nuevos")
	dias := fs.Int("dias", 7, "retención: además, el último de cada uno de los D días más recientes")
	borrar := fs.Bool("borrar", false, "retención: borrar de verdad (sin esto solo dice qué borraría)")
	fs.Parse(os.Args[2:])

	ctx := context.Background()
	s3c := clienteS3(ctx)
	bucket := os.Getenv("BUCKET")

	var err error
	switch os.Args[1] {
	case "respaldar":
		err = respaldar(ctx, s3c, bucket, *base, *fecha, *modo)
	case "ingenuo":
		err = ingenuo(ctx, s3c, bucket, *base, *fecha)
	case "sin-wait":
		err = sinWait(ctx, s3c, bucket, *base, *fecha)
	case "listar":
		err = listar(ctx, s3c, bucket, *base)
	case "retener":
		err = retener(ctx, s3c, bucket, *base, *ultimos, *dias, *borrar)
	case "restaurar":
		err = restaurar(ctx, s3c, bucket, *clave)
	case "verificar":
		err = verificar()
	case "bucket":
		_, err = s3c.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})
		if err == nil {
			fmt.Println("ok  bucket", bucket)
		}
	default:
		err = fmt.Errorf("subcomando desconocido %q", os.Args[1])
	}
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
}

func clienteS3(ctx context.Context) *s3.Client {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}
	// moto sirve los buckets por ruta (http://moto:5000/bucket/clave), no por subdominio.
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
		// Los objetos multiparte traen una suma «compuesta» que el SDK no valida al bajarlos, y lo avisa en cada
		// GetObject. La integridad la comprueba la restauración, que es la prueba que importa.
		o.DisableLogOutputChecksumValidationSkipped = true
	})
}

func claveDe(base, fecha, modo string) string {
	ext := ".dump"
	if strings.HasSuffix(modo, "gzip") {
		ext += ".gz"
	}
	return fmt.Sprintf("respaldos/%s/%s%s", base, fecha, ext)
}

// contador cuenta los bytes que pasan, sin guardarlos.
type contador struct{ n int64 }

func (c *contador) Write(p []byte) (int, error) { c.n += int64(len(p)); return len(p), nil }

// respaldar: pg_dump → (gzip) → io.Pipe → subida multiparte. Nada del respaldo vive entero en memoria.
func respaldar(ctx context.Context, s3c *s3.Client, bucket, base, fecha, modo string) error {
	t0 := time.Now()
	args := []string{"-Fc", "-d", base}
	if modo == "z0+gzip" {
		args = append(args, "-Z0") // que pg_dump no comprima: lo hace gzip
	}
	pg := exec.CommandContext(ctx, "pg_dump", args...)
	var stderr bytes.Buffer
	pg.Stderr = &stderr
	salida, err := pg.StdoutPipe()
	if err != nil {
		return err
	}
	if err := pg.Start(); err != nil {
		return err
	}

	pr, pw := io.Pipe()
	crudo, subido := &contador{}, &contador{}
	go func() {
		var w io.Writer = io.MultiWriter(pw, subido)
		var gz *gzip.Writer
		if strings.HasSuffix(modo, "gzip") {
			gz = gzip.NewWriter(w)
			w = gz
		}
		_, err := io.Copy(io.MultiWriter(w, crudo), salida)
		if err == nil && gz != nil {
			err = gz.Close()
		}
		// La línea que importa: si pg_dump terminó mal, el lector recibe ESE error y la subida se aborta.
		if werr := pg.Wait(); werr != nil && err == nil {
			err = fmt.Errorf("pg_dump: %v: %s", werr, strings.TrimSpace(stderr.String()))
		}
		pw.CloseWithError(err) // err == nil → EOF normal
	}()

	clave := claveDe(base, fecha, modo)
	// transfermanager sustituye a feature/s3/manager (que su go.mod ya marca Deprecated). Con un Body sin tamaño
	// conocido (un pipe) parte en trozos de PartSizeBytes y los sube en paralelo: multiparte.
	tm := transfermanager.New(s3c, func(o *transfermanager.Options) {
		o.PartSizeBytes = 8 << 20 // 8 MiB por parte
		o.Concurrency = 3
	})
	res, err := tm.UploadObject(ctx, &transfermanager.UploadObjectInput{Bucket: aws.String(bucket), Key: aws.String(clave), Body: pr})
	if err != nil {
		pr.CloseWithError(err) // que la goroutine y pg_dump no se queden bloqueados escribiendo
		return fmt.Errorf("subida abortada: %w", err)
	}
	fmt.Printf("ok  %s\n    pg_dump %s · subido %s · partes %d · %s · memoria pico %s\n",
		clave, mb(crudo.n), mb(subido.n), partes(subido.n, 8<<20), time.Since(t0).Round(100*time.Millisecond), picoRSS())
	_ = res
	return nil
}

// ingenuo: el respaldo entero a un []byte y luego PutObject. Funciona... hasta que la base crece.
func ingenuo(ctx context.Context, s3c *s3.Client, bucket, base, fecha string) error {
	t0 := time.Now()
	out, err := exec.CommandContext(ctx, "pg_dump", "-Fc", "-d", base).Output()
	if err != nil {
		return err
	}
	clave := claveDe(base, fecha, "fc")
	_, err = s3c.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(bucket), Key: aws.String(clave), Body: bytes.NewReader(out)})
	if err != nil {
		return err
	}
	fmt.Printf("ok  %s\n    pg_dump %s · subido de una vez · %s · memoria pico %s\n",
		clave, mb(int64(len(out))), time.Since(t0).Round(100*time.Millisecond), picoRSS())
	return nil
}

// sinWait: igual que respaldar, pero sin mirar cómo terminó pg_dump. Es el error que más se ve.
func sinWait(ctx context.Context, s3c *s3.Client, bucket, base, fecha string) error {
	pg := exec.CommandContext(ctx, "pg_dump", "-Fc", "-d", base)
	salida, err := pg.StdoutPipe()
	if err != nil {
		return err
	}
	pg.Stderr = io.Discard
	if err := pg.Start(); err != nil {
		return err
	}
	pr, pw := io.Pipe()
	go func() {
		gz := gzip.NewWriter(pw)
		io.Copy(gz, salida)
		gz.Close()
		pw.Close() // ← nunca se entera de que pg_dump falló
	}()
	clave := claveDe(base, fecha, "fc+gzip")
	if _, err := transfermanager.New(s3c).UploadObject(ctx, &transfermanager.UploadObjectInput{Bucket: aws.String(bucket), Key: aws.String(clave), Body: pr}); err != nil {
		return err
	}
	fmt.Printf("ok  %s  (el job terminó con código 0)\n", clave)
	return nil
}

type objeto struct {
	clave string
	fecha time.Time
	bytes int64
}

func objetos(ctx context.Context, s3c *s3.Client, bucket, base string) ([]objeto, error) {
	var res []objeto
	p := s3.NewListObjectsV2Paginator(s3c, &s3.ListObjectsV2Input{Bucket: aws.String(bucket), Prefix: aws.String("respaldos/" + base + "/")})
	for p.HasMorePages() {
		pag, err := p.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, o := range pag.Contents {
			nombre := strings.TrimPrefix(*o.Key, "respaldos/"+base+"/")
			// La fecha sale del NOMBRE, no de LastModified: un respaldo copiado o resubido conserva su fecha.
			f, err := time.Parse(formatoFecha, strings.SplitN(nombre, ".", 2)[0])
			if err != nil {
				continue
			}
			res = append(res, objeto{*o.Key, f, aws.ToInt64(o.Size)})
		}
	}
	sort.Slice(res, func(i, j int) bool { return res[i].fecha.After(res[j].fecha) })
	return res, nil
}

func listar(ctx context.Context, s3c *s3.Client, bucket, base string) error {
	lista, err := objetos(ctx, s3c, bucket, base)
	if err != nil {
		return err
	}
	for _, o := range lista {
		fmt.Printf("    %-45s %10s\n", o.clave, mb(o.bytes))
	}
	fmt.Printf("    %d respaldos\n", len(lista))
	// Una subida multiparte que nadie completa ni aborta no aparece en el listado, pero S3 cobra sus partes.
	mp, err := s3c.ListMultipartUploads(ctx, &s3.ListMultipartUploadsInput{Bucket: aws.String(bucket)})
	if err != nil {
		return err
	}
	fmt.Printf("    subidas multiparte a medias: %d\n", len(mp.Uploads))
	return nil
}

// retener: se quedan los N más nuevos y, además, el último de cada uno de los D días más recientes.
func retener(ctx context.Context, s3c *s3.Client, bucket, base string, n, d int, borrar bool) error {
	lista, err := objetos(ctx, s3c, bucket, base)
	if err != nil {
		return err
	}
	if len(lista) == 0 {
		return errors.New("no hay respaldos")
	}
	motivo := map[string]string{}
	for i, o := range lista {
		if i < n {
			motivo[o.clave] = fmt.Sprintf("entre los %d más nuevos", n)
		}
	}
	// los días se cuentan desde el respaldo más nuevo, no desde el reloj: si el job lleva días caído,
	// contar desde «hoy» borraría justo los últimos respaldos buenos.
	limite := lista[0].fecha.Truncate(24*time.Hour).AddDate(0, 0, -(d - 1))
	visto := map[string]bool{}
	for _, o := range lista {
		dia := o.fecha.Format("2006-01-02")
		if o.fecha.Before(limite) || visto[dia] {
			continue
		}
		visto[dia] = true
		if _, ya := motivo[o.clave]; !ya {
			motivo[o.clave] = "el último del " + dia
		}
	}
	var fuera []types.ObjectIdentifier
	for _, o := range lista {
		if m, ok := motivo[o.clave]; ok {
			fmt.Printf("    queda   %-40s (%s)\n", strings.TrimPrefix(o.clave, "respaldos/"), m)
		} else {
			fmt.Printf("    BORRA   %s\n", strings.TrimPrefix(o.clave, "respaldos/"))
			fuera = append(fuera, types.ObjectIdentifier{Key: aws.String(o.clave)})
		}
	}
	fmt.Printf("    quedan %d · se borran %d\n", len(lista)-len(fuera), len(fuera))
	if !borrar || len(fuera) == 0 {
		if !borrar {
			fmt.Println("    (ensayo: no se borró nada; usa -borrar)")
		}
		return nil
	}
	out, err := s3c.DeleteObjects(ctx, &s3.DeleteObjectsInput{Bucket: aws.String(bucket), Delete: &types.Delete{Objects: fuera}})
	if err != nil {
		return err
	}
	fmt.Printf("    borrados %d · errores %d\n", len(out.Deleted), len(out.Errors))
	return nil
}

// restaurar: GetObject → (gunzip) → stdin de pg_restore, en streaming, contra la base de PGHOST_DESTINO.
func restaurar(ctx context.Context, s3c *s3.Client, bucket, clave string) error {
	t0 := time.Now()
	obj, err := s3c.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(bucket), Key: aws.String(clave)})
	if err != nil {
		return err
	}
	defer obj.Body.Close()
	var r io.Reader = bufio.NewReader(obj.Body)
	if strings.HasSuffix(clave, ".gz") {
		gz, err := gzip.NewReader(r)
		if err != nil {
			return fmt.Errorf("el respaldo no es un gzip válido: %w", err)
		}
		r = gz
	}
	// Siempre a una base NUEVA: restaurar encima de otra a medias choca con lo que dejó el intento anterior.
	destino := append(os.Environ(), "PGHOST="+os.Getenv("PGHOST_DESTINO"))
	for _, q := range []string{"DROP DATABASE IF EXISTS tienda_restaurada", "CREATE DATABASE tienda_restaurada"} {
		c := exec.CommandContext(ctx, "psql", "-q", "-d", "postgres", "-c", q)
		c.Env = destino
		if out, err := c.CombinedOutput(); err != nil {
			return fmt.Errorf("%s: %v %s", q, err, out)
		}
	}
	// --single-transaction: o entra todo, o no queda nada a medias.
	pr := exec.CommandContext(ctx, "pg_restore", "--no-owner", "--exit-on-error", "--single-transaction", "-d", "tienda_restaurada")
	pr.Env = destino
	pr.Stdin = r
	var stderr bytes.Buffer
	pr.Stderr = &stderr
	if err := pr.Run(); err != nil {
		return fmt.Errorf("pg_restore: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	fmt.Printf("ok  restaurado %s en %s/tienda_restaurada · %s\n", clave, os.Getenv("PGHOST_DESTINO"), time.Since(t0).Round(100*time.Millisecond))
	return nil
}

// verificar: filas y una suma por tabla en el origen y en la base restaurada.
func verificar() error {
	const q = `SELECT 'clientes', count(*), coalesce(sum(id),0) FROM clientes
	           UNION ALL SELECT 'pedidos', count(*), coalesce(sum(monto),0) FROM pedidos ORDER BY 1`
	consulta := func(host, base string) (string, error) {
		c := exec.Command("psql", "-At", "-F", " ", "-d", base, "-c", q)
		c.Env = append(os.Environ(), "PGHOST="+host)
		out, err := c.CombinedOutput()
		return strings.TrimSpace(string(out)), err
	}
	a, err := consulta(os.Getenv("PGHOST"), "tienda")
	if err != nil {
		return fmt.Errorf("origen: %v %s", err, a)
	}
	b, err := consulta(os.Getenv("PGHOST_DESTINO"), "tienda_restaurada")
	if err != nil {
		return fmt.Errorf("destino: %v %s", err, b)
	}
	fmt.Printf("    %-9s %-30s %s\n", "tabla", "origen (filas · suma)", "restaurada")
	la, lb := strings.Split(a, "\n"), strings.Split(b, "\n")
	iguales := len(la) == len(lb)
	for i := range la {
		f := strings.Fields(la[i])
		g := []string{"-", "-", "-"}
		if i < len(lb) {
			g = strings.Fields(lb[i])
		}
		marca := "igual"
		if strings.Join(f, " ") != strings.Join(g, " ") {
			marca, iguales = "DISTINTA", false
		}
		fmt.Printf("    %-9s %-30s %-30s %s\n", f[0], f[1]+" · "+f[2], g[1]+" · "+g[2], marca)
	}
	if !iguales {
		return errors.New("la restauración NO coincide con el origen")
	}
	fmt.Println("ok  el respaldo restaura y coincide")
	return nil
}

func mb(n int64) string {
	if n < 1e6 {
		return fmt.Sprintf("%.0f KB", float64(n)/1e3)
	}
	return fmt.Sprintf("%.1f MB", float64(n)/1e6)
}

func partes(n, tam int64) int64 {
	if n == 0 {
		return 1
	}
	return (n + tam - 1) / tam
}

// picoRSS: el máximo de memoria residente del proceso (VmHWM de Linux).
func picoRSS() string {
	b, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return "?"
	}
	for _, l := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(l, "VmHWM:") {
			var kb int64
			fmt.Sscanf(strings.TrimSpace(strings.TrimPrefix(l, "VmHWM:")), "%d", &kb)
			return fmt.Sprintf("%.0f MB", float64(kb)/1024)
		}
	}
	return "?"
}
