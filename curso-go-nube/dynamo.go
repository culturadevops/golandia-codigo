package main

// El servicio DynamoDB. Y aqui viene la sorpresa del episodio 2: NO se parece en nada a S3.
//
//	S3         el metodo y la ruta dicen que quieres.  PUT /fotos/gato.png  ->  respuesta en XML
//	DynamoDB   SIEMPRE POST a "/", cuerpo JSON, y la operacion en una cabecera:
//	           X-Amz-Target: DynamoDB_20120810.PutItem
//
// Dos servicios de la misma nube, dos protocolos distintos. No es incoherencia: es historia.
// S3 es de 2006 y nacio como almacen de objetos sobre HTTP; DynamoDB es de 2012, cuando Amazon
// ya habia estandarizado el estilo "JSON con la operacion en la cabecera" para lo demas.
//
// Lo que se implementa aqui, y es todo lo que se usa un dia normal:
//
//	CreateTable · DescribeTable · DeleteTable · ListTables
//	PutItem · GetItem · DeleteItem
//	Query (con KeyConditionExpression completa) · Scan
//
// ⛔ Lo que NO: indices secundarios, UpdateItem, transacciones, streams, paginacion real y
// FilterExpression. No es una lista de tareas: es la frontera de un doble de pruebas.

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var (
	errNoExisteTabla = errors.New("la tabla no existe")
	errCondicion     = errors.New("la condicion no se cumple")
)

const prefijoTipo = "com.amazonaws.dynamodb.v20120810#"

type ServicioDynamo struct {
	tablas   *Tablas
	registro func(formato string, args ...any)
}

func NuevoDynamo(t *Tablas, registro func(string, ...any)) *ServicioDynamo {
	if registro == nil {
		registro = func(string, ...any) {}
	}
	return &ServicioDynamo{tablas: t, registro: registro}
}

func (s *ServicioDynamo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	destino := r.Header.Get("X-Amz-Target")
	op := destino
	if i := strings.LastIndex(destino, "."); i >= 0 {
		op = destino[i+1:]
	}
	cuerpo, err := io.ReadAll(r.Body)
	if err != nil {
		s.fallo(w, "SerializationException", err.Error())
		return
	}
	s.registro("POST / [%s] %d byte(s)", op, len(cuerpo))

	switch op {
	case "CreateTable":
		s.crearTabla(w, cuerpo)
	case "DescribeTable":
		s.describirTabla(w, cuerpo)
	case "DeleteTable":
		s.borrarTabla(w, cuerpo)
	case "ListTables":
		s.listarTablas(w)
	case "PutItem":
		s.ponerElemento(w, cuerpo)
	case "GetItem":
		s.obtenerElemento(w, cuerpo)
	case "DeleteItem":
		s.borrarElemento(w, cuerpo)
	case "Query":
		s.consultar(w, cuerpo)
	case "Scan":
		s.recorrer(w, cuerpo)
	default:
		// ⚠️ Fallo RUIDOSO y con el nombre de la operacion dentro. Un doble que contesta 200 a
		// algo que no implementa hace creer que la prueba paso.
		s.fallo(w, "UnknownOperationException",
			fmt.Sprintf("nubecita no implementa %q (episodio 2: S3 + DynamoDB)", op))
	}
}

// ── respuestas y errores ─────────────────────────────────────────────────
// El SDK convierte el campo `__type` en un tipo de error de Go que tu codigo puede preguntar con
// errors.As. Si contestas un 400 sin ese campo, el que te llama recibe "algo salio mal".

func (s *ServicioDynamo) responder(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/x-amz-json-1.0")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *ServicioDynamo) fallo(w http.ResponseWriter, tipo, mensaje string) {
	s.registro("  -> 400 %s", tipo)
	w.Header().Set("Content-Type", "application/x-amz-json-1.0")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"__type": prefijoTipo + tipo, "message": mensaje,
	})
}

func (s *ServicioDynamo) fallaPorError(w http.ResponseWriter, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, errNoExisteTabla):
		s.fallo(w, "ResourceNotFoundException", "Requested resource not found")
	case errors.Is(err, errCondicion):
		s.fallo(w, "ConditionalCheckFailedException", "The conditional request failed")
	default:
		s.fallo(w, "ValidationException", err.Error())
	}
	return true
}

// ── tablas ───────────────────────────────────────────────────────────────

type esquemaClave struct {
	AttributeName string `json:"AttributeName"`
	KeyType       string `json:"KeyType"` // HASH | RANGE
}

