package main

// Las pruebas de la salida de los contenedores. Como en el episodio 7, aqui no hay Docker: hay
// un demonio de mentira que habla su protocolo, porque su protocolo es HTTP.
//
//	go test ./... -count=1 -race -run Bitacora

import (
	"context"
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// marco arma una cabecera de Docker con su payload detras. Es literalmente lo que manda el
// demonio: ocho bytes y luego los datos.
func marco(flujo byte, texto string) []byte {
	c := make([]byte, 8)
	c[0] = flujo
	binary.BigEndian.PutUint32(c[4:], uint32(len(texto)))
	return append(c, texto...)
}

// 🥇 Un marco NO es una linea. Puede traer media, tres, o una partida por la mitad de una
// palabra. Esta prueba manda las tres formas a la vez.
func TestBitacoraDemultiplexaMarcos(t *testing.T) {
	var crudo []byte
	crudo = append(crudo, marco(1, "hola\nque tal\n")...)     // dos lineas en un marco
	crudo = append(crudo, marco(1, "una linea par")...)       // media linea...
	crudo = append(crudo, marco(1, "tida por la mitad\n")...) // ...y la otra mitad
	crudo = append(crudo, marco(2, "esto es un error\n")...)
	crudo = append(crudo, marco(1, "sin salto al final")...)

	var salieron []string
	var flujos []byte
	err := demultiplexar(strings.NewReader(string(crudo)), func(f byte, l []byte) {
		flujos = append(flujos, f)
		salieron = append(salieron, string(l))
	})
	if err != nil {
		t.Fatalf("demultiplexar: %v", err)
	}
	quiero := []string{"hola", "que tal", "una linea partida por la mitad",
		"esto es un error", "sin salto al final"}
	if len(salieron) != len(quiero) {
		t.Fatalf("salieron %d lineas y esperaba %d: %q", len(salieron), len(quiero), salieron)
	}
	for i := range quiero {
		if salieron[i] != quiero[i] {
			t.Fatalf("linea %d: %q y esperaba %q", i, salieron[i], quiero[i])
		}
	}
	// Y el flujo del que vino no se pierde por el camino: la cuarta es stderr.
	if flujos[3] != 2 {
		t.Fatalf("la linea de error tenia que venir del flujo 2 y vino del %d", flujos[3])
	}
}

// ⚠️ El mismo endpoint devuelve dos formatos distintos segun como se creo el contenedor. Con
// `Tty: true` no hay marcos, y hay que darse cuenta mirando los primeros ocho bytes.
func TestBitacoraSinMarcosCuandoHayTerminal(t *testing.T) {
	crudo := "primera\r\nsegunda\ntercera sin salto"
	var salieron []string
	if err := demultiplexar(strings.NewReader(crudo), func(_ byte, l []byte) {
		salieron = append(salieron, string(l))
	}); err != nil {
		t.Fatalf("demultiplexar: %v", err)
	}
	quiero := []string{"primera", "segunda", "tercera sin salto"}
	for i := range quiero {
		if i >= len(salieron) || salieron[i] != quiero[i] {
			t.Fatalf("sin marcos tendria que salir %q y salio %q", quiero, salieron)
		}
	}
}

// La hora la pone Docker, no nosotros: viene delante de cada linea y hay que separarla.
func TestBitacoraLaHoraLaPoneDocker(t *testing.T) {
	marca, mensaje := partirMarca("2026-08-25T10:00:00.500000000Z hola mundo")
	quiero := time.Date(2026, 8, 25, 10, 0, 0, 500000000, time.UTC).UnixMilli()
	if marca != quiero {
		t.Fatalf("la marca salio %d y tenia que ser %d", marca, quiero)
	}
	if mensaje != "hola mundo" {
		t.Fatalf("el mensaje salio %q", mensaje)
	}
	// Y si no viene con hora, no se rompe: se pone la de ahora y se deja la linea entera.
	_, m2 := partirMarca("una linea sin hora")
	if m2 != "una linea sin hora" {
		t.Fatalf("sin hora la linea tiene que quedarse entera, y quedo %q", m2)
	}
}

// El camino completo: el demonio manda marcos, y las lineas acaban en su grupo de registro.
func TestBitacoraSeguirLlenaElRegistro(t *testing.T) {
	var pedido string
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{version}/containers/{id}/logs", func(w http.ResponseWriter, r *http.Request) {
		pedido = r.URL.RawQuery
		_, _ = w.Write(marco(1, "2026-08-25T10:00:00Z arrancando\n"))
		_, _ = w.Write(marco(2, "2026-08-25T10:00:01Z algo fue mal\n"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	reg := NuevosRegistros()
	lan := NuevoLanzador(srv.Client(), srv.URL, "demo", "nubecita:4566", 4, 0, nil).
		ConRegistros(reg)
	reg.Asegurar("/aws/lambda/saludar", "flujo-1")
	lan.seguir(context.Background(), "abc123", "/aws/lambda/saludar", "flujo-1")

	// follow=1 y timestamps=1 no son opcionales: sin el primero esto seria un sondeo, y sin el
	// segundo la hora seria la nuestra.
	for _, q := range []string{"follow=1", "timestamps=1", "stdout=1", "stderr=1"} {
		if !strings.Contains(pedido, q) {
			t.Fatalf("a docker hay que pedirle %s, y se le pidio %q", q, pedido)
		}
	}

	pag, err := reg.Leer("/aws/lambda/saludar", "flujo-1", nil, nil, 0, "", true)
	if err != nil {
		t.Fatalf("leer: %v", err)
	}
	if len(pag.Apuntes) != 2 {
		t.Fatalf("tenian que llegar dos lineas y llegaron %d: %+v", len(pag.Apuntes), pag.Apuntes)
	}
	if pag.Apuntes[0].Mensaje != "arrancando" || pag.Apuntes[1].Mensaje != "algo fue mal" {
		t.Fatalf("las lineas llegaron mal: %+v", pag.Apuntes)
	}
	quiero := time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC).UnixMilli()
	if pag.Apuntes[0].Marca != quiero {
		t.Fatalf("la hora tiene que ser la de docker (%d) y es %d", quiero, pag.Apuntes[0].Marca)
	}
}

// Que Docker no quiera darnos los registros no puede tumbar nada: se anota y se sigue.
func TestBitacoraSeguirAguantaUnNo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"message":"No such container"}`))
	}))
	defer srv.Close()

	var dijo []string
	reg := NuevosRegistros()
	lan := NuevoLanzador(srv.Client(), srv.URL, "demo", "nubecita:4566", 4, 0,
		func(f string, a ...any) { dijo = append(dijo, f) }).ConRegistros(reg)
	lan.seguir(context.Background(), "abc123", "g", "f")

	if len(dijo) == 0 {
		t.Fatal("si docker dice que no, por lo menos hay que contarlo")
	}
	if len(reg.ListarGrupos("")) != 0 {
		t.Fatal("y no hay que inventarse ningun registro")
	}
}
