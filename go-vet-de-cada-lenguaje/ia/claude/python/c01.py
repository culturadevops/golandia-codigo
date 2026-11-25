"""Estadísticas de precios a partir de un archivo CSV de productos.

Formato esperado del CSV (con encabezado):

    producto,precio
    Teclado,25.50
    Monitor,189.99

Uso:
    python estadisticas_precios.py productos.csv [--delimitador ";"] [--encoding utf-8]
"""

from __future__ import annotations

import argparse
import csv
import statistics
import sys
from dataclasses import dataclass
from decimal import Decimal, InvalidOperation
from pathlib import Path


@dataclass(frozen=True, slots=True)
class Producto:
    nombre: str
    precio: Decimal


@dataclass(frozen=True, slots=True)
class Estadisticas:
    promedio: Decimal
    mediana: Decimal
    mas_caro: Producto
    total_productos: int


class ErrorCSV(ValueError):
    """Error de formato o contenido en el archivo CSV."""


def _normalizar(texto: str) -> str:
    return texto.strip().lower()


def _parsear_precio(valor: str, numero_linea: int) -> Decimal:
    limpio = valor.strip().replace("$", "").replace(" ", "")
    # Admite coma decimal ("12,50") si no hay punto.
    if "," in limpio and "." not in limpio:
        limpio = limpio.replace(",", ".")
    try:
        precio = Decimal(limpio)
    except InvalidOperation as exc:
        raise ErrorCSV(f"Línea {numero_linea}: precio inválido {valor!r}") from exc
    if not precio.is_finite() or precio < 0:
        raise ErrorCSV(f"Línea {numero_linea}: precio fuera de rango {valor!r}")
    return precio


def leer_productos(
    ruta: Path,
    *,
    columna_producto: str = "producto",
    columna_precio: str = "precio",
    delimitador: str = ",",
    encoding: str = "utf-8-sig",
) -> list[Producto]:
    """Lee el CSV y devuelve la lista de productos con su precio."""
    with ruta.open(newline="", encoding=encoding) as archivo:
        lector = csv.DictReader(archivo, delimiter=delimitador)
        if lector.fieldnames is None:
            raise ErrorCSV("El archivo está vacío o no tiene encabezado.")

        columnas = {_normalizar(c): c for c in lector.fieldnames if c is not None}
        col_prod = columnas.get(_normalizar(columna_producto))
        col_prec = columnas.get(_normalizar(columna_precio))
        if col_prod is None or col_prec is None:
            raise ErrorCSV(
                f"Faltan columnas requeridas: se esperaban "
                f"{columna_producto!r} y {columna_precio!r}, "
                f"se encontraron {lector.fieldnames!r}"
            )

        productos: list[Producto] = []
        for fila in lector:
            numero_linea = lector.line_num
            nombre = (fila.get(col_prod) or "").strip()
            valor_precio = (fila.get(col_prec) or "").strip()
            if not nombre and not valor_precio:
                continue  # Fila vacía.
            if not nombre:
                raise ErrorCSV(f"Línea {numero_linea}: falta el nombre del producto.")
            if not valor_precio:
                raise ErrorCSV(f"Línea {numero_linea}: falta el precio de {nombre!r}.")
            productos.append(Producto(nombre, _parsear_precio(valor_precio, numero_linea)))

    return productos


def calcular_estadisticas(productos: list[Producto]) -> Estadisticas:
    """Calcula promedio, mediana y producto más caro."""
    if not productos:
        raise ValueError("No hay productos para calcular estadísticas.")

    precios = [p.precio for p in productos]
    return Estadisticas(
        promedio=statistics.mean(precios),
        mediana=statistics.median(precios),
        mas_caro=max(productos, key=lambda p: p.precio),
        total_productos=len(productos),
    )


def formatear(estadisticas: Estadisticas, decimales: int = 2) -> str:
    """Devuelve un resumen legible de las estadísticas."""
    cuantizador = Decimal(1).scaleb(-decimales)

    def fmt(valor: Decimal) -> str:
        return str(valor.quantize(cuantizador))

    return "\n".join(
        [
            f"Productos analizados: {estadisticas.total_productos}",
            f"Precio promedio:      {fmt(estadisticas.promedio)}",
            f"Precio mediano:       {fmt(estadisticas.mediana)}",
            f"Producto más caro:    {estadisticas.mas_caro.nombre} "
            f"({fmt(estadisticas.mas_caro.precio)})",
        ]
    )


def _crear_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="Calcula promedio, mediana y producto más caro desde un CSV."
    )
    parser.add_argument("archivo", type=Path, help="Ruta al archivo CSV.")
    parser.add_argument("--columna-producto", default="producto")
    parser.add_argument("--columna-precio", default="precio")
    parser.add_argument("--delimitador", default=",")
    parser.add_argument("--encoding", default="utf-8-sig")
    parser.add_argument("--decimales", type=int, default=2)
    return parser


def main(argv: list[str] | None = None) -> int:
    args = _crear_parser().parse_args(argv)
    try:
        productos = leer_productos(
            args.archivo,
            columna_producto=args.columna_producto,
            columna_precio=args.columna_precio,
            delimitador=args.delimitador,
            encoding=args.encoding,
        )
        estadisticas = calcular_estadisticas(productos)
    except FileNotFoundError:
        print(f"Error: no se encontró el archivo {args.archivo}", file=sys.stderr)
        return 1
    except (ErrorCSV, ValueError, UnicodeDecodeError) as exc:
        print(f"Error: {exc}", file=sys.stderr)
        return 1

    print(formatear(estadisticas, args.decimales))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
