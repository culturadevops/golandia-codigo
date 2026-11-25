package config

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func entorno(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestCargarValida(t *testing.T) {
	c, err := Cargar(entorno(map[string]string{
		"PUERTO": "8080", "BD_URL": "postgres://app:secreta@bd:5432/tareas", "APAGADO_MAX": "20s",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if c.ApagadoMax != 20*time.Second || c.TiempoPeticion != 10*time.Second || c.CostoBcrypt != 12 {
		t.Fatalf("valores inesperados: %+v", c)
	}
}

func TestCargarDiceTodosLosErrores(t *testing.T) {
	_, err := Cargar(entorno(map[string]string{
		"PUERTO": "80a", "TIEMPO_PETICION": "30", "NIVEL_LOG": "mucho",
	}))
	if err == nil {
		t.Fatal("esperaba error")
	}
	for _, v := range []string{"PUERTO", "BD_URL", "TIEMPO_PETICION", "NIVEL_LOG"} {
		if !strings.Contains(err.Error(), v) {
			t.Errorf("el error no menciona %s:\n%v", v, err)
		}
	}
}

func TestDSNNoEnseñaLaClave(t *testing.T) {
	d := DSN("postgres://app:secreta@bd:5432/tareas")
	for _, s := range []string{d.String(), fmt.Sprint(d), fmt.Sprintf("%v", Config{BDURL: d})} {
		if strings.Contains(s, "secreta") {
			t.Errorf("la clave se ve: %s", s)
		}
	}
}
