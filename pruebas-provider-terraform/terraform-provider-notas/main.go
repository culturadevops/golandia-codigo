// terraform-provider-notas: un provider propio para el Terraform de verdad (golandia 675).
// Habla con notas-api. Se usa sin publicar, con dev_overrides en .terraformrc.
package main

import (
	"context"
	"flag"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// trampas activa a propósito los errores típicos, para medirlos en la demo:
//
//	go build -ldflags "-X main.trampas=clave"   (o 404, lista, vacio; separadas por comas)
//
// Sin -ldflags, el provider es la versión correcta.
var trampas string

func trampa(n string) bool {
	for _, t := range strings.Split(trampas, ",") {
		if t == n {
			return true
		}
	}
	return false
}

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "arranca en modo depuración")
	flag.Parse()

	err := providerserver.Serve(context.Background(), func() provider.Provider { return &notasProvider{} },
		providerserver.ServeOpts{Address: "registry.terraform.io/golandia/notas", Debug: debug})
	if err != nil {
		log.Fatal(err)
	}
}
