package main

// El servicio CloudWatch Logs. Habla el mismo protocolo que DynamoDB y SQS —POST a "/", JSON, y
// la operacion en la cabecera— con su propio prefijo y su propia version:
//
//	X-Amz-Target: Logs_20140328.PutLogEvents
//
// ⚠️ Y trae una sorpresa pequeña que se ve nada mas mirar un cuerpo: aqui los campos van en
// minuscula. `logGroupName`, no `LogGroupName`. DynamoDB manda `TableName` y SQS manda
// `QueueUrl`, con mayuscula, y este manda `logGroupName`. Es el mismo protocolo, la misma
// empresa y la misma cabecera, y la convencion de nombres es distinta. No es un error de nadie:
// es que cada equipo eligio la suya y ya no se puede cambiar sin romper a todo el mundo.
//
// Lo que se implementa, que es lo que se usa un dia normal:
//
//	CreateLogGroup · DeleteLogGroup · DescribeLogGroups · PutRetentionPolicy
//	CreateLogStream · DeleteLogStream · DescribeLogStreams
//	PutLogEvents · GetLogEvents · FilterLogEvents
//
// ⛔ Lo que NO: metricas y filtros de metrica, suscripciones, `StartQuery` (o sea, Logs
// Insights), exportar a S3, cifrado con KMS ni Live Tail.

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

const objetivoLogs = "Logs_20140328."

type ServicioLogs struct {
	reg      *Registros
	registro func(formato string, args ...any)
	cuenta   string
	region   string
}

func NuevoLogs(r *Registros, registro func(string, ...any)) *ServicioLogs {
	if registro == nil {
		registro = func(string, ...any) {}
	}
	return &ServicioLogs{reg: r, registro: registro,
		cuenta: "000000000000", region: "us-east-1"}
}

func (s *ServicioLogs) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	destino := r.Header.Get("X-Amz-Target")
	op := destino
	if i := strings.LastIndex(destino, "."); i >= 0 {
		op = destino[i+1:]
	}
	cuerpo, err := io.ReadAll(r.Body)
	if err != nil {
		s.fallo(w, http.StatusBadRequest, "SerializationException", err.Error(), nil)
		return
	}
	s.registro("POST / [Logs %s] %d byte(s)", op, len(cuerpo))

	switch op {
	case "CreateLogGroup":
		s.crearGrupo(w, cuerpo)
	case "DeleteLogGroup":
		s.borrarGrupo(w, cuerpo)
	case "DescribeLogGroups":
		s.listarGrupos(w, cuerpo)
	case "PutRetentionPolicy":
		s.retener(w, cuerpo, false)
	case "DeleteRetentionPolicy":
		s.retener(w, cuerpo, true)
	case "CreateLogStream":
		s.crearFlujo(w, cuerpo)
	case "DeleteLogStream":
		s.borrarFlujo(w, cuerpo)
	case "DescribeLogStreams":
		s.listarFlujos(w, cuerpo)
	case "PutLogEvents":
		s.poner(w, cuerpo)
	case "GetLogEvents":
		s.leer(w, cuerpo)
	case "FilterLogEvents":
		s.filtrar(w, cuerpo)
	default:
		s.fallo(w, http.StatusBadRequest, "UnknownOperationException",
			"nubecita no implementa "+op, nil)
	}
}

func (s *ServicioLogs) responder(w http.ResponseWriter, v any) {
	// ⚠️ json-1.1, no 1.0. DynamoDB y SQS usan 1.0. Los SDK no se quejan si te equivocas, pero
	// esta escrito en la definicion del servicio y copiarlo mal es empezar a mentir.
	w.Header().Set("Content-Type", "application/x-amz-json-1.1")
	w.WriteHeader(http.StatusOK)
	if v == nil {
		_, _ = w.Write([]byte("{}"))
		return
	}
	_ = json.NewEncoder(w).Encode(v)
}

