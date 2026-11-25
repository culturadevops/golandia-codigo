package main

// Los permisos. Episodio 10, y es la otra mitad del episodio 4.
//
// Aquel terminaba diciendo, con todas las letras: «comprobar QUIEN eres no es decidir QUE puedes
// hacer». Esto es lo segundo.
//
//	AUTENTICAR   ¿eres quien dices? -> la firma. Episodio 4.
//	AUTORIZAR    ¿puedes hacer esto? -> las politicas. Este.
//
// Y son dos sistemas separados de verdad, no dos partes del mismo: la firma se comprueba con un
// secreto que solo saben dos, y el permiso se comprueba con unas reglas escritas que puede leer
// cualquiera. Una firma valida y cero permisos es una situacion perfectamente normal.
//
// 🥇 LAS TRES REGLAS DE LA EVALUACION, y en este orden:
//
//	1. por defecto NO. Sin un `Allow` que te toque, no pasas.
//	2. un `Allow` explicito te deja pasar.
//	3. un `Deny` explicito GANA SIEMPRE, aunque haya mil `Allow`.
//
// De la tercera sale el patron que usa todo el mundo sin saber por que: pones un `Deny` general
// como red de seguridad, porque NADA lo puede levantar. Ni un administrador, ni una politica
// nueva, ni un despiste.
//
// 🥇 Y el comodin. `s3:*` no quiere decir «S3»: quiere decir «todo lo que empiece por `s3:`», y
// eso incluye `s3:DeleteBucket`. El asterisco no sabe de servicios, sabe de letras.

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Efecto es lo que dice una sentencia: dejar o negar.
type Efecto string

const (
	Permitir Efecto = "Allow"
	Negar    Efecto = "Deny"
)

// Sentencia es una regla. `Action` y `Resource` admiten comodines.
type Sentencia struct {
	Sid      string   `json:"Sid,omitempty"`
	Effect   Efecto   `json:"Effect"`
	Action   []string `json:"Action"`
	Resource []string `json:"Resource"`
}

// Politica es una lista de sentencias con su version.
type Politica struct {
	Version   string      `json:"Version"`
	Statement []Sentencia `json:"Statement"`
}

// ⚠️ En una politica de verdad, `Action` y `Resource` pueden ser una CADENA o una LISTA. El JSON
// de la biblioteca estandar no acepta las dos formas en el mismo campo, asi que hay que
// escribirlo a mano. Es feo y es obligatorio: si no, la mitad de las politicas del mundo real no
// se pueden leer.
func (s *Sentencia) UnmarshalJSON(crudo []byte) error {
	var bruto struct {
		Sid      string          `json:"Sid"`
		Effect   Efecto          `json:"Effect"`
		Action   json.RawMessage `json:"Action"`
		Resource json.RawMessage `json:"Resource"`
	}
	if err := json.Unmarshal(crudo, &bruto); err != nil {
		return err
	}
	s.Sid, s.Effect = bruto.Sid, bruto.Effect
	var err error
	if s.Action, err = unaOVarias(bruto.Action); err != nil {
		return fmt.Errorf("Action: %w", err)
	}
	if s.Resource, err = unaOVarias(bruto.Resource); err != nil {
		return fmt.Errorf("Resource: %w", err)
	}
	if s.Effect != Permitir && s.Effect != Negar {
		return fmt.Errorf("Effect es %q y solo entiendo Allow o Deny", s.Effect)
	}
	return nil
}

func unaOVarias(crudo json.RawMessage) ([]string, error) {
	if len(crudo) == 0 {
		return nil, nil
	}
	var lista []string
	if err := json.Unmarshal(crudo, &lista); err == nil {
		return lista, nil
	}
	var una string
	if err := json.Unmarshal(crudo, &una); err != nil {
		return nil, err
	}
	return []string{una}, nil
}

// ── el comodin ───────────────────────────────────────────────────────────

// encaja compara un patron con comodines contra un texto.
//
// ⚠️ Los comodines de una politica NO son expresiones regulares: son solo `*` (cualquier cosa,
// incluida ninguna) y `?` (un caracter). Un punto es un punto. Confundirlos es como escribir una
// politica en otro idioma y creer que dice lo que quieres.
func encaja(patron, texto string) bool {
	// Sin comodines, es una comparacion normal.
	if !strings.ContainsAny(patron, "*?") {
		return patron == texto
	}
	return comodines([]rune(patron), []rune(texto))
}

func comodines(p, t []rune) bool {
	for len(p) > 0 {
		switch p[0] {
		case '*':
			// Se prueba a que el asterisco se coma 0, 1, 2... caracteres.
			for i := 0; i <= len(t); i++ {
				if comodines(p[1:], t[i:]) {
					return true
				}
			}
			return false
		case '?':
			if len(t) == 0 {
				return false
			}
			p, t = p[1:], t[1:]
		default:
			if len(t) == 0 || p[0] != t[0] {
				return false
			}
			p, t = p[1:], t[1:]
		}
	}
	return len(t) == 0
}

