from typing import List, Tuple
import csv
from statistics import median, mean

def leer_archivo_csv(ruta_archivo: str) -> List[Tuple[str, float]]:
    """
    Lee un archivo CSV y devuelve una lista de tuplas (nombre, precio).
    
    :param ruta_archivo: Ruta del archivo CSV.
    :return: Lista de tuplas (nombre, precio).
    """
    productos = []
    with open(ruta_archivo, mode='r', newline='', encoding='utf-8') as file:
        csv_reader = csv.reader(file)
        next(csv_reader)  # Salta la primera línea (encabezado, si es necesario)
        for nombre, precio_str in csv_reader:
            try:
                precio = float(precio_str)
                productos.append((nombre, precio))
            except ValueError:
                print(f"Warning: No se pudo convertir '{precio_str}' a float. Se omitirá el producto.")
    return productos

def calcular_promedio(productos: List[Tuple[str, float]]) -> float:
    """
    Calcula el promedio de precios de los productos.
    
    :param productos: Lista de tuplas (nombre, precio).
    :return: Promedio de precios.
    """
    precios = [precio for _, precio in productos]
    return mean(precios)

def calcular_mediana(productos: List[Tuple[str, float]]) -> float:
    """
    Calcula la mediana de precios de los productos.
    
    :param productos: Lista de tuplas (nombre, precio).
    :return: Mediana de precios.
    """
    precios = [precio for _, precio in productos]
    return median(precios)

def calcular_producto_mas_caro(productos: List[Tuple[str, float]]) -> Tuple[str, float]:
    """
    Calcula el producto más caro.
    
    :param productos: Lista de tuplas (nombre, precio).
    :return: Tupla (nombre, precio) del producto más caro.
    """
    max_producto = max(productos, key=lambda x: x[1])
    return max_producto

if __name__ == "__main__":
    productos = leer_archivo_csv("productos.csv")
    promedio = calcular_promedio(productos)
    mediana = calcular_mediana(productos)
    producto_caro = calcular_producto_mas_caro(productos)
    print(f"Promedio de precios: {promedio}")
    print(f"Mediana de precios: {mediana}")
    print(f"Producto más caro: {producto_caro}")