// fallo escribe un error con la forma que espera el SDK. `extra` deja meter campos propios del
// error, que en esta API existen y son utiles: vease `InvalidSequenceTokenException`.
func (s *ServicioLogs) fallo(w http.ResponseWriter, estado int, tipo, mensaje string,
	extra map[string]any) {
	s.registro("  -> %d %s", estado, tipo)
	cuerpo := map[string]any{"__type": tipo, "message": mensaje}
	for k, v := range extra {
		cuerpo[k] = v
	}
	w.Header().Set("Content-Type", "application/x-amz-json-1.1")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(cuerpo)
}

// deError traduce un error del almacen a la excepcion que toca. Se hace en un sitio para que dos
// operaciones no contesten cosas distintas al mismo problema.
func (s *ServicioLogs) deError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errNoExisteGrupo), errors.Is(err, errNoExisteFlujo):
		s.fallo(w, http.StatusBadRequest, "ResourceNotFoundException", err.Error(), nil)
	case errors.Is(err, errYaExisteGrupo), errors.Is(err, errYaExisteFlujo):
		s.fallo(w, http.StatusBadRequest, "ResourceAlreadyExistsException", err.Error(), nil)
	default:
		s.fallo(w, http.StatusBadRequest, "InvalidParameterException", err.Error(), nil)
	}
}

func (s *ServicioLogs) arnGrupo(nombre string) string {
	return "arn:aws:logs:" + s.region + ":" + s.cuenta + ":log-group:" + nombre + ":*"
}

func (s *ServicioLogs) arnFlujo(grupo, flujo string) string {
	return "arn:aws:logs:" + s.region + ":" + s.cuenta + ":log-group:" + grupo +
		":log-stream:" + flujo
}

// ── grupos ───────────────────────────────────────────────────────────────

type peticionGrupo struct {
	LogGroupName       string `json:"logGroupName"`
	LogGroupNamePrefix string `json:"logGroupNamePrefix"`
	RetentionInDays    int32  `json:"retentionInDays"`
	Limit              int    `json:"limit"`
}

func (s *ServicioLogs) crearGrupo(w http.ResponseWriter, cuerpo []byte) {
	var p peticionGrupo
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		s.fallo(w, http.StatusBadRequest, "SerializationException", err.Error(), nil)
		return
	}
	if p.LogGroupName == "" {
		s.fallo(w, http.StatusBadRequest, "InvalidParameterException",
			"falta logGroupName", nil)
		return
	}
	if err := s.reg.CrearGrupo(p.LogGroupName); err != nil {
		s.deError(w, err)
		return
	}
	s.registro("  -> grupo %s creado", p.LogGroupName)
	s.responder(w, nil)
}

func (s *ServicioLogs) borrarGrupo(w http.ResponseWriter, cuerpo []byte) {
	var p peticionGrupo
	_ = json.Unmarshal(cuerpo, &p)
	if err := s.reg.BorrarGrupo(p.LogGroupName); err != nil {
		s.deError(w, err)
		return
	}
	s.registro("  -> grupo %s borrado", p.LogGroupName)
	s.responder(w, nil)
}

func (s *ServicioLogs) retener(w http.ResponseWriter, cuerpo []byte, quitar bool) {
	var p peticionGrupo
	_ = json.Unmarshal(cuerpo, &p)
	dias := p.RetentionInDays
	if quitar {
		dias = 0
	}
	if err := s.reg.Retener(p.LogGroupName, dias); err != nil {
		s.deError(w, err)
		return
	}
	s.responder(w, nil)
}

func (s *ServicioLogs) listarGrupos(w http.ResponseWriter, cuerpo []byte) {
	var p peticionGrupo
	_ = json.Unmarshal(cuerpo, &p)
	fichas := s.reg.ListarGrupos(p.LogGroupNamePrefix)
	salida := make([]map[string]any, 0, len(fichas))
	for _, f := range fichas {
		g := map[string]any{
			"logGroupName": f.Nombre,
			"creationTime": f.Creado.UnixMilli(),
			"arn":          s.arnGrupo(f.Nombre),
			"storedBytes":  f.Bytes,
			// metricFilterCount va siempre: el SDK lo espera y ponerlo a cero es la verdad.
			"metricFilterCount": 0,
		}
		if f.Retencion > 0 {
			g["retentionInDays"] = f.Retencion
		}
		salida = append(salida, g)
	}
	s.responder(w, map[string]any{"logGroups": salida})
}