type defAtributo struct {
	AttributeName string `json:"AttributeName"`
	AttributeType string `json:"AttributeType"` // S | N | B
}

type peticionCrearTabla struct {
	TableName            string         `json:"TableName"`
	KeySchema            []esquemaClave `json:"KeySchema"`
	AttributeDefinitions []defAtributo  `json:"AttributeDefinitions"`
}

func (s *ServicioDynamo) fichaTabla(t *Tabla) map[string]any {
	esquema := []map[string]string{{"AttributeName": t.PK.Nombre, "KeyType": "HASH"}}
	atrib := []map[string]string{{"AttributeName": t.PK.Nombre, "AttributeType": t.PK.Tipo}}
	if t.SK != nil {
		esquema = append(esquema, map[string]string{"AttributeName": t.SK.Nombre, "KeyType": "RANGE"})
		atrib = append(atrib, map[string]string{"AttributeName": t.SK.Nombre, "AttributeType": t.SK.Tipo})
	}
	return map[string]any{
		"TableName": t.Nombre,
		// ACTIVE desde el primer instante: en Amazon la tabla tarda en crearse y hay que esperar.
		// Aqui no, y eso es una ventaja del doble, no una mentira: se dice en el README.
		"TableStatus":          "ACTIVE",
		"KeySchema":            esquema,
		"AttributeDefinitions": atrib,
		"CreationDateTime":     float64(t.Creada.UnixMilli()) / 1000.0,
		"ItemCount":            s.tablas.Cuenta(t.Nombre),
		"TableSizeBytes":       0,
		"TableArn":             "arn:aws:dynamodb:us-east-1:000000000000:table/" + t.Nombre,
		"BillingModeSummary":   map[string]any{"BillingMode": "PAY_PER_REQUEST"},
	}
}

func (s *ServicioDynamo) crearTabla(w http.ResponseWriter, cuerpo []byte) {
	var p peticionCrearTabla
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		s.fallo(w, "SerializationException", err.Error())
		return
	}
	tipos := map[string]string{}
	for _, a := range p.AttributeDefinitions {
		tipos[a.AttributeName] = a.AttributeType
	}
	var pk Clave
	var sk *Clave
	for _, k := range p.KeySchema {
		c := Clave{Nombre: k.AttributeName, Tipo: tipos[k.AttributeName]}
		if c.Tipo == "" {
			s.fallo(w, "ValidationException",
				"el atributo "+k.AttributeName+" esta en KeySchema pero no en AttributeDefinitions")
			return
		}
		if k.KeyType == "HASH" {
			pk = c
		} else {
			copia := c
			sk = &copia
		}
	}
	if pk.Nombre == "" {
		s.fallo(w, "ValidationException", "falta la clave de particion (KeyType HASH)")
		return
	}
	t, nuevo := s.tablas.Crear(p.TableName, pk, sk)
	if !nuevo {
		s.fallo(w, "ResourceInUseException", "Table already exists: "+p.TableName)
		return
	}
	s.registro("  -> 200 tabla %q creada", p.TableName)
	s.responder(w, map[string]any{"TableDescription": s.fichaTabla(t)})
}

func (s *ServicioDynamo) describirTabla(w http.ResponseWriter, cuerpo []byte) {
	var p struct {
		TableName string `json:"TableName"`
	}
	_ = json.Unmarshal(cuerpo, &p)
	t, ok := s.tablas.Buscar(p.TableName)
	if !ok {
		s.fallo(w, "ResourceNotFoundException", "Requested resource not found: Table: "+p.TableName)
		return
	}
	s.responder(w, map[string]any{"Table": s.fichaTabla(t)})
}

func (s *ServicioDynamo) borrarTabla(w http.ResponseWriter, cuerpo []byte) {
	var p struct {
		TableName string `json:"TableName"`
	}
	_ = json.Unmarshal(cuerpo, &p)
	t, ok := s.tablas.Borrar(p.TableName)
	if !ok {
		s.fallo(w, "ResourceNotFoundException", "Requested resource not found: Table: "+p.TableName)
		return
	}
	s.registro("  -> 200 tabla %q borrada", p.TableName)
	s.responder(w, map[string]any{"TableDescription": s.fichaTabla(t)})
}

func (s *ServicioDynamo) listarTablas(w http.ResponseWriter) {
	nombres := s.tablas.Listar()
	s.registro("  -> 200 %d tabla(s)", len(nombres))
	s.responder(w, map[string]any{"TableNames": nombres})
}

// ── elementos ────────────────────────────────────────────────────────────

