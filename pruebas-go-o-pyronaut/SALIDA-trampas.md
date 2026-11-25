# SALIDA-trampas — Pyronaut 0.1.2: lo que se tropieza al montarlo

> Corrida del 2026-11-23 07:48. En la etapa `build` de la imagen (gl-prac-pyronaut-build:1: python:3.13.17-slim + pyronaut 0.1.2 + `pyronaut setup` + el proyecto en /p).

```
### T1 · lo que instala `pyronaut setup` (du -sh)
1.2G	/root/.local/share/pyronaut
262M	/root/.m2

### T2 · el modo «un solo archivo» (pyronaut dev main.py) con la lógica en otro archivo
import logica
from micronaut.http.annotation import Get

@Get("/suma")
def suma(n: int = 0) -> str:
    return logica.suma(n)
rc=124
Message: Failed to initialize GraalPy context: ModuleNotFoundError: No module named 'logica'
Message: Failed to initialize GraalPy context: ModuleNotFoundError: No module named 'logica'

### T3 · un __init__.py en el paquete, como en cualquier proyecto de Python
rc=5
Processing failed: Pyronaut processing failed: Custom __init__.py files are not supported in Pyronaut applications. Micronaut generates package __init__.py files for the GraalPy VFS; remove [banco/__init__.py] from the project source.
```
