from typing import List
import csv
from statistics import median

def leer_csv_ruta(ruta: str) -> List[tuple]:
    """Lee un archivo CSV y devuelve una lista de tuplas con los productos y sus precios."""
    productos = []
    with open(ruta, mode='r', newline='', encoding='utf-8') as file:
        csv_reader = csv.reader(file)
        next(csv_reader)  # Salta la primera fila si es el encabezado
        for row in csv_reader:
            if len(row) == 2:
                producto, precio = row
                productos.append((producto, float(precio)))
    return productos

def calcular_promedio_precios(productos: List[tuple]) -> float:
    """Calcula el promedio de precios de los productos."""
    precios = [precio for _, precio in productos]
    return sum(precios) / len(precios)

def calcular_mediana_precios(productos: List[tuple]) -> float:
    """Calcula la mediana de precios de los productos."""
    precios = [precio for _, precio in productos]
    return median(precios)

def calcular_producto_mas_caro(productos: List[tuple]) -> tuple:
    """Calcula el producto más caro."""
    return max(productos, key=lambda x: x[1])

def main(ruta: str):
    productos = leer_csv_ruta(ruta)
    promedio = calcular_promedio_precios(productos)
    mediana = calcular_mediana_precios(productos)
    mas_caro = calcular_producto_mas_caro(productos)
    
    print(f"Promedio de precios: {promedio}")
    print(f"Mediana de precios: {mediana}")
    print(f"Producto más caro: {mas_caro[0]} con precio de {mas_caro[1]}")

# Ejemplo de uso
# main('productos.csv')
