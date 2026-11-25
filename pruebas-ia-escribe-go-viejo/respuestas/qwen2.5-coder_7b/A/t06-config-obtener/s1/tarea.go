package tarea

func Obtener(cfg map[string]interface{}, clave string, porDefecto interface{}) interface{} {
	if valor, existe := cfg[clave]; existe {
		return valor
	}
	return porDefecto
}
