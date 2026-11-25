"""Estadísticas de precios a partir de un archivo CSV de productos.

Formato esperado (con encabezado):

    producto,precio
    Teclado,25.50
    Monitor,180.00

Uso:
    python estadisticas_precios.py productos.csv [--delimitador ";"]
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
class Resumen:
    cantidad: int
    promedio: Decimal
    mediana: Decimal
    mas_caro: Producto


class ErrorCSV(Exception):
    """Error de formato o contenido en el archivo CSV."""


def _parsear_precio(texto: str, linea: int) -> Decimal:
    limpio = texto.strip().replace("$", "").replace(" ", "")
    # Admite coma decimal ("12,50") si no hay punto.
    if "," in limpio and "." not in limpio:
        limpio = limpio.replace(",", ".")
    try:
        precio = Decimal(limpio)
    except InvalidOperation as exc:
        raise ErrorCSV(f"Línea {linea}: precio inválido {texto!r}") from exc
    if not precio.is_finite() or precio < 0:
        raise ErrorCSV(f"Línea {linea}: precio fuera de rango {texto!r}")
    return precio


def leer_productos(
    ruta: Path,
    columna_producto: str = "producto",
    columna_precio: str = "precio",
    delimitador: str = ",",
) -> list[Producto]:
    with ruta.open(newline="", encoding="utf-8-sig") as archivo:
        lector = csv.DictReader(archivo, delimiter=delimitador)
        campos = {c.strip().lower(): c for c in (lector.fieldnames or [])}
        if columna_producto not in campos or columna_precio not in campos:
            raise ErrorCSV(
                f"El CSV debe tener las columnas '{columna_producto}' y "
                f"'{columna_precio}'; encontradas: {list(campos.values())}"
            )
        col_nombre = campos[columna_producto]
        col_precio = campos[columna_precio]

        productos: list[Producto] = []
        for fila in lector:
            linea = lector.line_num
            nombre = (fila.get(col_nombre) or "").strip()
            precio_txt = (fila.get(col_precio) or "").strip()
            if not nombre and not precio_txt:
                continue  # fila vacía
            if not nombre:
                raise ErrorCSV(f"Línea {linea}: falta el nombre del producto")
            productos.append(Producto(nombre, _parsear_precio(precio_txt, linea)))
    return productos


def calcular_resumen(productos: list[Producto]) -> Resumen:
    if not productos:
        raise ValueError("No hay productos para analizar")
    precios = [p.precio for p in productos]
    return Resumen(
        cantidad=len(productos),
        promedio=sum(precios, Decimal(0)) / len(precios),
        mediana=statistics.median(precios),
        mas_caro=max(productos, key=lambda p: p.precio),
    )


def formatear(valor: Decimal) -> str:
    return f"{valor.quantize(Decimal('0.01')):,}"


def _argumentos(argv: list[str] | None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("archivo", type=Path, help="Ruta al archivo CSV")
    parser.add_argument("--delimitador", default=",", help="Separador (por defecto ',')")
    parser.add_argument("--col-producto", default="producto")
    parser.add_argument("--col-precio", default="precio")
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = _argumentos(argv)
    try:
        productos = leer_productos(
            args.archivo,
            args.col_producto.lower(),
            args.col_precio.lower(),
            args.delimitador,
        )
        resumen = calcular_resumen(productos)
    except (OSError, ErrorCSV, ValueError) as exc:
        print(f"Error: {exc}", file=sys.stderr)
        return 1

    print(f"Productos:       {resumen.cantidad}")
    print(f"Precio promedio: {formatear(resumen.promedio)}")
    print(f"Precio mediano:  {formatear(resumen.mediana)}")
    print(
        f"Más caro:        {resumen.mas_caro.nombre} "
        f"({formatear(resumen.mas_caro.precio)})"
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())
