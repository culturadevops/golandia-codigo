# SALIDA — Excel y PDF con Go — segunda parte (ejecutado)

Generado por `python correr_extra.py`. **No se edita a mano.**

- fecha: 2026-10-11 03:22 UTC
- imagen: fab-go-pdf:1.27.2 (Docker, linux/amd64)
- máquina: 24 · Intel(R) Core(TM) Ultra 7 270K Plus núcleos/CPU
- los tiempos son de ESTA máquina, dentro de Docker en Windows.

## 7 · Importar el Excel desde Go

`$ go run ./cmd/importar   # sin y con RawCellValue`

```
RawCellValue=false Teclado, enero: "1,200.00 €" · ParseFloat: strconv.ParseFloat: parsing "1,200.00 €": invalid syntax
RawCellValue=true  Teclado, enero: "1200" · ParseFloat: <nil>
```

## 8 · Formato condicional y una lista desplegable

`$ go run ./cmd/condicional && unzip -p informe-plus.xlsx xl/worksheets/sheet1.xml | grep ...`

```
informe-plus.xlsx: verde si > 8000 en B2:M6 · lista en O2:O6
<conditionalFormatting sqref="B2:M6">
<cfRule type="cellIs" dxfId="0" priority="1" operator="greaterThan">
<dataValidation allowBlank="true" sqref="O2:O6" type="list">
```

## 9 · Un QR en el PDF

`$ go run ./cmd/pdf -qr https://golandia.dev/informe-2026 && pdfimages -list con-qr.pdf`

```
con-qr.pdf: total del año 242.605,79 €
imagen en la página 1: 29×29 píxeles
```

## 10 · La cabecera en la página 2, y 3.000 líneas

`$ pdftotext -f 2 -l 2 d300.pdf -   # la página 2
$ time ./pdf -detalle 3000 -comprimir`

```
Informe de ventas 2026 · año, señal, ñandú, €
Producto

74 páginas · 133973 bytes
3000 líneas en 2.27 s
```

## 11 · Servirlos por HTTP, sin tocar el disco

`$ go run ./cmd/servidor &  curl -D - http://127.0.0.1:8080/informe.xlsx  ...informe.pdf`

```
== GET /informe.xlsx
Content-Disposition: attachment; filename="informe.xlsx"
Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
   6152 bytes
== GET /informe.pdf
Content-Disposition: inline; filename="informe.pdf"
Content-Type: application/pdf
   1134 bytes
```
