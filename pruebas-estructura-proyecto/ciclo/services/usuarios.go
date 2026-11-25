package services

import "github.com/golandia/ciclo/models"

func NombreValido(n string) bool { return n != "" }

func Crear(n string) models.Usuario { return models.Usuario{Nombre: n} }