type peticionElemento struct {
	TableName                string            `json:"TableName"`
	Item                     Elemento          `json:"Item"`
	Key                      Elemento          `json:"Key"`
	ConditionExpression      string            `json:"ConditionExpression"`
	ExpressionAttributeNames map[string]string `json:"ExpressionAttributeNames"`
	ReturnValues             string            `json:"ReturnValues"`
}

func (s *ServicioDynamo) ponerElemento(w http.ResponseWriter, cuerpo []byte) {
	var p peticionElemento
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		s.fallo(w, "SerializationException", err.Error())
		return
	}
	cond, err := condicionSimple(p.ConditionExpression, p.ExpressionAttributeNames)
	if err != nil {
		s.fallo(w, "ValidationException", err.Error())
		return
	}
	anterior, err := s.tablas.Poner(p.TableName, p.Item, cond)
	if s.fallaPorError(w, err) {
		return
	}
	s.registro("  -> 200 guardado en %q", p.TableName)
	res := map[string]any{}
	if p.ReturnValues == "ALL_OLD" && anterior != nil {
		res["Attributes"] = anterior
	}
	s.responder(w, res)
}

func (s *ServicioDynamo) obtenerElemento(w http.ResponseWriter, cuerpo []byte) {
	var p peticionElemento
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		s.fallo(w, "SerializationException", err.Error())
		return
	}
	e, err := s.tablas.Obtener(p.TableName, p.Key)
	if s.fallaPorError(w, err) {
		return
	}
	res := map[string]any{}
	if e != nil {
		res["Item"] = e
	}
	// ⚠️ Que no exista NO es un error en DynamoDB: es un 200 sin campo Item. Si aqui devolvieras
	// un error, el codigo del que te llama trataria "no hay" como "se rompio".
	s.registro("  -> 200 %s", map[bool]string{true: "encontrado", false: "no existe"}[e != nil])
	s.responder(w, res)
}

func (s *ServicioDynamo) borrarElemento(w http.ResponseWriter, cuerpo []byte) {
	var p peticionElemento
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		s.fallo(w, "SerializationException", err.Error())
		return
	}
	cond, err := condicionSimple(p.ConditionExpression, p.ExpressionAttributeNames)
	if err != nil {
		s.fallo(w, "ValidationException", err.Error())
		return
	}
	anterior, err := s.tablas.Quitar(p.TableName, p.Key, cond)
	if s.fallaPorError(w, err) {
		return
	}
	res := map[string]any{}
	if p.ReturnValues == "ALL_OLD" && anterior != nil {
		res["Attributes"] = anterior
	}
	s.registro("  -> 200 borrado")
	s.responder(w, res)
}

// ── Query y Scan ─────────────────────────────────────────────────────────

type peticionConsulta struct {
	TableName                 string            `json:"TableName"`
	KeyConditionExpression    string            `json:"KeyConditionExpression"`
	ExpressionAttributeNames  map[string]string `json:"ExpressionAttributeNames"`
	ExpressionAttributeValues map[string]AV     `json:"ExpressionAttributeValues"`
	ScanIndexForward          *bool             `json:"ScanIndexForward"`
	Limit                     *int              `json:"Limit"`
	Select                    string            `json:"Select"`
}

// soloContar: con `--select COUNT` DynamoDB devuelve el numero y NO los elementos. Devolverlos
// igualmente parece inofensivo hasta que alguien cuenta un millon de filas y se lleva el millon
// de filas por la red.
func soloContar(sel string) bool { return strings.EqualFold(sel, "COUNT") }

func (s *ServicioDynamo) consultar(w http.ResponseWriter, cuerpo []byte) {
	var p peticionConsulta
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		s.fallo(w, "SerializationException", err.Error())
		return
	}
	todos, t, err := s.tablas.Recorrer(p.TableName)
	if s.fallaPorError(w, err) {
		return
	}
	cond, err := analizarClave(p.KeyConditionExpression, p.ExpressionAttributeNames,
		p.ExpressionAttributeValues, t)
	if err != nil {
		s.fallo(w, "ValidationException", err.Error())
		return
	}
	var salen []Elemento
	for _, e := range todos {
		if cond.cumple(e) {
			salen = append(salen, e)
		}
	}
	// Descendente: `ScanIndexForward: false`. Los elementos ya vienen ordenados por clave de
	// orden, asi que basta con darles la vuelta.
	if p.ScanIndexForward != nil && !*p.ScanIndexForward {
		for i, j := 0, len(salen)-1; i < j; i, j = i+1, j-1 {
			salen[i], salen[j] = salen[j], salen[i]
		}
	}
	escaneados := len(salen)
	if p.Limit != nil && *p.Limit >= 0 && *p.Limit < len(salen) {
		salen = salen[:*p.Limit]
	}
	s.registro("  -> 200 Query: %d de %d", len(salen), len(todos))
	res := map[string]any{"Count": len(salen), "ScannedCount": escaneados}
	if !soloContar(p.Select) {
		res["Items"] = vaciaSiNil(salen)
	}
	s.responder(w, res)
}

