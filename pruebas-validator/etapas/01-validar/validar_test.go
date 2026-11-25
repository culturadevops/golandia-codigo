// go-playground/validator: qué valida de verdad, y qué te devuelve cuando falla.
package validar

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
)

type Alta struct {
	Nombre string `validate:"required,min=2,max=40"`
	Correo string `validate:"required,email"`
	Edad   int    `validate:"gte=18,lte=120"`
	Rol    string `validate:"required,oneof=admin editor lector"`
	Activo bool   `validate:"required"` // ⚠️ trampa: false NO pasa required
	Tags   []string `validate:"dive,min=2"`
}

func TestLoQueAtrapa(t *testing.T) {
	v := validator.New()
	malo := Alta{
		Nombre: "A",                 // min=2
		Correo: "esto-no-es-correo", // email
		Edad:   -3,                  // gte=18
		Rol:    "jefe",              // oneof
		Activo: false,               // required + bool false
		Tags:   []string{"go", "x"}, // dive,min=2 -> "x" falla
	}
	err := v.Struct(malo)
	if err == nil {
		t.Fatal("deberia haber fallado")
	}
	errs := err.(validator.ValidationErrors)
	fmt.Printf("  fallos detectados: %d\n", len(errs))
	for _, e := range errs {
		fmt.Printf("    campo=%-8s regla=%-10s valor=%v\n", e.Field(), e.Tag(), e.Value())
	}
}

// ⭐ El error CRUDO, tal cual sale. Esto es lo que la documentación se salta.
func TestElErrorCrudoEsIlegible(t *testing.T) {
	v := validator.New()
	err := v.Struct(Alta{Nombre: "A", Correo: "x", Edad: 5, Rol: "jefe", Activo: true, Tags: nil})
	fmt.Println("  --- lo que sale si haces fmt.Println(err) ---")
	fmt.Println(" ", err)
}

// Convertirlo en algo que una persona entienda: el trabajo que hay que hacer.
func mensajes(err error) []string {
	var out []string
	for _, e := range err.(validator.ValidationErrors) {
		switch e.Tag() {
		case "required":
			out = append(out, fmt.Sprintf("%s es obligatorio", e.Field()))
		case "email":
			out = append(out, fmt.Sprintf("%s no parece un correo valido", e.Field()))
		case "min":
			out = append(out, fmt.Sprintf("%s necesita al menos %s", e.Field(), e.Param()))
		case "gte":
			out = append(out, fmt.Sprintf("%s tiene que ser %s o mas", e.Field(), e.Param()))
		case "oneof":
			out = append(out, fmt.Sprintf("%s debe ser uno de: %s", e.Field(), e.Param()))
		default:
			out = append(out, fmt.Sprintf("%s no cumple %s", e.Field(), e.Tag()))
		}
	}
	return out
}

func TestMensajesDecentes(t *testing.T) {
	v := validator.New()
	err := v.Struct(Alta{Nombre: "A", Correo: "x", Edad: 5, Rol: "jefe", Activo: true, Tags: nil})
	fmt.Println("  --- lo mismo, traducido a mano ---")
	for _, m := range mensajes(err) {
		fmt.Println("   ·", m)
	}
}

// ⚠️ LA TRAMPA NUMERO UNO: required contra el valor cero.
func TestRequiredContraElValorCero(t *testing.T) {
	v := validator.New()
	type T struct {
		Booleano bool `validate:"required"`
		Entero   int  `validate:"required"`
		Texto    string `validate:"required"`
	}
	err := v.Struct(T{Booleano: false, Entero: 0, Texto: ""})
	fmt.Printf("  con false, 0 y cadena vacia -> %d fallos\n", len(err.(validator.ValidationErrors)))
	for _, e := range err.(validator.ValidationErrors) {
		fmt.Printf("    %s falla 'required' con valor %v\n", e.Field(), e.Value())
	}
}

// ⚠️ Una etiqueta MAL ESCRITA: ¿avisa el compilador? ¿avisa alguien?
func TestEtiquetaMalEscrita(t *testing.T) {
	v := validator.New()
	type T struct {
		Correo string `validate:"emial"` // <- typo a proposito
	}
	defer func() {
		if r := recover(); r != nil {
			s := fmt.Sprint(r)
			if len(s) > 120 {
				s = s[:120]
			}
			fmt.Println("  PANIC en tiempo de EJECUCION:", strings.TrimSpace(s))
		}
	}()
	err := v.Struct(T{Correo: "no-es-correo"})
	fmt.Println("  resultado con la etiqueta mal escrita:", err)
}

// Un validador propio, con un caso real: un RUC peruano (11 digitos, empieza por 10 o 20).
func TestValidadorPropio(t *testing.T) {
	v := validator.New()
	v.RegisterValidation("ruc", func(fl validator.FieldLevel) bool {
		s := fl.Field().String()
		if len(s) != 11 {
			return false
		}
		if !strings.HasPrefix(s, "10") && !strings.HasPrefix(s, "20") {
			return false
		}
		for _, c := range s {
			if c < '0' || c > '9' {
				return false
			}
		}
		return true
	})
	type Empresa struct {
		RUC string `validate:"required,ruc"`
	}
	for _, caso := range []string{"20123456789", "10123456789", "30123456789", "2012345678", "2012345678X"} {
		err := v.Struct(Empresa{RUC: caso})
		fmt.Printf("    %-12s -> %v\n", caso, err == nil)
	}
}

func BenchmarkValidarStruct(b *testing.B) {
	v := validator.New()
	bueno := Alta{Nombre: "Ana", Correo: "ana@example.com", Edad: 30, Rol: "admin", Activo: true, Tags: []string{"go"}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.Struct(bueno)
	}
}
