from typing import List, Tuple
import csv
from statistics import median, mean

def leer_archivo_csv(archivo: str) -> List[Tuple[str, float]]:
    """
    Lee un archivo CSV y devuelve una lista de tuplas (nombre, precio).
    
    :param archivo: Ruta del archivo CSV.
    :return: Lista de tuplas (nombre, precio).
    """
    productos = []
    with open(archivo, mode='r', newline='', encoding='utf-8') as file:
        csv_reader = csv.reader(file)
        next(csv_reader)  # Saltear la cabecera si la hay
        for nombre, precio in csv_reader:
            productos.append((nombre, float(precio)))
    return productos

def calcular_promedio(productos: List[Tuple[str, float]]) -> float:
    """
    Calcula el promedio de precios de la lista de productos.
    
    :param productos: Lista de tuplas (nombre, precio).
    :return: Promedio de precios.
    """
    precios = [precio for _, precio in productos]
    return mean(precios)

def calcular_mediana(productos: List[Tuple[str, float]]) -> float:
    """
    Calcula la mediana de precios de la lista de productos.
    
    :param productos: Lista de tuplas (nombre, precio).
    :return: Mediana de precios.
    """
    precios = [precio for _, precio in productos]
    return median(precios)

def calcular_producto_mas_caro(productos: List[Tuple[str, float]]) -> Tuple[str, float]:
    """
    Calcula el producto más caro de la lista de productos.
    
    :param productos: Lista de tuplas (nombre, precio).
    :return: Tupla con el nombre y precio del producto más caro.
    """
    return max(productos, key=lambda x: x[1])

def main(archivo: str):
    productos = leer_archivo_csv(archivo)
    promedio = calcular_promedio(productos)
    mediana = calcular_mediana(productos)
    producto_mas_caro = calcular_producto_mas_caro(productos)
    print(f"Promedio de precios: {promedio}")
    print(f"Mediana de precios: {mediana}")
    print(f"Producto más caro: {producto_mas_caro[0]} con {producto_mas_caro[1]}")

if __name__ == "__main__":
    import sys
    if len(sys.argv) > 1:
        archivo = sys.argv[1]
        main(archivo)
    else:
        print("Por favor, proporciona el nombre del archivo CSV como argumento.")