func (s *ServicioDynamo) recorrer(w http.ResponseWriter, cuerpo []byte) {
	var p peticionConsulta
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		s.fallo(w, "SerializationException", err.Error())
		return
	}
	todos, _, err := s.tablas.Recorrer(p.TableName)
	if s.fallaPorError(w, err) {
		return
	}
	escaneados := len(todos)
	if p.Limit != nil && *p.Limit >= 0 && *p.Limit < len(todos) {
		todos = todos[:*p.Limit]
	}
	// ⚠️ Scan LEE LA TABLA ENTERA. Aqui da igual porque son cuatro elementos en memoria; en
	// Amazon eso es la diferencia entre unos centavos y una factura que duele.
	s.registro("  -> 200 Scan: %d elemento(s)", len(todos))
	res := map[string]any{"Count": len(todos), "ScannedCount": escaneados}
	if !soloContar(p.Select) {
		res["Items"] = vaciaSiNil(todos)
	}
	s.responder(w, res)
}

// vaciaSiNil evita que un slice nil salga como `null` en el JSON. El SDK lo aguanta, pero la CLI
// lo enseña como `null` en vez de `[]` y parece que algo se rompio.
func vaciaSiNil(e []Elemento) []Elemento {
	if e == nil {
		return []Elemento{}
	}
	return e
}

// ── las expresiones ──────────────────────────────────────────────────────
// Esta es la parte que de verdad enseña como piensa DynamoDB. `KeyConditionExpression` solo
// admite: igualdad EXACTA en la clave de particion, y opcionalmente UNA comparacion sobre la
// clave de orden. No hay OR, no hay otros atributos. Esa limitacion es el diseño.

type condicionClave struct {
	pkNombre string
	pkValor  AV
	skNombre string
	skOp     string // "" | = | < | <= | > | >= | BETWEEN | begins_with
	skA, skB AV
}

func (c *condicionClave) cumple(e Elemento) bool {
	if x, ok := compara(e[c.pkNombre], c.pkValor); !ok || x != 0 {
		return false
	}
	if c.skOp == "" {
		return true
	}
	v, hay := e[c.skNombre]
	if !hay {
		return false
	}
	if c.skOp == "begins_with" {
		s, ok := v["S"].(string)
		p, ok2 := c.skA["S"].(string)
		return ok && ok2 && strings.HasPrefix(s, p)
	}
	x, ok := compara(v, c.skA)
	if !ok {
		return false
	}
	switch c.skOp {
	case "=":
		return x == 0
	case "<":
		return x < 0
	case "<=":
		return x <= 0
	case ">":
		return x > 0
	case ">=":
		return x >= 0
	case "BETWEEN":
		y, ok2 := compara(v, c.skB)
		return ok2 && x >= 0 && y <= 0
	}
	return false
}

// trocear parte la expresion en piezas. Se hace a mano y no con strings.Split porque los
// simbolos van pegados a los nombres: "id=:v" es tan valido como "id = :v".
func trocear(expr string) []string {
	var piezas []string
	i := 0
	for i < len(expr) {
		c := expr[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n':
			i++
		case c == '(' || c == ')' || c == ',':
			piezas = append(piezas, string(c))
			i++
		case c == '=' || c == '<' || c == '>':
			if i+1 < len(expr) && expr[i+1] == '=' {
				piezas = append(piezas, expr[i:i+2])
				i += 2
			} else {
				piezas = append(piezas, string(c))
				i++
			}
		default:
			j := i
			for j < len(expr) && !strings.ContainsRune(" \t\n(),=<>", rune(expr[j])) {
				j++
			}
			piezas = append(piezas, expr[i:j])
			i = j
		}
	}
	return piezas
}

