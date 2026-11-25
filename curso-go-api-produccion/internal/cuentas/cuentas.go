// Package cuentas: usuarios, claves y sesiones.
//
// Por qué sesiones en Postgres y no JWT: esta API ya consulta Postgres en cada petición, así
// que buscar la sesión cuesta una consulta por clave primaria. A cambio, cerrar sesión
// funciona de verdad (se borra la fila) y no hay firma, algoritmo ni clave que rotar.
// Un JWT no se puede revocar antes de que venza sin volver a tener una tabla.
package cuentas

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"tareas/internal/web"
)

// DuracionSesion es lo que vale un token desde que se crea.
const DuracionSesion = 24 * time.Hour

// Servicio guarda usuarios y sesiones en Postgres.
type Servicio struct {
	bd       *pgxpool.Pool
	costo    int
	ficticio []byte // hash de una clave que nadie tiene: ver IniciarSesion
}

// Nuevo prepara el servicio. costo es el de bcrypt (12 en producción, 10 en los tests).
func Nuevo(bd *pgxpool.Pool, costo int) *Servicio {
	f, _ := bcrypt.GenerateFromPassword([]byte("clave-que-nadie-tiene"), costo)
	return &Servicio{bd: bd, costo: costo, ficticio: f}
}

// Normalizar es la forma única de un correo: sin espacios y en minúsculas.
func Normalizar(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

// Registrar crea un usuario y devuelve su id.
func (s *Servicio) Registrar(ctx context.Context, email, clave string) (int64, error) {
	email = Normalizar(email)
	campos := map[string]string{}
	if i := strings.IndexByte(email, '@'); i < 1 || i == len(email)-1 || len(email) > 254 {
		campos["email"] = "no parece un correo"
	}
	if len(clave) < 10 || len(clave) > 72 {
		campos["clave"] = "de 10 a 72 bytes" // 72: lo máximo que bcrypt usa
	}
	if err := web.Validacion(campos); err != nil {
		return 0, err
	}
	h, err := bcrypt.GenerateFromPassword([]byte(clave), s.costo)
	if err != nil {
		return 0, err
	}
	var id int64
	err = s.bd.QueryRow(ctx,
		`INSERT INTO usuarios (email, clave_hash) VALUES ($1, $2) RETURNING id`, email, h).Scan(&id)
	if pgErr := (*pgconn.PgError)(nil); errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return 0, web.Conflicto("ese correo ya está registrado")
	}
	return id, err
}

var errCredenciales = web.NoAutenticado("correo o clave incorrectos")

// IniciarSesion comprueba la clave y devuelve un token nuevo.
func (s *Servicio) IniciarSesion(ctx context.Context, email, clave string) (string, time.Time, error) {
	var id int64
	var h []byte
	err := s.bd.QueryRow(ctx, `SELECT id, clave_hash FROM usuarios WHERE lower(email) = $1`,
		Normalizar(email)).Scan(&id, &h)
	if errors.Is(err, pgx.ErrNoRows) {
		// Se compara igual, contra un hash ficticio: si no, «ese correo no existe» contesta
		// en 1 ms y «clave mal» en 200 ms, y el tiempo dice qué correos están registrados.
		_ = bcrypt.CompareHashAndPassword(s.ficticio, []byte(clave))
		return "", time.Time{}, errCredenciales
	}
	if err != nil {
		return "", time.Time{}, err
	}
	if bcrypt.CompareHashAndPassword(h, []byte(clave)) != nil {
		return "", time.Time{}, errCredenciales
	}
	return s.crearSesion(ctx, id)
}

func (s *Servicio) crearSesion(ctx context.Context, usuario int64) (string, time.Time, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", time.Time{}, err
	}
	token := base64.RawURLEncoding.EncodeToString(b)
	vence := time.Now().Add(DuracionSesion).UTC().Truncate(time.Second)
	_, err := s.bd.Exec(ctx, `INSERT INTO sesiones (token_hash, usuario_id, vence) VALUES ($1, $2, $3)`,
		huella(token), usuario, vence)
	return token, vence, err
}

func huella(token string) []byte {
	h := sha256.Sum256([]byte(token))
	return h[:]
}

// Usuario devuelve el dueño de un token vigente.
func (s *Servicio) Usuario(ctx context.Context, token string) (int64, error) {
	var id int64
	err := s.bd.QueryRow(ctx, `SELECT usuario_id FROM sesiones WHERE token_hash = $1 AND vence > now()`,
		huella(token)).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, web.NoAutenticado("sesión inválida o vencida")
	}
	return id, err
}

// Cerrar borra la sesión: el token deja de servir en ese mismo instante.
func (s *Servicio) Cerrar(ctx context.Context, token string) error {
	_, err := s.bd.Exec(ctx, `DELETE FROM sesiones WHERE token_hash = $1`, huella(token))
	return err
}
