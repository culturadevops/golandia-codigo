# 📍 Aquí está la EVIDENCIA. El código está en `<nubecita>\`

> Movido el 2026-08-25 a petición del dueño: *«el código de esa nueva herramienta dónde está?
> quiero que esté en una carpeta aparte»*.

## Por qué están separados

| dónde | qué | por qué ahí |
|---|---|---|
| **`<nubecita>\`** | el **código** de la herramienta (`main.go`, `almacen.go`, `s3.go`, `s3_test.go`, `README.md`, `go.mod`) | Es una **herramienta**, no material de un vídeo: crece por episodios, se puede publicar como repositorio, y se copia sin arrastrar medio canal detrás |
| **esta carpeta** | `SALIDA.md` — la salida **literal** de compilar, probar y demostrar | Es el **respaldo de lo que se afirma en pantalla**. El canon exige que sobreviva al vídeo, así que vive con el canal, igual que el resto de `recursos/pruebas-*` |

⛔ No se borró nada: los ficheros se **movieron**. Los desechables (`nubecita.exe`, `_demo/`,
`_servidor.log`) no se movieron porque los regenera el corredor en cada corrida.

## Cómo se regenera todo

```powershell
python canales/golandia/investigacion/pruebas_nubecita.py
```

Ese corredor compila desde `<nubecita>`, pasa `go vet`, ejecuta las pruebas contra el SDK
oficial de AWS, levanta el servidor y le da con la **CLI oficial de Amazon**. Después escribe
`SALIDA.md` aquí. Si algo no sale como se espera, escribe `SALIDA.FALLIDA.md` **al lado** y no
toca el bueno.

## Qué vídeos se apoyan en esto

| vídeo | slug | episodio |
|---:|---|---|
| 127 | `mi-propio-localstack-en-go` | 1 · S3 |
