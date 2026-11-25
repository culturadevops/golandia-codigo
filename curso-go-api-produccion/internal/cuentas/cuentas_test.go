package cuentas_test

import (
	"context"
	"errors"
	"testing"

	"tareas/internal/cuentas"
	"tareas/internal/pruebabd"
	"tareas/internal/web"
)

func codigo(err error) string {
	var e *web.Error
	if errors.As(err, &e) {
		return e.Codigo
	}
	if err != nil {
		return "otro: " + err.Error()
	}
	return ""
}

func TestCuentas(t *testing.T) {
	c := cuentas.Nuevo(pruebabd.Abrir(t), 10)
	ctx := context.Background()

	id, err := c.Registrar(ctx, "  Ana@Ejemplo.com ", "una clave larga")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Registrar(ctx, "ana@ejemplo.COM", "otra clave larga"); codigo(err) != "conflicto" {
		t.Fatalf("mismo correo con otras mayúsculas: %v", err)
	}
	if _, err := c.Registrar(ctx, "x", "corta"); codigo(err) != "validacion" {
		t.Fatalf("validación: %v", err)
	}
	if _, _, err := c.IniciarSesion(ctx, "ana@ejemplo.com", "clave mal puesta"); codigo(err) != "no_autenticado" {
		t.Fatalf("clave mal: %v", err)
	}
	if _, _, err := c.IniciarSesion(ctx, "nadie@ejemplo.com", "una clave larga"); codigo(err) != "no_autenticado" {
		t.Fatalf("correo que no existe: %v", err)
	}

	token, _, err := c.IniciarSesion(ctx, "ANA@ejemplo.com", "una clave larga")
	if err != nil {
		t.Fatal(err)
	}
	if u, err := c.Usuario(ctx, token); err != nil || u != id {
		t.Fatalf("usuario del token: %d %v", u, err)
	}
	if err := c.Cerrar(ctx, token); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Usuario(ctx, token); codigo(err) != "no_autenticado" {
		t.Fatalf("el token sigue valiendo tras cerrar sesión: %v", err)
	}
}
