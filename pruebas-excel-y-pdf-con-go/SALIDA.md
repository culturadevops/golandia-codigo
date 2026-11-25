# SALIDA — Excel y PDF con Go (ejecutado)

Generado por `python correr.py`. **No se edita a mano.**

- fecha: 2026-10-11 03:19 UTC
- imagen: fab-go-pdf:1.27.2 (Docker, linux/amd64)
- máquina: 24 · Intel(R) Core(TM) Ultra 7 270K Plus núcleos/CPU
- los tiempos son de ESTA máquina, dentro de Docker en Windows.

## La versión de Go y las librerías

`$ go version && grep -E 'excelize|maroto' go.mod`

```
go version go1.27.2 linux/amd64
	github.com/johnfercher/maroto/v2 v2.4.3
	github.com/xuri/excelize/v2 v2.11.0
```

## 1 · El informe en Excel

`$ go run ./cmd/excel`

```
informe.xlsx: 2 hojas, 5 productos × 12 meses, 18 fórmulas, 1 gráfico
10593 informe.xlsx
```

`$ unzip -l informe.xlsx   # un .xlsx es un zip con XML`

```
xl/worksheets/sheet2.xml
xl/worksheets/sheet1.xml
xl/worksheets/_rels/sheet2.xml.rels
xl/workbook.xml
xl/theme/theme1.xml
xl/styles.xml
xl/sharedStrings.xml
xl/drawings/drawing1.xml
```

## 2 · Leerlo desde Go: la fórmula

`$ go run ./cmd/leer`

```
N7 fórmula:        =SUM(N2:N6)
N7 GetCellValue:   ""
N7 CalcCellValue:  "242,605.79 €" (err: <nil>)
filas en Ventas:   7
```

## 3 · 100.000 filas: SetCellValue contra StreamWriter

`$ ./grande -modo normal -filas 100000;  ./grande -modo stream -filas 100000`

```
normal 100000 filas: 1.63s · fichero 3.1 MB · memoria pedida en total 1019 MB
stream 100000 filas: 380ms · fichero 3.1 MB · memoria pedida en total 90 MB
        máximo de memoria del proceso: 518004 KB
        máximo de memoria del proceso: 71432 KB
```

## 3b · StreamWriter: en orden, sin volver atrás

`$ go run ./cmd/orden   # fila 2 y luego fila 1`

```
fila 2: <nil>
fila 1: row 1 has already been written
```

## 4 · El mismo informe en PDF

`$ go run ./cmd/pdf && pdfinfo informe.pdf`

```
informe.pdf: total del año 242.605,79 €
Title:           Informe de ventas
Producer:        FPDF 1.7
Pages:           1
Page size:       595.28 x 841.89 pts (A4)
10203 informe.pdf
```

## 5 · 300 líneas más, y la compresión

`$ go run ./cmd/pdf -detalle 300 -comprimir=false;  ... -comprimir=true`

```
comprimir=false: 8 páginas · 286991 bytes
comprimir=true: 8 páginas · 14996 bytes
```

## 6 · Letras que no están en WinAnsi

`$ ./pdf -otros   vs   ./pdf -otros -fuente DejaVuSans.ttf   # y pdftotext + pdffonts`

```
== sin-fuente.pdf · 2000 bytes
Sucursales: Kraków, .ód. . ......
   fuente Helvetica-Bold · incrustada: no
   fuente Helvetica · incrustada: no
== con-fuente.pdf · 35479 bytes
Sucursales: Kraków, Łódź → Москва
   fuente utf8dejavu · incrustada: yes
   fuente utf8dejavuB · incrustada: yes
```
