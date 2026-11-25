// demo · conduce nubepanel sin teclado contra una nubecita de verdad y fotografía la pantalla.
//
//	go run ./10_proyecto/demo -nubecita ruta/a/nubecita.exe
//
// Arranca nubecita en un puerto libre, siembra buckets, colas y tablas con el SDK oficial de
// AWS, pulsa teclas como lo haría una persona y, al final, APAGA nubecita para ver qué hace el
// panel cuando se queda sin servidor. Cada «foto» es el View() real en ese instante.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbt "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"curso-tui/10_proyecto/panel"
	"curso-tui/comun"
)

// ── la cámara: un modelo que envuelve a otro (lección 6) ─────────────────────────────────
type fotoMsg string

type camara struct{ m tea.Model }

func (c camara) Init() tea.Cmd { return c.m.Init() }
func (c camara) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if f, ok := msg.(fotoMsg); ok {
		comun.Pantalla(string(f), c.m.View().Content)
		return c, nil
	}
	var cmd tea.Cmd
	c.m, cmd = c.m.Update(msg)
	return c, cmd
}
func (c camara) View() tea.View { return c.m.View() }

func puertoLibre() int {
	l, _ := net.Listen("tcp", "127.0.0.1:0")
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func esperarHTTP(url string) error {
	for i := 0; i < 100; i++ {
		if r, err := http.Get(url); err == nil {
			r.Body.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("nubecita no contesta en %s", url)
}

func sembrar(endpoint string) {
	ctx := context.Background()
	cfg := aws.Config{Region: "us-east-1", BaseEndpoint: aws.String(endpoint),
		Credentials: credentials.NewStaticCredentialsProvider("prueba", "prueba", "")}
	s3c := s3.NewFromConfig(cfg, func(o *s3.Options) { o.UsePathStyle = true })
	for b, n := range map[string]int{"fotos": 3, "copias-db": 2, "logs-2026": 5} {
		s3c.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(b)})
		for i := 0; i < n; i++ {
			s3c.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(b), Key: aws.String(fmt.Sprintf("obj-%d.txt", i)),
				Body: strings.NewReader("hola")})
		}
	}
	q := sqs.NewFromConfig(cfg)
	for c, n := range map[string]int{"pedidos": 4, "pedidos-muertos": 1, "correos": 0} {
		r, _ := q.CreateQueue(ctx, &sqs.CreateQueueInput{QueueName: aws.String(c)})
		for i := 0; i < n; i++ {
			q.SendMessage(ctx, &sqs.SendMessageInput{QueueUrl: r.QueueUrl, MessageBody: aws.String("pedido")})
		}
	}
	d := dynamodb.NewFromConfig(cfg)
	for t, n := range map[string]int{"usuarios": 3, "sesiones": 2} {
		d.CreateTable(ctx, &dynamodb.CreateTableInput{TableName: aws.String(t),
			AttributeDefinitions: []ddbt.AttributeDefinition{{AttributeName: aws.String("id"), AttributeType: ddbt.ScalarAttributeTypeS}},
			KeySchema:            []ddbt.KeySchemaElement{{AttributeName: aws.String("id"), KeyType: ddbt.KeyTypeHash}},
			BillingMode:          ddbt.BillingModePayPerRequest})
		for i := 0; i < n; i++ {
			d.PutItem(ctx, &dynamodb.PutItemInput{TableName: aws.String(t),
				Item: map[string]ddbt.AttributeValue{"id": &ddbt.AttributeValueMemberS{Value: fmt.Sprint(i)}}})
		}
	}
}

// encolar mete n mensajes en una cola mientras el panel está abierto: alguien más trabajando.
func encolar(endpoint, cola string, n int) {
	ctx := context.Background()
	q := sqs.NewFromConfig(aws.Config{Region: "us-east-1", BaseEndpoint: aws.String(endpoint),
		Credentials: credentials.NewStaticCredentialsProvider("prueba", "prueba", "")})
	u, _ := q.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{QueueName: aws.String(cola)})
	for i := 0; i < n; i++ {
		q.SendMessage(ctx, &sqs.SendMessageInput{QueueUrl: u.QueueUrl, MessageBody: aws.String("correo")})
	}
}

func main() {
	bin := flag.String("nubecita", "", "binario de nubecita")
	soloSembrar := flag.String("sembrar", "", "solo siembra esta nubecita ya arrancada y sale")
	flag.Parse()
	if *soloSembrar != "" {
		sembrar(*soloSembrar)
		return
	}
	puerto := puertoLibre()
	endpoint := fmt.Sprintf("http://127.0.0.1:%d", puerto)
	nube := exec.Command(*bin, "-puerto", fmt.Sprint(puerto), "-callado")
	nube.Stdout, nube.Stderr = io.Discard, io.Discard
	if err := nube.Start(); err != nil {
		fmt.Println("no arranca nubecita:", err)
		os.Exit(2)
	}
	defer nube.Process.Kill()
	if err := esperarHTTP(endpoint); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	sembrar(endpoint)
	fmt.Println("  nubecita sembrada: 3 buckets, 3 colas, 2 tablas")

	p := tea.NewProgram(camara{panel.Nuevo(panel.NuevaAWS(endpoint))},
		tea.WithInput(nil), tea.WithOutput(io.Discard), tea.WithWindowSize(80, 24), tea.WithoutSignals())
	t := func(s string) { p.Send(comun.Tecla(s)); time.Sleep(150 * time.Millisecond) }
	foto := func(s string) { p.Send(fotoMsg(s)); time.Sleep(100 * time.Millisecond) }
	go func() {
		time.Sleep(800 * time.Millisecond)
		foto("1 · al arrancar")
		t("tab")
		foto("2 · tab: las colas")
		encolar(endpoint, "correos", 6)
		fmt.Println("  (otro proceso mete 6 mensajes en «correos»; nadie toca el panel)")
		time.Sleep(panel.Intervalo + 700*time.Millisecond)
		foto("3 · se refresca solo")
		t("p")
		foto("4 · p: purgar «correos»")
		t("s")
		time.Sleep(600 * time.Millisecond)
		foto("5 · s: purgada")
		t("/")
		for _, c := range "muer" {
			t(string(c))
		}
		foto("6 · /muer: filtrando")
		t("esc")
		t("tab")
		foto("7 · tab: las tablas")
		nube.Process.Kill()
		nube.Wait()
		fmt.Println("  (se apaga nubecita)")
		time.Sleep(panel.Intervalo + 1500*time.Millisecond)
		foto("8 · sin servidor")
		t("q")
	}()
	_, err := p.Run()
	fmt.Println("  el panel terminó con q · error:", err)
}
