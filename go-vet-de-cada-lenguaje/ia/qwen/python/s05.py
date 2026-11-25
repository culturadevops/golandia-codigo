from typing import List, Dict
import csv
from statistics import median

def leer_archivo_csv(archivo: str) -> List[Dict[str, str]]:
    productos: List[Dict[str, str]] = []
    with open(archivo, mode='r', encoding='utf-8') as file:
        reader = csv.DictReader(file)
        for row in reader:
            productos.append(row)
    return productos

def calcular_promedio_precios(productos: List[Dict[str, str]]) -> float:
    precios = [float(precio) for precio in (producto['precio'] for producto in productos)]
    return sum(precios) / len(precios)

def calcular_mediana_precios(productos: List[Dict[str, str]]) -> float:
    precios = [float(precio) for precio in (producto['precio'] for producto in productos)]
    return median(precios)

def encontrar_producto_mas_caro(productos: List[Dict[str, str]]) -> Dict[str, str]:
    productos_ordenados = sorted(productos, key=lambda x: float(x['precio']), reverse=True)
    return productos_ordenados[0]

if __name__ == "__main__":
    productos = leer_archivo_csv('productos.csv')
    promedio = calcular_promedio_precios(productos)
    mediana = calcular_mediana_precios(productos)
    producto_mas_caro = encontrar_producto_mas_caro(productos)
    
    print(f"Promedio de precios: {promedio}")
    print(f"Mediana de precios: {mediana}")
    print(f"Producto más caro: {producto_mas_caro}")
