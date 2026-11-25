from typing import List, Tuple
import csv
from statistics import median, mean

def leer_archivo_csv(archivo: str) -> List[Tuple[str, float]]:
    """
    Lee un archivo CSV y devuelve una lista de tuplas con el nombre del producto y su precio.
    
    :param archivo: Ruta del archivo CSV.
    :return: Lista de tuplas (nombre, precio).
    """
    productos = []
    with open(archivo, mode='r', encoding='utf-8') as file:
        reader = csv.reader(file)
        next(reader)  # Saltar encabezado si hay
        for nombre, precio in reader:
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

def encontrar_producto_mas_caro(productos: List[Tuple[str, float]]) -> Tuple[str, float]:
    """
    Encuentra el producto más caro en la lista de productos.
    
    :param productos: Lista de tuplas (nombre, precio).
    :return: Tupla (nombre, precio) del producto más caro.
    """
    return max(productos, key=lambda x: x[1])

# Ejemplo de uso
if __name__ == "__main__":
    archivo = 'productos.csv'  # Asegúrate de que el archivo CSV tenga los nombres correctos
    productos = leer_archivo_csv(archivo)
    promedio = calcular_promedio(productos)
    mediana = calcular_mediana(productos)
    producto_caro = encontrar_producto_mas_caro(productos)
    
    print(f"Promedio de precios: {promedio}")
    print(f"Mediana de precios: {mediana}")
    print(f"Producto más caro: {producto_caro[0]} con precio ${producto_caro[1]:.2f}")
