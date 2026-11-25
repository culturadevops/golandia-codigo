package comun

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// DirCDK es la carpeta de la app CDK, relativa a la raíz del repo.
const DirCDK = "cdk"

// CDK ejecuta `cdk <args>` en la carpeta de la app, imprime la línea de comando y devuelve
// la salida. Las líneas que son ruido de la herramienta (avisos de «feature flags», tiempos,
// el aviso de change set) no se imprimen; todo lo demás, tal cual.
func CDK(args ...string) string {
	Config("us-east-1") // la salvaguarda: la app también despliega
	npx := "npx"
	if runtime.GOOS == "windows" {
		npx = "npx.cmd"
	}
	cmd := exec.Command(npx, append([]string{"--no-install", "cdk", "--no-color"}, args...)...)
	cmd.Dir = DirCDK
	cmd.Env = append(os.Environ(), "AWS_REGION=us-east-1", "AWS_DEFAULT_REGION=us-east-1", "CDK_DISABLE_VERSION_CHECK=1")
	b, err := cmd.CombinedOutput()
	fmt.Printf("  $ cdk %s\n", strings.Join(args, " "))
	var lineas []string
	for _, l := range strings.Split(strings.ReplaceAll(string(b), "\r", ""), "\n") {
		t := strings.TrimSpace(l)
		if t == "" || strings.Contains(t, "feature flags") || strings.HasPrefix(t, "✨") ||
			strings.Contains(t, "Hold on while") || strings.Contains(t, "cdk-telemetry") ||
			strings.HasPrefix(t, "Stack ARN") || strings.HasPrefix(t, "arn:aws:cloudformation") ||
			strings.Contains(t, "CDK collects") || strings.Contains(t, "NOTICES") {
			continue
		}
		lineas = append(lineas, "    "+l)
	}
	if err != nil {
		lineas = append(lineas, "    (cdk salio con error: "+err.Error()+")")
	}
	return strings.Join(lineas, "\n")
}

// Plantilla lee la plantilla sintetizada de una pila.
func Plantilla(pila string) map[string]any {
	b, err := os.ReadFile(filepath.Join(DirCDK, "cdk.out", pila+".template.json"))
	if err != nil {
		log.Fatal(err)
	}
	var t map[string]any
	json.Unmarshal(b, &t)
	return t
}

// Recursos cuenta los recursos de una plantilla por tipo, sin el de metadatos de CDK.
func Recursos(pila string) (total int, porTipo [][2]any) {
	cuenta := map[string]int{}
	for _, r := range Plantilla(pila)["Resources"].(map[string]any) {
		tipo := r.(map[string]any)["Type"].(string)
		if tipo == "AWS::CDK::Metadata" {
			continue
		}
		cuenta[tipo]++
		total++
	}
	for k, v := range cuenta {
		porTipo = append(porTipo, [2]any{k, v})
	}
	sort.Slice(porTipo, func(i, j int) bool {
		if porTipo[i][1].(int) != porTipo[j][1].(int) {
			return porTipo[i][1].(int) > porTipo[j][1].(int)
		}
		return porTipo[i][0].(string) < porTipo[j][0].(string)
	})
	return total, porTipo
}

// Buckets lista los buckets que existen de verdad (en el emulador).
func Buckets() []string {
	out, err := s3.NewFromConfig(Config("us-east-1"), func(o *s3.Options) { o.UsePathStyle = true }).
		ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err != nil {
		log.Fatal(err)
	}
	var n []string
	for _, b := range out.Buckets {
		n = append(n, aws.ToString(b.Name))
	}
	return n
}
