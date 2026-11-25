"""Estadísticas de precios a partir de un archivo CSV de productos.

El CSV debe tener encabezados con columnas de producto y precio, por ejemplo:

    producto,precio
    Teclado,25.50
    Monitor,189.99

Uso:
    python estadisticas_precios.py productos.csv
    python estadisticas_precios.py productos.csv --columna-producto nombre --columna-precio costo
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
    cantidad: int
    promedio: Decimal
    mediana: Decimal
    mas_caro: Producto


class ErrorCSV(Exception):
    """Error al leer o interpretar el archivo CSV."""


def parsear_precio(texto: str) -> Decimal:
    """Convierte un texto en precio, aceptando coma o punto decimal y símbolo $."""
    limpio = texto.strip().replace("$", "").replace(" ", "")
    if "," in limpio and "." in limpio:
        # Formato tipo 1.234,56 o 1,234.56: el último separador es el decimal.
        if limpio.rfind(",") > limpio.rfind("."):
            limpio = limpio.replace(".", "").replace(",", ".")
        else:
            limpio = limpio.replace(",", "")
    elif "," in limpio:
        limpio = limpio.replace(",", ".")
    try:
        precio = Decimal(limpio)
    except InvalidOperation as exc:
        raise ValueError(f"precio inválido: {texto!r}") from exc
    if not precio.is_finite() or precio < 0:
        raise ValueError(f"precio fuera de rango: {texto!r}")
    return precio


def leer_productos(
    ruta: Path,
    columna_producto: str = "producto",
    columna_precio: str = "precio",
    delimitador: str | None = None,
) -> list[Producto]:
    """Lee el CSV y devuelve la lista de productos con precio válido."""
    try:
        with ruta.open(encoding="utf-8-sig", newline="") as archivo:
            muestra = archivo.read(4096)
            archivo.seek(0)
            if delimitador is None:
                try:
                    delimitador = csv.Sniffer().sniff(muestra, delimiters=",;\t|").delimiter
                except csv.Error:
                    delimitador = ","
            lector = csv.DictReader(archivo, delimiter=delimitador)
            campos = [c.strip().lower() for c in (lector.fieldnames or [])]
            col_prod = columna_producto.strip().lower()
            col_precio = columna_precio.strip().lower()
            faltantes = [c for c in (col_prod, col_precio) if c not in campos]
            if faltantes:
                raise ErrorCSV(
                    f"faltan columnas {faltantes}; encontradas: {campos}"
                )

            productos: list[Producto] = []
            for numero_linea, fila in enumerate(lector, start=2):
                normalizada = {
                    (k or "").strip().lower(): (v or "") for k, v in fila.items()
                }
                nombre = normalizada[col_prod].strip()
                texto_precio = normalizada[col_precio]
                if not nombre and not texto_precio.strip():
                    continue  # línea vacía
                try:
                    precio = parsear_precio(texto_precio)
                except ValueError as exc:
                    print(f"Aviso: línea {numero_linea} ignorada ({exc})", file=sys.stderr)
                    continue
                productos.append(Producto(nombre or "(sin nombre)", precio))
            return productos
    except OSError as exc:
        raise ErrorCSV(f"no se pudo leer {ruta}: {exc}") from exc


def calcular_estadisticas(productos: list[Producto]) -> Estadisticas:
    """Calcula promedio, mediana y producto más caro."""
    if not productos:
        raise ValueError("no hay productos con precio válido")
    precios = [p.precio for p in productos]
    return Estadisticas(
        cantidad=len(productos),
        promedio=statistics.fmean(precios) and Decimal(sum(precios)) / len(precios),
        mediana=Decimal(statistics.median(precios)),
        mas_caro=max(productos, key=lambda p: p.precio),
    )


def formatear(estadisticas: Estadisticas) -> str:
    """Devuelve un resumen legible de las estadísticas."""
    q = Decimal("0.01")
    return "\n".join(
        [
            f"Productos analizados: {estadisticas.cantidad}",
            f"Precio promedio:      {estadisticas.promedio.quantize(q)}",
            f"Precio mediano:       {estadisticas.mediana.quantize(q)}",
            f"Producto más caro:    {estadisticas.mas_caro.nombre} "
            f"({estadisticas.mas_caro.precio.quantize(q)})",
        ]
    )


def construir_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="Calcula promedio, mediana y producto más caro de un CSV."
    )
    parser.add_argument("archivo", type=Path, help="ruta al archivo CSV")
    parser.add_argument("--columna-producto", default="producto")
    parser.add_argument("--columna-precio", default="precio")
    parser.add_argument("--delimitador", default=None, help="por defecto se detecta")
    return parser


def main(argv: list[str] | None = None) -> int:
    args = construir_parser().parse_args(argv)
    try:
        productos = leer_productos(
            args.archivo, args.columna_producto, args.columna_precio, args.delimitador
        )
        print(formatear(calcular_estadisticas(productos)))
    except (ErrorCSV, ValueError) as exc:
        print(f"Error: {exc}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