func analizarClave(expr string, nombres map[string]string, valores map[string]AV,
	t *Tabla) (*condicionClave, error) {
	if strings.TrimSpace(expr) == "" {
		return nil, errors.New("falta KeyConditionExpression")
	}
	nombre := func(s string) string {
		if strings.HasPrefix(s, "#") {
			if v, ok := nombres[s]; ok {
				return v
			}
		}
		return s
	}
	valor := func(s string) (AV, error) {
		if !strings.HasPrefix(s, ":") {
			return nil, fmt.Errorf("se esperaba un marcador de valor y llego %q", s)
		}
		v, ok := valores[s]
		if !ok {
			return nil, fmt.Errorf("el marcador %s no esta en ExpressionAttributeValues", s)
		}
		return v, nil
	}

	piezas := trocear(expr)
	c := &condicionClave{}
	i := 0
	primero := true
	for i < len(piezas) {
		if !primero {
			if !strings.EqualFold(piezas[i], "AND") {
				return nil, fmt.Errorf("solo se admite AND entre condiciones, no %q", piezas[i])
			}
			i++
		}
		primero = false
		if i >= len(piezas) {
			return nil, errors.New("la expresion termina en AND")
		}
		// begins_with(nombre, :valor)
		if strings.EqualFold(piezas[i], "begins_with") {
			if i+5 >= len(piezas) || piezas[i+1] != "(" || piezas[i+3] != "," ||
				piezas[i+5] != ")" {
				return nil, errors.New("begins_with se escribe begins_with(nombre, :valor)")
			}
			v, err := valor(piezas[i+4])
			if err != nil {
				return nil, err
			}
			c.skNombre, c.skOp, c.skA = nombre(piezas[i+2]), "begins_with", v
			i += 6
			continue
		}
		if i+2 >= len(piezas) {
			return nil, fmt.Errorf("condicion incompleta cerca de %q", piezas[i])
		}
		n, op := nombre(piezas[i]), piezas[i+1]
		if strings.EqualFold(op, "BETWEEN") {
			if i+4 >= len(piezas) || !strings.EqualFold(piezas[i+3], "AND") {
				return nil, errors.New("BETWEEN se escribe: nombre BETWEEN :a AND :b")
			}
			a, err := valor(piezas[i+2])
			if err != nil {
				return nil, err
			}
			b, err := valor(piezas[i+4])
			if err != nil {
				return nil, err
			}
			c.skNombre, c.skOp, c.skA, c.skB = n, "BETWEEN", a, b
			i += 5
			continue
		}
		v, err := valor(piezas[i+2])
		if err != nil {
			return nil, err
		}
		if n == t.PK.Nombre {
			if op != "=" {
				// ⛔ Esta es LA regla de DynamoDB, y por eso el error es explicito: sobre la
				// clave de particion solo hay igualdad. No hay "usuarios que empiecen por a".
				return nil, errors.New("sobre la clave de particion solo se admite =")
			}
			c.pkNombre, c.pkValor = n, v
		} else {
			if !strings.Contains("=,<,<=,>,>=", op) {
				return nil, fmt.Errorf("operador no admitido: %q", op)
			}
			c.skNombre, c.skOp, c.skA = n, op, v
		}
		i += 3
	}
	if c.pkNombre == "" {
		return nil, errors.New("la condicion tiene que fijar la clave de particion con =")
	}
	return c, nil
}

// condicionSimple entiende `attribute_exists(x)` y `attribute_not_exists(x)`, que es lo que hace
// falta para la escritura condicional: la unica herramienta de concurrencia que da DynamoDB.
//
//	"escribe esto SOLO si nadie lo escribio antes"  ->  attribute_not_exists(pk)
//
// Devuelve nil cuando no hay condicion: quien llama interpreta nil como "adelante".
func condicionSimple(expr string, nombres map[string]string) (func(Elemento) bool, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, nil
	}
	piezas := trocear(expr)
	if len(piezas) != 4 || piezas[1] != "(" || piezas[3] != ")" {
		return nil, fmt.Errorf("nubecita solo entiende attribute_exists(x) y "+
			"attribute_not_exists(x) en ConditionExpression, no %q", expr)
	}
	atributo := piezas[2]
	if v, ok := nombres[atributo]; ok {
		atributo = v
	}
	switch strings.ToLower(piezas[0]) {
	case "attribute_not_exists":
		return func(anterior Elemento) bool {
			if anterior == nil {
				return true
			}
			_, hay := anterior[atributo]
			return !hay
		}, nil
	case "attribute_exists":
		return func(anterior Elemento) bool {
			if anterior == nil {
				return false
			}
			_, hay := anterior[atributo]
			return hay
		}, nil
	}
	return nil, fmt.Errorf("funcion no admitida en ConditionExpression: %q", piezas[0])
}
