from __future__ import annotations
from typing import List, Dict
import csv
from statistics import median

def leer_archivo_csv(ruta_archivo: str) -> List[Dict[str, str]]:
    """
    Lee un archivo CSV y devuelve una lista de diccionarios con los productos y sus precios.
    
    :param ruta_archivo: Ruta del archivo CSV a leer.
    :return: Lista de diccionarios, donde cada diccionario contiene un producto y su precio.
    """
    productos = []
    with open(ruta_archivo, mode='r', encoding='utf-8') as archivo:
        lector_csv = csv.DictReader(archivo)
        for fila in lector_csv:
            productos.append(fila)
    return productos

def calcular_promedio_precios(productos: List[Dict[str, str]]) -> float:
    """
    Calcula el promedio de los precios de los productos.
    
    :param productos: Lista de diccionarios con los productos y sus precios.
    :return: Promedio de los precios.
    """
    precios = [float(producto['precio']) for producto in productos]
    return sum(precios) / len(precios) if precios else 0

def calcular_mediana_precios(productos: List[Dict[str, str]]) -> float:
    """
    Calcula la mediana de los precios de los productos.
    
    :param productos: Lista de diccionarios con los productos y sus precios.
    :return: Mediana de los precios.
    """
    precios = [float(producto['precio']) for producto in productos]
    return median(precios)

def calcular_producto_mas_caro(productos: List[Dict[str, str]]) -> Dict[str, str]:
    """
    Calcula el producto más caro.
    
    :param productos: Lista de diccionarios con los productos y sus precios.
    :return: Diccionario con el producto más caro y su precio.
    """
    if not productos:
        return {}
    productos_ordenados = sorted(productos, key=lambda x: float(x['precio']), reverse=True)
    return productos_ordenados[0]

if __name__ == "__main__":
    productos = leer_archivo_csv('productos.csv')
    promedio = calcular_promedio_precios(productos)
    mediana = calcular_mediana_precios(productos)
    producto_mas_caro = calcular_producto_mas_caro(productos)
    
    print(f"Promedio de precios: {promedio}")
    print(f"Mediana de precios: {mediana}")
    print(f"Producto más caro: {producto_mas_caro}")