// ── flujos ───────────────────────────────────────────────────────────────

type peticionFlujo struct {
	LogGroupName        string `json:"logGroupName"`
	LogStreamName       string `json:"logStreamName"`
	LogStreamNamePrefix string `json:"logStreamNamePrefix"`
	Limit               int    `json:"limit"`
}

func (s *ServicioLogs) crearFlujo(w http.ResponseWriter, cuerpo []byte) {
	var p peticionFlujo
	_ = json.Unmarshal(cuerpo, &p)
	if err := s.reg.CrearFlujo(p.LogGroupName, p.LogStreamName); err != nil {
		s.deError(w, err)
		return
	}
	s.registro("  -> flujo %s/%s creado", p.LogGroupName, p.LogStreamName)
	s.responder(w, nil)
}

func (s *ServicioLogs) borrarFlujo(w http.ResponseWriter, cuerpo []byte) {
	var p peticionFlujo
	_ = json.Unmarshal(cuerpo, &p)
	if err := s.reg.BorrarFlujo(p.LogGroupName, p.LogStreamName); err != nil {
		s.deError(w, err)
		return
	}
	s.responder(w, nil)
}

func (s *ServicioLogs) listarFlujos(w http.ResponseWriter, cuerpo []byte) {
	var p peticionFlujo
	_ = json.Unmarshal(cuerpo, &p)
	fichas, err := s.reg.ListarFlujos(p.LogGroupName, p.LogStreamNamePrefix)
	if err != nil {
		s.deError(w, err)
		return
	}
	salida := make([]map[string]any, 0, len(fichas))
	for _, f := range fichas {
		fl := map[string]any{
			"logStreamName": f.Nombre,
			"creationTime":  f.Creado.UnixMilli(),
			"arn":           s.arnFlujo(p.LogGroupName, f.Nombre),
			"storedBytes":   0,
		}
		if f.Apuntes > 0 {
			fl["firstEventTimestamp"] = f.Primero
			fl["lastEventTimestamp"] = f.Ultimo
			fl["lastIngestionTime"] = f.Subida.UnixMilli()
			fl["uploadSequenceToken"] = f.Token
		}
		salida = append(salida, fl)
	}
	s.responder(w, map[string]any{"logStreams": salida})
}

// ── escribir y leer ──────────────────────────────────────────────────────

type apunteJSON struct {
	Timestamp int64  `json:"timestamp"`
	Message   string `json:"message"`
}

type peticionPoner struct {
	LogGroupName  string       `json:"logGroupName"`
	LogStreamName string       `json:"logStreamName"`
	SequenceToken string       `json:"sequenceToken"`
	LogEvents     []apunteJSON `json:"logEvents"`
}

