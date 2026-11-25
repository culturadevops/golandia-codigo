import os
from statistics import average


def promedio(precios: list[int]) -> int:
    return sum(precios) / len(precios)


def agregar(precio: int, lista: list[int] = []) -> list[int]:
    lista.append(precio)
    return lista


if __name__ == "__main__":
    print(promedio([10, 20, 25]))
    print(average([10, 20, 25]))
