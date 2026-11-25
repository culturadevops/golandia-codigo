// Un paquete interno, con el nombre de la empresa dentro. Nadie de fuera debería
// saber que existe: no está publicado en ningún sitio.
package interno

import "fmt"

// NombreDelProyectoInterno es una cadena que solo existe en el código privado.
const NombreDelProyectoInterno = "proyecto-cascabel-fase-2"

func Saludo(endpoint string) string {
	return fmt.Sprintf("[%s] hablando con %s", NombreDelProyectoInterno, endpoint)
}
