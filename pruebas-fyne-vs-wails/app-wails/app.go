package main

// El estado vive en Go, igual que en la versión de Fyne. La diferencia es que
// la interfaz lo pide por un puente, no lo lee directamente.
type App struct{ n int }

func (a *App) Sumar() int {
	a.n++
	return a.n
}

func (a *App) Items() []string {
	return []string{"uno", "dos", "tres", "cuatro", "cinco"}
}
