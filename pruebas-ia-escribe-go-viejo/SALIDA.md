# SALIDA — 703 «La IA escribe Go viejo: lo medí con go fix»

> Regenerado el 2026-10-29 09:25 con `python ejecutar_todo.py`. Medido en `golang:1.27.2` (Docker, --cpus 2), go.mod `go 1.27`.
> No es un estudio: 24 tareas propias, una sola máquina, un modelo local y una muestra pequeña de Claude Code.

## Quién generó
- Local: Ollama `qwen2.5-coder:7b`, digest `dae161e27b0e90dd1856c8bb3209201fd6736d8eb66298e75ed87571486f4364`, `/api/generate`, temperature 0.7, seeds 1-3, num_predict 2048.
- Claude Code 2.1.312, `claude -p` desde un directorio temporal vacío, flags `-p --safe-mode --tools  --strict-mcp-config --permission-prompts none --no-session-persistence --output-format json --max-budget-usd 0.50`. Modelos en modelUsage: {'claude-haiku-5-5': 29, 'claude-opus-5-5': 30}. Costo total de las 30 llamadas: 0.47 USD. Sin seed: una muestra por tarea y variante.
- `--bare` no se pudo usar: solo acepta ANTHROPIC_API_KEY y la cuenta entra por OAuth (respuesta: «Not logged in · Please run /login»). Se usó `--safe-mode` (sin hooks, MCP, CLAUDE.md ni plugins) y `--tools ""`.
- Prompts exactos: prompts.json (A sin versión · B añade «Usa Go 1.27.» · C añade además la lista de modismos, solo local). Tareas: tareas.json.

## Controles del banco
- control-viejo: compila=True vet=True avisos de go fix=15 {'any': 1, 'atomictypes': 1, 'errorsastype': 1, 'forvar': 1, 'mapsloop': 1, 'newexpr': 2, 'rangeint': 1, 'reflecttypefor': 1, 'slicescontains': 1, 'slicessort': 1, 'stringscut': 1, 'stringscutprefix': 1, 'stringsseq': 1, 'waitgroupgo': 1}
- control-moderno: compila=True vet=True avisos de go fix=0 {}

## Tabla principal (por modelo y prompt)
| modelo | prompt | muestras | compila | go vet limpio* | con ≥1 aviso de go fix* | avisos* | usa el modismo nuevo de su tarea* |
|---|---|---|---|---|---|---|---|
| claude-code | A | 15 | 15/15 (100 %) | 15/15 (100 %) | 5/15 (33 %) | 5 | 4/15 (27 %) |
| claude-code | B | 15 | 15/15 (100 %) | 15/15 (100 %) | 1/15 (7 %) | 1 | 13/15 (87 %) |
| qwen2.5-coder_7b | A | 72 | 57/72 (79 %) | 56/57 (98 %) | 28/57 (49 %) | 33 | 0/57 (0 %) |
| qwen2.5-coder_7b | B | 72 | 56/72 (78 %) | 55/56 (98 %) | 27/56 (48 %) | 32 | 1/56 (2 %) |
| qwen2.5-coder_7b | C | 72 | 24/72 (33 %) | 24/24 (100 %) | 5/24 (21 %) | 5 | 11/24 (46 %) |

*sobre las que compilan.

## Avisos de go fix por modernizer (solo las que compilan)
- claude-code A: {'rangeint': 1, 'stringsseq': 1, 'mapsloop': 1, 'slicescontains': 1, 'reflecttypefor': 1}
- claude-code B: {'stringsseq': 1}
- qwen2.5-coder_7b A: {'any': 9, 'rangeint': 6, 'mapsloop': 5, 'stringsbuilder': 3, 'slicescontains': 3, 'stringscutprefix': 2, 'newexpr': 2, 'errorsastype': 1, 'stringsseq': 1, 'atomictypes': 1}
- qwen2.5-coder_7b B: {'any': 8, 'rangeint': 7, 'mapsloop': 4, 'stringsbuilder': 3, 'slicescontains': 3, 'stringscutprefix': 2, 'newexpr': 2, 'errorsastype': 1, 'stringsseq': 1, 'atomictypes': 1}
- qwen2.5-coder_7b C: {'minmax': 2, 'stringscutprefix': 2, 'stringscut': 1}

