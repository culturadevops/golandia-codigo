package main

// Las pruebas del almacen de registros. Aqui no hay HTTP: son las REGLAS, que es donde estan
// las sorpresas de este servicio.
//
//	go test ./... -count=1 -race -run Registros

import (
	"errors"
	"testing"
	"time"
)

func almacenConLineas(t *testing.T, grupo, flujo string, base time.Time, n int) *Registros {
	t.Helper()
	r := NuevosRegistros()
	if err := r.CrearGrupo(grupo); err != nil {
		t.Fatalf("crear grupo: %v", err)
	}
	if err := r.CrearFlujo(grupo, flujo); err != nil {
		t.Fatalf("crear flujo: %v", err)
	}
	apuntes := make([]Apunte, 0, n)
	for i := 0; i < n; i++ {
		apuntes = append(apuntes, Apunte{
			Marca:   base.Add(time.Duration(i) * time.Second).UnixMilli(),
			Mensaje: "linea " + pad12(int64(i)),
		})
	}
	if _, _, err := r.Poner(grupo, flujo, "", apuntes, base.Add(time.Minute)); err != nil {
		t.Fatalf("poner: %v", err)
	}
	return r
}

// 🥇 La diferencia entre un registro y una cola, escrita como prueba: leer no saca nada. Dos
// lectores ven lo mismo, y el segundo no se encuentra el registro vacio.
func TestRegistrosLeerNoConsume(t *testing.T) {
	base := time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC)
	r := almacenConLineas(t, "/aws/lambda/saludar", "flujo-1", base, 5)

	uno, err := r.Leer("/aws/lambda/saludar", "flujo-1", nil, nil, 0, "", true)
	if err != nil {
		t.Fatalf("primera lectura: %v", err)
	}
	dos, err := r.Leer("/aws/lambda/saludar", "flujo-1", nil, nil, 0, "", true)
	if err != nil {
		t.Fatalf("segunda lectura: %v", err)
	}
	if len(uno.Apuntes) != 5 || len(dos.Apuntes) != 5 {
		t.Fatalf("un registro no se consume al leerlo: %d y %d", len(uno.Apuntes), len(dos.Apuntes))
	}
	if uno.Apuntes[0].Mensaje != dos.Apuntes[0].Mensaje {
		t.Fatalf("los dos lectores tendrian que ver lo mismo")
	}
}

// 🥇 La trampa del bucle infinito: el token NUNCA viene vacio, ni cuando ya no hay mas. Un
// «mientras haya token» no termina jamas; se para comparando el token con el que mandaste.
func TestRegistrosElTokenNoSeAcabaNunca(t *testing.T) {
	base := time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC)
	r := almacenConLineas(t, "g", "f", base, 5)

	pag, err := r.Leer("g", "f", nil, nil, 100, "", true)
	if err != nil {
		t.Fatalf("leer: %v", err)
	}
	if pag.Adelante == "" {
		t.Fatal("nextForwardToken tiene que venir siempre, tambien cuando no queda nada")
	}

	// Y ahora el bucle bien hecho: se para cuando el token deja de moverse.
	visto, token, vueltas := 0, "", 0
	for {
		vueltas++
		if vueltas > 20 {
			t.Fatal("el bucle no termina: eso es exactamente el fallo que enseña esta prueba")
		}
		p, err := r.Leer("g", "f", nil, nil, 2, token, true)
		if err != nil {
			t.Fatalf("leer: %v", err)
		}
		visto += len(p.Apuntes)
		if p.Adelante == token {
			break
		}
		token = p.Adelante
	}
	if visto != 5 {
		t.Fatalf("paginando salieron %d apuntes y hay 5", visto)
	}
}

