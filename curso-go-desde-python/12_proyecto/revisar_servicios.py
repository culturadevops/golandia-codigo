# -*- coding: utf-8 -*-
"""Leccion 12 — Todo junto: un script que se porta bien dentro de una tuberia.

Lo que separa un script de juguete de uno que puedes meter en CI:
  · argumentos de verdad (argparse), con --help que se explica solo
  · registro por niveles (logging), a stderr, para no ensuciar la salida
  · salida legible para una persona O json para una maquina
  · CODIGOS DE SALIDA distintos segun lo que pase, que es lo que mira CI

Pruebalo:
    python revisar_servicios.py
    python revisar_servicios.py --json
    python revisar_servicios.py --umbral 0.3 -v
    python revisar_servicios.py --json ; echo "codigo de salida: $?"
"""
import argparse
import json
import logging
import sys
from dataclasses import dataclass, asdict

log = logging.getLogger("revisar")

SALIDA_OK = 0
SALIDA_HAY_PROBLEMAS = 1
SALIDA_ERROR_DE_USO = 2


@dataclass
class Servicio:
    nombre: str
    estado: str
    cpu: float

    @property
    def sano(self):
        return self.estado == "activo"


# En la vida real esto vendria de una API o de un fichero. Aqui va fijo
# para que el ejemplo se ejecute igual en cualquier maquina.
INVENTARIO = [
    Servicio("web-01", "activo", 0.91),
    Servicio("web-02", "parado", 0.00),
    Servicio("db-01", "activo", 0.42),
    Servicio("cache-01", "activo", 0.88),
]


def revisar(servicios, umbral):
    """Devuelve (caidos, saturados). No imprime: solo decide."""
    caidos = [s for s in servicios if not s.sano]
    saturados = [s for s in servicios if s.sano and s.cpu > umbral]
    log.debug("revisados %d servicios con umbral %.2f", len(servicios), umbral)
    return caidos, saturados


def construir_argumentos():
    p = argparse.ArgumentParser(
        description="Revisa el inventario de servicios y avisa de lo que no esta bien.",
        epilog="Codigos de salida: 0 todo bien · 1 hay problemas · 2 error de uso")
    p.add_argument("--umbral", type=float, default=0.8,
                   help="cpu por encima de la cual se avisa (por omision: 0.8)")
    p.add_argument("--json", action="store_true",
                   help="salida en json, para otra herramienta")
    p.add_argument("-v", "--verboso", action="store_true",
                   help="ensena tambien los mensajes de depuracion")
    return p


def main(argv=None):
    args = construir_argumentos().parse_args(argv)

    # El registro va a STDERR a proposito: asi `script --json > fichero.json`
    # deja el json limpio y los mensajes siguen viendose en la consola.
    logging.basicConfig(level=logging.DEBUG if args.verboso else logging.INFO,
                        stream=sys.stderr,
                        format="%(levelname)-8s %(message)s")

    if not 0 <= args.umbral <= 1:
        log.error("el umbral tiene que estar entre 0 y 1, y me diste %s", args.umbral)
        return SALIDA_ERROR_DE_USO

    caidos, saturados = revisar(INVENTARIO, args.umbral)

    if args.json:
        print(json.dumps({
            "umbral": args.umbral,
            "total": len(INVENTARIO),
            "caidos": [asdict(s) for s in caidos],
            "saturados": [asdict(s) for s in saturados],
        }, indent=2, ensure_ascii=False))
    else:
        print(f"Revisados {len(INVENTARIO)} servicios (umbral {args.umbral:.0%})")
        for s in caidos:
            print(f"  CAIDO      {s.nombre}")
        for s in saturados:
            print(f"  SATURADO   {s.nombre}  cpu {s.cpu:.0%}")
        if not caidos and not saturados:
            print("  todo correcto")

    if caidos or saturados:
        log.warning("%d caidos, %d saturados", len(caidos), len(saturados))
        return SALIDA_HAY_PROBLEMAS
    return SALIDA_OK


if __name__ == "__main__":
    sys.exit(main())
