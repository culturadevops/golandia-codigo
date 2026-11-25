// ⚠️ VULNERABLE A PROPÓSITO (demo de gosec y govulncheck, golandia 683). No desplegar.
package main

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/net/html"
)

var db *sql.DB

var clave = []byte("cambiar-en-produccion")

// Fallo de código 1: SQL armada con fmt.Sprintf.
func buscar(w http.ResponseWriter, r *http.Request) {
	q := fmt.Sprintf("SELECT nombre FROM productos WHERE id = '%s'", r.URL.Query().Get("id"))
	filas, err := db.Query(q)
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	defer filas.Close()
	fmt.Fprintln(w, "ok")
}

// Fallo de código 2: path traversal (archivo=../../etc/passwd).
// Fallo de código 4: el error de w.Write se ignora.
func factura(w http.ResponseWriter, r *http.Request) {
	ruta := filepath.Join("facturas", r.URL.Query().Get("archivo"))
	datos, err := os.ReadFile(ruta)
	if err != nil {
		http.Error(w, "no existe", http.StatusNotFound)
		return
	}
	w.Write(datos)
}

// Fallo de código 3: math/rand para un token de sesión.
func token() string {
	const letras = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = letras[rand.Intn(len(letras))]
	}
	return string(b)
}

// Fallo de código 5: TLS sin verificar el certificado.
var cliente = &http.Client{Transport: &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}}

// Dependencia 1: jwt.Parse SÍ se llama (github.com/golang-jwt/jwt/v4 v4.5.0).
func sesion(r *http.Request) bool {
	t, err := jwt.Parse(r.Header.Get("Authorization"), func(t *jwt.Token) (any, error) {
		return clave, nil
	})
	return err == nil && t.Valid
}

// Dependencia 2: golang.org/x/net/html (versión vieja) solo con EscapeString.
func saludo(w http.ResponseWriter, r *http.Request) {
	if !sesion(r) {
		http.Error(w, "sin sesión", http.StatusUnauthorized)
		return
	}
	fmt.Fprintf(w, "<p>Hola, %s</p>", html.EscapeString(r.URL.Query().Get("nombre")))
}

func nueva(w http.ResponseWriter, r *http.Request) {
	resp, err := cliente.Get("https://pagos.ejemplo.com/ping")
	if err == nil {
		resp.Body.Close()
	}
	fmt.Fprintln(w, token())
}

func main() {
	http.HandleFunc("/buscar", buscar)
	http.HandleFunc("/factura", factura)
	http.HandleFunc("/saludo", saludo)
	http.HandleFunc("/sesion/nueva", nueva)
	fmt.Println(http.ListenAndServe(":8080", nil))
}
