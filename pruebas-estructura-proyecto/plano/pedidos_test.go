package main

import "testing"

func TestResumen(t *testing.T) {
	if got := NuevoPedido("mouse", 2).Resumen(); got != "2 x mouse" {
		t.Fatalf("got %q", got)
	}
}
