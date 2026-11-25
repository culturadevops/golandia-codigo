package driver

import "fmt"

// Este es el patron real: database/sql registra sus drivers asi.
func init() { fmt.Println("  1. init() del paquete importado (driver)") }
