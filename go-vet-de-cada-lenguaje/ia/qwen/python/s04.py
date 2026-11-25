from typing import List
import csv
from statistics import median

def leer_archivo_csv(ruta_archivo: str) -> List[List[float]]:
    productos = []
    with open(ruta_archivo, mode='r', newline='') as file:
        csv_reader = csv.reader(file)
        next(csv_reader)  # Saltar encabezado
        for fila in csv_reader:
            nombre, precio = fila
            productos.append([nombre, float(precio)])
    return productos

def calcular_promedio(productos: List[List[float]]) -> float:
    precios = [precio for _, precio in productos]
    return sum(precios) / len(precios)

def calcular_mediana(productos: List[List[float]]) -> float:
    precios = [precio for _, precio in productos]
    return median(precios)

def calcular_producto_mas_caro(productos: List[List[float]]) -> List[float]:
    precios = [precio for _, precio in productos]
    max_precio = max(precios)
    for nombre, precio in productos:
        if precio == max_precio:
            return [nombre, precio]
    return []

def main(ruta_archivo: str) -> None:
    productos = leer_archivo_csv(ruta_archivo)
    promedio = calcular_promedio(productos)
    mediana = calcular_mediana(productos)
    mas_caro = calcular_producto_mas_caro(productos)
    
    print(f"Promedio: {promedio}")
    print(f"Mediana: {mediana}")
    print(f"Producto más caro: {mas_caro[0]} con precio de ${mas_caro[1]:.2f}")

# Ejemplo de uso
# main('productos.csv')
