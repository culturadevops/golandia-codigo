from typing import List, Tuple
import csv
from statistics import median

def cargar_productos_archivo_csv(archivo: str) -> List[Tuple[str, float]]:
    """
    Lee un archivo CSV y devuelve una lista de tuplas (nombre, precio).
    """
    productos = []
    with open(archivo, mode='r', newline='', encoding='utf-8') as file:
        csv_reader = csv.reader(file)
        next(csv_reader)  # Suponiendo que la primera fila es el encabezado
        for nombre, precio in csv_reader:
            productos.append((nombre, float(precio)))
    return productos

def calcular_promedio(productos: List[Tuple[str, float]]) -> float:
    """
    Calcula el promedio de los precios de una lista de productos.
    """
    precios = [precio for _, precio in productos]
    return sum(precios) / len(precios)

def calcular_mediana(productos: List[Tuple[str, float]]) -> float:
    """
    Calcula la mediana de los precios de una lista de productos.
    """
    precios = [precio for _, precio in productos]
    return median(precios)

def encontrar_producto_mas_caro(productos: List[Tuple[str, float]]) -> Tuple[str, float]:
    """
    Encuentra el producto más caro en una lista de productos.
    """
    return max(productos, key=lambda x: x[1])

def main(archivo: str) -> Tuple[float, float, Tuple[str, float]]:
    """
    Lee un archivo CSV, calcula el promedio, la mediana y el producto más caro.
    """
    productos = cargar_productos_archivo_csv(archivo)
    promedio = calcular_promedio(productos)
    mediana = calcular_mediana(productos)
    producto_mas_caro = encontrar_producto_mas_caro(productos)
    return promedio, mediana, producto_mas_caro

# Ejemplo de uso
# promedio, mediana, producto_mas_caro = main('productos.csv')
# print(f'Promedio: {promedio}')
# print(f'Mediana: {mediana}')
# print(f'Producto más caro: {producto_mas_caro}')
