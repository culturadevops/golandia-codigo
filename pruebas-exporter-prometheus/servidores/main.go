// Servidores TLS de prueba para el exporter (golandia 679). Genera su propia CA y tres certificados
// al arrancar, así las fechas son relativas a «ahora»:
//
//	:9441  pronto   vence en 5 días
//	:9442  vencido  venció hace 1 día
//	:9443  sano     vence en 90 días
//	:9444  lento    acepta la conexión TCP y nunca contesta el handshake
//
// Escribe la CA en -ca (para que el blackbox_exporter pueda validar la cadena).
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

func main() {
	nombre := flag.String("nombre", "servidores", "nombre DNS que llevan los certificados")
	caPath := flag.String("ca", "/compartido/ca.pem", "dónde escribir la CA")
	flag.Parse()

	ahora := time.Now()
	caKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	caTpl := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "CA de prueba golandia"},
		NotBefore: ahora.Add(-365 * 24 * time.Hour), NotAfter: ahora.Add(5 * 365 * 24 * time.Hour),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTpl, caTpl, &caKey.PublicKey, caKey)
	if err != nil {
		log.Fatal(err)
	}
	caCert, _ := x509.ParseCertificate(caDER)
	if err := os.WriteFile(*caPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER}), 0o644); err != nil {
		log.Fatal(err)
	}

	dia := 24 * time.Hour
	sitios := []struct {
		puerto      string
		desde, hast time.Duration
	}{
		{":9441", -30 * dia, 5 * dia},  // pronto
		{":9442", -91 * dia, -1 * dia}, // vencido
		{":9443", -1 * dia, 90 * dia},  // sano
	}
	for i, s := range sitios {
		key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		tpl := &x509.Certificate{
			SerialNumber: big.NewInt(int64(i + 2)), Subject: pkix.Name{CommonName: *nombre},
			DNSNames: []string{*nombre}, NotBefore: ahora.Add(s.desde), NotAfter: ahora.Add(s.hast),
			KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		}
		der, err := x509.CreateCertificate(rand.Reader, tpl, caCert, &key.PublicKey, caKey)
		if err != nil {
			log.Fatal(err)
		}
		cfg := &tls.Config{Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}}}
		ln, err := tls.Listen("tcp", s.puerto, cfg)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%s  vence %s", s.puerto, tpl.NotAfter.UTC().Format(time.RFC3339))
		go servir(ln)
	}

	// El lento: acepta y se queda callado. Un handshake TLS contra él no termina nunca.
	ln, err := net.Listen("tcp", ":9444")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf(":9444  lento (no contesta el handshake)")
	var colgadas []net.Conn
	for {
		c, err := ln.Accept()
		if err != nil {
			continue
		}
		colgadas = append(colgadas, c) // se guardan para que no las cierre el GC
	}
}

func servir(ln net.Listener) {
	for {
		c, err := ln.Accept()
		if err != nil {
			continue
		}
		go func() {
			defer c.Close()
			c.(*tls.Conn).Handshake()
			c.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 2\r\nConnection: close\r\n\r\nok"))
		}()
	}
}
