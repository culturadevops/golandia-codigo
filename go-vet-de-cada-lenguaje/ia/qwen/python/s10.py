from typing import List, Tuple
import csv
from statistics import median

def leer_archivo_csv(ruta_archivo: str) -> List[Tuple[str, float]]:
    """
    Lee un archivo CSV y devuelve una lista de tuplas con el nombre del producto y su precio.
    
    :param ruta_archivo: Ruta del archivo CSV.
    :return: Lista de tuplas (nombre, precio).
    """
    productos = []
    with open(ruta_archivo, mode='r', newline='', encoding='utf-8') as file:
        reader = csv.reader(file)
        next(reader)  # Salta la primera fila si es encabezado
        for nombre, precio in reader:
            productos.append((nombre, float(precio)))
    return productos

def calcular_promedio_precios(productos: List[Tuple[str, float]]) -> float:
    """
    Calcula el promedio de los precios de los productos.
    
    :param productos: Lista de tuplas (nombre, precio).
    :return: Promedio de precios.
    """
    precios = [precio for _, precio in productos]
    return sum(precios) / len(precios) if precios else 0

def calcular_mediana_precios(productos: List[Tuple[str, float]]) -> float:
    """
    Calcula la mediana de los precios de los productos.
    
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
    return max(productos, key=lambda x: x[1])

def main(ruta_archivo: str) -> Tuple[float, float, Tuple[str, float]]:
    """
    Función principal que lee el archivo CSV y calcula el promedio, mediana y producto más caro.
    
    :param ruta_archivo: Ruta del archivo CSV.
    :return: Tupla con el promedio, mediana y producto más caro.
    """
    productos = leer_archivo_csv(ruta_archivo)
    promedio = calcular_promedio_precios(productos)
    mediana = calcular_mediana_precios(productos)
    producto_mas_caro = calcular_producto_mas_caro(productos)
    return promedio, mediana, producto_mas_caro

if __name__ == "__main__":
    # Ejemplo de uso
    ruta_archivo = 'productos.csv'  # Ruta del archivo CSV
    promedio, mediana, producto_mas_caro = main(ruta_archivo)
    print(f"Promedio de precios: {promedio}")
    print(f"Mediana de precios: {mediana}")
    print(f"Producto más caro: {producto_mas_caro}")