// El lote tiene que venir ordenado por marca, y lo comprueba el servidor: si no, el registro
// dejaria de estar ordenado y los tokens no querrian decir nada.
func TestRegistrosLoteFueraDeOrden(t *testing.T) {
	base := time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC)
	r := almacenConLineas(t, "g", "f", base, 0)

	_, _, err := r.Poner("g", "f", "", []Apunte{
		{Marca: base.Add(2 * time.Second).UnixMilli(), Mensaje: "segunda"},
		{Marca: base.UnixMilli(), Mensaje: "primera"},
	}, base.Add(time.Minute))
	if !errors.Is(err, errFueraDeOrden) {
		t.Fatalf("un lote desordenado tiene que fallar, y fallo con: %v", err)
	}
}

// ⚠️ La ventana: lo que cae fuera se tira, y la llamada dice que si. Es la explicacion de «puse
// un log y no aparece» cuando el contenedor tiene la hora mal.
func TestRegistrosVentanaDeTiempo(t *testing.T) {
	ahora := time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC)
	r := almacenConLineas(t, "g", "f", ahora, 0)

	_, rech, err := r.Poner("g", "f", "", []Apunte{
		{Marca: ahora.Add(-30 * 24 * time.Hour).UnixMilli(), Mensaje: "de hace un mes"},
		{Marca: ahora.UnixMilli(), Mensaje: "de ahora"},
		{Marca: ahora.Add(5 * time.Hour).UnixMilli(), Mensaje: "de dentro de cinco horas"},
	}, ahora)
	if err != nil {
		t.Fatalf("esto NO es un error, es un rechazo: %v", err)
	}
	if rech.Viejos != 1 || rech.Nuevos != 1 {
		t.Fatalf("tenian que rechazarse uno viejo y uno futuro, y salio %+v", rech)
	}
	pag, _ := r.Leer("g", "f", nil, nil, 0, "", true)
	if len(pag.Apuntes) != 1 || pag.Apuntes[0].Mensaje != "de ahora" {
		t.Fatalf("solo tendria que haber entrado el del medio: %+v", pag.Apuntes)
	}
}

// 🥇 El sequenceToken: se acepta vacio (como hoy) y se rechaza equivocado (como antes). Y lo
// bueno es que el error trae DENTRO el token que tocaba.
func TestRegistrosSequenceToken(t *testing.T) {
	ahora := time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC)
	r := almacenConLineas(t, "g", "f", ahora, 0)

	bueno, _, err := r.Poner("g", "f", "", []Apunte{{Marca: ahora.UnixMilli(), Mensaje: "uno"}}, ahora)
	if err != nil {
		t.Fatalf("sin token tiene que dejar: %v", err)
	}
	esperado, _, err := r.Poner("g", "f", "no-es-este",
		[]Apunte{{Marca: ahora.UnixMilli(), Mensaje: "dos"}}, ahora)
	if !errors.Is(err, errTokenEquivocado) {
		t.Fatalf("con el token equivocado tiene que fallar, y fallo con: %v", err)
	}
	if esperado != bueno {
		t.Fatalf("el error tiene que traer el token bueno (%q) y trajo %q", bueno, esperado)
	}
	if _, _, err := r.Poner("g", "f", bueno,
		[]Apunte{{Marca: ahora.UnixMilli(), Mensaje: "dos"}}, ahora); err != nil {
		t.Fatalf("con el token bueno tiene que dejar: %v", err)
	}
}

// ⚠️ endTime es EXCLUSIVO. Pedir «hasta las 10:00:02» no trae el apunte de las 10:00:02.
func TestRegistrosElFinalEsExclusivo(t *testing.T) {
	base := time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC)
	r := almacenConLineas(t, "g", "f", base, 5)

	hasta := base.Add(2 * time.Second).UnixMilli()
	pag, err := r.Leer("g", "f", nil, &hasta, 0, "", true)
	if err != nil {
		t.Fatalf("leer: %v", err)
	}
	if len(pag.Apuntes) != 2 {
		t.Fatalf("con endTime exclusivo tendrian que salir 2 y salieron %d", len(pag.Apuntes))
	}
}

