package models

import "github.com/golandia/ciclo/services"

type Usuario struct{ Nombre string }

func (u Usuario) Valido() bool { return services.NombreValido(u.Nombre) }