func alguno(patrones []string, texto string) bool {
	for _, p := range patrones {
		if encaja(p, texto) {
			return true
		}
	}
	return false
}

// ── la evaluacion ────────────────────────────────────────────────────────

// Veredicto dice si se puede y por que. El «por que» es la mitad del valor: un error de permisos
// que no dice que falta es un error inutil.
type Veredicto struct {
	Permitido bool
	Motivo    string
	Sid       string // la sentencia que decidio, si hubo alguna
}

// Evaluar aplica las tres reglas, en orden.
//
// 🥇 Fijate en que el `Deny` se busca ENTERO antes de mirar ningun `Allow`. No es una
// optimizacion al reves: es la unica forma de que «un Deny gana siempre» sea cierto. Si se
// devolviera al primer `Allow` que encaja, un `Deny` escrito mas abajo no serviria de nada, y
// tendrias una politica que hace cosas distintas segun el ORDEN en que la escribiste.
func Evaluar(ps []Politica, accion, recurso string) Veredicto {
	// 1. ¿Hay algun Deny que toque? Se mira TODO antes de nada.
	for _, p := range ps {
		for _, s := range p.Statement {
			if s.Effect != Negar {
				continue
			}
			if alguno(s.Action, accion) && alguno(s.Resource, recurso) {
				return Veredicto{false,
					fmt.Sprintf("un Deny explicito prohibe %s sobre %s", accion, recurso), s.Sid}
			}
		}
	}
	// 2. ¿Hay algun Allow que toque?
	for _, p := range ps {
		for _, s := range p.Statement {
			if s.Effect != Permitir {
				continue
			}
			if alguno(s.Action, accion) && alguno(s.Resource, recurso) {
				return Veredicto{true, "", s.Sid}
			}
		}
	}
	// 3. Y si no, no. Por defecto NO.
	return Veredicto{false,
		fmt.Sprintf("ninguna politica permite %s sobre %s", accion, recurso), ""}
}

// ── quien tiene que permiso ──────────────────────────────────────────────

// Identidad es una clave de acceso con sus politicas.
type Identidad struct {
	Acceso    string
	Nombre    string
	Politicas []Politica
}

type Permisos struct {
	mu       sync.Mutex
	porClave map[string]*Identidad
	registro func(formato string, args ...any)
}

func NuevosPermisos(registro func(string, ...any)) *Permisos {
	if registro == nil {
		registro = func(string, ...any) {}
	}
	return &Permisos{porClave: map[string]*Identidad{}, registro: registro}
}

func (p *Permisos) Poner(acceso, nombre string, ps []Politica) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.porClave[acceso] = &Identidad{Acceso: acceso, Nombre: nombre, Politicas: ps}
}

func (p *Permisos) Buscar(acceso string) (*Identidad, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	i, ok := p.porClave[acceso]
	return i, ok
}

func (p *Permisos) Identidades() []*Identidad {
	p.mu.Lock()
	defer p.mu.Unlock()
	fuera := make([]*Identidad, 0, len(p.porClave))
	for _, i := range p.porClave {
		fuera = append(fuera, i)
	}
	sort.Slice(fuera, func(a, b int) bool { return fuera[a].Acceso < fuera[b].Acceso })
	return fuera
}

// Cargada es una identidad tal como viene del fichero: quien es, con que se identifica, y que
// puede hacer.
//
// 🥇 Los tres campos juntos, y eso NO es pereza: un usuario de IAM es exactamente eso. La clave
// dice quien eres, el secreto lo demuestra, y la politica dice que puedes. Separarlos en dos
// ficheros seria separar dos cosas que siempre se leen a la vez.
type Cargada struct {
	Nombre   string   `json:"nombre"`
	Secreto  string   `json:"secreto,omitempty"`
	Politica Politica `json:"politica"`
}

// CargarPoliticas lee un fichero con la forma:
//
//	{"prueba": {"nombre":"lector", "secreto":"...", "politica": {...}}}
func CargarPoliticas(crudo []byte) (map[string]Cargada, error) {
	var fuera map[string]Cargada
	dec := json.NewDecoder(strings.NewReader(string(crudo)))
	// ⚠️ Sin esto, una errata en el nombre de un campo se traga en silencio. Y una politica que
	// se carga a medias es peor que una que no se carga: no protege y parece que si.
	dec.DisallowUnknownFields()
	if err := dec.Decode(&fuera); err != nil {
		return nil, fmt.Errorf("el fichero de permisos no se entiende: %w", err)
	}
	return fuera, nil
}
