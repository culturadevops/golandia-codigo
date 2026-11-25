package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// falsificarAdmin simula a un atacante que no conoce la clave secreta:
// decodifica el payload (es texto plano en base64, no esta cifrado),
// cambia "admin" a true, y lo vuelve a codificar. La firma queda igual
// que antes, así que ya no corresponde al nuevo payload.
func falsificarAdmin(tokenString string) string {
	partes := strings.Split(tokenString, ".")
	payloadBytes, _ := base64.RawURLEncoding.DecodeString(partes[1])

	var claims map[string]interface{}
	json.Unmarshal(payloadBytes, &claims)
	claims["admin"] = true

	nuevoPayload, _ := json.Marshal(claims)
	nuevoPayloadCodificado := base64.RawURLEncoding.EncodeToString(nuevoPayload)

	return partes[0] + "." + nuevoPayloadCodificado + "." + partes[2]
}

const claveSecreta = "esta-es-la-clave-secreta-del-servidor"

func demoBcrypt() {
	fmt.Println("=== bcrypt: contraseñas ===")
	contrasena := "hola-mundo-123"

	inicio := time.Now()
	hash1, _ := bcrypt.GenerateFromPassword([]byte(contrasena), bcrypt.DefaultCost)
	duracionBcrypt := time.Since(inicio)

	hash2, _ := bcrypt.GenerateFromPassword([]byte(contrasena), bcrypt.DefaultCost)

	fmt.Printf("hash 1: %s\n", hash1)
	fmt.Printf("hash 2: %s\n", hash2)
	fmt.Printf("misma contraseña, hashes distintos: %v\n", string(hash1) != string(hash2))

	okCorrecta := bcrypt.CompareHashAndPassword(hash1, []byte(contrasena)) == nil
	okIncorrecta := bcrypt.CompareHashAndPassword(hash1, []byte("otra-cosa")) == nil
	fmt.Printf("verificar con la contraseña correcta: %v\n", okCorrecta)
	fmt.Printf("verificar con una contraseña incorrecta: %v\n", okIncorrecta)
	fmt.Printf("bcrypt tardo: %s\n", duracionBcrypt.Round(time.Millisecond))

	inicioRapido := time.Now()
	sum := sha256.Sum256([]byte(contrasena))
	duracionRapida := time.Since(inicioRapido)
	fmt.Printf("sha256 (NO usar para contraseñas): %s, tardo %s\n", hex.EncodeToString(sum[:8]), duracionRapida)
}

func crearToken(admin bool) string {
	claims := jwt.MapClaims{
		"usuario": "ana",
		"admin":   admin,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	firmado, _ := token.SignedString([]byte(claveSecreta))
	return firmado
}

// verificarInseguro decodifica el token SIN comprobar la firma.
func verificarInseguro(tokenString string) (jwt.MapClaims, error) {
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}
	claims, _ := token.Claims.(jwt.MapClaims)
	return claims, nil
}

// verificarSeguro exige la firma correcta con la clave real.
func verificarSeguro(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metodo de firma inesperado: %v", t.Header["alg"])
		}
		return []byte(claveSecreta), nil
	})
	if err != nil {
		return nil, err
	}
	claims, _ := token.Claims.(jwt.MapClaims)
	return claims, nil
}

func demoJWT() {
	fmt.Println()
	fmt.Println("=== JWT: firma y verificacion ===")

	tokenReal := crearToken(false)
	fmt.Printf("token real (admin=false): %s...\n", tokenReal[:40])

	// el atacante no conoce la clave secreta, pero puede cambiar el payload a mano
	// y volver a codificarlo en base64: la firma queda desactualizada, pero el
	// texto es perfectamente legible sin necesitar ninguna clave.
	tokenFalsificado := falsificarAdmin(tokenReal)

	fmt.Println()
	fmt.Println("--- verificarInseguro (ParseUnverified, sin comprobar firma) ---")
	claimsInseguro, _ := verificarInseguro(tokenFalsificado)
	fmt.Printf("admin segun el token falsificado: %v\n", claimsInseguro["admin"])

	fmt.Println()
	fmt.Println("--- verificarSeguro (jwt.Parse, comprobando la firma real) ---")
	_, errSeguro := verificarSeguro(tokenFalsificado)
	fmt.Printf("token falsificado, verificacion segura: %v\n", errSeguro)

	fmt.Println()
	fmt.Println("--- verificarSeguro con el token real, sin tocar ---")
	claimsOk, errOk := verificarSeguro(tokenReal)
	fmt.Printf("token real, error: %v, admin: %v\n", errOk, claimsOk["admin"])
}

func main() {
	demoBcrypt()
	demoJWT()
}
