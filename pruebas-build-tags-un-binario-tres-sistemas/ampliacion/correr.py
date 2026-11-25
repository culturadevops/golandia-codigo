# -*- coding: utf-8 -*-
"""build tags — AMPLIACIÓN del banco (golandia, vídeo 368, 2026-10-11). Escribe ampliacion/SALIDA.md.

Aparte del banco principal para no rehacer sus salidas. Trabaja sobre una COPIA de ../proyecto:
  1  el binario guarda con qué etiquetas se compiló (go version -m);
  2  los tests también llevan sufijo de sistema; go test -c para Windows desde Linux;
  3  cuántos ficheros de la librería estándar llevan //go:build;
  4  un cuarto destino: WebAssembly con GOOS=wasip1.

    python correr.py
"""
import sys
from pathlib import Path

AQUI = Path(__file__).resolve().parent
sys.path.insert(0, str(AQUI.parent.parent))
from banco_go import Banco  # noqa: E402

COPIA = "cp -r /w /tmp/p && cd /tmp/p && "
TEST_LINUX = ("cat > ruta_linux_test.go <<'X'\npackage main\n\nimport \"testing\"\n\n"
              "func TestRutaLinux(t *testing.T) {\n\tif got := rutaConfig(); got != \"/etc/app/config.toml\" {\n"
              "\t\tt.Fatal(got)\n\t}\n}\nX\n")
TEST_WIN = ("cat > ruta_windows_test.go <<'X'\npackage main\n\nimport \"testing\"\n\n"
            "func TestRutaWindows(t *testing.T) {\n\tif got := rutaConfig(); got != `C:\ProgramData\app\config.toml` {\n"
            "\t\tt.Fatal(got)\n\t}\n}\nX\n")

with Banco(AQUI, "build tags — ampliación", proyecto="../proyecto") as b:
    b.titulo("1 · El binario recuerda sus etiquetas")
    b.sh(COPIA + "go build -o libre . && go build -tags pro -o pro . && "
         "for x in libre pro; do echo \"== $x\"; go version -m $x | grep -E 'build\s+(-tags|GOOS|GOARCH)'; done",
         mostrar="go build -o libre . && go build -tags pro -o pro .\n$ go version -m libre;  go version -m pro")

    b.titulo("2 · Los tests también llevan sufijo")
    b.sh(COPIA + TEST_LINUX + TEST_WIN + "go test -v . 2>&1 | grep -E '^(=== RUN|--- |ok)' ; "
         "GOOS=windows go test -c -o ruta.test.exe . && ls -l ruta.test.exe | awk '{print $5, $9}' && "
         "GOOS=windows go vet . && echo 'go vet para windows: sin quejas'",
         mostrar="go test -v .     # en Linux\n$ GOOS=windows go test -c -o ruta.test.exe .")

    b.titulo("3 · La librería estándar")
    b.sh("cd $(go env GOROOT)/src && total=$(find . -name '*.go' -not -path '*/testdata/*' | wc -l) && "
         "con=$(grep -rl --include='*.go' --exclude-dir=testdata '^//go:build' . | wc -l) && "
         "echo \"ficheros .go: $total · con //go:build: $con\" && "
         "echo \"en os/: $(ls os/*_windows.go | wc -l) _windows.go · $(ls os/*_linux.go | wc -l) _linux.go · $(ls os/*_darwin.go 2>/dev/null | wc -l) _darwin.go\"",
         mostrar="cd $(go env GOROOT)/src && grep -rl '^//go:build' . | wc -l")

    b.titulo("4 · Un cuarto destino: WebAssembly")
    b.sh(COPIA + "GOOS=wasip1 GOARCH=wasm go build -o app.wasm . 2>&1 | head -3; "
         "echo \"ficheros que entran: $(GOOS=wasip1 GOARCH=wasm go list -f '{{.GoFiles}}' . 2>/dev/null)\"",
         mostrar="GOOS=wasip1 GOARCH=wasm go build -o app.wasm .")
    b.sh(COPIA + "printf '//go:build !linux && !darwin && !windows\n\npackage main\n\n"
         "func rutaConfig() string { return \"config.toml\" } // cualquier otro sistema\n' > ruta_otros.go && cat ruta_otros.go && "
         "GOOS=wasip1 GOARCH=wasm go build -o app.wasm . && ls -l app.wasm | awk '{print $5, $9}' && "
         "echo \"ficheros que entran: $(GOOS=wasip1 GOARCH=wasm go list -f '{{.GoFiles}}' .)\" && "
         "echo \"y en linux:          $(go list -f '{{.GoFiles}}' .)\"",
         mostrar="cat ruta_otros.go && GOOS=wasip1 GOARCH=wasm go build -o app.wasm .")
