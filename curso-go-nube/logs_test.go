package main

// Las pruebas de CloudWatch Logs con el SDK oficial de Amazon delante. Si el SDK habla con esto
// sin enterarse de que no es Amazon, es que el protocolo esta bien; si no, no.
//
//	go test ./... -count=1 -race -run Logs

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	logtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	smithy "github.com/aws/smithy-go"
)

func clienteLogs(url string) *cloudwatchlogs.Client {
	return cloudwatchlogs.New(cloudwatchlogs.Options{
		Region:       "us-east-1",
		Credentials:  credencialesFijas("prueba", "prueba"),
		BaseEndpoint: aws.String(url),
	})
}

func nubeConLogs(t *testing.T) (*cloudwatchlogs.Client, *Registros) {
	t.Helper()
	reg := NuevosRegistros()
	srv := httptest.NewServer(Raiz(http.NewServeMux(), map[string]http.Handler{
		objetivoLogs: NuevoLogs(reg, nil),
	}, nil, nil))
	t.Cleanup(srv.Close)
	return clienteLogs(srv.URL), reg
}

func apunte(cuando time.Time, texto string) logtypes.InputLogEvent {
	return logtypes.InputLogEvent{
		Timestamp: aws.Int64(cuando.UnixMilli()),
		Message:   aws.String(texto),
	}
}

// El camino de todos los dias: crear, escribir y leer, con el SDK de verdad.
func TestLogsElSDKEscribeYLee(t *testing.T) {
	c, _ := nubeConLogs(t)
	ctx := context.Background()
	ahora := time.Now()

	if _, err := c.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String("/aws/lambda/saludar")}); err != nil {
		t.Fatalf("CreateLogGroup: %v", err)
	}
	if _, err := c.CreateLogStream(ctx, &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String("/aws/lambda/saludar"),
		LogStreamName: aws.String("flujo-1")}); err != nil {
		t.Fatalf("CreateLogStream: %v", err)
	}
	puesto, err := c.PutLogEvents(ctx, &cloudwatchlogs.PutLogEventsInput{
		LogGroupName:  aws.String("/aws/lambda/saludar"),
		LogStreamName: aws.String("flujo-1"),
		LogEvents: []logtypes.InputLogEvent{
			apunte(ahora, "START RequestId: 1"),
			apunte(ahora.Add(time.Millisecond), "hola desde la funcion"),
		},
	})
	if err != nil {
		t.Fatalf("PutLogEvents: %v", err)
	}
	if aws.ToString(puesto.NextSequenceToken) == "" {
		t.Fatal("PutLogEvents tiene que devolver el siguiente token")
	}

	leido, err := c.GetLogEvents(ctx, &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String("/aws/lambda/saludar"),
		LogStreamName: aws.String("flujo-1"),
		StartFromHead: aws.Bool(true),
	})
	if err != nil {
		t.Fatalf("GetLogEvents: %v", err)
	}
	if len(leido.Events) != 2 {
		t.Fatalf("tenian que salir dos apuntes y salieron %d", len(leido.Events))
	}
	if aws.ToString(leido.Events[1].Message) != "hola desde la funcion" {
		t.Fatalf("el mensaje llego mal: %q", aws.ToString(leido.Events[1].Message))
	}
	// 🥇 Los dos tokens van siempre, tambien cuando ya no queda nada.
	if aws.ToString(leido.NextForwardToken) == "" || aws.ToString(leido.NextBackwardToken) == "" {
		t.Fatal("los dos tokens tienen que venir siempre: es la trampa del bucle infinito")
	}
}