// Sin token y sin startFromHead, lo que se devuelve es el FINAL. De un registro, lo que quieres
// ver casi siempre es lo ultimo.
func TestRegistrosSinTokenDaElFinal(t *testing.T) {
	base := time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC)
	r := almacenConLineas(t, "g", "f", base, 10)

	pag, err := r.Leer("g", "f", nil, nil, 3, "", false)
	if err != nil {
		t.Fatalf("leer: %v", err)
	}
	if len(pag.Apuntes) != 3 || pag.Apuntes[2].Mensaje != "linea "+pad12(9) {
		t.Fatalf("tendrian que ser las tres ultimas: %+v", pag.Apuntes)
	}
}

// ⚠️ Cada flujo esta ordenado por dentro; el conjunto no lo esta hasta que alguien lo ordena.
// Es la misma diferencia que hay entre «cada particion esta ordenada» y «el registro lo esta».
func TestRegistrosFiltrarOrdenaEntreFlujos(t *testing.T) {
	base := time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC)
	r := NuevosRegistros()
	_ = r.CrearGrupo("g")

	r.Anexar("g", "b", base.Add(1*time.Second).UnixMilli(), "hola de b")
	r.Anexar("g", "a", base.Add(0*time.Second).UnixMilli(), "hola de a")
	r.Anexar("g", "b", base.Add(3*time.Second).UnixMilli(), "adios de b")
	r.Anexar("g", "a", base.Add(2*time.Second).UnixMilli(), "hola otra vez de a")

	filas, err := r.Filtrar("g", "", "hola", nil, nil, 0)
	if err != nil {
		t.Fatalf("filtrar: %v", err)
	}
	if len(filas) != 3 {
		t.Fatalf("tres lineas dicen hola y salieron %d", len(filas))
	}
	for i := 1; i < len(filas); i++ {
		if filas[i].Apunte.Marca < filas[i-1].Apunte.Marca {
			t.Fatalf("el resultado de varios flujos tiene que salir ordenado: %+v", filas)
		}
	}
	if filas[0].Flujo != "a" {
		t.Fatalf("cada fila tiene que decir de que flujo viene, y la primera dice %q", filas[0].Flujo)
	}
}

// Lambda no crea sus grupos a mano: aparecen porque alguien escribio.
func TestRegistrosAsegurarCreaLoQueFalte(t *testing.T) {
	r := NuevosRegistros()
	r.Anexar(grupoDeFuncion("saludar"), "flujo", time.Now().UnixMilli(), "hola")

	grupos := r.ListarGrupos("")
	if len(grupos) != 1 || grupos[0].Nombre != "/aws/lambda/saludar" {
		t.Fatalf("el grupo tenia que crearse solo: %+v", grupos)
	}
	flujos, err := r.ListarFlujos("/aws/lambda/saludar", "")
	if err != nil || len(flujos) != 1 || flujos[0].Apuntes != 1 {
		t.Fatalf("el flujo tenia que crearse solo y tener una linea: %+v (%v)", flujos, err)
	}
}

func TestRegistrosLoQueNoExiste(t *testing.T) {
	r := NuevosRegistros()
	if _, err := r.Leer("no-esta", "f", nil, nil, 0, "", true); !errors.Is(err, errNoExisteGrupo) {
		t.Fatalf("un grupo que no esta tiene que decirlo: %v", err)
	}
	_ = r.CrearGrupo("g")
	if _, err := r.Leer("g", "no-esta", nil, nil, 0, "", true); !errors.Is(err, errNoExisteFlujo) {
		t.Fatalf("un flujo que no esta tiene que decirlo: %v", err)
	}
	if err := r.CrearGrupo("g"); !errors.Is(err, errYaExisteGrupo) {
		t.Fatalf("crear dos veces el mismo grupo tiene que fallar: %v", err)
	}
}
