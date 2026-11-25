// Lección 1 · protobuf: el número de campo es el contrato, el nombre no.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/proto"

	"curso-grpc/comun"
	"curso-grpc/gen/l01/antes"
	"curso-grpc/gen/l01/bien"
	"curso-grpc/gen/l01/despues"
)

func main() {
	comun.Titulo("1.1 · Por el cable no viaja el nombre del campo, viaja su número")
	viejo := &antes.Producto{Sku: "TAZA-01", Stock: 5}
	b, _ := proto.Marshal(viejo)
	fmt.Printf("  el cliente viejo manda: sku=%q stock=%d\n", viejo.Sku, viejo.Stock)
	fmt.Printf("  en el cable: %d bytes · % x\n", len(b), b)

	comun.Titulo("1.2 · La trampa: reutilizar el número de un campo borrado")
	var nuevo despues.Producto
	err := proto.Unmarshal(b, &nuevo)
	fmt.Printf("  el servidor nuevo lee: sku=%q precio_centavos=%d · error: %v\n", nuevo.Sku, nuevo.PrecioCentavos, err)
	fmt.Printf("  -> la taza ahora cuesta %d centavos. protoc no avisó y Unmarshal tampoco.\n", nuevo.PrecioCentavos)

	comun.Titulo("1.3 · Lo correcto: reserved, y un número nuevo")
	var ok bien.Producto
	err = proto.Unmarshal(b, &ok)
	fmt.Printf("  el mismo mensaje viejo, leído con reserved 2: sku=%q · precio presente: %v · error: %v\n",
		ok.Sku, ok.PrecioCentavos != nil, err)
	fmt.Printf("  el campo 2 no se pierde: queda guardado como desconocido (%d bytes)\n", len(ok.ProtoReflect().GetUnknown()))
	fmt.Println("  y si alguien intenta volver a usar el 2, protoc se niega:")
	for _, l := range protocConElDos() {
		fmt.Println("    " + l)
	}

	comun.Titulo("1.4 · La otra trampa: en proto3, «no vino» y «vale 0» son lo mismo")
	agotado, _ := proto.Marshal(&antes.Producto{Sku: "TAZA-01", Stock: 0})
	sinStock, _ := proto.Marshal(&antes.Producto{Sku: "TAZA-01"})
	fmt.Printf("  stock=0: %d bytes · stock sin poner: %d bytes · ¿iguales?: %v\n",
		len(agotado), len(sinStock), string(agotado) == string(sinStock))
	gratis, _ := proto.Marshal(&bien.Producto{Sku: "TAZA-01", PrecioCentavos: proto.Int32(0)})
	sinPrecio, _ := proto.Marshal(&bien.Producto{Sku: "TAZA-01"})
	var g, s bien.Producto
	proto.Unmarshal(gratis, &g)
	proto.Unmarshal(sinPrecio, &s)
	fmt.Printf("  con optional: precio=0 -> %d bytes, presente: %v · sin precio -> %d bytes, presente: %v\n",
		len(gratis), g.PrecioCentavos != nil, len(sinPrecio), s.PrecioCentavos != nil)
}

// protocConElDos compila un .proto que reutiliza el número reservado y devuelve lo que dice protoc.
func protocConElDos() []string {
	dir, _ := os.MkdirTemp("", "l01")
	defer os.RemoveAll(dir)
	src := `syntax = "proto3";
message Producto {
  reserved 2;
  string sku = 1;
  int32 descuento = 2;
}
`
	os.WriteFile(filepath.Join(dir, "x.proto"), []byte(src), 0o644)
	out, err := exec.Command("protoc", "-I", dir, "--descriptor_set_out=/dev/null", "x.proto").CombinedOutput()
	lineas := []string{}
	for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if l != "" {
			lineas = append(lineas, l)
		}
	}
	codigo := 0
	if e, ok := err.(*exec.ExitError); ok {
		codigo = e.ExitCode()
	}
	return append(lineas, fmt.Sprintf("(protoc, código de salida %d)", codigo))
}