// 🥇 El error que trae dentro la respuesta que te falta. Igual que el AccessDenied del episodio
// 10, y al reves que el de la firma: aqui el servidor SI sabe cual era el token bueno.
func TestLogsElErrorDelTokenTraeElTokenBueno(t *testing.T) {
	c, _ := nubeConLogs(t)
	ctx := context.Background()
	ahora := time.Now()

	_, _ = c.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String("g")})
	_, _ = c.CreateLogStream(ctx, &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName: aws.String("g"), LogStreamName: aws.String("f")})
	primero, err := c.PutLogEvents(ctx, &cloudwatchlogs.PutLogEventsInput{
		LogGroupName: aws.String("g"), LogStreamName: aws.String("f"),
		LogEvents: []logtypes.InputLogEvent{apunte(ahora, "uno")}})
	if err != nil {
		t.Fatalf("la primera escritura no lleva token y tiene que valer: %v", err)
	}

	_, err = c.PutLogEvents(ctx, &cloudwatchlogs.PutLogEventsInput{
		LogGroupName: aws.String("g"), LogStreamName: aws.String("f"),
		SequenceToken: aws.String("me-lo-invento"),
		LogEvents:     []logtypes.InputLogEvent{apunte(ahora, "dos")}})
	var malo *logtypes.InvalidSequenceTokenException
	if !errors.As(err, &malo) {
		t.Fatalf("tenia que ser un InvalidSequenceTokenException tipado y fue: %v", err)
	}
	if aws.ToString(malo.ExpectedSequenceToken) != aws.ToString(primero.NextSequenceToken) {
		t.Fatalf("el error tenia que traer el token bueno y trajo %q",
			aws.ToString(malo.ExpectedSequenceToken))
	}
}

// Lo que no existe se dice con la excepcion de este servicio, tipada, no con un 500 cualquiera.
func TestLogsLoQueNoExisteSeDiceBien(t *testing.T) {
	c, _ := nubeConLogs(t)
	_, err := c.PutLogEvents(context.Background(), &cloudwatchlogs.PutLogEventsInput{
		LogGroupName: aws.String("no-esta"), LogStreamName: aws.String("f"),
		LogEvents: []logtypes.InputLogEvent{apunte(time.Now(), "hola")}})
	var noEsta *logtypes.ResourceNotFoundException
	if !errors.As(err, &noEsta) {
		t.Fatalf("tenia que ser un ResourceNotFoundException y fue: %v", err)
	}
}

// DescribeLogStreams tiene que contar lo que hay: cuando empezo, cuando acabo y por donde va.
func TestLogsDescribeCuentaLoQueHay(t *testing.T) {
	c, reg := nubeConLogs(t)
	ctx := context.Background()
	base := time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC)

	reg.Anexar("/aws/lambda/saludar", "flujo-1", base.UnixMilli(), "primera")
	reg.Anexar("/aws/lambda/saludar", "flujo-1", base.Add(time.Second).UnixMilli(), "ultima")

	grupos, err := c.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{})
	if err != nil {
		t.Fatalf("DescribeLogGroups: %v", err)
	}
	if len(grupos.LogGroups) != 1 {
		t.Fatalf("tenia que haber un grupo y hay %d", len(grupos.LogGroups))
	}
	// ⚠️ El ARN de un grupo acaba en ":*", y eso es parte del nombre, no un comodin nuestro.
	if arn := aws.ToString(grupos.LogGroups[0].Arn); !strings.HasSuffix(arn, ":*") {
		t.Fatalf("el ARN de un grupo acaba en :* y salio %q", arn)
	}

	flujos, err := c.DescribeLogStreams(ctx, &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String("/aws/lambda/saludar")})
	if err != nil {
		t.Fatalf("DescribeLogStreams: %v", err)
	}
	if len(flujos.LogStreams) != 1 {
		t.Fatalf("tenia que haber un flujo y hay %d", len(flujos.LogStreams))
	}
	f := flujos.LogStreams[0]
	if aws.ToInt64(f.FirstEventTimestamp) != base.UnixMilli() ||
		aws.ToInt64(f.LastEventTimestamp) != base.Add(time.Second).UnixMilli() {
		t.Fatalf("las marcas del flujo salieron mal: %d y %d",
			aws.ToInt64(f.FirstEventTimestamp), aws.ToInt64(f.LastEventTimestamp))
	}
}

