from statistics import mean


def promedio(precios: list[int]) -> float:
    return sum(precios) / len(precios)


def agregar(precio: int, lista: list[int] | None = None) -> list[int]:
    lista = [] if lista is None else lista
    lista.append(precio)
    return lista


if __name__ == "__main__":
    print(promedio([10, 20, 25]))
    print(mean([10, 20, 25]))
