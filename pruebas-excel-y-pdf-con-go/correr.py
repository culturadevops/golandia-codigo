# -*- coding: utf-8 -*-
"""Excel y PDF con Go — banco ejecutable (golandia, vídeo 386). Escribe SALIDA.md.

proyecto/: un informe de ventas (5 productos × 12 meses, datos deterministas en datos/)
  cmd/excel   el .xlsx con excelize v2.11.0: estilos, moneda, fórmulas, filtro, panel fijo, gráfico;
  cmd/leer    abrirlo desde Go: GetCellValue de una fórmula frente a CalcCellValue;
  cmd/grande  100.000 filas con SetCellValue frente a StreamWriter (tiempo, memoria, máximo RSS);
  cmd/orden   StreamWriter no deja volver a una fila ya escrita;
  cmd/pdf     el mismo informe en PDF con maroto v2.4.3: compresión y fuentes UTF-8.
Imagen imagen/Dockerfile = golang:1.27.2 + poppler-utils (pdfinfo, pdffonts, pdftotext, pdftoppm)
+ fonts-dejavu-core. Las páginas en PNG (para el vídeo) quedan en proyecto/*.png.

    python correr.py
"""
import sys
from pathlib import Path

AQUI = Path(__file__).resolve().parent
sys.path.insert(0, str(AQUI.parent))
import banco_go  # noqa: E402

banco_go.IMAGEN = "fab-go-pdf:1.27.2"
DEJAVU = "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"

with banco_go.Banco(AQUI, "Excel y PDF con Go") as b:
    b.titulo("La versión de Go y las librerías")
    b.sh("go version && grep -E 'excelize|maroto' go.mod", mostrar="go version && grep -E 'excelize|maroto' go.mod")

    b.titulo("1 · El informe en Excel")
    b.sh("go run ./cmd/excel && ls -l informe.xlsx | awk '{print $5, $9}'", mostrar="go run ./cmd/excel")
    b.sh("unzip -l informe.xlsx | awk 'NR>3 && $4 ~ /xl\// {print $4}' | head -8",
         mostrar="unzip -l informe.xlsx   # un .xlsx es un zip con XML")

    b.titulo("2 · Leerlo desde Go: la fórmula")
    b.sh("go run ./cmd/leer", mostrar="go run ./cmd/leer")

    b.titulo("3 · 100.000 filas: SetCellValue contra StreamWriter")
    b.sh("go build -o /tmp/grande ./cmd/grande && for m in normal stream; do "
         "/usr/bin/time -f \"        máximo de memoria del proceso: %M KB\" /tmp/grande -modo $m -filas 100000; done",
         mostrar="./grande -modo normal -filas 100000;  ./grande -modo stream -filas 100000")

    b.titulo("3b · StreamWriter: en orden, sin volver atrás")
    b.sh("go run ./cmd/orden", mostrar="go run ./cmd/orden   # fila 2 y luego fila 1")

    b.titulo("4 · El mismo informe en PDF")
    b.sh("go run ./cmd/pdf && pdfinfo informe.pdf | grep -E 'Title|Producer|Pages|Page size' && "
         "ls -l informe.pdf | awk '{print $5, $9}' && pdftoppm -r 110 -png -f 1 -l 1 informe.pdf pagina-informe",
         mostrar="go run ./cmd/pdf && pdfinfo informe.pdf")

    b.titulo("5 · 300 líneas más, y la compresión")
    b.sh("for c in false true; do go run ./cmd/pdf -detalle 300 -comprimir=$c -o detalle-$c.pdf > /dev/null && "
         "echo \"comprimir=$c: $(pdfinfo detalle-$c.pdf | awk '/Pages/ {print $2}') páginas · $(stat -c %s detalle-$c.pdf) bytes\"; done",
         mostrar="go run ./cmd/pdf -detalle 300 -comprimir=false;  ... -comprimir=true")

    b.titulo("6 · Letras que no están en WinAnsi")
    b.sh("go run ./cmd/pdf -otros -comprimir -o sin-fuente.pdf > /dev/null && "
         f"go run ./cmd/pdf -otros -comprimir -fuente {DEJAVU} -o con-fuente.pdf > /dev/null && "
         "for f in sin-fuente con-fuente; do echo \"== $f.pdf · $(stat -c %s $f.pdf) bytes\"; "
         "pdftotext $f.pdf - | grep Sucursales; pdffonts $f.pdf | tail -n +3 | awk '{print \"   fuente\", $1, \"· incrustada:\", $(NF-4)}'; done && "
         "pdftoppm -r 110 -png -f 1 -l 1 sin-fuente.pdf pagina-sin-fuente && pdftoppm -r 110 -png -f 1 -l 1 con-fuente.pdf pagina-con-fuente",
         mostrar="./pdf -otros   vs   ./pdf -otros -fuente DejaVuSans.ttf   # y pdftotext + pdffonts")