// FilterLogEvents busca en TODOS los flujos del grupo a la vez, y dice de cual vino cada linea.
func TestLogsFiltrarBuscaEnTodosLosFlujos(t *testing.T) {
	c, reg := nubeConLogs(t)
	base := time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC)

	reg.Anexar("g", "a", base.UnixMilli(), "ERROR se cayo algo")
	reg.Anexar("g", "b", base.Add(time.Second).UnixMilli(), "todo bien")
	reg.Anexar("g", "b", base.Add(2*time.Second).UnixMilli(), "ERROR otra vez")

	res, err := c.FilterLogEvents(context.Background(), &cloudwatchlogs.FilterLogEventsInput{
		LogGroupName:  aws.String("g"),
		FilterPattern: aws.String("ERROR"),
	})
	if err != nil {
		t.Fatalf("FilterLogEvents: %v", err)
	}
	if len(res.Events) != 2 {
		t.Fatalf("hay dos lineas con ERROR y salieron %d", len(res.Events))
	}
	if aws.ToString(res.Events[0].LogStreamName) != "a" ||
		aws.ToString(res.Events[1].LogStreamName) != "b" {
		t.Fatalf("cada linea tiene que decir de que flujo vino: %+v", res.Events)
	}
}

// Y los permisos del episodio 10 valen aqui sin tocar nada: el servicio nuevo entra por la misma
// puerta. Lo unico suyo es como se traduce su peticion a un permiso.
func TestLogsPasaPorLosPermisos(t *testing.T) {
	perm := NuevosPermisos(nil)
	perm.Poner("prueba", "lector", []Politica{pol(t, `{"Version":"2012-10-17","Statement":[
		{"Effect":"Allow","Action":["logs:CreateLogGroup","logs:CreateLogStream","logs:GetLogEvents"],
		 "Resource":"arn:aws:logs:us-east-1:000000000000:log-group:/aws/lambda/saludar:*"}]}`)})

	reg := NuevosRegistros()
	raiz := Raiz(http.NewServeMux(), map[string]http.Handler{
		objetivoLogs: NuevoLogs(reg, nil),
	}, nil, nil)
	srv := httptest.NewServer(NuevoAutorizador(perm, nil).Envolver(raiz))
	defer srv.Close()

	c := clienteLogs(srv.URL)
	ctx := context.Background()

	if _, err := c.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String("/aws/lambda/saludar")}); err != nil {
		t.Fatalf("crear ese grupo estaba permitido y fallo: %v", err)
	}
	// Otro grupo, el mismo permiso: el recurso no encaja.
	_, err := c.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String("/aws/lambda/otra")})
	if codigo(err) != "AccessDeniedException" {
		t.Fatalf("tenia que ser AccessDeniedException y fue %q (%v)", codigo(err), err)
	}
	// Y escribir tampoco: la accion no esta en la politica.
	_, _ = c.CreateLogStream(ctx, &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName: aws.String("/aws/lambda/saludar"), LogStreamName: aws.String("f")})
	_, err = c.PutLogEvents(ctx, &cloudwatchlogs.PutLogEventsInput{
		LogGroupName: aws.String("/aws/lambda/saludar"), LogStreamName: aws.String("f"),
		LogEvents: []logtypes.InputLogEvent{apunte(time.Now(), "hola")}})
	var api smithy.APIError
	if !errors.As(err, &api) || api.ErrorCode() != "AccessDeniedException" {
		t.Fatalf("PutLogEvents no estaba permitido: %v", err)
	}
	// Y el mensaje dice exactamente que permiso falta y sobre que recurso.
	if !strings.Contains(api.ErrorMessage(), "logs:PutLogEvents") ||
		!strings.Contains(api.ErrorMessage(), "log-group:/aws/lambda/saludar:*") {
		t.Fatalf("el error tiene que decir que falta: %q", api.ErrorMessage())
	}
}