func (s *ServicioLogs) poner(w http.ResponseWriter, cuerpo []byte) {
	var p peticionPoner
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		s.fallo(w, http.StatusBadRequest, "SerializationException", err.Error(), nil)
		return
	}
	apuntes := make([]Apunte, 0, len(p.LogEvents))
	for _, e := range p.LogEvents {
		apuntes = append(apuntes, Apunte{Marca: e.Timestamp, Mensaje: e.Message})
	}
	token, rech, err := s.reg.Poner(p.LogGroupName, p.LogStreamName, p.SequenceToken,
		apuntes, time.Now())
	if errors.Is(err, errTokenEquivocado) {
		// 🥇 El error que trae la respuesta dentro. Lo mismo que hace el `AccessDenied` de los
		// permisos y lo contrario que hace el de la firma: aqui el servidor SI sabe cual era
		// el bueno, asi que lo dice.
		s.fallo(w, http.StatusBadRequest, "InvalidSequenceTokenException",
			"el sequenceToken no es el que toca", map[string]any{
				"expectedSequenceToken": token,
			})
		return
	}
	if err != nil {
		s.deError(w, err)
		return
	}
	res := map[string]any{"nextSequenceToken": token}
	// rejectedLogEventsInfo solo aparece si hubo rechazos, igual que en el servicio de verdad.
	if rech.Viejos > 0 || rech.Nuevos > 0 {
		info := map[string]any{}
		if rech.Viejos > 0 {
			info["tooOldLogEventEndIndex"] = rech.Viejos
		}
		if rech.Nuevos > 0 {
			info["tooNewLogEventStartIndex"] = len(p.LogEvents) - rech.Nuevos
		}
		res["rejectedLogEventsInfo"] = info
		s.registro("  -> %d apunte(s) rechazado(s) por fecha", rech.Viejos+rech.Nuevos)
	}
	s.responder(w, res)
}

type peticionLeer struct {
	LogGroupName  string `json:"logGroupName"`
	LogStreamName string `json:"logStreamName"`
	StartTime     *int64 `json:"startTime"`
	EndTime       *int64 `json:"endTime"`
	Limit         int    `json:"limit"`
	NextToken     string `json:"nextToken"`
	StartFromHead bool   `json:"startFromHead"`
}

func (s *ServicioLogs) leer(w http.ResponseWriter, cuerpo []byte) {
	var p peticionLeer
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		s.fallo(w, http.StatusBadRequest, "SerializationException", err.Error(), nil)
		return
	}
	pag, err := s.reg.Leer(p.LogGroupName, p.LogStreamName, p.StartTime, p.EndTime,
		p.Limit, p.NextToken, p.StartFromHead)
	if err != nil {
		s.deError(w, err)
		return
	}
	eventos := make([]map[string]any, 0, len(pag.Apuntes))
	for _, a := range pag.Apuntes {
		eventos = append(eventos, map[string]any{
			"timestamp": a.Marca, "message": a.Mensaje, "ingestionTime": a.Marca,
		})
	}
	// ⚠️ Los dos tokens van SIEMPRE, tambien cuando `eventos` viene vacio. Es lo que hace el
	// servicio de verdad y es la trampa del bucle infinito: se para comparando tokens, no
	// mirando si hay token.
	s.responder(w, map[string]any{
		"events":            eventos,
		"nextForwardToken":  pag.Adelante,
		"nextBackwardToken": pag.Atras,
	})
}

type peticionFiltrar struct {
	LogGroupName        string `json:"logGroupName"`
	LogStreamNamePrefix string `json:"logStreamNamePrefix"`
	FilterPattern       string `json:"filterPattern"`
	StartTime           *int64 `json:"startTime"`
	EndTime             *int64 `json:"endTime"`
	Limit               int    `json:"limit"`
}

func (s *ServicioLogs) filtrar(w http.ResponseWriter, cuerpo []byte) {
	var p peticionFiltrar
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		s.fallo(w, http.StatusBadRequest, "SerializationException", err.Error(), nil)
		return
	}
	filas, err := s.reg.Filtrar(p.LogGroupName, p.LogStreamNamePrefix, p.FilterPattern,
		p.StartTime, p.EndTime, p.Limit)
	if err != nil {
		s.deError(w, err)
		return
	}
	eventos := make([]map[string]any, 0, len(filas))
	for i, f := range filas {
		eventos = append(eventos, map[string]any{
			"logStreamName": f.Flujo,
			"timestamp":     f.Apunte.Marca,
			"message":       f.Apunte.Mensaje,
			"ingestionTime": f.Apunte.Marca,
			"eventId":       pad12(int64(i)),
		})
	}
	s.responder(w, map[string]any{
		"events": eventos, "searchedLogStreams": []any{},
	})
}
