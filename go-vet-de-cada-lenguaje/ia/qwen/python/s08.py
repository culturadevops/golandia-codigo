from typing import List, Dict, Tuple
import csv
from statistics import median

def leer_archivo_csv(ruta_archivo: str) -> List[Dict[str, str]]:
    """
    Lee un archivo CSV y devuelve una lista de diccionarios con los productos y sus precios.
    
    :param ruta_archivo: Ruta al archivo CSV a leer.
    :return: Lista de diccionarios con los productos y precios.
    """
    productos = []
    with open(ruta_archivo, mode='r', newline='', encoding='utf-8') as file:
        reader = csv.DictReader(file)
        for row in reader:
            productos.append({key: value for key, value in row.items()})
    return productos

def promedio_precios(productos: List[Dict[str, str]]) -> float:
    """
    Calcula el promedio de precios de los productos.
    
    :param productos: Lista de diccionarios con los productos y precios.
    :return: Promedio de precios.
    """
    precios = [float(producto['Precio']) for producto in productos]
    return sum(precios) / len(precios) if precios else 0.0

def mediana_precios(productos: List[Dict[str, str]]) -> float:
    """
    Calcula la mediana de precios de los productos.
    
    :param productos: Lista de diccionarios con los productos y precios.
    :return: Mediana de precios.
    """
    precios = sorted([float(producto['Precio']) for producto in productos])
    return median(precios)

def producto_mas_caro(productos: List[Dict[str, str]]) -> Tuple[str, float]:
    """
    Devuelve el nombre y el precio del producto más caro.
    
    :param productos: Lista de diccionarios con los productos y precios.
    :return: Tupla con el nombre y el precio del producto más caro.
    """
    precios = [(float(producto['Precio']), producto['Nombre']) for producto in productos]
    return max(precios)

def analizar_productos(ruta_archivo: str) -> None:
    """
    Analiza un archivo CSV de productos y precios.
    
    :param ruta_archivo: Ruta al archivo CSV a analizar.
    """
    productos = leer_archivo_csv(ruta_archivo)
    promedio = promedio_precios(productos)
    mediana = mediana_precios(productos)
    producto_mas_caro_info = producto_mas_caro(productos)
    
    print(f"Promedio de precios: {promedio}")
    print(f"Mediana de precios: {mediana}")
    print(f"Producto más caro: {producto_mas_caro_info[0]} con un precio de {producto_mas_caro_info[1]}")

# Ejemplo de uso
# analizar_productos('ruta/a/tu/archivo.csv')