## ¿Cambió algo la frase «Usa Go 1.27»? (mismo modelo, misma tarea, misma seed)
- qwen2.5-coder_7b: 46 de 72 respuestas B idénticas byte a byte a su A.
- claude-code: 2 de 15 respuestas B idénticas byte a byte a su A.

## Por qué no compilan (primer error)
- qwen2.5-coder_7b A: 9× "…" imported and not used; 5× undefined: strings; 1× s.indexOf undefined (type string has no field or method indexOf)
- qwen2.5-coder_7b B: 10× "…" imported and not used; 5× undefined: strings; 1× s.indexOf undefined (type string has no field or method indexOf)
- qwen2.5-coder_7b C: 37× "…" imported and not used; 6× undefined: atomic; 3× assignment mismatch: 1 variable but strings.Cut returns 3 values; 1× not enough return values; 1× multiple-value errors.AsType[*fs.PathError](err, &pathErr) (value of type (*fs.PathError, 
- Imports sin usar en el prompt C (local), por paquete: {'strings': 20, 'errors': 9, 'sync': 8, 'fmt': 6, 'sync/atomic': 2, 'math': 1, 'reflect': 1, 'path/filepath': 1}
- `go tool fix help` en golang:1.27.2: 26 analizadores (go-tool-fix-help.txt). `go help fix`: «exit with a non-zero status if the diff is not empty» (go-help-fix.txt).

## go vet: lo que encontró
- qwen2.5-coder_7b/A/t20-error-validacion/s3: tarea.go:17:5: second argument to errors.As must be a non-nil pointer to either a type that implements error, or to any interface type
- qwen2.5-coder_7b/B/t20-error-validacion/s3: tarea.go:17:5: second argument to errors.As must be a non-nil pointer to either a type that implements error, or to any interface type

## Por tarea: ¿usa el modismo nuevo? (A / B / C)
| tarea | modismo | Claude A | Claude B | local A | local B | local C |
|---|---|---|---|---|---|---|
| t01-suma-cuadrados | rangeint | 0/1 | 1/1 | 0/3 | 0/3 | 2/3 |
| t02-repetir | rangeint | — | — | 0/3 | 0/3 | 0/3 |
| t03-acotar | minmax | 0/1 | 0/1 | 0/3 | 0/3 | 0/3 |
| t04-mayor-de-tres | minmax | — | — | 0/3 | 0/3 | 2/3 |
| t05-describir | any | 1/1 | 1/1 | 0/3 | 1/3 | 0/3 |
| t06-config-obtener | any | — | — | 0/3 | 0/3 | 0/3 |
| t07-clave-valor | stringscut | 1/1 | 1/1 | 0/3 | 0/3 | 0/3 |
| t08-dominio-email | stringscut | — | — | 0/3 | 0/3 | 0/3 |
| t09-token-bearer | stringscutprefix | 0/1 | 1/1 | 0/3 | 0/3 | 0/3 |
| t10-palabras-largas | stringsseq | 0/1 | 0/1 | 0/3 | 0/3 | 0/3 |
| t11-sumar-csv | stringsseq | — | — | 0/3 | 0/3 | 0/3 |
| t12-linea-tabla | stringsbuilder | 1/1 | 1/1 | 0/3 | 0/3 | 3/3 |
| t13-copiar-mapa | mapsloop | 0/1 | 1/1 | 0/3 | 0/3 | 0/3 |
| t14-clonar-mapa | mapsloop | — | — | 0/3 | 0/3 | 0/3 |
| t15-contiene | slicescontains | 0/1 | 1/1 | 0/3 | 0/3 | 0/3 |
| t16-algun-negativo | slicescontains | — | — | 0/3 | 0/3 | 1/3 |
| t17-mediana | slicessort | 0/1 | 1/1 | 0/3 | 0/3 | 0/3 |
| t18-al-reves | slicesbackward | — | — | 0/3 | 0/3 | 0/3 |
| t19-ruta-error | errorsastype | 0/1 | 1/1 | 0/3 | 0/3 | 0/3 |
| t20-error-validacion | errorsastype | — | — | 0/3 | 0/3 | 0/3 |
| t21-config-punteros | newexpr | 0/1 | 1/1 | 0/3 | 0/3 | 0/3 |
| t22-paralelo | forvar/waitgroupgo | 0/1 | 1/1 | 0/3 | 0/3 | 0/3 |
| t23-nombre-tipo | reflecttypefor | 0/1 | 1/1 | 0/3 | 0/3 | 0/3 |
| t24-contador-atomico | atomictypes | 1/1 | 1/1 | 0/3 | 0/3 | 3/3 |

## Diffs de go fix -diff de Claude Code (completos)
### claude-code/A/t01-suma-cuadrados/s1 {'rangeint': 1}
```diff
--- /w/m/m0000/tarea.go (old)
+++ /w/m/m0000/tarea.go (new)
@@ -3,7 +3,7 @@
 // SumaCuadrados devuelve la suma de i*i para i desde 0 hasta n-1.
 func SumaCuadrados(n int) int {
 	suma := 0
-	for i := 0; i < n; i++ {
+	for i := range n {
 		suma += i * i
 	}
 	return suma
```
### claude-code/A/t10-palabras-largas/s1 {'stringsseq': 1}
```diff
--- /w/m/m0005/tarea.go (old)
+++ /w/m/m0005/tarea.go (new)
@@ -9,7 +9,7 @@
 // en blanco, tienen más de 3 letras.
 func PalabrasLargas(texto string) int {
 	cuenta := 0
-	for _, palabra := range strings.Fields(texto) {
+	for palabra := range strings.FieldsSeq(texto) {
 		letras := 0
 		for _, r := range palabra {
 			if unicode.IsLetter(r) {
```
### claude-code/A/t13-copiar-mapa/s1 {'mapsloop': 1}
```diff
--- /w/m/m0007/tarea.go (old)
+++ /w/m/m0007/tarea.go (new)
@@ -1,9 +1,9 @@
 package tarea
 
+import "maps"
+
 // Mezclar copia todas las entradas de src en dst, sobrescribiendo las claves
 // que ya existan. dst no debe ser nil si src tiene entradas.
 func Mezclar(dst, src map[string]int) {
-	for k, v := range src {
-		dst[k] = v
-	}
+	maps.Copy(dst, src)
 }
```
### claude-code/A/t15-contiene/s1 {'slicescontains': 1}
```diff
--- /w/m/m0008/tarea.go (old)
+++ /w/m/m0008/tarea.go (new)
@@ -1,11 +1,8 @@
 package tarea
 
+import "slices"
+
 // EsAdmin devuelve true si el slice de roles contiene el texto "admin".
 func EsAdmin(roles []string) bool {
-	for _, rol := range roles {
-		if rol == "admin" {
-			return true
-		}
-	}
-	return false
+	return slices.Contains(roles, "admin")
 }
```
### claude-code/A/t23-nombre-tipo/s1 {'reflecttypefor': 1}
```diff
--- /w/m/m0013/tarea.go (old)
+++ /w/m/m0013/tarea.go (new)
@@ -6,7 +6,7 @@
 // (por ejemplo, error). Para tipos sin nombre (como []int o map[string]int)
 // devuelve su representación textual.
 func NombreTipo[T any]() string {
-	t := reflect.TypeOf((*T)(nil)).Elem()
+	t := reflect.TypeFor[T]()
 	if nombre := t.Name(); nombre != "" {
 		return nombre
 	}
```
### claude-code/B/t10-palabras-largas/s1 {'stringsseq': 1}
```diff
--- /w/m/m0020/tarea.go (old)
+++ /w/m/m0020/tarea.go (new)
@@ -9,7 +9,7 @@
 // en blanco, tienen más de 3 letras.
 func PalabrasLargas(texto string) int {
 	cuenta := 0
-	for _, palabra := range strings.Fields(texto) {
+	for palabra := range strings.FieldsSeq(texto) {
 		letras := 0
 		for _, r := range palabra {
 			if unicode.IsLetter(r) {
```
