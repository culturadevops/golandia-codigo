# -*- coding: utf-8 -*-
"""Excel y PDF con Go — segunda parte del banco (golandia, vídeo 386). Escribe SALIDA-extra.md.

Aparte de correr.py para no rehacer sus tiempos (que ya están en el guion). Usa el informe.xlsx
que deja correr.py, y añade: importar el Excel desde Go (formato frente a RawCellValue), formato
condicional y lista desplegable, el PDF con un QR, la cabecera en la página 2, 3.000 líneas con
tiempo, y los dos ficheros servidos por HTTP sin tocar el disco.

    python correr.py && python correr_extra.py
"""
import sys
from pathlib import Path

AQUI = Path(__file__).resolve().parent
sys.path.insert(0, str(AQUI.parent))
import banco_go  # noqa: E402

banco_go.IMAGEN = "fab-go-pdf:1.27.2"

with banco_go.Banco(AQUI, "Excel y PDF con Go — segunda parte", salida="SALIDA-extra.md") as b:
    b.titulo("7 · Importar el Excel desde Go")
    b.sh("go run ./cmd/importar", mostrar="go run ./cmd/importar   # sin y con RawCellValue")

    b.titulo("8 · Formato condicional y una lista desplegable")
    b.sh("go run ./cmd/condicional && unzip -p informe-plus.xlsx xl/worksheets/sheet1.xml | "
         "grep -o '<conditionalFormatting[^>]*>\|<cfRule[^>]*>\|<dataValidation [^>]*>'",
         mostrar="go run ./cmd/condicional && unzip -p informe-plus.xlsx xl/worksheets/sheet1.xml | grep ...")

    b.titulo("9 · Un QR en el PDF")
    b.sh("go run ./cmd/pdf -comprimir -qr https://golandia.dev/informe-2026 -o con-qr.pdf && "
         "pdfimages -list con-qr.pdf | awk 'NR>2 {print \"imagen en la página\", $1 \":\", $4 \"×\" $5, \"píxeles\"}' && "
         "pdftoppm -r 110 -png -f 1 -l 1 con-qr.pdf pagina-qr",
         mostrar="go run ./cmd/pdf -qr https://golandia.dev/informe-2026 && pdfimages -list con-qr.pdf")

    b.titulo("10 · La cabecera en la página 2, y 3.000 líneas")
    b.sh("go build -o /tmp/pdf ./cmd/pdf && /tmp/pdf -detalle 300 -comprimir -o d300.pdf > /dev/null && "
         "pdftotext -f 2 -l 2 d300.pdf - | head -3 && pdftoppm -r 110 -png -f 2 -l 2 d300.pdf pagina-dos && "
         "/usr/bin/time -f '3000 líneas en %e s' /tmp/pdf -detalle 3000 -comprimir -o d3000.pdf > /dev/null && "
         "echo \"$(pdfinfo d3000.pdf | awk '/Pages/ {print $2}') páginas · $(stat -c %s d3000.pdf) bytes\"",
         mostrar="pdftotext -f 2 -l 2 d300.pdf -   # la página 2\n$ time ./pdf -detalle 3000 -comprimir")

    b.titulo("11 · Servirlos por HTTP, sin tocar el disco")
    b.sh("go build -o /tmp/srv ./cmd/servidor && (/tmp/srv &) && sleep 1 && for f in informe.xlsx informe.pdf; do "
         "echo \"== GET /$f\"; curl -s -D - -o /tmp/$f http://127.0.0.1:8080/$f | grep -iE '^content-(type|disposition)'; "
         "echo \"   $(stat -c %s /tmp/$f) bytes\"; done",
         mostrar="go run ./cmd/servidor &  curl -D - http://127.0.0.1:8080/informe.xlsx  ...informe.pdf")
